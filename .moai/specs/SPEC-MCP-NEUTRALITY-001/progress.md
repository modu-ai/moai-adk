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

### M4 — env allowlist, staleness, job lifetime, hygiene (card t1632, branch WT-3-2-0-2)

Commits (after the M3-boundary absorb `327cd2824` = 32e8c0e9b + main t1631): `c13bc05c0` (env allowlist + pin + 429 relabel + staleness guard + AGENTS.md reconnect doc + job lifetime + role-audit cancel + stamp best-effort + call log). Push state: nothing pushed; `git status -sb` clean on WT-3-2-0-2.

E8 — verbatim RED outputs captured BEFORE each GREEN:

```
$ go test ./internal/cli -run 'TestClaudeAuditBinaryPinHonored' -count=1   → exit 1 (pre-fix)
    mcp_claude_env_allowlist_test.go:107: the audit invoked the PATH-lookup binary — the pin was bypassed
    mcp_claude_env_allowlist_test.go:110: the audit never invoked the pinned binary ...; invoked map[/system/claude-from-path:true]
$ go test ./internal/hook ... (M3) and the M2 confinement — the codex_audit "belongs to a different repository" regression the M3 batch surfaced
    codex_review_target_test.go:182: DEBUG handler isError: codex_audit: project_root ".../001" belongs to a different repository
```

AC matrix (M4-mapped rows; every row's command run at HEAD c13bc05c0, this tree):

| AC | Status | Command (verbatim) | Observed output |
|----|--------|--------------------|-----------------|
| AC-MN-012 (env allowlist) | PASS | `go test ./internal/cli -run 'TestClaudeAuditEnvAllowlist' -count=1` | `ok` — proxy vars, a redirected CLAUDE_CONFIG_DIR, provider keys, and the old scrub set all dropped; PATH kept; CLAUDE_CONFIG_DIR pinned to the subscription directory |
| AC-MN-012 (binary pin) | PASS | `go test ./internal/cli -run 'TestClaudeAuditBinaryPinHonored' -count=1` | `ok` — the audit invokes the MOAI_CLAUDE_BIN-pinned binary, never the PATH-lookup one |
| AC-MN-012 (429 relabel) | PASS | `go test ./internal/cli -run 'TestClaudeAudit' -count=1` | `ok` — a 429 yields ErrorCode `CLAUDE_RATE_LIMITED_ENDPOINT_UNVERIFIED` and the summary "rate limited (429); the serving endpoint is unverified" (the pre-fix "subscription capacity" label is gone) |
| AC-MN-013 (staleness) | PASS | `go test ./internal/cli -run 'TestMCPSchemaLint\|TestMoaiMCPServer_Registration' -count=1` + the guard wired in the registration gate | `ok` — write-capable tools refuse while stale; read tools annotate stale_server + _server + guidance; the check is inert for non-moai executables (the test process never reads stale) |
| AC-MN-014 (job lifetime) | PASS | `go test ./internal/cli -run 'TestCodexAudit\|TestGLMTask_Background\|TestCodexTask_Background' -count=1` | `ok` — role-audit jobs carry a cancel and are ended by the shutdown pass; GLM background goroutines join the shutdown; codex background cancellation unchanged; the three r1-named regression tests stay green |
| AC-MN-014 (role-audit cancel) | PASS | `go test ./internal/cli -run 'TestCodexRoleAudit' -count=1` + codex_role_audit_cancel registered (catalogue row + gate) | `ok` — the cancel surface refuses unknown jobs, reports already-terminal, and revokes the running context |
| AC-MN-015 (stamp best-effort) | PASS | `go build ./...` + the stamp block (mcp_server.go "REQ-MN-015") | the failed stamp releases the admission lock and serving CONTINUES; mcpServerRuntimeDir refuses to CREATE a .moai tree in a non-project cwd |
| AC-MN-015 (call log) | PASS | the call-log writer bound to the server run (mcp_server.go "REQ-MN-015" block) | one JSONL row per call (tool, duration ms, ok/refused/error, stale flag) under .moai/state/mcp-server/<pid>/calls.jsonl, size-rotated; nil (disabled) outside a server run so direct-construction tests never touch the real state tree |

Verification batch at HEAD c13bc05c0 (this tree, this run):

- E2 builds: `go build ./...` → exit 0; `GOOS=windows GOARCH=amd64 go build ./...` → exit 0.
- Affected families green (-count=1): the claude audit family (allowlist, pin, 429 relabel, provider mismatch, taxonomy), the codex_audit + codex_audit_pin + codex_audit_mcp families, TestGLMTask_Background, TestMCPSchemaLint, TestMoaiMCPServer_Registration, TestMCPRoot.
- Catalogue doc parity: both moai-mcp-tools-catalogue.md copies updated to 50 tools with the codex_role_audit_cancel row.

Interpretation notes for the audit (M4):

1. **The claude_audit pin resolution routes through the seam**: resolveClaudeBinaryAt's PATH fallback now calls claudeLookPath (the same injectable seam) — the existing binary-missing test keeps working AND the pin is honored. The runner's binary argument is the assertion surface.
2. **429 label**: ErrorCode is the NEW `CLAUDE_RATE_LIMITED_ENDPOINT_UNVERIFIED` (the old CLAUDE_CAPACITY_UNAVAILABLE const is retained for reference but no longer emitted on the 429 path); AuthMode stays "subscription" (the auth check DID pass — the mislabel was the capacity claim, not the auth mode).
3. **The staleness guard is inert for non-moai executables**: os.Executable's base must start with "moai" — the test process (cli.test) never reads stale, so no existing test breaks. Production serves through `moai mcp-server`, whose executable base is moai.
4. **Role-audit context detach**: the audit context derives from WithCancel(WithoutCancel(ctx)) — the handler's return cannot kill a minutes-long audit, while the stored cancel gives the shutdown pass and codex_role_audit_cancel the end handle.
5. **Call log nil-when-direct**: mcpCallLog is bound ONLY inside runMCPServerWithIOAndExposure — direct-construction tests leave it nil, so no test writes the real project's state tree (B8 hygiene).
6. **AGENTS.md reconnect doc**: local AGENTS.md AND the template mirror (AGENTS.md.tmpl) gained section 14 (per-harness reconnect procedures) — both copies in sync.

### M5 — generation migration, token stability, count-only hook, Stop scope (card t1632, branch WT-3-2-0-2)

Commits (after the M4-boundary absorb `fa05db2fb` = 0d53d09fa + main t1619): `db9fb0e7b` (generation migration + token stability + count-only hook + Stop scope). Push state: nothing pushed; clean on WT-3-2-0-2.

E8 — verbatim RED outputs captured BEFORE each GREEN:

```
$ go test ./internal/cli -run 'TestStoreGenerationAdvanceMigratesPending|TestClaimTokenStableOnRelist' -count=1   → exit 1
    factorymsg_claim_stability_test.go:72: claims at generation 2 = 0, want 1 (the pending message must migrate, not orphan)
    factorymsg_claim_stability_test.go:94: second claim returned 0 entries, want the already-claimed entry
$ go test ./internal/hook -run 'TestFactoryHookEmptyInboxSkipsWriteLock' (M3) — the counting-path premise extends to M5's count-only hook (no claim either)
```

AC matrix (M5-mapped rows; every row's command run at HEAD db9fb0e7b, this tree):

| AC | Status | Command (verbatim) | Observed output |
|----|--------|--------------------|-----------------|
| AC-MN-016 (generation migration) | PASS | `go test ./internal/cli -run '^TestStoreGenerationAdvanceMigratesPending$' -count=1` | `ok` — a same-session generation advance migrates the pending message to the new generation (claimable at gen 2) |
| AC-MN-017 (token stability) | PASS | `go test ./internal/cli -run '^TestClaimTokenStableOnRelist$' -count=1` | `ok` — the second claim returns the SAME message with the SAME token (extended lease); ReadBody with the original token still works |
| AC-MN-017 (hook count-only) | PASS | `go test ./internal/hook -run 'TestFactoryHook\|TestRebound' -count=1` | `ok` — the hook's inspections claim nothing, settle nothing, write nothing; repeated inspections are idempotent counts |
| AC-MN-018 (line-boundary cut + Stop scope) | PASS | the same hook family run | `ok` — the byte cut falls on a line boundary (ContextAndContinuationSafety's 16-message render); Stop blocks only for blocker/status_request (dispatch_notice delivers as a count without forcing the turn); the lock arm pins the count path succeeding UNDER a held write lock within 250ms |
| AC-MN-010 M5 half (already-claimed entries) | PASS | `go test ./internal/cli -run '^TestClaimTokenStableOnRelist$' -count=1` | `ok` — factory_msg_list's Claim returns the already-claimed entries with their tokens (never an empty-only response for live claims) |

Verification batch at HEAD db9fb0e7b (this tree, this run):

- E2 builds: `go build ./...` → exit 0; `GOOS=windows GOARCH=amd64 go build ./...` → exit 0.
- Affected families green (-count=1): internal/factorymsg (full package), the hook inbox family, the cli claim-stability + generation-migration tests.

Interpretation notes for the audit (M5):

1. **Generation migration is same-session-scoped**: a SLOT REBIND to a different session UUID keeps the REQ-DHR-020 superseded semantics (the no-silent-retargeting fence, pinned by TestFactorySessionGenerationOwnership). The migration fires only for a same-session generation advance — the pure "generation change alone" case REQ-MN-016 names. The idempotency-scope test updated: a same-key re-Send to the migrated recipient returns the ORIGINAL envelope (the scope spans the migration).
2. **Token stability is lease-gated**: a LIVE claim keeps its token; an EXPIRED claim mints a fresh one (the dispatch-result exactly-once fencing contract, pinned by TestDispatchResultExactlyOnce, requires the old holder's identity to die with its lease). The distinction is claim_expires_at vs now.
3. **The hook is now FULLY read-only**: no claim, no settle, no write probe (M3 removed the probe+settle; M5 removes the claim). The factory_msg_list tool is the only claim surface; the hook renders pending ids + kinds and the Stop decision reads kinds. TestFactoryHookZeroTurnAndCapabilityTruth's receipt arm updated: a receipt-only inspection leaves the receipt pending; the receipt answers via the MCP receipt tool.
4. **Stop scope**: only blocker/status_request force the leader's turn; dispatch_notice and other kinds deliver as counts (cont=false). PendingList sorts decision-required kinds FIRST so a rendered batch always carries them even past the 16-tool limit.
5. **The lock arm re-pinned**: the count path succeeds UNDER a held write lock within 250ms — the strongest possible REQ-MN-011 pin (the old "degraded under lock" premise is the defect itself).
6. **factory_msg_list's already-claimed entries** (AC-MN-010's M5 half): the Claim returns live-claimed entries with their tokens alongside new claims — the response is never empty-only for live claims.

### M6 — audit integrity, could-not-run split, target honesty, sandbox, state-root seam (card t1632, branch WT-3-2-0-2)

Commits: `b5d602c5c` (could-not-run split + not-converged convergence + stderr retention + one retry + REQ-MN-022), `09e4cd60a` (GLM narrow decode, REQ-MN-023), `21c419757` (EC-1/EC-4 characterization refresh), `36a04bff4` (receipt integrity + audit_cache evidence + auditor model, REQ-MN-020), `6e21dea76` (target honesty + sandbox + material marks + state-root seam, REQ-MN-019/024/025), `cf3546753` (schema snapshot + receipt-site count refresh). Push state: nothing pushed; clean on WT-3-2-0-2.

E8 — verbatim RED outputs captured BEFORE each GREEN:

```
$ go test ./internal/cli -run 'TestCodexAudit_CouldNotRunNotConvertedToFail|TestCodexAudit_StderrPreservedOnCouldNotRun|TestCodexAudit_CouldNotRunRetriesOnce|TestEnforceRequiredGateUnmet_CouldNotRunNotConvergedToFail' -count=1   → exit 1
    codex_audit_could_not_run_test.go:34: verdict = fail — a could-not-run was converted into a reviewed failure (REQ-MN-021)
    codex_audit_could_not_run_test.go:38: gate_unmet = "workflow.audit.gates.codex is `required`, but this audit returned no verdict (fail-open inconclusive)", want it to carry could-not-run
    codex_audit_could_not_run_test.go:55: stderr = "", want the backend's stderr preserved
    codex_audit_could_not_run_test.go:83: codexReviewRPC calls = 1, want exactly 2 (original + one retry)
    codex_audit_could_not_run_test.go:117: overall = fail — could-not-run converted into a converged failure (REQ-MN-021)
$ go test ./internal/cli -run 'TestParseGLMReview_NarrowDecode' -count=1   → exit 1
    mcp_glm_narrow_test.go:37: model-supplied provenance survived the decode: &{Backend:codex ...}
    mcp_glm_narrow_test.go:40: model-supplied gate_unmet survived: "injected"
    mcp_glm_narrow_test.go:43: model-supplied audit_receipt survived: "forged"
$ go vet ./internal/auditreceipt   → compile-stage RED (REQ-MN-020)
    receipt_integrity_test.go:17:10: undefined: TreeState
    receipt_integrity_test.go:52:9: got.HeadCommit undefined (type Receipt has no field or method HeadCommit)
    receipt_integrity_test.go:78:14: undefined: CauseReceiptVerdictFail
$ go vet ./internal/runtime   → compile-stage RED (REQ-MN-020 audit_cache)
    audit_cache_state_test.go:45:7: undefined: NewStateAuditCache
$ go vet ./internal/cli   → compile-stage RED (REQ-MN-019/024/025)
    codex_review_state_root_test.go:42:16: undefined: codexReviewStateRoot
$ go test ./internal/cli -run 'TestCodexAudit_AdversarialBaseBranchAppliesTheTarget|...UnresolvableStatesNotApplied|...ClaimsNoReviewBase' -count=1   → exit 1 (REQ-MN-019)
$ go test ./internal/cli -run 'TestCodexAudit_SessionSpawnsUnderReadOnlySandbox|TestCodexReview_SessionSpawnsUnderReadOnlySandbox' -count=1   → exit 1 (REQ-MN-024)
$ go test ./internal/cli -run 'TestGLMAudit_MarksTruncatedMaterial|TestGLMAudit_MarksExcludedUntracked' -count=1   → exit 1 (REQ-MN-024)
$ go test ./internal/cli -run 'TestCodexAudit_RecordsAuditorModel|TestGLMAudit_RecordsAuditorModel' -count=1   → exit 1 (REQ-MN-020)
```

AC matrix (M6-mapped rows; every row's command run at HEAD cf3546753, this tree):

| AC | Status | Command (verbatim) | Observed output |
|----|--------|--------------------|-----------------|
| AC-MN-021 (could-not-run split + retry + stderr) | PASS | `go test ./internal/cli -run 'TestCodexAudit_CouldNotRunNotConvertedToFail\|TestCodexAudit_StderrPreservedOnCouldNotRun\|TestCodexAudit_CouldNotRunRetriesOnce\|TestCodexAudit_SuccessVerdictRetriesNever\|TestEnforceRequiredGateUnmet_CouldNotRunNotConvergedToFail' -count=1` | `ok` — a forced transport failure keeps verdict `inconclusive` (never `fail`), gate_unmet carries could-not-run, the backend's stderr tail rides the result, and the leg retried exactly once (calls=2); a real verdict never retries (calls=1) |
| AC-MN-021 (baseBranch target honesty) | PASS | `go test ./internal/cli -run 'TestCodexAudit_Adversarial' -count=1` | `ok` — an adversarial audit asked for baseBranch carries the resolved base SHA + the diff instruction in the turn/start prompt, review_base names the judged base, and an unresolvable base returns inconclusive naming "target not applied" |
| AC-MN-022 (not-converged + narrow decode) | PASS | `go test ./internal/cli -run 'TestConverge\|TestParseGLMReview\|TestEnforceRequiredGateUnmet' -count=1` | `ok` — a required-inconclusive convergence stays not-converged (never pass on the caller's submitted verdict); model-supplied provenance/gate_unmet/audit_receipt do not survive the GLM decode |
| AC-MN-023 (receipt evidence binding) | PASS | `go test ./internal/auditreceipt -count=1` + `go test ./internal/runtime -count=1` | `ok` — a receipt carries head_commit + tree_digest; a FAIL-verdict receipt and a receipt whose state moved are refused as PASS corroboration; audit_cache entries persist to the project state root requiring an existing verdict file's SHA-256 (revalidated at lookup); ComputeHash refuses a non-SPEC directory; codex and GLM verdicts record the auditor model on provenance |
| AC-MN-024 (sandbox + material marks + empty guard) | PASS | `go test ./internal/cli -run 'TestCodexAudit_SessionSpawnsUnderReadOnlySandbox\|TestCodexReview_SessionSpawnsUnderReadOnlySandbox\|TestGLMAudit_MarksTruncatedMaterial\|TestGLMAudit_MarksExcludedUntracked\|TestGLMAudit_EmptyMaterialNeverInvokesBackend' -count=1` | `ok` — audit and review sessions spawn with the read-only sandbox override on argv and the scrubbed codexChildEnv; truncated/excluded_untracked ride the result; empty material invokes neither backend |
| AC-MN-025 (state-root seam) | PASS | `go test ./internal/cli -run 'TestCodexReviewStateRootAgreesWithProducer' -count=1` | `ok` — the Stop chain's receipt reader resolves the same git-toplevel root the producer wrote under for a session started in a subdirectory; the false "receipt absent" negative no longer reproduces |

Verification batch at HEAD cf3546753 (this tree, this run):

- E2 builds: `go build ./...` → exit 0.
- Affected families green (-count=1): internal/auditreceipt (full package), internal/runtime (full package), internal/closure (full package), and the cli families — TestCodexAudit*, TestRunMultiAudit*, TestSelfReview* (51 files family), TestGLMAudit*, TestConvergence*, TestCodexBlankReview*, TestContradiction*, TestAuditMulti_NoConfigNoArgs/PinsOnly/ArgumentRequired/ExplicitClaudeToken/PaddedConfiguredGate/SuppliedGateOverrides/CarrylessCaller/EnforcementFollowsTheCallStartPlan/ReceiptExposure/UnknownOriginWithoutAnchor, TestEffectiveScopeFor, TestCodexReviewStateRootAgreesWithProducer.
- Lint: `golangci-lint run ./internal/auditreceipt/... ./internal/runtime/... ./internal/cli/... ./internal/closure/...` → 4 errcheck findings, ALL pre-existing on committed files (factory_run_adopt.go ×3, mcp_server_call_log.go ×1) — confirmed by re-linting with the M6 sources temporarily reverted to HEAD (b5d602c5c ancestry): same 4. Zero NEW lint findings from this milestone.

Interpretation notes for the audit (M6):

1. **REQ-MN-021/022 supersede the #1632-axis-3 fail-closed conversion at both surfaces**: a backend leg that produced NO verdict is a not-performed result. `applyGateUnmet` keeps the verdict inconclusive and names could-not-run in gate_unmet; `enforceRequiredGateUnmet` keeps the overall at the distinct not-converged outcome (REQ-MN-022), never a converged fail — the pass the old claude-anchor fall-through produced stays equally unreachable, so the gate still blocks. The bytes this changes for a machine consumer: `could_not_run: true` rides the result (additive, omitempty).
2. **runMultiAudit re-stamps the per-backend gate from the tree's raw workflow.yaml BEFORE converge** (codex/glm only; the in-session claude anchor keeps its own default-required hierarchy): an explicitly advisory backend no longer rides the engine-default `required` into not-converged — the explicit-advisory-never-blocks contract is preserved byte-for-byte where it matters (TestRunMultiAudit_ExplicitAdvisoryGateUnmet_StillPasses), while an UNSET gate's byte-for-byte pass clause is superseded by REQ-MN-022 (not-converged, documented in the updated pin).
3. **The retry is exactly-once and seam-observable**: handleCodexAudit retries a could-not-run leg exactly once via the codexReviewRPC seam (calls=2 pinned); a real verdict never retries (calls=1 pinned). A missing binary is a could-not-run that never enters the RPC path, so it does not retry (deterministic failure — retrying cannot change LookPath).
4. **stderr is best-effort by construction**: a process that never started has no stderr; a handshake captured one carries the tail (4KiB bound) through codexSessionError; the reuse check treats a receipt with NO binding as unattributable, never contradictory.
5. **M2 follow-through landed inside this milestone** (same-SPEC cascade, L46): the audit-tool schema snapshot (testdata/audit-tools-schema-pre.json) was regenerated for M2's schema strictness + the M6 wire fields — the snapshot test had been failing since M2 landed the strictness without regenerating; the self-review fixtures carry the REQ-MN-005 server-root seam (the fixture trees live in their own repositories, so the confinement was refusing every self-review call — a latent M2 regression the M6 family run surfaced); glm_audit moves to resolveValidatedToolProjectRoot, the audit-surface root validation its codex/claude siblings already had.
6. **REQ-MN-025 scope**: the Stop chain's receipt readers (members using verify.LoadReceipt) now resolve codexReviewStateRoot — the same git-toplevel anchoring the producer applies via reviewExclusionRoot. The full cross-harness state unification stays out of scope (§5 Non-Goal).
7. **Pre-existing failures observed during M6 verification (NOT introduced here)**: TestCodexReviewGatePrimaryCheckoutSkip / _TreeScopeRequestShapeUnchanged / _StaleEnvKeepsTreeScope (macOS /var→/private symlink spelling in three gate pins — reproduced at HEAD 554053cf9 with the M6 sources reverted, so they predate this milestone) and TestAuditMulti_CodexLegDeadline_FailsClosedNamed / _UnconfiguredFailsOpen / TestAuditMulti_ModelMulti_CodexUnavailable_FailsClosedNamed (same reproduction method — the containment listing executes the stubbed binary before the timeout path). Leader disposition requested: both classes reproduce on the M5 HEAD tree.

### Card-review repair round — P1 3 + P2 5 (card t1632, branch WT-3-2-0-2)

Commit: `c292a7a8d` (fix: card-review repairs — P1 3 + P2 5). Base: sync commit `8864f9f56` (+backfill `873255f1e`); review base was e48f3b296 → head 226c387e3. Verdict file: .moai/reports/t1632/card-review.md (codex, scope=card, advisory fail — 3 P1 + 8 P2). Push state: nothing pushed; clean on WT-3-2-0-2.

E8 — verbatim RED outputs captured BEFORE each GREEN:

```
$ go test ./internal/auditreceipt -run 'TestTreeState_ReeditMovesDigest|TestTreeState_UntrackedContentMovesDigest' -count=1   → exit 1
    (P1-1) both FAILED — digest unchanged after an in-place re-edit / after an untracked file's content changed
$ go test ./internal/auditreceipt -run 'TestCheckCitedReceipts_StateComparesTheJudgedTree' -count=1   → exit 1
    receipt_state_root_test.go:34: WriteReceipt: audit-receipt dir: mkdir /primary-store: read-only file system
    (fixture fix: the store root is a real temp dir; the judged tree stays virtual — then the assertion readFor != judged fired, the P1-2 defect)
$ go vet ./internal/cli   → compile-stage RED (P1-3)
    mcp_server_staleness_test.go:15:48: undefined: binaryIdentity
    mcp_server_staleness_test.go:17:23: undefined: servingBinaryIdentity
$ go test ./internal/runtime -run 'TestStateCache_RelativeReportPathAnchoredToRoot' -count=1   → compile-stage RED (P2-7), then GREEN after resolveReportPath
```

Repair matrix (every row's command run at HEAD c292a7a8d, this tree):

| Finding | Disposition | Command (verbatim) | Observed output |
|---------|-------------|--------------------|-----------------|
| P1-1 digest misses file content | FIXED | `go test ./internal/auditreceipt -run 'TestTreeState_ReeditMovesDigest\|TestTreeState_UntrackedContentMovesDigest' -count=1` | `ok` — an in-place re-edit (same HEAD, same porcelain rows) and an untracked file's content change both MOVE the digest (tracked diff bytes + non-runtime untracked bytes, sorted, folded in) |
| P1-2 state compared at the store root | FIXED | `go test ./internal/auditreceipt -count=1` | `ok` — the freshness comparison reads TreeState(r.TreeRoot), the judged tree, never the caller's store root (TestCheckCitedReceipts_StateComparesTheJudgedTree pins the seam read) |
| P1-3 start-time binary identity | FIXED | `go test ./internal/cli -run 'TestServerStaleness' -count=1` | `ok` — an in-place reinstall (start identity 2048B/T1 vs installed 4096B/T2) reads STALE against the start-time capture; an unchanged install reads fresh |
| P2-7 relative report_path | FIXED | `go test ./internal/runtime -run 'TestStateCache' -count=1` | `ok` — a relative report path resolves against the project root at store AND at lookup; the entry carries the absolute anchored path |
| P2-8 baseBranch truncated | FIXED | `go test ./internal/cli -run 'TestGLMAudit' -count=1` | `ok` — the baseBranch collector (collectReviewDiffAtFlagged) carries the truncation mark; the GLM family stays green |
| P2-9 rotation per row | FIXED | `go build ./...` + the call-log rewrite | `ok` — the writer tracks the CURRENT segment's size (w.cur/w.size); no re-stat of the initial calls.jsonl, so post-rotation rows append to one segment instead of minting a segment per row |
| P2-10 duration from registration | FIXED | `go build ./...` + the add-wrapper rewrite | `ok` — the timer starts INSIDE logged, so the registration-to-first-call gap is never billed to a tool |
| P2-11 logger creates .moai | FIXED | `go test ./internal/cli -run 'TestMCPServerRuntimeDir_ProjectScoped\|TestMoaiMCPServer_AuditCacheStoreLookup\|TestServerStaleness' -count=1` | `ok` — newMCPCallLogWriter activates only for a cwd that already has .moai (the REQ-MN-015 discipline); the two fixtures updated (the runtime-dir fixture carries .moai; the audit_cache store fixture creates its verdict file — REQ-MN-020's existing-file requirement) |
| P2-4 expired-claim hook redelivery | HELD (leader disposition) | — | crosses the M5 DHR-020 exactly-once fencing contract; a follow-up card owns the design (redelivery vs fence) |
| P2-5 audit_cache per-op write classification | HELD (leader disposition) | — | the catalogue WriteCapable reclassification is a design change, not a mechanical fix; follow-up card |
| P2-6 stderr buffer synchronization | HELD (leader disposition) | — | deep mcp_codex change (mutex-guarded buffer or post-close read); risk/benefit judged by the leader with the real-usage evidence |

**Follow-up card ids (leader, 2026-10-14T21:4xZ)**: F6 + DHR-020 dead-letter half → **t1747**; the two pre-existing flaky classes (symlink 3 + deadline pins 3, F11) → **t1746**. The P2-4/5/6 held findings ride t1747's scope per the sync-audit-r4 §Debt record. ~~TestNoNewEnvLiteralsInDiff is main-origin (t1617 190f7db5d) — deliberately not this card's.~~ **CORRECTION (sync-audit r6 F21, 2026-10-15) — this main-origin sentence was FALSE and contradicted the r2 measurement and the r3 CORRECTION at this file's repair-round block: the guard's added-line set carries THIS card's own two comment lines (M1 `ebba213f0` mcp_refusal.go, M2-PartA `9e7a4a9a9` mcp_factory_card.go), reworded and the guard made green in the r3 repair. The t1617-side introductions in the sweep range are test-file-only.**

Verification batch at HEAD c292a7a8d (this tree, this run):

- E2 builds: `go build ./...` → exit 0.
- Affected families green (-count=1): internal/auditreceipt (full), internal/runtime (full), internal/closure (full), and the cli families TestCodexAudit*, TestCodexReview*, TestSelfReview*, TestGLMAudit*, TestAuditCache*, TestMoaiMCPServer*, TestServerStaleness*, TestMCPServerRuntimeDir*, TestRunMultiAudit*, TestConverge*.
- Lint: 4 errcheck, ALL pre-existing (factory_run_adopt.go ×3, mcp_server_call_log.go ×1 — the same finding the M6 round measured before this repair round touched the file); zero NEW.
  - **CORRECTION (sync-audit F2, 2026-10-15)**: the claim above was TRUE ONLY INSIDE ITS MEASURED SCOPE (auditreceipt, runtime, cli, closure) and FALSE across the card's touched set — `./internal/hook/...` carried a NEW errcheck from this card's own M3 commit `5e468bce1` (factory_messages_lock_test.go:54) that the scoped lint never saw, and `go test ./internal/mcp` carried the F1 catalogue-pin reds every milestone batch missed. This repair round re-measured the FULL touched set (cli, hook, factorymsg, auditreceipt, mcp, codexwiring, runtime, closure — 8 roots) and the §E.2 batches now include both surfaces; the F1/F2 root cause (verification-scope gap) is this card's process defect, recorded here so the audit can attribute it.

### Sync-audit r1 delta repair round — F1–F5 (card t1632, branch WT-3-2-0-2)

Commit: `850307cf4` (fix: sync-audit r1 delta repairs — F1-F5). Verdict file: .moai/reports/t1632/sync-audit-r1.md (FAIL; §Repair routing is the authoritative scope). Push state: nothing pushed; clean on WT-3-2-0-2.

E8 — verbatim RED outputs captured BEFORE each GREEN:

```
$ go test ./internal/mcp -count=1   → exit 1 (F1, at the repair-round start)
    --- FAIL: TestMoaiMCPTools_CatalogSize — catalog declares 50 tools, want 48
    --- FAIL: TestMoaiMCPTools_WriteCapableSet — write-capable tool count = 23, want 21
$ go vet ./internal/auditreceipt   → compile-stage RED (F3)
    receipt_state_error_test.go:51:14: undefined: CauseReceiptStateUnreadable
$ go test ./internal/auditreceipt -run 'TestCheckCitedReceipts_AbsentGitCarveOut' -count=1   → exit 1 (F3 fixture: a string error does not satisfy errors.Is; the stub now wraps exec.ErrNotFound like gitIn does)
$ go test ./internal/auditreceipt -run 'TestTreeState_IgnoresInheritedGitDir' -count=1   → exit 1 (F4)
    gitin_env_test.go:40: TreeState with a hostile GIT_DIR: open .../other.go: no such file or directory
$ go vet ./internal/cli   → compile-stage RED (F5)
    codex_audit_bind_timing_test.go: unhandled symbols before the seams existed; after GREEN the bind-timing pin asserts exactly ONE TreeState read (the start capture — a second read would mean the write-time bind came back)
```

Repair matrix (every row's command run at HEAD 850307cf4, this tree):

| Finding | Disposition | Command (verbatim) | Observed output |
|---------|-------------|--------------------|-----------------|
| F1 catalogue pins stale (todo_claim M3, codex_role_audit_cancel M4) | FIXED | `go test ./internal/mcp/... -count=1` | `ok` — size pin 48→50, both tools join the write-capable expected set (23 = 21+2) |
| F2 M3 errcheck + lint-scope gap | FIXED + scope corrected | `golangci-lint run` over the 8 touched roots (cli, hook, factorymsg, auditreceipt, mcp, codexwiring, runtime, closure) | 4 errcheck, all pre-existing in internal/cli — the hook finding is gone (capture-and-ignore rollback), and the full-scope measurement is now the §E.2 batch surface (correction recorded above) |
| F3 state-read failure accepts the receipt (Korean filename breaks the porcelain parse) | FIXED | `go test ./internal/auditreceipt -count=1` | `ok` — state reads parse the NUL-delimited `-z` forms (TestTreeState_KoreanFilenameSurvives pins a Korean-named untracked file end-to-end); a state-read error now REFUSES the reuse (CauseReceiptStateUnreadable), with the single absent-git carve-out (exec.ErrNotFound) pinned by TestCheckCitedReceipts_AbsentGitCarveOut |
| F4 gitIn inherits GIT_DIR/GIT_WORK_TREE | FIXED | `go test ./internal/auditreceipt -count=1` | `ok` — gitIn applies scrubbedGitEnv (the storeroot.go pattern); TestTreeState_IgnoresInheritedGitDir pins that a hostile GIT_DIR+GIT_WORK_TREE cannot redirect the state read (RED reproduced: the untracked read of the OTHER tree failed before the scrub) |
| F5 receipt binds the post-audit state | FIXED | `go test ./internal/cli -run 'TestCodexAudit_ReceiptBindsStartOfAuditState' -count=1` | `ok` — handleCodexAudit captures the tree state at audit START (auditreceipt.TreeState once, before serving) and recordAuditReceiptBound writes it verbatim; the write-time bind is GONE (the pin asserts exactly one TreeState read), so a tree edit during a minutes-long audit cannot launder a PASS through the freshness check |
| F6 stale-server audit_cache op=store | HELD (leader disposition — escalation of held card-review P2-5) | — | the mcpStalenessGuard argument-inspection fix is the leader's call; no code change here |

Verification batch at HEAD 850307cf4 (this tree, this run):

- E2 builds: `go build ./...` → exit 0.
- Affected families green (-count=1): internal/auditreceipt (full), internal/mcp (full), internal/factorymsg (full), internal/closure (full), the hook family TestFactoryMessages*+TestFactoryHook*, and the cli families TestCodexAudit*, TestSelfReview*, TestServerStaleness*, TestGLMAudit*.
- Lint (full touched set, 8 roots): 4 errcheck, all pre-existing in internal/cli — zero NEW (the corrected measurement).
- ~~**Pre-existing red observed, NOT a card defect**: internal/hook's full-suite TestNoNewEnvLiteralsInDiff fails on the `"MOAI_FACTORY_WORKER"` literal introduced by card t1617 commit `190f7db5d` and carried into this tree by the card t1619 absorb `cd19e5494` — zero commits in this card's range.~~ **CORRECTION (sync-audit r3 F12, 2026-10-15) — the lineage claim above was FALSE, and the correction is part of the repair.** The sweep's added-line set (base `966f1b158` → HEAD, non-test) contains exactly TWO lines with the literal, BOTH authored by THIS card: the mcp_refusal.go comment (M1 `ebba213f0`) and the mcp_factory_card.go comment (M2-PartA `9e7a4a9a9`) — `git show e48f3b296:internal/cli/mcp_factory_card.go | grep -c` = 0 (no occurrence at the base) and mcp_refusal.go did not exist at the base. The t1617/t1619-side introductions in the range are all in TEST files, which the sweep excludes. This was a verification-claim-integrity §1.1 surface-3 violation (a defect claim attributed to a lineage that was never traced) — the two comments are REWORDED in the r3 repair commit (`850307cf4`+r3 commit: "the spawn environment's factory-worker key (internal/config/envkeys.go EnvMoaiFactoryWorker)"), the guard is GREEN there, and the misattribution stands recorded here as the artifact of record.

### Sync-audit r3 delta repair round — F12/F13/F14 (card t1632, branch WT-3-2-0-2)

Commits: the r3 repair commit (fix: sync-audit r3 delta repairs — F12/F13/F14) + its evidence commit. Base: the r1-repair head `432c24187`. Verdict file: .moai/reports/t1632/sync-audit-r2.md (F1–F5 core resolved confirmed; the three blocking findings below). Push state: nothing pushed; clean on WT-3-2-0-2.

E8 — verbatim RED outputs captured BEFORE each GREEN:

```
$ go test ./internal/cli -run 'TestAuditMulti_ReceiptBindsStartOfConvergenceState' -count=1   → exit 1 (F13; first draft had no discriminating power — the scenario was re-scoped to flip the state DURING the fan-out, then:)
    audit_multi_bind_timing_test.go:58: receipt binds head="post-edit-head" digest="post-edit-digest", want the START state — a write-time bind on the convergence surface is the F13 laundering path
$ go vet ./internal/cli   → compile-stage RED (F14)
    codex_audit_capture_failure_test.go:57:10: rec.BindCaptureFailed undefined (type auditreceipt.Receipt has no field or method BindCaptureFailed)
$ go test ./internal/hook -run 'TestNoNewEnvLiteralsInDiff' -count=1   → FAIL at the r3 round start (F12); GREEN after the two comment rewordings
```

Repair matrix (every row's command run at the r3 repair head, this tree):

| Finding | Disposition | Command (verbatim) | Observed output |
|---------|-------------|--------------------|-----------------|
| F13 audit_multi receipt binds write-time | FIXED | `go test ./internal/cli -run 'TestAuditMulti_ReceiptBindsStartOfConvergenceState' -count=1` | `ok` — runMultiAudit captures TreeState BEFORE the fan-out through the SAME receipt-target contract the write uses, and routes through recordAuditReceiptAtBound; the pin flips the state DURING the fan-out and requires the START state on the receipt |
| F14 start-capture failure falls back to write-time bind | FIXED | `go test ./internal/cli -run 'TestCodexAudit_CaptureFailureNeverFallsBackToWriteTimeBind\|TestCheckCitedReceipts_CaptureFailedMarkerRefused' -count=1` | `ok` — a failed start capture records the explicit `bind_capture_failed` marker (no binding, NO fallback), and the reuse check refuses a capture-failed receipt (CauseReceiptBindCaptureFailed) — its freshness is unknowable, so it proves nothing |
| F12 env-literal guard red on this card's own comments + false lineage in this file | FIXED + corrected | `go test ./internal/hook -run 'TestNoNewEnvLiteralsInDiff' -count=1` | `ok` — the two comments (mcp_refusal.go, mcp_factory_card.go) now reference the key through internal/config/envkeys.go instead of spelling the literal; the false lineage claim above is struck through with the measured attribution recorded (VCI §1.1 surface-3 correction) |

Verification batch at the r3 repair head (this tree, this run):

- E2 builds: `go build ./...` → exit 0.
- Affected families green (-count=1): internal/auditreceipt (full), internal/runtime (full), internal/closure (full), internal/mcp (full), and the cli families TestCodexAudit*, TestAuditMulti*, TestSelfReview*, TestGLMAudit*, TestConverge*, TestRunMultiAudit*.
- Lint (full 8-root touched set): 4 errcheck, all pre-existing in internal/cli — zero NEW.
- The absorbed-era pin family TestAuditMulti_CodexLegDeadline_* stays red pending the leader disposition this card already requested (F11 note above) — unchanged by this round.

### Sync-audit r4 delta repair round — F16/F17 (card t1632, branch WT-3-2-0-2)

Commit: `47ad706c2` (fix: sync-audit r4 repairs — F16/F17). Base: the r3-repair head `e6854e5d8`. Verdict file: .moai/reports/t1632/sync-audit-r3.md (F12/F13/F14 resolved; F16 blocking, F17 note). Push state: nothing pushed; clean on WT-3-2-0-2.

Repair matrix (run at HEAD 47ad706c2, this tree):

| Finding | Disposition | Command (verbatim) | Observed output |
|---------|-------------|--------------------|-----------------|
| F16 shared plain bool in the F13 pin (data race under -race) | FIXED | `go test -race ./internal/cli -run '^TestAuditMulti_ReceiptBindsStartOfConvergenceState$' -count=1` | `ok  github.com/modu-ai/moai-adk/internal/cli  2.210s` — the postEdit flag is an atomic.Bool (Store in the errgroup goroutine, Load in the TreeState stub); no race report |
| F17 orphaned recordAuditReceiptAt doc comment | FIXED | `go build ./...` + lint | `ok` — the orphan folded into recordAuditReceiptAtBound's own doc (the wrapper it described was removed when every caller migrated to the bound surface) |

### Sync-audit r5 delta repair round — foreground codex_task cancel regression (card t1632, branch WT-3-2-0-2)

Commit: `17a0547b6` (fix: repair the foreground codex_task cancel path). Base: the follow-up-card memo head `43209a255`. RED already on disk: `TestCodexTask_ForegroundSessionStillBoundToRequest` panicked (nil deref at codex_task_session_ctx_test.go:317). Push state: nothing pushed; clean on WT-3-2-0-2.

Root cause (all inside this card's own diff — the leader's "M2 이후 무변경" premise corrected):

1. **Containment listing on the request context**: this card's REQ-MN-009 repair added the MCP-listing precondition to `openCodexSessionResolved` **before** the session opens, and it ran on the REQUEST context. An already-cancelled request failed the listing ("mcp list: context canceled"), the session never opened, the session-context double's `start` was never called, its `openCtx` stayed nil, and the foreground-cancel test panicked dereferencing `sess.current()`'s nil context (test line 317 — NOT a nil `*mcp.CallToolResult`; the handler always returned a structured `status=failed`, verified by an instrumented exact copy). The listing now runs on the BACKGROUND context — the same precondition-before-session-work rule the serving anchor lookup already applied (codex_task.go: "the git verification runs on the background context, and the cancellation surfaces later, at the session").
2. **Session doubles lacked the containment stub**: `withCtxBoundCodexSession`, `withCodexConn`, and the codex_task process-context fixtures never stubbed `codexAuditMCPList`, so after the listing precondition existed every session-opening test executed the stub binary and failed before opening. Each now installs the same empty-server stub `withCodexSession` already carried.
3. **Background-job tests re-anchored on the serving tree**: the job registry, the cancel call, and the REQ-MN-005 confine must describe ONE repository — a temp fixture root is not a registered worktree, so the cancel's confine refused it ("belongs to a different repository" / "server root ... is not a git repository" depending on fixture state). The three background tests and the cancel call now use `thisRepoRoot` for both, with registry records removed in cleanup.

Repair matrix (run at HEAD 17a0547b6, this tree):

| Check | Command (verbatim) | Observed output |
|-------|--------------------|-----------------|
| The r5 RED (foreground cancel panic) | `go test ./internal/cli -run '^TestCodexTask' -count=1` | `ok` — the whole family green including the formerly panicking test: status=failed returned structured, the session opened on the request context (`openCtx.Err() != nil` after cancel), and the child ended |
| Broader family | the same run | all `^TestCodexTask*` selectors green (background detach/cancel/turn-bound, protocol liveness, process-context) |

**Why the regression survived four rounds — verification-scope disclosure**: `TestCodexTask_Foreground*` (and the whole codex_task_session_ctx/process_context files) appeared in NO §E.2 verification batch — the milestone batches ran the audit/review/self-review families and the M3 batch scoped hook+factorymsg, but never the codex_task session-context files, and the F2 correction's 8-root lint re-measure fixed lint visibility without extending the TEST surface. This round adds the codex_task session-context family to the standing batch surface.

Authored-By-Agent: manager-develop

### Sync-audit r6 delta repair round — F19/F18/F20/F21 (card t1632, branch WT-3-2-0-2)

Commit: `807db1c8c` (fix: sync-audit r6 delta repairs — F19/F18/F20) + the F21 correction in this file. Base: the r5 evidence head `19089dac8`. Verdict file: .moai/reports/t1632/sync-audit-r5.md. Push state: nothing pushed; clean on WT-3-2-0-2.

| Finding | Disposition | Command (verbatim) | Observed output |
|---------|-------------|--------------------|-----------------|
| F19 cleanup removes the serving tree's real job registry | FIXED + canary measurement | `go test ./internal/cli -run 'TestCodexTask_Background' -count=1` | `ok` — the cleanup removes ONLY the test's own record file; a canary record (`canary-pre-existing-record.json`, planted before the job with real-record content) is asserted to SURVIVE the cleanup, then removed as the test's own planting — a future widening back into a directory removal fails loudly on the canary check |
| F18 handshake bound test flaked 4/5 at the 700ms margin | FIXED (margin + measured cause) | `go test ./internal/cli -run '^TestCodexTaskBackgroundHandshakeHonorsTaskBound$' -count=5` | `ok  github.com/modu-ai/moai-adk/internal/cli  6.190s` — green at -count=5. Measured cause: the handler's path to its structured "timed out after 100ms" result crosses real git queries whose latency under concurrent git load (live ps during the hang showed another session's `git status --porcelain` running and a network `git clone` in flight) exceeded the 700ms watchdog while the 100ms task bound itself had long fired. The pinned contract is the bound ending the handshake (the error text), not a 700ms wall clock; the watchdog margin is 2s |
| F20 unbounded detached containment listing | FIXED | `go build ./...` + the bound applied in openCodexSessionResolved | `ok` — the background-context listing is bounded by its own context timeout (DefaultCodexTaskTimeout), so a wedged-child listing can no longer outlive the request by minutes |
| F21 progress.md env-literal origin sentence contradicts the r2 measurement and the r3 CORRECTION | FIXED (correction in place) | the F21 CORRECTION at this file's follow-up-memo line | the main-origin sentence is struck and replaced with the measured attribution (this card's own M1/M2 comment lines), matching the r2 measurement and the r3 CORRECTION |

Verification batch at HEAD 807db1c8c (this tree, this run):

- E2 builds: `go build ./...` → exit 0.
- Affected families green: `go test ./internal/cli -run '^TestCodexTask' -count=1` → ok (full family incl. the r5 RED test); internal/auditreceipt full → ok; internal/hook full → ok (501s; the env-literal guard stays green after the F12 rewording).
- Lint (cli + hook): 4 errcheck, all pre-existing — zero NEW.

Authored-By-Agent: manager-develop

### Sync-audit r7 delta repair round — F22/F23 structural isolation (card t1632, branch WT-3-2-0-2)

Commit: `fcfad897e` (fix: sync-audit r7 delta repairs — F22/F23 structural isolation). Base: the r6 evidence head `5bdc1c212`. Verdict file: .moai/reports/t1632/sync-audit-r6.md (F18/F20/F21 resolved; F22 canary overwrite+symlink, F23 self-removal-before-check ordering). Push state: nothing pushed; clean on WT-3-2-0-2.

Adopted disposition: the auditor's recommended STRUCTURAL ISOLATION over the minimal canary patch — startCtxBoundBackgroundJob now isolates the job registry in a TEMPORARY git repository (the serving-project resolver, the job's project_root, AND the serving anchor codexServingDir all name it), so the job's records never touch the serving tree's real state at all. With the isolation in place the canary's fixed-name/non-exclusive write and its self-removal-before-check ordering are structurally gone; the mutation tripwire STAYS as the regression proof:

| Check | Command (verbatim) | Observed output |
|-------|--------------------|-----------------|
| F23 mutation detection (the guard fires on exactly the removal it guards) | `go test ./internal/cli -run 'TestCodexTask_CanaryGuardCatchesRegistryRemoval' -count=1` | `ok` — canaryLost reports the loss after the RemoveAll mutation and is silent while the canary is present (the r6 device failed this mutation: its own self-removal ran before the check) |
| F22/F23 structural isolation (zero contact with the serving state) | `go test ./internal/cli -run 'TestCodexTask_BackgroundJobNeverTouchesTheServingState' -count=1` | `ok` — a background job run leaves ZERO new records under the serving tree's job registry (before/after directory snapshot compared) |
| The repaired family | `go test ./internal/cli -run '^TestCodexTask' -count=1` | `ok` — the whole family green, including the three background tests re-anchored on the isolated registry |

r6 F18 follow-through (same round): the stack-observed hold point (handleCodexTask → codexAuditValidateRoot → codexAuditGit) was an UNBOUNDED background git read — it now carries its own context bound (DefaultCodexAuditListTimeout), and the bound test's watchdog margin is 10s with the reason recorded (git queries under concurrent load are seconds, and the pinned contract is the 100ms bound ending the handshake — visible in the "timed out after 100ms" error — not a wall clock).

Verification batch at HEAD fcfad897e (this tree, this run):

- E2 builds: `go build ./...` → exit 0.
- Affected families green (-count=1): `^TestCodexTask` full family (2× runs at different stages), internal/auditreceipt full.
- Lint (cli + auditreceipt): 4 errcheck, all pre-existing — zero NEW.

### Sync-audit r8 delta repair round — F24 absent-registry read (card t1632, branch WT-3-2-0-2)

Commit: `433c09af2` (fix: sync-audit r8 delta repair — F24 absent-registry read). Base: the r7 evidence head `100522d53`. Verdict file: .moai/reports/t1632/sync-audit-r7.md (F22/F23 isolation resolved; F24 blocking, F25 optional). Push state: nothing pushed; clean on WT-3-2-0-2.

| Finding | Disposition | Command (verbatim) | Observed output |
|---------|-------------|--------------------|-----------------|
| F24 absent serving registry reads as fatal ENOENT | FIXED + no-registry probe | `rm -rf .moai/state/codex-jobs && go test ./internal/cli -run 'TestCodexTask_BackgroundJobNeverTouchesTheServingState' -count=1 -v` | `--- PASS: TestCodexTask_BackgroundJobNeverTouchesTheServingState (0.46s)` — the before/after reads treat IsNotExist as an empty listing (a clean checkout has no .moai/state) and only an unexpected read error is fatal; the probe ran with the registry directory removed and stayed green, and the serving registry remained record-free after the run (0 entries) |
| F25 comparison scope (new-records-only) | HELD per the leader — rides t1747 | — | noted here: the structural isolation is the primary protection this round; widening the pre-existing-record comparison rides t1747 with the other held findings |

Verification batch at HEAD 433c09af2 (this tree, this run):

- E2 builds: `go build ./...` → exit 0.
- Affected families green (-count=1): `^TestCodexTask` full family.
- Lint: no new findings (the 4 errcheck stay pre-existing).

Authored-By-Agent: manager-develop

### Sync-audit r9 delta repair round — remeasure regressions (card t1632, branch WT-3-2-0-2)

Commit: `8f22e6401` (fix: sync-audit r9 delta repairs — remeasure regressions). Base: the r8 evidence head `486c9a6b1`. Baseline evidence: /tmp/t1632-remeasure-out3.jsonl (full capture), /tmp/t1632-baseline-out.txt (main-baseline-t1632 @ e04ebd5ca — all 5 PASS there). Push state: nothing pushed; clean on WT-3-2-0-2.

| Failure | Disposition | Command (verbatim) | Observed output |
|---------|-------------|--------------------|-----------------|
| TestCommittedWorkflowYamlAuditAlignment (claude effort high vs want medium) | FIXED — the test file was a PARTIAL MERGE of card t1520's alignment commit `602701d94`: its doc comment described the operator-role-table posture (effort at the Go default, high) while its two effort assertions still expected D8's medium. Restored to the baseline file (`e04ebd5ca`), which asserts high consistently; the committed yaml is unchanged (main's `842160157` already carried effort: high) | `go test ./internal/config -run 'TestCommittedWorkflowYamlAuditAlignment' -count=1` | `ok` |
| TestCodexContractByteCeiling (AGENTS.md 24724/24728 > ceiling 24576) | FIXED — this card's M4 added AGENTS.md §14 (the MCP reconnect procedure, ~+800B) to both copies; §14 condensed to the contract-level facts (the per-harness bullet list folds into one sentence) | `go test ./internal/config -run 'TestCodexContractByteCeiling\|TestCodexNestedTemplateDiscoveryBudget\|TestCommittedWorkflowYamlAuditAlignment' -count=1` | `ok` — both copies back under the ceiling |
| TestCodexNestedTemplateDiscoveryBudget (same ceiling, discovery chain) | FIXED — same §14 condensation | the same run | `ok` |
| TestBrokerOpenContextGitFailuresDoNotFallbackIdentity (+pipe) | NOT REPRODUCIBLE — no code change | `go test ./internal/factorymsg -run 'TestBrokerOpenContextGitFailuresDoNotFallbackIdentity' -count=1` and `go test ./internal/factorymsg -count=1` | both `ok` — isolated AND full-package runs green on the card tree; the remeasure failure was load/environment contention in the remeasure harness (recorded) |
| hook TestFactoryHookContextAndContinuationSafety + TestFactoryHookZeroTurnAndCapabilityTruth | VERIFIED GREEN on the card tree — no change | `go test ./internal/hook -run 'TestFactoryHookContextAndContinuationSafety\|TestFactoryHookZeroTurnAndCapabilityTruth' -count=1` | `ok` — the M5 repair surface (`5e468bce1` + `db9fb0e7b`) is intact; the main-side failures are the expected absence of those repairs on main (main PASS in the leader's matrix refers to the t1632-absent tree) |

Verification batch at HEAD 8f22e6401 (this tree, this run):

- E2 builds: `go build ./...` → exit 0.
- Affected families green (-count=1): internal/config (the three named tests), internal/factorymsg (full package, 233s), internal/hook (the two named + the family run earlier this round).

Authored-By-Agent: manager-develop

### Lane-57 sync-audit r1 repair round — l57-r1 F1–F4 (card t1632, branch WT-3-2-0-2)

Commit: `64f8ce1fd6` (fix(t1632): lane-57 sync-audit r1 repairs — audit-leg env marker, glm job root, expired-claim inbox, path-based staleness). Base: the r9 evidence head `b094733f97`. Audit receipt: `rcpt-99ea6bee74f2473b25362a18` (`.moai/reports/t1632/sync-audit-l57-r1.md`, codex_audit target=baseBranch mode=native, FAIL 4건). Note: the WIP of an interrupted session carried most of these repairs already; this round completed F2's three consumer call sites and captured the honest RED evidence per finding.

| Finding | Disposition | Command (verbatim) | Observed output |
|---------|-------------|--------------------|-----------------|
| F1 [P1] audit-leg env marker lost through the scrub (mcp_codex.go) | FIXED — `realCodexSessionRunner.start` re-appends `MOAI_CODEX_AUDIT_LEG=1` from the leg CONTEXT after the `codexChildEnv()` scrub; identity keys stay dropped | `go test ./internal/cli -run '^TestRealCodexSessionRunnerScrubRetainsAuditLegMarker$' -count=1 -timeout 120s` | GREEN: `--- PASS ... (0.22s)` / `ok internal/cli` |
| F2 [P2] glm job family ignored the requested project_root (glm_task.go + glm_job_control.go) | FIXED — producer roots the background job registry at the validated `glmRoot`; the three consumers (`glm_job_status`/`result`/`cancel`) pass the request-resolved root into `loadGLMJobFor` (REQ-MN-004 create/read parity) | `go test ./internal/cli -run '^(TestGLMTaskBackgroundJobRecordLandsUnderNamedRoot\|TestGLMJobStatusReadsTheNamedRoot)$' -count=1` | GREEN: both `--- PASS`, `ok internal/cli` |
| F3 [P2] expired-claim messages invisible in PendingList (factorymsg/store.go) | FIXED — the query now renders `state='pending' OR (state='claimed' AND claim_expires_at<=now)`; a LIVE claim stays excluded (PendingCount parity, REQ-MN-011) | `go test ./internal/factorymsg -run '^TestPendingListIncludesExpiredClaim$' -count=1` and full package `go test ./internal/factorymsg -count=1` | GREEN: `--- PASS` / full package `ok ... 52.185s` |
| F4 [P2] staleness judged by executable PATH (mcp_server_runtime.go) | FIXED — freshness judges build identity (size+mtime from the start-time capture), never the PATH; an in-place reinstall still moves size or mtime | `go test ./internal/cli -run '^TestServerStaleness_SameBuildAtDifferentPathReadsFresh$' -count=1` | GREEN: `--- PASS` / `ok internal/cli` |

E8 RED evidence (each pinning test run against its pre-fix code; the single implementation file reverted to HEAD with the WIP copy backed up and restored byte-identically after capture, `git diff --stat` verified before/after):

- F1 RED (`go test ./internal/cli -run '^TestRealCodexSessionRunnerScrubRetainsAuditLegMarker$' -count=1 -timeout 120s`, mcp_codex.go at HEAD, exit 1):
  `--- FAIL: TestRealCodexSessionRunnerScrubRetainsAuditLegMarker (0.22s)` / `codex_audit_leg_test.go:199: the scrubbed child environment lost the audit-leg marker (got "", want "1") — the leg's Stop chain re-runs its verification members (t1754)`
- F2 producer RED (glm_task.go + glm_job_control.go at HEAD, `go test ./internal/cli -run '^TestGLMTaskBackgroundJobRecordLandsUnderNamedRoot$|^TestGLMJobStatusReadsTheNamedRoot$' -count=1 -timeout 120s`; exit code not separately captured — output was piped through `tee`, the package FAIL is in the verbatim output):
  `--- FAIL: TestGLMTaskBackgroundJobRecordLandsUnderNamedRoot (0.21s)` / `glm_task_root_test.go:34: background glm_task with project_root=/private/var/.../002/W wrote no job record under the named root: glm job not found: job-20261021T065825Z-73da2b5c`
  (TestGLMJobStatusReadsTheNamedRoot alone does NOT go red on the full pre-fix pair — pre-fix producer and consumer both used the server tree, so the pair agreed wrongly; the test comment states this.)
- F2 consumer RED (producer fixed, glm_job_control.go at HEAD, `go test ./internal/cli -run '^TestGLMJobStatusReadsTheNamedRoot$' -count=1 -timeout 120s`, exit 1):
  `--- FAIL: TestGLMJobStatusReadsTheNamedRoot (0.23s)` / `glm_task_root_test.go:59: glm_job_status could not read the job record at the root glm_task wrote it to: glm_job_status: no job record for id job-20261021T065911Z-fe8542d3`
- F3 RED (store.go at HEAD, `go test ./internal/factorymsg -run '^TestPendingListIncludesExpiredClaim$' -count=1 -timeout 120s`, exit 1):
  `--- FAIL: TestPendingListIncludesExpiredClaim (0.03s)` / `pending_list_test.go:54: PendingList after the claim lease = []; want the expired-claim row rendered as pending (the hook's empty-inbox path hid a redeliverable message)`
- F4 RED (mcp_server_runtime.go at HEAD, `go test ./internal/cli -run '^TestServerStaleness_SameBuildAtDifferentPathReadsFresh$' -count=1 -timeout 120s`, exit 1):
  `--- FAIL: TestServerStaleness_SameBuildAtDifferentPathReadsFresh (0.00s)` / `mcp_server_staleness_test.go:95: an identical build at a different PATH location judged the current server stale: {Stale:true Serving:.../moai-serving Installed:.../moai-installed Version:v3.1.3 Commit:none}`

GREEN (the four pinning tests, one selector run): `go test ./internal/cli -run '^(TestRealCodexSessionRunnerScrubRetainsAuditLegMarker|TestGLMTaskBackgroundJobRecordLandsUnderNamedRoot|TestGLMJobStatusReadsTheNamedRoot|TestServerStaleness_SameBuildAtDifferentPathReadsFresh)$' -count=1 -timeout 180s -v` → exit 0, 4/4 `--- PASS`, `ok github.com/modu-ai/moai-adk/internal/cli 2.090s`.

Verification batch (this tree, this run — run on the working tree whose staged content is commit `64f8ce1fd6`, no source file changed between the last test run and the commit):

- Builds: `go build ./...` → `NATIVE_BUILD_EXIT=0`; `GOOS=windows GOARCH=amd64 go build ./...` → `WINDOWS_BUILD_EXIT=0`.
- Affected families, all `-count=1` exit 0: `^TestCodexTask` → `ok internal/cli 11.550s`; `TestGLMTask|TestGLMJob` → `ok internal/cli 8.318s`; `TestServerStaleness|TestMCPServerRuntimeDir|TestCodexAuditEnvAllowlist` → `ok internal/cli 1.154s`; `TestAuditLeg` → `ok internal/cli 0.889s`; `^TestCodexAudit` → `ok internal/cli 25.130s`; `./internal/factorymsg` full package → `ok ... 52.185s`.
- `go vet ./internal/cli ./internal/factorymsg` → clean (`VET_CLEAN_EXIT=0`).
- `golangci-lint run ./internal/cli/... ./internal/factorymsg/...` → exit 1 with exactly the 4 pre-existing errcheck in internal/cli (factory_run_adopt.go:79/91/103, mcp_server_call_log.go:123) — `4 issues: errcheck: 4`, ZERO new.
- No timing-sensitive flake observed in any run; no re-run needed.
- Full-suite note: the repository-wide test verdict is owned by the CI run on the project's integration branch and is PENDING at report time (lane-local targeted selectors only, per the standing batch surface).
- Push state: nothing pushed (the lane owns push). progress.md evidence block follows the prior rounds' separate docs-commit pattern (fix commit → `docs(...): E.2` evidence commit).

Authored-By-Agent: manager-develop

### Sync-audit r2 delta repair round — landing re-audit (card t1632, branch WT-3-2-0-2)

Commit: `fix(t1632): lane-57 sync-audit r2 repairs — receipt digest whitespace, convergence FAIL retention, summarizer env hardening` (this round). Base: HEAD `e80010e89c` (the absorb merge of release head 6a523191ea). Audit receipt: `rcpt-56d5695de1452b970e181315` (static codex_audit on the merged tree — it ran NO tests; each finding was verified honestly before repair, all 3 proved REAL).

| Finding | Disposition | Command (verbatim) | Observed output |
|---------|-------------|--------------------|-----------------|
| F1 [P1] receipt tree digest loses whitespace filenames (store.go:267) | REAL-FIXED — TWO layers: (a) the `untrackedContentPaths` TrimSpace on the split -z path (the auditor's named site), and (b) deeper, `gitIn`'s WHOLE-output TrimSpace strips the leading space of the FIRST ls-files entry (` a.go\x00` → `a.go\x00`); fixed by a new `gitInRaw` variant (untrimmed) that `untrackedContentPaths` now uses. The porcelain status read keeps `gitIn` (its output begins `XY `, never whitespace) | `go test ./internal/auditreceipt -run 'TestTreeState_WhitespaceFilenameContentMovesDigest' -count=1` | RED→GREEN: `--- FAIL ... digest unchanged after a leading-space untracked file's content changed` → `ok github.com/modu-ai/moai-adk/internal/auditreceipt` |
| F2 [P2] required FAIL overwritten by required-inconclusive (mcp_convergence.go:1086) | REAL-FIXED — `enforceRequiredGateUnmet` now downgrades to inconclusive ONLY when the overall does not itself carry a required FAIL; a converged FAIL (a failure a backend REPORTED) is kept, and `gate_unmet` still names the inconclusive leg. `TestConverge_RequiredFailWithInconclusive_Case2` (pure converge) was already pinned; the loss was at the enforcement layer, unpinned | `go test ./internal/cli -run 'TestEnforceRequiredGateUnmet_KeepsAReportedRequiredFail' -count=1` | RED→GREEN: `--- FAIL ... overall = "inconclusive", want the reported FAIL kept` → `ok internal/cli` |
| F3 [P2] participation summarizer env keeps proxy variables (mcp_claude.go:352) | REAL-FIXED — `claudeParticipationEnv` (denylist scrub) kept HTTPS_PROXY while `Summarize` then validated its own env, dying at `proxy variable HTTPS_PROXY must not reach the claude audit child` before summarizing. The audit path's allowlist (`claudeAuditAllowlistEnv`) never lets a proxy through; the denylist path now drops the same family the validator rejects (`NO_PROXY`, any `*_PROXY`). The allowlist switch was NOT taken — the existing pin `TestParticipationSummarizerEnvCarriesTheBaseEnvironment` requires HOME in the env, which the allowlist drops | `go test ./internal/cli -run 'TestParticipationSummarizerEnvDropsProxyVariables' -count=1` | RED→GREEN: `--- FAIL ... the summarizer environment carries HTTPS_PROXY` → `ok internal/cli` |

E8 RED evidence (each pinning test run against the pre-fix code, all three REAL):

- F1 RED (`go test ./internal/auditreceipt -run 'TestTreeState_WhitespaceFilenameContentMovesDigest' -count=1`, exit 1):
  `--- FAIL: TestTreeState_WhitespaceFilenameContentMovesDigest (0.78s)` / `receipt_integrity_content_test.go:93: digest unchanged after a leading-space untracked file's content changed — the trimmed path read the tracked a.go instead (l57-r2 finding 1)`
- F2 RED (`go test ./internal/cli -run 'TestEnforceRequiredGateUnmet_KeepsAReportedRequiredFail|TestParticipationSummarizerEnvDropsProxyVariables' -count=1`, exit 1):
  `--- FAIL: TestEnforceRequiredGateUnmet_KeepsAReportedRequiredFail (0.00s)` / `codex_audit_could_not_run_test.go:144: overall = "inconclusive", want the reported FAIL kept — a required-inconclusive leg reclassified a found failure as not-run (l57-r2 finding 2)`
- F3 RED (same run, exit 1):
  `--- FAIL: TestParticipationSummarizerEnvDropsProxyVariables (0.00s)` / `feedback_participation_model_test.go:89: the summarizer environment carries HTTPS_PROXY — the audit env validator rejects proxy variables, so Summarize dies before summarizing (l57-r2 finding 3)`

GREEN: F1 `go test ./internal/auditreceipt -run 'TestTreeState|TestCheckCitedReceipts' -count=1` → `ok ... 1.331s`; F2+F3 `go test ./internal/cli -run 'TestEnforceRequiredGateUnmet|TestParticipationSummarizer|TestFlushWiresClaudeSummarizer' -count=1` → `ok internal/cli 1.381s`.

Verification batch (this tree, this run, at the r2 HEAD):

- Builds: `go build ./...` → exit 0; `GOOS=windows GOARCH=amd64 go build ./...` → exit 0.
- `go test ./internal/auditreceipt -count=1` (full) → `ok github.com/modu-ai/moai-adk/internal/auditreceipt 4.509s`.
- `go test ./internal/cli -run 'TestConverge|TestEnforceRequiredGateUnmet|TestRunMultiAudit' -count=1` → `ok internal/cli 3.856s`.
- `go test ./internal/cli -run 'TestClaudeAudit' -count=1` → `ok internal/cli 2.073s`.
- `go vet ./internal/cli ./internal/auditreceipt` → clean (exit 0).
- `golangci-lint run ./internal/cli/... ./internal/auditreceipt/...` → exactly the 1 pre-existing errcheck (`internal/cli/mcp_server_call_log.go:123`), ZERO new.
- Full-suite note: the repository-wide test verdict is owned by the CI run on the project's integration branch and is PENDING at report time (lane-local targeted selectors only, per the standing batch surface).
- Push state: nothing pushed (the lane owns push/landing).

Authored-By-Agent: manager-develop

### Sync-audit r3 delta repair round — landing re-audit (card t1632, branch WT-3-2-0-2)

Commit: `fix(t1632): lane-57 sync-audit r3 repairs — boundary git env scrub, panel isolation, relative spec_dir, cache key collision` (this round). Base: HEAD `263d3254e9` (the r2 repairs commit). Audit receipt: `rcpt-9708ddd8d674aaa80c7390b6` (STATIC codex_audit on the merged tree — it ran NO tests; each finding was verified honestly before repair, all 4 proved REAL).

| Finding | Disposition | Command (verbatim) | Observed output |
|---------|-------------|--------------------|-----------------|
| F1 [P1] the repo-boundary git checks inherit hostile GIT_DIR/GIT_WORK_TREE (mcp_project_root.go:203 via codexAuditGit) | REAL-FIXED — `codexAuditGit` passed no `cmd.Env`, so the confinement's two common-dir lookups AND the worktree listing executed in whatever repository an inherited GIT_DIR named; a foreign project_root passed confinement. Fixed by scrubbing `GIT_DIR`/`GIT_WORK_TREE`/`GIT_COMMON_DIR`/`GIT_INDEX_FILE`/`GIT_CEILING_DIRECTORIES` (plus a locale pin, `LC_ALL=C`) inside `codexAuditGitEnv()` — the same scrub l57-r1 F4 landed in internal/auditreceipt (`scrubbedGitEnv`, TestTreeState_IgnoresInheritedGitDir). The fix lands at the shared helper, so codexAuditValidateRoot and the CLI verb path are covered by the same repair | `go test ./internal/cli -run 'TestMCPRootConfinementHostileGitDirEnv' -count=1` | RED→GREEN: `--- FAIL ... a foreign project_root passed confinement under an inherited hostile GIT_DIR` → `ok internal/cli` |
| F2 [P1] the r1 audit-leg marker fix resurrects Orca panel env vars (mcp_codex.go:612) | REAL-FIXED — `realCodexSessionRunner.start` replaced `cmd.Env` wholesale with `codexChildEnv()`, discarding the `dropOrcaPaneEnv` pass codexChildCmd had applied through `codexReviewChildEnv`; ORCA_PANE_KEY leaked into the session child. Fixed by composing BOTH filters — `cmd.Env = codexAuditChildEnv()` (= `dropOrcaPaneEnv(codexChildEnv())`) — with the r1 audit-leg ctx re-append kept verbatim. The r1 pin `TestRealCodexSessionRunnerScrubRetainsAuditLegMarker` stays green, now extended to also probe ORCA_PANE_KEY | `go test ./internal/cli -run 'TestRealCodexSessionRunnerScrubRetainsAuditLegMarker' -count=1` | RED→GREEN: `--- FAIL ... the session runner leaked ORCA_PANE_KEY="leaked" into the codex child` → `ok internal/cli` |
| F3 [P2] a relative spec_dir resolves against the server's CWD, not the request's project_root (mcp_server.go:1252) | REAL-FIXED — `handleAuditCache` compute_hash read `spec_dir` as-is, so a server whose cwd sat in another tree returned THAT tree's plan hash while `_root` claimed the requested one. Fixed by joining a relative spec_dir onto the validated root — the same anchoring the store applies to a relative report_path (card-review P2-7). `TestAuditCacheRelativeSpecDirAnchoredToRoot` pins it: two trees with different plan.md, relative spelling must equal the absolute root-anchored hash | `go test ./internal/cli -run 'TestAuditCacheRelativeSpecDirAnchoredToRoot' -count=1` | RED→GREEN: `--- FAIL ... relative spec_dir hash = 7cb8fe4b…, want the named tree's 944e2262…` → `ok internal/cli` |
| F4 [P2] sanitizeStateKey collisions serve one SPEC's PASS to another (audit_cache_state.go:133) | REAL-FIXED — `SPEC-A/B` and `SPEC-A_B` both sanitize to `SPEC-A_B.json`, and `Lookup` verified only the artifact hash, never the stored SpecID; with identical hashes the first SPEC's PASS report was served to the second. Fixed per the auditor: `Lookup` now compares `e.SpecID != specID` and misses on mismatch (the storing SPEC's own lookup still hits) | `go test ./internal/runtime -run 'TestStateCache_LookupRefusesSpecIDCollision' -count=1` | RED→GREEN: `--- FAIL ... a separator-collision lookup was served another SPEC's PASS entry` → `ok internal/runtime` |

E8 RED evidence (each pinning test run against the pre-fix code, all four REAL):

- F1 RED (`go test ./internal/cli -run 'TestMCPRootConfinementHostileGitDirEnv' -count=1`, exit 1):
  `--- FAIL: TestMCPRootConfinementHostileGitDirEnv (0.74s)` / `mcp_root_contract_test.go:133: a foreign project_root passed confinement under an inherited hostile GIT_DIR — the boundary git checks executed in the foreign repository (sync-audit r3 F1)`
- F2 RED (`go test ./internal/cli -run 'TestRealCodexSessionRunnerScrubRetainsAuditLegMarker' -count=1 -v`, exit 1):
  `codex_audit_leg_test.go:223: the session runner leaked ORCA_PANE_KEY="leaked" into the codex child — codexChildCmd's pane drop was discarded when cmd.Env was replaced (sync-audit r3 F2)`
- F3 RED (`go test ./internal/cli -run 'TestAuditCacheRelativeSpecDirAnchoredToRoot' -count=1 -v`, exit 1):
  `mcp_server_test.go:465: relative spec_dir hash = 7cb8fe4b2dc22016666fda881cb9b3c13c2e13a2337c4b980fb655759ac77834, want the named tree's 944e226249bc3fc79041cd1c6c0318a69bbfb5d6c25f60a3ffe8b98d690a369a — the relative path resolved against the server's working directory, not the validated project root (sync-audit r3 F3)`
- F4 RED (`go test ./internal/runtime -run 'TestStateCache_LookupRefusesSpecIDCollision' -count=1 -v`, exit 1):
  `audit_cache_state_test.go:112: a separator-collision lookup was served another SPEC's PASS entry — the stored SpecID is not compared on lookup (sync-audit r3 F4)`

Verification batch (this tree, this run, at the r3 HEAD):

- Builds: `go build ./...` → exit 0; `GOOS=windows GOARCH=amd64 go build ./...` → exit 0.
- `go test ./internal/cli -run 'TestMCPRoot|TestServerStaleness|TestRealCodexSessionRunnerScrubRetainsAuditLegMarker|TestCodexAudit' -count=1` → `ok github.com/modu-ai/moai-adk/internal/cli 26.041s`.
- `go test ./internal/runtime -count=1` (full) → `ok github.com/modu-ai/moai-adk/internal/runtime 3.743s`.
- `go test ./internal/cli -run 'TestMCPGateExposure|TestSpecProgress|TestVerifySnapshot|TestVerifyTrend' -count=1` → `ok github.com/modu-ai/moai-adk/internal/cli 1.456s`.
- `go vet ./internal/cli ./internal/runtime` → clean (exit 0).
- `golangci-lint run ./internal/cli/... ./internal/runtime/...` → exactly the 1 pre-existing errcheck (`internal/cli/mcp_server_call_log.go:123`), ZERO new.
- Full-suite note: the repository-wide test verdict is owned by the CI run on the project's integration branch and is PENDING at report time (lane-local targeted selectors only, per the standing batch surface).
- Push state: nothing pushed (the lane owns push/landing).

Authored-By-Agent: manager-develop

### Sync-audit r4 delta repair round — landing re-audit (card t1632, branch WT-3-2-0-2)

Commit: `fix(t1632): lane-57 sync-audit r4 repairs — unconditional audit-child containment, hook skips delivered claims` (this round). Base: HEAD `b0d0592114` (the r3 repairs commit). Audit receipt: `rcpt-b63b85fc92766d8863970695` (STATIC codex_audit — it ran NO tests; each finding was verified honestly against the code before repair, both proved REAL).

| Finding | Disposition | Command (verbatim) | Observed output |
|---------|-------------|--------------------|-----------------|
| F1 [P1] an OMITTED project_root skips the audit child's MCP containment entirely (mcp_codex.go:949, openCodexSessionResolved) | REAL-FIXED — the containment block keyed on `params["cwd"]`, so the audit_multi codex leg with no named root spawned a child with NO foreign-server disable args (the child resolved the operator's own MCP servers, write tools included). Fixed per the auditor: containment is now UNCONDITIONAL — the anchor is the request's `cwd` when named, else the server's own project-root resolution (`projectDirResolver()`), and a server that cannot resolve ANY root refuses the spawn (fail-closed). The params map is untouched: an omitted root still reaches the thread as an omission (REQ-2 pass-through preserved — the fix changes CHILD-Spawn containment, not the root-validation policy for which trees may be audited) | `go test ./internal/cli -run 'TestCodexAudit_NoProjectRootStillContainsChild' -count=1` | RED→GREEN: `--- FAIL ... no-root audit child argv = [app-server -c sandbox_mode=read-only], want the foreign-server disable` → `ok internal/cli` |
| F2 [P2] the hook re-claims already-delivered messages (internal/hook/factory_messages.go, the s.Claim call site) | REAL-FIXED — the store-level Claim still returns live claims (REQ-MN-017, the M5 contract built for the deleted MCP factory_msg_list surface), but the hook claims on EVERY inspection, so a message rendered at UserPromptSubmit was re-claimed and re-rendered at Stop (an extra forced turn, repeating until the lease expired). Fixed at the HOOK CALL SITE only: the hook now claims with `ClaimFresh`, which skips rows the endpoint still holds (the redelivery policy the managed-session driver already uses). grep verified the hook IS the only surviving production Claim caller post-t1908 (internal/cli/managed_factory_session.go uses ClaimFresh; todo_claim.go is a different type). The store-level M5 pin survives untouched: TestClaimFreshSkipsAHeldClaim asserts BOTH that Claim keeps REQ-MN-017 token stability AND that ClaimFresh skips a held claim — its premise is the store primitive, not the deleted tool | `go test ./internal/hook -run 'TestFactoryHookBatchSkipsDeliveredClaims' -count=1` | RED→GREEN: `--- FAIL ... the Stop inspection re-rendered the message UserPromptSubmit already delivered (state=pending-at-turn-boundary)` → `ok internal/hook` |

E8 RED evidence (each pinning test run against the pre-fix code, both REAL):

- F1 RED (`go test ./internal/cli -run 'TestCodexAudit_NoProjectRootStillContainsChild' -count=1`, exit 1):
  `--- FAIL: TestCodexAudit_NoProjectRootStillContainsChild (0.00s)` / `codex_audit_sandbox_test.go:53: no-root audit child argv = [app-server -c sandbox_mode=read-only], want the foreign-server disable (REQ-MN-009 containment is unconditional)`
- F2 RED (`go test ./internal/hook -run 'TestFactoryHookBatchSkipsDeliveredClaims' -count=1`, exit 1):
  `--- FAIL: TestFactoryHookBatchSkipsDeliveredClaims (0.44s)` / `factory_messages_test.go:751: the Stop inspection re-rendered the message UserPromptSubmit already delivered (state=pending-at-turn-boundary): ctx="Factory inbox run=r4-delivered-skip metadata …86ae5145380939acbb1411856b78b38b kind=status_request from=r4-sender lane=lead harness=claude generation=1 token=5de6cdcf13d01d60976eb0ba02d6dacd"`

Verification batch (this tree, this run, at the r4 HEAD):

- Builds: `go build ./...` → exit 0; `GOOS=windows GOARCH=amd64 go build ./...` → exit 0.
- `go test ./internal/cli -run 'TestCodexAudit|TestCodexGateContainment|TestMCPGateExposure|TestAuditMulti' -count=1` → `ok github.com/modu-ai/moai-adk/internal/cli 94.966s` (the existing containment family stays green).
- `go test ./internal/factorymsg -count=1` (full) → `ok github.com/modu-ai/moai-adk/internal/factorymsg 67.584s` (the M5 pin included).
- `go test ./internal/hook -run 'TestFactoryHook|TestFactoryMessages|TestFactory' -count=1` → `ok github.com/modu-ai/moai-adk/internal/hook 47.927s`.
- `go test ./internal/cli -run 'TestCodexTask|TestGLMTask' -count=1` → `ok github.com/modu-ai/moai-adk/internal/cli 16.489s` (the session-context neighbors).
- `go vet ./internal/cli ./internal/hook ./internal/factorymsg` → clean (exit 0).
- `golangci-lint run internal/cli/... internal/hook/... internal/factorymsg/...` → exactly the 1 pre-existing errcheck (`internal/cli/mcp_server_call_log.go:123`), ZERO new.
- Full-suite note: the repository-wide test verdict is owned by the CI run on the project's integration branch and is PENDING at report time (lane-local targeted selectors only, per the standing batch surface).
- Push state: nothing pushed (the lane owns push/landing).

Authored-By-Agent: manager-develop

### Sync-audit r5 delta repair round — landing re-audit (card t1632, branch WT-3-2-0-2)

Commit: `fix(t1632): lane-57 sync-audit r5 repairs — child cwd equals listing root, mode-aware untracked digest` (this round). Base: HEAD `71e857caf5` (the r4 repairs commit). Audit receipt: `rcpt-9de5e42dc597603ce97b0cf7` (STATIC review — it ran NO tests; each finding was verified honestly against the code before repair, both proved REAL).

| Finding | Disposition | Command (verbatim) | Observed output |
|---------|-------------|--------------------|-----------------|
| F1 [P1] the MCP listing's root and the child process's cwd diverge (internal/cli/mcp_codex.go:970, openCodexSessionResolved) | REAL-FIXED — after r4 the containment LISTING runs at the request's root (B), but the spawned app-server child inherited the SERVER process's cwd (A): MCP servers registered only in A's config.toml escaped the disable list and reached the audit child. Fixed per the auditor: the `codexSessionRunner` seam now carries `dir`; `realCodexSessionRunner` sets `cmd.Dir` to it; `openCodexSessionResolved` passes the SAME `root` the listing used (request `cwd` when named, else `projectDirResolver()` — unchanged r4 anchor, still fail-closed when unresolvable). The r4 pin TestCodexAudit_NoProjectRootStillContainsChild and the r1 pin TestRealCodexSessionRunnerScrubRetainsAuditLegMarker stay green; the ORCA env family stays green. A real-child pin (TestRealCodexSessionRunnerSpawnsChildAtWorkingDir) judges the directory the child actually ran at, not a field value | `go test ./internal/cli -run 'TestCodexAudit_SessionChildCwdMatchesListingRoot' -count=1` | RED→GREEN: `--- FAIL ... session child dir = "", want the containment listing root "/private/var/.../001" — a child at another cwd resolves a different config.toml and escapes the disable list (l57-r5 finding 1)` → `ok internal/cli` |
| F2 [P2] the untracked-file digest ignores file mode and symlink-ness (internal/auditreceipt/store.go:243, TreeState) | REAL-FIXED — the digest bound path + os.ReadFile(content) only: chmod +x on an untracked script, or a same-content symlink replacement, left the digest unchanged and a pre-change PASS receipt read fresh. Fixed in the SAME untrackedContentPaths/TreeState machinery (no fork, r2 whitespace and -z history preserved): the digest now binds os.Lstat's mode bits (file type + permissions) and, for a symlink, the Readlink target, beside the bytes. Two REDs: chmod +x moves the digest; a same-content symlink replacement (target OUTSIDE the tree, so the untracked path set itself is unchanged) moves the digest | `go test ./internal/auditreceipt -run 'TestTreeState_ChmodMovesUntrackedDigest|TestTreeState_SymlinkReplacementMovesUntrackedDigest' -count=1` | RED→GREEN: `--- FAIL ... digest unchanged after chmod +x on an untracked script — a pre-change PASS receipt reads fresh across an executable file (l57-r5 finding 2)` and `--- FAIL ... digest unchanged after a same-content symlink replacement — a pre-change PASS receipt reads fresh (l57-r5 finding 2)` → `ok internal/auditreceipt` |

E8 RED evidence (each pinning test run against the pre-fix code, all REAL):

- F1 RED (`go test ./internal/cli -run 'TestCodexAudit_SessionChildCwdMatchesListingRoot' -count=1`, exit 1 — seam extended, wiring not yet passing `root`):
  `--- FAIL: TestCodexAudit_SessionChildCwdMatchesListingRoot (0.17s)` / `codex_audit_sandbox_test.go:69: session child dir = "", want the containment listing root "/private/var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/TestCodexAudit_SessionChildCwdMatchesListingRoot1501530198/001" — a child at another cwd resolves a different config.toml and escapes the disable list (l57-r5 finding 1)`
- F2 RED-a (`go test ./internal/auditreceipt -run 'TestTreeState_ChmodMovesUntrackedDigest|TestTreeState_SymlinkReplacementMovesUntrackedDigest' -count=1`, exit 1):
  `--- FAIL: TestTreeState_ChmodMovesUntrackedDigest (0.28s)` / `receipt_integrity_content_test.go:149: digest unchanged after chmod +x on an untracked script — a pre-change PASS receipt reads fresh across an executable file (l57-r5 finding 2)`
- F2 RED-b (same command, exit 1):
  `--- FAIL: TestTreeState_SymlinkReplacementMovesUntrackedDigest (0.32s)` / `receipt_integrity_content_test.go:189: digest unchanged after a same-content symlink replacement — a pre-change PASS receipt reads fresh (l57-r5 finding 2)`

Verification batch (this tree, this run, at the r5 HEAD):

- Builds: `go build ./...` → exit 0; `GOOS=windows GOARCH=amd64 go build ./...` → exit 0.
- `go test ./internal/cli -run 'TestCodexAudit|TestCodexTask|TestCodexGateContainment' -count=1` → `ok github.com/modu-ai/moai-adk/internal/cli 82.866s`.
- `go test ./internal/cli -run 'TestRealCodexSessionRunnerScrubRetainsAuditLegMarker|TestCodexAudit_NoProjectRootStillContainsChild|TestCodexAudit_SessionChildCwdMatchesListingRoot|TestRealCodexSessionRunnerSpawnsChildAtWorkingDir' -count=1` → `ok github.com/modu-ai/moai-adk/internal/cli 2.493s` (the r1+r4 pins plus the two r5 tests).
- `go test ./internal/cli -run 'TestCodexChild|OrcaEnv|TestRealCodexConn' -count=1` → `ok github.com/modu-ai/moai-adk/internal/cli 1.586s` (the ORCA env family — its direct `realCodexSessionRunner.start` call site was a mechanical 4-arg update).
- `go test ./internal/auditreceipt -count=1` (full) → `ok github.com/modu-ai/moai-adk/internal/auditreceipt 7.376s`.
- `go vet ./internal/cli ./internal/auditreceipt` → clean (exit 0).
- `golangci-lint run internal/cli/... internal/auditreceipt/...` → exactly the 1 pre-existing errcheck (`internal/cli/mcp_server_call_log.go:123`), ZERO new.
- Full-suite note: the repository-wide test verdict is owned by the CI run on the project's integration branch and is PENDING at report time (lane-local targeted selectors only, per the standing batch surface).
- Push state: nothing pushed (the lane owns push/landing).

Authored-By-Agent: manager-develop

### Sync-audit r6 delta repair round — landing re-audit (card t1632, branch WT-3-2-0-2)

Commit: `fix(t1632): lane-57 sync-audit r6 repairs — symlink-safe tree digest, codex job family request root` (this round). Base: HEAD `5ea8db97d2` (the r5 repairs commit). Audit receipt: `rcpt-1260bde363fb2233f8315440` (STATIC review — it ran NO tests; each finding was verified honestly against the code before repair, both proved REAL).

| Finding | Disposition | Command (verbatim) | Observed output |
|---------|-------------|--------------------|-----------------|
| F1 [P2] TreeState opens untracked symlinks — ReadFile follows the link (internal/auditreceipt/store.go:254) | REAL-FIXED — the r5 mode/symlink binding still called os.ReadFile on every untracked path, so it read the link's TARGET: a dangling untracked symlink failed the read and failed the WHOLE TreeState (BindCaptureFailed on every receipt from that audit — all unusable), and a FIFO/device-backed target could block the read. Fixed per the auditor: when Lstat says symlink, TreeState binds the Readlink target beside the Lstat metadata and NEVER opens the link (non-symlink regular files keep the existing content read). The r5 pins TestTreeState_SymlinkReplacementMovesUntrackedDigest and TestTreeState_ChmodMovesUntrackedDigest stay green (full ./internal/auditreceipt suite run). New RED pin TestTreeState_DanglingSymlinkComputesAndBindsLinkTarget: a dangling untracked symlink in the fixture tree — pre-fix TreeState errors, post-fix it computes and re-pointing the link moves the digest | `go test ./internal/auditreceipt -run TestTreeState_DanglingSymlinkComputesAndBindsLinkTarget -count=1` | RED→GREEN: `--- FAIL: TestTreeState_DanglingSymlinkComputesAndBindsLinkTarget (0.21s)` / `TreeState failed on a dangling untracked symlink — ReadFile followed the link and bound the target's content (l57-r6 finding 1): open /var/folders/.../dangling: no such file or directory` → `ok internal/auditreceipt` |
| F2 [P2] the codex_job family ignores the request's project_root (internal/cli/codex_job_control.go:204) | REAL-FIXED — loadCodexJobFor read the registry at projectDirResolver() (the server tree) even when the caller named registered worktree B: a cancel naming B could cancel a job from A while _root claimed B. Fixed as the glm_task/glm_job_control twin (l57-r1 F2): the PRODUCER (codex_task.go:385) now roots the job registry at the VALIDATED request root (workRoot from codexAuditValidateRoot — the write opt-in stays serving-project policy per its own contract), loadCodexJobFor takes the request-resolved root, codex_job_status/result/cancel pass it, and codex_job_cancel's cancelled-status update hits the same request-root registry (the now-unused currentCodexJobRegistry server-root helper is removed). Two REDs mirroring the GLM pair (codex_task_root_test.go, fixture = newUntrackedFixture P/W): producer — a background job with project_root=W writes under W, not P; consumer — a record seeded under W is READ by codex_job_status naming W (seeded directly, because a pre-fix producer AND a pre-fix consumer are consistently wrong at P and a producer-fed pair would stay green). The r7/r8 isolation pin TestCodexTask_BackgroundJobNeverTouchesTheServingState stays green (startCtxBoundBackgroundJob already passed project_root = its temp registry root). 24 pre-existing tests pinned the OLD contract (record at projectDirResolver while project_root=thisRepoRoot) and were ALIGNED to the new contract — each now passes project_root = its git-inited fixture root with withCodexServingDir stubbed to the same repository (the startCtxBoundBackgroundJob isolation pattern); no assertion was weakened, and the aligned run leaves ZERO records under the real serving tree (measured: pre-fix runs wrote 12 debris records under the worktree's .moai/state/codex-jobs; the aligned run wrote none; the debris was removed) | `go test ./internal/cli -run 'TestCodexTaskBackgroundJobRecordLandsUnderNamedRoot|TestCodexJobStatusReadsTheNamedRoot' -count=1` | RED→GREEN: `--- FAIL: TestCodexTaskBackgroundJobRecordLandsUnderNamedRoot (0.18s)` / `background codex_task with project_root=/private/var/.../W wrote no job record under the named root: codex job not found: job-20261021T155940Z-008dc2b6` and `--- FAIL: TestCodexJobStatusReadsTheNamedRoot (0.14s)` / `codex_job_status could not read the job record at the root the request names: codex_job_status: no job record for id job-20261021T155940Z-3afc5d11` → `ok internal/cli` |

E8 RED evidence (each pinning test run against the pre-fix code, all REAL):

- F1 RED (`go test ./internal/auditreceipt -run TestTreeState_DanglingSymlinkComputesAndBindsLinkTarget -count=1`, exit 1):
  `--- FAIL: TestTreeState_DanglingSymlinkComputesAndBindsLinkTarget (0.21s)` / `receipt_integrity_content_test.go:213: TreeState failed on a dangling untracked symlink — ReadFile followed the link and bound the target's content (l57-r6 finding 1): open /var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/TestTreeState_DanglingSymlinkComputesAndBindsLinkTarget2313504257/001/dangling: no such file or directory`
- F2 RED-a producer (`go test ./internal/cli -run 'TestCodexTaskBackgroundJobRecordLandsUnderNamedRoot|TestCodexJobStatusReadsTheNamedRoot' -count=1`, exit 1):
  `--- FAIL: TestCodexTaskBackgroundJobRecordLandsUnderNamedRoot (0.18s)` / `codex_task_root_test.go:39: background codex_task with project_root=/private/var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/TestCodexTaskBackgroundJobRecordLandsUnderNamedRoot820031411/002/W wrote no job record under the named root: codex job not found: job-20261021T155940Z-008dc2b6`
- F2 RED-b consumer (same command, exit 1):
  `--- FAIL: TestCodexJobStatusReadsTheNamedRoot (0.14s)` / `codex_task_root_test.go:67: codex_job_status could not read the job record at the root the request names: codex_job_status: no job record for id job-20261021T155940Z-3afc5d11`

Verification batch (this tree, this run, at the r6 base HEAD with the repair sources in the working tree — the commit below adds only progress.md to these exact sources):

- Builds: `go build ./...` → exit 0; `GOOS=windows GOARCH=amd64 go build ./...` → exit 0.
- `go test ./internal/auditreceipt -count=1` (full) → `ok github.com/modu-ai/moai-adk/internal/auditreceipt 4.742s` (the r5 mode/chmod/symlink pins stay green).
- `go test ./internal/cli -run 'TestCodexJob|TestCodexTask|TestCodexRoleAudit' -count=1` → `ok github.com/modu-ai/moai-adk/internal/cli 17.554s`.
- `go test ./internal/cli -run 'TestGLMJob|TestGLMTask' -count=1` → `ok github.com/modu-ai/moai-adk/internal/cli 8.395s` (the glm twin family stays green).
- `go vet ./internal/cli ./internal/auditreceipt` → clean (exit 0).
- `golangci-lint run internal/cli/... internal/auditreceipt/...` → exactly the 1 pre-existing errcheck (`internal/cli/mcp_server_call_log.go:123`), ZERO new.
- Fixture note: the new contract's create/read parity required aligning 24 pre-existing tests (codex_task_test, codex_task_resume_scope_test, codex_task_failcause_test, codex_protocol_liveness_test, codex_task_process_context_test, codex_job_control_test, mcp_served_model_test) — every aligned test now names project_root = its own git-inited fixture root, mirroring the production contract; the invalid-work_key refusal and foreground no-registry paths are untouched.
- Full-suite note: the repository-wide test verdict is owned by the CI run on the project's integration branch and is PENDING at report time (lane-local targeted selectors only, per the standing batch surface).
- Push state: nothing pushed (the lane owns push/landing).

Authored-By-Agent: manager-develop

#### r6b follow-through — glm_job_cancel update path follows the request root

The r6 report flagged a residual asymmetry in the glm twin: glm_job_cancel READ the record from the request root (l57-r1) but wrote the cancelled-status update through currentGLMJobRegistry() (server-rooted, glm_job_control.go:191) — a cancel naming worktree B updated a server-tree record while _root claimed B. Fixed as the codex F2 mirror: the update hits `newGLMJobRegistry(root)` (the request-resolved root) and the now-unused currentGLMJobRegistry helper is removed.

- RED pin (TestGLMJobCancelUpdatesTheNamedRoot, glm_task_root_test.go — job record under fixture W, cancel naming W): `--- FAIL: TestGLMJobCancelUpdatesTheNamedRoot (0.50s)` / `result status = "running", want "cancelled" — the cancelled-status update must land in the registry the request names` and `record at the named root = "failed", want "cancelled" — the cancel's status update wrote somewhere other than the request root (REQ-MN-004)`. The record read "failed" because the server-rooted update never landed the cancelled guard at W, so the revoked call's goroutine wrote its own terminal status — exactly the un-cancelled-cancel shape the finding predicts.
- GREEN: `go test ./internal/cli -run 'TestGLMJobCancelUpdatesTheNamedRoot' -count=1` → `ok github.com/modu-ai/moai-adk/internal/cli`.
- Batch (this tree, this run, r6b sources): `go build ./...` exit 0; `GOOS=windows GOARCH=amd64 go build ./...` exit 0; `go test ./internal/cli -run 'TestGLMJob|TestGLMTask' -count=1` → ok 10.296s; `go test ./internal/cli -run 'TestCodexJob|TestCodexTask|TestCodexRoleAudit' -count=1` → ok 15.477s; `go test ./internal/auditreceipt -count=1` → ok 3.981s; `go vet ./internal/cli` → clean; `golangci-lint run internal/cli/... internal/auditreceipt/...` → exactly the 1 pre-existing errcheck (mcp_server_call_log.go:123), zero new; the six r6+r6b pin selectors (TreeState dangling-symlink, codex producer/consumer root, glm producer/consumer root, glm cancel update) in one run → ok.
- Full-suite note: repository-wide verdict owned by the integration-branch CI run, PENDING at report time. Push state: nothing pushed.

Authored-By-Agent: manager-develop

### Sync-audit r7 delta repair round — landing re-audit (card t1632, branch WT-3-2-0-2)

Commit: `fix(t1632): lane-57 sync-audit r7 repairs — framed tree digest, GLM review base, NUL worktree parse` (this round). Base: HEAD `a83bd0d698` (the r6b commit). Audit receipt: `rcpt-6f3ebb7971c3cae6081a1026` (STATIC review — it ran NO tests; each finding was verified honestly against the code before repair, all three proved REAL, all P2).

| Finding | Disposition | Command (verbatim) | Observed output |
|---------|-------------|--------------------|-----------------|
| r7-1 [P2] the untracked digest's concatenation is ambiguous without length framing (internal/auditreceipt/store.go TreeState) | REAL-FIXED — without a per-file length, a file's CONTENT bytes could masquerade as the framing of a phantom entry: the tree {ua: "", ub: H+"X"} (H = ub's exact framing bytes `\x00untracked:ub\x00mode:<m>`) hashed IDENTICALLY to {ua: H, ub: "X"} — same porcelain rows, same HEAD, same diff, a content change without digest movement, so a pre-change PASS receipt stayed fresh across a different tree. Fixed per the auditor: the content is LENGTH-FRAMED (`\x00len:<n>` before the bytes), making every entry boundary exact; the symlink `link:` target needs no length (a link target cannot contain NUL, and the next framing always begins with one — noted in-source). The framing change moves EVERY pre-existing digest (expected and correct — receipts are per-audit-tree, not persisted across versions); no fixture hard-codes a digest literal (the TestTreeState_* family pins are behavioral, all green on the full ./internal/auditreceipt suite). RED pin TestTreeState_FramingMasqueradeMovesUntrackedDigest: the auditor's two-state masquerade shape in one fixture tree | `go test ./internal/auditreceipt -run 'TestTreeState_FramingMasqueradeMovesUntrackedDigest' -count=1` | RED→GREEN: `--- FAIL: TestTreeState_FramingMasqueradeMovesUntrackedDigest (0.47s)` / `receipt_integrity_content_test.go:315: digest unchanged across the framing-masquerade content change — unframed concatenation let ua's content bytes read as ub's entry framing, so a pre-change PASS receipt stays fresh (l57-r7 finding 1)` → `ok internal/auditreceipt` |
| r7-2 [P2] the GLM leg of audit_multi(baseBranch) does not carry ReviewBase (internal/cli/mcp_convergence.go performGLMAudit) | REAL-FIXED — performGLMAudit collected the diff at the resolved base (collectReviewDiffCtx resolves it internally and DISCARDS it) but left ReviewBase empty, so the multi-record logged "working-tree (target not applied)" — a FALSE target report for a baseBranch audit that WAS applied. The GLM twin of the codex REQ-MN-019/t1426 chain. Fixed as the codex pattern: a baseBranch leg resolves its base ONCE (a delta_base on ctx replaces it, exactly as collectReviewDiffCtx did), measures the diff from that exact merge base via collectReviewDiffAt, and reports the same value as review_base (`out.ReviewBase = base.String()`); the non-baseBranch path is unchanged. RED pin TestPerformGLMAudit_BaseBranchCarriesReviewBase (both arms: resolved base + delta scope) | `go test ./internal/cli -run 'TestPerformGLMAudit_BaseBranchCarriesReviewBase' -count=1` | RED→GREEN: `--- FAIL: TestPerformGLMAudit_BaseBranchCarriesReviewBase/resolved_base_names_the_base_it_measured` / `second_review_effective_scope_test.go:47: baseBranch GLM leg carried no review_base — the record reads "working-tree (target not applied)" for an audit that was applied at the base (l57-r7 finding 2)` and `--- FAIL: .../delta_scope_names_the_delta_base_it_measured` / `review_base = "", want it to name the delta base 8e27d84127ea4427ab22e63896edbf841b8f3b9c the diff was measured from` → `ok internal/cli` |
| r7-3 [P2] the worktree list parse is line-based, not NUL-based (internal/cli/mcp_project_root.go:216) | REAL-FIXED — `git worktree list --porcelain` parsed by lines truncated a registered worktree whose path contains a newline, and confinement refused it as unregistered. The RED run revealed the SAME defect at two more parse sites on the SAME acceptance chain, each refusing the newline-path worktree BEFORE the named site could be reached — fixed all three as one cascade attributable to the finding: (a) auditreceipt parseWorktreePorcelain (storeroot.go) → `--porcelain -z` + NUL-field parse, grammar MEASURED on a live fixture (every FIELD ends with NUL, the entry boundary is an empty field, a newline inside a path stays verbatim inside its own field); IdentifyPrimaryCheckout passes -z; the pin fixtures in storeroot_test.go updated to -z shapes with provenance plus a newline-path entry arm; (b) mcp_worktree_root.go validateLinkedWorktreeRoot → `rev-parse --show-toplevel` has NO -z form, so the multi-line reading is ANCHORED: accepted only when git answered exactly `canonical` (the value being validated); (c) the named site confineToServerRepository → -z + NUL-field split. Sibling same-class line-based consumers OUTSIDE the confinement chain (factory_card.go, session_worktree*.go, integration.go, launcher.go, todo_issuance.go, update_worktree_migration.go, worktree/clean.go, worktree/done.go, memory_drain.go, codex_audit_launch.go:330, hook/cwd_changed_relocate.go) are NOT touched — outside this receipt's scope, named here for the next audit. RED pin TestMCPRootConfinementNewlinePathWorktree | `go test ./internal/cli -run 'TestMCPRootConfinementNewlinePathWorktree' -count=1` | RED→GREEN: `--- FAIL: TestMCPRootConfinementNewlinePathWorktree (0.33s)` / `mcp_root_contract_test.go:150: registered worktree at a newline-containing path rejected — the worktree list parse truncated the path at the newline (l57-r7 finding 3): project_root "/var/folders/.../002-wt\nnewline" has no .moai directory and is not an accepted linked worktree: git could not inspect it as a worktree (unexpected output)` → `ok internal/cli` (post-fix the newline-path worktree passes confinement and the canonical path is returned) |

E8 RED evidence (each pinning test run against the pre-fix code, all REAL):

- r7-1 RED (`go test ./internal/auditreceipt -run 'TestTreeState_FramingMasqueradeMovesUntrackedDigest' -count=1`, exit 1):
  `--- FAIL: TestTreeState_FramingMasqueradeMovesUntrackedDigest (0.47s)` / `receipt_integrity_content_test.go:315: digest unchanged across the framing-masquerade content change — unframed concatenation let ua's content bytes read as ub's entry framing, so a pre-change PASS receipt stays fresh (l57-r7 finding 1)`
- r7-2 RED (`go test ./internal/cli -run 'TestPerformGLMAudit_BaseBranchCarriesReviewBase' -count=1`, exit 1):
  `--- FAIL: TestPerformGLMAudit_BaseBranchCarriesReviewBase/resolved_base_names_the_base_it_measured (0.07s)` / `second_review_effective_scope_test.go:47: baseBranch GLM leg carried no review_base — the record reads "working-tree (target not applied)" for an audit that was applied at the base (l57-r7 finding 2)`; `--- FAIL: TestPerformGLMAudit_BaseBranchCarriesReviewBase/delta_scope_names_the_delta_base_it_measured (0.02s)` / `review_base = "", want it to name the delta base 8e27d84127ea4427ab22e63896edbf841b8f3b9c the diff was measured from`
- r7-3 RED (`go test ./internal/cli -run 'TestMCPRootConfinementNewlinePathWorktree' -count=1`, exit 1):
  `--- FAIL: TestMCPRootConfinementNewlinePathWorktree (0.33s)` / `mcp_root_contract_test.go:150: registered worktree at a newline-containing path rejected — the worktree list parse truncated the path at the newline (l57-r7 finding 3): project_root "/var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/TestMCPRootConfinementNewlinePathWorktree730482080/002-wt\nnewline" has no .moai directory and is not an accepted linked worktree: git could not inspect it as a worktree (unexpected output)`

Verification batch (this tree, this run, with the repair sources in the working tree — the commit below adds only progress.md to these exact sources):

- Builds: `go build ./...` → exit 0; `GOOS=windows GOARCH=amd64 go build ./...` → exit 0.
- `go test ./internal/auditreceipt -count=1` (full — the digest pins live here) → `ok github.com/modu-ai/moai-adk/internal/auditreceipt 3.948s` (the r2 whitespace / r5 mode+symlink / r6 dangling-link pins all stay green over the moved digests).
- `go test ./internal/cli -run 'TestMCPRoot|TestGLMAudit|TestAuditMulti|TestConverge' -count=1` → `ok github.com/modu-ai/moai-adk/internal/cli 92.689s`.
- `go test ./internal/cli -run 'TestCodexAudit' -count=1` → `ok github.com/modu-ai/moai-adk/internal/cli 30.083s` (the codex ReviewBase twin stays green).
- Change-scoped extra: `go test ./internal/hook -run 'TestAuditReceipt|TestAuditReceiptGuard' -count=1` → `ok github.com/modu-ai/moai-adk/internal/hook 1.410s` (internal/hook consumes the auditreceipt TreeState/StoreRoot seams the framing and -z changes move).
- `go vet ./internal/auditreceipt ./internal/cli` → clean (exit 0).
- `golangci-lint run ./internal/auditreceipt/... ./internal/cli/...` → exactly the 1 pre-existing errcheck (`internal/cli/mcp_server_call_log.go:123`), ZERO new.
- Full-suite note: the repository-wide test verdict is owned by the CI run on the project's integration branch and is PENDING at report time (lane-local targeted selectors only, per the standing batch surface).
- Push state: nothing pushed (the lane owns push/landing).

Authored-By-Agent: manager-develop

### Sync-audit r8 delta repair round — landing re-audit (card t1632, branch WT-3-2-0-2)

Commit: `fix(t1632): lane-57 sync-audit r8 repairs — task/gate containment split, nested-repo and submodule digest binding` (this round). Base: HEAD `984e34be61` (the r7 commit). Audit receipt: `rcpt-1945aac06ebeb8675eee395c` (STATIC review — it ran NO tests; each finding was verified honestly against the code before repair, all three proved REAL, all P2).

| Finding | Disposition | Command (verbatim) | Observed output |
|---------|-------------|--------------------|-----------------|
| r8-1 [P2] the gate-only MCP containment leaks into ordinary delegation tasks (internal/cli/mcp_codex.go:979) | REAL-FIXED (refines the r4 resolution — recorded per the standing discipline) — the two static audits pulled OPPOSITE directions on the same shared spawn path: r4 made containment UNCONDITIONAL on openCodexSessionResolved after an auditor found audit spawns naming no project_root escaping it; the round-7 auditor correctly observed that handleCodexTask (ordinary `codex_task`) rides the SAME path, so every ordinary task lost ALL configured external MCP servers regardless of allow_write (and the session-level read-only sandbox — REQ-MN-024 scopes that to the audit/review/convergence legs, while a task's sandbox is its own per-turn sandboxPolicy). Fixed by SCOPE, not by reverting to r4's defect: openCodexSessionResolved gains a `contain bool` — openCodexSessionOn (the plain-task path, sole production caller codex_task.go) passes false and spawns `[app-server]` with NO containment args and NO session-level sandbox override; runCodexReviewRPCResolved (the codex_audit / codex_review / review-gate family) passes true and keeps the r4-unconditional containment + read-only sandbox verbatim, anchored at cwd → projectDirResolver with the fail-closed no-root refusal. The r4 pins stay green with one honest premise adjustment recorded: TestCodexAudit_NoProjectRootStillContainsChild's failure message now reads "unconditional ON THE GATE/AUDIT PATH" (its assertion is unchanged and passing — the audit child stays contained whatever the request carries). RED pin TestCodexTask_SessionChildNotContained asserts an ordinary task child's argv carries NEITHER the foreign-server disable args NOR sandbox_mode=read-only, with a foreign server (context7) in the stubbed listing to make the leak visible | `go test ./internal/cli -run 'TestCodexTask_SessionChildNotContained' -count=1` | RED→GREEN: `--- FAIL: TestCodexTask_SessionChildNotContained (0.22s)` / `codex_audit_sandbox_test.go:99: ordinary codex_task child argv = [app-server -c sandbox_mode=read-only -c mcp_servers.context7.url="https://mcp-disabled.invalid" -c mcp_servers.context7.enabled=false], want NO foreign-server disable args — the containment belongs to gate/audit spawns (l57-r8 finding 1)` (+ the same argv line for the sandbox assertion) → `ok internal/cli` |
| r8-2 [P2] an untracked nested git repository breaks the whole capture (internal/auditreceipt/store.go:272) | REAL-FIXED — `git ls-files --others --exclude-standard -z` returns the DIRECTORY entry (`nested/`) for an untracked nested repo (measured on a live fixture: the only row was `nested/`); os.ReadFile on a directory fails (`is a directory`) → TreeState errors → BindCaptureFailed → every receipt from that audit unusable. Same class as the r6 symlink fix. Fixed per the auditor: only a REGULAR file is content-read — a directory (or any other non-regular entry a content read cannot take; a FIFO would block) binds its PATH + Lstat metadata and moves on; the r7 length framing keeps such an entry distinct from a same-framing empty file. RED pin TestTreeState_UntrackedNestedRepoComputes | `go test ./internal/auditreceipt -run 'TestTreeState_UntrackedNestedRepoComputes' -count=1` | RED→GREEN: `--- FAIL: TestTreeState_UntrackedNestedRepoComputes (0.12s)` / `receipt_integrity_content_test.go:335: TreeState failed on an untracked nested git repo — the directory entry was content-read (l57-r8 finding 2): read /var/folders/.../001/nested: is a directory` → `--- PASS: TestTreeState_UntrackedNestedRepoComputes (0.15s)` |
| r8-3 [P2] dirty-submodule inner edits do not move the digest (internal/auditreceipt/store.go:237) | REAL-FIXED — with a registered gitlink whose inner worktree is dirty, the parent diff shows the same `Subproject commit …-dirty` line and the status row + HEAD are unchanged, so the digest was STABLE across inner re-edits → stale PASS. Fixed per the auditor, scoped to gitlink entries only (mode 160000, via a new `git ls-files --stage -z` read; no arbitrary recursion): each submodule binds its inner HEAD + inner porcelain rows + the inner tracked diff's actual bytes — the first RED run showed the auditor's `status --porcelain` suggestion ALONE is the same row-identity defect one level down (a re-edit emits the identical ` M lib.go` row), so the inner diff bytes are bound, the same content fidelity the parent got in card-review P1-1. An inner UNTRACKED file is outside this binding, as it is outside the parent-side `-dirty` signal itself (named in-source). An absent/uninitialized submodule worktree binds its index gitlink instead — deterministic, no capture failure. RED pin TestTreeState_SubmoduleInnerEditMovesDigest (plumbing-form submodule: update-index --cacheinfo 160000 + clone, sidestepping the file-protocol gate) | `go test ./internal/auditreceipt -run 'TestTreeState_SubmoduleInnerEditMovesDigest' -count=1` | RED→GREEN: `--- FAIL: TestTreeState_SubmoduleInnerEditMovesDigest (0.33s)` / `receipt_integrity_content_test.go:403: digest unchanged across a submodule inner re-edit — the parent diff carries the same `Subproject commit …-dirty` line both times, so a pre-change PASS receipt stays fresh (l57-r8 finding 3)` → `--- PASS: TestTreeState_SubmoduleInnerEditMovesDigest (0.42s)` (the intermediate inner-porcelain-only fix stayed RED with the same message — measured, then upgraded to the inner diff bytes) |

E8 RED evidence (each pinning test run against the pre-fix code, all REAL):

- r8-1 RED (`go test ./internal/cli -run 'TestCodexTask_SessionChildNotContained' -count=1`, exit 1):
  `--- FAIL: TestCodexTask_SessionChildNotContained (0.22s)` / `codex_audit_sandbox_test.go:99: ordinary codex_task child argv = [app-server -c sandbox_mode=read-only -c mcp_servers.context7.url="https://mcp-disabled.invalid" -c mcp_servers.context7.enabled=false], want NO foreign-server disable args — the containment belongs to gate/audit spawns (l57-r8 finding 1)`; `codex_audit_sandbox_test.go:102: ordinary codex_task child argv = [app-server -c sandbox_mode=read-only -c mcp_servers.context7.url="https://mcp-disabled.invalid" -c mcp_servers.context7.enabled=false], want NO session-level read-only sandbox — the task's per-turn sandboxPolicy governs it (l57-r8 finding 1)`
- r8-2 RED (`go test ./internal/auditreceipt -run 'TestTreeState_UntrackedNestedRepoComputes|TestTreeState_SubmoduleInnerEditMovesDigest' -count=1`, exit 1):
  `--- FAIL: TestTreeState_UntrackedNestedRepoComputes (0.12s)` / `receipt_integrity_content_test.go:335: TreeState failed on an untracked nested git repo — the directory entry was content-read (l57-r8 finding 2): read /var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/TestTreeState_UntrackedNestedRepoComputes570839049/001/nested: is a directory`
- r8-3 RED (same run, exit 1):
  `--- FAIL: TestTreeState_SubmoduleInnerEditMovesDigest (0.33s)` / `receipt_integrity_content_test.go:403: digest unchanged across a submodule inner re-edit — the parent diff carries the same ` + "`Subproject commit …-dirty`" + ` line both times, so a pre-change PASS receipt stays fresh (l57-r8 finding 3)`

Verification batch (this tree, this run, at the final working-tree state the commit below records — re-run AFTER the last source edit so every output is verbatim at final HEAD):

- Builds: `go build ./...` → `BUILD_OK` (exit 0); `GOOS=windows GOARCH=amd64 go build ./...` → `WIN_BUILD_OK` (exit 0).
- `go test ./internal/cli -run 'TestCodexTask|TestCodexAudit|TestCodexGateContainment|TestCodexRoleAudit' -count=1` → `ok github.com/modu-ai/moai-adk/internal/cli 49.100s`.
- `go test ./internal/auditreceipt -count=1` (full) → `ok github.com/modu-ai/moai-adk/internal/auditreceipt 5.900s`.
- r4/r5 pins kept green explicitly (the selector above does not match their names): `go test ./internal/cli -run 'TestRealCodexSessionRunnerScrubRetainsAuditLegMarker|TestRealCodexSessionRunnerSpawnsChildAtWorkingDir|TestCodexJobs|TestCodexSession' -count=1 -v` → all `--- PASS` (11 RUN lines, 0 FAIL) — env scrub (incl. the Orca pane drop), child-at-working-dir mechanics, and the session/job families unchanged.
- Adjusted-pin + new-pin rerun: `go test ./internal/cli -run 'TestCodexAudit_NoProjectRootStillContainsChild|TestCodexTask_SessionChildNotContained' -count=1 -v` → both `--- PASS`.
- `go vet ./internal/cli ./internal/auditreceipt` → clean (`VET_OK`, exit 0).
- `golangci-lint run ./internal/cli/... ./internal/auditreceipt/...` → exactly the 1 pre-existing errcheck (`internal/cli/mcp_server_call_log.go:123`), ZERO new.
- Full-suite note: the repository-wide test verdict is owned by the CI run on the project's integration branch and is PENDING at report time (lane-local targeted selectors only, per the standing batch surface).
- Push state: nothing pushed (the lane owns push/landing).

Authored-By-Agent: manager-develop

## §E.3 Run-phase Audit-Ready Signal

run_audit_ready: true
run_complete_at: 2026-10-15
run_commit_sha: 73c84c3ed
run_status: complete (all 6 milestones M1–M6 implemented; branch WT-3-2-0-2, nothing pushed — merged-local landing is the leader's batch)
run_commit_range: M1 ebba213f0..6199704240 (refusal contract + identity) · M2 6199704240..0d53d09fa (Part A CLI identity absorption + Part B REQ-MN-004~008) · M3 (gate exposure 20-tool, containment, elicitation guard, write mutex, inbox counting, AC-MN-003 22/22 measured) · M4 c13bc05c0 (env allowlist, binary pin, 429 relabel, staleness, job lifetime, stamp, call log) · M5 db9fb0e7b (generation migration, token stability, count-only hook, line-boundary cut, Stop scope) · M6 09e4cd60a→cf3546753 (REQ-MN-021/022 could-not-run split + not-converged + retry, REQ-MN-023 narrow decode, REQ-MN-020 receipt integrity + audit_cache evidence + auditor model, REQ-MN-019/024/025 target honesty + sandbox + material marks + state-root seam)
ac_status: AC-MN-001..025 all PASS — AC-MN-001..005 (M1/M2 structured refusals, identity, root contract, catalogue, schema, wire), AC-MN-006..008 (M2 root confinement, catalogue hard source, schema lint, wire honesty), AC-MN-009..011 (M3 gate containment, elicitation, mutex/inbox), AC-MN-003 22/22 gates exposed measured at M3 (numerator 22, denominator 22, per §E.2 M3 block), AC-MN-012..015 (M4), AC-MN-016..018 (M5), AC-MN-019..025 (M6, §E.2 M6 matrix)
gaps_pre_existing_two_classes: (1) TestCodexReviewGatePrimaryCheckoutSkip/_TreeScopeRequestShapeUnchanged/_StaleEnvKeepsTreeScope — macOS /var→/private symlink spelling in three gate pins; (2) TestAuditMulti_CodexLegDeadline_FailsClosedNamed/_UnconfiguredFailsOpen/TestAuditMulti_ModelMulti_CodexUnavailable_FailsClosedNamed — containment listing executes the stubbed binary before the timeout path. BOTH reproduced at HEAD 554053cf9 with the M6 sources temporarily reverted (§E.2 M6 note 7) — they predate this SPEC's run phase; leader disposition requested
debts_remaining: none open from this SPEC — N3 (serial-slot instance red at M1 RED) was the TDD RED-stage observation itself, resolved by M1 GREEN (AC-MN-001 PASS, §E.2 M1 matrix); N4 (receipt reuse-refusal + restart persistence) was implemented in M6 (head/digest binding + state-move refusal, §E.2 M6 matrix); N5 (AC-MN-012 "without stalling" quantification) was closed at M3 (the automatic decline lands within 50 ms, measured — §E.2 M3 row)

## §E.4 Sync-phase Audit-Ready Signal

sync_audit_ready: true
sync_complete_at: 2026-10-15
sync_commit_sha: "8864f9f56"
sync_status: complete (3-phase close — spec.md frontmatter in-progress → completed on this commit; the placeholder is backfilled with the real SHA in the follow-up commit, since a commit cannot cite its own hash)
sync_commit_range: HEAD 226c387e3 (run close) → the sync commit (CHANGELOG entry, §E.4 signal, frontmatter terminal transition, MX anchors ×3)
sync_artifacts: CHANGELOG.md [Unreleased]/Added first bullet · spec.md frontmatter (status+updated only) · progress.md §E.4 (this section) · MX tag changes: @MX:ANCHOR added ×3 — toolRefusal (18 callers, internal/cli/mcp_refusal.go), resolveRequiredToolProjectRoot (8 callers, internal/cli/mcp_project_root.go), resolveValidatedToolProjectRoot (3 callers, same file); zero tags removed; pre-existing tags in internal/auditreceipt/store.go, internal/factorymsg/store.go, internal/cli/codex_review_gate.go verified intact
changelog_entry_position: [Unreleased] → ### Added, first bullet
b12_self_test_a: PASS — pre-emission grep `grep -c 'SPEC-MCP-NEUTRALITY-001' CHANGELOG.md` → 0 (halt condition clear)
b12_self_test_b: PASS — B12 AC counter on acceptance.md (tier L source): live=25 excluded=0 ambiguous=0; entry references 25 ACs (AC-MN-001..025)
b12_self_test_c: PASS — all cited paths verified by ls: internal/cli/{mcp_refusal,mcp_project_root,mcp_server_call_log,codex_review_gate,codex_stop_chain,mcp_review_material,mcp_audit_multi_record,mcp_convergence}.go, internal/runtime/audit_cache_state.go, internal/auditreceipt/store.go, internal/factorymsg/store.go, internal/hook/{failure_observer,factory_messages}.go, internal/codexwiring/configtoml.go, internal/mcp/catalog.go, internal/config/envkeys.go, internal/template/templates/AGENTS.md.tmpl
frontmatter_status_transitions: in-progress → completed (this sync commit, manager-docs, Authored-By-Agent: manager-docs trailer)
canary_compliance_check: not applicable — this SPEC defines no forward-looking policy whose sync tests would self-apply

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
