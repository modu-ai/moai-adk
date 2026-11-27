---
id: SPEC-20261127-installed-binary-name-moai
title: "설치되는 실행 파일 이름을 moai 하나로: 새 설치는 bin/moai-cli 를 만들지 않고, 이미 있는 moai-cli 는 읽기만 해요"
status: approved
revision: 1
created: 2026-11-27
parent: SPEC-20261121-moai-name
approval:
  by: user
  at: 2026-11-27
  basis: "리더 노트 10780 (USER-APPROVAL, 2026-11-27, 리더 세션의 질문 도구): 사용자가 직접 골랐어요 — \"승인, 호환 링크 없음 (Recommended)\". D-1 (a)와 나머지 기본값(Q2 아니요, Q3 건드리지 않음, D-4 상대 링크, D-6 보고하고 멈춤, Q5 그대로)을 승인하고 부모 SPEC-20261121-moai-name 의 조각 5 를 이 SPEC 이 대신하는 것도 승인했어요. 이미 있는 moai-cli 를 지우는 일은 승인에 없어요."
  digest: "sha256:cbab784d5f855f933fefd0ccbaacbd573190c9651d8af3fa01f4fc86000b9eaa"
---

# 설치되는 실행 파일 이름을 moai 하나로: 새 설치는 bin/moai-cli 를 만들지 않고, 이미 있는 moai-cli 는 읽기만 해요

## Intent

사용자 결정(2026-11-27, 리더 세션의 질문 도구, 원장 노트 10487): "moai-cli는 더이상 사용하지 않는다". 지금 제품의 설치는 반대로 움직여요. 설치 스크립트는 `versions/<v>/bin/moai-cli` 를 복사하고 `bin/moai` 를 그 하드 링크로 둬요(`install/install.sh:421-424`, `install/install.ps1:509-511`). `moai update` 의 이름 검사(`src/bin_names.rs` `plan_for`)는 `bin/moai-cli` 가 일반 파일이어야 한다고 기대해요. 그래서 `bin/moai` 하나만 있는 설치(리더의 손 설치, 노트 10690)는 `bin_names.state` 가 `blocked`(`bin/moai-cli is missing or not a regular file`)로 나와요. 이 설치의 출발 사실은 빌드 59 미리보기예요(`update-59-preview.out` 의 `bin_names`).

이 SPEC 은 설치된 모양만 바꿔요: 새 설치는 `bin/moai` 하나만 두고, `moai update` 와 런처 이름 검사와 switch 는 `moai-cli` 를 기대하지 않고, 사용자가 읽는 스킬·문서·도움말은 `moai-cli` 를 실행할 명령으로 적지 않아요. 이미 설치된 `moai-cli`(파일, 링크, PATH 링크)는 읽고 보고만 해요. 지우거나 새로 고치지 않아요.

부모 SPEC `SPEC-20261121-moai-name` 의 조각 5("바이너리와 패키지")는 같은 파일 이름을 맡지만 `moai-cli` 를 호환 링크로 남기는 설계예요. 이 SPEC 이 그 조각의 설치 모양 부분을 대신하고, 호환 링크는 Decisions 의 D-1 이 정해요. 부모 문서를 고치는 줄은 이 SPEC 의 승인 뒤에 써요.

### 조각 경계

한 SPEC, 네 조각이고 이 순서로 구현해요. 조각마다 커밋이 따로 있고 착지는 리더가 정해요(조각 2·3 이 크면 자식 SPEC 으로 나눠요).

| 조각 | 내용 | 요구 |
| --- | --- | --- |
| 1 | 이름 검사, 설치 식별자, 영수증 찾기, 남은 `moai-cli` 보고 (Rust) | REQ-5, 6, 7, 8 |
| 2 | 설치 스크립트 둘과 `install/README.md` | REQ-1, 2, 3, 4 |
| 3 | switch 의 런처 이름 (`src/update/switch.rs`) | REQ-9 |
| 4 | 스킬 fallback, 문서, 도움말, 골든, 저장소 스캔 | REQ-10, 11 |

## Non-goals

- 빌드 타깃과 패키지 이름: `Cargo.toml` 의 `[[bin]] moai-cli`, `CARGO_BIN_EXE_moai-cli`, `MOAI_CLI_BIN` 은 부모의 조각 5 가 맡아요(테스트가 그 이름으로 바이너리를 불러요).
- 릴리스 아카이브 이름 `moai-cli-<target>.tar.xz` 와 `release.yml` 의 두 바이너리 검사: 부모의 조각 6 이 맡아요. 설치 스크립트는 아카이브 안에서 읽는 파일 이름만 바꿔요.
- 영수증의 `product: "moai-cli"` 값과 `install::PRODUCT`: 스키마가 그 값만 받아서(`validate_receipt`) 바꾸면 기존 영수증이 모두 읽히지 않아요. 파일 이름이 아니라 식별 문자열이에요(Q5).
- 이미 설치된 버전 폴더의 모양을 바꾸는 일과 `moai-cli` 를 지우는 일.
- 옛 상태 폴더에서 새 상태 폴더로 옮기는 단계(`src/home_install.rs` 의 `version_bin`, `alias_table`): 옛 상태 폴더의 버전은 모두 `moai-cli` 모양이고 그 단계는 그것을 읽어요.
- 훅·상태 줄 명령 안의 `moai-cli`: `moai update` 가 처음 쓴 그대로인 항목은 이미 `moai` 로 고쳐요(`src/hook_legacy.rs`, `src/status_line.rs`). 사용자가 고친 항목은 남겨 두고 충돌로 보여요.
- Windows 와 Linux 확인(`not_run`).

## Requirements

- REQ-1: When an install script places a build, it shall copy the archive's `moai` file to `versions/<version>/bin/moai` (an archive that holds only `moai-cli` is copied under the name `moai`) and create no `bin/moai-cli`. · failure: none — an archive without either file is REQ-2.
- REQ-2 (REQ-1): IF the archive holds neither `moai` nor `moai-cli`, THEN the install script shall stop before it writes anything and name the missing file.
- REQ-3: When the version folder already holds a regular `bin/moai` with the archive's bytes, the install script shall report the version as installed and write nothing beyond the missing release parts. · failure: none — a folder holding other bytes is refused as today.
- REQ-4: When an install script links the commands into the link folder, it shall create the link `moai` only and leave an existing link `moai-cli` as it is. · failure: none — a link folder entry that is a plain file keeps today's warning.
- REQ-5: When `moai update` runs from a version folder whose `bin/moai` is a regular file and which has no `bin/moai-cli`, the name check shall report the state `unchanged` and write nothing. · failure: none — a folder outside this shape is REQ-6.
- REQ-6: When the version folder holds `bin/moai-cli`, the name check shall judge it as it does today (hard link, either link layout, create, replace_copy, blocked) and create or refresh no `bin/moai-cli`. · failure: none — the earlier shapes keep their states.
- REQ-7: When a build looks for the install receipt of an earlier version folder, the version lookup shall find it whether that folder holds `bin/moai-cli` with a hard link `bin/moai` or only `bin/moai`. · failure: none — an unreadable receipt is reported as today.
- REQ-8: When the name check or the update preview sees a `bin/moai-cli` or a link `moai-cli` in the link folder, it shall list that path as left in place and change nothing. · failure: none — nothing found, nothing listed.
- REQ-9: When the switch part looks for the Rust launcher beside the script, it shall use the launcher the decision D-4 names and no longer require a `moai-cli` beside the script. · failure: none — a missing launcher keeps the held state with a sentence naming the file it looked for.
- REQ-10: When the skill, the shipped documents, the help text or the help goldens name a command to run, they shall name `moai` and never `moai-cli`. · failure: none — a sentence that names the earlier file on disk as history or as a thing to report is allowed by the scan list.
- REQ-11 (REQ-10): IF a shipped text names `moai-cli` as a command to run outside the scan list, THEN the repository scan test shall fail and name the file and line.

## Acceptance

- AC-1 (REQ-1, REQ-3): 임시 HOME 에서 `file://` 릴리스 목록과 가짜 아카이브(`moai`·`moai-cli` 둘 다 든 것)로 `install.sh` 를 돌리면 `versions/<v>/bin/moai` 는 일반 파일이고 `bin/moai-cli` 는 없어요. 같은 명령을 다시 돌리면 "already installed" 로 끝나고 아무것도 새로 쓰지 않아요. 1 passed · verify: `cargo test --locked --test install_cli ac1_ -- --nocapture`
- AC-2 (REQ-1): `moai-cli` 만 든 아카이브는 `bin/moai` 라는 이름으로 복사되고 `bin/moai-cli` 는 없어요. 1 passed · verify: `cargo test --locked --test install_cli ac2_ -- --nocapture`
- AC-3 (REQ-2): 두 파일 모두 없는 아카이브는 아무것도 쓰기 전에 멈추고 빠진 파일 이름을 말해요(버전 폴더와 임시 폴더가 남지 않음). 1 passed · verify: `cargo test --locked --test install_cli ac3_ -- --nocapture`
- AC-4 (REQ-4): 링크 폴더에 `moai` 링크만 새로 생기고, 미리 있던 `moai-cli` 링크는 대상이 그대로예요. 1 passed · verify: `cargo test --locked --test install_cli ac4_ -- --nocapture`
- AC-5 (REQ-5): 임시 HOME 의 버전 폴더에 실제 바이너리를 `bin/moai` 하나로만 두고 `moai update --json` 을 돌리면 `bin_names.state` 가 `unchanged` 이고 `--confirm` 뒤 `bin_names_result` 도 `unchanged`, 쓴 파일이 없어요. 1 passed · verify: `cargo test --locked --test bin_names_cli ac5_only_moai -- --nocapture`
- AC-6 (REQ-6): 하드 링크, 두 링크 모양, `moai-cli` 만 있는 폴더(`create`), 다른 프로그램인 `bin/moai`(`blocked`)의 기대는 지금과 같고, `bin/moai-cli` 는 어느 상태에서도 새로 생기지 않아요. verify: `cargo test --locked --test bin_names_cli -- --nocapture`
- AC-7 (REQ-7): 하드 링크 폴더(영수증이 `moai-cli` 경로 아래)에서 `bin/moai` 하나뿐인 새 폴더로 올린 빌드가 이전 영수증을 찾아 `installed` 가 `none` 이 아니에요. 새 폴더의 `bin/moai` 는 자기 경로가 설치 식별자예요. 1 passed · verify: `cargo test --locked --test update_installed_version_by_name ac7_ -- --nocapture`
- AC-8 (REQ-8): `bin/moai-cli`(파일, 링크)와 링크 폴더의 `moai-cli` 가 있으면 미리보기가 그 경로를 "left in place" 로 나열하고, 실행 전후 파일 목록·링크 대상·inode 가 그대로예요. 1 passed · verify: `cargo test --locked --test bin_names_cli ac8_ -- --nocapture`
- AC-9 (REQ-9): 이름 바꾸기 스크립트(v1)가 `path/moai` 에 있고 `path/moai-cli` 가 없는 임시 HOME 에서, 기본 결정(D-4)의 런처가 응답하면 switch 미리보기가 링크를 제안하고 확인 뒤 스크립트를 옆에 `moai.before-<UTC>` 로 보관해요. 런처가 없으면 `held` 와 찾은 파일 이름을 말해요. 1 passed · verify: `cargo test --locked --test update_switch_cli ac9_ -- --nocapture`
- AC-10 (REQ-10, REQ-11): 저장소 스캔 테스트가 스킬(`skills/`), 템플릿, `install/README.md`, `README*.md`, `src/help.rs` 의 도움말 글, 도움말 골든 여덟 개에서 `moai-cli` 를 실행할 명령으로 적은 줄을 찾지 못하고, 허용 목록의 줄은 이유를 달고 있어요. 스킬 문장은 Rust 서명이 아닐 때 막힘을 보고하라고 말해요(D-6). 1 passed · verify: `cargo test --locked --test shipped_text_never_runs_moai_cli -- --nocapture`
- AC-11 (REQ-10): 도움말 골든과 `bin-names-plan` 설명이 새 문장과 같아요. verify: `cargo test --locked --test help_layout -- --nocapture`
- AC-12 (REQ-1, REQ-2, REQ-3, REQ-4, REQ-5, REQ-6, REQ-7, REQ-8, REQ-9, REQ-10, REQ-11): 실제 이 Mac 의 설치에서 `moai update` 미리보기를 읽기만 해서 `bin_names` 가 더 이상 `blocked` 가 아니고 남은 `moai-cli` 가 나열되는지 본다 · manual: 리더가 구현이 착지한 빌드를 설치한 뒤 `moai update` 미리보기를 한 번 읽어 노트 10690 의 사실(`blocked`)과 비교해요. 확인 전에는 met 로 적지 않아요.
- AC-13 (REQ-1, REQ-2, REQ-3, REQ-4, REQ-5, REQ-6, REQ-7, REQ-8, REQ-9, REQ-10, REQ-11): SPEC 검사를 통과해요 · verify: `moai spec check SPEC-20261127-installed-binary-name-moai`

## Scope

- write: `.moai/specs/SPEC-20261127-installed-binary-name-moai/**`, `src/bin_names.rs`, `src/install.rs`, `src/update.rs`, `src/update/switch.rs`, `install/install.sh`, `install/install.ps1`, `install/README.md`, `skills/moai/SKILL.md`, `tests/bin_names_cli.rs`, `tests/install_cli.rs`, `tests/update_switch_cli.rs`, `tests/update_installed_version_by_name.rs`, `tests/shipped_text_never_runs_moai_cli.rs` (코드와 테스트는 승인 뒤 구현이고 초안 단계에서 쓰는 것은 SPEC 폴더뿐이에요; install.rs 는 identity_exe 한 곳, update.rs 는 영수증 찾기 두 곳이에요)
- line lease (shared files, one `moai card line-lease` edit each): `src/help.rs` (`bin-names-plan` 행), `tests/fixtures/help/all.txt`, `tests/fixtures/help/all-60.txt`, `tests/fixtures/help/all-80.txt`, `tests/fixtures/help/all-dumb.txt` (골든은 `MOAI_HELP_GOLDEN_WRITE=1` 로 다시 만듦), `tests/areas.toml` (새 테스트 첫 낱말), `tools/repo_scans.json` (새 스캔 테스트 등록), `skills/moai/references/source-manifest.json` (스킬 파일 해시), `docs/design-review/appendix/operations/bin_names.md` (문서 한 줄 이름)
- read-only: `src/home_install.rs`, `src/hook_legacy.rs`, `src/status_line.rs`, `Cargo.toml`, `.github/workflows/release.yml`, `dist-workspace.toml`, `.moai/specs/SPEC-20261121-moai-name/spec.md`, `.moai/specs/SPEC-20261127-launcher-bin-layout/spec.md`.
- forbidden: 실제 HOME 과 그 아래 모든 폴더(상태 폴더 둘과 Claude 폴더), 다른 프로젝트, 네트워크 호출(설치 스크립트 테스트는 `file://` 만), 릴리스·태그·푸시, 아카이브 이름과 빌드 타깃 이름, `CLAUDE.md`.

## Interview

- Q1: `moai-cli` 를 호환 링크로 남기나요? → A: 사용자 결정 10487 은 "더이상 사용하지 않는다" 예요. 기본값은 남기지 않음(D-1). 선택지와 대가는 D-1 에 있어요.
- Q2: 새 설치가 링크 폴더에 `moai-cli` 링크를 만드나요? → A: 기본값은 아니요(REQ-4). 미리 있는 링크는 그대로 둬요.
- Q3: 사용자가 고친 훅·상태 줄 명령이 `moai-cli` 를 부르면? → A: 기본값은 건드리지 않음. 처음 쓴 그대로인 항목은 이미 `moai update` 가 `moai` 로 고쳐요(비목표). 대가는 D-1 의 (a) 에 적었어요.
- Q4: switch 의 런처는 무엇인가요? → A: D-4.
- Q5: 영수증의 `product: "moai-cli"` 문자열은 바꾸나요? → A: 기본값은 그대로(비목표). 스키마를 올려 두 값을 읽게 하는 일은 별도 SPEC 이에요.
- Q6: 스킬의 "`moai --version` 이 Rust 서명이 아니면 `moai-cli` 를 써라" 문장은? → A: 대체 실행 파일이 없어져요. 기본값은 "다른 `moai` 가 PATH 앞에 있다고 보고하고 멈춘다"(D-6).
- Q7: 아카이브에 `moai` 가 들어 있나요? → A: `release.yml` 이 두 바이너리를 모두 만들었다고 검사하고 `Cargo.toml` 이 두 `[[bin]]` 을 둬서 들어 있다고 읽혀요. 실제 릴리스 아카이브는 아직 없어서 직접 열어 보지 못했어요(unverified). 그래서 REQ-1 이 `moai-cli` 만 든 아카이브도 받아요.
- Q8: 이 초안을 승인하나요, 호환 링크는 어떻게 하나요, 부모 조각 5 를 이 SPEC 이 대신해도 되나요? → A: 사용자가 직접 골랐어요(2026-11-27, 리더 세션의 질문 도구, 노트 10780): "승인, 호환 링크 없음 (Recommended)". D-1 (a)와 나머지 기본값(Q2 아니요, Q3 건드리지 않음, D-4 상대 링크, D-6 보고하고 멈춤, Q5 그대로)을 승인하고 부모 조각 5 를 대신하는 것도 승인했어요. 이미 있는 `moai-cli` 를 지우는 일은 승인에 없어요. Q7 은 확인하지 못한 채로 남아요(unverified).

## Plan

1. SPEC 초안과 `moai spec check`. tier 2(HOME 에 쓰는 설치 모양)라서 승인 전에는 아래를 시작하지 않아요. 레인은 승인을 직접 기록하지 않아요.
2. 조각 1 RED: AC-5~AC-8 테스트. base 에서 AC-5, AC-7(새 폴더), AC-8 은 실패하고 AC-6 은 통과해야 해요. GREEN: `plan_for` 에 `bin/moai` 단독 분기 하나, `identity_exe`·영수증 찾기 확인, 남은 `moai-cli` 나열.
3. 조각 2 RED/GREEN: AC-1~AC-4. 가짜 아카이브는 테스트가 실제 빌드를 `moai`·`moai-cli` 이름으로 복사해 `tar -cJf` 로 만들고 `.sha256` 을 옆에 둬요(`xz` 가 없는 호스트는 그 테스트를 건너뛰고 `not_run` 으로 적어요).
4. 조각 3: AC-9. 조각 4: AC-10, AC-11, 골든 재생성(줄 임대), 스킬 해시.
5. 가드 뮤턴트(조각마다 새 가드 테스트 하나씩): (a) 단독 `bin/moai` 분기를 지움 → AC-5 실패, (b) 설치 스크립트가 `bin/moai-cli` 를 다시 만들게 함 → AC-1 실패, (c) 링크 폴더에 `moai-cli` 링크를 만들게 함 → AC-4 실패, (d) 스캔 허용 목록을 비움 → AC-10 의 이유 확인 실패.
6. 위험 경로(HOME 에 쓰고 링크를 만드는 변경): 임시 HOME 에서 실제 스크립트와 실제 바이너리로 AC-1~AC-5, AC-8 을 돌려 `bin/moai-cli` 가 생기지 않고 미리 있던 `moai-cli` 가 그대로임을 보여요. 막다른 길: 다른 프로그램인 `bin/moai`(`blocked`)도 같은 테스트에서 보여요.
7. `cargo fmt --all`, AC 의 verify 명령을 적힌 그대로 heavy slot 에서 실행, `moai check affected --card` 한 번, READY 에 tier 2 와 조각 경계를 적어요.
8. 승인 뒤 한 줄: 부모 SPEC 의 조각 5 문단과 D-6·D-21 이 이 SPEC 을 가리키게 하는 개정은 부모 SPEC 의 승인과 함께 따로 써요(이 SPEC 의 구현 범위가 아니에요).

## Risks

- R-1: 새 설치에는 `moai-cli` 가 없어요. `moai-cli` 를 이름으로 부르던 사용자 스크립트와 사용자가 고친 훅·상태 줄이 끊겨요. 링크 폴더의 옛 `moai-cli` 링크는 그대로 남아서 옛 빌드를 계속 가리켜요: 새 빌드로 바뀐 줄 알고 옛 빌드가 실행될 수 있어요. REQ-8 이 그 링크를 미리보기에 보여 주는 이유예요(D-1).
- R-2: 설치 식별자는 경로에서 나와요. 새 폴더는 `bin/moai`, 옛 폴더는 `bin/moai-cli` 가 식별자예요. 다른 폴더의 영수증은 두 이름을 이미 읽어요(`folder_receipt_ids`, `SPEC-20261124-installed-version-by-name`). AC-7 이 새 폴더로 올릴 때를 고정해요.
- R-3: `execution_link_version` 의 후보(`path/moai`, `path/moai-cli`, `bin/moai`)는 이미 `moai` 를 포함해서 바꾸지 않아요. 옛 `path/moai-cli` 링크만 가리키는 기계는 그 링크를 읽는 쪽이 그대로 받아요.
- R-4: 설치 스크립트 테스트는 `xz` 가 필요해요. 없는 호스트에서는 건너뛰어서 AC-1~AC-4 가 그 호스트에서는 증거가 아니에요(Windows 의 `install.ps1` 은 `not_run`).
- R-5: 아카이브에 `moai` 가 정말 들어 있는지 실제 릴리스로 확인하지 못했어요(Q7). REQ-1 이 `moai-cli` 만 든 아카이브도 받아서 이 가정이 틀려도 설치는 깨지지 않아요.
- R-6: `--confirm` 이 들어간 테스트는 환경을 비우고 돌려요(`CLAUDE_CONFIG_DIR` 을 물려받으면 실제 Claude 폴더에 쓸 수 있어요).
- R-7: 이름 필터가 테스트 0개와 맞아도 exit 0 이에요. 각 AC 는 `1 passed` 까지 증거로 적어요.
- R-8: 허용 목록은 이름을 이력으로 적는 문장을 위한 것이에요. 목록이 길어지면 스캔이 무뎌져서, 줄마다 이유를 달아요(AC-10).

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 호환 링크를 두지 않아요(기본값). 선택지: (a) 두지 않음 — 새 폴더에 `moai-cli` 없음, 옛 `moai-cli` 는 보고만 / (b) 한 릴리스 동안 새 폴더에 `bin/moai-cli` 를 `moai` 로 이은 링크로 둠 (부모 조각 5 의 설계) / (c) 설치 때마다 링크 폴더의 `moai-cli` 링크를 새 `bin/moai` 로 다시 가리킴. 대가: (a) 는 사용자가 고친 훅·상태 줄과 스크립트가 끊기고 옛 링크가 옛 빌드로 남음(R-1, 미리보기가 알림), (b) 는 사용자 말("더이상 사용하지 않는다")과 어긋나고 나중에 지우는 일이 한 번 더 필요해요, (c) 는 HOME 에 `moai-cli` 를 계속 쓰는 일이라 "새로 고치지 않는다"와 어긋나요 | 사용자 결정 10487, 부모 조각 5, 사용자 선택 10780 (승인, 2026-11-27) | SPEC 개정 |
| D-2 | 설치 스크립트가 복사하는 파일은 아카이브의 `moai`, 없으면 `moai-cli` 를 `moai` 라는 이름으로 | R-5, Q7 | 한 줄 |
| D-3 | 이름 검사는 `bin/moai` 단독을 `unchanged` 로 보고 쓰지 않아요. 기존 모양의 판정은 바꾸지 않음 | 빌드 59 미리보기, `SPEC-20261127-launcher-bin-layout` 의 요구 1~7 이 이미 정한 모양 | 분기 제거 |
| D-4 | switch 의 런처는 실행 중인 빌드 폴더의 `bin/moai` 이고 링크는 그 빌드로 가는 상대 링크예요(기본값). 선택지: (a) 그것 / (b) 옛 `path/moai-cli` 가 있을 때만 제안하고 없으면 `held`. 대가: (a) 는 링크 모양(옆의 `moai-cli` 가 아님)이 바뀌어 `is_the_link` 와 되돌리기 판정을 함께 바꿔요, (b) 는 새 설치에서는 switch 가 늘 `held` 예요 | 부모의 Go 경로 제거 SPEC 이 정한 switch 의 역할 | 조각 3 만 되돌림 |
| D-5 | 영수증의 `product` 문자열과 `PRODUCT` 상수는 그대로 | `validate_receipt` 가 그 값만 받음, 비목표 | 별도 SPEC |
| D-6 | 스킬의 fallback 은 "다른 `moai` 가 PATH 앞에 있으니 멈춘다"고 보고하는 문장이에요(기본값). 대체 명령을 적지 않아요 | `CLAUDE.md`: 검증된 대체가 없으면 막고 보고 | 문장 한 줄 |
| D-7 | 옛 상태 폴더 이동 단계(`home_install.rs`)는 이번에 바꾸지 않음 | 거기 있는 버전은 모두 옛 모양 | 후속 |

## Result

(구현 전. 이 SPEC 은 초안이고 승인과 구현이 남아 있어요.)
