// Copyright 2014 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package wicore

import (
	"github.com/maruel/wi/pkg/key"
)

// KeyboardMode defines the keyboard mapping (input mode) to use.
//
// Unlike vim, there's no Ex mode. It's unnecessary because the command window
// is a Window on its own, instead of a additional input mode on the current
// Window.
type KeyboardMode int

const (
	// CommandMode is the mode where typing letters results in commands, not
	// content editing. It's named Normal mode in vim.
	//
	// TODO(maruel): Rename for consistency?
	CommandMode KeyboardMode = iota + 1
	// EditMode is the mode where typing letters results in content, not commands.
	EditMode
	// AllMode is to bind keys independent of the current mode. It is useful for
	// function keys, Ctrl-<letter>, arrow keys, etc.
	AllMode
)

// KeyBindings stores the mapping between keyboard entry and commands. This
// includes what can be considered "macros" as much as casual things like arrow
// keys.
type KeyBindings interface {
	// Set registers a keyboard mapping. In practice keyboard mappings
	// should normally be registered on startup. Returns false if a key mapping
	// was already registered and was lost. Set cmdName to "" to remove a key
	// binding.
	Set(mode KeyboardMode, key key.Press, cmdName string) bool

	// Get returns a command if registered, nil otherwise.
	Get(mode KeyboardMode, key key.Press) string
}

// GetKeyBindingCommand traverses the Editor's Window tree to find a View that
// has the key binding in its Keyboard mapping.
func GetKeyBindingCommand(e Editor, mode KeyboardMode, key key.Press) string {
	active := e.ActiveWindow()
	for {
		cmdName := active.View().KeyBindings().Get(mode, key)
		if cmdName != "" {
			return cmdName
		}
		active = active.Parent()
		if active == nil {
			return ""
		}
	}
}
