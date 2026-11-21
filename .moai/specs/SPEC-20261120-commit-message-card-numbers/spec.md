---
id: SPEC-20261120-commit-message-card-numbers
title: "trailer-check refuses card numbers in commit messages"
status: draft
revision: 2
created: 2026-11-20
parent: null
---

# trailer-check refuses card numbers in commit messages

## Intent

2026-11-19 규칙(카드 번호는 SPEC id·가지·커밋·PR 에 쓰지 않는다) 이후에도 커밋 메시지에 카드 번호가 실렸고, 감사만이 그것을 찾아 수리 가지(-r2)를 만들게 했어요(리더 세션 점검 보고의 P3). 기계가 먼저 거절해야 해요.

이 조각은 trailer-check 가 범위의 각 커밋 제목·본문에서 카드 번호 꼴(`C-<n>`, `c-<n>`)을 검사하게 해요. 규칙 시작은 기록된 커밋으로 정해요(`--card-numbers-since`, 그 커밋 자신은 면제 — `--verified-since` 와 같은 모양): 플래그가 없으면 규칙이 꺼져 기존 범위의 결과가 변하지 않아요. SPEC id(SPEC-YYYYMMDD-slug), 버전(v4.2647.9), 16진 SHA 는 꼴이 아니라 통과하고, `AC-<n>` 같은 이웃 꼴과 원장 출력 인용(따옴표 안의 `card C-<n>`)도 통과해요. 같은 스캔을 `--message-file` 로 제공해 커밋 전 임시 메시지를 판정하고, 준비된 commit-msg 훅 파일을 설치 폴더에 두어 사람이 한 줄로 설치할 수 있게 해요. local_check·CI 배선과 시작 커밋 기록은 이 조각 뒤 리더 몫이에요.

수리 조각(2026-11-21, 감사 B6 로 직접 겪은 사고를 기계로 잡는 남은 절반): 자동 설치. project-update 의 git-hook 스텝이 사용자 프로젝트의 훅 폴더에 훅을 설치하고(훅이 없을 때만), 훅 스크립트는 설치된 moai 가 없거나 구버전이라 `--message-file` 을 모를 때 커밋을 막지 않아요 — 제거 환경에서 커밋이 막히는 사고는 이 검사의 목적 밖이에요. 이 저장소 자체의 local_check·CI 배선과 시작 커밋 기록은 여전히 리더 몫이에요.

## Non-goals

- local_check·CI 의 배선과 시작 커밋 기록: 리더가 기록한 뒤 별도로.
- 이미 훅이 있을 때 그 내용을 바꾸는 것: 사용자 훅은 건드리지 않고 needs_decision 으로 돌려요.
- .git_hooks/ 의 Go 세대 파일: 보존 폴더라 손대지 않아요.
- Out of scope: 네트워크, 카드 번호의 출력 사용(검사가 거절 문구에 번호 자체를 이름하는 것은 예외 — 번호는 비밀이 아니에요).

## Requirements

- REQ-1: When trailer-check judges a range with a recorded start (`--card-numbers-since`), it shall report a failing `card_number` finding for every commit after that start whose subject or body carries the card-number shape `C-<n>` or `c-<n>`, with the named start commit itself exempt. · failure: none — REQ-2 keeps the rule off while no start is recorded.
- REQ-2: While the flag is absent, the judgement shall leave every commit's card-number shape unreported, so existing ranges answer exactly as before this change.
- REQ-3 (REQ-1): IF the shape is preceded on its line by an odd number of double quotes, THEN it shall be exempt as quoted ledger output, along with the neighbour shapes `AC-<n>` and any version or hexadecimal SHA.
- REQ-4: While `--message-file` names a draft message file, the command shall judge that one message with the same scan always active, printing `ok` or failing findings, and exit 1 when a card number is found.
- REQ-5: A ready commit-msg hook file shall sit beside the installer scripts, calling the same check on the message file git hands the hook.
- REQ-6: While `moai project-update` plans a project that is a git repository, the git-hook step shall offer installing the hook file into that project's hook folder — the folder `git rev-parse --git-path hooks` names, so a `core.hooksPath` setting is honored — when the folder holds no `commit-msg` hook, keep the step `unchanged` when a hook with the same bytes is already there, and answer `needs_decision` when a different hook file is present, while a project without a `.git` folder gets no such step. · failure: none — pinned by AC-6.
- REQ-7: The hook file shall pass a clean message, refuse the commit with exit 1 and the check's findings when the scan finds a card number, and skip with a warning instead of refusing when the installed moai is missing or its `trailer-check` answers a usage error, so an uninstall or an older install never blocks a commit. · failure: none — pinned by AC-7.

## Acceptance

- AC-1 (REQ-1): a fixture commit whose body says `wording after C-150` fails with a `card_number` finding once `--card-numbers-since` names an earlier commit; the start commit itself with the same shape passes.
  verify: `cargo test --locked --offline --test trailer_check_cli`
- AC-2 (REQ-3): messages naming a SPEC id, a version and a hexadecimal SHA, an `AC-<n>` neighbour, and a quoted ledger refusal all pass.
  verify: `cargo test --locked --offline --test trailer_check_cli`
- AC-3 (REQ-2): without the flag the same card-number commit passes and `card_numbers_from` is null.
  verify: `cargo test --locked --offline --test trailer_check_cli`
- AC-4 (REQ-4, REQ-5): `--message-file` fails a draft carrying a card number and passes a clean one; the hook file calls the same command.
  verify: `cargo test --locked --offline --test trailer_check_cli`
- AC-6 (REQ-6): in a temporary git project `project-update` plans the git-hook step `ready`, its apply installs the hook into the hook folder executable, a second plan answers `unchanged`, a project holding a different `commit-msg` answers `needs_decision`, and a non-git folder has no such step.
  verify: `cargo test --locked --offline --test git_hook_cli`
- AC-7 (REQ-7): with a stub moai the hook passes a clean message (exit 0), refuses a card-number message (exit 1, findings on stderr), and — with a stub that answers 2 or nothing on the PATH — skips with a warning (exit 0).
  verify: `cargo test --locked --offline --test git_hook_cli`
- AC-5: the neighbours that read the trailer module and the un-narrowed READY set pass.
  verify: `cargo test --locked --offline --test trailer_verified_cli`
  verify: `cargo test --locked --offline --test bundle_check_cli`
  verify: `cargo test --locked --offline --test bundle_verify_cli`

## Scope

- write: `.moai/specs/SPEC-20261120-commit-message-card-numbers/spec.md`, `src/trailers.rs`, `src/ops/trailer_check.rs`, `tests/trailer_check_cli.rs`, `install/commit-msg`; 수리 조각이 더하는 것: `src/git_hooks.rs`, `src/ops/git_hook_apply.rs`, `src/project_update.rs`(git-hook 스텝), `tests/git_hook_cli.rs`, `docs/design-review/appendix/operations/git_hook_apply.md`, `src/help.rs`(새 명령 등록)
- read-only: `src/hooks.rs`, `.git_hooks/`
- forbidden: 실제 HOME, 네트워크, 이미 있는 사용자 훅의 교체, local_check·CI 배선(리더 몫)

## Plan

1. RED: trailer_check_cli 에 카드 번호 장면(실패·면제·꺼짐·message-file)을 새 기대로 두고 지금 코드로 깨져요.
2. GREEN: log 형식에 제목·본문을 더하고(parse_log 6필드), 스캔·면제·시작 규칙·message-file 모드를 trailers.rs 에, 플래그를 ops 에 넣어요; install/commit-msg 훅 파일을 둬요.
3. 변이(꼴 경계, 따옴표 면제, 시작 면제, message-file 판정) → 비축소 READY 검사(02:36Z 규칙) → tier 1 감사 → READY.
4. 수리 조각 1: install/commit-msg 의 skip 경계(없는 moai·usage 오류는 경고 후 통과)와 시험(git_hook_cli 에 스텁 moai 장면).
5. 수리 조각 2: src/git_hooks.rs(훅 폴더·기존 훅 판정 — git 스폰은 scrub_git_env 경유)와 ops git-hook-apply(설치·receipt), OPERATIONS 조각·help 등록.
6. 수리 조각 3: project-update 의 git-hook 스텝(plan→apply argv)과 시험, clippy·fmt·bins 전체.

## Risks

- log 형식 확장이 bundle·verify 같은 trailer-check 사용처에 닿아요: 결과 JSON 은 키 추가뿐(기존 시험은 필드별 단정), bundle 대상을 READY 검사에 넣어 확인해요.
- 꼴 경계가 오탐을 낼 수 있어요: 앞·뒤 경계와 인용 면제를 시험으로 고정하고, 거절 문구는 번호와 줄 위치만 이름해요.
- 훅이 커밋을 막는 도구라 fail-closed 실수가 치명적이에요: skip 경계(없는 moai·usage 오류)를 시험으로 고정하고, 훅이 없는 폴더에는 늘 사용자 훅이 이기요(덮지 않음).
- git 스폰의 환경은 scrub_git_env 를 지나요: git_env_scrub 감사 시험이 새 스폰을 잡게 두고, 시험의 git fixture 는 tests/common/git.rs 의 가드(gc.auto 등)를 따라요.
