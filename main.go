// Copyright 2013 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"github.com/maruel/wi/wi-plugin"
	"github.com/nsf/termbox-go"
	"github.com/nsf/tulib"
	"io/ioutil"
	"log"
	"net/rpc"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

const (
	// Major.Minor.Bugfix. All plugins should be recompiled on Minor version
	// change.
	VERSION = "0.0.1"
)

// UI

// It is normally expected to be drawn via an ssh/mosh connection so it should
// be "bandwidth" optimized, where bandwidth doesn't mean 1200 bauds anymore.
type terminal struct {
	width        int
	height       int
	window       wi.Window
	lastActive   []wi.Window
	events       <-chan termbox.Event
	outputBuffer tulib.Buffer
}

func (t *terminal) Draw() {
	// Descend the whole Window tree and find the invalidated window to redraw.
	// TODO(maruel): Optimize: If a floating window is invalidated, redraw all
	// visible windows.
	// Do a depth first search.
	for _, window := range t.window.ChildrenWindows() {
		if window.IsInvalid() {
		}
	}
	termbox.Flush()
}

func (t *terminal) ActiveWindow() wi.Window {
	return t.lastActive[0]
}

func (t *terminal) Height() int {
	return t.height
}

func (t *terminal) Width() int {
	return t.width
}

func (t *terminal) onResize() {
	// Recreate the buffer, which queries the new sizes.
	t.outputBuffer = tulib.TermboxBuffer()
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	// Resize the Windows.
	t.window.SetRect(tulib.Rect{0, 0, t.Width(), t.Height()})
}

func (t *terminal) eventLoop() int {
	for {
		event := <-t.events
		switch event.Type {
		case termbox.EventKey:
			active := t.ActiveWindow()
			cmd := active.Keyboard().OnKey(event)
			if cmd != "" {
				active.Command().Execute(active, cmd)
			}
		case termbox.EventMouse:
			// TODO(maruel): MouseDispatcher.
			break
		case termbox.EventResize:
			t.onResize()
		case termbox.EventError:
			// TODO(maruel): Not sure what situations can trigger this.
			os.Stderr.WriteString(event.Err.Error())
			return 1
		}
		t.Draw()
	}
	return 0
}

// The root window doesn't have anything to view in it. It will contain two
// child windows, the main content window and the status bar.
func MakeDisplay() *terminal {
	cmd_dispatcher := MakeCommandDispatcher()
	RegisterDefaultCommands(cmd_dispatcher)
	key_dispatcher := MakeKeyboardDispatcher()
	RegisterDefaultKeyboard(key_dispatcher)
	window := makeWindow(cmd_dispatcher, key_dispatcher, nil, makeBlankView(), wi.Fill)
	events := make(chan termbox.Event, 32)
	terminal := &terminal{
		events:     events,
		window:     window,
		lastActive: []wi.Window{window},
	}
	window.NewChildWindow(makeStatusView(), wi.Bottom)
	window.NewChildWindow(makeBlankView(), wi.Fill)
	terminal.onResize()
	go func() {
		for {
			events <- termbox.PollEvent()
		}
	}()
	return terminal
}

type window struct {
	cmd_dispatcher  wi.CommandDispatcher
	key_dispatcher  wi.KeyboardDispatcher
	parent          wi.Window
	rect            tulib.Rect
	childrenWindows []wi.Window
	view            wi.View
	docking         wi.DockingType
	border          wi.BorderType
	isInvalid       bool
}

func (w *window) Command() wi.CommandDispatcher {
	return w.cmd_dispatcher
}

func (w *window) Keyboard() wi.KeyboardDispatcher {
	return w.key_dispatcher
}

func (w *window) Parent() wi.Window {
	return w.parent
}

func (w *window) ChildrenWindows() []wi.Window {
	return w.childrenWindows[:]
}

func (w *window) NewChildWindow(view wi.View, docking wi.DockingType) {
	w.childrenWindows = append(w.childrenWindows, makeWindow(&commandDispatcher{}, &keyboardDispatcher{}, w, view, docking))
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
	w.Invalidate()
}

func (w *window) IsInvalid() bool {
	return w.isInvalid
}

func (w *window) Invalidate() {
	w.isInvalid = true
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

func makeWindow(cmd_dispatcher wi.CommandDispatcher, key_dispatcher wi.KeyboardDispatcher, parent wi.Window, view wi.View, docking wi.DockingType) wi.Window {
	return &window{
		cmd_dispatcher: cmd_dispatcher,
		key_dispatcher: key_dispatcher,
		parent:         parent,
		view:           view,
		docking:        docking,
	}
}

type view struct {
	naturalX int
	naturalY int
	buffer   wi.TextBuffer
}

func (v *view) SetBuffer(buffer wi.TextBuffer) {
	v.buffer = buffer
}

func (v *view) Buffer() wi.TextBuffer {
	return v.buffer
}

func (v *view) NaturalSize() (x, y int) {
	return v.naturalX, v.naturalY
}

// Empty non-editable window.
func makeBlankView() wi.View {
	return &view{0, 0, nil}
}

// The status line.
func makeStatusView() wi.View {
	return &view{1, -1, nil}
}

// The command box.
func makeCommandView() wi.View {
	return &view{1, -1, nil}
}

// A dismissable modal dialog box.
func makeAlertView() wi.View {
	return &view{1, -1, nil}
}

// Config

type config struct {
	ints    map[string]int
	strings map[string]string
}

func (c *config) GetInt(name string) int {
	return c.ints[name]
}

func (c *config) GetString(name string) string {
	return c.strings[name]
}

func (c *config) Save() {
}

func MakeConfig() wi.Config {
	return &config{}
}

// Control

type command struct {
	handler   wi.CommandHandler
	shortDesc string
	longDesc  string
}

func (c *command) Handle(w wi.Window, args ...string) {
	c.handler(w, args...)
}

func (c *command) ShortDesc() string {
	return c.shortDesc
}

func (c *command) LongDesc() string {
	return c.longDesc
}

type commandDispatcher struct {
	commands map[string]wi.Command
}

func (c *commandDispatcher) Execute(w wi.Window, cmd string, args ...string) {
	v, _ := c.commands[cmd]
	if v == nil {
		parent := w.Parent()
		if parent != nil {
			parent.Command().Execute(parent, cmd, args...)
		} else {
			// This is the root command, surface the error.
			c.Execute(w, "alert", "Command \""+cmd+"\" is not registered")
		}
	} else {
		v.Handle(w, args...)
	}
}

func (c *commandDispatcher) Register(name string, cmd wi.Command) bool {
	_, ok := c.commands[name]
	c.commands[name] = cmd
	return !ok
}

func MakeCommandDispatcher() wi.CommandDispatcher {
	return &commandDispatcher{make(map[string]wi.Command)}
}

type keyboardDispatcher struct {
	mappings map[string]string
}

func (k *keyboardDispatcher) OnKey(event termbox.Event) string {
	key := tulib.KeyToString(event.Key, event.Ch, event.Mod)
	return k.mappings[key]
}

func (k *keyboardDispatcher) Register(key string, cmd string) bool {
	_, ok := k.mappings[key]
	k.mappings[key] = cmd
	return !ok
}

func MakeKeyboardDispatcher() wi.KeyboardDispatcher {
	return &keyboardDispatcher{make(map[string]string)}
}

// Registers the native commands.
func RegisterDefaultCommands(dispatcher wi.CommandDispatcher) {
	dispatcher.Register(
		"alert",
		&command{
			func(w wi.Window, args ...string) {
				// TODO: w.Root().NewChildWindow(makeDialog(root))
				log.Printf("Faking an alert: %s", args)
			},
			"Shows a modal message",
			"Prints a message in a modal dialog box.",
		})
	dispatcher.Register(
		"open",
		&command{
			func(w wi.Window, args ...string) {
				log.Printf("Faking opening a file: %s", args)
			},
			"Opens a file in a new buffer",
			"Opens a file in a new buffer.",
		})
	dispatcher.Register(
		"new",
		&command{
			func(w wi.Window, args ...string) {
				log.Printf("Faking opening a new buffer: %s", args)
			},
			"Create a new buffer",
			"Create a new buffer.",
		})
	dispatcher.Register(
		"shell",
		&command{
			func(w wi.Window, args ...string) {
				log.Printf("Faking opening a new shell: %s", args)
			},
			"Opens a shell process",
			"Opens a shell process in a new buffer.",
		})
	dispatcher.Register("doc",
		&command{
			func(w wi.Window, args ...string) {
				// TODO: MakeWindow(Bottom)
				docArgs := make([]string, len(args)+1)
				docArgs[0] = "doc"
				copy(docArgs[1:], args)
				dispatcher.Execute(w, "shell", docArgs...)
			},
			"Search godoc documentation",
			"Uses the 'doc' tool to get documentation about the text under the cursor.",
		})
	dispatcher.Register(
		"quit",
		&command{
			func(w wi.Window, args ...string) {
				// For all the View, question if fine to quit.
				// If not fine, "prompt" y/n to force quit. If n, stop there.
				// - Send a signal to each plugin.
				// - Send a signal back to the main loop.
				log.Printf("Faking quit: %s", args)
			},
			"Quits",
			"Quits the editor. Optionally bypasses writing the files to disk.",
		})
	dispatcher.Register("help",
		&command{
			func(w wi.Window, args ...string) {
				// Creates a new Window with a ViewHelp.
				log.Printf("Faking help: %s", args)
			},
			"Prints help",
			"Prints general help or help for a particular command.",
		})
	// DIRECTION = up/down/left/right
	// window_DIRECTION
	// window_close
	// cursor_move_DIRECTION
	// add_text/insert/delete
	// undo/redo
	// verb/movement/multiplier
	// Modes, select (both column and normal), command.
}

// Registers the default keyboard mapping. Keyboard mapping simply execute the
// corresponding command. So to add a keyboard map, the corresponding command
// needs to be added first.
// TODO(maruel): This should be remappable via a configuration flag, for
// example vim flavor vs emacs flavor.
func RegisterDefaultKeyboard(key_dispatcher wi.KeyboardDispatcher) {
	key_dispatcher.Register("F1", "help")
	key_dispatcher.Register("Ctrl-C", "quit")
}

// Plugins

// loadPlugin starts a plugin and returns the process.
func loadPlugin(server *rpc.Server, f string) *os.Process {
	log.Printf("Would have run plugin %s", f)
	cmd := exec.Command(f)
	cmd.Env = append(os.Environ(), "WI=plugin")

	stdin, err := cmd.StdinPipe()
	if err != nil {
		panic(err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		panic(err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		panic(err)
	}
	if err := cmd.Start(); err != nil {
		// Surface the error as an "alert", since it's not a fatal error.
		log.Fatal(err)
	}

	// Fail on any write to Stderr.
	go func() {
		buf := make([]byte, 2048)
		n, _ := stderr.Read(buf)
		if n != 0 {
			panic(fmt.Sprintf("Plugin %s failed: %s", f, buf))
		}
	}()

	// Before starting the RPC, ensures the version matches.
	expectedVersion := wi.CalculateVersion()
	b := make([]byte, 40)
	n, err := stdout.Read(b)
	if err != nil {
		// Surface the error as an "alert", since it's not a fatal error.
		log.Fatal(err)
	}
	if n != 40 {
		// Surface the error as an "alert", since it's not a fatal error.
		log.Fatal("Unexpected size")
	}
	actualVersion := string(b)
	if expectedVersion != actualVersion {
		// Surface the error as an "alert", since it's not a fatal error.
		log.Fatalf("For %s; expected %s, got %s", f, expectedVersion, actualVersion)
	}

	// Start the RPC server for this plugin.
	go func() {
		server.ServeConn(wi.MakeReadWriteCloser(stdout, stdin))
	}()

	return cmd.Process
}

// loadPlugins loads all the plugins and returns the process handles.
func loadPlugins(display wi.Display) []*os.Process {
	// TODO(maruel): Get the path of the executable. It's a bit involved since
	// very OS specific but it's doable. Then all plugins in the same directory
	// are access.
	searchDir := "."
	files, err := ioutil.ReadDir(searchDir)
	if err != nil {
		return nil
	}
	if len(files) == 0 {
		// Save registering RPC stuff when unnecessary.
		return nil
	}

	out := []*os.Process{}
	server := rpc.NewServer()
	// TODO(maruel): http://golang.org/pkg/net/rpc/#Server.RegisterName
	// It should be an interface with methods of style DoStuff(Foo, Bar) Baz
	//server.RegisterName("Display", display)
	for _, f := range files {
		name := f.Name()
		if !strings.HasPrefix(name, "wi-plugin-") {
			continue
		}
		// Crude check for executable test.
		if runtime.GOOS == "windows" {
			if !strings.HasSuffix(name, ".exe") {
				continue
			}
		} else {
			if f.Mode()&0111 == 0 {
				continue
			}
		}
		out = append(out, loadPlugin(server, name))
	}
	return out
}

func main() {
	log.SetFlags(log.Lmicroseconds)
	command := flag.Bool("c", false, "Runs the commands specified on startup")
	version := flag.Bool("v", false, "Prints version and exit")
	flag.Parse()

	// Process this one early. No one wants version output to take 1s.
	if *version {
		println(VERSION)
		os.Exit(0)
	}

	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	defer termbox.Close()
	termbox.SetInputMode(termbox.InputAlt)

	display := MakeDisplay()

	plugins := loadPlugins(display)
	defer func() {
		for _, p := range plugins {
			// TODO(maruel): Nicely terminate them.
			p.Kill()
		}
	}()

	active := display.ActiveWindow()
	if *command {
		for _, i := range flag.Args() {
			active.Command().Execute(active, i)
		}
	} else if flag.NArg() > 0 {
		for _, i := range flag.Args() {
			active.Command().Execute(active, "open", i)
		}
	} else {
		// If nothing, opens a blank editor.
		active.Command().Execute(active, "new")
	}

	// Run the message loop.
	out := display.eventLoop()

	// Normal exit.
	termbox.SetCursor(0, 0)
	termbox.Flush()

	os.Exit(out)
}
