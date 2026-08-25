---
id: SPEC-ZONE-REGISTRY-HARDEN-001
title: "zone-registry 감사 후속 F1-F3 — clause 완결화 · 튜플 digest pinning · 문서 의미론 정렬"
version: "0.2.0"
status: completed
created: 2026-08-25
updated: 2026-08-25
author: manager-spec
priority: P2
phase: "v3.1.3 target"
module: ".claude/rules/moai/core/zone-registry.md, internal/template/templates/.claude/rules/moai/workflow/ci-autofix-protocol.md, internal/constitution, .moai/specs/SPEC-ZONE-REGISTRY-RESYNC-001/plan.md"
lifecycle: spec-anchored
tags: "zone-registry, constitution, registry-guard, digest-pinning, clause-canary, doc-consistency"
tier: M
era: V3R6
related_specs: [SPEC-ZONE-REGISTRY-RESYNC-001]
depends_on: [SPEC-ZONE-REGISTRY-RESYNC-001]
---

# SPEC: zone-registry 감사 후속 F1-F3 — clause 완결화 · 튜플 digest pinning · 문서 의미론 정렬

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-08-25 | manager-spec | 최초 작성 — t232 sync-audit verdict(2026-08-25, PASS 0.92)의 optional Findings F1/F2/F3를 카드 t268 Tier M 번들로 구조화. 측정 기준 트리 `db1362739`(= origin/main, 워크트리 t268)에서 대상 위치를 내용 기반으로 재좌표 |
| 0.2.0 | 2026-08-25 | manager-spec | plan-audit iter1(PASS-WITH-DEBT 0.825) 결함 3건 적용 — D1 AC-ZRH-009 신설(REQ-ZRH-006 커버리지), D2 M1 anchor 값 정정(`#semantic-failure-no-auto-patch`), D3 M3 소유권 노트(오케스트레이터 재위임 아래 manager-spec 집행). status draft 유지 |

## 1. 문제 — 측정된 형태

전제: SPEC-ZONE-REGISTRY-RESYNC-001(card t232, PR #1646 머지 `39c677f47`)은 PASS 0.92로 종결됐다. 종결 감사 `.moai/reports/t232/sync-audit-verdict-2026-08-25.md`의 Findings 중 F1/F2/F3는 전부 [minor][optional]·비차단이며, 감사관 권고 2항이 "후속 카드 1장에 F1(clause 재선택) + F2(ID digest pinning) + F3(문서/의미론 정리) 묶기"를 명시했다. 본 SPEC이 그 후속 카드(t268)다.

아래 좌표는 전부 본 워크트리 트리 `db1362739`에서 **내용으로 재좌표**한 값이다(감사 보고서의 행 번호는 t232 트리 기준 — 행갈이가 다를 수 있어 그대로 믿지 않는다).

### 1.1 F1 — clause가 행갈이 지점에서 절단 (CONST-V3R5-010)

- 레지스트리: `.claude/rules/moai/core/zone-registry.md:763`(`- id: CONST-V3R5-010`), clause 값은 `:768`
- 현재 clause: `"Semantic failures (data race, deadlock, panic, test assertion failure) MUST"`
- 핀된 원본: `.claude/rules/moai/workflow/ci-autofix-protocol.md:106-108`(레지스트리 `file:` 필드가 지목하는 template-script-free 트윈)

```
106: [ZONE:Frozen] [HARD] Semantic failures (data race, deadlock, panic, test assertion failure) MUST
107: NOT be automatically patched. The orchestrator MUST immediately escalate via
108: AskUserQuestion with the diagnosis report.
```

clause가 106행 끝(`MUST`)에서 잘려 있어, **저장 텍스트만 읽으면 MUST-NOT 금지가 MUST 의무로 읽힌다.** 단일 행 verbatim·유일 적중 계약은 만족하므로 기계 AC 위반이 아니라 데이터 품질 흠이다. 감사관이 제시한 재선택 예: `The orchestrator MUST immediately escalate via AskUserQuestion with the diagnosis report.` — 이 문장은 현재 107-108행에 걸쳐 있어(단일 행 적중 0회, 본 트리实测 `grep -c -F` → 0) **원본 단락의 서식 전용 rewrap이 선행되어야 한다.**

### 1.2 F2 — 가드가 엔트리 수(101)만 pin (개수 보존 치환 탈출)

- 가드: `internal/constitution/registry_sync_test.go` — 상수 `:45`(`wantRegistryEntries = 101`), 개수 단정 `:135-137`
- 탈출 구멍: 엔트리 **삭제**는 count로 잡히지만, **개수를 보존하는 ID 치환·zone/zone_class/canary_gate 치환**은 현재 아무 검사에도 걸리지 않는다(file/anchor/clause는 각자의 검사층이 있으나 id/zone/zone_class/canary_gate 4필드는 무방비).
- 감사관 조치: 정렬된 `(id, zone, zone_class, canary_gate)` 튜플 집합의 digest pinning.

### 1.3 F3 — plan.md 문서 의미론 vs 구현·acceptance 의미론 불일치

- 문서: `.moai/specs/SPEC-ZONE-REGISTRY-RESYNC-001/plan.md:108` — 리터럴 체크 설계 문단이 `strings.Count(rawFileContent, clause)`(**발생 의미론**: 같은 행 2회 적중 = 2)로 기술
- 구현: `internal/constitution/registry_sync_test.go:363-379` `literalHitCount` — 주석이 명시적으로 "the `grep -F -c` equivalent: the number of **LINES** in raw that contain clause"(**라인 수 의미론**: 같은 행 2회 적중 = 1)
- acceptance: `SPEC-ZONE-REGISTRY-RESYNC-001/acceptance.md:50` — `grep -F -c` 로 이미 구현과 일치
- 현 데이터는 양쪽 의미론에서 once=97(감사 독립 측정)이라 양쪽 통과 — 동작 결함이 아니라 **설계 문단의 서술 오기**다.

## 2. 요구사항 (GEARS)

REQ 접두사: `REQ-ZRH` (Zone-Registry Harden).

| ID | Pattern | Requirement |
|----|---------|-------------|
| REQ-ZRH-001 | Ubiquitous | The CONST-V3R5-010 clause shall be a self-contained complete sentence — a reader of the stored clause value alone shall receive the rule's full obligation without any text outside the quoted span. |
| REQ-ZRH-002 | Unwanted | The CONST-V3R5-010 clause shall not terminate mid-predicate at a line-wrap boundary of the pinned source file (the `…MUST` half of the original `MUST NOT`). |
| REQ-ZRH-003 | Ubiquitous | The source rewrap that places the selected sentence on a single line shall be formatting-only — the whitespace-normalized content of both ci-autofix-protocol.md twins shall be byte-identical before and after the change (base `db1362739`). |
| REQ-ZRH-004 | Ubiquitous | The registry-sync guard shall pin a SHA-256 hex digest over the sorted `(id, zone, zone_class, canary_gate)` tuple lines of all registry entries, asserted per mirror alongside the existing entry-count pin. |
| REQ-ZRH-005 | Event-detected | When a registry mutation preserves the entry count while altering any pinned tuple field (ID substitution or zone/zone_class/canary_gate substitution), the digest assertion shall fail. |
| REQ-ZRH-006 | Ubiquitous | The digest failure message shall print the computed digest and the same-change update procedure, so deliberate registry growth updates `wantRegistryEntries` and the digest constant together. |
| REQ-ZRH-007 | Ubiquitous | The SPEC-ZONE-REGISTRY-RESYNC-001 plan.md literal-check paragraph shall document the implemented line-count semantics (`grep -F -c` equivalent — the number of LINES containing the clause as a literal substring), replacing the `strings.Count` occurrence-semantics description, with a provenance annotation naming this SPEC. |
| REQ-ZRH-008 | Ubiquitous | After all three fixes, the pre-existing guard contracts shall hold unchanged — validator drift 0, literal buckets once=97 / zero=0 / multi=0, retired_exempt=4, anchor checks 101/101, and the two zone-registry.md mirrors byte-identical. |

### 2.1 사용자 스토리

레지스트리 유지보수자(GOOS)는 sync-auditor가 남긴 3건의 optional 흠을 별도 왕복 없이 한 번의 카드로 청산하고 싶다. F1 후 레지스트리를 읽는 사람은 잘린 반쪽 의무문이 아니라 완결된 규칙 문장을 읽고, F2 후 개수 보존 치환 변이는 차단 가드에 걸리며, F3 후 폐쇄된 SPEC의 설계 문단은 구현·acceptance가 실제로 재는 의미론과 같은 말을 한다.

## 3. 제약

- **범위는 F1-F3 정확히 세 축** — 인접 개선·레지스트리 재설계·각 finding이 명시하지 않은 추가 hardening 금지. 세 finding 모두 감사에서 optional·비차단이었으므로 scope 확장은 금지다.
- **매처 불변**: `internal/constitution/validator.go` 는 무편집(RESYNC-001 D1 계약 계승 — `git diff 1ae6e5c36..HEAD -- validator.go` = 0라인이 이후에도 유지되어야 한다).
- **트윈·미러 규율**: ci-autofix-protocol.md(배포판↔템플릿 원본)과 zone-registry.md(로컬↔템플릿) 각각 한쪽만 편집하고 끝내지 않는다 — 편집 후 양쪽 `cmp` 바이트 동일.
- **Template-First**: 템플릿 경로 파일은 `internal/template/templates/` 원본을 먼저 편집 → `make build` → 배포판 동기. 템플릿 neutrality 검사(SPEC-ID·날짜·SHA 등 C3/C7 금지 클래스)를 유발하지 않는 서식 전용 변경이어야 한다.
- **개발 방식**: `constitution.development_mode: tdd` — F2 신규 검사 코드는 RED(변이가 통과해버리는 실패)를 먼저 관측한 뒤 GREEN.
- **개수 불변**: 본 SPEC은 레지스트리 증감을 하지 않는다(`wantRegistryEntries = 101` 유지, digest는 현 101엔트리 집합을 pin).

## 4. Tier 분류

**Tier M** — 영향 파일 6개(아래 §5 참조, 미러 포함)로 Tier M 파일 범위(5-15)에 부합. 신규 LOC는 가드 보강 ~80행 + 데이터/문서 편집 소량으로 Tier S의 "<300 LOC·<5 files"를 파일 수에서 벗어난다. REQ 8 / AC 9 — Tier M 상한(16/16) 이내. 카드 t268의 Tier M 지정과 일치.

## 5. 변경 대상 파일 (전부 `db1362739` 기준 재좌표)

| # | 파일 | 변경 |
|---|------|------|
| 1 | `.claude/rules/moai/core/zone-registry.md` (`:768`) | F1 — CONST-V3R5-010 clause 재선택 |
| 2 | `internal/template/templates/.claude/rules/moai/core/zone-registry.md` (`:768`) | F1 — 1번과 동일 편집(미러) |
| 3 | `internal/template/templates/.claude/rules/moai/workflow/ci-autofix-protocol.md` (`:106-108`) | F1 — 서식 전용 rewrap(템플릿 원본 선행) |
| 4 | `.claude/rules/moai/workflow/ci-autofix-protocol.md` (`:106-108`) | F1 — 3번과 동일 편집(트윈) |
| 5 | `internal/constitution/registry_sync_test.go` (`:42-53`, `:135-137` 인근) | F2 — digest 상수·계산·단정·변이 서브테스트 |
| 6 | `.moai/specs/SPEC-ZONE-REGISTRY-RESYNC-001/plan.md` (`:108`) | F3 — 의미론 정정 + 정정 출처 주석 |

## 6. Out of Scope

### Out of Scope — F4 프로세스 부채 (별도 카드)

- 부활(정지 티메이트 재기동) 재발 방지 격리 강화 — 감사관 권고 3항이 별도 후속 카드를 명시
- 활성 감사 중 워크트리 단독-작성자 규율 — 위와 같은 별도 카드

### Out of Scope — F5 CI 초록 관측

- 판정 시점 PR 헤드 CI 25종 pending 관측 — 머지 시점 확인 사항이지 본 SPEC 작업이 아님

### Out of Scope — 가드 의미론 전환

- `literalHitCount`의 발생(occurrence) 의미론 전환 — 감사가 "문서 정리 **또는** 의미론 전환"의 양안을 제시했으나 본 SPEC은 문서 정렬안을 채택. 구현·acceptance는 이미 라인 수 의미론으로 정합하므로 동작 변경 없이 해소된다

### Out of Scope — 인접 표면 무편집

- `.claude/rules/local/ci-autofix-protocol.md`(dev 원본)의 같은 문장 rewrap — 레지스트리가 pin하는 파일이 아니며 어떤 기계 계약도 이 파일의 행 구조에 의존하지 않는다(본 트리 grep·registry `file:` 전수 확인)
- CONST-V3R5-010 외 다른 엔트리의 clause "개선" — F1은 이 엔트리 1건만 지목
- digest에 file/anchor/clause 필드 추가 — 각 필드는 이미 고유 검사층(verbatim·anchor 해석)이 지키며, 4필드 외 확장은 finding이 명시하지 않은 hardening
- `moai constitution validate` OK 경로의 `(0 entries checked)` 하드코딩(`internal/cli/constitution.go:386`) — 감사가 별계 결함으로 기록·존체
- zone-registry 재설계(스키마 변경·엔트리 재편)·validator.go 매처 변경
