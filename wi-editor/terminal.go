// Copyright 2013 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package editor

import (
	"github.com/maruel/tulib"
	"github.com/nsf/termbox-go"
)

// TermBox is the interface to termbox so it can be mocked in unit test.
type TermBox interface {
	Size() (int, int)
	Flush()
	PollEvent() termbox.Event
	Buffer() tulib.Buffer
}

type termBoxImpl struct {
}

func (t termBoxImpl) Size() (int, int) {
	return termbox.Size()
}

func (t termBoxImpl) Flush() {
	if err := termbox.Flush(); err != nil {
		panic(err)
	}
}

func (t termBoxImpl) PollEvent() termbox.Event {
	return termbox.PollEvent()
}

func (t termBoxImpl) Buffer() tulib.Buffer {
	w, h := t.Size()
	return tulib.Buffer{
		Cells: termbox.CellBuffer(),
		Rect:  tulib.Rect{0, 0, w, h},
	}
}

// Logger is the interface to log to. It must be used instead of
// log.Logger.Printf() or testing.T.Log(). This permits to collect logs for a
// complete test case.
type Logger interface {
	Logf(format string, v ...interface{})
}
