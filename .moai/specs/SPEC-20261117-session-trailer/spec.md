---
id: SPEC-20261117-session-trailer
title: 관문 신원 trailer — Moai-Spec와 Moai-Session (SDD-6)
status: draft
revision: 1
created: 2026-11-17
parent: SPEC-20261117-sdd-foundation
---

## Intent

사용자 결정 U-B(2026-11-17, guarded): 병합 관문의 신원은 커밋 trailer `Moai-Spec: <SPEC-ID>`와 `Moai-Session: <session>`이고,
계획했던 `Moai-Claim: C-<n>#<gen>`을 대신해요(카드 번호는 공유 장소에 쓰지 않음). 이 SPEC은 session id가 무엇인지 정하고, trailer를
만드는 `moai trailers`와 범위를 확인하는 `moai trailer-check`를 더하고, audit-gate가 판정의 감사자 session을 같은 모양으로 읽게 해요.

## Non-goals

- 커밋 hook이나 commit guard가 trailer를 강제하기(`claim_verify`·`hook_engine`은 C-80이 감사 중이라 건드리지 않음).
- 리더 병합 관문 G9–G11을 trailer로 바꾸기(다음 조각).
- `internal/sdd.md` 고치기(SDD-5가 아직 병합되지 않음; 규칙은 새 참조 `internal/commit-trailers`에).
- launcher가 OMP 세션에 id 변수를 넣기.

## Requirements

- REQ-1: The session id shall be the value `session-current` reads from the host environment (`CLAUDE_CODE_SESSION_ID`, then `CODEX_THREAD_ID`), accepted only as 8 to 64 hex digits and `-`.
- REQ-2: When an agent asks for its commit trailers with a SPEC id, `moai trailers` shall print `Moai-Spec: <id>` and `Moai-Session: <session id>` as trailer lines.
- REQ-3 (REQ-2): IF no host variable gives a session id in the accepted shape, THEN `moai trailers` shall print the `Moai-Spec` line only and report `session_missing` with the reason.
- REQ-4 (REQ-2): IF the SPEC id is not of the form `SPEC-<yyyymmdd>-<slug>`, THEN `moai trailers` shall refuse it and print no trailer.
- REQ-5: When `moai trailer-check` reads a revision range, it shall report every commit without `Moai-Spec`, with a malformed `Moai-Spec` or `Moai-Session`, or without `Moai-Session`. · failure: none — 실패 판정은 REQ-6·REQ-7이에요.
- REQ-6 (REQ-5): IF a commit's `Moai-Spec` or `Moai-Session` value is malformed or repeated, THEN the check shall fail.
- REQ-9: Where the checked branch is not a `spec/` branch, the check shall report a missing `Moai-Spec` without failing.
- REQ-7 (REQ-5): IF the range starts with `-` or git cannot read it, THEN the check shall refuse before reading any commit.
- REQ-8: Where a verdict comment carries a `Moai-Session:` line below its first line, the audit gate shall treat that session as the auditor's identity next to its lane name.

## Acceptance

- AC-1 (REQ-1…REQ-4): trailer 줄, session 없음, 모양이 틀린 host 값, 카드 번호·옛 SPEC id 거절.
  verify: `cargo test --locked --offline --test trailer_check_cli trailers_name_the_spec_and_this_sessions_id`; `cargo test --locked --offline --lib trailers::`
- AC-2 (REQ-5…REQ-7, REQ-9): 임시 저장소의 범위에서 다섯 가지 발견, `spec/` 가지와 아닌 가지의 실패 수, 옵션 모양 범위 거절.
  verify: `cargo test --locked --offline --test trailer_check_cli trailer_check_reports_missing_and_malformed_trailers`
- AC-3 (REQ-8): 판정의 `Moai-Session`이 저자 session이면 PASS가 아니고, FAIL은 같은 session의 PASS만 풀어요.
  verify: `cargo test --locked --offline --lib audit_gate::`; `cargo test --locked --offline --test audit_gate_cli`
- AC-4 (REQ-1…REQ-9): 이웃 검사.
  verify: `cargo test --locked --offline --test gate_check_cli`; `cargo test --locked --offline --test git_env_scrub`;
  `cargo test --locked --offline --test home_isolation_scan`; `cargo test --locked --offline --test product_cli`;
  `cargo test --locked --offline --test help_layout`; `cargo test --locked --offline --bin moai`;
  `cargo test --locked --offline --test agent_defs_package`; `cargo fmt --check`;
  `cargo clippy --locked --offline --all-targets -- -D warnings`;
  `python3 -m unittest tests/repo_scans_test.py tests/ci_scope_test.py tests/source_manifest_test.py`

## Scope

- write: `src/trailers.rs`, `src/ops/trailers.rs`, `src/ops/trailer_check.rs`, `src/lib.rs`, `src/help.rs`, `src/audit_gate.rs`,
  `tests/trailer_check_cli.rs`, `tests/agent_defs_package.rs`, `tests/fixtures/help/*.txt`,
  `skills/moai/references/internal/commit-trailers.md`, `skills/moai/references/chains/index.md`,
  `skills/moai/references/source-manifest.json`, `docs/design-review/appendix/operations/trailers.md`,
  `docs/design-review/appendix/operations/trailer_check.md`, `docs/design-review/appendix/operations/audit_gate.md`,
  `.moai/specs/SPEC-20261117-session-trailer/spec.md`
- read-only: `src/session_current.rs`, `src/spec_doc.rs`, `src/gate.rs`
- forbidden: `src/claim_verify.rs`, `src/hook_engine.rs`(C-80 감사 중), `skills/moai/workflows/mode-agent.md`(크기 한도), 실제 HOME

## Plan

1. 시험 먼저: CLI 시험(임시 저장소·HOME), audit-gate의 감사자 session 시험.
2. `src/trailers.rs`(형식·로그 해석·판정), 명령 둘, audit-gate의 `Moai-Session` 줄 읽기.
3. 새 참조 `internal/commit-trailers`, chains 규칙 한 줄, skill 목록.

## Risks

- OMP와 `codex exec` 같은 곳에는 session id 변수가 없을 수 있어요 → 그 커밋은 `Moai-Spec`만 갖고, 감사 관문은 저자를 모르는 것으로 읽어 tier 1·2 PR이 기다려요. launcher가 변수를 넣는 일은 다음 조각.
- `CODEX_THREAD_ID`는 Codex 문서에 없는 변수예요(`session_current`의 주석) → Codex lane이 재서 확인해야 해요.
- 감사 판정이 `Moai-Session` 줄을 쓰지 않으면 감사자 비교는 lane 이름만으로 해요(예전과 같음).

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D1 | session id = 기존 `session-current`가 읽는 host 변수 값(새 출처 없음) | 이미 있는 것 재사용, 비밀이 아님, 카드 번호 아님 | 출처 목록 바꾸기 |
| D2 | 새 명령 `moai trailers --spec`과 `moai trailer-check <range>` | 조정자 지시 | 명령 빼기 |
| D3 | `Moai-Session` 없음은 알리기만 하고 실패시키지 않음; `Moai-Spec` 없음은 `spec/` 가지에서만 실패 | session id가 없는 host가 있음 | 둘 다 실패로 |
| D4 | 규칙은 새 참조 `internal/commit-trailers`(SDD-5의 `sdd.md`가 아직 없음), chains 규칙이 이름을 부름; `mode-agent.md`는 그대로 | 조정자 지시, 크기 한도 | SDD-5 뒤 `sdd.md`로 옮기기 |
| D5 | 판정의 감사자 session은 첫 줄 아래 `Moai-Session:` 줄(커밋과 같은 키·모양); 같은 감사자 = session이 같을 때, 없으면 lane | 형식 맞추기(조정자 지시 4) | lane만 비교 |

## Result

아직 비어 있어요(구현 뒤 채워요).
