// +build !js

package goglfw

import (
	"os"
	"runtime"

	glfw "github.com/shurcooL/glfw3"
	"github.com/shurcooL/webgl"
	"golang.org/x/tools/godoc/vfs"
)

func Init() error {
	runtime.LockOSThread()

	return glfw.Init()
}

func Terminate() error {
	return glfw.Terminate()
}

func CreateWindow(width, height int, title string, monitor *Monitor, share *Window) (*Window, error) {
	var m *glfw.Monitor
	if monitor != nil {
		m = monitor.Monitor
	}
	var s *glfw.Window
	if share != nil {
		s = share.Window
	}

	w, err := glfw.CreateWindow(width, height, title, m, s)
	if err != nil {
		return nil, err
	}

	window := &Window{Window: w}

	window.Context = webgl.NewContext()

	return window, err
}

func SwapInterval(interval int) error {
	return glfw.SwapInterval(interval)
}

type Window struct {
	*glfw.Window

	Context *webgl.Context
}

type Monitor struct {
	*glfw.Monitor
}

func PollEvents() error {
	return glfw.PollEvents()
}

type CursorPositionCallback func(w *Window, xpos float64, ypos float64)

func (w *Window) SetCursorPositionCallback(cbfun CursorPositionCallback) (previous CursorPositionCallback, err error) {
	wrappedCbfun := func(_ *glfw.Window, xpos float64, ypos float64) {
		cbfun(w, xpos, ypos)
	}

	p, err := w.Window.SetCursorPositionCallback(wrappedCbfun)
	_ = p

	// TODO: Handle previous.
	return nil, err
}

type MouseMovementCallback func(w *Window, xdelta float64, ydelta float64)

var lastMousePos [2]float64 // HACK.

// TODO: For now, this overrides SetCursorPositionCallback; should support both.
func (w *Window) SetMouseMovementCallback(cbfun MouseMovementCallback) (previous MouseMovementCallback, err error) {
	lastMousePos[0], lastMousePos[1], _ = w.Window.GetCursorPosition()
	wrappedCbfun := func(_ *glfw.Window, xpos float64, ypos float64) {
		xdelta, ydelta := xpos-lastMousePos[0], ypos-lastMousePos[1]
		lastMousePos[0], lastMousePos[1] = xpos, ypos
		cbfun(w, xdelta, ydelta)
	}

	p, err := w.Window.SetCursorPositionCallback(wrappedCbfun)
	_ = p

	// TODO: Handle previous.
	return nil, err
}

type KeyCallback func(w *Window, key Key, scancode int, action Action, mods ModifierKey)

func (w *Window) SetKeyCallback(cbfun KeyCallback) (previous KeyCallback, err error) {
	wrappedCbfun := func(_ *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
		cbfun(w, Key(key), scancode, Action(action), ModifierKey(mods))
	}

	p, err := w.Window.SetKeyCallback(wrappedCbfun)
	_ = p

	// TODO: Handle previous.
	return nil, err
}

type CharCallback func(w *Window, char rune)

func (w *Window) SetCharCallback(cbfun CharCallback) (previous CharCallback, err error) {
	wrappedCbfun := func(_ *glfw.Window, char rune) {
		cbfun(w, char)
	}

	p, err := w.Window.SetCharCallback(wrappedCbfun)
	_ = p

	// TODO: Handle previous.
	return nil, err
}

type ScrollCallback func(w *Window, xoff float64, yoff float64)

func (w *Window) SetScrollCallback(cbfun ScrollCallback) (previous ScrollCallback, err error) {
	wrappedCbfun := func(_ *glfw.Window, xoff float64, yoff float64) {
		cbfun(w, xoff, yoff)
	}

	p, err := w.Window.SetScrollCallback(wrappedCbfun)
	_ = p

	// TODO: Handle previous.
	return nil, err
}

type MouseButtonCallback func(w *Window, button MouseButton, action Action, mods ModifierKey)

func (w *Window) SetMouseButtonCallback(cbfun MouseButtonCallback) (previous MouseButtonCallback, err error) {
	wrappedCbfun := func(_ *glfw.Window, button glfw.MouseButton, action glfw.Action, mods glfw.ModifierKey) {
		cbfun(w, MouseButton(button), Action(action), ModifierKey(mods))
	}

	p, err := w.Window.SetMouseButtonCallback(wrappedCbfun)
	_ = p

	// TODO: Handle previous.
	return nil, err
}

type FramebufferSizeCallback func(w *Window, width int, height int)

func (w *Window) SetFramebufferSizeCallback(cbfun FramebufferSizeCallback) (previous FramebufferSizeCallback, err error) {
	wrappedCbfun := func(_ *glfw.Window, width int, height int) {
		cbfun(w, width, height)
	}

	p, err := w.Window.SetFramebufferSizeCallback(wrappedCbfun)
	_ = p

	// TODO: Handle previous.
	return nil, err
}

func (w *Window) GetKey(key Key) (Action, error) {
	a, err := w.Window.GetKey(glfw.Key(key))
	return Action(a), err
}

func (w *Window) GetMouseButton(button MouseButton) (Action, error) {
	a, err := w.Window.GetMouseButton(glfw.MouseButton(button))
	return Action(a), err
}

func (w *Window) GetInputMode(mode InputMode) (int, error) {
	return w.Window.GetInputMode(glfw.InputMode(mode))
}

func (w *Window) SetInputMode(mode InputMode, value int) error {
	return w.Window.SetInputMode(glfw.InputMode(mode), value)
}

type Key glfw.Key

const (
	KeyLeftShift  = Key(glfw.KeyLeftShift)
	KeyRightShift = Key(glfw.KeyRightShift)
	Key1          = Key(glfw.Key1)
	Key2          = Key(glfw.Key2)
	Key3          = Key(glfw.Key3)
)

type MouseButton glfw.MouseButton

const (
	MouseButton1 = MouseButton(glfw.MouseButton1)
	MouseButton2 = MouseButton(glfw.MouseButton2)
)

type Action glfw.Action

const (
	Release = Action(glfw.Release)
	Press   = Action(glfw.Press)
	Repeat  = Action(glfw.Repeat)
)

type InputMode int

const (
	Cursor             = InputMode(glfw.Cursor)
	StickyKeys         = InputMode(glfw.StickyKeys)
	StickyMouseButtons = InputMode(glfw.StickyMouseButtons)
)

const (
	CursorNormal   = int(glfw.CursorNormal)
	CursorHidden   = int(glfw.CursorHidden)
	CursorDisabled = int(glfw.CursorDisabled)
)

type ModifierKey int

const (
	ModShift   = ModifierKey(glfw.ModShift)
	ModControl = ModifierKey(glfw.ModControl)
	ModAlt     = ModifierKey(glfw.ModAlt)
	ModSuper   = ModifierKey(glfw.ModSuper)
)

// Open opens a named asset.
//
// For now, assets are read directly from the current working directory.
func Open(name string) (vfs.ReadSeekCloser, error) {
	return os.Open(name)
}
