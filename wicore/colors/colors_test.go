// Copyright 2014 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package colors

import (
	"testing"

	"github.com/maruel/ut"
)

func TestNearestEGA(t *testing.T) {
	ut.AssertEqual(t, Black, NearestEGA(RGB{1, 1, 1}))
	ut.AssertEqual(t, White, NearestEGA(RGB{253, 253, 253}))
}
