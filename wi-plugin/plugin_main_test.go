// Copyright 2014 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package wi

import (
	"testing"
)

func TestCalculateVersion(t *testing.T) {
	v := CalculateVersion()
	if v != "d997c07769b9114084a9764d0c236600ec78c979" {
		t.Fatal(v)
	}
}
