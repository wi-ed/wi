// Copyright 2013 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// wi - Bringing text based editor technology past 1200 bauds. See README.md for more details.
package main

import (
	"flag"
	"fmt"
	"github.com/maruel/wi/wi-plugin"
	"github.com/nsf/termbox-go"
	"github.com/nsf/tulib"
	"log"
	"os"
	"time"
)

const (
	// Major.Minor.Bugfix. All plugins should be recompiled with wi-plugin
	// changes.
	version = "0.0.1"
)

var quitFlag = false

// UI

type commandQueueItem struct {
	cmdName string
	args    []string
	keyName string
}

// It is normally expected to be drawn via an ssh/mosh connection so it should
// be "bandwidth" optimized, where bandwidth doesn't mean 1200 bauds anymore.
type terminal struct {
	rootWindow     wi.WindowFull
	lastActive     []wi.WindowFull
	terminalEvents <-chan termbox.Event
	commandsQueue  chan commandQueueItem
	outputBuffer   tulib.Buffer
	languageMode   wi.LanguageMode
}

func (t *terminal) Version() string {
	return version
}

func (t *terminal) PostCommand(cmdName string, args ...string) {
	t.commandsQueue <- commandQueueItem{cmdName, args, ""}
}

func (t *terminal) postKey(keyName string) {
	t.commandsQueue <- commandQueueItem{keyName: keyName}
}

func (t *terminal) WaitQueueEmpty() {
}

func (t *terminal) ExecuteCommand(w wi.Window, cmdName string, args ...string) {
	cmd := wi.GetCommand(t, w, cmdName)
	if cmd == nil {
		t.ExecuteCommand(w, "alert", fmt.Sprintf(getStr(t.CurrentLanguage(), notFound), cmdName))
	} else {
		cmd.Handle(t, w, args...)
	}
}

func (t *terminal) CurrentLanguage() wi.LanguageMode {
	return t.languageMode
}

func drawRecurse(w wi.Window, buffer tulib.Buffer) {
	for _, child := range w.ChildrenWindows() {
		drawRecurse(child, buffer)
		// TODO(maruel): Draw non-client area.
		if child.View().IsInvalid() {
			child.View().DrawInto(buffer)
		}
	}
}

// draw descends the whole Window tree and find the invalidated window to
// redraw.
func (t *terminal) draw() {
	drawRecurse(t.rootWindow, t.outputBuffer)

	if err := termbox.Flush(); err != nil {
		panic(err)
	}
}

func (t *terminal) ActiveWindow() wi.Window {
	return t.lastActive[0]
}

func (t *terminal) ActivateWindow(w wi.Window) {
	if w.View().IsDisabled() {
		t.ExecuteCommand(w, "alert", getStr(t.CurrentLanguage(), activateDisabled))
		return
	}

	// First remove w from t.lastActive, second add w as t.lastActive[0].
	// This kind of manual list shuffling is really Go's achille heel.
	// TODO(maruel): There's no way I got it right on the first try without a
	// unit test.
	for i, v := range t.lastActive {
		if v == w {
			if i > 0 {
				copy(t.lastActive[:i], t.lastActive[1:i+1])
				t.lastActive[0] = w
			}
			return
		}
	}

	// This Window has never been active.
	l := len(t.lastActive)
	t.lastActive = append(t.lastActive, nil)
	copy(t.lastActive[:l], t.lastActive[1:l])
	t.lastActive[0] = w
}

func (t *terminal) onResize() {
	// Recreate the buffer, which queries the new sizes.
	t.outputBuffer = tulib.TermboxBuffer()
	/*
		if err := termbox.Clear(termbox.ColorDefault, termbox.ColorDefault); err != nil {
			panic(err)
		}
	*/
	// Resize the Windows. This also invalidates it, which will also force a
	// redraw.
	t.rootWindow.SetRect(tulib.Rect{0, 0, t.outputBuffer.Width, t.outputBuffer.Height})
}

// eventLoop handles both commands and events from the terminal. This function
// runs in the UI goroutine.
func (t *terminal) eventLoop() int {
	fakeChan := make(chan time.Time)
	var drawTimer <-chan time.Time = fakeChan
	for {
		select {
		case i := <-t.commandsQueue:
			if i.keyName != "" {
				// Convert the key press into a command. The trick is that we don't
				// know the active window, there could be commands already enqueued
				// that will change the active window, so using the active window
				// directly or indirectly here is an incorrect assumption.
				cmdName := wi.GetKeyBindingCommand(t, i.keyName)
				if cmdName != "" {
					t.ExecuteCommand(t.ActiveWindow(), cmdName)
				}
			} else {
				t.ExecuteCommand(t.ActiveWindow(), i.cmdName, i.args...)
			}
			if quitFlag {
				drawTimer = time.After(1 * time.Millisecond)
			} else {
				// TODO(maruel): Only trigger when a Window was invalidated.
				drawTimer = time.After(15 * time.Millisecond)
			}

		case event := <-t.terminalEvents:
			switch event.Type {
			case termbox.EventKey:
				k := keyEventToName(event)
				if k != "" {
					t.postKey(k)
				}
			case termbox.EventMouse:
				// TODO(maruel): MouseDispatcher. Mouse events are expected to be
				// resolved to the window that is currently active, unlike key presses.
				// Life is inconsistent.
				break
			case termbox.EventResize:
				// The terminal window was resized, resize everything, independent of
				// the enqueued commands.
				t.onResize()
			case termbox.EventError:
				// TODO(maruel): Not sure what situations can trigger this.
				t.PostCommand("alert", event.Err.Error())
			}
			// TODO(maruel): Only trigger when a Window was invalidated.
			drawTimer = time.After(15 * time.Millisecond)

		case <-drawTimer:
			if quitFlag {
				return 0
			}
			t.draw()
			drawTimer = fakeChan
		}
	}
	return 0
}

// makeEditor creates the Editor object. The root window doesn't have
// anything to view in it. It will contain two child windows, the main content
// window and the status bar.
func makeEditor() *terminal {
	// The root view is important, it defines all the global commands. It is
	// pre-filled with the default native commands and keyboard mapping, and it's
	// up to the plugins to add more global commands on startup.
	rootView := makeView(-1, -1)
	RegisterDefaultCommands(rootView.Commands())

	rootWindow := makeWindow(nil, rootView, wi.DockingFill)
	terminalEvents := make(chan termbox.Event, 32)
	terminal := &terminal{
		terminalEvents: terminalEvents,
		commandsQueue:  make(chan commandQueueItem, 500),
		rootWindow:     rootWindow,
		lastActive:     []wi.WindowFull{rootWindow},
		languageMode:   wi.LangEn,
	}

	RegisterDefaultKeyBindings(terminal)

	terminal.onResize()
	go func() {
		for {
			terminalEvents <- termbox.PollEvent()
		}
	}()
	return terminal
}

// window implements Window. It keeps its own buffer of its display.
type window struct {
	parent          wi.Window
	buffer          tulib.Buffer // includes the border
	rect            tulib.Rect
	childrenWindows []wi.Window
	view            wi.View
	docking         wi.DockingType
	border          wi.BorderType
	isInvalid       bool
}

func (w *window) Parent() wi.Window {
	return w.parent
}

func (w *window) ChildrenWindows() []wi.Window {
	return w.childrenWindows[:]
}

func (w *window) NewChildWindow(view wi.View, docking wi.DockingType) wi.Window {
	child := makeWindow(w, view, docking)
	w.childrenWindows = append(w.childrenWindows, child)
	return child
}

func (w *window) Remove(child wi.Window) {
	for i, v := range w.childrenWindows {
		if v == child {
			copy(w.childrenWindows[i:], w.childrenWindows[i+1:])
			w.childrenWindows[len(w.childrenWindows)-1] = nil
			w.childrenWindows = w.childrenWindows[:len(w.childrenWindows)-1]
			return
		}
	}
	panic("Trying to remove a non-child Window")
}

func (w *window) Rect() tulib.Rect {
	return w.rect
}

func (w *window) SetRect(rect tulib.Rect) {
	// TODO(maruel): Add if !w.rect.IsEqual(rect) {}
	w.rect = rect
	w.buffer = tulib.NewBuffer(rect.Width, rect.Height)
	w.Invalidate()
}

func (w *window) Invalidate() {
	w.isInvalid = true
}

func (w *window) Buffer() tulib.Buffer {
	if w.isInvalid {
		// Ask the view to draw into its buffer.
		buffer := w.buffer
		if w.border != wi.BorderNone {
			// Create a temporary buffer for the view to draw into.
			// TODO(maruel): Make it smarter so we do not need to double-copy the data
			// constantly. It's very inefficient. At least it's bearable since the
			// amount of data will usually be under 200x200=40000 elements.
			buffer = tulib.NewBuffer(w.rect.Width-1, w.rect.Height-1)
			// TODO(maruel): Draw border.
		}
		w.view.DrawInto(buffer)
		// TODO(maruel): Copy back.
	}
	return w.buffer
}

func (w *window) Docking() wi.DockingType {
	return w.docking
}

func (w *window) SetDocking(docking wi.DockingType) {
	if w.docking != docking {
		w.docking = docking
		w.Invalidate()
	}
}

func (w *window) SetView(view wi.View) {
	if view != w.view {
		w.view = view
		w.Invalidate()
	}
}

func (w *window) View() wi.View {
	return w.view
}

func makeWindow(parent wi.Window, view wi.View, docking wi.DockingType) wi.Window {
	return &window{
		parent:    parent,
		view:      view,
		docking:   docking,
		border:    wi.BorderNone,
		isInvalid: true,
	}
}

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
	buffer      wi.TextBuffer
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

func (v *view) DrawInto(buffer tulib.Buffer) {
	// TODO(maruel): Plugable drawing function.
	buffer.Set(0, 0, termbox.Cell{'A', termbox.ColorRed, termbox.ColorRed})
}

func (v *view) NaturalSize() (x, y int) {
	return v.naturalX, v.naturalY
}

func (v *view) SetBuffer(buffer wi.TextBuffer) {
	v.buffer = buffer
}

func (v *view) Buffer() wi.TextBuffer {
	return v.buffer
}

// Empty non-editable window.
func makeView(naturalX, naturalY int) wi.View {
	return &view{
		commands:    makeCommands(),
		keyBindings: makeKeyBindings(),
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
	return makeView(1, -1)
}

func makeStatusViewName() wi.View {
	// View name.
	// TODO(maruel): Register events of Window activation, make itself Invalidate().
	// TODO(maruel): Drawing code.
	return makeView(1, -1)
}

func makeStatusViewPosition() wi.View {
	// Position, % of file.
	// TODO(maruel): Register events of movement, make itself Invalidate().
	// TODO(maruel): Drawing code.
	return makeView(1, -1)
}

// The command box.
func makeCommandView() wi.View {
	return makeView(1, -1)
}

// A dismissable modal dialog box. TODO(maruel): An infobar that auto-dismiss
// itself after 5s.
func makeAlertView() wi.View {
	return makeView(1, 1)
}

func main() {
	log.SetFlags(log.Lmicroseconds)
	command := flag.Bool("c", false, "Runs the commands specified on startup")
	version := flag.Bool("v", false, "Prints version and exit")
	flag.Parse()

	// Process this one early. No one wants version output to take 1s.
	if *version {
		println(version)
		os.Exit(0)
	}

	if err := termbox.Init(); err != nil {
		panic(err)
	}
	defer termbox.Close()
	termbox.SetInputMode(termbox.InputAlt | termbox.InputMouse)

	editor := makeEditor()
	plugins := loadPlugins(editor)
	defer func() {
		for _, p := range plugins {
			// TODO(maruel): Nicely terminate them.
			if err := p.Kill(); err != nil {
				panic(err)
			}
		}
	}()

	// Add the status bar. At that point plugins are loaded so they can override
	// add_status_bar if they want.
	editor.PostCommand("add_status_bar")

	if *command {
		for _, i := range flag.Args() {
			editor.PostCommand(i)
		}
	} else if flag.NArg() > 0 {
		for _, i := range flag.Args() {
			editor.PostCommand("open", i)
		}
	} else {
		// If nothing, opens a blank editor.
		editor.PostCommand("new")
	}

	// Run the message loop.
	out := editor.eventLoop()

	// Normal exit.
	termbox.SetCursor(0, 0)
	if err := termbox.Flush(); err != nil {
		panic(err)
	}
	os.Exit(out)
}
