---
id: SPEC-V3R4-CI-FASTTRACK-001
title: "CI/CD Fast Track for Single-Developer Workflow (Path-Filter + Review Bot Consolidation)"
version: "0.1.0"
status: draft
created: 2026-05-17
updated: 2026-05-17
author: manager-spec
priority: P1
phase: "v3.0.0"
module: ".github/workflows"
lifecycle: spec-anchored
tags: "ci, cd, github-actions, paths-filter, review-bot, single-developer, productivity"
---

# Acceptance Criteria — SPEC-V3R4-CI-FASTTRACK-001

All criteria are binary PASS / FAIL. Run-PR cannot merge until every AC reports PASS
(except AC-CIFT-008 which is informational performance gate).

## AC-CIFT-001 — Docs-Only PR Fast Path

**Given** run-PR is merged into main and a new test PR is created with ONLY markdown
changes (e.g., a 1-line edit to `README.md`).

**When** the test PR triggers CI:

```bash
gh pr create --title "Test docs-only path filter" --body "AC-CIFT-001 manual gate"
TEST_PR=$(gh pr view --json number -q .number)
gh pr checks "$TEST_PR" --watch
```

**Then**:

- The `Test (ubuntu-latest)` check MUST report SUCCESS (via skip-marker job).
- The `Test (macos-latest)` and `Test (windows-latest)` checks MUST also report
  SUCCESS (via skip-marker matrix).
- The actual Go test job MUST be skipped (`if:` guard) — verify via
  `gh pr checks "$TEST_PR" --json` checking for skipped state.
- The PR MUST be mergeable (all required checks satisfied).

**Verification commands**:

```bash
gh pr checks "$TEST_PR" --json name,state -q \
  '.[] | select(.name | startswith("Test (")) | {name, state}'
```

Expected output: 3 entries with state `SUCCESS` (the skip-marker pass through).

**Binary**: PASS if all 3 matrix checks SUCCESS AND PR mergeable; FAIL otherwise.

**Implements**: REQ-CIFT-001.

## AC-CIFT-002 — CodeQL Paths-Ignore

**Given** the same docs-only test PR from AC-CIFT-001.

**When** the CodeQL workflow trigger is examined:

```bash
gh run list --workflow=codeql.yml --branch="$(gh pr view "$TEST_PR" --json headRefName -q .headRefName)"
```

**Then**:

- Either NO CodeQL workflow run is created for this PR (paths-ignore consumed the
  trigger), OR
- If `paths-ignore` does not satisfy branch protection alone, the CodeQL check status
  on the PR is `SUCCESS` (per GitHub's documented "skipped → success" behavior for
  paths-ignore in some cases).

**Verification**:

```bash
gh pr checks "$TEST_PR" --json name,state -q \
  '.[] | select(.name == "CodeQL") | .state'
```

Expected: `SUCCESS` or absent (no row).

**Binary**: PASS if CodeQL check is SUCCESS or absent; FAIL if CodeQL check is RUNNING
indefinitely or FAILURE.

**Implements**: REQ-CIFT-002.

## AC-CIFT-003 — 5 Review Workflows Removed

**Given** run-PR is merged into main.

**When** the verification command runs:

```bash
for wf in codex-review gemini-review glm-review llm-panel claude-code-review.optional; do
  test ! -f ".github/workflows/${wf}.yml" || echo "FAIL: ${wf}.yml still exists"
done
ls .github/workflows/ | grep -E "^(codex-review|gemini-review|glm-review|llm-panel|claude-code-review\.optional)\.yml$" | wc -l
```

**Then**:

- All 5 files MUST be absent (the `test ! -f` loop emits no FAIL line).
- The `grep | wc -l` MUST return exactly `0`.

Cross-check via GitHub API (post-merge):

```bash
gh api repos/modu-ai/moai-adk/actions/workflows --jq \
  '.workflows[] | select(.path | test("(codex-review|gemini-review|glm-review|llm-panel|claude-code-review\\.optional)\\.yml$")) | .path' \
  | wc -l
```

Expected: `0` (workflows are de-registered).

**Binary**: PASS if all 3 checks return 0; FAIL otherwise.

**Implements**: REQ-CIFT-003.

## AC-CIFT-004 — `claude-code-review.yml` + Audit-Preserved Workflows Survive

**Given** run-PR is merged.

**When** the verification command runs:

```bash
for wf in claude-code-review claude review-quality-gate; do
  test -f ".github/workflows/${wf}.yml" || echo "FAIL: ${wf}.yml missing"
done
```

**Then**:

- `claude-code-review.yml` MUST exist (PRESERVE per user approval).
- `claude.yml` MUST exist (PRESERVE per AD-004 audit: codex-independent `@claude` trigger).
- `review-quality-gate.yml` MUST exist (PRESERVE per AD-004 audit: codex-independent
  Claude Code Review check_run severity parser).

Additional codex-independence verification:

```bash
grep -nE 'codex|gemini|glm[^a-z]' \
  .github/workflows/claude.yml \
  .github/workflows/review-quality-gate.yml \
  .github/workflows/claude-code-review.yml \
  || echo "no matches (codex-independent confirmed)"
```

Expected: 0 matches (or only matches in unrelated context like comments).

**Binary**: PASS if 3 files exist AND no codex/gemini/glm CLI invocations in their
bodies; FAIL otherwise.

**Implements**: REQ-CIFT-003, REQ-CIFT-004.

## AC-CIFT-005 — lefthook + Makefile Preflight

**Given** run-PR is merged.

**When** the verification commands run:

```bash
test -f lefthook.yml
grep -E "^pre-push:" lefthook.yml
grep -E "^\.PHONY:.*preflight|^preflight:" Makefile
make preflight  # must exit 0 on a clean main checkout
```

**Then**:

- `lefthook.yml` exists at repo root.
- `lefthook.yml` contains a `pre-push:` block.
- `Makefile` declares `preflight` as a phony target.
- `make preflight` exits 0 (assuming clean main checkout with no lint errors and tests
  passing under `-race -short`).

Additionally, the lefthook install path SHOULD be documented:

```bash
grep -nE "brew install lefthook|lefthook install" \
  README.md CONTRIBUTING.md CLAUDE.local.md 2>/dev/null | head -3
```

Expected: at least 1 reference (location chosen by run-phase, typically README onboarding).

**Binary**: PASS if all 4 file checks pass AND `make preflight` exits 0 AND lefthook
install reference exists; FAIL otherwise.

**Implements**: REQ-CIFT-005.

## AC-CIFT-006 — Nightly Full-Matrix Workflow Created

**Given** run-PR is merged.

**When** the verification commands run:

```bash
test -f .github/workflows/nightly-full-matrix.yml
grep -E '^\s+- cron: "0 3 \* \* \*"' .github/workflows/nightly-full-matrix.yml
grep -E 'workflow_dispatch:' .github/workflows/nightly-full-matrix.yml
grep -E "tags: \['v\*'\]|tags:\s*\['v\\*'\]" .github/workflows/nightly-full-matrix.yml
grep -E 'os: \[ubuntu-latest, macos-latest, windows-latest\]' .github/workflows/nightly-full-matrix.yml
grep -E 'actions/github-script@v7' .github/workflows/nightly-full-matrix.yml
grep -E 'createComment\b|createIssue\b' .github/workflows/nightly-full-matrix.yml
gh workflow list | grep -E "Nightly Full Matrix"
```

**Then**:

- File exists.
- cron `0 3 * * *` configured.
- `workflow_dispatch` trigger present.
- tag-push trigger present.
- 3-OS matrix configured.
- `actions/github-script@v7` step present (for issue creation).
- Issue creation step present (createIssue or createComment for dedup).
- GitHub registers the workflow.

Optional (informational, NOT blocking AC): trigger one manual run via
`gh workflow run nightly-full-matrix.yml` and observe the matrix succeeds.

**Binary**: PASS if all 8 grep / test checks pass; FAIL otherwise.

**Implements**: REQ-CIFT-006.

## AC-CIFT-007 — CLAUDE.local.md §18.7 Doctrine Sync

**Given** run-PR is merged.

**When** the verification commands run:

```bash
# §18.7 must contain exactly 4 required check names in the gh api PATCH payload
awk '/^### §18\.7/,/^### §18\.8/' CLAUDE.local.md | grep -oE '"(Lint|Test \(ubuntu-latest\)|Build \(linux/amd64\)|CodeQL|Test \(macos-latest\)|Test \(windows-latest\))"' | sort -u

# 3-tier philosophy must be documented
awk '/^### §18\.7/,/^### §18\.8/' CLAUDE.local.md | grep -cE "Tier 1|Tier 2|Tier 3|3-tier"

# Cross-reference to this SPEC
grep -n "SPEC-V3R4-CI-FASTTRACK-001" CLAUDE.local.md

# (B) rationale documented
awk '/^### §18\.7/,/^### §18\.8/' CLAUDE.local.md | grep -cE "1인 개발|2026-05-17"
```

**Then**:

- The 4 required check names (`Lint`, `Test (ubuntu-latest)`, `Build (linux/amd64)`,
  `CodeQL`) MUST appear in §18.7 gh api PATCH payload.
- The 2 removed check names (`Test (macos-latest)`, `Test (windows-latest)`) MUST NOT
  appear in the required_status_checks contexts array (they may appear elsewhere as
  documentation of removal).
- The 3-tier philosophy section grep count MUST be ≥ 4 (Tier 1, Tier 2, Tier 3, and at
  least one usage of "3-tier").
- The SPEC ID cross-reference grep MUST return ≥ 1 match.
- The (B) rationale grep MUST return ≥ 1 match.

**Binary**: PASS if all 5 conditions satisfied; FAIL otherwise.

**Implements**: REQ-CIFT-007.

## AC-CIFT-008 — Lessons #18 Capture

**Given** run-PR is merged AND T8 has appended the entry.

**When** the verification commands run:

```bash
LESSONS=~/.claude/projects/-Users-goos-MoAI-moai-adk-go/memory/lessons.md
grep -n "^## #18" "$LESSONS"
awk '/^## #18/,/^## #19|^---|EOF/' "$LESSONS" | grep -cE "Category|Incorrect|Correct|Why|How to apply"
awk '/^## #18/,/^## #19|EOF/' "$LESSONS" | grep -cE "1인 개발|3-tier|paths-filter|nightly"
```

**Then**:

- The `## #18` heading exists exactly once.
- The 5 protocol sections (Category, Incorrect, Correct, Why, How to apply) all appear
  (count ≥ 5).
- Key concepts (1인 개발, 3-tier, paths-filter, nightly) appear in the body (count ≥ 4).

**Binary**: PASS if all 3 grep checks satisfied; FAIL otherwise.

**Implements**: REQ-CIFT-008.

## AC-CIFT-009 — Go Test Suite Still Passes (No Accidental Code Touch)

**Given** run-PR is merged (CI / build / doctrine layer changes only).

**When** the test suite runs locally and in CI:

```bash
go test -race ./...
```

**Then**:

- The command MUST exit 0 with no NEW test failures attributable to this SPEC.
- Pre-existing flaky tests (e.g., `TestSupervisor_NonZeroExit` ETXTBSY race per
  CLAUDE.local.md §18.11) MAY fail on first attempt; retry once.
- CI Test (ubuntu-latest) check on the run-PR itself MUST be SUCCESS.

**Binary**: PASS if `go test -race ./...` exits 0 (after ≤ 1 retry of pre-existing flaky
tests) AND run-PR's `Test (ubuntu-latest)` is SUCCESS; FAIL otherwise.

**Implements**: CC-3 (no production code changes).

## Performance Gates

### PG-1 — PR-Level Wait Reduction (Qualitative)

**Threshold**: For docs-only PRs (CI changes ONLY markdown / `.moai/specs/` / `.moai/docs/` /
`.claude/rules/` files), per-PR wait time SHALL reduce by ≥ 80% compared to pre-SPEC
baseline.

**Measurement**:

- Pre-SPEC baseline (current main): docs-only PR triggers full 3-OS matrix (~5-6 min wait).
- Post-SPEC: docs-only PR triggers skip-marker matrix (< 30 sec wait).
- Expected reduction: ~90%.

**Verification** (manual, post-merge):

```bash
# Create test docs-only PR after run-PR merge
# Observe CI completion time via gh pr checks --watch
```

**Binary**: ADVISORY (informational). If reduction < 50%, surface as comment on run-PR
review but do not block merge.

### PG-2 — Workflow Count Reduction

**Threshold**: `.github/workflows/*.yml` file count SHALL decrease by net 4 (delete 5,
add 1 = -4).

**Verification**:

```bash
PRE=$(git show "$(gh pr view --json baseRefName -q .baseRefName)":.github/workflows/ | wc -l)
POST=$(ls .github/workflows/*.yml | wc -l)
echo "delta=$((POST - PRE))"
```

Expected: `delta=-4` (or equivalent measurement: pre=18, post=14).

**Binary**: PASS if `delta == -4`; FAIL otherwise.

## Edge Cases

### EC-1 — Mixed PR (Docs + Go Change)

**Behavior**: PR modifies BOTH markdown (e.g., README) AND Go code (e.g.,
`internal/cli/init.go`). paths-filter detect job evaluates both filters and sets
`go_code = true`. The full test matrix runs (skip-marker job's `if:` evaluates to
false).

**Expected**: All 3 OS test jobs (ubuntu/macos/windows) run; skip-marker job(s) do not
run. Wait time matches non-docs baseline (5-6 min) — intentional, not regression.

**Verification**: Post-merge, create a mixed PR and observe `gh pr checks` rolls full
matrix.

### EC-2 — Release Tag Pushed

**Behavior**: User pushes `v3.0.0-rc1` tag. Two workflows trigger:

1. `release.yml` (existing GoReleaser): builds + publishes binaries
2. `nightly-full-matrix.yml` (new T6): runs full 3-OS test matrix against the tag

**Expected**: Both workflows run in parallel. Tag-push trigger of nightly serves as
release-gate verification. If nightly fails, GoReleaser may still publish (separate
workflows) — release blockage requires manual user action based on issue created by
nightly's `notify-on-failure` job (filtered by `github.event_name == 'schedule'` —
tag push does NOT auto-create issue; manual review).

**Verification**: Post-tag, observe both workflows in `gh run list`.

### EC-3 — lefthook Missing on Developer Machine

**Behavior**: Developer has never run `brew install lefthook`. `.git/hooks/pre-push` is
not lefthook wrapper. `git push` proceeds without pre-flight check.

**Expected**: Graceful degradation. No error, no block. Push succeeds. CI catches any
regression (Tier 1 ubuntu-latest required check + Tier 3 nightly).

**Verification**: User remove lefthook hooks (`lefthook uninstall`) and verifies push
still works.

### EC-4 — claude-code-review.yml ANTHROPIC_API_KEY Secret Missing

**Behavior**: Repo secret rotation removed `ANTHROPIC_API_KEY` accidentally. New PR
opens. `claude-code-review.yml` action fails with `401 Unauthorized`.

**Expected**: claude-code-review check fails. Since claude-code-review is NOT in branch
protection's 4 required checks (it's an informational review bot, not a required gate
per (B) decision), the PR remains mergeable. User receives review failure notification
via PR comment.

**Verification**: Post-merge, simulate by temporarily revoking the secret in a test
fork (NOT production repo).

## Definition of Done

The SPEC is **DONE** when ALL of the following are true:

1. Plan-PR is merged into main with status `MERGED` per `gh pr view`.
2. Run-PR is merged into main with status `MERGED`.
3. Sync-PR is merged into main with status `MERGED`.
4. All 9 acceptance criteria (AC-CIFT-001 through AC-CIFT-009) report PASS.
5. PG-2 (workflow count reduction) PASS (delta == -4); PG-1 (PR wait reduction)
   advisory observed ≥ 80%.
6. `CLAUDE.local.md` §18.7 reflects the 4-item required check list + 3-tier philosophy.
7. `lessons.md` #18 entry captures the 1-developer CI 3-tier pattern.
8. `nightly-full-matrix.yml` registered in `gh workflow list` AND first manual
   `workflow_dispatch` run succeeds.
9. `lefthook` install path documented in onboarding docs (`README.md` or equivalent).
10. CHANGELOG.md (top of file) carries an entry describing the CI fast-track rollout.
11. The frontmatter `status` field in spec.md is `completed` (post-sync transition).
