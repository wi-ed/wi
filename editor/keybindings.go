// Copyright 2014 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package editor

import "github.com/maruel/wi/wi_core"

type keyBindings struct {
	commandMappings map[string]string
	editMappings    map[string]string
}

func (k *keyBindings) Set(mode wi_core.KeyboardMode, keyName string, cmdName string) bool {
	var ok bool
	if mode == wi_core.AllMode || mode == wi_core.CommandMode {
		_, ok = k.commandMappings[keyName]
		k.commandMappings[keyName] = cmdName
	}
	if mode == wi_core.AllMode || mode == wi_core.EditMode {
		_, ok = k.editMappings[keyName]
		k.editMappings[keyName] = cmdName
	}
	return !ok
}

func (k *keyBindings) Get(mode wi_core.KeyboardMode, keyName string) string {
	if mode == wi_core.CommandMode {
		return k.commandMappings[keyName]
	}
	if mode == wi_core.EditMode {
		return k.editMappings[keyName]
	}
	v, ok := k.commandMappings[keyName]
	if !ok {
		return k.editMappings[keyName]
	}
	return v
}

func makeKeyBindings() wi_core.KeyBindings {
	return &keyBindings{make(map[string]string), make(map[string]string)}
}

// Commands.

func cmdKeyBind(c *wi_core.CommandImpl, cd wi_core.CommandDispatcherFull, w wi_core.Window, args ...string) {
	location := args[0]
	modeName := args[1]
	keyName := args[2]
	cmdName := args[3]

	if location == "global" {
		w = wi_core.RootWindow(w)
	} else if location != "window" {
		cmd := wi_core.GetCommand(cd, w, "key_bind")
		cd.ExecuteCommand(w, "alert", cmd.LongDesc(cd, w))
		return
	}

	var mode wi_core.KeyboardMode
	if modeName == "command" {
		mode = wi_core.CommandMode
	} else if modeName == "edit" {
		mode = wi_core.CommandMode
	} else if modeName == "all" {
		mode = wi_core.AllMode
	} else {
		cmd := wi_core.GetCommand(cd, w, "key_bind")
		cd.ExecuteCommand(w, "alert", cmd.LongDesc(cd, w))
		return
	}
	// TODO(maruel): Refuse invalid keyName.
	w.View().KeyBindings().Set(mode, keyName, cmdName)
}

// RegisterKeyBindingCommands registers the keyboard mapping related commands.
func RegisterKeyBindingCommands(dispatcher wi_core.Commands) {
	cmds := []wi_core.Command{
		&wi_core.CommandImpl{
			"key_bind",
			4,
			cmdKeyBind,
			wi_core.CommandsCategory,
			wi_core.LangMap{
				wi_core.LangEn: "Binds a keyboard mapping to a command",
			},
			wi_core.LangMap{
				wi_core.LangEn: "Usage: key_bind [window|global] [command|edit|all] <key> <command>\nBinds a keyboard mapping to a command. The binding can be to the active view for view-specific key binding or to the root view for global key bindings.",
			},
		},

		&wi_core.CommandAlias{"keybind", "key_bind", nil},
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
func RegisterDefaultKeyBindings(cd wi_core.CommandDispatcher) {
	wi_core.PostCommand(cd, "key_bind", "global", "all", "F1", "help")
	wi_core.PostCommand(cd, "key_bind", "global", "command", ":", "show_command_window")
	// TODO(maruel): Temporary.
	wi_core.PostCommand(cd, "key_bind", "global", "all", "Ctrl-c", "quit")
}
