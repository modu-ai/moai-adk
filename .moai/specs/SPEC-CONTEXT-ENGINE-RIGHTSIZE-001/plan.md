# plan.md — SPEC-CONTEXT-ENGINE-RIGHTSIZE-001

> Tier M implementation plan. Milestones ordered by decision-reversibility (highest-change-likelihood first): the code_comments expressive transition is most user-visible (affects how every agent comments code globally), so M1 leads. Mechanical dedup follows. Template sync + regression verification close.

---

## §A. Context

Anthropic's 2026-07-24 "new rules of context engineering" guidance shifts absolute/prohibitive rule language toward judgement-delegating informational language. GOSS chose the **conservative "B-group only"** application: transition only the 2 clearly-stylistic absolutes (code-comments language + Tool Selection "Use X instead of Y"), preserve all mechanical guardrails (C-group: Multi-File Decomposition, Reproduction-First Bug Fix) and safety invariants (A-group: 66 Frozen markers + 5 anchor doctrines). This plan implements that decision via 4 milestones — expressive transition, Tool Selection consolidation, template neutralization, regression verification.

Pre-verified baseline (2026-07-28 direct grep, not handoff text):
- `CLAUDE.md` `[HARD]` x 15
- `.claude/rules/moai/` `[ZONE:Frozen]` x 66 across 13 files
- `.claude/rules/moai/` `[ZONE:Evolvable]` x 98
- `.claude/rules/moai/` `MUST` x 269 / `NEVER` x 14
- `moai-constitution.md` "Tool Selection Priority" block at lines ~106-117 (5-bullet "Use X instead of Y")
- `agent-common-protocol.md` "Tool Selection by Task" table at line ~231 (canonical detailed SSOT — unchanged)
- `moai-constitution.md:77` TRUST 5 Readable line "Clear naming, English comments" (M2.1 target)
- `plan-auditor.md:144` already delegates to SSOT (M1.3 — no defect)

---

## §B. Known Issues

None blocking. Two non-issues confirmed during baseline verification:

1. **M1.3 (plan-auditor.md)** — initially flagged in the handoff as "15-tool-name duplication"; verified 2026-07-28 that line ~144 already consolidates this to a single SSOT cross-reference. The 32 tool-name mentions in the surrounding audit-context prose are load-bearing (they describe the audit work, not duplicate guidance). **No edit required; falsifiable regression AC included.**
2. **CLAUDE.md Tool Selection** — initially suggested as an M1 dedup target; verified absent (grep returned no match). Out of scope per §F.

---

## §C. Pre-flight Baselines (capture before M1)

Run these once at M1 start to lock the baseline. Each milestone's verification references these numbers.

```bash
# Baseline 1a — Frozen total count (target: 66)
grep -rc '\[ZONE:Frozen\]' .claude/rules/moai/ | awk -F: '{s+=$2} END{print s}'

# Baseline 1b — Frozen per-file distribution (target: 13 files, sum 66)
# Captured at M1 start to a baseline file; M4 / AC-CER-005(b) verify via diff.
# Per `feedback_ac_command_vacuous_green_traps` + plan-auditor D1: the total
# count alone is not falsifiable — a silent removal in one file offset by a
# parallel-session addition elsewhere preserves the total. The per-file
# distribution closes that gap.
grep -rc '\[ZONE:Frozen\]' .claude/rules/moai/ | grep -v ':0$' | sort \
  > .moai/specs/SPEC-CONTEXT-ENGINE-RIGHTSIZE-001/baseline-frozen-distribution.txt
# Expected contents (2026-07-28 verified, 13 files / sum 66):
#   .claude/rules/moai/core/agent-common-protocol.md:1
#   .claude/rules/moai/core/askuser-protocol.md:3
#   .claude/rules/moai/core/moai-constitution.md:1
#   .claude/rules/moai/design/constitution.md:19
#   .claude/rules/moai/development/agent-authoring.md:1
#   .claude/rules/moai/development/branch-origin-protocol.md:7
#   .claude/rules/moai/development/sprint-round-naming.md:1
#   .claude/rules/moai/workflow/ci-autofix-protocol.md:10
#   .claude/rules/moai/workflow/ci-watch-protocol.md:8
#   .claude/rules/moai/workflow/orchestration-mode-selection.md:3
#   .claude/rules/moai/workflow/spec-workflow.md:3
#   .claude/rules/moai/workflow/worktree-integration.md:8
#   .claude/rules/moai/workflow/worktree-state-guard.md:1

# Baseline 2 — Evolvable count (target: 98)
grep -rc '\[ZONE:Evolvable\]' .claude/rules/moai/ | awk -F: '{s+=$2} END{print s}'

# Baseline 3 — moai-constitution.md "Use X instead of Y" bullets (target: 5)
grep -c '^- Use .* instead of' .claude/rules/moai/core/moai-constitution.md

# Baseline 4 — moai-constitution.md "English comments" (target: 1 line, may be 2 matches in surrounding prose)
grep -n 'English comments' .claude/rules/moai/core/moai-constitution.md

# Baseline 5 — plan-auditor.md SSOT cross-ref (target: observable)
grep -n 'agent-common-protocol.md.*Tool Selection by Task' .claude/agents/moai/plan-auditor.md

# Baseline 6 — repo-wide moai spec lint (record pre-edit state)
moai spec lint --json 2>/dev/null | tail -5 || echo "(moai spec lint not available — use go run ./cmd/moai spec lint)"
```

---

## §D. Constraints

1. **Conservative scope lock** — C-group items MUST NOT be touched. If any run-phase step proposes transitioning Multi-File Decomposition or Reproduction-First Bug Fix, return a blocker report and re-delegate to manager-spec.
2. **Template-first obligation** (CLAUDE.local.md §2 [HARD]) — every edit to `.claude/rules/moai/core/moai-constitution.md` MUST be mirrored to `internal/template/templates/.claude/rules/moai/core/moai-constitution.md` with §25 neutralization.
3. **No `make build` required** — `.claude/rules/` is not `//go:embed`-ed; only `internal/template/templates/` is. Template edits do not affect the compiled binary.
4. **Pathspec commits** (per `feedback_shared_checkout_concurrent_commit_race` and `feedback_index_level_commit_shared_checkout`) — commit each milestone via explicit pathspec, NOT `git add -A`, to avoid absorbing parallel-session uncommitted work on the shared main checkout.
5. **Plan-phase commit subject uses `feat()` prefix** (per `feedback_plan_commit_subject_feat_prefix`) — `feat(SPEC-CONTEXT-ENGINE-RIGHTSIZE-001): plan-phase artifacts (Tier M, 3 artifacts)`.
6. **Verification before claim** (per `feedback_claimed_correction_never_applied` + `verification-claim-integrity.md` §1.1 surface 2) — every milestone's "done" claim MUST cite a verbatim grep output, not a summary.

---

## §E. Self-Verification (per-milestone grep evidence)

Each milestone's self-verification block is co-located with the milestone in §F below (the canonical verification command + expected output). The §E matrix consolidates them for the manager-develop §E.1 deliverable.

---

## §F. Milestones

### M1 — Expressive transition: code_comments line (highest UX impact)

**Why first**: This line (`moai-constitution.md:77` TRUST 5 Readable: "Clear naming, English comments") is the single most user-visible rule in this SPEC. Every agent that comments code inherits it. Reversibility is high (semantic shift), so review focuses here.

**Edit target** (`.claude/rules/moai/core/moai-constitution.md` line ~77, inside the TRUST 5 Readable bullet):

Before:
```
- Readable: Clear naming, English comments
```

After:
```
- Readable: Clear naming; comments match the surrounding code's language and density (per `code_comments` setting in `.moai/config/sections/language.yaml`, default English)
```

**Semantic intent preserved**: the rule still promotes readable code; the unconditional "English" absolute is replaced with (a) deference to the existing config mechanism (already respected at `agent-common-protocol.md:126` and `CLAUDE.md:200`) and (b) Anthropic's "match the surrounding code" judgement-delegation framing.

**Self-verification (run after edit)**:
```bash
# 1. Old absolute form gone
grep -c 'Clear naming, English comments' .claude/rules/moai/core/moai-constitution.md
# Expected: 0

# 2. New config-respecting form present
grep -c 'match the surrounding code.*language and density' .claude/rules/moai/core/moai-constitution.md
# Expected: 1

# 3. Config mechanism explicitly referenced
grep -c 'code_comments.*language.yaml' .claude/rules/moai/core/moai-constitution.md
# Expected: 1
```

**Commit subject**: `feat(SPEC-CONTEXT-ENGINE-RIGHTSIZE-001): M1 code_comments expressive transition (TRUST 5 Readable)`

---

### M2 — Tool Selection consolidation + informational reframing

**Why second**: Tool Selection is the second B-group target (combining M1.1 dedup + M2.2 absolute→informational). Both edits share the same prose block in `moai-constitution.md`, so they land in one milestone.

**Edit target** (`.claude/rules/moai/core/moai-constitution.md` § Tool Selection Priority, lines ~106-117):

Before (5-bullet absolute "Use X instead of Y" block):
```markdown
## Tool Selection Priority

Use specialized tools over general alternatives.

Rules:
- Use Read instead of cat/head/tail
- Use Edit instead of sed/awk
- Use Write instead of echo redirection
- Use Grep instead of grep/rg commands
- Use Glob instead of find/ls
```

After (collapsed informational pointer + canonical SSOT cross-reference):
```markdown
## Tool Selection Priority

Prefer the dedicated tool over a general alternative when one is fit for purpose — it improves accuracy and reduces round-trip latency. The canonical tool-by-task table lives in `.claude/rules/moai/core/agent-common-protocol.md` § Tool Selection by Task (that table is the single source of truth; this section intentionally carries no duplicate list).
```

**Semantic intent preserved**: the guidance still points agents toward dedicated tools; the unconditional "Use X instead of Y" / "Use Grep instead of grep/rg" prohibitions are replaced with judgement-delegating "prefer the dedicated tool" + a single SSOT pointer. M1.1 dedup and M2.2 reframing collapse into this one edit.

**No edit** to `agent-common-protocol.md § Tool Selection by Task` — that table is the canonical SSOT and is unchanged (per §B.1 M1.1 "Canonical SSOT retained").

**Self-verification (run after edit)**:
```bash
# 1. "Use X instead of Y" 5-bullet block gone from moai-constitution.md
grep -c '^- Use .* instead of' .claude/rules/moai/core/moai-constitution.md
# Expected: 0

# 2. Single-line SSOT pointer present
grep -c 'agent-common-protocol.md.*Tool Selection by Task' .claude/rules/moai/core/moai-constitution.md
# Expected: 1

# 3. Judgement-delegating language present ("prefer" or "fit for purpose")
grep -Ec 'prefer the dedicated tool|fit for purpose' .claude/rules/moai/core/moai-constitution.md
# Expected: >= 1

# 4. agent-common-protocol.md § Tool Selection by Task table unchanged (canonical SSOT)
grep -c '^### Tool Selection by Task' .claude/rules/moai/core/agent-common-protocol.md
# Expected: 1

# 5. plan-auditor.md SSOT cross-reference survived (M1.3 regression guard)
grep -c 'agent-common-protocol.md.*Tool Selection by Task' .claude/agents/moai/plan-auditor.md
# Expected: >= 1
```

**Commit subject**: `feat(SPEC-CONTEXT-ENGINE-RIGHTSIZE-001): M2 Tool Selection consolidation (M1.1 dedup + M2.2 informational reframe)`

---

### M3 — Template mirror synchronization + §25 neutralization

**Why third**: Template sync is mechanical but mandatory (CLAUDE.local.md §2 [HARD]). It follows the source-edit milestones so the neutralization step has a stable source to mirror.

**Edit targets** (`internal/template/templates/.claude/rules/moai/core/`):
1. `moai-constitution.md` — apply the SAME M1 + M2 edits as the source file.
2. `agent-common-protocol.md` — no source-side change; mirror unchanged (SSOT canonical).

**§25 neutralization check** (per CLAUDE.local.md §25 + `feedback_local_template_sync_neutralize_first`):
- NO SPEC ID (`SPEC-CONTEXT-ENGINE-RIGHTSIZE-001`) introduced into the template mirror.
- NO REQ tokens (`REQ-CER-*`).
- NO audit citations ("per SPEC-CONTEXT-ENGINE-RIGHTSIZE-001", "Audit N Finding AX").
- The new prose ("match the surrounding code's language and density (per `code_comments` setting in `.moai/config/sections/language.yaml`, default English)") is generic and template-safe — it names a config file and a general principle, not an internal SPEC.
- The new SSOT pointer ("canonical tool-by-task table lives in `.claude/rules/moai/core/agent-common-protocol.md`") is a generic relative path — template-safe.

**Self-verification (run after edit)**:
```bash
# 1. Template mirror has the same M1 transition
grep -c 'match the surrounding code.*language and density' internal/template/templates/.claude/rules/moai/core/moai-constitution.md
# Expected: 1

# 2. Template mirror has the same M2 consolidation
grep -c 'agent-common-protocol.md.*Tool Selection by Task' internal/template/templates/.claude/rules/moai/core/moai-constitution.md
# Expected: 1

# 3. NO SPEC ID leaked into template (§25)
grep -rc 'SPEC-CONTEXT-ENGINE-RIGHTSIZE-001' internal/template/templates/.claude/rules/moai/core/
# Expected: 0

# 4. NO REQ token leaked into template (§25)
grep -rc 'REQ-CER-' internal/template/templates/.claude/rules/moai/core/
# Expected: 0

# 5. byte-parity between source and template for the edited regions
diff <(sed -n '75,82p' .claude/rules/moai/core/moai-constitution.md) <(sed -n '75,82p' internal/template/templates/.claude/rules/moai/core/moai-constitution.md)
# Expected: no diff (the moai-constitution.md line ~77 region matches)
```

**Commit subject**: `feat(SPEC-CONTEXT-ENGINE-RIGHTSIZE-001): M3 template mirror sync + §25 neutralization`

---

### M4 — Regression verification (A-group + C-group + lint parity)

**Why last**: Pure verification — no edits. Closes the SPEC by establishing that A-group Frozen count, C-group mechanical guardrails, and lint state are all preserved.

**Verification batch** (run as a single-turn multi-Bash parallel batch per `agent-common-protocol.md § Parallel Execution`):

```bash
# A-group Frozen total count preserved (REQ-CER-005, falsifiable (a))
grep -rc '\[ZONE:Frozen\]' .claude/rules/moai/ | awk -F: '{s+=$2} END{print s}'
# Expected: >= 66 (baseline was 66)

# A-group Frozen per-file distribution preserved (REQ-CER-005, falsifiable (b) — D1)
diff -u \
  .moai/specs/SPEC-CONTEXT-ENGINE-RIGHTSIZE-001/baseline-frozen-distribution.txt \
  <(grep -rc '\[ZONE:Frozen\]' .claude/rules/moai/ | grep -v ':0$' | sort)
# Expected: empty diff (no `<` / `>` lines — see AC-CER-005(b))

# A-group AskUserQuestion doctrines preserved (REQ-CER-006)
grep -Er 'AskUserQuestion-Only|Subagent Prohibitions|ToolSearch.*select:AskUserQuestion' .claude/rules/moai/core/askuser-protocol.md .claude/rules/moai/core/agent-common-protocol.md
# Expected: matches in BOTH files (non-zero per file)

# A-group safety invariants preserved (REQ-CER-007)
grep -LEr 'BRANCH_GUARD_VIOLATION|verification-claim-integrity|Native.*goal.*Prohibition' .claude/rules/moai/workflow/main-checkout-branch-guard.md .claude/rules/moai/core/verification-claim-integrity.md .claude/rules/moai/workflow/goal-directive.md
# Expected: empty (-L lists files WITHOUT match; empty means all 3 files still have the invariant)

# C-group Multi-File Decomposition preserved (REQ-CER-008)
grep -Ec 'Multi-File Change Decomposition|Multi-File Decomposition' CLAUDE.md
# Expected: >= 2 (HARD bullet + §7 Rule 2)

# C-group Reproduction-First Bug Fix preserved (REQ-CER-008)
grep -c 'Reproduction-First Bug Fix' CLAUDE.md
# Expected: >= 2 (HARD bullet + §7 Rule 4)

# M1.3 regression — plan-auditor SSOT reference survived
grep -c 'agent-common-protocol.md.*Tool Selection by Task' .claude/agents/moai/plan-auditor.md
# Expected: >= 1

# moai spec lint — no new findings attributable to M1/M2
moai spec lint --json 2>/dev/null | tail -20 || go run ./cmd/moai spec lint 2>&1 | tail -20
# Compare to §C Baseline 6; new findings referencing moai-constitution.md lines ~77 or ~106-117 = regression

# go test parity — no Go code touched, suite remains green
go test ./... 2>&1 | tail -5
# Expected: ok / PASS lines; no new FAIL
```

**Commit subject**: `chore(SPEC-CONTEXT-ENGINE-RIGHTSIZE-001): M4 regression verification (A-group + C-group + lint parity)`

---

## §G. Anti-Patterns

- **AP-CER-001 — Driven-by refactor**: While editing `moai-constitution.md`, the implementer notices an unrelated stale comment and "cleans it up". → Stay within the M1+M2 edit regions; drive-by refactors create noise and risk regressions outside the SPEC's falsifiable AC set.
- **AP-CER-002 — C-group scope creep**: The implementer, emboldened by the B-group transition, also softens Multi-File Decomposition or Reproduction-First Bug Fix "for consistency". → C-group is explicitly preserved per §F; any such proposal MUST return a blocker report for manager-spec re-delegation.
- **AP-CER-003 — Ghost AC for M1.3**: The implementer, reading the original handoff, "fixes" `plan-auditor.md` line ~144 by rewriting tool guidance inline. → M1.3 is verification-only (the SSOT ref already exists); rewriting it introduces the duplication M1.1 was meant to remove.
- **AP-CER-004 — Template blind-copy**: The implementer mirrors M1/M2 edits to the template WITHOUT running the §25 neutralization grep, then claims sync complete. → Blind copy is forbidden (`feedback_local_template_sync_neutralize_first`); the §25 grep (M3 self-verification commands 3-4) MUST be cited.
- **AP-CER-005 — Unobserved "66 preserved" claim**: The implementer reports "Frozen count preserved" without running the grep. → Per `verification-claim-integrity.md §1.1 surface 2`, every preservation claim MUST cite the verbatim grep output.
- **AP-CER-006 — `git add -A` on shared checkout**: The implementer uses `git add -A` to stage M1/M2/M3 changes. → Pathspec-only commits are mandatory (§D constraint 4); `git add -A` on the shared main checkout risks absorbing parallel-session work.

---

## §H. Cross-References

- **spec.md**: `.moai/specs/SPEC-CONTEXT-ENGINE-RIGHTSIZE-001/spec.md` (requirements + scope + A/C-group classification).
- **acceptance.md**: `.moai/specs/SPEC-CONTEXT-ENGINE-RIGHTSIZE-001/acceptance.md` (falsifiable AC matrix).
- **Schema SSOT**: `.claude/rules/moai/development/spec-frontmatter-schema.md` § Status Transition Ownership Matrix (`(none) → draft` owned by manager-spec via `feat(SPEC-{ID}): plan-phase artifacts ({tier}, {N} artifacts)`).
- **Template-First**: CLAUDE.local.md §2 [HARD] + §25 neutralization doctrine (`.moai/docs/template-internal-isolation-doctrine.md`).
- **Applied lessons (memory)**:
  - `feedback_claimed_correction_never_applied` — every edit verified by direct grep at edit time.
  - `feedback_defect_claim_verification` — M1.3 classified no-defect after verification.
  - `feedback_local_template_sync_neutralize_first` — M3 includes §25 neutralization grep.
  - `feedback_guard_observation_must_be_falsifiable` — A-group ACs use grep counts.
  - `feedback_plan_commit_subject_feat_prefix` — plan-phase commit uses `feat()` prefix.
  - `feedback_shared_checkout_concurrent_commit_race` — pathspec-only commits.
  - `feedback_index_level_commit_shared_checkout` — pathspec staging discipline.

---
