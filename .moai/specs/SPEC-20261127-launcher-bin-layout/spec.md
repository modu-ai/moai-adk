---
id: SPEC-20261127-launcher-bin-layout
title: "런처 이름 검사가 링크로 이어진 bin 두 이름을 설치된 모양으로 받아들임"
status: draft
revision: 1
created: 2026-11-27
---

# 런처 이름 검사가 링크로 이어진 bin 두 이름을 설치된 모양으로 받아들임

## Intent

`moai update` 의 런처 단계가 스스로 막혀요. 설치 폴더 `versions/<v>/bin` 에서 `moai` 가 일반 파일이고 `moai-cli` 가 그 파일을 가리키는 심볼릭 링크면 `bin_names.state` 가 `blocked`(`bin/moai-cli is missing or not a regular file`)로 나와요 (`src/bin_names.rs:150-160`).

두 이름이 같은 설치의 실행 파일 하나를 가리키면 설치된 모양으로 받아들이고 아무것도 쓰지 않아요. 링크가 다른 곳을 가리키면 지금처럼 거절해요. 이 변경은 거절 하나를 푸는 것이라 tier 2 로 다뤄요.

저장소가 추적하는 설치 경로(`install/install.sh:422-424`, `install/install.ps1:509-511`, `src/bin_names.rs:267`)는 이미 하드 링크 모양(`moai-cli` 가 일반 파일, `moai` 가 그 하드 링크)을 써요. 반대 모양을 쓴 절차는 추적 코드에서 찾지 못했어요. 이 Mac 에서는 `4.2648.43-dev` 부터 반대 모양이고 그 앞 버전은 하드 링크 모양이에요 (읽기만 함).

## Non-goals

- 반대 모양을 쓴 추적 밖 설치 절차를 고치는 일 (카드 기준 1, 아래 AC-9).
- 설치된 버전 폴더의 모양을 바꾸는 일. 같은 폴더에서 모양을 바꾸면 그 폴더의 영수증을 못 읽게 돼요 (R-2).
- 설치 식별자(`install::identity_exe`, `install_id`) 변경.
- 새 상태 이름 추가, `src/update/switch.rs` 변경 (`launcher_state` 는 링크를 따라가요).
- 낡은 도움말·문서 글(`src/ops/bin_names_plan.rs:8-12`, `docs/rust-product-path.md:1171`) 고치기: 후속으로 리더가 정해요.
- Windows 와 Linux 확인.

## Requirements

- REQ-1: When `bin/moai-cli` is a symbolic link whose resolved path is the sibling `bin/moai` and that `bin/moai` is a regular file by its own metadata, the name check shall report the state `unchanged`.
- REQ-2: When `bin/moai` is a symbolic link whose resolved path is the sibling `bin/moai-cli` and that `bin/moai-cli` is a regular file by its own metadata, the name check shall report the state `unchanged`.
- REQ-3 (REQ-1): If `bin/moai-cli` is not a regular file and is not the link of REQ-1, then the name check shall report `blocked` with the reason it gives today.
- REQ-4 (REQ-2): If `bin/moai` exists, is not a regular file and is not the link of REQ-2 while `bin/moai-cli` is a regular file, then the name check shall report `blocked` with the reason it gives today.
- REQ-5: When the apply runs on a plan of REQ-1 or REQ-2, the name check shall write nothing, make no hard link, rename nothing and report an empty `written` list. · failure: none — a plan of these states reaches no write step.
- REQ-6: The name check shall judge a symbolic link by its resolved path and the file kind of the target, never by the bytes it leads to.
- REQ-7: When `moai update` runs from a version folder in either linked layout, the update command shall report the `bin_names` state `unchanged` and leave both names as they were. · failure: none — a layout outside REQ-1 and REQ-2 is judged by REQ-3 and REQ-4.

요구사항 설명: 1·2 는 받아들이는 두 모양(상대 링크와 절대 링크 모두), 3·4 는 계속 거절하는 경우예요. 끊어진 링크, 폴더를 가리키는 링크, 바이트가 같은 다른 파일을 가리키는 링크, 다른 버전 폴더나 HOME 밖을 가리키는 링크가 모두 거절이에요. 5 는 쓰기 없음, 6 은 "같은 설치"를 바이트로 판정하지 않는다는 규칙, 7 은 update 전체 경로예요. 일반 파일 두 개를 비교하는 지금의 규칙(`unchanged`, `replace_copy`)은 그대로예요.

## Acceptance

- AC-1 (REQ-1, REQ-5): 상대 링크와 절대 링크 두 경우 모두, 두 이름 어느 쪽으로 실행해도 `unchanged` 이고 apply 의 `written` 이 비며 이름 목록·링크 대상·inode 가 그대로예요. 1 passed · verify: `cargo test --locked --test bin_names_cli ac1_ -- --nocapture`
- AC-2 (REQ-2, REQ-5): `bin/moai` 가 형제 `bin/moai-cli` 를 가리키는 상대·절대 링크에서 AC-1 과 같은 결과예요. 1 passed · verify: `cargo test --locked --test bin_names_cli ac2_ -- --nocapture`
- AC-3 (REQ-3): `moai-cli` 링크가 끊어짐, 폴더, 폴더인 형제 `moai`, 다른 버전 폴더, HOME 밖을 가리키면 `blocked` 이고 아무것도 쓰지 않아요. 계획은 링크가 아닌 일반 파일로 실행해요. 1 passed · verify: `cargo test --locked --test bin_names_cli ac3_ -- --nocapture`
- AC-4 (REQ-3, REQ-6): `moai-cli` 링크가 `bin/moai` 와 바이트가 같은 다른 파일을 가리키면 `blocked` 이고 아무것도 쓰지 않아요. 1 passed · verify: `cargo test --locked --test bin_names_cli ac4_ -- --nocapture`
- AC-5 (REQ-4, REQ-6): `moai` 링크가 끊어짐, 폴더, 바이트가 같은 다른 파일, 다른 버전 폴더를 가리키면 `blocked` 이고 아무것도 쓰지 않아요. 1 passed · verify: `cargo test --locked --test bin_names_cli ac5_ -- --nocapture`
- AC-6 (REQ-5, REQ-7): 임시 HOME 의 버전 폴더에 실제 바이너리를 하드 링크, 반대(상대·절대), 거울 모양으로 놓고 `moai update --json` 을 돌리면 `bin_names.state` 가 `unchanged` 예요. `--confirm` 뒤 `bin_names_result.state` 도 `unchanged` 이고 두 이름은 그대로예요. 1 passed · verify: `cargo test --locked --test bin_names_cli ac6_ -- --nocapture`
- AC-7 (REQ-5): 이름이 하나뿐이면 지금처럼 하드 링크를 만들고 다시 돌리면 `unchanged` 예요. 1 passed · verify: `cargo test --locked --test bin_names_cli create_links_the_alias -- --nocapture`
- AC-8 (REQ-4): 다른 프로그램, 다른 곳의 사본을 가리키는 링크, 폴더인 `bin/moai` 는 계속 `blocked` 예요. 1 passed · verify: `cargo test --locked --test bin_names_cli another_program_a_link_or_a_folder -- --nocapture`
- AC-9 (REQ-7): 반대 모양을 쓴 절차를 찾아 모든 설치 경로가 한 모양을 쓰게 해요 (카드 기준 1) · manual: 추적 코드에는 그 모양을 쓰는 곳이 없고, 리더가 원장 노트 10651 에서 그 모양을 쓴 절차는 리더의 손 설치 순서(빌드 뒤 `bin/moai` 를 상태 폴더의 versions 아래 dev 빌드 폴더로 복사하고 `ln -sfn` 으로 상태 폴더의 `bin/moai` 링크를 바꾸는 순서, 추적된 설치 스크립트가 아님)라고 확인했어요. 그 순서가 다음 빌드 설치부터 제품의 설치 경로(`install/install.sh` 의 하드 링크 모양)를 따르는지를 리더가 읽어 확인하고, 읽은 파일·줄과 결과를 카드 노트에 적어요
- AC-10 (REQ-1, REQ-2, REQ-3, REQ-4, REQ-5, REQ-6, REQ-7): SPEC 검사를 통과해요 · verify: `moai spec check SPEC-20261127-launcher-bin-layout`

## Scope

- write: `.moai/specs/SPEC-20261127-launcher-bin-layout/**`, `src/bin_names.rs`, `tests/bin_names_cli.rs`
- read-only: `src/install.rs`, `src/update.rs`, `src/update/switch.rs`, `src/update/extra_skill.rs`, `src/home_install.rs`, `install/install.sh`, `install/install.ps1`
- forbidden: the real HOME state folder and Claude folder, any other project, any network call. Tests use temporary HOMEs with a cleared environment.

## Interview

- Q1: `moai` 가 일반 파일이고 `moai-cli` 가 링크인 모양을 설치된 모양으로 받아들이나요? → A: 리더 지시예요(세션 간 메시지). 받아들이되 링크가 같은 설치를 가리키는지 확인하고, 다른 곳이면 지금처럼 거절해요.
- Q2: 거울 모양(`moai` 가 `moai-cli` 를 가리키는 링크)도 받아들이나요? → A: 기본값은 받아들임 (D-2). 카드 기준 2 가 "하드 링크나 그것을 가리키는 링크" 라고 적었어요.
- Q3: 반대 모양은 어떤 절차가 썼나요? → A: 리더의 손 설치 순서예요(원장 노트 10651). 추적 코드에는 없어요 (AC-9 는 manual 확인).

## Plan

1. SPEC 초안과 `moai spec check`. tier 2 라서 승인 전에는 아래를 시작하지 않아요. 레인은 승인을 직접 기록하지 않아요.
2. RED: `tests/bin_names_cli.rs` 에 AC-1~AC-6 테스트를 넣고 191-193행의 거울 모양 기대를 바꿔요. base 에서 AC-1, AC-2, AC-6 은 실패하고 AC-3~AC-5, AC-8 은 통과해야 해요.
3. GREEN: `src/bin_names.rs` 에 형제 링크 판정 함수 하나와 분기 두 개를 넣어요.
4. 가드 뮤턴트: (a) 해석 경로 비교를 해석 성공 여부로 바꿈 → AC-3, AC-4, AC-5 실패, (b) 해석 경로 비교를 바이트 비교로 바꿈 → AC-4, AC-5 실패, (c) 대상의 일반 파일 확인을 뺌 → AC-3 실패.
5. 위험 경로: 임시 HOME 에서 실제 바이너리로 받아들이는 모양은 쓰기 0, 다른 곳을 가리키는 링크는 거절임을 Rust 테스트로 보여요.
6. `cargo fmt --all`, AC 의 verify 명령을 적힌 그대로 heavy slot 에서 실행, 바뀐 테스트 대상은 nextest 로도 실행, READY 때 `moai check affected --card <base>..<head>` 한 번.
7. `## Result` 작성 (AC-9 는 met 로 적지 않음). READY 에 tier 2 를 명시해요.

## Risks

- R-1: 거절이 줄어요. 링크로 이어진 일반 파일이 다른 프로그램이어도 `unchanged` 예요 (`src/bin_names.rs:112-116`). 지금의 `--version` 확인(195-217행)은 일반 파일 `moai` 에만 닿아요. 이 상태는 아무것도 쓰지 않아요.
- R-2: 설치 식별자가 모양마다 달라요. 반대 모양은 `bin/moai`, 하드 링크와 거울 모양은 `bin/moai-cli` 예요. 다른 버전 폴더의 영수증은 두 이름 모두 읽지만, 같은 폴더의 모양을 바꾸면 예전 이름의 영수증을 읽지 않아요 (`src/update.rs:1311-1327`, `1367-1368`).
- R-3: 링크가 다른 버전 폴더를 가리키는 이름으로 실행하면 그 폴더가 검사 대상이에요 (지금과 같음). 거절 테스트는 일반 파일로 실행해요.
- R-4: 형제로 해석되는 간접 링크는 받아들여요. 중간 링크가 다른 버전으로 옮겨가면 다시 `blocked` 예요.
- R-5: `--confirm` 테스트는 환경을 비우고 돌려야 해요. 물려받은 `CLAUDE_CONFIG_DIR` 이 있으면 실제 Claude 폴더에 쓸 수 있어요.
- R-6: 이름 필터가 테스트 0개와 맞아도 exit 0 이에요. 각 AC 는 `1 passed` 까지 증거로 적어요.
- R-7: 링크 대상의 대소문자가 디스크와 다를 때의 해석 결과는 확인하지 않았어요. Windows 와 Linux 는 `not_run` 이에요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 반대 모양을 `unchanged` 로 보고하고 쓰지 않음 | 리더 지시. 새 상태 이름은 범위 밖 읽는 쪽이 필요함 | 분기 제거 |
| D-2 | 거울 모양도 받아들임 (기본값). `tests/bin_names_cli.rs:191-193` 의 `blocked` 기대를 다른 곳을 가리키는 링크의 거절 확인으로 바꿈 | 카드 기준 2 | 두 번째 분기 제거, 기대 복원 |
| D-3 | "같은 설치"는 해석 경로가 형제 파일이고 그 파일이 일반 파일일 때만 | 바이트가 같은 사본은 다른 설치 | 규칙 변경은 SPEC 개정 |
| D-4 | 한 모양은 하드 링크. 추적 중인 설치 경로와 설치된 폴더는 바꾸지 않음 | 추적 코드가 이미 그 모양, R-2 | 후속 카드 |
| D-5 | 설치 식별자는 바꾸지 않음 | 폴더 사이 영수증은 두 이름을 이미 읽음 | 후속 SPEC |
| D-6 | 실행 파일이 해석 대상과 같은지는 추가로 확인하지 않음 | R-1 은 쓰기가 없음 | 조건 한 줄 추가 |

## Result

코드 head 는 `93d006f965`(base `fa2b49dc44`)예요. 모든 실행은 2026-11-27 에 macOS, debug 빌드로 heavy 슬롯 안에서 했어요.

레드(`213e64b086`, 테스트만 있는 커밋): `ac1_`, `ac2_`, `ac6_` 세 검사가 각각 exit 101(`test-failure`)로 실패했어요. 실패 이유는 `bin/moai-cli is missing or not a regular file` 과 `bin/moai is a link or a folder` 였어요. 그린(`93d006f965`): 같은 세 검사가 exit 0 으로 통과했고 `moai tdd green` 이 기록했어요.

`93d006f965` 의 코드에서 Acceptance 의 verify 명령을 적힌 그대로(슬롯 래퍼만 붙여) 실행했어요. 테스트 이름마다 `1 passed` 를 읽었어요.

- AC-1: met — `cargo test --locked --test bin_names_cli ac1_ -- --nocapture` 1 passed, exit 0.
- AC-2: met — `cargo test --locked --test bin_names_cli ac2_ -- --nocapture` 1 passed, exit 0.
- AC-3: met — `cargo test --locked --test bin_names_cli ac3_ -- --nocapture` 1 passed, exit 0.
- AC-4: met — `cargo test --locked --test bin_names_cli ac4_ -- --nocapture` 1 passed, exit 0.
- AC-5: met — `cargo test --locked --test bin_names_cli ac5_ -- --nocapture` 1 passed, exit 0.
- AC-6: met — `cargo test --locked --test bin_names_cli ac6_ -- --nocapture` 1 passed, exit 0. 하드 링크, 반대(상대·절대), 거울(상대·절대) 다섯 모양 모두예요.
- AC-7: met — `cargo test --locked --test bin_names_cli create_links_the_alias -- --nocapture` 1 passed, exit 0.
- AC-8: met — `cargo test --locked --test bin_names_cli another_program_a_link_or_a_folder -- --nocapture` 1 passed, exit 0. 거울 모양 기대 하나를 "다른 버전 폴더를 가리키는 링크는 blocked" 로 바꿨어요(결정 D-2).
- AC-9: not met — 반대 모양을 쓴 절차를 찾지 못했어요. 추적 코드에는 쓰는 곳이 없어요(`install/install.sh:422-424`, `install/install.ps1:509-511`, `src/bin_names.rs:267` 은 하드 링크 모양). 이 Mac 에서는 `4.2648.43-dev` 부터 반대 모양이에요(읽기만 함). 카드는 이 기준을 열어 둔 채 랜딩해요(Q1, 노트 10487).
- AC-10: met — `moai spec check SPEC-20261127-launcher-bin-layout` 오류 0, 경고 1(AC-9 가 unknown).

가드 뮤턴트는 `Edit` 로 걸고 `bin_names_cli` 전체를 돌린 뒤 `Edit` 로 되돌렸어요(되돌린 뒤마다 `git diff --exit-code --stat HEAD` exit 0).

- 해석한 경로 비교를 "둘 다 해석됨"으로 약하게 바꿈: `ac4_`, `ac5_`, `another_program_a_link_or_a_folder_is_left_alone` 실패(exit 101).
- 해석한 경로 비교를 바이트 해시 비교로 바꿈: `ac4_`, `ac5_` 실패(exit 101).
- 형제가 일반 파일인지 확인을 뺌: `ac3_` 실패(exit 101).

한계: Windows 와 Linux 는 `not_run` 이에요(심볼릭 링크 테스트는 unix 전용). 링크 대상의 대소문자가 디스크와 다른 경우(R-7)는 확인하지 않았어요. `ac1_` 과 `ac2_` 는 반복마다 단언하므로 레드에서는 첫 반복(상대 링크)만 실패로 실행됐어요.
