# t341 M0 — does the live PostToolUse payload carry Bash stdout?

Card t341 · SPEC-SELECTOR-CENSUS-001 · AC-SEC-000
Observed 2026-08-30 in `.claude/worktrees/t341`, branch `WT-selector-census`, HEAD `744cfab5e`
(develop `ee50984ab` absorbed). Session `ebe06935-2c47-4d8a-8cc4-04df14519ace`.

## Claim

**Yes — the live payload reaches `deriveFromOutputText`, and the output text decides.**
Measured behaviourally, not by capturing the payload.

Secondary, and consequential: in the observed call the **exit-code branch did not fire on a
shell exit code of 0**. The most economical reading is that the live payload carries no
`exit_code` for a Bash call, which would make `deriveFromOutputText` the live-dominant path
and `deriveFromExitCode` reachable mainly through the legacy `ToolOutput` fallback and tests.

## Evidence

The two classifier branches were made to **disagree**, so the record identifies which one ran.
A deliberately failing test was piped through `cat`, making the shell exit code 0 while stdout
carried a `--- FAIL` precise marker:

    $ go test ./internal/hook/ -run '^TestT341M0Probe$' -count=1 | cat
    --- FAIL: TestT341M0Probe (0.00s)
        zzz_t341_m0_probe_test.go:10: t341 M0 probe: deliberate failure
    FAIL
    FAIL	github.com/modu-ai/moai-adk/internal/hook	0.595s
    FAIL

Shell exit code: **0** (`cat` is the last stage of the pipeline).

The record the live hook wrote, tail of
`/Users/goos/MoAI/moai-adk-go/.moai/evolution/telemetry/usage-2026-08-30.jsonl`
(sink per `internal/telemetry/recorder.go:38-47`):

    {"ts":"2026-08-30T11:05:50.099478Z","session_id":"ebe06935-…","outcome":"success","path_kind":"code-change"}
    {"ts":"2026-08-30T11:05:54.769815Z","session_id":"ebe06935-…","outcome":"error","is_test_fail":true}

The first row is the `Write` that created the probe file; the second is the `go test` call.

## Reading

`classifyTestCommand` consults `deriveFromExitCode` first (`evidence_writer.go:69`), which on
`exit_code == 0` returns **pass** (`:163`) and never reaches the text branch. The record says
`is_test_fail: true`. A pass-returning exit-code branch cannot produce that, so the fail came
from the text branch's `--- FAIL` precise marker (`:210`) — which means the payload carried
the stdout for it to match.

The three questions AC-SEC-000 asks, answered as far as this method reaches:

- **(a) stdout present** — yes, established by the disagreement above.
- **(b) `exit_code` location** — **not established**. The observation shows only that no
  exit-code *pass* signal fired for a shell rc of 0. Absent field and present-but-different
  value are not separated by this probe.
- **(c) wrapped JSON vs plain** — **not established**. `decodeToolResponse`
  (`evidence_writer.go:111`) handles both, so a match proves nothing about which arrived.

## Gaps

- **The raw payload was not captured.** AC-SEC-000's literal artifact
  (`.moai/reports/t341/live-payload.json`, `jq -e .`) does not exist, and this file does not
  substitute for it. The only capture points are the shell wrapper and the hook binary, both
  in the **primary checkout**, shared with other live sessions — editing either to log
  payloads was rejected rather than done quietly. AC-SEC-000 is therefore **partially**
  satisfied: its underlying question (a) is answered, its artifact is still owed.
- (b) and (c) remain open, as stated above.

## Residual risk

If the live payload genuinely carries no `exit_code`, then AC-SEC-003's sample (f) — the node
built-in runner, whose only pass signal today is the exit-code path — pins a branch that does
not fire in production. That does not make (f) wrong: the function is reachable through the
legacy `ToolOutput` fallback and directly under test, and the veto is specified ahead of both
branches regardless. It does mean the *live* consequence of (f) is smaller than its wording
suggests, and M1 should record which branch it actually observed for each runner rather than
inferring from this one call.
