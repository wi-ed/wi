// Copyright 2014 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package lang

import (
	"testing"

	"github.com/maruel/ut"
)

func TestGetStr(t *testing.T) {
	a := Map{
		En: "Foo",
		Fr: "Bar",
	}
	ut.AssertEqual(t, "Bar", a.Get(Fr))
	ut.AssertEqual(t, "Foo", a.Get(Es))
	ut.AssertEqual(t, "Foo", a.Get(En))
}
