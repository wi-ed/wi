// Copyright 2014 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package main

import (
	"github.com/maruel/wi/wi-plugin"
	"testing"
)

func TestGetStr(t *testing.T) {
	a := langMap{
		wi.LangEn: "Foo",
		wi.LangFr: "Bar",
	}
	if getStr(wi.LangFr, a) != "Bar" {
		t.Fail()
	}
	if getStr(wi.LangEs, a) != "Foo" {
		t.Fail()
	}
	if getStr(wi.LangEn, a) != "Foo" {
		t.Fail()
	}
}
