# SPEC-UPDATE-MIGRATION-FIX-001 — Progress

SPEC ID: SPEC-UPDATE-MIGRATION-FIX-001
Card: t1578
Status: in-progress (run phase)
Tier: M

## §E.1 Plan-phase Audit-Ready Signal

- Artifacts authored: spec.md, plan.md, acceptance.md, research.md,
  decision-index.md, progress.md (this file) — all under
  `.moai/specs/SPEC-UPDATE-MIGRATION-FIX-001/`.
- Authored on tree: 2aab5f797 (branch WT-update-migration-fixes, card
  worktree `.moai/worktrees/t1578`); the six artifacts landed as lane
  commit f569be5d8 — the plan-artifact baseline is f569be5d8, and the
  anchor-path delta 2aab5f797..f569be5d8 is empty (acceptance.md EV-7),
  so all plan-phase measurements carry over. BASELINE RE-CUT NOTICE: the
  worktree fast-forwarded from 81786284e to 2aab5f797 (main absorbed, 155
  commits) mid-research. All research anchors re-verified on 2aab5f797
  (acceptance.md EV-6): the deny-migration map, its test file, and
  internal/template/deployer_mode.go are UNCHANGED between the two bases;
  internal/cli/update.go carries a 57-line change (reconciliation preview
  rename, migrateProjectCommonAssets removal-arm relocation into the sync
  flow gated on userAssetsInstalled, participation step added to the skip
  block) — none of it touches the SPEC's conclusions; the skip-path block
  was re-read byte-identical and still carries NO integrity probe.
- Plan-phase measurements recorded: acceptance.md evidence ledger EV-1
  through EV-7 (all read-only, all run in this worktree; EV-6 is the
  re-verification batch on the re-cut baseline, including the
  normalization guard re-run — `ok ... 2.866s`, exit 0 on 2aab5f797; EV-7
  extends the anchors to f569be5d8 and records the EV-5 grep's actual
  execution). Authoring-discipline correction recorded at EV-5: one
  ledger row was initially written before its command ran; it has been
  re-measured and the correction is stated in the entry itself.
- Scope decisions: card item (3) out-of-scope with rationale (spec.md C.1);
  card item (4) in-scope as M3 (spec.md C.2); both recorded in
  decision-index.md (Q3 evidence-needed; Q1/Q2 implementation-level
  defaults applied at plan close).
- Known-issue classification: K1 (card P1) and K2 (card P2) verified as
  already repaired upstream on this tree — K1 by t1569 M2 / PR #1792
  (measured: normalization guard green), K2 by SPEC-USER-ASSET-INSTALL-001
  mechanism retirement. The only new implementation is the version-match
  integrity probe (K3). Branch contingency: if run-phase M1-b measures a
  live empty-directory producer, M2 escalates to repair per spec.md R1.
- Plan status: audit-ready.
audit_ready: true

## §E.2 Run-phase Evidence

Attribution for every entry: the command, its verbatim output, and the
baseline (this run, this tree, HEAD SHA at capture). The M1 capture HEAD is
`008d2e2a7` (branch `WT-update-migration-fixes`, worktree
`.moai/worktrees/t1578`).

### Pre-flight (SPEC §C, run 2026-10-10, HEAD 008d2e2a7)

- `git rev-parse --show-toplevel` → `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1578` (matches the spawn value)
- `git branch --show-current` → `WT-update-migration-fixes`
- `git rev-parse --short HEAD` → `008d2e2a7`
- `go build ./...` → exit 0, no output
- `GOOS=windows GOARCH=amd64 go build ./...` → exit 0, no output
- `golangci-lint run --timeout=2m 2>&1 | tail -5` → `0 issues.` (baseline: zero findings before any edit)
- Working tree before this run: one modified file, `progress.md`. Its only
  change was one appended §G line (ceiling-outcome record, 2026-10-09T16:19:39Z)
  written before this run. It is kept byte-identical and is committed with M1.

### M1 — measurement record (no implementation code)

**M1-a — K1 reproduction attempt (SPEC §C C3 command, re-measured on 008d2e2a7).**

```
$ go test ./internal/cli/ -run 'TestRunUpdate_V3Path_NormalizesLegacyRootDenyEntries' -count=1
ok  	github.com/modu-ai/moai-adk/internal/cli	1.337s
```

Exit 0. Classification: **K1 repaired** by card t1569 M2 (PR #1792, commit 87da06367). The guard is GREEN on this tree, so no migration code change is authorized.

Pin-strength read (M2-a, read-only): the test-owned `legacyRootDenySpecifiers` list has 9 entries (`internal/cli/update_deny_migration_test.go:423-433`). The fixture `legacyRootDenyFixture` carries all 9 plus two user-custom rules. The test asserts that every legacy form is absent, every canonical form is present, the custom rules survive, and the `[settings] Normalized` line is printed. The pin is at least as strong as the plan describes. No code change is needed for M2-a.

**M1-b — K2 reproduction attempt (fixture measurements, temporary probe, not committed).**

The probe was a temporary `internal/cli/zz_m1b_measure_test.go`, deleted before this commit. Command: `go test ./internal/cli/ -run 'TestM1B_' -count=1 -v`.

Project side. A full file-level template sync runs on a v3 fixture through the real `runUpdate` (`runUpdateInFixture`):

```
M1B-PROJECT update output lines=45
M1B-PROJECT root-absent: .claude/skills (stat .../.claude/skills: no such file or directory)
M1B-PROJECT root-absent: .claude/agents/moai (stat .../.claude/agents/moai: no such file or directory)
M1B-PROJECT root-absent: .agents/skills (stat .../.agents/skills: no such file or directory)
M1B-PROJECT root-absent: .codex/agents/moai (stat .../.codex/agents/moai: no such file or directory)
(Elision: the temporary sandbox fixture path prefix is shown as `...` in the four stat lines above; every other character is verbatim.)
M1B-PROJECT RESULT swept=0 zero_file=0
--- PASS: TestM1B_ProjectSideSyncFixture (0.90s)
```

Installer side. The production installer (`newUserAssetInstaller(tmpHome)`, which is the same constructor `runUserAssetUpdatePhase` uses) installs L0 plus all 11 optional packs into a temp user home:

```
M1B-INSTALL packs=11 installed=686 refreshed=0 failures=0 collisions=0 divergences=0
M1B-INSTALL RESULT catalog_skill_targets_checked=116 empty_targets=0 missing_targets=0
M1B-INSTALL SWEEP user_root_dirs=208 zero_file=0
--- PASS: TestM1B_InstallerSideCatalogSelection (0.71s)
```

Classification: **K2 mechanism retired on this tree.** The project payload carries no managed skill or agent root (`swept=0` is the by-design exclusion from `isCommonAssetRoot`). The installer produced no empty directory target under the widest selection. Branch decision: **"mechanism retired; guard to be pinned in M2."** Contingency R1 (escalate M2 to repair) is NOT triggered. The REQ-UMF-006 verify-and-report is still implemented unconditionally in M2, as the plan requires.

**M1-c — K3 RED anchor (EV-4 command re-run on 008d2e2a7).**

```
$ sed -n '/if syncSkipped {/,/^\t}$/p' internal/cli/update.go
	if syncSkipped {
		// A version-matched update runs no sync and no merge, so the retired
		// per-agent model/effort keys are stripped here, after its own backup.
		// A user-cancelled merge returns the same skipped=true; the helper
		// re-evaluates the version predicate and leaves that case untouched.
		if err := stripRetiredModelConfigOnVersionMatch(cmd, out, "."); err != nil {
			updateLedger.requiref(sevWarn, "retired model-key removal failed: %v", err)
		}
		// Card t1527 D5 + repair round: the deferred render carries the block
		// on this early return too — no explicit call here.
		// SPEC-FEEDBACK-PARTICIPATION-001 (REQ-ANON-004): a version-matched
		// update is still a finished plain template-sync run, so the ask runs
		// here too; its own gates (mode flags, terminal, CI, asked) decide
		// whether anything prompts. A failure warns; it never fails the update.
		if err := runParticipationStep(cmd, out); err != nil {
			_, _ = fmt.Fprintln(out, tui.CheckLine("warn", "Participation ask", "failed", err.Error(), &th))
		}
		runParticipationFlushAtUpdate(cmd.ErrOrStderr())
		return nil
	}
```

Exit 0. The block holds no integrity probe call. **RED anchor for AC-UMF-001 and AC-UMF-002 confirmed on 008d2e2a7.**

**B2 cross-SPEC conflict scan (run before any edit).**

`grep -rn "Retired\|superseded" internal/cli internal/userassets` returns 228 lines across 84 files. The matches in the files this run touches (`update.go`, `update_deny_migration*.go`, `update_model_key_strip.go`, `update_template_sync.go`, `user_asset_phase.go`, `install.go`, `installer_test.go`) concern the retired v2 deny-rule strip, the retired model-key strip, and superseded SPEC references. None reverses this SPEC. **No reversal to record.**

**C2 — production evidence preserved.** `cp /tmp/moaikr-force-update.log .moai/reports/t1578/moaikr-force-update.log`. Both files have sha256 `c1c17110f013d69a398270f0686083e58963464e7a7f1816f0cdd965759edc47`. The copy sits under the gitignored `.moai/reports/*` path (`.gitignore:235`), so it is a local, uncommitted copy.

**M1 findings.** K1 is repaired and guarded. K2 is retired, with the guard pinned in M2 and no repair escalation. K3's RED anchor is confirmed. M2 proceeds as the regression-guard milestone.

### M2–M4

_pending_

## §E.3 Run-phase Audit-Ready Signal

_pending run-phase_

## §E.3 Run-phase Audit-Ready Signal

_pending run-phase_

## §E.4 Sync-phase Audit-Ready Signal

_pending sync-phase_

## §F Phase 4 Mode Selection

Plan→run Kickoff decision record (autonomous transition, auto-semantics §9.1):

```text
decision record: decided_by=lane-27 orchestrator (card t1578) evidence_refs=.moai/reports/t1578/plan-audit-iter3.md (verdict PASS, overall 0.94 >= Tier M 0.80, blocking 0, convergence pass, codex pass, audited_sha 817b5b86f) + plan-audit-iter2.md (rcpt-5bb5f28c03f4602d0879cd46) ladder_path=autonomous-kickoff §9.1 (verdict PASS + score >= threshold + artifact-hash unchanged on the Go ComputeHash subject set + no blocker open)
```

Mode selection inputs: tier=M; scope≈6 files (probe + guard tests + install.go handler); domain count=1 (Go CLI internal/cli + internal/userassets); file language mix=Go + SPEC artifacts; concurrency benefit=LOW (coding-heavy); Agent Teams prereqs=not requested.

| Mode | Selected | Rationale |
|------|----------|-----------|
| direct | no | semantic multi-file change, not a typo fix |
| serial | **yes** | coding-heavy Go implementation (Anthropic coding-task caveat) |
| fanout | no | research-heavy work only; single domain, single writer |
| sweep | no | not a mechanical-uniform ≥30-file transform |

Decision: **serial** (one manager-develop spawn, milestones M1→M4 in sequence).

Justification: the implementation is coding-heavy Go work in one subsystem family (update path + userassets installer); per Anthropic's coding-task parallelism caveat the sequential single-agent path is the safe default, and the write contract is one writer per tree (this worktree). sweep/fanout offer no concurrency benefit here; direct is below the semantic-change bar.

Boundary cases: none — all four mode criteria resolved unambiguously.


## §G Override and Refusal Record

- 2026-10-09T13:00:28Z SPEC-UPDATE-MIGRATION-FIX-001 ceiling-refusal outcome=hold reasons="plan-audit ceiling reached (round count 3 >= tier ceiling 2); the verdict matches no admitting arm and holds, entry blocked (REQ-ACE-006) — release path: the split/new-SPEC route of REQ-ACE-005 or an operator decision recorded in progress.md §G" evidence=/Users/goos/moai/moai-adk-go/.moai/worktrees/t1578/.moai/reports/t1578/plan-audit-iter1.md,/Users/goos/moai/moai-adk-go/.moai/worktrees/t1578/.moai/reports/t1578/plan-audit-iter2.md,/Users/goos/moai/moai-adk-go/.moai/worktrees/t1578/.moai/reports/t1578/plan-audit-iter3.md
- 2026-10-09T13:05:26Z SPEC-UPDATE-MIGRATION-FIX-001 ceiling-refusal outcome=hold reasons="plan-audit ceiling reached (round count 3 >= tier ceiling 2); the verdict matches no admitting arm and holds, entry blocked (REQ-ACE-006) — release path: the split/new-SPEC route of REQ-ACE-005 or an operator decision recorded in progress.md §G" evidence=/Users/goos/moai/moai-adk-go/.moai/worktrees/t1578/.moai/reports/t1578/plan-audit-iter1.md,/Users/goos/moai/moai-adk-go/.moai/worktrees/t1578/.moai/reports/t1578/plan-audit-iter2.md,/Users/goos/moai/moai-adk-go/.moai/worktrees/t1578/.moai/reports/t1578/plan-audit-iter3.md
- 2026-10-09T13:11:28Z SPEC-UPDATE-MIGRATION-FIX-001 ceiling-outcome outcome=pass-through reasons="plan-audit ceiling reached (round count 3 >= tier ceiling 2 + 1 delta rounds); the verdict is admission-clean and admits without a question (REQ-ACE-013)" evidence=/Users/goos/moai/moai-adk-go/.moai/worktrees/t1578/.moai/reports/t1578/plan-audit-iter1.md,/Users/goos/moai/moai-adk-go/.moai/worktrees/t1578/.moai/reports/t1578/plan-audit-iter2.md,/Users/goos/moai/moai-adk-go/.moai/worktrees/t1578/.moai/reports/t1578/plan-audit-iter3.md
- 2026-10-09T13:12:37Z SPEC-UPDATE-MIGRATION-FIX-001 ceiling-outcome outcome=pass-through reasons="plan-audit ceiling reached (round count 3 >= tier ceiling 2 + 1 delta rounds); the verdict is admission-clean and admits without a question (REQ-ACE-013)" evidence=/Users/goos/moai/moai-adk-go/.moai/worktrees/t1578/.moai/reports/t1578/plan-audit-iter1.md,/Users/goos/moai/moai-adk-go/.moai/worktrees/t1578/.moai/reports/t1578/plan-audit-iter2.md,/Users/goos/moai/moai-adk-go/.moai/worktrees/t1578/.moai/reports/t1578/plan-audit-iter3.md
- 2026-10-09T16:19:39Z SPEC-UPDATE-MIGRATION-FIX-001 ceiling-outcome outcome=pass-through reasons="plan-audit ceiling reached (round count 3 >= tier ceiling 2 + 1 delta rounds); the verdict is admission-clean and admits without a question (REQ-ACE-013)" evidence=/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1578/.moai/reports/t1578/plan-audit-iter1.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1578/.moai/reports/t1578/plan-audit-iter2.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1578/.moai/reports/t1578/plan-audit-iter3.md
