# SPEC-TCD-LLM-DECIDER-001 — Implementation Plan

Tier M · card t1352 · plan-phase authored 2026-10-03 at HEAD `681ee30b5` (worktree
`.claude/worktrees/t1352`, branch `WT-llm-card-decider`, == local develop tip). Milestones are
ordered dependency-correct with the riskiest existing-path change (M3, lock scope) isolated from
the purely additive ones; the decision records most likely to change (§C OD-A..OD-D) lead the
doc. Priority labels only — no time estimates.

## §A Context

SPEC-TODO-CLASSIFY-DISPATCH-001 shipped the classification seam with `DefaultCardDecider` as
the shipped backend and deferred the `Decider(llm)` implementation (its amendment 0.3.1, leader
ruling 2026-09-29 OD-2; operator standing-delegation provenance, Jev ask noul 0.8, gate 0.50).
Card t1352 (operator directive 2026-10-03) delivers that deferred half: an LLM-based
`CardDecider` behind the seam, so identity `llm` has a shipped implementation. The card is Low
priority by the queue's own framing — the plan keeps the surface minimal and default-off.

## §B Known facts (measured in this tree at `681ee30b5`)

1. **The seam** — `var todoCardDecider kanban.CardDecider = kanban.DefaultCardDecider{}`
   (`internal/cli/todo_classify.go:31`); tests already replace it, so the selection shape has a
   precedent.
2. **The entry points** — `newTodoAddCmd` RunE resolves `dec := todoCardDecider` and swaps in
   `todoDeciderFromClassificationFile` for `--classification-file` (`todo.go:752-759`);
   `runTodoAddAppendRoot` defaults a nil decider to the package default (`todo.go:787-789`) and
   is shared with the MCP `todo_add` tool (REQ-SD-024).
3. **The locked write** — the classification is resolved inside `todoStoreAt(root).Mutate(...)`
   via `todoClassifyInLock` (`todo.go:809-823`, `todo_classify.go:76-84`); the fallback branch
   (error → `DefaultCardClassification()` + exactly one notice,
   `todoClassificationFallbackNotice`, `todo_classify.go:36`) is already single-sourced.
4. **The closed sets and validator** — `classification.go:33-69` (identity `llm` accepted,
   `jev`/`llm+jev` named refusals at `:117-119`), `ValidateCardClassification`
   (`:109-123`), `StaticCardDecider` (`:266`), `UnavailableCardDecider` (`:279`),
   `DefaultCardClassification` (`:98-105`).
5. **The GLM transport precedent** — `internal/cli/mcp_glm.go` calls the z.ai Anthropic-compatible
   endpoint DIRECTLY: `config.DefaultGLMBaseURL + glmMessagesPath` (`:314`;
   `DefaultGLMBaseURL = "https://api.z.ai/api/anthropic"`, `defaults.go:272`), credential via
   `loadGLMKey → ~/.moai/.env.glm` (`:93-95`), a `glmHTTPDoer` test seam (`:84-91`), a bounded
   HTTP timeout (`:78-82`), and `extractJSONObject` (`:399`) for fenced-response parsing. The
   same package — every one of these is directly reusable.
6. **Constants homes** — `internal/config/envkeys.go` (no `MOAI_TODO_*` variable exists yet;
   env-over-code precedents: `EnvSessionWorktree`, `EnvFactoryAutoDispatch`; loud-refusal
   precedent: `EnvClaudeBin` — "an invalid pin fails the launch rather than silently falling
   back") and `internal/config/defaults.go` (`Default*` const block).
7. **The refusal shape** — `exitCodeError{code: 2, ...}` as used by
   `todoDeciderFromClassificationFile` (`todo_classify.go:64-67`).
8. **Methodology** — `quality.yaml` `constitution.development_mode: tdd` → run phase is
   RED-GREEN-REFACTOR.

## §C Decision records (OD-A..OD-D — resolved; two carry a leader-escalation note)

- **OD-A — activation trigger: env var `MOAI_TODO_DECIDER`, default-off (RESOLVED).** Accepted
  set exactly `{default, llm}`; unset/empty/`default` → `DefaultCardDecider`; `llm` → the LLM
  decider; any other value → exit-2 usage refusal naming the accepted set (loud, per the
  `EnvClaudeBin` precedent — an operator-authored misconfiguration must not silently become
  `default`); `jev`/`llm+jev` refused by the parent SPEC's named wording. Rejected
  alternatives: (1) a new YAML config key — heavier (loader struct + template + local config)
  and against the card's dormant framing; (2) a CLI flag — cannot express "standing judgment
  backend", only per-invocation; (3) default-on — a default-on LLM call at every card add is
  the one shape the card's Low-priority note rules out. Leader-escalation note: none needed —
  the default-off choice is the conservative direction; flipping it later is an env change, not
  a schema change.
- **OD-B — judgment source and transport: in-process HTTPS to the configured GLM endpoint via
  `internal/glmcred` (RESOLVED).** The decider mirrors the `mcp_glm.go` transport pattern —
  same base URL constant, same `/v1/messages` path, same credential reader, its own injectable
  doer seam, its own (shorter) timeout. NOT `StaticCardDecider` — that is the shipped
  judgment-file transport, not what the card asks for. NOT `internal/jevcred`/TypeSafe — C4/C1
  route it out. This amends the parent's "product computes no LLM call" header invariant for
  exactly this gated surface — recorded as constraint C2 of spec.md with rationale.
  Leader-escalation note: the provider/credential choice (GLM via glmcred) is the decision a
  leader may re-point; the REQs are provider-shaped only through C2/REQ-TLD-001's wording, so a
  re-point is a plan edit, not a REQ rewrite.
- **OD-C — failure taxonomy: every LLM-side failure takes the transport branch (RESOLVED).**
  Absent credentials, connection failure, non-2xx, timeout, non-JSON, and out-of-set fields ALL
  resolve to: fail-safe default + exactly ONE `todoClassificationFallbackNotice` line. A
  response whose `decider` field claims another identity is NOT a failure — the identity is
  forced to `llm` before validation and the judgment is accepted (REQ-TLD-004, AC-TLD-004);
  claimed-identity handling sits outside this taxonomy. The exit-2 branch stays exclusive to
  what the
  OPERATOR authored: `--classification-file` content (parent REQ-TCD-004, unchanged) and the
  `MOAI_TODO_DECIDER` selector value itself. Principle: operator-authored input is validated
  loudly; model-produced output degrades quietly-but-noticed. An add must never fail because a
  remote model misbehaved (parent REQ-TCD-003 semantics); an add must never silently run with a
  misconfigured selector the operator believes is `llm`.
- **OD-D — latency and lock scope: judgment computed outside `Mutate`, attached inside
  (RESOLVED).** Today the in-lock classify is microseconds (deterministic deciders). An LLM
  round-trip is seconds, with a 10-second ceiling. Inside the lock, N concurrent adds serialize
  into up to N×timeout and every queue reader (`todo list`, `factory next`) blocks behind them;
  outside, the calls overlap and only the write serializes. Mechanically: when the selected
  decider is the LLM one, the entry point pre-classifies
  (`cls, err := dec.Classify(text)` before `Mutate`), on error prints the one notice and takes
  `DefaultCardClassification()`, then hands `kanban.StaticCardDecider{Class: cls}` into the
  lock — `todoClassifyInLock` stays untouched (it re-stamps `ClassifiedAt` at the write,
  preserving the stamp-at-write discipline single-sourced). The parent's visibility invariant
  holds: visibility is created by the locked write, not by the computation.

## §D Constraints, Dependencies, Risks

### §D.1 Dependencies

- **depends_on: SPEC-TODO-CLASSIFY-DISPATCH-001** (parent/ prior SPEC) — seam, closed sets,
  fail-safe semantics, fallback notice. Frontmatter reads `status: completed` at this tree →
  run-phase depends_on pre-flight passes. If the t1407 amendment line re-lands it as
  `in-progress` before run entry, the pre-flight flags it — resolve at the gate then (the
  amendment's own record says the amendment completed).
- The parent SPEC's artifacts are read-only for this SPEC.

### §D.2 Related surfaces (consume, never modify)

- `internal/kanban/classification.go` — interface, closed sets, validator, `StaticCardDecider`.
- `internal/cli/mcp_glm.go` — transport pattern constants (`glmMessagesPath` is package-private;
  referencing it from the new file keeps ONE spelling in the package; if plan-audit objects to
  the coupling, mirror the value with an equality-pin test instead — run-phase discretion).
- `internal/config/envkeys.go`, `internal/config/defaults.go` — new constants only.

### §D.3 Constraints (from spec.md §C, binding on run)

- C1 jev boundary · C2 recorded invariant amendment · C3 single validator · C4 constants
  discipline · C5 no new config-file key · C6 scope discipline. All tests: `t.TempDir()`, no
  real network (every LLM test injects the doer), no OTEL `t.Setenv` in parallel tests;
  `MOAI_TODO_DECIDER` tests use `t.Setenv` and stay non-parallel.

### §D.4 PRESERVE list (untouched paths)

- `internal/kanban/**` — no Go change (the decider lives in `internal/cli`).
- `internal/cli/todo_classify.go` — ONE comment-only amendment (M2, the D2 carve-out): the
  header invariant "the product computes no LLM call and knows no external tooling"
  (`todo_classify.go:12-15`) becomes false in this package the moment the LLM decider lands,
  so it is reworded to the OD-B-amended statement (constraint C2) — the sole sanctioned LLM
  computation is this SPEC's gated decider, and `--classification-file` remains the sole
  operator-supplied judgment injection path. The fallback notice constant (`:36`) and the
  in-lock helper (`:76-84`) stay byte-unchanged; the lock-scope restructure lives in the entry
  points, not here.
- `internal/cli/factory_card.go` and every lease/claim path — explicitly out (see §D.5).
- `internal/template/templates/**` — no template change (C5).
- The parent SPEC directory `.moai/specs/SPEC-TODO-CLASSIFY-DISPATCH-001/`.

### §D.5 Risks

- **Concurrent lane t1458 (atomic serial lease).** t1458 works
  `internal/kanban/factory_card.go` and the lease/claim machinery. This plan's file list does
  NOT include `factory_card.go`, `internal/kanban/**`, or any lease path — the package-level
  overlap is limited to `internal/cli` (both lanes compile it; the files are disjoint). The one
  shared FILE risk is `internal/cli/todo.go`: this plan edits only the add-path entry points
  (`newTodoAddCmd` RunE, `runTodoAddAppendRoot`'s nil branch); if t1458's scope ever expands
  into the add path, the two landings serialize through the integration window and the later
  lane re-measures the merged tree. No t1458 SPEC artifact is touched.
- **The fallback notice is line-counted.** `todoClassificationFallbackNotice`
  (`todo_classify.go:36`) is exactly one line by design (operators and tests count fallbacks by
  counting lines) — never reword it.
- **Timeout budget.** The audit client's 120s ceiling (`mcp_glm.go:78`) is the wrong budget for
  an interactive add; the decider gets its own constant (default 10s, `defaults.go`) — do not
  reuse the audit value.
- **Network-in-suite hazard.** Zero real network in tests: the doer seam makes every scenario
  an `httptest.Server` or a stub; a test needing a real endpoint is a defect.
- **`t.Setenv` discipline.** `MOAI_TODO_DECIDER` tests mutate process env — keep them
  non-parallel (`t.Setenv` already enforces this); never set OTEL vars (AGENTS.local.md §6).

## §E Self-Verification (run phase reports each item with command + verbatim output + HEAD)

- E1. AC PASS/FAIL matrix — AC-TLD-001..007, each with its named test's actual output.
- E2. Cross-platform: `go build ./...` and `GOOS=windows GOARCH=amd64 go build ./...` → exit 0.
- E3. Coverage: `go test -cover ./internal/cli/... ./internal/config/...` ≥ 85% on the touched
  packages.
- E4. Boundary greps: `scripts/jev` references in `internal/ pkg/ cmd/
  internal/template/templates/` = 0 (positive control: local-only docs carry hits); no inline
  `os.Getenv("MOAI_TODO_DECIDER")` outside `envkeys.go`; no endpoint literal outside
  `defaults.go`.
- E5. Lint: `golangci-lint run` (CI version) — NEW issues only reported vs baseline.
- E6. RED evidence: per TDD, verbatim pre-GREEN failing output for each RED-now AC (E8 of the
  delegation template).
- E7. Concurrency: `go test -race ./internal/cli/...` for the lock-scope tests (AC-TLD-005).

## §F Milestones

- **M1 — Activation surface + selector (additive; flips nothing).** `envkeys.go`:
  `EnvTodoDecider = "MOAI_TODO_DECIDER"` (documented constant). `defaults.go`:
  `DefaultTodoClassifyLLMTimeout = 10 * time.Second`. New file
  `internal/cli/todo_decider_select.go`: the standing-decider selector (env read once per add
  invocation; accepted set `{default, llm}`; invalid → `exitCodeError{2}` naming the accepted
  set and the jev refusals). Entry-point wiring in `todo.go` (`newTodoAddCmd` RunE +
  `runTodoAddAppendRoot` nil branch): `--classification-file` first (wins, LLM not invoked),
  else the selector. Tests: selector matrix, refusal messages, precedence, and the default-path
  regression guard (AC-TLD-002, AC-TLD-007). RED first: selector tests fail on `681ee30b5`
  (no env reading exists).
- **M2 — The LLM decider implementation (new file `internal/cli/todo_classify_llm.go`).**
  `llmCardDecider` implementing `kanban.CardDecider`: request build (system prompt constraining
  the response to the closed sets + JSON-only; card text as the user message), injectable doer
  + key-loader seams (the `mcp_glm.go` pattern), endpoint from config constants, timeout from
  the M1 constant, response parse (`extractJSONObject`), validate via
  `kanban.ValidateCardClassification` with the identity FORCED to `llm` before validation,
  reason bounded to one line. Model: the flash/task slot — the existing package constant
  `glmTaskDefaultModel` (`config.DefaultGLMHigh`, `mcp_glm.go:58-63`), never the audit pin
  slot; `max_tokens` capped at 512 (a three-field + one-line-reason JSON response); no
  reasoning_effort (an empty value omits the field, `mcp_glm.go:113`). Every failure returns
  an error — the fallback branch stays in the seam (OD-C), never synthesized inside the
  decider. This milestone also lands the comment-only header amendment of
  `todo_classify.go:12-15` recorded in §D.4 (constraint C2 — behavior unchanged, notice text
  untouched). Tests via `httptest`: happy path,
  non-JSON, out-of-set, claimed identity `human`/`jev`, 500, unreachable, slow-past-timeout
  (AC-TLD-001, AC-TLD-003, AC-TLD-004, AC-TLD-006).
- **M3 — Lock-scope restructure (the one existing-path change).** In both add entry points:
  when the selected decider is the LLM one, pre-classify before `Mutate`; on error print the
  ONE notice + fail-safe default; hand `StaticCardDecider{Class: cls}` into the lock
  (OD-D). Tests: concurrent-add timing (both judgments outside the lock — wall-clock bound),
  reader-not-blocked assertion, notice-count, `-race` (AC-TLD-005, AC-TLD-003's notice clause).
- **M4 — Guards, regression, re-measure.** Jev-boundary greps with positive control; secret
  marker test (fake credential through the key-loader seam; zero marker bytes in all process
  output on happy AND failure paths, with the positive control that the fake server DID receive
  it); default-path regression mutation check; re-measure owning packages
  (`go vet`, `go test -timeout 30m ./internal/cli/... ./internal/config/...`, coverage,
  `golangci-lint`, `moai spec lint`); evidence to `.moai/reports/t1352/`.

## §G Anti-Patterns (run-phase refusals)

- Do NOT widen the accepted selector set beyond `{default, llm}` (`human` has no product
  implementation — accepting it would record an identity nothing produced).
- Do NOT re-spell the closed value sets in the new file (C3) or reword the fallback notice.
- Do NOT add a YAML config key, template file, or provider abstraction (C5, Out of Scope).
- Do NOT touch `internal/kanban/**`, `factory_card.go`, lease paths, or the parent SPEC
  artifacts (§D.4).
- Do NOT let a decider error refuse the add (OD-C) or let a slow decider run inside `Mutate`
  (OD-D).
- Do NOT log the credential: no key material in errors, notices, or test output.

## §H Cross-references

- spec.md `§C` (constraints C1-C6), `acceptance.md` (AC-TLD-001..007 + adoption table).
- Parent SPEC artifacts: `.moai/specs/SPEC-TODO-CLASSIFY-DISPATCH-001/{spec,plan,acceptance}.md`.
- Transport precedent: `internal/cli/mcp_glm.go`; constants homes: `internal/config/envkeys.go`,
  `internal/config/defaults.go`.
