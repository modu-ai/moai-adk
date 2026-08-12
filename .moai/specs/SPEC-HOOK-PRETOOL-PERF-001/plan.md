# Implementation Plan — SPEC-HOOK-PRETOOL-PERF-001

> Plan-phase artifact. Tier M. The milestones below are ordered by **decision-reversibility** (most likely to change first), NOT by execution order. M0 (profiling milestone) is the gating decision; M1 (cache shape) is the highest-change-likelihood code decision; M2 (lazy slice) is a pure refinement; M3 (validation) closes the loop.

## §A. Context (Why this plan, not another)

The PreToolUse hook path forks a fresh `moai` process and loads the full merged config on every invocation. Normal-path cost is 0.05s; under system load the fork+exec+full-config structure amplifies intermittently past 5s (33 timeouts / 30 days, all `PreToolUse:Bash`, <1% of invocations). The 5→10s timeout widening is immediate mitigation; this plan is the ROOT fix.

**Why config cache + lazy slice, not a daemon or a fast-path:**

| Direction | Verdict | Rationale |
|-----------|---------|-----------|
| Config disk cache (mtime-keyed) | **PRIMARY** | Attacks the ~20-section re-parse+re-merge charged to every invocation. Highest leverage, lowest risk: pure optimization, fail-open on miss, no security surface. |
| Lazy config slice | **SECONDARY** | Attacks the slice mismatch — PreToolUse only consumes security+branchguard+gate but pays for ~20 sections. Composes with cache (lazy on miss, full on hit). |
| Profiling milestone (before & after) | **GATE** | The spike is not reproducible at will. Without measurement we cannot confirm the cache+lazy moves the needle. M0 is the gate that confirms or revises the directions. |
| Fast-path (in-shell or minimal-Go short-circuit) | **DEFERRED** | Highest security risk — a fast-path that lets a destructive primitive through is a regression (REQ-PERF-007 / AC-PERF-007). The shell wrapper's existing bash-risk warn is the only in-shell short-circuit and is unchanged. The fast-path is recorded as a binding REQ so any follow-up inherits the constraint, but is NOT implemented here. |
| Long-lived hook daemon | **OUT (C-2)** | Large architectural change; lifecycle/IPC/security surface dwarf the cache fix. Justified only by a follow-up SPEC if profiling shows cache+lazy insufficient. |

## §B. Known Issues

- **B-1 (spike is not reproducible at will)**: a profiler attached to a single invocation measures the 0.05s normal path. The profiling milestone MUST use simulated concurrent-hook stress (≥8 parallel `moai hook pre-tool` invocations against a fixture project), not waiting for an organic spike. [Resolved by M0 design + C-6 constraint.]
- **B-2 (mtime resolution varies by OS / filesystem)**: macOS APFS, Linux ext4, and Windows NTFS have different mtime granularity and clock-skew behavior. The fingerprint MUST record both mtime AND size, and treat equal-mtime-but-different-size as invalidation. [Resolved by M1 fingerprint design.]
- **B-3 (atomic rename cross-filesystem)**: `os.Rename` is atomic only within the same filesystem. The cache temp file MUST be created in the same directory as the final cache file (`.moai/state/`), not in `/tmp`. The `internal/config/atomicfile` helper (SPEC-CONFIG-ATOMIC-WRITE-001) already implements this pattern — reuse, do not reinvent. [Resolved by M1 reuse decision.]
- **B-4 (concurrent writers stampeding on first invocation after a section edit)**: N parallel PreToolUse processes all observe a cache miss, all re-merge, all write. This is correct (all writes are atomic, last-writer-wins, content is deterministic for a given section set) but wasteful. A future follow-up MAY add a lock file; this SPEC accepts the stampede because the re-merge is cheap and the cache is an optimization. [Accepted as residual-risk; not blocking.]
- **B-5 (worktree path resolution)**: a hook running inside a worktree under `~/.moai/worktrees/...` MUST resolve `$CLAUDE_PROJECT_DIR` first (C-5), so it reads/writes the worktree's own cache, not the primary checkout's. The existing `internal/hook/observer.go` path-resolution helper centralizes this. [Resolved by C-5 + M1 reuse of observer.go helper.]

## §C. Pre-flight (before run-phase entry)

- [ ] **C-PRE-1**: Confirm the 5→10s timeout widening has ALREADY landed in `internal/template/templates/.claude/settings.json.tmpl` AND `internal/hook/CLAUDE.md`. This SPEC is the root fix; if the widening is not present, the immediate mitigation is missing and must land first (out of scope for this SPEC, but a precondition).
- [ ] **C-PRE-2**: Confirm `Workflow.BranchGuard.Enabled` defaults `false` in `internal/config/defaults.go` (the diagnosis evidence depends on this — if it is true, the diagnosis changes).
- [ ] **C-PRE-3**: Confirm `internal/config/atomicfile` is available for reuse (SPEC-CONFIG-ATOMIC-WRITE-001 merged). If absent, fall back to inline write-temp+rename.
- [ ] **C-PRE-4**: Confirm `internal/hook/observer.go` exposes the `$CLAUDE_PROJECT_DIR`-first path-resolution helper (B-5 depends on it).

## §D. Constraints (recap from spec.md §D)

- **C-1 fail-open invariant** — corrupt cache / missing state dir / FS error → full re-merge path, never stall.
- **C-2 no daemon** — out of scope; follow-up only.
- **C-3 no new config format** — cache file is the existing `*Config` struct encoded, plus a `schema_version` field.
- **C-4 template neutrality** — any template change goes through `internal/template/templates/` first + `make build`.
- **C-5 worktree path awareness** — `$CLAUDE_PROJECT_DIR`-first resolution.
- **C-6 profiling reproducibility** — simulated concurrent-hook stress, not organic-spike-waiting.

## §E. Self-Verification (run-phase §E skeleton — placeholders only)

> manager-spec populates §E.1 only. §E.2–§E.4 are placeholder headings for manager-develop / manager-docs.

### §E.1 Plan-phase Audit-Ready Signal

- `plan_status:` _<pending plan-audit>_
- `plan_complete_at:` _<pending>_
- Tier: M (3 artifacts: spec.md + plan.md + acceptance.md).
- REQ count: 10 (REQ-PERF-001 .. REQ-PERF-010).
- AC count: 10 (AC-PERF-001 .. AC-PERF-010), of which AC-PERF-007 is the make-or-break SECURITY AC.

### §E.2 Run-phase Evidence

_<pending run-phase>_

### §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

### §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

## §F. Milestones (ordered by decision-reversibility — highest-change first)

### M0 — Profiling milestone (GATE; confirms or revises M1/M2 directions)

**Why first**: the spike is not reproducible at will. Without baseline measurement, M1's cache shape is a guess. M0 is the gate that confirms the cache+lazy directions actually move the needle.

**Deliverables**:
- A reusable profiling harness under `internal/hook/perf/` (or `cmd/perf/`) that:
  - Spawns ≥8 parallel `moai hook pre-tool` invocations against a fixture project.
  - Repeats across ≥5 batches.
  - Captures per-phase wall-time: fork/exec, config load, security scan, total.
  - Reports p50, p99, and max-tail.
- A written baseline report committed to `.moai/specs/SPEC-HOOK-PRETOOL-PERF-001/baseline.md` (NOT a plan-phase artifact; a run-phase evidence doc) capturing the pre-change numbers.

**Gate condition**: M0 produces a baseline report. If the baseline shows the config-load phase is NOT the dominant cost under concurrent stress (i.e. the diagnosis is wrong), STOP and revise the plan before M1.

**Anti-pattern (AP-1)**: implementing M1 before M0 confirms the config-load phase is the dominant cost.

### M1 — Config disk cache (PRIMARY root fix)

**Why second**: this is the highest-change-likelihood code decision — the cache file format, fingerprint shape, and invalidation predicate are the design choices most likely to be revisited once M0's baseline is in hand.

**Deliverables**:
- A cache-file format: JSON encoding of the existing `*Config` struct + a header carrying `schema_version`, `fingerprint` (per-section `{path → {mtime, size}}` map), and `written_at`.
- A cache-read path in `internal/config/`: on `Load`, check `.moai/state/config-cache.json`; if present and fingerprint matches every current section file (same mtime AND size, no missing sections), serve the cached config without per-section reads. If absent or mismatched, fall through to the existing re-merge path (no behavior change).
- A cache-write path: after a successful re-merge, write the cache via `atomicfile.Write` (write-temp + rename in `.moai/state/`).
- Invalidation predicate: mtime-change OR size-change OR section-deletion OR schema-version-mismatch OR corrupt-file.
- Unit tests: AC-PERF-001 (hit skips parse), AC-PERF-002 (mtime change), AC-PERF-003 (deletion), AC-PERF-004 (corrupt fails open), AC-PERF-008 (atomic concurrent write), AC-PERF-009 (location under state dir).

**Reuse**: `internal/config/atomicfile` (SPEC-CONFIG-ATOMIC-WRITE-001); `internal/hook/observer.go` path-resolution (B-5).

### M2 — Lazy config slice (SECONDARY refinement)

**Why third**: pure refinement. Once the cache exists, the lazy slice only matters on a cache miss. It shrinks the cost of the miss path but does not change the cache shape.

**Deliverables**:
- A `LoadSlice(projectRoot, sliceNames ...string)` (or equivalent) on the config provider that reads only the named sections.
- A wiring change in `preToolHandler` (`internal/hook/pre_tool.go:370`) so that on a cache miss the handler requests only `security` + `branchguard` + `gateconfig` rather than the full set.
- Unit test: AC-PERF-005 (strict-subset assertion on the section files opened).

**Risk**: the wiring change touches `preToolHandler`, which is the hottest handler. Keep the change minimal — inject a `ConfigSliceProvider` interface rather than refactoring the handler's existing config access.

### M3 — Validation milestone (AFTER implementation; closes the loop)

**Why last**: pure validation. Re-runs M0's profiling harness against the post-change tree and confirms the tail dropped.

**Deliverables**:
- A post-change profiling report (`.moai/specs/SPEC-HOOK-PRETOOL-PERF-001/postchange.md`) with the same per-phase breakdown as the baseline.
- A delta summary: baseline vs post-change p50/p99/max-tail.
- The security-regression test (AC-PERF-007): enumerate every destructive primitive in the Bash Risk-Amplifier set as a separate case, assert the fast-path does NOT fire for any of them. (Even though no fast-path is implemented in this SPEC, the test lands as a forward guard.)
- Gate condition (AC-PERF-006): post-change max-tail SHALL be lower than baseline max-tail by ≥30% under concurrent stress, OR the milestone documents why the cache+lazy approach is insufficient and recommends a follow-up (fast-path or daemon).

**Anti-pattern (AP-4)**: declaring M3 complete without AC-PERF-007's regression test in place.

## §G. Anti-Patterns

- **AP-1 (implementing before measuring)**: implementing M1 before M0's baseline confirms config-load is the dominant concurrent-stress cost. The spike is structural and not reproducible at will; measurement must precede the code change.
- **AP-2 (mtime-only fingerprint)**: a fingerprint keyed only on mtime misses deletions (REQ-PERF-003) and size-change-without-mtime-change (B-2). The fingerprint MUST record both mtime AND size, and MUST treat section deletion as invalidation.
- **AP-3 (non-atomic cache write)**: a cache write that is not write-temp+rename-in-same-directory exposes readers to partial files (B-3). Reuse `internal/config/atomicfile`; do NOT write to `/tmp` and rename cross-filesystem.
- **AP-4 (fast-path without security regression test)**: a fast-path that short-circuits PreToolUse invocations without AC-PERF-007's regression test landing first is a security regression waiting to happen. AC-PERF-007 enumerates the destructive-primitive set and MUST be present before any fast-path merges.
- **AP-5 (daemon scope-creep)**: introducing a long-lived hook daemon within this SPEC. C-2 forbids it; a daemon's lifecycle/IPC/security surface is a follow-up SPEC's scope.

## §H. Cross-References

- **spec.md**: `/Users/goos/MoAI/moai-adk-go/.moai/specs/SPEC-HOOK-PRETOOL-PERF-001/spec.md` — §A baseline, §C REQs, §D constraints, §E out-of-scope.
- **acceptance.md**: `/Users/goos/MoAI/moai-adk-go/.moai/specs/SPEC-HOOK-PRETOOL-PERF-001/acceptance.md` — 10 ACs including AC-PERF-007 (make-or-break security).
- **SPEC-PRETOOL-GATE-MOVE-001** (completed) — complementary, non-overlapping; relocated the synchronous vet/lint/test gate off the PreToolUse budget but did not touch the residual fork+exec+config cost.
- **SPEC-CONFIG-ATOMIC-WRITE-001** — provides `internal/config/atomicfile` for reuse in REQ-PERF-008 / M1.
- **`.claude/rules/moai/development/coding-standards.md` § Bash Risk-Amplifier Doctrine** — the destructive-primitive set REQ-PERF-007 / AC-PERF-007 binds to.
- **`internal/hook/CLAUDE.md`** — confirms the 10s timeout widening as immediate mitigation pending this SPEC.
