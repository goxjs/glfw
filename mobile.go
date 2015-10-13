// +build mobile

package glfw

import (
	"log"
	"math"
	"os"
	"runtime"
	"time"

	"github.com/go-gl/glfw/v3.1/glfw"
	"golang.org/x/mobile/app"
	"golang.org/x/mobile/event/lifecycle"
	"golang.org/x/mobile/event/paint"
	"golang.org/x/mobile/event/size"
	"golang.org/x/mobile/event/touch"
	"golang.org/x/mobile/gl"
	"golang.org/x/tools/godoc/vfs"
)

func init() {
	runtime.LockOSThread()
}

var contextWatcher ContextWatcher

// Init initializes the library.
//
// A valid ContextWatcher must be provided. It gets notified when context becomes current or detached.
// It should be provided by the GL bindings you are using, so you can do glfw.Init(gl.ContextWatcher).
func Init(cw ContextWatcher) error {
	contextWatcher = cw
	log.Println("Init: not implemented")
	return nil
}

func Terminate() {
	log.Println("Terminate: not implemented")
}

func CreateWindow(width, height int, title string, monitor *Monitor, share *Window) (*Window, error) {
	// TODO.
	log.Println("CreateWindow start...")
	defer log.Println("                  ...end")

	// THINK: Problem here. This is blocking, yet it can't be in a goroutine because app.main needs to be on main thread.
	//
	//        app.Main called on thread 458207, but app.init ran on 458203
	//        exit status 1
	//
	//        What to do? Perhaps using internals of app instead of trying to import it may be the only viable choice?
	app.Main(func(a app.App) {
		var glctx gl.Context
		var sz size.Event
		for e := range a.Events() {
			switch e := a.Filter(e).(type) {
			case lifecycle.Event:
				switch e.Crosses(lifecycle.StageVisible) {
				case lifecycle.CrossOn:
					glctx, _ = e.DrawContext.(gl.Context)
					onStart(glctx)
					a.Send(paint.Event{})
				case lifecycle.CrossOff:
					onStop(glctx)
					glctx = nil
				}
			case size.Event:
				sz = e
				//touchX = float32(sz.WidthPx / 2)
				//touchY = float32(sz.HeightPx / 2)
			case paint.Event:
				if glctx == nil || e.External {
					// As we are actively painting as fast as
					// we can (usually 60 FPS), skip any paint
					// events sent by the system.
					continue
				}

				onPaint(glctx, sz)
				a.Publish()
				// Drive the animation by preparing to paint the next frame
				// after this one is shown.
				a.Send(paint.Event{})
			case touch.Event:
				//touchX = e.X
				//touchY = e.Y
			}
		}
	})

	window := &Window{}
	return window, nil
}

var started = time.Now()

// TODO.
func onStart(glctx gl.Context) {
	contextWatcher.OnMakeCurrent(glctx)
}
func onStop(glctx gl.Context) {
	contextWatcher.OnDetach()
}
func onPaint(glctx gl.Context, sz size.Event) {
	glctx.ClearColor(0.8*(0.5*(1.0+float32(math.Sin(time.Since(started).Seconds())))), 0.3, 0.01, 1)
	glctx.Clear(gl.COLOR_BUFFER_BIT)
}

func SwapInterval(interval int) {
	glfw.SwapInterval(interval)
}

func (w *Window) MakeContextCurrent() {
	log.Println("MakeContextCurrent: not implemented")
	// In reality, context is available on each platform via GetGLXContext, GetWGLContext, GetNSGLContext, etc.
	// Pretend it is not available and pass nil, since it's not actually needed at this time.
	contextWatcher.OnMakeCurrent(nil)
}

func DetachCurrentContext() {
	log.Println("DetachCurrentContext: not implemented")
	contextWatcher.OnDetach()
}

type Window struct {
}

func (w *Window) ShouldClose() bool {
	// TODO.
	return false
}

func (w *Window) GetSize() (width, height int) {
	log.Println("GetSize: not implemented")
	return 0, 0
}

func (w *Window) GetFramebufferSize() (width, height int) {
	log.Println("GetFramebufferSize: not implemented")
	return 0, 0
}

type Monitor struct {
}

func GetPrimaryMonitor() *Monitor {
	panic("not implemented")
}

func PollEvents() {
	// TODO.
	time.Sleep(time.Second / 60)
}

type CursorPosCallback func(w *Window, xpos float64, ypos float64)

func (w *Window) SetCursorPosCallback(cbfun CursorPosCallback) (previous CursorPosCallback) {
	log.Println("SetCursorPosCallback: not implemented")
	return nil
}

type MouseMovementCallback func(w *Window, xpos float64, ypos float64, xdelta float64, ydelta float64)

var lastMousePos [2]float64 // HACK.

func (w *Window) SetMouseMovementCallback(cbfun MouseMovementCallback) (previous MouseMovementCallback) {
	panic("not implemented")
}

type KeyCallback func(w *Window, key Key, scancode int, action Action, mods ModifierKey)

func (w *Window) SetKeyCallback(cbfun KeyCallback) (previous KeyCallback) {
	panic("not implemented")
}

type CharCallback func(w *Window, char rune)

func (w *Window) SetCharCallback(cbfun CharCallback) (previous CharCallback) {
	panic("not implemented")
}

type ScrollCallback func(w *Window, xoff float64, yoff float64)

func (w *Window) SetScrollCallback(cbfun ScrollCallback) (previous ScrollCallback) {
	panic("not implemented")
}

type MouseButtonCallback func(w *Window, button MouseButton, action Action, mods ModifierKey)

func (w *Window) SetMouseButtonCallback(cbfun MouseButtonCallback) (previous MouseButtonCallback) {
	panic("not implemented")
}

type FramebufferSizeCallback func(w *Window, width int, height int)

func (w *Window) SetFramebufferSizeCallback(cbfun FramebufferSizeCallback) (previous FramebufferSizeCallback) {
	log.Println("SetFramebufferSizeCallback: not implemented")
	return nil
}

func (w *Window) GetKey(key Key) Action {
	panic("not implemented")
}

func (w *Window) GetMouseButton(button MouseButton) Action {
	panic("not implemented")
}

func (w *Window) GetInputMode(mode InputMode) int {
	panic("not implemented")
}

func (w *Window) SetInputMode(mode InputMode, value int) {
	panic("not implemented")
}

type Key glfw.Key

// TODO: Use x/mobile/app stuff.

const (
	KeySpace        = Key(glfw.KeySpace)
	KeyApostrophe   = Key(glfw.KeyApostrophe)
	KeyComma        = Key(glfw.KeyComma)
	KeyMinus        = Key(glfw.KeyMinus)
	KeyPeriod       = Key(glfw.KeyPeriod)
	KeySlash        = Key(glfw.KeySlash)
	Key0            = Key(glfw.Key0)
	Key1            = Key(glfw.Key1)
	Key2            = Key(glfw.Key2)
	Key3            = Key(glfw.Key3)
	Key4            = Key(glfw.Key4)
	Key5            = Key(glfw.Key5)
	Key6            = Key(glfw.Key6)
	Key7            = Key(glfw.Key7)
	Key8            = Key(glfw.Key8)
	Key9            = Key(glfw.Key9)
	KeySemicolon    = Key(glfw.KeySemicolon)
	KeyEqual        = Key(glfw.KeyEqual)
	KeyA            = Key(glfw.KeyA)
	KeyB            = Key(glfw.KeyB)
	KeyC            = Key(glfw.KeyC)
	KeyD            = Key(glfw.KeyD)
	KeyE            = Key(glfw.KeyE)
	KeyF            = Key(glfw.KeyF)
	KeyG            = Key(glfw.KeyG)
	KeyH            = Key(glfw.KeyH)
	KeyI            = Key(glfw.KeyI)
	KeyJ            = Key(glfw.KeyJ)
	KeyK            = Key(glfw.KeyK)
	KeyL            = Key(glfw.KeyL)
	KeyM            = Key(glfw.KeyM)
	KeyN            = Key(glfw.KeyN)
	KeyO            = Key(glfw.KeyO)
	KeyP            = Key(glfw.KeyP)
	KeyQ            = Key(glfw.KeyQ)
	KeyR            = Key(glfw.KeyR)
	KeyS            = Key(glfw.KeyS)
	KeyT            = Key(glfw.KeyT)
	KeyU            = Key(glfw.KeyU)
	KeyV            = Key(glfw.KeyV)
	KeyW            = Key(glfw.KeyW)
	KeyX            = Key(glfw.KeyX)
	KeyY            = Key(glfw.KeyY)
	KeyZ            = Key(glfw.KeyZ)
	KeyLeftBracket  = Key(glfw.KeyLeftBracket)
	KeyBackslash    = Key(glfw.KeyBackslash)
	KeyRightBracket = Key(glfw.KeyRightBracket)
	KeyGraveAccent  = Key(glfw.KeyGraveAccent)
	KeyWorld1       = Key(glfw.KeyWorld1)
	KeyWorld2       = Key(glfw.KeyWorld2)
	KeyEscape       = Key(glfw.KeyEscape)
	KeyEnter        = Key(glfw.KeyEnter)
	KeyTab          = Key(glfw.KeyTab)
	KeyBackspace    = Key(glfw.KeyBackspace)
	KeyInsert       = Key(glfw.KeyInsert)
	KeyDelete       = Key(glfw.KeyDelete)
	KeyRight        = Key(glfw.KeyRight)
	KeyLeft         = Key(glfw.KeyLeft)
	KeyDown         = Key(glfw.KeyDown)
	KeyUp           = Key(glfw.KeyUp)
	KeyPageUp       = Key(glfw.KeyPageUp)
	KeyPageDown     = Key(glfw.KeyPageDown)
	KeyHome         = Key(glfw.KeyHome)
	KeyEnd          = Key(glfw.KeyEnd)
	KeyCapsLock     = Key(glfw.KeyCapsLock)
	KeyScrollLock   = Key(glfw.KeyScrollLock)
	KeyNumLock      = Key(glfw.KeyNumLock)
	KeyPrintScreen  = Key(glfw.KeyPrintScreen)
	KeyPause        = Key(glfw.KeyPause)
	KeyF1           = Key(glfw.KeyF1)
	KeyF2           = Key(glfw.KeyF2)
	KeyF3           = Key(glfw.KeyF3)
	KeyF4           = Key(glfw.KeyF4)
	KeyF5           = Key(glfw.KeyF5)
	KeyF6           = Key(glfw.KeyF6)
	KeyF7           = Key(glfw.KeyF7)
	KeyF8           = Key(glfw.KeyF8)
	KeyF9           = Key(glfw.KeyF9)
	KeyF10          = Key(glfw.KeyF10)
	KeyF11          = Key(glfw.KeyF11)
	KeyF12          = Key(glfw.KeyF12)
	KeyF13          = Key(glfw.KeyF13)
	KeyF14          = Key(glfw.KeyF14)
	KeyF15          = Key(glfw.KeyF15)
	KeyF16          = Key(glfw.KeyF16)
	KeyF17          = Key(glfw.KeyF17)
	KeyF18          = Key(glfw.KeyF18)
	KeyF19          = Key(glfw.KeyF19)
	KeyF20          = Key(glfw.KeyF20)
	KeyF21          = Key(glfw.KeyF21)
	KeyF22          = Key(glfw.KeyF22)
	KeyF23          = Key(glfw.KeyF23)
	KeyF24          = Key(glfw.KeyF24)
	KeyF25          = Key(glfw.KeyF25)
	KeyKP0          = Key(glfw.KeyKP0)
	KeyKP1          = Key(glfw.KeyKP1)
	KeyKP2          = Key(glfw.KeyKP2)
	KeyKP3          = Key(glfw.KeyKP3)
	KeyKP4          = Key(glfw.KeyKP4)
	KeyKP5          = Key(glfw.KeyKP5)
	KeyKP6          = Key(glfw.KeyKP6)
	KeyKP7          = Key(glfw.KeyKP7)
	KeyKP8          = Key(glfw.KeyKP8)
	KeyKP9          = Key(glfw.KeyKP9)
	KeyKPDecimal    = Key(glfw.KeyKPDecimal)
	KeyKPDivide     = Key(glfw.KeyKPDivide)
	KeyKPMultiply   = Key(glfw.KeyKPMultiply)
	KeyKPSubtract   = Key(glfw.KeyKPSubtract)
	KeyKPAdd        = Key(glfw.KeyKPAdd)
	KeyKPEnter      = Key(glfw.KeyKPEnter)
	KeyKPEqual      = Key(glfw.KeyKPEqual)
	KeyLeftShift    = Key(glfw.KeyLeftShift)
	KeyLeftControl  = Key(glfw.KeyLeftControl)
	KeyLeftAlt      = Key(glfw.KeyLeftAlt)
	KeyLeftSuper    = Key(glfw.KeyLeftSuper)
	KeyRightShift   = Key(glfw.KeyRightShift)
	KeyRightControl = Key(glfw.KeyRightControl)
	KeyRightAlt     = Key(glfw.KeyRightAlt)
	KeyRightSuper   = Key(glfw.KeyRightSuper)
	KeyMenu         = Key(glfw.KeyMenu)
)

type MouseButton glfw.MouseButton

const (
	MouseButton1 = MouseButton(glfw.MouseButton1)
	MouseButton2 = MouseButton(glfw.MouseButton2)
	MouseButton3 = MouseButton(glfw.MouseButton3)

	MouseButtonLeft   = MouseButton(glfw.MouseButtonLeft)
	MouseButtonRight  = MouseButton(glfw.MouseButtonRight)
	MouseButtonMiddle = MouseButton(glfw.MouseButtonMiddle)
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
	// TODO: Support mobile.
	return os.Open(name)
}

// ---

func WaitEvents() {
	panic("not implemented")
}

func PostEmptyEvent() {
	panic("not implemented")
}

func DefaultWindowHints() {
	panic("not implemented")
}

type CloseCallback func(w *Window)

func (w *Window) SetCloseCallback(cbfun CloseCallback) (previous CloseCallback) {
	panic("not implemented")
}

type RefreshCallback func(w *Window)

func (w *Window) SetRefreshCallback(cbfun RefreshCallback) (previous RefreshCallback) {
	panic("not implemented")
}

type SizeCallback func(w *Window, width int, height int)

func (w *Window) SetSizeCallback(cbfun SizeCallback) (previous SizeCallback) {
	panic("not implemented")
}

type CursorEnterCallback func(w *Window, entered bool)

func (w *Window) SetCursorEnterCallback(cbfun CursorEnterCallback) (previous CursorEnterCallback) {
	panic("not implemented")
}

type CharModsCallback func(w *Window, char rune, mods ModifierKey)

func (w *Window) SetCharModsCallback(cbfun CharModsCallback) (previous CharModsCallback) {
	panic("not implemented")
}

type PosCallback func(w *Window, xpos int, ypos int)

func (w *Window) SetPosCallback(cbfun PosCallback) (previous PosCallback) {
	panic("not implemented")
}

type FocusCallback func(w *Window, focused bool)

func (w *Window) SetFocusCallback(cbfun FocusCallback) (previous FocusCallback) {
	panic("not implemented")
}

type IconifyCallback func(w *Window, iconified bool)

func (w *Window) SetIconifyCallback(cbfun IconifyCallback) (previous IconifyCallback) {
	panic("not implemented")
}

type DropCallback func(w *Window, names []string)

func (w *Window) SetDropCallback(cbfun DropCallback) (previous DropCallback) {
	panic("not implemented")
}
