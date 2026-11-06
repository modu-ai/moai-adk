auditor-model: claude-sonnet-5-5

verdict: FAIL
audited_sha: 4a9ef2674462dbd089d9ae8b19ed88f349585b9b

# t1854 sync audit (Class B, no SPEC)

Overall: FAIL. Driven by the required codex leg (`overall_verdict: fail`, `residual_risk_note: required-backend FAIL: codex`) on one P3 documentation-accuracy finding (F1). The code fix itself is sound; the defect is in the three template notes.

## Dimension scores

| Dimension | Score | Verdict | Evidence |
|---|---|---|---|
| Functionality (40%) | 90/100 | PASS | Targeted run: `ok github.com/modu-ai/moai-adk/internal/homestate 19.113s`. Parser read through (below): `- `/`* ` accepted, conflicts still refused, look-alike hint cannot convert a valid signal to a refusal. |
| Security (25%) | 90/100 | PASS | Read-only string parse of a bounded file (`readBoundedFile`); no new I/O, exec, or path handling. Hint echoes a key from a repo-local progress.md into an error string via `%q` (quoted). |
| Craft (20%) | 80/100 | PASS | `gofmt -l internal/homestate` printed nothing. RED-first commit order confirmed (74f2864c8 precedes 343a33746). Full-package runs are cited from verdict.md, not re-run (slot held). |
| Consistency (15%) | 60/100 | FAIL | Template notes say "the kickoff gate reads only that form", contradicting the fixed parser (F1). |

## Findings

- F1 [P3] [blocking: the required codex gate returned fail on it] internal/template/templates/.claude/agents/moai/manager-spec.md:138 (same sentence in internal/template/templates/.codex/agents/moai/manager-spec.toml:125 and, as "reads a bare `audit_ready: true` line (no bullet marker)", in internal/template/templates/.claude/rules/moai/development/spec-frontmatter-schema.md:116) - The notes state the gate "reads only that form" / implies bullet markers are unsupported. After 343a33746 the parser accepts `- ` and `* ` bullets. Required fix: reword all three notes to "write a bare `audit_ready: true` line (recommended); one leading `- ` or `* ` list marker is also accepted; a look-alike key such as `audit-ready` is ignored and named in the refusal", then run `make build` to regenerate the `manager-spec` hash in internal/template/catalog.yaml and re-run the template package tests.
- F2 [optional] internal/homestate/card_audit_kickoff.go:68-72 - Only the literal `- ` / `* ` prefixes are stripped. `-\taudit_ready: true`, `-audit_ready: true`, and `- - audit_ready: true` are ignored, and the first two get no look-alike hint, because the normalised key starts with `_` and fails the `audit_ready` prefix test. Matches the stated intent (one `- `/`* ` marker); a hint for these is a nicety only.
- F3 [optional] internal/homestate/card_audit_kickoff.go:84-88 - Values `True`, `TRUE`, `true # note` are `conflicting` (refused) with no hint. Same as pre-fix behaviour; unchanged and consistent with "keep refusing conflicting values".
- F4 [optional, behaviour change to note] A bulleted `- audit_ready: <non-true>` line (for example a placeholder or an illustrative example) in §E.1 now counts as conflicting where it was previously ignored. `grep` of `.moai/specs/**/progress.md` for bulleted `audit_ready` found one line, `.moai/specs/SPEC-UPDATE-MIGRATION-FIX-001/progress.md:441: - audit_ready: true` (value true, so newly accepted, not newly refused). No existing bulleted non-true value found.

## Parser break attempts (read-through of the diff, plus codex overlay)

| Input | Result by code reading |
|---|---|
| `-  audit_ready: true` (extra spaces) | accepted (key is TrimSpace'd after prefix strip) |
| `  - audit_ready: true` (nested/indented) | accepted (line trimmed first) |
| `audit_ready:true` | accepted (value trimmed) |
| `-audit_ready: true`, `-<TAB>audit_ready: true` | ignored (F2) |
| look-alike line BEFORE a valid line | accepted; `ignored` is dropped when `ready && !conflicting` |
| fenced yaml with bullet | accepted (fence lines have no colon-key match; covered by test case) |
| `True` / `TRUE` / `true # note` | refused as conflicting (pre-existing, F3) |
| bullet true + bare false | refused (covered by test case) |

I did not run these inputs; they come from reading the diff. The codex leg reports running 18 additional overlay cases at HEAD `4a9ef2674` that passed (conflict rejection, and acceptance when a valid signal coexists with a look-alike).

## Evidence (five-section format)

**Claim.** The bullet-form fix and refusal hint behave as intended; the template notes are inconsistent with it.

**Evidence.**
- `go test ./internal/homestate -run 'TestAuditReadyBulletForms|TestFDA_AuditDecider' -count=1` -> `ok  	github.com/modu-ai/moai-adk/internal/homestate	19.113s`
- `gofmt -l internal/homestate` -> (no output)
- `moai verify audit-plan --project-root <toplevel>` -> `"cross_model_active": true`, `"cross_model_required": true`, `"enforced_required": ["codex"]`, claude and glm `off`.
- `mcp__moai__audit_multi` (project_root = this worktree, target baseBranch), digest members: `overall_verdict: fail`; `gate_unmet`: absent from the result (the result carries `residual_risk_note: required-backend FAIL: codex`); `plan_source: config`; per backend: `backend: codex, gate: required, verdict: fail`. `audit_receipt: rcpt-1cee52f249f7667fb5f08dcb`.
- Codex leg verbatim (bounded): "Verdict: fail - [P3] ... manager-spec.md:138 ... `the kickoff gate reads only that form`은 실제 동작과 다르다. 현재 파서는 `- audit_ready: true`와 `* audit_ready: true`도 승인하며, `TestAuditReadyBulletForms`가 통과했다. 같은 설명이 Codex 템플릿과 `spec-frontmatter-schema.md`에도 있다." (Translation: the statement is wrong; the parser also approves `- ` and `* ` forms; the same text is in the Codex template and spec-frontmatter-schema.md.)

**Baseline-attribution.** Tree `.moai/worktrees/t1854`, HEAD `4a9ef2674462dbd089d9ae8b19ed88f349585b9b`, base `fb749d372`. Tests built from this tree by `go test`. The installed `moai` binary (commit ee981a932, a strict ancestor of HEAD per the tool's `build_lag` note) was used for `verify audit-plan` and the MCP server only; it did not judge the Go tests. The codex result was produced by a server running that older build, so the review text describes this tree only through the diff it was given.

**Gaps.**
- The full `internal/homestate` and `internal/template` package runs were not re-run (slot held by the leader's turn order); I cite verdict.md's carried-over lines, which this audit did not observe.
- `go vet` and `golangci-lint` not run.
- The edge-case table above is from code reading, not executed by me.
- `gate_unmet` is absent as a member of the audit_multi result rather than present and empty.
- Convergence result file for the orchestrator's `--result-file` check: not written by me (read-only auditor); the orchestrator writes the digest members above to `<toplevel>/.moai/state/audit-plan-result.json` and runs the check. This verdict is not final until that check passes.

**Residual-risk.** Template text drift (F1) will mislead manager-spec output. The F4 behaviour change could refuse a card whose §E.1 carries a bulleted non-`true` example, though none was found in the repo's SPECs.

## Binding run conditions

No SPEC and no `progress.md` for this card; no binding conditions to dispose.

## Iteration history

Iteration 1: FAIL (required codex gate).

## Recommendations

Apply the F1 rewording to the three notes, `make build`, re-run the template package tests in the slot, and request a delta re-audit scoped to F1.
