# Subscription context probe evidence

## Claim

Sol accepted 921825 observed input tokens twice; Astra accepted 921821 twice. Both rejected an estimated 921883 input tokens. Boundary widths: 58 and 62 tokens. All successful requests recalled three needles at start/middle/end. This does not prove a 1050000 total context window or server output reservation.

## Evidence

Existing MoAI ResolveFresh and SendAuthorized, subscription endpoint only. No credential/header printing. Request bound16MiB, timeout150s, response bound2MiB. stream=true/store=false/reasoning=medium; max_output_tokens omitted. Repeated ASCII unit " x". Server usage on successes consistently equals repetitions+71; failures have no usage, so failed counts remain estimates.

| Model | Repetitions | Observed input | Estimated rejected input | Result | Needles | Seconds | Artifact |
|---|---:|---:|---:|---|---|---:|---|
| gpt-6-astra | 1000000 | - | 1000071 | EXPECTED_REJECTION | False | 2.77 | [astra-1000000.json](astra-1000000.json) |
| gpt-6-astra | 800000 | 800071 | - | PASS | True | 27.95 | [astra-800000.json](astra-800000.json) |
| gpt-6-astra | 921000 | 921071 | - | PASS | True | 33.98 | [astra-921000.json](astra-921000.json) |
| gpt-6-astra | 921500 | 921571 | - | PASS | True | 33.49 | [astra-921500.json](astra-921500.json) |
| gpt-6-astra | 921750 | 921821 | - | PASS | True | 32.00 | [astra-921750-repeat.json](astra-921750-repeat.json) |
| gpt-6-astra | 921750 | 921821 | - | PASS | True | 35.10 | [astra-921750.json](astra-921750.json) |
| gpt-6-astra | 921812 | - | 921883 | EXPECTED_REJECTION | False | 2.88 | [astra-921812.json](astra-921812.json) |
| gpt-6-astra | 921875 | - | 921946 | EXPECTED_REJECTION | False | 1.88 | [astra-921875.json](astra-921875.json) |
| gpt-6-astra | 922000 | - | 922071 | EXPECTED_REJECTION | False | 2.30 | [astra-922000.json](astra-922000.json) |
| gpt-5.6-sol | 1000 | 1071 | - | PASS | True | 2.23 | [sol-1000.json](sol-1000.json) |
| gpt-5.6-sol | 1000000 | - | 1000071 | EXPECTED_REJECTION | False | 3.16 | [sol-1000000.json](sol-1000000.json) |
| gpt-5.6-sol | 400000 | 400071 | - | PASS | True | 13.58 | [sol-400000.json](sol-400000.json) |
| gpt-5.6-sol | 800000 | 800071 | - | PASS | True | 18.91 | [sol-800000.json](sol-800000.json) |
| gpt-5.6-sol | 872000 | 872071 | - | PASS | True | 27.51 | [sol-872000.json](sol-872000.json) |
| gpt-5.6-sol | 921000 | 921071 | - | PASS | True | 34.48 | [sol-921000.json](sol-921000.json) |
| gpt-5.6-sol | 921464 | 921535 | - | PASS | True | 21.72 | [sol-921464.json](sol-921464.json) |
| gpt-5.6-sol | 921696 | 921767 | - | PASS | True | 22.92 | [sol-921696.json](sol-921696.json) |
| gpt-5.6-sol | 921754 | 921825 | - | PASS | True | 30.39 | [sol-921754-repeat.json](sol-921754-repeat.json) |
| gpt-5.6-sol | 921754 | 921825 | - | PASS | True | 22.33 | [sol-921754.json](sol-921754.json) |
| gpt-5.6-sol | 921812 | - | 921883 | EXPECTED_REJECTION | False | 4.17 | [sol-921812.json](sol-921812.json) |
| gpt-5.6-sol | 921928 | - | 921999 | EXPECTED_REJECTION | False | 2.35 | [sol-921928.json](sol-921928.json) |
| gpt-5.6-sol | 921929 | - | 922000 | EXPECTED_REJECTION | False | 2.60 | [sol-921929.json](sol-921929.json) |
| gpt-5.6-sol | 922000 | - | 922071 | EXPECTED_REJECTION | False | 7.52 | [sol-922000.json](sol-922000.json) |
| gpt-5.6-sol | 950000 | - | 950071 | EXPECTED_REJECTION | False | 2.12 | [sol-950000.json](sol-950000.json) |

Exact commands in summary.json journeys[].command; bounded complete logs in .runs/. Go test PASS means harness completion, not inference acceptance. Every context rejection was HTTP200 with SSE response.failed.

```text
server_error_code: context_length_exceeded
server_error_type: invalid_request_error
server_error_param: input
server_error_message: Your input exceeds the context window of this model. Please adjust your input and try again.
```

## Baseline-attribution

Worktree /Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified; branchWT-unified-gateway; HEAD81c1d58f9 plus existing dirty tree. go1.26.8 darwin/arm64; Python3.9.6. Only overlay evidence files authored; application source and installed binary unchanged.

## Gaps

Rejected-token counts estimated. Synthetic ASCII/needle recall does not establish real project semantic quality or complete retention. No long output, tool/multimodal history or production Claude UI probe. Local272k/1MiB limits bypassed for server measurement.

## Residual-risk

Server policy/account/model rollout may change limit. Output reservation and hidden overhead unknown. Do not equate observed input maximum with total context. Product limits need estimator and tool/system overhead margin.
