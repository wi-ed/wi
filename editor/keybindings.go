// Copyright 2014 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package editor

import (
	"log"
	"sort"

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

func cmdKeyBind(c *wi.CommandImpl, cd wi.CommandDispatcherFull, w wi.Window, args ...string) {
	location := args[0]
	modeName := args[1]
	keyName := args[2]
	cmdName := args[3]

	if location == "global" {
		w = wi.RootWindow(w)
	} else if location != "window" {
		cmd := wi.GetCommand(cd, w, "key_bind")
		cd.ExecuteCommand(w, "alert", cmd.LongDesc(cd, w))
		return
	}

	var mode wi.KeyboardMode
	if modeName == "command" {
		mode = wi.CommandMode
	} else if modeName == "edit" {
		mode = wi.CommandMode
	} else if modeName == "all" {
		mode = wi.AllMode
	} else {
		cmd := wi.GetCommand(cd, w, "key_bind")
		cd.ExecuteCommand(w, "alert", cmd.LongDesc(cd, w))
		return
	}
	// TODO(maruel): Refuse invalid keyName.
	w.View().KeyBindings().Set(mode, keyName, cmdName)
}

func keyLogRecurse(w *window, cd wi.CommandDispatcherFull, mode wi.KeyboardMode) {
	// TODO(maruel): Create a proper enumerator.
	keys := w.view.KeyBindings().(*keyBindings)
	var mapping *map[string]string
	if mode == wi.CommandMode {
		mapping = &keys.commandMappings
	} else if mode == wi.EditMode {
		mapping = &keys.editMappings
	} else {
		panic("Errr, fix me")
	}
	names := make([]string, 0, len(*mapping))
	for k := range *mapping {
		names = append(names, k)
	}
	sort.Strings(names)
	for _, name := range names {
		log.Printf("  %s  %s: %s", w.ID(), name, (*mapping)[name])
	}
	for _, child := range w.childrenWindows {
		keyLogRecurse(child, cd, mode)
	}
}

func cmdKeyLog(c *privilegedCommandImpl, e *editor, w *window, args ...string) {
	log.Printf("CommandMode commands")
	keyLogRecurse(e.rootWindow, e, wi.CommandMode)
	log.Printf("EditMode commands")
	keyLogRecurse(e.rootWindow, e, wi.EditMode)
}

// RegisterKeyBindingCommands registers the keyboard mapping related commands.
func RegisterKeyBindingCommands(dispatcher wi.Commands) {
	defaultCommands := []wi.Command{
		&wi.CommandImpl{
			"key_bind",
			4,
			cmdKeyBind,
			wi.CommandsCategory,
			wi.LangMap{
				wi.LangEn: "Binds a keyboard mapping to a command",
			},
			wi.LangMap{
				wi.LangEn: "Usage: key_bind [window|global] [command|edit|all] <key> <command>\nBinds a keyboard mapping to a command. The binding can be to the active view for view-specific key binding or to the root view for global key bindings.",
			},
		},
		&privilegedCommandImpl{
			"key_log",
			0,
			cmdKeyLog,
			wi.DebugCategory,
			wi.LangMap{
				wi.LangEn: "Logs the key bindings",
			},
			wi.LangMap{
				wi.LangEn: "Logs the key bindings, this is only relevant if -verbose is used.",
			},
		},

		&wi.CommandAlias{"keybind", "key_bind", nil},
	}
	for _, cmd := range defaultCommands {
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
func RegisterDefaultKeyBindings(cd wi.CommandDispatcher) {
	wi.PostCommand(cd, "key_bind", "global", "all", "F1", "help")
	wi.PostCommand(cd, "key_bind", "global", "command", ":", "show_command_window")
	// TODO(maruel): Temporary.
	wi.PostCommand(cd, "key_bind", "global", "all", "Ctrl-c", "quit")
}
