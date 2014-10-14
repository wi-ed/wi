// Copyright 2014 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package editor

import "github.com/maruel/wi/wi_core"

// commands is the map of registered commands.
type commands struct {
	commands map[string]wi_core.Command
}

func (c *commands) Register(cmd wi_core.Command) bool {
	name := cmd.Name()
	_, ok := c.commands[name]
	c.commands[name] = cmd
	return !ok
}

func (c *commands) Get(cmd string) wi_core.Command {
	return c.commands[cmd]
}

func makeCommands() wi_core.Commands {
	return &commands{make(map[string]wi_core.Command)}
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
	CategoryValue  wi_core.CommandCategory
	ShortDescValue wi_core.LangMap
	LongDescValue  wi_core.LangMap
}

func (c *privilegedCommandImpl) Name() string {
	return c.NameValue
}

func (c *privilegedCommandImpl) Handle(cd wi_core.CommandDispatcherFull, w wi_core.Window, args ...string) {
	if c.ExpectedArgs != -1 && len(args) != c.ExpectedArgs {
		cd.ExecuteCommand(w, "alert", c.LongDesc(cd, w))
	}
	// Convert types to internal types.
	e := cd.(*editor)
	wInternal := w.(*window)
	c.HandlerValue(c, e, wInternal, args...)
}

func (c *privilegedCommandImpl) Category(cd wi_core.CommandDispatcherFull, w wi_core.Window) wi_core.CommandCategory {
	return c.CategoryValue
}

func (c *privilegedCommandImpl) ShortDesc(cd wi_core.CommandDispatcherFull, w wi_core.Window) string {
	return wi_core.GetStr(cd.CurrentLanguage(), c.ShortDescValue)
}

func (c *privilegedCommandImpl) LongDesc(cd wi_core.CommandDispatcherFull, w wi_core.Window) string {
	return wi_core.GetStr(cd.CurrentLanguage(), c.LongDescValue)
}

// Commands

func cmdCommandAlias(c *wi_core.CommandImpl, cd wi_core.CommandDispatcherFull, w wi_core.Window, args ...string) {
	if args[0] == "window" {
	} else if args[0] == "global" {
		w = wi_core.RootWindow(w)
	} else {
		cmd := wi_core.GetCommand(cd, w, "command_alias")
		cd.ExecuteCommand(w, "alert", cmd.LongDesc(cd, w))
		return
	}
	alias := &wi_core.CommandAlias{args[1], args[2], nil}
	w.View().Commands().Register(alias)
}

// RegisterCommandCommands registers the top-level native commands.
func RegisterCommandCommands(dispatcher wi_core.Commands) {
	cmds := []wi_core.Command{
		&wi_core.CommandImpl{
			"command_alias",
			3,
			cmdCommandAlias,
			wi_core.CommandsCategory,
			wi_core.LangMap{
				wi_core.LangEn: "Binds an alias to another command",
			},
			wi_core.LangMap{
				// TODO(maruel): For complex aliasing, use macro?
				wi_core.LangEn: "Usage: command_alias [window|global] <alias> <name>\nBinds an alias to another command. The alias can either be local to the window or global",
			},
		},

		&wi_core.CommandAlias{"alias", "command_alias", nil},
	}
	for _, cmd := range cmds {
		dispatcher.Register(cmd)
	}
}
