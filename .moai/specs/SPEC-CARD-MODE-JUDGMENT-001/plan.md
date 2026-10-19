# plan.md — SPEC-CARD-MODE-JUDGMENT-001

Documents-only change, Template-First, Tier L. Milestones are ordered by decision reversibility — the wording most likely to be disputed first, the mechanical sync last — and by priority, never by duration. Requirements: `spec.md` §4. Criteria: `acceptance.md`. Constraints and candidate wording: `design.md`. Measurements and line numbers: `research.md`.

## A. Technical approach

1. **Edit order inside every milestone: template side first, then the identical edit on the deployed side** (`.claude/...`), then run the milestone's read-only checks. Never the reverse; a deployed-only edit is a Template-First violation and AC-MJ-001..009 read both sides.
2. **One home per rule.** M1 writes C's `### Mode judgment`; M2 writes M's `### Lease renewal`; M3 writes D's Routing bullet. Other sites receive a marker or a pointer, not the rule (REQ-MJ-012, `design.md` §D-2).
3. **Do not edit a ledger-pinned line that has no reason to change** (`design.md` §D-3): F:30, K:32/36/68 stay verbatim. Lines whose scope must change (F:119, F:122, K:28, A §5.1, W §0a, M:26, D:133) are rewritten, and each ledger row whose `after_text` is rewritten (at least ALB-0276) is updated in the same commit.
4. **Worded fresh.** New text copies no ledger-covered line of ≥ 40 units into a `paths:`-scoped file (AC-ALB-021(2)), and carries no card id, board id, SPEC id, date, or long hex run (AC-MJ-013).
5. **Class vs mode.** Say once, in C, that class (A/B/C) names the stages and mode names whether a card may run beside others; the two axes are independent.

## B. Milestones

### M1 — Mode judgment and explicit issuance (Priority High; highest change-likelihood: user-facing procedure)

Files (template, then deployed): C, G, F.

1. C: add `### Mode judgment — serial or parallelizable` after the parallelism paragraph (line 34), before `## Card size ceiling`. Content per `design.md` §D-9: file-set predicate, the two `serial` triggers, reason format, explicit `--classification-file` (decider `human`), reissue-not-default, and the Class C `--spec` sentence containing `cannot be added after assign`.
2. G: rewrite line 52 so the issuance instruction carries `--classification-file <path|->` and cites `factory-dispatch-cards.md` § Mode judgment; qualify lines 267-270 and 330-333 for Class C (REQ-MJ-005). Leave line 56's factual sentence and the `--force` row alone (REQ-MJ-015).
3. F: add one new [HARD] line directly after line 30 (the explicit-classification obligation, citing C). Leave line 30 verbatim.
4. Check: AC-MJ-001, -002, -003, -004 green; AC-MJ-010 and -017 still green. `TestTodoSkillDocumentsClassification` neutrality: no SPEC id / date / hex in G.

Open items touching M1: none blocking. Q1 default (agents untouched) and Q2 default (`doc-generation.md` untouched) apply; Q5 default keeps the board's words verbatim.

### M2 — Executing-card-only lease renewal and scoped standing cron (Priority High; most pinned)

Files (template, then deployed): M, K, F, A, W, plus `internal/template/testdata/binding_ledger.json` rows.

1. M: add `### Lease renewal — only the card being executed` under `## Factory Mode mechanics` with the tokens of AC-MJ-007; rewrite line 26 to point at it (the `moai factory renew` verb and `factory_lease.duration` text stay). Define *executing* once (`spec.md` §3).
2. K line 28: marker form (`design.md` §D-4); keep ``every renewal period of `workflow.factory_lease.renew_interval` `` on that line; net ≤ 0 units against the 9,999 baseline after an offsetting trim of ≥ 8 units from a non-ledger-covered core line (verify with `git grep` that no test pins the line first).
3. F line 119 (the ledger-pinned compressed line) and line 122 (the pinned long paragraph): add the marker; update ledger row ALB-0276's `after_text` to the new line in the same commit. Keep the `### Lane waits are explicit, and stalls are watched` tokens (`CronCreate`, `recurring`, `disk evidence`).
4. A §5.1 (lines 113-146): scope "armed once … and kept until the completion report" and "Each fire runs the watchdog's renewal first" to the executing card; keep the §5 tokens of `design.md` §D-3; reflow so the marker shares a physical line with each `kept`/`keeps` phrase (the per-line trap).
5. W §0a: add the scope clause; keep heading word `Renew`, `moai factory renew`, the closed `renew_result` set, "does not end the pass", and order (Renew before Measure). A pass that does not renew carries the previous `renew_result` forward.
6. Check: AC-MJ-007, -008, -009, -012 green. If the ledger row edit is needed, `TestBindingLedgerIntegrity` is the proof (M5).

### M3 — Free lane and assignment input (Priority Medium; the assignment-input sentence is held for Q3)

Files (template, then deployed): D.

1. D line 133 (Routing bullet): replace "free means the lane's previous card reached `done`…" with the executing-card definition (token `no card is actually executing`), the merge-ready / held / serial-waiting inclusion, and the one-sentence separation from the live-session condition citing M § Boundaries and carrying the phrase `not a free lane` (REQ-MJ-006, AC-MJ-005c). Line 35 (the diagram) keeps the term unchanged.
2. **Held for verdict Q3** — write the `assignment input` sentence (REQ-MJ-007, token `assignment input`) only after the operator's verdict. Under the default reading the sentence names a candidate set and leaves promotion to the existing clauses; under the Alternate, REQ-MJ-007 and AC-MJ-006 are amended first.
3. Check: AC-MJ-005 green; AC-MJ-006 green only after Q3.

### M4 — Mechanical sync of generated and fixture files (Priority Medium; mechanical, last)

1. `internal/template/catalog.yaml`: recompute the stored hashes for the entries whose skill directory changed — `moai` (contains `workflows/gtd.md`) and `moai-lane-watchdog` — with `go run ./internal/template/scripts/gen-catalog-hashes.go --entry moai` and `--entry moai-lane-watchdog`. (`moai-factory-foreman` is untouched by default.)
2. `binding_ledger.json`: confirm every rewritten covered line has its row updated (M2 step 3) and no row's `before_text`/`after_text` absence remains.
3. Re-run AC-MJ-013 (`E-013a` command) now that a diff exists; a non-empty stat names a leaking file.
4. Agent files are not edited by default, so `make agents-emit` and the `.codex` TOMLs are untouched; if Q1 flips to include them, run `make agents-emit` and re-run the emit check.

### M5 — Verification and evidence (Priority High; the gate)

Read-only checks first, in one parallel batch (AC-MJ-001..013, -016, -017 commands from `acceptance.md`), outputs redirected to files under the card's scratch directory with the exit code and a bounded tail surfaced; the lines that decide each claim go into `.moai/reports/t1894/verdict.md`.

Then the **scoped Go checks** (AC-MJ-014). Never `go test ./...`:

- `internal/template`, one invocation: `go test ./internal/template -count=1 -run '^(TestBindingLedgerIntegrity|TestRoleCoreStubGate|TestDeployedAlwaysLoadedCharBudget|TestRuleTemplateMirrorDrift|TestDeclaredRuleMirrorForks|TestReviewOwnership_CardReviewDoctrine|TestFactoryDispatchBodyCarriesRenewIntervalLinkage|TestLaneRecheckDoctrine|TestLaneWatchdogRenewsFirst|TestLaneWatchdogRecordsRefusalAndMeasures|TestWorkflowRulePathsPinned|TestTemplateNoInternalContentLeak|TestCatalogHashCoversSkillSubfiles)$'`
- `internal/cli`, one invocation: `go test ./internal/cli -count=1 -run '^(TestAutoPickMirrorParity|TestAutoPickDocDoctrine|TestTodoSkillDocumentsClassification)$'`

Record, per verdict, the empty-sweep check (each `-run` must report a non-zero test count — `[no tests to run]` is a failure, `verification-completeness.md` §1.1) and the judging build (build from this tree and invoke it by path, or state that the installed build is not a strict ancestor — `verification-claim-integrity.md` §2.2).

**Resource-slot and admission note (run phase; not an action now).** Compiling the `internal/cli` test binary is the heavy step (the full `internal/cli` suite is on the repository's heavy-run list). Before the Go invocations the lane takes the resource lease (`moai slot acquire --resource <name> --max-duration <bound>` … `moai slot release --resource <name>`; `resource-slot-lease.md`), follows the leader's resource-entry rules (load and memory admission, staggered start, one Go invocation per lane at a time), and runs each verification as a single `unset <VARS> && <command>` call when it needs a scrubbed environment (`AGENTS.md` §4). `make build` (which also runs the agent- and command-emit drift checks) is not required for a documents-only change and is taken only if the leader asks for it, in the same lease.

## C. Order and dependencies

M1 → M2 → M3 → M4 → M5, with M3 step 2 waiting on Q3 and M5 running after M4. M1 and M2 are independent of Q3. A run phase that cannot proceed on Q3 delivers M1, M2, M3 step 1, M4, M5 and reports AC-MJ-006 as blocked rather than green.

## D. Risks

| Risk | Where it shows | Mitigation |
|---|---|---|
| Core stub over 10,000 units | AC-MJ-012; `TestRoleCoreStubGate` | marker form (+9) plus an 8-unit offset; never the pointer form in K |
| Ledger row left stale | `TestBindingLedgerIntegrity` "obligation lost" | update the row in the same commit as the line |
| A pinned token lost in a scoped rewrite | `TestLaneRecheckDoctrine`, renew-interval test, watchdog tests | `design.md` §D-3 token list checked line by line |
| Unscoped tail line in hard-wrapped A | AC-MJ-008(a) | reflow so marker and `kept` share a line |
| Template-only edit | AC-MJ-001..009 show `:0`/`0` on the deployed operand; `TestAutoPickMirrorParity` | edit template then deployed within the milestone |
| `factory-dispatch.md` fork absorbed or duplicated | AC-MJ-017; `TestAutoPickMirrorParity` | apply the same edits to both copies; never copy the worktree-sweep sentence into the template |
| Injected SessionStart lane notice disagrees with the scoped rule | spec R-1 | out of scope (Q4); the leader's lane notice (AC-MJ-015) names it; a follow-up card is proposed in the report |

## E. Commits (run phase)

One commit per milestone, each naming the card id in its message, explicit pathspecs only (the primary-checkout sweep prohibition applies in spirit inside the worktree too): template and deployed files of one milestone together, `catalog.yaml` and `binding_ledger.json` with the milestone that changes what they pin. The plan-phase commit carries only `.moai/specs/SPEC-CARD-MODE-JUDGMENT-001/*`.
