# research.md — SPEC-CLI-TUX-V3-005

> **Provenance**: Read-only investigation of the EXISTING `internal/cli/printer/printer.go` interface and all 38 `fmt.Print*` call sites in non-test `internal/cli/` sources. This research is the SSOT for migration mapping, gap analysis, and milestone ordering.

---

## §A. Printer Interface Surface (ACTUAL — from printer.go:64-112)

The Printer interface is ALREADY DESIGNED and IMPLEMENTED. This SPEC migrates onto it; it does not invent methods. The complete method catalogue:

### Core status methods (all write to stderr)

| Method | Signature | Output channel | Render-mode behavior |
|--------|-----------|----------------|---------------------|
| `Info(format, args...)` | `Info(format string, args ...any)` | stderr | TTY: colored glyph + msg; Plain: glyph + msg; JSON: `{"level":"info","msg":...}` |
| `Warn(format, args...)` | `Warn(format string, args ...any)` | stderr | TTY: colored "Warning: " prefix; JSON: `{"level":"warn",...}` |
| `Error(format, args...)` | `Error(format string, args ...any)` | stderr | TTY: colored "Error: " prefix; JSON: `{"level":"error",...}` |
| `Success(format, args...)` | `Success(format string, args ...any)` | stderr | TTY: colored ok-glyph; JSON: `{"level":"success",...}` |

### Data method (writes to stdout)

| Method | Signature | Output channel | Render-mode behavior |
|--------|-----------|----------------|---------------------|
| `Data(v any) error` | `Data(v any) error` | stdout | JSON mode: `json.Marshal(v)` single line; text modes: `fmt.Fprintln(stdout, v)` |

### Feedback handle methods (write to stderr, return handles)

| Method | Returns | Output channel | Use case |
|--------|---------|----------------|----------|
| `Step(name)` | `StepHandle` | stderr | Long-running step with `Update`/`Complete`/`Fail` |
| `Spinner(label)` | `SpinnerHandle` | stderr | Animated spinner with `Update`/`Done`/`Fail` |
| `Progress(label, total)` | `ProgressHandle` | stderr | Progress bar with `Set`/`Done`/`Fail` |
| `Mode()` | `Mode` | (query) | Reports resolved render mode (ModeTTY/ModePlain/ModeJSON) |

### What the interface DOES NOT have (gap analysis input)

- **No styled-stdout method**: there is no method for writing lipgloss-rendered rich content to stdout. `Data()` is the only stdout method, and it is designed for machine-consumable single-line output.
- **No interactive prompt method**: there is no method for writing a prompt and reading stdin. The Printer is a one-way output gateway.
- **No raw write method**: there is no `Write([]byte)` passthrough. All output is structured through the method catalogue above.

---

## §B. Call-Site Inventory (38 total, non-test, measured 2026-07-14)

Canonical grep command (from SPEC-003 acceptance.md:33):
```bash
grep -rn 'fmt\.Printf\|fmt\.Println\|fmt\.Print(' internal/cli --include='*.go' | grep -v _test.go | wc -l
```
Result: **38**

### Per-file distribution

| File | Count | Classification |
|------|-------|----------------|
| `internal/cli/uikit/banner.go` | 12 | TUI render (lipgloss styled) — **GAP** |
| `internal/cli/state.go` | 11 | CLI status + JSON data — migration target |
| `internal/cli/migration.go` | 8 | CLI status + JSON data (Korean) — migration target |
| `internal/cli/worktree/tmux_integration.go` | 5 | CLI status (session info) — migration target |
| `internal/cli/worktree/new.go` | 1 | **COMMENTED OUT** (godoc example, line 433) — dead |
| `internal/cli/branch_protection.go` | 1 | Interactive y/N prompt — **GAP** (also `nolint:unused` deferred code) |

---

## §C. Migration Mapping Per Call Site

### C.1 `internal/cli/state.go` (11 calls) — MIGRATION TARGET

| Line | Current call | Proposed Printer method | Channel change | Notes |
|------|-------------|------------------------|----------------|-------|
| 99 | `fmt.Printf("No checkpoint found for phase %s, SPEC %s\n", ...)` | `p.Info(...)` | stdout → stderr | Status message, not data |
| 110 | `fmt.Println(string(data))` (JSON) | `p.Data(state)` or `p.Data(string(data))` | stdout → stdout (preserved) | JSON mode: structured; text: string |
| 120-123 | `fmt.Printf("Phase: %s\n", ...)` (4 lines) | `p.Data(multiLineString)` or restructure | stdout → stdout (preserved) | Multi-line human format — [DECISION: Data() composed string per user 2026-07-14] |
| 125 | `fmt.Printf("Blocker: kind=%s...\n", ...)` | (part of human format block) | stdout → stdout | Conditional line within human format |
| 129 | `fmt.Printf("Checkpoint:\n %s\n", ...)` | (part of human format block) | stdout → stdout | Conditional line within human format |
| 149 | `fmt.Println("No blockers found")` | `p.Info(...)` | stdout → stderr | Status message |
| 175 | `fmt.Println("No outstanding blockers found")` | `p.Info(...)` | stdout → stderr | Status message |
| 185 | `fmt.Println(string(data))` (JSON blocker) | `p.Data(latestBlocker)` | stdout → stdout (preserved) | JSON output |

**Channel decision**: `moai state show` output is currently all-stdout. The status messages ("No checkpoint found", "No blockers found") should move to stderr (Info), while the JSON/data output stays on stdout (Data). The human-readable multi-line block (lines 120-129) is a structured data display — [DECISION 2026-07-14: route to stdout via Data() as a composed multi-line string (preserves stdout behavior for scripted consumers of `moai state show`)].

### C.2 `internal/cli/migration.go` (8 calls) — MIGRATION TARGET

| Line | Current call | Proposed Printer method | Channel change | Notes |
|------|-------------|------------------------|----------------|-------|
| 54 | `fmt.Println("실행할 pending 마이그레이션이 없습니다.")` | `p.Info(...)` | stdout → stderr | Korean status message |
| 58 | `fmt.Printf("성공: %d개 마이그레이션 적용됨 (버전: %v)\n", ...)` | `p.Success(...)` | stdout → stderr | Korean success message |
| 99 | `fmt.Println(string(data))` (JSON) | `p.Data(output)` | stdout → stdout (preserved) | JSON output (`--json` flag) |
| 104 | `fmt.Printf("현재 버전: %d\n", current)` | (human status block) | stdout → stdout or stderr | Korean human format |
| 106 | `fmt.Printf("Pending 마이그레이션 (%d개): %v\n", ...)` | (human status block) | stdout → stdout or stderr | Korean human format |
| 108 | `fmt.Println("Pending 마이그레이션 없음 (최신 상태)")` | (human status block) | stdout → stdout or stderr | Korean human format |
| 111 | `fmt.Printf("최근 적용: %s (버전 %d)\n", ...)` | (human status block) | stdout → stdout or stderr | Korean human format |
| 149 | `fmt.Printf("성공: 버전 %d로 롤백됨\n", ...)` | `p.Success(...)` | stdout → stderr | Korean success message |

**Channel decision**: Same pattern as state.go — `--json` flag uses `p.Data()`, human format is status output. The human-format block (lines 104-111) is a multi-line status display that maps to a single composed `p.Data()` call. [DECISION 2026-07-14: stdout Data() composed string, same as state.go (per user 2026-07-14)].

### C.3 `internal/cli/worktree/tmux_integration.go` (5 calls) — MIGRATION TARGET

| Line | Current call | Proposed Printer method | Channel change | Notes |
|------|-------------|------------------------|----------------|-------|
| 100 | `fmt.Printf("Tmux session created: %s\n", ...)` | `p.Info(...)` or `p.Success(...)` | stdout → stderr | Session creation result |
| 101 | `fmt.Printf("Panes created: %d\n", ...)` | `p.Info(...)` | stdout → stderr | Session creation result |
| 102 | `fmt.Printf("Attached: %v\n", ...)` | `p.Info(...)` | stdout → stderr | Session creation result |
| 103 | `fmt.Printf("Worktree path: %s\n", ...)` | `p.Info(...)` | stdout → stderr | Session creation result |
| 104 | `fmt.Printf("To attach: tmux attach-session -t %s\n", ...)` | `p.Info(...)` | stdout → stderr | Instructional output |

**Channel decision**: These are human-facing status messages about a completed tmux session creation. They should route to stderr as `p.Info()`. The function needs Printer injection (currently it has no Printer parameter — the cobra command's `RunE` must construct and pass a Printer).

**Wiring note**: The function containing these calls is invoked from a cobra `RunE` closure. The wiring pattern is: construct Printer via `printer.New(printer.WithWriters(cmd.OutOrStdout(), cmd.ErrOrStderr()))` in the `RunE`, then pass it to the function. This may require a signature change to accept a `printer.Printer` parameter.

### C.4 `internal/cli/worktree/new.go` (1 call) — DEAD CODE

| Line | Current call | Status |
|------|-------------|--------|
| 433 | `//	fmt.Println(sessionName) // Output: ...` | **COMMENTED OUT** — godoc example for `GenerateTmuxSessionName` |

**Disposition**: No migration action. This line is inside a godoc comment block (lines 430-433) illustrating function usage. It is matched by the grep but is not executable code. It counts toward the baseline (38) but cannot be "migrated" — uncommenting it would be a code change, not a migration. The ratchet AC is satisfied by migrating the other 37 calls (or any subset producing count < 38).

### C.5 `internal/cli/branch_protection.go` (1 call) — GAP (deferred code)

| Line | Current call | Status |
|------|-------------|--------|
| 44 | `fmt.Printf("%s [y/N]: ", prompt)` | Interactive prompt in `ttyConfirmer.Confirm()` |

**Context**: `ttyConfirmer` is annotated `nolint:unused` with comment "SPEC-V3R6-CI-BASELINE-DRIFT-001 §D.1 deferred (interactive prompt path)". The ACTIVE code path uses `yesConfirmer` (always returns true) with the `--yes-branch-protection` flag. `ttyConfirmer` is RETAINED for a follow-up interactive prompt SPEC but is currently dead code.

**Source self-contradiction note (D5)**: the `Confirmer` interface doc (branch_protection.go:19-20) says "uses an interactive TTY confirmer by default", while the `ttyConfirmer` type comment says "Currently unwired". The dead-code premise rests on the `nolint:unused` linter directive (golangci-lint's unused analyzer), which is the authoritative signal — the doc/type-comment contradiction is in the source, not a SPEC reasoning error.

**Gap**: The Printer interface has NO interactive prompt method. The Printer is a one-way output gateway (stderr/stdout writes only); it does not read stdin.

**Disposition options**:
1. **Exclude from scope**: `ttyConfirmer` is deferred dead code; migrating it adds no behavioral value. Record as out-of-scope with rationale.
2. **Migrate the Printf to `p.Info(prompt)`**: but the prompt is an interactive write-expecting-read, and Info adds a newline + glyph, breaking the prompt format.
3. **Defer to the follow-up interactive prompt SPEC**: when `ttyConfirmer` is activated, the Printer may need a `Prompt` method or the prompt stays on a raw writer.

**Recommendation**: Option 1 (exclude from scope) — CONFIRMED. The line counts toward the baseline (38) but is dead code. [DECISION 2026-07-14: EXCLUDED as linter-confirmed dead code (nolint:unused). No migration; the 24 migratable sites suffice for the ratchet.]

### C.6 `internal/cli/uikit/banner.go` (12 calls) — GAP (TUI render)

| Lines | Current calls | Pattern |
|-------|--------------|---------|
| 90-101 | `PrintBanner` — 7 calls | `fmt.Println(bannerStyle.Render(...))`, `fmt.Println(dimStyle.Render(...))`, `fmt.Println()` (blank lines), `fmt.Println("  " + pillRow)` |
| 113-117 | `PrintWelcomeMessage` — 5 calls | `fmt.Println(titleStyle.Render(...))`, `fmt.Println(dimStyle.Render(...))`, `fmt.Println()` |

**Context**: `PrintBanner` and `PrintWelcomeMessage` render rich TUI content using lipgloss styles (`bannerStyle`, `dimStyle`, `titleStyle`). The styled content is pre-rendered via `lipgloss.Style.Render()` into a string, then written to stdout via `fmt.Println`. Blank lines (`fmt.Println()`) are spacing elements in the visual layout.

**Gap**: The Printer interface has NO method for writing pre-styled rich content to stdout. `Data()` is designed for machine-consumable single-line output (`fmt.Fprintln(stdout, v)` — it would work for a string but semantically mislabels banner output as "data"). There is no `RawStdout(string)` or `Styled(string)` method.

**This is the CRITICAL architecture decision (M1 gate)**: 12 of 38 calls are in this category. Options:

1. **Add a Printer method for styled/raw stdout** — e.g. `Render(s string)` that writes pre-styled content to stdout. This extends the interface and enables banner migration. Risk: the Printer's channel-discipline model (stdout=data, stderr=status) is muddied by adding a "styled stdout" method that is neither pure data nor pure status.
2. **Exclude banner.go as TUI render out-of-scope** — banner output is architecturally distinct from CLI status/data output. It is a visual branding element rendered at command entry, not a command-result output. The ratchet AC is satisfied without these 12 calls.
3. **Route banner through Data()** — treat the banner string as data output. Semantically incorrect (banner is not machine-consumable data), but technically works since `Data()` does `fmt.Fprintln(stdout, v)` for string values in text modes.

**Recommendation**: Option 2 (exclude as TUI render) — CONFIRMED. Banner rendering is a distinct architectural concern from CLI output-channel discipline. The Printer interface was designed for status/data events, not for rich visual layout. Forcing banner through Printer would either require an interface extension (option 1, adding complexity) or semantic misuse of Data() (option 3). [DECISION 2026-07-14: EXCLUDED as TUI render (lipgloss). 12 calls remain in baseline; ratchet met by the 24 migratable sites. No interface extension.]

---

## §D. Gap Analysis Summary

| Gap type | Call sites | Count | Proposed disposition |
|----------|-----------|-------|---------------------|
| Interactive prompt (no Printer method) | `branch_protection.go:44` | 1 | Exclude (dead code, `nolint:unused`) |
| TUI render (no Printer method) | `uikit/banner.go:90-117` | 12 | Exclude (architecturally distinct from CLI status/data) |
| **Total gap** | | **13** | |
| **Migratable** | state.go(11) + migration.go(8) + tmux(5) | **24** | Migrate to Printer methods |
| **Dead code** | worktree/new.go:433 | **1** | No action (godoc comment) |
| **TOTAL** | | **38** | |

**Ratchet impact**: migrating 24 calls → count drops from 38 to 14 (well below baseline). Even migrating just the 5 tmux_integration.go calls → count drops to 33 (below baseline). The ratchet AC is achievable with ANY subset of the 24 migratable calls.

---

## §E. Existing Output-Abstraction Patterns (cross-reference)

The codebase already mixes several output strategies. This SPEC consolidates toward Printer but does NOT retroactively refactor these existing patterns (they use `io.Writer` / `fmt.Fprint` with explicit writers, not bare `fmt.Print*`):

| Pattern | Location | Strategy | In scope? |
|---------|----------|----------|-----------|
| Confirmer interface | `branch_protection.go:7-11` | Dependency injection | No (not fmt.Print*) |
| GhClient interface | `branch_protection.go:25-32` | Dependency injection | No |
| `printJSONResult(out interface{...})` | `doctor_permission.go` | `io.Writer` parameter | No (uses writer, not bare fmt) |
| `MergeFallbackRecorder` | `update/backup/restore.go` | `func(..., errOut io.Writer)` | No (uses writer) |
| `fmt.Fprintf(cmd.ErrOrStderr(), ...)` | `harness.go:254`, `update.go:229+` | Explicit cobra writer | No (uses cmd writer, not bare fmt) |
| `fmt.Fprintln(out, tui.CheckLine(...))` | `update.go:229,239,278,...` | Explicit `out` writer | No (uses writer) |

These patterns are NOT matched by the ratchet grep (`fmt.Fprintf` with an explicit writer is not `fmt.Printf`/`Println`/`Print(`). They are architecturally sound (explicit writer parameter) and do not need migration.

---

## §F. Migration Ordering by Risk

Ordered by decision-reversibility (highest-reversibility first):

1. **M1 — Architecture gate (banner.go + branch_protection.go scope decision)**: RESOLVED 2026-07-14. Both gap sites EXCLUDED (banner.go as TUI render, branch_protection.go as dead code). The Printer interface needs NO extension. The gate is cleared — proceed directly to M2.
2. **M2 — state.go (11 calls)**: medium risk. Mix of JSON data and human format. Channel routing decision needed. Characterization tests required.
3. **M3 — migration.go (8 calls)**: medium risk. Same pattern as state.go, Korean strings. Characterization tests required.
4. **M4 — worktree/tmux_integration.go (5 calls)**: lowest risk. Pure status output, direct Info() mapping. Needs Printer wiring (signature change).
5. **M5 — Ratchet verification + gap-site documentation**: verify count < 38, record gap dispositions.

---

## §G. Tier Assessment

**Proposed: Tier M (standard)** — confirmed.

Justification:
- Single domain: CLI output layer (`internal/cli/`).
- Interface already exists: Printer is designed and tested (SPEC-001 completed).
- 5 non-test files affected (state.go, migration.go, tmux_integration.go, banner.go, branch_protection.go).
- Migration is behavior-preserving (DDD ANALYZE-PRESERVE-IMPROVE), not new-feature development.
- The interface-design risk that would justify Tier L is RETIRED (Printer exists since SPEC-001).

The one Tier-L signal (banner.go architecture decision) is a SCOPE gate, not a design-from-scratch effort. If the M1 gate decides to extend the Printer interface (option 1), the SPEC may be re-tiered to L. Default disposition (exclude banner.go) keeps this at Tier M.
