# SPEC-SESSION-CC-VERSION-002 — Progress

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
plan_complete_at: 2026-10-05
plan-audit: PASS 0.96 (iteration 2/2 — Tier M ceiling, blocking 0). Verdict file
`.moai/reports/t1515/plan-audit.md`; audited_sha 099250516768f7b6850e276a237620c245d5c4ef;
plan_artifact_hash ec087536e430c042e510c37d751338d684affef5deebdb30b731b2a7c616dbb8.
History: iteration 1 FAIL 0.91 (D1-D3 blocking, D4-D7 optional) → one repair round (D1-D7
applied by the author, all verified by the auditor's own re-reading) → delta re-audit PASS.
Auditor's named lint gap closed by the lane's own run: `go run ./cmd/moai spec lint
.moai/specs/SPEC-SESSION-CC-VERSION-002` → "No findings" (2026-10-05, this tree); the
auditor's background lint file `/tmp/t1515-lint-iter2.txt` landed empty (incomplete run) —
superseded by the observation above.

Post-audit addendum (lane, 2026-10-06): the verdict was revised twice after iteration 2 —
the auditor's hook-forced codex verification found P2-C (AC-SCV-014's literal version pin,
a two-way mutant); at the Tier M ceiling the leader approved a hunk-limited exception
(recorded in the verdict file and `.moai/reports/t1515/lane-note-kickoff-revision.md`); the
single-hunk repair (observation-based provenance) reread to the FINAL verdict:
`plan-audit: PASS 0.96`, blocking 0, plan_artifact_hash
`8e3ac1a906b4a6e477d85b42fe09620e073789ea219b510a06a35e81aea58ac8`. The iteration-2 hash
citation above (`ec087536…`) is the superseded intermediate; the Kickoff that ran on it was
validated post-hoc — the final verdict passes on the same evidence set.

## §E.2 Run-phase Evidence

Run-phase tree: `WT-session-cc-version`, base `099250516` → HEAD `dac6ceef2` (6 commits, no
push — the lane reports the merge SHA). All measurements below are this run, this tree.

**M1 — the claude option model and its derivation seam** (commit `9dda7e867`). New
`internal/cli/lane_resume_model.go`: three measured classes (required-value / optional-value /
boolean / unknown-as-zero), the compile-time snapshot converted from the predecessor's tables
with source-version provenance (Claude Code 2.1.289, 2026-10-04), the commander-help synopsis
parser, the `claudeHelpSynopsis` probe seam (bounded 3s, procInfoFunc pattern), and the
apply/refresh silent-degradation path. SPEC frontmatter flipped draft → in-progress on the
same commit; the plan-phase artifacts landed unchanged with it. RED (AC-SCV-012, captured
before the implementation existed): `go test ./internal/cli/ -run
'^TestClaudeOptionModelDerivation$' -count=1 -v` → exit 1, `undefined:
activeClaudeOptionModel / parseClaudeOptionModel / claudeOptionModel /
claudeOptionRequiredValue ...` `[build failed]`. GREEN: the same command → 4/4 `--- PASS`.

**M2 — the two-segment interpretation** (commit `b71bb0209`). One `scanResumeArgs` walk in two
polarities; the three hand-synced maps dissolved (grep 9 → 0); the launcher surface survives
as repo-owned `launcherSegmentValueFlags`; the derivation refreshes once per process at the
two launcher-entry call sites (`runClaudeEntry`, `runFactoryLaneRelaunch`). Parser repair
found by the pinned suite: the real help's wrapped description continuations can BEGIN with a
flag mention ("--append-system-prompt included — sent,"), which silently reclassified the
option boolean; the synopsis segment is now gated on the option block's shallow indent
(`claudeHelpSynopsisMaxIndent` 6, measured layout two spaces). REDs (captured on the
pre-rework tree, unpiped): `go test ./internal/cli/ -run '^TestOptionModelPolarityDefaults$'
-count=1 -v` → exit 1, `--- FAIL .../guard,_required-value_option,_does_not_fire` ("the guard
judged the model-known required value"), 3 cells coinciding with old behavior; and
`'^TestOptionModelResidualCompound$'` → exit 1, both cells ("the stripped option shielded a
resume token — a silently leaked resume" / "expected the compound-condition false refusal to
be observed"). GREEN: the 19-function pinned suite 19/19 top-level `--- PASS` (38 subtest
lines), `git diff --name-only 099250516 -- internal/cli/lane_resume_test.go` → empty.

**M3 — install-root-anchored version extraction** (commit `cfb2425fe`).
`versionSegmentFromPath` anchors on the `mappingPathNamesClaudeBinary` product-directory
shapes (`…/claude/versions/<v>`, `…/claude-code/<v>`, trailing `…/<v>/claude`); an unrelated
prefix can no longer satisfy the read; the ccversion.go DEBT marker set leaves with the
rework. RED (AC-SCV-016 re-capture, unpiped, pre-implementation tree, single invocation):
`go test ./internal/session/ -run '^TestVersionSegmentAnchoredToInstallRoot$' -count=1 -v` →
exit 1, `versionSegmentFromPath("/opt/versions/9/tools/claude/versions/2.1.281") = "9", want
"2.1.281"` (plus the `other/versions/3.0.0` → "" and `claude-code/2.1.284` → "9" shapes).
GREEN: the anchored table + all four degradation/mapping preservation tests `--- PASS`.

**M4 — live re-measure and marker retirement** (commit `45242e053`). The installed claude was
re-measured in this run (not carried forward): `command -v claude` →
`/Users/goos/.local/bin/claude`, realpath `/Users/goos/.local/share/claude/versions/2.1.289`;
`claude --version` → `2.1.289 (Claude Code)`; `claude --help` exit 0 in **0.098s** (inside the
3s derivation bound). **Live-measured version token: 2.1.289** — the snapshot's
source-version provenance comment carries the same token (AC-SCV-014's equality), refreshed
with the re-measure date (2026-10-05). All 43 snapshot entries reproduce from the live-derived
model with zero diffs; `--remote-control-session-name-prefix <prefix>` joins the required-value
class from the measure (r5 P2①); `TestRemoteControlPrefixValue` pins the instance (a prefix
value literally reading `--resume` passes the validator, the guard does not fire). Both
discharged DEBT marker sets read 0 (`@MX:UPGRADE: t1515` in lane_resume.go left with the M2
map deletion, in ccversion.go with the M3 rework). The plan §A.4 residual survives by design
(the snapshot fallback remains) and carries the fresh marker on the snapshot:
`@MX:DEBT` + `@MX:CEILING` + `@MX:UPGRADE` carrying the compound condition verbatim
(derivation down + option snapshot-absent + resume-shaped value → validator false refusal;
never a silently leaked resume under relaunch).

**M5 — verification batch** (commits `c190de078`, `dac6ceef2`). The full-suite gate caught two
isolation defects in the M1-era model tests, repaired without production-code changes: (1) the
fixture help carried an invented 4/8-space indentation the M2 synopsis-indent gate correctly
rejects — replaced with the real two-space layout (the gate round's own derivation-test re-run
gap closed by this repair); (2) the degradation subtests assumed the ambient active model
equals the snapshot — in the full suite an earlier entry-path test's once-per-process
derivation has already swapped it — they now pin the starting state they assert against.
golangci-lint on the changed packages: pre-edit baseline 0 issues; after the work 1 NEW
finding (`claudeOptionUnknown` unused) → fixed by making the boolean/unknown polarity an
explicit case in the walk; 0 issues at HEAD `dac6ceef2`. The lane's two style suggestions
(`strings.SplitSeq`, `maps.Copy`) were not flagged by golangci-lint and stay advisory-only.

**Scoped verification (env-scrubbed, `unset MOAI_KANBAN_ID MOAI_KANBAN_LEAD_ADDR
MOAI_KANBAN_SETTINGS_INJECTED &&` prefixed throughout):**

| Gate | Command (at HEAD `dac6ceef2` unless noted) | Result |
|---|---|---|
| AC-SCV-011 | `-run '^(TestResumeScanStopsAtClaudeSeparator\|TestSeparatorWinsOverAmbiguousValue\|TestSeparatorInterplaySkipsValues\|TestPostSeparatorLauncherFlagsAreInert\|TestPostSeparatorLauncherFlagsInertGuard\|TestPreSeparatorLauncherFlagsKeepValueBehavior)$' -count=1 -v` | 6/6 `--- PASS` |
| AC-SCV-012 | `-run '^TestClaudeOptionModelDerivation$' -v` + `grep -n "exec.Command" internal/cli/lane_resume_model_test.go internal/cli/lane_resume_test.go` | `--- PASS` 4/4 subtests; grep exit 1 (0 hits) |
| AC-SCV-013 | `-run '^TestOptionModelPolarityDefaults$' -v` | 4/4 subtests `--- PASS` |
| AC-SCV-014 | `grep -rn "claudeValueTakingOptions\|launcherValueTakingOptions\|ambiguousValueOptions" internal/cli/ --include="*.go"`; `grep -rn "@MX:UPGRADE: t1515" internal/cli/lane_resume.go internal/session/ccversion.go`; `grep -n "source-version provenance" internal/cli/lane_resume_model.go` | 0 hits; 0 hits; line 80 hit, token 2.1.289 = the §E.2 live-measured token; residual cells in `TestOptionModelResidualCompound` 2/2 PASS |
| AC-SCV-015 | the 19-function selector `-count=1 -v`; `git diff --name-only 099250516 -- internal/cli/lane_resume_test.go`; `-run '^TestRemoteControlPrefixValue$' -v` | 19/19 `--- PASS` (zero `--- FAIL`); empty (byte-identical); `--- PASS` |
| AC-SCV-016 | `go test ./internal/session/ -run '^TestVersionSegmentAnchoredToInstallRoot$' -v` | `--- PASS` (5 shapes incl. the overlay repro) |
| AC-SCV-017 | `go test -timeout 30m ./internal/session/ -run '^(TestVersionDegradationRendersUnknown\|TestRunningVersionFromDeletedBinary\|TestRunningVersionFromInjectedMapping\|TestInstalledVersionFromResolvedPath)$' -count=1 -v` | 4/4 `--- PASS` |
| E2 builds | `go build ./...`; `GOOS=windows GOARCH=amd64 go build ./...`; `go vet ./internal/cli/... ./internal/session/...` | exit 0 / exit 0 / clean |
| E4 boundary | `grep -rn 'AskUserQuestion' <changed files>` ; base-vs-now filtered package count | 0 hits on changed files; base 44 = now 44 (zero NEW; the pre-existing 44 are doc-string mentions and the agentlint linter's own identifiers) |
| E5 lint | `golangci-lint run --timeout=4m ./internal/cli/... ./internal/session/...` | 0 issues at HEAD (baseline 0 pre-edit; 1 NEW found and fixed in-run) |
| E3 coverage | `go test -timeout 60m -cover ./internal/cli/ ./internal/session/` | `ok internal/cli 2820.425s coverage: 85.0%`; `ok internal/session 8.520s coverage: 86.0%` (profile: `lane_resume.go` all functions 100%; `lane_resume_model.go` parse/apply/segment 100%, refresh 80% — the sync.Once LookPath-fail branch; `cc.go` runClaudeEntry 92.9%; `factory_lane_relaunch.go` entry 100%; `ccversion.go` versionSegmentFromPath 100%, installedCCVersion 77.8% pre-existing symlink-error branch) |
| §C scoped gate | `go test -timeout 60m -cover -coverprofile ./internal/cli/ ./internal/session/ -count=1` (the 30m form's deviation recorded below) | exit 0, zero `--- FAIL`, both packages `ok` |

**Scoped-run wall-time note (measurement environment, not a test failure).** The full
`./internal/cli` suite could not complete inside the 30m gate on this machine during this
run: attempt 1 exit FAIL at exactly 1800.9s (this attempt additionally surfaced the two model
test isolation defects — the repair above), attempt 2 exit FAIL at exactly 1800.9s with **zero
`--- FAIL` test results** (the timeout panic dumped `TestInitPluginStep_OptOut` still running),
attempt 3 (60m ceiling, `-cover`) was launched into a machine load storm (load average 517 —
dozens of concurrent test binaries from other lanes, the same multi-lane-suite shape as the
2026-08-15 load-413 incident) and was withdrawn by this agent rather than adding to the
storm. Attempt 4 — launched only after a load monitor read the machine back under 10, under
the `test-internal-cli-suite` slot lease, 60m ceiling — completed green (2820.4s, exit 0,
zero `--- FAIL`, both packages `ok`). The 30m gate form itself is recorded as a deviation: on
a quiet machine the package's known runtime (worst-case 1118s single-package baseline,
t1253; CI's dedicated `-timeout 25m`, card t1519) fits, but under multi-lane load it does
not; CI remains the full-suite verdict surface per the gitflow lane protocol §8. **No gap
remains: the scoped gate and E3 are both measured green at HEAD `dac6ceef2`.**

**Post-close repairs addendum (card-review r1, 2026-10-06, commit `42ed25265`).** Three
findings, each RED-first then fixed: (P1) `parseClaudeOptionModel` now classifies FIRST-WINS
per flag — a later flag-prefixed prose line (even at the shallow indent) cannot reclassify an
option whose synopsis already classified it; RED `TestClaudeHelpReferenceLinesNeverOverwrite`
exit 1 ("--print classified 1, want boolean", the guard missed a resume behind --print), GREEN
same test + the derivation fixture. (P2①) the derivation probe sets `cmd.WaitDelay` (1s) —
exec's default of zero waits for stdout pipe EOF after Cancel, so a wrapper exiting while its
own backgrounded child holds the pipe blocked launcher entry; RED
`TestClaudeHelpSynopsisBoundedUnderLingeringChild` (new sibling `lane_resume_probe_test.go`,
throwaway fixture script, never claude — the derivation test files named by the AC-SCV-012
no-spawn grep stay free of process references) exit 1 "returned after 27.24s", GREEN 1.2s.
(P2②) `versionSegmentFromPath` judges the product-directory shapes in PREFERENCE order
(native …/claude/versions/<v> over npm …/claude-code/<v>, and within a shape the candidate
closest to the path end), so `/opt/claude-code/9/tools/claude/versions/2.1.281` reads
2.1.281; RED exit 1 `= "9"`, GREEN all six anchored shapes + four preservation tests.

**Post-close scoped re-measurement (pending item).** The post-repair two-package run launched
into a fresh load storm (load 38-60): `./internal/session` completed green (`ok 20.893s,
coverage 86.1%`); `./internal/cli` hit the 60m wall (3601.4s) with exactly two `--- FAIL` —
`TestFactoryLeaseRecordStallBounded` and `TestFactoryLeaseDriftLogVerbWorktreeWriteWaits`,
both wall-clock timing-bound tests of the F1 lease machinery. Attribution: `git log
099250516..HEAD` over the factory_lease files is EMPTY — no commit on this branch touched
them; both assert sub-second timing bounds that the load storm violates (one passed in
isolation, one persists at load ~43). PENDING: one package-level re-run in a quiet window
(load < 15) for the clean local gate; CI remains the verdict surface. All repair-scoped
families are green at `42ed25265` (the six model/repair tests, r5 instance, polarity and
residual matrix; lint 0 issues; both builds exit 0).

## §E.3 Run-phase Audit-Ready Signal

run_status: audit-ready
run_complete_at: 2026-10-06
run HEAD: `dac6ceef2` (base `099250516`; 6 commits, unpushed — the lane reports the merge
SHA). E1-E6 complete and green: the seven-AC matrix of §E.2 all PASS with command + verbatim
output at HEAD; both cross-platform builds exit 0; coverage 85.0% (cli) / 86.0% (session),
every new/changed function ≥ 80% and the walk functions 100%; boundary grep zero NEW (changed
files 0 hits); golangci-lint 0 issues at HEAD (1 NEW found in-run and fixed); PRESERVE
byte-identity held (`lane_resume_test.go` diff vs base empty; changed-path scope exactly
`internal/cli/` + `internal/session/` + this SPEC directory). The two M1-era model-test
isolation defects were caught by the full-suite gate and repaired in-run (commits
`c190de078`, `dac6ceef2`); their RED-then-GREEN observation is the §E.2 M5 record. The r5
instance is pinned by `TestRemoteControlPrefixValue`; the §A.4 residual is measured by
`TestOptionModelResidualCompound` and carries its fresh marker.

## §E.4 Sync-phase Audit-Ready Signal

sync_status: complete
sync_commit_sha: 5a5d179f5   # post-close convergence (t1465 패턴) — card-review r1~r4 전 라운드 수리 트리의 최종 커밋; 원 close 커밋 6f646a29d (D3 backfill). card-review r1 FAIL(P1 1+P2 2)→수리→r2 FAIL(테스트 P2)→수리→r3 FAIL(테스트 P2, 상한)→리더 ① 마이크로 수리+사전확정 조건→r4 PASS(신규 0). 전체 이력 .moai/reports/t1515/card-review.md
sync_complete_at: 2026-10-06
b12_self_test_a: pass   # grep -c 'SPEC-SESSION-CC-VERSION-002' CHANGELOG.md → 0 before emission (duplicate-free), 1 after (exactly the entry added)
b12_self_test_c: pass   # every file path named in the CHANGELOG entry verified present: internal/cli/lane_resume.go, internal/cli/lane_resume_model.go, internal/cli/cc.go, internal/cli/factory_lane_relaunch.go, internal/session/ccversion.go
changelog_entry_position: "[Unreleased] → Fixed, 첫 번째 불릿"
frontmatter_status_transitions:
  spec_md: "in-progress → completed (sync 커밋에서 적용)"
  plan_md: unchanged
  acceptance_md: unchanged
  progress_md: §E.4 블록 + 레인 §E.1 post-audit addendum 동승

## §F Phase 4 Mode Selection

- Inputs: tier M; scope ~6 files (2 source + 2 test + SPEC frontmatter/progress); 1 domain
  (Go, internal/cli + internal/session); coding-heavy; one writer per card tree.
- serial: selected — coding-heavy single-domain implementation (Anthropic coding-task
  caveat); one write-capable delegation in the lane's card worktree.
- fanout: not selected — no multi-domain research fan-out warranted.
- sweep: not selected — semantic new-code work, not mechanical-uniform; well under file floor.
- direct: not selected — multi-file TDD implementation beyond trivial.
- Decision: serial

decision record: decided_by=claude-lane-t1515 (plan→run Kickoff, autonomous form per auto-semantics §9.1) evidence_refs=.moai/reports/t1515/plan-audit.md PASS 0.96 iter2 blocking-0 + plan_artifact_hash ec087536e430c042e510c37d751338d684affef5deebdb30b731b2a7c616dbb8 unchanged (lane-recomputed, byte-identical) + spec-lint no-findings (lane-run) + codex cross-model findings all resolved ladder_path=plan-run-kickoff-autonomous
