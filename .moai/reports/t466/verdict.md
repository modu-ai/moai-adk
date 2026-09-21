# sync-auditor verdict — SPEC-UPDATE-HOOK-DELIVERY-001 (card t466)

Auditor: sync-auditor (independent, fresh context) · Tree: WT-update-hook-delivery @ 940adf966 (worktree .claude/worktrees/t466, porcelain-clean at audit start)
Audit date: 2026-09-03 · Operator decision under audit: Option B (detect + guide, READ-ONLY)

## Overall Verdict: PASS-WITH-DEBT

Weighted score 89.4/100 (Functionality 90×0.40 + Security 95×0.25 + Craft 88×0.20 + Consistency 80×0.15).
Must-pass firewall (Functionality + Security): both pass independently — no firewall trip.
No blocking findings. The three debt items (F1–F3) are record-accuracy defects, none behavioral;
F1's user-facing half (CHANGELOG identity sentence) is the one item worth a follow-up.

## Dimension Scores

| Dimension | Score | Verdict | Justification |
|-----------|-------|---------|---------------|
| Functionality (40%) | 90/100 | PASS | All 10 binding ACs re-verified by the auditor's own runs (11 unit tests green; update+merge packages green; real-CLI smoke reproduced both directions; goldens green after inherited-red fix). Deduction: AC-UHD-003's adoption-gate letter (four mechanical RED elements in §E.2) not satisfied as specified — substance equivalent evidence substituted (F3). |
| Security (25%) | 95/100 | PASS | READ-ONLY mandate proven mechanically: zero write primitives in `doctor_hook_delivery.go` (grep for WriteFile/OpenFile/os.Create/os.Remove/Rename/Chmod/Truncate/WriteString → 0 matches; full file read confirms only `os.ReadFile` + `readHookOptInEnabled` → `config.LoadSystemHookOptInEnabled`, a read path); tests assert settings.json mtime+content invariance and no-file-creation. Display names surfaced in reports come from the template side only; script-path extraction regex excludes control characters. |
| Craft (20%) | 88/100 | PASS | New-code coverage measured by auditor: checkHookDelivery 81.7% / renderedTemplateHooks 69.2% / hookEntryIdentities 96.3% / sortedTemplateEventKeys 100.0% — all 12 uncovered blocks are exception/defensive/verbose arms (block-level listing verified), none on a detection path. golangci-lint 0 issues on touched packages. Tests assert BOTH directions (missing→warn, parity→silent), assert mtime+content, and use a test-side extraction pattern to avoid tautology. Deduction: two functions under the 85% function-level aspiration; package-level 85%/90% gate unmeasured (Gap G1/G4). |
| Consistency (15%) | 80/100 | PASS | Code matches existing doctor-check patterns (DiagnosticCheck, uikit statuses, workspaceChecks registration, golden snapshots, English comments, %w wrapping). Deduction: the shipped entry-identity rule (path+matcher) diverges from the rule recorded in design.md §G, progress.md M1, and the CHANGELOG identity sentence (all path-only) — F1; §E.2's boundary-evidence range and file list are inaccurate as written — F2; M2 executed off plan-letter — F3. |

## Per-AC matrix (13)

| AC | Status | Auditor-verified basis (commands run in this audit, this tree 940adf966) |
|----|--------|--------------------------------------------------------------------------|
| AC-UHD-001 (guard: new event key delivered) | PASS | Merge path untouched by the card: `git diff 7664729ab^..940adf966 -- internal/` = 8 files, zero non-test lines under internal/cli/update or internal/merge. `go test ./internal/cli/update/... ./internal/merge/... -count=1` → 7× `ok`. Generic template-new-key guard exists (`TestMergeUserFilesAddsTemplateEntry`, base_test.go:164). Hook-specific new-key characterization not written (plan M2(a) letter gap, F3) — guard holds by no-change + generic test. |
| AC-UHD-002 (deletion preserved) | PASS | `TestMergeKeepsUserDeletionInCarriedEventKey` green in the run above; asserts length==1 and no resurrection of handle-session-start-navigator.sh. |
| AC-UHD-003 (core gap resolved, Option B) | PASS (substance) | Defect surface pinned: `TestMergeDropsTemplateAdditionInsideCarriedEventKey` green ⇒ the silent drop is live and permanently characterized. Resolution evidenced by auditor's independent real-CLI smoke (below): warn names event key + per-matcher entries + remediation, rc=0. Adoption-gate LETTER (failing command + verbatim RED output + exit code + fixture + tree SHA in §E.2) not satisfied — F3. |
| AC-UHD-004 | N/A | Option A gated; not selected (M1, design.md §G). |
| AC-UHD-005 | N/A | Option A gated; not selected. |
| AC-UHD-006 (read-only specific report) | PASS | `TestCheckHookDelivery_FlagsMissingEntryInCarriedEventKey` green — asserts event key + missing entry name + `git status --porcelain` guidance + Detail under verbose + `assertReadOnlySettings` (mtime+content). Real-CLI smoke A reproduced end-to-end (below). |
| AC-UHD-007 | N/A | Option C gated; not selected. |
| AC-UHD-008 (invalid JSON graceful) | PASS | `TestCheckHookDelivery_InvalidJSONWarnsGracefully` green — warn names settings.json; broken file byte-unchanged. |
| AC-UHD-009 (non-array hook value) | PASS | `TestCheckHookDelivery_NonArrayHookValueWarnsAndSkips` green — anomaly named, key skipped. Note F4: the "other keys resolve" half uses a synthetic non-template key the loop never iterates — that half is unproven by this test (primary half proven; loop `continue` semantics read). |
| AC-UHD-010 (idempotence) | PASS | `TestCheckHookDelivery_SilentWhenAllEntriesPresent` green; the check is stateless and read-only (verified in source), so repeated runs are idempotent; update side unchanged (Option B). |
| AC-UHD-011 (non-hook keys byte-identical) | PASS | Read-only by construction (Security row); comparison direction pinned by `TestCheckHookDelivery_DoesNotFlagTemplateNewEventKeys` + `TestCheckHookDelivery_UserAuthoredEntryNotFlagged`, both green. |
| AC-UHD-012 (pre-commit axis untouched) | PASS | Auditor re-measurement: card-scope `git diff 7664729ab^..940adf966 -- internal/` = 8 files (doctor.go +4, detector +256, detector tests +408, 3 goldens, merge/base_test.go +131, catalog.yaml 1); grep for `installPreCommitHookOptional` / `.git/hooks` over BOTH card scope and full d592b0551..940adf966 range → 0 matches. §E.2's cited range/file list inaccurate (F2) but the conclusion holds on every scope measured. |
| AC-UHD-013 (missing/hook-free informational) | PASS | `TestCheckHookDelivery_MissingSettingsJSONInformational` (asserts no settings.json created) + `TestCheckHookDelivery_NoHooksKeyInformational` green; real-CLI smoke B: `ok Hook Delivery skipped — .claude/settings.json carries no hooks`, rc=0. |

Quality gates: `golangci-lint run --timeout=2m ./internal/cli/... ./internal/cli/update/... ./internal/merge/...` → `0 issues.` · `TestDoctorGolden` → `ok` (inherited stale-golden red resolved inside 2b4582a43, dual attribution verified in the golden diff: Hook Delivery line + Agent Emit Embed message from 8f6cc5d7f/t427). GOOS=windows build was run by the lane (§E.2) and not re-run by the auditor (Gap G5).

## Commands run by the auditor (verbatim outputs)

```
$ go test ./internal/cli/ -run 'TestCheckHookDelivery|TestHookEntryIdentities' -count=1 -v
--- PASS: TestCheckHookDelivery_FlagsMissingEntryInCarriedEventKey (0.00s)
--- PASS: TestCheckHookDelivery_SilentWhenAllEntriesPresent (0.00s)
--- PASS: TestCheckHookDelivery_MissingSettingsJSONInformational (0.00s)
--- PASS: TestCheckHookDelivery_NoHooksKeyInformational (0.00s)
--- PASS: TestCheckHookDelivery_InvalidJSONWarnsGracefully (0.00s)
--- PASS: TestCheckHookDelivery_NonArrayHookValueWarnsAndSkips (0.00s)
--- PASS: TestCheckHookDelivery_DoesNotFlagTemplateNewEventKeys (0.00s)
--- PASS: TestCheckHookDelivery_UserAuthoredEntryNotFlagged (0.00s)
--- PASS: TestCheckHookDelivery_OptInAwareRendering (0.00s)
--- PASS: TestHookEntryIdentities (0.00s)
--- PASS: TestCheckHookDelivery_RegisteredInWorkspaceChecks (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	0.977s

$ go test ./internal/cli/update/... ./internal/merge/... -count=1
ok  github.com/modu-ai/moai-adk/internal/cli/update        0.366s
ok  github.com/modu-ai/moai-adk/internal/cli/update/backup 0.842s
ok  github.com/modu-ai/moai-adk/internal/cli/update/deploy 1.131s
ok  github.com/modu-ai/moai-adk/internal/cli/update/merge  1.777s
ok  github.com/modu-ai/moai-adk/internal/cli/update/plan   2.434s
ok  github.com/modu-ai/moai-adk/internal/cli/update/report 2.072s
ok  github.com/modu-ai/moai-adk/internal/merge             1.331s

$ go test ./internal/cli/ -run 'TestDoctorGolden' -count=1
ok  	github.com/modu-ai/moai-adk/internal/cli	0.860s

$ golangci-lint run --timeout=2m ./internal/cli/... ./internal/cli/update/... ./internal/merge/...
0 issues.

$ go tool cover -func (targeted run) → doctor_hook_delivery.go:
checkHookDelivery 81.7% / renderedTemplateHooks 69.2% / hookEntryIdentities 96.3% / sortedTemplateEventKeys 100.0%
(12 uncovered blocks, all exception/defensive/verbose arms — block list verified against source lines 65-67, 84-88, 91-100, 119-120, 134-136, 172-174, 182-184, 186-188, 190-192, 211-212)

$ grep -nE 'WriteFile|OpenFile|os\.Create|os\.Remove|Rename|Chmod|Chown|Truncate|WriteString|Write\(' internal/cli/doctor_hook_delivery.go
(no output — 0 matches)

Real-CLI smoke (auditor-built /tmp/t466-moai from this tree, /tmp fixtures):
Fixture A: settings.json = {"hooks":{"PreToolUse":[]}}
  warn Hook Delivery  hooks.PreToolUse missing handle-pre-tool.sh (matcher Agent|Task); … (matcher AskUserQuestion); … (matcher SendMessage|TaskStop); … (matcher Write|Edit|Bash); re-add each missing entry … git status --porcelain | grep '^ D'
  0 ok, 1 warn, 0 fail   RC_A=0
Fixture B: settings.json = {"model":"opus"}
  ok Hook Delivery  skipped — .claude/settings.json carries no hooks
  1 ok, 0 warn, 0 fail   RC_B=0
```

## Findings (structured defect-list)

- F1 [MINOR] [optional] `internal/cli/doctor_hook_delivery.go:214-234` vs `.moai/specs/SPEC-UPDATE-HOOK-DELIVERY-001/design.md:70` / `progress.md:18` (M1) / `CHANGELOG.md` [Unreleased] Added entry — The shipped identity rule is handler-path + `|matcher`; the three records describe handler-path only. The implementation is NECESSARY, not cosmetic: the template's `hooks.PreToolUse` carries `handle-pre-tool.sh` in 4 blocks with 4 matchers (template settings.json.tmpl:53-91) — a path-only identity collapses them to one, so a user carrying any ONE block would read as full parity with 3 blocks missing. Tests guard the implemented rule (a path-only revert fails `FlagsMissingEntryInCarriedEventKey` — the fixture's missing block shares the script with the kept block). But the M3 commit message cites "(design.md §G rule)" while deviating from §G, and the CHANGELOG's user-facing identity sentence misdescribes shipped behavior. Required fix: correct the CHANGELOG identity sentence, and record the supersession of design.md §G's identity clause (amendment path or an explicit note); progress.md M1 line likewise.
- F2 [MINOR] [optional] `progress.md:53` (§E.2 AC-UHD-012 row) — Evidence cites `git diff d592b0551..HEAD -- internal/` "touches only internal/cli/{doctor.go, doctor_hook_delivery.go, doctor_hook_delivery_test.go, testdata/doctor-*.golden}". Actual output over that range: 53 files / +2498 (the branch absorbed origin/develop at b77ae5d5e, including t461's `internal/template/templates/.git_hooks/pre-commit` change via 8ab11ed99), and the card's own list omits `internal/cli/update/merge/base_test.go` (M2's file). The conclusion (this SPEC touched no pre-commit axis) is TRUE — auditor-verified at card scope 7664729ab^..940adf966 (8 files) and by 0-match greps on both scopes. Required fix: re-scope the §E.2 command citation to the card's commits (or note the absorption), add base_test.go to the list. Record-accuracy only; not blocking.
- F3 [MINOR] [optional] plan.md M2 vs executed — plan.md:44-45 and §E (line 37) required (a) THREE characterization tests (REQ-UHD-001/002/003) and (b) a RED observation for AC-UHD-003 with four mechanical elements (failing command + verbatim RED output + exit code + fixture + tree SHA) captured at M2; acceptance.md:18 makes the four elements an adoption gate. Executed: one characterization test (002) + a defect-surface test green-at-birth; §E.2 carries no four-element RED cell (the §E.2 "Mutation check" is the DETECTOR's mutant evidence from M3, a different axis). Substance is equivalent and independently re-verified by this audit (the defect-surface test passing proves the drop is live; the detector tests + smoke prove the resolution), so AC-UHD-003 stands on substance. Required fix: record the deviation explicitly (a §E.2 note acknowledging the substitution of green-characterization for RED-observation and the two undelivered characterization tests) rather than leaving the adoption-gate letter silently unmet.
- F4 [MINOR] [optional] `internal/cli/doctor_hook_delivery_test.go:239-260` — `TestCheckHookDelivery_NonArrayHookValueWarnsAndSkips` builds its "other carried key" as `eventKey + "-full"`, a name absent from the template's key set, so the comparison loop never visits it: the AC-UHD-009 clause "other keys resolve normally" is unproven by this test (the anomaly half is proven). Required fix if tightening: use a second real template event key, fully populated, alongside the anomalous one.
- F5 [INFO] [optional] CHANGELOG says "13 acceptance criteria, 10 PASS / 3 N/A" — matches §E.3 counts (10/3/0) and the §E.2 matrix. Verified accurate; recorded here as a checked claim, not a defect.

## Gaps (explicitly NOT observed)

- G1: Full `internal/cli` package suite — NOT run by auditor (600s package-timeout history under lane load; CI owns the full-suite verdict). Targeted runs only.
- G2: Real `moai update` execution — NOT observed. Merge path untouched is established by diff (0 non-test lines under internal/cli/update, internal/merge), not by execution.
- G3: opt_in=true CLI smoke — NOT executed end-to-end; covered only by `TestCheckHookDelivery_OptInAwareRendering` (unit).
- G4: Package-level coverage vs the 85%/90% §4 gate — NOT measured at package level (targeted-run basis only; §E.2 discloses this honestly).
- G5: `GOOS=windows GOARCH=amd64 go build ./...` — lane-reported pass (§E.2/§E.3), not re-run by auditor.
- G6: Hugo build of docs-site — lane-reported warning-free with 7/7 section parity; auditor verified 4-locale section counts (7/7/7/7) and the Hook Delivery section's presence/position in all four locales, but did not rebuild Hugo.

## Residual-risk

- The detector cannot distinguish user-deleted from never-delivered (design.md §G accepted cost): a user who deliberately removed a template entry is re-warned on every doctor run, forever, until Option A lands. Accepted at M1; wording stays neutral.
- User-customized matchers produce false "missing" reports (the matcher component of the identity): a user who edited a template block's matcher is warned for the block they possess in modified form. This is the flip side of the F1 divergence — the implemented rule chose precision over tolerance, a trade-off the records never made.
- If a future template hook entry references no `.claude/hooks/moai` script, its canonical JSON becomes both identity and display name (doctor_hook_delivery.go:242) — verbose report lines; harmless today (all template entries carry scripts).
- F1's stale records (design.md §G, progress.md M1, CHANGELOG sentence) are now frozen into a completed SPEC; without a follow-up they misdescribe the shipped identity rule indefinitely.

## Sync-surface verification

- CHANGELOG entry present, top of [Unreleased]→Added; B12 self-tests a/b/c recorded in §E.4; AC count claim 13 (10 PASS/3 N/A) cross-checked against §E.2/§E.3 — consistent (F5).
- docs-site 4-locale parity: en/ja/zh/ko doctor.md each carry 7 `##` sections including the new `Hook Delivery` section (ko at line 55, others at 53) with badge v3.1.4; content localized natively (ko page read by auditor — accurate: read-only, matcher naming, opt-out handling, remediation).
- §E.4 completeness: sync_commit_sha placeholder→85743a34a backfilled at 940adf966 (diff verified); §E.3 run_commit_sha backfill (7664729ab, f51cb973d, 2b4582a43) recorded with the ownership-route note (manager-docs under lead dispatch); canary/canary-compliance, docs_site_decision, readme_decision, mx_tag_changes fields present.
- spec.md frontmatter: in-progress → completed inside the single sync commit 85743a34a; body untouched (diff verified — frontmatter only); updated bumped same-day.
- Scope discipline across the 6 commits: all within Option B blast radius; no Option-A machinery, no drive-by refactors, no untouched-file surprises (a05c17164's catalog.yaml hash refresh is the declared develop-absorb chore).

Classification: sync-phase audit verdict, card t466 · Evidence basis: all commands above executed in this audit run against tree 940adf966 in worktree /Users/goos/MoAI/moai-adk-go/.claude/worktrees/t466.
