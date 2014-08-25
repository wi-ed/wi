// Copyright 2014 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package editor

import (
	"github.com/maruel/wi/wi-plugin"
	"github.com/nsf/termbox-go"
)

type keyBindings struct {
	commandMappings map[string]string
	editMappings    map[string]string
}

func (k *keyBindings) Set(mode wi.KeyboardMode, keyName string, cmdName string) bool {
	var ok bool
	if mode == wi.AllMode || mode == wi.CommandMode {
		_, ok = k.commandMappings[keyName]
		k.commandMappings[keyName] = cmdName
	}
	if mode == wi.AllMode || mode == wi.EditMode {
		_, ok = k.editMappings[keyName]
		k.editMappings[keyName] = cmdName
	}
	return !ok
}

func (k *keyBindings) Get(mode wi.KeyboardMode, keyName string) string {
	if mode == wi.CommandMode {
		return k.commandMappings[keyName]
	}
	if mode == wi.EditMode {
		return k.editMappings[keyName]
	}
	v, ok := k.commandMappings[keyName]
	if !ok {
		return k.editMappings[keyName]
	}
	return v
}

func makeKeyBindings() wi.KeyBindings {
	return &keyBindings{make(map[string]string), make(map[string]string)}
}

// keyEventToName returns the user printable key name like 'a', Ctrl-Alt-F1,
// Delete, etc.
func keyEventToName(event termbox.Event) string {
	out := ""
	if event.Mod&termbox.ModAlt != 0 {
		out = "Alt-"
	}
	switch event.Key {
	case termbox.KeyF1:
		out += "F1"
	case termbox.KeyF2:
		out += "F2"
	case termbox.KeyF3:
		out += "F3"
	case termbox.KeyF4:
		out += "F4"
	case termbox.KeyF5:
		out += "F5"
	case termbox.KeyF6:
		out += "F6"
	case termbox.KeyF7:
		out += "F7"
	case termbox.KeyF8:
		out += "F8"
	case termbox.KeyF9:
		out += "F9"
	case termbox.KeyF10:
		out += "F10"
	case termbox.KeyF11:
		out += "F11"
	case termbox.KeyF12:
		out += "F12"
	case termbox.KeyInsert:
		out += "Inset"
	case termbox.KeyDelete:
		out += "Delete"
	case termbox.KeyHome:
		out += "Home"
	case termbox.KeyEnd:
		out += "End"
	case termbox.KeyPgup:
		out += "PageUp"
	case termbox.KeyPgdn:
		out += "PageDown"
	case termbox.KeyArrowUp:
		out += "Up"
	case termbox.KeyArrowDown:
		out += "Down"
	case termbox.KeyArrowLeft:
		out += "Left"
	case termbox.KeyArrowRight:
		out += "Right"
	case termbox.KeyCtrlSpace: // KeyCtrlTilde, KeyCtrl2
		if event.Ch == 0 {
			out += "Ctrl-Space"
		}
	case termbox.KeyCtrlA:
		out += "Ctrl-A"
	case termbox.KeyCtrlB:
		out += "Ctrl-B"
	case termbox.KeyCtrlC:
		out += "Ctrl-C"
	case termbox.KeyCtrlD:
		out += "Ctrl-D"
	case termbox.KeyCtrlE:
		out += "Ctrl-E"
	case termbox.KeyCtrlF:
		out += "Ctrl-F"
	case termbox.KeyCtrlG:
		out += "Ctrl-G"
	case termbox.KeyBackspace: // KeyCtrlH
		out += "Backspace"
	case termbox.KeyTab: // KeyCtrlI
		out += "Tab"
	case termbox.KeyCtrlJ:
		out += "Ctrl-J"
	case termbox.KeyCtrlK:
		out += "Ctrl-K"
	case termbox.KeyCtrlL:
		out += "Ctrl-L"
	case termbox.KeyEnter: // KeyCtrlM
		out += "Enter"
	case termbox.KeyCtrlN:
		out += "Ctrl-N"
	case termbox.KeyCtrlO:
		out += "Ctrl-O"
	case termbox.KeyCtrlP:
		out += "Ctrl-P"
	case termbox.KeyCtrlQ:
		out += "Ctrl-Q"
	case termbox.KeyCtrlR:
		out += "Ctrl-R"
	case termbox.KeyCtrlS:
		out += "Ctrl-S"
	case termbox.KeyCtrlT:
		out += "Ctrl-T"
	case termbox.KeyCtrlU:
		out += "Ctrl-U"
	case termbox.KeyCtrlV:
		out += "Ctrl-V"
	case termbox.KeyCtrlW:
		out += "Ctrl-W"
	case termbox.KeyCtrlX:
		out += "Ctrl-X"
	case termbox.KeyCtrlY:
		out += "Ctrl-Y"
	case termbox.KeyCtrlZ:
		out += "Ctrl-Z"
	case termbox.KeyEsc: // KeyCtrlLsqBracket, KeyCtrl3
		out += "Esc"
	case termbox.KeyCtrl4: // KeyCtrlBackslash
		out += "Ctrl-4"
	case termbox.KeyCtrl5: // KeyCtrlRsqBracket
		out += "Ctrl-5"
	case termbox.KeyCtrl6:
		out += "Ctrl-6"
	case termbox.KeyCtrl7: // KeyCtrlSlash, KeyCtrlUnderscore
		out += "Ctrl-7"
	case termbox.KeySpace:
		out += "Space"
	case termbox.KeyBackspace2: // KeyCtrl8
		out += "Backspace2"
	}
	if event.Ch != 0 {
		out += string(event.Ch)
	}
	return out
}

// Registers the default keyboard mapping. Keyboard mapping simply execute the
// corresponding command. So to add a keyboard map, the corresponding command
// needs to be added first.
//
// TODO(maruel): This should be remappable via a configuration flag, for
// example vim flavor vs emacs flavor. I'm not sure it's worth supporting this
// without a restart.
func RegisterDefaultKeyBindings(cd wi.CommandDispatcher) {
	cd.PostCommand("keybind", "global", "all", "F1", "help")
	cd.PostCommand("keybind", "global", "all", "Ctrl-C", "quit")
	cd.PostCommand("keybind", "global", "command", ":", "show_command_window")
}
