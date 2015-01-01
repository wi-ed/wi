// generated by go run ../tools/wi-event-generator/main.go -output event_registry_decl.go; DO NOT EDIT

package wicore

import (
	"github.com/maruel/wi/pkg/key"
)

// EventRegistry permits to register callbacks that are called on events.
//
// When the callback returns false, the next registered events are not called.
//
// Warning: This interface is automatically generated.
type EventRegistry interface {
	EventsDefinition

	// Unregister unregisters a callback. Returns an error if the event was not
	// registered.
	Unregister(eventID EventID) error

	RegisterCommands(callback func(a EnqueuedCommands) bool) EventID
	RegisterDocumentCreated(callback func(a Document) bool) EventID
	RegisterDocumentCursorMoved(callback func(a Document, b int, c int) bool) EventID
	RegisterEditorKeyboardModeChanged(callback func(a KeyboardMode) bool) EventID
	RegisterTerminalKeyPressed(callback func(a key.Press) bool) EventID
	RegisterTerminalResized(callback func() bool) EventID
	RegisterViewCreated(callback func(a View) bool) EventID
	RegisterWindowCreated(callback func(a Window) bool) EventID
	RegisterWindowResized(callback func(a Window) bool) EventID
}
