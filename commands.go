// Copyright 2014 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package main

import (
	"github.com/maruel/wi/wi-plugin"
	"log"
)

type langMap map[wi.LanguageMode]string

type command struct {
	handler   wi.CommandHandler
	category  wi.CommandCategory
	shortDesc langMap
	longDesc  langMap
}

func (c *command) Handle(cd wi.CommandDispatcherFull, w wi.Window, args ...string) {
	c.handler(cd, w, args...)
}

func (c *command) Category() wi.CommandCategory {
	return c.category
}

func (c *command) ShortDesc(lang wi.LanguageMode) string {
	desc, ok := c.shortDesc[lang]
	if !ok {
		desc = c.shortDesc[wi.LangEn]
	}
	return desc
}

func (c *command) LongDesc(lang wi.LanguageMode) string {
	desc, ok := c.longDesc[lang]
	if !ok {
		desc = c.longDesc[wi.LangEn]
	}
	return desc
}

// commandAlias references another command.
type commandAlias struct {
	command string
}

// Alias looks up the Commands to find the aliased command, so it can
// return the relevant details.
func (c *commandAlias) Alias(cd wi.CommandDispatcherFull) wi.Command {
	return wi.GetCommand(cd, nil, c.command)
}

type commands struct {
	commands map[string]wi.Command
}

func (c *commands) Register(name string, cmd wi.Command) bool {
	_, ok := c.commands[name]
	c.commands[name] = cmd
	return !ok
}

func (c *commands) Get(cmd string) wi.Command {
	return c.commands[cmd]
}

func makeCommands() wi.Commands {
	return &commands{make(map[string]wi.Command)}
}

// Default commands

func cmdAlert(cd wi.CommandDispatcherFull, w wi.Window, args ...string) {
	wi.RootWindow(w).NewChildWindow(makeView(1, -1), wi.DockingFloating)
	//w2.Activate()
}

func cmdAddStatusBar(cd wi.CommandDispatcherFull, w wi.Window, args ...string) {
	w.NewChildWindow(makeStatusView(), wi.DockingBottom)
}

func cmdOpen(cd wi.CommandDispatcherFull, w wi.Window, args ...string) {
	log.Printf("Faking opening a file: %s", args)
}

func cmdNew(cd wi.CommandDispatcherFull, w wi.Window, args ...string) {
	if len(args) != 0 {
		cd.PostCommand("alert", "Command 'new' doesn't accept arguments")
	} else {
		w.NewChildWindow(makeView(-1, -1), wi.DockingFill)
	}
}

func cmdShell(cd wi.CommandDispatcherFull, w wi.Window, args ...string) {
	log.Printf("Faking opening a new shell: %s", args)
}

func cmdDoc(cd wi.CommandDispatcherFull, w wi.Window, args ...string) {
	// TODO: MakeWindow(Bottom)
	docArgs := make([]string, len(args)+1)
	docArgs[0] = "doc"
	copy(docArgs[1:], args)
	//dispatcher.Execute(w, "shell", docArgs...)
}

func cmdQuit(cd wi.CommandDispatcherFull, w wi.Window, args ...string) {
	// TODO(maruel): For all the View, question if fine to quit via
	// view.IsDirty(). If not fine, "prompt" y/n to force quit. If n, stop there.
	// - Send a signal to each plugin.
	// - Send a signal back to the main loop.
	log.Printf("Faking quit: %s", args)
}

func cmdHelp(cd wi.CommandDispatcherFull, w wi.Window, args ...string) {
	// TODO(maruel): Creates a new Window with a ViewHelp.
	log.Printf("Faking help: %s", args)
}

func cmdKeyBind(cd wi.CommandDispatcherFull, w wi.Window, args ...string) {
	if len(args) != 3 {
		cmd := wi.GetCommand(cd, nil, "keybind")
		cd.PostCommand("alert", cmd.LongDesc(cd.CurrentLanguage()))
	}
	var mode wi.KeyboardMode
	if args[0] == "command" {
		mode = wi.CommandMode
	} else if args[0] == "edit" {
		mode = wi.CommandMode
	} else if args[0] == "all" {
		mode = wi.AllMode
	} else {
		cmd := wi.GetCommand(cd, nil, "keybind")
		cd.PostCommand("alert", cmd.LongDesc(cd.CurrentLanguage()))
	}
	w.View().KeyBindings().Set(mode, args[1], args[2])
}

var defaultCommands = map[string]wi.Command{
	"alert": &command{
		cmdAlert,
		wi.WindowCategory,
		langMap{
			wi.LangEn: "Shows a modal message",
		},
		langMap{
			wi.LangEn: "Prints a message in a modal dialog box.",
		},
	},
	"add_status_bar": &command{
		cmdAddStatusBar,
		wi.WindowCategory,
		langMap{
			wi.LangEn: "Adds the standard status bar",
		},
		langMap{
			wi.LangEn: "Adds the standard status bar to the active window. This command exists so it can be overriden by a plugin, so it can create its own status bar.",
		},
	},
	"open": &command{
		cmdOpen,
		wi.WindowCategory,
		langMap{
			wi.LangEn: "Opens a file in a new buffer",
		},
		langMap{
			wi.LangEn: "Opens a file in a new buffer.",
		},
	},
	"new": &command{
		cmdNew,
		wi.WindowCategory,
		langMap{
			wi.LangEn: "Create a new buffer",
		},
		langMap{
			wi.LangEn: "Create a new buffer.",
		},
	},

	// Editor process lifetime management.
	"quit": &command{
		cmdQuit,
		wi.WindowCategory,
		langMap{
			wi.LangEn: "Quits",
		},
		langMap{
			wi.LangEn: "Quits the editor. Optionally bypasses writing the files to disk.",
		},
	},

	// High level commands.
	"shell": &command{
		cmdShell,
		wi.WindowCategory,
		langMap{
			wi.LangEn: "Opens a shell process",
		},
		langMap{
			wi.LangEn: "Opens a shell process in a new buffer.",
		},
	},
	"doc": &command{
		cmdDoc,
		wi.WindowCategory,
		langMap{
			wi.LangEn: "Search godoc documentation",
		},
		langMap{
			wi.LangEn: "Uses the 'doc' tool to get documentation about the text under the cursor.",
		},
	},
	"help": &command{
		cmdHelp,
		wi.WindowCategory,
		langMap{
			wi.LangEn: "Prints help",
		},
		langMap{
			wi.LangEn: "Prints general help or help for a particular command.",
		},
	},
	"keybind": &command{
		cmdKeyBind,
		wi.CommandsCategory,
		langMap{
			wi.LangEn: "Binds a keyboard mapping to a command",
		},
		langMap{
			wi.LangEn: "Usage: keybind [command|edit|all] <key> <command>\nBinds a keyboard mapping to a command. The binding can be to the active view for view-specific key binding or to the root view for global key bindings.",
		},
	},
	// DIRECTION = up/down/left/right
	// window_DIRECTION
	// window_close
	// cursor_move_DIRECTION
	// add_text/insert/delete
	// undo/redo
	// verb/movement/multiplier
	// Modes, select (both column and normal), command.
	// keybind global all Ctrl-C quit
	// keybind active edit <left> move_left
	// ...
}

// RegisterDefaultCommands registers the top-level native commands. This
// includes the window management commands, opening a new file buffer (it's a
// text editor after all) and help, quitting, etc. It doesn't includes handling
// a file buffer itself, it's up to the relevant view to add the corresponding
// commands. For example, "open" is implemented but "write" is not!
func RegisterDefaultCommands(dispatcher wi.Commands) {
	for name, cmd := range defaultCommands {
		dispatcher.Register(name, cmd)
	}
}
