// Copyright 2014 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package wi_core

import (
	"fmt"
	"strings"
)

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
