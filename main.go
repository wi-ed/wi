// Copyright 2013 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// wi - Bringing text based editor technology past 1200 bauds. See README.md
// for more details.
package main

import (
	"flag"
	"fmt"
	"github.com/maruel/wi/wi-plugin"
	"github.com/nsf/termbox-go"
	"github.com/nsf/tulib"
	"log"
	"os"
	"time"
)

const (
	// Major.Minor.Bugfix. All plugins should be recompiled with wi-plugin
	// changes.
	version = "0.0.1"
)

var quitFlag = false

// UI

type commandQueueItem struct {
	cmdName string
	args    []string
	keyName string
}

// It is normally expected to be drawn via an ssh/mosh connection so it should
// be "bandwidth" optimized, where bandwidth doesn't mean 1200 bauds anymore.
type terminal struct {
	rootWindow     *window
	lastActive     []wi.Window
	terminalEvents <-chan termbox.Event
	viewReady      chan bool // A View.Buffer() is ready to be drawn.
	commandsQueue  chan commandQueueItem
	languageMode   wi.LanguageMode
	keyboardMode   wi.KeyboardMode
}

func (t *terminal) Version() string {
	return version
}

func (t *terminal) PostCommand(cmdName string, args ...string) {
	log.Printf("PostCommand(%s, %s)", cmdName, args)
	t.commandsQueue <- commandQueueItem{cmdName, args, ""}
}

func (t *terminal) postKey(keyName string) {
	log.Printf("PostKey(%s)", keyName)
	t.commandsQueue <- commandQueueItem{keyName: keyName}
}

func (t *terminal) WaitQueueEmpty() {
	panic("Implement me!")
}

func (t *terminal) ExecuteCommand(w wi.Window, cmdName string, args ...string) {
	log.Printf("ExecuteCommand(%s)", cmdName)
	cmd := wi.GetCommand(t, w, cmdName)
	if cmd == nil {
		t.ExecuteCommand(w, "alert", fmt.Sprintf(getStr(t.CurrentLanguage(), notFound), cmdName))
	} else {
		cmd.Handle(t, w, args...)
	}
}

func (t *terminal) CurrentLanguage() wi.LanguageMode {
	return t.languageMode
}

func (t *terminal) KeyboardMode() wi.KeyboardMode {
	return t.keyboardMode
}

func drawRecurse(w *window, offsetX, offsetY int, out *tulib.Buffer) {
	log.Printf("drawRecurse(%s, %d, %d); %v", w.View().Title(), offsetX, offsetY, w.Rect())
	out.Blit(w.Rect(), offsetX, offsetY, w.Buffer())
	// TODO(maruel): Only draw the non-occuled frames!
	for _, child := range w.childrenWindows {
		// TODO(maruel): Handle DockingFloating.
		drawRecurse(child, offsetX+w.rect.X+w.viewRect.X, offsetY+w.rect.Y+w.viewRect.Y, out)
	}
}

// draw descends the whole Window tree and redraw Windows.
func (t *terminal) draw() {
	log.Print("draw()")
	b := tulib.TermboxBuffer()
	drawRecurse(t.rootWindow, 0, 0, &b)
	// TODO(maruel): Determine if Flush() is intelligent and skip the forced draw
	// when nothing changed.
	if err := termbox.Flush(); err != nil {
		panic(err)
	}
}

func (t *terminal) ActiveWindow() wi.Window {
	return t.lastActive[0]
}

func (t *terminal) ActivateWindow(w wi.Window) {
	log.Printf("ActivateWindow(%s)", w.View().Title())
	if w.View().IsDisabled() {
		t.ExecuteCommand(w, "alert", getStr(t.CurrentLanguage(), activateDisabled))
		return
	}

	// First remove w from t.lastActive, second add w as t.lastActive[0].
	// This kind of manual list shuffling is really Go's achille heel.
	// TODO(maruel): There's no way I got it right on the first try without a
	// unit test.
	for i, v := range t.lastActive {
		if v == w {
			if i > 0 {
				copy(t.lastActive[:i], t.lastActive[1:i+1])
				t.lastActive[0] = w
			}
			return
		}
	}

	// This Window has never been active.
	l := len(t.lastActive)
	t.lastActive = append(t.lastActive, nil)
	copy(t.lastActive[:l], t.lastActive[1:l])
	t.lastActive[0] = w
}

func (t *terminal) ViewReady(v wi.View) {
	go func() {
		t.viewReady <- true
	}()
}

func (t *terminal) onResize() {
	// Resize the Windows. This also invalidates it, which will also force a
	// redraw if the size changed.
	w, h := termbox.Size()
	t.rootWindow.SetRect(tulib.Rect{0, 0, w, h})
}

// eventLoop handles both commands and events from the terminal. This function
// runs in the UI goroutine.
func (t *terminal) eventLoop() int {
	fakeChan := make(chan time.Time)
	var drawTimer <-chan time.Time = time.After(5 * time.Millisecond)
	keyBuffer := ""
	for {
		select {
		case i := <-t.commandsQueue:
			if i.keyName != "" {
				// Convert the key press into a command. The trick is that we don't
				// know the active window, there could be commands already enqueued
				// that will change the active window, so using the active window
				// directly or indirectly here is an incorrect assumption.
				if i.keyName == "Enter" {
					t.ExecuteCommand(t.ActiveWindow(), keyBuffer)
					keyBuffer = ""
				} else {
					cmdName := wi.GetKeyBindingCommand(t, t.KeyboardMode(), i.keyName)
					if cmdName != "" {
						t.ExecuteCommand(t.ActiveWindow(), cmdName)
					} else if len(i.keyName) == 1 {
						keyBuffer += i.keyName
					}
				}
			} else {
				t.ExecuteCommand(t.ActiveWindow(), i.cmdName, i.args...)
			}

		case event := <-t.terminalEvents:
			switch event.Type {
			case termbox.EventKey:
				k := keyEventToName(event)
				if k != "" {
					t.postKey(k)
				}
			case termbox.EventMouse:
				// TODO(maruel): MouseDispatcher. Mouse events are expected to be
				// resolved to the window that is currently active, unlike key presses.
				// Life is inconsistent.
				break
			case termbox.EventResize:
				// The terminal window was resized, resize everything, independent of
				// the enqueued commands.
				t.onResize()
			case termbox.EventError:
				// TODO(maruel): Not sure what situations can trigger this.
				t.PostCommand("alert", event.Err.Error())
			}

		case <-t.viewReady:
			// Taking in account a 60hz frame is 18.8ms, 5ms is going to be generally
			// processed within the same frame. This delaying results in significant
			// bandwidth saving on loading.
			if drawTimer == fakeChan {
				drawTimer = time.After(5 * time.Millisecond)
			}

		case <-drawTimer:
			if quitFlag {
				return 0
			}
			t.draw()
			drawTimer = fakeChan
		}
	}
	return 0
}

// makeEditor creates the Editor object. The root window doesn't have
// anything to view in it. It will contain two child windows, the main content
// window and the status bar.
func makeEditor() *terminal {
	// The root view is important, it defines all the global commands. It is
	// pre-filled with the default native commands and keyboard mapping, and it's
	// up to the plugins to add more global commands on startup.
	rootView := makeView("Root", -1, -1)
	RegisterDefaultCommands(rootView.Commands())

	rootWindow := makeWindow(nil, rootView, wi.DockingFill)
	terminalEvents := make(chan termbox.Event, 32)
	terminal := &terminal{
		terminalEvents: terminalEvents,
		commandsQueue:  make(chan commandQueueItem, 500),
		viewReady:      make(chan bool),
		rootWindow:     rootWindow,
		lastActive:     []wi.Window{rootWindow},
		languageMode:   wi.LangEn,
		keyboardMode:   wi.EditMode,
	}

	RegisterDefaultKeyBindings(terminal)

	terminal.onResize()
	go func() {
		for {
			terminalEvents <- termbox.PollEvent()
		}
	}()
	return terminal
}

// window implements Window. It keeps its own buffer of its display.
type window struct {
	parent          *window
	windowBuffer    tulib.Buffer // includes the border
	rect            tulib.Rect
	viewRect        tulib.Rect
	childrenWindows []*window
	view            wi.View
	docking         wi.DockingType
	border          wi.BorderType // Use .effectiveBorder() instead.
	fg              termbox.Attribute
	bg              termbox.Attribute
}

func (w *window) String() string {
	return fmt.Sprintf("Window(%s, %v)", w.View().Title(), w.Rect())
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

func (w *window) ViewRect() tulib.Rect {
	return w.viewRect
}

func isEqual(lhs tulib.Rect, rhs tulib.Rect) bool {
	return lhs.X == rhs.X && lhs.Y == rhs.Y && lhs.Width == rhs.Width && lhs.Height == rhs.Height
}

var singleBorder = []rune{'\u2500', '\u2502', '\u250D', '\u2510', '\u2514', '\u2518'}
var doubleBorder = []rune{'\u2550', '\u2551', '\u2554', '\u2557', '\u255a', '\u255d'}

func (w *window) SetRect(rect tulib.Rect) {
	log.Printf("%s.SetRect(%v)", w, rect)
	// SetRect() recreates the buffer and immediately draws the borders.
	if !isEqual(w.rect, rect) {
		w.rect = rect
		// Internal consistency check.
		if w.parent != nil {
			parentViewRect := w.parent.viewRect
			if parentViewRect.Width < w.rect.X+w.rect.Width {
				panic(fmt.Sprintf("Child %v Width is greater than parent %v", w, w.parent))
			}
			if parentViewRect.Height < w.rect.Y+w.rect.Height {
				panic(fmt.Sprintf("Child %v Height is greater than parent %v", w, w.parent))
			}
		}
		w.windowBuffer = tulib.NewBuffer(w.rect.Width, w.rect.Height)
		if w.effectiveBorder() != wi.BorderNone {
			w.viewRect = tulib.Rect{1, 1, w.rect.Width - 2, w.rect.Height - 2}
			w.drawBorder()
		} else {
			w.viewRect = tulib.Rect{0, 0, w.rect.Width, w.rect.Height}
		}
		if w.viewRect.Width < 0 {
			w.viewRect.Width = 0
		}
		if w.viewRect.Height < 0 {
			w.viewRect.Height = 0
		}
		w.view.SetSize(w.viewRect.Width, w.viewRect.Height)
	}
	w.resizeChildren()
}

// resizeChildren() resizes all the children Window.
func (w *window) resizeChildren() {
	// When borders are used, w.viewRect.X and .Y are 1.
	remaining := w.viewRect
	var fill wi.Window
	for _, child := range w.childrenWindows {
		switch child.Docking() {
		case wi.DockingFill:
			fill = child

		case wi.DockingFloating:
			// Floating uses its own thing.
			child.SetRect(child.Rect())

		case wi.DockingLeft:
			w, _ := child.View().NaturalSize()
			if w > remaining.Width {
				w = remaining.Width
			}
			tmp := remaining
			tmp.Width = w
			child.SetRect(tmp)
			remaining.Y += w
			remaining.Width -= w

		case wi.DockingRight:
			w, _ := child.View().NaturalSize()
			if w > remaining.Width {
				w = remaining.Width
			}
			tmp := remaining
			tmp.X += (remaining.Width - w)
			tmp.Width = w
			child.SetRect(tmp)
			remaining.Width -= w

		case wi.DockingTop:
			_, h := child.View().NaturalSize()
			if h > remaining.Height {
				h = remaining.Height
			}
			tmp := remaining
			tmp.Height = h
			child.SetRect(tmp)
			remaining.X += h
			remaining.Height -= h

		case wi.DockingBottom:
			_, h := child.View().NaturalSize()
			if h > remaining.Height {
				h = remaining.Height
			}
			log.Printf("Argh %v", w.rect)
			log.Printf("Argh %v", remaining)
			log.Printf("Argh %v", w.viewRect)
			tmp := remaining
			tmp.Y += (remaining.Height - h)
			tmp.Height = h
			log.Printf("Argh %v", tmp)
			child.SetRect(tmp)
			remaining.Height -= h

		default:
			panic("Fill me")
		}
	}
	if fill != nil {
		fill.SetRect(remaining)
	}
}

func (w *window) Buffer() *tulib.Buffer {
	w.windowBuffer.Blit(w.viewRect, 0, 0, w.view.Buffer())
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

// drawBorder draws the borders right away in the Window's buffer.
func (w *window) drawBorder() {
	s := doubleBorder
	if w.border == wi.BorderSingle {
		s = singleBorder
	}
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
}

func (w *window) cell(r rune) termbox.Cell {
	return termbox.Cell{r, w.fg, w.bg}
}

// effectiveBorder automatically collapses borders when the Window Rect is too
// small.
func (w *window) effectiveBorder() wi.BorderType {
	if w.rect.Width < 5 || w.rect.Height < 3 {
		return wi.BorderNone
	}
	return w.border
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

type nullWriter int

func (nullWriter) Write([]byte) (int, error) {
	return 0, nil
}

func Main() int {
	log.SetFlags(log.Lmicroseconds | log.Lshortfile)
	command := flag.Bool("c", false, "Runs the commands specified on startup")
	version := flag.Bool("v", false, "Prints version and exit")
	verbose := flag.Bool("verbose", false, "Logs debugging information to wi.log")
	flag.Parse()

	// Process this one early. No one wants version output to take 1s.
	if *version {
		println(version)
		os.Exit(0)
	}

	if *verbose {
		if f, err := os.OpenFile("wi.log", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666); err == nil {
			defer func() {
				_ = f.Close()
			}()
			log.SetOutput(f)
		}
	} else {
		log.SetOutput(new(nullWriter))
	}

	if err := termbox.Init(); err != nil {
		panic(err)
	}
	defer termbox.Close()
	termbox.SetInputMode(termbox.InputAlt | termbox.InputMouse)

	editor := makeEditor()
	plugins := loadPlugins(editor)
	defer func() {
		for _, p := range plugins {
			// TODO(maruel): Nicely terminate them.
			if err := p.Kill(); err != nil {
				panic(err)
			}
		}
	}()

	// Add the status bar. At that point plugins are loaded so they can override
	// add_status_bar if they want.
	editor.PostCommand("add_status_bar")

	if *command {
		for _, i := range flag.Args() {
			editor.PostCommand(i)
		}
	} else if flag.NArg() > 0 {
		for _, i := range flag.Args() {
			editor.PostCommand("open", i)
		}
	} else {
		// If nothing, opens a blank editor.
		editor.PostCommand("new")
	}

	// Run the message loop.
	return editor.eventLoop()
}

func main() {
	os.Exit(Main())
}
