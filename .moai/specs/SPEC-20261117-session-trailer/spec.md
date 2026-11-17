---
id: SPEC-20261117-session-trailer
title: 관문 신원 trailer — Moai-Spec와 Moai-Session (SDD-6)
status: draft
revision: 3
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

- REQ-1: The `Moai-Session` value shall be `<session id>:<worktree key>`: the id `session-current` reads from the host environment (`CLAUDE_CODE_SESSION_ID`, then `CODEX_THREAD_ID`, 8 to 64 hex digits and `-`) and the first 12 hex digits of the sha256 of the worktree's canonical git directory.
- REQ-2: When an agent asks for its commit trailers with a SPEC id, `moai trailers` shall print `Moai-Spec: <id>` and `Moai-Session: <session id>` as trailer lines.
- REQ-3 (REQ-2): IF no host variable gives a session id in the accepted shape, or the folder is in no git worktree, THEN `moai trailers` shall print the `Moai-Spec` line only and report `session_missing` with the reason.
- REQ-4 (REQ-2): IF the SPEC id is not of the form `SPEC-<yyyymmdd>-<slug>`, THEN `moai trailers` shall refuse it and print no trailer.
- REQ-5: When `moai trailer-check` reads a revision range, it shall report every commit without `Moai-Spec`, with a malformed `Moai-Spec` or `Moai-Session`, or without `Moai-Session`. · failure: none — 실패 판정은 REQ-6·REQ-7이에요.
- REQ-6 (REQ-5): IF a commit's `Moai-Spec` or `Moai-Session` value is malformed or repeated, THEN the check shall fail.
- REQ-9: Where the checked branch is not a `spec/` branch, the check shall report a missing `Moai-Spec` without failing.
- REQ-10 (REQ-5): IF any commit fails the check, THEN `moai trailer-check` shall exit 1 and still print its findings.
- REQ-11 (REQ-5): IF a trailer value holds the log's field separator, THEN the check shall report that commit as `malformed_trailers` and fail.
- REQ-12 (REQ-5): IF a `Moai-Session` value is a bare session id without the worktree key, THEN the check shall report it as `legacy_session` without failing.
- REQ-13: Where `--branch` names a branch, the check shall judge the range for that branch instead of the checkout's HEAD.
- REQ-14: When the audit gate compares identities, it shall match two full values exactly and a legacy bare id with every value of its session.
- REQ-7 (REQ-5): IF the range starts with `-` or git cannot read it, THEN the check shall refuse before reading any commit.
- REQ-8: Where a verdict comment carries a `Moai-Session:` line below its first line, the audit gate shall treat that session as the auditor's identity next to its lane name.

## Acceptance

- AC-1 (REQ-1…REQ-4): trailer 줄, session 없음, 모양이 틀린 host 값, 카드 번호·옛 SPEC id 거절.
  verify: `cargo test --locked --offline --test trailer_check_cli trailers_name_the_spec_and_this_sessions_id`; `cargo test --locked --offline --lib trailers::`
- AC-2 (REQ-5…REQ-7, REQ-9…REQ-13): 임시 저장소의 범위에서 다섯 가지 발견, `spec/` 가지와 아닌 가지의 실패 수, 옵션 모양 범위 거절.
  verify: `cargo test --locked --offline --test trailer_check_cli trailer_check_reports_missing_and_malformed_trailers`; `cargo test --locked --offline --lib trailers::tests::a_value_with_a_separator_or_control_character_is_malformed`
- AC-3 (REQ-8, REQ-14): 판정의 `Moai-Session`이 저자 session이면 PASS가 아니고, FAIL은 같은 session의 PASS만 풀어요.
  verify: `cargo test --locked --offline --lib audit_gate::`; `cargo test --locked --offline --test audit_gate_cli`
- AC-5 (REQ-1, REQ-3): 한 저장소의 두 worktree는 같은 session 환경에서 다른 값을, 한 worktree는 늘 같은 값을 받아요(하위 폴더 포함).
  verify: `cargo test --locked --offline --test trailer_check_cli the_session_value_names_the_worktree`; `cargo test --locked --offline --lib trailers::`
- AC-4 (REQ-1…REQ-14): 이웃 검사.
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
- (감사 N1) `Moai-Session` 값은 스스로 적는 값이에요: 저자가 아무 값이나 쓸 수 있고, 관문은 lane 이름처럼 그 값을 믿어요 → 실수는
  잡지만 일부러 속이는 것은 못 막아요; 독립 감사와 리더 확인이 그 몫이에요(`internal/commit-trailers` 4).
- (감사 N2) session id는 작업 내내 같지 않아요: Claude Code는 `/clear` 뒤에 새 id를 줘요(환경 변수 문서 "updated on `/clear`") → 한
  저자의 커밋이 여러 id를 가질 수 있고, lane 이름 비교가 대비책이에요.
- `CODEX_THREAD_ID`는 Codex 문서에 없는 변수예요(`session_current`의 주석) → Codex lane이 재서 확인해야 해요.
- 감사 판정이 `Moai-Session` 줄을 쓰지 않으면 감사자 비교는 lane 이름만으로 해요(예전과 같음).

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D1 | session id = 기존 `session-current`가 읽는 host 변수 값(새 출처 없음); `/clear` 뒤 바뀌고 스스로 적는 값(감사 N1·N2) | 이미 있는 것 재사용, 비밀이 아님, 카드 번호 아님 | 출처 목록 바꾸기 |
| D7 | `Moai-Session` = `<host session id>:<worktree key>`(정식 git dir의 sha256 앞 12자리): Claude Code 하위 에이전트가 부모의 session id를 물려받아 id만으로는 저자와 감사 에이전트를 가르지 못해요(재감사 B3, 감사자가 같은 `c3f1ed16-…`를 측정). 감사 에이전트는 자기 worktree에서 돌아요. 옛 맨 id는 한 번의 이행 기간 동안 `legacy_session`(실패 아님)으로 받고, 관문은 그 id를 그 session 전체로 봐요(보수적). `--branch`(재감사 N4) | 리더 결정 2026-11-17(기록, 되돌릴 수 있음) | 맨 session id로 돌아가기 |
| D6 | `trailer-check`는 실패가 있으면 결과 JSON을 낸 채 exit 1(감사 B1, `audit-delta`와 같은 방식); 구분 문자를 품은 값은 `malformed_trailers`(감사 N3) | CI·스크립트가 종료 코드로 판단 | exit 0 유지 |
| D2 | 새 명령 `moai trailers --spec`과 `moai trailer-check <range>` | 조정자 지시 | 명령 빼기 |
| D3 | `Moai-Session` 없음은 알리기만 하고 실패시키지 않음; `Moai-Spec` 없음은 `spec/` 가지에서만 실패 | session id가 없는 host가 있음 | 둘 다 실패로 |
| D4 | 규칙은 새 참조 `internal/commit-trailers`(SDD-5의 `sdd.md`가 아직 없음), chains 규칙이 이름을 부름; `mode-agent.md`는 그대로 | 조정자 지시, 크기 한도 | SDD-5 뒤 `sdd.md`로 옮기기 |
| D5 | 판정의 감사자 session은 첫 줄 아래 `Moai-Session:` 줄(커밋과 같은 키·모양); 같은 감사자 = session이 같을 때, 없으면 lane | 형식 맞추기(조정자 지시 4) | lane만 비교 |

## Result

구현 에이전트가 채웠어요. 감사가 다시 확인해요. 기준 `5bf6f29a32`(audit-gate head) 위, audit-gate 수리 `0b12c389b8`를
합침(`8d586f16ab`), 이 worktree의 `target/`, `CARGO_INCREMENTAL=0`, `CARGO_BUILD_JOBS=4`, macOS arm64, cargo는 한 번에 하나.

RED(2026-11-17T10:03:13Z): `trailer_check_cli` 2개는 명령이 없어 usage(exit 2), `audit_gate::tests::a_verdict_names_its_auditors_session_like_a_commit_does`는
저자 session의 PASS를 통과시킴(`pass` ≠ `missing`). 통과 10:06:13Z.

- AC-1: met — `trailer_check_cli trailers_name_the_spec_and_this_sessions_id`, `--lib trailers::` 2
- AC-2: met — `trailer_check_cli trailer_check_reports_missing_and_malformed_trailers`(임시 저장소)
- AC-3: met — `--lib trailers:: audit_gate:: gate:: session_current:: spec_doc::` 38, `audit_gate_cli` 3
- AC-4: met — `gate_check_cli` 45, `gate_collect` 44, `gate_merge` 16, `gate_stacked_2859` 9, `audit_admission` 7, `git_env_scrub` 3, `home_isolation_scan` 16, `product_cli` 75, `help_layout` 10, `--bin moai` 27, `agent_defs_package` 18, `github_assets_test` 37 OK, 저장소 scan 대상 모두 exit 0; fmt 0, `rustfmt src/ops` 0, clippy 0; Python 88 OK; manifest ok; `package-inspect` ok; `moai spec check --all` ok
- 스스로 쓰기: 이 가지의 커밋은 `moai trailers --spec SPEC-20261117-session-trailer`가 만든 줄을 써요(`session_source: env:CLAUDE_CODE_SESSION_ID`).
  `moai trailer-check 0b12c389b8..HEAD`: ok, 커밋 2개, 실패 0 — 첫 WIP 커밋(`2ef8f25529`)은 `Moai-Session`이 없어 알림만(그 커밋은 이 명령이 생기기 전).

mutation 5개 모두 KILLED: Go 세대 SPEC id 받기, `Moai-Spec` 없음을 모든 가지에서 실패, 아무 `Moai-Session` 값 받기, `-`로 시작하는 범위를 git에 넘기기,
판정의 감사자 session 안 읽기.

실행 안 함: 실제 Codex·OMP 세션의 id 변수 측정, 전체 `cargo test`, Linux·Windows.
