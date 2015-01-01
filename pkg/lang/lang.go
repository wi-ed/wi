// Copyright 2014 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// package lang handles localization of language UI.
package lang

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
		return m[En]
	}
	return s
}
