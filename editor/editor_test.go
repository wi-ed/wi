// Copyright 2014 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package editor

import (
	"log"
	"testing"

	"github.com/maruel/ut"
	"github.com/maruel/wi/wiCore"
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

// TODO(maruel): Add a test with small display (10x2) and ensure it's somewhat
// usable.

func TestMainImmediateQuit(t *testing.T) {
	t.Parallel()
	terminal := NewTerminalFake(80, 25, []TerminalEvent{})
	editor, err := MakeEditor(terminal, true)
	ut.AssertEqual(t, nil, err)
	defer editor.Close()
	wiCore.PostCommand(editor, "editor_quit")
	ut.AssertEqual(t, 0, editor.EventLoop())
	// TODO(maruel): Print something.
	for y := 0; y < terminal.Height; y++ {
		for x := 0; x < terminal.Width; x++ {
			c := terminal.Buffer.Get(x, y)
			ut.AssertEqual(t, '\u0000', c.R)
		}
	}
}

func TestMainInvalidThenQuit(t *testing.T) {
	t.Parallel()
	terminal := NewTerminalFake(80, 25, []TerminalEvent{})
	editor, err := MakeEditor(terminal, true)
	ut.AssertEqual(t, nil, err)
	defer editor.Close()
	wiCore.PostCommand(editor, "invalid")
	wiCore.PostCommand(editor, "editor_quit")
	ut.AssertEqual(t, 0, editor.EventLoop())
	// TODO(maruel): Print something.
	for y := 0; y < terminal.Height; y++ {
		for x := 0; x < terminal.Width; x++ {
			c := terminal.Buffer.Get(x, y)
			ut.AssertEqual(t, '\u0000', c.R)
		}
	}
}
