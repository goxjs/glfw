// +build !js

package goglfw

import "github.com/go-gl/glfw/v3.1/glfw"

type Hint int

const (
	AlphaBits = Hint(glfw.AlphaBits)
	Samples   = Hint(glfw.Samples)
)

func WindowHint(target Hint, hint int) {
	glfw.WindowHint(glfw.Hint(target), hint)
}
