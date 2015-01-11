// Copyright 2014 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package editor

import (
	"log"

	"github.com/maruel/wi/wicore"
	"github.com/maruel/wi/wicore/lang"
)

func cmdDoc(c *wicore.CommandImpl, e wicore.EditorW, w wicore.Window, args ...string) {
	// TODO(maruel): Grab the current word under selection if no args is
	// provided. Pass this token to shell.
	docArgs := make([]string, len(args)+1)
	docArgs[0] = "doc"
	copy(docArgs[1:], args)
	//dispatcher.Execute(w, "shell", docArgs...)
}

func cmdHelp(c *wicore.CommandImpl, e wicore.EditorW, w wicore.Window, args ...string) {
	// TODO(maruel): Creates a new Window with a ViewHelp.
	log.Printf("Faking help: %s", args)
}

func cmdShell(c *wicore.CommandImpl, e wicore.EditorW, w wicore.Window, args ...string) {
	log.Printf("Faking opening a new shell: %s", args)
}

// RegisterTodoCommands registers the top-level native commands that are yet to
// be implemented.
//
// TODO(maruel): Implement these commands properly and move to the right place.
func RegisterTodoCommands(dispatcher wicore.CommandsW) {
	cmds := []wicore.Command{
		&wicore.CommandImpl{
			"doc",
			-1,
			cmdDoc,
			wicore.WindowCategory,
			lang.Map{
				lang.En: "Search godoc documentation",
			},
			lang.Map{
				lang.En: "Uses the 'doc' tool to get documentation about the text under the cursor.",
			},
		},
		&wicore.CommandImpl{
			"help",
			-1,
			cmdHelp,
			wicore.WindowCategory,
			lang.Map{
				lang.En: "Prints help",
			},
			lang.Map{
				lang.En: "Prints general help or help for a particular command.",
			},
		},
		&wicore.CommandImpl{
			"shell",
			-1,
			cmdShell,
			wicore.WindowCategory,
			lang.Map{
				lang.En: "Opens a shell process",
			},
			lang.Map{
				lang.En: "Opens a shell process in a new buffer.",
			},
		},
	}
	for _, cmd := range cmds {
		dispatcher.Register(cmd)
	}
}
