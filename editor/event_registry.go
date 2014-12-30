// Copyright 2014 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package editor

import (
	"errors"
	"sync"

	"github.com/maruel/wi/wicore"
)

type eventRegistry struct {
	lock                sync.Mutex
	nextID              wicore.EventID
	documentCreated     map[wicore.EventID]func(doc wicore.Document)
	documentCursorMoved map[wicore.EventID]func(doc wicore.Document)
	terminalResized     map[wicore.EventID]func()
	terminalKeyPressed  map[wicore.EventID]func(key wicore.KeyPress)
	viewCreated         map[wicore.EventID]func(view wicore.View)
	windowCreated       map[wicore.EventID]func(window wicore.Window)
	windowResized       map[wicore.EventID]func(window wicore.Window)
}

func (e *eventRegistry) Unregister(eventID wicore.EventID) error {
	e.lock.Lock()
	defer e.lock.Unlock()
	// TODO(maruel): Make something more reasonable and less error-prone.
	if _, ok := e.documentCreated[eventID]; ok {
		delete(e.documentCreated, eventID)
		return nil
	}
	if _, ok := e.documentCursorMoved[eventID]; ok {
		delete(e.documentCursorMoved, eventID)
		return nil
	}
	if _, ok := e.terminalResized[eventID]; ok {
		delete(e.terminalResized, eventID)
		return nil
	}
	if _, ok := e.terminalKeyPressed[eventID]; ok {
		delete(e.terminalKeyPressed, eventID)
		return nil
	}
	if _, ok := e.viewCreated[eventID]; ok {
		delete(e.viewCreated, eventID)
		return nil
	}
	if _, ok := e.windowCreated[eventID]; ok {
		delete(e.windowCreated, eventID)
		return nil
	}
	if _, ok := e.windowResized[eventID]; ok {
		delete(e.windowResized, eventID)
		return nil
	}
	return errors.New("trying to unregister an non existing event listener")
}

// TODO(maruel): Use "go generate" to reduce the copy-pasta.

func (e *eventRegistry) RegisterDocumentCreated(callback func(doc wicore.Document)) wicore.EventID {
	e.lock.Lock()
	defer e.lock.Unlock()
	i := e.nextID
	e.nextID++
	e.documentCreated[i] = callback
	return i
}

func (e *eventRegistry) onDocumentCreated(doc wicore.Document) {
	items := func() []func(doc wicore.Document) {
		e.lock.Lock()
		defer e.lock.Unlock()
		items := make([]func(doc wicore.Document), 0, len(e.documentCreated))
		for _, c := range e.documentCreated {
			items = append(items, c)
		}
		return items
	}()
	for _, item := range items {
		item(doc)
	}
}

func (e *eventRegistry) RegisterDocumentCursorMoved(callback func(doc wicore.Document)) wicore.EventID {
	e.lock.Lock()
	defer e.lock.Unlock()
	i := e.nextID
	e.nextID++
	e.documentCursorMoved[i] = callback
	return i
}

func (e *eventRegistry) onDocumentCursorMoved(doc wicore.Document) {
	items := func() []func(doc wicore.Document) {
		e.lock.Lock()
		defer e.lock.Unlock()
		items := make([]func(doc wicore.Document), 0, len(e.documentCursorMoved))
		for _, c := range e.documentCursorMoved {
			items = append(items, c)
		}
		return items
	}()
	for _, item := range items {
		item(doc)
	}
}

func (e *eventRegistry) RegisterTerminalResized(callback func()) wicore.EventID {
	e.lock.Lock()
	defer e.lock.Unlock()
	i := e.nextID
	e.nextID++
	e.terminalResized[i] = callback
	return i
}

func (e *eventRegistry) onTerminalResized() {
	items := func() []func() {
		e.lock.Lock()
		defer e.lock.Unlock()
		items := make([]func(), 0, len(e.terminalResized))
		for _, c := range e.terminalResized {
			items = append(items, c)
		}
		return items
	}()
	for _, item := range items {
		item()
	}
}

func (e *eventRegistry) RegisterTerminalKeyPressed(callback func(key wicore.KeyPress)) wicore.EventID {
	e.lock.Lock()
	defer e.lock.Unlock()
	i := e.nextID
	e.nextID++
	e.terminalKeyPressed[i] = callback
	return i
}

func (e *eventRegistry) onTerminalKeyPressed(key wicore.KeyPress) {
	items := func() []func(key wicore.KeyPress) {
		e.lock.Lock()
		defer e.lock.Unlock()
		items := make([]func(key wicore.KeyPress), 0, len(e.terminalKeyPressed))
		for _, c := range e.terminalKeyPressed {
			items = append(items, c)
		}
		return items
	}()
	for _, item := range items {
		item(key)
	}
}

func (e *eventRegistry) RegisterViewCreated(callback func(view wicore.View)) wicore.EventID {
	e.lock.Lock()
	defer e.lock.Unlock()
	i := e.nextID
	e.nextID++
	e.viewCreated[i] = callback
	return i
}

func (e *eventRegistry) onViewCreated(view wicore.View) {
	items := func() []func(view wicore.View) {
		e.lock.Lock()
		defer e.lock.Unlock()
		items := make([]func(view wicore.View), 0, len(e.viewCreated))
		for _, c := range e.viewCreated {
			items = append(items, c)
		}
		return items
	}()
	for _, item := range items {
		item(view)
	}
}

func (e *eventRegistry) RegisterWindowCreated(callback func(window wicore.Window)) wicore.EventID {
	e.lock.Lock()
	defer e.lock.Unlock()
	i := e.nextID
	e.nextID++
	e.windowCreated[i] = callback
	return i
}

func (e *eventRegistry) onWindowCreated(window wicore.Window) {
	items := func() []func(window wicore.Window) {
		e.lock.Lock()
		defer e.lock.Unlock()
		items := make([]func(window wicore.Window), 0, len(e.windowCreated))
		for _, c := range e.windowCreated {
			items = append(items, c)
		}
		return items
	}()
	for _, item := range items {
		item(window)
	}
}

func (e *eventRegistry) RegisterWindowResized(callback func(window wicore.Window)) wicore.EventID {
	e.lock.Lock()
	defer e.lock.Unlock()
	i := e.nextID
	e.nextID++
	e.windowResized[i] = callback
	return i
}

func (e *eventRegistry) onWindowResized(window wicore.Window) {
	items := func() []func(window wicore.Window) {
		e.lock.Lock()
		defer e.lock.Unlock()
		items := make([]func(window wicore.Window), 0, len(e.windowResized))
		for _, c := range e.windowResized {
			items = append(items, c)
		}
		return items
	}()
	for _, item := range items {
		item(window)
	}
}
