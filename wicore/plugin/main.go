// Copyright 2014 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package plugin implements the common code to implement a wi plugin.
package plugin

import (
	"fmt"
	"io"
	"log"
	"net/rpc"
	"os"

	"github.com/maruel/wi/wicore"
	"github.com/maruel/wi/wicore/lang"
)

// PluginImpl is the base implementation of interface wicore.Plugin. Embed this
// structure and override the functions desired.
type PluginImpl struct {
	Name        string
	Description lang.Map
}

func (p *PluginImpl) String() string {
	return fmt.Sprintf("Plugin(%s, %d)", p.Name, os.Getpid())
}

// Details implements wicore.Plugin.
func (p *PluginImpl) Details() wicore.PluginDetails {
	return wicore.PluginDetails{
		p.Name,
		p.Description.String(),
	}
}

// Init implements wicore.Plugin.
func (p *PluginImpl) Init(e wicore.Editor) {
}

// Close implements wicore.Plugin.
func (p *PluginImpl) Close() error {
	return nil
}

// pluginRPC implements wicore.PluginRPC and implement common bookeeping.
type pluginRPC struct {
	conn         io.Closer
	langListener wicore.EventListener
	plugin       wicore.Plugin
	e            *editorProxy
}

func (p *pluginRPC) GetInfo(l lang.Language, out *wicore.PluginDetails) error {
	lang.Set(l)
	*out = p.plugin.Details()
	return nil
}

func (p *pluginRPC) Init(details wicore.EditorDetails, ignored *int) error {
	p.e.id = details.ID
	p.e.version = details.Version
	p.langListener = p.e.RegisterEditorLanguage(func(l lang.Language) {
		// Propagate the information.
		lang.Set(l)
	})
	p.plugin.Init(p.e)
	return nil
}

func (p *pluginRPC) Quit(int, *int) error {
	// TODO(maruel): Is it really worth cancelling event listeners? It's just
	// unnecessary slow down, we should favor performance in the shutdown code.
	if p.langListener != nil {
		_ = p.langListener.Close()
		p.langListener = nil
	}
	p.e = nil
	err := p.plugin.Close()
	if p.conn != nil {
		_ = p.conn.Close()
		p.conn = nil
	}
	return err
}

// editorProxy is an experimentation.
type editorProxy struct {
	wicore.EventRegistry
	deferred     chan func()
	id           string
	activeWindow wicore.Window
	factoryNames []string
	keyboardMode wicore.KeyboardMode
	version      string
}

func (e *editorProxy) ID() string {
	return e.id
}

func (e *editorProxy) ActiveWindow() wicore.Window {
	return e.activeWindow
}

func (e *editorProxy) ViewFactoryNames() []string {
	out := make([]string, len(e.factoryNames))
	for i, v := range e.factoryNames {
		out[i] = v
	}
	return out
}

func (e *editorProxy) AllDocuments() []wicore.Document {
	return nil
}

func (e *editorProxy) AllPlugins() []wicore.PluginDetails {
	return nil
}

func (e *editorProxy) KeyboardMode() wicore.KeyboardMode {
	return e.keyboardMode
}

func (e *editorProxy) Version() string {
	return e.version
}

// Main is the function to call from your plugin to initiate the communication
// channel between wi and your plugin.
//
// Returns the exit code that should be passed to os.Exit().
func Main(plugin wicore.Plugin) int {
	if os.ExpandEnv("${WI}") != "plugin" {
		fmt.Fprint(os.Stderr, "This is a wi plugin. This program is only meant to be run through wi itself.\n")
		return 1
	}
	// TODO(maruel): Take garbage from os.Stdin, put garbage in os.Stdout.
	fmt.Print(wicore.CalculateVersion())

	// TODO(maruel): Pipe logs into os.Stderr and not have the editor process
	// kill the plugin process in this case.
	conn := wicore.MakeReadWriteCloser(os.Stdin, os.Stdout)
	server := rpc.NewServer()
	reg, deferred := wicore.MakeEventRegistry()
	e := &editorProxy{
		reg,
		deferred,
		"",
		nil,
		[]string{},
		wicore.Normal,
		"",
	}
	p := &pluginRPC{
		e:      e,
		conn:   os.Stdin,
		plugin: plugin,
	}
	// Statically assert the interface is correctly implemented.
	var objPluginRPC wicore.PluginRPC = p
	if err := server.RegisterName("PluginRPC", objPluginRPC); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		return 1
	}
	// Expose an object which doesn't have any method beside the ones exposed.
	// Otherwise it spew the logs with noise.
	objEventTriggerRPC := struct{ wicore.EventTriggerRPC }{p.e}
	if err := server.RegisterName("EventTriggerRPC", objEventTriggerRPC); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		return 1
	}
	log.Printf("wicore.plugin.Main() now serving")
	server.ServeConn(conn)
	return 0
}
