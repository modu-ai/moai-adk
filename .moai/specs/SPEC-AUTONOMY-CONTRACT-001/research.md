# research.md — SPEC-AUTONOMY-CONTRACT-001

> Plan-phase investigation, measured against worktree HEAD `ca1d5dc43` (branch `WT-contract-schema`,
> based on `develop`). Every statement below cites the file it was read from in this tree.

## Reuse Analysis — `internal/mission`

The card requires reusing the sealed mission contract behind `/moai goal --auto`. What was read:

| File | What it provides | Reuse decision |
|---|---|---|
| `internal/mission/contract.go` | `MissionContract` (JSON-tagged), closed `Action` constants, `SealMissionContract`: completeness check → clone + sort set fields → `json.Marshal` → SHA-256 → `SealedContract{Hash, Bytes}` | **Technique reused.** Clone, sort set fields, marshal, SHA-256 is the SPEC contract digest algorithm. Calling `SealMissionContract` on a mission projection is **deferred to A2** (iteration-1 plan-audit D3/D17). |
| `internal/mission/policy.go` | `ValidateMissionDecision(sealed, snapshot, decision, now)` — fail-closed decision validator; per-action required evidence; scope containment (`targetInsideScope`) | **Deferred to A2.** Its scope check is exact-or-prefix (`targetInsideScope`, policy.go:112-118), so a projected glob such as `internal/foo/**` would reject every in-scope target; A2 owns the glob-to-prefix translation (design.md § Forward Note — Mission Projection). |
| `internal/mission/receipt.go` | `OperationReceipt` and the receipt state machine | Not reused in A1 (runtime operations are A2's domain). |
| `internal/mission/completion_receipt.go` | `Integrity` field digest pattern: zero the integrity field, canonicalize maps, marshal, SHA-256, compare with `crypto/subtle`; atomic write through `internal/atomicfile` | **Pattern reused.** The signature block's `contract_sha256` follows the same "digest over everything except the integrity-bearing field" rule, compared in constant time; writes go through `internal/atomicfile`. |
| `internal/mission/governance_receipt.go` | Path-contained receipt writer (`containedGovernancePath`), binding validation against expectations | Path-containment approach reused for resolving `.moai/specs/<ID>/contract.yaml` (reject symlink escape / `..`). |
| `internal/mission/auto_state.go` | Session-scoped mission state under `.moai/state/` | Not reused: a SPEC contract is a committed artifact, not session state. |

What is **new** (no mission equivalent), or deferred: the YAML schema and strict decoding; the SPEC-contract fields
without a mission counterpart (`ownership.never`, `invariants`, `reobserve`, `review`, `plan_audit`);
the acceptance hash and AC count binding; the human-presence signing flow and signature block; batch
signing; the `moai contract` CLI; the `workflow.autonomy` configuration; the reason-code verifier; the
agent-environment refusal. Deferred to A2: the mission projection and its sealed hash.

Vocabulary mismatch observed: the mission action names use underscores (`local_develop_merge`,
`batch_push`) while the operator-approved design uses hyphenated tokens (`local-merge-develop`,
`push-develop`) and adds `worktree`, which has no mission action. Decision: the contract keeps the
approved hyphenated vocabulary and maps explicitly (`design.md` § Action Vocabulary) rather than
renaming either side.

## AC Count — where the canonical counter lives

- The AC-count semantics are published as an awk program between `# MOAI-AC-COUNTER-BEGIN` and
  `# MOAI-AC-COUNTER-END` in `.claude/agents/moai/manager-docs.md` (read in this tree). It honors a
  `moai-ac-prefix` declaration, counts distinct live IDs, excludes `[RETIRED]`/`[REF]`-marked IDs, and
  exits 3 on ambiguity.
- `internal/spec/ac_count_clause_test.go` extracts that program (`extractCounterCommand`) and checks the
  corpus snapshot `.moai/reports/t338/ac-count-baseline.txt`; `scripts/ac-baseline/check-staged.sh`
  uses the same extraction at commit time.
- `internal/spec/parser.go` `ParseAcceptanceCriteria` is a different parser (structural AC lines for
  lint) and does not implement the marker/prefix semantics, so it is not the count source.

Decision: port the awk counter to Go for verify (no subprocess, windows-capable), and pin the port to
the published program with a full-corpus parity test that reuses `extractCounterCommand`. Risk
recorded in `plan.md` § Risks (two implementations of one measurement).

New `acceptance.md` files are reported, not failed, by `TestACCounterFullCorpusMatchesBaseline`
(the "absent-from-snapshot … reported, not failed" branch), so this SPEC's own `acceptance.md` needs no
snapshot regeneration at plan time.

## Configuration

- `internal/config/types.go` models workflow sub-keys as small structs (`IntegrationLockConfig`,
  `SettingsDriftGateConfig`, `AgentStopGuardConfig`, …) on the workflow config type. The autonomy block
  follows that pattern.
- `internal/config/resolver.go` decodes YAML in non-strict mode ("unknown fields in yaml are silently
  ignored"), which is why A1 need not reserve A2's `new_api_detector` key (spec.md §C.3).
- `internal/config/envkeys.go` defines `EnvAutonomyTier = "MOAI_AUTONOMY_TIER"`. A1 adds no environment
  key; the config path `workflow.autonomy.mode` and value set `guided | contract` are disjoint from the
  tier tokens.
- Neither the local nor the template `workflow.yaml` currently contains an `autonomy` key (grep in this
  tree returned no match).

## CLI and constitution

- `internal/cli/spec_status.go` has `stdinIsTerminal()` with an overridable hook (`stdinIsTerminalFn`)
  for tests — the TTY seam the signer reuses.
- Cobra registration pattern: `internal/cli/integration.go` (`Use: "integration [command]"`) and
  `internal/cli/goal.go` (`Use: "goal"`). No `contract` command exists (grep returned no match).
- `internal/constitution/loader.go` `LoadRegistry` and `registry_path.go` `ResolveRegistryPath` load the
  zone registry (`.claude/rules/moai/core/zone-registry.md`, 108 `CONST-` mentions); `constitution:<glob>`
  invariants resolve against it.

## Related SPECs read

- `SPEC-GTD-AUTONOMY-001` (completed, Tier L) — origin of `internal/mission`.
- `SPEC-AUTONOMY-TIERS-001` (completed, Tier M) — `MOAI_AUTONOMY_TIER` mode token and permission bundle;
  orthogonal axis.
- `SPEC-AUTONOMY-RUN-GOAL-001` (completed, Tier M) — run-phase goal wrapping with Kickoff preserved.

## Iteration-2 measurements (plan-audit iteration 1 follow-up)

- `go list -deps ./internal/constitution | grep -xE 'os/exec|net'` → `net`; `go list -deps ./internal/config`
  → `net`; `go list -deps gopkg.in/yaml.v3 | grep -cxE 'os/exec|net'` → `0`. Hence the verification core
  takes config values and registry rule IDs as inputs and imports neither package (design.md § Package
  Layout).
- `go.mod` has no doublestar glob dependency (only `gopkg.in/yaml.v3` among the relevant libraries);
  the ownership matcher is built on the standard library.
- Agent markers: `internal/config/envkeys.go:494` defines `EnvClaudeCodeSessionID = "CLAUDE_CODE_SESSION_ID"`;
  `CLAUDECODE` was observed set in this session's Bash environment. No Codex session-marker constant was
  found in `internal/` (only `CODEX_HOME`, a user configuration location).
- Template scan: `grep -nE 'SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}|(^|[^A-Za-z0-9])t[0-9]{3,}([^0-9]|$)|20[0-9]{2}-[0-9]{2}-[0-9]{2}'`
  on `internal/template/templates/.moai/config/sections/workflow.yaml` → `exit=1` (no match); on a copy with
  `# see SPEC-AUTONOMY-CONTRACT-001` appended → line 263 printed, `exit=0`. The single-segment pattern
  `SPEC-[A-Z][A-Z0-9]+-[0-9]{3}` did **not** match the planted multi-segment ID (`exit=1`), so it is not used.

## Iteration-3 measurements (plan-audit iteration 2 follow-up)

- Frozen-file sources (design.md § Frozen Files): `.claude/rules/moai/core/zone-registry.md` has 57
  `zone: Frozen` entries (first at line 83) over 12 distinct `file:` values; `internal/hook/pre_tool.go:1231`
  declares `var frozenInstructionFiles = []string{"CLAUDE.md", "CLAUDE.local.md"}`, applied by basename in
  `checkHarnessFrozenZone` (loop at line 1242).
- AC-013 semantics: `go list ./internal/contract` → `directory not found`, `list_exit=1`;
  `./internal/atomicfile` → `list_exit=0`, `deps_exit=0`, `noexecnet_exit=0`; `./internal/config` →
  prints `net`, `noexecnet_exit=1`.
- AC pass convention: `go test -v -count=1 -run '^TestAC_CONTRACT_999' ./internal/atomicfile/` →
  exit 0, `ok … [no tests to run]`, zero `--- PASS:` lines; `-run '^TestReplace_OntoAbsentDestination'`
  → `--- PASS: TestReplace_OntoAbsentDestination (0.00s)`.
- Extended template scan (adds `A-Q[0-9]`, `[Tt]his repository`): current template → `exit=1`; copy with
  two planted comment lines → both printed (263, 264), `exit=0`.
