// Copyright 2014 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package editor

import (
	"github.com/maruel/wi/pkg/lang"
	"github.com/maruel/wi/wicore"
)

// commands is the map of registered commands.
type commands struct {
	commands map[string]wicore.Command
	names    []string
}

func (c *commands) Register(cmd wicore.Command) bool {
	name := cmd.Name()
	_, ok := c.commands[name]
	c.commands[name] = cmd
	c.names = nil
	return !ok
}

func (c *commands) Get(cmd string) wicore.Command {
	return c.commands[cmd]
}

func (c *commands) GetNames() []string {
	if c.names == nil {
		c.names = make([]string, 0, len(c.commands))
		for name := range c.commands {
			c.names = append(c.names, name)
		}
	}
	return c.names
}

func makeCommands() wicore.Commands {
	return &commands{make(map[string]wicore.Command), nil}
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
	CategoryValue  wicore.CommandCategory
	ShortDescValue lang.Map
	LongDescValue  lang.Map
}

func (c *privilegedCommandImpl) Name() string {
	return c.NameValue
}

func (c *privilegedCommandImpl) Handle(e wicore.Editor, w wicore.Window, args ...string) {
	if c.ExpectedArgs != -1 && len(args) != c.ExpectedArgs {
		e.ExecuteCommand(w, "alert", c.LongDesc())
	}
	// Convert types to internal types.
	ed := e.(*editor)
	wInternal := w.(*window)
	c.HandlerValue(c, ed, wInternal, args...)
}

func (c *privilegedCommandImpl) Category(e wicore.Editor, w wicore.Window) wicore.CommandCategory {
	return c.CategoryValue
}

func (c *privilegedCommandImpl) ShortDesc() string {
	return c.ShortDescValue.String()
}

func (c *privilegedCommandImpl) LongDesc() string {
	return c.LongDescValue.String()
}

// Commands

func cmdCommandAlias(c *wicore.CommandImpl, e wicore.Editor, w wicore.Window, args ...string) {
	if args[0] == "window" {
	} else if args[0] == "global" {
		w = wicore.RootWindow(w)
	} else {
		cmd := wicore.GetCommand(e, w, "command_alias")
		e.ExecuteCommand(w, "alert", cmd.LongDesc())
		return
	}
	alias := &wicore.CommandAlias{args[1], args[2], nil}
	w.View().Commands().Register(alias)
}

// RegisterCommandCommands registers the top-level native commands.
func RegisterCommandCommands(dispatcher wicore.Commands) {
	cmds := []wicore.Command{
		&wicore.CommandImpl{
			"command_alias",
			3,
			cmdCommandAlias,
			wicore.CommandsCategory,
			lang.Map{
				lang.En: "Binds an alias to another command",
			},
			lang.Map{
				// TODO(maruel): For complex aliasing, use macro?
				lang.En: "Usage: command_alias [window|global] <alias> <name>\nBinds an alias to another command. The alias can either be local to the window or global",
			},
		},

		&wicore.CommandAlias{"alias", "command_alias", nil},
	}
	for _, cmd := range cmds {
		dispatcher.Register(cmd)
	}
}
