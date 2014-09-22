// Copyright 2014 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package editor

import (
	"github.com/maruel/wi/wi-plugin"
)

type keyBindings struct {
	commandMappings map[string]string
	editMappings    map[string]string
}

func (k *keyBindings) Set(mode wi.KeyboardMode, keyName string, cmdName string) bool {
	var ok bool
	if mode == wi.AllMode || mode == wi.CommandMode {
		_, ok = k.commandMappings[keyName]
		k.commandMappings[keyName] = cmdName
	}
	if mode == wi.AllMode || mode == wi.EditMode {
		_, ok = k.editMappings[keyName]
		k.editMappings[keyName] = cmdName
	}
	return !ok
}

func (k *keyBindings) Get(mode wi.KeyboardMode, keyName string) string {
	if mode == wi.CommandMode {
		return k.commandMappings[keyName]
	}
	if mode == wi.EditMode {
		return k.editMappings[keyName]
	}
	v, ok := k.commandMappings[keyName]
	if !ok {
		return k.editMappings[keyName]
	}
	return v
}

func makeKeyBindings() wi.KeyBindings {
	return &keyBindings{make(map[string]string), make(map[string]string)}
}

// RegisterDefaultKeyBindings registers the default keyboard mapping. Keyboard
// mapping simply execute the corresponding command. So to add a keyboard map,
// the corresponding command needs to be added first.
//
// TODO(maruel): This should be remappable via a configuration flag, for
// example vim flavor vs emacs flavor. I'm not sure it's worth supporting this
// without a restart.
func RegisterDefaultKeyBindings(cd wi.CommandDispatcher) {
	wi.PostCommand(cd, "keybind", "global", "all", "F1", "help")
	wi.PostCommand(cd, "keybind", "global", "all", "Ctrl-C", "quit")
	wi.PostCommand(cd, "keybind", "global", "command", ":", "show_command_window")
}
