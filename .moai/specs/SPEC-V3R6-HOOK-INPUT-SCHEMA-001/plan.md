# Implementation Plan — SPEC-V3R6-HOOK-INPUT-SCHEMA-001

> Tier S. WHAT/WHY lives in `spec.md`. This plan covers HOW: the chosen mechanisms, milestone ordering, the REQ→AC→file map, and risks. No time estimates (priority labels + phase ordering only).

## §A. Context

Two confirmed `internal/hook` stdin-parse defects, both verified against current source this session:

1. **`globs` type mismatch** — `HookInput.Globs` is typed `string` (types.go:255) but Claude Code's InstructionsLoaded event sends a JSON array. `json.Unmarshal` fails on every rule-file load.
2. **Empty-stdin robustness gap** — `ReadInput` (protocol.go:26) raises `ErrHookInvalidInput` on empty stdin; the dispatcher returns non-zero, which the harness surfaces as a Bash tool failure for PreToolUse.

Both fixes are minimum-mechanism, single-package, and align with the existing graceful-fallback philosophy already present in `validateInput`.

## §B. Known Facts (verified, not assumed)

- `Globs` has **zero read consumers** in the package — confirmed by `grep -rn '\.Globs' internal/hook/ --include='*.go'` (only the struct declaration matches). No string-form consumer exists to regress.
- The flexible-typing precedent (`json.RawMessage`) already lives in the **same struct**: `ElicitationRequest` (types.go:249), `PermissionSuggestions` (types.go:260).
- `ReadInput` is the single shared reader for all events (protocol.go:26); fixing empty-stdin there benefits every event handler uniformly.
- `validateInput` already defaults missing `session_id` / `cwd` / `hook_event_name` rather than erroring (protocol.go:75-101) — the no-op-on-empty extends this established intent.
- The CLI dispatcher `runHookEvent` (hook.go:173-176) returns the `ReadInput` error verbatim → non-zero exit. No CLI-layer change is required if the reader returns a usable `*HookInput` + nil error for empty input.

## §C. Technical Approach

### C.1 Globs type choice (decision)

**Chosen: `Globs []string`** (type-safe slice), NOT a custom `UnmarshalJSON` and NOT `json.RawMessage`.

Rationale:
- The diagnostic evidence shows Claude Code (v2.1.69+) emits `globs` **as an array**. `[]string` decodes the array form natively with zero custom code — the minimum mechanism that satisfies REQ-HIS-001.
- There is **no existing string-form consumer** to preserve (verified `grep`), so the "preserve compatibility" clause in REQ-HIS-001 is satisfied trivially — nothing reads `Globs` today, so changing `string`→`[]string` regresses nothing.
- A custom `UnmarshalJSON` accepting *both* a bare string and an array would be defensible **only if** an older Claude Code emitted the bare-string form AND a consumer depended on it. Neither holds. Adding a dual-form decoder here is speculative flexibility (anti-overengineering: "Don't build abstractions for hypothetical future requirements") — rejected for Tier S.
- `json.RawMessage` (the sibling-field pattern) defers decoding to the consumer. Since the only realistic consumer wants the patterns as a list, `[]string` gives a directly-usable typed value with no downstream decode step. Preferred over `RawMessage` here.

Escape hatch (documented, not built): if a future Claude Code version reverts to emitting a bare string, the field can be promoted to a custom `UnmarshalJSON` that accepts both forms at that time — a localized, test-driven change. This SPEC does not pre-build it.

### C.2 Empty-stdin graceful no-op (decision)

**Fix site: `ReadInput` in `internal/hook/protocol.go`** (the shared reader source), not the per-event CLI commands.

Approach: after `io.ReadAll`, if the read bytes are empty or contain only whitespace, short-circuit and return a minimal valid `*HookInput` (with `validateInput`'s defaults applied — `session_id: "unknown"`, `hook_event_name` injected by the caller via `runHookEvent`) and a `nil` error. The caller (`runHookEvent`) then proceeds normally, dispatches to the handler with a default input, writes an empty `HookOutput`, and exits 0.

Why the reader source (not the CLI layer):
- One change fixes empty-stdin for **all** events uniformly (PreToolUse, InstructionsLoaded, Stop, ...), not just `pre-tool`.
- It matches the existing graceful-fallback design already embodied in `validateInput`.
- It keeps the CLI dispatcher (`runHookEvent`) untouched — lower blast radius.

Malformed-but-non-empty JSON: out of the strict requirement, but per spec §A the hook is a non-blocking observer. The minimum change targets the empty/blank/whitespace case (the confirmed symptom). Non-empty-malformed handling MAY be left returning `ErrHookInvalidInput` unless the implementer finds the same tool-failure propagation; if so, the same no-op-default pattern applies. Keep the change scoped to what AC-2 asserts.

### C.3 Tests (table-driven, Go convention)

- `types_test.go` (or `instructions_loaded_test.go`): add an array-form `globs` decode case → AC-1.
- `protocol_test.go`: **rewrite the two EXISTING cases at `protocol_test.go:183-192`** — `name: "empty stdin"` (input `""`) and `name: "whitespace only stdin"` (input `"   "`) — from `wantErr: true / errTarget: ErrHookInvalidInput` to `wantErr: false` + a non-nil zero-value `*HookInput` assertion. These are **intentional behavior inversions of the current pinned behavior, not new test rows** — the auditor (D1) confirmed these two cases currently pin the old erroring behavior and must flip. Add/confirm a well-formed-input case asserting unchanged field values → AC-3.
- CLI smoke (`internal/cli/hook_test.go`) — **mandatory** (D3): `moai hook pre-tool </dev/null` exits 0 → AC-4.

## §D. Milestones (priority-ordered, no time estimates)

| Milestone | Work | REQ | AC |
|-----------|------|-----|-----|
| M1 | RED: add array-form `globs` decode test; **rewrite the existing `empty stdin` + `whitespace only stdin` cases at `protocol_test.go:183-192` from `wantErr:true / errTarget:ErrHookInvalidInput` to `wantErr:false` + non-nil zero-value `*HookInput` assertion** — these are intentional behavior inversions, not new tests (D1) | REQ-HIS-001/002 | AC-1, AC-2 |
| M2 | GREEN: change `Globs string` → `Globs []string` in types.go:255 | REQ-HIS-001 | AC-1 |
| M3 | GREEN: empty/whitespace short-circuit in `ReadInput` returning default `*HookInput` + nil err | REQ-HIS-002 | AC-2 |
| M4 | Regression: full `go test ./internal/hook/...` green. **The green-suite assertion explicitly EXCLUDES the 2 inverted cases at protocol_test.go:183-192 — those are expected to change from erroring to no-op success, so "stays green" means green AFTER their rewrite, not unchanged (D1).** Confirm well-formed input unchanged | REQ-HIS-003 | AC-3 |
| M5 | **Mandatory (D3)**: CLI smoke `moai hook pre-tool </dev/null` exits 0 — the empty-stdin reader change governs this exact path (user-visible symptom) | REQ-HIS-002 | AC-4 |

Development mode: TDD (RED before GREEN per project `quality.yaml`). M1 must capture the array-form `globs` failure AND flip the two empty/whitespace cases before M2/M3 fix them.

## §E. Risks and Mitigations

| Risk | Likelihood | Mitigation |
|------|-----------|------------|
| `[]string` breaks a hidden `Globs` string consumer | Very low — `grep` confirmed none exists | Re-run `grep -rn '\.Globs' internal/hook/` at M2; if a consumer appears, switch to custom `UnmarshalJSON` |
| Empty-stdin no-op masks a genuine truncated-payload bug elsewhere | Low | Scope the short-circuit to empty/blank/whitespace only; non-empty-malformed still errors unless tool-failure propagation is reconfirmed |
| **(D2 run-phase gate) The live `pre-tool` failure payload may have been truncated-non-empty, not strictly-empty** | Medium — the Go error `unexpected end of JSON input` is identical for empty OR truncated JSON | **manager-develop MUST, before M3, reproduce/confirm from `~/.moai/logs/hook-stderr.log` whether the live `pre-tool` failure payload was strictly-empty (`""`/whitespace) or truncated-non-empty (partial JSON). If truncated, surface a BLOCKER — AC-2's empty-only boundary would not resolve the symptom and the scope needs a decision. Do NOT broaden AC-2 autonomously; this is a run-phase confirmation gate (D2).** |
| Type change ripples to JSON-encode paths | Very low | `Globs` is decode-only (input struct); `json:"globs,omitempty"` on a slice omits empty correctly |
| Coverage drop in `internal/hook` | Low | M1 adds tests before the change; net coverage non-decreasing |

## §F. Anti-Patterns to Avoid

- **Dual-form custom decoder built speculatively** — rejected in C.1; no string-form producer or consumer exists. Build the minimum (`[]string`).
- **Fixing empty-stdin in each CLI subcommand** — fix the shared `ReadInput` source once (C.2).
- **Swallowing non-empty malformed JSON silently as part of this SPEC** — keep scope to the empty/blank case asserted by AC-2.
- **Touching settings.json / handle-*.sh** — out of scope per spec §B; the bug is in the Go decode path.

## §G. Cross-References

- `spec.md` §A (diagnostic evidence + verified line numbers), §C (GEARS REQ), §3 (AC)
- `internal/hook/types.go:255` (Globs), `:249`/`:260` (RawMessage precedent)
- `internal/hook/protocol.go:26` (ReadInput), `:75-101` (validateInput graceful-fallback precedent)
- `internal/hook/errors.go:16` (ErrHookInvalidInput)
- `internal/cli/hook.go:173-176` (dispatcher error propagation)
- `internal/hook/CLAUDE.md` (module conventions — JSON I/O contract, exit-code discipline)

## §H. Open Question for plan-auditor

- **Globs type: `[]string` vs dual-form custom `UnmarshalJSON`.** This plan chooses `[]string` as the minimum mechanism, justified by (a) verified zero existing `Globs` consumers and (b) Claude Code v2.1.69+ emitting the array form. If plan-auditor judges cross-version resilience (older CC emitting a bare string) to be a hard requirement rather than a speculative future, the requirement should be re-tightened in spec.md §C to mandate dual-form acceptance — otherwise `[]string` stands.
