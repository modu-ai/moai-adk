---
id: SPEC-20261120-retire-decide
title: "Retire the unused decision model library"
status: approved
revision: 2
created: 2026-11-20
approval: {by: user, at: 2026-11-21, basis: "AskUserQuestion selection on revision 2: 승인 (Recommended)", digest: "sha256:272e376f401825a2a574d2662d911fb9594c2f174e3023a58c6afd51fb7e742d"}
---

# Retire the unused decision model library

## Intent

사용자 결정(2026-11-21 KST): 결정 모델(Jev·OpenAI)은 더 이상 쓰지 않는다. 인터뷰 질문은 SPEC-20261120-drop-decision-question 에서 지웠고, 남은 `src/decide/`(6,883줄)는 `src/lib.rs` 의 `pub mod decide;` 로만 연결돼 있고 모듈 밖에서 부르는 곳도 `moai decide` 명령도 없어요. 이 SPEC 은 그 라이브러리와 그것만 검사하던 시험 조각을 지우고, 결정 모델을 다루던 이전 SPEC 다섯 개를 `superseded` 로 표시하고, 결정 로그에 이번 두 사용자 결정을 남겨요.

## Requirements

- REQ-1: When the crate is built, the source tree shall hold neither a `src/decide/` folder nor a `decide` module declaration in `src/lib.rs` · failure: none — a deletion; the build and the tests are the check.
- REQ-2: When `tests/home_isolation_scan.rs` or the scan registry `tools/repo_scans.json` is read, neither shall name the decide library, its folder or its removed test while every other scan keeps running · failure: none — the scan test itself is the check.
- REQ-3: When a reader opens one of the five earlier decision model SPECs (`SPEC-20261118-project-decide-model`, `SPEC-20261118-decide-control`, `SPEC-20261118-decide-round-trip`, `SPEC-20261119-decide-transport-flaky`, `SPEC-20261119-outbound-screen-flake`), its status shall read `superseded` with an approval record, a line that names this SPEC and, where `moai spec check` asks for one, an `## Interview` record of this session's real question and answer · failure: none — `moai spec check` is the check.
- REQ-4: When a reader opens `docs/decisions-2026-11-08.md`, the log shall hold one row for the retired decision model and one row for keeping the `AGENTS.md` and `.gitignore` managed sections hand-maintained · failure: none — a documentation change.

## Acceptance

- AC-1 (REQ-1): `src/decide/` 없음, `src/lib.rs` 에 `mod decide` 없음, 라이브러리 시험이 통과. verify: `test ! -e src/decide`; `cargo test --locked --offline --lib`
- AC-2 (REQ-2): 홈 격리 스캔이 decide 조각 없이 통과하고, 스캔 등록부에 decide 항목이 없음. verify: `cargo test --locked --offline --test home_isolation_scan`; `grep -c -i decide tools/repo_scans.json`
- AC-3 (REQ-3): 다섯 SPEC 이 `superseded` 이고 검사 오류 없음. verify: `moai spec check --all`
- AC-4 (REQ-4): 결정 로그에 두 행. verify: `grep -c 'SPEC-20261120-retire-decide' docs/decisions-2026-11-08.md`
- AC-5: 가벼운 검사와 저장소 스캔. verify: `git diff --check`; `cargo fmt --check`; `cargo clippy --locked --offline --all-targets -- -D warnings`; `cargo test --locked --offline --test windows_unix_gate_scan --test doc_references --test git_env_scrub`

## Scope

- write: `src/decide/`(삭제), `src/lib.rs`, `tests/home_isolation_scan.rs`, `tools/repo_scans.json`, `docs/decisions-2026-11-08.md`, `.moai/specs/SPEC-20261118-project-decide-model/spec.md`, `.moai/specs/SPEC-20261118-decide-control/spec.md`, `.moai/specs/SPEC-20261118-decide-round-trip/spec.md`, `.moai/specs/SPEC-20261119-decide-transport-flaky/spec.md`, `.moai/specs/SPEC-20261119-outbound-screen-flake/spec.md`, `.moai/specs/SPEC-20261120-retire-decide/spec.md`
- read-only: `src/redact.rs` 와 키 보관소(`moai key-set`·`key-status`) — 다른 기능도 쓰므로 그대로; `skills/moai/references/internal/decide.md` — 질문 분류 규칙이라 결정 모델과 무관; 결정 모델을 경로로만 언급하는 다른 SPEC 들(기록이라 고치지 않음)
- forbidden: 실제 HOME·키 보관소 접근, 네트워크, 카드 번호를 SPEC id·브랜치·커밋에 사용

## Interview

- Q1: 아무 데서도 쓰지 않는 `src/decide/`(결정 모델 코드)를 지울까요? → A: 지우기 (Recommended) (사용자, 2026-11-21)
- Q2: `AGENTS.md`·`.gitignore` 관리 구역 충돌을 어떻게 정리할까요? → A: 지금 모양 유지로 확정 (Recommended) — 손으로 관리, 다음 생성은 skip-conflicts
- Q3: 이전 SPEC 다섯 개를 superseded 로 표시하는 범위로 승인할까요? → A: 승인 (Recommended)
- Q4: revision 2 — 범위에 `tools/repo_scans.json` 을 넣고, 2026-11-19 이후 SPEC 두 개에는 이번 문답을 Interview 로 적어 superseded 로 표시할까요? → A: 승인 (Recommended)

## Plan

1. `src/decide/` 와 `src/lib.rs:37` 삭제, `tests/home_isolation_scan.rs` 의 decide 전용 상수·스캔·시험 제거(다른 스캔은 유지) — 빌드와 AC-1·AC-2 확인.
2. 이전 SPEC 다섯 개를 `status: superseded` 로 바꾸고 승인 기록(이번 사용자 선택)과 "SPEC-20261120-retire-decide 가 대체함" 줄 추가.
3. 결정 로그에 두 행 추가, AC 검사 실행.
4. 키 보관소의 `typesafe`·`openai` 공급자 이름 정리는 후속 과제로 남김(다른 SPEC).
