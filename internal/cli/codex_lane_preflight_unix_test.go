//go:build !windows

// codex_lane_preflight_unix_test.go — card t1779 card-review round 2: the
// fake-ps sandbox test is Unix-only. The #!/bin/sh PATH injection and the
// real ps probe it drives (homestate.ProcessListProbeError returns nil
// unconditionally on Windows) have no Windows half, so the test rides the
// same !windows tag as the probe itself; the platform-neutral preflight
// tests stay in codex_lane_preflight_test.go untagged.

package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/homestate"
)

// TestCodexLaneLaunchEnvironmentRefusalSandboxFakePS — the END-TO-END
// sandbox path: a fake ps on PATH that fails the way a sandboxed ps does
// (stderr text + exit 1) drives the REAL homestate probe and the REAL
// signature match — not an injected string error, which is exactly what
// card-review P2-1 found the first test could not see (card t1779).
func TestCodexLaneLaunchEnvironmentRefusalSandboxFakePS(t *testing.T) {
	dir := t.TempDir()
	script := filepath.Join(dir, "ps")
	body := "#!/bin/sh\necho 'ps: Operation not permitted' >&2\nexit 1\n"
	if err := os.WriteFile(script, []byte(body), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir)

	prevProbe, prevStdio := codexLaneSandboxProbeFn, codexLaneStdioIsTerminalFn
	codexLaneSandboxProbeFn = func() error { return homestate.ProcessListProbeError(os.Getpid()) }
	codexLaneStdioIsTerminalFn = func() bool { return true }
	t.Cleanup(func() { codexLaneSandboxProbeFn, codexLaneStdioIsTerminalFn = prevProbe, prevStdio })

	if got := codexLaneLaunchEnvironmentRefusal([]string{"TERM=xterm-256color"}); !strings.Contains(got, "forbids process introspection") {
		t.Fatalf("sandboxed-ps refusal = %q, want the sandbox line", got)
	}
}
