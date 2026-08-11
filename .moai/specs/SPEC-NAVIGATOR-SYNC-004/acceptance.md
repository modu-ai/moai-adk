# acceptance.md — SPEC-NAVIGATOR-SYNC-004 (BAS Epic M2 — Falconer Route)

> Acceptance criteria for the M2 Route layer. Each AC is a binary-testable Given-When-Then,
> traceable to a spec.md §C REQ. The coverage AC (AC-NS4-010) names the exact command + ratio formula
> (verification-claim-integrity §1.1 surface 3 + §2).

## §A. Severity Legend

- **MUST** — failure blocks merge (M2 cannot ship without this passing).
- **SHOULD** — failure is a debt item, mergeable with a recorded `[NEEDS CLARIFICATION]` or follow-up.
- **MAY** — quality bar, not a gate.

## §B. Traceability Matrix

| REQ | AC(s) | Severity | Milestone |
|---|---|---|---|
| REQ-NS4-001 (on-demand CLI, no PostToolUse) | AC-NS4-001a, AC-NS4-001b | MUST | M2.4 |
| REQ-NS4-002 (three read-only inputs) | AC-NS4-002 (sub-assertions a/b/c) | MUST | M2.1 |
| REQ-NS4-003 (work-item promotion) | AC-NS4-003 (sub-assertions a/b/c) | MUST | M2.1 |
| REQ-NS4-004 (owner = code path, three paths) | AC-NS4-004 (table-driven: 3 source_kinds × confidence) | MUST | M2.2 |
| REQ-NS4-005 (consumer-only) | AC-NS4-005a, AC-NS4-005b | MUST | M2.5 |
| REQ-NS4-006 (non-overlap) | AC-NS4-006 | MUST | M2.5 |
| REQ-NS4-007 (independent output artifact) | AC-NS4-007 | MUST | M2.3 |
| REQ-NS4-008 (atomic-rename + idempotence + provenance) | AC-NS4-008a, AC-NS4-008b | MUST | M2.3 |
| REQ-NS4-009 (fail-open on every error mode) | AC-NS4-009 (table-driven: 7 modes) | MUST | M2.4 |
| REQ-NS4-010 (≥70% accuracy, measured) | AC-NS4-010 | MUST | M2.5 |
| REQ-NS4-011 (Hidden subcommand) | AC-NS4-011 | MUST | M2.4 |
| REQ-NS4-012 (template-first) | AC-NS4-012 | MUST | M2.5 |

## §C. Definition of Done

- Every MUST AC in §D shows PASS with a cited command + observed output (no narrative "passes").
- `go test ./internal/navigator/route/ -count=1` exits 0.
- `go test ./internal/navigator/route/ -run TestRouteAccuracy -v` prints an observed accuracy percentage ≥ 70.0.
- `go test ./internal/navigator/route/ -run TestRouteNonOverlap -count=1` exits 0.
- `go test ./internal/cli/ -run TestNavigatorRoute -count=1` exits 0 (Hidden-subcommand assertion).
- `moai spec lint --strict .moai/specs/SPEC-NAVIGATOR-SYNC-004/` reports 0 errors (the repo-wide baseline of pre-existing findings is NOT attributed to this SPEC).
- An idempotence test (two consecutive `navigator-route` runs on the same inputs) produces byte-identical `work-items.json` (AC-NS4-008b).
- No new `moai hook navigator-route` subcommand exists; the Route layer is invoked from `/moai project`, not from a PostToolUse hook (AC-NS4-001 sub-assertion a).
- `internal/navigator/sync/`, `internal/navigator/detect/`, `internal/navigator/tiers/`, `internal/hook/navigator_detect*.go`, and `internal/mx/` are byte-unchanged by the M2 run-phase diff (AC-NS4-005a).

## §D. Acceptance Criteria (Given-When-Then)

### AC-NS4-001a — on-demand CLI invocation produces output (REQ-NS4-001, MUST, M2.4)

**Given** a fixture corpus at `internal/navigator/route/testdata/route-corpus/` containing a non-empty `audit-report.json` + a detect-state directory with ≥1 JSONL row + a `nav-graph.json`,
**When** the `navigator-route` subcommand is invoked with the corpus as the project root,
**Then** the Route layer produces `.moai/project/navigator/work-items.{md,json}` with a non-empty `work_items[]` array, the process exits 0, and no user-facing error is surfaced.

### AC-NS4-001b — no PostToolUse real-time path (REQ-NS4-001, MUST, M2.4)

**Given** the M2 run-phase diff,
**When** the integration checks run,
**Then** BOTH sub-assertions hold:

- **(a) no PostToolUse hook branch**: `grep -rn 'navigator-route\|navigator_route\|runNavigatorRoute' internal/hook/` returns zero matches — the Route layer registers NO PostToolUse branch, NO `handle-navigator-route.sh` wrapper exists (`ls .claude/hooks/moai/handle-navigator-route.sh 2>/dev/null` shows no file), and the only entry point is the `navigator-route` Hidden cobra subcommand.
- **(b) no forked hook chain**: `grep -rn 'navigator-route' internal/cli/` shows the subcommand registration is a Hidden cobra command (mirroring `navigator-sync`/`navigator-tiers`), NOT a hook handler.

### AC-NS4-002 — three read-only inputs consumed (REQ-NS4-002, MUST, M2.1)

**Given** a fixture corpus with all three inputs present (audit-report.json + detect JSONL + nav-graph.json),
**When** the Route layer runs,
**Then** ALL THREE sub-assertions hold:

- **(a) audit-report.json consumed**: the `work_items[]` array contains entries with `source_kind: "audit-missing"` AND `source_kind: "audit-orphan"`, their `source_entry` fields matching the fixture's `missing[]` and `orphan[]` elements verbatim.
- **(b) detect JSONL consumed across sessions**: the `work_items[]` array contains entries with `source_kind: "detect"`, one per deduplicated `changed_path` across ALL `*.jsonl` files in the detect-state fixture directory (not just one session file). A path edited in two session files produces ONE detect work item (latest `changed_at` wins).
- **(c) nav-graph.json consumed for owner resolution**: at least one `audit-missing` work item whose fixture design doc carries an `@NAV:SYM` token has `owner_path` resolved to the symbol's declaration package (looked up via the `nav-graph.json` fixture), with `confidence: "medium"` — proving the graph was consulted, not just the audit fields.

### AC-NS4-003 — work-item promotion (5 fields + dedup) (REQ-NS4-003, MUST, M2.1)

**Given** a non-empty input set,
**When** the Route layer completes,
**Then** ALL THREE sub-assertions hold:

- **(a) 5 fields present**: every element of `work_items[]` carries exactly `source_kind` ∈ `{audit-missing, audit-orphan, detect}`, `source_entry` (the original entry verbatim), `owner_path` (non-empty absolute path), `action` (non-empty one-line string), and `confidence` ∈ `{high, medium, low}`. No extra fields, no missing fields.
- **(b) deduplication stable**: the same input set promoted twice (within one run or across two runs on the same HEAD) produces the same `work_items[]` array, sorted by `(source_kind, owner_path, source_entry.identifier)`. The same finding appearing in both audit and detect does NOT produce two work items if the `source_entry.identifier` collides — the audit source wins (audit is the authoritative roll-up; detect is the real-time supplement).
- **(c) action directive is non-generic**: the `action` string names the closing action per `source_kind` (orphan → "link this SPEC to a design feature or document its design rationale"; missing → "create a SPEC for this design feature or link existing code"; detect → "verify the affected doc rows still hold after this edit"). A generic "fix this" or empty string fails the assertion.

### AC-NS4-004 — owner = code/doc path, three resolution paths × confidence (REQ-NS4-004, MUST, M2.2)

**Given** a fixture corpus covering all three `source_kind`s and all three confidence levels,
**When** the owner-resolution logic runs,
**Then** for every row in the table below, the emitted work item's `owner_path` + `confidence` match the expected values, AND `owner_path` is an absolute path (never a person, team, or email).

| case | source_kind | fixture input | expected owner_path | expected confidence |
|---|---|---|---|---|
| 004a orphan-direct | audit-orphan | `implementation_path: "internal/foo.go"` (non-empty) | `/abs/…/internal/foo.go` | high |
| 004b orphan-empty-impl | audit-orphan | `implementation_path: ""` (empty) | `/abs/…/.moai/specs/<spec-id>/` (SPEC-dir fallback) | low |
| 004c missing-via-symbol | audit-missing | design doc with `@NAV:SYM:auth.ParseBearer` token, symbol resolves to `internal/auth/login.go` via graph | `/abs/…/internal/auth/login.go` (or its package dir) | medium |
| 004d missing-doc-fallback | audit-missing | design doc with NO `@NAV:SYM` token (or B5 heuristic unreliable) | `/abs/…/.moai/project/tech.md` (the design-doc `source.file`) | low |
| 004e detect-direct | detect | `changed_path: "/abs/…/internal/auth/login.go"` | `/abs/…/internal/auth/login.go` | high |

**Owner-is-path invariant**: `grep -E '@|mailto:|users/' <(jq -r '.work_items[].owner_path' work-items.json)` returns zero matches across ALL test cases — no owner_path contains an email, a username, or a person reference. (The `users/` pattern catches a `CODEOWNERS`-style path that is actually a person anchor.)

### AC-NS4-005a — consumer-only: M0/M1/M4 + mx byte-unchanged (REQ-NS4-005, MUST, M2.5)

**Given** the M2 run-phase diff is applied,
**When** `git diff --name-only origin/main...HEAD` is run,
**Then** the output does NOT contain any path under `internal/navigator/sync/`, `internal/navigator/detect/`, `internal/navigator/tiers/`, `internal/hook/navigator_detect*.go`, or `internal/mx/` (M2 consumes these, does not modify them). Command: `git diff --name-only origin/main...HEAD | grep -E '^internal/(navigator/(sync|detect|tiers)|hook/navigator_detect|mx)/' ; [ $? -eq 1 ]` (grep exit 1 = no matches = PASS).

### AC-NS4-005b — consumer-only: read is via public API + emitted JSON (REQ-NS4-005, MUST, M2.5)

**Given** the M2 source files,
**When** `grep -rn 'os.WriteFile\|os.Rename' internal/navigator/route/*.go` is run,
**Then** no write/rename call targets `internal/navigator/sync/`, `internal/navigator/detect/`, `internal/navigator/tiers/`, `internal/hook/`, or `internal/mx/` paths. (The only writes are `work-items.{md,json}` under `.moai/project/navigator/` and the log under `.moai/logs/`.) AND the Route layer reads audit-report.json + nav-graph.json via `os.ReadFile` (not by executing `navigator-audit.sh` or `navigator-sync`).

### AC-NS4-006 — non-overlap with predecessor chains + M0/M1/M4 (REQ-NS4-006, MUST, M2.5)

**Given** the M2 production source files (`internal/navigator/route/*.go` excluding `*_test.go`) + `internal/cli/navigator_route.go`,
**When** `grep -rn 'capability-map\.md\|audit-report\.\(md\|json\)\|capability-symbols\.\(md\|json\)\|nav-graph\.json\|tiers\.json\|blueprint/\|decisions/' internal/navigator/route/ internal/cli/navigator_route.go` is run,
**Then** the ONLY matches for `audit-report.json`, `nav-graph.json`, and the detect JSONL path are in a READ context (`os.ReadFile`, `os.Open`, `filepath.Glob`), and there are ZERO matches for `capability-map.md`, `capability-symbols.{md,json}`, `tiers.json`, `blueprint/`, or `decisions/` as write targets. A new `internal/navigator/route/nonoverlap_test.go` encodes this assertion and carries forward the pattern from `internal/navigator/sync/nonoverlap_test.go` + `internal/hook/navigator_detect_nonoverlap_test.go`.

### AC-NS4-007 — independent output artifact (REQ-NS4-007, MUST, M2.3)

**Given** a non-empty input set,
**When** the Route layer completes,
**Then** BOTH files exist at `.moai/project/navigator/work-items.md` AND `.moai/project/navigator/work-items.json`, the `.json` file parses as valid JSON with top-level keys `provenance` + `work_items` (array), the `.md` file renders the same work-item set grouped by `source_kind`, and the two files are consistent (the count of work items in the `.json` array equals the count of rows in the `.md` tables, per `source_kind`). No file other than these two (+ their `.tmp` transients + the log line) is written.

### AC-NS4-008a — atomic-rename + provenance, no wall-clock (REQ-NS4-008, MUST, M2.3)

**Given** a successful Route layer run,
**When** `work-items.json` is inspected,
**Then** the `provenance` block carries `route_commit_sha` (matches `git rev-parse HEAD` at run time) and `captured_at` (matches `git log -1 --format=%cI` of that SHA), and there is NO `time.Now()`-derived field anywhere in the artifact. A grep test asserts no wall-clock field name appears: `grep -E '"(generated_at|created_at|timestamp)":' work-items.json` returns zero matches (the canonical field is `captured_at`, sourced from git, not from the system clock).

### AC-NS4-008b — idempotence: byte-identical re-run (REQ-NS4-008, MUST, M2.3)

**Given** the same input set (unchanged audit-report.json + detect state + nav-graph.json) at the same HEAD,
**When** `navigator-route` is invoked twice in succession,
**Then** the two `work-items.json` outputs are byte-identical (`diff <run1> <run2>` returns empty, OR `cmp -s` exits 0). The same holds for `work-items.md`. (The atomic-rename `.tmp` + rename is a no-op write when the content is unchanged.)

### AC-NS4-009 — fail-open across 8 error modes (REQ-NS4-009, MUST, M2.4)

**Given** any of the 8 failure modes in the table below,
**When** `navigator-route` is invoked,
**Then** for every row the Route layer produces the expected `exit-0` behavior and the expected `empty-or-partial` output. In ALL 8 cases: exit 0, no user-facing error, no cascade into sibling `/moai project` steps.

| case | trigger | expected exit-0 | expected empty-or-partial output |
|---|---|---|---|
| 009a audit absent | `audit-report.json` does NOT exist | exit 0, promote from detect + nav-graph only (partial) | work-items from detect only (or empty if detect also absent); one log line |
| 009b detect state absent/empty | `.moai/state/navigator-detect/` does not exist OR contains no `*.jsonl` | exit 0, promote from audit only (partial) | work-items from audit only; one log line |
| 009c nav-graph absent | `nav-graph.json` does NOT exist | exit 0, owner resolution degrades (all `audit-missing` → low-confidence doc fallback) | work-items with degraded owners (missing → low); one log line |
| 009d unparseable JSON | one input (audit OR nav-graph OR a detect JSONL line) is invalid JSON | exit 0, skip the malformed input/line, continue with well-formed inputs | partial work-item set (well-formed inputs promoted); one log line per malformed input/line |
| 009e schema-invalid | `audit-report.json` is valid JSON but missing the `orphan[]` array (or `nav-graph.json` missing `nodes[]`) | exit 0, degrade the affected source (audit missing → no audit-missing/orphan work items; nav-graph → no symbol resolution) | partial work-item set; one log line |
| 009f owner-resolution error | a symbol node in the graph references a `source_path` that does not exist on disk | exit 0, mark the work item `confidence: low`, owner_path = the fallback path, continue | work-item with degraded owner (not dropped); one log line |
| 009g all-inputs-absent | ALL three inputs absent (fresh checkout, no audit, no detect, no graph) | exit 0, write NO output | no `work-items.{md,json}` written; one summary log line |
| 009h timeout | Route layer configured with a 500ms context timeout and an input set large enough to exceed it (or a test that cancels the context mid-promotion) | exit 0, return partial work-item set collected so far (possibly empty) | partial or empty work-item set; NO log line (context cancellation is not an error to advertise) |

### AC-NS4-010 — ≥70% Route accuracy, mechanically measured (REQ-NS4-010, MUST, M2.5)

**Given** the fixture corpus at `internal/navigator/route/testdata/route-corpus/` (6 `missing[]` entries: 3 with a resolvable `@NAV:SYM` token → medium, 3 without → low; 12 `orphan[]` entries: 10 with non-empty `implementation_path` → high, 2 with empty → low; 12 detect rows across 2 session files: all → high; corpus total = 30) and the synthetic `nav-graph.json` fixture covering the symbol nodes referenced by the `missing` corpus,

**Dual-arithmetic (D1 fix — the floor survives the B5 fallback)**:
- Happy path: actionable = 3 medium + 10 high-orphan + 12 high-detect = 25; accuracy = 25/30 = **83.3%** ≥ 70% ✓.
- B5 fallback (3 medium-missing → low): actionable = 0 + 10 + 12 = 22; accuracy = 22/30 = **73.3%** ≥ 70% ✓.
**When** the command `go test ./internal/navigator/route/ -run TestRouteAccuracy -v` is run,
**Then** the test emits an observed accuracy percentage (the ratio `(actionable work items) / (total input findings)`, where actionable = `confidence` ∈ `{high, medium}` and total = all `missing` + `orphan` + deduplicated `detect` findings) and the observed percentage is `>= 70.0`. The test FAILS (non-zero exit) if the observed percentage is `< 70.0` and prints the observed value on failure. The percentage printed on the PASS line is the Evidence; the ≥70% assertion is the Claim.

**Attribution (verification-claim-integrity §2)**: the accuracy number is attributable to the exact `go test … -v` invocation + its stdout; it is NOT a carried-over estimate, NOT a sampled figure, NOT a narrative "high accuracy". A subsequent run on the same fixture corpus produces the same percentage (deterministic).

### AC-NS4-011 — Hidden cobra subcommand (REQ-NS4-011, MUST, M2.4)

**Given** the M2 run-phase diff,
**When** the CLI integration checks run,
**Then** BOTH sub-assertions hold:

- **(a) navigator-route is Hidden**: `go test ./internal/cli/ -run TestNavigatorRouteHidden -v` asserts the `navigator-route` cobra command has `Hidden: true`, mirroring `internal/cli/navigator_tiers_test.go:65-70` (which asserts the same for `navigator-tiers`).
- **(b) no top-level moai subcommand**: `grep -rn 'navigator-route' internal/cli/root.go` (or the root-command registration site) returns zero matches — `navigator-route` is registered as a subcommand under the project step sequence, NOT as a top-level `moai navigator-route` user-facing command.

### AC-NS4-012 — template-first (REQ-NS4-012, MUST, M2.5)

**Given** the M2 run-phase diff touches any template-managed file under `.claude/`,
**When** `git diff --name-only origin/main...HEAD | grep '^internal/template/templates/'` is run,
**Then** every template-managed path modified locally ALSO appears in `internal/template/templates/` (the template source is never bypassed). AND if `internal/template/templates/.claude/settings.json` (or any template file) changed, `make build` regenerated `catalog.yaml` and the regenerated catalog is committed in the same PR. (If M2 ships NO distributed surface — the expected case, pure CLI + runtime Go — this AC reduces to: no template path in the diff, no catalog regen required — document this in the PR body.)

## §E. Indirect Verification (cross-cutting)

- **LSP / lint gate**: `go vet ./internal/navigator/route/... ./internal/cli/...` and `golangci-lint run ./internal/navigator/route/... ./internal/cli/...` exit 0.
- **Race detector**: `go test -race ./internal/navigator/route/` exits 0 (the promotion engine + the atomic-write are concurrency-safe; the multi-session JSONL read is the race-relevant surface).
- **Subagent boundary** (`internal/hook/CLAUDE.md` C-HRA-008 family): `grep -rn 'AskUserQuestion\|mcp__askuser' internal/navigator/route/*.go internal/cli/navigator_route.go` returns zero matches — the Route layer runs in subagent/CLI context and never prompts.

## §F. Edge Cases (quality bar, not merge gates)

- **Empty audit (all matched, no missing/orphan)**: `audit-report.json` with `missing: []`, `orphan: []`, `matched: [N]` → Route layer emits work-items from detect only (or empty if detect also absent). Not an error.
- **Detect JSONL with malformed lines interspersed**: per-line fail-open (like M1's per-edge fail-open) — well-formed lines promote, malformed lines log + skip.
- **Symbol node in graph references a non-existent path**: owner-resolution marks the work item `confidence: low`, owner_path = the fallback (design-doc path for missing, SPEC-dir for orphan). Not fatal.
- **Same path appears in both audit-orphan and detect**: deduplication by `(source_kind, owner_path, source_entry.identifier)` — the two source_kinds are distinct, so both work items are emitted (the orphan says "this SPEC has no design feature"; the detect says "this code was just edited"). They are different findings even if they share an owner_path.
- **Very large audit (200+ findings) + large detect state (500+ rows)**: the 500ms p99 budget (§D Constraints) must hold. The promotion engine is O(n) in findings; the owner-resolution is O(n) for direct paths + O(n·k) for symbol lookups (k = symbol edges in the graph, bounded). No quadratic blow-up.
- **Detect JSONL spanning many sessions (10+ files)**: the multi-session read + dedup-by-changed_path must be correct. A path edited in 10 sessions produces ONE detect work item (latest `changed_at`).
- **Provenance inputs block records the right commits**: the `provenance.inputs.audit_commit` matches `audit-report.json`'s `audit_commit`; `provenance.inputs.nav_graph_commit` matches `nav-graph.json`'s `provenance.extract_commit_sha`; `provenance.inputs.detect_sessions` lists the session IDs whose JSONL was consumed.

## §G. Forward-Looking Checks (advisory, not gates)

- **M3 readiness**: the `work-items.json` schema (AC-NS4-003 sub-assertion a + AC-NS4-007) is the contract M3 Fix consumes. Any M2→M3 schema drift should be caught by an M3 plan-phase audit, not an M2 AC. The `action` field is the hint M3 uses to decide the fix strategy (e.g. "create a SPEC" → M3 drafts a SPEC stub; "verify the affected rows" → M3 re-runs the bound doc through `--compare-to`).
- **PostToolUse real-time path future**: if a debounce/cadence policy is later desired (e.g. "promote only on session end", or "promote only when ≥N detections accumulate for one path"), a follow-up SPEC may add a PostToolUse or SessionEnd hook that calls the same `route.Promote` engine. The engine is pure (M2.1) so the real-time path would be a thin wrapper. This is explicitly deferred (§F Out of Scope in spec.md) — the on-demand CLI is the M2 deliverable.
- **External-tracker integration future**: a future SPEC MAY consume `work-items.json` read-only to create GitHub issues / Linear tickets. The `owner_path` + `action` + `confidence` fields are designed to be tracker-friendly (a tracker maps `owner_path` to a CODEOWNERS team, `action` to a ticket body, `confidence` to a priority). M2 does NOT implement the tracker write — it emits the local artifact that a future tracker integration consumes.
- **B5 heading-range heuristic resolution**: the run-phase recon (plan.md §B B5) decides whether the missing→symbol lookup is a MUST or a SHOULD. If the heading→line-range mapping proves unreliable, the lookup downgrades to a whole-doc fallback and the accuracy metric's medium-confidence count drops. Under the rebalanced corpus (§E / AC-NS4-010), the B5 fallback yields accuracy = 22/30 = **73.3%** ≥ 70% — the floor survives the fallback per the explicit dual-arithmetic, with 3.3pp headroom. The test will still surface the drop (83.3% → 73.3%) as a visible regression signal, but it will not fail the ≥70% gate.
