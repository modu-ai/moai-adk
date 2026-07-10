# progress.md — SPEC-HOOK-OFFICIAL-COMPLIANCE-001

> Plan-phase skeleton. Run-phase evidence (§E.2/§E.3) is populated by manager-develop; sync-phase close (§E.4) by manager-docs. This file carries ONLY the §E.1 plan-phase signal at creation; §E.2-§E.4 are placeholder headings (no populated evidence, commit SHAs, or audit-ready YAML at plan-phase).

---

## §A. Current Phase

Plan-phase (draft). Artifacts: spec.md + plan.md + acceptance.md + progress.md (this file).

---

## §B. Artifact Status

| Artifact | Status | Notes |
|----------|--------|-------|
| spec.md | draft | 21 REQs (HOC-001..021), 8 milestones M1-M8, Out of Scope §G |
| plan.md | draft | 8 milestones, dedup verdicts §D, Go inspection requirements |
| acceptance.md | draft | 36 ACs (AC-HOC-001..036), severity classification §D.1 |
| progress.md | draft | this file — §E skeleton only |

---

## §C. Milestone Progress

| Milestone | Title | Priority | Status |
|-----------|-------|----------|--------|
| M1 | Blocking gate JSON contract + PreToolUse exit-code (HIGH) | High | done (7/7 AC, orchestrator-independent verify) |
| M2 | Doctrine refresh (8-point single pass) | Medium | done (AC-008..016, 9/9 AC) commit pending |
| M3 | Async observation taps (UserPromptSubmit/Stop/SubagentStop) | Medium | done (AC-017/018) landed 7e641d959 |
| M4 | Timeout headroom (TeammateIdle/TaskCompleted/PreCompact) | Medium | done (AC-019/020) landed 7e641d959 |
| M5 | Matcher resolution + Go verification (FileChanged/ConfigChange) | Medium | done (AC-021/022) — Go inspection: no settings.json/wrapper change needed |
| M6 | Fail-open semantics correction (WorktreeCreate/PermissionRequest) | Medium | done (AC-023/024) |
| M7 | Input hardening (MOAI_HOOK_STDERR_LOG + gateguard escaping) | Low | not-started |
| M8 | Coverage holes + defects (MultiEdit/csharp/exec-form/compact) | Low | M8a done (AC-027 MultiEdit) landed 7e641d959; M8b-d (AC-028..030) remain |

---

## §D. Blocker Log

_(empty — plan-phase)_

---

## §E.1 Plan-phase Audit-Ready Signal

```yaml
plan_status: audit-ready
plan_complete_at: 2026-07-10
plan_artifacts: [spec.md, plan.md, acceptance.md, progress.md]
tier: L
req_count: 21
ac_count: 36
dedup_specs_cited:
  - SPEC-HOOK-SESSIONSTART-PROBE-001 (probe DONE — Rec #6 partial)
  - SPEC-HOOK-FACTFORCE-ADVISORY-001 (exit-0 DONE — Rec #7 partial)
  - SPEC-DIVECC-HOOK-FAILURE-MODE-AUDIT-001 (doctrine file exists — Rec #2 is an update)
primary_source: .moai/reports/hooks-improvement-plan-20260710.html
```

---

## §E.2 Run-phase Evidence

### M1 (HIGH) — Blocking gate JSON contract + PreToolUse exit-code — DONE

**§E.2 Go-inspection finding (binding per spec.md §E.2):** `internal/hook/pre_tool.go` deny path returns `NewDenyOutput(reason)` (line 380 + 7 more sites) → `*HookOutput` with `HookSpecificOutput.PermissionDecision="deny"` + `hookEventName:"PreToolUse"` (`types.go:448-458`), ExitCode zero-value (0). `pre_tool.go` never sets `ExitCode=2`; CLI `internal/cli/hook.go:228/359` `if output.ExitCode==2 { os.Exit(2) }` exists but is not reached by the PreToolUse deny path. **Conclusion: path (b) — exit 0 + JSON `permissionDecision:"deny"`.** Therefore `exit $?` (AC-003) evaluates to exit 0 today (behavior-preserving) and is future-proof if the handler ever emits exit 2.

**E1 — AC PASS/FAIL matrix (AC-HOC-001..007), orchestrator-independent verification:**

| AC | Status | Evidence (command → observed) |
|----|--------|-------------------------------|
| AC-001 | PASS | `sed -n '40,50p' team-ac-verify.sh \| grep -cE 'continue.*false\|stopReason'` → 1; `grep -c 'decision.*block'` in window → 0 (live + template) |
| AC-002 | PASS | sync-phase-quality-gate.sh `hookSpecificOutput` count 3 + `"hookEventName":"Stop"` present |
| AC-003 | PASS | handle-pre-tool.sh `exit $?` count 3 (live + template) |
| AC-004 | PASS | handle-pre-tool.sh moai-invocation lines with `2>>...MOAI_HOOK_STDERR_LOG` → 0 |
| AC-005 | PASS | handle-session-start.sh probe `hookEventName:"SessionStart"` present (live + template, 2 matches each) |
| AC-006 | PASS | agent-common-protocol.md byte-identical template↔live (`diff -q` IDENTICAL); clause (b) + table row carry `continue:false`+`stopReason` |
| AC-007 | PASS | `go test ./internal/template/ -run 'TestRuleTemplateMirror\|TestHookOfficialCompliance'` → `ok 0.459s` |

**E2 — Cross-platform build:** `go build ./...` → exit 0.
**E3 — Coverage:** N/A (no Go source changed; shell + doctrine edits only; new characterization test added).
**E4 — Subagent boundary:** `grep -rn 'AskUserQuestion' internal/hook/ internal/cli/` matches are pre-existing docs/comments only — 0 NEW introductions.
**E5 — Lint:** `golangci-lint run ./internal/template/` → 0 issues. (initial stringsseq diagnostics resolved: `for ln := range strings.SplitSeq` single-variable form at 3 sites; line-157 index-needing site stays `strings.Split`.)
**E6 — Push state:** NOT pushed (local working-tree). M1 + SPEC artifacts staged for a single pathspec commit.
**E7 — Deferred:** none.

### M3 (MEDIUM) — Async observation taps — DONE (backfill, landed 7e641d959)

Orchestrator-verified prior to this delegation (per task §A): settings.json.tmpl harness-observe entries for UserPromptSubmit, Stop, SubagentStop now carry `"async": true`, and InstructionsLoaded carries `"async": true`. AC-017 (≥3 async taps) + AC-018 (InstructionsLoaded async) satisfied. No re-edit by this delegation (M3 in HEAD `7e641d959`).

### M4 (MEDIUM) — Timeout headroom — DONE (backfill, landed 7e641d959)

Orchestrator-verified prior to this delegation (per task §A): settings.json.tmpl timeouts raised — TeammateIdle 60s, TaskCompleted 60s, PreCompact 30s (all >5s baseline). AC-019 (TeammateIdle/TaskCompleted >5) + AC-020 (PreCompact >5) satisfied. No re-edit (M4 in HEAD `7e641d959`).

### M8a (LOW) — MultiEdit matcher — DONE (backfill, landed 7e641d959)

Orchestrator-verified prior to this delegation (per task §A): settings.json.tmpl PostToolUse status-transition matcher = `Write|Edit|MultiEdit` (line 104). AC-027 satisfied. M8b-d (csharp glob / agent-hook exec form / compact naming — AC-028..030) remain out of this delegation's scope.

### M2 (MEDIUM) — Doctrine refresh (8-point single pass) — DONE (commit 773fddb1a, origin main)

8 stale points in hooks-system.md + hook-independence.md row (g) corrected in a single pass. Template-First (live ↔ template byte-identical). Edit set (14 string replacements across the 2 doctrine files × 2 mirrors):

- AC-008 (GAP-LC-01): PreCompact Can Block `No`→`Yes`
- AC-009 (GAP-STT-03): SubagentStop stdout +`decision:block+reason` / `additionalContext`
- AC-010 (GAP-OBS-04): Notification matcher 4→8 values (+elicitation_complete/elicitation_response/agent_needs_input/agent_completed); SessionEnd 4→6 (+bypass_permissions_disabled/other); StopFailure 4→10 (+overloaded/oauth_org_not_allowed/invalid_request/model_not_found/server_error/unknown)
- AC-011 (GAP-CEW-03): FileChanged matcher = literal filenames (NOT regex/glob), split on `|` into literals
- AC-012 (GAP-PE-02): UserPromptSubmit stdout +`decision:{block,reason}`/`sessionTitle`/`suppressOriginalPrompt`; PermissionRequest +`decision.behavior`/`updatedInput`/`updatedPermissions`
- AC-013 (GAP-PE-03): Elicitation/ElicitationResult handler-type = command+http+mcp_tool only (prompt/agent silently discarded)
- AC-014 (GAP-GD-06): exit-2 description rewritten from "universal" to per-event enum (honoring vs ignoring event lists)
- AC-015 (GAP-TOOL-04): PostToolUse async constraint — async PostToolUse can only deliver `additionalContext`
- AC-016 (GAP-GD-03): hook-independence.md row (g) sync-phase-quality-gate timeout 10s→60s (Stop)

**E1 — AC PASS:** all 9 AC greps (AC-008..016) PASS against `core/hooks-system.md` + `development/hook-independence.md` (canonical paths).
**E2 — Build:** `go build ./...` exit 0; `GOOS=windows GOARCH=amd64 go build ./...` exit 0.
**E5 — Lint:** `golangci-lint run --timeout=3m` 0 issues (clean baseline preserved).
**E5b — Mirror parity:** `TestRuleTemplateMirrorDrift` PASS (hooks-system.md mirror byte-identical).
**E5c — Neutrality:** `TestInternalContentLeak` PASS (no SPEC IDs/REQ tokens/SHAs leaked into templates).
**E6 — Push:** `773fddb1a` pushed to `origin main` (b3f66a69b..773fddb1a); post-push `git rev-list --left-right origin/main...HEAD` = `0 0`.
**Gaps:** acceptance.md AC greps reference `development/hooks-system.md` (stale path — file lives at `core/hooks-system.md`, cross-referenced by hook-independence.md §8). Substantive content correct; flagged for manager-spec path fix in a follow-up. StopFailure 10 values + Notification 8 values sourced from the audit report's official-doc enumeration (code.claude.com/docs/en/hooks, 2026-07-10 fetch); the moai-adk Go `stop_failure.go` only special-cases 4 (rate_limit/authentication_failed/billing_error/max_output_tokens) — the other 6 fall through to a default systemMessage path.

### M5 (MEDIUM) — Matcher resolution + Go verification — DONE (no file edit; inspection-only milestone)

**§E.2 Go inspection (binding per spec.md §E.2) — AC-021 (FileChanged matcher resolution):**

`internal/hook/file_changed.go` `Handle()` (lines 84-122) is a **pure downstream consumer** of `input.FilePath`. It performs NO matcher matching — it reads `input.ChangeType` and `filepath.Ext(input.FilePath)` against `supportedExtensions` and runs the MX scan. Matcher evaluation is owned entirely by the **Claude Code runtime** (the platform), which invokes the wrapper only after the matcher filter passes. `internal/hook/protocol.go` confirms the handler receives the already-filtered payload (no matcher field in `HookInput`). Per the official doc (code.claude.com/docs/en/hooks, audit GAP-GD-04), FileChanged matchers are **literal filenames** — the value is **split on `|`** and each segment is registered as a literal filename. Therefore the current settings.json.tmpl FileChanged block `"matcher": ".env|.envrc|.gitignore"` is **CORRECT** (splits into 3 literal filenames: .env, .envrc, .gitignore) — the "where literal-only, register individual entries" condition does NOT hold (the runtime supports pipe-split literals). **No settings.json change.** M2's AC-011 doctrine note documents the literal-with-pipe-split semantics — the doctrine and shipping runtime are reconciled.

**§E.2 Go inspection (binding per spec.md §E.2) — AC-022 (ConfigChange block channel):**

`internal/hook/config_change.go` `Handle()` (lines 62-90) **always returns `&HookOutput{}, nil`** (empty output = continue, exit 0). Line 87-88 comment: "Empty output = continue". The handler **NEVER blocks** — there is no block path. All validation + reload runs in a background goroutine (`runReload`, 5s deadline); failures are logged and swallowed (`slog.Warn` + early return), never propagated to the response. The handler emits **neither** stdout-JSON `decision:block+reason` **nor** exit 2 + stderr. Therefore the AC-022 condition "where exit 2 + stderr, make stderr event-conditional" does **NOT hold** → **no `handle-config-change.sh.tmpl` change needed.** The wrapper's existing `2>>"$MOAI_HOOK_STDERR_LOG"` on the 3 exec branches is harmless (captures incidental stderr to the log; no block-reason needs to reach Claude Code via stderr).

**E7 — Blocker report:** NONE. Go inspection revealed NO handler-side defect. `file_changed.go`'s downstream-consumer design is correct (matcher semantics are platform-owned). `config_change.go`'s never-block design is intentional (observer/reloader, not a gate). No Go fix required (spec.md §G out-of-scope respected).

**E1 — AC PASS:** AC-021 satisfied via doctrine documentation (AC-011 literal note in M2) + Go inspection (file_changed.go downstream-consumer confirmation). AC-022 satisfied via Go inspection (config_change.go never-blocks confirmation). Both ACs' verification blocks in acceptance.md call for "Go inspection evidence recorded in run-phase E7 report" — recorded above.
**E2 — Build:** N/A (no file change in M5; inspection only).
**E6 — Commits:** M5 evidence recorded in this progress.md update (no separate code commit needed — M5 is inspection-only; the M2 doctrine note already reconciles AC-021 with the shipping runtime).

### M6 (MEDIUM) — Fail-open semantics correction (WorktreeCreate/PermissionRequest) — DONE

Header-comment guard added to `handle-worktree-create.sh` + `handle-permission-request.sh` (.tmpl + .sh mirrors) stating the wrapper MUST NOT be registered unless moai is guaranteed resolvable, and the misleading "Claude Code handles missing hooks gracefully" comment corrected to name the fail-open hazard for active-creator (WorktreeCreate: empty stdout + exit 0 ABORTS creation) and security-gate (PermissionRequest: exit 0 = ALLOW neutralizes the permission gate) events. SessionStart probe (`handle-session-start.sh`) gained a PermissionRequest-priority dimension flagging the security-negative silent-allow over the observer-event degradation.

**E1 — AC PASS/FAIL matrix:**

| AC | Status | Evidence (command → observed) |
|----|--------|-------------------------------|
| AC-023 | PASS | `head -20 handle-worktree-create.sh.tmpl \| grep -iE 'MUST NOT.*register\|guaranteed.*resolvable'` → 2 matches ("MUST NOT be registered", "guaranteed resolvable"); same for `handle-permission-request.sh.tmpl` |
| AC-024 | PASS | `grep -iE 'PermissionRequest\|permission.*allow\|security.*negative' .claude/hooks/moai/handle-session-start.sh` → 3 matches (PermissionRequest-priority dimension comment + PRIORITY HAZARD line + security-negative silent-allow text) |

**E2 — Cross-platform build:** `go build ./...` → exit 0; `GOOS=windows GOARCH=amd64 go build ./...` → exit 0.
**E3 — Coverage:** N/A (no Go source changed; shell + comment edits only).
**E4 — Subagent boundary:** N/A (hook scripts, not Go subagent domain — and 0 NEW AskUserQuestion refs introduced).
**E5 — Lint:** `golangci-lint run --timeout=3m` → 0 issues.
**E5b — Mirror parity + neutrality:** `go test ./internal/template/... -run 'TestRuleTemplateMirror|TestHookOfficialCompliance|TestInternalContentLeak'` → `ok 0.393s`.
**Gaps:** acceptance.md AC-023 verification uses `head -20` windowed grep — confirmed the guard lands within lines 5-13 (well inside the 20-line window). AC greps reference `development/hooks-system.md` (stale path defect — pre-existing, for manager-spec follow-up).
**Residual-risk:** the header guard is advisory documentation (the plan-preferred lighter touch over loud-abort); it does not mechanically prevent registration. A future runtime that tightens this could enforce it, but that is out of M6 scope.

### Path-discrepancy finding (acceptance.md staleness — for manager-spec follow-up)

acceptance.md AC-HOC-008..016 verification greps reference `.claude/rules/moai/development/hooks-system.md`. The canonical hooks-system.md lives at `.claude/rules/moai/core/hooks-system.md` (confirmed via `find` — no `development/hooks-system.md` exists; hook-independence.md §8 cross-references `core/hooks-system.md` as the canonical event/execution-type reference). The AC greps were run against the correct `core/` path and all PASS. The `development/` path in acceptance.md is stale; flagged for a manager-spec path correction in a follow-up (NOT fixed here — acceptance.md body content is outside manager-develop's ownership per the forbidden-crossings matrix).

### Worktree integration + race absorption (operational note)

manager-develop ran in an auto-materialized worktree (`worktree-agent-af45e51ce8b611a9b`, commit `7ea9edda5`). Orchestrator integrated via `git checkout 7ea9edda5 -- <M1 paths>` after a cherry-pick was blocked by an unrelated staged rename. A parallel session's `git stash -u` had absorbed the untracked SPEC dir; orchestrator recovered the 4 SPEC artifacts from `stash@{0}^3` (byte-identical to manager-spec's plan-phase output, acceptance.md carries the D1/D2/D4 revision). No content lost. Per [[feedback_shared_checkout_concurrent_commit_race]], integration used pathspec checkout (not `git add -A`).

---

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

---

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

---

## §F Phase 0.95 Mode Selection

Input parameters: tier=L; scope≈11 files (5 hook template/live pairs + agent-common-protocol + test); domain count=1 (hook compliance); language mix=shell+markdown+Go-inspection; concurrency benefit=LOW (coding-heavy, tightly-coupled edits); Agent Teams prereqs=not met (team.enabled=false).

Decision: **sub-agent** (Mode 5) — sequential manager-develop per milestone.

Justification: Tier L + coding-heavy + single-domain → Anthropic coding-task parallelism caveat applies (most coding tasks involve fewer truly parallelizable tasks than research). M1 was implemented by one manager-develop in an auto-materialized worktree (cycle_type=tdd); M2-M8 are sequential single-domain doctrine/wrapper edits that do not benefit from fan-out. Mode 4 (parallel) and Mode 6 (workflow) are not selected: the work is not research-heavy nor a ≥30-file uniform mechanical transform.

---

## §H Recursive Self-Diagnosis Log

_<pending run-phase>_

---

## §I Token Accounting

_<pending sync-close>_
