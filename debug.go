// Copyright 2014 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Use "go build -tags debug" to have access to the code in this file.

// +build debug

package main

import (
	"flag"
	"io"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime/pprof"
	"time"

	"github.com/nsf/termbox-go"
)

var (
	crash      = flag.Duration("crash", 0, "Crash after specified duration")
	prof       = flag.String("http", "", "Start a profiling web server; access via <prof>/debug/pprof; see https://golang.org/pkg/net/http/pprof/ for more details")
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
		go func() {
			log.Println(http.ListenAndServe(*prof, nil))
		}()
	}
	return o
}
