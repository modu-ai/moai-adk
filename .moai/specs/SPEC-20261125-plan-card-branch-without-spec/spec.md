---
id: SPEC-20261125-plan-card-branch-without-spec
title: "PLAN and SDD text: a card branch without a SPEC for tier 0 and 1"
status: approved
revision: 1
created: 2026-11-25
approval:
  by: user
  at: 2026-11-25
  basis: "사용자가 리더 대화의 선택지에서 직접 고른 답(카드 노트 9186): \"승인, D-1~D-3 기본값 (Recommended)\"; 리더가 기록 직전에 moai spec show 로 digest 를 다시 읽었어요"
  digest: "sha256:c297e8bc36bdd9f01eb8597c6dbd8d89c8211ea0763c583b000fd84f29003f21"
---

# PLAN and SDD text: a card branch without a SPEC for tier 0 and 1

## Interview

- Q1: SPEC 은 어떤 변경에 필요한가요? 지금 `sdd.md` 는 "Every change that modifies the repository has a SPEC" 라고 말하고 `plan.md` 는 모든 계획을 SPEC 으로 끝내요. → A: tier 2 변경(삭제, 권한, 관문, HOME)에만이고, 나머지는 카드 본문과 기준이 설계예요. 근거는 기록된 사용자 결정(카드 본문의 "결정: SPEC 은 tier 2 변경에만, 나머지는 카드 본문과 기준이 설계"와 저장소 규칙 문서 Flow 3단계)이고, 이 SPEC 이 새로 묻는 것은 아니에요. 다른 선택지는 모든 변경에 SPEC(지금 글), feature 카드에만 SPEC 이었어요.
- Q2: 감사는 무엇에 FAIL 을 주나요? `sdd.md` 의 감사 문장은 "FAILs on any drift" 라고 해요. → A: 네 가지만(기준과 다른 동작, 데이터 손실, 보안 문제, 약해진 관문)이고 나머지는 비차단 줄이에요(`references/audit.md` "What blocks"). 같은 사용자 결정의 기록이에요.
- Q3: 사실 한 가지를 묻는 READ 도 여러 갈래로 나누나요? → A: 아니요. 출처 하나로 답이 나는 한 가지 사실 물음은 나누지 않아요(카드 기준 5, 전수 조사 보고서의 `read.md` 항목).
- Q4: SPEC 없는 카드의 브랜치 이름과 묶음 착지는 어떻게 하나요? → A: 아직 정해지지 않았어요. 아래 Decisions 의 D-1, D-2 가 선택지와 기본값을 가진 열린 결정이고, 승인은 사용자가 해요.

## Intent

PLAN 글은 모든 계획이 SPEC 과 승인으로 끝난다고 말하고, SDD 글은 "모든 변경에 SPEC" 과 "감사는 어떤 어긋남에도 FAIL" 을 말해요. 이는 tier 2 만 SPEC 을 쓰고 FAIL 은 네 가지뿐이라는 기록된 결정과 달라서, 레인이 tier 0·1 카드에도 SPEC 과 승인을 만들고 승인 대기로 멈춰요. 이 SPEC 은 계획 글에 두 갈래(SPEC 없는 카드, SPEC 과 승인이 있는 tier 2)를 단계로 쓰고, 계획 에이전트가 돌려준 쓰기 집합과 열린 물음이 어디로 가는지 적고, SDD 글을 같은 결정에 맞추고, READ 글에 한 가지 사실 물음의 갈래를 더해요. 코드와 관문은 바꾸지 않아요.

## Non-goals

- `moai` 명령, 승인 digest 계산, 착지 관문, 감사 관문은 바꾸지 않아요. 승인과 digest 줄은 그 일을 하는 다른 SPEC 의 몫이고 여기서 고치지 않아요.
- SPEC 없는 카드를 묶음 착지가 받게 하는 코드는 이 SPEC 의 일이 아니에요(D-2 의 기본이 그 간극을 글에 적고 later 줄로 남겨요).
- `run.md`, `sync.md`, `audit.md` 와 다른 글은 고치지 않아요. 같은 어긋남이 거기 있으면 보고만 해요.

## Requirements

- REQ-1: When the PLAN workflow plans a card, it shall sort the card by its Write set into tier 0 or 1, where the card body and criteria are the design and no SPEC is written, or tier 2, where a SPEC the user approved is the design. · failure: none — 분류는 읽기만 하고 아무것도 쓰지 않아요.
- REQ-2 (REQ-1): IF the tier cannot be told from the Write set, THEN the PLAN workflow shall treat the card as tier 2 and record why in the SPEC.
- REQ-3: When a card is tier 0 or 1, the PLAN workflow shall take the criteria of the card as the acceptance criteria and write no SPEC and ask for no approval. · failure: none — SPEC 이 없으면 승인 단계도 없어요.
- REQ-4: When a card is tier 2, the PLAN workflow shall write the SPEC with its `## Interview`, state the open choices with their options and default, and hand only the revision the user approved to RUN. · failure: none — 승인 기록은 지금처럼 사용자의 선택만이에요.
- REQ-5: When `moai-plan` returns its Write set, Shared spots and Tests that read changed text, the PLAN workflow shall say that the Write set and those tests become the card's scope and that each Shared spot is written under `moai card line-lease`. · failure: none — 계획이 비면 범위는 카드가 가진 대로예요.
- REQ-6: When `moai-plan` returns Open questions, the PLAN workflow shall say that a tier 2 card writes each as an open-question marker of `internal/sdd` in its SPEC and a tier 0 or 1 card writes the unresolved ones as a `LANE-BLOCKED kind=decision` note on the card. · failure: none — 열린 물음이 없으면 줄이 없어요.
- REQ-7: When the SDD reference states when a SPEC is needed, it shall say that a SPEC is for tier 2 changes and that for every other change the card body and criteria are the design. · failure: none — 이 문장이 옛 "모든 변경" 문장을 바꿔요.
- REQ-8: When the SDD reference states what an audit fails, it shall say that a FAIL is for the four kinds of `references/audit.md` only and that any other drift is a non-blocking finding. · failure: none — 같은 말을 한 곳(`references/audit.md`)이 소유해요.
- REQ-9: When a READ question is one fact that one source answers, the READ workflow shall answer it directly without fanning out to several `moai-research` agents. · failure: none — 여러 갈래 물음은 지금처럼 나눠요.
- REQ-10: The PLAN workflow and the SDD reference shall name no Issue and carry no sentence that ties a SPEC or a card to a pull request, except that the only pull request is a release's. · failure: none — 글자 정리라서 동작이 없어요.

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-3, REQ-4): `workflows/plan.md` 의 절차에 tier 0·1 갈래(SPEC 없음, 카드 기준이 기준)와 tier 2 갈래(SPEC, Interview, 승인)가 단계 둘로 있고, 모르면 tier 2 라는 줄이 있어요 · verify: `cargo test --locked --offline --test sdd_guidance card_branch_ac1_`
- AC-2 (REQ-5, REQ-6): `workflows/plan.md` 가 쓰기 집합과 테스트가 카드의 범위(`moai card scope`)로, 공유 지점이 `moai card line-lease` 로, 열린 물음이 SPEC 의 열린 물음 표지(`internal/sdd` 5항의 표지)나 `LANE-BLOCKED kind=decision` 노트로 간다고 말하고, 쓴 명령은 모두 이 빌드의 도움말에 있어요 · verify: `cargo test --locked --offline --test sdd_guidance card_branch_ac2_`
- AC-3 (REQ-7, REQ-8, REQ-10): `references/internal/sdd.md` 가 "SPEC 은 tier 2 변경" 문장과 FAIL 네 가지 문장을 갖고, "Every change that modifies the repository has a SPEC" 와 "FAILs on any drift" 와 "cards stay personal to-do items" 가 없고, `workflows/plan.md` 에 `Issue` 낱말이 없어요 · verify: `cargo test --locked --offline --test sdd_guidance card_branch_ac3_`
- AC-4 (REQ-9): `workflows/read.md` 에 한 가지 사실 물음은 `moai-research` 로 나누지 않는다는 줄이 있고 갈래를 나누는 줄은 그대로예요 · verify: `cargo test --locked --offline --test sdd_guidance card_branch_ac4_`
- AC-5 (REQ-1…REQ-10): 글을 고정한 이웃 테스트가 그대로 통과해요(고정 문장은 이유와 함께 고쳐요) · verify: `cargo test --locked --offline --test skill_structure`

## Scope

- write: `.moai/specs/SPEC-20261125-plan-card-branch-without-spec/`
- write: `skills/moai/workflows/plan.md`, `skills/moai/workflows/read.md`, `skills/moai/references/internal/sdd.md`
- write: `tests/sdd_guidance.rs`, `tests/agent_delegation_text.rs`, `tests/skill_structure.rs`
- write: `tests/fixtures/update_tux/*.json` (스킬 글 해시가 바뀌는 만큼만)
- line-lease: `skills/moai/references/source-manifest.json` (해시 줄만, 같은 커밋)

## Plan

1. 테스트를 먼저 써요: `tests/sdd_guidance.rs` 에 `card_branch_ac1_` … `card_branch_ac4_` 와 옛 "모든 변경" 고정 문장의 교체.
2. `plan.md`: Output·절차에 두 갈래, 쓰기 집합·열린 물음 자리, Issue 문장. `moai-plan` 이 돌려주는 절 이름은 `agents/plan.md` 의 것을 그대로 써요.
3. `sdd.md`: 항목 1(When)·항목 2 앞머리의 "personal to-do" 문구·감사 문장. 승인 항목(Approval)은 손대지 않아요.
4. `read.md`: 한 가지 사실 물음 갈래.
5. 매니페스트 해시와 `update_tux` 고정 파일, 이웃 테스트, 카드 영수증.

## Risks

- 같은 날 승인 digest 를 바꾸는 SPEC 이 `plan.md` 와 `sdd.md` 의 승인 줄을 고쳐요. 이 SPEC 은 그 줄을 건드리지 않고 줄 단위로 맞춰요. 먼저 착지하는 쪽 위에 다른 쪽이 줄만 이어 붙여요.
- 고정 테스트가 옛 문장을 지켜요(`tests/sdd_guidance.rs` 의 "Every change that modifies the repository has a SPEC" 와 `ORDER` 문장). 문장을 바꾸면 테스트도 같이 바꾸고 이유를 SPEC Result 에 적어요.
- SPEC 이 없는 카드의 `Moai-Spec` 트레일러와 묶음 착지: `moai trailer-check` 는 `spec/` 가 아닌 브랜치에서 `Moai-Spec` 이 없어도 통과해요(도움말에서 읽었고 실행하지 않았어요). `moai card land-batch` 는 SPEC 을 이름 짓지 않은 카드를 `no_spec` 으로 떨궈요(`src/card/land/batch.rs` 를 읽었고 실행하지 않았어요). 단일 `moai card land` 의 판정 검사(`src/card/land/check.rs`)에서는 SPEC 을 읽는 곳을 찾지 못했어요(끝에서 끝까지 실행으로 확인하지 않았어요).

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | SPEC 없는 카드의 브랜치는 `spec/` 로 시작하지 않는 이름(기본 `card/<slug>-<yyyymmdd>`, 슬러그는 일의 이름이고 카드 번호가 아니에요)으로 예약해요. 대안은 (나) `spec/<slug>-<yyyymmdd>` 를 그대로 쓰고 최소 SPEC 을 만들기(결정과 어긋나요), (다) 이름 모양을 이 SPEC 에서 정하지 않고 카드가 예약한 이름을 그대로 쓰기 | `spec/` 브랜치에서는 SPEC 이 먼저 있어야 파일을 쓸 수 있고(도구 사전 점검 hook), 트레일러 문서의 4항("A commit that belongs to no SPEC does not go on a `spec/` branch")이 있어요 | 글 한 줄을 되돌려요 |
| D-2 | 묶음 착지가 SPEC 없는 카드를 받을 때까지 tier 0·1 카드는 `moai card land` 로 하나씩 착지해요(글이 그 간극을 말하고 later 줄이 코드 변경을 이름 지어요). 대안은 (나) SPEC 없는 카드도 묶음에 넣을 수 있게 코드를 먼저 고치고 글은 그 뒤에 쓰기, (다) 묶음 때문에 모든 카드에 최소 SPEC | 묶음 착지는 합친 커밋에 `Moai-Spec` 을 달아서 SPEC 이 없으면 `no_spec` 으로 떨궈요(`batch.rs`) | 글 한 줄을 되돌려요 |
| D-3 | 계획 단계에서 tier 를 모르면 tier 2 로 다뤄요 | `references/audit.md` 의 바닥 규칙이 모를 때가 아니라 경로로 정해서, 모르는 쪽이 안전해요 | 한 줄을 되돌려요 |

## Result

구현 head 는 카드의 READY 노트에 있어요(영수증도 거기에 있어요).

- AC-1: met — `cargo test --locked --offline --test sdd_guidance card_branch_ac1_` 가 통과했어요(1개). `plan.md` 에 `## Card branches` 절이 있고, 1단계 Sort(바닥 목록 경로나 삭제·권한·관문·HOME 변경이면 tier 2, 아니면 tier 0 또는 1, 모르면 tier 2), 2단계 tier 0 또는 1(SPEC 없음, 승인 없음), 3단계 tier 2(SPEC, Interview, 승인)가 있어요.
- AC-2: met — `card_branch_ac2_` 가 통과했어요. 4단계가 계획 에이전트의 Write set 과 바뀐 글을 읽는 Tests 를 카드 범위(`moai card take --scope`, `moai card scope`)로, Shared spots 를 `moai card line-lease` 로, Open questions 를 tier 2 SPEC 의 열린 물음 표지나 `LANE-BLOCKED kind=decision` 노트로 보내요. 테스트가 절이 이름 지은 모든 `moai card <verb>` 의 `--help` 를 실행했고 모두 있었어요.
- AC-3: met — `card_branch_ac3_` 가 통과했어요. `sdd.md` 가 "A SPEC is for tier 2 changes" 를 말하고, 카드 본문과 기준이 그 아래 변경의 설계이고, SPEC 없는 카드는 `spec/` 로 시작하지 않는 브랜치에서 일하고, 감사는 네 가지에만 FAIL 을 준다고 해요. 옛 문장 셋("Every change that modifies the repository has a SPEC", "FAILs on any drift", "cards stay personal to-do items")은 없어졌고 `plan.md` 에 `Issue` 낱말이 없어요. `tests/sdd_guidance.rs` 에서 옛 첫 문장 고정은 "A SPEC is for tier 2 changes" 로 바꿨어요.
- AC-4: met — `card_branch_ac4_` 가 통과했어요. `read.md` 가 한 가지 사실 물음은 바로 답하고 그 밖에는 갈래 나눔을 그대로 둬요.
- AC-5: met — `cargo test --locked --offline --test skill_structure` 가 통과했어요(`agent_delegation_text`, `core_rules_2876`, `next_flow_skill_scan`, `spec_card_number_observe`, `spec_branch_name_first` 와 한 번에 실행했고 exit 0). `cargo test --locked --offline --test product_cli skill_package` 는 고친 뒤 통과했어요(13개). 카드 영수증의 첫 실행은 `product_cli skill_package::routes_reference_existing_workflows` 에서 한 번 실패했어요. `plan.md` 의 Sort 단계가 패키지 경로를 이름 지었기 때문이에요. 지금은 "the audit contract (the `audit` reference)" 라고 쓰고 같은 커밋에서 매니페스트 해시를 새로 고쳤어요.
- Manifest: `skills/moai/references/source-manifest.json` 에서 `plan.md`, `read.md`, `sdd.md` 의 sha256 값 셋을 줄 임대 아래서 바꿨고, head 에서 49개 항목 중 49개가 파일 바이트와 같아요. `tests/fixtures/update_tux/*.json` 과 범위에 적은 테스트 파일 `agent_delegation_text.rs`, `skill_structure.rs` 는 바꿀 필요가 없었어요.
- REQ-10 간극과 Non-goals 의 긴장: `sdd.md` 6항에서 ", and repeat the digest in the pull request body" 절을 지웠어요. Non-goals 는 승인 줄을 승인 digest SPEC 의 몫으로 두는데, 그 SPEC 을 가진 레인이 절 제거를 거절했어요. 그래서 이 변경은 그 절만 지우고, 두 변경은 착지 때 그 한 줄에서 만나요(다른 SPEC 의 문장을 두되 그 절은 뺀 채로 두세요).
- not_run: `card/<slug>-<yyyymmdd>` 브랜치에서 SPEC 없는 카드를 끝에서 끝까지 실행하지 않았어요(도구 사전 점검 hook, `moai card submit` 과 그 트레일러 관문을 그런 브랜치에서 돌리지 않았어요). SPEC 없는 카드의 `moai card land` 도 실행하지 않았어요. 이 둘은 확인되지 않은 채로 남아요.
- not_run: 고치기 전에 product_cli 테스트를 RED 로 돌리지 않았어요.
