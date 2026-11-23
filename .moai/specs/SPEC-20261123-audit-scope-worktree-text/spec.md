---
id: SPEC-20261123-audit-scope-worktree-text
title: "The audit agent cleans up and checks only inside the audit worktree named in scope, and no text says it runs in its own isolated worktree"
status: draft
revision: 1
created: 2026-11-23
---

# The audit agent cleans up and checks only inside the audit worktree named in scope, and no text says it runs in its own isolated worktree

## Intent

감사 규칙 글(SPEC-20261123-audit-rule-text)이 착지하면서 감사 에이전트의 `isolation` 이 `none` 이 됐어요. 에이전트는 이제 자기 워크트리를 갖지 않고, 착지자가 만든 감사 워크트리(브리프의 `scope`)를 읽어요. 그런데 몇 곳이 옛 말을 그대로 써요:

- `agents/audit.md` 의 정리 규칙과 검증 줄이 "your own worktree" 라고 적어요. 에이전트의 셸은 부모(착지자) 폴더에서 시작하므로, 상대 경로 `rm -rf target/debug` 나 검사가 부모 트리에서 돌 수 있어요.
- `skills/moai/workflows/audit.md` 와 `skills/moai/workflows/run.md` 의 정리 줄이 같은 말을 써요. `run.md` 의 그 줄은 구현 에이전트(`moai-impl`)에도 해당해요.
- `skills/moai/workflows/update.md` 와 `src/worktree_include.rs` 가 `moai-audit` 도 `isolation: worktree` 로 Claude Code 가 만든 워크트리를 쓴다고 적어요. 지금은 착지자가 평범한 git 으로 만든 워크트리를 읽어요.
- 등급 줄의 핵심 낱말을 지워도 테스트가 모르고(네 글이 서로 같기만 하면 통과), `Isolation:` 줄은 고정돼 있지 않아요.
- 등급 줄 안의 "the floor list below" 는 아래에 목록이 없는 세 글(`agents/audit.md`, `workflows/run.md`, `workflows/audit.md`)에서 가리키는 곳이 없어요.

이 SPEC 은 글만 고쳐요. 코드와 등급 줄 자체는 그대로예요.

### 출처 (2026-11-23 에 이 저장소의 팁 `b20f525f91` 에서 읽은 것)

- `agents/audit.md` 전체, `skills/moai/workflows/audit.md` 98~106행, `skills/moai/workflows/run.md` 36~50·125~140행, `skills/moai/workflows/update.md` 46행, `src/worktree_include.rs` 1~10행, `skills/moai/references/audit.md` 의 "Audit tiers".
- `tests/worktree_include_guard.rs` 가 `update.md` 의 `moai-impl` and `moai-audit` 낱말을 고정해요.
- Claude Code 의 `EnterWorktree` 도구 설명: 고정된 에이전트도 `path` 로 `.claude/worktrees/` 아래 기존 워크트리로 들어갈 수 있고 그 전환은 그 에이전트에만 걸려요. 이 SPEC 은 그 도구를 이름 붙이지 않고 "그 안에서 시작하라" 만 적어요.
- `.worktreeinclude` 가 복사되는 대상은 `src/worktree_include.rs` 가 인용한 문서(https://code.claude.com/docs/en/worktrees.md, 2026-11-13 읽음)의 "Claude Code 가 git 으로 만드는 워크트리" 예요. 이번에 그 문서를 다시 읽지 않았어요(미확인으로 둬요).

## Non-goals

- 감사 에이전트의 `isolation` 값 변경, 등급 줄 네 개의 문구 변경, 감사 계약(`references/audit.md`)의 문구 변경.
- 구현 에이전트(`moai-impl`)의 격리 방식 변경.
- 착지자가 감사 워크트리를 만드는 방법(평범한 git)과 `moai worktree-gc-apply` 의 변경.

## Requirements

- REQ-1: The cleanup rule and the verification line of `agents/audit.md` shall name the audit worktree named in `scope` as the only place the agent deletes build output and say that its shell may start in the parent's folder.
- REQ-2: The first step of `agents/audit.md` shall tell the agent to run every command, the checks, the builds, the status read and the cleanup, inside the audit worktree named in `scope`.
- REQ-3: The `Isolation:` line and the `Cleanup` line of `skills/moai/workflows/audit.md` shall say that the agent reads and cleans the audit worktree named in `scope` and runs every command inside it.
- REQ-4: The `Cleanup` line of `skills/moai/workflows/run.md` shall name the worktree the agent works in, never "its own worktree".
- REQ-5: The text of `skills/moai/workflows/update.md` and the comment of `src/worktree_include.rs` shall say that only the subagent worktrees of `moai-impl` are made by Claude Code and that the audit worktree is made by the lander with plain git.
- REQ-6: The four tier lines shall keep their key words, checked in the audit contract and in each of the three texts that carry them.
- REQ-7: `skills/moai/workflows/run.md` and `skills/moai/workflows/audit.md` shall point to the floor list in the audit contract's section "Audit tiers" next to the tier lines.
- REQ-8: A scan test in `tests/next_flow_skill_scan.rs` shall fail when any of REQ-1 to REQ-7 is undone.

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-3, REQ-4): `agents/audit.md`, `workflows/audit.md` and `workflows/run.md` hold no "own worktree" in a cleanup or isolation line and carry the scope-worktree words · verify: `cargo test --locked --offline --test next_flow_skill_scan refs_audit_cleanup_names_the_scope_worktree`
- AC-2 (REQ-3): the `Isolation:` line of `workflows/audit.md` holds its pinned words · verify: `cargo test --locked --offline --test next_flow_skill_scan refs_audit_isolation_line_is_pinned`
- AC-3 (REQ-5): `update.md` and `worktree_include.rs` no longer say `moai-audit` uses an isolated worktree and name plain git · verify: `cargo test --locked --offline --test next_flow_skill_scan refs_no_text_says_the_audit_agent_has_an_isolated_worktree` and `cargo test --locked --offline --test worktree_include_guard`
- AC-4 (REQ-6): a changed tier line with one key word removed in all four texts fails the key-word test · verify: `cargo test --locked --offline --test next_flow_skill_scan refs_audit_tier_lines_keep_their_key_words`
- AC-5 (REQ-7): both workflows name the contract's floor list next to the tier lines · verify: `cargo test --locked --offline --test next_flow_skill_scan refs_audit_floor_list_is_pointed_to`
- AC-6 (REQ-8): every test that read a rewritten phrase stays green · verify: `cargo test --locked --offline --test next_flow_skill_scan --test agent_defs_package --test worktree_include_guard --test lane_rules_2401`

## Scope

- write: `.moai/specs/SPEC-20261123-audit-scope-worktree-text/**`, `agents/audit.md`, `skills/moai/workflows/audit.md`, `skills/moai/workflows/run.md`, `skills/moai/workflows/update.md`, `src/worktree_include.rs`, `tests/next_flow_skill_scan.rs`, `tests/worktree_include_guard.rs`, `tests/agent_defs_package.rs`, `skills/moai/references/source-manifest.json`
- read-only: `skills/moai/references/audit.md`, `src/gate.rs`
- forbidden: the audit contract's wording, the four tier lines, any code outside the one comment of `src/worktree_include.rs`

## Plan

1. The SPEC.
2. RED tests in `tests/next_flow_skill_scan.rs` for AC-1, AC-2, AC-3, AC-4, AC-5.
3. The text edits; the one phrase pinned by `tests/worktree_include_guard.rs` changes with the text.
4. The source manifest is rewritten with `tools/source_manifest.py write` (a generator, not a test).
5. Targeted tests (AC-6), then one `check affected`, then READY with head, base and the receipt.

## Risks

- R-1: The installed agent and skill change only with the next `moai update`; until then the installed copies keep the old words.
- R-2: A host other than Claude Code may not let the agent start inside the worktree. The text then says to give each command the worktree as its folder; the existing `blocked` stop (worktree missing or HEAD not `head`) is unchanged.
- R-3: The `.worktreeinclude` claim about plain git rests on the 2026-11-13 read of the official page; it was not re-read for this change.

## Result

Status: SPEC draft (revision 1); the text slice is built. The SPEC carries no approval; the user approves it.

As built:
- `agents/audit.md`: the first step, the cleanup rule and the verification line name the audit worktree named in `scope`; the cleanup rule says the shell may start in the parent's folder where a relative delete would hit the parent's build output.
- `skills/moai/workflows/audit.md`: the `Isolation:` line says the agent runs every command inside the audit worktree (its shell starts in the lander's folder); the cleanup line names the audit worktree in `scope`; a pointer sentence names the contract's floor list.
- `skills/moai/workflows/run.md`: the cleanup line names the worktree the agent works in (its own where the host offers one, else this worktree); the same pointer sentence stands under the tier lines. It names the contract's section and no package path, because `tests/product_cli.rs` allows no package path in a workflow.
- `skills/moai/workflows/update.md` and `src/worktree_include.rs`: only `moai-impl` has a worktree Claude Code makes; the audit worktree is made by the lander with plain git (the claim rests on the 2026-11-13 read of the official page, not re-read).
- `tests/next_flow_skill_scan.rs`: five new tests for AC-1 to AC-5, red before the text edits (4 of 5 failed; the key-word test passes on the existing text and kills three planted key-word removals). `tests/worktree_include_guard.rs` and `tests/agent_defs_package.rs` follow the new phrases (the second keeps the old phrase for `moai-impl`). The source manifest was rewritten with `tools/source_manifest.py write`.
- AC-1 to AC-5 are met by those tests; AC-6 holds: `next_flow_skill_scan` 56, `agent_defs_package` 19, `lane_rules_2401` 3, `product_cli` 74 and `worktree_include_guard` 2 passed in this run.

Not run here: `moai check affected` (the READY note carries its receipt).
