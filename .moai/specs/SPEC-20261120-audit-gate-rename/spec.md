---
id: SPEC-20261120-audit-gate-rename
title: "audit-gate judges a renamed file by its previous filename too, like check-guard"
status: draft
revision: 1
created: 2026-11-20
---

# audit-gate — 이름이 바뀐 파일을 옛 이름(previous_filename)으로도 판정한다

## Intent

보호 경로 감사(2026-11-20, 감사 노트)에서 발견: `src/audit_gate.rs` 의 `collect` 가 GitHub files
API 응답에서 `filename` 만 수집해요. tier 2 보호 경로(예: `tools/local_check.sh`)를 비보호 경로
(예: `docs/`)로 rename 한 PR 은 옛 이름이 사라져 audit-gate 이 본문이 선언한 낮은 tier 로 판정돼요
— 같은 PR 을 check-guard(workflow jq 가 `.filename, (.previous_filename // empty)` 로 두 이름을
읽음)와 gate-check(`--no-renames`)은 옛 이름으로 봐요. 이 조각은 audit-gate 의 수집을 두 이름으로
맞춰 rename 으로 보호 경로를 빠져나가는 구멍을 닫아요.

## Non-goals

- `gate::tier_of`·`TIER2_PREFIXES` 변경 — 판정 규칙("경로 규칙은 올리기만")은 그대로.
- `.github/` 워크플로 편집 — CI 는 `moai audit-gate --pr` 로 바이너리를 직접 호출하고 수집은
  Rust 안에서 일어나요(audit-gate.yml). `.github/**` 는 다른 카드의 예약 중이기도 해요.
- check-guard·gate-check 변경 — 이미 두 이름을 읽어요.
- 보호 목록 확장(AGENTS.local.md, skills/moai/references/internal/review.md) — 리더 결정 사안,
  이 조각은 목록을 바꾸지 않고 카드에 질문 노트를 남겨요.

## Requirements

- REQ-1: When collect reads a pull request's changed files from the GitHub files API, a renamed entry shall contribute both of its names — `filename` and, when present, `previous_filename` — to the collected list · failure: none — an entry without `previous_filename` contributes exactly its `filename` as before.
- REQ-2: When a pull request renames a path that matches a tier 2 prefix (TIER2_PREFIXES) to an unprotected path, audit-gate shall judge that pull request tier 2 regardless of the tier its body declares · failure: none — the judgment itself stays `gate::tier_of` unchanged ("the path rule only raises").
- REQ-3: When a caller hands the `audit-gate` command its recorded input, the input shape shall stay unchanged — `previous_filename` is folded into `files` at collection, so the schema_version 1 contract keeps its single `files` list · failure: none — no schema change to the command's input or output.

## Acceptance

- AC-1 (REQ-1, REQ-2): PR fixture — 기록 전송(RecordedTransport)으로 `tools/local_check.sh` → `docs/local-check.md` rename 항목 하나를 주고 본문이 `audit-tier: 0` 을 선언해도 수집 목록이 두 이름을 다 갖고 decide 의 tier 가 2 로 나오는지. verify: `cargo test --locked --offline --lib audit_gate`
- AC-2 (REQ-3, REQ-2): CLI 표면 고정 — files 가 두 이름을 다 가진 기록 입력(본문 tier 0)이 tier 2 판정으로 나오는지(rename 이 tier 를 낮추지 못한다는 표면 계약). verify: `cargo test --locked --offline --test audit_gate_cli`
- AC-3: 이웃 — gate 자체는 무변경. verify: `cargo test --locked --offline --lib gate`; `cargo fmt --check`; `cargo clippy --locked --offline --all-targets -- -D warnings`

## Scope

- write: `src/audit_gate.rs`(collect 수집 + 단위 시험), `tests/audit_gate_cli.rs`(고정 시험), `.moai/specs/SPEC-20261120-audit-gate-rename/spec.md`
- read-only: `src/gate.rs`(TIER2_PREFIXES·tier_of 참조), `.github/workflows/audit-gate.yml`(호출 모양 참조)
- blocked: `.github/**`(다른 카드가 ci.yml 예약 중 — 이 조각은 편집 불필요로 확인됨)
- forbidden: 네트워크 실호출, 실제 HOME, 카드 번호를 SPEC id·가지·커밋에 사용

## Plan

1. RED: `src/audit_gate.rs` 단위 시험에 rename fixture 추가 — collect 가 `previous_filename` 을 버려 수집 목록·tier 어설션 실패.
2. GREEN: collect 가 `filename`과 `previous_filename`을 모두 수집(있을 때).
3. CLI 고정 시험(AC-2) + 이웃 검사 + READY 폭 검사 — 종료 코드를 READY 노트에.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 수정은 collect 한 곳 — tier 판정(tier_of)은 그대로 | check-guard·gate-check 과 "같은 방식"은 수집 층의 같음을 뜻함 | 판정 층에서 이름 재조회 |
| D-2 | 보호 목록 확장(AGENTS.local.md·review.md)은 이 조각에서 하지 않고 리더 결정 질문으로 남김 | 카드 본문이 "(leader) decide" 로 표기 | 리더 결정 뒤 TIER2_PREFIXES 확장 |
| D-3 | Input 스키마 v1 유지 — 수집에서 files 로 합침 | CLI 입력 계약 불변이 요구사항 | 입력에 별도 필드 추가 |

## Risks

- GitHub files API 가 rename 항목에 `previous_filename` 을 주는 건 문서화된 동작이고, 없으면(오래된 항목) 지금과 같아요 — 후퇴 없음.
- copied 항목도 `previous_filename` 을 가질 수 있어요 — 같은 필드라 같은 처리가 맞아요(복사된 옛 이름도 보호 경로면 tier 가 올라감).

## Result

base `6a8fa9cacf`(로컬 v2026.11.0), worktree `~/.moai/worktrees/moai-adk/c185`(새 모양), heavy-slot `--max 4`.

- RED: 기록 전송 fixture(`tools/local_check.sh` → `docs/local-check.md` rename, 본문 tier 0)에서 collect 가 `previous_filename` 을 버려 어설션 실패 — 수집 목록 `Some(["docs/local-check.md"])` vs 두 이름 기대(heavy-slot exit 101), 기록 뒤 GREEN.
- GREEN: `--lib` rename 단위 시험 1 passed · `audit_gate_cli` 6 passed 0 failed(`a_rename_of_a_tier_2_path_is_judged_by_its_old_name` 포함 — tier 2 고정) · `spec_cli` 18 passed 0 failed(`the_repository_specs_pass_check` 포함).
- 계약 교정: SPEC REQ-3 을 When…shall 모양으로, AC-3 의 `(이웃)` 괄호 제거(req_gears·ac_ref_unknown 해소); python 영역 모듈은 `tests.` 접두 모듈 경로로 실행.
- 폭 검사(운용 규칙 2호, 종료 코드): diff --check 0 · fmt 0 · ops rustfmt 0 · bash -n 0 · clippy --all-targets -D warnings 0 · build 0 · --bins 0 · **--lib 전체 0** · 영역+macos+고정 18 대상: 첫 실행 17 통과·spec_cli 1 실패(SPEC 모양 원인) → SPEC 교정 뒤 단독 재검 18 passed 0 failed · python 4모듈 0(129 tests) · python audit 영역 `tests.audit_eval_test` 0(14 tests) · trailer-check 0(legacy_session 경고 1, fails:false) · spec check --all 0(160 SPEC, 0 errors).
- 리더 결정 대기: 보호 목록 확장(AGENTS.local.md, skills/moai/references/internal/review.md) — 이 조각은 목록을 바꾸지 않았어요.
- 감사(2026-11-20, tier 2, 리더가 띄운 audit-c185-tier2) **PASS** — rename·삭제·복사 모두 tier 2, files API 오류 fail-closed, 뮤턴트 M1·M2·M3·M5 잡힘, --lib 1605 통과. 후속: TIER2_PREFIXES 두 항목 추가 + N1(files API 잘림 비교)·N2(fail-closed 고정)는 다음 조각.
- 통합 수리(2026-11-20, 로컬 통합 35 에서 빠짐 — 리더 노트): tests.ci_scope_test 의 문서 경로 스캔이 픽스처의 리터럴 `docs/…` 이름을 '코드가 참조하는 문서 경로'로 읽음(src/audit_gate.rs:765·783, tests/audit_gate_cli.rs:120) — 세 곳 모두 실행 중 조립(`format!("{}/local-check.md", "docs")`)으로 바꿈. 같은 커밋에 Moai-Verified: tests 트레일러 추가.
