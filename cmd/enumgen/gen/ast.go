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

func visitorImplSpec(r *namingRegistry, enumIdent string, members []*ast.Ident, visitorReturnIdent ast.Expr) *ast.GenDecl {
	// type __ExampleVisitor struct {
	// 	__VisitA func(A) error
	// 	__VisitB func(B) error
	// }

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

	var fields []*ast.Field
	for _, m := range members {
		fields = append(fields,
			&ast.Field{
				Names: []*ast.Ident{
					ast.NewIdent(fmt.Sprintf("__%s", r.visitMethodName(enumIdent, m.String()))),
				},
				Type: &ast.FuncType{
					Params: &ast.FieldList{
						List: []*ast.Field{
							{
								Type: m,
							},
						},
					},
					Results: results,
				},
			})
	}

	return &ast.GenDecl{
		Tok: token.TYPE,
		Specs: []ast.Spec{
			&ast.TypeSpec{
				Name: ast.NewIdent(fmt.Sprintf("__%s", r.visitorTypeName(enumIdent))),
				Type: &ast.StructType{
					Fields: &ast.FieldList{
						List: fields,
					},
				},
			},
		},
	}
}

func visitorFactoryImpl(registry *namingRegistry, enumIdent, visitorFactoryName string, members []*ast.Ident, visitorReturnIdent ast.Expr) *ast.FuncDecl {
	// func NewExampleVisitor(
	// 	__VisitA func(e A) error,
	// 	__VisitB func(e B) error,
	// ) ExampleVisitor {
	// 	return &__ExampleVisitor{
	// 		__VisitA: __VisitA,
	// 		__VisitB: __VisitB,
	// 	}
	// }

	var (
		visitorFactory  = ast.NewIdent(visitorFactoryName)
		visitorType     = ast.NewIdent(registry.visitorTypeName(enumIdent))
		visitorImplType = ast.NewIdent(fmt.Sprintf("__%s", registry.visitorTypeName(enumIdent)))
		visitorReturns  *ast.FieldList

		args        []*ast.Field
		compositeKv []ast.Expr
	)

	if visitorReturnIdent != nil {
		visitorReturns = &ast.FieldList{
			List: []*ast.Field{
				{
					Type: visitorReturnIdent,
				},
			},
		}
	}

	for _, m := range members {
		fieldName := ast.NewIdent(fmt.Sprintf("__%s", registry.visitMethodName(enumIdent, m.String())))
		args = append(args, &ast.Field{
			Names: []*ast.Ident{
				fieldName,
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
				Results: visitorReturns,
			},
		})
		compositeKv = append(compositeKv, &ast.KeyValueExpr{
			Key:   fieldName,
			Value: fieldName,
		})
	}

	return &ast.FuncDecl{
		Name: visitorFactory,
		Type: &ast.FuncType{
			Params: &ast.FieldList{
				List: args,
			},
			Results: &ast.FieldList{
				List: []*ast.Field{
					{
						Type: visitorType,
					},
				},
			},
		},
		Body: &ast.BlockStmt{
			List: []ast.Stmt{
				&ast.ReturnStmt{
					Results: []ast.Expr{
						&ast.UnaryExpr{
							Op: token.AND,
							X: &ast.CompositeLit{
								Type: visitorImplType,
								Elts: compositeKv,
							},
						},
					},
				},
			},
		},
	}
}

func visitorImpl(registry *namingRegistry, enumIdent string, member *ast.Ident, visitorReturnIdent ast.Expr) *ast.FuncDecl {
	// func (v __ExampleVisitor) VisitA(e A) error {
	// 	return v.__VisitA(e)
	// }

	var (
		stmts   []ast.Stmt
		results *ast.FieldList
		enumVal = ast.NewIdent("e")
		visitor = ast.NewIdent("v")
	)

	if visitorReturnIdent != nil {
		// func (v __ExampleVisitor) VisitA(e A) visitorReturnIdent {
		// 	return v.__VisitA(e)
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
							Sel: ast.NewIdent(fmt.Sprintf("__%s", registry.visitMethodName(enumIdent, member.String()))),
						},
						Args: []ast.Expr{
							enumVal,
						},
					},
				},
			},
		}
	} else {
		// func (v __ExampleVisitor) VisitA(e A) {
		// 	v.__VisitA(e)
		// }
		stmts = []ast.Stmt{
			&ast.ExprStmt{
				X: &ast.CallExpr{
					Fun: &ast.SelectorExpr{
						X:   visitor,
						Sel: ast.NewIdent(registry.visitMethodName(enumIdent, member.String())),
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
						visitor,
					},
					Type: ast.NewIdent(fmt.Sprintf("__%s", registry.visitorTypeName(enumIdent))),
				},
			},
		},
		Name: ast.NewIdent(registry.visitMethodName(enumIdent, member.String())),
		Type: &ast.FuncType{
			Params: &ast.FieldList{
				List: []*ast.Field{
					{
						Names: []*ast.Ident{
							enumVal,
						},
						Type: member,
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
