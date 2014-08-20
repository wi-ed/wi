// Copyright 2014 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package main

import (
	"fmt"
	"github.com/maruel/wi/wi-plugin"
	"log"
	"time"
)

type command struct {
	handler   wi.CommandHandler
	category  wi.CommandCategory
	shortDesc langMap
	longDesc  langMap
}

// Handle runs the handler.
func (c *command) Handle(cd wi.CommandDispatcherFull, w wi.Window, args ...string) {
	c.handler(cd, w, args...)
}

func (c *command) Category(cd wi.CommandDispatcherFull) wi.CommandCategory {
	return c.category
}

func (c *command) ShortDesc(cd wi.CommandDispatcherFull) string {
	return getStr(cd.CurrentLanguage(), c.shortDesc)
}

func (c *command) LongDesc(cd wi.CommandDispatcherFull) string {
	return getStr(cd.CurrentLanguage(), c.longDesc)
}

// commandAlias references another command.
type commandAlias struct {
	command string
}

func (c *commandAlias) Handle(cd wi.CommandDispatcherFull, w wi.Window, args ...string) {
	cd.ExecuteCommand(w, c.command, args...)
}

func (c *commandAlias) Category(cd wi.CommandDispatcherFull) wi.CommandCategory {
	cmd := wi.GetCommand(cd, nil, c.command)
	if cmd != nil {
		return c.Category(cd)
	}
	return wi.UnknownCategory
}

func (c *commandAlias) ShortDesc(cd wi.CommandDispatcherFull) string {
	return fmt.Sprintf(getStr(cd.CurrentLanguage(), aliasFor), c.command)
}

func (c *commandAlias) LongDesc(cd wi.CommandDispatcherFull) string {
	return fmt.Sprintf(getStr(cd.CurrentLanguage(), aliasFor), c.command)
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
	// TODO(maruel): Create an infobar that automatically dismiss itself after 5s.
	if len(args) != 1 {
		cmd := wi.GetCommand(cd, nil, "alert")
		cd.ExecuteCommand(w, "alert", cmd.LongDesc(cd))
	}
	wi.RootWindow(w).NewChildWindow(makeAlertView(args[0]), wi.DockingFloating)
	//w2.Activate()
	go func() {
		<-time.After(5 * time.Second)
		// TODO(maruel): Dismiss.
	}()
}

func cmdAddStatusBar(cd wi.CommandDispatcherFull, w wi.Window, args ...string) {
	// Create a tree of views that is used for alignment.
	w2 := w.NewChildWindow(makeStatusViewCenter(), wi.DockingBottom)
	w2.NewChildWindow(makeStatusViewName(), wi.DockingLeft)
	w2.NewChildWindow(makeStatusViewPosition(), wi.DockingRight)
}

func cmdOpen(cd wi.CommandDispatcherFull, w wi.Window, args ...string) {
	// The Window and View are created synchronously. The View is populated
	// asynchronously.
	log.Printf("Faking opening a file: %s", args)
}

func cmdNew(cd wi.CommandDispatcherFull, w wi.Window, args ...string) {
	if len(args) != 0 {
		cmd := wi.GetCommand(cd, nil, "alias")
		cd.ExecuteCommand(w, "alert", cmd.LongDesc(cd))
	} else {
		w.NewChildWindow(makeView("New doc", -1, -1), wi.DockingFill)
	}
}

func cmdShell(cd wi.CommandDispatcherFull, w wi.Window, args ...string) {
	log.Printf("Faking opening a new shell: %s", args)
}

func cmdDoc(cd wi.CommandDispatcherFull, w wi.Window, args ...string) {
	// TODO(maruel): Grab the current word under selection if no args is
	// provided. Pass this token to shell.
	docArgs := make([]string, len(args)+1)
	docArgs[0] = "doc"
	copy(docArgs[1:], args)
	//dispatcher.Execute(w, "shell", docArgs...)
}

func isDirtyRecurse(cd wi.CommandDispatcherFull, w wi.Window) bool {
	for _, child := range w.ChildrenWindows() {
		if isDirtyRecurse(cd, child) {
			return true
		}
		v := child.View()
		if v.IsDirty() {
			cd.ExecuteCommand(w, "alert", fmt.Sprintf(getStr(cd.CurrentLanguage(), viewDirty), v.Title()))
			return true
		}
	}
	return false
}

func cmdQuit(cd wi.CommandDispatcherFull, w wi.Window, args ...string) {
	// TODO(maruel): For all the View, question if fine to quit via
	// view.IsDirty(). If not fine, "prompt" y/n to force quit. If n, stop there.
	// - Send a signal to each plugin.
	// - Send a signal back to the main loop.
	root := wi.RootWindow(w)
	if !isDirtyRecurse(cd, root) {
		quitFlag = true
		// ViewReady will wake up the command event loop so it detects it's time to
		// quit.
		cd.ViewReady(w.View())
	}
}

func cmdHelp(cd wi.CommandDispatcherFull, w wi.Window, args ...string) {
	// TODO(maruel): Creates a new Window with a ViewHelp.
	log.Printf("Faking help: %s", args)
}

func cmdAlias(cd wi.CommandDispatcherFull, w wi.Window, args ...string) {
	if len(args) != 3 {
		cmd := wi.GetCommand(cd, nil, "alias")
		cd.ExecuteCommand(w, "alert", cmd.LongDesc(cd))
		return
	}
	if args[0] == "window" {
	} else if args[0] == "global" {
		w = wi.RootWindow(w)
	} else {
		cmd := wi.GetCommand(cd, nil, "alias")
		cd.ExecuteCommand(w, "alert", cmd.LongDesc(cd))
		return
	}
	cmdName := args[2]
	cmd := wi.GetCommand(cd, w, cmdName)
	if cmd == nil {
		cd.ExecuteCommand(w, "alert", fmt.Sprintf(getStr(cd.CurrentLanguage(), notFound), cmdName))
		return
	}
	w.View().Commands().Register(args[1], cmd)
}

func cmdKeyBind(cd wi.CommandDispatcherFull, w wi.Window, args ...string) {
	if len(args) != 4 {
		cmd := wi.GetCommand(cd, nil, "keybind")
		cd.ExecuteCommand(w, "alert", cmd.LongDesc(cd))
		return
	}
	location := args[0]
	modeName := args[1]
	keyName := args[2]
	cmdName := args[3]

	if location == "global" {
		w = wi.RootWindow(w)
	} else if location != "window" {
		cmd := wi.GetCommand(cd, nil, "keybind")
		cd.ExecuteCommand(w, "alert", cmd.LongDesc(cd))
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
		cmd := wi.GetCommand(cd, nil, "keybind")
		cd.ExecuteCommand(w, "alert", cmd.LongDesc(cd))
		return
	}
	w.View().KeyBindings().Set(mode, keyName, cmdName)
}

func cmdShowCommandWindow(cd wi.CommandDispatcherFull, w wi.Window, args ...string) {
	if len(args) != 0 {
		cmd := wi.GetCommand(cd, nil, "show_command_window")
		cd.ExecuteCommand(w, "alert", cmd.LongDesc(cd))
		return
	}

	// Create the Window with the command view and attach it to the currently
	// focused Window.
	cmdWindow := makeCommandView()
	w.NewChildWindow(cmdWindow, wi.DockingFloating)
}

// Native commands.
var defaultCommands = map[string]wi.Command{

	// WindowCategory

	// TODO(maruel): Use a 5 seconds infobar.
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
	"q": &commandAlias{"quit"},

	// CommandsCategory

	"alias": &command{
		cmdAlias,
		wi.CommandsCategory,
		langMap{
			wi.LangEn: "Binds an alias to another command",
		},
		langMap{
			// TODO(maruel): For complex aliasing, use macro?
			wi.LangEn: "Usage: alias [window|global] <alias> <name>\nBinds an alias to another command. The alias can either be local to the window or global",
		},
	},
	"keybind": &command{
		cmdKeyBind,
		wi.CommandsCategory,
		langMap{
			wi.LangEn: "Binds a keyboard mapping to a command",
		},
		langMap{
			wi.LangEn: "Usage: keybind [window|global] [command|edit|all] <key> <command>\nBinds a keyboard mapping to a command. The binding can be to the active view for view-specific key binding or to the root view for global key bindings.",
		},
	},
	"show_command_window": &command{
		cmdShowCommandWindow,
		wi.CommandsCategory,
		langMap{
			wi.LangEn: "Shows the interactive command window",
		},
		langMap{
			wi.LangEn: "This commands exists so it can be bound to a key to pop up the interactive command window.",
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
	// 'screenshot', mainly for unit test; open a new buffer with the screenshot, so it can be saved with 'w'.
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
