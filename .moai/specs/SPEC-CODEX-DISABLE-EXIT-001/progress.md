# progress.md — SPEC-CODEX-DISABLE-EXIT-001

## §E.1 Plan-phase Audit-Ready Signal

plan_status: pending-plan-audit
plan_complete_at: (pending — set at plan-phase close / audit PASS)
baseline_tree: a4855f0b2
card: t548
tier: S (artifact set 4 files per dispatch order)
M1_status: done (caller census performed at plan phase — evidence below)

### Plan-phase Evidence — M1 caller census (commands + verbatim outputs)

All searches run with `rg` against this worktree at `a4855f0b2`, this run, 2026-09-08.

1. The card's literal form — `rg -n --no-heading "codex skills" --hidden -g '!.git' -g '!node_modules'` → **4 hits, all documentation of the REJECTED form** (`.moai/specs/SPEC-CODEX-SKILL-DISABLE-001/plan.md:74` records its design-time rejection; `SPEC-CODEX-SKILL-PATH-SLASH-001` prose). Zero invocations — the form does not exist as a CLI surface.

2. The actual verb — `rg -n --no-heading "skills disable" --hidden -g '!.git' -g '!node_modules'` → **17 hits**, partitioned:
   - In-repo invocations: `.moai/reports/t502/e2e-verb.sh:19,41,43,68` (one dev-only E2E evidence script from the verb's own card t502; not shipped).
   - Go callers: `internal/cli/codex_skills_disable_test.go` — 10+ direct `runCodexSkillDisable(...)` calls (`:293,317,348,369,533,597,617,645,666`) + `newSkillsCmd()` at `:511`; they consume the returned `error`, not process exit codes.
   - Source/doc mentions: `internal/cli/skills.go:8`, `codex_skills_disable.go:387`, `CHANGELOG.md:20,386`, `.moai/project/codemaps/entry-points.md:59`, SPEC prose.

3. Sibling form — `rg -n --no-heading "skills prune" --hidden -g '!.git'` → **0 hits**. The sibling verb's real surface is `moai clean --codex-skills` (`internal/cli/codex_skills_prune.go:157`; `Kept:` at `:195`).

4. Templates — `rg -n "skills disable" internal/template/templates/` → **0 hits** (no hook wrapper, no template).

5. CI — `rg -n "skills" .github/workflows/ | rg -i "disable|prune|codex"` → **0 hits**.

6. docs-site — `rg -ln "skills disable|codex skills|skills prune|skills.config" docs-site/content/` → 8 files (doctor.md ×4 locales, moai-clean.md ×4 locales — all about `[[skills.config]]` shape, none about the disable verb); `rg -n "moai skills" docs-site/content/` → **0 hits**. The verb is undocumented in the user docs.

7. README ×4 — `rg -n "skills disable|skills prune|codex-skills" README*.md` → only `moai clean --codex-skills` (line 749 ×4). The disable verb is undocumented.

**Census verdict**: zero production in-repo callers (scripts / hooks / CI / templates). The exit-code contract exists for outside-this-repo consumers and as a published CLI surface; no published surface documents the Skipped branch's exit code, so there is no documented promise an exit-code change would break.

### Plan-phase Evidence — code reading (this tree, `a4855f0b2`)

- Runner branch table B1-B10 read at `internal/cli/codex_skills_disable.go:394-455` (spec.md §B.1). Tally: nil returns at `:399,408,414,421,424,434,454`; `fmt.Errorf` at `:402` (name resolution), `:441` (backup), `:451` (write); plus command-level errors at `skills.go:90,96`.
- Skipped reasons enumerated in `upsertCodexSkillDisable` (`:231-311`): no path `:237-239`; unencodable char `:261-263`; duplicates `:282-284`; unrecognized line `:293-295`; no rewritable key `:302-304`. Reasons (c)/(e) are actionable by `moai clean --codex-skills` — exit 0 hides that handoff.
- Documented intent read at `:387-393` (doc comment: fail-open scoped to missing inputs; name-not-found = typo = non-zero "so a script can see it") and `CHANGELOG.md:20` (mirror-absent-zero rationale). **Unchanged/Skipped bucketing is undocumented on every surface.**
- Sibling fail-open doctrine read at `codex_skills_prune.go:159-162`; per-entry `Kept:` never affects exit (`:182-196`).
- Verb-tree shape read at `skills.go:53-55`: the `skills` family contains exactly one verb (disable).

### Plan-phase Evidence — adjudication analysis

Both outcomes prescribed in spec.md §E. Outcome A (document-only) is viable but must carry an explicit card-rule-2 waiver (Unchanged/Skipped share bucket 0). Outcome B (Skipped → non-zero; Unchanged and absent-inputs stay 0) is the RECOMMENDED default: rule 2 is a stated operator requirement the current code violates; the code's own doc comment scopes fail-open to missing inputs, not refusals; the two rule-2-actionable skip reasons hand off to a different verb; and the census shows no documented exit-code promise for Skipped to break. Scope axis resolved at plan phase (spec.md §D): verb-local, with the single-target-vs-sweep distinction documented against `moai clean --codex-skills`.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
