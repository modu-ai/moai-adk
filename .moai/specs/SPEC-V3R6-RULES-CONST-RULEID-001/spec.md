---
id: SPEC-V3R6-RULES-CONST-RULEID-001
title: "Broaden constitution ruleIDPattern to accept the V3R6 namespace"
version: "0.1.0"
status: in-progress
created: 2026-06-19
updated: 2026-06-19
author: manager-spec
priority: P0
phase: "v3.0.0"
module: "internal/constitution"
lifecycle: spec-anchored
tags: "constitution, validator, ruleid, regex, bugfix, v3r6"
depends_on: []
tier: S
---

## HISTORY

| Date | Version | Change | Author |
|------|---------|--------|--------|
| 2026-06-19 | 0.1.0 | Plan-phase 산출물 생성 (Tier S LEAN, ACs inline). 형제 SPEC VERSION-FORMAT-001 plan-audit 중 발견된 결정론적 검증 버그 대응. | manager-spec |

## §A. Context / 배경

Sprint 16 rules-improvement 코호트의 5번째 SPEC. 형제 SPEC `SPEC-V3R6-...-VERSION-FORMAT-001`의 plan-audit 과정에서, `moai constitution validate`가 HEAD에서 항상 RED(non-zero exit)임이 **결정론적으로 검증**되었다.

### 실측 증거 (live, 본 plan-phase 세션에서 재현)

```
constitution validate: FAILED — 0 error(s) found

Error: load registry: entry 114 validation error: rule ID "CONST-V3R6-001" does not match pattern "^CONST-V3R[25]-\d{3,}$"
exit status 1
```

### 근본 원인 (검증 완료)

1. **`internal/constitution/rule.go:10`** — `const ruleIDPattern = ` 백틱 `^CONST-V3R[25]-\d{3,}$` 백틱. 문자 클래스 `[25]`는 **V3R2와 V3R5만** 허용하고 **V3R6을 배제**한다.
2. **`.claude/rules/moai/core/zone-registry.md`** — 배포된 레지스트리에 이제 `CONST-V3R6-NNN` 네임스페이스가 존재한다 (`id: CONST-V3R6-001` 1건, header 포함 `CONST-V3R6` 매치 2건).
3. **연쇄 효과** — 이 엔트리의 ID가 regex 검증에 실패하므로 `LoadRegistry`(`internal/constitution/validator.go:178`)가 **SentinelDrift 루프(validator.go:261-271)에 도달하기 전에** abort한다. 결과적으로 `moai constitution validate`는 항상 non-zero로 종료된다.

### 네임스페이스 집합 (검증 완료 — 본 SPEC의 핵심 입력)

배포 레지스트리 `id:` 필드 전수 조사 결과, 실재하는 CONST 네임스페이스는 **정확히 3종**이다:

| 네임스페이스 | 엔트리 수 |
|--------------|-----------|
| `CONST-V3R2` | 73 |
| `CONST-V3R5` | 41 |
| `CONST-V3R6` | 1 |

**V3R3, V3R4 네임스페이스는 존재하지 않는다.** 기존 테스트 `TestRuleValidateV3R5ID`(rule_test.go:147-148)는 `CONST-V3R3-001`/`CONST-V3R4-001`을 명시적으로 INVALID(`wantErr: true`)로 단언하여 이 부재를 보증한다.

### 패턴 수정 선택 (run-phase 입력, 본 SPEC에서 확정)

세 가지 후보:

| 후보 | 패턴 | V3R3/V3R4 영향 | 평가 |
|------|------|----------------|------|
| A (열거) | `^CONST-V3R[256]-\d{3,}$` | V3R3/V3R4 계속 거부 | **권장** — 실재 집합(2,5,6)과 정확히 일치, 기존 회귀 테스트 무손상 |
| B (범위) | `^CONST-V3R[2-6]-\d{3,}$` | V3R3/V3R4 **새로 허용** | 비권장 — 부재 네임스페이스를 의도치 않게 허용, `TestRuleValidateV3R5ID` v3r3/v3r4 케이스가 RED로 전환됨 |
| C (교체) | 명시적 OR 그룹 | 동일 | 과설계 |

**확정: 후보 A `^CONST-V3R[256]-\d{3,}$`** — 문자 클래스 `[256]`는 V3R2/V3R5/V3R6 **정확히 실재하는 3종**만 허용한다. 이는 컨텍스트가 제안한 `[2-6]` 범위보다 안전하다: `[2-6]`은 부재 네임스페이스 V3R3/V3R4를 새로 허용하여 기존 회귀 테스트(`TestRuleValidateV3R5ID`의 v3r3-invalid/v3r4-invalid 케이스)를 깨뜨린다.

### CLI load-path (검증 완료)

`moai constitution validate`가 어느 레지스트리를 읽는지 확인:

- CLI 진입점 `internal/cli/constitution.go:24` — `const constitutionRegistryRelPath = ".claude/rules/moai/core/zone-registry.md"`
- `resolveRegistryPath(cwd)`가 **프로젝트 cwd 기준 배포 경로**(`.claude/rules/moai/core/zone-registry.md`)를 해석한다 — embedded/template 사본이 아니다.
- `§25 divergence` 확인: 배포 사본 `CONST-V3R6` 매치 2건 / 템플릿 사본(`internal/template/templates/.claude/rules/moai/core/zone-registry.md`) 0건.

따라서 **CLI는 배포 사본을 읽으며**, 배포 사본에 `CONST-V3R6-001`이 존재한다. `rule.go`의 패턴을 넓히는 것은 V3R6 load-abort를 제거하기 위한 **필요 조건이지만 충분 조건은 아니다** — 패턴 수정 후에도 게이트는 별도 원인(73건의 사전 존재 CONST-V3R2/V3R5 clause DRIFT, 아래 §A.1 참조)으로 인해 여전히 exit 1을 반환한다. 템플릿 사본의 부재는 본 SPEC과 무관(아래 Out of Scope 참조).

### §A.1 un-mask 발견 (plan-audit iter-2, 결정론적 검증 완료)

패턴 수정이 V3R6 load-abort를 제거하면, 그동안 abort에 **가려져 있던** 73건의 잠재 DRIFT가 노출된다. 이는 본 plan-phase 세션에서 `[256]` 패턴을 임시 적용해 직접 측정·재확인했다:

- `[256]` 패턴 적용 시 `moai constitution validate` → **exit 1, "found 73 error(s)"** (DRIFT 종류, V3R6 ID-pattern abort 아님).
- V3R6 abort 신호 `does not match pattern` 발생 횟수 → **0** (패턴 수정이 load-abort를 정확히 제거함을 확인).
- 73건은 사전 존재 CONST-V3R2/V3R5 엔트리의 `clause:` 텍스트가 source 파일의 verbatim substring이 아닌 경우다. 예: `grep -c 'SPEC+EARS format' .claude/rules/moai/workflow/spec-workflow.md` → 0 (source가 "SPEC+GEARS format"으로 이미 변경됨).

**메커니즘**: `LoadRegistry`가 V3R6 ID에서 abort할 때는 post-load DRIFT 루프(validator.go:264)에 **도달하지 못했다**. 패턴 수정으로 레지스트리가 끝까지 로드되면, 그제서야 DRIFT 루프가 실행되어 73건을 발견한다. 이 73건은 본 SPEC이 **un-mask할 뿐 소유하지 않는다** — 73-drift cleanup은 별도 SPEC 후보(`SPEC-V3R6-RULES-CONST-DRIFT-CLEANUP-001`)이며 zone-registry clause/source 편집(SSOT-DEDUP 영역)이다. 아래 Out of Scope 참조.

> **run-phase 구현자 주의 (D3)**: 패턴 수정 후 `moai constitution validate`가 **여전히 exit 1을 반환하는 것은 EXPECTED이며 범위 밖**이다. 73건의 un-masked DRIFT는 본 SPEC이 수정하지 않는다. 본 SPEC의 in-scope 성공 신호는 `grep -c 'does not match pattern'` → **0** (V3R6 load-abort 제거)이지, full-gate exit 0이 아니다. 구현자는 이를 결함으로 오인하지 말 것.

## §B. Requirements (GEARS)

### REQ-RCR-001 (Ubiquitous) — V3R6 ID 허용

The `ruleIDRegexp` shall accept rule IDs in the `CONST-V3R6-NNN` format (NNN = 3 or more digits), in addition to the existing `CONST-V3R2-NNN` and `CONST-V3R5-NNN` formats.

### REQ-RCR-002 (Ubiquitous) — 회귀 무손상 (additive only)

The `ruleIDRegexp` shall continue to accept every `CONST-V3R2-NNN` and `CONST-V3R5-NNN` ID that it accepted before this change, and shall continue to reject `CONST-V3R3-NNN` and `CONST-V3R4-NNN` (absent namespaces). The pattern change is additive: V3R6 newly allowed, nothing previously-allowed removed and nothing previously-rejected newly allowed except V3R6.

### REQ-RCR-003 (Event-driven) — validate load-abort 제거 (necessary, not sufficient)

**When** an operator runs `moai constitution validate` against a project whose deployed `zone-registry.md` contains the `CONST-V3R6-001` entry, the constitution validator shall load the registry **without aborting on the `CONST-V3R6-001` ID-pattern check** — i.e. the `CONST-V3R6-001` entry loads successfully and the post-load DRIFT loop runs to completion. This requirement is **necessary-but-not-sufficient** for full-gate exit 0: any remaining exit-1 is attributable to the 73 pre-existing CONST-V3R2/V3R5 clause DRIFTs that this SPEC un-masks but does NOT own (§A.1). The in-scope success signal is the absence of the V3R6 load-abort, NOT a clean exit code.

### REQ-RCR-004 (Event-detected) — reproduction-first

**When** the run-phase begins (cycle_type=tdd), the implementer shall first add a failing reproduction test asserting that a `CONST-V3R6-001`-style ID is accepted by `Rule.Validate()` / `ruleIDRegexp`, confirm it is RED against the current pattern, then apply the pattern fix and confirm it is GREEN.

### REQ-RCR-005 (Ubiquitous) — doc-comment 정합

The doc comments at `internal/constitution/rule.go:8` and `:22` (currently stating "CONST-V3R2-NNN or CONST-V3R5-NNN format") shall be updated to include the V3R6 namespace, so the comment matches the broadened pattern.

## §C. Acceptance Criteria (inline — Tier S LEAN)

각 AC는 grep/command-verifiable. run-phase에서 독립 재검증.

### AC-RCR-001 — 패턴이 V3R6 허용 + V3R2/V3R5 회귀 무손상

- **Given** the fix is applied to `internal/constitution/rule.go`,
- **When** `Rule.Validate()` is called with `CONST-V3R6-001` (and any other confirmed-existing V3R6 ID),
- **Then** it returns nil (no error), AND `CONST-V3R2-150` / `CONST-V3R5-039` still return nil.
- **Verify**: `grep -n 'ruleIDPattern' internal/constitution/rule.go` shows the broadened pattern; the reproduction test (AC-RCR-002) passes for V3R6 + existing V3R2/V3R5 cases.

### AC-RCR-001b — V3R3/V3R4 계속 거부 (부재 네임스페이스 가드)

- **Given** the fix is applied,
- **When** `Rule.Validate()` is called with `CONST-V3R3-001` or `CONST-V3R4-001`,
- **Then** it still returns an error (these namespaces do not exist and must remain rejected).
- **Verify**: the existing `TestRuleValidateV3R5ID` cases `v3r3-invalid` / `v3r4-invalid` (rule_test.go:147-148) remain GREEN — confirming the chosen pattern is `[256]` (enumeration), NOT `[2-6]` (range that would newly allow V3R3/V3R4).
- **Note (run-phase decision input)**: this AC is the regression guard that disqualifies the `[2-6]` range candidate. Pattern MUST be `^CONST-V3R[256]-\d{3,}$`.

### AC-RCR-002 — reproduction test RED→GREEN

- **Given** a new test case is added to `internal/constitution/rule_test.go` asserting `CONST-V3R6-001` is a valid ID,
- **When** the test runs against the **pre-fix** pattern,
- **Then** it FAILS (RED) — the test must be confirmed RED before the fix.
- **And When** the test runs against the **post-fix** pattern,
- **Then** it PASSES (GREEN).
- **Verify**: `go test ./internal/constitution/ -run 'TestRuleValidate.*V3R6'` (or the new test name) — RED on pre-fix commit, GREEN on post-fix. Capture both outputs.

### AC-RCR-003 (keystone) — `moai constitution validate` load-abort cleared

> **iter-2 reframe (plan-audit D1)**: 원래 "exit 0" 단언은 UNSATISFIABLE이었다 — `[256]` 패턴 적용 후에도 73건의 사전 존재 DRIFT로 인해 validate는 여전히 exit 1이다(§A.1, 결정론적 검증 완료). 본 AC는 load-abort 제거(in-scope)를 단언하며, full-gate exit 0(out-of-scope)을 단언하지 않는다.

- **Given** the fix is applied AND `go build` succeeds,
- **When** `go run ./cmd/moai constitution validate` is executed against this project root,
- **Then** the validator no longer aborts at the `CONST-V3R6-001` ID-pattern check; the V3R6 entry loads successfully and the post-load DRIFT loop runs to completion.
- **Verify (keystone, in-scope)**: `go run ./cmd/moai constitution validate 2>&1 | grep -c 'does not match pattern'` → **0** (the V3R6 load-abort is gone). Paste the observed output into progress.md §E.2.
- **Expected-but-out-of-scope**: `go run ./cmd/moai constitution validate; echo "exit=$?"` will still print `exit=1` because the 73 pre-existing CONST-V3R2/V3R5 clause DRIFTs are now un-masked. This exit-1 is EXPECTED and is NOT a failure of this SPEC — those drifts are owned by a separate SPEC candidate (§A.1). Capture this exit code in progress.md §E.2 with the annotation "73 un-masked DRIFT, out-of-scope".
- **Cross-SPEC (necessary-not-sufficient)**: this AC clears the V3R6 ID-pattern blocker that sibling `VERSION-FORMAT-001 AC-VFM-001a` (the `moai constitution validate` gate) shares. But VERSION-FORMAT-001's gate reaching exit 0 requires THREE things together: (1) this SPEC's V3R6 ID-pattern fix, (2) resolution of the 73 CONST-V3R2/V3R5 clause drifts (separate SPEC), and (3) VERSION-FORMAT-001's own Opus-clause edits. **No single SPEC in this cohort reaches full-gate exit 0 alone.** See §E for the corrected dependency framing.

### AC-RCR-004 — Go toolchain clean

- **Given** the fix + new test are applied,
- **When** the standard Go gate runs,
- **Then** all of the following are clean:
  - `go build ./...` (no compile error)
  - `go vet ./...` (no vet finding)
  - `go test ./internal/constitution/...` (all pass, including pre-existing TestRule* cases)
- **Verify**: paste each command's output into progress.md §E.2. No unrelated package touched.

### AC-RCR-005 — doc-comment 정합

- **Given** the pattern is broadened,
- **When** `internal/constitution/rule.go:8` and `:22` doc comments are read,
- **Then** they mention the V3R6 namespace (not only V3R2/V3R5).
- **Verify**: `grep -n 'V3R6\|V3R2.*V3R5\|namespace' internal/constitution/rule.go` shows updated comments referencing all three namespaces.

### AC-RCR-006 — additive-only behavior (no unrelated change)

- **Given** the change set,
- **When** the diff is reviewed,
- **Then** the ONLY behavioral change is: `CONST-V3R6-NNN` IDs are newly accepted. No other validator behavior (DRIFT detection, FROZEN_WITHOUT_CANARY, SOURCE_FILE_MISSING, zone-class validation) changes.
- **Verify**: `git diff` scope limited to `internal/constitution/rule.go` (pattern + 2 doc comments) and `internal/constitution/rule_test.go` (new test case). No other file modified.

## §D. Exclusions / 범위 밖

이 SPEC은 **validator 패턴 버그만** 수정한다. 아래 항목은 명시적으로 범위 밖이다.

### Out of Scope — zone-registry.md 배포-vs-템플릿 divergence (§25)

- 배포 사본(`CONST-V3R6` 2건)과 템플릿 사본(`internal/template/templates/.claude/rules/moai/core/zone-registry.md`, 0건)의 divergence는 **형제 SPEC CATALOG-SCRUB-001 / SSOT-DEDUP-001의 mirror-parity AC가 소유**한다. 본 SPEC으로 끌어오지 않는다.
- 본 SPEC은 `rule.go`(Go 소스, 비-템플릿 파일)만 수정하므로 Template-First / mirror-parity 정책이 적용되지 않는다.

### Out of Scope — 신규 V3R6 constitution 엔트리 작성

- `CONST-V3R6-NNN` 신규 헌법 엔트리를 zone-registry.md에 추가하는 작업은 범위 밖. 본 SPEC은 **이미 존재하는** `CONST-V3R6-001`이 유효해지도록 패턴만 넓힌다.

### Out of Scope — V3R3/V3R4 네임스페이스 신설

- V3R3/V3R4는 실재하지 않으며, 본 SPEC은 이들을 새로 허용하지 않는다(`[256]` 열거 채택). 미래 네임스페이스 추가는 별도 SPEC.

### Out of Scope — validator의 다른 sentinel 로직 변경

- DRIFT, FROZEN_WITHOUT_CANARY, SOURCE_FILE_MISSING, INVALID_ZONE_CLASS 검사 로직은 손대지 않는다(REQ-RCR-002 additive-only 보증).

## §E. Cross-References

- 형제(blocked-on-this): `SPEC-V3R6-...-VERSION-FORMAT-001` — AC-VFM-001a (`moai constitution validate` 게이트)는 본 SPEC의 AC-RCR-003이 GREEN이 될 때까지 RED baseline 위에 있다. **본 SPEC은 VERSION-FORMAT-001의 validator-gated AC들의 precondition이다.**
- 형제(owns §25 divergence): CATALOG-SCRUB-001 / SSOT-DEDUP-001 — zone-registry deployed-vs-template mirror AC.
- 구현 대상: `internal/constitution/rule.go` (패턴 + doc comment), `internal/constitution/rule_test.go` (reproduction test).
- 영향 받는 진입점: `internal/constitution/validator.go:178` `LoadRegistry` → `Validate`; `internal/cli/constitution.go:24` CLI load-path.
- run-phase cycle_type: **tdd** (Go 소스 변경, reproduction-first).
