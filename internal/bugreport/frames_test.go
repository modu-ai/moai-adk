package bugreport

import (
	"reflect"
	"runtime"
	"testing"
)

// TestFramesExcludePathsLinesAndForeignModules pins the frame filter: only
// module-prefixed function names survive, with the prefix stripped; foreign
// modules, runtime internals, capture-package helpers, and path/line-shaped
// names never reach a payload.
func TestFramesExcludePathsLinesAndForeignModules(t *testing.T) {
	frames := []runtime.Frame{
		{Function: "runtime.gopanic"},
		{Function: "github.com/modu-ai/moai-adk/internal/bugreport.Capture"},      // capture helper dropped
		{Function: "github.com/modu-ai/moai-adk/internal/bugreport.framesFromPC"}, // capture helper dropped
		{Function: "github.com/modu-ai/moai-adk/internal/cli.Execute"},
		{Function: "somevendor.com/other/pkg.Do"},                     // foreign module dropped
		{Function: "github.com/modu-ai/moai-adk/internal/leak.go:42"}, // path/line shape dropped
		{Function: "github.com/modu-ai/moai-adk/internal/navigator/route/run.Recover"},
		{Function: ""}, // unnamed frame dropped
	}

	got := FilterFrames(frames)

	want := []string{
		"internal/cli.Execute",
		"internal/navigator/route/run.Recover",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("FilterFrames = %v, want %v", got, want)
	}

	// Every kept frame is a bare symbol path: never an absolute path, never a
	// line number, never argument text. (The shapes above already exercise
	// the exclusions; this assertion states the property the filter exists
	// for, so a future edit that keeps a wrong frame fails here too.)
	for _, name := range got {
		if err := ValidateFrameName(name); err != nil {
			t.Fatalf("kept frame %q fails the frame-name validator: %v", name, err)
		}
	}
}

// TestFrameLimitCapsInnermostFirst pins the 12-frame bound: with more moai
// frames than the limit, exactly the first FrameLimit in the stack's own order
// survive (runtime.Callers hands the innermost frame first, so the kept slice
// is the innermost 12).
func TestFrameLimitCapsInnermostFirst(t *testing.T) {
	var frames []runtime.Frame
	for i := 0; i < 30; i++ {
		frames = append(frames, runtime.Frame{
			Function: ModulePrefix + "internal/deep/pkg.Func" + itoa(i),
		})
	}

	got := FilterFrames(frames)

	if len(got) != 12 {
		t.Fatalf("frame count = %d, want the 12-frame limit", len(got))
	}
	if got[0] != "internal/deep/pkg.Func0" {
		t.Fatalf("first frame = %q, want Func0 (stack order preserved: innermost first)", got[0])
	}
	if got[11] != "internal/deep/pkg.Func11" {
		t.Fatalf("12th frame = %q, want Func11", got[11])
	}
}

// TestFunctionNamesWithGoSubstringAreKept pins the review finding's exact
// case: a function whose name merely CONTAINS the substring ".go"
// (cli.goalProjectRoot) is a function name, not a file path. Dropping it
// would lose fingerprint information; an anchored file-suffix check is what
// distinguishes the two.
func TestFunctionNamesWithGoSubstringAreKept(t *testing.T) {
	const name = "internal/cli.goalProjectRoot"
	if err := ValidateFrameName(name); err != nil {
		t.Fatalf("ValidateFrameName(%q) = %v, want accepted (function name, not a path)", name, err)
	}
	got := FilterFrames([]runtime.Frame{{Function: ModulePrefix + name}})
	if len(got) != 1 || got[0] != name {
		t.Fatalf("function frame dropped by the filter: %v, want [%q]", got, name)
	}
}

func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	digits := []byte{}
	for i > 0 {
		digits = append([]byte{byte('0' + i%10)}, digits...)
		i /= 10
	}
	return string(digits)
}
