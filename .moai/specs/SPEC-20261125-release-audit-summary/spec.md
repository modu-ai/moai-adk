---
id: SPEC-20261125-release-audit-summary
title: "One command gathers the release range's verdicts from the ledger into one PR comment, and the audit gate reads that comment"
status: draft
revision: 1
created: 2026-11-25
---

# One command gathers the release range's verdicts from the ledger into one PR comment, and the audit gate reads that comment

## Intent

릴리스 PR 의 `audit-gate` 는 지금 사람이 손으로 붙인 `AUDIT #` 댓글만 읽고 원장을 읽지 않아요(`src/audit_gate.rs`). 그런데 판정은 원장(`card.verdict` 이벤트)에 있고, 착지는 git 의 first-parent 줄과 원장의 `card.landed` 행에 있어요. 릴리스 PR 하나에는 카드가 수백 개 묶이므로 "착지마다 `AUDIT #` 댓글 하나"는 게이트의 "정확히 이 head" 규칙과 맞지 않고, PR 커밋 목록은 250개에서 끊겨 작성자를 알 수 없어요(게이트는 닫혀요). 사용자 선택(2026-11-24T23:58:11Z, 카드 본문): 명령 하나가 릴리스 범위의 first-parent 착지마다 판정(감사 세션, head, PASS)을 모아 PR 댓글 하나로 올리고, 게이트가 그 형태를 읽어요. 판정 없는 착지는 이름으로 나열하고 게이트는 실패해요. 서명 태그는 팀 모드 때의 후속이에요. 원격 쓰기는 사용자가 릴리스를 고른 때에만이므로, 명령의 기본은 아무것도 보내지 않는 미리보기(dry-run)이고 올리기는 명시한 플래그와 확인 값이 있을 때만이에요.

## Interview

- Q1: 판정이 없는 착지(직접 커밋, 판정 없이 닫힌 옛 카드)를 어떻게 다룰까요? 선택지: (가) 엄격: first-parent 커밋마다 PASS 가 있어야 게이트가 통과해요 (나) 낮은 tier 착지는 PASS 없이 통과해요 (다) 옛 카드에 소급 판정을 적어요. → A: (가) 엄격. 사용자가 2026-11-25T19:32:12Z 에 리더 세션의 선택 질문에 고른 답 그대로 "(가) 엄격 (Recommended)"이고 원장 노트 9473 에 있어요. 첫 릴리스 전 판정 없는 커밋 처리는 따로 정해요(릴리스 계획 문제로 리더가 아침 보고에 올려요). 비용은 D-1 에 있어요: 지금 범위로는 첫 릴리스가 통과하지 못해요.
- Q2: 판정 없는 행의 카드가 다른 head 의 PASS 이벤트를 갖고 있으면 그것으로 채울까요? → A: 아직 묻지 않았어요. 기본: 예, 착지 검사(`check::judge`)와 같은 규칙으로 판정해요(D-2).
- Q3: 이 명령이 올린 댓글을 게이트가 어떻게 믿을까요? → A: 아직 묻지 않았어요. 기본: 멤버의 댓글만 읽고, 이름 붙인 head 와 base 가 PR 과 같고, 줄 수와 이어짐이 맞고, 감사자가 줄마다 적힌 제외 세션에 없을 때만 PASS 로 읽어요. 줄의 사실 자체(작성자, 판정)는 멤버 댓글을 믿어요: 지금 멤버가 가짜 `AUDIT #` PASS 를 쓰는 것과 같은 신뢰 수준이에요(Risks).

## Non-goals

- 서명된 태그로 판정을 올리는 일(카드 본문: 팀 모드 때의 후속).
- 릴리스 브랜치 이름 `release/v{head}.{주번호}.{n}`(별도 카드의 몫).
- `.github/workflows/audit-gate.yml` 의 수정: 게이트는 출력의 `state` 와 `comment` 만 읽으므로 필요하지 않고, 이 파일은 CI 면이에요.
- 첫 릴리스 범위의 판정 없는 착지를 소급해서 채우는 일(D-1 선택지 (다)).
- first-parent 커밋마다 GitHub API 로 걸어서 완전성을 증명하는 일(후속; 호출 횟수 비용은 확인되지 않았어요).
- 기존 `AUDIT #` 댓글 경로의 변경: 그 규칙은 한 줄도 바뀌지 않아요.

## Requirements

- REQ-1: When the command runs on a base and a head, the command shall print exactly one line for each first-parent commit of `base..head`, oldest first. · failure: none — a commit with no verdict is still a line (REQ-2).
- REQ-2: When a first-parent commit has no passing verdict on the ledger, the command shall print its line as `NONE` or `FAIL` naming the commit and the cards it closed. · failure: none — a missing verdict is a line, never a skipped one.
- REQ-3: When one first-parent commit closed several cards, the command shall list every one of those cards on that commit's one line. · failure: none — a commit with one card is unchanged.
- REQ-4: While any card listed on a line has no passing verdict, the command shall not print that line as `PASS`.
- REQ-5: While a verdict's reviewer wrote or held the card, the command shall not count that verdict as a pass.
- REQ-6: The command shall list on every `PASS` line the sessions that wrote or held its cards.
- REQ-7: The command shall by default print the comment, its digest and the missing landings and send nothing.
- REQ-8: When `--post` is given with a PR, a repository, the plan's digest, a permission reference and a complete summary, the command shall post the planned comment once. · failure: none — a missing value, a changed plan or an incomplete summary is refused by REQ-9.
- REQ-9: IF `--post` lacks one of those values, or the plan or the ledger changed since the plan, or the summary is incomplete, THEN the command shall refuse and post nothing.
- REQ-10: IF the comment would exceed 60000 bytes, THEN the command shall refuse instead of truncating it.
- REQ-11: When a member's comment starts with the summary header, names exactly the PR head and a base equal to the merge base, and its lines form an unbroken chain from that base to that head, the audit gate shall read it as a summary. · failure: none — a comment that does not parse is a `fail` by REQ-12.
- REQ-12: IF a member's summary has a `FAIL` line or does not parse, THEN the audit gate shall decide `fail`.
- REQ-13: While a summary line is `NONE`, the audit gate shall decide `missing` and name the landings.
- REQ-14: When the audit gate reads a `PASS` line, the audit gate shall count it only if its reviewer is not among the line's excluded sessions (compared with `trailers::same_identity`) and that list is not empty. · failure: none — a reviewer among them is `missing` by REQ-13's rule for an uncounted line.
- REQ-15: The audit gate shall decide every comment that is not a summary exactly as before this SPEC.
- REQ-16: While an `AUDIT #` FAIL is open, the audit gate shall not decide `pass` for a valid summary.

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-3, REQ-4): 임시 원장과 임시 저장소에서 first-parent 커밋마다 한 줄, 오래된 것 먼저; 행이 없는 직접 커밋과 판정 없는 행은 이름으로 나열되고, 한 병합 커밋이 닫은 카드는 한 줄에 모두 나와요 · verify: `cargo test --locked --offline --test card release_audit_summary_cli::one_line_per_first_parent_commit_and_none_is_dropped`
  verify: `cargo test --locked --offline --test card release_audit_summary_cli::a_landing_without_a_verdict_is_listed_by_name`
  verify: `cargo test --locked --offline --test card release_audit_summary_cli::a_batch_commit_lists_every_card_it_closed_on_one_line`
  verify: `cargo test --locked --offline --test card release_audit_summary_cli::a_fast_forwarded_lane_absorbs_only_its_own_commits`
- AC-2 (REQ-5, REQ-6): 카드를 쓰거나 예약한 세션이 감사자인 판정은 PASS 줄이 되지 않아요; 판정 이벤트만 있고 행에 이름이 없는 카드는 착지 규칙으로 판정해요; 착지 뒤에 기록된 FAIL 은 `FAIL` 줄이에요 · verify: `cargo test --locked --offline --test card release_audit_summary_cli::a_reviewer_who_wrote_or_held_the_card_is_not_a_pass_line`
  verify: `cargo test --locked --offline --test card release_audit_summary_cli::a_pass_event_without_a_relied_row_is_judged_by_the_land_rules`
  verify: `cargo test --locked --offline --test card release_audit_summary_cli::a_fail_recorded_after_the_landing_marks_the_line_fail`
- AC-3 (REQ-7, REQ-8, REQ-9, REQ-10): 기본 실행은 gh 를 부르지 않고 쓰지 않아요; 올리기는 플래그, PR, digest, 허가 참조, 완전한 요약이 모두 있어야 하고 가짜 gh 로 정확히 한 번 올려요; 계획 뒤 원장이나 범위가 바뀌면 올리지 않고, 한도를 넘으면 자르지 않고 거절해요 · verify: `cargo test --locked --offline --test card release_audit_summary_cli::the_dry_run_prints_the_comment_and_digest_and_sends_nothing`
  verify: `cargo test --locked --offline --test card release_audit_summary_cli::posting_needs_flag_pr_digest_permission_and_a_complete_summary`
  verify: `cargo test --locked --offline --test card release_audit_summary_cli::one_post_with_the_planned_bytes_then_nothing`
  verify: `cargo test --locked --offline --test card release_audit_summary_cli::a_changed_ledger_or_range_after_the_plan_posts_nothing`
  verify: `cargo test --locked --offline --test card release_audit_summary_cli::a_summary_over_the_cap_is_refused_not_cut`
  verify: `cargo test --locked --offline --lib release_audit_summary::tests::render_and_parse_round_trip`
- AC-4 (REQ-11, REQ-12, REQ-13, REQ-14): 완전한 요약과 비작성자 PASS 는 tier 2 에서 통과해요; `NONE` 줄은 이름과 함께 `missing`; 제외 세션에 든 감사자, 빈 목록, 낡은 맨 id 대 전체 id 는 통과가 아니에요; 다른 head 나 다른 base, 끊긴 이음이나 틀린 줄 수, 비멤버 댓글은 읽히지 않아요 · verify: `cargo test --locked --offline --lib audit_gate::tests::a_complete_summary_with_non_author_passes_passes_tier_2`
  verify: `cargo test --locked --offline --lib audit_gate::tests::a_landing_without_a_verdict_fails_the_gate_by_name`
  verify: `cargo test --locked --offline --lib audit_gate::tests::a_reviewer_among_the_excluded_sessions_is_not_a_pass`
  verify: `cargo test --locked --offline --lib audit_gate::tests::a_summary_for_another_head_or_base_does_not_count`
  verify: `cargo test --locked --offline --lib audit_gate::tests::a_gap_in_the_chain_or_a_wrong_count_is_refused`
  verify: `cargo test --locked --offline --lib audit_gate::tests::a_non_member_summary_does_not_count`
- AC-5 (REQ-15, REQ-16): 열린 `AUDIT #` FAIL 이 유효한 요약을 막고, 실패하는 요약이 옛 경로의 통과를 막고, 병합 기준은 요약이 있을 때만 읽고, 명령이 낸 글이 곧 게이트가 읽는 글이에요; 옛 동작은 그대로예요 · verify: `cargo test --locked --offline --lib audit_gate::tests::an_open_fail_still_blocks_a_valid_summary`
  verify: `cargo test --locked --offline --lib audit_gate::tests::a_failing_summary_blocks_even_when_the_old_path_passes`
  verify: `cargo test --locked --offline --lib audit_gate::tests::collect_reads_the_merge_base_only_when_a_summary_is_present`
  verify: `cargo test --locked --offline --test card release_audit_summary_cli::the_comment_the_command_prints_is_what_audit_gate_reads`
  verify: `cargo test --locked --offline --test audit_gate_cli a_release_summary_comment_decides_the_gate`
  verify: `cargo test --locked --offline --test audit_gate_cli`
  verify: `cargo test --locked --offline --test gate_registry_tier`

## Scope

- write: `.moai/specs/SPEC-20261125-release-audit-summary/spec.md`
- write: `src/release_audit_summary.rs`, `src/ops/release_audit_summary.rs`, `src/audit_gate.rs`
- write: `tests/card/release_audit_summary_cli.rs`, `tests/audit_gate_cli.rs`, `tests/gate_registry_tier.rs`
- write: `docs/design-review/appendix/operations/release_audit_summary.md`, `docs/design-review/appendix/operations/audit_gate.md`
- write (D-3, D-4 기본안; 승인 뒤 `card scope` 로 넓혀요): `src/card/land/check.rs`(`authors_of` 와 `Authors` 를 `pub(crate)` 로), `src/ops/audit_gate.rs`(도움말 한 문장, 선택)
- shared files (한 줄 임대 `card line-lease`, 커밋 뒤 바로 반납): `src/lib.rs`(`pub mod release_audit_summary;`), `src/help.rs`(GROUPS 와 EXAMPLES 한 줄씩), `tests/fixtures/help/all.txt`, `all-dumb.txt`, `all-80.txt`, `all-60.txt`(`MOAI_HELP_GOLDEN_WRITE=1` 로 다시 만들어요), `tests/card/main.rs`(`mod release_audit_summary_cli;`), 필요할 때만 `tools/repo_scans.json`
- read-only: `src/ledger/workflow.rs`, `src/card/land.rs`, `src/forge/**`, `src/comment_authority.rs`, `src/gate.rs`
- forbidden: `.github/workflows/audit-gate.yml`, 실제 HOME 과 원장, 실제 GitHub(PR 과 댓글 만들기 포함: 시험은 가짜 gh 와 RecordedTransport 만), 원장 스키마 변경, 파이썬

## Design

댓글 형식(첫 줄은 `AUDIT #` 로 시작하지 않아요; 본문에는 16진수, id, 숫자만 있고 카드 제목은 없어요):

```
RELEASE-AUDIT-SUMMARY head=<40hex> base=<40hex> landings=<N>
S s1 <session id>            (세션 표, 세션마다 한 줄)
L <sha40> p=<첫 부모 12hex> PASS v=<card>:<감사한 head 12hex>:<tier>:<감사자 sN>:<제외 sN.sN>:<이벤트 id>
L <sha40> p=<12hex> ABS <대상 sha12>
L <sha40> p=<12hex> NONE c=<card,card> | <제목 정리본 ≤40바이트>
L <sha40> p=<12hex> FAIL why=<코드> c=<card,…> | <제목>

<!-- moai-release-audit-summary <head 40hex> -->
```

- 줄은 first-parent 커밋마다 하나이고 오래된 것부터예요. `ABS` 는 더 새로운 착지에 흡수된 커밋(빨리감기한 레인의 중간 커밋)이고 대상 줄을 가리켜요. 흡수는 `card.land_started` 의 `tip` 을 알 때만 인정해요; 모르면 중간 커밋은 자기 `NONE` 줄이에요(닫힌 쪽으로).
- 제외 세션은 `authors_of(L^1, 감사한 head)` 의 커밋 작성자와 카드의 예약 보유자예요(착지 검사의 `self_verdict` 와 같은 집합). 세션 표를 두는 것은 줄마다 긴 id 를 되풀이하면 한도를 넘기 때문이에요.
- 제목은 제목 문자열에만 `clean_public_text` 를 써요(첫 줄에는 쓰지 않아요: `audit` 로 시작하는 줄에 `· ` 가 붙어요). 올리는 바이트는 `summary_comment_body(body, marker)` 이고 digest 가 그 바이트를 덮어요. 마커는 한 번만 올리게 해요(`post_summary` 는 갱신하지 않아요).
- 크기: PASS 줄 하나가 약 100바이트이므로 436줄이 약 44KB 로 추정해요(측정 아님). 60000바이트가 넘으면 거절해요(GitHub 의 65536자 한도 아래; 공식 문구는 읽지 못했어요).

게이트 규칙(기존 규칙 1, 2, tier 0 반환과 `AUDIT #` 경로는 그대로):

1. `Input` 에 `#[serde(default)] release_base: Option<String>` 을 더해요(PR base 와 head 의 merge base). `collect()` 는 어떤 댓글의 첫 줄이 `RELEASE-AUDIT-SUMMARY ` 로 시작할 때만 읽어요. `schema_version` 은 1 이에요. 이 필드를 리터럴로 만드는 `tests/gate_registry_tier.rs` 한 줄을 고쳐요.
2. 멤버의 요약에 `FAIL` 줄이 있거나 머리글이 있는데 읽히지 않으면 열린 FAIL 이에요(`summary-fail`, `summary-malformed`).
3. tier 1 과 2: 이 head 를 이름으로 댄 멤버 요약은 모두 통과해야 해요. `NONE` 줄이 하나라도 있으면 옛 경로가 통과해도 `missing`(`summary-landings`, 이름 10개까지와 개수)이에요. 요약이 없으면 옛 경로의 결과를 그대로 돌려줘요.
4. 요약마다 확인: `base == release_base`(읽지 못하면 `missing`/`summary-base`), 줄 수 == `landings` 이고 1 이상, 이어짐(첫 `p` 가 base 앞 12자, 각 `p` 가 앞 줄 sha 앞 12자, 마지막 sha 가 head, 중복 없음), `ABS` 대상이 더 나중의 비-ABS 줄, PASS 줄마다 `v=` 가 하나 이상이고 정의된 id 이고 제외 목록이 비지 않음(작성자 모름은 닫힘), 감사자가 제외 목록에 없음.

게이트가 스스로 확인하는 것: 멤버의 댓글, 정확히 PR head, GitHub 의 merge base 와 같은 base, 줄 수와 이어짐의 일관성, 줄의 제외 목록. 멤버 댓글을 믿는 것: 목록이 실제 이력과 완전하다는 것, 감사자·작성자·판정의 사실, 원장 판단이 옳았다는 것.

명령(`release-audit-summary PROJECT --base <40hex> --head <40hex> [--home DIR]`): `--base`/`--head` 는 문자열 인자로 두고 `run()` 안에서 검증해요(필수 인자는 조각 변이 시험의 값 `1` 을 받아야 해요). 기본은 미리보기: `text`, `plan_digest`, `landings`, `counts`, `missing[]`, `complete`, `writes:false`, `network:"none"` 을 내고 gh 를 부르지 않아요. 올리기는 `--post --pr N --repo OWNER/REPO --plan-digest D --permission-reference R` 이 모두 있어야 하고, digest 를 다시 계산해 비교하고, `permission_reference_ok` 를 확인하고, 불완전한 요약은 `--allow-incomplete` 없이는 거절하고, `GithubConnector::post_summary` 로 한 번 올려요. `--require-complete` 는 불완전한 요약을 오류로 만들어요.

## Plan

1. 이 SPEC 의 승인을 기다려요(tier 2: 게이트 경로). 승인 뒤에 `card scope` 로 범위를 넓히고 공유 파일의 한 줄 임대를 받아요.
2. 문법: `src/release_audit_summary.rs` 에 `render` 와 `parse` 를 시험 먼저 써요(왕복, 깨진 줄 거절, 한도, 첫 줄이 `AUDIT #` 로 시작하지 않음). 만드는 쪽과 게이트가 같은 문법을 써요.
3. 만드는 쪽 RED: 임시 `TempHome` 과 `commit-tree` 로 만든 임시 저장소, 원시 SQL 로 넣은 `card.landed`·`card.verdict` 행으로 `tests/card/release_audit_summary_cli.rs` 를 써요.
4. 만드는 쪽 GREEN: first-parent 목록, 원장 읽기 전용으로 `card.landed`·`card.land_started` 읽기, 행과 커밋 맞추기와 흡수, 카드마다 감사한 head 고르고 `check::judge` 로 판정, 제외 세션 모으기.
5. 명령과 조각 문서, `lib.rs`·`help.rs` 한 줄, 도움말 골든 네 개 다시 만들기.
6. 올리기: 플래그, digest 재확인, 마커, 한도와 불완전 거절을 가짜 gh 와 RecordedTransport 로 시험해요.
7. 게이트 RED 와 GREEN: `release_base`, 열린 FAIL 확장, 요약 규칙, 마지막에 `tests/audit_gate_cli.rs`. 끝에서 끝까지: 명령의 미리보기 글을 멤버 댓글로 게이트에 넣으면 `pass`, 한 줄을 빼면 `missing`.
8. 문서(`audit_gate.md`)와 SPEC Result, `cargo fmt --all`, `moai check affected --card` 한 번. READY 에 위험 경로 실행(가짜 gh 로 올리는 길을 임시 폴더에서)과 `guard-mutant:` 줄(아래 변이)을 담아요.

변이(READY 에 줄마다 `guard-mutant: <시험> | <뺀 코드> | exit <n>`; 정확히 한 번 편집하고 git 으로 복원): 기준 1 쪽은 (1) 행 없는 커밋의 줄 빼기 (2) 조상/`from_tip` 검사 빼기 (3) 판정 없는 행 건너뛰기 (4) 공유 head 의 첫 카드만 (5) `--first-parent` 빼기 (6) 최신순 (7) 예약 보유자 빼기 (8) 작성자 빼기 (9) 늦은 FAIL 건너뛰기 (10) 한도에서 자르기 (11) 미리보기에서 올리기 (12) digest 재확인 빼기; 기준 2 쪽은 (13) 감사자 ∉ 제외 목록 검사 빼기 (14) 빈 제외 목록 허용 (15) `same_identity` 대신 `==` (16) `NONE` 무시 (17) head 일치 빼기 (18) base 일치 빼기 (19) 줄 수·이어짐 검사 빼기 (20) 요약에서 `retain_members` 빼기 (21) 요약을 열린 FAIL 규칙보다 먼저 판정 (22) 옛 PASS 가 실패 요약을 덮게 (23) `v=` 없는 PASS 줄 허용.

## Risks

- 게이트 경로, tier 2: SPEC 승인은 사용자예요. 감사는 작성하지 않은 세션의 전체 감사여야 해요.
- 첫 릴리스 범위(`origin/v2026.11.0..NeXT`)는 지금 그대로는 통과하지 못해요(D-1 (가) 선택 시): 첫 부모 커밋 436개 중 53개가 카드 없는 직접 커밋이고, 착지 행 293개 중 252개가 판정을 이름 대지 않고, 판정 이벤트가 있는 카드는 일부예요. D-1 의 (나) 또는 (다)를 사용자가 정하기 전에는 릴리스 PR 의 이 게이트가 `missing` 이에요. 이 숫자는 계획 단계에서 읽기 전용으로 센 값이고 이 SPEC 의 시험은 아니에요.
- 규모: 한 댓글에 약 500~550 착지가 상한이라는 추정이에요. 넘으면 명령이 오류를 내요("댓글 하나"와 범위 크기가 부딪혀요).
- 최초로 이 변경을 싣는 릴리스는 옛 게이트(PR base 에서 빌드)가 판정해요. 옛 게이트는 250 커밋이 넘는 tier 2 PR 을 통과시키지 못해요(D-6: 사용자 결정, 이 카드는 워크플로를 고치지 않아요).
- 멤버가 자기 일관성 있는 이음 줄을 꾸며 올리면 통과할 수 있어요: 멤버가 가짜 `AUDIT #` PASS 를 쓰는 것과 같은 신뢰 수준이에요. 실수로 빠뜨리기, 잘림, 낡은 head, 틀린 base 는 막혀요. 완전성의 증명(커밋마다 API)과 서명 태그는 후속이에요.
- 새 댓글은 `audit-gate` 작업을 다시 시작하지 않아요: 올린 뒤 `gh run rerun` 이 필요해요. CI 로그는 `reason` 을 가려서 빠진 착지의 이름은 댓글과 `moai audit-gate --pr` 에서만 보여요.
- `Input` 에 필드를 더하면 리터럴로 만드는 곳이 깨져요(`tests/gate_registry_tier.rs:761` 하나).
- `judge` 는 현재 원장으로 다시 판정하므로 나중에 감사자가 카드를 예약하면 줄이 `self_verdict` 로 바뀌어요: 보수적이고 버그가 아니에요.
- 올리기는 `gh` 프로세스를 시작하므로 READY 에 위험 경로 실행이 필요해요. `tests/card/main.rs` 와 공유 파일의 한 줄은 커밋 뒤 바로 반납해요.

## Decisions

D-2 부터 D-8 은 레인 추천값으로 가요(리더 배차 2026-11-25, 승인 때 함께 읽혀요).

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 판정 없는 착지: 엄격(first-parent 커밋마다 PASS 필요). 사용자 선택 2026-11-25T19:32:12Z, 노트 9473 | 카드 본문 "게이트는 실패해요"; 사용자가 (가)를 골랐어요(낮은 tier 면제 (나)와 소급 판정 (다)는 고르지 않았어요) | 줄마다 tier 를 읽는 면제를 후속 SPEC 으로(사용자 결정) |
| D-2 | 행이 판정을 이름 대지 않으면 카드의 판정 이벤트를 착지 검사 규칙으로 판정해요 | 같은 규칙 한 곳, 약화 없음 | 이름 대지 않은 행을 `NONE` 으로 |
| D-3 | 시험 파일은 `tests/card/release_audit_summary_cli.rs`(첫 단어 `release` 는 `card` 영역); `tests/card/main.rs` 한 줄은 임대 | `test_area_layout` | 최상위 `tests/audit_release_summary_cli.rs`(main.rs 수정 없음) |
| D-4 | 작성자는 `src/card/land/check.rs` 의 `authors_of`/`Authors` 를 `pub(crate)` 로 열어 같은 원천을 써요 | 진실의 원천 하나 | 새 모듈에 25줄 복사와 같음 시험 |
| D-5 | 불완전한 요약은 `--allow-incomplete` 없이 올리지 않아요 | 한 head 의 모든 요약이 통과해야 해서 잘못 올린 것은 지우지 않고 고칠 수 없어요 | 플래그 없이 올리기 허용 |
| D-6 | 이 변경을 처음 싣는 릴리스는 옛 게이트가 판정해요; 이 카드는 워크플로를 고치지 않아요 | CI 면은 사용자의 선택 | 릴리스 때 사용자가 처리 방법을 정해요 |
| D-7 | `land_started` 행이 없는 빨리감기 중간 커밋은 자기 `NONE` 줄이에요 | 닫힌 쪽으로 | 흡수 추정 허용 |
| D-8 | 올리기는 한 명령의 `--post` 플래그와 digest·허가 참조로 확인해요(기존 `card-summary-plan`/`apply` 의 두 명령 방식이 아님) | 카드가 "명령 하나"를 말해요 | plan 과 apply 두 명령으로 나누기 |
