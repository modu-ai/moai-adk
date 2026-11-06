# m0-baseline.md — SPEC-AGENT-CORE4-001 (piece 1, card t1649)

This file is the M0 baseline. It is the only file in the M0 commit, so the commit graph witnesses
that every measurement below was taken before any M1 or M2 change (`verification-claim-integrity.md`
§2.3). The M0 commit SHA cannot appear here (a commit does not know its own hash); it is the
commit that introduces this file, and the M1 run records it in `progress.md` §E.2.

## 1. Baseline tree

- Baseline = the parent of the M0 commit, `<M0>^`:
  `da248a38da6e5cdb1257642cb5b2c473220d0954` (`Merge main into WT-3-3-10 (card t1649, absorb before M0 baseline)`).
- Branch `WT-3-3-10`, worktree `.moai/worktrees/t1649`. Local `main` (`23296dcc98`) is an ancestor of the baseline
  (`git merge-base --is-ancestor main da248a38da` exit 0), so foreign commits on the guarded paths are inside it.
- Toolchain for every cell below: `go version go1.26.8 darwin/arm64`. No installed `moai` binary produced any cell.
- Every command was run unpiped through `bash -c` from the worktree root; each `exit=` is the real exit code.

## 2. M0 abort rule — the eight old lines of spec §C.1 against the baseline C2 files

Each old line of spec.md §C.1 (spec.md lines 107, 112, 117, 122, 127, 132, 137, 142 with the diff marker `- `
stripped) was compared for byte equality with the named line of the baseline C2 file
(`internal/template/templates/.claude/agents/moai/<file>`). All eight matched, so M0 does not abort:

```text
MATCH builder-harness.md:38 (spec L107)
MATCH builder-harness.md:155 (spec L112)
MATCH e2e-tester.md:92 (spec L117)
MATCH e2e-tester.md:143 (spec L122)
MATCH manager-design.md:50 (spec L127)
MATCH manager-design.md:203 (spec L132)
MATCH manager-design.md:206 (spec L137)
MATCH super-advisor.md:129 (spec L142)
```

## 3. Ledger commands (acceptance.md §2) re-run on the baseline

Same commands as the `acceptance.md` §2 ledger (measured there on `06095bb8fa`); outputs below are this run's,
on `da248a38da`. Only the elapsed times differ from the ledger.

```text
$ cmp .claude/agents/moai/builder-harness.md internal/template/templates/.claude/agents/moai/builder-harness.md
.claude/agents/moai/builder-harness.md internal/template/templates/.claude/agents/moai/builder-harness.md differ: char 2170, line 38
exit=1
$ cmp .codex/agents/moai/builder-harness.toml internal/template/templates/.codex/agents/moai/builder-harness.toml
.codex/agents/moai/builder-harness.toml internal/template/templates/.codex/agents/moai/builder-harness.toml differ: char 2123, line 32
exit=1
$ cmp .claude/agents/moai/e2e-tester.md internal/template/templates/.claude/agents/moai/e2e-tester.md
.claude/agents/moai/e2e-tester.md internal/template/templates/.claude/agents/moai/e2e-tester.md differ: char 5529, line 92
exit=1
$ cmp .codex/agents/moai/e2e-tester.toml internal/template/templates/.codex/agents/moai/e2e-tester.toml
.codex/agents/moai/e2e-tester.toml internal/template/templates/.codex/agents/moai/e2e-tester.toml differ: char 5468, line 85
exit=1
$ cmp .claude/agents/moai/manager-design.md internal/template/templates/.claude/agents/moai/manager-design.md
.claude/agents/moai/manager-design.md internal/template/templates/.claude/agents/moai/manager-design.md differ: char 2213, line 50
exit=1
$ cmp .codex/agents/moai/manager-design.toml internal/template/templates/.codex/agents/moai/manager-design.toml
.codex/agents/moai/manager-design.toml internal/template/templates/.codex/agents/moai/manager-design.toml differ: char 2171, line 44
exit=1
$ cmp .claude/agents/moai/super-advisor.md internal/template/templates/.claude/agents/moai/super-advisor.md
.claude/agents/moai/super-advisor.md internal/template/templates/.claude/agents/moai/super-advisor.md differ: char 8289, line 129
exit=1
$ cmp .codex/agents/moai/super-advisor.toml internal/template/templates/.codex/agents/moai/super-advisor.toml
.codex/agents/moai/super-advisor.toml internal/template/templates/.codex/agents/moai/super-advisor.toml differ: char 8055, line 124
exit=1
$ ls internal/template/agentemit/source
ls: internal/template/agentemit/source: No such file or directory
exit=1
$ make agents-emit-check
ok  	github.com/modu-ai/moai-adk/internal/template/agentemit	0.187s
exit=0
$ go test ./internal/template/agentemit -run '^TestSourceSetIsOneToOne$' -v -count=1
testing: warning: no tests to run
PASS
ok  	github.com/modu-ai/moai-adk/internal/template/agentemit	0.182s [no tests to run]
exit=0
$ go test ./internal/template/agentemit -run '^TestClaudeRenderFromSource$' -v -count=1
testing: warning: no tests to run
PASS
ok  	github.com/modu-ai/moai-adk/internal/template/agentemit	0.199s [no tests to run]
exit=0
$ go test ./internal/template/agentemit -run '^TestCodexRenderFromSource$' -v -count=1
testing: warning: no tests to run
PASS
ok  	github.com/modu-ai/moai-adk/internal/template/agentemit	0.216s [no tests to run]
exit=0
$ go test ./internal/template/agentemit -run '^TestRenderUnmappedCapabilityFails$' -v -count=1
testing: warning: no tests to run
PASS
ok  	github.com/modu-ai/moai-adk/internal/template/agentemit	0.198s [no tests to run]
exit=0
$ go test ./internal/template/agentemit -run '^TestCarriedFourCopiesEqual$' -v -count=1
testing: warning: no tests to run
PASS
ok  	github.com/modu-ai/moai-adk/internal/template/agentemit	0.186s [no tests to run]
exit=0
$ go test ./internal/template/agentemit -run '^TestEmitCheckFailsOnHandEdit$' -v -count=1
testing: warning: no tests to run
PASS
ok  	github.com/modu-ai/moai-adk/internal/template/agentemit	0.190s [no tests to run]
exit=0
$ go test ./internal/template/agentemit -run '^TestGoldenMarkdownMismatchDetected$' -v -count=1
testing: warning: no tests to run
PASS
ok  	github.com/modu-ai/moai-adk/internal/template/agentemit	0.180s [no tests to run]
exit=0
$ go test ./internal/harness/rosterguard -run '^TestNumeralSourceDirDischarged$' -v -count=1
testing: warning: no tests to run
PASS
ok  	github.com/modu-ai/moai-adk/internal/harness/rosterguard	0.275s [no tests to run]
exit=0
$ go test ./internal/template -run '^TestAgentsLocalSection20Text$' -v -count=1
testing: warning: no tests to run
PASS
ok  	github.com/modu-ai/moai-adk/internal/template	0.417s [no tests to run]
exit=0
$ grep -c 'Distribution mirror, hand-edited' AGENTS.local.md
1
exit=0
```

## 4. Blob ids of the 24 C2/C3 files at `<M0>^`

Command: `git ls-tree -r da248a38da6e5cdb1257642cb5b2c473220d0954 -- internal/template/templates/.claude/agents/moai internal/template/templates/.codex/agents/moai`
(24 lines; exit 0).

```text
100644 blob becfbbe3f0b99c1ec2d012c06e34537c37314e60	internal/template/templates/.claude/agents/moai/builder-harness.md
100644 blob 0a5f4f61f3fc34dbdc60e5c6fa3058ced5e2228c	internal/template/templates/.claude/agents/moai/e2e-tester.md
100644 blob de86e8a4359c9d1c1abb8a90181bcfe526cc21f6	internal/template/templates/.claude/agents/moai/manager-design.md
100644 blob 8f5edfbbed0f426d97573f647fe52da394cd30ab	internal/template/templates/.claude/agents/moai/manager-develop.md
100644 blob 00130e576b06c5442fc790bd61abccdb37c96890	internal/template/templates/.claude/agents/moai/manager-docs.md
100644 blob ca21a45b6f049972267de898e166b879df94d825	internal/template/templates/.claude/agents/moai/manager-git.md
100644 blob 7c047b382b7699e44be21a0f31bf84db126150b0	internal/template/templates/.claude/agents/moai/manager-lead.md
100644 blob 9e1c3ac4dceb1cd8597758e2a65fd2b677df2437	internal/template/templates/.claude/agents/moai/manager-spec.md
100644 blob 2a1c7afa6430cd769ba26dcad0d212cf1cad53d6	internal/template/templates/.claude/agents/moai/manager-todo.md
100644 blob 3ced1482d51663a67c566d63b25f95557f5d7632	internal/template/templates/.claude/agents/moai/plan-auditor.md
100644 blob 0e8e8ff11130ab511924d337cdb53d706fb63a34	internal/template/templates/.claude/agents/moai/super-advisor.md
100644 blob 3c945933f765f334fb1e33e20c99a6938a8f707e	internal/template/templates/.claude/agents/moai/sync-auditor.md
100644 blob 7c2a850ae0806ec81a797eee5bb221d65288d902	internal/template/templates/.codex/agents/moai/builder-harness.toml
100644 blob 2dff73b1d563145e93543a70853f19cd183144c8	internal/template/templates/.codex/agents/moai/e2e-tester.toml
100644 blob 7abbceaffbd4c6b7a3fd993f7b905502110369ec	internal/template/templates/.codex/agents/moai/manager-design.toml
100644 blob 3e809e75adecf1251af72a3b0d1940dfe3edb1bc	internal/template/templates/.codex/agents/moai/manager-develop.toml
100644 blob b0f12ca5a3d78a287b803a8711ded1dafa406c28	internal/template/templates/.codex/agents/moai/manager-docs.toml
100644 blob d1f5d8c6233b61376f765524194fb5806631c3e6	internal/template/templates/.codex/agents/moai/manager-git.toml
100644 blob 428751c64f7a32cb98713bbaeee676324ffd8dc9	internal/template/templates/.codex/agents/moai/manager-lead.toml
100644 blob d4fbffd26f366fdb6ea0a7e8d9254e844f403808	internal/template/templates/.codex/agents/moai/manager-spec.toml
100644 blob b94348d1559ca3d0e4a2f60cdf0f9ec4c075dd51	internal/template/templates/.codex/agents/moai/manager-todo.toml
100644 blob f9a5b64b30c636ce7d2c334ad9633dcd1f16a3c7	internal/template/templates/.codex/agents/moai/plan-auditor.toml
100644 blob a080098313130d50c123f2445d07b0ce3b71b1f7	internal/template/templates/.codex/agents/moai/super-advisor.toml
100644 blob 3a5313201eb31790c0c4f8ba9983ab54beed0c2a	internal/template/templates/.codex/agents/moai/sync-auditor.toml
```

## 5. The 30 test names of the agentemit package at `<M0>^`

Command: `go test ./internal/template/agentemit -list '^Test'` (exit 0). AC-AC4-011(b) requires every name below to
remain in the list at M1 and M2.

```text
TestEmitAllFlatPrefixLayoutFallback
TestEmitAllSandboxEmittedWhenConfirmed
TestRoleSandboxRejectsWriteToolsAndInvalidModes
TestParseManifestFailClosed
TestParseAgentDocStructuralErrors
TestEmitAllRejectsUnrepresentableBodies
TestParseAgentDocParsesFrontmatterContract
TestParseAgentDocRejectsBrokenSources
TestEmitAllRoundTripsBodyVerbatim
TestEmitAllMCPServerMapping
TestEmitAllModelAndEffortFromFrontmatter
TestEmitAllCodexModelEffortRules
TestEmitAllRejectsMaxUltraAndUnmappedModel
TestEmitAllOmitsSandboxMode
TestEmitAllDeterministic
TestEmitAllMarkdownIdentityIsPassThrough
TestEmitAllFailClosedNegatives
TestEmitAllFailClosedDuplicateName
TestLoadManifestSelfValidates
TestAuditRoleLauncherInstructionSurface
TestCodexAuditRolesReadOnlyScopedException
TestGoldenCommittedArtifactsMatchEmission
TestRealSetMarkdownTreeUnmodified
TestRealSetCodexShape
TestRealSetBodiesByteEqual
TestEmbedFSPresenceAndByteEquality
TestNeutralityByInheritance
TestPublicationPathHygiene
TestRolePermissionContractNoSilentDrop
TestRolePermissionUnsupportedNeverPass
ok  	github.com/modu-ai/moai-adk/internal/template/agentemit	0.224s
```

## 6. Roster guard numeral layer at `<M0>^`

Narrowed run, not the package-wide suite (the package-wide run needs the shared go-test-heavy slot, which this
card's lane controls): `go test -count=1 ./internal/harness/rosterguard/ -run '^TestNumeral'` (exit 0).

```text
ok  	github.com/modu-ai/moai-adk/internal/harness/rosterguard	30.825s
```

## 7. Gaps

- The package-wide rosterguard suite (`go test ./internal/harness/rosterguard/...`) and the package-wide
  `./internal/template` suite were NOT run at M0: only the narrowed selectors above were. The repository-wide test verdict
  belongs to CI on the integration branch and is PENDING.
- The C1 (`.claude/agents/moai`) and C3p (`.codex/agents/moai`) blobs of the carried four are not recorded here;
  the ledger cells in section 3 carry their differences from C2/C3 as `cmp` outputs.
