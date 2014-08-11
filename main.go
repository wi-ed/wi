// Copyright 2013 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// wi - Bringing text based editor technology past 1200 bauds. See README.md
// for more details.
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
	rootWindow     wi.Window
	lastActive     []wi.Window
	terminalEvents <-chan termbox.Event
	commandsQueue  chan commandQueueItem
	outputBuffer   tulib.Buffer
	languageMode   wi.LanguageMode
	keyboardMode   wi.KeyboardMode
}

func (t *terminal) Version() string {
	return version
}

func (t *terminal) PostCommand(cmdName string, args ...string) {
	log.Printf("PostCommand(%s, %s)", cmdName, args)
	t.commandsQueue <- commandQueueItem{cmdName, args, ""}
}

func (t *terminal) postKey(keyName string) {
	log.Printf("PostKey(%s)", keyName)
	t.commandsQueue <- commandQueueItem{keyName: keyName}
}

func (t *terminal) WaitQueueEmpty() {
	panic("Oops")
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

func (t *terminal) KeyboardMode() wi.KeyboardMode {
	return t.keyboardMode
}

func drawRecurse(w wi.Window, buffer *tulib.Buffer) {
	for _, child := range w.ChildrenWindows() {
		drawRecurse(child, buffer)
		if child.IsInvalid() {
			buffer.Blit(child.Rect(), 0, 0, child.Buffer())
		}
	}
}

// draw descends the whole Window tree and find the invalidated window to
// redraw.
func (t *terminal) draw() {
	log.Print("draw()")
	drawRecurse(t.rootWindow, &t.outputBuffer)

	if err := termbox.Flush(); err != nil {
		panic(err)
	}
}

func (t *terminal) ActiveWindow() wi.Window {
	return t.lastActive[0]
}

func (t *terminal) ActivateWindow(w wi.Window) {
	log.Printf("ActivateWindow(%s)", w.View().Title())
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
	keyBuffer := ""
	for {
		select {
		case i := <-t.commandsQueue:
			if i.keyName != "" {
				// Convert the key press into a command. The trick is that we don't
				// know the active window, there could be commands already enqueued
				// that will change the active window, so using the active window
				// directly or indirectly here is an incorrect assumption.
				if i.keyName == "<enter>" {
					t.ExecuteCommand(t.ActiveWindow(), keyBuffer)
					keyBuffer = ""
				} else {
					cmdName := wi.GetKeyBindingCommand(t, t.KeyboardMode(), i.keyName)
					if cmdName != "" {
						t.ExecuteCommand(t.ActiveWindow(), cmdName)
					} else if len(i.keyName) == 1 {
						keyBuffer += i.keyName
					}
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
		lastActive:     []wi.Window{rootWindow},
		languageMode:   wi.LangEn,
		keyboardMode:   wi.EditMode,
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
	windowBuffer    tulib.Buffer // includes the border
	rect            tulib.Rect
	childrenWindows []wi.Window
	view            wi.View
	docking         wi.DockingType
	border          wi.BorderType
	fg              termbox.Attribute
	bg              termbox.Attribute
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

func isEqual(lhs tulib.Rect, rhs tulib.Rect) bool {
	return lhs.X == rhs.X && lhs.Y == rhs.Y && lhs.Width == rhs.Width && lhs.Height == rhs.Height
}

var singleBorder = []rune{'\u2500', '\u2502', '\u250D', '\u2510', '\u2514', '\u2518'}
var doubleBorder = []rune{'\u2550', '\u2551', '\u2554', '\u2557', '\u255a', '\u255d'}

func (w *window) SetRect(rect tulib.Rect) {
	if isEqual(w.rect, rect) {
		return
	}
	w.rect = rect
	w.Invalidate()
	if w.border != wi.BorderNone {
		// Draw the borders right away.
		w.windowBuffer = tulib.NewBuffer(rect.Width, rect.Height)
		s := doubleBorder
		if w.border == wi.BorderSingle {
			s = singleBorder
		}
		w.windowBuffer.Set(0, 0, termbox.Cell{s[3], w.fg, w.bg})
		w.windowBuffer.Set(0, rect.Height-1, termbox.Cell{s[5], w.fg, w.bg})
		w.windowBuffer.Set(rect.Width-1, 0, termbox.Cell{s[4], w.fg, w.bg})
		w.windowBuffer.Set(rect.Width-1, rect.Height-1, termbox.Cell{s[6], w.fg, w.bg})
		w.windowBuffer.Fill(tulib.Rect{1, 0, rect.Width - 2, 0}, termbox.Cell{s[0], w.fg, w.bg})
		w.windowBuffer.Fill(tulib.Rect{1, rect.Height - 1, rect.Width - 2, rect.Height - 1}, termbox.Cell{s[0], w.fg, w.bg})
		w.windowBuffer.Fill(tulib.Rect{0, 1, 0, rect.Height - 2}, termbox.Cell{s[1], w.fg, w.bg})
		w.windowBuffer.Fill(tulib.Rect{rect.Width - 1, 1, rect.Width - 1, rect.Height - 2}, termbox.Cell{s[1], w.fg, w.bg})
	} else {
		w.windowBuffer = tulib.NewBuffer(0, 0)
	}
}

func (w *window) IsInvalid() bool {
	return w.isInvalid || w.View().IsInvalid()
}

func (w *window) Invalidate() {
	w.isInvalid = true
}

func (w *window) Buffer() *tulib.Buffer {
	if w.border == wi.BorderNone {
		w.isInvalid = false
		return w.view.Buffer()
	}
	if w.isInvalid {
		w.isInvalid = false
		// Ask the view to draw into its buffer.
		w.windowBuffer.Blit(tulib.Rect{1, 1, w.rect.Width - 2, w.rect.Height - 2}, 0, 0, w.view.Buffer())
	}
	return &w.windowBuffer
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
		fg:        termbox.ColorWhite,
		bg:        termbox.ColorBlack,
	}
}

func Main() int {
	log.SetFlags(log.Lmicroseconds | log.Lshortfile)
	command := flag.Bool("c", false, "Runs the commands specified on startup")
	version := flag.Bool("v", false, "Prints version and exit")
	verbose := flag.Bool("verbose", false, "Logs debugging information to wi.log")
	flag.Parse()

	// Process this one early. No one wants version output to take 1s.
	if *version {
		println(version)
		os.Exit(0)
	}

	if *verbose {
		if f, err := os.OpenFile("wi.log", os.O_CREATE|os.O_WRONLY, 0666); err == nil {
			defer func() {
				_ = f.Close()
			}()
			log.SetOutput(f)
		}
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
	return editor.eventLoop()
}

func main() {
	os.Exit(Main())
}
