---
id: SPEC-MIRROR-DOGFOOD-001
title: "Plan — mirror parity restore + dogfood record relocation"
created: 2026-09-23
card: t1086
---

# Plan — SPEC-MIRROR-DOGFOOD-001 (Tier M, card t1086)

## §A Context

Base: worktree `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1086`, branch
`WT-mirror-drift`, HEAD `d323f68fd` (= origin/develop, already absorbed — the
re-measurement starts from this state per the card). RED:
`TestRuleTemplateMirrorDrift/worktree-integration.md`, sentinel
`RULE_TEMPLATE_MIRROR_DRIFT` (source 61473 B vs mirror 61638 B, exit code 1 — full raw
output at `../evidence/red-baseline-d323f68fd.txt`). Direction settled by measurement —
see research.md §6-§7 and spec.md §B/§E.

## §B Known Issues

- The test failure message suggests `cp local -> template` — a KNOWN-BAD direction for
  this case (violates §25 neutrality C1-C8). Run phase must follow spec.md REQ-MD-001
  (template -> local), NOT the failure message's suggestion.
- Dogfood records placed inside `.claude/rules/moai/` do not survive `moai update` —
  the systemic cause this SPEC repairs for the t1067 record (broader sweep out of scope).
- The census pointer `.moai/reports/t1067/census-20260922.md` was never committed
  (absent from fs, index, and full history — research.md §2 disposition); the relocated
  record must mark it historical, not live evidence.

## §C Pre-flight

1. Confirm worktree HEAD is still `d323f68fd` (or re-measure divergence per AGENTS.md §2
   before any commit — values read earlier in the turn are stale by definition).
2. Confirm the working tree carries no modified/staged entries outside
   `.moai/specs/SPEC-MIRROR-DOGFOOD-001/` (untracked plan artifacts are permitted until
   the plan commit lands).
3. Confirm RED still reproduces: `go test ./internal/template/ -run
   'TestRuleTemplateMirrorDrift' -count=1` exits non-zero with the
   worktree-integration.md sentinel (the RED-now anchor for AC-MD-001).

## §D Constraints

- Single-commit changeset [HARD]: both changed files in ONE commit. The record file is
  mirror-free by design — no template counterpart exists or may be created.
- Template tree [HARD]: zero writes under `internal/template/templates/**`; zero edits
  to `internal/template/rule_template_mirror_test.go`.
- No re-authoring [HARD]: the local copy takes the template bytes verbatim; the record
  file carries the displaced record's substance (SPEC pointer, census path + disposition,
  card id, dates), not a summary written from memory.
- Record file loading scope [HARD]: `.claude/rules/local/wt-ac-restatement-record.md`
  MUST carry a `paths:` frontmatter with the exact glob value
  `"**/.claude/rules/moai/workflow/worktree-integration.md"` (rule-authoring.md
  convention) — it is a card-pinned
  historical record, not turn guidance, and must never join the always-loaded surface
  (rule-authoring.md slot 1; this discharges the statement duty by not creating an
  always-loaded file).
- MX tags: None (documentation-only; explicit Phase 14 finding).

## §E Self-Verification

Run-phase verification batch (all commands, verbatim output retained as evidence):

1. `go test ./internal/template/ -run 'TestRuleTemplateMirrorDrift' -count=1 -v` —
   expect exit 0, all 9 subtests PASS (AC-MD-001).
2. Pre-commit scope check: after staging BOTH changed files by explicit pathspec,
   `git diff --cached --name-only` — expect EXACTLY the 2 in-scope paths, neither under
   `internal/template/templates/` and neither the mirror test file (AC-MD-002, first
   assertion).
3. `cmp .claude/rules/moai/workflow/worktree-integration.md
   internal/template/templates/.claude/rules/moai/workflow/worktree-integration.md` —
   expect exit 0, no output (AC-MD-003).
4. `git ls-files .claude/rules/local/` + grep the pinned file for its four record
   elements (SPEC-AUDIT-EXPORT-CLAUSE-001 pointer, census path + disposition, card
   t1067, 2026-09-22) and its `paths:` frontmatter (AC-MD-004).
5. `go test ./internal/template/ -run TestTemplateNoInternalContentLeak -count=1` —
   expect exit 0 (AC-MD-005).
6. `make build` — expect exit 0 (AC-MD-006).
7. Post-commit scope check: `git show --name-only --format= <M1-commit-SHA>` — expect
   EXACTLY the 2 in-scope paths (AC-MD-002, second assertion; this is the check that
   survives the SPEC artifacts joining the branch).

## §F Milestones

Ordering follows decision-reversibility: the record-file placement decision (where the
displaced content lives, its loading scope — highest change-likelihood) leads; the
mechanical parity restore and build follow.

### M1 — Parity restore + record relocation (Priority High)

1. Create `.claude/rules/local/wt-ac-restatement-record.md` (NEW, tracked) carrying the
   relocated t1067 record: the AC-AEC-013 restatement pointer
   (`SPEC-AUDIT-EXPORT-CLAUSE-001/acceptance.md:622-680`), the census path
   `.moai/reports/t1067/census-20260922.md` with its historical disposition (never
   committed; research.md §2), the card-t1067 attribution, and the 2026-09-22
   measurement dates, plus a one-line statement of why the record lives here
   (durability vs managed root). The file MUST open with `paths:` frontmatter carrying
   the exact glob `"**/.claude/rules/moai/workflow/worktree-integration.md"`
   (conditional load — never always-loaded).
2. Restore byte parity: `cp internal/template/templates/.claude/rules/moai/workflow/worktree-integration.md
   .claude/rules/moai/workflow/worktree-integration.md` (template -> local; the ONLY
   sanctioned cp direction).
3. Verify per §E items 1-6, commit BOTH files in one commit staged by explicit
   pathspec after re-reading `git status --short`
   (`fix(SPEC-MIRROR-DOGFOOD-001): restore worktree-integration.md mirror parity and relocate t1067 dogfood record (card t1086)`),
   then verify §E item 7 (commit-scoped name-only = exactly 2 paths).

Single milestone is intentional (2 files, no inter-file dependency risk beyond the
single-commit changeset).

## §G Anti-Patterns

- Following the test failure message's `cp local -> template` suggestion (REJECTED — §25 neutrality).
- "Fixing" the RED by excluding the file from the mirror test allowlist.
- Splitting the two file changes into separate commits (breaks the single-commit changeset).
- Creating the record file without `paths:` frontmatter (it would become an always-loaded session rule — rule-authoring.md slot 1 — and carry a statement duty instead).
- Editing the template copy "to keep both sides' content" — the template side is already correct; dual-edit is only proper when both sides must change, which is not this case.
- Writing the record file from memory instead of transcribing the displaced record elements from the local copy's current hunk 2 text (research.md §2 quotes them verbatim).

## §H Cross-References

- `../spec.md` — requirements REQ-MD-001..REQ-MD-005 + §F constraints
- `../acceptance.md` — AC matrix AC-MD-001..AC-MD-006 (two-cell, RED-now observed)
- `../research.md` — measured findings (pinned to `d323f68fd`) + census disposition
- `../evidence/red-baseline-d323f68fd.txt` — full raw RED output (exit 1)
- CLAUDE.local.md §2 (Template-First), §4.1 (gitflow lane: merge target is develop, push is lead-batched)
