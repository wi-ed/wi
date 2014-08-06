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
	/*for _, window := range t.window.ChildrenWindows() {
	}*/
	termbox.Flush()
}

func (t *terminal) ActiveWindow() wi.Window {
	return t.lastActive[0]
}

func (t *terminal) ActivateWindow(w wi.Window) {
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
			cmdName := wi.GetKeyBindingCommand(t, keyEventToName(event))
			if cmdName != "" {
				cmd := wi.GetCommand(t, cmdName)
				if cmd != nil {
					cmd.Handle(t.ActiveWindow())
				}
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

// makeDisplay creates the Display object. The root window doesn't have
// anything to view in it. It will contain two child windows, the main content
// window and the status bar.
func makeDisplay() *terminal {
	// The root view is important, it defines all the global commands. It is
	// pre-filled with the default native commands and keyboard mapping, and it's
	// up to the plugins to add more global commands on startup.
	rootView := makeView(-1, -1)
	RegisterDefaultCommands(rootView.Commands())
	RegisterDefaultKeyBindings(rootView.KeyBindings())

	// TODO(maruel): Add callback to allow plugins to hook into it
	// (a)synchronously before creating the root Window.
	window := makeWindow(nil, rootView, wi.DockingFill)
	events := make(chan termbox.Event, 32)
	terminal := &terminal{
		events:     events,
		window:     window,
		lastActive: []wi.Window{window},
	}

	terminal.onResize()
	go func() {
		for {
			events <- termbox.PollEvent()
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

type view struct {
	commands    wi.Commands
	keyBindings wi.KeyBindings
	title       string
	isDirty     bool
	isInvalid   bool
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

func (v *view) DrawInto(buffer tulib.Buffer) {
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

// The status line.
func makeStatusView() wi.View {
	return makeView(1, -1)
}

// The command box.
func makeCommandView() wi.View {
	return makeView(1, -1)
}

// A dismissable modal dialog box.
func makeAlertView() wi.View {
	return makeView(1, 1)
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

	display := makeDisplay()
	plugins := loadPlugins(display)
	defer func() {
		for _, p := range plugins {
			// TODO(maruel): Nicely terminate them.
			p.Kill()
		}
	}()

	// Add the status bar.
	wi.ExecuteCommand(display, "add_status_bar")

	if *command {
		for _, i := range flag.Args() {
			wi.ExecuteCommand(display, i)
		}
	} else if flag.NArg() > 0 {
		for _, i := range flag.Args() {
			wi.ExecuteCommand(display, "open", i)
		}
	} else {
		// If nothing, opens a blank editor.
		wi.ExecuteCommand(display, "new")
	}

	// Run the message loop.
	out := display.eventLoop()

	// Normal exit.
	termbox.SetCursor(0, 0)
	termbox.Flush()

	os.Exit(out)
}
