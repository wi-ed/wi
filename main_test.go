// Copyright 2014 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package main

import (
	"github.com/maruel/tulib"
	"github.com/nsf/termbox-go"
	"log"
	"testing"
)

type termBoxFake struct {
	X      int
	Y      int
	events []termbox.Event
	buffer tulib.Buffer
}

func (t *termBoxFake) Size() (int, int) {
	return t.X, t.Y
}

func (t *termBoxFake) Flush() {
}

func (t *termBoxFake) PollEvent() termbox.Event {
	if len(t.events) == 0 {
		select {}
	}
	e := t.events[0]
	t.events = t.events[1:]
	return e
}

func (t *termBoxFake) Buffer() tulib.Buffer {
	return t.buffer
}

func makeTermBoxFake(width, height int, events []termbox.Event) *termBoxFake {
	return &termBoxFake{
		width,
		height,
		events,
		tulib.NewBuffer(width, height),
	}
}

func TestInnerMain(t *testing.T) {
	// TODO(maruel): This has persistent side-effect. Figure out how to handle
	// "log" properly. Likely by using the same mechanism as used in package
	// "subcommands".
	log.SetOutput(new(nullWriter))

	// TODO(maruel): Add a test with very small display (10x2) and ensure it's
	// somewhat usable.
	termBox := makeTermBoxFake(80, 25, []termbox.Event{})
	result := innerMain(true, true, []string{"quit"}, termBox)
	if result != 0 {
		t.Fatalf("Exit code: %v", result)
	}

	// TODO(maruel): Check the content of the cells via termBox.buffer.Cells.
}
