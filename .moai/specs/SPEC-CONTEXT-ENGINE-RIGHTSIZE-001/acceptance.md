# acceptance.md — SPEC-CONTEXT-ENGINE-RIGHTSIZE-001

> Acceptance Criteria matrix. Every AC is falsifiable (grep-countable or observable command output). No AC relies on prose judgment. Per `feedback_guard_observation_must_be_falsifiable`, A-group preservation ACs use grep counts over text-pattern claims; per `feedback_claimed_correction_never_applied`, every "done" claim cites the verbatim grep output at the time of verification.

---

## §A. AC Design Principles

### §A.1 Falsifiability rules (codified per applied lessons)

- **A1 — grep observable**: every "preserved" / "removed" / "added" AC cites a `grep -c` count or a `grep -L` empty-listing. Prose-only claims ("the rule still exists") are not acceptable.
- **A2 — verbatim evidence**: every AC PASS row cites the verbatim grep output line at verification time (per `verification-claim-integrity.md §1.1 surface 2 + §2 baseline-attribution).
- **A3 — vacuous-green trap**: avoid patterns from `feedback_ac_command_vacuous_green_traps` — guard regexes are not re-implemented in the AC (the guard is the authority); AC greps use scoped literal patterns, not BRE alternation constructs (the escaped-pipe form is forbidden because BSD grep parses it as a literal pipe and yields 0 matches — see D2 fix).
- **A4 — defect-claim discipline**: per `feedback_defect_claim_verification`, no AC is authored for a non-existent defect (M1.3 — no defect exists; only a regression-guard AC is included).
- **A5 — no ghost AC**: per `feedback_claimed_correction_never_applied`, an AC is included ONLY for a change the SPEC actually makes. M1.3 gets a "SSOT reference survived" AC (regression guard), NOT a "fixed" AC.

### §A.2 Severity convention

| Severity | Meaning |
|---|---|
| **Critical** | Blocks run-phase completion. A-group preservation or C-group preservation violation. |
| **High** | Blocks the SPEC's core constructive change (M1/M2 transitions). |
| **Medium** | Operational requirement (template sync, lint parity). |
| **Low** | Cosmetic / documentation only. |

---

## §B. Acceptance Criteria Matrix

### AC-CER-001 — M1 code_comments expressive transition (High)

**Given** the `moai-constitution.md` TRUST 5 Readable line at ~line 77,
**When** the M1 milestone completes,
**Then** the unconditional "English comments" absolute is replaced by config-respecting judgement language.

Falsifiable checks:
```bash
# (a) Old form gone
grep -c 'Clear naming, English comments' .claude/rules/moai/core/moai-constitution.md
# Expected: 0

# (b) New form present — references config mechanism + match-surrounding-code language
grep -c 'match the surrounding code.*language and density' .claude/rules/moai/core/moai-constitution.md
# Expected: 1

# (c) Config mechanism explicitly named (deference to code_comments setting, not absolute)
grep -c 'code_comments.*language.yaml' .claude/rules/moai/core/moai-constitution.md
# Expected: 1
```

**Traceability**: REQ-CER-003.

---

### AC-CER-002 — M2 Tool Selection consolidation (5-bullet block removed) (High)

**Given** the `moai-constitution.md` § Tool Selection Priority block (~lines 106-117),
**When** the M2 milestone completes,
**Then** the 5-bullet "Use X instead of Y" list is replaced by a single-line informational SSOT pointer.

Falsifiable checks:
```bash
# (a) 5-bullet absolute block gone
grep -c '^- Use .* instead of' .claude/rules/moai/core/moai-constitution.md
# Expected: 0

# (b) Single-line SSOT pointer present
grep -c 'agent-common-protocol.md.*Tool Selection by Task' .claude/rules/moai/core/moai-constitution.md
# Expected: 1

# (c) Judgement-delegating language present (prefer / fit-for-purpose)
grep -Ec 'prefer the dedicated tool|fit for purpose' .claude/rules/moai/core/moai-constitution.md
# Expected: >= 1
```

**Traceability**: REQ-CER-001, REQ-CER-004.

---

### AC-CER-003 — Canonical SSOT retained (M1.1 second half) (High)

**Given** `agent-common-protocol.md § Tool Selection by Task` is the canonical SSOT,
**When** M2 completes,
**Then** that table is unchanged in semantic intent.

Falsifiable checks:
```bash
grep -c '^### Tool Selection by Task' .claude/rules/moai/core/agent-common-protocol.md
# Expected: 1 (heading preserved)

# Spot-check the table is intact (Task | Preferred Tool | Avoid structure)
grep -c '^| Find files by name | Glob |' .claude/rules/moai/core/agent-common-protocol.md
# Expected: 1
```

**Traceability**: REQ-CER-002.

---

### AC-CER-004 — M1.3 plan-auditor SSOT reference preserved (regression guard, Medium)

**Given** `plan-auditor.md:~144` already delegates tool guidance to the canonical SSOT (verified 2026-07-28, no defect existed pre-edit),
**When** M2 completes,
**Then** that SSOT cross-reference is still observable (the M2 consolidation did not orphan the plan-auditor's pointer).

Falsifiable check:
```bash
grep -c 'agent-common-protocol.md.*Tool Selection by Task' .claude/agents/moai/plan-auditor.md
# Expected: >= 1
```

**Rationale (per `feedback_defect_claim_verification`)**: this is a regression-guard AC, NOT a fix AC. The original handoff flagged M1.3 as a defect, but direct verification showed the SSOT ref was already in place. Per `feedback_claimed_correction_never_applied`, no ghost "fixed" AC is authored.

**Traceability**: REQ-CER-009.

---

### AC-CER-005 — A-group `[ZONE:Frozen]` count preserved (Critical)

**Given** the 2026-07-28 verified baseline of 66 `[ZONE:Frozen]` markers across 13 files in `.claude/rules/moai/`,
**When** all milestones (M1-M4) complete,
**Then** the Frozen count remains greater than or equal to 66 (no Frozen marker was removed or downgraded to Evolvable).

Falsifiable checks:
```bash
# (a) Total count across .claude/rules/moai/ — necessary but NOT sufficient
grep -rc '\[ZONE:Frozen\]' .claude/rules/moai/ | awk -F: '{s+=$2} END{print s}'
# Expected: >= 66

# (b) Per-file distribution diff — falsifies the "removal-in-one-file offset by
#     addition-elsewhere" bypass that (a) cannot catch (per plan-auditor D1).
#     Baseline file is captured at M1 start per plan.md §C Baseline 1b.
diff -u \
  .moai/specs/SPEC-CONTEXT-ENGINE-RIGHTSIZE-001/baseline-frozen-distribution.txt \
  <(grep -rc '\[ZONE:Frozen\]' .claude/rules/moai/ | grep -v ':0$' | sort)
# Expected: empty diff (no `<` / `>` lines). A `<` line (file count dropped) is
#           a regression even when (a) still passes. A `>` line (file count rose)
#           is also a regression (an addition elsewhere should not silently mask it).
```

**Traceability**: REQ-CER-005, NFR-CER-001.

---

### AC-CER-006 — A-group AskUserQuestion doctrines preserved (Critical)

**Given** AskUserQuestion-Only Interaction + Subagent Prohibitions + Deferred Tool Preload are the 3 anchor doctrines in `.claude/rules/moai/core/`,
**When** all milestones complete,
**Then** all three remain observable at their canonical paths.

Falsifiable checks:
```bash
# (a) AskUserQuestion-Only — anchor rule
grep -Ec 'AskUserQuestion is the only user-facing question channel|AskUserQuestion-Only' .claude/rules/moai/core/askuser-protocol.md
# Expected: >= 1

# (b) Subagent Prohibitions — HARD (the canonical Frozen HARD line at
#     agent-common-protocol.md:~15; pattern corrected 2026-07-29 after the D2
#     ERE-conversion review surfaced that the original pattern matched a
#     non-existent phrase — "Subagents MUST NOT invoke/prompt" was the actual
#     wording, not the longer pattern authored in v0.1.0)
grep -c '\[HARD\] Subagents MUST NOT prompt the user' .claude/rules/moai/core/agent-common-protocol.md
# Expected: >= 1

# (c) Deferred Tool Preload — ToolSearch contract
grep -c 'ToolSearch(query: "select:AskUserQuestion")' .claude/rules/moai/core/askuser-protocol.md
# Expected: >= 1
```

**Traceability**: REQ-CER-006, NFR-CER-002.

---

### AC-CER-007 — A-group safety invariants preserved (Critical)

**Given** Branch Guard + Verification-Claim Integrity + Native `/goal` Prohibition are the 3 safety invariants,
**When** all milestones complete,
**Then** all three remain observable at their canonical paths.

Falsifiable checks:
```bash
# (a) Branch Guard — BRANCH_GUARD_VIOLATION sentinel
grep -c 'BRANCH_GUARD_VIOLATION' .claude/rules/moai/workflow/main-checkout-branch-guard.md
# Expected: >= 1

# (b) Verification-Claim Integrity — §1.1 surfaces 1-3
grep -Ec 'no unobserved-claim|1\.1 Binding scope' .claude/rules/moai/core/verification-claim-integrity.md
# Expected: >= 1

# (c) Native /goal Prohibition
grep -Ec 'Native.*/.*goal.*Prohibition|## Native `/goal` Prohibition' .claude/rules/moai/workflow/goal-directive.md
# Expected: >= 1
```

**Traceability**: REQ-CER-007, NFR-CER-002.

---

### AC-CER-008 — C-group mechanical guardrails preserved (Critical)

**Given** Multi-File Decomposition (`CLAUDE.md §7 Rule 2`) + Reproduction-First Bug Fix (`CLAUDE.md §7 Rule 4`) are explicitly preserved as C-group mechanical guardrails (NOT transitioned),
**When** all milestones complete,
**Then** both rules remain in force at their verified CLAUDE.md locations with their `[HARD]` bullets intact.

Falsifiable checks:
```bash
# (a) Multi-File Decomposition — §7 Rule 2 + HARD bullet
grep -Ec 'Multi-File Change Decomposition|Multi-File Decomposition' CLAUDE.md
# Expected: >= 2 (HARD bullet at line ~19 + §7 Rule 2 at line ~151)

# (b) Reproduction-First Bug Fix — §7 Rule 4 + HARD bullet
grep -c 'Reproduction-First Bug Fix' CLAUDE.md
# Expected: >= 2 (HARD bullet at line ~19 + §7 Rule 4 at line ~157)

# (c) Both still carry the HARD marker
grep -Ec '\[HARD\] Multi-File Decomposition|\[HARD\] Reproduction-First Bug Fix' CLAUDE.md
# Expected: >= 2
```

**Rationale (C-group classification, per `feedback_guard_signal_proves_call_not_effect`)**: these rules are mechanical guardrails against orchestrator scope-creep and bug-fix confirmation-bias. Their enforcement survives even when B-group stylistic rules move to judgement-delegation. The "보수(B-group only)" decision explicitly preserves them.

**Traceability**: REQ-CER-008.

---

### AC-CER-009 — Template mirror synced + §25 neutralized (Medium)

**Given** the template-first obligation (CLAUDE.local.md §2 [HARD]) + §25 template-internal-isolation doctrine,
**When** M3 completes,
**Then** the template mirror `internal/template/templates/.claude/rules/moai/core/moai-constitution.md` carries the same M1+M2 edits, AND no SPEC ID / REQ token / audit citation leaked into the template.

Falsifiable checks:
```bash
# (a) M1 edit mirrored
grep -c 'match the surrounding code.*language and density' internal/template/templates/.claude/rules/moai/core/moai-constitution.md
# Expected: 1

# (b) M2 edit mirrored
grep -c 'agent-common-protocol.md.*Tool Selection by Task' internal/template/templates/.claude/rules/moai/core/moai-constitution.md
# Expected: 1

# (c) §25 — NO SPEC ID leaked
grep -rc 'SPEC-CONTEXT-ENGINE-RIGHTSIZE-001' internal/template/templates/.claude/rules/moai/core/
# Expected: 0

# (d) §25 — NO REQ token leaked
grep -rc 'REQ-CER-' internal/template/templates/.claude/rules/moai/core/
# Expected: 0

# (e) §25 — NO audit citation leaked
grep -Erc 'Audit.*Finding|per SPEC-CONTEXT' internal/template/templates/.claude/rules/moai/core/
# Expected: 0 matches across every file (every per-file count line reads ":0")
```

**Traceability**: REQ-CER-010, NFR-CER-003.

---

### AC-CER-010 — No behavioral regression (lint + test parity) (Medium)

**Given** this SPEC edits only `.claude/rules/moai/*.md` (no Go code) and template mirrors,
**When** M4 completes,
**Then** `moai spec lint` reports no new findings attributable to M1/M2 edits, AND `go test ./...` remains green.

Falsifiable checks:
```bash
# (a) moai spec lint — compare to §C Baseline 6
moai spec lint --json 2>/dev/null | tail -20 || go run ./cmd/moai spec lint 2>&1 | tail -20
# Expected: no new findings referencing moai-constitution.md lines ~77 or ~106-117
# (repo-wide pre-existing debt is out of scope per §F Token-budget sweep)

# (b) go test parity — no Go code touched, suite green
go test ./... 2>&1 | tail -10
# Expected: "ok" lines; no new FAIL
```

**Baseline-attribution** (per `verification-claim-integrity.md §2`): the lint result MUST be attributed to the post-M4 tree state, not a carry-over from a prior unrelated run. The §C Baseline 6 capture at M1 start is the comparison reference.

**Traceability**: REQ-CER-011.

---

## §C. Severity Breakdown

| Severity | Count | AC IDs |
|---|---|---|
| Critical | 4 | AC-CER-005, AC-CER-006, AC-CER-007, AC-CER-008 (A-group + C-group preservation) |
| High | 3 | AC-CER-001, AC-CER-002, AC-CER-003 (M1/M2 constructive transitions) |
| Medium | 3 | AC-CER-004, AC-CER-009, AC-CER-010 (regression guard + template sync + lint parity) |
| Low | 0 | — |

**Pass criterion**: all 10 ACs PASS at M4 completion. A single Critical failure blocks run-phase completion per the implementation-kickoff approval gate.

---

## §D. Indirect Verification (not authored as AC — context only)

The following are observable side effects of a successful M1/M2 but are NOT authored as ACs because they are either (a) downstream consequences of an AC-CER row that already falsifies the claim, or (b) too coarse to be falsifiable:

- **Tool Selection prose length reduction in `moai-constitution.md`** — the section shrinks from ~12 lines to ~3 lines. Observable via `wc -l` on the region, but AC-CER-002 already falsifies the transition via grep.
- **Reduced MUST count in `moai-constitution.md`** — M2 removes 5 "Use X instead of Y" lines that implicitly carried MUST semantics. Observable via `grep -c '\bMUST\b'` delta, but AC-CER-002 already falsifies the removal.
- **Stylistic "voice" consistency** — judgement-delegating language should feel consistent across the new prose. This is a stylistic judgment, not falsifiable; intentionally NOT an AC.

---

## §E. Closure Gate (Definition of Done)

A SPEC is "done" when ALL of the following hold at M4 completion:

1. **All 10 ACs PASS** with verbatim grep output cited per row (no prose-only claims).
2. **All 4 milestones committed** with the canonical commit-subject prefix (`feat(SPEC-CONTEXT-ENGINE-RIGHTSIZE-001): M{n} ...` for M1-M3, `chore(SPEC-CONTEXT-ENGINE-RIGHTSIZE-001): M4 ...` for M4).
3. **Pathspec-only staging** — every commit was staged via explicit pathspec, NOT `git add -A`.
4. **No C-group scope creep** — Multi-File Decomposition and Reproduction-First Bug Fix remain `[HARD]` (AC-CER-008 PASS).
5. **No A-group regression** — `[ZONE:Frozen]` count >= 66 (AC-CER-005 PASS) and all 5 anchor doctrines observable (AC-CER-006 + AC-CER-007 PASS).
6. **Template sync + §25 neutralization verified** (AC-CER-009 PASS).
7. **Lint parity verified** (AC-CER-010 PASS).

---

## §F. Forward-Looking Checks (post-close observations, NOT closure gates)

These are NOT required for SPEC closure but SHOULD be observed in the sync-phase / next-few-SPECs window:

- **F1 — Agent adoption**: do downstream agent definitions (`.claude/agents/**/*.md`) update their code-comment expectations to match the new "match surrounding code" language? If they still hard-code "English comments", a follow-up SPEC may be needed to propagate the transition.
- **F2 — C-group candidate review**: if future Anthropic guidance further weakens absolute rules, re-evaluate whether Multi-File Decomposition / Reproduction-First Bug Fix remain C-group or move to B-group. (Out of scope for this SPEC; explicitly deferred.)
- **F3 — Token-budget sweep**: a separate follow-up SPEC could reduce the 269 `MUST` / 14 `NEVER` count repo-wide. (Out of scope per §F of spec.md.)

---
