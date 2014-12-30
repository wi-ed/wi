// Copyright 2014 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

//go:generate stringer -output=command_string.go -type=CommandCategory

package wicore

import (
	"fmt"
	"strings"
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

// CommandID describes a command in the queue.
type CommandID struct {
	ProcessID    int
	CommandIndex int
}

func (c CommandID) String() string {
	return fmt.Sprintf("%d:%d", c.ProcessID, c.CommandIndex)
}

// CommandDispatcher owns the command queue. Use this interface to enqueue
// commands for execution.
type CommandDispatcher interface {
	// PostCommands appends several Command calls at the end of the queue. Using
	// this function guarantees that all the commands will be executed in order
	// without commands interfering.
	//
	// `callback` is called synchronously after the command is executed.
	PostCommands(cmds [][]string, callback func()) CommandID
}

// CommandDispatcherFull is a superset of CommandDispatcher for internal use.
type CommandDispatcherFull interface {
	CommandDispatcher

	// ExecuteCommand executes a command now. This is only meant to run a command
	// reentrantly; e.g. running a command triggers another one. This usually
	// happens for key binding, command aliases, when a command triggers an error.
	ExecuteCommand(w Window, cmdName string, args ...string)

	// ActiveWindow returns the current active Window.
	ActiveWindow() Window

	// RegisterViewFactory makes a nwe view available by name.
	RegisterViewFactory(name string, viewFactory ViewFactory) bool

	CurrentLanguage() LanguageMode
}

// CommandHandler executes the command cmd on the Window w.
type CommandHandler func(cd CommandDispatcherFull, w Window, args ...string)

// Command describes a registered command that can be triggered directly at the
// command prompt, via a keybinding or a plugin.
type Command interface {
	// Name is the name of the command.
	Name() string
	// Handle executes the command.
	Handle(cd CommandDispatcherFull, w Window, args ...string)
	// Category returns the category the command should be bucketed in, for help
	// documentation purpose.
	Category(cd CommandDispatcherFull, w Window) CommandCategory
	// ShortDesc returns a short description of the command in the language
	// requested. It defaults to English if the description was not translated in
	// this language.
	ShortDesc(cd CommandDispatcherFull, w Window) string
	// LongDesc returns a long explanation of the command in the language
	// requested. It defaults to English if the description was not translated in
	// this language.
	LongDesc(cd CommandDispatcherFull, w Window) string
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

// CommandImplHandler is the CommandHandler to use when coupled with CommandImpl.
type CommandImplHandler func(c *CommandImpl, cd CommandDispatcherFull, w Window, args ...string)

// CommandImpl is the boilerplate Command implementation.
type CommandImpl struct {
	NameValue      string
	ExpectedArgs   int // If >= 0, the command will be aborted if the number of arguments is not exactly this value. Set to -1 to disable verification. On abort, an alert with the long description of the command is done.
	HandlerValue   CommandImplHandler
	CategoryValue  CommandCategory
	ShortDescValue LangMap
	LongDescValue  LangMap
}

// Name implements Command.
func (c *CommandImpl) Name() string {
	return c.NameValue
}

// Handle implements Command.
func (c *CommandImpl) Handle(cd CommandDispatcherFull, w Window, args ...string) {
	if c.ExpectedArgs != -1 && len(args) != c.ExpectedArgs {
		cd.ExecuteCommand(w, "alert", c.LongDesc(cd, w))
	}
	c.HandlerValue(c, cd, w, args...)
}

// Category implements Command.
func (c *CommandImpl) Category(cd CommandDispatcherFull, w Window) CommandCategory {
	return c.CategoryValue
}

// ShortDesc implements Command.
func (c *CommandImpl) ShortDesc(cd CommandDispatcherFull, w Window) string {
	return GetStr(cd.CurrentLanguage(), c.ShortDescValue)
}

// LongDesc implements Command.
func (c *CommandImpl) LongDesc(cd CommandDispatcherFull, w Window) string {
	return GetStr(cd.CurrentLanguage(), c.LongDescValue)
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
func (c *CommandAlias) Handle(cd CommandDispatcherFull, w Window, args ...string) {
	// The alias is executed inline. This is important for command queue
	// ordering.
	cmd := GetCommand(cd, w, c.CommandValue)
	if cmd != nil {
		cmd.Handle(cd, w, args...)
	} else {
		// TODO(maruel): This makes assumption on "alert".
		cmd = GetCommand(cd, w, "alert")
		txt := fmt.Sprintf(GetStr(cd.CurrentLanguage(), AliasNotFound), c.NameValue, c.CommandValue)
		cmd.Handle(cd, w, txt)
	}
}

// Category implements Command.
func (c *CommandAlias) Category(cd CommandDispatcherFull, w Window) CommandCategory {
	cmd := GetCommand(cd, w, c.CommandValue)
	if cmd != nil {
		return c.Category(cd, w)
	}
	return UnknownCategory
}

// ShortDesc implements Command.
func (c *CommandAlias) ShortDesc(cd CommandDispatcherFull, w Window) string {
	return fmt.Sprintf(GetStr(cd.CurrentLanguage(), AliasFor), c.merged())
}

// LongDesc implements Command.
func (c *CommandAlias) LongDesc(cd CommandDispatcherFull, w Window) string {
	return fmt.Sprintf(GetStr(cd.CurrentLanguage(), AliasFor), c.merged())
}

func (c *CommandAlias) merged() string {
	out := c.CommandValue
	if len(c.ArgsValue) != 0 {
		out += " " + strings.Join(c.ArgsValue, " ")
	}
	return out
}

// Utility functions.

// PostCommand appends a Command at the end of the queue.
// It is a shortcut to cd.PostCommands([][]string{[]string{cmdName, args...}},
// callback). Sadly, using `...string` means that callback cannot be the last
// parameter.
func PostCommand(cd CommandDispatcher, callback func(), cmdName string, args ...string) CommandID {
	line := make([]string, len(args)+1)
	line[0] = cmdName
	copy(line[1:], args)
	return cd.PostCommands([][]string{line}, callback)
}

// GetCommand traverses the Window hierarchy tree to find a View that has
// the command cmd in its Commands mapping. If Window is nil, it starts with
// the Editor's active Window.
func GetCommand(cd CommandDispatcherFull, w Window, cmdName string) Command {
	if w == nil {
		w = cd.ActiveWindow()
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
