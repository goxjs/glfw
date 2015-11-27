// +build js

package glfw

var hints = make(map[Hint]int)

type Hint int

const (
	AlphaBits                       Hint = 0x00021004
	Samples                         Hint = 0x0002100D
	Depth                           Hint = 0x00021100 // github.com/goxjs/glfw original hint for JS
	Stencil                         Hint = 0x00021101 // github.com/goxjs/glfw original hint for JS
	PremultipliedAlpha              Hint = 0x00021102 // github.com/goxjs/glfw original hint for JS
	PreserveDrawingBuffer           Hint = 0x00021103 // github.com/goxjs/glfw original hint for JS
	PreferLowPowerToHighPerformance Hint = 0x00021104 // github.com/goxjs/glfw original hint for JS
	FailIfMajorPerformanceCaveat    Hint = 0x00021105 // github.com/goxjs/glfw original hint for JS
)

func WindowHint(target Hint, hint int) {
	hints[target] = hint
}
