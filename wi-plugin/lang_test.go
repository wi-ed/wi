// Copyright 2014 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package wi

import (
	"github.com/maruel/ut"
	"testing"
)

func TestGetStr(t *testing.T) {
	a := LangMap{
		LangEn: "Foo",
		LangFr: "Bar",
	}
	ut.AssertEqual(t, "Bar", GetStr(LangFr, a))
	ut.AssertEqual(t, "Foo", GetStr(LangEs, a))
	ut.AssertEqual(t, "Foo", GetStr(LangEn, a))
}
