// Copyright 2013 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package editor

import (
	"github.com/maruel/wi/wi-plugin"
	"log"
	"time"
	"unicode/utf8"
)

// TODO(maruel): Plugable drawing function.
type drawInto func(v wi.View, buffer wi.Buffer)

type view struct {
	commands    wi.Commands
	keyBindings wi.KeyBindings
	title       string
	isDirty     bool
	isDisabled  bool
	naturalX    int
	naturalY    int
	actualX     int
	actualY     int
	onAttach    func(v *view, w wi.Window)
	buffer      *wi.Buffer
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

func (v *view) IsDisabled() bool {
	return v.isDisabled
}

func (v *view) NaturalSize() (x, y int) {
	return v.naturalX, v.naturalY
}

func (v *view) SetSize(x, y int) {
	log.Printf("View(%s).SetSize(%d, %d)", v.Title(), x, y)
	v.actualX = x
	v.actualY = y
	v.buffer = wi.NewBuffer(x, y)
}

func (v *view) Buffer() *wi.Buffer {
	//log.Printf("View(%s).Buffer(%d, %d)", v.Title(), v.actualX, v.actualY)
	r, _ := utf8.DecodeRuneInString(v.Title())
	v.buffer.Fill(wi.Rect{0, 0, v.actualX, v.actualY}, wi.MakeCell(r, wi.Red, wi.Black))
	/*
		l := wi.LabelParams{
			wi.ColorBlue,
			wi.ColorBlack,
			wi.AlignLeft,
			'b',
			true,
		}
		v.buffer.DrawLabel(wi.Rect{0, 0, v.actualX, 1}, &l, []byte(v.Title()))
	*/
	return v.buffer
}

func (v *view) OnAttach(w wi.Window) {
	if v.onAttach != nil {
		v.onAttach(v, w)
	}
}

// Empty non-editable window.
func makeView(title string, naturalX, naturalY int) *view {
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
func statusRootViewFactory(args ...string) wi.View {
	// TODO(maruel): OnResize(), query the root Window size, if y<=5 or x<=15,
	// set the root status Window to y=0, so that it becomes effectively
	// invisible when the editor window is too small.
	return makeView("Status Root", 1, 1)
}

func statusNameViewFactory(args ...string) wi.View {
	// View name.
	// TODO(maruel): Register events of Window activation, make itself Invalidate().
	// TODO(maruel): Drawing code.
	return makeView("Status Name", 15, 1)
}

func statusPositionViewFactory(args ...string) wi.View {
	// Position, % of file.
	// TODO(maruel): Register events of movement, make itself Invalidate().
	// TODO(maruel): Drawing code.
	return makeView("Status Position", 15, 1)
}

// The command dialog box.
// TODO(maruel): Position it 5 lines below the cursor in the parent Window's
// View. Do this via onAttach.
func commandViewFactory(args ...string) wi.View {
	return makeView("Command", 30, 1)
}

func infobarAlertViewFactory(args ...string) wi.View {
	out := "Alert: " + args[0]
	l := utf8.RuneCountInString(out)
	v := makeView(out, l, 1)
	v.onAttach = func(v *view, w wi.Window) {
		go func() {
			// Dismiss after 5 seconds.
			<-time.After(5 * time.Second)
			wi.PostCommand(w, "window_close", w.Id())
		}()
	}
	return v
}

func RegisterDefaultViewFactories(e Editor) {
	e.RegisterViewFactory("command", commandViewFactory)
	e.RegisterViewFactory("infobar_alert", infobarAlertViewFactory)
	e.RegisterViewFactory("status_name", statusNameViewFactory)
	e.RegisterViewFactory("status_position", statusPositionViewFactory)
	e.RegisterViewFactory("status_root", statusRootViewFactory)
}
