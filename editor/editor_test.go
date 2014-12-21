// Copyright 2014 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package editor

import (
	"io/ioutil"
	"log"
	"testing"

	"github.com/maruel/ut"
	"github.com/maruel/wi/wiCore"
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

	wiCore.PostCommand(editor, "editor_bootstrap_ui")
	wiCore.PostCommand(editor, "new")
	// Supporting this command requires using "go test -tags debug"
	// wiCore.PostCommand(editor, "log_all")
	wiCore.PostCommand(editor, "editor_quit")
	ut.AssertEqual(t, 0, editor.EventLoop())

	expected := wiCore.NewBuffer(80, 25)
	expected.Fill(wiCore.MakeCell(' ', wiCore.BrightYellow, wiCore.Black))
	expected.DrawString("Dummy content", 0, 0, wiCore.CellFormat{Fg: wiCore.BrightYellow, Bg: wiCore.Black})
	expected.DrawString("Really", 0, 1, wiCore.CellFormat{Fg: wiCore.BrightYellow, Bg: wiCore.Black})
	expected.DrawString("Status Name    Status Root                                       Status Position", 0, 24, wiCore.CellFormat{Fg: wiCore.Red, Bg: wiCore.LightGray})
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

	wiCore.PostCommand(editor, "editor_bootstrap_ui")
	wiCore.PostCommand(editor, "invalid")
	wiCore.PostCommand(editor, "editor_quit")
	ut.AssertEqual(t, 0, editor.EventLoop())

	expected := wiCore.NewBuffer(80, 25)
	expected.Fill(wiCore.MakeCell(' ', wiCore.Red, wiCore.Black))
	expected.DrawString("Root", 0, 0, wiCore.CellFormat{Fg: wiCore.Red, Bg: wiCore.Black})
	expected.DrawString("Status Name    Status Root                                       Status Position", 0, 24, wiCore.CellFormat{Fg: wiCore.Red, Bg: wiCore.LightGray})
	ut.AssertEqual(t, len(expected.Cells), len(terminal.Buffer.Cells))
	for i := 0; i < len(expected.Cells); i++ {
		ut.AssertEqualIndex(t, i, expected.Cells[i], terminal.Buffer.Cells[i])
	}
}
