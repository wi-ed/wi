// Copyright 2013 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package main

import (
	"fmt"
	"github.com/maruel/tulib"
	"github.com/maruel/wi/wi-plugin"
	"github.com/nsf/termbox-go"
	"log"
	"strings"
)

var singleBorder = []rune{'\u2500', '\u2502', '\u250D', '\u2510', '\u2514', '\u2518'}
var doubleBorder = []rune{'\u2550', '\u2551', '\u2554', '\u2557', '\u255a', '\u255d'}

func isEqual(lhs tulib.Rect, rhs tulib.Rect) bool {
	return lhs.X == rhs.X && lhs.Y == rhs.Y && lhs.Width == rhs.Width && lhs.Height == rhs.Height
}

type drawnBorder int

const (
	drawnBorderNone drawnBorder = iota
	drawnBorderLeft
	drawnBorderRight
	drawnBorderTop
	drawnBorderBottom
	drawnBorderAll
)

// window implements wi.Window. It keeps its own buffer of its display.
type window struct {
	parent          *window
	childrenWindows []*window
	windowBuffer    tulib.Buffer // includes the border
	rect            tulib.Rect   // Window Rect as described in wi.Window.Rect().
	clientAreaRect  tulib.Rect   // Usable area within the Window, the part not obscured by borders.
	viewRect        tulib.Rect   // Window View Rect, which is the client area not used by childrenWindows.
	view            wi.View
	docking         wi.DockingType
	border          wi.BorderType
	effectiveBorder drawnBorder       // effectiveBorder automatically collapses borders when the Window Rect is too small and is based on docking.
	fg              termbox.Attribute // Default text color, to be used in borders.
	bg              termbox.Attribute // Default background color, to be used in borders
}

func (w *window) String() string {
	return fmt.Sprintf("Window(%s, %v)", w.View().Title(), w.Rect())
}

// Returns a string representing the tree.
func (w *window) Tree() string {
	// Not the most performant implementation but does the job.
	out := w.String() + "\n"
	for _, child := range w.childrenWindows {
		for _, line := range strings.Split(child.Tree(), "\n") {
			if line != "" {
				out += ("  " + line + "\n")
			}
		}
	}
	return out
}

func (w *window) Parent() wi.Window {
	// TODO(maruel): Understand why this is necessary at all.
	if w.parent != nil {
		return w.parent
	}
	return nil
}

func (w *window) ChildrenWindows() []wi.Window {
	out := make([]wi.Window, len(w.childrenWindows))
	for i, v := range w.childrenWindows {
		out[i] = v
	}
	return out
}

func (w *window) NewChildWindow(view wi.View, docking wi.DockingType) wi.Window {
	log.Printf("Window(%s).NewChildWindow(%s, %s)", w.View().Title(), view.Title(), docking)
	for _, child := range w.childrenWindows {
		if child.Docking() == docking {
			panic("TODO(maruel): Likely not a panic, maybe a fallback?")
			return nil
		}
	}
	child := makeWindow(w, view, docking)
	w.childrenWindows = append(w.childrenWindows, child)
	w.resizeChildren()
	return child
}

func (w *window) Remove(child wi.Window) {
	for i, v := range w.childrenWindows {
		if v == child {
			copy(w.childrenWindows[i:], w.childrenWindows[i+1:])
			w.childrenWindows[len(w.childrenWindows)-1] = nil
			w.childrenWindows = w.childrenWindows[:len(w.childrenWindows)-1]
			return
		}
	}
	panic("Trying to remove a non-child Window")
}

func (w *window) Rect() tulib.Rect {
	return w.rect
}

func (w *window) SetRect(rect tulib.Rect) {
	// SetRect() recreates the buffer and immediately draws the borders.
	if !isEqual(w.rect, rect) {
		w.rect = rect
		// Internal consistency check.
		if w.parent != nil {
			if !w.rect.FitsIn(w.parent.clientAreaRect) {
				panic(fmt.Sprintf("Child %v doesn't fit parent's client area %v: %v; %v", w, w.parent, w.parent.clientAreaRect, w.rect.Intersection(w.parent.clientAreaRect)))
			}
		}

		w.windowBuffer = tulib.NewBuffer(w.rect.Width, w.rect.Height)
		w.updateBorder()
	}
	// Still flow the call through children Window, so DockingFloating are
	// properly updated.
	w.resizeChildren()
}

// calculateEffectiveBorder calculates window.effectiveBorder.
func calculateEffectiveBorder(r tulib.Rect, d wi.DockingType) drawnBorder {
	switch d {
	case wi.DockingFill:
		return drawnBorderNone

	case wi.DockingFloating:
		if r.Width >= 5 && r.Height >= 3 {
			return drawnBorderAll
		}
		return drawnBorderNone

	case wi.DockingLeft:
		if r.Width > 1 && r.Height > 0 {
			return drawnBorderRight
		}
		return drawnBorderNone

	case wi.DockingRight:
		if r.Width > 1 && r.Height > 0 {
			return drawnBorderLeft
		}
		return drawnBorderNone

	case wi.DockingTop:
		if r.Height > 1 && r.Width > 0 {
			return drawnBorderBottom
		}
		return drawnBorderNone

	case wi.DockingBottom:
		if r.Height > 1 && r.Width > 0 {
			return drawnBorderTop
		}
		return drawnBorderNone

	default:
		panic("Unknown DockingType")
	}
}

// resizeChildren() resizes all the children Window.
func (w *window) resizeChildren() {
	log.Printf("%s.resizeChildren()", w)
	// When borders are used, w.clientAreaRect.X and .Y are likely 1.
	remaining := w.clientAreaRect
	var fill wi.Window
	for _, child := range w.childrenWindows {
		switch child.Docking() {
		case wi.DockingFill:
			fill = child

		case wi.DockingFloating:
			// Floating uses its own thing.
			child.SetRect(child.Rect())

		case wi.DockingLeft:
			width, _ := child.View().NaturalSize()
			if width >= remaining.Width {
				width = remaining.Width
			} else if child.border != wi.BorderNone {
				width += 1
			}
			tmp := remaining
			tmp.Width = width
			remaining.X += width
			remaining.Width -= width
			child.SetRect(tmp)

		case wi.DockingRight:
			width, _ := child.View().NaturalSize()
			if width >= remaining.Width {
				width = remaining.Width
			} else if child.border != wi.BorderNone {
				width += 1
			}
			tmp := remaining
			tmp.X += (remaining.Width - width)
			tmp.Width = width
			remaining.Width -= width
			child.SetRect(tmp)

		case wi.DockingTop:
			_, height := child.View().NaturalSize()
			if height >= remaining.Height {
				height = remaining.Height
			} else if child.border != wi.BorderNone {
				height += 1
			}
			tmp := remaining
			tmp.Height = height
			remaining.Y += height
			remaining.Height -= height
			child.SetRect(tmp)

		case wi.DockingBottom:
			_, height := child.View().NaturalSize()
			if height >= remaining.Height {
				height = remaining.Height
			} else if child.border != wi.BorderNone {
				height += 1
			}
			tmp := remaining
			tmp.Y += (remaining.Height - height)
			tmp.Height = height
			remaining.Height -= height
			child.SetRect(tmp)

		default:
			panic("Fill me")
		}
	}
	if fill != nil {
		fill.SetRect(remaining)
		w.viewRect.X = 0
		w.viewRect.Y = 0
		w.viewRect.Width = 0
		w.viewRect.Height = 0
		w.view.SetSize(0, 0)
	} else {
		w.viewRect = remaining
		w.view.SetSize(w.viewRect.Width, w.viewRect.Height)
	}
}

func (w *window) Buffer() *tulib.Buffer {
	if w.viewRect.Width != 0 && w.viewRect.Height != 0 {
		w.windowBuffer.Blit(w.viewRect, 0, 0, w.view.Buffer())
	}
	return &w.windowBuffer
}

func (w *window) Docking() wi.DockingType {
	return w.docking
}

func (w *window) SetDocking(docking wi.DockingType) {
	if w.docking != docking {
		w.docking = docking
	}
}

func (w *window) SetView(view wi.View) {
	panic("To test")
	if view != w.view {
		w.view = view
		w.windowBuffer.Fill(w.viewRect, w.cell(' '))
	}
}

// updateBorder calculates w.effectiveBorder, w.clientAreaRect and draws the
// borders right away in the Window's buffer.
//
// It's called by SetRect() and will be called by SetBorder (if ever
// implemented).
func (w *window) updateBorder() {
	if w.border == wi.BorderNone {
		w.effectiveBorder = drawnBorderNone
	} else {
		w.effectiveBorder = calculateEffectiveBorder(w.rect, w.docking)
	}

	s := doubleBorder
	if w.border == wi.BorderSingle {
		s = singleBorder
	}

	switch w.effectiveBorder {
	case drawnBorderNone:
		w.clientAreaRect = tulib.Rect{0, 0, w.rect.Width, w.rect.Height}

	case drawnBorderLeft:
		w.clientAreaRect = tulib.Rect{1, 0, w.rect.Width - 1, w.rect.Height}
		w.windowBuffer.Fill(tulib.Rect{0, 0, 1, w.rect.Height}, w.cell(s[1]))

	case drawnBorderRight:
		w.clientAreaRect = tulib.Rect{0, 0, w.rect.Width - 1, w.rect.Height}
		w.windowBuffer.Fill(tulib.Rect{w.rect.Width - 1, 0, 1, w.rect.Height}, w.cell(s[1]))

	case drawnBorderTop:
		w.clientAreaRect = tulib.Rect{0, 1, w.rect.Width, w.rect.Height - 1}
		w.windowBuffer.Fill(tulib.Rect{0, 0, w.rect.Width, 1}, w.cell(s[0]))

	case drawnBorderBottom:
		w.clientAreaRect = tulib.Rect{0, 0, w.rect.Width, w.rect.Height - 1}
		w.windowBuffer.Fill(tulib.Rect{0, w.rect.Height - 1, w.rect.Width, 1}, w.cell(s[0]))

	case drawnBorderAll:
		w.clientAreaRect = tulib.Rect{1, 1, w.rect.Width - 2, w.rect.Height - 2}
		// Corners.
		w.windowBuffer.Set(0, 0, w.cell(s[2]))
		w.windowBuffer.Set(0, w.rect.Height-1, w.cell(s[4]))
		w.windowBuffer.Set(w.rect.Width-1, 0, w.cell(s[3]))
		w.windowBuffer.Set(w.rect.Width-1, w.rect.Height-1, w.cell(s[5]))
		// Lines.
		w.windowBuffer.Fill(tulib.Rect{1, 0, w.rect.Width - 2, 1}, w.cell(s[0]))
		w.windowBuffer.Fill(tulib.Rect{1, w.rect.Height - 1, w.rect.Width - 2, w.rect.Height - 1}, w.cell(s[0]))
		w.windowBuffer.Fill(tulib.Rect{0, 1, 1, w.rect.Height - 2}, w.cell(s[1]))
		w.windowBuffer.Fill(tulib.Rect{w.rect.Width - 1, 1, w.rect.Width - 1, w.rect.Height - 2}, w.cell(s[1]))

	default:
		panic("Unknown drawnBorder")
	}

	if w.clientAreaRect.Width < 0 {
		panic("Fix this case")
		w.clientAreaRect.Width = 0
	}
	if w.clientAreaRect.Height < 0 {
		panic("Fix this case")
		w.clientAreaRect.Height = 0
	}
}

func (w *window) cell(r rune) termbox.Cell {
	return termbox.Cell{r, w.fg, w.bg}
}

func (w *window) View() wi.View {
	return w.view
}

func makeWindow(parent *window, view wi.View, docking wi.DockingType) *window {
	return &window{
		parent:  parent,
		view:    view,
		docking: docking,
		//border:  wi.BorderNone,
		border: wi.BorderDouble,
		fg:     termbox.ColorWhite,
		bg:     termbox.ColorBlack,
	}
}

func drawRecurse(w *window, offsetX, offsetY int, out *tulib.Buffer) {
	log.Printf("drawRecurse(%s, %d, %d); %v", w.View().Title(), offsetX, offsetY, w.Rect())
	if w.Docking() == wi.DockingFloating {
		// Floating Window are relative to the screen, not the parent Window.
		offsetX = 0
		offsetY = 0
	}
	// TODO(maruel): Only draw the non-occuled frames!
	dest := w.Rect()
	dest.X += offsetX
	dest.Y += offsetY
	out.Blit(dest, 0, 0, w.Buffer())
	for _, child := range w.childrenWindows {
		drawRecurse(child, dest.X, dest.Y, out)
	}
}
