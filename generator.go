package main

import (
	"bytes"
	"fmt"
	"github.com/samber/lo"
	"go/ast"
	exact "go/constant"
	"go/token"
	"go/types"
	"golang.org/x/tools/go/packages"
	"log"
	"strings"
	"text/template"
)

type Generator struct {
	buf bytes.Buffer
	pkg *Package
}

// Package holds information about a Go package
type Package struct {
	dir      string
	name     string
	defs     map[*ast.Ident]types.Object
	files    []*File
	typesPkg *types.Package
}

// File holds a single parsed file and associated data.
type File struct {
	pkg  *Package  // Package to which this file belongs.
	file *ast.File // Parsed AST.
	// These fields are reset for each type being generated.
	typeName    string  // Name of the constant type.
	values      []Value // Accumulator for constant values of that type.
	trimPrefix  string
	lineComment bool
}

// Value represents a declared constant.
type Value struct {
	originalName string // The name of the constant before transformation
	name         string // The name of the constant after transformation (i.e. camel case => snake case)
	// The value is stored as a bit pattern alone. The boolean tells us
	// whether to interpret it as an int64 or a uint64; the only place
	// this matters is when sorting.
	// Much of the time the str field is all we need; it is printed
	// by Value.String.
	value  uint64 // Will be converted to int64 when needed.
	signed bool   // Whether the constant is a signed type.
	str    string // The string representation given by the "go/exact" package.
}

// parsePackage analyzes the single package constructed from the patterns and tags.
// parsePackage exits if there is an error.
func (g *Generator) parsePackage(patterns []string, tags []string) {
	cfg := &packages.Config{
		Mode:  packages.LoadSyntax,
		Tests: false,
	}
	pkgs, err := packages.Load(cfg, patterns...)
	if err != nil {
		log.Fatal(err)
	}
	if len(pkgs) != 1 {
		log.Fatalf("error: %d packages found", len(pkgs))
	}
	g.addPackage(pkgs[0])
}

// addPackage adds a type checked Package and its syntax files to the generator.
func (g *Generator) addPackage(pkg *packages.Package) {
	g.pkg = &Package{
		name:  pkg.Name,
		defs:  pkg.TypesInfo.Defs,
		files: make([]*File, len(pkg.Syntax)),
	}

	for i, file := range pkg.Syntax {
		g.pkg.files[i] = &File{
			file: file,
			pkg:  g.pkg,
		}
	}
}

func (g *Generator) Printf(format string, args ...interface{}) {
	_, _ = fmt.Fprintf(&g.buf, format, args...)
}

func (f *File) genDecl(node ast.Node) bool {
	decl, ok := node.(*ast.GenDecl)
	if !ok || decl.Tok != token.CONST {
		return true
	}

	var typ = ""

	for _, spec := range decl.Specs {
		vspec := spec.(*ast.ValueSpec)
		//fmt.Printf("%+v\n", vspec)
		if vspec.Type == nil && len(vspec.Values) > 0 {
			typ = ""
			continue
		}
		if vspec.Type != nil {
			ident, ok := vspec.Type.(*ast.Ident)
			if !ok {
				continue
			}
			typ = ident.Name
		}
		if typ != f.typeName {
			continue
		}

		for _, n := range vspec.Names {
			if n.Name == "_" {
				continue
			}

			obj, ok := f.pkg.defs[n]
			if !ok {
				log.Fatalf("no value for constant %s", n)
			}
			info := obj.Type().Underlying().(*types.Basic).Info()
			if info&types.IsInteger == 0 {
				log.Fatalf("can't handle non-integer constant type %s", typ)
			}
			value := obj.(*types.Const).Val() // Guaranteed to succeed as this is CONST.
			if value.Kind() != exact.Int {
				log.Fatalf("can't happen: constant is not an integer %s", n)
			}
			i64, isInt := exact.Int64Val(value)
			u64, isUint := exact.Uint64Val(value)
			if !isInt && !isUint {
				log.Fatalf("internal error: value of %s is not an integer: %s", n, value.String())
			}
			if !isInt {
				u64 = uint64(i64)
			}
			v := Value{
				originalName: n.Name,
				name:         n.Name,
				value:        u64,
				signed:       info&types.IsUnsigned == 0,
				str:          value.String(),
			}
			if c := vspec.Comment; f.lineComment && c != nil && len(c.List) == 1 {
				v.name = strings.TrimSpace(c.Text())
			}

			f.values = append(f.values, v)
		}
	}

	return false
}

func (g *Generator) generate(
	typeName, trimPrefix, addPrefix, transform string,
	templateFile string, lineComment bool,
) error {

	values := make([]Value, 0, 100)
	for _, file := range g.pkg.files {
		file.typeName = typeName
		file.lineComment = lineComment
		ast.Inspect(file.file, file.genDecl)
		values = append(values, file.values...)
	}

	if len(values) == 0 {
		return fmt.Errorf("no values defined for type %s", typeName)
	}

	values = trimValueNames(values, trimPrefix)

	values = transformValueNames(values, transform)

	values = prefixValueNames(values, addPrefix)

	// Open Template File
	tmpl, err := template.ParseFiles(templateFile)
	if err != nil {
		return err
	}

	// Execute Template
	err = tmpl.Execute(&g.buf, map[string]any{
		"Values": lo.Map(values, func(item Value, index int) map[string]any {
			return map[string]interface{}{
				"name":         item.name,
				"value":        item.value,
				"str":          item.str,
				"originalName": item.originalName,
				"signed":       item.signed,
			}
		}),
		"Name": typeName,
	})
	if err != nil {
		return err
	}

	return nil
}
