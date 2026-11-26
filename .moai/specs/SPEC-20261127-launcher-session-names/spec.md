---
id: SPEC-20261127-launcher-session-names
title: "Launcher names lanes agent-<name>, unique on the whole machine, and the leader leader-<project name>"
status: draft
revision: 2
created: 2026-11-27
---

# Launcher names lanes agent-<name>, unique on the whole machine, and the leader leader-<project name>

## Interview

- Q1: 세션 이름에서 프로젝트 접두사를 빼도 되나요? → A: "agent-이름, leader-이름으로 각 이름은 다른 프로젝트에서도 orca에서 오인이 없도록 이름이 겹쳐지지 않도록만 하면 앞에 프로젝트명을 추가하지 않아도 되장 낳을까? 그렇다면 이렇게 이름을 작성하도록 moai cc/glm 론처 수정해라."(사용자, 2026-11-27, 카드 C-587 본문이 옮긴 말 그대로)
- Q2: 리더 이름은 무엇인가요? → A: "leader-<프로젝트명> (Recommended)"(사용자의 선택, 2026-11-27, 같은 리더 세션의 질문 도구; 예 `leader-moai-adk`). 레인은 `agent-<이름>` 이에요.
- Q3: CLAUDE.md 의 세션 이름 문장을 이 변경으로 고쳐도 되나요? → A: 아직 묻지 않았어요. 동료 세션의 메시지는 승인이 아니라서(C-584 노트 10450, 10452 의 선례) SPEC-READY 에 쓸 문장을 인용해 사용자 본인의 선택을 받아요(D-6).

## Intent

세션 이름은 크로스 세션 메시지의 주소예요. 2026-11-25 의 `SPEC-20261125-session-names-project-prefix`(C-531)는 오인을 막으려고 이름 앞에 프로젝트를 붙였어요(`<project>-leader`, `<project>-agent-<name>`). 사용자는 2026-11-27 에 방향을 바꿨어요. 이름에 프로젝트를 넣지 않아도 되고, 대신 **이 컴퓨터의 어느 프로젝트에서도** 살아 있는 세션의 이름과 겹치지 않기만 하면 돼요. 리더는 `leader-<프로젝트 이름>` 이에요. 이 SPEC 은 그 방향의 구현이에요.

지금 겹침을 보는 범위는 클론 하나예요. 이름 잠금(`<name>.live`)과 등록부가 `<clone common dir>/.moai-cli/sessions/` 에 있어서(`session::agent_seq_dir`), 다른 저장소의 런처는 이 잠금을 볼 수 없어요. 그래서 프로젝트 접두사를 빼려면 HOME 단위 기록이 먼저 있어야 해요.

읽어서 확인한 사실(2026-11-27, NeXT `2e97d012e5`):

- 이름을 만드는 코드는 둘뿐이에요. `src/session.rs`(`project_name` 은 `MAX_NAME_LEN` 64 에서 `-agent-` 와 가장 긴 영어 이름을 빼서 줄여요, `leader_session_name`, `agent_prefix`, `pick_agent_name_seeded`)와 `src/card/wip_limit.rs:87-96`(`live_lanes` 가 `<project>-leader` 와 `<project>-agent-` 를 같은 식으로 다시 만들고 `canon` 으로 접두사 있는 이름을 짧은 이름으로 읽어요). 이름이 `agent-` 로 시작하는지, `-leader` 로 끝나는지를 보고 역할을 짐작하는 곳은 더 찾지 못했어요(`strip_prefix("agent-")` 는 `agent-<번호>` 항목용 `session::agent_number_of` 뿐이에요).
- `pick_agent_name_seeded` 는 이미 `agent-<n>` 과 `<project>-agent-<n>` 두 모양을 다 "쓰는 중"으로 읽어요. 이름 잠금은 호스트가 도는 동안 런처가 쥐는 `flock` 이에요(`session::lock_named_live`, 열린 파일 기술자라서 호스트가 끝나거나 런처가 죽으면 OS 가 풀어요). `session_launch` 는 호스트를 `.status()` 로 띄우고 기다려요.
- `session-register --lane-check <이름>` 은 이름의 접두사를 보지 않고 이 클론의 잠금과 등록부만 봐요(`lane_registry::lane_check`). 그래서 새 모양에서도 다른 프로젝트의 레인은 `not_lane` 으로 거절돼요.
- `moai card inbox` 의 `--leader` 기본값 `leader` 는 리더 노트의 작성자 이름표이고 세션 이름이 아니에요(`src/card/inbox.rs:20`). 이 SPEC 은 그것을 바꾸지 않아요.
- HOME 상태 폴더는 `<HOME>/.moai`(`home_workdir::root`)이고, HOME 은 `home_workdir::home()` 이 `HOME`(Windows 는 `USERPROFILE`)에서 읽어요.

읽는 사람: 이 변경을 승인하는 사용자, 구현하는 레인, 독립 감사자.

이 SPEC 이 `SPEC-20261125-session-names-project-prefix` 에서 대체하는 것: REQ-1, REQ-2, REQ-3(이름 모양과 겹침 범위), REQ-4 의 줄이는 길이(접미사 계산), REQ-6, REQ-12 의 출력, REQ-13 의 `/rename` 줄, REQ-14 의 결정 로그 줄, D-1, D-3 의 줄이기 계산, AC-1, AC-2, AC-4. 그 SPEC 의 REQ-7~REQ-11, REQ-15 와 D-4~D-9(보내기 전 확인, 같은 이름 두 줄 규칙, 첫 행동, `--lane-check`)는 그대로예요.

## Non-goals

- 이미 떠 있는 세션의 이름을 모델이나 런처가 바꾸지 않아요. 옛 이름(`<project>-leader`, `<project>-agent-<name>`, 번호 없는 `agent-<name>`)은 사용자의 재시작이나 `/rename` 때 바뀌고, 그동안 두 모양을 다 읽어요.
- `card inbox` 와 리더 노트의 작성자 이름표 `leader` 는 바뀌지 않아요(위 사실).
- 등록부의 리더 항목 이름 `leader`(`session-register --name leader`), 영어 이름 목록(48개)과 고르는 방식, 우회·advisor·모델 규칙, 첫 입력은 바뀌지 않아요. advisor 쪽 변경은 C-584 의 몫이에요.
- 역할 플래그 없는 `moai cc|glm --name <이름>` 은 HOME 잠금을 잡지 않아요(D-3).
- 호스트의 `ListAgents`, Orca, 이미 실행 중이거나 옛 런처로 띄운 다른 저장소의 세션을 이 기록이 알게 하는 일은 하지 않아요(Risks R-1).
- 실제 HOME 과 `~/.claude` 는 쓰지 않아요. 테스트는 임시 HOME 에서만 돌아요.

## Requirements

- REQ-1: When `moai cc` or `moai glm` is given `-a` or `--agent` without `--name`, the launcher shall start the host with the session name `agent-<name>`, where `<name>` is the first name that REQ-4 lets it claim, taken in the order of the launcher's seeded pick from the names of the list that no live session of this clone holds in either form (`agent-<name>` or `<project name>-agent-<name>`). · failure: none — a list that is used up is REQ-2.
- REQ-2 (REQ-1): IF no name of the list can be claimed, THEN the launcher shall start no host and refuse with one message that names `--name`.
- REQ-3: When `moai cc` or `moai glm` is given `-l` or `--leader` without `--name`, the launcher shall start the host with the session name `leader-<project name>`, where the project name is composed as it is today and cut so that the whole name fits the registry's 64 characters. · failure: none — an empty result is the word `project`.
- REQ-4: When the launcher starts a host under a name it chose, or under a `--name` given beside `-l` or `-a`, it shall first take the exclusive lock of the file `<HOME>/.moai/session-names/<name>.live`, then the clone's own `<name>.live` lock, hold both until the host has exited and release both if the second cannot be taken. · failure: none — REQ-5 and REQ-7 cover the refusals.
- REQ-5 (REQ-4): IF the machine-wide lock of a name is held by a launcher of any project, THEN the launcher shall, for `-a` without `--name`, go on to the next candidate, and for `-l` or a `--name` beside a role flag, start no host and refuse with one message that names the project the holder recorded and `--name`.
- REQ-6: When the launcher holds a machine-wide lock, it shall write the name, the project name, its process id and the time into the lock file after the lock is taken, as information only. · failure: none — a failed write leaves the lock held and the launch goes on.
- REQ-7 (REQ-4): IF the home folder cannot be resolved, or `<HOME>/.moai/session-names` or a lock file in it is a link, is not a plain folder or file, or cannot be opened, THEN the launcher shall start no host and refuse with the reason.
- REQ-8: When `--dry-run` is given with `-l` or `-a`, the launcher shall only probe the machine-wide locks, name the session it would start, and create no folder, file or lock. · failure: none — the probe answers `taken` for a lock it cannot judge.
- REQ-9: When `session-register --names` runs, it shall print the project name, the leader session name `leader-<project name>`, the lane name prefix `agent-` and, as `earlier`, the two earlier forms `<project name>-leader` and `<project name>-agent-` built with the earlier cut. · failure: none — outside a git repository it answers an error that says so.
- REQ-10: While a session is named in either the new or an earlier form, `session-register --lane-check` shall judge it by the clone's lock and registry only, so that a lane of another project, whatever its name, answers `not_lane`. · failure: none — the answer and its reasons stay as they are.
- REQ-11: When `card take` counts the live lanes, `wip_limit::live_lanes` shall read the leader as `leader-<project name>` or `<project name>-leader` and a lane as `agent-<name>` or `<project name>-agent-<name>`, and count a lane that holds both forms once. · failure: none — the earlier cut is used for the earlier forms.
- REQ-12: When a skill text, a help text or a repository document spells the launcher's session names, the document shall spell `leader-<project name>` and `agent-<name>`, the `/rename` line of the workflows included. · failure: none — the help golden files and the repository scan tests pin the text.

## Acceptance

- AC-1 (REQ-1, REQ-3, REQ-5): 임시 HOME 하나에 프로젝트 A·B 가 있고, A 의 가짜 호스트가 `agent-<x>` 를 쥐고 있어요. B 의 `-a` 는 `<x>` 를 고르는 시드에서도 다른 이름을 받고, 두 프로젝트에서 같은 이름이 동시에 나오지 않아요. 같은 프로젝트 이름이 `leader-<p>` 를 쥐면 두 번째 `-l` 은 거절돼요 · verify: `cargo test --locked --offline --test launcher_machine_names_cli collision_`
- AC-2 (REQ-2): 48개 이름이 모두 HOME 단위로 쥐어져 있으면 `-a` 는 호스트를 띄우지 않고 `--name` 을 말해요 · verify: `cargo test --locked --offline --test launcher_machine_names_cli used_up_`
- AC-3 (REQ-4, REQ-6): 가짜 호스트가 도는 동안 두 잠금이 잡혀 있고(다른 프로세스의 시도가 거절됨), 호스트가 끝나면 둘 다 풀려요. 클론 잠금을 못 잡으면 HOME 잠금도 풀려요. 잠금 파일에 이름·프로젝트·pid 가 적혀요 · verify: `cargo test --locked --offline --test launcher_machine_names_cli claim_`
- AC-4 (REQ-7, REQ-8): HOME 이 없거나 폴더·파일이 링크이면 호스트 없이 이유와 함께 거절되고, `--dry-run` 은 폴더와 파일을 만들지 않으며 트리가 그대로예요 · verify: `cargo test --locked --offline --test launcher_machine_names_cli refuse_` and `... dry_run_`
- AC-5 (REQ-5): 역할 플래그 옆의 `--name agent-ben` 이 다른 프로젝트에 쥐어져 있으면 그 프로젝트 이름과 `--name` 을 말하며 거절돼요. 역할 플래그 없는 `--name` 은 HOME 잠금 없이 전과 같이 돌아요 · verify: `cargo test --locked --offline --test launcher_machine_names_cli named_`
- AC-6 (REQ-9, REQ-10): `--names` 가 새 이름·접두사·`earlier` 를 내고, `--lane-check` 가 옛 모양 이름과 새 모양 이름 모두 이 클론의 잠금이 있을 때만 `lane` 이라고 하며 다른 프로젝트의 레인은 `not_lane` 이에요 · verify: `cargo test --locked --offline --test session_lane_check_cli`
- AC-7 (REQ-11): 카드 잡기의 레인 수가 `agent-ben` 과 `<p>-agent-ben` 을 한 레인으로 세고 옛·새 리더 이름을 레인으로 세지 않아요 · verify: `cargo test --locked --offline --test card card_wip_limit::`
- AC-8 (REQ-12): 도움말 골든, 스킬 글, 문서가 새 모양을 말하고 옛 모양을 런처 이름으로 말하는 곳이 없어요(CLAUDE.md 줄은 D-6 의 사용자 선택 뒤) · verify: `cargo test --locked --offline --test launcher_role_flags` and `cargo test --locked --offline --test help_layout` and `cargo test --locked --offline --test lane_team_names` and `cargo test --locked --offline --test repo_claude_md`

## Scope

- write: `.moai/specs/SPEC-20261127-launcher-session-names/**`, `src/session.rs`, `src/session/machine_names.rs`(new), `src/main.rs`(the `launch` name and lock blocks only), `src/card/wip_limit.rs`(`live_lanes` names only), `src/ops/session_register.rs`(`--names` only), `src/lane_registry.rs`(only if a doc line or name helper needs it), `skills/moai/workflows/work.md`, `skills/moai/workflows/leader.md`, `skills/moai/references/flows/team-messages.md`, `docs/user-entry-draft.md`, `docs/host-support-matrix.md`, `docs/claude-md-rule-provenance.md`, `docs/decisions-2026-11-08.md`, `docs/rust-product-path.md`, `tests/launcher_machine_names_cli.rs`(new), `tests/launcher_role_flags.rs`, `tests/launcher_retire_role_flags.rs`, `tests/lane_team_names.rs`, `tests/session_lane_check_cli.rs`, `tests/session_registry_owner_id.rs`, `tests/repo_claude_md.rs`
- write by line lease only, one line or hunk each, given back after the commit: `src/help.rs`(the lines that spell the names, near 1678-1687 and 1775), `tests/fixtures/help/*.txt`, `skills/moai/references/source-manifest.json`, `tests/areas.toml` and `tools/repo_scans.json`(the new test file), `CLAUDE.md`(the session-name sentence, only after the user's own pick, D-6)
- read-only: `src/home_workdir.rs`, `src/launcher_init.rs`, `src/project_id.rs`, `src/identity.rs`, `src/clone_root.rs`, `src/card/inbox.rs`
- forbidden: the real HOME and its `.claude` folder, any process kill outside a test's own child, a network call, Python, `CLAUDE.md` without the user's own pick, any weakening of a protection hook or of `--lane-check`

## Plan

1. RED: `tests/launcher_machine_names_cli.rs` (stub host from `tests/launcher_session_locks.rs`, two temp projects under one temp HOME, a seed that would pick the held name) and the changed pins in `tests/launcher_role_flags.rs`, `tests/lane_team_names.rs`, `tests/session_lane_check_cli.rs` fail on today's code; record why.
2. `src/session/machine_names.rs`: `claim(home, name, project)` (plain folder and file checks as `refuse_unless_regular`, exclusive `try_lock`, body written after the lock), `probe(home, name)` (shared `try_lock`, no write, no folder), and the message for a taken name. `session.rs` declares the module and changes `project_name` (cut 57 for `leader-`), `leader_session_name`, `agent_prefix` and the pick so that it returns the name together with its held locks.
3. `src/main.rs` `launch`: resolve HOME once with `home_workdir::home()` and pass it down; claim machine lock then clone lock; hold both as `live_named` is held; dry run only probes.
4. `wip_limit::live_lanes` and `session-register --names` read both forms with the earlier cut for the earlier ones.
5. Texts: `work.md:9`, `leader.md:7`, `team-messages.md`, the docs and the help lines; `CLAUDE.md` waits for the user's pick.
6. Dangerous path (the launcher starts a process and writes a lock file under HOME): the real binary runs only as Rust integration tests with a stub host in a temp HOME (decision 10013). One `guard-mutant:` line per new guard test (REQ-5, REQ-7, REQ-8, the clone-lock release), tier 2.
7. Receipt: `moai check affected --card <base>..<head>` once at READY, with the digest of this SPEC.

## Risks

- R-1 (how far "unique" reaches): the record knows only names claimed by launches of a binary that carries this change. It cannot see sessions that are already running, sessions an older binary started in another repository (the 2026-11-25 mix-up of `agent-ben` came from one), or plain sessions renamed by hand. `session-register --lane-check` stays the only proof that a lane belongs to this project; the new names do not replace it.
- R-2 (file locks): a HOME on a file system without working `flock` (some network mounts) cannot hold the lock; REQ-7 refuses when the lock cannot be taken, so the launcher says so instead of going on unchecked.
- R-3 (killed launcher): when the launcher is killed while its host lives on, the lock is released with it and the name looks free. The clone-level lock already has the same gap.
- R-4 (long project names): an earlier-form session of a project whose name was cut under the old rule is read with the earlier cut (REQ-9, REQ-11); a name cut under neither rule would be missed.
- R-5 (files stay): the lock files are never deleted; at most the 48 names plus one leader file per project name stay, a few lines each.
- R-6 (overlapping cards): C-584 (`launch`, `default_advisor`, the same `CLAUDE.md` sentence), C-582 (`wip_limit.rs`), C-585 (`work.md`) and C-299, C-409 (`src/main.rs`) touch the same files. Each overlap is split by function in a PAIR-NOTE; the card that lands later merges.
- R-7 (`CLAUDE.md`): tier 2 floor path. Not written without the user's own pick (D-6).

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 이름 기록은 `<HOME>/.moai/session-names/<name>.live` 파일의 `flock` 이에요(기본). 대안: 임대가 있는 JSON 파일(낡은 기록 정리와 경쟁을 따로 풀어야 해요). | 카드가 HOME 단위 잠금이나 임대를 말했어요. `flock` 은 원자적으로 잡히고 호스트가 끝나거나 런처가 죽으면 OS 가 풀어서 낡은 기록이 생기지 않아요. | 모듈을 지우면 클론 단위 잠금만 남아요. |
| D-2 | `-l` 도 HOME 에서 잡아요(기본). 대안: 레인만. | 같은 프로젝트 이름의 두 클론이나 같은 이름의 두 프로젝트가 같은 `leader-<p>` 를 쓰면 이 SPEC 이 막으려는 오인이 그대로예요. | 리더 이름은 클론 단위 잠금만 쓰게 해요. |
| D-3 | 역할 플래그 옆의 `--name` 은 HOME 에서 잡고, 역할 플래그 없는 `--name` 은 안 잡아요(기본). 대안: 모든 이름 있는 실행. | 역할 플래그가 이 SPEC 의 이름을 만들어요. 평범한 `--name` 은 사용자가 같은 이름을 여러 프로젝트에 일부러 쓸 수 있어서 동작을 바꾸지 않아요. | 조건 한 줄이에요. |
| D-4 | HOME 을 못 읽거나 폴더·파일이 쓸 수 없으면 거절해요(기본). 대안: 경고만 하고 확인 없이 진행. | 확인 없이 진행하면 "겹치지 않는다" 가 말뿐이 돼요. | 거절을 경고로 바꿔요. |
| D-5 | 옛 모양(`<p>-leader`, `<p>-agent-<name>`)은 리더 수와 레인 수를 셀 때 계속 읽어요(기본). 대안: 새 모양만. | 이미 떠 있는 세션은 옛 이름 그대로라 카드 잡기의 한도가 틀려져요. | 읽는 줄을 지워요. |
| D-6 | `CLAUDE.md` 의 세션 이름 문장은 사용자 본인의 선택을 받은 뒤에만 고쳐요(기본). 쓸 문장은 SPEC-READY 에 인용해요. 대안: 이번에는 고치지 않고 문서·스킬만. | 동료 세션의 메시지는 승인이 아니에요(선례 C-584 노트 10450·10452). 그 문장은 C-584 도 건드려서 마지막에 합쳐요. | 문장을 되돌려요. |
| D-7 | 기록 파일 본문은 이름·프로젝트 이름·pid·시각이에요(기본). 경로는 적지 않아요. 대안: 클론 경로 포함. | 거절 문장이 어느 프로젝트가 쥐었는지 말하면 충분하고, HOME 파일에 작업 폴더 경로를 남기지 않아요. | 필드를 줄이거나 늘려요. |
| D-8 | tier 2 로 봐요(런처가 프로세스를 띄우고 HOME 아래에 쓰며 `CLAUDE.md` 를 고쳐요). 대안: tier 1. | 바닥 경로 목록(`CLAUDE.md`)과 위험 경로 규칙(프로세스를 시작하는 변경)이에요. | 해당 없음. |
