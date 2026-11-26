---
id: SPEC-20261125-card-worktree-hook-config
title: "A card take or reclaim says when the worktree's hook settings point at files that are not there"
status: draft
revision: 1
created: 2026-11-25
---

# A card take or reclaim says when the worktree's hook settings point at files that are not there

## Intent

Go 세대 프로젝트의 `.claude/settings.json` 은 훅을 `$CLAUDE_PROJECT_DIR/.claude/hooks/moai/*.sh` 로 가리켜요. 세션을 기본 체크아웃에서 시작해 `EnterWorktree` 로 카드 워크트리에 들어가면 그 변수는 기본 체크아웃으로 풀려서 `Stop hook failed … handle-stop.sh: No such file or directory` 가 나요(다른 저장소 리더의 보고, 그 저장소의 실측이고 이 저장소에서는 재현되지 않았어요). 레인이 이유를 모른 채 오류를 보지 않도록 `card take` 와 `card reclaim` 이 워크트리의 훅 설정이 가리키지만 없는 파일을 한 줄로 알리고, WORK 글이 레인 세션을 어디서 시작하는지 말해요. 사용자의 설정 파일은 고치지 않아요.

## Requirements

- REQ-1: When a card is taken or reclaimed into a worktree, the answer shall carry one warning that names the hook scripts under `.claude/hooks/moai/` which the worktree's `.claude/settings.json` points at and the worktree does not have.
- REQ-2 (REQ-1): IF every script the settings point at is there, or the settings point at none, THEN the answer shall carry no hook warning.
- REQ-3 (REQ-1): IF the settings file is absent, or the file or the hook folder cannot be read, THEN the take or reclaim shall succeed with no hook warning.
- REQ-4 (REQ-1): The check shall only read the settings file and list the hook folder.
- REQ-5: The WORK workflow shall say where a lane session starts so that the hook settings resolve to files that exist.

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-3, REQ-4): 임시 폴더의 `.claude/settings.json` 이 가리키는 스크립트 중 없는 것만 이름순으로 돌려주고(있는 것·가리키지 않는 것·Go 세대가 아닌 훅은 빼고), 설정이 없거나 읽을 수 없으면 아무것도 돌려주지 않으며, 읽기가 폴더에 아무것도 쓰지 않아요 · verify: `cargo nextest run --lib -E 'test(go_hooks::tests::missing_scripts)'`
- AC-2 (REQ-1, REQ-2, REQ-3): 실제 git 연결 워크트리와 임시 HOME 에서 `card take` 와 `card reclaim` 의 응답이 훅 설정이 가리키는 없는 파일을 `warnings` 한 줄로 말하고, 없는 파일이 없으면 그 줄이 없으며, 두 명령은 그래도 성공해요 · verify: `cargo nextest run --test card -E 'test(the_hook_settings_of_the_worktree)'`
- AC-3 (REQ-5): `skills/moai/workflows/work.md` 가 훅 설정이 없는 파일을 가리킨다는 `take` 경고가 보이면 레인 세션을 그 워크트리 폴더에서 시작한다고 말하고, 사용자의 설정 파일을 고치지 않는다고 말해요 · verify: `rg -n 'hook settings' skills/moai/workflows/work.md`

## Scope

- write: `src/go_hooks.rs`(없는 스크립트를 계산하는 읽기 전용 함수와 단위 시험), `src/card/dispatch.rs`(`take`·`reclaim` 응답의 `warnings` 한 줄), `skills/moai/workflows/work.md`, `tests/card/card_take_worktree_cli.rs`, `tests/card/card_reclaim_cli.rs`, 이 SPEC 폴더. 스킬 글을 고치면 소스 매니페스트는 line-lease 안에서 `moai check fix` 로 맞춰요.
- write 하지 않는 것: 사용자의 `.claude/settings.json` 과 어떤 훅 설정 파일, 훅 스크립트, 게이트와 권한, 다른 프로젝트.

## Plan

1. RED: `go_hooks` 의 `missing_scripts` 단위 시험을 먼저 써서 실패를 봐요.
2. GREEN: `go_hooks::missing_scripts(project)` 를 더해요(`settings`, `split`, `scripts` 를 그대로 써요).
3. `take` 와 `reclaim` 의 응답에 경고를 더하고(읽기 오류는 경고 없음), 통합 시험으로 고정해요.
4. WORK 글에 레인 세션의 시작 위치를 적고 매니페스트를 맞춰요.

## Result

head a7ee9ee498687622c9b79cc8722e0548ef858ec5(base 9c8b7fb76313f0bd2d90950e3aa52531d3151904)에서 `verify` 명령을 쓴 그대로 다시 실행했어요(2026-11-26).

- AC-1: met — `cargo nextest run --lib -E 'test(go_hooks::tests::missing_scripts)'` 3/3 통과
- AC-2: met — `cargo nextest run --test card -E 'test(the_hook_settings_of_the_worktree)'` 2/2 통과
- AC-3: met — `rg -n 'hook settings' skills/moai/workflows/work.md` exit 0(40행)

가드 뮤턴트 6/6 exit 100(READY 노트 9779). 승인은 노트 9715(리더 위임 판단)이고 승인 블록은 착지 때 리더가 써요.
