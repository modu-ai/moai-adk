---
id: SPEC-20261125-integration-guard-install
title: "the integration branch guard hook is installed by the managed git-hook installer, a session start notices a missing hook, and card finish asks the same verdict check as the landing"
status: draft
revision: 3
created: 2026-11-25
branch: spec/20261125-integration-guard-install
---

# the integration branch guard hook is installed by the managed git-hook installer, a session start notices a missing hook, and card finish asks the same verdict check as the landing

## Intent

통합 브랜치(`NeXT`)를 지키는 `reference-transaction` 훅은 글과 테스트에만 있고 어디에도 설치되지 않아요. 사용자 지시(2026-11-24T23:35:48Z)로 한 전수 조사(`~/.moai/reports/moai-adk/2026-11-24-rules-that-do-not-work.md` 1·2·5번)가 이 두 가지를 쟀어요: `git reflog show NeXT` 에 손 fast-forward 머지가 136번, `moai card land` 가 1번. 그리고 `card.landed` 이벤트 204건이 리더의 `card finish` 경유예요. 오늘(2026-11-25) 우회 모드 세션에서 `git branch -f NeXT HEAD` 가 거절 없이 실행됐고, 이 저장소 `.git/hooks` 에는 `commit-msg`·`pre-commit` 뿐이에요(이 트리에서 읽음). 사용자가 이 카드와 도구 사전점검 카드의 착지를 전역 빌드 설치의 조건으로 정했어요(카드 노트 8925).

읽은 사실(NeXT 팁 `df6e6a3f22`, 이 트리):

1. 훅 텍스트는 이미 있어요. `src/card/guard.rs:51` 의 `hook_text()` 가 POSIX 셸 스크립트를 주고, `tests/card/card_guard_cli.rs`·`card_land_cli.rs` 가 시험용 저장소에 직접 써서 써요. 설치 명령이 없어요(`SPEC-20261121-ref-guard` 의 Non-goals 가 "훅 파일 설치"를 뺐어요).
2. 설치기가 이미 있어요. `src/git_hooks/mod.rs` 의 관리 이름은 `MANAGED = ["pre-commit", "pre-push"]` 이고, `git-hooks-plan`·`git-hooks-apply`·`git-hooks-revert` 와 `moai update` 의 `--confirm-git-hooks` 층이 같은 모듈을 써요. ref-guard SPEC 이 말한 "세 번째 관리 이름"이 바로 이 자리예요. 이 모듈은 `std`·`serde_json`·`sha2` 만 쓰고(`tests` 가 경로로 컴파일해요), 사용자가 직접 둔 훅·고친 훅은 건드리지 않아요.
3. 기록 파일은 이 저장소에 이미 있어요. `.git/moai/integration-branch` 가 `refs/heads/NeXT` 를 가리켜요(첫 `card land` 가 썼어요). 훅은 이 기록이 없으면 아무것도 막지 않아요.
4. `card finish` 는 착지 증거(`--head`, 없으면 원장의 마지막 `card.landed` 의 head)를 NeXT 팁 안에 들어 있는지만 보고(`src/card/finish.rs:142` `landing_evidence`), 판정 이벤트를 읽지 않아요. 반면 `card land` 가 카드를 닫을 때는 `check::read`(`src/card/land/check.rs:154`)가 낸 PASS 가 모든 모드에서 필요해요(`SPEC-20261124-land-closes-card` REQ-4). 그래서 손 머지 뒤 `card finish` 로 닫으면 판정 없이 카드가 닫혀요.
5. 세션 시작 훅(`src/hook_session_start.rs`)은 파일만 읽고 다른 프로그램을 시작하지 않아요. 한 줄을 더하는 자리가 있어요.

이 SPEC 은 세 가지를 해요. (a) 설치기가 `reference-transaction` 을 세 번째 관리 훅으로 설치하고 기록을 써요. (b) 세션 시작이 훅이 없으면 한 줄로 알려요. (c) `card finish` 가 착지와 같은 판정 검사를 해요. 이 저장소에 실제로 설치하는 일은 사용자의 선택이에요(Decisions D-2).

판 1(head fbfc58557b, digest `sha256:53688f93f71b697d555ff94269323ce9d1e25c0a2070f3df8287eb7e11386f57`)은 아래 Interview 의 두 답으로 승인됐고(리더 노트 8937), 판 2 는 구현 중에 읽은 사실 셋을 고쳐요: (1) 기존 설치기는 지울 Go 훅이 있을 때만 빈 이름을 설치해서, 세 번째 훅이 훅 없는 저장소에는 설치되지 않아요. 설치 조건을 REQ-1 에 적어요(D-6). (2) `tests/git_hooks_cli.rs` 가 `src/git_hooks/mod.rs` 를 경로로 컴파일해서 이 파일은 프로필을 읽을 수 없고, 미리보기 동사 파일 `src/ops/git_hooks_plan.rs` 도 같은 값을 넘겨야 해요(쓰기 목록에 더함). (3) 카드 테스트는 `tests/card/main.rs` 한 묶음(`--test card`)이라 AC-8 의 검증 명령 `--test card_guard_cli …` 는 없는 대상이었어요: AC-8 을 `--test card -- card_guard_cli card_land_cli card_finish_cli` 로 고치고 나머지를 AC-9 로 나눠요. 방향과 D-1(선택지 a)은 그대로예요. 판 2 의 승인은 다시 받아요. 판 3 은 새 테스트 파일의 이름만 바꿔요: `tests/card_finish_verdict_cli.rs` 는 `card` 영역(`tests/card/main.rs` 가 있는 영역)이라 `tests/card/` 안에 있어야 하는데 그 `main.rs` 를 다른 카드가 쥐고 있어서, 영역 낱말이 다른 `tests/gate_finish_verdict_cli.rs` 로 바꿔요(리더 결정, 카드 노트 9002). 쓰기 범위와 AC-5·AC-6 의 검증 줄의 파일 이름만 바뀌고 방향과 D-1~D-6 은 그대로예요. 판 3 의 승인은 다시 받아요.

## Interview

- Q1: 이 SPEC 의 첫 판(revision 1, head fbfc58557b)을 승인할까요? → A: 승인(사용자의 선택 "승인, D-1 은 모든 모드에서 거절 (Recommended)", 리더 노트 8937 이 옮겼어요)
- Q2: D-1, `card finish` 가 감사 판정 없는 head 를 만나면 어떻게 할까요? → A: 선택지 a, 모든 모드에서 거절해요(같은 선택)
- Q3: 이 SPEC 의 개정 2(head 082a20e849)를 승인할까요? → A: 승인(사용자의 선택 "승인 (Recommended)", 리더 노트 8943 이 옮겼어요)

## Non-goals

- 이 저장소의 `.git/hooks`·`.git/moai` 에 쓰는 일. 코드와 시험은 임시 저장소에서만 설치해요. 실제 저장소 설치는 사용자가 `moai git-hooks-plan` 으로 보고 `--confirm-git-hooks` 토큰으로 직접 적용해요(저장소 보호 설정 변경, `guarded`).
- `work_flow.land_verdict: enforce` 로 바꾸는 일(`.moai/project.yaml`). 사용자가 "이 카드 착지 뒤 켜기"를 골랐고(카드 노트 8824), 착지 뒤 리더가 후속 카드로 켜요. 이 SPEC 은 프로필을 건드리지 않아요.
- 보호 훅의 판정 규칙을 바꾸는 일(`SPEC-20261121-ref-guard`). 훅 텍스트의 바이트는 그대로예요. 훅을 `install/git-hooks/reference-transaction` 파일로 옮기되 내용은 한 바이트도 바꾸지 않아요.
- `git branch -C`·`-c` 가 훅 없이 브랜치를 옮기는 빈틈(ref-guard AC-20)과 `.git` 을 일부러 고치는 호출. 보호 훅은 사고를 막지 사보타주를 막지 않아요.
- 도구 사전점검(tool-preflight)이 우회 모드에서도 통합 브랜치 이동을 거절하게 하는 일(별도 카드).
- 기록 파일이 없고 프로필에도 통합 브랜치가 없는 프로젝트의 보호. 그런 프로젝트에는 훅이 아무것도 안 막고, 세션 시작도 아무 줄을 안 더해요.

## Requirements

- REQ-1: When the git-hook installer plans, applies or reverts, the installer shall manage the hook `reference-transaction` as a third managed name whose installed bytes equal the text of `install/git-hooks/reference-transaction`, the same bytes `guard::hook_text()` returns · failure: none — a free name gets the executable hook only in a repository whose profile names an integration branch or that already holds the guard record, no other repository gets it and no retire of a Go-generation hook installs it there, a hook the user holds or edited stays untouched and blocks that name, and the other two names behave as before.
- REQ-2: When the installer applies the managed set in the primary worktree of a repository whose profile names an integration branch, the installer shall write the guard record in the common git folder when it is absent and name the record state in its answer · failure: none — a linked worktree or a profile that names no branch writes no record and the answer says `not_written` with the reason, and an existing record is never rewritten.
- REQ-3: When a repository has the installed hook and the record, the guard shall refuse every hand move of the integration branch that `SPEC-20261121-ref-guard` AC-1 lists and let the move of `card land` through · failure: none — the same refusals and the same pass as the hook gave when a test wrote it by hand.
- REQ-4: When a session starts in a repository whose profile names an integration branch, the session-start handler shall add one line that names the missing guard exactly when the hook file is absent or not executable or the record is absent · failure: none — the handler reads two files and starts no program, both present adds no line, a read that fails adds no line, and a profile that names no branch adds no line.
- REQ-5: When `card finish` closes a card that changes the repository on a landing head of a NeXT-Flow profile, the card command shall run the verdict check of the landing for that head with the head already inside the tip and refuse the close with the check's code and reason when it finds no usable PASS, in every mode · failure: none — the refusal writes nothing to the ledger and names the verdict step of the landing, and the head the landing's start row names is closed on that row's verdict.
- REQ-6: When `card finish` closes a card on a merged pull request, a report or a note, the card command shall behave as it did before this SPEC · failure: none — no verdict check runs for a card with no landing head on a NeXT-Flow profile.
- REQ-7: When the installer, the session-start line or the finish check changes what a person sees or types, the card operations document and the two git-hook operations documents shall say so · failure: none — the three documents name the third hook, the record, the `not_written` answer, the session-start line and the finish refusal codes.

## Acceptance

- AC-1 (REQ-1): the plan of a temporary repository lists `reference-transaction` with the shipped bytes, the apply installs it executable beside the other two, a user hook of that name is kept and blocks only that name, a hook edited after the install is kept, the revert puts back what it retired, a repository whose profile names no integration branch and holds no record gets no hook and an unchanged plan, and a test pins that the shipped file equals `guard::hook_text()` · verify: `cargo test --locked --offline --test integration_guard_install ac1_`
- AC-2 (REQ-2): in a temporary primary worktree with a profile that names `NeXT` the apply writes `refs/heads/NeXT` as the record and the answer says `written`, a second apply leaves the record and says `present`, a linked worktree and a profile with no branch say `not_written` with the reason, and the record is written after the hook · verify: `cargo test --locked --offline --test integration_guard_install ac2_`
- AC-3 (REQ-3): in a temporary repository set up only through `git-hooks-plan` and `git-hooks-apply` (the dangerous path itself, not a hand-written hook), `git update-ref`, `git branch -f`, `git branch -D`, `git merge --ff-only` in the worktree that has the branch checked out and `git branch -m` leave the branch where it was and exit non-zero, and a `card land` of an audited head moves it · verify: `cargo test --locked --offline --test integration_guard_install ac3_`
- AC-4 (REQ-4): the session-start answer for a temporary repository with a profile that names `NeXT` carries one extra line when the hook is absent, when the hook is not executable and when the record is absent, carries none when both are present or when the profile names no branch, and starts no program · verify: `cargo test --locked --offline --test hook_session_start_cli guard_line_`
- AC-5 (REQ-5): `card finish --head` of a landed code card with no verdict record, with a verdict for another head and with a FAIL refuses with `no_verdict`, `verdict_other_head` and `verdict_fail`, leaves the card and the ledger as they were in observe and enforce mode, closes on a PASS recorded by a session other than the lander, and closes a head landed by `card land` on the verdict of its start row · verify: `cargo test --locked --offline --test gate_finish_verdict_cli ac5_`
- AC-6 (REQ-6): `card finish` with `--evidence-report`, `--evidence-note`, a merged pull request link on a profile with no integration branch and a research card closes as before with no verdict read · verify: `cargo test --locked --offline --test gate_finish_verdict_cli ac6_`
- AC-7 (REQ-7): the three operations documents name the third hook, the record state words, the session-start line and the finish refusal codes · verify: `cargo test --locked --offline --test integration_guard_install ac7_`
- AC-8 (REQ-3, REQ-5): the existing guard, landing and finish tests of the card test target still pass, and every test that closed a landed code card with `card finish --head` before this SPEC passes after the fixture of that test records the verdict it needs, none by a weaker assertion · verify: `cargo test --locked --offline --test card -- card_guard_cli card_land_cli card_finish_cli`
- AC-9 (REQ-1, REQ-4, REQ-5): the existing git-hook, update-layer, session-start and cleanup-outbox tests still pass · verify: `cargo test --locked --offline --test git_hooks_cli --test update_git_hooks_cli --test update_tux_git_hooks --test hook_session_start_cli --test card_cleanup_outbox_cli`

## Scope

- write: `.moai/specs/SPEC-20261125-integration-guard-install/spec.md`
- write: `install/git-hooks/reference-transaction`, `src/git_hooks/mod.rs`, `src/git_hooks/update.rs`, `src/ops/git_hooks_plan.rs`, `src/ops/git_hooks_apply.rs`
- write: `src/card/guard.rs`, `src/card/finish.rs`, `src/hook_session_start.rs`
- write: `tests/integration_guard_install.rs`, `tests/gate_finish_verdict_cli.rs`, `tests/git_hooks_cli.rs`, `tests/hook_session_start_cli.rs`
- write: `tests/card/card_finish_cli.rs`, `tests/card_cleanup_outbox_cli.rs`- write: `docs/design-review/appendix/operations/git_hooks_plan.md`, `docs/design-review/appendix/operations/git_hooks_apply.md`
- read-only: `src/card/land.rs`, `src/card/land/check.rs`, `src/card/landed.rs`, `src/profile.rs`
- forbidden: the `.git/hooks` and `.git/moai` folders of this repository, `.moai/project.yaml`, `CLAUDE.md`

## Plan

1. 테스트 먼저(RED): `tests/integration_guard_install.rs` (AC-1·2·3·7, 실제 git 과 임시 저장소, 설치는 `git-hooks-apply` 로만), `tests/gate_finish_verdict_cli.rs` (AC-5·6), `tests/hook_session_start_cli.rs` 의 `guard_line_` 시험 (AC-4). 각각 변경 전 트리에서 실패하는 것을 먼저 읽어요.
2. `install/git-hooks/reference-transaction`: `hook_text()` 의 현재 바이트를 그대로 옮긴 파일. `src/card/guard.rs` 의 `hook_text()` 는 같은 파일을 `include_str!` 로 돌려줘요(원본이 하나).
3. `src/git_hooks/mod.rs`: `MANAGED` 가 셋, `text_of` 가 세 이름을 알아요. 이 파일은 `tests/git_hooks_cli.rs` 가 경로로 컴파일해서 `crate::` 를 쓸 수 없으므로, 계획·적용에 "이 저장소에서 보호 훅을 설치할지"를 불리언으로 받는 짝 함수(`plan_with_guard`, `apply_with_guard`)를 더하고 기존 `plan`·`apply`·`apply_with_pause` 의 모양은 그대로 둬요(불리언 거짓). 세 번째 이름은 불리언이 참일 때만 설치하고, Go 훅을 지우는 실행이 "빈 이름을 함께 설치"하는 목록에도 그 불리언이 거짓이면 넣지 않아요. `tests/git_hooks_cli.rs` 의 `MANAGED` 고정 시험을 셋으로 갱신해요.
4. `src/git_hooks/update.rs`, `src/ops/git_hooks_plan.rs`, `src/ops/git_hooks_apply.rs`: 불리언은 프로필이 통합 브랜치를 가리키거나 기록이 이미 있을 때 참이에요(`update.rs` 의 한 함수가 계산하고 세 곳이 같은 값을 써서 미리보기와 적용의 토큰이 어긋나지 않아요). 적용이 끝난 뒤 `guard::protected(project, true)` 로 기록을 쓰고 답에 `guard: {record: written|present|not_written, why}` 를 더해요.
5. `src/hook_session_start.rs`: 프로필이 통합 브랜치를 가리키면 `<공통 git 폴더>/hooks/reference-transaction` 의 존재·실행 비트와 `moai/integration-branch` 의 존재를 읽어 한 줄(`MoAI: integration branch <name> is not guarded: <what is missing> (moai git-hooks-plan)`)을 더해요.
6. `src/card/finish.rs`: `Proven::Landing` 일 때 `check::Input { on_tip: false, verdict_head: head, lander: req.session, .. }` 로 `check::read` 를 부르고, 사용할 수 있는 PASS 가 없으면 `land` 의 `no_pass_reason` 과 같은 말로 거절해요(쓰기 전).
7. 기존 `finish` 테스트 두 파일의 준비 단계가 닫기 전에 판정을 기록하게 고쳐요(선택지 a 일 때만).
8. 문서 세 개. 그중 카드 연산 문서(`docs/design-review/appendix/operations/card.md`)는 프로필의 공유 파일이라 스코프에 넣지 않고 줄 임대(`card line-lease`)로 한 번에 써요. 도움말 고정 파일(`tests/fixtures/help/*.txt`), `tests/fixtures/state_folder_literals/allow.tsv`, `tools/repo_scans.json` 도 어떤 검사가 이름을 댈 때만 같은 방식으로 한 줄씩 써요. 그다음 가벼운 검사(`cargo fmt --all`, clippy, 장부 테스트 `areas.toml`·`repo_scans.json`·`allow.tsv`·source-manifest), 가드 뮤턴트(Result 에 줄로), 그 다음 `check affected --card` 한 번, READY 노트.

## Risks

- 훅이 설치된 저장소에서는 리더의 손 fast-forward 머지가 거절돼요(ref-guard 의 Risks 에 이미 적힌 대로 `card land` 로 착지해요). 설치가 사용자의 선택이라 이 SPEC 의 코드만으로는 이 저장소의 흐름이 바뀌지 않아요.
- `card finish` 가 판정 없는 head 를 항상 거절하면 지금 손 머지로 닫던 흐름이 막혀요. 의도한 결과지만 리더가 먼저 판정을 받아야 닫을 수 있어요(Decisions D-1 의 선택지 b 를 고르면 관찰 모드에서는 보고만 해요).
- 기존 `finish` 테스트 65곳(`--head` 호출: `tests/card/card_finish_cli.rs` 59, `tests/card_cleanup_outbox_cli.rs` 6)이 판정 없이 닫아요. 선택지 a 에서는 이 준비 단계가 판정 기록을 포함하도록 고쳐야 하고, 약한 단언으로 통과시키지 않아요. 선택지 b 에서는 이 수정이 필요 없어요. 이 차이가 D-1 을 고르는 비용이에요.
- `moai` 가 `card guard` 를 모르는 낡은 빌드가 설치한 훅은 통합 브랜치의 모든 이동을 거절해요(ref-guard 가 이미 적은 실패 닫힘). 설치기와 `card guard` 가 같은 바이너리라 이 SPEC 의 설치는 그 조합을 만들지 않아요.
- 기록 파일이 있어도 훅이 없는 상태(이 저장소의 지금 모습)와 훅은 있는데 기록이 없는 상태를 세션 시작이 같은 줄로 알려요. 구분은 줄의 `<what is missing>` 이 해요.
- 실제 저장소에 설치하는 순간부터 우회 모드 세션도 `git branch -f NeXT` 가 거절돼요. 되돌리는 길은 `moai git-hooks-revert <stamp>` 예요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | `card finish` 는 NeXT-Flow 프로필의 착지 head 를 닫을 때 모든 모드에서 사용할 수 있는 PASS 를 요구해요(선택지 a, 권장). 선택지 b 는 `card land --keep-open` 처럼 모드를 따라 관찰 모드에서는 보고만 해요. | 카드 기준 3 이 "판정 없는 head 를 거절"이라고 적었고, `card land` 가 카드를 닫을 때는 이미 모든 모드에서 PASS 를 요구해요(`SPEC-20261124-land-closes-card` REQ-4). 같은 닫기를 두 길이 다르게 판정하면 `finish` 가 우회로가 돼요(조사 보고서 2번). | 사용자 선택. 승인 뒤에는 REQ-5 한 줄과 AC-5 를 바꾸는 판이에요. |
| D-2 | 이 저장소에 훅을 실제로 설치하는 일은 이 SPEC 의 밖이에요. 코드는 `git-hooks-plan`·`git-hooks-apply` 의 세 번째 이름만 더하고, 리더가 계획을 읽은 뒤 사용자가 `--confirm-git-hooks` 토큰으로 직접 적용해요. | 저장소 보호 설정 변경은 `guarded` 예요(스킬의 선택 분류 2번). 사용자가 이 카드와 도구 사전점검 카드의 착지를 전역 빌드 설치의 조건으로 정했어요(노트 8925). | 적용 전에는 쓰기가 없어요. 적용 뒤에는 `moai git-hooks-revert <stamp>`. |
| D-3 | `work_flow.land_verdict: enforce` 는 이 카드가 착지한 뒤 별도로 켜요. | 사용자 선택 2026-11-24T23:45:12Z "이 카드 착지 뒤 켜기"(노트 8824). 이 카드의 마지막 단계로 프로필 한 줄을 넣는 길도 노트가 열어 뒀지만, 훅과 `finish` 변경이 착지하기 전에 프로필을 바꾸면 그 착지 자체가 막힐 수 있어서 후속으로 둬요. | `.moai/project.yaml` 의 키 한 줄. |
| D-4 | 훅 텍스트의 원본은 `install/git-hooks/reference-transaction` 한 파일이고 `hook_text()` 는 그 파일을 `include_str!` 로 돌려줘요. | 설치기 모듈은 `crate::card` 를 참조하지 못하고(경로 컴파일), 두 군데에 같은 글을 두면 어긋나요. | 파일을 `guard.rs` 의 문자열로 되돌리고 설치기가 그 바이트를 받는 판. |
| D-6 | 세 번째 훅은 통합 브랜치가 있는 저장소(프로필이 가리키거나 기록이 이미 있음)에서만 설치해요. | 기록이 없으면 훅은 아무것도 막지 않아요. 모든 git 프로젝트에 설치하면 `moai update` 의 미리보기에 아무 일도 안 하는 훅의 설치 줄과 확인 토큰이 생겨요. 기존 `update`·`git-hooks` 테스트의 준비 단계에는 `integration_branch` 가 없어서 이 조건 아래에서는 그 출력이 그대로예요(읽음). | 불리언을 항상 참으로 두는 판. |
| D-5 | 세션 시작의 한 줄은 훅·기록 둘 중 하나만 빠져도 같은 문장 틀로 쓰고 빠진 것을 이름으로 말해요. | 세션 시작은 파일만 읽고 결정하지 않는 규칙(`SPEC-20261123-hook-session-start`)이에요. | 줄 하나를 지우는 판. |
