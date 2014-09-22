// Copyright 2013 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package editor contains the UI toolkit agnostic unit-testable part of the wi
// editor. It brings text based editor technology past 1200 bauds.
//
// It is in a standalone package for a few reasons:
//
// - godoc will generate documentation for this code.
//
// - it can be unit tested without having a dependency on termbox.
//
// - hide away ncurse idiocracies (like Ctrl-H == Backspace) which could be
// supported on Windows or native UI.
//
// This package is not meant to be a general purpose reusable package, the
// primary maintainer of this project likes having a web browseable
// documentation. Using a package is an effective workaround the fact that
// godoc doesn't general documentation for "main" package.
//
// See ../README.md for user information.
package editor
