# Plan: SPEC-V3R6-HARNESS-LEARNER-FIX-001

## 1. Implementation Strategy

This is a **Tier S single-file targeted edit**. The change surface is `.claude/skills/moai-harness-learner/SKILL.md` exclusively. No code, no config, no test fixtures — pure documentation correction to align with the canonical orchestrator-subagent boundary contract.

### Change Surface

**Single file**: `.claude/skills/moai-harness-learner/SKILL.md` (162 LOC, last modified 2026-05-15).

**Edit regions** (3 discrete Edit operations):

1. **Frontmatter line 20** (REQ-HLF-001):
   - Old: `allowed-tools: Bash,Read,Write,Edit,AskUserQuestion`
   - New: `allowed-tools: Bash,Read,Write,Edit`

2. **Body line 35** (REQ-HLF-004 + REQ-HLF-003):
   - Old: `**Key constraint** [HARD]: 'moai harness apply' returns a JSON payload. This skill MUST receive that payload and surface it via 'AskUserQuestion'. The CLI itself does NOT prompt the user.`
   - New: `**Key constraint** [HARD]: 'moai harness apply' returns a JSON payload. This skill produces the payload; the orchestrator surfaces it via 'AskUserQuestion'. The CLI itself does NOT prompt the user. Canonical contract: '.claude/rules/moai/core/askuser-protocol.md § Orchestrator-Subagent Boundary' (CONST-V3R5-001/002/003).`

3. **Body lines 80-95** (REQ-HLF-002 + REQ-HLF-004) — Step 3 section rewrite:
   - Old: `### Step 3: Surface via AskUserQuestion\n\n[HARD] This skill (not the CLI) calls AskUserQuestion. The CLI only provides the payload.\n\n[example of skill-direct invocation]`
   - New: `### Step 3: Produce structured payload for orchestrator consumption\n\n[HARD] This skill produces a structured payload; the MoAI orchestrator surfaces it via 'AskUserQuestion'. Canonical contract: '.claude/rules/moai/core/askuser-protocol.md § Orchestrator-Subagent Boundary'.\n\nPayload schema (REQ-HLF-002):\n - proposal_id, target_path, field_key, current_value, new_value, observation_count, confidence, recommended_action\n\nThe skill emits this payload as its tool output. The orchestrator reads the payload, preloads 'AskUserQuestion' via 'ToolSearch', and surfaces the approve/reject decision to the user. On user approval, the orchestrator re-delegates to this skill with 'action=apply'; on rejection, with 'action=skip'.`

### Section preservation contract (REQ-HLF-005)

The following structural elements MUST remain unchanged:

- 4-Tier observation ladder (Observation 1× / Heuristic 3× / Rule 5× / Auto-Update 10×)
- L1-L5 safety architecture references (table at line ~154)
- Downstream GAN-loop contract section
- @MX:NOTE annotation at line 26 (V3R4 contract preservation note — informational, not actionable)
- Description at line 3 (frontmatter description field) — references AskUserQuestion as part of the *user-facing flow description*; this is the orchestrator's role and is not a violation. **Preserve verbatim.**

### Approach: 3 sequential Edit operations

Tier S workflow does not require parallel Edit. The 3 regions are well-separated; sequential Edit is safer and easier to verify per AC. Each Edit's `old_string` MUST include enough surrounding context to be unique within the file.

## 2. Verification Commands (parallel batch — single orchestrator turn)

Per `.claude/rules/moai/workflow/verification-batch-pattern.md`, all 7 AC verifications run as a single multi-Bash call in the run-phase completion turn:

```bash
# AC-HLF-001: frontmatter compliance
grep "^allowed-tools" .claude/skills/moai-harness-learner/SKILL.md | grep -c "AskUserQuestion"
# Expected: 0

# AC-HLF-002: canonical reference present
grep -n "askuser-protocol.md" .claude/skills/moai-harness-learner/SKILL.md
# Expected: ≥ 1 match

# AC-HLF-003: anti-pattern "this skill calls" removed
grep -c "This skill.*calls AskUserQuestion\|This skill (not the CLI) calls AskUserQuestion" .claude/skills/moai-harness-learner/SKILL.md
# Expected: 0

# AC-HLF-004: anti-pattern "surface via" removed
grep -c "Surface.*via AskUserQuestion\|Surface payload via .*AskUserQuestion" .claude/skills/moai-harness-learner/SKILL.md
# Expected: 0

# AC-HLF-005: preservation of structural elements
grep -c "4-Tier\|Tier 4 auto-update\|L1.*L5\|safety architecture" .claude/skills/moai-harness-learner/SKILL.md
# Expected: ≥ 4

# AC-HLF-006: structured payload field present
grep -n "proposal_id" .claude/skills/moai-harness-learner/SKILL.md
# Expected: ≥ 1 match

# AC-HLF-007: build integrity (embedded.go regeneration safe)
go test ./internal/template/...
# Expected: PASS
```

## 3. Commit Plan

**Single commit on main** (Late-Branch — no branch creation in plan-phase per SPEC-V3R5-LATE-BRANCH-001 REQ-LB-005).

```
fix(SPEC-V3R6-HARNESS-LEARNER-FIX-001): moai-harness-learner allowed-tools subagent boundary compliance

P0-S1 from .claude/skills/ audit 2026-05-21: remove AskUserQuestion from
moai-harness-learner allowed-tools and rewrite Step 3 to align with the
orchestrator-subagent boundary contract (CONST-V3R5-001/002/003).

The skill now produces a structured payload; the orchestrator surfaces it via
AskUserQuestion. Canonical contract: .claude/rules/moai/core/askuser-protocol.md
§ Orchestrator-Subagent Boundary.

Verification: 7/7 ACs PASS (AC-HLF-001..007).
- AC-HLF-001: allowed-tools no longer contains AskUserQuestion
- AC-HLF-002: askuser-protocol.md cited as canonical contract
- AC-HLF-003/004: "this skill calls AskUserQuestion" anti-pattern removed
- AC-HLF-005: 4-Tier ladder + L1-L5 safety architecture preserved
- AC-HLF-006: structured payload schema documented
- AC-HLF-007: go test ./internal/template/... PASS (no embed regression)

Scope: single file (.claude/skills/moai-harness-learner/SKILL.md), 3 Edit
operations (frontmatter line 20 + body line 35 + Step 3 section lines 80-95).

🗿 MoAI <email@mo.ai.kr>
```

**No PR creation in plan-phase**. Per Late-Branch workflow, the orchestrator decides PR creation timing after run-phase completion.

## 4. Rollback Procedure

If post-edit verification fails (any AC returns unexpected value) or `go test ./internal/template/...` regresses:

1. Identify the failing AC and the corresponding Edit region.
2. If single Edit caused the regression: `git diff HEAD -- .claude/skills/moai-harness-learner/SKILL.md` to inspect the change, then re-Edit with corrected `old_string`/`new_string`.
3. If multiple Edits compound a failure: `git checkout HEAD -- .claude/skills/moai-harness-learner/SKILL.md` to revert the entire file, then re-apply the 3 Edits with revised plan.
4. If `embedded.go` regenerates unexpectedly: `git status internal/template/embedded.go`; if modified, the change is expected (skill content embedded). Run `make build` if necessary, then re-test.

**No code-level rollback needed** — this SPEC modifies only a documentation file. Worst case: revert the single file.

## 5. Brownfield Strategy (PRESERVE / EXTEND / REPLACE)

- **PRESERVE** (REQ-HLF-005):
  - Frontmatter fields other than `allowed-tools` (line 1-19, 21-22): name, description, version, etc.
  - `@MX:NOTE` annotation at line 26.
  - Lines 27-79 (introduction, Role description, Tier ladder, observation flow Step 1-2).
  - Lines 96+ (any content after Step 3 — likely L1-L5 safety table, GAN-loop contract, examples).
  - Line 154 L5 row reference to "Surface via AskUserQuestion (this skill)" — this is a *contract description in a safety table*, not an instruction to the skill itself. **Preserve verbatim** because the orchestrator's surface step is part of the L5 description.
    - If grep AC-HLF-003/004 flag this line as a violation: re-examine the regex. The grep regex targets *active assertions* ("This skill calls...", "Surface payload via..."), not contract-table cells. Verify via line-number inspection.

- **EXTEND**:
  - Add structured payload schema (REQ-HLF-002) to the Step 3 section.
  - Add canonical contract citation (`.claude/rules/moai/core/askuser-protocol.md`) at line 35 (constraint paragraph) and in the rewritten Step 3.

- **REPLACE** (narrow surface only):
  - Line 20: `allowed-tools` value (remove `,AskUserQuestion`).
  - Line 35: constraint paragraph rewording.
  - Lines 80-95: Step 3 section heading and body.

**Working tree hygiene (B8)**: The 11 dirty PRESERVE files listed in the delegation contract (.claude/settings.json, internal/merge/* (5 files), internal/cli/init_layout.go, internal/cli/wizard/fullscreen.go, internal/cli/wizard/review.go, internal/hook/.moai/, .moai/harness/usage-log.jsonl, .claude/commands/99-release.md, .claude/skills/moai/workflows/release.md) MUST NOT be touched during run-phase Edit operations. The orchestrator's commit MUST stage only `.claude/skills/moai-harness-learner/SKILL.md` and (if regenerated) `internal/template/embedded.go`.

## 6. Risk Mitigation Reference

See spec.md § 6 Risks (R-1, R-2, R-3). Mitigations are integrated into:
- R-1 (section entanglement): handled by 3 sequential Edits with sufficient `old_string` context + AC-HLF-005 preservation grep.
- R-2 (embedded.go regression): handled by AC-HLF-007 build test.
- R-3 (downstream cross-references): handled by pre-flight grep during run-phase: `grep -rn "moai-harness-learner.*AskUserQuestion" .claude/ | grep -v "/moai-harness-learner/SKILL.md"`. If matches found, fold into the same fix or escalate as a structured blocker report to the orchestrator (no `AskUserQuestion` from subagent per agent-common-protocol.md).

---

**Plan-auditor expectation**: Tier S threshold 0.75. LEAN workflow: 0 BLOCKING findings → 1-iteration PASS. SHOULD/INFO findings folded into run-phase if material. Score-regression STOP escalation enabled (max 3 iterations).

**Delegation contract**: manager-develop (cycle_type=ddd, since the existing file has structural content to preserve — characterize-then-improve discipline). Single-file edit; Section B6 (spec-lint heading) and B4 (frontmatter schema) are applicable known issues; B1 (cross-platform build tags), B2 (cross-SPEC conflict), B3 (subagent boundary grep — *meta-relevance*: this SPEC IS the boundary fix), B5 (CI 3-tier), B7 (observer path) are not relevant to this Tier S documentation change.
