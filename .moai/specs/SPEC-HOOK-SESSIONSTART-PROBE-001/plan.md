---
id: SPEC-HOOK-SESSIONSTART-PROBE-001
title: "SessionStart Probe — Implementation Plan"
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

# Plan — SPEC-HOOK-SESSIONSTART-PROBE-001

## Design — Branch Resolution (the design dilemma, resolved)

The prompt requires explicitly naming the chosen design branch among three
options (A/B/C), with rationale and rejection of the others.

### §Design.1 The constraint that drives the dilemma

The SessionStart Go handler (`internal/hook/session_start.go:45` `Handle()`)
only executes AFTER a wrapper successfully resolved the binary (the wrapper
`exec`s `moai hook session-start`). Therefore the Go handler CANNOT directly
observe the "all-3-tiers-absent" state — by the time it runs, a tier was found.
The probe logic MUST therefore live in a layer that runs BEFORE the Go handler:
either the wrapper itself (Branch A), a new Go subcommand invoked from the
wrapper fallback (Branch B), or a separate probe script (Branch C).

### §Design.2 Branch evaluation

| Criterion | Branch A (bash wrapper) | Branch B (Go probe + subcommand) | Branch C (separate probe script) |
|-----------|-------------------------|----------------------------------|----------------------------------|
| Minimal change | 1 wrapper, ~15 LOC | 1 wrapper + new Go subcommand + CLI wiring | 1 wrapper + new bash script |
| Additive-only | Yes | Yes | Yes |
| Success-path byte-identical | Yes (tiers 1/2/3 untouched) | Yes | Yes |
| Once-per-session dedup | source-gate in bash | Go state file | bash state file |
| Non-blocking | Yes | Yes | Yes |
| Self-contained (§7 Q1) | Yes — does NOT inherit mode A | **Paradox — see below** | Yes — does NOT inherit mode A |
| New file inventory | None | Go subcommand + CLI test | New bash script + chmod |

### §Design.3 Branch B rejected — structural paradox

Branch B proposes: add `moai hook session-start-probe`, invoked from the
wrapper's fallback path before `exit 0`. But the fallback only fires when moai
is unresolvable in ALL 3 tiers — which means invoking `moai hook
session-start-probe` from that state would ALSO fail (moai is absent by
definition of reaching the fallback). Branch B does not actually surface a
warning in the precise condition that triggers it; the Go subcommand is
unreachable when needed. **Branch B is structurally unsound and is rejected.**

### §Design.4 Branch C rejected — needless moving parts

Branch C adds a new self-contained bash probe script (e.g.,
`.claude/hooks/moai/sessionstart-probe.sh`) invoked from the wrapper fallback.
Compared to Branch A, this adds: (1) a new file in the hook layer, (2) a new
chmod concern, (3) a second site to maintain when the warning content evolves,
(4) a new entry in the hook-independence.md §3 catalogue (a new shared
condition?). The benefit (separation of concerns) does not justify the
multi-surface cost for a Tier S single-pass SPEC. **Branch C is rejected as
over-engineered for the minimal scope.**

### §Design.5 Branch A selected

**Selected: Branch A (bash wrapper).** Modify ONLY the fallback branch of
`handle-session-start.sh` (and its template mirror `handle-session-start.sh.tmpl`)
to emit a non-blocking warning before the existing `exit 0`. The tier-1/2/3
resolution chain is untouched; the success path is byte-identical; the warning
emission is self-contained bash (does NOT call moai, does NOT inherit mode A).

**Reconciliation with hook-independence.md §6 "wrappers NOT to be edited":**
the §6 clause is the constraint of THAT audit-and-doctrine deliverable ("the
deliverable inspects and classifies; it does not change hook behavior"). The §6
Recommendation itself explicitly defers the probe to "a separate follow-up —
adding the probe is a hook/SessionStart change, not part of this
audit-and-doctrine deliverable". This SPEC is that follow-up. The scope of the
edit is named precisely: exactly the SessionStart wrapper's fallback branch,
not the 30 others (REQ-HOOK-005).

### §Design.6 Dedup mechanism — source-gated emission

SessionStart fires on 4 matchers: `startup | resume | clear | compact`. Within
a single long-running session, `compact` can fire many times — without dedup,
the warning would spam on every compaction.

**Chosen: source-gated emission.** Emit the warning only when `source ==
"startup"` (the very first SessionStart of a brand-new session). Subsequent
`resume`/`clear`/`compact` events within the same session are suppressed.

Rationale:
- `startup` fires exactly once per session lifetime (truly new Claude Code
  session). resume/clear/compact are within-session events.
- No state file, no marker cleanup, no time-window logic — idempotent and
  stateless.
- Bash parsing without `jq`: the wrapper layer does NOT depend on `jq`
  (hook-independence.md §3 row D — only 2 of 3 governance gates use jq). The
  probe reads stdin and extracts `source` via a simple grep/sed pipeline.

**stdin lifecycle correctness:** in the success path, the wrapper `exec`s moai
BEFORE touching stdin (stdin is forwarded to moai via the exec). In the fallback
branch (reached only when all 3 tiers are absent), the wrapper reads stdin via
`$(cat)` to extract `source`. The two paths are mutually exclusive — no stdin
double-consumption.

**Design limitation — resume coverage (accepted for Tier S simplicity):** the
probe gates emission on `source=startup` only (REQ-HOOK-004). This creates a
narrow coverage gap: if the moai binary is removed **MID-SESSION** (after
startup succeeded with the binary present) and the session is later resumed
(`source=resume`), then at the original `startup` event the fallback branch
was NOT reached (binary was present, success path taken, no warning), and at
the subsequent `resume` event the fallback branch IS reached (binary now
absent) but warning emission is suppressed by REQ-HOOK-004's `source=startup`
gate. The net effect: a session that started healthy but lost its moai binary
mid-session receives NO surfaced warning on resume. This gap is **accepted**
under Tier S simplicity (stateless source-gating — no marker file, no state
tracking, no cross-event session identity). Resume-coverage is deferred to a
forward-looking follow-up SPEC (backlog candidate placeholder ID:
`SPEC-HOOK-SESSIONSTART-PROBE-RESUME-001` — NOT created in this SPEC; listed
here as a backlog pointer only).

### §Design.7 Warning surface — dual-channel (primary model-context, secondary audit-log)

The warning is emitted on two channels simultaneously, with **distinct roles**
(primary vs secondary — these are NOT two equivalent user-facing surfaces):

1. **Primary channel — stdout JSON `hookSpecificOutput.additionalContext`**:
   SessionStart's stdout field, surfaced to the **model's context** via
   injection (per `.claude/rules/moai/core/hooks-system.md` § Hook Event
   stdin/stdout Reference). The model typically relays this to the user in its
   next response, but relay is **not deterministic** — the model may also act
   on the warning without surfacing it verbatim, or the user may not notice the
   relay. The word "surfaced" throughout this SPEC denotes this model-context
   injection (the primary actionable signal), NOT a guaranteed user-visible
   render in the session UI.

2. **Secondary channel — `$HOME/.moai/logs/hook-stderr.log`**: the existing
   sink already plumbed via `MOAI_HOOK_STDERR_LOG` env var at the top of the
   wrapper. This is an **after-the-fact audit trail** the user can grep — it is
   NOT a real-time user surface. It exists for forensic / debugging purposes
   when the primary channel's model-relay did not occur or was not noticed,
   and its value is post-hoc discoverability, not session-start alerting.

Both emissions are non-blocking; `exit 0` is preserved as the terminal exit
code (REQ-HOOK-003). The two channels are intentionally redundant: the primary
channel is what makes the signal actionable at session start (model context),
while the secondary channel guarantees post-hoc discoverability independent of
model behavior. Conflating the two (e.g., claiming the stderr log is "surfaced
to the user") would over-state the design's capability — the stderr log is
greppable, not surfaced.

### §Design.8 Graceful degradation of the probe itself

If the warning emission itself fails (e.g., `$HOME/.moai/logs/` unwritable,
stdout closed), the wrapper MUST still `exit 0` (hook-independence.md §7 Q4 —
graceful degrade). The probe's failure mode degrades to "no warning" rather
than "session halt". Concretely: every emission step is best-effort
(`|| true`, `2>/dev/null || true`); the terminal `exit 0` is unconditional.

## §A Context

See spec.md §A for the problem statement, verified evidence, and classification
citations. This plan operates against hook-independence.md §3 row A, §5, §6
Recommendation — no new risk discovery is performed here.

## §B Known Issues

None — greenfield additive change. No prior attempts at this probe exist in the
hook layer (verified: `grep -rn 'sessionstart-probe\|session-start-probe'
.claude/hooks/moai/ internal/template/templates/.claude/hooks/moai/` → empty).

## §C Pre-flight (before M1 implementation)

- [ ] Confirm SPEC ID `SPEC-HOOK-SESSIONSTART-PROBE-001` is free in
      `.moai/specs/` (done at plan-phase; re-check at run-phase start).
- [ ] Confirm plan-auditor verdict is PASS or PASS-WITH-DEBT (Phase 0.5 gate).
- [ ] Confirm Implementation Kickoff Approval obtained (CLAUDE.local.md §19.1
      plan→run HUMAN GATE; Phase 0.5 SKIP ≠ run-phase entry authorization).
- [ ] Read current `.claude/hooks/moai/handle-session-start.sh` (33 lines,
      1234 bytes) and its template mirror side-by-side; confirm they are
      byte-identical modulo the `.tmpl` suffix.

## §D Constraints

- **Tier S** — minimal, single-pass, additive-only. See §F for the single M1
  milestone (no multi-milestone split).
- **No new shared failure mode** — the probe is self-contained bash; it adds no
  new shared dependency (§7 Q2). It does NOT call moai, does NOT depend on jq,
  does NOT introduce a new shared config file.
- **Template-First** — both local `.sh` and template `.sh.tmpl` edited in the
  same commit (REQ-HOOK-006).
- **Settings.json untouched** — SessionStart registration (matcher
  `startup|resume|clear|compact`, timeout 30s) is NOT modified.

## §E Self-Verification (run-phase)

The run-phase (manager-develop, cycle_type=tdd) self-verifies via the §E.2
evidence matrix in progress.md. The verification commands (canonical batch):

```bash
# 1. AC-HOOK-005: 30 non-SessionStart wrappers untouched
git diff HEAD~1 --name-only -- .claude/hooks/moai/handle-*.sh \
  | grep -v 'handle-session-start.sh' | wc -l   # STRICT 0

# 2. AC-HOOK-006: Template-First mirror parity
diff <(awk '/^# Not found/,/^exit 0/' .claude/hooks/moai/handle-session-start.sh) \
     <(awk '/^# Not found/,/^exit 0/' internal/template/templates/.claude/hooks/moai/handle-session-start.sh.tmpl) \
  && echo "PARITY OK"

# 3. AC-HOOK-002 + AC-HOOK-003: fallback emits warning + exit 0
PATH=/usr/bin:/bin HOME=$(mktemp -d) bash .claude/hooks/moai/handle-session-start.sh \
  <<< '{"source":"startup"}' ; echo "exit=$?"

# 4. AC-HOOK-004: source-gated dedup (resume/clear/compact suppressed)
for src in resume clear compact; do
  PATH=/usr/bin:/bin HOME=$(mktemp -d) bash .claude/hooks/moai/handle-session-start.sh \
    <<< "{\"source\":\"$src\"}" | grep -c 'moai'   # STRICT 0 for each
done

# 5. Success-path regression (any tier resolves → byte-identical)
# Requires moai on PATH or in $HOME/go/bin; characterize with snapshot test.

# 6. Lint + test
go test ./internal/hook/...   # session_start_test.go unchanged; regression
golangci-lint run --timeout=2m
```

## §F Milestones (single-pass — Tier S)

**M1 — SessionStart probe implementation** (single milestone).

Scope:
- Edit `.claude/hooks/moai/handle-session-start.sh` fallback branch.
- Edit `internal/template/templates/.claude/hooks/moai/handle-session-start.sh.tmpl`
  fallback branch (byte-identical content modulo template-variable semantics —
  but this wrapper uses no template variables, so the content is verbatim
  identical).
- Add characterization tests:
  - `.claude/hooks/moai/test_sessionstart_probe.bats` (or equivalent bash test
    harness) — covers AC-002/003/004/007.
  - Success-path regression test — covers AC-001.
- Verify AC-005 (grep invariant: 30 wrappers untouched).
- Verify AC-006 (diff parity: local == template).

Exit criteria: all 7 ACs GREEN; `go test ./internal/hook/...` passes
(unchanged); `golangci-lint` clean.

**No M2, M3, ...** — Tier S single-pass.

## §G Anti-Patterns (what to avoid)

- **AP-1: Editing the tier-1/2/3 resolution chain.** The chain is untouched;
  only the fallback branch (after all 3 fail) is modified. Any change to the
  tier checks themselves violates REQ-HOOK-001 (success-path byte-identical).

- **AP-2: Adding `exit 2` or `set -e` in the fallback.** The fallback MUST
  exit 0 (REQ-HOOK-003). The probe is advisory; blocking session start would
  contradict hook-independence.md §7 Q4 (graceful degrade) and the
  Recovery-Signal Carve-Out principle (`agent-common-protocol.md` § Recovery-
  Signal Carve-Out — SessionStart is a non-blocking event).

- **AP-3: Calling `moai ...` from the fallback.** The fallback exists because
  moai is unresolvable in all 3 tiers. Calling moai from the fallback is the
  Branch B paradox (§Design.3). The probe is pure bash — no moai invocation.

- **AP-4: Using `jq` to parse stdin.** The wrapper layer does NOT depend on jq
  (hook-independence.md §3 row D — distinct from the governance gates).
  Introducing jq here would add a new shared dependency to the wrapper layer,
  inflating §3's catalogue. Use grep/sed.

- **AP-5: Editing the 30 non-SessionStart wrappers.** The scope is
  SessionStart-localized (REQ-HOOK-005). Any change to other handle-*.sh files
  is unrequested scope creep.

- **AP-6: Over-stating the risk in the warning content.** The warning MUST
  state the precise trigger (all 3 tiers absent) — NOT "moai not in PATH"
  (which would be the first-tier-only over-statement forbidden by
  verification-claim-integrity §1.1 surface 3 and hook-independence.md §5
  "Precise trigger for mode A").

- **AP-7: Skipping Template-First mirror.** Editing only the local `.sh`
  without the template `.sh.tmpl` violates REQ-HOOK-006 + CLAUDE.local.md §2
  [HARD]. Both must be in the same commit.

- **AP-8: Time-based dedup marker file.** A mtime-window marker file adds a
  state-management surface (cleanup, race on concurrent sessions) that
  source-gating avoids. Stick to source=startup gate (§Design.6).

## §H Cross-References

- spec.md — requirements (REQ-HOOK-001..007) + acceptance traceability.
- acceptance.md — Given-When-Then ACs (AC-HOOK-001..007).
- `.claude/rules/moai/development/hook-independence.md` §3 row A, §5, §6
  Recommendation, §7 Q1/Q4/Q5 — design constraints.
- `CLAUDE.local.md` §2 Template-First Rule, §14 `.HomeDir`/`.GoBinPath`
  prohibition, §19.1 Implementation Kickoff Approval.
- `.claude/rules/moai/core/verification-claim-integrity.md` §1.1 surface 3 —
  binds the genuine-risk trigger wording (no over-statement).
- `.claude/rules/moai/core/hooks-system.md` § stdin/stdout Reference
  (SessionStart `hookSpecificOutput.additionalContext`), § Timeout Configuration
  (SessionStart 30s).

---

Version: 0.1.0
Status: draft (plan-phase authoring complete; plan-auditor verdict pending)
