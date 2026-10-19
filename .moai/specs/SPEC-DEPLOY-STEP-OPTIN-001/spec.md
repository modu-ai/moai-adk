---
id: SPEC-DEPLOY-STEP-OPTIN-001
title: "배포본 자동화 단계별 명시 opt-in — 기본은 모두 수동, 설정은 하나의 해석기가 판정하고 MoAI Web이 보여 주고 검증한다"
version: "0.2.1"
status: draft
created: 2026-10-19
updated: 2026-10-19
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: "internal/config"
lifecycle: spec-anchored
tier: L
tags: "automation, opt-in, fail-closed, stage-policy, web-settings, update-preservation, template-first, card-t1895"
related_specs: [SPEC-REVIEW-GATE-REPAIR-001, SPEC-PROJECT-CONTINUATION-KEY-001, SPEC-TODO-ENABLE-FLAG-001, SPEC-UPDATE-YAML-PRESERVE-001, SPEC-GITHUB-FLOW-DEFAULT-001, SPEC-TODO-CLASSIFY-DISPATCH-001]
---

# SPEC-DEPLOY-STEP-OPTIN-001 — 배포본 자동화 단계별 명시 opt-in

카드: **t1895** (class C, P1, Template-First) · 기준 코드 트리 `8249bff8c3933d647b4f99e59048595f9808e83a` (작업 브랜치 `WT-opt-in-moai`) · Tier L (판정 근거는 plan.md §A)

## HISTORY

- 2026-10-19 · v0.2.1 · manager-spec · Q8 경계 재정 반영: 판정(project 계층, 명시 `auto`)은 유효하되 실제 설정 파일 쓰기(`.moai/config/sections/workflow.yaml`)는 운영자 keep-set 으로 별도 운영자 승인을 게이트로 요구한다(spec §D.1 · plan M6 (3)·§C·§F · design §6 · decision-index Q8).
- 2026-10-19 · v0.2.0 · manager-spec · 제품 결정 행 판정 기록과 일관성 스윕. Q1–Q9, Q14 판정(운영자 상시 규칙 — Jev, 불충분 시 Codex 교차; 리더 run tn564n 중계)을 decision-index 에 기록했다. Q4=(b) 로 REQ-DSO-017 · AC-DSO-021 · M7 을 빼 여섯 단계 SPEC 이 되었다(017·021 은 결번). Q5=(a) 로 단계별 fallback 확정. Q7=(c) 로 읽지 않는 `auto_push`/`auto_pr` 제거를 M6 에 둔다. Q8=(a) OPERATOR-HELD(리더 거부권 대기). Q9=(b) 로 즉시 중지를 `.moai/state/` 정지 표식으로 바꿨다. Q11 은 열려 있다.
- 2026-10-19 · v0.1.1 · manager-spec · 1회차 plan-audit(FAIL 0.85, `.moai/reports/t1895/plan-audit.md`) 개정. D1: REQ-DSO-016 을 쓰기+알림으로 정렬(결정 Q15, DECIDED). D2: REQ-DSO-010 의 게이트 지점을 공유 임대 이음매로 확장. D3: M2 의 Q3·Q4 재범위 기록. D4: AC-DSO-021 순서·결과 사용. D5: 완료 정의 (2) 재진술. D6 은 수정 없음.
- 2026-10-19 · v0.1.0 · manager-spec · 최초 작성(draft). 입력 메모(`.moai/reports/t1895/OPTIN-DESIGN.md`, ADK 읽기 전용 조사, main `74ad197380` 기준)는 가설 원천으로만 쓰고, 모든 전제를 이 트리에서 다시 측정했다(research.md 증거 원장 L01–L63). 메모와 달라진 전제는 research.md §3 에 모았다. 제품 수준 결정(Q1–Q9, Q14)과 증거 대기 결정(Q11)은 decision-index.md 에 열린 채로 두었고, 구현 수준 결정(Q10, Q12, Q13)은 규칙 적용 결과로 채웠다. 이 행은 감사 결과를 주장하지 않는다.

---

## §0 지배 원칙 [HARD]

> **배포본은 명시한 단계만 자동으로 움직인다. 명시하지 않았거나 읽을 수 없거나 잘못 쓴 단계는 수동이다. 이 판정은 코드 한 곳(해석기)에서 나고, 실행 경로와 웹 화면은 같은 판정을 읽는다.**

자동화에 동의한다는 것은 단계마다 따로 적는 일이다. git 모드, 원격 저장소 유무, 다른 단계의 동의는 동의가 아니다. 판정이 한 곳에서만 나야 웹 화면이 보여 주는 값과 실제 동작이 어긋나지 않는다(research.md §2 의 L20·L46 이 보여 주듯 지금은 키와 동작이 이미 어긋나 있다).

---

## §A 배경 (측정)

모든 항목은 research.md 의 증거 원장(명령, 원문 stdout, 종료 코드, 트리 SHA)에서 가져왔다. 원장 번호를 괄호로 단다.

### A.1 지금 외부 변경은 단계별 동의 없이 일어난다

- github-flow 에서 `moai factory complete` 의 전달 본체는 push(`factory_card_pr.go:395`) → PR 생성(`:410`) → auto-merge 요청(`:444`) → pr-open 기록(`:449`)을 이어서 호출한다(L03). 이 경로가 읽는 설정은 워크플로 종류 하나다(L09). 정책을 읽는 토큰이 이 파일에 없다(L08, 종료 코드 1).
- 네 번의 원격 변경 앞에 있는 재확인은 모두 lease·version·만료 확인이다(L54). 정책 재확인은 없다.
- t1882(PR draft·auto-merge 제거)의 변경은 이 트리에 없다: `--draft` 토큰 없음(L04), `"--auto"` 요청 그대로(L05), 이력에 t1882 없음(L06, L07). t1882 는 큐에서 `picked` 상태다.
- 자동 배차의 기본값은 코드에 있다. `laneDispatchSelection` 은 `--no-auto-dispatch` 가 없으면 항상 auto 를 돌려준다(`factory.go:323`, L11). 그 선택은 설정 키가 아니라 실행마다 찍는 환경 변수 `MOAI_FACTORY_AUTO_DISPATCH` 다(L12). 이 값을 읽는 곳은 안내문을 고르는 두 곳뿐이고(L46), `moai factory next` 는 인자 없이 부르면 우선순위 순서로 카드를 집는다(L45).

### A.2 설정 키는 있어도 동작과 이어져 있지 않다

- `git_strategy.<mode>.automation.auto_push` / `auto_pr` 은 구조체와 기본값에만 있고 읽는 비테스트 Go 코드가 없다(L20). 배포 템플릿의 team 프로필은 `auto_push: true` 로 나간다(L18, L19). 이 키로는 동의를 표현할 수 없다.
- `moai init` 은 원격 저장소가 있으면 `personal`, 없으면 `manual` 모드를 고른다(L21). 모드는 동의가 아니다.
- 이미 수동인 항목은 템플릿에서 확인했다: worktree `auto_create/auto_merge/auto_cleanup`(L22), `local_main_integration`·`candidate_ci`·`jev`(L24), `codex.review_gate`(L23). 이 저장소의 배포본은 `auto_create: true`, `auto_cleanup: true`, review gate `true` 를 자기 설정으로 갖는다(L25, L53). 보존 대상이다.
- 기본값이 수동 쪽으로 닫혀 있지 않은 이웃 키가 있다. `workflow.project.continuation` 은 키가 없거나 값이 틀려도 `card` 로 풀려 카드 발급 쪽으로 열린다(L13, L14). `workflow.todo.enabled` 는 키가 없거나 설정을 읽지 못해도 켜진 것으로 풀린다(L17). 카드가 말한 "project.continuation 키 누락" 항목의 실측 결과는 이렇다: 키는 템플릿(L15), 웹 스키마(L16), Go 기본값(L14)에 모두 있다. 문제는 키의 누락이 아니라 **누락이 열린 쪽(card)으로 풀리는 것**이다.

### A.3 웹은 편집은 하지만 효력·출처·변경 기록을 모른다

- `/settings`, `POST /save` 는 있고(L26), 루프백 Host 와 same-origin 검사를 거치며(L27), 저장은 직렬화되고 검증 실패는 쓰기를 멈춘다(L28). 파일 쓰기는 임시 파일 뒤 rename 이다(L32).
- 읽기 모델은 문자열 스칼라만 돌려준다(L31). revision 개념이 웹과 설정 패키지에 없다(L29). 저장 감사 기록을 쓰는 코드가 없다(L30).
- 기존 8단 설정 해석기(`config.Source` 의 builtin/project/local 등)는 있으나 호출자가 doctor 하나이고(L41), local 계층 경로를 CWD 기준 상대 경로로 읽는다(L40). 이 트리에는 `.moai/config/local` 디렉터리가 없다(L42). 이 디렉터리는 저장소에서 무시된다(L49).

### A.4 update 와 보존

- 3-way 병합은 사용자 값이 base 와 같으면 새 템플릿 값을 쓴다(`node_merge.go:449`, L33). 템플릿에 없는 사용자 키는 모두 보존한다(L34). 따라서 **템플릿이 값을 싣지 않으면** "사용자가 명시한 값이 예전 기본값과 같다"는 구분 문제가 생기지 않는다. 값을 싣는 설계는 나중에 기본값이 바뀔 때 명시 값을 조용히 갈아치울 수 있다(design.md §6).
- `AGENTS.local.md`·`CLAUDE.local.md` 는 배포 템플릿에 없다(L35, 양성 대조 L36).

### A.5 카드 항목과 관계

- 외부 변경 호출 지점에서 이 SPEC 의 run 은 t1882(같은 `factory_card_pr.go`) 착지 뒤에 시작한다. t1752(완료 기준 설정)는 웹 스키마 파일이 겹친다. t1884 는 폐기·재발급(t1900)되어 문서 정합은 그쪽이다. t1644(웹 2페이지 재구성)는 보류이며 이 SPEC 은 설정 섹션만 다룬다. 파일 겹침과 순서 결과는 plan.md §E.

---

## §B 요구사항 (GEARS)

단계 이름과 키 위치는 Q1 판정(`workflow.automation.stages.<stage>`), 값 형태는 Q2 판정(닫힌 문자열 `manual`/`auto`)으로 확정되어 더 이상 작업 이름이 아니다. 이 SPEC 은 여섯 단계 `todo_creation` / `dispatch` / `push` / `pr_creation` / `review_repair` / `merge` 를 다룬다. release/deploy 단계는 Q4=(b) 판정(2026-10-19)으로 이 SPEC 의 범위에서 빠졌다(§E).

### REQ-DSO-001 — 여섯 단계는 서로 독립이다 (Ubiquitous, M2)

The stage-policy resolver shall define exactly six automation stages — `todo_creation`, `dispatch`, `push`, `pr_creation`, `review_repair`, `merge` — and shall resolve each stage to exactly one of `manual` or `auto`.

### REQ-DSO-002 — 명시한 유효한 auto 만 auto 다 (Ubiquitous, M2)

The stage-policy resolver shall resolve a stage as `manual` unless a valid `auto` is explicitly present for that stage, and shall attach a reason code to every `manual` that is not an explicit `manual`.

### REQ-DSO-003 — 누락·잘못된·부분 설정은 해당 단계를 수동으로 푼다 (event-driven, M2)

When a stage's value is absent, null, of the wrong type, outside the domain or duplicated, or when the stage block or the whole file is absent, unreadable or not a mapping, the stage-policy resolver shall resolve each affected stage as `manual` with a reason code and shall resolve every other stage by its own explicit valid value.

잘못된 값 하나가 지우는 범위는 Q5=(a) 판정(2026-10-19)으로 정해졌다: 잘못된 값이 있는 단계만 수동이고 유효한 형제 단계는 그대로 읽는다. 블록 전체가 수동이 되는 경우는 블록 자체를 읽지 못할 때뿐이며, 그때는 지정된 단계가 없으므로 여섯 단계가 모두 수동이다.

### REQ-DSO-004 — 한 단계의 auto 는 다른 단계를 열지 않는다 (Ubiquitous, M2·M3)

The stage-policy resolver shall treat the seven stages independently, and an automatic stage shall not perform, imply or enable any earlier or later stage.

물리적으로 앞 단계가 있어야 하는 단계(PR 생성은 원격 브랜치가, merge 요청은 열린 PR이 필요하다)는 앞 단계를 대신 수행하지 않고, 수동인 앞 단계의 이름을 대며 멈춘다(REQ-DSO-008, REQ-DSO-009).

### REQ-DSO-005 — 판정은 값·출처·이유·revision 을 함께 돌려준다 (Ubiquitous, M2)

The stage-policy resolver shall return, for each stage, the effective value, the source (`default`, `project`, `local-override` or `stop-marker`), the reason code, and one policy revision that changes if and only if a stage-relevant input of any tier changes.

`local-override` 가 `project` 를 이긴다. 출처는 값이 실제로 어느 계층에서 왔는지를 뜻한다. 키가 없으면 `default` 다. 정지 표식이 서면 그것이 모든 계층을 이긴다(REQ-DSO-015, 결정 Q9).

### REQ-DSO-006 — 판정은 한 곳에서만 난다 (Ubiquitous, M2·M3)

All runtime callers and the web settings view shall obtain stage values from the stage-policy resolver, including prose workflows through the resolver's read verb, and no other code shall read the stage keys.

### REQ-DSO-007 — push 는 push 단계가 auto 일 때만 한다 (event-driven, M3)

When the card delivery edge reaches the remote push and the `push` stage resolves to `manual`, the delivery edge shall stop before pushing, leave the card record unchanged, and name the stage and the way to enable it or to push by hand in its message.

### REQ-DSO-008 — PR 생성은 pr_creation 이 auto 일 때만 하고 push 를 대신하지 않는다 (event-driven, M3)

When the card delivery edge reaches pull request creation, the delivery edge shall create the pull request only if the `pr_creation` stage resolves to `auto`, and when the card branch is not on the remote because `push` is `manual`, it shall stop and name the `push` stage instead of pushing.

### REQ-DSO-009 — merge 요청은 merge 가 auto 일 때만 한다 (event-driven, M3)

When the card delivery edge reaches the merge request and the `merge` stage resolves to `manual`, the delivery edge shall not request merge or auto-merge, and shall record the pull request as opened.

t1882 가 착지하면 이 요구의 대상 호출이 바뀐다. 요구는 호출 형태가 아니라 "merge 단계가 auto 가 아니면 병합 요청이 없다"에 묶는다(plan.md §E, 결정 Q11).

### REQ-DSO-010 — 자가 배차는 dispatch 가 auto 일 때만 한다 (event-driven, M3)

When a lane session would pick its next card by itself, the lane launch, the session-start lane rule, and every self-dispatch entry that leases through the shared card-lease seam — the bare `moai factory next` verb, the MCP `factory_next` verb, and the `todo --auto-lane` lease path — shall treat the `dispatch` stage as the gate at that shared seam, a stage that is not `auto` shall make each entry take the manual-dispatch path, and a nominated `moai factory next --card <id>` shall remain allowed.

게이트는 진입마다 따로 두지 않고 세 진입이 함께 거치는 공유 임대 이음매 한 곳에 둔다. 이음매 뒤의 새 자가 배차 진입도 같은 게이트를 거친다.

`--no-auto-dispatch` 는 계속 수동으로만 좁힌다. auto 로 넓히는 실행 옵션은 두지 않는다.

### REQ-DSO-011 — 자동 카드 생성은 todo_creation 이 auto 일 때만 한다 (event-driven, M3)

When an automatic producer asks the todo queue to create a card, the card-creation verb shall create it only if the `todo_creation` stage resolves to `auto`, and an explicit operator `moai todo add` shall never be gated.

`workflow.project.continuation` 은 표현 선택 키로 남고 카드 발급 여부는 이 단계가 정한다(Q6=(a) 판정, 2026-10-19).

### REQ-DSO-012 — 리뷰 지적 수리 요청은 review_repair 가 auto 일 때만 한다 (event-driven, M3)

When the codex review gate would keep a turn going to have the review findings repaired and the `review_repair` stage resolves to `manual`, the gate shall allow the turn to end with a system message that names the state, and the review itself shall still run and report as configured.

### REQ-DSO-013 — 끄는 일은 다음 경계에서 효력이 나고 진행 중 작업을 죽이지 않는다 (State-driven, M3)

While a stage is changed, the runtime shall re-resolve the policy at each card lease, execution start and external-change boundary, shall not terminate or delete a running command, worktree, lease or evidence file, and shall not re-run a boundary a card has already passed.

경계가 통과될 때 그 판정의 revision 을 전환 기록에 남긴다.

### REQ-DSO-014 — update 는 명시 설정을 덮어쓰지 않는다 (Ubiquitous, M5)

The update path shall preserve explicit stage values, including values equal to any earlier default, comments, unknown keys, the local override tier and operator-approved local factory settings, and shall not change the effective value of any stage for a project that carries no stage block.

### REQ-DSO-015 — 웹은 단계별 효력·출처·영향을 보여 주고 저장 시 검증하며 기록을 남긴다 (Ubiquitous, M4)

The web settings view shall show for each stage the effective value, the source, an override marker, a plain-language impact statement and the policy revision, shall validate a save before writing, shall answer a save with the re-read policy and revision, shall record each save in an audit record, and shall offer a stop action that writes the stop marker and a release action that removes it; while the marker stands every boundary shall treat every stage as `manual` regardless of tier values, and the view shall show the marker state.

즉시 중지는 설정 값을 쓰지 않는다(Q9=(b) 판정, 2026-10-19). 표식은 `.moai/state/` 아래 파일이고, 명시적 release 로만 지워지며 조용히 만료되지 않는다(잊힌 표식이 모든 것을 수동으로 두는 쪽이 안전하다). 표식은 local override 도 이긴다.

### REQ-DSO-016 — 웹 저장은 안전 경계 안에서만 쓴다 (Unwanted, M4)

When a stage save arrives from a non-loopback Host, a cross-site origin, or with an invalid, duplicate or unknown stage, the web server shall reject it without writing, and concurrent saves shall be serialized.

When a stage save targets a stage pinned by the local override tier, the web server shall write the new project-tier value, shall leave the stage's effective value and the override untouched while the override stands, and shall disclose that fact (`overridden-by-local`) in the save response and the audit record.

로컬이 고정한 단계의 저장을 거부하지 않는 이유: 세 계층 모델에서 project 계층의 쓰기는 local override 를 이길 수 없으므로 "override 를 덮어쓰지 않는다"는 규칙은 그대로 지켜진다. 거부하면 override 가 사라진 뒤에 적용될 운영자의 의사가 기록되지 않는다. 즉시 중지가 모든 계층 위에서 표식으로 작동하는 것(REQ-DSO-015)과 별개의 규칙이며, 저장과 표식은 서로를 대체하지 않는다(결정 Q15, Q9).

### REQ-DSO-018 — 템플릿은 단계 값을 싣지 않고 기본값은 코드에서 수동이다 (Ubiquitous, M2·M6)

The shipped templates shall carry the stage keys only as commented documentation and shall carry no stage value, the code default of every stage shall be `manual`, and the project mirror of every changed template file under `.claude/`, `.agents/` and `.codex/` shall be byte-identical to its template source.

`.moai/config/sections/*.yaml` 은 이 저장소에서 운영자 자신의 파일이라 템플릿과 다르다(research.md L56). 그 파일들은 바이트 일치 대신 주석 블록이 양쪽에 있는지로 확인한다. 기본값을 `auto` 로 바꾸는 일은 호환 파괴 변경이며 이 SPEC 의 범위 밖이다.

---

## §C AC 와 검증 구속

- 모든 인수 기준은 acceptance.md 에 있다(AC-DSO-001–020, 022 — 021 은 Q4=(b) 판정으로 철회된 결번이다). 요구와 1:1 이상 연결되고 고아 기준은 없다.
- 이 SPEC 은 plan 시점에 구현 테스트를 갖지 않는다. 각 기준의 RED-now 셀은 현재 트리의 결함이나 부재를 보이는 증거 원장 항목을 가리킨다(원장 항목은 명령, 원문 stdout, 종료 코드, 트리 SHA 네 요소를 모두 가진다). RED-now 를 만들 수 없는 기준은 회귀 가드로 분류하고 통과로 기록하지 않는다.
- 검증 명령은 작업 트리 루트에서 단일 호출로 실행하며, 실행된 테스트 수를 함께 기록한다. 0개는 측정 불가다.

## §D 실행 순서 구속

1. 결정 색인의 행 가운데 각 마일스톤을 막는 행의 `Operator verdict` 가 채워져야 그 마일스톤이 시작한다(plan.md M1). 2026-10-19 개정 2 기준: M2 의 Q1·Q2·Q5, M3 의 Q3·Q6·Q14, M4 의 Q9, M5·M6 의 Q7·Q8 이 판정되었고, M3 의 Q11(EVIDENCE-NEEDED)만 열려 t1882 착지까지 M3 을 막는다. M7 은 Q4=(b) 판정으로 철회되었다. M2 는 Q3·Q4 판정을 기다리지 않고 모델을 만들며, 판정이 M2 뒤에 떨어지면 재범위(re-scope) 기록으로 진행한다(plan.md M1) — Q4=(b) 재범위는 개정 2에서 이 기록대로 실행되었다.
2. run 은 t1882 가 착지한 뒤에 시작한다. 착지한 트리에서 AC-DSO-006–008 의 RED-now 셀을 다시 측정하고 다시 고정한다(verification-completeness §4).
3. 각 마일스톤의 RED 선행은 커밋 그래프로 검증한다(plan.md §D 마일스톤 순서 조항). SPEC 도구에 순서 검증 동사는 없다(research.md L43, L44).
4. 이 저장소 자신의 승인된 로컬 factory 설정을 새 단계 키로 옮기는 일(M6 (3))은 템플릿 변경과 별도 커밋이며 **운영자 keep-set 게이트 단계다** — 별도 운영자 승인을 받은 뒤에만 실행하고, 승인 전에는 건너뛰고 보고한다(결정 Q8, 리더 재정 2026-10-19).

---

## §E 범위 밖 (Out of Scope)

### Out of Scope — 이 SPEC이 만들지 않는 것

- 사용자·역할별 접근 제어(ACL)와 신원 확인. 웹은 기존 루프백 Host·same-origin 검사 안에서 동작하며, 감사 기록의 호출 주체는 그 채널에서 얻을 수 있는 값(운영체제 사용자, 채널 이름)까지다(결정 Q10, 구현 수준).
- 웹에서 local override 계층을 쓰는 기능. 웹은 project 계층만 쓰고 local 계층은 읽기 전용으로 표시한다.
- 실행 중인 명령·세션·lease·워크트리를 강제로 끝내는 기능. 끄는 일은 다음 경계에서 효력이 난다.
- release/deploy 단계의 선언과 게이트, 새 게시 호출을 막는 정적 검사. Q4=(b) 판정(2026-10-19)으로 REQ-DSO-017 · AC-DSO-021 · 마일스톤 M7 을 이 SPEC 에서 빼고 후속 SPEC 이 가져간다. 배포본에 게시 경로가 생길 때 그때 정의한다(배포본에는 그런 경로가 없다, L51).
- 로컬 commit(`auto_commit`)의 정책화. 여섯 단계는 외부로 나가는 변경과 자동 진행을 뜻하며 로컬 commit 은 포함하지 않는다.
- 기존 github-flow 사용자의 동작을 update 가 조용히 auto 로 옮겨 주는 migration. 새 단계 키의 부재는 수동이다.
- `AGENTS.local.md`·`CLAUDE.local.md` 의 배포 또는 수정. 두 파일은 배포 템플릿에 없다(L35).

### Out of Scope — 다른 카드의 몫

- PR draft 게시·auto-merge 제거·ready 전환(t1882). 같은 `factory_card_pr.go` 를 건드리므로 순서만 이 SPEC 이 따른다.
- 카드 완료(done) 기준 설정 `git_strategy.done_on`(t1752). 웹 스키마 파일이 겹치지만 키와 동작이 다르다.
- 문서·로컬 지침 정합(t1884 폐기, 재발급 t1900). 이 SPEC 의 결과를 문서에 반영하는 일은 그쪽이다.
- 웹 2페이지 재구성(t1644, 보류). 이 SPEC 은 설정 섹션 하나만 더한다.
- 완료된 SPEC-REVIEW-GATE-REPAIR-001 의 개정. 새 상태 이름은 이 SPEC 이 정의하고 완료 SPEC 의 어휘 표는 건드리지 않는다.

---

## §F 후속·잔여 위험

- **F1** 웹 화면의 영향 설명 문구(ko/en)는 run 단계에서 i18n 키로 만든다. 문구 자체의 번역 품질 검토는 sync 감사 몫이다.
- **F2** `git_strategy.<mode>.automation.auto_push/auto_pr` 은 읽는 코드가 없다(L20). Q7=(c) 판정(2026-10-19): 구조체와 템플릿에서 제거하고(M6), 기존 사용자 파일의 키는 update 가 보존한다(REQ-DSO-014).
- **F3** `dispatch` 단계가 막는 것은 자가 배차(인자 없는 `moai factory next`, MCP `factory_next`, `--auto-lane` 임대와 안내문·규칙 선택)다. 리더가 `--card` 로 지명하는 배차는 운영자 행위다. Q14=(a) 판정(2026-10-19)으로 확정되었다.
- **F4** 이 저장소의 승인된 로컬 factory 설정(L25, L53)은 새 해석기에서 키가 없으므로 수동으로 읽힌다. 현행 동작을 유지하려면 명시 `auto` 를 적어야 한다. Q8=(a) 판정(OPERATOR-HELD): project 계층에 적는다(M6). 판정은 유효하되 실제 설정 파일 쓰기는 운영자 keep-set 이며 별도 운영자 승인을 게이트로 요구한다(리더 재정 2026-10-19).
