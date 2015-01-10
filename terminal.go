// Copyright 2014 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// This file implements the conversion of editor.Terminal to termbox's.

package main

import (
	"fmt"

	"github.com/maruel/wi/editor"
	"github.com/maruel/wi/wicore"
	"github.com/maruel/wi/wicore/colors"
	"github.com/maruel/wi/wicore/key"
	"github.com/maruel/wi/wicore/raster"
	"github.com/nsf/termbox-go"
)

// TermBox implements the editor.Terminal interface that interacts with termbox.
type TermBox struct {
}

// Size implements editor.Terminal.
func (t TermBox) Size() (int, int) {
	return termbox.Size()
}

// SeedEvents implements editor.Terminal.
func (t TermBox) SeedEvents() <-chan editor.TerminalEvent {
	// Converts termbox.Event into editor.TerminalEvent. This removes the need to
	// have an hard dependency of editor on termbox-go; this makes both unit
	// testing easier and future-proof the editor.
	c := make(chan editor.TerminalEvent)
	wicore.Go("SeedEvents", func() {
		for {
			e := termbox.PollEvent()
			switch e.Type {
			case termbox.EventKey:
				// TODO(maruel): Key translation.
				c <- editor.TerminalEvent{
					Type: editor.EventKey,
					Key:  termboxKeyToKeyPress(e),
				}
			case termbox.EventResize:
				c <- editor.TerminalEvent{
					Type: editor.EventKey,
					Size: editor.Size{Width: e.Width, Height: e.Height},
				}
			case termbox.EventError:
				close(c)
				return
			}
		}
	})
	return c
}

// termboxKeyToKeyPress returns the key.Press compatible event.
func termboxKeyToKeyPress(k termbox.Event) key.Press {
	out := key.Press{}
	if k.Mod&termbox.ModAlt != 0 {
		out.Alt = true
	}
	switch termbox.Key(k.Key) {
	case termbox.KeyF1:
		out.Key = key.F1
	case termbox.KeyF2:
		out.Key = key.F2
	case termbox.KeyF3:
		out.Key = key.F3
	case termbox.KeyF4:
		out.Key = key.F4
	case termbox.KeyF5:
		out.Key = key.F5
	case termbox.KeyF6:
		out.Key = key.F6
	case termbox.KeyF7:
		out.Key = key.F7
	case termbox.KeyF8:
		out.Key = key.F8
	case termbox.KeyF9:
		out.Key = key.F9
	case termbox.KeyF10:
		out.Key = key.F10
	case termbox.KeyF11:
		out.Key = key.F11
	case termbox.KeyF12:
		out.Key = key.F12
	case termbox.KeyInsert:
		out.Key = key.Insert
	case termbox.KeyDelete:
		out.Key = key.Delete
	case termbox.KeyHome:
		out.Key = key.Home
	case termbox.KeyEnd:
		out.Key = key.End
	case termbox.KeyPgup:
		out.Key = key.PageUp
	case termbox.KeyPgdn:
		out.Key = key.PageDown
	case termbox.KeyArrowUp:
		out.Key = key.Up
	case termbox.KeyArrowDown:
		out.Key = key.Down
	case termbox.KeyArrowLeft:
		out.Key = key.Left
	case termbox.KeyArrowRight:
		out.Key = key.Right

	case termbox.KeyCtrlSpace: // KeyCtrlTilde, KeyCtrl2
		// This value is 0, which cannot be distinguished from non-keypress.
		if k.Ch == 0 {
			out.Ctrl = true
			out.Key = key.Space
		} else {
			// Normal keypress code path.
			out.Ch = k.Ch
		}

	case termbox.KeyCtrlA:
		out.Ctrl = true
		out.Ch = 'a'
	case termbox.KeyCtrlB:
		out.Ctrl = true
		out.Ch = 'b'
	case termbox.KeyCtrlC:
		out.Ctrl = true
		out.Ch = 'c'
	case termbox.KeyCtrlD:
		out.Ctrl = true
		out.Ch = 'd'
	case termbox.KeyCtrlE:
		out.Ctrl = true
		out.Ch = 'e'
	case termbox.KeyCtrlF:
		out.Ctrl = true
		out.Ch = 'f'
	case termbox.KeyCtrlG:
		out.Ctrl = true
		out.Ch = 'g'
	case termbox.KeyBackspace: // KeyCtrlH
	case termbox.KeyBackspace2:
		out.Key = key.Backspace
	case termbox.KeyTab: // KeyCtrlI
		out.Key = key.Tab
	case termbox.KeyCtrlJ:
		out.Ctrl = true
		out.Ch = 'j'
	case termbox.KeyCtrlK:
		out.Ctrl = true
		out.Ch = 'k'
	case termbox.KeyCtrlL:
		out.Ctrl = true
		out.Ch = 'l'
	case termbox.KeyEnter: // KeyCtrlM
		out.Key = key.Enter
	case termbox.KeyCtrlN:
		out.Ctrl = true
		out.Ch = 'n'
	case termbox.KeyCtrlO:
		out.Ctrl = true
		out.Ch = 'o'
	case termbox.KeyCtrlP:
		out.Ctrl = true
		out.Ch = 'p'
	case termbox.KeyCtrlQ:
		out.Ctrl = true
		out.Ch = 'q'
	case termbox.KeyCtrlR:
		out.Ctrl = true
		out.Ch = 'r'
	case termbox.KeyCtrlS:
		out.Ctrl = true
		out.Ch = 's'
	case termbox.KeyCtrlT:
		out.Ctrl = true
		out.Ch = 't'
	case termbox.KeyCtrlU:
		out.Ctrl = true
		out.Ch = 'u'
	case termbox.KeyCtrlV:
		out.Ctrl = true
		out.Ch = 'v'
	case termbox.KeyCtrlW:
		out.Ctrl = true
		out.Ch = 'w'
	case termbox.KeyCtrlX:
		out.Ctrl = true
		out.Ch = 'x'
	case termbox.KeyCtrlY:
		out.Ctrl = true
		out.Ch = 'y'
	case termbox.KeyCtrlZ:
		out.Ctrl = true
		out.Ch = 'z'
	case termbox.KeyEsc: // KeyCtrlLsqBracket, KeyCtrl3
		out.Key = key.Escape
	case termbox.KeyCtrl4: // KeyCtrlBackslash
		out.Ctrl = true
		out.Ch = '4'
	case termbox.KeyCtrl5: // KeyCtrlRsqBracket
		out.Ctrl = true
		out.Ch = '5'
	case termbox.KeyCtrl6:
		out.Ctrl = true
		out.Ch = '6'
	case termbox.KeyCtrl7: // KeyCtrlSlash, KeyCtrlUnderscore
		out.Ctrl = true
		out.Ch = '7'
	case termbox.KeySpace:
		out.Key = key.Space
	default:
		panic(fmt.Sprintf("Unhandled key %x", k.Key))
	}
	return out
}

// Converts a RGB color into the nearest termbox color.
func rgbToTermBox(c colors.RGB) termbox.Attribute {
	switch colors.NearestEGA(c) {
	case colors.Black:
		return termbox.ColorBlack
	case colors.Blue:
		return termbox.ColorBlue
	case colors.Green:
		return termbox.ColorGreen
	case colors.Cyan:
		return termbox.ColorCyan
	case colors.Red:
		return termbox.ColorRed
	case colors.Magenta:
		return termbox.ColorMagenta
	case colors.Brown:
		return termbox.ColorYellow
	case colors.LightGray:
		return termbox.ColorWhite
	case colors.DarkGray:
		return termbox.ColorBlack | termbox.AttrBold
	case colors.BrightBlue:
		return termbox.ColorBlue | termbox.AttrBold
	case colors.BrightGreen:
		return termbox.ColorGreen | termbox.AttrBold
	case colors.BrightCyan:
		return termbox.ColorCyan | termbox.AttrBold
	case colors.BrightRed:
		return termbox.ColorRed | termbox.AttrBold
	case colors.BrightMagenta:
		return termbox.ColorMagenta | termbox.AttrBold
	case colors.BrightYellow:
		return termbox.ColorYellow | termbox.AttrBold
	case colors.White:
		return termbox.ColorWhite | termbox.AttrBold
	default:
		return termbox.ColorDefault
	}
}

// Blit converts the editor.Buffer format into termbox format.
func (t TermBox) Blit(b *raster.Buffer) {
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
			cell := b.Cell(x, y)
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

// SetCursor() moves the terminal cursor.
func (t TermBox) SetCursor(col, line int) {
	termbox.SetCursor(col, line)
}
