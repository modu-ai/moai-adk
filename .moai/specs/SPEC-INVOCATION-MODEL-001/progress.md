# Progress — SPEC-INVOCATION-MODEL-001

> Lifecycle: plan → run → sync. This file tracks phase evidence. Plan-phase populates §E.1 only; §E.2/§E.3 are owned by manager-develop (run-phase); §E.4 by manager-docs (sync-phase).

---

## §E.1 Plan-phase Audit-Ready Signal

- **Status**: draft (plan-phase artifacts created 2026-07-01).
- **SPEC ID self-check**: `SPEC-INVOCATION-MODEL-001` decomposition → SPEC ✓ | INVOCATION ✓ | MODEL ✓ | 001 ✓ → PASS (canonical regex `^SPEC(-[A-Z][A-Z0-9]*)+-\d{3}$`).
- **Artifacts**: spec.md, plan.md, acceptance.md, progress.md (4-file Tier M set).
- **REQ coverage**: REQ-IM-001..014 (5 doctrine + 6 feedback + 3 cross-cutting) → AC-IM-001..022 (AC-IM-021 per-command citation anchor + AC-IM-022 run-phase official-docs verification added per plan-auditor D1).
- **Tier**: M (frontmatter `tier: M` — sets plan-auditor 0.80 threshold; aligns frontmatter with plan/progress).
- **plan-auditor**: iter-1 FAIL 0.81 (no blocking defects; all additive). D1-D7 corrections applied — D5 tier, D1 per-command citations + `/loop` framing correction, D2 orchestrator-direct reframe of REQ-IM-014, D3 2-guaranteed + `go version` best-effort diagnostics, D4 definite loader wiring (`feedbackFileWrapper`/`loadFeedbackSection`/`Loader.Load()`), D6 GEARS Event-driven relabels (REQ-IM-005/009), D7 integration/file-load test for AC-IM-012.
- **Deliverables**: (1) invocation-model alignment doctrine rule; (2) /moai feedback enhancement (diagnostic attach, dup detection, gh-fallback, target-repo config-ization).
- **Design decisions RESOLVED** (all 3 fixed by orchestrator, plan.md §H): (1) rule path = `.claude/rules/moai/workflow/native-invocation-model.md`; (2) config = `.moai/config/sections/feedback.yaml` key `feedback.repository` default `modu-ai/moai-adk`; (3) diagnostic = HYBRID — 2 guaranteed items (`moai version`, `uname`/OS) + `go version` best-effort (tool build-provenance) + optional orchestrator-mediated error context (best-effort). No open questions remain.
- **Milestones**: M1 doctrine rule → M2 config-ization (TDD) → M3 feedback workflow → M4 build+verify.
- **Frontmatter**: 12 canonical fields present; validated against spec-frontmatter-schema.md.
- **Out of Scope**: 5 `### Out of Scope —` sub-headings present in spec.md §E (Axis-A reuse, legacy retirement, runtime enforcement, config schema, feedback UX).

## §E.2 Run-phase Evidence

### AC-IM-022 — Official-docs classification verification (RECORDED BEFORE the M1 doctrine matrix commit)

Per verification-claim-integrity §1.1, the 9-command PROGRAMMATIC/HUMAN-ONLY classification matrix is NOT asserted from memory. It was verified at run-phase by fetching the official Claude Code docs.

- **Command run** (WebFetch unavailable in the isolated run-phase agent context → curl fallback of the same official pages):
  - `curl -sL https://code.claude.com/docs/en/commands` → HTTP 200, 1259176 bytes (the commands reference — `[Skill]`/`[Workflow]` marker source)
  - `curl -sL https://code.claude.com/docs/en/skills` → HTTP 200 (the `skills.md` — prompt-based defining quote + `Skill`-tool bridge statement)
- **Observed evidence** (verbatim extracts):
  - commands reference definition entries: `/code-review [low|medium|high|xhigh|max|ultra] [--fix] [--comment] [target]` **Skill** ; `/simplify [target]` **Skill** ; `/loop [interval] [prompt]` **Skill** ("Omit the interval and Claude self-paces between iterations") ; `/deep-research <question>` **Workflow** ; `/goal [condition|clear]` (no marker) ; `/clear [name]` (no marker) ; `/compact [instructions]` (no marker)
  - `skills.md`: "Unlike most built-in commands, which execute fixed logic directly, bundled skills are prompt-based: they give Claude detailed instructions and let it orchestrate the work using its tools."
  - `skills.md`: "A few built-in commands are also available through the Skill tool, including /init, /review, and /security-review. Other built-in commands such as /compact are not."
  - `skills.md`: "Add disable-model-invocation: true to prevent Claude from triggering it automatically." + "available in every session unless disabled with the disableBundledSkills setting" + "Disable all skills by denying the Skill tool in /permissions".
- **Verified classification** (6 PROGRAMMATIC / 3 HUMAN-ONLY):
  - PROGRAMMATIC bundled skill/workflow: `/code-review`, `/simplify`, `/loop`, `/deep-research`
  - PROGRAMMATIC built-in-exposed-via-Skill-tool: `/security-review`, `/review`
  - HUMAN-ONLY built-in (no Skill-tool bridge): `/goal`, `/clear`, `/compact`
- **DIVERGENCE recorded** (LIVE observation overrides plan.md §D1 provisional table): plan.md provisionally tagged `/security-review` and `/review` as HUMAN-ONLY. The official `skills.md` shows both ARE available through the Skill tool → they are orchestrator-invocable → reclassified PROGRAMMATIC (built-in-exposed sub-case). The `/moai review` per-subcommand note in the doctrine reflects this: its justification rests on Axis A + broader-orchestration composition, NOT on a HUMAN-ONLY premise. This divergence is documented in the doctrine's "Classification-divergence note".
- **`/loop` framing correction**: commands reference confirms native `/loop` is a bundled `[Skill]` that self-paces when the interval is omitted — NOT merely a "fixed time-interval scheduler" as goal-directive.md frames it. The doctrine carries the acknowledge+correct cross-reference (goal-directive.md NOT edited).

### Per-milestone commit SHAs (worktree branch `worktree-agent-abae346a5e201f249`)

- M1 (doctrine rule + mirror + progress §E.2 evidence + spec.md draft→in-progress): `846a0b463`
- M2 (feedback config-ization TDD): `769acbc4f`
- M3 (feedback workflow enhancement): `7b974e6c9`
- M4 (build + parity + verification + §E.3 signal): _<this commit>_

> Run-phase environment note: this run-phase executed in a runtime-materialized L1 isolation worktree (`Agent(isolation: "worktree")`), NOT the main checkout that Section A of the delegation assumed. Commits land on the worktree branch `worktree-agent-abae346a5e201f249`; the orchestrator reconciles them to main (the untracked SPEC plan-dir in main requires the mv-backup merge-unblock procedure per the worktree-untracked-plandir hazard). WebFetch was unavailable in this agent context; the official-docs verification above used the curl fallback of the same pages.

## §E.3 Run-phase Audit-Ready Signal

### AC binary matrix — 22/22 PASS

| AC | Status | Verification |
|----|--------|--------------|
| AC-IM-001 | PASS | grep doctrine: PROGRAMMATIC + HUMAN-ONLY + `Skill()` + `[Skill]`/`[Workflow]` heuristic present |
| AC-IM-002 | PASS | grep doctrine: verbatim skills.md "Unlike most built-in commands…bundled skills are prompt-based" quote present |
| AC-IM-003 | PASS | grep doctrine: all 9 native commands present, each tagged |
| AC-IM-004 | PASS | grep doctrine: Axis A + Axis B thesis present |
| AC-IM-005 | PASS | grep doctrine: feedback per-subcommand note + review --security example present |
| AC-IM-006 | PASS | grep doctrine: disable-model-invocation + disableBundledSkills caveat present |
| AC-IM-007 | PASS | grep feedback.md: `moai version` + `uname` guaranteed; `go version` best-effort present |
| AC-IM-008 | PASS | grep feedback.md: "tool-diagnostic information only" + "MUST NOT attach arbitrary user file contents" |
| AC-IM-009 | PASS | grep feedback.md: `gh issue list --repo <resolved-target> --search` + candidate-report |
| AC-IM-010 | PASS | grep feedback.md: `gh auth status` + rate-limit + "save the drafted issue body locally" fallback |
| AC-IM-011 | PASS | `go test ./internal/config/ -run TestFeedbackRepositoryDefault` → PASS (default resolves modu-ai/moai-adk) |
| AC-IM-012 | PASS | `go test ./internal/config/ -run TestFeedbackRepositoryOverride` → PASS (integration/file-load, override resolves) |
| AC-IM-013 | PASS | grep feedback.md: `gh issue create --repo <resolved-target>`; no bare hardcode as sole target |
| AC-IM-014 | PASS | `diff` feedback.md local↔template → IDENTICAL |
| AC-IM-015 | PASS | `diff` native-invocation-model.md local↔template → IDENTICAL |
| AC-IM-016 | PASS | `diff` feedback.yaml local↔template → IDENTICAL |
| AC-IM-017 | PASS | `make build` exit 0; `go build ./...` exit 0 (binary compiled with embedded new templates) |
| AC-IM-018 | PASS | `go test ./internal/template/ -run TestTemplateNoInternalContentLeak` → PASS |
| AC-IM-019 | PASS | `grep -c AskUserQuestion feedback.md` = 5 (unchanged baseline; 0 new for dedupe) |
| AC-IM-020 | PASS | `moai spec lint .../spec.md` → "✓ No findings — all SPEC documents are valid" |
| AC-IM-021 | PASS | grep doctrine: each of 9 commands carries a citation anchor; `/loop` framing correction + goal-directive.md cross-ref present |
| AC-IM-022 | PASS | Official-docs verification recorded in §E.2 above (curl commands.md/skills.md, classification + divergence recorded) BEFORE the M1 matrix commit |

### Build / coverage / lint

- **Cross-platform build**: `go build ./...` exit 0 + `GOOS=windows GOARCH=amd64 go build ./...` exit 0.
- **Coverage** (feedback resolver path, `go test -coverprofile ./internal/config/`): `FeedbackRepository` 100.0%, `NewDefaultFeedbackConfig` 100.0%, `loadFeedbackSection` 100.0% (≥ 85% threshold met).
- **Lint**: `golangci-lint run ./internal/config/... ./internal/template/...` → 0 issues (no NEW findings).
- **spec lint**: `moai spec lint spec.md` → clean (frontmatter 12-field + OutOfScopeRule pass).
- **Full suite** (`go test ./...`): touched packages (`internal/config`, `internal/template`) GREEN. NOTE: 2 pre-existing environment-flaky failures in `internal/hook/wrapper_test.go` (`TestHookWrapper_MoaiBinaryFallback`, `TestHookWrapper_ValidJSON`) — "signal: killed" on subprocess exec under parallel-sandbox load; `internal/hook` was NOT touched by this SPEC (git diff-verified), both PASS in isolation, and the immediate full-suite re-run returned exit 0. Environment artifact, not a regression from this SPEC.

### Residual risk

- Commits M1-M4 land on the L1 isolation worktree branch `worktree-agent-abae346a5e201f249`, NOT on main. Orchestrator reconciliation to main is required (the 4 SPEC plan-dir artifacts are untracked in main → merge needs the mv-backup untracked-plandir unblock procedure). manager-develop could not push to main from the isolated worktree.
- The `internal/hook` full-suite flake could recur under load; it is orthogonal to this SPEC's scope.

## §E.4 Sync-phase Audit-Ready Signal

### Sync-phase audit evidence

- **CHANGELOG.md entry created**: `[SPEC-INVOCATION-MODEL-001]` added to [Unreleased]/Added section on 2026-07-01 (orchestrator-direct append, pre-verification: grep -c = 0 before append). AC count verified = 22. Deliverable file paths verified via ls: (1) `.claude/rules/moai/workflow/native-invocation-model.md` (13.3 KB), (2) `.claude/skills/moai/workflows/feedback.md` (8.4 KB), (3) `.moai/config/sections/feedback.yaml` (356 bytes).
- **README.md assessment**: No changes required. SPEC is infrastructure/doctrine (not a user-facing feature addition). High-level README describes architecture; detailed deliverables are documented in the deliverable files themselves and the CHANGELOG entry.
- **spec.md frontmatter transition**: status `in-progress` → `completed` (applied 2026-07-01 on spec.md line 5).
- **Template mirror parity**: all 3 deliverables byte-identical between local and template (verified during run-phase M4).
- **Build + lint baseline**: `go build ./...` exit 0; `golangci-lint run ./internal/config/ ./internal/template/...` → 0 new issues; `moai spec lint spec.md` → clean.
- **SPEC-ID format**: SPEC-INVOCATION-MODEL-001 verified correct per canonical regex at plan-phase §E.1.

### Divergence recorded (classification refinement logged)

During run-phase official-docs verification (AC-IM-022 in §E.2 above), two native commands were reclassified from the plan.md provisional HUMAN-ONLY to PROGRAMMATIC:
- `/security-review` — official `skills.md` confirms it is available via the Skill tool (built-in-exposed-via-Skill-tool case)
- `/review` — same; official `skills.md` confirms Skill-tool availability

The native-invocation-model.md doctrine carries a "Classification-divergence note" cross-reference to this finding (documented in §E.2 line 39 above). The `/moai review` per-subcommand note in the doctrine reflects this divergence (its PROGRAMMATIC justification rests on Axis A orchestration composition, not a HUMAN-ONLY premise). No spec.md body edit was performed; divergence is recorded here for audit trail.

### Three-phase close metadata

- **plan-phase**: draft (plan.md recorded) + audit-ready signal (§E.1 above)
- **run-phase**: M1-M4 commits on L1 worktree branch + orchestrator reconciliation pending; audit-ready signal (§E.3 above, 22/22 AC PASS, build GREEN, lint clean)
- **sync-phase**: CHANGELOG + README + spec.md frontmatter + progress.md §E.4 (this section) + git commit to main + push (orchestrator-direct)

sync_commit_sha: 2614eab148abadaac1a81bd90aefc9dd9b3fba8f

> **sync_commit_sha backfill procedure** (per lifecycle-redesign-001 REQ-LR-009): The sync_commit_sha field will be populated with the commit SHA of the sync-phase final commit AFTER the commit is created and its SHA is returned by `git rev-parse HEAD` (per the canonical 3-phase-close atomic commit contract, the sync commit carries the status transition + frontmatter + CHANGELOG + progress.md updates atomically; sync_commit_sha is backfilled as a follow-up precision commit within the sync phase, not a separate Mx commit). This allows the orchestrator to record the exact SHA that closed the SPEC.
