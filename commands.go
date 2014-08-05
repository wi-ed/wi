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
// TODO(maruel): It should reference it by name, not by pointer, so that
// updating the original command will have an effect on the alias too. This
// complexifies the calls, since each function will have to do a lookup on the
// command first, then return the data found.
type commandAlias struct {
	command string
}

// Alias looks up the CommandDispatcher to find the aliased command, so it can
// return the relevant details.
func (c *commandAlias) Alias(dispatcher wi.CommandDispatcher) wi.Command {
	//return dispatcher.GetCommand(command)
	return nil
}

// Dispatcher

type commandDispatcher struct {
	commands map[string]wi.Command
}

func (c *commandDispatcher) Execute(w wi.Window, cmd string, args ...string) {
	v, _ := c.commands[cmd]
	if v == nil {
		parent := w.Parent()
		if parent != nil {
			parent.View().Command().Execute(parent, cmd, args...)
		} else {
			// This is the root command, surface the error.
			c.Execute(w, "alert", "Command \""+cmd+"\" is not registered")
		}
	} else {
		v.Handle(w, args...)
	}
}

func (c *commandDispatcher) Register(name string, cmd wi.Command) bool {
	_, ok := c.commands[name]
	c.commands[name] = cmd
	return !ok
}

func MakeCommandDispatcher() wi.CommandDispatcher {
	return &commandDispatcher{make(map[string]wi.Command)}
}

// Default commands

func cmdAlert(w wi.Window, args ...string) {
	// TODO: w.Root().NewChildWindow(makeDialog(root))
	log.Printf("Faking an alert: %s", args)
}

func cmdOpen(w wi.Window, args ...string) {
	log.Printf("Faking opening a file: %s", args)
}

func cmdNew(w wi.Window, args ...string) {
	log.Printf("Faking opening a new buffer: %s", args)
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
}

// RegisterDefaultCommands registers the top-level native commands. This
// includes the window management commands, opening a new file buffer (it's a
// text editor after all) and help, quitting, etc. It doesn't includes handling
// a file buffer itself, it's up to the relevant view to add the corresponding
// commands. For example, "open" is implemented but "write" is not!
func RegisterDefaultCommands(dispatcher wi.CommandDispatcher) {
	for name, cmd := range defaultCommands {
		dispatcher.Register(name, cmd)
	}
	// DIRECTION = up/down/left/right
	// window_DIRECTION
	// window_close
	// cursor_move_DIRECTION
	// add_text/insert/delete
	// undo/redo
	// verb/movement/multiplier
	// Modes, select (both column and normal), command.
}
