package spec

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strings"
)

// DriftRecord represents a single SPEC status drift entry
type DriftRecord struct {
	SPECID            string
	FrontmatterStatus string
	GitImpliedStatus  string
	Drifted           bool
}

// DriftReport represents a complete drift detection report
type DriftReport struct {
	Records []DriftRecord
	Count   int
}

// DetectDrift scans all SPECs and compares frontmatter status against git log
// Returns a report with all drift records and the total drift count
func DetectDrift(baseDir string) (*DriftReport, error) {
	specsDir := filepath.Join(baseDir, ".moai", "specs")

	// Check if specs directory exists
	if _, err := os.Stat(specsDir); os.IsNotExist(err) {
		return &DriftReport{Records: []DriftRecord{}, Count: 0}, nil
	}

	// Read all SPEC directories
	entries, err := os.ReadDir(specsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read specs directory: %w", err)
	}

	var records []DriftRecord
	driftCount := 0

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		specID := entry.Name()
		specDir := filepath.Join(specsDir, specID)

		// Parse frontmatter status
		frontmatterStatus, err := ParseStatus(specDir)
		if err != nil {
			// Skip SPECs that can't be parsed
			continue
		}

		// (③) terminal-state authority — superseded/archived/rejected frontmatter는
		// 어떤 git commit convention으로도 positive infer할 수 없으므로 frontmatter가
		// authoritative다. audit.go checkV3R6Drift의 terminal early-return과 정합.
		// D3 record-emission contract: record를 Drifted:false로 PRESERVE (drop 금지) —
		// Records[] 소비자(M5 residual classifier)가 전체 SPEC corpus를 관측하도록 한다.
		// @MX:NOTE: [AUTO] terminal frontmatter는 git 추론보다 우선한다.
		// @MX:REASON: SPEC-V3R6-DRIFT-LEGACY-CONVENTION-001 mechanism ③ —
		//   terminal state는 commit convention으로 positive infer 불가 (audit.go와 동일 정책).
		if isTerminalStatus(frontmatterStatus) {
			records = append(records, DriftRecord{
				SPECID:            specID,
				FrontmatterStatus: frontmatterStatus,
				GitImpliedStatus:  "terminal-exempt",
				Drifted:           false,
			})
			continue
		}

		// (④) era/grandfather alignment — grandfather-protected era (EraFinal==true:
		// V2.x/V3R2-R4/V3R5)는 drift 보고에서 제외한다. era.go의 LoadEraSignalsFromDir +
		// ClassifyEra를 READ-ONLY로 재사용하여 moai spec audit과 동일한 grandfather 정책을
		// 적용한다. D3 record-emission contract: record를 Drifted:false로 PRESERVE (drop 금지).
		// @MX:NOTE: [AUTO] grandfather era는 drift 면제 — audit.go grandfather 정책과 정합.
		// @MX:REASON: SPEC-V3R6-DRIFT-LEGACY-CONVENTION-001 mechanism ④ —
		//   ClassifyEra의 H-1..H-4 progress.md 신호 체인에 의존 (standalone created-date 비교 금지, AP-3).
		if signals, sigErr := LoadEraSignalsFromDir(specDir); sigErr == nil {
			if era, _ := ClassifyEra(signals); era.EraFinal() {
				records = append(records, DriftRecord{
					SPECID:            specID,
					FrontmatterStatus: frontmatterStatus,
					GitImpliedStatus:  "era-exempt",
					Drifted:           false,
				})
				continue
			}
		}

		// Get git-implied status
		gitStatus, err := getGitImpliedStatus(specID)
		if err != nil {
			// If git history is empty or unavailable, skip
			continue
		}

		// Check for drift
		drifted := frontmatterStatus != gitStatus

		record := DriftRecord{
			SPECID:            specID,
			FrontmatterStatus: frontmatterStatus,
			GitImpliedStatus:  gitStatus,
			Drifted:           drifted,
		}

		records = append(records, record)

		if drifted {
			driftCount++
		}
	}

	// Sort records by SPEC-ID for consistent output
	sort.Slice(records, func(i, j int) bool {
		return records[i].SPECID < records[j].SPECID
	})

	return &DriftReport{
		Records: records,
		Count:   driftCount,
	}, nil
}

// gitLogWindowSize determines the maximum number of commits getGitImpliedStatus inspects from git log.
//
// @MX:NOTE: [AUTO] Rationale for N=50: average matching git-log commits per SPEC are 5-10, providing a 5-10x safety margin.
// @MX:REASON: SPEC-V3R4-LINT-STATUS-CHORE-SKIP-001 OQ1 (original decision) +
//
//	SPEC-V3R4-LINT-SPECID-GREP-FIX-001 (impact assessment of the word-boundary filter — no change).
//	When changing N, see plan.md §7 OQ1.
const gitLogWindowSize = 50

// getGitImpliedStatus analyzes the git log for a SPEC-ID and infers its lifecycle status.
//
// The walker applies two filters sequentially:
//  1. chore-skip filter (LSCSK-001): excludes chore(spec): sweep commits
//  2. word-boundary filter (LSGF-001): blocks substring collisions (e.g., HARNESS-001 vs HARNESS-NAMESPACE-001)
//
// Adopts the status of the first commit with a meaningful classification (i.e., ClassifyPRTitle returns a non-empty status).
// If all N commits are skip candidates, returns an error, and
// the calling lint rule (StatusGitConsistencyRule) treats this as a skip condition.
//
// @MX:ANCHOR: [AUTO] getGitImpliedStatus — entry point for git-implied status inference
// @MX:REASON: called from two sites — StatusGitConsistencyRule.Check + DetectDrift (fan_in=2);
//
//	the core walker incorporates both LSCSK-001 (chore-skip) and LSGF-001 (word-boundary) fixes.
func getGitImpliedStatus(specID string) (string, error) {
	// Decide default branch — prefer main; fall back to master (preserves current behavior)
	branch := "main"
	if _, err := exec.Command("git", "rev-parse", "--verify", "main").Output(); err != nil {
		branch = "master"
	}

	// Fetch up to N commits referencing the SPEC-ID, newest-first
	cmd := exec.Command("git", "log", branch, "--oneline", "--no-merges",
		"--grep="+specID, fmt.Sprintf("-%d", gitLogWindowSize))
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git log failed: %w", err)
	}

	if len(output) == 0 {
		return "", fmt.Errorf("no git history found for %s", specID)
	}

	// Iterate line by line — newest first
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		line := scanner.Text()
		// Split out the commit hash (format: "<hash> <title>")
		parts := strings.SplitN(line, " ", 2)
		if len(parts) < 2 {
			// Skip malformed lines
			continue
		}
		commitTitle := parts[1]

		// skip-pattern matching — commits excluded from lifecycle inference, e.g., chore(spec):
		// @MX:NOTE: [AUTO] skip chore(spec) commits — guards against a bootstrapping bug
		// @MX:REASON: fixes the SPEC-V3R4-LINT-STATUS-CHORE-SKIP-001 core defect where a sweep commit masked the real impl commit
		if shouldSkipCommitTitle(commitTitle) {
			continue
		}

		// word-boundary SPEC-ID filter (LSGF-001) — blocks substring collisions
		// e.g., prevents specID="SPEC-V3R4-HARNESS-001" from matching "SPEC-V3R4-HARNESS-NAMESPACE-001"
		// @MX:NOTE: [AUTO] LSGF-001 word-boundary filter
		// @MX:REASON: blocks the false positive where a NAMESPACE supersede commit was adopted as the walker's first match
		if !commitMatchesSPECID(commitTitle, specID) {
			continue
		}

		// Classify the commit title
		_, status, err := ClassifyPRTitle(commitTitle)
		if err != nil {
			// On classification failure, safely skip to the next commit
			continue
		}

		if status == "" {
			// Safety net for unknown prefixes — do not treat an empty status as a meaningful classification; continue searching
			continue
		}

		// Meaningful classification found → return immediately
		return status, nil
	}

	// All N commits exhausted without a meaningful classification → return an error
	// StatusGitConsistencyRule::Check (lint.go:897-900) does not emit a finding when err != nil
	// @MX:NOTE: [AUTO] walker exhaustion at N=50 signals "unknown" — the lint rule treats it as skip to prevent false positives
	// @MX:REASON: SPECs whose commits all match skip patterns are excluded from git-consistency checks (fail-safe)
	return "", fmt.Errorf("no classifiable commit within window of %d for %s", gitLogWindowSize, specID)
}

// commitMatchesSPECID checks whether a commit title contains the exact SPEC-ID token.
//
// Because git log --grep=<specID> performs substring matching,
// for example a search for specID="SPEC-V3R4-HARNESS-001"
// also matches a "plan(spec): SPEC-V3R4-HARNESS-NAMESPACE-001 ..." commit — that is a defect.
//
// This function uses ExtractSPECIDs (transitions.go) to extract only exact SPEC-ID tokens
// and then checks whether the target specID is contained in that set.
//
// @MX:NOTE: [AUTO] commitMatchesSPECID — word-boundary SPEC-ID filter (LSGF-001)
// @MX:REASON: SPEC-V3R4-LINT-SPECID-GREP-FIX-001 — blocks the defect where git log --grep substring matching
//
//	caused a NAMESPACE supersede commit to be adopted as the walker's first match.
//	Reusing ExtractSPECIDs introduces zero external dependencies.
func commitMatchesSPECID(commitTitle, specID string) bool {
	extracted := ExtractSPECIDs(commitTitle)
	return slices.Contains(extracted, specID)
}

// shouldSkipCommitTitle checks whether a commit title matches a known skip pattern.
//
// Skip-pattern commits represent metadata maintenance work that must be excluded from
// lifecycle status inference (frontmatter sweeps, lint.skip registration, etc.).
//
// 두 가지 skip 분류:
//  1. metadata-sweep chore: chore(spec): / chore(specs): (LSCSK-001, AC-LSCSK-003 보존)
//  2. SPEC-ID-scoped backfill chore: chore(SPEC-XXX-NNN): ...backfill... — 단,
//     close-infix(4-phase close / Mx-phase audit-ready)를 포함하지 않을 때만 skip.
//     newest-first walker가 backfill chore를 건너뛰고 그 아래의 진짜 close commit에
//     도달하도록 한다 (SPEC-V3R6-DRIFT-CONVENTION-ALIGN-001 REQ-DCA-002/003).
//
// D5 guard (REQ-DCA-005 / AP-2): backfill과 close-infix가 결합된 단일 subject는
// skip하지 않는다 — close-infix가 ClassifyPRTitle에서 이겨 completed로 분류되어야 한다.
//
// @MX:NOTE: [AUTO] shouldSkipCommitTitle — metadata-sweep + narrow backfill skip filter
// @MX:REASON: core helper for SPEC-V3R4-LINT-STATUS-CHORE-SKIP-001 (AC-LSCSK-003 regression guard)
//
//   - SPEC-V3R6-DRIFT-CONVENTION-ALIGN-001 narrow backfill-skip; 변경 시 AC-LSCSK-003 + AC-DCA-003/008 재실행 의무.
func shouldSkipCommitTitle(title string) bool {
	// Case-insensitive prefix match (plan.md §7 OQ1: strings.HasPrefix + ToLower chosen)
	lower := strings.ToLower(strings.TrimSpace(title))

	// (1) metadata-sweep chore: chore(spec): / chore(specs): (AC-LSCSK-003 보존)
	if strings.HasPrefix(lower, "chore(spec):") ||
		strings.HasPrefix(lower, "chore(specs):") {
		return true
	}

	// (2) SPEC-ID-scoped backfill chore: skip하되 close-infix가 있으면 skip 금지 (D5 guard).
	if isSPECIDScopedChore(lower) &&
		strings.Contains(lower, "backfill") &&
		!closeInfixMatch(lower) {
		return true
	}

	return false
}

// specIDScopedChorePattern은 chore(SPEC-XXX-NNN): 형태의 SPEC-ID-scoped chore prefix를
// 매칭한다 (이미 소문자화된 title 대상). metadata-sweep chore(spec):/chore(specs):와
// 구분하기 위해 `spec-` 뒤에 토큰이 이어지는지 확인한다.
//
// @MX:NOTE: [AUTO] SPEC-ID-scoped chore 식별 — backfill skip을 lifecycle-bearing chore에만 적용.
// @MX:REASON: chore(spec):/chore(specs): metadata-sweep과 혼동 방지 (REQ-DCA-002).
var specIDScopedChorePattern = regexp.MustCompile(`^chore\(spec-[a-z0-9-]+-[0-9]+\):`)

// isSPECIDScopedChore는 (소문자) title이 SPEC-ID-scoped chore prefix인지 검사한다.
func isSPECIDScopedChore(lowerTitle string) bool {
	return specIDScopedChorePattern.MatchString(lowerTitle)
}

// isTerminalStatus는 frontmatter status가 terminal lifecycle state인지 검사한다.
// terminal state (superseded/archived/rejected)는 어떤 git commit convention으로도
// positive infer할 수 없으므로 frontmatter가 authoritative다 (mechanism ③).
// audit.go checkV3R6Drift의 terminal early-return과 동일한 enum을 사용한다.
//
// @MX:NOTE: [AUTO] terminal status enum — audit.go의 terminal early-return과 정합.
// @MX:REASON: SPEC-V3R6-DRIFT-LEGACY-CONVENTION-001 — terminal frontmatter authority (③).
func isTerminalStatus(status string) bool {
	switch status {
	case "superseded", "archived", "rejected":
		return true
	default:
		return false
	}
}

// DriftCount is a convenience function that returns only the drift count
func DriftCount(baseDir string) (int, error) {
	report, err := DetectDrift(baseDir)
	if err != nil {
		return 0, err
	}
	return report.Count, nil
}
