// Copyright 2014 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package editor

import (
	"github.com/maruel/wi/pkg/key"
	"github.com/maruel/wi/pkg/lang"
	"github.com/maruel/wi/wicore"
)

type keyBindings struct {
	normalMappings map[key.Press]string
	insertMappings map[key.Press]string
}

func (k *keyBindings) Set(mode wicore.KeyboardMode, key key.Press, cmdName string) bool {
	if !key.IsValid() {
		return false
	}
	var ok bool
	if mode == wicore.AllMode || mode == wicore.Normal {
		_, ok = k.normalMappings[key]
		k.normalMappings[key] = cmdName
	}
	if mode == wicore.AllMode || mode == wicore.Insert {
		_, ok = k.insertMappings[key]
		k.insertMappings[key] = cmdName
	}
	return !ok
}

func (k *keyBindings) Get(mode wicore.KeyboardMode, key key.Press) string {
	if !key.IsValid() {
		return ""
	}
	if mode == wicore.Normal || mode == wicore.AllMode {
		if v, ok := k.normalMappings[key]; ok {
			return v
		}
	}
	if mode == wicore.Insert || mode == wicore.AllMode {
		if v, ok := k.insertMappings[key]; ok {
			return v
		}
	}
	return ""
}

func (k *keyBindings) GetAssigned(mode wicore.KeyboardMode) []key.Press {
	out := []key.Press{}
	if mode == wicore.Normal || mode == wicore.AllMode {
		for k := range k.normalMappings {
			out = append(out, k)
		}
	}
	if mode == wicore.Insert || mode == wicore.AllMode {
		for k := range k.insertMappings {
			out = append(out, k)
		}
	}
	return out
}

func makeKeyBindings() wicore.KeyBindings {
	return &keyBindings{make(map[key.Press]string), make(map[key.Press]string)}
}

// Commands.

func cmdKeyBind(c *wicore.CommandImpl, e wicore.Editor, w wicore.Window, args ...string) {
	location := args[0]
	modeName := args[1]
	keyName := args[2]
	cmdName := args[3]

	if location == "global" {
		w = wicore.RootWindow(w)
	} else if location != "window" {
		cmd := wicore.GetCommand(e, w, "key_bind")
		e.ExecuteCommand(w, "alert", cmd.LongDesc())
		return
	}

	var mode wicore.KeyboardMode
	if modeName == "command" {
		mode = wicore.Normal
	} else if modeName == "edit" {
		mode = wicore.Normal
	} else if modeName == "all" {
		mode = wicore.AllMode
	} else {
		cmd := wicore.GetCommand(e, w, "key_bind")
		e.ExecuteCommand(w, "alert", cmd.LongDesc())
		return
	}
	// TODO(maruel): Refuse invalid keyName.
	k := key.StringToPress(keyName)
	w.View().KeyBindings().Set(mode, k, cmdName)
}

// RegisterKeyBindingCommands registers the keyboard mapping related commands.
func RegisterKeyBindingCommands(dispatcher wicore.Commands) {
	cmds := []wicore.Command{
		&wicore.CommandImpl{
			"key_bind",
			4,
			cmdKeyBind,
			wicore.CommandsCategory,
			lang.Map{
				lang.En: "Binds a keyboard mapping to a command",
			},
			lang.Map{
				lang.En: "Usage: key_bind [window|global] [command|edit|all] <key> <command>\nBinds a keyboard mapping to a command. The binding can be to the active view for view-specific key binding or to the root view for global key bindings.",
			},
		},

		&wicore.CommandAlias{"keybind", "key_bind", nil},
	}
	for _, cmd := range cmds {
		dispatcher.Register(cmd)
	}
}
