---
id: SPEC-20261121-moai-name
title: "The temporary name moai-cli leaves the product: every remaining occurrence, by kind and slice"
status: draft
revision: 2
created: 2026-11-21
---

# The temporary name moai-cli leaves the product

## Intent

사용자 지시(2026-11-21, 리더 대화 원문): "moai-cli 전수 조사해서 또 수정에 빠져 있는거 찾아서 완전 제거해라". 사용자 결정(2026-11-22, 카드 노트 6286): 릴리스 태그는 `v4.<yearweek>.<build>`, 내려받는 파일은 `moai-adk-<version>-<platform>` 이고 `moai-cli` 이름은 남기지 않아요. 임시 이름 `moai-cli`(바이너리·패키지·크레이트·폴더·변수·릴리스 이름)는 제품에서 완전히 사라지고 이름은 `moai` 하나예요.

이 SPEC 은 그 전체의 목록과 순서예요. 조각마다 이 SPEC 의 개정과 착지가 따로 있고, 각 조각의 요구는 아래 Requirements 에 그 조각이 시작될 때 더해요. 지금 개정(2)은 조각 1 이에요: 남은 발생 전부를 종류별·조각별로 나열하고, 출력되는 버전 줄이 `moai` 를 이름하기 전에 읽는 쪽이 두 이름을 모두 받게 해요.

### 왜 다시 쓰나 (2026-11-23 확인)

개정 1 의 조각 1 커밋 `6f0896b2f8` 은 NeXT 의 조상이지만, 병합 `b6e7f8cfcb`("통합 브랜치가 이미 하는 것은 가져온다")가 이 SPEC 파일과 `src/bin_names.rs` 의 변경을 지웠어요(`git log --full-history -m --diff-filter=D` 로 확인). 그래서 NeXT 팁 `2ade55ba2d` 에는 SPEC 도, 두 이름을 받는 같은 빌드 검사도 없어요. 이 개정이 그 일을 팁 위에서 다시 해요.

### 지금 코드에서 읽는 쪽 (측정: NeXT `2ade55ba2d`)

| 읽는 쪽 | 위치 | 두 이름을 받나 |
| --- | --- | --- |
| 설치 세대 판별 `generation_of` | `src/install.rs` `is_rust_signature` | 받아요 (`moai-cli` 와 `moai`, 단위 시험이 고정) |
| 같은 빌드 검사 (`bin-names-plan` 의 `bin/moai`) | `src/bin_names.rs` | 아니요 — 정확히 `moai-cli <sig>` 만 |
| 릴리스 워크플로우 버전 검사 | `.github/workflows/release.yml` 단계 `verify` | 아니요 — 정확히 `moai-cli ${VERSION} (rust)` 만 |
| 스킬 fallback 문장 | `skills/moai/SKILL.md` | 아니요 — `moai-cli <version> (rust)` 만 |
| Python 도구·시험(`tools/smoke`, `tools/statusline_golden.py`, `tests/*.py`) | Python | 이 카드 밖(Rust-only 카드) |

settings cleaner 는 버전 줄을 읽지 않아요(`~/.moai-cli/path/moai-cli` 런처 경로와 훅 명령을 지울 뿐) — 이름을 바꾸는 쪽은 조각 5 예요.

## Non-goals

- 다른 카드가 이미 쥔 것은 여기서 다시 하지 않아요: 상태 폴더 기본값과 HOME 마이그레이션(path-root 카드), 프로젝트 측 폴더 이동, Go 링크와 name-switch 스크립트 제거, Python 파일(Rust-only 카드), Codex/OMP 제거.
- 조각 4(캐시·임시 폴더 이름)와 조각 10(crate 이름)은 레인이 멈춘 뒤 리더가 예약해요.
- 키 볼트 유도선은 불변이에요(조각 9, guarded: 볼트 내용을 열지 않아요).
- 설치된 `~/.moai` 와 `~/.moai-cli` 의 링크와 폴더는 건드리지 않아요. 그 정리는 착지 뒤 사용자 선택이에요.
- NeXT 에 있는 빨강 1개(`mode_workflow_steps`)는 이 SPEC 의 범위 밖이에요(수리 중, 2026-11-23 리더 배차).

## Requirements

- REQ-1: When the same-build check of `bin-names-plan` reads the first `--version` line of a separate `bin/moai`, it shall count `moai <build version> (rust)` and `moai-cli <build version> (rust)` as this build and any other line as another program · failure: none — a line it does not know leaves `bin/moai` untouched and reports `blocked`.
- REQ-2: When the release workflow verifies the built binaries, it shall accept the first `--version` line `moai ${VERSION} (rust)` or `moai-cli ${VERSION} (rust)` for each binary · failure: none — REQ-7 handles a line that is not one of the two.
- REQ-3: The skill's fallback sentence shall name the Rust signature `moai <version> (rust)` first and accept `moai-cli <version> (rust)` as the same signature from an earlier build · failure: none.
- REQ-4: The install generation reader shall keep accepting both names · failure: none — pinned by the existing unit test `generation_of`.
- REQ-5: When `moai --version` runs, the first line shall begin with `moai ` and end with ` (rust)` for both binary names · failure: none — a build that still prints the earlier name stays readable through REQ-1 to REQ-4 until this requirement lands.
- REQ-6: The occurrence table of this SPEC shall list every kind of remaining occurrence with its slice · failure: none.
- REQ-7 (REQ-2): IF a built binary prints a first `--version` line other than those two, THEN the release workflow shall fail the job with a message naming both accepted forms before it writes the release note.

## Order that keeps installed copies working

설치된 훅·스킬·설정은 옛 이름을 부르는 채로 있어요. 순서가 그것을 지켜요.

1. **읽는 쪽을 먼저 둘 다 받게 한다(조각 1a, 이 개정).** 아직 아무 바이너리도 이름을 바꿔 찍지 않아요. 같은 빌드 검사·릴리스 검사·스킬 문장이 두 이름을 받아서, 이후 단계에서 어느 쪽이 찍어도 안 깨져요.
2. **찍는 이름을 `moai` 로 바꾼다(조각 1b).** `src/main.rs` 의 clap 이름 한 줄. 옛 설치(`moai-cli <ver> (rust)` 를 찍는 빌드)는 읽는 쪽이 계속 받으므로 읽기가 안 깨져요. `src/main.rs` 는 C-397 이 쥐고 있어 pair 요청 노트로 맡겨 두었고(카드 C-397 의 노트 `pair-request-c318-main-rs-20261123T0825Z`), 답이 오면 이어요.
3. **도움말·거절 문구·골든(조각 2).** 사용자에게 보이는 글만이에요.
4. **바이너리와 패키지(조각 5).** 새 이름 `moai` 가 실제 파일이 되고 `moai-cli` 는 호환 링크로 남아, 옛 훅·설정이 부르는 `moai-cli` 가 계속 돼요. 링크는 업데이트 단계가 지워요(카드 본문 5). 임시 HOME 에서 도는 Rust 시험이 이 호환 읽기를 증명해요.
5. **릴리스 이름(조각 6).** 태그 `v4.<yearweek>.<build>`, 파일 `moai-adk-<version>-<platform>`(결정 6286). 옛 설치 스크립트가 보게 되는 것은 그때 이 SPEC 이 적어요.

## Occurrence table

측정: 2026-11-23 NeXT `2ade55ba2d`, `git ls-files -z | xargs -0 rg -c -e 'moai-cli|moai_cli'` — **1061 파일, 7208 줄**. 종류별 줄 수는 리더의 조사(2026-11-21, tip `6e1bedb4f3`, 999 파일·6969 줄, `classified2.tsv`)에서 가져왔고 이 개정에서 다시 세지 않았어요(그 사이 늘었을 수 있어요; 조각 11 의 검사 시험이 새 발생을 막아요).

| Kind | Lines (2026-11-21) | What it is | Slice |
| --- | --- | --- | --- |
| i-a state folder in tests | 1499 | tests 의 `.moai-cli` 상태 폴더 경로 | 8 (path-root 뒤, resolver 로) |
| f tracked `.moai-cli` tree + ignores | 1284 | tracked 상태 트리 157 파일·ignore·scan 설정 | 7 (path-root 착지 뒤) |
| h other docs/history | 974 | 문서·design-review·specs 이력 | 11 (이력은 이력으로 유지) |
| c1 crate path in src | 695 | `moai_cli::` crate 경로(src) | 10 (리더 예약, 레인 멈춤) |
| i-c crate path in tests | 683 | `moai_cli::`(tests) | 10 |
| k comments in src | 392 | src 주석 | 11 |
| j Python | 252 | Python 파일 | Rust-only 카드 (제외) |
| i-o binary word in tests | 348 | tests 의 바이너리 이름 그 밖 | 8 |
| i2 unit tests in src | 206 | src 내부 단위 시험 | 10 |
| g skills/templates/agents | 392 | 스킬·템플릿 텍스트 | 1a (fallback 문장), 2, 11 |
| i-d env/file names in tests | 187 | tests 의 env·파일 이름 | 3 |
| i3 fixtures+goldens | 155 | fixture·golden | 2 |
| e installer+release | 110 | install/·.github/·scripts/·dist-workspace | 1a (릴리스 검사 한 곳), 5, 6 |
| e2 dev scripts+config | 45 | .cargo/·.config/ 개발 설정 | 3, 5 |
| c2 binary-name literal in src | 43 | src 의 `moai-cli` 리터럴 | 1a (같은 빌드 검사), 1b (clap 이름), 2 (사용자 문자열) |
| d env/file names in src | 34 | `MOAI_CLI_` 계열(src) | 3 |
| b user-visible strings | 20 | help·거절 메시지 | 2 (상태 경로분은 path-root 뒤) |
| c0 manifests | 10 | Cargo.toml·lock·.cargo | 5 |
| a state folder in src | 6 | src 의 `.moai-cli` 상태 폴더 | path-root 카드 (제외) |

## Acceptance

- AC-1 (REQ-1): 별도 `bin/moai` 가 `moai <sig>` 또는 `moai-cli <sig>` 를 찍으면 `replace_copy`, 다른 빌드 번호·`moai-adk 3.2.0`·모르는 이름은 `blocked` 예요 · verify: `cargo nextest run --test bin_names_signature_cli -E 'test(/^a_separate_moai/)'`
- AC-2 (REQ-2): 릴리스 워크플로우의 `verify` 단계 스크립트를 가짜 바이너리로 실제로 돌려, 두 이름의 네 조합은 통과하고 다른 버전·Go 줄·모르는 이름·`(rust)` 없는 줄은 실패해요 · verify: `cargo nextest run --test bin_names_signature_cli -E 'test(/^the_release_check/)'`
- AC-3 (REQ-3): 스킬의 버전 검사 문장이 `moai <version> (rust)` 를 먼저, `moai-cli <version> (rust)` 를 같은 시그니처로 적고 fallback 은 `moai-cli` 예요 · verify: `cargo nextest run --test bin_names_signature_cli -E 'test(/^the_skill_fallback/)'`
- AC-4 (REQ-4): 설치 세대 판별이 두 이름을 rust 로 받아요 · verify: `cargo nextest run --lib -E 'test(/generation_of|rust_signature/)'`
- AC-5 (REQ-5): `--version` 의 첫 줄이 `moai <version> (rust)` 이고 usage·help 이름이 `moai` 예요 — 조각 1b · verify: 1b 개정에서 적어요.
- AC-6 (REQ-6): 이 문서의 Occurrence table 이 종류별 줄 수·의미·조각을 전부 나열해요 · verify: `moai spec check .moai/specs/SPEC-20261121-moai-name/spec.md` 의 errors 가 0 이에요.

## Scope

- write (조각 1a): `.moai/specs/SPEC-20261121-moai-name/spec.md`, `src/bin_names.rs`, `skills/moai/SKILL.md`, `skills/moai/references/source-manifest.json`(SKILL.md 의 sha256 한 줄), `.github/workflows/release.yml`(단계 `verify` 만), `tests/bin_names_signature_cli.rs`
- read-only: `src/install.rs`(이미 두 이름을 받음), 조사 자료(리더 scratchpad `mcli/`)
- forbidden: 다른 카드가 쥔 파일(`src/main.rs`·`src/lib.rs` 는 C-397, 상태 폴더 resolver·상수는 path-root 카드, Python 파일은 Rust-only 카드), 조각 4·10 의 파일, 설치된 홈 상태 폴더(`.moai`, `.moai-cli`), 카드 번호를 SPEC id·가지·커밋·코드에 넣기

## Plan

1. RED: `tests/bin_names_signature_cli.rs`(같은 빌드 검사, 릴리스 스크립트, 스킬 문장)이 빨강인 것을 먼저 봐요.
2. `src/bin_names.rs`: 같은 빌드 검사를 함수로 빼고 두 이름을 받게 해요(+ 단위 시험).
3. `.github/workflows/release.yml` 단계 `verify`: 두 이름을 받고, 메모에는 바이너리가 찍은 첫 줄을 그대로 적어요.
4. `skills/moai/SKILL.md` 문장 재작성과 `source-manifest.json` 의 해시 갱신.
5. 확인: 위 AC 의 시험, `moai check affected`, `moai spec check`. 읽는 쪽이 끝난 뒤 조각 1b(clap 이름)를 C-397 의 답에 따라 이어요.

## Risks

- clap 이름을 바꾸는 순간 옛 이름만 받는 읽는 쪽이 깨져요 — 그래서 1a 가 1b 보다 먼저예요(위 Order).
- 릴리스 워크플로우의 태그 이름 검사(`moai-adk-v${VERSION}`)는 이 조각에서 바꾸지 않아요 — 결정 6286 의 새 태그는 조각 6 이 바꿔요.
- 스킬 문장은 옛 이름의 설치가 남는 동안에도 참이어야 해요 — fallback 은 계속 `moai-cli` 예요(조각 5 의 호환 링크가 사라질 때 같이 지워요).

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | `VERSION_SIGNATURE` 상수는 이름을 담지 않는다(`<ver> (rust)`) | 이름은 clap 이 앞에 붙여요 — 한 줄 바꿈으로 충분해요 | 이름을 상수에 넣기 |
| D-2 | `install.rs` 의 `generation_of` 는 코드를 바꾸지 않는다 | 이미 `^(moai-cli\|moai) [0-9]\S* \(rust\)$` 를 받아요 | 한쪽으로 좁히기 |
| D-3 | settings cleaner 의 런처 경로 `$HOME/.moai-cli/path/moai-cli` 는 조각 5 | 이름 바꿈이 아니라 호환 읽기의 문제예요 | 조각 1 로 당기기 |
| D-4 | 키 볼트 유도선은 불변 | 기존 볼트 항목을 읽을 수 없게 되는 것을 막아요(guarded, 볼트 내용 미개봉) | 유도선 버전 올리기 |
| D-5 | 릴리스 워크플로우의 메모에는 바이너리가 찍은 첫 줄을 그대로 적는다 | 옛·새 이름이 섞여 있는 동안 메모가 거짓이 되지 않아요 | 고정 문구로 되돌리기 |

## Result

- (조각 1a 진행 중)
