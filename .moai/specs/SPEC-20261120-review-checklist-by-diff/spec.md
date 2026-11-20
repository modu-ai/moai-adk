---
id: SPEC-20261120-review-checklist-by-diff
title: "review-package hands the reviewer only the checklist items the diff triggers (list read from base)"
status: draft
revision: 1
created: 2026-11-20
---

# review-package: diff 로 고르는 점검표

## Intent

감사자에게 가는 점검 규칙이 두 군데에 평평하게 쌓여 있어요. `REVIEW.md` 는 14개 항목이 PR 마다 통째로 가고, `skills/moai/workflows/verify.md` 의 "Change-specific rules"(5개)와 "Shell and helper rules"(4개)는 26,334 바이트 문서 안에 산문으로 있어서 항상 통째로 읽혀요. 둘 다 "이 diff 에 맞는 것만" 고르지 못하고, 항목마다 왜 생겼는지(Issue·PR·SPEC)를 기계가 읽을 곳이 없어요. 참고한 글(claude.dev "How the Claude Code team thinks about verification", 2d 항목)의 모양은 항목마다 diff 에서 읽히는 발동 조건·점검 내용·근거를 적고, 리뷰어는 조건이 맞는 몇 개만 보는 거예요.

이 SPEC 은 `tools/review_checklist.json` 을 새로 두고, `moai review-package` 가 **기준(base) 커밋의** 이 파일을 읽어 diff 와 맞는 항목만 `checklist` 키로 내게 해요. 기준 커밋에서만 읽으므로 PR 이 이 파일을 고쳐도 그 PR 의 점검표는 기준 커밋 것이에요(`REVIEW.md` 와 같은 원칙, `src/review.rs` 머리말). 경로 맞춤은 새로 만들지 않고 `tools/repo_scans.json` 이 쓰는 `repo_scans::in_scope`(prefix + 선택 suffix)를 그대로 써요.

옮기는 규칙(재조사로 다시 센 것): `verify.md` 는 두 절에 항목 9개가 있고, 그중 WF-OBS 번호가 붙은 것이 8개(#9 #25 #26 #29 #30, #17 #19 #16)예요. 아홉째 "Host worktree guard"(Issue #2836)는 번호가 없고 같은 절에 있어서 함께 옮겨요. `REVIEW.md` 에서는 경로가 정해진 5개(`tests/` 의 git 도우미와 바이너리 복사, `docs/` 를 읽는 시험, `skills/moai/**` 매니페스트, 골든 파일, `Cargo.lock`)를 옮겨요. 나머지 9개는 경로로 읽히지 않는 일반 규칙이라 `REVIEW.md` 에 남아요. 항목 중 "항상 맞는" 것(경로 prefix 가 빈 문자열)은 감사자의 작업 방식에 대한 규칙(WF-OBS #9, #16, #17, #19, Issue #2836)이고, 나머지는 diff 로 발동해요.

## Non-goals

- `src/gate.rs`, `tools/protected_checks.json` 편집. 보호 목록에 이 파일을 넣는 일은 그 파일들을 쓰는 다른 작업이 해요. 이 SPEC 은 항목이 들어 있는지만 시험해요(REQ-9).
- `skills/moai/workflows/verify.md`·`skills/moai/workflows/audit.md`·`skills/moai/references/source-manifest.json` 편집. 다른 작업이 이 경로를 쥐고 있어서 이번 조각에서 쓰지 않아요. verify.md 에서 규칙을 빼는 일과 audit.md 의 Review steps 1 에 `checklist` 를 한 줄 적는 일은 후속이에요(Risks 참조).
- `AGENTS.local.md` 의 원인이 적힌 규칙 옮기기. 후속이에요.
- 모델 호출·네트워크. 모든 판정은 문자열과 경로 비교예요.
- `diff_pattern` 을 정규식으로 만들기. 글자 그대로(literal) 비교만 해요(새 크레이트 없음).

## Requirements

- REQ-1: When `moai review-package` builds a package, `src/review.rs` shall read `tools/review_checklist.json` from the BASE commit only and carry in a `checklist` key, in the table's order, each item whose trigger the diff meets, with its `id`, `check`, `source`, `added` and `matched` (the changed paths that met the trigger and the first `diff_pattern` string found in them, or null).
- REQ-2 (REQ-1): IF the base commit has no `tools/review_checklist.json`, THEN `src/review.rs` shall return an empty `checklist` and add no `skipped` entry.
- REQ-3: When an item is judged, `src/review.rs` shall fire it only if at least one changed path is in the item's `trigger.paths` by `repo_scans::in_scope` and, when the item has `diff_pattern`, an added line of that same path contains one of its strings literally.
- REQ-4 (REQ-1): IF `tools/review_checklist.json` at base is not valid, THEN `src/review.rs` shall return an empty `checklist` and add one `skipped` entry for the file whose reason names the first problem, where valid means: JSON, `schema_version` 1, a list `items`, per item exactly the keys `id`, `trigger`, `check`, `source`, `added` (`trigger` holds `paths` and optionally `diff_pattern`), a unique `id` of lowercase letters, digits and hyphens, a non-empty `trigger.paths` whose entries have a string `prefix` and optional string `suffixes`, a non-empty list of non-empty strings for `diff_pattern` when present, a non-empty `check`, a `source` naming an `Issue #<n>`, a `PR #<n>` or a `SPEC-<8 digits>-<slug>`, an `added` date as `YYYY-MM-DD`, and a file of at most 150,000 bytes.
- REQ-5 (REQ-3): IF the diff cannot be split into exactly one section per changed path (a type change shows two sections), THEN `src/review.rs` shall judge only the items without `diff_pattern` and add one `skipped` entry for the file with the reason `diff-sections-mismatch`.
- REQ-6: When `tests/review_checklist_cli.rs` runs, the suite shall build a package from a temporary repository whose base holds this repository's own `tools/review_checklist.json`, and fail when that file is invalid or an item lacks `source` or `trigger` · failure: none — the file is read, never written.
- REQ-7: When `tests/review_checklist_cli.rs` runs, the suite shall fail when the checklist lacks an item for any of the nine rules of the `Change-specific rules` and `Shell and helper rules` sections of `skills/moai/workflows/verify.md`, or for the five path-fixed rules named in the Intent, or when an item's `check` does not carry that rule's text word for word while the rule is still in its source file · failure: none — the equivalence is the deliverable.
- REQ-8: When `tests/review_checklist_cli.rs` runs, the suite shall fail for any item of the real file that has no case pair — a diff that fires it and a diff that does not — and for any case whose firing diff does not fire exactly that item or whose quiet diff fires it · failure: none — a missing case surfaces as a failing test.
- REQ-9: When `tests/review_checklist_cli.rs` runs, the suite shall fail unless `tools/protected_checks.json` lists `tools/review_checklist.json` · failure: none — the list is read, never written, and the entry itself comes from the task that owns the list.
- REQ-10: When this SPEC is closed, its Result section shall list, for the last 30 merged pull requests, how many checklist items each would have fired with the real file, as a distribution, and the command that produced it · failure: none — the measurement is reported, not enforced.
- REQ-11: When `REVIEW.md` is read by `moai review-package`, the file shall carry one sentence saying that path-specific items arrive in the `checklist` key and none of the five moved rules · failure: none — the file stays within the byte and link limits, which `tests/review_md_repo.rs` already checks.
- REQ-12 (REQ-1): IF the pull request head holds its own `tools/review_checklist.json`, THEN `src/review.rs` shall ignore that copy and return the items of the base commit's file.

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-12): 기준 커밋의 파일에서 diff 와 맞는 항목만 table 순서로 `checklist` 에 담기고, 기준에 파일이 없으면 빈 목록이며 PR 쪽 사본은 무시돼요.
  verify: `cargo test --offline --locked --test review_checklist_cli -- base_only`
- AC-2 (REQ-3): 경로만 맞는 항목, 경로와 `diff_pattern` 이 함께 맞아야 하는 항목, 삭제된 파일, 다른 경로의 같은 글자가 각각 기대대로 발동하거나 안 해요.
  verify: `cargo test --offline --locked --test review_checklist_cli -- trigger`
- AC-3 (REQ-4, REQ-5): 잘못된 파일(깨진 JSON, 빠진 키, 모르는 키, 겹친 id, 근거 없는 source, 잘못된 날짜, 150,000 바이트 경계 N/N+1)과 타입 변경 diff 는 각각 `skipped` 이유를 남겨요.
  verify: `cargo test --offline --locked --test review_checklist_cli -- invalid`; `cargo test --offline --locked --lib review::tests`
- AC-4 (REQ-6): 이 저장소의 실제 파일이 유효하고 모든 항목에 `source`·`trigger` 가 있어요.
  verify: `cargo test --offline --locked --test review_checklist_cli -- real_file`
- AC-5 (REQ-7): verify.md 9개 규칙과 REVIEW.md 5개 규칙이 하나도 빠지지 않고 항목에 있어요.
  verify: `cargo test --offline --locked --test review_checklist_cli -- moved_rules`
- AC-6 (REQ-8): 모든 항목에 발동 diff 와 비발동 diff 시험이 있어요.
  verify: `cargo test --offline --locked --test review_checklist_cli -- every_item_has_cases`
- AC-7 (REQ-9): 보호 목록에 이 파일이 들어 있어요.
  verify: `cargo test --offline --locked --test review_checklist_cli -- protected_list`
- AC-8 (REQ-11): REVIEW.md 는 한도 안에서 통째로 읽혀요.
  verify: `cargo test --offline --locked --test review_md_repo`
- AC-9 (REQ-10): 최근 병합된 PR 30개의 발동 수 분포가 Result 에 있어요.
  verify: Result 절의 명령을 다시 실행해 같은 분포가 나오는지 봐요.

## Scope

- write: `tools/review_checklist.json`, `src/review.rs`, `REVIEW.md`, `tests/review_checklist_cli.rs`, `tests/review_package_cli.rs`, `tests/review_md_repo.rs`, `.moai/specs/SPEC-20261120-review-checklist-by-diff/**`; `src/repo_scans.rs` 는 `in_scope` 를 다시 쓰는 데 필요할 때만(지금은 이미 `pub`)
- read-only: `skills/moai/workflows/verify.md`(대조 시험이 읽어요), `tools/protected_checks.json`, `src/ops/review_package.rs`, `src/gate.rs`
- forbidden: `src/gate.rs`, `tools/protected_checks.json`, `skills/moai/workflows/verify.md`, `skills/moai/workflows/audit.md`, `skills/moai/references/source-manifest.json`, 실제 HOME, 네트워크 쓰기

## Plan

1. RED: `tests/review_checklist_cli.rs` — 기준에서 읽기·PR 사본 무시·발동·비발동·유효성·150,000 경계·타입 변경 시험을 먼저 쓰고 실패를 확인해요.
2. GREEN: `src/review.rs` 에 `Item`·`checklist` 파싱, `in_scope` 재사용, diff 구간 나누기, `Package.checklist` 를 더해요.
3. 점검표 파일: `tools/review_checklist.json` 에 verify.md 9개와 REVIEW.md 5개를 글자 그대로 옮기고 항목마다 `source`·`added` 를 git 이력(`git log -S`)에서 읽은 값으로 적어요.
4. 대조·사례 시험(REQ-6~9)과 `REVIEW.md` 줄이기(REQ-11).
5. 이웃 검사와 `tools/self_check.sh --pre-ready`. 마지막에 최근 병합 PR 30개 분포를 재서 Result 에 적어요.

## Decisions

- D1 (되돌릴 수 있음, 범주 none): 카드 본문은 "변경 종류별 규칙 8개"라고 쓰지만 `verify.md` 의 두 절에는 항목이 9개예요(WF-OBS 번호 8개 + 번호 없는 Host worktree guard). 아홉 개를 모두 옮겨요. 대안: 번호 붙은 8개만 옮기고 아홉째는 verify.md 에 남기기(한 절에 항목 하나만 남아 어색). 되돌리기: 그 항목을 JSON 에서 빼고 verify.md 에 두면 돼요.
- D2 (되돌릴 수 있음, 범주 none): 작업 방식 규칙 5개(WF-OBS #9, #16, #17, #19, Issue #2836)는 diff 로 읽을 수 없으므로 빈 prefix 로 모든 PR 에 발동해요. 대안: verify.md 에 남기기(그러면 "하나도 빠지지 않게 옮김"이 어긋나요). 되돌리기: 항목을 빼면 돼요.
- D3 (되돌릴 수 있음, 범주 none): `source` 는 `Issue #<n>`·`PR #<n>`·`SPEC-<id>` 만 받아요. 카드 번호(`C-<n>`)는 2026-11-19 22:15Z 의 결정으로 저장소 산출물에 쓰지 않으므로 받지 않아요(카드 본문은 "Issue·카드·SPEC·PR" 라고 적었어요). 되돌리기: 허용 모양 한 줄 추가.
- D4 (되돌릴 수 있음, 범주 none): 파일이 없을 때는 `skipped` 항목 없이 빈 목록(`REVIEW.md` 없음과 같은 처리, 기존 `skipped` 단정을 깨지 않아요). 파일이 있는데 잘못이면 `skipped` 로 크게 알려요.
- D5 (되돌릴 수 있음, 범주 none): `diff_pattern` 은 추가된 줄의 literal 부분 문자열(여럿 중 하나). 정규식은 새 크레이트가 필요하고 결정성 검증이 늘어서 쓰지 않아요.

## Risks

- 보호 목록(`tools/protected_checks.json`·`src/gate.rs` 의 `TIER2_PREFIXES`)에 이 파일을 넣는 편집은 다른 작업이 쥐고 있어요. 그 작업이 병합되기 전에는 REQ-9 시험이 빨갛게 남아요. 순서: 그쪽을 먼저 병합하고 이 가지가 그 결과를 합친 뒤 통과시켜요.
- verify.md 에서 규칙을 빼는 일, audit.md 의 Review steps 1 에 `checklist` 를 한 줄 적는 일, `skills/moai/references/source-manifest.json` 의 sha256 갱신은 다른 작업이 그 경로를 놓을 때까지 미뤄져요. 그때까지 규칙은 verify.md 와 점검표에 모두 있고(중복, 누락 아님), REQ-7 시험은 verify.md 의 글자와 항목 `check` 의 일치를 봐요.
- `diff_pattern` 의 literal 일치는 가설이에요: 맞는 줄이 없어도 발동할 수 있고(주석), 틀려도 놓칠 수 있어요(다른 표기). 항목은 감사자가 눈으로 확인할 후보일 뿐 판정이 아니에요.
- 같은 diff 에서 구간 수와 변경 경로 수가 다르면(타입 변경) `diff_pattern` 항목은 판정하지 않고 `skipped` 에 이유를 남겨요. 판정 못 한 항목은 발동하지 않는 것으로 취급돼요.

## Result

(닫을 때 채워요.)
