// Copyright 2014 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Use "go build -tags debug" to have access to the code and commands in this
// file.

// +build debug

package editor

import (
	"log"
	"sort"

	"github.com/maruel/wi/wiCore"
)

func commandLogRecurse(w *window, cd wiCore.CommandDispatcherFull) {
	// TODO(maruel): Create a proper enumerator.
	cmds := w.view.Commands().(*commands)
	names := make([]string, 0, len(cmds.commands))
	for k := range cmds.commands {
		names = append(names, k)
	}
	sort.Strings(names)
	for _, n := range names {
		c := cmds.commands[n]
		log.Printf("  %s  %s: %s", w.ID(), c.Name(), c.ShortDesc(cd, w))
	}
	for _, child := range w.childrenWindows {
		commandLogRecurse(child, cd)
	}
}

func cmdCommandLog(c *privilegedCommandImpl, e *editor, w *window, args ...string) {
	// Start at the root and recurse.
	commandLogRecurse(e.rootWindow, e)
}

func keyLogRecurse(w *window, cd wiCore.CommandDispatcherFull, mode wiCore.KeyboardMode) {
	// TODO(maruel): Create a proper enumerator.
	keys := w.view.KeyBindings().(*keyBindings)
	var mapping *map[string]string
	if mode == wiCore.CommandMode {
		mapping = &keys.commandMappings
	} else if mode == wiCore.EditMode {
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
	keyLogRecurse(e.rootWindow, e, wiCore.CommandMode)
	log.Printf("EditMode commands")
	keyLogRecurse(e.rootWindow, e, wiCore.EditMode)
}

func cmdLogAll(c *wiCore.CommandImpl, cd wiCore.CommandDispatcherFull, w wiCore.Window, args ...string) {
	cd.ExecuteCommand(w, "command_log")
	cd.ExecuteCommand(w, "window_log")
	cd.ExecuteCommand(w, "view_log")
	cd.ExecuteCommand(w, "key_log")
}

func cmdViewLog(c *privilegedCommandImpl, e *editor, w *window, args ...string) {
	names := make([]string, 0, len(e.viewFactories))
	for k := range e.viewFactories {
		names = append(names, k)
	}
	sort.Strings(names)
	log.Printf("View factories:")
	for _, name := range names {
		log.Printf("  %s", name)
	}
}

func cmdWindowLog(c *wiCore.CommandImpl, cd wiCore.CommandDispatcherFull, w wiCore.Window, args ...string) {
	root := wiCore.RootWindow(w)
	log.Printf("Window tree:\n%s", root.Tree())
}

// RegisterDebugCommands registers all debug related commands in Debug build.
func RegisterDebugCommands(dispatcher wiCore.Commands) {
	cmds := []wiCore.Command{
		&privilegedCommandImpl{
			"command_log",
			0,
			cmdCommandLog,
			wiCore.DebugCategory,
			wiCore.LangMap{
				wiCore.LangEn: "Logs the registered commands",
			},
			wiCore.LangMap{
				wiCore.LangEn: "Logs the registered commands, this is only relevant if -verbose is used.",
			},
		},
		&privilegedCommandImpl{
			"key_log",
			0,
			cmdKeyLog,
			wiCore.DebugCategory,
			wiCore.LangMap{
				wiCore.LangEn: "Logs the key bindings",
			},
			wiCore.LangMap{
				wiCore.LangEn: "Logs the key bindings, this is only relevant if -verbose is used.",
			},
		},
		&wiCore.CommandImpl{
			"log_all",
			0,
			cmdLogAll,
			wiCore.DebugCategory,
			wiCore.LangMap{
				wiCore.LangEn: "Logs the internal state (commands, view factories, windows)",
			},
			wiCore.LangMap{
				wiCore.LangEn: "Logs the internal state (commands, view factories, windows), this is only relevant if -verbose is used.",
			},
		},
		&privilegedCommandImpl{
			"view_log",
			0,
			cmdViewLog,
			wiCore.DebugCategory,
			wiCore.LangMap{
				wiCore.LangEn: "Logs the view factories",
			},
			wiCore.LangMap{
				wiCore.LangEn: "Logs the view factories, this is only relevant if -verbose is used.",
			},
		},
		&wiCore.CommandImpl{
			"window_log",
			0,
			cmdWindowLog,
			wiCore.DebugCategory,
			wiCore.LangMap{
				wiCore.LangEn: "Logs the window tree",
			},
			wiCore.LangMap{
				wiCore.LangEn: "Logs the window tree, this is only relevant if -verbose is used.",
			},
		},

		// 'editor_screenshot', mainly for unit test; open a new buffer with the screenshot, so it can be saved with 'w'.
	}
	for _, cmd := range cmds {
		dispatcher.Register(cmd)
	}
}
