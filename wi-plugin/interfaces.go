// Copyright 2014 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// This file defines all the interfaces to be used by the wi editor and to be
// accessable by plugins.

package wi

import (
	// TODO(maruel): Stop leaking this modules in the interface.
	"github.com/nsf/tulib"
)

// UI

// DockingType defines the relative position of a Window relative to its parent.
type DockingType int

const (
	// The window is not constrained by the parent window size and location.
	DockingFill DockingType = iota
	DockingFloating
	DockingLeft
	DockingRight
	DockingTop
	DockingBottom
	DockingCenter
)

// BorderType defines the type of border for a Window.
type BorderType int

const (
	// Width is 0.
	BorderNone BorderType = iota
	// Width is 1.
	BorderSingle
	// Despite its name, width is 1.
	BorderDouble
)

// Display is the output device. It shows the root window which covers the
// whole screen estate.
type Display interface {
	// Redraws all the invalidated windows.
	Draw()

	// ActiveWindow returns the current active Window.
	ActiveWindow() Window
	// ActivateWindow activates a Window.
	ActivateWindow(w Window)

	Height() int
	Width() int
}

// CommandCategory is used to put commands into sections for help purposes.
type CommandCategory int

const (
	// Commands relating to manipuling windows and UI in general.
	WindowCategory CommandCategory = iota
	// TODO(maruel): Add other categories.
)

// Window is a View container. It defines the position, Z-ordering via
// hierarchy and decoration. It can have multiple child windows. The child
// windows are not bounded by the parent window. The Window itself doesn't
// interact with the user, since it only has a non-client area (the border).
// All the client area is covered by the View.
type Window interface {
	Parent() Window
	ChildrenWindows() []Window
	NewChildWindow(view View, docking DockingType) Window
	// Remove detaches a child window tree from the tree. Care should be taken to
	// not remove the active Window.
	Remove(w Window)

	// Rect returns the position based on the Display, not the parent Window.
	Rect() tulib.Rect
	SetRect(rect tulib.Rect)

	// Invalidate forces the Window to be redrawn at next drawing. Otherwise
	// drawing this Window will be skipped.
	Invalidate()

	// Buffer returns the display buffer for this Window.
	Buffer() tulib.Buffer

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

// View is content presented in a Window. For example it can be a TextBuffer or
// a command box. View define the key binding and commands supported so it
// responds to user input.
type View interface {
	// Commands returns the commands registered for this specific view. For
	// example a text window will have commands specific to the file type
	// enabled.
	Commands() Commands

	// KeyBindings returns the key bindings registered for this specific key. For
	// example the 'command' view has different behavior on up/down arrow keys
	// than a text editor view.
	KeyBindings() KeyBindings

	// Title is View's title, which can be the current file name or any other
	// relevant detail.
	Title() string
	// IsDirty is true if the content should be saved before quitting.
	IsDirty() bool
	// IsInvalid is true if the View needs to be redraw.
	IsInvalid() bool

	// Draws itself into a buffer.
	DrawInto(buffer tulib.Buffer)

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

// Command describes a registered command that can be triggered directly at the
// command prompt, via a keybinding or a plugin.
type Command interface {
	Handle(w Window, args ...string)
	Category() CommandCategory
	ShortDesc() string
	LongDesc() string
}

// Commands stores the known commands. This is where plugins can add new
// commands. Each View contains its own Commands.
type Commands interface {
	// Register registers a command so it can be executed later. In practice
	// commands should normally be registered on startup. Returns false if a
	// command was already registered and was lost.
	Register(cmdName string, cmd Command) bool

	// Get returns a command if registered, nil otherwise.
	Get(cmdName string) Command
}

// KeyBindings stores the mapping between keyboard entry and commands. This
// includes what can be considered "macros" as much as casual things like arrow
// keys.
type KeyBindings interface {
	// Register registers a keyboard mapping. In practice keyboard mappings
	// should normally be registered on startup. Returns false if a key mapping
	// was already registered and was lost.
	Register(keyName string, cmdName string) bool

	// Get returns a command if registered, nil otherwise.
	Get(keyName string) string
}

// GetCommand traverses the Display's active Window hierarchy tree to find a
// View that has the command cmd in its Commands mapping.
func GetCommand(d Display, cmdName string) Command {
	return GetCommandWindow(d.ActiveWindow(), cmdName)
}

// GetCommandWindow traverses the Window hierarchy tree to find a View that has
// the command cmd in its Commands mapping.
func GetCommandWindow(w Window, cmdName string) Command {
	for {
		cmd := w.View().Commands().Get(cmdName)
		if cmd != nil {
			return cmd
		}
		w := w.Parent()
		if w == nil {
			return nil
		}
	}
}

// ExecuteCommand executes the command if possible or prints an error message
// otherwise.
func ExecuteCommand(d Display, cmdName string, args ...string) {
	ExecuteCommandWindow(d.ActiveWindow(), cmdName, args...)
}

// ExecuteCommandWindow executes the command if possible or prints an error
// message otherwise.
func ExecuteCommandWindow(w Window, cmdName string, args ...string) {
	cmd := GetCommandWindow(w, cmdName)
	if cmd == nil {
		ExecuteCommandWindow(
			w, "alert", "Command \""+cmdName+"\" is not registered")
	} else {
		cmd.Handle(w, args...)
	}
}

// GetKeyBindingCommand traverses the Display's Window tree to find a View that
// has the key binding in its Keyboard mapping.
func GetKeyBindingCommand(d Display, keyName string) string {
	active := d.ActiveWindow()
	for {
		cmdName := active.View().KeyBindings().Get(keyName)
		if cmdName != "" {
			return cmdName
		}
		active := active.Parent()
		if active == nil {
			return ""
		}
	}
}

// RootWindow, given any Window in the tree, returns the root Window.
func RootWindow(w Window) Window {
	for {
		if w.Parent() == nil {
			return w
		}
		w = w.Parent()
	}
}
