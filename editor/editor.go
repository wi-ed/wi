// Copyright 2013 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package editor

import (
	"fmt"
	"io"
	"log"
	"time"

	"github.com/maruel/wi/wi_core"
)

const (
	// Major.Minor.Bugfix. All plugins should be recompiled with wi_core
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

// Editor is the inprocess wi_core.Editor interface. It adds the process life-time
// management functions to the public interface wi_core.Editor.
//
// It is very important to call the Close() function upon termination.
type Editor interface {
	io.Closer

	wi_core.Editor

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
	terminal       Terminal                       // Abstract terminal interface to the real terminal.
	rootWindow     *window                        // The rootWindow is always DockingFill and set to the size of the terminal.
	lastActive     []wi_core.Window               // Most recently used order of Window activatd.
	viewFactories  map[string]wi_core.ViewFactory // All the ViewFactory's that can be used to create new View.
	terminalEvents <-chan TerminalEvent           // Events coming from Terminal.SeedEvents().
	viewReady      chan bool                      // A View.Buffer() is ready to be drawn.
	commandsQueue  chan commandQueueItem          // Pending commands to be executed.
	languageMode   wi_core.LanguageMode           // Actual language used.
	keyboardMode   wi_core.KeyboardMode           // Global keyboard mode is either CommandMode or EditMode.
	plugins        Plugins                        // All loaded plugin processes.
	quitFlag       bool                           // If true, a shutdown is in progress.
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

func (e *editor) ExecuteCommand(w wi_core.Window, cmdName string, args ...string) {
	log.Printf("ExecuteCommand(%s, %s, %s)", w, cmdName, args)
	if w == nil {
		w = e.ActiveWindow()
	}
	cmd := wi_core.GetCommand(e, w, cmdName)
	if cmd == nil {
		e.ExecuteCommand(w, "alert", fmt.Sprintf(wi_core.GetStr(e.CurrentLanguage(), notFound), cmdName))
	} else {
		cmd.Handle(e, w, args...)
	}
}

func (e *editor) CurrentLanguage() wi_core.LanguageMode {
	return e.languageMode
}

func (e *editor) KeyboardMode() wi_core.KeyboardMode {
	return e.keyboardMode
}

// draw descends the whole Window tree and redraw Windows.
func (e *editor) draw() {
	log.Print("draw()")
	// TODO(maruel): Cache the buffer.
	w, h := e.terminal.Size()
	out := wi_core.NewBuffer(w, h)
	drawRecurse(e.rootWindow, 0, 0, out)
	e.terminal.Blit(out)
}

func (e *editor) ActiveWindow() wi_core.Window {
	return e.lastActive[0]
}

func (e *editor) activateWindow(w wi_core.Window) {
	log.Printf("ActivateWindow(%s)", w.View().Title())
	if w.View().IsDisabled() {
		e.ExecuteCommand(w, "alert", wi_core.GetStr(e.CurrentLanguage(), activateDisabled))
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

func (e *editor) RegisterViewFactory(name string, viewFactory wi_core.ViewFactory) bool {
	_, present := e.viewFactories[name]
	e.viewFactories[name] = viewFactory
	return !present
}

func (e *editor) onResize() {
	// Resize the Windows. This also invalidates it, which will also force a
	// redraw if the size changed.
	w, h := e.terminal.Size()
	e.rootWindow.setRect(wi_core.Rect{0, 0, w, h})
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
							cmdName := wi_core.GetKeyBindingCommand(e, e.KeyboardMode(), keyName)
							if cmdName != "" {
								e.ExecuteCommand(e.ActiveWindow(), cmdName)
							} else {
								e.ExecuteCommand(e.ActiveWindow(), "alert", fmt.Sprintf(wi_core.GetStr(e.CurrentLanguage(), notMapped), keyName))
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
	rootView := makeStaticDisabledView("Root", -1, -1)

	// These commands are generic commands, they do not require specific access.
	RegisterCommandCommands(rootView.Commands())
	RegisterKeyBindingCommands(rootView.Commands())
	RegisterViewCommands(rootView.Commands())
	RegisterWindowCommands(rootView.Commands())
	RegisterDocumentCommands(rootView.Commands())
	RegisterEditorCommands(rootView.Commands())

	rootWindow := makeWindow(nil, rootView, wi_core.DockingFill)
	e := &editor{
		terminal:       terminal,
		rootWindow:     rootWindow,
		lastActive:     []wi_core.Window{rootWindow},
		viewFactories:  make(map[string]wi_core.ViewFactory),
		terminalEvents: terminal.SeedEvents(),
		viewReady:      make(chan bool),
		commandsQueue:  make(chan commandQueueItem, 500),
		languageMode:   wi_core.LangEn,
		keyboardMode:   wi_core.EditMode,
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

// Commands

func cmdAlert(c *wi_core.CommandImpl, cd wi_core.CommandDispatcherFull, w wi_core.Window, args ...string) {
	cd.ExecuteCommand(w, "window_new", "0", "bottom", "infobar_alert", args[0])
}

func cmdEditorBootstrapUI(c *wi_core.CommandImpl, cd wi_core.CommandDispatcherFull, w wi_core.Window, args ...string) {
	// TODO(maruel): Use onAttach instead of hard coding names.
	cd.ExecuteCommand(w, "window_new", "0", "bottom", "status_root")
	cd.ExecuteCommand(w, "window_new", "0:1", "left", "status_name")
	cd.ExecuteCommand(w, "window_new", "0:1", "right", "status_position")
}

func isDirtyRecurse(cd wi_core.CommandDispatcherFull, w wi_core.Window) bool {
	for _, child := range w.ChildrenWindows() {
		if isDirtyRecurse(cd, child) {
			return true
		}
		v := child.View()
		if v.IsDirty() {
			cd.ExecuteCommand(w, "alert", fmt.Sprintf(wi_core.GetStr(cd.CurrentLanguage(), viewDirty), v.Title()))
			return true
		}
	}
	return false
}

func cmdEditorQuit(c *privilegedCommandImpl, e *editor, w *window, args ...string) {
	if len(args) >= 1 {
		e.ExecuteCommand(w, "alert", c.LongDesc(e, w))
		return
	} else if len(args) == 1 {
		if args[0] != "force" {
			e.ExecuteCommand(w, "alert", c.LongDesc(e, w))
			return
		}
	} else {
		// TODO(maruel): For all the View, question if fine to quit via
		// view.IsDirty(). If not fine, "prompt" y/n to force quit. If n, stop
		// there.
		// - Send a signal to each plugin.
		// - Send a signal back to the main loop.
		if isDirtyRecurse(e, e.rootWindow) {
			return
		}
	}

	e.quitFlag = true
	// editor_redraw wakes up the command event loop so it detects it's time to
	// quit.
	wi_core.PostCommand(e, "editor_redraw")
}

func cmdEditorRedraw(c *privilegedCommandImpl, e *editor, w *window, args ...string) {
	go func() {
		e.viewReady <- true
	}()
}

func cmdCommandAlias(c *wi_core.CommandImpl, cd wi_core.CommandDispatcherFull, w wi_core.Window, args ...string) {
	if args[0] == "window" {
	} else if args[0] == "global" {
		w = wi_core.RootWindow(w)
	} else {
		cmd := wi_core.GetCommand(cd, w, "command_alias")
		cd.ExecuteCommand(w, "alert", cmd.LongDesc(cd, w))
		return
	}
	alias := &wi_core.CommandAlias{args[1], args[2], nil}
	w.View().Commands().Register(alias)
}

func cmdShowCommandWindow(c *wi_core.CommandImpl, cd wi_core.CommandDispatcherFull, w wi_core.Window, args ...string) {
	// Create the Window with the command view and attach it to the currently
	// focused Window.
	cd.ExecuteCommand(w, "window_new", w.ID(), "floating", "command")
}

// RegisterEditorCommands registers the top-level native commands.
func RegisterEditorCommands(dispatcher wi_core.Commands) {
	defaultCommands := []wi_core.Command{
		&wi_core.CommandImpl{
			"alert",
			1,
			cmdAlert,
			wi_core.WindowCategory,
			wi_core.LangMap{
				wi_core.LangEn: "Shows a modal message",
			},
			wi_core.LangMap{
				wi_core.LangEn: "Prints a message in a modal dialog box.",
			},
		},
		&wi_core.CommandImpl{
			"editor_bootstrap_ui",
			0,
			cmdEditorBootstrapUI,
			wi_core.WindowCategory,
			wi_core.LangMap{
				wi_core.LangEn: "Bootstraps the editor's UI",
			},
			wi_core.LangMap{
				wi_core.LangEn: "Bootstraps the editor's UI. This command is automatically run on startup and cannot be executed afterward. It adds the standard status bar. This command exists so it can be overriden by a plugin, so it can create its own status bar.",
			},
		},
		&privilegedCommandImpl{
			"editor_quit",
			-1,
			cmdEditorQuit,
			wi_core.EditorCategory,
			wi_core.LangMap{
				wi_core.LangEn: "Quits",
			},
			wi_core.LangMap{
				wi_core.LangEn: "Quits the editor. Use 'force' to bypasses writing the files to disk.",
			},
		},
		&privilegedCommandImpl{
			"editor_redraw",
			0,
			cmdEditorRedraw,
			wi_core.EditorCategory,
			wi_core.LangMap{
				wi_core.LangEn: "Forcibly redraws the terminal",
			},
			wi_core.LangMap{
				wi_core.LangEn: "Forcibly redraws the terminal.",
			},
		},
		&wi_core.CommandImpl{
			"show_command_window",
			0,
			cmdShowCommandWindow,
			wi_core.CommandsCategory,
			wi_core.LangMap{
				wi_core.LangEn: "Shows the interactive command window",
			},
			wi_core.LangMap{
				wi_core.LangEn: "This commands exists so it can be bound to a key to pop up the interactive command window.",
			},
		},
		&wi_core.CommandAlias{"q", "editor_quit", nil},
		&wi_core.CommandAlias{"q!", "editor_quit", []string{"force"}},
		&wi_core.CommandAlias{"quit", "editor_quit", nil},
	}
	for _, cmd := range defaultCommands {
		dispatcher.Register(cmd)
	}
}
