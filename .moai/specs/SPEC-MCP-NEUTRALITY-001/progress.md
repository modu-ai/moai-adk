# SPEC-MCP-NEUTRALITY-001 — progress

## §E.1 Plan-phase Audit-Ready Signal

audit_ready: true
plan_status: audit-ready
plan_complete_at: 2026-10-14
artifacts: spec.md, plan.md, acceptance.md, design.md, research.md (Tier L) + decision-index.md (decision_gate: on)
plan_audit_verdict: PASS-WITH-DEBT (r3, overall_score 0.91; trajectory 0.625 → 0.825 → 0.91)
plan_audit_sha: 9be172fce7
plan_artifact_hash: 61225de45bf1ed864b5af169632dca8db6499c1c4ba6216aaaee3937f599270c
plan_audit_receipt: rcpt-7323ca7bef2c885633529458 (codex required-gate pass, 0 findings)
plan_audit_verdict_file: .moai/reports/t1632/plan-audit-r3.md
debts: N3 (serial-slot instance red at M1 RED), N4 (AC-MN-023 reuse-refusal + restart-persistence depth, M6), N5 (AC-MN-012 "without stalling" quantification, M3)

## §E.2 Run-phase Evidence

### M1 — structured refusals + identity (card t1632, branch WT-3-2-0-2, HEAD 6199704240)

Commits (oldest first): `ebba213f0` (refusal contract + FactoryRoleOperator + ErrRunNotActive sentinel + spec.md draft→in-progress), `9e7a4a9a9` (handler wiring + instruments), `92cabf6ef` (observer split), `6199704240` (Codex env allowlist). Push state: nothing pushed (lane protocol — merged-local comes later); `git status -sb` = `## WT-3-2-0-2` with no upstream and a clean tree.

E8 — verbatim RED outputs (captured BEFORE each GREEN; final pre-GREEN run):

```
$ go test ./internal/cli -run '^(TestMCPRefusal_SerialSlotStructured|TestMCPFactoryActor_OwnedAndStaleRun|TestMCPOperatorMark_Required|TestTodoClaimMCP_Mirror)$' -count=1   → exit 1
--- FAIL: TestMCPRefusal_SerialSlotStructured (2.22s)
    mcp_refusal_test.go:135: refusal rode the error channel (is_error true); text=factory_next: factory next: refused serial-slot: another serial card is in flight and holds the serial slot
--- FAIL: TestMCPFactoryActor_OwnedAndStaleRun (9.65s)      [actor lane-2 arm: the stage SUCCEEDED as env lane-1 — the actor argument was ignored; stale-run arm: toolErr "factory card not found"]
--- FAIL: TestMCPOperatorMark_Required (12.95s)             [unmarked decide/todo_claim/todo_add arms SUCCEEDED — env absence granted operator]
--- FAIL: TestTodoClaimMCP_Mirror (0.58s)
    todo_claim_test.go:410: refusal rode the error channel (is_error true); text=todo_claim: moai claim: refused — lane boundary: ...
```

```
$ go test ./internal/hook -run '^TestFailureObserver_ExcludesPolicyRefusals$' -count=1   → exit 1
    failure_observer_refusal_test.go:79: inbox rows = 3, want exactly 1 (the genuine error only):
    [tool_failure:mcp__moai__factory_next:UnknownFailure, tool_failure:mcp__moai__todo_claim:UnknownFailure, tool_failure:Bash:ExitError]
```

AC matrix (M1-mapped rows; every row's command run at HEAD 6199704240, this tree):

| AC | Status | Command (verbatim) | Observed output |
|----|--------|--------------------|-----------------|
| AC-MN-001 (serial-slot instance; card gate 1) | PASS | `go test ./internal/cli -run '^TestMCPRefusal_SerialSlotStructured$' -count=1` | `ok github.com/modu-ai/moai-adk/internal/cli` — refusal is_error false, `refused:true, code:"serial-slot"`, text block carries the same JSON |
| AC-MN-001 class flip (EV-001 instrument) | PASS | same run selector incl. `TestTodoClaimMCP_Mirror` | `ok` — the lane-session todo_claim refusal is structured, CLI-identical text in the reason |
| AC-MN-002 (observer split; card gate 2) | PASS | `go test ./internal/hook -run '^TestFailureObserver_ExcludesPolicyRefusals$' -count=1` | `ok` — 2 refusal payloads + 1 genuine error → exactly 1 inbox row (the genuine Bash failure); usage-log gains 1 tool_failure for the same seam reason |
| AC-MN-004 (actor owned + stale-run) | PASS | `go test ./internal/cli -run '^TestMCPFactoryActor_OwnedAndStaleRun$' -count=1` | `ok` — actor lane-2 on a lane-1 card → `owned`, is_error false; actor lane-1 on a retired run → `stale-run`, is_error false; actor lane-1 on the active run proceeds |
| AC-MN-005 (positive operator mark) | PASS | `go test ./internal/cli -run '^TestMCPOperatorMark_Required$' -count=1` | `ok` — unmarked session refuses (unknown attribution, no closed-set code); lane session refuses with the CLI-identical boundary text; `MOAI_FACTORY_ROLE=operator` proceeds (factory_decide, todo_claim, todo_add) |
| AC-MN-003 (card gate 3) | NOT M1 | — | M3 flips it (plan §F M3 measurement harness) |

Verification batch at HEAD 6199704240 (this tree, this run):

- E2 builds: `go build ./...` → exit 0; `GOOS=windows GOARCH=amd64 go build ./...` → exit 0 (darwin/arm64 also verified).
- E3 coverage (change-scoped, per plan D8; full-matrix judgment belongs to CI): `mcp_refusal.go` 11 funcs average **94.2%** (toolRefusal 75% = the unreachable marshal-error branch; mcpRenewActorLabel 75%); `internal/hook` observer additions: recordToolFailureEvent 94.1%, mcpPolicyRefusalMark 100%, mcpRefusalInValue 90.9%, mcpRefusalInRawText 90.0%. Instrument: `go test -coverprofile` over the affected families.
- E4 boundary grep: `grep -rn "AskUserQuestion" internal/cli internal/hook` — the only hits are pre-existing documentation strings (harness.go "does not directly call", mcp_server.go result-shaping comment); zero in the M1 diff.
- E5 lint: `golangci-lint run` — internal/cli **0 issues** (isolated-cache run after waiting out a sibling session's lint lock), internal/hook + internal/config + internal/codexwiring + internal/factorymsg **0 issues**. No NEW findings, no baseline findings surfaced in the touched packages.
- Affected families green (-count=1): TestSD_AC010/014(×2)/015/016/024, TestFactoryMsg*/TestFactoryMCPIdentity*/TestFactoryLeadNotice*, TestFactoryNextNominateMCPParity, TestFactoryNominate/Quota/Complete sets, TestQWR_* (quota worktrees), TestTodoClaimMCP_Mirror, TestTodoAddMCPCarriesPresentation, `./internal/factorymsg` package, `./internal/hook` failure family, `./internal/config` + `./internal/codexwiring` packages.

Interpretation notes for the audit:

1. **Identity-class refusals carry no code** (REQ-MN-003's "refused with refused: true"): the closed set (plan D3) names no member for unknown-attribution / lane-boundary / not-a-lane, and inventing one is forbidden — the payload omits `code`. The five named classes carry their codes. EV-001's green prose says "refused:true+code"; the serial-slot instance (AC-MN-001's named class) carries `code: serial-slot`; the mirror arm asserts the structured shape with the code-omitting identity payload.
2. **Refusal tokens outside the closed set** (dropped, held, recorded, quota-hold, hold-marker, blocked, backend-skip, foreign-worktree, unknown-card) and the Codex merge-edge refusal keep today's error-channel shape — REQ-MN-001 names exactly five classes; broadening is a D3 SPEC amendment.
3. **Argument errors stay isError:true** per design D3 step 2 (missing project_root, blank card, malformed run id, legacy actor label) — the pinned wire shape.
4. **The observer exclusion covers both writes** (usage-log event + lessons-inbox stub) at the one seam: a policy refusal is not a malfunction for either consumer; AC-MN-002's metric (the inbox) is asserted directly.
5. **CLI/MCP gate divergence is sanctioned by plan D1**: factory_card.go is off-limits (parallel repair t1696), so the positive operator mark and the actor argument live at the MCP layer; the CLI verbs absorb after t1696 lands (lane board conditional task #5).
6. **Pre-existing load-sensitive flake classified**: TestFactoryLeaseQueueLockStallBounded failed once inside a 10-minute broad union run (timing assertion over by 19.7 ms; the union also hit go's 10m package timeout); isolated run and -count=3 pass. Not touched by this diff (CLI lease wait budget path).
7. **Existing behavior-pinning tests updated with the change** (the EV-018 pattern): the freeze snapshot (AC-SD-022) re-baselined with provenance, the MCP parity inputs pin gained `actor`, TestSD_AC016's unmarked success arm now stamps the operator mark, and the stage parity twin's refusal arm moved onto fresh fixtures (it previously passed by matching an illegal-transition error on both sides — the holder refusal it names requires a card still at plan).

### M2 — root contract, catalogue, schemas, wire honesty (card t1632, branch WT-3-2-0-2)

Commits (oldest first, after the leader's absorb merge `d1c7033429` = 8227a091b9 + main t1696/t1677): `8767d545b7` (**M2 Part A, CLI absorb**) — the M1 identity contract on the CLI verb layer: factoryDecideGate as the ONE predicate both decide surfaces run (positive operator mark admits / lane keeps the boundary refusal / unmarked is unknown attribution), `--actor` on next/stage/complete/renew with shared legacy-label validation, and an explicit run whose row exists but is not active refuses stale before any expired row is read (row-based: a run with no row has nothing expired to read). `decc922a6` (**M2 root confinement + factory_msg**) — validateProjectRoot confined to a registered worktree of the server's own repository (codexAuditValidateRoot acceptance set generalized), home rejected explicitly, non-string rejected as an invalid argument (the GetString coercion closed), factory_msg anchored at the caller's root with the four mutating members requiring it. `045733087` (**M2 body**) — REQ-MN-004 across the remaining families, REQ-MN-006 catalogue hard source, REQ-MN-007 schema lint, REQ-MN-008 wire honesty.

E8 — verbatim RED outputs captured BEFORE each GREEN (Part A + the M2 clusters):

```
$ go test ./internal/cli -run '^(TestFactoryDecideGate_PositiveMark|TestFactoryActorFlag|TestFactoryStaleRunRefused)$' -count=1   → exit 1
    factory_identity_absorb_test.go:40: cli decide accepted an unmarked session — env absence still grants operator
    factory_identity_absorb_test.go:103: cli stage --actor lane-2: err = unknown flag: --actor, want the F1 holder refusal naming lane-2
    factory_identity_absorb_test.go:142: cli stale-run refusal = factory stage: card evidence refused: artifact path "" is not usable, want the not-active refusal
$ go test ./internal/cli -run '^(TestMCPRootConfinement|TestMCPRootRequiredForMutating)$' -count=1   → exit 1
    mcp_root_contract_test.go:48: the home directory was accepted as a project root
    mcp_root_contract_test.go:62: sibling repository .../002 was accepted
    mcp_root_contract_test.go:81: project_root 42 was accepted            [GetString coercion]
    mcp_root_contract_test.go:118: mutating family tool without project_root was accepted (stale-run refused it first — the argument contract did not exist)
$ go test ./internal/cli -run 'TestMCPSchemaLint_ServedTools' -count=1   → exit 1
    schema lint findings (95): R1 additionalProperties not declared (sweep-wide) · R2 the 19 undescribed factory_msg_* properties · R4 seven read-only tools carrying the SDK default destructive hint (audit_cache, claude_audit, codex_job_result, codex_job_status, codex_review, codex_role_audit_result, ...)
$ grep -c "todo_claim" internal/mcp/catalog.go   → 0   [the catalogue row absent — the serve arm of TestMCPCatalogue_ServesTodoClaim is red by construction]
$ go test ./internal/cli -run 'TestFactoryMsgList_EmptyIsArrayAndTextMirrors' -count=1 (pre-fix probe)   → exit 1
    mcp_wire_honesty_test.go:66: empty claim set serialized as null: {"body_lookup":"factory_msg_body","messages":null}
```

AC matrix (M2-mapped rows; every row's command run at HEAD 045733087, this tree):

| AC | Status | Command (verbatim) | Observed output |
|----|--------|--------------------|-----------------|
| AC-MN-004/005 CLI half (Part A) | PASS | `go test ./internal/cli -run '^(TestFactoryDecideGate_PositiveMark\|TestFactoryActorFlag\|TestFactoryStaleRunRefused)$' -count=1` | `ok` — CLI decide flips to the positive mark, `--actor` decides whose label acts, a completed run refuses stale |
| AC-MN-005 MCP (factory_decide gate shared) | PASS | same batch incl. `TestMCPOperatorMark_Required` | `ok` — one predicate, both surfaces, identical texts |
| AC-MN-006 (resolver + mutating wire + _root) | PASS | `go test ./internal/cli -run '^TestMCPRootRequiredForMutating$' -count=1` | `ok` — absent project_root → isError true, text exactly `factory_msg_send: project_root is required for mutating calls`; correct calls carry `_root{dir, source:"param"}` (TestMCPRootConfinement's param arm) |
| AC-MN-007 (confinement) | PASS | `go test ./internal/cli -run '^TestMCPRootConfinement$' -count=1` | `ok` — home rejected explicitly, sibling repository rejected, registered worktree accepted (canonical), project_root 42 → invalid argument, never coerced |
| AC-MN-008 (todo_claim catalogue + fail-closed gate) | PASS | `go test ./internal/cli -run 'TestMCPCatalogue' -count=1` | `ok` — todo_claim in tools/list, HasTool pins, the gate panics on a catalogue-absent name and admits a row |
| AC-MN-009 (schema lint) | PASS | `go test ./internal/cli -run 'TestMCPSchemaLint' -count=1` | `ok` — R1 additionalProperties policy declared on every served tool (strict, applied at the registration gate), R2 every property described (the 19 factory_msg properties included), R3 required ⊆ declared (goal_status dropped the session_id required the enforcement never had), R4 read-only never destructive (the SDK default flipped on seven tools); handler-read args pinned (codex_task.model + project_root, goal_status.session_id) |
| AC-MN-010 M2 half (wire honesty) | PASS | `go test ./internal/cli -run '^TestFactoryMsgList_EmptyIsArrayAndTextMirrors$' -count=1` | `ok` — zero claims → `"messages":[]` (never null); the text block carries the same JSON as structuredContent |
| AC-MN-003 (card gate 3) | NOT M2 | — | M3 flips it |

Verification batch at HEAD 045733087 (this tree, this run):

- E2 builds: `go build ./...` → exit 0; `GOOS=windows GOARCH=amd64 go build ./...` → exit 0.
- E3 coverage (change-scoped, per plan D8): `mcp_project_root.go` M2 functions — resolveToolProjectRootWithSource 80.0%, toolProjectRootArgument 100.0%, resolveRequiredToolProjectRoot 77.8%, confineToolProjectRoot 80.0%, confineToServerRepository 80.0%. Instrument: `go test -coverprofile` over the refusal/root/catalogue families.
- E4 boundary grep: the only AskUserQuestion hits in internal/cli are pre-existing CLI documentation strings (harness.go, pr_watch_cmd.go — CLI surfaces, not the MCP handlers this diff touches); zero in the M2 diff.
- E5 lint: `golangci-lint run` over internal/cli, internal/mcp, internal/hook, internal/config, internal/codexwiring, internal/factorymsg → **0 issues**.
- Affected families green (-count=1): the root-contract sweep (confinement, required-root, catalogue, schema lint, registration guard), session_msg (4 tools + data-argument family), jev_ask, goal arms, session_list, audit_cache, codex_job_status/result/cancel, glm_task + glm_job_status/result/cancel, factory_msg 5, graph root honor family, verify_snapshot/trend root family, spec redirect family, factory CLI absorb tests.

Interpretation notes for the audit (M2):

1. **Part A ordering**: Part A (`8767d545b7`) landed BEFORE the M2 body per the delegation; the leader's absorb `d1c7033429` carried t1696, so internal/cli/factory_card.go edits happened after re-verification that the off-limits constraint was lifted.
2. **Row-based CLI stale-run**: the CLI refuses an explicit run when a run ROW exists and is not active; a run with no row passes through (nothing expired to read — the verb's own LoadCard answers). The MCP path keeps the stricter existence-based check (M1). The distinction keeps 40 fixture files' semantics intact and is documented on resolveFactoryCardRun.
3. **Confinement resolves the server root through projectDirResolver** (the same seam the codex/glm handlers swap): in-process fixtures substitute the server's repository there, so the confinement compares against the root the serving machinery actually resolves.
4. **verify_trend's no-root arm premise updated**: the store-root rules (SPEC-MCP-WORKTREE-UNTRACKED-001) keep ONE state store per repository — a same-repository fallback worktree resolves to the primary's store, so root-based record invisibility no longer holds; the guarantee the test pins moved to the `_root` provenance (fallback source + warning + the store source actually read). TestVerifySnapshot's isolation half still holds (verify.Load is identity-keyed).
5. **toolJSONWithRoot normalizes typed payloads to maps** (to merge `_root` at the top level); two tests asserting typed StructuredContent (factorymsg.Envelope, []jev.Answer) decode via a JSON round-trip now.
6. **session_msg broker anchoring moved**: newSessionMsgStoreFor(root) anchors the broker at the caller's confined root — the pre-M2 "takes NO project_root argument (design.md §6)" note is superseded by REQ-MN-004's family list.
7. **Pre-existing on the absorbed base, leader disposition**: TestFactoryLeaseQueueLockStallBounded and TestFactoryLeaseDriftLogVerbWorktreeWriteWaits exceed their pinned timing bounds without this diff (attribution run recorded in the Part A commit message).
8. **Scope note**: codex_role_audit* and codex_setup already carried their own root contracts (the codexAuditValidateRoot path this SPEC generalizes); audit_multi/codex_audit keep their pass-through resolution (resolveOptionalToolProjectRoot) per their REQ-2 semantics, now confinement-checked when a root is supplied.

### M3 — gate containment, elicitation, write mutex, inbox counting path (card t1632, branch WT-3-2-0-2)

Commits (after the M2-boundary absorb `d68d5172d` = eaac2057c + main t1707): `5e468bce1` (gate containment + elicitation guard + write mutex + inbox counting path + AC-MN-003 harness). Push state: nothing pushed; `git status -sb` clean on WT-3-2-0-2.

E8 — verbatim RED outputs captured BEFORE each GREEN:

```
$ go test ./internal/hook -run 'TestFactoryHookEmptyInboxSkipsWriteLock' -count=1   → exit 1
    factory_messages_lock_test.go:60: empty inbox under a held write lock: state = "degraded: database is locked (5) (SQLITE_BUSY)", want "empty-or-receipt-only" (the counting path must not take the write lock)
$ go test ./internal/cli -run 'TestMCPSchemaLint_ServedTools' (M2 tree, pre-exposure) — the gate-exposure and containment tests did not exist (EV-012's instruments are M3 builds; the disposition is regression-guard, not a plan-phase pass)
$ grep -rn "RequestElicitation\|WithElicitation" internal/ --include="*.go" (pre-guard, production files)   → 0 hits   [the invariant already held; the guard test pins it]
```

AC matrix (M3-mapped rows; every row's command run at HEAD 5e468bce1, this tree):

| AC | Status | Command (verbatim) | Observed output |
|----|--------|--------------------|-----------------|
| AC-MN-009 (gate exposure + containment) | PASS | `go test ./internal/cli -run 'TestMCPGateExposure\|TestCodexGateContainment' -count=1` | `ok` — the served set equals the enumerated 20-tool subset exactly (extra=[] missing=[]); audit_cache op=store refused naming the gate exclusion; containment args disable the foreign server, keep moai with the gate exposure override, and a listing failure refuses |
| AC-MN-010 (never-initiate + mutex) | PASS | `go test ./internal/cli -run 'TestMCPServer_NeverInitiatesElicitation\|TestMCPStateWriteMutex_ConcurrentWritersSerialize' -count=1 -race` | `ok` — zero RequestElicitation/WithElicitation references under internal/; 30 concurrent goal_arm/verify_snapshot/audit_cache writers serialize through the one mutex, -race clean |
| AC-MN-011 (inbox counting path) | PASS | `go test ./internal/hook -run 'TestFactoryHookEmptyInboxSkipsWriteLock' -count=1` | `ok` — with the broker write lock held by another connection, the empty inbox answers "empty-or-receipt-only" (the old path degraded with "database is locked (5) (SQLITE_BUSY)" — captured in the E8 block above) |
| AC-MN-012 + debt N5 (decline quantified) | PASS | `go test ./internal/cli -run 'TestManagedGateElicitationDeclineBounded' -count=1 -v` | `ok` — the automatic decline lands within 50 ms (measured well under), reply echoes the request id verbatim, result is {"action":"decline"}; "without stalling" is now a bounded-clock assertion, not prose |
| **AC-MN-003 (card gate 3)** | **PASS — measured** | `go test ./internal/cli -run 'TestACMN003_GateCallSuccessRate' -count=1 -v` | **`AC-MN-003 gate call success rate: 22/22 (100%) over 22 sampled calls`** — the fixture call set (initialize + tools/list + one call per enumerated subset tool) against the gate-exposure server with a gate-session-shaped in-process client; success = non-error, non-refused |
| AC-MN-013 (empty-inbox counting) | PASS | the same lock test + the delivery-path assertions in TestFactoryHookZeroTurnAndCapabilityTruth | `ok` — the counting path claims nothing, settles nothing, writes nothing; the delivery path claims exactly what it renders and settles receipts |

Verification batch at HEAD 5e468bce1 (this tree, this run):

- E2 builds: `go build ./...` → exit 0; `GOOS=windows GOARCH=amd64 go build ./...` → exit 0.
- E5 lint + format: gofmt clean on internal/cli, internal/hook, internal/factorymsg.
- Race lane: the guard + mutex tests green under `-race`.
- Affected families green (-count=1): the gate exposure/containment/harness set, the codex_audit + codex_audit_pin + codex_audit_mcp families (26 tests), the managed hardening family (the pre-existing elicitation-decline policy rows), the hook inbox family, the M2 root/schema/catalogue families.
- Catalogue doc parity: both moai-mcp-tools-catalogue.md copies updated to 49 tools with the todo_claim row (the registration-catalogue guard asserts the count).

Interpretation notes for the audit (M3):

1. **Gate exposure mechanism**: `MOAI_MCP_EXPOSURE=gate` / `moai mcp-server --exposure gate` intersects the enablement map with plan §F M3's enumerated table (moaiGateExposureTools). A new read-only tool must decide gate visibility in that table — the exposure test asserts served == enumerated, so an undecided tool fails the build's test run.
2. **Gate child containment is fail-closed**: the codex app-server spawn through the review-gate/audit path appends the foreign-server disable args + the moai gate-exposure args override, and a `codex mcp list` failure REFUSES the spawn (the role-audit precedent). The moai disable pair is dropped — the gate session is the moai server's client.
3. **codex_audit/review root validation is pre-confinement by scope**: the audit/delegation surface legitimately names ANOTHER repository's tree to audit, so the M2 server-repository confinement would have broken the audit's own fixture family (caught by this milestone's batch — an M2 regression the M3 batch surfaced and fixed via resolveValidatedToolProjectRoot). The wrong-tree hazard stays closed: unusable paths are rejected, never replaced.
4. **Receipt settling moved to the delivery path**: REQ-MN-011 forbids the counting path to write, so a receipt-only inspection leaves the receipt pending; the terminal acknowledged state arrives on the next non-empty (delivery) inspection or the factory_msg_receipt tool. TestFactoryHookZeroTurnAndCapabilityTruth's receipt arm was rewritten for that ordering.
5. **The exposure var is package-level** (moaiCurrentExposure): one server per process is the production reality; gateExposureClient restores the previous value so concurrent-family tests don't inherit "gate".
6. **AC-MN-003 denominator**: the harness samples initialize + tools/list + one call per enumerated subset tool (20) = 22 calls, matching acceptance.md §C's definition ("the denominator is the fixture call set — initialize, tools/list, and one call per tool in the REQ-MN-009 read-only subset"). The number is reported in this section with the explicit sample count the card gate requires; the plan-audit re-reads it from here.

## §E.3 Run-phase Audit-Ready Signal

_pending run-phase_

## §E.4 Sync-phase Audit-Ready Signal

_pending sync-phase_

## §F Phase 4 Mode Selection

Input parameters:
- tier: L (25 REQ / 25 AC, 6 milestones M1-M6)
- scope (file count): ~15-25 Go source + test files across internal/cli (mcp_*), internal/hook, internal/factorymsg, internal/auditreceipt, internal/codexwiring
- domain count: 3+ (MCP server layer, hook observer, messaging broker, audit receipts) — but coding-heavy
- file language mix: ~100% Go (+ markdown SPEC artifacts)
- concurrency benefit: LOW (coding-heavy — sequential milestone dependencies; M1 refusal contract is consumed by M2-M6)
- Agent Teams prereqs: not applicable (no operator request)

Mode evaluation:
- direct: not selected — semantic multi-file implementation, not trivial
- fanout: not selected — coding-heavy work violates Anthropic's coding-task parallelism caveat; milestones share the refusal contract and have inter-file dependencies
- sweep: not selected — semantic new-code work, not mechanical-uniform transformation
- serial: **selected** — one manager-develop spawn per milestone, sequential M1→M6

Decision: serial

Justification: the milestone chain is a layered contract build — M1's structured-refusal + actor-argument result shape is consumed by every later milestone (M2 project_root, M3 gate sessions, M6 audit integrity). Parallel spawns would each need the finished M1 contract and would contend for the same files (mcp_server.go dispatch tables). Anthropic's guidance: coding tasks involve fewer truly parallelizable tasks than research; serial is the safe default. Kickoff gate met in autonomous form (audit-cross: plan-audit-r3 PASS-WITH-DEBT 0.91 + codex pass rcpt-7323ca7bef2c885633529458; decision record .moai/reports/t1632/kickoff-decision.md; artifacts unchanged since verdict).
