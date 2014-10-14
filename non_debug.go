// Copyright 2014 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// +build !debug

package main

import (
	"io"
	"log"
)

type nullWriter struct{}

func (nullWriter) Write(b []byte) (int, error) {
	return len(b), nil
}

func DebugHook() io.Closer {
	// It is important to get rid of log output on stderr as it would conflict
	// with the editor's use of the terminal. Sadly the strings are still
	// rasterized, I don't know of a way to get rid of this.
	log.SetFlags(0)
	log.SetOutput(nullWriter{})
	return nil
}
