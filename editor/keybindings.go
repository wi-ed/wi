// Copyright 2014 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package editor

import "github.com/maruel/wi/wiCore"

type keyBindings struct {
	commandMappings map[wiCore.KeyPress]string
	editMappings    map[wiCore.KeyPress]string
}

func (k *keyBindings) Set(mode wiCore.KeyboardMode, key wiCore.KeyPress, cmdName string) bool {
	if !key.IsValid() {
		return false
	}
	var ok bool
	if mode == wiCore.AllMode || mode == wiCore.CommandMode {
		_, ok = k.commandMappings[key]
		k.commandMappings[key] = cmdName
	}
	if mode == wiCore.AllMode || mode == wiCore.EditMode {
		_, ok = k.editMappings[key]
		k.editMappings[key] = cmdName
	}
	return !ok
}

func (k *keyBindings) Get(mode wiCore.KeyboardMode, key wiCore.KeyPress) string {
	if !key.IsValid() {
		return ""
	}
	if mode == wiCore.CommandMode {
		return k.commandMappings[key]
	}
	if mode == wiCore.EditMode {
		return k.editMappings[key]
	}
	v, ok := k.commandMappings[key]
	if !ok {
		return k.editMappings[key]
	}
	return v
}

func makeKeyBindings() wiCore.KeyBindings {
	return &keyBindings{make(map[wiCore.KeyPress]string), make(map[wiCore.KeyPress]string)}
}

// Commands.

func cmdKeyBind(c *wiCore.CommandImpl, cd wiCore.CommandDispatcherFull, w wiCore.Window, args ...string) {
	location := args[0]
	modeName := args[1]
	keyName := args[2]
	cmdName := args[3]

	if location == "global" {
		w = wiCore.RootWindow(w)
	} else if location != "window" {
		cmd := wiCore.GetCommand(cd, w, "key_bind")
		cd.ExecuteCommand(w, "alert", cmd.LongDesc(cd, w))
		return
	}

	var mode wiCore.KeyboardMode
	if modeName == "command" {
		mode = wiCore.CommandMode
	} else if modeName == "edit" {
		mode = wiCore.CommandMode
	} else if modeName == "all" {
		mode = wiCore.AllMode
	} else {
		cmd := wiCore.GetCommand(cd, w, "key_bind")
		cd.ExecuteCommand(w, "alert", cmd.LongDesc(cd, w))
		return
	}
	// TODO(maruel): Refuse invalid keyName.
	key := wiCore.KeyPressFromString(keyName)
	w.View().KeyBindings().Set(mode, key, cmdName)
}

// RegisterKeyBindingCommands registers the keyboard mapping related commands.
func RegisterKeyBindingCommands(dispatcher wiCore.Commands) {
	cmds := []wiCore.Command{
		&wiCore.CommandImpl{
			"key_bind",
			4,
			cmdKeyBind,
			wiCore.CommandsCategory,
			wiCore.LangMap{
				wiCore.LangEn: "Binds a keyboard mapping to a command",
			},
			wiCore.LangMap{
				wiCore.LangEn: "Usage: key_bind [window|global] [command|edit|all] <key> <command>\nBinds a keyboard mapping to a command. The binding can be to the active view for view-specific key binding or to the root view for global key bindings.",
			},
		},

		&wiCore.CommandAlias{"keybind", "key_bind", nil},
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
func RegisterDefaultKeyBindings(cd wiCore.CommandDispatcher) {
	wiCore.PostCommand(cd, nil, "key_bind", "global", "all", "F1", "help")
	wiCore.PostCommand(cd, nil, "key_bind", "global", "command", ":", "show_command_window")
	// TODO(maruel): Temporary.
	wiCore.PostCommand(cd, nil, "key_bind", "global", "all", "Ctrl-c", "quit")
}
