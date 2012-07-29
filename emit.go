package main

import (
	"fmt"
	"go/ast"
	"go/token"
)

func (c *Compiler) emit(format string, params ...interface{}) {
	fmt.Fprintf(c.wr, format, params...)
}

func (c *Compiler) emitArrayType(node *ast.ArrayType) {
	if node.Len == nil {
		c.emit("(slice ")
	} else if _, ok := node.Len.(*ast.Ellipsis); ok {
		c.emit("(array... ")
	} else {
		c.emit("(array ")
		c.emitExpr(node.Len)
		c.emit(" ")
	}
	c.emitType(node.Elt)
	c.emit(")")
}

func (c *Compiler) emitAssignStmt(node *ast.AssignStmt) {
	c.emit("(%s ", goOpToSchemeOp(node.Tok.String()))
	if len(node.Lhs) == 1 {
		c.emitExpr(node.Lhs[0])
	} else {
		sep := "("
		for _, expr := range node.Lhs {
			c.emit(sep)
			c.emitExpr(expr)
			sep = " "
		}
		c.emit(")")
	}
	for _, expr := range node.Rhs {
		c.emit(" ")
		c.emitExpr(expr)
	}
	c.emit(")")
}

func (c *Compiler) emitBasicLit(node *ast.BasicLit) {
	switch node.Kind {
	case token.CHAR:
		c.emit("#\\%s", []byte(node.Value)[1:2])
	case token.STRING:
		// TODO newlines
		c.wr.Write([]byte(goStringToSchemeString(node)))
	default:
		// avoid printf's (MISSING):
		fmt.Fprint(c.wr, node.Value)
	}
}

func (c *Compiler) emitBinaryExpr(node *ast.BinaryExpr) {
	c.emit("(%s ", goOpToSchemeOp(node.Op.String()))
	c.emitExpr(node.X)
	c.emit(" ")
	c.emitExpr(node.Y)
	c.emit(")")
}

func (c *Compiler) emitBlockStmt(node *ast.BlockStmt) {
	if node.List == nil { return }
	for _, stmt := range node.List {
		c.emit(" ")
		c.emitStmt(stmt)
	}
}

func (c *Compiler) emitBranchStmt(node *ast.BranchStmt) {
	// (break), (continue), (goto label), (fallthrough)
	c.emit("(%s", node.Tok.String())
	if node.Label != nil {
		c.emit(" %s", node.Label.String())
	}
	c.emit(")")
}

func (c *Compiler) emitCallExpr(node *ast.CallExpr) {
	c.emit("(")
	if node.Ellipsis != 0 {
		c.emit("apply... ")
	}
	c.emitExpr(node.Fun)
	for _, arg := range node.Args {
		c.emit(" ")
		c.emitExpr(arg)
	}
	//if node.Ellipsis != 0 {
	//	c.emit(" ...")
	//}
	c.emit(")")
}

// CaseClause
// ChanDir

func (c *Compiler) emitChanType(node *ast.ChanType) {
	c.emit("(chan")
	switch node.Dir {
	case ast.RECV:
		c.emit("<-")
	case ast.SEND:
		c.emit("<-!")
	}
	c.emitType(node.Value)
	c.emit(")")
}

// CommClause

func (c *Compiler) emitComment(node *ast.Comment) {
}

func (c *Compiler) emitCommentGroup(node *ast.CommentGroup) {
}

func (c *Compiler) emitCompositeLit(node *ast.CompositeLit) {
	c.emit("#(")
	c.emitType(node.Type)
	for _, arg := range node.Elts {
		c.emit(" ")
		c.emitExpr(arg)
	}
	c.emit(")")
}

func (c *Compiler) emitDecl(node ast.Decl) {
	switch a := node.(type) {
	case *ast.GenDecl:
		c.emitGenDecl(a)
	case *ast.FuncDecl:
		c.emitFuncDecl(a)
	}
}

// DeclStmt

func (c *Compiler) emitDeferStmt(node *ast.DeferStmt) {
	c.emit("(defer ")
	c.emitCallExpr(node.Call)
	c.emit(")")
}

func (c *Compiler) emitEllipsis(node *ast.Ellipsis) {
	c.emit("... ")
	c.emitExpr(node.Elt)
}

func (c *Compiler) emitEmptyStmt(node *ast.EmptyStmt) {
	c.emit("(void)")
}

func (c *Compiler) emitExpr(node ast.Expr) {
	switch a := node.(type) {
	case *ast.BasicLit:       c.emitBasicLit(a)
	case *ast.CompositeLit:   c.emitCompositeLit(a)
	case *ast.Ellipsis:       c.emitEllipsis(a)
	case *ast.Ident:          c.emit(a.Name)

	// Expr
	case *ast.BinaryExpr:     c.emitBinaryExpr(a)
	case *ast.CallExpr:       c.emitCallExpr(a)
	case *ast.IndexExpr:      c.emitIndexExpr(a)
	case *ast.KeyValueExpr:   c.emitKeyValueExpr(a)
	case *ast.ParenExpr:      c.emitExpr(a.X)
	case *ast.SelectorExpr:   c.emitSelectorExpr(a)
	case *ast.SliceExpr:      c.emitSliceExpr(a)
	case *ast.StarExpr:       c.emitStarExpr(a)
	case *ast.TypeAssertExpr: c.emitTypeAssertExpr(a)
	case *ast.UnaryExpr:      c.emitUnaryExpr(a)

	// Type
	case *ast.ArrayType:      c.emitArrayType(a)
	case *ast.ChanType:       c.emitChanType(a)
	case *ast.FuncType:       c.emitFuncType(a)
	case *ast.InterfaceType:  c.emitInterfaceType(a)
	case *ast.MapType:        c.emitMapType(a)
	case *ast.StructType:     c.emitStructType(a)

	default:
		c.emit("<expr:%v>", node)
	}
}

// ExprStmt

func (c *Compiler) emitField(node *ast.Field) {
	if len(node.Names) == 0 {
		c.emitType(node.Type)
		return
	}
	c.emit("#(")
	for _, name := range node.Names {
		c.emit("%s ", name.String())
	}
	c.emitType(node.Type)
	c.emit(")")
}

// FieldFilter

func (c *Compiler) emitFieldList(node *ast.FieldList) {
	for _, field := range node.List {
		c.emit(" ")
		c.emitField(field)
	}
}

func (c *Compiler) emitFile(node *ast.File) {
	c.emit("(package %s ", node.Name.Name)
	ast.Walk(c, node)
	c.emit(")")
}

func (c *Compiler) emitForStmt(node *ast.ForStmt) {
	if node.Init == nil && node.Post == nil {
		c.emit("(while ")
		if node.Cond == nil {
			c.emit("#t")
		} else {
			c.emitExpr(node.Cond)
		}
		c.emit(" ")
		c.emitBlockStmt(node.Body)
		c.emit(")")
		return
	}

	c.emit("(for ")
	if node.Init == nil {
		c.emit("#f")
	} else {
		c.emitStmt(node.Init)
	}
	c.emit(" ")
	if node.Cond == nil {
		c.emit("#t")
	} else {
		c.emitExpr(node.Cond)
	}
	c.emit(" ")
	if node.Post == nil {
		c.emit("#f")
	} else {
		c.emitStmt(node.Post)
	}
	c.emit(" ")
	c.emitBlockStmt(node.Body)
	c.emit(")")
}

func (c *Compiler) emitFuncDecl(node *ast.FuncDecl) {
	// "(define-func (%s %s) %s)", name, type, body
	if node.Recv == nil {
		c.emit("(define-func (%s ", node.Name.Name)
	} else {
		c.emit("(define-func (")
		c.emitFieldList(node.Recv)
		c.emit(" %s ", node.Name.Name)
	}
	c.emitFuncType(node.Type)
	c.emit(")")
	c.emitBlockStmt(node.Body)
	c.emit(")")
}

// FuncLit

func (c *Compiler) emitFuncType(node *ast.FuncType) {
	// It is the responsibility of the caller to
	// write "func " or whatever is appropriate
	// because we have no idea at the point if this
	// is being called from a Decl/Stmt/Expr, etc.
	c.emitFieldList(node.Params)
	c.emitFuncResults(node.Results)
}

func (c *Compiler) emitFuncResults(node *ast.FieldList) {
	c.emit(" ")
	if node == nil || len(node.List) == 0 {
		c.emit("(void)")
		return
	}
	if len(node.List) == 1 {
		c.emitField(node.List[0])
		return
	}

	c.emit("(values")
	for _, field := range node.List {
		c.emit(" ")
		c.emitField(field)
	}
	c.emit(")")
}

//type emitter func (c *Compiler, node ast.Node)
//func (c *Compiler) emitList(nodes []ast.Node, emit emitter) {
//    sep := ""
//    for _, node := range nodes {
//        c.emit(sep)
//        emit(c, node)
//		sep = " "
//    }
//}

func (c *Compiler) emitGenDecl(node *ast.GenDecl) {
	if node.Tok == token.IMPORT {
		// "(import \"%s\")", path
		// "(import (as %s \"%s\"))", name, path
		// "(import (dot \"%s\"))", path
		c.emit("(import")
		for _, spec := range node.Specs {
			c.emit(" ")
			c.emitImportSpec(spec.(*ast.ImportSpec))
		}
		c.emit(")")
		return
	}

	// otherwise
	c.emit("(define-%s", node.Tok.String())
	switch node.Tok {
	case token.TYPE:
		// "(define-type %s %s)", name, type
		for _, spec := range node.Specs {
			c.emit(" ")
			c.emitTypeSpec(spec.(*ast.TypeSpec))
		}
	case token.CONST:
		// "(define-const %s)", name
		// "(define-const (= %s %s))", name, value
		// "(define-const (= #(%s %s) %s))", name, type, value
		fallthrough
	case token.VAR:
		// "(define-var (= %s %s))", name, value
		// "(define-var (= (%s) %s))", name(s), value(s)
		// "(define-var (= #(%s %s) %s))", name(s), type, value(s)
		// "(define-var #(%s %s))", name(s), type
		for _, spec := range node.Specs {
			c.emit(" ")
			c.emitValueSpec(spec.(*ast.ValueSpec))
		}
	}
	c.emit(")")
}

func (c *Compiler) emitGoStmt(node *ast.GoStmt) {
	c.emit("(go ")
	c.emitCallExpr(node.Call)
	c.emit(")")
}

// Ident

func (c *Compiler) emitIfStmt(node *ast.IfStmt) {
	c.emit("(when")
	if node.Init != nil {
		c.emit("* ")
		c.emitStmt(node.Init)
	}
	c.emit(" ")
	c.emitExpr(node.Cond)
	c.emit(" ")
	c.emitBlockStmt(node.Body)
	c.emit(")")
	// TODO: else
}

func (c *Compiler) emitImportSpec(node *ast.ImportSpec) {
	if node.Name != nil {
		if node.Name.String() == "." {
			c.emit("(dot ")
		} else {
			c.emit("(as %s ", node.Name.String())
		}
		c.emitBasicLit(node.Path)
		c.emit(")")
		return
	}
	c.emitBasicLit(node.Path)
}

// Importer

func (c *Compiler) emitIncDecStmt(node *ast.IncDecStmt) {
	c.emit("(%s ", node.Tok.String())
	c.emitExpr(node.X)
	c.emit(")")
}

func (c *Compiler) emitIndexExpr(node *ast.IndexExpr) {
	c.emit("(index ")
	c.emitExpr(node.X)
	c.emit(" ")
	c.emitExpr(node.Index)
	c.emit(")")
}

func (c *Compiler) emitInterfaceType(node *ast.InterfaceType) {
	c.emit("(interface ")
	// TODO: check if this works!
	c.emitFieldList(node.Methods)
	c.emit(")")
}

func (c *Compiler) emitKeyValueExpr(node *ast.KeyValueExpr) {
	c.emit("#:%s ", node.Key.(*ast.Ident).Name)
	c.emitExpr(node.Value)
}

func (c *Compiler) emitLabeledStmt(node *ast.LabeledStmt) {
	c.emit("(label %s ", node.Label.String())
	c.emitStmt(node.Stmt)
	c.emit(")")
}

func (c *Compiler) emitMapType(node *ast.MapType) {
	c.emit("(map-type ")
	c.emitType(node.Key)
	c.emit(" ")
	c.emitType(node.Value)
	c.emit(")")
}

// MergeMode
// Node
// ObjKind
// Object
// Package
// ParenExpr

func (c *Compiler) emitRangeStmt(node *ast.RangeStmt) {
	c.emit("(range (%s ", node.Tok.String())
	if node.Value == nil {
		c.emitExpr(node.Key)
	} else {
		c.emit("(")
		c.emitExpr(node.Key)
		c.emit(" ")
		c.emitExpr(node.Value)
		c.emit(")")
	}
	c.emit(" ")
	c.emitExpr(node.X)
	c.emit(")")
	c.emitBlockStmt(node.Body)
	c.emit(")")
}

func (c *Compiler) emitReturnStmt(node *ast.ReturnStmt) {
	c.emit("(return")
	for _, arg := range node.Results {
		c.emit(" ")
		c.emitExpr(arg)
	}
	c.emit(")")
	
}

// Scope

// helper function
func (c *Compiler) emitSelectCommClause(node *ast.CommClause) {
	if node.Comm == nil {
		c.emit("(else")
	} else {
		c.emit("(")
		c.emitStmt(node.Comm)
	}
	if node.Body != nil {
		for _, stmt := range node.Body {
			c.emit(" ")
			c.emitStmt(stmt)
		}
	}
	c.emit(")")
}

func (c *Compiler) emitSelectStmt(node *ast.SelectStmt) {
	c.emit("(comm!")
	for _, stmt := range node.Body.List {
		c.emit(" ")
		c.emitSelectCommClause(stmt.(*ast.CommClause))
	}
	c.emit(")")
}

func (c *Compiler) emitSelectorExpr(node *ast.SelectorExpr) {
	if id, ok := node.X.(*ast.Ident); ok {
		c.emit("%s.%s", id.Name, node.Sel.Name)
		return
	}
	c.emit("(dot ")
	c.emitExpr(node.X)
	c.emit(" %s)", node.Sel.Name)
}

func (c *Compiler) emitSendStmt(node *ast.SendStmt) {
	c.emit("(<-! ")
	c.emitExpr(node.Chan)
	c.emit(" ")
	c.emitExpr(node.Value)
	c.emit(")")
}

func (c *Compiler) emitSliceExpr(node *ast.SliceExpr) {
	c.emit("(index ")
	c.emitExpr(node.X)
	c.emit(" ")
	if node.Low == nil {
		c.emit("#f")
	} else {
		c.emitExpr(node.Low)
	}
	c.emit(" ")
	if node.High == nil {
		c.emit("#f")
	} else {
		c.emitExpr(node.High)
	}
	c.emit(")")
}

// Spec

func (c *Compiler) emitStarExpr(node *ast.StarExpr) {
	//if id, ok := node.X.(*ast.Ident); ok {
	//	c.emit("*%s", id.Name)
	//	return
	//}
	//if sel, ok := node.X.(*ast.SelectorExpr); ok {
	//	c.emit("*")
	//	c.emitSelectorExpr(sel)
	//	return
	//}
	c.emit("(ptr ")
	c.emitType(node.X)
	c.emit(")")
}

func (c *Compiler) emitStmt(node ast.Stmt) {
	switch a := node.(type) {

	// Stmt
	case *ast.AssignStmt:     c.emitAssignStmt(a)
	case *ast.BlockStmt:      c.emitBlockStmt(a)
	case *ast.BranchStmt:     c.emitBranchStmt(a)
	case *ast.DeclStmt:       c.emitDecl(a.Decl)
	case *ast.DeferStmt:      c.emitDeferStmt(a)
	case *ast.EmptyStmt:      c.emitEmptyStmt(a)
	case *ast.ExprStmt:       c.emitExpr(a.X)
	case *ast.ForStmt:        c.emitForStmt(a)
	case *ast.GoStmt:         c.emitGoStmt(a)
	case *ast.IfStmt:         c.emitIfStmt(a)
	case *ast.IncDecStmt:     c.emitIncDecStmt(a)
	case *ast.LabeledStmt:    c.emitLabeledStmt(a)
	case *ast.RangeStmt:      c.emitRangeStmt(a)
	case *ast.ReturnStmt:     c.emitReturnStmt(a)
	case *ast.SelectStmt:     c.emitSelectStmt(a)
	case *ast.SendStmt:       c.emitSendStmt(a)
	case *ast.SwitchStmt:     c.emitSwitchStmt(a)
	case *ast.TypeSwitchStmt: c.emitTypeSwitchStmt(a)

	default:
		c.emit("<stmt:%v>", node)
	}
}

func (c *Compiler) emitStructType(node *ast.StructType) {
	c.emit("(struct")
	c.emitFieldList(node.Fields)
	c.emit(")")
}

// helper function
func (c *Compiler) emitSwitchCaseClause(node *ast.CaseClause, cond bool) {
	if node.List == nil || len(node.List) == 0 {
		c.emit("(else ")
		for _, stmt := range node.Body {
			c.emit(" ")
			c.emitStmt(stmt)
		}
		c.emit(")")
		return
	}
	c.emit("(")
	if cond {
		c.emitExpr(node.List[0])
	} else {
		sep := "("
		for _, expr := range node.List {
			c.emit(sep)
			c.emitExpr(expr)
			sep = " "
		}
		c.emit(")")
	} 
	c.emit(" ")
	for _, stmt := range node.Body {
		c.emit(" ")
		c.emitStmt(stmt)
	}
	c.emit(")")
}

func (c *Compiler) emitSwitchStmt(node *ast.SwitchStmt) {
	cond := node.Tag == nil
	if cond {
		c.emit("(cond!")
	} else {
		c.emit("(case!")
	}
	if node.Init != nil {
		c.emit("* ")
		c.emitStmt(node.Init)
	}
	if !cond {
		c.emit(" ")
		c.emitExpr(node.Tag)
	}
	for _, stmt := range node.Body.List {
		c.emit(" ")
		c.emitSwitchCaseClause(stmt.(*ast.CaseClause), cond)
	}
	c.emit(")")
}

func (c *Compiler) emitType(node ast.Expr) {
	c.emitExpr(node)
}

func (c *Compiler) emitTypeAssertExpr(node *ast.TypeAssertExpr) {
	c.emit("(as ")
	c.emitExpr(node.X)
	c.emit(" ")
	if node.Type == nil {
		c.emit("type")
	} else {
		c.emitType(node.Type)
	}
	c.emit(")")
}

func (c *Compiler) emitTypeSpec(node *ast.TypeSpec) {
	c.emit(node.Name.Name)
	c.emit(" ")
	c.emitType(node.Type)
}

// helper function
func (c *Compiler) emitTypeSwitchCaseClause(node *ast.CaseClause) {
	c.emitSwitchCaseClause(node, false)
}

func (c *Compiler) emitTypeSwitchStmt(node *ast.TypeSwitchStmt) {
	c.emit("(type!")
	if node.Init != nil {
		c.emit("* ")
		c.emitStmt(node.Init)
	}
	c.emit(" ")
	c.emitStmt(node.Assign)
	for _, stmt := range node.Body.List {
		c.emit(" ")
		c.emitTypeSwitchCaseClause(stmt.(*ast.CaseClause))
	}
	c.emit(")")
}

func (c *Compiler) emitUnaryExpr(node *ast.UnaryExpr) {
	c.emit("(%s ", goOpToSchemeOp(node.Op.String()))
	c.emitExpr(node.X)
	c.emit(")")
}

// helper function
func (c *Compiler) emitValueNames(ids []*ast.Ident) {
	buffer := []byte{}
	for _, id := range ids {
		buffer = append(buffer, ' ')
		buffer = append(buffer, id.Name...)
	}
	if len(ids) == 1 {
		c.emit("%s", string(buffer[1:]))
	} else {
		c.emit("(%s)", string(buffer[1:]))
	}
}

func (c *Compiler) emitValueTypedNames(ids []*ast.Ident, t ast.Expr) {
	buffer := []byte{}
	for _, id := range ids {
		buffer = append(buffer, id.Name...)
		buffer = append(buffer, ' ')
	}
	c.emit("#(%s", string(buffer))
	c.emitType(t)
	c.emit(")")
}

func (c *Compiler) emitValueSpec(node *ast.ValueSpec) {
	if node.Type != nil {
		if node.Values != nil {
			// "(define-const (= #(%s %s) %s))", name, type, value
			// "(define-var (= #(%s %s) %s))", name(s), type, value(s)
			c.emit("(= ")
			c.emitValueTypedNames(node.Names, node.Type)
			for _, arg := range node.Values {
				c.emit(" ")
				c.emitExpr(arg)
			}
			c.emit(")")
		} else {
			// "(define-var #(%s %s))", name(s), type
			c.emitValueTypedNames(node.Names, node.Type)
		}
	} else {
		if node.Values != nil {
			// "(define-const (= %s %s))", name, value
			// "(define-var (= %s %s))", name, value
			// "(define-var (= (%s) %s))", name(s), value(s)
			c.emit("(= ")
			c.emitValueNames(node.Names)
			for _, arg := range node.Values {
				c.emit(" ")
				c.emitExpr(arg)
			}
			c.emit(")")
		} else {
			// "(define-const %s)", name
			c.emitValueNames(node.Names)
		}
	}
}

// Visitor