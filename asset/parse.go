package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/microsoft/typescript-go/internal/ast"
	"github.com/microsoft/typescript-go/internal/core"
	"github.com/microsoft/typescript-go/internal/parser"
)

func runParse(args []string) int {
	if len(args) < 2 {
		fmt.Fprintln(os.Stderr, "Usage: parseall <ts-directory> <output-asset-file>")
		return 2
	}
	dir := args[0]
	outputPath := args[1]

	fmt.Printf("Parsing everything in %s\n", dir)

	var failed bool
	var files []File
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error accessing %s: %v\n", path, err)
			failed = true
			return nil
		}
		if info.IsDir() {
			return nil
		}
		ext := filepath.Ext(path)
		if _, ok := SupportedExtensions[ext]; !ok {
			return nil
		}
		data, err := ioutil.ReadFile(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to read %s: %v\n", path, err)
			failed = true
			return nil
		}
		opts := ast.SourceFileParseOptions{
			FileName:         path,
			JSDocParsingMode: ast.JSDocParsingModeParseAll,
		}
		scriptKind := core.GetScriptKindFromFileName(path)
		parsedFile := parser.ParseSourceFile(opts, string(data), scriptKind)
		assetFile := File{
			Path:    path,
			Content: string(data),
			Tree:    flattenAST(parsedFile),
			Errors:  diagnosticsToAnnotations(parsedFile.Diagnostics()),
		}
		files = append(files, assetFile)
		return nil
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error walking directory: %v\n", err)
		return 2
	}
	if failed {
		return 1
	}

	f, err := os.Create(outputPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create output file: %v\n", err)
		return 2
	}
	defer f.Close()

	project := Project{Files: files}
	if err := Encode(project, f); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to encode project asset: %v\n", err)
		return 2
	}
	fmt.Printf("Wrote asset to %s\n", outputPath)
	return 0
}

// flattenAST walks the AST in preorder and returns a flat list of asset.Nodes.
func flattenAST(file *ast.SourceFile) []Node {
	var nodes []Node
	if file == nil {
		return nodes
	}
	visitor := &ast.NodeVisitor{}
	visitor.Visit = func(n *ast.Node) *ast.Node {
		if n == nil {
			return n
		}
		node := Node{
			Range: Range{Offset: n.Pos(), EndOffset: n.End()},
			Type:  n.Kind.String(),
		}
		nodes = append(nodes, node)
		visitor.VisitEachChild(n)
		return n
	}
	visitor.Visit(file.AsNode())
	return nodes
}

// diagnosticsToAnnotations converts parser diagnostics to asset.Annotations.
func diagnosticsToAnnotations(diags []*ast.Diagnostic) []Annotation {
	var anns []Annotation
	for _, d := range diags {
		anns = append(anns, Annotation{
			Range: Range{Offset: d.Pos(), EndOffset: d.End()},
			Text:  d.Message(),
		})
	}
	return anns
}
