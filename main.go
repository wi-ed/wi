// Copyright 2013 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package main

import (
	//"bytes"
	"flag"
	//"fmt"
	"github.com/nsf/termbox-go"
	"github.com/nsf/tulib"
	"log"
	"os"
	//"os/exec"
	//"path/filepath"
	//"strconv"
)

const (
	// Major.Minor.Bugfix. All plugins should be recompiled on Minor version
	// change.
	VERSION = "0.0.1"
)

// UI

// Display is the output device. It shows the root window which covers the
// whole screen estate.
type Display interface {
	Draw()
	RootWindow() Window
	ActiveWindow() Window
	Height() int
	Width() int
	EventLoop() int
}

// Window is a View container. It defines the position, Z-ordering via
// hierarchy and decoration. It can have multiple child windows. The child
// windows are not bounded by the parent window.
type Window interface {
	// Each Window has its own keyboard dispatcher for Window specific commands,
	// for example the 'command' window has different behavior than a golang
	// editor window.
	KeyboardDispatcher

	// Rect returns the position based on the Display, not the parent Window.
	Rect() tulib.Rect
	Parent() Window
	NewChildWindow(view View)

	SetView(view View)
	View() View
}

// TextBuffer is the content. It may only contain partial information in the
// case of large file or file opened via high latency I/O.
type TextBuffer interface {
	Lines() int
}

// View is the presentation of a TextBuffer in a Window. It responds to user
// input.
type View interface {
	SetBuffer(buffer TextBuffer)
	Buffer() TextBuffer
}

// It is normally expected to be drawn via an ssh/mosh connection so it should
// be "bandwidth" optimized, where bandwidth doesn't mean 1200 bauds anymore.
type terminal struct {
	width          int
	height         int
	window         Window
	lastActive     []Window
	events         <-chan termbox.Event
	outputBuffer   tulib.Buffer
	key_dispatcher KeyboardDispatcher
}

func (t *terminal) Draw() {
	termbox.Flush()
}

func (t *terminal) RootWindow() Window {
	return t.window
}

func (t *terminal) ActiveWindow() Window {
	return t.lastActive[0]
}

func (t *terminal) Height() int {
	return t.height
}

func (t *terminal) Width() int {
	return t.width
}

func (t *terminal) onResize() {
	t.outputBuffer = tulib.TermboxBuffer()
	// Resize the Windows.
}

func (t *terminal) EventLoop() int {
	for {
		event := <-t.events
		switch event.Type {
		case termbox.EventKey:
			t.key_dispatcher.OnKey(event)
		case termbox.EventMouse:
			// TODO(maruel): MouseDispatcher.
			break
		case termbox.EventResize:
			//termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
			t.onResize()
		case termbox.EventError:
			os.Stderr.WriteString(event.Err.Error())
			return 1
		}
		t.Draw()
	}
	return 0
}

func MakeDisplay(key_dispatcher KeyboardDispatcher) Display {
	window := MakeWindow(nil, MakeView())
	events := make(chan termbox.Event, 32)
	terminal := &terminal{
		events:         events,
		window:         window,
		lastActive:     []Window{window},
		key_dispatcher: key_dispatcher,
	}
	terminal.onResize()
	go func() {
		for {
			events <- termbox.PollEvent()
		}
	}()
	return terminal
}

type window struct {
	KeyboardDispatcher
	parent          Window
	rect            tulib.Rect
	childrenWindows []Window
	view            View
}

func (w *window) Parent() Window {
	return w.parent
}

func (w *window) NewChildWindow(view View) {
	w.childrenWindows = append(w.childrenWindows, MakeWindow(w, view))
}

func (w *window) Rect() tulib.Rect {
	return w.rect
}

func (w *window) SetView(view View) {
	w.view = view
}

func (w *window) View() View {
	return w.view
}

func MakeWindow(parent Window, view View) Window {
	return &window{parent: parent, view: view}
}

type view struct {
	buffer TextBuffer
}

func (v *view) SetBuffer(buffer TextBuffer) {
	v.buffer = buffer
}

func (v *view) Buffer() TextBuffer {
	return v.buffer
}

func MakeView() View {
	return &view{}
}

// Config

// Configuration manager.
type Config interface {
	GetInt(name string) int
	Save()
}

type config struct {
	ints map[string]int
}

func (c *config) GetInt(name string) int {
	return c.ints[name]
}

func (c *config) Save() {
}

func MakeConfig() Config {
	return &config{}
}

// Control

type CommandHandler func(cmd string, args ...string)

type Command interface {
	Handle(cmd string, args ...string)
	ShortDesc() string
	LongDesc() string
}

type command struct {
	handler   CommandHandler
	shortDesc string
	longDesc  string
}

func (c *command) Handle(cmd string, args ...string) {
	c.handler(cmd, args...)
}

func (c *command) ShortDesc() string {
	return c.shortDesc
}

func (c *command) LongDesc() string {
	return c.longDesc
}

// CommandDispatcher receives commands and dispatches them. This is where
// plugins can add new commands. The dispatcher runs in the UI thread and must
// be non-blocking.
type CommandDispatcher interface {
	// Execute executes a command through the dispatcher.
	Execute(cmd string, args ...string)
	// Register registers a command so it can be executed later. In practice
	// commands should normally be registered on startup. Returns false if a
	// command was already registered and was lost.
	Register(cmd string, command Command) bool
}

/// KeyboardDispatcher receives keyboard input, processes it and expand macros
//as necessary, then send the generated commands to the CommandDispatcher.
type KeyboardDispatcher interface {
	OnKey(event termbox.Event)
	// Register registers a keyboard mapping so it can be executed later. In
	// practice keyboard mappings should normally be registered on startup.
	// Returns false if a key mapping was already registered and was lost.
	Register(key string, command Command) bool
}

type commandDispatcher struct {
	commands map[string]Command
}

func (c *commandDispatcher) Execute(cmd string, args ...string) {
	v, _ := c.commands[cmd]
	if v == nil {
		c.Execute("error", "Command \""+cmd+"\" is not registered")
	} else {
		v.Handle(cmd, args...)
	}
}

func (c *commandDispatcher) Register(cmd string, command Command) bool {
	_, ok := c.commands[cmd]
	c.commands[cmd] = command
	return !ok
}

func MakeCommandDispatcher() CommandDispatcher {
	return &commandDispatcher{make(map[string]Command)}
}

type keyboardDispatcher struct {
	dispatcher CommandDispatcher
	mappings   map[string]Command
}

func (k *keyboardDispatcher) OnKey(event termbox.Event) {
}

func (k *keyboardDispatcher) Register(key string, command Command) bool {
	_, ok := k.mappings[key]
	k.mappings[key] = command
	return !ok
}

func MakeKeyboardDispatcher(dispatcher CommandDispatcher) KeyboardDispatcher {
	return &keyboardDispatcher{dispatcher, make(map[string]Command)}
}

// Registers the native commands.
func RegisterDefaultCommands(display Display, dispatcher CommandDispatcher) {
	dispatcher.Register(
		"error",
		&command{
			func(cmd string, args ...string) {
				println("Faking an error")
			},
			"Prints an error message",
			"Prints an error message.",
		})
	dispatcher.Register(
		"open",
		&command{
			func(cmd string, args ...string) {
				println("Faking opening a file")
			},
			"Opens a file in a new buffer",
			"Opens a file in a new buffer.",
		})
	dispatcher.Register(
		"new",
		&command{
			func(cmd string, args ...string) {
				println("Faking opening a file")
			},
			"Create a new buffer",
			"Create a new buffer.",
		})
	dispatcher.Register(
		"shell",
		&command{
			func(cmd string, args ...string) {
				println("Faking a shell")
			},
			"Opens a shell process",
			"Opens a shell process in a new buffer.",
		})
	dispatcher.Register("doc",
		&command{
			func(cmd string, args ...string) {
				docArgs := make([]string, len(args)+1)
				docArgs[0] = "doc"
				copy(docArgs[1:], args)
				dispatcher.Execute("shell", docArgs...)
			},
			"Search godoc documentation",
			"Uses the 'doc' tool to get documentation about the text under the cursor.",
		})
	dispatcher.Register(
		"quit",
		&command{
			func(cmd string, args ...string) {
				println("Faking quit")
			},
			"Quits",
			"Quits the editor. Optionally bypasses writing the files to disk.",
		})
	dispatcher.Register("help",
		&command{
			func(cmd string, args ...string) {
				println("Faking help")
			},
			"Prints help",
			"Prints general help or help for a particular command.",
		})
}

type keyShortcut struct {
	command string
	key     string
}

// RegisterShortcut registers a shortcut to a command, so that the keyboard
// mapping inherits the command's help.
func RegisterShortcut(display Display, cmd_dispatcher CommandDispatcher, key_dispatcher KeyboardDispatcher, key string, command string) {
	// TODO c := &keyShortcut{}
}

// Registers the default keyboard mapping. Most keyboard mapping should simply
// execute the corresponding command.
// TODO(maruel): This should be remappable via a configuration flag, for
// example vim flavor vs emacs flavor.
func RegisterDefaultKeyboard(display Display, cmd_dispatcher CommandDispatcher, key_dispatcher KeyboardDispatcher) {
	RegisterShortcut(display, cmd_dispatcher, key_dispatcher, "F1", "help")
	RegisterShortcut(display, cmd_dispatcher, key_dispatcher, "Ctrl-C", "quit")
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

	cmd_dispatcher := MakeCommandDispatcher()
	key_dispatcher := MakeKeyboardDispatcher(cmd_dispatcher)
	display := MakeDisplay(key_dispatcher)
	RegisterDefaultCommands(display, cmd_dispatcher)
	RegisterDefaultKeyboard(display, cmd_dispatcher, key_dispatcher)

	if *command {
		for _, i := range flag.Args() {
			cmd_dispatcher.Execute(i)
		}
	} else if flag.NArg() > 0 {
		for _, i := range flag.Args() {
			cmd_dispatcher.Execute("open", i)
		}
	} else {
		// If nothing, opens a blank editor.
		cmd_dispatcher.Execute("new")
	}

	// Run the message loop.
	out := display.EventLoop()

	// Normal exit.
	termbox.SetCursor(0, 0)
	termbox.Flush()

	os.Exit(out)
}
