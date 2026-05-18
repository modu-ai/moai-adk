---
id: SPEC-V3R4-CI-FASTTRACK-001
title: "CI/CD Fast Track for Single-Developer Workflow (Path-Filter + Review Bot Consolidation)"
version: "0.1.0"
status: completed
created: 2026-05-17
updated: 2026-05-17
author: manager-spec
priority: P1
phase: "v3.0.0"
module: ".github/workflows"
lifecycle: spec-anchored
tags: "ci, cd, github-actions, paths-filter, review-bot, single-developer, productivity"
---

# Scenarios — SPEC-V3R4-CI-FASTTRACK-001

Wave-by-Wave test plans. This is a single-Wave SPEC; all scenarios apply to the
run-PR validation gate.

## 0. Wave 0 — Skip-Marker Proof-of-Concept (sandbox, NOT merged)

**Scenario**: Before Wave 1 implementation, validate the skip-marker pattern in
isolated sandbox PR. Confirms (a) GitHub Actions' "two jobs with identical `name`,
mutually exclusive `if:` guards" pattern works on this runner config, (b) the actual
check-run name that satisfies branch protection's `CodeQL` and `Test (ubuntu-latest)`.

**Given**:

- Plan-PR merged into main.
- Empty `test/skip-marker-poc` branch from main HEAD.

**When**:

```bash
# 0. Pre-verify the actual check-run names emitted by current main
gh api repos/modu-ai/moai-adk/commits/main/check-runs --jq '.check_runs[].name' | sort -u
# Expected: includes "Analyze (Go) (go)" and "Test (ubuntu-latest)"

# 1. Create sandbox PoC workflow + markdown-only diff (see plan.md T0 deliverable)
git checkout -b test/skip-marker-poc
cat > .github/workflows/skip-marker-poc.yml <<'EOF'
# (T0 deliverable verbatim — see plan.md T0)
EOF
echo "<!-- PoC test -->" >> README.md
git add .github/workflows/skip-marker-poc.yml README.md
git commit -m "T0: skip-marker pattern PoC"
git push -u origin test/skip-marker-poc

# 2. Open sandbox PR
gh pr create --base main --title "T0: skip-marker PoC (do not merge)" --body "Wave 0 sandbox" --draft
POC_PR=$(gh pr view --json number -q .number)
gh pr checks "$POC_PR" --watch
```

**Then**:

- `PoC Test` check on the PoC PR reports `SUCCESS` (via the skip-marker job).
- Workflow run history shows `test-skip-marker` job ran, `test` job is in `skipped`.
- design.md AD-002 § "Wave 0 PoC Result" is updated with:
  - PASS / FAIL verdict (binary)
  - Observed satisfying check-run name (one of `Analyze (Go) (go)` / `CodeQL` etc.)
  - Reference link to sandbox PR (preserve forever as evidence)

**Failure mode**: If `PoC Test` reports `PENDING-forever` or `FAILURE`, the skip-marker
pattern is unsupported on this runner / branch protection config. Abort Wave 1, escalate
to alternative design (e.g., GitHub Actions `workflow_call` gating). community
discussion #13690 may have updated guidance.

**Cleanup**:

```bash
gh pr close "$POC_PR" --delete-branch
git push origin --delete test/skip-marker-poc
git checkout main && git branch -D test/skip-marker-poc
```

**Cross-reference**: design.md AD-002 § Wave 0 PoC Result, plan.md T0, plan.md §6
Implementation Sequence "Wave 0".

## 1. Docs-Only PR Fast Track (Wave 1 — post run-PR merge)

**Scenario**: After Wave 1 run-PR merge into main, a docs-only PR triggers paths-filter,
skip-marker job emits success for all 3 OS matrix slots AND for the CodeQL check, the
actual Go test job + analyze job are skipped, and the PR is mergeable.

**Given**:

- Wave 0 PASS recorded in design.md AD-002.
- Run-PR `feat/SPEC-V3R4-CI-FASTTRACK-001-fasttrack-impl` is merged into main.
- Working tree on `main` checkout at the post-merge HEAD.

**When**:

```bash
# Create a test branch with a single markdown edit
git checkout -b test/docs-fast-track
echo "<!-- AC-CIFT-001 test edit -->" >> README.md
git commit -am "test: docs-only PR for fast track verification"
git push -u origin test/docs-fast-track

# Open PR
gh pr create --title "test: docs-only PR for fast track" --body "AC-CIFT-001 verification" --base main
TEST_PR=$(gh pr view --json number -q .number)

# Wait for CI to settle
gh pr checks "$TEST_PR" --watch
```

**Then**:

- All 3 `Test (<os>)` checks report SUCCESS (via T1 skip-marker).
- The CodeQL required check (whichever name Wave 0 recorded as satisfying — typically
  `CodeQL` workflow-name match) reports SUCCESS (via T2 `analyze-skip-marker`).
- Detail page shows the skip-marker job ran (workflow run history); the actual `test`
  job and `analyze` job are in `skipped` state.
- PR is mergeable (`gh pr view --json mergeable -q .mergeable` == `MERGEABLE`).

**Failure mode**: If any `Test (<os>)` check is FAILURE or PENDING-forever, the
paths-filter pattern or skip-marker name template is wrong. Inspect `ci.yml` `detect`
job outputs and the `test-skip-marker` job's `name:` field. Re-apply T1. If CodeQL
check is PENDING-forever, the codeql.yml `analyze-skip-marker` `name:` does not match
the branch-protection contexts entry — re-apply T2.0 verification.

**Cleanup**: `gh pr close "$TEST_PR" --delete-branch`

**Cross-reference**: AC-CIFT-001, AC-CIFT-002.

## 2. Mixed PR (Docs + Go Change) — Full Matrix Runs

**Scenario**: PR with BOTH markdown and Go code changes triggers full 3-OS matrix.

**Given**:

- Run-PR merged.
- Working tree on main.

**When**:

```bash
git checkout -b test/mixed-pr
echo "<!-- mixed test -->" >> README.md
# Trivial Go comment edit
sed -i.bak 's|^// Package cli|// Package cli (mixed test)|' internal/cli/init.go && rm internal/cli/init.go.bak
git commit -am "test: mixed PR — docs + Go"
git push -u origin test/mixed-pr
gh pr create --title "test: mixed PR matrix gate" --base main
TEST_PR=$(gh pr view --json number -q .number)
gh pr checks "$TEST_PR" --watch
```

**Then**:

- All 3 `Test (<os>)` checks report SUCCESS (via the *real* test job, not skip-marker).
- Workflow run history shows the actual `test` job ran on all 3 OS.
- Wait time matches non-docs baseline (5-6 min) — intentional behavior.

**Failure mode**: If skip-marker job runs (instead of real test), paths-filter
`go_code` pattern missed the Go change. Inspect filter pattern. Re-apply T1.

**Cleanup**: `gh pr close "$TEST_PR" --delete-branch && git checkout main && git branch -D test/mixed-pr`

**Cross-reference**: AC-CIFT-001 (EC-1).

## 3. 5 Review Workflows Deleted

**Scenario**: After run-PR merge, `git log` shows 5 file deletions; `gh api` confirms
workflows de-registered; new PRs trigger only `claude-code-review.yml`.

**Given**:

- Run-PR merged.

**When**:

```bash
# Local file check
for wf in codex-review gemini-review glm-review llm-panel claude-code-review.optional; do
  test ! -f ".github/workflows/${wf}.yml" || echo "FAIL: ${wf}.yml still exists"
done

# Remote workflow registry check
gh api repos/modu-ai/moai-adk/actions/workflows \
  --jq '.workflows[] | select(.path | test("(codex-review|gemini-review|glm-review|llm-panel|claude-code-review\\.optional)\\.yml$")) | .path' \
  | wc -l
```

**Then**:

- Local check loop emits no FAIL line.
- Remote registry returns 0 matching paths.

**Failure mode**: If any file still exists locally OR remotely, T3 was incomplete.
Re-run `git rm <files>`, commit, push, and verify GitHub workflows registry sync
(may take ~30 sec to propagate).

**Cross-reference**: AC-CIFT-003.

## 4. lefthook Pre-Push Enforces Preflight

**Scenario**: After lefthook install, attempted push runs `make preflight`. On lint
or test failure, push is blocked.

**Given**:

- Run-PR merged.
- `brew install lefthook` already done OR newly installed.

**When**:

```bash
# Install lefthook hooks
lefthook install

# Verify hook wrapper
cat .git/hooks/pre-push | head -3

# Make a deliberately broken commit
git checkout -b test/preflight-gate
cat > /tmp/broken.go <<'EOF'
package broken
func main() { undefined_function() }  // intentional compile error
EOF
mv /tmp/broken.go internal/broken_test_file.go
git add internal/broken_test_file.go
git commit -m "test: broken code for preflight gate"

# Attempt push (should FAIL pre-push hook)
git push origin test/preflight-gate
echo "exit=$?"
```

**Then**:

- `lefthook install` succeeds and updates `.git/hooks/pre-push`.
- `git push` exits non-zero (push blocked by pre-push hook).
- `lefthook` output shows `preflight` command failed (likely at `lint-fast` step).

**Bypass test**:

```bash
LEFTHOOK=0 git push origin test/preflight-gate
echo "exit=$?"
```

- exits 0 (lefthook bypassed; push succeeds with warning).

**Cleanup**:

```bash
git push origin --delete test/preflight-gate
git checkout main && git branch -D test/preflight-gate
rm internal/broken_test_file.go
```

**Failure mode**: If pre-push hook does not block on lint error, check `lefthook.yml`
`pre-push.commands.preflight.run` references `make preflight` correctly. Re-apply T5.

**Cross-reference**: AC-CIFT-005.

## 5. Release PR Multi-OS Workflow First Run

**Scenario**: Trigger release PR multi-OS workflow via `workflow_dispatch` and verify
3-OS matrix execution, OR create a release/* branch PR to test automatic trigger.

**Given**:

- Run-PR merged.
- `gh` CLI authenticated.

**When** (manual trigger):

```bash
gh workflow run release-pr-multi-os.yml
sleep 5
RUN_ID=$(gh run list --workflow=release-pr-multi-os.yml --limit=1 --json databaseId -q '.[0].databaseId')
gh run watch "$RUN_ID"
```

Or **When** (release branch PR trigger):

```bash
git checkout -b release/v2.20.0-rc1 origin/main
echo "# Release v2.20.0-rc1" >> CHANGELOG.md
git add CHANGELOG.md && git commit -m "release: v2.20.0-rc1 bump"
git push -u origin release/v2.20.0-rc1
gh pr create --base main
```

**Then**:

- `gh workflow list` includes "Release PR Multi-OS".
- Manual trigger or PR open creates a new workflow run.
- All 3 matrix jobs (ubuntu-latest, macos-latest, windows-latest) execute.
- If all matrix jobs succeed, `notify-on-failure` job does not run.
- If any matrix job fails, an issue is created with prefix "Release PR Multi-OS failure".

**Failure mode**: If workflow_dispatch is not visible in `gh workflow list`, check
`on.workflow_dispatch:` block in release-pr-multi-os.yml. Wait ~30 sec for GitHub
to sync after run-PR merge.

**Cross-reference**: AC-CIFT-006 (EC-2).

## 6. Branch Protection Alignment with §18.7

**Scenario**: GitHub branch protection API returns exactly 4 required checks, matching
the post-(B) baseline and CLAUDE.local.md §18.7 documentation.

**Given**:

- Run-PR merged.
- Orchestrator's (B) PATCH already applied (baseline before this SPEC).

**When**:

```bash
# Fetch live branch protection state
gh api repos/modu-ai/moai-adk/branches/main/protection/required_status_checks \
  --jq '.contexts | sort'

# Cross-check with §18.7 documentation
awk '/^### §18\.7/,/^### §18\.8/' CLAUDE.local.md \
  | grep -oE '"(Lint|Test \(ubuntu-latest\)|Build \(linux/amd64\)|CodeQL)"' \
  | sort -u
```

**Then**:

- API response contexts (sorted): `["Build (linux/amd64)", "CodeQL", "Lint", "Test
  (ubuntu-latest)"]` (exactly 4).
- §18.7 grep matches exactly the same 4 names.

**Failure mode**: If API returns 6 contexts, the (B) PATCH was reverted somewhere.
Re-apply via `gh api -X PATCH ... required_status_checks --input -`. If §18.7 grep
does not match, T7 was incomplete — re-apply.

**Cross-reference**: AC-CIFT-007.

## 7. Lessons #19 Capture Verified

**Scenario**: lessons.md contains entry #19 with all 5 protocol sections (Category,
Incorrect, Correct, Why, How to apply).

**Given**:

- Run-PR merged.
- T8 has appended the entry.

**When**:

```bash
LESSONS=~/.claude/projects/-Users-goos-MoAI-moai-adk-go/memory/lessons.md
test -f "$LESSONS"
grep -n "^## #19" "$LESSONS"
awk '/^## #19/,/^## #20|EOF/' "$LESSONS" \
  | grep -cE "Category|Incorrect|Correct|Why|How to apply"
awk '/^## #19/,/^## #20|EOF/' "$LESSONS" \
  | grep -cE "1인 개발|3-tier|paths-filter|release-PR"
```

**Then**:

- `lessons.md` file exists.
- `## #19` heading exists exactly once.
- 5 protocol sections present (count ≥ 5).
- Key concept keywords present (count ≥ 4).

**Failure mode**: If any grep returns short, re-apply T8 with verbatim REQ-CIFT-008 body.

**Cross-reference**: AC-CIFT-008.

## 8. Spec Lint Clean

**Scenario**: The 5 plan-phase artifacts pass strict spec lint with zero findings.

**Given**: Plan-PR (this branch) artifacts in place.

**When**:

```bash
moai spec lint --strict
```

**Then**:

- Output contains `✓ No findings` (or equivalent zero-finding indicator) for
  `SPEC-V3R4-CI-FASTTRACK-001`.
- Exit code 0.

**Failure mode**: Inspect findings. Common causes:

- Frontmatter snake_case alias (`created_at:` instead of `created:`); fix per
  `spec-frontmatter-schema.md` SSOT.
- Missing required frontmatter field (12-field canonical schema).
- Status enum mismatch.
- EARS keyword absent in requirements.

Re-run after fix.

**Cross-reference**: All ACs implicitly depend on this gate before plan-PR merge.

## 9. Out of Scope (mirror)

Items explicitly NOT tested by these scenarios (deferred):

- Self-hosted runner cost / setup
- Test sharding implementation
- All-OS removal from main matrix (Linux-only build)
- claude-code-review.yml internals (trigger / payload / auth)
- review bot reactivation (codex / gemini / glm)
- macOS-latest required check restoration
- GitHub Actions → external CI provider migration
- v2.20.0-rc1 release tag publication
- claude.yml or review-quality-gate.yml auto-removal
- CI cache tuning (Go module cache TTL)

These are documented in spec.md §3 (Non-Goals) and plan.md §7 (Out of Scope) as DROP
items; their absence here is intentional.
