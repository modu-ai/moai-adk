# Research — SPEC-SUBCOMMAND-RETIRE-001

> Plan-phase research artifact. Captures the verified removal inventory, empirical
> justification, dependency-exclusivity verification, and CI-guard test analysis
> that the spec/plan/acceptance artifacts build on. All file existence, grep, and
> structural facts below were observed during plan-phase investigation (not assumed).

## §A. Empirical Justification (cited — verified earlier this session, not re-derived)

| Signal | Observation | Implication |
|--------|-------------|-------------|
| SPEC center of gravity | `.moai/specs/SPEC-*` count = 392 | plan/run/sync is the overwhelming center; the 5 targets are peripheral |
| Half-done cleanup | commit `1ece11578` removed `brain`·`design` from the docs-site menu/content, but the workflow files, SKILL.md router, dependent skills, and catalog entries remain live | this SPEC completes a started-but-unfinished retirement |
| e2e dead-weight | `e2e.md` is the largest workflow file (452 lines / 20.5 KB) with 0 output commits and 0 `<command-name>` invocations across 263 session transcripts | strongest dead-weight signal of the 5 |
| Domain fit | moai-adk-go = Go CLI tool by a 1-person OSS author | brand/UI design, 7-phase ideation, browser E2E are out of its domain |
| coverage redundancy | `go test -cover` already embedded in gate/run phases | `/moai coverage` is redundant |
| security re-routing | OWASP audit can route through retained `moai-ref-*` reference skills | `/moai security` removable without capability gap |

## §B. Verified Removal Inventory

### B.1 Command definition surface (CORRECTION to original task assumption)

The task stated "NO `.claude/commands/moai/{design,brain,e2e,coverage,security}.md` files exist." **This is false.** All 5 command files DO exist in both trees and are the real subcommand definition surface. Each is a thin `Skill("moai")` router:

```
.claude/commands/moai/{brain,design,e2e,coverage,security}.md          (local — all plain .md)
internal/template/templates/.claude/commands/moai/                      (template):
  brain.md, design.md, security.md        → plain .md
  coverage.md.tmpl, e2e.md.tmpl           → .md.tmpl (templated)
```

Body of each (verbatim shape): `Use Skill("moai") with arguments: <subcommand> $ARGUMENTS`. Removing a subcommand therefore deletes its command file in both trees (note the mixed `.md` / `.md.tmpl` forms in the template tree).

### B.2 Workflow files (both trees; NO team/ variants for any of the 5 — verified)

```
.claude/skills/moai/workflows/{design,brain,e2e,coverage,security}.md
internal/template/templates/.claude/skills/moai/workflows/{...same 5...}.md
```

### B.3 Dependent skills (7 — both trees confirmed present)

| Skill | Catalog tier (verified in `internal/template/catalog.yaml`) | Files/tree |
|-------|--------------------------------------------------------------|-----------|
| `moai-domain-ideation` | **core.skills** (line 91) | 1 |
| `moai-domain-research` | **core.skills** (line 96) | 1 |
| `moai-domain-design-handoff` | optional-pack:design (line 184) | 3 |
| `moai-domain-brand-design` | optional-pack:design (line 169) | 1 |
| `moai-domain-copywriting` | optional-pack:design (line 174) | 1 |
| `moai-workflow-design` | optional-pack:design (line 189) | 1 |
| `moai-workflow-gan-loop` | optional-pack:design (line 194) | 1 |

**Key structural fact**: 2 of the 7 are **core skills** (ideation, research); 5 are **optional-pack:design**. The design pack also contains `moai-domain-humanize` (line 179) which is **RETAINED** (used by docs-site ko/ja humanize per project memory) — so removing 5 design-pack skills leaves the pack non-empty.

### B.4 SKILL.md intent router surface (both trees, `.claude/skills/moai/SKILL.md`)

- **Header subcommand list** (lines 5-7): lists all subcommands incl. brain/design/coverage/e2e/security — must drop the 5.
- **Priority 1 explicit routing** (lines 58, 62, 71, 72, 74): `brain` (aliases ideate/idea), `design` (aliases brief/brand), `coverage` (alias cov), `e2e` (alias e2e-test), `security` (aliases audit/sec).
- **Priority 3 natural-language** (lines 85, 87): line 85 "Planning and design language … routes to plan" (keep — design word routes to plan, which survives); line 87 "Security language … routes to security" (must reroute to review/Agent security scope).
- **Quick Reference dedicated blocks**: brain (105-111), security (143-148), coverage (192-197), e2e (199-204), design (206-212).

### B.5 catalog.yaml (manual edit — generator is hash-only)

`make build` runs `go run ./internal/template/scripts/gen-catalog-hashes.go --all`, which is **hash-only**: it updates the `hash:` field of EXISTING entries and does NOT auto-discover or remove entries (verified by reading the generator header + struct). Therefore the 7 skill entries MUST be **manually deleted** from `internal/template/catalog.yaml` (lines 91-98 for the 2 core skills; lines 169-196 region for the 5 design-pack skills, preserving the `moai-domain-humanize` entry). Sequencing matters: if a skill dir is removed but its catalog entry remains, `gen-catalog-hashes --all` will try to hash a missing `SKILL.md` and error. Remove catalog entries + skill dirs together, then `make build`.

## §C. CI-Guard Test Analysis (HIGHEST-RISK surface)

Count model verified: **35 skills + 7 agents = 42 total catalog entries**. After removing 7 skills → **28 skills / 35 total**; agent count unchanged (all 7 removals are skills).

| Test file | Hardcoded expectation | Required change |
|-----------|-----------------------|-----------------|
| `catalog_tier_audit_test.go` | `const expectedSkillCount = 35` (line 154); `expectedAgentCount = 7` (unchanged) | 35 → **28**; append history-comment entry |
| `catalog_loader_test.go` | `const expectedTotal = 42` (line 53) | 42 → **35**; append history-comment entry |
| `embed_catalog_test.go` | `const wantTotal = 42` (line 44) | 42 → **35**; append history-comment entry |
| `commands_audit_test.go` | `TestBrainCommandThinPattern` (line 110) hardcodes `.claude/commands/moai/brain.md` existence | **delete the test function** (brain.md removed). `TestCommandsThinPattern` + `TestCommandsFrontmatterConsistency` iterate `cmdFiles` dynamically — no name-list change needed |
| `agentless_audit_test.go` | **THREE** package-level path lists read deleted files via `fs.ReadFile`→`t.Fatalf` (plan-auditor iter-1 D1 correction): (1) `var utilitySkillPaths` L31 has `coverage.md` → `TestAgentlessUtilityNoLLMControlFlow` L77 + `TestUtilitySkillsContainModeFlagIgnoredSentinel` L129; (2) `var implementationSkillPaths` L44 has `design.md` → `TestImplementationSkillsContainPipelineRejectionSentinel` L159; (3) `runDesignSkillPaths` (local, L186) has `design.md` → `TestRunDesignSkillsContainModeUnknownSentinel` | **remove `coverage.md` from `utilitySkillPaths`, `design.md` from `implementationSkillPaths`, and `design.md` from `runDesignSkillPaths`** (3 edits). Stale-comment hygiene: L28 "5 utility"→4, `@MX:ANCHOR fan_in=5`→4, L38 "4 implementation"→3 |

**Cleared from the risk list (verified NOT affected):**
- `internal_content_leak_test.go` — its "design" references are all the `internal/design/` **Go package** (DTCG tokens) regex (`internal/(spec|cli|hook|ciwatch|design)/`), unrelated to the design subcommand. No change required.
- `catalog_slim_audit_test.go` — assertions iterate `cat.AllEntries()` **dynamically** (`audited++`/`t.Logf`); the "17 optional skills / 25 hidden / 40 reachable" figures are **comments only**, not constants. Build will not break. Comment refresh is optional (SHOULD, hygiene).
- `skill_authoring_paths_glob_test.go` (lines 93, 141) — uses `design.md` as a **string fixture** for glob-pattern matching, not a file read. Not affected; optional cosmetic fixture swap.
- `catalog_loader_test`/`embed_catalog_test` agent-count side — unchanged (no agent removed).
- `TestWorkflowTriggerCoverage` (catalog_tier_audit_test.go:438) — walks the workflow dir + builds `knownSkills` from catalog **dynamically**; only breaks if a *remaining* workflow declares `required-skills` pointing at a removed skill. Verified: no remaining workflow references any of the 7 removed skills (grep clean).

## §D. Dependency-Exclusivity Verification (grep, scratch/spec/backup excluded)

### D.1 The 7 skills — external (non-removed, non-self) references:

| Surface | Reference | Disposition |
|---------|-----------|-------------|
| `.claude/rules/moai/core/glm-web-tooling.md` (+template) | "Cross-referenced by: … `moai-domain-research/SKILL.md` …" (line 12) | dangling after research removal → **edit** (drop the cross-ref token) |
| `.claude/skills/moai-domain-humanize/SKILL.md` (+template) | line 149: "`moai-domain-copywriting`: the generation counterpart …" | humanize RETAINED, copywriting REMOVED → **rewrite** to remove dependency (REQ-SCR-010) |
| `.claude/rules/moai/core/zone-registry.md` (+template) | FROZEN mirror entries 051-149 for `design/constitution.md` | **OUT OF SCOPE** (zone-protected FROZEN; documents design methodology, not the slash command) |
| `.claude/rules/moai/design/constitution.md` (+template) | FROZEN design pipeline rule | **OUT OF SCOPE** (FROZEN — header confirms "FROZEN/EVOLVABLE zone"; describes design methodology + archived `expert-frontend` role carve-out) |
| `internal/template/templates/.moai/design/README.md` | design subsystem readme | **OUT OF SCOPE** (design rule subsystem, deferred) |
| `.claude/skills/moai/SKILL.md`, `workflows/brain.md`, `workflows/design.md` | self (being removed/edited) | handled by M1 |

ideation/research/design-handoff referenced by NON-removed workflows: **only `brain.md`** (being removed). ✓ Safe.

### D.2 The 5 subcommands — references in rules/output-styles/CLAUDE.md (in-scope cleanup):

| Surface | Disposition |
|---------|-------------|
| `.claude/output-styles/moai/moai.md` (+template) — line 121 `… or /moai security` | **edit** (reroute to review/Agent security scope) |
| `CLAUDE.md` (+template) — §3 "Unified Skill: /moai design" section + §3 subcommand list (brain/design/coverage/e2e/security/codemaps) | **edit** (drop the 5 + the design section) |
| `.claude/rules/moai/workflow/spec-workflow.md` (+template) | **edit** (drop dangling subcommand refs) |
| `.claude/rules/moai/core/zone-registry.md`, `.claude/rules/moai/design/constitution.md` | **OUT OF SCOPE** (FROZEN/zone-protected) |

### D.3 docs-site residual (commit 1ece11578 already removed brain·design):

| Locale | quality-commands pages present | Menu (main.yaml) |
|--------|-------------------------------|------------------|
| en / ko / ja / zh | `moai-coverage.md` + `moai-e2e.md` (each locale) | `coverage` (276-280), `e2e` (282-286) entries (4-locale each) |

- **No `moai-security.md` page exists** in any locale (security never had a docs page). Cross-links to coverage/e2e appear in `quality-commands/{_index.md,_meta.yaml,moai-review.md,moai-codemaps.md}` and ko `getting-started/introduction.md` + `core-concepts/what-is-moai-adk.md`.
- Removal = 8 content files (coverage+e2e ×4 locales) + menu entries + cross-link cleanup, under §17 4-locale parity doctrine.

## §E. Replacement-Path Capability Map (no capability gap left behind)

| Retired subcommand | Capability | Surviving path |
|--------------------|-----------|----------------|
| `security` | OWASP Top-10 audit, deps/secrets scan | `/moai review --security` (line 173-175 already uses `Agent(general-purpose)` security scope) + natural-language "security audit" → `Agent(general-purpose)` loading retained `moai-ref-owasp-checklist`/`moai-ref-secops`/`moai-ref-supply-chain`/`moai-ref-llm-security` (optional-pack:devops, NOT removed) |
| `coverage` | coverage gap analysis | `go test -cover` embedded in `/moai gate` + `/moai run` |
| `brain` | 7-phase ideation → proposal | `/moai plan` (Socratic requirement clarification + diverge-converge) |
| `design` | brand/UI design pipeline | `moai web` console + `.moai/project/brand/` config retained; design methodology rule (FROZEN) persists as reference |
| `e2e` | browser E2E (Chrome/Playwright) | **no in-tool replacement** — out of Go-CLI domain; users invoke Playwright/Chrome directly |

## §F. Precedent Retire-SPECs (templates for run-phase pattern)

- `SPEC-V3R6-DESIGN-SYSTEM-RETIRE-001` — prior design-system retirement (net -1 skill; catalog count history in test comments)
- `SPEC-V3R6-SEQ-THINKING-RETIRE-001` — has a dedicated `seq_thinking_retire_audit_test.go` (a retirement-guard test pattern; this SPEC reuses count-constant updates rather than adding a new guard test)
- `SPEC-V3R5-CORE-SLIM-001` / `-B-001` — catalog skill-count slimming precedent
