# SPEC-HOOK-BACKSLASH-SYMLINK-001 — Implementation Plan

> Stateless artifact. Lifecycle status lives in `spec.md` frontmatter.

## A. Context

- Card t1556 (factory lane-12). Defect source: card t1533 codex review gate round 1,
  new out-of-card P1 — evidence (primary, in-tree):
  `.moai/reports/t1556/codex-review-gate-1.md`
  (provenance: `.moai/worktrees/t1533/.moai/reports/t1533/codex-review-gate-1.md`,
  a disposable tree).
- Defect site read and confirmed in this tree (`cad44a751`):
  `internal/hook/pre_tool.go:1397` —
  `parts := strings.Split(strings.ReplaceAll(filepath.ToSlash(p), "\\", "/"), "/")`
  inside `resolvePhysicalWalk`, the walker behind `resolveThroughExistingParent` (the
  card-t1530 repair of `checkFileAccess`).
- Attack mechanics (derived from the code, matching the reproduced finding): a
  project directory literally named `innocent\dir` symlinked outside the project.
  The unconditional split turns the component into `innocent` + `dir`; the probe of
  `<project>/innocent` fails Lstat as missing, so the walker returns
  `<project>/innocent/dir/file.txt` as an in-project unresolved tail; the boundary
  check passes; the real Write follows `innocent\dir` → outside. The validated path
  and the acted-on path diverge — the same divergence class t1530 (lexical `..`) and
  t1510 (`zoneResolve`) closed.

## B. Known Issues

- The defect is an input-normalization error INSIDE the t1530 walking pattern, not a
  failure of the pattern: the walk's probe-existing-then-rejoin structure is correct
  once it is fed correctly spelled segments. The leader hint asked to evaluate whether
  `resolveThroughExistingParent` applies here — evaluated: it IS the defective code;
  the repair narrows to platform-appropriate segmentation, and the walk (probe existing
  components via EvalSymlinks, rejoin the missing tail, pop `..` against resolved
  prefixes, depth-bound fail-closed) is retained as-is. A repair agent MAY still choose
  a different concrete mechanism (e.g. build-tagged split helper) if it satisfies
  REQ-HBS-001/004.
- `zoneSlash` (`protected_zone_path.go:36`) shares the unconditional conversion and
  feeds the protected-zone DENY-DECISION path (`resolveZoneTarget` →
  `checkProtectedZone`) — a live sibling surface of the same defect family; deferred
  to a named follow-up card in the t1454 residual ledger (spec.md §F), and must not
  be "fixed in passing" here.

## C. Pre-flight

1. `git rev-parse --short HEAD` in the card worktree — confirm the tree the RED
   measurement is pinned to.
2. Confirm baseline green: `go test -timeout 30m -count=1 ./internal/hook/` passes on
   the pre-fix tree except for the new reproduction test once added.
3. Confirm the t1530-era tests that pin the walk's guarantees (new-file rejoin, `..`
   physical pop, depth bound) are located and will act as the regression net.

## D. Constraints

- Reproduction-first (Rule 4): the failing test lands BEFORE the fix, and its RED on
  the pre-fix tree is observed and recorded (two-cell adoption,
  `verification-completeness.md` §2).
- No new dependencies; no API changes outside `internal/hook`.
- Windows semantics untouched: on Windows the conversion to `/`-segmentation stays
  correct (Windows treats `\` and `/` as interchangeable separators).
- Symlink-fixture tests skip on platforms that cannot create directory symlinks
  unprivileged; separator-semantics units stay platform-independent so every CI
  runner covers them.

## E. Self-Verification

- E1: RED cell of AC-HBS-001 observed on the pre-fix tree (verbatim go test output,
  pinned SHA).
- E2: GREEN cell of AC-HBS-001..AC-HBS-007 observed after the fix, same command form.
- E3: `GOOS=windows go build ./...` exit 0.
- E4: `go vet ./internal/hook/` and `golangci-lint run internal/hook/...` clean.
- E5: external destination file does not exist after the post-fix reproduction run
  (observed with a file-existence check, not asserted).

## F. Milestones

Ordered by decision-reversibility: the reproduction (highest-change-likelihood, proves
the defect shape) leads; the platform-mechanism decision follows; mechanical
hardening trails.

- **M1 — Reproduction test (RED) — DONE AT PLAN PHASE** [Priority High]
  Authored at plan phase per plan-audit iteration-1 D1:
  `internal/hook/pre_tool_backslash_repro_test.go` carries
  `TestCheckFileAccessPosixBackslashSymlinkEscape` (guard-level: deny + no external
  write, with a non-deny branch performing the actual write so the criterion is not
  vacuous) and `TestResolveThroughExistingParentPosixBackslashSymlinkDivergence`
  (resolver-level: resolution must land outside). Both observed RED on the pre-fix
  tree (exit 1; four-element cells `RED-HBS-001`/`RED-HBS-003` in acceptance.md;
  tree `785cfaaff`, code identical to `cad44a751`). The test file is committed WITH
  the SPEC revision commit, before any fix exists — the ordering claim rests on the
  recorded observation, taken now. Run phase proceeds directly to M2; the test file
  flips GREEN at M2.
- **M2 — Platform-appropriate segmentation (fix)** [Priority High]
  Make the walk's segment split treat `\` as a separator only on Windows (runtime
  GOOS check or build-tagged helper — repair agent's choice, constrained by
  REQ-HBS-001/004). Keep the POSIX split on `/` alone; keep the Windows split on both.
  The walker body (probe / rejoin / `..` pop / depth bound) is unchanged.
- **M3 — Platform-semantics + regression units (GREEN)** [Priority High]
  - Flip M1's reproduction to GREEN; add the false-positive unit
    (legitimate `weird\name.txt` new file inside the project → allowed,
    AC-HBS-004) — platform-independent, runs everywhere.
  - Add/keep a Windows-shape unit proving `\`-segmentation on the Windows path
    (string-level, no Windows FS required, AC-HBS-005).
  - Keep the t1530 guarantee tests green (new-file rejoin, `..` pop, depth bound).
- **M4 — Package re-measurement + evidence** [Priority Medium]
  `go test -timeout 30m -count=1 ./internal/hook/`, `GOOS=windows go build ./...`,
  vet + lint; write the measured evidence into `progress.md` §E.2/§E.3 for the
  auditor.

## G. Anti-Patterns

- Do NOT "fix" `zoneSlash` or any other ToSlash call site in the same change — scope
  discipline (spec.md §F).
- Do NOT deny every backslash-bearing POSIX path — that breaks REQ-HBS-005 legitimate
  names; the fix is resolution fidelity, not character blacklisting.
- Do NOT normalize the path before the walk with `filepath.Clean`/`Join` — that
  re-opens the t1530 lexical-`..` hole; `absoluteUncleaned` concatenation stays.
- Do NOT gate the deny on the walk's ok=false fallback alone — the fallback path
  exists for legitimate new files and must keep its existing semantics.

## H. Cross-References

- Evidence: `.moai/worktrees/t1533/.moai/reports/t1533/codex-review-gate-1.md` (card
  t1533 review round 1, the [P1] POSIX-backslash finding).
- Defect family: SPEC-INTERNAL-SECURITY-001 (REQ-SEC-007 original resolution), the
  t1510 `zoneResolve` physical-walk repair (`protected_zone_path.go`), the t1530
  `resolveThroughExistingParent` repair (`pre_tool.go`).
- Doctrine: `.claude/rules/moai/development/verification-completeness.md` §2 (two-cell
  adoption), AGENTS.local.md §4 (scoped test execution).
