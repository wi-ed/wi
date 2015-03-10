// Copyright 2013 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package editor

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/wi-ed/wi/wicore"
	"github.com/wi-ed/wi/wicore/colors"
	"github.com/wi-ed/wi/wicore/lang"
	"github.com/wi-ed/wi/wicore/raster"
)

var singleBorder = []rune{'\u2500', '\u2502', '\u250D', '\u2510', '\u2514', '\u2518'}
var doubleBorder = []rune{'\u2550', '\u2551', '\u2554', '\u2557', '\u255a', '\u255d'}

type drawnBorder int

const (
	drawnBorderNone drawnBorder = 0
	drawnBorderLeft             = 1 << iota
	drawnBorderRight
	drawnBorderTop
	drawnBorderBottom
	drawnBorderAll = drawnBorderLeft | drawnBorderRight | drawnBorderTop | drawnBorderBottom
)

// window implements wicore.Window. It keeps its own buffer of its display.
//
// window is the only structure guaranteed to be in the wi main process. View
// and Documents can be served via plugins.
type window struct {
	id              int // window ID relative to the parent.
	lastChildID     int // last ID used for a children window.
	parent          *window
	e               wicore.Editor
	childrenWindows []*window
	windowBuffer    *raster.Buffer // includes the border
	rect            raster.Rect    // Window Rect as described in wicore.Window.Rect().
	clientAreaRect  raster.Rect    // Usable area within the Window, the part not obscured by borders.
	viewRect        raster.Rect    // Window View Rect, which is the client area not used by childrenWindows.
	view            wicore.ViewW   // View that renders the content. It may be nil if this Window has no content.
	docking         wicore.DockingType
	border          wicore.BorderType
	effectiveBorder drawnBorder       // effectiveBorder automatically collapses borders when the Window Rect is too small and is based on docking.
	borderFormat    raster.CellFormat // Format to be used in borders. It can be different from .View().DefaultFormat().
}

// wicore.Window interface.

func (w *window) String() string {
	return fmt.Sprintf("Window(%s, %s, %v)", w.ID(), w.View().Title(), w.Rect())
}

func (w *window) ID() string {
	if w.parent == nil {
		// editor.rootWindow.id is always 0.
		return fmt.Sprintf("%d", w.id)
	}
	return fmt.Sprintf("%s:%d", w.parent.ID(), w.id)
}

func (w *window) Parent() wicore.Window {
	// TODO(maruel): Understand why this is necessary at all.
	if w.parent != nil {
		return w.parent
	}
	return nil
}

func (w *window) ChildrenWindows() []wicore.Window {
	out := make([]wicore.Window, len(w.childrenWindows))
	for i, v := range w.childrenWindows {
		out[i] = v
	}
	return out
}

func (w *window) Rect() raster.Rect {
	return w.rect
}

func (w *window) Docking() wicore.DockingType {
	return w.docking
}

func (w *window) View() wicore.View {
	return w.view
}

// Private methods.

// Recursively detach a window tree.
func detachRecursively(w *window) {
	for _, c := range w.childrenWindows {
		detachRecursively(c)
	}
	w.parent = nil
	w.childrenWindows = nil
}

func recurseIDToWindow(w *window, fullID string) *window {
	parts := strings.SplitN(fullID, ":", 2)
	intID, err := strconv.Atoi(parts[0])
	if err != nil {
		// Element is not a valid number, it's an invalid reference.
		return nil
	}
	for _, child := range w.childrenWindows {
		if child.id == intID {
			if len(parts) == 2 {
				return recurseIDToWindow(child, parts[1])
			}
			return child
		}
	}
	return nil
}

// Converts a wicore.Window.ID() to a window pointer. Returns nil if invalid.
//
// "0" is the special reference to the root window.
func (e *editor) idToWindow(id string) *window {
	cur := e.rootWindow
	if id != "0" {
		if !strings.HasPrefix(id, "0:") {
			log.Printf("Invalid id: %s", id)
			return nil
		}
		cur = recurseIDToWindow(cur, id[2:])
	}
	return cur
}

// setRect sets the rect of this Window, based on the parent's Window own
// Rect(). It updates Rect() and synchronously updates the child Window that
// are not DockingFloating.
func (w *window) setRect(rect raster.Rect) {
	// setRect() recreates the buffer and immediately draws the borders.
	if w.rect != rect {
		w.rect = rect
		// Internal consistency check.
		if w.parent != nil {
			if !w.rect.In(w.parent.clientAreaRect) {
				panic(fmt.Sprintf("Child %v doesn't fit parent's client area %v: %v", w, w.parent, w.parent.clientAreaRect))
			}
		}

		w.windowBuffer = raster.NewBuffer(w.rect.Width, w.rect.Height)
		w.updateBorder()
	}
	// Still flow the call through children Window, so DockingFloating are
	// properly updated.
	w.resizeChildren()
}

// calculateEffectiveBorder calculates window.effectiveBorder.
func calculateEffectiveBorder(r raster.Rect, d wicore.DockingType) drawnBorder {
	switch d {
	case wicore.DockingFill:
		return drawnBorderNone

	case wicore.DockingFloating:
		if r.Width >= 5 && r.Height >= 3 {
			return drawnBorderAll
		}
		return drawnBorderNone

	case wicore.DockingLeft:
		if r.Width > 1 && r.Height > 0 {
			return drawnBorderRight
		}
		return drawnBorderNone

	case wicore.DockingRight:
		if r.Width > 1 && r.Height > 0 {
			return drawnBorderLeft
		}
		return drawnBorderNone

	case wicore.DockingTop:
		if r.Height > 1 && r.Width > 0 {
			return drawnBorderBottom
		}
		return drawnBorderNone

	case wicore.DockingBottom:
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
	var fill *window
	for _, child := range w.childrenWindows {
		switch child.Docking() {
		case wicore.DockingFill:
			fill = child

		case wicore.DockingFloating:
			// Floating uses its own thing.
			// TODO(maruel): Not clean. Doesn't handle root Window resize properly.
			child.setRect(child.Rect())

		case wicore.DockingLeft:
			width, _ := child.View().NaturalSize()
			if width >= remaining.Width {
				width = remaining.Width
			} else if child.border != wicore.BorderNone {
				width++
			}
			tmp := remaining
			tmp.Width = width
			remaining.X += width
			remaining.Width -= width
			child.setRect(tmp)

		case wicore.DockingRight:
			width, _ := child.View().NaturalSize()
			if width >= remaining.Width {
				width = remaining.Width
			} else if child.border != wicore.BorderNone {
				width++
			}
			tmp := remaining
			tmp.X += (remaining.Width - width)
			tmp.Width = width
			remaining.Width -= width
			child.setRect(tmp)

		case wicore.DockingTop:
			_, height := child.View().NaturalSize()
			if height >= remaining.Height {
				height = remaining.Height
			} else if child.border != wicore.BorderNone {
				height++
			}
			tmp := remaining
			tmp.Height = height
			remaining.Y += height
			remaining.Height -= height
			child.setRect(tmp)

		case wicore.DockingBottom:
			_, height := child.View().NaturalSize()
			if height >= remaining.Height {
				height = remaining.Height
			} else if child.border != wicore.BorderNone {
				height++
			}
			tmp := remaining
			tmp.Y += (remaining.Height - height)
			tmp.Height = height
			remaining.Height -= height
			child.setRect(tmp)

		default:
			panic("Fill me")
		}
	}
	if fill != nil {
		fill.setRect(remaining)
		w.viewRect.X = 0
		w.viewRect.Y = 0
		w.viewRect.Width = 0
		w.viewRect.Height = 0
		w.view.SetSize(0, 0)
	} else {
		w.viewRect = remaining
		w.view.SetSize(w.viewRect.Width, w.viewRect.Height)
	}
	wicore.PostCommand(w.e, nil, "editor_redraw")
}

func (w *window) buffer() *raster.Buffer {
	// TODO(maruel): Redo API.
	// Opportunistically refresh the view buffer.
	if w.viewRect.Width != 0 && w.viewRect.Height != 0 {
		b := w.windowBuffer.SubBuffer(w.viewRect)
		b.Blit(w.view.Buffer())
	}
	return w.windowBuffer
}

/* TODO(maruel): Is it needed at all?
func (w *window) setView(view wicore.View) {
	if view != w.view {
		w.view = view
		b := w.windowBuffer.SubBuffer(w.viewRect)
		b.Fill(w.cell(' '))
		wicore.PostCommand(w, "editor_redraw")
	}
	panic("To test")
}
*/

// updateBorder calculates w.effectiveBorder, w.clientAreaRect and draws the
// borders right away in the Window's buffer.
//
// It's called by setRect() and will be called by SetBorder (if ever
// implemented).
func (w *window) updateBorder() {
	if w.border == wicore.BorderNone {
		w.effectiveBorder = drawnBorderNone
	} else {
		w.effectiveBorder = calculateEffectiveBorder(w.rect, w.docking)
	}

	s := doubleBorder
	if w.border == wicore.BorderSingle {
		s = singleBorder
	}

	// TODO(maruel): Switch to a bitmask check by incrementally reducing w.clientAreaRect.
	switch w.effectiveBorder {
	case drawnBorderNone:
		w.clientAreaRect = raster.Rect{0, 0, w.rect.Width, w.rect.Height}

	case drawnBorderLeft:
		w.clientAreaRect = raster.Rect{1, 0, w.rect.Width - 1, w.rect.Height}
		w.windowBuffer.SubBuffer(raster.Rect{0, 0, 1, w.rect.Height}).Fill(w.cell(s[1]))

	case drawnBorderRight:
		w.clientAreaRect = raster.Rect{0, 0, w.rect.Width - 1, w.rect.Height}
		w.windowBuffer.SubBuffer(raster.Rect{w.rect.Width - 1, 0, 1, w.rect.Height}).Fill(w.cell(s[1]))

	case drawnBorderTop:
		w.clientAreaRect = raster.Rect{0, 1, w.rect.Width, w.rect.Height - 1}
		w.windowBuffer.SubBuffer(raster.Rect{0, 0, w.rect.Width, 1}).Fill(w.cell(s[0]))

	case drawnBorderBottom:
		w.clientAreaRect = raster.Rect{0, 0, w.rect.Width, w.rect.Height - 1}
		w.windowBuffer.SubBuffer(raster.Rect{0, w.rect.Height - 1, w.rect.Width, 1}).Fill(w.cell(s[0]))

	case drawnBorderAll:
		w.clientAreaRect = raster.Rect{1, 1, w.rect.Width - 2, w.rect.Height - 2}
		// Corners.
		*w.windowBuffer.Cell(0, 0) = w.cell(s[2])
		*w.windowBuffer.Cell(0, w.rect.Height-1) = w.cell(s[4])
		*w.windowBuffer.Cell(w.rect.Width-1, 0) = w.cell(s[3])
		*w.windowBuffer.Cell(w.rect.Width-1, w.rect.Height-1) = w.cell(s[5])
		// Lines.
		w.windowBuffer.SubBuffer(raster.Rect{1, 0, w.rect.Width - 2, 1}).Fill(w.cell(s[0]))
		w.windowBuffer.SubBuffer(raster.Rect{1, w.rect.Height - 1, w.rect.Width - 2, w.rect.Height - 1}).Fill(w.cell(s[0]))
		w.windowBuffer.SubBuffer(raster.Rect{0, 1, 1, w.rect.Height - 2}).Fill(w.cell(s[1]))
		w.windowBuffer.SubBuffer(raster.Rect{w.rect.Width - 1, 1, w.rect.Width - 1, w.rect.Height - 2}).Fill(w.cell(s[1]))

	default:
		panic("Unknown drawnBorder")
	}

	if w.clientAreaRect.Width < 0 {
		w.clientAreaRect.Width = 0
		panic("Fix this case")
	}
	if w.clientAreaRect.Height < 0 {
		w.clientAreaRect.Height = 0
		panic("Fix this case")
	}
}

func (w *window) getBorderFormat() raster.CellFormat {
	c := w.borderFormat
	if c.Empty() {
		// Defaults to the view format.
		c = w.view.DefaultFormat()
		if c.Empty() && w.parent != nil {
			// Defaults to the parent's format.
			c = w.parent.getBorderFormat()
		}
	}
	return c
}

func (w *window) cell(r rune) raster.Cell {
	return raster.Cell{r, w.getBorderFormat()}
}

func makeWindow(parent *window, view wicore.ViewW, docking wicore.DockingType) *window {
	log.Printf("makeWindow(%s, %s, %s)", parent, view.Title(), docking)
	var e wicore.Editor
	id := 0
	if parent != nil {
		e = parent.e
		parent.lastChildID++
		id = parent.lastChildID
	}
	// It's more complex than that but it's a fine default.
	border := wicore.BorderNone
	if docking == wicore.DockingFloating {
		border = wicore.BorderDouble
	}
	return &window{
		id:      id,
		parent:  parent,
		e:       e,
		view:    view,
		docking: docking,
		border:  border,
		borderFormat: raster.CellFormat{
			Fg: colors.White,
			Bg: colors.Black,
		},
	}
}

// drawRecurse recursively draws the Window tree into buffer out.
func drawRecurse(w *window, offsetX, offsetY int, out *raster.Buffer) {
	log.Printf("drawRecurse(%s, %d, %d); %v", w.View().Title(), offsetX, offsetY, w.Rect())
	if w.Docking() == wicore.DockingFloating {
		// Floating Window are relative to the screen, not the parent Window.
		offsetX = 0
		offsetY = 0
	}
	// TODO(maruel): Only draw non-occuled Windows!
	dest := w.Rect()
	dest.X += offsetX
	dest.Y += offsetY
	out.SubBuffer(dest).Blit(w.buffer())

	fillFound := false
	for _, child := range w.childrenWindows {
		// In the case of DockingFill, only the first one should be drawn. In
		// particular, the DockingFloating child of an hidden DockingFill will not
		// be drawn.
		if child.docking == wicore.DockingFill {
			if fillFound {
				continue
			}
			fillFound = true
		}
		drawRecurse(child, dest.X, dest.Y, out)
	}
}

// Commands

func cmdWindowActivate(c *privilegedCommandImpl, e *editor, w *window, args ...string) {
	windowName := args[0]

	child := e.idToWindow(windowName)
	if child == nil {
		e.ExecuteCommand(w, "alert", isNotValidWindow.Sprintf(windowName))
		return
	}
	e.activateWindow(child)
}

func cmdWindowClose(c *privilegedCommandImpl, e *editor, w *window, args ...string) {
	windowName := args[0]

	child := e.idToWindow(windowName)
	if child == nil {
		e.ExecuteCommand(w, "alert", isNotValidWindow.Sprintf(windowName))
		return
	}
	for i, v := range child.parent.childrenWindows {
		if v == child {
			copy(w.childrenWindows[i:], w.childrenWindows[i+1:])
			w.childrenWindows[len(w.childrenWindows)-1] = nil
			w.childrenWindows = w.childrenWindows[:len(w.childrenWindows)-1]
			detachRecursively(v)
			wicore.PostCommand(e, nil, "editor_redraw")
			return
		}
	}
}

func cmdWindowNew(c *privilegedCommandImpl, e *editor, w *window, args ...string) {
	windowName := args[0]
	dockingName := args[1]
	viewFactoryName := args[2]

	parent := e.idToWindow(windowName)
	if parent == nil {
		if viewFactoryName != "infobar_alert" {
			e.ExecuteCommand(w, "alert", isNotValidWindow.Sprintf(windowName))
		}
		return
	}

	docking := wicore.StringToDockingType(dockingName)
	if docking == wicore.DockingUnknown {
		if viewFactoryName != "infobar_alert" {
			e.ExecuteCommand(w, "alert", invalidDocking.Sprintf(dockingName))
		}
		return
	}
	// TODO(maruel): Only the first child Window with DockingFill is visible.
	// TODO(maruel): Reorder .childrenWindows with
	// CommandDispatcherFull.ActivateWindow() but only with DockingFill.
	// TODO(maruel): Also allow DockingFloating.
	//if docking != wicore.DockingFill {
	for _, child := range parent.childrenWindows {
		if child.Docking() == docking {
			if viewFactoryName != "infobar_alert" {
				e.ExecuteCommand(w, "alert", cantAddTwoWindowWithSameDocking.Sprintf(docking))
			}
			return
		}
	}
	//}

	viewFactory, ok := e.viewFactories[viewFactoryName]
	if !ok {
		if viewFactoryName != "infobar_alert" {
			e.ExecuteCommand(w, "alert", invalidViewFactory.Sprintf(viewFactoryName))
		}
		return
	}
	// TODO(maruel): e.nextViewID is an implementation detail, it's wrong.
	view := viewFactory(e, e.nextViewID, args[3:]...)
	e.nextViewID++

	child := makeWindow(parent, view, docking)
	if docking == wicore.DockingFloating {
		width, height := view.NaturalSize()
		if child.border != wicore.BorderNone {
			width += 2
			height += 2
		}
		// TODO(maruel): Handle when width or height > scren size.
		// TODO(maruel): Not clean. Doesn't handle root Window resize properly.
		rootRect := e.rootWindow.Rect()
		child.rect.X = (rootRect.Width - width - 1) / 2
		child.rect.Y = (rootRect.Height - height - 1) / 2
		child.rect.Width = width
		child.rect.Height = height
	}
	parent.childrenWindows = append(parent.childrenWindows, child)
	parent.resizeChildren()
	// Call OnAttach() after the Window is attached to the parent.
	view.OnAttach(child)
	e.activateWindow(child)
}

func cmdWindowSetDocking(c *privilegedCommandImpl, e *editor, w *window, args ...string) {
	windowName := args[0]
	dockingName := args[1]

	child := e.idToWindow(windowName)
	if child == nil {
		e.ExecuteCommand(w, "alert", isNotValidWindow.Sprintf(windowName))
		return
	}
	docking := wicore.StringToDockingType(dockingName)
	if docking == wicore.DockingUnknown {
		e.ExecuteCommand(w, "alert", invalidDocking.Sprintf(dockingName))
		return
	}
	if w.docking != docking {
		// TODO(maruel): Check no other parent's child window have the same dock.
		w.docking = docking
		w.parent.resizeChildren()
		wicore.PostCommand(w.e, nil, "editor_redraw")
	}
}

func cmdWindowSetRect(c *privilegedCommandImpl, e *editor, w *window, args ...string) {
	windowName := args[0]

	child := e.idToWindow(windowName)
	if child == nil {
		e.ExecuteCommand(w, "alert", isNotValidWindow.Sprintf(windowName))
		return
	}
	r := raster.Rect{}
	var err1, err2, err3, err4 error
	r.X, err1 = strconv.Atoi(args[1])
	r.Y, err2 = strconv.Atoi(args[2])
	r.Width, err3 = strconv.Atoi(args[3])
	r.Height, err4 = strconv.Atoi(args[4])
	if err1 != nil || err2 != nil || err3 != nil || err4 != nil {
		e.ExecuteCommand(w, "alert", invalidRect.Sprintf(args[1], args[2], args[3], args[4]))
		return
	}
	child.setRect(r)
}

// RegisterWindowCommands registers all the commands relative to window
// management.
func RegisterWindowCommands(dispatcher wicore.CommandsW) {
	cmds := []wicore.Command{
		&privilegedCommandImpl{
			"window_activate",
			1,
			cmdWindowActivate,
			wicore.WindowCategory,
			lang.Map{
				lang.En: "Activate a window",
			},
			lang.Map{
				lang.En: "Active a window. This means the Window will have keyboard focus.",
			},
		},
		&privilegedCommandImpl{
			"window_close",
			1,
			cmdWindowClose,
			wicore.WindowCategory,
			lang.Map{
				lang.En: "Closes a window",
			},
			lang.Map{
				lang.En: "Closes a window. Note that any window can be closed and all the child window will be destroyed at the same time.",
			},
		},
		&privilegedCommandImpl{
			"window_new",
			-1,
			cmdWindowNew,
			wicore.WindowCategory,
			lang.Map{
				lang.En: "Creates a new window",
			},
			lang.Map{
				lang.En: "Usage: window_new <parent> <docking> <view name> <view args...>\nCreates a new window. The new window is created as a child to the specified parent. It creates inside the window the view specified. The Window is activated. It is invalid to add a child Window with the same docking as one already present.",
			},
		},
		&privilegedCommandImpl{
			"window_set_docking",
			2,
			cmdWindowSetDocking,
			wicore.WindowCategory,
			lang.Map{
				lang.En: "Change the docking of a window",
			},
			lang.Map{
				lang.En: "Changes the docking of this Window relative to the parent window. This will forces an invalidation and a redraw.",
			},
		},
		&privilegedCommandImpl{
			"window_set_rect",
			5,
			cmdWindowSetRect,
			wicore.WindowCategory,
			lang.Map{
				lang.En: "Move a window",
			},
			lang.Map{
				lang.En: "Usage: window_set_rect <window> <x> <y> <w> <h>\nMoves a Window relative to the parent window, unless it is floating, where it is relative to the view port.",
			},
		},
	}
	for _, cmd := range cmds {
		dispatcher.Register(cmd)
	}
}
