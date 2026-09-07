# SPEC-CODEX-STALE-SPLIT-FOURTH-001 — progress

card: t534 · worktree `.claude/worktrees/t534` · branch `WT-stale-msg-polarity`
tree pin at plan-phase: **`e0c904f58`**

## §E.1 Plan-phase Audit-Ready Signal

Plan-phase artifacts authored: `spec.md`, `plan.md`, `acceptance.md`, `research.md`, `progress.md`.

- Tier S declared; 8 requirements / 8 acceptance criteria, both at the ceiling, neither over.
- SPEC ID regex self-check executed as Bash: `PASS`.
- Frontmatter carries all 12 canonical fields plus `era`, `tier`, `related_specs`.
- Open decision settled at plan-phase (not deferred to run): the fourth counter renders
  **unconditionally** — spec.md §B.1, pinned by AC-SSF-004.
- Partial supersession of `REQ-CEF-010` / `AC-CEF-011` recorded with its preserved scope — spec.md §B.2.
- `[NEEDS CLARIFICATION]` markers: **none**.
- Correction against the card brief, carried forward: the broken live assertions are three
  (`doctor_codex_test.go:279`, `:349`, `:943`), not the single `(1 enabled, 0 disabled, 1
  unspecified)` string named in the brief — that string exists only in a comment. Measured in
  research.md §4.

- Audit provenance: **single-backend Claude audit, and that is the configured path** — not a gap.
  `grep -rn 'audit_model' .moai/config/sections/*.yaml` returns no output (exit 1) in this tree, so
  no `audit_model` is configured and the cross-model fan-out (`audit_multi`) is not the applicable
  entry point. Closed, not open.
- Plan-audit verdict 0.88 against the Tier S threshold 0.75 — **PASS**. The four blocking findings
  (D1 no-`default:` clause unverifiable, D2 AC-SSF-004's Verify not observing its Then, D3 M2's RED
  obligation unbound, D4 AC-SSF-004's RED-now cell + §D.2 collapsed sub-classes) are applied in
  acceptance.md; requirement and AC counts are unchanged at 8 / 8.

Status: `draft`. Awaiting plan-audit and Implementation Kickoff Approval.

## §E.2 Run-phase Evidence

**Tree-pin restamp.** Run-phase HEAD at entry was **`9575e8843`**, not the plan-phase pin
`e0c904f58`. The restamp is legitimate and the baselines carry over, because the two files under
change are byte-identical between the two commits — `git diff --stat e0c904f58..HEAD --
internal/cli/doctor_codex.go internal/cli/doctor_codex_test.go` produced NO output (exit 0). The
only commit between them (`9575e8843`) is the plan-phase artifact commit. Every cell below is
therefore attributed to `9575e8843`.

### AC-SSF-001 RED (captured BEFORE the M1 render change — §D.4 obligation)

Command, at `9575e8843` with only the M2 guard test added and `doctor_codex.go` untouched
(`grep -c 'non-boolean' internal/cli/doctor_codex.go` → `0`, exit 1):

```
$ go test ./internal/cli/ -run 'TestCheckCodexWiring_NonBooleanEnabledCountedSeparatelyInStaleSplit' -count=1 -v -timeout 1800s
exit=1
```

Verbatim output (trimmed to the two failure blocks and the verdict lines; temp paths shortened):

```
=== RUN   TestCheckCodexWiring_NonBooleanEnabledCountedSeparatelyInStaleSplit
=== RUN   .../non-boolean_leaves_unspecified_empty,_bare_booleans_unmoved
    doctor_codex_test.go:386: declared split does not carry the non-boolean entry in its own bucket:
      "…/.codex/config.toml declares 4 [[skills.config]] entries; 4 with a path that no longer
       exists (1 enabled, 2 disabled, 1 unspecified) — remove the stale entries or restore the
       skill files …; 1 declare `enabled` with a value that is not a bare TOML boolean — …"
=== RUN   .../absent_key_stays_unspecified_alongside_a_non-boolean
    doctor_codex_test.go:410: absent key and non-boolean value did not separate:
      "…declares 2 [[skills.config]] entries; 2 with a path that no longer exists
       (0 enabled, 0 disabled, 2 unspecified) …; 1 declare no `enabled` key; 1 declare `enabled`
       with a value that is not a bare TOML boolean — …"
--- FAIL: TestCheckCodexWiring_NonBooleanEnabledCountedSeparatelyInStaleSplit (0.01s)
    --- FAIL: .../non-boolean_leaves_unspecified_empty,_bare_booleans_unmoved (0.00s)
    --- FAIL: .../absent_key_stays_unspecified_alongside_a_non-boolean (0.00s)
FAIL	github.com/modu-ai/moai-adk/internal/cli	0.873s
```

The RED is red for the stated reason: both lines in the first block come from ONE run — the same
entry is `1 unspecified` in the advisory and `not a bare TOML boolean` in the fatal finding. That
is the two-names-in-one-run defect, observed rather than inferred. RED commit: `a9c9a275d`
(guard test only, production file untouched).

### AC matrix

| AC | Class | Status | Command | Observed output (verbatim, trimmed) |
|---|---|---|---|---|
| AC-SSF-001 | regression guard | **PASS** | `go test ./internal/cli/ -run 'TestCheckCodexWiring_NonBooleanEnabledCountedSeparatelyInStaleSplit' -count=1 -v -timeout 1800s` (exit 0) | `--- PASS: TestCheckCodexWiring_NonBooleanEnabledCountedSeparatelyInStaleSplit (0.07s)` + both sub-tests PASS; swept-count 1/1 named test |
| AC-SSF-002 | MUST-PASS control | **PASS** | `go test ./internal/cli/ -run 'TestCheckCodexWiring_UnspecifiedEnabledReportedSeparately' -count=1 -v -timeout 1800s` (exit 0) | `--- PASS: TestCheckCodexWiring_UnspecifiedEnabledReportedSeparately (0.00s)`; swept-count 1/1. Asserts `(0 enabled, 0 disabled, 1 unspecified, 1 non-boolean)` — the absent-key entry stayed in `unspecified` |
| AC-SSF-003 | MUST-PASS control | **PASS** | `go test ./internal/cli/ -run 'TestCheckCodexWiring_StaleHomeSkillsReported\|TestCodexSkillPath_AbsoluteExistingAndMissing' -count=1 -v -timeout 1800s` (exit 0) | `--- PASS: TestCheckCodexWiring_StaleHomeSkillsReported (0.00s)` + `--- PASS: TestCodexSkillPath_AbsoluteExistingAndMissing (0.00s)`; swept-count **2/2** as the criterion requires |
| AC-SSF-004 | MUST-PASS release-blocking | **PASS** | `go test ./internal/cli/ -run 'TestCheckCodexWiring_StaleHomeSkillsReported' -count=1 -v -timeout 1800s` (exit 0) | `--- PASS: TestCheckCodexWiring_StaleHomeSkillsReported (0.00s)`; swept-count 1/1. Line 279 now asserts the full four-member phrase `(1 enabled, 2 disabled, 0 unspecified, 0 non-boolean)`, so the Then is observed rather than merely coexisting. RED baseline: `grep -c 'non-boolean' internal/cli/doctor_codex.go` → `0`, exit 1, at `9575e8843` |
| AC-SSF-005 | MUST-PASS control | **PASS** | same command as AC-SSF-003 (exit 0) | Both tests keep their unedited leading-count (`3 stale skill entries` / `1 stale skill entry`, `N with a path that no longer exists`) and remove-directive assertions; the advisory grade (`uikit.CheckWarn`) assertion is likewise unedited |
| AC-SSF-006 | MUST-PASS diff criterion | **PASS** | `git diff --unified=0 e0c904f58..HEAD -- internal/cli/doctor_codex.go` | Every hunk header is `@@ … @@ func codexStaleSkillFinding()`; first hunk at line 821. No hunk touches `codexEnabledShapeFinding` (lines 755-790). The switch carries `case codexwiring.SkillEnabledNonBoolean:` and the surviving `default:` increments `missingUnknownState`, which is NOT one of the four named buckets |
| AC-SSF-007 | MUST-PASS diff criterion | **PASS** | `git diff e0c904f58..HEAD -- internal/cli/doctor_codex_test.go` | Three edited assertions (was lines 279 / 349 / 943), each preceded by a comment naming `SPEC-CODEX-STALE-SPLIT-FOURTH-001`; the t508 "deliberately does NOT grow a fourth bucket" paragraph is **rewritten** under a `PRIOR DECISION, REVERSED` heading quoting the original verbatim, not deleted |
| AC-SSF-008 | MUST-PASS diff criterion | **PASS** | `git diff -- internal/cli/ > /tmp/t534-diff-all.txt` then `grep -n '^+.*t\.Setenv("HOME"' /tmp/t534-diff-all.txt` → exit 1, no output (run against the working tree before the M1 commit; the same diff is now `e0c904f58..b0cb9318a`) | No `t.Setenv("HOME", …)` added. Both new fixtures go through `writeCodexHomeConfig` (`t.TempDir()`-backed) + `stubCodexHome` (pins `CODEX_HOME` and the `codexUserHomeDir` seam). No real `~/.codex` path is read or written |

### Suite / vet / scope

```
$ go test ./internal/cli/ -count=1 -timeout 1800s
exit=0
ok  	github.com/modu-ai/moai-adk/internal/cli	479.657s

$ go test ./internal/codexwiring/ -count=1 -timeout 600s
exit=0
ok  	github.com/modu-ai/moai-adk/internal/codexwiring	0.656s

$ go vet ./internal/cli/...
exit=0   (no output)

$ gofmt -l internal/cli/doctor_codex.go internal/cli/doctor_codex_test.go
(no output)

$ git diff --name-only e0c904f58..HEAD -- '*.go'
internal/cli/doctor_codex.go
internal/cli/doctor_codex_test.go
```

### Deviations from plan.md, recorded rather than smoothed over

1. **Milestone order was M2-RED → M1 → M3, not M1 → M2 → M3.** plan.md §F orders M1 first, but
   acceptance.md §D.4 requires AC-SSF-001's RED captured BEFORE the M1 render change. The two are
   in tension; §D.4 wins because a RED reconstructed after the change is not an observation. The
   guard was therefore authored and observed failing first (`a9c9a275d`), and M1 landed after
   (`b0cb9318a`).
2. **A fifth counter, `missingUnknownState`, exists in the production code.** plan.md §F M1
   permits a retained `default:` that panics or records an explicit unknown state where Go does
   not force exhaustiveness. This repository's `.golangci.yml` enables only
   `errcheck, govet, ineffassign, staticcheck, unused` — no `exhaustive` linter — so compile- or
   lint-time exhaustiveness is unavailable, and `internal/cli/CLAUDE.md` forbids `panic()`. The
   counter is included in the `missing` sum and surfaced by a clause that fires only when it is
   non-zero (unreachable while `SkillEnabled` has four states), mirroring the existing
   `indeterminate` clause's shape. It is NOT one of the four partition members and no `default:`
   arm increments a named bucket.
3. **M3's forward-pointer HISTORY row in `SPEC-CODEX-ENABLED-FATAL-001/spec.md` was SKIPPED.**
   plan.md §F M3 explicitly permits the skip and asks that it be recorded. The skip is taken on
   ownership grounds: manager-develop's artifact boundary forbids modifying another SPEC's body
   content. The supersession is load-bearing in this SPEC's §B.2 (REQ-SSF-007), which the forward
   pointer only duplicates.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-09-07
run_commit_sha: b0cb9318a          # M1 (final code commit); M2-RED = a9c9a275d
run_status: audit-ready
ac_pass_count: 8
ac_fail_count: 0
preserve_list_post_run_count: 0    # no PRESERVE-list file modified outside the two permitted files
l44_pre_commit_fetch: not-run      # lane-local worktree, no push in this card's run phase
l44_post_push_fetch: not-run       # no push performed (dispatch: do not push, do not open a PR)
new_warnings_or_lints_introduced: 0  # go vet ./internal/cli/... exit 0; gofmt clean
cross_platform_build:
  darwin_arm64: pass               # `go build ./...` exit 0 (no output); suite built and ran on the host
  windows_amd64: pass              # `GOOS=windows GOARCH=amd64 go build ./...` exit 0 (no output) — compile only; Windows tests are CI's verdict
total_run_phase_files: 2           # internal/cli/doctor_codex.go, internal/cli/doctor_codex_test.go
m1_to_mN_commit_strategy: "two commits — a9c9a275d (M2 RED guard, test only) then b0cb9318a (M1 render + bucket + assertion updates); M3 is documentation-only"
```

## §E.4 Sync-phase Audit-Ready Signal

**CHANGELOG.md.** Added under `### Fixed` in `[Unreleased]`, immediately after the
SPEC-CODEX-ENABLED-FATAL-001 entry (topically adjacent — same check, same file). B12 self-tests:
(1) pre-emission `grep -c 'SPEC-CODEX-STALE-SPLIT-FOURTH-001' CHANGELOG.md` → `1` (no prior
duplicate); (2) AC count `grep -oE 'AC-([A-Z0-9]+-)*[0-9]+' acceptance.md | sort -u | wc -l` → `8`,
matching the CHANGELOG's stated "8 acceptance criteria" and the 7 MUST-PASS / 1 regression-guard
split cited from acceptance.md §D.2 verbatim; (3) every file path cited in the entry
(`internal/cli/doctor_codex.go`, `.moai/reports/t534/{run-evidence,reproduction}.md`) verified via
`ls` before commit.

**README / docs-site.** Neither documents the stale-path advisory message or its bucket split.
Control: `grep -n "moai doctor" README.md` → 3 matches (the tool is documented broadly), so the
zero below is an observed absence, not an unreached scan.

```
$ grep -rn "with a path that no longer exists\|Codex Wiring" README.md README.ko.md README.ja.md README.zh.md
(no output, exit 1)
$ grep -rln "Codex Wiring\|codex.*wiring" docs-site/content
docs-site/content/en/advanced/codex-dual-harness.md
$ grep -n "enabled\|non-boolean\|unspecified\|stale" docs-site/content/en/advanced/codex-dual-harness.md
(no relevant match — the page covers the skills mirror, not the wiring diagnostic's message bodies)
```

The docs-site `codex-dual-harness.md` page exists and mentions "Codex Wiring" once, but not this
advisory's message content. This gap is **already owned**: card t535 covers the missing docs-site
Codex Wiring callout (4-locale). No docs-site edit made here — recorded, not actioned, per the
dispatch instruction to avoid duplicating t535's scope.

```yaml
sync_complete_at: 2026-09-07
sync_commit_sha: 1b027e311
sync_status: audit-ready
b12_self_test_a: PASS   # pre-emission grep = 1, no duplicate
b12_self_test_b: PASS   # AC count 8 == CHANGELOG claim (7 MUST-PASS + 1 regression guard)
b12_self_test_c: PASS   # all cited file paths verified via ls
changelog_entry_position: "### Fixed, immediately after SPEC-CODEX-ENABLED-FATAL-001"
frontmatter_status_transitions:
  spec_md: "in-progress -> completed (this sync commit)"
canary_compliance_check:
  applicable: false   # this SPEC defines no forward-looking policy that its own sync tests
```
