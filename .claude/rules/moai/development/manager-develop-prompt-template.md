---
description: "manager-develop delegation Prompt Template — Tier M/L SPEC run-phase 5-section standard. Load only when authoring a SPEC delegation."
paths: ".moai/specs/**,.claude/agents/moai/manager-develop.md,.claude/skills/moai/workflows/run.md"
---

# manager-develop Delegation Prompt Template

## Applicability

[ZONE:Evolvable] [HARD] The Section A-E 5-section delegation template defined in this rule is **REQUIRED for Tier M and Tier L SPEC delegations** and **OPTIONAL for Tier S** delegations. Tier S SPECs (≤300 LOC, <5 files affected, 2 artifacts per the LEAN workflow) MAY use minimal delegation prompts (~500-800 tokens) covering only:

## cycle_type Mode Reference

Per the canonical agent catalog policy, the `manager-develop` agent operates in one of three `cycle_type` modes selected per the run-phase task profile:

| cycle_type | Loop pattern | When to use | Iteration contract | Canonical reference |
|------------|--------------|-------------|---------------------|---------------------|
| `ddd` | ANALYZE-PRESERVE-IMPROVE | Existing codebases with minimal test coverage (< 10% per quality.yaml `development_mode: ddd` selection); characterization-test-first preservation of behavior | No fixed iteration limit; one cycle per logical refactoring chunk | `.claude/rules/moai/workflow/spec-workflow.md` § Run Phase DDD Mode |
| `tdd` | RED-GREEN-REFACTOR | Default — all new development work, brownfield projects with pre-RED analysis (≥ 10% coverage per quality.yaml `development_mode: tdd` selection); test-first development | No fixed iteration limit; one cycle per behavior specification | `.claude/rules/moai/workflow/spec-workflow.md` § Run Phase TDD Mode |
| `autofix` | **DIAGNOSE-PATCH-VERIFY** | CI auto-fix loop after `scripts/ci-watch/run.sh` detects a failing required check; semantic-failure-safe patching of lint / build / type errors | **Maximum 3 iterations** per PR push (per-PR-push counter, not per-session); escalation to `AskUserQuestion` after iteration 3 with no auto-resume timeout | `.claude/rules/moai/workflow/ci-autofix-protocol.md` |

### cycle_type=autofix DIAGNOSE-PATCH-VERIFY pattern

Each iteration of the autofix loop executes the three-step DIAGNOSE-PATCH-VERIFY pattern:

1. **DIAGNOSE**: Read the failing CI check output (provided by the orchestrator from `scripts/ci-watch/run.sh`). Identify the root cause — lint rule violation, build error, type error, missing dependency, etc.
2. **PATCH**: Apply a minimal fix that addresses the root cause without expanding scope. The autofix loop MUST NOT modify `.env`, `.env.*`, credentials files, secrets, or `scripts/ci-watch/run.sh` or any Wave 2 infrastructure scripts.
3. **VERIFY**: Re-run the failing check locally; if exit 0, push the patch as a new commit on the PR branch. If still failing, increment the iteration counter and repeat from DIAGNOSE.

### autofix escalation contract

After 3 iterations without success, the loop MUST halt and the orchestrator MUST trigger an `AskUserQuestion` blocking call presenting the user with at least: (a) continue with manual investigation, (b) revert the offending change and re-plan, (c) abort with structured failure report. Semantic failures (data race, deadlock, panic, test assertion failure) MUST NOT be auto-patched without human approval.

### Logged at

Every autofix iteration MUST be logged to `.moai/logs/ci-autofix/` with timestamp, patch summary, and CI result.



- Goal (single-paragraph task description)
- Deliverables (concrete file/commit list)
- Constraints (PRESERVE list, forbidden commands)
- Self-verification (AC PASS/FAIL matrix)

When applying the minimal form for Tier S, Section B (Known Issues B1-B12) MAY be filtered to relevant categories only or omitted entirely if no listed risk applies. Section C (Pre-flight) MAY be reduced to the single most-relevant baseline command.

When the SPEC tier is M or L, the full Section A-E template SHOULD be applied; Section B (known issues) MAY filter B1-B12 categories by domain relevance (e.g., a documentation-only SPEC may omit B1 cross-platform build tags).

Tier classification reference: `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Complexity Tier (S/M/L).

> [ZONE:Evolvable] [HARD] Every Tier M/L `manager-develop` subagent delegation prompt MUST include the 5 sections of this template (Context / Known Issues / Pre-flight / Constraints / Self-Verification Deliverables). Tier S MAY use the minimal form from the Applicability section above. Omission (at Tier M/L) increases the risk of repeated re-delegation.

This rule standardizes the delegation-quality improvements derived from a meta-analysis. Goal: pre-empt defects via 1-pass delegation.

## 1. Standard Delegation Prompt 5-Section Structure

### Section A — Context (location + branch + SPEC artifact paths)

Mandatory to state:
- Work location (project root absolute path)
- Current branch + HEAD SHA (clarify where the manager stacks additional commits)
- SPEC artifact paths (`.moai/specs/SPEC-XXX/{spec,plan,acceptance,progress}.md`) + line count
- plan-auditor verdict (PASS score + whether a re-run is recommended)
- Existing infrastructure (PRESERVE targets + EXTEND targets)

### Section B — Known Issues Auto-Injection (most important)

[ZONE:Evolvable] [HARD] The following 12 categories of known issues MUST be auto-included in the delegation prompt. Omission = re-delegation risk.

**B1. Cross-platform Build Tags**
- Force build tags when using the syscall package
- Recommended: separate files with `//go:build !windows` + `//go:build windows`
- Verification: `GOOS=windows GOARCH=amd64 go build ./...` MUST pass

**B2. Cross-SPEC Policy Conflict Pre-Scan**
- Check the affected package's retired/superseded SPECs (e.g., a prior harness retirement)
- Run `grep -r "Retired\|TestHarnessRetirement\|deprecation-marker" internal/<pkg>`
- On conflict: state the reversal in the SPEC body or define a new SPEC scope

**B3. C-HRA-008 / Subagent Boundary Discipline**
- No AskUserQuestion calls in subagent-domain code such as `internal/harness/`, `internal/hook/`
- Verification: `grep -rn 'AskUserQuestion\|mcp__askuser' <pkg> | grep -v "_test.go" | grep -v "// "` yields 0 matches
- CI guard test required: `<pkg>/subagent_boundary_test.go`

**B4. Frontmatter Canonical Schema**
- Use `created:`/`updated:`/`tags:` (snake_case aliases prohibited)
- Reference: `.claude/rules/moai/development/spec-frontmatter-schema.md`

**B5. CI 3-tier Awareness**
- spec-lint, golangci-lint, Test (per OS) can each fail separately
- pre-existing baseline vs NEW defect classification

**B6. spec-lint Heading Convention**
- `## Out of Scope` (h2) alone triggers a `MissingExclusions` ERROR
- A `### <X.Y> Out of Scope` (h3) sub-section is required

**B7. observer.go / capture path resolution**
- `input.CWD` empty → `os.Getwd()` fallback leaks the working dir (cause of the `internal/hook/.moai/` anomaly)
- Recommended: prefer `$CLAUDE_PROJECT_DIR`

**B8. Working Tree Hygiene**
- Do NOT modify runtime-managed files (`.moai/harness/usage-log.jsonl`, `.moai/state/`)
- Relies on session_end's `cleanupBogusRootDir` (the `{}/`  literal directory is a cleanup target)
- Do NOT include unrelated untracked files in commits (`git add` specific paths only)

**B9. Git Commit + Push Performed Directly (Hybrid Trunk 1-person OSS)**
- manager-develop is recommended to perform commit + push directly within this SPEC scope (direct-to-main — Hybrid Trunk 1-person OSS policy, Tier S/M)
- Conventional Commits format required (`feat(SPEC-...): M{N} <subject>`)
- Both per-M separate commits + final push, or per-M push, are allowed
- Never use `--no-verify` (a warn-only pre-commit hook is normal)
- Exceptions: (a) the orchestrator performs the push on a parallel-session race, (b) the orchestrator handles it when user confirmation is needed in an AC PASS-WITH-DEBT state, (c) on an explicit blocker report
- This rule does **NOT** apply to manager-docs — for manager-docs, commit + push in the /moai sync workflow is the deliverable itself

**B10. Untouched Paths PRESERVE (Scope Discipline)**
- Never modify the working tree beyond this SPEC's plan.md §A.5 PRESERVE list
- Take extra care while a parallel manager-develop instance is running (do not touch other directories' scope)
- Do not touch runtime-managed files (`.moai/harness/*`, `.moai/state/*`, `.moai/cache/*`)
- Do not touch unrelated SPEC directories (other SPECs' plan-phase artifacts)
- Do not touch parallel-session research/audit artifacts (`.moai/research/*`)

**B11. AskUserQuestion Prohibited (Subagent Boundary)**
- Subagents must not interact with the user directly (CLAUDE.md §8 + askuser-protocol.md §Orchestrator–Subagent Boundary)
- On finding a blocker, return a structured blocker report (the orchestrator runs AskUserQuestion + re-delegates)
- Blocker report format: 4 options + each option's change/impact/risk/ETA stated
- Never ask free-form prose questions (no "? how should we proceed?" pattern in the response body)

**B12. Sync-phase CHANGELOG emission discipline (manager-docs only)**
- Before drafting CHANGELOG entries, `Read` every implementation file referenced in the SPEC plan.md (do NOT rely on plan.md description alone — plan-phase placeholders may diverge from final implementation).
- Before appending to `CHANGELOG.md` `[Unreleased]` section, run `grep -c '<SPEC-ID>' CHANGELOG.md` — if the count is ≥1, halt emission and return blocker report (avoid duplicate entries from parallel BATCH-SYNC sessions).
- Verify file paths claimed in CHANGELOG match actual `ls <package-path>` output before committing.
- Verify AC count in CHANGELOG matches `acceptance.md` (SSOT) — NOT `progress.md` (which may include deferred AC).
- Origin: an earlier CHANGELOG cleanup root cause analysis (BATCH-SYNC line hallucination incident).

### Section C — Pre-flight Check List (mandatory verification before starting)

Run by the delegated manager-develop before any code change:

```bash
# 1. Check current branch + baseline
git branch --show-current
git rev-parse HEAD

# 2. Pre-check cross-platform build feasibility
go build ./...
GOOS=windows GOARCH=amd64 go build ./...

# 3. Measure the existing lint baseline (to distinguish NEW vs pre-existing)
golangci-lint run --timeout=2m 2>&1 | tail -5

# 4. Print the list of PRESERVE target files
ls <PRESERVE_GLOB>

# 5. Check retired/superseded SPECs of affected packages
grep -r "Retired\|TestHarnessRetirement\|superseded" internal/<target_pkg> || echo "no conflicts"
```

### Section D — Constraints (DO NOT VIOLATE)

Explicit list in each delegation prompt:
- Enumeration of PRESERVE target files (when applying the Brownfield strategy)
- List of unrelated untracked/modified files (do not modify)
- Forbidden commands (`--no-verify`, `--amend`, force-push to main, …)
- Required commands (Conventional Commits, `🗿 MoAI` trailer, …)
- Binary constraints such as C-HRA-008 (grep 0 matches)

### Section E — Self-Verification Deliverables

> Each E-item is reported per the verification-claim-integrity 5-section format (Claim / Evidence / Baseline-attribution / Gaps / Residual-risk) — see `.claude/rules/moai/core/verification-claim-integrity.md` §3.

When manager-develop reports completion, it MUST include self-verification of the following items:

**E1. AC Binary PASS/FAIL Matrix**
| AC | Status | Verification Command | Actual Output |
|----|--------|---------------------|---------------|
| AC-XXX-001 | PASS | `go test -run TestX ./pkg` | `PASS — ok  pkg 0.5s` |

**E2. Cross-Platform Build result**
```
$ go build ./...                          → exit 0
$ GOOS=windows GOARCH=amd64 go build ./... → exit 0
```

**E3. Coverage measurement (≥85% threshold per package)**
```
$ go test -cover ./internal/<pkg>/...
```

**E4. Subagent Boundary Grep (C-HRA-008 family)**
```
$ grep -rn 'AskUserQuestion' <pkg> | grep -v "_test.go" | grep -v "// "
(no output expected)
```

**E5. Lint Status (distinguish NEW vs baseline)**
```
$ golangci-lint run --timeout=2m
# On NEW issues, report explicitly; mark pre-existing baseline separately
```

**E6. Branch HEAD + Push state**
- List of new commit SHAs
- Result of `git push origin <branch>`

**E7. Blocker Report (if any)**
- When the delegation prompt did not specify a needed user decision, report it as a structured blocker (NEVER call AskUserQuestion)

## 2. Delegation Prompt Authoring Workflow (from the orchestrator's perspective)

```
1. Compose Section A (SPEC artifact paths + current git state)
2. Auto-inject Section B (select the 12 categories via keyword matching from lessons memory)
3. Section C standard pre-flight checks (copy verbatim from above)
4. Section D constraints (extract from SPEC + PRESERVE list + working tree state)
5. Section E deliverables (copy verbatim from above)
```

## 3. Anti-Patterns

Cases of non-compliance with this template — increased re-delegation risk:

- Missing Section B → cross-platform build / cross-SPEC conflicts discovered after the fact
- Missing Section E → the orchestrator verifies serially (~10 min extra loss)
- 1-liner delegation like "implement the SPEC" → manager-develop makes silent assumptions due to an insufficient prompt
- Missing PRESERVE enumeration → unintended file modifications

## 4. Related Layers (meta-analysis §3)

This rule standardizes Layer A (delegation prompt quality improvement). Subsequent layers:

- Layer B (parallel delegation — Agent Teams): a separate rule is needed
- Layer C (Background CI watch): standardize the `gh pr checks --watch` pattern
- Layer D (verification parallelization): orchestrator self-discipline
- Layer F (automatic lessons capture): SubagentStop hook extension

## 5. Verification (this rule applied to itself)

On the next SPEC delegation, measure by phase ordering (not wall-clock targets, per `agent-common-protocol.md` § Time Estimation):
- 1-pass success rate — Priority High target (baseline: 33%)
- Re-delegation count — Priority Medium target (baseline: 3)
- Overall completion — track by milestone progression, not duration

---

Version: 1.0.0
Status: Active — applies to all manager-develop delegations
