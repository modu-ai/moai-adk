---
id: SPEC-20261120-codex-statusline
title: "moai update installs a Codex status line with the elements and order closest to the moai Claude status line"
status: draft
revision: 1
created: 2026-11-20
---

# moai update installs a Codex status line with the elements and order closest to the moai Claude status line

## Intent

2026-11-20T20:3xZ 사용자 지시("codex 의 statusline 인데 claude code statusline 은 똑같이 아니더라도 최대한 제공할 수 있는 배열 순서와 요소들을 추가해서 설치가 되도록 개선해라")에서 나온 카드예요. Codex CLI 는 사용자 config(`$CODEX_HOME/config.toml`, 기본 `~/.codex/config.toml`)의 `[tui] status_line` 항목 이름 목록과 그 순서로 footer 를 정해요(이 기계 `~/.codex/config.toml:865`, 노트 4565 — 14 항목, `status_line_use_colors = true`). 지금 `moai update` 의 사용자 계층(skills·agents·routing·style·bin_names·hooks·keys·home·switch·cleanup)은 Codex 상태줄을 다루지 않고, 프로젝트 계층의 codex-config 스텝(`src/project_hosts.rs`)만 `MOAI_STATUS_LINE_PRESET`(#2928 S-5)을 프로젝트 `.codex/config.toml`에 설치해요. 그 프리셋은 C-203/C-212 이전 Claude 레이아웃 기준이라 지금 레이아웃(워크트리 🌳 가 셋째 줄 끝)과 어긋나요. 이 SPEC 은 (1) 사용자 계층에 Codex 상태줄 설치를 더하고(미리보기 → 사용자 yes → 적용, 사용자 값 절대 덮지 않음, 되돌리기 제공), (2) 프리셋을 현재 Claude 세 줄 레이아웃에서 다시 유도해요.

## Interview

- Q1: Codex 가 상태줄을 어떻게 정하나요? → A: `[tui] status_line` 항목 이름 배열, 순서가 곧 표시 순서예요. 항목 목록은 저장소 기준 41개(`CODEX_STATUS_LINE_ITEMS`, openai/codex rust-v0.169.0 source, #2928 S-5)이고 공식 문서(config-reference)는 항목을 열거하지 않아요 — 공식 문서 재확인은 moai-research 노트로 근거를 남겨요(REQ-13).
- Q2: 사용자 값이 있으면요? → A: 덮지 않아요. Claude statusLine 규칙과 같은 세 갈래: 키가 없으면 설치, MoAI 가 쓴 값(소유 기록)이면 새 프리셋으로 갱신, 그 외 사용자 값이면 미리보기에 키·현재 값·제안 값을 보여 주고 아무것도 쓰지 않아요(노트 4565 "덮어쓰지 않고 제안만").
- Q3: 되돌리기는요? → A: 적용 전 config 파일 전체를 비공개(0600) 복사로 남기고 결과에 복사 경로와 되돌리기 명령을 실어요 — 파일을 통째로 돌리는 것이 키별 값까지 정확히 원래대로 돌려놓아요(Claude status_line 의 cp 되돌리기 선례).
- Q4: config 파일이 아예 없으면요? → A: 제안하지 않아요(skip). Codex 를 쓰지 않는 HOME 에 파일을 새로 만들지 않고, 기존 update UI 골든도 그대로 유지돼요.
- Q5: 프리셋 내용은요? → A: 현재 Claude 레이아웃에서 Codex 가 보여 줄 수 있는 항목만, 식별 → 저장소/브랜치 → 사용량 순서로. 도출 결과는 기존 10항목에서 `thread-name`(🌳 워크트리 자리)을 맨 끝으로 옮긴 것 하나예요. 카드 수·🗿 MoAI 버전·🏷️ 태그·⬇️🔀 forge 수는 Codex 항목이 없어서 매핑 문서에 "불가"로 기록해요.
- Q6: command 항목은요? → A: 연구 노트의 결론을 따라요. 41개 목록에 사용자 명령을 실행하는 항목은 없고, 있어도 이 SPEC 은 프리셋 설치만 다뤄요(확장은 후속 카드).

## Non-goals

- 새 하위명령이나 `moai statusline --host codex` 렌더 추가(연구 결론에 따라 후속).
- 프로젝트 계층 codex-config 스텝의 동작 변경(프리셋 상수만 바뀌고, 소유 갱신 경로는 기존 3-way 그대로).
- `status_line_use_colors` 등 다른 [tui] 키의 읽기·쓰기.
- 사용자 값이 있을 때의 replace 선택지(제안만).

## Requirements

- REQ-1: When the user's Codex config file exists and has no `tui.status_line` key, `moai update` shall preview a Codex status line part that installs the moai preset list.
- REQ-2 (REQ-1): IF the Codex config file is missing, THEN `moai update` shall skip the Codex status line part, naming the reason.
- REQ-3 (REQ-1): IF the user confirms the update with the combined digest, THEN `moai update` shall write the preset list into the config file's `tui.status_line` key.
- REQ-4: When `tui.status_line` holds a value MoAI did not write, the preview shall show the key with the user's value and the proposed value.
- REQ-5 (REQ-4): IF the key holds a value MoAI did not write, THEN `moai update` shall write no byte of the config file at any confirmation level.
- REQ-6: When `tui.status_line` holds a value MoAI wrote (its ownership record), `moai update` shall plan an update of the key to the current preset, applied after the user's yes.
- REQ-7 (REQ-6): IF the recorded value no longer sits in the file, THEN `moai update` shall preview the key as a user value, leaving it unowned.
- REQ-8: The preset list shall be the one derived from the current Claude three-line layout, in the order identity, repository and branch, usage, with `thread-name` last.
- REQ-9: The apply shall keep a private (0600) copy of the config file before its write.
- REQ-10: The preview and the result shall name the config file path, the kept copy, and the command that restores it.
- REQ-11: The layer's plan digest shall bind into the `moai update` combined digest.
- REQ-12: The layer's revert entry shall sit in `revert_order` between `hooks` and `style`.
- REQ-13: The research note on the card shall list every Codex status line item with its source quote and whether each moai item maps to it or cannot be shown.

## Acceptance

- AC-1 (REQ-1, REQ-3): `CODEX_HOME` 을 임시 폴더로 둔 시험에서, config.toml 이 있고 키가 없으면 미리보기가 설치 계획을 내고, `--confirm` 뒤 파일의 키가 프리셋 목록을 담아요.
  verify: `cargo test --locked --offline --test update_user_layer_cli`
- AC-2 (REQ-4, REQ-5): 사용자 값이 있으면 미리보기가 키·현재 값·제안 값을 보여 주고, `--confirm` 해도 파일 바이트가 그대로예요.
  verify: `cargo test --locked --offline --test update_user_layer_cli`
- AC-3 (REQ-6, REQ-7): 소유 기록 값이 파일에 그대로 있으면 update 로 계획되고 yes 뒤 새 프리셋으로 쓰고, 값이 다르면 사용자 값으로 미리보기돼요.
  verify: `cargo test --locked --offline --test update_user_layer_cli`
- AC-4 (REQ-2): config.toml 이 없으면 층이 이유와 함께 skip 이고 요약에 줄이 없어요.
  verify: `cargo test --locked --offline --test update_user_layer_cli`
- AC-5 (REQ-9, REQ-10, REQ-11, REQ-12): 적용 결과에 백업 경로·되돌리기 명령·config 경로가 있고, 백업 모드가 0600 이며, 되돌리기를 돌리면 키 값이 원래대로 돌아오고, 결합 digest 가 층을 묶으며, `revert_order` 가 hooks 와 style 사이를 이름해요.
  verify: `cargo test --locked --offline --test update_user_layer_cli`
- AC-6 (REQ-8, REQ-13): 프리셋 상수가 REQ-8 순서가 되고, 연구 노트(카드)가 항목별 출처 인용과 매핑 가능·불가를 적어요.
  verify: `cargo test --locked --offline --lib codex_config` 와 카드 노트 인용

## Decisions

- D-1: 계층 코드는 `src/hosts/codex_config.rs` 의 기계 위에 사용자 config용 plan/apply 함수로 더해요(새 모듈·lib.rs 줄 편집 없음 — 가장 저렴한 소스; TOML 보존·검증·항목 목록·프리셋이 그 파일에 이미 있음). 배선만 `src/update.rs` 가 해요.
- D-2: 프리셋 재유도: `model, reasoning, codex-version`(줄1: 🧠 🤔 🤖) → `project-name, git-branch, branch-changes, pull-request-number`(줄2: 📡 🅱️ ↑↓ 🔗) → `context-remaining, weekly-limit, thread-name`(줄3: 🔋 CTX, 🔋 Week, 🌳 끝). `MOAI_STATUS_LINE_PRESET` 은 `thread-name` 이 끝으로 옮겨진 10항목이 되고, 매핑표는 docs/codex-adapter-contract.md 가 갱신해요. 사용자의 14항목 목록(노트 4565)은 사용자 값이라 그대로예요.
- D-3: 소유 기록은 update_keys 의 영수증 모양(홈 상태 폴더, `codex-statusline-` 접두어)을 따라요: 적용한 키 값의 digest 를 적고, 다음 계획 때 그 값이 파일에 그대로 있으면 MoAI 소유로 판정해요.
- D-4: 되돌리기는 파일 복사 복원(0600, `write_backup` 선례)이고 결과의 `revert` 에 명령 문자열을 실어요. 새 되돌리기 하위명령은 만들지 않아요.
- D-5: config 파일이 없으면 층이 이유와 함께 skip 이고 요약에 한 줄도 내지 않아요(기존 update UI 골든 불변).
- D-6: 연구(moai-research, 공식 문서)가 끝나면 항목 목록·command 항목 여부를 확인하고, 목록이 0.171.0 에서 늘었다면 `CODEX_STATUS_LINE_ITEMS` 갱신을 같은 커밋에서 검토해요(연구 노트 근거).

## Scope

- write:
  - .moai/specs/SPEC-20261120-codex-statusline/spec.md
  - src/hosts/codex_config.rs
  - src/update.rs
  - docs/codex-adapter-contract.md
  - tests/update_user_layer_cli.rs
  - tests/update_cli.rs

- read-only:
  - 그 밖의 모든 것 — 사용자의 실제 Codex config 폴더 포함(시험은 임시 CODEX_HOME 만)

## Plan

1. SPEC 작성과 `spec check`(빌드 바이너리).
2. 연구 노트 확인 → 프리셋·항목 목록 확정(필요 시 SPEC 개정).
3. RED: `tests/update_user_layer_cli.rs` 에 임시 `CODEX_HOME` 시험(없음→설치, 사용자 값→보존, 소유→갱신, 백업·되돌리기, 요약·revert_order, 파일 없음→skip).
4. GREEN: `codex_config.rs` 사용자 config plan/apply + `update.rs` 배선.
5. `MOAI_STATUS_LINE_PRESET` 순서 갱신 + docs/codex-adapter-contract.md 매핑표 갱신 + 프로젝트 계층 기존 시험 확인.
6. 검사: `moai check affected` + 영역 시험 무필터, 종료 코드를 READY 노트에.

## Risks

- R1: `CODEX_HOME` 해석(host_home::variable_dir)의 빈 값·상대 값 규칙(C-79)을 어기면 사용자 폴더를 잘못 읽어요 → 변수 규칙 시험을 먼저 두고 걸치게 해요.
- R2: 사용자 config 는 비밀(MCP env)을 담을 수 있어요 → 미리보기는 `tui.status_line` 값만 보여 주고, 백업은 0600 이에요.
- R3: 프리셋 상수 변경이 프로젝트 계층 기존 시험과 골든을 깰 수 있어요 → 영역 시험 무필터로 확인해요.
- R4: 결합 digest 에 층이 더해져 기존 update 시험이 digest 불일치로 깨질 수 있어요 → 해당 시험을 같은 커밋에서 갱신해요.

## Result

(감사 PASS 뒤 작성)
