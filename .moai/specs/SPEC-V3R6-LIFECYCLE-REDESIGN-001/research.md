---
id: SPEC-V3R6-LIFECYCLE-REDESIGN-001
research_version: "0.2.0"
spec_version: "0.2.0"
status: draft
created: 2026-06-18
updated: 2026-06-19
author: manager-spec
tier: L
---

# Research — SPEC-V3R6-LIFECYCLE-REDESIGN-001

## §A. Research Questions

- RQ-1: What is the exact drift surface for Axis A (4-phase / Mx-phase / mx_commit_sha / §E.5 / Mx chore)?
- RQ-2: What is the exact naming-migration surface for Axis B (Sprint / cohort / Round / Wave)?
- RQ-3: How many existing SPECs are affected by the H-4 rewrite (era.go migration impact)?
- RQ-4: What is the progress.md §E layout distribution across the catalog?
- RQ-5: What does the de-facto SDD standard (GitHub Spec Kit) use as work-unit vocabulary?
- RQ-6: Does Addy Osmani's SDD article corroborate the Spec Kit vocabulary?

## §B. Methodology

All measurements were taken on 2026-06-18 against the working tree at `/Users/goos/MoAI/moai-adk-go` (branch `main`, pre-redesign). Tooling:
- `grep -rl` for file-level surface counts.
- `moai spec audit --json` for authoritative era classification (the canonical Go implementation).
- `mcp__web_reader__webReader` for the Spec Kit README fetch.
- Direct source inspection of `internal/spec/era.go` and `internal/spec/audit.go`.

Exclusion filter (applied to all grep counts unless noted): `.claude/worktrees/`, `.claude/agent-memory/`, `.moai/specs/`, `.moai/reports/`, `.moai/backups/`, `.moai/research/`, `.moai/plans/`, `.moai/logs/`, `.moai/state/`, `.claude/skills/harness/` (user-owned harness skills).

## §C. Findings

### §C.1 RQ-1 — Axis A Drift Surface (14 files)

**Command**: `grep -rl '4-phase\|Mx-phase\|mx_commit_sha\|§E\.5\|Mx chore\|Mx-phase Audit-Ready' .claude/ | grep -v '<exclusions>'

**Result**: 14 files.

Breakdown by category:
- **Core rules (6)**:
  - `.claude/rules/moai/core/verification-claim-integrity.md`
  - `.claude/rules/moai/development/spec-frontmatter-schema.md`
  - `.claude/rules/moai/development/agent-patterns.md`
  - `.claude/rules/moai/workflow/lifecycle-sync-gate.md`
  - `.claude/rules/moai/workflow/archived-agent-rejection.md`
  - `.claude/rules/moai/workflow/spec-workflow.md`
- **Agent definitions (5)**:
  - `.claude/agents/moai/manager-spec.md`
  - `.claude/agents/moai/manager-docs.md`
  - `.claude/agents/moai/manager-git.md`
  - `.claude/agents/harness/workflow-specialist.md`
  - (manager-develop.md touched indirectly via §E references)
- **Hooks (1)**: `.claude/hooks/moai/status-transition-ownership.sh`
- **Output styles (1)**: `.claude/output-styles/moai/moai.md`
- **Skills (2)**:
  - `.claude/skills/harness-moaiadk-patterns/SKILL.md`
  - `.claude/skills/moai/workflows/plan/spec-assembly.md`

### §C.2 RQ-2 — Axis B Naming-Migration Surface (102 files)

**Command**: `grep -rl 'Sprint [0-9]\|코호트\|cohort\|Round [0-9]\|Wave [0-9]\|스프린트\|라운드' .claude/ .moai/ | grep -v '<exclusions>'

**Result**: 102 files.

Breakdown by priority tier (per REQ-LR-015):
- **T1 — Canonical rule files (11)**: `.claude/rules/moai/` excluding sprint-round-naming.md (which is the SSOT to be rewritten in M6).
- **T2 — Agents and output-styles (~3)**: `.claude/output-styles/moai/moai.md`, selected agent definitions.
- **T3 — Skills (~7)**: `.claude/skills/moai-domain-database/SKILL.md`, `.claude/skills/moai-workflow-ci-loop/SKILL.md`, `.claude/skills/moai/workflows/{plan/clarity-interview,project/codebase-analysis,project/meta-harness,project/mode-detection}.md`.
- **T4 — Project docs (~10)**: `.moai/docs/{autonomous-workflow-strategy,git-workflow-doctrine,harness-delivery-strategy}.md`, `.moai/project/codemaps/{docs-truth,overview}.md`, `.moai/project/db/queries.md`, `.moai/release/*.md`, `.moai/brain/IDEA-*/**`, `.moai/scripts/status-drift-cleanup.go`.
- **T5 — Archived/historical (~71)**: `.moai/archive/skills/v2.16/**`, `.moai/archive/skills/v3.0/**`, `.moai/design/v3-legacy/**`, `.moai/design/v3-redesign/**`, `.moai/design/v3-research/**`, `.moai/design/SPEC-V3R3-CLI-TUI-001/**`, `.moai/design/web-console-handoff/**`. Best-effort migration; most are frozen historical artifacts.

### §C.3 RQ-3 — Era Migration Impact (V3R6 count is a MOVING baseline — D3)

**Command**: `moai spec audit --json`

> **Moving-baseline note (D3)**: the V3R6 count is NOT a frozen literal. At the original plan-phase measurement (2026-06-18) it was **48**; a re-measurement on 2026-06-19 reported **V3R6 = 53 / modern_era_clean = 77 / total_specs = 357**, and it is still rising because a parallel session is authoring more V3R6 SPECs concurrently. Every count below is therefore an **illustrative plan-phase snapshot**, NOT a mechanical pass condition. The authoritative baseline `N` MUST be **captured at run-phase M1 start** (`moai spec audit --json`); all count-dependent acceptance criteria assert **invariance** (post-migration V3R6 count == the M1-captured baseline N), never equality to a frozen literal. See acceptance.md AC-LR-003 and §E quality gate.

**Result (illustrative snapshot — 2026-06-18; re-measure at M1)**:

| Metric | Value (e.g. as of 2026-06-18; 2026-06-19 in parens) |
|--------|-------|
| total_specs | 351 (357 on 2026-06-19) |
| grandfathered (V2.x + V3R2-R4 + V3R5) | 270 (271) |
| modern_era_clean (V3R6 H-4 satisfied) | 75 (77) |
| drift_findings total | 319 |
| V3R6 drift findings (H-4 detected) | 48 (53) |
| MUST-FIX findings | 6 |
| EraAutoDetected findings | 312 |
| Y_Y_N_Y findings | 4 |
| Y_Y_Y_Y_StatusDrift findings | 2 |
| AuditError | 1 |

Era distribution (from drift_findings):
- V2.x: 145 (H-1, progress.md absent)
- V3R2-R4: 118 (H-2, progress.md without §E.* markers)
- V3R5: 7 (H-3, §E.2 present but sync_commit_sha missing)
- V3R6: 48 (H-4, full predicate)
- unclassified: 1

**The V3R6 SPECs (N≈53 as of 2026-06-19, re-measure at M1) are the migration-affected set.** Correction (D1): rewriting H-4 to drop the `§E.5 + mx_commit_sha` requirement WITHOUT a migration window does **NOT** reclassify them to V3R5 (the H-3 empty-sync_sha condition does not match a SPEC with a populated sync_sha — see §D.3). The genuine regression vector is **H-6 unclassified**, which the H-5 date/phase heuristic already prevents for the entire current V3R6 population (the re-derived H-6 at-risk set is empty — §D.4). The dual-predicate window (REQ-LR-006) + auto-fold backfill (REQ-LR-007 / M3) remain valuable as defense-in-depth + classification-rationale precision (and as future-proofing against a not-yet-authored V3R6 SPEC with a pre-`modernEraThreshold` `created:` and no modern `phase:`), but they are NOT correctness-critical for the current catalog.

### §C.4 RQ-4 — progress.md §E Layout Distribution

**Commands**: `grep -rl '§E\.[2345]' .moai/specs/SPEC-*/progress.md`

| §E section | Files containing it |
|------------|---------------------|
| §E.2 (run-evidence) | 82 |
| §E.3 (run-audit) | 71 |
| §E.4 (sync) | 59 |
| §E.5 (Mx) | 83 |
| Total progress.md files | 206 |

Interpretation (counts are illustrative plan-phase snapshots — moving baseline per D3; re-measure at M1):
- ~83 SPECs carry the `§E.5` marker — the 5-section layout.
- ~82 carry `§E.2` — the run-evidence start marker (the H-2/H-3/H-4 discriminator).
- ~59 carry `§E.4` — the sync section (will become the sole sync marker after redesign).
- The V3R6 SPECs (RQ-3) are the subset of the `§E.5`-bearing SPECs that ALSO have populated `sync_commit_sha` AND `mx_commit_sha` (the full legacy H-4 predicate); of the current 53, **29** already carry `§E.4`+`sync_sha` (caught directly by the NEW H-4) and **11** are legacy-layout (`§E.5`+`mx_sha`, no `§E.4`) relying on the legacy-fallback/H-5 — see §D.4.

### §C.5 RQ-5 — GitHub Spec Kit Vocabulary (De-facto SDD Standard)

**Source**: `https://github.com/github/spec-kit` (fetched 2026-06-18 via `mcp__web_reader__webReader`).

**Verbatim Spec Kit Core Commands**:

| Command | Agent Skill | Description (verbatim) |
|---------|-------------|------------------------|
| `/speckit.constitution` | `speckit-constitution` | Create or update project governing principles and development guidelines |
| `/speckit.specify` | `speckit-specify` | Define what you want to build (requirements and user stories) |
| `/speckit.plan` | `speckit-plan` | Create technical implementation plans with your chosen tech stack |
| `/speckit.tasks` | `speckit-tasks` | Generate actionable task lists for implementation |
| `/speckit.taskstoissues` | `speckit-taskstoissues` | Convert generated task lists into GitHub issues for tracking and execution |
| `/speckit.implement` | `speckit-implement` | Execute all tasks to build the feature according to the plan |
| `/speckit.converge` | `speckit-converge` | Assess the codebase against spec/plan/tasks and append remaining work as new tasks |

**Spec Kit Core Philosophy (verbatim)**:
> Spec-Driven Development is a structured process that emphasizes:
> - Intent-driven development where specifications define the "what" before the "how"
> - Rich specification creation using guardrails and organizational principles
> - Multi-step refinement rather than one-shot code generation from prompts
> - Heavy reliance on advanced AI model capabilities for specification interpretation

**Vocabulary present in Spec Kit**: `spec`, `specifications`, `plan`, `tasks`, `constitution`, `feature`, `branch`, `integration`, `agent skill`, `skill`, `preset`, `extension`.

**Branch naming convention (verbatim example)**: `001-create-taskify` — sequential numeric prefix tied to the spec directory `specs/001-create-taskify/`.

**Vocabulary ABSENT from Spec Kit README** (verified by reading the full fetched content):
- `Sprint` — ABSENT
- `Cohort` — ABSENT
- `Wave` — ABSENT
- `Round` — ABSENT
- `Epic` — ABSENT

**Implication**: MoAI's `SPEC` and `skill` naming already align with Spec Kit. `Sprint`/`cohort`/`Round`/`Wave` are legacy Agile terms unaligned with SDD. `Epic` is MoAI's chosen replacement for `Sprint` — it is NOT borrowed from Spec Kit (which has no multi-feature grouping term beyond "project"), but it is an industry-recognized Agile/SAFe term that does not conflict with SDD vocabulary.

### §C.6 RQ-6 — Addy Osmani SDD Article

**Status**: CORROBORATION INCOMPLETE. The `mcp__web_reader__webReader` call to the candidate Addy Osmani URL returned HTTP 500 (network error, error id `20260618220403126908d1060f4309_call_5cb8cacb5d3144de9d249ece`). Per the verification-claim-integrity doctrine, I will NOT cite a specific Addy Osmani article URL or specific principles without a verified fetch. The user's context referenced "Addy Osmani's 5 principles + 6 core areas" — this is noted as an unverified claim and is NOT load-bearing for this SPEC. The SPEC relies on the GitHub Spec Kit citation (RQ-5, verified) as its primary industry anchor.

**Deferred to run phase**: a follow-up web search to locate and verify the Addy Osmani SDD article, IF the orchestrator/user wants secondary corroboration. The SPEC's Axis B redesign stands on the Spec Kit citation alone.

## §D. Analysis

### §D.1 Migration Strategy Selection (era.go H-4)

Three strategies were considered:

| Strategy | Mechanism | Pros | Cons |
|----------|-----------|------|------|
| **S1: Auto-migrate (fold §E.5 → §E.4)** | One-time backfill script folds §E.5 content into §E.4 for the 48 SPECs | Clean end state; single predicate | Touches 48 progress.md files; requires migration log |
| **S2: Grandfather the 48** | Treat the 48 as "V3R6-legacy" protected from the new predicate | No file mutation | Permanent bifurcation; drift detector complexity |
| **S3: Dual-predicate window (permanent)** | H-4 accepts EITHER new OR legacy predicate forever | No migration; no file mutation | Permanent legacy support; defeats the redesign purpose |

**Selected**: **S1 (auto-migrate)** + a temporary S3 window during the migration. Rationale:
- S1 produces the clean end state (4-section layout everywhere).
- The temporary S3 window (REQ-LR-006) prevents misclassification between M1 (code rewrite) and M3 (backfill).
- After M3, the S3 window can be retired (or kept as a defensive fallback).
- S2 is rejected because it permanently bifurcates the V3R6 population.
- Pure S3 is rejected because it defeats the redesign.

### §D.2 Epic Taxonomy Mapping

| Old term | New term | Semantic mapping | Disposition |
|----------|----------|------------------|-------------|
| Sprint | Epic | Multi-SPEC grouping (time-unit or thematic) | Renamed |
| cohort | (folded into Epic) | Epic-internal sub-grouping (e.g. "Epic N Lane A") | Removed as standalone term |
| Round | (folded into Milestone) | Within-SPEC SSE-stall sub-division | Removed; use Milestone |
| Wave | (retired) | Legacy pre-Round term (per AP-SRN-004) | Removed |
| Milestone | Milestone (retained) | Within-SPEC ordered step (M1, M2, ...) | Unchanged |
| SPEC | SPEC (retained) | Single work unit | Unchanged (SDD-aligned) |
| Constitution | Constitution (new) | Project-level governance (Spec Kit alignment) | Added as canonical term |

**Final canonical taxonomy (4 terms)**: Epic, SPEC, Milestone, Constitution.

### §D.3 Worked Example — Era Reclassification Trace (CORRECTED — verified against `internal/spec/era.go` lines 102-146)

> **Correction note (D1)**: an earlier draft of this trace claimed that after the H-4 rewrite a legacy-layout SPEC would match **H-3 → V3R5 (regression)**. That was FALSE. H-3 (`era.go:130`) is `if hasSyncSection && syncSHA == ""` — it fires ONLY when `sync_commit_sha` is **empty**; a SPEC carrying a non-empty `sync_commit_sha` does NOT match H-3. The true fall-through is **H-5** (`era.go:140`, `matchesModernPhase(phase) || isAfterModernThreshold(created)` → V3R6), and only if H-5 also misses does control reach **H-6 → unclassified**. The corrected trace below matches the actual `ClassifyEra` control flow.

Consider a hypothetical SPEC `SPEC-EXAMPLE-001` with the legacy 5-section progress.md and a modern `created:` date (`>= 2026-04-01`):

**Before redesign (current H-4, `era.go:135`)**:
- §E.2 present, §E.5 present, sync_commit_sha = "abc123", mx_commit_sha = "def456"
- H-4 matches (`hasSyncSection && hasMxSection && syncSHA != "" && mxSHA != ""`) → **V3R6**

**After M1 (new H-4, NO migration window — what actually happens)**:
- New H-4 predicate (`§E.2 + §E.4 + sync_sha`): NO (§E.4 absent) → H-4 misses.
- H-3 (`§E.2 present && sync_sha == ""`): NO (sync_sha = "abc123" is non-empty) → **H-3 does NOT match** (the old "regression to V3R5" claim was wrong here).
- H-5 (`matchesModernPhase(phase) || isAfterModernThreshold(created)`): YES (created >= 2026-04-01) → **V3R6 (via H-5 date heuristic — NO regression).**

**After M1 (new H-4, WITH dual-predicate window, REQ-LR-006 — preferred rationale)**:
- New H-4 predicate: NO (§E.4 absent).
- Legacy fallback predicate (`§E.2 + §E.5 + sync_sha + mx_sha`): YES → **V3R6 (via legacy predicate — an explicit, precise rationale, beating the H-5 date-heuristic fallback).**

**After M3 (backfill folds §E.5 → §E.4)**:
- §E.2 present, §E.4 present (folded), sync_commit_sha = "abc123".
- New H-4 predicate (`§E.2 + §E.4 + sync_sha`): YES → **V3R6 (via new predicate, clean end state).**

**The genuine H-6 at-risk edge case** (the ONLY scenario that produces a real regression):
- A V3R6 SPEC that lacks `§E.4`, lacks the legacy `§E.5`+`mx_sha` predicate, AND has neither a modern `phase:` (matching `v3r6`/`v3.0*`) NOR `created >= 2026-04-01`.
- New H-4: NO. H-3: NO (if sync_sha non-empty). Legacy fallback: NO. H-5: NO (pre-threshold date, no modern phase). → **H-6 → unclassified (the genuine regression).**
- **Empirically (plan-phase snapshot 2026-06-18): this at-risk set is EMPTY** — see §D.4. Every current V3R6 SPEC carries `created >= 2026-04-01` and/or a modern `phase:`, so H-5 catches all of them even without the migration window. Because the catalog is moving (a parallel session is authoring more V3R6 SPECs), this MUST be re-measured at run-phase M1 start.

This corrected trace shows the dual-predicate window (REQ-LR-006) is **defense-in-depth + rationale precision**, not misclassification prevention; the genuine regression vector is H-6 unclassified (currently empty), not H-3 V3R5.

### §D.4 Re-derived At-Risk Set (D1 — empirical, plan-phase snapshot 2026-06-18)

Measured via `moai spec audit --json` (for the V3R6 set) + per-SPEC frontmatter inspection of `phase:`/`created:` against the H-5 predicate (`matchesModernPhase` OR `created >= modernEraThreshold "2026-04-01"`):

| Disposition after H-4 rewrite | Count (illustrative — e.g. N≈53 V3R6 as of plan-phase) |
|-------------------------------|------------------|
| Caught directly by NEW H-4 (`§E.2 + §E.4 + sync_sha`) | 29 |
| Lacks `§E.4`; carry legacy `§E.5`+`mx_sha` → rely on legacy-fallback OR H-5 | 11 |
| H-5 safety net coverage (modern `phase:` AND/OR `created >= 2026-04-01`) | all 53 (45 phase+date, 8 date-only, 0 phase-only) |
| **Genuine H-6 at-risk (fall to unclassified)** | **0** |

Reproduction command (re-run at M1 to capture the current N — the literal 53 is illustrative, NOT a frozen pass condition):
```bash
moai spec audit --json | python3 -c '
import json,sys,os,re
d=json.load(sys.stdin)
v3r6=[f["spec_id"] for f in d["drift_findings"] if f.get("era")=="V3R6"]
at_risk=[]
for sid in v3r6:
    p=f".moai/specs/{sid}/spec.md"
    if not os.path.exists(p): continue
    fm=re.search(r"^---\n(.*?)\n---", open(p,encoding="utf-8",errors="replace").read(), re.S)
    fm=fm.group(1) if fm else ""
    phase=next((l.split(":",1)[1].strip().strip(chr(34)) for l in fm.splitlines() if l.strip().startswith("phase:")), "")
    created=next((l.split(":",1)[1].strip().strip(chr(34)) for l in fm.splitlines() if l.strip().startswith("created:")), "")
    pl=phase.lower()
    modern = ("v3r6" in pl) or pl.startswith("v3.0") or (created >= "2026-04-01" if created else False)
    if not modern: at_risk.append(sid)
print("V3R6 total:", len(v3r6), "| genuine H-6 at-risk:", len(at_risk), at_risk)
'
```

## §E. Open Questions

- OQ-1: Should the dual-predicate window (REQ-LR-006) be permanent (defensive) or retired after M3? **Deferred to plan-auditor / user.** Recommendation: retire after M3 + 1 release cycle; keep as defensive fallback if the audit shows any residual edge cases.
- OQ-2: Should `Constitution` (REQ-LR-012) be accompanied by a new `/moai constitution` slash command (Spec Kit alignment)? **Deferred to follow-up SPEC** (EX-6). This SPEC references the term for vocabulary alignment only.
- OQ-3: Should the 270 grandfather-protected SPECs be offered an optional opt-in migration to the 4-section layout? **Recommendation: NO** (N4 + grandfather clause). They remain era-protected.
- OQ-4: The Addy Osmani SDD article (RQ-6) could not be verified — is secondary corroboration required, or is the Spec Kit citation sufficient? **Recommendation: Spec Kit citation is sufficient** for Axis B; Addy Osmani corroboration is a nice-to-have for research.md completeness.
- OQ-5: The naming-migration surface (102 files) includes ~71 archived/historical files (T5) — should these be migrated at all? **Recommendation: NO** (frozen historical artifacts; best-effort only if actively referenced).

## §F. Sources

- **GitHub Spec Kit README**: `https://github.com/github/spec-kit` — fetched 2026-06-18 via `mcp__web_reader__webReader`. Cited verbatim in §C.5.
- **`moai spec audit --json`**: canonical era classification (internal Go implementation `internal/spec/audit.go`). Run 2026-06-18 against the working tree.
- **`internal/spec/era.go`**: H-1..H-6 heuristic table source. Lines 86-146 (ClassifyEra) + lines 167-178 (hasAnyProgressMarker).
- **`internal/spec/audit.go`**: `checkV3R6Drift` at lines 224-300; THREE §E.5-keyed findings — `Y_Y_Y_Y_StatusDrift` (251-266), `Y_Y_N_Y` (268-282), `Y_N_N_Y` (284-297); `FindingType` constants at lines 54-65 (verified D2).
- **`internal/spec/transitions.go`**: `closeInfix4Phase = "4-phase close"` (line 74) + `closeInfixMatch` (line 80) — the drift walker's only positive `completed` signal (verified D4).
- **Addy Osmani SDD article**: NOT verified (HTTP 500 on fetch). Noted as open question OQ-4.

## §G. Reproducibility

All commands required to reproduce the findings:

```bash
# RQ-1 (Axis A drift surface)
grep -rl '4-phase\|Mx-phase\|mx_commit_sha\|§E\.5\|Mx chore\|Mx-phase Audit-Ready' .claude/ \
  | grep -v '\.claude/worktrees/\|\.claude/agent-memory/\|\.claude/skills/harness/' | sort -u | wc -l

# RQ-2 (Axis B naming surface)
grep -rl 'Sprint [0-9]\|코호트\|cohort\|Round [0-9]\|Wave [0-9]\|스프린트\|라운드' .claude/ .moai/ \
  | grep -v '\.claude/worktrees/\|\.claude/agent-memory/\|\.moai/specs/\|\.moai/reports/\|\.moai/backups/\|\.moai/research/\|\.moai/plans/\|\.moai/logs/\|\.moai/state/\|\.claude/skills/harness/' | sort -u | wc -l

# RQ-3 (Era migration impact)
moai spec audit --json | python3 -c "import json,sys; d=json.load(sys.stdin); print('V3R6:', sum(1 for f in d['drift_findings'] if f.get('era')=='V3R6'))"

# RQ-4 (progress.md §E layout)
for marker in '§E\.2' '§E\.3' '§E\.4' '§E\.5'; do
  echo -n "$marker: "; grep -rl "$marker" .moai/specs/SPEC-*/progress.md 2>/dev/null | wc -l
done

# RQ-5 (Spec Kit fetch — manual via mcp__web_reader__webReader)
# URL: https://github.com/github/spec-kit
```
