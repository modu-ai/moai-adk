# Research — SPEC-DIVECC-HOOK-FAILURE-MODE-AUDIT-001

> Plan-phase grounding for the N1 hook shared-failure-mode audit. This file records the **verbatim** read-only evidence gathered against the moai-adk tree, per `.claude/rules/moai/core/verification-claim-integrity.md` §1.1 surface 3 + §2 (baseline attribution: the command run + the output observed, in this run, against this tree).

---

## §A. Verbatim shared-failure-mode evidence (the VERIFIED premise)

### A.1 Hook-layer inventory

**Command run** (in `.claude/hooks/moai/`):
```
ls handle-*.sh | wc -l
```
**Observed output**: `31` (31 `handle-*.sh` wrapper scripts).

Plus the 3 governance gates (`status-transition-ownership.sh`, `sync-phase-quality-gate.sh`, `team-ac-verify.sh`) → 34 hook scripts total in `.claude/hooks/moai/`.

### A.2 Shared failure mode A — moai-binary resolution chain

**Command run**:
```
grep -l 'command -v moai' *.sh | wc -l       # → 31
grep -l 'command -v moai' *.sh               # → exactly the 31 handle-*.sh wrappers
for f in handle-*.sh; do grep -q 'command -v moai' "$f" || echo "$f"; done   # → (empty)
```
**Observed output**: all 31 wrappers match `command -v moai`; the "wrappers WITHOUT the gate" loop printed nothing → **0 wrappers lack the chain**. The 3 governance gates do NOT appear in the `grep -l 'command -v moai'` result.

**Full read of a representative wrapper** (`handle-stop.sh`) — the resolution chain verbatim:
```bash
# Try moai command in PATH
if command -v moai &> /dev/null; then
    exec moai hook stop 2>>"$MOAI_HOOK_STDERR_LOG"
fi
# Try default ~/go/bin/moai
if [ -f "$HOME/go/bin/moai" ]; then
    exec "$HOME/go/bin/moai" hook stop 2>>"$MOAI_HOOK_STDERR_LOG"
fi
# Try ~/.local/bin/moai (Linux install location)
if [ -f "$HOME/.local/bin/moai" ]; then
    exec "$HOME/.local/bin/moai" hook stop 2>>"$MOAI_HOOK_STDERR_LOG"
fi
# Not found - exit silently (Claude Code handles missing hooks gracefully)
exit 0
```
A second wrapper (`handle-session-start.sh`) was read fully and carries the **identical** chain (only the `hook <event>` token differs). The chain is a generated, uniform pattern across all 31 wrappers.

**Conclusion**: shared failure mode A is a 3-tier resolution chain (PATH → `$HOME/go/bin` → `$HOME/.local/bin` → silent `exit 0`), identical across all 31 wrappers. The correlated-degradation trigger is the **conjunction** of all three tiers being absent, NOT the first `command -v moai` tier alone.

> **Refinement record (evidence-grounded honesty).** The orchestrator's grounding stated mode A as "a single condition — the `moai` binary present in PATH". The full read of the wrappers shows a 3-tier fallback chain, so the precise trigger is "moai absent from PATH AND `$HOME/go/bin` AND `$HOME/.local/bin`". The shared-failure-mode conclusion is unchanged (all 31 share the identical chain and degrade together), but the trigger condition is narrowed from "not in PATH" to "not resolvable by any of the 3 tiers". Per `verification-claim-integrity.md`, a risk claim may not be over-stated any more than a success claim may be.

**Corroborating documented real failure** (`CLAUDE.md` §17 Troubleshooting):
> `moai hook subagent-stop` fails | Binary not in PATH | Run `which moai`

This is the paper's "shared failure mode" warning made concrete: a real, already-documented incident where the shared moai-binary dependency degraded a hook.

### A.3 Shared failure mode B — `--skip-hook` bypass (governance gates)

**Command run**:
```
grep -l 'skip-hook' *.sh
```
**Observed output**:
```
status-transition-ownership.sh
sync-phase-quality-gate.sh
team-ac-verify.sh
```
All 3 governance gates (and only those 3) carry the `--skip-hook` bypass. The wrappers do NOT.

**Full read — the bypass is identical across all 3 gates** (verbatim from each gate's top):
```bash
# Opt-out flag
if [ "$1" = "--skip-hook" ]; then
    echo "{\"skipped\": true, \"reason\": \"--skip-hook flag\"}" >&2
    mkdir -p "${CLAUDE_PROJECT_DIR:-$PWD}/.moai/logs"
    echo "$(date -u +%Y-%m-%dT%H:%M:%SZ) [<gate-name>] skipped via --skip-hook" \
        >> "${CLAUDE_PROJECT_DIR:-$PWD}/.moai/logs/hook-skip.log"
    exit 0
fi
```
(Line refs at read time: `status-transition-ownership.sh` lines 10-16; `sync-phase-quality-gate.sh` lines 57-63; `team-ac-verify.sh` lines 27-33.)

**Conclusion**: one flag (`--skip-hook` as `$1`) disables all 3 gates. This is by-design (per `agent-common-protocol.md`, the `--skip-hook` opt-out is logged to `.moai/logs/hook-skip.log` for audit). It is a **documented shared-bypass surface**, classified `acceptable-by-design` — but it IS a single condition that correlates all three gates' degradation, so it belongs in the catalogue.

---

## §B. Full governance-gate independence cross-tab (the §A audit detail)

All 3 governance gates were read in full. The cross-tab below records which shared dependencies each gate carries — this is the data REQ-DIVECC-004 requires the audit to produce.

| Shared dependency | status-transition-ownership.sh | sync-phase-quality-gate.sh | team-ac-verify.sh |
|---|---|---|---|
| (a) moai-PATH resolution chain (`command -v moai`) | **NO** (self-contained bash) | **NO** | **NO** |
| (b) `--skip-hook` bypass (`$1`) | YES | YES | YES |
| (c) shared skip-log file `.moai/logs/hook-skip.log` | YES | YES | YES |
| (d) `jq` dependency | YES (graceful degrade → exit 0 / allow) | NO (uses git/grep/awk only) | YES (graceful degrade → allow-all) |
| (e) `${CLAUDE_PROJECT_DIR:-$PWD}` root fallback | YES | YES | YES |
| (f) `set -e` convention | YES (line 7) | YES (line 21) | YES (line 24) |
| (g) audit-log write (own log file) | YES (`status-transition-audit.log`) | YES (`sync-quality-gate.log`) | YES (`team-ac-verify.log`) |
| (h) configured timeout | 5s (PostToolUse) | 10s (Stop) | 10s (TaskCompleted, dormant) |
| (i) advisory/warn-first (does not hard-block by default) | YES (always exit 0; exit 2 reserved) | YES (exit 2 only if `MOAI_SYNC_GATE_BLOCKING=1`) | exit 2 on `--reject` stub / TaskCompleted reject |

### B.1 Positive / independent signal (REQ-DIVECC-006)

Row (a) is the key positive finding: the 3 governance gates do **NOT** share the moai-binary resolution chain (mode A) — the strongest shared failure of the wrapper layer. The governance layer depends on `git`, `jq`, `grep`, `awk` instead of the `moai` binary. So when `moai` leaves all 3 resolution tiers and the entire wrapper layer goes silent, the governance gates still function (subject to their own `jq`/`git` availability). This is genuine defense depth that the wrapper layer's internal uniformity lacks.

### B.2 Notable per-gate observations

- **status-transition-ownership.sh** — advisory only; always exits 0 (lines 66-79 note that exit 2 is "reserved for future ownership-mismatch enforcement"). Depends on `jq` (graceful no-op if absent, line 19-22).
- **sync-phase-quality-gate.sh** — the only gate that does NOT depend on `jq`; it uses `git` + `grep` + `awk`. Warn-first: a `block` decision is logged but the hook exits 0 unless `MOAI_SYNC_GATE_BLOCKING=1` (lines 222-230). Source-able for unit testing via a `BASH_SOURCE` guard (lines 68-71).
- **team-ac-verify.sh** — dormant unless `workflow.yaml` `team.enabled: true`; carries a `--reject` stub emitting `ledger_note` (REQ-LEDGER-002 of a sibling SPEC). Depends on `jq` for active verification (graceful allow-all if absent, lines 79-82).

---

## §C. Wrapper-layer fallback-branch finding (REQ-DIVECC-005)

The orchestrator's grounding asked whether the `handle-*.sh` wrappers already carry an `else` fallback when `moai` is absent. **They do** — but not as an `else`; as the 2nd and 3rd tiers of the resolution chain (`$HOME/go/bin/moai`, then `$HOME/.local/bin/moai`), then a final silent `exit 0` with the comment "Claude Code handles missing hooks gracefully". So:

- A mitigation recommendation of "add a fallback when moai is absent" is **partially already satisfied** — the wrappers fall back across 3 binary locations and degrade silently (not crash).
- The residual genuine-risk is the **all-three-absent** case: when moai is in none of the 3 locations, all 31 wrappers go silent simultaneously and the entire moai-hook delegation layer is a no-op with no surfaced warning (the `exit 0` is silent by design). Whether this silent-simultaneous-degradation warrants a louder signal (e.g. a one-time SessionStart probe that warns if `moai` is unresolvable) is the candidate mitigation recommendation for the genuine-risk classification — recorded as a *recommendation*, deferred to a follow-up SPEC for implementation per REQ-DIVECC-012.

---

## §D. Classification input (feeds REQ-DIVECC-003)

| Shared failure mode | Scope | Proposed classification | Rationale seed |
|---|---|---|---|
| A — moai-binary resolution chain | 31 wrappers | **genuine-risk** (low-likelihood, high-correlation) | All 31 degrade together if moai unresolvable in all 3 tiers; silent by design; documented real incident exists (CLAUDE.md §17). Mitigation = louder signal, not rewrite. |
| B — `--skip-hook` bypass | 3 gates | **acceptable-by-design** | Documented, audit-logged opt-out (`hook-skip.log`); intentional operator override; Chesterton's Fence applies. |
| C — shared `hook-skip.log` file | 3 gates | **acceptable-by-design** | A shared audit sink is the intended design (single skip ledger). |
| D — `jq` dependency | 2 of 3 gates | **acceptable-by-design** | Both jq-dependent gates degrade gracefully (no-op / allow-all) when jq absent — degradation is bounded, not correlated-catastrophic. |
| E — `${CLAUDE_PROJECT_DIR:-$PWD}` fallback | 3 gates + wrappers | **acceptable-by-design** | Standard project-root resolution convention; failure mode is benign (falls back to CWD). |
| F — `set -e` convention | 3 gates | **acceptable-by-design** | Shared safety convention, not a shared failure trigger. |

> The classifications above are the **proposed** plan-phase seeds. The run-phase audit confirms or revises each with its stated rationale (REQ-DIVECC-003 requires the final classification + rationale to live in the audit output / `hook-independence.md`).

---

## §E. Method note

All evidence in this file was produced by read-only inspection (`ls`, `grep -l`, `grep -q` loop, and full `Read` of `handle-stop.sh`, `handle-session-start.sh`, and all 3 governance gates) at plan-phase, against the working tree at `/Users/goos/MoAI/moai-adk-go`. No hook script was modified. The run-phase audit re-runs these commands to re-confirm the evidence is current before writing the catalogue (the evidence is a baseline measurement, not a carry-over — re-attribution at run-phase per `verification-claim-integrity.md` §2).
