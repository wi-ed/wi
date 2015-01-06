// Copyright 2014 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Use "go build -tags debug" to have access to the code and commands in this
// file.

// +build debug

package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"expvar"
	"flag"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime/pprof"
	"sort"
	"strings"
	"sync"

	"github.com/maruel/wi/editor"
	"github.com/maruel/wi/wicore"
	"github.com/maruel/wi/wicore/key"
	"github.com/maruel/wi/wicore/lang"
)

var (
	httpServer = flag.String("http", "", "Start a debug web server to observe internal states")
	cpuprofile = flag.String("cpuprofile", "", "Write cpu profile to file; use \"go tool pprof wi <file>\" to read the data; See https://blog.golang.org/profiling-go-programs for more details")
	data       debugData
)

type debugData struct {
	wg        sync.WaitGroup
	lock      sync.Mutex
	newData   *sync.Cond
	quit      bool
	logBuffer bytes.Buffer // Memory leak containing the process buffer. TODO(maruel): limit its size.
	logFile   io.Closer
	profFile  io.Closer
}

func (d *debugData) Close() error {
	d.lock.Lock()
	d.quit = true
	d.lock.Unlock()
	d.newData.Broadcast()
	if d.logFile != nil {
		d.logFile.Close()
		d.logFile = nil
	}
	if d.profFile != nil {
		pprof.StopCPUProfile()
		d.profFile.Close()
		d.profFile = nil
	}
	d.wg.Wait()
	return nil
}

func (d *debugData) Write(p []byte) (int, error) {
	d.lock.Lock()
	defer d.lock.Unlock()
	d.newData.Broadcast()
	return d.logBuffer.Write(p)
}

// streamTo doesn't update b.lastRead so multiple calls can happen
// simultaneously. It forcibly flushes the output so it is sent to the
// underlying TCP connection.
func (d *debugData) streamTo(w io.Writer) (int, error) {
	d.wg.Add(1)
	defer d.wg.Done()
	f, ok := w.(http.Flusher)
	offset := 0
	var err error
	d.lock.Lock()
	defer d.lock.Unlock()
	for !d.quit {
		buf := d.logBuffer.Bytes()
		if len(buf) != offset {
			var n int
			n, err = w.Write(buf[offset:])
			offset += n
			if ok && n != 0 {
				f.Flush()
			}
		}
		if err != nil {
			break
		}
		d.newData.Wait()
	}
	return offset, err
}

func debugHook() io.Closer {
	data.newData = sync.NewCond(&data.lock)
	log.SetFlags(log.Lmicroseconds | log.Lshortfile)
	log.SetOutput(&data)
	if f, err := os.OpenFile("wi.log", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666); err == nil {
		data.logFile = f
		log.SetOutput(io.MultiWriter(&data, f))
	}

	if *cpuprofile != "" {
		if f, err := os.OpenFile(*cpuprofile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666); err == nil {
			data.profFile = f
			pprof.StartCPUProfile(f)
		} else {
			log.Printf("Failed to open %s: %s", *cpuprofile, err)
			*cpuprofile = ""
		}
	}

	// TODO(maruel): Investigate adding our own profiling for RPC.
	// http://golang.org/pkg/runtime/pprof/
	// TODO(maruel): Add pprof.WriteHeapProfile(f) when desired (?)

	if *httpServer != "" {
		http.HandleFunc("/", rootHandler)
		http.HandleFunc("/favicon.ico", faviconHandler)
		http.HandleFunc("/log", logHandler)
		wicore.Go("HTTPserver", func() {
			log.Println(http.ListenAndServe(*httpServer, nil))
		})
	}
	return &data
}

func debugHookEditor(e editor.Editor) {
	expvar.Publish("active_window", funcString(func() string { return e.ActiveWindow().String() }))
	expvar.Publish("commands", funcJSON(func() interface{} { return commands(e) }))
	expvar.Publish("documents", funcJSON(func() interface{} { return documents(e) }))
	expvar.Publish("view_factories", funcJSON(func() interface{} { return viewFactories(e) }))
	expvar.Publish("windows", funcJSON(func() interface{} { return windows(e) }))
	expvar.NewInt("pid").Set(int64(os.Getpid()))

	cmds := []wicore.Command{
		&wicore.CommandImpl{
			"command_log",
			0,
			cmdCommandLog,
			wicore.DebugCategory,
			lang.Map{
				lang.En: "Logs the registered commands",
			},
			lang.Map{
				lang.En: "Logs the registered commands, this is only relevant if -verbose is used.",
			},
		},
		&wicore.CommandImpl{
			"key_log",
			0,
			cmdKeyLog,
			wicore.DebugCategory,
			lang.Map{
				lang.En: "Logs the key bindings",
			},
			lang.Map{
				lang.En: "Logs the key bindings, this is only relevant if -verbose is used.",
			},
		},
		&wicore.CommandImpl{
			"log_all",
			0,
			cmdLogAll,
			wicore.DebugCategory,
			lang.Map{
				lang.En: "Logs the internal state (commands, view factories, windows)",
			},
			lang.Map{
				lang.En: "Logs the internal state (commands, view factories, windows), this is only relevant if -verbose is used.",
			},
		},
		&wicore.CommandImpl{
			"view_log",
			0,
			cmdViewLog,
			wicore.DebugCategory,
			lang.Map{
				lang.En: "Logs the view factories",
			},
			lang.Map{
				lang.En: "Logs the view factories, this is only relevant if -verbose is used.",
			},
		},
		&wicore.CommandImpl{
			"window_log",
			0,
			cmdWindowLog,
			wicore.DebugCategory,
			lang.Map{
				lang.En: "Logs the window tree",
			},
			lang.Map{
				lang.En: "Logs the window tree, this is only relevant if -verbose is used.",
			},
		},

		// 'editor_screenshot', mainly for unit test; open a new buffer with the screenshot, so it can be saved with 'w'.
	}
	dispatcher := wicore.RootWindow(e.ActiveWindow()).View().Commands()
	for _, cmd := range cmds {
		dispatcher.Register(cmd)
	}

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
	e.RegisterTerminalKeyPressed(func(key key.Press) {
		log.Printf("TerminalKeyPressed(%s)", key)
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

// prettyPrintJSON pretty-prints a JSON buffer. Accepts list and dict.
func prettyPrintJSON(in []byte) []byte {
	var data interface{}
	var asMap map[string]interface{}
	if err := json.Unmarshal(in, &asMap); err != nil {
		var asList []interface{}
		if err := json.Unmarshal(in, &asList); err != nil {
			data = err.Error()
		} else {
			data = asList
		}
	} else {
		data = asMap
	}
	out, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return []byte(err.Error())
	}
	return out
}

var tmplRoot = template.Must(template.New("root").Parse(`<!DOCTYPE html>
	<html>
	<head>
		<title>wi internals</title>
		<meta charset="utf-8">
		<style>
		h1 {
			font-size: 1.1em;
		}
		.data_table {
			width: 100%;
		}
		.content {
			/*font-family:Consolas,Monaco,Lucida Console,Liberation Mono,DejaVu Sans Mono,Bitstream Vera Sans Mono,Courier New, monospace;
			*/
			max-height: 300px;
			overflow: auto;
		}
		table.data_table tbody tr:nth-child(even) {
			background-color: #eeeeee;
		}
		</style>
	</head>
	<body>
	<h1>wi internal details</h1>
	<ul>
		<li>
			<a href="/log">Process log (same as in wi.log)</a>.
		</li>
		<li>
			<a href="/debug/pprof/">Profiling information</a>.
			For more information, see <a href="https://golang.org/pkg/net/http/pprof/">golang.org/pkg/net/http/pprof/</a>.
		</li>
		<li>
			<a href="/debug/vars">Raw JSON expvar</a>.
			For more information, see <a href="https://golang.org/pkg/expvar/">golang.org/pkg/expvar/</a>.
		</li>
	</ul>
	<hr>
	<table class="data_table">
		<thead>
			<tr>
				<th>Name</th>
				<th>Value</th>
			</tr>
		</thead>
		<tbody>
		{{range .Values}}
			<tr>
				<td>{{index . 0}}</td>
				<td><div class="content"><pre>{{index . 1}}</pre></div></td>
			</tr>
		{{end}}
		</tbody>
	</table>
	</body>
	</html>`))

func rootHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.Redirect(w, r, "/", http.StatusMovedPermanently)
		return
	}
	d := struct {
		Values [][2]string
	}{
		[][2]string{},
	}
	expvar.Do(func(kv expvar.KeyValue) {
		v := kv.Value.String()
		if _, ok := kv.Value.(expvar.Func); ok {
			v = string(prettyPrintJSON([]byte(v)))
		}
		d.Values = append(d.Values, [2]string{kv.Key, v})
	})
	if err := tmplRoot.Execute(w, d); err != nil {
		io.WriteString(w, err.Error())
	}
}

func logHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	data.streamTo(w)
}

func faviconHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "image/x-image")
	wiPNG, _ := base64.StdEncoding.DecodeString("iVBORw0KGgoAAAANSUhEUgAAABAAAAAQCAYAAAAf8/9hAAAAVUlEQVQ4y2NgGGjACGP8////PzkGMFHqAhQDrklLY1WELI5LDcN/KLgqJfUfGaDz0QHJXkB3AROxCkkKxGvS0gxaT58SZQiGAVpPn+LlD/J0MEINAAC5TUkhJn+lswAAAABJRU5ErkJggg==")
	w.Write(wiPNG)
}

type funcString func() string

func (f funcString) String() string {
	return f()
}

type funcJSON func() interface{}

func (f funcJSON) String() string {
	v, _ := json.MarshalIndent(f(), "", "  ")
	return string(v)
}

func commandRecurse(w wicore.Window, buf []string) []string {
	cmds := w.View().Commands()
	for _, name := range cmds.GetNames() {
		c := cmds.Get(name)
		buf = append(buf, fmt.Sprintf("%-3s  %-21s: %s", w.ID(), c.Name(), c.ShortDesc()))
	}
	for _, child := range w.ChildrenWindows() {
		buf = commandRecurse(child, buf)
	}
	return buf
}

func commands(e wicore.Editor) interface{} {
	// Start at the root and recurse.
	out := commandRecurse(wicore.RootWindow(e.ActiveWindow()), []string{})
	sort.Strings(out)
	return out
}

func documents(e wicore.Editor) interface{} {
	return e.AllDocuments()
}

func viewFactories(e wicore.Editor) interface{} {
	names := e.ViewFactoryNames()
	sort.Strings(names)
	return names
}

func recurseTree(w wicore.Window) map[string]interface{} {
	out := map[string]interface{}{
		"rect":  w.Rect(),
		"title": w.View().Title(),
		"id":    w.ID(),
	}
	children := []interface{}{}
	for _, child := range w.ChildrenWindows() {
		children = append(children, recurseTree(child))
	}
	// Use z_ so it's the last item, for easier browsing.
	if len(children) != 0 {
		out["z_children"] = children
	}
	return out
}

func windows(e wicore.Editor) interface{} {
	return recurseTree(wicore.RootWindow(e.ActiveWindow()))
}

func cmdCommandLog(c *wicore.CommandImpl, e wicore.Editor, w wicore.Window, args ...string) {
	out := commandRecurse(wicore.RootWindow(e.ActiveWindow()), []string{})
	sort.Strings(out)
	for _, i := range out {
		log.Printf("  %s", i)
	}
}

func keyLogRecurse(w wicore.Window, e wicore.Editor, mode wicore.KeyboardMode) {
	bindings := w.View().KeyBindings()
	assigned := bindings.GetAssigned(mode)
	names := make([]string, 0, len(assigned))
	for _, k := range assigned {
		names = append(names, k.String())
	}
	sort.Strings(names)
	for _, name := range names {
		log.Printf("  %s  %s: %s", w.ID(), name, bindings.Get(mode, key.StringToPress(name)))
	}
	for _, child := range w.ChildrenWindows() {
		keyLogRecurse(child, e, mode)
	}
}

func cmdKeyLog(c *wicore.CommandImpl, e wicore.Editor, w wicore.Window, args ...string) {
	log.Printf("Normal commands")
	rootWindow := wicore.RootWindow(e.ActiveWindow())
	keyLogRecurse(rootWindow, e, wicore.Normal)
	log.Printf("Insert commands")
	keyLogRecurse(rootWindow, e, wicore.Insert)
}

func cmdLogAll(c *wicore.CommandImpl, e wicore.Editor, w wicore.Window, args ...string) {
	e.ExecuteCommand(w, "command_log")
	e.ExecuteCommand(w, "window_log")
	e.ExecuteCommand(w, "view_log")
	e.ExecuteCommand(w, "key_log")
}

func cmdViewLog(c *wicore.CommandImpl, e wicore.Editor, w wicore.Window, args ...string) {
	names := e.ViewFactoryNames()
	sort.Strings(names)
	log.Printf("View factories:")
	for _, name := range names {
		log.Printf("  %s", name)
	}
}

func tree(w wicore.Window) string {
	out := w.String() + "\n"
	for _, child := range w.ChildrenWindows() {
		for _, line := range strings.Split(tree(child), "\n") {
			if line != "" {
				out += ("  " + line + "\n")
			}
		}
	}
	return out
}

func cmdWindowLog(c *wicore.CommandImpl, e wicore.Editor, w wicore.Window, args ...string) {
	root := wicore.RootWindow(w)
	log.Printf("Window tree:\n%s", tree(root))
}
