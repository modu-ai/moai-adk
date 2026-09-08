# t367 verdict — plan-auditor GEARS rubric realigned to canon: Event-detected is the fifth pattern, Unwanted is legacy-only

Card: t367 (G2a, lane-9) · Branch: WT-gears-canon-rubric · Base: local develop `ddc3be6fa` · Commit: this verdict's commit (edit + verdict in one)

## Claim

`.claude/agents/moai/plan-auditor.md` (and its template mirror) mislabeled the GEARS pattern list at line 72: it named "Unwanted — The <subject> shall not [action]" as the fifth GEARS pattern, called it the "GEARS canonical negative form", and inverted the DEPRECATED annotation ("use shall not"). All three canon sources converge the opposite: the fifth GEARS pattern is **Event-detected** ("When [undesired-condition-detected], the <subject> shall [response]"), IF/THEN is the deprecated legacy form, and shall-not/Unwanted is not a GEARS pattern at all. The rubric now names Event-detected as the fifth pattern and states Unwanted is legacy-EARS-only (never canonical, never steered toward, countable only under the Score-1.0 legacy-equivalents allowance).

## Evidence

Three canon sources, re-observed in this tree at this run:

1. `.claude/skills/moai-workflow-spec/SKILL.md:59` — GEARS table row: `Event-detected (replaces IF/THEN) | "**When** <undesired-condition-detected>, the <subject> shall <response>" | IF <condition> THEN <action> **[DEPRECATED — use WHEN <event-detected>]**`.
2. `docs-site/content/en/workflow-commands/moai-plan.md:468` — "deprecation of the IF/THEN pattern (normalized to WHEN)"; :492-495 quotes the lint emission.
3. `internal/spec/lint.go:756-757` — code `LegacyEARSKeyword`, message `GEARS migration: replace IF/THEN with WHEN/event normalization`.

Post-edit verification, this run, this tree:

- `go test ./internal/template/ -run 'TestSanitizedPairParity' -count=1` → `ok ... 0.493s` (parity holds for the edited pair).
- Inverted-phrasing sweep: `grep -c "GEARS canonical negative"` and `grep -c "use shall not"` → 0 hits in BOTH copies; `grep -c "Event-detected"` → 1 hit in both copies.
- `git diff --stat` → exactly 2 files (`plan-auditor.md` × 2 copies, +3/-2 each). Line 158's unrelated "shall not execute it further" prose untouched. No test files reference the "Unwanted" wording (swept pre-edit).

## Baseline-attribution

All measurements this run, this tree (WT-gears-canon-rubric): the three canon-source greps above (file:line observed in this worktree); the parity test (this run, exit 0); the sweeps (this run, counts as printed). The defect itself was observed pre-edit on the same tree (absorbed develop `ddc3be6fa`).

## Gaps (explicitly NOT observed)

- The full `TestSanitizedPairParity` subtest for plan-auditor.md specifically was not isolated by name — the whole matching set ran and passed.
- The full `internal/template` package suite and `internal/spec` lint suite were not run (lane-local scope: the change is 2 markdown lines in an agent body; the parity guard is the covering test).
- CI's full-suite verdict belongs to the lead's batch push.

## Residual-risk

- The score-0.75..0.25 bands ("EARS/GEARS patterns" wording) still speak of patterns generically; if a future revision re-enumerates legacy EARS forms there, the same drift could recur — the fix here pins the Score-1.0 list only, which is where the inverted annotation lived.
- A pre-existing red was observed along the way and is NOT this card's: `TestCatalogHashParity` / `TestManifestHashFormat` catalog drift (entry `sync-auditor`), attributed to the lead's `4244c4a06` batch, repair queued with the operator (see `.moai/reports/t353/verdict.md` Finding 2 for the attribution).
