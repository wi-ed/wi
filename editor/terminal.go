// Copyright 2013 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package editor

import (
	"github.com/maruel/wi/wicore"
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
	Blit(b *wicore.Buffer)
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
	Key  wicore.KeyPress
	Size Size
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
	Buffer *wicore.Buffer
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
func (t *TerminalFake) Blit(b *wicore.Buffer) {
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
		wicore.NewBuffer(width, height),
	}
}
