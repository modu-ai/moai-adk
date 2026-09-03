# t216 / D-2 — the 11 present-but-unwired hook scripts

> Investigation record, card t216, axis D-2. Read-only; HEAD `a9eb896ce`,
> worktree `.claude/worktrees/t216`, branch `WT-hook-wiring-drift`.
> Persisted by the orchestrator — the investigating agent is read-only and
> holds no Write tool.

## Claim

Of the 11 scripts on disk that `.claude/settings.json` does not name, **only 4
are genuinely dead**, and all 4 are dead *by a recorded decision*, not by
neglect. **2 are actively reachable** by non-settings mechanisms the
settings-only view cannot see. **2 are deliberately deregistered** with a
documented root cause. **3 are ambiguous** and need a call.

Two audit findings need correcting:

1. **The "34 entries" figure is wrong.** There are **33 hook entries**, not 34.
   `grep -c '"type": "command"'` returns 34 because the **`statusLine` block is
   also `"type": "command"`** and is not a hook. The duplicate-matcher story
   still holds (32 distinct scripts across 33 entries), but the arithmetic that
   produced 34 counted a non-hook.
2. **`handle-agent-hook.sh` is not unwired.** It is registered **10 times via
   agent frontmatter `hooks:` blocks** (5 local, 5 template-mirrored) — a real
   Claude Code invocation surface. The audit classified its agent-`.md` hits as
   "docs"; they are YAML hook registrations inside frontmatter, not prose.

And one audit judgment the dispatch asked to re-verify independently splits in
two: `handle-session-start-compact.sh` is **correctly** by-design-unwired;
`handle-session-start-navigator.sh` is **not** — it was wired, then removed as
apparent collateral in a build-recovery commit.

## Evidence

### Baseline reproduction (this worktree, HEAD `a9eb896ce`)

```
$ find .claude/hooks/moai -maxdepth 1 -name '*.sh' -exec basename {} \; | sort > present.txt   # 43
$ grep -o 'hooks/moai/[A-Za-z0-9._-]*\.sh' .claude/settings.json | sed 's|hooks/moai/||' | sort -u > wired.txt   # 32
$ comm -13 present.txt wired.txt     # wired-but-missing
(empty)
$ comm -23 present.txt wired.txt     # present-but-unwired
chain-event.sh
handle-agent-hook.sh
handle-elicitation-result.sh
handle-elicitation.sh
handle-notification.sh
handle-session-start-compact.sh
handle-session-start-navigator.sh
handle-task-created.sh
handle-worktree-create.sh
handle-worktree-remove.sh
team-ac-verify.sh
```

43 / 32 / 0 / 11 reproduce exactly.

### Entry-count correction

```
$ grep -c '"type": "command"' .claude/settings.json
34
$ python3 -c "...walk d['hooks']..."   → total hook entries 33
$ python3 -c "import json;print(json.load(open('.claude/settings.json')).get('statusLine'))"
{'type': 'command', 'command': '$CLAUDE_PROJECT_DIR/.moai/status_line.sh', 'refreshInterval': 10}
```

**33 hook entries, 32 distinct scripts, 1 duplicate.** The 34th `"type":
"command"` is `statusLine`.

### The duplicate matcher — intentional

```
.claude/settings.json  PreToolUse[0]  matcher="Write|Edit|Bash"  → handle-pre-tool.sh  timeout=10
.claude/settings.json  PreToolUse[1]  matcher="Agent|Task"       → handle-pre-tool.sh  timeout=10
```

Identical in the shipped template (`settings.json.tmpl:47` matcher `:52`; `:58`
matcher `:63`). The two matchers do genuinely different work inside the shared
handler:

- `internal/hook/pre_tool.go:431/514/532/544` — the `Bash`/`Write`/`Edit` paths
  (dangerous-pattern deny, branch guard, integration lock, file-access deny,
  Write content scan).
- `internal/hook/pre_tool.go:597` — `if input.ToolName == "Agent" || input.ToolName == "Task"`
  → `checkAgentModel`, the per-agent model guard.

Provenance: `git log -S'"matcher": "Agent|Task"' -- …settings.json.tmpl` →
`b0d3b61f8 feat: moai web console redesign + per-agent model enforcement (#1376) (#1410)`
— added deliberately, for that feature.

**Verdict: intentional, and harmless.** The two tool sets are disjoint, so no
double-fire is possible. Merging into one `Write|Edit|Bash|Agent|Task` matcher
would be behaviourally identical; the split is a readability/ownership choice.
One stale artifact worth noting: `handle-pre-tool.sh:39` still comments "matcher
scope is Write|Edit|Bash" — written before the second matcher existed.

### Per-script table

| script | local settings.json | template settings.json | referenced by (file:line, exhaustive) | verdict |
|---|---|---|---|---|
| `chain-event.sh` | no | **yes** — `settings.json.tmpl:198`, SubagentStop, unconditional (outside the `{{ if .HookOptIn.Enabled }}` block ending `:195`) | `internal/chain/node.go:27` (doc comment naming it as the SubagentStop producer of `EventCompletionEdge`) | **reachable-via-template-settings** — live for every distributed project; dead only in this checkout |
| `handle-agent-hook.sh` | no | no | **agent frontmatter `hooks:` (real registrations, 10):** `.claude/agents/moai/manager-develop.md:21` (Stop), `sync-auditor.md:21` (Stop), `manager-spec.md:22` (Stop), `manager-docs.md:22` (PostToolUse) + `:27` (Stop), and the 5 template mirrors. Also `.moai/manifest.json:173`, `internal/cli/update.go:1079`, `.claude/rules/moai/core/agent-hooks.md:7,28,34,39`, `hooks-system.md:338` | **reachable-via-agent-frontmatter** — NOT unwired. Settings-only counting cannot see this surface |
| `handle-elicitation.sh` | no | no | `internal/template/retired_wrappers_test.go:18`, `m002_settings_cleanup_test.go:55`, `hook-independence.md:102` | **dead — by decision** |
| `handle-elicitation-result.sh` | no | no | `retired_wrappers_test.go:19`, `m002_settings_cleanup_test.go:56`, `hook-independence.md:102` | **dead — by decision** |
| `handle-notification.sh` | no | no | `retired_wrappers_test.go:17`, `m002_settings_cleanup_test.go:54,86,303,558`, `hook-independence.md:101` | **dead — by decision** |
| `handle-task-created.sh` | no | no | `retired_wrappers_test.go:20`, `m002_settings_cleanup_test.go:57,87`, `hook-independence.md:102` | **dead — by decision** |
| `handle-session-start-compact.sh` | no | no | `internal/hook/session_start_compact.go:12-16` (explicit design note); production path `internal/cli/deps.go:237` `HookRegistry.Register(hook.NewSessionStartCompactHandler())`; manual path `internal/cli/hook.go:151-164,557-574`. `.claude/settings.local.json:151` is a `Bash(cp …)` permission string, **not** wiring | **reachable-via-in-binary-registry** — production firing rides `handle-session-start.sh` → `moai hook session-start`; the wrapper is the manual/isolated surface. Deliberately unwired to avoid double-fire |
| `handle-session-start-navigator.sh` | no (never — pickaxe over `.claude/settings.json` is empty) | no (**was**, removed) | `internal/template/navigator_hook_test.go:3,19` (executes it directly under test), `.claude/skills/moai-workflow-project/references/navigator.md:157` + template mirror `:150` | **ambiguous — needs a decision** |
| `handle-worktree-create.sh` | no | no | `.claude/skills/moai-workflow-worktree/SKILL.md:309` + template `:310` (presence checklist, not invocation), `hook-independence.md:105` | **dead — deliberately deregistered, documented root cause** |
| `handle-worktree-remove.sh` | no | no | `SKILL.md:309`/`:310`, `hook-independence.md:105` | **dead — same decision** |
| `team-ac-verify.sh` | no (never — pickaxe empty across all settings paths and all history) | no (never) | `internal/template/hook_official_compliance_test.go:62,73`, `internal/hook/user_decision_capture.go:16` (comment), `internal/codexadapter/output.go:26` (comment), `agent-common-protocol.md:38,98` + mirror, `agent-common-protocol-reference.md:208,214,231,291` + mirror, `hook-independence.md:33,69,71,79,89` + mirror, `.claude/skills/moai/workflows/goal.md:152`, `.claude/agents/harness/hook-ci-specialist.md:69` | **ambiguous — needs a decision** |

### Search space covered for every `dead` verdict

An absence claim is only as good as its named scope. For each of the 11 names,
matched as a literal filename with
`grep -rn --binary-files=without-match -E '<name>'`:

- **Path roots:** the whole worktree, plus explicitly `Makefile`, `.github/` (all
  workflows/actions), and `scripts/` — the last three returned **zero hits** for
  all 11.
- **Include globs:** `*.sh`, `*.go`, `*.json`, `*.tmpl`, `*.yml`, `*.yaml`,
  `*.py`, `*.mjs`, `*.js`, `Makefile`, plus a second pass over `*.md` restricted
  to `.claude/` and `internal/template/templates/`.
- **Excluded:** `.git/`, `.moai/reports/` (prior audit prose), `node_modules/`.
- **Settings surfaces enumerated:** `.claude/settings.json`,
  `.claude/settings.local.json`, `internal/template/templates/.claude/settings.json.tmpl`,
  `~/.claude/settings.json`, `~/.claude/settings.local.json`,
  `~/.moai/claude-profiles/moai-adk/settings.json` — the last three returned **no
  hits for any of the 11**.

**Sibling-hook sourcing ruled out:**

```
$ grep -rln 'team-ac-verify|chain-event|worktree-create|worktree-remove|agent-hook|session-start-navigator|session-start-compact' .claude/hooks/moai/
chain-event.sh  handle-agent-hook.sh  handle-session-start-compact.sh
handle-worktree-create.sh  handle-session-start-navigator.sh
team-ac-verify.sh  handle-worktree-remove.sh
```

Every hit is the file naming **itself** in its own header. **No wired script
sources or execs an unwired sibling.**

**Go-invokes-shell ruled out:**

```
$ grep -rn 'hooks/moai' internal --include='*.go' | grep -v _test.go
internal/statusline/context_usage.go:92   (comment)
internal/spec/audit_transition.go:5       (comment)
internal/cli/update.go:942,950            (directory removal, not invocation)
internal/migration/migrations/m001_hardcoded_path.go:36  (file glob for rewriting)
internal/hook/session_start_compact.go:12 (comment)
internal/defs/dirs.go:414                 (path constant)
```

**No Go code execs a `.sh` wrapper.** The binary is only ever the callee.

### Template twins and `.sh` / `.sh.tmpl` pair agreement

| script | template twin | form | `diff` local vs template |
|---|---|---|---|
| `chain-event.sh` | yes | `.sh` | **DIFFERS** |
| `handle-agent-hook.sh` | yes | **`.sh` AND `.sh.tmpl`** | identical (and `.sh` == `.sh.tmpl`) |
| `handle-elicitation.sh` | yes | `.sh.tmpl` | identical |
| `handle-elicitation-result.sh` | yes | `.sh.tmpl` | identical |
| `handle-notification.sh` | yes | `.sh.tmpl` | identical |
| `handle-session-start-compact.sh` | yes | `.sh` | identical |
| `handle-session-start-navigator.sh` | yes | `.sh` | identical |
| `handle-task-created.sh` | yes | `.sh.tmpl` | identical |
| `handle-worktree-create.sh` | yes | `.sh.tmpl` | identical |
| `handle-worktree-remove.sh` | yes | `.sh.tmpl` | identical |
| `team-ac-verify.sh` | yes | `.sh` | identical |

**All 11 have template twins — none is local-only.** Removing any of them is a
distributed act, not a local cleanup.

The one drift:

```
$ diff .claude/hooks/moai/chain-event.sh internal/template/templates/.claude/hooks/moai/chain-event.sh
2c2
< # SPEC-CHAIN-CORE-001 REQ-CHAIN-012 — chain-event hook wrapper.
---
> # chain-event hook wrapper — appends completion edge to the chain ledger.
```

Comment-only; no behavioural divergence (and the template side is correctly
SPEC-ID-free per the neutrality rule). The `.tmpl` files for the six
`.sh.tmpl`-form scripts contain no Go template tokens — byte-identical to the
rendered `.sh`. The four scripts carrying a genuine `.sh` + `.sh.tmpl` pair in
the template dir are `handle-agent-hook`, `handle-stop-goal`,
`handle-task-completed`, `handle-teammate-idle`; only `handle-agent-hook` is
among the 11, and its pair **agrees byte-for-byte**.

### Was it ever wired?

```
$ git log --oneline -S'<name>' -- '*settings.json' '*settings.json.tmpl'
```

| script | history | reading |
|---|---|---|
| `chain-event.sh` | `435bc2bbd feat(SPEC-CHAIN-CORE-001): Origin-Trail Chain — Phase 1 (#1485)`; `--name-only` shows it touched **only** the tmpl | **Never wired locally.** Template-only; the local settings was never regenerated |
| `handle-elicitation.sh`, `handle-elicitation-result.sh`, `handle-notification.sh`, `handle-task-created.sh` | added `0fb0c2829` / `17d420b1a` / `78f673761` / `7aeac6f43`; **removed together** in `a165706e6 feat(SPEC-V3R2-RT-006): Hook Handler 27-Event Coverage (#984)`, `-44` lines from `.claude/settings.json`, `-60` from the tmpl. Body: *"Pattern A gate applied to 4 RETIRE-OBS-ONLY handlers … All 4 audit tests GREEN."* | **Deliberately retired**, enforced by `retired_wrappers_test.go` + `m002_settings_cleanup_test.go` |
| `handle-worktree-create.sh`, `handle-worktree-remove.sh` | added `239b217fa`, removed `2606ed05b`, re-added `daa720a4c`/`7aeac6f43`, **removed again** in `a3239d3de fix(hook): WorktreeCreate/Remove 등록 해제 + 공식 컨트랙트 문서 정정` | **Deliberately deregistered with a recorded regression as the cause.** CC v2.1.49+ `WorktreeCreate` is an *active creator* — the hook's stdout must be the worktree path. MoAI's observer-style handler returned `{}`, which CC read as a path → *"WorktreeCreate hook returned a path that is not a directory: {}"*, breaking `isolation: worktree` for 5 agents. Non-registration is now documented policy |
| `handle-session-start-compact.sh` | **empty** | Never wired anywhere, by design (`session_start_compact.go:12-16`) |
| `team-ac-verify.sh` | **empty** | Never wired anywhere, ever, in any settings surface |
| `handle-session-start-navigator.sh` | added to the **template only** by `2c87d195f feat(SPEC-PROJECT-NAVIGATOR-001) (#1354)` at SessionStart/timeout 5; **removed** by `7171880a9 feat: complete accumulated internal code — build recovery (#1409)` | see below |

### Why `handle-session-start-navigator.sh` is ambiguous, not "unreachable by design"

`load-path.md` inherits "unwired = unreachable = fine". The history says
otherwise. Commit `7171880a9` removed **two** hook entries from the template in
the same diff:

```
$ git show 7171880a9 -- internal/template/templates/.claude/settings.json.tmpl | grep '^[-+@]'
@@ -10,12 +10,6 @@
-            "args": [".../handle-session-start-navigator.sh"],  "timeout": 5,
@@ -133,12 +127,6 @@
-            "args": [".../handle-codex-review-gate.sh"],  "timeout": 900,
@@ -501,9 +489,8 @@   (permission-list reversions)
@@ -553,16 +540,7 @@   (permission-list reversions)
```

`handle-codex-review-gate.sh` was **restored** two commits later by
`e8e46396f fix(hooks): make the two review gates reachable … (#1421)`. The
navigator entry was not. The commit body of `7171880a9` describes Go-side
recovery work and never mentions removing a hook; the permission-list hunks in
the same diff are plain reversions to an older file state. **This has the
signature of a template regression that was partially repaired, not a decision.**

Corroborating: `internal/template/navigator_hook_test.go` still executes the
script directly and asserts AC-PN-009/010/012 — the acceptance criteria are alive
and green while the delivery mechanism is gone. Separately,
`.moai/project/navigator/navigator.md` does not exist in this tree
(`ls .moai/project/navigator/` → only `symbols/`), so even a restored wiring
would fail-open at the script's first `[ ! -r "$NAV_FILE" ]` guard.

**What would settle it:** an owner statement on SPEC-PROJECT-NAVIGATOR-001 —
either (a) the ambient auto-brief is still wanted, in which case restore the
template SessionStart entry alongside a `navigator.md` generation path, or (b) it
was superseded, in which case retire `navigator_hook_test.go` and the
`references/navigator.md:157` claim so the docs stop describing a hook that
cannot fire.

### Why `team-ac-verify.sh` is ambiguous

The doctrine calls it "dormant", but that word is doing two jobs:

- The **script's** meaning (`team-ac-verify.sh:5-7`): *"this hook exits 0
  immediately unless workflow.yaml declares `team.enabled: true`"* — it
  self-gates at runtime. Its header states its trigger as *"TaskCompleted event
  when team.enabled: true in workflow.yaml"*.
- The **actual state**: it has **never** appeared in any settings.json in the
  entire history. `TaskCompleted` in `.claude/settings.json` entry 25 is wired to
  `handle-task-completed.sh`, which does not call it
  (`grep 'hooks/moai\|\.sh"' handle-task-completed.sh` → one self-naming comment
  at `:14`). No Go code registers it; the only Go hits are comments.

So flipping `team.enabled: true` would **not** activate it. The self-gate is
unreachable because the invocation is.
`agent-common-protocol-reference.md:291` describes it as *"TaskCompleted in team
mode (dormant — harness `thorough` + team prerequisites)"*, which reads as
"registered but gated" and is not what the wiring says. Meanwhile
`hook_official_compliance_test.go` enforces its reject-shape contract, keeping a
contract green for a path nothing can enter.

**What would settle it:** decide whether team mode is meant to fire it. If yes,
it needs a `TaskCompleted` entry in `settings.json.tmpl` (the script already
self-gates, so an unconditional entry is safe). If no, the reference doc's
"dormant" wording should be corrected to "not registered — future opt-in",
matching the language already used for the worktree hooks.

## Baseline-attribution

Inherited from `rest-and-wiring.md` and **re-derived here rather than taken on
trust**: the 43 / 32 / 0 counts, the present-but-unwired membership list, and the
`chain-event.sh` template-vs-local drift direction.

**Corrected:** the "34 entries" figure (→ 33 hook entries + 1 `statusLine`); the
classification of `handle-agent-hook.sh` (→ reachable via agent frontmatter).

**Independently verified rather than inherited from `load-path.md`:** the
`handle-session-start-compact.sh` judgment (**confirmed**, with the in-binary
registration path supplied as the reason) and the
`handle-session-start-navigator.sh` judgment (**not confirmed** — reclassified
from "unreachable by design" to ambiguous-regression).

**New in this pass:** the full unwiring provenance for all 11, the template-twin
and pair-agreement matrix, and the `pre_tool.go:597` justification for the
duplicate matcher.

## Gaps

1. **The Claude Code runtime was not executed**, so **agent-frontmatter `hooks:`
   support is asserted from `agent-hooks.md` and the frontmatter YAML, not
   observed firing.** If that runtime feature were silently unsupported,
   `handle-agent-hook.sh` would fall back to dead. `.moai/logs/` carries no
   agent-hook trace usable to confirm.
2. The `chain-event.sh` comparison used the local template `.tmpl`. `origin/main`
   was not read (the prior audit did, reporting 1 hit); this worktree's template
   agrees.
3. The prior audit's plugin-`hooks.json` sweep (6 files under `~/.claude/plugins`
   and the profile cache) is inherited, not re-run — those live outside the
   worktree.
4. `.claude/settings.json` was not diffed against a fresh render of the template,
   so the full local-vs-template wiring drift is only partly characterized here.
   A second drift beyond `chain-event.sh` was noticed and belongs to the D-1
   axis: the template wires **`status-transition-ownership.sh` three times** under
   `PostToolUse` with `if` predicates (`settings.json.tmpl:78,86,94`) while local
   wires it **once**.

## Residual-risk

- **Removing any of the 4 dead scripts is a distributed act.** All four have
  template twins (`.sh.tmpl` form) and are pinned by `retired_wrappers_test.go` +
  `m002_settings_cleanup_test.go`; the m002 migration exists to strip their
  entries from *user* settings files on upgrade. Deleting the wrappers without
  retiring those tests and re-checking the migration's expectations would break
  the build and could leave upgraded user projects pointing at absent files —
  which the `hook-missing.log` fallback would absorb silently.
- **The two worktree hooks should stay on disk.** `worktree-integration.md`
  documents a future active-creator opt-in path (`git worktree add` + stdout
  redirect). Deleting them discards that opt-in surface and the recorded
  regression context.
- **`chain-event.sh` is the one live gap in this checkout.**
  `internal/chain/node.go:27` documents `EventCompletionEdge` as produced by that
  hook on SubagentStop; distributed projects get it, this repo does not. Any
  chain-lineage feature developed or tested here runs against a ledger missing
  completion edges — a silent dev/production divergence. (D-1 establishes the
  further fact that no ledger exists at all, for an independent reason.)
- **Two green test suites guard paths nothing can enter**
  (`navigator_hook_test.go`, `hook_official_compliance_test.go` AC-HOC-001). They
  keep passing regardless of the wiring decision, so neither ambiguity surfaces on
  its own through CI.
