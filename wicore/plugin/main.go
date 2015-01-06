// Copyright 2014 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package plugin implements the common code to implement a wi plugin.
package plugin

import (
	"fmt"
	"net/rpc"
	"os"

	"github.com/maruel/wi/wicore"
)

// Main is the function to call from your plugin to initiate the communication
// channel between wi and your plugin.
func Main() {
	if os.ExpandEnv("${WI}") != "plugin" {
		fmt.Fprint(os.Stderr, "This is a wi plugin. This program is only meant to be run through wi itself.\n")
		os.Exit(1)
	}
	// TODO(maruel): Take garbage from os.Stdin, put garbage in os.Stdout.
	fmt.Print(wicore.CalculateVersion())

	client := rpc.NewClient(wicore.MakeReadWriteCloser(os.Stdin, os.Stdout))
	// Do something with client.
	err := client.Call("Editor", "Height", nil)
	if err != nil {
		panic(err)
	}
	os.Exit(0)
}
