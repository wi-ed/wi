// Copyright 2015 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package parse contains simple go parser code to be used to generate code
// from interfaces.
package parse

import (
	"bytes"
	"errors"
	"fmt"
	"go/ast"
	"go/format"
)

// Arg is one parameter or return value.
type Arg struct {
	Name string
	Type string
}

// Method is a simplification of ast.FuncType using only strings.
type Method struct {
	Name    string
	Params  []Arg
	Results []Arg
}

// FindType finds a file level type declaration and returns it if found.
func FindType(f *ast.File, inputType string) *ast.TypeSpec {
	// Look at all file level declarations.
	for _, decl := range f.Decls {
		y, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}
		// Search for a type specification.
		for _, s := range y.Specs {
			t, ok := s.(*ast.TypeSpec)
			if !ok {
				continue
			}
			if inputType != t.Name.Name {
				continue
			}
			return t
		}
	}
	return nil
}

// processFieldList processes params or results of a method.
func processFieldList(list *ast.FieldList) ([]Arg, error) {
	if list == nil || len(list.List) == 0 {
		return []Arg{}, nil
	}
	out := make([]Arg, 0, len(list.List))
	for _, param := range list.List {
		typeName := ""
		selector, ok := param.Type.(*ast.SelectorExpr)
		if ok {
			ident, ok := selector.X.(*ast.Ident)
			if !ok {
				return out, errors.New("failed to process field")
			}
			typeName = ident.Name + "." + selector.Sel.Name
		} else {
			ident, ok := param.Type.(*ast.Ident)
			if !ok {
				return out, errors.New("failed to process field")
			}
			typeName = ident.Name
		}
		for _, name := range param.Names {
			arg := Arg{
				Name: name.Name,
				Type: typeName,
			}
			out = append(out, arg)
		}
	}
	return out, nil
}

// EnumInterface enumerates all the methods of an interface.
//
// Useful to generate code from an interface.
func EnumInterface(t *ast.TypeSpec) ([]Method, error) {
	typeName := t.Name.Name
	i, ok := t.Type.(*ast.InterfaceType)
	if !ok {
		return nil, fmt.Errorf("expected %s to be an interface", typeName)
	}
	out := make([]Method, 0, len(i.Methods.List))
	for _, m := range i.Methods.List {
		methodName := m.Names[0].Name
		methodFunc, ok := m.Type.(*ast.FuncType)
		if !ok {
			return out, fmt.Errorf("expected %s.%s to be a method", typeName, methodName)
		}
		params, err := processFieldList(methodFunc.Params)
		if err != nil {
			return out, fmt.Errorf("%s.%s: params %s", i, typeName, methodName, err)
		}
		results, err := processFieldList(methodFunc.Results)
		if err != nil {
			return out, fmt.Errorf("%s.%s: results %s", i, typeName, methodName, err)
		}
		method := Method{
			Name:    methodName,
			Params:  params,
			Results: results,
		}
		out = append(out, method)
	}
	return out, nil
}

// FormatSource runs the equivalent of gofmt on the buffer and returns the
// formatted version.
//
// If the code is not valid go syntax, prefix the buffer with a comment about
// this issue and include the buffer unformatted.
func FormatSource(buf []byte) ([]byte, error) {
	src, err := format.Source(buf)
	if err != nil {
		b := bytes.Buffer{}
		fmt.Printf("// ERROR: internal error: invalid Go generated: %s\n", err)
		fmt.Printf("// Compile the package to analyze the error.\n\n")
		_, _ = b.Write(buf)
		src = b.Bytes()
	}
	return src, err
}
