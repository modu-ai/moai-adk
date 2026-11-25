---
id: SPEC-20261122-next-flow
title: "The flow is named NeXT-Flow: the profile value next-flow, the old value m-flow still read and reported once, every shipped text says NeXT-Flow"
status: draft
revision: 2
created: 2026-11-22
parent: null
---

# NeXT-Flow: the flow's new name, the old value kept as an alias

## Intent

사용자 결정(2026-11-22, 재설계 마스터 카드의 결정 기록): 흐름 이름이 통합 브랜치의 새 이름 NeXT 를 따라 NeXT-Flow 로 바뀌어요(was M-Flow; 브랜치 개명은 아직 안 됐고 따로 진행해요). 이름의 뜻은 사용자의 말이에요: NeXT = New · engineering · eXperience · Together, 한 줄 소개 "Work in parallel. Land once. / 병렬로 개발하고, 한 번만 통합한다." 흐름이 작동하는 방식은 뜻과 갈라 설명해요: 레인은 head 와 base 만 제출하고 옮겨진 tip 을 다시 병합하지 않고, 착지 전에 독립 판정과 트리에 묶인 영수증이 오고, 착지는 한 번만 병합해 compare-and-swap 으로 가지를 옮기고 카드를 닫으며 다시 돌려도 두 번 옮기지 않고, 참조 가드가 다른 이동을 거부하며 모든 착지는 fast-forward 로 복사돼요. 프로필 값은 next-flow 가 되고, m-flow 라고 적힌 프로필은 같은 흐름으로 읽히며 옛 이름이라고 한 줄 알려요. 출하 글은 모두 NeXT-Flow 라고 쓰고 이름의 뜻은 한 곳에서만 설명해요.

### What exists (tip `b1db6ad006`, 잰 시각 2026-11-22T03:2xZ)

- 흐름 값의 자리: `WORK_FLOWS` 가 네 값 `m-flow`(맨 앞), `github-flow`, `trunk-based`, `git-flow` 를 정의해요(`src/profile.rs:184`), `work_flow` 키 문서는 `:158`. 새 프로필 틀은 `m-flow` 를 적어요(`src/profile.rs:1376`).
- 이름의 퍼짐: `m-flow`/`M-Flow` 는 src+skills+workflows 에서 모두 91곳(2026-11-22T03:2xZ 실측). 출하 글 중 `skills/moai/workflows/sync.md`, `skills/moai/workflows/project.md`, `skills/moai/references/adapters/herdr.md`, `skills/moai/references/internal/interview.md` 가 이름을 담고, `src/card/*`·`src/bundle.rs`·`src/statusline/*` 의 사용자 facing 글과 `docs/` 의 날짜가 찍힌 기록도 담아요.
- 값 문자열과 글자 이름의 구분: src 의 대부분은 값 `m-flow` 를 다루는 코드예요(읽기 경로는 그대로 남아요). 이름을 바꾸는 건 사용자가 읽는 글이고, 값 처리는 별칭으로 늘어나요.
- 텍스트 검사 시험의 틀: `tests/home_isolation_scan.rs`, `tests/mode_guidance_scan.rs` 같은 검사 시험이 있어서 같은 모양으로 흐름 이름 검사를 만들어요.

## Non-goals

- 이 조각의 코드·시험: SPEC 만 써요.
- 통합 브랜치 개명(아직 v2026.11.0; NeXT 개명은 다른 카드), 진입 스텝과 한 흐름 글 자리(step 6 의 자기 카드), 프로필 스키마 바꿈(값 하나만 늘어요).
- `docs/` 의 날짜가 찍힌 기록 문서 고치기: 역사 기록은 옛 이름을 그대로 담아요(REQ-3 의 예외).
- 네 글자 뜻의 재정의: 결정 기록의 문구 그대로예요.

## Requirements

- REQ-1: The work_flow key shall accept the value next-flow and select the same flow, flow text and checks the value m-flow selects · failure: none — a value that is neither next-flow nor a known earlier value is refused with the accepted list as today.
- REQ-2: A profile whose work_flow value is m-flow shall be read as the same flow and answered once with one line naming that value the old name of NeXT-Flow · failure: none — the line never blocks the read and never rewrites the profile.
- REQ-3: Every shipped text that names the flow shall say NeXT-Flow, with exactly one shipped text explaining the name's meaning in the user's words and the flow's behaviour · failure: none — dated documents under docs/ that record the old name as history are outside this rule.
- REQ-4: A repository text test shall refuse the string M-Flow in every shipped text except the one line that says it is the old name · failure: none — the refusal lists every file it names.
- REQ-5: Every write path that records a flow choice shall write the value next-flow for a new choice · failure: none — an existing profile's stored value is never rewritten by a reader.

## Acceptance

- AC-1 (REQ-1, REQ-2): a temporary project whose profile says next-flow and one that says m-flow select the same flow, and the old value is answered once with the old-name line · verify: `cargo test --locked --offline --test flow_rename_cli ac1_`
- AC-2 (REQ-5): a new flow choice records next-flow and an existing m-flow profile is left as it is · verify: `cargo test --locked --offline --test flow_rename_cli ac2_`
- AC-3 (REQ-3, REQ-4): the repository text test passes with exactly one explaining text and the one old-name line, and planting M-Flow in a shipped text fails it · verify: `cargo test --locked --offline --test flow_name_scan flow_`

## Scope

- this slice, write: `.moai/specs/SPEC-20261122-next-flow/**`
- later slices, write: `src/profile.rs` (the value list, the default templates), the user-facing strings of `src/card/*`, `src/bundle.rs`, `src/statusline/*`, the shipped texts under `skills/**` that name the flow, `tests/flow_rename_cli.rs`, `tests/flow_name_scan.rs`, the help goldens the wording feeds
- forbidden: rewriting stored profiles of real projects; the dated documents under `docs/`; the integration branch

## Plan

1. (done) SPEC (this slice).
2. Value: next-flow joins the accepted list, the reader maps the old value with the one old-name line, new choices record next-flow; the temp-project tests of AC-1 and AC-2.
3. Texts: the shipped texts say NeXT-Flow, one text explains the four letters, the flow-name scan test of AC-3; the help goldens the wording feeds are regenerated.

## Risks

- The sweep is wide (91 matches measured). The value string stays as an accepted alias in code, so the change is smaller than the raw count, but every user-facing string must be visited; the scan test keeps the sweep honest afterwards.
- New names trip existing greps: the help goldens and other word-scan tests may fail when texts change; the slice re-pins what it touches and runs the neighbouring checks.
- The one old-name line the scan test allows must stay exactly one; a second old-name mention in a shipped text fails the test by design.
- The letters' meanings are quoted from the decision record; rewording them in flight would drift from the user's decision.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | The old value m-flow is read forever as the same flow, with one old-name line. | The user decision: existing profiles keep working. | Warn louder later. |
| D-2 | The value spelling is next-flow; the display name is NeXT-Flow. | The decision record names the profile value and the branch; values are lowercase today. | Rename the value. |
| D-3 | Exactly one shipped text explains the name's meaning — the user's words (New · engineering · eXperience · Together with the tagline) — and the flow's behaviour; the technical letters are the behaviour description, never the meaning of the name. | The user's decision on the card replaces the earlier letter meanings. | Explain in every text. |
| D-4 | Dated documents under docs/ are exempt from the name rule. | They are history records of the time the old name was true. | Sweep them too. |
| D-5 | The flow-name check is a repository text test in the shape of the existing scan tests. | Criterion 2 asks for a repository text test; the shape is proven. | Move it into a CLI test. |
| D-6 | The flow-value lines inside files other cards reserve (the flow comparisons, the value list, the CI FLOWS row) change on this SPEC's own branch, by the leader's decision on the pair requests; no other line of those files. | The leader's decision note on the card; the leader re-checks the merged tree at landing. | Wait for each holder instead. |

## Result

코드 조각은 현재 tip(1637e28a7b)에서 다시 적용했어요. 11-22 에 올린 옛 조각(head de20a94966)은 감사도 착지도 받지 못했고 그 사이 `interview.md`·`project.md`·`next-flow.md` 가 다시 쓰여서, 옛 diff 는 참고로만 쓰고 충돌한 글은 tip 글을 기준으로 새로 고쳤어요.

- 값: `WORK_FLOWS` 에 `next-flow` 가 맨 앞이고 `m-flow` 는 계속 받아요. `is_local_flow` 가 두 값을 같은 흐름으로 보고, `old_flow_notice` 가 옛 값을 한 줄로 알려요(stderr, 프로필은 다시 쓰지 않아요). 비교 다섯 곳(`landed.rs`, `dispatch.rs`, `work.rs`, `ops/bundle.rs`, CI 표의 `FlowRow`)이 두 값을 인정해요.
- 이름: 사람이 읽는 글은 `flow_display_name` 으로 NeXT-Flow 를 보여 줘요(생성되는 CLAUDE.md 의 작업 흐름 줄, 세션 시작 글). 카드 오류 문장과 도움말 글, `/moai project` 질문 은행의 `work_flow` 선택지도 NeXT-Flow 라고 말하고, 선택지에 "변경마다 PR 없음, 다른 에이전트가 검사한 뒤 로컬 통합 브랜치에 착지" 한 줄이 붙어요. 이름의 뜻을 설명하는 출하 글은 `flows/next-flow` 하나뿐이에요(tip 에서 이미 그래요).
- 이번에 하지 않은 것: `tools/self_check.sh` 의 프로필 가지 읽기(다른 카드가 예약 중이고 이 문제와 상관이 없어요), herdr 글·테스트(tip 에서 없어졌어요), 묶음 PR 길의 이름 `m-flow/bundle-…`(별도로 정해야 할 일이라 건드리지 않았어요), 질문지의 다른 결함(다른 카드).
- 승인 상태: 이 revision 2 의 digest 승인은 제가 확인하지 못했어요. 리더가 `moai spec show` 로 대조해 주세요.

검증(임시 HOME·임시 프로젝트에서만, 진짜 `~/.claude` 는 읽지도 쓰지도 않았어요):
- AC-1: met — `cargo test --locked --offline --test flow_rename_cli ac1_` 통과(next-flow 와 m-flow 프로필이 같은 흐름을 고르고 옛 값은 한 줄로 한 번만 알려요).
- AC-2: met — `cargo test --locked --offline --test flow_rename_cli ac2_` 통과(새 선택은 next-flow 로 기록되고 기존 m-flow 프로필은 그대로예요).
- AC-3: met — `cargo test --locked --offline --test flow_name_scan flow_` 통과(질문 은행이 `next-flow (NeXT-Flow)` 를 보여 주는지와 표시 이름 함수를 고정한 새 테스트 둘 포함).
- 이웃 테스트: `generic_flow_questions` 10, `claude_md_layout_cli` 16, `agent_defs_package`, `next_flow_skill_scan`, `cli_text_citations`, `source_manifest_facts`, `card` 묶음의 `card_done_mflow_cli` 25(next-flow 짝 1개 추가), `hook_dispatch_cli` 53(세션 시작 줄의 표시 이름을 다시 고정하고 next-flow 짝 1개 추가) 통과. 전체 영수증의 첫 테스트 실행에서 `hook_dispatch_cli` 3개만 실패했고(옛 글 고정) 위처럼 고쳤어요. 옛 글을 고정한 단언은 새 글로 다시 고정했어요.
- 지킴이 뮤턴트(각각 복원, 트리 깨끗함 확인): 수용 목록에서 `next-flow` 제거 → `flow_rename_cli` ac1·ac2 실패; `work.rs` 비교를 `== "m-flow"` 로 → `next_flow_done_closes…` 실패; CLAUDE.md 줄의 표시 이름 제거 → `claude_md_layout_cli` 2개 실패; 질문 은행 행을 옛 값으로 → `flow_name_the_interview_offers_next_flow_by_name` 실패. 모두 exit 100. `ops/bundle.rs`·`hook_engine.rs` 는 뮤턴트를 만들지 않았어요(not_run).
