---
id: SPEC-20261121-moai-name
title: "The temporary name moai-cli leaves the product: every remaining occurrence, by kind and slice"
status: draft
revision: 1
created: 2026-11-21
---

# The temporary name moai-cli leaves the product

## Intent

사용자 지시(2026-11-21, 리더 대화 원문): "moai-cli 전수 조사해서 또 수정에 빠져 있는거 찾아서 완전 제거해라". 임시 이름 `moai-cli`(바이너리·패키지·크레이트·폴더·변수·릴리스 이름)는 제품이 `moai` 로 완전히 걷어낀다. 조사(tip 6e1bedb4f3, 읽기만): **999 파일·6969 줄**이 moai-cli 또는 moai_cli 를 담고 있어요(리더 scratchpad mcli/classified2.tsv — 19 종류 분류).

이 SPEC 은 카드의 조각 1 이에요: 남은 발생 전부를 종류별·조각별로 목록화하고, 버전 시그니처와 스킬 fallback 문구가 `moai` 를 이름하되 읽는 쪽은 두 이름을 모두 받아요.

## Non-goals

- 다른 카드가 이미 쥔 것은 여기서 다시 하지 않아요: 상태 폴더 기본값과 HOME 마이그레이션(path-root 카드 — src 의 resolver 3 개와 src/lib.rs 상수, `.moai-cli` 상태 폴더 경로 대부분), 프로젝트 측 폴더 이동, Go 링크와 name-switch 스크립트 제거(Go-link 레인 — install/commit-msg 포함), Python 파일(Rust-only 카드), Codex/OMP 제거(C-290·C-291·C-292·C-313).
- 조각 4(캐시·임시 폴더 이름)와 조각 10(crate 이름)은 레인이 멈춘 뒤 리더가 예약해요.
- 키 볼트 유도선(카드 본문 9)은 기존 볼트 항목을 읽을 수 없게 만들지 않으려 불변으로 두고 SPEC 이 결정을 기록해요(guarded — 볼트 내용을 열지 않아요).

## Requirements

- REQ-1: The version output first line shall name `moai`.
- REQ-2: The same-build check of the binary names shall accept a version line naming either binary.
- REQ-3: The install generation reader shall accept a version line naming either binary.
- REQ-4: The skill fallback text shall verify the Rust signature naming `moai`.
- REQ-5: The skill fallback text shall accept the older printed name on read.
- REQ-6: Every remaining occurrence of the old name shall be listed by kind with its slice in the occurrence table of this SPEC.

## Occurrence table (survey tip 6e1bedb4f3, classified2.tsv)

| Kind | Lines | What it is | Slice |
| --- | --- | --- | --- |
| i-a state folder in tests | 1499 | tests 의 `.moai-cli` 상태 폴더 경로 | 8(path-root 뒤; resolver 로) |
| f tracked `.moai-cli` tree + ignores | 1284 | tracked 상태 트리 157 파일·ignore·scan 설정 | 7(path-root 착지 뒤) |
| h other docs/history | 974 | 문서·design-review·specs 이력 | 11(이력은 이력으로 유지) |
| c1 crate path in src | 695 | `moai_cli::` crate 경로(src) | 10(리더 예약, 레인 멈춤) |
| i-c crate path in tests | 683 | `moai_cli::`(tests) | 10 |
| k comments in src | 392 | src 주석 | 11 |
| j Python | 252 | Python 파일 | Rust-only 카드(제외) |
| i-o binary word in tests | 348 | tests 의 바이너리 이름 그 밖 | 8 |
| i2 unit tests in src | 206 | src 내부 단위 시험 | 10 |
| g skills/templates/agents | 392 | 스킬·템플릿 텍스트 | 1(fallback 문구)·2·11 |
| i-d env/file names in tests | 187 | tests 의 env·파일 이름 | 3 |
| i3 fixtures+goldens | 155 | fixture·golden | 2 |
| e installer+release | 110 | install/·.github/·scripts/·dist-workspace | 5·6 |
| e2 dev scripts+config | 45 | .cargo/·.config/ 개발 설정 | 3·5 |
| c2 binary-name literal in src | 43 | src 의 `moai-cli` 리터럴 | 1(버전·clap)·2(사용자 문자열) |
| d env/file names in src | 34 | `MOAI_CLI_` 계열(src) | 3 |
| b user-visible strings | 20 | help·거절 메시지(`.moai-cli` 상태 경로 중심) | 2(상태 경로분은 path-root 뒤) |
| c0 manifests | 10 | Cargo.toml·lock·.cargo | 5 |
| a state folder in src | 6 | src 의 `.moai-cli` 상태 폴더 | path-root 카드(제외) |

## Acceptance

- AC-1 (REQ-1): `--version` 의 첫 줄이 `moai <version> (rust)` 로 시작하고 usage·help 이름이 `moai` 이에요.
  verify: `cargo test --locked --offline --test bin_names_cli` 의 version_line 어설션과 `src/main.rs` 의 clap name.
- AC-2 (REQ-2): 같은 빌드 검사가 두 이름의 버전 줄을 모두 같은 빌드로 받아요.
  verify: `cargo test --locked --offline --lib bin_names::tests::the_same_build_check_accepts_both_binary_names`
- AC-3 (REQ-3): 설치 세대 판별이 두 이름의 버전 줄을 모두 rust 로 받아요 — 기존 코드가 이미 그렇고 시험이 그것을 고정해요.
  verify: `cargo test --locked --offline --test install_cli` 의 generation 시험.
- AC-4 (REQ-4): 스킬 fallback 문구가 `moai <version> (rust)` 를 검사해요.
  verify: `grep -n "moai <version>" skills/moai/SKILL.md` 가 있어요.
- AC-5 (REQ-5): 스킬 fallback 문구가 옛 이름 줄도 rust 로 받는다고 적어요.
  verify: `grep -n "both lines count as Rust" skills/moai/SKILL.md` 가 있어요.
- AC-6 (REQ-6): 발생 표(위)가 종류별 줄 수·의미·조각을 전부 나열해요.
  verify: 이 문서의 Occurrence table — spec check 통과.

## Scope

- write: `.moai/specs/SPEC-20261121-moai-name/spec.md`, `src/bin_names.rs`, `src/main.rs`(C-313 반환 뒤), `skills/moai/SKILL.md`
- read-only: `src/install.rs`(이미 두 이름을 받음), 조사 자료(리더 scratchpad mcli/)
- forbidden: 타 레인 hold 파일(path-root resolver·src/lib.rs 상수, Go-link name-switch·install/commit-msg, Python, src/card), 조각 4·10 의 파일, 카드 번호를 SPEC id·가지·커밋·코드에 넣기

## Plan

1. 읽는 쪽을 두 이름으로: bin_names 의 같은 빌드 검사에 함수를 두고 단위 시험으로 두 이름을 고정해요.
2. 스킬 fallback 문구를 새 시그니처로 재작성하고 옛 이름 줄도 받는다고 적어요.
3. SPEC 을 쓰고(이 문서) clap name 변경은 src/main.rs 반환 뒤에 넣어요.
4. 확인: targeted 시험, spec check, 커밋.

## Risks

- clap name 을 바꾸는 순간 같은 빌드 검사(bin_names)와 설치 읽는 쪽이 옛 이름 줄을 만날 수 있어요 — 읽는 쪽을 먼저 두 이름으로 만들어 순서를 지켜요(이 조각의 순서).
- 스킬 텍스트는 설치가 옛 이름인 기간에도 참이어야 해요 — fallback 경로 안내(`moai-cli` 옛 설치/`moai` 새 설치)를 함께 적어요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | src/lib.rs 의 VERSION_SIGNATURE 는 그대로 — 이름을 담지 않는다 | 시그니처 문자열에 이름이 없고 이름은 clap 이 붙여요; lib.rs 는 path-root 레인이 hold 해요 | 이름을 상수에 넣기 |
| D-2 | src/install.rs 의 generation_of/is_rust_signature 는 이미 두 이름을 받아 코드 변경 없음 — 시험으로 고정만 | 감사 판독(#2791 AC-2)이 `^(moai-cli\|moai) …(rust)$` 를 이미 받아요 | 한쪽으로 좁히기 |
| D-3 | go_settings_clean 의 `$HOME/.moai-cli/path/moai-cli` 런처 경로는 이 SPEC 의 대상에서 뺀다 | Go 세대 상태 폴더 경로라 path-root 카드와 같은 이동이에요 | 이 SPEC 이 경로를 바꾸기 |
| D-4 | vault 유도선은 불변 | 기존 볼트 항목을 읽을 수 없게 되는 것을 막아요(guarded, 볼트 내용 미개봉) | 유도선 버전 올리기 |

## Result

- (조각 1 진행 중 — clap name 은 src/main.rs 반환 뒤, 나머지는 아래)
- AC-2: met — `cargo test --locked --offline --lib bin_names::tests::the_same_build_check_accepts_both_binary_names` 통과(`moai <sig>`·`moai-cli <sig>` 모두 같은 빌드, Go·다른 빌드·출력 없음은 아님).
- AC-4: met — SKILL.md fallback 문구 재작성(`moai <version> (rust)` 검사 + 설치 시기별 경로 안내).
- AC-5: met — 같은 문구에 옛 이름 줄(`moai-cli <version> (rust)`)도 rust 로 받는다고 적혀 있어요.
- AC-3: met(기존 동작) — install.rs 의 generation_of/is_rust_signature 는 `^(moai-cli|moai) [0-9]\S* \(rust\)$` 를 이미 받아요(코드 변경 없음 — 시험 고정은 install_cli 기존 generation 시험).
- AC-6: met — 위 Occurrence table(19 종류 전부, 조각 매핑 포함) — spec check 통과.
- AC-1: 보류 — src/main.rs 의 clap name 은 C-313 보유(src/main.rs)라 pair 요청 노트를 남겼어요(카드 노트 pair-request-main-rs-20261121T193632Z); 반환 뒤 이어서.
