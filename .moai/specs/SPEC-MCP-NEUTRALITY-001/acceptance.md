---
id: SPEC-MCP-NEUTRALITY-001
title: "Acceptance criteria — MCP refusal structuring and harness neutrality"
version: "0.1.0"
created: 2026-10-14
updated: 2026-10-14
author: manager-spec
---

# SPEC-MCP-NEUTRALITY-001 — Acceptance Criteria

## §A Adoption discipline (RED-now + green path)

Per `.claude/rules/moai/development/verification-completeness.md` §2, every release-blocking AC below is adopted as a TWO-CELL pair at run phase: a **RED-now cell** (the criterion observed red on the pre-implementation tree, pinned to the tree SHA) and a **green path cell** (the Milestone that flips it and the passing output). The run phase measures the RED cell first — an AC never observed red has proven nothing. RED must be red for the stated reason: each AC below names the defect behavior that makes it red today. Verification commands are single-invocation, read-only forms run from the worktree root; test names are the run-phase worker's to fix precisely (the names below indicate the family).

## §B AC Matrix

| AC | Given / When / Then | REQ | Flipped by |
|----|---------------------|----------|-----------|
| AC-MN-001 | Given a fixture holding a live serial-slot lease owned by another endpoint, When a factory verb is called through the MCP tool path, Then the response has `is_error` false, structured `refused: true` with `code: serial-slot`, and the text block carries the same refusal JSON. RED today: the path returns `isError: true`. | REQ-MN-001 — **card gate 1: 정책 거부 is_error 0** | M1 |
| AC-MN-002 | Given the failure observer with a scratch inbox file, When the suite forces N policy refusals and exactly 1 genuine transport error, Then the inbox gains exactly 1 new mcp row and 0 refusal rows. RED today: refusal rows are recorded as mcp failures. | REQ-MN-002 — **card gate 2: lessons-inbox mcp 신규 0** | M1 |
| AC-MN-003 | Given the gate-session measurement harness, When the measurement runs against the Codex gate fixtures, Then it prints a numeric success rate (percent) with the sample count; a run that prints no number is a FAIL. | REQ-MN-009 + gate path — **card gate 3: Codex 게이트 호출 성공률 측정** | M3 |
| AC-MN-004 | Given a spawn recorded as lane-A, an active run R1 and a completed run R0, When the tool is called with `actor: lane-B` on R1 and with `actor: lane-A` on R0, Then both refuse (`refused: true`, codes `owned` / `stale-run`, `is_error` false) and a lane-A call on R1 proceeds. | REQ-MN-003 | M1 |
| AC-MN-005 | Given a session with no lane env and no operator mark, When `factory_decide` is called, Then it refuses with unknown attribution; with the positive operator mark it proceeds. RED today: the unmarked session acts as operator. | REQ-MN-003 | M1 |
| AC-MN-006 | Given the server on a registered worktree, When a root-less family tool receives `project_root: 42`, Then an invalid-argument error results (no silent coercion); When a mutating family tool omits `project_root`, Then a refusal with a naming hint; When called correctly, Then the result carries `_root{dir, source: "param"}`. | REQ-MN-004 | M2 |
| AC-MN-007 | Given `project_root: "<home>"` and `project_root: "<sibling repo>"`, When a root-taking tool is called, Then both are rejected; a registered worktree of the server's repository is accepted. RED today: `$HOME` succeeds. | REQ-MN-005 | M2 |
| AC-MN-008 | Given the built server, When `tools/list` is called, Then `todo_claim` appears; given a test registering a name absent from the catalogue, Then the build-time gate fails it. | REQ-MN-006 | M2 |
| AC-MN-009 | Given the schema lint test over all served tools, When run on the fixed tree, Then it passes; run on a fixture missing a property description or carrying an undeclared handler-read argument, Then it fails. | REQ-MN-007 | M2 |
| AC-MN-010 | Given `factory_msg_list` with zero eligible targets, When called, Then `messages` is `[]` plus any already-claimed entries (never null); given a structured-result tool, Then the text block carries the same JSON as `structuredContent`. | REQ-MN-008, REQ-MN-017 | M2/M5 |
| AC-MN-011 | Given a codex gate child launch, When the argv is inspected, Then non-moai MCP servers are disabled and the moai server for gate sessions exposes only the read-relevant subset. | REQ-MN-009 | M3 |
| AC-MN-012 | Given an elicitation request from the client, When the server handles it, Then an automatic decline response returns without stalling or interleaving stdin writes; given two concurrent writers to one `.moai/state` artifact, Then the `-race` test shows serialized writes with no corruption. | REQ-MN-010 | M3 |
| AC-MN-013 | Given an empty factory inbox, When a reader inspects it, Then no write lock is taken and no write-probe deadline is paid; hook readers observe counts without claiming. RED today: the path produced 1,619 deadline-exceeded errors. | REQ-MN-011 | M3 |
| AC-MN-014 | Given the claude_audit child env built by the allowlist, When inspected, Then it contains neither `CLAUDE_CONFIG_DIR` nor proxy/CA variables, pins the subscription config dir, honors `MOAI_CLAUDE_BIN`, and a forced 429 is labeled rate-limited with the endpoint unverified. | REQ-MN-012 | M4 |
| AC-MN-015 | Given the installed binary's mtime/commit changed relative to the serving process, When any tool is called, Then the result annotates `_server` + `stale_server: true` with reconnect guidance and a mutating call refuses. | REQ-MN-013 | M4 |
| AC-MN-016 | Given a GLM job and a role-audit job alive at server exit, When the server shuts down, Then both contexts are cancelled; a status query on a dead-owner record reports `stale`; `codex_role_audit_cancel` exists; a timed-out foreground codex_task returns partial output plus its thread id. | REQ-MN-014 | M4 |
| AC-MN-017 | Given a read-only project directory and a non-project cwd, When the server starts there, Then it still serves despite the stamp-write failure and creates no `.moai` in the foreign cwd; the rotated per-process call log gains one row per tool call. | REQ-MN-015 | M4 |
| AC-MN-018 | Given an endpoint with a live claim, When it re-claims, Then the token is unchanged and the lease extended; after a `/clear`, peer attribution resolves the live identity or refuses with guidance (never the stale spawn-time id). RED today: every re-claim mints a new token. | REQ-MN-017 | M5 |
| AC-MN-019 | Given a receiver re-registering from generation 1 to 2 with pending gen-1 messages, When the re-register completes, Then each message is claimable at gen-2 or dead-lettered with sender notification — orphan count 0; given more messages than the hook byte cap, When the hook renders, Then the cut falls on a line boundary and claimed count equals visible count. | REQ-MN-016 | M5 |
| AC-MN-020 | Given a Stop event with only count-deliverable kinds, When the hook runs, Then no block (no forced turn); given a blocker-class message, Then exactly one block; given a degraded broker, Then the Stop side logs the state. | REQ-MN-018 | M5 |
| AC-MN-021 | Given a forced transport failure on a required-gate review, When the leg completes, Then the verdict is not `fail`, `gate_unmet` carries could-not-run, stderr is preserved, and exactly one retry fired; given `codex_audit` asked for `target: baseBranch`, Then either the resolved base SHA and diff instruction are in the prompt or the result is inconclusive with "target not applied", and the second-review record states the effective scope. RED today: transport failure becomes verdict fail; the codex leg silently ignores the target. | REQ-MN-019, REQ-MN-021 | M6 |
| AC-MN-022 | Given a required backend inconclusive and the caller's own `claude_verdict: pass`, When convergence runs, Then the outcome is not-converged (never pass); given an unparsable model reply containing `provenance` fields, When the GLM decoder runs, Then none of the model-supplied provenance/receipt/gate/review-base fields survive. | REQ-MN-021 | M6 |
| AC-MN-023 | Given a receipt with `codex_verdict: fail` and a PASS line candidate, When the store evaluates it, Then corroboration is refused; given a receipt whose head/digest differs from the current tree, Then reuse is refused; codex and GLM verdicts record the auditor model. RED today: a FAIL receipt reinforces PASS. | REQ-MN-020 | M6 |
| AC-MN-024 | Given a codex audit/review/convergence leg, When the child runs, Then a read-only sandbox policy and a scrubbed env apply; given truncated material or excluded untracked files, Then the result marks `truncated` / `excluded_untracked`; given empty material, Then the backend is not invoked; gate state is digest-keyed, atomically written, session-id-validated, and a could-not-run receipt is never reused as a pass within any TTL. | REQ-MN-022, REQ-MN-023, REQ-MN-024 | M6 |
| AC-MN-025 | Given the nested-worktree fixture reproducing the reviewer overlay (receipt written under one root, Codex Stop chain reading under another), When the chain evaluates, Then producer and reader resolve the same root and the false "receipt absent" negative no longer reproduces. RED today: the overlay reproduction reports the receipt absent. | REQ-MN-025 | M6 |

## §C Edge cases

- `project_root` pointing at `/tmp`, a non-git directory, and a fresh repo with no commits (clear errors, no partial writes).
- Concurrent claims from two distinct endpoints on the same inbox (raced path — one wins, the loser gets `raced`).
- Lease expiry between claim and read (token genuinely expired → re-claim mints a new token; the old token's read refuses with guidance).
- A gate session reconnecting after a binary reinstall mid-session (staleness annotation + reconnect guidance).
- Stop firing while the broker DB is unavailable (degraded-state log, no crash, no silent swallow).
- `session_id`/`job_id` values carrying `../` (validated, refused).

## §D Quality gates

- TRUST 5: Tested (per-change coverage ≥85% on touched packages, `-race` + `-count` repeats for concurrency tests), Readable, Unified (gofmt), Secured (path-join and env-allowlist validations), Trackable (Conventional Commits, card id in every commit).
- `go build ./...` and `GOOS=windows GOARCH=amd64 go build ./...` exit 0.
- golangci-lint: zero NEW findings vs baseline.
- The three card completion criteria (AC-MN-001/002/003) are the terminal gates — a Milestone is not done while any of them is unmeasured.

## §E Definition of Done

All 25 ACs green with attribution triples (command + verbatim output + HEAD SHA); RED-now evidence archived per cluster (E8); the three card gates measured and reported with numbers; no Out-of-Scope item absorbed; `progress.md` §E.2/§E.3 populated by manager-develop.
