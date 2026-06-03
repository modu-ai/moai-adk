package spec

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// writeSpecFixture는 baseDir/.moai/specs/<specID>/ 아래에 spec.md(+선택적 progress.md)를
// 작성한다. DetectDrift가 디렉토리 이름을 specID로 사용하고 ParseStatus가 spec.md
// frontmatter status를 읽는 규약에 맞춘다.
//
// progressMD가 ""이면 progress.md를 생성하지 않는다 (H-1 → V2.x grandfather 재현).
func writeSpecFixture(t *testing.T, baseDir, specID, frontmatterStatus, created, progressMD string) {
	t.Helper()

	specDir := filepath.Join(baseDir, ".moai", "specs", specID)
	if err := os.MkdirAll(specDir, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", specDir, err)
	}

	specMD := "---\n" +
		"id: " + specID + "\n" +
		"title: \"fixture\"\n" +
		"version: \"0.1.0\"\n" +
		"status: " + frontmatterStatus + "\n" +
		"created: " + created + "\n" +
		"updated: " + created + "\n" +
		"author: test\n" +
		"priority: P2\n" +
		"phase: \"v3.0.0\"\n" +
		"module: \"internal/spec\"\n" +
		"lifecycle: spec-anchored\n" +
		"tags: \"fixture\"\n" +
		"---\n\n# fixture\n"
	if err := os.WriteFile(filepath.Join(specDir, "spec.md"), []byte(specMD), 0o644); err != nil {
		t.Fatalf("write spec.md: %v", err)
	}

	if progressMD != "" {
		if err := os.WriteFile(filepath.Join(specDir, "progress.md"), []byte(progressMD), 0o644); err != nil {
			t.Fatalf("write progress.md: %v", err)
		}
	}
}

// findRecord는 DetectDrift 결과에서 specID에 해당하는 record를 찾는다.
func findRecord(report *DriftReport, specID string) (DriftRecord, bool) {
	for _, r := range report.Records {
		if r.SPECID == specID {
			return r, true
		}
	}
	return DriftRecord{}, false
}

// initGitInDir는 baseDir에 git repo를 초기화하고 주어진 commit들을
// oldest→newest 순서로 만든다. DetectDrift가 getGitImpliedStatus를 통해
// 'main' 브랜치의 git log를 읽으므로, fixture baseDir 자체를 'main' 브랜치를 가진
// git repo로 만든다.
func initGitInDir(t *testing.T, baseDir string, commitTitles []string) {
	t.Helper()
	runGit := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = baseDir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	runGit("init", "-b", "main")
	runGit("config", "user.email", "test@example.com")
	runGit("config", "user.name", "Test User")
	for _, title := range commitTitles {
		runGit("commit", "--allow-empty", "-m", title)
	}
}

// chdirTo는 t.Cleanup로 복원되는 CWD 변경 헬퍼다.
func chdirTo(t *testing.T, dir string) {
	t.Helper()
	orig, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(orig); err != nil {
			t.Logf("restore cwd 실패 (무시): %v", err)
		}
	})
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir %s: %v", dir, err)
	}
}

// progressV3R6은 H-4 (§E.2 + §E.5 + 두 SHA) → V3R6 modern era를 재현하는
// progress.md 콘텐츠를 만든다.
const progressV3R6 = "## §E.2 Sync-phase Audit-Ready Signal\n" +
	"sync_commit_sha: \"a1b2c3d4e5f6\"\n\n" +
	"## §E.5 Mx-phase Audit-Ready Signal\n" +
	"mx_commit_sha: \"f6e5d4c3b2a1\"\n"

// progressV3R2R4는 progress.md가 있으나 §E.* 마커가 없는 (H-2 → V3R2-R4 grandfather)
// 콘텐츠를 만든다.
const progressV3R2R4 = "# Progress\n\nplan complete.\n"

// TestDetectDrift_TerminalStateAuthoritative는 AC-DLC-003 (mechanism ③) 검증.
// frontmatter가 terminal state (superseded/archived/rejected)이면 git 추론 상태와
// 무관하게 Drifted=false인 record가 PRESERVE되어야 한다 (D3 record-emission contract).
func TestDetectDrift_TerminalStateAuthoritative(t *testing.T) {
	terminalStates := []string{"superseded", "archived", "rejected"}

	for _, ts := range terminalStates {
		t.Run(ts, func(t *testing.T) {
			baseDir := t.TempDir()
			specID := "SPEC-TERMINAL-001"
			// V3R6 progress (modern era) — terminal-state 검사가 era보다 먼저든 나중이든
			// terminal frontmatter가 authoritative여야 한다. modern era로 두어
			// "terminal authority가 era exemption과 독립적으로 동작"을 확인한다.
			writeSpecFixture(t, baseDir, specID, ts, "2026-05-01", progressV3R6)

			// git 추론이 implemented 등 non-terminal로 떨어지도록 commit 구성.
			// (terminal 상태는 어떤 commit convention으로도 추론 불가 — feat → implemented)
			initGitInDir(t, baseDir, []string{
				"feat(" + specID + "): M1 implementation",
			})
			chdirTo(t, baseDir)

			report, err := DetectDrift(baseDir)
			if err != nil {
				t.Fatalf("DetectDrift: %v", err)
			}

			rec, ok := findRecord(report, specID)
			if !ok {
				t.Fatalf("record for %s 누락 — D3 record-emission contract 위반 (terminal record는 PRESERVE되어야 함)", specID)
			}
			if rec.Drifted {
				t.Errorf("%s (status=%s): Drifted=true, want false (terminal frontmatter authoritative)", specID, ts)
			}
		})
	}
}

// TestDetectDrift_GrandfatherEraExempt는 AC-DLC-004 (mechanism ④) 검증.
// grandfather-protected era (EraFinal==true: V2.x/V3R2-R4/V3R5)는 frontmatter↔git
// mismatch가 있어도 Drifted=false record가 PRESERVE되어야 한다 (D3).
func TestDetectDrift_GrandfatherEraExempt(t *testing.T) {
	t.Run("H-1 V2.x (progress.md absent)", func(t *testing.T) {
		baseDir := t.TempDir()
		specID := "SPEC-V2X-001"
		// progress.md 없음 → H-1 → V2.x → EraFinal==true.
		// frontmatter=completed, git 추론=planned → 의도적 mismatch.
		writeSpecFixture(t, baseDir, specID, "completed", "2026-01-15", "")
		initGitInDir(t, baseDir, []string{
			"plan(spec): " + specID + " — initial draft",
		})
		chdirTo(t, baseDir)

		report, err := DetectDrift(baseDir)
		if err != nil {
			t.Fatalf("DetectDrift: %v", err)
		}
		rec, ok := findRecord(report, specID)
		if !ok {
			t.Fatalf("record for %s 누락 — D3 contract 위반 (grandfather record는 PRESERVE되어야 함)", specID)
		}
		if rec.Drifted {
			t.Errorf("%s (V2.x grandfather): Drifted=true, want false", specID)
		}
	})

	t.Run("H-2 V3R2-R4 (progress.md without §E.* markers, created after threshold)", func(t *testing.T) {
		baseDir := t.TempDir()
		specID := "SPEC-V3R2-LATE-001"
		// AP-3 era-subtlety guard: created 2026-04-23 > modernEraThreshold(2026-04-01)
		// 이지만 progress.md에 §E.* 마커가 없으므로 H-2 → V3R2-R4 (grandfather)가
		// H-5 date tie-breaker보다 먼저 발동해야 한다.
		writeSpecFixture(t, baseDir, specID, "completed", "2026-04-23", progressV3R2R4)
		initGitInDir(t, baseDir, []string{
			"plan(spec): " + specID + " — planned",
		})
		chdirTo(t, baseDir)

		report, err := DetectDrift(baseDir)
		if err != nil {
			t.Fatalf("DetectDrift: %v", err)
		}
		rec, ok := findRecord(report, specID)
		if !ok {
			t.Fatalf("record for %s 누락 — D3 contract 위반", specID)
		}
		if rec.Drifted {
			t.Errorf("%s (V3R2-R4 grandfather, created after threshold): Drifted=true, want false — H-2 must fire before H-5 date tie-breaker (AP-3)", specID)
		}
	})
}

// TestDetectDrift_V3R6StillDetected는 era exemption이 over-suppress하지 않음을 검증
// (AC-DLC-004 control). H-4 modern era (V3R6) SPEC에 genuine mismatch가 있으면
// Drifted=true여야 한다.
func TestDetectDrift_V3R6StillDetected(t *testing.T) {
	baseDir := t.TempDir()
	specID := "SPEC-V3R6-MODERN-001"
	// H-4 V3R6 + frontmatter=in-progress 이지만 git 추론=implemented → genuine drift.
	// (in-progress는 terminal 아님, V3R6는 grandfather 아님 → 정상 검출되어야 함)
	writeSpecFixture(t, baseDir, specID, "in-progress", "2026-05-01", progressV3R6)
	initGitInDir(t, baseDir, []string{
		"feat(" + specID + "): M1 implementation",
	})
	chdirTo(t, baseDir)

	report, err := DetectDrift(baseDir)
	if err != nil {
		t.Fatalf("DetectDrift: %v", err)
	}
	rec, ok := findRecord(report, specID)
	if !ok {
		t.Fatalf("record for %s 누락", specID)
	}
	if !rec.Drifted {
		t.Errorf("%s (V3R6 modern, in-progress↔implemented mismatch): Drifted=false, want true — exemption must NOT over-suppress modern-era drift", specID)
	}
}
