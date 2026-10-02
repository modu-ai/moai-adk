# SPEC-AUDIT-MODEL-CONVERGE-001 — design

> Tier L design artifact (0.1.3). It carries the reasoning behind spec.md §B.2
> and the shapes the run phase builds to. Measurements cited here are in
> research.md §R.1, §R.5 and §R.6 (tree `739556db5`, code identical to
> `c50da9c2f`). This file has no frontmatter `status:` — the SPEC's lifecycle
> state lives in spec.md only.

## §D.1 The resolver table (REQ-ACV-001)

One row per `audit.model` token. A cell is the gate the token assigns to the
backend and, in brackets, the source of that cell. `config.model` cells are
**explicit**; `default` cells are not.

| `audit.model` | claude | codex | glm | Reading |
|---|---|---|---|---|
| (empty / key absent) | required [default] | required [default] | advisory [default] | The engine default. Not an operator choice: fail-open exactly as today. |
| `claude` | required [config.model] | required [default] | advisory [default] | Decision D7'. Claude's gate is the operator's and explicit; codex and glm are left to the default profile, so `audit_multi` behaves as it does today and `cross_model_active` stays false. An operator who wants Claude alone sets `audit.gates.codex: off` and `audit.gates.glm: off`. |
| `codex` | off [config.model] | required [config.model] | off [config.model] | Codex alone (the skill's existing "codex reviews alone"). |
| `glm` | off [config.model] | off [config.model] | required [config.model] | GLM alone. |
| `multi` | required [config.model] | required [config.model] | advisory [config.model] | The cross-model profile: codex required, GLM advisory, Claude the anchor. Same gates as the default; the difference is the source — an operator wrote it, so it is fail-closed. |

Why `multi` and the empty token share their gates: the distributed default
profile was always "claude + codex required, glm advisory"
(`internal/config/audit_models.go`, `NewDefaultWorkflowConfig`). What was missing
was a way to say "I mean it". The token is that statement.

Why `claude` is the asymmetric row (D7'): the code pairs the `claude` token with
the default gates (`defaults.go:1272-1304`), `audit_models.go` calls it both "the
default single-backend audit model" and part of the default profile, and whether
the settings wizard persists an untouched `claude` is unobserved. If `claude`
meant Claude alone and some project had persisted it, that project would silently
lose codex and GLM in `audit_multi` — a change in the direction of less review.
If it means "Claude explicit, rest default" and some operator wanted Claude
alone, the cost is one visible edit. The two errors are not symmetric, so the row
takes the safe reading. The `codex` and `glm` tokens are never a default value, so
their "alone" reading carries no such risk.

## §D.2 Precedence, sources, and what counts as explicit (REQ-ACV-002, B.2.3)

Per backend, highest first:

1. `argument` — the gate the caller put in the tool call's `gates` object for
   that backend.
2. `config.gates` — the tree's `audit.gates.<backend>`.
3. `config.model` — the cell of §D.1 for the tree's `audit.model` token.
4. `default` — the distributed default (claude required, codex required, glm
   advisory).

A gate is **explicit** when its source is `config.gates` or `config.model`.
Explicitness is what the three existing fail-closed readers key on today
(`enforceRequiredGateUnmet`, `applyGateUnmet`, the receipt predicate): they read
the RAW `audit.gates` block and ignore the engine default, because a default
nobody wrote must not flip every project to fail-closed. The resolver keeps that
rule and widens "wrote" from "wrote `audit.gates`" to "wrote `audit.gates` or an
`audit.model` token". Only a `required` explicit gate enforces anything; an
explicit `off` or `advisory` never blocks. A caller-supplied `required` (source
`argument`) is NOT explicit: today it is not enforced either, and changing that
would alter existing callers (plan.md §B OQ-3).

## §D.3 Where it lives, what feeds it, who calls it (REQ-ACV-004, REQ-ACV-005)

```
internal/config/audit_plan.go     the pure resolver  (no file, env or clock)
        ^                ^                 ^                    ^
internal/cli/            internal/cli/     internal/cli/        internal/auditreceipt/
mcp_audit_multi.go       mcp_worktree_     audit_plan_cmd.go    store.go
(fallback plan)          root.go           (moai verify         (CodexGateRequired)
                         resolveAuditGates  audit-plan)
                         -> enforceRequiredGateUnmet (convergence)
                         -> applyGateUnmet (codex_audit)
```

**Raw reads only (REQ-ACV-004).** The resolver is fed the `AuditConfig` value the
tree's `workflow.yaml` carries and nothing else. The readers that feed it are the
existing raw ones: `loadWorkflowAuditSection` / `workflowAuditPins`
(`internal/cli/audit_pin.go:34-64`), the receipt store's own struct
(`internal/auditreceipt/store.go:236-257`), and `resolveAuditGates`
(`internal/cli/mcp_worktree_root.go:101-113`). None constructs
`NewDefaultWorkflowConfig()`; that function sets `Audit.Model: claude` and the
default gates, so a default-merged read would hand `claude` to every project. A
static check (AC-ACV-004) and a pins-only fixture (AC-ACV-008) pin this from both
sides.

**Which reader the verb uses (REQ-ACV-012).** `workflowAuditPins` turns a read or
parse error into a zero `AuditConfig` (`audit_pin.go:58-64`), which would make an
unreadable file look like an absent one. The verb therefore calls
`loadWorkflowAuditSection` directly (for the tree, or for the primary checkout of
a config-orphaned worktree, through the same routing `resolveAuditGates` uses) so
it can tell absent (nil error, zero value) from unreadable (an error).
`audit_multi`'s own handler keeps `workflowAuditPins` and so keeps today's reading
of an unreadable file for a direct caller (spec.md R-4).

- `internal/config` cannot import `internal/cli`, so the resolver cannot sit
  beside the handler; it takes the already-loaded `AuditConfig` value and the
  caller's supplied gates and returns a value.
- `resolveAuditGates(root)` keeps its signature `(config.AuditGates, string)` but
  returns the plan's *explicit* gates. That one change makes the convergence
  enforcement and the single-backend `codex_audit` rule read through the
  resolver with no edit to either of them. The unidentifiable-primary rule (codex
  assumed `required`, with its note) is untouched.
- `internal/auditreceipt/store.go` has its own tiny YAML struct for the raw codex
  gate. It gains `audit.model` and calls the resolver for the codex entry. A
  missing file, unreadable file, or invalid value keeps reading as "not required"
  on this hook-side path; the loud error belongs to `audit_multi` and the verb.
- **Import weight (decision).** This adds the edge `internal/auditreceipt ->
  internal/config`. Measured: no cycle, and `internal/config` carries 29
  module-path dependencies; `internal/hook` already imports `internal/config`
  (research.md R-26), so the hook-side weight is already paid and the whole
  `moai` binary links every package either way. A leaf package for the pure
  resolver was rejected: its inputs and outputs are `config.AuditConfig` /
  `config.AuditGates`, so a leaf would duplicate those types or still import
  `internal/config`. Accepted.
- Existing structural guards bind where the code may go:
  `TestConvergence_NoNewAuditModelEnum` forbids defining `AuditModel* =` in
  `mcp_convergence.go`; `TestConvergence_NoDirectFrontmatterRead` forbids the
  words `frontmatter` / `agent_overrides` there; the web guard forbids the
  resolver symbol in `internal/web`. The resolver is outside all three.

## §D.4 The read-only surface: a verb in `moai verify` (D1, D10, B.2.2)

D1 left the choice between an MCP tool and a CLI verb open; the leader's ruling
D10 placed the verb. Measured costs of the alternatives:

| | New MCP tool (`audit_plan`) | CLI verb in `moai verify` | `audit_multi` plan-only flag |
|---|---|---|---|
| New code | catalogue entry (`internal/mcp/catalog.go`, 45 -> 46), registration in `mcp_server.go`, handler | one cobra file + test, registered through the existing `verifyExtraCommands` hook | a branch in the handler |
| Fan-out | the registration-vs-catalogue guard test, the write-capable-set pin, per-tool settings fields derived from the catalogue, the tool count in `moai-mcp-tools.md` and its catalogue companion (both say 45) and their template mirrors, the per-tool permission entries (`tool-policy.yaml`, with `make build` running `tool-policy-drift-check`), the `tools:` list of both auditors, the emitted `.codex` role TOMLs | none beyond the verb | none |
| Reaches a running session | **No, not until reconnect.** An MCP server keeps the build it started with (research.md R-13). | Yes — the installed binary is resolved per call | n/a |
| Read-only by the catalogue's own classification | yes | yes | **No** — `audit_multi` is catalogued write-capable |
| Harness-neutral | tool names differ per harness | any harness with a shell | no |

**Placement and spelling (D10).** `moai spec audit` already means SPEC-era
classification and drift findings; a top-level `moai audit` would be confusable
with it. The existing `moai verify` namespace holds `record`, `check`, `sync-gate`
and `codex-review` — named for the thing they act on, hyphenated when two words.
`audit-plan` follows that shape and names a *plan*, not an action; `plan-audit` and
`sync-audit` would read as "run the audit". One verb serves both auditors.

**Input — the `--project-root` trap (REQ-ACV-010).** `verify`'s persistent
`--project-root` flag defaults to `$CLAUDE_PROJECT_DIR`, then the working
directory (`internal/cli/verify.go:57,75-91`). In a worktree-isolated session
`CLAUDE_PROJECT_DIR` names the primary checkout, so a bare call would read the
wrong tree's configuration and say nothing about it. The auditors therefore pass
their own toplevel explicitly: one plain call to learn it (`git rev-parse
--show-toplevel`), then `moai verify audit-plan --project-root <that path>`. The
verb registers through `verifyExtraCommands` and resolves its root through
`verifyResolveRoot`.

**Unknown verbs do not fail (REQ-ACV-014).** Measured: `moai verify audit-plan
--help` on the installed binary prints the group's own help and exits 0
(research.md R-25). An older binary therefore cannot be detected by exit code, and
"no output contract" is not a safe detector — it would also cover a refused,
crashed or timed-out call. The old-binary case is recognised **positively**: the
output is the `verify` group's help, which contains `Shared diagnostic snapshot
contract` and contains neither the verb's JSON member `config_status` nor an
`audit-plan:` line. Any other way of not getting a plan is a different failure.
If the group's help text ever changes, the signature stops matching and the
outcome falls into the blocking branch — the safe direction.

## §D.5 Output of `moai verify audit-plan` (REQ-ACV-010..012, 016)

JSON on stdout, exit 0:

```json
{
  "project_root": "/abs/path/of/the/tree/that/was/read",
  "config_status": "ok",
  "model": "multi",
  "model_source": "config",
  "backends": [
    {"backend": "claude", "gate": "required", "source": "config.model", "explicit": true},
    {"backend": "codex",  "gate": "required", "source": "config.model", "explicit": true},
    {"backend": "glm",    "gate": "advisory", "source": "config.model", "explicit": true}
  ],
  "cross_model_active": true,
  "cross_model_required": true,
  "enforced_required": ["claude", "codex"]
}
```

- `config_status` is `ok`, `absent` (no `workflow.yaml` or no `audit` block — the
  distributed default plan, `model_source: "default"`, every `explicit: false`; a
  pins-only `audit` block is `absent` for this purpose), or `unreadable`.
- **Unreadable (REQ-ACV-012).** The plan is not guessed: `config_status:
  "unreadable"`, `cross_model_active: "unknown"`, `cross_model_required:
  "unknown"`, an empty `backends`, and a `note` naming the file and the cause. It
  is a distinct, non-passing state; it is not the default plan and not the legacy
  path (REQ-ACV-015).
- `cross_model_active`: some backend other than `claude` is explicit and not
  `off`. `cross_model_required`: some backend other than `claude` is explicit and
  `required`. `enforced_required`: every backend that is explicit and `required` —
  the set whose unmet gate fails the verdict and whose entries the checker
  inspects.
- Invalid configuration (REQ-ACV-011): nothing on stdout, the REQ-ACV-003 text on
  stderr as `audit-plan: audit_model "grok" unknown (want one of
  claude|codex|glm|multi)`, exit 1. A system error (cannot resolve a tree) exits 2
  per `internal/cli/CLAUDE.md`, also `audit-plan:`-prefixed.

**The checker (REQ-ACV-016, decision D12).** With `--result-file <path>` the verb
also compares the plan with an `audit_multi` result and adds

```json
"convergence_check": {"ok": false, "unmet": ["codex"], "reason": "codex: no per_backend_verdicts entry"}
```

`ok` is true only when, for every `enforced_required` backend, the result has a
`per_backend_verdicts` entry whose `verdict` is exactly `pass` or `fail` and whose
`gate` is `required` (a `verdict` that is missing, null, empty, `inconclusive` or
any other value — `"pas"`, say — is an unmet gate named in `convergence_check`),
**and** the result's `plan_source` is `config` whenever any plan
entry has a `config.*` source. A file that is missing, unreadable, a directory,
empty, larger than the size bound below, or not a JSON object is an `audit-plan:`
error on stderr, exit 1, with no plan on stdout — never a pass. The checker is
**pure**: its only inputs are the resolved plan and the one file it is handed; it
opens exactly the path the caller names, once, enumerates and reads nothing under
`.moai/state`, takes no session id, and changes nothing. It never changes the verdict; the caller reads it.
It is the production consumer of `plan_source` — an additive, `omitempty` member
of the result that an older server does not emit, which is the signal that the
server that produced the result did not read the configuration.

*The input form (amended at 0.1.5).* The result is passed as a FILE path, not as a
command-line JSON argument. Measured during M4: the worktree-isolation guard
refuses every command that carries braces or quotes, heredocs and compound commands
included, so an inline `--result '<json>'` cannot be run by an auditor in a guarded
session and is not kept. A path argument is a plain word: the guard ran
`moai verify audit-plan --result-file <abs path>` with the path under the
worktree's `.moai/state/` and with a path under `/tmp` (both reached the CLI, which
answered with its own unknown-flag error on the build that predates the flag;
neither was refused). The auditor writes the result with its `Write` tool and
passes only the path — the same family as `moai handoff save --stdin`, which reads
what the caller wrote. The file holds a digest, not the whole result — only
`overall_verdict`, `gate_unmet`, `plan_source` and, per backend, `backend`, `gate`
and `verdict`; the full `ConvergenceResult` is also accepted (unknown members are
ignored), since a file needs no shell quoting.

*Size bound.* The verb refuses a file larger than 256 KiB (262,144 bytes) with an
`audit-plan:` error, exit 1. Measured on this tree: three real persisted
`ConvergenceResult` files of the iteration audits are 4,626, 10,747 and 20,473
bytes; a digest is under 1 KiB. 256 KiB is about twelve times the largest measured
full result and a few hundred times a digest, so no real result is refused and a
runaway or wrong file is, without the verb reading an unbounded input.

*Where the auditor writes.* A path its `Write` tool can reach and the guard accepts:
a file under the audited worktree's own `.moai/state/` (gitignored — measured with
`git check-ignore`), e.g. `<toplevel>/.moai/state/audit-plan-result.json`, written
fresh before each check so a stale file at that name is overwritten. The auditor
tools differ: `plan-auditor` carries `Write` and `Edit`; `sync-auditor` carries
neither (`tools:` lists Read, Grep, Glob, Bash, task tools, Skill and the audit MCP
tools, with `permissionMode: plan`) and stays read-only (no `Write` is added; its
`tools:` list is not touched). The rule, which the M7 instruction text carries:
`plan-auditor` writes the file itself; in the sync phase the ORCHESTRATOR session,
which holds `Write`, writes the digest file and runs
`moai verify audit-plan --result-file <path>`, and a cold `sync-auditor` returns its
`audit_multi` result members (the digest members above) in its report for the
orchestrator to write and check.

*Why the earlier failure modes cannot arise.* The first design read the result the
server persisted under a session id. It needed the right reader for a
config-orphaned worktree's tree-qualified file name, a binding between the file
and the audit being checked, a validated id, and a guarantee that a failed save
did not leave an earlier pass in place. None of that exists now: there is no store
read, so no choice of loader and no tree-qualified name to get wrong; there is no
session id, so no id to turn into a path; the verb reads exactly the one path the
caller names, once, so it cannot pick a sibling tree's tree-qualified file or
select by a session prefix; the caller writes that file from its own call's result
immediately before the check, so a stale copy is overwritten; and a failed save of
the server's persisted copy changes nothing the checker looks at. The
price is that the checker trusts the caller to pass the result faithfully (spec.md
R-8); the receipt guard is the independent backstop.

**What a verb run writes (REQ-ACV-010).** The verb itself writes nothing. The CLI's
own start-up does: `Execute` runs `InitDependencies()` for every non-trivial
subcommand (`internal/cli/root.go:82-84`), which registers
`buildSessionEndHandler(cwd)` (`internal/cli/deps.go:235`); that function creates
`<cwd>/.moai/reports` whenever `<cwd>/.moai/logs` exists
(`deps.go:337-346`). Measured on the installed binary in a scratch directory
holding only `.moai/logs`: after `moai verify audit-plan` (an unknown verb on that
build) the directory held `logs`, `reports` and `state` (research.md R-33). The
requirement and the criterion are therefore stated against what is true and
testable: the verb writes no regular file under the audited tree's `.moai`, and
the start-up directory creation under the working directory is named as pre-existing
and out of this SPEC's reach.

## §D.6 `audit_multi` integration (REQ-ACV-006..009)

- The handler reads the call's `gates` object as *supplied-only* (today's reader
  fills the three defaults, which erases the difference between "caller said
  required" and "caller said nothing"). It asks the resolver for the audited
  tree's plan — fed by `workflowAuditPins`, raw — with the supplied gates; an
  error returns the one hard tool error (like the unusable `project_root`),
  because a mistyped token silently becoming the default is the failure this SPEC
  exists to remove.
- The plan's gates feed the existing fan-out unchanged. Nothing about the
  backends, the independence rule, the in-session anchor, or the convergence table
  (cases 1-4) changes.
- Enforcement is untouched code reading new values: `runMultiAudit` already asks
  `workflowAuditGates(root)` for the explicit gates and fails the overall verdict
  for an explicit `required` backend whose entry is `inconclusive`
  (`enforceRequiredGateUnmet`), keeping the backend's own entry `inconclusive` and
  naming it in `gate_unmet` and `residual_risk_note`. After §D.3, a `required` that
  came from the `multi` token is explicit. Claude is already forced
  explicit-required when its gate is required.
- `ConvergenceResult` gains one `omitempty` string, `plan_source`, set to `config`
  when any backend's gate came from the tree's configuration. With no
  configuration and no gates the member is absent, which REQ-ACV-007's "no member
  the pre-change result did not carry" pins by golden file. Its consumer is the
  checker of §D.5.
- Receipts: `recordAuditReceipt` runs when codex participated and
  `receiptCodexGateRequired(root)` — now resolver-based — is true, so under
  `model: multi` the id comes back as `audit_receipt` even when the gate was unmet,
  and the SubagentStop guard demands a corroborated receipt for an auditor PASS
  (REQ-ACV-009).
- GLM- and GPT-launched lanes are unchanged and need no requirement: the existing
  tests `TestAuditMulti_GPTAndGLMOriginsRunActualClaude_AC_CLA_009_010` hold the
  behaviour — `audit_multi` ignores a supplied `claude_verdict` when the launch
  provider is not `claude` and runs the independent Claude backend, and the Claude
  gate is forced explicit-required. The concurrent fan-out is likewise held by
  `TestRunMultiAudit_ParallelFanOut_AC_AMM_002` (the regression-guard
  requirements that restated these were cut, decision D13).

## §D.7 Auditor flow (REQ-ACV-013, 014, 015, 016)

Both agents, in this order:

1. Learn the agent's own toplevel (`git rev-parse --show-toplevel`, a plain call)
   and run `moai verify audit-plan --project-root <that path>`.
2. **Outcome of step 1.**
   - *A plan* (JSON with `config_status` ok or absent): continue.
   - *The `verify` group's help* (the positive old-binary signature of §D.4): the
     **legacy path** (REQ-ACV-014). The agent follows the pre-change instruction —
     reads the project's `audit_model` and picks the backend as the pre-change text
     said — with one fixed point: wherever it calls `audit_multi` it passes **no
     `gates` argument**, so a configured plan or gate is applied by a plan-aware
     server and not weakened by explicit default arguments; it calls `audit_multi`
     whenever the tree's `workflow.yaml` sets an audit model other than `claude` or
     any `audit.gates` key (closing the case of `model: claude` with
     `gates.glm: required`, which the pre-change text reviewed with Claude only).
     It records "plan surface unreachable, legacy path used" in Gaps
     (`verification-claim-integrity.md` §3.1). No new failure.
   - *`config_status: unreadable`, an `audit-plan:` error, or any other way of not
     getting a plan* — refused, crashed, timed out, non-zero exit without the
     marker, malformed output (REQ-ACV-015): record the cause as a Gap; no PASS on
     this audit. Not the legacy path.
3. `cross_model_active: false` — the single-model path of today: a Claude main
   session reviews in session; a GPT or GLM main session calls `claude_audit`. The
   distributed default and an explicit `claude` token land here.
4. `cross_model_active: true` — call `audit_multi` with no `gates` argument (so the
   tree's own plan applies), with `project_root` set to the agent's toplevel,
   `target`, `focus`, and `claude_verdict` only from a Claude main session. Fold
   the result with the existing table in `moai-ref-cross-model-audit`.
5. `plan-auditor`: write the digest of the result just obtained to a file (§D.5,
   "Where the auditor writes"), run `moai verify audit-plan --project-root <path>
   --result-file <file>` and read `convergence_check`. A cold `sync-auditor` has no
   `Write`: it returns the digest members of its result in its report and the sync
   orchestrator does this step (§D.8). `ok: false` —
   an unmet gate named in Gaps and Residual-risk, no PASS, and the agent says it is
   an unmet gate and not a reviewed defect.
6. `gate_unmet` non-empty in the result: the overall verdict is `fail`; same
   reporting, and the `audit_receipt` ids are cited in the `AUDIT-VERDICT` line.

**The line between unreachable and unmet (B.2.5).** Only the positive old-binary
signature falls back to the legacy path. A configured required backend that does
not answer (decision D3), an unreadable configuration, a verb that fails to run
any other way, and a result that fails the check are all fail-closed, by name. The
legacy path exists so that an old binary is never an outage; it does not exist to
excuse a backend or an environment failure.

The prose "per the project's `audit_model`" and the four-line mode list that asks
the agent to branch on a token move into the labelled legacy paragraph and out of
the main path of both agents and of the skill's "When to use" table. The local
`sync-auditor.md` carries a `<!-- moai:closure-second-review:start -->` … `:end
-->` block (lines 178-194 local, 162-178 template) between the `audit_multi` bullet
and the other bullets; the edit must keep those markers.

## §D.8 The sync verdict: 4-dimension kept, `audit_multi` added (REQ-ACV-017)

Why the sync half needs a change at all: `sync-audit-4dim.js` judges are Claude
Explore agents, and on a clean PASS the `Binding promotion` paragraph of
`workflows/sync.md` makes that verdict binding and forbids spawning the cold
`sync-auditor` — the only sync-path agent that calls `audit_multi`. Decision D6
keeps the 4-dimension verdict (and its A3 fast path) as the Claude input and adds
the cross-model call beside it. Decision D9 keeps the rule in prose and in the
machine-readable result; the Go `IsBinding` predicate, which has no production
caller and whose `(false, reason)` already means "spawn the cold auditor", is not
changed.

**Flow** (orchestrator-side, in `workflows/sync.md`, after FO-SYNC-1 returns):

1. The script runs exactly as today and returns its verdict.
2. The orchestrator applies the existing binding rule. On any existing fallback
   trigger (INCOMPLETE, a zero-scored dimension, a contested finding) it spawns the
   cold `sync-auditor` as today; that agent runs the verb and `audit_multi`
   (REQ-ACV-013) and returns its result members in its report, and the orchestrator
   then writes the digest file and runs the checker (REQ-ACV-016, step 5), making
   no other extra call on this branch.
3. On a clean PASS the orchestrator runs `moai verify audit-plan --project-root
   <own toplevel>` and applies the outcome table of §D.7 step 2: the positive
   old-binary signature — legacy path (the existing binding rule as today, A3
   unchanged) with the named Gap; unreadable, an error, or any other failure to
   run — a Gap, the verdict not binding; a plan with `cross_model_required: false`
   — binding as before, nothing called.
4. With `cross_model_required: true` it calls `audit_multi`: no `gates` argument,
   `project_root` set to its own toplevel, the same `target` the cold auditor would
   use for this sync, and as `claude_verdict` an object synthesized from the
   workflow output — verdict `pass`, a one-line summary naming the harmonic mean,
   threshold and tier, and the workflow's findings mapped to findings; no judge
   reasoning text (the independence rule). The call is sequential, after the
   workflow returns. It writes only audit receipts and convergence state, never the
   audited source or docs, so the sole-writer invariant of `sync.md`
   (`manager-docs` is the only source/doc writer) is untouched. In a GPT- or
   GLM-launched session the supplied verdict is ignored and the independent Claude
   backend runs (§D.6).
5. It writes the digest file (`<toplevel>/.moai/state/audit-plan-result.json`, with
   its own `Write`) and runs the checker (`--result-file <file>`, §D.5).
6. The sync verdict is **binding only when all hold**: the 4-dimension PASS; the
   `audit_multi` overall verdict `pass` with an empty `gate_unmet`; and
   `convergence_check.ok: true`.
7. Otherwise it is **not binding**, for a named reason: the convergence verdict is
   `fail` (a reviewed defect — its findings go to the report); `gate_unmet` names
   the missing backend(s); the checker names an unmet backend (including a result
   from a server that predates the plan); or `audit_multi` cannot be called while
   the plan is reachable and requires it ("audit_multi unreachable", a Gap). The
   4-dimension verdict is never used alone in place of the missing half. The cold
   `sync-auditor` is not spawned for these reasons: it would repeat the same call.

**The binding statement.** The orchestrator's statement of which verdict is
binding carries `cross_model: required (<backends>)`, `audit_multi:
pass|fail|unavailable`, `gate_unmet: <backends|none>`, `audit_receipt: <ids|none>`
and `plan_check: ok|unmet (<backends>)`; on a failure it reads `binding: no —
cross-model gate unmet: codex` (or `audit_multi unreachable`), never a bare
four-dimension PASS. Where the statement is persisted is not specified in `sync.md`
(plan.md §B OQ-9).

**Where it lives (one file pair).** `.claude/skills/moai/workflows/sync.md`, the
`Binding promotion` paragraph under FO-SYNC-1 (local copy and template copy): the
new steps, the rule, the statement lines, and a pointer sentence in FO-SYNC-1. The
workflow script `.claude/workflows/sync-audit-4dim.js` is not edited: its agents
are `Explore` agents with no write capability and its contract is "no agent call
between judge collection and the returned verdict"; `audit_multi` is catalogued
write-capable. Its header comment keeps describing an unconditional happy path
(spec.md R-11). Template copies of `sync.md` must carry no internal SPEC or card
identifiers (`TestTemplateNoInternalContentLeak`).

**Effect on the A3 fast path.** Kept for the Claude half: the four-dimension
verdict is still produced once, in parallel, and on a clean PASS still replaces the
cold auditor's Claude review. The cost is one `audit_multi` call per sync when
codex is required (up to three external legs, §D.10); nothing is added when it is
not, and nothing is added while the verb is absent (legacy path).

## §D.9 GLM-launched lanes

Unchanged and stated, not invented (§D.6): in a lane launched with the GLM backend
the in-session auditor subagent inherits the session model, so it is not a Claude
judge; `audit_multi` obtains the Claude verdict from the independent Claude
subscription backend (`source: "mcp_claude_audit"`, pinned `{claude-opus-5-5,
high}`) and forces the Claude gate explicit-required. Under `multi` the independent
judges in such a lane are the Claude backend and codex; GLM, the author's own
family there, stays advisory.

## §D.10 The codex turn, the deadline, and the budget (REQ-ACV-020)

**What bounds the `audit_multi` call.** The legs run concurrently through one
`errgroup`; wall-clock is the slowest leg. Before this SPEC: Claude `5 *
time.Minute` (`claudeAuditTimeout`, `mcp_claude.go:27`), GLM 120 s
(`glmAuditHTTPTimeout`, `mcp_glm.go:80`), codex no deadline of its own (the request
context only). The 900 s figure belongs to the two Stop-hook wrappers
(`config.DefaultCodexReviewGateTimeout` / `DefaultMultiReviewGateTimeout`) — and
**neither binds this call**: the multi-review gate only reads a result already
persisted by `persistConvergenceResult` (`internal/cli/multi_review_gate.go:13,79,
120`), and the codex review gate wraps its own review in its own timeout
(`codex_review_gate.go:96`); neither file references `runMultiAudit` or
`performCodexAudit` (research.md R-27). The auditors' `audit_multi` call is an MCP
tool call bounded by the host's tool timeout, none of which is configured in this
repository (its effect was not observed).

**Why a deadline alone is not enough — the turn reader.** `awaitCodexTurnReview`
(`internal/cli/mcp_codex.go:1260-1269`) returns `bestCodexReviewText(reviewText,
agentText), nil` in three places: when the context has ended (`:1264`), when the
stream closes (`:1268`), and on `turn/completed` (`:1319`). Only the last is a
completed turn; the other two return whatever partial review text exists with a nil
error, and `runTurn` (`:932-941`) turns text into a verdict — it already maps a
non-nil turn error to `inconclusive` ("whatever review text was collected is
codex's PLACEHOLDER … it must never reach the synthesizer"). The reader checks the
context only *between lines*, and `conn.recv()` blocks until codex emits a line or
the process dies (a test records that "closing the conn is what unblocks the
reader", `codex_protocol_liveness_test.go:329-332`). So a deadline manifests in
production as the process being killed (`exec.CommandContext`,
`mcp_codex.go:518-540`) and its stdout closing — i.e. as the **stream-closed**
arm. The deadline therefore only helps if that arm is also an `inconclusive`.
Measured (research.md R-32, acceptance.md E29): a stream cut after a review item
and before `turn/completed` currently yields the partial body's verdict (a `fail`
from a finding bullet, for the fixture used).

**The change (REQ-ACV-020).** Make the two non-completed returns errors:
`awaitCodexTurnReview` returns `("", err)` with the cause named — `codex review
turn ended by context: <ctx error>` and `codex stream closed before the turn
completed`. `runTurn` already turns that into an `inconclusive` carrying the cause;
nothing else changes in it. The reader has two callers of `runTurn`
(`codex_task.go:181` and `mcp_codex.go:1031`), so the review gate, `codex_audit` and
`codex_task` inherit it. `codex_task` races its own context against the turn and
is unaffected only for that context/timer race; for the stream-closed arm its
outcome changes from `completed` with partial output to `failed` with the cause
(spec.md R-13). Existing codex test families must stay
green (AC-ACV-020).

**The deadline, derived from the repository's codex siblings.**
`internal/config/defaults.go` already holds the codex bounds as `var`s, each
documented as distinct from the others so that tuning one cannot move another:
`DefaultCodexReviewGateTimeout` (900 s, a Stop hook), `DefaultCodexTaskTimeout`
(600 s, one `codex_task` turn — delegated work), and `DefaultCodexAuditTimeout`
(20 minutes, "ONE read-only audit process … An audit reads a SPEC and its tree and
can take minutes", `defaults.go:589-593`, consumed by `codexAuditExec` in
`internal/cli/codex_audit_launch.go:648`). The `audit_multi` codex leg is the same
kind of work as the last — a read-only adversarial review of a SPEC or diff — so:

1. *Value.* `DefaultCodexAuditLegTimeout = DefaultCodexAuditTimeout` (20 minutes), a
   derived package variable, not a restated literal; no measured codex leg duration
   is claimed (spec.md R-5).
2. *Why this sibling.* The task bound bounds delegated work, not a review; the
   Stop-hook bound bounds a hook and does not apply to this call; the Claude stage
   limit bounds a different backend.
3. *Home and test seam.* Beside those siblings in `defaults.go`, as a `var` like
   them ("Not a compile-time const … so a test can shorten it"). Tests shorten
   `config.DefaultCodexAuditLegTimeout` and restore it with `t.Cleanup`; no second
   seam variable and no constant in `mcp_codex.go`.
4. *Budget arithmetic.* With the legs concurrent, the worst-case wall-clock of one
   `audit_multi` call is max(Claude 5 min, GLM 120 s, codex 20 min) = 20 minutes,
   reached only when codex hangs. A shorter limit would turn a legitimately slow
   review into an unmet gate for the whole pipeline (spec.md R-2); no measurement
   says where that line is, so it is the operator-visible knob (plan.md §B OQ-8).
5. *Mechanism.* `performCodexAudit` (the `audit_multi` leg, in `mcp_convergence.go`)
   wraps its context in `context.WithTimeout` with the variable and passes it on to
   the session start, so the process is killed at the deadline and the stream-closed
   arm above fires. After the RPC returns, if the leg's own deadline ended it and
   the caller's context did not, the leg replaces the summary with `codex leg timed
   out after <limit>` (wording only); a caller cancellation keeps the reader's
   `ended by context` cause.
6. *Scope.* Only `audit_multi`'s codex leg gets the deadline; `codex_audit` keeps
   the request context. Under a `required` codex gate from the tree's configuration
   an `inconclusive` entry becomes an unmet gate naming codex (D3); under the
   engine default it stays the fail-open `inconclusive` it always was for a missing
   backend.

The plan can only choose among the three existing legs, so the worst case per
audit is three external calls and, with the plan-audit retry ceiling of 3, three
fan-outs per SPEC plan phase; on the sync side, one additional `audit_multi`
fan-out per sync when codex is required. Quota exhaustion is not predicted; it
appears as an `inconclusive` leg and, for codex under `multi`, as an unmet gate.

## §D.11 Config surface and template neutrality (REQ-ACV-019)

- Committed `.moai/config/sections/workflow.yaml` (a tracked dev-repo file): add
  `model: multi` under `audit:`, change the claude pin `effort: medium` to `high`,
  and correct the header comment ("Claude defaults to claude-opus-5-5/medium" ->
  `high`).
- `internal/template/templates/.moai/config/sections/workflow.yaml` already carries
  the claude pin at `high` and has no `model` or `gates` key (research.md R-8): **no
  template change is needed or made**, and the Go default `Audit.Model` stays
  `claude` (`TestAuditConfig_DefaultProfile`). The default lives in code and the
  distributed template must not leak the sub-key (read from the
  `internal/config/defaults.go` comment; `CLAUDE.local.md`, which names the
  principle, is absent from this worktree and the primary checkout). The
  distributed shape is therefore a pins-only `audit` block, which the resolver must
  read as the default plan (AC-ACV-008).
- Settings schema / wizard: `ValidAuditModels()` and the `closedSeam` fields in
  `internal/settings/schema_sections.go` are unchanged; only the comment that names
  `activeAuditBackend` is updated. The wizard keeps writing the token verbatim,
  which now takes effect; whether it persists the default `claude` unprompted is
  not observed, and D7' makes the answer harmless.
- `moai update`: `workflow.yaml` is merge-restored (3-way) and user values are
  preserved per `internal/cli/update_clean_install.go:495`; only the comment was
  read, not exercised.

## §D.12 Cleanup (REQ-ACV-018)

`multiConvergenceImplemented` and `activeAuditBackend` are removed: the convergence
engine exists and the resolver supersedes the validation. The three
`TestActiveAuditBackend_*` tests go with them. `buildAuditEnvBlock` stays. The web
guard `TestWebConsole_AuditNoForkedInterpreter` scans for the string
`activeAuditBackend`; once that symbol is gone the scan passes vacuously, so the
sentinel is retargeted to the resolver symbol and the test gains a positive control
— it fails if the sentinel is absent from non-test source.

## §D.13 Rollout — what is true, per session

The verb's source does not exist on develop until this card merges. The sequence is:
card merged into develop; a build made from develop is installed; **each** live
session reconnects its MCP server (or restarts); only then does the new path apply
to that session's later audits.

- While the verb is absent (an installed binary that predates it), the auditors and
  the sync step take the legacy path with a named Gap (REQ-ACV-014): the pre-change
  behaviour, so a missing verb blocks nothing.
- After the CLI is installed, the verb is reachable at once (it is resolved per
  call), but every live lane and leader session still holds the MCP server it
  started with, which predates the install. That session's `audit_multi` results
  lack `plan_source`, so the checker (REQ-ACV-016) reports an unmet gate by name —
  fail-closed, visible, and bounded to that session until it reconnects. This is a
  real window, not an absence of one. The documented check is `moai doctor`, which
  compares a running server with the installed binary.
- The activation text — the local copies of the two auditors, the skill and
  `sync.md` — is the card's last milestone (plan.md M7), so nothing instructs the new
  path before every mechanism it relies on is in the tree.
- The leader's hand-off after the install: reconnect or restart every live session
  that will run a plan-audit or sync-audit, and run `moai doctor` in each (plan.md
  §J).

## §D.14 Test hermeticity: one `TestMain` scrub (decision D13c)

The committed `model: multi` becomes visible to any test that resolves its project
root from `CLAUDE_PROJECT_DIR`, which lane sessions export. Measured
(research.md R-31): the only production code that resolves a root for the audit
gate readers from that variable is in `internal/cli` (`resolveProjectDir`);
`internal/hook/main_test.go:51,68` already unsets it for the whole hook test binary;
`internal/cli/main_test.go` already clears the ambient factory environment, `MOAI_HOME`,
git variables and sandboxes the home and receipt roots, but not
`CLAUDE_PROJECT_DIR`. One added `os.Unsetenv(config.EnvClaudeProjectDir)` in that
`TestMain`, plus one test that fails if the variable is set at test start, replaces
the per-test helper; CI never sets the variable, so nothing that passes there can
depend on the ambient value. `internal/auditreceipt` and `internal/config` receive
explicit tree roots and need nothing.
