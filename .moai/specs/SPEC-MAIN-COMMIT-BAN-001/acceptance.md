# Acceptance — SPEC-MAIN-COMMIT-BAN-001

> One falsifiable AC per REQ cluster. Guard ACs follow RED→GREEN: AC-4 is the RED-first core
> (observed red on the pre-implementation tree for the stated reason: `checkBranchState` has no
> commit-family matcher, so `git commit` on a protected branch in the primary returns allow).
> Scope verify rows: AC-3 (Scope 1), AC-11 (Scope 2), AC-13 (Scope 3), AC-15 (Scope 4), AC-16 (umbrella).

## §D AC Matrix

| AC ID | REQ | Severity | Summary |
|---|---|---|---|
| AC-1 | REQ-1.1 | MUST-PASS | Local AGENTS.md §2: unconditional commit permission removed; protected-branch exception naming `main` present |
| AC-2 | REQ-1.2 | MUST-PASS | Template AGENTS.md.tmpl carries the generic protected-branch clause; neutrality + leak guards green on touched files |
| AC-3 | REQ-1.3/1.4 | MUST-PASS | Scope 1 verify: worktree-run grep 4/4 (`AGENTS.local.md` §4.1, lane-protocol §1, gitflow chain, git-workflow-doctrine) + lead-run primary probe (`CLAUDE.local.md` §4.1, post-disposition) |
| AC-4 | REQ-2.1/2.2 | MUST-PASS | RED→GREEN: commit/revert/cherry-pick on protected branch @ primary + configured + non-exempt → deny, `BRANCH_GUARD_VIOLATION:` prefix |
| AC-5 | REQ-2.1 | MUST-PASS | Commit on non-protected branch @ primary + configured → allow |
| AC-6 | REQ-2.1 | MUST-PASS | Commit on protected branch @ worktree cwd → allow (discriminant holds) |
| AC-7 | REQ-2.5 | MUST-PASS | Empty deny list → allow AND zero HEAD-resolver invocations (stub count == 0) |
| AC-8 | REQ-2.3 | MUST-PASS | HEAD-resolution failure → allow + advisory appended to `.moai/logs/branch-guard-audit.log`; detached HEAD → allow |
| AC-9 | REQ-2.1 | MUST-PASS | Both exemption axes (`manager-git` identity; env sentinel) suppress the deny |
| AC-10 | REQ-3.1/3.2/3.4 | MUST-PASS | Config surface: `deny_commits_on` in struct/defaults/template/local; D8 presence test green; `^TestStructYAMLSymmetry_` green; no new opt-in flag; template `enabled` stays false |
| AC-11 | REQ-2 (verify) | MUST-PASS | Scope 2 verify: scoped go vet + golangci-lint (CI version) + `go test ./internal/hook/ ./internal/core/git/...` green |
| AC-12 | REQ-4.1-4.6 | MUST-PASS | Residue procedure documented with fixed ordering, origin/main target + rationale, modified-tracked-set boundary named, restore-FORBIDDEN warning, rehearsal step (real-state-class recipe), re-armed switch-to-main post-condition, documented-exception note, incident reference, operator-side CLAUDE.local.md sync step |
| AC-13 | REQ-4.4 (verify) | MUST-PASS | Scope 3 verify: scratch-clone rehearsal at the primary's real state class executed; observed-refusal log at `.moai/reports/t1337/`; real primary untouched during run phase (ref-state read-only checks only) |
| AC-14 | REQ-3.3 | MUST-PASS | `git_strategy.manual.lead_push_threshold` exists in `ModeProfile` (default 0) + template `git-strategy.yaml.tmpl` + LOCAL yaml value `20`; D8 presence test green |
| AC-15 | REQ-5.1-5.3 | MUST-PASS | Scope 4 verify: gitflow chain doc names the key + counter command + window interaction; `AGENTS.local.md` §4.1 cites the key (worktree-run); machine-readable parse asserts local value 20 |
| AC-16 | (umbrella) | MUST-PASS | Sentinel-phrase grep batch: every file named in plan §D5/M3 shows its expected phrase in one enumerated command |

## §D.1 Severity / Traceability

- All ACs MUST-PASS — Tier M, operator-directed policy codification, behavior-affecting (guard).
- Traceability: AC-1..3 → REQ-1; AC-4..9 → REQ-2; AC-10, AC-14 → REQ-3; AC-12..13 → REQ-4;
  AC-15 → REQ-5; AC-16 cross-cutting.

## §D.2 Given-When-Then Scenarios (load-bearing)

### AC-4 — Protected-commit deny (RED→GREEN core)

```
GIVEN Workflow.BranchGuard.Enabled == true AND deny_commits_on == ["main"]
  AND a Bash HookInput with command "git commit -m x"
  AND the command cwd resolving to the primary checkout
  AND HEAD branch resolving to "main"
  AND no exemption env / agent type
WHEN the preToolHandler.Handle is invoked
THEN the hook returns DecisionDeny
  AND reason starts with "BRANCH_GUARD_VIOLATION:"
  AND reason names the protected branch ("main") and "primary checkout"
  AND reason does NOT suggest delegating to a manager-git agent
```

**RED cell** (baseline pinned): measured-red premise established at iter1 audit SHA
`145c3d98c` — `checkBranchState` carries no commit-family matcher (pattern set read verbatim),
so the commit case falls through the allow path; at run-phase RED execution the four elements
(command `go test ./internal/hook/ -run '^TestProtectedCommit$' -count=1`, verbatim output, exit
code, actual tree SHA) are captured and re-pinned in the evidence path (verification-completeness
§2.1). **GREEN cell**: M2 flips it; the anchored selector → all subtests PASS.

### AC-7 — Inert default (zero-cost short-circuit)

```
GIVEN deny_commits_on == [] (the shipped default)
  AND a Bash HookInput with command "git commit -m x" at a primary-checkout cwd
WHEN checkProtectedCommit runs
THEN it returns allow
  AND the swapped ResolveHeadBranch stub records 0 invocations
```

**Falsifiable**: subtest asserts `stub.calls == 0` (package-var swap idiom, sibling of the M6
deny-origin test).

### AC-8 — Fail-open preserved

```
GIVEN deny_commits_on == ["main"] and the guard enabled
  AND ResolveHeadBranch returns an error (git missing / rev-parse non-zero)
WHEN checkProtectedCommit runs
THEN decision is allow
  AND a structured advisory line is appended to .moai/logs/branch-guard-audit.log
  AND the resolved cwd appears in the advisory
```

Detached-HEAD twin: resolver returns `("", nil)` → allow, no advisory requirement.

### AC-13 — Rehearsal evidence, real tree untouched

```
GIVEN a scratch clone of the repository in a temp directory at the primary's REAL state class
  (clone → git switch c8f245c2c → develop-state content materialized as unstaged modifications)
WHEN the Scope 3 procedure steps execute there (inventory + diff precondition + branch switch + fetch + branch -f main origin/main)
THEN the ACTUAL refusal set (modified-tracked-set / untracked-overwrite forms) is transcribed
  to .moai/reports/t1337/rehearsal evidence — a clean clone at origin/main is NOT the
  rehearsal subject (it would observe silent removal, not the refusal; audit F5)
  AND the REAL primary checkout's branch state is unchanged during run phase
```

**Falsifiable**: ref-state reads from the worktree against the shared object store, frozen at
rehearsal pre-flight (R2 form): `MAIN_AT_START=$(git rev-parse main)` captured before the
rehearsal, then `git rev-parse main` == `$MAIN_AT_START` after it (main is compared to its own
frozen value — no moving ref decides the invariant); rehearsal evidence file exists and names
each step's observed output (no assumed behavior).

### AC-14 — Machine-readable threshold

```
GIVEN .moai/config/sections/git-strategy.yaml (LOCAL, tracked)
WHEN parsed (yq or a Go test reading the file)
THEN git_strategy.manual.lead_push_threshold == 20
  AND internal/config ModeProfile exposes the field with default 0
  AND the D8 presence test (TestTemplateConfigCarriesMainCommitBanKeys) is green
  AND the symmetry family stays green — verified by the scoped package run in §D.3
      (context: the symmetry harness cannot see these keys — workflow/git-strategy are
      outside symmetryCases, one-level-deep checkSymmetry, measured; the presence test is
      the instrument for the new keys, plan D8; no -run selector is cited for the family
      because its `TestStructYAMLSymmetry_*` names cannot be fully anchored without
      enumerating all seven)
```

**Falsifiable**:
```bash
yq '.git_strategy.manual.lead_push_threshold' .moai/config/sections/git-strategy.yaml   # → 20
go test ./internal/config/ -run '^TestTemplateConfigCarriesMainCommitBanKeys$' -count=1  # → ok
```

### AC-16 — Sentinel-phrase probe sets (umbrella; audit F1/F6 split)

**Worktree-run probe set** (lane executes; designated sentinel = the literal token
`commit-dead`, embedded by the D5 edit spec — a clause-specific phrase, not a generic verb):
```bash
grep -l 'commit-dead' AGENTS.md AGENTS.local.md .claude/rules/local/gitflow-lane-protocol.md .moai/docs/gitflow-integration-chain.md .moai/docs/git-workflow-doctrine.md   # → 5 files
grep -l 'branch -f main origin/main' .moai/docs/gitflow-integration-chain.md        # → hit
grep -l 'lead_push_threshold' .moai/docs/gitflow-integration-chain.md AGENTS.local.md   # → 2 files
grep -l 'deny_commits_on' .moai/config/sections/workflow.yaml internal/config/types.go  # → 2 files
grep -l 'commit-protected' internal/template/templates/AGENTS.md.tmpl               # → hit
```
**Primary-run probe set** (LEAD executes at the primary; the lane reports it as a pending
checklist item):
```bash
grep -l 'commit-dead' <primary>/CLAUDE.local.md    # → hit (post-disposition, REQ-4.6)
```
Expected: every probe non-zero; any empty result is a Gap, not a pass. All worktree-set target
files measured to exist in this tree (audit F1 re-measurement).

## §D.3 Verification batch (run/sync phase, lane-local scope per AGENTS.md §4)

```bash
go test ./internal/hook/ ./internal/core/git/... ./internal/config/... -count=1   # scoped, not ./...
go vet ./internal/hook/ ./internal/core/git/... ./internal/config/...
golangci-lint run internal/hook/... internal/config/...   # CI-pinned version (v2.1.6)
go test ./internal/template/ -run '^(TestTemplateNeutralityAudit|TestSanitizedPairParity|TestRuleTemplateMirrorDrift|TestDeclaredRuleMirrorForks|TestTemplateNoInternalContentLeak)$' -count=1
go test ./internal/config/ -run '^TestTemplateConfigCarriesMainCommitBanKeys$' -count=1
yq '.git_strategy.manual.lead_push_threshold' .moai/config/sections/git-strategy.yaml
yq '.workflow.branch_guard.deny_commits_on' .moai/config/sections/workflow.yaml
```

Gaps declared up front: full-suite `go test ./...` is CI's verdict (lane-local scoping rule);
the real-primary residue execution is the operator's post-merge act, verified by the procedure's
own post-steps, not by this SPEC's ACs; the primary-run probe set (AC-16) is lead-executed, not
lane-executed.
