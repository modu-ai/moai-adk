# Context token measurement (rc.2 C8 #2184)

`context-bundle` reports a size in bytes and a token figure labelled
`{"method": "bytes_div_4", "model_scope": "none"}`. That figure is an estimate, never a
measurement (`src/context/mod.rs`). The byte budget is 256 KiB unless the request or the
profile key `context_max_bytes` says otherwise (user decision 01:1x(11-09),
`docs/decisions-2026-11-08.md`). The decision also says the real tokenizer comes **after a
per-host measurement**. This file is the frame for that measurement. No row below has
been measured yet.

## Error table (frame only)

The error is `(bytes_div_4 − measured) / measured`. A positive value means the estimate
over-counts.

| sample | bytes | `bytes_div_4` | Claude measured | Claude error | Codex measured | Codex error | measured at (UTC) | host build |
|---|---|---|---|---|---|---|---|---|
| S1 English prose (README section) | | | not_run | not_run | not_run | not_run | | |
| S2 Korean prose (decisions log rows) | | | not_run | not_run | not_run | not_run | | |
| S3 Rust source (`src/context/mod.rs`) | | | not_run | not_run | not_run | not_run | | |
| S4 JSON (a bundle output) | | | not_run | not_run | not_run | not_run | | |
| S5 mixed bundle at the 256 KiB default | | | not_run | not_run | not_run | not_run | | |

State: **not_run**. Measuring needs the operator's host sessions and accounts, which a lane
does not hold. Fill a cell only from an observed number. Leave the others `not_run`.

## Procedure

1. **Fix the sample.** Run `context-bundle` on the sample at a recorded commit. Record
   `bundle_digest`, `budget.used_bytes` and `tokens_estimate.value` from its output. The
   sample bytes are exactly the bundle's `content` fields, so the measurement and the estimate
   read the same bytes.
2. **Measure on each host**, with the host's own counter, never a third-party tokenizer
   standing in for it:
   - Claude: the input-token count the host reports for a turn that carries only the sample,
     minus the same turn carrying an empty sample, which removes the fixed overhead.
   - Codex: the same difference over the input-token usage Codex reports.
   Record the host name, its build/version and the UTC time (`date -u`) beside the number.
   Neither host's counting method was verified in this card. Check each host's current
   documentation before the first measurement and record the URL and the date you checked it.
3. **Compute the error** with the formula above and fill the row. Repeat each sample three
   times. When the counts differ, record the spread instead of picking one.
4. **Decide.** If every error stays within ±25 %, keep `bytes_div_4` as the default
   estimate. Otherwise open a card to add a per-host factor or tokenizer, citing this table.
   The threshold is this card's proposal and needs confirmation, not a recorded decision.

## Not covered

- No external backend is called by this card (decision 01:1x: no cost ceiling set means
  zero external calls).
- `tokens_estimate.model_scope` stays `none` until a measured method replaces the estimate.
