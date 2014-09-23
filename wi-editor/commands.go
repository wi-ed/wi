// Copyright 2014 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package editor

import (
	"fmt"
	"github.com/maruel/wi/wi-plugin"
	"log"
	"sort"
)

// commands is the map of registered commands.
type commands struct {
	commands map[string]wi.Command
}

func (c *commands) Register(cmd wi.Command) bool {
	name := cmd.Name()
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

// privilegedCommandImplHandler is the CommandHandler to use when coupled with
// privilegedCommandImpl.
type privilegedCommandImplHandler func(c *privilegedCommandImpl, e *editor, w *window, args ...string)

// privilegedCommandImpl is the boilerplate Command implementation for builtin
// commands that can access the editor directly.
//
// Native (builtin) commands can mutate the editor.
//
// This command handler has access to the internals of the editor. Because of
// this, it can only be native commands inside the editor process.
type privilegedCommandImpl struct {
	NameValue      string
	ExpectedArgs   int // If >= 0, the command will be aborted if the number of arguments is not exactly this value. Set to -1 to disable verification. On abort, an alert with the long description of the command is done.
	HandlerValue   privilegedCommandImplHandler
	CategoryValue  wi.CommandCategory
	ShortDescValue wi.LangMap
	LongDescValue  wi.LangMap
}

func (c *privilegedCommandImpl) Name() string {
	return c.NameValue
}

func (c *privilegedCommandImpl) Handle(cd wi.CommandDispatcherFull, w wi.Window, args ...string) {
	if c.ExpectedArgs != -1 && len(args) != c.ExpectedArgs {
		cd.ExecuteCommand(w, "alert", c.LongDesc(cd, w))
	}
	// Convert types to internal types.
	e := cd.(*editor)
	wInternal := w.(*window)
	c.HandlerValue(c, e, wInternal, args...)
}

func (c *privilegedCommandImpl) Category(cd wi.CommandDispatcherFull, w wi.Window) wi.CommandCategory {
	return c.CategoryValue
}

func (c *privilegedCommandImpl) ShortDesc(cd wi.CommandDispatcherFull, w wi.Window) string {
	return wi.GetStr(cd.CurrentLanguage(), c.ShortDescValue)
}

func (c *privilegedCommandImpl) LongDesc(cd wi.CommandDispatcherFull, w wi.Window) string {
	return wi.GetStr(cd.CurrentLanguage(), c.LongDescValue)
}

// Default commands

func cmdAlert(c *wi.CommandImpl, cd wi.CommandDispatcherFull, w wi.Window, args ...string) {
	cd.ExecuteCommand(w, "window_new", "0", "bottom", "infobar_alert", args[0])
}

func commandLogRecurse(w *window, cd wi.CommandDispatcherFull) {
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

func cmdLogAll(c *wi.CommandImpl, cd wi.CommandDispatcherFull, w wi.Window, args ...string) {
	cd.ExecuteCommand(w, "command_log")
	cd.ExecuteCommand(w, "window_log")
	cd.ExecuteCommand(w, "view_log")
	cd.ExecuteCommand(w, "key_log")
}

func cmdEditorBootstrapUI(c *wi.CommandImpl, cd wi.CommandDispatcherFull, w wi.Window, args ...string) {
	// TODO(maruel): Use onAttach instead of hard coding names.
	cd.ExecuteCommand(w, "window_new", "0", "bottom", "status_root")
	cd.ExecuteCommand(w, "window_new", "0:1", "left", "status_name")
	cd.ExecuteCommand(w, "window_new", "0:1", "right", "status_position")
}

func cmdDocumentNew(c *wi.CommandImpl, cd wi.CommandDispatcherFull, w wi.Window, args ...string) {
	cmd := make([]string, 3+len(args))
	cmd[0] = w.ID()
	cmd[1] = "fill"
	cmd[2] = "new_document"
	copy(cmd[3:], args)
	cd.ExecuteCommand(w, "window_new", cmd...)
}

func cmdDocumentOpen(c *wi.CommandImpl, cd wi.CommandDispatcherFull, w wi.Window, args ...string) {
	// The Window and View are created synchronously. The View is populated
	// asynchronously.
	log.Printf("Faking opening a file: %s", args)
}

func isDirtyRecurse(cd wi.CommandDispatcherFull, w wi.Window) bool {
	for _, child := range w.ChildrenWindows() {
		if isDirtyRecurse(cd, child) {
			return true
		}
		v := child.View()
		if v.IsDirty() {
			cd.ExecuteCommand(w, "alert", fmt.Sprintf(wi.GetStr(cd.CurrentLanguage(), viewDirty), v.Title()))
			return true
		}
	}
	return false
}

func cmdEditorQuit(c *privilegedCommandImpl, e *editor, w *window, args ...string) {
	// TODO(maruel): For all the View, question if fine to quit via
	// view.IsDirty(). If not fine, "prompt" y/n to force quit. If n, stop there.
	// - Send a signal to each plugin.
	// - Send a signal back to the main loop.
	if !isDirtyRecurse(e, e.rootWindow) {
		e.quitFlag = true
		// editor_redraw wakes up the command event loop so it detects it's time to
		// quit.
		wi.PostCommand(e, "editor_redraw")
	}
}

func cmdEditorRedraw(c *privilegedCommandImpl, e *editor, w *window, args ...string) {
	go func() {
		e.viewReady <- true
	}()
}

func cmdCommandAlias(c *wi.CommandImpl, cd wi.CommandDispatcherFull, w wi.Window, args ...string) {
	if args[0] == "window" {
	} else if args[0] == "global" {
		w = wi.RootWindow(w)
	} else {
		cmd := wi.GetCommand(cd, w, "command_alias")
		cd.ExecuteCommand(w, "alert", cmd.LongDesc(cd, w))
		return
	}
	alias := &wi.CommandAlias{args[1], args[2]}
	w.View().Commands().Register(alias)
}

func cmdShowCommandWindow(c *wi.CommandImpl, cd wi.CommandDispatcherFull, w wi.Window, args ...string) {
	// Create the Window with the command view and attach it to the currently
	// focused Window.
	cd.ExecuteCommand(w, "window_new", w.ID(), "floating", "command")
}

// RegisterDefaultCommands registers the top-level native commands. This
// includes the window management commands, opening a new file buffer (it's a
// text editor after all) and help, quitting, etc. It doesn't includes handling
// a file buffer itself, it's up to the relevant view to add the corresponding
// commands. For example, "open" is implemented but "write" is not!
func RegisterDefaultCommands(dispatcher wi.Commands) {
	// Native commands.
	defaultCommands := []wi.Command{
		&wi.CommandImpl{
			"alert",
			1,
			cmdAlert,
			wi.WindowCategory,
			wi.LangMap{
				wi.LangEn: "Shows a modal message",
			},
			wi.LangMap{
				wi.LangEn: "Prints a message in a modal dialog box.",
			},
		},
		&wi.CommandImpl{
			"document_new",
			-1,
			cmdDocumentNew,
			wi.WindowCategory,
			wi.LangMap{
				wi.LangEn: "Create a new buffer",
			},
			wi.LangMap{
				wi.LangEn: "Create a new buffer.",
			},
		},
		&wi.CommandImpl{
			"document_open",
			-1,
			cmdDocumentOpen,
			wi.WindowCategory,
			wi.LangMap{
				wi.LangEn: "Opens a file in a new buffer",
			},
			wi.LangMap{
				wi.LangEn: "Opens a file in a new buffer.",
			},
		},
		&wi.CommandImpl{
			"editor_bootstrap_ui",
			0,
			cmdEditorBootstrapUI,
			wi.WindowCategory,
			wi.LangMap{
				wi.LangEn: "Bootstraps the editor's UI",
			},
			wi.LangMap{
				wi.LangEn: "Bootstraps the editor's UI. This command is automatically run on startup and cannot be executed afterward. It adds the standard status bar. This command exists so it can be overriden by a plugin, so it can create its own status bar.",
			},
		},
		&privilegedCommandImpl{
			"editor_quit",
			0,
			cmdEditorQuit,
			wi.EditorCategory,
			wi.LangMap{
				wi.LangEn: "Quits",
			},
			wi.LangMap{
				wi.LangEn: "Quits the editor. Optionally bypasses writing the files to disk.",
			},
		},
		&privilegedCommandImpl{
			"editor_redraw",
			0,
			cmdEditorRedraw,
			wi.EditorCategory,
			wi.LangMap{
				wi.LangEn: "Forcibly redraws the terminal",
			},
			wi.LangMap{
				wi.LangEn: "Forcibly redraws the terminal.",
			},
		},

		&wi.CommandImpl{
			"command_alias",
			3,
			cmdCommandAlias,
			wi.CommandsCategory,
			wi.LangMap{
				wi.LangEn: "Binds an alias to another command",
			},
			wi.LangMap{
				// TODO(maruel): For complex aliasing, use macro?
				wi.LangEn: "Usage: command_alias [window|global] <alias> <name>\nBinds an alias to another command. The alias can either be local to the window or global",
			},
		},
		&privilegedCommandImpl{
			"command_log",
			0,
			cmdCommandLog,
			wi.DebugCategory,
			wi.LangMap{
				wi.LangEn: "Logs the registered commands",
			},
			wi.LangMap{
				wi.LangEn: "Logs the registered commands, this is only relevant if -verbose is used.",
			},
		},
		&wi.CommandImpl{
			"log_all",
			0,
			cmdLogAll,
			wi.DebugCategory,
			wi.LangMap{
				wi.LangEn: "Logs the internal state (commands, view factories, windows)",
			},
			wi.LangMap{
				wi.LangEn: "Logs the internal state (commands, view factories, windows), this is only relevant if -verbose is used.",
			},
		},
		&wi.CommandImpl{
			"show_command_window",
			0,
			cmdShowCommandWindow,
			wi.CommandsCategory,
			wi.LangMap{
				wi.LangEn: "Shows the interactive command window",
			},
			wi.LangMap{
				wi.LangEn: "This commands exists so it can be bound to a key to pop up the interactive command window.",
			},
		},

		&wi.CommandAlias{"alias", "command_alias"},
		&wi.CommandAlias{"new", "document_new"},
		&wi.CommandAlias{"open", "document_open"},
		&wi.CommandAlias{"q", "editor_quit"},
		&wi.CommandAlias{"quit", "editor_quit"},

		// DIRECTION = up/down/left/right
		// window_DIRECTION
		// cursor_move_DIRECTION
		// add_text/insert/delete
		// undo/redo
		// verb/movement/multiplier
		// Modes, select (both column and normal), command.
		// ...
	}
	for _, cmd := range defaultCommands {
		dispatcher.Register(cmd)
	}
}
