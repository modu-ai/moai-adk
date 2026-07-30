# SPEC-CLIFIX-LINTER-STALE-001 — Implementation Plan

## §A Context

P3 row of the CLI audit roadmap: make the quality gates audit the real system again. The theme is staleness — checks written for a retired system shape (17-agent roster, hand-listed skills) silently stopped checking anything. Re-scoped in v0.2.0 after plan-audit FAIL (0.62): two of the six original defect premises were already fixed upstream (doctor_skills by SPEC-V3R6-DOCTOR-FALSE-SIGNAL-001; canonicalEffortMatrix by the profile-matrix derivation), leaving four live defects.

## §B Known Issues (findings inventory — anchors re-derived against `origin/main` 5832f0671)

| # | File anchor (content-token verified 2026-07-29) | Defect | Fix direction |
|---|---|---|---|
| 1 | `internal/cli/agentlint/agent_lint.go:1183-1235` (parseYAMLFrontmatter; setField TODO-stub cases for `skills`/`hooks` at L1229-1234 populate nothing) | hand-rolled frontmatter parser never fills Hooks/Skills/Sandbox → LR-04 permanently inert | yaml.v3 unmarshal into the frontmatter struct |
| 2 | `internal/cli/agentlint/agent_lint.go:774-777` (writeHeavyAgents slice names `expert-backend`, `expert-frontend`, `expert-refactoring`, `researcher` — all archived) | writeHeavyAgents stale; LR-05 (isolation-drift) matches only dead names, never the live write-heavy roster | reconcile the slice to the CLAUDE.md §4 retained-agent catalog (delete the 4 archived entries; optionally derive from the same source as `canonicalEffortMatrix`) |
| 3 | `internal/cli/agentlint/agent_lint.go:185-191` (scanPaths includes both live `.claude/agents/moai/` and `internal/template/templates/.claude/agents/moai/`) + `:877-967` (checkDuplicateMandateBlocks accumulates `mandateBlocks` across ALL files at L892/L939, emits LR-07 for every block past the first at L951-963, no path-pairing) | live + template-mirror double scan → LR-07 structural false positives | dedupe by live↔mirror path pairing (key each mandate block by content hash, then suppress the second occurrence when the owning paths form a live/mirror pair) |
| 4 | `internal/cli/help.go:57` (`{"moai brain", "Ideation workflow"}`) | phantom `moai brain` in root help | remove the entry; gate help-row inclusion on cobra command-tree registration so future phantom commands cannot reappear |
| 5 | `internal/cli/taskledger/taskledger.go:67-139` (ClaimTask; `targetTaskID = taskID` at L101 assigned BEFORE the pending-search loop L102-106; post-loop check at L126 is only `targetTaskID == ""`, so nonexistent/completed IDs reach the CLAIMED write at L132-134) | ClaimTask claims nonexistent/completed tasks "successfully" | after the search loop, if the explicit-ID branch found no pending match, reset `targetTaskID = ""` so the L126 guard fires |

Dropped premises (no longer defects — recorded here for audit traceability):
- ~~doctor_skills.go static allowlist~~ — already fixed by SPEC-V3R6-DOCTOR-FALSE-SIGNAL-001 (#1088). `doctor_skills.go:33-36` `knownCoreSkills()` calls `template.EmbeddedMoaiSkillNames()`; derivation covered by `doctor_skills_test.go:21/60`.
- ~~canonicalEffortMatrix 17-agent hand-list~~ — already derived. `agent_lint.go:630-641` `buildCanonicalEffortMatrix()` reads `template.DefaultProfileMatrix()[template.PerformanceTierMedium]`. The §G anti-pattern warning against "regenerating as another frozen hand-list" is already respected by the canonical matrix; only the `writeHeavyAgents` slice (#2 above) carries stale names.

## §C Pre-flight

1. Confirm P0-P2 SPECs merged (taskledger.go base for ClaimTask, agentlint package extraction, workflow_lint.go os.Exit removal). All three `depends_on` are `completed` per orchestrator-verified state — no re-audit; verify by reading frontmatter only.
2. Enumerate the live agent roster from `.claude/agents/moai/*.md` + Explore built-in (10 retained agents) — this is the reconciliation input for #2 (writeHeavyAgents) and the dedupe map input for #3.
3. **Baseline the current agentlint output** so the acceptance §D.5 "LR-07 false-positive count drops to 0" claim is falsifiable. Run `moai agent lint --format=json > /tmp/clifix-linter-baseline.json` (or, if the CLI binary is not built, `go test ./internal/cli/agentlint/... -run TestRunAgentLint -v`) and record: (a) total finding count, (b) per-rule finding count, (c) specifically the LR-07 finding count and the files it fires on. Cite this baseline verbatim in progress.md §E.2 alongside the post-fix run so the delta is observable. **Note**: the command is `moai agent lint` (parent `agent` + sub `lint`), NOT `moai agentlint` — the agentlint CLI was extracted to a subcommand during SPEC-CLI-SUBPKG-SPLIT-001.

## §D Constraints

- The linter realigns to the system; the system does not change to satisfy the linter (no agent/skill file edits to make checks pass, except test fixtures).
- Catalog-derived allowlists must fail loudly (diagnostic) when the catalog cannot be loaded — not silently pass everything.
- LR-07 dedupe must not mask true duplicates (two distinct live agents with the same name remain findings).
- help.go gating must be data-driven from the cobra command tree, not another hand-list.

## §E Self-Verification

- E1: AC matrix PASS/FAIL against acceptance.md.
- E2: `go test ./internal/cli/agentlint/... ./internal/cli/... -count=1` verbatim.
- E3: agentlint package coverage not below baseline.
- E5: `golangci-lint run` no new findings.

## §F Milestones (priority order)

- M1 — Parser (REQ-001): yaml.v3 frontmatter parsing; RED fixture proving Hooks/Skills/Sandbox were unpopulated, GREEN after; keep parse-error resilience (do not abort whole run on first broken file — regression guard only, full fix deferred to HYGIENE-001).
- M2 — writeHeavyAgents clean-up (REQ-002): delete the 4 archived entries (`expert-backend`, `expert-frontend`, `expert-refactoring`, `researcher`) from the slice at `agent_lint.go:774-777`; reconcile against the CLAUDE.md §4 retained-agent catalog; violation fixture asserting a retained write-heavy agent without `isolation: worktree` fires LR-05 and an archived-name fixture does NOT. (The `canonicalEffortMatrix` at L630-641 is already derived — leave it untouched.)
- M3 — LR-07 dedupe (REQ-003): path-pairing + content-hash suppression in `checkDuplicateMandateBlocks` (L877-967); live/mirror fixture pair → 0 findings; genuine same-name duplicate fixture → 1 finding.
- M4 — Surface truths (REQ-005, REQ-006, REQ-007): help.go phantom `moai brain` removal + cobra-tree registration gate; ClaimTask pending-state validation at `taskledger.go:101-128` (reset `targetTaskID` when the explicit-ID pending search fails); full regression suite demonstrating each previously-dead check fires AND does not fire on the negative fixture; §E self-verification + cite the §C baseline delta.

## §G Anti-Patterns and Risks

- Execution order: P0→P1→P2→P3→P4; this SPEC is fourth. Shared-file overlap: `taskledger.go` (CRITICAL-001 — the O_APPEND ledger write at L75/L132-134 must be preserved when adding pending-state validation) and the agentlint package (CONTRACT-001 workflow_lint.go os.Exit removal).
- Anti-pattern: re-introducing a frozen hand-list for `writeHeavyAgents`. The canonicalEffortMatrix at L630-641 already derives from `template.DefaultProfileMatrix()`; the writeHeavyAgents fix SHOULD derive from the same source or from the agent catalog rather than naming a fresh static roster that will drift again. If derivation is infeasible within Tier M scope, deleting the 4 archived names is the minimum acceptable fix (a comment citing CLAUDE.md §4 as the SSOT is required either way).
- Risk: yaml.v3 parsing is stricter than the hand parser — previously-tolerated malformed agent frontmatter may surface as new findings; triage them as genuine findings (report), not parser bugs.
- Risk: dedupe keyed on content hash breaks when mirrors are sanitized variants (not byte-identical) — key on path mapping first (live `.claude/agents/moai/` ↔ `internal/template/templates/.claude/agents/moai/`), hash as fallback. Genuine same-name duplicates across two unrelated live paths must remain findings.
- Scope-discipline reminder (D9): all four fixtures (parser, writeHeavyAgents, LR-07, ClaimTask) live under `internal/cli/agentlint/testdata/` and `internal/cli/taskledger/` test files — NOT under `.claude/agents/` or `.claude/skills/`. The linter realigns to the system; agent/skill files are not edited to make checks pass.

## §H Cross-References

- Findings SSOT: `.moai/reports/cli-improvement-audit-20260710.html` §3 clusters 2/4/5, §5 P3.
- Depends on: SPEC-CLIFIX-CRITICAL-001, SPEC-CLIFIX-CONTRACT-001, SPEC-CLIFIX-CONCURRENCY-001. Followed by: SPEC-CLIFIX-HYGIENE-001.
- Agent catalog SSOT: CLAUDE.md §4 (10 retained agents); template catalog: internal/template/catalog.yaml.
