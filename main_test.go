// Copyright 2014 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package main

import (
	"log"
	"testing"
)

func TestInnerMain(t *testing.T) {
	// TODO(maruel): This has persistent side-effect. Figure out how to handle
	// "log" properly. Likely by using the same mechanism as used in package
	// "subcommands".
	log.SetOutput(new(nullWriter))

	result := innerMain(true, []string{"quit"})
	if result != 0 {
		t.Fatalf("Exit code: %v", result)
	}
}
