// Copyright 2013 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package editor

import (
	"io"
	"log"
	"time"

	"github.com/wi-ed/wi/wicore"
	"github.com/wi-ed/wi/wicore/key"
	"github.com/wi-ed/wi/wicore/lang"
	"github.com/wi-ed/wi/wicore/raster"
)

const (
	// Major.Minor.Bugfix.
	version = "0.0.1"
)

// Editor is the inprocess wicore.Editor interface. It adds the process life-time
// management functions to the public interface wicore.Editor.
//
// It is very important to call the Close() function upon termination.
type Editor interface {
	io.Closer

	wicore.EditorW

	// EventLoop runs the event loop until the command "quit" executes
	// successfully.
	EventLoop() int
}

// editor is the global structure that holds everything together. It implements
// the Editor interface.
type editor struct {
	wicore.EventRegistry
	quit          chan int
	terminal      Terminal                      // Abstract terminal interface to the real terminal.
	rootWindow    *window                       // The rootWindow is always DockingFill and set to the size of the terminal.
	lastActive    []wicore.Window               // Most recently used order of Window activatd.
	documents     []wicore.Document             // All loaded documents.
	viewFactories map[string]wicore.ViewFactory // All the ViewFactory's that can be used to create new View.
	viewReady     chan bool                     // A View.Buffer() is ready to be drawn.
	keyboardMode  wicore.KeyboardMode           // Global keyboard mode instead of per Window, it's more logical for users.
	plugins       Plugins                       // All loaded plugin processes.
	nextViewID    int
}

func (e *editor) Close() error {
	if e.plugins == nil {
		return nil
	}
	err := e.plugins.Close()
	e.plugins = nil
	return err
}

func (e *editor) ID() string {
	// There shall be only one.
	return "editor"
}

func (e *editor) Version() string {
	return version
}

func (e *editor) onTerminalMetaKeyPressed(k key.Press) {
	if !k.IsValid() {
		panic("Unexpected non-key")
	}
	if !k.IsMeta() {
		panic("Unexpected non-meta")
	}
	cmdName := wicore.GetKeyBindingCommand(e, e.KeyboardMode(), k)
	if cmdName != "" {
		// The command is executed inline, since the key was already enqueued in
		// the event queue.
		e.ExecuteCommand(e.ActiveWindow(), cmdName)
	} else {
		e.ExecuteCommand(e.ActiveWindow(), "alert", notMapped.Formatf(k))
	}
}

func (e *editor) onTerminalKeyPressed(k key.Press) {
	if !k.IsValid() {
		panic("Unexpected non-key")
	}
	if k.IsMeta() {
		panic("Unexpected meta")
	}
}

func (e *editor) ExecuteCommand(w wicore.Window, cmdName string, args ...string) {
	log.Printf("ExecuteCommand(%s, %s, %s)", w, cmdName, args)
	if w == nil {
		w = e.ActiveWindow()
	}
	cmd := wicore.GetCommand(e, w, cmdName)
	if cmd == nil {
		e.ExecuteCommand(w, "alert", notFound.Formatf(cmdName))
	} else {
		cmd.Handle(e, w, args...)
	}
}

func (e *editor) onCommands(cmds wicore.EnqueuedCommands) {
	for _, cmd := range cmds.Commands {
		e.ExecuteCommand(e.ActiveWindow(), cmd[0], cmd[1:]...)
	}
	if cmds.Callback != nil {
		cmds.Callback()
	}
}

func (e *editor) KeyboardMode() wicore.KeyboardMode {
	return e.keyboardMode
}

// draw descends the whole Window tree and redraw Windows.
func (e *editor) draw() {
	log.Print("draw()")
	// TODO(maruel): Cache the buffer.
	w, h := e.terminal.Size()
	out := raster.NewBuffer(w, h)
	drawRecurse(e.rootWindow, 0, 0, out)
	e.terminal.Blit(out)
}

func (e *editor) AllDocuments() []wicore.Document {
	out := make([]wicore.Document, len(e.documents))
	for i, v := range e.documents {
		out[i] = v
	}
	return out
}

func (e *editor) AllPlugins() []wicore.PluginDetails {
	out := make([]wicore.PluginDetails, len(e.plugins))
	for i, v := range e.plugins {
		out[i] = v.Details()
	}
	return out
}

func (e *editor) ActiveWindow() wicore.Window {
	return e.lastActive[0]
}

func (e *editor) activateWindow(w wicore.Window) {
	view := w.View()
	log.Printf("ActivateWindow(%s)", view.Title())
	if view.IsDisabled() {
		e.ExecuteCommand(w, "alert", activateDisabled.String())
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
	e.TriggerViewActivated(view)
}

func (e *editor) RegisterViewFactory(name string, viewFactory wicore.ViewFactory) bool {
	_, present := e.viewFactories[name]
	e.viewFactories[name] = viewFactory
	return !present
}

func (e *editor) ViewFactoryNames() []string {
	names := make([]string, 0, len(e.viewFactories))
	for name := range e.viewFactories {
		names = append(names, name)
	}
	return names
}

func (e *editor) onTerminalResized() {
	// Resize the Windows. This also invalidates it, which will also force a
	// redraw if the size changed.
	w, h := e.terminal.Size()
	e.rootWindow.setRect(raster.Rect{0, 0, w, h})
}

func (e *editor) onDocumentCursorMoved(doc wicore.Document, col, line int) {
	// TODO(maruel): Obviously wrong.
	e.terminal.SetCursor(col, line)
}

func (e *editor) terminalLoop(terminal Terminal) {
	for event := range terminal.SeedEvents() {
		switch event.Type {
		case EventKey:
			if event.Key.IsValid() {
				if event.Key.IsMeta() {
					e.TriggerTerminalMetaKeyPressed(event.Key)
				} else {
					e.TriggerTerminalKeyPressed(event.Key)
				}
			}
		case EventResize:
			e.TriggerTerminalResized()
		}
	}
}

// EventLoop handles both commands and events from the editor. This function
// runs in the UI goroutine.
func (e *editor) EventLoop() int {
	fakeChan := make(chan time.Time)
	var drawTimer <-chan time.Time = fakeChan
	for {
		select {
		case i := <-e.quit:
			return i

		case <-e.viewReady:
			// Taking in account a 60hz frame is 18.8ms, 5ms is going to be generally
			// processed within the same frame. This delaying results in significant
			// bandwidth saving on loading.
			if drawTimer == fakeChan {
				drawTimer = time.After(5 * time.Millisecond)
			}

		case <-drawTimer:
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

func (e *editor) isDirty() bool {
	for _, doc := range e.documents {
		if doc.IsDirty() {
			return true
		}
	}
	return false
}

func (e *editor) loadPlugins() {
	paths, err := enumPlugins(getPluginsPaths())
	if err != nil {
		log.Printf("Failed to enum plugins: %s", err)
	} else {
		e.plugins, err = loadPlugins(paths)
		// Failing to load plugins is not a hard error.
		log.Printf("Loaded %d plugins", len(e.plugins))
		if err != nil {
			log.Printf("Failed to load plugins: %s", err)
		}
	}
	// Trigger the RPC to initialize each plugin concurrently. Init() does not
	// wait for the plugin to be fully initialized.
	for _, plugin := range e.plugins {
		plugin.Init(e)
	}
}

// MakeEditor creates an object that implements the Editor interface. The root
// window doesn't have anything to view in it.
//
// The editor contains a root window and a root view. It's up to the caller to
// add child Windows in it. Normally it will be done via the command
// "editor_bootstrap_ui" to add the status bar, then "new" or "open" to create
// the initial text buffer.
//
// It is fine to run it concurrently in unit test, as no global variable shall
// be used by the object created by this function.
func MakeEditor(terminal Terminal, noPlugin bool) (Editor, error) {
	lang.Set(lang.En)
	reg := makeEventRegistry()
	e := &editor{
		EventRegistry: reg,
		quit:          make(chan int),
		terminal:      terminal,
		rootWindow:    nil,                         // It is set below due to circular reference.
		lastActive:    make([]wicore.Window, 1, 8), // It is set below.
		documents:     []wicore.Document{},
		viewFactories: make(map[string]wicore.ViewFactory),
		viewReady:     make(chan bool),
		keyboardMode:  wicore.Normal,
		nextViewID:    1,
	}

	// The root view is important, it defines all the global commands. It is
	// pre-filled with the default native commands and keyboard mapping, and it's
	// up to the plugins to add more global commands on startup.
	rootView := makeStaticDisabledView(e, 0, "Root", -1, -1)

	// These commands are generic commands, they do not require specific access.
	cmds := rootView.CommandsW()
	RegisterCommandCommands(cmds)
	RegisterKeyBindingCommands(cmds)
	RegisterViewCommands(cmds)
	RegisterWindowCommands(cmds)
	RegisterDocumentCommands(cmds)
	RegisterEditorDefaults(rootView)

	RegisterDefaultViewFactories(e)

	e.rootWindow = makeWindow(nil, rootView, wicore.DockingFill)
	e.rootWindow.e = e
	e.lastActive[0] = e.rootWindow

	e.RegisterTerminalMetaKeyPressed(e.onTerminalMetaKeyPressed)
	e.RegisterTerminalKeyPressed(e.onTerminalKeyPressed)
	e.RegisterTerminalResized(e.onTerminalResized)
	e.RegisterCommands(e.onCommands)
	e.RegisterDocumentCursorMoved(e.onDocumentCursorMoved)

	if !noPlugin {
		e.loadPlugins()
	}

	e.TriggerWindowCreated(e.rootWindow)
	e.TriggerViewCreated(rootView)
	// This forces creating the default buffer.
	e.TriggerTerminalResized()
	wicore.Go("terminalLoop", func() { e.terminalLoop(terminal) })
	//e.TriggerEditorLanguage(lang.Active())
	//e.TriggerEditorKeyboardModeChanged(e.keyboardMode)
	return e, nil
}

// Commands

func cmdAlert(c *wicore.CommandImpl, e wicore.EditorW, w wicore.Window, args ...string) {
	e.ExecuteCommand(w, "window_new", "0", "bottom", "infobar_alert", args[0])
}

func cmdEditorBootstrapUI(c *wicore.CommandImpl, e wicore.EditorW, w wicore.Window, args ...string) {
	e.ExecuteCommand(w, "window_new", "0", "bottom", "status_root")
}

func cmdEditorCommandWindow(c *wicore.CommandImpl, e wicore.EditorW, w wicore.Window, args ...string) {
	// Create the Window with the command view and attach it to the currently
	// focused Window.
	e.ExecuteCommand(w, "window_new", w.ID(), "floating", "command")
}

func cmdEditorQuit(c *privilegedCommandImpl, e *editor, w *window, args ...string) {
	if len(args) >= 1 {
		e.ExecuteCommand(w, "alert", c.LongDesc())
		return
	} else if len(args) == 1 {
		if args[0] != "force" {
			e.ExecuteCommand(w, "alert", c.LongDesc())
			return
		}
	} else {
		if e.isDirty() {
			// TODO(maruel): For each dirty Document, "prompt" y/n to force quit. If
			// 'n', stop there.
			return
		}
		// TODO(maruel):
		// - Send a signal to each plugin.
		// - Send a signal back to the main loop.
	}

	// This tells the editor.EventLoop() to quit. This is synchronous.
	e.quit <- 0
}

func cmdEditorRedraw(c *privilegedCommandImpl, e *editor, w *window, args ...string) {
	wicore.Go("viewReady", func() {
		e.viewReady <- true
	})
}

// RegisterEditorDefaults registers the top-level native commands and key
// bindings.
func RegisterEditorDefaults(view wicore.ViewW) {
	cmds := []wicore.Command{
		&wicore.CommandImpl{
			"alert",
			1,
			cmdAlert,
			wicore.WindowCategory,
			lang.Map{
				lang.En: "Shows a modal message",
			},
			lang.Map{
				lang.En: "Prints a message in a modal dialog box.",
			},
		},
		&wicore.CommandImpl{
			"editor_bootstrap_ui",
			0,
			cmdEditorBootstrapUI,
			wicore.WindowCategory,
			lang.Map{
				lang.En: "Bootstraps the editor's UI",
			},
			lang.Map{
				lang.En: "Bootstraps the editor's UI. This command is automatically run on startup and cannot be executed afterward. It adds the standard status bar. This command exists so it can be overriden by a plugin, so it can create its own status bar.",
			},
		},
		&wicore.CommandImpl{
			"editor_command_window",
			0,
			cmdEditorCommandWindow,
			wicore.CommandsCategory,
			lang.Map{
				lang.En: "Shows the interactive command window",
			},
			lang.Map{
				lang.En: "This commands exists so it can be bound to a key to pop up the interactive command window.",
			},
		},
		&privilegedCommandImpl{
			"editor_quit",
			-1,
			cmdEditorQuit,
			wicore.EditorCategory,
			lang.Map{
				lang.En: "Quits",
			},
			lang.Map{
				lang.En: "Quits the editor. Use 'force' to bypasses writing the files to disk.",
			},
		},
		&privilegedCommandImpl{
			"editor_redraw",
			0,
			cmdEditorRedraw,
			wicore.EditorCategory,
			lang.Map{
				lang.En: "Forcibly redraws the terminal",
			},
			lang.Map{
				lang.En: "Forcibly redraws the terminal.",
			},
		},
		&wicore.CommandAlias{"q", "editor_quit", nil},
		&wicore.CommandAlias{"q!", "editor_quit", []string{"force"}},
		&wicore.CommandAlias{"quit", "editor_quit", nil},
	}
	commands := view.CommandsW()
	for _, cmd := range cmds {
		commands.Register(cmd)
	}

	bindings := view.KeyBindingsW()
	bindings.Set(wicore.AllMode, key.Press{Key: key.F1}, "help")
	bindings.Set(wicore.AllMode, key.Press{Ch: ':'}, "editor_command_window")
	bindings.Set(wicore.AllMode, key.Press{Ctrl: true, Ch: 'c'}, "quit")
	bindings.Set(wicore.Insert, key.Press{Key: key.Escape}, "key_set_normal")
}
