// Copyright 2014 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package wicore

// LanguageMode is a language selection for UI purposes.
// TODO(maruel): The name is poor.
type LanguageMode string

// Known languages.
//
// TODO(maruel): Add new languages when translating the application.
const (
	LangEn LanguageMode = "en"
	LangEs LanguageMode = "es"
	LangFr LanguageMode = "fr"
)

// LangMap is the mapping of strings based on the language.
type LangMap map[LanguageMode]string

// GetStr returns the string for the language if present, defaults to LangEn.
func GetStr(lang LanguageMode, m LangMap) string {
	s, ok := m[lang]
	if !ok {
		return m[LangEn]
	}
	return s
}
