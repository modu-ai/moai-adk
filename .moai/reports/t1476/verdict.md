# t1476 verdict — AMI-009 wording tension

## Claim
1. The tension was alive at develop `2b9e4a4d0`: REQ-AMI-009 / AC-AMI-009 (v0.6.0) required `.moai/logs/agent-model-audit.jsonl` to stay unwritten, while the shipped observer writes one row per Agent spawn.
2. Resolved by an in-place amendment (spec.md v0.9.0, `completed → in-progress`, `prior_completed_sha: ad0454dec`) that aligns the text with the shipped, operator-approved behaviour (opt-in `llm.agent_overrides_consume`, observe-never-blocks row with `override_consumption`). No production code changed; guard test added.

## Evidence
- Writer: `internal/hook/agent_model_guard.go` `appendAuditJSONL` → `os.OpenFile(... agent-model-audit.jsonl, O_APPEND|O_CREATE ...)`, called unconditionally from `checkAgentModel`.
- Pre-dates the "unwritten" text: `git show 336cd54d8:internal/hook/agent_model_guard.go | grep -n "appendAgentModelAudit(h"` → `176:	appendAgentModelAudit(h.projectRoot(), agentModelAuditRecord{` (336cd54d8 is the v0.6.0 amendment commit).
- Positive control / guard test: `go test -count=1 -run TestAgentModelAuditRowPerSpawn -v ./internal/hook/` → `--- PASS: TestAgentModelAuditRowPerSpawn (0.00s)` — two Handle() spawns produce exactly 2 rows (`declared`, `inherit`, `override_consumption: off`), no deny.
- `go vet ./internal/hook/` clean; `go test -count=1 -run AgentModel ./internal/hook/` → `ok`.
- `moai spec lint --strict SPEC-AGENT-MODEL-INHERIT-001` → `0 error(s), 2 warning(s)`, exit 0. Both warnings (acceptance.md:37, :38 — AC-AMI-023/024 `-run` anchoring) are on lines this card did not touch (`git diff` touched acceptance.md line 23 only).
- Commit hook: `ac-baseline-guard: ... no count moved` (25 REQ / 25 AC).

## Baseline-attribution
All commands run in this run against worktree `.claude/worktrees/agent-a562e1ed254e73ddd`, base = develop `2b9e4a4d0` (HEAD == develop at start, no absorb needed). `moai` binary used for lint is the installed build, not rebuilt from this tree (§2.2 Gap).

## Gaps
- Lint is 0 errors but not 0 findings: 2 pre-existing WARNINGs left untouched (out of scope; the suggested `^(...)$` anchors can empty the selection — needs a `-list` check before changing).
- Installed `moai` build provenance vs tree HEAD not compared.
- Codex's isolated prune repro was not re-run; the finding was confirmed by reading `prune_logs.go:171`.
- No plan-audit / sync re-close: SPEC is now `in-progress` under the amendment and needs a sync re-close (manager-docs) to return to `completed`.

## Residual-risk
- The audit file grows without bound on a continuously used project (mtime-keyed whole-file age-out). Now recorded as an open residual in design §G; a follow-up card for row-level/size-based pruning is recommended.
