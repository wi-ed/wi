// Copyright 2014 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package editor

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/rpc"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"

	"github.com/maruel/wi/wicore"
)

// TODO(maruel): Implement the RPC to make plugins work.

// Plugin represents a live plugin process.
type Plugin interface {
	io.Closer
}

// pluginProcess represents an out-of-process plugin.
type pluginProcess struct {
	proc *os.Process
}

func (p *pluginProcess) Close() error {
	if p.proc != nil {
		// TODO(maruel): Nicely terminate the child process via an RPC.
		return p.proc.Kill()
	}
	return nil
}

// pluginInline is a "plugin" that lives in the same process. It is used for
// "stock" plugins and for unit testing.
type pluginInline struct {
}

func (p *pluginInline) Close() error {
	return nil
}

// Plugins is the collection of Plugin instances, it represents all the live
// plugin processes.
type Plugins []Plugin

// Close implements io.Closer.
func (p Plugins) Close() error {
	var out error
	for _, instance := range p {
		if err := instance.Close(); err != nil {
			out = err
		}
	}
	return out
}

// loadPlugin starts a plugin and returns the process.
func loadPlugin(server *rpc.Server, args ...string) (Plugin, error) {
	log.Printf("loadPlugin(%v)", args)
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Env = append(os.Environ(), "WI=plugin")

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		return nil, err
	}

	// Fail on any write to Stderr.
	go func() {
		buf := make([]byte, 2048)
		n, _ := stderr.Read(buf)
		if n != 0 {
			// TODO(maruel): Must not panic but instead send an alert command.
			panic(fmt.Sprintf("Plugin %v failed: %s", args, buf))
		}
	}()

	// Before starting the RPC, ensures the version matches.
	expectedVersion := wicore.CalculateVersion()
	b := make([]byte, 40)
	if _, err := stdout.Read(b); err != nil {
		return nil, err
	}
	actualVersion := string(b)
	if expectedVersion != actualVersion {
		return nil, fmt.Errorf("unexpected wicore version; expected %s, got %s", expectedVersion, actualVersion)
	}

	// Start the RPC server for this plugin.
	go func() {
		server.ServeConn(wicore.MakeReadWriteCloser(stdout, stdin))
	}()

	return &pluginProcess{cmd.Process}, nil
}

// executeRaw starts a plugin present as source file for quick hacking. It
// first compile the file then run it.
func executeRaw(server *rpc.Server, filePath string) (Plugin, error) {
	return loadPlugin(server, "go", "run", filePath)
}

// EnumPlugins enumerate the plugins that should be loaded.
//
// TODO(maruel): Get the path of the executable. It's a bit involved since very
// OS specific but it's doable. Then all plugins in the same directory are
// access.
func EnumPlugins(searchDir string) ([]string, error) {
	files, err := ioutil.ReadDir(searchDir)
	if err != nil {
		return nil, err
	}
	if len(files) == 0 {
		return nil, nil
	}

	out := []string{}
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
		out = append(out, name)
	}
	return out, nil
}

func loadPlugins(pluginExecutables []string) (Plugins, error) {
	type x struct {
		Plugin
		error
	}
	c := make(chan x)
	server := rpc.NewServer()
	// TODO(maruel): http://golang.org/pkg/net/rpc/#Server.RegisterName
	// It should be an interface with methods of style DoStuff(Foo, Bar) Baz
	//server.RegisterName("Editor", e)
	go func() {
		var wg sync.WaitGroup
		for _, name := range pluginExecutables {
			wg.Add(1)
			go func(s *rpc.Server, n string) {
				defer wg.Done()
				if p, err := loadPlugin(s, n); err != nil {
					c <- x{error: fmt.Errorf("failed to load %s: %s", n, err)}
				} else {
					c <- x{Plugin: p}
				}
			}(server, name)
		}
		// Wait for all the plugins to be loaded.
		wg.Wait()
		close(c)
	}()

	// Convert to a slice.
	var wg sync.WaitGroup
	out := make(Plugins, 0, len(pluginExecutables))
	errs := make([]error, 0)
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := range c {
			if i.error != nil {
				errs = append(errs, i.error)
			} else {
				out = append(out, i.Plugin)
			}
		}
	}()
	wg.Wait()

	var err error
	if len(errs) != 0 {
		tmp := ""
		for _, e := range errs {
			tmp += e.Error() + "\n"
		}
		err = errors.New(tmp[:len(tmp)-1])
	}
	return out, err
}
