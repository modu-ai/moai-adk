auditor-model: claude-sonnet-5-5

# SPEC Review Report: SPEC-CONTAINMENT-MAPPING-DOC-001
Iteration: 4 (delta on round 3 defects N3, N4; CN-4 and MP-8 re-run in full)
Verdict: FAIL
Overall Score: 0.92
Plan Artifact Hash: 4d2823d08e46311e7830c6daa626282f5bf75495c7afbd4bda7d3a8fdf93f77c
Auditor Version: plan-auditor/v1 (Tier M threshold 0.80)

verdict: FAIL
audited_sha: a45ba6eb8424546f8d54db6ca10efc58f71d3412
overall_score: 0.92
must_pass_failed: 0
blocking_count: 1
plan_artifact_hash: 4d2823d08e46311e7830c6daa626282f5bf75495c7afbd4bda7d3a8fdf93f77c (auditor-computed `cat .moai/specs/SPEC-CONTAINMENT-MAPPING-DOC-001/*.md | shasum -a 256`; not the runtime ComputeHash, see Gaps G6)
scope: delta
fix_scope: acceptance.md#AC-CMM-010 acceptance.md#E10b acceptance.md#Evidence-ledger-E10b-note acceptance.md#Guard-measurements
defect_class: N5=ac-wording
reread_hunks: acceptance.md#AC-CMM-010 acceptance.md#E10b
convergence_overall: fail
required_backend: codex fail

Reasoning context ignored per M1 Context Isolation (the card prompt's description of fix round 3 was used only to know which commands to re-run; every finding below rests on the artifacts and on commands run in this audit).

The aggregate (0.92) is above the Tier M threshold (0.80) and above round 3 (0.90), so no STOP signal applies. The verdict is FAIL because the required codex backend returned `fail` (one P2 finding, receipt rcpt-25166887c6e4413d16283e54) and my independent reproduction confirmed it (N5 below). `moai verify audit-plan --result-file` returned `convergence_check.ok: true, unmet: []`: the gate was met (codex ran and answered), and the answer was fail. A required-gate failure is not absorbed into the aggregate.

## Prior-defect status

| Id | Status | Basis |
|---|---|---|
| N3 (`.moai/reports` exclusion hid other cards' tracked evidence) | RESOLVED | Throwaway repo: deleting another card's tracked file gives narrowed E10a exit 1 and E10b ` D .moai/reports/t9999/verdict.md`; the round-2 form `':!.moai/reports'` gives exit 0 on the same state; sibling-prefix dir `t17601` deletion is seen (exit 1, so the pathspec matches at a directory boundary); deleting only this card's file is invisible (exit 0, by design). Real tree: E10a exit 0, E10b empty. REQ-CMM-001, AC-CMM-010, header pin paragraph, Guard measurements, DoD and plan.md M4 all carry the `.moai/reports/t1760` scope. |
| N4 (AC-CMM-008 promised any-case months, E08 spells only May/MAY) | RESOLVED | AC-CMM-008 now states May counts only as `May`/`MAY` and lists mixed-case May under Known limit. Verbatim E08 on a 20-line hit fixture (incl. `OcToBeR 16`, `oCt 16`, `16th October`, `1st Dec`, `May 4`, `MAY 4`, `4 May`, Hangul-attached forms, `10-16`, `10/16`, `3월`): count 20. 19-line zero fixture (`may 4`, `mAy 4`, `maY 4`, `4 MAy`, `4 may`, `developer를`, `development`, `30,000건`, `2,500건`, `v3.2.0은`, `spec-anchored`, `Mark 5`, `5 Marketing`, the quoted source fragments, the two source URLs): count 0. Edge fixture: only `sections 10-12 of the guide` counts (documented Known limit); `Oct-16`, `Oct16`, `the 16th of October`, `October, 16`, `10.16`, `20261016` uncounted, as listed. Wording and command now agree. |
| N1 | no regression | E10a exit 0 / E10b empty on the real tree at HEAD a45ba6eb8 |
| N2 | no regression | `october 16`/`OCTOBER 16`/`16 October`-class shapes counted (see N4 row) |
| D1 | no regression | `t1760은`, `2026년`, `develop를`, `SPEC-CONTAINMENT-MAPPING-DOC-001을` counted |
| D2 | no regression | E07 on the passing fixture exit 0; one-path mutant `TOO-FEW-DISTINCT n=1`; `.md.NONEXISTENT` `BAD-SUFFIX`; missing path `MISSING`; seven distinct `TOO-FEW-DISTINCT n=7`, each exit 1 |

## Must-Pass Results

- [PASS] MP-1 REQ numbers: REQ-CMM-001..011 sequential (spec.md:L66-L76). Traceability verb: `COLLECTED: 11 REQ definitions (acceptance input: read)`, no `UNCOVERED:`/`ORPHAN:` lines.
- [PASS] MP-2 GEARS (requirement layer, spec.md REQ-CMM-001..011): each REQ is Ubiquitous/Event-driven/State-driven or a legacy Unwanted form (REQ-CMM-010 "shall not"), labelled in the text. ACs are Given/When/Then in acceptance.md (verification layer, not graded here).
- [PASS] MP-3 frontmatter: all 12 canonical fields present, correct types (spec.md:L1-L15); `tier: M` optional field present; no rejected aliases.
- [N/A] MP-4: single-document, non-tooling SPEC; no language-specific tool named.
- [PASS] MP-5 D7: the only SPEC id in spec.md is its own (`grep -Eo` output: `SPEC-CONTAINMENT-MAPPING-DOC-001`); no reconciliation candidate.
- [PASS] MP-6 D8: the D8 awk verb printed nothing (no `syscall` in spec.md).
- [PASS] MP-7: `grep -rn '\[NEEDS CLARIFICATION' plan.md` exit 1, no match (no research.md in Tier M).
- [PASS] MP-8: release-blocking criteria AC-CMM-001..009 (E01-E09). Each ledger command re-executed on the current tree (doc absent): E01 exit 1, stdout empty; E02 exit 2 (ugrep missing-file warning); E03 exit 2 (awk can't open file); E04a/E04b exit 2; E05 exit 2; E06 exit 2; E07 exit 2; E08 exit 2; E09 exit 2. All reproduce the recorded RED (stdout, stderr, exit). None is a test runner. E10a/E10b are labelled regression-guard (not release-blocking) with probe evidence; they were also executed (see N3 row).
- [PASS] MP-9: CN-4 verb `COLLECTED: 4 milestones in plan order (M1 M2 M3 M4), 12 exit bindings, 8 ordering candidates`; no `CONFLICT:` line. Candidates read: acceptance.md:258 ("AC-CMM-010 ... run-phase tip ... before the sync phase begins") orders against the sync phase, which the plan does not schedule as a milestone; plan.md M4 Exit binds AC-CMM-010 at the run tip, consistent. The other seven candidates (L7, L24, L53, L136, L219, L228, L243) order nothing across milestones.

## Category Scores

| Dimension | Score | Rubric Band | Evidence |
|-----------|-------|-------------|----------|
| Clarity | 0.90 | between 0.75 and 1.0 | AC-CMM-008 and AC-CMM-010 now say what their commands do, except the ignored-file claim (N5) |
| Completeness | 1.0 | 1.0 | HISTORY spec.md:L19-L24, Why 1.1, Requirements, Out of Scope H3s spec.md:L89, L94, L99, L103 with `-` bullets, frontmatter complete |
| Testability | 0.90 | 0.75-1.0 | Every AC binary; AC-CMM-010 claim overreaches by one class (N5) |
| Traceability | 1.0 | 1.0 | 11/11 REQs mapped (Traceability table acceptance.md:L77-L89), verb output clean |

## Defects Found

D1. N5 — acceptance.md:L53 (AC-CMM-010) and L205, L212 (E10b ledger) — AC-CMM-010 states "a change under any other card's directory under `.moai/reports/` is seen by both commands (observed on a throwaway repo)", and the E10b ledger records `untracked .moai/reports/t9999/new.md added (another card's directory): stdout ?? ...`. In this repository `.gitignore:235` is `.moai/reports/*`, so a NEW file under another card's directory is ignored and `git status --porcelain --untracked-files=all` does not list it. Reproduced: `git check-ignore -v .moai/reports/t9999/new.md` printed `.gitignore:235:.moai/reports/*	.moai/reports/t9999/new.md` (exit 0); the required codex backend reproduced the same in a copy of the real `.gitignore` (E10a exit 0, E10b empty). The ledger probe ran in a throwaway repo that carries no such ignore rule, so its `??` line does not transfer to this tree. The claim is wider than the command, on exactly the axis N3 was about. Deleted/modified tracked files (the N3 case) are seen; added ignored files are not. — Severity: major — Class: blocking (a criterion the document states is false as written) — Required fix: narrow the AC-CMM-010 sentence and the closing note after the ledger to "a modified, deleted, or newly tracked/non-ignored file under any other card's directory is seen"; add to AC-CMM-010 (and Known limit) that an added file that `.gitignore` ignores (`.moai/reports/*`, including a new file under another card's directory) is invisible to both commands; mark the throwaway `??` probe as taken without the repository's `.gitignore`. No command change needed (ignored files are machine-local, not repository content).

D2. O1 — acceptance.md:L53 — "In this tree two files under `.moai/reports/t1760/` are tracked" and "1789 tracked files": measured now 4 and 1791 (`git ls-files .moai/reports/t1760`: plan-audit.md, plan-audit-r2.md, plan-audit-r3.md, progress.md; `git ls-files .moai/reports | wc -l` 1791). The figures are attributed to earlier trees, so no verdict rests on them — Severity: minor — Class: optional — Fix: add "at tree 04890fc" to each count or drop the counts.

D3. O2 — plan.md:L157 (carried residual) — merging `main` into the card branch before landing would put other cards' files into the diff against the fixed base `629bb51cd...` and turn E10a red at the run tip. Inference from the fixed-base design; the landing sequence was not measured here — Severity: minor — Class: optional — Fix: none required; if the run phase hits it, the lane resolves it by recording the merge-base change and re-pinning (verification-completeness.md §4 re-measure on rebase).

## Regression Check (Iteration 2+)

- N3: RESOLVED (above). N4: RESOLVED (above). N1, N2, D1, D2: no regression.
- Round-3 optional O5 (template-mirror check only at run tip) and O6 (`10-12` false positive): unchanged and dispositioned in plan.md § J.1; O6 re-observed (count 1).

## Recommendation

1. D1 (N5): fix the AC-CMM-010 sentence "a change under any other card's directory under `.moai/reports/` is seen by both commands" to the tracked/non-ignored scope, add the ignored-file limit, and mark the E10b `??` probe as taken without the real `.gitignore`. Hunk-limited: acceptance.md AC-CMM-010 and the E10b note. Cross-layer sweep (verification-completeness.md §3): spec.md REQ-CMM-001 binds "no file ... outside"; a git-ignored file is not repository content, so REQ-CMM-001 needs at most a clarifying word ("tracked or non-ignored file"), not a new command.
2. Optional D2/D3 as stated.
3. Next audit can be `scope: reread` of the two hunks plus the full CN-4 verb.

## Evidence (Claim / Evidence / Baseline-attribution / Gaps / Residual-risk)

### Claim
C1. N3 is resolved. C2. N4 is resolved. C3. N1, N2, D1, D2 do not regress. C4. MP-1..MP-9 hold. C5. The revised AC-CMM-010 wording promises more than E10a/E10b do for ignored added files (N5).

### Evidence (commands run in this audit; outputs verbatim unless trimmed)
Tree: `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1760`.
- `git rev-parse HEAD` -> `a45ba6eb8424546f8d54db6ca10efc58f71d3412`.
- `moai verify audit-plan --project-root <tree>` -> `config_status ok`, `cross_model_active: true`, `enforced_required: ["codex"]` (claude off, glm off).
- `mcp__moai__spec_audit` (project_root=<tree>, filter SPEC-CONTAINMENT-MAPPING-DOC-001) -> `{"drift_findings":[{"finding_type":"EraAutoDetected","severity":"INFO","era":"V3R6","details":{"heuristic_matched":"H-5 (modern phase or created date)"}}],"modern_era_clean":1,"total_specs":1}`.
- E10a (verbatim ledger command) on the real tree: `E10a exit=0`. E10b: empty stdout, `E10b exit=0`.
- Throwaway repo (scratchpad `probe4r4`, base commit 81cd122b718e983c232b7d4e74f7fa5d9b049c90 holding `.moai/reports/t9999/verdict.md`, `.moai/reports/t1760/plan-audit.md`, `src/keep.txt`):
  - clean tree: `clean E10a=0`, E10b empty.
  - `rm .moai/reports/t9999/verdict.md`: `narrow form E10a=1`, `round2 form E10a=0`, E10b ` D .moai/reports/t9999/verdict.md`.
  - second commit ab5f148f5413baa5bedaa43d49f31fb9e11996dc adds `.moai/reports/t17601/v.md`; deleting it: `t17601 deletion narrow E10a=1`.
  - only this card's `.moai/reports/t1760/plan-audit.md` deleted: `own-card-only deletion E10a=0`; `git status --porcelain` shows ` D .moai/reports/t1760/plan-audit.md` (hidden from E10b by the pathspec, by design).
- E08 verbatim (extracted from the ledger by script, run with `eval` against fixtures `r4_hit.txt`, `r4_zero.txt`, `r4_edge.txt` in the scratchpad): `r4_hit lines=20 E08 count=20`; `r4_zero lines=19 E08 count=0`; `r4_edge lines=7 E08 count=1` (`grep -n` form: `1:sections 10-12 of the guide`).
- Ledger commands E01-E09 on the absent document: exit 1 / 2 / 2 / 2 / 2 / 2 / 2 / 2 / 2 in ledger order, stderr as recorded (ugrep missing-file warnings; awk `can't open file`).
- E03/E05/E07 on the scratchpad fixtures: `f_pass.md` exit 0 for all three; E03 `m_e03_only1` `COUNT BP-2=0 COUNT BP-5=0 COUNT BP-6=0` exit 1; `m_e03_flag` `BAD-FLAG` exit 1; E05 `m_e05_nosec` `NO-PATH-SECTION` exit 1; `m_e05_hdr` `HEADER-COUNT=0` exit 1; E07 four mutants (above) exit 1.
- Traceability verb: `COLLECTED: 11 REQ definitions (acceptance input: read)`. CN-4 verb: `COLLECTED: 4 milestones in plan order (M1 M2 M3 M4), 12 exit bindings, 8 ordering candidates`. D8 verb: no output. D7: `grep -Eo` -> only the SPEC's own id. MP-7 grep exit 1.
- `mcp__moai__audit_multi` (project_root=<tree>, target=baseBranch, focus="acceptance command scope guard", no gates, no claude_verdict): `overall_verdict: fail`, codex `fail`, `audit_receipt: rcpt-25166887c6e4413d16283e54`, `plan_source: config`, `participant_count: 1`, one P2 finding (acceptance.md:199, ignored out-of-scope files pass E10a/E10b). Digest written to `.moai/state/audit-plan-result.json`, then `moai verify audit-plan --project-root <tree> --result-file <that path>` -> `convergence_check: {"ok": true, "unmet": []}`.
- Independent confirmation of the codex finding: `git check-ignore -v .moai/reports/t9999/new.md` -> `.gitignore:235:.moai/reports/*	.moai/reports/t9999/new.md`, exit 0.

### Baseline-attribution
Every figure above was measured in this run against HEAD a45ba6eb8424546f8d54db6ca10efc58f71d3412 (SPEC content as of commit 1e7588b6f; a45ba6eb8 adds only report files). The installed `moai` binary and the running moai MCP server are build `ee981a932` (MCP server banner: `v3.2.0-rc.36-t1822c (commit: ee981a932, built: 2026-10-15T21:56:00Z)`); I did not compare it with the tree's HEAD (see G4). The throwaway repos' base commits are listed beside each probe.

### Gaps
G1. Tool refusal 1: my first compound command that created the probe repo in the scratchpad (`rm -rf` + `mkdir` + `git init` + `git add -A` + `git commit` in one call) was denied by the permission system. I did not retry it; I rebuilt the same fixture as separate plain steps (mkdir, `git init`, files via the Write tool, `git add` with explicit pathspecs, `git commit`). The measurements are unaffected, but the substitution is named here.
G2. Tool refusal 2 (cleanup): `rm -rf` of the probe repo `/private/tmp/claude-501/-Users-goos-MoAI-moai-adk-go/39eb45b2-9233-459e-a4d1-94f379f1058c/scratchpad/probe4r4` (one compound call that also hashed the SPEC files) was denied. I did not work around it. The probe repo (three tracked files, two commits, a deleted-file state) remains in the session scratchpad, outside the tree; the scratchpad also still holds earlier-round fixtures (`f_pass.md`, `m_*.md`, `e07_*.md`, `e08_*.txt`, `gitprobe/`) that predate this audit and which I used read-only.
G3. Not checked: claude and glm audit backends (both gate `off` in the audit plan), so the convergence has one participant (`participant_count: 1`, `disagreement_flag: null`).
G4. Tool provenance (verification-claim-integrity.md §2.2): the installed `moai` build's commit `ee981a932` was not compared with the tree's HEAD ancestry. The measured behaviours are git/grep/awk, not moai logic, except `moai verify audit-plan` and the MCP tools.
G5. Not re-observed: the carried-over fix-round-1/2 probes (E10 untracked under `.moai/docs`, tracked `.moai/docs/agent-lint.md` appended) that the ledger itself marks as not re-run; the full 46/51-line fixtures of round 3 (rebuilt smaller here); `moai spec ceiling` / the Tier M iteration ceiling (the leader reads the ceiling policy).
G6. `plan_artifact_hash` is an auditor-computed sha256 of the four SPEC markdown files in `cat` order, not the runtime `ComputeHash`; admission may compute a different value.
G7. Side effect: I wrote `.moai/state/audit-plan-result.json` (required by the plan check). It is not shown by `git status --porcelain` (ignored).

### Residual-risk
R1. N5 is a wording defect only: ignored files under `.moai/reports/` are machine-local and not repository content, so the practical scope-creep risk is low; the sentence still overstates a check and a requirement-adjacent criterion.
R2. E08 and E10a/E10b are checked only on the shapes enumerated; unlisted date spellings (Known limit) remain reviewer territory.
R3. The run-phase landing sequence (merge of `main` into the card branch) may change the E10a diff against the fixed base (D3).
R4. A single required backend (codex) produced the convergence; its P2 was reproduced by me, but a second independent model did not run.

## Operational Notes (unverified)

- Measure whether the lane's landing merges `main` into the card branch before E10a runs at the run tip — command: `git merge-base --is-ancestor main WT-<slug>` on a finished lane tree. Status: assumption.
- Measure the Tier M plan-audit iteration ceiling before another fix round — command: `moai spec ceiling SPEC-CONTAINMENT-MAPPING-DOC-001`. Status: assumption (not run here).

AUDIT-VERDICT: FAIL spec=SPEC-CONTAINMENT-MAPPING-DOC-001 receipts=rcpt-25166887c6e4413d16283e54
