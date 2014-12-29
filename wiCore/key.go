// Copyright 2014 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package wiCore

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
	if k.Key == KeyNone {
		if k.Ch == ' ' {
			out += "Space"
		} else {
			out += string(k.Ch)
		}
	} else {
		out += k.String()
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
