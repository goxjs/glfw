// +build !js

package goglfw

import (
	"os"
	"runtime"

	"github.com/ajhager/webgl"
	glfw "github.com/shurcooL/glfw3"
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

type Key glfw.Key

const (
	KeyLeftShift  = Key(glfw.KeyLeftShift)
	KeyRightShift = Key(glfw.KeyRightShift)
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

// Open opens a named asset.
//
// For now, assets are read directly from the current working directory.
func Open(name string) (vfs.ReadSeekCloser, error) {
	return os.Open(name)
}
