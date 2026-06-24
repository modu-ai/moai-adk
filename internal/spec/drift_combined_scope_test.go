package spec

import (
	"testing"
)

// TestDeriveScopePrefix는 scope-prefix 파생 로직을 검증한다 (mechanism ①).
// full SPEC-ID에서 trailing distinguishing-segment(+number)를 strip하여 combined-scope
// 그룹이 사용하는 prefix(SPEC-{PREFIX})를 얻는다.
func TestDeriveScopePrefix(t *testing.T) {
	tests := []struct {
		name   string
		specID string
		want   string
	}{
		{"single distinguishing segment", "SPEC-CCSYNC-CLAUDEMD-001", "SPEC-CCSYNC"},
		{"another sibling", "SPEC-CCSYNC-TOOLCAT-001", "SPEC-CCSYNC"},
		{"multi-segment id strips only trailing pair", "SPEC-V3R6-DRIFT-LEGACY-CONVENTION-001", "SPEC-V3R6-DRIFT-LEGACY"},
		{"two-segment family", "SPEC-ABC-FOO-001", "SPEC-ABC"},
		{"single-domain id (no further strip target)", "SPEC-FOO-001", "SPEC"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := deriveScopePrefix(tt.specID)
			if got != tt.want {
				t.Errorf("deriveScopePrefix(%q) = %q, want %q", tt.specID, got, tt.want)
			}
		})
	}
}

// TestCombinedScopeCloseMatches는 3-gate combined-scope 매칭 로직을 검증한다
// (REQ-DLC-001/002, AC-DLC-001/012). FALLBACK-ONLY는 DetectDrift 통합에서 적용되므로
// 여기서는 gate (a) prefix + (b) close-infix + (c) distinguishing-segment word-boundary
// token match를 검증한다.
func TestCombinedScopeCloseMatches(t *testing.T) {
	tests := []struct {
		name    string
		subject string
		specID  string
		want    bool
	}{
		{
			name:    "FOO matches (FOO + BAR) combined close",
			subject: "chore(SPEC-ABC): Mx-phase audit-ready signal + 4-phase close (FOO + BAR)",
			specID:  "SPEC-ABC-FOO-001",
			want:    true,
		},
		{
			// SPEC-V3R6-LIFECYCLE-REDESIGN-001 AC-LR-012 (REQ-LR-020, D4): 새 canonical
			// "3-phase close" infix로도 combined-scope close가 인식되어야 한다.
			name:    "FOO matches (FOO + BAR) combined close with 3-phase close infix (REQ-LR-020)",
			subject: "chore(SPEC-ABC): sync-phase audit-ready signal + 3-phase close (FOO + BAR)",
			specID:  "SPEC-ABC-FOO-001",
			want:    true,
		},
		{
			name:    "BAR matches (FOO + BAR) combined close",
			subject: "chore(SPEC-ABC): Mx-phase audit-ready signal + 4-phase close (FOO + BAR)",
			specID:  "SPEC-ABC-BAR-001",
			want:    true,
		},
		{
			name:    "OTHER NOT named in (FOO + BAR) → no match (gate c)",
			subject: "chore(SPEC-ABC): Mx-phase audit-ready signal + 4-phase close (FOO + BAR)",
			specID:  "SPEC-ABC-OTHER-002",
			want:    false,
		},
		{
			// D-NEW-1: word-boundary token match. FOO must NOT be cleared by (FOOBAR + BAZ).
			name:    "FOO NOT falsely cleared by (FOOBAR + BAZ) — word-boundary (D-NEW-1)",
			subject: "chore(SPEC-ABC): Mx-phase audit-ready signal + 4-phase close (FOOBAR + BAZ)",
			specID:  "SPEC-ABC-FOO-001",
			want:    false,
		},
		{
			// converse: FOOBAR matches its own combined close
			name:    "FOOBAR matches (FOOBAR + BAZ)",
			subject: "chore(SPEC-ABC): Mx-phase audit-ready signal + 4-phase close (FOOBAR + BAZ)",
			specID:  "SPEC-ABC-FOOBAR-001",
			want:    true,
		},
		{
			// gate (b): no close-infix → no match (bare combined-scope feat/chore)
			name:    "no close-infix → no match (gate b)",
			subject: "chore(SPEC-ABC): partial work (FOO + BAR)",
			specID:  "SPEC-ABC-FOO-001",
			want:    false,
		},
		{
			// gate (a): subject prefix names a full SPEC-ID (with -NNN), not a scope-prefix → no match
			name:    "prefix carries trailing -NNN → not a combined-scope subject (gate a)",
			subject: "chore(SPEC-ABC-FOO-001): Mx-phase audit-ready signal + 4-phase close",
			specID:  "SPEC-ABC-FOO-001",
			want:    false,
		},
		{
			// real-repo shape: CCSYNC combined close names both CLAUDEMD + TOOLCAT
			name:    "CLAUDEMD matches real CCSYNC combined close",
			subject: "chore(SPEC-CCSYNC): Mx-phase 4-phase close (CLAUDEMD + TOOLCAT status→completed, §E.5) + CHANGELOG fact correction",
			specID:  "SPEC-CCSYNC-CLAUDEMD-001",
			want:    true,
		},
		{
			name:    "TOOLCAT matches real CCSYNC combined close",
			subject: "chore(SPEC-CCSYNC): Mx-phase 4-phase close (CLAUDEMD + TOOLCAT status→completed, §E.5) + CHANGELOG fact correction",
			specID:  "SPEC-CCSYNC-TOOLCAT-001",
			want:    true,
		},
		{
			// load-bearing collision case: DYNWF NOT named in the CCSYNC close → no match
			name:    "DYNWF NOT named in CCSYNC close → no match (load-bearing collision guard)",
			subject: "chore(SPEC-CCSYNC): Mx-phase 4-phase close (CLAUDEMD + TOOLCAT status→completed, §E.5) + CHANGELOG fact correction",
			specID:  "SPEC-CCSYNC-DYNWF-001",
			want:    false,
		},
		{
			// hyphen-delimited prefix boundary: SPEC-CCSYNC scope does NOT apply to SPEC-CCSYNCEXTRA-001
			// (deriveScopePrefix(SPEC-CCSYNCEXTRA-001) = SPEC-CCSYNC? no — strips trailing -NNN only).
			// This case is primarily guarded at the DetectDrift integration boundary (prefix grep),
			// but the matcher must also not cross the hyphen boundary.
			name:    "SPEC-CCSYNCEXTRA-001 not matched by SPEC-CCSYNC combined close (hyphen boundary)",
			subject: "chore(SPEC-CCSYNC): Mx-phase 4-phase close (EXTRA + TOOLCAT)",
			specID:  "SPEC-CCSYNCEXTRA-001",
			want:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := combinedScopeCloseMatches(tt.subject, tt.specID)
			if got != tt.want {
				t.Errorf("combinedScopeCloseMatches(%q, %q) = %v, want %v", tt.subject, tt.specID, got, tt.want)
			}
		})
	}
}

// TestDetectDrift_CombinedScopeFallback는 AC-DLC-001 BINDING 검증 — 하나의 combined-scope
// close가 양쪽 sibling을 모두 completed로 resolve하는지 deterministic fixture로 검증한다.
// secondary scope-prefix grep fallback이 per-SPEC primary walk가 completed를 못 찾을 때만
// fire하고, 양쪽 sibling을 모두 non-drift로 만든다.
func TestDetectDrift_CombinedScopeFallback(t *testing.T) {
	baseDir := t.TempDir()

	// 두 sibling SPEC: frontmatter completed + V3R6 progress.md (grandfather-exempt 아님 → git walk 진입).
	writeSpecFixture(t, baseDir, "SPEC-ABC-FOO-001", "completed", "2026-05-01", progressV3R6)
	writeSpecFixture(t, baseDir, "SPEC-ABC-BAR-001", "completed", "2026-05-01", progressV3R6)

	// git history (oldest→newest): 두 feat + combined-scope close.
	// 핵심: combined close subject는 SPEC-ABC만 명명 (full sibling ID 없음) →
	// per-SPEC primary walk(--grep=SPEC-ABC-FOO-001)에는 안 잡히고, secondary
	// prefix grep(--grep=SPEC-ABC)에만 잡힌다.
	initGitInDir(t, baseDir, []string{
		"feat(SPEC-ABC-FOO-001): M1 implementation",
		"feat(SPEC-ABC-BAR-001): M1 implementation",
		"chore(SPEC-ABC): Mx-phase audit-ready signal + 4-phase close (FOO + BAR)",
	})
	chdirTo(t, baseDir)

	report, err := DetectDrift(baseDir)
	if err != nil {
		t.Fatalf("DetectDrift: %v", err)
	}

	for _, id := range []string{"SPEC-ABC-FOO-001", "SPEC-ABC-BAR-001"} {
		rec, ok := findRecord(report, id)
		if !ok {
			t.Fatalf("record for %s 누락", id)
		}
		if rec.Drifted {
			t.Errorf("%s: Drifted=true, want false (combined-scope close가 secondary prefix-grep으로 completed resolve해야 함). GitImpliedStatus=%q", id, rec.GitImpliedStatus)
		}
	}
}

// TestDetectDrift_CombinedScopeCollisionGuard는 AC-DLC-012 검증 — combined close에 명명되지
// 않은 same-prefix sibling은 fallback으로 false-clear되면 안 된다 (LSGF-001 collision guard).
func TestDetectDrift_CombinedScopeCollisionGuard(t *testing.T) {
	baseDir := t.TempDir()

	// OTHER는 SPEC-ABC prefix이지만 close (FOO + BAR)에 명명되지 않음 → fallback 적용 안 됨.
	// frontmatter completed인데 git 추론은 implemented (feat) → genuine drift로 남아야 한다.
	writeSpecFixture(t, baseDir, "SPEC-ABC-OTHER-002", "completed", "2026-05-01", progressV3R6)

	initGitInDir(t, baseDir, []string{
		"feat(SPEC-ABC-OTHER-002): M1 implementation",
		"chore(SPEC-ABC): Mx-phase audit-ready signal + 4-phase close (FOO + BAR)",
	})
	chdirTo(t, baseDir)

	report, err := DetectDrift(baseDir)
	if err != nil {
		t.Fatalf("DetectDrift: %v", err)
	}

	rec, ok := findRecord(report, "SPEC-ABC-OTHER-002")
	if !ok {
		t.Fatalf("record for SPEC-ABC-OTHER-002 누락")
	}
	// gate (c) 실패 → fallback 미적용 → primary walk status(implemented) 유지 → completed↔implemented drift.
	if !rec.Drifted {
		t.Errorf("SPEC-ABC-OTHER-002: Drifted=false, want true — combined close (FOO + BAR)에 명명되지 않았으므로 false-clear되면 안 됨 (collision guard). GitImpliedStatus=%q", rec.GitImpliedStatus)
	}
}

// TestDetectDrift_CombinedScopeNoCloseInfixNoFallback는 negative 케이스 — close-infix가 없는
// combined-scope commit은 fallback을 trigger하면 안 된다 (gate b). genuine-⑤ 보호.
func TestDetectDrift_CombinedScopeNoCloseInfixNoFallback(t *testing.T) {
	baseDir := t.TempDir()

	writeSpecFixture(t, baseDir, "SPEC-ABC-FOO-001", "completed", "2026-05-01", progressV3R6)

	// combined-scope commit이지만 close-infix 없음 → fallback 미적용 → drift 유지.
	initGitInDir(t, baseDir, []string{
		"feat(SPEC-ABC-FOO-001): M1 implementation",
		"chore(SPEC-ABC): partial sweep work (FOO + BAR)",
	})
	chdirTo(t, baseDir)

	report, err := DetectDrift(baseDir)
	if err != nil {
		t.Fatalf("DetectDrift: %v", err)
	}

	rec, ok := findRecord(report, "SPEC-ABC-FOO-001")
	if !ok {
		t.Fatalf("record for SPEC-ABC-FOO-001 누락")
	}
	if !rec.Drifted {
		t.Errorf("SPEC-ABC-FOO-001: Drifted=false, want true — close-infix 없는 combined-scope는 fallback 미적용 (gate b, genuine-⑤ 보호). GitImpliedStatus=%q", rec.GitImpliedStatus)
	}
}
