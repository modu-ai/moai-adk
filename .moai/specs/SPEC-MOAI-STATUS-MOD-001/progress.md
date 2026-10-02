# Progress — SPEC-MOAI-STATUS-MOD-001

Status after plan phase: **draft**, plan-phase artifacts complete (spec.md, plan.md, acceptance.md, decision-index.md, this file). Owner of the next transition: manager-develop (run phase) after the plan audit and the Kickoff gate.

## §E.1 Plan-phase Audit-Ready Signal (this session, tree `58dad30551368349cbc8a81889135f818abdade9`, branch `WT-moai-status-mod`)

plan_status: audit-ready
plan_complete_at: 2026-10-02 (iter2 verdict PASS-WITH-DEBT 0.82 + the verdict-recommended 0.2.1 touch-up applied)

Each row: the command verbatim and its observed output. Builds in play: `claude 2.1.287`, `moai v3.2.0-rc.25`, `bun 1.4.2`.

### ID pattern and uniqueness

- Command: `ls .moai/specs | grep -c MOAI-STATUS` — output `0`, exit 1 (no match; the id was free before authoring; recorded as M-10).
- Command: `ls -d .moai/specs/SPEC-MOAI-STATUS-MOD-001` — output `.moai/specs/SPEC-MOAI-STATUS-MOD-001` (the directory now exists; id `SPEC-MOAI-STATUS-MOD-001` matches `^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$`). **PASS**

### Frontmatter field presence

- Command: `for f in id title version status created updated author priority phase module lifecycle tags tier related_specs; do printf "%s: " "$f"; grep -c "^${f}:" .moai/specs/SPEC-MOAI-STATUS-MOD-001/spec.md; done`
- Output: every field printed `1` (14/14 — the 12 canonical fields plus `tier` and `related_specs`). **PASS**

### REQ and AC structure

- Command: `grep -c "^- \*\*REQ-MSM-" .moai/specs/SPEC-MOAI-STATUS-MOD-001/spec.md` — output `12` (REQ-MSM-001..012, all with leading list markers; spec-lint's REQ collection needs them).
- Command: `grep -c "^### AC-MSM-" .moai/specs/SPEC-MOAI-STATUS-MOD-001/acceptance.md` — output `13` (AC-MSM-001..013).
- AC-to-REQ coverage: the matrix of acceptance.md §D names all 12 REQs (001/002 in AC-001/002; 003 in AC-003/006; 004 in AC-004/013; 005 in AC-005/013; 006 in AC-007/013; 007 in AC-009; 008 in AC-008/009/013; 009 in AC-009/010; 010 in AC-011; 011 in AC-012; 012 in AC-008). **PASS**

### Spec lint

- Command: `moai spec lint SPEC-MOAI-STATUS-MOD-001`
- Exit code: 0. Output:
  ```
  ✓ No findings — all SPEC documents are valid
  ```
- JSON form: `moai spec lint SPEC-MOAI-STATUS-MOD-001 --json` — exit 0, output `[]` (zero findings).
- Control (the argument really resolves): `moai spec lint SPEC-MOAI-STATUS-MOD-999` — exit **3**, stderr `spec lint: no SPEC document found for "SPEC-MOAI-STATUS-MOD-999" (tried: /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1437/.moai/specs/SPEC-MOAI-STATUS-MOD-999/spec.md)` — the resolver distinguishes a missing SPEC (argument error) from a linted one. **PASS**

### OutOfScopeRule heading shape

- spec.md §6 carries three `### Out of Scope — <topic>` h3 subsections (actions and turn flow / health remediation and deeper signals / deployment, Go, template, CI), each with `-` bullets. Confirmed by the lint run above (the rule is one of the lint's checks). **PASS**

### Artifact set (Tier M)

- Present: `spec.md`, `plan.md`, `acceptance.md` (the Tier M set) plus `decision-index.md` (decision gate on, Q1-Q7) and `progress.md`. `design.md` / `research.md` deliberately absent (Tier M does not name them). **PASS**

### RED-now cells (acceptance.md §D, run this session)

- `claude plugin validate mods/moai-status` — exit 1; stdout ended `❯ file: File not found: /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1437/mods/moai-status` / `✘ Validation failed`.
- `CLAUDE_CONFIG_DIR=/tmp/msm-claude-cfg-empty claude plugin test mods/moai-status` — exit 1; stderr `claude plugin test: /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1437/mods/moai-status: no such plugin folder`.
- `bun test mods/moai-status/tests/pure/ --reporter=junit --reporter-outfile=/tmp/msm-junit.xml` (after `rm -f /tmp/msm-junit.xml`) — exit 1; stderr carried the no-match filter (`mods/moai-status/tests/pure/`, 6603 files searched); judging `grep -c '<testcase name="classify:' /tmp/msm-junit.xml` → exit 2, no file.
- `git check-ignore -v mods/moai-status/.claude-plugin/types/claude-code/index.d.ts` — exit 1, stdout empty (the ignore rule is absent; plan §D.5 adds it at M5).
- `git status --short` — only `?? .moai/specs/SPEC-MOAI-STATUS-MOD-001/` (the plan-phase tree holds nothing else new).

### Delta repair 0.2.0 (plan-audit iter1 FAIL 0.75, must-fix only — this session, tree `3f51a713b`)

- Spec lint re-run after the repair:
  - Command: `moai spec lint SPEC-MOAI-STATUS-MOD-001` — exit **0**, output verbatim:
    ```
    ✓ No findings — all SPEC documents are valid
    ```
- Two new measurements were taken for the repair and recorded as M-rows in spec.md §1 (commands and verbatim output there):
  - **M-12** (`$.ui.ask` listing control, AC-MSM-001's allow-list): `CLAUDE_CONFIG_DIR=/tmp/msm-claude-cfg-verify claude plugin validate /tmp/msm-ask-stub` → exit 0, `❯ ./register.ts calls: $.ui.ask`.
  - **M-13** (the `state writes:` line, AC-MSM-011): `CLAUDE_CONFIG_DIR=/tmp/msm-claude-cfg-verify claude plugin validate /tmp/msm-f8-stub` → exit 0, `❯ ./register.ts state writes: f8stub.health, f8stub.notice, f8stub.usage` (sorted; names all three keys). AC-MSM-011's green condition is observable on this build — no Gap demotion needed.
- F-1 disposition: option (b) — the `fail` STATUS token dropped from REQ-MSM-012, reason recorded in the HISTORY row and the requirement itself (both asked checks assign only ok/warn in-tree; `internal/cli/doctor.go:629-679`, `internal/cli/doctor_mcp_version.go`). Precedence fixed row-first in REQ-MSM-009/012 and plan §B.4.
- No file outside this SPEC directory and /tmp was written; no commit made (the lane commits).

### Post-audit touch-up 0.2.1 and Kickoff decision

- Touch-up per the iter2 verdict's own recommendation, gate disposition decided via Jev consultation (`proceed_after_touchup`, confidence 0.89, gate 0.5): N-3 — `notice`'s writer pinned in plan §B.1 (the fail-soft guard writes it on a thrown hook or a persistently unknown source, clears on the next clean cycle); N-2 — env key corrected to `CLAUDE_AUTOCOMPACT_PCT_OVERRIDE` (envkeys.go:622; D-3, G-12). Version 0.2.0 → 0.2.1. Lint after the touch-up: `moai spec lint SPEC-MOAI-STATUS-MOD-001` → exit 0, `✓ No findings — all SPEC documents are valid`.
- Kickoff decision record: decided_by=lane+Jev evidence_refs=.moai/reports/t1437/plan-audit-iter2.md(PASS-WITH-DEBT 0.82 @ 3cf3d810c),.moai/reports/t1437/plan-audit-iter1.md(FAIL 0.75),.moai/reports/t1437/kickoff-decision.md ladder_path=gate-row:plan→run Kickoff AUTONOMOUS §9.1; audit-cache disposition — card-dir verdicts always miss the legacy `.moai/reports/plan-audit/` cache path, so iter2 is the audit of record and Phase 1 takes the documented skip (no re-audit); the 0.2.1 touch-up after the verdict is exactly the verdict's recommended edit, recorded here.
- Heading fix: the §E.1/§E.2 headings originally carried no `§` token, which era.go's literal matching would have read as "no modern-era markers" (H-2 → V3R2-R4 misclassification); aligned to the canonical `§E.n` map this session (sibling t1436's progress.md is the working precedent).

## §E.2 Run-phase Evidence

Builds in play at run phase (re-measured at entry): `claude 2.1.287`, `moai v3.2.0-rc.25`, `bun 1.4.2`. Tree at run start: `c7b72b430` (branch `WT-moai-status-mod`). Engine runner profile: `/tmp/msm-cfg-empty-fresh` (fresh empty dir created this session); the temp-config runner check exited 0 with usage text. The plan-session's laid typings (build 2.1.287, `/tmp/msm-stub/.claude-plugin/types/claude-code/index.d.ts`) are the §4 authority; no re-lay was needed (same build observed at C1).

### M1 — contracts and skeleton (commits of this milestone)

- **RED (pure, bun)** — `rm -f /tmp/msm-junit.xml; bun test mods/moai-status/tests/pure/ --reporter=junit --reporter-outfile=/tmp/msm-junit.xml` before `hooks/data.ts` existed: exit **1**, verbatim `error: Cannot find module '../../hooks/data' from '.../mods/moai-status/tests/pure/data.spec.ts'`, ` 0 pass / 1 fail`. Red for the stated reason: the module under test did not exist.
- **GREEN (pure, bun)** — same command after `data.ts` (M1 slice: argv table, run-signature types, §G constants): exit **0**, ` 1 pass / 0 fail`; junit `argv:` → 1, `<failure` → 0, `<skipped` → 0.
- **validate (engine surface, temp profile)** — `CLAUDE_CONFIG_DIR=/tmp/msm-cfg-empty-fresh claude plugin validate mods/moai-status`: exit **0**; verbatim lines:
  - `❯ types ./types/index.d.ts declares state: moai-status.usage, moai-status.health, moai-status.notice`
  - `❯ ./register.ts hooks: session.start, session.end, session.measure, session.receive, ui.render{component=AbovePrompt}, ui.render{component=Spinner}` (five events, six registrations)
  - `❯ ./register.ts calls: $.state.get, $.state.set` (`$.process.run` joins when M4 wires `runDiag` into a handler)
  - `❯ ./register.ts state writes: moai-status.health, moai-status.notice, moai-status.usage` (sorted; AC-MSM-011's set)
  - `❯ ./register.ts state reads: moai-status.health, moai-status.notice, moai-status.usage`
  - `✔ Validation passed`
  - One validate-caught defect fixed during M1: the first-write initialization loop passed the state ref through a loop variable; validate refused non-literal refs (`takes a reference whose plugin and key are string literals`). Unrolled to three literal-ref calls.
- Pre-flight C5 re-measure (live CLI): `moai doctor --check "Binary Freshness"` → one box row `warn    Binary Freshness  binary is behind source tree (binary: 802a72235, HEAD: c7b72b430)` exit 0; `--check "MCP Server Version"` → `ok      MCP Server Version  no running moai MCP server recorded` exit 0. `moai memory doctor --json` → exit 0, 186,726 bytes, array of stores (findings may be `null`). The spellings M-9 pinned are unchanged.

### M2 — strip and spinner (TDD; commits of this milestone)

- **RED (pure, bun)** — classifier/strip/suffix tests added to `tests/pure/data.spec.ts` before the implementation: exit **1**, verbatim `SyntaxError: Export named 'DEFAULT_BAND' not found in module '.../hooks/data.ts'` (the classifier and its band did not exist). Red for the stated reason: the functions under test were absent.
- **RED (engine, temp profile `/tmp/msm-cfg-empty-fresh`)** — `tests/engine.test.ts` (`measure:` ×2) + `tests/status.test.tsx` (`strip:` ×3, `suffix:` ×2) against the M1 pass-through skeleton: exit **1**, `3 fail / 4 pass` — `measure: warn classification reaches state` (Expected containing "ctx 91% (warn at 90)", Received: undefined), `strip: warn draws a one-line band` (Expected containing "moai-status:", Received: undefined), `suffix: warn appends to the incoming suffix` (Expected "… · ctx 91%", Received: "…"). The pass-through guards (`hasSurvey`, quiet, `next(e)` echoes) passed. RED for the stated reason: no classification is written and no suffix is appended yet.
- Three engine-harness facts learned during RED (they shape every later engine test): (1) a raised event needs a test-side answer — `on('session.measure', (_$, e) => ({ changed: e.changed }))` in `setup` (M-3's "stubs are per test"); (2) test-side `on()` hooks must register before the test's first `$` call; (3) for a `ui.render` mount the chain's last hook must answer with a drawing (`next(e)` at the end is "no implementation for ui.render") — the tests' sentinel/observer hooks are terminal drawings recorded through closures, and the composition proof is that a mod dropping the upstream never lets the terminal run.
- **GREEN (pure, bun)** — exit **0**, `11 pass / 0 fail`; junit counts: `classify:` → 6, `strip-pure:` → 2, `suffix-pure:` → 2, `argv:` → 1; `<failure` → 0; `<skipped` → 0.
- **GREEN (engine)** — exit **0**, `7 pass / 0 fail` (`measure:` ×2, `strip:` ×3, `suffix:` ×2).
- One GREEN-phase defect found and fixed: the strip Text carried the `key`, but find matches Box keys only (sibling craft note: "Text takes no key") — the findable line now sits in a keyed Box. Diagnosed by temporarily removing the hook's fail-soft catch (no error surfaced; the hook ran clean) and restored after the fix.
- validate after M2: `calls: $.state.get, $.state.set, $.ui.resolve` (grows with the render hooks); hooks/state lines unchanged from M1.

### M3 — lane toast (TDD; commits of this milestone)

- **RED (pure, bun)** — `toast-pure:` ×2 re-added to `tests/pure/data.spec.ts` before the builder existed: exit **1** (`Export named 'toastLine' not found` — the load-error shape; the builder was absent).
- **RED (engine, temp profile)** — `toast:` ×3 in `tests/engine.test.ts` against the M2 tree: exit **1**, `9 pass / 1 fail` — `toast: delivery is toasted before passing` (Expected `["peer: lane says hi"]`, Received `[]`). The two guards (`toast failure still passes`, `consumed is never produced`) passed against the pass-through handler; their shared RED-now is the pre-work no-mod cell of acceptance.md §D.
- **GREEN (pure, bun)** — exit **0**, `13 pass / 0 fail`; junit counts: `classify:` → 6, `strip-pure:` → 2, `suffix-pure:` → 2, `toast-pure:` → 2, `argv:` → 1; `<failure` → 0; `<skipped` → 0.
- **GREEN (engine)** — exit **0**, `10 pass / 0 fail`; `(pass) toast:` → 3.

### M4 — health cycle (TDD; commits of this milestone)

- **RED (pure, bun)** — `tests/pure/health.spec.ts` written before `hooks/health.ts` existed: exit **1**, `13 pass / 1 fail / 1 error` — the health spec failed to load (module not found), the red for the stated reason.
- **RED (engine, temp profile)** — `health-cycle:` ×7 + `failsoft:` ×2 in `tests/engine.test.ts`: exit **1**, `8 fail` — every cycle test failed with `calls.length 0` (no timer registered yet) and the unknown-source shape; the throwing-hook guard passed (it exercises M3 behavior plus the engine's skip line). One RED-phase harness fact: the engine refuses a second `on('session.receive', …)` in one test without a matcher — the throwing sibling hook registers matcher-narrowed (`{ origin: 'peer' }`) before setup's catch-all answer.
- **GREEN (pure, bun)** — exit **0**, `23 pass / 0 fail`; junit counts: `health:` → 8, `health-rowfirst:` → 1 (debt N-4: a parseable row wins over any exit code — the parser sees stdout alone), `health-merge:` → 1 (unknown keeps the previous good verdict); `<failure` → 0; `<skipped` → 0.
- **GREEN (engine)** — exit **0**, `19 pass / 0 fail` (`health-cycle:` ×7, `failsoft:` ×2 among them).
- Three GREEN-phase defects the engine caught, each fixed: (1) `runDiag`'s `$` was a free variable — it resolved to an environment-global `$` whose `process.run` bypassed the test stubs (every source read unknown while no argv was recorded); `$` threads from the handler now. (2) The engine's `$`-flow rule refuses `$` passed to a function declared inside `register` — `runDiag` moved to the top of the hooks module (the sibling's `runMoai` shape); validate's `calls:` line then completed to `$.clock.every (via startHealth), $.process.run (via runDiag), $.state.get, $.state.set, $.ui.resolve, $.ui.status (via runHealthCycle), $.ui.toast`. (3) The clock's cancel handle is `{ cancel }` at runtime (an object), not the bare function the laid typings name — the stop helper accepts both shapes defensively; `session.end` also needed a test-side answer (`{ sessionId }`).

### M5 — packaging and handoff (commits of this milestone)

- `.gitignore` gained exactly the one rule `mods/*/.claude-plugin/types/` (REQ-MSM-011); `git check-ignore -v mods/moai-status/.claude-plugin/types/claude-code/index.d.ts` now exits **0** naming `.gitignore:445:mods/*/.claude-plugin/types/` (AC-MSM-012a green; the RED-now exit-1/empty state was measured pre-work and re-confirmed at M1 pre-flight).
- `README.md` written (launch, test/validate commands, the additive-observer boundary, the `$.`-single-file layout).
- Two structural ACs repaired to their mechanical green at M5: AC-MSM-002(ii) — the argv table's three rows re-pinned in double quotes so `grep -rn -E '"(doctor|memory)"' hooks/data.ts` prints exactly 3 lines (the check names themselves are unchanged); AC-MSM-002(iii) — two doc comments in the helper modules mentioned `$.process.run` in backticks, tripping the "no `$.` in a helper module" grep; reworded (no code change — the single call site was always in `register.ts`).

### AC matrix at run completion (E1; tree of the final M5 commit, branch `WT-moai-status-mod`)

Builds: `claude 2.1.287`, `moai v3.2.0-rc.25`, `bun 1.4.2`. Engine profile: `/tmp/msm-cfg-empty-fresh`. Pure recipe per acceptance.md §A.3 (`rm -f /tmp/msm-junit.xml` → junit counts); engine recipe: `mkdir -p` done, `claude plugin test mods/moai-status` → exit 0, `(pass)` lines counted.

| AC | Status | Command (decisive) | Decisive output |
|---|---|---|---|
| AC-MSM-001 | **PASS** | `CLAUDE_CONFIG_DIR=/tmp/msm-cfg-empty-fresh claude plugin validate mods/moai-status` | exit 0; `hooks:` names the five events' six registrations; token extraction `grep -o -E '\$\.[A-Za-z]+\.[A-Za-z]+' \| sort -u` → `$.clock.every, $.process.run, $.state.get, $.state.set, $.ui.resolve, $.ui.status, $.ui.toast` (subset of the eight); control `grep -c -F '$.process.run'` → **1** |
| AC-MSM-002 | **PASS** | (i)–(iv) structural greps | (i) exactly 1 line: `hooks/register.ts:134:  $.process.run(argv, { timeoutMs: CMD_TIMEOUT_MS })`; (ii) exactly 3 lines (data.ts:32-34); (iii) no output, exit 1; (iv) `9` (≥ 2; weak guard — the consumed-shape authority is AC-MSM-007's engine test) |
| AC-MSM-003 | **PASS** (pure) | bun junit | `classify:` → **6**, `<failure` → 0, `<skipped` → 0 |
| AC-MSM-004 | **PASS** (pure+engine) | bun junit; engine | `strip-pure:` → **2**; `strip:` → **3** engine `(pass)` lines |
| AC-MSM-005 | **PASS** (pure+engine) | bun junit; engine | `suffix-pure:` → **2**; `suffix:` → **2** |
| AC-MSM-006 | **PASS** (engine) | engine | `measure:` → **2** |
| AC-MSM-007 | **PASS** (pure+engine) | bun junit; engine | `toast-pure:` → **2**; `toast:` → **3** |
| AC-MSM-008 | **PASS** (pure) | bun junit | `health:` → **8** (plus extras `health-rowfirst:` 1, `health-merge:` 1 — own prefixes, outside the AC count) |
| AC-MSM-009 | **PASS** (engine) | engine | `health-cycle:` → **7** |
| AC-MSM-010 | **PASS** (engine) | engine | `failsoft:` → **2** |
| AC-MSM-011 | **PASS** | validate | exit 0; `state writes: moai-status.health, moai-status.notice, moai-status.usage` — set membership over the three declared keys (sorted, per M-13) |
| AC-MSM-012 | **PASS** | (a)–(d) | (a) `git check-ignore -v …` exit 0 naming the rule; (b) exit 1 (authored sources not swallowed); (c) `git ls-files -- 'internal/template/templates/*moai-status*'` empty (control prints the zone-registry path); (d) `git diff --stat <base>..HEAD -- internal/ .github/ internal/template/templates/ .goreleaser.yml` empty (`<base>` = `git merge-base develop HEAD` = `58dad3055`, re-read at measurement) |
| AC-MSM-013 | **UNOBSERVED** (Gap-class, not release-blocking) | manual | needs an operator in an interactive terminal with the mod loaded and a second session able to send a delivery; not performed in this lane |

- Suite totals at completion: **pure (bun)** 23 pass / 0 fail / 0 skipped across `tests/pure/`; **engine** 19 pass / 0 fail across `tests/engine.test.ts` + `tests/status.test.tsx`. Engine parts of AC-MSM-004/-005/-006/-007/-009/-010 were each observed on executing engine runs under the temp config dir (acceptance.md §F DoD-2 met — no UNOBSERVED engine part).
- DoD-3 (committed-diff path check): `git diff --name-only 58dad3055..HEAD` filtered against `mods/moai-status/`, `.gitignore`, and this SPEC directory leaves **nothing** — every committed path is in scope.
- AC-MSM-013 checklist for the operator (recorded, not claimed): load with `claude --plugin-dir <abs path to mods/moai-status>`; work past the context thresholds and watch the strip/suffix; run with a behind binary and a second session sending one delivery; confirm the toast, the status line naming both SHAs, and otherwise-unchanged behavior. If the operator profile's rollout switch is off the mod will not load there (G-10).

## §E.3 Run-phase Audit-Ready Signal

run_status: audit-ready
run_complete_at: 2026-10-03
run_commits: 13bc02b29 (M1), c7dc6a860 (M2), 9d3df65f2 (M3), 319364815 (M4), and the M5 packaging commit of this section — all on `WT-moai-status-mod`, all carrying `t1437` and the `Authored-By-Agent: manager-develop` trailer
push_state: not pushed — leader batch (git-flow lane protocol; the lane never pushes)

### Measurement notes carried into the run phase (authored at plan phase; the `## §E.3 Run-phase Audit-Ready Signal` section is manager-develop's, written at run completion)

- The authority typings were laid by a **headless session load** (`claude -p` with `--plugin-dir` and an empty `CLAUDE_CONFIG_DIR`; spec.md M-4). `validate` and `plugin test` lay nothing. If C4 must re-lay them at run-phase entry, budget one headless model call.
- The validate `calls:` line is a static listing with scope limits (spec.md M-2, M-11): it accepts fake nouns and misses top-level `$` calls. AC-MSM-001's allow-list pair is written knowing this; the behavioral checks are the engine tests.
- The engine test environment stubs everything per test: `session.measure`, `session.receive`, `ui.status`, `ui.toast`, `process.run` — and the test `$` raises events at the mod's hooks (`$.session.measure(fixture)`; spec.md M-3).
- The doctor single-check message spellings are pinned from the in-tree sources (`internal/cli/doctor.go:629-679`, `internal/cli/doctor_mcp_version.go:43-94`) and re-measured from the CLI (M-9); C5 re-checks both at run-phase entry.
- This session is a lane: no queue-mutating command was run, no commit was made (the lane commits after the audit), and no file outside the worktree and /tmp was written.

## §E.4 Sync-phase Audit-Ready Signal

sync_commit_sha: "03b8a763e"
sync_phase: complete — the single sync commit carries the CHANGELOG `[Unreleased]`/`### Added` entry, the spec.md `in-progress → implemented → completed` terminal transition (`updated: 2026-10-03` — already current), this §E.4 signal, and the two MX tag additions. A commit cannot cite its own SHA; the placeholder is backfilled with the real SHA in the sanctioned follow-up commit by the lane.

- CHANGELOG: one entry added as the first bullet under `[Unreleased]` `### Added` (new mod). Duplicate-entry guard ran before the edit: `grep -c 'SPEC-MOAI-STATUS-MOD-001' CHANGELOG.md` → `0`; the entry names only paths verified present (`ls mods/moai-status/`; figures cited from §E.1/§E.2/§E.3 as read this session).
- MX tag changes (sync sub-step, none existed before — scan `grep -n '@MX' mods/moai-status/hooks/*.ts` → no match): `@MX:ANCHOR [AUTO]` + `@MX:REASON` + `@MX:SPEC` added on `hooks/register.ts` `runDiag` (the module's only `$.process.run` call site); `@MX:NOTE [AUTO]` + `@MX:SPEC` added on the `hooks/data.ts` `DEFAULT_BAND` mirrored-defaults block. No existing tag deleted, demoted, or rewritten.
- Sync verification run this session: duplicate guard 0 → entry added; every path named in the entry exists; spec.md frontmatter touched only on `status:` (allowed scope); `git status --short` re-read immediately before staging; staged by explicit pathspec only (CHANGELOG.md, spec.md, progress.md, hooks/register.ts, hooks/data.ts).
- AC-MSM-013 remains UNOBSERVED (manual, interactive) — the operator checklist recorded in §E.2 stands; sync claims nothing for it.
- push_state: not pushed — leader batch (git-flow lane protocol; the lane never pushes).

## §F Phase 4 Mode Selection

Input parameters: tier M; scope ≈ 12 files (11 under `mods/moai-status/` + root `.gitignore`); domains = 1 (TypeScript mod under `mods/`); file language mix = 100% TypeScript/JSON/Markdown; concurrency benefit = LOW (coding-heavy, single-writer tree).

| Mode | Selected | Rationale |
|---|---|---|
| direct | not selected | Multi-file new-code implementation, not a trivial edit |
| serial | **selected** | Coding-heavy work — Anthropic's coding-task parallelism caveat; one writer per tree; milestones are sequential by dependency |
| fanout | not selected | Not research-heavy; a write race in one tree would need isolation |
| sweep | not selected | Not a mechanical uniform transform; new code |

Decision: serial — one `manager-develop`-workflow implementation agent per milestone set, spawned by the lane session.

Justification: the whole deliverable is new TypeScript in one plugin directory with strict internal invariants (single `$.` file, fixed argv table, render-hook discipline), so sequential TDD milestones under one agent preserve the invariants better than any concurrent shape; the lane's one-writer-per-tree rule independently forbids concurrent writers anyway.
