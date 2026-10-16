auditor-model: claude-sonnet-5-5
relay-note: byte-identical copy of card-review.md (written by the sync-auditor), placed under the sync-audit.md name the stage gate reads; relayed by lane-39.

# Card t1757 review, round 2 (Class B, no SPEC), cumulative 754752a1d + 226e99de0

verdict: PASS-WITH-DEBT
audited_sha: 226e99de01bc7527fce9e4ce2d4217332c90d2f4

Overall: PASS-WITH-DEBT. Both round-1 blockers are repaired and each repair was reproduced independently. Cross-model gate: codex (required) returned `pass`, receipt `rcpt-2502460cc153b79fa357194b`, `plan_source: config`, `gate_unmet` empty, claude/glm gates `off`. Debt: O1/O2/O3 from round 1, unchanged and judged non-blocking (see Findings). Real Windows behaviour remains unmeasured (Gap).

## Findings

- B1 [closed] doctor_windows_hook_shell.go:118-122. `hookShellWindowsDefault` now derives goos from `uikit.GoosArch()` (honours `MOAI_GOOS_OVERRIDE`), the same seam the goldens pin (`doctor_golden_test.go` sets `testos`). Reproduced the round-1 Windows-forcing overlay against the new tree: goldens pass (E4). The control mutant (default constructor forced to "windows", i.e. the round-1 behaviour) fails all four golden tests and the new test (E5).
- B2 [closed] doctor_windows_hook_shell.go:82-85. The default (non-verbose) `Message` carries the hedge: "hook launch will fail" (dependency fact) then "may block tool calls when a guard fails (inferred from the changelog wording \"hooks that set no shell\"; not confirmed for exec-form hooks, not measured on Windows)". The test asserts those hedge tokens on `Message`, not on `Detail` (test calls with verbose=true but inspects c.Message). Not self-contradicting: "will fail" concerns launching `bash`, "may block" concerns the inferred consequence.
- O1 [P2, optional, debt] `exec.LookPath("bash")` uses the doctor process PATH, not the hook-effective PATH; a System32 WSL `bash.exe` resolves OK. Can over- or under-warn. The message states "bash not found on PATH" which is literally what was measured (doctor PATH), so it is not misleading beyond the hedge. Unchanged.
- O2 [P3, optional, debt] countBashExecHooks counts every `command == "bash"` regardless of event/args/shell. Right for the shipped template; over-warns on customised configs. The consequence sentence is hedged ("may"), so no added dishonesty. Unchanged.
- O3 [optional, debt] no docs-site page for the new check. Not required by the card.
- N1 [P3, optional] "hook launch will fail" is stated firmly although it rests on O1's PATH assumption. Acceptable: it follows from "bash not found on PATH" and a Windows exec-form `bash` command; not measured on Windows (listed in Gaps).

## Evidence (five sections)

### Claim
C1 tree builds and vets for Windows. C2 target tests pass. C3 B1 fixed: goldens stable when the host OS is Windows. C4 new test discriminates. C5 B2 hedge in the default message. C6 repair commit touched only the two files claimed.

### Evidence
E1 `go build ./...` -> `build exit=0`, no output.
E2 `GOOS=windows go vet ./internal/cli/` -> `vet exit=0`, no output.
E3 `go test ./internal/cli/ -run 'HookShell|DoctorGolden|BinaryLag_DoctorCheckName' -count=1 -v` (filtered to result lines):
```
--- PASS: TestBinaryLag_DoctorCheckNameSetIsUnchanged (0.00s)
--- PASS: TestBinaryLag_DoctorCheckNameGuardRejectsMutations (0.01s)
--- PASS: TestDoctorGolden_Light (0.31s)
--- PASS: TestDoctorGolden_Dark (0.18s)
--- PASS: TestDoctorGolden_NoColor (0.22s)
--- PASS: TestDoctorGolden_IgnoresCallerCodexHome (0.23s)
--- PASS: TestHookShellWindows_BashMissing_Warns (0.00s)
--- PASS: TestHookShellWindows_BashPresent_OK (0.00s)
--- PASS: TestHookShellWindows_NoBashHooks_OK (0.00s)
--- PASS: TestHookShellWindows_NonWindows_OK (0.00s)
--- PASS: TestHookShellWindows_MissingSettings_Warns (0.00s)
--- PASS: TestHookShellWindows_DefaultHonorsGoosOverride (0.00s)
    --- PASS: .../non-windows_override_is_not_applicable (0.00s)
    --- PASS: .../windows_override_follows_the_windows_branch (0.00s)
--- PASS: TestHookShellWindows_RegisteredAsClaudeSurface (0.00s)
ok  	github.com/modu-ai/moai-adk/internal/cli	2.304s
```
E4 Windows-host simulation, production files untouched: `go test -overlay ov1.json` where the overlay replaces `internal/cli/uikit/banner.go` with a copy whose `goos := runtime.GOOS` is `goos := "windows"` (so `GoosArch()` with no override reports windows):
```
--- PASS: TestDoctorGolden_Light (0.37s)
--- PASS: TestDoctorGolden_Dark (0.28s)
--- PASS: TestDoctorGolden_NoColor (0.48s)
--- PASS: TestDoctorGolden_IgnoresCallerCodexHome (0.23s)
--- PASS: TestHookShellWindows_DefaultHonorsGoosOverride (0.00s)
ok  	github.com/modu-ai/moai-adk/internal/cli	2.547s
```
E5 mutant (overlay `ov2.json`, default constructor `goos := "windows"` ignoring the seam, i.e. round-1 behaviour):
```
--- FAIL: TestDoctorGolden_Light (0.64s)
--- FAIL: TestDoctorGolden_Dark (0.27s)
--- FAIL: TestDoctorGolden_NoColor (0.19s)
--- FAIL: TestDoctorGolden_IgnoresCallerCodexHome (0.20s)
    doctor_windows_hook_shell_test.go:115: got status=warn msg="not checked: .claude/settings.json not found", want OK 'not applicable'
--- FAIL: TestHookShellWindows_DefaultHonorsGoosOverride/non-windows_override_is_not_applicable (0.00s)
FAIL
```
E6 `git diff 754752a1d 226e99de0 --stat` -> two files only (`doctor_windows_hook_shell.go` 31, `doctor_windows_hook_shell_test.go` 22); `git status --short` empty after all overlay runs (overlay files live in the scratchpad).
E7 `moai verify audit-plan` -> `cross_model_active: true`, `enforced_required: ["codex"]`. `audit_multi` (target baseBranch, project_root = this tree) -> `overall_verdict: pass`, `gate_unmet`: none, `plan_source: config`, codex gate required verdict pass, findings none, `audit_receipt: rcpt-2502460cc153b79fa357194b`.

### Baseline-attribution
All commands ran in this session against tree /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1757, HEAD 226e99de01bc7527fce9e4ce2d4217332c90d2f4. The Go build/vet/test results come from the local Go toolchain on this tree. The `moai` CLI/MCP build that produced E7 is commit ec5a4f392, a strict ancestor of the tree HEAD (reported `build_lag`); its audit output describes the codex verdict on the diff, and the Go evidence is independent of that lag.

### Gaps
- Real Windows behaviour not measured (darwin host): whether exec-form `bash` fails to launch, whether a failing guard blocks tool calls, and whether Git Bash is absent. B1 was shown by GOOS injection through the seam, not on a Windows runner. The Windows multi-OS gate (`pr-multi-os-gate.yml`) remains the real check.
- The 2.1.299 changelog effect on exec-form hooks is still unobserved; the message hedges it and this audit confirmed the hedge, not the claim.
- Full `go test ./internal/cli/` and `./internal/template/` not re-run this round (template unchanged since round 1, where it passed); only the requested filtered run.
- No refused tool call this round except one compound command rejected by the worktree guard, split into plain commands. `.moai/lessons-inbox.jsonl` not inspected.

### Residual-risk
- Windows users with Git Bash installed but absent from the doctor process PATH still get a warning (O1); the warning text is hedged on consequence but states "not found on PATH" as measured.
- User-customised settings can over-warn (O2).

## Recommendations
1. O1/O2 as follow-up cards if the check proves noisy; O3 optional docs.
2. Let the Windows multi-OS gate confirm B1 on a real runner.

Iteration history: round 1 FAIL (754752a1d; B1 GOOS golden divergence, B2 over-reach; codex receipt rcpt-6dec858250736939aab29286). Round 2 PASS-WITH-DEBT (226e99de0; B1/B2 closed, O1-O3 carried as debt; codex receipt rcpt-2502460cc153b79fa357194b).

AUDIT-VERDICT: PASS-WITH-DEBT spec=t1757 receipts=rcpt-2502460cc153b79fa357194b
