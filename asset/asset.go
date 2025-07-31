package main

import (
	"encoding/binary"
	"io"
)

type Project struct {
	Files []File
}

type File struct {
	Path        string
	Content     string
	Tree        []Node
	Errors      []Annotation
	Resolutions []Resolution
}

type Node struct {
	Range
	Type string
}

type Range struct {
	Offset    int
	EndOffset int
}

// Annotation represents a message annotation (or a parser error) for a range/offset.
type Annotation struct {
	Range // EndOffset == Offset for single offset positions.
	Text  string
}

// Resolution represents a resolution of a symbol/imports in a file.
type Resolution struct {
	Range
	Definitions []Definition
}

type Definition struct {
	Path string
	Range
}

func Encode(p Project, writer io.WriteSeeker) error {
	e := encoder{
		strings: make(map[string]int),
		writer:  writer,
	}
	e.writeHeader(uint32(len(p.Files)))
	if e.writeErr != nil {
		return e.writeErr
	}

	for _, f := range p.Files {
		e.writeFile(f)
		if e.writeErr != nil {
			return e.writeErr
		}
	}

	// Write the position of the string table start at the beginning of the output.
	var buf [4]byte
	binary.LittleEndian.PutUint32(buf[:], uint32(e.written))
	if _, err := e.writer.Seek(5 /*after magic and version*/, io.SeekStart); err != nil {
		return err
	}
	_, e.writeErr = e.writer.Write(buf[:])
	if e.writeErr != nil {
		return e.writeErr
	}

	// Seek to the end of the file before writing the string table.
	if _, err := e.writer.Seek(0, io.SeekEnd); err != nil {
		return err
	}

	e.writeStringTable()
	return e.writeErr

}

// encoder takes care of encoding the asset data into a binary format.
type encoder struct {
	strings      map[string]int
	stringsSlice []string
	writer       io.WriteSeeker

	written  int   // accumulated number of bytes written.
	writeErr error // accumulated error, so we don't check every write for errors.
}

const (
	magic                = 0xde
	assetEncodingVersion = uint32(1)
)

func (e *encoder) writeHeader(numFiles uint32) {
	e.writeByte(magic)
	e.writeUint32(assetEncodingVersion)
	// Reserved for the offset position of the strings table.
	// Filled at the end of the encoding process.
	e.writeUint32(0)
	e.writeUint32(numFiles)
}

func (e *encoder) writeFile(f File) {
	e.writeString(f.Path)
	e.writeString(f.Content)
	e.writeUint32(uint32(len(f.Tree)))

	for _, n := range f.Tree {
		e.writeNode(n)
	}
	e.writeUint32(uint32(len(f.Errors)))
	for _, a := range f.Errors {
		e.writeAnnotation(a)
	}
	e.writeUint32(uint32(len(f.Resolutions)))
	for _, r := range f.Resolutions {
		e.writeResolution(r)
	}
}

// writeNode adds a node to the expected AST.
func (e *encoder) writeNode(n Node) {
	e.writeUint32(uint32(n.Offset))
	e.writeUint32(uint32(n.EndOffset))
	e.writeString(n.Type)
}

// writeAnnotation adds a message annotation (or a parser error) for a range/offset.
func (e *encoder) writeAnnotation(ann Annotation) {
	e.writeUint32(uint32(ann.Offset))
	e.writeUint32(uint32(ann.EndOffset))
	e.writeString(ann.Text)
}

// writeResolution adds a resolution of a symbol in a file.
func (e *encoder) writeResolution(r Resolution) {
	e.writeUint32(uint32(r.Offset))
	e.writeUint32(uint32(r.EndOffset))
	e.writeUint32(uint32(len(r.Definitions)))
	for _, d := range r.Definitions {
		e.writeDefinition(d)
	}
}

// writeDefinition adds a definition of a symbol in a file.
func (e *encoder) writeDefinition(d Definition) {
	e.writeString(d.Path)
	e.writeUint32(uint32(d.Offset))
	e.writeUint32(uint32(d.EndOffset))
}

// writeStringTable writes the string table at the end of the asset.
func (e *encoder) writeStringTable() {
	e.writeUint32(uint32(len(e.stringsSlice)))
	for _, s := range e.stringsSlice {
		if e.writeErr != nil {
			return
		}

		e.writeUint32(uint32(len(s)))
		n, err := e.writer.Write([]byte(s))
		e.written += n
		e.writeErr = err
	}
}

func (e *encoder) writeByte(b byte) {
	if e.writeErr != nil {
		return
	}
	n, err := e.writer.Write([]byte{b})
	e.written += n
	e.writeErr = err
}

func (e *encoder) writeString(content string) {
	pos := e.addString(content)
	e.writeUint32(uint32(pos))
}

func (e *encoder) writeUint32(v uint32) {
	if e.writeErr != nil {
		return
	}
	var bytes [4]byte
	binary.LittleEndian.PutUint32(bytes[:], v)
	n, err := e.writer.Write(bytes[:])
	e.written += n
	e.writeErr = err
}

// addString adds a string to the string table and returns its index. It does not write the string
// to the asset, as that happens when writing the string table at the end.
func (e *encoder) addString(s string) uint32 {
	if idx, ok := e.strings[s]; ok {
		return uint32(idx)
	}
	e.strings[s] = len(e.stringsSlice)
	e.stringsSlice = append(e.stringsSlice, s)
	return uint32(len(e.stringsSlice) - 1)
}
