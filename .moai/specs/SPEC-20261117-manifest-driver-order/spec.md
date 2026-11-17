---
id: SPEC-20261117-manifest-driver-order
title: source-manifest 순서를 하나로(병합 드라이버와 write가 같은 바이트, 중복 키 거절)
status: draft
revision: 1
created: 2026-11-17
parent: null
---

## Intent

카드 C-92. `source_manifest.py write`와 병합 드라이버는 `files` 항목을 기존 순서대로(우리 쪽 먼저, 상대에만 있는 것은
끝에) 써서, 같은 항목이 가지마다 다른 줄에 있을 수 있었어요. 2026-11-17 #3162를 GitHub가 글자로 병합하자 tip
`17c19a2759`의 manifest에 `references/internal/build-folder.md` 키가 두 번 들어갔고, push CI의
`source_manifest_test.test_this_repository_manifest_is_what_the_tool_writes`가 실패했어요. JSON을 읽으면 뒤의 값이 이겨서
중복은 조용히 지나갔어요.

## Non-goals

- 항목 안의 키 순서(`sha256`, `origin`)나 위쪽 키 순서 바꾸기.
- GitHub의 병합 방식 바꾸기(드라이버는 로컬 git에서만 돌아요).

## Requirements

- REQ-1: The manifest renderer shall write the `files` keys in sorted order, for `write`, `check` and the merge driver alike. · failure: none — one serializer; REQ-2 covers a broken input.
- REQ-2 (REQ-1): IF a manifest holds a key twice, THEN `check` and `write` shall stop with exit 2 and name the key.
- REQ-3: When the driver merges two manifests, its output shall be byte-identical to `write` on the same merged tree. · failure: none — the driver's own refusals stay as before.

## Acceptance

- AC-1 (REQ-1): 손으로 끝에 붙인 새 항목을 `check`가 거절하고(exit 1), `write` 뒤에는 정렬된 자리(기존 키 사이)에 있어요.
  verify: `python3 -m unittest tests.source_manifest_test -k sorted_order`
- AC-2 (REQ-2): 키가 두 번 있으면 `check`와 `write`가 exit 2로 멈추고 그 키를 말해요.
  verify: `python3 -m unittest tests.source_manifest_test -k duplicate_key`
- AC-3 (REQ-3): 세 가지 병합에서 드라이버 결과와 글자 충돌 뒤 `write` 결과가 바이트로 같아요(기존 키 사이에 들어가는 새 파일 포함).
  verify: `python3 -m unittest tests.source_manifest_test -k bytes_write_writes`
- AC-4 (REQ-1, REQ-2): 오늘의 경우(같은 새 항목을 양쪽이 가짐, 한쪽은 드라이버가 씀, 마지막은 글자 병합)에 키가 두 번 남지 않아요.
  verify: `python3 -m unittest tests.source_manifest_test -k never_duplicate`
- AC-5 (REQ-1…REQ-3): 이 저장소의 manifest가 새 규칙으로 맞고 이웃 검사가 통과해요.
  verify: `python3 -m unittest tests.source_manifest_test tests.repo_scans_test tests.ci_scope_test tests.skill_references_test`; `python3 -I tools/source_manifest.py check`;
  `cargo test --locked --offline --test product_cli skill_package`; `cargo test --locked --offline --test agent_defs_package`

## Scope

- write: `tools/source_manifest.py`, `tools/merge_driver_manifest.py`, `tests/source_manifest_test.py`,
  `skills/moai/references/source-manifest.json`(정렬과 중복 하나 제거), `.moai/specs/SPEC-20261117-manifest-driver-order/spec.md`
- read-only: `.gitattributes`
- forbidden: 실제 HOME, 네트워크 쓰기(push, PR, Issue)

## Plan

1. 시험 먼저(RED): 정렬, 중복 키, 드라이버 == write, 오늘의 경우.
2. `parse`에 `object_pairs_hook`(중복 키 거절), `render`가 `files`를 정렬. 드라이버는 이미 `render`·`parse`를 써요.
3. 이 저장소 manifest: tip의 중복 하나를 손으로 빼고 `write`.
4. 시험, mutation 3개.

## Risks

- 이번 한 번 manifest 전체 순서가 바뀌어요 → 열린 가지는 다음 병합에서 manifest가 충돌할 수 있어요. 드라이버나 `write`가
  풀고, 이후로는 순서가 하나라 같은 문제가 다시 생기지 않아요.
- 각 clone에 설치된 드라이버 사본은 이 변경 뒤 `--check`가 `stale`이에요 → 병합 뒤 `python3 -I tools/merge_driver_manifest.py --install`을 다시 돌려야 해요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D1 | 순서는 `files` 키 정렬(코드 포인트 순) | 역사와 상관없는 한 가지 순서; 새 파일이 기존 키 사이에 들어감 | 기존 순서 유지 |
| D2 | 중복 키는 `check`·`write` 모두 exit 2(사람이 고름) | 어느 쪽 값이 맞는지 도구가 모름 | 뒤의 값 채택 |

## Result

구현 에이전트가 2026-11-18에 채웠어요(UTC 2026-11-17 16:44–16:56Z). 감사가 다시 확인해요. 상태는 `draft` 그대로예요.
기준 tip `17c19a2759`. RED: 새 시험 4개가 바뀌기 전 코드에서 실패 — 끝에 붙인 항목이 그대로 남음, 중복 키가 이름 없이
일반 불일치(exit 1)로만 나옴, 드라이버 결과와 `write` 결과의 순서가 다름, 오늘의 경우 글자 병합이 깨끗하게 끝나며 키가
두 번 남음 → GREEN.

- AC-1~AC-4: met — `tests.source_manifest_test` 43개 OK(새 시험 4개 포함; 원래 끝 순서를 가정하던 두 시험은 정렬 순서로 고침).
  mutation M1(`render`가 주어진 순서 유지) KILLED, M2(중복 거절 빼기) KILLED, M3(맨 위 키만 중복 검사) KILLED
- AC-5: met — `python3 -I tools/source_manifest.py check` 0(tip의 중복 하나를 손으로 빼고 `write`; 46개 항목 정렬),
  Python `repo_scans`·`ci_scope`·`skill_references` 49개 OK, `product_cli` 75, `agent_defs_package` 18, `codex_adapter` 22,
  `moai spec check --all` 0

실행 안 함: 전체 `cargo test`, GitHub에서의 실제 병합(로컬 `git merge`로 같은 글자 병합을 재현).
