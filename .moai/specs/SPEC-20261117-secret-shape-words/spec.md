---
id: SPEC-20261117-secret-shape-words
title: 비밀 모양 검사가 task-based 같은 보통 낱말을 거절하지 않게
status: draft
revision: 1
created: 2026-11-17
parent: null
---

## Intent

카드 C-90. 프로필 글 검사 `profile::portable()`은 공유 목록 `SECRET_SHAPES`의 OpenAI식 키 앞부분(s, k, 하이픈)을
글 어디에서나 찾아서 `task-based`, `risk-free`, `desk-top` 같은 보통 낱말도 거절했어요. 같은 목록을 쓰는 곳은
`scopes::path_ok`, `install` 영수증 필드, `redact::contains_secret_shape`(hook 영수증·로그 필드, 메시지, goal, forge
보고)예요. `spec::check_text`는 이미 그 앞부분을 낱말 시작에서만 셌어요. 이 규칙을 하나의 함수
`crate::secret_shaped`로 모아 모든 사용처가 같게 판정해요. 다른 모양(`ghp_`, `github_pat_`, PEM 머리, 비밀번호·
토큰·API 키 대입)은 그대로 어디서든 거절해요.

## Non-goals

- 키 뒤 길이 조건 더하기: 짧은 값(`install_cli`의 앞부분 뒤 6글자짜리 영수증 고정 시험 같은 것)이 통과하게 되어 거절이 약해져요.
- 쓰기 hook(`hook_guards::carries_secret`), `agents_local_merge::secret_kind`, `decide::outbound`, `redact_outgoing`,
  진단 가리기 `redact_secret_shapes`: 이 목록 판정을 쓰지 않거나 이미 자기 규칙이 있어요. 동작이 바뀌지 않아요.

## Requirements

- REQ-1: The shared check shall count the OpenAI-style key prefix only at the start of a word, first in the text or after a character that is not an ASCII letter or digit. · failure: none — a pure predicate; REQ-3 pins what still refuses.
- REQ-2: Every caller of the shared shape list shall use that one check. · failure: none — the callers are named in Scope.
- REQ-3 (REQ-1, REQ-2): IF the prefix starts a word or any other listed shape appears anywhere, THEN the caller shall refuse the text as before.
- REQ-4: The documents that state the custom-text rules shall state the word-start rule. · failure: none — text, pinned by AC-3.

## Acceptance

- AC-1 (REQ-1, REQ-3): 앞 글자 경계(ASCII 글자·숫자 `0` `9` `a` `z` `A` `Z`는 낱말 안, 이웃 기호 `/` `:` `@` `[` `` ` `` `{`와
  공백·`_`·`-`·비ASCII 글자·맨 앞은 낱말 시작), 대소문자, 다른 모양의 낱말 안 위치가 시험으로 고정돼요.
  verify: `cargo test --locked --offline --lib redact`
- AC-2 (REQ-2, REQ-3): 프로필 글, 범위 경로, CLI `project-plan`이 `task-based`·`risk-free`·`desk-top`을 받고, 낱말 시작
  키 모양과 기존 거절(`product_cli`의 프로필 고정 시험, `install_cli`의 영수증 고정 시험)은 그대로 거절해요.
  verify: `cargo test --locked --offline --lib profile`; `cargo test --locked --offline --lib scopes`; `cargo test --locked --offline --test product_cli`; `cargo test --locked --offline --test install_cli`
- AC-3 (REQ-4): `internal/versioning`의 사용자 글 규칙과 `internal/interview`의 안내 문장이 고정돼요.
  verify: `cargo test --locked --offline --test agent_defs_package`; `python3 -I tools/source_manifest.py check`
- AC-4 (REQ-1…REQ-4): 이웃 검사가 통과해요.
  verify: `cargo test --locked --offline --lib spec`; `cargo test --locked --offline --test spec_cli`; `cargo test --locked --offline --test home_isolation_scan`;
  `cargo test --locked --offline --bin moai`; `cargo fmt --check`; `cargo clippy --locked --offline --all-targets -- -D warnings`;
  `python3 -m unittest tests.repo_scans_test tests.ci_scope_test tests.source_manifest_test`

## Scope

- write: `src/lib.rs`, `src/profile.rs`, `src/scopes.rs`, `src/install.rs`, `src/redact.rs`, `src/spec.rs`,
  `tests/product_cli.rs`, `tests/agent_defs_package.rs`, `skills/moai/references/internal/versioning.md`,
  `skills/moai/references/internal/interview.md`, `skills/moai/references/source-manifest.json`, `docs/rust-product-path.md`,
  `.moai/specs/SPEC-20261117-secret-shape-words/spec.md`
- read-only: `src/hook_guards.rs`, `src/agents_local_merge.rs`, `src/decide/outbound.rs`
- forbidden: 실제 HOME, 네트워크 쓰기(push, PR, Issue)

## Plan

1. 사용처 찾기: `SECRET_SHAPES`와 키 앞부분 문자열을 `src/`에서 grep.
2. `crate::secret_shaped`를 `lib.rs`에 두고 profile·scopes·install·redact·spec이 그것을 부르게.
3. 경계 시험(앞 글자 N/N+1), 프로필·경로·CLI 시험, 문서 두 곳과 운영 문서, 고정 문장.
4. 시험·형식·clippy·Python, mutation 3개.

## Risks

- 낱말 시작에서 이 앞부분으로 시작하는 보통 낱말은 여전히 거절돼요 → 드물고, 길이 조건을 더하면 짧은 키 모양 거절이
  약해져서 이번에는 넣지 않아요(D2).
- `contains_secret_shape`를 쓰는 hook 영수증·로그 필드는 이제 `task-based` 같은 값을 기록해요 → 키 모양은 그대로 빠져요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D1 | 규칙을 `spec::check_text`의 기존 낱말 시작 규칙으로 통일 | 이미 시험된 규칙(#2001 r26), 거절이 약해지는 곳 없음 | `secret_shaped`를 모든 곳 substring으로 |
| D2 | 키 뒤 길이 조건은 넣지 않음 | 리더 지시는 "실제 키 모양 거절을 약하게 하지 말 것"; 뒤 20글자 하한이면 앞부분 뒤 6글자(영수증)와 14글자(프로필) 고정 시험 값이 통과하게 됨 | 앞부분 뒤 글자 수 하한 추가 |
| D3 | `redact_secret_shapes`(진단 가리기)는 그대로 substring | 더 가리는 쪽이라 안전, 카드 밖 | 같은 함수로 바꾸기 |

## Result

구현 에이전트가 2026-11-17에 채웠어요. 감사가 다시 확인해요. 상태는 `draft` 그대로예요(승인은 사용자 몫).
기준 `2489ad7c39`. 빌드 폴더는 이 worktree의 `target/`, `CARGO_INCREMENTAL=0`, `CARGO_BUILD_JOBS=4`, 무거운 작업 슬롯
안에서 cargo 한 번에 하나. RED는 mutation으로 보였어요: 옛 동작(어디서나 셈)을 되살리면 새 시험이 실패해요.

- AC-1: met — `--lib redact`의 `sk_counts_only_at_the_start_of_a_word`. mutation M1(어디서나 셈, 카드의 버그) KILLED,
  M2(숫자를 낱말 밖으로 봄) KILLED
- AC-2: met — `--lib profile`·`scopes` 시험, `product_cli` 76(새 `ordinary_words_with_sk_inside_are_accepted` 포함),
  `install_cli` 7. mutation M3(프로필이 공유 검사를 건너뜀) KILLED
- AC-3: met — `agent_defs_package` 18, manifest check 0
- AC-4: met — `--lib` profile·scopes·install·redact·spec·hook_engine·hooks·message·goal·forge::report 154,
  `spec_cli` 17, `home_isolation_scan` 16, `archive_privacy_suite` 13, `go_state_cli` 61, `card_summary_cli` 16,
  `public_text` 6, `binary_state_cli` 11, `--bin moai` 27; fmt 0, `rustfmt src/ops` 0, clippy 0; Python 4개 모듈 85개 OK;
  `moai spec check --all` 0; `moai package-inspect skills/moai` 0

수리 한 번: 첫 Python 실행에서 `ci_scope_test`가 `scopes` 시험 글자의 `docs/` 경로를 문서 참조로 읽었어요 → 시험 경로를
`web/`로 바꾸고 다시 돌려 85개 OK.

실행 안 함: 전체 `cargo test`(바뀐 곳 시험만), Linux·Windows.
