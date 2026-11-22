---
id: SPEC-20261121-rust-only-text-scans
title: "The protections that only Python modules hold on documents, metadata and skill references live in Rust tests"
status: draft
revision: 1
created: 2026-11-21
---

# 문서·메타데이터 점검을 Rust 시험으로

## Intent

조각 1(`SPEC-20261121-rust-only-inventory`)의 판정표에서 "파이썬에만 있는 보호" 중 문서·메타데이터·스킬 참조를 읽는 시험을 Rust 로 옮기는 조각이에요. 사용자 결정(2026-11-21)은 "테스트는 rust로만"이고, 이 SPEC 의 첫 착지는 같은 성격의 파이썬 모듈 13개(사례 79개)를 다뤄요. 나머지 시험(GitHub 양식, 워크플로·설치 스크립트, 바이너리를 실행하는 시험)은 같은 방식으로 뒤따르는 착지에서 이 SPEC 의 Result 에 적어요.

이 head 는 SPEC 만 담고 코드는 없어요. 착지 순서상 경로 뿌리와 Go 연결 정리가 먼저 들어가야 해서 구현은 그 뒤에 시작해요.

판정표에서 옮길 대상(사례 수는 `ast` 로 센 값):

| 파이썬 모듈 | 사례 | 보호하는 것 | Rust 대상 |
|---|---|---|---|
| `tests/host_matrix_test.py` | 5 | 호스트 지원 행렬 문서의 행(ZCode·Codex·Claude)과 라벨 규칙 | `tests/docs_scans.rs` |
| `tests/omp_host_survey_test.py` | 9 | 호스트별 기능 조사표가 코드와 맞음 | `tests/docs_scans.rs` |
| `tests/oss_docs_test.py` | 18 | 공개 문서(README·라이선스·행동 강령·보안)의 존재·링크·금지 문구와 그 점검기 자신의 단위 사례 | `tests/docs_scans.rs` |
| `tests/release_rehearsal_test.py` | 5 | 릴리스 리허설 문서의 단계별 행과 증거 | `tests/docs_scans.rs` |
| `tests/space_applied_record_test.py` | 7 | `docs/design-review/applied.md` 표와 fixes 폴더 대응 | `tests/docs_scans.rs` |
| `tests/skill_web_pages_test.py` | 4 | `## Web pages` 절이 워크플로마다 한 번 | `tests/docs_scans.rs` |
| `tests/agents_local_build_folder_test.py` | 5 | check set 이 `CARGO_TARGET_DIR` 을 정하지 않고 규칙 문구가 맞음 | `tests/docs_scans.rs` |
| `tests/cargo_metadata_test.py` | 5 | `Cargo.toml` 메타데이터가 README 사실과 맞음 | `tests/docs_scans.rs` |
| `tests/dev_profile_sha2_test.py` | 2 | dev 프로파일에서 sha2 최적화 | `tests/docs_scans.rs` |
| `tests/platform_targets_test.py` | 5 | dist 대상이 지원 플랫폼 셋이고 설치 스크립트가 맞게 고름 | `tests/docs_scans.rs` |
| `tests/support_matrix_test.py` + `tools/check_support_matrix.py` | 5 | supported 행은 통과한 영수증을 인용, 문서에 비밀 모양 없음 | `tests/docs_scans.rs` |
| `tests/skill_references_test.py` | 5 | 추적 파일의 `skills/...` 참조가 실제 파일을 가리킴 | `tests/skill_references_scan.rs` |
| `tests/dist_include_test.py` | 4 | 릴리스 아카이브가 스킬 패키지와 기본 에이전트를 담음 | `tests/release_suite.rs` |

## Non-goals

- 파이썬 파일 삭제(조각 4 가 해요). 이 조각은 파이썬 시험을 그대로 두고, 같은 사실을 Rust 가 단정하게 해요.
- `tools/local_check.sh` 의 `unittest` 단계와 파이썬 슬롯 제거. 파이썬 시험이 하나라도 남아 있는 동안은 그 단계가 필요해서, 마지막 파이썬 시험이 없어지는 조각이 걷어요(결정 D8).
- GitHub 양식·워크플로·설치 스크립트·`local_check.sh`·`self_check.sh` 시험, 바이너리를 실행하는 시험, 도구 시험. 뒤따르는 착지와 조각 3 이 맡아요.
- 제품 코드(`src/**`) 변경과 새 크레이트.

## Requirements

- REQ-1: The file tests/docs_scans.rs shall hold a Rust test for every case of the eleven docs and metadata Python modules that no Rust test already covers · failure: none — a case without a Rust test is found in the case table of the Result
- REQ-2: The file tests/skill_references_scan.rs shall hold a Rust test for every case of tests/skill_references_test.py · failure: none — a case without a Rust test is found in the case table of the Result
- REQ-3: The file tests/release_suite.rs shall hold a Rust test for every case of tests/dist_include_test.py · failure: none — a case without a Rust test is found in the case table of the Result
- REQ-4: Each ported Rust test shall read the document or manifest at run time and take its expected value from the source that has it · failure: none — a copied literal that can drift is a review finding
- REQ-5: When a document fact that a Python case pinned is broken, the matching Rust test shall fail · failure: none — a mutant that survives names a missing assertion
- REQ-6: The ported Rust tests shall run without Python and without a shell command · failure: none — a spawn of either is found by a scan of the new files
- REQ-7: The rules of tools/repo_scans.json that name tests.skill_references_test and tests.space_applied_record_test shall name the Rust test targets that replace them · failure: none — a rule that still names a Python module is found by the table check
- REQ-8: Each new Rust test file that reads files across the tree shall have its entry in tools/repo_scans.json · failure: none — a missing entry is refused by the table check
- REQ-9 (REQ-1): IF a Python case cannot be written as a Rust test, THEN the Result shall name the case with its reason and the Python module stays in place.
- REQ-10: This slice shall delete no Python file · failure: none — a deletion is found by the check in AC-6

## Acceptance

- AC-1 (REQ-1, REQ-4, REQ-6): 문서·메타데이터 시험 사례가 `tests/docs_scans.rs` 에서 모두 통과해요.
  verify: `cargo test --offline --locked --test docs_scans`
- AC-2 (REQ-2): 스킬 참조 시험이 통과해요.
  verify: `cargo test --offline --locked --test skill_references_scan`
- AC-3 (REQ-3): 릴리스 아카이브 구성 시험이 통과해요.
  verify: `cargo test --offline --locked --test release_suite -- dist_include`
- AC-4 (REQ-5): 문서 사실을 깨뜨리는 변형 5개 이하가 모두 잡혀요.
  verify: Result 절의 변형 표.
- AC-5 (REQ-7, REQ-8): 규칙 표가 새 대상과 새 파일 항목을 받아들여요.
  verify: `cargo test --offline --locked --test repo_scans_cli`
- AC-6 (REQ-10): 파이썬 파일이 하나도 지워지지 않아요.
  verify: `git diff --name-status $(git merge-base HEAD v2026.11.0)..HEAD | grep -E '^D.*[.]py$'` 이 아무것도 내지 않아요.
- AC-7 (REQ-9): 79개 사례가 모두 Rust 시험 이름이나 남긴 이유를 가져요.
  verify: Result 절의 사례 대조표.

## Scope

- write: `.moai/specs/SPEC-20261121-rust-only-text-scans/**`, `tests/docs_scans.rs`(새 파일), `tests/skill_references_scan.rs`(새 파일), `tests/release_suite.rs`, `tools/repo_scans.json`(보호 경로: 바꾸는 PR 은 check-guard 절차를 따라요)
- read-only: 위 파이썬 모듈 13개와 `tools/check_support_matrix.py`, 이들이 읽는 문서(`docs/**`, `README*`, `Cargo.toml`, `install/**`, `skills/**`)
- forbidden: `src/**`, 파이썬 파일 삭제와 수정, 실제 HOME, 다른 조각의 모듈

## Plan

1. RED: 사례마다 Rust 시험을 먼저 쓰고, 각 시험이 문서 사실을 깨뜨리면 실패하는 것을 변형으로 보여요(5개 이하).
2. GREEN: 새 시험 파일 둘과 `tests/release_suite.rs` 추가분이 통과하게 해요.
3. 규칙 표: `tools/repo_scans.json` 의 두 규칙(`tracked-skill-references`, `design-review-fixes`)이 Rust 대상을 가리키고, 새 파일 항목을 더해요. `tests/repo_scans_cli.rs` 와 `tests/check_affected_cli.rs` 가 파이썬 단계 문구를 단정하는 곳을 확인해요.
4. 대조표: 사례 79개와 Rust 시험 이름을 Result 에 적어요. 옮기지 못한 사례는 이유를 적고 그 모듈은 남겨요.
5. 이웃 검사: 새 파일이 파이썬이나 셸을 부르지 않는지 훑는 검사, 저장소 전체 단어 훑기 시험(새 이름이 걸리는지), `repo_scans` 표 점검.

뒤따르는 착지(각각 이 SPEC 의 Result 에 적거나 새 SPEC 으로 나눠요): B GitHub 양식·템플릿·비밀 모양(YAML 방식은 D4), C 스킬 글 픽스처(`setup_approval_fixture_test`, `sync_drift_fixture_test`)와 `repo_scans` 표 점검, D 워크플로·설치 스크립트·`local_check.sh`·`self_check.sh` 시험, E 바이너리를 실행하는 시험(`check_pr_*` 셋, 상태줄 래퍼, e2e 호스트).

## Decisions

결정은 되돌릴 수 있고 범주는 `none` 이에요. 결정자는 리더이고 기한은 이 SPEC 의 구현 시작 전이에요. 답이 없으면 기본값으로 진행해요.

- D8 `unittest` 단계와 파이썬 슬롯을 걷는 시점. 카드의 조각 2 문구는 이 조각에서 걷는다고 하지만, 도구 시험 약 30개가 조각 3·4 까지 남아서 지금 걷으면 그 시험들이 돌지 않아요. 선택지: (a) 마지막 파이썬 시험이 없어지는 조각에서 걷고 그 전에는 `check affected` 가 파이썬 모듈을 덜 고르게만 함, (b) 이 조각에서 걷고 남은 도구 시험은 손으로 돌림. 기본값: (a).
- D9 같은 사실을 파이썬과 Rust 가 둘 다 단정하는 기간(조각 4 까지). 선택지: (a) 조각 4 까지 둘 다 둠, (b) 옮긴 모듈은 이 조각에서 파이썬 쪽을 지움(카드 규칙은 지우기를 조각 4 로 못 박아서 승인 필요). 기본값: (a).

## Risks

- 문서 시험은 문서가 바뀔 때마다 깨져서, Rust 로 옮기면서 "문서를 고치는 사람이 시험도 같이 고쳐야 한다"는 부담이 그대로 남아요. 기대값을 소스에서 끌어오는 것(REQ-4)이 그 부담을 줄이는 방법이에요.
- `oss_docs_test.py` 의 사례 중 점검기 자신을 시험하는 단위 사례(`signature_reader_reads_the_source`, `forbidden_samples_are_flagged` 등)는 Rust 점검기 함수가 있어야 옮길 수 있어요. 함수를 시험 파일 안에 두고 그 함수를 시험해요.
- `tools/repo_scans.json` 은 보호 경로라서 이 파일을 바꾸는 PR 은 check-guard 의 절차를 거쳐요.
- `tests/release_suite.rs` 를 다른 카드가 쥐고 있을 수 있어요. 구현 전에 예약을 확인해서 쥐고 있으면 짝 요청을 남겨요.

## Result

이 head 는 SPEC 만이에요. 구현과 대조표는 구현 착지에서 채워요.
