// Copyright 2014 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package main

import (
	"github.com/maruel/wi/wi-plugin"
)

type langMap map[wi.LanguageMode]string

var activateDisabled = langMap{
	wi.LangEn: "Can't activate a disabled view.",
}

var aliasFor = langMap{
	wi.LangEn: "Alias for \"%s\".",
}

var aliasNotFound = langMap{
	wi.LangEn: "\"%s\" is an alias to command \"%s\" but this command is not registered.",
}

var notFound = langMap{
	wi.LangEn: "Command \"%s\" is not registered.",
}

var viewDirty = langMap{
	wi.LangEn: "View \"%s\" is not saved, aborting quit.",
}

// getStr returns the string for the language if present, defaults to wi.LangEn.
func getStr(lang wi.LanguageMode, m langMap) string {
	s, ok := m[lang]
	if !ok {
		return m[wi.LangEn]
	}
	return s
}
