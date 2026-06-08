# Implementation Plan — SPEC-PREPUSH-WIRING-001

**Tier**: S (minimal) · **cycle_type**: tdd (RED-GREEN-REFACTOR)

## A. Context

Wire the dormant pre-push convention engine end-to-end by appending a gated
`moai hook pre-push` invocation to the distributed pre-push shell hook. `make ci-local`
is retained; the convention block runs only after ci-local passes and is a no-op by
default (engine self-gates on `enforce_on_push`, which ships `false`).

Ground truth (verified by file reads — see §H Cross-References for exact anchors):
- The pre-push hook exists in **3 byte-identical copies** (mirror set), of which only the
  template↔constant pair is test-enforced; the root copy is hand-maintained.
- The Go engine (`runPrePush`) is complete and self-gating — no Go logic change needed.
- git's pre-push protocol passes ref-update lines on stdin, NOT commit messages — the
  appended shell block must translate.

## B. Known Issues / Risks

1. **git stdin semantic mismatch (highest risk)** — git pre-push stdin is ref-update lines
   (`<local ref> <local oid> <remote ref> <remote oid>`), but `runPrePush` expects commit
   MESSAGES (one per line). The shell block MUST translate via a `while read ... done` loop
   computing a range per ref and piping `git log --format=%s <range>` to the subcommand.
2. **stdin consumed-once** — git's ref-update stdin can be read only once. `make ci-local`
   precedes the convention block. Design decision (§D) must guarantee stdin is captured
   before any consuming step.
3. **Zero-SHA edge cases** — new branch (remote oid all-zeros) and branch delete (local oid
   all-zeros) must be handled: enumerate locally-new commits vs skip the ref entirely.
4. **Mirror drift** — 3 copies must stay byte-identical; forgetting the root copy or
   skipping `make build` produces a `TestPrePushTemplateMatchesConstant` failure or a stale
   `embedded.go`.
5. **Template neutrality** — the block ships to 16 languages; no internal markers, generic
   POSIX shell only, `command -v moai` guard for non-moai projects.

## C. Pre-flight Checks

```bash
# Confirm the three mirror copies are currently byte-identical (baseline)
go test ./internal/cli/ -run TestPrePushTemplateMatchesConstant -count=1
diff internal/template/templates/.git_hooks/pre-push .git_hooks/pre-push && echo ROOT_MATCHES_TEMPLATE

# Confirm the Go engine subcommand dispatches (no change planned here)
grep -n "func runPrePush" internal/cli/hook_pre_push.go
grep -n 'Use:   "pre-push"' internal/cli/hook_pre_push.go

# Confirm the template default stays false (do NOT change it)
grep -n "enforce_on_push" internal/template/templates/.moai/config/sections/git-convention.yaml
```

## D. Design Decisions

### D.1 Placement (REQ-PPW-002)

The convention block is inserted **after** the `ci-local` fail-check `fi` (template line 39)
and **before** the final `exit 0` (template line 41). This guarantees the block runs only
after ci-local passes (the fail branch `exit "$EXIT_CODE"` already returns before reaching
it). Appending after `exit 0` would make the block dead code — explicitly avoided.

### D.2 stdin capture ordering (REQ-PPW-008)

Decision: **capture git's ref-update stdin into a shell variable at the top of the script,
before `make ci-local`**, then feed the captured value to the convention block.

Rationale: stdin is consumed-once. Capturing early (`REFS="$(cat)"` near the top, after the
`SKIP_MOAI_PREPUSH` bypass) is the robust, make-implementation-agnostic choice — it does not
rely on assuming `make ci-local` never reads stdin. The captured variable is then replayed
into the translation loop. `set -eu` is already active (template line 5); the capture must
tolerate empty stdin (a push can arrive with no ref-update lines in edge runs) without
tripping `set -e`.

> Implementation note for run phase: `REFS="$(cat)"` under `set -e` is safe for empty input
> (cat exits 0). The translation loop reads from the variable, e.g. via a here-string
> `while read ...; do ...; done <<EOF` / `printf '%s\n' "$REFS" | while read ...`. The run
> phase selects the exact idiom; the design constraint is "capture before ci-local, replay
> after".

### D.3 Translation loop (REQ-PPW-005/006/007)

git-recommended idiom: `while read local_ref local_oid remote_ref remote_oid; do ... done`.
Per ref:
- local_oid all-zeros (`0000...`) → branch delete → **skip** (REQ-PPW-007).
- remote_oid all-zeros → new remote ref → range enumerates locally-new commits
  (e.g. `git log --format=%s <local_oid>` limited to commits not on any remote, or
  `git log --format=%s <local_oid> --not --remotes`) (REQ-PPW-006). Note: on a clone with
  no remote-tracking refs, `--not --remotes` resolves to nothing and the range degrades to
  all ancestors of `<local_oid>` — a benign superset (over-validates, never under-validates),
  acceptable because `enforce_on_push` is opt-in (plan-auditor iter-1 D3).
- otherwise → range `git log --format=%s <remote_oid>..<local_oid>` (REQ-PPW-005).

Collected `%s` subject lines are piped to `moai hook pre-push` (which reads `/dev/stdin`,
one message per line, per `readStdinLines`).

### D.4 moai-presence guard (REQ-PPW-003)

The convention block is wrapped in `if command -v moai >/dev/null 2>&1; then ... fi`. The
engine additionally self-gates on `enforce_on_push` (REQ-PPW-004), so the block is a no-op
by default even when moai IS installed.

### D.5 Mirror discipline (REQ-PPW-009/010)

Template-First (CLAUDE.local §2): edit `internal/template/templates/.git_hooks/pre-push`
first, then mirror the byte-identical change into the `prePushHookContent` constant and the
root `.git_hooks/pre-push` copy, then run `make build` to regenerate `embedded.go`.

## E. Self-Verification (run-phase exit gate)

```bash
# 1. Byte-parity (template ↔ constant) — canonical, not a hand grep diff
go test ./internal/cli/ -run TestPrePushTemplateMatchesConstant -count=1   # exit 0

# 2. Root copy mirror parity (hand-maintained leg)
diff internal/template/templates/.git_hooks/pre-push .git_hooks/pre-push && echo ROOT_OK

# 3. Install assertion (must contain moai hook pre-push token now)
go test ./internal/cli/ -run 'TestInstallPrePushHook' -count=1 -v          # all PASS

# 4. Full package suite (cascading regressions)
go test ./internal/cli/ ./internal/template/... -count=1

# 5. embedded.go regenerated
make build

# 6. Lint
golangci-lint run --timeout=2m ./internal/cli/...
```

## F. Milestones

| Milestone | Description | Files | Verify |
|-----------|-------------|-------|--------|
| **M1 (RED)** | Add the failing wiring test: extend `TestInstallPrePushHook_FreshRepo` `wantStrings` to include the `moai hook pre-push` invocation token; add any new RED test asserting the moai invocation appears before the final `exit 0` in the template. Confirm RED (test fails because hook content lacks the token). | `internal/cli/hook_install_test.go` (+ optional new placement test) | `go test ./internal/cli/ -run 'TestInstallPrePushHook' -count=1` fails RED |
| **M2 (GREEN — template)** | Append the gated convention block to the distributed template (capture stdin top, translation loop after ci-local fail-check, `command -v moai` guard). Template-First. | `internal/template/templates/.git_hooks/pre-push` | block present; placement after fail-check `fi`, before final `exit 0` |
| **M3 (GREEN — mirrors)** | Mirror the byte-identical change into `prePushHookContent` constant and the root `.git_hooks/pre-push` copy. Run `make build` to regenerate `embedded.go`. | `internal/cli/hook_install.go`, `.git_hooks/pre-push`, `internal/template/embedded.go` (generated) | `TestPrePushTemplateMatchesConstant` PASS; `diff` root==template OK |
| **M4 (REFACTOR + verify)** | Run §E self-verification gate in full. Confirm install assertion now GREEN, byte-parity GREEN, full suite GREEN, lint clean. No Go engine change. | (verification only) | All §E commands pass |

> Tier S: no design.md / research.md. M1-M4 only. The ordering M2→M3 is sequential
> (mirror legs depend on the template content being finalized first).

## G. Anti-Patterns to Avoid

- **Appending the block after `exit 0`** → dead code (REQ-PPW-002 / D.1). Place before it.
- **Reading git stdin twice** → second read gets nothing. Capture once into a variable (D.2).
- **Piping ref-update lines directly to `moai hook pre-push`** → the engine expects commit
  messages, not `<ref> <sha> <ref> <sha>` lines. Translate first (D.3).
- **Editing the constant or root copy without the template (or vice versa)** → byte-parity
  test failure / silent mirror drift. Edit template first, mirror both legs (D.5).
- **Hand-editing `embedded.go`** → overwritten by `make build`; run the build instead.
- **Flipping `enforce_on_push` to `true`** → out of scope; the wired hook must stay a no-op
  by default.
- **Leaking internal markers (SPEC ID, REQ, dates) into the template block** → template
  neutrality violation (CLAUDE.local §15/§25). Generic POSIX shell only.

## H. Cross-References (exact ground-truth anchors)

- `internal/template/templates/.git_hooks/pre-push` (42 lines) — `set -eu`:5;
  SKIP bypass:7-10; `make ci-local`:19-25; fail-check exit:34-39; final `exit 0`:41.
- `internal/cli/hook_install.go` — `moaiPrePushMarker`:16; `prePushHookContent`:27-68
  (byte-identical-mirror comment:24-26).
- `.git_hooks/pre-push` — dev-repo root copy, byte-identical, git-tracked, NOT covered by
  `make build` (only `internal/template/templates/` is embedded) → hand-maintained leg.
- `internal/cli/hook_pre_push.go` — `runPrePush`:33; `isEnforceOnPushEnabled`:154-167;
  `readStdinLines` (reads `/dev/stdin`):169-186; subcommand registration:18-30. DO NOT MODIFY.
- `internal/cli/wave1_sync_test.go` — `TestPrePushTemplateMatchesConstant`:19 (template↔constant only).
- `internal/cli/hook_install_test.go` — `TestInstallPrePushHook_FreshRepo`:15
  (`wantStrings`:45-48 = `[moaiPrePushMarker, "ci-local"]`).
- `internal/template/templates/.moai/config/sections/git-convention.yaml`:23 — `enforce_on_push: false` (stays).
