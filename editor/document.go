// Copyright 2014 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package editor

import (
	"io"
	"log"

	"github.com/maruel/wi/wiCore"
)

// ReadWriteSeekCloser is a generic handle to a file.
// TODO(maruel): No idea why package io doesn't provide this interface.
type ReadWriteSeekCloser interface {
	io.Closer
	io.ReadWriteSeeker
}

// document is a live editable docuent.
// TODO(maruel): This will probably have to be moved into wiCore, since
// documents could be useful to plugins (?)
// TODO(maruel): editor needs to have the list of opened document. They may have multiple views associated to a single document.
type document struct {
	filePath string              // filePath encoded in unicode. This can cause problems with systems not using an unicode code page.
	handle   ReadWriteSeekCloser // Handle to the file. For unsaved files, it's empty.
}

// RenderInto renders a view of a document.
func (d *document) RenderInto(buffer *wiCore.Buffer, offsetLine, offsetColumn int) {
	// TODO(maruel): Implement
}

func cmdDocumentNew(c *wiCore.CommandImpl, cd wiCore.CommandDispatcherFull, w wiCore.Window, args ...string) {
	cmd := make([]string, 3+len(args))
	//cmd[0] = w.ID()
	cmd[0] = wiCore.RootWindow(w).ID()
	cmd[1] = "fill"
	cmd[2] = "new_document"
	copy(cmd[3:], args)
	cd.ExecuteCommand(w, "window_new", cmd...)
}

func cmdDocumentOpen(c *wiCore.CommandImpl, cd wiCore.CommandDispatcherFull, w wiCore.Window, args ...string) {
	// The Window and View are created synchronously. The View is populated
	// asynchronously.
	log.Printf("Faking opening a file: %s", args)
}

// RegisterDocumentCommands registers the top-level native commands to manage
// documents.
func RegisterDocumentCommands(dispatcher wiCore.Commands) {
	cmds := []wiCore.Command{
		&wiCore.CommandImpl{
			"document_new",
			0,
			cmdDocumentNew,
			wiCore.WindowCategory,
			wiCore.LangMap{
				wiCore.LangEn: "Create a new buffer",
			},
			wiCore.LangMap{
				wiCore.LangEn: "Create a new buffer.",
			},
		},
		&wiCore.CommandImpl{
			"document_open",
			1,
			cmdDocumentOpen,
			wiCore.WindowCategory,
			wiCore.LangMap{
				wiCore.LangEn: "Opens a file in a new buffer",
			},
			wiCore.LangMap{
				wiCore.LangEn: "Opens a file in a new buffer.",
			},
		},

		&wiCore.CommandAlias{"new", "document_new", nil},
		&wiCore.CommandAlias{"o", "document_open", nil},
		&wiCore.CommandAlias{"open", "document_open", nil},
	}
	for _, cmd := range cmds {
		dispatcher.Register(cmd)
	}
}
