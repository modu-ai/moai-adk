# SPEC-SEC-DEEPSCAN-001 — Design

## §S Surface Decision (why `/moai review --deep`, not `/moai security`)

Three candidate surfaces were considered:

| Candidate | Verdict | Reason |
|-----------|---------|--------|
| New `/moai security` subcommand | **REJECTED** | `/moai security` was RETIRED by SPEC-SUBCOMMAND-RETIRE-001 (completed). Its retirement REQ explicitly re-routes security-audit requests to `/moai review --security` + `Agent(general-purpose)`. Reviving it contradicts a completed decision and would re-introduce the retired intent-router row. |
| New top-level `/moai deepscan` | **REJECTED** | A brand-new subcommand adds intent-router surface, a new command wrapper, and docs-site burden for a capability that is a rigor-upgrade of an existing lens. Higher cost, no reuse of the `/moai review` scope machinery. |
| `--deep` mode on `/moai review` | **CHOSEN** | Reuses the existing scope-selection (`--repo` / `--staged` / `--branch` / `--file`), the Security perspective, and the sync-auditor skeptical synthesis. Additive and non-breaking. Aligns with `native-invocation-model.md` Axis A (compose over reinvent) — `/moai review` already composes a security pass into one pipeline step. |

**Composition semantics**: `--deep` is a *depth* modifier on the existing *breadth* review. `--deep` alone ⇒ security-focused deep scan. `--security --deep` ⇒ explicit. `--deep` may combine with a scope flag (`--repo` = whole tree; `--branch B` / `--staged` / new `--commit SHA` = diff scopes). `--patch` is an independent opt-in that enables phase 6.

## §A Architecture — Playbook-drives-Workflow topology

```
/moai review --deep [--patch] [--repo|--branch B|--commit SHA|--staged|--file P]
   │  (thin command: .claude/commands/moai/review.md — Skill routing, <20 lines)
   ▼
moai SKILL.md Intent Router (review row — UNCHANGED; no new "security" row)
   │
   ▼
workflow skill: .claude/skills/moai/workflows/review.md  ← --deep Mode playbook (NEW section)
   │  Orchestrator reads playbook, collects prefs via AskUserQuestion (scope, --patch, degradation)
   │  Orchestrator selects execution path (availability-gated):
   │     PRIMARY  → Workflow() (Mode 6) runtime-constructed from the playbook
   │     FALLBACK → Mode-4 bounded parallel fan-out (3-5 concurrent Agent())
   │     DEGRADED → single-pass /moai review --security + native /security-review
   ▼
6-phase pipeline (below) — reuses sync-auditor, Agent(general-purpose) security reviewers,
   Agent(isolation:"worktree") for patch drafting, Skill() for hunt-phase domain knowledge
   ▼
results: .moai/reports/security-deepscan-<timestamp>/  (report.md + findings.jsonl + revision.json + .gitignore [+ patches/])
```

Key design principle: **the shipped artifact is the playbook (template-safe markdown), NOT a `.js` script.** Per `dynamic-workflows.md`, "Claude writes the script for the task" and `.claude/workflows/` is user-owned / not template-managed. So the orchestrator constructs the Workflow at runtime from the playbook's phase descriptions. A user MAY save the generated script into their own `.claude/workflows/` for reuse, but MoAI ships no static script (REQ-SDS-062). This resolves the template-neutrality tension cleanly.

## §B Six-Phase Pipeline Design

Phases map onto `pipeline()` (sequential) with `parallel()` fan-out inside phases 3-4.

| # | Phase | Primitive shape | Agent role | Tools / scope | Skill injection |
|---|-------|-----------------|------------|---------------|-----------------|
| 1 | Architecture map | single read-only agent (or Explore) | recon | Read/Grep/Glob — READ-ONLY | — |
| 2 | Threat model | single agent, consumes phase-1 map | threat-modeler | Read/Grep — READ-ONLY | `moai-ref-owasp-checklist` (baseline) |
| 3 | Vulnerability hunt | `parallel()` — N hunt agents (per-area / per-manifest) | hunter | Read/Grep/Glob — READ-ONLY | per-area: `moai-ref-owasp-checklist` / `moai-ref-llm-security` / `moai-ref-secops` / `moai-ref-supply-chain` (`Skill()`, on-demand) |
| 4 | Adversarial verification | `parallel()` — 3 voters PER candidate finding | REACHABILITY / IMPACT / DEFENSES voters | Read/Grep — READ-ONLY | — |
| 5 | Report | orchestrator (or synthesizer agent) | reporter | Write to results dir ONLY | — |
| 6 | Patch (gated by `--patch`) | `Agent(isolation:"worktree")` drafter + independent reviewer | patch-author + patch-reviewer | drafter: Write in SCRATCH clone only; reviewer: read-only | — |

Phases 1-4 are strictly read-only against the user's working tree (REQ-SDS-012). Phase 5 writes only under the results dir. Phase 6 writes only inside the isolated scratch clone and emits `.patch` artifacts into the results dir — never the live tree.

### §B.1 Determinism note (Workflow primitive constraint)

When the orchestrator constructs the Workflow script, the script body MUST be deterministic (no `Date.now()`/`Math.random()` CALL in the body — per `dynamic-workflows.md`). The timestamp for the results-dir name and the revision stamp is injected via the script's input args OR stamped onto the results AFTER the run returns — never generated inside the script body. This keeps resume-caching valid.

### §B.2 Purpose-driven effort per phase (dynamic-workflows.md taxonomy)

| Phase | Purpose | Recommended (model, effort) |
|-------|---------|-----------------------------|
| 1 Architecture map | read-only-extract | (haiku, low) |
| 2 Threat model | synthesize | (sonnet, high) |
| 3 Hunt | research | (sonnet/opus, high/xhigh) |
| 4 Adversarial voters | verify-judge | (sonnet/opus, xhigh) |
| 5 Report | synthesize | (sonnet, high) |
| 6 Patch author | implement | (sonnet/opus, xhigh) |
| 6 Patch reviewer | verify-judge | (sonnet/opus, xhigh) |

## §C Adversarial Panel Mapping (absorbed 3-voter panel → MoAI adversarial-verify)

The absorbed 3-voter panel maps 1:1 onto the dynamic-workflow built-in adversarial-verify pattern (independent REFUTE-skewed voters, 2-of-3 quorum, perspective-diverse verify).

```
candidate finding F_i (from phase 3)
   │
   ├── parallel() ──► Voter REACHABILITY: "Is the vulnerable code path actually reachable
   │                    by an attacker-controlled input?"  → affirm | refute
   ├──────────────► Voter IMPACT:        "If reached, what is the concrete blast radius?
   │                    Is the stated impact real, not theoretical?" → affirm | refute
   └──────────────► Voter DEFENSES:      "Do existing defenses (validation, authz, framework
                        guards) already neutralize this?"  → affirm (undefended) | refute (defended)
   │
   ▼ quorum reduction (in-script or orchestrator)
   affirm_count = #voters affirming F_i
   IF affirm_count >= 2  → ADMIT F_i to report
        IF affirm_count == 3 → confidence MAY be "high"
        IF affirm_count == 2 → confidence CAPPED at "medium"   (REQ-SDS-021)
   IF affirm_count <  2  → REJECT F_i → "unconfirmed candidates" appendix only (REQ-SDS-023)
```

**Voter independence (REQ-SDS-022)**: each voter is a separate `Agent()` spawn (in the degraded Mode-4 path) or a separate workflow `agent()` (primitive path), given ONLY the finding claim + the code context — NOT the hunt agent's reasoning chain. Voters are REFUTE-skewed: their default posture is "disprove this finding", which is the skeptical-evaluation stance sync-auditor already embodies (`agent-common-protocol.md` § Skeptical Evaluation Stance).

**Precedent**: the local `.claude/workflows/sync-audit-4dim.js` (4 parallel read-only judges + in-script harmonic-mean verdict) is the structural precedent for a per-item parallel-judge-then-reduce pattern. The deep-scan panel is the same shape with a 2-of-3 quorum reduction instead of a harmonic mean.

## §D Patch Design (scratch clone + reviewer vouch)

```
confirmed finding F_i  (only when --patch present)
   │
   ├── Agent(isolation:"worktree")  [DRAFTER]
   │     - operates in an isolated scratch copy (L1 worktree)
   │     - drafts the minimal fix for F_i ONLY
   │     - emits a unified diff (F<i>.patch), never touches the live tree
   │
   └── Agent()  [REVIEWER — independent spawn, distinct from DRAFTER]
         - reads F<i>.patch + the finding + surrounding code (read-only)
         - vouches for THREE claims (all-or-nothing):
             (a) addresses only the one finding F_i
             (b) introduces no new vulnerability
             (c) leaves behavior otherwise unchanged
         - all 3 vouched  → write F<i>.patch into results/patches/  (REQ-SDS-033: never applied)
         - any 1 unvouched → write a short note (results/patches/F<i>.note.md) instead  (REQ-SDS-032)
```

Invariants:
- DRAFTER and REVIEWER are **different spawns** (independence). The playbook MUST make this a HARD instruction.
- The pipeline NEVER runs `git apply`, `git add`, `git commit`, or `git push` against the user's repo (REQ-SDS-033). Each vouched finding yields exactly one `.patch` file the user applies manually (one finding = one patch = one PR).
- The scratch clone is disposed after drafting (`moai worktree done` semantics via the runtime); no residue in the live tree.

## §E Results Directory Schema

```
.moai/reports/security-deepscan-<YYYYMMDD-HHMMSS>/
├── .gitignore          # contains a single line: *   (REQ-SDS-044 — stray git add cannot sweep it)
├── report.md           # human report (REQ-SDS-041)
├── findings.jsonl      # machine-readable, one finding per line (REQ-SDS-042)
├── revision.json       # revision stamp (REQ-SDS-043)
└── patches/            # only when --patch; F<i>.patch or F<i>.note.md per confirmed finding
```

### report.md — per confirmed finding
```
### F<i> — <title>
- Severity:   <critical|high|medium|low>
- Confidence: <high|medium>          (medium max when panel non-unanimous — REQ-SDS-021)
- Impact:     <concrete blast radius>
- Exploit scenario: <step-by-step reachability narrative>
- Recommendation: <fix direction>
- Panel: REACHABILITY=<affirm|refute> IMPACT=<...> DEFENSES=<...> (<N>/3)
```
Rejected candidates appear ONLY under a trailing `## Unconfirmed candidates (did not reach 2-of-3)` appendix (REQ-SDS-023).

### findings.jsonl — one line per confirmed finding
```json
{"id":"F1","severity":"high","confidence":"medium","title":"...","impact":"...","exploit":"...","recommendation":"...","panel":{"reachability":true,"impact":true,"defenses":false},"location":{"path":"...","hint":"..."}}
```

### revision.json — revision stamp
```json
{"scanned_commit":"<sha>","effort_tier":"<low|medium|high|xhigh|max>","working_tree_included":true,"scope":"repo|branch|commit|staged|file","generated_at":"<injected-timestamp>"}
```

## §F Graceful Degradation Ladder

Availability is checked BEFORE launch (workflow agents cannot prompt mid-run — the degradation choice is made up front, EC-5).

| Rung | Condition | Path | Rigor preserved? |
|------|-----------|------|------------------|
| PRIMARY | Dynamic Workflows available (v2.1.154+, not disabled) | `Workflow()` Mode 6 — full fan-out (16 concurrent / 1000 total) | full |
| FALLBACK | Workflows unavailable BUT Mode-4 viable | Mode-4 bounded parallel (3-5 concurrent `Agent()`), findings batched | **2-of-3 quorum preserved** (REQ-SDS-052) — only scale drops |
| DEGRADED | Neither viable | single-pass `/moai review --security` + native `/security-review` (Axis A reuse) | reduced (single-pass, no per-finding panel) — clearly labeled in the report |

Availability signal(s): runtime version < v2.1.154, `CLAUDE_CODE_DISABLE_WORKFLOWS=1`, or `disableWorkflows: true` ⇒ not PRIMARY. The playbook names these observable signals (REQ-SDS-050).

## §G Distribution Design (run-phase order of operations)

1. Template tree FIRST: edit `internal/template/templates/.claude/skills/moai/workflows/review.md` (`--deep` section) → edit `internal/template/templates/.claude/commands/moai/review.md.tmpl` (`argument-hint`).
2. Local sync: byte-identical copy for the workflow skill; rendered command wrapper.
3. Optional: a one-line note in `.moai/config/sections/delegation.yaml` (per-phase `Skill()` injection documented; static preload UNCHANGED).
4. `make build` re-embeds; no catalog/agent-count change (REQ-SDS-062).
5. Neutrality + parity verification (plan.md §E).

## §H Test/CI Mapping

| Gate | Proves |
|------|--------|
| `TestCommandsThinPattern` / `TestCommandsFrontmatterConsistency` | command wrapper stays thin + valid argument-hint |
| `diff` template ↔ local review.md | AC-SDS-060 byte-parity |
| `internal_content_leak_test.go` + `template-neutrality-check.yaml` | AC-SDS-061 neutrality |
| `expectedAgentCount` / `expectedSkillCount` unchanged | AC-SDS-062 (no new agent/skill) |
| grep suite in acceptance.md §D.2 | AC-SDS-001..052 content presence |
| `go build ./...` + `go test ./internal/template/...` | repo-wide regression freedom |
