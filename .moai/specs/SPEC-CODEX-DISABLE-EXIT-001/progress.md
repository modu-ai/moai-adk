# progress.md — SPEC-CODEX-DISABLE-EXIT-001

## §E.1 Plan-phase Audit-Ready Signal

plan_status: pending-plan-audit
plan_complete_at: (pending — set at plan-phase close / audit PASS)
baseline_tree: a4855f0b2
card: t548
tier: S (artifact set 4 files per dispatch order)
M1_status: done (caller census performed at plan phase — evidence below)

### Plan-phase Evidence — M1 caller census (commands + verbatim outputs)

All searches run with `rg` against this worktree at `a4855f0b2`, this run, 2026-09-08.

1. The card's literal form — `rg -n --no-heading "codex skills" --hidden -g '!.git' -g '!node_modules'` → **4 hits, all documentation of the REJECTED form** (`.moai/specs/SPEC-CODEX-SKILL-DISABLE-001/plan.md:74` records its design-time rejection; `SPEC-CODEX-SKILL-PATH-SLASH-001` prose). Zero invocations — the form does not exist as a CLI surface.

2. The actual verb — `rg -n --no-heading "skills disable" --hidden -g '!.git' -g '!node_modules'` → **17 hits**, partitioned:
   - In-repo invocations: `.moai/reports/t502/e2e-verb.sh:19,41,43,68` (one dev-only E2E evidence script from the verb's own card t502; not shipped).
   - Go callers: `internal/cli/codex_skills_disable_test.go` — 10+ direct `runCodexSkillDisable(...)` calls (`:293,317,348,369,533,597,617,645,666`) + `newSkillsCmd()` at `:511`; they consume the returned `error`, not process exit codes.
   - Source/doc mentions: `internal/cli/skills.go:8`, `codex_skills_disable.go:387`, `CHANGELOG.md:20,386`, `.moai/project/codemaps/entry-points.md:59`, SPEC prose.

3. Sibling form — `rg -n --no-heading "skills prune" --hidden -g '!.git'` → **0 hits**. The sibling verb's real surface is `moai clean --codex-skills` (`internal/cli/codex_skills_prune.go:157`; `Kept:` at `:195`).

4. Templates — `rg -n "skills disable" internal/template/templates/` → **0 hits** (no hook wrapper, no template).

5. CI — `rg -n "skills" .github/workflows/ | rg -i "disable|prune|codex"` → **0 hits**.

6. docs-site — `rg -ln "skills disable|codex skills|skills prune|skills.config" docs-site/content/` → 8 files (doctor.md ×4 locales, moai-clean.md ×4 locales — all about `[[skills.config]]` shape, none about the disable verb); `rg -n "moai skills" docs-site/content/` → **0 hits**. The verb is undocumented in the user docs.

7. README ×4 — `rg -n "skills disable|skills prune|codex-skills" README*.md` → only `moai clean --codex-skills` (line 749 ×4). The disable verb is undocumented.

**Census verdict**: zero production in-repo callers (scripts / hooks / CI / templates). The exit-code contract exists for outside-this-repo consumers and as a published CLI surface; no published surface documents the Skipped branch's exit code, so there is no documented promise an exit-code change would break.

### Plan-phase Evidence — code reading (this tree, `a4855f0b2`)

- Runner branch table B1-B10 read at `internal/cli/codex_skills_disable.go:394-455` (spec.md §B.1). Tally: nil returns at `:399,408,414,421,424,434,454`; `fmt.Errorf` at `:402` (name resolution), `:441` (backup), `:451` (write); plus command-level errors at `skills.go:90,96`.
- Skipped reasons enumerated in `upsertCodexSkillDisable` (`:231-311`): no path `:237-239`; unencodable char `:261-263`; duplicates `:282-284`; unrecognized line `:293-295`; no rewritable key `:302-304`. Reasons (c)/(e) are actionable by `moai clean --codex-skills` — exit 0 hides that handoff.
- Documented intent read at `:387-393` (doc comment: fail-open scoped to missing inputs; name-not-found = typo = non-zero "so a script can see it") and `CHANGELOG.md:20` (mirror-absent-zero rationale). **Unchanged/Skipped bucketing is undocumented on every surface.**
- Sibling fail-open doctrine read at `codex_skills_prune.go:159-162`; per-entry `Kept:` never affects exit (`:182-196`).
- Verb-tree shape read at `skills.go:53-55`: the `skills` family contains exactly one verb (disable).

### Plan-phase Evidence — adjudication analysis

Both outcomes prescribed in spec.md §E. Outcome A (document-only) is viable but must carry an explicit card-rule-2 waiver (Unchanged/Skipped share bucket 0). Outcome B (Skipped → non-zero; Unchanged and absent-inputs stay 0) is the RECOMMENDED default: rule 2 is a stated operator requirement the current code violates; the code's own doc comment scopes fail-open to missing inputs, not refusals; the two rule-2-actionable skip reasons hand off to a different verb; and the census shows no documented exit-code promise for Skipped to break. Scope axis resolved at plan phase (spec.md §D): verb-local, with the single-target-vs-sweep distinction documented against `moai clean --codex-skills`.

### M2 Adjudication Decision — OPERATOR (via lead relay, 2026-09-08)

- **Decision: Outcome B — exit-code repair.** Skipped (deliberate refusal) → non-zero; Unchanged (already disabled) → stays 0; absent-inputs (B1/B3/B4) → stay 0; name-resolution (B2) → stays non-zero.
- Channel: lead-relayed AskUserQuestion operator judgment, 2026-09-08 (relay message carries the verdict; same relay granted Implementation Kickoff Approval for run-phase entry).
- Evidence carried into this record (the 4 operator-cited grounds, all from §E.1/§B above): (1) caller census zero production in-repo callers; (2) Unchanged/Skipped exit codes undocumented on every published surface; (3) 2 of 5 Skipped reasons hand off to `moai clean --codex-skills` and exit 0 hides that handoff; (4) no measurable breakage (census).
- Card rule 2 (Unchanged ≠ Skipped) enforced per AC-CDE-002, including the RED-now cell observed on `a4855f0b2`.
- Outcome A remains documented in spec.md §E as the rejected alternative (unpicked prescription stays intact).

## §E.2 Run-phase Evidence

Run phase executed 2026-09-08, worktree `.claude/worktrees/t548`, branch `WT-codex-disable-exit`, Outcome B (M2 decision above). Commits: M4-RED `554f93e35` → M3+M4 `c7a9e0830` → M5 docs `<this commit>`.

### Milestone summary

- **M4-first (RED)**: `TestRunCodexSkillDisableSkippedExitsNonZero` authored (duplicate-entry guard-refusal fixture) and observed RED on the pre-change tree BEFORE any implementation edit. Verbatim E8 below.
- **M3 (GREEN)**: `codexSkillDisableSkipped` branch (`codex_skills_disable.go`) returns `fmt.Errorf("refused to disable %q for Codex: %s", opts.Skill, v.Reason)`; the `Skipped:` stdout line is preserved. Re-run → GREEN.
- **M4 rest**: `TestRunCodexSkillDisableUnchangedExitsZero` (Unchanged stays 0 — the other half of rule 2) + `TestRunCodexSkillDisableFailsOpenOnUnresolvedHome` (B3 absent-input, previously uncovered) added. Existing name-resolution tests (AC-CDE-004) pass unchanged.
- **M5**: `--help` Long text (`skills.go`) carries the per-class exit-code table; the runner doc comment restates the three-class contract truthfully. CHANGELOG entry DEFERRED to sync-phase (manager-docs surface) → WRITTEN at sync close: the [Unreleased] § Fixed entry documents the Skipped branch's behavior change (Skipped now exits non-zero; Unchanged and absent-inputs stay 0) and the `--help` exit-code contract.

### E1 — AC PASS/FAIL matrix (attribution: command + verbatim output + HEAD)

| AC | Status | Verification command | Observed result |
|----|--------|---------------------|-----------------|
| AC-CDE-001 | PASS | `grep -c '<marker>' internal/cli/codex_skills_disable.go` per §B.1 marker | `Nothing to disable`→1, `Refusing to write`→1, `Unchanged: `→2, `Skipped: `→4, `[dry-run]`→1, `Disabled `→1 — all ≥1 |
| AC-CDE-002 | PASS | `go test ./internal/cli/ -run 'TestRunCodexSkillDisableSkippedExitsNonZero\|TestRunCodexSkillDisableUnchangedExitsZero' -count=1` | `ok github.com/modu-ai/moai-adk/internal/cli` exit 0; Skipped test observed RED pre-change (E8), GREEN post-change |
| AC-CDE-003 | PASS | `go test ./internal/cli/ -run 'TestRunCodexSkillDisableMirrorAbsentExitsZero\|TestRunCodexSkillDisableFailsOpenOnAbsentConfig\|TestRunCodexSkillDisableFailsOpenOnUnresolvedHome' -count=1` | all PASS (existing B1/B4 tests unchanged + new B3 test) |
| AC-CDE-004 | PASS | `go test ./internal/cli/ -run 'TestRunCodexSkillDisableUnresolvedNameFailsWithoutWriting\|TestRunCodexSkillDisableAmbiguousNameFailsWithoutWriting' -count=1` | both PASS, files unchanged by this SPEC |
| AC-CDE-005 | PASS | `grep -c "verb-local" .moai/specs/.../spec.md` → 3; `git diff --stat a4855f0b2 -- internal/cli/codex_skills_prune.go` → empty | both hold |
| AC-CDE-006 | PASS | `go run ./cmd/moai skills disable --help` | Long help carries the `performed 0 / refused 1 / absent 0` table, exit 0 |
| AC-CDE-007 | PASS | progress.md §E.1 M2 record (this file) vs first run-phase commit | M2 decision record existed in §E.1 before commit `554f93e35`; record carried in the plan-phase commit lineage |

### E8 — verbatim pre-GREEN RED (four elements)

1. **Command**: `go test ./internal/cli/ -run 'TestRunCodexSkillDisableSkippedExitsNonZero' -count=1`
2. **Verbatim stdout**:
```
--- FAIL: TestRunCodexSkillDisableSkippedExitsNonZero (0.00s)
    codex_skills_disable_test.go:748: guard refusal returned nil error — a refused request is indistinguishable from a performed one
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	0.849s
FAIL
```
3. **Exit code**: 1
4. **Tree SHA**: `c6f6193e4` — at that moment `git diff a4855f0b2 --stat -- internal/cli/codex_skills_disable.go internal/cli/skills.go internal/cli/codex_skills_prune.go` printed EMPTY (implementation sources byte-identical to baseline `a4855f0b2`); the only working-tree change was the uncommitted RED test itself. Raw output also persisted at `.moai/state/verify/t548/red-skipped-nonzero.txt`.

The RED is red for the stated reason: the branch returned nil (B6 pre-change behavior), exactly what AC-CDE-002 Outcome B targets.

### E2 — cross-platform build

```
$ go build ./...                             → exit 0
$ GOOS=windows GOARCH=amd64 go build ./...   → exit 0
```

### E3 — coverage

```
$ go test -cover ./internal/cli/ -count=1
ok  	github.com/modu-ai/moai-adk/internal/cli	401.746s	coverage: 81.4% of statements
```

Package-wide 81.4% is the pre-existing level of this large package, not a regression from this SPEC: the touched runner measures 95.0% (`go tool cover -func`, `runCodexSkillDisable`), every added executable line is covered by the three new exit-contract tests, and the additions cannot lower statement coverage (new tests only add covered paths). The uncovered remainder lives in pre-existing helpers this SPEC does not touch (`writeCodexConfigPreservingMode` 50%, `newSkillsDisableCmd` cobra wiring 42.9%, `defaultSkillsProjectRoot` 0%).

### E4 — subagent boundary

```
$ grep -rn "AskUserQuestion\|mcp__askuser" internal/cli/codex_skills_disable.go internal/cli/skills.go internal/cli/codex_skills_disable_test.go
(exit 1 — no matches; run with command grep to bypass the shell's ugrep wrapper)
```

Note: plan.md §E.4 cited `TestNew_NoAskUserQuestion` as a package-wide guard; that test name does not exist in this tree (`[no tests to run]` observed). The real pattern is per-file static guards (`TestNewInventory_NoAskUserQuestion`, `TestNoAskUserQuestionInChain`, …); no new file was created by this SPEC, so no new guard is owed. The direct grep above is the binding evidence.

### E5 — lint

```
$ golangci-lint run internal/cli/...
0 issues.
```

Zero NEW findings (total is zero, so no baseline split needed).

### E6 — commits + push state

| Commit | Subject |
|--------|---------|
| `554f93e35` | test(SPEC-CODEX-DISABLE-EXIT-001): M4 Skipped exit-contract test, RED on pre-change tree (t548) — carries spec.md `draft → in-progress` |
| `c7a9e0830` | feat(SPEC-CODEX-DISABLE-EXIT-001): M3 Skipped guard-refusal exits non-zero (t548) |
| `<this commit>` | docs(SPEC-CODEX-DISABLE-EXIT-001): M5 exit-code contract on the verb surface (t548) |

Push state: **not pushed** — lane does not push (git-flow lane protocol); the lead batch-pushes `origin/develop` after recording the local merge SHA.

## §E.3 Run-phase Audit-Ready Signal

run_complete_at: 2026-09-08
run_commit_sha: pending-backfill-run
run_status: complete
ac_pass_count: 7
ac_fail_count: 0
preserve_list_post_run_count: 0
l44_pre_commit_fetch: not-performed (worktree lane-local work; STEP 0 verified baseline `a4855f0b2` lineage instead — dispatch preconditions held)
l44_post_push_fetch: n/a (lane does not push; lead performs the batch push and its landing verification)
new_warnings_or_lints_introduced: 0
cross_platform_build.darwin: pass
cross_platform_build.windows: pass
total_run_phase_files: 3
m1_to_mN_commit_strategy: 3 commits (M4-RED test / M3+M4 implementation / M5 help+docs), conventional subjects carrying SPEC ID + card t548

## §E.4 Sync-phase Audit-Ready Signal

sync_complete_at: 2026-09-08
sync_commit_sha: pending-backfill-sync
sync_status: complete
card: t548
changelog_entry_position: CHANGELOG.md [Unreleased] § Fixed (top of section)
b12_self_test_a: pass — pre-emission `grep -c 'SPEC-CODEX-DISABLE-EXIT-001' CHANGELOG.md` → 0 (exit 1, no duplicate entry); control grep `Unreleased` → 3 hits confirming the file was actually read (the shell's grep is a ugrep wrapper — run via `command grep`)
b12_self_test_b: pass — 7 distinct live AC identifiers in acceptance.md (AC-CDE-001..007, each appearing once, zero `[RETIRED]`/`[REF]` markers) == 7 referenced in the CHANGELOG entry
b12_self_test_c: pass — every write-surface path claimed in the entry verified present via `git diff a4855f0b2..HEAD --stat` (internal/cli/codex_skills_disable.go, internal/cli/skills.go, internal/cli/codex_skills_disable_test.go) and the SPEC link read directly
frontmatter_status_transitions.in-progress_to_implemented_to_completed: this sync commit (the completed transition merged into the sync commit per the 3-phase close; `status` + `updated` only)
mx_tag_check: pass — the production diff is one unexported branch return (`return fmt.Errorf(...)` inside the existing `runCodexSkillDisable` switch) plus doc comments and a `--help` string literal; no new exported symbol, no goroutine, no complexity growth — no @MX tag owed
sync_audit_note: implementation diff re-read from git (not progress.md prose) before the CHANGELOG entry was authored, per B12 discipline
