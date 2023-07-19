package gen

import "go/ast"

func findEnumIdentFromFields(pkgName string, s *ast.StructType) (ast.Expr, bool) {
	for _, f := range s.Fields.List {
		if len(f.Names) == 0 {
			if expr, ok := f.Type.(*ast.IndexExpr); ok {
				enumIdent, ok := extractEnumIdent(pkgName, expr)
				if ok {
					return enumIdent, true
				}
			}
		}
	}
	return nil, false
}

func extractEnumIdent(pkgName string, expr *ast.IndexExpr) (ast.Expr, bool) {
	if sel, ok := expr.X.(*ast.SelectorExpr); ok {
		if pkg, ok := sel.X.(*ast.Ident); ok {
			if pkg.String() != pkgName {
				return nil, false
			}
		}
		if sel.Sel.String() != enumSymbol {
			return nil, false
		}
	}
	return expr.Index, true
}

type enumDefinition struct {
	ident     *ast.Ident
	enumIdent ast.Expr
}

func checkEnumDefinition(enumPackage string, spec *ast.TypeSpec) (*enumDefinition, bool) {
	switch s := spec.Type.(type) {
	case *ast.StructType:
		// type A struct { enum.Value[Ident] }
		if ident, ok := findEnumIdentFromFields(enumPackage, s); ok {
			return &enumDefinition{
				ident:     spec.Name,
				enumIdent: ident,
			}, true
		}
	case *ast.IndexExpr:
		// type A enum.Value[Ident]
		if ident, ok := extractEnumIdent(enumPackage, s); ok {
			return &enumDefinition{
				ident:     spec.Name,
				enumIdent: ident,
			}, true
		}
	}
	return nil, false
}
