---
id: SPEC-V3R5-CONSTITUTION-DUAL-001
title: "Constitution Dual-Zone Formalization with Validate CLI — Acceptance Criteria"
version: "0.1.0"
status: draft
created: 2026-05-20
updated: 2026-05-20
author: GOOS Kim
priority: P1
phase: "v3.5.0"
module: ".claude/rules/moai + internal/constitution + internal/cli"
lifecycle: spec-anchored
tags: "constitution, dual-zone, frozen, evolvable, zone-registry, mega-sprint, w1, acceptance"
---

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-05-20 | GOOS Kim (via MoAI orchestrator) | Initial acceptance criteria — 5 Given/When/Then scenarios + edge cases + TRUST 5 mapping |
| 0.1.0 | 2026-05-20 | GOOS Kim (via MoAI orchestrator) | Iteration 2 revision — addressed plan-auditor BLOCKING defects (zone-registry 75→72, HARD rules 102→111 empirical, AC methodology unified) + SHOULD defects (5 new ACs for traceability, V3R5-001 namespace, plan.md/acceptance.md Out of Scope sections, 3 sentinel REQs added, AC-CDL-005 split). |

---

## 1. 검증 개요 (Verification Overview)

본 문서는 SPEC-V3R5-CONSTITUTION-DUAL-001 의 acceptance criteria 를 정의한다.

검증 원칙:

- 모든 AC 는 binary pass/fail (관찰 가능한 증거 기반)
- TRUST 5 quality gate 통과 필수
- 5초 이내 validate 명령 완료 (performance criterion)
- plan-auditor 가 본 acceptance.md 의 모든 scenario 를 독립 검증

**Empirical baseline (binding for all ACs)** at main HEAD `3bd2aa291`:

- 15 canonical source files (per spec.md §2.2) contain N=111 [HARD] occurrences
- zone-registry.md contains 72 entries (CONST-V3R2-001..046 + 049 + 051..072 + 150..152, with 3 internal gaps at 047/048/050)
- Phase B target: 39 new entries (111 − 72 = 39) using CONST-V3R5-001..039 parallel namespace

---

## 2. Given/When/Then 시나리오 (10 Core ACs)

### AC-CDL-001 — D1 annotation completeness

**Given**: The 15 canonical constitution source files (enumerated in spec.md §2.2) exist:

- `CLAUDE.md`
- `.claude/rules/moai/core/agent-common-protocol.md`
- `.claude/rules/moai/core/askuser-protocol.md`
- `.claude/rules/moai/core/moai-constitution.md`
- `.claude/rules/moai/design/constitution.md`
- `.claude/rules/moai/workflow/ci-autofix-protocol.md`
- `.claude/rules/moai/workflow/ci-watch-protocol.md`
- `.claude/rules/moai/workflow/context-window-management.md`
- `.claude/rules/moai/workflow/session-handoff.md`
- `.claude/rules/moai/workflow/spec-workflow.md`
- `.claude/rules/moai/workflow/worktree-integration.md`
- `.claude/rules/moai/workflow/worktree-state-guard.md`
- `.claude/rules/moai/development/agent-authoring.md`
- `.claude/rules/moai/development/branch-origin-protocol.md`
- `.claude/rules/moai/development/skill-authoring.md`

The empirical total count of `[HARD]` rule occurrences is X=111 at main HEAD `3bd2aa291`.

**When**:

```bash
FILES=(
  CLAUDE.md
  .claude/rules/moai/core/agent-common-protocol.md
  .claude/rules/moai/core/askuser-protocol.md
  .claude/rules/moai/core/moai-constitution.md
  .claude/rules/moai/design/constitution.md
  .claude/rules/moai/workflow/ci-autofix-protocol.md
  .claude/rules/moai/workflow/ci-watch-protocol.md
  .claude/rules/moai/workflow/context-window-management.md
  .claude/rules/moai/workflow/session-handoff.md
  .claude/rules/moai/workflow/spec-workflow.md
  .claude/rules/moai/workflow/worktree-integration.md
  .claude/rules/moai/workflow/worktree-state-guard.md
  .claude/rules/moai/development/agent-authoring.md
  .claude/rules/moai/development/branch-origin-protocol.md
  .claude/rules/moai/development/skill-authoring.md
)

HARD_COUNT=$(grep -hcE '\[HARD\]' "${FILES[@]}" | awk '{s+=$1} END {print s}')
ZONE_COUNT=$(grep -hcE '\[ZONE:(Frozen|Evolvable)\]' "${FILES[@]}" | awk '{s+=$1} END {print s}')
echo "HARD: $HARD_COUNT, ZONE: $ZONE_COUNT"
```

**Then**: ZONE_COUNT must be greater than or equal to HARD_COUNT (Y ≥ X). Every [HARD] rule has at least one zone marker. (Multiple zone markers on a single rule are permitted but produce a `DUPLICATE_ZONE_MARKER` warning per REQ-CDL-018.)

**Edge case**: A [HARD] rule may span multiple lines. The grep counts line occurrences, so multi-line rules with one marker per rule are counted as a single occurrence in both X and Y. Acceptance: rules counted by `[HARD]` keyword occurrence, not by logical rule count.

---

### AC-CDL-002 — D2 100% coverage

**Given**: The `zone-registry.md` file has been extended with new entries (CONST-V3R5-001 onwards) covering previously unmapped [HARD] rules. The total entry count in the registry is N. Pre-SPEC baseline N=72; Phase B target N≥111.

**When**:

```bash
N=$(./moai constitution list --format json | jq '.entries | length')

# M = sum of [HARD] occurrences across the 15 source files enumerated in spec.md §2.2
FILES=(  # identical list as AC-CDL-001
  CLAUDE.md
  .claude/rules/moai/core/agent-common-protocol.md
  .claude/rules/moai/core/askuser-protocol.md
  .claude/rules/moai/core/moai-constitution.md
  .claude/rules/moai/design/constitution.md
  .claude/rules/moai/workflow/ci-autofix-protocol.md
  .claude/rules/moai/workflow/ci-watch-protocol.md
  .claude/rules/moai/workflow/context-window-management.md
  .claude/rules/moai/workflow/session-handoff.md
  .claude/rules/moai/workflow/spec-workflow.md
  .claude/rules/moai/workflow/worktree-integration.md
  .claude/rules/moai/workflow/worktree-state-guard.md
  .claude/rules/moai/development/agent-authoring.md
  .claude/rules/moai/development/branch-origin-protocol.md
  .claude/rules/moai/development/skill-authoring.md
)
M=$(grep -hcE '\[HARD\]' "${FILES[@]}" | awk '{s+=$1} END {print s}')

echo "Registry entries: $N, HARD rules in sources: $M"
```

**Then**: N must be greater than or equal to M (target: N≥111). No [HARD] rule in any of the 15 canonical source files remains unmapped. Additionally, `./moai constitution validate --strict` exits with code 0 (no `ZONE_UNREGISTERED` errors).

**Source file list (binding)**: The 15 files above are the canonical M-computation source. Any future addition of constitution files requires updating spec.md §2.2 and re-running M computation. This is the **unified counting methodology** that resolves the iteration 1 D3 defect (AC-CDL-001 and AC-CDL-002 now both use the same 15-file grep sum).

**Edge case (V3R5 namespace)**: Newly added entries (CONST-V3R5-001..) coexist with existing CONST-V3R2-* entries as a parallel namespace. Both contribute to N. The V3R2 sequence retains its 3 internal gaps (047/048/050) unchanged.

---

### AC-CDL-003 — D3 validate CLI happy path

**Given**: The `zone-registry.md` is fully in sync with all referenced source files. Every entry's `clause:` text exists at the recorded `anchor:` in the recorded `file:`. No source files are missing. No [HARD] rules are unregistered.

**When**:

```bash
./moai constitution validate --strict --format json
echo "Exit code: $?"
```

**Then**:

- Exit code: 0
- JSON output structure (semantic equivalence):
  ```json
  {
    "status": "ok",
    "drift_count": 0,
    "missing_count": 0,
    "unregistered_count": 0,
    "entries": []
  }
  ```
- Execution time: < 5 seconds (measured via `time` command) for current corpus (≤ 150 entries, ≤ 15 source files).

**Edge case (warnings)**: Even on happy path, the validator may emit warnings for `DUPLICATE_ZONE_MARKER` (REQ-CDL-018) or `STALE_ENTRY` (REQ-CDL-019, entry timestamp older than 90 days). With `--strict` alone, warnings do NOT fail the command. Acceptance: status remains `ok`, exit code 0. With `--strict --fail-on-warning`, warnings promote to errors.

---

### AC-CDL-004 — D3 validate CLI drift detection

**Given**: A test fixture or controlled modification has changed the text of a [HARD] rule in one of the 15 canonical files. The corresponding zone-registry entry's `clause:` field no longer matches the source.

Example modification:

```diff
- [HARD] All user-facing questions MUST go through AskUserQuestion ...
+ [HARD] All user-directed inquiries SHOULD route through AskUserQuestion ...
```

**When**:

```bash
./moai constitution validate --strict --format json
echo "Exit code: $?"
```

**Then**:

- Exit code: 1
- JSON output contains at least one entry with `status: DRIFT`:
  ```json
  {
    "status": "drift",
    "drift_count": 1,
    "missing_count": 0,
    "unregistered_count": 0,
    "entries": [
      {
        "id": "CONST-V3R2-XXX",
        "file": "...",
        "anchor": "...",
        "status": "DRIFT",
        "detail": "(diff summary)"
      }
    ]
  }
  ```
- Textual output (without `--format json`): human-readable diff summary identifying the affected entry id, file, and approximate location.

**Edge case (anchor relocation)**: If the [HARD] rule text is unchanged but moved to a different anchor section, the validator treats this as drift (anchor mismatch) rather than text drift. Acceptance: `status: DRIFT` with `detail` indicating anchor mismatch.

**Multi-key coverage**: AC-CDL-004 also covers fixtures for SOURCE_FILE_MISSING (REQ-CDL-013, exit 2), ZONE_UNREGISTERED (REQ-CDL-008), FROZEN_WITHOUT_CANARY (REQ-CDL-015), ANCHOR_NOT_FOUND (REQ-CDL-016), DUPLICATE_ID (REQ-CDL-017, exit 1 always), STALE_ENTRY (REQ-CDL-019, warning only). One fixture per error key in `validator_test.go`.

---

### AC-CDL-005a — CI integration (automatable)

**Given**: A pull request is opened targeting the `main` branch of the moai-adk-go repository. The `.github/workflows/ci.yml` file has been updated with the `Constitution Validate` job (per Phase D Task D-1). The CI workflow YAML explicitly clears `MOAI_CONSTITUTION_SKIP_VALIDATE` env to prevent bypass.

**When**: GitHub Actions workflow `ci.yml` triggers on the PR. The `constitution-validate` job runs `./moai constitution validate --strict --format json` with `CI=true` env.

**Then**:

- If the PR's changes maintain registry-source sync: job succeeds (exit 0), step appears as ✓ in PR checks list, PR is mergeable (subject to other required checks).
- If the PR introduces drift (e.g., modified [HARD] rule without updating registry): job fails (exit code 1), step appears as ✗, PR is blocked from merging until drift is resolved.
- This is **auto-verifiable** within a single CI run — no human intervention required.

---

### AC-CDL-005b — Branch protection (manual maintainer verification)

**Given**: The Phase D Task D-2 branch protection update has been applied by a maintainer with admin permission after the SPEC merge (PR-merge time + maintainer manual step).

**When**: The maintainer (or any verifier with read access) invokes:

```bash
gh api /repos/modu-ai/moai-adk/branches/main/protection | jq '.required_status_checks.contexts'
```

**Then**: The response array contains 5 entries: `Lint`, `Test (ubuntu-latest)`, `Build (linux/amd64)`, `CodeQL`, `Constitution Validate`. The CLAUDE.local.md §18.7 baseline document also reflects the 4→5 update in the same commit cycle as Phase D.

**Verification classification**: This AC is explicitly classified as **manual verification — verified by maintainer applying `gh api PATCH ...` after PR merge** because branch protection updates require admin permission that CI runners do not possess. The verification cannot complete within a single CI run.

**Edge case (skip override at CI level)**: If a maintainer attempts to set `MOAI_CONSTITUTION_SKIP_VALIDATE=1` in the workflow env to bypass validation, the CI workflow YAML enforces `env: { MOAI_CONSTITUTION_SKIP_VALIDATE: '' }`. Acceptance: even if the env is set elsewhere, CI fails on drift. (Local-only override per REQ-CDL-011 still works for developers.)

---

### AC-CDL-006 — REQ-CDL-002 zone_class enum compliance

**Given**: All zone-registry entries (both pre-existing 72 CONST-V3R2-NNN and 39+ newly added CONST-V3R5-NNN) declare a `zone_class:` field per Phase B Task B-2 retroactive assignment.

**When**:

```bash
ALLOWED=("frozen-canonical" "frozen-safety" "evolvable-tuning" "evolvable-experimental")

# Extract all zone_class values from registry
./moai constitution list --format json | jq -r '.entries[].zone_class' | sort -u > /tmp/zone_classes.txt

# Assert each is in the allow-list
while read -r zc; do
  if [[ ! " ${ALLOWED[*]} " =~ " ${zc} " ]]; then
    echo "INVALID zone_class: $zc"
    exit 1
  fi
done < /tmp/zone_classes.txt
```

**Then**: Every entry's `zone_class` matches one of the 4 allowed values: `frozen-canonical`, `frozen-safety`, `evolvable-tuning`, `evolvable-experimental`. `moai constitution validate --strict` rejects entries with invalid `zone_class` values with error key `INVALID_ZONE_CLASS` and exit code 1.

**Edge case (legacy entries without zone_class)**: Pre-existing 72 entries must have `zone_class` retroactively assigned during Phase B Task B-2. An entry missing `zone_class` triggers `INVALID_ZONE_CLASS` (treated as null which is not in the allow-list).

---

### AC-CDL-007 — REQ-CDL-005 ID format compliance for new entries

**Given**: The set of zone-registry entries newly introduced by this SPEC (all entries with `id` matching `^CONST-V3R5-`). Target: 39+ new entries beginning at CONST-V3R5-001.

**When**:

```bash
# Extract new V3R5 entry IDs
NEW_IDS=$(./moai constitution list --format json | jq -r '.entries[].id' | grep -E '^CONST-V3R5-')

# Verify each matches the canonical regex
for id in $NEW_IDS; do
  if [[ ! "$id" =~ ^CONST-V3R5-[0-9]{3}$ ]]; then
    echo "INVALID ID format: $id"
    exit 1
  fi
done

# Verify sequencing starts at 001 and is contiguous (no internal gaps in V3R5 namespace)
FIRST=$(echo "$NEW_IDS" | sort | head -1)
[[ "$FIRST" == "CONST-V3R5-001" ]] || { echo "First V3R5 ID is not CONST-V3R5-001: $FIRST"; exit 1; }
```

**Then**: Every new entry ID matches the regex `^CONST-V3R5-[0-9]{3}$` exactly. The V3R5 namespace begins at `CONST-V3R5-001` and is contiguous (no internal gaps introduced by this SPEC). Parallel V3R2 namespace remains untouched (including its 3 historical gaps at 047/048/050).

---

### AC-CDL-008 — REQ-CDL-009 live reload — no restart between writes

**Given**: An in-progress shell session in which `./moai constitution list --zone evolvable` has been invoked once and produced output reflecting the current registry state.

**When**: A test fixture modifies `zone-registry.md` (e.g., adds one new entry with `zone: Evolvable`), then immediately re-invokes `./moai constitution list --zone evolvable` within the same shell session without restart:

```bash
COUNT_BEFORE=$(./moai constitution list --zone evolvable --format json | jq '.entries | length')

# Modify registry: add one Evolvable entry via fixture
cat >> .claude/rules/moai/core/zone-registry.md <<EOF
- id: CONST-V3R5-999
  zone: Evolvable
  zone_class: evolvable-tuning
  file: CLAUDE.md
  anchor: "#testing"
  clause: "[HARD] (test fixture)"
  canary_gate: false
EOF

COUNT_AFTER=$(./moai constitution list --zone evolvable --format json | jq '.entries | length')

echo "Before: $COUNT_BEFORE, After: $COUNT_AFTER"
```

**Then**: COUNT_AFTER = COUNT_BEFORE + 1. The second invocation reflects the updated entry set. No daemon, cache, or restart is required. Validator test fixture `TestList_ReflectsUpdatesWithoutRestart` covers this scenario.

---

### AC-CDL-009 — REQ-CDL-011 MOAI_CONSTITUTION_SKIP_VALIDATE override

**Given**: A controlled drift scenario exists (registry entry clause does not match source). Without override, `moai constitution validate --strict` returns exit code 1.

**When**:

```bash
# Without override: drift detected
./moai constitution validate --strict --format json
EXIT_WITHOUT=$?

# With override: bypass
MOAI_CONSTITUTION_SKIP_VALIDATE=1 ./moai constitution validate --strict --format json 2> /tmp/stderr.log
EXIT_WITH=$?

echo "Without override: $EXIT_WITHOUT, With override: $EXIT_WITH"
grep -q "WARN: validation skipped" /tmp/stderr.log && echo "Warning emitted: yes"
```

**Then**:

- EXIT_WITHOUT = 1 (drift detected)
- EXIT_WITH = 0 (bypass)
- stderr contains `WARN: validation skipped (MOAI_CONSTITUTION_SKIP_VALIDATE=1)`
- stdout (with `--format json`) does NOT report the drift entries (bypass is complete)

**Edge case (CI workflow)**: The CI workflow YAML explicitly clears this env (`env: { MOAI_CONSTITUTION_SKIP_VALIDATE: '' }`) so the override is **local-developer only**, never available in CI.

---

### AC-CDL-010 — REQ-CDL-014 read-only assertion

**Given**: A controlled test environment where the project source files and zone-registry are made read-only (`chmod -R -w .` or equivalent), except for `/tmp` scratch area used for output.

**When**:

```bash
# Make project read-only (except /tmp)
chmod -R -w .

# Run validator
./moai constitution validate --strict --format json > /tmp/validate-output.json
EXIT=$?

# Revert
chmod -R +w .

echo "Exit: $EXIT"
cat /tmp/validate-output.json
```

**Then**: The command produces a complete validation report (success or drift detection) without any write attempt to source files, registry, configuration, or state. No `EACCES`/`EROFS` errors due to attempted writes. The validator code review confirms no `os.WriteFile`, `os.Create`, or equivalent calls. Validator test fixture `TestValidate_ReadOnlyAssertion` (using `t.TempDir()` + `chmod -w`) covers this scenario in CI.

---

## 3. Edge Cases (보강 시나리오)

### EC-CDL-001 — Anchor mismatch handling

**Scenario**: A [HARD] rule is duplicated across two anchors (e.g., both `#user-interaction-boundary` and `#askuserquestion-protocol`). The registry entry points to one anchor; the other is not registered.

**Expected behavior**: validator reports the unregistered anchor as `ZONE_UNREGISTERED` if `--strict`. Without `--strict`, emits warning.

**Acceptance**: Test fixture covers this case in `validator_test.go`. Coverage report shows the branch.

### EC-CDL-002 — Markdown formatting differences

**Scenario**: registry clause stored as `"[HARD] All user-facing questions MUST go through AskUserQuestion ..."` but source file has equivalent text with different whitespace (`"[HARD]  All user-facing  questions"` — double space).

**Expected behavior**: validator normalizes whitespace (collapse multiple spaces to single) before substring matching. Treats this as OK, not DRIFT.

**Acceptance**: Whitespace normalization tested via `TestValidate_WhitespaceNormalization` fixture.

### EC-CDL-003 — Concurrent registry edits (DUPLICATE_ID)

**Scenario**: Two PRs simultaneously add entries to `zone-registry.md`, both using ID `CONST-V3R5-001` (collision).

**Expected behavior**: Out of scope for validator (git merge conflict resolves this at VCS level). However, validator MUST detect duplicate IDs at parse time and report `DUPLICATE_ID` error with exit code 1 regardless of `--strict` (REQ-CDL-017).

**Acceptance**: Test fixture with duplicate IDs verifies validator rejects with clear error message. Covered by AC-CDL-004 multi-key fixture set.

### EC-CDL-004 — Large registry performance

**Scenario**: Future registry growth to 500 entries across 30 source files.

**Expected behavior**: validate completes in < 10 seconds (degraded but acceptable). Current corpus (≤ 150 entries, 15 sources) completes in < 5 seconds (REQ-CDL-003).

**Acceptance**: Benchmark test (`BenchmarkValidate`) ensures linear scaling, not exponential.

### EC-CDL-005 — Code-fence content false positive

**Scenario**: A markdown code block contains `[HARD]` as illustrative example (not an actual rule).

**Expected behavior**: validator excludes code-fenced content from [HARD] rule extraction. Only inline (paragraph) [HARD] occurrences count.

**Acceptance**: Test fixture with code-fence containing `[HARD]` verifies exclusion.

### EC-CDL-006 — Empty source file

**Scenario**: A constitution `.md` file is empty (0 bytes) but referenced by registry entries.

**Expected behavior**: validator reports all referencing entries as `SOURCE_FILE_MISSING` (or new `SOURCE_FILE_EMPTY` status). Exit code 2.

**Acceptance**: Edge case test covers this.

### EC-CDL-007 — DUPLICATE_ZONE_MARKER warning

**Scenario**: A single [HARD] rule line has two ZONE markers applied (operator error): `[ZONE:Frozen] [ZONE:Evolvable] [HARD] ...`.

**Expected behavior**: validator emits `DUPLICATE_ZONE_MARKER` warning (REQ-CDL-018). Exit code unchanged unless `--strict --fail-on-warning`.

**Acceptance**: Test fixture with multi-marker line verifies warning emission without exit-code change.

### EC-CDL-008 — STALE_ENTRY warning

**Scenario**: A registry entry's last-update timestamp is older than 90 days relative to invocation time.

**Expected behavior**: validator emits `STALE_ENTRY` warning (REQ-CDL-019). Observation-only; does not affect exit code.

**Acceptance**: Test fixture manipulates entry mtime (or includes embedded timestamp metadata) to trigger warning. Covered in `TestValidate_StaleEntryWarning`.

---

## 4. Definition of Done (DoD)

본 SPEC 은 다음 조건 **모두** 만족 시 완료:

### 4.1 Phase 별 완료 조건

| Phase | DoD |
|-------|-----|
| Phase A | AC-CDL-001 PASS, 15 .md files marked, no [HARD] rule unmarked (ZONE_COUNT ≥ HARD_COUNT) |
| Phase B | AC-CDL-002 PASS (N≥111), AC-CDL-006 PASS (zone_class enum), AC-CDL-007 PASS (V3R5-001..NNN format), HISTORY updated |
| Phase C | AC-CDL-003 PASS, AC-CDL-004 PASS (all 10 sentinel keys), AC-CDL-008 PASS, AC-CDL-009 PASS, AC-CDL-010 PASS, coverage ≥ 85%, golangci-lint zero warnings |
| Phase D | AC-CDL-005a PASS (CI auto), AC-CDL-005b PASS (manual maintainer post-merge admin step) |

### 4.2 통합 완료 조건

- All 10 ACs (AC-CDL-001..010) binary PASS (note: 005 split into 005a + 005b)
- All 8 EDGE cases (EC-CDL-001..008) covered by test fixtures
- `./moai constitution validate --strict --format json` returns `status: ok`
- `./moai constitution list --zone evolvable | wc -l` 과 `--zone frozen | wc -l` 합산이 총 entry 수와 일치 (no entries unclassified)
- spec.md, plan.md, acceptance.md 의 status 가 `completed` 로 전환
- v3.5.0 release notes 에 본 SPEC 의 deliverables 기록

### 4.3 TRUST 5 Quality Gate

| Pillar | 검증 방법 | 통과 기준 |
|--------|-----------|-----------|
| Tested | `go test -cover ./internal/constitution/` | ≥ 85% coverage |
| Readable | `golangci-lint run ./internal/constitution/ ./internal/cli/constitution_validate.go` | zero warnings |
| Unified | `gofmt -l ./internal/constitution/` | empty output (formatted) |
| Secured | validator code review | read-only (REQ-CDL-014), no shell injection, no path traversal |
| Trackable | `git log --oneline feat/SPEC-V3R5-CONSTITUTION-DUAL-001..` | Conventional Commits 형식, Phase 별 분리 |

### 4.4 회귀 방지 (Regression Prevention)

- 기존 `moai constitution list` 동작 검증: 기존 72 entries 출력 unchanged (snapshot test)
- 기존 `moai constitution guard` / `amend` 동작 변경 없음 (C-CDL-006 orthogonality)
- 기존 `moai constitution list --zone frozen` 결과: 신규 entry 도 포함되어야 함 (확장 검증)
- 기존 CI workflows: `Lint`, `Test (ubuntu-latest)`, `Build (linux/amd64)`, `CodeQL` 모두 PASS 유지
- `make preflight` 로컬 실행: 1-2분 budget 유지 (validate 추가로 인한 추가 시간 < 10초)

---

## 5. plan-auditor 검증 포인트 (Independent Review Criteria)

본 acceptance.md 는 plan-auditor 의 다음 차원에서 검증된다:

### D1 — Brief Quality (목표 ≥ 0.85)

- 10 ACs binary verifiable (PASS/FAIL clear)
- Edge cases enumerate (≥ 8)
- EARS 형식 매핑 명시 (REQ-CDL-NNN → AC-CDL-NNN cross-reference in spec.md §8)
- All 19 REQs have ≥1 AC mapping (100% traceability)

### D2 — Phase Decomposition

- Phase A/B/C/D 별 DoD 명시
- Sequential dependency 명시 (A→B→C→D)
- 각 phase 별 commit 가능 (atomic)

### D3 — Risk Management

- Edge case scenarios (EC-CDL-001..008) 가 plan.md §7 Risk Matrix 와 정합성 유지
- TRUST 5 통과 기준 명시
- 회귀 방지 시나리오 포함

### D4 — Frontmatter Compliance

- 12-field canonical schema 준수
- snake_case alias 없음 (`created`/`updated`/`tags` 사용; `created_at`/`updated_at`/`labels` 금지)
- HISTORY 섹션 frontmatter 직후 위치

### D5 — Exclusion Discipline

- spec.md §5.2 EXCL-001~006 와 정합성 유지
- 본 acceptance.md §6 "Out of Scope" cross-reference
- 본 acceptance.md 는 implementation 검증만 다룸 (구현 자체는 plan.md scope)

### D6 — Lint Baseline

- 본 SPEC 자체의 lint warnings ≤ baseline
- 신규 entry 가 stale entry 를 생성하지 않음

---

## 6. Scope Boundaries

### 6.1 Out of Scope

See `spec.md` §5.2 for the canonical exclusion list (EXCL-001 through EXCL-006). Brief summary for acceptance.md siblings:

- **EXCL-001**: PreToolUse Frozen Guard hook acceptance criteria deferred to W3 HARNESS-AUTONOMY-001
- **EXCL-002**: agent/skill frontmatter `zone:` field acceptance criteria deferred to T3 Full / SPEC-V3R5-AGENT-ZONE-001
- **EXCL-003**: expert-backend / expert-frontend / expert-mobile retirement acceptance criteria deferred to W2 CORE-SLIM-001
- **EXCL-004**: Workflow rules beyond §2.2 enumerated 15 files — retroactive AC coverage deferred
- **EXCL-005**: design/constitution.md structural change acceptance criteria forbidden — inline marker verification only
- **EXCL-006**: precommit/generation-time check acceptance criteria deferred — CI-time verification only

본 acceptance.md 는 validate CLI 의 binary AC verification 에 집중하며, 다음은 본 acceptance.md scope 외:

- validator's internal algorithm performance tuning (deferred to follow-up SPEC if needed)
- registry schema evolution (e.g., adding new fields beyond zone_class) — deferred
- i18n of error messages (sentinel error keys remain language-agnostic per C-CDL-002)

---

## 7. 후속 검증 (Post-Completion Verification)

본 SPEC 완료 후 다음 검증 자동 실행:

1. **W2 SPEC 작성 시**: SPEC-V3R5-CORE-SLIM-001 의 expert-backend retirement 정당성이 본 SPEC 의 zone-registry FROZEN canonical 17 agent list 를 참조 (확인 필요)

2. **W3 SPEC 작성 시**: SPEC-V3R5-HARNESS-AUTONOMY-001 의 PreToolUse Frozen Guard hook 이 본 SPEC 의 `zone-registry.md` 파일을 정상 파싱 (smoke test)

3. **v3.5.0 release 시**: `moai doctor` 명령에 constitution validate 항목 통합 (옵션, 후속 SPEC 검토)
