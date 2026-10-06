package bugreport

import (
	"errors"
	"runtime"
	"strings"
	"testing"
)

// baseFingerprintInput returns the canonical fingerprint fixture the stability
// and sensitivity tests share. Every field is a valid, representative value so
// the tests measure the fingerprint, not the validators.
func baseFingerprintInput() CanonicalInput {
	return CanonicalInput{
		Build:  BuildIdentity{Version: "v3.2.0", Commit: "abcdef1234567"},
		OS:     "darwin",
		Arch:   "arm64",
		Kind:   KindPanic,
		Frames: []string{"internal/cli.Execute", "internal/navigator/route/run.Recover"},
	}
}

func TestFingerprintStable(t *testing.T) {
	in := baseFingerprintInput()

	a := Fingerprint(in)
	b := Fingerprint(in)

	if len(a) != 16 {
		t.Fatalf("fingerprint length = %d, want 16 hex characters (%q)", len(a), a)
	}
	for _, r := range a {
		isHex := (r >= '0' && r <= '9') || (r >= 'a' && r <= 'f')
		if !isHex {
			t.Fatalf("fingerprint %q carries non-lowercase-hex character %q", a, r)
		}
	}
	if a != b {
		t.Fatalf("fingerprint not stable: %q then %q for identical input", a, b)
	}
}

func TestFingerprintInputsAllMatter(t *testing.T) {
	base := baseFingerprintInput()
	baseFP := Fingerprint(base)

	cases := []struct {
		name  string
		mutat func(*CanonicalInput)
	}{
		{"version", func(c *CanonicalInput) { c.Build.Version = "v3.2.1" }},
		{"commit", func(c *CanonicalInput) { c.Build.Commit = "abcdef1234568" }},
		{"os", func(c *CanonicalInput) { c.OS = "windows" }},
		{"arch", func(c *CanonicalInput) { c.Arch = "amd64" }},
		{"kind", func(c *CanonicalInput) { c.Kind = KindInternalError }},
		{"frame_changed", func(c *CanonicalInput) { c.Frames[0] = "internal/cli.Root.Execute" }},
		{"frame_added", func(c *CanonicalInput) { c.Frames = append(c.Frames, "internal/hook.Dispatch") }},
		{"frame_dropped", func(c *CanonicalInput) { c.Frames = c.Frames[:1] }},
		{"frame_reordered", func(c *CanonicalInput) {
			c.Frames = []string{c.Frames[1], c.Frames[0]}
		}},
	}
	for _, tc := range cases {
		mutated := base
		tc.mutat(&mutated)
		if got := Fingerprint(mutated); got == baseFP {
			t.Errorf("%s: fingerprint unchanged after mutating %s (%q)", tc.name, tc.name, got)
		}
	}
}

func TestZeroFramePanicStaysLocal(t *testing.T) {
	// A stack with no moai frame at all filters to nothing.
	got := FilterFrames([]runtime.Frame{
		{Function: "runtime.gopanic"},
		{Function: "foreign.example/pkg.Do"},
		{Function: "github.com/someone-else/module.Client.Call"},
	})
	if len(got) != 0 {
		t.Fatalf("foreign-only stack filtered to %v, want empty", got)
	}

	// The local-only predicate fires for a panic with zero moai frames...
	if !IsLocalOnly(KindPanic, nil) {
		t.Fatal("IsLocalOnly(KindPanic, nil) = false, want true")
	}
	if !IsLocalOnly(KindPanic, []string{}) {
		t.Fatal("IsLocalOnly(KindPanic, []) = false, want true")
	}
	// ...and never for another kind: those capture from a moai call site,
	// so an empty frame list for them means the caller passed something wrong,
	// not that the signal is local (the Build validator refuses it instead).
	if IsLocalOnly(KindHookTimeout, nil) {
		t.Fatal("IsLocalOnly(KindHookTimeout, nil) = true, want false")
	}

	// Building a payload for a zero-frame panic is refused outright: the
	// signal stays local, no payload exists to queue or send.
	_, err := Build(KindPanic, nil, nil, BuildIdentity{Version: "v3.2.0", Commit: "abcdef1234567"})
	if !errors.Is(err, ErrZeroFramesStayLocal) {
		t.Fatalf("Build(KindPanic, no frames) error = %v, want ErrZeroFramesStayLocal", err)
	}
}

// genericCanary is the M1 canary: a generic function that captures its own
// runtime frame so the test can pin how the runtime names a generic frame.
// The design's premise (inferred when written) is that the runtime prints
// "[...]" for the type arguments instead of the concrete types — observed and
// pinned here rather than assumed.
func genericCanary[T any](_ T) string {
	var pc [16]uintptr
	n := runtime.Callers(1, pc[:])
	frames := runtime.CallersFrames(pc[:n])
	for {
		f, more := frames.Next()
		if strings.Contains(f.Function, "genericCanary") {
			return f.Function
		}
		if !more {
			break
		}
	}
	return ""
}

func TestGenericFunctionFrameNameShape(t *testing.T) {
	name := genericCanary(42)
	if name == "" {
		t.Fatal("generic canary did not find its own frame")
	}

	// Pinned shape (observed 2026-10-07, go 1.26: "main.genericCanary[...]"
	// in the canary probe): the runtime prints the generic symbol with "[...]"
	// for the type arguments — the concrete type argument values never enter
	// the name. This is the property the design relies on when it reads
	// Function and nothing else from a frame.
	if !strings.Contains(name, "[...]") {
		t.Fatalf("generic frame name %q does not carry the [...] placeholder; re-pin the shape the runtime actually prints", name)
	}

	// The canary frame itself is a capture-package frame, so the filter
	// drops it — that is the drop-helper rule working, and this pins it.
	if got := FilterFrames([]runtime.Frame{{Function: name}}); len(got) != 0 {
		t.Fatalf("capture-package generic frame survived the filter: %v", got)
	}

	// A generic frame in a NON-capture package survives with the module
	// prefix stripped — brackets and dots included.
	elsewhere := ModulePrefix + "internal/somepkg.GenericHelper[...]"
	filtered := FilterFrames([]runtime.Frame{{Function: elsewhere}})
	want := "internal/somepkg.GenericHelper[...]"
	if len(filtered) != 1 || filtered[0] != want {
		t.Fatalf("filtered generic frame = %v, want [%q]", filtered, want)
	}
}
