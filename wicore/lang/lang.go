// Copyright 2014 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package lang handles localization of language UI.
package lang

import (
	"fmt"
	"strings"
	"sync"
)

var lock sync.Mutex
var active = En

// Language is used to declare the language for UI purpose.
type Language string

// Known languages.
//
// TODO(maruel): Add new languages when translating the application.
const (
	En   Language = "en"
	Es   Language = "es"
	Fr   Language = "fr"
	FrCa Language = "fr_ca"
)

// Map is the mapping of strings based on the language.
type Map map[Language]string

// Get returns the string for the language if present, defaults to En.
func (m Map) Get(lang Language) string {
	s, ok := m[lang]
	if !ok {
		items := strings.Split(string(lang), "_")
		s, ok = m[Language(items[0])]
		if !ok {
			s, ok = m[En]
		}
	}
	return s
}

func (m Map) String() string {
	return m.Get(Active())
}

// Formatf returns fmt.Sprintf(m.String(), args...) as a shortcut.
func (m Map) Formatf(args ...interface{}) string {
	return fmt.Sprintf(m.String(), args...)
}

// Active returns the active language for the process.
func Active() Language {
	lock.Lock()
	defer lock.Unlock()
	return active
}

// Set sets the active language for the process.
func Set(lang Language) {
	lock.Lock()
	defer lock.Unlock()
	active = lang
}
