// Copyright 2014 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package editor

import (
	"github.com/maruel/wi/wi-plugin"
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

// TODO(maruel): Add a test with small display (10x2) and ensure it's somewhat
// usable.

func TestMainImmediateQuit(t *testing.T) {
	t.Parallel()
	editor := MakeEditor(NewTerminalFake(80, 25, []TerminalEvent{}))
	wi.PostCommand(editor, "editor_quit")
	result := Main(true, editor)
	if result != 0 {
		t.Fatalf("Exit code: %v", result)
	}
	// TODO(maruel): Check the content of the cells via termBox.buffer.Cells.
}

func TestMainInvalidThenQuit(t *testing.T) {
	t.Parallel()
	editor := MakeEditor(NewTerminalFake(80, 25, []TerminalEvent{}))
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
