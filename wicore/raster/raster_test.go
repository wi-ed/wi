// Copyright 2014 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package raster

import (
	"testing"

	"github.com/maruel/ut"
	"github.com/maruel/wi/wicore/colors"
)

func TestRect(t *testing.T) {
	ut.AssertEqual(t, true, Rect{}.Empty())
	ut.AssertEqual(t, true, Rect{}.In(Rect{}))
	ut.AssertEqual(t, true, Rect{1, 1, 2, 2}.In(Rect{0, 0, 10, 10}))
}

func TestCellFormat(t *testing.T) {
	ut.AssertEqual(t, true, CellFormat{}.Empty())
}

func TestCell(t *testing.T) {
	ut.AssertEqual(t, Cell{R: 'a', F: CellFormat{Fg: colors.Red, Bg: colors.Blue}}, MakeCell('a', colors.Red, colors.Blue))
}

func TestBuffer(t *testing.T) {
	ut.AssertEqual(t, "Buffer(0, 0, 0)", NewBuffer(-1, -1).String())

	b := NewBuffer(4, 4)
	b2 := b.SubBuffer(Rect{1, 2, 1, 1})
	b3 := b.SubBuffer(Rect{3, 3, 2, 2})
	b.Fill(MakeCell('a', colors.Red, colors.Black))
	b.DrawString("FOOO", 0, 0, CellFormat{Fg: colors.Red})
	b.DrawString("foooo", 0, 3, CellFormat{colors.Brown, colors.Magenta, true, true, false})
	// Outside.
	b.DrawString("ZZ", 4, 0, CellFormat{colors.Brown, colors.Magenta, true, true, false})
	// Blit between two subbuffers inside the same buffer.
	b2.Blit(b3)
	b2.Cell(0, 0).F.Fg = colors.Blue

	ut.AssertEqual(t, Cell{}, *b.Cell(4, 4))
	ut.AssertEqual(t, "FOOO", string(b.Line(0).Runes()))
	ut.AssertEqual(t, "aaaa", string(b.Line(1).Runes()))
	ut.AssertEqual(t, "a…aa", string(b.Line(2).Runes()))
	ut.AssertEqual(t, "foo…", string(b.Line(3).Runes()))

	fRed := CellFormat{Fg: colors.Red}
	fBlu := CellFormat{Fg: colors.Blue, Bg: colors.Magenta, Italic: true, Underline: true}
	fBro := CellFormat{Fg: colors.Brown, Bg: colors.Magenta, Italic: true, Underline: true}
	ut.AssertEqual(t, []CellFormat{fRed, fRed, fRed, fRed}, b.Line(0).Formats())
	ut.AssertEqual(t, []CellFormat{fRed, fRed, fRed, fRed}, b.Line(1).Formats())
	ut.AssertEqual(t, []CellFormat{fRed, fBlu, fRed, fRed}, b.Line(2).Formats())
	ut.AssertEqual(t, []CellFormat{fBro, fBro, fBro, fBro}, b.Line(3).Formats())

	ut.AssertEqual(t, Buffer{Cells: CellStride{}}, *b.SubBuffer(Rect{0, 0, -1, -1}))

	NewBuffer(0, 0).Fill(MakeCell('a', colors.Red, colors.Blue))
}

func TestFormatText(t *testing.T) {
	data := [][]string{
		{"hello", "hello"},
		{"\000hello", "NULhello"},
		{"\001", "^A"},
		{"	a", "	a"},
	}
	for i, v := range data {
		ut.AssertEqualIndex(t, i, v[1], FormatText(v[0]))
	}
}

func TestElideText(t *testing.T) {
	ut.AssertEqual(t, "", ElideText("foo", -1))
	ut.AssertEqual(t, "", ElideText("foo", 0))

	data := [][]string{
		{"hel", "hel"},
		{"hell", "he…"},
		{"hello", "he…"},
		{"\000hello", "\000h…"},
	}
	for i, v := range data {
		ut.AssertEqualIndex(t, i, v[1], ElideText(v[0], 3))
	}
}
