auditor-model: claude-sonnet-5-5

# Sync audit iter2 - card t1854 (delta re-audit of F1)

verdict: PASS
audited_sha: ba38ecf07b175d7aefa5b8f2dc0f6393915b6f5f

Overall: PASS. Iter-1 F1 is resolved. No new findings. Open Gap: the template package test run (lane's slot turn).

## Dimension scores (flat-weighted)

| Dimension | Score | Verdict | Basis |
|---|---|---|---|
| Functionality (40%) | 90 (carried from iter1) | PASS | Delta touches docs and a catalog hash only. Targeted tests re-run green (below). |
| Security (25%) | 90 (carried from iter1) | PASS | No code change in the delta; no input surface touched. |
| Craft (20%) | 80 (carried from iter1) | PASS | No code or test change; coverage not re-measured, so the iter-1 figure stands as a carry-over (Gap). |
| Consistency (15%) | 60 -> 88 (re-scored) | PASS | The three notes now match the parser and each other. |

Must-pass firewall (Functionality + Security): both pass. Consistency was the only failing dimension in iter1.

## Findings

- F1 (iter1, P3, blocking): RESOLVED. The three notes no longer claim "reads only". Notes now state the bare form is canonical and one leading `- ` or `* ` is tolerated.
- No new findings. Optional observation (not blocking): the notes say "names it in the refusal". The parser returns the look-alike key only when the readiness check fails; a section with a valid `audit_ready: true` plus a look-alike passes silently. The sentence is true as a description of the refusal path.

## Claim / Evidence / Baseline-attribution / Gaps / Residual-risk

### Claim 1: every sentence in the three notes is true against the parser now.
Evidence (parser read at HEAD, `internal/homestate/card_audit_kickoff.go` lines 42-94):
- Accepts one leading `- ` or `* ` (lines 67-71 CutPrefix, if/else-if so only one marker).
- Key must equal `audit_ready` exactly (line 77); any other key whose lower-cased, `-`/space-to-`_` normalised form starts with `audit_ready` is recorded as ignored (lines 78-81), e.g. `audit-ready`.
- Ignored key is returned only when `ready && !conflicting` is false (lines 90-93) and the caller appends `; ignored look-alike key %q, write a bare `audit_ready: true` line` to the refusal (lines 34-36).
Notes at HEAD (verbatim from `git show ba38ecf07`):
- manager-spec.md:138 and manager-spec.toml:125: "bare `audit_ready: true` line (the canonical form; the kickoff gate also tolerates one leading `- ` or `* ` bullet marker, but ignores any other spelling such as `audit-ready` and names it in the refusal)".
- spec-frontmatter-schema.md:116: "write a bare `audit_ready: true` line (the kickoff gate also tolerates one leading `- ` or `* ` bullet marker)".
Result: each clause maps to a parser line. True.

### Claim 2: the notes agree with each other and with the refusal text.
The .md and .toml agent notes are byte-identical sentences. The schema row is a strict subset (omits the look-alike clause), not a contradiction. The refusal text says "write a bare `audit_ready: true` line", the same canonical form. Agree.

### Claim 3: no other template or doc still carries the old phrasing.
Commands: `grep -rn "audit_ready" internal/template/templates .claude` returned exactly three hits (the three reworded notes; `.claude` has none). `grep -rnE "reads only|no bullet marker|no \`- \`"` filtered on audit/bullet over the same trees returned no lines. Repo-wide `grep -rIl audit_ready` outside `.moai/specs` and `.moai/reports` returned only CHANGELOG.md; line 54 describes the bullet acceptance correctly (it was written in the run phase and states acceptance, not exclusivity).
Result: none remain.

### Claim 4: catalog.yaml change is only the touched template hash, and 04c83ecbc added only the verdict file.
Evidence: `git diff 4a9ef2674..HEAD -- internal/template/catalog.yaml` shows one hunk, one line: manager-spec hash `fc837944...` to `332974b6...`. `shasum -a 256 internal/template/templates/.claude/agents/moai/manager-spec.md` printed `332974b6913e2a6e778f4b90882ba5896712274cb3d76d9b1f00e48488c9739f`, equal to the new catalog hash. `grep -c "manager-spec.toml\|spec-frontmatter-schema" internal/template/catalog.yaml` printed 0, so the other two files are uncatalogued and need no hash line. `git show --name-status 04c83ecbc` printed `A .moai/reports/t1854/sync-audit.md` only. `git diff --stat 4a9ef2674..HEAD` lists five files: the iter-1 verdict, catalog.yaml, and the three templates.

### Claim 5: targeted tests pass.
Command: `go test ./internal/homestate -run 'TestAuditReadyBulletForms|TestFDA_AuditDecider' -count=1`
Observed: `ok  	github.com/modu-ai/moai-adk/internal/homestate	13.775s`

### Cross-model leg (codex, gate required)
Tool: mcp__moai__audit_multi, project_root = this worktree, target baseBranch, single call.
Result members: overall_verdict=pass; gate_unmet absent; plan_source=config; per_backend_verdicts=[{backend: codex, gate: required, verdict: pass, summary: "Verdict: pass", findings: []}]; fail_open_backends=[]; disagreement_flag=null; participant_count=1.
audit_receipt: rcpt-2e33fe188e6c3c51894e8c0b
build_commit ee981a932 is an ancestor of HEAD ba38ecf07 (tool reported build lag); the codex leg reads the diff, not this binary's gate logic. Named as a provenance Gap below.
This verdict is not final until the orchestrator's convergence check passes (`moai verify audit-plan --result-file`); the auditor does not write the result digest.

### Baseline-attribution
All greps, git commands, and the test run were executed in this run against tree HEAD ba38ecf07b175d7aefa5b8f2dc0f6393915b6f5f. `moai verify audit-plan` (installed build ee981a932, strict ancestor of HEAD) resolved codex=required, claude=off, glm=off.

### Gaps
- The `internal/template` package test (embedded-template and catalog hash tests) was NOT run by this audit; the lane re-runs it in its own slot turn and cites it separately. The catalog hash equality was checked by `shasum` only. Open Gap, not a pass.
- The full `internal/homestate` package run was not executed (only the two targeted tests).
- Coverage and lint were not re-measured; Craft 80 is carried from iter1 because the delta has no code.
- The installed moai binary lags HEAD (ee981a932 vs ba38ecf07); `audit-plan` output reflects that older build's config reader.
- Consistency re-score 88 is a judgment on three prose sentences, supported by the greps above.

### Residual-risk
- A future parser change (for example accepting a numbered list marker) would again outdate the notes; no test pins the note wording to the parser.
- Deployed copies under `.claude/` in user projects update only through `moai update`; the repo-local `.claude` here carries no `audit_ready` text to reconcile.

## Binding run conditions
None found for this card (no SPEC, Class B); nothing to dispose.

## Iteration history
- iter1 (audited 4a9ef2674): FAIL, F1 P3 Consistency 60. Functionality 90, Security 90, Craft 80.
- iter2 (audited ba38ecf07): PASS. F1 resolved; Consistency 88; others carried.

## Recommendations
- Lane: run the `internal/template` package test in its slot turn and cite it beside this verdict.
