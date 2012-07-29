package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"os"
	"strings"
)

type Buffer struct {
}

type Compiler struct {
//	br bufio.Reader
//	bw bufio.Writer
//	rd io.Reader
//	rw io.ReadWriter
	wr io.Writer
}

func NewBuffer() *Buffer {
	return &Buffer{}
}

func NewCompiler() *Compiler {
	return &Compiler{}
}

func (c *Compiler) Compile(rd io.Reader, wr io.Writer) (err error) {
	c.wr = wr
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "", rd, 0)
	if err != nil {
		return err
	}
	c.emitFile(file)
	if f, ok := c.wr.(io.Closer); ok {
		err = f.Close()
	}
	return
}

func (c *Compiler) compileFile(filename string) error {
	return c.compileFileTo(filename, os.Stdout)
}

func (c *Compiler) compileFileTo(filename string, wr io.Writer) error {
	rd, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	return c.Compile(rd, wr)
}

func (c *Compiler) compileString(input string) error {
	return c.Compile(strings.NewReader(input), os.Stdout)
}

func (c *Compiler) Visit(node ast.Node) (w ast.Visitor) {
	//fmt.Printf("\n- %v\n", node)
	switch a := node.(type) {
	case *ast.GenDecl:
		c.emitGenDecl(a)
		return nil
	case *ast.FuncDecl:
		c.emitFuncDecl(a)
		return nil
	}
	return c
}

func goOpToSchemeOp(name string) string {
	var table = map[string]string{
		"!": "not",
		"&": "bitwise-and",
		"&&": "and",
		"&=": "bitwise-and=",
		"&^": "bitwise-but",
		"&^=": "bitwise-but=",
		"^": "bitwise-xor",
		"^=": "bitwise-xor=",
		"|": "bitwise-or",
		"|=": "bitwise-or=",
		"||": "or",
	}
	if table[name] != "" {
		return table[name]
	}
	return name
}

func goStringToSchemeString(node *ast.BasicLit) string {
	//internalBuf := make([]byte, 1024)
	//buf := bytes.NewBuffer(internalBuf)
	//err := printer.Fprint(buf, token.NewFileSet(), node)
	//if err != nil {
	//	panic(err)
	//}
	//return strings.Replace(node.Value, "\n", "\\n", 1)
	return node.Value
}
