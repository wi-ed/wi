// Copyright 2013 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package editor

import "github.com/maruel/wi/wiCore"

// ColorMode is the coloring mode in effect.
//
// TODO(maruel): Define coloring modes. Could be:
//   - A file type. Likely defined by a string, not a int.
//   - A diff view mode.
//   - No color at all.
type ColorMode int

// documentView is the View of a Document. There can be multiple views of the
// same document, each with their own cursor position.
// TODO(maruel): In some cases, the cursor position could be shared. A good
// example is vimdiff in 4-way mode.
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
	selection       wiCore.Rect // selection if any. TODO(maruel): Selection in columnMode vs normal selection vs line selection.
}

func (v *documentView) Buffer() *wiCore.Buffer {
	v.buffer.Fill(wiCore.Cell{' ', v.defaultFormat})
	v.document.RenderInto(v.buffer, v, v.offsetLine, v.offsetColumn)
	// TODO(maruel): Draw the cursor over.
	// TODO(maruel): Draw the selection over.
	return v.buffer
}

func documentViewFactory(args ...string) wiCore.View {
	// TODO(maruel): Sort out "use max space".
	return &documentView{
		view: view{
			commands:      makeCommands(),
			keyBindings:   makeKeyBindings(),
			title:         "<Empty document>", // TODO(maruel): Title == document.filePath ?
			naturalX:      100,
			naturalY:      100,
			defaultFormat: wiCore.CellFormat{Fg: wiCore.BrightYellow, Bg: wiCore.Black},
		},
		document: makeDocument(),
	}
}
