---
id: SPEC-V3R6-DOCS-PREREWRITE-FLAG-001
title: "Docs Pre-Rewrite Flag Citations — --factory-run / --clear-policy citation cleanup (card t1726)"
version: "0.1.2"
status: draft
created: 2026-10-19
updated: 2026-10-19
author: Goos Kim
priority: P1
phase: "v3.2.0"
module: "docs-site/content"
lifecycle: spec-anchored
tags: "docs, docs-site, i18n, factory-mode, cli-reference, readme, flag-citations, v3.2"
tier: M
related_specs: [SPEC-V3R6-DOCS-PREREWRITE-CLEANUP-001]
---

# SPEC-V3R6-DOCS-PREREWRITE-FLAG-001 — 플래그 인용 정리 (`--factory-run`·`--clear-policy`)

## HISTORY

| Date | Version | Author | Change |
|------|---------|--------|--------|
| 2026-10-19 | 0.1.0 | manager-spec | 플랜 페이즈 최초 작성 (card t1726, Class C, docs-only). 모든 계수와 RED-now 셀은 이 트리 74ad197380에서 이 실행에 직접 측정했다(레인 전제 기록 `.moai/reports/t1726/progress.md`와 일치 재확인). 운영자 결정 기록 d-20261019T015908Z-220a(2026-10-19)에 따라 Q13(치환/삭제 판별 규칙)과 Q17(4 로캘 한 번에)을 decision-index에 적용본으로 수록했다. Q18은 이 결정으로 정해지지 않은 별개 열린 행이다(기본값 (a)는 planner 잠정값). REQ·AC는 모두 신규 id(부모 AC-PRC-001 범위 밖, verdict D3)를 쓴다. |
| 2026-10-19 | 0.1.1 | manager-spec | 플랜 감사 1차(iter-1) 차반 결함 수리 — verdict FAIL 0.94, receipt rcpt-25df35500d7f88da2fabbf8d. D1: plan.md F.1 M1에 Q18 운영자 판정(리더 경유)을 런 페이즈 진입 선결 조건으로 명시하고 planner 기본값 (a)를 잠정 계획 입력으로 한정(자율 진입 승인 아님). D2: 원문 대조 판정 — `factoryPrintClearPolicyLine`("No value instructs a session clear")과 `TestSD_AC020_ClearPolicies`가 레인 `/clear` 안내가 대체됐음을 확인하여 factory-mode 77행 clear-each 셀·134행, README 87행을 STALE-3으로 재분류하고 REQ-DPF-005의 보존 범위를 살아 있는 선택자·값 인용으로 축소(launchers 37행 값 나열은 LIVE 유지 — AC020은 complete 출력의 clear 지시를 금할 뿐 문서의 값 나열을 금하지 않음). D3: AC-DPF-006을 커밋+스테이지+작업트리+언트래킹 다중 상태 열거로 강화하고 재핀 규칙 명시. D4: E-02를 실측 원문 stdout(ugrep 7.8.4, 트리 13cfafa454)으로 재기록. D5: "세 로캘" 계수 문구 정정. D6: 미존재 GATE-001 참조에 계획 전 SPEC 표기. REQ·AC 식별자 불변. |
| 2026-10-19 | 0.1.2 | manager-spec | 플랜 감사 2차(iter-2) 차반 결함 수리 — verdict FAIL 0.88(점수 회귀 STOP; D1~D6은 재확인 해소, 재오픈 안 함), receipt rcpt-454d40d50ba7205d6fdcec4d. D7: AC-DPF-002 green path의 keep-절을 범위표·plan.md M2와 일치하게 수정(`--clear-policy <값>` 행 자체와 `clear-when-full` 셀 서술은 유지, `relaunch` 셀과 `clear-each` 셀의 `/clear` 요청 서술은 치환 — `clear-each`는 세션 안 카드 전환 서술로)하고, STALE-3 부재를 검증하는 release-blocking AC-DPF-007을 RED-now 셀 E-05(현재 HEAD b484ecacce에서 12줄 실측, exit 0)와 짝으로 신설하며 §A.1 매트릭스 행을 추가했다. D8(선택): AC-DPF-006 허용 목록 대조의 합집합에 경로스펙 없는 `git diff --name-only <base>..HEAD`를 추가. plan.md F.2 M2·F.3 M3·E1·E2·E5가 AC-DPF-007을 참조한다. REQ-DPF-001~006 식별자 불변, 기존 AC-DPF-001~006 재번호 없음, Q18 행 무변경. |

## Overview (WHY)

문서가 두 플래그를 인용하는 방식이 코드의 실제 상태와 어긋난 지점을 정리한다. 두 플래그 모두 코드에는 살아 있다. `--factory-run`은 런 선택자(run id로 합류·이어받기)로 파서와 재개 검증에서 살아 있고(`internal/cli/factory.go:126`, `:153-163`, `internal/cli/lane_resume.go:61`), `--clear-policy`는 REQ-SD-020의 레인 전용 선택 값으로 살아 있다(`internal/cli/factory.go:75`, `:127`, `:836`). 그러나 `--clear-policy`의 relaunch 값이 가리키는 **소비 경로**는 카드 t1554로 대체됐다. `internal/cli/factory_card.go:2484`의 `factoryRelaunchSupersededNote`가 그 상태를 기록한다: "카드 소비는 통합 `moai todo --auto` 엔진으로 옮겨졌고, 런처는 레인 세션 하나를 띄운다".

따라서 이 카드의 판단 단위는 플래그가 아니라 **인용(citation)** 이다. 플래그의 살아 있는 동작(런 선택자, 레인 정책 선택과 값 나열)을 서술하는 인용은 남고, 대체된 동작을 현재 동작인 양 서술하는 인용은 정리 대상이다. 대체된 동작은 둘이다: 카드별 relaunch 소비 순환(t1554로 대체)과 레인 `/clear` 안내(SPEC-FACTORY-LANE-NO-CLEAR-001 REQ-LNC-007/009로 대체 — `factoryPrintClearPolicyLine`은 어떤 정책 값도 세션 clear를 지시하지 않고 카드 전환을 세션 안 전환으로 출력하며, `TestSD_AC020_ClearPolicies`가 그 출력에 clear 지시가 없음을 단언한다). 이 분류는 decision-index의 열린 행 Q18(제품 수준)의 판단 입력이며, 이 SPEC의 계획은 그 기본값 (a)(대체된 인용만 정리, 살아 있는 인용 보존)을 적용해 작성됐다. 기본값 (a)는 잠정 계획 입력일 뿐이며 런 페이즈 진입은 Q18 운영자 판정(리더 경유)을 기다린다.

## Scope (WHAT)

측정된 표면은 정확히 12개다(4개 문서 페이지 × 4 로캘 + README 4종). 파일별 인용 줄과 분류:

| 파일 (×해당 로캘) | 줄 | `--factory-run` | `--clear-policy` | 분류 |
|---|---|---|---|---|
| `docs-site/content/{ko,en,ja,zh}/advanced/factory-mode.md` | 65, 67 | 각 2줄 | — | LIVE — 선택자·합류/이어받기 서술, 유지 |
| 같은 파일 | 77 | — | 각 1줄 | **STALE-2 + STALE-3** — `relaunch` 셀이 대체된 카드별 새 세션 기동을 현재 의미로 서술(STALE-2), `clear-each` 셀이 대체된 레인 `/clear` 요청을 현재 의미로 서술(STALE-3) → 두 셀 모두 치환 대상. 행 자체(`--clear-policy <값>` 나열)와 `clear-when-full` 셀(임계치 조건 서술)은 LIVE 유지 |
| 같은 파일 | 134 | — | 각 1줄 | **STALE-3** — "카드가 `done`에 닿으면 그 레인을 `/clear`한다"는 대체된 레인 `/clear` 안내 → 치환 대상(치환 문장은 현재 카드 전환 서술로 플래그 인용을 유지) |
| `docs-site/content/{ko,en,ja,zh}/cli-reference/launchers.md` | 36 | 각 1줄 | — | LIVE, 유지 |
| 같은 파일 | 37 | — | 각 1줄 | LIVE — 값 나열(clear-each/clear-when-full/relaunch), 소비 서술 없음, 유지 |
| `README.md`, `README.ko.md`, `README.ja.md`, `README.zh.md` | 72, 99 | 각 2줄 | — | LIVE, 유지 |
| 같은 파일 | 74 | — | 각 1줄 | **STALE-1** — "카드를 순회하는 `--clear-policy relaunch` 루프"를 현재 대비책으로 서술 → 치환 대상 |
| 같은 파일 | 87 | — | 각 1줄 | **STALE-3** — "카드가 끝나면 레인이 `/clear`를 요청한다"는 대체된 레인 `/clear` 서술 → 치환 대상(치환 문장은 현재 카드 전환 서술로 플래그 인용을 유지) |

분류 원칙(명시): **LIVE** = `--factory-run` 선택자 또는 `--clear-policy`의 살아 있는 측면(REQ-SD-020 레인 선택, 값 나열, `clear-when-full`의 임계치 조건)을 서술하는 인용. **STALE** = 대체된 동작을 현재 동작인 양 서술하는 인용으로 세 부류이다. **STALE-1/2** = t1554로 대체된 카드별 relaunch 소비(런처가 카드마다 새 세션을 띄우는 순환) 서술. **STALE-3** = SPEC-FACTORY-LANE-NO-CLEAR-001 REQ-LNC-007/009로 대체된 레인 `/clear` 안내 서술. STALE-3 판정 근거(원문 대조, iter-1 D2): `internal/cli/factory_card.go`의 `factoryPrintClearPolicyLine`은 clear-each(기본값 포함)에 카드 전환을 세션 안 전환(`ExitWorktree(keep)` → `moai factory next` → `EnterWorktree`)으로 출력하고 relaunch에 세션 종료 요청을 출력하며, godoc은 "No value instructs a session clear"로 끝난다. `TestSD_AC020_ClearPolicies`(`internal/cli/factory_m6_test.go`)는 complete 출력이 `/clear` 지시를 담지 않음을 단언한다. launchers 37행의 값 나열은 LIVE로 유지한다 — AC020 검사는 complete 출력의 clear 지시를 금할 뿐 플래그의 존재나 문서의 값 나열을 금하지 않는다. STALE 치환 문장은 살아 있는 플래그(`--clear-policy`)를 계속 이름 붙이므로 문서군별 플래그 카운트는 치환 전후 동일하게 유지된다. 계측 근거: `grep -c` 라인 카운트와 인용 원문은 acceptance.md 증거 장부(E-01~E-04)에 있다.

## Requirements (GEARS)

**REQ-DPF-001** (Ubiquitous): 네 README 파일(`README.md`, `README.ko.md`, `README.ja.md`, `README.zh.md`)은 카드 t1554로 소비 경로가 대체된 `--clear-policy relaunch` 카드별 순환을 현재 동작인 것처럼 인용하는 문장을 담지 **shall not** 한다(STALE-1).

**REQ-DPF-002** (Ubiquitous): `docs-site/content/{ko,en,ja,zh}/advanced/factory-mode.md`의 `--clear-policy` 선택 행은 `relaunch` 값의 소비를 "감독 런처가 카드마다 다음 카드용 새 세션을 띄운다"는 대체된 의미로 서술 **shall not** 한다(STALE-2).

**REQ-DPF-003** (Event-driven, Q13 적용): **When** 이 SPEC이 STALE 분류(STALE-1·2·3) 인용을 정리할 때, 문서는 교체가 될 현재 문서·명령이 존재하면 그 인용을 현재 메커니즘(STALE-1·2는 통합 `moai todo --auto` 엔진의 단일 레인 세션 소비, STALE-3는 세션 안 카드 전환)을 이름 붙인 서술로 **shall** 치환하고, 존재하지 않으면 해당 인용 행을 **shall** 삭제한다. 세 STALE 분류 모두 현재 교체 대상이 존재하므로 기본 처분은 치환이다.

**REQ-DPF-004** (Event-driven, Q17 적용): **When** 이 SPEC이 12개 표면 중 한 로캘의 인용을 편집할 때, 같은 문서의 네 로캘(ko·en·ja·zh) 중 편집한 로캘을 제외한 나머지 로캘에 동일 편집을 같은 변경 세트 안에서 **shall** 적용한다. 로캘별 분산 편집은 하지 않으며, 런 페이즈 결과물은 하나의 커밋에 담아도 된다.

**REQ-DPF-005** (Ubiquitous, 보존 — iter-1 D2로 범위 축소): 12개 표면은 `--factory-run` 선택자 인용(런 id 합류·이어받기)과 `--clear-policy`의 살아 있는 측면 인용 — REQ-SD-020 레인 선택, 값 나열(launchers 37행), `clear-when-full`의 임계치 조건 서술 — 을 **shall** 유지한다. 이 보존은 대체된 서술을 덮지 않는다: 레인 `/clear` 안내(STALE-3)와 카드별 relaunch 소비 서술(STALE-1·2)은 이 REQUIREMENT의 보존 대상이 아니며 REQ-DPF-003의 치환 대상이다. 이 REQUIREMENT는 STALE 정리가 살아 있는 인용을 덧불어 지우는 변형을 막는 가드다.

**REQ-DPF-006** (Unwanted): 이 SPEC의 구현은 `internal/`과 `cmd/` 아래 어떤 파일도 수정 **shall not** 한다. help-text에 플래그를 추가하는 선택지를 포함한다(Out of Scope, 별도 SPEC 필요).

## Decision Inputs

- **Q18**(제품 수준, 열림): decision-index.md의 완전한 행이 소유한다. planner 기본값 (a)는 잠정 계획 입력일 뿐 운영자 판정이 아니며 자율 진입 승인도 아니고, 행의 `Operator verdict:`는 비어 있다 — 런 페이즈 진입(킥오프)은 리더 경유 운영자 판정을 기다린다. GATE 자식 SPEC(card t1717, 계획 전 — `.moai/specs/`에 아직 없다)이 이 답을 입력으로 받는다.
- **Q13·Q17**(운영자 결정 기록 d-20261019T015908Z-220a, 2026-10-19): 각각 치환/삭제 판별 규칙과 4 로캘 동시 변경. decision-index.md에 이 SPEC 적용본 행으로 수록됐고 REQ-DPF-003·REQ-DPF-004에 반영됐다.

## Out of Scope

이 SPEC이 하지 않는 것을 명시해 나중에 조용히 청구되지 않게 한다.

### Out of Scope — internal/cmd 변경과 help-text 추가

- `internal/`, `cmd/` 아래 어떤 파일의 변경도 하지 않는다. 여기에는 `--factory-run`·`--clear-policy`를 `factory`, `goal`, `codex`, `cc`의 help 텍스트에 추가하는 일이 포함된다(Q18의 선택지 (c)). 이 선택지는 코드 변경을 수반하므로 이 문서 전용 SPEC의 범위 밖이며, 선택되면 별도 SPEC이 필요하다(부모 인덱스 Appendix A FLAG 행의 전제와 동일).
- relaunch 소비를 다시 카드별 기동으로 되돌리는 코드 변경. 대체는 t1554가 했고, 이 SPEC은 문서만 정리한다.

### Out of Scope — 대체된 relaunch 소비의 재문서화

- STALE-1·2·3을 치환한 자리에 대체된 카드별 순환이나 레인 `/clear` 안내를 새로운 문구로 되살려 적는 일.
- `moai todo --auto` 통합 엔진의 동작을 새로 상세 문서화하는 일(그 내용은 t1554 계열과 factory-mode 문서의 기존 절이 담당한다). 치환 문장은 현재 메커니즘을 이름 붙이는 최소 서술에 머문다.

### Out of Scope — 이 카드 밖의 문서 정리

- 부모 인덱스의 다른 정리 항목: 부재 명령 인용(Q13의 나머지 적용 주체인 SWEEPA~SWEEPF), 배지, CG 문단, 이미지, 리다이렉트, CI 잡. FLAG는 플래그 인용 줄만 담당한다.
- 12개 표면 안에서도 플래그 인용과 무관한 줄의 편집. `factory-mode.md`는 SWEEPA와, README 4종은 README 자식과 파일을 공유하지만 담당 줄이 다르다(verdict D3 — 부모 AC-PRC-001에서 플래그 항목이 빠져 있는 근거).
- ko·en·ja·zh 이외 로캘로의 번역 확장.

## Dependencies and Parallel Work

- 부모 인덱스: SPEC-V3R6-DOCS-PREREWRITE-CLEANUP-001 (revision 0.3.x). 이 SPEC은 그 11개 자식 중 FLAG다.
- 파일 공유: SWEEPA(card t1721)와 `factory-mode.md` ×4, README 자식(card t1720)과 README 4종 — 다른 줄. 착지 순서 충돌을 줄이려면 부모 Appendix A의 권장(README 먼저, FLAG는 그 뒤)을 따른다.
- GATE 자식(card t1717, 계획 전 — SPEC-V3R6-DOCS-PREREWRITE-GATE-001은 `.moai/specs/`에 아직 없고 그 플랜 페이즈가 이 참조를 소유한다)은 Q18 답을 입력으로 받는다. Q18이 열려 있어도 이 SPEC의 기본값 (a) 적용 분량(STALE-1·2·3 정리)은 GATE의 인용 검사기와 무충돌이다.
- 전제 계측 기록: `.moai/reports/t1726/progress.md`(레인, 트리 74ad197380).
