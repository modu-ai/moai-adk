---
id: SPEC-HANDOFF-RESUME-ASK-001
title: "Handoff resume ask — launcher question, in-session ask mode, and /clear auto-capture across Claude Code and Codex"
version: "0.1.0"
status: draft
created: 2026-10-21
updated: 2026-10-21
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: "internal/hook, internal/hook/handoff, internal/cli, internal/config, internal/settings, internal/homestate"
lifecycle: spec-anchored
tier: M
tags: "handoff, session-start, session-end, launcher, resume, codex, auto-capture, card-t1672"
card: t1672
related_specs: [SPEC-HANDOFF-AUTORESUME-001, SPEC-HANDOFF-NEUTRAL-001, SPEC-INFINITE-GOAL-001]
---

# SPEC-HANDOFF-RESUME-ASK-001 — Handoff resume ask

## A. History

| Date | Version | Change |
|---|---|---|
| 2026-10-10 | — | Operator approved the three-layer design (launcher question / in-session ask / clear capture) in `.moai/reports/moai-handoff-resume-plan-20261010.md`; board record `d-20261010T113836Z-a35a`. |
| 2026-10-21 | 0.1.0 | Initial plan-phase draft (card t1672, Tier M, Class C, P1). Tree `8249bff8c3933d647b4f99e59048595f9808e83a`, branch `WT-clear-cc-codex`; every code fact in §C was re-read at that tree. |

**Tier rationale (one line):** Tier M — the run-phase Go scope is ~11 files with well under 1000 LOC of new behavior across one subsystem (the session-handoff chain), and the design is already fixed by the operator-approved plan document, so no `design.md`/`research.md` uncertainty remains; the rules/template/output-style edits are one documentation-drift family corrected in sync, not a second subsystem. Artifact set: `spec.md` + `plan.md` + `acceptance.md` (+ `progress.md`, uncounted).

## B. Problem Statement

The operator request (card t1672, quoted from the dispatch): when a handoff is requested or `/clear` is run, a handoff should exist; when a new session starts, the user should be asked whether to load the previous handoff or start fresh, on both Claude Code and Codex, all toggleable by configuration.

Today (§C) the only consumer path is `moai handoff show` (manual paste) or the `auto` mode's single inject-and-consume cell on `/clear`. A session started fresh has no resume path at all — the injector deliberately excludes `startup` — and nothing creates a handoff at `/clear` time: the session-end hook fires on every exit reason and never reads `reason`. The result: a `/clear` without a prior `moai handoff save` loses the session's context, and the user must remember to request a handoff before clearing.

This SPEC adds three layers, both harnesses, all toggleable: an **L1 launcher question** (the guaranteed path — `moai cc` / `moai codex` / `moai glm` ask in the terminal when a live handoff exists), an **L2 in-session ask** (a new `handoff.mode: ask` value where the SessionStart hook injects only a short notice on `startup` and `clear`, plus a new `moai handoff resume` verb), and an **L3 `/clear` capture** (the session-end hook machine-captures a light handoff on `reason=clear` when none exists, within a per-harness budget). Model-authored handoffs keep their existing doctrine unchanged.

## C. Current structure (code facts re-read at tree `8249bff8c3933d647b4f99e59048595f9808e83a`; line numbers are coordinates at that tree and drift, the names are the contract)

1. **Store.** The handoff store is the `resume_handoffs` table in the project's factory database (`internal/homestate/factory.go:93`): `status` enum `pending|claimed|consumed|failed|expired|cleared`, body, directives, conversation language, claim columns, and a unique partial index `one_pending_resume_handoff` on `status='pending'` (`factory.go:117`) — at most one live pending row can exist. There is **no `kind`/provenance column**; the auto-capture marker this SPEC needs is a schema addition (the `ALTER TABLE` migration block at `factory.go:396` is the precedent). A sibling `handoff_events` audit table exists (`factory.go:135`: `flow`, `handoff_id`, `from_status`, `to_status`, `detail`). The stale TTL is `DefaultHandoffStaleTTL = 7 * 24h` (`internal/config/defaults.go:829`). The rules documents' `pending.json` references are the drift this card corrects: `.claude/rules/moai/workflow/session-handoff.md:25` is the one live hit (measured; `-format.md` and `-examples.md` carry none).
2. **SessionStart injection.** `handoffInjectHandler` (`internal/hook/handoff_inject.go`) implements one branch table whose single INJECT+CONSUME cell is `source == "clear" ∧ mode == "auto" ∧ live pending`. `mode != "auto"` is a pure no-op (manual never touches pending, even stale). An unknown mode value is coerced to `manual` with a warning (`handoffConfig`, `handoff_inject.go:196`); validation does not reject it. Non-clear sources (`startup`, `resume`, `compact`) get a notice-only stderr hint when `guide: true` — no additionalContext, no consumption; `startup` is deliberately excluded from any injection path. The injected text is rendered in four locales (ko/en/ja/zh, en fallback) by `handoff_inject_render.go`; the body is appended verbatim.
3. **SessionEnd.** `HookInput.Reason` exists (`internal/hook/hook.go:233`: `clear | logout | prompt_input_exit | bypass_permissions_disabled | other`) but `sessionEndHandler` (`internal/hook/session_end.go`) never reads it — no capture device exists.
4. **CLI.** `moai handoff` registers exactly `save`, `clear`, `show` (`internal/cli/handoff.go:48`). `show` is read-only (pending row first, latest consumed fallback) and never claims. There is no `resume` verb.
5. **Config.** `HandoffConfig` carries exactly two keys, `Mode` and `Guide` (`internal/config/types.go:1370`); `DefaultHandoffMode = "manual"` (`defaults.go:704`). Both `handoff.yaml` copies — the template mirror `internal/template/templates/.moai/config/sections/handoff.yaml` and this project's `.moai/config/sections/handoff.yaml` — carry only `mode: manual` and `guide: false`. The web-settings edit enum is the closed set `handoffModeValues = {"manual", "auto"}` (`internal/settings/schema_sections.go:555`). A third copy of the file exists as a test fixture (`internal/settings/testdata/sections/handoff.yaml`).
6. **Launchers.** `moai cc` (`internal/cli/cc.go`), `moai glm` (`internal/cli/glm.go`), and `moai codex` (`internal/cli/codex_launcher.go`) start sessions with no handoff prompt. The Codex launcher already has a first-prompt channel: the CLI's positional PROMPT argument (`codex_launcher.go:115`, exercised by `codexLaneAutoPrompt` at `:969`). The TTY seam `isInteractiveStdin` (`internal/cli/init_update_notice.go:69`) and `writerIsTerminal` (`internal/cli/glamour_style.go:137`) are assignable variables; `runSelectionInteractive` (`factory_run_select.go:244`) is the precedent for a launcher-gate seam.
7. **Output style.** `.claude/output-styles/moai/moai.md` describes the auto-resume flow (the section citing `handoff.mode: auto` and the emission-time save obligation) and its template mirror exists (`internal/template/templates/.claude/output-styles/moai/moai.md`); neither carries a `pending.json` reference (measured 0 hits — see §G on the card-text divergence).

## D. Requirements (GEARS)

Terms. **Live pending**: a `resume_handoffs` row with `status='pending'` that is not past the stale TTL. **Expired**: a `pending` row past the stale TTL. **Interactive**: the launcher's stdin is a terminal (the `isInteractiveStdin` seam). **Ask-notice**: a short injected notice that names the waiting handoff and directs the session to ask the user whether to resume on the session's first turn. **Auto-capture record**: a `resume_handoffs` row created by the session-end hook, marked with a provenance kind distinct from model-authored saves. **Capture budget**: Claude Code 10 seconds, Codex 3 seconds — the wall-time ceiling for the capture path on that harness.

- **REQ-HRA-001** (Event-driven): When a MoAI launcher (`moai cc`, `moai codex`, `moai glm`) starts while `handoff.launcher_prompt` is true, stdin is a terminal, and a live pending resume handoff exists for the project, the launcher shall print a terminal question stating that a handoff is waiting and offering exactly three choices — **resume**, **fresh start**, and **view** — rendered in the project's configured `conversation_language` (ko/en/ja/zh; any other or unreadable value falls back to en), with no preselected choice and deterministic handling of an empty answer or end of input (fresh start, record preserved).
- **REQ-HRA-002** (Event-driven): When the operator chooses resume, the launcher shall claim the pending record through the existing transactional claim (a concurrent claimer wins atomically) and shall start the session with the claimed body verbatim as the session's first prompt (the Codex launcher through its existing positional PROMPT channel; the Claude Code and GLM launchers through the positional initial-prompt argument of the spawned CLI). When the claim is lost to a concurrent claimer, the launcher shall start the session without the body and without an error.
- **REQ-HRA-003** (Event-driven): When the operator chooses fresh start, the launcher shall leave the pending record in `pending` status untouched, shall append one decline entry to the `handoff_events` audit trail (detail naming the launcher and the operator's decline), and shall start the session with no injected body. When the operator chooses view, the launcher shall print the stored body verbatim to the terminal, shall claim and consume nothing, and shall then re-present the question.
- **REQ-HRA-004** (State-driven): While the launcher's stdin is not a terminal, the launcher shall ask nothing, shall start the session unchanged, and shall leave the pending record untouched (the `launcher_prompt` toggle is unobservable on a non-interactive start).
- **REQ-HRA-005** (Event-driven): When `handoff.mode` is `ask` and a SessionStart event arrives with a live pending resume handoff, the SessionStart handoff-inject handler shall inject an ask-notice as additionalContext on the `startup` and `clear` sources only; the notice shall never contain the saved body; the handler shall not claim, consume, or modify the record on any source under `ask`; and the notice shall be rendered in the session's conversation_language (ko/en/ja/zh, en fallback). On the `resume` and `compact` sources under `ask` the handler shall inject nothing (silent no-op; record preserved).
- **REQ-HRA-006** (Event-driven): When `handoff.mode` is `ask` and the pending record is expired, the handler shall apply the same stale-TTL expiry the `auto` mode applies today — the record is expired and no notice is emitted.
- **REQ-HRA-007** (Event-driven): When `moai handoff resume` runs while a live pending record exists, the verb shall claim it (transactional compare-and-swap), print the claimed body verbatim to stdout exactly once, and leave the record consumed; when no live pending record exists or the record is expired, the verb shall exit non-zero, print a one-line error localized to ko/en/ja/zh (en fallback), and print no body.
- **REQ-HRA-008** (Ubiquitous): The SessionStart behaviors of the `manual` and `auto` modes shall remain exactly as they are at the base tree — `manual` a pure no-op on every source (never touching pending, even stale), `auto` keeping its single inject+consume cell (`source == clear` ∧ live pending), its notice-only stderr hint on non-clear sources when `guide: true`, and its stale-TTL expiry — and the `ask` addition shall alter neither.
- **REQ-HRA-009** (Ubiquitous): The in-session ask path shall never block a session start and shall never invoke AskUserQuestion (subagent boundary); on Codex, where no question tool exists, the ask-notice directs the session to stop with a two-choice report (resume / fresh start), and that path stays best-effort — the launcher question is the guaranteed path for Codex.
- **REQ-HRA-010** (Event-driven): When a SessionEnd event arrives with `reason == "clear"`, `handoff.capture_on_clear` is true, and no live pending record exists, the session-end handler shall create exactly one auto-capture record — marked with the auto-capture provenance kind, carrying the ending session's id as `saved_by_session`, and a light body composed of the session id, the current branch, the HEAD SHA, the active SPEC id, and the session's last user prompt (a fact that cannot be read is recorded as an empty field) — within the capture budget of the running harness, performing no network call and no model call; the capture path is fail-open: any error leaves no record and never blocks session end.
- **REQ-HRA-011** (Event-driven): When a SessionEnd event arrives with `reason == "clear"` and a live pending record already exists — saved by this session or any other — the handler shall leave it untouched and create no record; when the reason is anything other than `clear` (`logout`, `prompt_input_exit`, `bypass_permissions_disabled`, `other`, or absent), the handler shall create no record regardless of the toggle.
- **REQ-HRA-012** (Ubiquitous): The handoff configuration section shall carry exactly four keys — `mode` (`manual | auto | ask`), `guide` (bool), `capture_on_clear` (bool), `launcher_prompt` (bool). Template defaults: `mode: manual`, `guide: false`, `capture_on_clear: false`, `launcher_prompt: true`; a missing or unreadable config falls back to that default set (existing behavior unchanged). This project's own `handoff.yaml` shall set `mode: ask` and `capture_on_clear: true`.
- **REQ-HRA-013** (Event-detected): When configuration loading encounters a `handoff.mode` value outside `{manual, auto, ask}`, the loader shall reject the configuration with a validation error naming the key and the offending value; the web-settings schema enum shall list all three values; the SessionStart handler's defensive coercion of an unknown mode to `manual` remains in place as fail-open defense-in-depth for stores that predate validation.
- **REQ-HRA-014** (Ubiquitous): The session-handoff rules documents and their template mirrors shall reference the `resume_handoffs` table in the project's factory database as the handoff store (never `.moai/state/handoff/pending.json`), shall document the four config keys and the manual/auto/ask semantics, and shall remain generic deployables with no project-specific values in any template-mirror file; the output-style section describing the auto-resume flow shall describe the ask mode's notice behavior alongside it.

## E. Decision Points

All four are design defaults applied under the operator-approved plan (board `d-20261010T113836Z-a35a`, 2026-10-10); none is open, and none requires an operator question at plan-audit.

| DP | Decision | Disposition |
|---|---|---|
| Q1 | Which sources the ask-notice covers. | **Default applied: `startup` + `clear` only** (the card names exactly these two); `resume`/`compact` are silent no-ops under `ask` (REQ-HRA-005) — a resumed session already passed the launcher question. |
| Q2 | Expired-pending behavior under `ask`. | **Default applied: same stale-TTL expiry as `auto`** (REQ-HRA-006) — one expiry semantic, not two. |
| Q3 | Launcher answer that is empty, unreadable, or end-of-input. | **Default applied: fresh start, record preserved** (REQ-HRA-001) — the launcher must proceed, and proceeding must not consume. |
| Q4 | Locale source for the launcher question, which runs before any session exists. | **Default applied: the project's configured `conversation_language`, read at launcher time; unreadable → en** (REQ-HRA-001). |

## F. Out of Scope

### Out of Scope — Codex `/new` SessionStart source value

- What source value (if any) Codex's in-session `/new` delivers to a SessionStart hook is unmeasured. It is recorded as an open question (§G), excluded from every requirement here, and not a completion criterion; the Codex in-session path is best-effort and the launcher question is its guaranteed path.

### Out of Scope — Machine-capture summary quality

- No improvement to the quality of machine-captured bodies. The capture performs no model call; when quality matters, the existing request-a-handoff path (`moai handoff save`) remains the quality path, unchanged with its five triggers and save obligation.

### Out of Scope — Model-authored handoff doctrine

- The model-authored handoff rules — the five emission triggers, the emission-time save obligation, the paste-ready 6-block format — are unchanged by this SPEC. The capture adds a machine path beside them, never a replacement.

### Out of Scope — manual / auto behavior changes

- Any change to `manual` (pure no-op) or `auto` (single inject+consume cell, guide hints, stale expiry) behavior (REQ-HRA-008 pins the baseline; the matrix test re-asserts every manual/auto cell).

### Out of Scope — Memory handoffs and goal re-arm

- The SessionEnd memory-handoff rows (`memory_handoffs` table) and their separate state machine; the auto-capture never consumes them. The capture embeds no goal and no directives; embedded-goal re-arm semantics stay owned by SPEC-INFINITE-GOAL-001.

### Out of Scope — Launcher flows this SPEC does not own

- Factory lane-join selection and its chooser (SPEC-LANE-JOIN-RUN-SELECT-001); any change to the paste-ready body format; handoff retention/TTL policy changes; the `memory handoff` and `factory handoff` flows that share factory.db but not the `resume_handoffs` table.

## G. Verification Gaps and Open Questions

- **Codex SessionStart payload `source` value** for a plain startup is unmeasured at this tree. Run phase measures it once and records the value in the report (AC-HRA-014). The 36-cell matrix (AC-HRA-001) asserts the Claude Code source axis (`startup | clear | resume | compact`); if the Codex measurement shows a different source vocabulary, the report records the mapping and the ask-notice applies to the measured startup-equivalent source. Not blocking: the Codex in-session path is best-effort by design (REQ-HRA-009).
- **Codex `/new` source value**: unmeasured, out of scope (§F), carried as an open question for a future measurement — no requirement depends on it.
- **Last-user-prompt fact source for the capture**: the hook-side read mechanism (e.g. a transcript path from the SessionEnd payload) is resolved at run phase; REQ-HRA-010 already makes the field fail-open (empty when unreadable), so no criterion depends on the mechanism chosen.
- **Card-text divergence measured** (recorded, also in `progress.md` §E.1): the card expects an output-style "banner 7" pending.json drift to correct; at this tree `.claude/output-styles/` carries **zero** `pending.json` references — the live drift is the single hit at `.claude/rules/moai/workflow/session-handoff.md:25` (plus its template mirror). The output-style obligation here is the ask-mode amendment of the auto-flow section (REQ-HRA-014), not a pending.json fix.
- **Launcher TTY behavior and the Codex launcher question** are measured manually once each (AC-HRA-008, AC-HRA-009, AC-HRA-014); no automated test drives a real terminal, matching the SPEC-LANE-JOIN-RUN-SELECT-001 precedent.
