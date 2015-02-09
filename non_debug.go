// Copyright 2014 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// +build !debug

package main

import (
	"io"
	"io/ioutil"
	"log"

	_ "github.com/maruel/circular"
	"github.com/maruel/wi/editor"
)

func debugHook() io.Closer {
	// It is important to get rid of log output on stderr as it would conflict
	// with the editor's use of the terminal. Sadly the strings are still
	// rasterized, I don't know of a way to get rid of this.
	log.SetFlags(0)
	log.SetOutput(ioutil.Discard)
	return nil
}

func debugHookEditor(e editor.Editor) {
}
