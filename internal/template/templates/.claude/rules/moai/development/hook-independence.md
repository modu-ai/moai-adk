---
paths: "**/.claude/hooks/**"
---

# Hook Independence — Shared-Failure-Mode Catalogue + Authoring Checklist

Doctrine for the MoAI hook layer (`.claude/hooks/moai/`). This rule names every
shared failure mode across the hook scripts, classifies each as
`acceptable-by-design` or `genuine-risk`, and gives hook authors a forward-looking
checklist so a new hook does not silently inherit every existing shared mode.

This is **policy-layer doctrine**, not mechanical enforcement. There is no lint
rule or CI gate behind it — it changes hook-author behavior by being read, the
same way the verification-claim-integrity doctrine is read. Mechanical
enforcement (a meta-hook that scans new hook scripts for shared-condition
introduction) is intentionally out of scope; deciding "is this degradation
catastrophic?" requires author judgment, not a mechanically-decidable property.

## 1. Why hook independence matters

Defense-in-depth is the design pattern of layering independent checks so that one
layer's failure does not collapse the whole defense. The pattern has a sharp
failure mode of its own: **defense-in-depth fails when its layers share a failure
mode.** A defense assembled from components that all collapse under one condition
is not depth — it is one layer wearing many hats.

The MoAI hook layer is exactly such a defense surface. It is built from two
families of scripts:

- **wrapper scripts** (`handle-*.sh`) — thin forwarders that delegate a Claude
  Code hook event to the `moai hook <event>` subcommand.
- **governance gates** (`status-transition-ownership.sh`,
  `sync-phase-quality-gate.sh`, `team-ac-verify.sh`) — self-contained bash that
  enforce a policy (ownership transitions, sync quality, per-AC verification).

A shared failure mode in either family correlates the degradation of many scripts
under one condition. This is not hypothetical: `CLAUDE.md` §17 Troubleshooting
already documents a real incident in this family — a `moai hook` invocation
failing because the `moai` binary was not resolvable on PATH. That is the
shared-failure-mode hazard made concrete, and it is the motivation for this
catalogue.

## 2. Definition: "shared failure mode"

A **shared failure mode** is any single dependency, condition, flag, or convention
whose failure correlates the degradation of **two or more** hook scripts.

The unit of analysis is the *correlation*, not the dependency in isolation:

- A dependency that only one script has is **not** a shared mode.
- A dependency that ten scripts each have, but whose failure degrades each one
  *independently* (not simultaneously), is a **weak** shared mode.
- The **strongest** shared modes are those where one condition flips many scripts
  at once — the wrapper layer's moai-binary chain (one condition, many wrappers)
  and the governance layer's bypass flag (one flag, all gates).

## 3. The shared-failure-mode catalogue

Each row cites the command run and the output observed (reproducible by any
reviewer against the current tree), per the no-unobserved-claim invariant. The
counts below are the observed-at-audit-time values; re-running the cited command
re-confirms them.

| Mode | Scope | Evidence (command → observed) | Classification | Rationale |
|------|-------|-------------------------------|----------------|-----------|
| **A — moai-binary resolution chain** | 31 wrapper scripts | `ls handle-*.sh \| wc -l` → 31; `grep -l 'command -v moai' handle-*.sh \| wc -l` → 31; `for f in handle-*.sh; do grep -q 'command -v moai' "$f" \|\| echo "$f"; done` → (empty, i.e. 0 wrappers lack the chain) | **genuine-risk** | All 31 wrappers carry the *identical* 3-tier resolution chain and degrade together if `moai` is unresolvable in all 3 tiers. The degradation is **silent by design** (final `exit 0`, "Claude Code handles missing hooks gracefully") and a documented real incident exists (`CLAUDE.md` §17). Low-likelihood but high-correlation + silent → genuine-risk. |
| **B — `--skip-hook` bypass** | 3 governance gates | `grep -l 'skip-hook' *.sh` → the 3 gate filenames only (wrappers do NOT match); each gate checks `if [ "$1" = "--skip-hook" ]` and `exit 0` | **acceptable-by-design** | One flag as `$1` disables all 3 gates — but the bypass is a *documented, audit-logged operator override*. Each gate appends a skip record to the shared `.moai/logs/hook-skip.log` before exiting (see `agent-common-protocol.md` § Hook Invocation Surface). Chesterton's Fence: an intentional, logged opt-out is not a defect. |
| **C — shared `hook-skip.log`** | 3 governance gates | every gate writes to `${CLAUDE_PROJECT_DIR:-$PWD}/.moai/logs/hook-skip.log` on bypass (same path literal in all 3) | **acceptable-by-design** | A single shared skip ledger IS the intended design — one place to audit every bypass. A shared *sink* (not a shared *gate*) is benign; its failure mode (log path unwritable) degrades only the audit trail, not the gate logic. |
| **D — `jq` dependency** | 2 of 3 gates (NOT sync-gate) | `grep -c jq status-transition-ownership.sh sync-phase-quality-gate.sh team-ac-verify.sh` → status:5 / **sync:0** / team:5 | **acceptable-by-design** | The 2 jq-dependent gates degrade gracefully (no-op / allow-all) when `jq` is absent — bounded, not correlated-catastrophic. The exception is load-bearing: `sync-phase-quality-gate.sh` uses `git`/`grep`/`awk` and does NOT depend on `jq`. The layer is NOT homogeneous; a "all gates depend on jq" claim would be false. |
| **E — `${CLAUDE_PROJECT_DIR:-$PWD}` root-fallback** | 3 gates + wrappers | `grep -l 'CLAUDE_PROJECT_DIR' *.sh` → gates + wrappers share the convention | **acceptable-by-design** | Standard project-root resolution convention. Failure mode is benign: when `CLAUDE_PROJECT_DIR` is unset, the scripts fall back to `$PWD` rather than crashing. A shared *convention* with a benign fallback is not a shared *failure* trigger. |
| **F — `set -e` convention** | 3 gates | `grep -l 'set -e' status-transition-ownership.sh sync-phase-quality-gate.sh team-ac-verify.sh` → all 3 | **acceptable-by-design** | A shared safety convention (exit on unhandled error), not a shared failure trigger. It makes each gate fail *closed* on an unexpected error; that is the desired behavior, not a correlated-degradation hazard. |

## 4. Governance-gate independence cross-tab

REQ-coverage detail: for each governance gate, whether it shares each enumerated
dependency. The key reading is **row (a)** (the positive signal) and **row (d)**
(the `jq` exception — do NOT homogenize the layer).

| Shared dependency | status-transition-ownership.sh | sync-phase-quality-gate.sh | team-ac-verify.sh |
|-------------------|:------------------------------:|:--------------------------:|:-----------------:|
| (a) moai-PATH resolution chain (`command -v moai`) | **NO** (self-contained bash) | **NO** | **NO** |
| (b) `--skip-hook` bypass (`$1`) | YES | YES | YES |
| (c) shared skip-log `.moai/logs/hook-skip.log` | YES | YES | YES |
| (d) `jq` dependency | YES (graceful degrade) | **NO** (git/grep/awk only) | YES (graceful degrade) |
| (e) `${CLAUDE_PROJECT_DIR:-$PWD}` root-fallback | YES | YES | YES |
| (f) `set -e` convention | YES | YES | YES |
| (g) configured timeout | 5s (PostToolUse) | 60s (Stop) | 10s (TaskCompleted) — **dormant / not wired** |

> **Row (g) caveat — `team-ac-verify.sh` is dormant.** The TaskCompleted gate is
> not registered in the live hook configuration: `grep -c 'team-ac-verify'
> .claude/settings.json` → 0. It is forward-looking (activates only under harness
> `thorough` + team mode prerequisites). Its 10s timeout is the configured value
> *should it be wired*, not an active timeout today. Do not read row (g) as
> evidence the gate is active.

### 4.1 Positive signal — the governance layer does NOT share mode A

Row (a) is the load-bearing positive finding. The 3 governance gates do **NOT**
share the moai-binary resolution chain — the strongest shared failure of the
wrapper layer. The governance gates are self-contained bash depending on `git`,
`jq` (2 of 3), `grep`, and `awk`, never on the `moai` binary. So when `moai`
leaves all 3 resolution tiers and the entire 31-wrapper layer goes silent, the
governance gates still function (subject to their own `git`/`jq` availability).

This is genuine defense depth that the wrapper layer's internal uniformity lacks:
the wrapper layer's strongest shared failure does **not** propagate to the
governance layer. A defense whose two families share their strongest failure mode
would be depth-in-name-only; here they do not, which is the property this audit
set out to verify.

## 5. The wrapper resolution chain (mode A detail)

The `handle-*.sh` wrappers do not gate on a bare `command -v moai`; they carry a
**3-tier resolution chain** followed by a silent fallback. The chain, identical
across all 31 wrappers (only the `hook <event>` token differs):

1. `command -v moai` on PATH → `exec moai hook <event>`
2. else `$HOME/go/bin/moai` (default Go install location) → `exec` it
3. else `$HOME/.local/bin/moai` (Linux install location) → `exec` it
4. else `exit 0` — silent no-op ("Claude Code handles missing hooks gracefully")

**Precise trigger for mode A.** The correlated-degradation trigger is the
**conjunction of all three tiers being absent** — `moai` unresolvable on PATH
AND in `$HOME/go/bin` AND in `$HOME/.local/bin`. It is NOT "moai not in PATH"
(the first tier alone): the 2nd and 3rd tiers are a real fallback. Stating the
trigger as the first-tier check alone would over-state the risk — and a risk
claim may not be over-stated any more than a success claim may be.

**The wrappers already carry a fallback.** A mitigation of "add a fallback when
`moai` is absent" is *partially already satisfied* — the wrappers fall back across
3 binary locations and degrade silently (not crash). The residual genuine-risk is
narrow: the **all-three-absent** case, where all 31 wrappers go silent
simultaneously with no surfaced warning.

## 6. Mitigation recommendations

For each `genuine-risk` mode, a mitigation *recommendation* (not an
implementation). Recommendations only — no hook script is edited by this doctrine.

- **Mode A (genuine-risk) — Recommendation**: consider a one-time SessionStart
  probe that warns (once, non-blocking) when `moai` is unresolvable in all 3
  tiers, converting the silent-simultaneous-degradation into a surfaced signal.
  This is a *recommendation*, deferred to a separate follow-up — adding the probe
  is a hook/SessionStart change, not part of this audit-and-doctrine deliverable.
  The wrappers themselves are NOT to be edited (the deliverable inspects and
  classifies; it does not change hook behavior).

All other catalogue modes are `acceptable-by-design` and therefore carry no
mitigation obligation (their shared correlation is intentional and bounded).

## 7. Authoring checklist — "does this new hook introduce an independent failure mode?"

Before adding a new hook, answer each question. The goal is to keep a new hook
from silently inheriting every existing shared mode (a fourth gate that shares all
of B/C/D/E/F, or a new wrapper that shares mode A, adds correlation without depth).

1. **Does this hook depend on the `moai` binary?** If yes, does it carry the
   3-tier resolution chain + silent `exit 0` fallback (like the wrappers), or is
   it self-contained bash (like the governance gates)? Self-contained is more
   independent — it does not inherit mode A.
2. **Does this hook add a *new* shared condition** that, if it fails, would
   degrade other hooks too — a new shared config file, a new shared binary, a new
   shared env var? If yes, name it and classify it (`acceptable-by-design` vs
   `genuine-risk`) before merging.
3. **If this hook shares the `--skip-hook` bypass**, is the bypass logged to
   `.moai/logs/hook-skip.log` (the audit invariant that keeps the bypass in the
   `acceptable-by-design` class)? An unlogged bypass is a defect.
4. **Does this hook degrade gracefully** (no-op / allow / logged) when its
   dependencies are absent, or does it crash/block? Graceful degradation keeps a
   shared dependency in the `acceptable-by-design` class; a crash/block on a
   shared dependency pushes it toward `genuine-risk`.
5. **Is the degradation surfaced** when the shared condition fails, or silent? A
   silent simultaneous degradation across many hooks is the genuine-risk shape —
   consider a one-time probe/warning rather than a silent `exit 0`.

A hook that answers "self-contained / no new shared condition / logged bypass /
graceful / surfaced" introduces no new shared failure mode. A hook that answers
otherwise should record the new mode in the catalogue above (§3) with its
classification, so the next author inherits an accurate picture.

## 8. Cross-references (SSOT — cross-reference, do not duplicate)

This doctrine cross-references the following canonical surfaces. It does NOT copy
their content — each remains the single source of truth for its own subject:

- `.claude/rules/moai/core/hooks-system.md` — canonical hook event + execution-type
  reference (the wrapper / gate event vocabulary, timeout policy, the
  PostToolUse / Stop / TaskCompleted event semantics). This catalogue names *which*
  scripts share *which* dependency; hooks-system.md owns the event model itself.
- `.claude/rules/moai/core/agent-common-protocol.md` § Hook Invocation Surface —
  the 3 governance gates' owning REQs, exit-code semantics (exit 0 = continue,
  exit 2 = block/reject), and the `--skip-hook` audit-log convention (mode B's
  by-design justification). The Stop self-gate caveat and the orchestrator's
  exit-2 → AskUserQuestion translation responsibility also live there.
- `.claude/rules/moai/workflow/runtime-recovery-doctrine.md` §4 — the
  Recovery-Signal Carve-Out (Stop / PostToolUse hooks SHOULD exit 0 on recovery
  turns to avoid the death-spiral). A related but separate hook concern; this
  catalogue cross-references it for completeness. The carve-out is documentation-
  only at this layer (current hooks do not parse `stopReason`).

---

Version: 1.0.0
Classification: Canonical Reference (policy-layer doctrine) — do not duplicate the
cross-referenced surfaces; cross-reference this file's catalogue instead.
