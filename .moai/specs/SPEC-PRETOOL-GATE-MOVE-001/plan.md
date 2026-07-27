# SPEC-PRETOOL-GATE-MOVE-001 — Implementation Plan

> Tier M plan-phase artifact. Cross-references spec.md (requirements) and acceptance.md (AC matrix).
> Origin: census Priority-1 item P1-B (`.moai/reports/census-2026-07-27-handoff.md` line 619, defect §C-2 line 107).

## §A Context

### §A.1 Mission summary

Tier M SPEC fixing census C-2 (CRITICAL): the PreToolUse commit-quality gate's deny verdict is silently dropped because the gate needs up to 210s (vet 30s + lint 60s + test 120s, synchronous) but the PreToolUse hook budget is 5s. User-approved direction (e): relocate the heavy gate OFF PreToolUse to a native git pre-commit hook (runs in the user's shell, outside Claude Code's 5s budget). PreToolUse retains only fast security checks (ast-grep, frozen-zone).

### §A.2 Worktree + branch + base

- Worktree: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/p1b-gate-move`
- Branch: `feat/SPEC-PRETOOL-GATE-MOVE-001`
- Base: `origin/main @ 3e6c92ef7` (= SPEC-FALSE-ALLCLEAR-GUARD-001 PR #1183 merge)
- All work happens inside the worktree; git operations via `git -C /Users/goos/MoAI/moai-adk-go/.claude/worktrees/p1b-gate-move`.

### §A.3 SPEC artifacts

| Artifact | Path (worktree-relative)              | Owner           |
|----------|---------------------------------------|-----------------|
| spec.md  | `.moai/specs/SPEC-PRETOOL-GATE-MOVE-001/spec.md`       | manager-spec    |
| plan.md  | `.moai/specs/SPEC-PRETOOL-GATE-MOVE-001/plan.md`       | manager-spec    |
| acceptance.md | `.moai/specs/SPEC-PRETOOL-GATE-MOVE-001/acceptance.md` | manager-spec |
| progress.md   | `.moai/specs/SPEC-PRETOOL-GATE-MOVE-001/progress.md`   | manager-spec (§E.1) → manager-develop (§E.2/§E.3) → manager-docs (§E.4) |

### §A.4 Plan-phase artifact set

Tier M = 3 artifacts + progress.md = 4 files. No research.md (M1 unknowns are empirical-verification, not literature research — fixture-based tests resolve them, not domain deep-dives).

### §A.5 PRESERVE list (do NOT touch — scope discipline)

- `internal/hook/quality/astgrep_gate.go` + ast-grep PreToolUse scanner path (just tuned by SPEC-FALSE-ALLCLEAR-GUARD-001 PR #1183).
- Frozen-zone guard PreToolUse path.
- `internal/hook/quality/gate.go` `IsGitCommit` regex (still used by PreToolUse fast path).
- `QualityGate.detectToolchain()` + 16-language `toolchains` table.
- SPEC-PRECOMMIT-001 fast-subset behavior (`gofmt -l` + `go vet` on staged Go files) — **extended, not replaced**.
- Pre-push installer: `PrePushInstaller`, `prePushHookContent`, `TestPrePushTemplateMatchesConstant`, `SKIP_MOAI_PREPUSH` env var.
- All other settings.json hook registrations (SessionStart, PostToolUse, Stop, SubagentStop, etc.).

### §A.6 EXTEND list (will touch — pending M1 design decision)

**Primary path (e-1, default — extend SPEC-PRECOMMIT-001's installer)**:
- `internal/cli/hook_install.go` — extend `preCommitHookContent` const to additionally invoke the heavy gate (after the existing fast subset).
- `internal/template/templates/.git_hooks/pre-commit` — mirror the Go const byte-for-byte.
- `internal/cli/wave1_sync_test.go` — extend the byte-identity test to cover the extended pre-commit content (mirror of `TestPrePushTemplateMatchesConstant`).
- `internal/cli/hook_install_test.go` — extend installer tests to assert the heavy-gate invocation token is present after fresh install + after update.
- `.claude/rules/moai/development/coding-standards.md` §Bash Risk-Amplifier Doctrine — extend the destructive-primitive set to include `git commit --no-verify`.
- `internal/hook/pre_tool.go` + `internal/hook/pre_tool_test.go` — FIRM COMMITMENT (F5 mechanical, plan-audit iter-1): the `--no-verify` detection is implemented as a PreToolUse guard with a fast sub-millisecond regex. AC-PGM-004 is mechanically enforceable MUST-PASS via this guard; a `bypassPermissions` agent cannot bypass it because the PreToolUse deny is the sole blocking mechanism under that mode. No longer optional pending-M1-design — the cross-evidence (deny is the ONLY blocking mechanism under `bypassPermissions`) made the mechanical path the only defensible choice.

**Fallback path (e-prime, if M1 rules out (e-1))**:
- New CLI surface exposing `moai gate` as a standalone verb (if not already present — verify at M1).
- Documentation: `README.md`, `.moai/docs/` updates instructing users/CI to invoke `moai gate` before commit.
- No pre-commit hook extension; the pre-commit hook stays as SPEC-PRECOMMIT-001's fast subset.

## §B Known Issues Auto-Injection (filtered to relevant categories)

### §B.1 Cross-platform build tags

Not applicable — the relocation does not introduce `syscall`-package usage. SPEC-PRECOMMIT-001 REQ-PC-019 parity: any new Go installer code uses only portable primitives. Verify with `GOOS=windows GOARCH=amd64 go build ./...`.

### §B.2 Cross-SPEC policy conflict pre-scan

**Conflict surface**: SPEC-PRECOMMIT-001 (completed) and SPEC-GATE-001 (implemented) both touch the commit-quality surface. This SPEC extends SPEC-PRECOMMIT-001's hook body and operationalizes SPEC-GATE-001's REQ-GATE-011 no-bypass intent. No rewrite of either prior SPEC body; both are cross-referenced.

**Pre-scan command** (M1 step):
```bash
grep -rn "Retired\|superseded\|SPEC-PRECOMMIT\|SPEC-GATE" \
  internal/cli/hook_install.go internal/cli/hook_pre_push.go internal/hook/quality/
```

### §B.3 C-HRA-008 / Subagent boundary

The relocated gate runs in git's pre-commit hook (user's shell), NOT in any subagent context. The subagent boundary rule (no `AskUserQuestion` in `internal/hook/` or `internal/cli/`) is preserved by construction. Verify:
```bash
grep -rn 'AskUserQuestion\|mcp__askuser' internal/cli/hook_install.go internal/hook/quality/ \
  | grep -v _test.go | grep -v "// "
# Expected: 0 matches
```

### §B.4 Frontmatter canonical schema

This SPEC's frontmatter uses canonical `created:` / `updated:` / `tags:` (no snake_case aliases). Verified at write time against `.claude/rules/moai/development/spec-frontmatter-schema.md`.

### §B.5 CI 3-tier awareness

spec-lint, golangci-lint, Test (per OS) can each fail independently. Distinguish NEW defects from pre-existing baseline at M4 (record baseline at M1 §C pre-flight).

### §B.6 spec-lint heading convention

spec.md §F Out of Scope uses `### Out of Scope — <topic>` H3 sub-headings (not a bare H2) per the `OutOfScopeRule` lint (`MissingExclusions` finding).

### §B.7 observer.go / capture path resolution

Not directly applicable — the relocated gate does not run as an in-process hook handler. The `resolveQualityProjectDir` helper (gate.go) remains the canonical cwd resolver for in-process gate calls.

### §B.8 Working-tree hygiene

Do NOT modify runtime-managed files (`.moai/harness/usage-log.jsonl`, `.moai/state/*`). Use `git add` with explicit pathspec, never `git add -A`.

### §B.9 Git commit + push (repo-local PR policy)

Per `.claude/rules/moai/workflow/repo-local-pr-policy.md`, this repo has `enforce_admins: true` — **all** tiers (S/M/L) require PR. Self-merge is allowed; CI must pass. The orchestrator's mission authorizes direct commits to the `feat/SPEC-PRETOOL-GATE-MOVE-001` branch (worktree-isolated), but the final landing is via PR (no direct push to `main`).

### §B.10 Untouched paths PRESERVE

See §A.5 above. The worktree branches from `origin/main @ 3e6c92ef7` (SPEC-FALSE-ALLCLEAR-GUARD-001 merge); the ast-grep scanner code is fresh and MUST NOT be touched.

### §B.11 AskUserQuestion prohibited

manager-spec is a subagent. No `AskUserQuestion` calls. If a load-bearing unknown blocks plan authoring, return the canonical `## Missing Inputs` blocker report.

### §B.12 Sync-phase CHANGELOG (manager-docs only)

Not applicable at plan-phase. manager-docs owns CHANGELOG at sync.

## §C Pre-flight Check List (executed at run-phase M1, before any code change)

```bash
# 1. Worktree state (staleness rule — re-read before any commit/push)
git -C /Users/goos/MoAI/moai-adk-go/.claude/worktrees/p1b-gate-move branch --show-current
git -C /Users/goos/MoAI/moai-adk-go/.claude/worktrees/p1b-gate-move rev-parse HEAD

# 2. Cross-platform build feasibility
go build ./...
GOOS=windows GOARCH=amd64 go build ./...

# 3. Existing lint baseline (distinguish NEW vs pre-existing at M4)
golangci-lint run --timeout=2m 2>&1 | tail -5

# 4. PRESERVE-target listing (visually confirm none will be touched)
ls internal/hook/quality/astgrep_gate.go internal/cli/hook_install.go \
   internal/template/templates/.git_hooks/

# 5. Pre-existing SPEC conflicts
grep -rn "Retired\|superseded" internal/cli/hook_install.go internal/hook/quality/ \
  || echo "no conflicts"

# 6. Confirm moai CLI gate surface exists (informs M2 primary vs fallback)
go run ./cmd/moai gate --help 2>&1 | head -5
```

## §D Constraints (DO NOT VIOLATE)

- **PRESERVE enumeration** (see §A.5): ast-grep scanner, frozen-zone guard, `IsGitCommit` regex, `detectToolchain`, pre-push installer, SPEC-PRECOMMIT-001 fast-subset behavior.
- **Forbidden commands**: `git push --no-verify`, `git push --force`, `git reset --hard`, `git rebase -i`, `--amend` on plan-phase commits.
- **Required commands**: Conventional Commits `feat(SPEC-PRETOOL-GATE-MOVE-001): M{n} <subject>`; `🗿 MoAI` trailer on every commit.
- **Template-First cycle**: edit template FIRST, then `make build`, then sync local mirror.
- **Binary constraints (grep-clean gates)**:
  - `grep -rn 'AskUserQuestion' internal/cli/hook_install.go internal/hook/quality/ | grep -v _test.go` → 0 matches (C-HRA-008).
  - `grep -c 'SPEC-PRETOOL-GATE-MOVE\|REQ-PGM' internal/template/templates/.git_hooks/pre-commit` → 0 (template neutrality, REQ-PGM-009).
  - `make build` succeeds; `go test ./internal/cli/... ./internal/hook/...` exits 0.
  - Byte-identity: `preCommitHookContent` const byte-equal to `internal/template/templates/.git_hooks/pre-commit` (REQ-PGM-008).

## §E Self-Verification (skeleton — populated at run-phase per ownership matrix)

Per progress.md §E map (`spec-frontmatter-schema.md` § progress.md Section Map):
- **§E.1 Plan-phase Audit-Ready Signal** — populated by manager-spec at plan-phase (this file's `plan_status:`).
- **§E.2 Run-phase Evidence** — populated by manager-develop (forbidden to this agent).
- **§E.3 Run-phase Audit-Ready Signal** — populated by manager-develop (forbidden to this agent).
- **§E.4 Sync-phase Audit-Ready Signal** — populated by manager-docs; carries `sync_commit_sha` (forbidden to this agent).

The §E.2-§E.4 placeholders are emitted in progress.md per the §E skeleton HARD obligation; they are NOT populated by this agent.

## §F Milestones

Ordered by decision-reversibility (highest-change-likelihood decisions first).

### M1 — Empirical verification of the 3 load-bearing unknowns + PreCommitInstaller structure read

**Why first**: every downstream design decision (extend vs replace, single vs layered, fallback trigger) depends on these empirical findings. Building M2 against an unverified assumption risks a full rewrite. M1 is fixture-based and resolves three load-bearing unknowns the orchestrator flagged + one structural read.

**Sub-tasks**:

- **M1.a — git pre-commit fires under Claude Code's `git commit`** (census-P1-B unknown #1):
  - Fixture: `/tmp/p1b-m1a-<timestamp>` fresh git repo with a `.git/hooks/pre-commit` that writes a sentinel file (`HOOK_FIRED`) to a known path and exits 0.
  - Action: invoke `git commit -m "test"` via the Bash tool (NOT a manual shell — the test must exercise the actual code path Claude Code uses).
  - Pass criterion: `HOOK_FIRED` exists after the attempt.
  - Document in progress.md §E.2 with the verbatim Bash result + fixture setup.

- **M1.b — `--no-verify` bypass behavior** (census-P1-B unknown #2):
  - Same fixture, action `git commit --no-verify -m "test"`.
  - Pass criterion (expected): `HOOK_FIRED` does NOT exist — git suppresses the hook entirely under `--no-verify`.
  - Implication: confirms that the `--no-verify` bypass defense MUST live on the PreToolUse side (M3 coding-standards.md destructive-primitive extension), not in the hook itself.

- **M1.c — reject-surfacing path** (census-P1-B unknown #3):
  - Fixture: `.git/hooks/pre-commit` always exits 1 with a specific marker string on stderr (e.g. `P1B_REJECT_MARKER_<random>`).
  - Action: `git commit -m "test"` via Bash tool.
  - Pass criterion (expected): the Bash tool result includes `P1B_REJECT_MARKER_<random>` in stderr; the marker is visible to Claude.
  - Implication: confirms that git's stderr surfaces to Claude Code, satisfying REQ-PGM-011 without additional plumbing.

- **M1.d — `PreCommitInstaller` structure read** (no edit):
  - Read `internal/cli/hook_install.go` and confirm: `preCommitHookContent` const location + current content; `PreCommitInstaller` type; `installPreCommitHookOptional` helper; `fileHasMoaiMarker` helper; `ErrUserHookExists` sentinel; the byte-identity test in `internal/cli/wave1_sync_test.go`.
  - Output: a structural summary in progress.md §E.2 that anchors the M2 extension shape.

- **M1.e — `moai gate` CLI surface check** (informs M2 primary vs fallback):
  - Run `go run ./cmd/moai gate --help`. If a `gate` subcommand exists and invokes `QualityGate.Run`, the M2 primary path can shell out to `moai gate` from the hook body. If not, M2 primary path uses a direct Go entrypoint or a new thin CLI verb.

**M1 output**: progress.md §E.2 carries the 3 findings + the PreCommitInstaller structure summary + the `moai gate` surface check. The findings determine which M2 sub-path is taken.

**Fallback trigger**: if M1.a finds git pre-commit does NOT fire under Claude Code's Bash invocation, escalate to **§F.M2-fallback** (census option (d) split-gate, or option (e-prime) standalone `moai gate` CLI). REQ-PGM-012 binds this fallback.

### M2 — Relocate the heavy gate to git pre-commit (primary path e-1)

**Decision shape (decided by M1 findings)**:
- **Primary (e-1)**: extend `preCommitHookContent` const to invoke the heavy gate AFTER the existing fast subset (`gofmt -l` + `go vet`). The fast subset stays as the first line; the heavy gate is layered on top.
- **Fallback (e-prime)**: if M1.a fails — standalone `moai gate` CLI the user/CI runs explicitly. The pre-commit hook is not the surface; documentation and (optionally) a wrapper are the deliverables.

**Sub-tasks (primary path)**:
- Edit `internal/cli/hook_install.go` `preCommitHookContent` const: after the existing fast subset, add a guarded invocation of the heavy gate. The invocation MUST:
  - Respect `SKIP_MOAI_PRECOMMIT=1` bypass env var (REQ-PGM-010, SPEC-PRECOMMIT-001 REQ-PC-012 parity).
  - Use `command -v moai >/dev/null 2>&1` to skip gracefully when `moai` is not on PATH (16-language neutrality: non-moai projects pass silently — REQ-PGM-007/009).
  - Preserve the existing `# MoAI-ADK pre-commit hook` marker in the first 3 lines (SPEC-PRECOMMIT-001 REQ-PC-007).
  - Route the heavy gate's failure output to stderr (REQ-PGM-011) and exit non-zero on failure (REQ-PGM-001).
- Edit `internal/template/templates/.git_hooks/pre-commit` to mirror the const byte-for-byte (REQ-PGM-008).
- Run `make build` to recompile the `//go:embed all:templates` binary (CLAUDE.local.md §2 Template-First).

**Files affected (primary)**:
- `internal/cli/hook_install.go`
- `internal/template/templates/.git_hooks/pre-commit`

**Fallback path (e-prime) files**:
- New or extended CLI verb in `internal/cli/` exposing `moai gate` as a standalone command.
- Documentation: `README.md`, `.moai/docs/` updates instructing users/CI to invoke `moai gate` before commit.
- No pre-commit hook extension.

### M3 — Bypass defense + coding-standards extension

**Why separate from M2**: the bypass defense is a doctrinal + PreToolUse-side concern, independent of the relocation surface. It can be developed in parallel with M2's template work (subject to the orchestrator's no-concurrent-write-capable-agents rule).

**Sub-tasks**:
- Extend `.claude/rules/moai/development/coding-standards.md` §Bash Risk-Amplifier Doctrine destructive-primitive set: add `git commit --no-verify` next to the existing `git push --no-verify` entry (doctrine-level documentation of the destructive-primitive set; both bypass git's safety net).
- Extend `internal/hook/pre_tool.go` to detect `git commit --no-verify` (standalone arg OR substring of a compound Bash command) via a fast sub-millisecond regex and emit deny/ask via the existing PreToolUse hook path. **FIRM COMMITMENT (F5 mechanical, plan-audit iter-1) — NOT optional**: under `defaultMode: bypassPermissions` the PreToolUse deny is the sole blocking mechanism, so the detection MUST live on the PreToolUse path; a doctrine-only documentation grep would leave a `bypassPermissions` agent free to issue `--no-verify` and bypass the relocated gate. The detection MUST be a fast regex check (sub-millisecond) — the heavy gate is no longer on PreToolUse, so the 5s budget is no longer a constraint for this detection either.
- Add/extend `internal/hook/pre_tool_test.go` with a mechanical test fixture: stage a `git commit --no-verify -m test` PreToolUse payload → assert the guard emits deny. This is the AC-PGM-004 mechanical test (NOT a documentation grep).

**Files affected**:
- `.claude/rules/moai/development/coding-standards.md`
- `internal/hook/pre_tool.go` (FIRM — F5 mechanical)
- `internal/hook/pre_tool_test.go` (FIRM — F5 mechanical)

### M4 — Tests (reachability, language-neutrality, template-neutrality, bypass defense, byte-identity)

**Sub-tasks** (each maps to one or more ACs in acceptance.md §A):

- **Reachability test** (AC-PGM-001, AC-PGM-002): stage a Go file with a `go vet` failure → run the relocated gate → assert commit rejected; mirror with a clean file → assert commit passes.
- **Budget-independence test** (AC-PGM-003, END-TO-END fixture per F1 amendment): exercise Claude Code's Bash tool end-to-end — staged bad commit + wall-clock observation t>5s (via `/usr/bin/time -p`) + commit rejected by the relocated git pre-commit gate. NOT satisfiable by a standalone `gate.Run()` timing test (F1 amendment — the budget-independence claim must ground at the integration boundary, not a synthetic sub-component).
- **Bypass-defense test** (AC-PGM-004, mechanical MUST-PASS per F5 amendment): PreToolUse guard at `internal/hook/pre_tool.go` mechanically denies `git commit --no-verify` via fast sub-millisecond regex — the test exercises the actual guard, NOT a documentation grep.
- **Language-neutrality test** (AC-PGM-005): parameterized test across ≥3 non-Go languages (Python + `ruff check .` failing, Rust + `cargo clippy -D warnings` failing, Node.js + `eslint .` failing). Each test stages a language-specific defect file and asserts the relocated gate catches it.
- **Fast-check preservation tests** (AC-PGM-006, AC-PGM-007): regression tests that ast-grep scanner and frozen-zone guard still fire on PreToolUse (unchanged by this SPEC — baseline captured at M1 §C pre-flight).
- **Template-neutrality grep test** (AC-PGM-008): grep the distributed template for internal SPEC IDs / REQ tokens / commit SHAs — expect 0 matches.
- **Byte-identity test** (AC-PGM-010): extend the `TestPrePushTemplateMatchesConstant` mirror to cover the extended pre-commit content; assert `preCommitHookContent` const == template file byte-for-byte.
- **Install-path tests** (AC-PGM-011, AC-PGM-012): extend `internal/cli/hook_install_test.go` — (a) fresh-repo install carries the heavy-gate invocation token; (b) update overwrites marker-carrying hooks and preserves foreign hooks (`ErrUserHookExists`).
- **SKIP_MOAI_PRECOMMIT bypass test** (AC-PGM-013): env-var set → hook prints bypass notice and exits 0 without running the heavy gate.
- **Error-surfacing test** (AC-PGM-014): use the M1.c fixture pattern; assert the rejection marker reaches the Bash tool result.
- **No-PreToolUse-regression test** (AC-PGM-015): assert `IsGitCommit` still detects `git commit`; `detectToolchain` still resolves all 16 languages; ast-grep + frozen-zone still fire.

**Files affected**:
- `internal/cli/hook_install_test.go` (extend)
- `internal/cli/wave1_sync_test.go` (extend byte-identity test)
- `internal/hook/quality/gate_test.go` (extend or add reachability/language-neutrality tests)
- New: `internal/cli/precommit_relocation_test.go` (or similar) for the integration tests

### M5 — Sync (3-phase close)

- manager-docs sync-phase: CHANGELOG entry, README/docs-site update if the pre-commit hook behavior is user-visible, frontmatter transition to `completed`.
- Single sync commit carries the `implemented → completed` transition.
- `progress.md` §E.4 `sync_commit_sha` populated by manager-docs (with D3 backfill pattern if needed — the SHA cannot be known within its own commit).
- Per CLAUDE.local.md §23 + `repo-local-pr-policy.md`: landing is via PR (self-merge allowed; CI must pass).

## §G Anti-Patterns

- **AP-PGM-001 — Touching PRESERVE surfaces**: editing `astgrep_gate.go`, the frozen-zone guard, `IsGitCommit`, `detectToolchain`, or the pre-push installer violates scope discipline and risks regressing SPEC-FALSE-ALLCLEAR-GUARD-001 or SPEC-PRECOMMIT-001.
- **AP-PGM-002 — Hard-coding Go in the distributed template**: writing `if [[ -f go.mod ]]; then ...` in the template without a `command -v go` guard regresses 16-language neutrality for projects without a Go toolchain.
- **AP-PGM-003 — Internal-content leak into template**: embedding `SPEC-PRETOOL-GATE-MOVE-001` or `REQ-PGM-NNN` tokens in `internal/template/templates/.git_hooks/pre-commit` violates §25 isolation and fails AC-PGM-008 grep.
- **AP-PGM-004 — Skipping `make build`**: editing the template without recompiling the binary means `//go:embed all:templates` serves stale content; the byte-identity test (AC-PGM-010) catches this only if the test runs post-build.
- **AP-PGM-005 — Treating `--no-verify` as solvable in the hook**: git does not pass `--no-verify` to the hook; the hook is simply not invoked. The bypass defense MUST live in PreToolUse (coding-standards.md destructive-primitive extension).
- **AP-PGM-006 — Silent-fail the hook on missing `moai`**: if the relocated gate invocation fails silently when `moai` is not on PATH, the gate becomes a no-op for users who installed via a non-standard path. Use `command -v moai` guard + stderr notice.
- **AP-PGM-007 — Regressing SPEC-PRECOMMIT-001 byte-identity**: extending the const without extending the test leaves the byte-identity invariant unverified post-change.
- **AP-PGM-008 — Treating M1 findings as optional**: REQ-PGM-012 binds the fallback trigger. Skipping M1 and assuming git pre-commit fires under Claude Code risks building the entire relocation on an unverified assumption.

## §H Cross-References

- spec.md: `.moai/specs/SPEC-PRETOOL-GATE-MOVE-001/spec.md`
- acceptance.md: `.moai/specs/SPEC-PRETOOL-GATE-MOVE-001/acceptance.md`
- progress.md: `.moai/specs/SPEC-PRETOOL-GATE-MOVE-001/progress.md`
- Census report: `.moai/reports/census-2026-07-27-handoff.md` §C-2 (line 107), §7 Priority 1 P1-B (line 619)
- Prior SPECs: SPEC-GATE-001 (implemented, original gate), SPEC-PRECOMMIT-001 (completed, installer being extended), SPEC-FALSE-ALLCLEAR-GUARD-001 (PR #1183, just-tuned ast-grep scanner, worktree base)
- CLAUDE.local.md §2 (Template-First), §15 (16-language neutrality), §23 (PR-mandatory policy), §25 (internal-content isolation)
- `internal/cli/hook_install.go` — `PreCommitInstaller` / `preCommitHookContent` / `fileHasMoaiMarker` / `ErrUserHookExists` / `installPreCommitHookOptional`
- `internal/cli/wave1_sync_test.go` — `TestPrePushTemplateMatchesConstant` (byte-identity gate to mirror)
- `.claude/rules/moai/development/coding-standards.md` §Bash Risk-Amplifier Doctrine
- `.claude/rules/moai/development/spec-frontmatter-schema.md` § Status Transition Ownership Matrix (close-subject full-ID mandate)
