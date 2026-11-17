---
id: SPEC-20261117-versioning-refusals
title: "버저닝 custom 글의 거절 목록 — 경로 문자와 Unicode 기본 무시 문자 전체"
status: draft
revision: 1
created: 2026-11-17
---

## Intent

버저닝 카드(C-69)의 독립 감사가 남긴 두 가지 후속(R2-N1, R2-N2)을 고쳐요. 프로필의 `versioning.format`이 `custom`일
때 사용자의 글 한 줄을 받는데,

1. 문서(`internal/versioning`, `internal/interview`, `operations/project_plan.md`)의 거절 목록에 `/`와 `\`가 빠졌어요.
   프로필의 portable 검사(`src/profile.rs` `portable`)는 둘을 거절해요.
2. 보이지 않는 문자 검사(`src/versioning.rs` `hidden`)가 Unicode의 기본 무시 문자(Default_Ignorable_Code_Point) 일부만
   막았어요. U+034F, U+115F–U+1160, U+17B4–U+17B5, U+180B–U+180F(U+180E 제외), U+3164, U+FFA0, U+FFF0–U+FFF8,
   U+1BCA0–U+1BCA3, U+1D173–U+1D17A, U+E0080–U+E00FF, U+E01F0–U+E0FFF가 통과했어요.

## Requirements

요구 문장은 영어 GEARS예요(`src/spec.rs` `parse_gears`). `When` 요구가 없어 실패 경로 요구도 없어요.

- REQ-1: The custom versioning text check shall refuse every code point of the Unicode 18.0.0 Default_Ignorable_Code_Point set.
- REQ-2: The custom versioning text check shall keep refusing the line and paragraph separators U+2028 and U+2029.
- REQ-3: The embedded default-ignorable table shall equal the ranges of a verbatim excerpt of the Unicode source file kept in the repository.
- REQ-4: Every refusal list of custom versioning text in the skill references and the operations appendix shall name `/` and `\`.
- REQ-5: The refusal list of `internal/versioning` shall name every default-ignorable range the code refuses.
- REQ-6: The change shall add no crate dependency.
- REQ-7: The repository shall carry the full Unicode License V3 notice beside the copied Unicode excerpt, outside the skill package.
- REQ-8: The profile check shall refuse `\` in custom versioning text as it refuses `/`.

## Acceptance

- AC-1 (REQ-1, REQ-2): 범위마다 처음과 끝(N)은 거절되고, 바로 바깥 이웃(N+1)은 다른 이유로 거절되는 것이 아니면 통과해요.
  감사가 이름을 든 코드 포인트는 모두 거절돼요.
  verify: `cargo test --locked --offline --lib versioning::tests::custom_text_refuses_markup_and_every_hidden_range`
- AC-2 (REQ-3, REQ-5, REQ-7): 표가 원본 발췌(`tests/fixtures/unicode/default-ignorable-18.0.0.txt`)의 범위와 합계 4,174와
  같고, `internal/versioning`과 `operations/project_plan.md`가 모든 범위와 `/`·`\`를 적고, 발췌 머리가 같은 폴더의
  `LICENSE-Unicode-V3.txt`를 가리키고 그 파일이 Unicode License V3 전문이에요.
  verify: `cargo test --locked --offline --lib versioning::tests::default_ignorable_table_matches_the_unicode_source`
- AC-3 (REQ-4): 세 문서가 `/`와 `\`를 적어요.
  verify: `grep -c '`/`' skills/moai/references/internal/versioning.md skills/moai/references/internal/interview.md docs/design-review/appendix/operations/project_plan.md` → 각 1 이상
- AC-4 (REQ-6): `Cargo.toml`과 `Cargo.lock`이 바뀌지 않았어요.
  verify: `git diff --exit-code origin/v2026.11.0 -- Cargo.toml Cargo.lock`
- AC-6 (REQ-7, REQ-8): 라이선스 파일이 받은 원본과 같고 skill 패키지 밖에 있으며, 프로필이 `\`를 거절해요.
  verify: `shasum -a 256 tests/fixtures/unicode/LICENSE-Unicode-V3.txt` → `e7a93b009565cfce55919a381437ac4db883e9da2126fa28b91d12732bc53d96`;
  `git ls-files skills | grep -c unicode` → `0`; `cargo test --locked --offline --lib profile::`
- AC-5 (REQ-1…REQ-8): 바뀐 곳에 닿는 검사가 통과해요.
  verify: `cargo test --locked --offline --test versioning_cli`; `cargo test --locked --offline --test agent_defs_package`;
  `cargo test --locked --offline --test product_cli`; `cargo test --locked --offline --bin moai`; `cargo fmt --check`;
  `cargo clippy --locked --offline --all-targets -- -D warnings`; `python3 -I tools/source_manifest.py check`;
  `python3 -m unittest tests/repo_scans_test.py tests/ci_scope_test.py tests/skill_references_test.py tests/source_manifest_test.py`

## Scope

- write: `src/versioning.rs`, `src/profile.rs`(시험만), `tests/fixtures/unicode/default-ignorable-18.0.0.txt`,
  `tests/fixtures/unicode/LICENSE-Unicode-V3.txt`,
  `skills/moai/references/internal/versioning.md`, `skills/moai/references/internal/interview.md`,
  `skills/moai/references/source-manifest.json`, `docs/design-review/appendix/operations/project_plan.md`,
  `.moai/specs/SPEC-20261117-versioning-refusals/spec.md`
- read-only: `src/lib.rs`
- forbidden: `src/headver.rs`, `src/update.rs`, `Cargo.toml`, `Cargo.lock`, 다른 열린 가지의 파일

## Plan

1. 원본 받기: `https://www.unicode.org/Public/18.0.0/ucd/DerivedCoreProperties.txt`(파일 sha256 `09c928886a178fca…`,
   2026-11-17 읽음)에서 Default_Ignorable_Code_Point 절을 그대로 발췌해 시험 자료로 둬요.
2. RED: 표를 읽는 시험과 범위 끝 N/N+1 시험을 먼저 쓰고, 옛 `hidden`에서 U+034F로 실패하는지 확인해요.
3. GREEN: `hidden`이 17개 범위 표와 U+2028–U+2029를 써요.
4. 문서 세 곳의 목록과 manifest, 검사, 커밋, mutation.

## Decisions

- D1 Unicode 판은 18.0.0이에요: `latest` 주소가 2026-11-17에 준 판이고, 시험 자료와 코드 주석에 판과 주소를 적었어요.
- D2 crate를 더하지 않고 범위를 코드에 넣었어요. 이웃한 원본 줄은 한 범위로 합쳐 17개예요(합계 4,174는 같음).
- D3 U+2028–U+2029는 기본 무시 문자가 아니지만 줄을 바꾸므로 계속 거절해요.
- D4 범위 바깥 이웃의 통과 확인은 다른 범위나 U+2028–U+2029, 제어 문자인 이웃을 건너뛰어요. 17개 범위의 이웃 34개 중
  U+2029 하나만 건너뛰어 33개를 확인해요.
- D5 `/`와 `\`는 코드에 새로 넣지 않았어요: 프로필의 portable 검사가 이미 거절하고, 이번 수리는 문서가 그 사실을 적게 하는 것이에요.
- D6 (감사 FAIL 수리, 리더 결정) 발췌가 Unicode Data File의 복사본이라 Unicode License V3 19–21행("this copyright and
  permission notice appear with all copies of the Data Files")에 따라 전문을 같은 폴더에 둬요:
  `https://www.unicode.org/license.txt`를 2026-11-17 10:02:53Z에 받았고(sha256
  `e7a93b009565cfce55919a381437ac4db883e9da2126fa28b91d12732bc53d96`, 39줄, 1,995바이트) 바이트 그대로
  `tests/fixtures/unicode/LICENSE-Unicode-V3.txt`에 뒀어요. 발췌 머리 주석 3줄이 그 파일을 가리켜요. 데이터 절은 원본과 바이트가
  같아요. 두 파일은 `tests/` 아래라 skill 패키지(`skills/moai`)와 설치물에 들어가지 않아요(`Cargo.toml`은 `publish = false`).

## Result

구현 에이전트가 2026-11-17에 채웠어요. SYNC와 AUDIT가 다시 확인해요. 기준 `25881165be`(origin/v2026.11.0), 변경 커밋
`949d2ddbc9`. 빌드 폴더는 이 worktree의 `target/`, `CARGO_INCREMENTAL=0`, `CARGO_BUILD_JOBS=4`, macOS arm64,
cargo는 `moai heavy-slot-acquire` 슬롯 안에서 한 번에 하나.

- AC-1: met — RED(표만 넣고 `hidden` 그대로): exit 101, `custom_text_refuses_markup_and_every_hidden_range` "U+034F".
  (그 전 실행은 시험 해석기가 발췌의 머리 주석 줄을 읽어 실패한 TOOL_FAILURE였고, `#` 줄을 건너뛰게 고쳤어요.)
  GREEN: `--lib versioning::` exit 0(20 passed).
- AC-2: met — 같은 실행에서 `default_ignorable_table_matches_the_unicode_source` 통과(범위 17개, 합계 4,174).
- AC-3: met — `versioning.md` 2, `interview.md` 1, `project_plan.md` 1.
- AC-4: met — `git diff --exit-code origin/v2026.11.0 -- Cargo.toml Cargo.lock` exit 0.
- AC-5: met — `versioning_cli` 0(15); `agent_defs_package` 0(18); `product_cli` 0(75); `--bin moai` 0(27);
  `--lib profile::` 0(10); `cargo fmt --check` 0(처음 1, 줄바꿈 두 곳 뒤 0); `rustfmt --check --edition 2021 src/ops/*.rs` 0;
  clippy `-D warnings` 0; `tools/source_manifest.py check` 0; Python 4개 파일 OK; `moai spec check --all` 0(처음 1:
  REQ-4 주어에 쉼표가 있어 GEARS 해석 실패, 문장을 고친 뒤 0); `git diff --check` 0.
- mutation 3개(커밋 뒤, 각각 `git restore` 뒤 tree 깨끗): U+034F 범위를 다른 범위 복사로 바꿈 → exit 101, 두 시험 FAIL;
  마지막 범위 끝 U+E0FFF → U+E0FFE → exit 101 "U+E0FFF"와 표 불일치; `hidden`에서 U+2028–U+2029 줄 삭제 → exit 101
  "U+2028". 되돌린 뒤 exit 0.

실행 안 함: 전체 `cargo test`, Linux·Windows.
