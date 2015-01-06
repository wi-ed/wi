// Copyright 2013 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package editor

import (
	"github.com/maruel/wi/pkg/colors"
	"github.com/maruel/wi/pkg/key"
	"github.com/maruel/wi/wicore"
)

// commandView would normally be in a floating Window near the current cursor
// on the last focused Window or at the very last line at the bottom of the
// screen.
type commandView struct {
	view
	text string
}

func (v *commandView) Buffer() *wicore.Buffer {
	v.buffer.Fill(wicore.Cell{' ', v.DefaultFormat()})
	v.buffer.DrawString(v.text, 0, 0, v.DefaultFormat())
	return v.buffer
}

func (v *commandView) onTerminalKeyPressed(k key.Press) {
	// TODO(maruel): React to keys.
	if k.Ch != '\000' {
		v.text += string(k.Ch)
	} else {
		switch k.Key {
		case key.Escape:
			// Dismiss window.
		case key.Enter:
			// Execute command.
		case key.Space:
			v.text += " "
		case key.Tab:
			// Command completion.
		}
	}
}

// The command dialog box.
//
// TODO(maruel): Position it 5 lines below the cursor in the parent Window's
// View. Do this via onAttach.
func commandViewFactory(e wicore.Editor, args ...string) wicore.View {
	bindings := makeKeyBindings()
	// Fill up the key bindings. This includes basic cursor movement, help, etc.
	//bindings.Set(wicore.AllMode, key.Press{Key: key.Enter}, "execute_command")
	//bindings.Set(wicore.AllMode, key.Press{Key: key.Escape}, "window_close")
	v := &commandView{
		view{
			commands:      makeCommands(),
			keyBindings:   bindings,
			title:         "Command",
			naturalX:      30,
			naturalY:      1,
			defaultFormat: wicore.CellFormat{Fg: colors.Green, Bg: colors.Black},
		},
		"",
	}
	event := e.RegisterTerminalKeyPressed(v.onTerminalKeyPressed)
	e.RegisterViewActivated(func(v wicore.View) {
		_ = event.Close()
	})
	return v
}
