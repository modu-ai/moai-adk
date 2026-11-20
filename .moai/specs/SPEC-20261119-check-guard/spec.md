---
id: SPEC-20261119-check-guard
title: "check-guard: a PR cannot weaken the checks that judge it (list read from base)"
status: draft
revision: 2
created: 2026-11-19
---

# check-guard: 보호 검사 경로를 바꾸는 PR 이 심판 규칙을 약화시키지 못하게 한다

## Intent

오늘 PR 은 자기를 심판하는 규칙을 PR head 에서 바꿀 수 있어요: `REVIEW.md` 와 `agents/audit.md` 는
`gate::tier0_safe`(src/gate.rs:267-275) 덕분에 `audit-tier: 0` 선언만으로 통과할 수 있고,
`tools/ci_scope.py`·`src/audit_gate.rs` 는 tier 1 이면 돼요. 카드 C-97 (리더 결정 노트 2336,
2026-11-18T23:49:39Z)은 보호 검사 경로를 담는 **목록을 base 가지에** 두고, 그 목록을 base 에서
읽는 검사 `check-guard` 로 이 경로를 바꾸는 PR 에 'self-modifying check' 라벨을 붙이고
audit-tier 2 를 강제해요. 이미 `audit-gate.yml` 이 같은 모양의 선례예요: 게이트는 base 커밋에서
빌드되어 PR 이 심판 규칙을 바꿀 수 없어요. 이 카드는 ruleset 26791615 와 필수 검사 이름을
바꾸지 않아요(카드 기준 3, guarded — 필요화는 별도 사용자 결정, 아래 Decisions D3).

강제 방법(설정 D1): `check-guard` 는 보호 경로가 변경된 PR 이 본문 첫 10 줄에서 정확히
`audit-tier: 2` 를 선언하지 않으면 실패해요. `audit-gate`(src/audit_gate.rs:144-148)는 선언된
tier 로 `gate::tier_of` 를 돌리므로, 선언된 2 는 곧 게이트의 실제 tier-2 요구(비저자 PASS)로
이어져요. 선언 없음은 게이트가 이미 tier 2 로 계산하지만(audit_gate.rs:148), 이 검사는 명시적
선언을 요구해 저자가 강제를 인지하게 해요.

## Non-goals

- ruleset 26791615, 필수 검사 이름/구성, 브랜치 보호 변경 (guarded, 카드 밖).
- `.github/moai/labels.yml` 편집 — 라벨은 가드가 런타임에 `gh label create --force` 로 시도하고
  결과를 요약에 기록해요 (설정 D2). 매니페스트 등재는 별도 사용자 절차(00-E #2020).
- `src/gate.rs` 의 `TIER2_PREFIXES` 변경 — src/** 는 이 카드 범위 밖이에요. 게이트의 경로 규칙은
  지식으로 남고, 이 카드는 base 목록 쪽을 채워요.
- `pull_request_target` 사용, 어떤 secret 도 읽기, PR head 의 스크립트 실행.
- 기존 `ci`·`audit-gate` 워크플로의 동작 변경.

## Requirements

- REQ-1: When a `pull_request` event starts on `v2026.11.0`, `.github/workflows/check-guard.yml` shall check out `github.event.pull_request.base.sha` with the pinned checkout action and execute the checked-out copy of `tools/check_guard.py` from that base checkout.
- REQ-13 (REQ-1): IF the base checkout or the guarded script step fails, THEN the `check-guard` job shall end red and execute no step from the pull request head.
- REQ-2: When the `check-guard` job runs, the job shall read the pull request's changed files from the GitHub files API over all pages with both `filename` and `previous_filename`, and the pull request body from `gh pr view --json body`.
- REQ-3: When `tools/check_guard.py` reads the protected list, the guard shall take the entries from the base checkout's `tools/protected_checks.json` and match an entry containing `*` as a whole-path glob, an entry ending with `/**` as a path prefix, and every other entry as an exact path · failure: none — the three shapes are total over any entry.
- REQ-4: When a changed path matches a protected entry and the body declares exactly `audit-tier: 2` in the `gate::parse_body_tier` grammar, `tools/check_guard.py` shall exit 0, attempt the `self-modifying check` label, and list the protected paths and the label outcome in its summary.
- REQ-5 (REQ-4): IF the declaration is absent or reads 0 or 1, THEN `tools/check_guard.py` shall exit 1 with a `declare audit-tier: 2` reason in its summary and still attempt the label.
- REQ-6 (REQ-4): IF no changed path matches a protected entry, THEN `tools/check_guard.py` shall exit 0 and record that reason in its summary.
- REQ-7 (REQ-2): IF the list file is missing or its JSON is broken, or the file-list input is empty, unreadable or carries a control character, or the body is unreadable, THEN `tools/check_guard.py` shall exit 1 and name that reason in its summary.
- REQ-8: IF an option is missing or malformed, THEN `tools/check_guard.py` shall refuse with usage and exit 2.
- REQ-9 (REQ-4): IF a label create or add call fails, THEN `tools/check_guard.py` shall keep the verdict unchanged and record the label outcome in its summary.
- REQ-10: When a pull request is audited, `agents/audit.md` shall direct the auditor to read `REVIEW.md` and `agents/audit.md` from the pull request's base ref with `git show <base-sha>:<path>` · failure: none — an instruction document names no runtime failure branch.
- REQ-11: IF `.github/workflows/check-guard.yml` deviates from the REQ-1 and REQ-2 structure contract — the pinned checkout action, the `pull_request` trigger only, the base.sha ref, the minimal permissions — or `agents/audit.md` loses the REQ-10 instruction, THEN `tests/github_assets_test.py` shall fail.
- REQ-12: When `tests/check_guard_test.py` runs, the suite shall cover REQ-3 to REQ-9 with fixture file lists — protected and unprotected changes, a pull request editing the protected list itself, rename old and new names, deletions, the safe failures, glob matching, and a label failure that cannot flip the verdict — without network or `gh` · failure: none — the coverage itself is the deliverable and a missing case surfaces as an audit finding.

## Acceptance

- AC-1 (REQ-3, REQ-4): 보호 경로 + `audit-tier: 2` 선언 → exit 0, 라벨 시도, 요약에 경로 나열.
  verify: `PYTHONDONTWRITEBYTECODE=1 python3 -m unittest tests.check_guard_test -k Pass`
- AC-2 (REQ-5): 보호 경로 + 선언 없음/0/1 → exit 1, 이유 `declare audit-tier: 2`.
  verify: `PYTHONDONTWRITEBYTECODE=1 python3 -m unittest tests.check_guard_test -k Force`
- AC-3 (REQ-3): glob `src/gate*.rs` 는 `src/gate.rs`·`src/gate_collect.rs` 를 잡고 `src/audit_gate.rs` 는
  정확 항목으로, `.github/workflows/**` 는 그 폴더만 잠가요.
  verify: `PYTHONDONTWRITEBYTECODE=1 python3 -m unittest tests.check_guard_test -k Match`
- AC-4 (REQ-5, REQ-12): PR 이 `tools/protected_checks.json`·`REVIEW.md` 를 약화해 편집해도 가드는
  base 목록으로 판정해 같은 경로를 또 잠가요(자기 수정이 판정을 못 바꿔요).
  verify: `PYTHONDONTWRITEBYTECODE=1 python3 -m unittest tests.check_guard_test -k SelfEdit`
- AC-5 (REQ-12): 이름변경(보호→보호, 보호→비보호 old 이름)과 삭제가 잡혀요.
  verify: `PYTHONDONTWRITEBYTECODE=1 python3 -m unittest tests.check_guard_test -k Rename`
- AC-6 (REQ-6): 비보호(docs) 변경만 있으면 exit 0.
  verify: `PYTHONDONTWRITEBYTECODE=1 python3 -m unittest tests.check_guard_test -k Unprotected`
- AC-7 (REQ-7, REQ-8, REQ-9): 목록 파손·빈 파일 목록·제어 문자·본문 누락은 exit 1, 옵션 오류는 usage exit 2,
  라벨 실패는 판정 불변.
  verify: `PYTHONDONTWRITEBYTECODE=1 python3 -m unittest tests.check_guard_test -k Safe`
- AC-8 (REQ-1, REQ-2, REQ-10, REQ-11): check-guard.yml 은 base.sha 체크아웃 + base 사본 실행 + pull_request 전용 +
  pinned 액션 + 최소 권한을 유지하고, agents/audit.md 는 base ref 읽기 지시를 담아요.
  verify: `PYTHONDONTWRITEBYTECODE=1 python3 -m unittest tests.github_assets_test -k Guard`; `PYTHONDONTWRITEBYTECODE=1 python3 -m unittest discover -s tests -p 'github_assets_test.py'`

## Scope

- write: `.github/workflows/check-guard.yml`, `tools/check_guard.py`, `tools/protected_checks.json`,
  `tests/check_guard_test.py`, `tests/github_assets_test.py`, `agents/audit.md`,
  `.moai/specs/SPEC-20261119-check-guard/**`
- read-only: `src/gate.rs`(TIER2_PREFIXES·tier0_safe 참조), `src/audit_gate.rs`, `.github/workflows/ci.yml`,
  `.github/workflows/audit-gate.yml`, `.github/moai/labels.yml`
- forbidden: ruleset 26791615 와 필수 검사 이름, `.github/moai/labels.yml` 편집, `src/**` 편집,
  실제 HOME·원장·키, 네트워크 쓰기, `pull_request_target`

## Plan

1. RED: `tests/check_guard_test.py` — fake 실행기 주입으로 REQ-3~9·REQ-12 시험을 먼저 쓰고 실패 확인.
2. GREEN: `tools/check_guard.py` + `tools/protected_checks.json` 구현.
3. 워크플로: `.github/workflows/check-guard.yml` — audit-gate.yml 모양(base 체크아웃, pinned 액션,
   최소 권한)으로 gh API 결과를 임시 파일로 넘겨요.
4. REQ-10: `agents/audit.md` 에 base ref 읽기 문단 추가; `tests/github_assets_test.py` 에 구조 계약 시험(REQ-11).
5. 이웃 검사: `python3 -m unittest discover -s tests`, `cargo fmt --check`, `cargo clippy --all-targets -D warnings`,
   `moai trailer-check origin/v2026.11.0..HEAD`.

## Decisions

- D1 (되돌릴 수 있음, 범주 none): 강제 방법 — "본문 `audit-tier: 2` 선언 요구 + 미선언 시 검사 실패".
  대안: gate.rs TIER2_PREFIXES 확장(src/** 범위 밖이라 불가), 라벨을 게이트가 읽게 함(src/audit_gate.rs
  범위 밖). 선언된 2 는 audit_gate.rs:144-148 에서 그대로 tier 2 가 되므로 강제가 게이트까지 이어져요.
- D2 (되돌릴 수 있음, 범주 none): 라벨 런타임 생성 fallback — `gh label create --force` 후
  PR 부착, 실패는 요약 기록만. `.github/moai/labels.yml` 은 범위 밖.
- D3 (사용자 결정 항목): `check-guard` 를 ruleset 의 필수 검사로 올리는 일은 이 카드 밖(guarded) —
  SPEC 은 검사를 준비만 해요.

## Risks

- 라벨 생성 권한이 GITHUB_TOKEN 에 없으면 라벨이 계속 안 붙어요 → 요약에 매번 기록되고, 강제(판정)는
  라벨과 무관하게 작동해요. 필요하면 별도 결정으로 매니페스트 등재.
- files API 페이지가 30 넘는 PR(>3000 파일)은 가드가 안전 실패로 멈춰요 → 대형 PR 은 실패 이유가
  요약에 보여요(거짓 통과보다 안전).
- `previous_filename` 이 없는 오래된 이벤트 형식은 jq 단계에서 빈 값으로 흘러요 → 빈 줄은 가드가 무시해요.
