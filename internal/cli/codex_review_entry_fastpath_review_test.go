package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/hook"
	"github.com/modu-ai/moai-adk/internal/verify"
)

func TestCodexEntryFastpath_ReviewRejectsMalformedTimestamp(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	prefix := now.Format("2006-01-02T15:04:05")
	day := now.Format("2006-01-02")
	year := now.Format("2006")
	for _, tc := range []struct {
		name, timestamp, location string
		valid                     bool
	}{
		{"unknown suffix", prefix + "bogus", "check", false},
		{"missing zone", prefix, "check", false},
		{"truncated offset", prefix + "+01", "check", false},
		{"offset range", prefix + "+99:99", "check", false},
		{"month range", year + "-13-01T00:00:00Z", "check", false},
		{"day range", year + "-02-30T00:00:00Z", "check", false},
		{"hour range", day + "T24:00:00Z", "check", false},
		{"minute range", day + "T00:60:00Z", "check", false},
		{"second range", day + "T00:00:60Z", "check", false},
		{"trailing bytes", prefix + "Zbogus", "check", false},
		{"snapshot timestamp", prefix + "bogus", "snapshot", false},
		{"other check timestamp", prefix + "bogus", "other", false},
		{"valid UTC", now.Format(time.RFC3339Nano), "check", true},
		{"valid fraction", now.Add(123456789 * time.Nanosecond).Format(time.RFC3339Nano), "check", true},
		{"valid positive offset", now.In(time.FixedZone("east", 5*3600+30*60)).Format(time.RFC3339Nano), "check", true},
		{"valid negative offset", now.In(time.FixedZone("west", -4*3600)).Format(time.RFC3339Nano), "check", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := newCodexEntryFixture(t, codexEntryFixtureOpts{enabled: true, primaryScope: "review", dirty: true,
				stubVersion:     "moai-adk v3.2.0-1-g" + codexEntryHeadPlaceholder + "-dirty",
				reviewerVersion: "reviewer fixture v1",
				receipt:         &codexEntryReceiptSpec{verdict: "pass", recordedAt: now}})
			key, err := verify.Key(context.Background(), f.dir)
			if err != nil {
				t.Fatal(err)
			}
			path := verify.SnapshotPath(f.dir, key)
			raw, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			var snapshot map[string]any
			if err := json.Unmarshal(raw, &snapshot); err != nil {
				t.Fatal(err)
			}
			checks := snapshot["checks"].([]any)
			switch tc.location {
			case "snapshot":
				snapshot["recorded_at"] = tc.timestamp
			case "other":
				snapshot["checks"] = append(checks, map[string]any{"command": "another check", "recorded_at": tc.timestamp})
			default:
				checks[0].(map[string]any)["recorded_at"] = tc.timestamp
			}
			raw, err = json.MarshalIndent(snapshot, "", "  ")
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(path, raw, 0o644); err != nil {
				t.Fatal(err)
			}
			_, loadErr := verify.Load(f.dir, key)
			if tc.valid != (loadErr == nil) {
				t.Fatalf("Go snapshot authority: valid=%v load error=%v", tc.valid, loadErr)
			}
			wantForward := 1
			if tc.valid {
				wantForward = 0
			}
			run := runCodexEntryWrapper(t, f)
			forward, kick, argv := codexEntryAwaitArgv(t, f, wantForward, 0)
			if run.exitCode != 0 || forward != wantForward || kick != 0 {
				t.Fatalf("timestamp authority: exit=%d forward=%d kick=%d, want forward=%d kick=0\nargv:\n%s\nstderr:\n%s",
					run.exitCode, forward, kick, wantForward, argv, run.stderr)
			}
		})
	}
}

func TestCodexEntryFastpath_ReviewMixedRootsFallThrough(t *testing.T) {
	for _, projectFirst := range []bool{false, true} {
		t.Run(fmt.Sprintf("project_first=%v", projectFirst), func(t *testing.T) {
			f := newCodexEntryFixture(t, codexEntryFixtureOpts{enabled: true, primaryScope: "review", dirty: true,
				stubVersion: "moai-adk v3.2.0-1-g" + codexEntryHeadPlaceholder + "-dirty"})
			other := t.TempDir()
			path := filepath.Join(other, ".moai", "config", "sections", "workflow.yaml")
			if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(path, []byte("workflow:\n  codex:\n    review_gate:\n      enabled: false\n"), 0o644); err != nil {
				t.Fatal(err)
			}
			input := &hook.HookInput{CWD: f.dir, ProjectDir: other}
			if got := resolveProjectDirFromInput(input); got != other {
				t.Fatalf("Go config root = %q, want %q", got, other)
			}
			if got := reviewScopeSessionDir(input, other); got != f.dir {
				t.Fatalf("Go session root = %q, want %q", got, f.dir)
			}
			if !readCodexReviewGateEnabled(f.dir) || readCodexReviewGateEnabled(other) {
				t.Fatal("fixture must have an enabled cwd gate and disabled project_dir gate")
			}
			stdin := fmt.Sprintf(`{"cwd":%q,"project_dir":%q}`, f.dir, other)
			if projectFirst {
				stdin = fmt.Sprintf(`{"project_dir":%q,"cwd":%q}`, other, f.dir)
			}
			run := runCodexEntryWrapperStdin(t, f, stdin)
			forward, kick, argv := codexEntryAwaitArgv(t, f, 1, 0)
			if run.exitCode != 0 || forward != 1 || kick != 0 {
				t.Fatalf("mixed roots must defer to Go: exit=%d forward=%d kick=%d\nargv:\n%s\nstderr:\n%s",
					run.exitCode, forward, kick, argv, run.stderr)
			}
		})
	}
	t.Run("matching roots retain fast path", func(t *testing.T) {
		f := newCodexEntryFixture(t, codexEntryFixtureOpts{enabled: true, primaryScope: "review", dirty: true,
			stubVersion: "moai-adk v3.2.0-1-g" + codexEntryHeadPlaceholder + "-dirty"})
		run := runCodexEntryWrapperStdin(t, f, fmt.Sprintf(`{"cwd":%q,"project_dir":%q}`, f.dir, f.dir))
		forward, kick, argv := codexEntryAwaitArgv(t, f, 0, 1)
		if run.exitCode != 0 || forward != 0 || kick != 1 {
			t.Fatalf("matching roots retain fast path: exit=%d forward=%d kick=%d\nargv:\n%s\nstderr:\n%s",
				run.exitCode, forward, kick, argv, run.stderr)
		}
	})
}

func TestCodexEntryFastpath_ReviewYAMLPoliciesFallThrough(t *testing.T) {
	for _, tc := range []struct{ name, policy string }{
		{"alias", "policy: &skip skip\nworkflow:\n  codex:\n    review_gate:\n      enabled: true\n      tree_scope: *skip\n      primary_scope: review\n"},
		{"anchor", "workflow:\n  codex:\n    review_gate:\n      enabled: true\n      tree_scope: &skip skip\n      primary_scope: review\n"},
		{"merge", "policy: &skip\n  tree_scope: skip\nworkflow:\n  codex:\n    review_gate:\n      <<: *skip\n      enabled: true\n      primary_scope: review\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := newCodexEntryFixture(t, codexEntryFixtureOpts{enabled: true, primaryScope: "review", dirty: true,
				stubVersion: "moai-adk v3.2.0-1-g" + codexEntryHeadPlaceholder + "-dirty"})
			path := filepath.Join(f.dir, ".moai", "config", "sections", "workflow.yaml")
			if err := os.WriteFile(path, []byte(tc.policy), 0o644); err != nil {
				t.Fatal(err)
			}
			if got := readCodexReviewGateTreeScope(f.dir); got != "skip" {
				t.Fatalf("Go YAML policy = %q, want skip", got)
			}
			run := runCodexEntryWrapper(t, f)
			forward, kick, argv := codexEntryAwaitArgv(t, f, 1, 0)
			if run.exitCode != 0 || forward != 1 || kick != 0 {
				t.Fatalf("YAML policy must defer to Go: exit=%d forward=%d kick=%d\nargv:\n%s\nstderr:\n%s",
					run.exitCode, forward, kick, argv, run.stderr)
			}
		})
	}
}

// GNU awk reserves RT as the scalar record terminator. Give the real local
// awk the same scalar binding so this portability regression runs on mawk
// hosts too; this is not a native GNU awk execution claim.
func TestCodexEntryFastpath_AwkRecordTerminatorRemainsScalar(t *testing.T) {
	for _, tc := range []struct {
		name       string
		configOnly bool
		wantKick   int
	}{
		{"reviewable change kicks", false, 1},
		{"config-only change stays suppressed", true, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := newCodexEntryFixture(t, codexEntryFixtureOpts{enabled: true, primaryScope: "review", dirty: true,
				configOnly: tc.configOnly, stubVersion: "moai-adk v3.2.0-1-g" + codexEntryHeadPlaceholder + "-dirty"})
			awkPath := filepath.Join(f.toolsDir, "awk")
			realAwk, err := filepath.EvalSymlinks(awkPath)
			if err != nil {
				t.Fatal(err)
			}
			if err := os.Remove(awkPath); err != nil {
				t.Fatal(err)
			}
			shim := "#!/bin/bash\nexec " + codexEntryQuoteShell(realAwk) + " -v RT=record-terminator \"$@\"\n"
			if err := os.WriteFile(awkPath, []byte(shim), 0o755); err != nil {
				t.Fatal(err)
			}
			run := runCodexEntryWrapper(t, f)
			forward, kick, argv := codexEntryAwaitArgv(t, f, 0, tc.wantKick)
			if run.exitCode != 0 || forward != 0 || kick != tc.wantKick || strings.Contains(run.stderr, "RT") {
				t.Fatalf("scalar RT: exit=%d forward=%d kick=%d, want forward=0 kick=%d\nargv:\n%s\nstderr:\n%s",
					run.exitCode, forward, kick, tc.wantKick, argv, run.stderr)
			}
		})
	}
}

func TestCodexEntryFastpath_ReviewRejectsTrailingJSONData(t *testing.T) {
	for _, verdict := range []string{"pass", "inconclusive"} {
		for _, tc := range []struct {
			name, suffix string
			valid        bool
		}{
			{"junk on next line", "\njunk", false},
			{"second object", "\n{}", false},
			{"extra closing brace", "\n}", false},
			{"top-level comma", ",", false},
			{"same-line junk", "junk", false},
			{"non-JSON whitespace", "\n\f", false},
			{"whitespace", " \t\r\n\n", true},
		} {
			t.Run(verdict+"/"+tc.name, func(t *testing.T) {
				f := newCodexEntryFixture(t, codexEntryFixtureOpts{enabled: true, primaryScope: "review", dirty: true,
					stubVersion:     "moai-adk v3.2.0-1-g" + codexEntryHeadPlaceholder + "-dirty",
					reviewerVersion: "reviewer fixture v1",
					receipt:         &codexEntryReceiptSpec{verdict: verdict, recordedAt: time.Now()}})
				key, err := verify.Key(context.Background(), f.dir)
				if err != nil {
					t.Fatal(err)
				}
				path := verify.SnapshotPath(f.dir, key)
				raw, err := os.ReadFile(path)
				if err != nil {
					t.Fatal(err)
				}
				raw = append(raw, []byte(tc.suffix)...)
				if err := os.WriteFile(path, raw, 0o644); err != nil {
					t.Fatal(err)
				}
				_, loadErr := verify.Load(f.dir, key)
				if tc.valid != (loadErr == nil) {
					t.Fatalf("Go snapshot authority: valid=%v load error=%v", tc.valid, loadErr)
				}
				wantForward := 1
				if tc.valid {
					wantForward = 0
				}
				run := runCodexEntryWrapper(t, f)
				forward, kick, argv := codexEntryAwaitArgv(t, f, wantForward, 0)
				if run.exitCode != 0 || forward != wantForward || kick != 0 {
					t.Fatalf("trailing JSON: exit=%d forward=%d kick=%d, want forward=%d kick=0\nargv:\n%s\nstderr:\n%s",
						run.exitCode, forward, kick, wantForward, argv, run.stderr)
				}
			})
		}
	}
}
