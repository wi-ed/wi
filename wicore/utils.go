// Copyright 2014 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Utility functions.

package wicore

import (
	"io"
	"reflect"

	"github.com/maruel/interfaceGUID"
	"github.com/wi-ed/wi/wicore/key"
	"github.com/wi-ed/wi/wicore/raster"
)

// CalculateVersion returns the hex string of the hash of the primary
// interfaces for this package.
//
// It traverses the Editor type recursively, expanding all types referenced
// recursively. This data is used to generate an hash that represents the
// "version" of this interface.
func CalculateVersion() string {
	// TODO(maruel): EditorW, Plugin and PluginRPC.
	return interfaceGUID.CalculateGUID(reflect.TypeOf((*EditorW)(nil)).Elem())
}

// GetKeyBindingCommand traverses the Editor's Window tree to find a View that
// has the key binding in its Keyboard mapping.
func GetKeyBindingCommand(e Editor, mode KeyboardMode, key key.Press) string {
	active := e.ActiveWindow()
	for {
		cmdName := active.View().KeyBindings().Get(mode, key)
		if cmdName != "" {
			return cmdName
		}
		active = active.Parent()
		if active == nil {
			return ""
		}
	}
}

// RootWindow returns the root Window when given any Window in the tree.
func RootWindow(w Window) Window {
	for {
		if w.Parent() == nil {
			return w
		}
		w = w.Parent()
	}
}

// PositionOnScreen returns the exact position on screen of a Window.
func PositionOnScreen(w Window) raster.Rect {
	out := w.Rect()
	if w.Docking() == DockingFloating {
		return out
	}
	for {
		w = w.Parent()
		if w == nil {
			break
		}
		// Take in account the parent Window position.
		r := w.Rect()
		out.X += r.X
		out.Y += r.Y
		if w.Docking() == DockingFloating {
			break
		}
	}
	return out
}

// MultiCloser closes multiple io.Closer at once.
type MultiCloser []io.Closer

// Close implements io.Closer.
func (m MultiCloser) Close() (err error) {
	for _, i := range m {
		err1 := i.Close()
		if err1 != nil {
			err = err1
		}
	}
	return
}

// MakeReadWriteCloser creates a io.ReadWriteCloser out of one io.ReadCloser
// and one io.WriteCloser.
func MakeReadWriteCloser(reader io.ReadCloser, writer io.WriteCloser) io.ReadWriteCloser {
	return struct {
		io.Reader
		io.Writer
		io.Closer
	}{reader, writer, MultiCloser{reader, writer}}
}
