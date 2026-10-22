package cli

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/hook"
)

func TestCodexFindingAnchorPreservesWindowsPaths(t *testing.T) {
	for _, tc := range []struct {
		name, body, file string
		line             int
	}{
		{"drive slash", "`C:/repo/.claude/settings.json:13` drifted", "C:/repo/.claude/settings.json", 13},
		{"drive backslash", "`C:\\repo\\internal\\main.go:17` unsafe input", `C:\repo\internal\main.go`, 17},
		{"UNC", "`\\\\server\\share\\main.go:19` unsafe input", `\\server\share\main.go`, 19},
		{"distinct drives stay ambiguous", "`C:/repo/main.go:17` and `D:/repo/main.go:23` disagree", "", 0},
		{"title and body stay ambiguous", "`C:/repo/.claude/settings.json:13` drifted\n  fix `C:/repo/internal/main.go:25`", "", 0},
		{"URL is not a file anchor", "see https://example.invalid/internal/main.go:17", "", 0},
		{"URL does not hide a local anchor", "see https://example.invalid/guide.md:17 and `internal/main.go:23`", "internal/main.go", 23},
		{"relative path", "`internal/main.go:17` unsafe input", "internal/main.go", 17},
		{"repeat keeps first line", "`C:/repo/main.go:17` and `C:/repo/main.go:23`", "C:/repo/main.go", 17},
	} {
		t.Run(tc.name, func(t *testing.T) {
			file, line, ok := codexFindingAnchorOf(tc.body)
			if file != tc.file || line != tc.line || ok != (tc.file != "") {
				t.Fatalf("anchor = %q:%d, ok=%v; want %q:%d", file, line, ok, tc.file, tc.line)
			}
		})
	}
}

func TestStopParseCapValidPayloadRoundTripsNativePath(t *testing.T) {
	for _, event := range []hook.EventType{hook.EventStop, hook.EventPreToolUse, hook.EventPostToolUse} {
		const root = `C:\Users\fixture\project with spaces`
		var input hook.HookInput
		if err := json.Unmarshal([]byte(validPayload(event, root)), &input); err != nil {
			t.Fatalf("valid %s fixture is not JSON: %v", event, err)
		}
		if input.CWD != root || input.HookEventName != string(event) {
			t.Fatalf("valid fixture changed its event or native cwd: %+v", input)
		}
	}
}

func TestAutoLivenessNativePIDProbeWithoutExternalTools(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	lv := newAutoLiveness()
	if !lv.pidAlive(os.Getpid()) {
		t.Fatal("current process must be live without an external kill command")
	}
	for _, pid := range []int{-1, 0} {
		if lv.pidAlive(pid) {
			t.Errorf("invalid pid %d reported live", pid)
		}
	}
}

func TestCanonicalWorktreePathRefusesNonTrimStableInput(t *testing.T) {
	root := t.TempDir()
	for _, input := range []string{" " + root, root + " ", "\t" + root, root + "\t"} {
		if got, err := canonicalWorktreePath(input); err == nil {
			t.Errorf("non-trim-stable input %q was accepted as %q", input, got)
		}
	}
}

func TestAutoLivenessNativePIDKeepsOwnerWhenProcessScanUnavailable(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	root := t.TempDir()
	lv := newAutoLiveness()
	lv.registryEntries = func() ([]autoRegistryEntry, error) {
		return []autoRegistryEntry{{Cwd: filepath.Join(autoOwnerWorktrees(root, "t1")[0], "session"), PID: os.Getpid()}}, nil
	}
	lv.processCWDs = func() ([]string, error) { return nil, errors.New("process cwd scan unavailable") }
	alive, degraded, err := lv.ownerAlive(root, "t1")
	if !alive || !degraded || err != nil {
		t.Fatalf("live registry owner must prevent rescue when the other channel is unavailable: alive=%v degraded=%v err=%v", alive, degraded, err)
	}
}
