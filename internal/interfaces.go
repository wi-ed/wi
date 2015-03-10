// Copyright 2015 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package internal

import (
	"github.com/wi-ed/wi/wicore"
	"github.com/wi-ed/wi/wicore/lang"
)

// PluginRPC is the low-level interface exposed by the plugin for use by
// net/rpc. net/rpc forces the interface to be in a rigid format.
type PluginRPC interface {
	// GetInfo is the fisrt function to be called synchronously. It must return
	// immediately.
	GetInfo(ignored lang.Language, out *wicore.PluginDetails) error
	// Init is called on plugin startup. All initialization should be done there.
	Init(in wicore.EditorDetails, ignored *int) error
	// Quit is called on editor termination. The editor waits for the function to
	// return.
	Quit(in int, ignored *int) error
}
