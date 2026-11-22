---
id: SPEC-20261122-next-flow
title: "The flow is named NeXT-Flow: the profile value next-flow, the old value m-flow still read and reported once, every shipped text says NeXT-Flow"
status: draft
revision: 1
created: 2026-11-22
parent: null
---

# NeXT-Flow: the flow's new name, the old value kept as an alias

## Intent

사용자 결정(2026-11-22, 재설계 마스터 카드의 결정 기록): 흐름 이름이 통합 브랜치 NeXT 를 따라 NeXT-Flow 로 바뀌어요(_was M-Flow_). 글자의 뜻: N = No tip chase(레인은 head 와 base 만 제출하고 옮겨진 tip 을 다시 병합하지 않아요), e = evidence before landing(독립 판정과 트리에 묶인 영수증이 착지 전에 와요), X = eXactly-once landing(착지 명령은 한 번만 병합하고 compare-and-swap 으로 가지를 옮기고 카드를 닫아요; 다시 돌려도 두 번 옮기지 않아요), T = Tip guarded and copied(참조 가드가 다른 이동을 거부하고 모든 착지는 fast-forward 로 원격에 올라가요). 프로필 값은 next-flow 가 되고, m-flow 라고 적힌 프로필은 같은 흐름으로 읽히며 옛 이름이라고 한 줄 알려요. 출하 글은 모두 NeXT-Flow 라고 쓰고 네 글자의 뜻은 한 곳에서만 설명해요. 이 조각은 SPEC 만 쓰고 코드는 없어요.

### What exists (tip `b1db6ad006`, 잰 시각 2026-11-22T03:2xZ)

- 흐름 값의 자리: `WORK_FLOWS` 가 네 값 `m-flow`(맨 앞), `github-flow`, `trunk-based`, `git-flow` 를 정의해요(`src/profile.rs:184`), `work_flow` 키 문서는 `:158`. 새 프로필 틀은 `m-flow` 를 적어요(`src/profile.rs:1376`).
- 이름의 퍼짐: `m-flow`/`M-Flow` 는 src+skills+workflows 에서 모두 91곳(2026-11-22T03:2xZ 실측). 출하 글 중 `skills/moai/workflows/sync.md`, `skills/moai/workflows/project.md`, `skills/moai/references/adapters/herdr.md`, `skills/moai/references/internal/interview.md` 가 이름을 담고, `src/card/*`·`src/bundle.rs`·`src/statusline/*` 의 사용자 facing 글과 `docs/` 의 날짜가 찍힌 기록도 담아요.
- 값 문자열과 글자 이름의 구분: src 의 대부분은 값 `m-flow` 를 다루는 코드예요(읽기 경로는 그대로 남아요). 이름을 바꾸는 건 사용자가 읽는 글이고, 값 처리는 별칭으로 늘어나요.
- 텍스트 검사 시험의 틀: `tests/home_isolation_scan.rs`, `tests/mode_guidance_scan.rs` 같은 검사 시험이 있어서 같은 모양으로 흐름 이름 검사를 만들어요.

## Non-goals

- 이 조각의 코드·시험: SPEC 만 쓰어요.
- 통합 브랜치 개명(이미 NeXT), 진입 스텝과 한 흐름 글 자리(step 6 의 자기 카드), 프로필 스키마 바꿈(값 하나만 늘어요).
- `docs/` 의 날짜가 찍힌 기록 문서 고치기: 역사 기록은 옛 이름을 그대로 담아요(REQ-3 의 예외).
- 네 글자 뜻의 재정의: 결정 기록의 문구 그대로예요.

## Requirements

- REQ-1: The work_flow key shall accept the value next-flow and select the same flow, flow text and checks the value m-flow selects · failure: none — a value that is neither next-flow nor a known earlier value is refused with the accepted list as today.
- REQ-2: A profile whose work_flow value is m-flow shall be read as the same flow and answered once with one line naming that value the old name of NeXT-Flow · failure: none — the line never blocks the read and never rewrites the profile.
- REQ-3: Every shipped text that names the flow shall say NeXT-Flow, with exactly one shipped text explaining the four letters in their decided meanings · failure: none — dated documents under docs/ that record the old name as history are outside this rule.
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
| D-3 | Exactly one shipped text explains the four letters. | The card body: "explain the four letters once". | Explain in every text. |
| D-4 | Dated documents under docs/ are exempt from the name rule. | They are history records of the time the old name was true. | Sweep them too. |
| D-5 | The flow-name check is a repository text test in the shape of the existing scan tests. | Criterion 2 asks for a repository text test; the shape is proven. | Move it into a CLI test. |

## Result

(Filled when the work is done. This is the SPEC-only slice: no code, no acceptance criterion run yet; the value, the alias rule and the letters are recorded above from the user's decision.)
