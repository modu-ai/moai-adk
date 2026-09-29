---
id: SPEC-ALIAS-PASSTHROUGH-001
title: "Pass profile model aliases through to Claude Code --model verbatim — fix the compile-time alias-snapshot substitution in moai cc"
version: "0.1.0"
status: draft
created: 2026-09-29
updated: 2026-09-29
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: internal/cli
lifecycle: spec-anchored
tags: "model-alias, launcher, expandModelString, passthrough, claude-code, glm-slot"
era: V3R6
tier: M
related_specs: [SPEC-SONNET55-BUMP-001]
issue_number: 1730
---

# SPEC-ALIAS-PASSTHROUGH-001 — Alias passthrough to `claude --model`

## HISTORY

| Version | Date | Change |
|---|---|---|
| 0.1.0 | 2026-09-29 | Initial draft. Card t1315, GitHub #1730 (Class C design change, Tier M). Tree measured at `02ad57bbe`. Card anchors corrected against this tree (§A.3) — the issue's line numbers predate the t1322 and later merges. |

## §A Background

`moai cc` does not hand a profile model alias such as `opus[1m]` to Claude Code. The launcher
resolves it through the compile-time `template.ModelAliasTable` and launches
`claude --model <canonical-id>[1m]`. The table is a snapshot of one moment, so whenever Claude
Code moves an alias to a newer model, `moai cc` keeps launching the older model until a MoAI
release bumps the table. A user who picks the alias precisely to track the newest model does not
get it (GitHub #1730).

Card t1322 (SPEC-SONNET55-BUMP-001) already bumped the table to Opus 5.5 and is merged into this
tree's base (`02ad57bbe`), so today's visible symptom is gone — for now. This SPEC's essence is
fixing the **root mechanism** (the snapshot substitution itself), NOT another table bump. The
next upstream alias move reopens the gap until another bump ships; that recurrence is the defect
this SPEC removes.

### A.1 Where the substitution actually lives (measured at `02ad57bbe`)

| Surface | Evidence |
|---|---|
| Alias→ID substitution | `internal/cli/launcher.go` `expandModelString` (:1142-1155): looks the base up in `template.ModelAliasTable` (:1147) and replaces it with the canonical id; the `[1m]` suffix is preserved across substitution via `splitModelSuffix` (:1160-1166). Unknown values and full ids pass through unchanged (:1149). |
| Launch call site | `launcher.go` :711 `model = resolveMainSessionModel(model, glmBackend)`. The model value originates from `prefs.Model` → `settings["DO_CLAUDE_MODEL"]` (:631-633, raw alias — no expansion there) or a `--model` extraArg (:672-676, incl. the `-m` short form; `--model=` value form :679-680). `buildArgs` emits `--model <model>` (:735-737); the process is handed to `claudeBin` via `execOrSpawnClaudeFunc` (:838). |
| Claude vs GLM branch | `resolveMainSessionModel` (:1235-1248): Claude backend → `expandModelString` (:1237, the defect); GLM backend → reverse-maps canonical ids to slot aliases via `template.ModelAliasFromCanonicalID` (:1243) because z.ai's Anthropic-compat shim binds `--model` to slot aliases through `ANTHROPIC_DEFAULT_*_MODEL` env (:1226-1234). |
| Alias SSOT table | `internal/template/model_policy.go` `ModelAliasTable` (:80-86): `"opus": ModelIDOpus55` (:81, the t1322 bump), `"sonnet": ModelIDSonnet55` (:82), `"fable"`, `"haiku"`, `"opusplan"` (self-mapped CC-native routing alias :85). `ModelIDOpus55 = "claude-opus-5-5"` (:42). `ModelDeprecatedCanonicalIDs` (:97-104) maps superseded ids back to aliases. `ModelAliasPickerValues` (:144-156) exposes `fable[1m]`, `opus[1m]`, `sonnet[1m]`, `haiku` (1M unification — bare opus/sonnet/fable no longer offered but remain valid stored values). |
| Table consumers (display/validation/legacy) | `profile_setup.go` `normalizeModel` (:65) + `promoteTo1M` (:99) + `normalizeModelLegacy1M` (:113) — prefs migration, legacy-id→alias; `internal/web/validate.go` `modelOptionList` (:48) — validation union of picker values + table keys + `[1m]` variants; `internal/cli/schema_bridge.go` `schemaOptionBridge` (:118-135) — TUI option labels; `internal/template/glm_effort_overlay.go` (:277-285) — alias-keyed GLM slot switch. All are alias-keyed and independent of the launch substitution. |
| Non-Claude launchers | `internal/cli/codex_launcher.go` carries no model expansion (only a usage-string mention of `--model o3` as a passthrough example, :627). GPT gateway launcher was withdrawn (card t857). |
| Test pinning of the defect | `internal/cli/launcher_test.go` `TestExpandModelString` (:674-713) asserts `expandModelString("opus")` → `template.ModelIDOpus55` and `("opus[1m]")` → `ModelIDOpus55+"[1m]"`; `TestResolveMainSessionModel_GLMAvoidsCanonicalID` (:926-959) case "claude backend alias expands to canonical id" (:938) pins the same substitution. Both run GREEN on this tree — the defective behavior is live and test-pinned. |

### A.2 Observed defect evidence on this tree (plan-phase measurement)

Command: `go test -run '^(TestExpandModelString|TestResolveMainSessionModel_GLMAvoidsCanonicalID)$' ./internal/cli/ -v`

Observed (verbatim, this run, tree `02ad57bbe`): all 23 subtests PASS, including
`--- PASS: TestExpandModelString/opus_alias_1m_resolves ` (asserts
`expandModelString("opus[1m]") == template.ModelIDOpus55 + "[1m]"`, i.e. the launcher substitutes
the compile-time id) and `--- PASS: TestResolveMainSessionModel_GLMAvoidsCanonicalID/claude_backend_alias_expands_to_canonical_id `
(asserts `resolveMainSessionModel("opus", false) == template.ModelIDOpus55`).

Exit code 0. This is the substitution mechanism working as coded — and the defect: the id it
substitutes is frozen at MoAI compile time, while Claude Code resolves the same alias to whatever
its current target is. The issue's own probe (CC 2.1.283, `claude -p --model 'opus[1m]'`)
reports `modelUsage` for `claude-opus-5-5[1m]`; given `claude-opus-5[1m]` it reports
`claude-opus-5[1m]` — the alias and the pinned id demonstrably diverge once upstream moves.

### A.3 Card-anchor corrections (tree wins)

| Issue claimed (v3.1.2 / develop `c000a1fcb`) | This tree (`02ad57bbe`) says |
|---|---|
| `expandModelString` at `launcher.go:1136`, alias lookup at `:1141` | `:1142` / `:1147` (later merges shifted the block) |
| "The profile model reaches it at `:1231`" | Prefs model reaches `resolveMainSessionModel` at the launch call site `:711`; the `expandModelString` call inside it is `:1237` |
| `ModelIDOpus55` at `model_policy.go:52`, `"opus"` row at `:78-79` | `:42` / `:81` |
| "On v3.1.2 the same row maps to `ModelIDOpus5`" | Row already maps to `ModelIDOpus55` (t1322 merged into this base) — symptom gone today, mechanism unchanged |

A second sweep after the plan-audit verdict (finding F-3) corrected six in-SPEC citations that had drifted ±3-5 lines — the extraArgs parse, `normalizeModel`, `promoteTo1M`, `normalizeModelLegacy1M`, `modelOptionList`, and the profile_setup `splitModelSuffix` consumers. §A.1 and REQ-ALP-010 carry the corrected values, re-measured on this tree.

### A.4 Claude Code's own alias semantics (the authority being delegated to)

The issue verified (`claude -p --model 'opus[1m]' --output-format json`, CC 2.1.283) that Claude
Code accepts the alias directly and resolves it to its current target, `[1m]` suffix included.
Claude Code natively supports the `[1m]` context-window modifier on aliases and full ids alike
(`splitModelSuffix`'s doc comment at `launcher.go:1157-1159` states this). Claude Code also
validates its own `--model` value: a genuinely invalid value errors visibly at launch. MoAI
therefore holds no information about alias semantics that Claude Code does not hold fresher.

## §B Requirements (GEARS)

- **REQ-ALP-001** (Ubiquitous) The Claude-backend launch path shall pass the resolved model string to `claude --model` verbatim, with no compile-time `ModelAliasTable` substitution anywhere on that path.
- **REQ-ALP-002** (Ubiquitous) Every stored or flagged model form shall pass through unchanged on the Claude-backend path: base aliases (`opus`, `sonnet`, `fable`, `haiku`, `opusplan`), their `[1m]` variants (the picker surface), full canonical ids (current and deprecated), unknown values, and the empty string (which emits no `--model` flag at all, byte-identical to today).
- **REQ-ALP-003** (Event-driven) When the launch backend is GLM, `resolveMainSessionModel` shall keep reverse-mapping canonical ids (current and deprecated) to slot aliases and shall keep passing alias values through unchanged — the GLM branch's behavior is byte-identical to today's.
- **REQ-ALP-004** (Ubiquitous) The alias→id substitution role of `expandModelString` shall be removed; the function shall have no production caller and shall be deleted rather than left as dead code. `splitModelSuffix` shall be retained (the GLM branch and `profile_setup.go` still consume it).
- **REQ-ALP-005** (Ubiquitous) `ModelAliasTable` shall remain the SSOT for every non-launch consumer — wizard picker values (`ModelAliasPickerValues`), prefs normalization (`normalizeModel`), web/settings validation unions (`modelOptionList`, `schemaOptionBridge`), and the GLM reverse mapping (`ModelAliasFromCanonicalID`) — and no consumer of those surfaces shall change behavior.
- **REQ-ALP-006** (Event-driven) When a passthrough acceptance test asserts its expectation, the assertion shall be table-independent: it shall reference the input string round-tripping (or a literal), never `ModelIDOpus55`, `ModelIDSonnet55`, `ModelAliasTable`, or any derived table value — so the next upstream alias move cannot re-break the regression proof.
- **REQ-ALP-007** (Ubiquitous) Stale documentation shall be updated in the same change: the launcher comment stating "under a Claude backend short aliases expand to canonical ids as before (byte-identical to expandModelString)" (:701-704), the `resolveMainSessionModel` doc comment (:1224-1234), the `ModelAliasTable` forward-direction note naming `expandModelString` as consumer (`model_policy.go` :68-69), and the `@MX:ANCHOR` reason line listing `expandModelString` (:79).
- **REQ-ALP-008** (Event-driven) When the Claude-backend launch path receives an unknown or legacy token, it shall pass it through unchanged and shall emit no new warning; legacy-id→alias mapping stays owned by the prefs-normalization layer (`normalizeModel`) and the GLM reverse map, never by the launch path.
- **REQ-ALP-009** (Event-driven) When the launch reproduction harness runs (scratch directory holding an empty `.moai/`, isolated profile base, a stub `claude` binary placed ahead of any real one), the captured `claude` argv shall carry the stored profile alias verbatim as the value of `--model`.
- **REQ-ALP-010** (Ubiquitous) The `settings.local.json` `DO_CLAUDE_MODEL` channel (:631-633, raw value, no expansion) and the `--model` extraArgs override (:672-676, still overrides prefs, still passes through) shall remain byte-identical in behavior.

## §C Decisions

- **D-1 (alias passthrough — chosen)** On the Claude-backend launch path, `--model` receives the stored/flagged value verbatim; Claude Code resolves alias semantics at launch time, `[1m]` suffix included. The user who picks an alias tracks Claude Code's current target for that alias without waiting for a MoAI release; the user who wants a fixed model sets a full id, which already passes through unchanged. Rationale: Claude Code is the authority on what its aliases mean, and it accepts them directly (§A.4); MoAI's snapshot can only lag.
  **Rejected — snapshot-refresh** (another table bump now, or a live-fetched/refreshed table at launch): a bump re-fixes today's symptom and reintroduces the defect at the next upstream move (the recurrence IS the defect); a live refresh adds a lookup or network dependency to the launch path to recover information Claude Code already has. **Rejected — hybrid** (recognize known aliases, substitute only unknowns, or vice versa): alias recognition needs a second alias vocabulary that itself goes stale — a new instance of the exact defect class being removed — while buying nothing passthrough does not already provide, since unknown values already pass through today.
- **D-2 (passthrough is total on the Claude path)** All forms pass through verbatim (REQ-ALP-002): base aliases, `[1m]` variants, `opusplan`, full ids, unknown values. No alias recognizer is added. Known behavior deltas, both intended: (a) `opus[1m]` now reaches Claude Code as `opus[1m]` instead of `claude-opus-5-5[1m]`; (b) a bare `opus` (a legacy stored value the picker no longer offers but still accepts) reaches Claude Code verbatim instead of the table's id — under Claude Code's 1M unification its bare alias targets the same model family, and any future drift is now Claude Code's live truth, which is the point of the fix.
- **D-3 (unknown/legacy = silent passthrough)** Unknown and legacy tokens pass through unchanged with no new warning (REQ-ALP-008). Rejected warn-and-pass: unknown-passthrough is long-established pinned behavior (`custom-xyz` cases), legitimate custom gateway model names would trigger the warning noise, and a warning carries no actionable information — Claude Code errors visibly on genuinely invalid values. Rejected launch-path legacy-id reverse mapping: a user who pinned `claude-opus-4-8` wants that fixed model; mapping it to the alias would silently unpin them. The wizard migration (`normalizeModel`) already normalizes legacy ids to aliases at the prefs layer, which is the correct home for that mapping.
- **D-4 (non-Claude backends unchanged)** The GLM branch keeps its canonical→slot-alias reverse map (REQ-ALP-003): z.ai's shim binds `--model` to slot aliases through `ANTHROPIC_DEFAULT_*_MODEL` env, and a literal `claude-*` id there bypasses slot routing — the GLM branch is already alias-passthrough and stays as-is. The codex launcher has no model expansion to change. `DO_CLAUDE_MODEL` and `--model` extraArgs behavior unchanged (REQ-ALP-010).
- **D-5 (the table keeps its non-launch jobs)** `ModelAliasTable` remains SSOT for display, validation, prefs migration, and GLM slots (REQ-ALP-005); launch-time validation is Claude Code's job. No validation loosens or tightens: `modelOptionList`'s union (picker + table keys + `[1m]` variants) still gates what the wizard/web surfaces accept as stored values, and the launch path never consults the table.
- **D-6 (t1322 coordination, closed)** SPEC-SONNET55-BUMP-001's REQ-SSB-011/D-4 anticipated this SPEC absorbing the table contents after it; t1322 is merged into this base (`02ad57bbe`), so the clause resolves as "t1315 lands on top, mechanism change only". No table-content edit occurs in this SPEC; `ModelIDOpus55` and the `"opus": ModelIDOpus55` row remain exactly as t1322 landed them.

## §D Non-Functional Constraints

- Table-independence invariant: passthrough assertions never reference table constants (REQ-ALP-006). This is the property that makes the regression proof survive the next upstream alias move.
- No local full-suite run (gitflow lane discipline): affected packages only — `./internal/cli/... ./internal/template/... ./internal/web/...` — with `-timeout 30m`; CI owns the full-suite verdict.
- Cross-platform: the launcher is POSIX-exec + Windows spawn (`execOrSpawnClaudeFunc` build-tagged); the reproduction harness must not depend on POSIX-only shell semantics for its core assertion (the exec-seam capture is portable; any true-binary PATH-stub variant is POSIX-guarded and supplementary).
- Go code/comments in English; error wrapping via `%w`; no new hardcoded model-id literals outside `model_policy.go` (the existing id-hardcoding discipline).
- No commit/push by the implementing agent inside the card worktree outside the lane's integration window (card worktree discipline; the lane commits).

## §E Research Items

- **E.1 Claude Code accepts `opus[1m]` directly** — VERIFIED by the issue reporter (CC 2.1.283, `claude -p --model 'opus[1m]' --output-format json` → `modelUsage` for `claude-opus-5-5[1m]`), and independently by this card's dispatch ("verified for `opus[1m]`"). Full-id input already passes through today and is unaffected.
- **E.2 Bare aliases remain valid Claude Code input** — VERIFIED at CLI level by the live probe (this plan-phase session, F-1 discharge; coordinator observed the identical outcome): `claude -p --model 'opus' --output-format json "Reply with ok."` → exit 0, `"is_error": false`, `"subtype": "success"`, `"result": "ok"` — Claude Code does NOT reject a bare alias at launch. **Gateway caveat**: this lane runs behind a GLM gateway (`ANTHROPIC_BASE_URL` set) and `modelUsage` recorded `glm-5.3-flash` (canonicalModel `glm-5.3-flash`, provider `firstParty`), so the **Anthropic-native** alias→model mapping was NOT observed — the probe establishes CLI-level acceptance only. Residual: if the Anthropic-native side pins bare `opus` to an old model, a legacy stored bare-alias value launches that old model after this change (recorded in acceptance.md §D.2; mitigations: the documented fixed-model path — full IDs pass through — and prefs-layer normalization keeping new saves on known aliases). Repo corroboration for the picker surface: `model_policy.go` :145-149 (base aliases remain valid stored values). The AC-ALP-001 harness observes the launcher's argv without a network call — the launcher side is fully provable in-suite; CC's native side is the accepted residual.
- **E.3 opusplan routing alias** — no full-id expansion exists on either side (table self-maps it, :85); passthrough is behavior-identical for it. No research gap.
- **E.4 t1322 merge state** — VERIFIED on this tree: `"opus": ModelIDOpus55` at `model_policy.go:81` and `ModelIDOpus55 = "claude-opus-5-5"` at :42. D-6 recorded; no coordination clause needed in acceptance.

## §F Out of Scope

### F.1 Out of Scope — table contents and bump cadence
- Any edit to `ModelAliasTable` rows, `ModelDeprecatedCanonicalIDs`, `ModelIDOpus55`/`ModelIDSonnet55` constants, or `ModelAliasPickerValues` — t1322 owns contents; this SPEC owns the mechanism only (D-6).
- Any alias-refresh, config-manifest, or versioned-table mechanism (D-1 rejection).

### F.2 Out of Scope — non-Claude backends
- The GLM branch's reverse-mapping logic and the z.ai slot env wiring (`setGLMEnv`, `buildEnvForGLMLaunch`, `glm_effort_overlay.go`) — byte-identical, verification only (D-4).
- The codex launcher, the withdrawn GPT gateway path, and `DO_CLAUDE_*` settings semantics beyond what REQ-ALP-010 pins.

### F.3 Out of Scope — display, validation, and migration surfaces
- `normalizeModel` / `promoteTo1M` / `normalizeModelLegacy1M` behavior (REQ-ALP-005 pins them unchanged).
- `modelOptionList`, `schemaOptionBridge`, wizard picker UX, and web console validation semantics.
- Statusline rendering and served-model gate/stop verdict logic (model-agnostic already).

### F.4 Out of Scope — launch behavior beyond the model flag
- Effort resolution (`resolveLaunchEffort`), profile leases, permission-mode sync, `--continue` fallback, block-cap injection — untouched.
- Warning/error UX for genuinely invalid model values (Claude Code's job; no new MoAI messaging, REQ-ALP-008).
