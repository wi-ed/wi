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
	"path/filepath"
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
func loadPlugin(server *rpc.Server, cmdLine []string) (Plugin, error) {
	log.Printf("loadPlugin(%v)", cmdLine)
	cmd := exec.Command(cmdLine[0], cmdLine[1:]...)
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

	first := make(chan error)

	// Fail on any write to Stderr.
	wicore.Go("stderrReader", func() {
		buf := make([]byte, 2048)
		n, _ := stderr.Read(buf)
		if n != 0 {
			first <- fmt.Errorf("plugin %v failed: %s", cmdLine, buf)
		}
	})

	wicore.Go("stdoutReader", func() {
		// Before starting the RPC, ensures the version matches.
		expectedVersion := wicore.CalculateVersion()
		b := make([]byte, 40)
		if _, err := stdout.Read(b); err != nil {
			first <- err
		}
		actualVersion := string(b)
		if expectedVersion != actualVersion {
			first <- fmt.Errorf("unexpected wicore version; expected %s, got %s", expectedVersion, actualVersion)
		}
		first <- nil
	})

	err = <-first
	if err != nil {
		return nil, err
	}

	// Start the RPC server for this plugin.
	wicore.Go("RPCserver", func() {
		server.ServeConn(wicore.MakeReadWriteCloser(stdout, stdin))
	})

	return &pluginProcess{cmd.Process}, nil
}

// getPluginsPaths returns the search paths for plugins.
//
// Currently look at ".", each element of $GOPATH/bin and in $WIPLUGINSPATH.
func getPluginsPaths() []string {
	out := []string{}
	for _, i := range filepath.SplitList(os.Getenv("GOPATH")) {
		out = append(out, filepath.Join(i, "bin"))
	}
	for _, i := range filepath.SplitList(os.Getenv("WIPLUGINSPATH")) {
		out = append(out, i)
	}
	return out
}

// enumPlugins enumerate the plugins that should be loaded.
//
// It returns the command lines to use to start the processes. It support
// normal executable, standalone source file and directory containing multiple
// source files.
//
// Source files will incur a ~500ms to ~1s compilation overhead, so they should
// eventually be compiled. Still, it's very useful for quick prototyping.
func enumPlugins(searchDirs []string) ([][]string, error) {
	out := [][]string{}
	var err error
	for _, searchDir := range searchDirs {
		files, err2 := ioutil.ReadDir(searchDir)
		if err2 != nil {
			err = err2
		}
		if len(files) == 0 {
			continue
		}

		for _, f := range files {
			name := f.Name()
			if !strings.HasPrefix(name, "wi-plugin-") {
				continue
			}
			filePath := filepath.Join(searchDir, name)

			if f.IsDir() {
				// Compile on-the-fly a directory of source files.
				// TODO(maruel): When built with -tags debug, pass it along.
				files, err2 := filepath.Glob(filepath.Join(filePath, "*.go"))
				if len(files) == 0 || err2 != nil {
					continue
				}
				i := []string{"go", "run"}
				for _, t := range files {
					i = append(i, filepath.Join(filePath, t))
				}
				out = append(out, i)
				continue
			}

			if strings.HasSuffix(name, ".go") {
				// Compile on-the-fly a source file.
				// TODO(maruel): When built with -tags debug, pass it along.
				out = append(out, []string{"go", "run", filePath})
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
			out = append(out, []string{filePath})
		}
	}
	return out, err
}

func loadPlugins(pluginExecutables [][]string) (Plugins, error) {
	type x struct {
		Plugin
		error
	}
	c := make(chan x)
	server := rpc.NewServer()
	// TODO(maruel): http://golang.org/pkg/net/rpc/#Server.RegisterName
	// It should be an interface with methods of style DoStuff(Foo, Bar) Baz
	//server.RegisterName("Editor", e)
	wicore.Go("loadPlugins", func() {
		var wg sync.WaitGroup
		for _, cmd := range pluginExecutables {
			wg.Add(1)
			wicore.Go("loadPlugin", func() {
				func(s *rpc.Server, n []string) {
					defer wg.Done()
					if p, err := loadPlugin(s, n); err != nil {
						c <- x{error: fmt.Errorf("failed to load %v: %s", n, err)}
					} else {
						c <- x{Plugin: p}
					}
				}(server, cmd)
			})
		}
		// Wait for all the plugins to be loaded.
		wg.Wait()
		close(c)
	})

	// Convert to a slice.
	var wg sync.WaitGroup
	out := make(Plugins, 0, len(pluginExecutables))
	errs := make([]error, 0)
	wg.Add(1)
	wicore.Go("pluginReaper", func() {
		defer wg.Done()
		for i := range c {
			if i.error != nil {
				errs = append(errs, i.error)
			} else {
				out = append(out, i.Plugin)
			}
		}
	})
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
