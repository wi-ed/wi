// Copyright 2014 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package main

import (
	"github.com/maruel/wi/wi-plugin"
)

// TODO(maruel): Implement simple config feature. Not everything should be a plugin?

type config struct {
	ints    map[string]int
	strings map[string]string
}

func (c *config) GetInt(name string) int {
	return c.ints[name]
}

func (c *config) GetString(name string) string {
	return c.strings[name]
}

func (c *config) Save() {
}

func MakeConfig() wi.Config {
	return &config{}
}