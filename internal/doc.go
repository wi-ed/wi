// Copyright 2015 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package internal contains the symbols that need to be exported so it can be
// used in RPC via net/rpc but is not meant to be an API for end user plugins.
//
// It is based on https://golang.org/s/go14internal design principle.
package internal
