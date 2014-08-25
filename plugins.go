// Copyright 2014 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package main

import (
	"fmt"
	"github.com/maruel/wi/wi-plugin"
	"io/ioutil"
	"log"
	"net/rpc"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
)

// TODO(maruel): Implement the RPC to make plugins work.

type plugin interface {
	Terminate()
}

type pluginImpl struct {
	proc *os.Process
}

func (p *pluginImpl) Terminate() {
	// TODO(maruel): Nicely terminate them.
	if err := p.proc.Kill(); err != nil {
		panic(err)
	}
}

type plugins []plugin

func (p plugins) Terminate() {
	for _, instance := range p {
		instance.Terminate()
	}
}

// loadPlugin starts a plugin and returns the process.
func loadPlugin(server *rpc.Server, f string) plugin {
	log.Printf("loadPlugin(%s)", f)
	cmd := exec.Command(f)
	cmd.Env = append(os.Environ(), "WI=plugin")

	stdin, err := cmd.StdinPipe()
	if err != nil {
		panic(err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		panic(err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		panic(err)
	}
	if err := cmd.Start(); err != nil {
		// Surface the error as an "alert", since it's not a fatal error.
		log.Fatal(err)
	}

	// Fail on any write to Stderr.
	go func() {
		buf := make([]byte, 2048)
		n, _ := stderr.Read(buf)
		if n != 0 {
			panic(fmt.Sprintf("Plugin %s failed: %s", f, buf))
		}
	}()

	// Before starting the RPC, ensures the version matches.
	expectedVersion := wi.CalculateVersion()
	b := make([]byte, 40)
	n, err := stdout.Read(b)
	if err != nil {
		// Surface the error as an "alert", since it's not a fatal error.
		log.Fatal(err)
	}
	if n != 40 {
		// Surface the error as an "alert", since it's not a fatal error.
		log.Fatal("Unexpected size")
	}
	actualVersion := string(b)
	if expectedVersion != actualVersion {
		// Surface the error as an "alert", since it's not a fatal error.
		log.Fatalf("For %s; expected %s, got %s", f, expectedVersion, actualVersion)
	}

	// Start the RPC server for this plugin.
	go func() {
		server.ServeConn(wi.MakeReadWriteCloser(stdout, stdin))
	}()

	return &pluginImpl{cmd.Process}
}

// loadPlugins loads all the plugins and returns the process handles.
func loadPlugins(e wi.Editor) plugins {
	// TODO(maruel): Get the path of the executable. It's a bit involved since
	// very OS specific but it's doable. Then all plugins in the same directory
	// are access.
	searchDir := "."
	files, err := ioutil.ReadDir(searchDir)
	if err != nil {
		return nil
	}
	if len(files) == 0 {
		// Save registering RPC stuff when unnecessary.
		return nil
	}

	var wg sync.WaitGroup
	c := make(chan plugin)
	server := rpc.NewServer()
	// TODO(maruel): http://golang.org/pkg/net/rpc/#Server.RegisterName
	// It should be an interface with methods of style DoStuff(Foo, Bar) Baz
	//server.RegisterName("Editor", e)
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		name := f.Name()
		if !strings.HasPrefix(name, "wi-plugin-") {
			continue
		}
		// Crude check for executable test.
		if runtime.GOOS == "windows" {
			if !strings.HasSuffix(name, ".exe") {
				continue
			}
		} else {
			if f.Mode()&0111 == 0 {
				continue
			}
		}
		wg.Add(1)
		go func(s *rpc.Server, n string) {
			c <- loadPlugin(s, n)
			wg.Done()
		}(server, name)
	}

	var wg2 sync.WaitGroup
	out := make(plugins, 0)
	wg2.Add(1)
	go func() {
		for i := range c {
			out = append(out, i)
		}
		wg2.Done()
	}()

	// Wait for all the plugins to be loaded.
	wg.Wait()

	// Convert to a slice.
	close(c)
	wg2.Wait()
	return out
}
