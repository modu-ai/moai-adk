package constitution

import (
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
	"time"
)

// SPEC-CON-AMEND-APPLY-001 M1 — evolution-log writer and reader acceptance
// tests (AC-CAA-006, 007, 009, 010, 011, 018). Every fixture lives under
// t.TempDir(); the only real file touched is .moai/research/evolution-log.md,
// opened read-only by AC-CAA-010 (b).

// humanFormatSection is the "## HISTORY" .. end-of-file span of the real
// .moai/research/evolution-log.md, copied verbatim (AC-CAA-010 (a)). The
// three-backtick fence is written as ~~~ here because a Go raw string cannot
// hold a backtick; realFence restores it.
const humanFormatSection = `## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-05-13 | SPEC-V3R2-HRN-002 (manager-spec) | Initial file creation — EVO-HRN-002 first entry. |

---

## EVO-HRN-002

~~~yaml
id: EVO-HRN-002
spec_id: SPEC-V3R2-HRN-002
timestamp: "2026-05-13"
zone: Frozen
target_file: .claude/rules/moai/design/constitution.md
target_section: "§11 GAN Loop Contract"
amendment_type: "sub-section insertion"

before_snippet: |
  # (§11.4.1 did not exist)
  # §11.4 Sprint Contract Protocol defined durable Sprint Contract state,
  # but did not constrain evaluator memory scope to ephemeral-per-iteration.

after_snippet: |
  ### §11.4.1 Evaluator Memory Scope (Principle 4)
  Evaluator judgment memory SHALL be ephemeral per iteration...
  Sprint Contract state SHALL be durable across iterations...
  Implementation: the GAN loop runner SHALL respawn evaluator-active for each iteration...
  Configuration: evaluator.memory_scope: per_iteration (FROZEN; value cannot be changed without a new CON-002 amendment cycle)

canary_verdict: "CanaryUnavailable (0 of 3 required subjects; v3 corpus has no completed GAN-loop evaluation artifacts)"
canary_log_ref: ".moai/specs/SPEC-V3R2-HRN-002/canary-fresh-memory-eval.txt"

con_002_evidence_ref: ".moai/specs/SPEC-V3R2-HRN-002/con-002-amendment-evidence.md"
con_002_layers:
  layer1_frozen_guard: "PASS"
  layer2_canary: "PASS (CanaryUnavailable — no prior evaluator runs to regress)"
  layer3_contradiction: "PASS"
  layer4_rate_limiter: "PASS (2 of 3 used in v3.x cycle)"
  layer5_human_oversight: "PENDING-FINAL (maintainer approval in landing PR)"

approver: "Goos Kim (bobby@afamily.kr) — approval recorded in landing PR description"
rationale_cite: "R1 §9 (Zhuge et al. 2024 arXiv:2410.10934 Agent-as-a-Judge anti-pattern: cumulative evaluator memory) + Principle 4 (design-constitution §11.4.1)"

const_registry_entry: "CONST-V3R2-153"
design_log_ref: ".moai/design/v3-research/evolution-log.md"

status: "landed"
version_before: "3.4.0"
version_after: "3.5.0"
~~~

---

_End of evolution log._
`

// realFence restores the markdown code fences of a fixture written with ~~~.
func realFence(s string) string { return strings.ReplaceAll(s, "~~~", "`"+"``") }

// writeLog writes content to <dir>/evolution-log.md and returns the path.
func writeLog(t *testing.T, dir, content string) string {
	t.Helper()
	p := filepath.Join(dir, "evolution-log.md")
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatalf("write log fixture: %v", err)
	}
	return p
}

// stripFixturePaths removes every given path from s, in its cleaned absolute
// form and its symbolic-link-resolved form (longest first), so that the digits
// a temporary directory carries cannot satisfy a number assertion
// (acceptance.md common rules, numeric-assertion rule).
func stripFixturePaths(s string, paths ...string) string {
	var forms []string
	for _, p := range paths {
		if abs, err := filepath.Abs(p); err == nil {
			forms = append(forms, filepath.Clean(abs))
		}
		if res, err := filepath.EvalSymlinks(p); err == nil {
			forms = append(forms, res)
		}
		forms = append(forms, p)
	}
	// Longest first, so a resolved /private/var/... form is removed before
	// the /var/... form it contains.
	sort.Slice(forms, func(i, j int) bool { return len(forms[i]) > len(forms[j]) })
	for _, f := range forms {
		if f != "" {
			s = strings.ReplaceAll(s, f, "")
		}
	}
	return s
}

// hasWholeNumber reports whether n occurs in s as a whole number — not
// preceded or followed by another digit.
func hasWholeNumber(s, n string) bool {
	return regexp.MustCompile(`(^|[^0-9])` + regexp.QuoteMeta(n) + `([^0-9]|$)`).MatchString(s)
}

// containsPathForm reports whether s contains p in its cleaned absolute form
// or its symbolic-link-resolved form.
func containsPathForm(s, p string) bool {
	if abs, err := filepath.Abs(p); err == nil && strings.Contains(s, filepath.Clean(abs)) {
		return true
	}
	if res, err := filepath.EvalSymlinks(p); err == nil && strings.Contains(s, res) {
		return true
	}
	return false
}

// AC-CAA-006 — the writer emits snake_case keys and zone names.
func TestAppendEvolutionLog_SnakeCaseAndZoneNames(t *testing.T) {
	integerZone := regexp.MustCompile(`(?m)^\s*zone_?(before|after)\s*:\s*-?[0-9]+\s*$`)
	cases := []struct {
		zone Zone
		name string
	}{
		{ZoneEvolvable, "Evolvable"},
		{ZoneFrozen, "Frozen"},
	}
	for _, tc := range cases {
		logPath := filepath.Join(t.TempDir(), "evolution-log.md")
		entry := &AmendmentLog{
			ID:            "LEARN-20260911-001",
			RuleID:        "CONST-V3R2-003",
			ZoneBefore:    tc.zone,
			ZoneAfter:     tc.zone,
			ClauseBefore:  "old clause",
			ClauseAfter:   "new clause",
			CanaryVerdict: "skipped",
			ApprovedBy:    "human",
			ApprovedAt:    time.Date(2026, 9, 11, 0, 0, 0, 0, time.UTC),
		}
		if err := AppendEvolutionLog(logPath, entry); err != nil {
			t.Fatalf("AppendEvolutionLog(%s): %v", tc.name, err)
		}
		raw, err := os.ReadFile(logPath)
		if err != nil {
			t.Fatal(err)
		}
		got := string(raw)
		for _, want := range []string{"rule_id:", "approved_at:", "zone_before: " + tc.name, "zone_after: " + tc.name} {
			if !strings.Contains(got, want) {
				t.Errorf("zone %s: written log lacks %q:\n%s", tc.name, want, got)
			}
		}
		for _, bad := range []string{"ruleid:", "approvedat:", "zonebefore:"} {
			if strings.Contains(got, bad) {
				t.Errorf("zone %s: written log carries legacy key %q:\n%s", tc.name, bad, got)
			}
		}
		if integerZone.MatchString(got) {
			t.Errorf("zone %s: written log carries an integer zone line:\n%s", tc.name, got)
		}
	}
}

// AC-CAA-007 — legacy concatenated keys and integer zones still read; the
// snake_case key wins when both forms are present.
func TestLoadEvolutionLogs_LegacyKeys(t *testing.T) {
	content := "# Evolution Log\n\n" +
		"---\n" +
		"id: LEARN-20260911-001\n" +
		"ruleid: CONST-V3R2-003\n" +
		"zonebefore: 1\n" +
		"zoneafter: 1\n" +
		"clausebefore: old clause\n" +
		"clauseafter: new clause\n" +
		"canaryverdict: skipped\n" +
		"contradictions: []\n" +
		"approvedby: human\n" +
		"approvedat: 2026-09-11T00:00:00Z\n" +
		"rolledback: false\n" +
		"rollbackreason: \"\"\n" +
		"rollbackat: null\n" +
		"---\n" +
		"---\n" +
		"id: LEARN-20260911-002\n" +
		"rule_id: A\n" +
		"ruleid: B\n" +
		"approved_at: 2026-09-11T01:00:00Z\n" +
		"---\n"
	logs, err := LoadEvolutionLogs(writeLog(t, t.TempDir(), content))
	if err != nil {
		t.Fatalf("LoadEvolutionLogs: %v", err)
	}
	if len(logs) != 2 {
		t.Fatalf("entries = %d, want 2: %+v", len(logs), logs)
	}
	if logs[0].RuleID != "CONST-V3R2-003" {
		t.Errorf("legacy RuleID = %q, want CONST-V3R2-003", logs[0].RuleID)
	}
	if logs[0].ZoneBefore != ZoneEvolvable || logs[0].ZoneAfter != ZoneEvolvable {
		t.Errorf("legacy zones = %v/%v, want Evolvable/Evolvable", logs[0].ZoneBefore, logs[0].ZoneAfter)
	}
	if want := time.Date(2026, 9, 11, 0, 0, 0, 0, time.UTC); !logs[0].ApprovedAt.Equal(want) {
		t.Errorf("legacy ApprovedAt = %v, want %v", logs[0].ApprovedAt, want)
	}
	if logs[1].RuleID != "A" {
		t.Errorf("both-forms RuleID = %q, want A (snake_case wins)", logs[1].RuleID)
	}
}

// markdownNoiseLog builds the AC-CAA-009 fixture; extraRule inserts a third
// horizontal rule before the first entry.
func markdownNoiseLog(extraRule bool) string {
	var b strings.Builder
	b.WriteString("# Evolution Log\n\n")
	b.WriteString("| Version | Date | Description |\n")
	b.WriteString("|---------|------|-------------|\n")
	b.WriteString("| 0.1.0 | 2026-09-01 | first |\n\n")
	b.WriteString("---\n\n")
	b.WriteString("Prose between two horizontal rules.\n\n")
	b.WriteString("---\n\n")
	if extraRule {
		b.WriteString("---\n\n")
	}
	b.WriteString("---\n")
	b.WriteString("id: LEARN-20260901-001\n")
	b.WriteString("rule_id: CONST-V3R2-010\n")
	b.WriteString("approved_by: human\n")
	b.WriteString("approved_at: 2026-09-01T10:00:00Z\n")
	b.WriteString("---\n")
	b.WriteString("---\n")
	b.WriteString("id: LEARN-20260901-002\n")
	b.WriteString("rule_id: CONST-V3R2-011\n")
	b.WriteString("approved_by: human\n")
	b.WriteString("approved_at: 2026-09-01T11:00:00Z\n")
	b.WriteString("---\n")
	return b.String()
}

// AC-CAA-009 — markdown rules and table separators do not disturb parsing,
// and the count does not depend on how many rules precede the entries.
func TestLoadEvolutionLogs_MarkdownNoise(t *testing.T) {
	for _, extra := range []bool{false, true} {
		logs, err := LoadEvolutionLogs(writeLog(t, t.TempDir(), markdownNoiseLog(extra)))
		if err != nil {
			t.Fatalf("extraRule=%v: LoadEvolutionLogs: %v", extra, err)
		}
		if len(logs) != 2 {
			t.Fatalf("extraRule=%v: entries = %d, want 2: %+v", extra, len(logs), logs)
		}
		wantIDs := []string{"LEARN-20260901-001", "LEARN-20260901-002"}
		wantRules := []string{"CONST-V3R2-010", "CONST-V3R2-011"}
		for i := range logs {
			if logs[i].ID != wantIDs[i] || logs[i].RuleID != wantRules[i] {
				t.Errorf("extraRule=%v: entry %d = (%q, %q), want (%q, %q)",
					extra, i, logs[i].ID, logs[i].RuleID, wantIDs[i], wantRules[i])
			}
		}
	}
}

// AC-CAA-010 — human-format entries map as specified; malformed ones fail closed.
func TestLoadEvolutionLogs_HumanFormat(t *testing.T) {
	t.Run("verbatim_block", func(t *testing.T) {
		logs, err := LoadEvolutionLogs(writeLog(t, t.TempDir(), realFence(humanFormatSection)))
		if err != nil {
			t.Fatalf("LoadEvolutionLogs: %v", err)
		}
		if len(logs) != 1 {
			t.Fatalf("entries = %d, want 1: %+v", len(logs), logs)
		}
		got := logs[0]
		if got.ID != "EVO-HRN-002" {
			t.Errorf("ID = %q, want EVO-HRN-002", got.ID)
		}
		if got.RuleID != "CONST-V3R2-153" {
			t.Errorf("RuleID = %q, want CONST-V3R2-153", got.RuleID)
		}
		if want := time.Date(2026, 5, 13, 0, 0, 0, 0, time.UTC); !got.ApprovedAt.Equal(want) {
			t.Errorf("ApprovedAt = %v, want %v", got.ApprovedAt, want)
		}
		if got.ZoneBefore != ZoneFrozen || got.ZoneAfter != ZoneFrozen {
			t.Errorf("zones = %v/%v, want Frozen/Frozen", got.ZoneBefore, got.ZoneAfter)
		}
		if got.RolledBack {
			t.Error("RolledBack = true, want false")
		}
		if got.RollbackAt != nil {
			t.Errorf("RollbackAt = %v, want nil", got.RollbackAt)
		}
	})

	t.Run("real_file_readonly", func(t *testing.T) {
		// Read-only drift witness on the real log: LoadEvolutionLogs only reads.
		realLog := filepath.Join("..", "..", ".moai", "research", "evolution-log.md")
		logs, err := LoadEvolutionLogs(realLog)
		if err != nil {
			t.Fatalf("LoadEvolutionLogs(real): %v", err)
		}
		for _, l := range logs {
			if l.ID == "EVO-HRN-002" {
				return
			}
		}
		t.Errorf("real log: no entry with ID EVO-HRN-002 among %d entries", len(logs))
	})

	t.Run("malformed_timestamp_fails_closed", func(t *testing.T) {
		content := "# Evolution Log\n\n## EVO-X-001\n\n~~~yaml\nid: EVO-X-001\nzone: Frozen\ntimestamp: \"not-a-date\"\n~~~\n"
		_, err := LoadEvolutionLogs(writeLog(t, t.TempDir(), realFence(content)))
		if err == nil {
			t.Fatal("LoadEvolutionLogs: want an error for an unparseable timestamp, got nil")
		}
		if !strings.Contains(err.Error(), "EVO-X-001") {
			t.Errorf("error %q does not name the entry EVO-X-001", err)
		}
	})
}

// AC-CAA-011 — the rate limiter now sees the human-format entry.
func TestRateLimiter_SeesHumanFormatEntry(t *testing.T) {
	logPath := writeLog(t, t.TempDir(), realFence(humanFormatSection))
	proposal := &AmendmentProposal{RuleID: "CONST-V3R2-003", Before: "a", After: "b"}

	within := &rateLimiter{now: func() time.Time { return time.Date(2026, 5, 13, 1, 0, 0, 0, time.UTC) }}
	err := within.Admit(proposal, logPath)
	var rle *ErrRateLimitExceeded
	if !errors.As(err, &rle) {
		t.Fatalf("Admit at 2026-05-13T01:00Z = %v, want *ErrRateLimitExceeded (24h cooldown)", err)
	}

	later := &rateLimiter{now: func() time.Time { return time.Date(2026, 6, 13, 0, 0, 0, 0, time.UTC) }}
	if err := later.Admit(proposal, logPath); err != nil {
		t.Fatalf("Admit at 2026-06-13T00:00Z = %v, want nil (the rejection must come from the entry)", err)
	}
}

// AC-CAA-018 — the fail-closed error names file, line, and key.
func TestLoadEvolutionLogs_FailClosedErrorLocation(t *testing.T) {
	t.Run("unparseable_timestamp", func(t *testing.T) {
		// Lines 1-5 prose; the fence opens on line 6; timestamp sits on file
		// line 10 (L1) and on block-relative line 4.
		content := "# Evolution Log\n" + // 1
			"\n" + // 2
			"Prose line one.\n" + // 3
			"Prose line two.\n" + // 4
			"\n" + // 5
			"~~~yaml\n" + // 6
			"id: EVO-X-001\n" + // 7  (block line 1)
			"zone: Frozen\n" + // 8
			"approver: reviewer\n" + // 9
			"timestamp: \"not-a-date\"\n" + // 10 (block line 4)
			"~~~\n" // 11
		dir := t.TempDir()
		logPath := writeLog(t, dir, realFence(content))
		_, err := LoadEvolutionLogs(logPath)
		if err == nil {
			t.Fatal("want an error, got nil")
		}
		msg := err.Error()
		if !containsPathForm(msg, logPath) {
			t.Errorf("error %q does not name the log path %s", msg, logPath)
		}
		stripped := stripFixturePaths(msg, logPath, dir)
		for _, want := range []string{"timestamp", "EVO-X-001"} {
			if !strings.Contains(stripped, want) {
				t.Errorf("error %q lacks %q", msg, want)
			}
		}
		if !hasWholeNumber(stripped, "10") {
			t.Errorf("error %q lacks the file line 10 (path-stripped: %q)", msg, stripped)
		}
		if hasWholeNumber(stripped, "4") {
			t.Errorf("error %q carries the block-relative line 4 (path-stripped: %q)", msg, stripped)
		}
	})

	t.Run("missing_timestamp", func(t *testing.T) {
		// Lines 1-3 prose; the opening --- sits on line 4; id: sits on file
		// line 7 (L2) and on segment-relative line 3.
		content := "# Evolution Log\n" + // 1
			"\n" + // 2
			"Prose.\n" + // 3
			"---\n" + // 4
			"rule_id: CONST-V3R2-012\n" + // 5 (segment line 1)
			"approved_by: human\n" + // 6
			"id: LEARN-20260911-009\n" + // 7 (segment line 3)
			"---\n" // 8
		dir := t.TempDir()
		logPath := writeLog(t, dir, content)
		_, err := LoadEvolutionLogs(logPath)
		if err == nil {
			t.Fatal("want an error, got nil")
		}
		msg := err.Error()
		if !containsPathForm(msg, logPath) {
			t.Errorf("error %q does not name the log path %s", msg, logPath)
		}
		stripped := stripFixturePaths(msg, logPath, dir)
		for _, want := range []string{"approved_at", "LEARN-20260911-009"} {
			if !strings.Contains(stripped, want) {
				t.Errorf("error %q lacks %q", msg, want)
			}
		}
		if !hasWholeNumber(stripped, "7") {
			t.Errorf("error %q lacks the file line 7 (path-stripped: %q)", msg, stripped)
		}
		if hasWholeNumber(stripped, "3") {
			t.Errorf("error %q carries the segment-relative line 3 (path-stripped: %q)", msg, stripped)
		}
	})
}
