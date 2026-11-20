---
id: SPEC-20261120-decide-leftovers
title: "Close the last decision model leftovers"
status: approved
revision: 1
created: 2026-11-20
approval: {by: user, at: 2026-11-21, basis: "AskUserQuestion selection: 승인 (Recommended)", digest: "sha256:850f84f998c0783384ffd83e121137b370b5926c51a8818c23c59e5da2338ee4"}
---

# Close the last decision model leftovers

## Intent

SPEC-20261120-retire-decide 가 결정 모델 라이브러리(`src/decide/`)를 지운 뒤 남은 두 가지를 정리해요. (1) 초안 SPEC 세 개(card-secret-screen, c68-audit-followups, secret-shape-words)가 지워진 `src/decide/outbound.rs`·`normalise.rs`를 읽기 전용 참조로 적고 있어 읽는 사람이 헷갈려요. 기록이라 본문은 두고 안내 줄만 달아요. (2) `typesafe` 라는 낱말이 남은 곳을 모두 읽어 보니 세 종류예요: 예전 세대가 사용자 HOME 에 만든 비밀 파일 이름 `.env.typesafe`(`src/home_layout.rs`, `src/home_retire.rs`와 그 시험 — 실제 파일의 이전 처리 규칙이라 그대로 둠), 예전 프로필 호환을 고정하는 시험(`tests/project_decide_model.rs` — 그대로 둠), 키 보관소 시험의 예시 공급자 이름(`tests/key_set_prompt_tty_test.py` — 공급자 이름은 아무 소문자 이름이나 되므로 중립 이름으로 바꿈).

## Requirements

- REQ-1: When a reader opens one of the three draft SPECs, the SPEC shall state under its title that `src/decide/` was removed by SPEC-20261120-retire-decide and that its read-only references to that folder are historical · failure: none — a documentation change; `moai spec check` is the check.
- REQ-2: When `tests/key_set_prompt_tty_test.py` runs, the test shall use the neutral sample provider name `demo` in place of `typesafe` with every assertion keeping its meaning · failure: none — the test run is the check.
- REQ-3: When a Go-generation key file named `.env.typesafe` is classified, `src/home_layout.rs` and `src/home_retire.rs` shall keep treating it as a secret that is retired by rename only · failure: none — no source change; the existing tests pin it.

## Acceptance

- AC-1 (REQ-1): 세 SPEC 에 안내 줄이 있고 검사 오류가 없음. verify: `grep -c 'SPEC-20261120-retire-decide' .moai/specs/SPEC-20261119-card-secret-screen/spec.md .moai/specs/SPEC-20261119-c68-audit-followups/spec.md .moai/specs/SPEC-20261117-secret-shape-words/spec.md`; `moai spec check --all`
- AC-2 (REQ-2): 키 입력 시험이 새 이름으로 통과하고 `typesafe` 가 그 파일에 없음. verify: `python3 -m unittest tests.key_set_prompt_tty_test`; `grep -c typesafe tests/key_set_prompt_tty_test.py`
- AC-3 (REQ-3): 파일 이름 규칙 시험이 그대로 통과(소스 무변경). verify: `cargo test --locked --offline --lib home_layout home_retire`; `cargo test --locked --offline --test home_retire_cli`; `git diff --exit-code v2026.11.0 HEAD -- src`
- AC-4: 가벼운 검사. verify: `git diff --check`; `cargo fmt --check`

## Scope

- write: `.moai/specs/SPEC-20261119-card-secret-screen/spec.md`, `.moai/specs/SPEC-20261119-c68-audit-followups/spec.md`, `.moai/specs/SPEC-20261117-secret-shape-words/spec.md`, `tests/key_set_prompt_tty_test.py`, `.moai/specs/SPEC-20261120-decide-leftovers/spec.md`
- read-only: `src/home_layout.rs`, `src/home_retire.rs`, `tests/home_retire_cli.rs`, `tests/project_decide_model.rs`, `src/key_vault.rs`
- forbidden: 실제 HOME·키 보관소 접근, 네트워크, 카드 번호를 SPEC id·브랜치·커밋에 사용, `templates/bundle-packs/` 의 Go 세대 스킬 글(제품 내용이라 별도 결정)

## Interview

- Q1: 초안 SPEC 3개의 `src/decide/` 참조와 `typesafe` 이름 잔재를 모두 처리할까요? → A: 모두 처리 (사용자, 2026-11-21 대화)
- Q2: 위 범위(안내 줄 추가, 예시 공급자 이름 교체, `.env.typesafe` 규칙·예전 프로필 호환 시험 무변경)로 승인할까요? → A: 승인 (Recommended)

## Plan

1. 세 SPEC 의 제목 바로 아래에 `> 안내(2026-11-21): src/decide/ 는 SPEC-20261120-retire-decide 에서 삭제됐어요. 아래 읽기 전용 참조는 기록이에요.` 한 줄을 넣고 `spec check` 를 확인.
2. `tests/key_set_prompt_tty_test.py` 의 예시 공급자 `typesafe` 를 `demo` 로 바꾸고(인자, 파일 이름 `demo.key`, 프롬프트 문구 단언) 시험을 돌림.
3. 파일 이름 규칙 시험을 소스 무변경 상태로 돌려 `.env.typesafe` 처리가 그대로임을 확인.
4. `templates/bundle-packs/` 의 TypeSafe 언급 정리는 후속 과제로 남김.
