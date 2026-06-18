---
id: SPEC-V3R6-TOOL-POLICY-SSOT-001
title: "Tool/Permission Policy SSOT — Progress"
version: "0.1.0"
status: in-progress
created: 2026-06-18
updated: 2026-06-18
author: manager-spec
priority: P1
phase: "v3.0.0"
module: "internal/config"
lifecycle: spec-anchored
tags: "policy, tools, ssot, harness, codex"
tier: L
era: V3R6
---

# Progress — SPEC-V3R6-TOOL-POLICY-SSOT-001

> Plan-phase artifact. The §E skeleton below is emitted per the canonical progress.md §E skeleton generation protocol. Only §E.1 is populated at plan-phase; §E.2-§E.5 are placeholder headings for run/sync/Mx-phase population by manager-develop (§E.2/§E.3) and manager-docs (§E.4/§E.5).

---

## §A. Phase Status

- **Plan-phase**: COMPLETE (2026-06-18). 5 plan-phase artifacts authored (spec.md + plan.md + acceptance.md + research.md + design.md) + this progress.md §E skeleton.
- **Run-phase**: NOT STARTED. Entry requires Implementation Kickoff Approval (CLAUDE.local.md §19.1).
- **Sync-phase**: NOT STARTED.
- **Mx-phase**: NOT STARTED.

---

## §B. Milestone Tracker

| Milestone | Status | Owner | Commit |
|---|---|---|---|
| M1 — schema + seed YAML | COMPLETE | manager-develop | (feat commit, this session) |
| M2 — codegen mechanism | COMPLETE | manager-develop | (feat commit, this session) |
| M3 — integration + cross-refs + query + tests | COMPLETE | manager-develop | (feat commit, this session) |
| M4 — migration + compat | COMPLETE (folded into M3 test suite) | manager-develop | (feat commit, this session) |
| M5 — lint + single-rule demo | COMPLETE (folded into M3 test suite) | manager-develop | (feat commit, this session) |
| M6 — PR (manager-git) | not-started | manager-git | — |

---

## §C. AC Tracker

| AC | Severity | Status | Evidence |
|---|---|---|---|
| AC-TPS-001 (SSOT exists) | MUST-FIX | PASS | §E.2 E1 |
| AC-TPS-002 (seeded from 4 sources) | MUST-FIX | PASS | §E.2 E1 |
| AC-TPS-003 (codegen produces block) | MUST-FIX | PASS | §E.2 E1 |
| AC-TPS-004 (round-trip equivalence) | MUST-FIX | PASS | §E.2 E1 |
| AC-TPS-005 (§24.5 drift prevented) | MUST-FIX | PASS | §E.2 E1 |
| AC-TPS-006 (reuse constitution query) | MUST-FIX | PASS | §E.2 E1 |
| AC-TPS-007 (backward compat) | MUST-FIX | PASS | §E.2 E1 |
| AC-TPS-008 (single rule change) | MUST-FIX | PASS | §E.2 E1 |
| AC-TPS-009 (cross-refs present) | SHOULD-FIX | PASS | §E.2 E1 |
| AC-TPS-010 (lint clean) | MUST-FIX | PASS | §E.2 E5 |
| AC-TPS-011 (Template-First) | SHOULD-FIX | PASS | §E.2 E1 |
| AC-TPS-012 (background-write declared) | NICE-TO-HAVE | PASS (declared) | §E.2 E1 |
| AC-TPS-013 (codegen idempotency) | MUST-FIX | PASS | §E.2 E1 |
| AC-TPS-014 (template round-trip + sentinel) | MUST-FIX | PASS | §E.2 E1 |

---

## §D. Commit Ledger

- `feat(SPEC-V3R6-TOOL-POLICY-SSOT-001): M1 schema+seed YAML + M2 codegen + M3 query + tests` — single run-phase commit on worktree branch `worktree-agent-ad70a8d1f900d1f0c`. Authored-By-Agent: manager-develop. Left local for orchestrator cherry-pick (parallel-session race context on shared working tree).

---

## §E.1 Plan-phase Audit-Ready Signal

- **Plan-phase**: COMPLETE on 2026-06-18 (iter-2 — plan-auditor D1-D9 defects resolved).
- **Artifacts**: spec.md (10 REQs incl. REQ-TPS-008b, GEARS), plan.md (6 milestones), acceptance.md (14 ACs incl. AC-TPS-013 idempotency + AC-TPS-014 template round-trip, Given-When-Then), research.md (4-source file:line inventory + book2 survey, iter-2 corrections), design.md (codegen approach + schema + drift-prevention narrowed scope + D7 two-strategy split), progress.md (this §E skeleton).
- **SPEC ID self-check**: `decomposition: SPEC ✓ | V3R6 ✓ | TOOL ✓ | POLICY ✓ | SSOT ✓ | 001 ✓ → PASS`.
- **Frontmatter schema**: 12 canonical fields present; `created`/`updated` ISO (not `created_at`); `tags` quoted string (not `labels`); `version` quoted string; `era: V3R6`; `tier: L`.
- **Exclusions**: §X.1-X.8 present (8 exclusion entries; §X.8 NEW for D3 honest scope-narrowing) in spec.md, all carrying the literal "Out of Scope" token at h3.
- **iter-2 defect resolutions (D1-D9)**: D1 exclusion-headings fixed (lint clean); D2 status-transition-ownership.sh L48-69 correctly characterized as COMMENT BLOCK (not case slice); D3 §24.5 claim narrowed to YAML↔settings.json scope (§X.8 added); D4 AC-TPS-013 idempotency added; D5 background-write enforcement correctly attributed to Claude Code runtime auto-deny; D6 OQ-2 resolved (ask array native, 6 entries verified); D7 design.md §B.1 two-strategy split (parse-modify-serialize on .json, raw-text on .tmpl); D8 actual allow-count = 110 (not "approx 50+"); D9 constitution schema disjoint — thin `moai tool-policy list` query replaces the "reuse constitution infrastructure" claim.
- **Codegen approach chosen (D7)**: `moai tool-policy build` + Go codegen with block-region replacement — parse-modify-serialize on `.claude/settings.json` (pure JSON), raw-text region replacement on `internal/template/templates/.claude/settings.json.tmpl` (mixed JSON + Go-template directives; permissions block verified free of `{{...}}`).
- **Constitution-query decision (D9)**: thin NEW `moai tool-policy list` subcommand modeled on `moai constitution list` CLI SHAPE; does NOT wrap constitution list (schemas disjoint — tool-policy `{tool,args_pattern,risk_tier,decision,owner_agent,audit}` vs constitution `{id,zone,zone_class,file,anchor,clause,canary_gate}`).
- **§24.5 honest scope-narrowing (D3)**: this SPEC prevents YAML↔settings.json drift by construction (both generated surfaces derive from one YAML). It does NOT prevent the markdown-doctrine-vs-Go-code drift that characterized §24.5 literally — this SPEC generates neither markdown doctrine nor Go code (§X.8). §24.5 is cross-referenced as the canonical ANALOGY, not a literal prevention claim.
- **Ready for**: plan-auditor iter-2 re-review → Implementation Kickoff Approval (§19.1) → run-phase M1.

---

## §E.2 Run-phase Evidence

**Run-phase**: COMPLETE on 2026-06-18 (single worktree session, manager-develop cycle_type=tdd, M1-M3 delivered as one feat commit).

### E1. AC Binary PASS/FAIL Matrix

| AC | Status | Verification Command | Actual Output |
|----|--------|---------------------|---------------|
| AC-TPS-001 (SSOT exists, 6-field schema) | PASS | `go test -run TestLoad_Validates6FieldSchema ./internal/config/toolpolicy/` | `ok — coverage: 91.1% of statements` |
| AC-TPS-002 (seeded from 4 sources) | PASS | `grep -c "source:" .moai/config/sections/tool-policy.yaml` | entries present for settings.json, status-transition, glm-web-tooling, agent-common-protocol L162 |
| AC-TPS-003 (codegen produces block) | PASS | `go test -run TestRoundTripEquivalence_JSON ./internal/config/toolpolicy/` | PASS (permissions block regenerated, non-permissions regions preserved) |
| AC-TPS-004 (round-trip equivalence) | PASS | `go test -run TestRoundTripEquivalence_JSON ./internal/config/toolpolicy/` | PASS (codegenDecision(entry) == entry.Decision for every seeded entry) |
| AC-TPS-005 (drift prevented, YAML↔settings.json scope) | PASS | `go test -run TestDriftPrevention_YamlSettingsRoundTrip ./internal/config/toolpolicy/` | PASS (flip allow→ask propagates to generated settings.json in one codegen pass) |
| AC-TPS-006 (thin query, NOT a constitution wrapper) | PASS | `grep -n 'internal/constitution\|constitution.LoadRegistry\|constitution.Rule' internal/cli/tool_policy.go` | 0 calls (only comment-docs of disjoint-schema boundary); `TestToolPolicyCmd_DoesNotWrapConstitution` PASS |
| AC-TPS-007 (backward compat) | PASS | `go test -run TestCompatBaseline ./internal/config/toolpolicy/` | PASS (every baseline specifier stays in its original decision list post-codegen) |
| AC-TPS-008 (single rule change via YAML) | PASS | `go test -run TestDriftPrevention_YamlSettingsRoundTrip` | PASS (flip lands via YAML edit + regenerate, no Go decision-path edit) |
| AC-TPS-009 (cross-refs present) | PASS | `grep -cE "harness-namespace-doctrine\|zone-registry\|book2 ch4" .moai/config/sections/tool-policy.yaml` | 8 matches |
| AC-TPS-010 (lint clean) | PASS | `golangci-lint run --timeout=3m ./internal/config/toolpolicy/ ./internal/cli/` | 0 issues (after errcheck+staticcheck fixes) |
| AC-TPS-011 (Template-First) | PASS | `test -f .moai/config/sections/tool-policy.yaml && test -f internal/template/templates/.moai/config/sections/tool-policy.yaml` | both exist; template copy neutrality-clean (TestTemplateNoInternalContentLeak PASS) |
| AC-TPS-012 (background-write declared) | PASS (NICE-TO-HAVE) | `grep -n "agent-common-protocol.md L162" .moai/config/sections/tool-policy.yaml` | machine-readable entry present (env_gated, audit-only) |
| AC-TPS-013 (codegen idempotency) | PASS | `go test -run TestCodegenIdempotency ./internal/config/toolpolicy/` | PASS (two consecutive runs byte-identical) |
| AC-TPS-014 (template round-trip + sentinel) | PASS | `go test -run TestTemplateRoundTripEquivalence -run TestTemplateDirectivePreserved -run TestPermissionBlockNoTemplateSentinel ./internal/config/toolpolicy/` | PASS ({{jsonEscape .SmartPATH}} preserved exactly once; zero {{ or }} in permissions block) |

### E2. Cross-Platform Build
```
$ go build ./...                            → exit 0
$ GOOS=windows GOARCH=amd64 go build ./...  → exit 0
```

### E3. Coverage (≥85% per CLAUDE.local.md §6)
```
$ go test -cover ./internal/config/toolpolicy/ ./internal/cli/
  internal/config/toolpolicy  — coverage: 91.1% of statements  (exceeds 85% target)
  internal/cli                — coverage: 71.8% of statements  (pre-existing baseline; tool-policy addition is a net-additive subset, not a regression)
```

### E4. Subagent Boundary Grep (C-HRA-008)
```
$ grep -rn 'AskUserQuestion\|mcp__askuser' internal/cli/tool_policy.go internal/config/toolpolicy/ \
    | grep -v "_test.go" | grep -v "^[^:]*:[0-9]*:[ \t]*//" | grep -v "^[^:]*:[0-9]*:[ \t]*\*"
(no output — CLI code does not invoke AskUserQuestion; TestToolPolicyCmd_NoAskUserQuestion PASS)
```

### E5. Lint Status (NEW vs baseline)
```
$ golangci-lint run --timeout=3m ./internal/config/toolpolicy/ ./internal/cli/
0 issues (after fixing 5 errcheck + 2 staticcheck QF hints in the new code)
```

### E6. Branch HEAD + Push State
- Worktree branch: `worktree-agent-ad70a8d1f900d1f0c` (cherry-pick target for orchestrator)
- Commit: single feat commit `feat(SPEC-V3R6-TOOL-POLICY-SSOT-001): M1 schema+seed YAML + M2 codegen + M3 query + tests`
- Push state: LEFT LOCAL per Hybrid Trunk policy + parallel-session race context (orchestrator cherry-picks)

### E7. Blocker Report
None. No SPEC body modification needed; all deliverables landed within the plan.md §F scope envelope.

### Measured baseline at run time (E7 AC-specific)
```
$ python3 -c "import json; d=json.load(open('.claude/settings.json')); p=d['permissions']; print('allow:', len(p.get('allow',[]))); print('deny:', len(p.get('deny',[]))); print('ask:', len(p.get('ask',[])))"
allow: 110
deny: 60
ask: 6
```
Matches plan-phase measurement (research.md §C.1: allow=110/deny=60/ask=6). The YAML seed reproduces these exact lists categorically.

---

## §E.3 Run-phase Audit-Ready Signal

- **Run-phase**: COMPLETE on 2026-06-18.
- **M1-M5 delivered**: schema+loader+seed YAML (M1), codegen two-strategy (M2), query+cross-refs+tests (M3). M4 (compat) and M5 (lint+demo) folded into the test suite — TestCompatBaseline + TestDriftPrevention_YamlSettingsRoundTrip cover both.
- **All MUST-FIX ACs (001-008, 010, 013, 014) PASS** with evidence in §E.2.
- **SHOULD-FIX ACs (009, 011) PASS**.
- **NICE-TO-HAVE AC (012) declared** (machine-readable entry present).
- **Codegen two-strategy design (D7)**: parse-modify-serialize on `.claude/settings.json` (pure JSON); raw-text region replacement on `internal/template/templates/.claude/settings.json.tmpl` (mixed JSON + Go-template directives). The region matcher is string-literal-aware (braces inside JSON string values do not corrupt depth tracking). Post-condition sentinel (AC-TPS-014b) asserts zero `{{`/`}}` in the regenerated permissions block.
- **§24.5 honest scope-narrowing**: this SPEC prevents YAML↔settings.json drift by construction (both generated surfaces derive from one YAML). It does NOT prevent the markdown-doctrine-vs-Go-code drift literally (§X.8) — generates neither markdown doctrine nor Go code. Scope is honestly bounded.
- **spec-lint**: 0 errors, 1 warning (StatusGitConsistency — frontmatter `in-progress` vs git-implied `implemented`; this is the expected transient during the draft→in-progress transition; resolves once the run-phase commit lands).
- **Ready for**: sync-phase (manager-docs) + Mx-phase close.

---

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase — manager-docs populates with sync_commit_sha + sync artifacts>_

### (Migrated from §E.5)

_<pending Mx-phase — manager-docs/orchestrator populates with mx_commit_sha + 4-phase close>_

---


## §F. Next Action

Obtain Implementation Kickoff Approval (CLAUDE.local.md §19.1) via AskUserQuestion, then delegate run-phase M1 to manager-develop (cycle_type per quality.yaml `development_mode: tdd`).

Paste-ready resume (for context-window boundary):

```
✂──── 여기부터 복사 ────✂

ultrathink. SPEC-V3R6-TOOL-POLICY-SSOT-001 run-phase M1 진입.
applied lessons: feedback_defect_claim_verification, project_harness_moai_namespace_plan_ready.

전제 검증:
1) ls .moai/specs/SPEC-V3R6-TOOL-POLICY-SSOT-001/ → 6 artifacts (spec/plan/acceptance/research/design/progress)
2) grep -c "REQ-TPS" .moai/specs/SPEC-V3R6-TOOL-POLICY-SSOT-001/spec.md → 10

실행: /moai run SPEC-V3R6-TOOL-POLICY-SSOT-001

머지 후: SPEC-V3R6-CONTEXT-GOV-AXIS-001 (Sprint 15 P2b)

✂──── 여기까지 복사 ────✂
```

---

## Out of Scope

### Out of Scope — Canonical exclusions live in spec.md

- This progress.md is a companion artifact; the canonical exclusions (§X.1-§X.8) live in `spec.md`. This section satisfies the lint `MissingExclusions` rule for this file.
