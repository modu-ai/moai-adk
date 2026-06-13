# Implementation Plan — SPEC-HARNESS-LOOP-CLOSURE-001

**Tier**: S (minimal) · **cycle_type**: tdd (RED-GREEN-REFACTOR) · single-pass (no Round split)

## A. Context

Prove the harness apply/lineage/rollback mechanism end-to-end. The pipeline is fully
wired (Observe → … → Apply / Rollback) but has NEVER closed its loop: zero applies,
no snapshots dir, no per-transition manifest. This SPEC ADDS a per-transition lineage
manifest (M6) and ensures the existing apply/snapshot/rollback Go path is proven by
tests so ONE clean human-gated apply can flow end-to-end.

Ground truth (verified by file reads — see §H for exact anchors):
- `Apply()` (applier.go:176) already: (1) calls `evaluator.Evaluate()`, (2) returns
  `ApplyPendingError` on `DecisionPendingApproval`, (3) returns a rejection error on
  `DecisionRejected`, (4) creates a snapshot then modifies the SKILL.md on
  `DecisionApproved`. The ONLY gap is: no lineage entry is written on accept OR reject.
- `createSnapshot` (applier.go:218) + `RestoreSnapshot` (applier.go:272) are complete
  and correct — snapshot byte-backup + manifest.json before write; restore reads
  manifest + restores. NO change to snapshot logic is needed.
- L1 Frozen Guard (`frozen_guard.go:35 IsFrozen`, hardcoded prefixes 18-23) already
  rejects non-`my-harness-*` targets. NO change to the guard is needed.
- `auto_apply: false` (harness.yaml:116) → the L5 human gate is already the default.
- `WritePromotion` (learner.go:145) is the canonical append-JSONL idiom to mirror for
  the new lineage writer (auto-mkdir parent, marshal, `O_APPEND|O_CREATE|O_WRONLY`).
- No `my-harness-*` skill currently exists in the live tree — the Go path is proven via
  `t.TempDir()` fixtures; authoring a live skill is an OPERATIONAL follow-through (§Exclusions).

## B. Known Issues / Risks

1. **Lineage write must not abort the apply on its own failure** — a manifest write
   failure should NOT silently corrupt the apply. Decision (§D.2): lineage write is
   best-effort-after-effect on the accept path (after the SKILL.md write succeeds) and
   after-effect on the reject path; a lineage write error is wrapped + returned but the
   primary transition (the file write, or the rejection) has already happened. This
   mirrors `LogViolation`'s non-blocking posture (frozen_guard.go:79) but the Apply
   contract still surfaces the error to the caller. The run phase confirms the exact
   ordering against the existing `Apply` switch.
2. **Reject-path placement** — the current `Apply()` returns early on `DecisionRejected`
   (applier.go:185-186) BEFORE any file work. The lineage write for the reject case must
   be inserted at that early-return site, writing `decision:"rejected"` with the layer
   reason, then returning the (preserved) rejection error.
3. **Pending path must NOT write lineage** — `DecisionPendingApproval` returns
   `ApplyPendingError` (applier.go:188-191). Pending is not a transition; do NOT write a
   lineage entry there (REQ-HLC-005). Only the eventual approved/rejected resolution writes.
4. **snapshotBase vs manifest path** — the lineage manifest lives at
   `.moai/harness/learning-history/manifest.jsonl` (a FIXED relative path under the
   learning-history dir), distinct from the per-apply snapshot dirs under
   `snapshotBase`. The writer must accept the manifest path (or its dir) as a parameter
   so tests can point it at `t.TempDir()` — do NOT hardcode the absolute repo path in
   the writer (template neutrality + testability).
5. **Backward compat** — `manifest.jsonl` is a brand-new file. Do NOT touch
   `usage-log.jsonl` or `tier-promotions.jsonl`. The `LineageEntry` struct uses
   `omitempty` so future fields and existing reads stay compatible (REQ-HLC-002/011).
6. **C-HRA-008 boundary** — no `AskUserQuestion` anywhere in `internal/harness/`. The new
   `lineage.go` is pure file I/O; the human gate stays at the orchestrator. Keep the
   subagent_boundary_test.go guard green (REQ-HLC-012).
7. **Template neutrality** — all new code is in `internal/harness/`; nothing in
   `internal/template/templates/**`; no SPEC ID / date / SHA in shipped code.

## C. Pre-flight Checks

```bash
# 1. Branch + baseline
git branch --show-current
git rev-parse HEAD

# 2. Build green + cross-platform (lineage.go is pure os/json — must build on windows too)
go build ./internal/harness/...
GOOS=windows GOARCH=amd64 go build ./internal/harness/...

# 3. Existing apply/snapshot tests green (baseline before adding lineage)
go test ./internal/harness/ -run 'TestApply|TestRestoreSnapshot|TestSnapshot' -count=1

# 4. C-HRA-008 boundary baseline (must already be green)
go test ./internal/harness/ -run TestSubagentBoundary -count=1
grep -rn 'AskUserQuestion\|mcp__askuser' internal/harness/ | grep -v "_test.go" | grep -v "^[^:]*:[0-9]*:[[:space:]]*//" || echo "BOUNDARY_CLEAN"

# 5. Confirm no lineage symbols pre-exist (this SPEC introduces them)
grep -rn 'LineageEntry\|WriteLineageEntry\|LoadManifest\|manifest.jsonl' internal/harness/ || echo "NO_PREEXISTING_LINEAGE"

# 6. Confirm the two protected runtime artifacts (do not touch)
ls -l .moai/harness/usage-log.jsonl .moai/harness/learning-history/tier-promotions.jsonl
```

## D. Design Decisions

### D.1 LineageEntry struct (types.go, additive — REQ-HLC-002)

Add to `internal/harness/types.go` a struct mirroring the `Promotion` JSONL style:

```go
// LineageEntry is one append-only transition record in manifest.jsonl (M6 auditable lineage).
type LineageEntry struct {
    ProposalID     string    `json:"proposal_id"`
    TargetPath     string    `json:"target_path"`
    AppliedSurface string    `json:"applied_surface,omitempty"` // frontmatter field key on accept; empty on reject
    Decision       string    `json:"decision"`                  // "approved" | "rejected"
    Timestamp      time.Time `json:"timestamp"`
    Reason         string    `json:"reason,omitempty"`
}
```

`Decision` is a plain string with two documented values (`"approved"` / `"rejected"`) to
keep the JSONL self-describing and avoid a new enum type (Tier S simplicity). `omitempty`
on `AppliedSurface`/`Reason` keeps reject entries (no surface) and minimal entries compact.

### D.2 lineage.go (NEW — REQ-HLC-001/002, mirror learner.go:145)

`internal/harness/lineage.go` exposes two functions:

```go
// WriteLineageEntry appends one entry to manifestPath (auto-creates parent dir).
func WriteLineageEntry(manifestPath string, entry LineageEntry) error

// LoadManifest reads all entries from manifestPath (returns empty slice if absent).
func LoadManifest(manifestPath string) ([]LineageEntry, error)
```

- `WriteLineageEntry` mirrors `WritePromotion` (learner.go:145-176): `MkdirAll` parent,
  default `Timestamp` to `time.Now().UTC()` when zero, `json.Marshal`, append `'\n'`,
  open `O_APPEND|O_CREATE|O_WRONLY` 0o644, write.
- `LoadManifest` reads the file line-by-line, `json.Unmarshal` per line, skips blank
  lines; a missing file returns `([]LineageEntry{}, nil)` (not an error — backward compat).
- The manifest path is a PARAMETER (not hardcoded) so tests use `t.TempDir()`. The
  production caller passes `<learning-history-dir>/manifest.jsonl`.

### D.3 Apply() integration (applier.go:176 — REQ-HLC-003/004/005)

Extend `Apply()` minimally. It gains a manifest path so it can write lineage. Two insertion
points:

- **Reject path** (current applier.go:185-186, `case DecisionRejected`): BEFORE returning
  the rejection error, append a `LineageEntry{Decision:"rejected", Reason: <layer reason>,
  ProposalID, TargetPath}` (no `AppliedSurface`). Then return the preserved rejection error.
  Active harness left unchanged (no snapshot, no SKILL.md write) — REQ-HLC-004.
- **Accept path** (current applier.go:193-213, `case DecisionApproved` → snapshot → file
  modify): AFTER the SKILL.md modification succeeds, append a `LineageEntry{Decision:
  "approved", AppliedSurface: proposal.FieldKey, ProposalID, TargetPath, Reason: <approved>}`.
- **Pending path** (applier.go:188-191): UNCHANGED — return `ApplyPendingError`, write NO
  lineage (REQ-HLC-005).

Signature decision (run-phase confirms exact shape): the manifest path is threaded through
`Apply` (preferred — explicit + testable), e.g. an added parameter or an Applier field set
at construction. The run phase picks the least-invasive form that keeps every existing
`Apply` caller compiling (or updates the single call-site if a param is added). The
constraint is: the manifest path MUST be injectable for `t.TempDir()` tests.

### D.4 Snapshot/rollback (NO code change — REQ-HLC-009/010)

`createSnapshot` already creates the snapshots base dir (`MkdirAll`, applier.go:224) on
first apply and writes a byte-backup + manifest.json before the SKILL.md write.
`RestoreSnapshot` already restores byte-identically. This SPEC adds NO snapshot code; it
adds TESTS proving the snapshot dir is created on first apply and that rollback restores
byte-identical content (AC-HLC-006/007).

### D.5 lineage_test.go (NEW — RED first)

Table-driven + scenario tests covering: write→load round-trip, append-only (two writes →
two entries), missing-file LoadManifest returns empty, approved-entry shape, rejected-entry
shape, snapshot-created-on-first-apply, rollback byte-identical. All fixtures under
`t.TempDir()`; a fixture `my-harness-x/SKILL.md` is hand-written in the temp dir for the
accept path, and a fixture frozen-zone target (e.g. `.claude/agents/moai/x.md` relative)
for the L1 reject path.

## E. Self-Verification (run-phase exit gate)

```bash
# 1. Full harness package suite (cascading regressions)
go test ./internal/harness/... -count=1

# 2. Coverage on the new lineage path (>=85% target)
go test -cover ./internal/harness/

# 3. Cross-platform build (lineage.go is pure os/json)
go build ./internal/harness/...
GOOS=windows GOARCH=amd64 go build ./internal/harness/...

# 4. C-HRA-008 subagent boundary (MUST stay green)
go test ./internal/harness/ -run TestSubagentBoundary -count=1
grep -rn 'AskUserQuestion\|mcp__askuser' internal/harness/ | grep -v "_test.go" | grep -v "^[^:]*:[0-9]*:[[:space:]]*//" || echo "BOUNDARY_CLEAN"

# 5. Backward compat — protected runtime artifacts unchanged in the working tree
git status --porcelain .moai/harness/usage-log.jsonl .moai/harness/learning-history/tier-promotions.jsonl
# Expected: empty (no diff to either protected artifact)

# 6. Lint
golangci-lint run --timeout=2m ./internal/harness/...

# 7. Template neutrality — no leak of harness lineage code into templates
grep -rn 'LineageEntry\|WriteLineageEntry\|manifest.jsonl' internal/template/templates/ || echo "TEMPLATE_CLEAN"
```

## F. Milestones

| Milestone | Description | Files | Verify |
|-----------|-------------|-------|--------|
| **M1 (RED)** | Add failing `lineage_test.go`: write/load round-trip, append-only, missing-file→empty, approved-entry shape, rejected-entry shape, snapshot-created-on-first-apply, rollback-byte-identical. Reference the not-yet-existing `LineageEntry` / `WriteLineageEntry` / `LoadManifest` so the package fails to compile (RED). | `internal/harness/lineage_test.go` (NEW) | `go test ./internal/harness/ -run TestLineage` fails to compile / RED |
| **M2 (GREEN — types + writer)** | Add `LineageEntry` to `types.go` (additive, omitempty). Add `lineage.go` with `WriteLineageEntry` + `LoadManifest`, mirroring `learner.go:145`. | `internal/harness/types.go`, `internal/harness/lineage.go` (NEW) | write/load/append-only/missing-file tests GREEN |
| **M3 (GREEN — Apply integration)** | Thread an injectable manifest path into `Apply()`. Append `decision:"approved"` after the SKILL.md modify; append `decision:"rejected"` at the `DecisionRejected` early-return; leave the pending path unchanged. Update the single `Apply` call-site if a param is added. | `internal/harness/applier.go` (+ caller if needed) | approved/rejected lineage tests GREEN; snapshot-on-first-apply + rollback-byte-identical tests GREEN |
| **M4 (REFACTOR + verify)** | Run §E self-verification gate in full. Confirm full suite GREEN, coverage ≥85% on lineage path, cross-platform build, C-HRA-008 green, protected artifacts unchanged, lint clean, template clean. No snapshot logic change, no safety-pipeline change, no FROZEN-constant change. | (verification only) | All §E commands pass |

> Tier S: no design.md / research.md. Single-pass M1-M4 (no Round split — task count well
> under the SSE-stall threshold). The ordering M2→M3 is sequential (Apply integration
> depends on the writer existing).

## G. Anti-Patterns to Avoid

- **Writing lineage on the pending path** → pending is not a transition (REQ-HLC-005).
  Only approved/rejected resolutions write.
- **Letting a lineage write failure corrupt or hide the primary transition** → the file
  write (accept) or the rejection (reject) is the primary effect; lineage is the
  after-effect record. Surface the error but do not undo the transition (§B.1 / §D.2).
- **Hardcoding the manifest absolute path inside the writer** → breaks `t.TempDir()` tests
  and template neutrality. The path is a parameter.
- **Modifying `createSnapshot` / `RestoreSnapshot` / the safety pipeline / L1 guard** →
  all complete and correct; this SPEC only ADDS lineage + tests (Exclusions / C6).
- **Flipping `auto_apply` to `true`** → out of scope; the human gate is the point.
- **Touching `usage-log.jsonl` / `tier-promotions.jsonl`** → protected runtime artifacts
  (REQ-HLC-011). `manifest.jsonl` is the only new file.
- **Adding a quality verdict / scorer gate** → that is the deferred M2-lite SPEC (C7).
- **Wiring `AskUserQuestion` into the harness** → C-HRA-008 violation (REQ-HLC-012). The
  gate stays at the orchestrator via the existing `OversightProposal`/`ApplyPendingError`.
- **Leaking lineage code into `internal/template/templates/**`** → template neutrality (C4).

## H. Cross-References (exact ground-truth anchors)

- `internal/harness/applier.go` — `Apply`:176 (switch: `DecisionRejected`:185-186 [reject
  lineage insertion], `DecisionPendingApproval`:188-191 [unchanged], `DecisionApproved`:193
  → `createSnapshot`:199 → file modify:203-213 [approved lineage insertion]);
  `createSnapshot`:218 (MkdirAll snapshot base:224 — first-apply dir creation);
  `RestoreSnapshot`:272; `ApplyPendingError`:133; `snapshotManifest`:146 (DO NOT MODIFY
  snapshot/restore logic).
- `internal/harness/types.go` — `Promotion`:191 (JSONL style reference); ADD `LineageEntry`
  after the Phase-3 types block.
- `internal/harness/learner.go` — `WritePromotion`:145-176 (canonical append-JSONL idiom to
  mirror; DO NOT MODIFY).
- `internal/harness/safety/pipeline.go` — `Evaluate`:89 (`Decision` producer:
  `DecisionRejected` w/ `RejectedBy`+`Reason`:92-99/107-113/117-126/133-142;
  `DecisionPendingApproval`:147-153; `DecisionApproved`:155-157; DO NOT MODIFY).
- `internal/harness/safety/frozen_guard.go` — `IsFrozen`:35, `frozenPrefixes`:18-23
  (`.claude/agents/moai/`, `.claude/skills/moai-`, `.claude/rules/moai/`,
  `.moai/project/brand/`; DO NOT MODIFY).
- `internal/harness/subagent_boundary_test.go` — C-HRA-008 binary guard (MUST stay green).
- `.moai/config/sections/harness.yaml`:116 — `auto_apply: false` (stays).
- `.moai/harness/learning-history/` — manifest.jsonl lives here alongside
  tier-promotions.jsonl (production caller path).
- CLAUDE.local.md §15 (language neutrality), §25 (internal-content isolation).
