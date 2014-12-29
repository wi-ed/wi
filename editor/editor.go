// Copyright 2013 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package editor

import (
	"fmt"
	"io"
	"log"
	"sync/atomic"
	"time"

	"github.com/maruel/wi/wiCore"
)

const (
	// Major.Minor.Bugfix. All plugins should be recompiled with wiCore
	// changes.
	version = "0.0.1"
)

// commandItem is a command pending to be executed.
type commandItem struct {
	cmdName string          // Set on command execution
	args    []string        // Set on command execution
	key     wiCore.KeyPress // Set on key press
}

// commandQueueItem is a set of commandItem pending to be executed.
type commandQueueItem []commandItem

// Editor is the inprocess wiCore.Editor interface. It adds the process life-time
// management functions to the public interface wiCore.Editor.
//
// It is very important to call the Close() function upon termination.
type Editor interface {
	io.Closer

	wiCore.Editor

	// EventLoop runs the event loop until the command "quit" executes
	// successfully.
	EventLoop() int
}

// editor is the global structure that holds everything together. It implements
// the Editor interface.
type editor struct {
	terminal       Terminal                      // Abstract terminal interface to the real terminal.
	rootWindow     *window                       // The rootWindow is always DockingFill and set to the size of the terminal.
	lastActive     []wiCore.Window               // Most recently used order of Window activatd.
	documents      []wiCore.Document             // All loaded documents.
	viewFactories  map[string]wiCore.ViewFactory // All the ViewFactory's that can be used to create new View.
	lastCommandID  int64                         // Used by PostCommand.
	terminalEvents <-chan TerminalEvent          // Events coming from Terminal.SeedEvents().
	viewReady      chan bool                     // A View.Buffer() is ready to be drawn.
	commandsQueue  chan commandQueueItem         // Pending commands to be executed.
	languageMode   wiCore.LanguageMode           // Actual language used.
	keyboardMode   wiCore.KeyboardMode           // Global keyboard mode is either CommandMode or EditMode.
	plugins        Plugins                       // All loaded plugin processes.
	quitFlag       bool                          // If true, a shutdown is in progress.
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

func (e *editor) PostCommands(cmds [][]string) wiCore.CommandID {
	log.Printf("PostCommands(%s)", cmds)
	tmp := make(commandQueueItem, len(cmds))
	for i, cmd := range cmds {
		tmp[i].cmdName = cmd[0]
		tmp[i].args = cmd[1:]
	}
	e.commandsQueue <- tmp
	return wiCore.CommandID{0, int(atomic.AddInt64(&e.lastCommandID, 1))}
}

func (e *editor) postKey(key wiCore.KeyPress) {
	log.Printf("PostKey(%s)", key)
	e.commandsQueue <- commandQueueItem{commandItem{key: key}}
}

func (e *editor) ExecuteCommand(w wiCore.Window, cmdName string, args ...string) {
	log.Printf("ExecuteCommand(%s, %s, %s)", w, cmdName, args)
	if w == nil {
		w = e.ActiveWindow()
	}
	cmd := wiCore.GetCommand(e, w, cmdName)
	if cmd == nil {
		e.ExecuteCommand(w, "alert", fmt.Sprintf(wiCore.GetStr(e.CurrentLanguage(), notFound), cmdName))
	} else {
		cmd.Handle(e, w, args...)
	}
}

func (e *editor) CurrentLanguage() wiCore.LanguageMode {
	return e.languageMode
}

func (e *editor) KeyboardMode() wiCore.KeyboardMode {
	return e.keyboardMode
}

// draw descends the whole Window tree and redraw Windows.
func (e *editor) draw() {
	log.Print("draw()")
	// TODO(maruel): Cache the buffer.
	w, h := e.terminal.Size()
	out := wiCore.NewBuffer(w, h)
	drawRecurse(e.rootWindow, 0, 0, out)
	e.terminal.Blit(out)
}

func (e *editor) AllDocuments() []wiCore.Document {
	return e.documents[:]
}

func (e *editor) ActiveWindow() wiCore.Window {
	return e.lastActive[0]
}

func (e *editor) activateWindow(w wiCore.Window) {
	log.Printf("ActivateWindow(%s)", w.View().Title())
	if w.View().IsDisabled() {
		e.ExecuteCommand(w, "alert", wiCore.GetStr(e.CurrentLanguage(), activateDisabled))
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

func (e *editor) RegisterViewFactory(name string, viewFactory wiCore.ViewFactory) bool {
	_, present := e.viewFactories[name]
	e.viewFactories[name] = viewFactory
	return !present
}

func (e *editor) onResize() {
	// Resize the Windows. This also invalidates it, which will also force a
	// redraw if the size changed.
	w, h := e.terminal.Size()
	e.rootWindow.setRect(wiCore.Rect{0, 0, w, h})
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
							cmdName := wiCore.GetKeyBindingCommand(e, e.KeyboardMode(), keyName)
							if cmdName != "" {
								e.ExecuteCommand(e.ActiveWindow(), cmdName)
							} else {
								e.ExecuteCommand(e.ActiveWindow(), "alert", fmt.Sprintf(wiCore.GetStr(e.CurrentLanguage(), notMapped), keyName))
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
			// TODO(maruel): Temporary for unit tests. Figure out a way to enforce
			// drawing in test cases but not on real process exit, since it's useless.
			//if e.quitFlag {
			//	return 0
			//}

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
			if e.quitFlag {
				return 0
			}
		}
	}
}

func (e *editor) loadPlugins() {
	// TODO(maruel): Get path.
	paths, err := EnumPlugins(".")
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
	RegisterDebugCommands(rootView.Commands())
	RegisterCommandCommands(rootView.Commands())
	RegisterKeyBindingCommands(rootView.Commands())
	RegisterViewCommands(rootView.Commands())
	RegisterWindowCommands(rootView.Commands())
	RegisterDocumentCommands(rootView.Commands())
	RegisterEditorCommands(rootView.Commands())

	rootWindow := makeWindow(nil, rootView, wiCore.DockingFill)
	e := &editor{
		terminal:       terminal,
		rootWindow:     rootWindow,
		lastActive:     []wiCore.Window{rootWindow},
		documents:      []wiCore.Document{},
		viewFactories:  make(map[string]wiCore.ViewFactory),
		terminalEvents: terminal.SeedEvents(),
		viewReady:      make(chan bool),
		commandsQueue:  make(chan commandQueueItem, 500),
		languageMode:   wiCore.LangEn,
		keyboardMode:   wiCore.EditMode,
	}
	rootWindow.cd = e

	RegisterDefaultViewFactories(e)

	// This forces creating the default buffer.
	e.onResize()

	if !noPlugin {
		e.loadPlugins()
	}

	// Key bindings are loaded after the plugins, so a plugin has the chance to
	// hook the command 'key_bind' if desired. It's also the perfect time to hook
	// 'editor_bootstrap_ui' to customize the default look on startup.
	RegisterDefaultKeyBindings(e)
	return e, nil
}

// Commands

func cmdAlert(c *wiCore.CommandImpl, cd wiCore.CommandDispatcherFull, w wiCore.Window, args ...string) {
	cd.ExecuteCommand(w, "window_new", "0", "bottom", "infobar_alert", args[0])
}

func cmdEditorBootstrapUI(c *wiCore.CommandImpl, cd wiCore.CommandDispatcherFull, w wiCore.Window, args ...string) {
	cd.ExecuteCommand(w, "window_new", "0", "bottom", "status_root")
}

func (e *editor) isDirty() bool {
	for _, doc := range e.documents {
		if doc.IsDirty() {
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
		if e.isDirty() {
			// TODO(maruel): For each dirty Document, "prompt" y/n to force quit. If
			// 'n', stop there.
			return
		}
		// TODO(maruel):
		// - Send a signal to each plugin.
		// - Send a signal back to the main loop.
	}

	e.quitFlag = true
	// editor_redraw wakes up the command event loop so it detects it's time to
	// quit.
	wiCore.PostCommand(e, "editor_redraw")
}

func cmdEditorRedraw(c *privilegedCommandImpl, e *editor, w *window, args ...string) {
	go func() {
		e.viewReady <- true
	}()
}

func cmdShowCommandWindow(c *wiCore.CommandImpl, cd wiCore.CommandDispatcherFull, w wiCore.Window, args ...string) {
	// Create the Window with the command view and attach it to the currently
	// focused Window.
	cd.ExecuteCommand(w, "window_new", w.ID(), "floating", "command")
}

// RegisterEditorCommands registers the top-level native commands.
func RegisterEditorCommands(dispatcher wiCore.Commands) {
	cmds := []wiCore.Command{
		&wiCore.CommandImpl{
			"alert",
			1,
			cmdAlert,
			wiCore.WindowCategory,
			wiCore.LangMap{
				wiCore.LangEn: "Shows a modal message",
			},
			wiCore.LangMap{
				wiCore.LangEn: "Prints a message in a modal dialog box.",
			},
		},
		&wiCore.CommandImpl{
			"editor_bootstrap_ui",
			0,
			cmdEditorBootstrapUI,
			wiCore.WindowCategory,
			wiCore.LangMap{
				wiCore.LangEn: "Bootstraps the editor's UI",
			},
			wiCore.LangMap{
				wiCore.LangEn: "Bootstraps the editor's UI. This command is automatically run on startup and cannot be executed afterward. It adds the standard status bar. This command exists so it can be overriden by a plugin, so it can create its own status bar.",
			},
		},
		&privilegedCommandImpl{
			"editor_quit",
			-1,
			cmdEditorQuit,
			wiCore.EditorCategory,
			wiCore.LangMap{
				wiCore.LangEn: "Quits",
			},
			wiCore.LangMap{
				wiCore.LangEn: "Quits the editor. Use 'force' to bypasses writing the files to disk.",
			},
		},
		&privilegedCommandImpl{
			"editor_redraw",
			0,
			cmdEditorRedraw,
			wiCore.EditorCategory,
			wiCore.LangMap{
				wiCore.LangEn: "Forcibly redraws the terminal",
			},
			wiCore.LangMap{
				wiCore.LangEn: "Forcibly redraws the terminal.",
			},
		},
		&wiCore.CommandImpl{
			"show_command_window",
			0,
			cmdShowCommandWindow,
			wiCore.CommandsCategory,
			wiCore.LangMap{
				wiCore.LangEn: "Shows the interactive command window",
			},
			wiCore.LangMap{
				wiCore.LangEn: "This commands exists so it can be bound to a key to pop up the interactive command window.",
			},
		},
		&wiCore.CommandAlias{"q", "editor_quit", nil},
		&wiCore.CommandAlias{"q!", "editor_quit", []string{"force"}},
		&wiCore.CommandAlias{"quit", "editor_quit", nil},
	}
	for _, cmd := range cmds {
		dispatcher.Register(cmd)
	}
}
