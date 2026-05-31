// Package spec — atomic close orchestrator for `moai spec close` CLI.
//
// SPEC-V3R6-LIFECYCLE-SYNC-GATE-001 design.md §B.1 + §A.2.
//
// M1 scope (this milestone): public API surface + precondition matrix validation
// + dry-run path + structured error reporting + lock acquisition. M3 milestone
// completes the atomic-commit transaction (file staging + git commit + SHA
// backfill mechanism per design §B.1 Option D).
//
// Public contract:
//
//	closer.Close(specID, opts) → (*CloseResult, error)
//
// Result semantics:
//   - error == nil + Result.CommitSHA non-empty → atomic close succeeded
//   - error == nil + Result.NoOp == true → backfill-only on already-completed SPEC
//   - error == ErrPreconditionMissing → preconditions not met; staging untouched
//   - error == ErrSpecCloseLockHeld → another process holds the lock
//   - error == ErrDryRun → --dry-run requested; no staging side-effects
package spec

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// CloseOptions configures Close() invocations.
type CloseOptions struct {
	// BaseDir is the project root (where .moai/specs/ lives). Default ".".
	BaseDir string

	// BackfillOnly transitions only the missing fields (sync_commit_sha /
	// mx_commit_sha / status) without requiring full preconditions. Per AC-LSG-022.
	// When the SPEC is already fully completed, this becomes a no-op (exit 0,
	// 0 commits) per AC-LSG-018 (v0.1.2 reframe).
	BackfillOnly bool

	// DryRun prints the diff that would be applied and exits without staging
	// or committing. Returns ErrDryRun.
	DryRun bool

	// Force bypasses precondition checks. Reserved for emergency recovery; not
	// part of normal close flow. Use only when L60 backfill rule applies.
	Force bool

	// LogPath overrides the audit-trail log destination. When empty, the log is
	// written to <BaseDir>/.moai/logs/lifecycle-close.log per NFR-LSG-004. Tests
	// inject a t.TempDir() path so they never touch the real project log.
	LogPath string
}

// CloseResult is the structured output of Close().
type CloseResult struct {
	SpecID    string            `json:"spec_id"`
	CommitSHA string            `json:"commit_sha,omitempty"`
	// Transitions records which fields were updated and to what values.
	// Empty map indicates no-op (AC-LSG-018 fully-completed-noop fixture).
	Transitions map[string]string `json:"transitions"`
	// NoOp is true when the SPEC is already completed and no changes are needed.
	NoOp bool `json:"noop"`
	// Mode is "full-close" or "backfill-only" per AC-LSG-020 log fields.
	Mode string `json:"mode"`
	// PreconditionsFailed lists the precondition names that prevented close.
	// Populated only when error == ErrPreconditionMissing.
	PreconditionsFailed []string `json:"preconditions_failed,omitempty"`
	// Result is the in-memory outcome: "success" | "failure" | "noop".
	// NOTE: the on-disk AC-LSG-020 audit log uses a narrower {success, failure}
	// enum — "noop" maps to "success" there via lifecycleCloseLogResult, because
	// a no-op close IS a successful close (see lifecycleCloseLogEntry).
	Result string `json:"result"`
	// DurationMs is total wall-clock time of the close invocation.
	DurationMs int64 `json:"duration_ms"`
	// AuditedAt is the close timestamp (RFC3339) per AC-LSG-020 log schema.
	AuditedAt time.Time `json:"audited_at"`
}

// Sentinel errors for the close path. The CLI layer maps these to exit codes:
//   - ErrPreconditionMissing → exit 1 (AC-LSG-006, AC-LSG-014)
//   - ErrSpecCloseLockHeld   → exit 1 (AC-LSG-010, AC-LSG-021)
//   - ErrDryRun              → exit 0 (informational, --dry-run requested)
var (
	// ErrPreconditionMissing is returned when one or more close preconditions
	// are not met (e.g., missing §E.5 Mx section in progress.md). The
	// PreconditionsFailed field on CloseResult names the specific failure.
	ErrPreconditionMissing = errors.New("close precondition not met")

	// ErrDryRun is returned when opts.DryRun is set. It signals "preview only".
	// Not a true error — CLI exits 0 and prints the would-apply diff.
	ErrDryRun = errors.New("dry-run requested; no staging performed")

	// ErrAlreadyCompleted indicates the SPEC is already at status: completed.
	// In backfill-only mode this is converted to a no-op success (NoOp=true).
	// In full close mode it is surfaced to the caller.
	ErrAlreadyCompleted = errors.New("SPEC already at status: completed")
)

// @MX:ANCHOR: [AUTO] Close는 `moai spec close`의 단일 진입점 — CLI(spec_close.go) + 단위/통합 테스트가 호출(fan_in=3)
// @MX:REASON: [AUTO] no-op 판정 불변식(--backfill-only 모드에서 spec.md status=="completed"면 mx_commit_sha 백필 형태와 무관하게 무조건 no-op, 0 commit)이 AC-LSG-018/022 진실표에 묶여 있다. 이 술어를 변경하면 5개 이미-종료 SPEC의 dogfood가 깨진다.
//
// Close orchestrates the atomic 4-phase close transition for a SPEC.
//
// M1 implementation (this milestone): precondition matrix + lock + dry-run +
// no-op detection. M3 will add the atomic git commit transaction.
//
// The lock is held only for the duration of this call (released on return).
//
// Every invocation (success / no-op / failure / dry-run) appends one JSON line
// to the audit-trail log per NFR-LSG-004 (AC-LSG-020). The log destination is
// opts.LogPath, defaulting to <BaseDir>/.moai/logs/lifecycle-close.log. The
// empty-specID guard is the sole exception: it predates result construction and
// has no SPEC to attribute the log entry to.
func Close(specID string, opts CloseOptions) (result *CloseResult, err error) {
	startedAt := time.Now()

	if specID == "" {
		return nil, fmt.Errorf("Close: empty specID")
	}

	baseDir := opts.BaseDir
	if baseDir == "" {
		baseDir = "."
	}
	// REQ-CFS-002: baseDir를 절대 경로로 정규화한다. 기본값 "." (상대 경로)에
	// 의존하면 하위 git add / git restore --staged / git commit 작업이 cmd.Dir
	// 파일시스템 타이밍에 따라 일관되지 않게 경로를 해석한다. filepath.Abs는
	// 이미 절대 경로인 입력에 대해 멱등적이므로(§D.2 edge case) 회귀가 없다.
	if absBase, absErr := filepath.Abs(baseDir); absErr == nil {
		baseDir = absBase
	}

	mode := "full-close"
	if opts.BackfillOnly {
		mode = "backfill-only"
	}

	result = &CloseResult{
		SpecID:      specID,
		Transitions: map[string]string{},
		Mode:        mode,
		AuditedAt:   startedAt.UTC(),
	}

	// Single audit-trail emission per invocation (NFR-LSG-004 / AC-LSG-020).
	// Runs on every return path below; DurationMs is finalized here so all paths
	// record consistent wall-clock time without per-return duplication. The log
	// `result` field maps the in-memory "noop" value to "success" (no-op IS a
	// success per AC-LSG-020 — the log schema's result enum is {success, failure}
	// only). Log I/O failure is non-fatal: it never overrides the close outcome.
	defer func() {
		result.DurationMs = time.Since(startedAt).Milliseconds()
		writeLifecycleCloseLog(baseDir, opts.LogPath, result)
	}()

	// AC-LSG-010 — acquire per-SPEC lock; ErrSpecCloseLockHeld on contention.
	lock, lockErr := AcquireSpecCloseLock(baseDir, specID)
	if lockErr != nil {
		result.Result = "failure"
		if IsLockHeldError(lockErr) {
			return result, lockErr
		}
		return result, fmt.Errorf("acquire lock: %w", lockErr)
	}
	defer func() { _ = lock.Release() }()

	// Read spec.md + progress.md + acceptance.md
	specDir := filepath.Join(baseDir, ".moai", "specs", specID)
	if _, statErr := os.Stat(specDir); statErr != nil {
		result.Result = "failure"
		return result, fmt.Errorf("spec directory not found: %s", specDir)
	}

	state, loadErr := loadSpecCloseState(specDir, specID)
	if loadErr != nil {
		result.Result = "failure"
		return result, fmt.Errorf("load spec state: %w", loadErr)
	}

	// Detect no-op (AC-LSG-018 / AC-LSG-022 fully-completed fixture state).
	//
	// M1/M2 remediation (Defect 1): the no-op predicate keys ONLY on
	// `spec.md status == "completed"`. A terminal-state SPEC is already closed
	// regardless of how its §E.5 mx_commit_sha was backfilled (the 5 already-
	// discharged target SPECs left mx_commit_sha empty / `null` / `(this commit)`
	// placeholder / absent). The earlier triple-AND gate additionally required
	// non-empty SyncCommitSHA AND MxCommitSHA, which let only the literal both-
	// SHA-present fixture through and caused 4/5 production SPECs to fall through
	// to the precondition matrix or compute-transitions — violating AC-LSG-018's
	// 0-commit no-op requirement.
	//
	// Truth-table safety (AC-LSG-022): the three transition fixtures
	// (Y_N_N_Y / Y_Y_N_Y / Y_Y_Y_Y_StatusDrift) all carry status: implemented,
	// so they never enter this branch — only `fully-completed-noop` (status:
	// completed) and the 5 production SPECs do.
	if state.SpecMDStatus == "completed" {
		if opts.BackfillOnly {
			// Already complete + backfill-only = no-op success path.
			result.NoOp = true
			result.Result = "noop"
			return result, nil
		}
		result.Result = "failure"
		return result, ErrAlreadyCompleted
	}

	// Precondition matrix validation per AC-LSG-006 + AC-LSG-014.
	preconditionsFailed := validatePreconditions(state, opts)
	if len(preconditionsFailed) > 0 && !opts.Force {
		result.PreconditionsFailed = preconditionsFailed
		result.Result = "failure"
		return result, fmt.Errorf("%w: %s", ErrPreconditionMissing, strings.Join(preconditionsFailed, ", "))
	}

	// Compute transitions (which fields would change)
	transitions := computeTransitions(state, opts)
	result.Transitions = transitions

	// Dry-run path — return ErrDryRun without staging
	if opts.DryRun {
		result.Result = "success"
		return result, ErrDryRun
	}

	// Atomic commit transaction (M1/M3 remediation, Defect 3).
	// Writes the computed transitions to disk, stages ONLY this SPEC's spec.md +
	// progress.md by exact path, commits, and populates result.CommitSHA. On any
	// failure, performs a full rollback so no partial staging remains (AC-LSG-014).
	sha, txErr := performAtomicClose(baseDir, specID, state)
	if txErr != nil {
		result.Result = "failure"
		return result, fmt.Errorf("atomic close transaction: %w", txErr)
	}
	result.CommitSHA = sha
	result.Result = "success"
	return result, nil
}

// performAtomicClose executes the atomic close transaction for a non-no-op,
// non-dry-run, preconditions-met SPEC. It:
//
//  1. Rewrites spec.md frontmatter status → completed.
//  2. Rewrites progress.md §E.3 status → completed (when progress.md present).
//  3. Backfills §E.2 sync_commit_sha if missing/placeholder (BACKWARD reference
//     to the existing sync commit; the `(this commit)` placeholder resolves to
//     the most recent sync commit SHA touching this SPEC).
//  4. Backfills §E.5 mx_commit_sha per the L60 atomic-backfill chicken-and-egg
//     convention (placeholder acceptable — the no-op predicate no longer
//     requires a non-empty mx_commit_sha).
//  5. Stages EXACTLY this SPEC's spec.md + progress.md (explicit paths only —
//     NEVER `git add -A` / `.` / `-u`).
//  6. Commits and returns the new commit SHA.
//
// Atomicity (AC-LSG-014): any step failure triggers full rollback — staged
// paths are unstaged (`git restore --staged`) and on-disk file contents are
// restored from the in-memory snapshots captured at load time. After rollback,
// `git status --porcelain` shows no staged changes attributable to this call.
//
// @MX:ANCHOR: [AUTO] performAtomicClose는 Close()의 유일한 commit 수행 경로 — full-close happy path가 여기로 수렴(AC-LSG-001/014)
// @MX:REASON: [AUTO] explicit-path staging 불변식(이 SPEC의 spec.md + progress.md만 stage, 절대 git add -A 금지)과 rollback 원자성(어떤 step 실패 시 staged 0)이 active multi-session race 하에서 다른 세션의 parallel-WIP 파일을 침범하지 않도록 보장한다.
func performAtomicClose(baseDir, specID string, state *closeState) (commitSHA string, err error) {
	// Resolve the on-disk (absolute) paths used for os.WriteFile rollback.
	specMDPath := state.SpecMDPath
	progressMDPath := state.ProgressMDPath

	// REQ-CFS-001: git add / git restore --staged 에 넘길 경로는 baseDir 기준
	// 상대 경로로 변환한다. 절대 경로를 cmd.Dir=baseDir 환경에서 git에 넘기면
	// 파일시스템 타이밍에 따라 일관되지 않게 해석되어 CI race-detector에서만
	// 재현되는 경합을 유발한다. filepath.Rel 실패 시(다른 볼륨 등 §D.2 edge case)
	// 절대 경로로 폴백한다 — 조용히 삼키지 않고 명시적으로 처리한다.
	specMDGitPath := relToBaseOrAbs(baseDir, specMDPath)
	var progressMDGitPath string
	if progressMDPath != "" {
		progressMDGitPath = relToBaseOrAbs(baseDir, progressMDPath)
	}

	// Snapshot on-disk contents for rollback.
	origSpec := state.SpecMDContent
	origProgress := state.ProgressMDContent

	// staged tracks the exact paths we `git add`-ed, for precise unstaging on rollback.
	var staged []string

	// rollback restores file contents and unstages anything we staged. Best-effort:
	// rollback errors are joined into the returned error but never panic.
	rollback := func(cause error) (string, error) {
		// Restore on-disk file contents from snapshots (absolute paths for os.WriteFile).
		_ = os.WriteFile(specMDPath, []byte(origSpec), 0644)
		if progressMDPath != "" {
			_ = os.WriteFile(progressMDPath, []byte(origProgress), 0644)
		}
		// REQ-CFS-003: rollback의 git restore --staged 도 staging과 동일한
		// 상대 경로 해석을 사용한다 — stage한 정확히 그 경로만 unstage된다.
		if len(staged) > 0 {
			args := append([]string{"restore", "--staged"}, staged...)
			_ = runGitInDir(baseDir, args...)
		}
		return "", cause
	}

	// Step 1 — spec.md frontmatter status → completed.
	newSpec := rewriteSpecStatusCompleted(origSpec)
	if err := os.WriteFile(specMDPath, []byte(newSpec), 0644); err != nil {
		return rollback(fmt.Errorf("write spec.md: %w", err))
	}

	// Steps 2-4 — progress.md transitions + backfills (when progress.md present).
	if progressMDPath != "" {
		newProgress := origProgress
		// §E.3 status → completed.
		newProgress = rewriteProgressStatusCompleted(newProgress)
		// §E.2 sync_commit_sha backfill (BACKWARD reference to existing sync commit).
		// The dogfood SPEC carries the `(this commit)` placeholder here, which must
		// be resolved to the actual prior sync commit SHA.
		if needsSHABackfill(state.SyncCommitSHA) {
			resolved := resolveRecentSpecCommitSHA(baseDir, specID)
			if resolved != "" {
				newProgress = backfillProgressField(newProgress, "sync_commit_sha", resolved)
			}
		}
		// §E.5 mx_commit_sha backfill per L60 (placeholder acceptable — the close
		// commit's own SHA cannot be embedded in itself, chicken-and-egg).
		if needsSHABackfill(state.MxCommitSHA) {
			newProgress = backfillProgressField(newProgress, "mx_commit_sha", l60MxBackfillPlaceholder)
		}
		if err := os.WriteFile(progressMDPath, []byte(newProgress), 0644); err != nil {
			return rollback(fmt.Errorf("write progress.md: %w", err))
		}
	}

	// Step 5 — stage EXACTLY this SPEC's spec.md + progress.md by explicit path.
	// REQ-CFS-001: baseDir 기준 상대 경로를 git에 넘긴다(절대 경로 race 회피).
	if err := runGitInDir(baseDir, "add", specMDGitPath); err != nil {
		return rollback(fmt.Errorf("git add spec.md: %w", err))
	}
	staged = append(staged, specMDGitPath)
	if progressMDGitPath != "" {
		if err := runGitInDir(baseDir, "add", progressMDGitPath); err != nil {
			return rollback(fmt.Errorf("git add progress.md: %w", err))
		}
		staged = append(staged, progressMDGitPath)
	}

	// Step 6 — commit and capture the new SHA.
	subject := fmt.Sprintf("chore(%s): Mx-phase audit-ready signal + 4-phase close", specID)
	body := "Authored-By-Agent: orchestrator-direct"
	if err := runGitInDir(baseDir, "commit", "-m", subject, "-m", body); err != nil {
		return rollback(fmt.Errorf("git commit: %w", err))
	}

	sha, shaErr := gitRevParseHead(baseDir)
	if shaErr != nil {
		// Commit succeeded but we cannot read the SHA — surface as failure (do NOT
		// roll back a successful commit; report the read error instead).
		return "", fmt.Errorf("read commit SHA after commit: %w", shaErr)
	}
	return sha, nil
}

// relToBaseOrAbs는 abs(절대 파일 경로)를 baseDir 기준 상대 경로로 변환한다.
// git add / git restore --staged 에 cmd.Dir=baseDir 와 함께 넘길 경로 해석을
// 결정론적으로 만들기 위함이다(REQ-CFS-001/003). filepath.Rel이 실패하거나
// (다른 볼륨 등 §D.2 edge case) ".." 로 baseDir 밖을 벗어나는 경로를 산출하면
// 원래의 절대 경로로 폴백한다 — 조용히 삼키지 않고 안전한 입력만 상대화한다.
func relToBaseOrAbs(baseDir, abs string) string {
	rel, err := filepath.Rel(baseDir, abs)
	if err != nil || strings.HasPrefix(rel, "..") {
		return abs
	}
	return rel
}

// l60MxBackfillPlaceholder is the L60 atomic-backfill placeholder for the §E.5
// mx_commit_sha field. The close commit's own SHA cannot be embedded in itself
// (chicken-and-egg, plan.md §B1); this placeholder follows the established
// backward-compat convention. The remediated no-op predicate no longer requires
// a non-empty mx_commit_sha, so a placeholder is acceptable.
const l60MxBackfillPlaceholder = "(this commit)"

// needsSHABackfill reports whether a §E.2/§E.5 commit-SHA field value requires
// backfill. A value needs backfill when it is empty OR a `(this commit)`-style
// placeholder. `extractProgressField` (via cleanFieldValue) strips quotes but
// does NOT normalize `(this commit)` to empty, so the dogfood SPEC's
// `sync_commit_sha: "(this commit)"` arrives here as the literal placeholder.
//
// This predicate is transaction-local and does NOT alter cleanFieldValue, whose
// `(this commit)`-as-value behavior the era classification heuristics rely on.
func needsSHABackfill(value string) bool {
	v := strings.ToLower(strings.TrimSpace(value))
	v = strings.Trim(v, `"'`+"`")
	switch v {
	case "", "(this commit)", "(pending)", "<pending>":
		return true
	}
	return false
}

// runGitInDir runs a git subcommand in dir, returning a wrapped error on failure.
func runGitInDir(dir string, args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git %s: %w (output: %s)", strings.Join(args, " "), err, strings.TrimSpace(string(out)))
	}
	return nil
}

// gitRevParseHead returns the current HEAD commit SHA for the repo at dir.
func gitRevParseHead(dir string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git rev-parse HEAD: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

// resolveRecentSpecCommitSHA returns the most recent commit SHA whose message
// references the SPEC ID (a BACKWARD reference to the existing sync commit used
// to backfill an empty §E.2 sync_commit_sha). Returns "" when no such commit
// exists or git is unreachable — the caller then leaves the placeholder.
func resolveRecentSpecCommitSHA(dir, specID string) string {
	cmd := exec.Command("git", "log", "-1", "--format=%H", fmt.Sprintf("--grep=%s", specID))
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// rewriteSpecStatusCompleted rewrites the spec.md frontmatter `status:` line to
// `completed`. Idempotent: when already completed, the content is returned
// unchanged. Matches the canonical specStatusPattern anchor.
func rewriteSpecStatusCompleted(content string) string {
	return specStatusLinePattern.ReplaceAllString(content, "status: completed")
}

// specStatusLinePattern matches the frontmatter `status:` line (line-anchored,
// no leading indent — frontmatter fields are column-0). Distinct from
// specStatusPattern (which captures the value) — this one rewrites the whole line.
var specStatusLinePattern = regexp.MustCompile(`(?m)^status:[ \t]*.*$`)

// progressStatusLinePattern matches a `status:` line in progress.md body (allows
// leading whitespace or markdown-list prefix). Used to flip §E.3 status to completed.
var progressStatusLinePattern = regexp.MustCompile(`(?m)^([ \t]*(?:[-*][ \t]*)?` + "`?" + `status` + "`?" + `[ \t]*:[ \t]*).*$`)

// rewriteProgressStatusCompleted rewrites the first progress.md `status:` line to
// `completed`, preserving any leading whitespace / list-marker prefix. When no
// status line exists, the content is returned unchanged (caller already gated on
// progress.md presence; a missing §E.3 status line is tolerated).
func rewriteProgressStatusCompleted(content string) string {
	loc := progressStatusLinePattern.FindStringSubmatchIndex(content)
	if loc == nil {
		return content
	}
	// loc[2]:loc[3] is the captured prefix (indent + `status:` + spacing).
	prefix := content[loc[2]:loc[3]]
	return content[:loc[0]] + prefix + "completed" + content[loc[1]:]
}

// backfillProgressField rewrites an existing `field:` line in progress.md to the
// given value, or appends `field: value` when the field line is absent. Empty /
// placeholder existing values (null, none, "", `(this commit)`) are overwritten.
//
// The rewritten line preserves the original line's leading indent + optional
// markdown-list marker, but normalizes the `key:` separator to exactly one space
// before the value (so `mx_commit_sha:` with no trailing space yields
// `mx_commit_sha: <value>`, not `mx_commit_sha:<value>`).
func backfillProgressField(content, field, value string) string {
	// Capture group 1: leading indent + optional list marker + backtick-wrapped key.
	pattern := regexp.MustCompile(`(?m)^([ \t]*(?:[-*][ \t]*)?` + "`?" + regexp.QuoteMeta(field) + "`?" + `)[ \t]*:[ \t]*.*$`)
	if loc := pattern.FindStringSubmatchIndex(content); loc != nil {
		keyPrefix := content[loc[2]:loc[3]] // indent + (list marker) + key (no colon)
		return content[:loc[0]] + keyPrefix + ": " + value + content[loc[1]:]
	}
	// Field line absent — append under the existing body (best-effort; rare path).
	if !strings.HasSuffix(content, "\n") {
		content += "\n"
	}
	return content + field + ": " + value + "\n"
}

// lifecycleCloseLogEntry is the on-disk NFR-LSG-004 audit-trail schema. One
// such entry is appended (as a single JSON line) per Close() invocation.
//
// Distinct from CloseResult: the log `result` enum is {success, failure} ONLY
// (AC-LSG-020). A no-op close — whose in-memory CloseResult.Result is "noop" —
// serializes here as result: "success" with transitions: {} (empty object).
// This reconciliation lets BOTH AC-LSG-018's jq filter
// (`.result == "success" and .transitions == {}`) AND AC-LSG-020's
// `.result == "success"` filter match the 5 no-op dogfood closes.
type lifecycleCloseLogEntry struct {
	Timestamp   string            `json:"timestamp"`   // RFC3339
	SpecID      string            `json:"spec_id"`
	Mode        string            `json:"mode"`        // full-close | backfill-only
	Transitions map[string]string `json:"transitions"` // changed fields; {} when none
	CommitSHA   string            `json:"commit_sha"`
	Result      string            `json:"result"`      // success | failure
	DurationMs  int64             `json:"duration_ms"`
}

// lifecycleCloseLogResult maps the in-memory CloseResult.Result to the log
// schema's {success, failure} enum. "noop" maps to "success" (a no-op IS a
// successful close); any non-"failure" value is treated as success.
func lifecycleCloseLogResult(in string) string {
	if in == "failure" {
		return "failure"
	}
	return "success"
}

// writeLifecycleCloseLog appends one JSON line describing the close outcome to
// the audit-trail log. logPath overrides the destination; when empty it
// defaults to <baseDir>/.moai/logs/lifecycle-close.log per NFR-LSG-004. The
// parent directory is created if absent. All I/O errors are swallowed — the
// audit trail is best-effort and MUST NOT alter the close outcome.
func writeLifecycleCloseLog(baseDir, logPath string, result *CloseResult) {
	if result == nil {
		return
	}

	path := logPath
	if path == "" {
		path = filepath.Join(baseDir, ".moai", "logs", "lifecycle-close.log")
	}

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return
	}

	transitions := result.Transitions
	if transitions == nil {
		transitions = map[string]string{}
	}

	entry := lifecycleCloseLogEntry{
		Timestamp:   result.AuditedAt.UTC().Format(time.RFC3339),
		SpecID:      result.SpecID,
		Mode:        result.Mode,
		Transitions: transitions,
		CommitSHA:   result.CommitSHA,
		Result:      lifecycleCloseLogResult(result.Result),
		DurationMs:  result.DurationMs,
	}

	line, err := json.Marshal(entry)
	if err != nil {
		return
	}

	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return
	}
	defer func() { _ = f.Close() }()
	_, _ = f.Write(append(line, '\n'))
}

// closeState captures the parsed state of spec.md + progress.md needed for
// precondition validation AND the atomic close transaction (M3 remediation).
type closeState struct {
	SpecMDStatus     string // current spec.md frontmatter status
	HasSyncSection   bool   // §E.2 section present in progress.md
	HasMxSection     bool   // §E.5 section present in progress.md
	SyncCommitSHA    string // extracted §E.2 sync_commit_sha
	MxCommitSHA      string // extracted §E.5 mx_commit_sha
	ProgressMDStatus string // progress.md §E.3 status field
	ACAllPass        bool   // all MUST-PASS acceptance criteria PASS
	HasPassWithDebt  bool   // any AC marked PASS-WITH-DEBT (genuine verdict, not descriptive text)

	// Transaction inputs (M3): raw file paths + contents captured at load time so
	// the atomic transaction can rewrite them and roll back from in-memory snapshots.
	SpecMDPath        string // absolute path to spec.md
	SpecMDContent     string // raw spec.md bytes (pre-transition)
	ProgressMDPath    string // absolute path to progress.md ("" when absent)
	ProgressMDContent string // raw progress.md bytes ("" when absent)
}

// loadSpecCloseState reads spec.md + progress.md + acceptance.md and populates
// closeState. Missing optional files are tolerated (corresponding bool fields
// set to false).
func loadSpecCloseState(specDir, specID string) (*closeState, error) {
	state := &closeState{}

	// spec.md
	specMDPath := filepath.Join(specDir, "spec.md")
	specContent, err := os.ReadFile(specMDPath)
	if err != nil {
		return nil, fmt.Errorf("read spec.md: %w", err)
	}
	state.SpecMDPath = specMDPath
	state.SpecMDContent = string(specContent)
	if m := specStatusPattern.FindStringSubmatch(string(specContent)); len(m) > 1 {
		state.SpecMDStatus = strings.TrimSpace(m[1])
	}

	// progress.md (optional — V2.x SPECs lack this)
	progressMDPath := filepath.Join(specDir, "progress.md")
	if progressContent, perr := os.ReadFile(progressMDPath); perr == nil {
		body := string(progressContent)
		state.ProgressMDPath = progressMDPath
		state.ProgressMDContent = body
		state.HasSyncSection = hasProgressMarker(body, "§E.2")
		state.HasMxSection = hasProgressMarker(body, "§E.5")
		state.SyncCommitSHA = extractProgressField(body, "sync_commit_sha")
		state.MxCommitSHA = extractProgressField(body, "mx_commit_sha")
		// progress.md §E.3 status is read as `status:` line within §E.3 block.
		// For M1 we accept that if a top-level `status:` line exists, that is the value.
		// Refinement: M3 may need a per-section parser.
		state.ProgressMDStatus = extractProgressField(body, "status")
	}

	// acceptance.md AC PASS check (M3: precise marker detection)
	acMDPath := filepath.Join(specDir, "acceptance.md")
	if acContent, aerr := os.ReadFile(acMDPath); aerr == nil {
		acBody := string(acContent)
		// Defect 4 remediation: detect ONLY a genuine PASS-WITH-DEBT AC verdict
		// (table-cell verdict or bold **PASS-WITH-DEBT** marker), NOT the
		// descriptive precondition-definition phrase "no PASS-WITH-DEBT".
		state.HasPassWithDebt = hasGenuinePassWithDebtVerdict(acBody)
		// M3: treat presence of FAIL verdict markers as failure.
		hasFail := strings.Contains(strings.ToUpper(acBody), "**FAIL**") ||
			strings.Contains(strings.ToUpper(acBody), "| FAIL |") ||
			strings.Contains(strings.ToUpper(acBody), "| FAILED |")
		state.ACAllPass = !hasFail
	} else {
		// Missing acceptance.md is acceptable for Tier S SPECs (LEAN workflow).
		state.ACAllPass = true
	}

	return state, nil
}

// passWithDebtTableCell matches a genuine PASS-WITH-DEBT verdict in a markdown
// AC table cell. The cell MUST START with PASS-WITH-DEBT (after leading
// whitespace only): `| PASS-WITH-DEBT ... |`. Anchoring the verdict to the cell
// START distinguishes a real verdict column from a HISTORY / change-log cell that
// merely NARRATES a plan-auditor score mid-sentence
// (e.g. "| ... | iter-3 ... per plan-auditor iter-2 PASS-WITH-DEBT 0.873 ... |").
//
// Defect 4 remediation 보완: the prior pattern `\|[^|\n]*\bPASS-WITH-DEBT\b[^|\n]*\|`
// matched ANY cell containing the substring, false-positiving on HISTORY-cell
// narrations across ≥4 V3R6 SPECs (SESSION-LEGACY-COVERAGE-001,
// LOCAL-NAMESPACE-CONSOLIDATION-001, WORKFLOW-PLAN-GEARS-ALIGN-001,
// WORKFLOW-ORCHESTRATION-FIX-001), each of which has zero genuine run-phase debt.
// The bold form `| **PASS-WITH-DEBT** |` is intentionally NOT matched here (the
// cell starts with `**`) — it is covered by passWithDebtBoldVerdict instead.
var passWithDebtTableCell = regexp.MustCompile(`(?i)\|\s*PASS-WITH-DEBT\b[^|\n]*\|`)

// passWithDebtBoldVerdict matches a bold **PASS-WITH-DEBT** verdict marker
// (used in narrative AC verdict lines, e.g., "verdict: **PASS-WITH-DEBT** ...").
var passWithDebtBoldVerdict = regexp.MustCompile(`(?i)\*\*\s*PASS-WITH-DEBT\s*\*\*`)

// hasGenuinePassWithDebtVerdict reports whether acceptance.md contains a real
// PASS-WITH-DEBT AC verdict, as distinct from the descriptive precondition
// phrase "no PASS-WITH-DEBT" that merely DEFINES the concept (AC-LSG-014 text).
//
// Defect 4 (M1/M3 remediation): the prior naive substring match
// (strings.Contains(upper, "PASS-WITH-DEBT")) false-positived on
// acceptance.md's own AC-LSG-014 descriptive text
// "(sync section / mx section / AC PASS / no PASS-WITH-DEBT)", causing this
// SPEC's own --dry-run to falsely report a precondition failure when it has
// ZERO real debt ACs.
//
// Detection grounds on two genuine verdict forms ONLY:
//  1. a markdown table cell carrying the PASS-WITH-DEBT verdict, OR
//  2. a bold **PASS-WITH-DEBT** verdict marker.
//
// A bare "no PASS-WITH-DEBT" / "PASS-WITH-DEBT" appearing in flowing prose
// (precondition definition) matches NEITHER pattern and is correctly ignored.
func hasGenuinePassWithDebtVerdict(acBody string) bool {
	return passWithDebtTableCell.MatchString(acBody) ||
		passWithDebtBoldVerdict.MatchString(acBody)
}

// validatePreconditions checks the 4-phase precondition matrix per AC-LSG-006.
// Returns a slice of failing precondition names (empty slice = all pass).
//
// Backfill-only mode relaxes the spec.md status requirement (allows implemented
// or in-progress to backfill); other preconditions still apply.
func validatePreconditions(state *closeState, opts CloseOptions) []string {
	var failed []string

	// Precondition 1: §E.2 sync section present
	if !state.HasSyncSection {
		failed = append(failed, "missing §E.2 sync-phase audit-ready signal in progress.md")
	}

	// Precondition 2: §E.5 mx section present
	if !state.HasMxSection {
		failed = append(failed, "missing §E.5 Mx-phase audit-ready signal in progress.md")
	}

	// Precondition 3: all MUST-PASS AC PASS
	if !state.ACAllPass {
		failed = append(failed, "one or more acceptance criteria are not PASS")
	}

	// Precondition 4: no PASS-WITH-DEBT
	if state.HasPassWithDebt && !opts.Force {
		failed = append(failed, "PASS-WITH-DEBT marker present; use --force to override")
	}

	// In backfill-only mode the spec.md status requirement is relaxed
	// (the whole point of backfill is to bring an `implemented`/`in-progress`
	// SPEC up to `completed`). In full-close mode the spec.md status must be
	// `implemented` to proceed.
	if !opts.BackfillOnly {
		if state.SpecMDStatus != "implemented" && state.SpecMDStatus != "completed" {
			failed = append(failed,
				fmt.Sprintf("spec.md status is %q; full close requires status=implemented "+
					"(use --backfill-only for in-progress SPECs)", state.SpecMDStatus))
		}
	}

	return failed
}

// computeTransitions returns the field changes that the atomic close would apply.
// The returned map describes intent, not actual applied changes (M3 applies them).
func computeTransitions(state *closeState, opts CloseOptions) map[string]string {
	transitions := map[string]string{}

	if state.SpecMDStatus != "completed" {
		transitions["spec.md:frontmatter.status"] = "completed"
	}
	if state.ProgressMDStatus != "completed" {
		transitions["progress.md:§E.3.status"] = "completed"
	}
	if needsSHABackfill(state.SyncCommitSHA) {
		transitions["progress.md:§E.2.sync_commit_sha"] = "<derived-from-recent-sync-commit>"
	}
	if needsSHABackfill(state.MxCommitSHA) {
		transitions["progress.md:§E.5.mx_commit_sha"] = "<derived-from-recent-mx-commit>"
	}

	return transitions
}
