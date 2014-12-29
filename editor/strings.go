// Copyright 2014 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package editor

import (
	"github.com/maruel/wi/wicore"
)

var activateDisabled = wicore.LangMap{
	wicore.LangEn: "Can't activate a disabled view.",
}

var cantAddTwoWindowWithSameDocking = wicore.LangMap{
	wicore.LangEn: "Can't create two windows with the same docking \"%s\".",
}

var invalidDocking = wicore.LangMap{
	wicore.LangEn: "String \"%s\" does not refer to a valid Docking type.",
}

var invalidRect = wicore.LangMap{
	wicore.LangEn: "\"%s, %s, %s, %s\" does not refer to a valid Rect.",
}

var invalidViewFactory = wicore.LangMap{
	wicore.LangEn: "\"%s\" does not refer to a valid ViewFactory. Make sure the view factory was properly registered.",
}

var isNotValidWindow = wicore.LangMap{
	wicore.LangEn: "ID \"%s\" does not refer to a valid window ID.",
}

var notFound = wicore.LangMap{
	wicore.LangEn: "Command \"%s\" is not registered.",
}

// notMapped describes that a key is not mapped to any command.
var notMapped = wicore.LangMap{
	wicore.LangEn: "\"%s\" is not mapped to any command.",
}

var viewDirty = wicore.LangMap{
	wicore.LangEn: "View \"%s\" is not saved, aborting quit.",
}
