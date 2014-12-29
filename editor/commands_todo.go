// Copyright 2014 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package editor

import (
	"log"

	"github.com/maruel/wi/wiCore"
)

func cmdDoc(c *wiCore.CommandImpl, cd wiCore.CommandDispatcherFull, w wiCore.Window, args ...string) {
	// TODO(maruel): Grab the current word under selection if no args is
	// provided. Pass this token to shell.
	docArgs := make([]string, len(args)+1)
	docArgs[0] = "doc"
	copy(docArgs[1:], args)
	//dispatcher.Execute(w, "shell", docArgs...)
}

func cmdHelp(c *wiCore.CommandImpl, cd wiCore.CommandDispatcherFull, w wiCore.Window, args ...string) {
	// TODO(maruel): Creates a new Window with a ViewHelp.
	log.Printf("Faking help: %s", args)
}

func cmdShell(c *wiCore.CommandImpl, cd wiCore.CommandDispatcherFull, w wiCore.Window, args ...string) {
	log.Printf("Faking opening a new shell: %s", args)
}

// RegisterTodoCommands registers the top-level native commands that are yet to
// be implemented.
//
// TODO(maruel): Implement these commands properly and move to the right place.
func RegisterTodoCommands(dispatcher wiCore.Commands) {
	cmds := []wiCore.Command{
		&wiCore.CommandImpl{
			"doc",
			-1,
			cmdDoc,
			wiCore.WindowCategory,
			wiCore.LangMap{
				wiCore.LangEn: "Search godoc documentation",
			},
			wiCore.LangMap{
				wiCore.LangEn: "Uses the 'doc' tool to get documentation about the text under the cursor.",
			},
		},
		&wiCore.CommandImpl{
			"help",
			-1,
			cmdHelp,
			wiCore.WindowCategory,
			wiCore.LangMap{
				wiCore.LangEn: "Prints help",
			},
			wiCore.LangMap{
				wiCore.LangEn: "Prints general help or help for a particular command.",
			},
		},
		&wiCore.CommandImpl{
			"shell",
			-1,
			cmdShell,
			wiCore.WindowCategory,
			wiCore.LangMap{
				wiCore.LangEn: "Opens a shell process",
			},
			wiCore.LangMap{
				wiCore.LangEn: "Opens a shell process in a new buffer.",
			},
		},
	}
	for _, cmd := range cmds {
		dispatcher.Register(cmd)
	}
}
