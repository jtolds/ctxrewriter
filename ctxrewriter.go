// package ctxrewriter rewrites go source by adding a `ctx context.Context`
// argument to the beginning of every function definition, and by adding a
// `ctx` argument to the beginning of every function call.
package ctxrewriter

import (
	"bytes"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
)

const (
	ctxVariable = "ctx"
)

func rewriteExprs(exprs []ast.Expr) []ast.Expr {
	if exprs == nil {
		return nil
	}
	new_exprs := make([]ast.Expr, 0, len(exprs))
	for _, expr := range exprs {
		new_exprs = append(new_exprs, rewrite(expr).(ast.Expr))
	}
	return new_exprs
}

func rewriteStmts(stmts []ast.Stmt) []ast.Stmt {
	if stmts == nil {
		return nil
	}
	new_stmts := make([]ast.Stmt, 0, len(stmts))
	for _, stmt := range stmts {
		new_stmts = append(new_stmts, rewrite(stmt).(ast.Stmt))
	}
	return new_stmts
}

func rewrite(node ast.Node) ast.Node {
	switch v := node.(type) {
	default:
		panic(node)
	case *ast.ImportSpec, *ast.BasicLit, *ast.Ident,
		*ast.BranchStmt, *ast.EmptyStmt:
		return node

	case *ast.ArrayType:
		c := *v
		if c.Len != nil {
			c.Len = rewrite(c.Len).(ast.Expr)
		}
		c.Elt = rewrite(c.Elt).(ast.Expr)
		return &c
	case *ast.AssignStmt:
		c := *v
		c.Lhs = rewriteExprs(c.Lhs)
		c.Rhs = rewriteExprs(c.Rhs)
		return &c
	case *ast.BinaryExpr:
		c := *v
		c.X = rewrite(c.X).(ast.Expr)
		c.Y = rewrite(c.Y).(ast.Expr)
		return &c
	case *ast.BlockStmt:
		c := *v
		c.List = rewriteStmts(c.List)
		return &c
	case *ast.CallExpr:
		c := *v
		c.Fun = rewrite(c.Fun).(ast.Expr)
		c.Args = append([]ast.Expr{ast.NewIdent(ctxVariable)},
			rewriteExprs(c.Args)...)
		return &c
	case *ast.CaseClause:
		c := *v
		c.List = rewriteExprs(c.List)
		c.Body = rewriteStmts(c.Body)
		return &c
	case *ast.ChanType:
		c := *v
		c.Value = rewrite(c.Value).(ast.Expr)
		return &c
	case *ast.CommClause:
		c := *v
		if c.Comm != nil {
			c.Comm = rewrite(c.Comm).(ast.Stmt)
		}
		c.Body = rewriteStmts(c.Body)
		return &c
	case *ast.CompositeLit:
		c := *v
		if c.Type != nil {
			c.Type = rewrite(c.Type).(ast.Expr)
		}
		c.Elts = rewriteExprs(c.Elts)
		return &c
	case *ast.DeclStmt:
		c := *v
		c.Decl = rewrite(c.Decl).(ast.Decl)
		return &c
	case *ast.DeferStmt:
		c := *v
		c.Call = rewrite(c.Call).(*ast.CallExpr)
		return &c
	case *ast.Ellipsis:
		c := *v
		if c.Elt != nil {
			c.Elt = rewrite(c.Elt).(ast.Expr)
		}
		return &c
	case *ast.ExprStmt:
		c := *v
		c.X = rewrite(c.X).(ast.Expr)
		return &c
	case *ast.Field:
		c := *v
		c.Type = rewrite(c.Type).(ast.Expr)
		return &c
	case *ast.FieldList:
		c := *v
		if c.List != nil {
			new_list := make([]*ast.Field, 0, len(c.List))
			for _, field := range c.List {
				new_list = append(new_list, rewrite(field).(*ast.Field))
			}
			c.List = new_list
		}
		return &c
	case *ast.File:
		c := *v
		new_decls := make([]ast.Decl, 0, len(c.Decls)+1)
		new_decls = append(new_decls, &ast.GenDecl{
			Tok: token.IMPORT,
			Specs: []ast.Spec{
				&ast.ImportSpec{Path: &ast.BasicLit{
					Value: `"golang.org/x/net/context"`}}}})
		for _, decl := range c.Decls {
			new_decls = append(new_decls, rewrite(decl).(ast.Decl))
		}
		c.Decls = new_decls
		return &c
	case *ast.ForStmt:
		c := *v
		if c.Init != nil {
			c.Init = rewrite(c.Init).(ast.Stmt)
		}
		if c.Cond != nil {
			c.Cond = rewrite(c.Cond).(ast.Expr)
		}
		if c.Post != nil {
			c.Post = rewrite(c.Post).(ast.Stmt)
		}
		if c.Body != nil {
			c.Body = rewrite(c.Body).(*ast.BlockStmt)
		}
		return &c
	case *ast.FuncDecl:
		c := *v
		if c.Body != nil {
			c.Body = rewrite(c.Body).(*ast.BlockStmt)
		}
		c.Type = rewrite(c.Type).(*ast.FuncType)
		return &c
	case *ast.FuncLit:
		c := *v
		c.Type = rewrite(c.Type).(*ast.FuncType)
		if c.Body != nil {
			c.Body = rewrite(c.Body).(*ast.BlockStmt)
		}
		return &c
	case *ast.FuncType:
		c := *v
		c.Params = rewrite(c.Params).(*ast.FieldList)
		c.Params.List = append([]*ast.Field{{
			Names: []*ast.Ident{ast.NewIdent(ctxVariable)},
			Type: &ast.SelectorExpr{
				X:   ast.NewIdent("context"),
				Sel: ast.NewIdent("Context")}}}, c.Params.List...)
		if c.Results != nil {
			c.Results = rewrite(c.Results).(*ast.FieldList)
		}
		return &c
	case *ast.GenDecl:
		c := *v
		if c.Specs != nil {
			new_specs := make([]ast.Spec, 0, len(c.Specs))
			for _, spec := range c.Specs {
				new_specs = append(new_specs, rewrite(spec).(ast.Spec))
			}
			c.Specs = new_specs
		}
		return &c
	case *ast.GoStmt:
		c := *v
		c.Call = rewrite(c.Call).(*ast.CallExpr)
		return &c
	case *ast.IfStmt:
		c := *v
		if c.Init != nil {
			c.Init = rewrite(c.Init).(ast.Stmt)
		}
		if c.Cond != nil {
			c.Cond = rewrite(c.Cond).(ast.Expr)
		}
		if c.Body != nil {
			c.Body = rewrite(c.Body).(*ast.BlockStmt)
		}
		if c.Else != nil {
			c.Else = rewrite(c.Else).(ast.Stmt)
		}
		return &c
	case *ast.IncDecStmt:
		c := *v
		c.X = rewrite(c.X).(ast.Expr)
		return &c
	case *ast.IndexExpr:
		c := *v
		c.X = rewrite(c.X).(ast.Expr)
		c.Index = rewrite(c.Index).(ast.Expr)
		return &c
	case *ast.InterfaceType:
		c := *v
		c.Methods = rewrite(c.Methods).(*ast.FieldList)
		return &c
	case *ast.KeyValueExpr:
		c := *v
		c.Key = rewrite(c.Key).(ast.Expr)
		c.Value = rewrite(c.Value).(ast.Expr)
		return &c
	case *ast.LabeledStmt:
		c := *v
		c.Stmt = rewrite(c.Stmt).(ast.Stmt)
		return &c
	case *ast.MapType:
		c := *v
		c.Key = rewrite(c.Key).(ast.Expr)
		c.Value = rewrite(c.Value).(ast.Expr)
		return &c
	case *ast.ParenExpr:
		c := *v
		c.X = rewrite(c.X).(ast.Expr)
		return &c
	case *ast.RangeStmt:
		c := *v
		if c.Key != nil {
			c.Key = rewrite(c.Key).(ast.Expr)
		}
		if c.Value != nil {
			c.Value = rewrite(c.Value).(ast.Expr)
		}
		if c.X != nil {
			c.X = rewrite(c.X).(ast.Expr)
		}
		if c.Body != nil {
			c.Body = rewrite(c.Body).(*ast.BlockStmt)
		}
		return &c
	case *ast.ReturnStmt:
		c := *v
		c.Results = rewriteExprs(c.Results)
		return &c
	case *ast.SelectStmt:
		c := *v
		if c.Body != nil {
			c.Body = rewrite(c.Body).(*ast.BlockStmt)
		}
		return &c
	case *ast.SelectorExpr:
		c := *v
		c.X = rewrite(c.X).(ast.Expr)
		return &c
	case *ast.SendStmt:
		c := *v
		if c.Chan != nil {
			c.Chan = rewrite(c.Chan).(ast.Expr)
		}
		if c.Value != nil {
			c.Value = rewrite(c.Value).(ast.Expr)
		}
		return &c
	case *ast.SliceExpr:
		c := *v
		c.X = rewrite(c.X).(ast.Expr)
		if c.Low != nil {
			c.Low = rewrite(c.Low).(ast.Expr)
		}
		if c.High != nil {
			c.High = rewrite(c.High).(ast.Expr)
		}
		if c.Max != nil {
			c.Max = rewrite(c.Max).(ast.Expr)
		}
		return &c
	case *ast.StarExpr:
		c := *v
		c.X = rewrite(c.X).(ast.Expr)
		return &c
	case *ast.StructType:
		c := *v
		c.Fields = rewrite(c.Fields).(*ast.FieldList)
		return &c
	case *ast.SwitchStmt:
		c := *v
		if c.Init != nil {
			c.Init = rewrite(c.Init).(ast.Stmt)
		}
		if c.Tag != nil {
			c.Tag = rewrite(c.Tag).(ast.Expr)
		}
		if c.Body != nil {
			c.Body = rewrite(c.Body).(*ast.BlockStmt)
		}
		return &c
	case *ast.TypeAssertExpr:
		c := *v
		if c.X != nil {
			c.X = rewrite(c.X).(ast.Expr)
		}
		if c.Type != nil {
			c.Type = rewrite(c.Type).(ast.Expr)
		}
		return &c
	case *ast.TypeSpec:
		c := *v
		c.Type = rewrite(c.Type).(ast.Expr)
		return &c
	case *ast.TypeSwitchStmt:
		c := *v
		if c.Init != nil {
			c.Init = rewrite(c.Init).(ast.Stmt)
		}
		if c.Assign != nil {
			c.Assign = rewrite(c.Assign).(ast.Stmt)
		}
		if c.Body != nil {
			c.Body = rewrite(c.Body).(*ast.BlockStmt)
		}
		return &c
	case *ast.UnaryExpr:
		c := *v
		c.X = rewrite(c.X).(ast.Expr)
		return &c
	case *ast.ValueSpec:
		c := *v
		c.Values = rewriteExprs(c.Values)
		return &c
	}
}

func Process(source []byte) ([]byte, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "go.go", source, parser.ParseComments)
	if err != nil {
		return nil, err
	}
	var out bytes.Buffer
	err = printer.Fprint(&out, fset, rewrite(f))
	return out.Bytes(), err
}

func ProcessFile(filename string, inplace bool) error {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		return err
	}
	out := os.Stdout
	if inplace {
		fh, err := os.Create(filename)
		if err != nil {
			return err
		}
		defer fh.Close()
		out = fh
	}
	return printer.Fprint(out, fset, rewrite(f))
}
