// +build js

package goglfw

var hints = make(map[Hint]int)

type Hint int

const (
	AlphaBits Hint = 0x00021004
	Samples   Hint = 0x0002100D
)

func WindowHint(target Hint, hint int) error {
	hints[target] = hint
	return nil
}
