package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/microsoft/typescript-go/internal/ast"
	"github.com/microsoft/typescript-go/internal/core"
	"github.com/microsoft/typescript-go/internal/ls"
	"github.com/microsoft/typescript-go/internal/parser"
	"github.com/microsoft/typescript-go/internal/testutil/projecttestutil"
	"github.com/microsoft/typescript-go/internal/tspath"
)

func runResolve(args []string) int {
	if len(args) < 2 {
		fmt.Println("Usage: resolve <directory> <output_path> [limit]")
		return 1
	}
	dir := args[0]
	outputPath := args[1]
	fmt.Printf("Finding resolutions in %v...\n", dir)

	limit := 10000
	if len(args) > 2 {
		var err error
		limit, err = strconv.Atoi(args[2])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return 1
		}
	}
	start := time.Now()
	ctx := projecttestutil.WithRequestID(context.Background())
	files, err := parseFiles(dir, limit)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}
	fmt.Printf("Parsed %d files in %v\n", len(files), time.Since(start))
	service, done := createLanguageService(ctx, files)
	defer done()

	var totalResolutions int
	projectFiles := make([]File, 0, len(files))
	for file, content := range files {
		res, err := ResolutionsForFile(ctx, file, content, service)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return 1
		}
		projectFiles = append(projectFiles, File{
			Path:        file,
			Content:     content,
			Resolutions: res,
		})
		totalResolutions += len(res)
		printResolutionsInFile(file, content, res)
	}
	fmt.Printf("Found %v resolutions in %v files in %v\n", totalResolutions, len(files), time.Since(start))

	f, err := os.Create(outputPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create output file: %v\n", err)
		return 2
	}
	defer f.Close()

	project := Project{Files: projectFiles}
	if err := Encode(project, f); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to encode project asset: %v\n", err)
		return 2
	}
	fmt.Printf("Wrote asset to %s\n", outputPath)

	return 0
}

func ResolutionsForFile(ctx context.Context, path, content string, service *ls.LanguageService) ([]Resolution, error) {
	opts := ast.SourceFileParseOptions{
		Path:             tspath.Path(path),
		FileName:         path,
		JSDocParsingMode: ast.JSDocParsingModeParseAll,
	}
	scriptKind := core.GetScriptKindFromFileName(path)
	file := parser.ParseSourceFile(opts, content, scriptKind)
	if file == nil {
		return nil, fmt.Errorf("failed to parse file")
	}

	var ret []Resolution
	visitor := &ast.NodeVisitor{}
	visitor.Visit = func(n *ast.Node) *ast.Node {
		if n == nil {
			return n
		}
		if n.Kind != ast.KindIdentifier {
			visitor.VisitEachChild(n)
			return n
		}
		pos := service.PositionToLineAndCharacter(file, core.TextPos(n.Loc.Pos()+1))
		cands, err := service.ProvideDefinitionAsNodes(context.Background(), ls.FileNameToDocumentURI(path), pos)
		if err != nil {
			fmt.Printf("service.ProvideDefinitionAsNodes error: %v\n", err)
			return n
		}
		var defs []Definition
		for _, cand := range cands {
			name := ast.GetNameOfDeclaration(cand)
			if name == nil {
				continue
			}
			defs = append(defs, Definition{
				Path:  ast.GetSourceFileOfNode(name).FileName(),
				Range: Range{name.Loc.Pos(), name.Loc.End()},
			})
		}
		prefix := strings.TrimSuffix(content[n.Loc.Pos():n.Loc.End()], n.Text()) // strip WS from the node range
		stOffset := n.Loc.Pos() + len(prefix)
		name := content[stOffset:n.Loc.End()]
		if name != n.Text() {
			fmt.Printf("Warning: name mismatch for %s, got %q, expected %q\n", n.Kind, name, n.Text())
		}
		ret = append(ret, Resolution{
			Range:       Range{stOffset, n.Loc.End()},
			Definitions: defs,
		})
		return n
	}
	visitor.Visit(file.AsNode())
	return ret, nil
}

func printResolutionsInFile(fileName, content string, defs []Resolution) {
	fmt.Printf("********** File: %s **********\n", fileName)
	for _, def := range defs {
		name := content[def.Range.Offset:def.Range.EndOffset]
		fmt.Printf("Symbol: %q @ %s: [%d, %d)\n", name, fileName, def.Range.Offset, def.Range.EndOffset)
		for _, decl := range def.Definitions {
			prefix := "  Declaration:"
			if decl.Path != fileName {
				prefix = "  Cross-file declaration:"
			}
			fmt.Printf("%s @ %s: [%d, %d)\n", prefix, decl.Path, decl.Offset, decl.EndOffset)
		}
	}
}

func parseFiles(dir string, limit int) (map[string]string, error) {
	files := map[string]string{}
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error accessing %s: %v\n", path, err)
			return nil
		}
		if info.IsDir() || len(files) >= limit {
			return nil
		}
		ext := filepath.Ext(path)
		if _, ok := SupportedExtensions[ext]; !ok {
			return nil
		}
		data, err := ioutil.ReadFile(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to read %s: %v\n", path, err)
			return nil
		}
		files[path] = string(data)
		return nil
	})
	return files, err
}

func createLanguageService(ctx context.Context, files map[string]string) (*ls.LanguageService, func()) {
	projectService, _ := projecttestutil.Setup(files, nil)
	start := time.Now()
	var i int
	for fileName := range files {
		i++
		projectService.OpenFile(fileName, files[fileName], core.GetScriptKindFromFileName(fileName), "")
		if i%100 == 0 {
			fmt.Printf("Opening files, %v files opened out of %v...\n", i, len(files))
		}
	}
	fmt.Printf("Opened %d files in %v\n", i, time.Since(start))
	project := projectService.Projects()[0]
	return project.GetLanguageServiceForRequest(ctx)
}
