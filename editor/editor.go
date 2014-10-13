// Copyright 2013 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package editor

import (
	"fmt"
	"io"
	"log"
	"time"

	"github.com/maruel/wi/wi-plugin"
)

const (
	// Major.Minor.Bugfix. All plugins should be recompiled with wi-plugin
	// changes.
	version = "0.0.1"
)

// commandItem is a command pending to be executed.
type commandItem struct {
	cmdName string   // Set on command execution
	args    []string // Set on command execution
	key     KeyPress // Set on key press
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

// editor is the global structure that holds everything together. It implements
// the Editor interface.
type editor struct {
	terminal       Terminal                  // Abstract terminal interface to the real terminal.
	rootWindow     *window                   // The rootWindow is always DockingFill and set to the size of the terminal.
	lastActive     []wi.Window               // Most recently used order of Window activatd.
	viewFactories  map[string]wi.ViewFactory // All the ViewFactory's that can be used to create new View.
	terminalEvents <-chan TerminalEvent      // Events coming from Terminal.SeedEvents().
	viewReady      chan bool                 // A View.Buffer() is ready to be drawn.
	commandsQueue  chan commandQueueItem     // Pending commands to be executed.
	languageMode   wi.LanguageMode           // Actual language used.
	keyboardMode   wi.KeyboardMode           // Global keyboard mode is either CommandMode or EditMode.
	plugins        Plugins                   // All loaded plugin processes.
	quitFlag       bool                      // If true, a shutdown is in progress.
}

func (e *editor) Close() error {
	if e.plugins == nil {
		return nil
	}
	err := e.plugins.Close()
	e.plugins = nil
	return err
}

func (e *editor) Version() string {
	return version
}

func (e *editor) PostCommands(cmds [][]string) {
	log.Printf("PostCommands(%s)", cmds)
	tmp := make(commandQueueItem, len(cmds))
	for i, cmd := range cmds {
		tmp[i].cmdName = cmd[0]
		tmp[i].args = cmd[1:]
	}
	e.commandsQueue <- tmp
}

func (e *editor) postKey(key KeyPress) {
	log.Printf("PostKey(%s)", key)
	e.commandsQueue <- commandQueueItem{commandItem{key: key}}
}

func (e *editor) ExecuteCommand(w wi.Window, cmdName string, args ...string) {
	log.Printf("ExecuteCommand(%s, %s, %s)", w, cmdName, args)
	if w == nil {
		w = e.ActiveWindow()
	}
	cmd := wi.GetCommand(e, w, cmdName)
	if cmd == nil {
		e.ExecuteCommand(w, "alert", fmt.Sprintf(wi.GetStr(e.CurrentLanguage(), notFound), cmdName))
	} else {
		cmd.Handle(e, w, args...)
	}
}

func (e *editor) CurrentLanguage() wi.LanguageMode {
	return e.languageMode
}

func (e *editor) KeyboardMode() wi.KeyboardMode {
	return e.keyboardMode
}

// draw descends the whole Window tree and redraw Windows.
func (e *editor) draw() {
	log.Print("draw()")
	// TODO(maruel): Cache the buffer.
	w, h := e.terminal.Size()
	out := wi.NewBuffer(w, h)
	drawRecurse(e.rootWindow, 0, 0, out)
	e.terminal.Blit(out)
}

func (e *editor) ActiveWindow() wi.Window {
	return e.lastActive[0]
}

func (e *editor) activateWindow(w wi.Window) {
	log.Printf("ActivateWindow(%s)", w.View().Title())
	if w.View().IsDisabled() {
		e.ExecuteCommand(w, "alert", wi.GetStr(e.CurrentLanguage(), activateDisabled))
		return
	}

	// First remove w from e.lastActive, second add w as e.lastActive[0].
	// This kind of manual list shuffling is really Go's achille heel.
	// TODO(maruel): There's no way I got it right on the first try without a
	// unit test.
	for i, v := range e.lastActive {
		if v == w {
			if i > 0 {
				copy(e.lastActive[:i], e.lastActive[1:i+1])
				e.lastActive[0] = w
			}
			return
		}
	}

	// This Window has never been active.
	l := len(e.lastActive)
	e.lastActive = append(e.lastActive, nil)
	copy(e.lastActive[:l], e.lastActive[1:l])
	e.lastActive[0] = w
}

func (e *editor) RegisterViewFactory(name string, viewFactory wi.ViewFactory) bool {
	_, present := e.viewFactories[name]
	e.viewFactories[name] = viewFactory
	return !present
}

func (e *editor) onResize() {
	// Resize the Windows. This also invalidates it, which will also force a
	// redraw if the size changed.
	w, h := e.terminal.Size()
	e.rootWindow.SetRect(wi.Rect{0, 0, w, h})
}

// EventLoop handles both commands and events from the editor. This function
// runs in the UI goroutine.
func (e *editor) EventLoop() int {
	fakeChan := make(chan time.Time)
	var drawTimer <-chan time.Time = fakeChan
	keyBuffer := ""
	for {
		select {
		case cmds := <-e.commandsQueue:
			for _, cmd := range cmds {
				if cmd.key.IsValid() {
					keyName := cmd.key.String()
					if cmd.key.IsMeta() {
						if keyName == "Enter" {
							// TODO(maruel): Temporary, until the command window works.
							e.ExecuteCommand(e.ActiveWindow(), keyBuffer)
							keyBuffer = ""
						} else {
							// Convert the key press into a command. The trick is that we
							// don't know the active window, there could be commands already
							// enqueued that will change the active window, so using the
							// active window directly or indirectly here is an incorrect
							// assumption.
							cmdName := wi.GetKeyBindingCommand(e, e.KeyboardMode(), keyName)
							if cmdName != "" {
								e.ExecuteCommand(e.ActiveWindow(), cmdName)
							} else {
								e.ExecuteCommand(e.ActiveWindow(), "alert", fmt.Sprintf(wi.GetStr(e.CurrentLanguage(), notMapped), keyName))
							}
						}
					} else {
						if keyName != "" {
							// Accumulate normal key presses.
							keyBuffer += keyName
						}
					}
				} else {
					// A normal command.
					e.ExecuteCommand(e.ActiveWindow(), cmd.cmdName, cmd.args...)
				}
			}

		case event := <-e.terminalEvents:
			switch event.Type {
			case EventKey:
				if event.Key.IsValid() {
					e.postKey(event.Key)
				}
			case EventResize:
				// The terminal window was resized, resize everything, independent of
				// the enqueued commands.
				e.onResize()
			}

		case <-e.viewReady:
			// Taking in account a 60hz frame is 18.8ms, 5ms is going to be generally
			// processed within the same frame. This delaying results in significant
			// bandwidth saving on loading.
			if drawTimer == fakeChan {
				drawTimer = time.After(5 * time.Millisecond)
			}

		case <-drawTimer:
			if e.quitFlag {
				return 0
			}

			// Empty e.viewReady first.
		EmptyViewReady:
			for {
				select {
				case <-e.viewReady:
				default:
					break EmptyViewReady
				}
			}

			e.draw()
			drawTimer = fakeChan
		}
	}
	return 0
}

func (e *editor) LoadPlugins() error {
	// TODO(maruel): Get path.
	paths, err := EnumPlugins(".")
	if err != nil {
		return err
	}
	e.plugins = loadPlugins(paths)
	return nil
}

// MakeEditor creates an object that implements the Editor interface. The root
// window doesn't have anything to view in it.
//
// It's up to the caller to add child Windows in it. Normally it will be done
// via the command "editor_bootstrap_ui" to add the status bar, then "new" or
// "open" to create the initial text buffer.
//
// It is fine to run it concurrently in unit test, as no global variable shall
// be used by the object created by this function.
//
// Editor is closed by this function.
func MakeEditor(terminal Terminal, noPlugin bool) (Editor, error) {
	// The root view is important, it defines all the global commands. It is
	// pre-filled with the default native commands and keyboard mapping, and it's
	// up to the plugins to add more global commands on startup.
	rootView := makeView("Root", -1, -1)

	// These commands are generic commands, they do not require specific access.
	RegisterDefaultCommands(rootView.Commands())
	RegisterKeyBindingCommands(rootView.Commands())
	RegisterViewCommands(rootView.Commands())
	RegisterWindowCommands(rootView.Commands())

	rootWindow := makeWindow(nil, rootView, wi.DockingFill)
	e := &editor{
		terminal:       terminal,
		rootWindow:     rootWindow,
		lastActive:     []wi.Window{rootWindow},
		viewFactories:  make(map[string]wi.ViewFactory),
		terminalEvents: terminal.SeedEvents(),
		viewReady:      make(chan bool),
		commandsQueue:  make(chan commandQueueItem, 500),
		languageMode:   wi.LangEn,
		keyboardMode:   wi.EditMode,
	}
	rootWindow.cd = e

	RegisterDefaultViewFactories(e)

	// This forces creating the default buffer.
	e.onResize()

	if !noPlugin {
		if err := e.LoadPlugins(); err != nil {
			_ = e.Close()
			return nil, err
		}
	}

	// Key bindings are loaded after the plugins, so a plugin has the chance to
	// hook key_bind if desired.
	RegisterDefaultKeyBindings(e)
	return e, nil
}
