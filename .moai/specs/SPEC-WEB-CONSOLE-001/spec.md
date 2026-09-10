---
id: SPEC-WEB-CONSOLE-001
title: "MoAI Web Console — Browser-based Settings CRUD"
version: "0.3.0"
status: completed
created: 2026-06-03
updated: 2026-09-10
author: manager-spec
priority: P1
phase: "v3.0.0"
module: "internal/web"
lifecycle: spec-anchored
tags: "web, console, settings, crud, cli, localhost"
tier: M
related_specs: [SPEC-V3R3-WEB-001]
---

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-06-03 | manager-spec | Created — fulfils the CRUD "Console" requirement that motivated archiving SPEC-V3R3-WEB-001. Browser-based READ/WRITE of profile preferences + user/language/statusline settings via a `moai web` subcommand (Go single binary, net/http + go:embed). **SUPERSEDES SPEC-V3R3-WEB-001** (read-only Cockpit/Workflow-Tracker capability). The read-only invariant of the predecessor is deliberately discarded; the polling Workflow-Tracker panels are out of scope. |
| 0.3.0 | 2026-09-10 | manager-spec | In-place amendment (card t613): the REQ-WC-009 Host check is widened from mutating methods to every method on every route, static assets included. The original requirement text is retained verbatim. See § Amendments and the supersession block under REQ-WC-009. |

### Amendments

- **2026-09-10 (in-place amendment, v0.2.0 → v0.3.0)** — **The REQ-WC-009 Host-check method scope is widened from `POST`/`PUT`/`PATCH` to every request on every route, `/static/` assets included.** A foreign or absent `Host` is now rejected with HTTP 403 on reads as well as writes. `status: completed` is deliberately unchanged: this amends the contract of a closed SPEC. It does not reopen the SPEC, and no run-phase work is re-entered under this SPEC ID.
  - **Cause**: card t613 reproduced foreign-`Host` and absent-`Host` GET requests receiving HTTP 200 with settings content, while POST under the same Hosts was rejected. The exemption traces to an analysis gap in this SPEC's own plan (plan.md R3), not to a recorded risk acceptance. The full rationale sits in the supersession block under REQ-WC-009, and the evidence is `.moai/reports/t613/verdict.md`.
  - **Decision**: operator, 2026-09-10. Option A ("all-route Host gate") of verdict §5 was chosen over B (gate dynamic routes only, leave `/static/` open) and C (keep the exemption and record the risk acceptance).
  - **Explicitly preserved, unchanged**: the allowed Host set (`localhost`, `127.0.0.1`, `[::1]`, each with or without a port). The `Sec-Fetch-Site` same-origin CSRF gate of SPEC-INTERNAL-SECURITY-001 REQ-SEC-002, which stays scoped to state-changing routes. The GLM key reveal defense (`/glm-key/reveal` is POST-only, and its handler re-checks loopback). The "no token auth, no session store, no CSRF-token infrastructure" boundary of REQ-WC-009 and § Exclusions.
  - **Scope**: `spec.md` REQ-WC-009 (supersession block appended after the unchanged original text); `acceptance.md` AC-WC-009 (matrix row annotated, original scenario kept and marked, amended scenario added); `plan.md` Phase 6 and R3 (dated notes, original text intact). Prior completed version: `v0.2.0`. No other requirement, AC, or artifact is touched. Code, test migration, and user documentation changes are outside this amendment.
  - **Schema note**: the canonical `completed → in-progress (amendment)` transition (`.claude/rules/moai/development/spec-frontmatter-schema.md` § Status Transition Ownership Matrix) pairs `amendment_of:` with a status change to `in-progress`. That status change is not performed here, per the delegating instruction. `amendment_of:` is therefore omitted as well, and this sub-section is the amendment record. SPEC-MOAI-MCP-SERVER-001 and SPEC-TREND-MCP-001 used the same form.

> **Supersede note**: SPEC-V3R3-WEB-001 ("Cockpit Foundation + Workflow Tracker", status `archived`) was archived precisely because the user needed to EDIT settings from a browser (CRUD), which conflicted with that SPEC's read-only ambient-awareness invariant. The "Console" (CRUD) pivot proposed at that time was never written as a SPEC. This document IS that successor. Reusable infrastructure patterns (CLI subcommand + flags, loopback-only HTTP server, graceful shutdown, browser auto-open, `go:embed` static assets) are adapted from the predecessor's `plan.md`; the read-only invariant and the polling/tracker requirements are NOT carried over.

---

# SPEC-WEB-CONSOLE-001: MoAI Web Console — Browser-based Settings CRUD

## Background

MoAI-ADK exposes a terminal TUI wizard (`internal/cli/profile_setup.go`, huh/bubbletea) that lets a user configure profile preferences and the core project settings. The archived SPEC-V3R3-WEB-001 introduced a browser surface but constrained it to read-only ambient awareness, which blocked the user's actual need: editing settings from a browser. This SPEC delivers a minimal local web UI — the **MoAI Web Console** — that READs and WRITEs the named MoAI settings from a browser, reusing the existing validation and persistence logic so the browser surface is a thin equivalent of the terminal wizard, not a parallel implementation.

The Console is exposed as a new `moai web` cobra subcommand on the existing binary. It binds to loopback only (`127.0.0.1`), serves a single lightweight page with embedded static assets, and persists changes through the existing `internal/profile/` read/write/sync functions. There is no external database, no JS build toolchain, no live tracker, and no multi-user/auth model.

### Ground-truth references (verified)

| Path | Role in this SPEC |
|------|-------------------|
| `internal/profile/preferences.go` | `ProfilePreferences` struct = the CRUD form schema source. Canonical validation list `ValidPermissionModes` + `IsValidPermissionMode()`. Profile store via `ReadPreferences(name)` / `WritePreferences(name, prefs)` at `~/.moai/claude-profiles/<name>/preferences.yaml`. |
| `internal/profile/sync.go` | `SyncToProjectConfig(projectRoot, prefs)` = canonical persistence entry point that writes `user.yaml` + `language.yaml` + `statusline.yaml` (via `syncStatusline`). The web layer persists through this, never by writing YAML directly. |
| `internal/profile/profile.go` | Existing profile read/write helpers (profile listing / current / base dir). |
| `internal/cli/profile_setup.go` | Canonical value lists to reuse for server-side validation: `statuslineModeCanonical`, `statuslinePresetCanonical`, `statuslineThemeCanonical`, language options `en`/`ko`/`ja`/`zh`. |
| `internal/cli/profile_setup_translations.go` | Translation keys/labels reusable for the browser form. |
| `internal/cli/profile.go` | Existing terminal profile command (the Console is the browser equivalent of the same scope). |
| `internal/cli/root.go` | Cobra root command — register `web` subcommand here via `rootCmd.AddCommand(...)` (mirror `internal/cli/cc.go`, `internal/cli/brain.go`). |
| `.moai/config/sections/{user,language,statusline}.yaml` | Project-level CRUD targets (written via `SyncToProjectConfig`). |
| `.moai/specs/SPEC-V3R3-WEB-001/plan.md` | REFERENCE for reusable infra patterns ONLY (CLI flags `--port`/`--no-open`, loopback bind, graceful 5s drain, browser auto-open, `go:embed`). |

### `ProfilePreferences` fields (CRUD form schema)

`UserName`, `ConversationLang`, `GitCommitLang`, `CodeCommentLang`, `DocLang`, `ModelPolicy`, `Model`, `EffortLevel`, `PermissionMode` (+ deprecated `Bypass`, auto-migrated to `PermissionMode`), `StatuslineMode`, `StatuslinePreset`, `StatuslineSegments` (`map[string]bool`), `StatuslineTheme`.

## Goals

| Goal ID | Type | Target |
|---------|------|--------|
| Goal Primary | Capability | A user edits the named MoAI settings (profile preferences + user/language/statusline) from a browser and the change persists to disk through existing `internal/profile/` functions. |
| Goal Secondary 1 | Safety | The Console is reachable only from the local machine (loopback bind) and mutating requests reject cross-origin/foreign-Host attempts. |
| Goal Secondary 2 | Reuse | Validation and persistence reuse the existing canonical value lists and `internal/profile/` write/sync functions — zero parallel validation, zero direct YAML writes. |
| Goal Anti | Minimalism | No auth/session/token infrastructure, no external DB, no SSE/WebSocket, no SPA framework, no live workflow tracker. The smallest change that delivers browser CRUD over the named settings. |

Goal Anti is a HARD constraint on every requirement below — its violation reintroduces the over-engineering this SPEC explicitly avoids.

## GEARS Requirements

### REQ-WC-001 (Ubiquitous — CLI subcommand + flags)

The `moai` binary **shall** expose a `web` subcommand registered on the cobra root command, accepting `--port <int>` (default candidate 8080) and `--no-open` (boolean), with a working `moai web --help` listing both flags.

### REQ-WC-002 (Ubiquitous — loopback-only server lifecycle)

The Console HTTP server **shall** bind exclusively to the loopback interface `127.0.0.1:<port>` and **shall not** bind to `0.0.0.0` or any non-loopback address.

### REQ-WC-003 (Event-driven — graceful shutdown)

**When** the process receives SIGINT or SIGTERM, the Console server **shall** stop accepting new connections, drain in-flight requests for up to 5 seconds, close the listener, and exit with status 0.

### REQ-WC-004 (Event-driven — browser auto-open)

**When** the server starts and `--no-open` is not set, the Console **shall** attempt to open `http://127.0.0.1:<port>` in the default browser; **When** the auto-open attempt fails, the Console **shall** continue serving (auto-open failure is non-fatal).

### REQ-WC-005 (Ubiquitous — embedded static assets)

The Console **shall** serve its page shell, CSS, and client script from assets bundled into the binary via `go:embed` (under the new Go package, e.g. `internal/web/assets/`), requiring no separate JS runtime, build toolchain, or network fetch of frontend dependencies.

### REQ-WC-006 (Event-driven — settings READ)

**When** a user requests the Console page (`GET /`), the Console **shall** render the current values of the in-scope settings — `ProfilePreferences` fields for the selected profile plus the project-level `user.yaml` / `language.yaml` / `statusline.yaml` values — as a pre-populated editable form, reading current state through existing `internal/profile/` read functions (`ReadPreferences`).

### REQ-WC-007 (Event-driven — settings WRITE via existing persistence)

**When** a user submits the settings form (a mutating request, e.g. `POST`), the Console **shall** validate the submitted values using the existing canonical value lists (`ValidPermissionModes` / `IsValidPermissionMode`, `statuslineModeCanonical`, `statuslinePresetCanonical`, `statuslineThemeCanonical`, language `en`/`ko`/`ja`/`zh`) and, on success, persist them by calling the existing profile/sync functions (`WritePreferences` for the profile store and `SyncToProjectConfig` for `user.yaml`/`language.yaml`/`statusline.yaml`) — never by writing YAML directly.

### REQ-WC-008 (Event-driven — invalid input rejection)

**When** a submitted value fails the existing validation (e.g. an unrecognized `PermissionMode` or statusline theme), the Console **shall** reject the mutation, leave persisted state unchanged, and re-render the form with a per-field error message; the rejection **shall** reuse the existing validation predicates rather than a parallel rule set.

### REQ-WC-009 (Event-driven — write-safety Host check)

**When** the Console receives a mutating request (`POST`/`PUT`/`PATCH`), the Console **shall** verify the request `Host` header resolves to a loopback origin (e.g. `127.0.0.1:<port>` or `localhost:<port>`) and **shall** reject the request with HTTP 403 when the Host header does not match, to prevent DNS-rebinding/CSRF from other local origins. This check is the sole write-safety boundary beyond loopback binding — no token auth, session store, or CSRF-token infrastructure is introduced.

> **Supersession — REQ-WC-009 method scope (2026-09-10, v0.3.0).** The requirement text above is kept verbatim and is superseded **in part**. Its method scope, "a mutating request (`POST`/`PUT`/`PATCH`)", is replaced by "every request, of any HTTP method, on every route, static assets under `/static/` included". Several parts are unchanged: the loopback-origin verification, the HTTP 403 response on mismatch, the no-token / no-session / no-CSRF-token boundary, the `Sec-Fetch-Site` same-origin gate of SPEC-INTERNAL-SECURITY-001 REQ-SEC-002 (still scoped to state-changing routes), and the GLM key reveal defense (`/glm-key/reveal` is POST-only, and its handler re-checks loopback). Where the two texts disagree, the amended requirement below governs. See also HISTORY § Amendments.

#### REQ-WC-009 as amended 2026-09-10 (Event-driven — all-route Host check)

**When** the Console receives a request of any HTTP method on any route (dynamic pages, form endpoints, and static assets under `/static/` alike), the Console **shall** verify that the request `Host` header resolves to a loopback origin, accepting exactly `localhost`, `127.0.0.1`, and `[::1]`, each with or without a port. **When** the `Host` header names any other host or is absent, the Console **shall** reject the request with HTTP 403, serving none of the route's content and performing no state change. The purpose is to prevent DNS-rebinding reads as well as DNS-rebinding/CSRF writes from other origins. The `Sec-Fetch-Site` same-origin CSRF gate (SPEC-INTERNAL-SECURITY-001 REQ-SEC-002) **shall** remain scoped to state-changing routes and is not widened by this requirement. As in the original, no token auth, session store, or CSRF-token infrastructure is introduced.

**Rationale.** On local `develop` tree `d3b7d438d`, a foreign `Host` and an absent `Host` both received HTTP 200 on `/settings`, `/`, `/kanban`, `/monitor`, `/todo`, `/specs`, and `/static/app.js`. `/settings` echoed an injected profile marker, while `POST /save` under the same Hosts returned 403. The GET exemption originates in commit `b1ab60454` (this SPEC's M1), where plan.md R3 analyzed DNS rebinding for one impact only, "Unauthorized mutation"; read exposure was never analyzed. "Reads are safe" therefore entered the code comment and AC-WC-009 as a conclusion without analysis, and no risk-acceptance decision was recorded. It was an analysis gap, not an accepted risk. R3's mitigation, "loopback bind is the outer boundary", does not hold for reads under DNS rebinding: the browser connects to 127.0.0.1 while sending `Host: <attacker domain>` and, being same-origin from its own view, can read the response. The Host check is the only server-side defense on that path, so it must cover reads. Gating every method was measured to fail 79 existing `internal/web` tests with 403. The verdict attributes this to the tests' non-loopback default Host, judged from sampled failures rather than confirmed per test, and those tests move to a loopback Host.

**Provenance.** Card t613. Evidence: `.moai/reports/t613/verdict.md` (§2 origin, §3 reproduction matrix, §4 test impact). Operator decision 2026-09-10: verdict §5 option A, "all-route Host gate".

### REQ-WC-010 (Event-driven — graceful empty/error states)

**When** the selected profile has no `preferences.yaml` yet (zero-value preferences) or a project config section is absent, the Console **shall** render the form with neutral defaults (mirroring `ReadPreferences` returning a zero-value struct) rather than erroring; **When** a read or persistence operation returns an error, the Console **shall** surface a readable inline error and **shall not** leave the page blank or panic.

### REQ-WC-011 (Where — profile selection)

**Where** more than one profile exists on the machine, the Console **shall** let the user view the list of profiles, see which profile is current, and select which profile's preferences the form reads and writes; **Where** only the default profile exists, profile selection MAY be omitted from the rendered UI.

### REQ-WC-012 (Unwanted — scope boundary)

The Console **shall not** read or write any MoAI config section outside the named scope (profile preferences + `user.yaml` + `language.yaml` + `statusline.yaml`) — specifically it **shall not** touch `quality.yaml`, `workflow.yaml`, `harness.yaml`, `git-strategy.yaml`, or the remaining config sections.

## Files to Create

| Path | Purpose |
|------|---------|
| `internal/web/` (new package) | HTTP server, route handlers, validation glue, embedded assets. Exact internal layout finalized by manager-develop in run-phase. |
| `internal/web/assets/` | `go:embed` static assets — page shell (HTML), CSS, client script (vanilla JS or HTMX). |
| `internal/web/*_test.go` | Unit + integration tests (`go test ./internal/web/...`). |
| `internal/cli/web.go` (candidate) | Thin cobra `web` subcommand wrapper that wires flags and starts `internal/web` server (mirrors `internal/cli/cc.go` / `internal/cli/brain.go` thin-command pattern). |

## Files to Modify

| Path | Change |
|------|--------|
| `internal/cli/root.go` | Register the `web` subcommand via `rootCmd.AddCommand(...)` in `init()`. |

> Note: This is a binary feature (Go source under `internal/`), NOT a deployed template. Per CLAUDE.local.md §2 Template-First rule, only template-managed `.claude`/`.moai` assets need template mirroring; `internal/web/` Go source and its embedded assets do NOT go under `internal/template/templates/` and need no template mirror.

### Out of Scope

- Read-only Workflow Tracker / Cockpit panels (SPEC-V3R3-WEB-001 capability — not resurrected)
- Auth / session / token / CSRF-token infrastructure (loopback bind + Host check only)
- External database, SSE/WebSocket/live polling, SPA/JS build toolchain
- Config sections outside the named scope (`quality.yaml`, `workflow.yaml`, `harness.yaml`, `git-strategy.yaml`, and the rest)
- Parallel validation rule set, direct YAML writes from the web layer, template mirroring of `internal/web/`

## Exclusions (What NOT to Build)

[HARD] The following are explicitly OUT OF SCOPE. Implementing any of them is an anti-pattern for this SPEC:

- **Read-only Workflow Tracker / Cockpit panels** — the polling SPEC-ID/Phase/progress.md surface from SPEC-V3R3-WEB-001 is NOT resurrected. The read-only invariant is discarded; this SPEC is CRUD.
- **Parallel/duplicate validation** — do NOT author a second validation rule set. Reuse `ValidPermissionModes`/`IsValidPermissionMode` + the `*Canonical` lists + language options. Authoring parallel validation is a forbidden anti-pattern.
- **Direct YAML writes** — do NOT marshal/write `user.yaml`/`language.yaml`/`statusline.yaml`/`preferences.yaml` from the web layer. Persist ONLY through `WritePreferences` + `SyncToProjectConfig`. Direct YAML writes from the web layer are a forbidden anti-pattern.
- **Auth / session / token infrastructure** — no login, no session store, no CSRF token store, no API keys. Loopback binding + Host-header check is the entire access/write-safety model.
- **Multi-user / remote access** — no `0.0.0.0` binding, no TLS, no reverse-proxy assumptions, no remote clients.
- **External database** — no SQLite/Postgres/embedded KV. The filesystem (existing profile + config YAML) is the only store.
- **SSE / WebSocket / live polling** — CRUD via form `GET`/`POST` is sufficient; there is no live tracker to update.
- **SPA / JS build toolchain** — no React/Vue/Bun/Hono/multi-page SPA framework, no npm/bundler step. Vanilla JS or HTMX served from `go:embed` only.
- **Out-of-scope config sections** — `quality.yaml`, `workflow.yaml`, `harness.yaml`, `git-strategy.yaml`, and the other ~26 sections are NOT editable through the Console (REQ-WC-012).
- **Theming/plugin frameworks** — no theming engine beyond exposing the existing `StatuslineTheme` field; no plugin system, no future-proofing hooks.
- **Template mirroring** — `internal/web/` is binary Go source, not a `.claude`/`.moai` template asset; it does NOT belong under `internal/template/templates/`.
