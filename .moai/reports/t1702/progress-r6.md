# t1702 progress, round 6 (lane-19, run to2bbk, home-asset migration stage 2)

Scope given by the leader (stages.md section "2단계"): reuse internal/config/userroots.go, repair audit iter5 F2[P1] and P3, enumerate several profiles read-only, a judgment that is uncertain keeps the project copy. Operator ruling relayed 11-06: profile reach is decided by READ-ONLY verification of the profile roots - before a project copy is removed, read each profile root and check it holds the same copy; absent or uncertain keeps the project copy; NOTHING is written to a profile root. Codex gate: Claude audit only until the operator decides after t2096.

## Setup
- `moai factory next --card t1702 --run to2bbk` -> `t1702 stage=- worktree=t1702`, tree /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1702, branch WT-profile-skill-reach.
- The tree was 2103 commits behind local main (8 ahead, last 0887e69728). Merged local main in (898a8999f1). Conflicts: `internal/cli/user_asset_phase.go` (both sides added a function - `userAssetRootsLine` from this card, `recordUserAssetFailures` from t2014 - both kept) and the three `internal/cli/testdata/doctor-*.golden` (taken from main, then regenerated with `UPDATE_GOLDEN=1` from the merged code; against main's goldens the only differences are the summary counts that add this card's doctor check). `go build` of cli/userassets/config/hook clean; userassets, config and the hook zone/roots tests ok after the merge.

## Done (all RED first)
1. iter5 F2 [P1] - commit 2881bf2a2b. `ZoneEntry.Match` for ZoneUserRoot now folds `<base>@<profile>` to `<base>` (bases from `config.UserRootProfileBases`: claude-skills, claude-agents only; empty base/profile, other bases and look-alike slugs never fold) so the shipped `user-root:claude-skills/` and `user-root:claude-agents/` entries cover profile-derived roots. Tests: `internal/config/protected_zone_profile_test.go` (shipped YAML parsed as the deny sites parse it; 5 profile cases RED before; controls for fixed roots and look-alikes) and `internal/hook/t1702_profile_root_enforcement_test.go` (the deployed YAML through `zoneUserRootCovered` with a temp HOME, a profile and a manifest: 3 cases RED before; an untracked profile file and a junk profile name stay uncovered). Mutants: no fold -> the end-to-end test RED; fold every base -> the `agents-skills`/`codex-agents` controls RED. The iter5 audit asked for exactly this: a test on the DEPLOYED YAML at `zoneUserRootCovered`.
2. iter5 P3 - commit 8476be4d3f. `removeConfigRootMoaiHooks` (new, in update.go) runs the check and the removal through one `os.Root` on the config root (Go 1.26.8): a swap of `hooks` for a link after the check can no longer reach an external tree; the F1 refusal of a link at either component is kept. `internal/cli/update_hooks_toctou_test.go` makes the window deterministic through a seam (`afterMoaiHooksCheck`): RED before (the external file was deleted), controls pass before and after (a real directory is removed; a linked `hooks` is refused). Existing `TestEnsureGlobalSettingsEnv*`, `HooksRemovalRadius`, profile-reach and init-resume tests ok.
3. F3-iter5 (Shape 2 of `update_home_radius_test.go` has no discriminating power) was not touched: the audit classed it optional and non-blocking.

## STOP - a conflict between the existing design and the operator ruling (needs a decision)
The card branch already WRITES to profile roots: `userassets.ClaudeConfigRoots(home)` / `RealProfileDirNames`, install and prune over every real profile directory (`internal/userassets/install.go`, `remove.go`, `profile_reach_test.go`), the settings cleanup that reaches `<profile>/settings.json` and `<profile>/hooks/moai` (`update.go`: `ensureGlobalSettingsEnv`), `migrate_profile_reach_test.go`, `update_profile_reach_test.go`, and the update summary `userAssetRootsLine`. That is the design the card text called for ("install/update reach every Claude config root"), and it is the option the stage brief listed as "install puts copies in every profile root". The operator ruling picks the OTHER option for the home-asset migration: read-only verification, no writes to a profile root. Whether the existing install/cleanup writes (a) stay because they belong to the user-asset install path (a different feature from the removal judgment), (b) are removed or put behind a switch so nothing writes to a profile root, or (c) stay for the install and only the NEW removal judgment is read-only, is not stated and I did not guess. I changed none of those writers. The two repairs above are protections (they make writes to profile roots HARDER) and are consistent with either answer.

## Proposed interface for the read-only reach judgment (proposal only - not implemented; for lane-34's t1642 doctor judgment)
Where: package `internal/userassets`, next to `ClaudeConfigRoots`, using `config.UserProfileDirNames` for the one junk-filtered enumeration. All reads, no writes, no creation of directories.

    // ProfileReach is the read-only answer for ONE project copy.
    type ProfileReach struct {
        Profile string          // profile directory name
        Root    string          // absolute profile skills|agents root that was read
        State   ReachState      // see below
        Reason  string          // one line, for the report
    }
    type ReachState int
    const (
        ReachSame      ReachState = iota // the profile root holds the same copy (content hash equal)
        ReachAbsent                      // the profile root exists, the copy is not there
        ReachDiffers                     // a copy is there but its content differs
        ReachUnknown                     // unreadable, a link that leaves the root, a non-regular file, a read error
    )
    // CheckProfileReach reads, for every REAL profile (UserProfileDirNames), the
    // profile root of slug and compares rel with want (the project copy's hash).
    func CheckProfileReach(home string, slug RootSlug, rel string, want [32]byte) ([]ProfileReach, error)
    // SafeToRemoveProjectCopy is true only when at least one profile exists AND
    // every profile answered ReachSame; any other state, an enumeration error,
    // or zero profiles returns false (the project copy stays).
    func SafeToRemoveProjectCopy(reaches []ProfileReach) bool

Rules this interface carries: reading goes through `os.Root` on each profile root (no escape through a link, same as the P3 repair); `Absent`, `Differs` and `Unknown` all keep the project copy; zero profiles keeps it too; the junk filter is the shared `ValidProfileDirName`; the comparison key is the content hash the migration judgment already computes, never the path or the name. Lane-34's doctor can call `CheckProfileReach` and print each `ProfileReach` row; the removal path calls `SafeToRemoveProjectCopy`.

## Not done yet
- The interface above is not implemented, nor the tests for it.
- REQ-OWR-005 revision text (manager-spec) - the ruling says to produce it on this card if the requirement needs a change; not started because the answer to the conflict above decides whether it does.
- Audit: none yet. One Claude `sync-auditor` round is the operator's rule; the repo setting `audit.gates.codex: required` still applies and the operator will decide after the t2096 report (t2070's audit failed for that reason).
- `moai factory stage` was not walked for this card (the stage is `-`).

## Gaps
- The whole `internal/cli` and `internal/hook` suites were not run (the leader's constraint); the sweeps above are targeted: config (whole package, 37 s ok), hook zone/roots tests (ok), cli update/global-settings/profile tests (ok), userassets (ok).
- golangci-lint and `/verify` (the gate) were not run. The three doctor golden files were regenerated, not hand-edited, and their diff against main was read (summary counts only).
- No real user home or profile was touched; every test uses temporary HOME and profile directories.
- Windows: not executed. `GOOS=windows go vet ./internal/config ./internal/hook ./internal/cli` printed nothing (compiles and vets); `go test -c` for Windows was not run.
