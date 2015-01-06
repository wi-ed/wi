// Copyright 2014 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package wicore

import (
	"github.com/maruel/wi/wicore/lang"
)

// AliasFor describes that a command is an alias.
var AliasFor = lang.Map{
	lang.En: "Alias for \"%s\".",
}

// AliasNotFound describes an alias to another command did not resolve.
var AliasNotFound = lang.Map{
	lang.En: "\"%s\" is an alias to command \"%s\" but this command is not registered.",
}
