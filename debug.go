// Copyright 2014 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Use "go build -tags debug" to have access to the code in this file.

// +build debug

package main

import (
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
	"time"

	"github.com/maruel/wi/editor"
	"github.com/maruel/wi/wicore"
	"github.com/nsf/termbox-go"
)

var (
	crash      = flag.Duration("crash", 0, "Crash after specified duration")
	prof       = flag.String("http", "", "Start a profiling web server")
	cpuprofile = flag.String("cpuprofile", "", "Write cpu profile to file; use \"go tool pprof wi <file>\" to read the data; See https://blog.golang.org/profiling-go-programs for more details")
)

type onDebugClose struct {
	logFile  io.Closer
	profFile io.Closer
}

func (o onDebugClose) Close() error {
	if o.logFile != nil {
		o.logFile.Close()
	}
	if o.profFile != nil {
		pprof.StopCPUProfile()
		o.profFile.Close()
	}
	return nil
}

func debugHook() io.Closer {
	o := onDebugClose{}
	log.SetFlags(log.Lmicroseconds | log.Lshortfile)
	if f, err := os.OpenFile("wi.log", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666); err == nil {
		o.logFile = f
		log.SetOutput(f)
	}

	if *cpuprofile != "" {
		if f, err := os.OpenFile(*cpuprofile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666); err == nil {
			o.profFile = f
			pprof.StartCPUProfile(f)
		} else {
			log.Printf("Failed to open %s: %s", *cpuprofile, err)
			*cpuprofile = ""
		}
	}

	// TODO(maruel): Investigate adding our own profiling for RPC.
	// http://golang.org/pkg/runtime/pprof/
	// TODO(maruel): Add pprof.WriteHeapProfile(f) when desired (?)

	if *crash > 0 {
		// Crashes but ensure that the terminal is closed first. It's useful to
		// figure out what's happening with an infinite loop for example.
		time.AfterFunc(*crash, func() {
			o.Close()
			termbox.Close()
			panic("Timeout")
		})
	}

	if *prof != "" {
		http.HandleFunc("/", rootHandler)
		go func() {
			log.Println(http.ListenAndServe(*prof, nil))
		}()
	}
	return o
}

func debugHookEditor(e editor.Editor) {
	expvar.Publish("active_window", Func(func() string { return e.ActiveWindow().String() }))
	expvar.Publish("commands", FuncJSON(func() interface{} { return commands(e) }))
	expvar.Publish("view_factories", FuncJSON(func() interface{} { return viewFactories(e) }))
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
		<title>wi</title>
		<meta charset="utf-8">
		<style>
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
	<ul>
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

type Func func() string

func (f Func) String() string {
	return f()
}

type FuncJSON func() interface{}

func (f FuncJSON) String() string {
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

func viewFactories(e wicore.Editor) interface{} {
	names := e.ViewFactoryNames()
	sort.Strings(names)
	return names
}
