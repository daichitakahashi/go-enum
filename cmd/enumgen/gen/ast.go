package gen

import (
	"fmt"
	"go/ast"
	"go/token"
)

func importSpecs(s []*ast.ImportSpec) *ast.GenDecl {
	decl := &ast.GenDecl{
		Tok:   token.IMPORT,
		Specs: make([]ast.Spec, 0, len(s)),
	}
	for _, is := range s {
		decl.Specs = append(decl.Specs, is)
	}
	return decl
}

func visitorSpec(visitorTypeName string, members []*ast.Ident) *ast.TypeSpec {
	// ExampleVisitor interface {
	// 	VisitA(e A)
	// 	VisitB(e B)
	// }

	methodList := make([]*ast.Field, 0, len(members))
	for _, m := range members {
		methodList = append(methodList, &ast.Field{
			Names: []*ast.Ident{
				ast.NewIdent(fmt.Sprintf("Visit%s", m)),
			},
			Type: &ast.FuncType{
				Params: &ast.FieldList{
					List: []*ast.Field{
						{
							Names: []*ast.Ident{
								ast.NewIdent("e"),
							},
							Type: m,
						},
					},
				},
			},
		})
	}

	return &ast.TypeSpec{
		Name: ast.NewIdent(visitorTypeName),
		Type: &ast.InterfaceType{
			Methods: &ast.FieldList{
				List: methodList,
			},
		},
	}
}

func enumSpec(enumTypeName, visitorTypeName string) *ast.TypeSpec {
	// ExampleEnum interface {
	// 	Accept(v ExampleVisitor)
	// }

	return &ast.TypeSpec{
		Name: ast.NewIdent(enumTypeName),
		Type: &ast.InterfaceType{
			Methods: &ast.FieldList{
				List: []*ast.Field{
					{
						Names: []*ast.Ident{
							ast.NewIdent("Accept"),
						},
						Type: &ast.FuncType{
							Params: &ast.FieldList{
								List: []*ast.Field{
									{
										Names: []*ast.Ident{
											ast.NewIdent("v"),
										},
										Type: ast.NewIdent(visitorTypeName),
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func acceptImpl(ident *ast.Ident, visitorTypeName string) *ast.FuncDecl {
	// func (e A) Accept(v ExampleVisitor) {
	// 	v.VisitA(e)
	// }

	enumVal := ast.NewIdent("e")
	visitor := ast.NewIdent("v")

	return &ast.FuncDecl{
		Recv: &ast.FieldList{
			List: []*ast.Field{
				{
					Names: []*ast.Ident{
						enumVal,
					},
					Type: ident,
				},
			},
		},
		Name: ast.NewIdent("Accept"),
		Type: &ast.FuncType{
			Params: &ast.FieldList{
				List: []*ast.Field{
					{
						Names: []*ast.Ident{
							visitor,
						},
						Type: ast.NewIdent(visitorTypeName),
					},
				},
			},
		},
		Body: &ast.BlockStmt{
			List: []ast.Stmt{
				&ast.ExprStmt{
					X: &ast.CallExpr{
						Fun: &ast.SelectorExpr{
							X:   visitor,
							Sel: ast.NewIdent(fmt.Sprintf("Visit%s", ident)),
						},
						Args: []ast.Expr{
							enumVal,
						},
					},
				},
			},
		},
	}
}

func typeCheckDecl(enumTypeName string, members []*ast.Ident) *ast.GenDecl {
	// var _ = []ExampleEnum{
	// 	A{},
	// 	B{},
	// }

	enumMembers := make([]ast.Expr, 0, len(members))
	for _, m := range members {
		enumMembers = append(enumMembers, &ast.CompositeLit{
			Type: m,
		})
	}

	return &ast.GenDecl{
		Tok: token.VAR,
		Specs: []ast.Spec{
			&ast.ValueSpec{
				Names: []*ast.Ident{
					ast.NewIdent("_"),
				},
				Values: []ast.Expr{
					&ast.CompositeLit{
						Type: &ast.ArrayType{
							Elt: ast.NewIdent(enumTypeName),
						},
						Elts: enumMembers,
					},
				},
			},
		},
	}
}
