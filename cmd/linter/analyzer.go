package main

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
)

var Analyzer = &analysis.Analyzer{
	Name: "customlinter",
	Doc:  "checks panic, log.Fatal and os.Exit usage",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {

	var mainFunc *ast.FuncDecl

	if pass.Pkg.Name() == "main" {
		for _, file := range pass.Files {
			for _, decl := range file.Decls {
				if fn, ok := decl.(*ast.FuncDecl); ok && fn.Name.Name == "main" {
					mainFunc = fn
				}
			}
		}
	}

	for _, file := range pass.Files {

		ast.Inspect(file, func(n ast.Node) bool {

			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}

			if ident, ok := call.Fun.(*ast.Ident); ok {
				if ident.Name == "panic" {
					pass.Reportf(call.Pos(), "panic usage is forbidden")
				}
			}

			sel, ok := call.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}

			obj := pass.TypesInfo.Uses[sel.Sel]

			fn, ok := obj.(*types.Func)
			if !ok {
				return true
			}

			pkg := fn.Pkg()
			if pkg == nil {
				return true
			}

			if pkg.Path() == "log" && fn.Name() == "Fatal" {
				if !insideNode(mainFunc, call) {
					pass.Reportf(call.Pos(), "log.Fatal cannot be used outside main()")
				}
			}

			if pkg.Path() == "os" && fn.Name() == "Exit" {
				if !insideNode(mainFunc, call) {
					pass.Reportf(call.Pos(), "os.Exit cannot be used outside main()")
				}
			}

			return true
		})
	}

	return nil, nil
}

func insideNode(fn *ast.FuncDecl, node ast.Node) bool {
	if fn == nil {
		return false
	}

	start := fn.Pos()
	end := fn.End()

	return node.Pos() >= start && node.Pos() <= end
}
