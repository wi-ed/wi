// Copyright 2014 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// This file implements the conversion of editor.Terminal to termbox's.

package main

import (
	"github.com/maruel/wi/wi-editor"
	"github.com/maruel/wi/wi-plugin"
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
					Key:  termboxKeyToKeyPress(e),
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

// termboxKeyToKeyPress returns the wi.KeyPress compatible event.
func termboxKeyToKeyPress(key termbox.Event) editor.KeyPress {
	out := editor.KeyPress{}
	if key.Mod&termbox.ModAlt != 0 {
		out.Alt = true
	}
	switch termbox.Key(key.Key) {
	case termbox.KeyF1:
		out.Key = editor.KeyF1
	case termbox.KeyF2:
		out.Key = editor.KeyF2
	case termbox.KeyF3:
		out.Key = editor.KeyF3
	case termbox.KeyF4:
		out.Key = editor.KeyF4
	case termbox.KeyF5:
		out.Key = editor.KeyF5
	case termbox.KeyF6:
		out.Key = editor.KeyF6
	case termbox.KeyF7:
		out.Key = editor.KeyF7
	case termbox.KeyF8:
		out.Key = editor.KeyF8
	case termbox.KeyF9:
		out.Key = editor.KeyF9
	case termbox.KeyF10:
		out.Key = editor.KeyF10
	case termbox.KeyF11:
		out.Key = editor.KeyF11
	case termbox.KeyF12:
		out.Key = editor.KeyF12
	case termbox.KeyInsert:
		out.Key = editor.KeyInsert
	case termbox.KeyDelete:
		out.Key = editor.KeyDelete
	case termbox.KeyHome:
		out.Key = editor.KeyHome
	case termbox.KeyEnd:
		out.Key = editor.KeyEnd
	case termbox.KeyPgup:
		out.Key = editor.KeyPageUp
	case termbox.KeyPgdn:
		out.Key = editor.KeyPageDown
	case termbox.KeyArrowUp:
		out.Key = editor.KeyArrowUp
	case termbox.KeyArrowDown:
		out.Key = editor.KeyArrowDown
	case termbox.KeyArrowLeft:
		out.Key = editor.KeyArrowLeft
	case termbox.KeyArrowRight:
		out.Key = editor.KeyArrowRight
	case termbox.KeyCtrlSpace: // KeyCtrlTilde, KeyCtrl2
		if key.Ch == 0 {
			out.Ctrl = true
			out.Ch = ' '
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
		out.Key = editor.KeyBackspace
	case termbox.KeyTab: // KeyCtrlI
		out.Key = editor.KeyTab
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
		out.Key = editor.KeyEnter
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
		out.Key = editor.KeyEscape
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
		if key.Ch != 0 {
			out.Ch = key.Ch
		}
	}
	return out
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

// Blit converts the editor.Buffer format into termbox format.
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
