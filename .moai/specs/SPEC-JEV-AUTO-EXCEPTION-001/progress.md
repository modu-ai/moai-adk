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
  decided_by: orchestrator, following the precedent commit 185569ef3. Correction recorded after the run: plan-audit iteration 1 observed the UNMODIFIED completed SPECs lint-clean, not the amended state; the post-edit state was first observed at M2, where moai spec lint printed No findings for both SPECs.
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

Attribution: measured in this run on tree HEAD `5f8c6e051` plus the M1 files (the new test file, the `spec.md` frontmatter flip, this section); go1.26.8 darwin/arm64; golangci-lint v2.1.6, the CI-pinned version of `.github/workflows/ci.yml`; every Go command ran with the eleven lane variables unset in the same compound invocation. Machine load average was about 300, so wall times are inflated. Commit G: `6d012fd4d` (backfilled after M4; a commit cannot cite its own SHA, and AC-JAE-013 records G and K at M4).

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

### M2 and M3 — the linked commit K (`7983d9131`), evidence read back at M4

M2 (the two completed-SPEC bodies) was left uncommitted by design and M3 (the rest of the set, plus the arming constant flipped to `true`) was added to it, so M2 and M3 share one commit: **K = `7983d9131b018d88c56505e901559dc0ec139cb8`**, 19 files (`git show --stat --format=%h 7983d9131`: 80 insertions, 28 deletions): `jev.go`, `mcp_jev.go`, `workflow.yaml` x2, the catalogue x2, `SPEC-JEV-CORE-001/spec.md` (0.3.0 to 0.4.0), `SPEC-MANAGER-TODO-001/spec.md` (0.1.0 to 0.2.0), `jev-local-operations.md`, `agent-authoring.md` x2, `SKILL.md` x2, `CLAUDE.md` x2, the question-design skill x2, `internal/template/catalog.yaml`, and the one-line flip in `jev_auto_exception_test.go`. The list equals the plan's 19 (plan.md §E "Files changed"), and no file outside M2-M3 is in it.

### M4 — evidence and closure (this record; commit order and sweeps below)

**Claim.** AC-JAE-001..014 were each exercised against tree K: 13 PASS, 1 PASS-WITH-DEBT (AC-JAE-014, one clause), 0 FAIL.

**Baseline-attribution.** Every row below was measured in this run on tree HEAD `7983d9131` (`git status --short` empty, branch `WT-jev-auto-exception`); card base `CARD_BASE` = `c50da9c2f8aa1227073bd77caa07ca1c75b8d81b` from `git merge-base develop HEAD` taken in this run. Tools: go1.26.8 darwin/arm64; golangci-lint v2.1.6 (the CI pin); tree-local `moai` built in this run from this tree by `go build -o <scratchpad>/moai ./cmd/moai` and invoked by path (`moai version` prints `v3.1.3 none built unknown`: no ldflags, so the binary carries no commit stamp and its provenance is by construction, see Gaps). Machine load average 40.5 43.8 117.4 at 15:53 local. Every Go command ran as one compound `unset <11 lane variables> && go test -count=1 …` with an anchored `-run`, one package per invocation; `go test -list` came first and printed the swept names. Exit codes: `go test` prints `ok` only on exit 0 and that line is the observation (no separate field); `git grep` and build exits were made observable with `&& echo EXIT0 || echo EXIT1` (a no-match grep prints nothing here).

**Evidence — E1, the AC matrix (command, deciding output, exit):**

| AC | Result | Deciding output (this run, tree K) | Exit |
|---|---|---|---|
| 001 | PASS | `git grep -c -F "auto-scoped ranking exception" -- internal/jev/jev.go internal/cli/mcp_jev.go` gives `mcp_jev.go:1` `jev.go:1`; the same for `selection order only` and `mechanical filters`; `doc_display_only_test` in `jev.go` gives nothing; `NearDuplicateMark\|LaneQuestionRoute\|SkillSuggest` in both files gives nothing; `TestJevAutoExceptionWording/go-comments` `--- PASS` | 0, 0, 0, 1, 1, ok |
| 002 | PASS | literal `:1` in each `workflow.yaml` copy, `mechanical filters` `:1` in each; `TestJevDoctrineAmendment` (6 subtests incl. `rules-and-config`) `--- PASS`; wording `config-comment` and `mirror-parity` `--- PASS` | 0, 0, ok |
| 003 | PASS | literal `:2` and `never through this tool` `:2` in each catalogue copy; `TestMCPToolCatalogueDocsStayMirrorIdentical` `--- PASS`; wording `catalogue-rows` `--- PASS` | 0, 0, ok |
| 004 | PASS | `version: "0.4.0"` at `spec.md:4`; `v0.4.0` count 5 (>= 3); `mechanical filters` count 3 (>= 2); literal count 5; `moai spec lint SPEC-JEV-CORE-001` gives `✓ No findings — all SPEC documents are valid`; `git diff --stat CARD_BASE -- .moai/specs/SPEC-JEV-CORE-001/progress.md` empty; `TestJevDoctrineAmendment/spec` and wording `spec-core` `--- PASS`; `status: completed` kept | 0, 0, 0, 0, ok |
| 005 | PASS | literal count 3 in `SPEC-MANAGER-TODO-001/spec.md` (>= 2); `^- \*\*REQ-MT-0(14\|15)\*\*` count 2; `moai spec lint SPEC-MANAGER-TODO-001` `✓ No findings`; wording `spec-manager-todo` `--- PASS`; `status: completed` kept | 0, 0, 0, ok |
| 006 | PASS | literal `:1`, `기계적 필터` `:1`, pinned `예외는 한 곳뿐이다` `:1` in `jev-local-operations.md`; `git diff --stat CARD_BASE -- AGENTS.local.md` empty; `TestJevDoctrineAmendment/local-guide` and wording `local-guide` `--- PASS` | 0, 0, 0, ok |
| 007 | PASS | literal `:1` in each of `agent-authoring.md`, `SKILL.md`, `CLAUDE.md`, live and template (six lines), and in the question-design skill pair (`:1` each); `wc -c CLAUDE.md internal/template/templates/CLAUDE.md` gives `15658` and `15658` (<= `15733`; base was 15573, growth 85 bytes); `diff -q` silent and exit 0 for the `CLAUDE.md` pair and for the skill pair; the skill pair carries none of `internal/jev\|mcp__moai__jev\|jev_ask\|moai jev` (grep exit 1); wording `extension-rows` `--- PASS`; `TestJevQuestionDesignSkillCarriesNoCallPath` and `TestJevQuestionDesignSkillCopiesStayIdentical` `--- PASS`; `TestCatalogHashParity` `verified 49 catalog entries … 0 drift`; `git diff --stat CARD_BASE -- internal/template/catalog.yaml` shows 2 insertions, 2 deletions | 0, 0, 0, ok |
| 008 | PASS | `TestMCPToolCatalogueDocsStayMirrorIdentical` `--- PASS`; `TestTemplateNoInternalContentLeak` `--- PASS (1.99s)`; wording `mirror-parity` `--- PASS` | ok |
| 009 | PASS | `TestJevDoctrineAmendment` 6/6; `TestJevAmendmentLinkage` 6/6 with `tree` logging `JevDoctrineAmended = true`; `TestMCPToolCatalogueDocsStayMirrorIdentical` plus the five `--auto` guards, six `--- PASS`; `TestNoConsumerCallPathShips`, the two `internal/jev` import tests, the two question-design guards and `TestCatalogHashParity` `--- PASS`; `git diff --stat CARD_BASE -- activation_test.go contract_mode_blocks_test.go kickoff.go AGENTS.local.md` empty | ok, 0 |
| 010 | PASS | D-7 below; L11 baseline `git grep -c -F "DISPLAY-ONLY: the answer is a labelled model signal a person reads" -- internal/cli/mcp_jev.go` gives `internal/cli/mcp_jev.go:1`; `git diff --stat CARD_BASE -- todo_auto_rank.go defaults.go catalog.go` empty; `go build ./...` `BUILD_LINUXLOCAL_EXIT0`, `GOOS=windows GOARCH=amd64 go build ./...` `BUILD_WINDOWS_EXIT0`; `TestPackageImports_AreStandardLibraryOnly` and `TestImportClassifierPositiveControl` `--- PASS` | 0 |
| 011 | PASS | `-list` printed `TestJevAutoExceptionLinkage` `TestJevAutoExceptionWording`; `--- PASS: TestJevAutoExceptionLinkage (8.26s)` with 18 `falsifier/partial/<n>`, `arming-only`, `self-match`, `split-commits`, `dangling-anchor`, `all-in-one-commit`, `tree` all PASS; `tree` logs `armed=true`; falsifier logs read `partial amendment`, `first appears in … not …`, `dangling amendment`; the D1 control at G prints nothing (AC-013 below) | ok |
| 012 | PASS | `--- PASS: TestJevAutoExceptionWording (0.01s)`: the six falsifiers, `disclaimer-not-flagged`, `mirror-parity` and the seven group subtests (`go-comments` `config-comment` `catalogue-rows` `spec-core` `spec-manager-todo` `local-guide` `extension-rows`) all `--- PASS`, no `--- SKIP`; sweep control: 2 top-level `--- PASS`, 0 `--- FAIL` | ok |
| 013 | PASS | next block | see block |
| 014 | PASS-WITH-DEBT | sweeps below: every hit classified, every class-(i) passage carries the literal, both controls present, baselines re-measured equal; the clause "and no new file" is NOT met (see sweeps and finding 2) | n/a |

All of the above are the deciding lines; the full verbose output (about 330 lines for the two new guards) was read in the lane and is not kept (Residual-risk).

**AC-JAE-013 — commit order (graph reads in this run):** G = `6d012fd4db10e674b6341d57dcd2ec5430c66797`, K = `7983d9131b018d88c56505e901559dc0ec139cb8` (two different SHAs).

| Check | Command | Output | Exit |
|---|---|---|---|
| G is an ancestor of K | `git merge-base --is-ancestor 6d012fd4d 7983d9131 && echo ANCESTOR_YES` | `ANCESTOR_YES` | 0 |
| token absent at G | `git grep -c -F "jevAutoExceptionAmended = true" 6d012fd4d -- internal/template/jev_auto_exception_test.go \|\| echo NO_MATCH_exit_nonzero` | `NO_MATCH_exit_nonzero` | 1 |
| false constant at G | same with `= false` | `6d012fd4d:internal/template/jev_auto_exception_test.go:1` | 0 |
| token once at K | same with `= true` at `7983d9131` | `7983d9131:internal/template/jev_auto_exception_test.go:1` | 0 |
| only K adds the token | `git log --format=%h -S"jevAutoExceptionAmended = true" -- internal/template/jev_auto_exception_test.go` | `7983d9131` (one line) | 0 |
| K's file list | `git show --stat --format=%h 7983d9131` | the 19 files above, equal to the plan's list | 0 |
| mechanical one-commit check | `TestJevAutoExceptionLinkage/tree` | `--- PASS`, `armed=true` | ok |

**D-7 / AC-JAE-010 — comment-only.** `CARD_BASE` taken now by `git merge-base develop HEAD` = `c50da9c2f8aa1227073bd77caa07ca1c75b8d81b`. `git diff -U0 c50da9c2f8aa1227073bd77caa07ca1c75b8d81b -- internal/jev/jev.go internal/cli/mcp_jev.go` shows exactly two hunks, `mcp_jev.go @@ -10 +10,7` and `jev.go @@ -27,2 +27,10`, both inside comment blocks (the second also corrects the stale `doc_display_only_test.go` name to `display_only_test.go`); no hunk touches the import blocks or the `mcp.WithDescription` line (`mcp_jev.go:48`). The filter `… | grep -E '^[+-]' | grep -v -E '^(\+\+\+|---)' | grep -v -E '^[+-][[:space:]]*//'` printed nothing on this diff (`FILTER_EMPTY_nonmatch_exit_1`). Positive control: the same filter over `git diff -U0 7d8a9bdbc c50da9c2f -- internal/cli/todo_auto.go` printed four code lines (`+`, `+	landed  autoLandedLookup // …`, `+	jevRank autoJevRanker    // …`, `+	targets = autoRankTargets(out, rec, targets, opts)`), so the empty result is not a broken filter.

**Sweeps (plan §F V12; each excludes `.moai/specs`, `.moai/reports`, `CHANGELOG.md`; patterns exactly as written there).** Baselines were re-measured on the card base `c50da9c2f` in this run and equal the plan-time figures.

| Sweep | Baseline (re-measured at `c50da9c2f`) | Now (K) | Delta |
|---|---|---|---|
| primary `display-only\|display only` | 69 hits / 43 files | 76 hits / 44 files | +7 hits, +1 file |
| synonym `never reorder by inferred priority\|판단 자료\|모델 답을 입력으로도` | 5 hits | 6 hits | +1 |
| closed-target `a person reads\|labelled model\|queue mutation\|hard to undo` | 84 hits / 47 files | 96 hits / 48 files | +12 hits, +1 file |

Controls: `internal/mcp/catalog.go:97:	// Gated judgment wrapper (display-only): …` present in the primary sweep; `internal/cli/todo_triage.go:14` (and `:123`) present in the closed-target sweep.

Classification of every hit that is new since the baseline (the 7, 1 and 12 are accounted for in full; per-file counts for the live files at base and at K were read with `git grep -c`, template mirrors reconcile by the totals and by identical added lines):

| Sweep | File:line | Class | Reason |
|---|---|---|---|
| primary | `internal/cli/mcp_jev.go` (3rd hit, the added comment line `// stays display-only for every caller.`) | (i) amended surface | inside the amended comment block that carries `auto-scoped ranking exception`, `selection order only`, `workflow.jev.enabled`, `mechanical filters` |
| primary | `internal/template/jev_auto_exception_test.go` x6 (package-comment locator, file-comment locator, `REQ-MT-014` closed list, manager-todo line locator, retained-agents line locator, the fixture string `// The capability is display-only: …`) | (iii) different object, **new file** | locators and fixtures of the new wording guard; not a doctrine statement; same kind as the predecessor guard `contract_mode_blocks_test.go` in research.md §R1.2 |
| synonym | `internal/template/jev_auto_exception_test.go:188` (locator `never reorder by inferred priority`) | (iii) different object | guard locator |
| closed-target | `internal/jev/jev.go` (3rd hit, `a queue mutation` in the added comment) | (i) amended surface | inside the amended block with both literals |
| closed-target | `internal/template/jev_auto_exception_test.go` x11 | (iii) different object, **new file** | closed-target tables and fixtures of the guard |

Hits that exist at base and were rewritten in place keep their count (catalogue 3/3 and 2/2, `agent-authoring.md` 1/1, `workflow.yaml` 1/1, `CLAUDE.md` 1/1, `jev.go` 1/1 primary) and are class (i), now carrying the literal; `SKILL.md:180` (synonym) was rewritten in place the same way. Class-(i) check, `git grep -c -F "auto-scoped ranking exception"` per file: `jev.go:1`, `mcp_jev.go:1`, `workflow.yaml` x2 `:1`, catalogue x2 `:2`, the two completed SPECs `:5` and `:3`, the guide `:1`, `agent-authoring.md` / `SKILL.md` / `CLAUDE.md` live and template `:1` each, question-design skill live and template `:1` each; landed anchors still carry it (`manager-todo.md:2`, `kanban-dispatch.md:1`, `gtd.md:1`).

**Other checks (E2-E5).** E2: `go build ./...` and `GOOS=windows GOARCH=amd64 go build ./...` both reached their `echo` (exit 0). E3: no production logic changed (comments plus one test file); `go test -count=1 -cover ./internal/jev/` gives `ok … coverage: 95.6% of statements`, reported only as context; no coverage target applies to changed lines because none is executable. E4: `git grep -n AskUserQuestion -- internal/jev internal/cli/mcp_jev.go` printed nothing (`NO_ASKUSERQUESTION_MATCH`). E5: `golangci-lint run --timeout=10m ./internal/jev/... ./internal/template/` (v2.1.6) prints `0 issues.`; `go vet ./internal/template/ ./internal/jev/ ./internal/cli/` reached `VET_EXIT0`; `gofmt -l` on the three Go files printed nothing. `moai spec lint` (tree-local build, tree HEAD `7983d9131`): `SPEC-JEV-CORE-001`, `SPEC-MANAGER-TODO-001`, `SPEC-JEV-AUTO-EXCEPTION-001` each `✓ No findings — all SPEC documents are valid`. `moai spec audit --filter-spec <id> --include-grandfathered --json` for the three: each `total_specs 1`, `modern_era_clean 1`, one finding of type `EraAutoDetected`, severity INFO (heuristics H-4 for the two completed SPECs, H-5 for this SPEC); no drift finding. E8: the M1 RED recorded above (commit `6d012fd4d`) was not re-run.

**Guard results at K (swept names from `go test -list` first):** `./internal/template/` 5 names (`TestJevAutoExceptionLinkage` ok, `TestJevAutoExceptionWording` ok, `TestJevDoctrineAmendment` ok, `TestTemplateNoInternalContentLeak` ok, `TestSanitizedPairParity` ok); `./internal/cli/` 8 names (`TestMCPToolCatalogueDocsStayMirrorIdentical`, `TestAutoRankDoctrineAmendment`, `TestAutoRankMirrorParity`, `TestAutoRankMarkerDisclosure`, `TestAutoHelpAndRefusalDoNotAssertPickOrder`, `TestAutoRankAgentDoctrine`, `TestJevQuestionDesignSkillCarriesNoCallPath`, `TestJevQuestionDesignSkillCopiesStayIdentical`) all `--- PASS`; `./internal/contract/kickoff/` 1 name (`TestJevAmendmentLinkage`, 6 subtests PASS); `./internal/jev/` 2 names PASS; `./internal/jevmeasure/` 1 name PASS; `./internal/spec/` 1 name (`TestCatalogHashParity`, 0 drift). `TestTemplateNoInternalContentLeak` and `TestSanitizedPairParity` were run with `-v` through a `grep -E '^(--- |ok|FAIL|PASS)'` filter, so their exit code is read from the printed `ok` line.

**Run-phase findings no requirement names (SPEC body not edited; the sync phase reconciles):**

1. **Guard size.** `internal/template/jev_auto_exception_test.go` is 846 lines (`wc -l`); plan.md D-8 (`plan.md:183-187`) estimated "about 250-350 lines" for the test and "under 1000 on any reading" in total. The run's insertions are 869 at G (846 of them the test) and 80 at K, 949 together: still under 1000 (`git show --stat`), but with a margin of 51 lines instead of the roughly 500 the estimate implied; the 19-file count already put the SPEC above the tier table's 5-15 (OD-6).
2. **AC-JAE-014 "no new file" is not literally true.** The guard test is a new file in the primary sweep (43 to 44 files) and the closed-target sweep (47 to 48 files), and adds a synonym hit (5 to 6). It is class (iii) (guard locators and fixtures). acceptance.md AC-JAE-014 and research.md §R1 do not list it; the sync phase should add it to §R1 as class (iii) and relax the clause.
3. **Claim and over-reach expressions were refined in M1** (recorded above, plan §E allows it): two narrowing exemptions were added so four legitimate sentences stop being flagged; the eight documented outcomes are unchanged. One false positive remains by design of the M1 author: `ordering accuracy was measured by nobody`. The final amended passages pass (go-comments, config-comment, catalogue-rows, spec-core, spec-manager-todo, local-guide and extension-rows PASS).
4. **Linkage test timing.** M1 recorded 110-210 s for the linkage test at a load average of about 300; here `TestJevAutoExceptionLinkage` took 8.26 s (`tree` 2.70 s, `all-in-one-commit` 0.63 s, `split-commits` 0.52 s) at load averages 40.5 43.8 117.4. The cost is git subprocess launches, so it scales with machine load, not with the tree.
5. **`tree` runs `git log -S` once per registry row when armed** (18 rows) plus the presence reads; with `armed=false` it only asserts absence. The registry has 18 rows (17 surfaces plus the guard file itself), which is also why there are `falsifier/partial/0` through `/17`.
6. **K carries no `Authored-By-Agent` trailer** (`git log --format=%b -n 1 7983d9131` ends at `🗿 MoAI`); G carries `Authored-By-Agent: manager-develop`. acceptance.md Definition of Done item 4 names the trailers as the nearest readable witness of who wrote what, so for K the manager-spec / manager-develop split rests on the lane's attestation alone. K is a joint commit staged by the lane orchestrator, so a single-agent trailer would be wrong; the SPEC does not say what K should carry.
7. **`SPEC-MANAGER-TODO-001` went from 0.1.0 to 0.2.0** (a HISTORY row was added), while acceptance.md AC-JAE-005 asserts a new HISTORY row but no version bump for that SPEC (only `SPEC-JEV-CORE-001` v0.4.0 is asserted). No guard pins it.
8. **The `moai` binary carries no commit stamp** (`v3.1.3 none built unknown` without ldflags), so verification-claim-integrity §2.2's "name the judging build's commit" can only be met by construction (built from this tree in this run), not by reading the binary.
9. **`jev.go` now names the correct test file.** The stale `doc_display_only_test.go` reference (`jev.go:27`, ledger L15) is replaced by `display_only_test.go`; `git grep -c -F "doc_display_only_test" -- internal/jev/jev.go` is empty (exit 1). Not named by any requirement beyond AC-JAE-001.

**Gaps (explicitly not observed):**
- `make build` was not run, so the clause "if X3 or X5 survives … `make build` leaves no further diff" (AC-JAE-007) is unverified; `TestCatalogHashParity` (0 drift over 49 entries) is a different mechanism and proves hash parity only.
- `golangci-lint` was not run on `./internal/cli/` (the only change there is a comment in `mcp_jev.go`; `go vet` and `gofmt -l` were clean). It ran on `./internal/jev/... ./internal/template/`; a baseline run was not made because the result is zero issues.
- The remaining Definition of Done items: item 7 (the queue store and the live queue unchanged) was not observed, since the queue lives outside the repository and no queue command was run; item 4's process attestation (one writer, explicit-pathspec staging of K) cannot be read from the graph (finding 6); item 5's completion-report statements are the lane report's.
- No `git fetch` or push was run (the leader pushes `develop`), so `l44_pre_commit_fetch` and `l44_post_push_fetch` are recorded as not-run; the CI verdict for the repository-wide suites, including the darwin and windows test matrix, is PENDING at report time and belongs to the CI run on the integration branch.
- Template-mirror per-file hit counts were not read one by one with `git grep -c`; they reconcile through the totals (+7, +1, +12 each accounted for) and the identical added lines in the diff.
- The sweeps close only over their three stated patterns (spec.md §G R-8).
- That the amended prose is correct doctrine (not only that the literals and bounds are present) is outside what the guards test; the sync-audit reads it.

**Residual-risk:** the verbose output of the two new guards (about 330 lines, mostly the 18 `falsifier/partial` lines) is not kept; the deciding lines are above. A phrasing of the exception that avoids all three sweep patterns would still be missed (G-9 of the plan). The remaining claim-regexp false positive (finding 3) can trip a future legitimate sentence. The linked commit reverted whole is undetected (spec.md §G R-2). A green local run is an early signal only; the CI verdict decides.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_status: audit-ready
run_complete_at: 2026-10-02T06:53:53Z
run_commit_sha: 14ba2eca7          # the commit that carried this block (docs(SPEC-JEV-AUTO-EXCEPTION-001): M4 run-phase evidence); backfilled by the following commit, since a commit cannot cite its own hash
guard_commit: 6d012fd4d            # G, AC-JAE-013
linked_commit: 7983d9131           # K, AC-JAE-013
ac_pass_count: 13
ac_pass_with_debt_count: 1         # AC-JAE-014: the "no new file" clause (guard test file, class iii)
ac_fail_count: 0
preserve_list_post_run_count: 0    # the 7 files named by the AC-JAE-009 and AC-JAE-010 no-diff lists: none differs from CARD_BASE
l44_pre_commit_fetch: not-run      # no fetch; the leader pushes develop
l44_post_push_fetch: not-run       # no push by this lane
new_warnings_or_lints_introduced: 0   # golangci-lint v2.1.6 on ./internal/jev/... ./internal/template/: 0 issues; go vet clean on three packages
cross_platform_build:
  native_darwin_arm64: exit 0
  windows_amd64: exit 0
total_run_phase_files: 21          # distinct files over G and K (3 + 19, the guard test counted once)
m1_to_mN_commit_strategy: "G (guard, armed=false) then K (M2+M3 linked, armed=true), then this evidence commit and its SHA backfill; no push"
```

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_status: audit-ready
sync_complete_at: 2026-10-02T07:11:51Z
sync_commit_sha: 0757729f0   # the commit that carried this block (docs(SPEC-JEV-AUTO-EXCEPTION-001): sync-phase artifacts); backfilled by the following commit, since a commit cannot cite its own hash
card: t1403
tier: M
ac_source: .moai/specs/SPEC-JEV-AUTO-EXCEPTION-001/acceptance.md   # tier M: acceptance.md is the AC source; resolver state: resolved
ac_total: 14                    # live count from acceptance.md (counter: live=14 excluded=0 ambiguous=0); not taken from this file
ac_pass_count: 13               # as recorded in §E.3
ac_pass_with_debt_count: 1      # AC-JAE-014; resolved in the SPEC body reconciliation (guard source classified (iii))
ac_fail_count: 0
changelog_entry_position: "CHANGELOG.md [Unreleased] ### Added, first bullet (above the SPEC-TODO-AUTO-PRIORITY-001 entry)"
b12_self_test_a: "pre-emission grep: git grep -c 'SPEC-JEV-AUTO-EXCEPTION-001' -- CHANGELOG.md printed nothing, exit 1 (zero entries) before emission"
b12_self_test_b: "AC count: awk counter over acceptance.md printed 14 (live=14 excluded=0 ambiguous=0); the entry states 14, AC-JAE-001..014"
b12_self_test_c: "path verification: ls of every path the entry names (and the live/template mirrors) exited 0"
frontmatter_status_transitions:
  spec_md: "in-progress -> completed (single sync commit); updated: 2026-10-02 (unchanged, already the sync date)"
  plan_acceptance_progress: "not applicable: plan.md and acceptance.md are stateless on the status axis and carry no status field; progress.md carries none"
mx_tag_validation: "This change adds no @MX tags: comments, SPEC and rule prose, and one Go test file only. git grep -n '@MX' over jev_auto_exception_test.go, internal/jev/jev.go and internal/cli/mcp_jev.go found no tag in the test file and four pre-existing tag lines in the two production files (jev.go:5-6 ANCHOR+REASON, mcp_jev.go:71-72 WARN+REASON), none touched by this SPEC (the diff against the card base is one comment hunk in each file, at jev.go line 27 and mcp_jev.go line 10)."
docs_site_readme_check: "git grep -n -i -E 'display-only|display only' -- docs-site README.md README.ko.md README.en.md README.ja.md README.zh.md: one hit, docs-site/content/en/guides/mcp-server.md:208 (heading 'Judgment (gated, display-only)'), classified as describing the jev_ask MCP tool, which stays display-only; no edit. README files carry no Jev mention."
sync_audit: pending            # sync-auditor not yet run
decision_record_reread: pending   # the sync audit re-reads the plan->run Kickoff decision record of section F.1
```

