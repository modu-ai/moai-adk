// Package captureframesprobe is a test-only fixture package: it exists so
// the capture-frames test can call Capture from a frame OUTSIDE the capture
// package (the frame filter drops every internal/bugreport frame by design,
// so an in-package probe would be filtered out of its own assertion).
// Nothing in production imports it.
package captureframesprobe

import "github.com/modu-ai/moai-adk/internal/bugreport"

// Probe calls Capture from this package's frame — the caller frame the
// capture-frames test asserts on.
func Probe() {
	bugreport.Capture(bugreport.KindPanic, nil, "", nil)
}
