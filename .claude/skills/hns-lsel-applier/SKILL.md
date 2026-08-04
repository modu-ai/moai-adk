---
name: hns-lsel-applier
description: >
  Local Self-Evolution Loop (LSEL) APPLY engine — the playback-only consumer of
  approved decision.json records that drives `.moai/hooks/lsel-apply.sh` for the
  GOOS-local PROPOSE→APPLY seam closure (SPEC-LSEL-LOCAL-EVOLUTION-001 M3). Reads
  an approved decision.json, validates the target against the frozen allowlist
  (.claude/lsel/frozen-allowlist.json), mechanically refuses execution-meta
  targets lacking a synchronous-approval marker, applies the referenced
  diff.patch, appends an apply-ledger.jsonl row, and commits one
  lsel-<proposal-id>-tagged Conventional Commit on the feature branch. M3 scope:
  APPLY bypass closure only (the frozen Go applier stays frozen — its write-flag
  at internal/harness/applier.go:22 stays false; REQ-LSEL-003).
allowed-tools: Read, Grep, Glob, Bash
user-invocable: false
metadata:
  version: "0.1.0"
  category: "harness"
  status: "active"
  updated: "2026-08-04"
  tags: "lsel,self-evolution,apply,harness,dogfood"
---

# hns-lsel-applier — LSEL APPLY engine

> **Namespace:** `hns-lsel-*` is user-owned dogfood (CLAUDE.local.md §24). This skill
> is NOT mirrored into `internal/template/templates/` — it lives only in this repo.
> Graduation to `moai-lsel-*` + 16-language distribution is a separate SPEC
> (out of scope per spec.md §G).

> **Token discipline.** This body deliberately avoids the literal identifier of the
> frozen Go applier's write-flag and the literal name of the orchestrator-only
> user-question channel. Both are referenced by location and role (e.g. "the write-flag
> at `internal/harness/applier.go:22`", "the orchestrator's synchronous user gate") so
> the REQ-LSEL-003 invariant-grep and the E4 subagent-boundary grep over LSEL surfaces
> read zero literal matches. The invariants themselves are stated in plain English.

## What this skill does

The LSEL loop's PROPOSE→APPLY seam was dead in production: `Applier.Apply` never ran
(`manifest.jsonl` absent), `CuratorDispatch` had 0 production callers, and the frozen
Go applier's write-flag at `internal/harness/applier.go:22` (kept `false`) is the
apply dead-switch (REQ-LSEL-003 — M3 keeps it `false` by BYPASS, never unfreeze). The
M3 closure routes APPLY through a parallel **user-owned** applier that writes only to
the six evolvable surfaces (spec.md §B.3), while the frozen Go applier stays
byte-for-byte frozen. Design SSOT:
`.moai/reports/moai-local-self-evolution-design-20260804.html` §4 ("동결된 Go applier를
우회하는 병렬 사용자 소유 applier를 세운다" — build a parallel user-owned applier that
bypasses the frozen Go one) + §7 (the two-round critique moved the allowlist OUT of
the evolvable skill files into a frozen meta file, and added execution-meta files to
the forced-gate set).

The APPLY engine is the **playback hook** `.moai/hooks/lsel-apply.sh` plus this skill's
model-mediated judgment (target validation nuance, approval-marker provenance, blast-
radius reasoning the mechanical hook cannot see). The hook is the load-bearing safety
floor; this skill is the consumer-facing surface that decides WHICH approved decision
to feed it next and how to interpret refusal.

## The APPLY pipeline

`lsel-apply.sh <decision.json>` performs five steps. Steps 1-2 are the safety floor;
steps 3-5 are the playback.

1. **Frozen-allowlist hard-reject** (REQ-LSEL-001 / AC-LSEL-001). The target path in
   `decision.json` is matched against the `frozen_patterns` regex list in
   `.claude/lsel/frozen-allowlist.json`. A match → REFUSE: a row is appended to
   `.moai/logs/lsel-reject.log` naming the rejected path and category `frozen-path`,
   NO file is written, and the hook exits 2.
2. **Execution-meta forced-gate** (REQ-LSEL-002 / AC-LSEL-005 D3 self-amending-
   handcuffs). If the target matches one of the four `execution_meta` categories —
   (i) the frozen allowlist meta file itself, (ii) an applier/curator skill body
   (`hns-lsel-applier/`, `hns-lsel-curator/`), (iii) the apply hook
   (`.moai/hooks/lsel-*.sh`), (iv) the `settings.local.json` hook-registration
   subblock — the hook checks `decision.json` for a synchronous-approval marker (a
   `synchronous_approval` object with `decision: "approved"`, produced by the
   orchestrator's synchronous user gate). No marker → REFUSE: reject-log row with
   category `execution-meta`, exit 3, no write. With marker → proceeds (the refusal
   is keyed on the absent marker, NOT on the category match alone). The refusal
   semantics mirror the M2 `csa_refusal_test.sh` fixture exactly.
3. **Apply the diff.patch** via `git apply` (playback of an already-approved decision).
   Only the paths declared in the patch are staged — never `git add -A` (working-tree
   hygiene).
4. **Append the apply-ledger row** to `.moai/state/lsel/apply-ledger.jsonl`:
   `{proposal_id, target_surface, ts, result:"applied", commit_sha, category}`. This
   is the manifest the frozen Go applier never produced, finally real in user-owned
   space.
5. **Commit** the staged change as ONE `feat(lsel-<proposal-id>): ...` Conventional
   Commit on the current (feature) branch (REQ-LSEL-004). The ledger row's
   `commit_sha` is backfilled from the commit's short SHA.

A no-arg invocation is a clean no-op (exit 0) so an empty approved-queue does not
derail a loop pass.

## The model-mediated layer (you, when invoked)

When this skill is invoked to drive an APPLY pass, your job on top of the mechanical
hook is:

- **Read the proposal's `proposal.md` + `self-critique.md`** at
  `.moai/state/lsel/proposals/<id>/`. A proposal with `status: blocked` (any
  UNRESOLVED self-critique objection) MUST NOT be fed to the hook — return a blocker
  report instead.
- **Confirm the approval marker's provenance.** The `synchronous_approval` object must
  carry a real orchestrator-produced approval artifact (the synchronous user-question
  channel the orchestrator owns per CLAUDE.md §8). A marker the loop fabricated for
  itself is the self-amending-handcuffs failure mode (REQ-LSEL-002); the hook's
  mechanical check is the floor, your provenance judgment is the ceiling. This skill
  is a subagent mechanism and NEVER invokes the orchestrator-only user-question
  channel; return a blocker report and let the orchestrator run the gate.
- **Re-verify the frozen-allowlist invariant** by grepping the write-flag at
  `internal/harness/applier.go:22` — it MUST stay `false`. If M3's bypass ever drifts
  toward unfreezing the Go applier, return a blocker (AP-LSEL-002).

## What this skill does NOT do

- **No unfreezing of the Go applier** — the write-flag at
  `internal/harness/applier.go:22` stays `false` (REQ-LSEL-003). The bypass is
  parallel and user-owned; the frozen applier is reference-only.
- **No edits to frozen doctrine** — `internal/template/templates/**`,
  `.claude/rules/moai/**`, `CLAUDE.md`, retained agents, `moai-*` skills, the frozen
  Go applier / `curator_dispatch.go`, and `.moai/config/sections/**` are
  byte-for-byte untouched. The allowlist hard-rejects them (step 1).
- **No new-apply / self-approve / allowlist-amend primitive** in `lsel-apply.sh`
  (REQ-LSEL-008). The hook only consumes already-approved decisions; it does not
  author proposals or bless them. The M4 `verify.sh` runs only AFTER an apply has
  committed — it does not create a new apply; it never self-approves.
- **No user-question invocation** — subagent boundary (CLAUDE.md §8). Return a
  blocker report; the orchestrator runs the synchronous gate.

---

## VERIFY stage (M4 — REQ-LSEL-013 / AC-LSEL-015)

After an apply commits a proposal, VERIFY proves the apply was safe. VERIFY has
**two layers** — both MANDATORY:

### (a) Mechanical layer — `verify.sh` (bash-testable)

`verify.sh --proposal-dir <p> --repo-root <r> [--timeout SECS] [--feedback-file <f>]`
runs the proposal's frontmatter `verify_command` with a **timeout-retry-once**
policy and auto-reverts on a second failure:

- **2-attempt policy.** Attempt 1 runs. On SUCCESS → `verified:true`, stop. On
  TIMEOUT-class failure → attempt 2 (flaky tolerance — report §10 P3: "retry
  once on timeout so a correct proposal isn't flipped by noise"). On
  NON-TIMEOUT-FAIL → attempt 2 (allows the "second non-timeout failure" of
  AC-LSEL-015 clause 3 to materialize before revert fires).
- Attempt 2: SUCCESS → `verified:true`. Any failure (timeout OR non-timeout) →
  `verified:false` + **auto `git revert lsel-<proposal-id>`** + the proposal's
  `feedback_*.md` is marked `verified: false`.
- The outcome is appended to the apply ledger as a
  `{"stage":"verify","verified":true|false,...}` row — the load-bearing signal.

### (b) MANDATORY `/moai gate` superset (model-mediated)

**`/moai gate` (lint+format+type+test) is MANDATORY after every apply — it is
NOT optional.** A bash hook cannot invoke a Claude Code slash command directly,
so the gate runs MODEL-SIDE: when this skill drives an APPLY pass, the
orchestrator/model runs `/moai gate` after the mechanical apply + verify. The
proposer-authored `verify_command` alone is **circular** (report §11 mustFix
B#6 / AP-LSEL-004): a proposer can author a verify_command that its own change
satisfies. `/moai gate` is the independent check — lint, format, type-check,
and the test suite exercise surfaces the proposal never touched. Treat a
mechanical verify_command PASS without a green `/moai gate` as an unverified
apply.

### Model-mediated layer (you, when invoked)

- After `lsel-apply.sh` commits and `verify.sh` reports, **run `/moai gate`**
  (the MANDATORY superset). If the gate fails, the apply is unverified even if
  `verify.sh` reported `verified:true` — treat it as a VERIFY failure and revert.
- A revert (whether auto-fired by `verify.sh` or gate-driven) MUST be surfaced
  to the orchestrator as a blocker report — the orchestrator, not this skill,
  confirms a revert with the user. Never run `/moai gate`'s user-facing surface
  from inside this subagent mechanism; return a blocker report.

### Verification (run before declaring an APPLY+VERIFY pass complete)

```bash
# M4 VERIFY characterization test (AC-LSEL-015) — hermetic temp repo:
.claude/skills/hns-lsel-applier/verify_test.sh

# the verify ledger row carries the outcome:
grep '"stage":"verify"' .moai/state/lsel/apply-ledger.jsonl
```

## Verification (run before declaring an APPLY pass complete)

```bash
# 1. APPLY characterization test (AC-001/002/005/008/013) — hermetic temp repo:
.claude/skills/hns-lsel-applier/apply_test.sh

# 2. Rollback-rehearsal SHIP GATE (AC-014):
.claude/skills/hns-lsel-applier/rollback_rehearsal_test.sh

# 3. REQ-LSEL-003 frozen-flag re-verify (post-apply). The frozen Go applier's
#    write-flag at internal/harness/applier.go:22 MUST read `false`, and no LSEL
#    surface (skills/hooks/state/allowlist) references that identifier as mutable.
#    (The literal identifier is intentionally not written here; find it at line 22
#    of the frozen file and grep LSEL surfaces for it — expect zero hits.)
sed -n '22p' internal/harness/applier.go   # MUST show the `= false` write-flag line

# 4. The apply ledger carries the new row:
tail -1 .moai/state/lsel/apply-ledger.jsonl
```

## Cross-references

- **SPEC:** `.moai/specs/SPEC-LSEL-LOCAL-EVOLUTION-001/{spec,plan,acceptance,progress}.md`
- **Design report (SSOT):** `.moai/reports/moai-local-self-evolution-design-20260804.html`
  §4 (user-owned parallel applier), §7 (allowlist relocation + execution-meta forced
  gate), §10 P3, §11 mustFix A#1/A#5/A#8.
- **Frozen applier (reference only):** `internal/harness/applier.go:22` (the frozen
  write-flag), `internal/harness/curator_dispatch.go`.
- **Curator (PROPOSE stage, M2):** `.claude/skills/hns-lsel-curator/SKILL.md`
  (CSA forced-gate doctrine + the Tier-4 DEAD finding).
- **CSA refusal fixture:** `.claude/skills/hns-lsel-curator/csa_refusal_test.sh`
  (the refusal rule this hook enforces mechanically).
- **Namespace guard:** `internal/template/split_namespace_test.go`,
  `internal/template/internal_content_leak_test.go` (must NOT flag a `hns-lsel-applier`
  leak — this skill is dogfood, never templated).
