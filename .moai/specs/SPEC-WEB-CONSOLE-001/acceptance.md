# Acceptance — SPEC-WEB-CONSOLE-001: MoAI Web Console (Settings CRUD)

Acceptance criteria are mapped 1:1 to the GEARS requirements in `spec.md`. Each AC is observable (test output, HTTP status, persisted file state, or CLI exit code). This `acceptance.md` is the SSOT for the AC count.

## D. AC Matrix

| AC | Requirement | Verification |
|----|-------------|--------------|
| AC-WC-001 | REQ-WC-001 | `moai web --help` exits 0 and lists `--port` and `--no-open` |
| AC-WC-002 | REQ-WC-002 | Server bound to `127.0.0.1`, never `0.0.0.0` |
| AC-WC-003 | REQ-WC-003 | SIGINT/SIGTERM → 5s drain → exit 0 |
| AC-WC-004 | REQ-WC-004 | `--no-open` suppresses open; auto-open failure non-fatal |
| AC-WC-005 | REQ-WC-005 | Page/CSS/JS served from `go:embed`, no network fetch |
| AC-WC-006 | REQ-WC-006 | `GET /` renders current settings as pre-populated form |
| AC-WC-007 | REQ-WC-007 | Valid submit persists via `WritePreferences` + `SyncToProjectConfig` (round-trip) |
| AC-WC-008 | REQ-WC-008 | Invalid value rejected via existing validation; state unchanged |
| AC-WC-009 | REQ-WC-009 | Foreign `Host` on mutating request → HTTP 403 |
| AC-WC-010 | REQ-WC-010 | Zero-value profile / missing section → neutral defaults; read error → inline error, no panic |
| AC-WC-011 | REQ-WC-011 | Multi-profile: list + current marker + selection; single profile may omit UI |
| AC-WC-012 | REQ-WC-012 | Out-of-scope config sections never read/written |

## Scenarios (Given-When-Then)

### AC-WC-001 — CLI subcommand + flags
- **Given** the `moai` binary is built,
- **When** the user runs `moai web --help`,
- **Then** exit code is 0 and the output lists both `--port` (int, default 8080) and `--no-open` (bool).

### AC-WC-002 — loopback-only bind
- **Given** the Console server is started,
- **When** the listener address is inspected,
- **Then** it is `127.0.0.1:<port>` and binding to `0.0.0.0` or any non-loopback address does not occur.

### AC-WC-003 — graceful shutdown
- **Given** the Console server is running,
- **When** the process receives SIGINT or SIGTERM,
- **Then** it stops accepting new connections, drains in-flight requests for up to 5s, closes the listener, and exits with status 0.

### AC-WC-004 — browser auto-open
- **Given** the server is starting,
- **When** `--no-open` is set, **Then** no browser-open is attempted;
- **When** `--no-open` is not set and the open attempt fails, **Then** the server continues serving (failure is non-fatal, exit code unaffected).

### AC-WC-005 — embedded assets
- **Given** the binary with `go:embed` assets,
- **When** `GET /` and the static asset routes are requested,
- **Then** the page shell, CSS, and client script are served from embedded bytes with correct `Content-Type`, requiring no network fetch and no build toolchain.

### AC-WC-006 — settings READ renders current values
- **Given** a profile with existing `preferences.yaml` and present `user.yaml`/`language.yaml`/`statusline.yaml`,
- **When** the user requests `GET /`,
- **Then** the form is pre-populated with the current `ProfilePreferences` field values (read via `ReadPreferences`) and the current project-section values.

### AC-WC-007 — settings WRITE round-trips through existing persistence
- **Given** the Console form is open for a profile,
- **When** the user submits a valid change (e.g. `ConversationLang` `en`→`ko`, `StatuslineTheme`→`catppuccin-latte`),
- **Then** the server validates via existing predicates, calls `WritePreferences(name, prefs)` and `SyncToProjectConfig(projectRoot, prefs)`, returns 2xx, and a subsequent `ReadPreferences` + read of `language.yaml`/`statusline.yaml` reflects the new values (round-trip confirmed). No direct YAML write occurs.

### AC-WC-008 — invalid input rejected via existing validation
- **Given** the Console form is open,
- **When** the user submits an unrecognized `PermissionMode` (fails `IsValidPermissionMode`) or an unrecognized statusline theme (not in `statuslineThemeCanonical`),
- **Then** the mutation is rejected, persisted state is unchanged, the form re-renders with a per-field error, and the rejection used the existing validation predicate (not a parallel rule).

### AC-WC-009 — write-safety Host check
- **Given** the Console server is running,
- **When** a `POST`/`PUT`/`PATCH` arrives with a `Host` header that does not resolve to a loopback origin (`127.0.0.1:<port>` / `localhost:<port>`),
- **Then** the server responds HTTP 403 and persisted state is unchanged;
- **And** a `GET` request is not Host-gated (read remains accessible).

### AC-WC-010 — graceful empty/error states
- **Given** a profile with no `preferences.yaml` (zero-value preferences) or an absent project config section,
- **When** the user opens the Console,
- **Then** the form renders with neutral defaults (mirroring `ReadPreferences` zero-value) rather than an error;
- **And** when a read or persistence call returns an error, a readable inline error is shown and the page is neither blank nor panicking.

### AC-WC-011 — profile selection
- **Given** more than one profile exists on the machine,
- **When** the user opens the Console,
- **Then** the profile list is shown, the current profile is marked, and the user can select which profile's preferences the form reads/writes;
- **And** when only the default profile exists, the selection UI MAY be omitted.

### AC-WC-012 — scope boundary
- **Given** the Console is running,
- **When** any READ or WRITE operation executes,
- **Then** only profile preferences + `user.yaml` + `language.yaml` + `statusline.yaml` are touched, and `quality.yaml`/`workflow.yaml`/`harness.yaml`/`git-strategy.yaml` and the other sections are never read or written (verified by asserting those files are unchanged after a WRITE).

## Edge Cases

- **EC-1 — Port already in use**: `moai web --port 8080` when 8080 is bound → clear stderr error + non-zero exit (not a panic).
- **EC-2 — Concurrent submits**: two rapid valid submits → last-writer-wins through `WritePreferences`/`SyncToProjectConfig`; no partial/corrupt YAML (the existing write functions own atomicity semantics).
- **EC-3 — Deprecated `Bypass` field**: a profile with legacy `Bypass: true` reads back as `PermissionMode: "bypassPermissions"` (existing `ReadPreferences` migration); the form shows the migrated value.
- **EC-4 — `StatuslineSegments` map**: custom-preset segment toggles (`map[string]bool`) round-trip through the form without dropping keys.
- **EC-5 — Empty `StatuslineSegments` + absent statusline.yaml**: WRITE applies `syncStatusline` defaults (REQ-WC-010 + R5) rather than writing an empty map.

## Quality Gate Criteria

- `go build ./...` → exit 0
- `GOOS=windows GOARCH=amd64 go build ./...` → exit 0 (cross-platform)
- `go test ./internal/web/...` → all green
- Package coverage `internal/web` ≥ 85%
- `golangci-lint run --timeout=2m` → no NEW issues
- Subagent boundary: `grep -rn 'AskUserQuestion\|mcp__askuser' internal/web internal/cli/web.go | grep -v _test.go | grep -v '//'` → 0 matches
- `moai web --help` → exit 0, lists `--port` + `--no-open`

## Definition of Done

- [ ] All 12 ACs (AC-WC-001 … AC-WC-012) PASS with cited evidence.
- [ ] All 5 edge cases (EC-1 … EC-5) handled and tested.
- [ ] Quality-gate criteria above all pass.
- [ ] Persistence path verified to go exclusively through `WritePreferences` + `SyncToProjectConfig` (no direct YAML write in `internal/web`).
- [ ] Validation path verified to reuse existing predicates only (no parallel validation rule set in `internal/web`).
- [ ] Loopback bind + Host-check are the only access/write-safety mechanisms (no auth/session/token code present).
- [ ] No out-of-scope config section is read or written (REQ-WC-012 / AC-WC-012).
- [ ] MX tags placed per plan.md §G (ANCHOR on server entry, WARN on write/persist path, NOTE on Host-check); no unresolved @MX:TODO at sync.
