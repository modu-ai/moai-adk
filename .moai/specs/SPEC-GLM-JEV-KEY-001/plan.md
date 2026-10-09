---
id: SPEC-GLM-JEV-KEY-001
title: "Implementation plan — glm --key flag and moai jev command"
created: 2026-10-09
updated: 2026-10-10
---

# SPEC-GLM-JEV-KEY-001 — plan.md

## §A. Context

Factory card t1613 (Class C — key command cleanup), DOC-ONLY plan phase. Two deliverables: a `--key` persistent-save form on `moai glm`, and a new top-level `moai jev` command with `--key`. Downstream card t1612 (terminal-output redesign) is explicitly out of scope.

Code anchors verified by manager-spec in this tree at HEAD f7606c7bc (2026-10-09):

- `internal/cli/glm.go:39-118` — `glmCmd` (`Use: "glm [-p profile] [-f | -l] [-- claude-args...]"`, `GroupID: "launch"`, **`DisableFlagParsing: true`**, `RunE: runGLM`).
- `internal/cli/glm.go:142-147` — `init()` registers `glmSetupCmd`, `glmStatusCmd` on `glmCmd` and `glmCmd` on `rootCmd`.
- `internal/cli/glm.go:187-200` — manual subcommand routing inside `runGLM` (`case "setup": / "status": / "tools":`), running before all launch parsing; `glm.go:209` comment: "Placed after subcommand routing so `moai glm setup` is never intercepted".
- `internal/cli/glm.go:479-503` — `runGLMSetup` (trims, empty-key error, `saveGLMKey`, prints `GLM API key stored (%s)`); `glm.go:505-511` — `maskAPIKey` (first4 + `****` + last4; `****` for ≤8 chars).
- `internal/cli/glm.go:952-977` — `getGLMEnvPath`/`saveGLMKey`/`loadGLMKey` delegate to `internal/glmcred` (mode 0600 write + legacy wide-mode tightening; `MOAI_TEST_GLM_KEY` test seam; `glmcred.HomeDirFn` aliased to this package's `userHomeDirFn` in `glm.go:35-37`).
- `internal/jevcred/jevcred.go` — the TypeSafe credential SSOT already exists, deliberately modeled on glmcred: `Save` (mode 0600 + explicit chmod tightening, quoted dotenv form), `Load` (`MOAI_TEST_TYPESAFE_KEY` seam), `Path` (`~/.moai/.env.typesafe`, absolute-`MOAI_HOME` redirect), `View` (bounded disclosure: configured + final 4 chars, empty hint at ≤4 chars), dotenv key `TYPESAFE_API_KEY` (unexported), no reveal route.
- `internal/cli/mcp_jev.go:87`, `internal/cli/doctor_jev.go:82,88`, `internal/web/fieldsets_templ.go:1928`, `internal/web/handlers.go:121` — existing guidance/surfaces already name `~/.moai/.env.typesafe`.
- `internal/config/envkeys.go` — env-var name SSOT (`EnvHome = "MOAI_HOME"` at line 22; `EnvTestGLMKey`, `EnvTestTypeSafeKey` test seams).
- No `jev` command exists anywhere in `internal/cli` (grep for `jevCmd` / `Use:.*jev` → 0 rows in non-test files); `internal/cli/root.go:130-134` defines groups `launch` / `project` / `tools`.
- Stack: cobra v1.10.2 + fang wrapper (`.moai/project/tech.md:36-37`).

## §B. Known Issues

1. **`moai glm setup --key x` trap (pre-existing class, not fixed here).** With manual routing, `setup` receives raw args, so a value starting with `-` is stored verbatim today (`moai glm setup -p x` already behaves this way). Adding `--key` at the parent level makes `moai glm setup --key x` a plausible-but-wrong user spelling that would store the literal `--key`. The card's constraint ("`moai glm setup <key>` keeps working unchanged") forbids touching `runGLMSetup`, so this trap is documented, not fixed. t1612's guidance surface must not advertise `--key` beside `setup`.
2. **`glmCmd` has `DisableFlagParsing: true`** — a cobra flag definition alone would never be parsed for the glm parent. The `--key` handling must be a manual scan inside `runGLM` (same mechanism as the existing `--help` scan), not a `Flags().String` registration expected to parse itself. A flag definition on `glmCmd` is still added solely so the flag renders in help output.
3. **`moai jev` bare invocation has no prior behavior to preserve** — the command is new; its bare shape is pinned in §D.2, not inherited.
4. **Legacy `setup` mishandles embedded newlines today (measured, left unchanged).** `runGLMSetup` only `TrimSpace`s its argument (an embedded CR/LF survives the trim), `glmcred.EscapeValue` deliberately does not escape newlines (glmcred.go:135-138), and `glmcred.Load` reads line-by-line — so `moai glm setup` given an embedded-newline value stores a credential that reads back as its first line only, with the success message printed. The card's unchanged-setup constraint forbids fixing it here; the NEW `--key` paths refuse such values loudly instead (REQ-GJK-012). The acceptance-behavior divergence between the two glm save forms is deliberate and documented.
5. **glm separated-value check differs from the rule (resolved in run milestone M4; the rule is not relaxed).** The glm scan (glm.go:559-570) refuses a dash-leading separated token by a raw prefix test, and that test exempts the bare `-` (`next != "-"`). So `moai glm --key -` stores `-`, and `moai glm --key " -f"` passes the test and is stored trimmed as `-f`. §D.1 states the rule on the trimmed value, as `runJev` does (jev.go:68). Resolution: milestone M4 (§F) trims the separated token before the dash test and removes the bare-`-` exemption, so `moai glm --key -` and `moai glm --key " -f"` are refused with a usage error. REQ-GJK-004 is kept as written; the rule is not relaxed, and no bare-`-` exemption is admitted.

## §C. Pre-flight (run-phase re-verification)

Before the first commit, re-verify these still hold (cheap greps, not memory):

1. `grep -n 'case "setup":' internal/cli/glm.go` → the manual routing still precedes launch parsing.
2. `grep -rn "jevCmd" internal/cli/ --include='*.go' | grep -v _test` → still 0 rows before M1 (so the RED tests fail for the right reason).
3. `grep -n "func Save\|func Load\|func Path\|func View" internal/jevcred/jevcred.go` → signatures unchanged.
4. Re-read `git rev-parse --short HEAD` + `git branch --show-current` immediately before any commit (staleness rule).

## §D. Design Decisions (pinned)

Ordered by decision-reversibility: user-facing flows first, storage and mechanical items after.

### §D.1 Routing composition for `moai glm --key` (user-facing flow)

`glmCmd` keeps `DisableFlagParsing: true`. Inside `runGLM`, after the existing manual subcommand switch (`setup`/`status`/`tools`) and before `guardCGLaunchMode` / spawn-strip / profile parsing, insert a `--key` scan over `args`:

- Accepts `--key <value>` and `--key=<value>`. The separated form refuses a token whose trimmed value begins with `-`: a dash-leading token after `--key` is never taken as the value, so `moai glm --key -f` returns a usage error and stores nothing (REQ-GJK-004; AC-GJK-011). The `=` form is the per-command exception: `--key=-f` binds `-f` without consuming a following token and is stored, while jev refuses the same value (§D.2). The gap between this rule and the raw-token check at glm.go:559-570 is recorded in §B.5.
- The scan stops at the first bare `--`, mirroring the `--help` scan's precedent (glm.go:166-174): tokens after `--` are child passthrough (`-- claude-args...`) and are never scanned for `--key` — `moai glm -- --key x` passes the tokens to the child and stores nothing.
- Scan placement AFTER the subcommand switch satisfies REQ-GJK-006: `moai glm setup <key>` is routed before the scan can ever see it (same ordering rationale as the `--spawn` strip and its comment at `glm.go:209`).
- Scan placement BEFORE launch parsing satisfies REQ-GJK-002: a `--key` invocation stores and returns — it never reaches `guardCGLaunchMode`, entry parsing, or `exec`.
- Any additional token beside `--key <value>` → usage error naming the conflict, store nothing (REQ-GJK-003). Missing value → usage error (REQ-GJK-004). Empty-after-trim value → the setup path's `empty API key` error (REQ-GJK-005).
- No shorthand. `-k` is retired vocabulary on this command (glm.go Long, genealogy note).
- A `--key` entry is added to `glmCmd.Flags()` only so it renders in help output (criterion a); parsing stays manual.
- The `--key` value carries the same input validation as the jev path (§D.2): a trimmed value containing `\r` or `\n` is refused with a validation error before `saveGLMKey` is invoked — existing file untouched, nothing stored. The card's "same storage, same value" constraint binds what the two forms WRITE, not what they ACCEPT: a value the owning packages cannot round-trip (`EscapeValue` deliberately does not escape newlines; `Load` reads line-by-line) must not enter storage through a new path. Legacy `setup` keeps today's behavior (§B4).

### §D.2 New `moai jev` command shape (user-facing flow)

- New file `internal/cli/jev.go`: `jevCmd = &cobra.Command{Use: "jev [--key <api-key>]", Short: "Manage the Jev (TypeSafe) credential", RunE: runJev}`, registered in `init()` via `rootCmd.AddCommand(jevCmd)`, `GroupID: "tools"` (existing groups: launch/project/tools — a credential utility is not a launch command).
- Cobra parsing (NO `DisableFlagParsing`): `jevCmd.Flags().String("key", "", "Store the TypeSafe API credential")`. Flag parsing alone is not enough: pflag v1.0.10 takes the token after `--key` as the value unconditionally (flag.go:1020-1023), so `moai jev --key --help` would store the literal `--help`. The explicit refusal therefore lives in `runJev`: a trimmed value beginning with `-` is refused with a usage error in every spelling (`--key -f`, `--key --help`, `--key=-f`) before the writer is invoked (REQ-GJK-013; AC-GJK-011).
- `runJev`: flag set → trim; empty → error; leading `-` → usage error (the refusal above); **CR/LF check** — a trimmed value containing `\r` or `\n` is refused with a validation error BEFORE the writer is invoked, so the existing credential file stays byte-for-byte unchanged (no partial write; reject-before-write, mirroring `internal/web/jevkey.go` `validateJevKey`'s trim-then-`ContainsAny("\r\n")` ordering); then `jevcred.Save(value)`; print bounded confirmation (§D.3); return nil. Flag unset → `cmd.Help()`, return nil, exit 0 (REQ-GJK-009). No subcommands.

### §D.3 Confirmation disclosure

- glm `--key` path reuses the setup path's exact confirmation string and mask (`GLM API key stored (%s)` via `maskAPIKey`) — zero new glm strings, identical user-visible contract as `setup`.
- jev confirmation uses `jevcred.View()`'s bounded disclosure (configured + final 4 characters; a credential of ≤4 characters discloses nothing), per REQ-GJK-008. This diverges from `maskAPIKey` (first4+last4) on purpose: REQ-JEVC-020 is the jev domain's already-asserted disclosure contract and `jevcred` is its SSOT. The divergence is pre-existing domain policy, not new divergence.

### §D.4 Storage decisions (the pin)

| Axis | glm (`--key` and `setup` — identical) | jev (`--key`) |
|---|---|---|
| File path | `~/.moai/.env.glm` (`glmcred.Path()`) | `~/.moai/.env.typesafe` (`jevcred.Path()`) |
| Env-var / dotenv key | `GLM_API_KEY` (owned by `internal/glmcred`) | `TYPESAFE_API_KEY` (owned by `internal/jevcred`) |
| File format | dotenv, quoted value: `GLM_API_KEY="<escaped>"` | dotenv, quoted value: `TYPESAFE_API_KEY="<escaped>"` |
| File mode | 0600 at write + explicit chmod tightening of a pre-existing wider mode | same (jevcred.Save) |
| Writer | `glmcred.Save` (single writer, SPEC-GLM-KEY-INPUT-001 D-1) | `jevcred.Save` (single writer, SPEC-JEV-CORE-001 REQ-JEVC-018) |
| Test seam | `userHomeDirFn` override (aliased to `glmcred.HomeDirFn` at glm.go:35-37) + `MOAI_TEST_GLM_KEY` | `jevcred.HomeDirFn` (exported var) or absolute `MOAI_HOME` (`config.EnvHome`, envkeys.go:22) + `MOAI_TEST_TYPESAFE_KEY` |

**Guidance-line disposition (delegation question):** NONE of the existing guidance lines change. `mcp_jev.go:87`, `doctor_jev.go:82/88`, and the web fieldsets already name `~/.moai/.env.typesafe`, which is exactly the file the new command writes through `jevcred`. The glm.go Long text gains a `--key` line (addition, §D.7); its existing `Use 'moai glm setup <key>'` sentence stays.

### §D.5 Naming and constants

- No new environment-variable name is introduced; `envkeys.go` gains nothing. The dotenv key names stay package-owned (`glmcred` / `jevcred`), and new `internal/cli` code must not spell them as literals (enforced by the diff-scoped AC-GJK-013 grep — pre-existing rows in untouched files are out of scope).
- If run phase discovers a genuinely needed new env-var name, its constant belongs in `internal/config/envkeys.go` (hardcoding-prevention rule) — none is expected.

### §D.6 Test seams and isolation

- Tests redirect the credential location via the existing seams only — never the developer's real `$HOME`: override `userHomeDirFn` (glm family precedent) and `jevcred.HomeDirFn` (exported var), or use an absolute `MOAI_HOME` pointing at `t.TempDir()`.
- Go test isolation rules apply: `t.TempDir()`, no OTEL env in parallel tests (hns-moaiadk-best-practices).
- Launch seam: `launchClaudeFunc` (launcher.go:575; signature `(profile, args)`) is the seam AC-GJK-004 (zero invocations on the save path), AC-GJK-009 (zero invocations on the `--key K --help` precedence case), and AC-GJK-016 (the recorded argument list) use. The launcher appends a `--settings` pair and an optional `--effort` pair after the passthrough (`appendCrossSessionSettings`, crosssession_settings.go:86). AC-GJK-016's equality therefore needs both injection inputs inert: redirect `findProjectRootFn` (cc.go:21) to an empty `t.TempDir()`, as glm_test.go:61-63 does, and return no effort from `launchEffortPrefsFn` (launch_effort_settings.go:49). Without the redirect, the project's own `crosssession.yaml` (`inbound: "accept"` in this tree) would inject a `--settings` pair. The expected recorded list is `["--", "--key", K]`, separator included: the separator reaches the seam (`parseProfileFlag` keeps `--` and everything after it, launcher.go:879-882), and the only separator strip on this path (launcher.go:661-664, inside `runLaunchClaude`) runs after the seam. Env seed: set `MOAI_TEST_GLM_KEY` (`glmcred.EnvTestGLMKey`) to the test key K with `t.Setenv` before the run. Without a key, `runUnifiedLaunch` fails in `applyGLMMode` ("GLM API key not found", launcher.go:168, 286) before the launch call (launcher.go:247), so the seam records nothing; the seed is read first (glmcred.go:101) and creates no credential file.

### §D.7 User-facing strings (enumeration; 4-locale note)

New or touched strings, all English literals (the CLI has no message catalog; existing glm messages are literal English):

1. `glm.go` Long help: +1 flag line under `Flags:` (`--key <api-key>` — store the GLM API key and exit, same storage as `moai glm setup <key>`) and +1 examples line (`moai glm --key sk-xxx`). The existing `Use 'moai glm setup <key>'` sentence is unchanged.
2. `jev.go`: `Use` line, `Short` ("Manage the Jev (TypeSafe) credential"), `Long` help (mentions `--key` and the `~/.moai/.env.typesafe` location), `--key` flag usage string.
3. `jev.go` confirmation: e.g. `Jev credential stored` + bounded hint suffix (format per §D.3).
4. Error strings: glm `--key` usage-error (missing value / extra tokens), jev empty-credential error.

4-locale note: t1612's renderer owns en/ko/ja/zh parity for its surfaces. This card only needs the strings enumerable (this section is the enumeration); no translation work is planned and no i18n catalog exists for CLI messages today.

## §E. Self-Verification (of this plan)

- Every code anchor in §A was read from this tree by manager-spec this run (not carried from the delegation text): glm.go regions 25-234 / 470-529 / 940-1008, jevcred.go full, envkeys.go full, mcp_jev.go 55-109, handlers.go 95-154, root.go 125-145.
- SPEC ID regex check executed as Bash, verbatim output `PASS` (2026-10-09).
- Frontmatter of spec.md validated against the 12-field canonical schema (`.claude/rules/moai/development/spec-frontmatter-schema.md`); `phase: "v3.2.0"` is a release target (latest tag v3.1.2; v3.2.0 is the in-flight release), not a stage name.
- ID uniqueness: no `SPEC-GLM-JEV-KEY-*` in `.moai/specs/` (directory listing, 2026-10-09).
- No open clarification markers remain: none of the decisions below needs operator input — storage and disclosure are decided by completed SPEC-JEV-CORE-001; the two implementation-level rows carry defaults (decision-index.md Q3/Q4, DEFAULT-APPLIED).

## §F. Milestones

Run order: M1, M2, M4, M3. M4 changes production code (`internal/cli/glm.go`), so M3's verification runs after it, on the final tree.

### M1 (Priority High) — RED-first key-save surfaces

- RED (new test file `internal/cli/jev_key_test.go`, plus glm `--key` cases alongside or in the same file): `TestRootHelpListsJevCommand`, `TestJevHelpDocumentsKeyFlag`, `TestJevKeyFlagSavesCredential`, `TestJevBareInvocationPrintsHelpExitZero`, `TestJevKeyEmptyValueErrors`, `TestGlmKeyFlagSavesKeySameStorage`, `TestGlmKeyFlagRefusesExtraArgs`, `TestGlmKeyFlagMissingValueErrors`, `TestGlmKeyEmptyValueErrors`, `TestGlmKeyExecFlagAsValueRefused`, `TestGlmKeyEqualsFormSaves`, `TestJevKeyOptionTokenValueRefused`, `TestGlmHelpDocumentsKeyFlag`, `TestGlmKeyFlagNewlineValueRefusesAndPreserves`, `TestJevKeyNewlineValueRefusesAndPreserves`, `TestGlmKeyAfterDashDashPassthrough`. RED reasons: `jev` command does not exist (compile/behavior RED); `moai glm --key x` currently falls through to launch parsing, never a save. The dash-leading cases (`TestGlmKeyExecFlagAsValueRefused` for `--key -f`; `TestJevKeyOptionTokenValueRefused` for `--key --help`) are RED against a value-takes-next-token mutant, which stores `-f` or `--help`; `TestGlmKeyEqualsFormSaves` (its `--key=-f` assertion) is RED against a mutant that refuses dash-leading values in every spelling. `TestJevKeyOptionTokenValueRefused` covers only `--key --help` today; its jev cases are extended to `--key -f`, `--key -`, `--key " -f"`, and `--key=-f` (AC-GJK-011, REQ-GJK-013). Against the value-takes-next-token mutant the bare `-` and padded `" -f"` cases store `-` and `-f`, so they are RED too; the glm bare and padded cases are RED in M4. The passthrough test is RED by MUTANT, not by the empty tree — against the unimplemented tree it is vacuously green (no scan exists to misbehave), so its RED cell is: with the scan stubbed to whole-args (the mutant), the test MUST fail (the credential file is created and the launch seam records no argument list equal to `["--", "--key", "test-key-1234567890"]`, with `MOAI_TEST_GLM_KEY` seeded per §D.6).
- GREEN: implement `internal/cli/jev.go` (§D.2) and the `runGLM` scan (§D.1).
- Files: `internal/cli/jev.go` (new), `internal/cli/glm.go` (scan + help flag entry + Long lines), `internal/cli/jev_key_test.go` (new).

### M2 (Priority High) — composition and preservation

- Characterization / composition tests: `TestGlmSetupRoutingUnchanged` (`moai glm setup K2` still stores via setup after the scan exists — REQ-GJK-006), `TestKeyFormsShareStorageLastWriterWins` (`--key A` then `setup B` → B; reverse → A — card criterion: both forms write the SAME storage and SAME value format), existing internal/cli glm test family stays green untouched.
- Files: tests only; zero production change expected in this milestone.

### M4 (Priority High) — glm separated-value check to the trimmed rule (run re-open; resolves §B.5)

- RED (`internal/cli/jev_key_test.go`): (a) a new table-driven test `TestGlmKeyBareAndPaddedDashValueRefused` (to be created) over `moai glm --key -` and `moai glm --key " -f"`: each returns a usage error, `.env.glm` (seeded with a stored key) stays byte-for-byte unchanged, and the stored value appears in neither stdout nor stderr; (b) the equality assertion in `TestGlmKeyAfterDashDashPassthrough`: seed `MOAI_TEST_GLM_KEY` and the §D.6 redirects, then assert the recorded list equals `["--", "--key", K]`. The current test asserts containment only and logs without failing when the launch path is unreached (lines 398-405), so it cannot see a discarded passthrough. RED reasons on the current tree: `--key -` stores `-` (the `next != "-"` exemption, glm.go:559-570), and `--key " -f"` stores the trimmed `-f` (the raw prefix test does not see the leading space).
- GREEN: `scanGLMKeyValue` (`internal/cli/glm.go:554`) trims the separated token before the dash test and drops the `"-"` exemption. The `=` spelling branch is unchanged (REQ-GJK-004).
- Re-run after GREEN: (i) the discard mutant (post-`--` tokens dropped without a launch) must make `TestGlmKeyAfterDashDashPassthrough` fail; (ii) M3's gates and the AC-GJK-013 measurement on the final tree, with evidence pinned to the final commit SHA (see the M3 ordering bullet). The observations are recorded in progress.md §E.2 (manager-develop's write, not this plan's).
- Files: `internal/cli/glm.go`, `internal/cli/jev_key_test.go`.

### M3 (Priority Medium) — verification and surface checks

- Ordering: M3 runs after M4 (§F order M1, M2, M4, M3). Its gates (the package run, the focused selector, the glmcred and jevcred packages, `go vet`, gofmt) and the AC-GJK-013 measurement run on the final tree, the tree after M4 lands. Evidence is pinned to the commit SHA of that tree, read with `git rev-parse --short HEAD` immediately before the run (verification-completeness §4). Evidence taken before M4 lands does not count toward closure.
- Verification families (lane-local discipline; CI owns the full suite). Preservation front line (criterion c — the card modifies `glm.go`): `go test ./internal/cli/` green; the focused selector is SECONDARY and honestly scoped — `go test ./internal/cli/ -run 'Test(GLM|Glm|Jev|Key|Root)'` sweeps 145 name-matching functions (`TestKey` covers `TestKeyFormsShareStorageLastWriterWins`; `TestGlm`/`TestJev`/`TestRoot` cover the new cases) but NOT the whole pre-existing glm family (name-shape misses, measured: glm_test 11, glm_new_test 35, glm_compat 4, glm_team 6, glm_persist_gate 5, mcp_glm_parse 2); also `go test ./internal/glmcred/ ./internal/jevcred/`, `go vet ./internal/cli/`, gofmt on touched files. A `[no tests to run]` line is an empty sweep — treat it as failure, never as a pass (verification-completeness §1.1).
- Constant-ownership grep, scoped to this card's ADDED NON-TEST lines — the tree already carries 3 pre-existing matching rows in untouched files (`glm_tools.go:6` comment, `mcp_audit.go:30,32`), so a whole-tree 0-row verdict is permanently red; and `_test.go` is a sanctioned literal area (AGENTS.local.md hardcoding allowance), where the new AC-GJK-004 and AC-GJK-005 style tests legitimately assert `GLM_API_KEY=` dotenv content: `git diff origin/main...HEAD -- internal/cli/ ':(exclude)**/*_test.go' | grep '^+' | grep "TYPESAFE_API_KEY\|GLM_API_KEY"` → 0 rows. Executable premise, run as separate commands BEFORE the pipeline (a bare pipeline cannot distinguish a zero-match exit from a git-failure exit): `git merge-base origin/main HEAD; echo "merge_base_exit=$?"` must print the base followed by `merge_base_exit=0` AND `git diff --stat origin/main...HEAD -- internal/cli/` must be non-empty. Meaningful only once the implementation commits exist — a diff with no added lines sweeps nothing and asserts nothing. Merge-base form per gitflow-lane-protocol §8 (never a literal pinned base); the range uses the three-dot form `origin/main...HEAD`, the same diff as `merge-base..HEAD`, because a `$(git merge-base …)` range is refused by the worktree-session guard (`worktree-integration-ops.md` authoring rule 3). The card base is main (AGENTS.md §3); the local gitflow-lane-protocol §1 and §8 still name develop, and that conflict is an operator decision left open.
  - Pathspec observation record (verification-completeness §2 — a verdict formula is complete only once its discrimination was observed): the delegation's control SHA `48a96bcb1` does not resolve in this tree (`fatal: ambiguous argument '48a96bcb1': unknown revision`), so the two-arm control ran on the local test-only commit `eaa97e321` (touches only `internal/cli/managed_codex_tui_test.go`): `git diff eaa97e321^..eaa97e321 -- internal/cli/ ':(exclude)**/*_test.go' | wc -c` → **0**; the same range without the exclude → **7826**. The exclude arm is what keeps the matcher from flagging the legitimate test assertions as violations.
- Guidance-line accuracy: re-read `mcp_jev.go:87` and `doctor_jev.go:82,88` — content still true (no edit expected).
- MX: `internal/cli/jev.go` is unexported-symbol-only; one `@MX:NOTE` on the `jevcred` delegation (storage SSOT lives in `internal/jevcred`) is sufficient.
- Evidence lands in progress.md §E.2 (manager-develop's write, not this plan's).

## §G. Anti-Patterns

- Do NOT create `.env.jev`, a second writer, or a reveal route — storage and disclosure stay delegated to `internal/glmcred` / `internal/jevcred`.
- Do NOT intercept the `setup`/`status`/`tools` routing: the `--key` scan sits AFTER the manual subcommand switch, never before it.
- Do NOT add a `-k` shorthand on glm (retired vocabulary) or echo the full key anywhere.
- Do NOT touch `init_jev_wizard.go`, `workflow.jev.enabled`, the `jev_ask` tool, the web console, or any t1612 renderer surface.
- Do NOT register `--key` as a cobra-persistent flag on `glmCmd` expecting parsing — `DisableFlagParsing: true` makes the manual scan the only working mechanism.
- Do NOT spell `TYPESAFE_API_KEY` / `GLM_API_KEY` as literals in `internal/cli`.

## §H. Cross-References

- SPEC-GLM-KEY-INPUT-001 (completed) — glmcred single-writer SSOT, D-1/C-2.
- SPEC-JEV-CORE-001 (completed) — `internal/jevcred`, REQ-JEVC-018 (storage), REQ-JEVC-020 (bounded disclosure).
- SPEC-JEV-OPTIN-MEASURE-001 — web console Jev credential view model (consumes the same storage).
- Card t1612 — downstream terminal-output redesign (JEV guidance); out of scope here.
- decision-index.md — decision routing for this SPEC (decision_gate: on).
