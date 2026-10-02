# SPEC-JEV-AUTO-EXCEPTION-001 — Progress

## §E.1 Plan-phase Audit-Ready Signal

```yaml
plan_status: audit-ready
plan_complete_at: 2026-10-02T05:11:55Z   # revision 0.1.1 (iteration 1 delta); first plan completion was 2026-10-02T04:32:41Z
card: t1403
tier: M
plan_head: 1eef55dd9   # revision base HEAD (the revision commit follows it); first plan was authored at c50da9c2f
plan_audit: "iteration 1 of 2: FAIL 0.82 (audited_sha df226fe66; .moai/reports/t1403/plan-audit-iter1.md); iteration 2 of 2: PASS 0.94 against Tier M 0.80 (audited_sha b489d99f0; .moai/reports/t1403/plan-audit-iter2.md)"
artifact_sha256: e8f219f05fdb88aa23c419f912dfd658698e1964b7fc91c4493ae98b4ba37027   # combined ComputeHash over acceptance, plan, research, spec at b489d99f0; re-measured by the orchestrator and equal to the auditor's value
```

## §F Phase 4 Mode Selection

### F.1 Plan→run Kickoff gate (autonomous transition, auto-semantics §9.1)

| Condition | Observed | Evidence |
|---|---|---|
| Independent plan-audit verdict is PASS | PASS, 0.94 against the Tier M threshold 0.80 (iteration 2 of the Tier M ceiling of 2) | `.moai/reports/t1403/plan-audit-iter2.md` (`verdict: PASS` read from the file by the orchestrator) |
| Plan phase records audit-ready | `plan_status: audit-ready` | §E.1 above |
| Plan-artifact hashes unchanged since the verdict | equal | combined hash recomputed with the `ComputeHash` algorithm over acceptance, plan, research, spec at HEAD `b489d99f0`, tree clean: `e8f219f0…7027`, equal to the auditor's; per-file prefixes spec=f4238902 plan=ea7f7111 acceptance=98e97de1 research=b6aec967; artifact mtimes (14:05-14:14 local) precede the verdict file (14:28) |
| No blocker open | none | iteration-1 defects D1-D10 all RESOLVED per the iteration-2 regression table; the six minor items N1-N6 of iteration 2 are optional debts carried into the run delegation, not blockers; scope questions were settled by the operator (OD-1, OD-3) and the orchestrator (OD-2, OD-4, OD-5), tier by the auditor (OD-6: Tier M) |
| Keep-set case (environment-impossible / operator-held / irreversible external-shared) | none applies | the run edits files and commits locally in the card worktree; nothing is pushed (the factory leader pushes `develop`); no credential or runner is missing; no operator-held work is named by the card |

```text
decision record: decided_by=claude-code lane-2 orchestrator (factory lane, Kickoff autonomous transition) evidence_refs=.moai/reports/t1403/plan-audit-iter2.md(verdict=PASS score=0.94 audited_sha=b489d99f0),.moai/specs/SPEC-JEV-AUTO-EXCEPTION-001/progress.md#E.1,sha256 combined=e8f219f0 spec=f4238902 plan=ea7f7111 acceptance=98e97de1 ladder_path=gate-row plan→run Kickoff (AUTONOMOUS, auto-semantics §9.1)
```

Recorded 2026-10-02T05:30:27Z. The decision board under the moai home (auto-semantics §11) was NOT written, and this session did not attempt it: the predecessor card (t1400) recorded `factory_decide` as refused for a lane session and found no lane-writable board verb, and that finding was not re-measured here. This record lives here and in the card's local evidence directory. A reader must treat it as self-attested (auto-semantics §10).

### F.2 Orchestration mode (orchestration-mode-selection.md §D)

Input parameters: tier M; scope 19 files (13 distinct edits: six live/mirror pairs, one generated `catalog.yaml` hash line, plus the new guard test file) plus two completed-SPEC bodies; domain count 6 (Go comments and one new test, template mirrors, rule files, a skill, config YAML, SPEC bodies); file language mix: markdown 60%, Go 20%, YAML 20%; concurrency benefit LOW (coding-heavy, one writer per working tree, ordered landing G then K).

| Mode | Selected | Rationale |
|---|---|---|
| direct | no | not a trivial edit; a new guard test and a linked commit |
| serial | yes | default for coding-heavy work; the linked commit K needs one ordered writer |
| fanout | no | not research-heavy; parallel writers in one tree are prohibited |
| sweep | no | not one uniform mechanical transform; 19 files is below the soft ~30 boundary |

Decision: serial

Justification: milestones M1 to M4 of plan.md are ordered by obligation (guard commit G is an ancestor of the linked commit K), and every milestone writes the single card worktree, so one sub-agent per milestone in sequence is the only shape that keeps the commit graph witnessing the ordering. Boundary case: 19 files and 6 domains exceed the fanout thresholds (10 files, 3 domains), resolved toward the simpler mode because the work is coding-heavy and write-bound (§B.2).

Spawn shape: `general-purpose` carrying the manager-develop role charter for M1, M3 and M4 (a `manager-develop`-typed spawn auto-isolates into its own L1 tree and cannot write the card tree), and `manager-spec` for the two completed-SPEC bodies in M2.

## §G Operator Decisions

```yaml
- id: OD-1
  date: 2026-10-02
  question: scope extensions X1-X4 (spec.md §B.3) beyond the six card-named surfaces
  asked_via: AskUserQuestion, lane-2 session (recommendation_mode: pull, no option labelled recommended)
  options_shown: ["card six only", "X1 and X2", "X1-X4 all"]
  answer: "X1-X4 all"
  effect: spec.md §B.2/§B.3 stands as authored; no re-delegation; AC-JAE-006 and AC-JAE-007 stay in scope
- id: OD-2
  date: 2026-10-02
  decision: REQ-JEVO-009 accuracy-label set stays out of scope (spec.md §B.6, measured interaction with TestNoConsumerCallPathShips)
  decided_by: orchestrator on the author's evidence; no scope widening, so no operator question
- id: OD-3
  date: 2026-10-02
  question: plan-audit iteration 1 defect D2 — the ref skill moai-ref-jev-question-design (live and template mirror) carries the same "queue mutation" sentence as S3/S4; amend it or classify it out of scope
  asked_via: AskUserQuestion, lane-2 session (recommendation_mode: pull, no option labelled recommended)
  options_shown: ["include as X5", "classify out of scope (iii)"]
  answer: "include as X5"
  effect: surface X5 joins the amended set (two files, catalog.yaml hash regenerated as for X3); the closed-inventory sweep gains a closed-target-phrase pattern
- id: OD-4
  date: 2026-10-02
  decision: plan-audit D8 / assumptions A-2 and A-3 — both completed SPECs (SPEC-JEV-CORE-001, SPEC-MANAGER-TODO-001) are amended in place with a HISTORY row and keep status completed; the completed-to-in-progress amendment path is not used
  decided_by: orchestrator, following the precedent commit 185569ef3 and the audit's own check that moai spec lint and moai spec audit stay clean in that state
- id: OD-5
  date: 2026-10-02
  decision: assumption A-3 — the jev_ask tool description string (mcp_jev.go:48) and the internal/mcp catalog comment stay unchanged; the --auto ranking calls the client directly, never through handleJevAsk
  decided_by: orchestrator, on the author's measured reachability evidence and the iteration-1 audit's check of handleJevAsk call sites
- id: OD-6
  date: 2026-10-02
  decision: tier classification (Tier M kept by the author with 19 files against the tier table's 5-15 guidance) is not settled by the orchestrator; plan-audit iteration 2 judges it independently and states which threshold applies
  decided_by: orchestrator (deferral to the independent auditor)
```

## §E.2 Run-phase Evidence

### M1 — the guard, armed=false (commit G)

Attribution: measured in this run on tree HEAD `5f8c6e051` plus the M1 files (the new test file, the `spec.md` frontmatter flip, this section); go1.26.8 darwin/arm64; golangci-lint v2.1.6, the CI-pinned version of `.github/workflows/ci.yml`; every Go command ran with the eleven lane variables unset in the same compound invocation. Machine load average was about 300, so wall times are inflated. Commit G: `pending-backfill` (a commit cannot cite its own SHA; AC-JAE-013 records G and K at M4).

| Check | Command (scrub prefix omitted) | Deciding output | Exit |
|---|---|---|---|
| sweep control, before | `go test -count=1 -list '^(TestJevAutoExceptionLinkage\|TestJevAutoExceptionWording)$' ./internal/template/` | `ok  github.com/modu-ai/moai-adk/internal/template  0.490s`, no test name | 0 |
| sweep control, after | same | `TestJevAutoExceptionLinkage` `TestJevAutoExceptionWording` `ok … 0.837s` | 0 |
| AC-JAE-011 + AC-JAE-012 (armed=false) | `go test -count=1 -v -run '^(…Linkage\|…Wording)$' ./internal/template/` | 2 top-level `--- PASS`, 0 `--- FAIL`; 32 subtests PASS (18 `falsifier/partial/<n>`, `arming-only`, `self-match`, `split-commits`, `dangling-anchor`, `all-in-one-commit`, `tree` with `armed=false`, 6 wording falsifiers, `disclaimer-not-flagged`, `mirror-parity`); 7 group subtests SKIP, reason `surfaces unamended: armed=false; N passages located` with N = 2, 2, 4, 4, 2, 1, 8 = the located-passage count each group must find | 0 |
| arming literal absent | `git grep -c -F "jevAutoExceptionAmended = true" -- internal/template/jev_auto_exception_test.go` (staged) | (empty) | 1 |
| false constant present once | `git grep -c -F "jevAutoExceptionAmended = false" -- internal/template/jev_auto_exception_test.go` | `internal/template/jev_auto_exception_test.go:1` | 0 |
| probe control | `git grep -c -F "TestJevDoctrineAmendment" -- internal/template` | `internal/template/contract_mode_blocks_test.go:2` | 0 |
| neighbours | `TestJevAmendmentLinkage` (kickoff), `TestJevDoctrineAmendment` + `TestTemplateNoInternalContentLeak` (template), six cli guards, `TestNoConsumerCallPathShips`, `TestCatalogHashParity` | all `--- PASS`; `verified 49 catalog entries … 0 drift` | 0 |
| build / vet / format / lint | `GOOS=windows GOARCH=amd64 go build ./...`; `go vet ./internal/template/` (and with `GOOS=windows`); `gofmt -l`; `golangci-lint run ./internal/template/...` | 0, 0, empty, `0 issues.` | 0 |

RED (`EXPECTED_RED`, stub checkers returning nothing, same file): `--- FAIL` for `falsifier/partial/0` through `/17`, `arming-only`, `self-match`, `split-commits`, `dangling-anchor`, the six wording falsifiers, `mirror-parity` and the seven group subtests; the deciding lines read `checker accepted a registry missing internal/jev/jev.go`, `checker accepted markers split across commits`, `checker accepted a guard that spells its arming token`, `checker accepted a passage whose bound sits in another paragraph`, `located 0 passages, want 8`; the controls `all-in-one-commit`, `tree` and `disclaimer-not-flagged` passed against the stub, as a stub that finds nothing should.

Mutants (throwaway copies in the scratchpad, run through `go test -overlay`, nothing in the repository tree): bound checked passage-wide, closed targets ignored, long-form literals skipped, claim and over-reach sets removed, mirror compare removed, anchor check removed, commit comparison removed, arming row missing from the registry, anchors added to the commit comparison — each turned exactly its named subtest red (`bound-in-other-paragraph`, `closed-target-dropped`, `long-form-literal-missing`, `claim-phrasing` and `over-reach-phrasing`, `mirror-parity`, `dangling-anchor`, `split-commits`, `arming-only` and `self-match`, `all-in-one-commit`). A claim set matching the word `accurate` anywhere turned `disclaimer-not-flagged` red. The mutant that spells the arming literal in the guard's own source passed `tree` under `-overlay` because `tree` reads the on-disk file, which the overlay does not change: that probe is invalid for the on-disk read, and the class is carried by `falsifier/self-match`.

Claim and over-reach refinement (plan §E allows it): the reference expressions are kept; two narrowing exemptions were added. `validated as a whole` is not a claim when `answer` sits in the 30 characters before it; a negation (`not`, `never`, `no`, `nothing`, `nor`) in the 30 characters before `every|any|all moai todo|pick(s)` makes it a scope narrowing, not an over-reach. Verified by running the reference expressions alone against the accepted-sentence table: they flagged four legitimate sentences (`The answer set is validated as a whole…`, `An answer that is validated as a whole may set selection order only.`, `It does not apply to any moai todo pick outside the --auto cycle.`, `Nothing here changes all picks outside the cycle.`) and the exemptions clear exactly those four; the plan's eight documented outcomes are unchanged (the four flagged phrases stay flagged, the four accepted phrases stay accepted), and two sentences the exemptions must not clear (`The ordering is validated as a whole.`, `It applies to every moai todo pick outside the --auto cycle.`) stay flagged. Not fixed: `ordering accuracy was measured by nobody` is still a false positive.

Process attestation (M1): one writer in the card worktree; the files of this commit were staged by explicit pathspec; Gaps and residual risk are in the lane report.

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
