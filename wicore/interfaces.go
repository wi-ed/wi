// Copyright 2014 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// This file defines all the interfaces to be used by the wi editor and to be
// accessable by plugins.

//go:generate stringer -output=interfaces_string.go -type=BorderType,DockingType
//go:generate go run ../tools/wi-event-generator/main.go -output event_registry_decl.go

package wicore

import (
	"fmt"
	"io"

	"github.com/maruel/wi/pkg/key"
	"github.com/maruel/wi/pkg/lang"
)

// UI

// DockingType defines the relative position of a Window relative to its parent.
type DockingType int

// Available docking options.
const (
	// DockingUnknown is an invalid value.
	DockingUnknown DockingType = iota

	// DockingFill means the Window uses all the available space.

	DockingFill
	// DockingFloating means the Window is not constrained by the parent window
	// size and location.
	DockingFloating

	DockingLeft
	DockingRight
	DockingTop
	DockingBottom
)

// StringToDockingType converts a string back to a DockingType.
func StringToDockingType(s string) DockingType {
	switch s {
	case "fill":
		return DockingFill
	case "floating":
		return DockingFloating
	case "left":
		return DockingLeft
	case "right":
		return DockingRight
	case "top":
		return DockingTop
	case "bottom":
		return DockingBottom
	default:
		return DockingUnknown
	}
}

// BorderType defines the type of border for a Window.
type BorderType int

const (
	// BorderNone means width is 0.
	BorderNone BorderType = iota
	// BorderSingle means width is 1.
	BorderSingle
	// BorderDouble means width is 1 despite its name, only the glyph is
	// different.
	BorderDouble
)

// EventID is to be used to cancel an event listener.
type EventID int

/*
// EventID describes an event in the queue.
type EventID struct {
	ProcessID int // 0 is the main editor process, other values are for plugins.
	Index     int
}

func (c EventID) String() string {
	return fmt.Sprintf("%d:%d", c.ProcessID, c.Index)
}
*/

// EventsDefinition declares the valid events.
//
// Do not use this interface directly, use the automatically-generated
// interface EventRegistry instead.
type EventsDefinition interface {
	// TriggerCommands dispatches one or multiple commands to the current active
	// listener. Normally, it's the View contained to the active Window. Using
	// this function guarantees that all the commands will be executed in order
	// without commands interfering.
	//
	// `callback` is called synchronously after the command is executed.
	TriggerCommands(cmds EnqueuedCommands)
	TriggerDocumentCreated(doc Document)
	TriggerDocumentCursorMoved(doc Document, col, row int)
	TriggerEditorKeyboardModeChanged(mode KeyboardMode)
	TriggerEditorLanguage(l lang.Language)
	TriggerTerminalResized()
	TriggerTerminalKeyPressed(key key.Press)
	TriggerViewCreated(view View)
	TriggerWindowCreated(window Window)
	TriggerWindowResized(window Window)
}

// Editor is the output device and the main process context. It shows the root
// window which covers the whole screen estate.
type Editor interface {
	EventRegistry

	// ExecuteCommand executes a command now. This is only meant to run a command
	// reentrantly; e.g. running a command triggers another one. This usually
	// happens for key binding, command aliases, when a command triggers an error.
	//
	// TODO(maruel): Remove?
	ExecuteCommand(w Window, cmdName string, args ...string)

	// ActiveWindow returns the current active Window.
	ActiveWindow() Window

	// RegisterViewFactory makes a new view available by name.
	RegisterViewFactory(name string, viewFactory ViewFactory) bool

	// KeyboardMode is global to the editor. It matches vim behavior. For example
	// in a 2-window setup while in insert mode, using Ctrl-O, Ctrl-W, Down will
	// move to the next window but will stay in insert mode.
	//
	// Technically, each View could have their own KeyboardMode but in practice
	// it just creates a cognitive overhead without much benefit.
	KeyboardMode() KeyboardMode

	// Version returns the version number of this build of wi.
	Version() string
}

// Window is a View container. It defines the position, Z-ordering via
// hierarchy and decoration. It can have multiple child windows. The child
// windows are not bounded by the parent window if DockingFloating is used. The
// Window itself doesn't interact with the user, since it only has a non-client
// area (the border). All the client area is covered by the View.
//
// Split view is not supported. A 4-way merge setup can be created with the
// following Window setup as 4 child Window of the root Window:
//
//    +-----------+-----------+------------+
//    |  Remote   |Merge Base*|   Local    |
//    |DockingLeft|DockingFill|DockingRight|
//    |           |           |            |
//    +-----------+-----------+------------+
//    |              Result                |
//    |           DockingBottom            |
//    |                                    |
//    +------------------------------------+
//
// * The Merge Base View can be either:
//   - The root Window's View that is constained.
//   - A child Window set as DockingFill. In this case, the root Window View is
//     not visible.
//
// The end result is that this use case doesn't require any "split" support.
// Further subdivision can be done via Window containment.
//
// The Window interface exists for synchronous query but modifications
// (creation, closing, moving) are done asynchronously via commands. A set of
// privileged commands starting with the prefix "window_" can modify Window
// instances, designating the actual Window by its .ID() method.
type Window interface {
	fmt.Stringer

	// ID returns the unique id for this Window. The id is guaranteed to be
	// unique through the process lifetime of the editor.
	ID() string

	// Tree returns a textual representation of the Window hierarchy. It is only
	// for debugging purpose.
	// TODO(maruel): Remove.
	Tree() string

	// Parent returns the parent Window.
	Parent() Window
	// ChildrenWindows returns a copy of the slice of children windows.
	ChildrenWindows() []Window

	// Rect returns the position based on the parent Window area, except if
	// Docking() is DockingFloating.
	Rect() Rect

	// Docking returns where this Window is docked relative to the parent Window.
	// A DockingFloating window is effectively starting a new independent Rect.
	Docking() DockingType

	// View returns the View contained by this Window. There is exactly one.
	View() View
}

// View is content presented in a Window. For example it can be a TextBuffer or
// a command box. View define the key binding and commands supported so it
// responds to user input.
type View interface {
	fmt.Stringer
	io.Closer

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

	// IsDisabled returns false if the View can be activated to receive user
	// inputs at all.
	IsDisabled() bool

	// Buffer returns the display buffer for this Window.
	Buffer() *Buffer

	// NaturalSize returns the natural size of the content. It can be -1 for as
	// long/large as possible, 0 if indeterminate. The return value of this
	// function is not affected by SetSize().
	NaturalSize() (width, height int)
	// SetSize resets the View Buffer size.
	SetSize(x, y int)

	// OnAttach is called by the Window after it was attached.
	// TODO(maruel): Maybe split in ViewFull?
	OnAttach(w Window)

	// DefaultFormat returns the default coloring for this View. If this View has
	// an CellFormat.Empty()==true format, it will uses whatever parent Window's
	// View DefaultFormat().
	DefaultFormat() CellFormat
}

// ViewFactory returns a new View.
type ViewFactory func(e Editor, args ...string) View

// Document represents an open document. It can be accessed by zero, one or
// multiple View. For example the document may not be visible at all as a 'back
// buffer', may be loaded in a View or in multiple View, each having their own
// coloring and cursor position.
type Document interface {
	fmt.Stringer
	io.Closer

	// RenderInto renders a view of a document.
	//
	// TODO(maruel): Likely return a new Buffer instance instead, for RPC
	// friendlyness. To be decided.
	RenderInto(buffer *Buffer, view View, offsetLine, offsetColumn int)

	// IsDirty is true if the content should be saved before quitting.
	IsDirty() bool
}

// Config

// Config is the configuration manager.
type Config interface {
	GetInt(name string) int
	GetString(name string) string
	Save()
}

// Utility functions.

// RootWindow returns the root Window when given any Window in the tree.
func RootWindow(w Window) Window {
	for {
		if w.Parent() == nil {
			return w
		}
		w = w.Parent()
	}
}

// PositionOnScreen returns the exact position on screen of a Window.
func PositionOnScreen(w Window) Rect {
	out := w.Rect()
	if w.Docking() == DockingFloating {
		return out
	}
	for {
		w = w.Parent()
		if w == nil {
			break
		}
		// Take in account the parent Window position.
		r := w.Rect()
		out.X += r.X
		out.Y += r.Y
		if w.Docking() == DockingFloating {
			break
		}
	}
	return out
}
