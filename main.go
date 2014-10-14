// Copyright 2013 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package wi brings text based editor technology past 1200 bauds.
//
// This package contains only the non-unit-testable part of the editor.
//
//   - editor/ contains the editor logic itself. It is terminal-agnostic.
//   - wi_core/ (wi) contains the plugin glue. This module is shared by both
//     the editor itself and any wi-plugin-* for RPC.
//   - wi-plugin-sample/ is a sample plugin executable to `go install`. It is
//     both meant as a reusable skeleton to write a new plugin and as a way to
//     ensure the plugin system works.
//
// See README.md for more details.
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/maruel/wi/editor"
	"github.com/maruel/wi/wi_core"
	"github.com/nsf/termbox-go"
)

func mainImpl() int {
	// "flag" and "termbox" use a lot of global variables so they can't be easily
	// included in parallel tests.
	command := flag.Bool("c", false, "Runs the commands specified on startup")
	version := flag.Bool("v", false, "Prints version and exit")
	noPlugin := flag.Bool("no-plugin", false, "Disable loading plugins")
	flag.Parse()

	// Process this one early. No one wants version output to take 1s.
	if *version {
		println(version)
		return 0
	}

	if *command && flag.NArg() == 0 {
		fmt.Fprintf(os.Stderr, "error: -c implies specifying commands to execute")
		return 1
	}

	if err := termbox.Init(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to initialize terminal: %s", err)
		return 1
	}

	out := DebugHook()
	if out != nil {
		defer func() {
			_ = out.Close()
		}()
	}

	// It is really important that no other goroutine panic() or call
	// log.Fatal(), otherwise the terminal will be left in a broken state.
	defer termbox.Close()
	termbox.SetInputMode(termbox.InputAlt | termbox.InputMouse)

	e, err := editor.MakeEditor(&TermBox{}, *noPlugin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s", err)
		return 1
	}
	defer func() {
		_ = e.Close()
	}()

	wi.PostCommand(e, "editor_bootstrap_ui")
	if *command {
		for _, i := range flag.Args() {
			wi.PostCommand(e, i)
		}
	} else if flag.NArg() > 0 {
		for _, i := range flag.Args() {
			wi.PostCommand(e, "open", i)
		}
	} else {
		// If nothing, opens a blank editor.
		wi.PostCommand(e, "new")
	}
	return e.EventLoop()
}

func main() {
	os.Exit(mainImpl())
}
