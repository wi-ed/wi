// Copyright 2014 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package wi

import (
	"testing"
)

func TestCalculateVersion(t *testing.T) {
	v := CalculateVersion()
	// TODO(maruel): We don't care about the actual version. Just test the
	// underlying code to calculate a version is working.
	if v != "a22cdfe7d99e93fb642804b786521611e8e4c7c3" {
		t.Fatal(v)
	}
}
