// +build js

package goglfw

import (
	"fmt"

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
	canvas.Width = int(float64(width)*devicePixelRatio + 0.5)   // Nearest int.
	canvas.Height = int(float64(height)*devicePixelRatio + 0.5) // Nearest int.
	canvas.Style().SetProperty("width", fmt.Sprintf("%vpx", width), "")
	canvas.Style().SetProperty("height", fmt.Sprintf("%vpx", height), "")
	document.Body().AppendChild(canvas)

	text := document.CreateElement("div")
	textContent := fmt.Sprintf("%v %v (%v) @%v", dom.GetWindow().InnerWidth(), canvas.Width, float64(width)*devicePixelRatio, devicePixelRatio)
	text.SetTextContent(textContent)
	document.Body().AppendChild(text)

	return &Window{canvas}, nil
}

type Window struct {
	Canvas *dom.HTMLCanvasElement
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
		cbfun(w, float64(me.ClientX), float64(me.ClientY))
	})

	// TODO: Handle previous.
	return nil, nil
}

func (w *Window) ShouldClose() (bool, error) {
	return false, nil
}
