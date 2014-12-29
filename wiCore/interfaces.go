// Copyright 2014 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// This file defines all the interfaces to be used by the wi editor and to be
// accessable by plugins.

package wiCore

import (
	"fmt"
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

func (d DockingType) String() string {
	switch d {
	case DockingFill:
		return "DockingFill"
	case DockingFloating:
		return "DockingFloating"
	case DockingLeft:
		return "DockingLeft"
	case DockingRight:
		return "DockingRight"
	case DockingTop:
		return "DockingTop"
	case DockingBottom:
		return "DockingBottom"
	default:
		return "Unknown DockingType"
	}
}

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

func (b BorderType) String() string {
	switch b {
	case BorderNone:
		return "BorderNone"
	case BorderSingle:
		return "BorderSingle"
	case BorderDouble:
		return "BorderDouble"
	default:
		return "Unknown BorderType"
	}
}

// CommandCategory is used to put commands into sections for help purposes.
type CommandCategory int

const (
	// UnknownCategory means the command couldn't be categorized.
	UnknownCategory CommandCategory = iota
	// WindowCategory are commands relating to manipuling windows and UI in
	// general.
	WindowCategory
	// CommandsCategory are commands relating to manipulating commands, aliases,
	// keybindings.
	CommandsCategory
	// EditorCategory are commands relating to the editor lifetime.
	EditorCategory
	// DebugCategory are commands relating to debugging the app itself or plugins.
	DebugCategory

	// TODO(maruel): Add other categories.
)

func (c CommandCategory) String() string {
	switch c {
	case UnknownCategory:
		return "UnknownCategory"
	case WindowCategory:
		return "WindowCategory"
	case CommandsCategory:
		return "CommandsCategory"
	case DebugCategory:
		return "DebugCategory"
	default:
		return "Unknown CommandCategory"
	}
}

// KeyboardMode defines the keyboard mapping (input mode) to use.
type KeyboardMode int

const (
	// CommandMode is the mode where typing letters results in commands, not
	// content editing.
	CommandMode KeyboardMode = iota + 1
	// EditMode is the mode where typing letters results in content, not commands.
	EditMode
	// AllMode is to bind keys independent of the current mode. It is useful for
	// function keys, Ctrl-<letter>, arrow keys, etc.
	AllMode
)

// CommandID describes a command in the queue.
type CommandID struct {
	ProcessID    int
	CommandIndex int
}

func (c CommandID) String() string {
	return fmt.Sprintf("%d:%d", c.ProcessID, c.CommandIndex)
}

// CommandDispatcher owns the command queue. Use this interface to enqueue
// commands for execution.
type CommandDispatcher interface {
	// PostCommands appends several Command calls at the end of the queue. Using
	// this function guarantees that all the commands will be executed in order
	// without commands interfering.
	//
	// `callback` is called synchronously after the command is executed.
	PostCommands(cmds [][]string, callback func()) CommandID
}

// ViewFactory returns a new View.
type ViewFactory func(args ...string) View

// CommandDispatcherFull is a superset of CommandDispatcher for internal use.
type CommandDispatcherFull interface {
	CommandDispatcher

	// ExecuteCommand executes a command now. This is only meant to run a command
	// reentrantly; e.g. running a command triggers another one. This usually
	// happens for key binding, command aliases, when a command triggers an error.
	ExecuteCommand(w Window, cmdName string, args ...string)

	// ActiveWindow returns the current active Window.
	ActiveWindow() Window

	// RegisterViewFactory makes a nwe view available by name.
	RegisterViewFactory(name string, viewFactory ViewFactory) bool

	CurrentLanguage() LanguageMode
}

// Editor is the output device and the main process context. It shows the root
// window which covers the whole screen estate.
type Editor interface {
	CommandDispatcherFull

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
	CommandDispatcher

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

// Document represents an open document. It can be accessed by zero, one or
// multiple View. For example the document may not be visible at all as a 'back
// buffer', may be loaded in a View or in multiple View, each having their own
// coloring and cursor position.
type Document interface {
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

// Control

// CommandHandler executes the command cmd on the Window w.
type CommandHandler func(cd CommandDispatcherFull, w Window, args ...string)

// Command describes a registered command that can be triggered directly at the
// command prompt, via a keybinding or a plugin.
type Command interface {
	// Name is the name of the command.
	Name() string
	// Handle executes the command.
	Handle(cd CommandDispatcherFull, w Window, args ...string)
	// Category returns the category the command should be bucketed in, for help
	// documentation purpose.
	Category(cd CommandDispatcherFull, w Window) CommandCategory
	// ShortDesc returns a short description of the command in the language
	// requested. It defaults to English if the description was not translated in
	// this language.
	ShortDesc(cd CommandDispatcherFull, w Window) string
	// LongDesc returns a long explanation of the command in the language
	// requested. It defaults to English if the description was not translated in
	// this language.
	LongDesc(cd CommandDispatcherFull, w Window) string
}

// Commands stores the known commands. This is where plugins can add new
// commands. Each View contains its own Commands.
type Commands interface {
	// Register registers a command so it can be executed later. In practice
	// commands should normally be registered on startup. Returns false if a
	// command was already registered and was lost.
	Register(cmd Command) bool

	// Get returns a command if registered, nil otherwise.
	Get(cmdName string) Command
}

// KeyBindings stores the mapping between keyboard entry and commands. This
// includes what can be considered "macros" as much as casual things like arrow
// keys.
type KeyBindings interface {
	// Set registers a keyboard mapping. In practice keyboard mappings
	// should normally be registered on startup. Returns false if a key mapping
	// was already registered and was lost. Set cmdName to "" to remove a key
	// binding.
	Set(mode KeyboardMode, key KeyPress, cmdName string) bool

	// Get returns a command if registered, nil otherwise.
	Get(mode KeyboardMode, key KeyPress) string
}

// EventType is the type of the event being flowed through the Window hierarchy
// and plugins. EventListener receive these.
//
// TODO(maruel): Dedupe from editor/terminal.go, testing.
type EventType string

// TODO(maruel): Dedupe from editor/terminal.go, testing.
//
// TODO(maruel): Use int or string? int is faster, string is likely more
// "extendable".
const (
	EventDocumentCreated     EventType = "document_created"
	EventDocumentCursorMoved           = "document_cursor_moved"
	EventTerminalResized               = "terminal_resized"
	EventTerminalKeyPressed            = "terminal_key_pressed"
	EventViewCreated                   = "view_created"
	EventWindowCreated                 = "window_created"
	EventWindowResized                 = "window_resized"

// Etc.
)

// EventListener are called on events.
//
// TODO(maruel): Experimenting with the idea.
type EventListener interface {
	OnEvent(t EventType, i interface{})
}

// Utility functions.

// PostCommand appends a Command at the end of the queue.
// It is a shortcut to cd.PostCommands([][]string{[]string{cmdName, args...}},
// callback). Sadly, using `...string` means that callback cannot be the last
// parameter.
func PostCommand(cd CommandDispatcher, callback func(), cmdName string, args ...string) CommandID {
	line := make([]string, len(args)+1)
	line[0] = cmdName
	copy(line[1:], args)
	return cd.PostCommands([][]string{line}, callback)
}

// GetCommand traverses the Window hierarchy tree to find a View that has
// the command cmd in its Commands mapping. If Window is nil, it starts with
// the Editor's active Window.
func GetCommand(cd CommandDispatcherFull, w Window, cmdName string) Command {
	if w == nil {
		w = cd.ActiveWindow()
	}
	for {
		cmd := w.View().Commands().Get(cmdName)
		if cmd != nil {
			return cmd
		}
		w = w.Parent()
		if w == nil {
			return nil
		}
	}
}

// GetKeyBindingCommand traverses the Editor's Window tree to find a View that
// has the key binding in its Keyboard mapping.
func GetKeyBindingCommand(e Editor, mode KeyboardMode, key KeyPress) string {
	active := e.ActiveWindow()
	for {
		cmdName := active.View().KeyBindings().Get(mode, key)
		if cmdName != "" {
			return cmdName
		}
		active = active.Parent()
		if active == nil {
			return ""
		}
	}
}

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
