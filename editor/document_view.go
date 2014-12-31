// Copyright 2013 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package editor

import (
	"github.com/maruel/wi/pkg/key"
	"github.com/maruel/wi/wicore"
)

// ColorMode is the coloring mode in effect.
//
// TODO(maruel): Define coloring modes. Could be:
//   - A file type. Likely defined by a string, not a int.
//   - A diff view mode.
//   - No color at all.
type ColorMode int

// documentView is the View of a Document. There can be multiple views of the
// same document, each with their own cursor position.
//
// TODO(maruel): In some cases, the cursor position could be shared. A good
// example is vimdiff in 4-way mode.
//
// TODO(maruel): The part that is serializable has to be in its own structure
// for easier deserialization.
type documentView struct {
	view
	document        *document
	cursorLine      int
	cursorColumn    int
	cursorColumnMax int         // cursor position if the line was long enough.
	offsetLine      int         // Offset of the view of the document.
	offsetColumn    int         // Offset of the view of the document. Only make sense when wordWrap==false.
	wordWrap        bool        // true if word-wrapping is in effect. TODO(maruel): Implement.
	columnMode      bool        // true if free movement is in effect. TODO(maruel): Implement.
	colorMode       ColorMode   // Coloring of the file. Technically it'd be possible to have one file view without color and another with. TODO(maruel): Determine if useful.
	selection       wicore.Rect // selection if any. TODO(maruel): Selection in columnMode vs normal selection vs line selection.
}

func (v *documentView) Buffer() *wicore.Buffer {
	v.buffer.Fill(wicore.Cell{' ', v.defaultFormat})
	v.document.RenderInto(v.buffer, v, v.offsetLine, v.offsetColumn)
	// TODO(maruel): Draw the cursor over.
	// TODO(maruel): Draw the selection over.
	return v.buffer
}

func cmdDocumentCursorLeft(c *wicore.CommandImpl, cd wicore.CommandDispatcherFull, w wicore.Window, args ...string) {
	d, ok := w.View().(*documentView)
	if !ok {
		panic("Oops")
	}
	if d.cursorColumn--; d.cursorColumn == -1 {
		// Maybe make the wrap behavior optional.
		if d.cursorLine--; d.cursorLine == -1 {
			d.cursorLine = 0
			// TODO(maruel): Beep.
		}
	}
	// TODO(maruel): Err, need to implement: cd.onDocumentCursorMoved(d)
}

func documentViewFactory(e wicore.EventRegistry, args ...string) wicore.View {
	dispatcher := makeCommands()
	cmds := []wicore.Command{
		&wicore.CommandImpl{
			"document_cursor_left",
			-1,
			cmdDocumentCursorLeft,
			wicore.WindowCategory,
			wicore.LangMap{
				wicore.LangEn: "Moves cursor left",
			},
			wicore.LangMap{
				wicore.LangEn: "Moves cursor left.",
			},
		},
	}
	for _, cmd := range cmds {
		dispatcher.Register(cmd)
	}

	bindings := makeKeyBindings()
	bindings.Set(wicore.AllMode, key.Press{Key: key.Left}, "document_cursor_left")

	// TODO(maruel): Sort out "use max space".
	return &documentView{
		view: view{
			commands:      dispatcher,
			keyBindings:   bindings,
			title:         "<Empty document>", // TODO(maruel): Title == document.filePath ?
			naturalX:      100,
			naturalY:      100,
			defaultFormat: wicore.CellFormat{Fg: wicore.BrightYellow, Bg: wicore.Black},
		},
		document: makeDocument(),
	}
}
