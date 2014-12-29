// Copyright 2014 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// This file implements the conversion of editor.Terminal to termbox's.

package main

import (
	"github.com/maruel/wi/editor"
	"github.com/maruel/wi/wiCore"
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
	go func() {
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
	}()
	return c
}

// termboxKeyToKeyPress returns the wiCore.KeyPress compatible event.
func termboxKeyToKeyPress(key termbox.Event) wiCore.KeyPress {
	out := wiCore.KeyPress{}
	if key.Mod&termbox.ModAlt != 0 {
		out.Alt = true
	}
	switch termbox.Key(key.Key) {
	case termbox.KeyF1:
		out.Key = wiCore.KeyF1
	case termbox.KeyF2:
		out.Key = wiCore.KeyF2
	case termbox.KeyF3:
		out.Key = wiCore.KeyF3
	case termbox.KeyF4:
		out.Key = wiCore.KeyF4
	case termbox.KeyF5:
		out.Key = wiCore.KeyF5
	case termbox.KeyF6:
		out.Key = wiCore.KeyF6
	case termbox.KeyF7:
		out.Key = wiCore.KeyF7
	case termbox.KeyF8:
		out.Key = wiCore.KeyF8
	case termbox.KeyF9:
		out.Key = wiCore.KeyF9
	case termbox.KeyF10:
		out.Key = wiCore.KeyF10
	case termbox.KeyF11:
		out.Key = wiCore.KeyF11
	case termbox.KeyF12:
		out.Key = wiCore.KeyF12
	case termbox.KeyInsert:
		out.Key = wiCore.KeyInsert
	case termbox.KeyDelete:
		out.Key = wiCore.KeyDelete
	case termbox.KeyHome:
		out.Key = wiCore.KeyHome
	case termbox.KeyEnd:
		out.Key = wiCore.KeyEnd
	case termbox.KeyPgup:
		out.Key = wiCore.KeyPageUp
	case termbox.KeyPgdn:
		out.Key = wiCore.KeyPageDown
	case termbox.KeyArrowUp:
		out.Key = wiCore.KeyArrowUp
	case termbox.KeyArrowDown:
		out.Key = wiCore.KeyArrowDown
	case termbox.KeyArrowLeft:
		out.Key = wiCore.KeyArrowLeft
	case termbox.KeyArrowRight:
		out.Key = wiCore.KeyArrowRight

	case termbox.KeyCtrlSpace: // KeyCtrlTilde, KeyCtrl2
		// This value is 0, which cannot be distinguished from non-keypress.
		if key.Ch == 0 {
			out.Ctrl = true
			out.Ch = ' '
		} else {
			// Normal keypress code path.
			out.Ch = key.Ch
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
		out.Key = wiCore.KeyBackspace
	case termbox.KeyTab: // KeyCtrlI
		out.Key = wiCore.KeyTab
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
		out.Key = wiCore.KeyEnter
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
		out.Key = wiCore.KeyEscape
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
		out.Ch = ' '
	default:
		panic("Remove me")
	}
	return out
}

// Converts a RGB color into the nearest termbox color.
func rgbToTermBox(c wiCore.RGB) termbox.Attribute {
	switch wiCore.NearestEGAColor(c) {
	case wiCore.Black:
		return termbox.ColorBlack
	case wiCore.Blue:
		return termbox.ColorBlue
	case wiCore.Green:
		return termbox.ColorGreen
	case wiCore.Cyan:
		return termbox.ColorCyan
	case wiCore.Red:
		return termbox.ColorRed
	case wiCore.Magenta:
		return termbox.ColorMagenta
	case wiCore.Brown:
		return termbox.ColorYellow
	case wiCore.LightGray:
		return termbox.ColorWhite
	case wiCore.DarkGray:
		return termbox.ColorBlack | termbox.AttrBold
	case wiCore.BrightBlue:
		return termbox.ColorBlue | termbox.AttrBold
	case wiCore.BrightGreen:
		return termbox.ColorGreen | termbox.AttrBold
	case wiCore.BrightCyan:
		return termbox.ColorCyan | termbox.AttrBold
	case wiCore.BrightRed:
		return termbox.ColorRed | termbox.AttrBold
	case wiCore.BrightMagenta:
		return termbox.ColorMagenta | termbox.AttrBold
	case wiCore.BrightYellow:
		return termbox.ColorYellow | termbox.AttrBold
	case wiCore.White:
		return termbox.ColorWhite | termbox.AttrBold
	default:
		return termbox.ColorDefault
	}
}

// Blit converts the editor.Buffer format into termbox format.
func (t TermBox) Blit(b *wiCore.Buffer) {
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
