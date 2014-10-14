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

	"github.com/maruel/wi/wi_core"
)

func commandLogRecurse(w *window, cd wi_core.CommandDispatcherFull) {
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

func keyLogRecurse(w *window, cd wi_core.CommandDispatcherFull, mode wi_core.KeyboardMode) {
	// TODO(maruel): Create a proper enumerator.
	keys := w.view.KeyBindings().(*keyBindings)
	var mapping *map[string]string
	if mode == wi_core.CommandMode {
		mapping = &keys.commandMappings
	} else if mode == wi_core.EditMode {
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
	keyLogRecurse(e.rootWindow, e, wi_core.CommandMode)
	log.Printf("EditMode commands")
	keyLogRecurse(e.rootWindow, e, wi_core.EditMode)
}

func cmdLogAll(c *wi_core.CommandImpl, cd wi_core.CommandDispatcherFull, w wi_core.Window, args ...string) {
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

func cmdWindowLog(c *wi_core.CommandImpl, cd wi_core.CommandDispatcherFull, w wi_core.Window, args ...string) {
	root := wi_core.RootWindow(w)
	log.Printf("Window tree:\n%s", root.Tree())
}

func RegisterDebugCommands(dispatcher wi_core.Commands) {
	cmds := []wi_core.Command{
		&privilegedCommandImpl{
			"command_log",
			0,
			cmdCommandLog,
			wi_core.DebugCategory,
			wi_core.LangMap{
				wi_core.LangEn: "Logs the registered commands",
			},
			wi_core.LangMap{
				wi_core.LangEn: "Logs the registered commands, this is only relevant if -verbose is used.",
			},
		},
		&privilegedCommandImpl{
			"key_log",
			0,
			cmdKeyLog,
			wi_core.DebugCategory,
			wi_core.LangMap{
				wi_core.LangEn: "Logs the key bindings",
			},
			wi_core.LangMap{
				wi_core.LangEn: "Logs the key bindings, this is only relevant if -verbose is used.",
			},
		},
		&wi_core.CommandImpl{
			"log_all",
			0,
			cmdLogAll,
			wi_core.DebugCategory,
			wi_core.LangMap{
				wi_core.LangEn: "Logs the internal state (commands, view factories, windows)",
			},
			wi_core.LangMap{
				wi_core.LangEn: "Logs the internal state (commands, view factories, windows), this is only relevant if -verbose is used.",
			},
		},
		&privilegedCommandImpl{
			"view_log",
			0,
			cmdViewLog,
			wi_core.DebugCategory,
			wi_core.LangMap{
				wi_core.LangEn: "Logs the view factories",
			},
			wi_core.LangMap{
				wi_core.LangEn: "Logs the view factories, this is only relevant if -verbose is used.",
			},
		},
		&wi_core.CommandImpl{
			"window_log",
			0,
			cmdWindowLog,
			wi_core.DebugCategory,
			wi_core.LangMap{
				wi_core.LangEn: "Logs the window tree",
			},
			wi_core.LangMap{
				wi_core.LangEn: "Logs the window tree, this is only relevant if -verbose is used.",
			},
		},

		// 'editor_screenshot', mainly for unit test; open a new buffer with the screenshot, so it can be saved with 'w'.
	}
	for _, cmd := range cmds {
		dispatcher.Register(cmd)
	}
}
