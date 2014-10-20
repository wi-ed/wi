// Copyright 2013 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package editor

import (
	"log"
	"time"
	"unicode/utf8"

	"github.com/maruel/wi/wiCore"
)

// TODO(maruel): Likely move into wiCore for reuse.
type view struct {
	commands      wiCore.Commands
	keyBindings   wiCore.KeyBindings
	title         string
	isDisabled    bool
	naturalX      int
	naturalY      int
	actualX       int
	actualY       int
	window        wiCore.Window
	onAttach      func(v *view, w wiCore.Window)
	defaultFormat wiCore.CellFormat
	buffer        *wiCore.Buffer
}

// wiCore.View interface.

func (v *view) Commands() wiCore.Commands {
	return v.commands
}

func (v *view) KeyBindings() wiCore.KeyBindings {
	return v.keyBindings
}

func (v *view) Title() string {
	return v.title
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
	v.buffer = wiCore.NewBuffer(x, y)
}

func (v *view) OnAttach(w wiCore.Window) {
	if v.onAttach != nil {
		v.onAttach(v, w)
	}
	v.window = w
}

// DefaultFormat returns the View's format or the parent Window's View's format.
func (v *view) DefaultFormat() wiCore.CellFormat {
	if v.defaultFormat.Empty() && v.window != nil {
		w := v.window.Parent()
		if w != nil {
			return w.View().DefaultFormat()
		}
	}
	return v.defaultFormat
}

// A disabled static view.
type staticDisabledView struct {
	view
}

func (v *staticDisabledView) Buffer() *wiCore.Buffer {
	// TODO(maruel): Use the parent view format by default. No idea how to
	// surface this information here. Cost is at least a RPC, potentially
	// multiple when multiple plugins are involved in the tree.
	v.buffer.Fill(wiCore.Cell{' ', v.DefaultFormat()})
	v.buffer.DrawString(v.Title(), 0, 0, v.DefaultFormat())
	return v.buffer
}

// Empty non-editable window.
func makeStaticDisabledView(title string, naturalX, naturalY int) *staticDisabledView {
	return &staticDisabledView{
		view{
			commands:      makeCommands(),
			keyBindings:   makeKeyBindings(),
			title:         title,
			isDisabled:    true,
			naturalX:      naturalX,
			naturalY:      naturalY,
			defaultFormat: wiCore.CellFormat{Fg: wiCore.Red, Bg: wiCore.Black},
		},
	}
}

// The status line is a hierarchy of Window, one for each element, each showing
// a single item.
func statusRootViewFactory(args ...string) wiCore.View {
	// TODO(maruel): OnResize(), query the root Window size, if y<=5 or x<=15,
	// set the root status Window to y=0, so that it becomes effectively
	// invisible when the editor window is too small.
	v := makeStaticDisabledView("Status Root", 1, 1)
	v.defaultFormat.Bg = wiCore.LightGray
	v.onAttach = func(v *view, w wiCore.Window) {
		id := w.ID()
		w.PostCommands(
			[][]string{
				{"window_new", id, "left", "status_name"},
				{"window_new", id, "right", "status_position"},
			})
	}
	return v
}

func statusNameViewFactory(args ...string) wiCore.View {
	// View name.
	// TODO(maruel): Register events of Window activation, make itself Invalidate().
	v := makeStaticDisabledView("Status Name", 15, 1)
	v.defaultFormat = wiCore.CellFormat{}
	return v
}

func statusPositionViewFactory(args ...string) wiCore.View {
	// Position, % of file.
	// TODO(maruel): Register events of movement, make itself Invalidate().
	v := makeStaticDisabledView("Status Position", 15, 1)
	v.defaultFormat = wiCore.CellFormat{}
	return v
}

func infobarAlertViewFactory(args ...string) wiCore.View {
	out := "Alert: " + args[0]
	l := utf8.RuneCountInString(out)
	v := makeStaticDisabledView(out, l, 1)
	v.onAttach = func(v *view, w wiCore.Window) {
		go func() {
			// Dismiss after 5 seconds.
			<-time.After(5 * time.Second)
			wiCore.PostCommand(w, "window_close", w.ID())
		}()
	}
	return v
}

// RegisterDefaultViewFactories registers the builtins views factories.
func RegisterDefaultViewFactories(e Editor) {
	e.RegisterViewFactory("command", commandViewFactory)
	e.RegisterViewFactory("infobar_alert", infobarAlertViewFactory)
	e.RegisterViewFactory("new_document", documentViewFactory)
	e.RegisterViewFactory("status_name", statusNameViewFactory)
	e.RegisterViewFactory("status_position", statusPositionViewFactory)
	e.RegisterViewFactory("status_root", statusRootViewFactory)
}

// Commands

// RegisterViewCommands registers view-related commands
func RegisterViewCommands(dispatcher wiCore.Commands) {
	cmds := []wiCore.Command{}
	for _, cmd := range cmds {
		dispatcher.Register(cmd)
	}
}
