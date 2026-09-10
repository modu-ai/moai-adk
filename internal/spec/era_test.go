// era_test.go — TDD coverage for era classification engine.
// SPEC-V3R6-LIFECYCLE-SYNC-GATE-001 AC-LSG-002, AC-LSG-013, AC-LSG-017.
package spec

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestClassifyEra_HeuristicTable(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		signals  EraSignals
		wantEra  Era
		wantRule string // substring of returned rule label
	}{
		// AC-LSG-013 — Era auto-detection (no frontmatter era field)
		{
			name:     "H-1 progress.md absent → V2.x",
			signals:  EraSignals{ProgressMDExists: false},
			wantEra:  EraV2x,
			wantRule: "H-1",
		},
		{
			name: "H-2 progress.md present but no §E.* markers → V3R2-R4",
			signals: EraSignals{
				ProgressMDExists:  true,
				ProgressMDContent: "# Progress\n\n## Section A\nnotes",
			},
			wantEra:  EraV3R2R4,
			wantRule: "H-2",
		},
		{
			name: "H-3 §E.2 present but sync_commit_sha empty → V3R5",
			signals: EraSignals{
				ProgressMDExists: true,
				ProgressMDContent: `# Progress
## §E.2 Sync-phase signal
sync_commit_sha:
`,
			},
			wantEra:  EraV3R5,
			wantRule: "H-3",
		},
		{
			name: "H-3 §E.2 present + null sync_commit_sha → V3R5",
			signals: EraSignals{
				ProgressMDExists: true,
				ProgressMDContent: `## §E.2 Sync
sync_commit_sha: null
`,
			},
			wantEra:  EraV3R5,
			wantRule: "H-3",
		},
		{
			name: "H-4 (new) §E.2 + §E.4 + sync_commit_sha → V3R6 (3-phase predicate, REQ-LR-005)",
			signals: EraSignals{
				ProgressMDExists: true,
				ProgressMDContent: `## §E.2 Run-phase Evidence
## §E.4 Sync-phase Audit-Ready Signal
sync_commit_sha: abc1234567890abc1234567890abc1234567890a
`,
			},
			wantEra:  EraV3R6,
			wantRule: "H-4 (§E.2 + §E.4",
		},
		{
			name: "H-4-legacy §E.2 + §E.5 + both SHAs → V3R6 (migration window, REQ-LR-006)",
			signals: EraSignals{
				ProgressMDExists: true,
				ProgressMDContent: `## §E.2 Sync-phase signal
sync_commit_sha: abc1234567890abc1234567890abc1234567890a

## §E.5 Mx-phase audit-ready signal
mx_commit_sha: def4567890def4567890def4567890def4567890
`,
			},
			wantEra:  EraV3R6,
			wantRule: "H-4-legacy",
		},
		{
			name: "H-3 §E.2 + §E.4 present but sync_commit_sha empty → V3R5 (sync not yet complete)",
			signals: EraSignals{
				ProgressMDExists: true,
				ProgressMDContent: `## §E.2 Run-phase Evidence
## §E.4 Sync-phase Audit-Ready Signal
sync_commit_sha:
`,
			},
			wantEra:  EraV3R5,
			wantRule: "H-3",
		},
		// SPEC-ERA-H3-NARROWING-001 — H-3 is deferred when a modern-era signal is
		// present, so an in-flight modern SPEC (plan-phase §E.2 skeleton, sync not
		// yet closed) reaches H-5 instead of being frozen at V3R5.
		//
		// Mutation contract: removing `&& !hasModernEraSignal(signals)` from the H-3
		// predicate in era.go MUST break the next two cases (AC-EH3-001/002); removing
		// the `matchesModernPhase(...) ||` term from hasModernEraSignal MUST break
		// AC-EH3-002 alone (the corpus probe cannot distinguish the phase path).
		{
			name: "AC-EH3-001 H-3 deferred on created signal → V3R6 via H-5",
			signals: EraSignals{
				ProgressMDExists: true,
				ProgressMDContent: `## §E.2 Run-phase Evidence
sync_commit_sha:
`,
				FrontmatterCreated: "2026-08-25",
			},
			wantEra:  EraV3R6,
			wantRule: "H-5",
		},
		{
			name: "AC-EH3-002 H-3 deferred on phase signal → V3R6 via H-5",
			signals: EraSignals{
				ProgressMDExists: true,
				ProgressMDContent: `## §E.2 Run-phase Evidence
sync_commit_sha:
`,
				FrontmatterCreated: "",
				FrontmatterPhase:   "v3.0.0",
			},
			wantEra:  EraV3R6,
			wantRule: "H-5",
		},
		// Mutation contract: replacing the H-3 predicate with an unconditional skip
		// MUST break the next two cases (AC-EH3-003/004) — AC-EH3-004 by falling to
		// H-6 unclassified, which is the grandfather-loss vector.
		{
			name: "AC-EH3-003 pre-threshold created keeps grandfather protection → V3R5 via H-3",
			signals: EraSignals{
				ProgressMDExists: true,
				ProgressMDContent: `## §E.2 Run-phase Evidence
sync_commit_sha:
`,
				FrontmatterCreated: "2026-03-15",
				FrontmatterPhase:   "",
			},
			wantEra:  EraV3R5,
			wantRule: "H-3",
		},
		{
			name: "AC-EH3-004 no signal at all stays V3R5, does not fall to H-6",
			signals: EraSignals{
				ProgressMDExists: true,
				ProgressMDContent: `## §E.2 Run-phase Evidence
sync_commit_sha:
`,
				FrontmatterCreated: "",
				FrontmatterPhase:   "",
				FrontmatterEra:     "",
			},
			wantEra:  EraV3R5,
			wantRule: "H-3",
		},
		{
			name: "H-5 phase tie-breaker v3.0 → V3R6",
			signals: EraSignals{
				ProgressMDExists:  true,
				ProgressMDContent: "## §E.2 Sync\nsync_commit_sha: abc123\n",
				FrontmatterPhase:  "v3.0.0",
			},
			wantEra:  EraV3R6,
			wantRule: "H-5",
		},
		{
			name: "H-5 created date >= 2026-04-01 → V3R6",
			signals: EraSignals{
				ProgressMDExists:   true,
				ProgressMDContent:  "## §E.2 Sync\nsync_commit_sha: abc\n",
				FrontmatterCreated: "2026-05-25",
			},
			wantEra:  EraV3R6,
			wantRule: "H-5",
		},
		{
			name: "H-5 created date before threshold → not V3R6",
			signals: EraSignals{
				ProgressMDExists:   true,
				ProgressMDContent:  "## §E.2 Sync\nsync_commit_sha: abc\n",
				FrontmatterCreated: "2026-03-15",
			},
			// H-3 takes over because sync_commit_sha looks like a real SHA but mx absent
			// Actually H-3 only fires if sync_commit_sha empty. With non-empty + no mx →
			// fall through. Phase empty, created before threshold → H-6.
			// Wait: the §E.2 with sync_commit_sha != empty AND no mx section means
			// H-3 (sync empty) fails, H-4 (both present) fails, H-5 (phase/date) fails →
			// H-6 unclassified.
			wantEra:  EraUnclassified,
			wantRule: "H-6",
		},
		{
			name: "H-6 ambiguous → unclassified",
			signals: EraSignals{
				ProgressMDExists:  true,
				ProgressMDContent: "## §E.2 Sync\nsync_commit_sha: abc123\n", // sync present, no mx, no phase, no date
			},
			wantEra:  EraUnclassified,
			wantRule: "H-6",
		},

		// H-override: explicit frontmatter era field
		{
			name: "H-override V3R6 wins regardless of signals",
			signals: EraSignals{
				FrontmatterEra:   "V3R6",
				ProgressMDExists: false, // would otherwise be H-1 → V2.x
			},
			wantEra:  EraV3R6,
			wantRule: "H-override",
		},
		{
			name: "H-override V2.x wins",
			signals: EraSignals{
				FrontmatterEra:   "V2.x",
				ProgressMDExists: true,
				ProgressMDContent: `## §E.2
sync_commit_sha: abc123
## §E.5
mx_commit_sha: def456
`,
			},
			wantEra:  EraV2x,
			wantRule: "H-override",
		},
		{
			name: "H-override invalid value falls through to H-1",
			signals: EraSignals{
				FrontmatterEra:   "GARBAGE",
				ProgressMDExists: false,
			},
			wantEra:  EraV2x,
			wantRule: "H-1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			gotEra, gotRule := ClassifyEra(tt.signals)
			if gotEra != tt.wantEra {
				t.Errorf("ClassifyEra() era = %v, want %v (rule: %s)", gotEra, tt.wantEra, gotRule)
			}
			if tt.wantRule != "" && !containsSubstring(gotRule, tt.wantRule) {
				t.Errorf("ClassifyEra() rule = %q, want substring %q", gotRule, tt.wantRule)
			}
		})
	}
}

// TestClassifyEra_NoV3R5WhileModernSignal — SPEC-ERA-H3-NARROWING-001 AC-EH3-008.
//
// Combinatorial invariant guard: across the product of {§E.2 present} ×
// {sync_commit_sha present} × {created} × {phase}, no input that the H-5 predicate
// calls modern may classify as V3R5.
//
// The modern predicate is recomputed here from independent literals — it deliberately
// does NOT call hasModernEraSignal. Calling the production helper would make the test
// assume the very identity it exists to check: if the helper and the H-5 condition
// later diverge, a helper-driven guard keeps passing.
//
// The second assertion (rationale reaches H-5) is scoped to the H-3-eligible shape
// (§E.2 present AND sync_commit_sha empty) — the shape H-3 would otherwise have
// claimed. Outside that shape, earlier heuristics legitimately win (H-2 on a
// marker-less progress.md, H-4 on a closed sync), and demanding "H-5" there would
// require reordering the heuristic chain, which REQ-EH3-003 forbids.
//
// Mutation contract: removing `&& !hasModernEraSignal(signals)` from the H-3
// predicate in era.go MUST make this test fail.
func TestClassifyEra_NoV3R5WhileModernSignal(t *testing.T) {
	t.Parallel()

	sectionE2 := []bool{true, false}
	syncSHAPresent := []bool{true, false}
	createdValues := []string{"", "2026-03-15", "2026-08-25"}
	phaseValues := []string{"", "v3.0.0", "v3.2.0 target"}

	// Independent re-implementation of the H-5 predicate (test-local literals).
	modernSignal := func(created, phase string) bool {
		p := strings.ToLower(strings.TrimSpace(phase))
		if strings.Contains(p, "v3r6") || strings.HasPrefix(p, "v3.0") {
			return true
		}
		c := strings.TrimSpace(created)
		return c != "" && c >= "2026-04-01"
	}

	covered := 0
	for _, hasE2 := range sectionE2 {
		for _, hasSHA := range syncSHAPresent {
			for _, created := range createdValues {
				for _, phase := range phaseValues {
					content := "# Progress\n"
					if hasE2 {
						content += "## §E.2 Run-phase Evidence\n"
					}
					if hasSHA {
						content += "sync_commit_sha: abc1234567890abc1234567890abc1234567890a\n"
					} else {
						content += "sync_commit_sha:\n"
					}

					signals := EraSignals{
						ProgressMDExists:   true,
						ProgressMDContent:  content,
						FrontmatterCreated: created,
						FrontmatterPhase:   phase,
					}
					if !modernSignal(created, phase) {
						continue
					}
					covered++

					gotEra, gotRule := ClassifyEra(signals)
					if gotEra == EraV3R5 {
						t.Errorf("modern signal (created=%q phase=%q e2=%v sha=%v) classified V3R5 (rule %q)",
							created, phase, hasE2, hasSHA, gotRule)
					}
					// H-3-eligible shape: H-5 must actually be the branch that decides.
					if hasE2 && !hasSHA && !strings.HasPrefix(gotRule, "H-5") {
						t.Errorf("H-3-eligible modern input (created=%q phase=%q) resolved by %q, want H-5",
							created, phase, gotRule)
					}
				}
			}
		}
	}

	// Empty-sweep guard: a selector matching zero combinations is green for the
	// wrong reason. Of the 9 (created, phase) pairs, 5 are modern; times the 4
	// {§E.2} × {sync_commit_sha} shapes = 20 expected combinations.
	if covered != 20 {
		t.Fatalf("swept %d modern-signal combinations, want 20 — the product changed", covered)
	}
}

// AC-LSG-002 — Era classification 5 buckets coverage
func TestClassifyEra_FiveBucketCoverage(t *testing.T) {
	t.Parallel()

	buckets := map[Era]bool{
		EraV2x:          false,
		EraV3R2R4:       false,
		EraV3R5:         false,
		EraV3R6:         false,
		EraUnclassified: false,
	}

	cases := []EraSignals{
		{ProgressMDExists: false},                                                                                   // V2.x
		{ProgressMDExists: true, ProgressMDContent: "no markers"},                                                   // V3R2-R4
		{ProgressMDExists: true, ProgressMDContent: "## §E.2\nsync_commit_sha:\n"},                                  // V3R5
		{ProgressMDExists: true, ProgressMDContent: "## §E.2\nsync_commit_sha: abc\n## §E.5\nmx_commit_sha: def\n"}, // V3R6
		{ProgressMDExists: true, ProgressMDContent: "## §E.2\nsync_commit_sha: abc\n"},                              // unclassified (no mx)
	}

	for i, sig := range cases {
		era, rule := ClassifyEra(sig)
		buckets[era] = true
		t.Logf("case %d → era=%s rule=%s", i, era, rule)
	}

	for era, hit := range buckets {
		if !hit {
			t.Errorf("Era bucket %v not exercised (5-bucket coverage required by AC-LSG-002)", era)
		}
	}
}

// AC-LSG-017 — Backward compatibility: pre-V3R6 SPECs classify as era_final: true
func TestEra_GrandfatherClause(t *testing.T) {
	t.Parallel()

	grandfathered := []Era{EraV2x, EraV3R2R4, EraV3R5}
	for _, era := range grandfathered {
		if !era.EraFinal() {
			t.Errorf("Era %v should be grandfather-protected (EraFinal=true)", era)
		}
		if era.IsModern() {
			t.Errorf("Era %v should not be modern (IsModern=false)", era)
		}
	}

	modern := []Era{EraV3R6, EraUnclassified}
	for _, era := range modern {
		if era.EraFinal() {
			t.Errorf("Era %v should NOT be grandfather-protected", era)
		}
	}

	if !EraV3R6.IsModern() {
		t.Error("EraV3R6 should be modern (IsModern=true)")
	}
}

// LoadEraSignalsFromDir integration — verifies disk I/O extraction.
func TestLoadEraSignalsFromDir(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	specDir := filepath.Join(tempDir, "SPEC-TEST-001")
	if err := os.MkdirAll(specDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Write minimal spec.md with frontmatter
	specMD := `---
id: SPEC-TEST-001
title: "Test"
version: "0.1.0"
status: draft
created: 2026-05-25
updated: 2026-05-25
author: Test
priority: P1
phase: "v3.0.0"
module: "internal/test"
lifecycle: spec-anchored
tags: "test"
era: "V3R6"
---

# Body
`
	if err := os.WriteFile(filepath.Join(specDir, "spec.md"), []byte(specMD), 0644); err != nil {
		t.Fatal(err)
	}

	// Write progress.md with §E.2 + §E.5
	progressMD := `# Progress

## §E.2 Sync-phase signal
sync_commit_sha: abc1234

## §E.5 Mx-phase signal
mx_commit_sha: def5678
`
	if err := os.WriteFile(filepath.Join(specDir, "progress.md"), []byte(progressMD), 0644); err != nil {
		t.Fatal(err)
	}

	signals, err := LoadEraSignalsFromDir(specDir)
	if err != nil {
		t.Fatalf("LoadEraSignalsFromDir() error = %v", err)
	}

	if signals.FrontmatterEra != "V3R6" {
		t.Errorf("FrontmatterEra = %q, want V3R6", signals.FrontmatterEra)
	}
	if signals.FrontmatterPhase != "v3.0.0" {
		t.Errorf("FrontmatterPhase = %q, want v3.0.0", signals.FrontmatterPhase)
	}
	if signals.FrontmatterCreated != "2026-05-25" {
		t.Errorf("FrontmatterCreated = %q, want 2026-05-25", signals.FrontmatterCreated)
	}
	if !signals.ProgressMDExists {
		t.Error("ProgressMDExists = false, want true")
	}
	if !containsSubstring(signals.ProgressMDContent, "§E.2") {
		t.Error("ProgressMDContent missing §E.2 marker")
	}

	// End-to-end: classify
	era, rule := ClassifyEra(signals)
	if era != EraV3R6 {
		t.Errorf("end-to-end classification: era = %v rule=%s, want V3R6", era, rule)
	}
}

// extractProgressField unit coverage — both YAML and markdown styles.
func TestExtractProgressField(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		content string
		field   string
		want    string
	}{
		{"YAML style", "sync_commit_sha: abc123\n", "sync_commit_sha", "abc123"},
		{"YAML with whitespace", "  sync_commit_sha:  abc123  \n", "sync_commit_sha", "abc123"},
		{"markdown list", "- sync_commit_sha: abc123\n", "sync_commit_sha", "abc123"},
		{"markdown with backticks", "- `sync_commit_sha`: abc123\n", "sync_commit_sha", "abc123"},
		{"absent field", "other_field: xyz\n", "sync_commit_sha", ""},
		{"null value", "sync_commit_sha: null\n", "sync_commit_sha", ""},
		{"pending placeholder", "sync_commit_sha: <pending>\n", "sync_commit_sha", ""},
		{"empty string value", `sync_commit_sha: ""`, "sync_commit_sha", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := extractProgressField(tt.content, tt.field)
			if got != tt.want {
				t.Errorf("extractProgressField() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestNormalizeEra(t *testing.T) {
	t.Parallel()

	cases := []struct {
		input  string
		want   Era
		wantOK bool
	}{
		{"V3R6", EraV3R6, true},
		{"v3r6", EraV3R6, true},
		{"V2.x", EraV2x, true},
		{"V3R2", EraV3R2R4, true},
		{"V3R5", EraV3R5, true},
		{"unclassified", EraUnclassified, true},
		{"GARBAGE", EraUnclassified, false},
		{"", EraUnclassified, false},
	}

	for _, c := range cases {
		t.Run(c.input, func(t *testing.T) {
			t.Parallel()
			got, ok := normalizeEra(c.input)
			if got != c.want || ok != c.wantOK {
				t.Errorf("normalizeEra(%q) = (%v, %v), want (%v, %v)", c.input, got, ok, c.want, c.wantOK)
			}
		})
	}
}

// helper
func containsSubstring(s, sub string) bool {
	return len(sub) == 0 || (len(s) >= len(sub) && (s == sub || indexOf(s, sub) >= 0))
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
