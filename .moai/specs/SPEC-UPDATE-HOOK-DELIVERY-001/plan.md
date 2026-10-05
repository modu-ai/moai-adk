# SPEC-UPDATE-HOOK-DELIVERY-001 — Implementation Plan

> Tier M. Design-decision card: the core resolution (deliver / detect+guide / no-op+docs) is an OPEN decision owned by the operator at Implementation Kickoff Approval. Milestones are ordered by decision-reversibility — the data-model/identity decisions that most constrain everything downstream lead; mechanical verification work closes the plan.

## §A Context

- **Tree**: worktree `.claude/worktrees/t466`, branch `WT-update-hook-delivery`, plan-phase baseline `d592b0551`.
- **SPEC artifacts**: `.moai/specs/SPEC-UPDATE-HOOK-DELIVERY-001/{spec,plan,acceptance,design,research,progress}.md`
- **Module**: `internal/cli/update` (+ `internal/merge` for the strategy layer, `internal/cli` for the doctor surface).
- **Affected packages (re-verification scope at run-phase)**: `internal/cli/update/...`, `internal/merge/...`, `internal/cli/...` (doctor), `internal/template/...` (if template-side identity markers are chosen).

## §B Known Issues (relevant subset)

- **B5 (CI 3-tier)**: `internal/cli` is a critical package (90%+ coverage target). New merge/detection code lands with tests; the lint baseline on this tree is unknown at plan-phase — measure at M1 pre-flight.
- **B10 (PRESERVE)**: `.claude/hooks/moai/*.sh` wrappers, `.moai/state/`, `.moai/config/` are runtime-managed — never touched by this SPEC's mechanism.
- **Boundary hazard (card t461)**: `installPreCommitHookOptional` lives in the same package tree. Any diff touching that function is scope drift — REQ-UHD-012 / AC-UHD-012 guard it.

## §C Pre-flight (run before M1)

```bash
git branch --show-current && git rev-parse --short HEAD
go build ./... && GOOS=windows GOARCH=amd64 go build ./...
go test ./internal/cli/update/... ./internal/merge/...   # baseline before any change
golangci-lint run --timeout=2m internal/cli/update/... internal/merge/... 2>&1 | tail -5
```

Record baseline outputs in progress.md §E.2 before the first change commit.

## §D Constraints

- **PRESERVE**: today's delivered behavior for template-introduced event keys (REQ-UHD-001); deletion preservation with no resurrection (REQ-UHD-002); user-modified entries (REQ-UHD-003); all non-hook keys byte-identical (REQ-UHD-005).
- **FORBIDDEN**: any modification to `installPreCommitHookOptional` or the `.git/hooks/` write path; any write to settings.json outside the selected option's mechanism; hand-editing the `.sh`/`.sh.tmpl` hook wrapper pairs (template-first rule applies if a template-side marker is chosen — edit `internal/template/templates/` first, then `make build`).
- The operator's option decision MUST be recorded in progress.md before M1 implementation commits.

## §E Self-Verification

E1 AC matrix (13 rows, acceptance.md) · E2 cross-platform build (`GOOS=windows` included above) · E3 coverage for touched packages vs 85%/90% thresholds · E4 n/a (no subagent-boundary surface) · E5 lint delta vs M1 baseline · E6 branch HEAD + push state · E7 blockers · E8 RED evidence — AC-UHD-003's failing command + verbatim RED output and its exit code + fixture + tree SHA, captured at M2 before any GREEN (the AC's adoption gate; per verification-completeness §2).

## §F Milestones (ordered by decision-reversibility)

**M1 — Decision landing + identity/data-model record (Priority High; most change-likely)**
The operator selects the option at the Implementation Kickoff Approval gate. M1 records the verdict in progress.md and, for Option A, fixes the entry-identity scheme among the design.md axes (stable entry-identity key vs tombstone-in-file vs seen-registry sidecar) — the single decision that shapes every later milestone. Deliverable: decision record + chosen scheme's data model written into design.md §G (a short "resolution" sub-section, body edit re-delegated to manager-spec if needed per D-NEW-1).

**M2 — Characterization + RED tests (Priority High)**
Before any behavior change: (a) characterization tests pinning the three preserved behaviors (new event key delivered; deletion preserved without resurrection; user-modified entries kept — REQ-UHD-001/002/003); (b) the RED test for the core gap (AC-UHD-003) — template adds an entry inside a carried event key, asserting the selected option's outcome. RED output captured verbatim for §E.8. Existing 31 template tests in `internal/template/settings_test.go` remain green.

**M3 — Core mechanism implementation (Priority High)**
Implement the selected option against `internal/cli/update/merge/` (+ `internal/merge/` if the strategy layer changes, + `internal/cli/doctor.go` for Option B's check). Option A: the per-entry add-only delivery keyed on M1's identity scheme, wired so base-derivation no longer treats a carried event key's array as an opaque scalar for ADD purposes — without resurrecting tombstoned/absent-identity entries. Option B: the detection comparison (template hook set vs user file) + report surface in doctor/update output. Option C: explicit no-op — documentation plus (if chosen) a one-line update-output notice. All options satisfy REQ-UHD-005/006/011 (JSON validity, idempotence, malformed-input grace).

**M4 — Option-gated surface completion (Priority Medium)**
Only the selected option's remaining surface: Option A — tombstone/seen-registry lifecycle (user re-delete honored across subsequent updates; template removal of an entry retires its identity record) OR Option B — guidance text + exit-status semantics for the doctor check OR Option C — docs-site/README limitation note. (For the unselected options this milestone is N/A and is skipped with a progress.md note.)

**M5 — Verification sweep (Priority Medium; mechanical)**
Full affected-package tests, coverage measurement, cross-platform build, lint delta, idempotence double-run scenario, boundary grep proving `installPreCommitHookOptional` untouched (AC-UHD-012). This is the §E self-verification batch; findings land in progress.md §E.3.

## §G Anti-Patterns (this SPEC specifically)

- Do NOT "fix" the add-blindness by recursing `pruneToShared` into arrays generically — that resurrects user-deleted entries (REQ-UHD-002 violation). Array-awareness must be identity-scoped per M1's scheme.
- Do NOT write settings.json from the doctor check (Option B stays read-only).
- Do NOT let the mechanism touch user-authored hook entries absent from the template (Out of Scope).
