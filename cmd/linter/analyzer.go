package main

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

var Analyzer = &analysis.Analyzer{
	Name: "customlinter",
	Doc:  "checks panic, log.Fatal and os.Exit usage",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {

	for _, file := range pass.Files {

		ast.Inspect(file, func(n ast.Node) bool {

			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}

			// panic()
			if ident, ok := call.Fun.(*ast.Ident); ok {
				if ident.Name == "panic" {
					pass.Reportf(call.Pos(), "panic usage is forbidden")
				}
			}

			// log.Fatal / os.Exit
			if sel, ok := call.Fun.(*ast.SelectorExpr); ok {

				pkgIdent, ok := sel.X.(*ast.Ident)
				if !ok {
					return true
				}

				if (pkgIdent.Name == "log" && sel.Sel.Name == "Fatal") ||
					(pkgIdent.Name == "os" && sel.Sel.Name == "Exit") {

					if !insideMain(pass, call) {
						pass.Reportf(call.Pos(), "%s.%s cannot be used outside main()", pkgIdent.Name, sel.Sel.Name)
					}
				}
			}

			return true
		})
	}

	return nil, nil
}

func insideMain(pass *analysis.Pass, node ast.Node) bool {

	for _, file := range pass.Files {

		var currentFunc *ast.FuncDecl

		ast.Inspect(file, func(n ast.Node) bool {

			switch v := n.(type) {

			case *ast.FuncDecl:
				currentFunc = v

			case *ast.CallExpr:
				if v == node {

					if currentFunc == nil {
						return false
					}

					if currentFunc.Name.Name != "main" {
						return false
					}

					if pass.Pkg.Name() != "main" {
						return false
					}

					return true
				}
			}

			return true
		})
	}

	return false
}
