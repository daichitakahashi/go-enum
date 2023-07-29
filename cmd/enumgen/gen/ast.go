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

func visitorSpec(r *namingRegistry, enumIdent string, members []*ast.Ident) *ast.TypeSpec {
	// ExampleVisitor interface {
	// 	VisitA(e A)
	// 	VisitB(e B)
	// }

	methodList := make([]*ast.Field, 0, len(members))
	for _, m := range members {
		methodList = append(methodList, &ast.Field{
			Names: []*ast.Ident{
				ast.NewIdent(r.visitMethodName(enumIdent, m.String())),
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
		Name: ast.NewIdent(r.visitorTypeName(enumIdent)),
		Type: &ast.InterfaceType{
			Methods: &ast.FieldList{
				List: methodList,
			},
		},
	}
}

func enumSpec(r *namingRegistry, enumIdent string) *ast.TypeSpec {
	// ExampleEnum interface {
	// 	Accept(v ExampleVisitor)
	// }

	return &ast.TypeSpec{
		Name: ast.NewIdent(fmt.Sprintf("%sEnum", enumIdent)),
		Type: &ast.InterfaceType{
			Methods: &ast.FieldList{
				List: []*ast.Field{
					{
						Names: []*ast.Ident{
							ast.NewIdent(r.acceptMethodName(enumIdent)),
						},
						Type: &ast.FuncType{
							Params: &ast.FieldList{
								List: []*ast.Field{
									{
										Names: []*ast.Ident{
											ast.NewIdent("v"),
										},
										Type: ast.NewIdent(r.visitorTypeName(enumIdent)),
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

func acceptImpl(r *namingRegistry, enumIdent string, member *ast.Ident) *ast.FuncDecl {
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
					Type: member,
				},
			},
		},
		Name: ast.NewIdent(r.acceptMethodName(enumIdent)),
		Type: &ast.FuncType{
			Params: &ast.FieldList{
				List: []*ast.Field{
					{
						Names: []*ast.Ident{
							visitor,
						},
						Type: ast.NewIdent(r.visitorTypeName(enumIdent)),
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
							Sel: ast.NewIdent(r.visitMethodName(enumIdent, member.String())),
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

func typeCheckDecl(enumIdent string, members []*ast.Ident) *ast.GenDecl {
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
							Elt: ast.NewIdent(fmt.Sprintf("%sEnum", enumIdent)),
						},
						Elts: enumMembers,
					},
				},
			},
		},
	}
}
