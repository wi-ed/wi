// Copyright 2014 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package editor

import (
	"fmt"
	"io"
	"log"
	"strings"
	"unicode"

	"github.com/maruel/wi/wicore"
	"github.com/maruel/wi/wicore/lang"
	"github.com/maruel/wi/wicore/raster"
)

// ReadWriteSeekCloser is a generic handle to a file.
//
// TODO(maruel): No idea why package io doesn't provide this interface.
type ReadWriteSeekCloser interface {
	io.Closer
	io.ReadWriteSeeker
}

// document is a live editable document.
//
// TODO(maruel): This will probably have to be moved into wicore, since
// documents could be useful to plugins (?)
//
// TODO(maruel): editor needs to have the list of opened document. They may
// have multiple views associated to a single document.
//
// TODO(maruel): Strictly speaking, a Window could be in the wi parent process,
// a View in a plugin process and a Document in a separate plugin process (e.g.
// output from a live command, whatever). This means wicore.Document would need
// to be a proper interface.
type document struct {
	filePath string              // filePath encoded in unicode. This can cause problems with systems not using an unicode code page.
	fileType string              // One of the known file type. Generally described by a file extension, optionally followed by a version (?). TODO(maruel): Design.
	handle   ReadWriteSeekCloser // Handle to the file. For unsaved files, it's empty.
	content  []string            // Content as a slice of string, each being a line. In practice, it could be desired that a document not to be fully loaded in memory, or loaded asynchronously. TODO(maruel): Implement partial loading.
	isDirty  bool                // true if the content was not saved to disk.
}

func makeDocument() *document {
	return &document{
		// TODO(maruel): Obviously, no initial content.
		content: []string{"Dummy content\n", "Really\n"},
	}
}

func (d *document) ID() string {
	// TODO(maruel): This implies the same document should never be loaded twice.
	// It think it's a valid assumption, multiple DocumentView should be created
	// instead.
	// TODO(maruel): Implement uniqueness. And look at hardlinks and symlinks too
	// to dedupe paths.
	return fmt.Sprintf("document:%s", d.filePath)
}

func (d *document) String() string {
	return fmt.Sprintf("Document(%s)", d.filePath)
}

func (d *document) Close() error {
	return nil
}

func (d *document) RenderInto(buffer *raster.Buffer, view wicore.View, offsetColumn, offsetLine int) {
	for row, l := range d.content {
		// This will automatically elide text.
		if offsetColumn != 0 {
			// TODO(maruel): This is a hot path and should be optimized accordingly
			// by not requiring converting the full string.
			// TODO(maruel): Handle zero width space U+200B. It should (obviously)
			// not take any space.
			l = string([]rune(l)[offsetColumn:])
		}
		// It is particularly important on Windows, as "\n" would be rendered as an invalid character.
		l := strings.TrimRightFunc(l, unicode.IsSpace)
		buffer.DrawString(l, offsetColumn, row+offsetLine, view.DefaultFormat())
	}
}

func (d *document) FileType() wicore.FileType {
	return wicore.Scanning
}

func (d *document) IsDirty() bool {
	return d.isDirty
}

// Commands.

func cmdDocumentBuild(c *wicore.CommandImpl, e wicore.EditorW, w wicore.Window, args ...string) {
	e.ExecuteCommand(w, "alert", "Implement 'document_build' for your document")
}

func cmdDocumentNew(c *wicore.CommandImpl, e wicore.EditorW, w wicore.Window, args ...string) {
	cmd := make([]string, 3+len(args))
	//cmd[0] = w.ID()
	cmd[0] = wicore.RootWindow(w).ID()
	cmd[1] = "fill"
	cmd[2] = "new_document"
	copy(cmd[3:], args)
	e.ExecuteCommand(w, "window_new", cmd...)
}

func cmdDocumentOpen(c *wicore.CommandImpl, e wicore.EditorW, w wicore.Window, args ...string) {
	// The Window and View are created synchronously. The View is populated
	// asynchronously.
	log.Printf("Faking opening a file: %s", args)
}

func cmdDocumentRun(c *wicore.CommandImpl, e wicore.EditorW, w wicore.Window, args ...string) {
	e.ExecuteCommand(w, "alert", "Implement 'document_run' for your document")
}

// RegisterDocumentCommands registers the top-level native commands to manage
// documents.
func RegisterDocumentCommands(dispatcher wicore.CommandsW) {
	cmds := []wicore.Command{
		&wicore.CommandImpl{
			"document_build",
			0,
			cmdDocumentBuild,
			wicore.WindowCategory,
			lang.Map{
				lang.En: "Build a file",
			},
			lang.Map{
				lang.En: "Build a file.",
			},
		},
		&wicore.CommandImpl{
			"document_new",
			0,
			cmdDocumentNew,
			wicore.WindowCategory,
			lang.Map{
				lang.En: "Create a new buffer",
			},
			lang.Map{
				// TODO(maruel): Add a command to create a new buffer without a new
				// window. Wrapper command does the connection to create doc, open new
				// window, then load buffer into window.
				lang.En: "Create a new buffer. It also creates a new window to hold the document.",
			},
		},
		&wicore.CommandImpl{
			"document_open",
			1,
			cmdDocumentOpen,
			wicore.WindowCategory,
			lang.Map{
				lang.En: "Opens a file in a new buffer",
			},
			lang.Map{
				lang.En: "Opens a file in a new buffer.",
			},
		},
		&wicore.CommandImpl{
			"document_run",
			0,
			cmdDocumentRun,
			wicore.WindowCategory,
			lang.Map{
				lang.En: "Run a file",
			},
			lang.Map{
				lang.En: "Run a file.",
			},
		},

		&wicore.CommandAlias{"new", "document_new", nil},
		&wicore.CommandAlias{"o", "document_open", nil},
		&wicore.CommandAlias{"open", "document_open", nil},
	}
	for _, cmd := range cmds {
		dispatcher.Register(cmd)
	}
}
