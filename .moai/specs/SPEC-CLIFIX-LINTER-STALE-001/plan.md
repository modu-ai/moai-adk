# SPEC-CLIFIX-LINTER-STALE-001 — Implementation Plan

## §A Context

P3 row of the CLI audit roadmap: make the quality gates audit the real system again. The theme is staleness — checks written for a retired system shape (17-agent roster, hand-listed skills) silently stopped checking anything.

## §B Known Issues (findings inventory)

| # | File anchor (re-verify before edit) | Defect | Fix direction |
|---|---|---|---|
| 1 | agentlint/agent_lint.go:530-570 | hand-rolled frontmatter parser never fills Hooks/Skills/Sandbox → LR-04 permanently inert | yaml.v3 unmarshal into the frontmatter struct |
| 2 | agentlint/agent_lint.go:581-599 | canonicalEffortMatrix/writeHeavyAgents built for retired 17-agent roster; current 10 agents uncovered | regenerate from the 10-agent catalog (CLAUDE.md §4 SSOT / .claude/agents/moai/*) |
| 3 | agentlint/agent_lint.go:184,835 | live + template-mirror double scan → LR-07 structural false positives | dedupe by live↔mirror path mapping or content hash |
| 4 | doctor_skills.go:10-27 | static allowlist stale (comment 23 vs actual 22; live skills WARN, retired skills PASS) | derive from template catalog |
| 5 | help.go:57 | phantom `moai brain` in root help | remove; gate help entries on registered commands |
| 6 | team_spawn.go:345-352 | ClaimTask claims nonexistent/completed tasks "successfully" | validate pending match before claim |

## §C Pre-flight

1. Confirm P0-P2 SPECs merged (team_spawn.go and agentlint bases moved).
2. Enumerate the live agent roster from `.claude/agents/moai/*.md` + Explore built-in (10 total) and the template mirror set — this is the regeneration input for #2 and the dedupe map for #3.
3. Identify the catalog SSOT for skills (internal/template/catalog.yaml) for #4 derivation.
4. Baseline: run `moai agentlint` (or `go test ./internal/cli/agentlint/...`) and record current findings for before/after comparison.

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

- M1 — Parser: yaml.v3 frontmatter parsing; RED fixture proving LR-04 was inert, GREEN after; keep parse-error resilience (do not abort whole run on first broken file — regression guard only, full fix deferred per §C scope note).
- M2 — Roster/matrix: regenerate effort matrix + write-heavy list from the 10-agent catalog; violation fixture test.
- M3 — Dedupe + allowlist: LR-07 live/mirror dedupe; doctor_skills catalog derivation with load-failure diagnostic.
- M4 — Surface truths: help.go phantom removal + registered-command gate; ClaimTask pending-validation; full regression suite (REQ-LINT-001-007); §E self-verification.

## §G Anti-Patterns and Risks

- Execution order: P0→P1→P2→P3→P4; this SPEC is fourth. Shared-file overlap: team_spawn.go (CRITICAL-001 b — append fix must be preserved when adding validation), agentlint package (CONTRACT-001 workflow_lint.go os.Exit removal).
- Anti-pattern: regenerating the effort matrix as another frozen hand-list — prefer deriving from the catalog/agent files at build or init time, or add a drift test that fails when the roster changes.
- Anti-pattern: satisfying AC-LINT-001-004 by expanding the static list to today's skills — the requirement is derivation, not refresh.
- Risk: yaml.v3 parsing is stricter than the hand parser — previously-tolerated malformed agent frontmatter may surface as new findings; triage them as genuine findings (report), not parser bugs.
- Risk: dedupe keyed on content hash breaks when mirrors are sanitized variants (not byte-identical) — key on path mapping first, hash as fallback.

## §H Cross-References

- Findings SSOT: `.moai/reports/cli-improvement-audit-20260710.html` §3 clusters 2/4/5, §5 P3.
- Depends on: SPEC-CLIFIX-CRITICAL-001, SPEC-CLIFIX-CONTRACT-001, SPEC-CLIFIX-CONCURRENCY-001. Followed by: SPEC-CLIFIX-HYGIENE-001.
- Agent catalog SSOT: CLAUDE.md §4 (10 retained agents); template catalog: internal/template/catalog.yaml.
