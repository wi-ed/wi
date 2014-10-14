// Copyright 2014 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package editor

import (
	"github.com/maruel/wi/wiCore"
)

var activateDisabled = wiCore.LangMap{
	wiCore.LangEn: "Can't activate a disabled view.",
}

var cantAddTwoWindowWithSameDocking = wiCore.LangMap{
	wiCore.LangEn: "Can't create two windows with the same docking \"%s\".",
}

var invalidDocking = wiCore.LangMap{
	wiCore.LangEn: "String \"%s\" does not refer to a valid Docking type.",
}

var invalidRect = wiCore.LangMap{
	wiCore.LangEn: "\"%s, %s, %s, %s\" does not refer to a valid Rect.",
}

var invalidViewFactory = wiCore.LangMap{
	wiCore.LangEn: "\"%s\" does not refer to a valid ViewFactory. Make sure the view factory was properly registered.",
}

var isNotValidWindow = wiCore.LangMap{
	wiCore.LangEn: "ID \"%s\" does not refer to a valid window ID.",
}

var notFound = wiCore.LangMap{
	wiCore.LangEn: "Command \"%s\" is not registered.",
}

// notMapped describes that a key is not mapped to any command.
var notMapped = wiCore.LangMap{
	wiCore.LangEn: "\"%s\" is not mapped to any command.",
}

var viewDirty = wiCore.LangMap{
	wiCore.LangEn: "View \"%s\" is not saved, aborting quit.",
}
