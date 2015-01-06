// Copyright 2014 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package wicore

import (
	"fmt"
	"runtime"
)

// GotPanic must be listened to in the main thread to know if a panic() call
// occured in any goroutine running under Go().
var GotPanic <-chan interface{}

var gotPanic chan<- interface{}

func init() {
	p := make(chan interface{})
	GotPanic = p
	gotPanic = p
}

// Go wraps a call to wrap any panic() call and pipe it to the main goroutine.
//
// This permits clean shutdown, like terminal cleanup or plugin closure, when
// the program crashes.
func Go(name string, f func()) {
	go func() {
		defer func() {
			if i := recover(); i != nil {
				// TODO(maruel): No allocation in recover handling.
				buf := make([]byte, 2048)
				n := runtime.Stack(buf, false)
				if n != 0 {
					buf = buf[:n-1]
				}
				gotPanic <- fmt.Errorf("%s panicked: %s\nStack: %s", name, i, buf)
			}
		}()
		f()
	}()
}
