---
id: SPEC-MIRROR-DOGFOOD-001
title: "Research — TestRuleTemplateMirrorDrift/worktree-integration.md RED root cause"
created: 2026-09-23
card: t1086
---

# Research — Mirror Drift on worktree-integration.md (card t1086)

All measurements below were re-executed first-hand on THIS worktree
(`/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1086`), HEAD `d323f68fd`
(= origin/develop), branch `WT-mirror-drift`, working tree clean at measurement time.

## 1. RED reproduction (observed, this tree)

RED-now cell in the four-element form (verification-completeness.md §2.1; baseline
attribution per verification-claim-integrity.md §2):

| Element | Value |
|---|---|
| Command (single invocation, as cited) | `go test ./internal/template/ -run 'TestRuleTemplateMirrorDrift' -count=1 -v` |
| Exit code | `1` |
| Verbatim stdout | FULL raw output, 45 lines, unelided — attached at `evidence/red-baseline-d323f68fd.txt` (stderr joined; the file is the byte-verbatim stream of the command above) |
| Tree SHA (pinned) | `d323f68fd` (branch `WT-mirror-drift`, working tree clean at capture) |

Key line from the attached raw output:

```
    rule_template_mirror_test.go:174: RULE_TEMPLATE_MIRROR_DRIFT: source file .claude/rules/moai/workflow/worktree-integration.md differs from its mirror at .../internal/template/templates/.claude/rules/moai/workflow/worktree-integration.md (source 61473 bytes, mirror 61638 bytes); run 'cp ...' and stage both files
--- FAIL: TestRuleTemplateMirrorDrift (0.00s)
    --- FAIL: TestRuleTemplateMirrorDrift/worktree-integration.md (0.00s)
    --- PASS: TestRuleTemplateMirrorDrift/hooks-system.md (0.00s)
    --- PASS: TestRuleTemplateMirrorDrift/frontend.md (0.00s)
    (... 6 further subtests PASS)
```

8 of 9 subtests PASS; only `worktree-integration.md` is RED. The failure message suggests
`cp local -> template` — that suggestion is the REJECTED direction (see §7).

## 2. Full diff — 2 hunks, 22 changed lines

`diff local template` shows exactly two hunks:

- **Hunk 1 (local lines 610-612)** — local copy carries measurement provenance:
  "at Claude Code **2.1.278** (2026-09-22; every disposition traces to a recorded probe;
  measured 2026-09-22 — card t1067)". Template copy carries the same sentence with the
  date and card id stripped: "(every disposition traces to a recorded probe)".
- **Hunk 2 (local lines 653-664)** — local copy carries a dogfood record pointer:
  "the AC-AEC-013 restatement in `SPEC-AUDIT-EXPORT-CLAUSE-001/acceptance.md:622-680` …
  census at `.moai/reports/t1067/census-20260922.md` (card t1067, measured 2026-09-22)".
  Template copy carries the same teaching content re-stated NEUTRALLY as a generic bash
  example (refused form `B=$(git merge-base develop HEAD)` + executable restatement with
  plain verbs) — no SPEC IDs, no dates, no card ids, no `.moai/reports/` paths.

**Census-pointer disposition:** `.moai/reports/t1067/census-20260922.md` was never
committed — absent from this tree's filesystem, git index, AND full history
(`git log --all` on the path: 0 rows; `git ls-files .moai/reports/t1067/`: 0 entries;
measured 2026-09-23 at `d323f68fd`). `.moai/reports/` is local-only/untracked. The path
is HISTORICAL: it resolves only inside card t1067's worktree-era session records. The
relocated record file must carry this disposition (spec.md REQ-MD-004, acceptance.md
AC-MD-004) so the path is never cited as live evidence.

## 3. Attribution (observed, this tree)

```
git log --oneline -S "card t1067" -- .claude/rules/moai/workflow/worktree-integration.md
dbe1a6941 docs(SPEC-AC-GUARD-001): M2 AC-command boundary map + authoring rule at the t287 section (card t1067)
```

The string entered ONLY the local copy (via SPEC-AC-GUARD-001 M2, card t1067). The same
probe against the template-side path returns empty — the template copy has NEVER carried
it. t1072 (`427ec4455`, 3 lines) and t1073 (`7795226c2`, line 52) edited both copies
properly (dual-copy commits) and are NOT part of this divergence.

## 4. Neutrality state of the template copy (observed, this tree)

`go test ./internal/template/ -run TestTemplateNoInternalContentLeak -count=1` →
`ok github.com/modu-ai/moai-adk/internal/template` (PASS). The template copy is
neutral-clean today; adopting it as the parity target preserves that state.

## 5. Durability constraint — why relocation, not deletion-without-home

- `.claude/rules/local/` exists, is git-tracked (5 dev-only protocol files:
  `ci-autofix-protocol.md`, `ci-watch-protocol.md`, `gitflow-lane-protocol.md`,
  `lifecycle-sync-gate.md`, `repo-local-pr-policy.md`), and has NO template mirror —
  the established durable home for dev-only content (CLAUDE.local.md §2).
- The managed root `.claude/rules/moai/` is wiped-and-redeployed by every
  `moai update` (CLAUDE.local.md §2.3; `CleanMoaiManagedPaths` at
  `internal/cli/update/deploy/deploy.go`). Dogfood records
  placed inside the managed root have no durability — which is exactly how this record
  ended up as a one-off in-place divergence instead of a tracked local file.

## 6. Settled purpose determination (the card's [HARD] first question)

| Surface | Classification | May carry | Enforced by |
|---|---|---|---|
| Template copy (`internal/template/templates/.claude/rules/moai/**`) | NEUTRAL distributed artifact | No internal card records / SPEC IDs / dates / report paths (neutrality classes C1-C8) | `TestTemplateNoInternalContentLeak` + CI template-neutrality guard |
| Local managed copy (`.claude/rules/moai/**`) | Deployed artifact, byte-parity obligation | Same content as template; NOT a durable home for dogfood records (wiped by `moai update`) | `TestRuleTemplateMirrorDrift` |
| Dogfood records | Dev-only durable content | Card ids, SPEC IDs, measurement dates, report paths | Lives in `.claude/rules/local/` (tracked, unmanaged, no mirror) |

The fix direction is the RESULT of this table: parity is restored by taking the
template's neutral form (template -> local), and the displaced record moves to
`.claude/rules/local/`.

## 7. Rejected directions (with rationale)

1. **REJECT `cp` local -> template** (what the test failure message suggests): ships
   `card t1067`, `2026-09-22`, `SPEC-AUDIT-EXPORT-CLAUSE-001`, and
   `.moai/reports/` paths into the distributed template — violates §25 neutrality
   classes C1-C8 and would trip the CI template-neutrality guard.
2. **REJECT relaxing the byte-parity gate** (allowlisting/excluding
   `worktree-integration.md`): the divergence is a one-off misplaced record with a
   proper home available; weakening the gate hides future REAL drift in this file;
   the gate's non-vacuity is already evidenced by the current RED itself.

## 8. Verification completeness note

Per `.claude/rules/moai/development/verification-completeness.md` §2 (two-cell
discipline): AC-001 (mirror test green) is release-blocking and its RED-now cell is
ALREADY observed on this tree (§1 above, HEAD `d323f68fd`). The green path cell names
M1 as the milestone that flips it. No criterion in this SPEC is vacuous-by-construction.
