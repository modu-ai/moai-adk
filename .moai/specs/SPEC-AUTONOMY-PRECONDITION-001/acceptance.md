# SPEC-AUTONOMY-PRECONDITION-001 — Acceptance criteria

Every criterion is binary: it names fixtures, the processed call, and an observable outcome. The
`RED-now` note records why the criterion fails today; `Green path` names the milestone that turns
it green. No criterion here depends on work this SPEC does not own — A1's R5 and R6 are answered
(spec.md §C.4, §C.5), which is the defect (finding N5) that caused the split.

## §A — Push serializer (REQ-AP-001, REQ-AP-002, REQ-AP-007)

- **AC-AP-001** (maps REQ-AP-001) — **Given** `workflow.autonomy.mode: contract`, a resolved
  contract whose `actions` contains `push-develop` with `push_requires_lease: true`, and two live
  sessions A and B, **When** A's `git push origin develop` is processed at PreToolUse and then B's
  `git push origin develop` is processed while A holds the `push-develop` lease within its bound,
  **Then** A's call is allowed with its hook output byte-identical to the no-guard baseline and the
  lease record names A as holder, and B's call is denied with a reason whose first token is
  `PUSH_SERIALIZATION_VIOLATION:` and which names A.
  - Test: `TestPushSerializationDeniesSecondPush` in `internal/hook`.
  - RED-now: no push serializer and no `PUSH_SERIALIZATION_VIOLATION` sentinel exist
    (spec.md §C.1). Green path: M1.
- **AC-AP-002** (maps REQ-AP-001, REQ-AP-007) — **Given** the AC-AP-001 two-session setup,
  **When** the same two pushes are processed in each of three off conditions — (a)
  `workflow.autonomy.mode: guided`, (b) mode `contract` with a contract whose `actions` omits
  `push-develop`, and (c) mode `contract` with `push-develop` present but
  `push_requires_lease: false` — **Then** in every condition both pushes are allowed, no
  `push-develop` lease record is written, and no escalation record is written by either call.
  - Test: `TestPushSerializationInactiveWhenNotArmed` in `internal/hook`.
  - RED-now: the guard does not exist, so its inactivity cannot be asserted against it.
    Green path: M1.
- **AC-AP-003** (maps REQ-AP-002) — **Given** A holds the `push-develop` lease after an admitted
  push, **When** that push's PostToolUse reports a non-zero exit, **Then** the lease record is
  released, and a subsequent `git push origin develop` from B is admitted and records B as holder.
  - Test: `TestPushLeaseReleasedOnFailedPush` in `internal/hook`.
  - RED-now: no release path exists. Green path: M1.
- **AC-AP-004** (maps REQ-AP-002) — **Given** three separate fixtures in which (a) A's owning
  session is gone while the lease is held, (b) A's declared bound has elapsed, and (c) the lease
  record is present but undecodable, **When** B's `git push origin develop` is processed in each,
  **Then** in (a) and (b) B's push is admitted and the new lease record names A as displaced, and
  in (c) B's push is allowed with exactly one audit line recording the fail-open and no lease
  written.
  - Test: `TestPushLeaseReclaimAndFailOpen` in `internal/hook`.
  - RED-now: no reclaim or fail-open path exists. Green path: M1.

## §B — Contract-sign guard: recognition and fail-closed (REQ-AP-003, REQ-AP-004)

- **AC-AP-005** (maps REQ-AP-003) — **Given** any value of `workflow.autonomy.mode`, **When** each
  of `moai contract sign SPEC-X-001`, `'moai' contract "sign"`, `FOO=1 moai contract sign`,
  `env FOO=1 moai contract sign`, `command moai contract sign`, `exec moai contract sign`,
  `~/go/bin/moai contract sign`, `./bin/moai contract sign`, `moai --no-color contract sign`, and
  `sh -c 'moai contract sign'` is processed at PreToolUse, **Then** each is denied with a reason
  whose first token is `CONTRACT_SIGN_AGENT_VIOLATION:`, and the count of denied cases equals the
  count of cases supplied (no case silently skipped).
  - Test: `TestContractSignAgentInvocationDenied` in `internal/hook`.
  - RED-now: no sign guard and no `CONTRACT_SIGN_AGENT_VIOLATION` sentinel exist
    (spec.md §C.2). Green path: M2.
- **AC-AP-006** (maps REQ-AP-003) — **Given** the AC-AP-005 fixture set, **When** the positive
  controls `moai contract show SPEC-X-001`, `moai contract verify SPEC-X-001`,
  `echo "moai contract sign"`, `git log --grep sign`, and `git commit -m "sign the contract"` are
  processed, **Then** each is allowed with hook output byte-identical to the no-guard baseline, and
  no audit line is written for any of them.
  - Test: `TestContractSignPositiveControlsAllowed` in `internal/hook`.
  - RED-now: the guard does not exist, so no control can be shown to survive it. Green path: M2.
- **AC-AP-007** (maps REQ-AP-004) — **Given** any value of `workflow.autonomy.mode`, **When** each
  of `$(which moai) contract sign`, `$M contract sign`, `eval "moai contract sign"`, and
  `sh -c 'bash -c "moai contract sign"'` is processed, **Then** each is denied with a reason
  starting `CONTRACT_SIGN_AGENT_VIOLATION:` and carrying the literal token `unclassified`.
  - Test: `TestContractSignUnclassifiedDeniedClosed` in `internal/hook`.
  - RED-now: no fail-closed path exists. Green path: M2.
- **AC-AP-008** (maps REQ-AP-004; spec.md §C.2) — **Given** an installed binary for which
  `moai contract --help` exits non-zero with `Unknown command "contract"`, **When**
  `moai contract sign SPEC-X-001` is processed at PreToolUse, **Then** it is denied with
  `CONTRACT_SIGN_AGENT_VIOLATION:` — the denial does not depend on the verb being implemented, and
  the test asserts the unimplemented precondition before asserting the deny, so the case cannot
  pass vacuously by the command simply failing.
  - Test: `TestContractSignDeniedBeforeVerbExists` in `internal/hook`.
  - RED-now: the guard does not exist. Green path: M2.

## §C — Wrapper coverage (REQ-AP-005; finding N4)

- **AC-AP-009** (maps REQ-AP-005) — **Given** any value of `workflow.autonomy.mode`, **When** each
  of `script -q /dev/null moai contract sign`, `timeout 30 moai contract sign`,
  `sudo moai contract sign`, `sudo -n moai contract sign`, `stdbuf -oL moai contract sign`,
  `nice -n 10 moai contract sign`, `nohup moai contract sign`, and
  `xargs -n1 moai contract sign` is processed, **Then** each is denied with a reason starting
  `CONTRACT_SIGN_AGENT_VIOLATION:`, and each wrapper's own options are shown to have been skipped
  rather than read as the program word (the reason names `moai` as the resolved program).
  - Test: `TestContractSignWrapperBypassesDenied` in `internal/hook`.
  - RED-now: no wrapper set exists; finding N4 recorded this as an uncovered hole. Green path: M2.
- **AC-AP-010** (maps REQ-AP-005) — **Given** a wrapper name absent from the closed list — the
  fixture uses `chrt 0 moai contract sign` — **When** it is processed, **Then** it is denied as
  unclassified under REQ-AP-004 rather than allowed, and a mutant that removes the
  unknown-wrapper branch is shown to flip this case from denied to allowed while AC-AP-009 stays
  green, so the criterion is not satisfied by AC-AP-009's list alone.
  - Test: `TestContractSignUnknownWrapperFailsClosed` in `internal/hook`.
  - RED-now: neither the list nor its unknown-wrapper branch exists. Green path: M2.

## §D — Receipt path (REQ-AP-006; A1's answered R5)

- **AC-AP-011** (maps REQ-AP-006) — **Given** a SPEC directory containing
  `.moai/specs/SPEC-X-001/kickoff-receipt.json`, **When**
  `moai contract sign SPEC-X-001 --signer llm+jev --receipt .moai/specs/SPEC-X-001/kickoff-receipt.json`
  is processed, **Then** it is allowed and exactly one audit line naming that receipt path is
  written; **and** when the same command is processed with the receipt file absent, and separately
  with `--receipt` resolving to a path other than the fixed name, **Then** both are denied with
  `CONTRACT_SIGN_AGENT_VIOLATION:`.
  - Test: `TestContractSignReceiptPathAllowed` in `internal/hook`.
  - RED-now: no receipt branch exists. **Not** blocked on A1: the path shape is defined at
    `WT-contract-schema:.moai/specs/SPEC-AUTONOMY-CONTRACT-001/design.md` § Kickoff Receipt
    (spec.md §C.4). Green path: M2.
- **AC-AP-012** (maps REQ-AP-006) — **Given** a present, well-formed receipt file whose *contents*
  are forged (a `contract_sha256` that matches nothing), **When** the receipt-path invocation is
  processed, **Then** the guard allows it and writes no validation verdict of its own — the
  content verdict belongs to A1's validator (REQ-CONTRACT-023) — and the test asserts the absence
  of any guard-emitted validation outcome, so the guard is shown not to have silently taken on
  A1's job.
  - Test: `TestContractSignGuardDoesNotValidateReceipt` in `internal/hook`.
  - RED-now: no receipt branch exists. Green path: M2.

## §E — Documentation of the mode-independent deny (REQ-AP-010; finding N8)

- **AC-AP-013** (maps REQ-AP-010) — **Given** the template `workflow.yaml` autonomy block and the
  rule text that states contract-mode behavior, **When** both are read after the change, **Then**
  each states that the contract-sign deny is active under `guided`, and the
  "nothing changes under `guided`" statement is scoped in words to the escalation detector; **and**
  a grep over the changed template text finds no card id, SPEC id, internal date, or commit SHA.
  - Test: `TestAutonomyGuidedPromiseScopedToDetector` in `internal/config` (template text
    assertions) plus the template-neutrality grep in `internal/template`.
  - RED-now: the template autonomy block is introduced by A1 and carries no such statement.
    Green path: M3, after A1's block lands.

## §F — Mission-validator projection (REQ-AP-008; O9 / N7)

- **AC-AP-014** (maps REQ-AP-008) — **Given** a signed contract fixture and a proposed action,
  **When** the projection is asked for a verdict, **Then** the verdict for each of the five mission
  refusal reasons — `mission_not_running`, `mission_mismatch`, `policy_mismatch`, `stale_snapshot`,
  `expired_decision` — is produced by `mission.ValidateMissionDecision` and not by a local
  re-implementation (asserted by a fixture whose only difference is a field that mission alone
  rejects); **and** given a contract carrying a field with no mission counterpart, the projection
  returns a fail-closed error naming that field, and the exported signatures of
  `internal/mission` are unchanged from the pre-change baseline.
  - Test: `TestContractProjectsOntoMissionValidator` in `internal/contract` (or the package the
    projection lands in; a rename is recorded in progress.md §E.2).
  - RED-now: no projection exists. `ValidateMissionDecision` exists at
    `internal/mission/policy.go:200` (spec.md §C.7). Green path: M4.

## §G — Traceability

Every requirement has at least one criterion and every criterion maps one requirement.

| Requirement | Criteria |
|---|---|
| REQ-AP-001 | AC-AP-001, AC-AP-002 |
| REQ-AP-002 | AC-AP-003, AC-AP-004 |
| REQ-AP-003 | AC-AP-005, AC-AP-006 |
| REQ-AP-004 | AC-AP-007, AC-AP-008, AC-AP-010 |
| REQ-AP-005 | AC-AP-009, AC-AP-010 |
| REQ-AP-006 | AC-AP-011, AC-AP-012 |
| REQ-AP-007 | AC-AP-001, AC-AP-002 |
| REQ-AP-008 | AC-AP-014 |
| REQ-AP-009 | AC-AP-005 (the tool-call boundary case), AC-AP-013 |
| REQ-AP-010 | AC-AP-013 |

## §H — Quality gates and Definition of Done

- Change-scoped tests pass with a non-empty swept count, and every test named above appears as a
  `--- PASS:` line: `go test ./internal/hook/... ./internal/config/... ./internal/contract/... ./internal/mission/... -count=1 -v`.
- `GOOS=windows GOARCH=amd64 go build ./...` exits 0.
- `golangci-lint run` reports no new issue against the pre-change baseline measured on this tree.
- Coverage of each package this SPEC adds code to is at least 85%.
- Subagent boundary: no `AskUserQuestion` reference in the changed packages outside tests.
- Template neutrality: the changed template text carries no card id, SPEC id, internal date, or
  SHA (AC-AP-013 second limb).
- No escalation record is written by either component on any tested path (AC-AP-002).
- `internal/mission`'s exported surface is unchanged (AC-AP-014 second limb).
