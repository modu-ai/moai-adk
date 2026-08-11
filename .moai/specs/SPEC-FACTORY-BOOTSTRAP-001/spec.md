---
id: SPEC-FACTORY-BOOTSTRAP-001
title: "Factory Mode multi-session bootstrap guidance and companion entry"
version: "0.2.0"
status: draft
created: 2026-08-11
updated: 2026-08-11
author: manager-spec
priority: High
phase: "v3.1.0 target"
module: internal/cli, internal/hook, internal/factory
lifecycle: spec-anchored
tags: "factory-mode, bootstrap, multi-session, companion, session-start-hook, cross-session-messaging, settings-injection, cli-help, docs-site"
tier: L
related_specs: [SPEC-FACTORY-MODE-001, SPEC-KANBAN-BOOTSTRAP-001]
---

## HISTORY

- **v0.2.0** (2026-08-11) — Plan-phase audit revision (iteration 2). Closes plan-auditor findings D1-D5 from `review-1` (verdict FAIL 0.81 → targeting PASS ≥ 0.85 on re-audit). **D1 (blocking)**: §A.2 truth table completed to its 4 rows — the missing 4th row is `moai cc --name <role>-<run-id>` alone (no `-f`) → **no-op** (unchanged, no factory membership), a deliberate breaking change from `94025ce0a` where companion-shape `--name` alone entered companion mode; REQ-FB-001 amended to cover all four cases; AC-FB-027 added to pin the no-op. **D2**: design.md §7 typo `EnvMoai_FACTORY_LABEL` → `EnvMoaiFactoryLabel`. **D3**: §A.5 leader socket path source named as run-phase-derived through M3 (no producer exists today; `grep -rni 'socket\|LeadAddr\|leader_address' internal/factory/ internal/cli/factory.go internal/hook/session_start_factory.go` → 0 matches at HEAD `94025ce0a`); AC-FB-013(c) relaxed to assert a non-empty socket-path-shaped line. **D4**: AC-FB-016 fail-open sub-clause AC-FB-016a added for `factoryCompanionNotice` when `SplitCompanionLabel` returns `ok=false` or env vars are absent. **D5**: `related_specs:` frontmatter field retained (`moai spec lint` passes — sibling `SPEC-KANBAN-BOOTSTRAP-001` carries the same field, establishing repo convention); one-line note added to `progress.md` §E.1.

- **v0.1.0** (2026-08-11) — Initial plan-phase authoring. This SPEC is authored **on top of** commit `94025ce0a` ("feat(factory): announce companion session bootstrap from the SessionStart hook"), which landed a first implementation **out of order** — before any plan-phase artifacts for the multi-session bootstrap existed. That commit is recorded as prior art (§A.1) and is **revised, not reverted** by this SPEC. The predecessor `SPEC-FACTORY-MODE-001` (status: completed, v0.9.0) defined `-f` / `--factory` as a one-session **chain seed** (REQ-FM-001..006); this SPEC extends that surface into multi-session territory by redefining `-f` as **factory membership** whose role is disambiguated by the presence of `--name`. The unconditional bootstrap notice the prior-art commit emits is taken here as the "bell" fragment of a larger topology-config-gated design owned by `SPEC-KANBAN-BOOTSTRAP-001` (status: draft, deferred to the next release); the sibling owns the quorum, role-declaration carrier, and topology-config-gated dispatch protocol, and the boundary is recorded one-sidedly from this SPEC in §C. The five in-scope items are: (1) the `-f` redefinition, (2) `--settings` injection of `crossSessionInbound: accept`, (3) SessionStart notice revision, (4) CLI help text, and (5) a docs-site 4-locale page. The out-of-order landing is the reason this plan exists.

---

## §A. Context

### A.1 Prior art — commit `94025ce0a` (revised, not reverted)

The prior-art commit added 12 files / +724 lines and is the implementation this SPEC revises. It is prior art, not a deliverable of this SPEC. The files, by surface:

- **Entry switch** (`internal/cli/factory.go`) — `parseFactoryFlag`, `enterFactoryMode` (lead: sets `MOAI_FACTORY` + `MOAI_FACTORY_ID` + `MOAI_FACTORY_SPEC`), `enterFactoryCompanionMode` (companion: sets `MOAI_FACTORY_LABEL`, deliberately NOT `MOAI_FACTORY`), `parseCompanionLabel`, `rejectFactoryOnCG`, `recordFactorySession`, `captureEnvState`.
- **Vocabulary** (`internal/factory/bootstrap.go`) — `CompanionRoles = ["plan","run","review","sync"]`, `NewRunID` (base36 of Unix second), `CompanionLabel`, `SplitCompanionLabel`, `isCompanionRole`, `isRunIDShape`.
- **SessionStart hook** (`internal/hook/session_start_factory.go`) — `factoryBootstrapNotice`, `factoryLeadNotice` (prints run id + four companion commands), `factoryCompanionNotice` (prints "joined run X as the Y companion" — the role declaration this SPEC removes at §A.6).
- **Env keys** (`internal/config/envkeys.go`) — `EnvMoaiFactory`, `EnvMoaiFactorySpec`, `EnvMoaiFactoryID`, `EnvMoaiFactoryLabel`.
- **Dispatch wiring** (`internal/cli/cc.go`, `internal/cli/glm.go`) — both call `parseFactoryFlag` and `enterFactoryMode`/`enterFactoryCompanionMode` from the run path; the comments reading "SPEC-FACTORY-MODE-001: --factory / -f seeds a plan→run→verify→sync chain" (cc.go:91, glm.go:166) are the **current definition this SPEC revises**. Whether those comments are repointed to this SPEC at run-phase is a §B decision, not a spec-level one.
- **Block-cap inject** (`internal/cli/launcher_blockcap_infinite.go`) — `injectStopHookBlockCapForGoal` reads **either** `MOAI_FACTORY` **or** `MOAI_FACTORY_LABEL` and raises the cap unconditionally for either; this is the property REQ-FB-005 preserves.
- **Tests** — `factory_bootstrap_test.go`, `bootstrap_test.go`, `session_start_factory_test.go`, `launcher_blockcap_infinite_test.go` (the last carries the two AC003 preserve-tests, §A.7).

### A.2 The redefinition — "chain seed" → "factory membership"

`SPEC-FACTORY-MODE-001` defines `-f` / `--factory` as a one-session chain seed: one session carries the flag, and that session's orchestrator drives the whole `plan → run → verify → sync` chain. Multi-session factory operation needs the **same flag** to also mark a session as a factory member that is **not** driving the chain — a companion. The prior-art commit resolved this by giving companions a different env var (`MOAI_FACTORY_LABEL` instead of `MOAI_FACTORY`) but kept the dispatch in an `else if` structure (`cc.go:96-105`, `glm.go:169-176`) that short-circuits to the lead branch whenever `-f` is present, so a companion launched as `moai cc -f --name run-abc` is classified as a lead today and seeds a second chain. This SPEC redefines the flag:

| Launch form | Role | `MOAI_FACTORY` | `MOAI_FACTORY_LABEL` | Seeds chain? |
|---|---|---|---|---|
| `moai cc -f` (alone) | lead | set | unset | yes |
| `moai cc -f --name <role>-<run-id>` | companion | unset | set | no |
| `moai cc --name <non-companion>` (alone, no `-f`) | unchanged | unset | unset | no |
| `moai cc --name <role>-<run-id>` (alone, no `-f`) | unchanged | unset | unset | no |

The shape of `--name` is the discriminator **only when `-f` is present**: a companion-shape label (`<role>-<run-id>` where `role ∈ CompanionRoles` and `run-id` matches `isRunIDShape`) selects the companion branch when `-f` is also present; anything else leaves `--name` as claude's flag and passes it through untouched. When `-f` is absent, the launcher is a no-op regardless of `--name` shape — the two no-op rows collapse to one rule (see the breaking-change paragraph below).

#### A.2.1 Breaking change from `94025ce0a` — the 4th row

Under the prior-art commit `94025ce0a`, a companion-shape `--name` alone (`moai cc --name run-abc123`, no `-f`) entered companion mode — `MOAI_FACTORY_LABEL` was set, `enterFactoryCompanionMode(label)` was invoked, and the session joined the run as a companion. The `94025ce0a` commit message states this explicitly: "Companions launch under `--name <role>-<run-id>`".

Under this SPEC, that path is a **no-op**: no `MOAI_FACTORY*` env var is set, no factory session record is written, and `--name` is passed through to claude untouched. This is the sharpest edge of the "revised, not reverted" framing in the HISTORY v0.1.0 row — the `-f` redefinition to factory membership means companion entry **requires** `-f` (`-f --name <companion-shape>`), and the companion-shape-`--name`-alone path that was companion entry under `94025ce0a` is reclassified as a no-op. The operator intent is explicit: `--name` alone, regardless of shape, is untreated (operator instruction: "`--name` 단독 = 무처리"). AC-FB-027 pins this case; REQ-FB-001's no-`-f` clause governs it.

### A.3 Dispatch revision — evaluate both flags together

The current dispatch in `cc.go:96-105` (and the parallel `glm.go:169-176`):

```go
if specID, factoryEnabled, factoryArgs := parseFactoryFlag(filteredArgs); factoryEnabled {
    // lead branch — short-circuits
    defer enterFactoryMode(specID)()
    recordFactorySession(specID, factory.BackendClaude)
} else if label, isCompanion := parseCompanionLabel(filteredArgs); isCompanion {
    // companion branch — unreachable when -f present
    defer enterFactoryCompanionMode(label)()
}
```

is structurally incapable of recognizing a companion that also carries `-f`, because `parseFactoryFlag` returns `factoryEnabled=true` on any `-f` and the `else if` skips the companion-label parse. The revision evaluates both flags together — first parse `-f`, then parse `--name`, then select the branch on the combination — so a companion launched with both flags is correctly classified. The four-flag truth table in §A.2 is the requirement this dispatch implements (REQ-FB-001, REQ-FB-002, REQ-FB-003, REQ-FB-004).

### A.4 `crossSessionInbound` injection — why `--settings`

Inter-session communication in a factory run uses Claude Code's cross-session messaging (`ListAgents` / `SendMessage`). The `crossSessionInbound` settings field controls whether an inbound message is `accept`-ed, `hold`, or `refuse`-d by the receiving session; the **stricter tier wins** between project and local settings, so an operator whose `.claude/settings.json` or `settings.local.json` carries `crossSessionInbound: hold` (or leaves the field absent, which is the runtime default-hold behavior) cannot relax it from the project settings layer. Measured in this worktree: `crossSessionInbound` appears in **0** `internal/template/templates/` files and **0** Go source files (`grep -rn crossSessionInbound internal/ pkg/ cmd/` → 0); `--settings` appears in moai code in **0** places (the two `--settings` matches are in `moai-foundation-cc/reference/*.md` Claude Code reference docs, not in moai Go code). The accept/hold/refuse ladder therefore cannot be satisfied from any settings layer moai writes today.

The injection path is a **transient settings file** (written to a session-private tempdir, not to `.claude/` or `.moai/`) passed to the launched `claude` via `--settings <file>`. The `--settings` flag is documented to take the **strictest** merge of its argument with the project/local settings, which is precisely the property needed: it lets the launcher force `crossSessionInbound: accept` regardless of what an operator's local file says, without mutating the operator's persistent settings. The operator-supplied `--settings` case (REQ-FB-007) honors an explicit `--settings <file>` already on the command line by **not** injecting moai's own — the bootstrap notice then instructs the operator to verify `accept` is present in their file, because two `--settings` flags would collide and the operator's intent wins.

### A.5 SessionStart notice — lead content, companion content, and the role-name removal

The lead notice (REQ-FB-008, REQ-FB-009, REQ-FB-011) carries:

1. The run id (already printed by `factoryLeadNotice` today).
2. The **four companion launch lines**, each now including `-f` so the operator copies a factory-membership command, not a bare `--name` command (today's notice omits `-f`).
3. The **leader socket path** — the address companions send messages to on the cross-session messaging substrate (Claude Code's `ListAgents` / `SendMessage` API). No moai code produces this string today (`grep -rni 'socket\|LeadAddr\|leader_address' internal/factory/ internal/cli/factory.go internal/hook/session_start_factory.go` → 0 matches at HEAD `94025ce0a`); the M3 notice milestone either reads it from the platform API at SessionStart time or has the launcher capture it into a new env var (e.g. `MOAI_FACTORY_LEAD_ADDR`) when `enterFactoryMode` classifies a lead. The string's concrete producer is therefore a run-phase decision scoped to M3, not a spec-level name. AC-FB-013(c) asserts a non-empty socket-path-shaped line rather than naming an env var.
4. An **inbound-automation notice**: a one-line statement that cross-session messages are auto-accepted via the injected `--settings`, so the operator does not expect to see hold prompts.
5. The **SPEC identifier** — printed ONLY when `MOAI_FACTORY_SPEC` is set. The flag accepts an optional SPEC argument today (`parseFactoryFlag` returns `specID`); when the operator passed none, the notice omits the SPEC line rather than printing an empty placeholder.

The companion notice (REQ-FB-010) is **join-only and role-less**: it prints that the session joined run X, and **does not** print "as the Y companion". The current `factoryCompanionNotice` (session_start_factory.go:63-69) prints the role, and that print is removed here.

### A.6 Collision-avoidance rationale — why the companion notice must not name the role

`SPEC-KANBAN-BOOTSTRAP-001` carries `REQ-KS-006`, which widens the role-declaration contract from "a stable label set at launch" to "addressability **plus role declaration**, with the declaration resolvable from a session that is not the `lead`" (sibling §A.13, v0.3.0). That sibling is at the Tier L ceiling of 25 REQ and is **deferred** — when it lands, it owns the role-declaration carrier. If this SPEC's companion notice continues to print the role, two sources of role truth would exist at bootstrap: the unconditional notice here and the topology-config-gated declaration the sibling ships. They would not, in general, agree — the sibling's declaration is runtime-resolvable and may differ from the launch-time label. Removing the role from this SPEC's notice leaves the companion notice as a pure membership ack and lets the sibling's eventual declaration be the sole role truth. The forward reference is in §C.

### A.7 AC003 regression prohibition (PRESERVE)

The two block-cap acceptance tests in `internal/cli/launcher_blockcap_infinite_test.go` MUST stay green:

- `TestAC003_LauncherInjectsRaisedBlockCapForInfiniteGoal`
- `TestAC003_BlockCapDoctrineClauseSpecific`

They assert the doctrine clause documented at `launcher_blockcap_infinite.go` (the unconditional factory-branch raise sits ahead of the goal read). The `-f` redefinition changes **which branch a companion takes**, not **whether the cap is raised for a companion** — the cap raise for the label-only path is preserved by REQ-FB-005. These two tests are the regression guard and are called out as PRESERVE in `plan.md` §A.5 with their package path.

### A.8 CLI help — `Use` / `Long`, not flag registration

`moai cc` and `moai glm` both set `DisableFlagParsing: true` because they forward arbitrary args to `claude` / `glm`. A cobra `cmd.Flags()` registration is therefore silently **inert** — the flag is never parsed by cobra and the registration only misleads a reader into thinking the flag surface lives there. The documentation lives in the `Use:` field and the `Long:` / Flags prose block, which is what `moai cc --help` prints. Today `Use: "cc [-p profile] [-- claude-args...]"` mentions neither `-f` nor `--factory`; the revision requires both `-f` (lead) and `-f --name <role>-<id>` (companion) to appear in the help text (REQ-FB-012, REQ-FB-013, REQ-FB-014).

### A.9 docs-site — section decision

The new Factory Mode page lands under `docs-site/content/{en,ko,ja,zh}/multi-llm/factory-mode.md`. Justification (in `plan.md` §B.1): `multi-llm/` is the section for backend-mixing sessions (`moai cg`, GLM, tmux Agent Teams) and a factory run is exactly that — a lead plus companions, any of which may run on either backend (the lead notice itself says "Substitute 'moai glm' for 'moai cc' on any companion"). Placing it under `multi-llm/` adds a page to an existing section, so no new section is created, no `_meta.yaml` section-weight is introduced, and the `layouts/partials/menu.html:28-44` SVG switch needs no new case (the `multi-llm` icon already exists in the switch). `advanced/` was the alternative; it is rejected because `advanced/` carries single-session advanced patterns (worktrees, autonomy loops, deep configuration) and a multi-session page would be the odd one out. The page frontmatter is the new-page contract: `title`, `weight`, `draft:false` across all four locale files, plus a `sub:` entry under the `multi-llm` section in `data/menu/main.yaml` carrying a 4-locale `name` map and the `ref` (REQ-FB-015, REQ-FB-016, REQ-FB-017).

---

## §B. Requirements (GEARS notation)

### REQ-FB-001 — `-f` is factory membership, role disambiguated by `--name`

**Where** the launcher is `moai cc` or `moai glm`, **When** the operator passes `-f` / `--factory` and no companion-shape `--name` is present, the launcher SHALL classify the session as a **lead**. **Where** the launcher is `moai cc` or `moai glm`, **When** the operator passes `-f` / `--factory` and a companion-shape `--name` is present, the launcher SHALL classify the session as a **companion**. **Where** the launcher is `moai cc` or `moai glm`, **When** the operator does **not** pass `-f` / `--factory`, the launcher SHALL leave the session unchanged regardless of `--name` shape — no `MOAI_FACTORY*` env var is set, no factory session record is written, and `--name` is passed through to the launched backend untouched (the companion-shape-`--name`-alone case that was companion entry under `94025ce0a` is reclassified as a no-op by this clause; see §A.2.1). The §A.2 truth table enumerates the four cases this requirement governs.

### REQ-FB-002 — Both flags evaluated together, no short-circuit

**Where** the dispatch site is `cc.go` or `glm.go`, the launcher SHALL evaluate `-f` / `--factory` and `--name` **together** rather than short-circuiting to the lead branch on `-f` present, so that a session launched with both `-f` and a companion-shape `--name` reaches the companion branch.

### REQ-FB-003 — Companion does NOT seed the chain

**When** the launcher classifies a session as a companion ( REQ-FB-001 companion case ), the launcher SHALL set `MOAI_FACTORY_LABEL` and SHALL NOT set `MOAI_FACTORY`, `MOAI_FACTORY_ID`, or `MOAI_FACTORY_SPEC`, so the companion does not seed a `plan → run → verify → sync` chain.

### REQ-FB-004 — Lead still seeds the chain

**When** the launcher classifies a session as a lead ( REQ-FB-001 lead case ), the launcher SHALL set `MOAI_FACTORY`, `MOAI_FACTORY_ID`, and (when a SPEC argument was passed) `MOAI_FACTORY_SPEC`, preserving `SPEC-FACTORY-MODE-001` REQ-FM-001's chain-seed behavior.

### REQ-FB-005 — Block-cap raise preserved for both branches

**Where** `MOAI_FACTORY` is set (lead) OR `MOAI_FACTORY_LABEL` is set (companion), the block-cap inject (`launcher_blockcap_infinite.go`) SHALL raise `CLAUDE_CODE_STOP_HOOK_BLOCK_CAP` to `DefaultRaisedStopHookBlockCap`, preserving the prior-art unconditional behavior measured at HEAD `94025ce0a`.

### REQ-FB-006 — `crossSessionInbound: accept` injected via transient `--settings`

**When** the launcher classifies a session as a lead or companion AND the operator did NOT pass `--settings <file>` on the command line, the launcher SHALL write a transient settings file containing `{"crossSessionInbound": "accept"}` to a session-private tempdir and SHALL pass it to the launched `claude` / `glm` via `--settings <file>`, so inbound cross-session messages are auto-accepted regardless of the operator's project/local settings.

### REQ-FB-007 — Operator-supplied `--settings` honored

**When** the operator passed `--settings <file>` on the command line, the launcher SHALL NOT inject its own settings file, and the lead/companion bootstrap notice SHALL print an advisory instructing the operator to verify `crossSessionInbound: "accept"` is present in their supplied file.

### REQ-FB-008 — Lead notice content

**When** the SessionStart hook runs in a lead session, `factoryLeadNotice` SHALL emit, in order: (a) the run id; (b) the four companion launch lines, each carrying `-f` and the companion-shape `--name <role>-<run-id>`; (c) the leader socket path; (d) a one-line inbound-automation notice stating that cross-session messages are auto-accepted via the injected `--settings`.

### REQ-FB-009 — SPEC identifier conditional in notice

**Where** `MOAI_FACTORY_SPEC` is set, `factoryLeadNotice` SHALL append the SPEC identifier to the notice; **where** `MOAI_FACTORY_SPEC` is unset, the notice SHALL omit the SPEC line entirely rather than printing an empty placeholder.

### REQ-FB-010 — Companion notice is join-only and role-less

**When** the SessionStart hook runs in a companion session, `factoryCompanionNotice` SHALL emit a join acknowledgement that names the run id and SHALL NOT name the role, removing the prior-art `"as the %s companion"` clause to avoid colliding with the role-declaration contract the deferred `SPEC-KANBAN-BOOTSTRAP-001` `REQ-KS-006` owns.

### REQ-FB-011 — Companion launch lines carry `-f`

**Where** `factoryLeadNotice` enumerates the four companion launch lines, each line SHALL be of the form `moai cc -f --name <role>-<run-id>` (or `moai glm -f --name <role>-<run-id>`), so the operator copies a factory-membership command rather than the bare `--name` form the prior-art notice prints today.

### REQ-FB-012 — CLI help documents lead entry

**Where** the command is `moai cc` or `moai glm`, the `Use:` field and the `Long:` / Flags prose SHALL document `-f` / `--factory` as the lead entry switch and SHALL state that it seeds a `plan → run → verify → sync` chain in the launched session.

### REQ-FB-013 — CLI help documents companion entry

**Where** the command is `moai cc` or `moai glm`, the `Use:` / `Long:` / Flags prose SHALL document `-f --name <role>-<run-id>` as the companion entry form, SHALL enumerate the four roles (`plan`, `run`, `review`, `sync`), and SHALL state that a companion joins an existing run without seeding a chain.

### REQ-FB-014 — Help text lives in `Use` / `Long`, not flag registration

**Where** `DisableFlagParsing: true` is set on the command, the launcher SHALL carry the `-f` / `--factory` documentation in the `Use:` / `Long:` prose only and SHALL NOT register `-f` / `--factory` via `cmd.Flags()`, because a flag registration under disabled parsing is silently inert and misleads readers about the parse site.

### REQ-FB-015 — docs-site 4-locale Factory Mode page

**Where** the docs-site tree at `docs-site/content/{en,ko,ja,zh}/multi-llm/` exists, the publisher SHALL add a `factory-mode.md` page in each of the four locales, each carrying frontmatter `title`, `weight`, and `draft:false`, with body content covering the lead entry, the four companion entries, the multi-session bootstrap flow, and the cross-session-messaging communication substrate.

### REQ-FB-016 — Menu entry with 4-locale `name` map

**Where** `docs-site/data/menu/main.yaml` carries the `multi-llm` section's `sub:` list, the publisher SHALL add a `factory-mode` entry whose `name:` carries all four locale keys (`ko`, `en`, `ja`, `zh`) and whose `ref:` points at `/multi-llm/factory-mode`.

### REQ-FB-017 — No new docs-site section

**Where** the page is placed under `multi-llm/`, the publisher SHALL NOT create a new docs-site section, SHALL NOT add a new `_meta.yaml` section weight, and SHALL NOT require a new `layouts/partials/menu.html:28-44` SVG case, because the `multi-llm` icon already exists in the switch.

### REQ-FB-018 — AC003 tests preserved green

**While** the `-f` redefinition and dispatch revision are implemented, the two tests `TestAC003_LauncherInjectsRaisedBlockCapForInfiniteGoal` and `TestAC003_BlockCapDoctrineClauseSpecific` in `internal/cli/launcher_blockcap_infinite_test.go` SHALL continue to pass, guarding the unconditional block-cap raise measured at HEAD `94025ce0a`.

---

## §C. Out of Scope

The following are **NOT** owned by this SPEC. They belong to `SPEC-KANBAN-BOOTSTRAP-001` (status: draft, deferred to the next release) and are recorded here as a one-sided boundary from this SPEC — the sibling's own §C is the authoritative side and is not edited from here.

### Out of Scope — Topology-config-gated quorum and dispatch

- The shared work queue, the worker spawn automation, the failure-detection / liveness watch, and the board model are owned by the kanban sibling family (`SPEC-KANBAN-BOARD-001`, `SPEC-KANBAN-WORKTREE-001`, `SPEC-KANBAN-BOOTSTRAP-001`).
- The topology-config-gated quorum wait (`REQ-KS-007`, `REQ-KS-012`) — the entry switch printing guidance and waiting for a configured quorum before proceeding — is owned by `SPEC-KANBAN-BOOTSTRAP-001`.
- The role-declaration carrier and runtime role resolution (`REQ-KS-006` widened at sibling v0.3.0) is owned by the sibling; this SPEC deliberately **removes** the role name from its own companion notice (REQ-FB-010) to avoid a second source of role truth.
- The dispatch protocol (`REQ-KS-018`, `REQ-KS-019`) — lead routes work to the session whose declared role owns the column — is sibling territory.
- The topology-config-gated **multi-backend** command emission (`REQ-KS-013`) — emitting backend-specific launch commands conditioned on a topology config — is sibling territory.

### Out of Scope — Forward reference to the sibling's supersedence

**When** `SPEC-KANBAN-BOOTSTRAP-001` lands, the sibling's topology-config-gated guidance SHALL supersede the unconditional bootstrap notice this SPEC ships (REQ-FB-008, REQ-FB-010, REQ-FB-011). The notice **emit mechanism** — the SessionStart hook surface, the `factoryBootstrapNotice`/`factoryLeadNotice`/`factoryCompanionNotice` function trio, and the env-var-driven branch selection — is **consumed** (not re-authored) by the sibling at that point. This SPEC's notice is unconditional because the topology config does not yet exist; the sibling's notice is conditional on that config, and the unconditional → conditional upgrade is the sibling's deliverable, not a rework of this SPEC's mechanism.

### Out of Scope — Run-phase decisions

- Whether the `cc.go:91` / `glm.go:166` comments naming `SPEC-FACTORY-MODE-001` are repointed to this SPEC is a run-phase documentation decision (the comment describes the chain-seed behavior, which REQ-FB-004 preserves, so the comment is not stale per se — repointing is editorial).
- The transient settings file's exact path under `os.TempDir()` is a run-phase implementation detail, provided it is session-private and cleaned up on session exit.

### Out of Scope — Single-session factory mode

The single-session chain-seed behavior (`SPEC-FACTORY-MODE-001` REQ-FM-001..006) is the predecessor's territory and is **preserved** (REQ-FB-004), not redefined. This SPEC's redefinition is strictly on the multi-session axis: the lead is still the chain seed; the companion is the new factory-membership form.

---

## §D. Constraints

- **C1 — Template-First.** Any `internal/template/templates/` edit MUST be mirrored to the template source and `make build` MUST be run. The CLI help text and the docs-site page both touch template-managed surfaces; the mirror is scoped in `plan.md` §F.
- **C2 — Template neutrality (CLAUDE.local.md §25).** SPEC IDs, REQ tokens, audit citations, internal dates, and commit SHAs MUST NOT leak into `internal/template/templates/`. The docs-site page is user-facing documentation; its body carries no `REQ-FB-*` tokens, no `94025ce0a` SHA, and no `SPEC-FACTORY-BOOTSTRAP-001` identifier outside this SPEC's own artifacts.
- **C3 — Sibling off-limits.** `.moai/specs/SPEC-KANBAN-*` files MUST NOT be edited from this SPEC. The boundary in §C is stated from this side only.
- **C4 — AC003 preserve.** The two named block-cap tests in `internal/cli/launcher_blockcap_infinite_test.go` MUST stay green (REQ-FB-018).
- **C5 — Commits only.** No push, no PR. Work stays in the `feat/factory-bootstrap-guidance` branch inside this worktree.
- **C6 — Worktree isolation.** All artifacts land at `.moai/specs/SPEC-FACTORY-BOOTSTRAP-001/` relative to this worktree's root. Nothing is written to the primary checkout.
- **C7 — Orchestrator-only AskUserQuestion.** No `AskUserQuestion` invocation in CLI / hook code (C-HRA-008). The bootstrap notice is informational stdout; it does not prompt.
- **C8 — Fail-open throughout.** The SessionStart notice and the settings-file injection are fail-open: a transient-file write failure or an env-read anomaly degrades to emitting nothing (notice) or to launching without the injected settings (injection), never to blocking the session start.

---

## §E. Cross-References

- `SPEC-FACTORY-MODE-001` (completed, v0.9.0) — predecessor; defines `-f` / `--factory` as chain seed (REQ-FM-001..006), REQ-FM-023 (unconditional factory block-cap branch), REQ-FM-006 (`moai cg` rejection). This SPEC preserves the lead's chain-seed behavior.
- `SPEC-KANBAN-BOOTSTRAP-001` (draft, deferred) — sibling; owns topology-config-gated quorum, role-declaration carrier, dispatch protocol. This SPEC takes only the "bell" fragment (notice emit mechanism) and is consumed by the sibling when it lands.
- `SPEC-INFINITE-GOAL-001` — owns the raised Stop-hook block cap (`CLAUDE_CODE_STOP_HOOK_BLOCK_CAP`); this SPEC preserves the unconditional factory-branch raise that sits ahead of the goal read (REQ-FB-005, REQ-FB-018).
- `.claude/rules/moai/development/spec-frontmatter-schema.md` — canonical 12-field frontmatter schema; this SPEC's frontmatter conforms.
- `CLAUDE.local.md` §2 (Template-First), §15 (16-language neutrality), §25 (template content isolation) — constraint sources for C1/C2.
- `verification-claim-integrity.md` §1.1 surface 3 — every measured fact in §A is cited against its baseline (HEAD `94025ce0a` in this worktree); no claim here is inferred from text-pattern matching alone.

---

## §F. Acceptance Criteria Summary

Acceptance criteria (Given-When-Then) live in `acceptance.md`. The §D matrix there enumerates one or more `AC-FB-XXX` per REQ above, mapped by REQ ID and severity. This §F is a pointer; the canonical AC enumeration is `acceptance.md` §D.

---

## §G. Open Questions

None at plan-phase. The docs-site section decision (§A.9) is resolved here, not deferred. Run-phase decisions are scoped as Out of Scope in §C, not as open questions.

---

## §H. Status

`status: draft` — set at plan-phase artifact creation. The next transition (`draft → in-progress`) is owned by `manager-develop` per the Status Transition Ownership Matrix; this agent performs only the `(none) → draft` transition.
