// +build js

package glfw

var hints = make(map[Hint]int)

type Hint int

const (
	AlphaBits Hint = iota
	DepthBits
	StencilBits
	Samples

	// goxjs/glfw-specific hints for WebGL.
	PremultipliedAlpha
	PreserveDrawingBuffer
	PreferLowPowerToHighPerformance
	FailIfMajorPerformanceCaveat
)

func WindowHint(target Hint, hint int) {
	hints[target] = hint
}
