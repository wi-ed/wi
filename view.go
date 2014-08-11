// Copyright 2013 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package main

import (
	"github.com/maruel/wi/wi-plugin"
	"github.com/nsf/termbox-go"
	"github.com/nsf/tulib"
	"log"
)

// TODO(maruel): Plugable drawing function.
type drawInto func(v wi.View, buffer tulib.Buffer)

type view struct {
	commands    wi.Commands
	keyBindings wi.KeyBindings
	title       string
	isDirty     bool
	isInvalid   bool
	isDisabled  bool
	naturalX    int
	naturalY    int
	buffer      tulib.Buffer
}

func (v *view) Commands() wi.Commands {
	return v.commands
}

func (v *view) KeyBindings() wi.KeyBindings {
	return v.keyBindings
}

func (v *view) Title() string {
	return v.title
}

func (v *view) IsDirty() bool {
	return v.isDirty
}

func (v *view) IsInvalid() bool {
	return v.isInvalid
}

func (v *view) IsDisabled() bool {
	return v.isDisabled
}

func (v *view) NaturalSize() (x, y int) {
	return v.naturalX, v.naturalY
}

func (v *view) SetSize(x, y int) {
	log.Printf("View(%s).SetSize(%d, %d)", v.Title(), x, y)
	v.naturalX = x
	v.naturalY = y
	v.isInvalid = true
	v.buffer = tulib.NewBuffer(x, y)
}

func (v *view) Buffer() *tulib.Buffer {
	log.Printf("View(%s).Buffer(%d, %d)", v.Title(), v.naturalX, v.naturalY)
	// TODO(maruel): Plugable drawing function.
	v.buffer.Set(0, 0, termbox.Cell{'A', termbox.ColorRed, termbox.ColorRed})
	v.isInvalid = false
	return &v.buffer
}

// Empty non-editable window.
func makeView(title string, naturalX, naturalY int) wi.View {
	return &view{
		commands:    makeCommands(),
		keyBindings: makeKeyBindings(),
		title:       title,
		naturalX:    naturalX,
		naturalY:    naturalY,
	}
}

// The status line is a hierarchy of Window, one for each element, each showing
// a single item.
func makeStatusViewCenter() wi.View {
	// TODO(maruel): OnResize(), query the root Window size, if y<=5 or x<=15,
	// set the root status Window to y=0, so that it becomes effectively
	// invisible when the editor window is too small.
	return makeView("Status Center", 1, -1)
}

func makeStatusViewName() wi.View {
	// View name.
	// TODO(maruel): Register events of Window activation, make itself Invalidate().
	// TODO(maruel): Drawing code.
	return makeView("Status Name", 1, -1)
}

func makeStatusViewPosition() wi.View {
	// Position, % of file.
	// TODO(maruel): Register events of movement, make itself Invalidate().
	// TODO(maruel): Drawing code.
	return makeView("Status Position", 1, -1)
}

// The command box.
func makeCommandView() wi.View {
	return makeView("Command", 1, -1)
}

// A dismissable modal dialog box. TODO(maruel): An infobar that auto-dismiss
// itself after 5s.
func makeAlertView(text string) wi.View {
	return makeView("Alert:"+text, 1, 1)
}
