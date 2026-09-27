package cli

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/closure"
	"github.com/modu-ai/moai-adk/internal/closure/closuretest"
	"github.com/modu-ai/moai-adk/internal/kanban"
)

// ─── seams ───

// fixtureSeams points the A4 command seams at the fixture and restores the
// production values when the test ends. findProjectRootFn is overridden too:
// loadContractEnv resolves configuration from it, and the test process's
// working directory is this repository, never the fixture.
func fixtureSeams(t *testing.T, f *closuretest.Fixture) {
	t.Helper()
	savedQueue, savedRun, savedRoot := contractQueueRootFn, contractRunDirFn, findProjectRootFn
	contractQueueRootFn = func() string { return f.Root }
	contractRunDirFn = func() (string, error) { return f.Root, nil }
	findProjectRootFn = func() (string, error) { return f.Root, nil }
	t.Cleanup(func() {
		contractQueueRootFn, contractRunDirFn, findProjectRootFn = savedQueue, savedRun, savedRoot
	})
}

// queueWithCards installs queue items card→specID ("": no SPEC).
func queueWithCards(t *testing.T, f *closuretest.Fixture, cards map[string]string) {
	t.Helper()
	store := kanban.NewBacklogStore(kanban.BacklogPathForRoot(f.Root))
	err := store.Mutate(func(rec *kanban.BacklogRecord) error {
		for card, spec := range cards {
			item := kanban.BacklogItem{
				ID: card, Text: "fixture card " + card,
				AddedAt: "2026-09-27T00:00:00Z", State: kanban.BacklogStatePicked,
			}
			if spec != "" {
				s := spec
				item.SpecID = &s
			}
			rec.Items = append(rec.Items, item)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("queue setup: %v", err)
	}
}

// runReport invokes the report command with the current seams.
func runReport(t *testing.T, card string) (string, error) {
	t.Helper()
	c := &cobra.Command{}
	out := &bytes.Buffer{}
	c.SetOut(out)
	c.SetErr(out)
	err := runContractReport(c, card)
	return out.String(), err
}

// exitCodeOf extracts the exit-coded error's code, 0 when err is nil.
func exitCodeOf(t *testing.T, err error) int {
	t.Helper()
	if err == nil {
		return 0
	}
	var e *exitCodeError
	if errors.As(err, &e) {
		return e.code
	}
	t.Fatalf("unexpected error kind: %v", err)
	return -1
}

// contractDigest returns the fixture contract's recorded digest.
func contractDigest(t *testing.T, f *closuretest.Fixture) string {
	t.Helper()
	rep, _, _, _, _, err := closure.LoadBuildSideFiles(f.CardDir, closuretest.SpecID,
		closuretest.Autonomy(), nil, nil)
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	return rep.RecordedContractSHA256
}

// writeSecondReview writes one performed second-review record line.
func writeSecondReview(t *testing.T, f *closuretest.Fixture, digest, head, verdict string) {
	t.Helper()
	disagree := false
	line, err := json.Marshal(map[string]any{
		"schema_version": 1, "card": closuretest.Card, "contract_card": closuretest.Card,
		"spec_id": closuretest.SpecID, "contract_sha256": digest, "head_sha": head,
		"target": "baseBranch",
		"scope": map[string]any{
			"base_branch": "main", "base_sha": "base01", "head_sha": head,
			"changed_files": 3, "diff_sha256": "diff01",
		},
		"backends": []map[string]string{
			{"backend": "codex", "gate": "required", "verdict": verdict},
		},
		"participant_count": 1, "disagreement_flag": &disagree,
		"audit_receipt": "rcpt-1", "build_commit": "build01",
		"recorded_at": "2026-09-27T09:00:00Z",
	})
	if err != nil {
		t.Fatalf("marshal record: %v", err)
	}
	f.Write(filepath.Join(f.EvidenceDir, closure.SecondReviewFile), string(line)+"\n")
}

// buildReadyEvidence generates the report pair, records a performed pass
// review at the card HEAD, and regenerates the report.
func buildReadyEvidence(t *testing.T, f *closuretest.Fixture) {
	t.Helper()
	digest := contractDigest(t, f)
	writeSecondReview(t, f, digest, f.Head(), "pass")
	if _, err := runReport(t, closuretest.Card); err != nil {
		t.Fatalf("report after review: %v", err)
	}
}

// gitWorktreeRemove removes the card worktree from the fixture repository.
func gitWorktreeRemove(t *testing.T, f *closuretest.Fixture) {
	t.Helper()
	cmd := exec.Command("git", "worktree", "remove", "--force", f.CardDir)
	cmd.Dir = f.Root
	cmd.Env = closuretest.ScrubbedEnv()
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("worktree remove: %v: %s", err, out)
	}
}

// ─── AC-CLOSURE-001 — report command writes the pair or refuses ───

func TestAC_CLOSURE_001(t *testing.T) {
	f := closuretest.New(t)
	fixtureSeams(t, f)
	queueWithCards(t, f, map[string]string{
		"c1": closuretest.SpecID,
		"c3": "",                 // no SPEC ID
		"c5": closuretest.SpecID, // the contract names c1, not c5
	})
	// c4: a SPEC with no contract.yaml. c4 has no worktree, so its card
	// evidence home is the primary checkout — the SPEC lives there.
	f.Write(filepath.Join(f.Root, ".moai", "specs", "SPEC-FIXTURE-004", "spec.md"),
		"---\nid: SPEC-FIXTURE-004\nstatus: draft\n---\n")
	// c5: mapped to the SAME SPEC as c1 — the signed contract names c1, so
	// the report must refuse. The SPEC copy under the primary makes it
	// readable from c5's evidence home (the primary checkout).
	if err := copyDir(filepath.Join(f.CardDir, ".moai", "specs", closuretest.SpecID),
		filepath.Join(f.Root, ".moai", "specs", closuretest.SpecID)); err != nil {
		t.Fatalf("copy spec for c5: %v", err)
	}
	queueWithCards(t, f, map[string]string{"c4": "SPEC-FIXTURE-004"})

	out, err := runReport(t, "c1")
	if exitCodeOf(t, err) != 0 {
		t.Fatalf("report c1: exit %d: %v", exitCodeOf(t, err), err)
	}
	wantPath := filepath.Join(f.CardDir, ".moai", "reports", "c1", "closure-report.md")
	if !strings.Contains(out, wantPath) {
		t.Fatalf("output = %q, want the md path %q", out, wantPath)
	}
	for _, name := range []string{closure.ReportMDFile, closure.ReportJSONFile} {
		if _, statErr := os.Stat(filepath.Join(f.EvidenceDir, name)); statErr != nil {
			t.Fatalf("%s missing: %v", name, statErr)
		}
	}

	for _, tc := range []struct {
		card  string
		cause string
	}{
		{"c2", "not in the queue"},
		{"c3", "no SPEC ID"},
		{"c4", "contract.yaml"},
		{"c5", "names card"},
	} {
		_, verr := runReport(t, tc.card)
		if code := exitCodeOf(t, verr); code != 2 {
			t.Fatalf("report %s: exit %d err %v, want 2", tc.card, code, verr)
		}
		if !strings.Contains(fmt.Sprint(verr), tc.cause) {
			t.Fatalf("report %s: err %v does not name the cause %q", tc.card, verr, tc.cause)
		}
	}
	// No file was created under any .moai/reports for the refusals.
	for _, rel := range []string{"c2", "c3", "c4", "c5"} {
		for _, base := range []string{f.CardDir, f.Root} {
			dir := filepath.Join(base, ".moai", "reports", rel)
			if _, statErr := os.Stat(dir); !errors.Is(statErr, os.ErrNotExist) {
				t.Fatalf("refusal created %s: %v", dir, statErr)
			}
		}
	}
}

// ─── AC-CLOSURE-019 — push-check CLI ───

func runPushCheck(t *testing.T, f *closuretest.Fixture, mode string, args ...string) (string, int, error) {
	t.Helper()
	if mode != "" {
		f.Write(filepath.Join(f.Root, ".moai", "config", "sections", "workflow.yaml"),
			"workflow:\n  autonomy:\n    mode: "+mode+"\n    contract:\n      second_review: required\n")
		f.Write(filepath.Join(f.Root, ".moai", "config", "sections", "git-strategy.yaml"),
			"git_strategy:\n  mode: manual\n  manual:\n    workflow: git-flow\n    develop_branch: develop\n")
	}
	fixtureSeams(t, f)
	c := &cobra.Command{}
	out := &bytes.Buffer{}
	c.SetOut(out)
	c.SetErr(out)
	err := runContractPushCheck(c, args)
	return out.String(), exitCodeOf(t, err), err
}

func TestAC_CLOSURE_019(t *testing.T) {
	t.Run("ready", func(t *testing.T) {
		f := closuretest.New(t)
		fixtureSeams(t, f)
		queueWithCards(t, f, map[string]string{"c1": closuretest.SpecID})
		buildReadyEvidence(t, f)
		out, code, err := runPushCheck(t, f, "contract")
		if code != 0 {
			t.Fatalf("exit %d err %v out %s, want ready", code, err, out)
		}
		if !strings.Contains(out, "ready") {
			t.Fatalf("out = %q", out)
		}
	})

	t.Run("not ready names the SPEC and codes", func(t *testing.T) {
		f := closuretest.New(t)
		fixtureSeams(t, f)
		queueWithCards(t, f, map[string]string{"c1": closuretest.SpecID})
		if _, err := runReport(t, closuretest.Card); err != nil {
			t.Fatalf("report: %v", err)
		}
		// Merge the card branch into local develop so the pushed source
		// commit carries the SPEC (a sync-commit-like landing).
		f.Git(f.Root, "merge", "--no-ff", closuretest.Branch)
		out, code, err := runPushCheck(t, f, "contract")
		if code != 1 {
			t.Fatalf("exit %d err %v, want 1", code, err)
		}
		if !strings.Contains(out, closuretest.SpecID) || !strings.Contains(out, "second_review_not_performed") {
			t.Fatalf("out = %q, want SPEC with codes", out)
		}
	})

	t.Run("refspec from another tree", func(t *testing.T) {
		f := closuretest.New(t)
		fixtureSeams(t, f)
		queueWithCards(t, f, map[string]string{"c1": closuretest.SpecID})
		buildReadyEvidence(t, f)
		out, code, err := runPushCheck(t, f, "contract", "origin", "HEAD:develop")
		if code != 0 || !strings.Contains(out, "ready") {
			t.Fatalf("exit %d out %s err %v, want ready", code, out, err)
		}
	})

	t.Run("all flag is undetermined", func(t *testing.T) {
		f := closuretest.New(t)
		out, code, err := runPushCheck(t, f, "contract", "--all", "origin")
		if code != 1 || !strings.Contains(out, "push_check_undetermined") {
			t.Fatalf("exit %d out %s err %v, want undetermined", code, out, err)
		}
	})

	t.Run("missing remote ref is undetermined", func(t *testing.T) {
		f := closuretest.New(t)
		out, code, err := runPushCheck(t, f, "contract", "nowhere", "HEAD:develop")
		if code != 1 || !strings.Contains(out, "push_check_undetermined") {
			t.Fatalf("exit %d out %s err %v, want undetermined", code, out, err)
		}
	})

	t.Run("unknown flag exits 2", func(t *testing.T) {
		f := closuretest.New(t)
		_, code, err := runPushCheck(t, f, "contract", "--bogus")
		if code != 2 {
			t.Fatalf("exit %d err %v, want 2", code, err)
		}
	})

	t.Run("guided is inactive", func(t *testing.T) {
		f := closuretest.New(t)
		out, code, err := runPushCheck(t, f, "guided")
		if code != 0 || !strings.Contains(out, "inactive") {
			t.Fatalf("exit %d out %s err %v, want inactive", code, out, err)
		}
	})
}

// ─── AC-CLOSURE-020 — human verdict recording ───

type verdictSeam struct {
	env      map[string]string
	tty      bool
	lines    []string
	name     string
	email    string
	hasIdent bool
}

func withVerdictSeams(t *testing.T, s verdictSeam, f *closuretest.Fixture) {
	t.Helper()
	savedEnv, savedTTY, savedReader, savedIdent := contractGetenvFn, contractStdinIsTerminalFn, newContractLineReader, contractGitIdentityFn
	contractGetenvFn = func(key string) string { return s.env[key] }
	contractStdinIsTerminalFn = func() bool { return s.tty }
	newContractLineReader = func(r io.Reader) func() (string, error) {
		return func() (string, error) {
			if len(s.lines) == 0 {
				return "", fmt.Errorf("eof")
			}
			line := s.lines[0]
			s.lines = s.lines[1:]
			return line, nil
		}
	}
	contractGitIdentityFn = func(root string) (string, string, error) {
		if !s.hasIdent {
			return "", "", fmt.Errorf("no identity")
		}
		return s.name, s.email, nil
	}
	fixtureSeams(t, f)
	t.Cleanup(func() {
		contractGetenvFn, contractStdinIsTerminalFn, newContractLineReader, contractGitIdentityFn =
			savedEnv, savedTTY, savedReader, savedIdent
	})
}

func runVerdict(t *testing.T, args ...string) (string, int, error) {
	t.Helper()
	c := &cobra.Command{}
	out := &bytes.Buffer{}
	c.SetOut(out)
	c.SetErr(out)
	err := runContractVerdict(c, args[0], args[1], "")
	return out.String(), exitCodeOf(t, err), err
}

func TestAC_CLOSURE_020(t *testing.T) {
	refusals := []struct {
		name string
		seam verdictSeam
	}{
		{"agent marker", verdictSeam{env: map[string]string{"CLAUDECODE": "1"}, tty: true,
			lines: []string{"accept c1"}, hasIdent: true}},
		{"not a terminal", verdictSeam{tty: false, lines: []string{"accept c1"}, hasIdent: true}},
		{"wrong confirmation", verdictSeam{tty: true, lines: []string{"reject c1"}, hasIdent: true}},
	}
	for _, tc := range refusals {
		t.Run(tc.name+" refuses", func(t *testing.T) {
			f := closuretest.New(t)
			fixtureSeams(t, f)
			queueWithCards(t, f, map[string]string{"c1": closuretest.SpecID})
			if _, err := runReport(t, closuretest.Card); err != nil {
				t.Fatalf("report: %v", err)
			}
			withVerdictSeams(t, tc.seam, f)
			_, code, err := runVerdict(t, closuretest.Card, "accept")
			if code != 1 {
				t.Fatalf("exit %d err %v, want 1", code, err)
			}
			if _, statErr := os.Stat(filepath.Join(f.EvidenceDir, closure.ClosureVerdictFile)); !errors.Is(statErr, os.ErrNotExist) {
				t.Fatalf("verdict file exists after refusal: %v", statErr)
			}
		})
	}

	t.Run("no closure report refuses", func(t *testing.T) {
		f := closuretest.New(t)
		queueWithCards(t, f, map[string]string{"c1": closuretest.SpecID})
		withVerdictSeams(t, verdictSeam{tty: true, lines: []string{"accept c1"}, hasIdent: true}, f)
		_, code, err := runVerdict(t, closuretest.Card, "accept")
		if code != 1 {
			t.Fatalf("exit %d err %v, want 1", code, err)
		}
	})

	t.Run("records on the terminal path", func(t *testing.T) {
		f := closuretest.New(t)
		fixtureSeams(t, f)
		queueWithCards(t, f, map[string]string{"c1": closuretest.SpecID})
		if _, err := runReport(t, closuretest.Card); err != nil {
			t.Fatalf("report: %v", err)
		}
		withVerdictSeams(t, verdictSeam{tty: true, lines: []string{"accept c1"},
			name: "GOOS", email: "goos@example.com", hasIdent: true}, f)
		out, code, err := runVerdict(t, closuretest.Card, "accept")
		if code != 0 {
			t.Fatalf("exit %d err %v out %s, want 0", code, err, out)
		}
		data, err := os.ReadFile(filepath.Join(f.EvidenceDir, closure.ClosureVerdictFile))
		if err != nil {
			t.Fatalf("verdict file: %v", err)
		}
		var rec closure.VerdictRecord
		if err := json.Unmarshal(data, &rec); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if rec.Verdict != "accept" || rec.Operator.Name != "GOOS" || rec.Operator.Email != "goos@example.com" ||
			rec.Method != "interactive-tty" {
			t.Fatalf("record = %+v", rec)
		}
		// report_sha256 equals the canonical hash of the current report.
		reportData, err := os.ReadFile(filepath.Join(f.EvidenceDir, closure.ReportJSONFile))
		if err != nil {
			t.Fatalf("report: %v", err)
		}
		var parsed closure.Report
		if err := json.Unmarshal(reportData, &parsed); err != nil {
			t.Fatalf("parse report: %v", err)
		}
		if rec.ReportSHA256 != closure.CanonicalReportHash(&parsed) {
			t.Fatalf("report_sha256 = %q, want the canonical hash", rec.ReportSHA256)
		}
	})
}

// ─── AC-CLOSURE-024 — one evidence home for writers and readers ───

func TestAC_CLOSURE_024(t *testing.T) {
	f := closuretest.New(t)
	fixtureSeams(t, f)
	queueWithCards(t, f, map[string]string{"c1": closuretest.SpecID})

	// From the primary checkout and from a worktree of the same repository
	// (here: twice through the same seam), the report pair lands only in the
	// card evidence directory.
	for i := 0; i < 2; i++ {
		if _, err := runReport(t, closuretest.Card); err != nil {
			t.Fatalf("report run %d: %v", i+1, err)
		}
	}
	for _, name := range []string{closure.ReportMDFile, closure.ReportJSONFile} {
		if _, err := os.Stat(filepath.Join(f.EvidenceDir, name)); err != nil {
			t.Fatalf("%s missing from the card evidence dir: %v", name, err)
		}
	}
	// sources names the read paths (the card evidence home).
	data, err := os.ReadFile(filepath.Join(f.EvidenceDir, closure.ReportJSONFile))
	if err != nil {
		t.Fatalf("read report: %v", err)
	}
	var parsed closure.Report
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("parse: %v", err)
	}
	if !strings.Contains(parsed.Sources["progress"], f.CardDir) {
		t.Fatalf("sources[progress] = %q, want the card evidence home path", parsed.Sources["progress"])
	}

	// With the c1 worktree removed, the report writes into the primary
	// checkout's evidence directory. The SPEC (with its signed contract) is
	// copied to the primary first, so the contract still resolves there.
	specRel := filepath.Join(".moai", "specs", closuretest.SpecID)
	if err := copyDir(filepath.Join(f.CardDir, specRel), filepath.Join(f.Root, specRel)); err != nil {
		t.Fatalf("copy spec: %v", err)
	}
	f.Git(f.Root, "add", ".moai")
	f.Git(f.Root, "commit", "-m", "spec on primary")
	gitWorktreeRemove(t, f)

	out, err := runReport(t, closuretest.Card)
	if code := exitCodeOf(t, err); code != 0 {
		t.Fatalf("report after worktree removal: exit %d err %v", code, err)
	}
	wantPath := filepath.Join(f.Root, ".moai", "reports", "c1", "closure-report.md")
	if !strings.Contains(out, wantPath) {
		t.Fatalf("out = %q, want the primary checkout path %q", out, wantPath)
	}
}

// copyDir copies a directory tree (test helper for the removed-worktree case).
func copyDir(src, dst string) error {
	return filepath.WalkDir(src, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, 0o644)
	})
}
