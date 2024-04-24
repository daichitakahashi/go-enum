package gen

import "go/ast"

// Extract type parameter T from `pkgName.symbol[T]`.
func extractSymbolExpr(pkgName, symbol string, expr *ast.IndexExpr) (ast.Expr, bool) {
	if sel, ok := expr.X.(*ast.SelectorExpr); ok {
		if pkg, ok := sel.X.(*ast.Ident); ok {
			if pkg.String() != pkgName {
				return nil, false
			}
		}
		if sel.Sel.String() != symbol {
			return nil, false
		}
	}
	return expr.Index, true
}

// Extract interface used as enum identifier T from struct (`pkgName.MemberOf[T]`).
func findEnumIdentFromFields(pkgName string, s *ast.StructType) (ast.Expr, bool) {
	for _, f := range s.Fields.List {
		if len(f.Names) == 0 {
			if expr, ok := f.Type.(*ast.IndexExpr); ok {
				enumIdent, ok := extractSymbolExpr(pkgName, enumSymbol, expr)
				if ok {
					return enumIdent, true
				}
			}
		}
	}
	return nil, false
}

// Extract return type of visitor method from enum identifier interface (`pkgName.VisitorReturns[T]`)
func findVisitorReturnsFromFields(pkgName string, i *ast.InterfaceType) (ast.Expr, bool) {
	for _, f := range i.Methods.List {
		if len(f.Names) == 0 {
			if expr, ok := f.Type.(*ast.IndexExpr); ok {
				visitorReturnIdent, ok := extractSymbolExpr(pkgName, visitorReturnsSymbol, expr)
				if ok {
					return visitorReturnIdent, true
				}
			}
		}
	}
	return nil, false
}

type enumMemberDefinition struct {
	ident     *ast.Ident
	enumIdent ast.Expr
}

func extractEnumMemberDefinition(enumPackage string, spec *ast.TypeSpec) (*enumMemberDefinition, bool) {
	switch s := spec.Type.(type) {
	case *ast.StructType:
		// type A struct { enumPackage.MemberOf[Ident] }
		if ident, ok := findEnumIdentFromFields(enumPackage, s); ok {
			return &enumMemberDefinition{
				ident:     spec.Name,
				enumIdent: ident,
			}, true
		}
	case *ast.IndexExpr:
		// type A enumPackage.MemberOf[Ident]
		if ident, ok := extractSymbolExpr(enumPackage, enumSymbol, s); ok {
			return &enumMemberDefinition{
				ident:     spec.Name,
				enumIdent: ident,
			}, true
		}
	}
	return nil, false
}

type enumIdentDefinition struct {
	ident              *ast.Ident
	visitorReturnIdent ast.Expr
}

func extractEnumIdentDefinition(enumPackage string, spec *ast.TypeSpec) (*enumIdentDefinition, bool) {
	switch s := spec.Type.(type) {
	case *ast.InterfaceType:
		// type A interface { enumPackage.VisitorReturns[Ident] }
		if ident, ok := findVisitorReturnsFromFields(enumPackage, s); ok {
			return &enumIdentDefinition{
				ident:              spec.Name,
				visitorReturnIdent: ident,
			}, true
		}
	}
	return nil, false
}
