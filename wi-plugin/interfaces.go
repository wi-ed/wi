// Copyright 2014 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package wi

import (
	// TODO(maruel): Stop leaking these modules in the interface.
	"github.com/nsf/termbox-go"
	"github.com/nsf/tulib"
)

// UI

// BorderType defines the type of border for a Window.
type BorderType int

// DockingType defines the relative position of a Window relative to its parent.
type DockingType int

const (
	// The window is not constrained by the parent window size and location.
	Fill DockingType = iota
	Floating
	Left
	Right
	Top
	Bottom
	Center
)

const (
	// Width is 0.
	None BorderType = iota
	// Width is 1.
	Single
	// Width is 1.
	Double
)

// Display is the output device. It shows the root window which covers the
// whole screen estate.
type Display interface {
	// Redraws all the invalidated windows.
	Draw()
	ActiveWindow() Window
	Height() int
	Width() int
}

// Window is a View container. It defines the position, Z-ordering via
// hierarchy and decoration. It can have multiple child windows. The child
// windows are not bounded by the parent window.
type Window interface {
	// Each Window has its own command dispatcher. For example a text window will
	// have commands specific to the file type enabled.
	Command() CommandDispatcher

	// Each Window has its own keyboard dispatcher for Window specific commands,
	// for example the 'command' window has different behavior than a golang
	// editor window.
	Keyboard() KeyboardDispatcher

	Parent() Window
	ChildrenWindows() []Window
	NewChildWindow(view View, docking DockingType)
	// Remove detaches a child window tree from the tree.
	Remove(w Window)

	// Rect returns the position based on the Display, not the parent Window.
	Rect() tulib.Rect
	SetRect(rect tulib.Rect)

	IsInvalid() bool
	// Invalidate forces the Window to be redrawn at next drawing. Otherwise
	// drawing this Window will be skipped.
	Invalidate()

	Docking() DockingType
	// This will forces an invalidation.
	SetDocking(docking DockingType)

	SetView(view View)
	// This will forces an invalidation.
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
	// NaturalSize returns the natural size of the content. It can be -1 for as
	// long/large as possible, 0 if indeterminate.
	NaturalSize() (x, y int)
	SetBuffer(buffer TextBuffer)
	Buffer() TextBuffer
}

// Config

// Configuration manager.
type Config interface {
	GetInt(name string) int
	GetString(name string) string
	Save()
}

// Control

// CommandHandler executes the command cmd on the Window w.
type CommandHandler func(w Window, args ...string)

type Command interface {
	Handle(w Window, args ...string)
	ShortDesc() string
	LongDesc() string
}

// CommandDispatcher receives commands and dispatches them. This is where
// plugins can add new commands. The dispatcher runs in the UI thread and must
// be non-blocking.
type CommandDispatcher interface {
	// Execute executes a command through the dispatcher.
	Execute(w Window, cmd string, args ...string)
	// Register registers a command so it can be executed later. In practice
	// commands should normally be registered on startup. Returns false if a
	// command was already registered and was lost.
	Register(cmd string, command Command) bool
}

/// KeyboardDispatcher receives keyboard input, processes it and expand macros
//as necessary, then send the generated commands to the CommandDispatcher.
type KeyboardDispatcher interface {
	// OnKey converts a key press into a command string, or "" if unprocessed.
	OnKey(event termbox.Event) string
	// Register registers a keyboard mapping so it can be executed later. In
	// practice keyboard mappings should normally be registered on startup.
	// Returns false if a key mapping was already registered and was lost.
	Register(key string, cmd string) bool
}
