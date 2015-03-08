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
	"github.com/maruel/wi/wicore/colors"
	"github.com/maruel/wi/wicore/raster"
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
		_ = out.Close()
	}
}

func compareBuffers(t *testing.T, expected *raster.Buffer, actual *raster.Buffer) {
	ut.AssertEqual(t, expected.Height, actual.Height)
	ut.AssertEqual(t, expected.Width, actual.Width)
	// First compares lines of text, then colors.
	for l := 0; l < expected.Height; l++ {
		e := expected.Line(l)
		a := actual.Line(l)
		ut.AssertEqualIndex(t, l, string(e.Runes()), string(a.Runes()))
		ut.AssertEqualIndex(t, l, e.Formats(), a.Formats())
	}
}

func TestMainImmediateQuit(t *testing.T) {
	defer keepLog(t)()

	terminal := NewTerminalFake(80, 25, []TerminalEvent{})
	editor, err := MakeEditor(terminal, true)
	ut.AssertEqual(t, nil, err)
	defer func() {
		_ = editor.Close()
	}()

	wicore.PostCommand(editor, nil, "editor_bootstrap_ui")
	wicore.PostCommand(editor, nil, "new")
	// Supporting this command requires using "go test -tags debug"
	// wicore.PostCommand(editor, nil, "log_all")
	wicore.PostCommand(editor, nil, "editor_quit")
	ut.AssertEqual(t, 0, editor.EventLoop())

	expected := raster.NewBuffer(80, 25)
	expected.Fill(raster.MakeCell(' ', colors.BrightYellow, colors.Black))
	expected.DrawString("Dummy content", 0, 0, raster.CellFormat{Fg: colors.BrightYellow, Bg: colors.Black})
	expected.DrawString("Really", 0, 1, raster.CellFormat{Fg: colors.BrightYellow, Bg: colors.Black})
	expected.DrawString("Status Name    Normal                                            0,0            ", 0, 24, raster.CellFormat{Fg: colors.Red, Bg: colors.LightGray})
	expected.Cell(0, 0).F.Bg = colors.White
	expected.Cell(0, 0).F.Fg = colors.Black
	compareBuffers(t, expected, terminal.Buffer)
}

func TestMainInvalidThenQuit(t *testing.T) {
	defer keepLog(t)()

	terminal := NewTerminalFake(80, 25, []TerminalEvent{})
	editor, err := MakeEditor(terminal, true)
	ut.AssertEqual(t, nil, err)
	defer func() {
		_ = editor.Close()
	}()

	wicore.PostCommand(editor, nil, "editor_bootstrap_ui")
	wicore.PostCommand(editor, nil, "invalid")
	wicore.PostCommand(editor, nil, "editor_quit")
	ut.AssertEqual(t, 0, editor.EventLoop())

	expected := raster.NewBuffer(80, 25)
	expected.Fill(raster.MakeCell(' ', colors.Red, colors.Black))
	expected.DrawString("Root", 0, 0, raster.CellFormat{Fg: colors.Red, Bg: colors.Black})
	expected.DrawString("Status Name    Normal                                            Status Position", 0, 24, raster.CellFormat{Fg: colors.Red, Bg: colors.LightGray})
	compareBuffers(t, expected, terminal.Buffer)
}
