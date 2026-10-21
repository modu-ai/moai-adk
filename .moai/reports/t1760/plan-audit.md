auditor-model: claude-sonnet-5-5

# SPEC Review Report: SPEC-CONTAINMENT-MAPPING-DOC-001

Reasoning context ignored per M1 Context Isolation (none was passed; only spec.md, plan.md, acceptance.md, progress.md were read).

verdict: FAIL
audited_sha: cf5a52f55992e59fe9593fa508ace5d4dba59fc7
overall_score: 0.66
must_pass_failed: 1
blocking_count: 2
plan_artifact_hash: 0956f1a18204b0dc5762a4d3e2a5de5934f3a60f876616b4d398044568f35fda (auditor-computed sha256 over `cat .moai/specs/SPEC-CONTAINMENT-MAPPING-DOC-001/*.md`, NOT the runtime ComputeHash; see Gaps G6)
scope: full
fix_scope: acceptance.md#E08 acceptance.md#E07 acceptance.md#AC-CMM-007 acceptance.md#AC-CMM-008
defect_class: D1=ac-wording D2=ac-wording
convergence_overall: fail
required_backend: codex fail
required_backend_fail: codex

Iteration: 1/3. Tier M (threshold 0.80). Verdict: FAIL. Plan Artifact Hash: see above. Auditor Version: plan-auditor (sonnet-5-5 in-session, plus codex required gate).

## Must-Pass Results
- [PASS] MP-1 REQ numbering: REQ-CMM-001..011 sequential, no gap or duplicate, spec.md:L63-L73. REQ count 11 and AC count 14 are inside the Tier M ceiling of 16 each.
- [PASS] MP-2 GEARS (judged on the REQ layer in spec.md only): all eleven REQs carry a labelled GEARS/EARS pattern (Ubiquitous, Event-driven L66/L71, State-driven L67, Unwanted-legacy L72). The Given/When/Then entries in acceptance.md are ACs and were not graded here.
- [PASS] MP-3 frontmatter: all 12 fields present with right types, spec.md:L2-L14. `moai spec lint SPEC-CONTAINMENT-MAPPING-DOC-001` printed `0 error(s), 0 warning(s)` (one INFO OwnershipTransitionUnmeasured: commit has no Authored-By-Agent trailer). `mcp__moai__spec_audit` reported era V3R6, `modern_era_clean: 1`.
- [N/A] MP-4: single-repo documentation SPEC, no multi-language tooling.
- [PASS] MP-5 D7: only self-reference `SPEC-CONTAINMENT-MAPPING-DOC-001` (grep -Eo). No `REVIEW:` candidates.
- [PASS] MP-6 D8: `grep -c syscall spec.md` printed `0`.
- [PASS] MP-7: `grep -rn 'NEEDS CLARIFICATION'` over spec/plan/acceptance/progress printed nothing, exit 1; research.md does not exist (Tier M).
- [FAIL] MP-8 RED-now re-execution: E07 (AC-CMM-007, release-blocking) is refused by the worktree-session guard and cannot be re-executed (see D2). Per the MP-8 execution discipline a refused citation is not recorded as a pass and takes the section 2.1 demotion; the SPEC still classifies AC-007 release-blocking and the Definition of Done requires it green. The other release-blocking RED cells reproduce (Claim/Evidence below).
- [PASS, by hand] MP-9: the CN-4 verb was refused by the guard (G3). Read by hand: plan.md M1-M4 are numbered list items (plan.md:L71-L74), there are no `Exit:` lines, and acceptance.md ordering words (L5, L22, L130, L175, L189) order nothing across milestones. No conflict found. Mechanical NONE/CONFLICT was not observed.

## Category Scores
| Dimension | Score | Band | Evidence |
|---|---|---|---|
| Clarity | 0.75 | minor ambiguity | duplicate heading name `Integration into the release branch is self-served` (D9); AC-006 vs REQ-008 (D6) |
| Completeness | 1.0 | all sections | HISTORY spec.md:L19, WHY L25, REQ L59, ACs acceptance.md, `### Out of Scope — ...` x4 spec.md:L86-L103 |
| Testability | 0.50 | several mutant-admitting ACs | D1, D2, D3-D5; E08 and E07 defects measured below |
| Traceability | 0.75 | indirect | all 11 REQs mapped, no orphan (acceptance.md:L75-L87), but REQ-003 "verbatim" and REQ-007 "path § section in the defense cell" have no discriminating check |

## Defects Found
D1. acceptance.md:L153 (E08, AC-CMM-008) — `grep -w` is defeated by Hangul particles. With the Korean body the plan itself assumes (plan.md § B.2), `2026년`, `t1760은`, `develop를` all return count 0, so card ids, years and the word `develop` pass unseen. The observed-failure fixtures (acceptance.md:L175-L191) were ASCII only. Severity: critical — Class: blocking — Required fix: drop `-w`; use explicit ASCII boundaries, e.g. `grep -c -E '(^|[^A-Za-z0-9_])(([Tt][0-9]{3,4})|SPEC-[A-Z][A-Z0-9-]*|(19|20)[0-9]{2}|[Dd][Ee][Vv][Ee][Ll][Oo][Pp])([^A-Za-z0-9_]|$)'`; add Hangul-adjacent cases to the instrument-validation table. (Tested fix: catches all four Hangul cases, leaves `developer`, `development`, `30,000`, `2,500`, `v3.2.0` at 0.)
D2. acceptance.md:L145 (E07, AC-CMM-007) — the command uses awk `system("test -f " p)`; the guard in this very worktree session refuses it ("runs awk with a program that can execute commands"). The run phase runs in a card worktree, so the release-blocking AC cannot be executed as written. Separately: it counts occurrences, not distinct paths (eight repeats of one path give n=8, floor passes, measured), and codex reproduced that a `.NONEXISTENT` suffix after a real `.md` prefix passes. Severity: critical — Class: blocking — Required fix: replace with guard-admissible single-invocation forms (extract paths with grep -o, test existence without `system()`), require distinct paths >= 8, and match the full cited path.
D3. acceptance.md:L109 (E03, AC-CMM-003) — counts matching rows, not distinct IDs; four BP-1 rows give `4`. A fabricated fragment or wrong URL also passes E03/E04 (fixture mut2: all well-formed, exit/counts identical to pass). REQ-003 "verbatim" and REQ-004 are reviewer-only. Severity: major — Class: optional — Required fix: require each of BP-1, BP-2, BP-5, BP-6 exactly once; name the file where the run-phase raw-byte match output is recorded so AC-CMM-011 is checkable.
D4. acceptance.md:L130 (E05, AC-CMM-005) — `§` is accepted anywhere in the row, defense cell not isolated. Fixture mut1 (claim cell `§ x`, defense cell `vague defense`) exits 0 (measured; codex reproduced). Header/column order is not checked though plan.md § B.3 says the commands key on it. Severity: major — Class: optional — Required fix: test the defense-cell field (`$5`) for a path followed by `§`.
D5. acceptance.md:L167 (E10, AC-CMM-010) — `<PLAN_COMMIT_SHA>` placeholder is unfilled and its `<`/`>` make the command non-conforming as written; the recorded red probe `HEAD~1 HEAD ... exit 1` no longer reproduces (now exit 0, measured: HEAD~1..HEAD touches only the four SPEC files); a commit-to-commit diff ignores uncommitted/untracked changes (codex). `629bb51cda73fc70422769988b8c28182c9bd2d5` is already pinned and `git diff --name-only 629bb51cda... HEAD` shows only the four SPEC files, so it is a ready base. Severity: major — Class: optional — Required fix: use that SHA, drop the HEAD~1 probe (moving coordinate), add `git status --porcelain` scope check.
D6. acceptance.md:L35 vs spec.md:L70 — AC-CMM-006 demands at least one `gap` row unconditionally, REQ-CMM-008 permits a gap only when no defense is found; if both seed gaps are resolved the AC forces an unsupported gap or the "gap row saying a defense was found" the AC text hints at. Severity: minor — Class: optional — Required fix: make the AC conditional or state the allowed disposition.
D7. acceptance.md:L195-L201 vs spec.md:L63 — "No file outside the document and this SPEC directory changed" (DoD) is unqualified in time while REQ-CMM-001 binds the run phase only; the sync phase commonly adds a CHANGELOG entry. Severity: minor — Class: optional — Required fix: state E10 is evaluated at the run-phase tip.
D8. acceptance.md:L103 etc. (ledger E02-E09) — the `stdout:` field holds stderr text (`grep: ...: No such file or directory`); here `grep` is a shell function over ugrep and prints `ugrep: warning: ...` on stderr with empty stdout. Exit codes and RED reproduce; text is not verbatim stdout. Severity: minor — Class: optional.
D9. plan.md:L47 / spec cites — `factory-dispatch-mechanics.md § Integration into the release branch is self-served` names two headings (mechanics.md:L103 `##`, L234 `###`), and the same name exists at factory-dispatch.md:L214. Severity: minor — Class: optional — Required fix: pin the line or heading level when citing.
D10. plan.md:L15 — claim that `.moai/docs` siblings are Korean is half true: `factory-managed-session.md`, `gitflow-integration-chain.md` Korean; `git-workflow-doctrine.md`, `version-management.md` English. `language.yaml` has `documentation: ko`, so the Korean-body choice itself stands. Severity: minor — Class: optional.
D11. acceptance.md:L153 — E08 `SPEC-` branch uses `-i`, so `spec-anchored` would count; REQ-CMM-010 forbids dates but E08 only matches four-digit years. Severity: minor — Class: optional.
Codex-only findings (required backend): the four P2 items on E03, E05, E07, E10 are folded into D3, D4, D2, D5 above.

## Recommendation
1. Fix E08 (D1) and E07 (D2) first; re-run both on the pass fixture and on Hangul/duplicate/suffix mutants.
2. Tighten E03, E05, E10 (D3-D5) and resolve D6-D7 wording.
3. Re-audit the delta (acceptance.md E03/E05/E07/E08/E10, AC-006, DoD). CN-4 verb runs in full on re-audit.
Source integrity (below) is sound and needs no change.

## Evidence (Claim / Evidence / Baseline-attribution / Gaps / Residual-risk)

### Claim 1: the four sources and quoted fragments are real
Evidence: `curl -sS -L` raw download, HTTP 200 for all four: bp5 164787 B, bp6 192579 B, bp1 144487 B, bp2 452863 B. `grep -c -F` on raw bytes: all 18 fragments matched, counts identical to plan.md § C (e.g. `bp5 | post-merge and pre-deploy automated testing => 1`, `bp6 | users approved roughly 93% of permission prompts => 1`, `bp1 | the tokens are never reachable from the sandbox where Claude => 1`, `bp2 | Treat a text-only end of turn ... => 2`). Titles matched (bp6 2, bp2 2, others 1). Surrounding-text read: BP-5 says a merge queue "doesn't make much sense at agent scale" and names post-merge testing in place of a global-lock queue; BP-1 "each could fail or be replaced independently" concerns brain/hands/session interfaces, not machine failure domains (analogical row, seed Fit `differs`).
Baseline-attribution: this run, pages as served 2026-10-16; no tree dependency.
Gaps: WebFetch is not available in this auditor session, so the WebFetch cross-check was not run; raw-byte matching (the evidence the plan itself calls verbatim) was independent.

### Claim 2: cited MoAI defense sections exist in this worktree
Evidence: `grep -n -E '^#+ ...'` found all 15 headings, e.g. `factory-dispatch.md:90 ### The delegation channel is the queue`, `:134 ## Completion is read, never trusted`, `:166 ## Isolation is provisioned by MoAI...`, `auto-semantics.md:113 ### 5.1 The standing awaken carrier`, `:206 ## 9. The gate inventory and dispositions`, `factory-dispatch-gates.md:55 ## Verification load incident record`, `factory-dispatch-mechanics.md:16 ### Serializing a heavy run across lanes`, `:103` and `:234` (duplicate, D9), `cross-session-messaging-core.md:56`, `main-checkout-branch-guard.md:41 ## Mechanical Enforcement`, `worktree-integration.md:161`, `:194`, `moai-constitution.md:132 ## Security Boundaries`, `gitflow-lane-protocol.md:51 ## 3. Serialization`, `goal-directive.md:11`. Content checks: mechanics.md:L114 states FIFO, remeasure before queue, "seconds-long identity check plus git merge --no-ff ... no test suite inside the window"; L120 conflicts bullet; gates.md:L57 four lanes full suite, load 413; branch-guard text says opt-in, default false. Seed rows match.
Baseline-attribution: this worktree, HEAD cf5a52f55.
Gaps: auto-semantics.md body, worktree-integration.md body, resource-slot-lease.md were checked for heading only.

### Claim 3: RED-now cells reproduce (MP-8)
Evidence (document absent, tree HEAD cf5a52f55 differing from the pin only in the four SPEC files: `git diff --name-only 629bb51cda73fc70422769988b8c28182c9bd2d5 HEAD` lists spec/plan/acceptance/progress):
- E01 `test -f ...` exit 1, empty. E02 `grep -c -i -F 'reference-only' ...` stdout empty, stderr `ugrep: warning: ...No such file or directory`, exit 2. Same exit 2 for E03, E04a, E06, E08, E09. E05 awk: `awk: can't open file ...` exit 2. E04b analog of E04a, not separately run. E07: refused (D2).
- E10 green `git diff --quiet cf5a52f55 HEAD -- . ':!doc' ':!spec-dir'` exit 0 (guard accepts this plain form). E10 red probe from the ledger `HEAD~1 HEAD` now exit 0, does not reproduce (D5).
Baseline-attribution: this run, this tree.
Gaps: E07 not executed. All nine REDs are the same "file absent" cause, so they say nothing about each check's discrimination; discrimination was tested below.

### Claim 4: the checks discriminate (mutant probe)
Evidence on fixtures kept outside the repo (scratchpad `fx/`): E03 = 4, E04a/E04b = 4/4, E06 = 1, E05 exit 0 on the pass fixture. Mutants: mut1 (claim-cell `§` only, vague defense cells, one path repeated 8x) -> E05 exit 0, E07-style count n=8 distinct=1 floor-passes=1. mut2 (fabricated fragment, wrong URL) -> E03=4, E04 4/4. Hangul probe: `2026년`, `t1760은`, `develop를` all `0` under ugrep and /usr/bin/grep; fix pattern returns `1,1,1,1`.
Gaps: E07 itself could not run on any fixture; only its occurrence/distinct logic was reproduced without `system()`.

### Claim 5: template-deployment and scope findings
Evidence: `find internal/template/templates/.moai/docs -maxdepth 1 -type f` lists exactly six files (agent-lint, audit-artifact-convention, generic-patterns-guide, hook-stdin-fail-closed, session-handoff-appendix, todo-queue-storage); `.moai/docs` has 37 files; `ls ... | grep -c -E "workflow-doctrine|integration-chain|managed-session"` -> 0; gitflow-integration-chain.md:L3 declares the doc local-only ("로컬 전용 문서 — 템플릿에 미러하지 않는다"); `git check-ignore` exit 1 (not ignored). No `.moai/docs` file or doc matches `contain|mapping|multi-agent` before the SPEC. Premise grep for the three engineering URLs across `.moai/docs .claude/rules .claude/skills AGENTS.md docs-site/content` printed nothing, exit 1. `github_flow_sweep_guard_test.go` has moai-docs floor 33 (37 now); armed only under github-flow; the repo is git-flow. This supports no Template-First mirror; REQ-CMM-011 stands.
Gaps: the guard's scan function body (the exact `develop` pattern) was not read; `moai update` behaviour for untracked-by-template docs not measured (the plan states the same gap).

### Claim 6: tool provenance and cross-model gate
Evidence: `moai verify audit-plan --project-root .../t1760`: `cross_model_active: true`, `enforced_required: ["codex"]`, claude off, glm off. `mcp__moai__audit_multi` (no gates, project_root = this toplevel, target baseBranch, focus "acceptance command soundness and evidence integrity") returned `per_backend_verdicts: [codex required fail]`, `overall_verdict: fail`, `participant_count: 1`, `disagreement_flag: null`, `residual_risk_note: required-backend FAIL: codex`, `audit_receipt: rcpt-8085bfbb3b82e8d5d10b24a8`, `plan_source: config`, `build_commit: ee981a932`. Digest written to `.moai/state/audit-plan-result.json`, then `moai verify audit-plan --result-file ...` returned `convergence_check: {ok: true, unmet: []}` (the gate was met; the backend answered fail). The installed `moai` is `v3.2.0-rc.36-t1822c`, build commit `ee981a932` (`-dirty`); `git merge-base --is-ancestor ee981a932 HEAD` exit 1, so it is not a strict ancestor of HEAD; no tree-built binary exists (`bin/moai` absent). Lint and audit were judged by that installed build and the same-build MCP server.
Gaps: none for the required gate.

### Gaps (consolidated)
G1 WebFetch and the Grep tool are not available in this session; Bash `grep` (a shell function over ugrep) and curl were used instead. G2 The worktree-session guard refused: compound git/shell with function definitions or heredocs, `sh <script>`, `awk -f`, and awk with `system()`; these were split into plain commands. G3 Consequently the Group-4 traceability verb and the CN-4 verb were not executed mechanically; traceability and ordering were read by hand (matrix acceptance.md:L75-L87, M1-M4 plan.md:L71-L74). G4 E07 never executed. G5 `moai spec lint` output is a statement about the installed build only. G6 The runtime plan-artifact hash was not computed (no admitted verb used); the value above is a plain sha256. G7 The codex findings on E03/E05/E07/E10 were reproduced by me for E05 and the E07 counting logic only; E03 uniqueness and E10 uncommitted-change blindness follow from the commands' structure and codex's run.

### Residual-risk
A rewritten E08/E07 could still admit mutants not tried here (HTML-entity or full-width digits, a path in an anchor). Paraphrase drift in defense cells remains reviewer-only (AC-014). Upstream pages may change before run; the author's re-fetch handles that. codex residual note: `required-backend FAIL: codex`.

## Operational Notes (unverified)
- measured: Hangul-adjacent tokens evade `grep -w`; command `grep -c -w -i -E 't[0-9]{3,4}|...|develop' fx/y.md fx/c.md fx/d.md` printed 0 for each of the three (see Claim 4).
- inferred: run phase will hit the same worktree-guard refusal on any awk containing `system()` (same guard family as observed here); measure by running the final E07 inside the card worktree before the run begins.
- assumption: replacing E10's base by the pinned SHA keeps the guard green until the run phase edits only the document and the SPEC dir.

AUDIT-VERDICT: FAIL spec=SPEC-CONTAINMENT-MAPPING-DOC-001 receipts=rcpt-8085bfbb3b82e8d5d10b24a8
