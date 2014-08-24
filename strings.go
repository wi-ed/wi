// Copyright 2014 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package main

import (
	"github.com/maruel/wi/wi-plugin"
)

var activateDisabled = wi.LangMap{
	wi.LangEn: "Can't activate a disabled view.",
}

var aliasFor = wi.LangMap{
	wi.LangEn: "Alias for \"%s\".",
}

var aliasNotFound = wi.LangMap{
	wi.LangEn: "\"%s\" is an alias to command \"%s\" but this command is not registered.",
}

var notFound = wi.LangMap{
	wi.LangEn: "Command \"%s\" is not registered.",
}

var viewDirty = wi.LangMap{
	wi.LangEn: "View \"%s\" is not saved, aborting quit.",
}
