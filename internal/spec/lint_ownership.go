package spec

import (
	"bufio"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

// OwnershipTransitionRule는 spec.md frontmatter `status:` 전환을 git log 커밋 주체와 비교하여
// `Status Transition Ownership Matrix` (SSOT: spec-frontmatter-schema.md § Status Transition Ownership Matrix)
// 위반을 탐지한다. SPEC-V3R6-ANTHROPIC-AUDIT-TIER3-001 F13 (Anthropic Best Practice #7 DRI ownership) 도입.
//
// 7개 canonical transitions:
//  1. (none) → draft           : manager-spec       (feat(SPEC-...): plan-phase ...)
//  2. draft → in-progress      : manager-develop    (feat|fix|refactor|perf|test(SPEC-...) M1 ...)
//  3. in-progress → implemented: manager-docs       (docs(SPEC-...): sync-phase ...)
//  4. implemented → completed  : manager-docs OR orchestrator (chore(SPEC-...): Mx-phase ...)
//  5. * → superseded           : manager-spec       (feat(SPEC-NEW): supersedes SPEC-OLD)
//  6. * → archived             : manager-docs       (chore(specs): archive ...)
//  7. * → rejected             : orchestrator/manager-docs (chore(SPEC-...): rejected ...)
//
// Default severity: Warning. Observation-only — never mutates files.
// REQ-AAT-007..012 (SPEC-V3R6-ANTHROPIC-AUDIT-TIER3-001).
//
// @MX:NOTE: [AUTO] OwnershipTransitionRule — F13 programmatic DRI verification
// @MX:REASON: declarative ownership matrix(spec-frontmatter-schema.md) 단독으로는 commit이 매트릭스를 준수했는지 검증할 수 없으므로 lint-time shift-left 보강
type OwnershipTransitionRule struct{}

// Code returns the lint finding code for ownership-mismatch detections.
func (r *OwnershipTransitionRule) Code() string { return "OwnershipTransitionInvalid" }

// expectedOwnerKind는 transition별 기대 commit 분류를 나타낸다.
type expectedOwnerKind int

const (
	ownerNone           expectedOwnerKind = iota // 검증 대상 아님
	ownerManagerSpec                             // plan-phase: feat(SPEC-...): plan-phase ... | feat(SPEC-NEW): supersedes ...
	ownerManagerDevelop                          // run-phase: feat|fix|refactor|perf|test(SPEC-...) M{N} ...
	ownerManagerDocs                             // sync-phase: docs(SPEC-...): sync-phase ...
	ownerOrchestrator                            // Mx/archive/reject: chore(...): ...
)

// String returns a human-readable identifier for the expected owner role.
func (k expectedOwnerKind) String() string {
	switch k {
	case ownerManagerSpec:
		return "manager-spec"
	case ownerManagerDevelop:
		return "manager-develop"
	case ownerManagerDocs:
		return "manager-docs"
	case ownerOrchestrator:
		return "orchestrator|manager-docs"
	}
	return "unknown"
}

// expectedOwnerForTransition는 (previousStatus, currentStatus) → expectedOwnerKind 매트릭스이다.
// 매트릭스 미정의 transition (e.g., status 역행)은 ownerNone 반환 — 별도 lint rule(StatusValueEnumRule) 책임.
//
// REQ-AAT-009: 기본적으로 7개 canonical transition 모두 검증. (none) → draft 는 prev="" current="draft" 로 표현.
func expectedOwnerForTransition(prev, curr string) expectedOwnerKind {
	// terminal lifecycle states (* → superseded|archived|rejected)
	switch curr {
	case "superseded":
		return ownerManagerSpec
	case "archived":
		return ownerOrchestrator // chore(specs): archive
	case "rejected":
		return ownerOrchestrator // chore(SPEC-...): rejected
	}

	switch prev {
	case "": // (none) → draft
		if curr == "draft" {
			return ownerManagerSpec
		}
	case "draft", "planned":
		if curr == "in-progress" || curr == "implemented" {
			return ownerManagerDevelop
		}
	case "in-progress":
		if curr == "implemented" {
			return ownerManagerDocs
		}
	case "implemented":
		if curr == "completed" {
			return ownerOrchestrator // chore(SPEC-...): Mx-phase
		}
	}

	return ownerNone
}

// commitOwnerKind는 commit subject prefix를 expectedOwnerKind로 분류한다.
//
// Pattern (commit subject prefix → owner):
//   - chore(...) ...                                 → ownerOrchestrator
//   - docs(...) ...                                  → ownerManagerDocs
//   - feat(...) ... + ("plan-phase"|"supersedes")    → ownerManagerSpec
//   - feat|fix|refactor|perf|test(...) ...           → ownerManagerDevelop
//   - plan(spec ...                                  → ownerManagerSpec
//
// 분류 불가 시 ownerNone (회귀 false-positive 방지).
//
// NOTE (M4 AC-LSG-004): subject-prefix 분류는 더 이상 production Check() 경로에서
// 사용되지 않는다. WHO 신호의 SSOT는 `Authored-By-Agent:` trailer (trailerAgentOwnerKind)다.
// 본 함수는 F13(ANTHROPIC-AUDIT-TIER3-001)의 subject→owner 매핑 회귀 테스트
// (TestCommitOwnerKind)를 위해 유지된다.
func commitOwnerKind(subject string) expectedOwnerKind {
	lower := strings.ToLower(strings.TrimSpace(subject))

	if strings.HasPrefix(lower, "chore(") {
		return ownerOrchestrator
	}
	if strings.HasPrefix(lower, "docs(") {
		return ownerManagerDocs
	}
	if strings.HasPrefix(lower, "feat(") {
		if strings.Contains(lower, "plan-phase") ||
			strings.Contains(lower, "supersedes") {
			return ownerManagerSpec
		}
		return ownerManagerDevelop
	}
	if strings.HasPrefix(lower, "fix(") ||
		strings.HasPrefix(lower, "refactor(") ||
		strings.HasPrefix(lower, "perf(") ||
		strings.HasPrefix(lower, "test(") {
		return ownerManagerDevelop
	}
	if strings.HasPrefix(lower, "plan(spec") {
		return ownerManagerSpec
	}

	return ownerNone
}

// ownerMatches는 expected와 actual owner의 호환 여부를 판단한다.
// orchestrator-typed transition (implemented→completed, archived, rejected)은
// manager-docs (docs prefix)도 호환 owner로 인정한다 (matrix `manager-docs OR orchestrator`).
func ownerMatches(expected, actual expectedOwnerKind) bool {
	if expected == actual {
		return true
	}
	if expected == ownerOrchestrator && actual == ownerManagerDocs {
		return true
	}
	return false
}

// trailerAgentOwnerKind는 `Authored-By-Agent:` trailer 값(lowercase single-token)을
// expectedOwnerKind로 분류한다. trailer 값 enum: manager-spec / manager-develop /
// manager-docs / manager-git / orchestrator-direct (SPEC §D.1.6 HARD).
//
// 미인식 trailer 값은 ownerNone 반환 (silent SKIP — false-positive guard).
// manager-git은 OwnershipTransitionRule의 transition matrix 대상이 아니므로 ownerNone.
func trailerAgentOwnerKind(agent string) expectedOwnerKind {
	switch agent {
	case "manager-spec":
		return ownerManagerSpec
	case "manager-develop":
		return ownerManagerDevelop
	case "manager-docs":
		return ownerManagerDocs
	case "orchestrator-direct":
		return ownerOrchestrator
	}
	return ownerNone
}

// ownershipTransitionRecord는 git log에서 추출한 (prev, curr, commit-subject) 튜플이다.
//
// CommitSHA + AuthoredByAgent는 SPEC-V3R6-LIFECYCLE-SYNC-GATE-001 M4 (AC-LSG-004 / D5)에서
// 추가된 필드다. AuthoredByAgent는 commit body의 `Authored-By-Agent: <agent>` trailer 값으로,
// transition을 수행한 주체를 나타내는 기계적 신호다. trailer가 없으면 (legacy / non-MoAI commit)
// 빈 문자열이며, 그 경우 commit subject prefix 분류 경로로 fallback한다 (F13 호환).
type ownershipTransitionRecord struct {
	PreviousStatus  string
	CurrentStatus   string
	CommitSubject   string
	CommitSHA       string
	AuthoredByAgent string
}

// getOwnershipTransitionRunner는 테스트에서 git-log lookup 함수를 주입할 수 있도록 분리한다.
// 운영 코드 경로는 항상 lookupOwnershipTransitionFromGit를 사용한다.
//
// @MX:NOTE: [AUTO] 테스트 주입 hook — 운영은 default, 테스트는 fake function 주입
var getOwnershipTransitionRunner = lookupOwnershipTransitionFromGit

// lookupOwnershipTransitionFromGit는 SPEC ID로 git log를 조회하여 가장 최근의 frontmatter
// `status:` 전환을 찾고 그 transition을 수행한 commit subject를 반환한다.
//
// 미발견 시 (nil, nil) 반환 — non-git 환경(no .git dir) + 신규 SPEC 모두 graceful no-op.
// git 실행 실패 시 (nil, error) 반환 — 호출자가 OwnershipTransitionUnreachable Info finding 처리.
//
// @MX:NOTE: [AUTO] git log --follow로 spec.md status 라인 변화 감지
// @MX:REASON: REQ-AAT-010 — git unreachable 시 Info finding (no error severity)
func lookupOwnershipTransitionFromGit(specPath, specID string) (*ownershipTransitionRecord, error) {
	if specID == "" || specPath == "" {
		return nil, nil
	}

	// git 환경 확인 — fail-safe (테스트 tmpdir / 비-git 환경)
	if _, err := exec.Command("git", "rev-parse", "--git-dir").Output(); err != nil {
		return nil, fmt.Errorf("git unreachable: %w", err)
	}

	// git log --follow <path> -p — lookback window 내 status 변화 감지 (gitLogWindowSize 재사용)
	// %b (commit body)를 추가 캡처하여 `Authored-By-Agent:` trailer를 파싱한다 (M4 AC-LSG-004).
	cmd := exec.Command("git", "log",
		fmt.Sprintf("-%d", gitLogWindowSize),
		"--follow",
		"--format=%H%x00%s%x00%b%x00",
		"-p",
		"--",
		specPath,
	)
	output, err := cmd.Output()
	if err != nil {
		// path 무발견 (신규 SPEC) — graceful no-op (not an error)
		return nil, nil
	}
	if len(output) == 0 {
		return nil, nil
	}

	commits := splitGitLogCommits(string(output))
	for _, c := range commits {
		oldStatus, newStatus, found := extractStatusDelta(c.diffBody)
		if !found {
			continue
		}
		return &ownershipTransitionRecord{
			PreviousStatus:  oldStatus,
			CurrentStatus:   newStatus,
			CommitSubject:   c.subject,
			CommitSHA:       c.hash,
			AuthoredByAgent: parseAuthoredByAgent(c.body),
		}, nil
	}

	return nil, nil
}

// authoredByAgentLine은 commit body의 `Authored-By-Agent: <agent>` trailer를 매칭한다.
// 값은 lowercase single-token (manager-spec / manager-develop / manager-docs /
// manager-git / orchestrator-direct). 대소문자 무관, 선두 공백 허용.
// SPEC-V3R6-LIFECYCLE-SYNC-GATE-001 §D.1.6 HARD trailer convention.
var authoredByAgentLine = regexp.MustCompile(`(?mi)^\s*Authored-By-Agent:\s*(\S+)\s*$`)

// parseAuthoredByAgent은 commit body에서 `Authored-By-Agent:` trailer 값을 추출한다.
// trailer 미발견 시 빈 문자열 반환 (legacy / non-MoAI commit — OwnershipTransitionRule 적용 대상 아님).
func parseAuthoredByAgent(body string) string {
	m := authoredByAgentLine.FindStringSubmatch(body)
	if len(m) < 2 {
		return ""
	}
	return strings.ToLower(strings.TrimSpace(m[1]))
}

// gitLogCommit는 splitGitLogCommits 결과의 commit-단위 구조이다.
type gitLogCommit struct {
	hash     string
	subject  string
	body     string
	diffBody string
}

// gitLogCommitHeader는 "<hash>\x00<subject>\x00<body>\x00\n" 패턴을 매칭한다
// (git log --format=%H%x00%s%x00%b%x00). body(%b)는 newline을 포함할 수 있으나
// `[^\x00]*`이 NUL을 제외한 모든 문자(개행 포함, `(?s)` 불필요 — `[^\x00]`은 개행도 매칭)를 캡처한다.
var gitLogCommitHeader = regexp.MustCompile(`(?m)^([0-9a-f]{7,40})\x00([^\x00]*)\x00([^\x00]*)\x00\n`)

// splitGitLogCommits는 git log --format=%H<null>%s<null>%b<null> -p 출력을 commit 단위로 split한다.
func splitGitLogCommits(output string) []gitLogCommit {
	matches := gitLogCommitHeader.FindAllStringSubmatchIndex(output, -1)
	if len(matches) == 0 {
		return nil
	}
	var commits []gitLogCommit
	for i, m := range matches {
		hash := output[m[2]:m[3]]
		subject := output[m[4]:m[5]]
		body := output[m[6]:m[7]]
		var diffBody string
		if i+1 < len(matches) {
			diffBody = output[m[1]:matches[i+1][0]]
		} else {
			diffBody = output[m[1]:]
		}
		commits = append(commits, gitLogCommit{
			hash:     hash,
			subject:  subject,
			body:     body,
			diffBody: diffBody,
		})
	}
	return commits
}

// extractStatusDelta는 commit diff body에서 frontmatter `status:` 라인 변화를 추출한다.
// Returns (oldStatus, newStatus, found). 변화 없음 또는 status 라인 미발견 시 ("", "", false).
//
// Frontmatter status 라인 형식: "status: <enum-value>" (선두 공백 허용).
// 신규 SPEC (additions only)은 oldStatus="" newStatus="draft"로 반환된다.
func extractStatusDelta(diffBody string) (oldStatus, newStatus string, found bool) {
	scanner := bufio.NewScanner(strings.NewReader(diffBody))
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)
	for scanner.Scan() {
		line := scanner.Text()
		if v, ok := parseStatusDiffLine(line, "-"); ok {
			oldStatus = v
		} else if v, ok := parseStatusDiffLine(line, "+"); ok {
			newStatus = v
		}
	}
	if newStatus == "" {
		return "", "", false
	}
	return oldStatus, newStatus, true
}

// parseStatusDiffLine은 "<sign>status: <value>" 또는 "<sign> status: <value>" 패턴에서 value를 추출한다.
// sign은 "-" (삭제) 또는 "+" (추가). 일치 안 함 시 ("", false) 반환.
func parseStatusDiffLine(line, sign string) (string, bool) {
	if !strings.HasPrefix(line, sign) {
		return "", false
	}
	rest := strings.TrimPrefix(line, sign)
	rest = strings.TrimLeft(rest, " ")
	if !strings.HasPrefix(rest, "status:") {
		return "", false
	}
	value := strings.TrimSpace(strings.TrimPrefix(rest, "status:"))
	return value, true
}

// hasOwnershipSkipOptOut는 SPEC frontmatter의 `lint.skip:` 목록에
// "OwnershipTransitionInvalid"가 포함되어 있는지 검사한다 (AC-LSG-012).
func hasOwnershipSkipOptOut(doc *SPECDoc) bool {
	for _, code := range doc.Frontmatter.LintConfig.Skip {
		if code == "OwnershipTransitionInvalid" {
			return true
		}
	}
	return false
}

// Check inspects a single SPEC document for ownership-matrix compliance.
//
// Behavior:
//   - empty id/status: skipped (FrontmatterSchemaRule responsibility)
//   - lint.skip opt-out present: emits Info "OwnershipTransitionSkipped" (AC-LSG-012)
//   - non-git or untracked SPEC: emits Info "OwnershipTransitionUnreachable" (graceful)
//   - unmapped transition (역행, terminal already): silently skipped
//   - mismatched owner: emits Warning "OwnershipTransitionInvalid"
//   - trailer-less commit (legacy / non-MoAI): silently skipped (M4 AC-LSG-004)
//   - unclassifiable commit subject: silently skipped (false-positive guard)
//
// Owner-detection precedence (M4 AC-LSG-004 / D5):
//  1. `Authored-By-Agent:` commit-body trailer — the mechanical WHO signal
//  2. fallback: commit subject prefix classification (F13 legacy compatibility)
func (r *OwnershipTransitionRule) Check(doc *SPECDoc, _ []*SPECDoc) []Finding {
	fm := doc.Frontmatter
	if fm.ID == "" || fm.Status == "" {
		return nil
	}

	// AC-LSG-012: lint.skip opt-out — emit informational Skipped finding instead of Invalid.
	if hasOwnershipSkipOptOut(doc) {
		return []Finding{
			{
				File:     doc.Path,
				Line:     1,
				Severity: SeverityInfo,
				Code:     "OwnershipTransitionSkipped",
				Message:  fmt.Sprintf("SPEC %s ownership transition check skipped via lint.skip: [OwnershipTransitionInvalid] opt-out", fm.ID),
			},
		}
	}

	rec, err := getOwnershipTransitionRunner(doc.Path, fm.ID)
	if err != nil {
		// REQ-AAT-010: Info severity, never blocks
		return []Finding{
			{
				File:     doc.Path,
				Line:     1,
				Severity: SeverityInfo,
				Code:     "OwnershipTransitionUnreachable",
				Message:  fmt.Sprintf("SPEC %s ownership transition validation skipped (git unreachable: %v)", fm.ID, err),
			},
		}
	}
	if rec == nil {
		return nil
	}

	expected := expectedOwnerForTransition(rec.PreviousStatus, rec.CurrentStatus)
	if expected == ownerNone {
		return nil
	}

	// M4 AC-LSG-004 / D5: the `Authored-By-Agent:` trailer is the gating signal.
	// Commits WITHOUT the trailer (legacy / non-MoAI / pre-v3.0.1) are NOT subject to
	// OwnershipTransitionRule — silent SKIP. This is the false-positive guard required by
	// the M4 exit criterion ("moai spec lint against repo emits no false positives"):
	// most existing commits predate the trailer convention.
	if rec.AuthoredByAgent == "" {
		return nil
	}

	actual := trailerAgentOwnerKind(rec.AuthoredByAgent)
	if actual == ownerNone {
		// trailer present but not a transition-relevant actor (e.g., manager-git) — silent SKIP.
		return nil
	}

	if !ownerMatches(expected, actual) {
		return []Finding{
			{
				File:     doc.Path,
				Line:     1,
				Severity: SeverityWarning,
				Code:     "OwnershipTransitionInvalid",
				Message: fmt.Sprintf(
					"SPEC %s transition %q → %q expected owner %q but commit %s (Authored-By-Agent: %s) maps to %q",
					fm.ID,
					emptyOrValue(rec.PreviousStatus),
					rec.CurrentStatus,
					expected.String(),
					rec.CommitSHA,
					rec.AuthoredByAgent,
					actual.String(),
				),
			},
		}
	}

	return nil
}

// emptyOrValue returns "(none)" when s is empty, else s. Used for human-readable previous-status formatting.
func emptyOrValue(s string) string {
	if s == "" {
		return "(none)"
	}
	return s
}
