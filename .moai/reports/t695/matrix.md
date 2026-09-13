# t695 Matrix — full cell table with real gateway requests

Every cell below is a REAL request through a real gateway child process
(started from this worktree's build, ephemeral loopback port, isolated session
token, real credential store read in place). Raw per-cell output with verbatim
bodies: `results-*.jsonl` excerpts are quoted in investigation.md; the probe
tool is `probe_driver_main.go.txt`.

Legend: body = verbatim error body or reply excerpt. All 200 replies were real
model replies (`"ok"` / tool_use blocks / `"3"`), not synthesized.

## A. Operator live repro (pre-fix, production gateway) — recorded, not run by this lane

| ts (2026-09-13) | cell | result |
|---|---|---|
| ~17:32 | astra + low | API Error: 400 Bad Request (masked) |
| ~17:32 | sol + low | API Error: 400 Bad Request (masked) |
| ~17:32 | terra + default effort | SUCCESS |
| ~17:32 | terra + /effort low | API Error: 400 Bad Request (masked) |
| ~17:3x | /effort xhigh | API Error: 400 Bad Request (masked) |
| earlier | luna + low (lane-7, 6d392874) | 400 Bad Request at 08:07:14Z (masked) |
| earlier | sol + high (lanes 8/9/10) | OK across 3 lanes |
| 08:33:32Z + 08:42:33Z | sol + settings effortLevel=xhigh (family 3455fea2) | 400 (masked) |
| 08:32:27Z | sol effortLevel=low (family e28832af) | 400 (masked) |
| 08:33:01Z | terra first request (family 115e12d0, effort unconfirmed) | 400 (masked) |
| 08:31:57Z | astra first request (family b20da60c, effort unconfirmed) | 400 (masked) |
| 08:42:43Z | family 0efb66f7 | one 502 Bad Gateway (upstream-class, out of t695 scope) |

## B. Instrumented control pair (this lane, D1 build, pre-D3 policy) — CONFIRMED

| cell | status | body (verbatim) | ms |
|---|---|---|---|
| sol + low | 400 | `{"error":{"message":"unsupported effort","type":"invalid_request_error"},"type":"error"}` | 3 |
| luna + low | 400 | `{"error":{"message":"unsupported effort","type":"invalid_request_error"},"type":"error"}` | 0 |
| luna + high | 200 | `{"content":[{"text":"ok",...}],"model":"gpt-5.6-luna",...}` | 2542 |
| sol + medium | 200 | `{"content":[{"text":"ok",...}]}` | 2067 |
| sol + (no effort) | 200 | `{"content":[{"text":"ok",...}]}` | 1522 |

Control-pair verdict: model-independent, effort-axis defect CONFIRMED (the
400s are local: <=3 ms, no upstream round trip, real reason now visible).

## C. Post-fix plain cells — 4 models x 5 efforts x fresh (direct gateway HTTP, receipts = production-equivalent)

Final clean build. 20/20 GREEN.

| model | low | medium | high | xhigh | max |
|---|---|---|---|---|---|
| astra | 200 ok | 200 ok | 200 ok | 200 ok | 200 ok (wire: xhigh) |
| sol | 200 ok | 200 ok | 200 ok | 200 ok | 200 ok |
| terra | 200 ok | 200 ok | 200 ok | 200 ok | 200 ok |
| luna | 200 ok | 200 ok | 200 ok | 200 ok | 200 ok |

Pre-fix the same 20 cells would be 15 green / 5 red (low/xhigh/max rejected
locally); operator live repro already showed 4 of the red cells in production.

max stability sampling (same request repeated): sol 4/4 200, terra 4/4 200,
luna 6/7 200 (1 non-200 = receipt-less probe artifact, see investigation.md
§4; production-equivalent runs are green), astra raw max 0/4 (per-model
rejection) vs astra mapped-to-xhigh 3/3 200.

## D. Post-fix tool-bearing cells — 4 models x 5 efforts (direct gateway HTTP, receipts)

Single echo tool, tool_choice auto, prompt asks for a tool call. 20/20 GREEN —
each returned a real `tool_use` block (`{"name":"echo","input":{"text":"ok"},...}`,
`stop_reason":"tool_use"`).

## E. Post-fix subagent-style spawn-shaped cells (direct gateway HTTP, receipts)

Request shape mirrors a spawn's first turn: coordinator system prompt + an
`Agent` tool definition (`prompt`, `model` parameters). 8/8 GREEN — each
returned a real Agent `tool_use` call:

| cell | status | notes |
|---|---|---|
| spawnshape sol low / high | 200 / 200 | tool_use with model+prompt args |
| spawnshape terra low / high | 200 / 200 | tool_use |
| spawnshape luna low / high | 200 / 200 | tool_use |
| spawnshape astra low / high | 200 / 200 | tool_use |

## F. Model-arg discriminator cell (direct gateway HTTP)

| cell | status | body (verbatim) | meaning |
|---|---|---|---|
| model=claude-sonnet-5 on GPT gateway | 404 | `{"error":{"message":"Not Found","type":"not_found_error"},"type":"error"}` | a foreign model id is rejected at the catalog with a REASON-FUL 404 — **not** the masked "Bad Request" shape seen in live spawn failures |

## G. E2E cells — real `moai gpt` launcher + real Claude Code client (`claude --print`)

All post-fix. Fresh cells 1-4 establish the addendum's required green cells;
cells 5-8 cover session modes; cell 8 is a real subagent spawn.

| # | cell | mode | effort | result |
|---|---|---|---|---|
| 1 | luna | fresh | low | real reply: `ok` |
| 2 | astra | fresh | low | real reply: `ok` |
| 3 | terra | fresh | low | real reply: `ok` |
| 4 | sol | fresh | xhigh | real reply: `ok` |
| 5 | sol | continue existing conversation | low | real reply: `ok` (history-aware answer) |
| 6 | sol -> terra (same gpt-5.6 family) | continue + model switch | high | real reply: `ok` |
| 7 | gpt-5.6 -> astra (cross-family) | continue + model switch | low | real reply: `ok` |
| 8 | luna, Explore subagent spawn (inherit) | continue | low | real reply: `3` (subagent counted the words) |

## H. D6 spawn-surface cells (post-fix, D1 makes every rejection body visible)

| cell | result | reproduces lane-8 family body? |
|---|---|---|
| fresh sol + Explore spawn, model arg = opus | GREEN ("done") | no |
| continued sol session (with history) + Explore spawn, model arg = opus | GREEN ("done") | no |

Code-level note on the lead's hypothesis: a wire model of opus/sonnet can
never reach the family check — `catalog.Resolve` 404s first (section F, reason-
ful body). Lane-8's family-body failures therefore carried a GPT wire model
and failed the receipt/history check on REPLAYED PARENT HISTORY (fork/resume
of the parent transcript), not on the model id itself. The model arg may still
be the upstream-of-the-wire trigger (forcing a fork that replays parent
history); verifying that is t672-scope. See investigation.md §7.

Named-teammate spawn shape (t688 advisor case): NOT instrumented — the
team-mode surface is not drivable from a `-p` one-shot in this lane. The
masked-vs-real distinction between the two live mechanisms is still decisive;
see the verdict in the completion report.

Post-fix cell totals: 20 plain + 20 tool + 8 spawn-shaped + 1 discriminator
+ 8 e2e = 57 real green cells; 0 red cells on any (model x effort) combination
Claude Code can emit. Remaining non-green: the intentional 404 discriminator
(reason-ful, expected) and pre-fix historical cells in sections A-B.
