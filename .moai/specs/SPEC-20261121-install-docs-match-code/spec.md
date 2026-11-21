---
id: SPEC-20261121-install-docs-match-code
title: "README and install/README match the code; a test keeps the named commands and paths real"
status: draft
revision: 1
created: 2026-11-21
---

## Intent

install-flow 보고 §7 G10(코드가 이겨요)의 사용자 결정(2026-11-20T23:1xZ, 노트 4679): README·install/README·adapters/orca.md 를 코드에 맞추고, 문서가 이름하는 명령·경로가 실제로 있는지 보는 테스트를 더하며, 호스트 감지가 CLAUDE_CONFIG_DIR 을 따르게. 조사 결과와 조각:

- **README.md:40(state folders)** — 문서는 `.moai-cli/` (project) 였지만 코드는 새 프로젝트를 `.moai/` 로 시작해요(`state_root::create_dir`: MoAI state 가 전혀 없는 프로젝트는 이 세대의 폴더 — 이전 RC 나 Go 로 시작한 프로젝트는 자기 폴더를 유지) — `skills/moai/workflows/project.md` 는 이미 `.moai` 를 말해 문서끼리도 어긋났었어요. 수리: 새 프로젝트 `.moai/`·이전 RC 프로젝트 `.moai-cli/` 유지·사용자 state 는 `~/.moai-cli/`(`HOME_STATE_DIR`).
- **README.md:53(installer)** — "not published, not_run" — 실제로는 `install/install.sh`·`install/README.md` 가 있고 설치 주소는 `https://adk.mo.ai.kr/install.sh`(Go 세대와 같은 주소). 수리: 실제 주소·한 줄 명령·install/README 링크·attestation 문서 링크.
- **README.md:75(entry)** — `/moai read this project and set it up`(draft)를 진입처럼 안내 — 실제 설치 후 다음 단계는 `moai update` → `moai cc` 등 세션 시작 → 세션 안에서 `/moai project`(install.sh:581-598). 수리: 실제 진입을 안내하고 draft 는 draft 로(user-entry-draft 링크 유지).
- **install/README.md:32(질문 소스)** — "A pipe (curl … | sh) … never shows the menu" — 스크립트는 stdin 이 터미널이 아니면 `/dev/tty` 를 읽어 파이프에서도 그 터미널에 앉은 사람에게 질문해요(install.sh answer_source). 수리: 실제 동작(standard input → `/dev/tty` → 없음)으로.
- **adapters/orca.md** — 이미 저장소에 없어요(provisional Python adapter 제거 커밋이 adapters 를 지웠어요) — CLAUDE_CONFIG_DIR unsupported 문구는 사라진 것으로 해소.
- **호스트 감지(CLAUDE_CONFIG_DIR)** — `src/update_models.rs` 의 `installed_tools` 와 `install.sh` 의 `detect_hosts` 가 `~/.claude` 만 봐요 — **별도 조각**: 두 파일이 다른 카드의 예약(C-248·C-246, lease 유효)이라 반환 뒤 widen 해서 고쳐요(`host_home.rs` 의 resolver 가 정답 — Claude 는 `CLAUDE_CONFIG_DIR` else `~/.claude`).

## Non-goals

- `skills/moai/workflows/project.md` 의 `.moai` 문구: 이미 코드와 일치해요 — README 쪽이 고쳐질 것.
- verdict 줄 형식·verdict 늘려 읽기 같은 gate 주제: 이 카드는 설치 문서만.
- CODEX_HOME 감지: 결정 노트가 CLAUDE_CONFIG_DIR 만 이름해요 — Codex 쪽은 다음 카드로 리더에게 노트.
- `install/install.ps1` 의 Windows 질문 흐름: 보고서가 줄 단위로 읽지 않은 부분 — 이 카드도 건드리지 않아요.

## Requirements

- REQ-1: While the installation documents name state folders, the installer or the entry, they shall speak for the code — README's state-folder cell names `.moai/` for a new project and the kept `.moai-cli/` for a project that began on the earlier RC with the home's `.moai-cli/` as the user state, the installer section names the real addresses (`https://adk.mo.ai.kr/install.sh` and `install.ps1`) and links `install/README.md`, and the entry names what the installer prints (`moai cc`, then `/moai project` in the session) while the natural-language draft stays a draft link. · failure: none — pinned by AC-1.
- REQ-2: While install/README describes when the installer asks questions, its sentence shall match the script's answer source: standard input when it is a terminal, else the session's terminal, which the documented pipe still reaches, else no question. · failure: none — pinned by AC-2.
- REQ-3: A test (`tests/docs_installer_paths_test.py`) shall turn red when the documents drift: every repository-relative markdown link in README and install/README points at a file of the tree, every `moai <word>` and `moai-cli <word>` the three documents print is a subcommand of the binary's own help, the named skill entry is a file under `skills/moai`, the state-folder sentence matches `state_root` and `HOME_STATE_DIR`, and the question-source sentence matches the script. · failure: none — pinned by AC-3.
- REQ-4: While the host-detection files are another card's reservation, this SPEC's later slice shall make `update_models.rs`'s `installed_tools` and `install.sh`'s `detect_hosts` follow the host home resolver — Claude's folder is `CLAUDE_CONFIG_DIR` when it is set, else the home's `.claude` folder — recorded here as the remaining half, taken when the reservations return. · failure: none — the slice lands behind AC-3 with its own tests.

## Acceptance

- AC-1 (REQ-1): README's state-folder cell, installer section and entry paragraph carry the sentences above; no "not published", "not_run" or "never shows the menu" sentence remains in the two documents.
  verify: `python3 -m unittest tests.docs_installer_paths_test -v` (the state-folder and installer-line tests) plus `grep -c 'not_run\|not published' README.md` answering 0 for those sentences.
- AC-2 (REQ-2): install/README's question-source sentence names standard input, `/dev/tty` and the no-terminal case in that order, and the script's `answer_source` really reads that way.
  verify: `python3 -m unittest tests.docs_installer_paths_test -v` (the installer-line test).
- AC-3 (REQ-3): `python3 -m unittest tests.docs_installer_paths_test` — 5 passed, 0 failed, with a built `moai-cli` (the help comparison is the binary's own).
  verify: `python3 -m unittest tests.docs_installer_paths_test -v`
- AC-4 (REQ-4): not run in this slice — the host-detection files are other cards' reservations; the slice lands with its own tests when they return.
  manual: the host-detection files are other cards' reservations — the later slice adds the detection change and its own tests when they return.

## Scope

- write: this SPEC folder, `README.md`, `install/README.md`, `tests/docs_installer_paths_test.py`
- reservations this card met while taking (the later slice takes them when returned): `install/install.sh` is held by the language-once card (lease valid), `src/update_models.rs` by the GLM-key card (lease valid); both hold a live lease, so no take over their scope
- read-only: `src/host_home.rs`, `src/state_root.rs`, `src/lib.rs`, `install/install.sh`, `skills/moai/workflows/project.md`, `docs/release-attestation.md`, `docs/user-entry-draft.md`
- forbidden: 실제 HOME 쓰기, 문서가 코드를 따라오게 하는 대신 코드를 문서에 맞추는 것, 배포 skill 문서 변경

## Plan

1. RED: the five-scenario test (`tests/docs_installer_paths_test.py`) with the correct sentences as the expectations — the current README fails the state-folder and the installer-section tests.
2. GREEN: the four document repairs above (state folders, installer section, entry, question source).
3. 회귀: python unittest(이 시험 + skill_references_test — README 는 skills 밖이라 무관하지만 함께), spec check, diff-check — 커밋·push.
4. Later slice (REQ-4): widen with `src/update_models.rs` and `install/install.sh` after C-248·C-246 return — the detection follows `host_home`'s resolver, tests follow.

## Risks

- README 문구를 시험이 문장 단위로 묶어요: 문구를 다듬으면 시험도 함께 바뀌는 게 드리프트 검사의 비용이에요 — 문장의 사실(폴더·주소·명령)만 묶고 서사는 묶지 않아요.
- `moai <word>` 추출이 버전 표기(`moai-cli 4.…`)를 잡지 않게 정규식이 소문자 word 만 — 대문자·숫자로 시작하는 표기는 매칭 밖이에요.
- install.sh 는 다른 카드의 예약이라 이 조각은 읽기만 해요 — 조각 2 가 그 파일을 바꿀 때 README·install/README 의 문구(이미 실제 동작을 말함)는 다시 안 바뀌어야 해요 — 감지 수리는 문서에 쓰인 동작과 같은 방향이에요.
