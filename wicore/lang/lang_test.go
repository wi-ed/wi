// Copyright 2014 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package lang

import (
	"testing"

	"github.com/maruel/ut"
)

func TestGetDefaultEn(t *testing.T) {
	a := Map{
		En:   "Foo",
		FrCa: "Bar",
	}
	ut.AssertEqual(t, "Foo", a.Get(Es))
}

func TestGetDefaultNonCountry(t *testing.T) {
	a := Map{
		Fr: "Bar",
	}
	ut.AssertEqual(t, "Bar", a.Get(FrCa))
}

func TestGetDefaultMissing(t *testing.T) {
	a := Map{
		FrCa: "Bar",
	}
	ut.AssertEqual(t, "", a.Get(Es))
}

func TestString(t *testing.T) {
	a := Map{
		FrCa: "Bar",
	}
	Set(FrCa)
	ut.AssertEqual(t, "Bar", a.String())
}

func TestFormatf(t *testing.T) {
	a := Map{
		FrCa: "Bar %d",
	}
	Set(FrCa)
	ut.AssertEqual(t, "Bar 2", a.Formatf(2))
}
