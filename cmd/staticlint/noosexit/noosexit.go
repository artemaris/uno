// Package noosexit предоставляет анализатор, который запрещает прямой вызов os.Exit в функции main пакета main.
//
// Анализатор проверяет исходный код на наличие прямых вызовов os.Exit в функции main
// основного пакета и выдает ошибку, если такие вызовы найдены.
//
// Это помогает избежать неконтролируемого завершения программы и способствует
// более чистому коду с правильной обработкой ошибок.
package noosexit

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// Analyzer - анализатор, запрещающий использовать прямой вызов os.Exit в функции main пакета main.
var Analyzer = &analysis.Analyzer{
	Name:     "noosexit",
	Doc:      "запрещает прямой вызов os.Exit в функции main пакета main",
	Run:      run,
	Requires: []*analysis.Analyzer{inspect.Analyzer},
}

func run(pass *analysis.Pass) (interface{}, error) {
	// Получаем инспектор AST
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	// Фильтр для функций
	nodeFilter := []ast.Node{
		(*ast.FuncDecl)(nil),
	}

	// Проходим по всем функциям
	inspect.Preorder(nodeFilter, func(n ast.Node) {
		funcDecl := n.(*ast.FuncDecl)

		// Проверяем, что это функция main в пакете main
		if !isMainFunctionInMainPackage(pass, funcDecl) {
			return
		}

		// Проверяем тело функции на наличие os.Exit
		checkForOsExit(pass, funcDecl.Body)
	})

	return nil, nil
}

// isMainFunctionInMainPackage проверяет, является ли функция main в пакете main
func isMainFunctionInMainPackage(pass *analysis.Pass, funcDecl *ast.FuncDecl) bool {
	// Проверяем имя пакета
	if pass.Pkg.Name() != "main" {
		return false
	}

	// Проверяем имя функции
	if funcDecl.Name.Name != "main" {
		return false
	}

	// Проверяем, что функция не имеет параметров и возвращаемых значений
	if funcDecl.Type.Params != nil && len(funcDecl.Type.Params.List) > 0 {
		return false
	}

	if funcDecl.Type.Results != nil && len(funcDecl.Type.Results.List) > 0 {
		return false
	}

	return true
}

// checkForOsExit рекурсивно проверяет блок кода на наличие вызовов os.Exit
func checkForOsExit(pass *analysis.Pass, block *ast.BlockStmt) {
	if block == nil {
		return
	}

	ast.Inspect(block, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.CallExpr:
			if isOsExitCall(pass, node) {
				pass.Reportf(node.Pos(), "прямой вызов os.Exit в функции main запрещен")
			}
		}
		return true
	})
}

// isOsExitCall проверяет, является ли вызов вызовом os.Exit
func isOsExitCall(pass *analysis.Pass, call *ast.CallExpr) bool {
	// Проверяем, что это селектор (package.Function)
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}

	// Проверяем, что селектор относится к пакету os
	ident, ok := sel.X.(*ast.Ident)
	if !ok {
		return false
	}

	// Получаем информацию о типе
	obj := pass.TypesInfo.Uses[ident]
	if obj == nil {
		return false
	}

	// Проверяем, что это пакет os
	pkg, ok := obj.(*types.PkgName)
	if !ok {
		return false
	}

	if pkg.Imported().Path() != "os" {
		return false
	}

	// Проверяем, что вызывается функция Exit
	return sel.Sel.Name == "Exit"
}
