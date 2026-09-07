Thank you for the report. The symptom you described — a background job failing
almost immediately while reporting a false "10m0s timeout" — was confirmed as **two
separate defects** producing one misleading message, and both are fixed on the
development line (fix commit `7da2f11d1`). The release carrying this fix has not
been cut yet; we'll follow up here as soon as it ships.

1. **The misreport.** The turn supervisor watched the derived context's Done channel
   and unconditionally named its own bound, so a turn ended by the *caller's*
   context was reported as having exhausted a bound it never approached. In our
   reproduction, a caller-cancelled turn that ended after ~50 ms was reported as
   "timed out after 10m0s" — about 1/11,800 of the bound. The fix distinguishes the
   cause from the parent context: a real bound expiry still reports a timeout, a
   caller-ended turn reports cancellation, each with its own message and next step.
2. **The immediate failure.** A background job inherited the request-scoped
   context, which the MCP host cancels the moment the handler returns — and for
   `background: true` that return is immediate, so every background job died within
   milliseconds of creation. The job now runs on `context.WithoutCancel`, keeping
   values (the progress token included) while dropping the cancellation chain; the
   turn remains bounded by the codex_task timeout. The `codex_job_cancel` path was
   verified unaffected.

Regression guards run in both directions: mutating the discriminant to
"always caller-ended" fails the timeout direction, so the other side cannot come
back silently. We'll follow up here as soon as a release with this fix ships.
