# SPEC-JEV-GOAL-DIST-001 — Progress

Card: t1020 · Tier M · plan-phase artifacts authored 2026-09-20; revised 2026-09-22 (plan iter-2, v0.2.0). Split from `SPEC-JEV-INTEGRATION-001` on the M7+M8 seam.

## §E.1 Plan-phase Audit-Ready Signal

| Item | State |
|---|---|
| SPEC ID regex check | `PASS` (executed) |
| SPEC ID collision | none |
| Tier | M — REQ 14 / ceiling 16; AC 14 / ceiling 16 |
| Artifact set | spec.md · plan.md · acceptance.md · progress.md (Tier M: 3 artifacts + progress) |
| Requirements | 14 (REQ-JEVG-001 … REQ-JEVG-014; ids unchanged since 0.1.0) |
| Acceptance criteria | 14 (AC-JEVG-001 … AC-JEVG-014; 013/014 appended at iter-2, none renumbered) |
| Plan-audit history | iter-1 FAIL 0.68 (Tier M threshold 0.80; `.moai/reports/SPEC-JEV-GOAL-DIST-001/plan-audit-iter1.md`, defects D1-D10) → iter-2 revision resolving all ten (seat (i) withdrawn per §C.1 disposition; MCP wrapper gate-coupled; record destinations named at `docs/jev-negative-results.md`; baselines pinned at `ef3ad83e2`) |
| Baseline pin | plan iter-2 freeze `ef3ad83e2` (branch `WT-goal-dist`); re-pin to the run-entry SHA at run-phase start per acceptance.md §Baseline pin |
| Predecessor | `SPEC-JEV-CONSUMERS-001` (closed `completed`; M5 block recorded, N2 open — the provenance of the seat-(i) withdrawal) |
| Successor | none — last of four |
| Status transition | (none) → draft |

No open question is owned by this SPEC. The chain questions this SPEC records carry per-question statuses in spec.md §F (Q2/Q4 OPEN, CORE; Q3 OPEN, OPTIN; R1 RESOLVED; N1 SETTLED; **N2 OPEN and unowned — it blocked CONSUMERS M5 and is the reason seat (i) is withdrawn, not pending**). The aitmpl ops-checklist is recorded DEFERRED in spec.md §F.

**Implementation Kickoff Approval — GRANTED** (2026-09-22, operator message 「킥오프 진입 진행하자」; plan-audit iter-2 **PASS 0.88** attached, verdict `.moai/reports/SPEC-JEV-GOAL-DIST-001/plan-audit-iter2.md`; monotonic 0.68 → 0.88, Tier M threshold 0.80 met). Progression mode: **autonomous** default per goal.md §Progression Mode (operator declined to choose; `run.md` §Run-phase Autonomy `ac_converge` governs goal arming downstream of this gate). Owed at run entry: D11 five one-line cross-reference fixes (plan-audit-iter2.md, non-blocking); baseline re-pin to the run-entry SHA per acceptance.md §Baseline pin. Plan-audit skip-eligible at `/moai run` (PASS + 0.88 ≥ 0.80 + artifacts hash unchanged since `59b66a77b`) — the skip rationale MUST be recorded in the run-phase delegation prompt Section A. N2 remains OPEN-unowned; operator decision deferred, non-blocking for this SPEC. Run executes in THIS worktree (branch `WT-goal-dist`); lanes do not push — integration via the lead-named window per `.claude/rules/local/gitflow-lane-protocol.md`.

## §E.2 Run-phase Evidence

**Baseline re-pin (acceptance.md §Baseline pin — first act of run-phase §E.2 authorship).** Run entry landed on `c07aa8daa` (branch `WT-goal-dist`, 2026-09-22). The pre-SPEC baseline pinned at plan iter-2 (`ef3ad83e2`) is re-pinned to the run-entry SHA `c07aa8daa`; every before/after criterion in this SPEC is judged against `c07aa8daa`, never a branch name. Nothing product-tree-shaped sits between the two pins: `git diff ef3ad83e2..c07aa8daa --stat` names only the four SPEC artifacts (acceptance.md, plan.md, progress.md, spec.md — measured 2026-09-22, this worktree).

### Milestone evidence

Every row below was measured in this worktree against the re-pinned baseline `c07aa8daa` (or the commit named in the row); commands and verbatim outputs are abbreviated to their deciding lines, with the full runs in the cited test names.

**M7a — seat (i) disposition + regression proof (commit `1470c5fec`).**

- AC-JEVG-001 persisted shape: `go test ./internal/cli/ -run TestMissionBlockedQuestionPersistedShapeUnchanged` → `PASS` (state=blocked, last_blocker, contract lineage intact, mode 0600, and no `question`/`rout`/`classif`/`answer` key at any JSON depth). Diff method: `git diff c07aa8daa -- internal/mission/supervisor.go internal/mission/auto_state.go internal/cli/goal.go internal/cli/state.go | wc -l` → `0`.
- AC-JEVG-002: `go test ./internal/cli/ -run TestGoalLoopCarriesNoUserQuestionSurface` → `PASS` (loop files carry zero `AskUserQuestion` references outside comments; positive control `internal/mission/supervisor.go` fires).
- Observed failure (guard verification, `verification-completeness.md` §1.1): mutant `StateBlocked` → `StateRunning` in `persistMissionBlock` flipped the shape test red — `--- FAIL: TestMissionBlockedQuestionPersistedShapeUnchanged ... persisted state = running, want "blocked"` — then reverted to green.
- Seat-(i) not built: the only Go path holding lane questions still reads `blocker-*.json` (`internal/cli/state.go` `runShowBlocker`); no host path appeared in the tree (no re-plan trigger). N2 remains OPEN and unowned; CONSUMERS' recorded state not amended.

**M7b — Jev Noul as a separate receipt item (commit `6870aa11d`).**

- RED (verbatim, pre-GREEN): `internal/mission/governance_receipt_jev_test.go:73:32: undefined: AuxiliarySignal` (+9 more compile errors) — `FAIL github.com/modu-ai/moai-adk/internal/mission [build failed]`.
- GREEN: `AuxiliarySignal{QuestionID, Kind, Noul, Probability}` on `GovernanceReceipt.AuxiliarySignals` (`omitempty`); deliberately absent from `validateGovernanceBinding`; `canonicalGovernanceReceipt` deep-copies it so the integrity digest covers what the governor was shown.
- AC-JEVG-003: `go test ./internal/mission/ -run TestGovernanceReceiptRecordsJevNoul` → `PASS` (separate item round-trips, binding fields unchanged, file mode 0600). Backward compatibility: `TestGovernanceReceiptWithoutAuxiliaryItemKeepsPreSpecShape` → `PASS` (no `auxiliary_signals` key when none supplied, so pre-SPEC digests still validate).
- AC-JEVG-004: `git diff c07aa8daa --stat` over `.claude/agents/moai/mission-governor.md` and its template mirror → `0` lines on both, and the two copies byte-identical (`cmp` clean). Governor output shape untouched.
- AC-JEVG-005: `TestCompletionPredicateInputsCarryNoJevSymbol` → `PASS` (enumerated binding-field sweep asserted non-vacuous; `internal/jev` absent from the contract/receipt writers; positive control `jev_skill_suggest.go` fires).

**M8a — gate-inert MCP wrapper (commit `65461f10d`).**

- RED (verbatim, pre-GREEN, three cycles): (1) `undefined: newJevAskClient` / `undefined: handleJevAsk` — build failure; (2) `TestJevCallPath_HasExactlyTheDeclaredConsumers`: `internal/cli files outside the declared consumer set import internal/jev: [mcp_jev.go]`; (3) `moai-mcp-tools.md says "30" tools exposed; the registered set holds 31`.
- GREEN: `internal/cli/mcp_jev.go` — `jev_ask` wraps `jev.Client.Ask` only; no transport code in the wrapper (REQ-JEVG-006); the gate read is FIRST and short-circuits: gate off → zero client constructions (counted via the `newJevAskClient` seam), no network call, structured `gated_unavailable` result.
- AC-JEVG-013: `TestJevAskGateOffReportsGatedUnavailableAndConstructsNothing` → `PASS` (`constructed == 0`, `status == gated_unavailable`, not an error). Gate-on happy path (`TestJevAskGateOnDelegatesToJevClientAndMapsAnswers`, local httptest endpoint) and the no-credential typed absence both `PASS`.
- Guard reconciliation (B2 disposition): `TestJevCallPath_HasExactlyTheDeclaredConsumers` allowlist **EXTENDED** with `mcp_jev.go` cited to this SPEC's M8a — the extension its own comment declares as the sanctioned arrival path.
- AC-JEVG-007: `TestMCPToolCatalogueDocsStayMirrorIdentical` + `TestMCPToolCatalogueFiguresMatchRegistry` → `PASS` — both copies of `moai-mcp-tools.md` byte-identical (and the catalogue companion likewise), all four stub figures (31 total; 27 of the 31) and every companion count figure equal to `len(MoaiMCPToolNames())` = 31, family arithmetic derived from the registry (31 − 4 session-messaging = 27). Catalog size pin updated 30 → 31 per that test's own instruction. Project-root doc test unaffected: `jev_ask` declares no `project_root` input.
- Web console: `internal/web/assets/i18n.js` carries the gated `jev_ask` enablement label in all four locales; `go test ./internal/web/` → `ok` (i18n parity tests green).
- **Mirror-test enrollment decision (AC-JEVG-007 honesty note): DECLINE** enrolling `moai-mcp-tools.md` in `internal/template/rule_template_mirror_test.go` — equivalent-or-stronger dedicated enforcement already exists in `TestMCPToolCatalogueDocsStayMirrorIdentical` (byte-compare of BOTH catalogue files), and the dedicated test additionally asserts the figures against the registered set, which the generic mirror test cannot. Record made per plan §C2.
- **Gate-unrun disposition record (REQ-JEVG-007):** the wrapper ships present-but-unpresented. What the fitness gate still needs: the binding labelled-set measurement defined by `SPEC-JEV-OPTIN-MEASURE-001` REQ-JEVO-009 — each consumer's typed answers compared against its constant-answer baseline on a labelled set of the size OPTIN's Kickoff gate must fix (chain question Q3). Who owns it: `SPEC-JEV-OPTIN-MEASURE-001` owns the gate; consumer fit records sit under `SPEC-JEV-CONSUMERS-001` REQ-JEVN-016 condition (iii), which this record follows without amending. Until that gate runs, no shipped or template surface presents `jev_ask` as available. This record cites **no measurement figure** for the gate, by requirement.

**M8b — reference skill (commit `64b320632`).**

- AC-JEVG-008: `TestJevQuestionDesignSkillCarriesNoCallPath` → `PASS` — the forbidden token set (`internal/jev`, `mcp__moai__jev`, `jev_ask`, `moai jev`) absent from both copies of `.claude/skills/moai-ref-jev-question-design/SKILL.md`; positive controls (`jev_skill_suggest.go` for the import path, the catalogue doc for the registered name) both fire. `TestJevQuestionDesignSkillCopiesStayIdentical` → `PASS` (byte-identical).
- Skill content: question-design rules only — compute the question in code, always admit a no-match answer, keep the state small, phrase and record both noul polarities, typed absences are no signal.
- Template-first + catalog: entry added under `catalog.yaml` `core.skills` (hash filled by `gen-catalog-hashes.go --entry`); count pins moved per each test's own provenance convention (skills 34 → 35, total entries 46 → 47). `go test ./internal/template/` → `ok`.

**M8c — records (commit `3deef0992`).**

- AC-JEVG-012: `docs/jev-negative-results.md` §1 carries the measured figures (dead-call precision 29.2% against a 25% base rate; 2-class 58.9% against the 75.0% constant; English control 67.5% on 40 cards; gate sweep flat 0.30–0.80), cites the gitignored local evidence (`.moai/reports/t943/verdict.md`, primary checkout) and the maintainer guide (§30) as origins, and re-measures nothing. Figure-presence grep: 3 table lines cover all five figure patterns.
- AC-JEVG-014: §2 states the `scripts/jev/` disposition — uncommitted working copy, not committed, not distributed (no template mirror), not maintained — with `internal/jev` as the canonical implementation and nothing deleted from anyone's tree.
- REQ-JEVG-014 triage guard: `TestTodoTriageStaysModelFree` → `PASS` (zero `jev` references in `todo_triage.go`; positive control fires). Observed firing captured on a mutant (a temporary `jev` token in `todo_triage.go` flipped it red; removed, `git diff` clean, re-verified green).

### Post-run verification batch

- E2 build: `go build ./...` → `DARWIN_BUILD_OK`; `GOOS=windows GOARCH=amd64 go build ./...` → `WINDOWS_BUILD_OK`.
- E3 coverage: `internal/mission` 88.1%, `internal/mcp` 100.0% (both ≥ 85% target). `internal/template` 81.7% whole-package (the SPEC's changes there are YAML + test pins; the figure is the package baseline, not a regression — pre-SPEC figure not re-measured). `internal/cli` whole-package coverage runs ~10 min and hit the default test timeout on this machine with zero test failures (known shape); its verdict is deferred to CI per the CI 3-tier policy (B5).
- E4 subagent boundary: the dispatch's literal grep (`internal/cli` + `internal/jev`, non-test, minus `// `-prefixed lines) yields **32 hits, none a question surface**: multi-line comment-block continuation lines the single-line filter cannot strip, help-text string literals stating the prohibition ("does not call AskUserQuestion"), the `agentlint` enforcement instrument itself (`checkLiteralAskUserQuestion`), and a lint testdata fixture. The sealed loop's own files (`goal.go`, `goal_runnable.go`, `hook_stop_goal.go`) carry zero, pinned by `TestGoalLoopCarriesNoUserQuestionSurface`; invocation-shaped search `AskUserQuestion(` yields only the lint instrument + its fixture. Positive control (`internal/mission/supervisor.go` identifier) fires.
- E5 lint: `golangci-lint run --timeout=2m` → `0 issues.` both pre-flight (baseline `c07aa8daa`) and post-run — no NEW issues. `go vet` on `internal/cli`, `internal/mission`, `internal/mcp`, `internal/template` → clean.
- E6 commits: `1470c5fec` (M7a + draft→in-progress + re-pin) · `6870aa11d` (M7b) · `65461f10d` (M8a) · `64b320632` (M8b) · `3deef0992` (M8c) · this records commit. Push state: **not pushed (lane discipline — the lead batch-pushes develop after the integration window)**.
- E8 RED outputs: captured verbatim pre-GREEN per TDD cycle — M7b compile failure (`undefined: AuxiliarySignal`), M8a build failure + guard firing + figure drift, plus mutant-probe observed failures for the two characterization guards (M7a shape, M8c triage). Full verbatim text in the commit bodies and this section.

### Chain-status note

This SPEC's completion does **not** make the chain release-ready: the fitness gate (OPTIN REQ-JEVO-009) stands unrun, the consumers remain in their recorded gate-unrun state, and N2 is still open and unowned — seat (i) stays withdrawn, carried by spec.md §C.1 and plan.md §B1.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-09-22
run_commit_sha: pending-backfill-run
run_status: complete
ac_pass_count: 14
ac_fail_count: 0
preserve_list_post_run_count: 9
l44_pre_commit_fetch: "not performed — lane-local branch (WT-goal-dist), no remote ref consulted; HEAD + branch re-read immediately before every commit (plain git forms) per the worktree guard"
l44_post_push_fetch: "n/a — the lane never pushes; remote develop landing is the lead's batch-push (git-flow lane protocol 2026-09-02)"
new_warnings_or_lints_introduced: 0
cross_platform_build.darwin: pass
cross_platform_build.windows: pass
total_run_phase_files: 26
m1_to_mN_commit_strategy: "per-milestone commits on WT-goal-dist — M7a 1470c5fec, M7b 6870aa11d, M8a 65461f10d, M8b 64b320632, M8c 3deef0992, records commit (this one); no push, no PR, no tags"
```

`preserve_list_post_run_count: 9` — measured unchanged against the re-pinned baseline `c07aa8daa`: `internal/jev` (untouched; `git diff` empty), `.claude/agents/moai/mission-governor.md` + template mirror (AC-JEVG-004 diffs both, 0 lines), `internal/mission/supervisor.go`, `internal/mission/auto_state.go`, `internal/cli/goal.go`, `internal/cli/state.go` (AC-JEVG-001 diff set, 0 lines), `internal/cli/todo_triage.go` (restored byte-clean after the mutant probe), `.claude/settings.local.json` (untouched). `scripts/jev/` lives in the primary checkout, outside this worktree — not committed here by design (REQ-JEVG-012).

`run_commit_sha` is a placeholder by the D3 backfill pattern: a commit cannot cite its own SHA; the records commit's real SHA is backfilled in the follow-up commit.

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

## §F Phase 4 Mode Selection

Decision: **serial** — one `manager-develop` spawn carrying the milestone chain M7a→M7b→M8a→M8b→M8c sequentially. Alternatives not selected: direct (multi-file semantic implementation, not a typo-fix), fanout (coding-heavy, not research-heavy), sweep (<30 files, not one uniform mechanical transform), agent-team (no operator request).

| Input | Value |
|---|---|
| tier | M |
| scope (file count) | ~12-15 (Go: MCP wrapper registration + receipt item + tests; rules: `moai-mcp-tools.md` ×2 copies ×2 figures; reference skill ×2 copies; `docs/jev-negative-results.md` new; template emissions if any) |
| domain count | 4 (Go source, rules docs, skills, templates/records) |
| file language mix | Go + markdown |
| concurrency benefit | LOW — coding-heavy (Anthropic coding-task parallelism caveat) |
| agent-team prereqs | not requested (no `--team` / no Team scale label) |

Justification: the work is coding-heavy Go implementation with a dependency-ordered milestone chain (records last because the surfaces they document are only final after M8a/M8b), so the sequential single-spawn path is both the Anthropic-recommended default and the only mode whose milestone ordering the dependency chain permits. Selected 2026-09-22 by the run-entry session before the first run-phase `Agent()` spawn, per `orchestration-mode-selection.md` §D.
