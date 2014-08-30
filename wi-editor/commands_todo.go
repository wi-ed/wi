// Copyright 2014 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package editor

import (
	"github.com/maruel/wi/wi-plugin"
	"log"
)

func cmdShell(c *wi.CommandImpl, cd wi.CommandDispatcherFull, w wi.Window, args ...string) {
	log.Printf("Faking opening a new shell: %s", args)
}

func cmdDoc(c *wi.CommandImpl, cd wi.CommandDispatcherFull, w wi.Window, args ...string) {
	// TODO(maruel): Grab the current word under selection if no args is
	// provided. Pass this token to shell.
	docArgs := make([]string, len(args)+1)
	docArgs[0] = "doc"
	copy(docArgs[1:], args)
	//dispatcher.Execute(w, "shell", docArgs...)
}

func cmdHelp(c *wi.CommandImpl, cd wi.CommandDispatcherFull, w wi.Window, args ...string) {
	// TODO(maruel): Creates a new Window with a ViewHelp.
	log.Printf("Faking help: %s", args)
}

var todoCommands = []wi.Command{
	&wi.CommandImpl{
		"doc",
		cmdDoc,
		wi.WindowCategory,
		wi.LangMap{
			wi.LangEn: "Search godoc documentation",
		},
		wi.LangMap{
			wi.LangEn: "Uses the 'doc' tool to get documentation about the text under the cursor.",
		},
	},
	&wi.CommandImpl{
		"help",
		cmdHelp,
		wi.WindowCategory,
		wi.LangMap{
			wi.LangEn: "Prints help",
		},
		wi.LangMap{
			wi.LangEn: "Prints general help or help for a particular command.",
		},
	},
	&wi.CommandImpl{
		"shell",
		cmdShell,
		wi.WindowCategory,
		wi.LangMap{
			wi.LangEn: "Opens a shell process",
		},
		wi.LangMap{
			wi.LangEn: "Opens a shell process in a new buffer.",
		},
	},
}

// RegisterTodoCommands registers the top-level native commands that are yet to
// be implemented.
//
// TODO(maruel): Implement these commands properly and move to
// RegisterDefaultCommands().
func RegisterTodoCommands(dispatcher wi.Commands) {
	for _, cmd := range todoCommands {
		dispatcher.Register(cmd)
	}
}
