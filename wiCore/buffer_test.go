// Copyright 2014 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package wiCore

import (
	"testing"

	"github.com/maruel/ut"
)

func TestFormatText(t *testing.T) {
	data := [][]string{
		{"hello", "hello"},
		{"\000hello", "NULhello"},
		{"\001", "^A"},
		{"	a", "	a"},
	}
	for i, v := range data {
		ut.AssertEqualIndex(t, i, v[1], FormatText(v[0]))
	}
}
