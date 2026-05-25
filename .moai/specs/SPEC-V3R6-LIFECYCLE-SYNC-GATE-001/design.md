---
id: SPEC-V3R6-LIFECYCLE-SYNC-GATE-001
artifact: design
version: "0.1.1"
created: 2026-05-25
updated: 2026-05-25
---

## HISTORY

### v0.1.1 (2026-05-25, manager-spec — iter-2 cross-reference sync)
- No design changes in this iteration; version bumped for consistency with spec.md v0.1.1 + acceptance.md v0.1.1
- D5 trailer convention (`Authored-By-Agent: manager-develop`) cross-referenced in §B.4 (see acceptance.md AC-LSG-004 v0.1.1 wording for binding rule)
- §A.1 component diagram remains accurate for 5-deliverable architecture (D8 framing — 5 deliverables = 5 logical units realized by ~15 source files)

### v0.1.0 (2026-05-25, manager-spec)
- Initial design.md authored documenting architecture diagrams + decision rationale
- 5 deliverable component diagram with data flow
- Era classification heuristic table specification
- File lock mechanism design (per-SPEC scoping)
- Atomic close staging algorithm

---

## A. Architecture Overview

### A.1 Component Diagram (Logical)

```
┌─────────────────────────────────────────────────────────────────────────┐
│                     SPEC-V3R6-LIFECYCLE-SYNC-GATE-001                   │
│                          5 Deliverables (Logical View)                   │
└─────────────────────────────────────────────────────────────────────────┘

         ┌──────────────────────┐
         │  User Invocation     │
         │  (CLI / git commit)  │
         └──────────┬───────────┘
                    │
        ┌───────────┴───────────────────────┐
        │                                   │
        ▼                                   ▼
┌───────────────────┐               ┌───────────────────┐
│ moai spec close   │               │ git commit (hook) │
│ (CLI subcommand)  │               │                   │
└──────────┬────────┘               └──────────┬────────┘
           │                                    │
           ▼                                    ▼
┌─────────────────────┐                ┌─────────────────────────┐
│ internal/spec/      │                │ .claude/hooks/moai/     │
│ closer.go           │                │ handle-pre-commit-      │
│ (atomic close core) │                │ spec-status.sh          │
└──────────┬──────────┘                │ (mismatch detection)    │
           │                           └──────────┬──────────────┘
           │                                      │
           │ (read)                               │ (read)
           ▼                                      ▼
       ┌─────────────────────────────────────────────────┐
       │ .moai/specs/SPEC-XXX/                           │
       │   spec.md frontmatter (status field SSOT)       │
       │   progress.md §E.2/§E.3/§E.5 (sync/run/mx)      │
       │   acceptance.md AC status                       │
       └─────────────────────────────────────────────────┘
                            ▲
                            │ (audit reads)
                            │
                ┌───────────┴───────────┐
                │ moai spec audit       │
                │ (CLI subcommand)      │
                └───────────┬───────────┘
                            │
                            ▼
                ┌───────────────────────┐
                │ internal/spec/audit.go│
                │ (era classification)  │
                └───────────────────────┘

                  ┌────────────────────────────────┐
                  │ Defensive Lint Layer           │
                  │ internal/spec/lint_ownership.go│
                  │ OwnershipTransitionRule         │
                  │ (commit subject vs status diff)│
                  └────────────────────────────────┘

                  ┌────────────────────────────────────────┐
                  │ Protocol SSOT                          │
                  │ .claude/rules/moai/workflow/           │
                  │   lifecycle-sync-gate.md               │
                  │ (era heuristic table + grandfather    │
                  │  clause + frontmatter era semantics)  │
                  └────────────────────────────────────────┘
```

### A.2 Data Flow — Atomic Close Path

```
User: moai spec close SPEC-XXX
  │
  ▼
[CLI layer] internal/cli/spec_close.go
  │ parse flags (--backfill-only, --dry-run, --force)
  ▼
[Core layer] internal/spec/closer.go::Close(specID, opts)
  │
  ├──── (1) Acquire file lock .moai/state/spec-close-<SPEC-ID>.lock (flock-based)
  │     │ on contention: exit 1, "lock held"
  │     ▼
  ├──── (2) Read & parse spec.md frontmatter (yaml.Unmarshal)
  │     │ check: status == implemented (or in-progress with --backfill-only)
  │     ▼
  ├──── (3) Read & parse progress.md
  │     │ extract §E.2 sync_commit_sha, §E.5 mx_commit_sha
  │     │ check: both present OR --backfill-only + recent commits findable
  │     ▼
  ├──── (4) Read & parse acceptance.md AC status table
  │     │ check: all MUST-PASS AC == PASS, no PASS-WITH-DEBT
  │     ▼
  ├──── (5) Validate 4-phase precondition matrix
  │     │ on failure: exit 1, name missing artifact, NO staging
  │     ▼
  ├──── (6) Compute new state:
  │     │   spec.md.frontmatter.status = "completed"
  │     │   progress.md §E.3.status = "completed"
  │     │   progress.md §E.2.sync_commit_sha = <derived>
  │     │   progress.md §E.5.mx_commit_sha = <derived>
  │     │   spec.md §A Lifecycle Sync row updated
  │     ▼
  ├──── (7) Write all changes to disk + git add spec.md progress.md
  │     │ if --dry-run: print diff + skip commit
  │     ▼
  ├──── (8) git commit -m "chore(SPEC-XXX): 4-phase close — atomic\n\n🗿 MoAI <email@mo.ai.kr>"
  │     │ on commit failure: roll back staging (git restore --staged)
  │     ▼
  ├──── (9) Release file lock
  │     ▼
  └──── exit 0 + emit JSON summary if --json
```

### A.3 Data Flow — Audit Path

```
User: moai spec audit --json [--filter-era=V3R6] [--include-grandfathered]
  │
  ▼
[CLI layer] internal/cli/spec_audit.go
  │ parse flags
  ▼
[Core layer] internal/spec/audit.go::Audit(opts)
  │
  ├──── (1) glob .moai/specs/SPEC-*/spec.md → spec_files
  │     ▼
  ├──── (2) for each spec_file in parallel (goroutine pool):
  │     │   read spec.md frontmatter
  │     │   read progress.md (if present)
  │     │   classify era per §C.2 heuristic table
  │     │   if era ∈ {V2.x, V3R2-R4, V3R5}: mark grandfathered
  │     │   if era == V3R6:
  │     │     check cross-tab: progress.md present, sync section present, mx section present, status field consistency
  │     │     emit drift finding if applicable
  │     ▼
  ├──── (3) aggregate results
  │     ▼
  ├──── (4) emit output (JSON or human-readable)
  │     ▼
  └──── exit 0 (or 1 if --strict + drift findings present)
```

## B. Component Detail

### B.1 closer.go (~300 LOC) — Atomic Close Core

**Public API**:
```go
type CloseOptions struct {
    BackfillOnly bool
    DryRun       bool
    Force        bool
}

type CloseResult struct {
    SpecID        string
    CommitSHA     string
    Transitions   map[string]string // field → new value
    LockHeld      bool
    PreconditionsFailed []string
}

func Close(specID string, opts CloseOptions) (*CloseResult, error)
```

**Key invariants**:
- File lock per SPEC (`.moai/state/spec-close-<SPEC-ID>.lock`) acquired via `golang.org/x/sys/unix` flock OR fallback to atomic-create-file pattern on Windows
- All 5 state transitions staged or NONE staged (atomic rollback via `git restore --staged` on failure)
- spec.md + progress.md YAML/markdown parsing preserves comment lines + key ordering (use `gopkg.in/yaml.v3` Node API, not Marshal/Unmarshal)
- mx_commit_sha self-reference resolved via post-staging git commit-tree + amend (or `--allow-empty-message` placeholder; design decision below)

**Decision: mx_commit_sha self-reference resolution**:
The atomic close commit's SHA cannot be known until the commit object is created. Two approaches considered:
- **Option A: Post-staging amend** — stage changes, commit with placeholder SHA, then amend with actual SHA. Issue: `--amend` rewrites the commit, breaking the atomic guarantee semantically (two commit objects exist transiently).
- **Option B: Placeholder + sync-phase backfill** — stage with `mx_commit_sha: <PENDING>`, then sync-phase backfills via separate chore. Issue: re-introduces 2-commit cadence, defeating atomicity.
- **Option C (CHOSEN): Predictive SHA via `git commit-tree` + reflog** — stage changes, compute tree-ish + parent hash, predict commit SHA via SHA-256/SHA-1 computation, stage final state with predicted SHA, commit. Risk: SHA prediction fragile across git versions.
- **Option D (FALLBACK CHOSEN if C proves fragile)**: Accept that mx_commit_sha references the SHA at which mx-phase audit-ready signal was emitted (i.e., the previous Mx commit when using `--backfill-only`); for fresh close, mx_commit_sha is the close commit's own SHA, resolved via tag-after-commit pattern. The progress.md `§E.5 mx_commit_sha` field is intentionally allowed to self-reference; tooling handles the recursion gracefully.

**Chosen mechanism for v0.1.0 implementation**: Option D with tag-after-commit. M1 implementation may revisit if D proves operationally problematic.

### B.2 audit.go (~200 LOC) — Era Classification Engine

**Public API**:
```go
type AuditOptions struct {
    JSONOutput          bool
    FilterEra           string
    IncludeGrandfathered bool
    Strict              bool
}

type AuditResult struct {
    AuditedAt           time.Time
    TotalSpecs          int
    Grandfathered       int
    ModernEraClean      int
    DriftFindings       []DriftFinding
}

type DriftFinding struct {
    SpecID         string
    Era            string
    FindingType    string  // Y_N_N_Y, Y_Y_N_Y, Y_Y_Y_Y_StatusDrift
    Severity       string  // MUST-FIX, INFO
    Remediation    string
    Details        map[string]interface{}
}

func Audit(opts AuditOptions) (*AuditResult, error)
```

### B.3 Pre-Commit Hook (~80 LOC bash)

**File**: `.claude/hooks/moai/handle-pre-commit-spec-status.sh`

**Behavior**:
1. Read stdin JSON (Claude Code hook protocol)
2. Extract `staged_files` from git diff --cached --name-only
3. For each staged spec.md OR progress.md, parse frontmatter `status` field via `awk`/`yq`
4. Compare cross-file invariants
5. If commit message matches canonical 4-phase close pattern, enforce spec.md status: completed
6. Emit exit 0 (PASS) or exit 2 (BLOCK) + structured JSON to stdout
7. NEVER invoke AskUserQuestion (REQ-LSG-011 HARD)

**Output JSON schema on exit 2**:
```json
{
  "continue": false,
  "stopReason": "<descriptive>",
  "details": {
    "spec_id": "SPEC-XXX",
    "spec_md_status": "<value>",
    "progress_md_status": "<value>",
    "resolution_command": "moai spec close SPEC-XXX --backfill-only"
  }
}
```

### B.4 Spec-Lint Extension (~50 LOC)

**File**: `internal/spec/lint_ownership.go`

**Interface implementation**:
```go
type OwnershipTransitionRule struct{}

func (r *OwnershipTransitionRule) Check(ctx LintContext) []Finding {
    // 1. parse commit subject + author trailer
    // 2. parse staged diff for spec.md status field changes
    // 3. consult Status Transition Ownership Matrix from spec-frontmatter-schema.md
    // 4. emit OwnershipTransitionInvalid finding on mismatch
}
```

Registered in `internal/spec/lint.go` Rule registry alongside existing FrontmatterSchemaRule, MissingExclusions, LegacyEARSKeyword, etc.

### B.5 Rule File (~250 LOC)

**File**: `.claude/rules/moai/workflow/lifecycle-sync-gate.md`

**Section structure**:
1. `# Lifecycle Sync Gate Protocol — SSOT` (h1)
2. `## Era Classification Heuristic` (h2)
3. `## Grandfather Clause Policy` (h2)
4. `## Frontmatter Era Field Semantics` (h2)
5. `## Status Transition Ownership Matrix Cross-Reference` (h2)
6. `## Worked Example: Era Auto-Detection` (h2)

Each h2 section ≥ 30 lines authored with worked examples.

## C. Era Classification Heuristic Table

### C.1 Era Definitions

| Era | Period | Lifecycle Standard |
|-----|--------|-------------------|
| V2.x | Pre-2026-02 | No progress.md; SPEC implementation via direct commit |
| V3R2-R4 | 2026-02 ~ 2026-03 | progress.md introduced; no sync_commit_sha |
| V3R5 | 2026-03 ~ 2026-04 | sync section emerges; sync_commit_sha not enforced |
| V3R6 | 2026-04 ~ present | 4-phase modern standard (plan / run / sync / Mx); sync_commit_sha + mx_commit_sha required |

### C.2 Heuristic Detection Table

| Heuristic ID | Rule | Era Inferred |
|--------------|------|--------------|
| H-1 | `.moai/specs/SPEC-*/progress.md` absent | V2.x |
| H-2 | progress.md present, no §E.2 / §E.3 / §E.4 / §E.5 section markers | V3R2-R4 |
| H-3 | progress.md §E.2 present, sync_commit_sha field absent OR null | V3R5 |
| H-4 | progress.md §E.2 present + sync_commit_sha SHA value + §E.5 present + mx_commit_sha SHA value | V3R6 |
| H-5 | spec.md frontmatter `phase:` field references "v3.0" OR `v3R6` OR commit at created-date >= 2026-04-01 | V3R6 (tie-breaker) |
| H-6 | Auto-detection ambiguous (no heuristic matches) | unclassified |

**Override**: spec.md frontmatter optional `era:` field, if present, overrides auto-detection.

### C.3 Grandfather Clause

SPECs classified as V2.x, V3R2-R4, or V3R5 are **grandfather-clause-protected**:
- `moai spec audit` SHALL classify them as `era_final: true`
- NO drift findings emitted regardless of cross-tab pattern
- `--include-grandfathered` flag surfaces them in JSON output for completeness, but `severity: "INFO"` only

**Rationale**: Retroactive normalization of 145 historical SPECs is operationally infeasible and provides no production value. The grandfather clause acknowledges era-specific standards were legitimate at their time of authorship.

## D. File Lock Mechanism

### D.1 Lock Scope

Per-SPEC file lock: `.moai/state/spec-close-<SPEC-ID>.lock`

### D.2 Acquisition Protocol

**POSIX (Darwin/Linux)**:
```go
import "golang.org/x/sys/unix"

lockFile, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0644)
if err != nil { return err }
defer lockFile.Close()

if err := unix.Flock(int(lockFile.Fd()), unix.LOCK_EX|unix.LOCK_NB); err != nil {
    return ErrLockHeld
}
defer unix.Flock(int(lockFile.Fd()), unix.LOCK_UN)
```

**Windows (fallback)**:
Atomic-create-file pattern: `os.OpenFile(lockPath, os.O_CREATE|os.O_EXCL|os.O_RDWR, 0644)` — fails if file exists. Stale lock detection via PID + timestamp embedded in lock file body.

### D.3 Lock Release

Released via `defer` on function exit. On crash, stale lock file remains; auto-cleanup on next invocation if PID embedded in lock body is no longer alive (post-MVP enhancement; M1 may leave as known-issue requiring manual `rm .moai/state/spec-close-*.lock`).

## E. Open Design Decisions (deferred to M1)

1. **Tree-ish SHA prediction (Option C vs D in B.1)**: Whether to attempt Option C (predictive SHA via git commit-tree) or accept Option D (post-commit self-reference) in v0.1.0. Recommendation: start with Option D, escalate to Option C only if AC-LSG-001 verification fails.

2. **Hook bash dependency**: `yq` is convenient for YAML parsing but adds external dependency. Alternative: pure-bash `awk`-based parser. Recommendation: pure-bash `awk` for portability.

3. **Audit parallelism**: 200-SPEC fixture timing (NFR-LSG-001) suggests parallel goroutine pool (~8 workers). Single-threaded baseline may also meet 5s bound; defer optimization to M1 benchmark.

4. **CHANGELOG sync entry**: sync-phase will add CHANGELOG.md entry under v3.0.1 unreleased section noting `moai spec close` + `moai spec audit` introduction. Plan-phase MUST NOT modify CHANGELOG.md (A.5.3 Out of Scope).

## F. Cross-References

- spec.md §B requirements
- acceptance.md §B AC detailed specifications
- plan.md §F milestones
- research.md §B root cause analysis
- `.claude/rules/moai/development/spec-frontmatter-schema.md` § Status Transition Ownership Matrix
- `.claude/rules/moai/core/agent-common-protocol.md` § Hook Invocation Surface
