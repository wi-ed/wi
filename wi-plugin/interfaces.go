// Copyright 2014 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// This file defines all the interfaces to be used by the wi editor and to be
// accessable by plugins.

package wi

import (
	"fmt"
	// TODO(maruel): Stop leaking this modules in the interface. We can't "type
	// Buffer tulib.Buffer" or else we lose all the methods and will have to
	// reimplement them here.
	"github.com/maruel/tulib"
)

// UI

// DockingType defines the relative position of a Window relative to its parent.
type DockingType int

const (
	// The Window uses all the available space.
	DockingFill DockingType = iota
	// The Window is not constrained by the parent window size and location.
	DockingFloating
	DockingLeft
	DockingRight
	DockingTop
	DockingBottom
	DockingUnknown
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
	// Width is 0.
	BorderNone BorderType = iota
	// Width is 1.
	BorderSingle
	// Despite its name, width is 1, only the glyph is different.
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
	UnknownCategory CommandCategory = iota
	// Commands relating to manipuling windows and UI in general.
	WindowCategory
	// Commands relating to manipulating commands, aliases, keybindings.
	CommandsCategory
	// Commands relating to the editor lifetime.
	EditorCategory
	// Commands relating to debugging the app itself or plugins.
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

// Switches keyboard mapping based on the input mode. These modes are hardcode;
// adding a new mode would require rebuilding the editor (2 seconds, really).
type KeyboardMode int

const (
	_ KeyboardMode = iota
	// CommandMode is the mode where typing letters results in commands, not
	// content editing.
	CommandMode
	// EditMode is the mode where typing letters results in content, not commands.
	EditMode
	// AllMode is to bind keys independent of the current mode. It is useful for
	// function keys, Ctrl-<letter>, arrow keys, etc.
	AllMode
)

// CommandDispatcher owns the command queue. Use this interface to enqueue
// commands for execution.
type CommandDispatcher interface {
	// PostCommands appends several Command calls at the end of the queue. Using
	// this function guarantees that all the commands will be executed in order
	// without commands interfering.
	PostCommands(cmds [][]string)
}

// ViewFactory returns a new View.
type ViewFactory func() View

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
// instances, designating the actual Window by its .Id() method.
type Window interface {
	fmt.Stringer

	// Id returns the unique id for this Window. The id is guaranteed to be
	// unique through the process lifetime of the editor.
	Id() string

	// Tree returns a textual representation of the Window hierarchy. It is only
	// for debugging purpose.
	// TODO(maruel): Remove.
	Tree() string

	// Parent returns the parent Window.
	Parent() Window
	// ChildrenWindows returns a copy of the slice of children windows.
	ChildrenWindows() []Window

	// NewChildWindow() adds a View in a new child Window located at 'docking'
	// position. It is invalid to add a child Window with the same docking as one
	// already present. In this case, nil is returned.
	// TODO(maruel): Convert to a command.
	NewChildWindow(view View, docking DockingType) Window

	// Rect returns the position based on the parent Window area, except if
	// Docking() is DockingFloating.
	Rect() tulib.Rect

	// SetRect sets the rect of this Window, based on the parent's Window own
	// Rect(). It updates Rect() and synchronously updates the child Window that
	// are not DockingFloating.
	// TODO(maruel): Convert to a command.
	SetRect(rect tulib.Rect)

	// Buffer returns the display buffer for this Window. The Window
	// double-buffers the View buffer so it could stale data if the View is slow
	// to draw itself.
	Buffer() *tulib.Buffer

	Docking() DockingType

	// SetView replaces the current View with a new one. This forces an
	// invalidation and a redraw.
	// TODO(maruel): Convert to a command.
	SetView(view View)

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
	// IsDirty is true if the content should be saved before quitting.
	IsDirty() bool

	// IsDisabled returns false if the View can be activated to receive user
	// inputs at all.
	IsDisabled() bool

	// Buffer returns the display buffer for this Window.
	Buffer() *tulib.Buffer

	// NaturalSize returns the natural size of the content. It can be -1 for as
	// long/large as possible, 0 if indeterminate. The return value of this
	// function is not affected by SetSize().
	NaturalSize() (width, height int)
	// SetSize resets the View Buffer size.
	SetSize(x, y int)
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
	Set(mode KeyboardMode, keyName string, cmdName string) bool

	// Get returns a command if registered, nil otherwise.
	Get(mode KeyboardMode, keyName string) string
}

// Utility functions.

// PostCommand appends a Command at the end of the queue.
// It is a shortcut to cd.PostCommands([][]string{[]string{cmdName, args...}})
func PostCommand(cd CommandDispatcher, cmdName string, args ...string) {
	line := make([]string, len(args)+1)
	line[0] = cmdName
	copy(line[1:], args)
	cd.PostCommands([][]string{line})
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
		w := w.Parent()
		if w == nil {
			return nil
		}
	}
}

// GetKeyBindingCommand traverses the Editor's Window tree to find a View that
// has the key binding in its Keyboard mapping.
func GetKeyBindingCommand(e Editor, mode KeyboardMode, keyName string) string {
	active := e.ActiveWindow()
	for {
		cmdName := active.View().KeyBindings().Get(mode, keyName)
		if cmdName != "" {
			return cmdName
		}
		active := active.Parent()
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
func PositionOnScreen(w Window) tulib.Rect {
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
