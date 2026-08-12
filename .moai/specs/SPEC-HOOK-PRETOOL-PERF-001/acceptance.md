# Acceptance Criteria — SPEC-HOOK-PRETOOL-PERF-001

> Verification layer for SPEC-HOOK-PRETOOL-PERF-001. Each AC is `AC-PERF-XXX`, labeled `Given … When … Then …`, and binary-testable. The GEARS obligation lives in `spec.md` §C (`REQ-PERF-XXX`); this file is the verification layer and uses Given-When-Then as required by the plan-phase acceptance contract.

## A. Traceability Matrix

| AC | REQ | Severity | Verification kind |
|----|-----|----------|-------------------|
| AC-PERF-001 | REQ-PERF-001 | MUST | Unit test — cache hit skips per-section parse |
| AC-PERF-002 | REQ-PERF-002 | MUST | Unit test — mtime change invalidates |
| AC-PERF-003 | REQ-PERF-003 | MUST | Unit test — deletion invalidates |
| AC-PERF-004 | REQ-PERF-004 | MUST | Unit test — corrupt cache fails open |
| AC-PERF-005 | REQ-PERF-005 | SHOULD | Unit test — lazy slice on cache miss |
| AC-PERF-006 | REQ-PERF-006 | MUST | Profiling milestone — baseline + post-change timing |
| AC-PERF-007 | REQ-PERF-007 | MUST (SECURITY) | Regression test — fast-path must not bypass destructive-primitive scan |
| AC-PERF-008 | REQ-PERF-008 | MUST | Concurrency test — atomic cache write under parallel writers |
| AC-PERF-009 | REQ-PERF-009 | MUST | Grep/test — cache location under state dir |
| AC-PERF-010 | REQ-PERF-010 | MUST | Gating assertion — timeout narrowing is measurement-gated |

## B. Acceptance Criteria (Given-When-Then)

### AC-PERF-001 — Cache hit skips full per-section parse

**Given** a project with `.moai/config/sections/*.yaml` populated and a valid cache file at `.moai/state/config-cache.json` whose recorded mtime fingerprint matches every section file's current mtime,
**When** the PreToolUse handler requests the config via `ConfigProvider`,
**Then** the loader SHALL return the cached config WITHOUT opening any `.moai/config/sections/*.yaml` file, verified by a test that counts file-open syscalls (or section-loader invocations) and asserts the count is 0 on the cache-hit path.

### AC-PERF-002 — Section mtime change invalidates the cache

**Given** a valid cache file whose recorded mtime fingerprint references `quality.yaml` at mtime T0,
**When** `quality.yaml` is modified so its mtime advances to T1 (> T0) and the PreToolUse handler requests the config,
**Then** the loader SHALL re-read and re-merge `quality.yaml`, SHALL rewrite the cache with the new fingerprint (T1), and SHALL serve a config whose `Quality` field reflects the modified content (not the stale cached content).

### AC-PERF-003 — Section deletion invalidates the cache

**Given** a valid cache file whose recorded fingerprint lists `feedback.yaml` as present at cache-write time,
**When** `feedback.yaml` is removed from the filesystem (deleted) and the PreToolUse handler requests the config,
**Then** the loader SHALL detect the deletion as an invalidation signal (equal force to an mtime change), SHALL re-merge the surviving sections, SHALL rewrite the cache with a fingerprint that no longer references `feedback.yaml`, and SHALL NOT serve the stale cached `Feedback` field.

### AC-PERF-004 — Corrupt or schema-mismatched cache fails open

**Given** a cache file at `.moai/state/config-cache.json` that is (a) syntactically corrupt JSON, OR (b) carries a `schema_version` field that does not match the running binary's config schema version,
**When** the PreToolUse handler requests the config,
**Then** the loader SHALL NOT emit a user-facing error, SHALL NOT exit non-zero, SHALL fall back to the full re-merge path, SHALL serve a correct config to the handler, and SHALL overwrite the corrupt cache file with a valid one on successful re-merge.

### AC-PERF-005 — Lazy config slice on cache miss

**Given** a cache miss (no cache file, or fingerprint invalid),
**When** the PreToolUse handler requests ONLY the security policy + branch-guard flag + gate config,
**Then** the loader SHALL read only the section files backing those slices (e.g. `quality.yaml` for gate config, the security-pattern source) and SHALL NOT read unrelated sections (e.g. `language.yaml`, `statusline.yaml`, `ralph.yaml`) — verified by a test that asserts the set of section files opened on the lazy-load path is a strict subset of the full section set.

### AC-PERF-006 — Profiling milestone produces baseline + post-change timing

**Given** a simulated concurrent-hook stress harness (≥8 parallel `moai hook pre-tool` invocations against a fixture project, repeated across ≥5 batches),
**When** the profiling milestone runs BEFORE the cache+lazy implementation (baseline) and AGAIN AFTER the implementation (post-change),
**Then** the milestone SHALL produce a per-phase timing report (fork/exec wall-time, config-load wall-time, security-scan wall-time, total wall-time, max-wall-time tail) for BOTH runs, and the post-change total-wall-time max-tail SHALL be lower than the baseline max-tail by a margin the plan documents (e.g. ≥30% reduction in the p99 tail under concurrent stress). The measurement, not the code change, is the evidence of root-cause resolution.

### AC-PERF-007 (SECURITY — make-or-break) — Fast-path must not bypass destructive-primitive scan

**Given** a hypothetical fast-path (introduced by this SPEC, a follow-up SPEC, or any other change) that short-circuits any subset of PreToolUse Bash invocations before the full config + scan path,
**When** the fast-path receives a Bash command whose text matches any entry of the Bash Risk-Amplifier destructive-primitive set — `rm -rf`, `git push --force`, `git push -f`, `git push --no-verify`, `git commit --no-verify`, `git reset --hard`, `DROP TABLE`, `TRUNCATE`, `chmod -R 777` — OR whose compound-subcommand count (pipe `|`, `&&`, `||`, `;`, backtick, `$(...)`) exceeds `BASH_SUBCOMMAND_SOFT_CAP` (5),
**Then** the fast-path SHALL decline to short-circuit (i.e. SHALL route the invocation through the full config + security scan path), so that the destructive-primitive / security scan always runs on the commands that need it. A regression test SHALL enumerate each destructive primitive as a separate case and assert the fast-path does NOT fire for any of them. This AC binds the fast-path design space even though this SPEC defers the fast-path itself.

### AC-PERF-008 — Atomic cache write under concurrent writers

**Given** ≥4 concurrent goroutines (or ≥4 concurrent `moai hook pre-tool` subprocesses) that each observe a cache miss and each attempt to rewrite the cache,
**When** all writers race to write `.moai/state/config-cache.json`,
**Then** every reader observed during the race SHALL observe EITHER the previous valid cache file OR trigger the fail-open re-merge path (AC-PERF-004), and SHALL NEVER observe a partially-written file (no truncated JSON, no zero-byte intermediate state). Verified by a concurrency test using `sync.WaitGroup` + `errgroup` that performs N parallel writes and asserts every read returns a parseable config.

### AC-PERF-009 — Cache location under state dir

**Given** the cache-write path constant (or path-resolver function) defined by REQ-PERF-009,
**When** the loader resolves the cache file path,
**Then** the resolved path SHALL be lexically under the project's resolved state directory (default `.moai/state/`), verified by a test that resolves the path with a custom state dir and asserts `strings.HasPrefix(resolvedPath, stateDir)`. The cache file name SHALL be fixed (e.g. `config-cache.json`) so it is predictable for gitignore and cleanup.

### AC-PERF-010 — Timeout narrowing is measurement-gated

**Given** the 10s timeout currently configured in `internal/template/templates/.claude/settings.json.tmpl` + `internal/hook/CLAUDE.md`,
**When** a change attempts to narrow the timeout back toward the 5s MoAI policy default,
**Then** the change SHALL cite the post-change profiling milestone evidence (AC-PERF-006) demonstrating per-invocation cost under concurrent-hook stress has dropped below the threshold that previously produced the 33-timeouts/30-days tail. A change that narrows the timeout speculatively (without citing the measurement) is a regression and MUST be rejected at review.

## C. Edge Cases

- **C-EDGE-001 (cache file written by an older binary)**: a cache file produced by moai-adk v3.0.1 read by v3.0.2 carries an older `schema_version`. REQ-PERF-004 / AC-PERF-004 handle this as fail-open re-merge. The test SHALL include a schema-version-mismatch case.
- **C-EDGE-002 (cache file produced by a newer binary, read by older)**: the older binary either cannot parse the cache (fail-open) or parses it with an unknown schema_version (fail-open). Same handling as C-EDGE-001.
- **C-EDGE-003 (state dir does not exist / is not writable)**: the cache write fails. The loader SHALL fail open (full re-merge path, no user-facing error) and the PreToolUse handler SHALL still receive a correct config. The cache is an optimization, not a correctness dependency.
- **C-EDGE-004 (section file mtime equal to cache fingerprint mtime)**: equal mtimes are treated as cache-valid (no re-read). Only strictly-newer mtimes invalidate. This matches the standard mtime-cache convention and avoids infinite re-read loops.
- **C-EDGE-005 (clock skew / mtime in the future)**: a section file with an mtime in the future relative to the cache-write time is treated as newer (invalidation fires). The cache is conservative: when in doubt, re-merge.
- **C-EDGE-006 (worktree hook reading primary checkout's cache)**: C-5 requires `$CLAUDE_PROJECT_DIR`-first resolution. A worktree hook SHALL read/write the worktree's own `.moai/state/config-cache.json`, NOT the primary checkout's cache, so the primary checkout's cache fingerprint is not polluted by a worktree's section-file state.

## D. Definition of Done

- All MUST ACs (AC-PERF-001, -002, -003, -004, -006, -007, -008, -009, -010) PASS with attributable evidence (test output or profiling report).
- AC-PERF-005 (SHOULD) PASS OR an explicit `[NEEDS CLARIFICATION]` is recorded in `plan.md` with rationale for deferral.
- The profiling milestone (AC-PERF-006) produces a written baseline + post-change report committed to the SPEC directory.
- The make-or-break security AC (AC-PERF-007) has a regression test that enumerates EVERY destructive primitive in the Bash Risk-Amplifier set as a separate case.
- `go test ./internal/config/... ./internal/hook/...` exit 0.
- `golangci-lint run ./internal/config/... ./internal/hook/...` exit 0.
- No fast-path implementation is merged without AC-PERF-007's regression test landing first.
