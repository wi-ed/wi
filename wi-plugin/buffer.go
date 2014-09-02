// Copyright 2014 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package wi

var (
	//Magenta   = RGB{192, 0, 192}
	Black       = RGB{0, 0, 0}
	BrightRed   = RGB{255, 0, 0}
	BrightWhite = RGB{255, 255, 255}
	DarkGray    = RGB{128, 128, 128}
	Red         = RGB{192, 0, 0}
	White       = RGB{192, 192, 192}
)

// RGB represents the color of a single character on screen.
//
// Transparency per character is not supported, this is text mode after all.
type RGB struct {
	R, G, B uint8
}

// Rect is highly inspired by image.Rectangle but uses more standard origin +
// size instead of two points. It makes the usage much simpler. It implements a
// small subset of image.Rectangle.
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
	Ch        rune
	Fg        RGB
	Bg        RGB
	Italic    bool
	Underline bool
	Blinking  bool
}

func MakeCell(Ch rune, Fg, Bg RGB) Cell {
	return Cell{Ch: Ch, Fg: Fg, Bg: Bg}
}

type Buffer struct {
	Width  int
	Height int
	Cells  []Cell
}

// Line returns a single line in the buffer.
func (b *Buffer) Line(Y int) []Cell {
	if Y >= b.Height {
		return []Cell{}
	}
	base := Y * b.Width
	return b.Cells[base : base+b.Width]
}

func (b *Buffer) Set(X, Y int, cell Cell) {
	b.Line(Y)[X] = cell
}

func (b *Buffer) Fill(zone Rect, cell Cell) {
	// TODO(maruel): Bound checking.
	for y := zone.Y; y < zone.Y+zone.Height; y++ {
		line := b.Line(y)
		for x := zone.X; x < zone.X+zone.Width; x++ {
			line[x] = cell
		}
	}
}

// Blit copies src completely into b at offsets X and Y.
func (b *Buffer) Blit(X, Y int, src *Buffer) {
	for y := 0; y < src.Height; y++ {
		copy(b.Line(Y + y)[X:], src.Line(y))
	}
}

func NewBuffer(width, height int) *Buffer {
	return &Buffer{
		width,
		height,
		make([]Cell, width*height),
	}
}
