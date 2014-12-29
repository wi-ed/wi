// Copyright 2014 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package wicore

import (
	"strings"
)

// Key represents a non-character key.
type Key int

// Known non-character keys.
const (
	KeyNone Key = iota
	KeyF1
	KeyF2
	KeyF3
	KeyF4
	KeyF5
	KeyF6
	KeyF7
	KeyF8
	KeyF9
	KeyF10
	KeyF11
	KeyF12
	KeyF13
	KeyF14
	KeyF15
	KeyEscape
	KeyBackspace
	KeyTab
	KeyEnter
	KeyInsert
	KeyDelete
	KeyHome
	KeyEnd
	KeyArrowUp
	KeyArrowDown
	KeyArrowLeft
	KeyArrowRight
	KeyPageUp
	KeyPageDown
	keyLast
)

func (k Key) String() string {
	switch k {
	case KeyNone:
		return "<None>"
	case KeyF1:
		return "F1"
	case KeyF2:
		return "F2"
	case KeyF3:
		return "F3"
	case KeyF4:
		return "F4"
	case KeyF5:
		return "F5"
	case KeyF6:
		return "F6"
	case KeyF7:
		return "F7"
	case KeyF8:
		return "F8"
	case KeyF9:
		return "F9"
	case KeyF10:
		return "F10"
	case KeyF11:
		return "F11"
	case KeyF12:
		return "F12"
	case KeyF13:
		return "F13"
	case KeyF14:
		return "F14"
	case KeyF15:
		return "F15"
	case KeyEscape:
		return "Escape"
	case KeyBackspace:
		return "Backspace"
	case KeyTab:
		return "Tab"
	case KeyEnter:
		return "Enter"
	case KeyInsert:
		return "Insert"
	case KeyDelete:
		return "Delete"
	case KeyHome:
		return "Home"
	case KeyEnd:
		return "End"
	case KeyArrowUp:
		return "Up"
	case KeyArrowDown:
		return "Down"
	case KeyArrowLeft:
		return "Left"
	case KeyArrowRight:
		return "Right"
	case KeyPageUp:
		return "PageUp"
	case KeyPageDown:
		return "PageDown"
	default:
		return "<Invalid>"
	}
}

func StringToKey(key string) Key {
	switch key {
	case "F1":
		return KeyF1
	case "F2":
		return KeyF2
	case "F3":
		return KeyF3
	case "F4":
		return KeyF4
	case "F5":
		return KeyF5
	case "F6":
		return KeyF6
	case "F7":
		return KeyF7
	case "F8":
		return KeyF8
	case "F9":
		return KeyF9
	case "F10":
		return KeyF10
	case "F11":
		return KeyF11
	case "F12":
		return KeyF12
	case "F13":
		return KeyF13
	case "F14":
		return KeyF14
	case "F15":
		return KeyF15
	case "Escape":
		return KeyEscape
	case "Backspace":
		return KeyBackspace
	case "Tab":
		return KeyTab
	case "Enter":
		return KeyEnter
	case "Insert":
		return KeyInsert
	case "Delete":
		return KeyDelete
	case "Home":
		return KeyHome
	case "End":
		return KeyEnd
	case "Up":
		return KeyArrowUp
	case "Down":
		return KeyArrowDown
	case "Left":
		return KeyArrowLeft
	case "Right":
		return KeyArrowRight
	case "PageUp":
		return KeyPageUp
	case "PageDown":
		return KeyPageDown
	default:
		return KeyNone
	}
}

// KeyPress represents a key press.
//
// Only one of Key or Ch is set.
type KeyPress struct {
	Alt  bool
	Ctrl bool
	Key  Key  // Non-character key (e.g. F-keys, arrows, etc). Set to KeyNone when not used.
	Ch   rune // Character key, e.g. letter, number. Set to rune(0) when not used.
}

func (k KeyPress) String() string {
	if k.Key == KeyNone && k.Ch == 0 {
		return ""
	}
	out := ""
	if k.Ctrl {
		out += "Ctrl-"
	}
	if k.Alt {
		out += "Alt-"
	}
	if k.Key > KeyNone && k.Key < keyLast {
		out += k.Key.String()
	} else {
		if k.Ch == ' ' {
			out += "Space"
		} else {
			out += string(k.Ch)
		}
	}
	return out
}

// IsMeta returns true if a key press is a meta key.
func (k KeyPress) IsMeta() bool {
	return k.Alt || k.Ctrl || k.Key != KeyNone
}

// IsValid returns true if the object represents a key press.
func (k KeyPress) IsValid() bool {
	return k.Alt || k.Ctrl || k.Key != KeyNone || k.Ch != rune(0)
}

// Parses a string and returns a KeyPress.
func StringToKeyPress(keyName string) KeyPress {
	out := KeyPress{}
	if strings.HasPrefix(keyName, "Ctrl-") {
		keyName = keyName[5:]
		out.Ctrl = true
	}
	if strings.HasPrefix(keyName, "Alt-") {
		keyName = keyName[4:]
		out.Alt = true
	}
	rest := []rune(keyName)
	l := len(rest)
	if l == 1 {
		out.Ch = rest[0]
	} else if l > 1 {
		if keyName == "Space" {
			out.Ch = ' '
		} else {
			out.Key = StringToKey(keyName)
		}
	}
	return out
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
