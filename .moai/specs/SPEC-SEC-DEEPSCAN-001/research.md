# SPEC-SEC-DEEPSCAN-001 — Research

Codebase research performed 2026-07-24 against the live tree. All findings are grounded in observed files, not memory.

## §A Current `/moai review` surface (extension point)

- **Workflow skill**: `.claude/skills/moai/workflows/review.md` (20961 bytes) + byte-identical template sibling `internal/template/templates/.claude/skills/moai/workflows/review.md`. `user-invocable: false`, version 2.5.0, triggers keywords `["review","code review","security audit","quality check","code analysis"]`, agents `["sync-auditor"]`.
- **Command wrapper**: `.claude/commands/moai/review.md` (thin, `allowed-tools: Skill`, `argument-hint: "[--staged] [--branch] [--security]"`) rendered from `internal/template/templates/.claude/commands/moai/review.md.tmpl`.
- **Existing flags**: `--staged`, `--branch BRANCH`, `--security`, `--file PATH`, `--design`, `--critique`, `--lean`, `--repo`.
- **Existing execution shape**: Phase 2 is a Mode-4 parallel read-only fan-out (4 perspectives: Security / Performance / Quality / UX, ≤4 concurrent judges, within the 3-5 ceiling); **sync-auditor is the binding synthesis + verdict owner**. Per-perspective `Skill()` injection is already the pattern (Security → `moai-ref-owasp-checklist`).
- **Existing security depth**: Perspective 1 already carries an OWASP check, a dependency-vulnerability sub-scan (manifest enumeration + per-spawn `Agent(general-purpose)` security reviewer), an **incremental secrets scan with a `.moai/state/secrets-scan-checkpoint.txt` checkpoint**, and a data-isolation check. `/moai review` is explicitly **read-only, report-only** and "layered under `/moai loop`" as a queue supplier.
- **Consequence for this SPEC**: `--deep` is additive. It reuses the scope-selection (Phase 1), the sync-auditor synthesis, and the security ref-skill injection pattern already present. It does NOT modify the single-pass path.

## §B `/moai security` retirement — the binding constraint

- **SPEC-SUBCOMMAND-RETIRE-001** (`.moai/specs/SPEC-SUBCOMMAND-RETIRE-001/`, Tier L, status `completed`) removed `design` / `brain` / `e2e` / `coverage` / **`security`** from the template source ("most-aggressive option, permanent for all distributed users").
  - spec.md:122-126: "security-audit request after `/moai security` is removed, the orchestrator shall route the request to `Agent(general-purpose)` with security scope, loading the … `moai-ref-owasp-checklist` … `moai-ref-llm-security` reference skills — preserving the OWASP audit capability `/moai security` provided."
  - Removed aliases include `security`/`audit`/`sec` (spec.md:109).
- **Corroborating live evidence**: `moai-ref-owasp-checklist` SKILL.md § Target Agents states: "`/moai review --security` - Primary security-audit invocation surface (**replaces the retired `/moai security` subcommand** per SPEC-SUBCOMMAND-RETIRE-001)".
- **Design consequence**: a new `/moai security` surface is OFF the table. The deep scan MUST live under `/moai review --deep` (design.md §S). This is the single most load-bearing research finding for the surface decision.

## §C Dynamic-workflow primitives (reused engine)

- `.claude/rules/moai/workflow/dynamic-workflows.md`: the `Workflow()` primitive is a JS script the runtime executes to orchestrate subagents (v2.1.154+). Up to 16 concurrent / 1000 total per run. **No mid-run user input** (only permission prompts pause). Script body MUST be deterministic (no `Date.now()`/`Math.random()` call — timestamps injected via args or stamped after return). "Claude writes the script for the task."
- **Template-managed?** NO. "MoAI does not ship any saved workflows by default; the user-owned `.claude/workflows/` directory is not template-managed." Confirmed live: `ls internal/template/templates/.claude/workflows/` → does not exist. Local `.claude/workflows/` holds `codemaps-extract.js`, `plan-research-fanout.js`, `sync-audit-4dim.js`, `hns-oss-docs-run.js`, `hns-release-update-run.js` — all local/user-owned, none template-managed.
- **Design consequence**: the deep scan CANNOT ship a `.js` script in the template tree. The shipped artifact is the playbook; the Workflow is runtime-constructed (design.md §A). This is what keeps the SPEC template-neutral.
- **Adversarial-verify precedent**: `.claude/workflows/sync-audit-4dim.js` (11065 bytes) — 4 parallel read-only judges + in-script harmonic-mean verdict. Structural precedent for the per-finding parallel-judge-then-reduce panel (design.md §C). `plan-research-fanout.js` (3-4 read-only lens explorers + a contradiction-marking synthesizer) is the second precedent.
- **Mode selection**: `.claude/rules/moai/workflow/orchestration-mode-selection.md` — Mode 6 (workflow) is downstream of Implementation Kickoff Approval; Mode 4 (parallel, 3-5 concurrent) is the fallback. The degradation ladder (design.md §F) uses these two modes.

## §D Security reference skills (reused domain knowledge)

Confirmed present in BOTH local and template trees (`.claude/skills/` + `internal/template/templates/.claude/skills/`):
- `moai-ref-owasp-checklist` — OWASP Top 10, auth, input-validation, HTTP headers, secrets patterns, trust-boundary verification principles.
- `moai-ref-llm-security` — prompt-injection defense, OWASP LLM Top 10, MCP/agentic hardening, NIST AI RMF, MITRE ATLAS.
- `moai-ref-secops` — CI/CD hardening, container/K8s, IaC misconfig, SAST/DAST, OWASP API Top 10 operational defense.
- `moai-ref-supply-chain` — SBOM, dependency-confusion, malicious-package triage, SLSA, Sigstore.
- **Design consequence**: hunt-phase agents load these per-area via on-demand `Skill()` injection (REQ-SDS-011), NOT static preload (adding 4 heavy skills to every review preload is a token regression). This follows `skill-routing.md` §1.

## §E Distribution precedent — SPEC-E2E-REVIVAL-001 (Tier L, markdown-shipping)

The closest structural precedent (Tier L, workflow-skill + command edits, both trees, no/minimal Go). Its design.md §G "Distribution Design" and §H "Test/CI Mapping" are the template for this SPEC's M6 mechanical steps. Key difference: E2E-REVIVAL ADDED an agent (`e2e-tester`, catalog count 9→10); THIS SPEC adds NO agent (reuses sync-auditor + per-spawn `Agent(general-purpose)` + `Agent(isolation:"worktree")`), so `catalog.yaml` counts stay unchanged — a strictly lighter footprint.

## §F Hook system (relevant to SPEC-2, out of scope here)

- Hooks live in `.claude/hooks/moai/` (shell wrappers → `moai hook <event>`). Valid events: PreToolUse, PostToolUse, Stop, SessionStart, TeammateIdle, TaskCompleted (NOT "PreCommit"). Timeout default 5s (MoAI policy).
- The always-on in-session guardian (SPEC-2) is the natural home for a PostToolUse-style lightweight security lens. This SPEC-1 (on-demand deep scan) does NOT touch hooks — it is an explicitly-invoked orchestrator playbook. Recorded here to scope the SPEC-1/SPEC-2 boundary; hooks are out of scope for SPEC-1 (spec.md §C).

## §G Results-classification rule (SPEC vs report)

- `.claude/skills/moai-workflow-spec/SKILL.md` § "What Does NOT Belong in .moai/specs/": a Security Audit "Analyzes existing code" → `.moai/reports/security-audit-{DATE}/`. The deep scan's output is analysis of existing code ⇒ a REPORT.
- **Design consequence**: results go to `.moai/reports/security-deepscan-<ts>/`, NOT `.moai/specs/` (REQ-SDS-040). `.moai/reports/` is already gitignored (CLAUDE.local.md §2); the self-`.gitignore` (REQ-SDS-044) is belt-and-suspenders per the absorbed source.

## §H Resolved decisions (user-confirmed via AskUserQuestion)

1. Single-commit scope token: RESOLVED as a new `--commit <SHA>` flag (consistent with REQ-SDS-003; not an overload of `--branch`). Recorded in plan.md § Resolved Decisions.
2. Results-dir retention: RESOLVED as no auto-prune for the first cut (retention left to the user, revisit in SPEC-2). Recorded in spec.md §C Out of Scope + plan.md § Resolved Decisions.

Both were settled before Implementation Kickoff Approval; no open clarification remains.
