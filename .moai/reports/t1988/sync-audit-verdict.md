auditor-model: glm-5.3
verdict: PASS
audited_sha: 7a263a0580bb2283bd2dd6c1423c36c1733e18e0
rebind-note: audited_sha rebound from 8efe31c080f1d49b86898620ee83cef99fe52bbe to the structural gate head 7a263a0580bb2283bd2dd6c1423c36c1733e18e0 — the commit that tracks this verdict file (T11 requires the artifact at the recorded commit); the audit itself was performed and observed at 8efe31c080f1d49b86898620ee83cef99fe52bbe, whose tree differs from the gate head only by this evidence set (docs only, zero product files).

# t1988 sync-audit iter-2 — F1 repair confirmed (GH #1924/#1926, Class B)

Independent sync-auditor (depth-1 leaf), run tnbn7e, iteration 2. Iter-1 verdict (FAIL, F1) is preserved unchanged at `.moai/reports/t1988/sync-audit-verdict-iter1.md` (audited_sha `e3b7e7145010ea994ed90386404ff42703da1774`). This iteration is scoped to the enumerated F1 defect delta per the finding-consumption discipline: the only tree change between the two audited heads is commit `8efe31c080`, and the re-audit set is the F1 check plus spot re-runs of E1/E4 and the machine lines.

Audit plan at this iteration (re-read): `config_status: ok`, `cross_model_active: true`, gates — claude `off`, codex `required`, glm `off` (`config.gates`).

## Claim

| # | Claim | Result |
|---|---|---|
| C1 | (Iter-1 F1 required fix) The catalog pin was regenerated and committed with card id, exactly one line | VERIFIED (E7′) |
| C2 | `TestManifestHashFormat` is green at the new HEAD | VERIFIED (E8′) |
| C3 | E1 targeted refusal test still green at the new HEAD | VERIFIED (E1′) |
| C4 | E4 doc parity between the two `phase-execution.md` copies still holds | VERIFIED (E4′) |
| C5 | The delta-scoped required codex gate passes at the new HEAD | VERIFIED (E9′, receipt cited) |

Overall: **PASS** — iter-1 blocking finding F1 is disposed with observed evidence; no new blocking finding. Optional findings F2/F3 from iter-1 remain open as optional (non-blocking) items.

## Evidence (command + verbatim output, this run, this tree)

**E7′ — Delta scope and the one-line fix.**
Commands: `git rev-parse HEAD`; `git diff --name-only e3b7e7145010ea994ed90386404ff42703da1774..HEAD`
Output:
```
8efe31c080f1d49b86898620ee83cef99fe52bbe
internal/template/catalog.yaml
```
Working tree clean (`git status --short` empty). Commit `8efe31c080 fix(template): regenerate the moai skill whole-tree catalog pin after the phase-execution.md hash-definition fix (card t1988, sync-audit F1)` — card id present.
`git diff e3b7e71450..HEAD -- internal/template/catalog.yaml` output (the entire delta):
```
-              hash: 18df0ebde3a82e0718c904dedc9c3ac8970c2dd6cb4676d5b9e29be107898286
+              hash: 33ee619f96dcea5430dabdfce9db9b1ec912f62b80c76766b7660602e3898644
```
The new stored hash is byte-identical to the computed value iter-1 observed (`33ee619f96dcea5430dabdfce9db9b1ec912f62b80c76766b7660602e3898644`) — the pin now matches the deployed tree rather than masking the drift.

**E8′ — F1 check at the new HEAD.**
Command: `go test -count=1 ./internal/template -run TestManifestHashFormat`
Output:
```
ok  	github.com/modu-ai/moai-adk/internal/template	0.528s
```
exit=0. The iter-1 failure (`CATALOG_HASH_UNSTABLE: moai stored hash=18df0ebd…, computed hash=33ee619f…`) no longer reproduces.

**E1′ — Targeted refusal test (Functionality spot re-run).**
Command: `go test ./internal/cli/ -run TestFactoryUnknownAttributionErrorStatesManualStamp -count=1`
Output:
```
ok  	github.com/modu-ai/moai-adk/internal/cli	1.510s
```

**E4′ — Doc parity spot re-run.**
Command: `diff .claude/skills/moai/workflows/run/phase-execution.md internal/template/templates/.claude/skills/moai/workflows/run/phase-execution.md`
Output: (empty) `diff_exit=0` — the two copies remain byte-identical.

**E9′ — Delta-scoped cross-model convergence (required codex gate).**
Call: `mcp__moai__audit_multi` (project_root = this worktree, target = `baseBranch`, delta_base = `e3b7e7145010ea994ed90386404ff42703da1774`, focus = the regenerated catalog pin; no `gates` argument; GLM main session so no `claude_verdict`).
Digest members: `overall_verdict: pass`; `gate_unmet`: absent (empty); `plan_source: config`; per-backend: `codex / required / pass` (sole participant; claude and glm `off` per plan); `disagreement_flag: null`; `participant_count: 1`; `fail_open_backends: []`; `audit_receipt: rcpt-303010c513ae169fe23c8633`; server `build_commit: b1806498fe`. No findings returned.
Iter-1 receipt (same session, superseded scope): `rcpt-bafdd64fe6a0f83245d137d0`.
This auditor is read-only: the orchestrator writes the digest members fresh to `<toplevel>/.moai/state/audit-plan-result.json` and runs `moai verify audit-plan --project-root <toplevel> --result-file <path>`; this PASS is not final until that check passes.

**Iter-1 preserved verdict integrity.** An anchored count of the iter-1 machine verdict line in `.moai/reports/t1988/sync-audit-verdict-iter1.md` returns exactly `1` (grep, this run) — the preserved FAIL file carries exactly one machine verdict line, unmodified.

## Baseline-attribution

- Tree: `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1988` (git worktree), branch `WT-triage-github-issues`, audited HEAD `8efe31c080f1d49b86898620ee83cef99fe52bbe`, working tree clean at audit open. Parent audited head (iter-1): `e3b7e7145010ea994ed90386404ff42703da1774`.
- Every output above was produced in this run, against this tree, with the local Go toolchain compiling from tree source.
- Iter-1 checks E2 (old-wording falsifiability), E3 (writer enumeration), E5 (set-spec source existence), E6 (scope + gofmt) are carried forward, not re-executed: the commit graph shows the only change between the two audited heads is `internal/template/catalog.yaml` (E7′ name-only output), which none of those checks read. The carried evidence is attributed to iter-1's run at `e3b7e71450` and is valid here by the graph, not by re-observation.
- The MCP convergence server remains the installed build `b1806498fe` (not an ancestor of the audited HEAD, unchanged from iter-1); its pass verdict was corroborated by this auditor's own local run of the same test (E8′).

## Gaps (explicitly unobserved)

- Full `go test ./...` still not run (lane-local doctrine); `internal/cli` full suite and platform matrix unrun — CI carries them.
- E2/E3/E5/E6 not re-executed at `8efe31c080` — carried from iter-1 via the commit-graph delta (one file, none of those checks' subjects); see Baseline-attribution.
- The delta convergence scoped codex in prose only and wrote no session state or second-review record (tool-reported; see Residual-risk verbatim note); no claude or glm leg exists for the delta (both gates `off` per plan).
- plan-audit→kickoff end-to-end reproduction still not run (live factory-record mutation is dispatch-prohibited); mechanism verified at source/gate level only, unchanged from iter-1.

## Residual-risk

- Convergence ran with a single participant (codex `required`; claude/glm `off`); `disagreement_flag: null` — no agreement claim is made. Tool-reported note, verbatim: `"delta_base e3b7e7145010ea994ed90386404ff42703da1774: no claude or glm leg produced a verdict on it; codex was scoped in prose only (not mechanically bounded); no second-review record and no session state written"`.
- Any future drift under `templates/.claude/skills/moai/` re-reds `TestManifestHashFormat`; the regeneration (`go run ./internal/template/scripts/gen-catalog-hashes.go --all`) must be the last step of any template edit — F2's process recommendation (add the `internal/template` hash test to the lane-local required set for template-scope diffs) remains the standing mitigation.
- F3 (optional, open): the kickoff-side refusal still does not point auditors at the runtime hash definition, so an arbitrary-pipeline `plan_artifact_hash` can still wedge plan→kickoff despite the corrected docs — follow-up candidate outside this card.
- The local `.claude/skills` twin still has no integrity pin equivalent to the template catalog; single-sided edits of the pair would be caught only by a manual `diff`.
