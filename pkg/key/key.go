// Copyright 2014 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

//go:generate stringer -type=Key

// Package key implements generic key definition.
package key

import "strings"

// Key represents a non-character key.
type Key int

// Known non-character keys.
//
// Keys between None and Meta are not meta keys as they can be represented with
// a character, e.g. \n, \i, ' ', \t.
const (
	None Key = iota
	Enter
	Escape
	Space
	Tab
	Meta
	F1
	F2
	F3
	F4
	F5
	F6
	F7
	F8
	F9
	F10
	F11
	F12
	F13
	F14
	F15
	Backspace
	Delete
	Insert
	Home
	End
	PageUp
	PageDown
	Up
	Down
	Left
	Right
	last
)

// StringToKey parses the string presentation of a key back into a Key.
//
// Returns None on invalid key name.
func StringToKey(key string) Key {
	// TODO(maruel): Create tool to generate this automatically.
	switch key {
	case "Escape":
		return Escape
	case "Enter":
		return Enter
	case "Space":
		return Space
	case "Tab":
		return Tab
	case "Meta":
		return Meta
	case "F1":
		return F1
	case "F2":
		return F2
	case "F3":
		return F3
	case "F4":
		return F4
	case "F5":
		return F5
	case "F6":
		return F6
	case "F7":
		return F7
	case "F8":
		return F8
	case "F9":
		return F9
	case "F10":
		return F10
	case "F11":
		return F11
	case "F12":
		return F12
	case "F13":
		return F13
	case "F14":
		return F14
	case "F15":
		return F15
	case "Backspace":
		return Backspace
	case "Delete":
		return Delete
	case "Insert":
		return Insert
	case "Home":
		return Home
	case "End":
		return End
	case "PageUp":
		return PageUp
	case "PageDown":
		return PageDown
	case "Up":
		return Up
	case "Down":
		return Down
	case "Left":
		return Left
	case "Right":
		return Right
	default:
		return None
	}
}

// Press represents a key press.
//
// Only one of Key or Ch is set.
type Press struct {
	Alt  bool
	Ctrl bool
	Key  Key  // Non-character key (e.g. F-keys, arrows, etc). Set to None when not used.
	Ch   rune // Character key, e.g. letter, number. Set to rune(0) when not used.
}

func (k Press) String() string {
	if k.Key == None && k.Ch == 0 {
		return ""
	}
	out := ""
	if k.Ctrl {
		out += "Ctrl-"
	}
	if k.Alt {
		out += "Alt-"
	}
	if k.Key > None && k.Key < last {
		out += k.Key.String()
	} else {
		out += string(k.Ch)
	}
	return out
}

// IsMeta returns true if a key press is a meta key.
func (k Press) IsMeta() bool {
	return k.Alt || k.Ctrl || k.Key >= Meta
}

// IsValid returns true if the object represents a key press.
func (k Press) IsValid() bool {
	return k.Alt || k.Ctrl || k.Key != None || k.Ch != rune(0)
}

// StringToPress parses a string and returns a Press.
func StringToPress(keyName string) Press {
	out := Press{}
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
		out.Key = StringToKey(keyName)
	}
	return out
}
