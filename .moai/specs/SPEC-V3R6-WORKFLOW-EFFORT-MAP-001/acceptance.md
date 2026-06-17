# acceptance.md — SPEC-V3R6-WORKFLOW-EFFORT-MAP-001

> Acceptance criteria and Given-When-Then scenarios. Each AC is observable, grep-verifiable, or byte-parity-verifiable. AC IDs follow `AC-WEM-NNN` per the spec.md §G matrix.

## §D. Acceptance Criteria Matrix

### AC-WEM-001 (MUST) — SSOT taxonomy exists and is cited by downstream surfaces

**Given** the dynamic-workflows.md rule file carries the canonical purpose→(model,effort) taxonomy,
**When** a reader consults any of the 3 downstream surfaces (workflow.yaml comments, codemaps-extract.js header comment, session-handoff.md Block 1 field-spec),
**Then** the surface cross-references the dynamic-workflows.md taxonomy section OR reproduces its essential content.

**Evidence:**
```bash
grep -n "Purpose-driven model+effort selection\|purpose.*model.*effort" .claude/rules/moai/workflow/dynamic-workflows.md
# expect ≥ 1 match (the section heading)
```

### AC-WEM-002 (MUST) — Workflow agent() explicit-purpose rule documented

**Given** the dynamic-workflows.md taxonomy section exists,
**When** a workflow-script author reads the section,
**Then** the section states that `effort` SHALL be set explicitly per taxonomy rather than inherited from the session default.

**Evidence:**
```bash
grep -n "set.*effort.*explicitly\|effort.*explicitly.*per.*taxonomy\|SHALL.*set.*effort" .claude/rules/moai/workflow/dynamic-workflows.md
# expect ≥ 1 match
```

### AC-WEM-003 (MUST) — codemaps-extract.js sets effort: 'low' per stage

**Given** the codemaps-extract.js `Extract` phase invokes one Explore agent per package via `agent()`,
**When** the script runs,
**Then** the `agent()` call passes `effort: 'low'` (read-only-extract purpose per §D taxonomy).

**Expected literal form:** the `effort` key MUST appear as a **single-line object-literal property** on the `agent()` opts argument — i.e. a grep for `effort:` on the SAME line as `agentType: 'Explore'` MUST match. A multi-line opts object where `effort:` sits on its own line is also acceptable, but the AC verifiable form is: the line carrying `agentType:` also carries `effort:` (or the immediately adjacent line within the same opts literal). The check below verifies this directly.

**Evidence:**
```bash
# Match the single-line object-literal form: agentType and effort on ONE line.
grep -nE "agentType: *'Explore'.*effort: *'low'|effort: *'low'.*agentType: *'Explore'" .claude/workflows/codemaps-extract.js
# expect ≥ 1 match (effort: 'low' co-located with agentType: 'Explore' on the same line)

# Fallback: if the opts literal is multi-line, accept effort: 'low' within 3 lines after agentType.
grep -nA3 "agentType: 'Explore'" .claude/workflows/codemaps-extract.js | grep -E "effort: *'low'"
# expect ≥ 1 match
```

**Negative check (must NOT still inherit session default):**
```bash
# The agent() call MUST carry effort explicitly — i.e. the opts literal containing
# agentType: 'Explore' MUST also contain effort:. This is the positive form of
# "no longer inherits session default".
grep -nA3 "agentType: 'Explore'" .claude/workflows/codemaps-extract.js | grep -c "effort:"
# expect ≥ 1
```

### AC-WEM-004 (MUST) — workflow.yaml (local) role_profiles carry effort for all 7 roles

**Given** the local `.moai/config/sections/workflow.yaml` declares 7 role_profiles,
**When** the file is inspected post-edit,
**Then** each of the 7 roles (analyst, architect, designer, implementer, researcher, reviewer, tester) has an `effort:` key with a value from `{low, medium, high, xhigh, max}`.

**Evidence:**
```bash
# Count effort keys under role_profiles. The awk range MUST be bounded by the
# next TOP-LEVEL workflow: sibling at 4-space indent (token_budget:), because
# role_profiles is the LAST child of team: — patterns: PRECEDES role_profiles
# in the file, so the iter-0 end-anchor /^        patterns:/ never closes and
# over-captures to EOF. Verified 2026-06-17: the range below captures exactly
# 37 lines containing all 7 role keys and terminates cleanly at token_budget:.
awk '/^        role_profiles:/,/^    token_budget:/' .moai/config/sections/workflow.yaml | grep -cE '^            [a-z_]+:'
# expect 7 (the 7 role keys: analyst/architect/designer/implementer/researcher/reviewer/tester)

awk '/^        role_profiles:/,/^    token_budget:/' .moai/config/sections/workflow.yaml | grep -c 'effort:'
# expect 7 (one effort: per role)

# Verify each of the 7 roles individually has effort
for role in analyst architect designer implementer researcher reviewer tester; do
  awk '/^        role_profiles:/,/^    token_budget:/' .moai/config/sections/workflow.yaml \
    | awk -v r="$role" 'BEGIN{p=0} $0 ~ "^            "r":"{p=1} p&&/effort:/{print; p=0}' | head -1
done
# expect 7 non-empty lines
```

### AC-WEM-005 (MUST) — workflow.yaml (template mirror) role_profiles carry effort for all 7 roles, byte-aligned values with local

**Given** the template mirror `internal/template/templates/.moai/config/sections/workflow.yaml` is the user-distributable copy,
**When** it is inspected post-edit,
**Then** (a) each of the 7 roles has an `effort:` key, AND (b) the 7 values are byte-identical to the local file's values.

**Evidence:**
```bash
# (a) count — corrected awk range (see AC-WEM-004 note: patterns: precedes
#     role_profiles, so the end-anchor MUST be the next top-level workflow:
#     sibling token_budget: at 4-space indent)
awk '/^        role_profiles:/,/^    token_budget:/' internal/template/templates/.moai/config/sections/workflow.yaml | grep -c 'effort:'
# expect 7

# (b) value alignment — corrected awk range on BOTH sides
diff <(awk '/^        role_profiles:/,/^    token_budget:/' .moai/config/sections/workflow.yaml | grep 'effort:') \
     <(awk '/^        role_profiles:/,/^    token_budget:/' internal/template/templates/.moai/config/sections/workflow.yaml | grep 'effort:')
# expect 0 diff lines
```

### AC-WEM-006 (MUST) — No Go source under internal/ is modified by this SPEC's commits (decision #1 invariant)

**Given** decision #1 forbids Go source changes and REQ-WEM-006 makes the YAML field declarative-only,
**When** the SPEC's run-phase commits are inspected,
**Then** no Go file under `internal/` appears in the commits' file lists — the declarative `effort` field is consumed by the orchestrator (LLM) and workflow-script authors, never by Go code.

**PRIMARY evidence (git-diff-based — the real decision-#1 invariant):**
```bash
# List Go files under internal/ touched by this SPEC's commits.
# (Replace HEAD~N with the commit-range that spans this SPEC's run-phase commits,
#  or use the SPEC-ID grep form below which is commit-range-agnostic.)
git diff --name-only origin/main -- internal/ | grep '\.go$'
# expect: 0 lines (no Go files under internal/ modified by this SPEC)

# Commit-range-agnostic form (grep by SPEC-ID in commit subject):
git log --grep='SPEC-V3R6-WORKFLOW-EFFORT-MAP-001' --name-only --pretty=format: \
  | sort -u | grep -E '^internal/.*\.go$'
# expect: 0 lines
```

**SUPPORTING evidence only (struct-field-absence — explains WHY coupling is impossible, but is NOT the decision-#1 check).** The iter-0 PRIMARY was a grep for `grep 'RoleProfiles|role_profiles' internal/ | grep 'effort'`; that grep was **vacuous** because `RoleProfileEntry` (`internal/config/types.go:372`) declares only `Isolation/Mode/Model` fields — there is NO `Effort` field, so no Go code could reference it regardless of whether the YAML adds the key. A grep that proves nothing cannot be the PRIMARY check:
```bash
# Supporting: confirm the struct still has no Effort field (the field cannot couple
# even if a reader existed). This corroborates AC-WEM-006 but does not replace the
# git-diff PRIMARY above.
grep -A6 'type RoleProfileEntry struct' internal/config/types.go | grep -i 'effort'
# expect: 0 matches (field absent — declarative YAML cannot bind to Go)
```

### AC-WEM-007 (MUST) — session-handoff Block 1 documents purpose-conditional ultracode line

**Given** the session-handoff.md `Field-by-Field Specification` describes Block 1,
**When** the file is inspected post-edit,
**Then** Block 1 carries a sub-bullet documenting that a `/effort ultracode` re-set line is emitted ONLY when the next SPEC declares workflow fan-out, with explicit reference to the dynamic-workflows.md doctrine (ultracode not restored by `ultrathink.`).

**Evidence (structural — verifies the conditional SHAPE, not mere token presence):**
```bash
# PRIMARY: the line must be gated by a workflow-fan-out CONDITION. The iter-0
# check `grep -c ultracode ≥ 2` counted token presence and would pass even if
# "ultracode" appeared twice in unrelated prose. The check below verifies the
# actual conditional structure: the word "ultracode" co-occurs on the same line
# (or within a 2-line window) as a fan-out-condition term.
grep -nE -B0 -A2 'ultracode' .claude/rules/moai/workflow/session-handoff.md \
  | grep -iE 'workflow.*fan.?out|fan.?out|next SPEC.*workflow|workflow.*next SPEC|dynamic Workflow|Agent Teams'
# expect ≥ 1 match — the ultracode line is conditioned on workflow fan-out

# The literal re-set command form must also appear:
grep -nE '/effort ultracode' .claude/rules/moai/workflow/session-handoff.md
# expect ≥ 1 match (the literal re-set line)

# The doctrine reference (ultracode not restored by ultrathink.) must appear:
grep -nE 'ultracode.{0,40}(NOT|not).{0,30}restored|ultrathink.{0,40}(not|NOT).{0,30}restored' .claude/rules/moai/workflow/session-handoff.md
# expect ≥ 1 match (the doctrine reference)
```

### AC-WEM-008 (MUST) — output-styles/moai/moai.md §8 render surface carries same Block 1 update

**Given** moai.md §8 is the canonical render surface and session-handoff.md is the SSOT (bidirectional parity per session-handoff.md Cross-references),
**When** moai.md §8 is inspected post-edit,
**Then** the render surface carries the same ultracode conditional-line documentation as session-handoff.md Block 1 — i.e. the same purpose-conditional structure verified in AC-WEM-007, not just two token occurrences.

**Evidence (structural — mirrors AC-WEM-007's conditional-shape check):**
```bash
# PRIMARY: ultracode co-occurs with a fan-out-condition term (same conditional shape
# as AC-WEM-007 — NOT a bare `grep -c ultracode ≥ 2` which only proves token presence).
grep -nE -B0 -A2 'ultracode' .claude/output-styles/moai/moai.md \
  | grep -iE 'workflow.*fan.?out|fan.?out|next SPEC.*workflow|workflow.*next SPEC|dynamic Workflow|Agent Teams'
# expect ≥ 1 match

# The literal re-set command form:
grep -nE '/effort ultracode' .claude/output-styles/moai/moai.md
# expect ≥ 1 match
```

### AC-WEM-009a (MUST) — session-handoff.md mirror pair byte-parity (CI-enforced)

**Given** `internal/template/rule_template_mirror_test.go` (the REAL mirror-parity test — the iter-0 draft's `embedded_mirror_test.go` is a phantom file that does NOT exist) carries an allowlist covering `.claude/rules/moai/workflow/session-handoff.md` (allowlist lines 47, 51),
**When** the SPEC's run-phase commits land,
**Then** the session-handoff.md SSOT and its template mirror are byte-identical, verified BOTH by the CI test AND by a `diff` corroboration.

**Evidence:**
```bash
# PRIMARY — CI-enforced (the Go allowlist covers this pair):
go test ./internal/template/ -run TestRuleTemplateMirror
# expect: PASS (the allowlist entry for session-handoff.md enforces 0-diff at CI time)

# Corroboration — manual diff:
diff .claude/rules/moai/workflow/session-handoff.md internal/template/templates/.claude/rules/moai/workflow/session-handoff.md | wc -l
# expect 0
```

### AC-WEM-009b (MUST) — dynamic-workflows.md AND output-styles/moai/moai.md mirror pairs byte-parity (AC-enforced — NOT CI-enforced)

**Given** `internal/template/rule_template_mirror_test.go` allowlist does NOT cover `.claude/rules/moai/workflow/dynamic-workflows.md` or `.claude/output-styles/moai/moai.md` (allowlist covers session-handoff.md ONLY — see AC-WEM-009a),
**When** the SPEC's run-phase commits land,
**Then** both pairs are byte-identical, verified by `diff`/`cmp` (this AC), NOT by CI. Extending the Go allowlist to cover these two pairs would require a Go change under `internal/` and is therefore **deferred to a future SPEC** per decision #1 (no Go source changes).

**Evidence (AC-enforced — this `diff` IS the gate, because CI does not cover these pairs):**
```bash
diff .claude/rules/moai/workflow/dynamic-workflows.md internal/template/templates/.claude/rules/moai/workflow/dynamic-workflows.md | wc -l
# expect 0

diff .claude/output-styles/moai/moai.md internal/template/templates/.claude/output-styles/moai/moai.md | wc -l
# expect 0

#(cmp alternative, byte-exact):
cmp .claude/rules/moai/workflow/dynamic-workflows.md internal/template/templates/.claude/rules/moai/workflow/dynamic-workflows.md && echo OK
cmp .claude/output-styles/moai/moai.md internal/template/templates/.claude/output-styles/moai/moai.md && echo OK
# expect: OK for both
```

**Baseline note:** pre-edit (2026-06-17), both pairs are already 0-diff. This AC verifies the SPEC does NOT break that parity — and, because CI does NOT cover these pairs, this `diff` check is the ONLY automated gate.

### AC-WEM-010 (MUST) — dynamic-workflows.md has the taxonomy subsection

**Given** REQ-WEM-010 requires the taxonomy to land in dynamic-workflows.md,
**When** the file is inspected,
**Then** a section titled "Purpose-driven model+effort selection" (or semantic equivalent) exists and contains (a) the official effort citation, (b) the taxonomy table or a cross-reference to it, (c) the codemaps-extract.js worked example.

**Evidence:**
```bash
grep -n "Purpose-driven\|purpose.*driven.*model.*effort\|## .*Purpose" .claude/rules/moai/workflow/dynamic-workflows.md
# expect ≥ 1 match

grep -n "platform.claude.com/docs/en/build-with-claude/effort" .claude/rules/moai/workflow/dynamic-workflows.md
# expect ≥ 1 match (the official citation)

grep -n "codemaps-extract" .claude/rules/moai/workflow/dynamic-workflows.md
# expect ≥ 1 match (the worked example reference)
```

### AC-WEM-011 (MUST) — Template neutrality (no internal SPEC IDs / REQ tokens / SHAs / feedback_ refs / Audit citations in template mirror)

**Given** REQ-WEM-013 and `.moai/docs/template-internal-isolation-doctrine.md` forbid internal-content leakage into `internal/template/templates/`,
**When** the template-mirror edits are inspected,
**Then** none of the edited template files contain `SPEC-V3R6-WORKFLOW-EFFORT-MAP-001`, `REQ-WEM-`, commit SHAs, `feedback_` refs, or "Audit N Finding" citations introduced by THIS SPEC.

**Evidence:**
```bash
# SPEC ID leakage
grep -rn 'SPEC-V3R6-WORKFLOW-EFFORT-MAP' internal/template/templates/.claude/rules/moai/workflow/{dynamic-workflows,session-handoff}.md internal/template/templates/.claude/output-styles/moai/moai.md internal/template/templates/.moai/config/sections/workflow.yaml
# expect 0 matches

# REQ token leakage
grep -rn 'REQ-WEM-' internal/template/templates/.claude/rules/moai/workflow/{dynamic-workflows,session-handoff}.md internal/template/templates/.claude/output-styles/moai/moai.md internal/template/templates/.moai/config/sections/workflow.yaml
# expect 0 matches

# feedback_ ref leakage (NEW — pre-existing feedback_ refs in other template files are out of scope)
grep -n 'feedback_' internal/template/templates/.claude/rules/moai/workflow/{dynamic-workflows,session-handoff}.md internal/template/templates/.claude/output-styles/moai/moai.md
# expect 0 NEW matches introduced by this SPEC (pre-existing are grandfathered per the doctrine §25.1 allow-list)
```

**Note:** the template mirror's `effort` field comment in workflow.yaml MUST read `# effort: declarative metadata consumed by orchestrator/workflow-script authors; no Go code reads this field.` (NO SPEC ID suffix). The local workflow.yaml's comment MAY read `... per SPEC-V3R6-WORKFLOW-EFFORT-MAP-001 REQ-WEM-006.` (local file is not user-distributable).

### AC-WEM-012 (MUST) — Pre-existing dirty-tree changes NOT absorbed

**Given** REQ-WEM-014 excludes the 5 modified config YAMLs + untracked dirs,
**When** the SPEC's run-phase commits are inspected,
**Then** none of the excluded paths appear in the commits' file lists.

**Evidence:**
```bash
# List all files touched by this SPEC's commits (commits carry SPEC-ID in subject).
# Plan.md §C enumerates 13 paths = 3 doctrine SSOT + 3 template mirrors + 2 config
# (local + template) + 1 workflow script + 4 SPEC artifacts (spec/plan/acceptance/progress).
git log --grep='SPEC-V3R6-WORKFLOW-EFFORT-MAP-001' --name-only --pretty=format: | sort -u | grep -v '^$'
# verify the output contains ONLY the 13 paths enumerated in plan.md §C

# Negative check: excluded paths must NOT appear. The pattern is BROADENED to
# cover ALL sibling SPEC dirs under .moai/specs/ (other than THIS SPEC's own dir),
# not just the 5 config YAMLs + design/docs/reports dirs. The iter-0 regex missed
# parallel untracked SPEC dirs (e.g. .moai/specs/SPEC-V3R6-SESSION-ID-ATTRIBUTION-REPAIR-001/
# is in the dirty tree today). Two independent negative checks:
#
#   (1) the 5 pre-existing modified config-sections YAMLs + untracked design/docs/reports
#   (2) ANY sibling SPEC dir under .moai/specs/ that is NOT this SPEC's own dir
#
# Check (1):
git log --grep='SPEC-V3R6-WORKFLOW-EFFORT-MAP-001' --name-only --pretty=format: \
  | grep -E '\.moai/config/sections/(git-convention|language|llm|quality|user)\.yaml|\.moai/(design|docs|reports)/'
# expect 0 matches
#
# Check (2) — sibling SPEC dirs (this SPEC's own dir is EXCLUDED from the negative set):
git log --grep='SPEC-V3R6-WORKFLOW-EFFORT-MAP-001' --name-only --pretty=format: \
  | grep -E '^\.moai/specs/SPEC-' | grep -v '^\.moai/specs/SPEC-V3R6-WORKFLOW-EFFORT-MAP-001/'
# expect 0 matches (any match here = a sibling SPEC dir was absorbed)
```

## §D.1 Severity Classification

| AC | Severity | Rationale |
|----|----------|-----------|
| AC-WEM-001 | MUST | SSOT non-existence defeats the entire SPEC |
| AC-WEM-002 | MUST | Without the explicit-purpose rule, workflow authors keep inheriting |
| AC-WEM-003 | MUST | The codemaps-extract.js fix is the concrete proof-of-concept for the SSOT |
| AC-WEM-004 | MUST | role_profiles is half the mapping-coverage decision (#2) |
| AC-WEM-005 | MUST | Template mirror byte-alignment is the distributability contract |
| AC-WEM-006 | MUST | git-diff confirms no Go file under internal/ is modified by this SPEC's commits (decision-#1 invariant); struct-field-absence is supporting evidence only |
| AC-WEM-007 | MUST | Handoff Block 1 is decision #3 |
| AC-WEM-008 | MUST | Render-surface parity is the bidirectional-contract invariant |
| AC-WEM-009a | MUST | session-handoff.md byte-parity is CI-enforced (`rule_template_mirror_test.go` allowlist) |
| AC-WEM-009b | MUST | dynamic-workflows.md + output-styles/moai/moai.md byte-parity is AC-enforced (`diff` — these pairs are NOT in the Go allowlist; extending it would require a Go change, deferred per decision #1) |
| AC-WEM-010 | MUST | Taxonomy subsection is where the SSOT lives |
| AC-WEM-011 | MUST | Template neutrality is CI-enforced (`template-neutrality-check.yaml`) |
| AC-WEM-012 | MUST | Scope discipline is the REQ-WEM-014 invariant |

All 13 ACs are MUST (AC-WEM-001..008, 009a, 009b, 010, 011, 012 — AC-WEM-009 was split into 009a/009b at iter-1 to separate CI-enforced parity from AC-enforced parity). No SHOULD/DEFERRED ACs in this SPEC.

## §D.2 Indirect Verification (surfaces not directly grepped)

- **Orchestrator runtime consumption of the `effort` field:** NOT verified at run-phase. The field is declarative; orchestrator consumption is an LLM behavior, not a code path. The AC verifies the field EXISTS and is DOCUMENTED; runtime consumption is trusted to the orchestrator's rule-loading (which already consumes workflow.yaml for the other 3 fields).
- **Actual cost savings from `effort: 'low'` on codemaps-extract.js:** NOT verified at run-phase. This is a downstream metric that would require a token-cost A/B test; out of scope for this SPEC's acceptance. The AC verifies the field is set per the official recommendation; the cost outcome is trusted to the official guidance.

## §D.3 Closure Gates

This SPEC is closed when:
1. All 13 MUST ACs pass (evidence commands return expected counts).
2. `git log --grep='SPEC-V3R6-WORKFLOW-EFFORT-MAP-001'` shows the plan + run + sync + Mx commits per the canonical 4-phase lifecycle.
3. `progress.md` §E.5 carries the Mx-phase audit-ready signal with a real commit SHA (not a placeholder).
4. plan-auditor verdict ≥ 0.80 (Tier M threshold) at plan-phase.
5. sync-auditor verdict ≥ MUST-PASS threshold at sync-phase.

## §D.4 Forward-Looking Checks (post-close)

- A future SPEC may wire the declarative `effort` field into Go runtime enforcement (spawn wrapper, agent_lint). This SPEC explicitly defers that (§H Exclusions). The forward-looking check is: when that future SPEC lands, it MUST cite this SPEC as the declarative SSOT origin.
- A future workflow script (`.claude/workflows/*.js`) SHOULD consult the dynamic-workflows.md taxonomy section when authoring its `agent()` calls. This is a SHOULD (not enforced); the codemaps-extract.js pattern serves as the reference.

## §D.5 Edge Cases

- **Edge: session-handoff resume emitted when next-SPEC plan-phase is incomplete.** The orchestrator may not yet know whether the next SPEC needs workflow fan-out. Resolution: default to OMITTING the ultracode line (REQ-WEM-008) — the resumed session can always `/effort ultracode` interactively. Omitting is the safe default; emitting unconditionally is the AP called out in AC-WEM-007's anti-pattern list.
- **Edge: workflow.yaml local file diverges from template by >119 lines post-edit.** Acceptable per AC-WEM-005 — only the 7 `effort` values must align; the rest of the local-only divergence (comments, structural) is pre-existing and out of scope.
- **Edge: codemaps-extract.js runs under a Claude Code version that rejects `effort` on workflow subagents.** Not expected (effort is documented for workflow agent() opts), but if it occurs, the fix is a CC-version compatibility note, not a SPEC revert. The AC verifies the field is set; runtime rejection is a CC bug.

## §D.6 Quality Gate Criteria

- **Lint:** spec-lint clean on all 4 SPEC artifacts (spec.md / plan.md / acceptance.md / progress.md).
- **Mirror parity CI:** `internal/template/rule_template_mirror_test.go` passes (covers session-handoff.md pair — AC-WEM-009a). The other 2 pairs (dynamic-workflows.md, output-styles/moai/moai.md) are NOT CI-covered and are gated by AC-WEM-009b's `diff` check instead.
- **Template neutrality CI:** `.github/workflows/template-neutrality-check.yaml` passes (no internal content leakage).
- **No Go build required:** the SPEC's surfaces are doctrine/YAML/JS/MD; `make build` is not load-bearing.

## §D.7 Definition of Done

- [ ] All 13 MUST ACs pass with evidence commands returning expected counts
- [ ] 4-phase lifecycle complete (plan / run / sync / Mx)
- [ ] plan-auditor ≥ 0.80, sync-auditor MUST-PASS
- [ ] `progress.md` §E.5 carries real Mx commit SHA
- [ ] No regression in pre-existing `rule_template_mirror_test.go` (covers session-handoff.md) / `template-neutrality-check.yaml` CI
- [ ] The 5 dirty-tree YAMLs + untracked dirs remain unabsorbed (AC-WEM-012)
