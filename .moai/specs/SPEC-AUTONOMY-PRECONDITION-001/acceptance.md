# SPEC-AUTONOMY-PRECONDITION-001 — Acceptance criteria

Every criterion is binary: it names fixtures, the processed call, and an observable outcome. The
`RED-now` note records why the criterion fails today; `Green path` names the milestone that turns
it green.

**No criterion here depends on work this SPEC does not own** — the defect (finding N5) that caused
the split. That is a claim, so §G carries an **ownership column** naming, per criterion, the surface
its Then reads and who creates it; the blanket sentence does not stand in for that enumeration. The
iter-1 audit was right that v0.1.1's blanket claim was false: `AC-AP-013` read a template block A1
creates. It is repaired by moving the surface (spec.md §C.10, §H.5), not by softening the sentence.
R6 is answered (spec.md §C.5) and R5 was resolved by *removing* scope: receipt issuance is
A3's, so the receipt-path criteria are excluded rather than carried (spec.md §C.4, §H.4).
`AC-AP-011 [RETIRED]` and `AC-AP-012 [RETIRED]` were retired at v0.1.1 and their ids are **not**
re-used; the gap in the numbering is deliberate. Fifteen criteria: AC-AP-001..010, AC-AP-013..017.

**Every criterion carries a limb that goes RED when the guard is absent or fails open**, and the limb
is named in the criterion rather than left to the reader. The pattern is `AC-AP-008`'s: assert
something the guard must *produce* — a deny, a lease record, a written verdict — in the same test as
any absence limb, so an unwired or fail-open guard cannot satisfy the criterion by doing nothing. The
iter-1 audit found three criteria made only of absence assertions; `AC-AP-012 [RETIRED]` is retired, and
`AC-AP-002` and `AC-AP-006` are repaired below. Each criterion states its discriminator on a
`Discriminator:` line.

The `[RETIRED]` and `[REF]` tokens above and below are the AC-counter's reserved markers, not
decoration, and their **placement inside the code span is load-bearing**: the marker must sit
adjacent to the identifier with only spaces or tabs between, and a closing backtick breaks that
adjacency — the token goes **inside** the code span, not after it. Measured with the canonical
counter on this file at v0.1.2: `live=15 excluded=3 ambiguous=0`, exit 0 — 15 live criteria, 3 excluded
(2 retired here, 1 referenced from another SPEC).

Marking only *some* occurrences of an identifier is worse than marking none: the counter resolves
per identifier and halts with `AMBIGUOUS` rather than emitting a count, so every occurrence of a
retired or referenced identifier in this file carries its token — including the ones in this
paragraph's own neighbourhood, which is how that halt was observed here before it was fixed.

Every criterion is evaluated against fixtures. Per A1's required ordering (spec.md §E C4) the
`push-develop` action is not activated until A4 lands, and a signed contract confers no autonomy
until A3 — so no criterion here asserts that any autonomy was granted.

## §A — Push serializer (REQ-AP-001, REQ-AP-002, REQ-AP-007)

- **AC-AP-001** (maps REQ-AP-001) — **Given** `workflow.autonomy.mode: contract`, a
  `moai contract show --json` fixture reporting `actions` containing `push-develop` with
  `push_requires_lease: true`, and two live
  sessions A and B, **When** A's `git push origin develop` is processed at PreToolUse and then B's
  `git push origin develop` is processed while A holds the `push-develop` lease within its bound,
  **Then** A's call is allowed with its hook output byte-identical to the no-guard baseline and the
  lease record names A as holder, and B's call is denied with a reason whose first token is
  `PUSH_SERIALIZATION_VIOLATION:` and which names A.
  - Discriminator: the Then requires a **produced** deny carrying the sentinel and a **written** lease
    record naming A. An absent or fail-open serializer produces neither, so both limbs fail.
  - Test: `TestPushSerializationDeniesSecondPush` in `internal/hook`.
  - RED-now: no push serializer and no `PUSH_SERIALIZATION_VIOLATION` sentinel exist
    (spec.md §C.1). Green path: M1.
- **AC-AP-002** (maps REQ-AP-001, REQ-AP-007) — **Given** the AC-AP-001 two-session setup,
  **When** the same two pushes are processed in each of four off conditions — (a)
  `workflow.autonomy.mode: guided`, (b) mode `contract` with a contract whose `actions` omits
  `push-develop`, (c) mode `contract` with `push-develop` present but
  `push_requires_lease: false`, and (d) mode `contract` with the `show --json` field absent
  altogether (the state of the tree until A1 supplies it, spec.md §C.6) — and then once more in a
  fifth, **armed** condition (e) with all three activation conditions satisfied, **Then** in each of
  (a)-(d) both pushes are allowed, no `push-develop` lease record is written, and no escalation record
  is written by either call; **and in (e), in the same test, B's push is denied with the
  `PUSH_SERIALIZATION_VIOLATION:` sentinel and the lease record names A.**
  - Discriminator: condition (e) is the armed positive control, and it is what makes the criterion
    non-vacuous. A guard that is never wired, or that fails open on every path, satisfies (a)-(d)
    trivially and **fails (e)** — so the criterion is RED in exactly the state v0.1.1 would have
    reported as green. The four off conditions alone assert only absence and could not distinguish
    "correctly inactive" from "not present".
  - Test: `TestPushSerializationInactiveWhenNotArmed` in `internal/hook`.
  - RED-now: the guard does not exist, so neither its inactivity nor condition (e) can be observed.
    Green path: M1.
- **AC-AP-003** (maps REQ-AP-002) — **Given** A holds the `push-develop` lease after an admitted
  push, **When** that push's PostToolUse reports a non-zero exit, **Then** the lease record is
  released, and a subsequent `git push origin develop` from B is admitted and records B as holder.
  - Discriminator: the Then requires B's subsequent push to be **admitted with B recorded as holder**
    — a written record. With no serializer no lease record is ever written, so "records B as holder"
    fails; the release limb alone would have been absence-only.
  - Test: `TestPushLeaseReleasedOnFailedPush` in `internal/hook`.
  - RED-now: no release path exists. Green path: M1.
- **AC-AP-004** (maps REQ-AP-002) — **Given** three separate fixtures in which (a) A's owning
  session is gone while the lease is held, (b) A's declared bound has elapsed, and (c) the lease
  record is present but undecodable, **When** B's `git push origin develop` is processed in each,
  **Then** in (a) and (b) B's push is admitted and the new lease record names A as displaced, and
  in (c) B's push is allowed with exactly one audit line recording the fail-open and no lease
  written.
  - Discriminator: limbs (a) and (b) require a **new lease record naming A as displaced**, which only
    a live serializer writes. Limb (c) is absence-only by nature (fail-open) and rides in the same test
    as (a) and (b), so an unwired guard fails the test rather than passing on (c).
  - Test: `TestPushLeaseReclaimAndFailOpen` in `internal/hook`.
  - RED-now: no reclaim or fail-open path exists. Green path: M1.

## §B — Human-path sign deny: recognition and fail-closed (REQ-AP-003, REQ-AP-004)

- **AC-AP-005** (maps REQ-AP-003) — **Given** any value of `workflow.autonomy.mode` and a session with
  **no** `MOAI_FACTORY_ROLE` set in its environment, **When** each of
  `moai contract sign SPEC-X-001`, `moai contract sign --signer human SPEC-X-001`,
  `'moai' contract "sign"`, `FOO=1 moai contract sign`, `env FOO=1 moai contract sign`,
  `command moai contract sign`, `exec moai contract sign`, `~/go/bin/moai contract sign`,
  `./bin/moai contract sign`, `moai --no-color contract sign`, and `sh -c 'moai contract sign'` is
  processed at PreToolUse, **Then** each is denied with a reason whose first token is
  `CONTRACT_SIGN_AGENT_VIOLATION:`, and the count of denied cases equals the count of cases supplied
  (no case silently skipped).
  - The absent-variable Given is load-bearing, not decoration: every case here is on the **human**
    signing path (no `--signer`, or `--signer human` — spec.md §C.2), and REQ-AP-003 denies it in every
    session. A guard that gated this path on `MOAI_FACTORY_ROLE` would fail this criterion rather than
    pass it vacuously.
  - Discriminator: every limb requires a **produced** deny. An absent or fail-open guard denies
    nothing and fails on the first case; the equal-count limb additionally catches a guard that denies
    some shapes and silently skips others.
  - Test: `TestContractSignAgentInvocationDenied` in `internal/hook`.
  - RED-now: no sign guard and no `CONTRACT_SIGN_AGENT_VIOLATION` sentinel exist
    (spec.md §C.2). Green path: M2.
- **AC-AP-006** (maps REQ-AP-003) — **Given** the AC-AP-005 session, **When** the positive
  controls `moai contract show SPEC-X-001`, `moai contract verify SPEC-X-001`,
  `echo "moai contract sign"`, `git log --grep sign`, and `git commit -m "sign the contract"` are
  processed, **Then** each is allowed with hook output byte-identical to the no-guard baseline and no
  audit line is written for any of them; **and in the same test** `moai contract sign SPEC-X-001` is
  denied with the `CONTRACT_SIGN_AGENT_VIOLATION:` sentinel, and the counts of allowed and denied cases
  each equal the counts supplied.
  - Discriminator: the trailing deny case is the armed control. A guard that is never wired, or that
    fails open, allows all five controls **and** allows the sign case, failing the test — so passing
    here means the controls survived a guard that was demonstrably active, which is the claim. Without
    that limb the criterion asserted only absence and an unwired guard would have passed it.
  - Test: `TestContractSignPositiveControlsAllowed` in `internal/hook`.
  - RED-now: the guard does not exist, so no control can be shown to survive it. Green path: M2.
- **AC-AP-007** (maps REQ-AP-004) — **Given** any value of `workflow.autonomy.mode` and a session with
  no `MOAI_FACTORY_ROLE`, **When** each of `$(which moai) contract sign`, `$M contract sign`,
  `eval "moai contract sign"`, and `sh -c 'bash -c "moai contract sign"'` is processed, **Then**
  each is denied with a reason starting `CONTRACT_SIGN_AGENT_VIOLATION:` and carrying the literal
  token `unclassified`.
  - Every case carries `sign`, so REQ-AP-003 is in force and the fail-closed rule applies in every
    session. The `decide` counterparts are **not** here: with no role marker no deny rule is in force
    for `decide`, so an unclassifiable `decide` is allowed — asserted as the allow arm of AC-AP-016, and
    as a deny in AC-AP-015's marked session.
  - Discriminator: produced denies, plus the literal `unclassified` token — a guard that denied by
    falling through to the classified path would pass the deny limb and fail the token limb.
  - Test: `TestContractSignUnclassifiedDeniedClosed` in `internal/hook`.
  - RED-now: no fail-closed path exists. Green path: M2.
- **AC-AP-008** (maps REQ-AP-004; spec.md §C.2, §C.3) — **Given** an installed binary for which
  `moai contract --help` exits non-zero with `Unknown command "contract"`, **When**
  `moai contract sign SPEC-X-001` is processed in a session with no role marker, and
  `moai contract decide SPEC-X-001` is processed in a session whose `MOAI_FACTORY_ROLE` is `worker`,
  **Then** each is denied with `CONTRACT_SIGN_AGENT_VIOLATION:` — the denial does not depend on the
  verb being implemented, and the test asserts the unimplemented precondition before asserting the
  deny, so the case cannot pass vacuously by the command simply failing. This is the criterion that
  makes the `decide` deny evaluable today, while A3 has not yet defined the verb (spec.md §F O1).
  - Discriminator: the precondition assertion runs **first**, which turns "the verb does not exist"
    from an assumption into a measured value; and PreToolUse runs **before** execution, so no command
    failure can produce the deny. This is the construction the other criteria's discriminators copy.
  - Test: `TestContractSignDeniedBeforeVerbExists` in `internal/hook`.
  - RED-now: the guard does not exist. Green path: M2.

## §B.1 — Role-scoped deny: the non-interactive sign path and `decide` (REQ-AP-011, REQ-AP-012)

- **AC-AP-015** (maps REQ-AP-011) — **Given** a session whose environment sets `MOAI_FACTORY_ROLE` to
  `worker` (set by the test itself, per REQ-AP-012) and any value of `workflow.autonomy.mode`, **When**
  each of `moai contract sign --signer llm --receipt /tmp/r.json`,
  `moai contract sign --signer llm+jev --receipt /tmp/r.json`, `moai contract decide SPEC-X-001`,
  `'moai' contract "decide"`, `sudo moai contract decide`, and `eval "moai contract decide"` is
  processed, **Then** each is denied with a reason whose first token is
  `CONTRACT_SIGN_AGENT_VIOLATION:`, the `eval` case additionally carries the literal token
  `unclassified`, and the count of denied cases equals the count supplied.
  - Discriminator: produced denies in a marked session. A guard that ignores the marker allows all six
    and fails immediately.
  - Test: `TestContractRoleScopedDenyUnderWorkerMarker` in `internal/hook`.
  - RED-now: neither the guard nor the `MOAI_FACTORY_ROLE` constant exists —
    `grep -rn 'MOAI_FACTORY_ROLE' internal cmd pkg` returns 0 rows (spec.md §C.7). Green path: M2.
- **AC-AP-016** (maps REQ-AP-011) — **Given** the AC-AP-015 command set, **When** it is processed
  twice — once in a session where `MOAI_FACTORY_ROLE` is **unset**, and once where it is set to a value
  other than `worker` — **Then** in both runs every one of those calls is allowed with hook output
  byte-identical to the no-guard baseline and no audit line written; **and in the same test**
  `moai contract sign SPEC-X-001` and `moai contract sign --signer human SPEC-X-001` are denied with
  the `CONTRACT_SIGN_AGENT_VIOLATION:` sentinel in both runs.
  - Discriminator: the two trailing human-path denies are the armed control, and they are the whole
    reason this criterion is not vacuous. An unwired or fail-open guard allows the role-scoped calls
    **and** allows the human-path calls, failing the test. Without them, "the lead may still `decide`"
    would be satisfied by a guard that does not exist — the exact shape the iter-1 audit found in
    `AC-AP-002` and `AC-AP-006`.
  - This criterion is the one that would have caught the withdrawn outright deny: under v0.1.1's rule
    every call here was denied, so the lead's own `decide` path was denied (spec.md §C.3).
  - Test: `TestContractRoleScopedAllowWithoutWorkerMarker` in `internal/hook`.
  - RED-now: the guard does not exist, so no allow can be shown to survive it. Green path: M2.
- **AC-AP-017** (maps REQ-AP-012, REQ-AP-009) — **Given** the pre-change tree, in which
  `grep -rn 'MOAI_FACTORY_ROLE' internal cmd pkg` returns exactly 0 rows, **When** the change lands,
  **Then** `internal/config` exports a name constant whose value is exactly the string
  `MOAI_FACTORY_ROLE` and a role-value constant whose value is exactly the string `worker`; the guard
  and its tests reference those constants and not a repeated literal (asserted by
  `grep -c '"MOAI_FACTORY_ROLE"' internal/hook` returning 0); **and** the guard reads no environment
  variable outside the closed set {this constant} ∪ {`CLAUDECODE`, `CLAUDE_CODE_SESSION_ID`} — A1's
  marker set (`25283ebf8:…/design.md:416-426`) — asserted by enumerating the guard's `os.Getenv` /
  `os.LookupEnv` call sites and their arguments; and AC-AP-015 passes with the variable set by the test
  alone, with no launcher change and no dependency on card t1240.
  - Discriminator: the pre-change 0-row measurement is asserted first, so "this SPEC introduced the
    constant" is measured rather than assumed; and the final limb is a **cross-criterion** check — if
    the constants existed but the guard did not read them, AC-AP-015 fails.
  - Test: `TestFactoryRoleEnvConstant` in `internal/config` plus the literal-free grep in
    `internal/hook`.
  - RED-now: neither constant exists. Green path: M2.
  - Recorded with it, not asserted by it: `worker` is the **canonical** CLI role spelling
    (`internal/cli/factory.go:59`); `agent` (`:64`) is its retired pre-rename alias and is **not** an
    accepted value here — the operator resolved spec.md §F O5 to `worker` for exactly that reason. This
    criterion still does not measure what card t1240 stamps: if t1240 stamps anything other than
    `worker`, it passes while REQ-AP-011 denies nothing in production — spec.md §F O5 and §E C7 carry
    that, and no criterion here claims otherwise.

## §C — Wrapper coverage (REQ-AP-005; finding N4)

- **AC-AP-009** (maps REQ-AP-005) — **Given** any value of `workflow.autonomy.mode`, **When** each
  of `script -q /dev/null moai contract sign`, `timeout 30 moai contract sign`,
  `sudo moai contract sign`, `sudo -n moai contract sign`, `stdbuf -oL moai contract sign`,
  `nice -n 10 moai contract sign`, `nohup moai contract sign`, and
  `xargs -n1 moai contract sign` is processed, **Then** each is denied with a reason starting
  `CONTRACT_SIGN_AGENT_VIOLATION:`, and each wrapper's own options are shown to have been skipped
  rather than read as the program word (the reason names `moai` as the resolved program).
  - Discriminator: produced denies, plus the resolved-program limb — a guard that read a wrapper's
    own option as the program word would fail to name `moai` and fail that limb even where it denied.
  - Recorded, not asserted: `sudo` and `env -i` drop the environment of the **child**, not of the hook
    process, and the guard reads the calling session's environment at PreToolUse before any wrapper
    runs (spec.md §C.7) — so no case here is affected by that stripping. Whether every harness
    populates the hook process's environment identically to the session's is **not** measured by this
    criterion and is recorded as a Gap in progress.md §E.1.
  - Test: `TestContractSignWrapperBypassesDenied` in `internal/hook`.
  - RED-now: no wrapper set exists; finding N4 recorded this as an uncovered hole. Green path: M2.
- **AC-AP-010** (maps REQ-AP-005) — **Given** a wrapper name absent from the closed list — the
  fixture uses `chrt 0 moai contract sign` — **When** it is processed, **Then** it is denied as
  unclassified under REQ-AP-004 rather than allowed, and a mutant that removes the
  unknown-wrapper branch is shown to flip this case from denied to allowed while AC-AP-009 stays
  green, so the criterion is not satisfied by AC-AP-009's list alone.
  - **The mutation clause is a requirement on run-phase, not a measurement already taken.** No guard
    exists, so there is no branch to remove and no test to run: the mutant has **not** been built and
    the flip has **not** been observed. The assertion is first observed at run-phase M2, where the
    mutant is actually constructed and the flip recorded in progress.md §E.2. The criterion is not
    green until that observation exists.
  - Discriminator: the primary limb is a produced deny; the mutation limb is what stops the criterion
    from being satisfied by the closed list alone, and it is checked by building the mutant, not by
    reading the design.
  - Test: `TestContractSignUnknownWrapperFailsClosed` in `internal/hook`.
  - RED-now: neither the list nor its unknown-wrapper branch exists. Green path: M2.

## §D — No receipt exemption on the human path (REQ-AP-003; the withdrawn REQ-AP-006)

`AC-AP-011 [RETIRED]` and `AC-AP-012 [RETIRED]` (receipt path allowed / forged receipt) are
**retired**, and the transferred `AC-AE-025 [REF]` is **not carried** — receipt issuance is A3's
(spec.md §C.4, §H.4). Their ids are not re-used.

The positive obligation that replaces them is already carried by **AC-AP-005**, whose fixture set
contains no exempt form: since REQ-AP-003 grants no exemption on the human path, the absence of a
receipt branch is asserted by that deny being unconditional rather than by a criterion of its own. A
run-phase implementation that added a receipt exemption to the human path would fail AC-AP-005 on the
exempted shape.

**REQ-AP-011 is not this allowance returning.** It gates *who may invoke* the non-interactive signing
path A1 defines (`--signer llm` / `llm+jev`), on the role marker, and recognizes no receipt: it never
inspects `--receipt`, so nothing here depends on A3 issuing one. AC-AP-015 and AC-AP-016 pass a
receipt path as an opaque argument and assert only the deny and the allow.

## §E — Documentation of the mode-independent deny (REQ-AP-010; finding N8)

- **AC-AP-013** (maps REQ-AP-010) — **Given** the `paths:`-scoped rule file this change creates for
  the contract-sign guard and its `internal/template/templates/` mirror, **When** both are read after
  the change, **Then** each states that the contract-sign deny is mode-independent and active under
  `guided`, and states that the "nothing changes under `guided`" promise of
  SPEC-AUTONOMY-ESCALATION-001 REQ-AE-001 is scoped in words to the escalation detector; **and** the
  two files are byte-identical apart from the front-matter `paths:` value; **and** a grep over both
  finds no card id, SPEC id, internal date, or commit SHA.
  - **Retargeted at v0.1.2, which is the repair of the iter-1 audit's D1.** v0.1.1 read A1's template
    `workflow.yaml` autonomy block, which does not exist in this tree — nor does any rule text stating
    contract-mode behavior (spec.md §C.10 measures both absent). Both surfaces this criterion now reads
    are **created by this change**, so it is evaluable at M3 with no dependency on A1 or any other
    track. The two limbs whose surfaces are not this SPEC's are transferred, not assumed (spec.md §H.5).
  - Discriminator: every limb reads a file this change **creates**. If the change does not create it,
    the read fails and the criterion is RED — there is no state in which "nothing to check" reads as a
    pass. The mirror-parity limb additionally catches a local-only edit that skips the Template-First
    mirror.
  - Test: `TestContractSignGuardRuleDocumentsModeIndependence` in `internal/template` (both files read
    from disk, parity and neutrality asserted together).
  - RED-now: neither the rule file nor its mirror exists. Green path: M3, with no precondition outside
    this SPEC.

## §F — Mission-validator projection (REQ-AP-008; O9 / N7 / A1 R8)

- **AC-AP-014** (maps REQ-AP-008) — **Given** a signed-valid contract fixture whose `ownership.write`
  contains `internal/foo/**`, whose `budget.operations` is positive, and which carries the required
  top-level `card` field, and a proposed action, **When** the projection is asked for a verdict,
  **Then** all five of the following hold:
  1. a verdict is produced at all — the projected contract is **not** refused `incomplete_contract`,
     which requires the projection to have populated `Approved` and a positive
     `ResourceLimits.MaxOperations` (`internal/mission/contract.go:55-61`, `:82`);
  2. the verdict for each of the five mission refusal reasons — `mission_not_running`,
     `mission_mismatch`, `policy_mismatch`, `stale_snapshot`, `expired_decision` — is produced by
     `mission.ValidateMissionDecision`, asserted by a fixture whose only difference is a field that
     mission alone rejects;
  3. a target `internal/foo/x.go` is **inside** the projected scope, and a target
     `internal/foobar/x.go` is **outside** it — the `/**`-to-prefix translation neither narrows nor
     widens (`targetInsideScope`, `internal/mission/policy.go:112-118`);
  4. a contract whose `ownership.write` contains an inner-wildcard glob (`internal/*/x.go`), and
     separately one carrying a field that is neither projected nor on the deliberately-not-projected
     list, each return a fail-closed error **naming** the offending field or glob — while the `card`
     field, which is on that list, does **not** trigger one;
  5. the exported signatures of `internal/mission` are unchanged from the pre-change baseline.
  - **Limb 2's "not by a local re-implementation" is a requirement on run-phase, not a measurement
    already taken.** No projection exists, so nothing has been measured: the differentiating fixture
    has not been built and no verdict has been observed to come from mission rather than from a copy.
    That assertion is first observed at run-phase M4 and recorded in progress.md §E.2. The same applies
    to limb 5's baseline, which is captured at M4's start, not now.
  - Discriminator: limbs 1, 2 and 3 all require a **produced** verdict with a specific value, so a
    projection that is absent, or that fails closed on everything, fails them. Limb 4's fail-closed
    assertions ride in the same test, so the criterion cannot be satisfied by a projection that only
    ever refuses.
  - Test: `TestContractProjectsOntoMissionValidator` in `internal/contract` (or the package the
    projection lands in; a rename is recorded in progress.md §E.2).
  - RED-now: no projection exists. `ValidateMissionDecision` exists at
    `internal/mission/policy.go:200` (spec.md §C.8), and `targetInsideScope`'s exact-or-prefix
    containment is measured at `:112-118` (spec.md §C.11). Green path: M4.

## §G — Traceability and per-criterion ownership

Every live requirement has at least one criterion, and every criterion maps at least one requirement.
Three criteria map two requirements each — `AC-AP-002` (REQ-AP-001 + REQ-AP-007), `AC-AP-005`
(REQ-AP-003 + REQ-AP-009 through its absent-variable Given), `AC-AP-017` (REQ-AP-012 + REQ-AP-009) —
and each does so as two named limbs of one test rather than as a vague overlap. The
**Surface** and **Created by** columns are the enumeration that replaces the blanket "no criterion
depends on unowned work" claim: a reader checks the column rather than trusting the sentence. The
iter-1 audit's D1 was that the sentence was true of every row but one, and the exception was invisible
because no such column existed.

| Requirement | Criteria | Surface the Then reads | Created by |
|---|---|---|---|
| REQ-AP-001 | AC-AP-001, AC-AP-002 | the `push-develop` lease record, and a `show --json` **fixture** | this SPEC (the record); the fixture stands in for A1's command, so no criterion waits on it |
| REQ-AP-002 | AC-AP-003, AC-AP-004 | the lease record and one audit line | this SPEC |
| REQ-AP-003 | AC-AP-005, AC-AP-006 | the guard's PreToolUse decision on a command line | this SPEC |
| REQ-AP-004 | AC-AP-007, AC-AP-008, AC-AP-010 | the same decision, unclassified branch | this SPEC |
| REQ-AP-005 | AC-AP-009, AC-AP-010 | the same decision, wrapper branch | this SPEC |
| REQ-AP-006 | **withdrawn at v0.1.1** — no criteria; id retired, not re-used | — | — |
| REQ-AP-007 | AC-AP-001, AC-AP-002 | absence of an escalation record; unchanged hook output | this SPEC |
| REQ-AP-008 | AC-AP-014 | `mission.ValidateMissionDecision` and the projection | `internal/mission` exists (spec.md §C.8); the projection is this SPEC's |
| REQ-AP-009 | AC-AP-005 (the absent-variable Given), AC-AP-017 (the closed env-read set) | the guard's decision without a role marker, and its enumerated `os.Getenv` call sites | this SPEC |
| REQ-AP-010 | AC-AP-013 | a `paths:`-scoped rule file and its template mirror | **this SPEC** — retargeted at v0.1.2; v0.1.1 read A1's template block (spec.md §C.10, §H.5) |
| REQ-AP-011 | AC-AP-015, AC-AP-016 | the guard's decision with and without the role marker | this SPEC (the guard) + REQ-AP-012 (the marker constant) |
| REQ-AP-012 | AC-AP-017 | the exported constants in `internal/config/envkeys.go` | this SPEC — which is what removes the dependency on card t1240 |

`AC-AP-017` maps two requirements and is listed under both, which is the one place a criterion carries
a second mapping: REQ-AP-012's limb is the constants' existence, REQ-AP-009's is the closed env-read
set. They are separate limbs of one test rather than two criteria, because both are assertions about
the same enumerated call sites.

**Eleven live requirements, fifteen criteria.** Every live requirement has ≥1 criterion; every
criterion maps at least one requirement (three map two, named above); the set difference is empty in
both directions — no live requirement without a criterion, no criterion without a live requirement.
`AC-AP-011 [RETIRED]` / `AC-AP-012 [RETIRED]` retired with REQ-AP-006.

**What this table establishes and what it does not.** It establishes id-set inclusion and, through the
two right-hand columns, that no criterion's Then reads a surface another track creates. It does **not**
establish that each criterion semantically verifies the requirement it maps — that judgment is a
reader's, and the table is no substitute for it.

## §H — Quality gates and Definition of Done

- Change-scoped tests pass with a non-empty swept count, and every test named above appears as a
  `--- PASS:` line: `go test ./internal/hook/... ./internal/config/... ./internal/contract/... ./internal/mission/... -count=1 -v`.
- `GOOS=windows GOARCH=amd64 go build ./...` exits 0.
- `golangci-lint run` reports no new issue against the pre-change baseline measured on this tree.
- Coverage of each package this SPEC adds code to is at least 85%.
- Subagent boundary: no `AskUserQuestion` reference in the changed packages outside tests.
- Template neutrality: the changed and added template text carries no card id, SPEC id, internal
  date, or SHA (AC-AP-013 final limb), and the rule file and its mirror are byte-identical apart from
  `paths:`.
- No escalation record is written by either component on any tested path (AC-AP-002).
- No receipt exemption exists on the guard's human path (the negative limb of AC-AP-005).
- The role-scoped rule allows a no-marker session's `decide` (AC-AP-016), and no code path reads the
  literal `"MOAI_FACTORY_ROLE"` instead of the constant (AC-AP-017).
- `internal/mission`'s exported surface is unchanged (AC-AP-014 limb 5).
- Every criterion's armed limb is present and observed: no criterion is reported green on absence
  assertions alone, and the two run-phase-first assertions (AC-AP-010's mutant, AC-AP-014's
  no-re-implementation and baseline limbs) are recorded in progress.md §E.2 with the command that
  produced them.
