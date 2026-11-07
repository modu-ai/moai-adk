# t1702 progress, round 7 (lane-19, run to2bbk): the read-only profile-reach judgment

Leader dispatch (relayed): implement the read-only profile-reach judgment (RED first), have manager-spec draft the REQ-OWR-005 revision, inform lane-34 (t1642) of the interface, then one Claude audit.

## Setup
- `moai factory next --card t1702 --run to2bbk` -> `t1702 stage=- worktree=t1702`; tree /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1702, branch WT-profile-skill-reach, start HEAD 431f7c5a5d.

## Conflict from round 6 and how this round treats it
Round 6 stopped on: the branch already WRITES to profile roots (install/prune/settings cleanup) while the operator ruling (d-20261106T055331Z-c28f) wants the removal judgment to be read-only. Option (c) of that report is what was built: the NEW removal judgment is read-only; the existing writers were not touched. Whether the existing install-side writes should also go is still not decided and is not decided here.

## Done
1. `internal/userassets/profile_reach.go` (commit f5e05a105a): `CheckProfileReach(home, slug, rel, want [32]byte) ([]ProfileReach, error)`, `SafeToRemoveProjectCopy([]ProfileReach) bool`, states Same/Absent/Differs/Unknown (iota+1, the zero value is no state). Enumeration is `config.UserProfileDirNames` (junk filter shared with the hook). Per profile: Lstat the root (link or file -> Unknown, missing -> Absent), `os.OpenRoot`, `Lstat` the relative file inside it (missing -> Absent; link, directory, FIFO or a path leaving the root -> Unknown), bounded read (8 MiB, larger -> Unknown), sha256 compare. Only claude-skills and claude-agents have a profile root; any other slug returns no rows, and zero rows never licenses a removal. An unreadable profile store is an error, never an empty answer; an escaping `rel` is an error. Nothing is written or created.
2. Tests `internal/userassets/profile_reach_judgment_t1702_test.go`, 12 tests, all on a temporary HOME. RED first: before the implementation the package did not compile (`undefined: ProfileReach`, `CheckProfileReach`, `SafeToRemoveProjectCopy`, ...). GREEN: `go test ./internal/userassets -run 'TestCheckProfileReach_|TestSafeToRemoveProjectCopy_' -count=1 -v` -> 12 PASS; `-race` -> ok.
3. Mutants (each went RED, then was reverted): (a) removing the zero-rows guard of `SafeToRemoveProjectCopy` -> 3 tests failed (no-profiles, fixed slugs, unreadable store); (b) that guard back but `r.State != ReachSame` changed to `== ReachDiffers` -> 3 tests failed (absent, uncertain shapes, any-non-same). Final run after both reverts: ok with `-race`.
4. `gofmt -l internal/userassets` empty; `go vet ./internal/userassets` and `GOOS=windows go vet ./internal/userassets` print nothing.
5. REQ-OWR-005 revision proposal (manager-spec, `.moai/reports/t1702/req-owr-005-revision.md`). The SPEC (SPEC-UPDATE-OWNERSHIP-REPAIR-001) lives on card t1865's held branch; it was not edited. The proposal names three conflicts with the current text (read/stat ban; frozen call-site baseline and AC-OWR-08; the no-change list for `userroots.go` and `protected_zone.go`, which this branch changed for protection) and proposes: write ban kept absolute, one named read path (the two functions above), AC-OWR-08 as an allowlist scan.
6. lane-34 (t1642) was told the interface and acknowledged it; it waits for the final signature.

## Not wired
`CheckProfileReach` has no caller yet: the removal path and the `moai doctor` line are other cards (t1642 / stage 3). The judgment is tested alone.

## Gaps
- The whole `internal/userassets` package, `internal/cli`, `internal/hook` and `internal/config` suites were not run this round (leader's constraint); only the new tests plus vet.
- golangci-lint and `/verify` (the gate) were not run.
- Windows: vet only; the tests were not run there. The symlink cases use `t.Skipf` where links are unavailable.
- Not measured: behaviour on a case-insensitive filesystem; a profile root that is a link to another real profile root (reported Unknown by design, not tested against a real home).
- No real user home or profile was touched.
- The proposal text came from manager-spec; I read it but did not check it against the rest of the SPEC.
- Audit: none yet. The repo gate `audit.gates.codex: required` still applies. Decision d-20261106T064700Z-de2d is about 3.2 cards; t1702 was deferred to 3.3 (d-20261015T230738Z-e62a), so I do not treat it as covering a Codex call here. This round uses Claude only; if the Codex gate blocks a PASS, that is for the leader and operator.
