// Copyright 2014 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package editor

import (
	"github.com/wi-ed/wi/wicore/lang"
)

var activateDisabled = lang.Map{
	lang.En: "Can't activate a disabled view.",
}

var cantAddTwoWindowWithSameDocking = lang.Map{
	lang.En: "Can't create two windows with the same docking \"%s\".",
}

var invalidDocking = lang.Map{
	lang.En: "String \"%s\" does not refer to a valid Docking type.",
}

var invalidRect = lang.Map{
	lang.En: "\"%s, %s, %s, %s\" does not refer to a valid Rect.",
}

var invalidViewFactory = lang.Map{
	lang.En: "\"%s\" does not refer to a valid ViewFactory. Make sure the view factory was properly registered.",
}

var isNotValidWindow = lang.Map{
	lang.En: "ID \"%s\" does not refer to a valid window ID.",
}

var notFound = lang.Map{
	lang.En: "Command \"%s\" is not registered.",
}

// notMapped describes that a key is not mapped to any command.
var notMapped = lang.Map{
	lang.En: "\"%s\" is not mapped to any command.",
}

var viewDirty = lang.Map{
	lang.En: "View \"%s\" is not saved, aborting quit.",
}
