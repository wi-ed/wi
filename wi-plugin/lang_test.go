// Copyright 2014 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package wi

import (
	"testing"
)

func TestGetStr(t *testing.T) {
	a := LangMap{
		LangEn: "Foo",
		LangFr: "Bar",
	}
	if GetStr(LangFr, a) != "Bar" {
		t.Fail()
	}
	if GetStr(LangEs, a) != "Foo" {
		t.Fail()
	}
	if GetStr(LangEn, a) != "Foo" {
		t.Fail()
	}
}
