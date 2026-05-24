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

// ownershipTransitionRecord는 git log에서 추출한 (prev, curr, commit-subject) 튜플이다.
type ownershipTransitionRecord struct {
	PreviousStatus string
	CurrentStatus  string
	CommitSubject  string
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
	cmd := exec.Command("git", "log",
		fmt.Sprintf("-%d", gitLogWindowSize),
		"--follow",
		"--format=%H%x00%s%x00",
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
			PreviousStatus: oldStatus,
			CurrentStatus:  newStatus,
			CommitSubject:  c.subject,
		}, nil
	}

	return nil, nil
}

// gitLogCommit는 splitGitLogCommits 결과의 commit-단위 구조이다.
type gitLogCommit struct {
	hash     string
	subject  string
	diffBody string
}

// gitLogCommitHeader는 "<hash>\x00<subject>\x00\n" 패턴을 매칭한다 (git log --format=%H%x00%s%x00).
var gitLogCommitHeader = regexp.MustCompile(`(?m)^([0-9a-f]{7,40})\x00([^\x00]*)\x00\n`)

// splitGitLogCommits는 git log --format=%H<null>%s<null> -p 출력을 commit 단위로 split한다.
func splitGitLogCommits(output string) []gitLogCommit {
	matches := gitLogCommitHeader.FindAllStringSubmatchIndex(output, -1)
	if len(matches) == 0 {
		return nil
	}
	var commits []gitLogCommit
	for i, m := range matches {
		hash := output[m[2]:m[3]]
		subject := output[m[4]:m[5]]
		var diffBody string
		if i+1 < len(matches) {
			diffBody = output[m[1]:matches[i+1][0]]
		} else {
			diffBody = output[m[1]:]
		}
		commits = append(commits, gitLogCommit{
			hash:     hash,
			subject:  subject,
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

// Check inspects a single SPEC document for ownership-matrix compliance.
//
// Behavior:
//   - empty id/status: skipped (FrontmatterSchemaRule responsibility)
//   - non-git or untracked SPEC: emits Info "OwnershipTransitionUnreachable" (graceful)
//   - unmapped transition (역행, terminal already): silently skipped
//   - mismatched owner: emits Warning "OwnershipTransitionInvalid"
//   - unclassifiable commit subject: silently skipped (false-positive guard)
func (r *OwnershipTransitionRule) Check(doc *SPECDoc, _ []*SPECDoc) []Finding {
	fm := doc.Frontmatter
	if fm.ID == "" || fm.Status == "" {
		return nil
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

	actual := commitOwnerKind(rec.CommitSubject)
	if actual == ownerNone {
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
					"SPEC %s transition %q → %q expected owner %q but commit subject %q maps to %q",
					fm.ID,
					emptyOrValue(rec.PreviousStatus),
					rec.CurrentStatus,
					expected.String(),
					rec.CommitSubject,
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
