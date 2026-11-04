# Card t2028 — rule documents progress record (docs lane)

Scope: items 1, 3, 4, 5. Item 2 (sync-auditor regression table) belongs to builder-harness. Every edit is applied identically to the template source (`internal/template/templates/.claude/rules/moai/...`) and the live mirror (`.claude/rules/moai/...`).

## Item 1 — non-converging repair
- Existing-sentence check: grep over `.claude/rules/moai` for `consecutive|converg|twice in a row|repair (loop|round)` found only unrelated hits (goal-directive stagnation guard, runtime-recovery-doctrine same-rung circuit breaker at 3, agent-common-protocol-reference E1 bug-deadlock at 3). The nearest neighbour is `factory-dispatch-detail.md:168` "Ceiling. After a repair the lane re-reviews at most 2 times", which caps re-review count but does not stop a chain of repair-induced defects. No equivalent sentence: ADDED.
- Home: `workflow/factory-dispatch-detail.md`, new bullet "Non-converging repair" directly after the card-review "Ceiling" bullet (paths-scoped file, zero always-loaded cost).
- Threshold: 2 consecutive is a PROPOSED value, stated as such in the text, with rationale (one induced defect is noise, two in a row shows the approach is not converging; evidence card t2013 K6 -> F1 -> F8).
- prediction: in the next 10 lane cards that run a review/audit repair loop, a card whose progress record shows two consecutive repair-induced defects also records a revert/redesign/narrow-scope choice before any third repair commit; a third same-approach repair commit after two consecutive induced defects, with no recorded choice, falsifies this edit. verified: pending.

## Item 3 — never write secret values into logs/reports/evidence
- Existing-sentence check: `grep -i secret` over `core/moai-constitution.md` found only line 137 "Never commit secrets to version control" (commit-time, not log/report/evidence). AGENTS.md and AGENTS.md.tmpl had no `secret` hit at all. No equivalent: ADDED.
- Constitution: one bullet under "Security Boundaries" after the commit bullet.
- AGENTS.md decision: ADDED one sentence to section 7 (Tools and command output), in both `AGENTS.md` and `internal/template/templates/AGENTS.md.tmpl`. Reason: AGENTS.md is the only rule surface a Codex lane loads (the constitution is not), the cross-harness contract currently has no secret clause at all, and Codex lanes write the same verdict/evidence files. Mirroring: the template source is `AGENTS.md.tmpl`; the repo-root `AGENTS.md` is a rendered variant that already differs from the tmpl in a few paragraphs, so both were edited with the identical sentence rather than relying on a render step.
- prediction: no verdict, progress record, or log written by a lane after this lands contains a literal token/key/password value (a sweep `grep -rE` for known secret shapes over new `.moai/reports/**` finds 0 hits); one hit falsifies it. verified: pending.

## Item 4 — verify a new dependency exists before adding
- Existing-sentence check: grep for `dependenc|exist` in `development/coding-standards.md` found only the Duplicate-Prevention "exists in exactly one location" line; repo-wide grep for `slopsquat|hallucinated package|go list -m|npm view|pip index` in `.claude/rules/moai` found none. The Simplicity ladder step 5 ("already-installed dependency") covers reuse, not existence. ADDED.
- Home: new `## New Dependencies` section in `development/coding-standards.md` (before Advisory-Check Discipline), language-neutral (three registry lookups plus "or the language's equivalent"). The manager-develop agent definition is left to builder-harness.
- prediction: a dependency added to go.mod/package.json/pyproject in a card after this lands is accompanied by a cited lookup output in the progress record or commit body; a card adding a dependency whose name does not resolve in the registry reaching review falsifies it. verified: pending.

## Item 5 — shared-DB schema migration as keep-set category 3 example
- Existing-sentence check: `grep -i "migration|schema"` over `workflow/auto-semantics.md` found no hit; section 9 intro (lines 208-212) and the inventory rows name only origin push and branch disposal. ADDED.
- Edit: section 9 intro, category 3 now reads "irreversible operations touching external shared systems (for example an origin push, or a schema migration applied to a shared database)".
- prediction: a lane card that applies a schema migration to a shared database records a KEEP / operator-held disposition (not an autonomous decision record); an autonomous decision record covering a shared-DB schema migration falsifies it. verified: pending.

## Files touched (each in both trees)
- `core/moai-constitution.md` (always-loaded)
- `development/coding-standards.md` (paths-scoped via frontmatter)
- `workflow/auto-semantics.md` (paths-scoped via frontmatter)
- `workflow/factory-dispatch-detail.md` (paths-scoped via frontmatter)
- `AGENTS.md` and `internal/template/templates/AGENTS.md.tmpl` (always-loaded)

## Byte delta (wc -c, before -> after)
- moai-constitution.md 15155 -> 15296 (+141, always-loaded)
- coding-standards.md 13367 -> 13771 (+404, paths-scoped)
- auto-semantics.md 30589 -> 30670 (+81, paths-scoped)
- factory-dispatch-detail.md 29435 -> 30108 (+673, paths-scoped)
- AGENTS.md 22526 -> 22662 (+136, always-loaded; limit 40000)
- AGENTS.md.tmpl 22504 -> 22640 (+136)
- Always-loaded growth this card: +277 B across two files; no single edit exceeds the 1,000 B statement threshold of rule-authoring.md.

## Template-side follow-ups found by tests (recorded for traceability)
- The card id was removed from the template copy of the item 1 bullet: `TestTemplateNoInternalContentLeak` (class C9-card-id) and `TestReviewOwnership_CardReviewDoctrine` reject internal card tokens in template files. The evidence (card t2013, K6 -> F1 -> F8) lives only in this record.
- `internal/template/testdata/binding_ledger.json` row ALB-0115 (Security Boundaries list): `after_text` updated to include the new secret-values bullet; `before_text` is the anchor baseline and stays unchanged (changing it fails the row-coverage check). Without this, `TestBindingLedgerIntegrity` fails with "after-text absent for row ALB-0115".
- Catalog hash: `make build` ran `gen-catalog-hashes.go --all`; `catalog.yaml` came out unchanged (24603 bytes, no diff in `git status`), so no hash regeneration was needed for these rule/AGENTS edits.

## Verification (this run, tree = card branch before the commit)
- `go test ./internal/template/... -count=1 -timeout 40m`: exit 0; `ok internal/template 972.553s`, agentemit/commandemit/embedemit ok. Two earlier runs failed for the reasons above (card-id token, ledger row) plus environmental toolchain/default-10m-timeout noise.
- `make build`: exit 0.
