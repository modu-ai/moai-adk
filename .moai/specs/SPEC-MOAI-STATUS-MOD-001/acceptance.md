# Acceptance — SPEC-MOAI-STATUS-MOD-001

**Measurement tree (document-level pin, binds every criterion that carries no pin of its own):** `58dad30551368349cbc8a81889135f818abdade9`, branch `WT-moai-status-mod` (revision 0.2.0 — the iter1 delta repair touched plan-phase files only; the RED-now cells were run in the authoring session at the pinned tree, and the delta repair's two new measurements M-12/M-13 were taken at tree `3f51a713b`, recorded in spec.md §1). Judging builds: `claude 2.1.287`, `moai v3.2.0-rc.25`, `bun 1.4.2` (tool provenance, `verification-claim-integrity.md` §2.2). Every RED-now cell is a command run in this session; its streams are labeled and its exit code is its own field. Paths in output are the worktree's absolute paths.

## §A Dispositions that apply to every criterion

1. **Two runners, labeled.** A test-bearing criterion carries a **pure part** — judged under `bun test`, covering the classifiers, the parsers, the builders and the composer — and/or an **engine part** — judged under `claude plugin test`, covering anything that needs hook dispatch, `$.state`, the render hooks, `mount`, or stubbed `process.run` / `ui.status` / `ui.toast` / `session.measure` / `session.receive`. One criterion holds both parts in one body (one heading, one matrix row), each with its own test-name prefix. **Pure evidence is lane/developer evidence only**: `bun` is a user-local binary and no CI or tooling file of this repository references it (spec.md G-11, sibling M-15); it proves the pure functions, never hook dispatch or paint. The engine runner is the authority for the engine parts.
2. **How the engine runner is invoked.** Under an **empty temporary config directory**: `CLAUDE_CONFIG_DIR=/tmp/msm-claude-cfg-empty claude plugin test mods/moai-status` (the directory created empty with `mkdir -p`). Observed to execute tests on a scratchpad stub in this session (M-3, M-6) and accepted by the worktree guard in this env-prefixed form. The operator profile may refuse the runner via its rollout switch (sibling M-13) — a refusal is recorded as such and leaves the part UNOBSERVED; it is never a pass (spec.md G-10). Record which profile ran.
3. **Judging recipes.** *Pure:* `rm -f /tmp/msm-junit.xml`, then `bun test mods/moai-status/tests/pure/ --reporter=junit --reporter-outfile=/tmp/msm-junit.xml` (must exit 0), then for each prefix `grep -c '<testcase name="<prefix>' /tmp/msm-junit.xml` against the expected count, `grep -c '<failure' /tmp/msm-junit.xml` → `0` **and** `grep -c '<skipped' /tmp/msm-junit.xml` → `0` (bun counts a `test.skip` or `test.todo` as a present `<testcase>` with `<skipped`; the sibling's scratch control measured exactly that, their acceptance.md §A.3 — the count alone would accept a skipped test). In a non-TTY run bun prints no `(pass)` lines, so the junit file is the record. *Engine:* `mkdir -p /tmp/msm-claude-cfg-empty`, then `CLAUDE_CONFIG_DIR=/tmp/msm-claude-cfg-empty claude plugin test mods/moai-status > /tmp/msm-engine.out 2>&1` (must exit 0), then `grep -c -E '^\(pass\) <prefix>' /tmp/msm-engine.out` against the expected count and `grep -c -E '^\(fail\)' /tmp/msm-engine.out` → `0`. The observed engine line format is `(pass) <test name> [<ms>ms]` on stdout, also when redirected; a failing test prints `(fail) <name>` plus the assertion text, the summary ends ` N pass` / ` N fail` / `Ran N tests across N files`, and the exit code is 1 on any fail (M-3). The engine kit declares no skip API in its typings. A count of 0 is not a pass: an empty sweep asserts nothing (`verification-completeness.md` §1.1); bun exits 1 and the engine runner exits 1 when no test file exists.
4. **Closure.** An engine part closes on one executing run, at a recorded tree SHA, in which every named test passes. A refused runner is recorded as such and leaves the part UNOBSERVED; the SPEC stays open until an executing run exists, and neither the lane nor this SPEC waives that. Pure parts close under bun but never substitute for engine parts.
5. **Two-cell adoption.** Each criterion carries a RED-now cell and a green path naming the milestone that flips it (`verification-completeness.md` §2). Criteria that cannot be red by construction are labeled **regression-guard** and carry a positive control.
6. **Manual criteria are Gap-class.** AC-MSM-013 needs an interactive terminal, a loaded mod, and (for the live toast) a second session sending a delivery; it is not recorded as a pass until a person performs it, and it is not release-blocking (`verification-completeness.md` §2.1, undecidable disposition).

## §D AC matrix

| AC | Requirements | Runner | RED-now (tree `58dad3055`) | GREEN (milestone) |
|---|---|---|---|---|
| AC-MSM-001 | REQ-MSM-001, REQ-MSM-002 | validate | `claude plugin validate mods/moai-status` → exit 1 "File not found" | exit 0; `hooks:` names the five events' six registrations; the calls allow-list subset holds (M1-M4) |
| AC-MSM-002 | REQ-MSM-001, REQ-MSM-002 | structural grep | each grep → exit 2 (no such directory/file) | (i) 1 line, (ii) 3 lines, (iii) exit 1, (iv) 2 lines (M1-M4) |
| AC-MSM-003 | REQ-MSM-003 | pure | shared pure cell | 6 tests (M2) |
| AC-MSM-004 | REQ-MSM-004 | pure + engine | shared cells | pure 2 tests (M2); engine 3 tests (M2) |
| AC-MSM-005 | REQ-MSM-005 | pure + engine | shared cells | pure 2 tests (M2); engine 2 tests (M2) |
| AC-MSM-006 | REQ-MSM-003 | engine | shared engine cell | 2 tests (M2) |
| AC-MSM-007 | REQ-MSM-006 | pure + engine | shared cells | pure 2 tests (M3); engine 3 tests (M3) |
| AC-MSM-008 | REQ-MSM-008, REQ-MSM-012 | pure | shared pure cell | 8 tests (M4) |
| AC-MSM-009 | REQ-MSM-007, REQ-MSM-008, REQ-MSM-009 | engine | shared engine cell | 7 tests (M4) |
| AC-MSM-010 | REQ-MSM-009 | engine | shared engine cell | 2 tests (M4) |
| AC-MSM-011 | REQ-MSM-010 | validate | AC-MSM-001 cell | exit 0 and the `state writes:` line names `moai-status.*` keys (M1) |
| AC-MSM-012 | REQ-MSM-011 | git / grep | `git check-ignore -v mods/moai-status/.claude-plugin/types/claude-code/index.d.ts` → exit 1, stdout empty | exit 0 naming the new rule (M5); three regression guards stay as measured |
| AC-MSM-013 | REQ-MSM-004, REQ-MSM-005, REQ-MSM-006, REQ-MSM-008 | manual | not performed | manual launch observed (operator, interactive terminal; toast arm needs a second session) |

**The shared pure RED-now cell** (every `pure` part): command `bun test mods/moai-status/tests/pure/ --reporter=junit --reporter-outfile=/tmp/msm-junit.xml` after `rm -f /tmp/msm-junit.xml`; exit **1**; stdout `bun test v1.4.2 (744846f84)`; stderr:
```
The following filters did not match any test files in --cwd="/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1437":
 mods/moai-status/tests/pure/
6603 files were searched [141.00ms]

note: Tests need ".test", "_test_", ".spec" or "_spec_" in the filename (ex: "MyApp.test.ts")
```
(the `[Nms]` field and the file count vary run to run) and the judging count `grep -c '<testcase name="classify:' /tmp/msm-junit.xml` → stderr `ugrep: warning: /tmp/msm-junit.xml: No such file or directory`, exit **2**. The red's reason is the one the work removes: no pure test file exists.

**The shared engine RED-now cell** (every `engine` part): command `CLAUDE_CONFIG_DIR=/tmp/msm-claude-cfg-empty claude plugin test mods/moai-status`; exit **1**; stdout empty (0 bytes); stderr `claude plugin test: /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1437/mods/moai-status: no such plugin folder`. With the output redirected as in the recipe, the judging count `grep -c -E '^\(pass\) measure:' /tmp/msm-engine.out` → `0`, exit 1. The reason is the one the work removes: no mod folder exists. (The operator profile prints the rollout-switch refusal instead — not a RED, sibling M-13.)

## §D.1 AC detail

### AC-MSM-001 — the mod validates with exactly the additive surface
- **Given** `mods/moai-status/` with a manifest (a `types` pointer, because `$.state` is used), a `hooks/hooks.json` naming one module, and the hooks module,
- **When** `claude plugin validate mods/moai-status` runs,
- **Then** it exits 0, and its `./register.ts hooks:` line contains `session.start`, `session.end`, `session.measure`, `session.receive`, `ui.render{component=AbovePrompt}` and `ui.render{component=Spinner}` — five events, six registrations — and no other event name; and its `calls:` line passes the **allow-list subset check**: every `$.x.y` token extracted from the `calls:` line is one of REQ-MSM-002's eight nouns (`$.state.get`, `$.state.set`, `$.ui.toast`, `$.ui.status`, `$.ui.resolve`, `$.clock.every`, `$.clock.after`, `$.process.run`) — mechanically: `grep -o -E '\$\.[A-Za-z]+\.[A-Za-z]+' <validate output> | sort -u` prints only tokens from that set — **and** the control `grep -c -F '$.process.run'` on the same output → `1`.
- Judging: the token extraction and membership check over `claude plugin validate mods/moai-status` output (redirect to a file first), plus a read of the `hooks:` line. The listing control (M-12) is the standing proof that the extraction surface catches noun spellings.
- **RED-now** — command `claude plugin validate mods/moai-status`; exit **1**; stdout:
  ```
  Validating plugin manifest: /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1437/mods/moai-status

  ✘ Found 1 error:

    ❯ file: File not found: /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1437/mods/moai-status

  ✘ Validation failed
  ```
- Mutant: a `tool.call` registration → the `hooks:` line names it (fail); a `$.http.fetch` or `$.ui.ask` call anywhere the listing reaches → the extracted token set leaves the allowed eight (fail); removing all `$.process.run` calls → the control count `0`. **Listing control (M-12, this session):** `claude plugin validate /tmp/msm-ask-stub` on a scratchpad stub whose handler calls `$.ui.ask` exits 0 printing `❯ ./register.ts calls: $.ui.ask` — the listing catches forbidden noun spellings, so the subset check over the same line catches one the mod actually calls. **Limit, stated honestly:** the `calls:` line is a static listing, not an existence check (M-2, fake-noun control `$.ui.helloWorld` printed and exit 0), and it walks only the scope its analysis reaches (M-11) — the check therefore proves the listed surface is a subset of the eight nouns, not that no unreachable spelling exists; the engine tests (AC-MSM-002, AC-MSM-006, AC-MSM-007, AC-MSM-009) are the behavioral check.

### AC-MSM-002 — structural call-site checks
- **Given** the finished source under `mods/moai-status/hooks`, where every `$.…` call lives in `hooks/register.ts` and the helper modules `data.ts`, `health.ts` receive functions (`run`) or the resolved element table as arguments and never `$` (plan §B.1),
- **When** these run: (i) `grep -rn 'process\.run(' mods/moai-status --exclude-dir=.claude-plugin --exclude-dir=node_modules --exclude='*.test.*' --exclude='*.spec.*' --exclude='*.md'`; (ii) `grep -rn -E '"(doctor|memory)"' mods/moai-status/hooks/data.ts`; (iii) `grep -n -E '\$\.' mods/moai-status/hooks/data.ts mods/moai-status/hooks/health.ts`; (iv) `grep -c 'return next(e)' mods/moai-status/hooks/register.ts`,
- **Then** (i) prints exactly one line, in `hooks/register.ts`; (ii) prints exactly 3 lines (the three argv-table rows of plan §B.3); (iii) prints nothing and exits 1 (a `$.` in a helper module is a defect); (iv) prints at least 2 — a **weak structural guard**: the count survives a defective handler because every other pass-through (render branches, session start/end) also returns `next(e)`, so (iv) alone kills no consumed-shaped mutant; the authority for the no-consumed property is the engine observations of AC-MSM-006 and AC-MSM-007.
- **RED-now** — (i) exit **2**, stderr `ugrep: warning: mods/moai-status: No such file or directory`; (ii) exit **2**, same shape; (iii) exit **2** (two missing files — green is exit **1**, so exit 2 and exit 1 are told apart); (iv) exit **2**.
- Mutants: a second `process.run(` anywhere in `hooks/` → (i) two lines; an argv assembled from a template string with `doctor` spelled differently → (ii) fewer than 3; a `$.http.fetch` inside `health.ts` → (iii) matches; a measure handler that returns a `{ consumed }`-shaped value instead of `next(e)` → **(iv) still shows ≥ 2** (the other handlers keep counting) — this mutant is stated here precisely so nobody cites (iv) as killing it; AC-MSM-007's `toast: consumed is never produced` engine test is what fails.
- **Residual, stated honestly:** these pin occurrences of names and the places `$` may appear; they do not prove control flow. The behavioral checks for the pass-through guarantee are AC-MSM-006 and AC-MSM-007's engine tests, which observe the returned value of a dispatched hook.

### AC-MSM-003 — measure classification mirrors the gates (pure)
- **Given** the classifier of plan §B.4 — **`classifyMeasure(input, band)`**, a pure function of the pushed figure and explicit bounds (`band.autoCompactPct`, the soft/hard constants, the per-kind holds as plain fields) — and fixtures shaped like `SessionMeasureInput` (M-5): a 200K window with `percent` 89 and 91; a 1M window with `percent` 49 and 51; `percent` absent; a `five_hour` entry at `percentUsed` 89.9 and 90; a `seven_day` entry at 94.9 and 95; an unknown-kind entry at 99; a `hardCeiling` case where soft > naive hard (a 200K window — soft 90 — with an aggressive `band.autoCompactPct` of 60, making naive hard 70),
- **When** the pure classifier runs,
- **Then** six tests pass: `classify: context soft follows the window band` (200K → soft 90, so 89 is none and 91 is warn; 1M → soft 50, so 49 is none and 51 is warn), `classify: context critical at the hard ceiling` (defaults: min(95, 85+10) = 95, clamped up to soft when soft > it; the clamp branch itself is exercised by the overridden fixture — naive hard 70 < soft 90 clamps to 90 — reachable only because `autoCompactPct` is an input, not a frozen constant), `classify: absent percent is no reading, never zero` (a measure without `context.percent` classifies nothing and stores no 0), `classify: rate windows warn at their gate holds` (five_hour 90, seven_day 95; 89.9/94.9 are none), `classify: unknown kinds never warn` (a `spend_limit` entry at 99 stays info-only), `classify: unchanged classification is not rewritten` (the caller writes state only when the level moved).
- Judging: pure recipe (§A.3), prefix `classify:` → `6`, `<failure` → `0`, `<skipped` → `0`.
- **RED-now** — the shared pure cell. Mutant: a classifier using 85 for every window → the band test fails; one treating a missing `percent` as 0 → the absent test fails; one warning on `spend_limit` → the unknown-kind test fails; one ignoring `band.autoCompactPct` in favor of the frozen 85 → the clamp fixture classifies 95, not the clamped 90, and the hard-ceiling test fails.

### AC-MSM-004 — the strip draws, yields, and composes
- **Pure part.** **Given** the strip-line builder, **when** called with classifications, **then** `strip-pure: warn line names the figure and the threshold` (`ctx 91% (warn at 90)`; `5h quota 92%`), `strip-pure: nothing above info yields an empty line` pass. Judging: pure recipe, `strip-pure:` → `2`.
- **Engine part.** **Given** the mod's hooks, a `session.measure` stub-raised fixture, and the AbovePrompt component mounted via `$.ui.mount({ component: 'AbovePrompt', … })` with `ui.render` chains stubbed per test, **when** the band is drawn, **then** three tests pass: `strip: warn draws a one-line band` (the mounted tree carries a `Text` whose text contains `moai-status:`, **and still carries the upstream sentinel — the fixture's upstream `Text` resolved through `next(e)` draws beside the strip: REQ-MSM-004's compose requirement, later mods' drawings survive**), `strip: hasSurvey passes through` (the fixture upstream tree is returned unchanged — no `moai-status` text), `strip: quiet classification passes through` (same, with a below-threshold measure). Judging: engine recipe, `strip:` → `3`.
- **RED-now** — the shared pure cell and the shared engine cell. Mutant: a render hook that drops the upstream tree when drawing → the warn test fails (the upstream sentinel disappears) and the survey and quiet pass-through tests fail; one that draws on info → the quiet test fails.

### AC-MSM-005 — the spinner suffix rewrites only the suffix
- **Pure part.** **Given** the marker builder, **when** called with classifications, **then** `suffix-pure: warn context yields ctx marker` (` · ctx 91%`), `suffix-pure: nothing to say yields no marker` pass. Judging: pure recipe, `suffix-pure:` → `2`.
- **Engine part.** **Given** the Spinner component mounted with the engine's default `suffix` (`…`) and a measure fixture, **when** the line is drawn, **then** two tests pass: `suffix: warn appends to the incoming suffix` (the rendered props carry `suffix` ending ` · ctx 91%` and still starting with the incoming ellipsis; `word` and `message` unchanged), `suffix: quiet passes the event through` (`next(e)` resolved to the engine's own drawing). Judging: engine recipe, `suffix:` → `2`.
- **RED-now** — the shared cells. Mutant: a hook that overwrites `word` → the first engine test fails; one that replaces the ellipsis instead of appending → the same test's startswith assertion fails.

### AC-MSM-006 — the measure handler observes and never blocks
- **Given** the engine test kit raising `$.session.measure({...})` at the mod's hook with the `ui.status`-free, `process`-free environment stubbed per test,
- **When** the handler runs for a warn fixture and for a below-threshold fixture,
- **Then** two tests pass: `measure: warn classification reaches state` (`$.state` read after the dispatch carries the level), `measure: next(e) is always returned` (the dispatch's result echoes `{ changed }` and the delivery path — the engine's own observable — completes; the handler rejects nothing).
- Judging: engine recipe, `measure:` → `2`.
- **RED-now** — the shared engine cell. Mutant: a handler answering without `next` → the second test fails (the chain's result changes).

### AC-MSM-007 — the lane toast is a pure observer
- **Pure part.** **Given** delivery fixtures (a peer message with newlines and control characters, a `task-notification` with a long first line), **when** the line builder runs, **then** `toast-pure: one line from origin and excerpt` (no `\n`, starts with the origin kind, excerpt within 80 code points), `toast-pure: control characters stripped` pass. Judging: pure recipe, `toast-pure:` → `2`.
- **Engine part.** **Given** `$.session.receive({...})` raised at the mod's hook with `ui.toast` stubbed to record, **when** the delivery passes, **then** three tests pass: `toast: delivery is toasted before passing` (the stub recorded one line; the dispatch resolved to the queued delivery `{ text }`), `toast: toast failure still passes the delivery` (the stubbed toast throws; the dispatch still resolves `{ text }`), `toast: consumed is never produced` (no dispatch resolved a `{ consumed }` value on any path). Judging: engine recipe, `toast:` → `3`.
- **RED-now** — the shared cells. Mutant: a handler answering without `next` when the toast throws → the second engine test fails.

### AC-MSM-008 — health parsing against the pinned spellings (pure)
- **Given** the exact box-row outputs of M-9 (fresh, behind, newer-than-tree, different-branch; no-server, match, dev-build, stale-server) and a memory-doctor JSON shaped like M-8,
- **When** the parsers and composer run,
- **Then** eight tests pass: `health: behind row names both shas` (`binary is behind source tree (binary: 802a72235, HEAD: 58dad3055)` → warn segment carrying both), `health: fresh and different-branch rows are healthy`, `health: newer-than-tree row is warn undetermined`, `health: stale mcp row names pid and commits`, `health: no-server and match rows are healthy`, `health: summary line is not a verdict` (`0 ok, 1 warn, 0 fail` never classifies; a missing box row is unknown; a row whose STATUS token is not `ok`/`warn` — `fail` included, unreachable from the asked checks per REQ-MSM-012 — classifies as unknown, never healthy), `health: memory json worst over-cap store` (1422/50 → warn segment `memory 1422/50 files`; no over-cap store → healthy), `health: composer joins warns and clears when empty` (two warns joined with ` · `; empty → the clear signal). The row is parsed first: a non-zero doctor exit classifies as unknown only when no parseable row for the asked check is present, and an unparseable row is unknown (REQ-MSM-012, REQ-MSM-009).
- Judging: pure recipe, `health:` → `8`.
- **RED-now** — the shared pure cell. Mutant: a parser that greps any line containing `warn` → the summary-line test fails; one that treats a missing row as healthy → the same test fails.

### AC-MSM-009 — the health cycle runs, pins, clears, cancels, respects the floor and the timeout
- **Given** the engine test kit with `on('process.run', …)` answering the three fixtures by argv and **recording each call's `init`** (so `timeoutMs` is observable), a mocked clock (`mock.clock`) **that records the interval the timer registered with**, and `ui.status` stubbed to record,
- **When** the session starts, intervals elapse, and the session ends,
- **then** seven tests pass: `health-cycle: only table argv across ticks` (every recorded argv is one of the three; none else), `health-cycle: warn cycle pins one line` (the stubbed status received the composed line), `health-cycle: healthy cycle clears` (the stub received `undefined`), `health-cycle: a tick while one runs is dropped` (two intervals, one cycle — the process stub recorded one round), `health-cycle: timer cancelled on session.end` (advancing the clock after `session.end` records nothing), `health-cycle: interval respects the floor` (the recorded `$.clock.every` interval is at least `HEALTH_POLL_MIN_MS` — 15,000 ms, REQ-MSM-007), `health-cycle: run carries the 20s timeout` (every recorded `process.run` init carries `timeoutMs: 20000`, REQ-MSM-007).
- Judging: engine recipe, `health-cycle:` → `7` (all engine tests of this AC carry the `health-cycle:` prefix, distinct from the pure `health:` of AC-MSM-008).
- **RED-now** — the shared engine cell. Mutant: a timer surviving `session.end` → the cancel test fails; a second concurrent cycle → the drop test fails; a timer registered below the floor (e.g. 5,000 ms) → the interval test fails; a `runDiag` that drops `timeoutMs` (falling back to the engine's 30 s default) → the timeout test fails.

### AC-MSM-010 — fail-soft
- **Given** stubbed processes that reject (cannot start), return exit 2, and return unparseable output, and a hook body made to throw,
- **when** a cycle runs and a hook throws,
- **then** `failsoft: a rejected source shows unknown and keeps last good` (the pinned line carries `?` for that source and the other sources' real verdicts; `$.state` still holds the previous good classification) and `failsoft: a throwing hook does not reach the session` (the dispatch completes; the engine reports the skip as one line naming mod, event and reason — the observed M-3 shape) pass.
- Judging: engine recipe, `failsoft:` → `2`.
- **RED-now** — the shared engine cell. Mutant: a classifier that throws on a rejected run → the second test's dispatch fails visibly.

### AC-MSM-011 — the state contract
- **Given** `plugin.json` with `"types": "./types/index.d.ts"` and a contract declaring `interface PluginState` for `'moai-status'` with PascalCase-led exports (sibling M-17), and handlers that write the three state keys of plan §B.1 (`usage`, `health`, `notice`) through `{ plugin, key }` references,
- **When** `claude plugin validate mods/moai-status` runs,
- **Then** it exits 0 and its output contains a `./register.ts state writes:` line naming `moai-status.usage`, `moai-status.health` and `moai-status.notice` — the same three keys plan §B.1 declares. **The line was measured this session (M-13, tree `3f51a713b`):** on a scratchpad stub, validate printed `❯ ./register.ts state writes: f8stub.health, f8stub.notice, f8stub.usage` — key tokens spelled `plugin.key`, **sorted alphabetically, not in registration order**, so the judgment is set membership, never order.
- **RED-now** — the AC-MSM-001 cell (exit 1, file not found). Mutant: dropping the `types` pointer → validate exits 1 (`… is not declared: the manifest's types contract must name it in interface PluginState`, sibling M-17); a lowercase-led contract export → validate exits 1; a handler writing a key plan §B.1 does not declare → the writes line names a fourth token and the set-equality judgment fails.

### AC-MSM-012 — packaging and the no-Go no-template no-CI regression
- **Given** the finished mod and the `.gitignore` edit,
- **When** these run: `git check-ignore -v mods/moai-status/.claude-plugin/types/claude-code/index.d.ts` (a), `git check-ignore -v mods/moai-status/hooks/register.ts mods/moai-status/types/index.d.ts` (b), `git ls-files -- 'internal/template/templates/*moai-status*'` (c), `git diff --stat <base>..HEAD -- internal/ .github/ internal/template/templates/ .goreleaser.yml` (d, `<base>` = the card base expression `git merge-base develop HEAD` re-read at measurement time — the audit's card base was `58dad30551368349cbc8a81889135f818abdade9`),
- **Then** (a) exits 0 and names the new rule; (b) exits 1 (the rule does not swallow the authored source — a different path from the engine-written `.claude-plugin/types/`); (c) prints nothing; (d) prints nothing — no Go file, workflow file, template file, or release config changed on the branch.
- **RED-now (a)** — run at `58dad3055`: exit **1**, stdout empty.
- (b), (c), (d) are **regression-guards**, green today and incapable of red until someone mispaces the mod: (b) exit 1 (the directories do not exist yet — the same exit-2-vs-exit-1 caveat as AC-MSM-002 (iii) applies until M1, after which it is exit 1 on existing files); (c) exit 0, stdout empty, with control `git ls-files -- 'internal/template/templates/.claude/rules/moai/core/zone-registry.md'` printing that path, exit 0; (d) exit 0 with empty output at the plan-phase tree (nothing changed yet — it stays empty through the run phase, which is the guard). Mutant: a file added under `internal/template/templates/` named with `moai-status` makes (c) print its path; a rule `mods/*/types/` makes (b) exit 0; any `internal/` or `.github/` commit makes (d) print a stat line.

### AC-MSM-013 — manual: first interactive launch and a live toast (Gap-class, not release-blocking)
- **Given** an operator in an interactive terminal, the mod loaded with `claude --plugin-dir <absolute path to mods/moai-status>`, and a second session able to send a message,
- **When** they work past the soft/hard context thresholds, run with a behind binary, and receive one cross-session message,
- **Then** the band shows the strip at the thresholds and nothing below them, the spinner carries the marker while a warning holds, the toast shows one line per delivery, the status line names the behind binary with both SHAs, and the session behaves otherwise exactly as without the mod. The result is recorded in `progress.md` §E.2 with the date and the `claude` build; until then it is **unobserved**. If the operator profile's rollout switch is off the mod will not load there (G-10); the temp config dir does not stand in for that.

## §E Edge cases (covered by the tests above or listed as Gaps)

`percent` absent (no reading, never zero — REQ-MSM-003); `rateLimits` empty (off a subscription or before the first reading — no warn, no crash); a measure burst folding into one event; a delivery whose text is empty or entirely control characters (a bare origin-kind line); a toast while the transcript is in scrollback (the notification-bar shape — platform behavior, not asserted); `moai` missing from PATH (unknown source, REQ-MSM-009); a doctor row whose check name does not match the asked check (unknown, never another check's verdict); two warns at once (joined line); a `window` of 0 or absent (the band rule needs it — treated as no reading); the `vscode` and `mobile` surfaces (G-1 — AbovePrompt and Spinner are raised on terminal and desktop only, M-5).

## §F Definition of Done

1. AC-MSM-001..012 evidence in `progress.md` §E.2: command, verbatim output, exit code, tree SHA, `claude`, `moai` and (for pure parts) `bun` builds, each labeled **pure (bun)**, **engine**, or **UNOBSERVED**.
2. **Every engine part (in AC-MSM-004, -005, -006, -007, -009 and -010) is observed on an executing engine run under the temp config dir (§A.2), or the SPEC is not done.** A refusal does not close it; pure parts closed under bun do not substitute for it; no waiver is granted here. The operator profile's own refusal of the runner is reported by the leader to the operator and is not a SPEC blocker.
3. Every path printed by `git diff --name-only <base>..HEAD` (`<base>` = `git merge-base develop HEAD` re-read at measurement time) lies under `mods/moai-status/`, `.gitignore`, or this SPEC's directory — the committed diff is the check, not the working tree's cleanliness (a clean tree alone would satisfy any working-tree check vacuously).
4. AC-MSM-013 listed as Gap-class in the completion report, not as a pass.
5. Commit messages carry `t1437`; the verdict lives at `.moai/reports/t1437/verdict.md` (written by the lane).
