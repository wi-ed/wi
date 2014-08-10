// Copyright 2014 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package wi

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"net/rpc"
	"os"
	"reflect"
)

type set map[string]bool

func recurseType(h io.Writer, t reflect.Type, seen set) {
	kind := t.Kind()
	h.Write([]byte(kind.String()))
	if kind == reflect.Interface {
		name := t.Name()
		h.Write([]byte(name))
		if seen[name] {
			return
		}
		seen[name] = true
		for i := 0; i < t.NumMethod(); i++ {
			recurseMethod(h, t.Method(i), seen)
		}
	} else if kind == reflect.Struct {
		name := t.Name()
		h.Write([]byte(name))
		if seen[name] {
			return
		}
		seen[name] = true
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			h.Write([]byte(f.Name))
			recurseType(h, f.Type, seen)
		}
		for i := 0; i < t.NumMethod(); i++ {
			recurseMethod(h, t.Method(i), seen)
		}
	} else if kind == reflect.Array || kind == reflect.Chan || kind == reflect.Ptr || kind == reflect.Slice {
		recurseType(h, t.Elem(), seen)
	} else if kind == reflect.Map {
		recurseType(h, t.Key(), seen)
		recurseType(h, t.Elem(), seen)
	} else if kind >= reflect.Bool && kind <= reflect.Complex128 || kind == reflect.String {
		// Base types.
	} else {
		panic(kind.String())
	}
}

func recurseMethod(h io.Writer, m reflect.Method, seen set) {
	h.Write([]byte(m.Name))
	for i := 0; i < m.Type.NumIn(); i++ {
		recurseType(h, m.Type.In(i), seen)
	}
	for i := 0; i < m.Type.NumOut(); i++ {
		recurseType(h, m.Type.Out(i), seen)
	}
}

// CalculateVersion returns the hex string of the hash of the primary
// interfaces for this package.
//
// It traverses the Editor type recursively, expanding all types referenced
// recursively. This data is used to generate an hash, that will represent the
// version of this interface.
func CalculateVersion() string {
	h := sha1.New()
	recurseType(h, reflect.TypeOf((*Editor)(nil)).Elem(), make(set))
	return hex.EncodeToString(h.Sum(nil))
}

type multiCloser []io.Closer

func (m multiCloser) Close() (err error) {
	for _, i := range m {
		err1 := i.Close()
		if err1 != nil {
			err = err1
		}
	}
	return
}

func MakeReadWriteCloser(reader io.ReadCloser, writer io.WriteCloser) io.ReadWriteCloser {
	return struct {
		io.Reader
		io.Writer
		io.Closer
	}{reader, writer, multiCloser{reader, writer}}
}

// Main is the function to call from your plugin to initiate the communication
// channel between wi and your plugin. It takes care of the versioning.
func Main() {
	if os.ExpandEnv("${WI}") != "plugin" {
		fmt.Print("This is a wi plugin. This program is only meant to be run through wi itself.")
		os.Exit(1)
	}
	// TODO(maruel): Take garbage from os.Stdin, put garbage in os.Stdout.
	fmt.Print(CalculateVersion())

	client := rpc.NewClient(MakeReadWriteCloser(os.Stdin, os.Stdout))
	// Do something with client.
	err := client.Call("Editor", "Height", nil)
	if err != nil {
		panic(err)
	}
	os.Exit(0)
}
