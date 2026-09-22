# SPEC-MATRIX-GROUP-COMMENT-001 — Implementation plan

Card: **t1055**. Tier **S** — a comment-text change at two sites in one file.
There is one milestone; manufacturing more would misrepresent the size.

## §A Context

See `spec.md` §A and §B. Tree: `.claude/worktrees/t1055`, branch
`WT-retention-comment`, HEAD `d8304b49a`. Every coordinate below was measured in
this tree.

## §B Constraints

- **Comment text only.** No non-comment line of `internal/template/profile_matrix.go`
  changes; no other Go file changes.
- **Do not edit `:8`.** The profile-axis token is a report-only observation
  (`spec.md` §E) pending an operator decision.
- **Do not touch the group layer itself.** Removal is `SPEC-MODEL-MATRIX-CORE-001`'s
  question.

## §C The decision most likely to change — the replacement comment text

This is the only real design choice in the card, so it leads.

The replacement must carry three facts that the current text either omits or
denies, and must stay useful to the reader of `ResolveAgentModelEffort` (whose
lookup genuinely is by name):

1. Lookup here is by agent NAME, not by group — **unchanged and still true**; the
   removal-hazard fix must not discard the sentence that was correct.
2. `AgentGroup` has two consumers, named by path: `internal/cli/model.go`
   (display column) and `internal/web/agentfm.go` (membership gate).
3. The second is a **gate**: it discards the string and uses only the bool, to
   reject an override submission for a non-matrix agent. Removing the group layer
   therefore requires replacing that gate, not just relocating a column.

Proposed text for `profile_matrix.go:481-483` (exact wording is the run phase's
to finalise; these facts are the requirement):

```go
// Lookup is by agent NAME, not by group: per-agent cells split two of the former
// groups, so the group layer no longer carries routing information here. It is
// NOT dead, and it is not display-only: AgentGroup has two consumers of
// different kinds — internal/cli/model.go uses the group string as a table
// column (display), while internal/web/agentfm.go discards the string and uses
// only the membership bool as a validation gate, rejecting a frontmatter
// override submitted for a non-matrix agent. Removing the group layer must
// therefore replace that gate, not merely relocate the column.
```

For `profile_matrix.go:261`, the minimal correction is to stop asserting
"display-only" while keeping the sentence's actual point (the inner key is an
agent name, not a group) — e.g. "(not a group — per-agent cells split two of the
former groups; the group layer no longer keys this matrix, but see AgentGroup:
it still gates override validity)".

Both sites should point at `AgentGroup`, so a reader who arrives at either one
reaches the same two-consumer statement rather than a second summary that can
drift.

## §D Milestone

**M1 — align both comment sites with the code.**

1. Edit `internal/template/profile_matrix.go:481-483` per §C.
2. Edit `internal/template/profile_matrix.go:261` (and the wrapped continuation
   at 262 if the rewrite needs it) per §C.
3. Verify: `git diff -U0 -- internal/template/profile_matrix.go` shows only
   comment lines; `go build ./internal/template/...` and
   `go test ./internal/template/... ./internal/cli/... ./internal/web/...` pass.
4. Commit naming card t1055.

## §E Risks

| Risk | Mitigation |
|---|---|
| The rewrite drifts into asserting something else the code does not do — the same defect class the card fixes | Each claim in the new text maps to a measured coordinate in `spec.md` §B; AC-MGC-001 pins the two paths and the word "gate" |
| The rewrite is read as a licence to remove the group layer | The text states the gate must be *replaced*, not that the layer must survive; scope boundary restated in `spec.md` §D |
| A stray non-comment edit slips in | AC-MGC-003 is a mechanical diff check, not a reading |
| `:8` gets "helpfully" fixed in passing | Explicit out-of-scope bullet; AC-MGC-005 asserts the line is byte-identical |

## §F Self-verification

Run the acceptance commands in `acceptance.md` §D and paste their verbatim
output into `progress.md` §E.2.

## §G Cross-references

- `spec.md` §B — the measurement command and its negative control.
- `SPEC-MODEL-MATRIX-CORE-001` — owns the group-layer removal decision this
  comment feeds.
