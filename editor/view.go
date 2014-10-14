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
	v.buffer.Fill(wi.MakeCell(r, wi.Red, wi.Black))
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

type commandView struct {
	*view
}

func (v *commandView) Buffer() *wi.Buffer {
	r, _ := utf8.DecodeRuneInString(v.Title())
	v.buffer.Fill(wi.MakeCell(r, wi.Green, wi.Black))
	v.buffer.DrawString(v.Title(), 0, 0, wi.CellFormat{wi.Brown, wi.Black, false, false, false})
	return v.buffer
}

// The command dialog box.
// TODO(maruel): Position it 5 lines below the cursor in the parent Window's
// View. Do this via onAttach.
func commandViewFactory(args ...string) wi.View {
	return &commandView{makeView("Command", 30, 1)}
}

func documentViewFactory(args ...string) wi.View {
	// TODO(maruel): Sort out "use max space".
	//onAttach
	return makeView("", 100, 100)
}

func infobarAlertViewFactory(args ...string) wi.View {
	out := "Alert: " + args[0]
	l := utf8.RuneCountInString(out)
	v := makeView(out, l, 1)
	v.onAttach = func(v *view, w wi.Window) {
		go func() {
			// Dismiss after 5 seconds.
			<-time.After(5 * time.Second)
			wi.PostCommand(w, "window_close", w.ID())
		}()
	}
	return v
}

// RegisterDefaultViewFactories registers the builtins views factories.
func RegisterDefaultViewFactories(e Editor) {
	e.RegisterViewFactory("command", commandViewFactory)
	e.RegisterViewFactory("new_document", documentViewFactory)
	e.RegisterViewFactory("infobar_alert", infobarAlertViewFactory)
	e.RegisterViewFactory("status_name", statusNameViewFactory)
	e.RegisterViewFactory("status_position", statusPositionViewFactory)
	e.RegisterViewFactory("status_root", statusRootViewFactory)
}

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
func RegisterViewCommands(dispatcher wi.Commands) {
	defaultCommands := []wi.Command{
		&privilegedCommandImpl{
			"view_log",
			0,
			cmdViewLog,
			wi.DebugCategory,
			wi.LangMap{
				wi.LangEn: "Logs the view factories",
			},
			wi.LangMap{
				wi.LangEn: "Logs the view factories, this is only relevant if -verbose is used.",
			},
		},
	}
	for _, cmd := range defaultCommands {
		dispatcher.Register(cmd)
	}
}
