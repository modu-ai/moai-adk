# plan.md — SPEC-FACTORY-BOOTSTRAP-001

> Implementation plan. Order of milestones is by **decision-reversibility** (most-likely-to-change first), not by execution sequence — execution sequence is decided at run-phase. Each milestone carries priority, not time estimates (CLAUDE.md §7, agent-common-protocol.md § Time Estimation).

---

## §A. Existing Assets (Prior Art and Preserve)

### A.1 Prior-art commit `94025ce0a` — the implementation this plan revises

Already-on-disk assets at HEAD `94025ce0a` (branch `feat/factory-bootstrap-guidance`), measured in this worktree:

| Surface | File | Function(s) | Status |
|---|---|---|---|
| Entry switch | `internal/cli/factory.go` | `parseFactoryFlag`, `enterFactoryMode`, `enterFactoryCompanionMode`, `parseCompanionLabel`, `rejectFactoryOnCG`, `recordFactorySession`, `captureEnvState` | exists, **revised** by M2 |
| Vocabulary | `internal/factory/bootstrap.go` | `CompanionRoles`, `NewRunID`, `CompanionLabel`, `SplitCompanionLabel`, `isCompanionRole`, `isRunIDShape` | exists, **unchanged** |
| SessionStart hook | `internal/hook/session_start_factory.go` | `factoryBootstrapNotice`, `factoryLeadNotice`, `factoryCompanionNotice` | exists, **revised** by M4 |
| Env keys | `internal/config/envkeys.go` | `EnvMoaiFactory`, `EnvMoaiFactorySpec`, `EnvMoaiFactoryID`, `EnvMoaiFactoryLabel` | exists, **unchanged** |
| Dispatch (cc) | `internal/cli/cc.go:96-105` | `parseFactoryFlag` + `else if parseCompanionLabel` | exists, **revised** by M2 |
| Dispatch (glm) | `internal/cli/glm.go:169-176` | parallel | exists, **revised** by M2 |
| Block-cap | `internal/cli/launcher_blockcap_infinite.go` | `injectStopHookBlockCapForGoal` reads `EnvMoaiFactory OR EnvMoaiFactoryLabel` | exists, **unchanged** (REQ-FB-005) |
| Help text | `internal/cli/cc.go` `Use:` / `Long:`, `internal/cli/glm.go` `Use:` / `Long:` | neither mentions `-f` | exists, **revised** by M5 |
| Tests | `factory_bootstrap_test.go`, `bootstrap_test.go`, `session_start_factory_test.go`, `launcher_blockcap_infinite_test.go` | prior-art coverage | exists, **extended** by every milestone |

### A.2 Predecessor SPEC

`SPEC-FACTORY-MODE-001` (completed, v0.9.0) — REQ-FM-001..006 govern the flag parse and the `moai cg` rejection; REQ-FM-023 governs the unconditional factory block-cap branch. This plan **preserves** all of them (REQ-FB-004 for the chain seed, REQ-FB-005 / REQ-FB-018 for the block cap).

### A.3 Sibling boundary

`SPEC-KANBAN-BOOTSTRAP-001` (draft, deferred) — owns topology-config-gated quorum, role-declaration carrier (`REQ-KS-006`), dispatch protocol. This plan's notice revision (M4) deliberately removes the role name from the companion notice to avoid colliding with `REQ-KS-006`. The boundary is one-sided from this SPEC (§C of spec.md); the sibling's files are not edited.

### A.4 Measured facts (baseline HEAD `94025ce0a`, this worktree)

- `crossSessionInbound` in `internal/template/templates/`: **0** matches.
- `crossSessionInbound` in `internal/ pkg/ cmd/` Go code: **0** matches.
- `--settings` in moai Go code: **0** matches (the two `--settings` references are inside `moai-foundation-cc/reference/*.md` Claude Code reference docs, not moai code).
- `cc.go:96-105` / `glm.go:169-176` dispatch: `else if` short-circuits companion `-f` entry today.
- `factoryCompanionNotice` (`session_start_factory.go:63-69`): prints `"joined run %s as the %s companion"` today.
- `launcher_blockcap_infinite.go`: reads `EnvMoaiFactory != "" || EnvMoaiFactoryLabel != ""` (line ~50), unconditional raise.
- docs-site: 4 locales × ~133 pages, 14 sections; Factory Mode = 0 pages; new-page contract = frontmatter `title`/`weight`/`draft:false` + 4 files + `main.yaml` `sub:` entry.
- `moai cc` / `moai glm` `Use:` field: `"cc [-p profile] [-- claude-args...]"` — mentions neither `-f` nor `--factory`.

### A.5 PRESERVE — AC003 regression tests

Package path: `internal/cli/launcher_blockcap_infinite_test.go`. Two tests MUST stay green across every milestone:

- `TestAC003_LauncherInjectsRaisedBlockCapForInfiniteGoal`
- `TestAC003_BlockCapDoctrineClauseSpecific`

Rationale: the `-f` redefinition changes **which branch a companion takes** but MUST NOT change **whether the cap is raised for a companion** — the label-only path keeps the raise because `injectStopHookBlockCapForGoal` already OR's the two env vars. These two tests are the regression guard.

---

## §B. Milestones (priority-ordered by decision-reversibility)

### M0 — docs-site section decision (priority: Medium, reversibility: LOW)

**Decision**: place the Factory Mode page under `multi-llm/`, not `advanced/` or a new section.

**Why this is first**: a docs-site section choice is the most expensive to reverse after the page is published — moving a page between sections after the URL is indexed (and after the 4-locale sibling pages and `main.yaml` entry are written) churns URLs, sitemaps, and the menu hierarchy. It is settled before any code is written.

**Rationale**: `multi-llm/` carries backend-mixing sessions (`moai cg`, GLM panes, tmux Agent Teams). A factory run is exactly that — a lead plus N companions, any of which may run on either backend; the prior-art lead notice itself says "Substitute 'moai glm' for 'moai cc' on any companion." `advanced/` was the alternative and is rejected because it carries single-session advanced patterns (worktrees, autonomy loops, deep configuration); a multi-session page would be the odd one out there.

**Reversibility note**: if the placement turns out wrong post-merge, the reverse is a single `git mv` across four locale files plus a `main.yaml` `ref:` rewrite — costly but bounded. The decision is made here, not deferred.

### M1 — `crossSessionInbound` injection via transient `--settings` (priority: High, reversibility: LOW)

The new surface this SPEC introduces. The accept/hold/refuse ladder cannot be satisfied from project/local settings (stricter tier wins), so the launcher writes a transient settings file `{"crossSessionInbound": "accept"}` to a session-private tempdir under `os.TempDir()` and passes `--settings <file>` to the launched backend.

**Decision points** (settled here, not at run-phase):

1. **File location**: `os.TempDir()/moai-factory-<pid>-<random>.json`. Session-private by PID; cleaned up on session exit (the launcher's existing `defer restoreEnvState()` already runs at exit; a parallel `defer os.Remove(settingsFile)` is added).
2. **Operator-supplied `--settings` (REQ-FB-007)**: when the operator passed `--settings <file>`, moai does NOT inject its own; the bootstrap notice prints the advisory instead. Two `--settings` flags on the claude command line collide and the operator's intent wins, so injection is suppressed.
3. **Fail-open (C8)**: a write failure degrades to launching without the injected settings; the bootstrap notice then prints the same advisory as the operator-supplied case (because moai cannot guarantee `accept` either way).

**Reversibility note**: the file-location choice is easily reversed (a constant). The operator-supplied-honor contract is harder to reverse once documented in the docs-site page — locked here.

### M2 — `-f` redefinition + dispatch revision (priority: High, reversibility: MEDIUM)

The core semantic change. The dispatch in `cc.go` and `glm.go` is rewritten to evaluate `-f` and `--name` **together**:

```
1. parseFactoryFlag(filteredArgs) → (specID, factoryEnabled, rest)
2. parseCompanionLabel(rest) → (label, isCompanion)   [run regardless of factoryEnabled, per REQ-FB-002]
3. select branch:
   - factoryEnabled && isCompanion  → enterFactoryCompanionMode(label)   [companion]
   - factoryEnabled && !isCompanion → enterFactoryMode(specID)           [lead]
   - !factoryEnabled                → no-op                               [BREAKING from 94025ce0a: covers both !factoryEnabled && isCompanion (was companion entry under prior art — see spec.md §A.2.1) AND !factoryEnabled && !isCompanion; --name is passed through to claude untouched regardless of shape]
```

The `else if` becomes a flat `if/if` selection on the combination. The two `!factoryEnabled` rows of the §A.2 truth table collapse into one default branch because both are no-ops — the `isCompanion` parse result is inspected only when `-f` is present (it selects between companion and lead); when `-f` is absent the launcher does nothing regardless of `--name` shape. The block-cap inject is **untouched** (it already OR's the two env vars), satisfying REQ-FB-005 and REQ-FB-018 by construction — neither env var is set on the no-op path, so the cap is not raised for no-op sessions (which is correct: they are not factory members).

**Reversibility note**: the truth table is the spec (REQ-FB-001); the dispatch is its implementation. The implementation is easily refactored; the truth table — including the breaking-change 4th row — is locked here.

### M3 — SessionStart notice revision (priority: High, reversibility: MEDIUM)

`session_start_factory.go`:

- `factoryLeadNotice`: gain (c) leader socket path, (d) inbound-automation notice, (e) conditional SPEC line (only when `MOAI_FACTORY_SPEC` set), and the companion launch lines change from `moai cc --name X` to `moai cc -f --name X`.
- `factoryCompanionNotice`: lose the role clause — `"joined run %s as the %s companion"` becomes `"joined run %s"`. Role-name removal rationale at spec.md §A.6.

**Reversibility note**: the role-name removal is the load-bearing decision — it exists to avoid colliding with the sibling's `REQ-KS-006`. Re-adding the role would re-introduce the collision. Locked here.

### M4 — CLI help text (priority: Medium, reversibility: HIGH)

`cc.go` and `glm.go` `Use:` / `Long:` / Flags prose blocks updated to document `-f` (lead) and `-f --name <role>-<run-id>` (companion). No `cmd.Flags()` registration (the flag is forwarded to claude, not parsed by cobra — REQ-FB-014).

**Reversibility note**: pure documentation; easily revised post-merge. Listed this low because it carries no semantic load — the dispatch and notice revisions are where the spec lives.

### M5 — docs-site page + menu entry (priority: Medium, reversibility: MEDIUM)

Four files (`docs-site/content/{en,ko,ja,zh}/multi-llm/factory-mode.md`) plus `data/menu/main.yaml` `sub:` entry under `multi-llm`. Page body covers: lead entry, four companion entries, multi-session bootstrap flow, cross-session messaging substrate, the operator-supplied `--settings` advisory. No new section, no `_meta.yaml` section weight, no `menu.html` SVG case.

**Reversibility note**: once the URL is indexed and the 4-locale siblings are written, moving or removing the page is expensive. Listed after the code milestones because it documents the implemented behavior, not the reverse.

### M6 — Template-First mirror + `make build` (priority: High, reversibility: N/A)

Per C1, every `internal/template/templates/` edit is mirrored to the template source. The CLI help text (M4) touches `internal/cli/cc.go` / `glm.go`, **not** the template tree — the template tree does not carry the launcher Go source — so M4 has **no template mirror**. The docs-site page (M5) **does** live under `docs-site/`, which is **not** under `internal/template/templates/` (docs-site is its own tree), so M5 has **no template mirror either**. M6 is therefore a verification milestone, not an edit milestone: a `make build` run confirming no template source was touched, plus a `grep` confirming no SPEC ID / REQ token / commit SHA leaked into `internal/template/templates/` (C2).

**Reversibility note**: N/A — verification only.

---

## §C. Known Issues

- **None blocking at plan-phase.** The `-f`/`--name` truth table (REQ-FB-001) covers all four cases deterministically; the operator-supplied `--settings` case (REQ-FB-007) is settled in M1; the role-name removal (REQ-FB-010) is settled in M3. No `[NEEDS CLARIFICATION]` markers.

---

## §D. Pre-flight (before run-phase)

- [ ] Plan-audit verdict ≥ 0.85 (Tier L threshold).
- [ ] Implementation Kickoff Approval (plan→run HUMAN GATE) obtained.
- [ ] Branch `feat/factory-bootstrap-guidance` is at HEAD `94025ce0a`, tree clean, base `chore/revert-kanban-rename` (`24c4674b5`) ← origin/main; local ahead by 2, no race.

---

## §E. Constraints (recap from spec.md §D)

C1 Template-First · C2 template neutrality (no SPEC IDs / REQ tokens / SHAs in `internal/template/templates/`) · C3 `.moai/specs/SPEC-KANBAN-*` off-limits · C4 AC003 preserve · C5 commits only (no push, no PR) · C6 worktree isolation · C7 no AskUserQuestion in CLI/hook · C8 fail-open throughout.

---

## §F. Self-Verification (this plan's own consistency check)

- **F.1 REQ ↔ Milestone traceability**: every REQ-FB-001..018 maps to at least one milestone (M1 covers REQ-FB-006/007; M2 covers REQ-FB-001..005; M3 covers REQ-FB-008..011; M4 covers REQ-FB-012..014; M5 covers REQ-FB-015..017; M6 covers REQ-FB-018's preserve + C1/C2 verification). No REQ is unmiled.
- **F.2 AC003 preserve named**: §A.5 names both tests and their package path; REQ-FB-018 binds them.
- **F.3 Sibling boundary one-sided**: this plan edits no `.moai/specs/SPEC-KANBAN-*` file; §C of spec.md is the only side stated.
- **F.4 Prior art recorded**: §A.1 enumerates the 12 files / 7 function groups of `94025ce0a`; the HISTORY row records the commit as revised-not-reverted.
- **F.5 Measured facts cited**: §A.4 cites every measurement against HEAD `94025ce0a` in this worktree (the `verification-claim-integrity.md` §2 baseline-attribution).
- **F.6 Decision-reversibility ordering**: milestones are ordered M0 (docs-site section, lowest reversibility) → M1 (new --settings surface) → M2 (core semantic) → M3 (notice) → M4 (help text, highest reversibility) → M5 (docs page) → M6 (verify). The two most expensive-to-reverse decisions (section choice, operator-supplied-settings contract) lead.

---

## §G. Anti-Patterns

- **AP-1 — Reverting `94025ce0a`.** The prior-art commit is revised, not reverted. Reverting it would discard the working entry switch, vocabulary, and block-cap wiring, and would require re-landing them under this SPEC's name with no semantic gain.
- **AP-2 — Editing `.moai/specs/SPEC-KANBAN-*` from this plan.** The sibling is off-limits (C3). The boundary is stated from this side only.
- **AP-3 — Registering `-f` via `cmd.Flags()`.** Under `DisableFlagParsing: true` the registration is silently inert (REQ-FB-014). The flag is forwarded, not parsed.
- **AP-4 — Printing the role in the companion notice.** Re-introducing the role clause collides with the sibling's `REQ-KS-006` (spec.md §A.6).
- **AP-5 — Mutating the operator's persistent settings to inject `crossSessionInbound`.** The injection is via a transient `--settings <file>` (M1), never via a write to `.claude/settings.json` or `settings.local.json`. The stricter-tier-wins rule makes a persistent write both ineffective and intrusive.
- **AP-6 — Time estimates in milestones.** Milestones carry priority labels (High / Medium), not durations (CLAUDE.md §7, agent-common-protocol.md § Time Estimation).

---

## §H. Cross-References

- `spec.md` §A — context (the SSOT for prior art, redefinition, dispatch, notice, collision-avoidance, AC003 preserve, CLI help, docs-site).
- `acceptance.md` §D — AC matrix mapped per REQ.
- `design.md` — design alternatives considered (the four-way dispatch selection, the settings-injection surface, the notice-content ordering, the docs-site section shortlist).
- `research.md` — measurement record (every §A.4 fact re-traced there with the grep command and observed output).
- `.moai/specs/SPEC-FACTORY-MODE-001/spec.md` — predecessor.
- `.moai/specs/SPEC-KANBAN-BOOTSTRAP-001/spec.md` — sibling (read-only from this side).
- `CLAUDE.local.md` §2 / §15 / §25 — Template-First, language neutrality, content isolation.
