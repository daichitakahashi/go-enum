package gen

import (
	"bytes"
	"errors"
	"fmt"
	"go/ast"
	"go/format"
	"go/token"
	"log"
	"os"
	"strconv"

	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/imports"
)

const (
	packagePath          = "github.com/daichitakahashi/go-enum"
	enumSymbol           = "MemberOf"
	visitorReturnsSymbol = "VisitorReturns"
)

func loadPackage() (*packages.Package, error) {
	pkgs, err := packages.Load(&packages.Config{
		Mode:  packages.NeedSyntax | packages.NeedFiles | packages.NeedCompiledGoFiles | packages.NeedName,
		Dir:   ".",
		Tests: false,
	}, "")
	if err != nil {
		return nil, err
	}

	for _, pkg := range pkgs {
		// return first package
		return pkg, nil
	}

	return nil, errors.New("no package loaded")
}

func Run(wd, filename string, namingVisitor []NamingVisitorParams, namingAccept []NamingAcceptParams, namingVisitorFactory []NamingVisitorImplParams) {
	err := os.Chdir(wd)
	if err != nil {
		log.Fatal(err)
	}

	pkg, err := loadPackage()
	if err != nil {
		log.Fatal(err)
	}

	f := &ast.File{
		Name: ast.NewIdent(pkg.Name),
	}
	registry := newNamingRegistry(namingVisitor, namingAccept, namingVisitorFactory)

	type targetFile struct {
		file        *ast.File
		enumPackage string
	}
	collectTargetPackages := pipelineStage(func(in *ast.File, out chan targetFile) {
		var collected bool
		for _, spec := range in.Imports {
			importPath, err := strconv.Unquote(spec.Path.Value)
			if err != nil {
				log.Fatal(err)
			}
			if importPath == packagePath {
				name := "enum"
				if spec.Name != nil {
					name = spec.Name.String()
				}
				// TODO: consider dot import
				out <- targetFile{
					file:        in,
					enumPackage: name,
				}
				collected = true
			}
		}
		if collected {
			// collect import specs
			f.Decls = append(
				f.Decls,
				importSpecs(in.Imports),
			)
		}
	})

	collectEnumDefinitions := pipelineStage(func(in targetFile, out chan enumMemberDefinition) {
		for _, decl := range in.file.Decls {
			if typeDecl, ok := decl.(*ast.GenDecl); ok {
				for _, spec := range typeDecl.Specs {
					if typeSpec, ok := spec.(*ast.TypeSpec); ok {
						if def, ok := extractEnumMemberDefinition(in.enumPackage, typeSpec); ok {
							out <- *def
						}
					}
				}
			}
		}
	})
	collectEnumIdentDefinitions := pipelineStage(func(in targetFile, out chan enumIdentDefinition) {
		for _, decl := range in.file.Decls {
			if typeDecl, ok := decl.(*ast.GenDecl); ok {
				for _, spec := range typeDecl.Specs {
					if typeSpec, ok := spec.(*ast.TypeSpec); ok {
						if def, ok := extractEnumIdentDefinition(in.enumPackage, typeSpec); ok {
							out <- *def
						}
					}
				}
			}
		}
	})
	collectDefinitions := merge(collectEnumDefinitions, collectEnumIdentDefinitions, func(o1 []enumMemberDefinition, o2 []enumIdentDefinition) pair[[]enumMemberDefinition, map[string]enumIdentDefinition] {
		m := make(map[string]enumIdentDefinition)
		for _, d := range o2 {
			m[fmt.Sprint(d.ident)] = d
		}
		return pair[[]enumMemberDefinition, map[string]enumIdentDefinition]{
			left:  o1,
			right: m,
		}
	})

	// assemble enumInfo per enum type
	type enumInfo struct {
		ident              ast.Expr
		members            []*ast.Ident
		visitorReturnIdent ast.Expr
	}
	assembleEnumInfo := func(in pair[[]enumMemberDefinition, map[string]enumIdentDefinition]) <-chan enumInfo {
		var (
			dict = map[string]*enumInfo{}
			list []*enumInfo
		)

		for _, def := range in.left {
			enumIdent := fmt.Sprint(def.enumIdent)

			var visitorReturnIdent ast.Expr
			if e, ok := in.right[enumIdent]; ok {
				visitorReturnIdent = e.visitorReturnIdent
			}

			if info, ok := dict[enumIdent]; ok {
				info.members = append(info.members, def.ident)
			} else {
				info := &enumInfo{
					ident:              def.enumIdent,
					members:            []*ast.Ident{def.ident},
					visitorReturnIdent: visitorReturnIdent,
				}
				dict[enumIdent] = info
				list = append(list, info)
			}
		}

		out := make(chan enumInfo)
		go func() {
			defer close(out)
			for _, info := range list {
				out <- *info
			}
		}()
		return out
	}

	generateDecl := pipelineStage(func(in enumInfo, out chan ast.Decl) {
		enumIdent := fmt.Sprint(in.ident)
		out <- &ast.GenDecl{
			Tok: token.TYPE,
			Specs: []ast.Spec{
				visitorSpec(registry, enumIdent, in.members, in.visitorReturnIdent),
				enumSpec(registry, enumIdent, in.visitorReturnIdent),
			},
		}

		// implementations of Accept
		for _, m := range in.members {
			out <- acceptImpl(registry, enumIdent, m, in.visitorReturnIdent)
		}

		// type checks
		out <- typeCheckDecl(enumIdent, in.members)

		// visitor impl factory
		visitorFactory, found := registry.visitorImplFactoryName(enumIdent)
		if found {
			out <- visitorImplSpec(registry, enumIdent, in.members, in.visitorReturnIdent)
			out <- visitorFactoryImpl(registry, enumIdent, visitorFactory, in.members, in.visitorReturnIdent)
			for _, m := range in.members {
				out <- visitorImpl(registry, enumIdent, m, in.visitorReturnIdent)
			}
		}
	})

	decls :=
		pipe(collectTargetPackages,
			pipe(collectDefinitions,
				pipe(assembleEnumInfo,
					generateDecl,
				),
			),
		)(iterate(pkg.Syntax))

	for decl := range decls {
		f.Decls = append(f.Decls, decl)
	}
	if len(f.Decls) == 0 {
		log.Fatal("target type not found")
	}

	code, err := generateCode(f)
	if err != nil {
		log.Fatal(err)
	}
	err = os.WriteFile(filename, code, 0644)
	if err != nil {
		log.Fatal(err)
	}
}

const codeGeneratedMark = `// Code generated by enumgen. DO NOT EDIT.`

func generateCode(f *ast.File) ([]byte, error) {
	buf := bytes.NewBuffer(nil)

	err := format.Node(buf, token.NewFileSet(), f)
	if err != nil {
		return nil, err
	}

	source, err := imports.Process("", buf.Bytes(), &imports.Options{
		Fragment:   false,
		AllErrors:  true,
		Comments:   true,
		TabIndent:  true,
		TabWidth:   8,
		FormatOnly: false,
	})
	if err != nil {
		return nil, err
	}

	buf.Reset()
	buf.WriteString(fmt.Sprintf("%s\n\n", codeGeneratedMark))
	buf.Write(source)
	return buf.Bytes(), nil
}
