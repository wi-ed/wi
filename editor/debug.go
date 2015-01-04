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

	"github.com/maruel/wi/pkg/key"
	"github.com/maruel/wi/pkg/lang"
	"github.com/maruel/wi/wicore"
)

func commandLogRecurse(w wicore.Window) {
	cmds := w.View().Commands()
	names := cmds.GetNames()
	sort.Strings(names)
	for _, name := range names {
		c := cmds.Get(name)
		log.Printf("  %s  %s: %s", w.ID(), c.Name(), c.ShortDesc())
	}
	for _, child := range w.ChildrenWindows() {
		commandLogRecurse(child)
	}
}

func cmdCommandLog(c *wicore.CommandImpl, e wicore.Editor, w wicore.Window, args ...string) {
	// Start at the root and recurse.
	commandLogRecurse(wicore.RootWindow(w))
}

func keyLogRecurse(w wicore.Window, e wicore.Editor, mode wicore.KeyboardMode) {
	bindings := w.View().KeyBindings()
	assigned := bindings.GetAssigned(mode)
	names := make([]string, 0, len(assigned))
	for _, k := range assigned {
		names = append(names, k.String())
	}
	sort.Strings(names)
	for _, name := range names {
		log.Printf("  %s  %s: %s", w.ID(), name, bindings.Get(mode, key.StringToPress(name)))
	}
	for _, child := range w.ChildrenWindows() {
		keyLogRecurse(child, e, mode)
	}
}

func cmdKeyLog(c *wicore.CommandImpl, e wicore.Editor, w wicore.Window, args ...string) {
	log.Printf("Normal commands")
	rootWindow := wicore.RootWindow(e.ActiveWindow())
	keyLogRecurse(rootWindow, e, wicore.Normal)
	log.Printf("Insert commands")
	keyLogRecurse(rootWindow, e, wicore.Insert)
}

func cmdLogAll(c *wicore.CommandImpl, e wicore.Editor, w wicore.Window, args ...string) {
	e.ExecuteCommand(w, "command_log")
	e.ExecuteCommand(w, "window_log")
	e.ExecuteCommand(w, "view_log")
	e.ExecuteCommand(w, "key_log")
}

func cmdViewLog(c *wicore.CommandImpl, e wicore.Editor, w wicore.Window, args ...string) {
	names := e.ViewFactoryNames()
	sort.Strings(names)
	log.Printf("View factories:")
	for _, name := range names {
		log.Printf("  %s", name)
	}
}

func cmdWindowLog(c *wicore.CommandImpl, e wicore.Editor, w wicore.Window, args ...string) {
	root := wicore.RootWindow(w)
	log.Printf("Window tree:\n%s", root.Tree())
}

// RegisterDebugCommands registers all debug related commands in Debug build.
func RegisterDebugCommands(dispatcher wicore.Commands) {
	cmds := []wicore.Command{
		&wicore.CommandImpl{
			"command_log",
			0,
			cmdCommandLog,
			wicore.DebugCategory,
			lang.Map{
				lang.En: "Logs the registered commands",
			},
			lang.Map{
				lang.En: "Logs the registered commands, this is only relevant if -verbose is used.",
			},
		},
		&wicore.CommandImpl{
			"key_log",
			0,
			cmdKeyLog,
			wicore.DebugCategory,
			lang.Map{
				lang.En: "Logs the key bindings",
			},
			lang.Map{
				lang.En: "Logs the key bindings, this is only relevant if -verbose is used.",
			},
		},
		&wicore.CommandImpl{
			"log_all",
			0,
			cmdLogAll,
			wicore.DebugCategory,
			lang.Map{
				lang.En: "Logs the internal state (commands, view factories, windows)",
			},
			lang.Map{
				lang.En: "Logs the internal state (commands, view factories, windows), this is only relevant if -verbose is used.",
			},
		},
		&wicore.CommandImpl{
			"view_log",
			0,
			cmdViewLog,
			wicore.DebugCategory,
			lang.Map{
				lang.En: "Logs the view factories",
			},
			lang.Map{
				lang.En: "Logs the view factories, this is only relevant if -verbose is used.",
			},
		},
		&wicore.CommandImpl{
			"window_log",
			0,
			cmdWindowLog,
			wicore.DebugCategory,
			lang.Map{
				lang.En: "Logs the window tree",
			},
			lang.Map{
				lang.En: "Logs the window tree, this is only relevant if -verbose is used.",
			},
		},

		// 'editor_screenshot', mainly for unit test; open a new buffer with the screenshot, so it can be saved with 'w'.
	}
	for _, cmd := range cmds {
		dispatcher.Register(cmd)
	}
}

// RegisterDebugEvents registers the debug event listeners.
func RegisterDebugEvents(e wicore.EventRegistry) {
	// TODO(maruel): Generate automatically?
	e.RegisterCommands(func(cmds wicore.EnqueuedCommands) bool {
		//log.Printf("Commands(%v)", cmds)
		return true
	})
	e.RegisterDocumentCreated(func(doc wicore.Document) bool {
		log.Printf("DocumentCreated(%s)", doc)
		return true
	})
	e.RegisterDocumentCursorMoved(func(doc wicore.Document, col, row int) bool {
		log.Printf("DocumentCursorMoved(%s, %d, %d)", doc, col, row)
		return true
	})
	e.RegisterEditorKeyboardModeChanged(func(mode wicore.KeyboardMode) bool {
		log.Printf("EditorKeyboardModeChanged(%s)", mode)
		return true
	})
	e.RegisterEditorLanguage(func(l lang.Language) bool {
		log.Printf("EditorLanguage(%s)", l)
		return true
	})
	e.RegisterTerminalResized(func() bool {
		log.Printf("TerminalResized()")
		return true
	})
	e.RegisterTerminalKeyPressed(func(key key.Press) bool {
		log.Printf("TerminalKeyPressed(%s)", key)
		return true
	})
	e.RegisterViewCreated(func(view wicore.View) bool {
		log.Printf("ViewCreated(%s)", view)
		return true
	})
	e.RegisterWindowCreated(func(window wicore.Window) bool {
		log.Printf("WindowCreated(%s)", window)
		return true
	})
	e.RegisterWindowResized(func(window wicore.Window) bool {
		log.Printf("WindowResized(%s)", window)
		return true
	})
}
