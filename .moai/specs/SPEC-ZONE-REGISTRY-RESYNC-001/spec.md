---
id: SPEC-ZONE-REGISTRY-RESYNC-001
title: "zone-registry clause/anchor 재동기화 + 재발 차단 가드"
version: "0.4.0"
status: draft
created: 2026-08-24
updated: 2026-08-25
author: manager-spec
priority: P1
phase: "v3.1.3 target"
module: ".claude/rules/moai/core/zone-registry.md, internal/template/templates/.claude/rules/moai/core/zone-registry.md, internal/constitution, .github/workflows/ci.yml"
lifecycle: spec-anchored
tags: "zone-registry, constitution, drift, template-mirror, ci-guard, anchor-rot"
tier: M
era: V3R6
related_specs: [SPEC-V3R6-ZONE-REGISTRY-PACKAGING-001, SPEC-V3R5-CONSTITUTION-DUAL-001]
---

# SPEC: zone-registry clause/anchor 재동기화 + 재발 차단 가드

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-08-24 | manager-spec | 최초 작성 — clause 재동기화 + 가드, anchor 는 보조 축 |
| 0.2.0 | 2026-08-24 | manager-spec | 범위 확장 승인 반영 — anchor 탐지·수리를 정식 축으로 승격(죽은 `SentinelAnchorNotFound` 완결), slug 규칙을 요구사항화, 마일스톤 순서를 의존성으로 고정 |
| 0.3.0 | 2026-08-25 | manager-spec | plan-audit iter1(FAIL 0.75) 반영 — 자기참조 `file:` / 빈 clause / `\|\| true` / 부분 순회 mutant 4종을 AC 로 승격, 이중 트리 파일 수 오기 정정, paths-filter 구속 추가 |
| 0.4.0 | 2026-08-25 | manager-spec | plan-audit iter2(PASS-WITH-DEBT 0.925) 부채 #1 마감 — 빈 clause 금지를 **유일 적중** 요구로 강화(공백 한 칸·짧은 토큰 우회 차단), REQ 번호를 전역 오름차순으로 재배치(구 015 slug 규칙 → **012**, 구 012/013/014 → 013/014/015), §D8 인용 정정, 잔여 부채 2건(`CONST-V3R2-004` 근접 오답 / 평가 수 이중 카운트)을 §H 에 판정자와 함께 기록 |

## 1. 문제 — 측정된 형태

갓 `moai init` 한 프로젝트가 첫 `moai doctor` 에서 실패를 보고한다.

측정 트리: `.claude/worktrees/t232` @ `294b4b6ab`. 그 트리에서 빌드한 `bin/moai` 로 스크래치 프로젝트를 초기화하고 측정했다. 전문 출력은 `.moai/reports/t232/validate-repro.txt`.

```console
$ moai constitution validate ; echo exit=$?
Constitution validate: found 67 error(s).
exit=1

$ moai doctor | tail -2
  fail    Constitution Registry  registry loads (101 entries) but validate found 67 error(s)
  Pass 22    Warn 2    Fail 1
```

67건 전부가 동일 sentinel `[DRIFT] clause "…" not found in source "…"` 이다.

**이 숫자는 상수가 아니라 특정 트리의 측정치다.** 이슈 #1616 이 적은 65는 `v3.1.3-rc.0` 에서 잰 값이고, `294b4b6ab` 에서는 **101건 중 67건**이다. 규칙 문서가 움직일 때마다 값이 바뀌므로, 이 SPEC 은 어디서도 67을 목표값·판정 기준으로 쓰지 않는다 — AC의 GREEN 조건은 전부 **"드리프트 0"** 이고, 67은 "이 체크가 지금 무언가를 관측하고 있다"를 보이는 RED 근거로만 등장한다(측정 트리 SHA 를 항상 함께 적는다).

### 1.1 착지 후 재측정 — 67 → 63 (수리 대상은 줄지 않았다)

PR #1611(은퇴 엔트리 skip)이 `bf6083f13` 으로 main 에 들어간 뒤 재측정했다. 측정 트리: `WT-zone-registry-drift` @ `1ae6e5c36`(= `294b4b6ab` + origin/main 통합). 전문 출력은 `.moai/reports/t232/validate-postmerge.txt`.

```console
$ moai constitution validate ; echo exit=$?
Constitution validate: found 63 error(s).
exit=1
  4 retired entry/entries skipped ([SUPERSEDED …] marker); re-check them with --strict
```

**검증기 보고치만 줄었고, 실제로 깨진 clause 수는 그대로다.** 독립 분석기는 전과 동일하게 **clause 실패 68 / anchor 실패 17 / clause 통과·anchor 실패 9** 를 낸다. 차이 4건은 `CONST-V3R2-021/022/023/024` — 은퇴 마커를 단 엔트리들이고, #1611 이 이들을 비-strict 경로에서 건너뛰게 만들었을 뿐이다. 수리 범위는 **68건 그대로**이며 63 은 "지금 검증기가 보고하는 수"일 뿐이다.

### 1.2 열린 충돌 — 은퇴 엔트리와 AC-ZRR-002/003

위 재측정이 **SPEC 내부 충돌 하나를 드러냈다. run-phase 착수 전에 결정해야 한다.**

- `AC-ZRR-002/003` 은 **101건 전부**가 자기 `file:` 안에서 정확히 1회 적중할 것을 요구한다.
- #1611 은 은퇴 엔트리를 검증에서 제외하며, 그 근거는 "은퇴한 절이 인용하는 원문은 정의상 사라졌으므로 verbatim 검사는 실패만 할 수 있다" 이다.

측정 결과 은퇴 4건은 **전부 clause 실패**다(`clause_ok=False`, anchor 는 4건 모두 통과). 즉 두 계약이 같은 4건에 대해 반대를 말한다 — AC 는 고치라 하고, #1611 은 고칠 수 없는 것이라고 한다.

가능한 해소는 셋이며, **이 SPEC 은 아직 고르지 않았다**:

| 안 | 내용 | 대가 |
|---|---|---|
| A | AC 를 "은퇴 엔트리를 제외한 97건"으로 축소 | 은퇴 4건의 `file:`/`anchor:` 가 영구히 검증 밖에 남는다(anchor 는 지금 통과하지만 다음 편집에서 깨져도 아무도 모른다) |
| B | 은퇴 엔트리도 verbatim 을 요구 | #1611 의 전제와 정면 충돌. 사라진 원문을 인용하라는 요구가 된다 |
| C | 은퇴 엔트리는 clause 검사에서 빼되 **anchor 검사는 유지** | 계약이 갈라져 구현이 복잡해지지만, 두 근거를 모두 존중한다 |

C 가 유력해 보이나 근거를 더 재야 한다 — 은퇴 엔트리의 `anchor:` 가 무엇을 가리켜야 하는지(사라진 절인가, 그 절을 대체한 절인가)가 아직 측정되지 않았다.

사용자가 받는 레지스트리는 템플릿과 **바이트 동일**하다.

```console
$ diff -q .claude/rules/moai/core/zone-registry.md \
          internal/template/templates/.claude/rules/moai/core/zone-registry.md
(차이 없음)
```

`.claude/rules/moai` 는 `CleanMoaiManagedPaths`(`internal/cli/update/deploy/deploy.go`)가 관리하는 뿌리라, 사용자가 로컬에서 고쳐도 다음 `moai update` 가 템플릿 원본으로 덮는다. **수리는 반드시 템플릿 소스에서 이뤄져야 한다.**

## 2. 원인 — 세 층

### 2.1 matcher 는 정상이다. 데이터가 인용이 아니다

`internal/constitution/validator.go` 의 DRIFT 판정은 문서한 대로만 한다 — `normalizeWhitespace(clause)` 가 `normalizeWhitespace(stripCodeFences(source))` 의 부분문자열이어야 한다. 이 규칙을 독립 재구현(`.moai/reports/t232/analyze.py`)해 돌리면 검증기와 **동일한 ID 집합**이 나온다(analyzer 68 vs validator 67, 차이 1건은 analyzer 쪽 YAML 파싱 아티팩트로 확인됨).

67건의 실체는 두 갈래다.

- **약 61건은 패러프레이즈.** `clause:` 값이 교리를 요약한 문장이지 인용이 아니다. `CONST-V3R2-008` 의 clause `"Language-Aware Responses: All user-facing responses MUST be in user's conversation_language…"` 는 어떤 소스 파일에도 그 형태로 존재한 적이 없다. 정확 부분문자열 매처 아래에서 **구조적으로 통과 불가능한 값**이다.
- **6건은 짧은 요약 라벨.** `SPEC+EARS format`, `@MX TAG protocol`, `16-language neutrality`, `Template-First discipline`, `AskUserQuestion monopoly`, `Claude Code substrate`. 같은 원인의 짧은 형태.

즉 이 SPEC은 "매처를 데이터에 맞추는" 작업이 아니라 **데이터를 계약에 맞추는** 작업이다.

### 2.2 검증기가 보지 못하는 두 번째 결함 — anchor 17건

`Validate` 는 `clause` 만 본다. `anchor:` 를 해당 `file:` 의 heading 에 대고 해석하는 코드는 없다. 독립 해석 결과 **101건 중 17건이 자기 `file:` 안의 어떤 heading slug 와도 맞지 않는다.**

**anchor 축을 범위에 넣는 결정적 근거는 그중 9건이다.** 이 9건은 clause 가 통과한다. 즉 clause 만 고치는 수리를 마치면 이들은 `validate` 초록, doctor 초록, 사람 눈에도 "고쳐진 것처럼" 보이면서 실제로는 여전히 깨진 포인터를 들고 있다. **고쳐진 것처럼 보이는 결함**은 남아 있는 결함보다 나쁘다 — 다음 사람이 다시 들여다볼 이유를 없애기 때문이다.

측정된 9건: `CONST-V3R5-004/005/006/007/008/010/011/012/013` (전부 `ci-autofix-protocol.md`). 원인은 heading 개명이다.

| registry anchor | 현재 heading |
|---|---|
| `#iteration-limit` | `## Iteration Cap` |
| `#commit-strategy` | `## Patch Commit Rule — No Force-Push` |
| `#user-interaction-channel` | `## AskUserQuestion Boundary` |
| `#semantic-failure-handling` | `## Semantic Failure — No Auto-Patch` |
| `#audit-log` | `## Audit Log Requirement` |
| `#protected-files` | `## CI Infrastructure Preservation` |
| `#ci-auto-fix-loop-entry-condition` | `## Entry Condition` |

나머지 8건(`CONST-V3R2-003/008/009/010/011/028/029`, `CONST-V3R5-009`)은 clause 도 함께 깨져 있다. `CLAUDE.md` §1 이 세 불릿으로 압축되면서 `CONST-V3R2-008..011` 의 원문 교리는 `moai-constitution.md` 의 `## Response Language` / `## Parallel Execution` / `## Output Format` 으로 이사했다 — 교리가 사라진 게 아니라 포인터가 낡았다.

#### anchor 검사는 신규 범위가 아니라 미완의 완결이다

검증기는 이 검사를 **이미 이름 붙여 두었다.**

```go
// internal/constitution/validator.go:27-28
// SentinelAnchorNotFound is used when the anchor in a registry entry does not exist in the source file.
SentinelAnchorNotFound = "ANCHOR_NOT_FOUND"
```

리포 전체에서 `SentinelAnchorNotFound` 를 grep 하면 **위 두 줄(주석 + 선언)이 전부**다. 문자열 `ANCHOR_NOT_FOUND` 도 그 한 줄 외에 어디에도 없다. 설계되고 명명됐지만 `Validate` 에 배선된 적이 없는 **죽은 sentinel** — 즉 anchor 검사를 넣는 것은 새 요구를 발명하는 게 아니라 **의도됐으나 구현되지 않은 검사를 완결**하는 일이다.

#### slug 규칙이 곧 요구사항이다

재사용할 수 있는 markdown heading slug 헬퍼가 리포에 **없다.** 이름이 비슷한 것들은 전부 다른 것을 slug 화한다 — `internal/web/render_helpers.go` 의 `i18nSlug`(i18n 키), `internal/cli/preference/filestore.go` 의 `slugify`(경로), `internal/hook/session_end.go` 의 `projectSlug`(프로젝트 경로). heading 을 다루는 것은 하나도 없다.

따라서 anchor 검사를 배선하려면 **heading→slug 규칙을 명시적으로 정의해야 하고, 그 규칙 선택 자체가 요구사항이다** — 규칙이 다르면 실패 건수가 달라지기 때문이다. 위 17건은 아래 규칙 아래에서 측정된 값이다(`.moai/reports/t232/analyze.py`).

1. 코드 펜스(```` ``` ````) 안의 행은 heading 후보에서 제외한다
2. `#` 접두를 벗기고 앞뒤 공백을 제거한다
3. 백틱(`` ` ``)을 제거한다
4. 소문자화한다
5. `[a-z0-9]`, 공백, `-` 이외의 문자를 제거한다
6. 연속 공백을 단일 `-` 로 바꾸고 앞에 `#` 를 붙인다

### 2.3 조용히 썩은 이유 — `validate` 를 아무도 돌리지 않는다

```console
$ grep -c "constitution validate" Makefile .github/workflows/*.yml
0
```

`make constitution-check`(Makefile:75-78)와 CI `constitution-check` job(.github/workflows/ci.yml:445-475) 은 둘 다 `constitution **list**` 만 돌린다 — 레지스트리를 파싱만 한다. `constitution validate` 를 부르는 곳은 Makefile 에도, CI 에도, 어떤 Go 테스트에도 없다. 그 부재가 카탈로그의 3분의 2가 깨질 때까지 붉은 신호가 한 번도 안 뜬 이유다.

수리 경로에 대해서도 하나 못박아 둔다. `internal/constitution/pipeline.go:256-267` 의 `updateSourceFile` / `updateRegistryClause` 는 둘 다 `not yet implemented` 를 반환하는 스텁이다. 이 SPEC 의 수리는 **템플릿 레지스트리 파일과 그 로컬 미러에 대한 직접 편집**이며, 저 스텁을 구현하거나 호출하는 경로는 쓰지 않는다.

덧붙여 CI `constitution-check` job 은 `continue-on-error: true` 다. **이 job 에 `validate` 를 한 줄 끼워 넣는 것만으로는 가드가 되지 않는다** — 실패해도 PR 이 막히지 않는다. 가드 설계에서 이 사실이 구속 조건이다.

## 3. 이 카드의 대표 mutant

"validate 가 exit 0" 만 요구하는 AC는 **매처를 약화시키는 구현**으로 만족된다: fuzzy/토큰 중첩 매칭, 짧은 clause 스킵, 레지스트리 엔트리 삭제, `canary_gate` 뒤집기, 파일 제외 목록, `MOAI_CONSTITUTION_SKIP_VALIDATE=1` 상시화. 이런 구현은 AC를 통과하면서 낡은 포인터를 전부 그대로 두고, 드리프트 탐지기 자체를 파괴한다.

따라서 이 SPEC 의 AC는 전부 mutant 내성으로 작성한다 — 매처 불변 고정, 엔트리 수 보존, **수리 대상과 독립인 도구로의 검증**, 그리고 모든 체크에 사전 구현(RED) 기준값 명시. 상세는 `acceptance.md`, mutant 대응표는 `plan.md` §G.

## 4. 요구사항 (GEARS)

### 데이터 수리

- **REQ-ZRR-001** (Ubiquitous) — The zone registry shall carry, for every entry, a `clause:` value that is a contiguous single-line verbatim span occurring **exactly once** in the file named by that entry's `file:`, in both the local rules tree and the shipped template tree.
- **REQ-ZRR-002** (Ubiquitous) — The zone registry shall carry, for every entry, an `anchor:` value that resolves to a heading present in the file named by that entry's `file:`.
- **REQ-ZRR-003** (Event-driven) — **When** the doctrine cited by an entry has moved to a different file, the entry shall be re-pointed by correcting `file:` and `anchor:`, not by rewording `clause:` into a summary.

### 보존 (수리가 파괴해서는 안 되는 것)

- **REQ-ZRR-004** (Unwanted) — The repair shall not alter the DRIFT matching semantics of `internal/constitution/validator.go` (`normalizeWhitespace`, `stripCodeFences`, and the substring check in `Validate`).
- **REQ-ZRR-005** (Unwanted) — The repair shall not add, delete, or renumber registry entries, and shall not change any entry's `zone`, `zone_class`, or `canary_gate` value.
- **REQ-ZRR-006** (Unwanted) — The repair shall not introduce any skip, exclusion, threshold, or fuzzy-match path into the validation route (clause-length threshold, per-file exclusion list, token-overlap matching, or an expanded environment bypass).

### 가드

- **REQ-ZRR-007** (Event-driven) — **When** an edit to the rules tree or to the registry breaks any entry's clause or anchor, the guard shall fail the pull request containing that edit.
- **REQ-ZRR-008** (Ubiquitous) — The guard shall evaluate both registry mirrors: the local `.claude/rules/moai/core/zone-registry.md` against the local rules tree, and `internal/template/templates/.claude/rules/moai/core/zone-registry.md` against the template rules tree.
- **REQ-ZRR-009** (Ubiquitous) — The guard shall verify anchor resolution in addition to clause verbatim-ness, completing the check already named by the unwired `SentinelAnchorNotFound` sentinel.
- **REQ-ZRR-010** (Unwanted) — The guard shall not report success when validation was bypassed; **when** `MOAI_CONSTITUTION_SKIP_VALIDATE=1` is present in its environment, the guard shall fail rather than pass.
- **REQ-ZRR-011** (Unwanted) — The guard shall not run only in a `continue-on-error: true` CI job, and the step that runs it shall not be wrapped in `|| true`, a step-level `continue-on-error`, or any other exit-code suppression.
- **REQ-ZRR-012** (Ubiquitous) — The guard shall declare, in code, the heading-to-slug rule it applies (the six steps in §2.2), with a comment naming it as the rule the anchor failure count was measured under.

### 배포 규율

- **REQ-ZRR-013** (Ubiquitous) — The local registry and the template registry shall remain byte-identical, and the binary shall be re-embedded via `make build` in the same change.
- **REQ-ZRR-014** (Unwanted) — The template registry shall not carry SPEC IDs, internal dates, or commit SHAs.
- **REQ-ZRR-015** (Event-driven) — **When** `moai constitution validate` runs in a freshly initialized project built from this change, it shall exit 0, and `moai doctor` shall report no Constitution Registry failure.

## 5. 알려진 구속 조건 — 두 트리에서 동시에 verbatim 이어야 한다

레지스트리는 두 트리에 바이트 동일하게 배포되는데, 인용 대상 파일 17개 중 **2개 파일이 로컬과 템플릿에서 서로 다르고, 그 2개가 3개 엔트리를 물고 있다**(측정: `.moai/reports/t232/divergence.py`).

| 파일 | 영향 엔트리 |
|---|---|
| `.claude/rules/moai/development/coding-standards.md` | `CONST-V3R2-004`, `CONST-V3R2-005` |
| `.claude/rules/moai/development/skill-authoring.md` | `CONST-V3R5-038` |

이 3개 엔트리의 clause 는 **두 판본 모두에 존재하는 텍스트 구간**에서 골라야 한다. 공통 구간이 없으면 그것은 clause 선택 문제가 아니라 미러 드리프트 결함이며, 우회(엔트리 제외·매처 완화)가 아니라 blocker 로 올려야 한다. 나머지 15개 파일(엔트리 98건)은 두 트리에서 동일하므로 이 제약을 받지 않는다.

**실측 결과 공통 구간은 존재한다**(plan-audit iter1 이 두 파일을 전수 diff): 발산은 `coding-standards.md` 1줄(`git commit --no-verify` 불릿) + `skill-authoring.md` 3줄(SPEC-ID 중립화) = 총 4줄이고, 셋 다 영향 엔트리가 가리키는 절(`#language-policy` / `#thin-command-pattern` / `#key-format-rules`)과 무관하다. 즉 blocker 로 올라올 가능성은 사실상 없다 — run-phase 는 이 조사를 반복하지 않는다.

## 6. 범위 밖 (Non-goals)

### Out of Scope — 다른 카드가 소유한 축

- `[SUPERSEDED]` 표기 및 `canary_gate: false` 처리 축(#1595) — 카드 t201 소관. 이 SPEC 은 어떤 `canary_gate` 값도 바꾸지 않는다(REQ-ZRR-005).

### Out of Scope — 별도 설계가 필요한 축

- 프로젝트 로컬 오버레이 기구(이슈 #1616 의 2안: 사용자가 자기 프로젝트에서 레지스트리를 확장·재정의하는 층) — 별도 설계 사안이다. 후속 후보로만 기록하고 여기서 명세하지 않는다.

### Out of Scope — 하지 않는 일

- 규칙 문서 본문 수정. 이 SPEC 은 포인터(레지스트리)를 문서에 맞추는 것이지, 문서를 포인터에 맞추지 않는다. 인용을 만들려고 규칙 본문을 고치는 것은 범위 밖이다.
- 매처 개선·성능·리팩터링. `validator.go` 의 DRIFT 경로는 불변이 요구사항이다(REQ-ZRR-004).
- 레지스트리 커버리지 확장(새 HARD 절 등록). 엔트리 수는 101 그대로다(REQ-ZRR-005).
- `moai update` 가 로컬 전용 파일을 지우는 결함(CLAUDE.local.md §2.3) — 별개 결함, 별개 카드.

## 7. 미검증 항목 (Gaps)

- **개별 67건의 "올바른 인용문"은 아직 정해지지 않았다.** 이 SPEC 은 계약(단일 행 verbatim, anchor 해석 가능)만 정한다. 어떤 문장을 뽑을지는 run-phase 판단이며, 그 판단의 검증은 AC-ZRR-002/003/004 가 기계적으로 한다.
- **slug 규칙은 선택된 것이지 검증된 것이 아니다.** §2.2 의 6단계 규칙은 `analyze.py` 가 쓴 규칙이며, Claude Code / markdown 렌더러의 실제 anchor 해석과 바이트 단위로 같은지는 확인하지 않았다. 17이라는 수치는 **이 규칙 아래에서만** 참이다 — 그래서 REQ-ZRR-012 가 규칙을 코드에 명시하도록 요구한다. 규칙이 바뀌면 건수도 바뀐다는 사실 자체를 문서화하는 것이 여기서의 방어다.
- **fresh-init 재현은 1회 측정**이다(스크래치 프로젝트 1개, `--language go`). 다른 언어 옵션에서 레지스트리가 달리 배포되는지는 확인하지 않았다.
- **DRIFT 67 은 `294b4b6ab` 시점의 값이며, PR #1611 착지 전 측정이다.** #1611 은 `IsRetiredClause` 의 은퇴 판정 경로를 바꾸므로 은퇴 4건이 검증에서 빠져 값이 움직일 수 있다. run-phase 는 착지 후 재측정하고 **측정 트리 SHA 를 병기**한다. 재측정 값이 67 과 같더라도 **"착지 후 재측정 결과 동일"을 명시**한다 — 값이 변하지 않은 것과 재측정을 하지 않은 것은 문서에서 구별되지 않는데, 앞은 근거이고 뒤는 갭이다.
- **~~3건 이중 트리 제약(§5)의 해소 가능성 미확인~~ → 닫힘.** plan-audit iter1 이 두 파일을 전수 diff 한 결과 발산은 총 4줄(`coding-standards.md` 1 + `skill-authoring.md` 3)이고 전부 영향 엔트리의 인용 대상 절 밖이다 — 공통 구간 존재가 확인됐다(§5 참조). run-phase 는 이 조사를 반복하지 않는다.
- **slug 규칙이 렌더러의 실제 anchor 해석과 같은지도 감사에서 확인되지 않았다.** plan-audit iter1 이 이 항목을 자기 gap 으로 명시했다 — 17이라는 수치는 여전히 `analyze.py` 재구현 기준이다.
