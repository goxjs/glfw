// +build js

package goglfw

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	"github.com/gopherjs/gopherjs/js"
	"github.com/shurcooL/webgl"
	"golang.org/x/tools/godoc/vfs"
	"honnef.co/go/js/dom"
	"honnef.co/go/js/xhr"
)

var document = dom.GetWindow().Document().(dom.HTMLDocument)

func Init() error {
	document.Body().Style().SetProperty("margin", "0px", "")

	return nil
}

func Terminate() error {
	return nil
}

func CreateWindow(_, _ int, title string, monitor *Monitor, share *Window) (*Window, error) {
	// HACK: Go fullscreen?
	width := dom.GetWindow().InnerWidth()
	height := dom.GetWindow().InnerHeight()

	canvas := document.CreateElement("canvas").(*dom.HTMLCanvasElement)

	devicePixelRatio := js.Global.Get("devicePixelRatio").Float()
	canvas.Width = int(float64(width)*devicePixelRatio + 0.5)   // Nearest non-negative int.
	canvas.Height = int(float64(height)*devicePixelRatio + 0.5) // Nearest non-negative int.
	canvas.Style().SetProperty("width", fmt.Sprintf("%vpx", width), "")
	canvas.Style().SetProperty("height", fmt.Sprintf("%vpx", height), "")

	document.Body().AppendChild(canvas)

	document.SetTitle(title)

	// DEBUG: Add framebuffer information div.
	if false {
		//canvas.Height -= 30
		text := document.CreateElement("div")
		textContent := fmt.Sprintf("%v %v (%v) @%v", dom.GetWindow().InnerWidth(), canvas.Width, float64(width)*devicePixelRatio, devicePixelRatio)
		text.SetTextContent(textContent)
		document.Body().AppendChild(text)
	}

	w := &Window{
		canvas: canvas,
	}

	// Create GL context.
	{
		attrs := webgl.DefaultAttributes()
		attrs.Alpha = (hints[AlphaBits] > 0)
		attrs.Antialias = (hints[Samples] > 0)

		gl, err := webgl.NewContext(w.canvas.Underlying(), attrs)
		if err != nil {
			return nil, err
		}

		w.Context = gl
	}

	document.AddEventListener("mousedown", false, func(event dom.Event) {
		me := event.(*dom.MouseEvent)
		if !(me.Button >= 0 && me.Button <= 2) {
			return
		}

		w.mouseButton[me.Button] = Press
		if w.mouseButtonCallback != nil {
			w.mouseButtonCallback(w, MouseButton(me.Button), Press, 0)
		}

		me.PreventDefault()
	})
	document.AddEventListener("mouseup", false, func(event dom.Event) {
		me := event.(*dom.MouseEvent)
		if !(me.Button >= 0 && me.Button <= 2) {
			return
		}

		w.mouseButton[me.Button] = Release
		if w.mouseButtonCallback != nil {
			w.mouseButtonCallback(w, MouseButton(me.Button), Release, 0)
		}

		me.PreventDefault()
	})
	document.AddEventListener("contextmenu", false, func(event dom.Event) {
		event.PreventDefault()
	})

	document.AddEventListener("mousemove", false, func(event dom.Event) {
		me := event.(*dom.MouseEvent)

		w.cursorPosition[0], w.cursorPosition[1] = float64(me.ClientX), float64(me.ClientY)
		if w.cursorPositionCallback != nil {
			w.cursorPositionCallback(w, w.cursorPosition[0], w.cursorPosition[1])
		}
		if w.mouseMovementCallback != nil {
			w.mouseMovementCallback(w, float64(me.MovementX), float64(me.MovementY))
		}

		me.PreventDefault()
	})

	// Hacky mouse-emulation-via-touch.
	touchHandler := func(event dom.Event) {
		te := event.(*dom.TouchEvent)

		touches := te.Get("touches")
		if touches.Length() > 0 {
			t := touches.Index(0)

			if w.mouseMovementCallback != nil {
				w.mouseMovementCallback(w, t.Get("clientX").Float()-w.cursorPosition[0], t.Get("clientY").Float()-w.cursorPosition[1])
			}

			w.cursorPosition[0], w.cursorPosition[1] = t.Get("clientX").Float(), t.Get("clientY").Float()
			if w.cursorPositionCallback != nil {
				w.cursorPositionCallback(w, w.cursorPosition[0], w.cursorPosition[1])
			}
		}
		w.touches = touches

		te.PreventDefault()
	}
	document.AddEventListener("touchstart", false, touchHandler)
	document.AddEventListener("touchmove", false, touchHandler)
	document.AddEventListener("touchend", false, touchHandler)

	// Request first animation frame.
	js.Global.Call("requestAnimationFrame", animationFrame)

	return w, nil
}

type Window struct {
	Context *webgl.Context

	canvas *dom.HTMLCanvasElement

	cursorPosition [2]float64
	mouseButton    [3]Action

	cursorPositionCallback CursorPositionCallback
	mouseMovementCallback  MouseMovementCallback
	mouseButtonCallback    MouseButtonCallback

	touches js.Object // Hacky mouse-emulation-via-touch.
}

type Monitor struct {
}

func PollEvents() error {
	return nil
}

func (w *Window) MakeContextCurrent() error {
	return nil
}

type CursorPositionCallback func(w *Window, xpos float64, ypos float64)

func (w *Window) SetCursorPositionCallback(cbfun CursorPositionCallback) (previous CursorPositionCallback, err error) {
	w.cursorPositionCallback = cbfun

	// TODO: Handle previous.
	return nil, nil
}

type MouseMovementCallback func(w *Window, xdelta float64, ydelta float64)

func (w *Window) SetMouseMovementCallback(cbfun MouseMovementCallback) (previous MouseMovementCallback, err error) {
	w.mouseMovementCallback = cbfun

	// TODO: Handle previous.
	return nil, nil
}

type MouseButtonCallback func(w *Window, button MouseButton, action Action, mods ModifierKey)

func (w *Window) SetMouseButtonCallback(cbfun MouseButtonCallback) (previous MouseButtonCallback, err error) {
	w.mouseButtonCallback = cbfun

	// TODO: Handle previous.
	return nil, nil
}

type FramebufferSizeCallback func(w *Window, width int, height int)

func (w *Window) SetFramebufferSizeCallback(cbfun FramebufferSizeCallback) (previous FramebufferSizeCallback, err error) {
	dom.GetWindow().AddEventListener("resize", false, func(event dom.Event) {
		// HACK: Go fullscreen?
		width := dom.GetWindow().InnerWidth()
		height := dom.GetWindow().InnerHeight()

		devicePixelRatio := js.Global.Get("devicePixelRatio").Float()
		w.canvas.Width = int(float64(width)*devicePixelRatio + 0.5)   // Nearest non-negative int.
		w.canvas.Height = int(float64(height)*devicePixelRatio + 0.5) // Nearest non-negative int.
		w.canvas.Style().SetProperty("width", fmt.Sprintf("%vpx", width), "")
		w.canvas.Style().SetProperty("height", fmt.Sprintf("%vpx", height), "")

		cbfun(w, w.canvas.Width, w.canvas.Height)
	})

	// TODO: Handle previous.
	return nil, nil
}

func (w *Window) GetSize() (width, height int, err error) {
	// TODO: See if dpi adjustments need to be made.
	fmt.Println("Window.GetSize:", w.canvas.GetBoundingClientRect().Width, w.canvas.GetBoundingClientRect().Height)

	return w.canvas.GetBoundingClientRect().Width, w.canvas.GetBoundingClientRect().Height, nil
}

func (w *Window) GetFramebufferSize() (width, height int, err error) {
	return w.canvas.Width, w.canvas.Height, nil
}

func (w *Window) ShouldClose() (bool, error) {
	return false, nil
}

func (w *Window) SwapBuffers() error {
	<-animationFrameChan
	js.Global.Call("requestAnimationFrame", animationFrame)

	return nil
}

var animationFrameChan = make(chan struct{})

func animationFrame() {
	go func() {
		animationFrameChan <- struct{}{}
	}()
}

func (w *Window) GetCursorPosition() (x, y float64, err error) {
	return w.cursorPosition[0], w.cursorPosition[1], nil
}

func (w *Window) GetKey(key Key) (Action, error) {
	// TODO: Implement.
	return Release, nil
}

func (w *Window) GetMouseButton(button MouseButton) (Action, error) {
	if !(button >= 0 && button <= 2) {
		return 0, fmt.Errorf("button is out of range: %v", button)
	}

	// Hacky mouse-emulation-via-touch.
	if w.touches != nil {
		switch button {
		case MouseButton1:
			if w.touches.Length() == 1 || w.touches.Length() == 3 {
				return Press, nil
			}
		case MouseButton2:
			if w.touches.Length() == 2 || w.touches.Length() == 3 {
				return Press, nil
			}
		}

		return Release, nil
	}

	return w.mouseButton[button], nil
}

func (w *Window) GetInputMode(mode InputMode) (int, error) {
	return 0, errors.New("not yet impl")
}

var ErrInvalidParameter = errors.New("invalid parameter")
var ErrInvalidValue = errors.New("invalid value")

func (w *Window) SetInputMode(mode InputMode, value int) error {
	switch mode {
	case Cursor:
		switch value {
		case CursorNormal:
			document.Underlying().Call("exitPointerLock")
			w.canvas.Style().SetProperty("cursor", "initial", "")
			return nil
		case CursorHidden:
			document.Underlying().Call("exitPointerLock")
			w.canvas.Style().SetProperty("cursor", "none", "")
			return nil
		case CursorDisabled:
			w.canvas.Underlying().Call("requestPointerLock")
			return nil
		default:
			return ErrInvalidValue
		}
	case StickyKeys:
		return errors.New("not impl")
	case StickyMouseButtons:
		return errors.New("not impl")
	default:
		return ErrInvalidParameter
	}
}

type Key int

const (
	KeyLeftShift  Key = 340
	KeyRightShift Key = 344
)

type MouseButton int

const (
	MouseButton1 MouseButton = 0
	MouseButton2 MouseButton = 2 // Web MouseEvent has middle and right mouse buttons in reverse order.
	MouseButton3 MouseButton = 1 // Web MouseEvent has middle and right mouse buttons in reverse order.
)

type Action int

const (
	Release Action = 0
	Press   Action = 1
	Repeat  Action = 2
)

type InputMode int

const (
	Cursor InputMode = iota
	StickyKeys
	StickyMouseButtons
)

const (
	CursorNormal = iota
	CursorHidden
	CursorDisabled
)

type ModifierKey int

const (
	ModShift ModifierKey = iota
	ModControl
	ModAlt
	ModSuper
)

// Open opens a named asset.
func Open(name string) (vfs.ReadSeekCloser, error) {
	req := xhr.NewRequest("GET", name)
	req.ResponseType = xhr.ArrayBuffer

	err := req.Send(nil)
	if err != nil {
		return nil, err
	}

	b := js.Global.Get("Uint8Array").New(req.Response).Interface().([]byte)

	return nopCloser{bytes.NewReader(b)}, nil
}

type nopCloser struct {
	io.ReadSeeker
}

func (nopCloser) Close() error { return nil }
