---
id: SPEC-WEB-CONSOLE-001
title: "MoAI Web Console — Browser-based Settings CRUD"
version: "0.2.0"
status: implemented
created: 2026-06-03
updated: 2026-06-03
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
