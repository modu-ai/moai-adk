# SPEC-AUDIT-MODEL-CONVERGE-001 — design

> Tier L design artifact. It carries the reasoning behind spec.md §B.2 and the
> shapes the run phase builds to. Measurements cited here are in research.md
> §R.1 (tree `c50da9c2f`). This file has no frontmatter `status:` — the SPEC's
> lifecycle state lives in spec.md only.

## §D.1 The resolver table (REQ-ACV-001)

One row per `audit.model` token. A cell is the gate that token assigns to the
backend. The source of every non-default cell is `config.model`.

| `audit.model` | claude | codex | glm | Reading |
|---|---|---|---|---|
| (empty / key absent) | required | required | advisory | The engine default. Not an operator choice, so not explicit: fail-open exactly as today. |
| `claude` | required | off | off | Claude alone. The auditors keep today's single-model path. |
| `codex` | off | required | off | Codex alone (the skill's existing "codex reviews alone"). |
| `glm` | off | off | required | GLM alone. |
| `multi` | required | required | advisory | The cross-model profile: codex required, GLM advisory, Claude the anchor. Same gates as the default; the difference is the source — an operator wrote it, so it is fail-closed. |

Why `multi` and the empty token share their gates: the distributed default
profile was always "claude + codex required, glm advisory"
(`internal/config/audit_models.go`, `NewDefaultWorkflowConfig`). What was missing
was a way to say "I mean it". The token is that statement. This is also why the
`claude` token cannot map to the default profile: it would make `claude`,
`multi` and the empty token indistinguishable in gates, and the token would
still decide nothing. The `claude` row is nevertheless the one reading the code
does not settle on its own (`audit_models.go` calls it both "the default
single-backend model" and part of a claude+codex default profile). Operator
decision D4 reads it as Claude alone, **adopted at confidence 0.57 — plan-audit
please re-weigh** (plan.md §B OQ-1).

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
byte-identity). It is recorded as a flagged assumption (plan.md §B OQ-3).

## §D.3 Where it lives and who calls it (REQ-ACV-004, REQ-ACV-005)

```
internal/config/audit_plan.go     the pure resolver  (no file, env or clock)
        ^                ^                 ^                    ^
internal/cli/            internal/cli/     internal/cli/        internal/auditreceipt/
mcp_audit_multi.go       mcp_worktree_     audit_plan_cmd.go    store.go
(fallback plan)          root.go           (moai audit plan)    (CodexGateRequired)
                         resolveAuditGates
                         -> enforceRequiredGateUnmet (convergence)
                         -> applyGateUnmet (codex_audit)
```

- `internal/config` cannot import `internal/cli`, so the resolver cannot sit
  beside the handler; it takes the already-loaded `AuditConfig` value and the
  caller's supplied gates and returns a value. Reading `workflow.yaml` stays
  where it is (`workflowAuditPins` in `internal/cli/audit_pin.go`; the
  config-orphaned-worktree routing in `resolveAuditGates`).
- `resolveAuditGates(root)` keeps its signature `(config.AuditGates, string)`
  but now returns the plan's *explicit* gates. That one change makes the
  convergence enforcement and the single-backend `codex_audit` rule read through
  the resolver with no edit to either of them. The unidentifiable-primary rule
  (codex assumed `required`, with its note) is untouched.
- `internal/auditreceipt/store.go` has its own tiny YAML struct for the raw
  codex gate. It gains `audit.model` and calls the resolver for the codex
  entry. This adds the import edge `internal/auditreceipt -> internal/config`;
  measured: `internal/config` has no dependency on `internal/auditreceipt` and
  `internal/auditreceipt` currently imports no other moai package (research.md
  R-14), so no cycle. The hook-side predicate keeps its established posture — a
  missing file, unreadable file, or invalid value reads as "not required"; the
  loud error belongs to `audit_multi` and `moai audit plan`.
- Existing structural guards bind where the code may go:
  `TestConvergence_NoNewAuditModelEnum` forbids defining `AuditModel* =` in
  `mcp_convergence.go`; `TestConvergence_NoDirectFrontmatterRead` forbids the
  words `frontmatter` / `agent_overrides` there; the web guard forbids the
  resolver symbol in `internal/web`. The resolver is outside all three.

## §D.4 The read-only surface: CLI verb, not MCP tool (D1, B.2.2)

D1 left the choice open. Measured costs:

| | New MCP tool (`audit_plan`) | CLI verb `moai audit plan` | `audit_multi` plan-only flag |
|---|---|---|---|
| New code | catalogue entry (`internal/mcp/catalog.go`, 45 -> 46), registration in `mcp_server.go`, handler | one cobra file + test | a branch in the handler |
| Fan-out | the registration-vs-catalogue guard test, the write-capable-set pin, per-tool settings fields derived from the catalogue, the tool count in `moai-mcp-tools.md` and its catalogue companion (both say 45) and their template mirrors, the per-tool permission entries (`tool-policy.yaml`, with `make build` running `tool-policy-drift-check`), the `tools:` list of both auditors, the emitted `.codex` role TOMLs | none beyond the verb | none |
| Reaches a running session | **No, not until reconnect.** An MCP server keeps the build it started with; this session's server is `rc.23 d194083fb` while the installed binary is `rc.24 c50da9c2f` (research.md R-13). A tool added at HEAD is absent from `tools/list` until then. | Yes — the installed binary is resolved per call | n/a |
| Resolves the caller's tree | needs the caller to pass `project_root` (the server cannot know a worktree switch) | inherits the caller's working directory; `--project-root` is optional | needs `project_root` |
| Read-only by the catalogue's own classification | yes | yes (no tool catalogue) | **No** — `audit_multi` is catalogued write-capable (it files receipts and state); a flag that makes a write-capable tool read-only is a mode switch a permission gate cannot see |
| Harness-neutral | tool names differ per harness | any harness with a shell (Claude and Codex both) | no |

The MCP-over-CLI rule (`moai-mcp-tools.md`) prefers an MCP tool only when it is
in the agent's `tools:` list; both auditors already carry `Bash`. The verb wins
on fan-out, on reaching running sessions, and on neutrality; it costs one
Bash call per audit. The spelling `moai audit plan` adds a top-level `audit`
group (none exists; `moai spec audit` is the SPEC-lifecycle audit and is
unrelated) — plan.md §B OQ-5.

A Bash call from a worktree-isolated session must be a plain command; the verb
takes no required argument precisely so the common call is
`moai audit plan` with nothing to compute.

## §D.5 Output of `moai audit plan` (REQ-ACV-010..012)

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

- `config_status` is `ok`, `absent` (no `workflow.yaml` or no `audit` block —
  the distributed default plan, `model_source: "default"`, every `explicit:
  false`), or `unreadable` (parse failure — the default plan plus a `note`,
  mirroring what `audit_multi` does today).
- `cross_model_active`: some backend other than `claude` is explicit and not
  `off`. `cross_model_required`: some backend other than `claude` is explicit
  and `required`. `enforced_required`: every backend that is explicit and
  `required` — the set whose unmet gate fails the verdict.
- Invalid configuration: nothing on stdout, the REQ-ACV-003 text on stderr
  (`audit_model "grok" unknown (want one of claude|codex|glm|multi)` — the
  wording the retired `activeAuditBackend` used), exit 1. A system error
  (cannot resolve a tree) exits 2 per `internal/cli/CLAUDE.md`.
- It creates no file, opens no state directory, calls no backend, and has no
  `AskUserQuestion` path (the static guard of `internal/cli/CLAUDE.md`).

## §D.6 `audit_multi` integration (REQ-ACV-006..009, 018, 019)

- The handler reads the call's `gates` object as *supplied-only* (today's
  reader fills the three defaults, which erases the difference between "caller
  said required" and "caller said nothing"). It asks the resolver for the
  audited tree's plan with the supplied gates; an error returns the one hard
  tool error (like the unusable `project_root`), because a mistyped token
  silently becoming the default is the failure this SPEC exists to remove.
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
  "no member the pre-change result did not carry" pins by golden file.
- Receipts: `recordAuditReceipt` runs when codex participated and
  `receiptCodexGateRequired(root)` — now resolver-based — is true, so under
  `model: multi` the id comes back as `audit_receipt` even when the gate was
  unmet, and the SubagentStop guard demands a corroborated receipt for an
  auditor PASS (REQ-ACV-009).

## §D.7 Auditor flow (REQ-ACV-013, 014)

Both agents, in this order:

1. `moai audit plan` for the agent's own tree (plain Bash; no argument when the
   working directory is the tree).
2. `cross_model_active: false` — the single-model path of today: a Claude main
   session reviews in session; a GPT or GLM main session calls `claude_audit`.
   The distributed default lands here (all entries non-explicit).
3. `cross_model_active: true` — call `audit_multi` with no `gates` argument
   (so the tree's own plan applies), with `project_root` set to the agent's
   toplevel, `target`, `focus`, and `claude_verdict` only from a Claude main
   session. Fold the result with the existing table in
   `moai-ref-cross-model-audit`.
4. `gate_unmet` non-empty — the overall verdict is `fail`; the report says so in
   Gaps and Residual-risk, names the backend, states that it is an unmet gate and
   not a reviewed defect, and cites `audit_receipt` in the `AUDIT-VERDICT` line
   (a FAIL line is not refused for lacking a receipt, but the receipt is the
   record that the audit was attempted).
5. `moai audit plan` itself unreachable (an older build, a refusal) — record
   "plan surface unreachable" as a Gap per `verification-claim-integrity.md`
   §3.1 and do not reach PASS; the orchestrator re-delegates after the build.

The prose "per the project's `audit_model`" and the four-line mode list that
asks the agent to branch on a token it cannot read mechanically are removed from
both agents and from the skill's "When to use" table.

## §D.8 The sync verdict: the 4-dimension verdict kept, `audit_multi` added (REQ-ACV-015, REQ-ACV-021)

Why the sync half needs a change at all: `sync-audit-4dim.js` judges are Claude
Explore agents, and on a clean PASS the `Binding promotion` paragraph of
`workflows/sync.md` makes that verdict binding and forbids spawning the cold
`sync-auditor` — the only sync-path agent that calls `audit_multi`. Wiring the
auditors alone would leave sync Claude-only whenever the 4-dimension verdict is
clean. Decision D6 keeps the 4-dimension verdict (and its A3 fast path) as the
Claude input and adds the cross-model call beside it.

**Flow** (orchestrator-side, in `workflows/sync.md`, after FO-SYNC-1 returns):

1. The script runs exactly as today and returns its verdict.
2. The orchestrator builds the `FourDimVerdict` and consults the binding
   predicate. On any existing fallback trigger (INCOMPLETE, a zero-scored
   dimension, a contested finding) it spawns the cold `sync-auditor` as today;
   that agent runs `moai audit plan` and `audit_multi` itself (REQ-ACV-013), so
   the orchestrator makes no extra call on this branch.
3. On a clean PASS the orchestrator runs `moai audit plan`. With
   `cross_model_required: false` (the distributed default included) the verdict
   is binding exactly as before and nothing is called.
4. With `cross_model_required: true` it calls `audit_multi`: no `gates` argument
   (the tree's plan applies), `project_root` set to the orchestrator's own
   toplevel, the same `target` the cold auditor would use for this sync, and as
   `claude_verdict` an object synthesized from the workflow output — verdict
   `pass`, a one-line summary naming the harmonic mean, threshold and tier, and
   the workflow's findings mapped to findings; no judge reasoning text (the
   independence rule). The call is sequential, after the workflow returns. It
   writes only audit receipts and convergence state, never the audited source or
   docs, so the sole-writer invariant of `sync.md` (`manager-docs` is the only
   source/doc writer) is untouched. In a GPT- or GLM-launched session the
   supplied verdict is ignored and the independent Claude backend runs
   (§D.9).
5. The sync verdict is **binding only when both** the 4-dimension PASS and the
   `audit_multi` convergence pass, with an empty `gate_unmet`.
6. Otherwise it is **not binding**, for a named reason: the convergence verdict
   is `fail` (a reviewed defect — its findings go to the report); `gate_unmet`
   names the missing backend(s); or `audit_multi` cannot be called (tool absent
   or refused — recorded as "audit_multi unreachable", a Gap under
   `verification-claim-integrity.md` §3.1). The 4-dimension verdict is never
   used alone in place of the missing half. The cold `sync-auditor` is not
   spawned for these reasons: it would repeat the same call. The operator
   re-runs sync when the backend is back.

**Where each piece lives** (answer to "which files"):

| Piece | File (local copy and template copy) | Change |
|---|---|---|
| The added call and the "both must pass" rule | `.claude/skills/moai/workflows/sync.md` — the `Binding promotion` paragraph under FO-SYNC-1 | new step and rule; the existing three fallback triggers stay; a sentence in FO-SYNC-1 points at it |
| What the verdict record shows | the same paragraph | the orchestrator's statement of which verdict is binding carries `cross_model: required (<backends>)`, `audit_multi: pass|fail|unavailable`, `gate_unmet: <backends|none>`, `audit_receipt: <ids|none>`; when unmet: `binding: no — cross-model gate unmet: codex` (or `audit_multi unreachable`), never a bare four-dimension PASS |
| Header and description text of the script | `.claude/workflows/sync-audit-4dim.js` — the `VERDICT SCOPING` header comment and `meta.description` | text only: the verdict is the Claude input and is binding together with a passing `audit_multi` where codex is required, and the call is made by the orchestrator because this script's agents are read-only. Phases, schemas, verdict computation, output shape: unchanged |
| The machine predicate | `internal/runtime/sync_4dim_binding.go` | `FourDimVerdict` gains the cross-model requirement, the outcome (`pass`, `fail`, or absent) and the `gate_unmet` backends; after the existing triggers, `IsBinding()` returns `(true, "")` when no cross-model backend is required, and otherwise `(false, "cross-model …")` unless the outcome is `pass` with nothing unmet. The reasons begin `cross-model`, which is how the orchestrator tells "do not spawn the cold auditor" from the existing fallback reasons |

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
it is not.

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

- Legs before this SPEC: Claude `5 * time.Minute` (`claudeAuditTimeout`,
  `mcp_claude.go:27`), GLM 120 s (`glmAuditHTTPTimeout`, `mcp_glm.go:80`), codex
  no deadline of its own (the request context only). The legs run concurrently
  through one `errgroup`; wall-clock is the slowest leg.
- The 900 s figure belongs to the two Stop-hook wrappers
  (`handle-codex-review-gate.sh`, `handle-multi-review-gate.sh`,
  `config.DefaultCodexReviewGateTimeout` / `DefaultMultiReviewGateTimeout`). It
  is not an `audit_multi` bound and this SPEC does not touch those hooks. The
  auditors' `audit_multi` call is an MCP tool call bounded by the host's tool
  timeout; none is configured in this repository.

**The codex deadline (D5), derived.**

1. *Value.* The review job the codex leg runs — an adversarial review of the
   same diff — is the job the Claude leg runs, and the Claude leg's stage limit
   is the nearest existing constant. The codex limit is therefore defined as
   `codexAuditTimeout = claudeAuditTimeout` (5 minutes), a derived constant, not
   a restated literal.
2. *Fit under the 900 s Stop-hook budget.* The legs are concurrent, so the
   fan-out is bounded by the slowest: max(300 s, 120 s, 300 s) = 300 s, leaving
   900 − 300 = 600 s of the hook budget for everything else. Even if the three
   legs ran back to back, 300 + 120 + 300 = 720 s stays under 900 s.
3. *Why not other values.* Shorter would invent a tighter limit for the codex
   leg than for the Claude leg doing the same work, with no measured codex
   duration to justify it. Longer has no support in the repository. Real codex
   review durations were not measured (spec.md R-5, plan.md §B OQ-8).
4. *Home.* Beside the other codex constants in `internal/cli/mcp_codex.go`,
   where the Claude and GLM limits sit beside their backends. A package
   variable initialised from the constant is the test seam (the Claude limit is
   a constant and cannot be shortened in a test).
5. *Mechanism.* `performCodexAudit` (the `audit_multi` leg, in
   `mcp_convergence.go`) wraps its context in `context.WithTimeout`. The codex
   session subprocess is started with `exec.CommandContext`, so expiry stops
   the process and closes its stdout, which ends the blocked read.
6. *Expiry is `inconclusive`, never a verdict.* `awaitCodexTurnReview` returns
   whatever review text it has when its context ends (research.md R-21). So when
   the leg context ended by deadline and the caller's context did not, the leg
   returns `inconclusiveReview("codex leg timed out after <limit>")` regardless of
   any text the session produced. Under a `required` codex gate from the tree's
   configuration that entry becomes an unmet gate naming codex (D3); under the
   engine default it stays the fail-open `inconclusive` it always was for a
   missing backend.
7. *Scope.* Only `audit_multi`'s codex leg. The single-backend `codex_audit`
   handler keeps the request context.

- The plan can only choose among the three existing legs, so the worst case per
  audit is three external calls (one required codex, one Claude subscription,
  one advisory GLM) and, with the plan-audit retry ceiling of 3, three fan-outs
  per SPEC plan phase; on the sync side, one additional `audit_multi` fan-out per
  sync when codex is required (§D.8). Quota exhaustion is not predicted; it
  appears as an `inconclusive` leg and, for codex under `multi`, as an unmet
  gate.
- Bounds this SPEC adds and tests: the codex-leg deadline (AC-ACV-020) and the
  concurrency regression of REQ-ACV-018 (AC-ACV-017).

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
  the `internal/config/defaults.go` comment.)
- Settings schema / wizard: `ValidAuditModels()` and the `closedSeam` fields in
  `internal/settings/schema_sections.go` are unchanged; only the comment that
  names `activeAuditBackend` is updated. The wizard keeps writing the token
  verbatim, which now takes effect. Whether the wizard persists the default
  `claude` token unprompted is not observed (spec.md R-6).
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
