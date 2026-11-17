---
id: SPEC-20261117-versioning-format-chars
title: "custom 버전 문구가 Unicode 형식 문자(Cf)를 모두 거부"
status: draft
revision: 1
created: 2026-11-17
parent: SPEC-20261117-versioning-refusals
---

## Intent

C-71 감사 N3: custom 버전 문구는 Default_Ignorable_Code_Point(4,174개)를 거부하지만, Unicode가 그 집합에서 일부러 뺀 형식
문자는 받아요. 그중 U+FFF9–U+FFFB(행간 주석)와 U+13430–U+1343F(이집트 상형문자 형식 제어)는 보이지 않거나 둘레 글자의
배치를 바꿔요. 공식 `DerivedGeneralCategory.txt` 18.0.0(https://www.unicode.org/Public/18.0.0/ucd/extracted/DerivedGeneralCategory.txt,
2026-11-17 16:34:16Z 읽음, 파일 sha256 `d6b151d2d40ee9b1876d26f417980f45ffae47b6055ccf7203cb31f07a030f94`)의
General_Category=Format 절(21줄, 170개)을 기준으로, Cf 전체를 거부해요.

## Requirements

- REQ-1: IF the custom versioning text holds a code point of General_Category Cf in Unicode 18.0.0, THEN the profile validation shall refuse it.
  · failure: none
- REQ-2: The profile validation shall keep accepting Arabic letters and digits, Devanagari text with its combining marks, Korean text and emoji without joiners.
- REQ-3: The versioning reference and the `project-plan` operation page shall name every refused Cf range.

## Acceptance

- AC-1 (REQ-1): 코드의 Cf 표가 출처 발췌와 같고(170개), 각 범위의 양 끝(N)은 거부, 바깥 이웃(N+1)은 다른 이유로 거부되지 않으면 받아요.
  verify: `cargo test --lib versioning::tests::format_table_matches_the_unicode_source`
  verify: `cargo test --lib versioning::tests::custom_text_refuses_every_format_character`
- AC-2 (REQ-2): 아랍어·데바나가리·한국어 문구와 결합자 없는 이모지는 받아요.
  verify: `cargo test --lib versioning::tests::custom_text_refuses_every_format_character`
- AC-3 (REQ-3): 두 문서가 Cf 범위를 모두 적었는지 기존 표-문서 시험이 확인해요.
  verify: `cargo test --lib versioning::tests::format_table_matches_the_unicode_source`

## Scope

- write: `src/versioning.rs`, `tests/fixtures/unicode/general-category-cf-18.0.0.txt`,
  `skills/moai/references/internal/versioning.md`, `docs/design-review/appendix/operations/project_plan.md`,
  `skills/moai/references/source-manifest.json`(manifest 도구가 다시 씀),
  `.moai/specs/SPEC-20261117-versioning-format-chars/spec.md`
- read-only: `tests/fixtures/unicode/LICENSE-Unicode-V3.txt`, `tests/fixtures/unicode/default-ignorable-18.0.0.txt`
- forbidden: 실제 HOME, `.github/workflows/audit-gate.yml`

## Plan

1. 출처 읽기(시각·sha256), Cf 절만 발췌(C-71과 같은 라이선스 파일).
2. 코드 `FORMAT` 표와 `hidden()`, 시험(표-출처, N/N+1, 받아야 할 문구).
3. 문서 두 곳, manifest, 검사, mutation 3개.

## Decisions

- D1 Cf 전체를 거부해요(리더 권고). 근거: Cf 170개 중 DI에 없는 것은 U+0600–U+0605, U+06DD, U+070F, U+0890–U+0891,
  U+08E2, U+110BD, U+110CD(Prepended_Concatenation_Mark: 보이지만 뒤 숫자들에 걸쳐 그려져 숫자 표시를 바꿈), U+FFF9–U+FFFB,
  U+13430–U+1343F(보이지 않거나 배치를 바꿈)이에요. 평범한 아랍어(글자 Lo, 숫자 Nd)·데바나가리(Mn·Mc)·한국어는 Cf가 아니라
  그대로 받아요. 이모지 ZWJ 줄과 FE0F 이모지, 페르시아어 ZWNJ(U+200C)는 C-71부터 이미 거부돼요(DI); 이번에 바뀌지 않아요.
- D2 지시는 U+13430–U+13440이라 했지만 18.0.0 출처에서 Cf는 U+13430–U+1343F이고 U+13440은 Mn(EGYPTIAN HIEROGLYPH MIRROR
  HORIZONTALLY)이에요. 출처를 따르고, U+13440은 N+1 이웃으로 받아요.
- D3 출처 전체를 싣지 않고 Cf 절(제목 줄부터 total 줄까지)만 발췌해요. 표를 코드에만 두면 시험이 자기 자신과 비교하게 되니까요.

## Result

구현 에이전트가 채웠어요. 기준 `17c19a2759`(origin/v2026.11.0), 변경 커밋 `a091e3e97a`.

- RED: 새 시험을 쓰고 `hidden()`에서 `FORMAT`을 뺀 옛 동작으로 돌리니 `custom_text_refuses_every_format_character`가
  U+0600에서 FAIL(EXPECTED_RED). 되돌린 뒤 22 통과.
- AC-1: met — 표가 발췌와 같고 170개; 21개 범위의 양 끝 거부, 바깥 이웃 32개 받음(다른 이유로 거부되는 10개 제외).
  C-71 이웃 시험은 U+FFF9가 이제 형식 문자라 기대값을 `2 * 17 - 2`로 고쳤어요.
- AC-2: met — 아랍어·아랍 숫자, 데바나가리, 한국어, 🚀 문구가 받아져요.
- AC-3: met — 두 문서가 21개 범위를 모두 적었어요(표-문서 시험).
- mutation(커밋 뒤, 각각 `git restore` 뒤 tree 깨끗): `hidden()`이 `FORMAT`을 안 봄 → 1 FAIL; U+1343F를 U+13440으로 → 2 FAIL;
  U+FFF9–U+FFFB 줄 삭제 → 2 FAIL.
- 함께 돌림: `--lib profile::` 10, `versioning_cli` 15, `agent_defs_package` 18, `product_cli` 75, `--bin moai` 27, fmt,
  clippy `-D warnings`, Python 4개 88 OK, `source_manifest.py write`(versioning.md 다시 씀)·`check`, `moai spec check --all`.

실행 안 함: 전체 cargo test, Linux CI.
