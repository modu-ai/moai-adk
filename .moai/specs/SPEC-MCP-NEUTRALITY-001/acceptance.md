---
id: SPEC-MCP-NEUTRALITY-001
title: "Acceptance criteria — MCP refusal structuring and harness neutrality"
version: "0.2.0"
created: 2026-10-14
updated: 2026-10-23
author: manager-spec
---

# SPEC-MCP-NEUTRALITY-001 — Acceptance Criteria

## §A Adoption contract (two classes, stated per row)

Per `.claude/rules/moai/development/verification-completeness.md` §2, every AC carries a RED-now cell and a green path cell as a pair; the carrier is the evidence ledger in §C (table rows cite ledger ids, since table cells mangle shell metacharacters). Two classes are used, and every row states which:

- **Release-blocking (3 rows)** — the card's terminal gates (AC-MN-001/002/003, per §F). Each carries a RED-now cell with all four §2.1 elements — a single-invocation command, its verbatim stdout, its exit code, and the tree SHA — measured on this tree at plan phase, plus the green path cell naming the Milestone that flips it.
- **Milestone-gating (22 rows)** — classified regression-guard at plan phase under the §2.1 undecidable disposition: the defect each row forbids is observable, at plan phase, only through an instrument that does not exist yet (the TDD RED test the owning Milestone writes first). A missing test is absent execution, not a red (§1.3), so no plan-phase command can observe these rows red for their stated reason; each ledger entry cites the defect's in-tree or audit anchor instead, states the disposition, and is not recorded as a pass. Their RED cells become measurable at their Milestone's RED step; the SPEC's DoD still requires every row green at close.

Verification commands are single-invocation, read-only forms run from the worktree root. Test names inside green-path cells are the run-phase worker's to fix precisely; the names indicate the family.

## §B AC Matrix

| AC | Given / When / Then | REQ | EV | Flipped by |
|----|---------------------|-----|----|------------|
| AC-MN-001 | Given a fixture holding a live serial-slot lease owned by another endpoint, When a factory verb is called through the MCP tool path, Then the response has `is_error` false, structured `refused: true` with `code: serial-slot`, and the text block carries the same refusal JSON. | REQ-MN-001 — **card gate 1: 정책 거부 is_error 0** | EV-001 | M1 |
| AC-MN-002 | Given the failure observer with a scratch inbox file, When the suite forces N policy refusals and exactly 1 genuine transport error, Then the inbox gains exactly 1 new mcp row and 0 refusal rows. | REQ-MN-002 — **card gate 2: lessons-inbox mcp 신규 0** | EV-002 | M1 |
| AC-MN-003 | Given the gate-session measurement harness of plan §F M3, When the measurement runs against the gate fixtures, Then it prints the metric line `<successes>/<total> (<pct>%)` with the sample count; a run that prints no metric line is a FAIL. Metric: a gate call is successful when the moai MCP server returns a non-error, non-refused result to the gate-session-shaped client; the denominator is the fixture call set — initialize, tools/list, and one call per tool in the REQ-MN-009 read-only subset, issued by the M3 harness; the numerator is the successful count over that set; the instrument is the M3 measurement test in `internal/cli`. | REQ-MN-009 + gate path — **card gate 3: Codex 게이트 호출 성공률 측정** | EV-003 | M3 |
| AC-MN-004 | Given a spawn recorded as lane-A, an active run R1 and a completed run R0, When the tool is called with `actor: lane-B` on R1 and with `actor: lane-A` on R0, Then both refuse (`refused: true`, codes `owned` / `stale-run`, `is_error` false) and a lane-A call on R1 proceeds. | REQ-MN-003 | EV-004 | M1 |
| AC-MN-005 | Given a session with no lane env and no operator mark, When `factory_decide` is called, Then it refuses with unknown attribution; with the positive operator mark it proceeds. | REQ-MN-003 | EV-005 | M1 |
| AC-MN-006 | Given the server on a registered worktree, When a root-less family tool receives `project_root: 42`, Then an invalid-argument error results (no silent coercion); When a mutating family tool omits `project_root`, Then the call refuses at the tool layer with the wire result `{"isError": true, "content": [{"type": "text", "text": "<tool>: project_root is required for mutating calls"}]}` — argument validation, outside the five policy codes; no JSON-RPC `-32602` protocol error is produced for this case; When called correctly, Then the result carries `_root{dir, source: "param"}`; and When `factory_complete` runs with `project_root`, Then the integration lock root resolves through the same git-common-dir resolver the CLI uses. | REQ-MN-004, REQ-MN-005 | EV-006 | M2 |
| AC-MN-007 | Given `project_root: "<home>"` and `project_root: "<sibling repo>"`, When a root-taking tool other than the audit/delegation surface is called, Then both are rejected; a registered worktree of the server's repository is accepted. Exception: the audit/delegation surface (`codex_audit`, `audit_multi`, `glm_audit`, `claude_audit`) is validated but not confined to the server's repository, because an audit legitimately names another repository's tree (`resolveValidatedToolProjectRoot`, `internal/cli/mcp_project_root.go`; `progress.md` M3 note 3); a sibling repository is therefore accepted there. What that surface validates is: the root is symlink-canonicalized, and it must carry a `.moai` or be an accepted linked worktree; `codex_audit`, `glm_audit`, and `claude_audit` also reject a non-string `project_root` (their resolver). KNOWN GAPS relative to this row's first clause, recorded as debt for the operator and not validation the surface performs: the home directory is NOT rejected on the audit surface (that rejection lives only in the confined set, `confineToolProjectRoot`), and `audit_multi` treats a non-string `project_root` as absent and audits the default tree | REQ-MN-005 | EV-007 | M2 |
| AC-MN-008 | Given the built server, When `tools/list` is called, Then `todo_claim` appears; given a test registering a name absent from the catalogue, Then the build-time gate fails it. | REQ-MN-006 | EV-008 | M2 |
| AC-MN-009 | Given the schema lint test over all served tools, When run on the fixed tree, Then it passes; run on a fixture missing a property description or carrying an undeclared handler-read argument, Then it fails. | REQ-MN-007 | EV-009 | M2 |
| AC-MN-010 | Given `factory_msg_list` with zero eligible targets, When called, Then `messages` is `[]` plus any already-claimed entries (never null); given a structured-result tool, Then the text block carries the same JSON as `structuredContent`. | REQ-MN-008, REQ-MN-017 | EV-010 | M2/M5 |
| AC-MN-011 | Given a codex gate child launch, When the argv is inspected, Then non-moai MCP servers are disabled and the moai server for gate sessions exposes exactly the read-only subset enumerated in plan §F M3 (the test asserts the served subset equals the enumerated set). | REQ-MN-009 | EV-011 | M3 |
| AC-MN-012 | Given the elicitation guard test, When it sweeps the server's request-initiation surface, Then zero elicitation requests are initiated toward clients; given a scripted elicitation request arriving on the gate client path, When handled, Then an automatic decline returns without stalling or interleaving stdin writes; given two concurrent writers to one `.moai/state` artifact, Then the `-race` test shows serialized writes with no corruption. | REQ-MN-010 | EV-012 | M3 |
| AC-MN-013 | Given an empty factory inbox, When the counting path inspects it, Then no write lock is taken and no write-probe deadline is paid, and no message is claimed; delivery remains the rendering path of REQ-MN-016. | REQ-MN-011 | EV-013 | M3 |
| AC-MN-014 | Given the claude_audit child env built by the allowlist, When inspected, Then it contains neither `CLAUDE_CONFIG_DIR` nor proxy/CA variables, pins the subscription config dir, honors `MOAI_CLAUDE_BIN`, and a forced 429 is labeled rate-limited with the endpoint unverified. | REQ-MN-012 | EV-014 | M4 |
| AC-MN-015 | Given the installed binary's mtime/commit changed relative to the serving process, When any tool is called, Then the result annotates `_server` + `stale_server: true` with reconnect guidance and a mutating call refuses. | REQ-MN-013 | EV-015 | M4 |
| AC-MN-016 | Given a background job whose originating request's context has ended, When the job is queried, Then it is still running or completed (never cancelled by the request) — the pinned GLM regression tests `TestGLMTask_BackgroundJobSurvivesRequestContextEnd` and `TestGLMTask_DetachedBackgroundJobIsStillCancellable` stay green (the third pin, `TestCodexTask_BackgroundJobSurvivesRequestContextEnd`: superseded — surface removed by card t1909); given a GLM job and a role-audit job alive at server exit, When the server shuts down, Then both are cancelled; a status query on a dead-owner record reports `stale`; `codex_role_audit_cancel` exists; a timed-out foreground codex_task returns partial output plus its thread id (superseded — surface removed by card t1909). | REQ-MN-014 | EV-016 | M4 |
| AC-MN-017 | Given a read-only project directory and a non-project cwd, When the server starts there, Then it still serves despite the stamp-write failure and creates no `.moai` in the foreign cwd; the rotated per-process call log gains one row per tool call. | REQ-MN-015 | EV-017 | M4 |
| AC-MN-018 | Given an endpoint with a live claim, When it re-claims, Then the token is unchanged and the lease extended; after a `/clear`, peer attribution resolves the live identity or refuses with guidance (never the stale spawn-time id). | REQ-MN-017 | EV-018 | M5 |
| AC-MN-019 | Given a receiver re-registering from generation 1 to 2 with pending gen-1 messages, When the re-register completes, Then each message is claimable at gen-2 or dead-lettered with sender notification — orphan count 0; given more messages than the hook byte cap, When the delivery path renders, Then the cut falls on a line boundary and claimed count equals visible count. | REQ-MN-016 | EV-019 | M5 |
| AC-MN-020 | Given a Stop event with only count-deliverable kinds, When the counting path runs, Then no block (no forced turn); given a blocker-class message, Then exactly one block; given a degraded broker, Then the Stop side logs the state. | REQ-MN-018 | EV-020 | M5 |
| AC-MN-021 | Given a forced transport failure on a required-gate review, When the leg completes, Then the verdict is not `fail`, `gate_unmet` carries could-not-run, stderr is preserved, and exactly one retry fired; given `codex_audit` asked for `target: baseBranch`, Then either the resolved base SHA and diff instruction are in the prompt or the result is inconclusive with "target not applied", and the second-review record states the effective scope. | REQ-MN-019, REQ-MN-021 | EV-021 | M6 |
| AC-MN-022 | Given a required backend inconclusive and the caller's own `claude_verdict: pass`, When convergence runs, Then the outcome is not-converged (never pass); given an unparsable model reply containing `provenance` fields, When the GLM decoder runs, Then none of the model-supplied provenance/receipt/gate/review-base fields survive. | REQ-MN-021 | EV-022 | M6 |
| AC-MN-023 | Given a receipt with `codex_verdict: fail` and a PASS line candidate, When the store evaluates it, Then corroboration is refused; given a receipt whose head/digest differs from the current tree, Then reuse is refused; given an `audit_cache` store call, Then the entry requires an existing verdict file whose SHA-256 matches and a non-SPEC directory is refused (never the empty-input hash); codex and GLM verdicts record the auditor model. | REQ-MN-020 | EV-023 | M6 |
| AC-MN-024 | Given a codex audit/review/convergence leg, When the child runs, Then a read-only sandbox policy and a scrubbed env apply; given truncated material or excluded untracked files, Then the result marks `truncated` / `excluded_untracked`; given empty material, Then the backend is not invoked; gate state is digest-keyed, atomically written, session-id-validated, and a could-not-run receipt is never reused as a pass within any TTL. | REQ-MN-022, REQ-MN-023, REQ-MN-024 | EV-024 | M6 |
| AC-MN-025 | Given the nested-worktree fixture reproducing the reviewer overlay (receipt written under one root, Codex Stop chain reading under another), When the chain evaluates, Then producer and reader resolve the same root and the false "receipt absent" negative no longer reproduces. | REQ-MN-025 | EV-025 | M6 |

## §C RED-now Evidence Ledger

Tree pinned for all plan-phase measurements: `a33f31fc3efe78c02e68336d0b85bb128adcf015` (branch WT-3-2-0-3).

```
EV-001 (AC-MN-001 · release-blocking · RED, measured)
command : go test ./internal/cli -run '^TestTodoClaimMCP_Mirror$' -count=1
stdout  : ok  	github.com/modu-ai/moai-adk/internal/cli	1.721s
exit    : 0
tree    : a33f31fc3efe78c02e68336d0b85bb128adcf015
red why : the test's refusal arm calls the MCP todo_claim handler in a lane session
          and asserts the CLI-identical refusal text; sdCallTool (factory_m3_test.go)
          returns an error only when res.IsError is true, so the PASSING assertion
          proves the policy refusal rides the error channel (toolErr sets IsError:
          true, mcp_server.go) — the exact behavior AC-MN-001 forbids.
green   : M1 flips it — the same call returns is_error=false with refused:true+code;
          the mirror test's refusal arm asserts the structured shape.
```

```
EV-002 (AC-MN-002 · release-blocking · RED, measured)
command : grep -oE "refused|tool_failure" internal/hook/failure_observer.go
stdout  : tool_failure
          tool_failure
          tool_failure
          tool_failure
exit    : 0
tree    : 43ae15cf0d1ad05ac2ef5ff77699ecaa0c251f13
red why : single invocation, no pipes — the positive control is IN the stdout
          (four tool_failure references: the observer classifies and records
          tool failures) and the probe is the measured absence (zero refused
          lines: no exclusion exists); with EV-001 (every policy refusal is an
          error today) every refusal flows into the tool_failure rows — the
          behavior AC-MN-002 forbids. Behavioral corroboration: the §1
          production stat of 409 refusal rows recorded as mcp failures; the
          isolated-inbox reproduction (N forced refusals + 1 genuine error →
          exactly 1 mcp row) is M1's stronger cell. Wired at
          internal/hook/post_tool_failure.go (recordToolFailureEvent).
green   : M1 flips it — the observer classifies refused:true payloads and
          excludes them; the forced-refusal fixture yields exactly 1 genuine
          row.
```

```
EV-003 (AC-MN-003 · release-blocking · RED, measured)
command : go test ./internal/cli -run '^(TestCodexReviewGate_SubcommandRegistered|TestCodexGateCallSuccessRate)$' -count=1 -v
stdout  : === RUN   TestCodexReviewGate_SubcommandRegistered
          --- PASS: TestCodexReviewGate_SubcommandRegistered (0.00s)
          PASS
          ok  	github.com/modu-ai/moai-adk/internal/cli	1.060s
exit    : 0
tree    : a33f31fc3efe78c02e68336d0b85bb128adcf015
red why : the selector was proven valid by the control arm (the existing codex
          gate test matched and ran), and the measurement arm matched nothing —
          no run line for TestCodexGateCallSuccessRate: the success-rate
          instrument the card gate requires does not exist. Measured absence
          with a same-command positive control, not a spelling miss.
green   : M3 flips it — the harness of plan §F M3 prints the metric line
          defined in AC-MN-003.
```

```
EV-004 (AC-MN-004 · milestone-gating · regression-guard, disposition stated)
red     : stale run ids are accepted and identity is env-only (factoryLaneAdmission
          reads env; no actor argument exists).
why not measured now : the observable (structured refused codes on a supplied
          actor argument) requires the argument surface M1 builds; no plan-phase
          command can name an argument no handler reads. Disposition: not
          recorded as a pass.
green   : M1's TDD RED step constructs the red; then the cell above's GWT.
```

```
EV-005 (AC-MN-005 · milestone-gating · regression-guard)
red     : operator authority is decided by env ABSENCE (factoryLaneRefusal,
          factory_card.go) — an unmarked session acts as operator.
why not measured now : the red-for-stated-reason needs the positive-mark
          surface M1 builds (before it, the unmarked session is legitimately
          "unknown", not yet refused). Disposition: not recorded as a pass.
green   : M1.
```

```
EV-006 (AC-MN-006 · milestone-gating · regression-guard)
red     : non-string project_root is coerced and succeeds (lens-7 B1 probe:
          project_root:42 → source=server-cwd, success); lockRoot is not
          caller-passed (REQ-MN-004 clause absent today).
why not measured now : the probe needs a running server session or the M2
          in-process harness; the coercion site is verified in-tree
          (resolveToolProjectRootWithSource, mcp_project_root.go). Disposition:
          not recorded as a pass.
green   : M2.
```

```
EV-007 (AC-MN-007 · milestone-gating · regression-guard)
red     : root validation accepts any .moai-bearing directory incl. $HOME
          (lens-7 B2 probe; validateProjectRoot branch verified in-tree).
why not measured now : same instrument gap as EV-006. Disposition: not
          recorded as a pass.
green   : M2.
```

```
EV-008 (AC-MN-008 · milestone-gating · regression-guard)
red, measured : grep -oE "todo_claim|todo_list|todo_add" internal/mcp/catalog.go
          | sort | uniq -c → "   2 todo_add" / "   2 todo_list" / (todo_claim
          absent), exit 0, tree a33f31fc3e — sibling control present, todo_claim
          measured absent from the catalogue (the registered tool is dead).
why regression-guard : the row gates the M2 milestone, not the release; the
          tools/list half needs the built server.
green   : M2.
```

```
EV-009 (AC-MN-009 · milestone-gating · regression-guard)
red     : schemas carry the eleven undocumented factory_msg_* properties, the
          undeclared codex_task.model read (superseded — surface removed by
          card t1909), unset additionalProperties, wrong hints (lens-7 schema
          section); no schema lint test exists.
why not measured now : the lint test is itself the M2 deliverable. Disposition:
          not recorded as a pass.
green   : M2.
```

```
EV-010 (AC-MN-010 · milestone-gating · regression-guard)
red     : toolJSON emits "<tool>: ok" with data only in structuredContent
          (mcp_server.go:1106-1108 area); factory_msg_list serializes nil as
          "messages": null.
why not measured now : observable through the M2/M5 harness. Disposition: not
          recorded as a pass.
green   : M2 (shape) / M5 (list).
```

```
EV-011 (AC-MN-011 · milestone-gating · regression-guard)
red     : codex gate/audit children do not disable non-moai MCP servers and no
          gate-session subset exists (the containment codexAuditMCPDisableArgs
          applies only on the role-audit path).
why not measured now : the launch-argv instrument is M3's. Disposition: not
          recorded as a pass.
green   : M3.
```

```
EV-012 (AC-MN-012 · milestone-gating · regression-guard)
red     : the 38/38 gate-session MCP call failures with elicitation unhandled
          and the stdin double-write (redesign §6.3 problem column; audit-era
          measurement, not re-executable on this tree); SDK direction verified
          in-tree: mcp-go v1.1.1 (go.mod), zero elicitation call sites in
          internal/ — moai's server never initiates, so the live defect is the
          gate client path.
why not measured now : the guard test and the scripted elicitation harness are
          M3 deliverables. Disposition: not recorded as a pass.
green   : M3.
```

```
EV-013 (AC-MN-013 · milestone-gating · regression-guard)
red     : the empty-inbox read path takes the write lock under a 200ms probe —
          1,619 recorded "inbox context deadline exceeded" (t1622-6; audit-era
          count, cited not re-executed).
why not measured now : the lock-free counting path is M3's build. Disposition:
          not recorded as a pass.
green   : M3.
```

```
EV-014 (AC-MN-014 · milestone-gating · regression-guard)
red     : scrubClaudeAuditEnv is a denylist (CLAUDE_CONFIG_DIR and proxy pass
          through), 429 labeled subscription capacity, claudeLookPath bypasses
          the pin (lens-3 #8; sites verified in-tree).
why not measured now : the allowlist env builder is M4's. Disposition: not
          recorded as a pass.
green   : M4.
```

```
EV-015 (AC-MN-015 · milestone-gating · regression-guard)
red     : staleness advertised only at initialize; no per-call comparison
          against the installed binary (lens-7 B9).
why not measured now : the per-call annotation is M4's. Disposition: not
          recorded as a pass.
green   : M4.
```

```
EV-016 (AC-MN-016 · milestone-gating · regression-guard)
red     : GLM/role-audit jobs never cancelled on server exit; no owner stamp,
          no liveness check, no role-audit cancel tool; foreground timeout
          loses the thread (lens-3 #16/17; sites verified in-tree). The
          foreground codex_task thread clause: superseded — surface removed
          by card t1909.
why not measured now : the survival assertion is real but its GREEN form
          depends on M4's lifecycle changes; running the three pinned
          regression tests today shows the survival contract the fix must
          preserve, not the missing cancellation. Disposition: not recorded
          as a pass.
green   : M4 — the three regression tests stay green AND the new
          cancellation/stale/cancel/partial-output assertions pass. Since
          2026-10-23 the codex pin and the codex partial-output assertion are
          superseded — surface removed by card t1909; the two GLM pins and
          the GLM/role-audit cancellation, stale and cancel-tool assertions
          remain the green form.
```

```
EV-017 (AC-MN-017 · milestone-gating · regression-guard)
red     : stamp write fails closed and writeMCPServerRuntimeRecord creates
          .moai in any cwd (lens-7 B10; site verified in-tree); no call log.
why not measured now : the fixtures (read-only dir, foreign cwd server start)
          are M4's. Disposition: not recorded as a pass.
green   : M4.
```

```
EV-018 (AC-MN-018 · milestone-gating · regression-guard)
red     : every re-claim mints a new claim token (three re-claim sites; the
          recorded "claim identity mismatch" failures); currentFactoryPeer
          serves the spawn-time session id after /clear.
why not measured now : the token-stability contract is M5's store change; the
          existing Claim tests pin the rotation behavior the fix reverses.
          Disposition: not recorded as a pass.
green   : M5 — rotation-pinning tests updated with the behavior change, the
          new stability tests pass.
```

```
EV-019 (AC-MN-019 · milestone-gating · regression-guard)
red     : generation advance orphans messages (9 real orphans in the production
          DB; lens-4 §5 mechanism verified in code); the 2,048-byte cut claims
          messages it never renders.
why not measured now : the migration/dead-letter path and the line-boundary cut
          are M5 builds. Disposition: not recorded as a pass.
green   : M5.
```

```
EV-020 (AC-MN-020 · milestone-gating · regression-guard)
red     : the leader Stop hook blocks on every pending batch (every natural
          turn costs one forced turn; lens-6 §5).
why not measured now : the kind-scoped block is M5's. Disposition: not
          recorded as a pass.
green   : M5.
```

```
EV-021 (AC-MN-021 · milestone-gating · regression-guard)
red     : applyGateUnmet/enforceRequiredGateUnmet convert transport failure to
          verdict fail; the codex leg ignores the target while the second-review
          record claims baseBranch (lens-3 #2/#10; sites verified in-tree).
why not measured now : the could-not-run verdict path and the target-honest
          prompt are M6 builds. Disposition: not recorded as a pass.
green   : M6.
```

```
EV-022 (AC-MN-022 · milestone-gating · regression-guard)
red     : claudeVerdictOrDefault passes convergence on the caller's own verdict
          when a required backend is inconclusive; parseGLMReview unmarshals
          model text into the full ReviewOutput (lens-3 #13/#21).
why not measured now : the not-converged outcome and the narrow decoder are M6
          builds. Disposition: not recorded as a pass.
green   : M6.
```

```
EV-023 (AC-MN-023 · milestone-gating · regression-guard)
red     : auditreceipt/store.go rejects only GateUnmet/inconclusive — a fail
          receipt returns corroboration-eligible; receipts carry no head/digest
          (lens-3 #3); audit_cache is process-local, self-attesting, and
          compute_hash returns the empty-input hash for non-SPEC dirs (lens-3
          #24); no model provenance on codex/GLM verdicts.
why not measured now : the binding fields and validations are M6 builds.
          Disposition: not recorded as a pass.
green   : M6.
```

```
EV-024 (AC-MN-024 · milestone-gating · regression-guard)
red     : audit/review legs forward sandboxPolicy only when supplied and set no
          cmd.Env; truncation/untracked exclusion unmarked; codex_audit lacks
          the empty-material guard; gate state is per-session, non-atomic,
          unvalidated session_id path join, inconclusive reused 24h (lens-3
          #7/#9/#11/#12; sites verified in-tree).
why not measured now : the containment and state-integrity surfaces are M6
          builds. Disposition: not recorded as a pass.
green   : M6.
```

```
EV-025 (AC-MN-025 · milestone-gating · regression-guard)
red     : the reviewer overlay reproduction — receipt producer writes at one
          root, the Codex Stop chain reader resolves another, false "receipt
          absent" in nested sessions (t1597 issuer evidence).
why not measured now : the seam contract and its nested-worktree fixture are
          M6 builds; the reproduction predates this tree. Disposition: not
          recorded as a pass.
green   : M6.
```

## §D Edge cases

- `project_root` pointing at `/tmp`, a non-git directory, and a fresh repo with no commits (clear errors, no partial writes).
- Concurrent claims from two distinct endpoints on the same inbox (raced path — one wins, the loser gets `raced`).
- Lease expiry between claim and read (token genuinely expired → re-claim mints a new token; the old token's read refuses with guidance).
- A gate session reconnecting after a binary reinstall mid-session (staleness annotation + reconnect guidance).
- Stop firing while the broker DB is unavailable (degraded-state log, no crash, no silent swallow).
- `session_id`/`job_id` values carrying `../` (validated, refused).

## §E Quality gates

- TRUST 5: Tested (per-change coverage ≥85% on touched packages, `-race` + `-count` repeats for concurrency tests), Readable, Unified (gofmt), Secured (path-join and env-allowlist validations), Trackable (Conventional Commits, card id in every commit).
- `go build ./...` and `GOOS=windows GOARCH=amd64 go build ./...` exit 0.
- golangci-lint: zero NEW findings vs baseline.

## §F Definition of Done

All 25 ACs green with attribution triples (command + verbatim output + HEAD SHA); the three release-blocking gates measured with numbers per their ledger cells; RED-now evidence for every Milestone-gating row produced at its Milestone's RED step and archived (E8); no Out-of-Scope item absorbed. **The card is not closed while any of the three release-blocking gates is unmeasured** — each gate is measured by the Milestone that flips it (M1, M1, M3 respectively), and closure reports all three numbers.
