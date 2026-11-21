---
id: SPEC-20261121-gate-check-unread-inputs
title: "gate-check names what it read — the mo.ai.kr G1 was a verdict line in the wrong shape, not a read failure"
status: draft
revision: 1
created: 2026-11-21
---

## Intent

mo.ai.kr의 리더가 2026-11-21T00:2xZ 에 보고했어요: `moai gate-check --pr <N> --repo modu-ai/mo.ai.kr --dir . --tip refs/remotes/origin/main` 이 모든 PR 에서 `G1 no AUDIT verdict`·`ci=none`·`ignored_control_lines.reason no_member_authority (count 0)` — 병합된 PR(#1029·#1031)조차, 그날 이른 시간엔 G10 까지 갔던 PR 조차. 그 사이 그 저장소에서 바뀐 것은 Go 세대 마커 두 개가 `.moai/backups/` 로 이동해 state_root 가 mixed→rust 로 넘어간 일이었어요.

조사 결과(이 카드의 첫 절반, 실측 증거):

- **gate 체인은 state 를 읽지 않아요.** `src/gate.rs`·`src/gate_collect.rs`·`src/comment_authority.rs`·`src/forge/transport.rs` 전수에서 `state_root`·프로필·`config/sections` 참조는 0 건 — GitHub 읽기는 `gh`(gh 자체 인증, 토큰은 호출자의 `GH_TOKEN`) 한 번의 스폰이고 git 은 `--dir` 의 로컬 git 이라, state_root 전환은 이 결과에 닿을 길이 없어요.
- **원인은 판정 줄의 형식 불일치예요.** gate 가 읽는 verdict 줄은 `AUDIT #<pr> (<title>) head <40hex> base <40hex> — verdict: PASS|FAIL`(`moai audit-verdict-line` 이 출력, `gate::format_verdict_line`)인데, mo.ai.kr 의 MEMBER 코멘트는 리더 손기록 요약 `AUDIT PASS <sha> ci-run=<n> conclusion=success` — `verdict_line` 은 첫 줄이 `AUDIT #` 으로 시작하는 코멘트만 verdict 로 읽으므로 두 PR 모두 verdict 0 건 → G1. 실측: `gh api repos/modu-ai/mo.ai.kr/issues/{1029,1031}/comments` 로 읽은 MEMBER 코멘트 4 건 중 정식 줄 0 건, 비정식 줄 2 건(`AUDIT PASS …`), G1 사유는 그 차이를 말하지 않았어요. 익명 읽기 가설은 반증했어요 — unauthenticated `curl` 도 `author_association: MEMBER` 를 돌려줘요.
- `ci=none` 은 정보일 뿐(규칙은 안 읽어요) — watched check(`ci`/`rc-ci`/…)이 0 건이라는 뜻이고, 어떤 check-run 이름이 읽혔는지 출력에 없어 형식 문제와 읽기 실패가 구분되지 않았어요.

수리는 두 번째 절반: gate-check 가 무엇을 읽었는지 이름으로 말해요 — G1 의 사유 문장이 읽은 코멘트 수·member 수·형식이 어긋난 member 첫 줄(실제 사례 그대로)과 정식 줄 모양을 이름하고, 응답에 `reads` 요약(코멘트 수·association 분포·verdict 줄 수·lookalike 목록·check-run 수·이름들)이 붙어요. 이러면 다음에 비슷한 보고가 오면 한 줄로 원인이 밝혀져요.

## Non-goals

- mo.ai.kr 의 판정 코멘트 재작성: 그쪽 몫 — 이 카드의 출력이 정식 형식(`moai audit-verdict-line`)과 어긋난 줄을 이름해 주므로 재작성은 한 번의 재기록이에요.
- state_root mixed→rust 전환 자체와 Go 마커 백업 이동: 이 카드의 조사로 gate 결과와 무관함이 확인된 것으로 닫아요.
- verdict 줄 형식의 확장(`AUDIT PASS …` 도 읽게): 감사 줄은 한 형식 — 느슨하게 읽으면 위조 줄이 verdict 가 돼요.
- `ci_info` 문구 확장: 정보성 유지 — 대신 `reads` 요약이 check-run 이름들을 이름해요.

## Requirements

- REQ-1: While `moai gate-check` stops on G1 because no comment carries a verdict line, its reason shall name what was read — the number of member comments it holds (or `no member comments were read at all (0)` when there are none), and, when a member comment's first line starts with `AUDIT` but not with `AUDIT #`, that line's first 60 characters together with the verdict line's real shape (`AUDIT #<pr> (<title>) head <40hex> base <40hex> — verdict: PASS|FAIL`, printed by `moai audit-verdict-line`). · failure: none — pinned by AC-1.
- REQ-2: While `moai gate-check` answers (pass or stop, live or from a snapshot), the answer shall carry a `reads` summary — the comment count, the `author_association` distribution, the verdict-line count, the count and the list of member first lines that start with `AUDIT` but are not verdict lines (comment id, association, the first 60 characters), and the check-run count with the distinct check names — so an empty read and a read failure are distinguishable in the answer itself. · failure: none — pinned by AC-2.

## Acceptance

- AC-1 (REQ-1): a snapshot whose only member comment starts `AUDIT PASS …` stops on G1 with a reason naming `1 member comment(s) read`, the `AUDIT PASS 72225828…` line, and the `AUDIT #` shape; an empty comments list names `no member comments were read at all (0)`.
  verify: `cargo test --locked --offline --test gate_check_cli`
- AC-2 (REQ-2): the same runs carry `reads` with `comments`, `author_associations`, `verdict_lines`, `member_first_lines_that_are_not_verdict_lines`, `lookalike_first_lines`, `check_runs` and `check_run_names` — the wrong-shape case reports `verdict_lines: 0` and `check_run_names: ["build"]` with `ci_info: ci=none`.
  verify: `cargo test --locked --offline --test gate_check_cli`
- AC-3: the neighbours of the gate keep passing — the whole binary tests (the help rules read every command) and the live-transport collection tests.
  verify: `cargo test --offline --locked --bins` and `cargo test --offline --locked --lib gate`

## Scope

- write: this SPEC folder, `src/gate.rs` (the G1 reason and the `reads` summary), `src/ops/gate_check.rs` (the answer carries `reads`), `tests/gate_check_cli.rs`
- read-only: `src/gate_collect.rs`, `src/comment_authority.rs`, `src/forge/transport.rs` (the survey that closed the state_root question)
- forbidden: 실제 HOME·네트워크 쓰기(POST/merge), verdict 줄 형식의 느슨화, mo.ai.kr 의 코멘트를 이 카드가 고치는 것

## Plan

1. RED: gate_check_cli 에 두 장면 — 비정식 member 줄(실측 문구 그대로)과 빈 comments — 새 사유·`reads` 를 기대로, 지금 코드로 깨져요.
2. GREEN: `gate::reads_summary` 와 `no_verdict_reason` — 사유 문장은 결정적 힌트만(코멘트 수·member 수·가장 가까운 어긋난 줄 60 자·정식 형식), `reads` 는 ops 의 `decide` 가 pass·stop 양쪽에 붙여요.
3. 회귀: 전체 `--bins`(help 규칙이 모든 명령을 읽어요)·`--lib gate`·ci_info 의 전체 비교 시험은 `reads` 를 예외로(그 시험은 check-run 목록을 바꿔요) — clippy·fmt → tier 2 감사 → READY.

## Risks

- reason 문장이 길어요: 결정적 힌트(수·어긋난 줄·정식 형식)만 남기고 분포 같은 전체는 `reads` 로 — 기존 사유의 한 줄 문화 유지.
- lookalike 탐지(`AUDIT` 시작·`AUDIT #` 아님)는 거짓양성이 가능해요: 정보성 필드라 규칙은 안 읽고, 이름만 이름해요 — verdict 판정은 늘 `verdict_line` 하나예요.
- stdin 스냅샷은 member 만 남겨요(`keep_member_comments`): `reads.comments` 는 결정에 쓰인 코멘트 수로, 무시된 비멤버는 이미 `ignored_control_lines` 가 이름해요 — 두 요약이 겹치지 않아요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D1 | state_root 전환은 무관으로 닫는다 — gate 체인의 state 참조 0 건 전수 조사 | 조사는 실측(소스 전수 grep)과 PR 실측(코멘트 읽기 성공·형식 불일치)으로 끝났어요 | mo.ai.kr 의 다음 실행이 `reads` 로 다른 입력 실패를 가리키면 다시 열어요 |
| D2 | verdict 줄 형식은 늘리지 않는다 — 어긋난 줄을 이름해 안내한다 | 줄 형식은 위조 방지의 하나의 검문소(첫 줄 `AUDIT #`·형식 재독해 검증) — 느슨화는 위조 줄을 verdict 로 만들어요 | 사용자가 형식 변경을 요청하면 별도 카드로 |
| D3 | `reads` 는 정보성 — 어떤 규칙도 읽지 않는다 | `ci_info` 와 같은 자리 · 규칙 로직 불변(기존 48 시험 중 수정은 전체 비교 1 곳의 예외뿐) | — |

## Result

- AC-1: met (REQ-1) — 비정식 member 줄 장면: G1 사유가 `1 member comment(s) read`·어긋난 줄(`AUDIT PASS 72225828…`)·정식 형식(`AUDIT #`)을 이름해요; 빈 코멘트 장면: `no member comments were read at all (0)`. `cargo test --offline --locked --test gate_check_cli` — g1_names_what_it_read… · g1_names_an_empty_comment_list… 통과(48+3 passed — 수리 시험 3 건 더해요).
- AC-2: met (REQ-2) — 같은 실행의 답이 `reads` 를 담아요: comments·author_associations·verdict_lines·member_first_lines_that_are_not_verdict_lines·lookalike_first_lines·check_runs·check_run_names — 잘못된 형태 장면은 `verdict_lines: 0`·`check_run_names: ["build"]`·`ci_info: ci=none`.
- AC-3: met — `cargo test --offline --locked --bins` 통과(help 규칙이 모든 명령을 읽어요)·`cargo test --offline --locked --lib gate` 32 passed.
- 감사 수리(같은 감사자 audit-c258 델타): B1 — 사유·`reads.lookalike_first_lines[].first_line` 은 `redact_secret_shapes` 를 먼저 적용하고 그다음 60 자로 잘라요(자르기가 비밀 문자 자체를 보지 않아요 — 마스킹·경계 시험 고정). N1 — member 판정은 `comment_authority::is_member`(OWNER·MEMBER·COLLABORATOR 모두). N3 — g1 위의 중복 doc 주석 제거. not run: 없어요(이 카드의 나머지 조각 없음 — 남은 것은 mo.ai.kr 쪽 판정 코멘트 재기록이에요). closure_status: draft — 승인은 병합 시 사용자 선택 뒤예요.
