---
id: SPEC-20261126-install-docs-match-code
title: "설치 문서(README, install/README)가 코드와 같은 말을 하고 Rust 시험이 그 어긋남을 막는다"
status: approved
revision: 1
created: 2026-11-26
approval:
  by: leader
  at: 2026-11-26
  basis: "리더 위임 승인 노트 9550 (사용자 위임 2026-11-25 ~14:50Z \"나머진 네가 알아서 추천 해주는데로 선택해서\", 리더 판단이고 사용자 직접 선택은 아님); 리더가 착지 직전 moai spec show 로 digest 를 다시 읽었어요"
  digest: "sha256:20063e2e7ff1cc63824d583e1634b1836fb9eee66c050755f9af88a8b2bee55b"
---

## Interview

- none: the documents are made to say what the code already does; the one open pick (note 9549) was decided by the leader under the delegation.

# 설치 문서(README, install/README)가 코드와 같은 말을 하고 Rust 시험이 그 어긋남을 막는다

## Intent

설치 흐름 보고서 §7 G10(2026-11-21, "코드가 이겨요")과 그 사용자 결정(2026-11-20T23:1xZ)을 다시 시작해요: README, install/README 가 코드와 다른 말을 하는 곳을 코드에 맞추고, 문서가 이름하는 명령·옵션·링크가 실제로 있는지 보는 시험을 더하고, 호스트 감지가 `CLAUDE_CONFIG_DIR` 을 따르게 해요.

이 SPEC 은 `SPEC-20261121-install-docs-match-code`(가지 `spec/install-docs-match-code-20261121` 에만 있는 옛 SPEC, 병합된 적 없어요)를 이어받지만 그대로 쓰지 않아요: 옛 SPEC 은 시험을 Python(`tests/docs_installer_paths_test.py`)으로 정했는데 이 저장소의 시험은 Rust 뿐이고, 그 사이 Codex 와 OMP 가 빠지고 `adapters/` 가 없어지고 설치 스크립트가 바뀌었어요. 모든 문장을 NeXT 끝 `e946ae8102` 에서 다시 읽어 확인했어요(2026-11-26).

읽은 어긋남(코드가 이겨요):

- `README.md` 의 "State folders" 칸은 프로젝트 `.moai-cli/`, 사용자 `~/.moai-cli/` 라고 하지만 코드는 새 프로젝트를 `.moai/` 로 시작해요(`state_root::create_dir`, 이전 RC 로 시작한 프로젝트는 자기 `.moai-cli/` 를 유지) 사용자 상태는 `~/.moai/`(마커가 있거나 상태 파일이 없는 HOME) 또는 이전 `~/.moai-cli/` 예요(`home_root::dir`). FAQ 두 곳도 같은 말을 해요.
- `README.md` 의 "Install" 은 설치 스크립트가 없다고 하지만 `install/install.sh` 와 `install/install.ps1` 이 있고, 주소는 `https://adk.mo.ai.kr/install.sh`, `install.ps1` 이에요. 다만 Rust 릴리스는 아직 없어서(`gh release list` 가 읽은 2026-11-26 의 릴리스는 Go `v3.x` 뿐이에요) 그 주소가 서빙하는 것은 아직 Go 쪽 스크립트예요(`install/README.md` 의 말). 그래서 "없다"가 아니라 "스크립트는 있고 릴리스 전까지 그 주소는 Go 판을 서빙해요"가 사실이에요.
- `README.md` 의 "First run" 은 `/moai read this project and set it up` 을 진입으로 안내하지만 설치 스크립트가 마지막에 출력하는 다음 단계는 `moai update` → 프로젝트 폴더에서 `moai cc` → 세션 안에서 `/moai project` 예요.
- `install/README.md` 의 "A pipe (`curl … | sh`) has no terminal on stdin, so it never shows the menu" 는 스크립트의 `answer_source` 와 달라요: 표준 입력이 터미널이 아니어도 `/dev/tty` 를 열 수 있으면 파이프에서도 메뉴가 떠요.
- `src/update_models.rs` 의 `installed_tools` 와 `install/install.sh` 의 `detect_hosts` 는 Claude 폴더를 `~/.claude` 로만 봐요. `src/host_home.rs` 의 resolver 는 `CLAUDE_CONFIG_DIR` 이 있으면 그 폴더를 따라요(두 코드가 다른 말을 해요).
- `adapters/orca.md` 는 이미 저장소에 없어요: 바꿀 것이 없어요.

## Non-goals

- README 의 소개 문장(Cargo 설명과 같아야 해요)과 Codex·GitHub Issues 단어: `SPEC-20261122-remove-codex-shipped-text` 의 S2 조각이 소유해요(README 와 Cargo 설명을 같은 커밋으로). 이 SPEC 은 그 문장들을 건드리지 않아요.
- `install/install.ps1` 의 Windows 질문 흐름과 `moai-cli` → `moai` 이름 바꾸기, 태그 글자: 다른 카드의 일이에요.
- `skills/moai/**` 와 `docs/**` 의 문장: 이 SPEC 의 시험은 README 둘과 install/README 만 읽어요.
- 설치 스크립트가 Rust 릴리스를 서빙하게 하는 일(release step)과 주소 바꾸기.
- 기존 Python 시험(`tests/install_script_test.py`)을 고치거나 새 Python 을 쓰는 일.

## Requirements

- REQ-1: When `README.md` or `README.ko.md` names the state folders, the text shall name the `.moai` folder for a new project, the kept `.moai-cli` folder for a project that began on the earlier RC, and for the user state the `.moai` folder of the HOME when it holds this generation's marker or no state files and the earlier `.moai-cli` folder otherwise, as `state_root` and `home_root` decide · failure: none — pinned by AC-1.
- REQ-2: When `README.md` or `README.ko.md` describes the installer, the text shall name the scripts the tree holds (`install/install.sh`, `install/install.ps1`) with their published addresses and one-line commands, link `install/README.md`, and say that no Rust release is published yet and that the addresses serve the Go edition's scripts until it is, without the words `not_run` or "not published" · failure: none — pinned by AC-2.
- REQ-3: When `README.md` or `README.ko.md` describes the first run, the text shall give the steps the installer prints at its end (`moai update` when the skill is not installed, then `moai cc` in a project folder, then `/moai project` in the session) before the natural-language draft, which stays a link · failure: none — pinned by AC-3.
- REQ-4: When `install/README.md` describes when the installer asks a question, the sentence shall say the answer comes from standard input when it is a terminal, else from the session's terminal (`/dev/tty`) when it can be opened, else nobody is asked, as `answer_source` of `install/install.sh` reads · failure: none — pinned by AC-4.
- REQ-5: When a code span of `README.md`, `README.ko.md` or `install/README.md` has the form `moai <word>` or `moai-cli <word>` with a lower-case word, the word shall be a subcommand the built binary lists in `help --all` · failure: none — the test names the file, the line and the word.
- REQ-6 (REQ-5): IF the built binary's help cannot be read or lists no subcommand, THEN the test shall fail naming the command it ran, never skipping.
- REQ-7: When a code span of those three documents names a launcher option after `moai cc` or `moai glm` (`-l`, `-a`, `--name`), the option shall be listed by the built binary's `cc --help` · failure: none — the test names the file, the line and the option.
- REQ-8: When a relative markdown link of those three documents names a path, the path shall be a file or folder of the tree · failure: none — the test names the file, the line and the path.
- REQ-9: When `installed_tools` looks for Claude Code, it shall find the folder the host home resolver gives (`CLAUDE_CONFIG_DIR` when that variable relocates the folder, else the HOME's `.claude` folder, which is also the answer when the resolver refuses the variable) · failure: none — pinned by AC-5.
- REQ-10: When `detect_hosts` of `install/install.sh` looks for Claude Code, it shall check the folder `CLAUDE_CONFIG_DIR` names when that variable is set and the HOME's `.claude` folder otherwise · failure: none — pinned by AC-6.
- REQ-11: When `install/install.sh` finds no Claude Code, its warning shall name both places it looked: the `claude` command and the folder · failure: none — pinned by AC-6.

## Acceptance

- AC-1 (REQ-1): README.md 와 README.ko.md 의 상태 폴더 칸과 FAQ 문장이 `state_root::MOAI_DIR`, `crate::PROJECT_STATE_DIR`(이전 프로젝트 폴더), `crate::HOME_STATE_DIR` 에서 읽은 폴더 이름을 말하고, 코드의 상수를 바꾸면 시험이 깨져요 · verify: `cargo nextest run --locked --offline --test docs_installer_paths the_state_folder_sentences_match_the_state_root_rules`
- AC-2 (REQ-2): 설치 섹션이 트리에 있는 두 스크립트와 `https://adk.mo.ai.kr/install.sh`, `install.ps1` 을 이름하고 `install/README.md` 로 링크하며 "not published"·`not_run` 문장이 없어요 · verify: `cargo nextest run --locked --offline --test docs_installer_paths the_installer_section_names_the_scripts_and_the_entry_the_installer_prints`
- AC-3 (REQ-3): 첫 실행 문장이 `moai update`, `moai cc`, `/moai project` 의 차례로 있고, 설치 스크립트의 "Next steps" 출력에 같은 명령이 있어요(스크립트 본문에서 읽어요) · verify: `cargo nextest run --locked --offline --test docs_installer_paths the_installer_section_names_the_scripts_and_the_entry_the_installer_prints`
- AC-4 (REQ-4): install/README.md 의 질문 출처 문장이 표준 입력, `/dev/tty`, 묻지 않음을 이 차례로 말하고 "never shows the menu" 문장이 없고, `install/install.sh` 의 `answer_source` 에 같은 세 갈래가 있어요 · verify: `cargo nextest run --locked --offline --test docs_installer_paths the_question_source_sentence_matches_the_installer`
- AC-5 (REQ-9): `installed_tools_with` 가 `CLAUDE_CONFIG_DIR` 이 옮긴 폴더를 찾고, 변수를 풀면 홈의 `.claude` 를, resolver 가 변수를 거부해도 홈의 `.claude` 를 찾아요 · verify: `cargo nextest run --locked --offline --test update_model_pick the_installed_tools_follow_the_claude_config_dir the_installed_tools_fall_back_to_the_home_folder_when_the_variable_is_refused`
- AC-6 (REQ-10, REQ-11): 임시 HOME 에서 `detect_hosts` 함수를 `sh` 로 돌려 `CLAUDE_CONFIG_DIR` 이 가리키는 폴더가 있으면 `claude` 를, 변수를 풀고 홈의 `.claude` 만 있어도 `claude` 를, 둘 다 없으면 호스트 없음을 내고, 스크립트의 경고 줄이 `claude` 명령과 폴더를 함께 말해요 · verify: `cargo nextest run --locked --offline --test docs_installer_paths the_installer_host_detection_follows_the_claude_config_dir`
- AC-7 (REQ-5, REQ-6): 세 문서의 `moai <word>`·`moai-cli <word>` 코드 조각이 모두 `help --all` 의 하위 명령이고, 없는 명령을 한 줄 더한 문서는 파일·줄·낱말을 말하며 실패해요 · verify: `cargo nextest run --locked --offline --test docs_installer_paths every_command_the_documents_name_is_a_subcommand_of_the_binary`
- AC-8 (REQ-7): `moai cc`·`moai glm` 뒤의 옵션 조각이 `cc --help` 에 모두 있어요 · verify: `cargo nextest run --locked --offline --test docs_installer_paths every_launcher_option_the_documents_name_is_in_the_launcher_help`
- AC-9 (REQ-8): 세 문서의 상대 링크가 모두 트리의 파일이에요 · verify: `cargo nextest run --locked --offline --test docs_installer_paths every_relative_link_of_the_documents_points_at_a_file_of_the_tree`

## Scope

- write: `README.md`
- write: `README.ko.md`
- write: `install/README.md`
- write: `install/install.sh`
- write: `src/update_models.rs`
- write: `tests/update_model_pick.rs`
- write: `tests/docs_installer_paths.rs`
- write: `.moai/specs/SPEC-20261126-install-docs-match-code/**`
- write (shared file, only by `moai card line-lease`): `tools/repo_scans.json`
- write (shared file, only by `moai card line-lease`): `tests/areas.toml` (the word `docs` of the new test file's name)
- read-only: `src/state_root.rs`, `src/home_root.rs`, `src/host_home.rs`, `install/install.ps1`
- forbidden: `Cargo.toml`, `skills/**`, `docs/**`, `src/gate.rs`

## Plan

1. SPEC 초안을 커밋하고 결정과 추천 기본값을 `LANE-BLOCKED kind=decision` 으로 올려요. 승인(리더 위임) 전에는 코드와 문서를 쓰지 않아요(`install` 과 `src/update` 접두사가 tier 2 바닥 목록에 있어요).
2. 조각 A (install/ 밖): RED — `tests/docs_installer_paths.rs` 의 시험 다섯 개(AC-1, AC-2/3, AC-7, AC-8, AC-9)와 `tests/update_model_pick.rs` 의 시험 둘(AC-5)이 먼저 실패하는 것을 확인해요. GREEN — README.md, README.ko.md 를 코드에 맞추고 `installed_tools_with` 를 더해요.
3. 조각 B (install/ 아래, 이름 바꾸기 카드(태그 글자를 바꾸는 조각)의 예약이 풀린 뒤 NeXT 끝에서): `install/README.md` 의 질문 출처 문장(AC-4)과 `install/install.sh` 의 `detect_hosts` 와 경고 줄(AC-6)을 고쳐요. 줄이 겹치지 않는다고 그 카드의 레인과 합의했고, 착지 순서는 그쪽이 먼저예요.
4. `tools/repo_scans.json` 에 새 시험 파일을 규칙으로 등록해요(line-lease 한 커밋).
5. 영수증(`check affected --card`)과 guard-mutant 줄을 READY 에 담아요.

## Risks

- 문서를 읽는 시험은 문서 문구에 묶여서, 맞는 문구 수정도 시험을 깨뜨릴 수 있어요. 시험은 낱말 몇 개와 코드 상수만 보고 문장 전체를 고정하지 않아요(고칠 때 시험 메시지가 어느 낱말이 빠졌는지 말해요).
- `install/` 쓰기는 다른 카드의 예약과 같은 파일이에요. 그 카드는 `install/README.md` 18·40번째 줄과 `install/install.sh` 의 태그 줄만 바꾸고 줄이 겹치지 않는다고 답했어요(2026-11-26). 그래도 조각 B 는 그 카드가 착지한 뒤 NeXT 끝에서 시작해요.
- 시험이 실제 바이너리를 돌려요(`help --all`). 빌드가 없거나 도움말을 못 읽으면 건너뛰지 않고 실패해요(REQ-6).
- 도움말에 이미 있는 `moai-cli` 이름은 이름 바꾸기 카드가 `moai` 로 옮길 수 있어요. 시험은 두 이름 모두 받아서 그 카드와 겹치지 않아요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 옛 SPEC 을 쓰지 않고 새 SPEC(`SPEC-20261126-install-docs-match-code`)으로 다시 시작해요. 옛 가지와 그 워크트리는 건드리지 않아요. | 옛 SPEC 은 Python 시험·Codex·adapters 를 담고 병합된 적이 없어요(감독자 판단, 추천 기본값) | 옛 가지를 이어서 쓰는 길은 남아 있어요 |
| D-2 | README 의 소개 문장, Codex·GitHub Issues 단어는 건드리지 않아요. | `SPEC-20261122-remove-codex-shipped-text` S2 가 README 와 Cargo 설명을 같은 커밋으로 맡아요 | S2 가 끝난 뒤 별도 카드 |
| D-3 | README.ko.md 도 같은 문장을 같이 맞춰요. | 한쪽만 고치면 미러가 어긋나요(추천 기본값) | 영문만 고치고 한국어는 다음 카드로 |
| D-4 | 시험은 Rust 하나(`tests/docs_installer_paths.rs`)이고 명령·옵션 목록은 빌드한 바이너리의 도움말에서 읽어요. 손으로 적은 목록은 없어요. | 리더 지시(2026-11-26), 이 저장소는 Rust 시험만 써요 | — |
| D-5 | 호스트 감지는 Claude 만이에요(Codex 는 빠졌어요). `CLAUDE_CONFIG_DIR` 은 `host_home` 의 규칙을 그대로 따라요. | 사용자 선택(2026-11-20) 과 Codex 제거 결정(2026-11-21) | — |
| D-6 | `install/` 아래 쓰기(조각 B)는 그 이름 바꾸기 카드가 착지한 뒤 NeXT 끝에서 해요. | 그 레인과의 합의(2026-11-26): 줄은 겹치지 않지만 예약은 그쪽이에요 | 리더가 지금 쓰라고 하면 조각 B 를 앞당겨요 |
| D-7 | `install.sh` 의 `detect_hosts` 시험은 새 Rust 시험이 함수 본문을 임시 HOME 에서 `sh` 로 돌려요. 기존 Python 시험은 고치지 않아요. | Python 금지, 스크립트에는 Rust 로 돌릴 틀이 아직 없어요 | 틀이 생기면 그쪽으로 옮겨요 |

## Result

2026-11-25T23:23:59Z (`date -u`) 기준, macOS, 작업 폴더 `.claude/worktrees/docs-install-g10`, 바탕 `e946ae8102bc234fe083408d644b0c822215956e`(가지를 딴 `NeXT` 끝; 이후 `NeXT` 가 움직였고 검사는 따라가지 않았어요).

- 순서: SPEC 초안 `54a81365cf`, RED `6163dd0a80`(`moai tdd red`가 새 시험 5개의 실패를 기록, `test-failure` exit 100), GREEN `7c70f8ad1d`, 시험 등록 `77f5488e72`. `moai tdd green`이 RED 체크를 다시 돌려 통과를 확인했고 사이클은 `moai tdd done` 으로 닫았어요.
- AC-1 met: `the_state_folder_sentences_match_the_state_root_rules` 통과(README.md 와 README.ko.md 의 상태 폴더 칸과 FAQ 문장이 새 폴더, 이전 RC 가 유지하는 폴더, 사용자 상태 규칙을 `earlier`/`이전` 낱말과 함께 말해요).
- AC-2 met, AC-3 met: `the_installer_section_names_the_scripts_and_the_entry_the_installer_prints` 통과(설치 섹션이 두 스크립트·두 주소·`install/README.md` 링크를 이름하고, `moai update` → `moai cc` → `/moai project` 차례가 설치 스크립트의 Next steps 출력과 같아요).
- AC-4 met: `the_question_source_sentence_matches_the_installer` 통과(`install/README.md` 의 한 문장이 표준 입력, `/dev/tty`, 묻지 않음을 이 차례로 말하고 "never shows the menu" 는 없어요; 같은 문단 앞 문장의 "stdin 과 stdout 이 둘 다 터미널" 은 `install.ps1` 만 말하도록 고쳤어요).
- AC-5 met: `the_installed_tools_follow_the_claude_config_dir`, `the_installed_tools_fall_back_to_the_home_folder_when_the_variable_is_refused` 통과. `installed_tools_with` 는 resolver 를 쓰고, `installed_tools` 는 `host_home::resolve` 로 프로세스 환경을 읽어요(차이는 아래).
- AC-6 met: `the_installer_host_detection_follows_the_claude_config_dir` 통과(`detect_hosts` 를 임시 HOME 에서 `sh` 로 돌려 세 경우와 경고 줄의 두 낱말을 확인).
- AC-7 met, AC-8 met, AC-9 met: 세 가드 시험이 현재 문서에서 통과하고, 변이로 각각 죽어요(아래).
- 시험 실행: `docs_installer_paths` + `update_model_pick` 27개 통과(`target/irad/green1.log`), 등록 뒤 `repo_scans_table`, `repo_scans_cli`, `gate_registry_tier`, `docs_installer_paths`, `update_model_pick` 110개 통과(`target/irad/reg1.log`).
- guard mutants: 9개, 각각 정확한 한 줄 치환으로 적용하고 `git checkout` 으로 되돌렸으며 마지막 `git status --short` 는 비었어요. 모두 exit 100 으로 죽었고 `FAIL` 줄이 대상 시험의 이름이에요(`target/irad/mutants3.out`, 줄은 READY 노트에 있어요).
- 차이 1: `installed_tools(home)` 는 `installed_tools_with(home, Some(home), env)` 가 아니라 `host_home::resolve(Claude, home)` 를 불러요. 이 세션처럼 `CLAUDE_CONFIG_DIR` 가 있는 환경에서 `Some(home)` 은 시험용 임시 HOME 에도 변수를 적용해 기존 시험 2개를 깨뜨렸어요. `resolve` 는 `home` 이 계정 홈일 때만 변수를 세요.
- 차이 2: Decision D-6 (조각 B 는 이름 바꾸기 카드 착지 뒤)을 리더가 풀었어요(리더 메시지, 이름 바꾸기 카드의 READY 가 예약 경로를 반납한 뒤). 그 카드의 줄(install.sh 39·266·285·311·603, install.ps1, README 18·40)은 건드리지 않았고, 착지는 그쪽이 먼저예요.
- 차이 3: 새 시험 파일은 `tests/areas.toml` 에 첫 낱말 `docs` 가 있어야 해서 `spec` 영역의 `doc` 옆에 더했어요(공유 파일, line-lease).
- Gap: `README.md` 45번째 줄의 "릴리스 아카이브는 macOS arm64 에만 올라가요" 문장은 Rust 릴리스가 아직 없다는 사실과 어긋나 보이지만 릴리스 워크플로가 올리는 것을 읽지 않아 그대로 뒀어요. README.ko.md 는 같은 줄에 다른 문장(RC1 목표 범위)을 가져 영문과 어긋나요.
- Gap: README.ko.md 의 옛 GitHub Issue 장부 문구(다이어그램, "Issue 하나에 세션 하나")와 Codex 단어는 README.md 와 달리 그대로예요. Codex 단어는 Codex 제거 SPEC 의 조각이 맡아요.
- Gap: `install/install.ps1`, `README` 의 소개 문장, 설치 스크립트가 Rust 릴리스를 서빙하는 일은 범위 밖이에요.
- Gap: `CLAUDE_CONFIG_DIR` 가 없는 폴더를 가리키고 홈의 `.claude` 가 있을 때 `install.sh` 가 무엇을 해야 하는지는 SPEC 이 말하지 않아서 정하지 않았어요(지금은 변수가 가리키는 폴더만 봐요). 시험도 이 경우를 고정하지 않아요.
- Gap: 이 카드의 계획, 구현, 문서는 `moai-impl` 과 `moai-docs` 에이전트가 했고 별도 `moai-plan` 은 쓰지 않았어요(계획은 SPEC 의 Plan 이 그대로 대신해요). 위험한 경로는 없어요: 지우는 일과 프로세스를 띄우는 일이 없고, `sh` 는 임시 HOME 에서 `detect_hosts` 함수 본문만 돌려요.
- `check affected --card` 영수증과 그 단계별 exit 는 READY 노트에 있어요.
