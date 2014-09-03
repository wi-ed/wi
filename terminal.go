// Copyright 2014 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package main

import (
	"github.com/maruel/wi/wi-editor"
	"github.com/maruel/wi/wi-plugin"
	"github.com/nsf/termbox-go"
)

// TermBox implements the editor.Terminal interface that interacts with termbox.
type TermBox struct {
}

func (t TermBox) Size() (int, int) {
	return termbox.Size()
}

func (t TermBox) SeedEvents() <-chan editor.TerminalEvent {
	// Converts termbox.Event into editor.TerminalEvent. This removes the need to
	// have an hard dependency of wi-editor on termbox-go; this makes both unit
	// testing easier and future-proof the editor.
	c := make(chan editor.TerminalEvent)
	go func() {
		for {
			e := termbox.PollEvent()
			switch e.Type {
			case termbox.EventKey:
				// TODO(maruel): Key translation.
				c <- editor.TerminalEvent{
					Type: editor.EventKey,
					Key:  editor.Key{},
				}
			case termbox.EventResize:
				c <- editor.TerminalEvent{
					Type: editor.EventKey,
					Size: editor.Size{e.Width, e.Height},
				}
			case termbox.EventError:
				break
			}
		}
		close(c)
	}()
	return c
}

// Converts a RGB color into the nearest termbox color.
func rgbToTermBox(c wi.RGB) termbox.Attribute {
	switch wi.NearestEGAColor(c) {
	case wi.Black:
		return termbox.ColorBlack
	case wi.Blue:
		return termbox.ColorBlue
	case wi.Green:
		return termbox.ColorGreen
	case wi.Cyan:
		return termbox.ColorCyan
	case wi.Red:
		return termbox.ColorRed
	case wi.Magenta:
		return termbox.ColorMagenta
	case wi.Brown:
		return termbox.ColorYellow
	case wi.LightGray:
		return termbox.ColorWhite
	case wi.DarkGray:
		return termbox.ColorBlack | termbox.AttrBold
	case wi.BrightBlue:
		return termbox.ColorBlue | termbox.AttrBold
	case wi.BrightGreen:
		return termbox.ColorGreen | termbox.AttrBold
	case wi.BrightCyan:
		return termbox.ColorCyan | termbox.AttrBold
	case wi.BrightRed:
		return termbox.ColorRed | termbox.AttrBold
	case wi.BrightMagenta:
		return termbox.ColorMagenta | termbox.AttrBold
	case wi.BrightYellow:
		return termbox.ColorYellow | termbox.AttrBold
	case wi.White:
		return termbox.ColorWhite | termbox.AttrBold
	default:
		return termbox.ColorDefault
	}
}

// Convert the wi.Buffer format into termbox format.
func (t TermBox) Blit(b *wi.Buffer) {
	width, height := termbox.Size()
	cells := termbox.CellBuffer()
	if width > b.Width {
		width = b.Width
	}
	if height > b.Height {
		height = b.Height
	}

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			i := y*width + x
			cell := b.Get(x, y)
			cells[i].Ch = cell.R
			cells[i].Fg = rgbToTermBox(cell.F.Fg)
			// TODO(maruel): Not sure.
			if cell.F.Underline {
				cells[i].Fg |= termbox.AttrUnderline
			}
			cells[i].Bg = rgbToTermBox(cell.F.Bg)
			// TODO(maruel): Not sure. Some terminal may cause Bg&Bold to be Blinking.
			if cell.F.Italic {
				cells[i].Bg |= termbox.AttrUnderline
			}
		}
	}
	if err := termbox.Flush(); err != nil {
		panic(err)
	}
}
