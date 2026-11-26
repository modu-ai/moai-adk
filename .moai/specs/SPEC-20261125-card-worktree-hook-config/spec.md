---
id: SPEC-20261125-card-worktree-hook-config
title: "A card take or reclaim says when the hook settings point at files the session's start folder does not have"
status: draft
revision: 2
created: 2026-11-25
---

# A card take or reclaim says when the hook settings point at files the session's start folder does not have

## Intent

Go 세대 프로젝트의 `.claude/settings.json` 은 훅을 `$CLAUDE_PROJECT_DIR/.claude/hooks/moai/*.sh` 로 가리켜요. 세션을 기본 체크아웃에서 시작해 `EnterWorktree` 로 카드 워크트리에 들어가면 설정은 워크트리 것으로 옮겨 가고 그 변수는 세션을 시작한 폴더에 남아서, 시작 폴더에 스크립트가 없으면 `Stop hook failed … handle-stop.sh: No such file or directory` 가 나요(다른 저장소 리더의 실측; 공식 문서 `https://code.claude.com/docs/en/worktrees` 의 "Hook paths don't follow the worktree" 와 "The move takes … settings to that location", 읽은 시각 2026-11-26T03:01:37Z, 받은 파일 sha256 1b3eac588abbe62de80f6f6d4275c80de00d989883007cddcf9cdb773b373f20). 레인이 이유를 모른 채 오류를 보지 않도록 `card take` 와 `card reclaim` 이 시작 폴더에 없는 훅 파일을 한 줄로 알리고, 그대로 따르면 파일이 찾아지는 다음 걸음을 말해요. 사용자의 설정 파일은 고치지 않아요.

개정 1은 워크트리에 파일이 있는지를 봐서, 실측한 상황에서는 조용했고 안내한 자리에는 파일이 없었어요(독립 감사 FAIL). 개정 2는 시작 폴더를 기준으로 봐요.

## Non-goals

- 사용자의 설정 파일, 훅 스크립트, 게이트와 권한을 바꾸지 않아요.
- Go 세대가 아닌 훅의 파일은 보지 않아요.

## Interview

- Q1: 어느 폴더를 기준으로 파일이 있는지 보나요? -> A: 세션이 시작하는 폴더요. 레인 세션은 기본 체크아웃에서 시작해 `EnterWorktree` 로 들어가요(카드 실측과 공식 문서). `card take` 를 돌리는 셸에는 `CLAUDE_PROJECT_DIR` 이 없어서 기본 체크아웃을 시작 폴더로 읽어요(D-1).
- Q2: 어떤 다음 걸음을 말하나요? -> A: 따랐을 때 파일이 찾아지는 것만 말해요(리더 수리 조건). 워크트리가 다 가졌으면 거기서 세션을 시작하라고, 아니면 어느 시작 폴더도 다 갖지 못했고 설정은 사용자가 정한다고 말해요(D-3).

## Requirements

- REQ-1: When a card is taken or reclaimed into a worktree and the start folder lacks a hook script under `.claude/hooks/moai/` that the worktree's `.claude/settings.json` points at, the answer shall carry one warning that names those scripts and the start folder; the start folder is the primary checkout of the worktree's repository, and the worktree itself when no primary checkout can be read.
- REQ-2 (REQ-1): IF the start folder has every script the settings point at, or the settings point at none, THEN the answer shall carry no hook warning.
- REQ-3 (REQ-1): IF the settings file is absent, or it or a hook folder to be judged cannot be read, THEN the take or reclaim shall succeed with no hook warning.
- REQ-4 (REQ-1): The check shall only read: the worktree's settings file, the worktree's `.git` file with the git folder's `commondir`, and the listing of the hook folders it judges.
- REQ-5 (REQ-1): The warning shall name a next step after which the named scripts are found: IF the worktree has every script the settings point at, THEN start the lane session in the worktree folder; ELSE say that no start folder has them all and that the user decides about the hook settings, naming `moai go-hooks-plan <worktree>`.
- REQ-6 (REQ-5): The WORK workflow shall state the same two steps and say that a settings file is never edited to quiet the warning.

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-3, REQ-4): 임시 폴더에서 설정이 가리키는 스크립트 중 대상 폴더에 없는 것만 이름순으로 돌려주고(있는 것·가리키지 않는 것·Go 세대가 아닌 훅은 빼고), 설정이나 폴더를 읽을 수 없으면 오류로 답하며, 연결 워크트리의 `.git` 파일에서 기본 체크아웃을 읽고(아니면 없음), 아무것도 쓰지 않아요 · verify: `cargo nextest run --lib -E 'test(go_hooks::tests::missing_scripts)'`
- AC-2 (REQ-1, REQ-2, REQ-3, REQ-5): 실제 git 연결 워크트리와 임시 HOME 에서 `card take` 의 응답이 (가) 기본 체크아웃에 없고 워크트리에 있으면 스크립트와 기본 체크아웃을 말하고 워크트리에서 세션을 시작하라고 한 줄로 알리며 그 폴더에 파일이 실제로 있고, (나) 둘 다 없으면 그 걸음을 말하지 않고 `moai go-hooks-plan` 을 말하며, (다) 기본 체크아웃에 다 있으면 줄이 없고, (라) 설정이 없거나 읽을 수 없으면 줄이 없어요. `card reclaim` 도 (가)와 (다)가 같고, 명령은 모두 성공해요 · verify: `cargo nextest run --test card -E 'test(the_hook_settings_of_the_worktree)'`
- AC-3 (REQ-6): `skills/moai/workflows/work.md` 가 두 걸음과 `moai go-hooks-plan` 을 말해요 · verify: `rg -n 'go-hooks-plan' skills/moai/workflows/work.md`

## Scope

- write: `src/go_hooks.rs`(없는 스크립트를 계산하는 읽기 전용 함수와 단위 시험), `src/card/dispatch.rs`(`take`·`reclaim` 응답의 `warnings` 한 줄), `skills/moai/workflows/work.md`, `tests/card/card_take_worktree_cli.rs`, `tests/card/card_reclaim_cli.rs`, 이 SPEC 폴더. 스킬 글을 고치면 소스 매니페스트는 line-lease 안에서 `moai check fix` 로 맞춰요.
- write 하지 않는 것: 사용자의 `.claude/settings.json` 과 어떤 훅 설정 파일, 훅 스크립트, 게이트와 권한, 다른 프로젝트.

## Plan

1. RED: 시작 폴더 기준의 단위 시험과 통합 시험(가·나·다·라)을 먼저 고쳐 실패를 봐요.
2. GREEN: `missing_scripts` 가 설정을 읽는 폴더와 파일을 찾는 폴더를 따로 받고, `.git` 파일에서 기본 체크아웃을 읽는 함수를 더해, 경고 문장을 두 걸음으로 나눠요.
3. WORK 글을 같은 두 걸음으로 고치고 매니페스트를 맞춰요.
4. 가드 뮤턴트, `moai check affected --card`, READY.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 시작 폴더는 워크트리 저장소의 기본 체크아웃이에요(`.git` 파일의 `gitdir:` → `commondir` → `.git` 폴더의 부모, 프로세스 없이 읽기). 읽을 수 없으면 워크트리를 시작 폴더로 봐요. | 카드 실측과 공식 문서: 변수는 시작 폴더에 남아요. Bash 환경에는 `CLAUDE_PROJECT_DIR` 이 없어요(이 세션에서 확인). | 다른 시작 폴더를 알 방법이 생기면 그 값을 먼저 써요. |
| D-2 | 시작 폴더에 다 있고 워크트리에만 없으면 경고하지 않아요. | 기본 흐름(기본 체크아웃에서 시작 → `EnterWorktree`)에서는 오류가 없어요. 카드 기준 1의 "없는 파일을 가리키는"은 `$CLAUDE_PROJECT_DIR` 이 풀리는 폴더 기준이에요. | 워크트리에서 직접 시작하면 실패한다는 두 번째 문장을 더해요. |
| D-3 | 어느 폴더도 다 갖지 못하면 세션을 어디서 시작하라고 하지 않고, 설정은 사용자가 정한다고 말하며 읽기 전용 미리보기 `moai go-hooks-plan <worktree>` 를 이름으로 알려요. | 따랐을 때 풀리지 않는 걸음은 말하지 않아요. 설정 변경은 사용자의 몫이에요. | 명령 이름을 빼요. |

## Risks

- R-1: 세션을 기본 체크아웃이 아닌 폴더에서 시작하면 판단 기준이 실제와 달라요. 경고는 기준 폴더를 이름으로 말해서 읽는 사람이 알 수 있어요.
- R-2: 데스크톱 앱의 워크트리 세션은 설정도 기본 체크아웃에서 읽어요(같은 문서). 이 검사는 CLI 흐름 기준이에요.

## Result

개정 2의 구현 전이에요. 개정 1의 결과(head 4e1a134960)는 독립 감사 FAIL(노트 9871)로 무효예요.
