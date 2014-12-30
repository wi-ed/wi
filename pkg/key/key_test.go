// Copyright 2014 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package key

import (
	"testing"

	"github.com/maruel/ut"
)

func TestKey(t *testing.T) {
	for i := None; i < last; i++ {
		s := i.String()
		ut.AssertEqual(t, true, len(s) > 1)
		ut.AssertEqual(t, i, StringToKey(s))
	}
}

func TestPress(t *testing.T) {
	data := []Press{
		{false, false, None, 'a'},
		{false, false, None, 'A'},
		{false, false, None, 'é'},
		{false, false, Space, '\000'},
		{false, false, Tab, '\000'},
		{false, false, F1, '\000'},
	}
	for i, v := range data {
		ut.AssertEqualIndex(t, i, true, v.IsValid())
		v.Ctrl = true
		ut.AssertEqualIndex(t, i, v, StringToPress(v.String()))
		v.Alt = true
		ut.AssertEqualIndex(t, i, v, StringToPress(v.String()))
		v.Ctrl = false
		ut.AssertEqualIndex(t, i, v, StringToPress(v.String()))
	}
}
