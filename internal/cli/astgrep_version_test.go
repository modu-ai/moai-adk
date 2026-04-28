package cli

// astgrep_version_test.go covers AC-UTIL-002-08/09.
// Tests run in package cli (not cli_test) to access unexported vars/funcs.
//
// sync.Once のリセット戦略:
//   sgVersionOnce は *sync.Once (ポインタ) なので、テストでは新しいポインタに
//   置き換えることでリセット相当を実現する (sync.Once 値のコピーを避ける)。

import (
	"strings"
	"sync"
	"testing"
)

// resetSGVersionCache は sgVersionOnce と sgVersionResult をリセットし、
// defer で元の状態に戻すためのクリーンアップ関数を返します。
func resetSGVersionCache(newExec func(string) (string, error)) func() {
	origOnce := sgVersionOnce   // save pointer (not copy)
	origResult := sgVersionResult
	origExec := sgVersionExec

	sgVersionOnce = new(sync.Once) // new instance, no copy
	sgVersionResult = ""
	if newExec != nil {
		sgVersionExec = newExec
	}

	return func() {
		sgVersionOnce = origOnce // restore pointer
		sgVersionResult = origResult
		sgVersionExec = origExec
	}
}

// TestDetectSGVersion_ReturnsUnknownWhenSGMissing verifies that detectSGVersion()
// returns "unknown" when the sg binary is not available. AC-UTIL-002-08
func TestDetectSGVersion_ReturnsUnknownWhenSGMissing(t *testing.T) {
	defer resetSGVersionCache(func(_ string) (string, error) {
		return "", &execError{msg: "exec: sg: executable file not found in $PATH"}
	})()

	v := detectSGVersion()
	if v != "unknown" {
		t.Errorf("detectSGVersion() = %q, want \"unknown\" when sg is missing", v)
	}
}

// TestDetectSGVersion_ParsesTrimmedOutput verifies that detectSGVersion() returns
// the trimmed stdout of sg --version. AC-UTIL-002-08
func TestDetectSGVersion_ParsesTrimmedOutput(t *testing.T) {
	defer resetSGVersionCache(func(_ string) (string, error) {
		return "ast-grep 0.40.5\n", nil
	})()

	v := detectSGVersion()
	want := "ast-grep 0.40.5"
	if v != want {
		t.Errorf("detectSGVersion() = %q, want %q", v, want)
	}
}

// TestDetectSGVersion_ReturnsUnknownOnEmptyOutput verifies that empty stdout
// yields "unknown". AC-UTIL-002-08
func TestDetectSGVersion_ReturnsUnknownOnEmptyOutput(t *testing.T) {
	defer resetSGVersionCache(func(_ string) (string, error) {
		return "   ", nil // all whitespace → trimmed to empty
	})()

	v := detectSGVersion()
	if v != "unknown" {
		t.Errorf("detectSGVersion() = %q, want \"unknown\" for empty output", v)
	}
}

// TestDetectSGVersion_CachesViaOnce verifies that the executor is called exactly
// once regardless of how many times detectSGVersion() is invoked. AC-UTIL-002-09
func TestDetectSGVersion_CachesViaOnce(t *testing.T) {
	callCount := 0
	defer resetSGVersionCache(func(_ string) (string, error) {
		callCount++
		return "ast-grep 0.42.1", nil
	})()

	const iterations = 10
	var results []string
	for range iterations {
		results = append(results, detectSGVersion())
	}

	if callCount != 1 {
		t.Errorf("executor called %d times, want 1 (sync.Once caching)", callCount)
	}

	for i, r := range results {
		if !strings.Contains(r, "ast-grep") {
			t.Errorf("call %d: detectSGVersion() = %q, want version string", i, r)
		}
	}
}

// execError is a lightweight error type for simulating exec errors in tests.
type execError struct {
	msg string
}

func (e *execError) Error() string { return e.msg }
