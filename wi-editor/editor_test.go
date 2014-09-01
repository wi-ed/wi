// Copyright 2014 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package editor

import (
	"github.com/maruel/tulib"
	"github.com/maruel/wi/wi-plugin"
	"github.com/nsf/termbox-go"
	"log"
	"testing"
)

type nullWriter int

func (nullWriter) Write([]byte) (int, error) {
	return 0, nil
}

func init() {
	// TODO(maruel): This has persistent side-effect. Figure out how to handle
	// "log" properly. Likely by using the same mechanism as used in package
	// "subcommands".
	log.SetOutput(new(nullWriter))
}

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

// TODO(maruel): Add a test with very small display (10x2) and ensure it's
// somewhat usable.

func TestMainImmediateQuit(t *testing.T) {
	t.Parallel()
	editor := MakeEditor(makeTermBoxFake(80, 25, []termbox.Event{}))
	wi.PostCommand(editor, "editor_quit")
	result := Main(true, editor)
	if result != 0 {
		t.Fatalf("Exit code: %v", result)
	}
	// TODO(maruel): Check the content of the cells via termBox.buffer.Cells.
}

func TestMainInvalidThenQuit(t *testing.T) {
	t.Parallel()
	editor := MakeEditor(makeTermBoxFake(80, 25, []termbox.Event{}))
	wi.PostCommand(editor, "invalid")
	wi.PostCommand(editor, "editor_quit")
	result := Main(true, editor)
	if result != 0 {
		t.Fatalf("Exit code: %v", result)
	}
	// TODO(maruel): Check the content of the cells via termBox.buffer.Cells.
	//t.Logf("%v", termBox.buffer.Cells)
	//t.Fail()
}
