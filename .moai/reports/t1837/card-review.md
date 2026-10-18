# Card t1837 — card-review (advisory)

Backend: codex via `mcp__moai__codex_review`, `scope: card`, `project_root` = `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1837`.
Base: `fb749d372e61a6b10c53091362c76b4412827427` (merge base, recomputed per call). Tree: the card worktree above.

## Round 1 — reviewed HEAD dd4d40427 — verdict: fail

Verbatim finding (translated header kept as returned):

- [P2] Workflow 변경 검사를 Deliver 단계 앞으로 옮겨줘 — `internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-mechanics.md:155`
  The procedure says its order is fixed, but the operator check sits after step 3 Deliver and step 4 Close, so a `.github/workflows/**` change is checked only after the PR is opened and auto-merge requested; the pre-approval gate applies too late.

Disposition: FIXED. Step `5b` removed and reinserted as `2b` between Card-review and Deliver in both mirror copies; the cards section reworded to "stops before Deliver". Commit `1969709ce`. The reviewer was right: the gate was placed after the action it guards.

## Round 2 — reviewed HEAD 1969709ce — verdict: pass

Verbatim: "Verdict: pass ... 수정이 필요한 명확한 결함은 발견하지 못했어." Findings: none. The reviewer ran no tests or builds (static review only).

Re-review count used: 1 of 2 (ceiling not reached).

## Gaps

- The review is static; it observed no build or test output.
