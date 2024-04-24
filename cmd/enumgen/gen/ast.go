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

func visitorSpec(r *namingRegistry, enumIdent string, members []*ast.Ident, visitorReturnIdent ast.Expr) *ast.TypeSpec {
	// ExampleVisitor interface {
	// 	VisitA(e A)
	// 	VisitB(e B)
	// }

	// Visit(e A) visitReturnIdent
	var results *ast.FieldList
	if visitorReturnIdent != nil {
		results = &ast.FieldList{
			List: []*ast.Field{
				{
					Type: visitorReturnIdent,
				},
			},
		}
	}

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
				Results: results,
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

func enumSpec(r *namingRegistry, enumIdent string, visitorReturnIdent ast.Expr) *ast.TypeSpec {
	// ExampleEnum interface {
	// 	Accept(v ExampleVisitor)
	// }

	// Accept(v Visitor) visitorReturnIdent
	var results *ast.FieldList
	if visitorReturnIdent != nil {
		results = &ast.FieldList{
			List: []*ast.Field{
				{
					Type: visitorReturnIdent,
				},
			},
		}
	}

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
							Results: results,
						},
					},
				},
			},
		},
	}
}

func acceptImpl(r *namingRegistry, enumIdent string, member *ast.Ident, visitorReturnIdent ast.Expr) *ast.FuncDecl {
	var (
		stmts   []ast.Stmt
		results *ast.FieldList
		enumVal = ast.NewIdent("e")
		visitor = ast.NewIdent("v")
	)

	if visitorReturnIdent != nil {
		// func (e A) Accept(v ExampleVisitor) visitorReturnIdent {
		// 	return v.VisitA(e)
		// }
		results = &ast.FieldList{
			List: []*ast.Field{
				{
					Type: visitorReturnIdent,
				},
			},
		}
		stmts = []ast.Stmt{
			&ast.ReturnStmt{
				Results: []ast.Expr{
					&ast.CallExpr{
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
		}
	} else {
		// func (e A) Accept(v ExampleVisitor) {
		// 	v.VisitA(e)
		// }
		stmts = []ast.Stmt{
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
		}
	}

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
			Results: results,
		},
		Body: &ast.BlockStmt{
			List: stmts,
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
