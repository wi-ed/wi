// Copyright 2013 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package wi brings text based editor technology past 1200 bauds.
//
// This package contains only the non-unit-testable part of the editor.
//
//   - editor/ contains the editor logic itself. It is terminal-agnostic.
//   - wicore/ contains the plugin glue. This module is shared by both the
//     editor itself and any wi-plugin-* for RPC.
//   - wi-plugin-sample/ is a sample plugin executable to `go install`. It is
//     both meant as a reusable skeleton to write a new plugin and as a way to
//     ensure the plugin system works.
//
// This project supports 'Debug' and 'Release' builds. The Release build is the
// default, the Debug build has to be built explicitly. Use the following
// command to generate a Debug build:
//
//   go build -tags debug
//
// A debug build has additional functionalities:
//
//   - Logs to wi.log.
//   - Has additional flags, for example it can create cpu profiles via
//     -cpuprofile and optionally serve profiling data over a builtin web server
//     at http://localhost:6060/debug/pprof via net/http/pprof with flag
//     -http=:6060.
//   - Has additional commands defined, see editor/debug.go for the list.
//
// Run "wi -h" for help about the additional flags after doing a Debug build.
//
// See README.md for more details.
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/maruel/wi/editor"
	"github.com/maruel/wi/wicore"
	"github.com/nsf/termbox-go"
)

func terminalThread(returnCode chan<- int, mustClose chan<- func()) {
	// "flag" and "termbox" use a lot of global variables so they can't be easily
	// included in parallel tests.
	command := flag.Bool("c", false, "Runs the commands specified on startup")
	version := flag.Bool("v", false, "Prints version and exit")
	noPlugin := flag.Bool("no-plugin", false, "Disable loading plugins")
	flag.Parse()

	// Process this one early. No one wants version output to take 1s.
	if *version {
		println(version)
		returnCode <- 0
		return
	}

	if *command && flag.NArg() == 0 {
		fmt.Fprintf(os.Stderr, "error: -c implies specifying commands to execute")
		returnCode <- 1
		return
	}

	if err := termbox.Init(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to initialize terminal: %s", err)
		returnCode <- 1
		return
	}

	out := debugHook()
	if out != nil {
		defer func() {
			_ = out.Close()
		}()
	}

	// It is really important that all other goroutine wrap with handlePanic(),
	// otherwise the terminal will be left in a broken state.
	mustClose <- termbox.Close
	termbox.SetInputMode(termbox.InputAlt | termbox.InputMouse)

	e, err := editor.MakeEditor(&TermBox{}, *noPlugin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s", err)
		returnCode <- 1
		return
	}
	defer func() {
		_ = e.Close()
	}()
	debugHookEditor(e)

	wicore.PostCommand(e, nil, "editor_bootstrap_ui")
	if *command {
		for _, i := range flag.Args() {
			wicore.PostCommand(e, nil, i)
		}
	} else if flag.NArg() > 0 {
		for _, i := range flag.Args() {
			wicore.PostCommand(e, nil, "open", i)
		}
	} else {
		// If nothing, opens a blank editor.
		wicore.PostCommand(e, nil, "new")
	}
	returnCode <- e.EventLoop()
}

func mainImpl() int {
	returnCode := make(chan int)
	var closer func()
	mustClose := make(chan func())
	wicore.Go("terminalThread", func() { terminalThread(returnCode, mustClose) })
	for {
		select {
		case c := <-mustClose:
			closer = c
		case r := <-returnCode:
			if closer != nil {
				closer()
			}
			return r
		case p := <-wicore.GotPanic:
			fmt.Fprintf(os.Stderr, "Got panic!\n")
			if closer != nil {
				closer()
			}
			fmt.Fprintf(os.Stderr, "Panic: %s\n", p)
			return 1
		}
	}
}

func main() {
	os.Exit(mainImpl())
}
