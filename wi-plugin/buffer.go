// Copyright 2014 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package wi

import (
	"unicode/utf8"
)

var (
	Black         = RGB{0, 0, 0}
	Blue          = RGB{0, 0, 170}
	Green         = RGB{0, 170, 0}
	Cyan          = RGB{0, 170, 170}
	Red           = RGB{170, 0, 0}
	Magenta       = RGB{170, 0, 170}
	Brown         = RGB{170, 85, 0}
	LightGray     = RGB{170, 170, 170}
	DarkGray      = RGB{85, 85, 85}
	BrightBlue    = RGB{85, 85, 255}
	BrightGreen   = RGB{85, 255, 85}
	BrightCyan    = RGB{85, 255, 255}
	BrightRed     = RGB{255, 85, 85}
	BrightMagenta = RGB{255, 85, 255}
	BrightYellow  = RGB{255, 255, 85}
	White         = RGB{255, 255, 255}

	// Use colors in EGAColors for maximum compatibility with terminals that do
	// not support terminal-256.
	//
	// EGA colors <3 forever.
	// Random intertubes websites says things like:
	//   sudo apt-get install ncurses-term
	//   export TERM=xterm-256color
	//
	// "tput colors" can be used to determine the number of colors supported by
	// the terminal.
	EGAColors = []RGB{
		Black,
		Blue,
		Green,
		Cyan,
		Red,
		Magenta,
		Brown,
		LightGray,
		DarkGray,
		BrightBlue,
		BrightGreen,
		BrightCyan,
		BrightRed,
		BrightMagenta,
		BrightYellow,
		White,
	}

	// TODO(maruel): Add all colors for
	// gnome-256color/putty-256color/rxvt-256color/xterm-256color.
	// Debian package ncurses-term gets the DB for colors.
)

// RGB represents the color of a single character on screen.
//
// Transparency per character is not supported, this is text mode after all.
type RGB struct {
	R, G, B uint8
}

// NearestEGAColor returns the nearest colors for a 16 colors terminal.
func NearestEGAColor(c RGB) RGB {
	minDistance := 255 * 255 * 3
	out := Black
	for _, ega := range EGAColors {
		r := int(ega.R) - int(c.R)
		g := int(ega.G) - int(c.G)
		b := int(ega.B) - int(c.B)
		distance := r*r + b*b + g*g
		if distance < minDistance {
			minDistance = distance
			out = ega
		}
	}
	return out
}

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

// Eq reports whether r and s are equal.
func (r Rect) Eq(s Rect) bool {
	return r.X == s.X && r.Y == s.Y && r.Width == s.Width && r.Height == s.Height
}

// In reports whether every point in r is in s.
func (r Rect) In(s Rect) bool {
	if r.Empty() {
		return true
	}
	return s.X <= r.X && (r.X+r.Width) <= (s.X+s.Width) && s.Y <= r.Y && (r.Y+r.Height) <= (s.Y+s.Height)
}

// Cell represents the properties of a single character on screen.
//
// Some properties are ignored on different terminals.
type Cell struct {
	R rune
	F CellFormat
}

type CellFormat struct {
	Fg        RGB
	Bg        RGB
	Italic    bool
	Underline bool
	Blinking  bool
}

// MakeCell is a shorthand to return a Cell.
func MakeCell(R rune, Fg, Bg RGB) Cell {
	return Cell{R, CellFormat{Fg: Fg, Bg: Bg}}
}

// Buffer represents a buffer of Cells.
//
// The Cells slice can be shared across multiple Buffer when using SubBuffer().
// Width Height are guaranteed to be either both zero or non-zero.
type Buffer struct {
	Width  int
	Height int
	Stride int
	Cells  []Cell
}

var emptySlice = []Cell{}

// Line returns a single line in the buffer.
//
// If the requested line number if outside the buffer, an empty slice is
// returned.
func (b *Buffer) Line(Y int) []Cell {
	if Y >= b.Height {
		return emptySlice
	}
	base := Y * b.Stride
	return b.Cells[base : base+b.Width]
}

// Get gets a specific character cell.
//
// If the position is outside the buffer, an empty cell is returned.
func (b *Buffer) Get(X, Y int) Cell {
	line := b.Line(Y)
	if len(line) < X {
		return Cell{}
	}
	return line[X]
}

// Set sets a specific character cell.
//
// If the position is outside the buffer, the call is ignored.
func (b *Buffer) Set(X, Y int, cell Cell) {
	line := b.Line(Y)
	if len(line) < X {
		line[X] = cell
	}
}

// DrawString draws a string into the buffer.
//
// Text will be automatically elided if necessary.
func (b *Buffer) DrawString(s string, X, Y int, f CellFormat) {
	line := b.Line(Y)
	if len(line) < X {
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

// ElideText elide a string as necessary.
func ElideText(s string, width int) string {
	if width <= 0 {
		return ""
	}
	// TODO(maruel): Memory copy intensive.
	length := utf8.RuneCount([]byte(s))
	if length < width {
		return s
	}
	return s[:length-1] + "…"
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
	base := r.Y*b.Stride + r.X
	length := r.Height*b.Stride + r.Width
	if r.Width <= 0 || r.Height <= 0 {
		r.Width = 0
		r.Height = 0
		base = 0
		length = 0
	}
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
		make([]Cell, width*height),
	}
}
