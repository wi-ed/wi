// Copyright 2013 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// wi - Bringing text based editor technology past 1200 bauds.
//
// This package contains only the non-unit-testable part of the editor.
// Everything else is in wi-editor/.
//
// See README.md for more details.
package main

import (
	"flag"
	"fmt"
	"github.com/maruel/wi/wi-editor"
	"github.com/maruel/wi/wi-plugin"
	"github.com/nsf/termbox-go"
	"log"
	"os"
)

type nullWriter struct{}

func (nullWriter) Write([]byte) (int, error) {
	return 0, nil
}

func Main() int {
	// All of "flag", "log" and "termbox" use a lot of global variables so they
	// can't be easily included in parallel tests.
	log.SetFlags(log.Lmicroseconds | log.Lshortfile)
	command := flag.Bool("c", false, "Runs the commands specified on startup")
	version := flag.Bool("v", false, "Prints version and exit")
	noPlugin := flag.Bool("no-plugin", false, "Disable loading plugins")
	verbose := flag.Bool("verbose", false, "Logs debugging information to wi.log")
	flag.Parse()

	// Process this one early. No one wants version output to take 1s.
	if *version {
		println(version)
		return 0
	}

	if *verbose {
		if f, err := os.OpenFile("wi.log", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666); err == nil {
			defer func() {
				_ = f.Close()
			}()
			log.SetOutput(f)
		}
	} else {
		log.SetOutput(nullWriter{})
	}

	if *command && flag.NArg() == 0 {
		fmt.Fprintf(os.Stderr, "error: -c implies specifying commands to execute")
		return 1
	}

	if err := termbox.Init(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to initialize terminal: %s", err)
		return 1
	}
	// It is really important that no other goroutine panic() or call
	// log.Fatal(), otherwise the terminal will be left in a broken state.
	defer termbox.Close()
	termbox.SetInputMode(termbox.InputAlt | termbox.InputMouse)

	e := editor.MakeEditor(&TermBox{})
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
	wi.PostCommand(e, "window_log_tree")
	return editor.Main(*noPlugin, e)
}

func main() {
	os.Exit(Main())
}
