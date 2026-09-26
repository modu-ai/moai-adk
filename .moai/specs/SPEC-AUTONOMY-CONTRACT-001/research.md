# research.md — SPEC-AUTONOMY-CONTRACT-001

> Plan-phase investigation, measured against worktree HEAD `ca1d5dc43` (branch `WT-contract-schema`,
> based on `develop`). Every statement below cites the file it was read from in this tree.

## Reuse Analysis — `internal/mission`

The card requires reusing the sealed mission contract behind `/moai goal --auto`. What was read:

| File | What it provides | Reuse decision |
|---|---|---|
| `internal/mission/contract.go` | `MissionContract` (JSON-tagged), closed `Action` constants, `SealMissionContract`: completeness check → clone + sort set fields → `json.Marshal` → SHA-256 → `SealedContract{Hash, Bytes}` | **Reused two ways.** (1) The technique (clone, sort, marshal, hash) is the SPEC contract digest algorithm. (2) The function itself is called on the SPEC contract's mission projection; its hash is recorded as `signature.mission_contract_sha256`. |
| `internal/mission/policy.go` | `ValidateMissionDecision(sealed, snapshot, decision, now)` — fail-closed decision validator; per-action required evidence; scope containment (`targetInsideScope`) | **Enabled, not called in A1.** The projection exists so A2 can call this validator on a signed SPEC contract instead of writing a second one. |
| `internal/mission/receipt.go` | `OperationReceipt` and the receipt state machine | Not reused in A1 (runtime operations are A2's domain). |
| `internal/mission/completion_receipt.go` | `Integrity` field digest pattern: zero the integrity field, canonicalize maps, marshal, SHA-256, compare with `crypto/subtle`; atomic write through `internal/atomicfile` | **Pattern reused.** The signature block's `contract_sha256` follows the same "digest over everything except the integrity-bearing field" rule, compared in constant time; writes go through `internal/atomicfile`. |
| `internal/mission/governance_receipt.go` | Path-contained receipt writer (`containedGovernancePath`), binding validation against expectations | Path-containment approach reused for resolving `.moai/specs/<ID>/contract.yaml` (reject symlink escape / `..`). |
| `internal/mission/auto_state.go` | Session-scoped mission state under `.moai/state/` | Not reused: a SPEC contract is a committed artifact, not session state. |

What is **new** (no mission equivalent): the YAML schema and strict decoding; the SPEC-contract fields
without a mission counterpart (`ownership.never`, `invariants`, `reobserve`, `review`, `plan_audit`);
the acceptance hash and AC count binding; the human-presence signing flow and signature block; batch
signing; the `moai contract` CLI; the `workflow.autonomy` configuration; the reason-code verifier.

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
