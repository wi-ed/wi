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
)

// pluginRPC implements wicore.PluginRPC.
type pluginRPC struct {
	conn    io.Closer
	name    string
	onStart func() error
}

func (p *pluginRPC) GetInfo(ignored int, out *wicore.PluginDetails) error {
	out.Name = p.name
	return nil
}

func (p *pluginRPC) OnStart(int, *int) error {
	if p.onStart != nil {
		return p.onStart()
	}
	return nil
}

func (p *pluginRPC) Quit(value int, _ *int) error {
	if p.conn != nil {
		_ = p.conn.Close()
		p.conn = nil
	}
	return nil
}

// Main is the function to call from your plugin to initiate the communication
// channel between wi and your plugin.
func Main(name string, onStart func() error) {
	if os.ExpandEnv("${WI}") != "plugin" {
		fmt.Fprint(os.Stderr, "This is a wi plugin. This program is only meant to be run through wi itself.\n")
		os.Exit(1)
	}
	// TODO(maruel): Take garbage from os.Stdin, put garbage in os.Stdout.
	fmt.Print(wicore.CalculateVersion())

	conn := wicore.MakeReadWriteCloser(os.Stdin, os.Stdout)
	server := rpc.NewServer()
	p := &pluginRPC{os.Stdin, name, onStart}
	_ = server.RegisterName("PluginRPC", p)
	server.ServeConn(conn)
	os.Exit(0)
}
