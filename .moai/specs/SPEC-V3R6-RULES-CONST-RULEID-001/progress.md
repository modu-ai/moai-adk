# Progress — SPEC-V3R6-RULES-CONST-RULEID-001

> Tier S LEAN. spec.md (inline ACs) + progress.md (this file).
> §E.1 populated at plan-phase. §E.2–§E.5 are placeholder headings (populated by run/sync/Mx phases).

## §E.1 Plan-phase Audit-Ready Signal

- **SPEC ID self-check**: `decomposition: SPEC ✓ | V3R6 ✓ | RULES ✓ | CONST ✓ | RULEID ✓ | 001 ✓ → PASS` (canonical regex `^SPEC(-[A-Z][A-Z0-9]*)+-\d{3}$`, digit-only end anchor).
- **Frontmatter**: 12 canonical fields present + `tier: S`, `depends_on: []`. status=draft.
- **Out of Scope**: 4× `### Out of Scope —` H3 sub-headings with `-` bullets (satisfies OutOfScopeRule).
- **Deterministic verification done at plan-phase**:
  - `ruleIDPattern` = `^CONST-V3R[25]-\d{3,}$` confirmed at rule.go:10 (excludes V3R6).
  - Namespace set confirmed: V3R2 (73), V3R5 (41), V3R6 (1). V3R3/V3R4 absent.
  - `moai constitution validate` confirmed RED at HEAD (entry 114, `CONST-V3R6-001` regex fail, exit 1).
  - CLI load-path confirmed: reads deployed `.claude/rules/moai/core/zone-registry.md` (not template); deployed=2 / template=0 §25 divergence.
  - Pattern fix confirmed: `^CONST-V3R[256]-\d{3,}$` (enumeration, NOT `[2-6]` range — range would break TestRuleValidateV3R5ID v3r3/v3r4-invalid cases).
- **run-phase cycle_type**: tdd (reproduction-first).
- **Cross-SPEC dependency recorded**: this SPEC is a precondition for VERSION-FORMAT-001 AC-VFM-001a.

## §E.2 Run-phase Evidence

### AC-RCR-002 — reproduction test RED→GREEN (cycle_type=tdd)

**RED (pre-fix, 패턴 `^CONST-V3R[25]-\d{3,}$`)**:
```
$ go test ./internal/constitution/ -run 'TestRuleValidateV3R6ID' -v
=== NAME  TestRuleValidateV3R6ID/v3r6-valid
    rule_test.go:206: ID "CONST-V3R6-001": Validate() = rule ID "CONST-V3R6-001" does not match pattern "^CONST-V3R[25]-\\d{3,}$", want nil
=== NAME  TestRuleValidateV3R6ID/v3r6-valid-large
    rule_test.go:206: ID "CONST-V3R6-999": Validate() = rule ID "CONST-V3R6-999" does not match pattern "^CONST-V3R[25]-\\d{3,}$", want nil
--- FAIL: TestRuleValidateV3R6ID (0.00s)
    --- PASS: v3r5-regression / v3r3-still-invalid / v3r2-regression (0.00s)
    --- FAIL: v3r6-valid / v3r6-valid-large (0.00s)
    --- PASS: v3r4-still-invalid (0.00s)
FAIL    github.com/modu-ai/moai-adk/internal/constitution     0.492s
```
V3R6 케이스만 RED, regression 가드(v3r2/v3r3/v3r4/v3r5)는 정상 동작 — RED가 올바른 이유로 발생.

**GREEN (post-fix, 패턴 `^CONST-V3R[256]-\d{3,}$`)**:
```
$ go test ./internal/constitution/... -v | tail
--- PASS: TestRuleValidateV3R6ID (0.00s)
    --- PASS: v3r3-still-invalid / v3r2-regression / v3r6-valid-large / v3r5-regression / v3r6-valid / v3r4-still-invalid
PASS
ok      github.com/modu-ai/moai-adk/internal/constitution     0.403s
```

### AC-RCR-001 / AC-RCR-001b — V3R6 허용 + V3R2/V3R5 회귀 무손상 + V3R3/V3R4 계속 거부

전체 `TestRuleValidateV3R6ID` 6 케이스 PASS. `TestRuleValidateV3R5ID`(기존, v3r3/v3r4-invalid 포함)도 PASS — `[256]` enumeration이 `[2-6]` range가 아님을 회귀 테스트로 보증.

### AC-RCR-003 (keystone) — `moai constitution validate` load-abort cleared

```
$ go run ./cmd/moai constitution validate 2>&1 | grep -c 'does not match pattern'
0
$ go run ./cmd/moai constitution validate >/dev/null 2>&1; echo "exit=$?"
exit=1
$ go run ./cmd/moai constitution validate 2>&1 | grep -c '\[DRIFT\]'
73
```
- **in-scope 성공 신호**: `does not match pattern` = **0** → V3R6 load-abort 해소 확인. CONST-V3R6-001 엔트리가 정상 로드되어 post-load DRIFT 루프가 완전 실행됨.
- **out-of-scope exit=1**: 73건 un-masked DRIFT (사전 존재 CONST-V3R2/V3R5 clause). 본 SPEC이 un-mask만 수행, 소유하지 않음 (§A.1). 백로그 CONST-DRIFT-CLEANUP-001 소관. 본 SPEC의 결함 아님.

### AC-RCR-004 — Go toolchain clean

```
$ go build ./... ; echo "build-exit=$?"
build-exit=0
$ go vet ./internal/constitution/... ; echo "vet-exit=$?"
vet-exit=0
$ go test ./internal/constitution/...
ok      github.com/modu-ai/moai-adk/internal/constitution     0.396s
```

### AC-RCR-005 — doc-comment 정합

```
$ grep -n 'V3R6\|V3R2.*V3R5\|namespace' internal/constitution/rule.go
8: // ruleIDPattern is a regex constant for validating IDs in CONST-V3R2-NNN, CONST-V3R5-NNN, or CONST-V3R6-NNN format.
11: const ruleIDPattern = `^CONST-V3R[256]-\d{3,}$`
23:    // ID is the unique identifier in CONST-V3R2-NNN, CONST-V3R5-NNN, or CONST-V3R6-NNN format.
```

### AC-RCR-006 — additive-only behavior

`git diff` scope = 정확히 2 파일:
- `internal/constitution/rule.go`: pattern(`[25]`→`[256]`) + 2 doc comments. 다른 validator 로직 변경 없음.
- `internal/constitution/rule_test.go`: `TestRuleValidateV3R6ID` 1개 추가.

DRIFT/FROZEN_WITHOUT_CANARY/SOURCE_FILE_MISSING/INVALID_ZONE_CLASS 검사 로직은 무결섭 (regex 상수만 교체).

### Coverage

```
$ go test -cover ./internal/constitution/...
ok      github.com/modu-ai/moai-adk/internal/constitution     0.396s   coverage: 41.3% of statements
```
패키지 전체 41.3% — 본 SPEC은 1-line regex 상수 변경이므로, regex 라인은 TestRuleValidateV3R6ID(6 cases) + 기존 TestRuleValidate* 들로 충분히 커버됨.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-06-19T03:25:00+09:00
run_commit_sha: "562c96622"
run_status: implemented
ac_pass_count: 6
ac_fail_count: 0
preserve_list_post_run_count: 0
l44_pre_commit_fetch: "skipped — subagent context (orchestrator-side obligation)"
l44_post_push_fetch: "n/a — push는 orchestrator 권한"
new_warnings_or_lints_introduced: 0
cross_platform_build:
  darwin_arm64: pass
  linux_amd64: "not cross-checked (regex 상수, GOOS 미관련)"
  windows_amd64: "not cross-checked (동일)"
total_run_phase_files: 2
m1_to_mN_commit_strategy: "single M1 commit (Tier S, 1-line regex fix + reproduction test)"
```

**Self-verification (E1-E7)**:
- E1 AC matrix: 6/6 PASS (AC-RCR-001/001b/002/003/004/005/006)
- E2 Cross-platform build: darwin pass; regex 상수는 GOOS 무관
- E3 Coverage: 41.3% (1-line fix scope — 합리적)
- E4 Subagent-boundary grep: n/a (constitution 패키지는 subagent 도메인 아님)
- E5 Lint: go vet clean, new warnings 0
- E6 Push: pending orchestrator
- E7 Blocker: none

## §E.4 Sync-phase Audit-Ready Signal

sync_commit_sha: c55b53f41

Sync-phase orchestrator-direct (GLM manager-docs spawn context-limit fallback per `feedback_glm_orchestrator_direct_sync_mx`). frontmatter status in-progress → completed rides this sync commit (3-phase close per SPEC-V3R6-LIFECYCLE-REDESIGN-001 REQ-LR-008 — `completed` transition merged into the sync commit, no separate Mx chore). §E.5 Mx-phase retired (folded into §E.4 per LIFECYCLE-REDESIGN). CHANGELOG entry added (Sprint 16 RULES cohort).

## §E.5 Mx-phase Audit-Ready Signal

_<pending Mx-phase>_
