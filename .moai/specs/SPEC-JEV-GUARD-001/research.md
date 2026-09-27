---
id: SPEC-JEV-GUARD-001
title: "research — observation and verification log (plan phase)"
version: "0.1.0"
created: 2026-09-22
author: manager-spec
---

# research.md — SPEC-JEV-GUARD-001

All commands run 2026-09-22 in this worktree (`.claude/worktrees/t1083`), tree `cd99336bf`, branch `WT-jev-guard-green`. Cause chain inherited from t1066 — this log records VERIFICATION, not re-diagnosis.

## 1. The RED (observed)

- **Command**: `go test ./internal/jevmeasure/ -run TestNoConsumerCallPathShips`
- **Verbatim stdout** (bounded tail):
  ```
  --- FAIL: TestNoConsumerCallPathShips (0.05s)
      gate_demo_test.go:137: a consumer call path is present before its measurement: [SkillSuggest in /Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1083/internal/cli/jev_skill_suggest.go]
  FAIL
  FAIL	github.com/modu-ai/moai-adk/internal/jevmeasure	0.325s
  ```
- **Exit code**: 1
- Exactly ONE of the three markers trips (`SkillSuggest`); `NearDuplicateMark` and `LaneQuestionRoute` hold absence.

## 2. Guard test contract (read, gate_demo_test.go:97-158)

- Walks `internal/` for non-test `.go` files; asserts literal absence of the three markers.
- Positive control: confirms the walk read `internal/jevmeasure/measure.go` (`ConstantBaseline`) — a green with an empty walk would be meaningless; the control closes that hole.

## 3. Scope inventory (verification commands + results)

| Check | Command | Result |
|---|---|---|
| SPEC ID availability | `ls .moai/specs/ \| grep -i JEV` | `SPEC-JEV-GUARD-*` absent → free; regex self-check `PASS` |
| SkillSuggest non-test refs | `grep -rn 'SkillSuggest\|jev_skill_suggest\|jev-suggest' --include='*.go' internal/ cmd/ \| grep -v _test` | all hits inside `internal/cli/jev_skill_suggest.go` only |
| root.go registration | `grep -rn 'newJevSuggestCmd' --include='*.go' internal/ cmd/` | `root.go:218` + impl + its tests (note: `grep -n 'jev' root.go` returns nothing — the symbol carries capital `Jev`; the first probe was a case-sensitive miss, corrected by the symbol-level grep) |
| `jevNotice` | grep | defined `todo_jev_finding.go:258`; used by Consumer A + Consumer B. STAYS |
| `jevEnabled` | grep | defined `doctor_jev.go:126`; used by `doctor_jev.go`, `mcp_jev.go`, `todo_jev_finding.go`. STAYS |
| `installJevProbe` | grep | defined only in `todo_jev_finding_test.go:25` (Consumer C helper); the `jev_skill_suggest.go:405` hit is a comment. STAYS |
| kanban coupling claim | grep of `internal/kanban/foreman_queue_watch_test.go` | extracts `**Queue watch.**` block from `moai-kanban-foreman/SKILL.md` — unrelated, confirmed |
| md/yaml/json/tmpl sweep | `grep -rln 'jev-suggest' . --include='*.md' ...` | 2 SKILL.md copies + 4 codemap files + CONSUMERS progress.md (historical — untouched) |
| SKILL.md prose location | read | both copies lines 115-123: `### Skill Suggestion (gated — default off)` subsection |

## 4. SKILL.md divergence baseline (measured — corrected per plan-audit D1)

**Corrected measurement** (re-run after iter-1 FAIL, tree `cd99336bf`):

- **Command**: `diff .claude/skills/moai/SKILL.md internal/template/templates/.claude/skills/moai/SKILL.md` (exit 1)
- **Observed**: 20 hunks / 80 plain output lines / **41 diff content lines** (`grep -c '^[<>]'` → 41; `<` 21, `>` 20) / **21 changed lines across 4 categories**:
  - A — 17 single hunks, `For detailed orchestration:` lines (`${CLAUDE_SKILL_DIR}` vs literal path)
  - B — hunk `288,289c288,289` (2 lines), harness-Builder block (Builder paragraph + its orchestration line, same `${CLAUDE_SKILL_DIR}` variant)
  - C — hunk `335c335` (1 line), `moai cg -w` vs `moai cc -w`
  - D — hunk `407d406` (1 line), local-only `Last Updated: 2026-07-07`
- **Executable no-widening check** (comparison-side normalization; categories A+B collapse):
  - `diff <(sed 's|\${CLAUDE_SKILL_DIR}|.claude/skills/moai|g' .claude/skills/moai/SKILL.md) internal/template/templates/.claude/skills/moai/SKILL.md` → pre-edit: exactly 2 hunks (C + D shapes), exit 1
  - shape assertion `… | grep '^[<>]' | grep -vc -e 'moai c[gc] -w' -e 'Last Updated:'` → prints **0** (exit 1 from the zero-match grep is the pass signal)

The iter-1 artifacts stated "exactly 17 lines" — a miscount of category A alone, with internally inconsistent arithmetic (17 vs 16+2) and both the Builder category and lines 335/407 omitted. The fresh measurement agrees exactly with the auditor's 21/41; recorded as an observation correction, not a reconciliation. Whole-file byte-identity was never true on this tree; the AC criterion is prose absence + shape-pinned no-widening (spec.md §D.3, acceptance.md §D.4).

## 5. Upstream blockers (inherited, not re-diagnosed)

- t1066 F2: TypeSafe wire format — endpoint 422 requires `questions` dictionary-typed with per-question `type` discriminator; response `answers` also dictionary; R3 probe reproduced 200 after correction. Separate repair card. Measurement (and therefore Consumer B re-landing) is blocked on it.
- t1068: sibling card on the acbent absent-83-rows guard-AC axis — different files, do not touch.

## 6. Codemap staleness (post-withdrawal)

`.moai/project/codemaps/{docs-truth,entry-points,data-flow,modules}.md` reference `jev-suggest` (docs-truth.md:112; entry-points.md:7,63,70; data-flow.md:507; modules.md:59). Generated docs → sync-phase regeneration (`moai codemaps`) by manager-docs. Not run-phase surface.

## 7. Gaps (what this plan phase did NOT observe)

- CI verdict on origin/develop@f5fff2190 — read from the dispatch context, not re-verified here (network read not performed in plan phase).
- `go test ./...` full suite — deliberately not run (lane-local rule).
- The exact commit SHA that landed Consumer B in t1066 M6 — not pinned; recoverable via `git log --follow -- internal/cli/jev_skill_suggest.go` before deletion (run-phase pre-flight may pin it for the restoration record).

## Residual risk

- The SKILL.md divergence baseline (21 changed lines / 4 categories) was measured at `cd99336bf`; if another lane merges a SKILL.md change before t1083's integration, re-measure before judging AC-JEVG-004 (divergence comparison is against the pre-edit tree of THIS worktree at merge time). The AC's shape-pinned check is robust to line-number shifts from other edits, but a baseline-category edit by another lane (e.g., a `moai cg/cc` wording change) would need the shape list re-validated.
