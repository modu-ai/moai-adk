---
id: SPEC-20261123-claude-md-rewrite
title: "CLAUDE.md is rewritten to about 200 lines of team-shared facts and CLAUDE.local.md gets a machine-facts draft"
status: draft
revision: 1
created: 2026-11-23
parent: SPEC-20261122-claude-md-layout
---

# CLAUDE.md is rewritten to about 200 lines of team-shared facts and CLAUDE.local.md gets a machine-facts draft

## Intent

사용자 결정(2026-11-23): 프로젝트 `CLAUDE.md` 는 이 저장소 팀이 공유하는 사실만 담고, `CLAUDE.local.md` 는 개인·기계 사실만 담고, 배포용 전역 지침(`~/.claude/CLAUDE.md`)은 moai 전체 공통이어야 해요. 지금 `CLAUDE.md` 는 44,900 B(예산 45,056 B)에 일반 MoAI 규칙, 이 저장소 규칙, 리더 운영 세부, 출처 표 37행, 관리 구역이 섞여 있고, 공식 문서는 파일당 200줄 아래를 권해요. 안에는 지금 코드·결정과 어긋난 규칙도 여럿 있어요(아래 S1~S6). 이 SPEC 은 (1) 지금 규칙마다 남김·옮김·폐기를 적은 표(`rule-table.md`)를 먼저 만들어 조용히 사라지는 규칙이 없게 하고, (2) 표대로 `CLAUDE.md` 를 다시 쓰고 세부를 `docs/` 로 옮기고, (3) 어긋난 곳을 코드·결정에 맞추고, (4) `CLAUDE.local.md` 초안을 이 SPEC 폴더에 두고(실제 파일은 무시 파일이라 커밋하지 않고, 착지 뒤 리더가 옮겨요), (5) 이 파일을 읽는 테스트를 새 구조에 맞춰요. 전역 파일(`~/.claude/CLAUDE.md`, `rules/moai.md`)은 이 SPEC 의 범위가 아니에요(`SPEC-20261122-claude-md-home-routing`).

근거 보고서: 'MoAI 지침 파일 구조 조사'(2026-11-23 03:47–04:10Z 조사, Claude Code 2.1.330, 공식 문서 code.claude.com/docs/en/memory 와 /settings 를 그날 읽음).

### What exists

읽은 시각 2026-11-23T04:2xZ, 통합 tip `b3ba2b30d5`, 새 가지는 그 위에서 땄어요.

#### 낡은 규칙 여섯 곳(S1~S6)과 근거

| 곳 | 옛 글 | 지금 | 근거 |
| --- | --- | --- | --- |
| S1 원장 | "GitHub Issues are the work ledger", 클레임 댓글, `PLAN #<issue>`, PR 병합 | 로컬 카드(`moai card`)가 원장, PR 은 릴리스 때만, 착지는 `moai card land` | `docs/decisions-2026-11-08.md` 행 '작업 원장 = 로컬 카드'(2026-11-16); `moai card --help`, `moai card land --help`; `skills/moai/references/flows/next-flow.md`; `SPEC-20261121-card-land` |
| S2 론처 `-l\|-a` | "deprecated since #2164", 제거는 나중 카드 | 유지: 세션 이름과 접근 권한만 정하고 역할은 세션 안에서 입력 | `src/session.rs` 21-24행, `SPEC-20261118-launcher-agent-entry`(사용자 결정 2026-11-18), `moai --help` 의 `moai cc\|glm -l\|-a` |
| S3 모드 플래그 | "`/moai --mode-leader\|--mode-agent` 는 스킬이 모드를 고르는 유일한 철자, 두 모드 워크플로" | 스킬은 대시 옵션을 요청 글로 읽고(`skills/moai/SKILL.md` 라우팅 글), `skills/moai/workflows/` 에 모드 워크플로가 없어요(`work.md` 가 반복). 론처와 `lane_loop` 은 아직 `/moai --mode-agent` 줄을 찍어요 | `ls skills/moai/workflows`, `src/session.rs:47-48`, `src/lane_loop.rs:177`, `SPEC-20261122-next-flow-skills-agents` REQ-2(검증: `tests/claim_verify_cli.rs` 의 '지웠다'는 주석) |
| S4 호스트 | "Hosts: Claude Code and Codex", 관리 구역 `hosts: claude, codex` | Claude Code(GLM 은 그 실행기 경유). Codex·OMP 철회 | 사용자 결정 2026-11-21, `SPEC-20261121-remove-codex-support`, `src/session.rs` 머리말, `SPEC-20261121-codex-launcher-and-models-removed`; 프로필 `.moai/project.yaml` 의 `hosts` 가 아직 `codex` 를 가졌어요 |
| S5 경로·워크트리 | 프로젝트 상태 `.moai-cli/`, Go 의 `.moai/` 를 건드리지 말 것, 새 워크트리 `~/.moai/worktrees/<project>-<id>`, `.moai/` 는 무시 | 프로젝트 상태는 `.moai/`(추적 333개: 스펙·프로필·평가), `.moai-cli/` 는 있는 곳에서만 읽음, 새 워크트리 `.claude/worktrees/<name>`, 태그 `moai-adk-v{head}.{yearweek}.{build}`, 바이너리 `moai` | `git ls-files .moai`(333), `src/state_root.rs`, `SPEC-20261121-path-root-moai-default`, `SPEC-20261122-worktree-place`, `docs/release-checklist.md` 18행(결정 D6), `git worktree list` |
| S6 Python 테스트 | 구현 지도에 `tests/*_test.py`(Python 픽스처), `adapters/claude/`(임시 Python) | 테스트는 Rust만; `moai check affected` 는 Python 테스트를 고르지도 돌리지도 않고, 남은 Python 테스트 파일 56개는 한 카드가 걷어내요; `adapters/` 는 없어요 | `moai check --help`, `git ls-files tests \| rg _test.py$`(56), `ls adapters`(없음), 사용자 규칙 2026-11-21·2026-11-23 |

그 밖에 바로잡은 곳: 출하 글에서 하나로 합친 HeadVer 절 두 개; 태그 접두사 `moai-cli-v`(옛)와 `moai-adk-v`(결정 D6); 감사 배정(레인이 자기 `moai-audit` 에이전트를 시작)과 `audit.md` 의 층(tier) 표; 'Codex 창에 붙여 넣기' 문장; 심사 상태 `moai/audit` 의 서술은 릴리스 PR 에서만 해당.

#### 지침 파일 로딩(측정)

공식 문서는 하위 폴더에서 위쪽 폴더의 `CLAUDE.md` 를 모두 읽는다고 해요. 그래서 `.claude/worktrees/<name>` 아래 레인이 상위(주 체크아웃)의 `CLAUDE.md` 와 `CLAUDE.local.md` 를 겹쳐 읽는지가 질문이었어요. 2026-11-22T01:18Z 에 Claude Code 2.1.330 으로 잰 결과(`~/.moai/reports/moai-adk/claude-md-slice0-probe-2026-11-22.md`, 탐침 C1·C3: `git worktree add` 로 만든 `.claude/worktrees/w1` 과 `claude --worktree`): 모델에 가는 요청 본문에서 워크트리 자신의 추적 `CLAUDE.md` 와 규칙은 보이고, 주 체크아웃의 편집된 `CLAUDE.md` 는 보이지 않았고(겹쳐 읽지 않음), 주 체크아웃의 무시 파일 `CLAUDE.local.md` 는 보였어요. 워크트리가 체크아웃 폴더 밖(C2)이면 `CLAUDE.local.md` 도 안 보여요. 그래서 `CLAUDE.md` 는 중복 로드 비용이 없고, `CLAUDE.local.md` 는 `.claude/worktrees/` 아래 레인에게 한 번 보여요. 이 레인에서는 다시 재지 않았고(탐침 도구가 사라졌고 Claude Code 판본이 같아요: 2.1.330), 잰 것은 세션 시작 시점뿐이에요(탐침 5번).

#### 이 파일을 읽는 테스트

`tests/repo_claude_md.rs`(예산 44 KiB, 구역 제목, 문장 고정), `tests/recurring_problems_rule.rs`(`## Leader supervision`, `## Rule provenance`, `## Machine notes` 구역과 행), `tests/lane_next_rules.rs`(`PLAN #<issue>` 문장), `tests/claim_verify_cli.rs`(`1. Claim.`·`2. Worktree.` 줄). 저장소 전체 검사(`tools/repo_scans.json`, `tests/claude_python_adapter_removed.rs` 의 낱말 검사, `tests/next_flow_skill_scan.rs`)도 이 파일을 지나가요; `moai check affected --list` 가 고르고 READY 영수증이 결과를 보여요.

## Non-goals

- 전역 `~/.claude/CLAUDE.md` 와 `rules/moai.md` 를 쓰는 일, 제품이 만드는 `CLAUDE.md` 관리 구역의 생성기 코드(`src/agents.rs`)를 바꾸는 일.
- 스킬·에이전트·`audit.md` 의 글을 바꾸는 일(이 SPEC 은 그것들을 가리키기만 해요).
- `.claude/rules/*.md`(경로별 규칙)를 새로 쓰는 일: `.gitignore` 의 `.claude/` 무시 범위를 먼저 좁혀야 해서 `SPEC-20261122-claude-md-text-and-ignore` 가 맡아요.
- 실제 `CLAUDE.local.md` 를 쓰는 일: 무시 파일이라 이 가지에 넣지 않아요. 초안만 이 폴더에 둬요.
- `docs/design-review/autonomy-modes-spec.md` 처럼 이슈 시대 문서를 고치는 일.

## Requirements

- REQ-1: The authored `CLAUDE.md` shall stay within 200 lines including its managed section and hold only facts a team sharing this repository needs.
- REQ-2: The rule table of this SPEC shall list every rule of `CLAUDE.md` as it stood at `b3ba2b30d5` with the disposition kept, moved or discarded.
- REQ-3 (REQ-2): IF a row is kept or moved, THEN its destination file shall contain the row's anchor text verbatim.
- REQ-4 (REQ-2): IF a row is discarded, THEN the row shall name the decision or the fact that retires the rule.
- REQ-5: The rules named in S1 to S6 shall state what the code and the recorded decisions state today.
- REQ-6: The provenance table shall live in `docs/claude-md-rule-provenance.md` with the rows for the recurring-problems rule and the Korean wording rule.
- REQ-7: The detail of the lane loop and of leader supervision shall live in `docs/lane-rules.md` and `docs/leader-rules.md`, each named by a pointer line in `CLAUDE.md`.
- REQ-8: The draft `claude-local-draft.md` of this SPEC shall hold machine and personal facts only and none of the team rules.
- REQ-9: The tests that read `CLAUDE.md` shall pin the managed section, the line limit, the gate, HOLD and secret sentences, the six corrected facts and the absence of card numbers and personal lines.
- REQ-10: The decisions log shall record the user decisions of 2026-11-22 (`AGENTS.md` removed) and 2026-11-23 (this rewrite) with time, scope and how to undo them.
- REQ-11: The managed section and the `hosts` and `purpose` values of `.moai/project.yaml` shall name Claude Code only.
- REQ-12: This SPEC shall record the measured load of `CLAUDE.md` and `CLAUDE.local.md` for a lane under `.claude/worktrees/`.

## Acceptance

- AC-1 (REQ-1, REQ-6, REQ-7, REQ-9, REQ-11): `CLAUDE.md` has at most 200 lines, one managed section at the end whose `hosts` equal the profile's, no personal line, no card number, no provenance table, the pinned sentences, and a pointer to each of the two docs.
  verify: `cargo test --locked --offline --test repo_claude_md`
- AC-2 (REQ-2, REQ-3, REQ-4): every row of `rule-table.md` has a disposition, a destination and anchor or a basis, the ids are unique and continuous, and every anchor is found in its destination file.
  verify: `cargo test --locked --offline --test repo_claude_md rule_table`
- AC-3 (REQ-5): `CLAUDE.md` names local cards as the ledger, `-l|-a` as kept, `/moai --mode-agent` as request text, Claude Code and GLM as the hosts, `.claude/worktrees/` and `.moai/` as the places and Rust as the only test language.
  verify: `cargo test --locked --offline --test repo_claude_md stale_`
  verify: `cargo test --locked --offline --test lane_next_rules`
- AC-4 (REQ-6): the provenance file holds the recurring-problems and Korean rows with their sources and no card number.
  verify: `cargo test --locked --offline --test recurring_problems_rule`
- AC-5 (REQ-8): the local draft holds none of the team-rule anchors and names this machine's facts.
  verify: `cargo test --locked --offline --test repo_claude_md local_draft`
- AC-6 (REQ-10): the decisions log has the rows of 2026-11-22 and 2026-11-23.
  verify: `cargo test --locked --offline --test repo_claude_md decisions_log`
- AC-7 (REQ-12): the load measurement is in this SPEC with its source, version and limits (read above).
  verify: `cargo test --locked --offline --test repo_claude_md load_measurement`
  verify: `moai spec check --all --project .`
- AC-8: neighbouring checks on the changed files.
  verify: `moai check affected <base>..<head>`

## Scope

- write: `CLAUDE.md`, `docs/lane-rules.md`, `docs/leader-rules.md`, `docs/claude-md-rule-provenance.md`, `docs/agents-md-maintenance.md`, `docs/decisions-2026-11-08.md`, `.moai/project.yaml`, `tests/repo_claude_md.rs`, `tests/recurring_problems_rule.rs`, `tests/lane_next_rules.rs`, `tests/claim_verify_cli.rs`, `tests/fixtures/state_folder_literals/allow.tsv`, `.moai/specs/SPEC-20261123-claude-md-rewrite/**`
- read-only: `skills/moai/**`, `src/**`, `.gitignore`
- forbidden: the real `CLAUDE.local.md`, the user's Claude config folder, the profile folder of the session, `.claude/` of the primary checkout

## Plan

1. SPEC first (this file, `rule-table.md`, `claude-local-draft.md`); `moai spec check`.
2. Tests first: rewrite `repo_claude_md.rs` and `recurring_problems_rule.rs` for the new structure, retire the two pins whose rules are discarded (`agents_md_names_the_plan_comment_as_part_of_the_design` in `lane_next_rules.rs`, `claim_order_docs_agree` in `claim_verify_cli.rs`; table rows R-03, R-13 to R-15), and run the new tests against the old `CLAUDE.md` to see them fail (RED).
3. Write the three docs, then `CLAUDE.md` (when its reservation is free: another card holds it while it adds one provenance row, and that row then moves to `docs/claude-md-rule-provenance.md`), then `.moai/project.yaml` hosts and purpose and the decisions log.
4. Run the tests again (GREEN), `moai check affected <base>..<head>` once, then the READY note with its receipt. `CLAUDE.md` is on the tier-2 floor list of `audit.md`, so the audit is the full one.

## Risks

- Another card holds `CLAUDE.md` while it adds one row to the provenance table of the old file. This card works on everything else first and takes `CLAUDE.md` when that reservation ends; the row then moves to `docs/claude-md-rule-provenance.md`. A conflict at landing is the lander's to merge.
- Every session reads this file, so a wrongly dropped rule changes behaviour silently. The rule table and its anchor test are the guard; the earlier text stays in git (`b3ba2b30d5`).
- Other texts may point at sections of the old file (`## Agent loop`, `## Gate`, `Rule provenance`). The plan greps for them and repairs the pointers; a pointer that cannot be repaired is reported.
- The repository scans and wording tests read this file too; they decide at READY, through `moai check affected`, not by guess.
- `.moai/project.yaml` has an uncommitted change in the primary checkout (two detected-tool lines). This card changes only `hosts` and `purpose`; the two hunks do not overlap.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | Detail the session still must read goes to two `docs/` files with a short rule and a pointer in `CLAUDE.md`; no `@` import. | An import loads the file at launch like inline text, so it does not reduce cost (official memory page, read 2026-11-23); the card allows either. | Inline the two files again. |
| D-2 | The provenance table moves to `docs/claude-md-rule-provenance.md`, with retired rows marked as retired. | Official guidance: facts the code or history already holds do not belong in the file. | Move the table back. |
| D-3 | The managed section is edited by hand, as the user decided on 2026-11-20 (log row 관리 구역 손 관리: the next generation apply uses `skip-conflicts`), and the profile's `hosts` and `purpose` change so that a later regeneration says the same. | `moai agents-plan` on a copy answered `conflict` ("managed block edited after baseline"), as that decision expects. | Restore the two profile values and the two lines. |
| D-4 | Commit messages stay Korean in `CLAUDE.md`. `.moai/project.yaml` says `language.commit_messages: "en"`; the user decision of 2026-11-23 (amending the writing-language rule) and the commit history say Korean. The profile is left as it is and reported. | Core rule 21: code and recorded decisions win; the profile value is the user's own setting, not changed here. | Change the profile value or the sentence. |
| D-5 | `CLAUDE.md` is leader-owned: a lane changes it only through a card whose write scope names it. | The card of this SPEC and the earlier ownership line; the reservation guard (`moai card check`) enforces it. | Drop the sentence. |
| D-6 | Card numbers appear in no SPEC, commit message or document written by this change. | User rule of 2026-11-20 recorded in the provenance table (card-number commit messages as a cause). | None. |
| D-7 | Where a rule already has its full description in `skills/moai/references/audit.md` or `flows/next-flow.md`, `CLAUDE.md` and the docs point to it and copy only the one-line rule. | One full description per rule. | Copy the text. |
| D-8 | `CLAUDE.md` names `NeXT` as the integration branch and keeps one line saying `v2026.11.0` is a recovery ref that stopped at `b3ba2b30d5`. | User decision of 2026-11-23 relayed by the leader (card note, 04:22:57Z); the local branch `NeXT` is at `1b775565be` and `.moai/project.yaml` `work_flow.integration_branch` is `NeXT` on that branch. | Name the earlier branch again. |

## Result

Status: documents, tests and the profile changed; the SPEC front matter is not changed here (approval is recorded by the user's pick, not by a lane). One step is still open: the single merge of the integration branch `NeXT` after the card that also edited `CLAUDE.md` has landed, then the checks again.

Measured at head `9acf2dc3e7` on base `b3ba2b30d5` (`moai check affected b3ba2b30d5..9acf2dc3e7 --max 4`, exit 0): `diff_check=0 fmt=0 clippy=0 test=0 result=PASS`, 2461 tests passed in 51 targets, 0 skipped. Lines now: `CLAUDE.md` 119 (was 235), `docs/lane-rules.md` 50, `docs/leader-rules.md` 49, `docs/claude-md-rule-provenance.md` 57; `rule-table.md` has 113 rows.

- AC-1: met — `repo_claude_md` passed in that run: at most 200 lines (119), one managed section at the end with `hosts: claude.` equal to the profile, no personal line, no card number, no provenance table, the pinned gate, HOLD, secret and audit sentences, and a pointer to each of the three docs.
- AC-2: met — `repo_claude_md rule_table_has_a_disposition_for_every_rule_and_every_anchor_is_found` passed: 113 rows, ids R-01 to R-113, every kept or moved anchor found in its destination, every discarded row names its basis.
- AC-3: met — the `stale_` tests (ledger, launcher `-l|-a`, mode flags, hosts, paths and worktrees, integration branch, Python tests) and `lane_next_rules` passed. The integration branch is `NeXT` by the user decision of 2026-11-23 (decision D-8).
- AC-4: met — `recurring_problems_rule` passed (5 tests): the recurring-problems rule sits in Standing principles, its row and the Korean wording row sit in `docs/claude-md-rule-provenance.md` with their sources and no card number.
- AC-5: met — `repo_claude_md local_draft_holds_machine_facts_and_no_team_rule` passed; the draft names this machine's facts (Claude Code 2.1.330, `CLAUDE_CONFIG_DIR`, `sccache`, the lane cargo home) and repeats no kept team rule.
- AC-6: met — `repo_claude_md decisions_log_records_the_agents_md_removal_and_this_rewrite` passed; the log has the rows of 2026-11-22 (`AGENTS.md` removed), 2026-11-23 (this rewrite) and 2026-11-23 (integration branch `NeXT`), plus one backfill row that copies the flow decisions of 2026-11-20 to 2026-11-22 from their SPECs (the leader confirms it).
- AC-7: met — `load_measurement_is_recorded_in_the_spec` passed; the measurement is above with its source, version and limits. `moai spec check --all --project .` reported 0 errors (19 warnings in other SPECs, none in this one).
- AC-8: met — the `moai check affected` run above. One scan needed a change in the same card: `state_folder_literal_scan` listed the two new tests, so `tests/fixtures/state_folder_literals/allow.tsv` got two lines (repo_claude_md 2/9, recurring_problems_rule 0/1).

Not run: no test on Windows or Linux; the two git checks of `repo_claude_md` ran only in the worktree, not in a copy outside git. `.moai/project.yaml` still says `language.commit_messages: "en"` while `CLAUDE.md` and the history say Korean (decision D-4); it is left for the user.
