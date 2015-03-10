// Copyright 2013 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package editor

import (
	"github.com/wi-ed/wi/wicore"
	"github.com/wi-ed/wi/wicore/colors"
	"github.com/wi-ed/wi/wicore/key"
	"github.com/wi-ed/wi/wicore/lang"
	"github.com/wi-ed/wi/wicore/raster"
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
	cursorLine      int // cursor position is 0-based.
	cursorColumn    int
	cursorColumnMax int         // cursor position if the line was long enough.
	offsetLine      int         // Offset of the view of the document.
	offsetColumn    int         // Offset of the view of the document. Only make sense when wordWrap==false.
	wordWrap        bool        // true if word-wrapping is in effect. TODO(maruel): Implement.
	columnMode      bool        // true if free movement is in effect. TODO(maruel): Implement.
	colorMode       ColorMode   // Coloring of the file. Technically it'd be possible to have one file view without color and another with. TODO(maruel): Determine if useful.
	selection       raster.Rect // selection if any. TODO(maruel): Selection in columnMode vs normal selection vs line selection.
}

func (v *documentView) Close() error {
	err := v.view.Close()
	err2 := v.document.Close()
	if err != nil {
		return err
	}
	return err2
}

func (v *documentView) Buffer() *raster.Buffer {
	v.buffer.Fill(raster.Cell{' ', v.defaultFormat})
	v.document.RenderInto(v.buffer, v, v.offsetColumn, v.offsetLine)
	// TODO(maruel): Draw the cursor using proper terminal function.
	cell := v.buffer.Cell(v.offsetColumn+v.cursorColumn, v.offsetLine+v.cursorLine)
	cell.F.Bg = colors.White
	cell.F.Fg = colors.Black
	// TODO(maruel): Draw the selection over.
	return v.buffer
}

// cursorMoved triggers the event and ensures the cursor is visible.
func (v *documentView) cursorMoved(e wicore.Editor) {
	e.TriggerDocumentCursorMoved(v.document, v.cursorColumn, v.cursorLine)
	// TODO(maruel): Adjust v.offsetLine and v.offsetColumn as necessary.
	// TODO(maruel): Trigger redraw.
}

func (v *documentView) onKeyPress(e wicore.Editor, k key.Press) {
	// TODO(maruel): Only get when the View is active.
	l := v.document.content[v.cursorLine]
	v.document.content[v.cursorLine] = l[:v.cursorColumn] + string(k.Ch) + l[v.cursorColumn:]
	v.cursorColumn++
	v.cursorColumnMax = v.cursorColumn
	v.cursorMoved(e)
	// TODO(maruel): Implement dirty instead.
	e.TriggerTerminalResized()
}

func cmdToDoc(handler func(v *documentView, e wicore.EditorW)) wicore.CommandImplHandler {
	return func(c *wicore.CommandImpl, e wicore.EditorW, w wicore.Window, args ...string) {
		v, ok := w.View().(*documentView)
		if !ok {
			e.ExecuteCommand(w, "alert", "Internal error")
			return
		}
		handler(v, e)
	}
}

func cmdDocumentCursorLeft(v *documentView, e wicore.EditorW) {
	if v.cursorColumn == 0 {
		// TODO(maruel): Make wrap behavior optional.
		if v.cursorLine == 0 {
			// TODO(maruel): Beep.
			return
		}
		v.cursorLine--
		v.cursorColumn = len(v.document.content[v.cursorLine]) - 1
	}
	v.cursorColumnMax = v.cursorColumn
	v.cursorMoved(e)
}

func cmdDocumentCursorRight(v *documentView, e wicore.EditorW) {
	if v.cursorColumn == len(v.document.content[v.cursorLine])-1 {
		// TODO(maruel): Make wrap behavior optional.
		if v.cursorLine > len(v.document.content)-1 {
			// TODO(maruel): Beep.
			return
		}
		v.cursorLine++
		v.cursorColumn = 0
	} else {
		v.cursorColumn++
	}
	v.cursorColumnMax = v.cursorColumn
	v.cursorMoved(e)
}

func cmdDocumentCursorUp(v *documentView, e wicore.EditorW) {
	if v.cursorLine == 0 {
		// TODO(maruel): Beep.
		return
	}
	v.cursorLine--
	if v.cursorColumn >= len(v.document.content[v.cursorLine]) {
		v.cursorColumn = len(v.document.content[v.cursorLine]) - 1
	}
	v.cursorMoved(e)
}

func cmdDocumentCursorDown(v *documentView, e wicore.EditorW) {
	if v.cursorLine >= len(v.document.content)-1 {
		// TODO(maruel): Beep.
		return
	}
	v.cursorLine++
	if v.cursorColumn >= len(v.document.content[v.cursorLine]) {
		v.cursorColumn = len(v.document.content[v.cursorLine]) - 1
	}
	v.cursorMoved(e)
}

func cmdDocumentCursorHome(v *documentView, e wicore.EditorW) {
	if v.cursorLine != 0 || v.cursorColumnMax != 0 {
		v.cursorLine = 0
		v.cursorColumn = 0
		v.cursorColumnMax = v.cursorColumn
		v.cursorMoved(e)
	}
}

func cmdDocumentCursorEnd(v *documentView, e wicore.EditorW) {
	if v.cursorLine != len(v.document.content)-1 || v.cursorColumnMax != len(v.document.content[v.cursorLine])-1 {
		v.cursorLine = len(v.document.content) - 1
		v.cursorColumn = len(v.document.content[v.cursorLine]) - 1
		v.cursorColumnMax = v.cursorColumn
		v.cursorMoved(e)
	}
}

func documentViewFactory(e wicore.Editor, id int, args ...string) wicore.ViewW {
	dispatcher := makeCommands()
	cmds := []wicore.Command{
		&wicore.CommandImpl{
			"document_cursor_left",
			0,
			cmdToDoc(cmdDocumentCursorLeft),
			wicore.WindowCategory,
			lang.Map{
				lang.En: "Moves cursor left",
			},
			lang.Map{
				lang.En: "Moves cursor left.",
			},
		},
		&wicore.CommandImpl{
			"document_cursor_right",
			0,
			cmdToDoc(cmdDocumentCursorRight),
			wicore.WindowCategory,
			lang.Map{
				lang.En: "Moves cursor right",
			},
			lang.Map{
				lang.En: "Moves cursor right.",
			},
		},
		&wicore.CommandImpl{
			"document_cursor_up",
			0,
			cmdToDoc(cmdDocumentCursorUp),
			wicore.WindowCategory,
			lang.Map{
				lang.En: "Moves cursor up",
			},
			lang.Map{
				lang.En: "Moves cursor up.",
			},
		},
		&wicore.CommandImpl{
			"document_cursor_down",
			0,
			cmdToDoc(cmdDocumentCursorDown),
			wicore.WindowCategory,
			lang.Map{
				lang.En: "Moves cursor down",
			},
			lang.Map{
				lang.En: "Moves cursor down.",
			},
		},
		&wicore.CommandImpl{
			"document_cursor_home",
			0,
			cmdToDoc(cmdDocumentCursorHome),
			wicore.WindowCategory,
			lang.Map{
				lang.En: "Moves cursor to the beginning of the document",
			},
			lang.Map{
				lang.En: "Moves cursor to the beginning of the document.",
			},
		},
		&wicore.CommandImpl{
			"document_cursor_end",
			0,
			cmdToDoc(cmdDocumentCursorEnd),
			wicore.WindowCategory,
			lang.Map{
				lang.En: "Moves cursor to the end of the document",
			},
			lang.Map{
				lang.En: "Moves cursor to the end of the document.",
			},
		},
	}
	for _, cmd := range cmds {
		dispatcher.Register(cmd)
	}

	bindings := makeKeyBindings()
	bindings.Set(wicore.AllMode, key.Press{Key: key.Left}, "document_cursor_left")
	bindings.Set(wicore.AllMode, key.Press{Key: key.Right}, "document_cursor_right")
	bindings.Set(wicore.AllMode, key.Press{Key: key.Up}, "document_cursor_up")
	bindings.Set(wicore.AllMode, key.Press{Key: key.Down}, "document_cursor_down")
	bindings.Set(wicore.AllMode, key.Press{Key: key.Home}, "document_cursor_home")
	bindings.Set(wicore.AllMode, key.Press{Key: key.End}, "document_cursor_end")
	// vim style movement.
	bindings.Set(wicore.Normal, key.Press{Ch: 'h'}, "document_cursor_left")
	bindings.Set(wicore.Normal, key.Press{Ch: 'l'}, "document_cursor_right")
	bindings.Set(wicore.Normal, key.Press{Ch: 'k'}, "document_cursor_up")
	bindings.Set(wicore.Normal, key.Press{Ch: 'j'}, "document_cursor_down")

	// TODO(maruel): Sort out "use max space".
	// TODO(maruel): Load last cursor position from config.
	v := &documentView{
		view: view{
			commands:      dispatcher,
			keyBindings:   bindings,
			id:            id,
			title:         "<Empty document>", // TODO(maruel): Title == document.filePath ?
			naturalX:      100,
			naturalY:      100,
			defaultFormat: raster.CellFormat{Fg: colors.BrightYellow, Bg: colors.Black},
		},
		document: makeDocument(),
	}
	v.onAttach = func(_ *view, w wicore.Window) {
		v.cursorMoved(e)
	}
	v.events = append(v.events, e.RegisterTerminalKeyPressed(func(k key.Press) {
		v.onKeyPress(e, k)
	}))
	return v
}
