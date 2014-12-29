// Copyright 2013 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package editor

import "github.com/maruel/wi/wicore"

// commandView would normally be in a floating Window near the current cursor
// on the last focused Window or at the very last line at the bottom of the
// screen.
type commandView struct {
	view
}

func (v *commandView) Buffer() *wicore.Buffer {
	v.buffer.Fill(wicore.Cell{' ', v.DefaultFormat()})
	v.buffer.DrawString(v.Title(), 0, 0, v.DefaultFormat())
	return v.buffer
}

// The command dialog box.
// TODO(maruel): Position it 5 lines below the cursor in the parent Window's
// View. Do this via onAttach.
func commandViewFactory(args ...string) wicore.View {
	keys := makeKeyBindings()
	// Fill up the key bindings. This includes basic cursor movement, help, etc.
	//keys.Set(wicore.CommandMode, "Enter", "ExecuteCommmad")

	return &commandView{
		view{
			commands:      makeCommands(),
			keyBindings:   keys,
			title:         "Command",
			naturalX:      30,
			naturalY:      1,
			defaultFormat: wicore.CellFormat{Fg: wicore.Green, Bg: wicore.Black},
		},
	}
}
