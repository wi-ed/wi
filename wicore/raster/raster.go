// Copyright 2014 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package raster implements text buffering.
package raster

import (
	"fmt"
	"unicode/utf8"

	"github.com/maruel/wi/wicore/colors"
)

// Rect is highly inspired by image.Rectangle but uses more standard origin +
// size instead of two points. It makes the usage much simpler. It implements a
// small subset of image.Rectangle.
//
// Negative values are invalid.
type Rect struct {
	X, Y, Width, Height int
}

// Empty reports whether the rectangle area is 0.
func (r Rect) Empty() bool {
	return r.Width == 0 || r.Height == 0
}

// In reports whether every point in r is in s.
func (r Rect) In(s Rect) bool {
	if r.Empty() {
		return true
	}
	return s.X <= r.X && (r.X+r.Width) <= (s.X+s.Width) && s.Y <= r.Y && (r.Y+r.Height) <= (s.Y+s.Height)
}

// CellFormat describes all the properties of a single cell on screen.
type CellFormat struct {
	Fg        colors.RGB
	Bg        colors.RGB
	Italic    bool
	Underline bool
	Blinking  bool
}

// Empty reports whether the CellFormat is black on black. In that case it
// doesn't matter if it's italic or underlined.
func (c CellFormat) Empty() bool {
	return c.Fg == colors.Black && c.Bg == colors.Black
}

// Cell represents the properties of a single character on screen.
//
// Some properties are ignored on different terminals.
type Cell struct {
	R rune
	F CellFormat
}

// MakeCell is a shorthand to return a Cell.
func MakeCell(R rune, Fg, Bg colors.RGB) Cell {
	return Cell{R, CellFormat{Fg: Fg, Bg: Bg}}
}

// CellStride is a slice of cells.
type CellStride []Cell

// Runes returns runes as a slice.
func (c CellStride) Runes() []rune {
	out := make([]rune, len(c))
	for i, cell := range c {
		out[i] = cell.R
	}
	return out
}

// Formats returns cells format as a slice.
func (c CellStride) Formats() []CellFormat {
	out := make([]CellFormat, len(c))
	for i, cell := range c {
		out[i] = cell.F
	}
	return out
}

// Buffer represents a buffer of Cells.
//
// The Cells slice can be shared across multiple Buffer when using SubBuffer().
// Width Height are guaranteed to be either both zero or non-zero.
type Buffer struct {
	Width  int
	Height int
	Stride int
	Cells  CellStride
}

var emptySlice = CellStride{}

func (b *Buffer) String() string {
	return fmt.Sprintf("Buffer(%d, %d, %d)", b.Width, b.Height, b.Stride)
}

// Line returns a single line in the buffer.
//
// If the requested line number if outside the buffer, an empty slice is
// returned.
func (b *Buffer) Line(Y int) CellStride {
	if Y >= b.Height {
		return emptySlice
	}
	base := Y * b.Stride
	return b.Cells[base : base+b.Width]
}

// Cell returns the pointer to a specific character cell.
//
// If the position is outside the buffer, an empty temporary cell is returned.
func (b *Buffer) Cell(X, Y int) *Cell {
	line := b.Line(Y)
	if len(line) < X {
		return &Cell{}
	}
	return &line[X]
}

// DrawString draws a string into the buffer.
//
// Text will be automatically elided if necessary.
func (b *Buffer) DrawString(s string, X, Y int, f CellFormat) {
	line := b.Line(Y)
	if len(line) <= X {
		return
	}
	bytes := []byte(ElideText(s, len(line)-X))
	for x := X; x < len(line) && len(bytes) > 0; x++ {
		r, size := utf8.DecodeRune(bytes)
		line[x].R = r
		line[x].F = f
		bytes = bytes[size:]
	}
}

// Fill fills a buffer with a Cell.
//
// To fill a section of a buffer, use SubBuffer() first.
func (b *Buffer) Fill(cell Cell) {
	if b.Height == 0 {
		return
	}
	// First set the initial line.
	line0 := b.Line(0)
	for x := 0; x < b.Width; x++ {
		line0[x] = cell
	}
	// Then used optimized copy() to fill the rest.
	for y := 1; y < b.Height; y++ {
		copy(b.Line(y), line0)
	}
}

// FormatText formats special characters like code points below 32.
//
// TODO(maruel): This must add coloring too.
//
// TODO(maruel): Improve performance for the common case (no special character).
//
// TODO(maruel): Handles special unicode whitespaces. Since the editor is meant
// for mono-space font, all except U+0020 and \t should be escaped.
// https://en.wikipedia.org/wiki/Whitespace_character
func FormatText(s string) string {
	out := ""
	for _, c := range s {
		if c == 0 {
			out += "NUL"
		} else if c == 9 {
			// TODO(maruel): Need positional information AND desired tabwidth.
			out += string(c)
		} else if c <= 32 {
			out += "^" + string(c+'A'-1)
		} else {
			out += string(c)
		}
	}
	return out
}

// ElideText elide a string as necessary.
func ElideText(s string, width int) string {
	if width <= 0 {
		return ""
	}
	length := utf8.RuneCountInString(s)
	if length <= width {
		return s
	}
	return s[:width-1] + "…"
}

// Blit copies src into b.
//
// To copy a section of a buffer, use SubBuffer() first. Areas that falls
// either outside of src or of b are ignored.
func (b *Buffer) Blit(src *Buffer) {
	for y := 0; y < src.Height; y++ {
		copy(b.Line(y), src.Line(y))
	}
}

// SubBuffer returns a Buffer representing a section of the buffer, sharing the
// same cells.
func (b *Buffer) SubBuffer(r Rect) *Buffer {
	if r.X+r.Width > b.Width {
		r.Width = b.Width - r.X
	}
	if r.Y+r.Height > b.Height {
		r.Height = b.Height - r.Y
	}
	if r.Width <= 0 || r.Height <= 0 {
		return &Buffer{Cells: CellStride{}}
	}
	base := r.Y*b.Stride + r.X
	length := r.Height*b.Stride + r.Width - b.Width
	return &Buffer{
		r.Width,
		r.Height,
		b.Stride,
		b.Cells[base : base+length],
	}
}

// NewBuffer creates a fresh new buffer.
func NewBuffer(width, height int) *Buffer {
	if width <= 0 || height <= 0 {
		width = 0
		height = 0
	}
	return &Buffer{
		width,
		height,
		width,
		make(CellStride, width*height),
	}
}
