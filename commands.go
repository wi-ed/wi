// Copyright 2014 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package main

import (
	"github.com/maruel/wi/wi-plugin"
	"log"
)

type command struct {
	handler   wi.CommandHandler
	category  wi.CommandCategory
	shortDesc string
	longDesc  string
}

func (c *command) Handle(w wi.Window, args ...string) {
	c.handler(w, args...)
}

func (c *command) Category() wi.CommandCategory {
	return c.category
}

func (c *command) ShortDesc() string {
	return c.shortDesc
}

func (c *command) LongDesc() string {
	return c.longDesc
}

// commandAlias references another command.
type commandAlias struct {
	command string
}

// Alias looks up the Commands to find the aliased command, so it can
// return the relevant details.
func (c *commandAlias) Alias(w wi.Window) wi.Command {
	return wi.GetCommandWindow(w, c.command)
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

func cmdAlert(w wi.Window, args ...string) {
	wi.RootWindow(w).NewChildWindow(makeView(1, -1), wi.DockingFloating)
	log.Printf("Faking an alert: %s", args)
}

func cmdOpen(w wi.Window, args ...string) {
	log.Printf("Faking opening a file: %s", args)
}

func cmdNew(w wi.Window, args ...string) {
	if len(args) != 0 {
		wi.ExecuteCommandWindow(w, "alert", "Command 'new' doesn't accept arguments")
	} else {
		w.NewChildWindow(makeView(-1, -1), wi.DockingFill)
	}
}

func cmdShell(w wi.Window, args ...string) {
	log.Printf("Faking opening a new shell: %s", args)
}

func cmdDoc(w wi.Window, args ...string) {
	// TODO: MakeWindow(Bottom)
	docArgs := make([]string, len(args)+1)
	docArgs[0] = "doc"
	copy(docArgs[1:], args)
	//dispatcher.Execute(w, "shell", docArgs...)
}

func cmdQuit(w wi.Window, args ...string) {
	// For all the View, question if fine to quit.
	// If not fine, "prompt" y/n to force quit. If n, stop there.
	// - Send a signal to each plugin.
	// - Send a signal back to the main loop.
	log.Printf("Faking quit: %s", args)
}

func cmdHelp(w wi.Window, args ...string) {
	// Creates a new Window with a ViewHelp.
	log.Printf("Faking help: %s", args)
}

var defaultCommands = map[string]wi.Command{
	"alert": &command{
		cmdAlert,
		wi.WindowCategory,
		"Shows a modal message",
		"Prints a message in a modal dialog box.",
	},
	"open": &command{
		cmdOpen,
		wi.WindowCategory,
		"Opens a file in a new buffer",
		"Opens a file in a new buffer.",
	},
	"new": &command{
		cmdNew,
		wi.WindowCategory,
		"Create a new buffer",
		"Create a new buffer.",
	},

	// Editor process lifetime management.
	"quit": &command{
		cmdQuit,
		wi.WindowCategory,
		"Quits",
		"Quits the editor. Optionally bypasses writing the files to disk.",
	},

	// High level commands.
	"shell": &command{
		cmdShell,
		wi.WindowCategory,
		"Opens a shell process",
		"Opens a shell process in a new buffer.",
	},
	"doc": &command{
		cmdDoc,
		wi.WindowCategory,
		"Search godoc documentation",
		"Uses the 'doc' tool to get documentation about the text under the cursor.",
	},
	"help": &command{
		cmdHelp,
		wi.WindowCategory,
		"Prints help",
		"Prints general help or help for a particular command.",
	},
	// DIRECTION = up/down/left/right
	// window_DIRECTION
	// window_close
	// cursor_move_DIRECTION
	// add_text/insert/delete
	// undo/redo
	// verb/movement/multiplier
	// Modes, select (both column and normal), command.
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
