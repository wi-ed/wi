// Copyright 2013 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// wi-editor - Bringing text based editor technology past 1200 bauds.
//
// This package contains the unit-testable part of the wi editor. It is in a
// package so godoc will generate documentation for this code. While this
// package is not meant to be a general purpose reusable package, the primary
// maintainer of this project likes having a web browseable documentation.
// Using a package is an effective workaround the fact that godoc doesn't
// general documentation for "main" package.
//
// See ../README.md for more details.
package editor

import (
	"fmt"
	"github.com/maruel/tulib"
	"github.com/maruel/wi/wi-plugin"
	"github.com/nsf/termbox-go"
	"io"
	"log"
	"time"
)

const (
	// Major.Minor.Bugfix. All plugins should be recompiled with wi-plugin
	// changes.
	version = "0.0.1"
)

var quitFlag = false

// TermBox is the interface to termbox so it can be mocked in unit test.
type TermBox interface {
	Size() (int, int)
	Flush()
	PollEvent() termbox.Event
	Buffer() tulib.Buffer
}

type termBoxImpl struct {
}

func (t termBoxImpl) Size() (int, int) {
	return termbox.Size()
}

func (t termBoxImpl) Flush() {
	if err := termbox.Flush(); err != nil {
		panic(err)
	}
}

func (t termBoxImpl) PollEvent() termbox.Event {
	return termbox.PollEvent()
}

func (t termBoxImpl) Buffer() tulib.Buffer {
	w, h := t.Size()
	return tulib.Buffer{
		Cells: termbox.CellBuffer(),
		Rect:  tulib.Rect{0, 0, w, h},
	}
}

// Logger is the interface to log to. It must be used instead of
// log.Logger.Printf() or testing.T.Log(). This permits to collect logs for a
// complete test case.
type Logger interface {
	Logf(format string, v ...interface{})
}

// commandItem is a command pending to be executed.
type commandItem struct {
	cmdName string
	args    []string
	keyName string
}

// commandQueueItem is a set of commandItem pending to be executed.
type commandQueueItem []commandItem

// Editor is the inprocess wi.Editor interface. It adds the process life-time
// management functions to the public interface wi.Editor.
//
// It is very important to call the Close() function upon termination.
type Editor interface {
	io.Closer

	wi.Editor

	// Loads the plugins. This function should be called early but can be skipped
	// in case the plugins shouldn't be loaded.
	LoadPlugins() error

	// EventLoop runs the event loop until the command "quit" executes
	// successfully.
	EventLoop() int
}

// It is normally expected to be drawn via an ssh/mosh connection so it should
// be "bandwidth" optimized, where bandwidth doesn't mean 1200 bauds anymore.
type terminal struct {
	termBox        TermBox
	rootWindow     *window
	lastActive     []wi.Window
	viewFactories  map[string]wi.ViewFactory
	terminalEvents <-chan termbox.Event
	viewReady      chan bool // A View.Buffer() is ready to be drawn.
	commandsQueue  chan commandQueueItem
	languageMode   wi.LanguageMode
	keyboardMode   wi.KeyboardMode
	plugins        Plugins
}

func (t *terminal) Close() error {
	if t.plugins == nil {
		return nil
	}
	err := t.plugins.Close()
	t.plugins = nil
	return err
}

func (t *terminal) Version() string {
	return version
}

func (t *terminal) PostCommands(cmds [][]string) {
	log.Printf("PostCommands(%s)", cmds)
	tmp := make(commandQueueItem, len(cmds))
	for i, cmd := range cmds {
		tmp[i].cmdName = cmd[0]
		tmp[i].args = cmd[1:]
	}
	t.commandsQueue <- tmp
}

func (t *terminal) postKey(keyName string) {
	log.Printf("PostKey(%s)", keyName)
	t.commandsQueue <- commandQueueItem{commandItem{keyName: keyName}}
}

func (t *terminal) ExecuteCommand(w wi.Window, cmdName string, args ...string) {
	log.Printf("ExecuteCommand(%s)", cmdName)
	cmd := wi.GetCommand(t, w, cmdName)
	if cmd == nil {
		t.ExecuteCommand(w, "alert", fmt.Sprintf(wi.GetStr(t.CurrentLanguage(), notFound), cmdName))
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

// draw descends the whole Window tree and redraw Windows.
func (t *terminal) draw() {
	log.Print("draw()")
	b := tulib.TermboxBuffer()
	drawRecurse(t.rootWindow, 0, 0, &b)
	t.termBox.Flush()
}

func (t *terminal) ActiveWindow() wi.Window {
	return t.lastActive[0]
}

func (t *terminal) ActivateWindow(w wi.Window) {
	log.Printf("ActivateWindow(%s)", w.View().Title())
	if w.View().IsDisabled() {
		t.ExecuteCommand(w, "alert", wi.GetStr(t.CurrentLanguage(), activateDisabled))
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

func (t *terminal) PostDraw() {
	go func() {
		t.viewReady <- true
	}()
}

func (t *terminal) RegisterViewFactory(name string, viewFactory wi.ViewFactory) bool {
	_, present := t.viewFactories[name]
	t.viewFactories[name] = viewFactory
	return !present
}

func (t *terminal) onResize() {
	// Resize the Windows. This also invalidates it, which will also force a
	// redraw if the size changed.
	w, h := t.termBox.Size()
	t.rootWindow.SetRect(tulib.Rect{0, 0, w, h})
}

// EventLoop handles both commands and events from the terminal. This function
// runs in the UI goroutine.
func (t *terminal) EventLoop() int {
	fakeChan := make(chan time.Time)
	var drawTimer <-chan time.Time = fakeChan
	keyBuffer := ""
	for {
		select {
		case cmds := <-t.commandsQueue:
			for _, cmd := range cmds {
				// TODO(maruel): Temporary, until the command window works.
				if cmd.keyName != "" {
					// Convert the key press into a command. The trick is that we don't
					// know the active window, there could be commands already enqueued
					// that will change the active window, so using the active window
					// directly or indirectly here is an incorrect assumption.
					if cmd.keyName == "Enter" {
						t.ExecuteCommand(t.ActiveWindow(), keyBuffer)
						keyBuffer = ""
					} else {
						cmdName := wi.GetKeyBindingCommand(t, t.KeyboardMode(), cmd.keyName)
						if cmdName != "" {
							t.ExecuteCommand(t.ActiveWindow(), cmdName)
						} else if len(cmd.keyName) == 1 {
							keyBuffer += cmd.keyName
						}
					}
				} else {
					t.ExecuteCommand(t.ActiveWindow(), cmd.cmdName, cmd.args...)
				}
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
				wi.PostCommand(t, "alert", event.Err.Error())
			}

		case <-t.viewReady:
			// Taking in account a 60hz frame is 18.8ms, 5ms is going to be generally
			// processed within the same frame. This delaying results in significant
			// bandwidth saving on loading.
			if drawTimer == fakeChan {
				drawTimer = time.After(5 * time.Millisecond)
			}

		case <-drawTimer:
			if quitFlag {
				return 0
			}

			// Empty t.viewReady first.
		EmptyViewReady:
			for {
				select {
				case <-t.viewReady:
				default:
					break EmptyViewReady
				}
			}

			t.draw()
			drawTimer = fakeChan
		}
	}
	return 0
}

func (t *terminal) LoadPlugins() error {
	// TODO(maruel): Get path.
	paths, err := EnumPlugins(".")
	if err != nil {
		return err
	}
	t.plugins = loadPlugins(paths)
	return nil
}

// MakeEditor creates the Editor object. The root window doesn't have
// anything to view in it.
//
// It's up to the caller to add child Windows in it. Normally it will be done
// via the command "bootstrap_ui" to add the status bar, then "new" or "open"
// to create the initial text buffer.
func MakeEditor(termBox TermBox) *terminal {
	// The root view is important, it defines all the global commands. It is
	// pre-filled with the default native commands and keyboard mapping, and it's
	// up to the plugins to add more global commands on startup.
	rootView := makeView("Root", -1, -1)

	// These commands are generic commands, they do not require specific access.
	RegisterDefaultCommands(rootView.Commands())

	if termBox == nil {
		termBox = termBoxImpl{}
	}

	rootWindow := makeWindow(nil, rootView, wi.DockingFill)
	terminalEvents := make(chan termbox.Event, 32)
	terminal := &terminal{
		termBox:        termBox,
		rootWindow:     rootWindow,
		lastActive:     []wi.Window{rootWindow},
		terminalEvents: terminalEvents,
		viewReady:      make(chan bool),
		commandsQueue:  make(chan commandQueueItem, 500),
		languageMode:   wi.LangEn,
		keyboardMode:   wi.EditMode,
	}
	rootWindow.cd = terminal

	// These commands have intimate knowledge of the terminal.
	RegisterWindowCommands(terminal, rootView.Commands())

	RegisterDefaultKeyBindings(terminal)

	terminal.onResize()
	go func() {
		for {
			terminalEvents <- terminal.termBox.PollEvent()
		}
	}()
	return terminal
}

// Main() is the unit-testable part of Main() that is called by the "main"
// package.
//
// It is fine to run it concurrently in unit test, as no global variable shall
// be used by this function.
func Main(noPlugin bool, editor Editor) int {
	if !noPlugin {
		if err := editor.LoadPlugins(); err != nil {
			fmt.Printf("%s", err)
			return 1
		}
	}
	defer editor.Close()

	// Run the message loop.
	return editor.EventLoop()
}
