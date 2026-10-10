# SPEC-USER-SETTINGS-PROTECT-001 — Acceptance Criteria

This file is the verification layer. Each criterion is written as Given-When-Then and names its verifying command. A release-blocking criterion carries a RED cell with four elements: (a) the command, as one shell invocation; (b) its stdout, verbatim; (c) its exit code, as a separate field; (d) the tree SHA. The requirement layer (GEARS) lives in spec.md §2 and is not restated here.

Pre-implementation tree for every RED-now cell: `3975fe3cc25eb1cbc3be79aa16ff3bcd99f8ff21` (3975fe3cc) on `WT-3-2-0`. P-3 (plan.md §C, run phase) and M2 (plan.md §F, run phase) re-observe the in-scope cells on the post-landing run base, each with the run-base SHA recorded; both observations are recorded in progress.md beside the pinned-tree observations, and neither replaces the other.

Evidence sources: `.moai/reports/t1630/evidence/` (local and git-ignored, not tracked; the tracked binding is `.moai/specs/SPEC-USER-SETTINGS-PROTECT-001/evidence-manifest.json`) holds every probe source, overlay definition, and script cited below. Probes run through `go -C <worktree> test -overlay <overlay>` and write nothing into the repository. Scripts: `.moai/reports/t1630/evidence/home-manifest.sh`, `.moai/reports/t1630/evidence/check-codex-env.sh`, `.moai/reports/t1630/evidence/run-ac009.sh`, `.moai/reports/t1630/evidence/grep-cell.sh`, and `.moai/reports/t1630/evidence/fakebin/codex` (these scripts serve the moved criteria AC-006, AC-007, and AC-009; Out of Scope - moved to t1666). Probe-written files exist only under `.moai/reports/t1630/evidence/throwaway-home/`. Repair-round probes: `.moai/reports/t1630/evidence/probe-ac003b_test.go.txt` (overlay `.moai/reports/t1630/evidence/overlay-ac003b.json`, AC-003 part B) and the `TestAC011DBEntryListed` test in `.moai/reports/t1630/evidence/probe-ac011_test.go.txt` (AC-011 clause ii; Out of Scope - moved to t1666), and `.moai/reports/t1630/evidence/probe-ac005b_test.go.txt` (overlay `.moai/reports/t1630/evidence/overlay-ac005b.json`, AC-005 case ii).

Class key: release-blocking (the run cannot close without it), regression-guard (green today and must stay green; or undecidable at plan time, per verification-completeness §2.1: not release-blocking, and not recorded as a pass until its RED is observed in the run phase), gate (a precondition the run checks before it changes anything).

Executed-count gate: a selector counts as executed only when `-v` prints `=== RUN` lines. Each criterion states its minimum N. Zero executed, or a `no tests to run` line, is a failure.

Precondition cell (applies to the RED observations recorded in this file; all were taken on the pinned tree 3975fe3cc):
- (a) `git diff --stat 3975fe3cc HEAD -- internal cmd`, run at the revision under audit; it prints nothing when the Go sources at HEAD are identical to the pinned tree
- (b) stdout: empty
- (c) exit code: 0
- (d) tree: 3975fe3cc25eb1cbc3be79aa16ff3bcd99f8ff21 (the Go sources are identical to the pinned commit)
The run-base re-observation has its own precondition in plan.md §C (P-2 landing gate, then P-3 re-observation); it does not replace this cell. Out of Scope - moved to t1666: the post-t1619 baseline (a measurement pinned after commit 569a3fe5f).

Test-environment rule: the session does not redirect the shell HOME (the worktree guard refuses such commands), so no probe changes the shell HOME. The AC-003 part B probe sets HOME and MOAI_HOME inside its own test process to t.TempDir paths, so no read or write reaches the operator's home. Each in-scope probe is chosen so that it resolves no home path, or runs in a package that has no TestMain (`internal/config/toolpolicy`, `internal/core/project`). The cli and receipt probes belong to moved criteria (Out of Scope - moved to t1666).

## AC-001 — The default init path keeps the user's allow, ask, deny, and unmodelled keys (REQ-001)

- Given: a USER-scope settings file with `permissions.defaultMode` "plan", `allow`, `ask`, `deny`, `additionalDirectories`, and a sibling `env` key.
- When: the USER-scope writer runs on the default path (`toolpolicy.WriteUserDefaultMode`, which the empty-tier init path calls per plan-phase E-6).
- Then: `allow`, `ask`, `deny`, `additionalDirectories`, and `env` are unchanged. Only `permissions.defaultMode` may change, and only when it is absent (decided Q1; AC-003 part B).
- Contract anchor: SPEC-INIT-WIZARD-REPAIR-001 §4 (decision-index.md Q8, DECIDED). The M1 preservation test that §4 requires is absent from the tree (plan.md E-23); the run adds it under this criterion.
- Verifying command: `go -C /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1630 test -overlay /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1630/.moai/reports/t1630/evidence/overlay-ac001-003.json -count=1 -v -run '^TestAC001PreservesListsAndUnmodelledKeys$' ./internal/config/toolpolicy/` (run-phase: the same assertion authored in the repository; its name is recorded in progress.md).
- Class: release-blocking. Minimum executed: N=1.
- RED cell (observed on 3975fe3cc):
  - (a) command: the verifying command above.
  - (b) stdout:
    ```
    === RUN   TestAC001PreservesListsAndUnmodelledKeys
        zz_ac001_003_test.go:54: permissions.allow was removed; after={"permissions": {
                "defaultMode": "acceptEdits",
                "additionalDirectories": ["/tmp/x"]
              },"env":{"FOO":"1"}}
        zz_ac001_003_test.go:54: permissions.ask was removed; after={"permissions": {
                "defaultMode": "acceptEdits",
                "additionalDirectories": ["/tmp/x"]
              },"env":{"FOO":"1"}}
        zz_ac001_003_test.go:54: permissions.deny was removed; after={"permissions": {
                "defaultMode": "acceptEdits",
                "additionalDirectories": ["/tmp/x"]
              },"env":{"FOO":"1"}}
    --- FAIL: TestAC001PreservesListsAndUnmodelledKeys (0.00s)
    FAIL
    FAIL	github.com/modu-ai/moai-adk/internal/config/toolpolicy	0.257s
    FAIL
    ```
  - (c) exit code: 1
  - (d) tree: 3975fe3cc25eb1cbc3be79aa16ff3bcd99f8ff21
  - Executed: 1 (one `=== RUN`, one `--- FAIL`).
- Green path: after M2 the writer splices only `defaultMode`, and the three lists and `additionalDirectories` survive.

## AC-002 — The permissions region is byte-identical when the defaultMode already matches (REQ-002; card judgement 1)

- Given: a USER-scope settings file whose `permissions.defaultMode` already equals the resolved tier default, with the lists written in a non-canonical layout.
- When: the USER-scope writer runs (`toolpolicy.WriteUserDefaultMode`).
- Then: the file bytes are identical before and after the run, so the permissions section shows a diff of 0.
- Verifying command: `go -C /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1630 test -overlay /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1630/.moai/reports/t1630/evidence/overlay-ac001-003.json -count=1 -v -run '^TestAC002RegionBytesUnchangedWhenDefaultModeMatches$' ./internal/config/toolpolicy/`.
- Class: release-blocking. Minimum executed: N=1.
- RED cell (observed on 3975fe3cc):
  - (a) command: the verifying command above.
  - (b) stdout:
    ```
    === RUN   TestAC002RegionBytesUnchangedWhenDefaultModeMatches
        zz_ac001_003_test.go:72: file bytes changed
            before="{\n  \"permissions\": {\n      \"allow\": [\"Bash(make:*)\"],\n      \"defaultMode\":   \"acceptEdits\"\n  },\n  \"env\": {\"FOO\": \"1\"}\n}\n"
            after="{\n  \"permissions\": {\n    \"defaultMode\": \"acceptEdits\"\n  },\n  \"env\": {\"FOO\": \"1\"}\n}\n"
    --- FAIL: TestAC002RegionBytesUnchangedWhenDefaultModeMatches (0.00s)
    FAIL
    FAIL	github.com/modu-ai/moai-adk/internal/config/toolpolicy	0.089s
    FAIL
    ```
  - (c) exit code: 1
  - (d) tree: 3975fe3cc25eb1cbc3be79aa16ff3bcd99f8ff21
  - Executed: 1.
- Green path: after M2 the writer skips the rewrite when the resolved values already match, so the bytes are unchanged.

## AC-003 — The lists survive under either defaultMode state, and an existing defaultMode is kept on the default and automatic paths (decision Q1, DECIDED; the restated clause is in the spec.md §2 note on REQ-003)

Part A (release-blocking): the lists and unmodelled keys survive whether the existing defaultMode differs from or matches the resolved tier default.
- Verifying command (part A): `go -C /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1630 test -overlay /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1630/.moai/reports/t1630/evidence/overlay-ac001-003.json -count=1 -v -run '^TestAC003ListsPreservedUnderEitherDisposition$' ./internal/config/toolpolicy/`.
- Minimum executed: N=3 (the parent test and its two subtests `differing` and `matching`).
- RED cell (observed on 3975fe3cc; re-observed in revision 2 at HEAD 5dc6c4530, where `git diff --stat 3975fe3cc HEAD -- internal cmd` prints nothing):
  - (a) command: the verifying command for part A, verbatim and with no redirect: `go -C /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1630 test -overlay /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1630/.moai/reports/t1630/evidence/overlay-ac001-003.json -count=1 -v -run '^TestAC003ListsPreservedUnderEitherDisposition$' ./internal/config/toolpolicy/`.
  - (b) stdout (verbatim; re-observed at HEAD 5dc6c4530):
    ```
    === RUN   TestAC003ListsPreservedUnderEitherDisposition
    === RUN   TestAC003ListsPreservedUnderEitherDisposition/differing
        zz_ac001_003_test.go:79: permissions.allow was removed; after={"permissions": {
                "defaultMode": "acceptEdits",
                "additionalDirectories": ["/tmp/x"]
              },"env":{"FOO":"1"}}
        zz_ac001_003_test.go:79: permissions.ask was removed; after={"permissions": {
                "defaultMode": "acceptEdits",
                "additionalDirectories": ["/tmp/x"]
              },"env":{"FOO":"1"}}
        zz_ac001_003_test.go:79: permissions.deny was removed; after={"permissions": {
                "defaultMode": "acceptEdits",
                "additionalDirectories": ["/tmp/x"]
              },"env":{"FOO":"1"}}
    === RUN   TestAC003ListsPreservedUnderEitherDisposition/matching
        zz_ac001_003_test.go:83: permissions.allow was removed; after={"permissions": {
                "defaultMode": "acceptEdits",
                "additionalDirectories": ["/tmp/x"]
              },"env":{"FOO":"1"}}
        zz_ac001_003_test.go:83: permissions.ask was removed; after={"permissions": {
                "defaultMode": "acceptEdits",
                "additionalDirectories": ["/tmp/x"]
              },"env":{"FOO":"1"}}
        zz_ac001_003_test.go:83: permissions.deny was removed; after={"permissions": {
                "defaultMode": "acceptEdits",
                "additionalDirectories": ["/tmp/x"]
              },"env":{"FOO":"1"}}
    --- FAIL: TestAC003ListsPreservedUnderEitherDisposition (0.00s)
        --- FAIL: TestAC003ListsPreservedUnderEitherDisposition/differing (0.00s)
        --- FAIL: TestAC003ListsPreservedUnderEitherDisposition/matching (0.00s)
    FAIL
    FAIL	github.com/modu-ai/moai-adk/internal/config/toolpolicy	0.089s
    FAIL
    ```
  - (c) exit code: 1
  - (d) tree: 3975fe3cc25eb1cbc3be79aa16ff3bcd99f8ff21 (re-observed at HEAD 5dc6c4530)
  - Executed: 3.

Part B (release-blocking; decided Q1): an existing USER-scope `permissions.defaultMode` that differs from the resolved tier default is kept on the default init path (`project.ApplyAutonomyTierBundle` with an empty persisted tier, which resolves to semi-auto) and on the automatic path (persisted tier `automatic`). Both calls take the user settings path, and the probe compares the value in place.
- Verifying command (part B): `go -C /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1630 test -overlay /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1630/.moai/reports/t1630/evidence/overlay-ac003b.json -count=1 -v -run '^TestAC003DefaultModeKeptOnBothPaths$' ./internal/core/project/`.
- Class: release-blocking. Minimum executed: N=3 (the parent test and its subtests `default-init` and `automatic`).
- RED cell (observed on 3975fe3cc):
  - (a) command: the verifying command for part B.
  - (b) stdout:
    ```
    === RUN   TestAC003DefaultModeKeptOnBothPaths
    === RUN   TestAC003DefaultModeKeptOnBothPaths/default-init
        zz_ac003b_test.go:56: existing defaultMode "plan" was overwritten with "acceptEdits"; after={"permissions": {
                "defaultMode": "acceptEdits"
              }}
    === RUN   TestAC003DefaultModeKeptOnBothPaths/automatic
        zz_ac003b_test.go:56: existing defaultMode "plan" was overwritten with "auto"; after={"permissions": {
                "defaultMode": "auto"
              }}
    --- FAIL: TestAC003DefaultModeKeptOnBothPaths (0.00s)
        --- FAIL: TestAC003DefaultModeKeptOnBothPaths/default-init (0.00s)
        --- FAIL: TestAC003DefaultModeKeptOnBothPaths/automatic (0.00s)
    FAIL
    FAIL	github.com/modu-ai/moai-adk/internal/core/project	0.379s
    FAIL
    ```
  - (c) exit code: 1
  - (d) tree: 3975fe3cc25eb1cbc3be79aa16ff3bcd99f8ff21
  - Executed: 3.
- Green path: after M2 the writer writes `defaultMode` only when it is absent; both subtests pass.

## AC-004 — The settings template carries one permissions.defaultMode key with the value "default" (decision Q2, DECIDED; the restated clause is in the spec.md §2 note on REQ-004)

- Given: the settings template `internal/template/templates/.claude/settings.json.tmpl`.
- When: the template's `permissions.defaultMode` key is matched with the value `"default"`.
- Then: the template carries exactly one `"defaultMode"` key, and its value is `"default"` (decided Q2; the value is pinned by the lane in progress.md §G, G-17). The rules that a fresh init writes the key only where it is absent and that update never modifies it are not asserted by this criterion; they are stated in the spec.md §2 note on REQ-004, and their normative content is carried by follow-up card t1666.
- Verifying command: `grep -c '"defaultMode": *"default"' /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1630/internal/template/templates/.claude/settings.json.tmpl`; the count must print 1.
- Class: release-blocking (decided Q2). Minimum executed: N=1 (the count).
- RED cell (observed on 3975fe3cc, this session, value pattern):
  - (a) command: the verifying command above.
  - (b) stdout: `0`
  - (c) exit code: 1 (`grep -c` exits 1 on zero matches; the same command followed by `echo "template-defaultMode-value-count-exit=$?"` printed `template-defaultMode-value-count-exit=1`).
  - (d) tree: 3975fe3cc25eb1cbc3be79aa16ff3bcd99f8ff21
  - The count 0 against an expected 1 is red for the stated reason: the template carries no key (`internal/cli/launcher.go:738` and `:1121` state that the template stopped shipping a default).
- Green path: after M3 the template carries one key with the pinned value, and the count is 1.

## AC-005 — The PROJECT-scope policy path keeps user-added permissions.allow entries by set union, with and without a sidecar record (REQ-005; Q5 DECIDED)

- Given: a tool-policy document is present (`.moai/config/sections/tool-policy.yaml`), and the project settings file holds a user-added `permissions.allow` entry that the document does not list. Case (i): no sidecar record exists at `.moai/state/tool-policy/managed-allow.json`. Case (ii): the record exists with `last_generated`, and the project allow list also holds an entry in `last_generated` that the document no longer generates.
- When: `project.ApplyAutonomyTierBundle` runs the full-bundle path with the `automatic` tier.
- Then: case (i): every existing allow entry is kept, so the user-added entry is present and nothing is removed. Case (ii): user_added = existing allow entries minus `last_generated`; the result is the regenerated list union user_added, so the user-added entry is present and the `last_generated` entry the document no longer generates is absent. Only the user removes a user-added entry (decided Q5; diff-based detection pinned in progress.md §G, G-20).
- Verifying command: `go -C /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1630 test -overlay /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1630/.moai/reports/t1630/evidence/overlay-ac005.json -count=1 -v -run '^TestAC005PolicyPathKeepsUserAllowEntry$' ./internal/core/project/` (case i). Case (ii): `go -C /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1630 test -overlay /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1630/.moai/reports/t1630/evidence/overlay-ac005b.json -count=1 -v -run '^TestAC005WithRecordKeepsUserAddedEntry$' ./internal/core/project/`.
- Class: release-blocking (decided Q5). Minimum executed: N=1 per case.
- RED cell, case (i) (observed on 3975fe3cc; the assertion is the decided retention):
  - (a) command: the verifying command above.
  - (b) stdout:
    ```
    === RUN   TestAC005PolicyPathKeepsUserAllowEntry
        zz_ac005_test.go:59: user-added permissions.allow entry dropped; before={"permissions":{"allow":["Bash(user-added:*)"]}} after={"permissions": {
                "ask": [
                  "Bash(git push:*)"
                ],
                "deny": [
                  "Bash(git push --force:*)"
                ]
              }}
    --- FAIL: TestAC005PolicyPathKeepsUserAllowEntry (0.00s)
    FAIL
    FAIL	github.com/modu-ai/moai-adk/internal/core/project	0.365s
    FAIL
    ```
  - (c) exit code: 1
  - (d) tree: 3975fe3cc25eb1cbc3be79aa16ff3bcd99f8ff21
  - Executed: 1.
- RED cell, case (ii) (observed on 3975fe3cc):
  - (a) command: the case (ii) verifying command above.
  - (b) stdout:
    ```
    === RUN   TestAC005WithRecordKeepsUserAddedEntry
        zz_ac005b_test.go:79: user-added permissions.allow entry dropped with a sidecar record present; before={"permissions":{"allow":["Bash(old-managed:*)","Bash(user-added:*)"]}} after={"permissions": {
                "ask": [
                  "Bash(git push:*)"
                ],
                "deny": [
                  "Bash(git push --force:*)"
                ]
              }}
    --- FAIL: TestAC005WithRecordKeepsUserAddedEntry (0.01s)
    FAIL
    FAIL	github.com/modu-ai/moai-adk/internal/core/project	0.427s
    FAIL
    ```
  - (c) exit code: 1
  - (d) tree: 3975fe3cc25eb1cbc3be79aa16ff3bcd99f8ff21
  - Executed: 1.
- Green path: after M3 implements the diff-based set union, both cases pass unchanged.

## AC-006 — Every package that reaches a home resolver sandboxes MOAI_HOME in its TestMain; the verifier observes it at run time (REQ-006)

**Out of Scope - moved to t1666.** Reason: the sandbox criterion for REQ-006 moves with the test-isolation items; audit defects B5 (the AC-006 cell 1 order) and B10 (the AC-006 run-time verdict) sit in its cells. The text below is retained unchanged.

Each cell below is one command with its own stdout and exit code. The text checks are plan-phase observations only. The run-time check is the verdict.

- Verifying command (run-time, the verdict): `go -C /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1630 test -overlay /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1630/.moai/reports/t1630/evidence/overlay-ac006.json -count=1 -v -run '^TestAC006SandboxObserved$' ./internal/contract/receipt/`. The probe asserts, inside the test process, that MOAI_HOME and HOME are absolute roots. The run-phase repetition covers every reaching package listed in §A.4 Basis of plan.md (homestate, escalation, factory, factorymsg, web, contract/receipt), once TestMain installs the sandbox.
- Class: release-blocking. Minimum executed: N=1 for the observed cell.
- RED cells (observed on 3975fe3cc):
  - Cell 1 (text check, five TestMain packages). (a) `sh /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1630/.moai/reports/t1630/evidence/grep-cell.sh -L -E 'EnvHome|MOAI_HOME' /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1630/internal/homestate/main_test.go /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1630/internal/escalation/main_test.go /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1630/internal/factory/main_test.go /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1630/internal/factorymsg/main_test.go /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1630/internal/web/main_test.go`. (b) stdout: the five paths in order `internal/homestate/main_test.go`, `internal/escalation/main_test.go`, `internal/factory/main_test.go`, `internal/factorymsg/main_test.go`, `internal/web/main_test.go`, then `grep exit=1` (printed by the wrapper; the wrapper's grep resolves to a different implementation under `sh` than in the interactive shell, so the verdict is read from stdout). (c) exit code: 1 (printed by the wrapper). (d) tree: 3975fe3cc.
  - Cell 2 (existence check). (a) `ls /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1630/internal/contract/receipt/main_test.go`. (b) stdout: empty; the message `ls: …/internal/contract/receipt/main_test.go: No such file or directory` is returned on stderr. (c) exit code: 1. (d) tree: 3975fe3cc.
  - Cell 3 (run-time, the verdict). (a) the verifying command above. (b) stdout: `=== RUN   TestAC006SandboxObserved` / `    zz_ac006_probe_test.go:13: MOAI_HOME="" is not an absolute test-owned root` / `--- FAIL: TestAC006SandboxObserved (0.00s)` / `FAIL` / `FAIL	github.com/modu-ai/moai-adk/internal/contract/receipt	0.607s` / `FAIL`. (c) exit code: 1. (d) tree: 3975fe3cc. Executed: 1.
- Comment-only mutant (must fail the run-time check): `.moai/reports/t1630/evidence/probe-ac006-mutant_test.go.txt` adds a comment that names the sandbox and installs nothing; overlay `.moai/reports/t1630/evidence/overlay-ac006-mutant.json` adds it to the receipt package beside the probe.
  - Mutant text check: (a) `sh /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1630/.moai/reports/t1630/evidence/grep-cell.sh -L -E 'EnvHome|MOAI_HOME' /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1630/.moai/reports/t1630/evidence/probe-ac006-mutant_test.go.txt`. (b) stdout: empty; then `grep exit=0`. (c) exit code 0 (printed by the wrapper). The text check passes the mutant, which is the defect that D19 describes.
  - Mutant run-time check: (a) `go -C /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1630 test -overlay /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1630/.moai/reports/t1630/evidence/overlay-ac006-mutant.json -count=1 -v -run '^TestAC006SandboxObserved$' ./internal/contract/receipt/`. (b) stdout: identical to Cell 3. (c) exit code: 1. (d) tree: 3975fe3cc. The run-time check fails the mutant.
- Green path: after M6, Cell 2 returns a TestMain file for the receipt package, and the run-time check passes in all six packages.

## AC-007 — No test run writes under the operator's real home; the ~/.moai/run entry count does not increase (REQ-007; card judgement 2)

**Out of Scope - moved to t1666.** Reason: the throwaway-HOME leak verifier and the ~/.moai/run growth-0 criterion (REQ-007); audit defect B3 (AC-007). The text below is retained unchanged.

- Given: a before-manifest over the four home roots (`.moai`, `.claude`, `.codex`, `.agents`), taken on the operator's home by the operator (or on a throwaway account), before one scoped run.
- When: the scoped run of the reaching packages (REQ-007 set) runs under a throwaway HOME. The full-suite verdict is not part of this criterion; it is delegated to the CI workflow on the pushed branch.
- Then: the after-manifest equals the before-manifest. No path under the four roots is created, modified, or removed. The entry count under `~/.moai/run` is unchanged.
- Verifying command (verdict): `sh /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1630/.moai/reports/t1630/evidence/home-manifest.sh verify <home> <before> <after>`. The script follows symlinks (`find -L`). Any find failure exits 2, never an empty manifest. The script exits 1 when the manifests differ.
- Class: release-blocking. Minimum executed: N=1 for the leak probe (the RED is a verifier cell).
- RED cell (verifier, observed on 3975fe3cc, on the throwaway HOME `.moai/reports/t1630/evidence/throwaway-home`):
  - Step 1 — snapshot. (a) `sh /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1630/.moai/reports/t1630/evidence/home-manifest.sh snapshot /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1630/.moai/reports/t1630/evidence/throwaway-home /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1630/.moai/reports/t1630/evidence/manifest-ac007-before.txt`. (b) stdout: empty. (c) exit 0. (d) tree 3975fe3cc.
  - Step 2 — simulated write. (a) `go -C /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1630 test -overlay /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1630/.moai/reports/t1630/evidence/overlay-ac007.json -count=1 -v -run '^TestAC007LeakProbe$' ./internal/config/toolpolicy/`. (b) stdout: `=== RUN   TestAC007LeakProbe` / `--- PASS: TestAC007LeakProbe (0.00s)` / `PASS` / `ok  	github.com/modu-ai/moai-adk/internal/config/toolpolicy	0.249s`. (c) exit 0. (d) tree 3975fe3cc. The probe writes `.moai/reports/t1630/evidence/throwaway-home/.moai/run/ac007-leak/leak.txt`.
  - Step 3 — verify. (a) `sh /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1630/.moai/reports/t1630/evidence/home-manifest.sh verify /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1630/.moai/reports/t1630/evidence/throwaway-home /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1630/.moai/reports/t1630/evidence/manifest-ac007-before.txt /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1630/.moai/reports/t1630/evidence/manifest-ac007-after.txt`. (b) stdout (the diff, returned on the tool's error channel):
    ```
    0a1,4
    > 1791618368 30 -rw------- …/evidence/throwaway-home/.moai/run/ac007-leak/leak.txt
    > 1791618368 96 drwx------ …/evidence/throwaway-home/.moai
    > 1791618368 96 drwx------ …/evidence/throwaway-home/.moai/run
    > 1791618368 96 drwx------ …/evidence/throwaway-home/.moai/run/ac007-leak
    4d7
    < absent …/evidence/throwaway-home/.moai
    ```
    (The paths are abbreviated after `t1630/.moai/specs/SPEC-USER-SETTINGS-PROTECT-001/` in this display; the script prints full paths.) (c) exit 1. (d) tree 3975fe3cc.
- The RED shows that the verifier detects a write under a home root. Q1 does not enter this criterion.
- Scope of the verifier (REQ-007): the four roots of the operator's real home. The reaching packages (REQ-007 set, plan.md §A.4 Basis) run on the sandboxed home in run phase (M6).
- Green path: after M6, step 2 writes nothing under the home root, and step 3 prints `manifest unchanged` with exit 0.

## AC-008 — Under the MOAI_HOME sandbox, RunProjectDir and the store resolvers resolve under the sandbox root (REQ-008)

**Out of Scope - moved to t1666.** Reason: its green path needs the receipt TestMain sandbox install, the AC-006 mechanism that moves with it (REQ-008). The text below is retained unchanged.

- Given: a test process that inherits MOAI_HOME from its parent command and calls no helper that sets it.
- When: the child-process probe calls `homestate.RunProjectDir`, and the receipt `StoreDir` test runs.
- Then: the probe finds MOAI_HOME set, and RunProjectDir returns a path under it. The receipt StoreDir test passes.
- Verifying command: `go -C /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1630 test -overlay /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1630/.moai/reports/t1630/evidence/overlay-ac008.json -count=1 -v -run '^(TestAC008RunProjectDirUnderSandbox|TestStoreDirMatchesEscalation)$' ./internal/contract/receipt/`. Selector gate: N=2 executed (one probe, one named StoreDir test). The escalation package's `TestStoreDirUsesQueueProjectKey` is not part of this run; its package has a TestMain that runs unsandboxed setup, so it is run in run phase after M6.
- Class: release-blocking. Named tests cited: `TestStoreDirMatchesEscalation` (internal/contract/receipt/dir_test.go:14).
- RED cell (observed on 3975fe3cc):
  - (a) command: the verifying command above.
  - (b) stdout:
    ```
    === RUN   TestStoreDirMatchesEscalation
    --- PASS: TestStoreDirMatchesEscalation (0.15s)
    === RUN   TestAC008RunProjectDirUnderSandbox
        zz_ac008_test.go:19: MOAI_HOME is unset in the test process: the test binary is not sandboxed
    --- FAIL: TestAC008RunProjectDirUnderSandbox (0.00s)
    FAIL
    FAIL	github.com/modu-ai/moai-adk/internal/contract/receipt	0.662s
    FAIL
    ```
  - (c) exit code: 1
  - (d) tree: 3975fe3cc25eb1cbc3be79aa16ff3bcd99f8ff21
  - Executed: 2.
- Green path: after M6 the sandbox install in the receipt package sets MOAI_HOME, the probe passes, and the StoreDir test keeps passing.

## AC-009 — The review-gate live Codex test sets CODEX_HOME to a temporary root (REQ-009; card item b)

**Out of Scope - moved to t1666.** Reason: the Codex CODEX_HOME run cell and its check script (REQ-009); audit defect B2 (the AC-009 run path). The text below is retained unchanged.

- Given: the live review-gate test runs with a codex binary first on PATH and MOAI_SKIP_LIVE_CODEX unset. The binary is the evidence fake (`.moai/reports/t1630/evidence/fakebin/codex`): it answers `--version`, writes a names-only record to `.moai/reports/t1630/evidence/codex-env-ac009.names.txt` for any other call (the variable names and a presence check; no values), and never reaches a model.
- When: `TestHandleCodexReviewGate_LiveCodexBlocksInjectionAndKey` runs (`internal/cli/codex_review_gate_live_test.go:35`).
- Then: the codex child environment carries CODEX_HOME (presence check on the names-only record). The record holds no values, so it cannot show where the value lies; the run-phase test asserts that under TMPDIR itself, because its CODEX_HOME comes from `t.TempDir()`. The verdict is the check script, because the test skips after the fake codex's review call fails (`codex review turn did not complete`), so the test's own assertion never runs.
- Verifying command (run-time, the verdict): `sh /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1630/.moai/reports/t1630/evidence/check-codex-env.sh /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1630/.moai/reports/t1630/evidence/codex-env-ac009.names.txt`, run after `sh …/evidence/run-ac009.sh` (the run cell; the script exists because the worktree guard refuses an inline PATH prefix).
- Class: release-blocking. Minimum executed: N=1 in the run cell.
- Plan-phase text check (kept as an observed RED): (a) `sh …/.moai/reports/t1630/evidence/grep-cell.sh -n 'CODEX_HOME' /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1630/internal/cli/codex_review_gate_live_test.go`. (b) stdout: `grep exit=1` (no match). (c) exit code 1. (d) tree 3975fe3cc.
- Comment-only mutant (must fail the run-time check): `.moai/reports/t1630/evidence/mutant-ac009-comment-only.txt` holds only `// t.Setenv("CODEX_HOME", t.TempDir())`. Its text check matches (exit 0), so the text verifier passes it. The run-time check reads the names that the run cell recorded, which a comment cannot change, so the mutant fails the check exactly as the current tree does.
- RED cells (observed on 3975fe3cc; the run cell and the check cell are restated in this round from the names-only record. The earlier check observation on the value-bearing log had the same exit 1 and the same absent-name result; that log is no longer in the repository and is not cited further):
  - Run cell. (a) `sh /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1630/.moai/reports/t1630/evidence/run-ac009.sh` (sandboxed HOME, names-only fake). (b) stdout: `=== RUN   TestHandleCodexReviewGate_LiveCodexBlocksInjectionAndKey` / `codex review: inconclusive (review call failed: codex stdout closed before response to id=1)` / `    codex_review_gate_live_test.go:138: codex review turn did not complete — the producer recorded inconclusive with the error surfaced (correct behavior)` / `--- SKIP: TestHandleCodexReviewGate_LiveCodexBlocksInjectionAndKey (1.74s)` / `PASS` / `ok  	github.com/modu-ai/moai-adk/internal/cli	3.025s`. (c) exit code 0. (d) tree 3975fe3cc. Executed: 1 (skipped after the review call).
  - Check cell. (a) `sh /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1630/.moai/reports/t1630/evidence/check-codex-env.sh /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1630/.moai/reports/t1630/evidence/codex-env-ac009.names.txt`. (b) stdout: `FAIL: CODEX_HOME absent from the codex child environment (names-only record: /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1630/.moai/reports/t1630/evidence/codex-env-ac009.names.txt)`. (c) exit code 1. (d) tree 3975fe3cc.
- Green path: after M5 the test sets CODEX_HOME to a temporary root under `t.TempDir()`, the names-only record lists CODEX_HOME, and the check prints `PASS: CODEX_HOME present in the codex child environment (names-only record: …)` with exit 0.

## AC-010 — A package that omits the sandbox fails the guard with a named finding (REQ-010)

**Out of Scope - moved to t1666.** Reason: its green path needs every reaching package's TestMain to call the sandbox helper, the AC-006 mechanism that moves with it (REQ-010). The text below is retained unchanged.

- Guard name (named now): `TestSandboxGuard_ReachingPackagesInstallSandbox`, in `internal/testhome/guard_test.go` (package `testhome`, created in M6).
- Guard design: the guard parses each reaching package's TestMain with `go/parser` and `go/ast`. It passes only when the TestMain body contains a call to the shared sandbox helper. A comment or a string that names the helper does not pass. The finding text is `SANDBOX-MISSING: <package dir> reaches a home resolver but its TestMain does not call the sandbox helper`.
- Verifying command: `go -C /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1630 test -count=1 -v -run '^TestSandboxGuard_ReachingPackagesInstallSandbox$' ./internal/testhome/`.
- Mutant procedure (run in M6): an overlay replaces one reaching package's TestMain with a comment-only body. The same command with `-overlay` must fail with the `SANDBOX-MISSING` finding for that package. The tree run must pass.
- Class: release-blocking. Minimum executed: N=1.
- RED cell (observed on 3975fe3cc):
  - (a) command: the verifying command above.
  - (b) stdout: empty. The go tool's messages came through the tool's error channel, which is stderr: `# ./internal/testhome` / `stat /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1630/internal/testhome: directory not found` / `FAIL	./internal/testhome [setup failed]` / `FAIL`.
  - (c) exit code: 1
  - (d) tree: 3975fe3cc25eb1cbc3be79aa16ff3bcd99f8ff21
  - Executed: 0. The red reason is the absent package.
- Green path: after M6 the guard runs and passes on the tree, and fails on the mutant.

## AC-011 — clean --home scans run and db items; a run item is a candidate only when unreferenced (REQ-011; card item e; Q3 DECIDED)

**Out of Scope - moved to t1666.** Reason: the clean --home run and db candidates, decision Q3 (REQ-011), moved by the operator-delegated split (spec.md HISTORY, revision 2). The text below is retained unchanged.

- Given: a moai home root holding one aged run/ entry that no live record references, and one aged db/ entry.
- When: the clean-home candidate scan runs (`scanHomeCleanable`, retention 30 days).
- Then: (i) at least one candidate lies under run/; (ii) the aged db/ entry is listed among the candidates. Clause (iii) of the decided rule (a run/ entry referenced by a live record is not a candidate; a db/ entry is listed without `--force` and deleted only with `--force`, the record's `--yes`) is a regression-guard, not release-blocking (plan.md G-21).
- Verifying command: `go -C /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1630 test -overlay /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1630/.moai/reports/t1630/evidence/overlay-ac011.json -count=1 -v -run '^TestAC011CleanHomeCandidatesIncludeRun$' ./internal/cli/`. Clause (ii) verifying command: `go -C /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1630 test -overlay /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1630/.moai/reports/t1630/evidence/overlay-ac011.json -count=1 -v -run '^TestAC011DBEntryListed$' ./internal/cli/`. Run-phase: the same assertions authored in the repository.
- Class: release-blocking for clauses (i) and (ii). Clause (iii): regression-guard, not release-blocking; the run phase authors its RED before its GREEN; reason: plan.md G-21. Minimum executed: N=1 per clause.
- RED cell, clause (i) (observed on 3975fe3cc):
  - (a) command: the verifying command above.
  - (b) stdout: `=== RUN   TestAC011CleanHomeCandidatesIncludeRun` / `    zz_ac011_test.go:34: no run/ candidate among 0 candidates` / `--- FAIL: TestAC011CleanHomeCandidatesIncludeRun (0.00s)` / `FAIL` / `FAIL	github.com/modu-ai/moai-adk/internal/cli	1.223s` / `FAIL`.
  - (c) exit code: 1
  - (d) tree: 3975fe3cc25eb1cbc3be79aa16ff3bcd99f8ff21
  - Executed: 1.
- RED cell, clause (ii) (observed on 3975fe3cc):
  - (a) command: the clause (ii) verifying command above.
  - (b) stdout: `=== RUN   TestAC011DBEntryListed` / `    zz_ac011_test.go:58: no db/ entry listed among 0 candidates` / `--- FAIL: TestAC011DBEntryListed (0.00s)` / `FAIL` / `FAIL	github.com/modu-ai/moai-adk/internal/cli	0.931s` / `FAIL`.
  - (c) exit code: 1
  - (d) tree: 3975fe3cc25eb1cbc3be79aa16ff3bcd99f8ff21
  - Executed: 1.
- Plan-phase category scan (context only, not the verdict): `internal/cli/clean_home.go` builds candidates with categories projects (lines 236, 249), debug (286), releases (293, 458), logs (306), and backups (329). No run or db category appears among them.
- Green path: after M4 the scan returns a run/ candidate for the aged, unreferenced entry and lists the aged db/ entry. Clause (iii) is authored and observed in run phase before its GREEN.

## AC-012 — Dry-run stays the default and --force deletes only allowlisted categories (REQ-012)

**Out of Scope - moved to t1666.** Reason: it guards the clean --home allowlist that REQ-012 changes; REQ-012 and its pointer AC-011 move, so this regression guard has no subject in this SPEC. The text below is retained unchanged.

- Given: a sandboxed home with deletable entries in allowlisted categories and carved-out segments.
- When: the existing clean-home regression tests run.
- Then: all three pass.
- Verifying command: `go -C /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1630 test -count=1 -v -run '^(TestCleanHome_DryRunMutatesNothing|TestCleanHome_ForceDeletesOnlyAllowlistedCategories|TestCleanHomeCarveOut_ForcePreservesCarvedSegments)$' ./internal/cli/`.
- Selector gate: N=3 executed. Zero executed is a failure.
- Class: regression-guard. It covers the existing categories only; the run and db rules of decided Q3 are checked by AC-011.
- Observed baseline (3975fe3cc, plan phase): stdout `=== RUN   TestCleanHomeCarveOut_ForcePreservesCarvedSegments` / `--- PASS: TestCleanHomeCarveOut_ForcePreservesCarvedSegments (0.02s)` / `=== RUN   TestCleanHome_DryRunMutatesNothing` / `--- PASS: TestCleanHome_DryRunMutatesNothing (0.01s)` / `=== RUN   TestCleanHome_ForceDeletesOnlyAllowlistedCategories` / `--- PASS: TestCleanHome_ForceDeletesOnlyAllowlistedCategories (0.01s)` / `PASS` / `ok  	github.com/modu-ai/moai-adk/internal/cli	0.836s`; exit code 0; tree 3975fe3cc; executed 3.

## AC-013 — The ordering landings are ancestors of the run base before run starts (REQ-013; gate)

- Given: the run base tip and the three landing SHAs the leader names for t1619, t1578, and t1591.
- When: each SHA is tested for ancestry against the run base.
- Then: each test exits 0 before any run-phase change begins.
- Verifying command: `git merge-base --is-ancestor <landing-sha> <run-base-tip>` for each of the three SHAs.
- Class: gate. Not release-blocking.
- Observed (revision 2, HEAD 5dc6c4530; the plan-phase observation at worktree base 2aab5f797 gave the same exit values; representative commits named in plan.md E-14):
  - `git merge-base --is-ancestor 8108eb256 HEAD` → `t1619 ancestor-of-HEAD exit=1`
  - `git merge-base --is-ancestor e73a7cbf5 HEAD` → `t1578 ancestor-of-HEAD exit=1`
  - `git merge-base --is-ancestor a372a984c HEAD` → `t1591 ancestor-of-HEAD exit=1`
- The gate is red at plan time. The leader names the landing SHAs (plan.md G-13).

## Edge cases

EC observation command (cited by EC-1 to EC-4; re-observed in revision 2 at HEAD 5dc6c4530; exit code 0): `go -C /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1630 test -overlay /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1630/.moai/reports/t1630/evidence/overlay-ac001-003.json -count=1 -v -run '^TestECObservations$' ./internal/config/toolpolicy/`. Raw stdout:

```
=== RUN   TestECObservations
    zz_ac001_003_test.go:92: EC-1 absent file: err=<nil> after="{\n  \"permissions\": {\n    \"defaultMode\": \"acceptEdits\"\n  }\n}"
    zz_ac001_003_test.go:100: EC-2 no permissions key: err=render "/var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/TestECObservations1851103078/002/settings.json": permissions block not found: no "permissions": key in body after="{\"env\":{\"FOO\":\"1\"}}"
    zz_ac001_003_test.go:108: EC-3 invalid JSON: err=render "/var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/TestECObservations1851103078/003/settings.json": permissions block not found: no "permissions": key in body after="{not json"
    zz_ac001_003_test.go:116: EC-4 duplicate allow entries: err=<nil> after="{\"permissions\": {\n    \"defaultMode\": \"acceptEdits\"\n  }}"
--- PASS: TestECObservations (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/config/toolpolicy	0.074s
```

- EC-1 — A settings file that does not exist yet. Observed (EC observation command above; the same probe as at 3975fe3cc, re-observed at HEAD 5dc6c4530): the writer returns no error and creates `{"permissions": {"defaultMode": "acceptEdits"}}`, with no lists. Bound to AC-001 (the run adds the expectation as an assertion).
- EC-2 — A settings file with no `permissions` key. Observed (EC observation command above): the writer returns `render … permissions block not found: no "permissions": key in body` and leaves the file unchanged (`{"env":{"FOO":"1"}}`). Bound to AC-001 (run-phase assertion of the error path). The earlier wording "no lists invented" is replaced by this observation.
- EC-3 — A settings file that is not valid JSON. Observed (EC observation command above): the writer returns `permissions block not found` and leaves the file unchanged (`{not json`). The error path is the observed behaviour, not a claim about the design. Bound to AC-001 (run-phase assertion).
- EC-4 — A list with duplicate entries. Observed (EC observation command above): the writer returns no error, and it removes the whole `allow` list (`{"permissions": {"defaultMode": "acceptEdits"}}`), which is the S-1 defect. The expectation that duplicates are preserved as written is unverified until the fix lands. Bound to AC-001 (run-phase assertion).
- EC-5 — A HOME path that contains spaces or is a symbolic link. Not observed in plan phase. Unverified. Bound to AC-007 as a run-phase fixture on a throwaway account. Out of Scope - moved to t1666 (with AC-007).
- EC-6 — A test package whose test files reference no home resolver. Not observed; the set is a name-based measure (plan.md E-9, G-9). Unverified. Bound to the REQ-007 exclusion rule and to AC-006's package set, and checked by the run-phase reach measure. Out of Scope - moved to t1666 (REQ-007 and the AC-006 package set).

## Quality gate criteria

- QG-1: every in-scope release-blocking criterion (AC-001 to AC-005) has an observed four-element RED cell on the pinned tree, and gate AC-013 is observed red at plan time. No pending allowance exists for a release-blocking criterion. The criteria moved to t1666 (AC-006 to AC-012) are not judged by this revision.
- QG-2: every criterion names its verifying command.
- QG-3: no in-scope criterion depends on the operator's home directory. The one criterion that did, AC-007, is Out of Scope - moved to t1666.
- QG-4: every named test exists, or is authored in run phase with its name recorded in progress.md.
- QG-5: in-scope criteria are six (AC-001 to AC-005 and AC-013), within the Tier M ceiling of 16. The file retains 13 criteria; AC-006 to AC-012 (seven) are Out of Scope - moved to t1666.

## Definition of Done

- The release-blocking criteria in scope (AC-001, AC-002, AC-003 parts A and B, AC-004 with count 1, and AC-005) are green on the post-fix tree, with verbatim output recorded. The criteria moved to t1666 (AC-006 to AC-012) are not part of this Definition of Done.
- Done criterion kept from card 3.2-0-1: the permissions-section diff across `moai init` is 0 when the existing defaultMode already matches the resolved tier default (AC-002). This is not measured in this SPEC: AC-002 measures the USER-scope writer only (plan.md R-1); the init-level observation is a Gap carried to the run phase.
- Out of Scope - moved to t1666: the regression guards AC-012 and AC-011 clause (iii).
- The gate (AC-013) is green before run starts and again before sync.
- The RED cells are re-observed on the post-landing run base at P-3 and M2, and both observations are recorded in progress.md.
- Out of Scope - moved to t1666: AC-007's before-and-after manifest on a throwaway account.
- Q1, Q2, and Q5 are DECIDED by the pinned board citation in decision-index.md. Q4 and Q6 are closed by the evidence produced at M2. Q3 and Q7 are Out of Scope - moved to t1666. G-17 and G-20 are closed by the lane in progress.md §G; G-18 stays a disclosed gap.
- Out of Scope - moved to t1666: the run-phase deletion criterion for run and db candidates (REQ-011, REQ-012); its RED is observed first, as a four-element cell, in the follow-up.
- S-7 (t1594) remains blocked on card text (M8). No requirement covers it in this revision.
