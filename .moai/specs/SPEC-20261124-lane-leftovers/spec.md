---
id: SPEC-20261124-lane-leftovers
title: "lanes keep their throwaway copies and temporary HOMEs inside the worktree and report what is left in READY"
status: draft
revision: 4
created: 2026-11-24
---

# lanes keep their throwaway copies and temporary HOMEs inside the worktree and report what is left in READY

## Intent

재발 방지예요. 같은 라운드(2026-11-24)의 감사 노트 N6 두 개가 같은 일을 짚었어요: 한 레인은 공용 체크아웃에 0 바이트 로그 파일 하나와 시스템 임시 폴더의 `moai-danger.*` 폴더 둘을 남겼고, 다른 레인은 뮤턴트 복사본 약 800MB 를 자기 세션의 scratchpad 에 남겼어요. 이 숫자와 경로는 카드 본문과 감사 노트가 말한 것이고, 이 SPEC 을 쓰며 다시 재지는 않았어요.

원인은 두 가지예요. 레인은 절대 경로 삭제가 금지라(`CLAUDE.md`, `.claude/rules/moai/flow.md`) 저장소와 워크트리 밖에 만든 것을 스스로 치우지 못해요. 그리고 버리는 복사본·임시 HOME·`mktemp` 폴더를 어디에 만들지 정한 규칙이 없어요. 지금 `docs/lane-rules.md` 는 노트와 로그를 `target/<name>/` 에 두라고만 해요(Worktree entry and exit 의 마지막 줄).

이 SPEC 은 세 가지를 해요. (1) 그런 것은 `<worktree>/target/<name>/` 아래에만 만들고 READY 전에 상대 경로로 지운다는 한 문장을 레인 지침에 적어요: 워크트리 안은 레인이 상대 경로로 지울 수 있는 곳이에요. (2) READY 노트에 `leftovers:` 줄을 요구해서 지우지 못한 것이 노트에 적히게 하고, 줄이 없으면 `moai card note` 가 거절하고 길을 알려 줘요. (3) READY 를 쓰는 순간 공용 체크아웃에 레인이 만든 새 파일이 있는지 읽어서 답에 드러내요.

읽어서 확인한 것 하나가 설계를 정했어요: 공용 체크아웃의 `.gitignore` 는 `/.moai/*` 를 무시해요(`git check-ignore -v .moai/none.log` → `.gitignore:361:/.moai/*`, 이 실행에서 읽음). 그래서 카드가 말한 "공용 체크아웃 `git status` 비교" 는 `.moai/none.log` 같은 파일을 보지 못해요. 무시된 파일까지 읽어야 해요(REQ-4).

읽는 사람: 레인, 리더, 이 변경의 감사자.

## Interview

- Q1: 같은 문제를 만든 두 일은 어디서 왔는가? -> A: 카드 본문과 감사 노트 N6 두 개예요. 둘의 원인은 같아서(위치 규칙 없음 + 절대 경로 삭제 금지) 한 SPEC 으로 닫아요.
- Q2: 새 가드를 만들기 전에 두 곳 이상의 공개 구현을 비교해야 하는가(코어 규칙 19)? -> A: 아니요. 이 저장소에 같은 모양의 가드가 이미 둘 있어요: READY 노트의 `guard-mutant:` 줄(`src/bundle/mutant_line.rs`)과 `card note` 의 READY 영수증 검사(`src/card/ready_receipt.rs`). 새 줄은 그 둘과 같은 모양으로 읽고 거절해요.
- Q3: `<worktree>/target/<name>/` 의 것은 정리 때 같이 사라지는가? -> A: 저절로는 아니에요(카드 본문의 가정과 달라요). `docs/lane-rules.md` 가 말하듯 정리는 cargo 출력만 지우고 `target/<name>/` 의 노트가 있는 워크트리는 `worktree-gc-plan` 의 `remove` 후보가 아니에요. 그래서 레인이 READY 전에 자기 것을 상대 경로로 지우고(워크트리 안이라 허용), 못 지운 것만 `leftovers:` 에 적어요(REQ-1).
- Q4: 시스템 임시 폴더나 scratchpad 의 것을 잡는 검사를 더하는가? -> A: 아니요(Non-goals). git 이 볼 수 없는 곳이라 위치 규칙과 READY 줄이 보고의 전부예요. 이 한계는 Risks 에 적었어요.
- Q5: 새 거절 규칙은 observe 부터인가(SDD 규칙 11)? -> A: 부분만이에요. `leftovers:` 줄의 있고 없음은 읽는 일뿐이라 오탐이 없어서 카드대로 바로 거절해요. 공용 체크아웃 목록은 다른 레인의 파일이나 `.DS_Store` 같은 것에 오탐할 수 있어서 observe 로 시작해요. 이 선택은 D-1 로 리더에게 물어요.

## Non-goals

- 시스템 임시 폴더(`$TMPDIR`)와 scratchpad 를 훑거나 지우는 검사는 더하지 않아요. 이미 있는 찌꺼기를 지우는 일도 이 SPEC 이 아니에요(운영자의 몫이에요).
- 레인에게 절대 경로 삭제를 허용하지 않아요. `rm`/삭제 가드(`SPEC-20261122-*` 의 보호 훅)와 "워크트리를 지우지 않는다" 는 규칙은 그대로예요.
- 공용 체크아웃의 추적되는 파일의 변경은 보지 않아요: 카드가 말한 것은 "추적되지 않은 새 파일" 이에요.
- `CLAUDE.md` 는 바꾸지 않아요(카드 범위 밖). 지침은 `docs/lane-rules.md` 와 레인에게 가는 스킬 참조 문서에만 적어요.
- `guard-mutant:` 줄의 규칙과 `moai bundle check` 는 바꾸지 않아요.
- 영수증 검사(`check_note` 의 PASS 영수증 규칙)는 느슨하게 하지 않아요. 새 검사는 그 앞에 더해질 뿐이에요.

## Requirements

- REQ-1: The lane guidance (`docs/lane-rules.md` and the `self-check` reference of the moai skill) shall say that a throwaway copy of the repository, a temporary HOME and a `mktemp` folder are made only under `<worktree>/target/<name>/` (with `TMPDIR` pointed there for a run that makes its own temporary folders) and are removed by a relative path before READY.
- REQ-2: When `moai card note` gets a hand-back note (a first line that starts with `LANE-READY-INTEGRATION`, `REPAIRED` or `READY `) whose first line reads correctly, it shall keep the note only when the body holds exactly one readable `leftovers:` line · failure: see REQ-3.
- REQ-3 (REQ-2): IF the body holds no `leftovers:` line, more than one, an empty one, a template placeholder left as it is (`<path>`) or `none` together with a path, THEN `moai card note` shall refuse the note with the line to add and write nothing.
- REQ-4: When `moai card note` keeps a hand-back note, it shall add to its answer `shared_checkout_new`, the paths in the primary checkout that are untracked or ignored, are not on the fixed list of expected state (`target/`, `.claude/`, `CLAUDE.local.md`, `.mcp.json`, `.DS_Store`) and were created or changed after the card's reservation began · failure: see REQ-5.
- REQ-5 (REQ-4): IF the primary checkout cannot be read or the card holds no reservation, THEN `moai card note` shall keep the note and answer `shared_checkout: not_checked` with the reason.
- REQ-6: When `moai card note` reads the primary checkout, it shall change nothing there, not even the index's stat cache · failure: none — a read that changes the checkout is a defect that the before/after comparison of AC-3 shows.
- REQ-7: The `leftovers:` line shall appear in the Hand-off section of `docs/lane-rules.md`, on the operation page of `card`, in the `card note` help text and in the refusal text of REQ-3 as a line the lane can paste.

## Acceptance

- AC-1 (REQ-1): the two guidance files name `<worktree>/target/<name>/`, `TMPDIR` and removal by a relative path before READY · verify: `cargo nextest run --locked --offline --test card_ready_leftovers_cli the_lane_guidance_names_the_throwaway_place`
- AC-2 (REQ-2, REQ-3): in a temporary project, a READY note without the line is refused for each of the three markers, the refusal names the line to add, and the ledger holds no new note; each bad shape of REQ-3 is refused; `leftovers: none` and `leftovers: <path> | <path>` are kept with the PASS receipt present; a line inside a code fence or indented four spaces does not count · verify: `cargo nextest run --locked --offline --test card_ready_leftovers_cli a_ready_note_without_a_readable_leftovers_line_is_refused`
- AC-3 (REQ-4, REQ-5, REQ-6): with a temporary primary checkout and a linked worktree, a new untracked file and a new ignored file made in the primary after the reservation are in `shared_checkout_new`; a file older than the reservation, an entry on the fixed list and a file inside the lane's own worktree are not; the note is kept; the primary's `.git/index` bytes and modification time and its `git status` output are the same before and after; with the repository unreadable the answer has `shared_checkout: not_checked` and the note is kept · verify: `cargo nextest run --locked --offline --test card_ready_leftovers_cli the_shared_checkout_is_read_without_changing_it`
- AC-4 (REQ-2, REQ-3): the pure reader of the line (`none`, paths, placeholder, repeated, fenced, indented, `- ` prefix) is pinned by unit tests · verify: `cargo nextest run --locked --offline --lib card::ready_leftovers::tests`
- AC-5 (REQ-7): the Hand-off section, the `card` operation page, the `card note` help text and the refusal text each show `leftovers:` · verify: `cargo nextest run --locked --offline --test card_ready_leftovers_cli the_leftovers_line_is_shown_where_a_lane_looks`

## Scope

- write: `.moai/specs/SPEC-20261124-lane-leftovers/**`, `src/card/ready_leftovers.rs`(새), `src/card/mod.rs`(한 줄 `pub mod`), `src/card/work.rs`(호출 한 줄과 답 필드), `src/ops/card.rs`(`Note` 도움말 글만), `tests/card_ready_leftovers_cli.rs`(새), `docs/lane-rules.md`, `docs/design-review/appendix/operations/card.md`, `skills/moai/references/internal/self-check.md`, `skills/moai/references/source-manifest.json`(판 4: `self-check.md` 의 sha256 값 하나, 줄 임대), `tests/fixtures/help/**`(`card note` 항목만), `tests/card_ready_receipt_cli.rs`와 `tests/e2e_temp_home_cli.rs`(판 2: 기존 READY 노트 본문에 `leftovers: none` 줄만 더해요), `tools/repo_scans.json`(판 3: `reads-docs` 와 `reads-skills` 두 항목에 새 테스트를 규칙 파일과 `--test` 로 더해요, 줄 임대)
- read-only: `src/card/ready_receipt.rs`, `src/bundle/mutant_line.rs`, `tests/areas.toml`, `docs/claude-md-rule-provenance.md`
- forbidden: `CLAUDE.md`, `src/lib.rs`, 영수증 검사(`check_note`)를 느슨하게 하는 변경, 새 Python 테스트, 카드 번호를 SPEC id·가지·커밋·PR 에 쓰는 것, 실제 HOME, 다른 레인의 파일

## Plan

1. RED: `tests/card_ready_leftovers_cli.rs` 의 테스트 셋과 `card::ready_leftovers::tests` 를 먼저 쓰고 코드 없이 실패하는 것을 봐요. 새 테스트 파일의 첫 낱말 `card` 는 `tests/areas.toml` 에 이미 있어요. 테스트는 상태 폴더 이름(`.moai`)을 글자로 쓰지 않아서 `allow.tsv` 의 새 줄이 필요 없어요.
2. GREEN: `src/card/ready_leftovers.rs` 에 줄을 읽는 순수 함수(코드 울타리·4칸 들여쓰기·`- ` 접두사 규칙은 `mutant_line.rs` 와 같아요. `mutant_line.rs` 는 이 SPEC 의 읽기 전용이라 같은 열두 줄을 이 모듈에 두고, 하나로 합치는 일은 뒤 SPEC 의 몫으로 `## Result` 에 적어요)와 공용 체크아웃을 읽는 함수를 두고, `work.rs` 에서 영수증 검사 뒤에 한 줄을 불러요(영수증 거절 글이 앞서요). 공용 체크아웃은 `git worktree list --porcelain` 의 첫 항목이고 `git status --porcelain=v1 -z --ignored=matching -uall` 로 읽어요; 이 저장소의 `git_text` 가 `--no-optional-locks` 를 붙여서 index 를 건드리지 않아요. 예약 시작은 원장의 예약 `acquired_at` 이에요.
3. 글: `docs/lane-rules.md`(Worktree entry and exit 의 한 줄과 Hand-off 1번), `card.md`, 도움말, `self-check.md`.
3d. 판 4: `skills/moai/references/flows/next-flow.md` 는 팁의 크기(8051바이트)를 넘기지 못하게 `tests/next_flow_skill_scan.rs` 가 못박고 있어서(처음 이 실행의 `check affected` 영수증이 실패해서 읽음, 헤드 `b80512db75`) 그 파일에는 아무것도 더하지 않아요. 스킬 쪽 지침은 `self-check.md` 가 맡고, 이 파일의 sha256 을 적은 `source-manifest.json` 의 값 하나만 다시 계산해요. 검사 한도를 올리지 않아요.
3c. 판 3: NeXT 팁의 파일 읽기 검사(`every_file_reader_has_an_always_scan_or_an_exemption`)가 문서를 열어 읽는 새 테스트 파일을 `tools/repo_scans.json` 의 `always` 항목에 올리라고 해서(이 실행에서 읽음), `reads-docs` 와 `reads-skills` 에 올려요. 임시 저장소의 무시 규칙은 `.git/info/exclude` 로, 고정 파일 이름은 이 저장소에 없는 이름으로 바꿔서 `FILE_READ_EXEMPT` 를 건드리지 않아요.
3b. 판 2: 새 거절 때문에 기존 READY 노트 테스트 일곱(`card_ready_receipt_cli` 여섯, `e2e_temp_home_cli` 하나)이 실패해서(이 실행에서 읽음) 그 노트 본문에 `leftovers: none` 줄만 더해요. 영수증 검사 기대는 그대로예요.
4. 뮤턴트: 새 가드 테스트마다 코드를 지우고 실패를 봐요(줄 검사 제거, 목록 필터 제거, `--no-optional-locks` 제거, 예약 시각 필터 제거). 영향 테스트 한 번, READY.
5. observe 창: 이 SPEC 이 올라간 뒤 다음 READY 노트 10개까지(또는 리더가 정한 수까지) 공용 체크아웃 목록이 비지 않은 횟수와 오탐을 `## Result` 에 적어요. 거절로 바꾸는 일은 기록된 결정이에요(D-1).

## Risks

- 공용 체크아웃 목록이 다른 레인이나 운영자의 파일에 오탐할 수 있어요(`.DS_Store`, 같은 때에 도는 다른 레인의 파일). 그래서 이 부분은 observe 로 시작해요(D-1).
- `leftovers: none` 이라고 거짓으로 적을 수 있어요. 이 검사는 쓰는 것을 잊는 일을 막지, 보안 경계가 아니에요(영수증 검사와 같은 한계).
- 저장소와 워크트리 밖(`$TMPDIR`, scratchpad)에 만든 것은 git 이 보지 못해서 보고와 위치 규칙에 기대요. 이 검사가 그 일을 없애지는 못해요.
- 이미 도는 레인 세션은 새 규칙을 모르는 채 READY 를 써서 한 번 거절될 수 있어요. 거절 글이 붙일 줄을 보여 줘서 그 자리에서 고칠 수 있어요(REQ-7).
- `.claude/worktrees/` 의 다른 워크트리 이름이나 `target/` 을 목록에서 빼는 고정 목록이 실제 상태와 어긋날 수 있어요. 목록은 코드 한 곳에 상수로 두고, observe 창에서 오탐으로 드러난 것만 더해요.

- `docs/lane-rules.md` 는 같은 시각에 다른 카드의 예약이 잡고 있어서(줄 임대도 그 예약이 끝날 때까지 거절돼요) 이 변경의 글은 그 예약이 풀린 뒤에 들어가요.

## Decisions

- D-1: 거절의 범위. 선택지 (a) `leftovers:` 줄 검사는 지금 거절, 공용 체크아웃 목록은 observe(답에 드러내기만) — 기본값; (b) 둘 다 지금 거절; (c) 둘 다 observe 로 시작. 근거: 줄 검사는 있고 없음만 읽어 오탐이 없어요. 목록은 다른 레인의 파일에 오탐할 수 있어서 SDD 규칙 11 을 따라 observe 창이 필요해요(Q5). 되돌릴 수 있어요(다음 변경 하나). 분류: none. 기한: 없음(레인은 리더의 배차를 기다려요). 결정자: 리더(또는 사용자).
- D-2: 공용 체크아웃 파일을 가르는 기준. 선택지 (a) 예약 시작 이후 만들거나 바뀐 것(생성 시각과 수정 시각 중 큰 값) — 기본값; (b) 예약 때 목록을 원장에 저장해 비교(더 정확하나 원장 쓰기와 `card take` 변경이 필요해요, 다른 카드의 몫). 기본값은 원장도 `card take` 도 바꾸지 않아요. 되돌릴 수 있어요. 분류: none.
- D-3: 위치 규칙의 범위. 선택지 (a) 버리는 복사본·임시 HOME·`mktemp` 폴더만 `target/<name>/` — 기본값; (b) 레인의 모든 임시 파일(노트, 로그 포함)을 같은 곳으로 합치기. (b)는 scratchpad 를 쓰라는 호스트 지침과 부딪혀서 이 SPEC 에서는 하지 않아요. 되돌릴 수 있어요. 분류: none.

## Result

NeXT 팁 efb3a80ae9 를 합친 가지에서 읽은 것이에요(합치기는 리더 결정 노트 8196 의 (a); 같은 시각의 `check affected` 영수증은 READY 노트가 가져요). 모든 실행은 `RUSTC_WRAPPER=sccache`, 힙 슬롯 `--max 3`, `cargo nextest run --locked --offline` 이에요.

- AC-1: met — `the_lane_guidance_names_the_throwaway_place` 가 `docs/lane-rules.md` 와 `self-check.md` 의 `<worktree>/target/<name>/`·`TMPDIR`·`relative path`·`leftovers:` 를 읽고 통과해요(선택한 128개 테스트 128 통과, 로그 `target/lane-leftovers/fix2.log`).
- 첫 `check affected` 영수증(헤드 `b80512db75`)은 `result=FAIL`(`test=100`)이었어요: `next-flow.md` 가 못박힌 크기 8051바이트를 넘고(`next_flow_skill_scan` 넷) 소스 매니페스트의 sha256 이 낡았어요(`product_cli`·`source_manifest_facts` 둘). 판 4 가 `next-flow.md` 를 팁 그대로 되돌리고 `self-check.md` 의 해시 하나만 다시 계산해서 위 여섯이 통과해요(`fix2.log`: 128 통과).
- AC-2: met — `a_ready_note_without_a_readable_leftovers_line_is_refused` 통과(세 표지, 여섯 잘못된 모양, 두 읽을 수 있는 모양, 노트 수 확인). 뮤턴트: 줄 검사를 끈 복사본에서 이 테스트와 단위 테스트가 실패했어요(exit 100).
- AC-3: met — `the_shared_checkout_is_read_without_changing_it` 통과(새 추적 안 된 파일, 새 무시된 파일과 폴더가 `shared_checkout_new` 에 있고, 예약보다 오래된 파일·고정 목록·레인 워크트리 파일은 없고, `.git/index` 바이트와 수정 시각과 `git status` 출력이 같음). `a_card_without_a_reservation_is_not_checked_and_the_note_is_kept` 통과. 읽을 수 없는 저장소는 단위 테스트 `a_checkout_that_cannot_be_read_is_not_checked` 가 가려요(통합 테스트에서 저장소를 읽을 수 없게 만들지는 않았어요).
- AC-4: met — `card::ready_leftovers::tests` 여섯 개 통과. 뮤턴트: `.skip(1)` 을 지운 복사본이 처음에 살아남아(exit 0) 첫 줄만 있는 본문의 기대를 더했고, 다시 돌리자 실패했어요(exit 100).
- AC-5: met — `the_leftovers_line_is_shown_where_a_lane_looks` 통과(Hand-off 절, `card.md`, `card note --help`).
- 판 2 테스트 수리: `card_ready_receipt_cli`·`e2e_temp_home_cli` 24개 통과(고치기 전에 일곱 실패, `target/lane-leftovers/existing.log`).
- 뮤턴트 요약(모두 커밋된 트리의 복사본이 아니라 제자리 변경 뒤 `git checkout -- <path>` 로 복원, 일곱 개): 줄 검사 끔·고정 목록 필터 끔·예약 시각 필터 끔·`--ignored=matching` 뺌·`--no-optional-locks` 없는 `git status`·자리표시자 거절 뺌 — 모두 실패(exit 100); `.skip(1)` 뺌 — 처음 생존, 테스트를 더한 뒤 실패.
- 알려진 것: `mutant_line.rs` 가 읽기 전용이라 같은 열두 줄(코드 울타리·들여쓰기·`- ` 규칙)을 `ready_leftovers.rs` 에 두었어요. 하나로 합치는 일은 뒤 SPEC 의 몫이에요.
- observe 창: 아직 시작하지 않았어요(이 SPEC 이 NeXT 에 오른 뒤 다음 READY 10개). 오탐 횟수는 비어 있어요.
