// Copyright 2014 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package wi

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
