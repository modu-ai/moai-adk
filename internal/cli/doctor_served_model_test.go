package cli

import (
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/cli/uikit"
	"github.com/modu-ai/moai-adk/internal/config"
)

func servedRow(model string) string {
	b, _ := json.Marshal(map[string]any{"type": "assistant", "message": map[string]any{"model": model}})
	return string(b)
}

// writeServedFixture writes one subagent transcript plus its meta.json under
// <base>/projects/<slug>/<session>/subagents/.
func writeServedFixture(t *testing.T, base, slug, session, agentID, agentType string, rows []string) {
	t.Helper()
	dir := filepath.Join(base, "projects", slug, session, "subagents")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "agent-"+agentID+".jsonl"), []byte(strings.Join(rows, "\n")+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	meta, _ := json.Marshal(map[string]string{"agentType": agentType, "model": "opus"})
	if err := os.WriteFile(filepath.Join(dir, "agent-"+agentID+".meta.json"), meta, 0o600); err != nil {
		t.Fatal(err)
	}
}

func snapshotFiles(t *testing.T, roots ...string) map[string]string {
	t.Helper()
	out := map[string]string{}
	for _, root := range roots {
		_ = filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return err
			}
			b, rerr := os.ReadFile(p)
			if rerr != nil {
				return rerr
			}
			out[p] = string(b)
			return nil
		})
	}
	return out
}

// AC-SMA-011 — the doctor sweep: primary, worktree-prefix and listed-worktree
// slugs are swept, an unrelated slug is not, an empty sweep is informational
// rather than ok, nothing is written, and doctor's exit status is unchanged.
func TestServedModelCheck_Sweep(t *testing.T) {
	const primary = "/x/proj"
	const listedWorktree = "/y/lane2"
	glm := []string{servedRow("glm-5.3-flash"), servedRow("glm-5.3-flash")}
	opus := []string{servedRow("claude-opus-5-5")}
	noModel := []string{`{"type":"assistant","message":{}}`}

	project := t.TempDir()
	logPath := filepath.Join(project, ".moai", "logs", "agent-model-audit.jsonl")
	if err := os.MkdirAll(filepath.Dir(logPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(logPath, []byte(`{"agent":"plan-auditor","verdict":"ok"}`+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	exitOf := func(c DiagnosticCheck) error { return doctorExitStatus(countFailedChecks([]DiagnosticCheck{c})) }

	t.Run("a_primary_worktree_and_listed_slugs", func(t *testing.T) {
		configBase, homeBase := t.TempDir(), t.TempDir()
		writeServedFixture(t, configBase, "-x-proj", "s1", "p1", "plan-auditor", glm)
		writeServedFixture(t, configBase, "-x-proj", "s1", "p2", "sync-auditor", opus)
		writeServedFixture(t, homeBase, "-x-proj--claude-worktrees-lane1", "s2", "w1", "sync-auditor", glm)
		writeServedFixture(t, configBase, "-y-lane2", "s3", "l1", "plan-auditor", noModel)
		writeServedFixture(t, configBase, "-x-other", "s4", "o1", "plan-auditor", glm)
		before := snapshotFiles(t, configBase, homeBase, project)

		got := runServedModelScan(servedScanInputs{
			bases:         []string{configBase, homeBase},
			primary:       primary,
			worktreePaths: []string{listedWorktree},
			cfg:           &config.Config{},
		}, false)
		text := got.Message + "\n" + got.Detail

		if got.Status != uikit.CheckWarn {
			t.Fatalf("status = %q, want warn (text=%q)", got.Status, text)
		}
		for _, want := range []string{"swept 4", "ok 1", "served_drift 2", "unknown 1",
			configBase, homeBase, "-x-proj", "-x-proj--claude-worktrees-lane1", "-y-lane2",
			"agent-p1", "agent-w1", "agent-l1"} {
			if !strings.Contains(text, want) {
				t.Fatalf("report lacks %q:\n%s", want, text)
			}
		}
		if strings.Contains(text, "-x-other") || strings.Contains(text, "agent-o1") {
			t.Fatalf("an unrelated slug was swept:\n%s", text)
		}
		if after := snapshotFiles(t, configBase, homeBase, project); len(after) != len(before) {
			t.Fatalf("the sweep wrote files: before=%d after=%d", len(before), len(after))
		} else {
			for p, b := range before {
				if after[p] != b {
					t.Fatalf("the sweep changed %s", p)
				}
			}
		}
		control := DiagnosticCheck{Name: servedModelCheckName, Status: uikit.CheckOK}
		if (exitOf(got) == nil) != (exitOf(control) == nil) {
			t.Fatalf("doctor exit status moved: finding=%v control=%v", exitOf(got), exitOf(control))
		}
	})

	t.Run("b_empty_sweep_is_info_not_ok", func(t *testing.T) {
		configBase, homeBase := t.TempDir(), t.TempDir()
		got := runServedModelScan(servedScanInputs{
			bases:   []string{configBase, homeBase},
			primary: primary,
			cfg:     &config.Config{},
		}, false)
		text := got.Message + "\n" + got.Detail
		if got.Status != uikit.CheckInfo {
			t.Fatalf("status = %q, want info (never ok) for an empty sweep: %q", got.Status, text)
		}
		for _, want := range []string{"swept 0", configBase, homeBase, "-x-proj"} {
			if !strings.Contains(text, want) {
				t.Fatalf("empty-sweep report lacks %q:\n%s", want, text)
			}
		}
		if exitOf(got) != nil {
			t.Fatalf("an empty sweep changed doctor's exit status: %v", exitOf(got))
		}
	})
}
