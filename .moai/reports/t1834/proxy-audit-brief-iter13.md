# Proxy audit brief — card t1834, sync-audit iteration 13 (post-repair confirmation)

Role: sync-auditor (post-implementation skeptical evaluation). Card t1834, SPEC-INTEGRATION-STALE-BASE-001. Authority for this round: the leader's 10-21 re-dispatch of the card with completion instructions (run tn8rpj), recorded in progress.md §E.2.8 as the renewed operator-side authorization after decision d-20261018T222141Z-398b ended iter12's audit rounds. Exactly ONE confirming audit runs on this authorization.

## Audited object

- Tree (read-only except the verdict file): /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1834 — branch WT-10-17-class-8
- Audited SHA = tree HEAD: `1a16da10f6edc7b78c22bdb8460908108ab220b9` (clean tree; local main fully absorbed — verify `git merge-base --is-ancestor main HEAD` rc=0 yourself)
- project_root for every moai MCP call: /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1834
- Verdict output path: /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1834/.moai/reports/t1834/sync-audit-iter13.md
- delta_base: none (full HEAD-confirmation audit; judge this HEAD on its own evidence, prior verdicts do not carry over)
- Gates: codex = required (audit_multi / codex_audit through the moai MCP per the config gates; do not hand-pick gates; if the required leg cannot run, return a blocker report — never substitute another backend or hand-write a gate result).

## History since iter12 (read progress.md §E.2.8 and the named files)

- iter12 (sync-audit-iter12.md) FAILED at afce6d2f78 with two P2 blocking findings:
  - F1: `factoryAdoptedTip` (internal/cli/factory_card.go) accepted a re-measure record found under the tip's tree key without checking `record.Tree == tipTree` — the adoption exemption bypassed the gate's identity check (REQ-ISB-016).
  - F2: `factoryBranchOfWorktree` used `git rev-parse --abbrev-ref HEAD`, which prints `heads/<branch>` when a tag shares the branch name; callers built `refs/heads/heads/<branch>` and `complete` failed for healthy candidates.
- Since then (all commits end with the `🗿 MoAI` marker and carry `card t1834`; the lane verified fix lines, commit subjects, markers, and the clean tree on disk):
  - `2439da8c90` absorb merge (local main, 103 landings behind → 0)
  - `26dad7f9ef` RED F1 record-identity test
  - `e9e95f4d40` F1 fix (identity check inside `factoryAdoptedTip`, line ~2889)
  - `7891703291` RED F2 same-name-tag test v1 (merge-step end-to-end)
  - `1eaf9bb649` F2 end-to-end rescope to the adoption path — reason: the first end-to-end run exposed a PRE-EXISTING same-class defect on a non-card path (below), so the end-to-end subtest was rescoped to the path this card's fix actually owns; the resolution subtest RED was observed at the original commit
  - `1a16da10f6` F2 fix (`git symbolic-ref --quiet HEAD`, only `refs/heads/` stripped; detached HEAD keeps the former best-effort `""`; the fix comment states the hazard and the no-new-semantics rule)
- Transparency disclosure carried from the implementer's report: commit `26dad7f9ef` was re-created seconds after first creation via `git reset --soft HEAD~1` to repair a mangled trailer literal; no `--amend`, nothing pushed, nothing cited the broken form; the commit graph (test commit before fix commit) witnesses RED-before-GREEN.
- Confirming scoped runs already observed at 1a16da10f6 by the implementer (its Section E): factory Freshness|Remeasure ok 12.654s; config BaseFreshness|ShippedConfigKeys ok 3.655s; cli Freshness|TestIntegrationAcquireRefusesStaleBase|TestFactoryComplete|TestAcquire|TestIntegration ok 109.710s; gofmt clean; vet clean. Treat these as the implementer's claims; your Go evidence below is independent.

## Known context (no scope carve-outs — you judge everything you find)

- `internal/factory/integration_merge_step.go:950` reads the branch with `rev-parse --abbrev-ref HEAD` — same class as F2, PRE-EXISTING on a non-card path (the card's plan-pin preserve check showed this file untouched), observed live during the F2 end-to-end work, recorded in §E.2.8, routed for separate repair. Judge whether it blocks THIS head; do not treat it as already accepted.
- Other `--abbrev-ref` readers reported by grep (observation only, untested): hook_install.go:87, session_worktree.go:768/870, session_worktree_automerge.go:311, and iter12 F3 (`internal/cli/integration.go:126`).

## Scope of the audit

1. The card's non-.moai diff against main (`git diff main...HEAD --name-only -- . ':!.moai'`) must equal exactly 19 paths — iter12's 18 plus `internal/cli/factory_card_pr_gitenv_test.go` — and the factory-dispatch-gates.md pair must pass `diff -q`.
2. Read the repairs and judge: do they close iter12 F1 and F2; did the repairs introduce defects of their own (reverse cases: a healthy record naming its key tree is still adopted; a forged record naming another tree is refused on BOTH the pure gate and the complete path; same-named tag resolution returns the plain branch; same-named tag keeps the adoption path healthy; detached HEAD behavior unchanged); is every text statement consistent with the code (the F2 fix comment, the CHANGELOG entry, plan.md/spec.md/acceptance.md wording vs the rescope, progress.md §E.2.8 claims vs the tree).
3. Go evidence is REQUIRED this round, quoted verbatim: `go build ./...`; `go vet ./internal/cli/ ./internal/factory/ ./internal/hook/`; and the scoped tests, first pinned with `go test -list` per package, then ONE `go test -p 1 -count=1 -json` per scope with the raw JSONL saved under /private/tmp/t1834-iter13/ and the REAL exit code recorded; read failing names from the JSONL, never from tail/grep-trimmed text; do not re-run a scope already observed at this HEAD only to get more output. Scopes (one plain -run pattern per invocation — the worktree guard refuses '|' alternation; plain commands, no heredocs, no computed command text): ./internal/factory/ -run Freshness; ./internal/factory/ -run Remeasure; ./internal/cli/ -run Freshness; ./internal/cli/ -run Remeasure; ./internal/cli/ -run TestFactoryCompleteAdoptionRequiresRecordTreeIdentity; ./internal/cli/ -run TestFactoryBranchOfWorktreeWithSameNameTag; ./internal/cli/ -run TestAcquire; ./internal/cli/ -run TestIntegrationAcquire; ./internal/config/ -run BaseFreshness; and the template/ledger test covering the card's factory-dispatch-gates.md edit.
   Go entry conditions (boards 8c1e, e1b8, df10): before EVERY Go command (a) `moai slot status --resource go-test-heavy` must NOT show a landing or "remeasure record" run holding the slot (never acquire or force the slot yourself); (b) 1-minute load (`uptime`) <= 32; (c) memory free (`memory_pressure | grep -i 'free percentage'`) >= 50%. Wait with a Monitor until-loop; never a busy loop. ONE Go invocation at a time, `-p 1`. A Go check you could not run is a Gap and must be said plainly.

## Rules for the auditor

- No pattern kills (pkill, killall, kill $(pgrep ...)) — stop only a PID you started and recorded; stop agents with TaskStop; if a guard refuses a command, do not retry it.
- Do not edit any file except the verdict file; do not commit; do not push; do not ask questions; if blocked, return a blocker report with verbatim refusal text.

## Output format

Write the verdict to the path above, mirroring sync-audit-iter12.md's structure: `verdict:` and `audited_sha:` machine keys once each (audited_sha = 1a16da10f6edc7b78c22bdb8460908108ab220b9), the audit_multi digest (one line), ALL receipt ids of your own calls, the judging build commit next to the tree HEAD (state whether that build is a strict ancestor of the tree HEAD and what that means for what the gate could see), Claim / Evidence / Baseline-attribution / Gaps / Residual-risk, the test-list pins, the JSONL paths with real exit codes, the verbatim build/vet lines with the slot/load/memory readings taken before each Go command, and the final line `AUDIT-VERDICT: <PASS|PASS-WITH-DEBT|FAIL> spec=SPEC-INTEGRATION-STALE-BASE-001 receipts=<ids>`.

The lane (lane-6) stops writing to the tree from now until this verdict lands (CLI lease renewal continues).
