// +build js

package glfw

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"runtime"

	"github.com/gopherjs/gopherjs/js"
	"golang.org/x/tools/godoc/vfs"
	"honnef.co/go/js/dom"
)

var document = dom.GetWindow().Document().(dom.HTMLDocument)

var contextWatcher ContextWatcher

func Init(cw ContextWatcher) error {
	contextWatcher = cw
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

	if js.Global.Get("document").Get("body") == nil {
		body := js.Global.Get("document").Call("createElement", "body")
		js.Global.Get("document").Set("body", body)
		log.Println("Creating body, since it doesn't exist.")
	}
	document.Body().Style().SetProperty("margin", "0", "")
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

	// Use glfw hints.
	attrs := defaultAttributes()
	attrs.Alpha = (hints[AlphaBits] > 0)
	attrs.Antialias = (hints[Samples] > 0)

	// Create GL context.
	context, err := newContext(canvas.Underlying(), attrs)
	if err != nil {
		return nil, err
	}

	w := &Window{
		canvas:  canvas,
		context: context,
	}

	if w.canvas.Underlying().Get("requestPointerLock") == js.Undefined ||
		document.Underlying().Get("exitPointerLock") == js.Undefined {

		w.missing.pointerLock = true
	}
	if w.canvas.Underlying().Get("webkitRequestFullscreen") == js.Undefined ||
		document.Underlying().Get("webkitExitFullscreen") == js.Undefined {

		w.missing.fullscreen = true
	}

	if monitor != nil {
		if w.missing.fullscreen {
			log.Println("warning: Fullscreen API unsupported")
		} else {
			w.requestFullscreen = true
		}
	}

	dom.GetWindow().AddEventListener("resize", false, func(event dom.Event) {
		// HACK: Go fullscreen?
		width := dom.GetWindow().InnerWidth()
		height := dom.GetWindow().InnerHeight()

		devicePixelRatio := js.Global.Get("devicePixelRatio").Float()
		w.canvas.Width = int(float64(width)*devicePixelRatio + 0.5)   // Nearest non-negative int.
		w.canvas.Height = int(float64(height)*devicePixelRatio + 0.5) // Nearest non-negative int.
		w.canvas.Style().SetProperty("width", fmt.Sprintf("%vpx", width), "")
		w.canvas.Style().SetProperty("height", fmt.Sprintf("%vpx", height), "")

		if w.framebufferSizeCallback != nil {
			// TODO: Callbacks may be blocking so they need to happen asyncronously. However,
			//       GLFW API promises the callbacks will occur from one thread, so may want to do that.
			go w.framebufferSizeCallback(w, w.canvas.Width, w.canvas.Height)
		}
		if w.sizeCallback != nil {
			go w.sizeCallback(w, w.canvas.GetBoundingClientRect().Width, w.canvas.GetBoundingClientRect().Height)
		}
	})

	document.AddEventListener("keydown", false, func(event dom.Event) {
		w.goFullscreenIfRequested()

		ke := event.(*dom.KeyboardEvent)

		action := Press
		if ke.Repeat {
			action = Repeat
		}

		key := Key(ke.KeyCode)

		switch {
		case key == 16 && ke.Location == dom.KeyLocationLeft:
			key = KeyLeftShift
		case key == 16 && ke.Location == dom.KeyLocationRight:
			key = KeyRightShift
		}

		switch key {
		case KeyLeftShift, KeyRightShift, Key1, Key2, Key3, KeyEnter, KeyTab, KeyEscape, KeyF1, KeyF2, KeyLeft, KeyRight, KeyUp, KeyDown, KeyQ, KeyW, KeyE, KeyA, KeyS, KeyD, KeySpace:
			// Extend slice if needed.
			neededSize := int(key) + 1
			if neededSize > len(w.keys) {
				w.keys = append(w.keys, make([]Action, neededSize-len(w.keys))...)
			}
			w.keys[key] = action

			if w.keyCallback != nil {
				mods := ModifierKey(0) // TODO: ke.CtrlKey && !ke.AltKey && !ke.MetaKey && !ke.ShiftKey.

				go w.keyCallback(w, key, -1, action, mods)
			}
		default:
			fmt.Println("Unknown KeyCode:", ke.KeyCode)
		}

		ke.PreventDefault()
	})
	document.AddEventListener("keyup", false, func(event dom.Event) {
		w.goFullscreenIfRequested()

		ke := event.(*dom.KeyboardEvent)

		key := Key(ke.KeyCode)

		switch {
		case key == 16 && ke.Location == dom.KeyLocationLeft:
			key = KeyLeftShift
		case key == 16 && ke.Location == dom.KeyLocationRight:
			key = KeyRightShift
		}

		switch key {
		case KeyLeftShift, KeyRightShift, Key1, Key2, Key3, KeyEnter, KeyTab, KeyEscape, KeyF1, KeyF2, KeyLeft, KeyRight, KeyUp, KeyDown, KeyQ, KeyW, KeyE, KeyA, KeyS, KeyD, KeySpace:
			// Extend slice if needed.
			neededSize := int(key) + 1
			if neededSize > len(w.keys) {
				w.keys = append(w.keys, make([]Action, neededSize-len(w.keys))...)
			}
			w.keys[key] = Release

			if w.keyCallback != nil {
				mods := ModifierKey(0) // TODO: ke.CtrlKey && !ke.AltKey && !ke.MetaKey && !ke.ShiftKey.

				go w.keyCallback(w, key, -1, Release, mods)
			}
		default:
			fmt.Println("Unknown KeyCode:", ke.KeyCode)
		}

		ke.PreventDefault()
	})

	document.AddEventListener("mousedown", false, func(event dom.Event) {
		w.goFullscreenIfRequested()

		me := event.(*dom.MouseEvent)
		if !(me.Button >= 0 && me.Button <= 2) {
			return
		}

		w.mouseButton[me.Button] = Press
		if w.mouseButtonCallback != nil {
			go w.mouseButtonCallback(w, MouseButton(me.Button), Press, 0)
		}

		me.PreventDefault()
	})
	document.AddEventListener("mouseup", false, func(event dom.Event) {
		w.goFullscreenIfRequested()

		me := event.(*dom.MouseEvent)
		if !(me.Button >= 0 && me.Button <= 2) {
			return
		}

		w.mouseButton[me.Button] = Release
		if w.mouseButtonCallback != nil {
			go w.mouseButtonCallback(w, MouseButton(me.Button), Release, 0)
		}

		me.PreventDefault()
	})
	document.AddEventListener("contextmenu", false, func(event dom.Event) {
		event.PreventDefault()
	})

	document.AddEventListener("mousemove", false, func(event dom.Event) {
		me := event.(*dom.MouseEvent)

		var movementX, movementY float64
		if !w.missing.pointerLock {
			movementX = float64(me.MovementX)
			movementY = float64(me.MovementY)
		} else {
			movementX = float64(me.ClientX) - w.cursorPos[0]
			movementY = float64(me.ClientY) - w.cursorPos[1]
		}

		w.cursorPos[0], w.cursorPos[1] = float64(me.ClientX), float64(me.ClientY)
		if w.cursorPosCallback != nil {
			go w.cursorPosCallback(w, w.cursorPos[0], w.cursorPos[1])
		}
		if w.mouseMovementCallback != nil {
			go w.mouseMovementCallback(w, w.cursorPos[0], w.cursorPos[1], movementX, movementY)
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
			go w.scrollCallback(w, -we.DeltaX*multiplier, -we.DeltaY*multiplier)
		}

		we.PreventDefault()
	})

	// Hacky mouse-emulation-via-touch.
	touchHandler := func(event dom.Event) {
		w.goFullscreenIfRequested()

		te := event.(*dom.TouchEvent)

		touches := te.Get("touches")
		if touches.Length() > 0 {
			t := touches.Index(0)

			if w.mouseMovementCallback != nil {
				go w.mouseMovementCallback(w, t.Get("clientX").Float(), t.Get("clientY").Float(), t.Get("clientX").Float()-w.cursorPos[0], t.Get("clientY").Float()-w.cursorPos[1])
			}

			w.cursorPos[0], w.cursorPos[1] = t.Get("clientX").Float(), t.Get("clientY").Float()
			if w.cursorPosCallback != nil {
				go w.cursorPosCallback(w, w.cursorPos[0], w.cursorPos[1])
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
	canvas            *dom.HTMLCanvasElement
	context           *js.Object
	requestFullscreen bool // requestFullscreen is set to true when fullscreen should be entered as soon as possible (in a user input handler).
	fullscreen        bool // fullscreen is true if we're currently in fullscreen mode.

	// Unavailable browser APIs.
	missing struct {
		pointerLock bool // Pointer Lock API.
		fullscreen  bool // Fullscreen API.
	}

	cursorMode  int
	cursorPos   [2]float64
	mouseButton [3]Action

	keys []Action

	cursorPosCallback       CursorPosCallback
	mouseMovementCallback   MouseMovementCallback
	mouseButtonCallback     MouseButtonCallback
	keyCallback             KeyCallback
	scrollCallback          ScrollCallback
	framebufferSizeCallback FramebufferSizeCallback
	sizeCallback            SizeCallback

	touches *js.Object // Hacky mouse-emulation-via-touch.
}

func (w *Window) SetSize(width, height int) {
	fmt.Println("not yet implemented: SetSize", width, height)
}

// goFullscreenIfRequested performs webkitRequestFullscreen if it was scheduled. It is called only from
// user events, because that API will fail if called at any other time.
func (w *Window) goFullscreenIfRequested() {
	if !w.requestFullscreen {
		return
	}
	w.requestFullscreen = false
	w.canvas.Underlying().Call("webkitRequestFullscreen")
	w.fullscreen = true
}

type Monitor struct{}

func (m *Monitor) GetVideoMode() *VidMode {
	return &VidMode{
		// HACK: Hardcoded sample values.
		// TODO: Try to get real values from browser via some API, if possible.
		Width:       1680,
		Height:      1050,
		RedBits:     8,
		GreenBits:   8,
		BlueBits:    8,
		RefreshRate: 60,
	}
}

func GetPrimaryMonitor() *Monitor {
	// TODO: Implement real functionality.
	return &Monitor{}
}

func PollEvents() error {
	return nil
}

func (w *Window) MakeContextCurrent() {
	contextWatcher.OnBecomeCurrent(w.context)
}

func DetachCurrentContext() {
	contextWatcher.OnDetach()
}

func GetCurrentContext() *Window {
	panic("not yet implemented")
}

type CursorPosCallback func(w *Window, xpos float64, ypos float64)

func (w *Window) SetCursorPosCallback(cbfun CursorPosCallback) (previous CursorPosCallback) {
	w.cursorPosCallback = cbfun

	// TODO: Handle previous.
	return nil
}

type MouseMovementCallback func(w *Window, xpos float64, ypos float64, xdelta float64, ydelta float64)

func (w *Window) SetMouseMovementCallback(cbfun MouseMovementCallback) (previous MouseMovementCallback) {
	w.mouseMovementCallback = cbfun

	// TODO: Handle previous.
	return nil
}

type KeyCallback func(w *Window, key Key, scancode int, action Action, mods ModifierKey)

func (w *Window) SetKeyCallback(cbfun KeyCallback) (previous KeyCallback) {
	w.keyCallback = cbfun

	// TODO: Handle previous.
	return nil
}

type CharCallback func(w *Window, char rune)

func (w *Window) SetCharCallback(cbfun CharCallback) (previous CharCallback) {
	// TODO.
	return nil
}

type ScrollCallback func(w *Window, xoff float64, yoff float64)

func (w *Window) SetScrollCallback(cbfun ScrollCallback) (previous ScrollCallback) {
	w.scrollCallback = cbfun

	// TODO: Handle previous.
	return nil
}

type MouseButtonCallback func(w *Window, button MouseButton, action Action, mods ModifierKey)

func (w *Window) SetMouseButtonCallback(cbfun MouseButtonCallback) (previous MouseButtonCallback) {
	w.mouseButtonCallback = cbfun

	// TODO: Handle previous.
	return nil
}

type FramebufferSizeCallback func(w *Window, width int, height int)

func (w *Window) SetFramebufferSizeCallback(cbfun FramebufferSizeCallback) (previous FramebufferSizeCallback) {
	w.framebufferSizeCallback = cbfun

	// TODO: Handle previous.
	return nil
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
		if w.missing.pointerLock {
			log.Println("warning: Pointer Lock API unsupported")
			return
		}
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
	KeySpace        Key = 32
	KeyApostrophe   Key = -1
	KeyComma        Key = -1
	KeyMinus        Key = -1
	KeyPeriod       Key = -1
	KeySlash        Key = -1
	Key0            Key = -1
	Key1            Key = 49
	Key2            Key = 50
	Key3            Key = 51
	Key4            Key = -1
	Key5            Key = -1
	Key6            Key = -1
	Key7            Key = -1
	Key8            Key = -1
	Key9            Key = -1
	KeySemicolon    Key = -1
	KeyEqual        Key = -1
	KeyA            Key = 65
	KeyB            Key = -1
	KeyC            Key = -1
	KeyD            Key = 68
	KeyE            Key = 69
	KeyF            Key = -1
	KeyG            Key = -1
	KeyH            Key = -1
	KeyI            Key = -1
	KeyJ            Key = -1
	KeyK            Key = -1
	KeyL            Key = -1
	KeyM            Key = -1
	KeyN            Key = -1
	KeyO            Key = -1
	KeyP            Key = -1
	KeyQ            Key = 81
	KeyR            Key = -1
	KeyS            Key = 83
	KeyT            Key = -1
	KeyU            Key = -1
	KeyV            Key = -1
	KeyW            Key = 87
	KeyX            Key = -1
	KeyY            Key = -1
	KeyZ            Key = -1
	KeyLeftBracket  Key = -1
	KeyBackslash    Key = -1
	KeyRightBracket Key = -1
	KeyGraveAccent  Key = -1
	KeyWorld1       Key = -1
	KeyWorld2       Key = -1
	KeyEscape       Key = 27
	KeyEnter        Key = 13
	KeyTab          Key = 9
	KeyBackspace    Key = -1
	KeyInsert       Key = -1
	KeyDelete       Key = -1
	KeyRight        Key = 39
	KeyLeft         Key = 37
	KeyDown         Key = 40
	KeyUp           Key = 38
	KeyPageUp       Key = -1
	KeyPageDown     Key = -1
	KeyHome         Key = -1
	KeyEnd          Key = -1
	KeyCapsLock     Key = -1
	KeyScrollLock   Key = -1
	KeyNumLock      Key = -1
	KeyPrintScreen  Key = -1
	KeyPause        Key = -1
	KeyF1           Key = 112
	KeyF2           Key = 113
	KeyF3           Key = -1
	KeyF4           Key = -1
	KeyF5           Key = -1
	KeyF6           Key = -1
	KeyF7           Key = -1
	KeyF8           Key = -1
	KeyF9           Key = -1
	KeyF10          Key = -1
	KeyF11          Key = -1
	KeyF12          Key = -1
	KeyF13          Key = -1
	KeyF14          Key = -1
	KeyF15          Key = -1
	KeyF16          Key = -1
	KeyF17          Key = -1
	KeyF18          Key = -1
	KeyF19          Key = -1
	KeyF20          Key = -1
	KeyF21          Key = -1
	KeyF22          Key = -1
	KeyF23          Key = -1
	KeyF24          Key = -1
	KeyF25          Key = -1
	KeyKP0          Key = -1
	KeyKP1          Key = -1
	KeyKP2          Key = -1
	KeyKP3          Key = -1
	KeyKP4          Key = -1
	KeyKP5          Key = -1
	KeyKP6          Key = -1
	KeyKP7          Key = -1
	KeyKP8          Key = -1
	KeyKP9          Key = -1
	KeyKPDecimal    Key = -1
	KeyKPDivide     Key = -1
	KeyKPMultiply   Key = -1
	KeyKPSubtract   Key = -1
	KeyKPAdd        Key = -1
	KeyKPEnter      Key = -1
	KeyKPEqual      Key = -1
	KeyLeftShift    Key = 340
	KeyLeftControl  Key = -1
	KeyLeftAlt      Key = -1
	KeyLeftSuper    Key = -1
	KeyRightShift   Key = 344
	KeyRightControl Key = -1
	KeyRightAlt     Key = -1
	KeyRightSuper   Key = -1
	KeyMenu         Key = -1
)

type MouseButton int

const (
	MouseButton1 MouseButton = 0
	MouseButton2 MouseButton = 2 // Web MouseEvent has middle and right mouse buttons in reverse order.
	MouseButton3 MouseButton = 1 // Web MouseEvent has middle and right mouse buttons in reverse order.

	MouseButtonLeft   = MouseButton1
	MouseButtonRight  = MouseButton2
	MouseButtonMiddle = MouseButton3
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
	resp, err := http.Get(name)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("non-200 status: %s", resp.Status)
	}
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return nopCloser{bytes.NewReader(b)}, nil
}

type nopCloser struct {
	io.ReadSeeker
}

func (nopCloser) Close() error { return nil }

// ---

func WaitEvents() {
	// TODO.

	runtime.Gosched()
}

func PostEmptyEvent() {
	// TODO: Implement.
}

func DefaultWindowHints() {
	// TODO: Implement.
}

func (w *Window) SetClipboardString(str string) {
	// TODO: Implement.
}
func (w *Window) GetClipboardString() (string, error) {
	// TODO: Implement.
	return "", errors.New("GetClipboardString not implemented")
}

func (w *Window) SetTitle(title string) {
	document.SetTitle(title)
}

func (w *Window) Show() {
	// TODO: Implement.
}

func (w *Window) Hide() {
	// TODO: Implement.
}

func (w *Window) Destroy() {
	document.Body().RemoveChild(w.canvas)
	if w.fullscreen {
		if w.missing.fullscreen {
			log.Println("warning: Fullscreen API unsupported")
		} else {
			document.Underlying().Call("webkitExitFullscreen")
			w.fullscreen = false
		}
	}
}

type CloseCallback func(w *Window)

func (w *Window) SetCloseCallback(cbfun CloseCallback) (previous CloseCallback) {
	// TODO: Implement.

	// TODO: Handle previous.
	return nil
}

type RefreshCallback func(w *Window)

func (w *Window) SetRefreshCallback(cbfun RefreshCallback) (previous RefreshCallback) {
	// TODO: Implement.

	// TODO: Handle previous.
	return nil
}

type SizeCallback func(w *Window, width int, height int)

func (w *Window) SetSizeCallback(cbfun SizeCallback) (previous SizeCallback) {
	w.sizeCallback = cbfun

	// TODO: Handle previous.
	return nil
}

type CursorEnterCallback func(w *Window, entered bool)

func (w *Window) SetCursorEnterCallback(cbfun CursorEnterCallback) (previous CursorEnterCallback) {
	// TODO: Implement.

	// TODO: Handle previous.
	return nil
}

type CharModsCallback func(w *Window, char rune, mods ModifierKey)

func (w *Window) SetCharModsCallback(cbfun CharModsCallback) (previous CharModsCallback) {
	// TODO: Implement.

	// TODO: Handle previous.
	return nil
}

type PosCallback func(w *Window, xpos int, ypos int)

func (w *Window) SetPosCallback(cbfun PosCallback) (previous PosCallback) {
	// TODO: Implement.

	// TODO: Handle previous.
	return nil
}

type FocusCallback func(w *Window, focused bool)

func (w *Window) SetFocusCallback(cbfun FocusCallback) (previous FocusCallback) {
	// TODO: Implement.

	// TODO: Handle previous.
	return nil
}

type IconifyCallback func(w *Window, iconified bool)

func (w *Window) SetIconifyCallback(cbfun IconifyCallback) (previous IconifyCallback) {
	// TODO: Implement.

	// TODO: Handle previous.
	return nil
}

type DropCallback func(w *Window, names []string)

func (w *Window) SetDropCallback(cbfun DropCallback) (previous DropCallback) {
	// TODO: Implement. Can use HTML5 file drag and drop API?

	// TODO: Handle previous.
	return nil
}
