// +build js

package goglfw

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gopherjs/gopherjs/js"
	"honnef.co/go/js/dom"
)

var document = dom.GetWindow().Document().(dom.HTMLDocument)

func Init() error {
	document.Body().Style().SetProperty("margin", "0px", "")

	return nil
}

func Terminate() error {
	return nil
}

func CreateWindow(width, height int, title string, monitor *Monitor, share *Window) (*Window, error) {
	canvas := document.CreateElement("canvas").(*dom.HTMLCanvasElement)
	devicePixelRatio := js.Global.Get("devicePixelRatio").Float()
	canvas.Width = int(float64(width)*devicePixelRatio + 0.5)   // Nearest non-negative int.
	canvas.Height = int(float64(height)*devicePixelRatio + 0.5) // Nearest non-negative int.
	canvas.Style().SetProperty("width", fmt.Sprintf("%vpx", width), "")
	canvas.Style().SetProperty("height", fmt.Sprintf("%vpx", height), "")
	document.Body().AppendChild(canvas)

	document.SetTitle(title)

	// DEBUG: Add framebuffer information div.
	text := document.CreateElement("div")
	textContent := fmt.Sprintf("%v %v (%v) @%v", dom.GetWindow().InnerWidth(), canvas.Width, float64(width)*devicePixelRatio, devicePixelRatio)
	text.SetTextContent(textContent)
	document.Body().AppendChild(text)

	// TODO: A part of this should go into SetFramebufferSizeCallback and friends.
	/*dom.GetWindow().AddEventListener("resize", false, func(event dom.Event) {
		devicePixelRatio := js.Global.Get("devicePixelRatio").Float()
		canvas.Width = int(float64(width)*devicePixelRatio + 0.5)   // Nearest non-negative int.
		canvas.Height = int(float64(height)*devicePixelRatio + 0.5) // Nearest non-negative int.
		textContent := fmt.Sprintf("%v %v (%v) @%v", dom.GetWindow().InnerWidth(), canvas.Width, float64(width)*devicePixelRatio, devicePixelRatio)
		text.SetTextContent(textContent)
	})*/

	w := &Window{
		Canvas: canvas,
	}

	document.AddEventListener("mousedown", false, func(event dom.Event) {
		me := event.(*dom.MouseEvent)
		if !(me.Button >= 0 && me.Button <= 2) {
			return
		}

		w.mouseButton[me.Button] = Press

		me.PreventDefault()
	})
	document.AddEventListener("mouseup", false, func(event dom.Event) {
		me := event.(*dom.MouseEvent)
		if !(me.Button >= 0 && me.Button <= 2) {
			return
		}

		w.mouseButton[me.Button] = Release

		me.PreventDefault()
	})
	document.AddEventListener("contextmenu", false, func(event dom.Event) {
		event.PreventDefault()
	})

	// Request first animation frame.
	js.Global.Call("requestAnimationFrame", animationFrame)

	return w, nil
}

type Window struct {
	Canvas *dom.HTMLCanvasElement

	cursorPosition [2]float64
	mouseButton    [3]Action
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
	document.AddEventListener("mousemove", false, func(event dom.Event) {
		me := event.(*dom.MouseEvent)
		w.cursorPosition[0], w.cursorPosition[1] = float64(me.ClientX), float64(me.ClientY)
		cbfun(w, w.cursorPosition[0], w.cursorPosition[1])
	})

	// TODO: Handle previous.
	return nil, nil
}

type FramebufferSizeCallback func(w *Window, width int, height int)

func (w *Window) SetFramebufferSizeCallback(cbfun FramebufferSizeCallback) (previous FramebufferSizeCallback, err error) {
	// TODO: Actually set the callback.

	// TODO: Handle previous.
	return nil, err
}

func (w *Window) GetSize() (width, height int, err error) {
	// TODO: Handle units in a better, more general way.
	//       Currently assumes "px" units.
	widthString := strings.TrimSuffix(w.Canvas.Style().GetPropertyValue("width"), "px")
	heightString := strings.TrimSuffix(w.Canvas.Style().GetPropertyValue("height"), "px")

	width, err = strconv.Atoi(widthString)
	if err != nil {
		return 0, 0, err
	}
	height, err = strconv.Atoi(heightString)
	if err != nil {
		return 0, 0, err
	}

	return width, height, nil
}

func (w *Window) GetFramebufferSize() (width, height int, err error) {
	return w.Canvas.Width, w.Canvas.Height, nil
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

	return w.mouseButton[button], nil
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
