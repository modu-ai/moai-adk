# SPEC-AUDIT-MODEL-CONVERGE-001 — design

> Tier L design artifact (0.1.2). It carries the reasoning behind spec.md §B.2
> and the shapes the run phase builds to. Measurements cited here are in
> research.md §R.1 and §R.5 (tree `53a42f013`, code identical to `c50da9c2f`).
> This file has no frontmatter `status:` — the SPEC's lifecycle state lives in
> spec.md only.

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

Why `claude` is the asymmetric row (D7', and the error-direction reasoning kept
for the plan-audit): the code pairs the `claude` token with the default gates
(`defaults.go:1272-1304`), `audit_models.go` calls it both "the default
single-backend audit model" and part of the default profile, and whether the
settings wizard persists an untouched `claude` is unobserved. If `claude` meant
Claude alone and some project had persisted it, that project would silently lose
codex and GLM in `audit_multi` — a change in the direction of less review. If it
means "Claude explicit, rest default" and some operator wanted Claude alone, the
cost is one visible edit (`gates.codex: off`). The two errors are not symmetric,
so the row takes the safe reading. The `codex` and `glm` tokens are never a
default value, so their "alone" reading carries no such risk.

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
explicit `off` or `advisory` never blocks.

A caller-supplied `required` (source `argument`) is NOT explicit. Today it is
not enforced either, and changing that would alter existing callers (REQ-ACV-007
byte-identity). It is a flagged assumption (plan.md §B OQ-3).

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

**Raw reads only (REQ-ACV-004, plan-audit PA1-D5).** The resolver is fed the
`AuditConfig` value the tree's `workflow.yaml` carries and nothing else. The
three readers that feed it are the existing raw ones: `loadWorkflowAuditSection`
/ `workflowAuditPins` (`internal/cli/audit_pin.go:34-64`), the receipt store's
own struct (`internal/auditreceipt/store.go:236-257`), and `resolveAuditGates`
(`internal/cli/mcp_worktree_root.go:101-113`). None of them constructs
`NewDefaultWorkflowConfig()`; that function sets `Audit.Model: claude` and the
default gates, so any default-merged read would hand `claude` to every project
and — under any reading of the `claude` row — change behaviour for projects that
never wrote the key. A static check (AC-ACV-004) and two fixtures (a pins-only
`workflow.yaml` and an explicit `model: claude`, AC-ACV-008 and AC-ACV-006/011)
pin this from both sides.

**Which reader the verb uses (PA1-D2).** `workflowAuditPins` turns a read or
parse error into a zero `AuditConfig` (`audit_pin.go:58-64`), which would make
an unreadable file look like an absent one. The verb therefore calls
`loadWorkflowAuditSection` directly (for the tree, or for the primary checkout
of a config-orphaned worktree, through the same routing `resolveAuditGates`
uses) so it can tell absent (nil error, zero value) from unreadable (an error).
`audit_multi`'s own handler keeps `workflowAuditPins` and so keeps today's
reading of an unreadable file for a direct caller (spec.md R-4).

- `internal/config` cannot import `internal/cli`, so the resolver cannot sit
  beside the handler; it takes the already-loaded `AuditConfig` value and the
  caller's supplied gates and returns a value.
- `resolveAuditGates(root)` keeps its signature `(config.AuditGates, string)`
  but now returns the plan's *explicit* gates. That one change makes the
  convergence enforcement and the single-backend `codex_audit` rule read through
  the resolver with no edit to either of them. The unidentifiable-primary rule
  (codex assumed `required`, with its note) is untouched.
- `internal/auditreceipt/store.go` has its own tiny YAML struct for the raw
  codex gate. It gains `audit.model` and calls the resolver for the codex
  entry. A missing file, unreadable file, or invalid value keeps reading as
  "not required" on this hook-side path; the loud error belongs to
  `audit_multi` and the verb.
- **Import weight (decision, PA1-D11).** This adds the edge
  `internal/auditreceipt -> internal/config`. Measured: no cycle
  (`internal/config` has no dependency on `internal/auditreceipt`;
  `internal/auditreceipt` imports no other moai package today), and
  `internal/config` carries 29 module-path (third-party) dependencies. The hook
  path is the receipt store's heavy consumer, and `internal/hook` already
  imports `internal/config` (research.md R-26), so the hook-side import weight
  is already paid; the whole `moai` binary links every package either way. A
  separate leaf package for the pure resolver was considered and rejected: the
  resolver's inputs and outputs are `config.AuditConfig` / `config.AuditGates`,
  so a leaf would either duplicate those types or still import `internal/config`.
  Accepted.
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
| Reaches a running session | **No, not until reconnect.** An MCP server keeps the build it started with; this session's server is `rc.23 d194083fb` while the installed binary is `rc.24 c50da9c2f` (research.md R-13). | Yes — the installed binary is resolved per call | n/a |
| Read-only by the catalogue's own classification | yes | yes | **No** — `audit_multi` is catalogued write-capable |
| Harness-neutral | tool names differ per harness | any harness with a shell | no |

**Placement and spelling (D10).** `moai spec audit` already means SPEC-era
classification and drift findings; a top-level `moai audit` would be confusable
with it and adds a command group for one verb. The existing `moai verify`
namespace holds `record`, `check`, `sync-gate` and `codex-review` — named for the
thing they act on, hyphenated when two words. `audit-plan` follows that shape and
names a *plan*, not an action; the alternatives `plan-audit` and `sync-audit`
(the leader's example names) read as "run the plan audit" and would imply a verb
that performs an audit when this one never does. One verb serves both auditors.

**Input — the `--project-root` trap (REQ-ACV-012).** `verify`'s persistent
`--project-root` flag defaults to `$CLAUDE_PROJECT_DIR`, then the working
directory (`internal/cli/verify.go:57,75-91`). In a worktree-isolated session
`CLAUDE_PROJECT_DIR` names the primary checkout, not the worktree, so a bare call
would read the wrong tree's configuration and say nothing about it — the failure
`project_root` exists to prevent for the MCP tools. The auditors therefore pass
their own toplevel explicitly: one plain call to learn it (`git rev-parse
--show-toplevel`), then `moai verify audit-plan --project-root <that path>`. The
verb registers through `verifyExtraCommands` (the pattern `sync-gate` and
`codex-review` use) and resolves its root through `verifyResolveRoot`.

**Unknown verbs do not fail (REQ-ACV-014).** Measured: `moai verify audit-plan
--help` on the installed binary prints the group's own help and exits 0
(research.md R-25) — the group has no argument validation, so an unknown verb is
not an error. An older binary therefore cannot be detected by exit code. The
verb's **output contract** is what identifies it: stdout is one JSON object
carrying `config_status`, or stderr carries a line beginning `audit-plan:`. Any
other outcome — help text, "command not found", a refused call — means the verb
was not reached.

## §D.5 Output of `moai verify audit-plan` (REQ-ACV-010..012, 022, 023)

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
  distributed default plan, `model_source: "default"`, every `explicit: false`;
  a pins-only `audit` block is `absent` for this purpose), or `unreadable`.
- **Unreadable (REQ-ACV-023).** The plan is not guessed. The object carries
  `config_status: "unreadable"`, `cross_model_active: "unknown"`,
  `cross_model_required: "unknown"`, an empty `backends`, and a `note` naming the
  file and the cause. It is a distinct, non-passing state: the auditors and the
  sync orchestrator record it as a Gap and do not PASS; it is not the legacy path
  and not the default plan.
- `cross_model_active`: some backend other than `claude` is explicit and not
  `off`. `cross_model_required`: some backend other than `claude` is explicit and
  `required`. `enforced_required`: every backend that is explicit and `required`
  — the set whose unmet gate fails the verdict and whose entries the result check
  inspects.
- **Result check (REQ-ACV-022).** With `--check-session <id>` the verb also reads
  the convergence result the server persisted for that session
  (`.moai/state/audit-multi/<id>.json`, the tree-qualified name for a
  config-orphaned worktree, through the existing `loadConvergenceResult` reader)
  and adds
  ```json
  "convergence_check": {"session_id": "…", "found": true, "ok": false,
                        "unmet": ["codex"], "reason": "codex: no per_backend_verdicts entry"}
  ```
  `ok` is true only when, for every `enforced_required` backend, the result has a
  `per_backend_verdicts` entry whose `verdict` is not `inconclusive` and whose
  `gate` is `required`, **and** the result's `plan_source` is `config` whenever
  any plan entry has a `config.*` source. A missing file gives `found: false`,
  `ok: false`. The check never changes the verdict; the auditors read it. It is
  the production consumer of `plan_source` (an additive, `omitempty` member of
  the result that an older server does not emit — the signal that the server
  that produced the result did not read the configuration).
- Invalid configuration (REQ-ACV-011): nothing on stdout, the REQ-ACV-003 text on
  stderr as `audit-plan: audit_model "grok" unknown (want one of
  claude|codex|glm|multi)` (the wording the retired `activeAuditBackend` used),
  exit 1. A system error (cannot resolve a tree) exits 2 per
  `internal/cli/CLAUDE.md`, also `audit-plan:`-prefixed.
- It creates no file, opens no state directory for writing, calls no backend, and
  has no `AskUserQuestion` path (the static guard of `internal/cli/CLAUDE.md`).

## §D.6 `audit_multi` integration (REQ-ACV-006..009, 018, 019)

- The handler reads the call's `gates` object as *supplied-only* (today's
  reader fills the three defaults, which erases the difference between "caller
  said required" and "caller said nothing"). It asks the resolver for the
  audited tree's plan — fed by `workflowAuditPins`, raw — with the supplied
  gates; an error returns the one hard tool error (like the unusable
  `project_root`), because a mistyped token silently becoming the default is the
  failure this SPEC exists to remove.
- The plan's gates feed the existing fan-out unchanged. Nothing about the
  backends, the independence rule, the in-session anchor, or the convergence
  table (cases 1-4) changes.
- Enforcement is untouched code reading new values: `runMultiAudit` already asks
  `workflowAuditGates(root)` for the explicit gates and fails the overall
  verdict for an explicit `required` backend whose entry is `inconclusive`
  (`enforceRequiredGateUnmet`), keeping the backend's own entry `inconclusive`
  and naming it in `gate_unmet` and `residual_risk_note`. After §D.3, a
  `required` that came from the `multi` token is explicit, so codex missing /
  timed out / over quota under `model: multi` fails the verdict by name. Claude
  is already forced explicit-required when its gate is required.
- `ConvergenceResult` gains one `omitempty` string, `plan_source`, set to
  `config` when any backend's gate came from the tree's configuration. With no
  configuration and no gates the member is absent, which is what REQ-ACV-007's
  "no member the pre-change result did not carry" pins by golden file. Its
  consumer is the result check of §D.5.
- Receipts: `recordAuditReceipt` runs when codex participated and
  `receiptCodexGateRequired(root)` — now resolver-based — is true, so under
  `model: multi` the id comes back as `audit_receipt` even when the gate was
  unmet, and the SubagentStop guard demands a corroborated receipt for an
  auditor PASS (REQ-ACV-009).

## §D.7 Auditor flow (REQ-ACV-013, 014, 022, 023)

Both agents, in this order:

1. Learn the agent's own toplevel (`git rev-parse --show-toplevel`, a plain call)
   and run `moai verify audit-plan --project-root <that path>`.
2. **Outcome of step 1.**
   - *Verb not reached* (the output contract of §D.4 is absent — an installed
     binary that predates the verb, `command not found`, a refused call): the
     **legacy path**. The agent follows the pre-change instruction exactly —
     reads the project's `audit_model` as today's text says and picks the backend
     as today's text says — and records "plan surface unreachable, legacy path
     used" in Gaps (`verification-claim-integrity.md` §3.1: an absence is named,
     never silent). No new failure, no outage.
   - `config_status: unreadable`: record the file and cause as a Gap; no PASS on
     this audit. Not the legacy path.
   - `audit-plan:` error on stderr / exit 1 (invalid configuration): report the
     text as a Gap; no PASS.
   - A plan: continue.
3. `cross_model_active: false` — the single-model path of today: a Claude main
   session reviews in session; a GPT or GLM main session calls `claude_audit`.
   The distributed default and an explicit `claude` token land here.
4. `cross_model_active: true` — call `audit_multi` with no `gates` argument (so
   the tree's own plan applies), with `project_root` set to the agent's
   toplevel, a `session_id`, `target`, `focus`, and `claude_verdict` only from a
   Claude main session. Fold the result with the existing table in
   `moai-ref-cross-model-audit`.
5. Run `moai verify audit-plan --project-root <path> --check-session <session_id>`
   and read `convergence_check`. `ok: false` — an unmet gate named in Gaps and
   Residual-risk, no PASS, and the agent says it is an unmet gate and not a
   reviewed defect.
6. `gate_unmet` non-empty in the result: the overall verdict is `fail`; same
   reporting, and the `audit_receipt` ids are cited in the `AUDIT-VERDICT` line (a
   FAIL line is not refused for lacking a receipt, but the receipt is the record
   that the audit was attempted).

**The line between unreachable and unmet (B.2.6).** Only the absence of the
verb falls back to the legacy path. A configured required backend that does not
answer (decision D3), an unreadable configuration, and a result that fails the
check are all fail-closed, by name. The legacy path exists so that an old binary
is never an outage; it does not exist to excuse a backend.

The prose "per the project's `audit_model`" and the four-line mode list that
asks the agent to branch on a token move into the labelled legacy paragraph and
out of the main path of both agents and of the skill's "When to use" table. The
local `sync-auditor.md` carries a `<!-- moai:closure-second-review:start -->` …
`:end -->` block (lines 178-194 local, 162-178 template) between the
`audit_multi` bullet and the other bullets; the edit must keep those markers.

## §D.8 The sync verdict: 4-dimension kept, `audit_multi` added (REQ-ACV-015, 021, 024)

Why the sync half needs a change at all: `sync-audit-4dim.js` judges are Claude
Explore agents, and on a clean PASS the `Binding promotion` paragraph of
`workflows/sync.md` makes that verdict binding and forbids spawning the cold
`sync-auditor` — the only sync-path agent that calls `audit_multi`. Decision D6
keeps the 4-dimension verdict (and its A3 fast path) as the Claude input and adds
the cross-model call beside it. Decision D9 keeps the rule in prose and in the
machine-readable result; the Go `IsBinding` predicate, which has no production
caller and whose `(false, reason)` already means "spawn the cold auditor", is
**not** changed (no overload, no new meaning for that return value).

**Flow** (orchestrator-side, in `workflows/sync.md`, after FO-SYNC-1 returns):

1. The script runs exactly as today and returns its verdict.
2. The orchestrator applies the existing binding rule. On any existing fallback
   trigger (INCOMPLETE, a zero-scored dimension, a contested finding) it spawns
   the cold `sync-auditor` as today; that agent runs the verb, `audit_multi` and
   the result check itself (REQ-ACV-013, 022), so the orchestrator makes no extra
   call on this branch.
3. On a clean PASS the orchestrator runs `moai verify audit-plan --project-root
   <own toplevel>` and applies the same outcome table as §D.7 step 2: verb not
   reached — legacy path (the existing binding rule as today, A3 unchanged) with
   the named Gap; unreadable or invalid — a Gap, verdict not binding; a plan with
   `cross_model_required: false` — binding as before, nothing called.
4. With `cross_model_required: true` it calls `audit_multi`: no `gates` argument,
   `project_root` set to its own toplevel, a `session_id`, the same `target` the
   cold auditor would use for this sync, and as `claude_verdict` an object
   synthesized from the workflow output — verdict `pass`, a one-line summary
   naming the harmonic mean, threshold and tier, and the workflow's findings
   mapped to findings; no judge reasoning text (the independence rule). The call
   is sequential, after the workflow returns. It writes only audit receipts and
   convergence state, never the audited source or docs, so the sole-writer
   invariant of `sync.md` (`manager-docs` is the only source/doc writer) is
   untouched. In a GPT- or GLM-launched session the supplied verdict is ignored
   and the independent Claude backend runs (§D.9).
5. It runs the result check (`--check-session <session_id>`, §D.5).
6. The sync verdict is **binding only when all hold**: the 4-dimension PASS; the
   `audit_multi` overall verdict `pass` with an empty `gate_unmet`; and
   `convergence_check.ok: true`.
7. Otherwise it is **not binding**, for a named reason: the convergence verdict
   is `fail` (a reviewed defect — its findings go to the report); `gate_unmet`
   names the missing backend(s); the check names an unmet backend (including a
   result from a server that predates the plan); or `audit_multi` cannot be
   called while the plan is reachable and requires it ("audit_multi
   unreachable", a Gap). The 4-dimension verdict is never used alone in place of
   the missing half. The cold `sync-auditor` is not spawned for these reasons: it
   would repeat the same call. The operator re-runs sync when the backend is back.

**The binding statement (REQ-ACV-021).** The orchestrator's statement of which
verdict is binding carries `cross_model: required (<backends>)`,
`audit_multi: pass|fail|unavailable`, `gate_unmet: <backends|none>`,
`audit_receipt: <ids|none>` and `plan_check: ok|unmet (<backends>)`; on a failure
it reads `binding: no — cross-model gate unmet: codex` (or `audit_multi
unreachable`), never a bare four-dimension PASS. Where the statement is
persisted is not specified in `sync.md` (plan.md §B OQ-9).

**Where each piece lives** (answer to "which files"):

| Piece | File (local copy and template copy) | Change |
|---|---|---|
| The added call, the check, and the "both must pass" rule | `.claude/skills/moai/workflows/sync.md` — the `Binding promotion` paragraph under FO-SYNC-1 | new steps and rule; the three existing fallback triggers stay; a sentence in FO-SYNC-1 points at it |
| What the verdict record shows | the same paragraph | the statement lines above |
| Header and description text of the script | `.claude/workflows/sync-audit-4dim.js` — the `VERDICT SCOPING` header comment and `meta.description` | text only: the verdict is the Claude input and is binding together with a passing `audit_multi` where codex is required, and the call is made by the orchestrator because this script's agents are read-only. Phases, schemas, verdict computation, output shape: unchanged |

The script is not given the call because its agents are `Explore` agents with no
write capability and the script's own contract is "no agent call between judge
collection and the returned verdict" (pure JS aggregation); `audit_multi` is
catalogued write-capable. Template copies of `sync.md` and the script must carry
no internal SPEC or card identifiers (`TestTemplateNoInternalContentLeak`; the
script's only allowlisted identifier is the `SPEC-FOO-001` launch example).

**Effect on the A3 fast path.** Kept for the Claude half: the four-dimension
verdict is still produced once, in parallel, and on a clean PASS still replaces
the cold auditor's Claude review. The cost is one `audit_multi` call per sync
when codex is required (up to three external legs, §D.10); nothing is added when
it is not, and nothing is added while the verb is unreachable (legacy path).

## §D.9 GLM-launched lanes (REQ-ACV-019)

Unchanged and stated, not invented. In a lane launched with the GLM backend, the
in-session auditor subagent inherits the session model, so it is not a Claude
judge. `audit_multi` already ignores a supplied `claude_verdict` when the launch
provider is not `claude` and obtains the Claude verdict from the independent
Claude subscription backend (`source: "mcp_claude_audit"`, pinned
`{claude-opus-5-5, high}`); the Claude gate is forced explicit-required, so an
unavailable Claude backend fails the verdict too. Under `multi` the independent
judges in such a lane are therefore the Claude backend and codex; GLM, the
author's own family there, stays advisory.

## §D.10 Budget and the codex-leg deadline (REQ-ACV-018, REQ-ACV-020)

**What bounds the `audit_multi` call.** The legs run concurrently through one
`errgroup`; wall-clock is the slowest leg. Before this SPEC: Claude
`5 * time.Minute` (`claudeAuditTimeout`, `mcp_claude.go:27`), GLM 120 s
(`glmAuditHTTPTimeout`, `mcp_glm.go:80`), codex no deadline of its own (the
request context only). The 900 s figure belongs to the two Stop-hook wrappers
(`handle-codex-review-gate.sh`, `handle-multi-review-gate.sh`,
`config.DefaultCodexReviewGateTimeout` / `DefaultMultiReviewGateTimeout`) — and
**neither binds this call**: the multi-review gate only reads a result already
persisted by `persistConvergenceResult` (`internal/cli/multi_review_gate.go:13,79,
120`) and never runs `audit_multi`, and the codex review gate runs its own review,
not `audit_multi` (neither file references `runMultiAudit` or `performCodexAudit`;
research.md R-27). The auditors' `audit_multi` call is an MCP tool call bounded
by the host's tool timeout, none of which is configured in this repository (its
effect was not observed). The earlier "900 − 300 = 600 s margin" arithmetic
therefore had no subject and is withdrawn.

**The codex deadline (D5), derived from the repository's codex siblings.**
`internal/config/defaults.go` already holds the codex bounds as `var`s, each
documented as distinct from the others so that tuning one cannot move another:
`DefaultCodexReviewGateTimeout` (900 s, a Stop hook), `DefaultCodexTaskTimeout`
(600 s, one `codex_task` turn — delegated work), and `DefaultCodexAuditTimeout`
(20 minutes, "ONE read-only audit process … An audit reads a SPEC and its tree and
can take minutes", `defaults.go:589-593`, consumed by `codexAuditExec` in
`internal/cli/codex_audit_launch.go:648`). The `audit_multi` codex leg is the
same kind of work as the last of these — a read-only adversarial review of a SPEC
or diff — so its limit is derived from it:

1. *Value.* `DefaultCodexAuditLegTimeout = DefaultCodexAuditTimeout` (20 minutes),
   a derived package variable, not a restated literal. No new number is
   introduced and no measured codex leg duration is claimed (spec.md R-5).
2. *Why this sibling and not the others.* The task bound (600 s) bounds delegated
   work, not a review; the Stop-hook bound (900 s) bounds a hook that must not
   stall a commit and does not apply to this call (above); the Claude stage limit
   (5 minutes) bounds a different backend. The audit bound's own comment says
   what a review of a SPEC and its tree costs.
3. *Home and test seam.* Beside those siblings in `defaults.go`, as a `var` like
   them ("Not a compile-time const … so a test can shorten it"). The test
   shortens `config.DefaultCodexAuditLegTimeout` and restores it with
   `t.Cleanup`; no second seam variable and no constant in `mcp_codex.go`.
4. *Budget arithmetic.* With the legs concurrent, the worst-case wall-clock of one
   `audit_multi` call is max(Claude 5 min, GLM 120 s, codex 20 min) = 20 minutes,
   reached only when codex hangs; a healthy fan-out ends with its slowest healthy
   leg. The cost of the derivation is that a hung codex can delay an audit by up
   to 20 minutes before it fails closed, by name. A shorter limit would turn a
   legitimately slow review into an unmet gate for the whole pipeline (spec.md
   R-2); no measurement says where that line is, so it is the operator-visible
   knob (plan.md §B OQ-8).
5. *Mechanism.* `performCodexAudit` (the `audit_multi` leg, in
   `mcp_convergence.go`) wraps its context in `context.WithTimeout` with the
   variable. The codex session subprocess is started with `exec.CommandContext`
   (`mcp_codex.go:518-540`), so any end of the context stops the process and
   closes its stdout, which ends the blocked read.
6. *Any end of the leg context is `inconclusive` (PA1-D3).*
   `awaitCodexTurnReview` returns `bestCodexReviewText(reviewText, agentText), nil`
   on **any** context end (`mcp_codex.go:1263-1266`), so partial review text can
   reach the parser whether the context ended by the leg's own deadline or by the
   caller's cancellation. The leg therefore decides from the context, not from
   the returned text: after the RPC returns, if the leg context has ended for any
   reason, the leg returns `inconclusiveReview(...)` and discards whatever text
   the session produced. Only the summary wording differs — `codex leg timed out
   after <limit>` when the leg's own deadline ended it and the caller's context
   did not, `codex leg cancelled` otherwise. Under a `required` codex gate from
   the tree's configuration that entry becomes an unmet gate naming codex (D3);
   under the engine default it stays the fail-open `inconclusive` it always was
   for a missing backend.
7. *Scope.* Only `audit_multi`'s codex leg. The single-backend `codex_audit`
   handler keeps the request context.

- The plan can only choose among the three existing legs, so the worst case per
  audit is three external calls (one required codex, one Claude subscription,
  one advisory GLM) and, with the plan-audit retry ceiling of 3, three fan-outs
  per SPEC plan phase; on the sync side, one additional `audit_multi` fan-out per
  sync when codex is required (§D.8). Quota exhaustion is not predicted; it
  appears as an `inconclusive` leg and, for codex under `multi`, as an unmet
  gate.
- Bounds this SPEC adds and tests: the codex-leg deadline and cancellation arm
  (AC-ACV-020) and the concurrency regression of REQ-ACV-018 (AC-ACV-017).

## §D.11 Config surface and template neutrality (REQ-ACV-017)

- Committed `.moai/config/sections/workflow.yaml` (a tracked dev-repo file):
  add `model: multi` under `audit:`, change the claude pin `effort: medium` to
  `high`, and correct the header comment ("Claude defaults to
  claude-opus-5-5/medium" -> `high`).
- `internal/template/templates/.moai/config/sections/workflow.yaml` already
  carries the claude pin at `high` and has no `model` or `gates` key (research.md
  R-8): **no template change is needed or made**, and the Go default
  `Audit.Model` stays `claude` (`TestAuditConfig_DefaultProfile`). The same
  principle the Go default already documents for `session_worktree` applies —
  the default lives in code and the distributed template must not leak the
  sub-key. (`CLAUDE.local.md`, the document that names that principle, is not
  present in this worktree or the primary checkout; the principle is read from
  the `internal/config/defaults.go` comment.) The distributed shape is therefore
  a pins-only `audit` block, which the resolver must read as the default plan
  (AC-ACV-008).
- Settings schema / wizard: `ValidAuditModels()` and the `closedSeam` fields in
  `internal/settings/schema_sections.go` are unchanged; only the comment that
  names `activeAuditBackend` is updated. The wizard keeps writing the token
  verbatim, which now takes effect. Whether the wizard persists the default
  `claude` token unprompted is not observed (spec.md R-6); D7' makes the answer
  harmless.
- `moai update`: `workflow.yaml` is merge-restored (3-way) and user values are
  preserved per `internal/cli/update_clean_install.go:495`; only the comment was
  read, not exercised. A tracked local value `model: multi` is expected to
  survive.

## §D.12 Cleanup (REQ-ACV-016)

`multiConvergenceImplemented` and `activeAuditBackend` are removed: the
convergence engine exists and the resolver supersedes the validation. The three
`TestActiveAuditBackend_*` tests go with them. `buildAuditEnvBlock` is not part
of this SPEC and stays. The web guard `TestWebConsole_AuditNoForkedInterpreter`
scans for the string `activeAuditBackend`; once that symbol is gone the scan
passes vacuously, so the sentinel is retargeted to the resolver symbol and the
test gains a positive control — it fails if the sentinel is absent from
non-test source.

## §D.13 Rollout (D8) — behaviour first, ordering second

The verb's source does not exist on develop until this card merges, so "install
the build before the card merges" is impossible. The sequence is: card merged
into develop; a build made from develop is installed; the MCP server is
reconnected; only then does the new audit path apply to later audits. Until
then the auditors and the sync step take the legacy path (REQ-ACV-014), which is
the pre-change behaviour, so the window between merge and install is not an
outage and not fail-closed. The activation text — the local copies of the two
auditors, the skill and `sync.md` — is the card's last milestone (plan.md M7)
so that nothing instructs the new path before every mechanism it relies on is
in the tree. The single window that remains is a new CLI with the old MCP
server still running: the verb answers, the old server's result has no
`plan_source`, and the result check fails closed by name until the server is
reconnected (spec.md R-8). The Definition of Done carries the install,
reconnect and smoke steps (plan.md §J).
