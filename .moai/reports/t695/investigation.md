# t695 Investigation — GPT model 400 sweep (effort allowlist + masked 400s)

Card: t695 · Branch: `WT-gpt-model-400-sweep` · Worktree: `.claude/worktrees/t695`
Date: 2026-09-13 · All probe commands run from the worktree against a gateway
child started from this worktree's own build (`bin/moai`, commit under test).

## 0. Isolation statement

- No live gateway, session, or process was killed, restarted, or reconfigured.
- Probes ran through `gateway.StartChild` (the production `internal-gateway`
  child entry) on an ephemeral `127.0.0.1:<random>` port with a session token
  generated per run. Probe tooling: `probe_driver_main.go.txt` (this directory).
- GPT credentials were reused **read-only through the gateway process** from the
  real `~/.moai/gateway-auth` store. The store refuses symlinked paths and
  copying token material is prohibited, so the probe children ran with the real
  `MOAI_HOME`. The probe payload deliberately carried **no Conversation** (or,
  where receipts were required, an isolated receipt store under
  `/private/tmp/t695/receipts-*`), so the children wrote **nothing** under
  `~/.moai/state/gateway-conversations/`. No other family's directory was read,
  modified, or deleted.
- E2E launcher cells ran in a scratch cwd (`/private/tmp/t695/e2e`) and created
  their own new conversation families; no existing family was touched.
- Probe processes were driven in the foreground and stopped via
  `ChildProcess.Stop` (bounded 3s) on every path; no background load was left.
- Upstream cost discipline: every cell used a tiny prompt ("Reply with the
  single word: ok") and `max_tokens <= 64`.

## 1. Cause chain (confirmed by real requests, not only code reading)

1. `internal/gateway/translate/native_policy.go` `nativePolicy`: the effort
   allowlist accepted only `high` (plus `medium` under `PolicyGPTNative`.
   Claude Code sends `output_config.effort` for every `/effort` selection, so
   low/xhigh/max failed locally with `errors.New("unsupported effort")`.
2. `internal/gateway/openai.go` `openAITranslationError`: every non-replay
   translate error was converted to a reason-less 400
   `{"message":"Bad Request","type":"invalid_request_error"}` — masking cause 1
   (and every other policy rejection) from client debug logs. The
   `HistoryReplayError` branch already passed its message through; D1 extends
   that treatment to all translate errors (they are the adapter's own
   validation strings).
3. Pre/post control pair (instrumented, D1 build, pre-D3 policy):

```
control-sol-low-prefix   400  3ms    {"error":{"message":"unsupported effort","type":"invalid_request_error"},"type":"error"}
control-luna-low-prefix  400  0ms    {"error":{"message":"unsupported effort","type":"invalid_request_error"},"type":"error"}
control-luna-high-prefix 200  2542ms {"content":[{"text":"ok",...}]}
control-sol-medium-prefix 200 2067ms {"content":[{"text":"ok",...}]}
control-sol-noeffort     200  1522ms {"content":[{"text":"ok",...}]}
```

The 400s returned in <=3 ms with no upstream round trip — local gateway
rejection, upstream never reached. This converts the earlier "consistent with
diagnosis" reading into a confirmed, attributed mechanism.

## 2. D2 upstream probe (real requests, per model x effort)

Method: temporarily widened the local policy in a THROWAWAY probe build
(reverted before the final commit) so every string effort reached the upstream
subscription endpoint; measured the upstream HTTP result through the real
adapter path (PKCE subscription endpoint, credentials resolved by the real
store/broker).

Results (200 = upstream accepted and returned a real reply):

| model \ effort | low | medium | high | xhigh | max |
|---|---|---|---|---|---|
| gpt-6-astra | 200 | 200 | 200 | 200 | **non-200 4/4** |
| gpt-5.6-sol | 200 | 200 | 200 | 200 | 200 (4/4 incl. stability resample) |
| gpt-5.6-terra | 200 | 200 | 200 | 200 | 200 (4/4 incl. stability resample) |
| gpt-5.6-luna | 200 | 200 | 200 | 200 | 200 (6/7; the single non-200 is explained below) |

- `astra + max` failed every time (4/4) while `astra + xhigh` succeeded —
  a per-model upstream rejection of `effort=max`, not a transient.
- The single `luna + max` non-200 is **not** an effort rejection: the identical
  request succeeded 6/7 times. Diagnostics on the later terra cells (temp
  logging in the throwaway build) showed the mechanism: responses that carry
  opaque reasoning items require receipt-backed history authorization; the
  receipt-less probe children rejected such responses locally as
  `opaque response authorization unavailable` (surfaced as a 502). Production
  children always carry the Conversation payload, so this probe artifact does
  not exist there — confirmed by the production-equivalent sweep below being
  20/20.

### D2 also probed the two other masked-rejection paths (required cells)

- `output_config.format` other than the exact title-only json_schema
  (non-title json_schema, sol + luna): upstream **accepted** (200, reply
  followed the foreign schema). Post-fix decision: pass through any
  `{"type":"json_schema","schema":<object>}` for the GPT profile; the
  response-side `{title}` output contract stays bound to the exact title-only
  schema.
- `thinking.display: "summarized"` (GPT): upstream **accepted** when mapped to
  `reasoning.summary: "auto"` (sol + luna, with adaptive + low effort). The
  probe also exposed a coupled rejection: `thinking adaptive` was locally
  restricted to high/medium effort; upstream accepts adaptive with every
  allowlisted effort (200 on sol/luna at low).

## 3. D3 mapping decisions (all evidence-backed, no guessing)

| client value | wire value | rationale |
|---|---|---|
| low, medium, high, xhigh (any model) | same | upstream 200 on all 4 models |
| max (sol/terra/luna) | max | upstream 200, stable on resampling |
| max (gpt-6-astra) | **xhigh** | upstream rejects max on astra 4/4; xhigh is the nearest accepted tier and is 200 on astra |
| display omitted | (no summary field) | unchanged |
| display summarized (GPT) | `reasoning.summary: "auto"` | upstream 200 on sol/luna |
| adaptive thinking (GPT) | valid with any allowlisted effort | upstream 200 at low effort |
| non-title json_schema format (GPT) | forwarded verbatim | upstream 200; response-side title contract unchanged |
| unknown effort (e.g. "ultra") | rejected `unsupported effort` | explicit allowlist preserved |

Anthropic-native profile: byte-identical to its previous validated subset (the
D2 evidence was collected against the GPT subscription upstream only; its
native body is forwarded verbatim).

## 4. The terra 502 cluster (instrumented, resolved)

During the first receipt-less tool sweep, terra failed 10/10 tool-bearing cells
while plain cells stayed green. Temp diagnostics in the throwaway build
captured at the 502 sites:

```
stage=subscription-convert err=opaque response authorization unavailable
```

Root cause: terra's responses include opaque reasoning items; without a
receipt-backed History authority the adapter correctly fails closed. Re-running
with an isolated receipt store (production-equivalent) turned all terra tool
cells 200. Not a product defect; recorded because it explains both the terra
cluster and the single luna+max non-200.

## 5. Verification (D5)

- `go vet ./internal/gateway/... ./internal/cli/` — clean.
- `golangci-lint run --new-from-rev=74d872aaf internal/gateway/...` —
  `0 issues.` (the 219 pre-existing issues in the package are unchanged).
- Affected packages: `go test ./internal/gateway/... ./internal/cli/...`.
  Only failure: `TestAppServerSubprocessHTTPToolContinuation`, reproduced at
  pristine HEAD `74d872aaf` (extracted via `git archive` to /tmp and run
  there) — pre-existing environmental (python subprocess spawn), unrelated to
  this card.
- RED evidence for D1: with `internal/gateway/openai.go` temporarily restored
  from HEAD, `TestOpenAITranslationErrorSurfacesReasonWithoutWrappingChain`
  failed with `{"message":"Bad Request",...}`; GREEN after the fix.

## 6. Explicit gaps (not observed)

- The upstream response STATUS/BODY behind the astra+max non-200 is not
  observable through the adapter (it maps non-200 upstream to a generic
  error); the mapping decision rests on the 4/4 vs 200-at-xhigh contrast.
- The exact model id carried by live spawned-agent requests (t672/t688
  surfaces) was not extractable: client debug logs record the 400s but not
  request bodies, and the production gateway keeps no request log. The D6
  verdict therefore separates what is code-proven from what is hypothesis.
- The named-teammate spawn shape was not drivable from this lane (interactive
  team-mode surface, not reachable from a `-p` one-shot); it is covered by
  operator/lead live evidence and the mechanism analysis, not by a matrix cell.
