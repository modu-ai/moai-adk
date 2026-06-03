# Plan — SPEC-WEB-CONSOLE-001: MoAI Web Console (Settings CRUD)

## A. Context

This plan decomposes the browser-based settings Console into 7 ordered implementation phases. Unlike the predecessor SPEC-V3R3-WEB-001 (which deferred the stack decision to run-phase), the technology stack here is **locked by user decision**: Go single binary, `net/http` (stdlib, prefer zero new heavy deps), `go:embed` for static assets, lightweight frontend (vanilla JS or HTMX — pick the simplest that works), exposed as a `moai web` cobra subcommand. The plan therefore fixes concrete technology decisions while leaving exact internal file layout to manager-develop.

PRESERVE targets (do NOT reimplement): `internal/profile/` read/write/sync logic, the canonical validation value lists, and the existing terminal profile wizard. EXTEND targets: a new `internal/web/` package + a thin `web` subcommand registered in `internal/cli/root.go`.

## B. Tier Classification

**Tier M (Medium).** Rationale:
- New `internal/web/` package (HTTP server + handlers + validation glue + embedded assets) plus a thin CLI subcommand and one root.go registration edit.
- CRUD over ~14 `ProfilePreferences` fields + 3 project config sections, with a frontend asset bundle and integration tests.
- Estimated scope: roughly 5–15 files affected, 300–1000 LOC (new package + assets + tests), well within the Tier M envelope (and clearly above Tier S, below Tier L constitutional/>15-file threshold).
- plan-auditor PASS threshold for Tier M: **0.80**.
- Artifact set: 3 files (spec.md + plan.md + acceptance.md) — produced.
- Section A–E delegation template REQUIRED for run-phase (Tier M).

## C. Technology Decisions (LOCKED — not deferred)

| Area | Decision | Rationale |
|------|----------|-----------|
| HTTP server | Go stdlib `net/http` (`http.Server` + `http.ServeMux`) | Zero new heavy deps; loopback bind + graceful shutdown are stdlib-native (`http.Server.Shutdown(ctx)`). |
| Static assets | `go:embed` under `internal/web/assets/` | Single-binary distribution, no runtime fetch, no build toolchain. |
| Frontend | Vanilla JS form, OR HTMX if it reduces handler/markup complexity | Simplest that works. HTMX was the predecessor's candidate and is acceptable; a plain HTML `<form>` + `POST` is also acceptable. No SPA framework. |
| Validation | Reuse `internal/profile` + `internal/cli/profile_setup` canonical lists | `ValidPermissionModes`/`IsValidPermissionMode`, `statuslineModeCanonical`, `statuslinePresetCanonical`, `statuslineThemeCanonical`, language `en`/`ko`/`ja`/`zh`. No parallel validation. |
| Persistence | Reuse `internal/profile` write/sync | `WritePreferences(name, prefs)` for the profile store; `SyncToProjectConfig(projectRoot, prefs)` for `user.yaml`/`language.yaml`/`statusline.yaml`. No direct YAML writes. |
| CLI | `moai web [--port N] [--no-open]` registered in `internal/cli/root.go` | Mirrors the thin-command pattern of `internal/cli/cc.go` / `internal/cli/brain.go`. |
| Write-safety | Loopback bind + `Host`-header check on mutating requests | Minimal CSRF/DNS-rebinding defense; no token auth, no session store. |

## D. Implementation Phases

### Phase 1 — CLI subcommand registration + flag parsing (REQ-WC-001)
- Add a thin `web` cobra command (candidate `internal/cli/web.go`) with `--port <int>` (default 8080) and `--no-open` (bool).
- Register via `rootCmd.AddCommand(...)` in `internal/cli/root.go` `init()` (mirror `cc.go`/`brain.go`).
- Verify `moai web --help` lists both flags; exit-code discipline per `internal/cli/CLAUDE.md` (0 success, 1 user error, 2 system error).
- Subagent boundary: NO `AskUserQuestion` in CLI code (C-HRA-008); add the `TestWeb_NoAskUserQuestion` static guard mirroring `worktree/new_test.go`.
- Deliverable: subcommand + flag struct + registration + tests. Depends on: none.

### Phase 2 — HTTP server scaffold + loopback bind + graceful shutdown (REQ-WC-002, REQ-WC-003, REQ-WC-004)
- Bootstrap `http.Server` bound to `127.0.0.1:<port>` (HARD: never `0.0.0.0`).
- Register `GET /` (replaced by READ handler in Phase 4) + a static asset handler.
- SIGINT/SIGTERM handling → `http.Server.Shutdown(ctx)` with a 5s drain context → exit 0.
- `--no-open` handling: when false, attempt to open `http://127.0.0.1:<port>`; auto-open failure is non-fatal (server continues).
- Tests: lifecycle (start/stop), port-conflict error path, loopback-bind assertion.
- Deliverable: server bootstrap + signal handler + auto-open + tests. Depends on: Phase 1.

### Phase 3 — Embedded assets + page shell (REQ-WC-005)
- `go:embed` directive over `internal/web/assets/` (HTML page shell + CSS + client script).
- Page shell renders an empty form container that the READ handler populates.
- Tests: assets are embedded (served bytes match), `Content-Type` correct, no network fetch of frontend deps.
- Deliverable: embedded FS + page shell + asset handler + tests. Depends on: Phase 2.

### Phase 4 — READ handlers (REQ-WC-006, REQ-WC-010, REQ-WC-011)
- `GET /` renders current settings as a pre-populated editable form:
  - `ProfilePreferences` for the selected profile via `ReadPreferences(name)` (zero-value when no `preferences.yaml` → neutral defaults, not an error).
  - Project-level `user.yaml`/`language.yaml`/`statusline.yaml` current values.
- Profile selection (REQ-WC-011): list profiles + mark current + allow selecting which profile the form binds to. Omit selection UI when only the default profile exists.
- Graceful empty/error states (REQ-WC-010): neutral defaults for missing config; readable inline error on read failure; never blank, never panic.
- Tests: populated form for an existing profile; neutral-default form for a zero-value profile; multi-profile selection rendering; read-error path.
- Deliverable: READ handler(s) + profile-list rendering + tests. Depends on: Phase 3.

### Phase 5 — WRITE handlers reusing profile/sync + existing validation (REQ-WC-007, REQ-WC-008, REQ-WC-012)
- Form submit (`POST`) → bind form values to a `ProfilePreferences` + project-settings update.
- Validate via existing predicates ONLY: `IsValidPermissionMode`, `statusline*Canonical` membership, language-option membership. No parallel rules.
- On success: persist via `WritePreferences(name, prefs)` (profile store) + `SyncToProjectConfig(projectRoot, prefs)` (user/language/statusline). No direct YAML writes.
- On validation failure (REQ-WC-008): reject mutation, leave persisted state unchanged, re-render form with per-field error.
- Scope boundary (REQ-WC-012): touch ONLY the named scope; never quality/workflow/harness/git-strategy or other sections.
- Tests: valid submit round-trips through `WritePreferences`/`SyncToProjectConfig` (read-back asserts persisted values); invalid `PermissionMode`/theme rejected + state unchanged; out-of-scope sections untouched.
- Deliverable: WRITE handler(s) + validation glue + persistence glue + tests. Depends on: Phase 4.

### Phase 6 — Host-check middleware (REQ-WC-009)
- Middleware on mutating methods (`POST`/`PUT`/`PATCH`): verify `Host` resolves to a loopback origin (`127.0.0.1:<port>` or `localhost:<port>`); reject non-matching Host with HTTP 403.
- Minimal by design: no token auth, no session store, no CSRF-token infrastructure.
- Tests: loopback Host passes; foreign Host on a `POST` → 403 + state unchanged; `GET` is not Host-gated (read-only is safe).
- Deliverable: Host-check middleware + tests. Depends on: Phase 5.

### Phase 7 — Golden-path integration test (all REQs)
- Bootstrap a temp env with `t.TempDir()` (isolated profile dir + project config sections).
- Start the Console on a random free port (avoid 8080 conflicts).
- `GET /` → assert current values rendered.
- `POST` a valid change → assert HTTP 2xx → read back via `ReadPreferences`/config files → assert persisted (round-trip).
- `POST` an invalid value → assert rejection + unchanged state.
- `POST` with a foreign `Host` header → assert 403.
- Assert graceful shutdown on signal (server returns, exit 0 path).
- Deliverable: golden-path E2E test asserting READ→WRITE→round-trip + Host-check + empty-state. Depends on: Phases 1–6.

## E. Self-Verification (run-phase exit gate)

manager-develop reports completion with: AC PASS/FAIL matrix (acceptance.md SSOT), `go build ./...` + `GOOS=windows GOARCH=amd64 go build ./...` exit 0, `go test ./internal/web/...` green with ≥85% package coverage, subagent-boundary grep clean (`grep -rn 'AskUserQuestion' internal/web internal/cli/web.go | grep -v _test.go | grep -v '//'` → 0), `golangci-lint run` NEW-issue-free, and `moai web --help` output.

## F. Risk Analysis & Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| R1: Port conflict (8080 in use) | Server start fails | Clear stderr error + non-zero exit on bind failure; `--port` flag escape. |
| R2: Missing signal handling → zombie process | Ctrl+C does not terminate | SIGINT/SIGTERM handler + 5s drain + `Shutdown(ctx)` + exit 0 (verified by integration test). |
| R3: DNS-rebinding/CSRF from another local origin | Unauthorized mutation | Host-header check on mutating methods (REQ-WC-009); reject 403. Loopback bind is the outer boundary. |
| R4: Parallel validation drift | Browser accepts values the wizard rejects (or vice-versa) | Reuse existing predicates only; no parallel rule set (REQ-WC-008). |
| R5: Direct YAML write bypasses sync semantics | `statusline.yaml` segment defaults / migration lost | Persist ONLY via `WritePreferences` + `SyncToProjectConfig` (REQ-WC-007); the latter owns `syncStatusline` defaults. |
| R6: Cross-platform build break (signal/exec) | Windows build fails | Use stdlib `os/signal` (cross-platform). If any platform-specific primitive is needed, isolate with build tags; verify `GOOS=windows GOARCH=amd64 go build ./...`. |
| R7: Over-engineering creep (auth/SPA/DB) | Scope bloat, Goal Anti violation | Exclusions section is HARD; reviewer rejects any auth/session/DB/SPA/SSE addition. |

## G. MX Tag Plan

[MANDATORY] Place the following MX tags during run-phase (per `.claude/rules/moai/workflow/mx-tag-protocol.md`; `code_comments: ko` → Korean descriptions):

- **@MX:ANCHOR — Console server entry function** (Phase 2 deliverable). Reason: high fan_in invariant contract — the server entry point is referenced by the CLI subcommand, unit tests, and the integration test; mark the loopback-bind + graceful-shutdown invariant. @MX:REASON required.
- **@MX:WARN — settings WRITE / persist path** (Phase 5 deliverable). Reason: danger zone — this is the only code path that mutates on-disk user/project settings. @MX:REASON required (e.g. "must persist only via WritePreferences + SyncToProjectConfig; never direct YAML write").
- **@MX:NOTE — Host-check middleware** (Phase 6 deliverable). Reason: documents the deliberate minimal write-safety model (loopback + Host check, no token auth) so a future reader does not "add CSRF tokens for completeness".
- **@MX:TODO — per-phase incompletion hooks** (Phases 2–7). Resolved (removed) when each phase's acceptance scenario passes; unresolved @MX:TODO blocks the sync-phase quality gate.

## H. Dependencies

No new external Go dependencies are required or permitted by default. Allowed building blocks:
- Go stdlib: `net/http`, `os/signal`, `embed`, `context`, `path/filepath`.
- Existing internal packages: `internal/profile` (read/write/sync + validation), `internal/cli/profile_setup` (canonical value lists), `internal/cli` (root command registration).

[HARD] Adding HTMX is the ONLY acceptable client-side asset, served as an embedded static file (no npm). Adding a Go router/template dependency is discouraged — `net/http` + `html/template` (stdlib) suffice. Any new dependency decision is a run-phase blocker report to the orchestrator, not a silent `go get`.
