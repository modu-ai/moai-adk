package spec

import (
	"bufio"
	"context"
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

// driftDeps is the injectable dependency seam for the drift computation.
//
// It exists so tests can COUNT git-log invocations directly: the central claim of
// the single-pass design — "a constant number of git subprocesses, independent of
// the SPEC count" — is only verifiable through an observable seam, not by
// inspection.
//
// logAll is the counted operation. headSHA and branch sit deliberately OUTSIDE it:
// both are O(1), and a cache hit must be answerable without any git-log work at all.
//
// @MX:ANCHOR: [AUTO] driftDeps — the drift-path git seam.
// @MX:REASON: SPEC-SESSIONSTART-PERF-001 REQ-SSP-001 / AC-SSP-001 / AC-SSP-004 —
//
//	the O(1)-subprocess guarantee is a load-bearing contract of the session-start
//	critical path; removing this seam removes the only mechanical proof of it.
type driftDeps struct {
	// useCache gates the HEAD-SHA cache READ. False means "always recompute" —
	// the `moai spec drift --no-cache` authoritative path (REQ-SSP-006a).
	useCache bool

	// headSHA resolves the current HEAD commit SHA (the cache key).
	headSHA func() (string, error)

	// branch resolves the default branch name ("main", falling back to "master").
	branch func() string

	// logAll performs THE single git log pass: exactly one invocation per cache
	// miss, zero per cache hit — regardless of how many SPECs exist.
	logAll func(branch string) ([]commitRecord, error)
}

// realDriftDeps wires the production git implementations.
func realDriftDeps(useCache bool) driftDeps {
	return driftDeps{
		useCache: useCache,
		headSHA:  gitHeadSHA,
		branch:   cachedMainBranch,
		logAll:   gitLogAllFullMessage,
	}
}

// activeSpec is a SPEC that survived the cheap pre-filter and therefore still
// requires git classification.
type activeSpec struct {
	specID string
	status string
}

// DetectDrift scans all SPECs and compares frontmatter status against git log.
// Returns a report with all drift records and the total drift count.
//
// While HEAD is unchanged the result is served from the HEAD-SHA cache. Use
// DetectDriftFresh to force a recompute.
//
// @MX:ANCHOR: [AUTO] DetectDrift — public drift entry point (session-start critical path).
// @MX:REASON: fan_in >= 3 (moai spec drift, its --exit-code-on-drift post-run, and
//
//	session-start via DriftCount). It runs synchronously inside the SessionStart hook,
//	so its cost is charged directly to every Claude Code session launch.
func DetectDrift(baseDir string) (*DriftReport, error) {
	return detectDrift(baseDir, realDriftDeps(true))
}

// DetectDriftFresh recomputes drift while ignoring the HEAD-SHA cache.
//
// This is the authoritative on-demand path (REQ-SSP-006a): the cache is keyed only
// on HEAD, so an uncommitted frontmatter edit leaves a stale-cache window, and this
// is how an operator escapes it.
func DetectDriftFresh(baseDir string) (*DriftReport, error) {
	return detectDrift(baseDir, realDriftDeps(false))
}

// detectDrift is the seam-injectable core of drift detection.
//
// Control flow (SPEC-SESSIONSTART-PERF-001 design.md §M1.2):
//
//	(A) HEAD-SHA cache short-circuit — zero git-log work while HEAD is unchanged
//	(B) cheap pre-filter            — terminal + grandfather SPECs resolved with NO git work
//	(C) ONE git log pass            — a single subprocess, whatever N is
//	(D) per-SPEC in-memory walk     — the original two-stage semantics, verbatim
func detectDrift(baseDir string, deps driftDeps) (*DriftReport, error) {
	specsDir := filepath.Join(baseDir, ".moai", "specs")

	// Check if specs directory exists
	if _, err := os.Stat(specsDir); os.IsNotExist(err) {
		return &DriftReport{Records: []DriftRecord{}, Count: 0}, nil
	}

	// (A) HEAD-SHA cache short-circuit. A hit returns without touching deps.logAll.
	// A non-git checkout yields an empty head, which simply disables the cache.
	head, headErr := deps.headSHA()
	if headErr != nil {
		head = ""
	}
	if deps.useCache {
		if cached, ok := loadDriftCache(baseDir, head); ok {
			return cached, nil
		}
	}

	// Read all SPEC directories
	entries, err := os.ReadDir(specsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read specs directory: %w", err)
	}

	var records []DriftRecord
	var active []activeSpec
	driftCount := 0

	// (B) Cheap pre-filter — terminal-status and grandfather-era SPECs are resolved
	// from frontmatter alone and never enter the git-checked working set. On the real
	// corpus this removes the majority of SPECs before a single subprocess runs.
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

		// Survives the pre-filter: this SPEC needs git classification.
		active = append(active, activeSpec{specID: specID, status: frontmatterStatus})
	}

	// (C) THE single git log pass — one subprocess, independent of len(active).
	//
	// A failure here (no repository, no commits on the branch) leaves commits nil, so
	// every active SPEC exhausts its in-memory walk and is skipped. That reproduces
	// the previous per-SPEC `git log` failure path exactly: getGitImpliedStatus
	// returned an error and DetectDrift dropped the record.
	var commits []commitRecord
	if len(active) > 0 {
		if fetched, logErr := deps.logAll(deps.branch()); logErr == nil {
			commits = fetched
		}
	}

	// (D) Per-SPEC in-memory walk over the shared index.
	for _, a := range active {
		gitStatus, err := inMemImpliedStatus(commits, a.specID)
		if err != nil {
			// No git history, or no classifiable commit in the window — skip.
			continue
		}

		// (①) combined-scope secondary prefix lookup (FALLBACK-ONLY) — primary walk가
		// `completed`를 못 찾았는데 frontmatter가 `completed`라면, 이 SPEC은 scope-prefix를
		// 명명하는 combined-scope close commit으로 닫혔을 수 있다 (예: cf7d78a9c
		// "chore(SPEC-CCSYNC): ... 4-phase close (CLAUDEMD + TOOLCAT)"). 이런 close는
		// per-SPEC full-ID 윈도우에 절대 안 잡히므로 (full-ID가 아닌 scope-prefix만 명명),
		// scope-prefix 조회로만 도달 가능하다.
		// 3-gate(FALLBACK-ONLY + closeInfixMatch + distinguishing-segment word-boundary)로
		// LSGF-001 보존: 명명되지 않은 sibling은 over-attribute하지 않는다 (AP-5).
		// @MX:NOTE: [AUTO] combined-scope prefix fallback — 이제 shared index를 조회한다
		//
		//	(이전에는 per-SPEC secondary `git log --grep=<prefix>` 서브프로세스였다).
		//
		// @MX:REASON: SPEC-V3R6-DRIFT-LEGACY-CONVENTION-001 mechanism ① (M3) — combined-scope
		//
		//	close는 per-SPEC window에 없으므로 (full-ID 미명명) scope-prefix 조회만이 도달 가능.
		//	additive-only: exact-token primary walk는 불변, fallback은 primary가
		//	completed/terminal을 못 줄 때만 fire (genuine-⑤ 보호).
		if a.status == "completed" && gitStatus != "completed" && !isTerminalStatus(gitStatus) {
			if inMemCombinedScopeClose(commits, a.specID) {
				gitStatus = "completed"
			}
		}

		// Check for drift
		drifted := a.status != gitStatus

		records = append(records, DriftRecord{
			SPECID:            a.specID,
			FrontmatterStatus: a.status,
			GitImpliedStatus:  gitStatus,
			Drifted:           drifted,
		})

		if drifted {
			driftCount++
		}
	}

	// Sort records by SPEC-ID for consistent output
	sort.Slice(records, func(i, j int) bool {
		return records[i].SPECID < records[j].SPECID
	})

	report := &DriftReport{
		Records: records,
		Count:   driftCount,
	}

	// Persist keyed on the HEAD we computed against, so an unchanged HEAD is served
	// from cache next time. Best-effort — a write failure only costs a recompute.
	saveDriftCache(baseDir, head, report)

	return report, nil
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
	// Decide default branch — prefer main; fall back to master.
	// REQ-PERF-001-A: uses per-run cache to avoid redundant git rev-parse spawns.
	branch := cachedMainBranch()

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

// specIDTailPattern은 full SPEC-ID의 trailing `-<SEGMENT>-<NNN>` 쌍을 매칭한다.
// deriveScopePrefix가 이 마지막 쌍을 strip하여 combined-scope group이 사용하는
// scope-prefix(SPEC-{PREFIX})를 얻는다.
//
// @MX:NOTE: [AUTO] scope-prefix 파생 — combined-scope close의 SPEC-{PREFIX} 추출.
// @MX:REASON: SPEC-V3R6-DRIFT-LEGACY-CONVENTION-001 mechanism ① (M3) — distinguishing
//
//	segment(+number)를 strip하여 secondary prefix-grep key를 만든다.
var specIDTailPattern = regexp.MustCompile(`-[A-Za-z0-9]+-[0-9]+$`)

// deriveScopePrefix는 full SPEC-ID에서 trailing distinguishing-segment+number를 strip하여
// combined-scope group이 명명하는 scope-prefix를 반환한다. 예: SPEC-CCSYNC-CLAUDEMD-001 →
// SPEC-CCSYNC. 마지막 `-<SEGMENT>-<NNN>` 쌍만 strip한다 (multi-segment ID도 마지막 쌍만).
func deriveScopePrefix(specID string) string {
	return specIDTailPattern.ReplaceAllString(specID, "")
}

// combinedScopeCloseGroupPattern은 combined-scope close subject의 그룹 표기
// `(SEG1 + SEG2 + ...)` 또는 `(SEG1 + SEG2 ...추가설명)`에서 첫 괄호 그룹을 캡처한다.
// 그룹 내 토큰을 `+` 및 공백/구두점으로 split하여 distinguishing-segment 후보를 얻는다.
var combinedScopeCloseGroupPattern = regexp.MustCompile(`\(([^)]*)\)`)

// distinguishingSegmentToken는 영숫자 토큰(distinguishing segment 후보)을 추출하는 패턴이다.
var distinguishingSegmentToken = regexp.MustCompile(`[A-Za-z0-9]+`)

// combinedScopeCloseMatches는 combined-scope close subject가 주어진 specID의
// sibling을 close하는지 3-gate로 판정한다 (FALLBACK-ONLY gate는 DetectDrift에서 적용).
//
//	gate (a): subject prefix가 `chore(SPEC-<PREFIX>)` / `docs(SPEC-<PREFIX>)`이고
//	          <PREFIX>가 trailing `-NNN`을 갖지 않는다 (즉 full SPEC-ID가 아닌 scope-prefix).
//	gate (b): closeInfixMatch(subject) == true (정규 close 신호).
//	gate (c): subject의 combined-scope 그룹 `(SEG1 + SEG2 ...)`을 토큰 단위로 split하여
//	          specID의 distinguishing segment와 WORD-BOUNDARY/TOKEN 정확 일치하는지 검사한다
//	          (D-NEW-1: FOO는 (FOOBAR + BAZ)에 의해 false-clear되면 안 된다).
//
// @MX:NOTE: [AUTO] combined-scope 3-gate matcher — LSGF-001 보존 (word-boundary token match).
// @MX:REASON: SPEC-V3R6-DRIFT-LEGACY-CONVENTION-001 mechanism ① (M3) — historical
//
//	combined-scope close(예: cf7d78a9c "chore(SPEC-CCSYNC): ... (CLAUDEMD + TOOLCAT)")를
//	sibling에 매핑하되, 명명되지 않은 sibling은 over-attribute하지 않는다 (gate c, AP-5).
func combinedScopeCloseMatches(subject, specID string) bool {
	lower := strings.ToLower(strings.TrimSpace(subject))

	// gate (b): close-infix가 없으면 combined-scope fallback 후보가 아니다.
	if !closeInfixMatch(lower) {
		return false
	}

	prefix := deriveScopePrefix(specID)

	// gate (a): subject가 scope-prefix(trailing -NNN 없는 SPEC-{PREFIX})를 prefix-scope로
	// 명명하는지 검사한다. subject에서 추출한 모든 SPEC-ID 토큰 중 어느 것도 full-ID가 아니면서
	// (즉 ExtractSPECIDs가 full-ID 토큰을 잡지 않으면서) scope-prefix `SPEC-<PREFIX>(`가
	// subject에 등장해야 한다.
	//   - subject가 specID full-ID 토큰을 직접 명명하면 그것은 combined-scope가 아니라
	//     per-SPEC commit이므로 primary walk가 처리한다 (여기서 false → fallback 미적용).
	if slices.Contains(ExtractSPECIDs(subject), specID) {
		return false
	}
	// scope-prefix가 `chore(spec-<prefix>)` / `docs(spec-<prefix>)` 형태로 등장하는지 확인.
	// (이미 소문자) — scope-prefix 뒤에 `)` 또는 `:`가 와야 full-ID(-NNN)가 아님이 보장된다.
	lowerPrefix := strings.ToLower(prefix)
	hasScopePrefix := strings.Contains(lower, "chore("+lowerPrefix+")") ||
		strings.Contains(lower, "chore("+lowerPrefix+":") ||
		strings.Contains(lower, "docs("+lowerPrefix+")") ||
		strings.Contains(lower, "docs("+lowerPrefix+":")
	if !hasScopePrefix {
		return false
	}

	// gate (c): distinguishing-segment word-boundary token match.
	// specID의 distinguishing segment를 추출한다 (scope-prefix 제거 후 trailing -NNN 제거).
	seg := distinguishingSegment(specID, prefix)
	if seg == "" {
		return false
	}
	segLower := strings.ToLower(seg)

	// subject의 모든 괄호 그룹 `(...)`에서 토큰을 추출하여 정확 일치를 검사한다.
	// 첫 그룹은 `chore(SPEC-PREFIX)` prefix의 괄호일 수 있으므로 (예: `(spec-abc)`),
	// 모든 그룹을 스캔한다. 정확 토큰 일치이므로 prefix 그룹의 토큰(`spec`/`abc`)은
	// distinguishing segment(`foo`)와 충돌하지 않는다.
	allGroups := combinedScopeCloseGroupPattern.FindAllStringSubmatch(lower, -1)
	for _, g := range allGroups {
		if len(g) < 2 {
			continue
		}
		tokens := distinguishingSegmentToken.FindAllString(g[1], -1)
		if slices.Contains(tokens, segLower) {
			return true
		}
	}
	return false
}

// distinguishingSegment는 specID에서 scope-prefix를 제거하고 trailing -NNN을 제거하여
// distinguishing segment를 반환한다. 예: (SPEC-CCSYNC-CLAUDEMD-001, SPEC-CCSYNC) → CLAUDEMD.
// prefix가 hyphen-delimited boundary로 specID의 prefix가 아니면 "" 반환 (collision guard).
func distinguishingSegment(specID, prefix string) string {
	// hyphen-delimited prefix boundary: specID는 `SPEC-<PREFIX>-`로 시작해야 한다.
	if !strings.HasPrefix(specID, prefix+"-") {
		return ""
	}
	rest := strings.TrimPrefix(specID, prefix+"-") // 예: CLAUDEMD-001
	// trailing -NNN 제거 → CLAUDEMD
	if idx := strings.LastIndex(rest, "-"); idx > 0 {
		return rest[:idx]
	}
	return rest
}

// The former resolveCombinedScopeClose ran a SECOND `git log --grep=<scope-prefix>`
// subprocess per completed-but-unclosed SPEC. It is superseded by
// inMemCombinedScopeClose (drift_index.go), which answers the identical question
// against the shared single-pass index — same candidate window, same 3-gate matcher,
// zero extra subprocesses.

// DriftCount is a convenience function that returns only the drift count.
// Served from the HEAD-SHA cache while HEAD is unchanged.
func DriftCount(baseDir string) (int, error) {
	report, err := DetectDrift(baseDir)
	if err != nil {
		return 0, err
	}
	return report.Count, nil
}

// DriftCountFresh returns the drift count while ignoring the HEAD-SHA cache
// (REQ-SSP-006a — the `moai spec drift --count --no-cache` path).
func DriftCountFresh(baseDir string) (int, error) {
	report, err := DetectDriftFresh(baseDir)
	if err != nil {
		return 0, err
	}
	return report.Count, nil
}

// driftWorkFn is the seam DriftCountCtx runs under a deadline. Production points
// it at DriftCount; a test overrides it to prove the time-box abandons a slow,
// context-IGNORING worker. It is a package var (not a parameter) so the public
// DriftCountCtx signature stays a drop-in ctx-aware sibling of DriftCount.
var driftWorkFn = DriftCount

// DriftCountCtx computes the drift count under a caller-supplied deadline
// (SPEC-SESSIONSTART-PERF-001 REQ-SSP-015). The drift computation itself is a
// bounded, synchronous git + in-memory pass with no context awareness, so the
// time-box cannot be enforced cooperatively — it is enforced HERE by running the
// work in a goroutine and racing it against ctx.Done(). On deadline exceed the
// caller receives ctx.Err() immediately while the abandoned goroutine finishes
// into the buffered channel (never blocking on send → no leak).
//
// This is the safety net that guarantees the session-start advisory check can
// never block the critical path unboundedly, even on a pathological repository.
//
// @MX:WARN: [AUTO] launches a goroutine whose result is discarded on timeout.
// @MX:REASON: SPEC-SESSIONSTART-PERF-001 REQ-SSP-015 — the buffered (cap-1)
//
//	channel is load-bearing: an unbuffered channel would leak the worker
//	goroutine on every timeout, since the abandoned send would block forever.
func DriftCountCtx(ctx context.Context, baseDir string) (int, error) {
	type result struct {
		count int
		err   error
	}

	// Capture the seam ONCE, synchronously on the caller's goroutine, so the
	// spawned goroutine closes over a local copy. On a timeout the worker
	// goroutine is deliberately abandoned (still running); reading the mutable
	// driftWorkFn package var from inside it would race with a test's cleanup
	// reassignment. The local capture removes that shared access entirely.
	work := driftWorkFn

	ch := make(chan result, 1) // buffered: the worker never blocks on send.
	go func() {
		c, e := work(baseDir)
		ch <- result{count: c, err: e}
	}()

	select {
	case r := <-ch:
		return r.count, r.err
	case <-ctx.Done():
		return 0, ctx.Err()
	}
}
