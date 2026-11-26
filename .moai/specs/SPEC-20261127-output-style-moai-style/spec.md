---
id: SPEC-20261127-output-style-moai-style
title: "설치되는 출력 스타일 이름을 MoAI-Style 로 바꾸고, 영어가 아닌 언어의 AI 번역투를 원칙과 윤문 단계로 잡음"
status: draft
revision: 1
created: 2026-11-27
---

# 설치되는 출력 스타일 이름을 MoAI-Style 로 바꾸고, 영어가 아닌 언어의 AI 번역투를 원칙과 윤문 단계로 잡음

## Intent

카드 C-592 의 사용자 지시 두 가지예요 (원문은 카드 본문과 노트 10532).

하나, 설치되는 출력 스타일의 이름이 `MoAI-Easy` 에서 `MoAI-Style` 로 바뀌어요. 파일은 `output-styles/MoAI-Style.md`, 설정은 `"outputStyle": "MoAI-Style"` 이에요. 이미 `MoAI-Easy` 를 설치한 사용자의 HOME 은 `moai update` 가 옮겨요: 새 파일을 만들고, 옛 파일은 되돌리기 기록에 바이트째 보관한 뒤 `output-styles` 에서 치우고, `outputStyle` 값이 정확히 `MoAI-Easy` 이면 `MoAI-Style` 로 바꿔요. `home-style-revert` 가 세 가지를 원래대로 돌려놓아요. 사용자 HOME 의 파일 이름과 설정 값을 건드리는 일이라 tier 2 예요.

둘, 스타일 본문이 영어가 아닌 언어에서 AI 번역투를 막아요. 사용자 지시는 **단어 목록을 넣지 말고, 언어와 상관없는 원칙과 윤문 단계를 넣는 것**이에요 (노트 10532). 원칙은 "그 언어 개발자가 실제로 쓰는 표준어로 쓴다"이고, 윤문은 "보내기 전에 그 언어 개발자가 읽는다고 생각하고 다시 읽어 번역투를 고쳐 쓴다"예요. 번역투 낱말표(노트 10529)는 스타일에 넣지 않고 평가 자료(eval)와 테스트 픽스처에만 써요.

이 SPEC 이 랜딩된 뒤 처음 `moai update` 를 돌릴 때, 사용자 자신의 Claude 폴더(이 세션이 쓰는 스타일)가 처음으로 옮겨져요. 그때 사용자는 미리보기에서 이 계획을 먼저 봐요. 이번 레인과 테스트는 임시 HOME 만 써요.

## Non-goals

- 영어 이름 `audit` → `review` 바꾸기 (노트 첫 줄, 에이전트 이름, 문서, 게이트 검사). C-593 이 따로 맡아요.
- `CLAUDE.md` 고치기. 거기 있는 한국어 단어 목록(`CLAUDE.md:106` 근처)은 팀 규칙이고, 바꾸려면 `CLAUDE.md` 를 쓰기 범위에 둔 카드와 사용자의 직접 입력이 필요해요 (Q6).
- 일본어·중국어 예시 단어. 원어민이나 공식 출처로 확인한 것만 평가 자료에 넣어요. 노트 10529 의 일본어·중국어 예는 출처로 확인하지 않은 일반 지식이에요.
- 프로젝트의 `.claude/settings*.json` 고치기 (지금처럼 읽기만 하고 알림만 내요).
- 다른 이름의 스타일 파일(`MoAI.md`, `MoAI-Learn.md`)이나 사용자가 직접 만든 스타일을 건드리는 일.

옛 이름을 일부러 남기는 곳 (AC-12 의 허용 목록과 같아요):

- Go 세대 감지: `src/go_state.rs` 의 스타일 이름 비교, `src/claude_settings_merge.rs` 의 `go_generated` 테스트, `tests/go_state_*.rs`, `tests/go_project_real_run_cli.rs`. Go 세대 설치가 실제로 `MoAI-Easy` 를 썼기 때문이에요.
- 해시로 고정한 이전 렌더: `tests/fixtures/home-style/render-*.md` 와 `EARLIER_RENDERS`. 새 이름 앞의 글을 기록한 자료라 고치지 않고, 이번 변경 직전 렌더의 해시 하나만 보태요.
- 금지 배너: `tools/repo_scans.json` 과 `tests/style_banner_scan.rs` 의 `MoAI-Easy ★`.
- 상류 템플릿: `templates/claude/output-styles/moai-easy.md` 와 `NOTICE.md`. Apache-2.0 적응본이고 Go 세대 감지가 이 파일 이름을 읽어요 (Q5).
- 읽기만 하는 기록: 지난 SPEC, `docs/design-review/**`, `.moai-cli/**`.
- 마이그레이션의 원본: `src/home_style.rs` 의 옛 이름 상수, 옛 이름 테스트, 되돌리기 기록 호환 코드, html-report 의 옛 이름 줄.
- Python: `tests/ci_scope_test.py`. 이 저장소에서 Python 은 고치지도 돌리지도 않아요.

## Requirements

- REQ-1: The installer shall name the output style `MoAI-Style`; the style file is `output-styles/MoAI-Style.md` with `name: MoAI-Style` in its frontmatter, and the `outputStyle` value it writes is `MoAI-Style`.
- REQ-2: When the config folder holds `output-styles/MoAI-Easy.md` that is MoAI's own (the state record names its bytes, or its bytes equal a text MoAI shipped under that name), the plan shall list that file as a `retire` target. · failure: none — a file that is not MoAI's own is REQ-5, and a refused folder or link is REQ-9.
- REQ-3: The apply shall keep the bytes of a `retire` target in the receipt before it removes that file from `output-styles`.
- REQ-4: When the `outputStyle` value in `settings.json` is exactly the string `MoAI-Easy` and the old style file is retired or absent, the plan shall list that value as a `replace` target, and the apply changes that one value as text so that every other byte stays as the user wrote it. · failure: none — any other value keeps the behaviour of today (absent is set; a case variant or another name is a `conflict` and is not changed).
- REQ-5 (REQ-2): If the old file `output-styles/MoAI-Easy.md` exists and is not MoAI's own, then the plan shall mark it `conflict`.
- REQ-6 (REQ-5): If the old file is a `conflict`, then the apply shall leave that file and the `outputStyle` value as they are and still write the new style file.
- REQ-7 (REQ-5): If the old file is a `conflict`, then the update shall show a notice that names that file and says how to choose the style.
- REQ-8 (REQ-2): If the new style file is a `conflict` or `blocked`, then the plan shall leave the old file and the `outputStyle` value as they are.
- REQ-9 (REQ-2): If the folder or the old file is refused by the rules of the style file (a link that is not followed, a folder that is not plain), then the plan shall mark the old file `blocked` with the same reason text.
- REQ-10: The apply shall write in the order new style file, setting, old-file removal, and record each entry in the receipt before its write.
- REQ-11: When a revert runs on a receipt that retired a file, the revert shall restore the old file's bytes, remove the new style file and put the old `outputStyle` text back, each only while the target still holds what the apply wrote. · failure: none — a target changed since the apply is REQ-12.
- REQ-12 (REQ-11): If a target changed since the apply, then the revert shall skip that target and name it in `skipped`.
- REQ-13: The revert shall treat a receipt entry that records no path as an entry of the earlier name (`output-styles/MoAI-Easy.md`, value `MoAI-Easy`) and never write the new file name for it.
- REQ-14: When the old file is absent, the plan shall list no old-file target, so that a plan without an old file has the same targets as a plan before this change apart from the new names. · failure: none — there is nothing to move.
- REQ-15: The style source shall state, for any user language other than English, one principle that holds for every language: write the standard wording that working developers of that language and region actually use for software work, which may be a loanword, a native word or the English term itself, and never a dictionary rendering of the English word.
- REQ-16: The style source shall state a polish step: before sending a reply in a language other than English, the writer rereads the reply as a developer who speaks that language and rewrites what reads as translated (a dictionary rendering of an English word, English word order, a string of nouns, an awkward passive, a particle or spacing that breaks the orthography of that language).
- REQ-17: The style source shall hold no word list and no per-language example words.
- REQ-18: The style source shall keep identifiers, commands, paths and protocol lines as written.
- REQ-19: The style source shall ask for a gloss of a technical term only when the reader may not know it, and it never asks the writer to coin a native word.
- REQ-20: The translationese-to-polished pairs shall live in an eval fixture under `.moai/evals/style/`, outside the skill package and outside the manifest, with a source note for each Korean pair and no pair for a language whose wording a native speaker or an official source has not confirmed.
- REQ-21: The html-report tier map shall resolve the output style `MoAI-Style` to the `basic` tier and keep `MoAI-Easy` resolving to `basic` as the earlier name, in the skill and in its bundle-pack copy.
- REQ-22: The update screens shall count a `retire` target and a `replace` target among the changes of the style row.

## Acceptance

- AC-1 (REQ-1): 임시 HOME 에 처음 설치하면 계획이 `output-styles/MoAI-Style.md`(frontmatter `name: MoAI-Style`) 만들기와 `outputStyle` `MoAI-Style` 넣기를 말하고, 적용 뒤 `MoAI-Easy.md` 는 없어요. 1 passed · verify: `cargo test --locked --test home_style_cli ac1_ -- --nocapture`
- AC-2 (REQ-2, REQ-3, REQ-4, REQ-10, REQ-14): 임시 HOME 에 옛 파일(기록과 같은 바이트)과 `"outputStyle": "MoAI-Easy"` 와 다른 키가 있으면 계획이 옛 파일 `retire` 와 값 `replace` 를 말하고, 적용이 새 파일을 만들고 값 텍스트 하나만 바꾸며(다른 바이트는 그대로) 옛 파일을 `output-styles` 에서 치우고, 되돌리기 기록의 옛 바이트 sha256 이 옛 파일과 같고, 기록 파일이 새 파일의 sha256 을 가리켜요. 옛 파일이 없는 HOME 의 계획에는 옛 파일 대상이 없어요. 1 passed · verify: `cargo test --locked --test home_style_cli ac2_ -- --nocapture`
- AC-3 (REQ-11, REQ-12): AC-2 의 되돌리기가 옛 파일 바이트, 새 파일 제거, 설정 파일 바이트, 기록의 옛 sha256 을 되돌려요. 적용 뒤 사용자가 설정 파일의 다른 곳을 고친 경우는 값 텍스트 하나만 되돌리고, 새 파일이 바뀐 경우는 그 대상만 건너뛰고 이름을 말해요. 1 passed · verify: `cargo test --locked --test home_style_cli ac3_ -- --nocapture`
- AC-4 (REQ-5, REQ-6, REQ-7): 사용자가 고친 `MoAI-Easy.md` 가 있으면 옛 파일 대상이 `conflict` 이고, 적용 뒤 옛 파일과 `outputStyle` 이 바이트째 그대로이며 새 파일은 만들어지고 알림이 그 파일을 말해요. 1 passed · verify: `cargo test --locked --test home_style_cli ac4_ -- --nocapture`
- AC-5 (REQ-4): 옛 파일이 없고 `outputStyle` 이 `MoAI-Easy` 이면 값이 바뀌어요. 값이 없으면 지금처럼 넣고, `moai-easy`·`MOAI-EASY` 와 다른 이름은 지금처럼 `conflict` 로 두어 바꾸지 않아요. 1 passed · verify: `cargo test --locked --test home_style_cli ac5_ -- --nocapture`
- AC-6 (REQ-13): 이 변경 앞 형식의 되돌리기 기록(경로 없는 항목, `outputStyle` 텍스트가 `MoAI-Easy`)과 그 모양의 HOME 픽스처에서 되돌리기가 옛 경로와 옛 값으로 동작하고, 새 파일 이름은 한 번도 쓰지 않아요. 1 passed · verify: `cargo test --locked --test home_style_cli ac6_ -- --nocapture`
- AC-7 (REQ-8): 사용자가 만든 `MoAI-Style.md`(MoAI 것이 아님)가 있으면 새 파일 대상이 `conflict` 이고, 옛 파일과 `outputStyle` 은 바이트째 그대로예요. 1 passed · verify: `cargo test --locked --test home_style_cli ac7_ -- --nocapture`
- AC-8 (REQ-9): `CLAUDE_CONFIG_DIR` 프로필에서 `output-styles` 가 기본 폴더로 링크되어 있어도 옛 파일이 같은 링크 규칙으로 옮겨지고, 규칙이 거절하는 링크(기본 폴더 밖, 끊어진 링크)는 옛 파일 대상이 같은 사유로 `blocked` 이며 아무것도 옮겨지지 않아요. 1 passed · verify: `cargo test --locked --test home_style_cli ac8_ -- --nocapture`
- AC-9 (REQ-1, REQ-2, REQ-4, REQ-11, REQ-22): 환경을 비운 `moai update --json` 이 옛 HOME 의 옮기기를 미리보기로 보여 주고(스타일 줄이 `retire` 와 `replace` 를 변경으로 세요), `--confirm` 이 AC-2 의 결과를 내고, 다시 돌리면 `unchanged` 이고, 되돌리기 기록 id 로 `home-style-revert` 가 되돌려요. 실제 HOME 은 쓰지 않아요. 1 passed · verify: `cargo test --locked --test update_user_layer_cli ac9_ -- --nocapture`
- AC-10 (REQ-15, REQ-16, REQ-17, REQ-18, REQ-19): 스타일 원본에 원칙 문장(REQ-15)과 윤문 단계 문장(REQ-16)이 있고, 낱말표나 언어별 예시 낱말 줄이 없고(평가 픽스처의 번역투 낱말이 본문에 하나도 없어요), 식별자·명령·경로·프로토콜 줄을 그대로 둔다는 문장이 있고, 용어 풀이 규칙이 새 낱말을 지어내라고 하지 않아요. 1 passed · verify: `cargo test --locked --test home_style_ownership_cli ac10_ -- --nocapture`
- AC-11 (REQ-20): `.moai/evals/style/` 에 README, 번역투→윤문 쌍(`cases.json`), 대조군(`controls.json`)이 있고, 한국어 쌍마다 출처 노트가 있고, 확인 안 된 언어의 쌍은 없고, 스킬 패키지와 manifest 에 들어가지 않아요. 1 passed · verify: `cargo test --locked --test home_style_ownership_cli ac11_ -- --nocapture`
- AC-12 (REQ-1, REQ-21): 옛 이름 `MoAI-Easy` 는 위 Non-goals 의 허용 목록에 있는 파일에만 나타나요(허용 목록은 테스트 안에 파일과 이유로 적어요). 1 passed · verify: `cargo test --locked --test style_banner_scan ac12_ -- --nocapture`
- AC-13 (REQ-21): html-report 스킬과 bundle-pack 사본의 계층 표가 `MoAI-Style` → `basic` 이고 `MoAI-Easy` 도 `basic` 으로 남아 있어요. 1 passed · verify: `cargo test --locked --test style_banner_scan ac13_ -- --nocapture`
- AC-14 (REQ-15, REQ-16): 번역투 문장을 담은 요청에 대해 한국어 답이 평가 픽스처의 기대대로 윤문되어 나와요 · unknown: 모델을 실제로 돌리는 평가 실행기가 아직 없어요. 사람이나 모델 채점으로 픽스처를 한 번 읽어 확인하는 일은 후속 카드예요.
- AC-15 (REQ-1…REQ-14): 이웃 테스트가 그대로 통과해요 · verify: `cargo test --locked --test host_home_cli` 와 `cargo test --locked --test update_linked_config_files` 와 `cargo test --locked --test home_style_cli`
- AC-16 (REQ-1…REQ-22): SPEC 검사를 통과해요 · verify: `moai spec check SPEC-20261127-output-style-moai-style`

## Scope

- write: `.moai/specs/SPEC-20261127-output-style-moai-style/**`, `.moai/evals/style/**`, `src/home_style.rs`, `src/ops/home_style_plan.rs`, `src/ops/home_style_apply.rs`, `src/ops/home_style_revert.rs`, `src/update_tux.rs`, `src/update.rs`, `src/host_home.rs`, `skills/moai/references/*/output-style.md`, `skills/moai-html-report/SKILL.md`, `templates/bundle-packs/.claude/skills/moai-domain-html-report/SKILL.md`, `tests/home_style_cli.rs`, `tests/home_style_ownership_cli.rs`, `tests/style_banner_scan.rs`, `tests/update_user_layer_cli.rs`, `tests/update_linked_config_files.rs`, `tests/host_home_cli.rs`, `tests/e2e_temp_home_cli.rs`, `tests/product_cli.rs`, `tests/fixtures/home-style/**`, `tests/fixtures/update_tux/**`, `tests/fixtures/statusline/**`
- shared files (쓰기 범위가 아니라 `card line-lease` 로 짧게 써요): `skills/moai/references/source-manifest.json`(스타일 원본의 sha256 과 reason), `tests/fixtures/help/*.txt`(`home-style-plan` 줄), `tools/repo_scans.json`(필요할 때만). `skills/moai/references/*/output-style.md` 는 `home` 폴더의 한 파일이에요.
- read-only: `src/go_state.rs`, `src/claude.rs`, `src/claude_settings_merge.rs`, `src/home_layout.rs`, `src/state_root.rs`, `templates/claude/output-styles/**`, `CLAUDE.md`
- forbidden: 실제 HOME 아래의 Claude 폴더와 상태 폴더, 다른 프로젝트, 네트워크 호출, Python. 테스트는 임시 HOME 과 비운 환경(`CleanEnv`)만 써요.

## Interview

- Q1: 이름은 무엇이고 옛 이름은 어떻게 되나요? → A: 사용자 결정(노트 10530)이에요. 이름은 `MoAI-Style`, 파일도 `MoAI-Style.md` 예요. `moai update` 가 `MoAI-Easy.md` 를 옮기고 설정의 `outputStyle` 도 고쳐요. 옛 파일은 되돌리기 기록과 함께 보관하고 지우지 않아요.
- Q2: 어디까지 고치나요? → A: 사용자 결정(노트 10530): 분명한 오역과 흐름 낱말 모두예요. 조사는 붙여 써요(`tip을`).
- Q3: 번역 오류는 어떻게 막나요? → A: 사용자 직접 입력(노트 10532, 원문 그대로): "이런식으로 영어 -> 일본어, 중국어, 한국어 등 사용자 언어가 영어가 아닐 경우 현지 정서에 맞는 표준어로 모두 변경하도록 지침을 수정하자. 단어 몇개 하드 코딩 금지. AI 번역체를 윤문 처리를 하도록 해야 한다." 그래서 원칙과 윤문 단계만 스타일에 넣고 단어표는 평가 픽스처에만 써요.
- Q4: 옛 파일은 `output-styles` 에서 치우나요, 그대로 두나요? → A: 확인이 필요해요 (사용자 승인 시 정해요). 기본값은 **치우되 되돌리기 기록에 바이트째 보관**이에요. 노트 10530 의 "보관하고 지우지 않아요"를 그렇게 읽었어요. 그대로 두면 Claude Code 의 스타일 목록에 이름이 둘 보이고 옛 것은 더 이상 갱신되지 않아요. "그대로 두기"를 고르면 REQ-2 와 REQ-3 의 제거 단계, REQ-10 의 마지막 단계, AC-2·AC-3 의 제거와 복원 부분이 빠지고 값 바꾸기만 남아요.
- Q5: `templates/claude/output-styles/` 의 세 파일(`moai.md`, `moai-easy.md`, `moai-learn.md`)은요? → A: 기본값은 **그대로 둠**이에요. 상류 적응본이고 NOTICE 에 이름이 적혀 있고 Go 세대 감지가 이 파일 이름을 읽어요. 바꾸려면 사용자가 말해 주세요.
- Q6: `CLAUDE.md` 의 한국어 단어 목록은요? → A: 이 SPEC 의 범위 밖이에요. 사용자 지시는 스타일에 단어 목록을 넣지 말라는 것이고, `CLAUDE.md` 는 사용자가 레인 창에 직접 입력해서만 바뀌어요. 필요하면 후속으로 사용자가 정해요.
- Q7: 스타일을 이미 켜 둔 실행 중인 세션은요? → A: 새 세션부터 바뀐 이름을 읽어요. 지금 켜진 세션은 이름이 `MoAI-Easy` 로 보여요. html-report 의 옛 이름 줄이 그 세션을 받아요.

## Plan

1. SPEC 초안과 `moai spec check`. tier 2 라서 승인 전에는 아래를 시작하지 않아요. 레인은 승인을 직접 기록하지 않아요. 승인 뒤 쓰기 범위를 위 Scope 로 넓혀요.
2. 먼저 읽기: `src/update_tux.rs` 의 스타일 줄(`tally`, `change_mark`)과 `src/update.rs` 의 `same_writes` 쪽 쓰기 판정에 새 행동 단어 `retire`·`replace` 를 넣는 자리, `src/ops/home_style_*` 의 JSON 모양, C-460 의 변경 범위. C-460 이 `record_name` 과 같은 파일을 건드리므로 랜딩 순서는 리더가 정해요. 같은 줄을 동시에 고치지 않아요.
3. RED: AC-1…AC-13 의 테스트와 새 픽스처(옛 형식 되돌리기 기록, 옛 이름 HOME)를 먼저 쓰고 실패를 눈으로 봐요. 옛 이름 상수와 새 이름 상수는 한 곳에 두어요.
4. GREEN 1: 이름 바꾸기. `STYLE_NAME`, 스타일 파일 경로, 렌더 이름, 설명 글. 옛 이름 상수는 마이그레이션 원본으로만 남겨요. `EARLIER_RENDERS` 에 이번 변경 직전 렌더의 해시를 보태요.
5. GREEN 2: 옮기기. 옛 파일 대상(없으면 계획에 넣지 않아요), 설정 대상의 값 바꾸기(키마다 바꿀 값과 되돌릴 텍스트를 기록해요: 지금의 `SettingKey`·`InsertedKey` 에 선택 필드를 더해요), 적용 순서와 항목 기록, 되돌리기 기록 항목에 경로를 적는 선택 필드와 옛 형식 기본값, `entry_keys` 의 옛 형식 `outputStyle` 값 비교를 옛 이름으로, `revert_in` 의 새 종류. 기록 파일의 소유 바이트는 새 파일을 가리키고 되돌리면 옛 바이트로 돌아와요.
6. GREEN 3: 스타일 본문. 원칙, 윤문 단계, 용어 풀이 규칙 다시 쓰기. 평가 픽스처 `.moai/evals/style/`(README, cases, controls: `.moai/evals/audit/` 의 모양).
7. 곁가지: html-report 두 곳, 도움말 골든 줄, 업데이트 미리보기 픽스처, manifest 의 sha256 과 reason(`card line-lease` 로 짧게).
8. 가드 뮤턴트(2단계): (a) 옛 파일이 MoAI 것인지 확인을 뺌, (b) 값 바꾸기의 "정확히 `MoAI-Easy`" 조건을 느슨하게, (c) 새 파일이 막혔을 때 옛 파일을 그대로 두는 가드를 뺌, (d) 되돌리기에서 "아직 적은 그대로인가" 확인을 뺌, (e) 옛 형식 항목의 기본 경로를 새 이름으로 바꿈. 각각 한 번에 하나씩 `Edit` 로 걸고 테스트가 실패하는지 보고 되돌려요.
9. 위험 경로: 임시 HOME 에서 실제 바이너리로 옮기기와 되돌리기를 돌려요(`CleanEnv`). 실제 Claude 폴더는 건드리지 않아요.
10. `cargo fmt --all`, 각 AC 의 verify 명령을 적힌 그대로 heavy slot 에서, 바뀐 테스트 대상은 nextest 로도, READY 때 `moai check affected --card <base>..<head>` 한 번.
11. `## Result` 작성 (AC-14 는 met 로 적지 않음). READY 에 tier 2 를 명시해요.

## Risks

- R-1: 사용자 HOME 의 파일 이름과 설정 값을 바꿔요. 되돌리기 기록의 항목은 "적은 그대로일 때만" 되돌려요. 사용자가 그 사이 고쳤으면 건너뛰고 이름을 말해요.
- R-2: 옛 형식 기록의 `outputStyle` 텍스트 되돌리기가 지금은 `STYLE_NAME` 으로 값을 비교해요(`entry_keys`). 이름을 바꾸면서 옛 기록을 위한 비교를 옛 이름으로 두지 않으면 그 되돌리기가 조용히 막혀요. AC-6 이 막아요.
- R-3: 실행 중인 세션과 다른 도구가 `MoAI-Easy` 를 이름으로 읽어요. 옛 이름 줄을 일부러 남겨요 (Q7, REQ-21).
- R-4: 대상이 하나 늘어서 미리보기 JSON 의 모양과 `update_tux` 줄이 달라져요. 옛 파일이 없으면 대상이 없으므로 새 설치의 미리보기는 이름만 달라요 (REQ-14).
- R-5: C-460 이 같은 파일(`src/home_style.rs` 의 `record_name`, `update_tux` 의 스타일 줄)을 바꿔요. 랜딩 순서와 머지는 리더가 정해요.
- R-6: 원칙 문장은 영어로 쓴 본문이 모델에게 주는 지시예요. 모델이 원칙을 따르는지는 테스트가 아니라 평가 실행이 보여 주고, 그 실행기가 아직 없어요 (AC-14 unknown).
- R-7: 일본어·중국어에서 원칙이 맞게 동작하는지는 확인하지 않았어요. 원칙은 언어를 가리지 않게 썼고 예시 단어를 넣지 않았어요.
- R-8: 이 변경이 랜딩되고 나서 사용자의 첫 `moai update` 가 사용자 HOME 을 처음 옮겨요. 미리보기를 먼저 보여 주고 `--confirm` 은 사용자의 확인 뒤예요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 이름은 `MoAI-Style`, 파일도 `MoAI-Style.md` | 사용자 결정(노트 10530) | 상수 두 개 |
| D-2 | 옛 파일이 MoAI 것이면 되돌리기 기록에 보관하고 치움 (Q4 의 기본값) | 노트 10530 "보관하고 지우지 않아요" | `home-style-revert` |
| D-3 | 값은 정확히 `MoAI-Easy` 일 때만 바꿈. 대소문자만 다르거나 다른 이름은 지금처럼 | 사용자의 선택을 덮어쓰지 않음 | 조건 한 줄 |
| D-4 | 사용자가 고친 옛 파일은 그대로 두고 알림 | 지우거나 덮어쓰지 않음(상시 규칙) | 없음 |
| D-5 | 옛 파일이 없으면 계획에 옛 파일 대상을 넣지 않음 | 새 설치의 계획이 달라지지 않게 | 없음 |
| D-6 | 되돌리기 기록 항목에 경로를 적는 선택 필드를 더하고, 없으면 옛 경로와 옛 값 | 옛 기록이 계속 되돌려져야 함 (REQ-13) | 필드 제거 |
| D-7 | 스타일 본문에 단어표를 넣지 않고 번역투 쌍은 평가 픽스처에만 | 사용자 지시(노트 10532) | 없음 |
| D-8 | 평가 픽스처는 한국어 쌍만, 출처 노트를 붙여 | 확인 안 된 일본어·중국어 예를 넣지 않음 | 쌍 추가 |
| D-9 | 상류 템플릿 세 파일은 그대로 (Q5 의 기본값) | NOTICE 와 Go 세대 감지 | 후속 카드 |
| D-10 | 모델이 원칙을 따르는지는 AC-14 `unknown` 으로 두고 met 로 적지 않음 | 평가 실행기 없음 | 실행기가 생기면 verify 추가 |
| D-11 | 새 행동 단어는 `retire`(옛 파일)와 `replace`(설정 값)이고, 화면의 `change_mark` 와 `tally` 가 변경으로 세요 | `update_tux` 가 아는 단어에 `replace` 가 이미 있음 | 단어 제거 |
