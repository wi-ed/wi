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
	mappings map[string]string
}

func (k *keyBindings) Register(keyName string, cmdName string) bool {
	_, ok := k.mappings[keyName]
	k.mappings[keyName] = cmdName
	return !ok
}

func (k *keyBindings) Get(keyName string) string {
	return k.mappings[keyName]
}

func makeKeyBindings() wi.KeyBindings {
	return &keyBindings{make(map[string]string)}
}

// keyEventToName returns the user printable key name like 'a', Ctrl-Alt-<f1>,
// <delete>, etc.
func keyEventToName(event termbox.Event) string {
	return tulib.KeyToString(event.Key, event.Ch, event.Mod)
}

// Registers the default keyboard mapping. Keyboard mapping simply execute the
// corresponding command. So to add a keyboard map, the corresponding command
// needs to be added first.
// TODO(maruel): This should be remappable via a configuration flag, for
// example vim flavor vs emacs flavor.
func RegisterDefaultKeyBindings(keyBindings wi.KeyBindings) {
	keyBindings.Register("F1", "help")
	keyBindings.Register("Ctrl-C", "quit")
}
