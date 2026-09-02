---
id: SPEC-LANE-PUSH-BATCH-001
title: "Implementation plan — lane push prohibition; lead-side batched develop push"
version: "0.1.0"
created: 2026-09-02
author: manager-spec
---

# SPEC-LANE-PUSH-BATCH-001 — plan.md

## §A Context

- **Worktree**: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t430` — verify with
  `git rev-parse --show-toplevel` before any edit. Branch `WT-lead-batch-push`, base tree
  `ad272be20abff9e4f3b1b363fce3e48dac4c5132` (= origin/develop, the adjudicated canonical
  copy). Re-read HEAD immediately before committing.
- **SPEC artifacts**: `.moai/specs/SPEC-LANE-PUSH-BATCH-001/{spec,plan,acceptance,progress}.md`.
- **Adjudication (inherited, settled)**: canonical copy = `origin/develop @ ad272be20`. The
  primary checkout's `CLAUDE.local.md` working-tree copy is stale (zero unique lines,
  measured). Do not reconcile it; edit the worktree copies only.
- **Scope (amended 2026-09-02)**: 8 point-locations across 5 files — `CLAUDE.local.md`
  348+364; `gitflow-lane-protocol.md` 69 (+71 retry scenario) + 81;
  `git-workflow-doctrine.md` 103 + 329; `repo-local-pr-policy.md` 12;
  `git-local-workflow-doctrine.md` 150. Enumeration basis: orchestrator wide sweep
  (integration merge co-mentioned with push across `.moai/docs/`, `.claude/rules/local/`,
  `CLAUDE.local.md`), corroborated at plan-phase — a `grep -rln` over the push-phrase class
  returns exactly the 5 in-scope files, zero further hits.
- **Measured evidence to cite verbatim in the doctrine text** (REQ-LPB-003): 3 cards
  (t336, t372, t413), 22 commits, ONE push `09bf452c0..ad272be20`, Vercel builds 3→1.

### §A.5 PRESERVE list (must survive this SPEC untouched)

| Surface | Anchor | Rule |
|---|---|---|
| `CLAUDE.local.md` ~line 349 | `WT 브랜치 push·CI 직접 요청 금지` (운영자 지시 2026-09-01) | byte-identical |
| `CLAUDE.local.md` ~line 347 | `승인이 아니다` (window only after lead designation) | present, count 1 |
| `CLAUDE.local.md` ~line 351 | `sync는 병합` (sync completes BEFORE the merge) | present, count 1 |
| `CLAUDE.local.md` ~line 350 | `원격 머지가 확인되기 전까지 폐기하지 않는다` (no disposal before remote landing) | present, count 1 |
| `gitflow-lane-protocol.md` | §3 (직렬화), §5 (충돌), §8 (검증은 레인-로컬), §9 (rc 빌드), §10 (릴리스), §11 (develop 갱신) | byte-identical; `^## [0-9]` count stays 11 |
| `.claude/rules/local/repo-local-pr-policy.md` line 12 | "Remote CI on `origin/develop` is the verdict surface" sentence | kept while the "; lanes push" clause is replaced |
| `.moai/docs/git-workflow-doctrine.md` line 103 | §2·§4 protocol pointer | kept while the push clause is reworded |
| `.moai/docs/git-local-workflow-doctrine.md` line 150 | [RETIRED 2026-08-27] markers + `enforce_admins` note | kept while the push clause is reworded |
| `gitflow-lane-protocol.md` line 8 | 2026-08-27 transition provenance note | untouched (confirmed-leave) |
| `gitflow-lane-protocol.md` line 29 | §2 delegation sentence (the §2 EXCLUDED exception covers its push step) | untouched apart from the added exception |
| `gitflow-lane-protocol.md` line 34 | factual push-target statement | untouched (confirmed-leave) |
| `gitflow-lane-protocol.md` line 58 | §3 HEAD re-read rule (applies to the lead's push too) | untouched (confirmed-leave) |
| `.moai/docs/git-local-workflow-doctrine.md` line 157 | §23.9(a) — push-neutral as written | untouched (confirmed-leave) |
| `.claude/skills/moai/workflows/sync/delivery.md` | entire file | untouched (Out of Scope) |
| `internal/template/templates/**` | entire tree | untouched; no `make build` |

Note: `gitflow-lane-protocol.md` §4's retained sentence "원격 CI(`origin/develop`)가 통합
판정의 주체다" (line 73) stays — the verdict surface is unchanged; only the push actor and
cadence change.

## §B Known Issues

- **B-anchor-whitelist (AC-1 subtlety, updated per plan-audit D3)**: `git push origin
  develop` appears 3× in `CLAUDE.local.md` today — lines 348 (chain), 349 (the
  KEEP-VERBATIM 2026-09-01 bullet's historical parenthetical), 364 (운영 절차 code block).
  Only 348 and 364 are lane-facing removal targets; 349 is preserved verbatim by operator
  instruction. AC-LPB-001's green check uses: the chain-arrow form → 0 (G1); the bare-line
  form → EXACTLY 1 (G2 — the single sanctioned site is the lead batch-push block's runnable
  code line; NO formatting ban is imposed; the count is 1 at both RED and GREEN, the site
  flip carried by G1 + G3's whitelist); and the chain-scoped positive anchor G28
  (`^- 창을 받으면: .*병합 SHA` ≥1) that fails the evasion mutant ending the chain at
  release without the merge-SHA report.
- **B-grep-exit-semantics**: `grep -c` exits 0 when count > 0 and 1 when count is 0. For
  absence checks the PASSING signal is exit 1 + stdout `0` — do not read exit 1 as failure.
- **B-enumeration (amendment 2026-09-02)**: the orchestrator's wide sweep (integration
  merge co-mentioned with push across `.moai/docs/`, `.claude/rules/local/`,
  `CLAUDE.local.md`) closed the enumeration at 8 point-locations across 5 files; the
  plan-phase corroboration sweep (`grep -rln` over the push-phrase class) returns exactly
  the 5 in-scope files. The two formerly-reported siblings are now IN scope:
  `repo-local-pr-policy.md:12` and `git-local-workflow-doctrine.md:150`, plus
  `git-workflow-doctrine.md:103` as a second doctrine location (REQ-LPB-007 covers both
  doctrine points).
- **B6 (lint)**: spec.md carries `### Out of Scope — <topic>` H3 subsections with `-`
  bullets (OutOfScopeRule).
- **B11**: this is a subagent task — no AskUserQuestion; blockers go back to the
  orchestrator as structured reports.
- **B-run-commit**: run-phase commits on the card branch carry the card id (traceability
  carrier 2 of 3); e.g. `docs(t430): SPEC-LANE-PUSH-BATCH-001 M1 — lane push prohibition,
  lead batch push`.

## §C Pre-flight

```bash
git rev-parse --show-toplevel          # must print the t430 worktree path
git rev-parse --short HEAD             # expect ad272be20 (or later; re-pin if moved)
grep -n "git push origin develop" CLAUDE.local.md              # expect 3 hits (348/349/364)
grep -n "병합 후 레인이 직접 올린다" .claude/rules/local/gitflow-lane-protocol.md   # expect 1 hit (69)
grep -n "git push origin develop" .moai/docs/git-workflow-doctrine.md              # expect 1 hit (329)
```

If any expectation differs from the acceptance.md RED cells, STOP and report the divergence
— do not improvise against a moved tree.

## §D Constraints

- Edit exactly five files, at only the named locations: `CLAUDE.local.md`,
  `.claude/rules/local/gitflow-lane-protocol.md`, `.moai/docs/git-workflow-doctrine.md`,
  `.claude/rules/local/repo-local-pr-policy.md` (line 12 only),
  `.moai/docs/git-local-workflow-doctrine.md` (line 150 only). Nothing else.
- No edits under `internal/template/templates/`; no `make build`.
- PRESERVE list (§A.5) is byte-identical where marked.
- Conventional Commits; card id `t430` in every commit message on the branch.
- Do not re-adjudicate the canonical copy; do not touch delivery.md; do not add tooling.

## §E Self-Verification

Docs-only card: the E2/E3/E4/E8 items (build/coverage/boundary/RED-test) do not apply.
Verification is the grep batch of acceptance.md §3, run as parallel read-only Bash calls,
reported per the attribution triple — (a) command, (b) verbatim output, (c) tree SHA of this
run. Each AC row cites its command's verbatim stdout and exit code; absence checks report
`grep -c` stdout `0` + exit 1 as the PASSING signal. RED cells were measured at tree
`ad272be20` and are recorded with their verbatim outputs in acceptance.md §4.

## §F Milestones

Single milestone (Tier S, single-milestone card). Edit steps are ordered semantic-first —
the decisions most likely to receive review attention lead; mechanical line edits follow.

### M1 — Lane push prohibition + lead batch-push doctrine (all three files)

1. **Semantic core — `gitflow-lane-protocol.md` §4 rewrite** (REQ-LPB-005): replace
   "병합 후 레인이 직접 올린다: `git push origin develop`." with lead-actor wording; delete
   the lane-side rejected-push-retry paragraph ("거부되면 ... 절대 force 하지 않는다.");
   write the lead-side batch push + remote-landing verification; KEEP the verdict-surface
   sentence (line 73).
2. **Semantic core — `gitflow-lane-protocol.md` §2 exception** (REQ-LPB-004): add the
   repo-local exception naming delivery.md Step 3.2 step 6 as EXCLUDED in this repo (lead's
   batched push, 2026-09-02); literal token `EXCLUDED` required (AC anchor).
3. **Semantic core — `gitflow-lane-protocol.md` §7 lead duty** (REQ-LPB-006): collect SHAs
   from lane completion reports → batch decision → single push → verify remote landing →
   card done + disposal approval. Token `일괄` required (AC anchor).
4. **`gitflow-lane-protocol.md` §6 rewording** (REQ-LPB-005): "병합·push를 마치면" → lane
   wording without push; disposal gate wording kept, re-anchored to the lead's push landing.
5. **`CLAUDE.local.md` §4.1 lane sites** (REQ-LPB-002): (a) 창을 받으면 chain — remove
   `git push origin develop`, chain ends `moai integration release` → report local merge
   SHA; the chain line MUST contain the literal `병합 SHA` (chain-scoped anchor G28:
   `^- 창을 받으면: .*병합 SHA`);
   (b) 운영 절차 code block — remove the push line, add comment (push OUTSIDE the window, by
   the lead, batched); (c) 규율 2 — push actor = lead, batched (2026-09-02); (d) 규율 4 —
   window chain `acquire → merge → release`, push excluded.
6. **`CLAUDE.local.md` §4.1 NEW lead batch-push block** (REQ-LPB-003): insert after the
   운영 절차 code block, before "**로컬 CI를 두지 않는 이유**". Three elements (collection
   basis / push-time decision / remote-landing verification → card done + disposal approval)
   + the measured-evidence sentence (t336+t372+t413, 22 commits, `09bf452c0..ad272be20`,
   Vercel 3→1). Lane completion-report field list (line 346) gains the local merge SHA.
   The block's runnable code form contains `git push origin develop` as a bare line — the
   single G2 sanctioned site (exactly one file-wide).
7. **`.moai/docs/git-workflow-doctrine.md` (REQ-LPB-007)**: (a) §18.8 step 1 — drop
   `git push origin develop` from the comment; reference the lead batch (CLAUDE.local.md
   §4.1); token `리드 일괄` required (AC anchor). (b) §18.3 merge-strategy callout (line
   103) — reword "합친 뒤 `origin/develop`에 push한다" to lead-batch wording; keep the
   `gitflow-lane-protocol.md` §2·§4 pointer.
8. **`.moai/docs/git-local-workflow-doctrine.md` §23.7 (REQ-LPB-007)**: reword "합친 뒤
   push한다(`origin/develop` CI가 판정)" (line 150) to lead-batch wording; [RETIRED]
   markers and `enforce_admins` note untouched; token `일괄` required (AC anchor).
9. **`.claude/rules/local/repo-local-pr-policy.md` card-workflow bullet (REQ-LPB-004)**:
   replace "; lanes push `origin/develop`" (line 12) with lead-batch wording; keep the
   verdict-surface sentence; token `일괄` required (AC anchor).
10. **Verify**: run the acceptance.md §3 grep batch; confirm PRESERVE anchors unchanged;
    record results in progress.md §E.2.

**Canonical grep tokens (MUST appear verbatim — AC evidence anchors):** `일괄` (CL ≥1, LP
≥1, RPP ≥1, LWD ≥1), `병합 SHA` (CL ≥1), `원격 착지` (CL ≥1), `09bf452c0` (CL ≥1),
`EXCLUDED` (LP ≥1), `리드 일괄` (DOC ≥1), `verdict surface` (RPP ≥1, kept),
`gitflow-lane-protocol.md` (DOC ≥1, pointer kept), `병합 후 레인이 직접 올린다` (LP → 0),
`거부되면` (LP → 0), `병합·push` (LP → 0), `lanes push` (RPP → 0), `합친 뒤 push한다`
(LWD → 0), `에 push한다` (DOC → 0), `git push origin develop` (CL chain-arrow form → 0,
CL bare-line form → EXACTLY 1 — the lead batch-push block's code line; DOC → 0), `^- 창을
받으면: .*병합 SHA` (CL ≥1 — chain-scoped positive anchor). (CL = CLAUDE.local.md, LP = gitflow-lane-protocol.md,
DOC = git-workflow-doctrine.md, RPP = repo-local-pr-policy.md, LWD =
git-local-workflow-doctrine.md.)

## §G Anti-Patterns

- Do NOT edit delivery.md "while you're in there" — the §2 exception is the mechanism.
- Do NOT rewrite the 2026-09-01 [HARD] bullet; its historical parenthetical stays.
- Confirmed-leave lines stay untouched (plan §A.5): protocol lines 8/29/34/58, LWD line 157
  (§23.9(a)), and the kept sentences inside the amended lines (verdict-surface sentence,
  §2·§4 pointer, [RETIRED] markers, `enforce_admins` note).
- Do NOT convert the batch push into tooling (no CLI subcommand, no hook).
- Do NOT touch PRESERVE rows; do not run `make build`; do not sweep-stage commits.

## §H Cross-References

- spec.md §5 Out of Scope — delivery.md, primary working tree, tooling.
- `.claude/rules/local/repo-local-pr-policy.md` / `.moai/docs/git-local-workflow-doctrine.md`
  §23.7 — the two amendment surfaces carrying the lead-batch wording.
- acceptance.md §4 evidence ledger — RED cells pinned to tree `ad272be20`.
- `CLAUDE.local.md` §4.1 — the model this SPEC amends (2026-08-29 chain + 2026-09-01
  prohibition + this 2026-09-02 batch directive).
- `.claude/rules/local/gitflow-lane-protocol.md` §11 — develop refresh keyed to
  `origin/develop` (unchanged; it now moves only on the lead's batched push).
- `.claude/skills/moai/workflows/sync/delivery.md` Step 3.2 — the distributed procedure the
  §2 exception carves.
