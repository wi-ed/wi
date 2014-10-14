// Copyright 2014 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package wiCore

import (
	"fmt"
	"io"
	"net/rpc"
	"os"
	"reflect"

	"github.com/maruel/interface_guid"
)

// CalculateVersion returns the hex string of the hash of the primary
// interfaces for this package.
//
// It traverses the Editor type recursively, expanding all types referenced
// recursively. This data is used to generate an hash, that will represent the
// version of this interface.
func CalculateVersion() string {
	return interface_guid.CalculateGUID(reflect.TypeOf((*Editor)(nil)).Elem())
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

// MakeReadWriteCloser creates a io.ReadWriteCloser out of one io.ReadCloser
// and one io.WriteCloser.
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
