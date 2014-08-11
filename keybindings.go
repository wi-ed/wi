// Copyright 2014 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package main

import (
	"github.com/maruel/wi/wi-plugin"
	"github.com/nsf/termbox-go"
	"github.com/nsf/tulib"
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

// keyEventToName returns the user printable key name like 'a', Ctrl-Alt-<f1>,
// <delete>, etc.
// TODO(maruel): I dislike the format of tulib, redo.
func keyEventToName(event termbox.Event) string {
	return tulib.KeyToString(event.Key, event.Ch, event.Mod)
}

// Registers the default keyboard mapping. Keyboard mapping simply execute the
// corresponding command. So to add a keyboard map, the corresponding command
// needs to be added first.
// TODO(maruel): This should be remappable via a configuration flag, for
// example vim flavor vs emacs flavor. I'm not sure it's worth supporting this
// without a restart.
func RegisterDefaultKeyBindings(cd wi.CommandDispatcher) {
	cd.PostCommand("keybind", "global", "all", "F1", "help")
	cd.PostCommand("keybind", "global", "all", "Ctrl-C", "quit")
}
