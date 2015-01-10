// Copyright 2014 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// wi-plugin-sample is an example plugin for wi.
//
// This plugin serves two purposes:
//   - Ensure that the plugin system is actually working.
//   - Serve as a copy-pastable skeleton to help people who would like to write
//     a plugin.
//
// To try it out, from the wi/ directory, run `go build` then set the
// environment variable `WIPLUGINSPATH=.`, so that this directory is
// automatically compiled via `go run`. See ../editor/plugins.go for the gory
// details.
package main

import (
	"github.com/maruel/wi/wicore/plugin"
)

func onStart() error {
	// This is the place to do full initialization.
	// TODO(maruel): At this point, the proxy wicore.Editor should be up and
	// provided to this function.
	return nil
}

func main() {
	// This starts the control loop. See its doc for more up-to-date details.
	plugin.Main("wi-plugin-sample", onStart)
}
