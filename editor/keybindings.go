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
	commandMappings map[key.Press]string
	editMappings    map[key.Press]string
}

func (k *keyBindings) Set(mode wicore.KeyboardMode, key key.Press, cmdName string) bool {
	if !key.IsValid() {
		return false
	}
	var ok bool
	if mode == wicore.AllMode || mode == wicore.CommandMode {
		_, ok = k.commandMappings[key]
		k.commandMappings[key] = cmdName
	}
	if mode == wicore.AllMode || mode == wicore.EditMode {
		_, ok = k.editMappings[key]
		k.editMappings[key] = cmdName
	}
	return !ok
}

func (k *keyBindings) Get(mode wicore.KeyboardMode, key key.Press) string {
	if !key.IsValid() {
		return ""
	}
	if mode == wicore.CommandMode {
		return k.commandMappings[key]
	}
	if mode == wicore.EditMode {
		return k.editMappings[key]
	}
	v, ok := k.commandMappings[key]
	if !ok {
		return k.editMappings[key]
	}
	return v
}

func makeKeyBindings() wicore.KeyBindings {
	return &keyBindings{make(map[key.Press]string), make(map[key.Press]string)}
}

// Commands.

func cmdKeyBind(c *wicore.CommandImpl, cd wicore.Editor, w wicore.Window, args ...string) {
	location := args[0]
	modeName := args[1]
	keyName := args[2]
	cmdName := args[3]

	if location == "global" {
		w = wicore.RootWindow(w)
	} else if location != "window" {
		cmd := wicore.GetCommand(cd, w, "key_bind")
		cd.ExecuteCommand(w, "alert", cmd.LongDesc(cd, w))
		return
	}

	var mode wicore.KeyboardMode
	if modeName == "command" {
		mode = wicore.CommandMode
	} else if modeName == "edit" {
		mode = wicore.CommandMode
	} else if modeName == "all" {
		mode = wicore.AllMode
	} else {
		cmd := wicore.GetCommand(cd, w, "key_bind")
		cd.ExecuteCommand(w, "alert", cmd.LongDesc(cd, w))
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

// RegisterDefaultKeyBindings registers the default keyboard mapping. Keyboard
// mapping simply execute the corresponding command. So to add a keyboard map,
// the corresponding command needs to be added first.
//
// TODO(maruel): This should be remappable via a configuration flag, for
// example vim flavor vs emacs flavor. I'm not sure it's worth supporting this
// without a restart.
func RegisterDefaultKeyBindings(e wicore.EventRegistry) {
	wicore.PostCommand(e, nil, "key_bind", "global", "all", "F1", "help")
	wicore.PostCommand(e, nil, "key_bind", "global", "command", ":", "show_command_window")
	// TODO(maruel): Temporary.
	wicore.PostCommand(e, nil, "key_bind", "global", "all", "Ctrl-c", "quit")
}
