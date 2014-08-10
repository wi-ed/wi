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

// CommandCategory is used to put commands into sections for help purposes.
type CommandCategory int

const (
	UnknownCategory CommandCategory = iota
	// Commands relating to manipuling windows and UI in general.
	WindowCategory
	// Commands relating to manipulating commands, aliases, keybindings.
	CommandsCategory

	// TODO(maruel): Add other categories.
)

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

// LanguageMode is a language selection for UI purposes.
type LanguageMode string

const (
	// TODO: Add new languages when translating the application.

	LangEn = "en"
	LangEs = "es"
	LangFr = "fr"
)

// CommandDispatcher owns the command queue. Use this interface to enqueue
// commands for execution.
type CommandDispatcher interface {
	// PostCommand appends a Command at the end of the queue.
	PostCommand(cmdName string, args ...string)

	// WaitQueueEmpty waits for the enqueued commands to complete before moving
	// on. It can be used for example when quitting the process safely.
	WaitQueueEmpty()
}

type CommandDispatcherFull interface {
	CommandDispatcher

	// ExecuteCommand executes a command now. This is only meant to run a command
	// reentrantly; e.g. running a command triggers another one. This usually
	// happens for key binding, command aliases, when a command triggers an error.
	ExecuteCommand(w Window, cmdName string, args ...string)

	// ActiveWindow returns the current active Window.
	ActiveWindow() Window
	// ActivateWindow activates a Window.
	ActivateWindow(w Window)

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
// windows are not bounded by the parent window. The Window itself doesn't
// interact with the user, since it only has a non-client area (the border).
// All the client area is covered by the View.
type Window interface {
	// Parent returns the parent Window.
	Parent() Window
	// ChildrenWindows returns a copy of the slice of children windows.
	ChildrenWindows() []Window
	// TODO(maruel): Accept a Window, not a View. This permits more complex
	// window creation.
	NewChildWindow(view View, docking DockingType) Window
	// Remove detaches a child window tree from the tree. Care should be taken to
	// not remove the active Window.
	Remove(w Window)

	// Rect returns the position based on the Editor's root Window, not the
	// parent Window.
	Rect() tulib.Rect
	SetRect(rect tulib.Rect)

	IsInvalid() bool
	// Invalidate forces the Window to be redrawn at next drawing. Otherwise
	// drawing this Window will be skipped. In general the View should be
	// invalidated, not the Window. This is relevant when the non-client area
	// needs an update.
	Invalidate()

	// Buffer returns the display buffer for this Window. This indirectly clears
	// the Invalid bit.
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

	// IsDisabled returns false if the View can be activated to receive user
	// inputs at all.
	IsDisabled() bool

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
type CommandHandler func(cd CommandDispatcherFull, w Window, args ...string)

// Command describes a registered command that can be triggered directly at the
// command prompt, via a keybinding or a plugin.
type Command interface {
	// Handle executes the command.
	Handle(cd CommandDispatcherFull, w Window, args ...string)
	// Category returns the category the command should be bucketed in, for help
	// documentation purpose.
	Category(cd CommandDispatcherFull) CommandCategory
	// ShortDesc returns a short description of the command in the language
	// requested. It defaults to English if the description was not translated in
	// this language.
	ShortDesc(cd CommandDispatcherFull) string
	// LongDesc returns a long explanation of the command in the language
	// requested. It defaults to English if the description was not translated in
	// this language.
	LongDesc(cd CommandDispatcherFull) string
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
	// Set registers a keyboard mapping. In practice keyboard mappings
	// should normally be registered on startup. Returns false if a key mapping
	// was already registered and was lost. Set cmdName to "" to remove a key
	// binding.
	Set(mode KeyboardMode, keyName string, cmdName string) bool

	// Get returns a command if registered, nil otherwise.
	Get(mode KeyboardMode, keyName string) string
}

// Utility functions.

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
