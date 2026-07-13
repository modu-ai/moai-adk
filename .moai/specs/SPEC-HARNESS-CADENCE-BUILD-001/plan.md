---
id: SPEC-HARNESS-CADENCE-BUILD-001
version: "0.1.2"
status: completed
updated: 2026-07-13
---

# Implementation Plan — SPEC-HARNESS-CADENCE-BUILD-001

## §A Context

Integrate cadence (recurring schedules) into the v4 harness Builder under the user-confirmed **discovery-queue model**: (1) build-time recurrence question at the PLAN→GENERATE gate, (2) optional `schedule` field in manifest.json + lifecycle-verb awareness, (3) a new sanctioned cadence-bridge recipe class whose payload is a self-contained discovery prompt, (4) an ANALYZE external-research sub-step, (5) a Schedule Retrofit branch for already-built harnesses. The cadence-bridge catalog HARD invariant (scheduled runs never commit/push/enter run-phase) is preserved verbatim.

### §A.1 Measured baseline (verified 2026-07-13, this checkout)

| Anchor | Observation |
|---|---|
| `v4manifest.Manifest` | 8 top-level fields, `internal/harness/v4manifest/types.go`; decoder tolerates unknown fields |
| `v4manifest.Validate` | `internal/harness/v4manifest/validate.go` — single validation entry point |
| Lifecycle verbs | `ListHarnesses`/`EditHarness`/`RemoveHarness` in `internal/cli/harness/v4lifecycle.go`; cobra wrappers in `v4lifecycle_cmd.go`; `HarnessEntry{Name,Domain,EntryCommand}` |
| Doctor | `internal/cli/harness/doctor.go` — axis 2 = `v4manifest.Validate` reuse → Validate extension propagates to doctor for free; severities ERROR/INFO; command-only harnesses = INFO |
| cadence-bridge.md | Recipes 1-3; invariant stated once at catalog level (`grep -c "scheduled runs never commit, never push, never enter run-phase"` = 1); Discovery-to-Queue Contract; Cron-unavailable fallback clause |
| Template parity | All 4 target md files byte-identical to `internal/template/templates/` mirrors (diff exit 0 ×4) |
| Existing manifests | `release-update` has a full manifest (no schedule); `github`/`release` are command-only (no manifest — deliberate Runner asymmetry) |
| SPEC ID dedup | No existing `*CADENCE-BUILD*` SPEC; `SPEC-CADENCE-BRIDGE-001` owns the recipe catalog's creation and is related, not superseded |

## §B Known Issues / Hazards

1. **`harness-run` dispatch gap (pre-existing, out of scope)**: `v4manifest.CommandTemplate` routes `Skill("moai") harness-run <name> $ARGUMENTS`, but the moai SKILL.md documents no `harness-run` subcommand (grep = 0 matches). Recipe 4 deliberately does NOT depend on the entry command / Runner dispatch (payload = discovery prompt), so this gap does not block; recorded so nobody wires the scheduled payload through `/harness:<name>` assuming that path is exercised.
2. **`/loop` is session-scoped**: an armed loop dies with the session. The recurrence-question option descriptions and the ACTIVATE emission MUST state this (REQ-HCB-002/030); `moai harness list` surfacing the declared schedule is what lets a later session re-arm.
3. **Cron tool availability varies by runtime version**: degrade to `/loop` per the existing catalog fallback clause (REQ-HCB-025).
4. **RemoveHarness deletes the manifest**: the unregister notice must be computed BEFORE deletion (read manifest → remove → print). Implementation in the cobra wrapper or a pre-read inside `RemoveHarness`'s caller — do not reorder removal atomicity.
5. **Byte-parity + neutrality coupling**: because live md files must stay byte-identical to template mirrors, the added doctrine text may NOT carry this SPEC's ID even in the live tree. Write all four md edits in neutral prose (the existing cadence-bridge.md style proves this is workable).
6. **Doctor severity is binary (ERROR/INFO)**: an invalid schedule is ERROR (invariant/schema violation blocks the gate); an absent schedule produces no finding. Do not introduce a WARN tier.
7. **Shared-checkout races**: plan-phase artifacts are committed immediately (pathspec-scoped) per prior-incident discipline; run-phase spawns follow the pre-spawn sync check.

### Resolved clarifications (2026-07-13 — orchestrator AskUserQuestion round; decisions binding)

1. **Command-only thin-harness schedule persistence — DECIDED: registration-only.** No manifest fabrication for github/release; the Runner asymmetry is preserved and the `v4manifest.Validate` schema is untouched by the thin-harness path. REQ-HCB-052 stands as written; the rejected alternative (relaxing Validate to accept a thin schedule-bearing manifest) is closed.
2. **Recurrence question placement — DECIDED: folded into the existing PLAN→GENERATE AskUserQuestion gate round.** One added question in the same round (the ≤4-question round limit is respected); option descriptions MUST state the `/loop` session-scoped vs Cron persistent trade-off AND the discovery-only execution model. REQ-HCB-001/002/005 stand as written; the rejected alternative (separate post-gate round) is closed.
3. **Schedule field shape — DECIDED: minimal 3-field object.** `interval` / `mechanism: loop|cron` / `mode: "discovery-only"` literal; additive-only, absent-tolerated. NO `enabled` or registration bookkeeping fields — the runtime queue/CronList is the truth source for registration state. REQ-HCB-010..014 and the spec Exclusions stand as written.

## §C Pre-flight (run-phase entry checklist)

- [ ] `git fetch origin main` + divergence check + `moai session list --json` (pre-spawn sync discipline)
- [ ] Re-verify the 4 md byte-parity pairs still hold (`diff` ×4 — parallel sessions may have moved them)
- [ ] Re-verify `v4manifest.Validate` signature and doctor axis-2 reuse unchanged (content-token grep, not line numbers)
- [ ] Confirm no new `*CADENCE*` SPEC/branch landed since plan authoring
- [ ] Confirm the three resolved-clarification decisions (§B Resolved clarifications) still stand — no user reversal since the plan revision

## §D Constraints

- Frozen: Implementation Kickoff Approval gate, AskUserQuestion monopoly, cadence-bridge catalog invariant (textual preservation pinned by AC-HCB-022).
- Additive-only Go schema; zero behavior change for schedule-less manifests (AC-HCB-011/035).
- CLI never prompts (existing boundary tests extend to new code).
- Template-First: edit live + mirror in the same milestone; `make build` after template edits.
- Neutrality: no SPEC/REQ tokens in any of the 4 md files (live or mirror).
- No time estimates; priority labels only.

## §E Self-Verification (run-phase, per milestone)

Per-milestone evidence lands in `progress.md` §E.2 with verbatim command outputs (file-redirect contract for long outputs):

- E1: AC matrix PASS/FAIL table (acceptance.md §D as the checklist)
- E2: `go build ./...` + `make build` green
- E3: `go test ./internal/harness/v4manifest/... ./internal/cli/harness/... -count=1` then full `go test ./...`
- E4: boundary grep — no NEW user-prompt tool tokens in `internal/cli/harness/` beyond the measured baseline of 5 pre-existing documentation/help-text string matches (delta-framed per acceptance.md AC-HCB-034)
- E5: `golangci-lint run` on touched packages
- E6: byte-parity `diff` ×4 + neutrality grep on `internal/template/templates/`
- E7: `moai harness doctor` exit 0 on the real repo; `moai spec lint` findings for this SPEC = 0

## §F Milestones (ordered by decision reversibility — highest-change-likelihood first)

### M1 — Manifest schedule schema (Priority: High) — data-model decision

Files: `internal/harness/v4manifest/types.go`, `internal/harness/v4manifest/validate.go`, tests (`types_test.go`/`validate_test.go` or existing test files).

- Add `Schedule *Schedule \`json:"schedule,omitempty"\`` to `Manifest`; new `Schedule{Interval, Mechanism, Mode string}` struct with godoc naming the discovery-only invariant.
- Extend `Validate`: nil schedule → no-op; present → interval non-empty, mechanism ∈ {loop, cron}, mode == "discovery-only" (exact literal).
- Table-driven tests: valid loop/cron fixtures, absent-schedule regression, mode violation, mechanism violation, empty interval, unknown extra keys tolerated.
- REQs: HCB-010..014. ACs: HCB-010..013, HCB-070 (partial).

### M2 — Doctrine surfaces: recurrence question, recipe, research, retrofit (Priority: High) — user-facing flow decisions

Files (live tree): `.claude/rules/moai/workflow/cadence-bridge.md`, `.claude/skills/moai/workflows/harness-builder.md`, `.claude/skills/moai/workflows/harness-build-entry.md`.

- cadence-bridge.md: Recipe 4 "Scheduled Harness Discovery" (recipe class; payload = self-contained discovery prompt template; interval guidance; read-only rationale; loop AND cron forms; Cron-unavailable degradation pointer) + eligibility-table row (data-row count 6 → 7). Invariant sentence untouched, not restated (AC-HCB-022 pins count == 1). The self-contained payload inlines the `.moai/reports/cadence/<date>.md` queue-path literal — file-level literal count rises 1 → 2 by design (AC-HCB-023 delta-frame; the +1 lives inside the Recipe 4 payload block).
- harness-builder.md: (a) ANALYZE research sub-step (official CC docs via WebFetch/WebSearch, domain best practices, context7 resolve-library-id → query-docs; feeds PLAN aggregate; MCP-fallback + GLM-routing cross-refs; load-bearing-minimum skip clause); (b) PLAN gate gains the recurrence question + interval/mechanism capture + draft-manifest `schedule` recording; (c) Artifact 5 content contract lists optional `schedule` (3 sub-fields, discovery-only literal); (d) ACTIVATE gains the post-smoke-gate registration step (CronCreate prompt form / paste-ready `/loop` emission, session-scoped caveat).
- harness-build-entry.md: Schedule Retrofit branch — detection (existing `.claude/commands/harness/<name>.md` + scheduling intent) evaluated BEFORE the existing Phase 2 name-collision matrix (existing-name + scheduling-intent routes to Retrofit, never to the `<name>-v2` re-derive/rename path — REQ-HCB-050 precedence pin); recurrence round, manifest-bearing path (orchestrator-mediated manifest edit via `moai harness edit` path discovery, then registration), command-only path (registration-only, no manifest fabrication, user informed), dev-only isolation note.
- Anchor obligation (AC grep bounding): the new sections MUST carry the pinned heading anchors — a heading containing `Recurrence` (harness-builder.md PLAN gate), `Recipe 4` (cadence-bridge.md), `Schedule Retrofit` (harness-build-entry.md). The acceptance.md windowed greps awk-bound on these anchors (AC-HCB-001/002/021/024/051). Flat-window constraint: the awk idiom closes each window at the FIRST subsequent heading of ANY level — keep all windowed tokens above any sub-heading inside each new anchored section, or repeat the anchor token in the sub-heading. Ordering-token discipline: in the Schedule Retrofit window name `CronCreate` only in the registration step (after `moai harness edit`, never in the recurrence-round prose), and in the ACTIVATE section place the registration step's first `CronCreate` mention after the smoke-gate paragraph (AC-HCB-030/051 first-match ordering).
- All text neutral (no SPEC IDs). REQs: HCB-001..005, 020..025, 030 (doctrine half), 040..044, 050..053.

### M3 — CLI lifecycle awareness (Priority: Medium) — mechanical, follows M1

Files: `internal/cli/harness/v4lifecycle.go`, `internal/cli/harness/v4lifecycle_cmd.go`, tests; `internal/cli/harness/doctor_test.go` (fixture only — doctor.go itself expected unchanged via Validate reuse).

- `HarnessEntry` gains `Schedule *v4manifest.Schedule \`json:"schedule,omitempty"\``; `ListHarnesses` populates from manifest; text output appends `schedule: <interval> via <mechanism>` only when present.
- Remove wrapper: pre-read manifest schedule → after successful removal, print unregister notice (mechanism-specific). No notice when absent.
- Doctor test: fixture manifest with `mode: "write"` → ERROR finding via axis-2; absent-schedule fixtures unchanged.
- REQs: HCB-031..035, 030 (surfacing half). ACs: HCB-031a/b, 032, 033, 035.

### M4 — SKILL.md alignment + template mirrors + build (Priority: Medium) — mechanical propagation

Files: `.claude/skills/moai/SKILL.md` (Branch A.1 verb descriptions: list shows schedule; remove emits unregister notice) + `internal/template/templates/` mirrors of all 4 md files; `make build`.

- Copy each edited live file to its mirror (byte-identical); run `make build`; template CI guards green.
- REQs: HCB-060..062. ACs: HCB-060, 061, 062.

### M5 — Full verification sweep (Priority: Low) — closure

- Full `go test ./... -count=1`, `golangci-lint run`, `moai harness doctor` (real repo, exit 0), `moai spec lint` (this SPEC = 0 findings), byte-parity diff ×4, neutrality grep, §E evidence recorded in progress.md.
- ACs: HCB-070..073 + E1-E7 matrix.

## §G Anti-Patterns (do NOT)

- Do not restate the catalog invariant inside Recipe 4 (breaks the single-statement design; AC-HCB-022 fails on count > 1).
- Do not wire the scheduled payload through `/harness:<name>` / the Runner (write-capable specialists; also rides the unresolved harness-run gap — §B.1).
- Do not add a WARN doctor severity, a schedule-state registry, or an `enabled` bookkeeping field (scope creep; see Exclusions).
- Do not fabricate a manifest for command-only harnesses to carry a schedule.
- Do not put this SPEC's ID into any of the 4 mirrored md files (neutrality + byte-parity coupling).
- Do not verify ACs by token-presence alone where a behavior exists — Go-verifiable ACs assert through `go test` fixtures (Validate/List/Remove/Doctor behavior), not grep.
- Do not run `git add -A`; commit pathspec-scoped.

## §H Cross-References

- spec.md §B (REQ SSOT), acceptance.md §D (AC matrix SSOT)
- `.claude/rules/moai/workflow/cadence-bridge.md` — catalog invariant + queue contract (consumed/extended)
- `.claude/rules/moai/core/agent-common-protocol.md` § MCP Fallback Strategy / § Pre-Spawn Sync Check
- `.claude/rules/moai/core/glm-web-tooling.md` — GLM research routing
- CLAUDE.local.md §2 (Template-First), §21/§24 (dev-only + namespace isolation), §25 (template neutrality)
