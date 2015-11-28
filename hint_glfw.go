// +build !js

package glfw

import "github.com/go-gl/glfw/v3.1/glfw"

type Hint int

const (
	AlphaBits                            = Hint(glfw.AlphaBits)
	Samples                              = Hint(glfw.Samples)
	Depth                           Hint = 0x00021100 // it works on WebGL and is ignored on other environment
	Stencil                         Hint = 0x00021101 // it works on WebGL and is ignored on other environment
	PremultipliedAlpha              Hint = 0x00021102 // it works on WebGL and is ignored on other environment
	PreserveDrawingBuffer           Hint = 0x00021103 // it works on WebGL and is ignored on other environment
	PreferLowPowerToHighPerformance Hint = 0x00021104 // it works on WebGL and is ignored on other environment
	FailIfMajorPerformanceCaveat    Hint = 0x00021105 // it works on WebGL and is ignored on other environment
)

func WindowHint(target Hint, hint int) {
	// ignores hints it works only on browser
	if Depth <= target && target <= FailIfMajorPerformanceCaveat {
		return
	}
	glfw.WindowHint(glfw.Hint(target), hint)
}
