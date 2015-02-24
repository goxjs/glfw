// +build !js

package goglfw

import (
	"os"
	"runtime"

	"github.com/go-gl/glfw/v3.1/glfw"
	"github.com/shurcooL/gogl"
	"golang.org/x/tools/godoc/vfs"
)

func init() {
	runtime.LockOSThread()
}

func Init() error {
	return glfw.Init()
}

func Terminate() {
	glfw.Terminate()
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

	window.Context = gogl.NewContext()

	return window, err
}

func SwapInterval(interval int) {
	glfw.SwapInterval(interval)
}

type Window struct {
	*glfw.Window

	Context *gogl.Context
}

type Monitor struct {
	*glfw.Monitor
}

func PollEvents() {
	glfw.PollEvents()
}

type CursorPosCallback func(w *Window, xpos float64, ypos float64)

func (w *Window) SetCursorPosCallback(cbfun CursorPosCallback) (previous CursorPosCallback) {
	wrappedCbfun := func(_ *glfw.Window, xpos float64, ypos float64) {
		cbfun(w, xpos, ypos)
	}

	p := w.Window.SetCursorPosCallback(wrappedCbfun)
	_ = p

	// TODO: Handle previous.
	return nil
}

type MouseMovementCallback func(w *Window, xpos float64, ypos float64, xdelta float64, ydelta float64)

var lastMousePos [2]float64 // HACK.

// TODO: For now, this overrides SetCursorPosCallback; should support both.
func (w *Window) SetMouseMovementCallback(cbfun MouseMovementCallback) (previous MouseMovementCallback) {
	lastMousePos[0], lastMousePos[1] = w.Window.GetCursorPos()
	wrappedCbfun := func(_ *glfw.Window, xpos float64, ypos float64) {
		xdelta, ydelta := xpos-lastMousePos[0], ypos-lastMousePos[1]
		lastMousePos[0], lastMousePos[1] = xpos, ypos
		cbfun(w, xpos, ypos, xdelta, ydelta)
	}

	p := w.Window.SetCursorPosCallback(wrappedCbfun)
	_ = p

	// TODO: Handle previous.
	return nil
}

type KeyCallback func(w *Window, key Key, scancode int, action Action, mods ModifierKey)

func (w *Window) SetKeyCallback(cbfun KeyCallback) (previous KeyCallback) {
	wrappedCbfun := func(_ *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
		cbfun(w, Key(key), scancode, Action(action), ModifierKey(mods))
	}

	p := w.Window.SetKeyCallback(wrappedCbfun)
	_ = p

	// TODO: Handle previous.
	return nil
}

type CharCallback func(w *Window, char rune)

func (w *Window) SetCharCallback(cbfun CharCallback) (previous CharCallback) {
	wrappedCbfun := func(_ *glfw.Window, char rune) {
		cbfun(w, char)
	}

	p := w.Window.SetCharCallback(wrappedCbfun)
	_ = p

	// TODO: Handle previous.
	return nil
}

type ScrollCallback func(w *Window, xoff float64, yoff float64)

func (w *Window) SetScrollCallback(cbfun ScrollCallback) (previous ScrollCallback) {
	wrappedCbfun := func(_ *glfw.Window, xoff float64, yoff float64) {
		cbfun(w, xoff, yoff)
	}

	p := w.Window.SetScrollCallback(wrappedCbfun)
	_ = p

	// TODO: Handle previous.
	return nil
}

type MouseButtonCallback func(w *Window, button MouseButton, action Action, mods ModifierKey)

func (w *Window) SetMouseButtonCallback(cbfun MouseButtonCallback) (previous MouseButtonCallback) {
	wrappedCbfun := func(_ *glfw.Window, button glfw.MouseButton, action glfw.Action, mods glfw.ModifierKey) {
		cbfun(w, MouseButton(button), Action(action), ModifierKey(mods))
	}

	p := w.Window.SetMouseButtonCallback(wrappedCbfun)
	_ = p

	// TODO: Handle previous.
	return nil
}

type FramebufferSizeCallback func(w *Window, width int, height int)

func (w *Window) SetFramebufferSizeCallback(cbfun FramebufferSizeCallback) (previous FramebufferSizeCallback) {
	wrappedCbfun := func(_ *glfw.Window, width int, height int) {
		cbfun(w, width, height)
	}

	p := w.Window.SetFramebufferSizeCallback(wrappedCbfun)
	_ = p

	// TODO: Handle previous.
	return nil
}

func (w *Window) GetKey(key Key) Action {
	a := w.Window.GetKey(glfw.Key(key))
	return Action(a)
}

func (w *Window) GetMouseButton(button MouseButton) Action {
	a := w.Window.GetMouseButton(glfw.MouseButton(button))
	return Action(a)
}

func (w *Window) GetInputMode(mode InputMode) int {
	return w.Window.GetInputMode(glfw.InputMode(mode))
}

func (w *Window) SetInputMode(mode InputMode, value int) {
	w.Window.SetInputMode(glfw.InputMode(mode), value)
}

type Key glfw.Key

const (
	KeyLeftShift  = Key(glfw.KeyLeftShift)
	KeyRightShift = Key(glfw.KeyRightShift)
	Key1          = Key(glfw.Key1)
	Key2          = Key(glfw.Key2)
	Key3          = Key(glfw.Key3)
	KeyEnter      = Key(glfw.KeyEnter)
	KeyEscape     = Key(glfw.KeyEscape)
	KeyF1         = Key(glfw.KeyF1)
	KeyF2         = Key(glfw.KeyF2)
	KeyLeft       = Key(glfw.KeyLeft)
	KeyRight      = Key(glfw.KeyRight)
	KeyUp         = Key(glfw.KeyUp)
	KeyDown       = Key(glfw.KeyDown)
	KeyQ          = Key(glfw.KeyQ)
	KeyW          = Key(glfw.KeyW)
	KeyE          = Key(glfw.KeyE)
	KeyA          = Key(glfw.KeyA)
	KeyS          = Key(glfw.KeyS)
	KeyD          = Key(glfw.KeyD)
	KeySpace      = Key(glfw.KeySpace)
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
	CursorMode             = InputMode(glfw.CursorMode)
	StickyKeysMode         = InputMode(glfw.StickyKeysMode)
	StickyMouseButtonsMode = InputMode(glfw.StickyMouseButtonsMode)
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
