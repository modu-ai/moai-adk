---
id: SPEC-HOOK-SESSIONSTART-PROBE-001
title: "SessionStart Probe — Acceptance Criteria"
version: "0.1.0"
status: completed
created: 2026-07-07
updated: 2026-07-07
author: manager-spec
priority: P2
phase: "v14.4.0"
module: ".claude/hooks/moai"
lifecycle: spec-anchored
tags: "hook, sessionstart, probe, mode-a, genuine-risk"
tier: S
---

# Acceptance Criteria — SPEC-HOOK-SESSIONSTART-PROBE-001

## §D AC Matrix

7 acceptance criteria, one per REQ (traceable via spec.md §G). All MUST pass
GREEN for the run-phase to declare done. Tier S — single-pass M1, no milestone
split.

Each AC is machine-verifiable (command + observed output). Severity defaults to
MUST-PASS; deviations flagged explicitly.

| AC ID | Traceable REQ | Severity | Verifiable via |
|-------|---------------|----------|----------------|
| AC-HOOK-001 | REQ-HOOK-001 | MUST-PASS | success-path regression snapshot diff |
| AC-HOOK-002 | REQ-HOOK-002 | MUST-PASS | stubbed-PATH/HOME bash invocation + stdout/stderr capture |
| AC-HOOK-003 | REQ-HOOK-003 | MUST-PASS | `echo $?` after fallback invocation |
| AC-HOOK-004 | REQ-HOOK-004 | MUST-PASS | 4-source parameterized invocation + grep count |
| AC-HOOK-005 | REQ-HOOK-005 | MUST-PASS | `git diff --name-only` filtered-list count |
| AC-HOOK-006 | REQ-HOOK-006 | MUST-PASS | `diff` between local and template fallback branches |
| AC-HOOK-007 | REQ-HOOK-007 | MUST-PASS | warning-text content grep (4 mandated elements) |

## §D.1 Severity

All 7 ACs are MUST-PASS. Tier S has no NICE-TO-HAVE ACs — the scope is minimal
and every requirement is load-bearing. Any single AC FAIL → run-phase returns
blocker report, no partial credit.

## §D.2 Given-When-Then Scenarios

### AC-HOOK-001 — Success-path byte-identical (traceable to REQ-HOOK-001)

**Given** the existing 3-tier moai-binary resolution chain in
`.claude/hooks/moai/handle-session-start.sh` (tier 1 PATH, tier 2
`$HOME/go/bin/moai`, tier 3 `$HOME/.local/bin/moai`), AND a test environment
where at least one tier resolves (e.g., `command -v moai` succeeds, OR
`$HOME/go/bin/moai` exists),

**When** SessionStart fires with `{"source":"startup"}` on stdin,

**Then** the wrapper `exec`s `moai hook session-start` with byte-identical
stdin/stdout/stderr behavior as the pre-SPEC baseline (verified by
characterization snapshot: same stdin forwarded, same `MOAI_HOOK_STDERR_LOG`
rotation logic, same `exec` semantics). The fallback branch is NOT reached.

**Verification command:**
```bash
# Snapshot the pre-SPEC wrapper's tier-1 branch behavior, then diff against
# post-SPEC. The diff MUST be empty.
diff <(git show HEAD~1:.claude/hooks/moai/handle-session-start.sh \
       | awk '/^# Try moai command in PATH/,/^# Try default/') \
     <(awk '/^# Try moai command in PATH/,/^# Try default/' \
           .claude/hooks/moai/handle-session-start.sh)
# Expected: empty diff (byte-identical tier-1/2/3 chain)
```

### AC-HOOK-002 — All-3-tiers-absent emits surfaced warning (traceable to REQ-HOOK-002)

**Given** a stubbed environment where `moai` is unresolvable on PATH AND
`$HOME/go/bin/moai` does not exist AND `$HOME/.local/bin/moai` does not exist
(e.g., `PATH=/usr/bin:/bin`, `HOME=$(mktemp -d)`),

**When** SessionStart fires with `{"source":"startup"}` on stdin,

**Then** the wrapper fallback:
1. emits a warning line to `$HOME/.moai/logs/hook-stderr.log` containing the
   substring `moai` AND at least one of {`PATH`, `go/bin`, `.local/bin`}, AND
2. emits to stdout a JSON object matching
   `{"hookSpecificOutput":{"additionalContext":"<warning text>"}}` where
   `<warning text>` is non-empty.

**Verification command:**
```bash
TMPHOME=$(mktemp -d)
OUT=$(PATH=/usr/bin:/bin HOME="$TMPHOME" \
      bash .claude/hooks/moai/handle-session-start.sh \
      <<< '{"source":"startup"}')
# (a) stderr log written
grep -q 'moai' "$TMPHOME/.moai/logs/hook-stderr.log" && echo "STDERR-OK"
# (b) stdout JSON carries additionalContext
echo "$OUT" | grep -q 'hookSpecificOutput' \
  && echo "$OUT" | grep -q 'additionalContext' \
  && echo "STDOUT-OK"
```

### AC-HOOK-003 — Non-blocking exit 0 (traceable to REQ-HOOK-003)

**Given** the all-3-tiers-absent state from AC-HOOK-002,

**When** the wrapper fallback completes,

**Then** the wrapper's exit code is exactly `0` (NOT 1, NOT 2). Session start
proceeds uninterrupted.

**Verification command:**
```bash
PATH=/usr/bin:/bin HOME=$(mktemp -d) \
  bash .claude/hooks/moai/handle-session-start.sh \
  <<< '{"source":"startup"}' >/dev/null 2>&1
echo "exit=$?"
# Expected: exit=0
```

### AC-HOOK-004 — Once-per-session dedup (source-gated) (traceable to REQ-HOOK-004)

**Given** the all-3-tiers-absent state from AC-HOOK-002,

**When** SessionStart fires 4 times with `source` ∈ {`startup`, `resume`,
`clear`, `compact`} (one invocation per matcher),

**Then** the warning is emitted exactly ONCE (on the `startup` invocation);
the `resume`, `clear`, and `compact` invocations emit NO warning (stdout empty
of `additionalContext`, stderr log empty of new warning line).

**Verification command:**
```bash
TMPHOME=$(mktemp -d)
for src in startup resume clear compact; do
  OUT=$(PATH=/usr/bin:/bin HOME="$TMPHOME" \
        bash .claude/hooks/moai/handle-session-start.sh \
        <<< "{\"source\":\"$src\"}")
  WARN_COUNT=$(echo "$OUT" | grep -c 'additionalContext')
  if [ "$src" = "startup" ]; then
    [ "$WARN_COUNT" -eq 1 ] && echo "$src: WARN-EMITTED (expected)"
  else
    [ "$WARN_COUNT" -eq 0 ] && echo "$src: SUPPRESSED (expected)"
  fi
done
# Expected:
#   startup: WARN-EMITTED (expected)
#   resume: SUPPRESSED (expected)
#   clear: SUPPRESSED (expected)
#   compact: SUPPRESSED (expected)
```

### AC-HOOK-005 — 30 non-SessionStart wrappers untouched (traceable to REQ-HOOK-005)

**Given** the run-phase commit for this SPEC,

**When** `git diff HEAD~1 --name-only -- .claude/hooks/moai/handle-*.sh` is
filtered to exclude `handle-session-start.sh`,

**Then** the filtered list is empty (zero of the 30 other wrappers were
modified).

**Verification command:**
```bash
git diff HEAD~1 --name-only -- '.claude/hooks/moai/handle-*.sh' \
  | grep -v '^\.claude/hooks/moai/handle-session-start\.sh$' \
  | wc -l
# Expected: 0
```

### AC-HOOK-006 — Template-First mirror parity (traceable to REQ-HOOK-006)

**Given** the run-phase commit for this SPEC,

**When** the fallback branch of the local `.claude/hooks/moai/handle-session-start.sh`
is compared to the fallback branch of the template
`internal/template/templates/.claude/hooks/moai/handle-session-start.sh.tmpl`,

**Then** the two fallback branches are byte-identical (the wrapper uses no
template variables per CLAUDE.local.md §14, so the rendered output equals the
template source verbatim).

**Verification command:**
```bash
diff \
  <(awk '/^# Not found/,/^exit 0/' .claude/hooks/moai/handle-session-start.sh) \
  <(awk '/^# Not found/,/^exit 0/' \
    internal/template/templates/.claude/hooks/moai/handle-session-start.sh.tmpl)
# Expected: empty diff (byte-identical fallback branches)
```

### AC-HOOK-007 — Warning content actionability (traceable to REQ-HOOK-007)

**Given** the warning emitted by AC-HOOK-002,

**When** the warning text is inspected (from either the stderr log or the
stdout `additionalContext`),

**Then** the text contains all 4 mandated elements:
- (a) tier enumeration: the substrings `PATH`, `go/bin`, AND `.local/bin` all
  appear (the 3 absent tiers named).
- (b) consequence: the substring `31 wrappers` (or grammatical equivalent)
  appears, naming the silent-no-op scope.
- (c) non-blocking framing: the substring `non-blocking` (or grammatical
  equivalent such as `advisory`) appears.
- (d) remediation: at least one of {`reinstall`, `restore PATH`, `rebuild`,
  `make build`} appears (an action hint).

**Verification command:**
```bash
TMPHOME=$(mktemp -d)
WARN=$(PATH=/usr/bin:/bin HOME="$TMPHOME" \
       bash .claude/hooks/moai/handle-session-start.sh \
       <<< '{"source":"startup"}' \
       | grep -oE '"additionalContext":"[^"]*"' \
       | sed 's/"additionalContext":"//;s/"$//')

# (a) tier enumeration
echo "$WARN" | grep -qF 'PATH' && echo "$WARN" | grep -qF 'go/bin' \
  && echo "$WARN" | grep -qF '.local/bin' && echo "(a) TIERS-OK"
# (b) consequence
echo "$WARN" | grep -qE '31 wrappers|thirty-one wrappers' && echo "(b) CONSEQUENCE-OK"
# (c) non-blocking framing
echo "$WARN" | grep -qE 'non-blocking|advisory' && echo "(c) FRAMING-OK"
# (d) remediation
echo "$WARN" | grep -qE 'reinstall|restore PATH|rebuild|make build' \
  && echo "(d) REMEDIATION-OK"
```

## §D.3 Indirect Verification

No AC relies on indirect verification. All 7 ACs are directly observable via
the verification commands above (run-phase manager-develop executes them and
captures verbatim output for §E.2 evidence).

## §D.4 Closure Gates

**Definition of Done (run-phase):**
- All 7 ACs GREEN (verification commands produce the expected output).
- `go test ./internal/hook/...` passes (session_start_test.go unchanged —
  regression baseline).
- `golangci-lint run --timeout=2m` clean (no new findings in the modified
  files; bash files are out of golangci-lint's scope but in pre-commit scope).
- Commit subject follows `fix(SPEC-HOOK-SESSIONSTART-PROBE-001): M1 ...` per
  the Status Transition Ownership Matrix.

**Definition of Done (sync-phase):**
- Single sync commit `docs(SPEC-HOOK-SESSIONSTART-PROBE-001): sync-phase artifacts`
  carries the `in-progress → implemented → completed` transition.
- 3-phase close (plan→run→sync) complete.

## §D.5 Forward-Looking Checks (post-close)

- **hook-independence.md §3 catalogue update** — the next time
  hook-independence.md is revised, the catalogue row A rationale may be updated
  to note "mitigation Recommendation implemented in
  SPEC-HOOK-SESSIONSTART-PROBE-001 (SessionStart probe)". This is a forward-
  looking doc concern, NOT an AC for this SPEC's run-phase.
- **Follow-up consideration** — if the probe proves noisy in practice (e.g.,
  users with deliberately-uninstalled moai find the startup warning annoying),
  a future SPEC may add a silence-via-env-var mechanism (e.g.,
  `MOAI_SESSIONSTART_PROBE_SILENCE=1`). Out of scope for this SPEC.

## §D.6 Severity Escalation

None — Tier S, all MUST-PASS, no severity tiers.

## §D.7 Edge Cases (manager-develop MUST cover)

- **Empty stdin / malformed JSON:** if stdin is empty or `source` cannot be
  extracted (malformed JSON), the fallback should emit warning unconditionally
  (treat as if `source=startup` — safer to warn than to suppress). Verify via
  `<<< ''` and `<<< '{malformed}'`.
- **`$HOME/.moai/logs/` unwritable:** the warning emission to the stderr log
  fails — wrapper MUST still emit stdout JSON and exit 0 (graceful degrade per
  §Design.8).
- **Stdout closed (e.g., redirected to `/dev/null` by Claude Code):** the
  stdout JSON emission fails — wrapper MUST still write stderr log and exit 0.
- **PATH contains a non-executable file named `moai`:** `command -v moai`
  succeeds but exec fails. This is a pre-existing edge case (the wrapper today
  has no special handling) and is out of scope — the probe is about the
  all-3-tiers-absent state, not the exec-failed state.

---

Version: 0.1.0
Status: draft (plan-phase authoring complete; plan-auditor verdict pending)
