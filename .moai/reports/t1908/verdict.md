# t1908 Verdict — remove MoAI MCP cross-session messaging tools

Card: t1908 (Class C·P1) · Tree: `.moai/worktrees/t1908` · Branch `WT-mcp-codex-claude`

## Claim

1. The nine `session_msg_*` / `factory_msg_*` MCP tools are removed entirely (handlers, registrations, catalogue entries, sole-covering tests), on the post-merge tree.
2. `go build ./...`, `go vet ./internal/cli ./internal/mcp`, and `golangci-lint run ./internal/cli ./internal/mcp` exit clean on the merged tree.
3. A real MCP `tools/list` against the tree-local binary lists 39 tools; the nine removed names are ABSENT; keep-set samples (`codex_audit`, `audit_cache`, `spec_audit`) are PRESENT.
4. Remaining mcp_server/catalog tests pass (76 name-pinned RUNs in `internal/cli` + the full `internal/mcp` package).
5. Removal commit `8a368823b7`; base-refresh merge commit `83a2ffbc1b` (batch head `85c62aaff2`).

## Evidence

- **Removal commit**: `8a368823b7` — `refactor(mcp): remove MoAI MCP cross-session messaging tools` (13 paths, pathspec staging, `-F` message file).
- **Merge commit**: `83a2ffbc1b` — `Merge 85c62aaff2 into WT-mcp-codex-claude (card t1908, base refresh)`; conflict-free (30 files, exit 0; `MERGE_HEAD` empty after, tree clean).
- **Removed-file inventory**: `internal/cli/mcp_session_msg.go`, `internal/cli/mcp_session_msg_test.go`, `internal/cli/mcp_session_msg_boundary_test.go`, `internal/cli/mcp_session_msg_data_test.go`, `internal/cli/mcp_factory_msg.go`, `internal/cli/mcp_factory_msg_test.go` (git rm). Edited: `internal/cli/mcp_server.go` (registration block + `sessionmsg`/`factorymsg` imports), `internal/mcp/catalog.go` (nine entries), `internal/mcp/catalog_test.go` (48→39, write-capable 21→15), `internal/cli/mcp_jev_catalog_doc_test.go` (family arithmetic), `.claude/rules/moai/core/moai-mcp-tools-catalogue.md` + template mirror (48→39 figures, both messaging sections and family rows dropped; mirror `cmp`-identical). `closeFactoryToolStore` relocated verbatim into `internal/cli/managed_factory_session.go` (callers at `managed_factory_session.go:522`, `managed_codex_factory.go:911`).
- **tools/list measurement** (both pre- and post-merge): `go build -o /tmp/t1908-moai{,-merged} ./cmd/moai` → probe script `/tmp/t1908-probe/toolslist.py` spawns the binary's `mcp-server`, drives initialize → notifications/initialized → tools/list over stdio, writes the captured names to `/tmp/t1908-probe/tools-post-merge.txt`. Post-merge: `COUNT=39`; per-name assertions from the captured file: `session_msg_register|session_msg_list|session_msg_send|session_msg_poll|factory_msg_send|factory_msg_list|factory_msg_body|factory_msg_receipt|factory_msg_status` → ABSENT=YES (grep -cx = 0 each); `codex_audit` PRESENT=1, `audit_cache` PRESENT=1, `spec_audit` PRESENT=1. Server cleanup: stdin EOF + `wait(10)`, `kill` on the recorded Popen handle in `finally` (no pattern kill).
- **Build/vet/lint** (merged tree, this run): `go build ./...` → exit 0; `go vet ./internal/cli ./internal/mcp` → exit 0; `golangci-lint run ./internal/cli ./internal/mcp` → `0 issues.` exit 0.
- **Tests** (merged tree, this run): `go test ./internal/mcp -count=1` → `ok ... 0.134s`; `go test ./internal/cli -run 'TestMoaiMCP|TestMCPToolCatalogue|TestMCPAnnotation|TestMCPConsole|TestMoaiMCPServer|TestMCPAudit|TestMCPSelfReview|TestCodexReview' -count=1 -v` → 76 `=== RUN`, `ok ... 8.703s` (verbatim at `/tmp/t1908-probe/tests-post-merge.txt`).

## Baseline-attribution

- All figures above were produced in this run against the merged tree at HEAD `83a2ffbc1b` (branch `WT-mcp-codex-claude`). Judging build: tree-local binaries `/tmp/t1908-moai` (pre-merge, removal commit `8a368823b7`) and `/tmp/t1908-moai-merged` (merged `83a2ffbc1b`), both built from the tree under measurement — no installed-binary lag.
- The pre-merge tools/list (COUNT=39, `/tmp/t1908-probe/tools-pre-merge.txt`) and post-merge list are identical; the base refresh reintroduced none of the nine.
- Guard refusal recorded: the first `git merge` attempt (compound `cd <worktree> && git merge`) was denied by the BranchGuard (`BRANCH_GUARD_VIOLATION: git merge in primary checkout`) because this session's process cwd resets to the primary checkout between Bash calls. Not retried in that shape; the merge ran once as `git -C <card-worktree> merge …`, the remedy the deny names ("use a worktree"), and succeeded — the deny was a cwd-classification artifact, not a blocked operation.

## Gaps

- No full `go test ./...` on the merged tree (lane-local discipline: scope verification to the change; `internal/cli` + `internal/mcp` cover the touched packages, plus `go build ./...` over the whole module).
- Cross-references to the removed tools remain in prose surfaces outside the card scope and are left for sync-phase: `.claude/rules/moai/workflow/cross-session-messaging-detail.md` (+ mirror), `.claude/skills/moai-lane-watchdog/SKILL.md` (+ mirror), `internal/hook/factory_messages.go` / `factory_rebind.go` / `session_start_factory_i18n.go` (hook prompt text), `managed_codex_approval_test.go` / `managed_factory_session_test.go` / `factory_live_test.go` / `factory_operational_*_test.go` (compile-safe string references), and the Codex approval-override generator still emitting `.tools.factory_msg_send/receipt` scoping for tools that no longer exist (dead config, functionally harmless).
- The primary checkout's copy of the catalogue doc (47-tool generation) was not touched; only the card worktree copies were edited.

## Residual-risk

- The batch may land further commits after `85c62aaff2`; a later lane absorbing a newer batch head must re-run the tools/list + doc-figure checks (the merge that introduces `factory_renew`-style additions shifts the 39 figure).
- Sessions with a long-lived moai MCP server (started before this change) keep listing the nine tools until the server respawns — runtime restart, not a tree defect.
- `go test ./...` on the merged tree could surface failures in untouched packages interacting with the removed surface indirectly (e.g., discovery/live tests referencing the tool names in prompts); not observed, bounded by the targeted runs above.

---

# Chain-change addendum — sync-phase leftovers (leader directive, same card/PR)

## Claim

The operational references to the removed tools in the doctrine/skill/hook surfaces are cleaned; the dead Codex approval overrides are removed; the boundary-ambiguous items are listed for the leader, not edited; all gates re-pass on the new HEAD.

## Evidence (this run, tree = HEAD `a124c984ce` + the uncommitted cleanup, all verified before commit)

- **① prose**: `.claude/rules/moai/workflow/cross-session-messaging-detail.md` (+ sanitized template mirror — each copy edited in place, fork conventions preserved): the Codex-broker-path section rewritten — broker removed in t1908, coordination rides the queue-on-disk channel, the broker-reading table dropped. `.claude/skills/moai-lane-watchdog/SKILL.md` (+ byte-identical template mirror, `cmp`): §4 Codex step ⑤ reworded to the queue-on-disk channel. Both references were OPERATIONAL (they instructed tool calls), not prose-only.
- **② hook prompt text**: `internal/hook/factory_rebind.go` (rebound-notice tool tail dropped), `internal/hook/factory_messages.go` (inbox prefix reworded — notification-only, tool instructions dropped), `internal/hook/session_start_factory_i18n.go` (the four-locale `operationalStatus` strings reworded to keep one `%q` + the liveness caveat; field comment updated). No hook test pinned the dropped text; the surviving pins (`Factory inbox run=`, `factory lane rebound:` substrings) still pass.
- **③ approval-override generator**: `internal/cli/managed_codex_factory.go` `factoryMoAIMCPApprovalArgs()` now emits only the `default_tools_approval_mode="approve"` pair — the two dead `.tools.factory_msg_send/receipt` overrides removed; `internal/cli/managed_codex_approval_test.go` updated (asserts the default override and rejects any `.tools.factory_msg` naming).
- **Re-measurement** (tree-local binary `/tmp/t1908-moai-chain`, built at this working tree): `COUNT=39`; the nine removed names absent (grep -cx = 0 each); `codex_audit`/`audit_cache`/`spec_audit` present (=1). `go build ./...` exit 0; `go vet ./internal/cli ./internal/mcp ./internal/hook` exit 0; `golangci-lint run ./internal/cli ./internal/mcp ./internal/hook` → `0 issues.`; `go test ./internal/hook -count=1` ok (250s); name-pinned set (`TestMoaiMCP|TestMCPToolCatalogue|TestMCPAnnotation|TestMCPConsole|TestMoaiMCPServer|TestMCPAudit|TestMCPSelfReview|TestCodexReview|TestMoAIMCPApproval|TestManagedCodex|TestConfigToml`) → 212 RUNs, `internal/mcp` + `internal/cli` both ok.

## Gaps (listed for the leader, deliberately not edited)

- `cross-session-messaging-core.md` (always-loaded stub) + `cross-session-messaging.md` (role-injected full body), both copies: the § "Codex broker path (session messaging tools)" section still describes the removed broker as present. Names no removed tool identifier; ownership ambiguous vs the t1909 chain — not edited.
- `internal/cli/managed_factory_session.go` `managedFactoryInboxPrompt` (line ~458, Korean): instructs the managed codex session to read/settle via `factory_msg_body`/`factory_msg_receipt`. Same defect class, but the managed inbox flow's design depends on the removed tools (a redesign decision), and it sits in the t1909-adjacent codex-managed surface — not edited. Its test (`managed_factory_session_test.go`) pins the tool names.
- Prose/history keeps: `internal/sessionmsg/*.go` @MX fan_in annotations, `internal/config/defaults.go` SessionMsg constant comments (constants stay live for the retained package), `factory_live_test.go` / `factory_operational_*_test.go` / `managed_hardening_test.go` (compile-safe prompt/evidence strings), `mcp_code_tools_test.go:173` comment.

## Residual-risk

- The four-locale `operationalStatus` strings changed shape (tool call → run-id caveat); any downstream consumer matching the old leading tool-call syntax verbatim would miss — none found in the test sweep.
- The managed codex inbox flow remains functionally broken-by-design (claims render, model cannot settle) until the leader rules on the listed item.

