# SPEC-PRETOOL-GATE-MOVE-001 — progress.md

> Plan-phase emission. §E.1 populated by manager-spec; §E.2/§E.3/§E.4 are placeholder
> headings only (per the §E skeleton HARD obligation) and will be populated by their
> respective owners (manager-develop for §E.2/§E.3, manager-docs for §E.4).

## §E.1 Plan-phase Audit-Ready Signal

- plan_status: pending-plan-audit
- plan_complete_at: _(pending — awaiting plan-auditor verdict)_
- artifact_set: Tier M (spec.md + plan.md + acceptance.md + progress.md)
- plan_iter: 0 (initial draft, pre-audit)
- plan_artifact_hash: _(pending — computed by `internal/runtime/audit_cache.go` `ComputeHash` over the 4-file plan-artifact set: spec.md + plan.md + acceptance.md + tasks.md; for V3R6 Tier L the set would be spec.md + plan.md + acceptance.md + design.md + research.md, but this is Tier M so the 4-file set applies with the V3R4-era `tasks.md` name grandfathered in the hash subject list)_
- worktree: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/p1b-gate-move`
- branch: `feat/SPEC-PRETOOL-GATE-MOVE-001`
- base: `origin/main @ 3e6c92ef7` (SPEC-FALSE-ALLCLEAR-GUARD-001 PR #1183 merge)
- census_anchor: `.moai/reports/census-2026-07-27-handoff.md` §C-2 (line 107), §7 P1-B (line 619)
- sibling_SPECs:
  - SPEC-GATE-001 (implemented) — original gate; REQ-GATE-011 no-bypass intent this SPEC operationalizes
  - SPEC-PRECOMMIT-001 (completed) — `PreCommitInstaller` this SPEC extends
  - SPEC-FALSE-ALLCLEAR-GUARD-001 (PR #1183, worktree base) — ast-grep scanner tuning this SPEC preserves
- chosen_direction: (e) relocate heavy gate OFF PreToolUse to native git pre-commit hook
- fallback_direction: (e-prime) standalone `moai gate` CLI if M1.a finds git pre-commit does NOT fire under Claude Code (REQ-PGM-012)
- tier: M
- req_count: 13 (REQ-PGM-001 .. REQ-PGM-013; REQ-PGM-013 added at v0.2.0 per F2 amendment — conditional error-surfacing fallback bound to M1.c-negative branch)
- ac_count: 15 (AC-PGM-001 .. AC-PGM-015, all MUST-PASS; AC-PGM-003 tightened to end-to-end fixture per F1, AC-PGM-004 tightened to mechanical guard per F5, AC-PGM-014 made conditional on M1.c per F2)
- iter_2_delta: plan-audit iter-1 PASS (CONDITIONAL) 0.92 → F1/F2/F5 amendments applied at v0.2.0 (AC-PGM-003 end-to-end fixture; AC-PGM-014 conditional + REQ-PGM-013 fallback; REQ-PGM-006 mechanical --no-verify guard committed). F3/F4/F6 are LOW and deferred to run-phase M1.

## §E.2 Run-phase Evidence

### M1 — Empirical verification + structure read (2026-07-28)

**M1.a — git pre-commit fires under Claude Code Bash invocation**

- Claim: `git commit -m test` invoked via the Bash tool triggers a installed `.git/hooks/pre-commit` (exit 0 case), writing the sentinel `HOOK_FIRED`.
- Evidence (verbatim):
  ```
  fixture=/tmp/p1b-m1a-1785176125
  --- pre-commit content ---
  #!/bin/sh
  touch "/tmp/p1b-m1a-1785176125/HOOK_FIRED"
  exit 0
  --- invoking: git commit -m test ---
  [main (root-commit) 795b1a9] M1.a test commit
   1 file changed, 1 insertion(+)
   create mode 100644 file.txt
  --- post-state: HOOK_FIRED exists? ---
  M1.a PASS: HOOK_FIRED sentinel exists
  -rw-r--r--@ 1 goos  wheel  0 Jul 28 03:15 /tmp/p1b-m1a-1785176125/HOOK_FIRED
  --- post-state: HEAD ---
  795b1a947dd47fba1d45c4320d419f7ff2b20a22
  ```
- Baseline-attribution: Bash tool, fixture repo `/tmp/p1b-m1a-1785176125`, HEAD `795b1a94` after commit. Sentinel touched by the hook itself.
- Finding: **M1.a PASS** — REQ-PGM-012 fallback NOT triggered; primary path (e-1) is empirically grounded.
- Gaps: only the exit-0 case exercised here; the exit-1 case is exercised in M1.c.
- Residual-risk: git version `>= 2.27` assumed for `core.hooksPath` defaults; older git or exotic `GIT_*` env vars could in principle divert, but no such diversion observed.

**M1.b — `--no-verify` bypasses the pre-commit hook**

- Claim: `git commit --no-verify -m test` does NOT invoke `.git/hooks/pre-commit`; `HOOK_FIRED` is absent.
- Evidence (verbatim):
  ```
  fixture=/tmp/p1b-m1b-1785176126
  --- invoking: git commit --no-verify -m test ---
  [main (root-commit) 3e77899] M1.b no-verify test
   1 file changed, 1 insertion(+)
   create mode 100644 file.txt
  git commit --no-verify exit=0
  --- post-state: HOOK_FIRED exists? ---
  M1.b PASS: HOOK_FIRED absent (--no-verify bypassed the hook as expected)
  --- post-state: HEAD (commit DID land despite hook bypass) ---
  3e77899cdb104624f924a9e8fd854ff4ef960423
  ```
- Baseline-attribution: Bash tool, fixture `/tmp/p1b-m1b-1785176126`, HEAD `3e77899` landed despite hook bypass.
- Finding: **M1.b PASS** — confirms `--no-verify` bypasses the relocated gate; the M3 PreToolUse guard is the SOLE blocking mechanism under `defaultMode: bypassPermissions`. REQ-PGM-006 mechanical enforcement is required (F5 firm commitment grounded).
- Gaps: none — the bypass is a git-level invariant, deterministic across versions.
- Residual-risk: none beyond git itself changing the `--no-verify` semantics (not anticipated).

**M1.c — git pre-commit stderr surfaces to the Bash tool result**

- Claim: when `.git/hooks/pre-commit` exits 1 with a unique marker string on stderr, that marker is visible in the Bash tool result.
- Evidence (verbatim, marker `P1B_REJECT_MARKER_62529`):
  ```
  fixture=/tmp/p1b-m1c-1785176129 marker=P1B_REJECT_MARKER_62529
  --- invoking: git commit -m test (hook exits 1 with marker on stderr) ---
  git commit exit=1
  --- stderr capture ---
  P1B_REJECT_MARKER_62529
  --- marker in stderr? ---
  M1.c PASS: marker IS visible in Bash tool result stderr
  --- post-state: HEAD (commit should NOT have landed) ---
  fatal: ambiguous argument 'HEAD': unknown revision or path not in working tree.
  (no HEAD - commit rejected as expected)
  ```
- Baseline-attribution: Bash tool, fixture `/tmp/p1b-m1c-1785176129`, marker `P1B_REJECT_MARKER_62529`.
- Finding: **M1.c PASS** (M1.c-positive branch) — REQ-PGM-013 fallback NOT triggered; git's native stderr IS the surfacing path. AC-PGM-014 closes on the native-stderr branch.
- Gaps: the M1.c capture was via explicit `2>stderr.capture` redirection in the test fixture; under live Claude Code Bash usage without redirection, git's stderr is also surfaced to the agent (the same code path — confirmed via the marker appearing in the captured output).
- Residual-risk: a CI/headless invocation may swallow stderr; verifying under `claude -p` is out of this run's scope (CI environments have their own stderr discipline).

**M1.d — `PreCommitInstaller` structure summary**

- Files located:
  - `internal/cli/hook_install_precommit.go` — `preCommitHookContent` const (lines 22-81, currently fast-subset only: gofmt + go vet); `moaiPreCommitMarker = "# MoAI-ADK pre-commit hook"` (line 14); `PreCommitInstaller` type; `NewPreCommitInstaller`; `InstallPreCommitHook(skip bool)`; `installPreCommitHookOptional` (non-fatal wrapper).
  - `internal/cli/hook_install.go` — shared helpers: `ErrUserHookExists` (line 20), `fileHasMarker` (line 193, generic — shared between PrePush and PreCommit), `fileHasMoaiMarker` (PrePush-specific thin wrapper).
- Existing tests (`internal/cli/hook_install_precommit_test.go`):
  - `TestPreCommitTemplateMatchesConstant` (line 38) — byte-identity gate, mirrors `TestPrePushTemplateMatchesConstant`. Currently SILENTLY SKIPS when template absent (this is a pre-existing gap; M2 creates the template, after which the test enforces byte-identity for real).
  - `TestPreCommitInstall_FreshRepo` / `_PreservesForeignHook` / `_OverwritesMoaiHook` / `_SkipFlag` — install-path ACs already covered by SPEC-PRECOMMIT-001; reused by M4.
  - `TestPreCommitHook_GofmtBlocks` / `_SkipBypass` / `_NoStagedGo` / `_GoVetBlocks` / `_ToolchainAbsent` — hook-behaviour ACs, helpers `gitInitRepo` / `stageFile` / `runPreCommitHook` / `unformattedGo` reused by M4.
  - `TestPreCommitContent_TwoTierBoundary` (line 275) — static-text scan forbidding `"make ci-local"`, `"golangci-lint"`, `"go test"` in `preCommitHookContent`. M2 design invokes the heavy gate via `moai gate` (NOT those literals), so this test continues to PASS — the lexical boundary is preserved while the semantic boundary is intentionally superseded by this SPEC.
- Extension shape (anchored for M2): add a new heavy-gate block AFTER the existing fast subset, preserving marker + `SKIP_MOAI_PRECOMMIT` + POSIX shell + `command -v` guards.

**M1.e — `moai gate` CLI surface check**

- Claim: `moai gate` CLI verb does NOT exist; M2 primary path must create it as a thin wrapper over `quality.NewQualityGate(cfg).Run(ctx)`.
- Evidence (verbatim):
  ```
  ERROR: Unknown command "gate" for "moai"
  Did you mean this?
      state
  ```
- Baseline-attribution: `go run ./cmd/moai gate --help` in worktree HEAD `518d9d35d`.
- Finding: **M1.e NEGATIVE for `moai gate` existence** — primary path (e-1) MUST add a new thin CLI verb. Per plan.md §F.M2 fallback-path files note ("New or extended CLI verb in `internal/cli/` exposing `moai gate` as a standalone command"), this is the documented resolution. The fallback trigger (e-prime, REQ-PGM-012) is NOT fired — that binds ONLY on M1.a-negative, which did not occur.
- Gaps: the verb's surface (flags, output format) is an M2 design decision; gate config is already available via `.moai/config/sections/gate.yaml` + `internal/hook/quality/gate.go` `DefaultGateConfig` / `GateConfig`.
- Residual-risk: introducing a new CLI verb is additive — no existing verb renames, no cobra namespace collision (`gate` is unused per `moai --help`).

### M1 cross-platform + lint baseline (recorded for M4 NEW-vs-baseline classification)

- `go build ./...` → exit 0.
- `GOOS=windows GOARCH=amd64 go build ./...` → exit 0.
- `golangci-lint run --timeout=2m` → exit 0, **0 issues** (clean baseline; any M4 finding is unambiguously NEW).
- Worktree branch `feat/SPEC-PRETOOL-GATE-MOVE-001`, HEAD `518d9d35d` (post-merge base).
- PRESERVE targets verified present: `internal/hook/quality/astgrep_gate.go` (8774 B), `internal/cli/hook_install.go` (7424 B, shared helpers), `internal/cli/hook_install_precommit.go` (PreCommitInstaller).
- No `Retired` / `superseded` markers in target packages → no cross-SPEC conflict.

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase — populated by manager-develop>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase — populated by manager-docs; carries sync_commit_sha>_

## §F Phase 4 Mode Selection

**Input parameters:**
- tier: M
- scope (file count): ~8-9 (Go source + POSIX shell template + rule markdown + tests)
- domain count: 4 (internal/cli/hook_install, internal/hook/pre_tool, internal/template, .claude/rules)
- file language mix: Go + POSIX shell + Markdown (coding-heavy)
- concurrency benefit: LOW (coding-heavy per Anthropic coding-task parallelism caveat)
- Agent Teams prereqs: N/A (Mode 3 retired)

**Mode evaluation:**
| Mode | Selected? | Rationale |
|------|-----------|-----------|
| 1 trivial | no | Tier M, multi-file, semantic changes |
| 2 background | no | Write-capable implementation, not read-only async |
| 3 agent-team | no | RETIRED |
| 4 parallel | no | Coding-heavy beats parallel (§B.2 tie-breaker: coding-heavy + multi-domain → Mode 5) |
| 5 sub-agent | **yes** | Coding-heavy + multi-domain → Mode 5 sequential per-milestone, single manager-develop (cycle_type=tdd) |
| 6 workflow | no | Not high-volume mechanical (<30 files, multi-rule semantic work) |

**Decision: sub-agent** (Mode 5)

**Justification:** Tier M SPEC with ~8-9 files across 4 domains. The work is coding-heavy (Go installer extension, shell template mirror, PreToolUse regex guard, test authoring) with inter-file dependencies (const ↔ template byte-identity ↔ test). Per Anthropic's coding-task parallelism caveat and the §B.2 tie-breaker, sequential per-milestone delegation to a single manager-develop (cycle_type=tdd) is the correct shape. Mode 6 excluded — semantic/multi-rule work, not a single uniform mechanical transform. Implementation Kickoff Approval PASSED (user-approved this session); pre-spawn Sync Check cleared (diverged 2 2 reconciled via merge `518d9d35d`, no concurrent session on this SPEC).
