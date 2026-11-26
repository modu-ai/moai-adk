---
id: SPEC-20261125-session-names-project-prefix
title: "Session names carry the project: <project>-leader and <project>-agent-<english name>, with a registry check before a message is sent"
status: approved
revision: 1
created: 2026-11-25
branch: spec/session-names-project-prefix-20261125
approval:
  by: user
  at: 2026-11-25
  basis: "사용자가 리더 대화의 선택지에서 직접 고른 답(카드 노트 9133): \"승인, D-2~D-9 기본값 A (Recommended)\"; 리더가 기록 직전에 moai spec show 로 digest 를 다시 읽었어요"
  digest: "sha256:b757e1e5a36b401a0ee6be96d5386cfb3b9f1aa053e032c035864894224325c3"
---

# Session names carry the project: <project>-leader and <project>-agent-<english name>, with a registry check before a message is sent

## Interview

- Q1: 세션 이름에 프로젝트를 어떻게 넣나요? → A: "<프로젝트>-agent-<이름>"(사용자 2026-11-25T08:17:03Z, 선택 그대로; 미리보기 moai-adk-leader / moai-adk-agent-ben / finpc-leader / finpc-agent-ben, 카드 본문)
- Q2: 왜 필요한가요? → A: "moai cc -a/-l 의 이름 정의와 어떤 프로젝트에 어떤것인지 정의를 해야 … 프로젝트 마다 혼동이 없도록 처리를 하자"(사용자, 카드 본문이 옮긴 말)

## Intent

세션 이름은 크로스 세션 메시지의 주소예요. 지금 이름 `leader-<project>`, `agent-<english name>` 은 레인 이름에 프로젝트가 없어서, 2026-11-25 에 리더가 FINPC 저장소의 `agent-ben`, `agent-cara` 에 moai-adk 감사 지시를 잘못 보냈어요. 이 SPEC 은 이름을 `<project>-leader`, `<project>-agent-<english name>` 으로 바꿔서 이름만 봐도 어느 프로젝트의 세션인지 알게 하고, 리더가 메시지를 보내기 전에 받는 세션이 이 프로젝트 등록부의 레인인지 한 단계로 확인하게 하고, 레인이 일을 시작하기 전에 자기 저장소와 프로젝트 이름을 확인하게 해요. 이 SPEC 은 `SPEC-20261125-agent-team-session-names` 의 이름 규칙(REQ-1, REQ-3, 결정 D-1, D-9)을 대체해요.

읽어서 확인한 사실(2026-11-25, 이 워크트리 NeXT `3dfd96b7d8`):

- 이름을 만드는 코드는 `src/session.rs` 에 있어요: `project_name`(`project_id`, 없으면 체크아웃 폴더 이름을 레지스트리 문자로 고치고 `MAX_NAME_LEN` 64 에서 `leader-` 길이를 빼서 줄여요), `leader_session_name`, `pick_agent_name`(영어이름 24개 목록, `.live` 잠금과 등록부 항목을 건너뜀)이고 `src/main.rs` 가 `-l`/`-a` 에서 불러요. 이름이 `agent-` 로 시작한다는 가정은 `agent_number_of` 에도 있어요(`agent-<n>` 번호 항목 읽기, 옛 번호 이름용).
- 등록부와 `.live` 잠금은 클론 하나마다 `<clone common dir>/.moai-cli/sessions/` 에 있어요(`session::agent_seq_dir`). 그래서 같은 클론 안에서는 이름이 겹치지 않지만, 다른 저장소(FINPC)의 같은 이름 세션은 이 등록부에 없고 `ListAgents` 에는 함께 보여요. 이 실측: 이 클론의 등록부에는 `agent-ben.live`, `agent-cara.live`, `agent-dana.live`, `leader.json`, `leader.live` 가 있고 `ListAgents` 에는 `agent-ben`, `agent-dana` 가 두 줄씩 있어요. 하나는 다른 저장소 세션이라고 짐작되지만 읽어서 확인하지는 못했어요(짐작).
- 등록부 항목 `leader.json` 에는 호스트 세션 id 가 있어요(`session-register --show --name leader` 의 `session`). 그런데 이 id(`b3b7f70a-…`)의 앞 6자와 `ListAgents` 의 `[ref]`(예: `leader-moai-adk [1dbde7]`)가 같아 보이지 않아요. 지금 글(`team-messages.md` "Addressing")은 "세션 id 를 `ListAgents` 행에 맞춰라"라고 하는데 어떻게 맞추는지 확인하지 못했어요(Risks R-3, 결정 D-4).
- `agent` 이름의 `.live` 잠금은 런처가 `-a`/`--name` 으로 띄울 때만 잡혀요(`src/main.rs` 의 `lock_named_live`). 옵션 없이 띄우고 `/rename` 한 세션은 잠금도 항목도 없어서 등록부가 알지 못해요. `/moai agent` 는 지금 레인을 등록부에 등록하지 않아요(`session-register --mode agent` 호출이 스킬 글에 없어요).
- `moai session-register --show --name <이름>` 은 항목만 읽어서 `.live` 잠금만 있는 레인은 `free` 로 답해요(이 세션의 `agent-cara`, `agent-ben`, `agent-dana` 가 `free` 로 읽혔어요, 2026-11-25). 그래서 지금 명령으로는 "이 이름이 이 프로젝트의 레인인가"를 읽을 수 없어요.
- 프로젝트 프로필 `.moai/project.yaml` 에는 이름 키가 없고 `project_id` 만 있어요(이 저장소는 `moai-adk`). `project_id` 는 사람이 정한 값(`demo`)일 수도 있고 무작위 26자(`project_id::is_random_form`)일 수도 있어요.

읽는 사람: 이 변경을 승인하는 사용자, 구현하는 레인, 독립 감사자.

## Non-goals

- 이미 떠 있는 옛 이름 세션(`leader-moai-adk`, `agent-ben` 등)을 모델이나 런처가 바꾸지 않아요. 이름 변경은 사용자의 `/rename` 이나 재시작만 해요(모델은 `/rename` 을 실행할 수 없어요).
- 등록부의 리더 항목 이름 `leader`(`session-register --name leader`)는 바뀌지 않아요. 표시 이름과 등록부 이름은 따로예요.
- 영어이름 목록(24개), 우회·advisor·모델 규칙, 첫 입력, `moai cc|glm` 옵션 없는 launch 는 바뀌지 않아요.
- 크로스 세션 메시지를 보내는 일 자체(호스트 도구)나 Orca 메시지는 새로 만들지 않아요. 확인 명령은 읽기만 해요.
- 호스트의 `ListAgents` 가 다른 저장소의 세션을 숨기게 하는 일은 하지 않아요(호스트 기능).
- `interview.md`, `project.md`, 전역 설치본 `~/.claude/skills/moai`(착지 뒤 리더가 `moai update` 로 갱신해요), 실제 HOME, 네트워크, Python 은 쓰지 않아요.

## Requirements

- REQ-1: When `moai cc` or `moai glm` is given `-l` or `--leader` without `--name`, the launcher shall start the host with the session name `<project name>-leader`. · failure: none — the name is composed from the project name, and REQ-3 covers a name that is already taken.
- REQ-2: When `moai cc` or `moai glm` is given `-a` or `--agent` without `--name`, the launcher shall start the host with the session name `<project name>-agent-<english name>`, the first name of the launcher's list that REQ-3 does not exclude. · failure: none — a list that is used up is refused with the message that names `--name`.
- REQ-3: When a `<project name>-agent-<english name>` has a held session live lock or a registry entry with a running lease, the launcher shall skip that name. · failure: none — a list that is used up is refused with the message that names `--name`.
- REQ-15 (REQ-1): IF the live lock of `<project name>-leader` is held, THEN the launcher shall refuse `-l` with one message and start no host.
- REQ-4: The launcher shall compose the project name from the `project_id` of the project profile when it is not of the random 26-character form, else from the folder name of the clone's primary checkout, reduced to the characters of the registry's session names and cut so that the whole session name fits the registry's name length. · failure: none — an empty result is the word `project`.
- REQ-5: When `--name <name>` is given with `-l` or `-a`, the launcher shall use that name as given. · failure: none — a name the registry refuses is refused as it is today.
- REQ-6: When the launcher help or a repository document spells the session names, it shall spell `<project name>-leader` and `<project name>-agent-<english name>`. · failure: none — the help golden files and the repository scan tests pin the text.
- REQ-7: When a session name is checked, a read-only command of `moai session-register` shall answer whether that name is a lane of this clone's registry. · failure: none — the answer is `not_lane` with the reason when the registry is unreadable or has no such name, never `lane`.
- REQ-8 (REQ-7): IF the name's live lock is not held and its registry entry has no running lease, THEN the command shall answer `not_lane`. · failure: none — a symlinked sessions folder or an unreadable entry answers `not_lane` with the reason, and the command writes nothing.
- REQ-9: When the leader is about to send a message to a session other than the registered leader, the leader workflow shall tell the leader to run that command with the session's listed name first and to send only on `lane`. · failure: none — on `not_lane` nothing is sent and the leader writes one ledger note that names the row and the answer.
- REQ-10 (REQ-9): When two rows of `ListAgents` carry the same name, the leader workflow shall send to neither and write one ledger note, because the registry holds one entry per name. · failure: none — the note stands and the lane's own ledger notes reach the leader.
- REQ-11: When a session enters `/moai agent` or `/moai leader`, the workflow shall tell it to run `pwd` and read the project name before any other action. · failure: none — a session whose folder is not the project that a dispatch or the ledger names says `another repository` in one line and takes and writes nothing.
- REQ-12 (REQ-11): The read-only command of REQ-7 shall also print the project name the launcher uses and the leader session name and the agent name prefix of this project. · failure: none — outside a git repository it answers an error that says so.
- REQ-13: When `/moai agent` sets the mode in a session whose name is not already `<project name>-agent-<english name>`, the agent workflow shall tell the session to give the user the one line `/rename <project name>-agent-<english name>` once, and `/moai leader` the line `/rename <project name>-leader`. · failure: none — the fallback of the earlier SPEC (REQ-17: go on under the name the session has, short id, one ledger note) stands.
- REQ-14: When this change lands, the decisions log shall carry one line that replaces the 2026-11-25 shapes `leader-<project>` and `agent-<name>` with the shapes of REQ-1 and REQ-2. · failure: none — `CLAUDE.md`, the skill text and the tests that pin the old shapes carry the new ones, and the repository scan tests read these files.

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-4, REQ-5): `moai cc -l --dry-run` and `moai cc -a --dry-run` in a temporary project whose profile has `project_id` `demo` print `--name demo-leader` and `--name demo-agent-ben`; a random-form `project_id` uses the folder name; `--name` replaces the name; a name longer than the registry allows is cut to fit · verify: `cargo test --offline --locked --test launcher_role_flags ac1_`
- AC-2 (REQ-2, REQ-3, REQ-15): `-a` skips `demo-agent-ben` while its live lock is held or its entry has a running lease, `-l` is refused while `demo-leader` is held, and a used-up list names `--name` · verify: `cargo test --offline --locked --test launcher_role_flags ac2_`
- AC-3 (REQ-7, REQ-8, REQ-12): the command answers `lane` for a held live lock and for a running lease, `not_lane` for a name with neither, for an unreadable entry and for a symlinked folder, prints the project name and the two names of this project, and leaves the sessions folder byte-identical · verify: `cargo test --offline --locked --test session_lane_check_cli`
- AC-4 (REQ-6, REQ-14): the help golden files and `CLAUDE.md` carry the new shapes, no live text spells `leader-<project name>` or `agent-<english name>` as the launcher's name, and the decisions log has the replacing line · verify: `cargo test --offline --locked --test launcher_role_flags ac6_` and `cargo test --offline --locked --test help_layout` and `cargo test --offline --locked --test repo_claude_md`
- AC-5 (REQ-9, REQ-10, REQ-11, REQ-13): the leader workflow names the check before every send and the two-rows rule, the agent workflow and the leader workflow carry the first action (`pwd` and the project name) and the `/rename` line with the new shapes, and `team-messages.md` carries the same check in "Addressing" · verify: `cargo test --offline --locked --test lane_team_names`

## Decisions

각 결정은 되돌릴 수 있고 분류는 none 이에요. 추천이 맨 앞이고, 기본값은 승인 때 사용자가 다르게 고르지 않으면 적용돼요.

- D-1 (이름 모양, 사용자 결정 확정): `<project name>-leader`, `<project name>-agent-<english name>` 예 `moai-adk-leader`, `moai-adk-agent-ben`, `finpc-leader`, `finpc-agent-ben`. 영어이름 목록과 순서는 그대로예요.
- D-2 (프로젝트 이름의 출처): (A, 추천) `project_id` 가 사람이 정한 값이면 그것, 무작위 26자 형태면 클론의 기본 체크아웃 폴더 이름. 워크트리 폴더 이름은 쓰지 않아요(카드 워크트리는 일마다 이름이 달라서 같은 프로젝트의 세션이 다른 접두사를 가져요). (B) 항상 폴더 이름: 폴더를 옮기거나 클론 이름이 다르면 이름이 바뀌어요. (C) 항상 `project_id`: 무작위 id 의 프로젝트는 26자 접두사가 붙어 읽기 어려워요. 기본값은 A.
- D-3 (긴 이름과 특수문자): (A, 추천) 레지스트리 문자(영문·숫자·`.`·`_`·`-`) 밖은 `-` 로 고치고, 첫 글자가 영문·숫자가 되게 앞의 기호를 떼고, 이름 전체가 64자(`MAX_NAME_LEN`)를 넘지 않게 프로젝트 이름을 줄여요(가장 긴 접미사 `-agent-<영어이름>` 의 길이를 빼고). 결과가 비면 `project`. 같은 규칙을 지금의 `project_name` 이 이미 써요. (B) 넘으면 `-l`/`-a` 를 거절하고 `--name` 을 안내해요. 기본값은 A.
- D-4 (보내기 전 확인 명령과 모양): (A, 추천) `moai session-register` 에 읽기 전용 `--lane-check <listed name>` 을 더해요. 답은 `{"status":"lane"|"not_lane","name":…,"basis":"live_lock"|"registry_lease"|…,"reason":…}` 이고 `lane` 은 `.live` 잠금이 잡혔거나 등록부 항목의 임대가 이어질 때만이에요. 같은 명령 `--names` 는 이 프로젝트의 프로젝트 이름, 리더 이름, 에이전트 접두사를 출력해요(REQ-12). (B) 기존 `--show` 를 늘려서 `live` 필드를 더해요: 명령이 하나지만 `--show` 가 지금 항목만 말하는 계약이 바뀌어요. (C) 새 최상위 명령 `moai lane-check`: 이름이 눈에 띄지만 명령이 하나 더 늘어요. 호스트 `[ref]` 와 등록부 세션 id 가 맞는지는 확인되지 않아서(Intent 의 사실 셋째) 이름으로만 확인해요. 기본값은 A.
- D-5 (같은 이름이 두 줄일 때): (A, 추천) 어느 쪽에도 보내지 않고 원장에 한 줄을 남겨요(REQ-10). 런처가 한 클론에서 같은 이름을 못 띄우게 막으므로 두 줄은 다른 저장소의 세션이거나 옵션 없이 띄우고 `/rename` 한 세션이에요. (B) 짧은 id 로 보내요: 틀린 쪽에 보낼 수 있어요(오늘 사고의 원인). 기본값은 A.
- D-6 (이미 떠 있는 옛 이름 세션): (A, 추천) 그대로 두고 `--lane-check` 가 같은 규칙으로 답해요: 옛 `agent-ben` 도 이 클론의 `.live` 잠금이 잡혀 있으면 `lane` 이라 지시를 받아요. 사용자가 다음 재시작이나 `/rename` 때 새 이름으로 바꿔요. 이름 모양을 접두사로 검사하지 않아요. (B) 새 모양이 아니면 `not_lane` 으로 취급해서 지시를 보내지 않아요: 이름 혼동은 막지만 지금 도는 세션들이 모두 재시작될 때까지 지시를 못 받아요. 기본값은 A.
- D-7 (`--name` 을 준 경우): (A, 추천) 준 이름을 그대로 써요(REQ-5): 접두사를 붙이지 않고 등록부의 이름 검사만 거쳐요. 사용자가 이름을 완전히 정하는 길이라서 프로젝트 이름이 없을 수 있고, 그 세션은 `--lane-check` 에서 이름이 같다면 `lane` 이에요. (B) 접두사가 없으면 거절해요: 혼동은 줄지만 `--name` 의 지금 의미가 바뀌어요. 기본값은 A.
- D-8 (옵션 없이 띄운 세션의 등록): (A, 추천) `/moai agent` 는 레인이 등록부에 항목이 없을 때 `moai session-register . --name <its name> --mode agent` 로 등록해요(`--show` 가 `free` 일 때만, 이미 `held` 이면 건드리지 않아요). 그래야 `/rename` 한 세션도 `lane` 으로 읽혀요. (B) 등록하지 않아요: 옵션 없이 띄운 세션은 `not_lane` 이라 지시를 받지 못해요. 기본값은 A.
- D-9 (SPEC 크기): (A, 추천) 한 SPEC 으로 가요. 요구 15개는 `spec check` 의 경고 기준(12개)을 넘으면 그대로 두어요. (B) 런처(REQ-1..6)와 글·명령(REQ-7..14)으로 쪼개요. 기본값은 A.

## Scope

- write: `.moai/specs/SPEC-20261125-session-names-project-prefix/**`, `src/session.rs`, `src/main.rs`, `src/help.rs`, `src/ops/session_register.rs`, `src/lane_registry.rs`, `skills/moai/workflows/work.md`, `skills/moai/workflows/leader.md`, `skills/moai/references/flows/team-messages.md`, `CLAUDE.md`, `docs/decisions-2026-11-08.md`, `docs/claude-md-rule-provenance.md`, `docs/rust-product-path.md`, `docs/user-entry-draft.md`, `docs/host-support-matrix.md`, `tests/launcher_role_flags.rs`, `tests/lane_team_names.rs`, `tests/session_lane_check_cli.rs`(새 파일, 첫 낱말 `session` 은 `tests/areas.toml` 에서 읽어서 확인해요)
- write (읽어서 정확한 경로를 SCOPE-REQUEST 로 올려요): 이름 모양을 고정한 기존 테스트와 골든 — `tests/help_layout.rs`, `tests/fixtures/help/*`, `tests/repo_claude_md.rs`, `tests/launcher_session_locks.rs`, `tests/launcher_retire_role_flags.rs`, `tests/product_cli.rs`
- write (한 줄, line-lease): `src/lib.rs`·`src/help.rs` 의 한 줄 등록, `tools/repo_scans.json`(새 테스트 파일 등록), `skills/moai/references/source-manifest.json`
- read-only: `src/launcher_init.rs`, `src/project_id.rs`, `src/identity.rs`, `src/clone_root.rs`
- forbidden: `interview.md`, `project.md`, 전역 설치본과 실제 HOME, 네트워크, Python, 보호 훅을 약하게 하는 변경

## Plan

1. RED: `tests/launcher_role_flags.rs` 의 `ac1_`/`ac2_`/`ac6_` 와 `tests/session_lane_check_cli.rs`, `tests/lane_team_names.rs` 가 현재 동작에서 실패해요. 실패를 보고 나서 구현해요.
2. 런처 이름: `session::project_name` 이 기본 체크아웃 폴더와 무작위 id 규칙(D-2)을 따르고 접미사 길이를 빼서 줄이게(D-3), `leader_session_name` 과 `pick_agent_name` 의 모양을 `<project>-leader`, `<project>-agent-<name>` 으로 바꿔요. GREEN.
3. 확인 명령: `lane_registry` 에 읽기 전용 질의(`.live` 잠금, 임대, 읽을 수 없는 항목, 링크된 폴더는 `not_lane`)를 더하고 `session-register --lane-check`, `--names` 로 연결해요(D-4).
4. 글: `leader.md`(보내기 전 확인 단계, 같은 이름 두 줄 규칙), `work.md`(첫 행동 `pwd`·프로젝트 이름, `/rename` 줄, 등록 D-8), `team-messages.md` 의 Addressing 을 새 이름과 확인 단계로 고쳐요.
5. 문서: `CLAUDE.md` 의 세션 이름 문장, 결정 로그의 대체 줄, 출처 표, 세 `docs/**` 문장.
6. 위험 경로(이름·확인 명령은 읽기만 하고 삭제·권한을 건드리지 않아요): 임시 클론에서 `.live` 잠금을 잡은 `-a` 세션의 이름이 `--lane-check` 로 `lane` 으로 읽히고, 잠금이 풀린 뒤 `not_lane` 으로 읽히는 것을 실제 실행으로 READY 에 붙여요.
7. 가드 뮤턴트(tier 2: 런처, `CLAUDE.md`): 새 가드 테스트마다 `guard-mutant:` 한 줄. 영수증은 `moai check affected --card` 한 번, READY 에 SPEC digest 를 담아요.

## Risks

- R-1: 이름이 길어져서(`<project>-agent-<name>`) 프로젝트 이름이 긴 저장소는 잘려요(D-3). 잘린 이름이 다른 프로젝트와 같아질 확률은 낮지만 0 은 아니에요. 같은 이름이면 REQ-3 의 잠금이 한 클론 안에서 막고, 다른 클론이면 REQ-9 의 확인이 막아요.
- R-2: 옛 이름 세션과 새 이름 세션이 한동안 함께 떠요(D-6). 리더가 둘 다 `--lane-check` 로 같은 규칙으로 읽어요.
- R-3: 등록부의 세션 id 와 `ListAgents` 의 `[ref]` 가 맞는지 확인되지 않았어요(Intent 의 사실 셋째). 이 SPEC 은 이름으로만 확인하고 id 맞추기에 기대지 않아요. 구현 레인이 실제로 맞는지 확인하면 `team-messages.md` 의 "Addressing" 문장을 따로 고쳐야 해요.
- R-4: 옵션 없이 띄운 세션이 등록 전에는 `not_lane` 이라 첫 지시를 받지 못해요(D-8 이 등록으로 줄여요).
- R-5: `CLAUDE.md` 와 런처를 바꾸는 tier 2 라서 사용자 승인이 먼저예요. 승인 전에는 구현하지 않아요.
- R-6: `--lane-check` 가 `lane` 이라고 답해도 그 세션이 지금 일을 받을 수 있는 상태인지는 말하지 않아요. 상태(바쁨, 쉼)는 `ListAgents` 가 말해요.
