// Copyright 2014 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package wi

import (
	"fmt"
)

// CommandImplHandler is the CommandHandler to use when coupled with CommandImpl.
type CommandImplHandler func(c *CommandImpl, cd CommandDispatcherFull, w Window, args ...string)

// CommandImpl is the boilerplate Command implementation.
type CommandImpl struct {
	NameValue      string
	HandlerValue   CommandImplHandler
	CategoryValue  CommandCategory
	ShortDescValue LangMap
	LongDescValue  LangMap
}

func (c *CommandImpl) Name() string {
	return c.NameValue
}

func (c *CommandImpl) Handle(cd CommandDispatcherFull, w Window, args ...string) {
	c.HandlerValue(c, cd, w, args...)
}

func (c *CommandImpl) Category(cd CommandDispatcherFull, w Window) CommandCategory {
	return c.CategoryValue
}

func (c *CommandImpl) ShortDesc(cd CommandDispatcherFull, w Window) string {
	return GetStr(cd.CurrentLanguage(), c.ShortDescValue)
}

func (c *CommandImpl) LongDesc(cd CommandDispatcherFull, w Window) string {
	return GetStr(cd.CurrentLanguage(), c.LongDescValue)
}

// CommandAlias references another command by its name. It's important to not
// bind directly to the Command reference, so that if a command is replaced by
// a plugin, that the replacement command is properly called by the alias.
type CommandAlias struct {
	NameValue    string
	CommandValue string
}

func (c *CommandAlias) Name() string {
	return c.NameValue
}

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

func (c *CommandAlias) Category(cd CommandDispatcherFull, w Window) CommandCategory {
	cmd := GetCommand(cd, w, c.CommandValue)
	if cmd != nil {
		return c.Category(cd, w)
	}
	return UnknownCategory
}

func (c *CommandAlias) ShortDesc(cd CommandDispatcherFull, w Window) string {
	return fmt.Sprintf(GetStr(cd.CurrentLanguage(), AliasFor), c.CommandValue)
}

func (c *CommandAlias) LongDesc(cd CommandDispatcherFull, w Window) string {
	return fmt.Sprintf(GetStr(cd.CurrentLanguage(), AliasFor), c.CommandValue)
}
