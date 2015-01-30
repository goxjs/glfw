// +build !js

package goglfw

import glfw "github.com/shurcooL/glfw3"

type Hint int

const (
	AlphaBits = Hint(glfw.AlphaBits)
	Samples   = Hint(glfw.Samples)
)

func WindowHint(target Hint, hint int) {
	glfw.WindowHint(glfw.Hint(target), hint)
}
