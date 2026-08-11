# plan.md — SPEC-MCP-DEFAULT-ON-001

Implementation plan for the Epic's foundation SPEC. Milestones are ordered by **decision reversibility**: the contract changes that are hardest to walk back (the two amendments, the template shape) come first, and the mechanical documentation sweep comes last.

## §A. Context

### A.1 Base

Every citation in this plan was read at base commit `ed70e4354` in the worktree `.claude/worktrees/moai-mcp-integration`.

### A.2 The two `.mcp.json` files

| Path | State at base | State after this SPEC |
|---|---|---|
| `.mcp.json` (repo root) | 3 entries: `context7`, `chrome-devtools`, `playwright` | **4** entries: those three **plus** `moai` |
| `internal/template/templates/.mcp.json` | 3 entries, byte-identical to repo root | **1** entry: `moai` only |

They are byte-identical at base. **After this SPEC they diverge on purpose.** The repo root is a maintainer working tree where the third-party MCP servers are actively used; the template is a neutral distribution artifact whose active map is the harness's own server. A future reader who notices the divergence and "fixes" it will silently undo REQ-A-1 or REQ-A-2. This paragraph, plus REQ-A-2 in spec.md, is the record that the divergence is deliberate.

### A.3 Verified integration points

| Concern | Location | State at base |
|---|---|---|
| Tool registration | `internal/cli/mcp_server.go:105` `registerMoaiMCPTools` | 17 tools |
| Entry shape | `internal/cli/mcp_server.go:498` `buildMoaiMCPServerEntry` | `{"command":"moai","args":["mcp-server"]}`, no `env`; constants at `:35-48` |
| Provisioning gate | `internal/cli/init.go:159` `provisionMCPEntryIfOptedIn` | body opens `if !optedIn { return }` |
| Gate call site | `internal/cli/init.go:782` | passes `opts.MCPToolsOptIn` |
| Wizard question | `internal/cli/wizard/questions.go:440-446` | `mcp_tools_opt_in`, confirm, `Default: "false"` |
| Locale strings | `internal/cli/wizard/translations.go:170`, `:312`, `:454` | ko / ja / zh, all "Opt-in default-off (REQ-MCP-002)" |
| Template deployment | `internal/template/deployer.go:104` | `fs.WalkDir` with **no dotfile skip** — the template `.mcp.json` IS deployed |
| Update merge set | `internal/cli/update_template_sync.go:320-328` | returns only `.claude/settings.json` + `.moai/status_line.sh` |
| Neutrality guard | `internal/template/mcp_template_neutrality_test.go` | `mcpAllowedActiveKeys` = the 3 third-party keys (line ~37); assertions at 83 / 87 / 92 |
| Provisioning call-site test | `internal/cli/init_mcp_provision_test.go:27`, `:44`, `:87`, `:105` | OFF path, ON path, and that `runInit` calls the gate |

### A.4 The `moai update` clobber bug (REQ-A-4)

Two facts at base combine into a defect:

1. `deployer.go:104` walks the embedded FS with no dotfile skip, so `internal/template/templates/.mcp.json` **is** deployed into user projects.
2. `update_template_sync.go:320-328` omits `.mcp.json` from `collectMergeableFiles`, on the stated grounds that "MoAI no longer ships an MCP template (full MCP removal)".

Premise 2's comment contradicts fact 1 and also contradicts `internal/template/mcp_template_neutrality_test.go`, which exists precisely to guard the shipped file. The consequence is that `moai update` deploys the template `.mcp.json` over the user's copy without a 3-way merge, so any entry the user added is lost. Fixing this inside SPEC-A rather than deferring it is deliberate: this SPEC is the change that makes `.mcp.json` load-bearing for every user, so shipping the flip while the clobber remains would turn a latent bug into a routine one.

### A.5 PRESERVE list

- `internal/cli/mcp_server.go` tool registration, handlers, schemas — **no change**.
- `internal/cli/mcp_codex.go`, `mcp_glm.go`, `mcp_convergence*.go` — **no change**.
- `internal/web/**` — **no change** (SPEC-C).
- `.claude/agents/**` — **no change** (SPEC-B).
- The recipe catalogue and the `moai mcp add|remove|list` CLI — **no change**.
- The forbidden-token scan and `${VAR}`-literal check in `mcp_template_neutrality_test.go` — preserved; only the allowed-key set changes.

## §B. Milestones

### M1 — Amendments to the two completed SPECs (highest reversibility cost)

Ordered first because it is a **contract** change to closed work: if the owner's direction is going to be revisited, it should be revisited before any code moves.

- Amend `SPEC-MOAI-MCP-SERVER-001`: REQ-MCP-002 gate inversion, REQ-MCP-015 flag restatement, AC-MCP-002 + AC-MCP-006, HISTORY `### Amendments`, `version:` bump, `updated:` refresh, `status: completed` unchanged.
- Amend `SPEC-TREND-MCP-001`: REQ-TMC-003 inversion, REQ-TMC-001 count, AC-TMC-001 + AC-TMC-004, same HISTORY / version / status treatment. REQ-TMC-002 and REQ-TMC-004's `$comment` clause untouched.
- Both amendments carry the reversal rationale in prose next to the amended criterion (REQ-A-8).

**Schema tension, recorded rather than resolved.** The canonical `completed → in-progress (amendment)` transition in `.claude/rules/moai/development/spec-frontmatter-schema.md` pairs the `amendment_of:` field with a status change to `in-progress`. The owner directed that `status: completed` stay unchanged, so `amendment_of:` is omitted too (adding the field while the status stays `completed` would assert a transition that did not occur). The HISTORY `### Amendments` sub-section carries the record instead. Both amended SPECs state this in their own HISTORY. If a plan-auditor wants the canonical form instead, the change is mechanical: set both to `in-progress`, add `amendment_of:` self-referentially, and return them to `completed` at this SPEC's sync.

### M2 — Template + repo-root `.mcp.json` shape (REQ-A-1, REQ-A-2)

- Rewrite `internal/template/templates/.mcp.json` to the single `moai` entry, preserving `$schema` and `staggeredStartup` verbatim.
- Add `moai` to the repo-root `.mcp.json`, keeping the existing three.
- Update `mcpAllowedActiveKeys` and the three assertions in `internal/template/mcp_template_neutrality_test.go` (REQ-A-9); leave the forbidden-token regex set and the secret check alone.
- `make build`.

### M3 — Provisioning gate inversion (REQ-A-3)

- Flip the wizard question at `questions.go:440-446` to `Default: "true"`, keeping positive polarity. Rename the question ID away from `mcp_tools_opt_in` (its name asserts the old contract) and follow the rename through `applyWizardPage3ToOpts` and the `opts` field.
- Rename `provisionMCPEntryIfOptedIn` → a name describing default-on provisioning with an honored decline, and update the doc comment at `init.go:150-158`, which currently explains the opt-in rationale.
- Update `internal/cli/init_mcp_provision_test.go`: the OFF-path test (`:27`) becomes "explicit decline is honored", the ON-path test (`:44`) becomes the default path, and the `runInit`-calls-the-gate assertion (`:87`, `:105`) is preserved as-is — that assertion is about reachability, which the flip does not change.

### M4 — `moai update` 3-way merge (REQ-A-4)

- Add `.mcp.json` to `collectMergeableFiles` (`update_template_sync.go:320-328`).
- Replace the false comment at `:322` with an accurate one stating that `.mcp.json` IS shipped and IS merged.
- Add a test proving a user-added MCP entry survives `moai update`.

### M5 — Documentation reconciliation (REQ-A-5) — most mechanical, ordered last

- Three wizard locale strings (ko `translations.go:170`, ja `:312`, zh `:454`): replace "Opt-in default-off (REQ-MCP-002)" wording, and drop the `REQ-MCP-002` token — a REQ token in a distributed string is a §25 neutrality hazard independent of this SPEC.
- `.claude/rules/moai/core/settings-management.md:33` + template mirror: restate the "exactly three active third-party entries … the single local stdio server stays opt-in" paragraph.
- Template `CLAUDE.md:69` and `:230-236`: reconcile the MCP inventory, which currently names Context7 and claude-in-chrome as integrated.
- Template-First mirror + `make build` for every template-side edit (REQ-A-10).

## §C. Technical approach

The change is a **gate-direction flip plus a shape change**, not new machinery. Every seam it touches already exists: `buildMoaiMCPServerEntry` already produces the right entry, `provisionMoaiMCPServerEntryAt` already writes it through `mutateClaudeJSONAtomic`, and `runInit` already calls the gate. The work is inverting one boolean's default, reducing one JSON file, adding one string to one slice, and sweeping the prose that describes the old contract.

The one genuinely new artifact is the M4 merge-preservation test.

## §D. Anti-patterns

- **AP-A-1 — Reconciling the two `.mcp.json` files.** They diverge on purpose (§A.2).
- **AP-A-2 — Deleting the false comment at `update_template_sync.go:322` without replacing it.** The next reader will re-ask why `.mcp.json` is in the merge set; the corrected comment is the answer.
- **AP-A-3 — Negative-polarity wizard question.** "Skip MCP provisioning?" inverts the meaning of the confirm's default relative to every neighboring Page-3 question. Use a positive question with `Default: "true"`.
- **AP-A-4 — Weakening the neutrality guard while inverting it.** Only `mcpAllowedActiveKeys` and the three shape assertions change; the forbidden-token scan and the `${VAR}`-literal secret check stay.
- **AP-A-5 — Leaving a `REQ-` token in a distributed locale string.** The three translations currently embed `REQ-MCP-002`; the sweep removes it rather than updating it to a new REQ token.

## §E. Cross-references

- `SPEC-MOAI-MCP-SERVER-001` — amended by M1; the SPEC that built the server.
- `SPEC-TREND-MCP-001` — amended by M1; the SPEC that created the template `.mcp.json`.
- `SPEC-MCP-AGENT-WIRING-001` (SPEC-B) — depends on this SPEC.
- `SPEC-MCP-CONSOLE-001` (SPEC-C) — depends on SPEC-B.
- `.claude/rules/moai/development/spec-frontmatter-schema.md` § Status Transition Ownership Matrix — the amendment-transition tension recorded in M1.
