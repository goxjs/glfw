// Package glfw experimentally provides a glfw-like API
// with desktop (via glfw) and browser (via HTML5 canvas) backends.
//
// It is used for creating a GL context and receiving events.
package glfw

// ContextSwitcher is a general mechanism for switching between contexts.
type ContextSwitcher interface {
	// MakeContextCurrent takes a context and makes it current.
	// If given context is nil, then the current context is detached.
	MakeContextCurrent(context interface{})
}
