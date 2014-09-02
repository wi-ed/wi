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

func (t TermBox) Blit(b *wi.Buffer) {
	// Convert the wi.Buffer format into termbox format.
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
			c := b.Get(x, y)
			// TODO(maruel): Convert colors.
			cells[y*width+x].Ch = c.R
		}
	}
	if err := termbox.Flush(); err != nil {
		panic(err)
	}
}
