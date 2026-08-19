# t140 / GD-3 — `general-purpose` availability: investigation result

Investigation card. No code change proposed here; the fix-necessity verdict is the lead's.

Binary under test: `claude --version` → `2.1.235 (Claude Code)`
(path `/Users/goos/.local/share/claude/versions/2.1.235`, the same version the changelog item belongs to)

## 1. Which session classes lack `general-purpose` — OBSERVED

Read directly out of the shipped 2.1.235 bundle. The built-in agent list is selected by
one three-way function, and the list is built from it:

```js
function Ooi(){
  if (K.CLAUDE_AGENT_SDK_DISABLE_BUILTIN_AGENTS && Rn()) return "none";
  if (_v()) return "coordinator";
  return "default";
}
function eht(){
  let e = Ooi();
  if (e === "none") return [];                                    // no built-ins at all
  if (e === "coordinator") { ...getCoordinatorAgents() }          // see below
  let t = [eAe]; ...                                              // eAe = general-purpose
}
function Rn(){ return !host.launchOptions.isInteractive() }        // non-interactive
function _v(){                                                     // coordinator mode
  if (!Nn(process.env.CLAUDE_CODE_COORDINATOR_MODE)) return false;
  if (LH() && !sc() && !K.CLAUDE_CODE_REMOTE) return false;        // interactive & local & not remote -> off
  return true;
}
function ztS(){ return [ Swp ] }                                   // getCoordinatorAgents
// Swp = { agentType: bUr, whenToUse: "For executing tasks autonomously — research,
//         implementation, or verification.", tools:["*"], maxTurns:200,
//         permissionMode:"bubble", source:"built-in" }
// bUr = "worker"
// eAe = { agentType: "general-purpose", whenToUse: "General-purpose agent for
//         researching complex questions, ..." }
```

Therefore exactly **two** classes lack `general-purpose`:

| Class | Condition (all parts required) | Built-ins present |
|---|---|---|
| `none` | env `CLAUDE_AGENT_SDK_DISABLE_BUILTIN_AGENTS` truthy **AND** the session is non-interactive | none at all (`[]`) |
| `coordinator` | env `CLAUDE_CODE_COORDINATOR_MODE` truthy **AND** (non-interactive **OR** remote workspace **OR** `CLAUDE_CODE_REMOTE`) | exactly one: `worker` |

`coordinator` is the class the changelog does not mention and the official docs page does
not list. It is not merely "general-purpose removed" — the whole built-in set is replaced
by a single `worker` agent.

`default` (everything else) begins its list with `eAe`, i.e. `general-purpose`.

The fallback the changelog describes is the same `eAe`:
`let gr = t ?? eAe.agentType` — an omitted `subagent_type` resolves to general-purpose,
and where that is absent the throw site is
``throw new Vyt(`${AVo}. Available agents: ${FTi(Rr)}`)`` with
`AVo = "subagent_type is required: the general-purpose agent is not available in this session"`.

### Correction to a mid-investigation reading
`code.claude.com/docs/en/errors`, read via WebFetch, reported that the error does **not**
enumerate available agents. The binary contradicts that: the throw site appends
`. Available agents: <list>`. The fetched page was reported truncated at that section, so
the changelog wording ("listing the available agents") is the accurate one and the doc read
was wrong. Recorded because the wrong reading was acted on for one step.

## 2. Reproducible under this project's configuration? — NOT REPRODUCED, and the reason is measured

Neither trigger exists anywhere in this repository or in the live session:

```
$ grep -rn "DISABLE_BUILTIN_AGENTS" <repo> --include=*.{go,json,tmpl,sh,md,yaml}   -> 0 hits
$ grep -rn "COORDINATOR_MODE"       <repo> --include=*.{go,json,tmpl,sh,md,yaml}   -> 0 hits
$ [ -n "$CLAUDE_CODE_COORDINATOR_MODE" ]            -> UNSET
$ [ -n "$CLAUDE_AGENT_SDK_DISABLE_BUILTIN_AGENTS" ] -> UNSET
```

The other two documented removal paths are also absent:

- `permissions.deny` in `.claude/settings.json` (46 entries), `~/.claude/settings.json` (`[]`),
  `.claude/settings.local.json` (absent), and `settings.json.tmpl`: **no `Agent(...)` entry of
  any kind**. `disallowedTools`: absent from all four.
- No agent definition claims the name `general-purpose`
  (`grep -rln "^name: general-purpose" .claude/agents/ internal/template/templates/.claude/agents/` -> none),
  so no shadowing collision.
- No MoAI launcher passes `--print` / `-p` / `--disallowedTools` / `--agents`
  (`grep` over `internal/cli/*.go` non-test: the only `-p` is `--profile` at `launcher.go:816`).

Positive control: this very session (interactive, this repo's settings, 2.1.235) lists
`general-purpose` among its available agent types. So the `default` class is confirmed
present, not merely assumed.

## 3. Overlap with MoAI's session shapes — NO OVERLAP MEASURED

- MoAI's own Go code never launches a non-interactive `claude`. The headless shape MoAI
  *documents* (`dynamic-workflows.md:41`, a `claude -p` batch loop) satisfies `Rn()` but sets
  neither env var, so it resolves to `default` and keeps `general-purpose`.
- Workflow scripts never name `general-purpose` as `agentType`
  (`grep -rn "general-purpose" .claude/workflows/` -> 0); they name `Explore` or a project
  specialist, and otherwise take the runtime's default workflow subagent.
- The consumers that do depend on it — `archived-agent-rejection.md` §C rows 7-12,
  `manager-lead.md:42,62,135,175`, `super-advisor.md:6` — all run in the orchestrator's own
  interactive session, i.e. the `default` class.

## Gaps (explicitly not observed)

- **Who turns coordinator mode on is undetermined.** The setter `C5b(e)` writes
  `process.env.CLAUDE_CODE_COORDINATOR_MODE="1"` when passed `"coordinator"`, but no caller
  was located; the bundle is minified and the grep for call sites returned only the
  definition. The env name sits in a forwarded-env list beside `CLAUDE_BG_*` (background)
  entries, which *suggests* background/cloud sessions — that is INFERENCE, not observation.
- **No live negative reproduction.** No session was actually launched with either env var to
  watch the error fire. The claim above rests on reading the shipped bundle, not on running it.
- **Only three doc pages consulted**: `code.claude.com/docs/en/sub-agents`,
  `code.claude.com/docs/en/errors`. The env-vars page was not opened.
- **Managed (org/admin) settings not inspected** — a managed-policy `permissions.deny` outside
  this machine's files could deny `Agent(general-purpose)` and was not observable from here.

## Verdict input for the lead (not a decision)

Under measured conditions the two classes lacking `general-purpose` do not intersect any
session shape MoAI launches or documents. On item 3's own criterion that makes this a
one-line documentation item rather than a fallback-design item. The one thing that would
change that is the undetermined gap above: if background / cloud / Remote-Control sessions
are what set `CLAUDE_CODE_COORDINATOR_MODE`, then a MoAI session running there would see a
single `worker` built-in and every `Agent(subagent_type: "general-purpose")` spawn in
`archived-agent-rejection.md` §C, `manager-lead.md`, and `super-advisor.md` would fail.
That question is answerable and is the thing worth answering before sizing any card.
