// Copyright 2013 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package editor

import (
	"github.com/maruel/wi/wi-plugin"
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

type EventType int

const (
	EventKey = iota
	EventResize
)

// TerminalEvent represents an event that occured on the terminal.
type TerminalEvent struct {
	Type EventType // Type determines which other member will be valid for this event.
	Key  Key
	Size Size
}

// Key represents a key press.
type Key struct {
	// TODO(maruel): Redo.
	Modifier int
	Key      int
	Ch       rune
}

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
	X      int
	Y      int
	events []TerminalEvent
	buffer *wi.Buffer
}

func (t *TerminalFake) Size() (int, int) {
	return t.X, t.Y
}

func (t *TerminalFake) SeedEvents() <-chan TerminalEvent {
	out := make(chan TerminalEvent)
	go func() {
		for _, i := range t.events {
			out <- i
		}
	}()
	return out
}

func (t *TerminalFake) Blit(b *wi.Buffer) {
	t.buffer.Blit(b)
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
