// Copyright 2013 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package editor

import (
	"github.com/maruel/wi/wi_core"
)

// Terminal is the interface to the actual terminal termbox so it can be mocked
// in unit test or a different implementation than termbox can be used.
type Terminal interface {
	// Size returns the current size of the terminal window.
	Size() (int, int)
	// SeedEvents() returns a channel where events will be sent to.
	//
	// The channel will be closed when the terminal is closed.
	SeedEvents() <-chan TerminalEvent
	// Blit updates the terminal output with the buffer specified.
	//
	// It is important for the buffer to be the right size, otherwise the display
	// will be partially updated.
	Blit(b *wi.Buffer)
}

// EventType is the type of supported event.
type EventType int

// Supported event types.
const (
	EventKey = iota
	EventResize
)

// TerminalEvent represents an event that occured on the terminal.
type TerminalEvent struct {
	Type EventType // Type determines which other member will be valid for this event.
	Key  KeyPress
	Size Size
}

// Key represents a non-character key.
type Key int

// Known non-character keys.
const (
	KeyNone = iota
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
	switch k.Key {
	case KeyNone:
		if k.Ch == ' ' {
			out += "Space"
		} else {
			out += string(k.Ch)
		}
	case KeyF1:
		out += "F1"
	case KeyF2:
		out += "F2"
	case KeyF3:
		out += "F3"
	case KeyF4:
		out += "F4"
	case KeyF5:
		out += "F5"
	case KeyF6:
		out += "F6"
	case KeyF7:
		out += "F7"
	case KeyF8:
		out += "F8"
	case KeyF9:
		out += "F9"
	case KeyF10:
		out += "F10"
	case KeyF11:
		out += "F11"
	case KeyF12:
		out += "F12"
	case KeyF13:
		out += "F13"
	case KeyF14:
		out += "F14"
	case KeyF15:
		out += "F15"
	case KeyEscape:
		out += "Escape"
	case KeyBackspace:
		out += "Backspace"
	case KeyTab:
		out += "Tab"
	case KeyEnter:
		out += "Enter"
	case KeyInsert:
		out += "Insert"
	case KeyDelete:
		out += "Delete"
	case KeyHome:
		out += "Home"
	case KeyEnd:
		out += "End"
	case KeyArrowUp:
		out += "ArrowUp"
	case KeyArrowDown:
		out += "ArrowDown"
	case KeyArrowLeft:
		out += "ArrowLeft"
	case KeyArrowRight:
		out += "ArrowRight"
	case KeyPageUp:
		out += "PageUp"
	case KeyPageDown:
		out += "PageDown"
	default:
		out += "Unknown"
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

// Size represents the size of an UI element.
type Size struct {
	Width  int
	Height int
}

// Logger is the interface to log to. It must be used instead of
// log.Logger.Printf() or testing.T.Log(). This permits to collect logs for a
// complete test case.
//
// TODO(maruel): Move elsewhere.
type Logger interface {
	Logf(format string, v ...interface{})
}

// TerminalFake implements the Terminal and buffers the output.
//
// It is mostly useful in unit tests.
type TerminalFake struct {
	Width  int
	Height int
	Events []TerminalEvent
	Buffer *wi.Buffer
}

// Size implements Terminal.
func (t *TerminalFake) Size() (int, int) {
	return t.Width, t.Height
}

// SeedEvents implements Terminal.
func (t *TerminalFake) SeedEvents() <-chan TerminalEvent {
	out := make(chan TerminalEvent)
	go func() {
		for _, i := range t.Events {
			out <- i
		}
	}()
	return out
}

// Blit implements Terminal.
func (t *TerminalFake) Blit(b *wi.Buffer) {
	t.Buffer.Blit(b)
}

// NewTerminalFake returns an initialized TerminalFake which implements the
// interface Terminal.
//
// The terminal can be preloaded with fake events.
func NewTerminalFake(width, height int, events []TerminalEvent) *TerminalFake {
	return &TerminalFake{
		width,
		height,
		events,
		wi.NewBuffer(width, height),
	}
}
