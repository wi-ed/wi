// Copyright 2013 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package editor

import (
	"fmt"
	"log"
	"time"
	"unicode/utf8"

	"github.com/maruel/wi/wicore"
	"github.com/maruel/wi/wicore/colors"
	"github.com/maruel/wi/wicore/raster"
)

// TODO(maruel): Likely move into wicore for reuse.
type view struct {
	commands      wicore.Commands
	keyBindings   wicore.KeyBindings
	eventRegistry wicore.EventRegistry
	title         string
	isDisabled    bool
	naturalX      int // Desired size.
	naturalY      int
	actualX       int // Actual size in UI.
	actualY       int
	window        wicore.Window
	onAttach      func(v *view, w wicore.Window)
	defaultFormat raster.CellFormat
	buffer        *raster.Buffer
	events        []wicore.EventListener
}

// wicore.View interface.

func (v *view) String() string {
	return fmt.Sprintf("View(%s)", v.title)
}

func (v *view) Close() error {
	var err error
	for _, event := range v.events {
		err2 := event.Close()
		if err2 != nil {
			err = err2
		}
	}
	return err
}

func (v *view) Commands() wicore.Commands {
	return v.commands
}

func (v *view) KeyBindings() wicore.KeyBindings {
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
	v.buffer = raster.NewBuffer(x, y)
}

func (v *view) OnAttach(w wicore.Window) {
	if v.onAttach != nil {
		v.onAttach(v, w)
	}
	v.window = w
}

// DefaultFormat returns the View's format or the parent Window's View's format.
func (v *view) DefaultFormat() raster.CellFormat {
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

func (v *staticDisabledView) Buffer() *raster.Buffer {
	// TODO(maruel): Use the parent view format by default. No idea how to
	// surface this information here. Cost is at least a RPC, potentially
	// multiple when multiple plugins are involved in the tree.
	v.buffer.Fill(raster.Cell{' ', v.DefaultFormat()})
	v.buffer.DrawString(v.Title(), 0, 0, v.DefaultFormat())
	return v.buffer
}

// Empty non-editable window.
func makeStaticDisabledView(e wicore.Editor, title string, naturalX, naturalY int) *staticDisabledView {
	return &staticDisabledView{
		view{
			commands:      makeCommands(),
			keyBindings:   makeKeyBindings(),
			eventRegistry: e,
			title:         title,
			isDisabled:    true,
			naturalX:      naturalX,
			naturalY:      naturalY,
			defaultFormat: raster.CellFormat{Fg: colors.Red, Bg: colors.Black},
			events:        []wicore.EventListener{},
		},
	}
}

// The status line is a hierarchy of Window, one for each element, each showing
// a single item.
func statusRootViewFactory(e wicore.Editor, args ...string) wicore.View {
	// TODO(maruel): OnResize(), query the root Window size, if y<=5 or x<=15,
	// set the root status Window to y=0, so that it becomes effectively
	// invisible when the editor window is too small.
	v := makeStaticDisabledView(e, "Status Root", 1, 1)
	v.defaultFormat.Bg = colors.LightGray
	v.onAttach = func(v *view, w wicore.Window) {
		id := w.ID()
		e.TriggerCommands(
			wicore.EnqueuedCommands{
				[][]string{
					{"window_new", id, "left", "status_active_window_name"},
					{"window_new", id, "right", "status_position"},
					{"window_new", id, "fill", "status_mode"},
				},
				nil,
			})
	}
	return v
}

func statusActiveWindowNameViewFactory(e wicore.Editor, args ...string) wicore.View {
	// Active Window View name.
	// TODO(maruel): Register events of Window activation, make itself Invalidate().
	v := makeStaticDisabledView(e, "Status Name", 15, 1)
	v.defaultFormat = raster.CellFormat{}
	return v
}

func statusModeViewFactory(e wicore.Editor, args ...string) wicore.View {
	// Mostly for testing purpose, will contain the current mode "Insert" or "Command".
	v := makeStaticDisabledView(e, e.KeyboardMode().String(), 10, 1)
	v.defaultFormat = raster.CellFormat{}
	event := e.RegisterEditorKeyboardModeChanged(func(mode wicore.KeyboardMode) {
		v.title = mode.String()
	})
	v.events = append(v.events, event)
	return v
}

func statusPositionViewFactory(e wicore.Editor, args ...string) wicore.View {
	// Position, % of file.
	// TODO(maruel): Register events of movement, make itself Invalidate().
	v := makeStaticDisabledView(e, "Status Position", 15, 1)
	v.defaultFormat = raster.CellFormat{}
	event := e.RegisterDocumentCursorMoved(func(doc wicore.Document, col, row int) {
		v.title = fmt.Sprintf("%d,%d", col, row)
	})
	v.events = append(v.events, event)
	return v
}

func infobarAlertViewFactory(e wicore.Editor, args ...string) wicore.View {
	out := "Alert: " + args[0]
	l := utf8.RuneCountInString(out)
	v := makeStaticDisabledView(e, out, l, 1)
	v.onAttach = func(v *view, w wicore.Window) {
		wicore.Go("infobarAlert", func() {
			// Dismiss after 5 seconds.
			<-time.After(5 * time.Second)
			wicore.PostCommand(e, nil, "window_close", w.ID())
		})
	}
	return v
}

// RegisterDefaultViewFactories registers the builtins views factories.
func RegisterDefaultViewFactories(e Editor) {
	e.RegisterViewFactory("command", commandViewFactory)
	e.RegisterViewFactory("infobar_alert", infobarAlertViewFactory)
	e.RegisterViewFactory("new_document", documentViewFactory)
	e.RegisterViewFactory("status_active_window_name", statusActiveWindowNameViewFactory)
	e.RegisterViewFactory("status_mode", statusModeViewFactory)
	e.RegisterViewFactory("status_position", statusPositionViewFactory)
	e.RegisterViewFactory("status_root", statusRootViewFactory)
}

// Commands

// RegisterViewCommands registers view-related commands
func RegisterViewCommands(dispatcher wicore.Commands) {
	cmds := []wicore.Command{}
	for _, cmd := range cmds {
		dispatcher.Register(cmd)
	}
}
