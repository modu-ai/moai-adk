# Acceptance Criteria — SPEC-STEERING-ALIGN-CLAUDEMD-DIET-001

> GEARS-format acceptance criteria with re-runnable verification commands.
> Every AC names: the claim (GEARS), the command, and the PASS predicate.
> SSOT for AC enumeration is this file (NOT progress.md).

---

## D. Acceptance Criteria Matrix

### AC-CMD-001 — Per-tree line-count drops into the derived range

**GEARS (State-driven)**: **While** the diet is in effect, the always-loaded CLAUDE.md line count SHALL drop from 650 into the derived target range (400-470, plan.md §C.4 — revised UP after the D1 demotions) in BOTH trees, by the SAME amount.

**Given** the baseline 650/650 lines, **When** the diet is applied to both trees, **Then** each tree's `wc -l` lands in [400, 470] AND the two counts are equal.

```bash
L=$(wc -l < CLAUDE.md)
T=$(wc -l < internal/template/templates/CLAUDE.md)
echo "LIVE=$L TEMPLATE=$T"
# PASS: 400 <= L <= 470  AND  L == T  (the L==T equality also feeds AC-CMD-002 byte-parity; here it is the line-level cross-check, D6 note)
[ "$L" -ge 400 ] && [ "$L" -le 470 ] && [ "$L" -eq "$T" ] && echo "AC-CMD-001 PASS" || echo "AC-CMD-001 FAIL"
```

PASS predicate: both in range AND equal. The range was revised UP from iter-1's [350,430] because the D1 re-audit demoted 5 genuinely-unique sections from POINTER → KEEP (less content removed = higher floor). (If the honest classification lands slightly outside [400,470], run-phase updates the derived range in plan.md §C.4 with justification rather than over-cutting — REQ-CMD-009.) (D6: the `L==T` line-equality here is the line-level cross-check; AC-CMD-002's `diff` is the authoritative byte-parity gate — the two are complementary, not redundant.)

---

### AC-CMD-002 — Template ↔ live byte-parity preserved

**GEARS (Ubiquitous)**: The two CLAUDE.md trees SHALL remain byte-identical after the diet.

**Given** byte-identical baseline (diff exit 0), **When** the diet is applied to both, **Then** `diff` still exits 0.

```bash
diff CLAUDE.md internal/template/templates/CLAUDE.md; echo "DIFF_EXIT=$?"
# PASS: DIFF_EXIT=0
```

PASS predicate: `DIFF_EXIT=0` (no output).

---

### AC-CMD-003 — Load-bearing [HARD]-directive count preserved (≥ 14)

**GEARS (Unwanted behavior)**: The diet SHALL NOT reduce the `[HARD]` directive count below the protected baseline of 14, unless a dropped line is proven a verbatim SSOT duplicate (enumerated + justified in the run-phase report).

**Given** 14 `[HARD]` lines at baseline, **When** the diet is applied, **Then** `grep -c '[HARD]'` is ≥ 14 in both trees — OR any shortfall is itemized with a duplicate-of-SSOT justification.

```bash
HL=$(grep -c '\[HARD\]' CLAUDE.md)
HT=$(grep -c '\[HARD\]' internal/template/templates/CLAUDE.md)
ZL=$(grep -c '\[ZONE:' CLAUDE.md)
echo "HARD live=$HL template=$HT ; ZONE live=$ZL (baseline HARD=14 ZONE=14)"
# PASS: HL >= 14 AND HL == HT   (any drop below 14 requires enumerated justification in run report)
[ "$HL" -ge 14 ] && [ "$HL" -eq "$HT" ] && echo "AC-CMD-003 PASS" || echo "AC-CMD-003 REVIEW (justify drop)"
```

PASS predicate: live `[HARD]` count ≥ 14 AND equal across trees. A justified drop (verbatim SSOT duplicate) converts REVIEW→PASS with the run-phase enumeration.

---

### AC-CMD-004 — @import not double-counted as token reduction

**GEARS (Unwanted behavior)**: The 2 `@import` lines SHALL be accounted as structure-only; no `@import` restructuring SHALL be counted toward the token-reduction figure.

**Given** 2 `@import` lines at baseline, **When** the diet is reported, **Then** the `@import` count is unchanged (still 2, structure-only) AND the AC-CMD-006 reduction figure is attributed solely to M-DELETE/M-POINTER/M-SCOPE (run-phase report states this attribution explicitly).

```bash
I=$(grep -cE '^@[.]' CLAUDE.md)
echo "import_lines=$I (baseline 2; structure-only, excluded from token-reduction accounting)"
# PASS: I == 2  AND run-phase report attributes the byte reduction to M-DELETE/M-POINTER (NOT @import)
[ "$I" -eq 2 ] && echo "AC-CMD-004 PASS (verify attribution prose in run report)" || echo "AC-CMD-004 FAIL"
```

PASS predicate: `import_lines=2` unchanged AND the run-phase completion report explicitly excludes @import from the reduction attribution.

---

### AC-CMD-005 — Changelog footer removed, identity lines optionally retained

**GEARS (Ubiquitous)**: The "Changes in vX.Y.Z" changelog prose SHALL be removed from both trees; the `Version:`/`Language:`/`Core Rule:` identity lines MAY remain.

**Given** L638/L644 "Changes in v…" at baseline, **When** the diet is applied, **Then** `grep -c '^Changes in v'` is 0 in both trees.

```bash
CL=$(grep -c '^Changes in v' CLAUDE.md)
CT=$(grep -c '^Changes in v' internal/template/templates/CLAUDE.md)
echo "changes_footer live=$CL template=$CT (baseline 2 each)"
# PASS: CL == 0 AND CT == 0
[ "$CL" -eq 0 ] && [ "$CT" -eq 0 ] && echo "AC-CMD-005 PASS" || echo "AC-CMD-005 FAIL"
```

PASS predicate: 0 in both trees. (Identity `Version:` line presence is NOT asserted either way — optional per REQ-CMD-002.)

---

### AC-CMD-006 — Always-on byte-sum reduced per tree (attributed to real mechanisms)

**GEARS (State-driven)**: **While** the diet is in effect, each tree's byte count SHALL drop below the 35778 baseline, the reduction attributable solely to M-DELETE + M-POINTER + M-SCOPE (NOT @import).

**Given** 35778 bytes/tree at baseline, **When** the diet is applied, **Then** `wc -c` is < 35778 in both trees AND the two are equal.

```bash
BL=$(wc -c < CLAUDE.md)
BT=$(wc -c < internal/template/templates/CLAUDE.md)
echo "bytes live=$BL template=$BT (baseline 35778 each)"
# PASS: BL < 35778 AND BL == BT
[ "$BL" -lt 35778 ] && [ "$BL" -eq "$BT" ] && echo "AC-CMD-006 PASS" || echo "AC-CMD-006 FAIL"
```

PASS predicate: both < 35778 AND equal. The run-phase report must attribute the delta to M-DELETE/M-POINTER (cross-check with AC-CMD-004).

---

### AC-CMD-007 — Neutrality / internal-content isolation preserved (CI-guard-aligned, D3)

**GEARS (Unwanted behavior)**: The diet SHALL NOT inject any NEW moai-adk internal artifact (internal SPEC-ID, internal date, commit SHA, audit citation, `feedback_*`/memory path, `/Users/` path) into the CLAUDE.md body, as judged by the authoritative CI guard.

**(D3 — the AC is delegated to the CI guard, NOT a broad standalone grep.)** The iter-1 broad grep `grep -nE 'SPEC-…|feedback_|…|[0-9a-f]{9,40}'` FALSE-FAILED on a PRE-EXISTING neutral-baseline line — `CLAUDE.md` L459 (`internal/template/templates/CLAUDE.md` mirror) reads *"Per user policy 2026-05-17 … See `feedback_worktree_autonomous` memory"*, which matches both the `feedback_` pattern AND a date, yet is PRE-EXISTING content the SPEC's "neutral baseline" claim must accommodate. The authoritative neutrality guard is the CI test `TestTemplateNeutralityAudit` in `internal/template/template_neutrality_audit_test.go` (it carries the canonical allow-list distinguishing pre-existing/neutral content from genuine internal-leak injection). The AC delegates to it.

**Given** the CI guard passing at baseline, **When** the diet is applied, **Then** the CI guard STILL passes (the diet introduced no NEW leak).

```bash
# Authoritative CI guard (the canonical neutrality semantics, with its own allow-list).
go test ./internal/template/ -run TestTemplateNeutralityAudit 2>&1 | tail -5
# PASS: ok (the guard passes on the post-diet template tree)

# Optional informational delta (NOT the gate): only NEW forbidden tokens vs the pre-diet baseline matter.
# Baseline is captured pre-diet (plan.md §C.1); the diet must not ADD any match beyond the baseline set
# (e.g. the pre-existing L459 feedback_worktree_autonomous line is in the baseline and is NOT a regression).
git diff --no-index <(git show HEAD:internal/template/templates/CLAUDE.md) internal/template/templates/CLAUDE.md \
  | grep '^+' | grep -nE 'SPEC-[A-Z]+-[0-9]{3}|feedback_|/Users/|Audit [0-9]+ Finding' \
  || echo "no NEW internal-artifact token added by the diet"
```

PASS predicate: `go test … -run TestTemplateNeutralityAudit` exits 0 (the authoritative gate). The informational diff-grep is a defense-in-depth check that the diet ADDED no new forbidden token — pre-existing neutral-baseline content (the L459 `feedback_worktree_autonomous` reference) is NOT a regression and does NOT fail this AC.

---

### AC-CMD-008 — §4 nesting note reconciled with official mechanism

**GEARS (Ubiquitous)**: The §4 "Watch (Claude Code 2.1.172)" nesting note SHALL reflect the verified official mechanism — nested spawning via `Agent` in `tools`, depth fixed at 5 and not configurable.

**Given** the imprecise "opt-in / disabled by default" framing at baseline, **When** the correction is applied, **Then** the §4 note references the `Agent`-in-`tools` mechanism AND states the depth limit is fixed (not configurable).

```bash
# The corrected note must mention the tools-list mechanism AND the fixed depth.
sed -n '/## 4. Agent Catalog/,/## 5\./p' CLAUDE.md | grep -iE "tools|depth (five|5)|fixed|not configurable|omit.*Agent"
# PASS: at least one match describing the Agent-in-tools mechanism + fixed-depth fact
sed -n '/## 4. Agent Catalog/,/## 5\./p' CLAUDE.md \
  | grep -iqE "depth (five|5)|fixed and not configurable|Agent.*tools" \
  && echo "AC-CMD-008 PASS" || echo "AC-CMD-008 FAIL"
```

PASS predicate: the §4 note body matches the `Agent`-in-`tools` / fixed-depth-5 mechanism (run-phase authors the exact wording per REQ-CMD-008 AFTER WebFetch re-confirmation + reconciliation with `agent-authoring.md:L100`; the grep confirms the mechanism is named, not the old "opt-in/disabled by default"-only framing). SHOULD (D4) — non-blocking because the precise CC behavior is a residual risk, not an asserted fact.

---

### AC-CMD-009 — POINTER edits gated on prose-duplication re-verification (D1 run-phase precondition)

**GEARS (Event-driven)**: **When** the run-phase is about to apply a POINTER edit to a CLAUDE.md section, the run-phase SHALL FIRST re-grep the SSOT rule file for the section's DISTINCTIVE content; **When** the grep returns 0 hits of the distinctive content, the run-phase SHALL block the POINTER edit and reclassify the section to KEEP.

**Given** the surviving POINTER set (§7, §8, §10, §13, §14-Worktree-subsection, §15-non-CG per plan.md §C.2), **When** run-phase re-verifies each before editing, **Then** each shows ≥1 hit of its distinctive content; any 0-hit section is KEPT, not pointer-ized.

```bash
# Re-run the distinctive-content prose-duplication grep for each surviving POINTER candidate.
# A 0-hit result is a BLOCK signal (force KEEP), not a pass. Example for the demoted-vs-surviving split:
echo "--- surviving POINTER (expect >=1) ---"
grep -c 'Ambiguous pronouns\|Multi-interpretable\|Unclear boundaries' .claude/rules/moai/core/askuser-protocol.md   # §7
grep -c 'select:AskUserQuestion\|max 4 questions' .claude/rules/moai/core/askuser-protocol.md                        # §8
grep -c 'WebSearch\|WebFetch\|verify each URL' .claude/rules/moai/core/moai-constitution.md                          # §10
grep -c 'Level 1\|Level 2\|skillListingBudget' .claude/rules/moai/development/skill-authoring.md                     # §13
grep -c 'Worktree Isolation\|isolation.*worktree' .claude/rules/moai/workflow/worktree-integration.md               # §14-subsection
grep -c 'CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS\|SendMessage\|TeammateIdle' .claude/rules/moai/workflow/spec-workflow.md # §15-non-CG
echo "--- DEMOTED (must stay KEEP — 0 hits confirms the demotion) ---"
grep -c 'Token limit\|Permission error' .claude/rules/moai/core/agent-common-protocol.md          # §11 → expect 0
grep -c 'File Write Conflict\|Loop Prevention' .claude/rules/moai/core/moai-constitution.md        # §14-bullets → expect 0
grep -c 'moai cg' .claude/rules/moai/workflow/spec-workflow.md                                     # §15-CG → expect 0
grep -c '150,000\|150000' .claude/rules/moai/workflow/context-window-management.md                 # §16 → expect 0
# PASS: every surviving-POINTER grep >= 1 AND every demoted-section grep == 0
```

PASS predicate: every surviving POINTER candidate shows ≥1 distinctive-content hit before its edit; every demoted section shows 0 hits (confirming the KEEP decision). A surviving-POINTER candidate that drops to 0 at run-time (SSOT changed between plan and run) MUST be reclassified KEEP — the run-phase reports this as a plan-vs-run drift, not a silent deletion. **MUST / Blocking** — this is the core defense against the D1 over-cut failure mode.

---

### AC-CMD-010 — Behavioral-but-untagged anchors survive the diet (D2)

**GEARS (Unwanted behavior)**: The diet SHALL NOT drop the behavioral anchors that carry value WITHOUT a `[HARD]`/`[ZONE:*]` tag — specifically the §4 8-agent catalog table (the 8 retained agent names), the §4 archived-agent name list, and the §4 Selection Decision Tree. The 14-line `[HARD]` count (AC-CMD-003) does NOT cover these untagged-but-behavioral anchors.

**Given** the baseline §4 anchors, **When** the diet is applied, **Then** the count of retained-agent names, archived-agent names, and decision-tree markers does NOT drop.

```bash
# Retained-agent names (8 expected) — spawn-routing depends on these.
RA=$(grep -cE 'manager-spec|manager-develop|manager-docs|manager-git|plan-auditor|sync-auditor|builder-harness|Explore' CLAUDE.md)
# Archived-agent names (spawn-rejection depends on the list being present).
AA=$(grep -cE 'manager-strategy|manager-quality|manager-brain|manager-project|expert-backend|expert-frontend|expert-security|expert-devops' CLAUDE.md)
# Selection Decision Tree marker.
DT=$(grep -c 'Selection Decision Tree' CLAUDE.md)
echo "retained-agent-refs=$RA  archived-agent-refs=$AA  decision-tree=$DT"
# Baseline (capture pre-diet in plan.md §C.1): RA_base, AA_base, DT_base.
# PASS: RA >= RA_base AND AA >= AA_base AND DT >= 1  (no behavioral anchor dropped)
[ "$DT" -ge 1 ] && echo "AC-CMD-010 decision-tree present" || echo "AC-CMD-010 FAIL (decision tree dropped)"
```

PASS predicate: the 8 retained-agent names still appear (`RA` ≥ pre-diet baseline), the archived-agent list still appears (`AA` ≥ baseline — spawn-rejection guard intact), AND the Selection Decision Tree heading survives (`DT` ≥ 1). The pre-diet baselines RA_base/AA_base are captured in plan.md §C.1 pre-flight. **MUST / Blocking** — these untagged anchors are behavioral (spawn routing + spawn rejection) and the [HARD]-count guard does not protect them.

---

## D.1 Severity / blocking classification

| AC | Severity | Blocking? |
|----|----------|-----------|
| AC-CMD-001 | MUST | Blocking (the diet's primary measurable outcome) |
| AC-CMD-002 | MUST | Blocking (byte-parity is the Template-First invariant) |
| AC-CMD-003 | MUST | Blocking (behavioral [HARD] directives must survive) |
| AC-CMD-004 | MUST | Blocking (the @import-honesty constraint — the load-bearing one) |
| AC-CMD-005 | MUST | Blocking (changelog removal is a named M-DELETE deliverable) |
| AC-CMD-006 | MUST | Blocking (real byte reduction, real-mechanism attribution) |
| AC-CMD-007 | MUST | Blocking (neutrality CI guard — D3-aligned to `TestTemplateNeutralityAudit`) |
| AC-CMD-008 | SHOULD | Non-blocking (D4: precise CC mechanism is a residual risk; run-phase re-fetch + reconcile before asserting) |
| AC-CMD-009 | MUST | Blocking (D1: POINTER edits gated on prose-duplication re-verification — core over-cut defense) |
| AC-CMD-010 | MUST | Blocking (D2: behavioral-but-untagged §4 anchors survive — spawn routing + rejection) |

## D.2 Traceability (REQ → AC)

| REQ | AC | Theme |
|-----|----|-------|
| REQ-CMD-001 | AC-CMD-001, AC-CMD-006 | per-line-test body diet → measurable drop |
| REQ-CMD-002 | AC-CMD-005 | changelog-footer removal |
| REQ-CMD-003 | AC-CMD-001, AC-CMD-006 | rule-SSOT pointer-ization → drop |
| REQ-CMD-004 | AC-CMD-004 | @import token-neutrality honesty |
| REQ-CMD-005 | AC-CMD-002 | template-mirror parity |
| REQ-CMD-006 | AC-CMD-003 | [HARD]-directive retention |
| REQ-CMD-007 | AC-CMD-007 | neutrality / isolation |
| REQ-CMD-008 | AC-CMD-008 | §4 nesting-note correctness (D4 — SHOULD) |
| REQ-CMD-009 | AC-CMD-001 | derived target range (not arbitrary number) |
| REQ-CMD-003 (D1 strengthening) | AC-CMD-009 | POINTER edits gated on prose-duplication re-verification |
| REQ-CMD-001 / REQ-CMD-006 (D2) | AC-CMD-010 | behavioral-but-untagged §4 anchors survive |

## D.3 Definition of Done

- [ ] All 8 MUST-blocking ACs PASS (AC-CMD-001..007, AC-CMD-009, AC-CMD-010 — note AC-CMD-008 is SHOULD).
- [ ] AC-CMD-008 (SHOULD) reconciled AFTER run-phase WebFetch re-confirmation + `agent-authoring.md:L100` reconciliation (D4).
- [ ] AC-CMD-009 (D1): every surviving-POINTER section re-verified to carry distinctive SSOT prose before its edit; any 0-hit section reclassified KEEP.
- [ ] AC-CMD-010 (D2): §4 retained-agent (8) + archived-agent list + Selection Decision Tree survive.
- [ ] `go build ./...` exit 0 (embedded.go compiles after `make build`).
- [ ] `go test ./internal/template/...` green (mirror-parity + `TestTemplateNeutralityAudit`).
- [ ] spec-lint clean on this SPEC dir (`MissingExclusions` 0 — §D h3 sub-headings present).
- [ ] Both trees committed + pushed (Conventional Commits, no `--no-verify`).
- [ ] §A.5 PRESERVE-list paths untouched (`git status` shows only the 2 CLAUDE.md trees + this SPEC's 4 artifacts + embedded.go).
- [ ] Run-phase completion report attributes the byte reduction to M-DELETE/M-POINTER (excludes @import) per AC-CMD-004.

## D.4 Indirect verification note

The "token reduction" is verified INDIRECTLY via line/byte-sum (AC-CMD-001/006), not a real tokenizer — acceptable per the per-line-test framing (spec.md §F.2). If an exact token count is later required, run-phase MAY add a tokenizer measurement, but it is not a blocking AC here.
