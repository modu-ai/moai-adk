> **[REMOVED 2026-07-13]** The `gateguard-fact-force.sh` PreToolUse advisory hook referenced in M7 / REQ-HOC-017 / AC-026 below has been FULLY DECOMMISSIONED from both the local working tree and the distribution template. The hook script, `MOAI_FACT_FORCE` env var, `.moai/state/fact-force/` state dir, and `.moai/logs/fact-force-skip.log` skip log are all removed. References below are preserved as historical decision-record content only.

# plan.md — SPEC-HOOK-OFFICIAL-COMPLIANCE-001

> **Plan-phase only.** This document decomposes the 8 audit recommendations into 8 milestones (M1-M8) with per-Rec dedup verdicts. No code or hook/rule file is modified at plan-phase.

---

## §A. Context

### §A.1 Primary source

`.moai/reports/hooks-improvement-plan-20260710.html` — the SSOT for the audit. 32 gaps across 8 groups, Top 8 prioritized recommendations, 6 cross-cutting themes, official-feature adoption matrix, per-group detailed findings. Each gap carries `{id, severity, category, finding, officialBasis, recommendation}`.

### §A.2 Authoritative baseline

`code.claude.com/docs/en/hooks` (2026-07-10 fetch via z.ai webReader MCP, GLM-backend routing). The Report's per-gap `officialBasis` fields embed the canonical requirements from this doc.

### §A.3 Audit scope (39 files)

- 35 template hook files under `internal/template/templates/.claude/hooks/moai/` (32 `handle-*.sh.tmpl` + `gateguard-fact-force.sh` + `status-transition-ownership.sh` + `sync-phase-quality-gate.sh` + `team-ac-verify.sh`)
- 3 rule docs (`agent-hooks.md`, `hooks-system.md`, `hook-independence.md`)
- `settings.json.tmpl` wiring

### §A.4 Pre-flight baseline (2026-07-10)

Verified tree-state facts (observed by manager-spec, not carried over from the Report):

- `internal/template/templates/.claude/hooks/moai/team-ac-verify.sh:45` emits `{"decision":"block","reason":...,"ledger_note":...}` — confirmed GAP-GD-01
- `internal/template/templates/.claude/hooks/moai/sync-phase-quality-gate.sh:357` emits `{"decision":"block","reason":"...","systemMessage":"..."}` at top level — confirmed GAP-GD-02
- `internal/template/templates/.claude/hooks/moai/handle-pre-tool.sh.tmpl:55-56,61-62,67-68` all end `exit 0` overriding moai exit code; `2>>"$MOAI_HOOK_STDERR_LOG"` on all 3 — confirmed GAP-TOOL-01
- `internal/template/templates/.claude/hooks/moai/handle-session-start.sh.tmpl:44` emits `{"hookSpecificOutput":{"additionalContext":"..."}}` — NO `hookEventName` — confirmed GAP-LC2-01 (NOT fixed by SPEC-HOOK-SESSIONSTART-PROBE-001)
- PostToolUse observe has `async:true` (settings.json.tmpl:92,116); Stop(144)/SubagentStop(168)/UserPromptSubmit(222) observe taps do NOT — confirmed GAP-OBS-01/02
- TeammateIdle(241)/TaskCompleted(256) timeout=5; PreCompact(30) timeout=5; sync-phase-quality-gate(139) timeout=60 — confirmed GAP-STT-01/LC-03

---

## §B. Known Issues (B1-B12 filter — domain-relevant subset)

- **B1 Cross-platform**: no syscall usage expected; wrapper edits are shell-only; Go inspection is read-only. `GOOS=windows GOARCH=amd64 go build ./...` stays green (no Go change).
- **B6 spec-lint heading**: `### Out of Scope — <topic>` H3 sub-headings used (§G of spec.md) — satisfies `OutOfScopeRule`.
- **B7 observer.go / capture path resolution**: NOT in scope (no `internal/hook/subagent_stop.go` edit).
- **B8 Working Tree Hygiene**: run-phase MUST NOT touch `.moai/state/`, `.moai/harness/`, `.moai/logs/`, or unrelated untracked SPEC dirs. `git add` specific paths only.
- **B10 Untouched Paths PRESERVE**: run-phase edits only the files named in the milestone scope; parallel-session work (other hook SPECs) MUST NOT be touched.
- **B11 AskUserQuestion prohibition**: subagent boundary — run-phase returns blocker reports, never calls AskUserQuestion.

---

## §C. Pre-flight (run-phase entry)

```bash
# 1. Branch + baseline
git branch --show-current
git rev-parse HEAD

# 2. Cross-platform build (no Go change expected, but verify)
go build ./...
GOOS=windows GOARCH=amd64 go build ./...

# 3. Lint baseline (distinguish NEW vs pre-existing)
golangci-lint run --timeout=2m 2>&1 | tail -5

# 4. Template-mirror parity baseline
go test ./internal/template/... -run TestRuleTemplateMirror 2>&1 | tail -10

# 5. SPEC-lint clean baseline
moai spec lint SPEC-HOOK-OFFICIAL-COMPLIANCE-001 2>&1 | tail -5
```

---

## §D. Dedup Verdicts against 6 Overlapping SPECs

For each of the 8 recommendations, the dedup verdict against the overlapping SPECs named in the task. Each verdict is one of: **NEW** (owned by this SPEC), **DONE** (cite the completed SPEC, drop from scope), **PARTIAL** (a completed SPEC did one piece; this SPEC carries the residual).

### Rec #1 — Blocking gate JSON contract + PreToolUse exit-code (HIGH)

| Overlap SPEC | Status | Verdict |
|---|---|---|
| SPEC-V3R6-HOOK-CONTRACT-FIX-001 | `implemented` | **NO overlap.** That SPEC covers the WorktreeCreate/Remove plain-text-stdout active-creator contract — NOT team-ac-verify (TaskCompleted) or sync-phase-quality-gate (Stop) or handle-pre-tool (PreToolUse). The "contract" in that SPEC's name is a different contract. |

**Dedup outcome**: Rec #1 is **NEW** — owned entirely by this SPEC (M1).

### Rec #2 — hooks-system.md + hook-independence.md doctrine refresh (MEDIUM)

| Overlap SPEC | Status | Verdict |
|---|---|---|
| SPEC-DIVECC-HOOK-FAILURE-MODE-AUDIT-001 | `completed` | **PARTIAL (file-exists, content-stale).** That SPEC created `hook-independence.md` as the shared-failure-mode catalogue + authoring checklist. It did NOT refresh the 8 specific stale points this Rec targets (PreCompact Can Block, SubagentStop stdout, Notification matcher, FileChanged literal note, UserPromptSubmit/PermissionRequest fields, handler-type matrix, exit-2 universal, timeout 10s→60s). The doctrine file EXISTS (DIVECC created it); this SPEC UPDATES it with the 8 corrections. |

**Dedup outcome**: Rec #2 is **NEW** (a doctrine UPDATE, not a creation). Cite DIVECC-HOOK-FAILURE-MODE-AUDIT-001 as the creator of `hook-independence.md`. Owned by this SPEC (M2).

### Rec #3 — async:true for 3 observation taps (MEDIUM)

| Overlap SPEC | Status | Verdict |
|---|---|---|
| SPEC-V3R6-HOOK-ASYNC-EXPAND-001 | `implemented` | **NO overlap.** That SPEC covers FileChanged + ConfigChange + TaskCreated + Notification async. Its §6.3 explicitly EXCLUDES UserPromptSubmit/Stop/SubagentStop from its scope ("sync, blocking — exit 2 for quality gate; intentionally synchronous"). The 3 taps this Rec targets are a different event set. |

**Dedup outcome**: Rec #3 is **NEW** — owned entirely by this SPEC (M3).

### Rec #4 — Timeout headroom for TeammateIdle/TaskCompleted/PreCompact (MEDIUM)

No prior SPEC covers this. **NEW** — owned by this SPEC (M4).

### Rec #5 — FileChanged matcher + ConfigChange block channel (MEDIUM)

No prior SPEC covers this. **NEW** — owned by this SPEC (M5). Requires Go inspection per §E.2 of spec.md.

### Rec #6 — Per-event fail-open semantics (MEDIUM)

| Overlap SPEC | Status | Verdict |
|---|---|---|
| SPEC-HOOK-SESSIONSTART-PROBE-001 | `completed` | **PARTIAL.** That SPEC implemented the standing SessionStart moai-resolvability probe — the exact mechanism Rec #6 references ("implement the standing SessionStart moai-resolvability probe"). The probe is DONE. The residual is the per-event fail-open SEMANTICS correction: WorktreeCreate (fail-closed: empty-stdout + exit-0 = creation-ABORT) and PermissionRequest (security-negative: exit-0 = ALLOW silently neutralizes the permission gate). |

**Dedup outcome**: Rec #6 is **PARTIAL**. The probe part is DONE — cite SPEC-HOOK-SESSIONSTART-PROBE-001 (completed) and DROP from active scope. The WorktreeCreate fail-closed guard + PermissionRequest warning priority = **NEW residual** owned by this SPEC (M6). The probe is referenced as the existing mechanism the residual extends.

### Rec #7 — Input hardening: MOAI_HOOK_STDERR_LOG + gateguard JSON escaping (LOW)

| Overlap SPEC | Status | Verdict |
|---|---|---|
| SPEC-HOOK-FACTFORCE-ADVISORY-001 | `completed` | **PARTIAL.** That SPEC rewrote `gateguard-fact-force.sh` from exit-2-blocking to exit-0-advisory (the UX-defect fix). It did NOT address GAP-GD-08 (raw `$session_id`/`$file_path`/`$tool_name` interpolation into `printf '{...}'` state-file write — JSON escaping missing) nor GAP-STT-04/GAP-TOOL-02 (the shared `MOAI_HOOK_STDERR_LOG` unvalidated-input pattern across ~31 wrappers). |

**Dedup outcome**: Rec #7 is **PARTIAL**. The gateguard exit-0 behavior is DONE — cite SPEC-HOOK-FACTFORCE-ADVISORY-001 (completed). The MOAI_HOOK_STDERR_LOG allowlist + gateguard state-file escaping = **NEW residual** owned by this SPEC (M7).

### Rec #8 — Coverage holes + shell/code defects (LOW)

| Overlap SPEC | Status | Verdict |
|---|---|---|
| SPEC-HOOK-DEADCODE-001 | `in-progress` | **NO overlap.** That SPEC is Go dead code (`internal/hook/agents/` + `lifecycle/`, `dual_parse.go`, `HookInput.Data` — all Go symbols). Rec #8 is shell + settings.json defects: MultiEdit matcher (settings.json.tmpl), csharp glob (sync-phase-quality-gate.sh — shell), agent-hook exec form (agent frontmatter), compact naming (handle-compact.sh.tmpl). No collision. |

**Dedup outcome**: Rec #8 is **NEW** — owned entirely by this SPEC (M8).

### Dedup summary table

| Rec | Milestone | Verdict | Residual owned by this SPEC |
|-----|-----------|---------|-----------------------------|
| #1 | M1 | NEW | All 5 REQs |
| #2 | M2 | NEW (update) | All 2 REQs (DIVECC created the file) |
| #3 | M3 | NEW | All 2 REQs |
| #4 | M4 | NEW | All 2 REQs |
| #5 | M5 | NEW | All 2 REQs |
| #6 | M6 | PARTIAL | REQ-HOC-014, REQ-HOC-015 (probe DONE per SESSIONSTART-PROBE-001) |
| #7 | M7 | PARTIAL | REQ-HOC-016, REQ-HOC-017 (exit-0 DONE per FACTFORCE-ADVISORY-001) |
| #8 | M8 | NEW | All 4 REQs |
| **GAP-STT-05** (non-Rec gap, INFO) | — | **DONE** (resolved by SESSIONSTART-PROBE-001 probe; report §6 says "no new action") | — |

---

## §E. Self-Verification (run-phase E1-E7 deliverables — reference)

Per `.claude/rules/moai/development/manager-develop-prompt-template.md` §E. The run-phase reports each item per the verification-claim-integrity 5-section format (Claim / Evidence / Baseline-attribution / Gaps / Residual-risk):

- **E1**: AC binary PASS/FAIL matrix (AC-HOC-001 through AC-HOC-036)
- **E2**: Cross-platform build result (`go build ./...` + `GOOS=windows GOARCH=amd64`)
- **E3**: Coverage measurement — N/A (no Go change; Go inspection only)
- **E4**: Subagent-boundary grep (C-HRA-008): `grep -rn 'AskUserQuestion' internal/hook/ internal/cli/ | grep -v _test.go` — 0 matches
- **E5**: Lint status (NEW vs baseline)
- **E6**: Branch HEAD + push state
- **E7**: Blocker report (if Go inspection reveals a handler-side defect requiring a separate SPEC)

---

## §F. Milestones (priority-based, no time estimates)

### M1 — Blocking gate JSON contract + PreToolUse exit-code (HIGH, first)

**Priority**: High. **REQs**: REQ-HOC-001, REQ-HOC-002, REQ-HOC-003, REQ-HOC-004, REQ-HOC-005. **Gaps**: GAP-GD-01, GAP-GD-02, GAP-STT-02, GAP-TOOL-01, GAP-LC2-01.

**Scope**:
- `internal/template/templates/.claude/hooks/moai/team-ac-verify.sh` — reject path JSON form (REQ-HOC-001)
- `internal/template/templates/.claude/hooks/moai/sync-phase-quality-gate.sh` — Stop block path hookSpecificOutput nesting (REQ-HOC-002)
- `internal/template/templates/.claude/hooks/moai/handle-pre-tool.sh.tmpl` — exit-code passthrough + stderr preservation (REQ-HOC-003) — **REQUIRES Go inspection of `internal/hook/pre_tool.go` first** to confirm which path the handler uses
- `internal/template/templates/.claude/hooks/moai/handle-session-start.sh.tmpl` — hookEventName addition (REQ-HOC-004) + live mirror
- `.claude/rules/moai/core/agent-common-protocol.md` — Ledger Closure clause (b) mirror (REQ-HOC-005) + template mirror

**Go inspection requirement**: Before editing `handle-pre-tool.sh.tmpl`, the run-phase MUST inspect `internal/hook/pre_tool.go` (and the dispatcher path) to confirm whether the Go handler emits exit 2 + stderr or stdout JSON `permissionDecision:deny` for the reject path. The wrapper adjustment follows the observed Go behavior.

### M2 — Doctrine refresh (MEDIUM, single pass)

**Priority**: Medium. **REQs**: REQ-HOC-006, REQ-HOC-007. **Gaps**: GAP-LC-01, GAP-STT-03, GAP-OBS-04, GAP-CEW-03, GAP-PE-02, GAP-PE-03, GAP-GD-06, GAP-TOOL-04, GAP-GD-03.

**Scope**:
- `.claude/rules/moai/development/hooks-system.md` — 8-point refresh (REQ-HOC-006) + template mirror
- `.claude/rules/moai/development/hook-independence.md` — row (g) timeout 10s→60s (REQ-HOC-007) + template mirror

### M3 — Async observation taps (MEDIUM)

**Priority**: Medium. **REQs**: REQ-HOC-008, REQ-HOC-009. **Gaps**: GAP-OBS-01, GAP-OBS-02, GAP-LC2-02.

**Scope**:
- `internal/template/templates/.claude/settings.json.tmpl` — add `"async": true` to UserPromptSubmit/Stop/SubagentStop harness-observe entries (REQ-HOC-008) + InstructionsLoaded entry (REQ-HOC-009) + live mirror

### M4 — Timeout headroom (MEDIUM)

**Priority**: Medium. **REQs**: REQ-HOC-010, REQ-HOC-011. **Gaps**: GAP-STT-01, GAP-LC-03.

**Scope**:
- `internal/template/templates/.claude/settings.json.tmpl` — raise TeammateIdle/TaskCompleted/PreCompact timeouts (REQ-HOC-010, REQ-HOC-011) + live mirror
- If the TeammateIdle/TaskCompleted raise is deemed excessive, document the fast-pre-check + async-deferred split as the alternative (the split would be Go work — out of scope per spec.md §G; document only)

### M5 — Matcher resolution + Go verification (MEDIUM)

**Priority**: Medium. **REQs**: REQ-HOC-012, REQ-HOC-013. **Gaps**: GAP-GD-04, GAP-CEW-02.

**Scope**:
- `internal/hook/file_changed.go` + matcher engine inspection — resolve FileChanged literal-vs-regex (REQ-HOC-012)
- `internal/hook/config_change.go` inspection — confirm ConfigChange block channel (REQ-HOC-013)
- `internal/template/templates/.claude/settings.json.tmpl` — FileChanged matcher adjustment if literal-only confirmed (REQ-HOC-012) + live mirror
- `internal/template/templates/.claude/hooks/moai/handle-config-change.sh.tmpl` — conditional stderr policy if Go uses exit 2 + stderr (REQ-HOC-013) + live mirror
- `.claude/rules/moai/development/hooks-system.md` — document runtime interpretation + template mirror

### M6 — Fail-open semantics correction (MEDIUM)

**Priority**: Medium. **REQs**: REQ-HOC-014, REQ-HOC-015. **Gaps**: GAP-CEW-01, GAP-PE-01.

**Scope**:
- `internal/template/templates/.claude/hooks/moai/handle-worktree-create.sh.tmpl` — fail-closed guard / header comment (REQ-HOC-014) + live mirror
- `internal/template/templates/.claude/hooks/moai/handle-permission-request.sh.tmpl` — fail-open warning priority (REQ-HOC-015) + live mirror; coordinate with the existing SessionStart probe (SPEC-HOOK-SESSIONSTART-PROBE-001)

### M7 — Input hardening (LOW)

**Priority**: Low. **REQs**: REQ-HOC-016, REQ-HOC-017. **Gaps**: GAP-STT-04, GAP-TOOL-02, GAP-GD-08.

**Scope**:
- Shared wrapper pattern (sourced common fragment or per-wrapper inline) — MOAI_HOOK_STDERR_LOG allowlist (REQ-HOC-016) — apply once in the shared pattern
- `internal/template/templates/.claude/hooks/moai/gateguard-fact-force.sh` — state-file escaping (REQ-HOC-017) + live mirror

### M8 — Coverage holes + defects (LOW)

**Priority**: Low. **REQs**: REQ-HOC-018, REQ-HOC-019, REQ-HOC-020, REQ-HOC-021. **Gaps**: GAP-GD-05, GAP-GD-07, GAP-OBS-03, GAP-LC-02.

**Scope**:
- `internal/template/templates/.claude/settings.json.tmpl` — MultiEdit matcher extension (REQ-HOC-018) + live mirror
- `internal/template/templates/.claude/hooks/moai/sync-phase-quality-gate.sh` — csharp glob fix (REQ-HOC-019) + live mirror
- Agent frontmatter `handle-agent-hook.sh` invocation — exec form (REQ-HOC-020) — 4 agent files (manager-develop, manager-docs, manager-spec, sync-auditor) + template mirrors
- `internal/template/templates/.claude/hooks/moai/handle-compact.sh.tmpl` — stale comment fix / rename consideration (REQ-HOC-021) + live mirror

---

## §G. Anti-Patterns

- **Skipping Go inspection before wrapper edit** — REQ-HOC-003 (PreToolUse), REQ-HOC-012 (FileChanged), REQ-HOC-013 (ConfigChange) all require Go source observation first. Adjusting a wrapper on assumption violates spec.md §E.2 and `verification-claim-integrity.md` §1.1 surface 3.
- **Editing only the template OR only the live mirror** — Template-First Rule (CLAUDE.local.md §2) requires both; `rule_template_mirror_test.go` CI catches drift.
- **Bundle-editing all 8 milestones in one commit** — keep milestones in separate commits (`feat(SPEC-HOOK-OFFICIAL-COMPLIANCE-001): M1 ...`, etc.) for bisectability.
- **Modifying the moai-binary resolution chain in 30 wrappers** — out of scope (spec.md §G); only WorktreeCreate (M6) and PermissionRequest (M6) get per-wrapper attention, and only as guards/comments, not chain edits.
- **Treating GAP-LC2-01 as already-fixed by SESSIONSTART-PROBE-001** — VERIFIED unfixed: `handle-session-start.sh.tmpl:44` still omits `hookEventName`. M1 REQ-HOC-004 owns it.

---

## §H. Cross-References

- spec.md: `.moai/specs/SPEC-HOOK-OFFICIAL-COMPLIANCE-001/spec.md`
- acceptance.md: `.moai/specs/SPEC-HOOK-OFFICIAL-COMPLIANCE-001/acceptance.md`
- Primary audit report: `.moai/reports/hooks-improvement-plan-20260710.html`
- Official baseline: `code.claude.com/docs/en/hooks`
- Manager-develop delegation template: `.claude/rules/moai/development/manager-develop-prompt-template.md`
- Template-First Rule: CLAUDE.local.md §2
- Verification-claim integrity: `.claude/rules/moai/core/verification-claim-integrity.md`
