// Copyright 2014 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

//go:generate stringer -output=command_string.go -type=CommandCategory

package wicore

import (
	"fmt"
	"strings"

	"github.com/maruel/wi/pkg/lang"
)

// CommandCategory is used to put commands into sections for help purposes.
type CommandCategory int

const (
	// UnknownCategory means the command couldn't be categorized.
	UnknownCategory CommandCategory = iota
	// WindowCategory are commands relating to manipuling windows and UI in
	// general.
	WindowCategory
	// CommandsCategory are commands relating to manipulating commands, aliases,
	// keybindings.
	CommandsCategory
	// EditorCategory are commands relating to the editor lifetime.
	EditorCategory
	// DebugCategory are commands relating to debugging the app itself or plugins.
	DebugCategory

	// TODO(maruel): Add other categories.
)

// CommandHandler executes the command cmd on the Window w.
type CommandHandler func(e Editor, w Window, args ...string)

// Command describes a registered command that can be triggered directly at the
// command prompt, via a keybinding or a plugin.
type Command interface {
	// Name is the name of the command.
	Name() string
	// Handle executes the command.
	Handle(e Editor, w Window, args ...string)
	// Category returns the category the command should be bucketed in, for help
	// documentation purpose.
	Category(e Editor, w Window) CommandCategory
	// ShortDesc returns a short description of the command in the language
	// requested. It defaults to English if the description was not translated in
	// this language.
	ShortDesc(e Editor, w Window) string
	// LongDesc returns a long explanation of the command in the language
	// requested. It defaults to English if the description was not translated in
	// this language.
	LongDesc(e Editor, w Window) string
}

// Commands stores the known commands. This is where plugins can add new
// commands. Each View contains its own Commands.
type Commands interface {
	// Register registers a command so it can be executed later. In practice
	// commands should normally be registered on startup. Returns false if a
	// command was already registered and was lost.
	Register(cmd Command) bool

	// Get returns a command if registered, nil otherwise.
	Get(cmdName string) Command
}

// EnqueuedCommand is used internally to dispatch commands through
// EventRegistry.
type EnqueuedCommands struct {
	Commands [][]string
	Callback func()
}

// CommandImplHandler is the CommandHandler to use when coupled with CommandImpl.
type CommandImplHandler func(c *CommandImpl, e Editor, w Window, args ...string)

// CommandImpl is the boilerplate Command implementation.
type CommandImpl struct {
	NameValue      string
	ExpectedArgs   int // If >= 0, the command will be aborted if the number of arguments is not exactly this value. Set to -1 to disable verification. On abort, an alert with the long description of the command is done.
	HandlerValue   CommandImplHandler
	CategoryValue  CommandCategory
	ShortDescValue lang.Map
	LongDescValue  lang.Map
}

// Name implements Command.
func (c *CommandImpl) Name() string {
	return c.NameValue
}

// Handle implements Command.
func (c *CommandImpl) Handle(e Editor, w Window, args ...string) {
	if c.ExpectedArgs != -1 && len(args) != c.ExpectedArgs {
		e.ExecuteCommand(w, "alert", c.LongDesc(e, w))
	}
	c.HandlerValue(c, e, w, args...)
}

// Category implements Command.
func (c *CommandImpl) Category(e Editor, w Window) CommandCategory {
	return c.CategoryValue
}

// ShortDesc implements Command.
func (c *CommandImpl) ShortDesc(e Editor, w Window) string {
	return c.ShortDescValue.Get(e.CurrentLanguage())
}

// LongDesc implements Command.
func (c *CommandImpl) LongDesc(e Editor, w Window) string {
	return c.LongDescValue.Get(e.CurrentLanguage())
}

// CommandAlias references another command by its name. It's important to not
// bind directly to the Command reference, so that if a command is replaced by
// a plugin, that the replacement command is properly called by the alias.
type CommandAlias struct {
	NameValue    string
	CommandValue string
	ArgsValue    []string
}

// Name implements Command.
func (c *CommandAlias) Name() string {
	return c.NameValue
}

// Handle implements Command.
func (c *CommandAlias) Handle(e Editor, w Window, args ...string) {
	// The alias is executed inline. This is important for command queue
	// ordering.
	cmd := GetCommand(e, w, c.CommandValue)
	if cmd != nil {
		cmd.Handle(e, w, args...)
	} else {
		// TODO(maruel): This makes assumption on "alert".
		cmd = GetCommand(e, w, "alert")
		txt := fmt.Sprintf(AliasNotFound.Get(e.CurrentLanguage()), c.NameValue, c.CommandValue)
		cmd.Handle(e, w, txt)
	}
}

// Category implements Command.
func (c *CommandAlias) Category(e Editor, w Window) CommandCategory {
	cmd := GetCommand(e, w, c.CommandValue)
	if cmd != nil {
		return c.Category(e, w)
	}
	return UnknownCategory
}

// ShortDesc implements Command.
func (c *CommandAlias) ShortDesc(e Editor, w Window) string {
	return fmt.Sprintf(AliasFor.Get(e.CurrentLanguage()), c.merged())
}

// LongDesc implements Command.
func (c *CommandAlias) LongDesc(e Editor, w Window) string {
	return fmt.Sprintf(AliasFor.Get(e.CurrentLanguage()), c.merged())
}

func (c *CommandAlias) merged() string {
	out := c.CommandValue
	if len(c.ArgsValue) != 0 {
		out += " " + strings.Join(c.ArgsValue, " ")
	}
	return out
}

// Utility functions.

// PostCommand appends a Command at the end of the queue. It is a shortcut to
// e.TriggerCommands(EnqueuedCommands{...}).
func PostCommand(e EventRegistry, callback func(), cmdName string, args ...string) {
	line := make([]string, len(args)+1)
	line[0] = cmdName
	copy(line[1:], args)
	e.TriggerCommands(EnqueuedCommands{[][]string{line}, callback})
}

// GetCommand traverses the Window hierarchy tree to find a View that has
// the command cmd in its Commands mapping. If Window is nil, it starts with
// the Editor's active Window.
func GetCommand(e Editor, w Window, cmdName string) Command {
	if w == nil {
		w = e.ActiveWindow()
	}
	for {
		cmd := w.View().Commands().Get(cmdName)
		if cmd != nil {
			return cmd
		}
		w = w.Parent()
		if w == nil {
			return nil
		}
	}
}
