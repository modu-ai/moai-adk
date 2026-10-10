---
id: SPEC-REVIEW-GATE-REPAIR-001
title: "Codex 리뷰 게이트 수리 — 프롬프트 진입 교착을 풀고 리뷰 기준을 하나로 모은다"
version: "0.1.1"
status: draft
created: 2026-10-10
updated: 2026-10-10
author: manager-spec
priority: P0
phase: "v3.2.0 target"
module: "internal/cli"
lifecycle: spec-anchored
tier: L
tags: "codex, review-gate, turn-end, prompt-entry, receipt, merge-base, card-t1620"
---

# SPEC-REVIEW-GATE-REPAIR-001 — Codex 리뷰 게이트의 프롬프트 진입 교착을 풀고 리뷰 기준을 하나로 모은다

카드: **t1620** (class C — 훅·영수증·리뷰 범위·설정 네 서브시스템을 가로지르는 설계 변경) · 기준 코드 트리 `2aab5f797` (작업 브랜치 `WT-review-gate-fix`, plan 커밋 `6218869fa` 위)

## HISTORY

- 2026-10-10 · v0.1.0 · manager-spec · 최초 작성(draft). Tier L 판정 근거는 plan.md §A. 측정 원천은 research.md(인용 전수 재검증, 기준선 셀 실행). 운영자 판정이 필요한 항목 Q1–Q6 은 decision-index.md 에 개방 상태로 두었다.
- 2026-10-10 · v0.1.1 · manager-spec · plan-audit 1회차 FAIL(score 0.625, MP-7 실패; 근거 `.moai/reports/t1620/plan-audit.md` §1.4 의 D1–D12) 을 수리했다. 운영자 판정 Q1–Q6 과 보드 판정 D3·D4·D11 을 decision-index.md 에 원문으로 기록했다(보드 기록 d-20261010T053034Z-c50f, d-20261010T053834Z-422b, d-20261010T054047Z-3fc2). 요구사항 16개, 인수 기준 20개(Tier L 상한은 각각 25). 의미를 바꾸지 않는 감사 선택 항목(O1–O11)을 적용했다. **Q6 처분**: SPEC-GATE-BOTTLENECK-001 의 REQ-GBN-002·REQ-GBN-004·AC-GBN-004 는 이 SPEC 의 HISTORY 와 F5 에만 기록하며, 그 SPEC 은 편집하지 않는다. 재감사를 통과하기 전에는 Implementation Kickoff 를 열지 않는다.

---

## §0 지배 원칙 [HARD]

> **턴 종료 게이트만 차단하고, 프롬프트 진입은 차단하지 않는다. 검토가 이뤄지지 않은 상태는 "통과"로 재사용되지 않으며, 검토 없이 턴 종료가 허용될 때마다 화면에 드러난다.**

교착의 구조는 이렇다. 진입 단계에서 막힌 프롬프트는 모델에게 턴을 주지 않으므로, 리뷰 키가 움직일 방법이 없다(research.md §4, 설계 분석). 모델이 조치할 수 있는 지점은 턴 종료뿐이다. 그래서 검토 결과는 진입 시 추가 맥락으로 한 번 전달되고, 차단은 턴 종료에서만 일어나며, 그 차단은 키당 총 C 회로 상한을 가진다(C = 1). 검토되지 않은 채 허용되는 턴 종료는 systemMessage 로 드러난다(REQ-RG-016). 이 교착의 관측 범위와 관측하지 못한 전제는 §A.1 에 적었다.

---

## §A 배경 (측정)

### A.1 관측 사실 — 기계 로컬 fail 스냅샷 한 건 (recorded_at 2026-10-10T01:37:46+09:00, key 2aab5f797…:7f1e2657f0831d9f)

- 관측: 레인 저장소의 `.moai/worktrees/t1543/.moai/state/verify/snapshots/2aab5f797b75-7f1e2657f0831d9f.json`(기계 로컬, 커밋 대상 아님)에서 `codex-review` 검사는 `verdict: fail`, `exit_code: 1`, `recorded_at: 2026-10-10T01:37:46.265725+09:00`, `tool_version: codex-cli 0.161.0` 이다. 키는 `2aab5f797…:7f1e2657f0831d9f` 이며, 앞부분은 전체 SHA `2aab5f797b75983e132af451da68f69e3426557b`(이 트리의 기준 커밋)의 접두사다.
- 상세 파일의 발견 사항 두 건은 `internal/homestate/card_transition.go:287` 과 `internal/cli/factory_card.go:2894` 를 가리키며, 두 경로는 이 트리에 있다. 발견 사항이 옳은지는 이 SPEC 이 판단하지 않는다(research.md §4).
- 미관측(Gap) 세 항목:
  1. 레인 자체가 카드 변경을 하지 않았다는 전제. 이 세션의 git 가드가 다른 워크트리 대상 호출을 거부했다(research.md §8 Gap 1).
  2. 24시간 잠금 기간. 기록 시각에 유효기간 24시간(`internal/cli/codex_review_cache.go:29`)을 더한 값은 산술 결과이며 관측이 아니다(research.md §4).
  3. 진입 훅이 실제로 프롬프트를 막았다는 사실. 재현하지 않았다(research.md §3 #22, §8 Gap 2).

### A.2 리뷰 기준은 두 해석기가 서로 다른 설정 키를 읽는다

- 턴 종료 게이트와 card 범위 리뷰는 `cardMergeBase` 로 기준을 정한다. 이 함수는 `IntegrationTarget` 을 읽고(`internal/cli/codex_review_scope.go:189-191`), 그 값은 git-flow 수동 프로필에서 `develop_branch`(`.moai/config/sections/git-strategy.yaml:16`, 값 `develop`; 키 경로 `git_strategy.manual.develop_branch`)다(`internal/config/loader_integration_branch.go:123`).
- 감사 도구는 `resolveReviewBase` 를 쓰며, 이 함수는 첫 단계로 `worktree_base_branch` 를 읽는다(`internal/cli/mcp_review_material.go:159`, 값 `main`, `git-strategy.yaml:5`). 이 단계는 커밋 `f82a078bb`(2026-10-03, 카드 t1426)가 추가했다.
- 토폴로지(git 측정): `merge-base(main, develop) = db0c514d3`, `db0c514d3..main` 139 커밋, `db0c514d3..develop` 149 커밋.
- 같은 저장소에서 두 해석기는 서로 다른 기준을 고른다. 두 해석을 하나의 함수로 모으는 이유다(REQ-RG-005).

### A.3 완료된 SPEC 의 결정이 코드에서 뒤집혀 있다

- SPEC-CODEX-REVIEW-TARGET-001(status: completed) §A.7 은 codex 네이티브 경로가 `worktree_base_branch` 를 읽지 않는다고 결정했다(`.moai/specs/SPEC-CODEX-REVIEW-TARGET-001/spec.md` 102–111행).
- 카드 t1426 이 이 결정을 뒤집었다. 감사 기준 테스트의 주석이 이를 명시한다: "Card t1426 reverses the spec.md §A.7 non-read clause"(`internal/cli/codex_review_target_test.go`, AC-CRT-002 절).
- 해당 SPEC 의 본문에는 t1426 을 가리키는 항목이 없다(`grep -c t1426` = 0). 결정의 기록은 코드와 테스트뿐이다. 운영자 판정은 Q1 에서 받았다.

### A.4 진입 차단은 진행 중 SPEC 의 명시 요구다, 기록 처분은 Q6 이다

- SPEC-GATE-BOTTLENECK-001(status: in-progress)의 REQ-GBN-002(`spec.md:43–47`): "WHEN no fresh receipt exists at Stop, THE GATE SHALL allow the turn, start the review in the background, and leave the verdict to be enforced at the NEXT turn entry; a fresh FAIL at turn entry SHALL block. `async:true` SHALL NOT be used." 진입 훅의 헤더 주석(`internal/cli/codex_review_delay.go` 1–14행)도 같은 설계를 적었다. REQ-RG-001 은 이 진입 차단 절을 뒤집는다.
- 같은 SPEC 의 REQ-GBN-004(`spec.md:51–54`)는 모든 경로가 fail-open(missing/inconclusive reviewer 는 허용)을 지키도록 요구하고, AC-GBN-004(`spec.md:67–68`)의 inconclusive 절은 실패 시 허용을 단언한다. 이 SPEC 은 REQ-GBN-004 를 뒤집지 않는다. REQ-RG-001 과 REQ-RG-010 은 inconclusive 와 모든 진입 상태에서 허용을 유지한다.
- 기록 처분 Q6 = (b)(보드 기록 d-20261010T054047Z-3fc2): 이 SPEC 의 HISTORY 와 F5 에만 기록하고, 대상 SPEC 은 편집하지 않는다. 대상 SPEC 이 in-progress 인 채로 코드와 어긋나는 잔여 위험은 F5 에 적었다.

### A.5 영수증과 재사용 규칙

- 영수증의 결합 필드는 다섯 개다: head, tree_digest, config_digest, command, tool_version(`internal/verify/receipt.go:70-76`). 유효기간은 24시간이다(`internal/cli/codex_review_cache.go:29`).
- 현재 턴 종료 게이트는 수용된 영수증이 pass 든 inconclusive 든 허용한다(`internal/cli/codex_review_gate.go:176-178`). 관측: `TestReviewGate_InconclusiveReceiptIsNotPassEvidence` 가 통과하며, stderr 에 "reusing the cached inconclusive verdict for the unchanged tree" 를 남긴다(acceptance.md §C, 셀 B-2).

### A.6 설정 상태

- 추적 파일 `.moai/config/sections/workflow.yaml` 302행의 `codex.review_gate.enabled` 는 HEAD 에서 `true` 다(`git show HEAD:` 로 확인). 이 워크트리의 작업 트리는 HEAD 와 같다(`git status --short` 출력 없음, plan 단계 측정). 앞선 측정 세션이 기록한 리더의 미커밋 변경(`false`)은 이 워크트리에서 재현되지 않았다(research.md §5 P-10 정정, Gap).
- 스위치가 없을 때의 기본값은 커밋된 값이다(REQ-RG-011, 보드 판정 D11). 현재 Go 층은 판독 트리의 작업 트리 파일을 읽으므로(`internal/cli/mcp_codex.go:2503-2510`), 기본값을 커밋된 값으로 바꾸는 것은 M6 의 행위 변경이다.

### A.7 카드의 두 주장 정정

- 카드가 적은 `internal/hook/codex_review_*.go` 경로는 이 트리에 없다. 네 파일은 모두 `internal/cli/` 에 있다(research.md §2).
- "두 층을 모두 켜면 doctor 경고가 뜬다"는 주장은 코드에서 관측되지 않았다. `internal/cli/doctor*.go` 71개 파일 전체 검색에 해당 검사가 없고(research.md §5 P-6), 문서 한 문장(`.claude/rules/moai/workflow/factory-dispatch-detail.md:157`)만 있다(Q5).

---

## §B 요구사항 (GEARS)

### REQ-RG-001 — 진입 훅은 프롬프트를 막지 않는다 (Ubiquitous, M2)

The codex review entry hook shall not return a block decision for any receipt state.

어떤 영수증 상태에서도 진입 훅은 차단 결정을 내지 않는다(absent, stale, key-mismatch, fresh-pass, fresh-fail, fresh-inconclusive 모두 포함). 이 요구가 교착을 끊는다. 발견 사항의 전달은 REQ-RG-002 가 맡는다. 이 요구는 SPEC-GATE-BOTTLENECK-001 REQ-GBN-002 의 진입 차단 절을 뒤집는다(기록 처분 Q6 = (b), §A.4).

### REQ-RG-002 — 발견 사항은 키마다 한 번 추가 맥락으로 전달된다 (event-driven, M2)

When a prompt is submitted and a fresh fail receipt exists for the current review key, the entry hook shall add the fail detail to the model context as additional context exactly once for that key, and shall record that delivery against that key.

전달 수단은 UserPromptSubmit 출력의 `additionalContext` 다(`.claude/rules/moai/core/hooks-system.md:91`). 이미 전달된 키는 프롬프트마다 반복하지 않는다. 리뷰 키(영수증 결합 필드 다섯 개의 조합)가 바뀌면 새 전달 기록이 생긴다.

### REQ-RG-003 — 차단은 턴 종료에서만, 키당 C 회까지 한다 (event-driven, M3)

When a turn ends with a fresh fail receipt for the current review key, the block count for that key is below C, and the stop has not been released by the stop-hook-active flag, the turn-end gate shall block the stop and shall give the fail detail as the reason.

현행 턴 종료 차단을 유지한다. 모델이 조치할 수 있는 유일한 지점이기 때문이다(§0). 차단할 때마다 그 키의 차단 횟수를 1 올린다(design.md §3). C 의 값은 운영자 판정 Q2 에 따라 1 이다(REQ-RG-004).

### REQ-RG-004 — 키당 차단 횟수는 C 로 총량 제한된다 (event-driven, M3, C = 1)

When a turn ends with a fresh fail receipt for a review key whose block count has reached C, the turn-end gate shall allow the stop and shall show the unreviewed state in a visible system message.

차단 횟수는 리뷰 키 단위의 총량이다(보드 판정 D3 = 키당 총 C회로 제한 (허용이 카운터를 되돌리지 않음)). 허용은 차단 횟수를 되돌리지 않는다. 그 키에 fresh pass 영수증이 기록되면 그 키의 차단 횟수를 제거한다. 키가 바뀌면 새 키는 0 에서 시작한다. C = 1(운영자 판정 Q2)이므로, 같은 fail 키는 처음 한 번만 차단되고 그 뒤의 턴 종료는 모두 허용되며 메시지가 표시된다. 이 규칙은 턴 단위로 카운터를 되돌리는 방식이 만드는 차단·허용의 반복(sawtooth)을 막는다. 허용 뒤의 다음 턴 종료는 AC-RG-017 이 확인한다.

### REQ-RG-005 — 리뷰 기준은 하나의 해석 함수가 정한다 (Ubiquitous, M4)

The turn-end gate, the card-scope review tool, and the audit tools shall obtain the review base from one shared resolution function.

두 해석 경로가 같은 입력에 같은 기준을 내야 한다. 함수를 하나로 묶는 것만으로는 부족하고, 모든 호출 경로가 그 함수를 거쳐야 한다. 함수의 입력, 후보 집합, 거리 척도, 동점 규칙은 design.md §4 가 정한다.

### REQ-RG-006 — 기준은 후보 가운데 HEAD 에 가장 가까운 공통 조상이다 (state-driven, M4)

While the configured base refs yield at least one resolvable candidate, the shared resolution function shall select the candidate whose merge-base with HEAD has the fewest commits between that merge-base and HEAD.

후보 집합은 설정된 두 키 각각의 로컬 ref 와 원격 추적 ref 의 합집합이다(design.md §4.1). 거리는 `rev-list --count <merge-base>..HEAD` 다(design.md §4.2). 카드가 main 에서 잘렸다면 main 쪽 merge-base 가, develop 에서 잘렸다면 develop 쪽 merge-base 가 선택된다. 같은 거리의 동점 규칙은 design.md §4.3 이 정한다. merge-base SHA 가 다른 동점은 inconclusive 다(AC-RG-020).

### REQ-RG-007 — 두 설정 키가 다른 ref 를 가리켜도 판정은 막히지 않는다 (event-driven, M4, 운영자 판정 Q1)

When the configured integration target and the configured worktree base branch name different refs, the shared resolution function shall select the nearest candidate of the union of both keys' candidates under REQ-RG-006, and shall not return an inconclusive result on that ground alone.

운영자 판정 Q1 = (a) 합집합에서 최근접 후보. 키 불일치는 판정을 막지 않는다. 결과에는 두 설정 값의 이름을 함께 기록한다. 이 저장소에서 두 키는 항상 다르므로(§A.2), 이 요구는 이 저장소의 일상 경로다. 문언대로 키 불일치를 inconclusive 로 두면 모든 판정이 inconclusive 가 되므로 그 해석은 폐기되었다.

### REQ-RG-008 — 기준 후보가 없으면 다른 기준으로 넘어가지 않는다 (event-detected, M4)

When none of the configured base refs resolves in the tree, the shared resolution function shall return an inconclusive result and shall not substitute another base.

설정 밖의 ref(예: `origin/HEAD`)로의 대체도 금지된다. 현행 감사 경로의 단계 1–2 대체(`origin/HEAD`, `main`)는 이 요구와 충돌하며, 그 단언의 기대값 변경은 AC-RG-016 이 근거와 함께 명시한다.

### REQ-RG-009 — 재사용은 pass 와 fail 에만 허용한다 (Ubiquitous, M5)

The turn-end gate shall reuse a stored review receipt for an unchanged review key only when the stored verdict is pass or fail.

inconclusive 영수증은 pass 증거로 세지 않으며, 그 키에 대해 재사용하지 않는다.

### REQ-RG-010 — inconclusive 뒤의 턴 종료는 재검토를 시작하고 허용한다 (event-driven, M5, 간격 없음 Q4)

When a turn ends and the stored verdict for the current review key is inconclusive, the turn-end gate shall start a review in the background unless an in-flight marker for that key exists, and shall allow the stop.

운영자 판정 Q4 = (a) 간격 없음. 다음 턴 종료에서 다시 시작하며, 중복 방지는 진행 중 표식만 맡는다. 쿨다운은 없다. 허용은 SPEC-GATE-BOTTLENECK-001 REQ-GBN-004 의 fail-open 과 일치한다. 허용에 붙는 메시지는 REQ-RG-016 이 정한다.

### REQ-RG-011 — 실행 단위의 스위치는 주 체크아웃의 재정의 파일이고, 없으면 커밋된 값을 따른다 (Ubiquitous, M6, 운영자 판정 Q3·D11)

The codex review gate shall take its enablement for each run from one override file in the primary checkout that applies to every lane of that run. While that file is absent, or names a run other than the single active run, or no single active run resolves, the codex review gate shall take the committed workflow.yaml review_gate.enabled value.

운영자 판정 Q3 = (b) 주 체크아웃의 미추적 재정의 파일. 보드 판정 D11 = 커밋된 설정값을 따름: 재정의 파일이 없으면 게이트는 커밋된 `.moai/config/sections/workflow.yaml` 의 `codex.review_gate.enabled` 값을 따르고, 재정의 파일은 run 단위로만 그 값을 덮어쓴다.

- **재정의 파일의 위치**: 주 체크아웃의 `.moai/state/codex-review-gate/run-switch.json`. `.moai/state/` 는 `.gitignore` 401행으로 무시되므로 미추적이다. 파일 이름과 형식은 run 단계의 설계 세부다. 파일은 실행 식별자와 활성화 값(on/off)을 싣는다.
- **"run" 의 정의**: factory run 하나. `--run` 플래그가 가리키는 factory run id 로 지칭되며, 활성 run 은 동시에 하나다(`internal/cli/factory_card.go` 의 `--run` 도움말 "default: the single active run"). 리더가 `moai cc -f` 로 연 배치 하나가 한 run 이다.
- **커밋된 값**: 판독 트리의 HEAD 커밋에 있는 `.moai/config/sections/workflow.yaml` 의 `codex.review_gate.enabled`. 작업 트리의 미커밋 값은 쓰지 않는다. 판독은 `git show HEAD:` 같은 읽기 전용 조회다.
- **기본값의 이유**: 현재 Go 층은 빈 루트나 읽기 실패를 비활성으로 읽는다(`internal/cli/mcp_codex.go:2503–2510`). 그래서 기본값을 명시하지 않으면 루트를 풀지 못하는 경우 검토가 조용히 꺼진다. 기본값을 커밋된 값으로 정하면 이 공백이 사라진다.
- **적용 조건**: 재정의 파일은 자기가 가리키는 run 이 단일 활성 run 과 같을 때만 적용된다. 오래된 재정의 파일이 다음 run 에 조용히 남지 않게 하기 위한 조건이다.
- 추적 키 `codex.review_gate.enabled` 의 역할은 기본값의 원천으로만 남는다. 실행을 전역으로 끄거나 켜는 권한은 없다.

### REQ-RG-012 — 셸 층과 Go 층은 같은 스위치 값을 읽는다 (Ubiquitous, M6)

The shell wrapper and the Go handler shall resolve the override file at the same primary-checkout path and shall derive the committed default from the same committed source, so that both layers return the same enablement for the same run.

현재 Go 층은 `reviewGateConfigRoot`(`internal/cli/codex_review_gate.go:489`)를 거쳐 `auditreceipt.StoreRoot` 를 쓰는데, `StoreRoot` 는 트리에 workflow.yaml 이 있으면 그 트리를 그대로 돌려준다(`internal/auditreceipt/storeroot.go:243`). 따라서 재정의 파일의 경로는 `StoreRoot` 가 아니라 주 체크아웃을 돌려주는 `IdentifyPrimaryCheckout`(`internal/auditreceipt/storeroot.go:142`) 또는 같은 결과를 내는 해석기로 구한다. 셸 래퍼(`.claude/hooks/moai/handle-codex-review-entry.sh` 33–34행)는 지금 `CLAUDE_PROJECT_DIR` 의 workflow.yaml 을 읽으므로, 같은 해석을 따르도록 바뀐다. 해석기의 선택은 run 단계 설계 세부다.

### REQ-RG-013 — 현재 pass 영수증은 재검토 없이 허용한다 (State-driven, 회귀)

While a fresh pass receipt exists for the current review key, the entry hook and the turn-end gate shall allow without starting a new review.

### REQ-RG-014 — 검토 건너뜀의 결정은 바뀌지 않는다 (Ubiquitous, 회귀)

The codex review gate shall keep its current allow decision for disabled gates, active stop-hook releases, trees without reviewable changes, and absent reviewer binaries.

결정(허용)은 그대로다. 이 경우 턴 종료에 붙는 메시지는 REQ-RG-016 이 더한다. 메시지는 결정을 바꾸지 않는다.

### REQ-RG-015 — 턴 종료에 fresh 영수증이 없으면 재검토를 시작하고 허용한다 (event-driven, M3)

When a turn ends with no fresh receipt for the current review key (absent, stale, or key-mismatched), the turn-end gate shall start a review in the background unless an in-flight marker for that key exists, and shall allow the stop.

가장 흔한 턴 종료 상태다. 이 동작은 SPEC-GATE-BOTTLENECK-001 REQ-GBN-002 의 턴 종료 절(absent 또는 stale 이면 허용하고 백그라운드 검토를 시작한다)과 같다. 이 SPEC 은 그 요구를 이 SPEC 의 요구로 명시하며, 그 SPEC 은 편집하지 않는다(Q6 = (b)). 허용에 붙는 메시지는 REQ-RG-016 이 정한다.

### REQ-RG-016 — 검토 없이 허용되는 턴 종료는 상태를 명명한 메시지를 낸다 (event-driven, M3, 보드 판정 D4)

When the turn-end gate allows a stop and no fresh passing receipt backs that allow, the turn-end gate shall emit a visible system message that names the state.

보드 판정 D4 = 항상 메시지 표시. 이 문장은 "allows a stop without a fresh passing review" 를 문언대로 옮긴 것이다. 따라서 absent, stale, key-mismatch, inconclusive, 차단 횟수 소진, 스위치 꺼짐, 검토할 변경 없음, stop-hook 해제, reviewer 없음이 모두 포함된다. 키가 없는 상태(스위치 꺼짐, 검토할 변경 없음)도 포함하는 것은 §0 의 "검토가 되지 않은 상태는 드러나야 한다"에 따른다. 메시지는 Stop 훅의 `systemMessage` 로 낸다(`.claude/rules/moai/core/hooks-system.md:103`). 허용 결정 자체는 바뀌지 않는다(REQ-RG-014). 상태 어휘는 design.md §7 이 정한다.

---

## §C AC 와 검증 구속

- 모든 인수 기준은 acceptance.md 에 있다(AC-RG-001 ~ AC-RG-020, 요구와 1:1 이상 연결, 고아 기준 없음). 이 SPEC 은 plan 시점에 릴리스 게이트를 하나도 갖지 않는다. 이유: 행위 요구를 단언하는 테스트가 현재 트리에 없고(research.md §6, 테스트 이름 확인), 기존 행위 테스트는 결함을 고정할 뿐이다(acceptance.md §A). 릴리스 게이트로 승격하려면 run 단계가 구현 전 트리에서 RED 를 원문으로 기록해야 한다.
- 검증 명령은 항상 작업 트리 루트에서, 단일 호출로 실행한다. 실행된 테스트 수를 함께 기록하며, 0개는 측정 불가로 보고한다.

## §D 실행 순서 구속

1. 운영자 판정 Q1–Q6 과 보드 판정 D3·D4·D11 은 decision-index.md 에 기록되어 있다. run 단계는 여섯 Q 행의 `Operator verdict` 가 채워져 있는지 확인하고 시작한다.
2. 각 마일스톤은 RED(실패 확인) → GREEN → 회귀 순으로 진행한다. RED 출력은 progress.md §E.2 에 원문으로 남긴다.
3. **변경 세트(change set)는 커밋 하나다.** 두 면을 가진 파일은 같은 커밋에서 함께 갱신한다. 훅 래퍼(`.claude/hooks/moai/handle-codex-review-{entry,gate}.sh` 와 `internal/template/templates/.claude/hooks/moai/handle-codex-review-{entry,gate}.sh`)는 M6 의 한 커밋에 함께 담긴다. M7 의 규칙 문장 정정(F4)도 규칙 파일(`.claude/rules/moai/workflow/factory-dispatch-detail.md`)과 그 템플릿 미러를 한 커밋에서 맞춘다. 두 면이 서로 다른 커밋에 갈라지는 상태는 허용하지 않는다.

---

## §E 범위 밖 (Out of Scope)

### Out of Scope — 영수증 결합 기준의 변경 (F1)

- 영수증을 HEAD 대신 (merge base, 내용 digest) 에 묶는 변경은 이 SPEC 에서 하지 않는다. Follow-up F1 로 추적한다.
- 현재 결합 필드 다섯 개의 구성 변경도 하지 않는다.

### Out of Scope — 가시성, 중복 제거, kick 정리 (F2)

- 차단 사유의 중복 제거와 길이 상한, 미검토 상태의 systemMessage 표시 확장(REQ-RG-016 의 범위를 넘는 부분), kick 의 pid·liveness 기반 중복 제거와 레인 전체 동시성 상한, 마커 누수 정리, 추적 diff 가 비면 kick 하지 않는 규칙, 진입 훅 timeout 의 5초 정렬은 이 SPEC 밖이다. Follow-up F2 로 추적한다.

### Out of Scope — factory 저장소 경로 (F3)

- `internal/factory/todo_dispatch_current.go:53` 과 `internal/factory/todo_runtime.go:364` 의 두 주장은 이 SPEC 의 diff 에 없다. 카드 t1620 밖의 별도 후보로 추적한다.

### Out of Scope — 감사 도구의 나머지 결함

- `audit_multi` 의 codex 레그가 target 을 무시하는 문제, claude 자식 프로세스 env 의 denylist 방식, codex 샌드박스 정책 미전달, 작업 수명 정리 누락, factory 메시지의 null 반환은 카드 항목 밖이다. 연구 입력 lens-3 의 별도 후보로 남긴다.

### Out of Scope — doctor 진단 신설 (F4)

- 두 리뷰 계층이 동시에 켜졌을 때 doctor 경고를 추가하는 일은 이 SPEC 에서 하지 않는다(운영자 판정 Q5 = (b)). Follow-up F4 로 등록한다.

### Out of Scope — 두 리뷰 계층 중 하나의 폐기

- 턴 종료 게이트나 card-review 단계 가운데 하나를 없애는 결정은 이 SPEC 이 하지 않는다. 중복 검토는 후보 결함으로만 기록한다(규칙 문장 F4 와 연결).

### Out of Scope — GLM 턴 종료 게이트 신설

- 현재 GLM 턴 종료 게이트는 존재하지 않는다(research.md §3). 신설은 범위 밖이다.

### Out of Scope — SPEC-GATE-BOTTLENECK-001 의 편집 (F5)

- 대상 SPEC 의 REQ-GBN-002, REQ-GBN-004 와 AC-GBN-004 의 inconclusive 절은 편집하지 않는다(운영자 판정 Q6 = (b)). 처분 기록은 이 SPEC 의 HISTORY 와 F5 에만 있다.

### Out of Scope — 워크플로 설정 파일

- `.moai/config/sections/workflow.yaml` 은 편집하지 않는다. 스위치의 기본값은 그 파일의 커밋된 값을 읽을 뿐이다(REQ-RG-011).

---

## Follow-up

- **F1 — 영수증 결합 기준 변경.** 영수증을 HEAD 대신 (merge base, 내용 digest) 에 묶는다(카드 항목 4). 별도 후속 카드로 추적한다. 요구사항을 붙이지 않는다.
- **F2 — 가시성과 운영 정리.** 카드 항목 6–10: 미검토 상태를 systemMessage 로 드러내기(REQ-RG-016 밖의 범위), 차단 사유의 중복 제거와 길이 상한, kick 을 pid·liveness 로 중복 제거하고 레인 전체 상한을 두며 마커 약 64개 정리, 추적 diff 가 비면 kick 하지 않기, 진입 훅 timeout 을 5초에서 15초 예산과 맞추기. 요구사항을 붙이지 않는다.
- **F3 — factory 후보 두 건(카드 t1620 밖).** (a) `internal/factory/todo_dispatch_current.go:53` 의 백로그 SQLite 열기가 JSON 큐 이관 전에 빈 SQLite 를 만들 수 있다. (b) `internal/factory/todo_runtime.go:364` 의 할당 트랜잭션에서 외부 저장소 변경 뒤 내부 커밋(약 368행)이 실패하면 복구되지 않는다. 두 주장의 검증 결과는 research.md §7 에 있다. 요구사항을 붙이지 않는다.
- **F4 — 두 계층 동시 활성 doctor 점검 (운영자 판정 Q5 = (b)).** doctor 검사는 이 SPEC 에서 추가하지 않고 후속 항목으로 등록한다. 규칙 문장 `.claude/rules/moai/workflow/factory-dispatch-detail.md:157`(두 계층이 같은 카드 diff 를 두 번 검토한다는 문장)은 Q5 (b) 의 구성 요소로 run 단계(M7)에서 정정한다. 정정은 그 규칙 파일과 템플릿 미러를 같은 커밋에서 맞춘다. 이 plan 단계에서는 그 규칙 파일을 편집하지 않는다.
- **F5 — Q6 기록 처분 (운영자 판정 Q6 = (b)).** 이 SPEC 은 SPEC-GATE-BOTTLENECK-001 의 REQ-GBN-002(`spec.md:43–47`, 진입 차단 절), REQ-GBN-004(`spec.md:51–54`)와 AC-GBN-004(`spec.md:67–68`, inconclusive 절)를 편집하지 않는다. 처분은 이 SPEC 의 HISTORY 와 이 항목에만 기록한다. 잔여 위험: 대상 SPEC 은 in-progress 이며, 코드의 진입 차단 동작과 REQ-GBN-002 가 주석 없이 어긋난 상태로 남는다.

---

## §F 제약

- 이 SPEC 의 산출물은 `.moai/specs/SPEC-REVIEW-GATE-REPAIR-001/` 안에만 둔다.
- 작업 트리의 `.moai/config/sections/workflow.yaml` 은 건드리지 않는다.
- 기준 코드 SHA 는 `2aab5f797b75983e132af451da68f69e3426557b`(짧은 형식 `2aab5f797`)다. 설치된 빌드(`4f8aba061`)는 트리 기준의 후손 커밋이므로 트리의 증거로 인용하지 않는다.

## §G 참조

- research.md — 인용 재검증 등록부, 재현 근거, 전제 확인(§9: 수리 1회차 재검증)
- design.md — 상태 모델, 기준 해석 함수, 스위치 해석, 판정 표
- plan.md — 마일스톤과 기술 접근
- acceptance.md — 인수 기준 매트릭스와 기준선 셀
- decision-index.md — 운영자 판정 Q1–Q6 과 보드 판정 D3·D4·D11
- `.moai/specs/SPEC-CODEX-REVIEW-TARGET-001/spec.md` §A.7 · `.moai/specs/SPEC-GATE-BOTTLENECK-001/spec.md` REQ-GBN-002 · REQ-GBN-004 · AC-GBN-004
- `.moai/reports/t1620/plan-audit.md` §1.4 — 수리의 입력(감사 결함 D1–D12, 선택 항목 O1–O11). 이 SPEC 은 그 파일을 권한으로 인용하지 않는다.
