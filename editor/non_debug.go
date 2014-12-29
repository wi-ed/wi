// Copyright 2014 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// +build !debug

package editor

import (
	"github.com/maruel/wi/wicore"
)

// RegisterDebugCommands registers nothing in Release build.
func RegisterDebugCommands(dispatcher wicore.Commands) {
}
