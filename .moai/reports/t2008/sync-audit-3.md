auditor-model: claude-opus-5-5
verdict: FAIL
score: 0.81
audited_sha: 9754a838781cca589c730fe7bbbcb1965799b484

# Sync audit 3 (fix-confirm delta): card t2008 (Class B, no SPEC), PR #1961 (draft), branch WT-run-tdd-result

- Tree: /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t2008
- Scope: the F1-F3 defect delta of `sync-audit-2.md` (FAIL 0.74 @0e010a2d3f). Repair commits `git log e4d693349f..HEAD`: `a6f9fdc927` (RED), `32c93279ad` (GREEN), `9754a83878` (record). Delta diff: 12 files, +47/-27.
- HEAD and working tree, at the start and again just before this file was written: `git rev-parse HEAD` returned `9754a838781cca589c730fe7bbbcb1965799b484` and `git status --short` printed nothing.
- This file is untracked and ignored (`git check-ignore -v` → `.gitignore:235:.moai/reports/*`). Not staged, not committed.
- card-review.md: absent by decision. Reason: "Codex cost hold; independent audit Codex leg covers".
- One round only (the pre-approved delta round). Nothing spawned.

## Overall verdict

FAIL, 0.81. The required Codex gate returned `fail`, and that leg is fail-closed. I reached the same finding on my own before the Codex result arrived (N1).

- F1 is repaired.
- F3 is repaired.
- F2 is only partly repaired. The rule now asks for `HEAD` and `plan_artifact_hash` on every phase checkpoint line. But the three checkpoint templates the phase bodies actually tell the orchestrator to write still carry neither field. An orchestrator that follows those templates writes checkpoints with no identity, and the Resume Check then invalidates those phases on every resume.

All the scoped mechanical checks pass. The failure is that the text is incomplete, not that a check is red.

## Disposition of the sync-audit-2 blocking findings

| ID | Status | Evidence (live files at 9754a83878) |
|---|---|---|
| F1 identity block on every run | Repaired | `context-loading.md:164`: "Identity record (every run, whether the file was just created or already existed; the plan phase creates progress.md without these fields): if the block below is absent, append it now before the first phase runs." This now sits outside the "does not exist" branch, so the normal path (progress.md already created by manager-spec) reaches it. Pinned by the literal `"whether the file was just created or already existed"` in `run_resume_identity_test.go:24`. |
| F2 per-phase checkpoint values | Partly repaired, see N1 | `context-loading.md:172` prescribes `- Phase {N} complete: HEAD {sha}, plan_artifact_hash {hash}`, and `:157` compares HEAD "recorded on that phase's checkpoint line". The shared overwrite sentence is gone. But `phase-execution.md:218/392/412` still prescribe checkpoint lines without those fields. |
| F3 retired phase labels | Repaired | The retired-label grep over run.md and run/*.md printed nothing (`grep_exit=1`). The frontmatter now reads `Phase 1-10` / `Phase 11-20`, and these match the headings (`## Phase 1:`..`## Phase 10:` in phase-execution.md, `Phase 11`..`Phase 20` in task-decomposition.md). The four hand-offs now name Phase 2, Phase 2, Phase 2 and Phase 3, which is correct: the gate is Phase 1, Environment Assessment is Phase 2, and the next phase after it is Phase 3. The checker now searches the router, task-decomposition, context-loading and phase-execution for all eight new literals (`run_phase_reference_test.go:46-56`). |

## Your question: is the manager-spec / progress.md ordering now covered end to end?

Partly.

- The producer is now reachable on the normal path. manager-spec creates progress.md without identity fields. Then `/moai run` takes the "exists" branch at `:157`. It finds no checkpoints, so it skips nothing. Then `:164` appends the identity block because it is absent. On a later resume the block exists and is compared. So the gap F1 named is closed.
- The end-to-end chain still breaks at the per-phase checkpoint writers (N1). The two smaller timing and wording defects (N2, N3) fail safe and are recorded as debt.

## Findings (structured defect list)

- **N1 [Medium] [blocking] `.claude/skills/moai/workflows/run/phase-execution.md:218, :392, :412` (with `context-loading.md:172`). The per-phase checkpoint templates do not carry the identity values F2 requires.** Confidence: high. Codex reported the same independently (P2, high confidence). It ran an overlay test that failed with 6 missing fields, and passed once the fields were added.
  - The three templates are verbatim:
    - `- Phase 2 complete: memory_guard={enabled|disabled}, available_mb={N}, strategy={full|module|changed}`
    - `- Phase 7 complete: {N} acceptance criteria registered as pending tasks`
    - `- Phase 8 complete: {N} stub files created, LSP baseline captured`
  - These templates and `:172` prescribe two different formats for the same `- Phase {N} complete:` line.
  - Effect: an orchestrator that follows the phase body writes checkpoints for Phases 2, 7 and 8 with no HEAD and no hash. Under `:157` ("invalidate a phase ... that carries an unrecorded identity field ... and every later phase that consumed its output"), every resume re-runs from Phase 2. This fails safe (nothing is wrongly skipped), but resume is still unusable on the standard path. That is the same failure class as the original F1.
  - The test does not see it: `TestRunResumeIdentity` pins only the literal `- Phase {N} complete: HEAD` in context-loading.md.
  - Required fix:
    - Add `HEAD {sha}, plan_artifact_hash {hash}` to the three templates, e.g. `- Phase 2 complete: HEAD {sha}, plan_artifact_hash {hash}, memory_guard=...`.
    - Mirror the change into `internal/template/templates/` and regenerate the catalog hash.
    - Extend the test so that every `- Phase {N} complete:` template in the run bodies must carry both fields, and keep a mutant for a template missing them.

- **N2 [Low] [optional] `context-loading.md:156/164/170` and `:157`. The timing and comparison wording is inconsistent.**
  - The Resume Check runs "Before Phase 5" (`:156`, and `:133` "created at Phase 5 start"). The identity block is appended "before the first phase runs" (`:164`). Its `plan_artifact_hash` is "recorded by the Plan Audit Gate", which is Phase 1.
  - If the block is written at Phase 0, the hash is `unknown`, which counts as an unrecorded field.
  - `:157` binds HEAD to "that phase's checkpoint line" but compares the hash against "the recorded one", without saying which record.
  - Both readings fail safe.
  - Fix: say "at the Resume Check (before Phase 5)", and "the plan_artifact_hash recorded on that phase's checkpoint line".

- **N3 [Low] [optional] `context-loading.md:164`. The identity block is written only when absent, and never re-recorded.**
  - Repository root, worktree and lease owner live only in the block.
  - After a legitimate worktree move or a card re-lease (for example the 8859 lease recovery), every later resume differs on those fields and invalidates every phase, even after the phases are re-run in the new tree.
  - This fails safe. Fix: once invalidated phases are re-run, re-record the block (or move these fields onto the checkpoint line too).

- **N4 [Low] [optional] `internal/template/run_phase_reference_test.go:96-120`. The mutant test does not pin the eight new stale literals.**
  - The missing literals are `Phase 11-4`, `Phase 1-1.8`, `0.5~1.8` and the four `... Phase 1.` hand-off forms.
  - The hand-off check in phase-execution.md is a literal list, not a check against existing headings as sync-audit-2 suggested. For example, `skip to Phase 1 ` with no period, or `proceed to Phase 21`, would pass.
  - The deployed test does prove the current text is clean.

- **N5 [Info] [optional] `context-loading.md:109, :112`.** `Skip phases [0, 0.6, 2.0, 2.5, 2.75, 2.8a, 2.9, 2.10]` and `(Phase 10 + 2.8a)` are decimal labels from the old numbering. They predate this card (`git show origin/main:... | grep -n 2.8a` shows the same two lines), so they are out of scope. Recorded so they are not lost.

- **Carried optional debts F4-F7 (re-judged)**
  - F4: closed. `:166` now reads `{git rev-parse --show-toplevel of the tree the run executes in}`, so the self-contradiction is gone.
  - F5: still open, Low. `phase-execution.md:94-95` logs `verdict=PASS` for an admitted PASS-WITH-DEBT.
  - F6: still open, Low, fails safe. `--skip-audit` leaves `plan_artifact_hash` uncomputed, so the hash becomes `unknown` and the phase is invalidated.
  - F7: unchanged, out of scope. The `.moai/workflows/run*.md` snapshot drift predates the card.

## Claim / Evidence / Baseline-attribution / Gaps / Residual-risk

**Claim 1. The scoped template tests pass.**
- Evidence: `go test -count=1 -run 'TestRunPlan|TestRunResumeIdentity|TestRunPhaseReferences|TestSkillsContainPlanAuditGateMarkers|Mirror|Embed|Manifest|DateProvenance|BindingLedger' ./internal/template/` printed `ok  	github.com/modu-ai/moai-adk/internal/template	3.659s`.

**Claim 2. internal/skills passes, including the router LOC ceiling.**
- Evidence: `go test -count=1 ./internal/skills/` printed `ok  	github.com/modu-ai/moai-adk/internal/skills	0.085s` and `exit=0`.
- Evidence: `wc -l` printed `199 .claude/skills/moai/workflows/run.md` and `199 internal/template/templates/.claude/skills/moai/workflows/run.md`. The ceiling is 200.

**Claim 3. The TDD result contract hook test passes.**
- Evidence: `bash .claude/hooks/tests/test-tdd-result-contract.sh` printed `PASS: TDD distinguishes expected RED from regression and tool failures`.

**Claim 4. The live skill files are byte-identical to their mirrors.**
- Evidence: `cmp` printed `run.md=0`, `run/context-loading.md=0`, `run/phase-execution.md=0`, `run/task-decomposition.md=0`.

**Claim 5. No retired phase label remains in run.md or run/*.md.**
- Evidence: `grep -nE 'Phase 11~4|Phase 11-4|Phase 1-1\.8|0\.5~1\.8|Phase 1~1\.8|Phase 1\.8|11/2B|or 2B|2\.8[^a]|(entering|proceed to|continue to|skip to|proceeding to) Phase 1([^0-9]|$)' .claude/skills/moai/workflows/run.md .claude/skills/moai/workflows/run/*.md` printed no matches and `grep_exit=1`.
- Evidence (frontmatter): `phase-execution.md:2: description: "Run Phase 1-10 — ..."`, `:6: phase: "Phase 1-10: Phase Sequence and Preparation"`, `task-decomposition.md:2: description: "Run Phase 11-20 — ..."`, `:6: phase: "Phase 11-20: Implementation, Quality Validation, and Completion"`.
- Evidence (headings): `grep -nE '^#+ Phase [0-9]+'` listed `## Phase 1:` through `## Phase 10:` in phase-execution.md and `Phase 11` through `## Phase 20:` in task-decomposition.md.

**Claim 6. Static checks on the two test files are clean.**
- Evidence: `gofmt -l internal/template/run_phase_reference_test.go internal/template/run_resume_identity_test.go` printed nothing and `gofmt_exit=0`.
- Evidence: `go vet ./internal/template/` returned `vet_exit=0`.

**Claim 7. N1 is real.**
- Evidence: `sed -n` over phase-execution.md printed `218:- Phase 2 complete: memory_guard={enabled|disabled}, available_mb={N}, strategy={full|module|changed}`, `392:- Phase 7 complete: {N} acceptance criteria registered as pending tasks` and `412:- Phase 8 complete: {N} stub files created, LSP baseline captured`.
- Evidence: `grep -n "complete:"` across run bodies shows that `context-loading.md:172` is the only checkpoint form that carries `HEAD`.
- Codex reproduced it with an overlay test.

**Claim 8. Security: the delta adds no attack surface.** The delta changes only Markdown prose, two `_test.go` files and one catalog hash (`git diff --stat e4d693349f..HEAD`). No production Go changed, and no secret, exec or network surface was added.

**Baseline-attribution.** Every command above ran in this session against tree `9754a838781cca589c730fe7bbbcb1965799b484` with a clean working tree, from `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t2008`, using the Go toolchain on PATH.

**Judging build (VCI 2.2)**
- Installed `moai`: `moai version` → `v3.2.0-rc.39   v3.2.0-1041-gdad3603de1   built 2026-11-02T19:17:51Z`.
- MCP server that ran `audit_multi`: `build_commit: b1806498fe` (v3.2.0-rc.38, built 2026-10-21).
- `git merge-base --is-ancestor dad3603de1 HEAD` → `installed_ancestor_exit=1`, and `git merge-base --is-ancestor b1806498fe HEAD` → `mcp_ancestor_exit=1`. Both commits exist (`git cat-file -t` → `commit`), so neither build is a strict ancestor of HEAD.
- These builds judged only the audit plan and the Codex leg. The tests ran from tree source through `go test`.

**Gaps**
1. The worktree-isolation guard refused two compound Bash forms:
   - the scoped `go test` with a `| tail; echo ${pipestatus}` suffix;
   - the hook test with the same suffix.
   I re-ran both as plain single commands, and the outputs in Claims 1 and 3 come from those plain runs. The plain runs did not print an explicit exit code: `ok` from `go test` and the `PASS:` line from the hook are the evidence.
2. I did not run `/moai run` or a resume end to end. N1, N2 and N3 come from reading the instructions. Codex's overlay test covers N1 at text level only, and Codex marked full resume behavior NOT_RUN.
3. I did not run `go test ./...`, `./internal/cli`, golangci-lint or `make build` (out of scope by instruction). I did not read CI at this head.
4. The Codex leg was scoped through `delta_base=e4d693349f` in prose only, not mechanically bounded (tool note). Claude and GLM are gated `off` by config. I am not counting them as skipped legs.
5. `audit_multi` with `delta_base` writes no second-review record and no session state (tool note).

**Residual-risk**
- R1. All the new clauses are pinned by literal-presence tests, so real orchestrator behavior on resume is not exercised.
- R2. Customer projects receive the wording only on `moai update`.
- R3. `spec-workflow.md` still carries the legacy Grace Window paragraph (deferred by ruling).
- R4. Even after N1 is fixed, a phase body added later without the two fields would bring the gap back unless the test scans every checkpoint template, which is why the N1 fix includes that test.

## Dimension scores (flat weighted mode; built-in default profile, since no SPEC means no evaluator_profile)

| Dimension | Score | Verdict | Evidence |
|---|---|---|---|
| Functionality (40%) | 74/100 | FAIL | F1 and F3 repaired (Claim 5, disposition table). F2 is partial: three checkpoint producers still lack the identity fields (N1, Claim 7). |
| Security (25%) | 95/100 | PASS | Claim 8: the delta is prose, tests and a hash. |
| Craft (20%) | 80/100 | PASS | Claim 6 clean, Claims 1-3 green. Literal-only checker, and the new labels have no mutant pins (N4). |
| Consistency (15%) | 88/100 | PASS | Claim 4 mirrors byte-identical. Wording inconsistencies N2 and F5. |

Must-pass firewall: Functionality fails, so the overall result is FAIL regardless of the other scores. The required Codex gate is also FAIL.

## Binding run conditions

None apply. There is no SPEC progress.md and no `Binding run conditions` heading.

## Per-backend verdicts (audit_multi digest)

- overall_verdict: `fail`
- gate_unmet: not present in the result. `residual_risk_note`: "delta_base e4d693349ffcdd08d38e691bf397e80c5f3fb5e1: no claude or glm leg produced a verdict on it; codex was scoped in prose only (not mechanically bounded); no second-review record and no session state written | required-backend FAIL: codex"
- plan_source: `config` (`moai verify audit-plan`: model codex; claude off, codex required, glm off; `cross_model_active: true`)
- per_backend_verdicts:
  - backend `codex`, gate `required`, verdict `fail`. One P2 finding: context-loading.md:172 against phase-execution.md:218/392/412, the same as N1.
- participant_count: 1; fail_open_backends: []; disagreement_flag: null
- tree_root: `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t2008`; build_commit: `b1806498fe`
- audit_receipt: `rcpt-7adecf2d9fbf14de79cdc95f`

This verdict is not final until the orchestrator writes the digest to `.moai/state/audit-plan-result.json` and runs `moai verify audit-plan --project-root <toplevel> --result-file <path>`. Here, the required-backend fail already makes the convergence result `fail`.

## Iteration history

- sync-audit-1: judged `d0177299cf`, PASS-WITH-DEBT 0.88 (D1, D2). Not adopted.
- sync-audit-2: judged `0e010a2d3f`, FAIL 0.74 (F1-F3 blocking), receipt `rcpt-91596827d55973873e76a231`.
- sync-audit-3 (this file): judged `9754a83878`, FAIL 0.81. F1 and F3 repaired; F2 partial (N1 blocking). This was the one pre-approved delta round. Another round needs a new approval. The scope for that round would be the N1 delta: three template lines, their mirrors, the catalog hash and one test extension.
