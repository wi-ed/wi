// Copyright 2014 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package editor

import (
	"log"

	"github.com/maruel/wi/wi_core"
)

func cmdDocumentNew(c *wi_core.CommandImpl, cd wi_core.CommandDispatcherFull, w wi_core.Window, args ...string) {
	cmd := make([]string, 3+len(args))
	//cmd[0] = w.ID()
	cmd[0] = wi_core.RootWindow(w).ID()
	cmd[1] = "fill"
	cmd[2] = "new_document"
	copy(cmd[3:], args)
	cd.ExecuteCommand(w, "window_new", cmd...)
}

func cmdDocumentOpen(c *wi_core.CommandImpl, cd wi_core.CommandDispatcherFull, w wi_core.Window, args ...string) {
	// The Window and View are created synchronously. The View is populated
	// asynchronously.
	log.Printf("Faking opening a file: %s", args)
}

// RegisterDocumentCommands registers the top-level native commands to manage
// documents.
func RegisterDocumentCommands(dispatcher wi_core.Commands) {
	defaultCommands := []wi_core.Command{
		&wi_core.CommandImpl{
			"document_new",
			0,
			cmdDocumentNew,
			wi_core.WindowCategory,
			wi_core.LangMap{
				wi_core.LangEn: "Create a new buffer",
			},
			wi_core.LangMap{
				wi_core.LangEn: "Create a new buffer.",
			},
		},
		&wi_core.CommandImpl{
			"document_open",
			1,
			cmdDocumentOpen,
			wi_core.WindowCategory,
			wi_core.LangMap{
				wi_core.LangEn: "Opens a file in a new buffer",
			},
			wi_core.LangMap{
				wi_core.LangEn: "Opens a file in a new buffer.",
			},
		},

		&wi_core.CommandAlias{"new", "document_new", nil},
		&wi_core.CommandAlias{"o", "document_open", nil},
		&wi_core.CommandAlias{"open", "document_open", nil},
	}
	for _, cmd := range defaultCommands {
		dispatcher.Register(cmd)
	}
}
