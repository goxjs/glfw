// +build js

package goglfw

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"

	"github.com/gopherjs/gopherjs/js"
	"github.com/shurcooL/gogl"
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
	// THINK: Consider https://developer.mozilla.org/en-US/docs/Web/API/Window.open?

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
		attrs := gogl.DefaultAttributes()
		attrs.Alpha = (hints[AlphaBits] > 0)
		attrs.Antialias = (hints[Samples] > 0)

		gl, err := gogl.NewContext(w.canvas.Underlying(), attrs)
		if err != nil {
			return nil, err
		}

		w.Context = gl
	}

	document.AddEventListener("keydown", false, func(event dom.Event) {
		if w.keyCallback == nil {
			return
		}
		ke := event.(*dom.KeyboardEvent)

		action := Press
		if ke.Repeat {
			action = Repeat
		}

		mods := ModifierKey(0) // TODO: ke.CtrlKey && !ke.AltKey && !ke.MetaKey && !ke.ShiftKey.

		switch key := Key(ke.KeyCode); key {
		case KeyLeftShift, KeyRightShift, Key1, Key2, Key3, KeyEnter, KeyEscape, KeyF1, KeyF2, KeyLeft, KeyRight, KeyUp, KeyDown, KeyQ, KeyW, KeyE, KeyA, KeyS, KeyD, KeySpace:
			// Extend slice if needed.
			neededSize := int(key) + 1
			if neededSize > len(w.keys) {
				w.keys = append(w.keys, make([]Action, neededSize-len(w.keys))...)
			}
			w.keys[key] = action

			w.keyCallback(w, key, -1, action, mods)
		default:
			fmt.Println("Unknown KeyCode:", ke.KeyCode)
		}

		ke.PreventDefault()
	})
	document.AddEventListener("keyup", false, func(event dom.Event) {
		if w.keyCallback == nil {
			return
		}
		ke := event.(*dom.KeyboardEvent)

		mods := ModifierKey(0) // TODO: ke.CtrlKey && !ke.AltKey && !ke.MetaKey && !ke.ShiftKey.

		switch key := Key(ke.KeyCode); key {
		case KeyLeftShift, KeyRightShift, Key1, Key2, Key3, KeyEnter, KeyEscape, KeyF1, KeyF2, KeyLeft, KeyRight, KeyUp, KeyDown, KeyQ, KeyW, KeyE, KeyA, KeyS, KeyD, KeySpace:
			// Extend slice if needed.
			neededSize := int(key) + 1
			if neededSize > len(w.keys) {
				w.keys = append(w.keys, make([]Action, neededSize-len(w.keys))...)
			}
			w.keys[key] = Release

			w.keyCallback(w, key, -1, Release, mods)
		default:
			fmt.Println("Unknown KeyCode:", ke.KeyCode)
		}

		ke.PreventDefault()
	})

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

		w.cursorPos[0], w.cursorPos[1] = float64(me.ClientX), float64(me.ClientY)
		if w.cursorPosCallback != nil {
			w.cursorPosCallback(w, w.cursorPos[0], w.cursorPos[1])
		}
		if w.mouseMovementCallback != nil {
			w.mouseMovementCallback(w, w.cursorPos[0], w.cursorPos[1], float64(me.MovementX), float64(me.MovementY))
		}

		me.PreventDefault()
	})
	document.AddEventListener("wheel", false, func(event dom.Event) {
		we := event.(*dom.WheelEvent)

		var multiplier float64
		switch we.DeltaMode {
		case dom.DeltaPixel:
			multiplier = 0.1
		case dom.DeltaLine:
			multiplier = 1
		default:
			log.Println("unsupported WheelEvent.DeltaMode:", we.DeltaMode)
			multiplier = 1
		}

		if w.scrollCallback != nil {
			w.scrollCallback(w, -we.DeltaX*multiplier, -we.DeltaY*multiplier)
		}

		we.PreventDefault()
	})

	// Hacky mouse-emulation-via-touch.
	touchHandler := func(event dom.Event) {
		te := event.(*dom.TouchEvent)

		touches := te.Get("touches")
		if touches.Length() > 0 {
			t := touches.Index(0)

			if w.mouseMovementCallback != nil {
				w.mouseMovementCallback(w, t.Get("clientX").Float(), t.Get("clientY").Float(), t.Get("clientX").Float()-w.cursorPos[0], t.Get("clientY").Float()-w.cursorPos[1])
			}

			w.cursorPos[0], w.cursorPos[1] = t.Get("clientX").Float(), t.Get("clientY").Float()
			if w.cursorPosCallback != nil {
				w.cursorPosCallback(w, w.cursorPos[0], w.cursorPos[1])
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

func SwapInterval(interval int) error {
	// TODO: Implement.
	return nil
}

type Window struct {
	Context *gogl.Context

	canvas *dom.HTMLCanvasElement

	cursorMode  int
	cursorPos   [2]float64
	mouseButton [3]Action

	keys []Action

	cursorPosCallback     CursorPosCallback
	mouseMovementCallback MouseMovementCallback
	mouseButtonCallback   MouseButtonCallback
	keyCallback           KeyCallback
	scrollCallback        ScrollCallback

	touches *js.Object // Hacky mouse-emulation-via-touch.
}

type Monitor struct {
}

func PollEvents() error {
	return nil
}

func (w *Window) MakeContextCurrent() error {
	return nil
}

type CursorPosCallback func(w *Window, xpos float64, ypos float64)

func (w *Window) SetCursorPosCallback(cbfun CursorPosCallback) (previous CursorPosCallback, err error) {
	w.cursorPosCallback = cbfun

	// TODO: Handle previous.
	return nil, nil
}

type MouseMovementCallback func(w *Window, xpos float64, ypos float64, xdelta float64, ydelta float64)

func (w *Window) SetMouseMovementCallback(cbfun MouseMovementCallback) (previous MouseMovementCallback, err error) {
	w.mouseMovementCallback = cbfun

	// TODO: Handle previous.
	return nil, nil
}

type KeyCallback func(w *Window, key Key, scancode int, action Action, mods ModifierKey)

func (w *Window) SetKeyCallback(cbfun KeyCallback) (previous KeyCallback, err error) {
	w.keyCallback = cbfun

	// TODO: Handle previous.
	return nil, nil
}

type CharCallback func(w *Window, char rune)

func (w *Window) SetCharCallback(cbfun CharCallback) (previous CharCallback, err error) {
	// TODO.
	return nil, nil
}

type ScrollCallback func(w *Window, xoff float64, yoff float64)

func (w *Window) SetScrollCallback(cbfun ScrollCallback) (previous ScrollCallback, err error) {
	w.scrollCallback = cbfun

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

func (w *Window) GetSize() (width, height int) {
	// TODO: See if dpi adjustments need to be made.
	fmt.Println("Window.GetSize:", w.canvas.GetBoundingClientRect().Width, w.canvas.GetBoundingClientRect().Height)

	return w.canvas.GetBoundingClientRect().Width, w.canvas.GetBoundingClientRect().Height
}

func (w *Window) GetFramebufferSize() (width, height int) {
	return w.canvas.Width, w.canvas.Height
}

func (w *Window) ShouldClose() bool {
	return false
}

func (w *Window) SetShouldClose(value bool) {
	// TODO: Implement.
	// THINK: What should happen in the browser if we're told to "close" the window. Do we destroy/remove the canvas? Or nothing?
	//        Perhaps https://developer.mozilla.org/en-US/docs/Web/API/Window.close is relevant.
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

func (w *Window) GetCursorPos() (x, y float64) {
	return w.cursorPos[0], w.cursorPos[1]
}

func (w *Window) GetKey(key Key) Action {
	if int(key) >= len(w.keys) {
		return Release
	}
	return w.keys[key]
}

func (w *Window) GetMouseButton(button MouseButton) Action {
	if !(button >= 0 && button <= 2) {
		panic(fmt.Errorf("button is out of range: %v", button))
	}

	// Hacky mouse-emulation-via-touch.
	if w.touches != nil {
		switch button {
		case MouseButton1:
			if w.touches.Length() == 1 || w.touches.Length() == 3 {
				return Press
			}
		case MouseButton2:
			if w.touches.Length() == 2 || w.touches.Length() == 3 {
				return Press
			}
		}

		return Release
	}

	return w.mouseButton[button]
}

func (w *Window) GetInputMode(mode InputMode) int {
	switch mode {
	case CursorMode:
		return w.cursorMode
	default:
		panic(errors.New("not yet impl"))
	}
}

var ErrInvalidParameter = errors.New("invalid parameter")
var ErrInvalidValue = errors.New("invalid value")

func (w *Window) SetInputMode(mode InputMode, value int) {
	switch mode {
	case CursorMode:
		switch value {
		case CursorNormal:
			w.cursorMode = value
			document.Underlying().Call("exitPointerLock")
			w.canvas.Style().SetProperty("cursor", "initial", "")
			return
		case CursorHidden:
			w.cursorMode = value
			document.Underlying().Call("exitPointerLock")
			w.canvas.Style().SetProperty("cursor", "none", "")
			return
		case CursorDisabled:
			w.cursorMode = value
			w.canvas.Underlying().Call("requestPointerLock")
			return
		default:
			panic(ErrInvalidValue)
		}
	case StickyKeysMode:
		panic(errors.New("not impl"))
	case StickyMouseButtonsMode:
		panic(errors.New("not impl"))
	default:
		panic(ErrInvalidParameter)
	}
}

type Key int

const (
	KeyLeftShift  Key = 340
	KeyRightShift Key = 344
	Key1          Key = 49
	Key2          Key = 50
	Key3          Key = 51
	KeyEnter      Key = 13
	KeyEscape     Key = 27
	KeyF1         Key = 112
	KeyF2         Key = 113
	KeyLeft       Key = 37
	KeyRight      Key = 39
	KeyUp         Key = 38
	KeyDown       Key = 40
	KeyQ          Key = 81
	KeyW          Key = 87
	KeyE          Key = 69
	KeyA          Key = 65
	KeyS          Key = 83
	KeyD          Key = 68
	KeySpace      Key = 32
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
	CursorMode InputMode = iota
	StickyKeysMode
	StickyMouseButtonsMode
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
	b, err := xhr.Send("GET", name, nil)
	if err != nil {
		return nil, err
	}

	return nopCloser{bytes.NewReader(b)}, nil
}

type nopCloser struct {
	io.ReadSeeker
}

func (nopCloser) Close() error { return nil }
