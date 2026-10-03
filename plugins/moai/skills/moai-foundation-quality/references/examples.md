# TRUST 5 and Gate Examples

Worked examples of applying TRUST 5 and triaging gate/review output. These
are process walkthroughs, not library calls — MoAI enforces quality through
agents and gate commands, not a Python SDK.

## Example 1: Assessing a New Feature Change

A change adds an authentication endpoint. Walk through all five dimensions.

- **Tested**: Run the test command (e.g. `go test ./internal/auth/...`).
  Observe the output — are the new tests green? Is coverage at or above the
  threshold? Show the verbatim output, do not summarize.
- **Readable**: Are the handler and validator names intention-revealing
  (`validateCredentials`, not `vc`)? Is the token-refresh logic commented
  where non-obvious?
- **Unified**: Does it match the existing handler pattern in the file
  (same error-wrapping style, same response shape)? Run the formatter in
  check mode.
- **Secured**: Does it validate the credential input at the trust
  boundary? Does it avoid logging the raw credential? Consult the OWASP
  checklist for the authn surface (see moai-ref-owasp-checklist).
- **Trackable**: Does the commit read `feat(SPEC-{DOMAIN}-001): add login
  endpoint`? Is the SPEC referenced?

If any dimension is below bar, it is a gap to close — not a pass.

## Example 2: Triage of a `/moai gate` Run

`/moai gate` returns lint + format + type + test results in parallel.
Suppose it reports: 2 lint errors, 1 type error, tests green, format clean.

Triage:
1. **Type error** (major) — fix first; a type error often cascades into
   other findings. Fix directly or via `/moai fix`.
2. **Lint errors** (classify) — read each. If real defects, fix. If false
   positives, suppress per-line with a reason. If style preferences outside
   the rule set, defer or ignore.
3. **Re-run the gate** and show the output to verify the fixes landed.

Do NOT disable the linter rule globally to silence the 2 errors. Per-line
suppression with a stated reason is the sanctioned path.

## Example 3: Triage of a `/moai review` Run

`/moai review` flags a SQL query built by string concatenation (Secured
finding) and a 200-line function (Craft/Readable finding).

Triage:
1. **SQL injection** (critical, Secured) — parameterize the query. This is
   a must-fix before merge.
2. **200-line function** (major, Craft) — extract sub-functions. If the
   change is a hotfix and the function pre-exists, defer with an @MX:TODO
   and a stated reason rather than expanding scope.

Both findings map to TRUST 5 dimensions (Secured and Readable
respectively). The triage records which dimension each finding addresses.

## Example 4: Coverage Gap Assessment

The coverage report shows the touched package at 72%, below the 85%
threshold. The uncovered code is the new endpoint's error paths.

Triage:
1. The gap is a defect to close, not an accepted state.
2. Add tests for the error paths (missing-credential, expired-token,
   invalid-format). Run `go test -cover ./internal/auth/...` (or the
   project equivalent) and observe the new figure.
3. Do NOT claim "coverage is now X%" without re-running the command and
   observing the output in this run (verification-claim-integrity).

## Example 5: "Not Applicable" With Justification

A change updates a README only — no executable code. The assessment:

- **Tested**: "Not applicable — docs-only change, no executable code
  touched." (justified skip)
- **Readable**: Check the prose is clear and the links resolve.
- **Unified**: Check the markdown style matches other docs.
- **Secured**: "Not applicable — no input surface or trust boundary
  changed." (justified skip)
- **Trackable**: Commit reads `docs: update README install steps`.

The justified skips name why; unjustified skips are treated as gaps.

## Example 6: Technical Debt via @MX Tags

A review finds `@MX:TODO` and `@MX:DEBT` tags in the touched code.

Triage:
1. **@MX:TODO** — is the TODO in scope for this change? If yes, resolve it
   (mark GREEN). If no, leave it but confirm it has a stated reason.
2. **@MX:DEBT** with `@MX:CEILING` / `@MX:UPGRADE` — this is a deliberate
   working simplification. Confirm the ceiling and upgrade path are
   stated. Do not "fix" it unless the change crosses the stated ceiling.

Debt identified is not debt that must be fixed immediately. Record it,
prioritize it, schedule it. Acting on inferred debt without the domain's
verification tool is an unobserved defect claim.

## Example 7: sync-auditor Scoring (thorough harness)

At the thorough harness level, sync-auditor scores the change on four
dimensions as a fresh-judgment auditor. Suppose the scores are:
Functionality 0.9, Security 0.6, Craft 0.8, Consistency 0.85.

The overall score is the **harmonic mean**, not the average:
harmonic mean of (0.9, 0.6, 0.8, 0.85) leaves the low Security score
dragging the result down hard. This penalizes the lopsided quality
(functional but insecure).

sync-auditor's verdict: the Security dimension is below the must-pass bar.
The change does not PASS despite three strong dimensions. The fix: close
the security gap (input validation, credential handling) and re-run.

## Example 8: cycle_type and Quality

A SPEC is implemented via `cycle_type=tdd`. The quality shape:

- RED — a failing test is written first specifying the behavior.
- GREEN — minimal implementation makes the test pass.
- REFACTOR — the code is cleaned without changing behavior.

The Tested principle is front-loaded: the test exists before the
implementation. The Readable/Unified principles are addressed in REFACTOR.
Contrast with `cycle_type=ddd` (ANALYZE-PRESERVE-IMPROVE), where
characterization tests capture existing behavior before a
behavior-preserving transformation — here Tested is served by
characterization tests, not test-first.

See Skill("moai-workflow-tdd") and Skill("moai-workflow-ddd") for the full
cycle mechanics.
