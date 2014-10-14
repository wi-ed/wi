// Copyright 2013 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package editor

import (
	"log"
	"sort"
	"time"
	"unicode/utf8"

	"github.com/maruel/wi/wi_core"
)

// TODO(maruel): Plugable drawing function.
type drawInto func(v wi_core.View, buffer wi_core.Buffer)

type view struct {
	commands    wi_core.Commands
	keyBindings wi_core.KeyBindings
	title       string
	isDirty     bool
	isDisabled  bool
	naturalX    int
	naturalY    int
	actualX     int
	actualY     int
	onAttach    func(v *view, w wi_core.Window)
	buffer      *wi_core.Buffer
}

func (v *view) Commands() wi_core.Commands {
	return v.commands
}

func (v *view) KeyBindings() wi_core.KeyBindings {
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
	v.buffer = wi_core.NewBuffer(x, y)
}

func (v *view) OnAttach(w wi_core.Window) {
	if v.onAttach != nil {
		v.onAttach(v, w)
	}
}

// A disabled static view.
type staticDisabledView struct {
	view
}

func (v *staticDisabledView) Buffer() *wi_core.Buffer {
	v.buffer.DrawString(v.Title(), 0, 0, wi_core.CellFormat{Fg: wi_core.Red, Bg: wi_core.Black})
	return v.buffer
}

// Empty non-editable window.
func makeStaticDisabledView(title string, naturalX, naturalY int) *staticDisabledView {
	return &staticDisabledView{
		view{
			commands:    makeCommands(),
			keyBindings: makeKeyBindings(),
			title:       title,
			isDisabled:  true,
			naturalX:    naturalX,
			naturalY:    naturalY,
		},
	}
}

// The status line is a hierarchy of Window, one for each element, each showing
// a single item.
func statusRootViewFactory(args ...string) wi_core.View {
	// TODO(maruel): OnResize(), query the root Window size, if y<=5 or x<=15,
	// set the root status Window to y=0, so that it becomes effectively
	// invisible when the editor window is too small.
	return makeStaticDisabledView("Status Root", 1, 1)
}

func statusNameViewFactory(args ...string) wi_core.View {
	// View name.
	// TODO(maruel): Register events of Window activation, make itself Invalidate().
	return makeStaticDisabledView("Status Name", 15, 1)
}

func statusPositionViewFactory(args ...string) wi_core.View {
	// Position, % of file.
	// TODO(maruel): Register events of movement, make itself Invalidate().
	return makeStaticDisabledView("Status Position", 15, 1)
}

type commandView struct {
	view
}

func (v *commandView) Buffer() *wi_core.Buffer {
	v.buffer.DrawString(v.Title(), 0, 0, wi_core.CellFormat{Fg: wi_core.Red, Bg: wi_core.Black})
	return v.buffer
}

// The command dialog box.
// TODO(maruel): Position it 5 lines below the cursor in the parent Window's
// View. Do this via onAttach.
func commandViewFactory(args ...string) wi_core.View {
	return &commandView{
		view{
			commands:    makeCommands(),
			keyBindings: makeKeyBindings(),
			title:       "Command",
			naturalX:    30,
			naturalY:    1,
		},
	}
}

type documentView struct {
	view
}

func (v *documentView) Buffer() *wi_core.Buffer {
	v.buffer.DrawString(v.Title(), 0, 0, wi_core.CellFormat{Fg: wi_core.Red, Bg: wi_core.Black})
	return v.buffer
}

func documentViewFactory(args ...string) wi_core.View {
	// TODO(maruel): Sort out "use max space".
	//onAttach
	return &documentView{
		view{
			commands:    makeCommands(),
			keyBindings: makeKeyBindings(),
			title:       "<Empty document>",
			naturalX:    100,
			naturalY:    100,
		},
	}
}

func infobarAlertViewFactory(args ...string) wi_core.View {
	out := "Alert: " + args[0]
	l := utf8.RuneCountInString(out)
	v := makeStaticDisabledView(out, l, 1)
	v.onAttach = func(v *view, w wi_core.Window) {
		go func() {
			// Dismiss after 5 seconds.
			<-time.After(5 * time.Second)
			wi_core.PostCommand(w, "window_close", w.ID())
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

func cmdViewLog(c *privilegedCommandImpl, e *editor, w *window, args ...string) {
	names := make([]string, 0, len(e.viewFactories))
	for k := range e.viewFactories {
		names = append(names, k)
	}
	sort.Strings(names)
	log.Printf("View factories:")
	for _, name := range names {
		log.Printf("  %s", name)
	}
}

// RegisterViewCommands registers view-related commands
func RegisterViewCommands(dispatcher wi_core.Commands) {
	defaultCommands := []wi_core.Command{
		&privilegedCommandImpl{
			"view_log",
			0,
			cmdViewLog,
			wi_core.DebugCategory,
			wi_core.LangMap{
				wi_core.LangEn: "Logs the view factories",
			},
			wi_core.LangMap{
				wi_core.LangEn: "Logs the view factories, this is only relevant if -verbose is used.",
			},
		},
	}
	for _, cmd := range defaultCommands {
		dispatcher.Register(cmd)
	}
}
