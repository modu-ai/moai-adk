# Progress — SPEC-WORKTREE-BRANCH-GUARD-OPTIN-001

> Plan-phase skeleton. §E.2–§E.4 are populated by manager-develop (run-phase) and manager-docs (sync-phase); this agent emits only §E.1.

## §E.1 Plan-phase Audit-Ready Signal

- **Plan-phase artifacts emitted**: spec.md, plan.md, acceptance.md, progress.md (this file).
- **Tier**: M (standard set).
- **Era**: V3R6.
- **Status**: `draft` (set at creation per Status Transition Ownership — this agent owns `(none) → draft` only).
- **Pre-write self-checks executed**:
  - SPEC ID regex: `SPEC-WORKTREE-BRANCH-GUARD-OPTIN-001` → `PASS` (verified via Bash regex match against `^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$`).
  - SPEC ID dedup: no existing `.moai/specs/SPEC-WORKTREE-BRANCH-GUARD-OPTIN-001/` directory prior to creation.
  - Frontmatter 12 canonical fields present in spec.md (id, title, version, status, created, updated, author, priority, phase, module, lifecycle, tags) + era/tier/related_specs optional fields.
- **Evidence basis**: spec.md §D cites directly observed violations (`git stash list` / `git merge-base`) from this session — not inferred, not grep-only.
- **Design decisions recorded**: plan.md §E (D1 gate-at-call-site, D2 config struct placement, D3 regex tightening, D4 maintainer opt-in location).
- **Open clarification**: plan.md §E D3 leaves the implementer's preferred regex syntax (anchored vs. lookahead) as a behavior-not-syntax choice — the SPEC mandates the behavior (REQ-2), the implementer picks syntax.

## §E.2 Run-phase Evidence

| AC ID | Status | Verification Command | Observed Output |
|-------|--------|---------------------|-----------------|
| AC-REQ-1a | PASS | `go test ./internal/hook/ -run 'TestPreTool_BranchGuard_ConfigGate_DefaultOff' -v` | `--- PASS: TestPreTool_BranchGuard_ConfigGate_DefaultOff` — default config (Enabled=false), `git reset --hard` / `git switch` NOT denied |
| AC-REQ-1b | PASS | (covered by DefaultOff `explicit_false_switch`) | enabled=false explicit, `git switch feature/x` NOT denied |
| AC-REQ-2a | PASS | `go test ./internal/hook/ -run 'TestBranchStatePatterns_TrueNegatives' -v` | `git stash list` no-match (excluded by tightened regex) |
| AC-REQ-2b | PASS | (covered by TrueNegatives `git stash show -p stash@{0}`) | `git stash show -p` no-match |
| AC-REQ-2c | PASS | (covered by TrueNegatives `git merge-base HEAD origin/main`) | `git merge-base` no-match (trailing-whitespace anchor) |
| AC-REQ-2d | PASS | `go test ./internal/hook/ -run 'TestBranchStatePatterns_DangerousFormsPreserved' -v` | bare `git stash` + push/pop/apply/drop still match (deny) |
| AC-REQ-2e | PASS | (covered by DangerousFormsPreserved `git merge feature/x`) | `git merge feature/x` still matches (deny) |
| AC-REQ-3 | PASS | `go test ./internal/hook/ -run 'TestPreTool_BranchGuard_ConfigGate_FlagFlips' -v` | flipping ONLY `BranchGuard.Enabled` false→true flips decision allow→deny (non-vacuity) |
| AC-REQ-4 | PASS | `grep -rn 'enabled.*true' internal/template/templates/ \| grep -i branch` → 0 matches; `go test ./internal/config/ -run 'TestNewDefaultConfigContainsAllSections'` | default false asserted; no `enabled: true` in templates |
| AC-REQ-5 | PASS | `go test ./internal/template/ -run 'TestSanitizedPairParity\|TestTemplateNoInternalContentLeak\|TestRuleTemplateMirror'` | all 3 green; both rule-mirror sides updated (v1.2.0) |
| AC-REQ-6a | PASS | `TestPreTool_BranchGuard_ConfigGate_Exemptions/MOAI_BRANCH_GUARD_EXEMPT_env` | Enabled=true + EXEMPT=1 → allow |
| AC-REQ-6b | PASS | `TestPreTool_BranchGuard_ConfigGate_Exemptions/manager-git_agent_identity` | Enabled=true + AgentType=manager-git → allow |
| AC-REQ-6c | PASS | `TestPreTool_BranchGuard_ConfigGate_FailOpen` | Enabled=true + non-git projectDir → allow (fail-open preserved) |
| AC-REQ-7 | DEFERRED | (CHANGELOG entry — owned by manager-docs sync-phase) | n/a — run-phase does not own CHANGELOG |

### M2 RED evidence (verbatim, captured before GREEN)

`git stash list`, `git stash show`, `git merge-base` matched the old patterns:
```
branch_guard_test.go:310: matchBranchStateCommand("git stash show -p stash@{0}") = ("git stash", true), want false
branch_guard_test.go:310: matchBranchStateCommand("git merge-base HEAD origin/main") = ("git merge", true), want false
branch_guard_test.go:310: matchBranchStateCommand("git stash show") = ("git stash", true), want false
branch_guard_test.go:310: matchBranchStateCommand("git merge-base --is-ancestor A B") = ("git merge", true), want false
branch_guard_test.go:310: matchBranchStateCommand("git stash list") = ("git stash", true), want false
branch_guard_test.go:310: matchBranchStateCommand("git stash show -p") = ("git stash", true), want false
--- FAIL: TestBranchStatePatterns_TrueNegatives (0.00s)
```

### M3 RED evidence (verbatim, captured before GREEN)

Gate not yet wired at call site — Enabled=false still denied; Enabled=true yielded "ask" (checkBashCommand preempt on `git reset --hard`; switched the FlagFlips command to `git switch -c ...` which cleanly reaches checkBranchState):
```
pre_tool_branch_guard_optin_test.go:80: Handle(enabled=false, "git switch feature/x") = deny; want non-deny (REQ-1 default-off)
pre_tool_branch_guard_optin_test.go:80: Handle(enabled=false, "git switch -c feat/y") = deny; want non-deny (REQ-1 default-off)
--- FAIL: TestPreTool_BranchGuard_ConfigGate_DefaultOff (0.15s)
    pre_tool_branch_guard_optin_test.go:132: Handle(enabled) = "ask", want "deny" — flipping ONLY the flag MUST flip the decision (AC-REQ-3 non-vacuity)
--- FAIL: TestPreTool_BranchGuard_ConfigGate_FlagFlips (0.10s)
```

## §E.3 Run-phase Audit-Ready Signal

- **run_status**: pass
- **ac_pass_count**: 13 (AC-REQ-1a,1b,2a,2b,2c,2d,2e,3,4,5,6a,6b,6c)
- **ac_fail_count**: 0 (AC-REQ-7 deferred to sync-phase — CHANGELOG owned by manager-docs)
- **preserve_list_post_run_count**: 0 (isExemptAgent body, MOAI_BRANCH_GUARD_EXEMPT check, manager-git identity check, fail-open path, handle-pre-tool.sh.tmpl dispatch, settings.json.tmpl PreToolUse matcher — all byte-identical / behavior-identical)
- **new_warnings_or_lints_introduced**: 0 (golangci-lint clean; no new findings)
- **cross_platform_build.darwin**: pass (`go build ./...` exit 0)
- **cross_platform_build.windows**: pass (`GOOS=windows GOARCH=amd64 go build ./...` exit 0)
- **total_run_phase_files**: 8 (types.go, defaults.go, defaults_test.go, branch_guard.go, branch_guard_test.go, pre_tool.go, pre_tool_branch_guard_optin_test.go, pre_tool_test.go + 2 rule mirrors + pre_tool_branch_guard_integration_test.go + CLAUDE.local.md)
- **m1_to_mN_commit_strategy**: single commit covering M1-M7 (Tier M, behavior-coherent unit)
- **blocker**: M6 local-config path (`.moai/config/local/`) is NOT gitignored — documented in CLAUDE.local.md §22.9 BLOCKER note; the gitignore addition is out of this SPEC's scope (separate chore). The maintainer opt-in file is NOT created in the committed tree (would violate §25 template neutrality); the maintainer creates it locally.

## §E.4 Sync-phase Audit-Ready Signal

- **sync_status**: completed
- **sync_complete_at**: 2026-07-30
- **sync_commit_sha**: 9eca6cadc
- **changelog_entry_position**: `[Unreleased] → ### Changed` (SPEC-WORKTREE-BRANCH-GUARD-OPTIN-001 entry, AC-REQ-7 user-facing default-off communication)
- **frontmatter_status_transitions.spec**: `in-progress → implemented → completed` (single sync commit, merged close per Status Transition Ownership Matrix)
- **frontmatter_status_transitions.updated_refreshed**: 2026-07-30 (spec.md `updated:` field refreshed to sync commit date)
- **b12_self_test_a** (pre-emission grep): `grep -c 'SPEC-WORKTREE-BRANCH-GUARD-OPTIN-001' CHANGELOG.md` returned 0 BEFORE emission (no duplicate)
- **b12_self_test_b** (AC count match): acceptance.md SSOT AC rows = 14 (AC-REQ-1a/1b/2a/2b/2c/2d/2e/3/4/5/6a/6b/6c/7); CHANGELOG entry references the implementation scope covering all 14 (run-phase owned 1a-6c, sync-phase owned AC-REQ-7)
- **b12_self_test_c** (file path verification): every implementation file path claimed in the CHANGELOG entry (`internal/config/types.go`, `internal/config/defaults.go`, `internal/hook/branch_guard.go`, `internal/hook/pre_tool.go`, `.claude/rules/moai/workflow/main-checkout-branch-guard.md`, `CLAUDE.local.md`) verified to exist via `ls` before committing
- **canary_compliance_check.template_neutrality**: the sync commit did NOT edit `internal/template/templates/.moai/config/sections/workflow.yaml` to set `enabled: true` — template default remains `enabled: false` (§25)
- **canary_compliance_check.body_untouched**: spec.md / plan.md / acceptance.md BODY content unchanged on the sync commit (only spec.md frontmatter `status:` + `updated:` modified)
- **mx_tag_validation**: M5 rule mirror updated on both sides in run-phase (v1.2.0); no new @MX annotations required in sync-phase (the sync commit is doc-only: CHANGELOG + frontmatter + progress.md)
