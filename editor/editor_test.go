// Copyright 2014 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package editor

import (
	"io/ioutil"
	"log"
	"testing"

	"github.com/maruel/ut"
	"github.com/maruel/wi/wicore"
)

func init() {
	// TODO(maruel): This has persistent side-effect. Figure out how to handle
	// "log" properly. Likely by using the same mechanism as used in package
	// "subcommands".
	log.SetOutput(ioutil.Discard)
}

// TODO(maruel): Add a test with small display (10x2) and ensure it's somewhat
// usable.

func keepLog(t *testing.T) func() {
	out := ut.NewWriter(t)
	log.SetOutput(out)
	return func() {
		log.SetOutput(ioutil.Discard)
		out.Close()
	}
}

func TestMainImmediateQuit(t *testing.T) {
	defer keepLog(t)()

	terminal := NewTerminalFake(80, 25, []TerminalEvent{})
	editor, err := MakeEditor(terminal, true)
	ut.AssertEqual(t, nil, err)
	defer editor.Close()

	wicore.PostCommand(editor, nil, "editor_bootstrap_ui")
	wicore.PostCommand(editor, nil, "new")
	// Supporting this command requires using "go test -tags debug"
	// wicore.PostCommand(editor, nil, "log_all")
	wicore.PostCommand(editor, nil, "editor_quit")
	ut.AssertEqual(t, 0, editor.EventLoop())

	expected := wicore.NewBuffer(80, 25)
	expected.Fill(wicore.MakeCell(' ', wicore.BrightYellow, wicore.Black))
	expected.DrawString("Dummy content", 0, 0, wicore.CellFormat{Fg: wicore.BrightYellow, Bg: wicore.Black})
	expected.DrawString("Really", 0, 1, wicore.CellFormat{Fg: wicore.BrightYellow, Bg: wicore.Black})
	expected.DrawString("Status Name    Status Mode                                       Status Position", 0, 24, wicore.CellFormat{Fg: wicore.Red, Bg: wicore.LightGray})
	ut.AssertEqual(t, len(expected.Cells), len(terminal.Buffer.Cells))
	for i := 0; i < len(expected.Cells); i++ {
		ut.AssertEqualIndex(t, i, expected.Cells[i], terminal.Buffer.Cells[i])
	}
}

func TestMainInvalidThenQuit(t *testing.T) {
	defer keepLog(t)()

	terminal := NewTerminalFake(80, 25, []TerminalEvent{})
	editor, err := MakeEditor(terminal, true)
	ut.AssertEqual(t, nil, err)
	defer editor.Close()

	wicore.PostCommand(editor, nil, "editor_bootstrap_ui")
	wicore.PostCommand(editor, nil, "invalid")
	wicore.PostCommand(editor, nil, "editor_quit")
	ut.AssertEqual(t, 0, editor.EventLoop())

	expected := wicore.NewBuffer(80, 25)
	expected.Fill(wicore.MakeCell(' ', wicore.Red, wicore.Black))
	expected.DrawString("Root", 0, 0, wicore.CellFormat{Fg: wicore.Red, Bg: wicore.Black})
	expected.DrawString("Status Name    Status Mode                                       Status Position", 0, 24, wicore.CellFormat{Fg: wicore.Red, Bg: wicore.LightGray})
	ut.AssertEqual(t, len(expected.Cells), len(terminal.Buffer.Cells))
	for i := 0; i < len(expected.Cells); i++ {
		ut.AssertEqualIndex(t, i, expected.Cells[i], terminal.Buffer.Cells[i])
	}
}
