// Copyright 2014 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package editor

import (
	"github.com/maruel/wi/wi-plugin"
)

var activateDisabled = wi.LangMap{
	wi.LangEn: "Can't activate a disabled view.",
}

var cantAddTwoWindowWithSameDocking = wi.LangMap{
	wi.LangEn: "Can't create two windows with the same docking \"%s\".",
}

var invalidDocking = wi.LangMap{
	wi.LangEn: "String \"%s\" does not refer to a valid Docking type.",
}

var invalidRect = wi.LangMap{
	wi.LangEn: "\"%s, %s, %s, %s\" does not refer to a valid Rect.",
}

var invalidViewFactory = wi.LangMap{
	wi.LangEn: "\"%s\" does not refer to a valid ViewFactory. Make sure the view factory was properly registered.",
}

var isNotValidWindow = wi.LangMap{
	wi.LangEn: "ID \"%s\" does not refer to a valid window ID.",
}

var notFound = wi.LangMap{
	wi.LangEn: "Command \"%s\" is not registered.",
}

// notMapped describes that a key is not mapped to any command.
var notMapped = wi.LangMap{
	wi.LangEn: "\"%s\" is not mapped to any command.",
}

var viewDirty = wi.LangMap{
	wi.LangEn: "View \"%s\" is not saved, aborting quit.",
}
