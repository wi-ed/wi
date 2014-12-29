// Copyright 2014 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package wiCore

import (
	"testing"

	"github.com/maruel/ut"
)

func TestKey(t *testing.T) {
	for i := KeyNone; i < keyLast; i++ {
		ut.AssertEqual(t, i, KeyFromString(i.String()))
	}
}

func TestKeyPress(t *testing.T) {
	data := []KeyPress{
		{false, false, KeyNone, ' '},
		{false, false, KeyNone, 'a'},
		{false, false, KeyNone, 'A'},
		{false, false, KeyNone, 'é'},
		{false, false, KeyF1, '\000'},
	}
	for i, v := range data {
		ut.AssertEqualIndex(t, i, true, v.IsValid())
		v.Ctrl = true
		ut.AssertEqualIndex(t, i, v, KeyPressFromString(v.String()))
		v.Alt = true
		ut.AssertEqualIndex(t, i, v, KeyPressFromString(v.String()))
		v.Ctrl = false
		ut.AssertEqualIndex(t, i, v, KeyPressFromString(v.String()))
	}
}
