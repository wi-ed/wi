// Copyright 2013 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package editor

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
	// TODO(maruel): For combo box (e.g. drop down list of suggestions), it
	// should be drawBorderLeftBottomRight.

	drawnBorderNone drawnBorder = iota
	drawnBorderLeft
	drawnBorderRight
	drawnBorderTop
	drawnBorderBottom
	drawnBorderAll
)

// window implements wi.Window. It keeps its own buffer of its display.
type window struct {
	id              int
	nextChildId     int
	parent          *window
	cd              wi.CommandDispatcherFull
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

func (w *window) Id() string {
	return fmt.Sprintf("%s:%d", w.parent.Id(), w.id)
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
	log.Printf("%s.NewChildWindow(%s, %s)", w, view.Title(), docking)
	// Only the first child Window with DockingFill is visible.
	// TODO(maruel): Reorder .childrenWindows with
	// CommandDispatcherFull.ActivateWindow() but only with DockingFill.
	// TODO(maruel): Also allow DockingFloating.
	if docking != wi.DockingFill {
		for _, child := range w.childrenWindows {
			if child.Docking() == docking {
				panic("TODO(maruel): Likely not a panic, maybe a fallback?")
				return nil
			}
		}
	}
	child := makeWindow(w, view, docking)
	if docking == wi.DockingFloating {
		width, height := view.NaturalSize()
		if child.border != wi.BorderNone {
			width += 2
			height += 2
		}
		// TODO(maruel): Handle when width or height > scren size.
		// TODO(maruel): Not clean. Doesn't handle root Window resize properly.
		rootRect := wi.RootWindow(w).Rect()
		child.rect.X = (rootRect.Width - width - 1) / 2
		child.rect.Y = (rootRect.Height - height - 1) / 2
		child.rect.Width = width
		child.rect.Height = height
	}
	w.childrenWindows = append(w.childrenWindows, child)
	w.resizeChildren()
	return child
}

// Recursively detach a window tree.
func detachRecursively(w *window) {
	for _, c := range w.childrenWindows {
		detachRecursively(c)
	}
	w.parent = nil
	w.childrenWindows = nil
}

func (w *window) Remove(child wi.Window) {
	for i, v := range w.childrenWindows {
		if v == child {
			copy(w.childrenWindows[i:], w.childrenWindows[i+1:])
			w.childrenWindows[len(w.childrenWindows)-1] = nil
			w.childrenWindows = w.childrenWindows[:len(w.childrenWindows)-1]
			detachRecursively(v)
			w.cd.PostDraw()
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
			// TODO(maruel): Not clean. Doesn't handle root Window resize properly.
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
	w.cd.PostDraw()
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
		w.cd.PostDraw()
	}
}

func (w *window) SetView(view wi.View) {
	panic("To test")
	if view != w.view {
		w.view = view
		w.windowBuffer.Fill(w.viewRect, w.cell(' '))
		w.cd.PostDraw()
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
	var cd wi.CommandDispatcherFull
	id := 0
	if parent != nil {
		cd = parent.cd
		id = parent.nextChildId
		parent.nextChildId += 1
	}
	// It's more complex than that but it's a fine default.
	border := wi.BorderNone
	if docking == wi.DockingFloating {
		border = wi.BorderDouble
	}
	return &window{
		id:      id,
		parent:  parent,
		cd:      cd,
		view:    view,
		docking: docking,
		border:  border,
		fg:      termbox.ColorWhite,
		bg:      termbox.ColorBlack,
	}
}

// drawRecurse recursively draws the Window tree into buffer out.
func drawRecurse(w *window, offsetX, offsetY int, out *tulib.Buffer) {
	log.Printf("drawRecurse(%s, %d, %d); %v", w.View().Title(), offsetX, offsetY, w.Rect())
	if w.Docking() == wi.DockingFloating {
		// Floating Window are relative to the screen, not the parent Window.
		offsetX = 0
		offsetY = 0
	}
	// TODO(maruel): Only draw non-occuled Windows!
	dest := w.Rect()
	dest.X += offsetX
	dest.Y += offsetY
	out.Blit(dest, 0, 0, w.Buffer())

	fillFound := false
	for _, child := range w.childrenWindows {
		// In the case of DockingFill, only the first one should be drawn. In
		// particular, the DockingFloating child of an hidden DockingFill will not
		// be drawn.
		if child.docking == wi.DockingFill {
			if fillFound {
				continue
			}
			fillFound = true
		}
		drawRecurse(child, dest.X, dest.Y, out)
	}
}

// Commands

func cmdWindowClose(c *privilegedCommandImpl, e *editor, w *window, args ...string) {
	log.Printf("Faking closing a window: %s", args)
}

func cmdWindowLogTree(c *wi.CommandImpl, cd wi.CommandDispatcherFull, w wi.Window, args ...string) {
	root := wi.RootWindow(w)
	log.Printf("Window tree:\n%s", root.Tree())
}

func cmdWindowNew(c *privilegedCommandImpl, e *editor, w *window, args ...string) {
	/*
		parentName := args[0]
		viewFactoryName := args[1]
		docking := args[2]
		parentWindow := editor.GetWindow(parentName)
		viewFactory := editor.viewFactories[viewFactoryName]
		parentWindow.NewChildWindow(viewFactory(), docking)
	*/
}

func RegisterWindowCommands(e *editor, dispatcher wi.Commands) {
	var windowCommands = []wi.Command{
		&privilegedCommandImpl{
			"window_close",
			1,
			cmdWindowClose,
			wi.WindowCategory,
			wi.LangMap{
				wi.LangEn: "Closes a window",
			},
			wi.LangMap{
				wi.LangEn: "Closes a window. Note that any window can be closed and all the child window will be destroyed at the same time.",
			},
		},
		&wi.CommandImpl{
			"window_log_tree",
			0,
			cmdWindowLogTree,
			wi.DebugCategory,
			wi.LangMap{
				wi.LangEn: "Logs the tree in the log file",
			},
			wi.LangMap{
				wi.LangEn: "Logs the tree in the log file, this is only relevant if -verbose is used.",
			},
		},
		&privilegedCommandImpl{
			"window_new",
			4,
			cmdWindowNew,
			wi.WindowCategory,
			wi.LangMap{
				wi.LangEn: "Creates a new window",
			},
			wi.LangMap{
				wi.LangEn: "Creates a new window. The new window is created as a child to the specified parent. It creates the view specified that was previously registered.",
			},
		},
		// 'screenshot', mainly for unit test; open a new buffer with the screenshot, so it can be saved with 'w'.
	}
	for _, cmd := range windowCommands {
		dispatcher.Register(cmd)
	}
}
