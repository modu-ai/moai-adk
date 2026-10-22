package worktree

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/worktreeroot"
	"github.com/spf13/cobra"
)

func TestReviewVerifyRejectsInvalidJournals(t *testing.T) {
	for _, raw := range []string{`{broken`, `{"steps":"wrong"}`, `{"unknown_schema":true}`} {
		t.Run(raw, func(t *testing.T) {
			root := t.TempDir()
			dir := filepath.Join(root, ".moai", "state", "worktree-migrate")
			if err := os.MkdirAll(dir, 0755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(dir, "corrupt.json"), []byte(raw), 0600); err != nil {
				t.Fatal(err)
			}
			cmd := &cobra.Command{}
			var out bytes.Buffer
			cmd.SetOut(&out)
			if err := runMigrateVerify(cmd, root); err == nil {
				t.Fatalf("invalid journal verified as intact: %s", out.String())
			}
		})
	}
	t.Run("unreadable-entry", func(t *testing.T) {
		root := t.TempDir()
		if err := os.MkdirAll(filepath.Join(root, ".moai", "state", "worktree-migrate", "unreadable.json"), 0o755); err != nil {
			t.Fatal(err)
		}
		cmd := &cobra.Command{}
		var out bytes.Buffer
		cmd.SetOut(&out)
		if err := runMigrateVerify(cmd, root); err == nil {
			t.Fatalf("unreadable journal reported verified: %s", out.String())
		}
	})
}

func TestReviewJournalNamesDoNotOverwriteEarlierRuns(t *testing.T) {
	repo, tree, _ := rfxLegacyWorld(t)
	m4InRepo(t, repo)
	reviewMigrationStateDir(t, repo)
	if _, err := runMigrationEngine(repo, true, "review-first"); err != nil {
		t.Fatal(err)
	}
	first := sfxLatestJournal(t, repo)
	before, err := os.ReadFile(first)
	if err != nil {
		t.Fatal(err)
	}
	// Preserve real earlier rollback evidence at the timestamp names a new
	// run could select. A bounded window spans clock ticks during Git probes,
	// so this does not depend on two subprocess-heavy applies beating a timer.
	dir := filepath.Dir(first)
	var oldNames []string
	start := time.Now().UTC()
	for i := -2; i <= 60; i++ {
		oldName := filepath.Join(dir, start.Add(time.Duration(i)*time.Second).Format("20060102T150405")+".json")
		if err := os.WriteFile(oldName, before, 0o600); err != nil {
			t.Fatal(err)
		}
		oldNames = append(oldNames, oldName)
	}
	second := filepath.Join(repo, ".moai", "worktrees", "review-second")
	if err := os.MkdirAll(filepath.Dir(second), 0755); err != nil {
		t.Fatal(err)
	}
	m6Git(t, repo, "worktree", "add", "-q", "-b", "WT-review-second", "--", second)
	if _, err := runMigrationEngine(repo, true, "review-second"); err != nil {
		t.Fatal(err)
	}
	for _, oldName := range oldNames {
		after, err := os.ReadFile(oldName)
		if err != nil || !bytes.Equal(before, after) {
			t.Fatalf("later apply overwrote earlier journal for %s at %s: %v", tree, oldName, err)
		}
	}
}

func TestReviewRollbackRefusesMissingRecordedInstance(t *testing.T) {
	repo, _, _ := rfxLegacyWorld(t)
	for _, replacement := range []bool{false, true} {
		t.Run(map[bool]string{false: "missing", true: "different-source"}[replacement], func(t *testing.T) {
			old := filepath.Join(t.TempDir(), "old")
			dest := filepath.Join(t.TempDir(), "dest")
			if replacement {
				old = repo
			}
			dir := filepath.Join(repo, ".moai", "state", "worktree-migrate")
			if err := os.MkdirAll(dir, 0755); err != nil {
				t.Fatal(err)
			}
			body, err := json.Marshal(map[string]any{"steps": []journalStep{{OldPath: old, NewPath: dest, Ident: "different-tree:token", Outcome: "moved"}}})
			if err != nil {
				t.Fatal(err)
			}
			journal := filepath.Join(dir, "missing.json")
			if err := os.WriteFile(journal, body, 0600); err != nil {
				t.Fatal(err)
			}
			cmd := &cobra.Command{}
			var out bytes.Buffer
			cmd.SetOut(&out)
			if err := runMigrateRollback(cmd, repo, journal, false); err == nil {
				t.Fatalf("missing/different instance rollback succeeded: %s", out.String())
			}
			if !strings.Contains(out.String(), "refused") {
				t.Fatalf("missing instance needs refusal report: %s", out.String())
			}
		})
	}
}

func TestReviewIdentityUnavailableSkipIsAudited(t *testing.T) {
	repo, tree, _ := rfxLegacyWorld(t)
	reviewMigrationStateDir(t, repo)
	admin, err := gitDirPointer(tree)
	if err != nil {
		t.Fatal(err)
	}
	// A directory at the identity-token file refuses token creation on all
	// hosts, including privileged Linux, without chmod/capability assumptions.
	if err := os.Mkdir(filepath.Join(admin, tokenFileName), 0755); err != nil {
		t.Fatal(err)
	}
	ml := moveListPath(repo)
	out := migrateOneTree(repo, migratePlan{source: tree, destination: filepath.Join(t.TempDir(), "dest")}, ml, "review", nil)
	if out.Outcome != "skipped" {
		t.Fatalf("outcome=%+v", out)
	}
	raw, err := os.ReadFile(ml)
	if err != nil || !strings.Contains(string(raw), `"result":"skipped"`) || !strings.Contains(string(raw), "identity unavailable") {
		t.Fatalf("skip audit missing: %v, %s", err, raw)
	}
}

func TestReviewPartialMigrationReportSurvivesEngineError(t *testing.T) {
	repo, tree, _ := rfxLegacyWorld(t)
	m4InRepo(t, repo)
	reviewMigrationStateDir(t, repo)
	original := journalWrite
	journalWrite = func(string, []byte) error { return fmt.Errorf("review journal failure") }
	t.Cleanup(func() { journalWrite = original })
	cmd := &cobra.Command{}
	var out bytes.Buffer
	cmd.SetOut(&out)
	if err := runMigratePlanOrApply(cmd, repo, true, false); err == nil {
		t.Fatal("journal failure became successful apply")
	}
	if !strings.Contains(out.String(), tree) || !strings.Contains(out.String(), "failed") {
		t.Fatalf("engine error erased partial report: %s", out.String())
	}
}

func TestReviewUpdatePartialReportSurvivesEngineError(t *testing.T) {
	repo, tree, _ := rfxLegacyWorld(t)
	m4InRepo(t, repo)
	reviewMigrationStateDir(t, repo)
	original := journalWrite
	journalWrite = func(string, []byte) error { return fmt.Errorf("review update journal failure") }
	t.Cleanup(func() { journalWrite = original })
	var out bytes.Buffer
	if err := runUpdateConsolidation(repo, false, &out); err == nil {
		t.Fatal("journal failure became successful update consolidation")
	}
	if !strings.Contains(out.String(), tree) || !strings.Contains(out.String(), "failed") {
		t.Fatalf("engine error erased update partial report: %s", out.String())
	}
	if _, err := os.Stat(tree); err != nil {
		t.Fatalf("refused move lost its source: %v", err)
	}
}

func TestReviewMigrationRevalidatesBeforeEachMove(t *testing.T) {
	repo, first, _ := rfxLegacyWorld(t)
	m4InRepo(t, repo)
	reviewMigrationStateDir(t, repo)
	second := filepath.Join(filepath.Dir(first), "z-second")
	m6Git(t, repo, "worktree", "add", "-q", "-b", "WT-review-second", "--", second)
	originalProbe := activeProcessCWDProbe
	originalWrite := journalWrite
	becameActive := false
	activeProcessCWDProbe = func() ([]string, error) {
		if becameActive {
			return []string{second}, nil
		}
		return nil, nil
	}
	journalWrite = func(path string, data []byte) error {
		err := originalWrite(path, data)
		if bytes.Contains(data, []byte(`"outcome": "moved"`)) {
			becameActive = true
		}
		return err
	}
	t.Cleanup(func() { activeProcessCWDProbe = originalProbe; journalWrite = originalWrite })
	report, err := runMigrationEngine(repo, true, "review-revalidate")
	if err != nil {
		t.Fatal(err)
	}
	if !becameActive {
		t.Fatal("fixture did not finish first move")
	}
	if !m6Registered(t, repo, second) {
		t.Fatalf("later live tree moved using stale initial plan: %+v", report.rows)
	}
	found := false
	for _, row := range report.rows {
		if row.Source == second && row.Outcome == "skipped" && strings.Contains(row.SkipReason, "active process") {
			found = true
		}
	}
	if !found {
		t.Fatalf("newly active tree missing skipped report: %+v", report.rows)
	}
}

func reviewMigrationStateDir(t *testing.T, repo string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(moveListPath(repo)), 0o755); err != nil {
		t.Fatal(err)
	}
}

func TestReviewL1GuardUsesHomeDirectoryIdentity(t *testing.T) {
	repo, _, base := rfxLegacyWorld(t)
	home, alias := filepath.Join(base, "CaseHome"), filepath.Join(base, "casehome")
	if err := os.MkdirAll(home, 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("MOAI_HOME", home)
	project, err := worktreeroot.Root(repo)
	if err != nil {
		t.Fatal(err)
	}
	tree := filepath.Join(project, "release-review")
	m6Git(t, repo, "worktree", "add", "-q", "-b", "release-review", "--", tree)
	if err := os.MkdirAll(alias, 0755); err != nil {
		t.Fatal(err)
	}
	originalInfo, err := os.Stat(home)
	if err != nil {
		t.Fatal(err)
	}
	aliasInfo, err := os.Stat(alias)
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("MOAI_HOME", alias)
	want := os.SameFile(originalInfo, aliasInfo)
	if got := isL1WorktreePath(tree); got != want {
		t.Fatalf("L1=%v with same home identity=%v (%s/%s)", got, want, home, alias)
	}
}
