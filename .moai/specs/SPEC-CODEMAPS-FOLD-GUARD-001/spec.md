---
id: SPEC-CODEMAPS-FOLD-GUARD-001
title: "codemaps fold 판정 단위 산문 보존 가드 — 재생성이 접힌 단위를 다시 쓰지 못하게 한다"
version: "0.1.1"
status: completed
created: 2026-09-14
updated: 2026-09-14
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: "internal/graph, .moai/project/codemaps"
lifecycle: spec-anchored
tags: "codemaps, fold, guard, regression, record-only, t475, t688, t748"
era: V3R6
tier: S
depends_on: [SPEC-CODEMAPS-REFRESH-002, SPEC-GRAPH-STAMP-ANCESTRY-001]
related_specs: [SPEC-CODEMAPS-ACCURACY-001, SPEC-CODEMAPS-REFRESH-001]
---

# SPEC-CODEMAPS-FOLD-GUARD-001 — codemaps fold 판정 단위 산문 보존 가드

## HISTORY

| Version | Date | Change | Author |
|---------|------|--------|--------|
| 0.1.0 | 2026-09-14 | 최초 plan-phase 저작(카드 t748). 기준선은 워크트리 `.claude/worktrees/t748` @ `146faed9d`에서 전수 재측정: 생성기 5문서에 대한 fold 단위 토큰 히트 0(§A.2), `fold-judgments.txt` fold 행 5(§A.2), 가드 테스트 부재(§D AC-CFG-001 RED-now). Tier S 판정 근거는 §B.1. | manager-spec |
| 0.1.1 | 2026-09-14 | plan-audit iter1 PASS 0.86 수리 — D1~D6 전부 적용. D1: REQ-CFG-003에 생성기 5문서 존재·가독성 트리거 신설(읽기 오류 통과 = 공허 녹색 차단, fail-forbidden-root의 문서 수준 확장). D2: AC-CFG-001/003/004 헤더에 소관 REQ 토큰 부기. D3: Event-detected → Event-driven 표기 정규화(구조는 기존 GEARS 준수). D4: plan M1.4에 단일 판독 경로 핀(실트리 스캔과 /tmp 주입 스캔이 같은 판독 함수). D5: REQ-CFG-002에 탐지 한계 명시(토큰을 대지 않는 산문 재기술은 미탐지 — 잔여 수용 근거 포함). D6: §D 미관측 집합을 AC-CFG-002/003/004/005로 정정. 기존 REQ/AC id는 재번호 없이 유지. | manager-spec |

## §A. Problem Statement

### §A.1 사건 — 접힘이 허가로 오용된 한 번

SPEC-CODEMAPS-REFRESH-002의 run 중, 첫 재생성 통과(commit `e397ec00d`, 그 SPEC의 M2)가 **5개 fold 판정 단위 전부에 산문을 새로 썼다.** §A.3(a1) 판별식의 fold 처분 — *"판정과 그 근거를 기록하고, 그 단위에 대해 codemaps 산문을 바꾸지 않는다"*(`.moai/specs/SPEC-CODEMAPS-REFRESH-002/spec.md:163`) — 를 실행자는 판정이 아니라 **편입 허가**로 읽었다. M3(commit `cd03be0d3`)이 fold 단위 산문을 되돌리고 omission 단위 산문은 유지했으며, 되돌린 내용의 전수 목록과 단위별 처분은 `.moai/reports/t475/codemaps-accuracy-verification.md` §④-b(:550-556)에 기록돼 있다. t475의 sync-audit은 같은 리포트에 증거 파일 표시 부채(§④/§④-b)를 남겼다.

되돌림은 완료됐다. 이 카드의 소관은 산문을 다시 만지는 것이 아니라, **같은 재발이 기계적으로 불가능하거나 적색으로 크게 실패하도록 만드는** 것이다.

### §A.2 기준선 — 이 워크트리에서 재측정한 값

모든 수치는 워크트리 `.claude/worktrees/t748`, HEAD `146faed9d`에서 실행한 명령의 출력이다.

```
$ ls internal/graph/codemaps_fold_guard_test.go
ls: internal/graph/codemaps_fold_guard_test.go: No such file or directory     EXIT=1
```

```
$ grep -rn -F -e 'doctor_hook_delivery' -e 'step_git_env' -e 'prlink_landedref' \
    -e 'fieldsets_codex_templ' -e 'internal/core/git' \
    .moai/project/codemaps/{overview,modules,dependencies,entry-points,data-flow}.md
(무출력)                                                                       EXIT=1
```

```
$ grep -c '^fold ' .moai/project/codemaps/fold-judgments.txt
5                                                                             EXIT=0
```

- **생성기 5문서**(`overview.md`, `modules.md`, `dependencies.md`, `entry-points.md`, `data-flow.md` — t688 §B.1 G3가 명시한 집합)에는 fold 단위의 전체 경로와 파일명 토큰이 **0건**이다.
- 유일한 언급면은 기록 계층 `.moai/project/codemaps/fold-judgments.txt`다 — `fold <unit>` 5행 + `omission <unit>` 15행. t688(G3)이 이 파일이 생성기 5문서가 아님을 이미 명시했다.
- **단축형 `core/git`은 산문에 남는다**(예: `overview.md:103` "``core/git``은 인프라", `modules.md:171` "``core/git`` 위에 얹은 상위 유틸리티"). 이것이 바로 `internal/core/git`의 fold 근거다(t475 §⑥ 표) — **단축형은 위반이 아니라 접힌 결과물이다.** 위반 형태는 사건에서 되돌려진 것과 같은 **전체 경로·파일명 토큰**의 (재)등장이다.

### §A.3 세 축 — 층 구분을 유지한다 (발주 문구의 해석)

발주 문구 *"AC 접미 fold, reqLineWidePattern 캡처 실패, 검증 축은 비침묵"* 을 t475 산출물과 저장소 코드로 검증한 결과, 세 층은 서로 다른 축이며 이 SPEC은 셋을 접지 않는다.

1. **AC 접미 fold (기록 계층).** fold 판정은 AC/기록 계층에 산다 — AC-CM2-007의 기록 전용 분류는 **fold 판정 후보에만** 적용되고(`.moai/specs/SPEC-CODEMAPS-REFRESH-002/acceptance.md:163-165` "기록 전용 처분은 fold에만 적용된다"), 판정 본체는 t475 §⑥ 표(:102-135)와 `fold-judgments.txt`다. 사건은 정확히 이 축의 침범이었다: 기록 계층의 fold 결정이 산문 계층의 편입 허가로 오용됐다. 이 가드의 제1 보호 대상은 이 층이다.
2. **reqLineWidePattern 캡처 실패 (lint 계층 — 별개 축).** `internal/spec/lint_req_widen.go:8-59`는 REQ 정의줄 패턴이 담지 못하는 여섯 줄 형태를 문서화한 **기지의 lint 축**이다(`:59` 패턴 정의, 마크다운 리스트 불릿 앵커 + 캡처 그룹 누락 형태들). 발주 문구의 지시로 읽는 바: **히트-0 상태를 lint 쪽 캡처 실패가 원인일 때 산문을 써서 "고쳐서는 안 된다."** 측정이 빠진 것이지 서술이 빠진 것이 아니기 때문이다. 따라서 이 가드의 판정 근거는 REQ/lint 계층의 줄-형태 패턴과 무관한 **exact-substring 히트 규약**(§A.3(a)의 `grep -c -F` 관습 계승)으로 정의한다 — 캡처 축의 사각지대를 판정 근거로 수입하지 않는다. lint 축 자체의 수리는 이 카드 소관이 아니다(§ Out of Scope).
3. **검증 축은 비침묵 (판정 계층).** 회귀는 조용히 로그를 남기고 통과해서는 안 되며, 반드시 FAIL 해야 한다. 근거: `verification-completeness.md` §1.1(report-not-verdict — 결과를 인쇄하고 마지막 성공의 종료 코드를 돌려주는 탐침)과 t747 sync-audit F5 교훈(코퍼스 탐침이 통제 델타를 기록만 하고 단언하지 않았다). 이 가드는 Go 테스트로서 실패 시 `--- FAIL` + 비정상 종료 코드로 CI에 도달한다(§1.2(c) 도달성).

발주 문구의 해석이 산출물과 어긋나는 지점은 발견되지 않았다. 위 셋이 이 SPEC의 설계를 직접 결정한다: (1)→보호 집합의 원천은 기록 계층이고, (2)→판정 규약은 substring 히트이며 lint 캡처에 의존하지 않고, (3)→운반체는 종료 코드로 실패하는 자동 실행 검사다.

### §A.4 t688 위에서의 합성 — 스탬프 독립

t688(SPEC-GRAPH-STAMP-ANCESTRY-001, `completed`)이 신선도 기계에 (M1) 스탬프 조상성 선판정, (M2) push 가드, (M4) codemaps 재생성 + 재스탬프를 얹었다. 이 가드와의 관계:

- **직교 축이다.** `moai graph check`의 codemaps 계층은 described-source-diff(소스↔스탬프 드리프트)를 잰다. 사건의 결함 — fold 단위에 산문이 **기록됨** — 은 그 지표를 움직이지 않았고, 재스탬프만으로도 fresh가 된다. 즉 신선도 게이트는 이 결함에 **구조적으로 눈멀다**. 가드는 산문 내용(정확히는 fold 단위 토큰의 부재)을 재는 별개 축이다.
- **재스탬프는 가드 기대를 무효화하지 않는다.** t688 M4의 재앵커링(`provenance.json` 갱신)은 생성기 5문서의 산문을 바꾸지 않으므로, fold 단위 토큰 부재라는 가드의 기대는 스탬프 값과 무관하게 불변이다. REQ-CFG-004가 이를 요구로 고정한다.
- **합성은 중복이 아니다.** 가드는 `moai graph check`에 새 계층을 더하지 않는다(체커·CLI·워크플로 변경 0 — § Out of Scope). 두 검사는 같은 검증 배치에서 나란히 돌 수 있고, 한쪽 적색이 다른 쪽 수리로 풀리지 않는다 — 신선도 적색의 수리는 재생성+도달 가능한 재스탬프, fold 가드 적색의 수리는 산문 되돌림이다.

## §B. Scope

### §B.1 In Scope — Tier S 판정

- **M1**: `internal/graph`에 fold 가드 테스트 1개 신설 — 기록 계층에서 보호 집합을 읽고(데이터 주도), 5단위 floor 핀, 생성기 5문서의 fold 토큰 부재 단언, 기록 지속성 단언, 스탬프 무관성.
- **M2**: RED 관측 가능성 입증 — 주입 가능한 문서 디렉터리로 변조 입력 → FAIL, 복원 → PASS(동결 표면 `.moai/project/codemaps/**` 불변).
- **M3**: CI 도달성 확인 + 종결 증거.

Tier 판정: **S**. run-phase 변경은 테스트 파일 1개(예상 < 300 LOC), 프로덕션 코드 0줄, 영향 파일 < 5. 검증 단계 통합이 별도 아티팩트(acceptance.md)를 요구하지 않는다 — AC는 spec.md §3에 인라인으로 충분하고, Tier S 상한(REQ 8 / AC 8) 안에 들어온다. 검증 절차 통합이 워크플로·CLI 변경을 요구하지 않는다는 것이 §A.4에서 확정됐으므로 M으로 올릴 근거가 없다.

### Out of Scope — 접힘 정책과 판정의 재오픈

- fold↔omission 판정을 다시 내지 않는다. 판정 본체는 t475 §⑥(책임 질문 + 부모 산문 인용)이며, 이 카드는 그 결과를 **보존**할 뿐 재심하지 않는다.
- `fold-judgments.txt`나 5문서의 산문 내용을 편집하지 않는다. 사건의 되돌림은 완료됐다(§A.1) — 이 카드는 산문을 다시 쓰지 않는다.
- 미래의 정당한 재판정(fold→omission 전환 등)을 금지하지 않는다. 다만 그때 가드는 적색으로 실패하며, 재판정이 기록 계층에서 명시적으로 갱신되어야 한다는 신호가 그 실패다 — 조용한 재분류가 불가능해지는 것이 설계 의도다.

### Out of Scope — graph 체커·신선도 기계 변경

- `moai graph check`에 fold 계층을 추가하지 않는다. `internal/graph/check.go`·`check_citations.go`의 프로덕션 변경 0줄, 임계값·`gate.yaml` 변경 0, `.github/workflows/graph-freshness.yml` 변경 0. (t688이 방금 체커 계약을 안정시켰다 — REQ-GSA-012.)
- 신선도 임계값 40 논쟁을 다시 열지 않는다(SPEC-GRAPH-FRESHNESS-CADENCE-001 유지 판정 계승).

### Out of Scope — lint 캡처 축의 수리

- `reqLineWidePattern`(`internal/spec/lint_req_widen.go`)과 그 잔여 미캡처 형태의 수리는 별개 축이다. 이 카드는 그 축을 판정 근거로 수입하지 않을 뿐 수리하지도 않는다.

### Out of Scope — `docs-truth.md`와 문서 내용 저작

- `docs-truth.md`는 손으로 유지되는 생성기 밖 문서다(REQ-CM2-014 계승). 이 카드는 그 내용에 관여하지 않는다.

## §C. Requirements (GEARS)

- **REQ-CFG-001** (Ubiquitous) — **보호 집합의 원천은 기록 계층이다.** the fold guard shall derive the protected unit set from `.moai/project/codemaps/fold-judgments.txt`의 `fold <unit>` 행(데이터 주도)이며, t475 §⑥이 fold로 판정한 5단위 — `internal/core/git`, `internal/cli/doctor_hook_delivery.go`, `internal/hook/quality/step_git_env.go`, `internal/kanban/prlink_landedref.go`, `internal/web/fieldsets_codex_templ.go` — 를 **floor 집합으로 핀한다.** 손 열거만으로 집합을 만들지 않는다(REFRESH-002 plan-audit D1의 교훈 — 손 열거는 자기 결함을 재생산한다): 일반 규칙은 기록을 읽고, floor 핀은 기록이 조용히 파이는 것을 막는 이중 잠금이다.

- **REQ-CFG-002** (Event-driven) — **fold 단위 토큰의 산문 (재)등장은 실패한다.** **When** 생성기 5문서(`overview.md`, `modules.md`, `dependencies.md`, `entry-points.md`, `data-flow.md`) 중 하나라도 보호 단위의 토큰을 담는 것이 감지되면, the guard shall 비정상 종료 코드로 실패하고 어느 단위가 어느 문서의 몇 행에서 발견됐는지 이름을 대라. **관측면의 정의(판정 규약)**: "fold 단위 산문 존재"란, 문서의 한 행이 단위의 저장소 경로 또는 파일명 토큰(`fold-judgments.txt`에 적힌 그대로)을 exact-substring으로 담는 것이다 — §A.3(a) 히트 규약(`grep -c -F`)의 계승이다. **단축형 예외는 결함이 아니다**: `core/git`(전체 경로 아님)은 부모 산문의 접힌 결과물이며 — `overview.md:103`, `modules.md:171` 실측 — 위반 토큰이 아니다. 생성 관계 서술(예: "`.templ` → `_templ.go` 생성")은 산물 파일명을 대지 않는 한 허용된다(t475 §⑥ `fieldsets_codex_templ.go` fold 행의 처분과 일치).

  **탐지 한계(명시적 잔여)**: 이 규약은 토큰을 대는 재기술만 잡는다 — 단위의 경로·파일명 토큰을 전혀 대지 않고 산문을 다시 쓰는 재생성은 이 가드를 피한다. 이 잔여는 수용한다: 사건의 재생성기 서명은 토큰을 운반했다(t475 §④-b:558 "전체 경로 표기 6곳"), 그리고 토큰 대신 의미 단위의 산문 대조를 요구하는 대안들은 가드를 비기계적 판정으로 되돌리고 비침묵 축(축 3)을 약화시키므로 엄격히 더 나쁘다(§A.4의 직교성 논거).

- **REQ-CFG-003** (Event-driven) — **기록의 지속성 — 히트-0 상태는 판정과 함께 보인다.** **When** `fold-judgments.txt`가 부재하거나, 파싱 불가능하거나, floor 집합의 어느 단위 행이든 잃었음이 감지되면, the guard shall 실패한다. **When** 생성기 5문서 중 하나라도 부재하거나 읽을 수 없음이 감지되면, the guard shall 실패한다 — 읽기 오류를 통과로 처리하는 구현은 공허한 녹색(스윕 집합 0의 녹색, §A.3 축 3)이므로, 문서 수준에서도 plan.md §D의 fail-forbidden-root 규칙(루트 미발견은 skip 아닌 실패)이 확장 적용된다. fold 단위의 히트-0 상태는 **판정(fold 표기)과 함께 보이는 상태**다 — 기록이 사라지면 "접혀서 없다"와 "잊혀서 없다"가 구분되지 않는다(발주 제2 보호 축). 재판정에 의한 정당한 기록 갱신도 이 REQ 아래에서는 적색이며, 갱신이 기록 계층의 명시적 행위로 남도록 한다(§A.1 사건과 같은 조용한 경로 차단).

- **REQ-CFG-004** (Where, capability gate) — **스탬프 독립과 합성.** **Where** the fold guard evaluates, it shall not read `provenance.json` nor condition its verdict on `moai graph check`의 어떤 판정 — 스탬프 재앵커링(t688 M4)과 조상성 선판정(t688 G1)은 이 가드의 판정을 바꾸지 않는다. 그리고 the fold guard shall be implemented as a standalone regression test — `moai graph check`의 계층 집합, CLI 종료 코드 의미론, graph-freshness 워크플로에 어떤 변경도 요구하지 않는다(중복 계층 신설 금지, §A.4).

- **REQ-CFG-005** (Event-driven) — **RED 입증은 동결 표면을 건드리지 않는다.** **When** 가드의 RED 관측 가능성을 입증할 때(변조 입력 → FAIL → 복원 → PASS), the guard shall 주입 가능한 문서 디렉터리 입력을 받아들여 변조 사본 위에서 판정하며, 동결 표면인 실제 `.moai/project/codemaps/**`를 변조해서는 안 된다. live 표면을 변조했다가 되돌리는 입증 방식은 본 REQ의 위반이다 — 표면이 기준선인 저장소에서 그 방식은 사고의 재생산이다.

## §D. Acceptance Criteria (Given-When-Then — Tier S 인라인)

RED-now 셀은 `verification-completeness.md` §2.1의 4요소(명령 + verbatim stdout + exit + 트리 SHA)로 이 워크트리 `146faed9d`에서 실측했다. AC-CFG-002/003/004/005의 RED는 가드 존재가 전제라 저작 시점에 관측 불가하며, run-phase M2가 fixtures 위에서 RED→GREEN을 입증한다 — 그 전까지 이 넷은 **pass로 기록되지 않는다**(§2 미채택 상태).

- **AC-CFG-001 — 가드 존재 + 현재 트리 녹색 (REQ-CFG-001 · REQ-CFG-002 · REQ-CFG-004).** **Given** 되돌림이 완료된 현재 codemaps(§A.2: 토큰 히트 0, floor 5행) **When** 가드 테스트가 기본 스위트에서 실행되면 **Then** PASS다.
  - RED-now: `ls internal/graph/codemaps_fold_guard_test.go` → `ls: ... No such file or directory`, exit `1`, 트리 `146faed9d` — 가드가 아직 없다.
  - green path: M1이 테스트를 신설하고 `go test ./internal/graph/ -run 'Fold'`가 `ok` + exit `0`.

- **AC-CFG-002 — 변조 → FAIL, 복원 → PASS (RED 관측).** **Given** REQ-CFG-005의 주입 가능한 입력으로 준비한 생성기 문서 사본에 fold 단위 토큰 1행(예: `modules.md` 사본에 `internal/kanban/prlink_landedref.go` 행)을 주입했다 **When** 가드를 실행하면 **Then** FAIL이고 메시지가 단위·문서를 이름 대며, 사본을 복원하면 PASS다.
  - RED-now: 관측 불가(가드 부재) — run-phase M2가 입증. **이 AC는 그 증거가 수출되기 전까지 pass로 기록하지 않는다.**

- **AC-CFG-003 — 기록 floor (REQ-CFG-001 · REQ-CFG-003).** **Given** `fold-judgments.txt`의 fold 행 5(§A.2 실측: `grep -c '^fold '` → `5`, exit `0`, 트리 `146faed9d`) **When** floor 단위 행 하나를 잃은 기록 사본으로 가드를 돌리면 **Then** FAIL이고, 온전한 기록이면 PASS다.
  - RED-now: floor 자체는 현재 트리에 존재(위 실측) — 이 AC의 RED는 M2가 사본 변조로 입증한다.

- **AC-CFG-004 — 스탬프 독립 (REQ-CFG-004).** **Given** `provenance.json`의 스탬프가 재앵커링된 상태(또는 판독 불가 상태) **When** 가드를 실행하면 **Then** 판정이 스탬프 상태와 무관하게 동일하다. 그리고 `moai graph check`의 계층 출력에 fold 관련 신규 행이 없다(변경 0).
  - RED-now: 관측 불가(가드 부재) — M2 입증. 체커 무변경은 M3의 `git diff` 범위 검사로 판정.

- **AC-CFG-005 — 비침묵 도달성.** **Given** 가드가 `internal/graph` 패키지 테스트로 존재 **When** `go test ./internal/graph/...`가 기본 스위트(CI 포함)에서 실행되면 **Then** 가드 위반은 `--- FAIL` + 비정상 종료 코드로 표면에 오르고, 보호 집합이 빈 스윕으로 녹색을 낼 경로가 없다(floor 단언이 스윕보다 선행 — REQ-CFG-001의 이중 잠금).
  - RED-now: `ls internal/graph/codemaps_fold_guard_test.go` → 부재, exit `1`, 트리 `146faed9d`(AC-CFG-001과 동일 관측).
  - green path: M1 신설 후 기본 스위트에 편입됨(별도 CI 워크플로 편집 없이 — 기존 `go test` 스위트가 운반한다).

## §E. Cross-References

- **SPEC-CODEMAPS-REFRESH-002** (`completed`) — §A.3(a1) 판별식과 AC-CM2-007의 기록 전용 규정(:163-165), [HARD] :236의 omission 누출 차단. 이 가드는 그 계약의 **사후 회귀 잠금**이다.
- **SPEC-GRAPH-STAMP-ANCESTRY-001** (`completed`, t688) — 조상성 선판정·push 가드·생성기 5문서 명시(§B.1 G3)와 `fold-judgments.txt` 비생성 문서 선언. 이 가드의 합성 상대(§A.4).
- **`.moai/reports/t475/codemaps-accuracy-verification.md`** — §⑥ 판정 표(:102-135, 요약 :144-149: fold 5 / omission 15), §④-b 되돌림 전수 목록(:550-556), 기준 사본 `.moai/reports/t475/pre-regen/`. **동결 표면 — 이 카드가 수정하지 않는다.**
- **SPEC-CODEMAPS-ACCURACY-001 / SPEC-CODEMAPS-REFRESH-001** — 정확성 검증 3항목과 절차 정본. 이 카드는 그 어느 아티팩트도 재개하지 않는다.
- **verification-completeness.md** §1.1/§1.2/§2 — 비침묵 판정(§1.2(c) 도달성), 공허 스윕 차단, 2셀 채용 규율. AC 설계의 직접 근거.
- 카드 t748 — 발주 카드. 사건 개요와 세 축 문구의 출처.
