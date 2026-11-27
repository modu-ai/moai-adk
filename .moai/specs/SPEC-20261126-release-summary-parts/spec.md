---
id: SPEC-20261126-release-summary-parts
title: "릴리스 감사 요약을 여러 부분 댓글로 올리고 audit gate 가 부분 사슬 전체를 확인해요"
status: draft
revision: 1
created: 2026-11-26
---

# 릴리스 감사 요약을 여러 부분 댓글로 올리고 audit gate 가 부분 사슬 전체를 확인해요

## Intent

`release-audit-summary` 는 `base..head` 의 first-parent 커밋마다 한 줄을 적은 댓글 하나를 만들고, 60000바이트를 넘으면 거절해요(`src/release_audit_summary.rs:39` 의 `MAX_COMMENT_BYTES`). 계획 단계에서 첫 릴리스 범위를 미리 본 요약 글은 63067바이트였고, 판정 줄이 늘면 약 77,500바이트로 추정해요(추정이며 확인하지 않았어요). 그래서 지금은 범위 전체의 요약을 올릴 수 없어요.

이 SPEC 은 한 댓글에 들어가지 않는 요약을 `base..head` 의 이어진 구간마다 부분 댓글 하나로 나눠 올리고, `audit-gate` 가 묶음 전체의 사슬을 확인하게 해요. 줄 문법과 줄마다의 PASS 규칙은 그대로이고, 어떤 글도 자르지 않아요. 한 댓글에 들어가는 요약은 지금과 똑같아요. SPEC-20261125-release-audit-summary 의 REQ-10(한도를 넘으면 거절)은 이 SPEC 에서 "나누고, 나눌 수 없을 때만 거절"로 바뀌어요.

의존(이 SPEC 이 풀지 않아요): `audit-gate` 는 PR base 커밋에서 빌드돼요(`.github/workflows/audit-gate.yml:49`). 계획 때 읽은 첫 릴리스의 base(`6a8fa9cacf`)에는 요약을 읽는 코드가 없어요. SPEC-20261125-release-audit-summary 의 D-6 은 2026-11-26 에 사용자가 직접 "새 관문을 먼저 올리기"로 정했어요: 요약을 읽는 새 관문만 담은 작은 PR 을 옛 관문으로 먼저 `v2026.11.0` 에 올리고, 그다음 본 릴리스를 새 관문으로 판정해요. ruleset 은 바꾸지 않아요. 첫 릴리스의 요약이 부분으로 나뉘면 이 SPEC 의 게이트 변경도 그 작은 PR 에 들어가야 효과가 있어요(추론이며 사용자가 정한 내용이 아니에요). PR 을 여는 일과 푸시는 사용자의 릴리스 선택 뒤에만 해요.

## Interview

- Q1: 한도를 넘는 요약을 어떻게 올릴까요? 선택지: (가) 여러 부분 댓글과 게이트의 묶음 확인 (나) 개수와 PASS 아닌 줄만 담은 요약 하나. → A: 사용자에게는 아직 묻지 않았어요. 리더(사용자가 아니에요)는 (가)를 지지하고 (나)는 게이트를 약하게 해서 반대해요. 기본: (가). [NEEDS CLARIFICATION: (가)로 진행해도 되는지]
- Q2: 잘못 올라간 부분 댓글은 어떻게 고칠까요? → A: 아직 묻지 않았어요. 기본: 명령은 댓글을 고치거나 지우지 않고, 사람이 사용자의 선택을 받아 GitHub 에서 지운 뒤 다시 올려요. [NEEDS CLARIFICATION: 이 원격 삭제 절차를 받아들일지, 누가 지울지]
- Q3: 승인 한 번이 댓글 여러 개와 부분 실패 뒤의 재실행을 덮을까요? → A: 아직 묻지 않았어요. 기본: 같은 plan digest 와 같은 permission reference 로 남은 부분만 다시 올려요. [NEEDS CLARIFICATION: 재실행에 새 승인이 필요한지]
- Q4: 처음 싣는 릴리스의 게이트 base 는 어떻게 할까요? → A: 2026-11-26 사용자 직접 선택 "새 관문을 먼저 올리기". 새 관문만 담은 작은 PR 을 옛 관문으로 먼저 올리고 본 릴리스를 새 관문으로 판정해요. ruleset 은 바꾸지 않아요.

## Non-goals

- 워크플로와 게이트를 빌드하는 base 의 변경(이전 SPEC 의 D-6).
- 개수만 담은 요약, PASS 줄을 뺀 요약 같은 줄이는 형식.
- 판정 없는 착지의 소급 판정, 줄 문법(`PASS`/`ABS`/`NONE`/`FAIL`)과 `Verdict::counts` 의 변경.
- 댓글을 고치거나 지우는 명령, 새 판이 옛 판을 대신하는 규칙(withdraw 줄 포함).
- `src/forge/connector.rs`, `Input` 필드, 원장 스키마, `AUDIT #` 댓글 경로의 변경.
- 서명 태그, 커밋마다 GitHub API 로 완전성을 증명하는 일.

## Requirements

- REQ-1: While the single summary comment of a range is at most 60000 bytes, the command shall plan that one comment with the same bytes, marker and plan digest as before this SPEC.
- REQ-2: When the single summary comment of a range would exceed 60000 bytes, the command shall plan 2 to 99 part comments that together hold every landing line exactly once in first-parent order under one set id and one plan digest.
- REQ-3: The command shall keep every part comment within 60000 bytes with CRLF line ends counted, without truncating a comment, dropping a landing line or separating an `ABS` line from its target.
- REQ-4 (REQ-2): IF no split at the allowed boundaries keeps every part within the cap and the set within 99 parts, THEN the command shall refuse the plan and post nothing.
- REQ-5: The first line of a part comment shall start with `RELEASE-AUDIT-SUMMARY ` and read neither as an `AUDIT #` verdict nor as a control line.
- REQ-6: When `--post` is given for a split plan with every required value and the rebuilt plan digest equals the given one, the command shall post in part order each part whose marker is not yet on the pull request and no other comment.
- REQ-7 (REQ-6): IF a required value is missing or the rebuilt plan digest differs or an incomplete summary lacks `--allow-incomplete`, THEN the command shall refuse before it sends any part.
- REQ-8 (REQ-6): IF a part cannot be posted, THEN the command shall stop and end with an error that names the parts posted, already present and not sent.
- REQ-9: The command shall never edit or delete a comment of a pull request.
- REQ-10: When every part of one set that names the PR head is present, the audit gate shall count the set as passing only if its parts form one unbroken first-parent chain from the merge base to the PR head with no gap, overlap or repeated commit, its recomputed set id equals the named one and every line passes the per-line rule of a single summary.
- REQ-11 (REQ-10): IF two present parts of a set disagree on merge base or part count, hold different content for one part index, break a segment boundary or repeat a commit, or the recomputed id of a complete set differs, THEN the audit gate shall decide `fail` at every tier.
- REQ-12 (REQ-10): IF a member's summary or part comment does not read within the cap or a readable one for the PR head holds a `FAIL` line, THEN the audit gate shall decide `fail` at every tier.
- REQ-13 (REQ-10): IF the base a complete set names differs from the merge base of the PR or one of its lines has no counted verdict, THEN the audit gate shall decide `missing` at tier 1 and 2.
- REQ-14: While a set that names the PR head lacks a part and no `fail` rule applies, the audit gate shall decide `missing` at every tier and name the absent parts.
- REQ-15: While two member comments carry the same part with identical parsed content, the audit gate shall read them as one part.
- REQ-16: While more than one summary or set names the PR head, the audit gate shall decide `pass` only if each of them passes on its own.
- REQ-17: While no member comment is a part comment, the audit gate shall decide exactly as before this SPEC.

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-3, REQ-4, REQ-5): 한도 안의 범위는 지금과 같은 바이트의 댓글 하나이고, 넘는 범위는 모든 줄을 한 번씩 담은 부분들이며, `ABS` 묶음은 갈라지지 않고, 나눌 수 없으면 거절해요. 부분의 첫 줄은 판정 줄이나 제어 줄로 읽히지 않고 마커는 200바이트 안이에요 · verify: `cargo test --locked --offline --lib release_audit_summary::tests::a_range_within_the_cap_keeps_the_single_comment_bytes`
  verify: `cargo test --locked --offline --lib release_audit_summary::tests::a_range_over_the_cap_splits_into_parts_that_hold_every_line_once`
  verify: `cargo test --locked --offline --lib release_audit_summary::tests::a_split_never_separates_an_absorbed_line_from_its_target`
  verify: `cargo test --locked --offline --lib release_audit_summary::tests::part_render_and_parse_round_trip`
  verify: `cargo test --locked --offline --lib release_audit_summary::tests::the_first_line_of_a_part_never_starts_like_an_audit_verdict`
  verify: `cargo test --locked --offline --lib release_audit_summary::tests::a_part_marker_fits_the_connector_limit`
  verify: `cargo test --locked --offline --test card release_audit_summary_cli::a_range_over_the_cap_is_planned_as_parts`
  verify: `cargo test --locked --offline --test card release_audit_summary_cli::a_summary_over_the_cap_is_refused_not_cut`
- AC-2 (REQ-6, REQ-7, REQ-8, REQ-9): 승인 한 번이 부분을 순서대로 한 번씩 올리고, 거절은 첫 전송 전에 나고, 중간 실패 뒤 같은 명령은 빠진 부분만 올리며, 원장이 바뀌면 아무것도 올리지 않고, `PATCH` 와 `DELETE` 호출이 없어요 · verify: `cargo test --locked --offline --test card release_audit_summary_cli::one_approval_posts_every_part_once_in_order`
  verify: `cargo test --locked --offline --test card release_audit_summary_cli::posting_parts_refuses_before_the_first_send`
  verify: `cargo test --locked --offline --test card release_audit_summary_cli::a_retry_after_a_partial_post_sends_only_the_missing_parts`
  verify: `cargo test --locked --offline --test card release_audit_summary_cli::a_changed_ledger_after_a_partial_post_posts_nothing`
  verify: `cargo test --locked --offline --test card release_audit_summary_cli::one_post_with_the_planned_bytes_then_nothing`
- AC-3 (REQ-10, REQ-13, REQ-15): 완성된 묶음은 올린 순서와 상관없이 tier 2 에서 통과하고, 다른 base 나 세지 못하는 줄은 `missing` 이고, 내용이 같은 사본은 한 부분이에요 · verify: `cargo test --locked --offline --lib audit_gate::tests::a_complete_part_set_passes_tier_2`
  verify: `cargo test --locked --offline --lib audit_gate::tests::a_part_set_on_another_base_or_with_an_uncounted_line_is_missing`
  verify: `cargo test --locked --offline --lib audit_gate::tests::identical_copies_of_a_part_are_one_part`
- AC-4 (REQ-11, REQ-12): 부분 사이의 빈틈, 겹침, 중복 커밋, base 나 개수의 불일치, 같은 번호의 다른 내용, 손으로 고친 줄(id 불일치), 읽히지 않는 부분, 어느 부분의 `FAIL` 줄은 옛 PASS 가 있어도 tier 0 을 포함해 `fail` 이에요 · verify: `cargo test --locked --offline --lib audit_gate::tests::parts_that_contradict_each_other_fail`
  verify: `cargo test --locked --offline --lib audit_gate::tests::a_fail_line_or_an_unreadable_part_fails_at_every_tier`
  verify: `cargo test --locked --offline --lib release_audit_summary::tests::a_broken_part_is_refused`
- AC-5 (REQ-14, REQ-16): 빠진 부분이 있으면 모든 tier 에서 `missing` 이고, 두 묶음의 부분은 서로 채우지 못하고, 통과하는 묶음이 옛 단일 요약이나 다른 미완성 묶음을 덮지 못해요(반대도 같아요) · verify: `cargo test --locked --offline --lib audit_gate::tests::a_set_with_an_absent_part_is_missing_at_every_tier`
  verify: `cargo test --locked --offline --lib audit_gate::tests::parts_of_two_sets_never_complete_each_other`
  verify: `cargo test --locked --offline --lib audit_gate::tests::every_summary_and_set_of_the_head_must_pass`
- AC-6 (REQ-10, REQ-17): 부분만 있어도 병합 기준을 읽고, 명령이 낸 부분이 곧 게이트가 읽는 글이며, 부분이 없을 때의 판정은 그대로예요 · verify: `cargo test --locked --offline --lib audit_gate::tests::collect_reads_the_merge_base_when_only_parts_are_posted`
  verify: `cargo test --locked --offline --test card release_audit_summary_cli::the_parts_the_command_prints_are_what_audit_gate_reads`
  verify: `cargo test --locked --offline --test audit_gate_cli a_release_summary_part_set_decides_the_gate`
  verify: `cargo test --locked --offline --lib audit_gate::tests::a_gap_in_the_chain_or_a_wrong_count_is_refused`
  verify: `cargo test --locked --offline --test card release_audit_summary_cli::the_comment_the_command_prints_is_what_audit_gate_reads`
  verify: `cargo test --locked --offline --test audit_gate_cli`
  verify: `cargo test --locked --offline --test gate_registry_tier`

## Scope

- write: `.moai/specs/SPEC-20261126-release-summary-parts/spec.md`
- write: `src/release_audit_summary.rs`, `src/ops/release_audit_summary.rs`, `src/audit_gate.rs`
- write: `tests/card/release_audit_summary_cli.rs`, `tests/audit_gate_cli.rs`
- write: `docs/design-review/appendix/operations/release_audit_summary.md`, `docs/design-review/appendix/operations/audit_gate.md`
- read-only: `src/forge/connector.rs`, `src/forge/transport.rs`, `src/comment_authority.rs`, `src/gate.rs`, `tests/gate_registry_tier.rs`
- forbidden: `.github/workflows/audit-gate.yml`, 실제 HOME 과 원장, 실제 GitHub(테스트는 가짜 `gh` 와 RecordedTransport 만), 원장 스키마 변경, 파이썬

## Design

### 부분 댓글 문법

```
RELEASE-AUDIT-SUMMARY head=<40hex> base=<40hex> landings=<K> part=<i>/<n> set=<64hex> from=<40hex> to=<40hex>
S s1 <session id>
L <sha40> p=<12hex> ...          (줄 문법은 단일 요약과 같아요)

<!-- moai-release-audit-summary <head 40hex> set=<64hex> part=<i>/<n> -->
```

- `head`, `base` 는 묶음 전체의 PR head 와 병합 기준이에요. `landings` 는 이 부분의 줄 수예요. `from`, `to` 는 이 구간의 시작 커밋(제외)과 마지막 커밋이에요.
- `n` 은 2~99, `i` 는 1~n 이고 앞의 0 없는 10진수예요. 부분 하나짜리 묶음은 없어요.
- 부분 하나의 확인: 위 머리글, 부분 1 은 `from == base`, 부분 n 은 `to == head`, 기존 `check_chain(lines, from, to)`, 줄 수, 마커(있으면 head, set, i/n 일치), 60000바이트, 세션 표. 하나라도 어긋나면 읽히지 않는 댓글이에요. `ABS` 대상은 같은 부분 안의 더 나중 비-ABS 줄이어야 해요(지금 규칙 그대로).
- 부분마다 세션 표를 따로 가져요. 번호는 지금 만드는 쪽의 순서(감사자, 그다음 제외 목록)로 처음 쓰일 때 매겨요.
- 마커는 `n = 99` 일 때 156바이트로 `MARKER_MAX_BYTES`(200) 안이에요. 옛 단일 마커와 서로 부분 문자열이 아니에요.

### 묶음 id 와 plan digest

- set id: 각 부분의 본문에서 마커와 ` set=<값>` 칸을 뺀 렌더 글을 부분 순서대로 담은 JSON 배열의 SHA-256(16진수 64자)이에요. 만드는 쪽은 비밀 가리기(`redact`)를 지난 글로 계산하고, 게이트는 읽은 부분을 같은 함수로 다시 렌더해 계산해요. CRLF 는 영향이 없어요.
- plan digest: 마커까지 포함한 올릴 댓글 n 개의 JSON 배열에 대한 `crate::digest` 예요. 단일 요약의 digest 는 지금 그대로예요.

### 나누기

- 단일 댓글이 60000바이트 이하면 지금 형식 그대로예요. 넘을 때만 나눠요.
- 줄 i 뒤에서 자를 수 있는 것은 i 번째까지의 어떤 `ABS` 줄도 i 뒤의 줄을 가리키지 않을 때뿐이에요.
- 오래된 줄부터 들어가는 만큼 담아요. 들어가는지는 가리기를 지난 올릴 글(`part=99/99`, set 은 0 64자)의 바이트 수에 줄바꿈 수를 더해(CRLF) 60000 이하인지로 봐요. 결과는 입력이 같으면 같아요.
- 자를 수 없는 한 덩어리가 한도를 넘거나 99부분을 넘으면 거절해요.

### 명령의 출력과 올리기

- 나뉜 계획은 `set`, `plan_digest`, `parts[]`(`part`, `marker`, `bytes`, `text`)를 내고 최상위 `text`, `marker`, `bytes` 는 내지 않아요(일부를 전체로 읽지 않게). `counts`, `missing`, `complete`, `landings` 는 범위 전체예요.
- `--post` 한 번이 digest 하나와 permission reference 하나로 묶음 전체를 올려요. 모든 거절(값, digest, 불완전, 한도, 마커 모양)은 첫 전송 전이에요. 부분마다 `post_summary` 를 한 번 불러요(마커가 있으면 건너뛰고, POST 를 다시 보내지 않아요). 첫 실패에서 멈춰요.
- 성공 결과는 `state`(하나라도 올렸으면 `posted`, 아니면 `already_posted`), `writes`, `parts[]`(`part`, `state`, `comment_id`)예요.

### 게이트 판정 순서

| 순서 | 조건 | 결과 / rule |
| --- | --- | --- |
| 1 | 열린 `AUDIT #` FAIL | `fail` / `open-fail`(그대로) |
| 2 | 읽히지 않는 멤버 요약이나 부분(head 무관) | `fail` / `summary-malformed` |
| 3 | 이 head 의 단일 요약이나 부분에 `FAIL` 줄(묶음이 미완성이어도) | `fail` / `summary-fail` |
| 4 | 같은 set 의 부분끼리 모순, 또는 완성된 묶음의 id 불일치 | `fail` / `summary-malformed` |
| 5 | 이 head 의 묶음에 빠진 부분(tier 0 포함) | `missing` / `summary-parts` |
| 6 | tier 0 | `pass` / `tier-0` |
| 7 | 단일 요약과 묶음 각각: base 불일치 | `missing` / `summary-base` |
| 8 | 단일 요약과 묶음 각각: 세지 못하는 줄 | `missing` / `summary-landings` |
| 9 | 모두 통과 | `pass` / `release-summary` |
| - | 이 head 의 요약이 없음 | 옛 경로 그대로 |

- 묶음은 이 head 를 댄 부분 가운데 set 이 같은 것이에요. 이어짐은 부분 번호 순서로 봐요: 부분 1 의 `from` 이 base, 부분 k 의 `from` 이 부분 k-1 의 `to`, 부분 n 의 `to` 가 head 이고, 커밋은 묶음 전체에서 한 번만 나와요.
- `uncounted` 와 `Verdict::counts` 는 부분마다 그 부분의 세션 표로 그대로 써요.
- 결과의 `comment` 는 묶음이면 부분 1 의 id(5번은 있는 부분 중 번호가 가장 작은 것), 실패면 문제의 댓글 id 예요. 통과한 단위가 여럿이면 그 id 가 가장 큰 단위예요.
- `is_summary` 와 `collect` 의 병합 기준 읽기는 머리글 첫 낱말이 같아서 부분에도 적용돼요. `Input` 필드는 늘지 않아요.
- 게이트는 git 을 읽지 않아요. "first-parent 사슬"은 지금처럼 줄에 적힌 부모 값과 구간 경계의 이어짐이에요.

### 옛 단일 요약과 부분 묶음

- 같은 head 의 단일 요약과 묶음은 각각 통과해야 해요. 통과하는 쪽이 다른 쪽을 구하지 않아요. 불완전하게 올린 옛 단일 요약(`--allow-incomplete`)은 지워질 때까지 막아요.
- 옛 마커는 부분을 "이미 올림"으로 만들지 않아요. 같은 head 에 단일 요약이 이미 있을 때 새 단일 요약이 `already_posted` 가 되는 지금 동작은 그대로예요.
- 단일만 아는 게이트는 부분 머리글의 칸 수 때문에 `summary-malformed` 로 닫혀요. 요약을 모르는 게이트(첫 릴리스의 base)는 읽지 않아요.

### 잘못 올라간 부분 고치기

- 명령은 지우거나 고치지 않아요. 게이트는 어떤 댓글도 새 댓글로 덮지 않아요.
- 손으로 고쳐졌거나 틀린 부분(마커는 같음): 게이트는 `fail` 이나 `missing` 이고, 재실행은 마커 때문에 건너뛰어요. 사람이 그 댓글을 지운 뒤 같은 명령을 돌리면 그 부분만 올라가요.
- 원장이나 범위가 바뀜: digest 가 달라 거절돼요. 새 계획은 새 set 이에요. 통과하지 못하는 옛 묶음의 댓글을 사람이 지우고, 미리보기와 승인을 새로 받아 올려요.
- 새 PR head 는 읽히는 옛 요약을 벗어나게 하지만, 읽히지 않는 댓글은 head 와 상관없이 계속 막아요.
- 지워도 원장의 FAIL 은 사라지지 않아요: 다시 만든 계획에 `FAIL` 줄이 다시 나와요.

## Plan

1. 사용자의 승인과 Interview 답을 기다려요(tier 2). 카드를 잡고 위 `write:` 범위를 예약해요.
2. 부분 문법(`src/release_audit_summary.rs`): 테스트를 먼저 쓰고 `render`/`parse` 의 부분 형식, set id, 마커를 더해요. 기존 단일 형식의 테스트는 바꾸지 않아요.
3. 나누기와 미리보기(같은 파일, `tests/card/release_audit_summary_cli.rs`): `a_summary_over_the_cap_is_refused_not_cut` 은 이름을 지켜요(이전 SPEC 의 verify 줄이 가리켜요). 거절 대상을 자를 수 없는 `ABS` 덩어리(빨리감기 900커밋)로 바꾸고, 700개 직접 커밋은 새 테스트에서 부분이 돼요.
4. 올리기(같은 두 파일): 가짜 `gh` 가 지금은 `post-body.txt` 하나와 준비된 댓글 하나만 다뤄요. POST 마다 따로 저장하고, 저장한 댓글을 모두 돌려주고, k 번째 POST 를 실패시킬 수 있게 바꿔요.
5. 게이트(`src/audit_gate.rs`, `tests/audit_gate_cli.rs`): RED 뒤 GREEN. 모듈 문서의 규칙 5 를 고쳐요.
6. 문서 두 조각과 `src/ops/release_audit_summary.rs` 의 긴 도움말(첫 줄은 그대로).
7. `cargo fmt --all`, `moai check affected --card` 한 번. verify 줄마다 테스트가 1개 이상 실제로 돌았는지 확인해요. READY 에 위험 경로 실행(임시 HOME 에서 진짜 `moai` 바이너리와 가짜 `gh` 를 돌리는 Rust 통합 테스트로 부분 실패 뒤 재실행)과 `guard-mutant:` 줄을 담아요.

뮤턴트(하나씩): `ABS` 덩어리 안에서 자르기, CRLF 계산 빼기, 첫 전송 전 거절 빼기, 실패 뒤 계속 보내기, 구간 경계 확인 빼기, 묶음 전체 중복 확인 빼기, set id 재계산 빼기, 빠진 부분 허용, tier 0 반환을 부분 확인 앞으로 옮기기, 미완성 묶음의 `FAIL` 줄 무시, 내용이 다른 같은 번호 합치기, 새 묶음이 옛 묶음 덮기, 부분의 base 확인 빼기, 부분의 `uncounted` 빼기.

## Risks

- 게이트 경로, tier 2: 승인은 사용자이고, 작성하지 않은 세션의 전체 감사가 필요해요.
- 부분을 읽으려면 PR base 에 이 변경이 있어야 해요. 첫 릴리스에서는 먼저 올리는 작은 PR 에 이 게이트 변경이 빠지면 효과가 없어요(Intent 의 의존).
- 부분만 올라간 사이 게이트는 통과하지 않아요. 그 사이 원장이 바뀌면 올린 댓글을 사람이 지워야 해요.
- 멤버나 토큰 계정의 댓글이 부분 마커를 인용하면 그 부분을 건너뛰어 `missing` 으로 남아요. 비멤버 토큰 계정으로 올린 부분은 게이트가 읽지 않아요(지금 단일 요약과 같아요).
- 멤버는 요약을 꾸미거나 지우고 새로 쓸 수 있어요: 지금과 같은 신뢰 수준이에요.
- 새 댓글은 `audit-gate` 작업을 다시 시작하지 않아요. 앞선 상태는 재실행 전까지 남아요.
- 댓글 한도는 코드의 값만 확인했어요. 코드 주석은 GitHub 한도를 65536자라고 적고 공식 문구는 읽지 않았다고 해요: 미확인이에요. UTF-8 에서 60000바이트가 60000자 이하라는 것은 계획자의 추론이에요. GitHub 가 줄 끝을 CRLF 로 돌려주는 조건도 미확인이에요.
- 동시에 두 번 올리면 사본이 생길 수 있어요(같은 내용은 한 부분). 부분마다 댓글 목록을 다시 읽고, 게이트는 댓글 30페이지까지 읽어요.
- 올리기는 `gh` 프로세스를 시작해요: 위험 경로 실행이 필요해요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 부분 머리글은 같은 첫 낱말에 칸을 더해요(기본). 다른 선택: 새 낱말 | 새 낱말은 단일만 아는 게이트가 무시해서, 부분에 `FAIL` 줄이 있어도 옛 `AUDIT #` PASS 로 통과할 수 있어요 | 새 낱말과 `is_summary` 확장 |
| D-2 | 한도를 넘을 때만 나눠요(기본). 다른 선택: 항상 부분, `--parts` 플래그 | 단일 요약의 바이트와 digest 를 지켜요 | 플래그를 요구 |
| D-3 | set id 는 set 칸을 뺀 내용의 digest, plan digest 는 올릴 바이트 전체(기본). 다른 선택: 임의 id | id 를 글에 넣으므로 올릴 바이트의 digest 일 수 없고, 게이트가 다시 계산할 수 있어야 해요 | 임의 id(고친 줄을 못 잡아요) |
| D-4 | 빠진 부분은 tier 0 포함 `missing`(기본). 다른 선택: `fail` | 닫혀 있고, 올리기를 마치면 풀려요. 안 온 부분에 `FAIL` 줄이 있을 수 있어요 | `fail` 로 |
| D-5 | 내용이 같은 사본은 한 부분(기본). 다른 선택: 사본이 있으면 `fail` | 빈틈, 겹침, 커밋 중복이 생기지 않아요 | `fail` 로(사람이 지워야 해요) |
| D-6 | 이 head 의 모든 요약과 묶음이 통과해야 해요(기본). 다른 선택: 새 완성 묶음이 옛것을 대신 | 대신하게 하면 옛 `FAIL`·`NONE` 줄을 가릴 수 있어 약화예요 | 사용자 결정과 새 SPEC |
| D-7 | 명령은 댓글을 고치거나 지우지 않아요(기본). 다른 선택: 승인된 삭제 명령. withdraw 줄은 버려요 | 원격 삭제는 사용자의 선택이고 withdraw 줄은 게이트를 약하게 해요(Interview Q2) | 별도 SPEC 의 삭제 명령 |
| D-8 | 올리기 전에 다른 묶음이 있는지 읽어 거절하지 않아요(기본). 다른 선택: connector 에 읽기를 더해 거절 | 게이트가 이미 닫히고 connector 를 바꾸지 않아요 | connector 읽기 추가(범위가 늘어요) |
| D-9 | 같은 digest 의 재실행은 같은 permission reference 로(기본). 다른 선택: 재실행마다 새 승인 | Interview Q3 | 새 참조를 요구 |
| D-10 | 부분은 99개까지(기본). 다른 선택: 더 작은 상한 | 마커 길이가 정해지고 댓글 읽기 한도 안이에요 | 상한을 낮춰요 |
| D-11 | 명령 도움말의 첫 줄은 그대로(기본). 다른 선택: "PR comments" 로 | 첫 줄을 바꾸면 공유 도움말 골든 네 개가 바뀌어요 | 첫 줄과 골든을 함께 고쳐요 |
| D-12 | 부분 크기는 CRLF 를 센 바이트(기본). 다른 선택: 올린 바이트만 | 한도 가까이 담은 부분이 줄 끝 변환으로 읽히지 않는 일을 막아요 | 올린 바이트만 세요 |
