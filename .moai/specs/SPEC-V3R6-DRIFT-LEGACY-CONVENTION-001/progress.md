# Progress — SPEC-V3R6-DRIFT-LEGACY-CONVENTION-001

> Run-phase progress tracking for the `moai spec drift` comprehensive false-positive
> elimination + era/grandfather/terminal alignment + close-subject doctrine SPEC.
> Tier M, cycle_type=tdd (RED-GREEN-REFACTOR). Status transitioned draft→in-progress
> on the M1 commit per the Status Transition Ownership Matrix (manager-develop owns
> `draft → in-progress`).

## §E — Phase 0.95 Mode Selection

- **Input parameters**: tier = M; scope = ~4 files (drift.go, transitions.go, 2-3 new
  _test.go, 2 doctrine files + 1 template mirror); domain = Go source + doctrine markdown
  + template mirror; file language mix = mostly Go + some markdown; concurrency benefit =
  LOW (coding-heavy, single-package internal/spec, sequential milestone dependency);
  Agent Teams prereqs = not met (single-package coding task).
- **Decision**: `sub-agent` (Mode 5 — sequential sub-agent per milestone).
- **Justification**: This is coding-heavy work in a single package (`internal/spec`) with
  a strict milestone dependency chain (M1/M2 independent → M3 depends on classifier seam
  stability → M4 doctrine independent → M5 verification gate). Per Finding A4
  (coding-task parallelism caveat), sequential sub-agent is the correct default for coding
  tasks; parallel/agent-team modes add coordination overhead without benefit for a
  single-package edit.

## Milestone Progress

| Milestone | Status | Commit | Notes |
|-----------|--------|--------|-------|
| M1 — era exemption + terminal authority (drift.go) | done | d8d2d8d9c | DetectDrift ③+④; D3 record-emission contract |
| M2 — stale sync→completed rule correction (transitions.go) | done | 2fb2a5156 | 4-phase model: sync = implemented |
| M3 — combined-scope secondary prefix-grep fallback (drift.go) | done | 7493472e0 | 3-gate LSGF-001-safe; live CCSYNC CLAUDEMD+TOOLCAT resolved |
| M4 — close-subject full-ID doctrine amendment + mirror | done | 041e6fefb | lifecycle-sync-gate.md + spec-frontmatter-schema.md + template mirror (§25 C2 generalized) |
| M5 — verification gate + genuine-⑤ residual classification | done | d00061a25 | drift 54→8; audit byte-identical; 12/12 AC PASS; residual classified (5a/5b/5c handoff) |

## §E.2 Run-phase Evidence

| AC | Severity | Status | Verification | Actual Output |
|----|----------|--------|--------------|---------------|
| AC-DLC-001 | MUST-PASS | PASS | TestDetectDrift_CombinedScopeFallback + TestCombinedScopeCloseMatches + live jq | synthetic: BOTH FOO+BAR resolve completed; live: CLAUDEMD+TOOLCAT Drifted=false |
| AC-DLC-002 | MUST-PASS | PASS | TestClassifyPRTitle_StaleSyncRuleCorrected + live | bare sync/docs(sync)→implemented; close-infix→completed; STATUSLINE-STDINFIELDS absent from drift |
| AC-DLC-003 | MUST-PASS | PASS | TestDetectDrift_TerminalStateAuthoritative | superseded/archived/rejected → Drifted=false (record PRESERVED) |
| AC-DLC-004 | MUST-PASS | PASS | TestDetectDrift_GrandfatherEraExempt + V3R6StillDetected | V2.x/V3R2-R4 exempt; V3R6 still detected; AP-3 guard passes |
| AC-DLC-005 | MUST-PASS (gate) | PASS | drift_chore_skip_test + drift_test backfill | all chore-skip + backfill-skip regression green |
| AC-DLC-006 | MUST-PASS (gate) | PASS | TestGetGitImpliedStatus_SPECIDWordBoundary (5 subcases) | word-boundary protection unchanged; HARNESS001 NOT planned |
| AC-DLC-007 | MUST-PASS (gate) | PASS | main-HEAD binary vs my binary audit --json diff (same corpus) | byte-identical; era.go/audit.go call NONE of modified funcs |
| AC-DLC-008 | MUST-PASS | PASS | TestDetectDrift_CombinedScopeCollisionGuard + era V3R2/V3R3 exemption | grandfathered sibling not newly flagged |
| AC-DLC-009 | SHOULD-PASS | PASS | drift --count before/after | 54 → 7 (strict decrease) |
| AC-DLC-010 | MUST-PASS (gate) | PASS | go test ./internal/spec/... + go vet + golangci-lint | green; lint 0 issues |
| AC-DLC-011 | MUST-PASS | PASS | TestCloseSubjectDoctrineAmendment (portable Go regexp, D-NEW-2) | both doctrine files co-located prohibition; template mirror §25-generalized; mirror-parity + leak tests green |
| AC-DLC-012 | MUST-PASS | PASS | TestDetectDrift_CombinedScopeCollisionGuard + TestCombinedScopeCloseMatches (OTHER/FOOBAR/EXTRA) | non-sibling partial-prefix not mapped; D-NEW-1 FOO vs FOOBAR guard |

## PRESERVE Verification

- `internal/spec/era.go` — byte-unchanged (verified `git status --porcelain` clean): PASS
- `internal/spec/audit.go` — byte-unchanged (verified): PASS

## Baselines (captured pre-fix)

- `moai spec drift --count` = 54 (drifted records)
- `moai spec audit --json`: grandfathered=268, modern_era_clean=24, drift_findings=291
- Named exemplars (pre-fix): SPEC-CCSYNC-CLAUDEMD-001 [completed↔implemented Drifted],
  SPEC-CCSYNC-TOOLCAT-001 [completed↔implemented Drifted],
  SPEC-CCSYNC-DYNWF-001 [in-progress↔implemented Drifted — collision case].

## M5 Post-fix Results

- `moai spec drift --count` = 54 → **8** (strict decrease, AC-DLC-009). (Earlier intermediate
  read showed 7; the 7↔8 variance is parallel-session corpus churn — directional binding
  signal holds.)
- Named-exemplar jq proof (post-fix): SPEC-CCSYNC-CLAUDEMD-001 Drifted=false (combined-scope
  fallback), SPEC-CCSYNC-TOOLCAT-001 Drifted=false (combined-scope fallback),
  SPEC-CCSYNC-DYNWF-001 Drifted=false (genuinely closed by its own per-SPEC commit 21ac357c1
  by the parallel session; fallback NOT invoked — primary walk gives completed). The
  load-bearing collision guard (DYNWF NOT cleared by the (CLAUDEMD + TOOLCAT) combined close)
  remains pinned by the deterministic unit fixture TestCombinedScopeCloseMatches/DYNWF_NOT_named,
  immune to repo churn.
- Audit parity (AC-DLC-007): pristine-main binary vs post-fix binary `moai spec audit --json`
  on the same corpus → BYTE-IDENTICAL. era.go/audit.go call NONE of the modified functions.
- Coverage 87.6% (≥85%). go test ./... green (93 packages). go vet/lint clean.
  GOOS=windows build exit 0. Observation-only discipline verified (no new write primitive).

## Genuine-⑤ Residual Classification (INFORMATIONAL — handoff to follow-up SPEC)

The 8 post-fix residual drift records are NOT in this SPEC's mechanism ①②③④ scope. They
decompose into 3 NEW sub-classes for a follow-up SPEC (this SPEC MUST NOT clear them — AP-1):

**5a — post-close fix shadowing** (NEW mechanism): a `fix`/`feat` commit landed AFTER the
`4-phase close`, so the newest-first walker adopts the post-close commit (→implemented) before
reaching the close. The frontmatter `completed` is correct; git-walk infers `implemented`.
  - SPEC-V3R6-LIFECYCLE-SYNC-GATE-001 [completed↔implemented] — close 203ed896b, then fix 38d07a6a0
  - SPEC-V3R6-DOCS-I18N-PARITY-001 [completed↔implemented] — close present, then post-close fix

**5b — group-label combined close** (NEW combined-scope variant NOT matched by M3 prefix-grep):
closed by `chore(SPEC group C): Mx-phase close (...)` which names a free-text group label
("SPEC group C"), NOT a `SPEC-<PREFIX>` scope-prefix. `deriveScopePrefix` cannot derive a
hyphen-delimited prefix from a group label, so the M3 fallback correctly does not fire.
  - SPEC-V3R6-MAIN-RED-REMEDIATION-001 [completed↔in-progress]
  - SPEC-V3R6-PROMPT-CACHE-001 [completed↔implemented]

**5c — genuine no-close / frontmatter lag** (true mechanism ⑤ — the deferred operational set):
  - SPEC-AUTONOMY-RUN-GOAL-001 [implemented↔completed] — frontmatter BEHIND (git has close
    11871ce61 → completed; frontmatter still implemented; needs frontmatter→completed)
  - SPEC-V3R6-HARNESS-NAMESPACE-CLEANUP-001 [completed↔implemented] — no per-SPEC close commit
  - SPEC-V3R6-LOCAL-NAMESPACE-CONSOLIDATION-001 [completed↔in-progress] — no per-SPEC close commit
  - SPEC-V3R6-TEMPLATE-MIRROR-DRIFT-001 [completed↔in-progress] — retroactive chore, no close-infix

Follow-up SPEC recommendation: (i) handle 5a (post-close fix shadowing — walker should treat a
close-infix commit ANYWHERE in the window as authoritative for `completed`, even if a newer
non-close fix shadows it); (ii) handle 5b (group-label combined close — extend combinedScopeClose
matching to free-text group labels with an explicit sibling enumeration); (iii) 5c is operational
(actually run `moai spec close` / backfill frontmatter — NOT a detector change).

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-06-03
run_commit_sha: "d00061a25"
run_status: implemented
ac_pass_count: 12
ac_fail_count: 0
preserve_list_post_run_count: 2  # era.go + audit.go byte-unchanged
new_warnings_or_lints_introduced: 0  # golangci-lint 0 issues
cross_platform_build:
  darwin: pass
  windows: pass  # GOOS=windows GOARCH=amd64 go build ./internal/spec/... exit 0
total_run_phase_files: 7  # drift.go, transitions.go, +4 new/edited _test.go, 2 doctrine + 1 template mirror
m1_to_mN_commit_strategy: "per-milestone via L1 worktree → cherry-picked to branch (M1 d8d2d8d9c, M2 2fb2a5156, M3 7493472e0, M4 041e6fefb, M5 d00061a25)"
drift_count_before: 54
drift_count_after: 8
audit_parity: byte-identical
coverage_internal_spec: 87.6%
```

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-06-03
sync_commit_sha: "eae3cdad9"
sync_status: implemented
spec_frontmatter_transitions:
  - field: status
    old_value: in-progress
    new_value: implemented
  - field: updated
    old_value: 2026-06-03
    new_value: 2026-06-03
changelog_entry_position: "[Unreleased]/Fixed section"
progress_sections_updated:
  - "§E.3 Run-phase Audit-Ready Signal (§E.4 added for sync)"
  - "§E.4 Sync-phase Audit-Ready Signal (NEW)"
```

### (Migrated from §E.5)

```yaml
mx_complete_at: 2026-06-03
mx_commit_sha: "feff43cd8"
mx_status: completed
spec_frontmatter_transitions:
  - field: status
    old_value: implemented
    new_value: completed
  - field: version
    old_value: "0.1.1"
    new_value: "0.2.0"
four_phase_close:
  plan_artifacts: "authored by manager-spec (uncommitted in main), committed within run M1 d8d2d8d9c"
  run: "d8d2d8d9c (M1), 2fb2a5156 (M2), 7493472e0 (M3), 041e6fefb (M4), d00061a25 (M5)"
  sync: eae3cdad9
  backfill: 211d79936
  mx: "feff43cd8"
ac_final: "12/12 PASS"
drift_outcome: "54 → 8 (4 false-positive mechanisms ①②③④ resolved; 8 genuine-⑤ residual handed off to follow-up SPEC)"
einstein_md_failure: "pre-existing uncommitted template drift (last touched 3d588317e, out of scope; not introduced by this SPEC)"
close_subject_doctrine_dogfood: "this close commit uses the full SPEC-ID per REQ-DLC-011 — the close-subject full-ID mandate this SPEC shipped"
```
