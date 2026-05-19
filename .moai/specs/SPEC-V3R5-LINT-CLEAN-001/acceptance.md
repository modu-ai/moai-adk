# Acceptance Criteria — SPEC-V3R5-LINT-CLEAN-001

Each AC leaf provides:
- **Statement** (verbatim from spec.md §4)
- **Test command** (concrete shell)
- **Pass condition** (exact)
- **Evidence location** (where the artifact lives for audit)

All commands assume cwd = project root `/Users/goos/MoAI/moai-adk-go` and that `./bin/moai` is built (`make build` produces it).

**Terminology** (per `spec.md` §2.0 Glossary): "LCLN-Phase N" (Phase 1..4) = this SPEC's internal cleanup phases; "Mega-Sprint Wave W0..W4" = v3.5.0 release-roadmap SPECs. The plan-auditor iter-1 nomenclature "W<N>-LCLN" is retired in v0.2.0.

---

## AC-LCLN-001 — Final baseline reaches W2-deferred residual after this SPEC, and 0 after Mega-Sprint W2

**Statement** (top-level): The system shall reduce `moai agent lint --strict` baseline to 0 findings on `main` HEAD after all four LCLN-Phases merge AND Mega-Sprint W2 SPEC-V3R5-CORE-SLIM-001 lifecycle COMPLETEs.

**Test command** (run after LCLN-Phase 4 merges to `main`):

```bash
git checkout main && git pull origin main
./bin/moai agent lint --strict --format=json > /tmp/lint-post-LCLN-P4.json
jq '.summary' /tmp/lint-post-LCLN-P4.json
```

**Pass condition** (post-Phase-4, pre-W2): Exit code 0 AND `.summary.total` equals 12 (the W2-deferred post-Phase-2 residual). All 12 are LR-08 findings whose drift skill OR affected agent is in the W2-retirement set.

**Final 0/0 verification** (after Mega-Sprint W2 lifecycle COMPLETE — out of this SPEC's run-phase scope, but documented as the closure condition):

```bash
./bin/moai agent lint --strict --format=json | jq '.summary.total == 0'
```

Expected output: `true`.

**Evidence location**: `.moai/state/lint-baseline-post-LCLN-P4.json` (this SPEC's contribution), and `.moai/state/lint-baseline-post-W2-coreslim.json` (cross-SPEC observation after Mega-Sprint W2 merges).

### AC-LCLN-001.1 — Baseline captured before each LCLN-Phase

**Statement**: The system shall capture and commit a baseline JSON before each LCLN-Phase PR opens.

**Test command** (run at start of each LCLN-Phase):

```bash
PHASE=P1  # or P2/P3/P4
./bin/moai agent lint --strict --format=json > .moai/state/lint-baseline-pre-LCLN-${PHASE}.json
jq '.summary' .moai/state/lint-baseline-pre-LCLN-${PHASE}.json
```

**Pass condition**: JSON file exists at `.moai/state/lint-baseline-pre-LCLN-${PHASE}.json` with valid envelope (`.version` field = `"1.0"`, `.summary.total` is integer, `.violations` is array).

**Evidence location**: `.moai/state/lint-baseline-pre-LCLN-<PHASE>.json` (committed to repo as audit trail).

### AC-LCLN-001.2 — Each LCLN-Phase reduces a contiguous subset (per-phase sum-check)

**Statement**: Each LCLN-Phase PR shall reduce its target subset of findings exactly as per plan.md §3 milestone sum-check (60 / 4 / 30 / 70 per phase).

**Test command** (run after LCLN-Phase PR merges):

```bash
PHASE=P1
./bin/moai agent lint --strict --format=json > .moai/state/lint-baseline-post-LCLN-${PHASE}.json

PRE=$(jq '.summary.total' .moai/state/lint-baseline-pre-LCLN-${PHASE}.json)
POST=$(jq '.summary.total' .moai/state/lint-baseline-post-LCLN-${PHASE}.json)

# Exact reduction targets per LCLN-Phase (from plan.md §3 milestone sum-check):
# Phase 1: 60 reduction
# Phase 2: 4 reduction
# Phase 3: 30 reduction
# Phase 4: 70 reduction
case $PHASE in
  P1) EXPECTED_REDUCTION=60 ;;
  P2) EXPECTED_REDUCTION=4 ;;
  P3) EXPECTED_REDUCTION=30 ;;
  P4) EXPECTED_REDUCTION=70 ;;
esac

ACTUAL_REDUCTION=$((PRE - POST))
TOLERANCE=3  # allows ±3 for concurrent upstream Mega-Sprint perturbation

if [ "$ACTUAL_REDUCTION" -lt "$((EXPECTED_REDUCTION - TOLERANCE))" ] || [ "$ACTUAL_REDUCTION" -gt "$((EXPECTED_REDUCTION + TOLERANCE))" ]; then
  echo "FAIL: phase ${PHASE} reduced $ACTUAL_REDUCTION, expected ${EXPECTED_REDUCTION} ± ${TOLERANCE}"
  exit 1
fi
echo "PASS: phase ${PHASE} reduced $ACTUAL_REDUCTION (target ${EXPECTED_REDUCTION} ± ${TOLERANCE})"
```

**Pass condition**: Exit code 0; output begins with `PASS:`. Reduction is within ±3 of expected target.

**Evidence location**: `.moai/state/lint-baseline-pre-LCLN-<PHASE>.json` and `.moai/state/lint-baseline-post-LCLN-<PHASE>.json`.

### AC-LCLN-001.3 — Final post-Mega-Sprint-W2 exact equality check

**Statement**: After both this SPEC and Mega-Sprint W2 complete, `moai agent lint --strict` shall report 0 total findings (exact equality, per plan-auditor iter-1 Finding 8).

**Test command** (cross-SPEC, observed only after Mega-Sprint W2 merges):

```bash
./bin/moai agent lint --strict --format=json | jq '.summary.total == 0'
```

**Pass condition**: Output is exactly `true`. No disjunctive matching — only exact JSON-boolean equality.

**Evidence location**: CI log of the final cross-SPEC post-merge run, OR `.moai/state/lint-baseline-final.json`.

---

## AC-LCLN-002 — Delta-only NEW=0 semantics (D6)

**Statement**: The system shall preserve delta-only D6 NEW=0 semantics — each LCLN-Phase PR adds 0 new findings vs the prior baseline.

### AC-LCLN-002.1 — Pre-phase baseline captured to versioned JSON

**Statement**: Pre-phase baseline shall be captured to `.moai/state/lint-baseline-pre-LCLN-P<N>.json` and committed.

**Test command** (verify file exists in the LCLN-Phase PR's diff):

```bash
PHASE=P1
git log --diff-filter=A --name-only main..HEAD | grep "lint-baseline-pre-LCLN-${PHASE}.json"
```

**Pass condition**: Output contains the baseline filename exactly once (file added in PR).

**Evidence location**: PR diff contains `.moai/state/lint-baseline-pre-LCLN-<PHASE>.json` as an added file.

### AC-LCLN-002.2 — Delta diff command emits NEW count

**Statement**: Post-phase verification command shall emit the count of findings present post-phase but absent pre-phase, keyed by `file+line+rule`.

**Test command**:

```bash
PHASE=P1
NEW_FINDINGS=$(jq -s '
  def fp: .file + ":" + (.line|tostring) + ":" + .rule;
  ([.[1].violations[] | fp] - [.[0].violations[] | fp]) | length
' .moai/state/lint-baseline-pre-LCLN-${PHASE}.json .moai/state/lint-baseline-post-LCLN-${PHASE}.json)
echo "NEW findings introduced by LCLN-${PHASE}: $NEW_FINDINGS"
```

**Pass condition**: `NEW_FINDINGS` equals `0`.

**Evidence location**: LCLN-Phase PR description includes the output line `NEW findings introduced by LCLN-P<N>: 0`.

### AC-LCLN-002.3 — NEW count = 0 enforced per LCLN-Phase merge gate

**Statement**: The NEW count vs prior baseline shall equal 0 (LCLN-Phase merge gate).

**Test command** (used as LCLN-Phase PR merge precondition):

```bash
PHASE=P1
NEW=$(jq -s '
  def fp: .file + ":" + (.line|tostring) + ":" + .rule;
  ([.[1].violations[] | fp] - [.[0].violations[] | fp]) | length
' .moai/state/lint-baseline-pre-LCLN-${PHASE}.json .moai/state/lint-baseline-post-LCLN-${PHASE}.json)

if [ "$NEW" -ne 0 ]; then
  echo "BLOCK: LCLN-${PHASE} introduces $NEW new findings"
  exit 1
fi
echo "OK: NEW=0"
```

**Pass condition**: Exit code 0 (script reaches `echo "OK: NEW=0"`).

**Evidence location**: CI log of LCLN-Phase PR's pre-merge verification step.

---

## AC-LCLN-003 — Scope discipline (no FROZEN modification, no orthogonal regression, per-rule contradiction check)

**Statement**: The system shall not modify FROZEN zones nor regress orthogonal lints nor introduce per-rule regressions during cleanup.

### AC-LCLN-003.1 — `moai spec lint --strict` exit 0

**Test command** (run before and after each LCLN-Phase):

```bash
./bin/moai spec lint --strict
echo "Exit: $?"
```

**Pass condition**: Exit code 0; output ends with `✓ No findings — all SPEC documents are valid`.

**Evidence location**: CI log of LCLN-Phase PR.

### AC-LCLN-003.2 — `golangci-lint run ./...` exit 0 (conditional on default config per C8)

**Test command**:

```bash
# Conditional: re-baseline if .golangci.yml introduced (C8)
if [ -f .golangci.yml ] || [ -f .golangci.yaml ]; then
  echo "BLOCK: .golangci.yml present — re-baseline required per C8"
  exit 1
fi
golangci-lint run ./... --timeout=10m 2>&1 | tail -3
echo "Exit: $?"
```

**Pass condition**: Exit code 0; final line is `0 issues.` or `No issues found.` (under default 6-linter set).

**Evidence location**: CI log of LCLN-Phase PR.

### AC-LCLN-003.3 — `moai workflow lint` exit 0

**Test command**:

```bash
./bin/moai workflow lint
echo "Exit: $?"
```

**Pass condition**: Exit code 0; output contains `✓ No violations found`.

**Evidence location**: CI log of LCLN-Phase PR.

### AC-LCLN-003.4 — FROZEN-zone audit (design.md §1.1 + Appendix A superset)

**Statement**: Cleanup PRs shall not touch any guarded path per design.md §1.1 + Appendix A (superset of zone-registry Frozen entries, including operational invariants).

**Test command** (Frozen Guard script, run in LCLN-Phase PR pre-merge):

```bash
FROZEN_PATTERNS=(
  ".claude/rules/moai/core/moai-constitution.md"
  ".claude/rules/moai/core/agent-common-protocol.md"
  ".claude/rules/moai/core/askuser-protocol.md"
  ".claude/rules/moai/core/zone-registry.md"
  ".claude/rules/moai/design/constitution.md"
  ".claude/rules/moai/development/coding-standards.md"
  ".claude/rules/moai/development/agent-authoring.md"
  ".claude/rules/moai/workflow/spec-workflow.md"
  ".claude/rules/moai/workflow/mx-tag-protocol.md"
  "internal/cli/agent_lint.go"
  "internal/cli/agent_lint_test.go"
  "CLAUDE.md"
)

CHANGED=$(git diff --name-only main..HEAD)

VIOLATIONS=0
for pattern in "${FROZEN_PATTERNS[@]}"; do
  if echo "$CHANGED" | grep -F "$pattern" > /dev/null; then
    echo "BLOCK: FROZEN-zone violation: $pattern modified"
    VIOLATIONS=$((VIOLATIONS + 1))
  fi
done

if [ $VIOLATIONS -gt 0 ]; then
  exit 1
fi
echo "OK: no FROZEN-zone files modified"
```

**Pass condition**: Exit code 0; output `OK: no FROZEN-zone files modified`.

**Evidence location**: LCLN-Phase PR description includes the Frozen Guard output.

### AC-LCLN-003.5 — Per-rule count regression check (plan-auditor iter-1 Finding 12)

**Statement**: No individual LR-XX rule's finding count shall increase post-phase vs pre-phase. This is a strengthening of the prior weaker invariant "unique rule type count is preserved".

**Test command**:

```bash
PHASE=P1
PRE=.moai/state/lint-baseline-pre-LCLN-${PHASE}.json
POST=.moai/state/lint-baseline-post-LCLN-${PHASE}.json

# Build per-rule maps from each baseline and compare
jq -s '
  def by_rule: .violations | group_by(.rule) | map({rule: .[0].rule, count: length}) | from_entries[..] // {};
  (.[0].violations | group_by(.rule) | map({key: .[0].rule, value: length}) | from_entries) as $pre |
  (.[1].violations | group_by(.rule) | map({key: .[0].rule, value: length}) | from_entries) as $post |
  ($pre + $post | keys) | map(
    {rule: ., pre: ($pre[.] // 0), post: ($post[.] // 0)}
  ) | map(select(.post > .pre))
' $PRE $POST > /tmp/per-rule-regressions-${PHASE}.json

REGRESSIONS=$(jq 'length' /tmp/per-rule-regressions-${PHASE}.json)

if [ "$REGRESSIONS" -ne 0 ]; then
  echo "BLOCK: ${REGRESSIONS} per-rule regression(s) in LCLN-${PHASE}:"
  jq '.[]' /tmp/per-rule-regressions-${PHASE}.json
  exit 1
fi
echo "OK: no per-rule count increases"
```

**Pass condition**: Exit code 0; `REGRESSIONS=0` (no individual rule's count increased).

**Evidence location**: LCLN-Phase PR description; `/tmp/per-rule-regressions-LCLN-<PHASE>.json`.

---

## AC-LCLN-004 — Test suite passes (must-pass: 004.1; advisory: 004.2)

### AC-LCLN-004.1 — Full test suite passes (must-pass)

**Statement**: `go test ./...` shall pass on every LCLN-Phase PR.

**Test command**:

```bash
go test ./... 2>&1 | tail -10
echo "Exit: $?"
```

**Pass condition**: Exit code 0; no `FAIL` lines in output.

**Evidence location**: CI log of LCLN-Phase PR.

### AC-LCLN-004.2 — Per-package coverage stable (ADVISORY per Finding 14 Option B)

**Statement** (advisory only — NOT in Definition of Done must-pass set): Per-package coverage SHOULD NOT drop by more than 1.0 percentage point on any touched package. Coverage regression triggers a warning + acknowledgment in PR description, NOT a merge block.

**Test command** (per plan-auditor iter-1 Finding 9, using `git worktree` instead of fragile `git stash`):

```bash
PHASE=P1
WORKTREE_DIR=/tmp/coverage-pre-LCLN-${PHASE}

# Capture pre-phase coverage from a clean worktree of main
rm -rf $WORKTREE_DIR
git worktree add $WORKTREE_DIR origin/main 2>/dev/null
(cd $WORKTREE_DIR && go test ./... -coverprofile=/tmp/coverage-pre-${PHASE}.out 2>&1 \
  | grep -E "^ok\s+" | awk '{print $2 "," $5}' | sort) > /tmp/coverage-pre-${PHASE}.csv
git worktree remove --force $WORKTREE_DIR

# Capture post-phase coverage from current branch
go test ./... -coverprofile=/tmp/coverage-post-${PHASE}.out 2>&1 \
  | grep -E "^ok\s+" | awk '{print $2 "," $5}' | sort > /tmp/coverage-post-${PHASE}.csv

# Compare
REGRESSIONS=$(join -t, /tmp/coverage-pre-${PHASE}.csv /tmp/coverage-post-${PHASE}.csv | awk -F, '{
  gsub(/%/, "", $2); gsub(/%/, "", $3);
  if ($3 - $2 < -1.0) print "REGRESSION: " $1 " " $2 "% → " $3 "%";
}')

if [ -n "$REGRESSIONS" ]; then
  echo "WARN (advisory): Coverage regression detected:"
  echo "$REGRESSIONS"
  echo "Acknowledgment required in PR description; NOT a merge block."
else
  echo "OK: no coverage regression > 1.0pp"
fi
```

**Pass condition** (advisory): Either output is `OK:` OR PR description acknowledges the regression. Per Finding 14 Option B, this AC is NOT a merge gate — it informs but does not block.

**Evidence location**: LCLN-Phase PR description (if a regression is observed and acknowledged) OR CI log.

**Note**: Agent-definition edits should NOT affect Go test coverage in normal operation. This AC catches accidental Go-source edits that would indicate scope creep. Precondition: clean working tree required at test invocation time (the `git worktree add` pattern enforces this).

---

## AC-LCLN-005 — Mega-Sprint W1/W2/W3/W4 parallel safety (rebase hygiene)

**Statement**: The system shall remain mergeable with concurrent Mega-Sprint W1 (CONSTITUTION-DUAL-001), W2 (CORE-SLIM-001), W3 (HARNESS-AUTONOMY-001), W4 (PROJECT-MEGA-001) work.

### AC-LCLN-005.1 — Rebase onto latest main before each LCLN-Phase PR

**Test command** (run at the start of each LCLN-Phase's run-phase):

```bash
PHASE=P1
git fetch origin main
git rebase origin/main
# Verify the LCLN-Phase's baseline JSON is still valid against the new main
./bin/moai agent lint --strict --format=json > /tmp/lint-baseline-post-rebase.json
DIFF_TOTAL=$(jq -s '(.[0].summary.total - .[1].summary.total) | abs' \
  .moai/state/lint-baseline-pre-LCLN-${PHASE}.json /tmp/lint-baseline-post-rebase.json)
if [ "$DIFF_TOTAL" -gt 5 ]; then
  echo "WARN: main rebased, baseline shifted by $DIFF_TOTAL findings. Re-capture pre-LCLN-${PHASE} baseline."
fi
```

**Pass condition**: Rebase succeeds (no conflicts on agent-definition files); baseline drift |Δ| ≤ 5 findings.

**Evidence location**: LCLN-Phase PR commit history shows rebase-onto-main as the first commit.

### AC-LCLN-005.2 — W2-deferred set bound check (tightened per Finding 7)

**Statement**: The Mega-Sprint W2-deferred set, computed empirically at each LCLN-Phase entry, MUST contain a count in the range `[11, 16]` (= canonical 13 ± 3 upstream drift tolerance, where ±3 covers concurrent Mega-Sprint W1/W2 perturbation of the skill/agent file set; see `research.md` §2.2 for derivation).

**Test command** (run primarily for LCLN-Phase 4, since only Phase 4 depends on W2's outcome):

```bash
PHASE=P4
# Compute W2-deferred set from current baseline using union semantics
# (skill ∈ {moai-domain-backend, moai-domain-frontend, moai-domain-database})
# OR (agent ∈ {expert-backend, expert-frontend, expert-mobile})
W2_DEFER=$(jq '[.violations[] |
  select(.rule == "LR-08") |
  select(
    (.message | test("moai-domain-(backend|frontend|database)")) or
    (.file | test("expert-(backend|frontend|mobile)\\.md$"))
  )] | length' .moai/state/lint-baseline-pre-LCLN-${PHASE}.json)

echo "W2-deferred set size: $W2_DEFER"

if [ "$W2_DEFER" -lt 11 ] || [ "$W2_DEFER" -gt 16 ]; then
  echo "BLOCK: W2-deferred set size $W2_DEFER outside expected range [11, 16]"
  exit 1
fi
echo "OK: W2-deferred set size $W2_DEFER in expected range [11, 16]"
```

**Pass condition**: `W2_DEFER` ∈ `[11, 16]`. Empirical canonical (2026-05-19): 13.

**Evidence location**: `.moai/state/lint-w2-deferred.json` (artifact committed in LCLN-Phase 4 PR).

---

## AC-LCLN-007 — Template-first invariant enforcement (NEW per plan-auditor iter-1 Finding 6)

**Statement** (from spec.md REQ-LCLN-007): When a fix targets an agent definition, the edit shall be applied to `internal/template/templates/.claude/agents/moai/<agent>.md` first, then propagated via `make build` + sync. This AC catches the highest-leverage process invariant.

### AC-LCLN-007.1 — Template + live edits mirror, embedded.go regenerated

**Test command** (run in every LCLN-Phase PR pre-merge, EXCEPT Phase 2 which has different semantics — see §AC-LCLN-007.2):

```bash
PHASE=P1
# (a) Verify template-side file edits mirror live-side file edits (by filename match)
TEMPLATE_EDITS=$(git diff --name-only main..HEAD | grep -E '^internal/template/templates/\.claude/agents/moai/.*\.md$' | sed 's|internal/template/templates/||' | sort)
LIVE_EDITS=$(git diff --name-only main..HEAD | grep -E '^\.claude/agents/moai/.*\.md$' | sort)

if [ "$TEMPLATE_EDITS" != "$LIVE_EDITS" ]; then
  echo "BLOCK: template/live edits do not mirror:"
  echo "  Template-side edited files: $TEMPLATE_EDITS"
  echo "  Live-side edited files: $LIVE_EDITS"
  exit 1
fi

# (b) Verify embedded.go regenerated in the same PR
if ! git diff --name-only main..HEAD | grep -q "^internal/template/embedded\.go$"; then
  echo "BLOCK: internal/template/embedded.go not regenerated in this PR"
  echo "       Run 'make build' and commit the result."
  exit 1
fi

echo "OK: template-first invariant satisfied; embedded.go regenerated"
```

**Pass condition** (Phases 1, 3, 4): Exit code 0. Template + live edit sets match; `embedded.go` present in PR diff.

**Evidence location**: LCLN-Phase PR description Frozen Guard + Template-First section.

### AC-LCLN-007.2 — Phase 2 live-only deletion exception

**Statement**: LCLN-Phase 2 deletes `.claude/agents/moai/expert-mobile.md` only (live surface). The template counterpart was deleted in Mega-Sprint W0; Phase 2 has no template-side change. Therefore AC-LCLN-007.1's "mirror" check is replaced by an explicit asymmetry assertion for this phase.

**Test command** (Phase 2 only):

```bash
PHASE=P2
# Phase 2: expect ONLY live-side deletion of expert-mobile.md, NO template-side change, NO embedded.go change

DELETED_LIVE=$(git diff --name-only --diff-filter=D main..HEAD | grep "^\.claude/agents/moai/expert-mobile\.md$" || echo "")
DELETED_TEMPLATE=$(git diff --name-only --diff-filter=D main..HEAD | grep "^internal/template/templates/\.claude/agents/moai/expert-mobile\.md$" || echo "")

if [ -z "$DELETED_LIVE" ]; then
  echo "BLOCK: expected live deletion of .claude/agents/moai/expert-mobile.md not found"
  exit 1
fi
if [ -n "$DELETED_TEMPLATE" ]; then
  echo "BLOCK: unexpected template-side deletion of expert-mobile.md (Mega-Sprint W0 already deleted it)"
  exit 1
fi
echo "OK: Phase 2 asymmetric deletion validated"
```

**Pass condition** (Phase 2): Exit code 0. Live-side deletion present; template-side absent (was deleted in Mega-Sprint W0).

**Evidence location**: LCLN-Phase 2 PR description.

---

## AC-LCLN-Phase-specific gates

### AC-LCLN-P1-LR03

**Statement**: After LCLN-Phase 1 merges, LR-03 finding count is reduced to 0 (excluding mobile-LR-03 which is cleared by Phase 2).

**Test command**:

```bash
jq '[.violations[] | select(.rule == "LR-03") | select(.file | test("expert-mobile\\.md$") | not)] | length' \
  .moai/state/lint-baseline-post-LCLN-P1.json
```

**Pass condition**: Output is `0`.

### AC-LCLN-P1-LR12

**Test command**:

```bash
jq '[.violations[] | select(.rule == "LR-12")] | length' .moai/state/lint-baseline-post-LCLN-P1.json
```

**Pass condition**: Output is `0`.

### AC-LCLN-P1-LR06

**Test command**:

```bash
jq '[.violations[] | select(.rule == "LR-06") | select(.file | test("expert-mobile\\.md$") | not)] | length' \
  .moai/state/lint-baseline-post-LCLN-P1.json
```

**Pass condition**: Output is `0`.

### AC-LCLN-P1-LR05

**Test command**:

```bash
jq '[.violations[] | select(.rule == "LR-05")] | length' .moai/state/lint-baseline-post-LCLN-P1.json
```

**Pass condition**: Output is `0`.

### AC-LCLN-P2-mobile

**Statement**: After LCLN-Phase 2 merges, `.claude/agents/moai/expert-mobile.md` does not exist.

**Test command**:

```bash
[ ! -f .claude/agents/moai/expert-mobile.md ] && echo "PASS" || echo "FAIL: file still exists"
```

**Pass condition**: Output is `PASS`.

### AC-LCLN-P3-LR01

**Test command**:

```bash
jq '[.violations[] | select(.rule == "LR-01")] | length' .moai/state/lint-baseline-post-LCLN-P3.json
```

**Pass condition**: Output is `0`.

### AC-LCLN-P3-LR02

**Test command**:

```bash
jq '[.violations[] | select(.rule == "LR-02")] | length' .moai/state/lint-baseline-post-LCLN-P3.json
```

**Pass condition**: Output is `0`.

### AC-LCLN-P4-drift-classification

**Statement**: W2-deferred subset is enumerated and committed in LCLN-Phase 4 PR.

**Test command**:

```bash
[ -f .moai/state/lint-w2-deferred.json ] && \
  jq '.deferred_findings | length' .moai/state/lint-w2-deferred.json
```

**Pass condition**: File exists; deferred count ∈ `[11, 16]` (per AC-LCLN-005.2).

### AC-LCLN-P4-add-skills

**Statement**: After LCLN-Phase 4 merges, all W4-resolvable LR-08 findings are 0; only the W2-deferred residual (12 = 13 canonical set − 1 cleared by Phase 2) remains.

**Test command**:

```bash
TOTAL_LR08=$(jq '[.violations[] | select(.rule == "LR-08")] | length' .moai/state/lint-baseline-post-LCLN-P4.json)
W2_DEFER=$(jq '.deferred_findings | length' .moai/state/lint-w2-deferred.json)

if [ "$TOTAL_LR08" -eq "$W2_DEFER" ]; then
  echo "PASS: residual LR-08 = $W2_DEFER (W2-deferred only)"
else
  echo "FAIL: residual=$TOTAL_LR08, W2-deferred=$W2_DEFER"
  exit 1
fi
```

**Pass condition**: Output is `PASS: residual LR-08 = <N> (W2-deferred only)` where `<N>` ∈ `[11, 16]`.

---

## Definition of Done (DoD) — SPEC closure (must-pass set)

This SPEC is considered COMPLETE (status `completed`) when ALL of the following must-pass items hold:

- [ ] AC-LCLN-001 passes — final `moai agent lint --strict` total = 12 (W2-deferred residual; further reduction to 0 happens via Mega-Sprint W2)
- [ ] AC-LCLN-001.3 passes after Mega-Sprint W2 lifecycle COMPLETE (cross-SPEC observation)
- [ ] AC-LCLN-002.1, 002.2, 002.3 passed on every LCLN-Phase PR (NEW=0 enforced)
- [ ] AC-LCLN-003.1, 003.2, 003.3, 003.4, 003.5 passed on every LCLN-Phase PR (no orthogonal regression, no FROZEN modification, no per-rule increase)
- [ ] AC-LCLN-004.1 passed on every LCLN-Phase PR (`go test ./...` exit 0)
- [ ] AC-LCLN-005.1 verified: every LCLN-Phase PR rebased onto main pre-merge
- [ ] AC-LCLN-005.2 verified for Phase 4: W2-deferred set size ∈ `[11, 16]`
- [ ] AC-LCLN-007.1 verified for Phases 1, 3, 4: template-first invariant + embedded.go regeneration
- [ ] AC-LCLN-007.2 verified for Phase 2: asymmetric deletion
- [ ] All 4 LCLN-Phases merged: Phase 1, Phase 2, Phase 3, Phase 4
- [ ] `.moai/state/lint-w2-deferred.json` committed and references SPEC-V3R5-CORE-SLIM-001 as the dissolving SPEC
- [ ] HISTORY entry added to `spec.md` with version bump (0.2.0 → 0.3.0) documenting status `draft → completed`
- [ ] Cross-reference added to `project_v3r5_w0_lifecycle_complete` memory entry's downstream list

## Advisory (NOT must-pass) items

- AC-LCLN-004.2 (coverage delta) — informational only per plan-auditor iter-1 Finding 14 Option B. Regression triggers PR-description acknowledgment, NOT a merge block.

## Quality gate criteria (TRUST 5)

- **Tested**: Per-phase delta verification via shell scripts above. No new unit tests required (linter behavior unchanged; this SPEC fixes agents, not the linter).
- **Readable**: All edits maintain or improve agent description clarity. No introduction of cryptic abbreviations.
- **Unified**: Frontmatter format consistent across all touched agents (same field ordering, same enum values).
- **Secured**: No security implication (agent metadata only; no permission/auth changes).
- **Trackable**: Each LCLN-Phase PR has its own commit on `main` (squash merge per CLAUDE.local.md §18.3), AC-LCLN-XXX referenced in commit messages.

## Final post-Mega-Sprint-W2 verification (cross-SPEC)

After SPEC-V3R5-CORE-SLIM-001 lifecycle COMPLETEs:

```bash
./bin/moai agent lint --strict --format=json | jq '.summary.total == 0'
```

Expected output: `true` (exact equality, per Finding 8).

This SPEC's contribution to that outcome: established the cleanup pipeline and reduced baseline from 176 → 12. Mega-Sprint W2 closes the remaining 12 via agent/skill retirement.
