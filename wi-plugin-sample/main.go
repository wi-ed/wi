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
	"log"
	"os"

	"github.com/wi-ed/wi/wicore"
	"github.com/wi-ed/wi/wicore/key"
	"github.com/wi-ed/wi/wicore/lang"
	"github.com/wi-ed/wi/wicore/plugin"
)

type pluginImpl struct {
	plugin.Impl
	e wicore.Editor
}

// Init is the place to do full initialization. It is not required to implement
// this function.
func (p *pluginImpl) Init(e wicore.Editor) {
	p.e = e

	// TODO(maruel): Generate automatically?
	e.RegisterCommands(func(cmds wicore.EnqueuedCommands) {
		//log.Printf("Commands(%v)", cmds)
	})
	e.RegisterDocumentCreated(func(doc wicore.Document) {
		log.Printf("DocumentCreated(%s)", doc)
	})
	e.RegisterDocumentCursorMoved(func(doc wicore.Document, col, row int) {
		log.Printf("DocumentCursorMoved(%s, %d, %d)", doc, col, row)
	})
	e.RegisterEditorKeyboardModeChanged(func(mode wicore.KeyboardMode) {
		log.Printf("EditorKeyboardModeChanged(%s)", mode)
	})
	e.RegisterEditorLanguage(func(l lang.Language) {
		log.Printf("EditorLanguage(%s)", l)
	})
	e.RegisterTerminalResized(func() {
		log.Printf("TerminalResized()")
	})
	e.RegisterTerminalKeyPressed(func(k key.Press) {
		log.Printf("TerminalKeyPressed(%s)", k)
	})
	e.RegisterViewCreated(func(view wicore.View) {
		log.Printf("ViewCreated(%s)", view)
	})
	e.RegisterWindowCreated(func(window wicore.Window) {
		log.Printf("WindowCreated(%s)", window)
	})
	e.RegisterWindowResized(func(window wicore.Window) {
		log.Printf("WindowResized(%s)", window)
	})
}

// Close is the place to do full shut down. It is not required to implement
// this function.
func (p *pluginImpl) Close() error {
	p.e = nil
	return nil
}

func mainImpl() int {
	log.SetFlags(log.Lmicroseconds | log.Lshortfile)
	if f, err := os.OpenFile("wi-plugin-sample.log", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666); err == nil {
		log.SetOutput(f)
		defer func() {
			_ = f.Close()
		}()
	}
	// This starts the control loop. See its doc for more up-to-date details.
	p := &pluginImpl{
		plugin.Impl{
			"wi-plugin-sample",
			lang.Map{
				lang.En: "Sample plugin to be used as a template",
				lang.Fr: "Plugin exemple pour être utilisé comme modèle",
			},
		},
		nil,
	}
	return plugin.Main(p)
}

func main() {
	os.Exit(mainImpl())
}
