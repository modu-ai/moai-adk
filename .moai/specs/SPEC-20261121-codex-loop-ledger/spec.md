---
id: SPEC-20261121-codex-loop-ledger
title: "The Codex loop tools read the card ledger — the lane Stop hook decides from cards, not GitHub"
status: draft
revision: 1
created: 2026-11-21
---

## Intent

mo.ai.kr 의 모든 레인이 Codex 에서 무중단으로 못 도는 원인 조사(보고서 codex-leader-nonstop-2026-11-21.md, 02:59-03:07Z 실측)의 사용자 지시('원인을 찾아서 보고하고 수정해서 빌드해서 제공')의 코드 절반. 조사의 코드 관련 결론: Codex CLI 에는 타이머가 없고 유일한 이어 달리기 장치는 Stop 훅(`decision: block` 을 돌려주는 것)인데, moai-adk 가 지금 설치하는 Codex 레인 Stop 훅(`lane-stop-hook`, #2904)이 그 판정을 **GitHub Issue/claim 읽기**(`lane_loop::read_github`)로 해요 — 토큰이 없거나 읽기가 느린 환경에서 훅을 깔면 레인이 멈출 수 있어요.

이 SPEC 은 그 도구들의 판정 소스를 **카드 원장**으로 바꿔요. 조각 1(이 커밋): `lane-stop-hook` 의 판정을 `card::work::next` — `moai card next` 가 프린트하는 바로 그 뷰, GitHub 요청 0 건 — 로 교체해요: 세션이 hold 한 카드는 이어 하라(lease 가 살아 있으면 continue, 끝났으면 재확보), 없으면 `ready` 카드를 우선순위로 제안하고, `hold` 는 절대 제안하지 않으며, 제안할 것이 없으면 아무것도 프린트하지 않아 턴이 끝나도록. 원장을 못 읽으면 경고 한 줄과 함께 무출력 — 훅 실패가 레인을 멈추지 않아요(fail-open). 설치된 명령줄(`--lane L --repo O/R --snapshot …`)은 그 모양 그대로 받고 판정은 읽지 않아요 — 훅 템플릿(hooks.rs)은 바뀌지 않아요.

나머지 도구들은 이어지는 조각: `lane-next` 의 카드 원장 모드(HOLD·전제조건·structured state·scope 가용성·끝난 lease 를 C-266 규칙대로 존중 — `src/lane_next.rs`), `leader-round` 의 원장 요약(inbox·held/expired leases·train — `src/leader_round.rs`), goal 문장의 원장화(`src/ops/goal_*.rs` 의 tools/gen_goal_block.py 이식 잔여).

## Non-goals

- gate 의 판정 느슨화: G1-G12 는 그대로 — 이 카드는 Stop 훅이 **무엇을 읽는지**만 바꿔요.
- GitHub 운영의 lane-next 행동 교체: 원장 모드는 별도 조각에서 계약을 정하고 기존 GitHub 읽기는 유지해요.
- Orca 자동화·세션 배치 같은 mo.ai.kr 환경 부분: 보고서의 (a)-(d) 중 코드 밖이에요.
- `hooks.rs` 의 훅 템플릿 변경: 설치된 명령줄 모양을 유지해요(REQ-3).

## Requirements

- REQ-1: While `moai lane-stop-hook` runs, it shall decide from the project's card ledger through `card::work::next` alone — printing the block JSON with a one-line reason naming the card when the session holds a card (continue it; or re-claim it when the lease ended) or a `ready` card exists (take it, by priority), holding a `hold` card out of every offer, and printing nothing when no card can be worked on — while it makes no GitHub request at all. · failure: none — pinned by AC-1.
- REQ-2: While the ledger cannot be read, the hook shall print a warning line on stderr and nothing on stdout, so the turn ends instead of hanging and a lane never stops on a hook error. · failure: none — pinned by AC-2.
- REQ-3: While the installed command line names `--lane`, `--repo` or `--snapshot`, the hook shall accept them and not read them for the judgement — the deciding session is the hook's own environment (`CLAUDE_CODE_SESSION_ID` or `CODEX_THREAD_ID`) and the project is the working folder, unless `--project` names one. · failure: none — pinned by AC-1.

## Acceptance

- AC-1 (REQ-1): a lane whose session holds a card is blocked with `continue C-1 (active) …`; a lane with a ready card and nothing held is blocked with `take C-1 (p1): <title>`; a lane whose lease ended is blocked with `re-claim C-1 …`; a `hold` card and a project without a ledger print nothing (allow).
  verify: `cargo test --offline --locked --test lane_stop_hook_cli`
- AC-2 (REQ-2): with the ledger unreadable the hook prints the warning on stderr, nothing on stdout, and exits 0.
  verify: `cargo test --offline --locked --test lane_stop_hook_cli` (the ledger-absent scene) and the refusal path in `lane_stop_hook_main`.
- AC-3: the neighbours keep passing — the whole binary tests (the help rules read every command) and the library.
  verify: `cargo test --offline --locked --bins --lib`

## Scope

- write (slice 1): this SPEC folder, `src/main.rs` (`lane_stop_hook_main`'s judgement and the ledger answer), `tests/lane_stop_hook_cli.rs`
- later slices widen: `src/lane_next.rs`, `src/leader_round.rs`, `src/ops/goal_block.rs`, `src/hooks.rs` (the goal text's ledger source) — each taken when the slice starts
- read-only: `src/card/work.rs`, `src/ledger/`, `src/comment_authority.rs`
- forbidden: GitHub 읽기의 재도입(판정 경로), 훅 실패의 fail-closed, 실제 HOME 쓰기, 카드 원장에 쓰기(읽기만)

## Plan

1. Slice 1 (this commit): `lane_stop_hook_main` reads the ledger — RED first (`tests/lane_stop_hook_cli.rs`, five scenes: held→continue, ready→take, ended lease→re-claim, hold→nothing, no ledger→nothing), then the judgement over `card::work::next`.
2. Slice 2: `lane-next`'s ledger mode (HOLD, prerequisites, structured state, scope availability, ended leases per C-266) — taken with its tests.
3. Slice 3: `leader-round` summarises from the ledger (inbox, held/expired leases, train).
4. Slice 4: the goal text's ledger source (`ops/goal_*.rs`).
5. Neighbours: the whole `--bins --lib` after each slice; the Codex hooks documentation line in `interfaces.md` stays as it is (the hook command's shape is unchanged).

## Risks

- Stop 훅이 block 을 남발하면 세션이 끝나지 않아요: 제안은 원장이 정한 하나(다음 카드)뿐이고, 제안할 것이 없으면 무조건 allow — 무한 block 의 입력은 원장에 `ready` 카드가 있는 것 자체라 의도된 동작이에요.
- `card::work::next` 는 세션 env 에 의존해요: 훅은 세션 안에서 돌므로 같은 세션이지만, env 가 없는 실행(수동)은 holder 없이 판정해 `mine` 이 비어 ready 제안만 나와요 — 계약에 명시해요.
- 원장이 커지면 `next` 가 느려질 수 있어요: 원장은 단일 SQLite 파일 읽기라 GitHub 폴링보다 훨씬 가벼워요 — 훅 timeout(120s) 안에서 실측해요(AC-1 의 시험은 임시 원장).

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D1 | 판정 소스는 `card::work::next` — `moai card next` 의 뷰 재사용 | hold·ready·skipped(사유)·next 가 이미 원장 규칙대로(HOLD 미포함·scope 비충돌·p 순) 정리돼 있어요 — 훅은 그 뷰를 decision 문장으로 바꾸는 얇은 층 | 뷰가 부족해지면 별도 판정 함수로 |
| D2 | allow 는 무출력 — 기존 훅 계약 유지 | `lane_stop_hook_main` 은 빈 answer 를 프린트하지 않아요(기존 동작) — 훅 템플릿·문서 무수정 | — |
| D3 | 실패는 fail-open(경고 + 무출력) | 보고 (4): 훅을 깔면 레인이 멈출 수 있어요 — 원장 읽기 실패가 레인을 멈추는 것은 이 카드의 목적 밖 | 사용자가 fail-closed 를 고르면 플래그로 |
| D4 | `lane`·`repo`·`snapshot` 인자는 받되 판정은 안 읽어요 | 설치된 훅 명령줄(hooks-plan 템플릿)과 호환 — 템플릿 무수정 | 템플릿을 함께 바꾸는 별도 카드 |
