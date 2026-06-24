# Plan — SPEC-V3R6-ORCH-IGGDA-001

> **Tier L** — 6 milestones. Each milestone is a manager-develop delegation unit. Milestones are ordered by dependency; parallel execution within a milestone is permitted when files are independent. Per `.claude/rules/moai/development/sprint-round-naming.md`, Milestone (M1–M6) is the within-SPEC ordered-step unit.

---

## §A — Context

This plan implements the 5 deliverables (D1–D5) of SPEC-V3R6-ORCH-IGGDA-001. The implementation is **Template-First** (CLAUDE.local.md §2): every rule-file change goes into `internal/template/templates/` first, then `make build`, then `moai update`. The implementation amends a FROZEN invariant (Implementation Kickoff Approval) — design.md §F carries the rigorous justification.

**Approach summary (one paragraph)**: M1 lands the IGGDA 4-phase pipeline definition + the D2 safe-condition predicate in `orchestration-mode-selection.md` (the FROZEN-amend, the highest-stakes change — done first so plan-auditor can challenge it early). M2 lands the bounded recursive self-diagnosis loop in `run.md` + the keyword list. M3 lands the moai-aware Stop hook driver shell script. M4 lands the D5 independent-audit preservation regression tests. M5 wires the `moai spec audit --filter-spec` flag if it does not exist (minor Go). M6 closes with drift/lint updates + the 3-phase close (sync-phase).

---

## §B — Known Issues (B1–B12 auto-injection, filtered to relevant)

> Per `.claude/rules/moai/development/manager-develop-prompt-template.md` § Section B. Filtered to categories relevant to this SPEC (rules + hook + minor Go).

- **B2 (Cross-SPEC policy conflict)**: `run.md:122,124` carries the FROZEN Implementation Kickoff Approval invariant. Any M1 edit MUST preserve the `[ZONE:Frozen]` marker and MUST surface the safe-condition predicate as an AMENDMENT, not a removal. `grep -r "Implementation Kickoff Approval\|GATE-2" internal/template/templates/.claude/` MUST return the amended content, not a deletion.
- **B6 (spec-lint heading convention)**: this SPEC's spec.md carries `### Out of Scope — <topic>` H3 sub-headings with `-` bullets (satisfies `OutOfScopeRule`). The run-phase manager-develop MUST NOT break this when touching spec.md frontmatter.
- **B8 (Working tree hygiene)**: `.moai/harness/*`, `.moai/state/*`, `.moai/cache/*` are runtime-managed — DO NOT touch. The Stop hook driver reads `progress.md` but does NOT write to runtime-managed paths.
- **B9 (Git commit + push, Hybrid Trunk)**: per `.moai/docs/git-workflow-doctrine.md`, this SPEC is Tier L → PR creation via manager-git (NOT main-direct). M1–M5 commits land on a feature branch; M6 merges via PR.
- **B11 (AskUserQuestion prohibition in subagents)**: the recursive self-diagnosis sub-agent (`Agent(general-purpose)`) MUST NOT call `AskUserQuestion`. On semantic failure or iteration-3 halt, it returns a blocker report; the orchestrator runs `AskUserQuestion`.

---

## §C — Pre-flight (manager-develop executes before M1 code change)

```bash
# 1. Branch + baseline
git branch --show-current    # expect: feature/SPEC-V3R6-ORCH-IGGDA-001 or similar
git rev-parse HEAD

# 2. Confirm Template-First source location
ls internal/template/templates/.claude/skills/moai/workflows/run.md
ls internal/template/templates/.claude/rules/moai/workflow/orchestration-mode-selection.md

# 3. Existing lint baseline
golangci-lint run --timeout=2m 2>&1 | tail -5

# 4. spec-lint baseline
moai spec lint .moai/specs/SPEC-V3R6-ORCH-IGGDA-001/ 2>&1 | tail -10

# 5. FROZEN-invariant current state (must be preserved-as-amended)
grep -n "Implementation Kickoff Approval\|\[ZONE:Frozen\]" internal/template/templates/.claude/skills/moai/workflows/run.md | head -10

# 6. moai spec audit flag existence check (Q4)
grep -n "filter-spec\|FilterSpec" internal/spec/audit.go internal/cli/spec*.go 2>/dev/null || echo "FLAG ABSENT — M5 adds it"
```

---

## §D — Constraints (DO NOT VIOLATE)

- **PRESERVE**: the `[ZONE:Frozen]` marker on Implementation Kickoff Approval (`run.md:122`, `orchestration-mode-selection.md:14`, CLAUDE.local.md §19.1). Amend the BEHAVIOR, not the marker.
- **PRESERVE**: the 6 HARD safety conditions C1–C6 from SPEC-AUTONOMY-RUN-GOAL-001. C1 is amended (Path B), not removed; C2–C6 are inherited verbatim.
- **PRESERVE**: the 5 circuit-breaker invariants from runtime-recovery-doctrine.md §3. The recursive loop MUST comply.
- **PRESERVE**: independent auditor separation (plan-auditor, sync-auditor in fresh contexts).
- **FORBIDDEN**: `--no-verify`, `--amend`, force-push to main, modification of `.env` / credentials / `scripts/ci-watch/run.sh` (CONST-V3R5-011/013).
- **FORBIDDEN**: subagent `AskUserQuestion` calls (B11).
- **FORBIDDEN**: template edits that leak internal SPEC IDs / commit SHAs / macOS-bias paths (neutrality CI guard).
- **REQUIRED**: Conventional Commits + `🗿 MoAI <email@mo.ai.kr>` trailer on every commit.
- **REQUIRED**: Template-First — `internal/template/templates/` edit → `make build` → `moai update` → local verification.

---

## §E — Self-Verification (manager-develop §E1–E7 deliverables per M)

Each milestone's completion report MUST include the §E1–E7 self-verification matrix from `manager-develop-prompt-template.md`:
- E1: AC binary PASS/FAIL matrix (per the milestone's scoped ACs)
- E2: cross-platform build (`go build ./...` + `GOOS=windows GOARCH=amd64 go build ./...`)
- E3: coverage ≥ 85% on touched Go packages (M5 only — M1–M4 are rules/hook, coverage N/A)
- E4: subagent-boundary grep (`grep -rn 'AskUserQuestion' <pkg> | grep -v _test.go | grep -v "^.*//"` = 0 matches)
- E5: lint baseline (NEW vs pre-existing)
- E6: branch HEAD + push state
- E7: blocker report (if any)

---

## §F — Milestones (M1–M6)

### M1 — IGGDA pipeline definition + D2 safe-condition predicate (FROZEN-amend, highest-stakes, first)

**Owner**: manager-develop (cycle_type=ddd — brownfield rule amendment)
**Files** (Template-First):
- `internal/template/templates/.claude/rules/moai/workflow/orchestration-mode-selection.md` (D1 pipeline + D2 predicate, AMEND § Implementation Kickoff Approval + add § IGGDA Pipeline)
- `internal/template/templates/.claude/skills/moai/workflows/run.md:120-126` (D2 amendment — safe-condition predicate cross-reference)
- (Decision deferred to M1 kickoff) `internal/template/templates/.claude/rules/moai/workflow/iggda-pipeline.md` NEW sibling if the pipeline definition is too large for orchestration-mode-selection.md

**Deliverables**:
1. The safe-condition predicate (REQ-IGGDA-004) is documented in orchestration-mode-selection.md with the 4 conditions (a)–(d) and the auto-proceed vs explicit-gate decision logic.
2. The dangerous-domain keyword list (REQ-IGGDA-005, design.md §F.3) is enumerated.
3. The IGGDA 4-phase pipeline (REQ-IGGDA-008) is defined.
4. The `[ZONE:Frozen]` marker is PRESERVED on Implementation Kickoff Approval; the amendment is documented as a sub-section, not a marker removal.
5. `make build` regenerates embedded.go; `moai update` syncs to local.

**ACs covered**: AC-IGGDA-001, 002, 003, 004, 005a, 005b, 005c, 005d, 005e, 006, 007, 008 (12 ACs — the safe-condition predicate branches incl. all 5 explicit-gate sub-branches 005a–005e; see acceptance.md §D.1).

**Risk**: HIGH — FROZEN-invariant amendment. plan-auditor MUST scrutinize M1 before M2 begins.

---

### M2 — Bounded recursive self-diagnosis loop (D3)

**Owner**: manager-develop (cycle_type=ddd)
**Files** (Template-First):
- `internal/template/templates/.claude/skills/moai/workflows/run.md` (NEW section `## Recursive Self-Diagnosis Loop (bounded)`)
- `internal/template/templates/.claude/rules/moai/development/manager-develop-prompt-template.md` (reference injection — the loop's DIAGNOSE-PATCH-VERIFY inherits from autofix)

**Deliverables**:
1. The bounded loop (REQ-IGGDA-014 through REQ-IGGDA-019) is documented in run.md.
2. Mechanical vs semantic failure classification (REQ-IGGDA-015 + REQ-IGGDA-016) is enumerated with examples.
3. The max-3-iteration bound + escalation contract is documented.
4. The 5 circuit-breaker invariants compliance (REQ-IGGDA-017) is cross-referenced to runtime-recovery-doctrine.md §3.
5. The logging contract (REQ-IGGDA-018) — `progress.md §E Recursive Self-Diagnosis Log` — is defined.

**ACs covered**: AC-IGGDA-009 through AC-IGGDA-014, AC-IGGDA-021 through AC-IGGDA-026 (12 ACs — the bounded-loop + circuit-breaker + logging + forbidden-paths ACs whose content lives in run.md § Recursive Self-Diagnosis Loop; see acceptance.md §D.2/§D.3).

**Risk**: MEDIUM — the loop is new but inherits heavily from ci-autofix-protocol + autofix cycle_type.

---

### M3 — Moai-aware Stop hook driver (D4)

**Owner**: manager-develop (cycle_type=tdd — NEW shell script, test-first)
**Files** (Template-First):
- `internal/template/templates/.claude/hooks/moai/iggda-phase-driver.sh` (NEW)
- `internal/template/templates/.claude/hooks/moai/handle-iggda-phase-driver.sh` (NEW wrapper, per CLAUDE.local.md §7 shell-script-hook pattern)
- `internal/template/templates/.claude/settings.json.tmpl` (register the Stop hook)
- Tests: `internal/template/templates/.claude/hooks/moai/iggda-phase-driver_test.sh` (shell test) OR a Go test that exercises the hook via subprocess

**Deliverables**:
1. The Stop hook (REQ-IGGDA-009 through REQ-IGGDA-013) reads `progress.md` + invokes `moai spec audit --json --filter-spec=<SPEC-ID>` (M5 dependency — if M5 not yet done, the hook emits a graceful-degradation log and falls back to `progress.md`-only evaluation).
2. The Recovery-Signal Carve-Out compliance (REQ-IGGDA-011) — on recovery-turn `stopReason` or withheld-recoverable error, exit 0.
3. The hook returns exit codes + JSON with `ledger_note` field (REQ-IGGDA-013); NEVER calls `AskUserQuestion` directly (B11).
4. The hook is registered in settings.json Stop hook config with `timeout: 5` (per CLAUDE.local.md §7).

**ACs covered**: AC-IGGDA-015, 016, 017, 018, 019, 020 (6 ACs — the Stop hook driver ACs: `/goal` graceful degradation, exit-code/JSON return, subagent-boundary grep, timeout ≤5s, Windows stub [SHOULD-PASS], no runtime-path writes; see acceptance.md §D.2).

**Risk**: MEDIUM — shell hook + settings.json registration; cross-platform (Windows stub needed per CLAUDE.local.md §7).

---

### M4 — Independent-audit preservation regression (D5)

**Owner**: manager-develop (cycle_type=tdd)
**Files**:
- `internal/template/templates/.claude/rules/moai/workflow/orchestration-mode-selection.md` (add § IGGDA Independent-Audit Preservation — REQ-IGGDA-020 through REQ-IGGDA-023)
- Tests: a regression guard script (shell or Go) that verifies plan-auditor + sync-auditor spawn in fresh contexts (NOT continuation of implementer turn)

**Deliverables**:
1. The independent-audit preservation invariants (REQ-IGGDA-020 through REQ-IGGDA-023) are documented.
2. The "self-audit" vs "independent audit" disambiguation (REQ-IGGDA-022) is documented.
3. The regression guard proves: (a) plan-auditor is a separate `Agent()` spawn; (b) sync-auditor is a separate `Agent()` spawn; (c) a FAIL/INCONCLUSIVE verdict halts auto-advance.

**ACs covered**: AC-IGGDA-027, 028, 029, 030 (4 ACs — the independent-audit-preservation ACs: fresh-context spawn for plan-auditor/sync-auditor, FAIL/INCONCLUSIVE halt, self-audit vs independent-audit disambiguation; see acceptance.md §D.4).

**Risk**: LOW — documentation + regression guard; no FROZEN invariant touched.

---

### M5 — `moai spec audit --filter-spec` flag (if absent; minor Go)

**Owner**: manager-develop (cycle_type=tdd — Go flag addition, test-first)
**Files** (if flag is absent — Q4 pre-flight resolves this):
- `internal/spec/audit.go` (add `--filter-spec` flag + JSON output refinement)
- `internal/cli/spec_audit.go` (wire the flag)
- `internal/spec/audit_test.go` (test cases for the flag)

**Deliverables**:
1. If the flag exists: M5 is a no-op (close with a one-line commit documenting the flag's pre-existence).
2. If the flag is absent: `moai spec audit --json --filter-spec=SPEC-V3R6-ORCH-IGGDA-001` returns JSON with `drift_findings` filtered to the named SPEC. Coverage ≥ 85% on `internal/spec/audit.go`.

**ACs covered**: AC-IGGDA-037 (Go test suite passes for the `--filter-spec` flag addition; MUST-PASS when M5 touches Go, N/A when M5 is a no-op per Q4 pre-flight). NOT AC-025 — AC-025 (iteration log) belongs to M2.

**Risk**: LOW — additive Go flag; no existing behavior changed.

---

### M6 — Drift/lint updates + 3-phase close (sync-phase)

**Owner**: manager-docs (sync-phase per the canonical agent responsibility matrix)
**Files**:
- `.moai/specs/SPEC-V3R6-ORCH-IGGDA-001/spec.md` (frontmatter `status: draft → in-progress → implemented → completed` per the 3-phase close)
- `.moai/specs/SPEC-V3R6-ORCH-IGGDA-001/progress.md` (§E.2 run-phase evidence + §E.3 run-phase audit-ready signal + §E.4 sync-phase audit-ready signal + `sync_commit_sha`)
- `CHANGELOG.md` (entry)
- (If applicable) `README.md` / docs-site updates

**Deliverables**:
1. The 3-phase close (plan→run→sync) per SPEC-V3R6-LIFECYCLE-REDESIGN-001 — single sync commit carries `implemented → completed`.
2. `moai spec audit --json --filter-spec=SPEC-V3R6-ORCH-IGGDA-001` reports `drift_findings: []` (REQ-IGGDA-028).
3. spec-lint clean (0 findings).
4. The PR (Tier L) is created via manager-git.

**ACs covered**: AC-IGGDA-031, 032, 033, 034, 035, 036, 038 (7 ACs — the IGGDA-completeness terminal gate AC-031 + cross-cutting inherited C-series AC-032/033/034/035 + spec-lint AC-036 + template-neutrality AC-038; see acceptance.md §D.5/§D.6/§D.7).

**Risk**: LOW — sync-phase close is standard.

---

## §G — Anti-Patterns (M1–M6 specific)

- **AP-IGGDA-M1-001**: removing the `[ZONE:Frozen]` marker from Implementation Kickoff Approval in M1. The marker is PRESERVED; the behavior is amended via a sub-section.
- **AP-IGGDA-M2-001**: allowing the recursive loop to auto-fix semantic failures (race/deadlock/panic). REQ-IGGDA-016 is HARD — semantic failures escalate, never self-fix.
- **AP-IGGDA-M2-002**: exceeding the max-3-iteration bound without escalating. REQ-IGGDA-014 is HARD — iteration 4 is prohibited; escalation is mandatory.
- **AP-IGGDA-M3-001**: the Stop hook driver calling `AskUserQuestion` directly. REQ-IGGDA-013 + B11 forbid this; hooks return exit codes + JSON.
- **AP-IGGDA-M3-002**: the Stop hook driver blocking a recovery turn (exit 2 on a recovery signal). REQ-IGGDA-011 mandates exit 0 on recovery turns (Recovery-Signal Carve-Out).
- **AP-IGGDA-M4-001**: collapsing the independent auditor into the implementer's context. REQ-IGGDA-020 + REQ-IGGDA-021 forbid this.
- **AP-IGGDA-M5-001**: changing existing `moai spec audit` behavior (the flag is ADDITIVE; existing callers are unaffected).
- **AP-IGGDA-M6-001**: using a combined/abbreviated scope in the close commit subject (e.g., `chore(SPEC-V3R6-ORCH): 3-phase close`). Per the close-subject full-ID mandate, the subject MUST be `chore(SPEC-V3R6-ORCH-IGGDA-001): ... 3-phase close`.

---

## §H — Cross-References

- `.claude/rules/moai/development/manager-develop-prompt-template.md` § Section A-E delegation template (Tier L mandatory)
- `.claude/rules/moai/development/sprint-round-naming.md` — Milestone (M1–M6) is the within-SPEC ordered-step unit
- `.moai/docs/git-workflow-doctrine.md` — Tier L PR routing (manager-git)
- `.moai/specs/SPEC-V3R6-LIFECYCLE-REDESIGN-001/` — 3-phase close (plan→run→sync)
- `.moai/specs/SPEC-V3R6-DRIFT-LEGACY-CONVENTION-001/` — close-subject full-ID mandate
- `.moai/specs/SPEC-AUTONOMY-RUN-GOAL-001/plan.md` — sibling plan (run-phase autonomy milestones; IGGDA extends)

---

Version: 0.1.0 (plan-phase)
Status: Active — awaits plan-auditor independent audit before run-phase
