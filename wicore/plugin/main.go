// Copyright 2014 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package plugin implements the common code to implement a wi plugin.
package plugin

import (
	"fmt"
	"io"
	"net/rpc"
	"os"

	"github.com/maruel/wi/wicore"
	"github.com/maruel/wi/wicore/lang"
)

// Plugin is a simplified interface that a plugin exposes so the common plugin
// framework can efficiently communicate with the editor process.
type Plugin interface {
	Details() wicore.PluginDetails
	OnStart(e wicore.Editor) error
	OnQuit() error
}

// PluginImpl is the base implementation of interface Plugin. Embed this
// structure and override the functions desired.
type PluginImpl struct {
	Name        string
	Description lang.Map
}

func (p *PluginImpl) Details() wicore.PluginDetails {
	return wicore.PluginDetails{
		p.Name,
		p.Description.String(),
	}
}

func (p *PluginImpl) OnStart(e wicore.Editor) error {
	return nil
}

func (p *PluginImpl) OnQuit() error {
	return nil
}

// pluginRPC implements wicore.PluginRPC and implement common bookeeping.
type pluginRPC struct {
	conn         io.Closer
	langListener wicore.EventListener
	plugin       Plugin
	e            wicore.Editor
}

func (p *pluginRPC) GetInfo(l lang.Language, out *wicore.PluginDetails) error {
	lang.Set(l)
	*out = p.plugin.Details()
	return nil
}

func (p *pluginRPC) OnStart(int, *int) error {
	// TODO(maruel): Create the proxy.
	p.e = nil
	if p.e != nil {
		p.langListener = p.e.RegisterEditorLanguage(func(l lang.Language) {
			// Propagate the information.
			lang.Set(l)
		})
	}
	return p.plugin.OnStart(p.e)
}

func (p *pluginRPC) Quit(int, *int) error {
	// TODO(maruel): Is it really worth cancelling event listeners? It's just
	// unnecessary slow down, we should favor performance in the shutdown code.
	if p.langListener != nil {
		_ = p.langListener.Close()
		p.langListener = nil
	}
	p.e = nil
	err := p.plugin.OnQuit()
	if p.conn != nil {
		_ = p.conn.Close()
		p.conn = nil
	}
	return err
}

// Main is the function to call from your plugin to initiate the communication
// channel between wi and your plugin.
func Main(plugin Plugin) {
	if os.ExpandEnv("${WI}") != "plugin" {
		fmt.Fprint(os.Stderr, "This is a wi plugin. This program is only meant to be run through wi itself.\n")
		os.Exit(1)
	}
	// TODO(maruel): Take garbage from os.Stdin, put garbage in os.Stdout.
	fmt.Print(wicore.CalculateVersion())

	conn := wicore.MakeReadWriteCloser(os.Stdin, os.Stdout)
	server := rpc.NewServer()
	p := &pluginRPC{
		conn:   os.Stdin,
		plugin: plugin,
	}
	// The reason is to statically assert the interface is correctly implemented.
	var _ wicore.PluginRPC = p
	_ = server.RegisterName("PluginRPC", p)
	server.ServeConn(conn)
	os.Exit(0)
}
