# Plan — SPEC-FEEDBACK-ANON-PARTICIPATION-001

## §A Context

- Card t1498 (operator scope as corrected 2026-10-04): opt-in, real-name participation; automatic issue filing from the user's own `gh` account only for moai-adk-attributed tool errors; zero-token deterministic path with a model called only when a new issue is about to be published.
- Standalone SPEC, Tier L (new packages, more than 15 files touched across `internal/bugreport`, `internal/feedback`, `internal/cli`, `internal/config`, `internal/settings`, `internal/web`, templates, three skill-body copies, four docs pages). Frontmatter `depends_on: SPEC-FEEDBACK-AUTO-SUBMIT-001`.
- Branch and worktree: the card worktree branches from `develop`; the commit trail carries the card id.
- Methodology: TDD per `quality.yaml`; every named test in `acceptance.md` is written first (RED) and its failure observed before the milestone is called done.

## §B Open clarifications (Kickoff inputs)

Only the first marker touches an operator-held choice, and it gates only the policy flip in M6. Every deterministic milestone (M1 to M5, M7) proceeds without any answer here.

1. [NEEDS CLARIFICATION: ambiguous attribution — does an ambiguous verdict call the model, or stay local and unsent?] The card says deterministic rules come first and a model is used only when attribution is ambiguous; the leader's message says the model is called only at the moment an issue is published. Both variants are specified and tested (REQ-ANON-009, AC-ANON-011, AC-ANON-020); the shipped constant is `local` until answered. Context for the choice: the kinds that default to ambiguous are `hook_timeout` and two template sentinels (`ErrUnexpandedToken`, `ErrInvalidJSON`); the model sees only fixed fields, so for `hook_timeout` it has no more information than the rules do.
2. [NEEDS CLARIFICATION: how Go code reaches a model for the publish summary] No text-generation call exists in the Go tree (only the `claude` availability probe in `internal/cli/doctor.go`). Candidates are a headless invocation of the user's own `claude` binary with a fixed prompt on stdin, or no model at all (template text only). Non-blocking: the deterministic template fallback already files the issue (REQ-ANON-017), so M1 to M5 and M7 do not depend on the answer.
3. [NEEDS CLARIFICATION: flush trigger set] Flush runs at `moai feedback participation flush` and the end of `moai update`. Without a more frequent trigger, queued reports leave only when the user runs update. A session-start trigger would need a detached, time-boxed process and must never run inside a hook dispatch. Proposed default: the two triggers above.
4. [NEEDS CLARIFICATION: cap values] The proposed constants in `design.md` section 10 (7-day per-fingerprint window, 3 per day, 10 per week, queue bound 20, attempt limit 5, daily model-call cap 6) are proposals. Proposed default: ship them as written; they are constants, so changing one is a one-line change.
7. [NEEDS CLARIFICATION: withdrawal purge scope] Default specified: turning participation off makes the next flush discard unsent items and the spool (the sent-history log stays); the purge command removes everything. The alternative is to leave unsent items inert until purge.
8. [NEEDS CLARIFICATION: fingerprint merge granularity] The card defines the fingerprint to include version, commit, and operating system, so the same defect on two commits or two systems yields separate issues. Default: follow the card. The alternative is a coarser merge key with the finer fields as occurrence attributes.
9. [NEEDS CLARIFICATION: runtime defects of deployed agents and skills] The card lists "deployed harness defect" as a signal; the tree has a machine-detectable signal only for hooks and for shipped templates that fail to render or validate. A behavioural defect of an agent or skill has none, and inferring one from Claude Code tool-failure stubs is excluded (REQ-ANON-007). Default: cover only the render and validate sentinels in this SPEC.

(The numbering skips 5 and 6 on purpose: two earlier clarifications were closed by the operator's scope correction and are recorded in research.md section 6.)

## §C Milestones (ordered by decision reversibility: types and user-facing flows first, mechanical work last)

Priority High for M1 to M5; Medium for M6; Medium for M7. Complete each before starting the next; M2 and M3 may proceed in parallel worktrees only if their file sets stay disjoint, which they do except `internal/config/defaults.go` (integration serialised).

### M1 — Taxonomy, payload, fingerprint (new types, High)
Creates the new interfaces everything else binds to; most expensive to change later.
Files: `internal/bugreport/kind.go`, `verdict.go`, `payload.go`, `validate.go`, `fingerprint.go`, `frames.go`, and `_test.go` siblings; constants in `internal/config/defaults.go`.
Proves: AC-ANON-012, AC-ANON-013.

### M2 — Consent and setting surfaces (user-facing flow, High)
Files: `internal/template/templates/.moai/config/sections/feedback.yaml` and the local mirror `.moai/config/sections/feedback.yaml`; `internal/config/types.go`, `defaults.go`, `feedback_accessors.go`, `loader.go`; `internal/config/testdata/shipped_key_inventory.yaml`; `internal/settings/testdata/sections/feedback.yaml`, `schema_sections.go`, `schema_sections_test.go`; `internal/web/assets/i18n.js`, `settings_shell.go`, `feedback_panel_test.go`; `internal/cli/wizard/questions.go`, `translations.go`, `types.go`, `wizard.go`; the pin tests `question_removal_test.go`, `init_regroup_test.go`, `expansion_test.go`, the axis golden under `internal/cli/wizard/testdata/axis/`, and a new `participation_question_test.go`; `internal/cli/init_participation_wizard.go` (new), `init.go`; `internal/cli/update_participation.go` (new), `update.go`.
Proves: AC-ANON-001, 003, 004, 005, 006, 022.
Amendment: the commit message records that this re-adds an init question and amends the quiet-wizard decision (REQ-ANON-003).
Release ordering: this milestone must not ship in a release before M5, otherwise users are asked to consent to a sender that does not exist (section F).

### M3 — Detection and attribution wiring (High)
Files: `internal/bugreport/capture.go`, `spool.go`, `attribution.go`, `register.go`; `cmd/moai/main.go`; one capture call in each recover site listed in `design.md` section 2 (`internal/guardliveness/evaluator.go`, `internal/cli/codex_stop_chain.go`, `internal/resilience/circuit.go`, `internal/navigator/route/run.go`, `internal/navigator/fix/request.go`, `internal/escalation/detector.go`, `internal/hook/session_start_guard_liveness.go`, `internal/hook/user_decision_capture.go`, `internal/hook/session_start_binary_lag.go`, `internal/hook/navigator_detect.go`); `internal/hook/registry.go`; `internal/cli/preference/cmd.go`; `internal/cli/update_template_sync.go`; `internal/cli/update_clean_install.go`; guard tests including the recover-site walk.
Proves: AC-ANON-002, 007, 008, 009, 010.

### M4 — Local pipeline, preview, log, withdrawal (High)
Files: `internal/feedback/outbox/*.go` (drain, ledger, caps, tripwire, queue wiring, log); `internal/feedback/queue.go` (optional fingerprint fields, omitempty, backward compatible); `internal/cli/feedback_participation.go` (preview, purge, flush); tests.
Proves: AC-ANON-014, 015, 016, 023.

### M5 — Publication through the user's gh, deterministic text (High)
Files: `internal/feedback/publish/ghrunner.go`, `sender.go`, `lookup.go`, `contract.go`, `template.go`, `internal/feedback/testdata/bugreport_issue_v1.golden`; the flush call at the end of `update.go`; tests with a `gh` stub.
Proves: AC-ANON-017, 018, 021. After M5 the whole feature works with zero model calls.

### M6 — Model step and ambiguous policy (Medium)
Files: `internal/feedback/publish/model.go`, `adjudicate.go`, `budget.go`; tests with a counting stub. The channel and the policy value come from clarifications 1 and 2; until answered, the seam is wired to the template fallback and the policy is `local`.
Proves: AC-ANON-011, 019, 020.

### M7 — Mirrors, documentation, shipping guard (mechanical, Medium)
Files: `internal/template/templates/.claude/skills/moai/workflows/feedback.md`, `plugins/moai/skills/moai/workflows/feedback.md`, `.claude/skills/moai/workflows/feedback.md`; `docs-site/content/{en,ko,ja,zh}/utility-commands/moai-feedback.md`; `internal/template/auto_repair_guard_test.go` (new); `make build`.
Proves: AC-ANON-024, 025.

## §D MX tag plan

- `@MX:ANCHOR` on `bugreport.Capture` (called from the main recover, ten recover sites, hook registry, update paths: fan-in well above three) and on `bugreport.Fingerprint` (every dedupe, cap, title, and comment keys on it).
- `@MX:WARN` with `@MX:REASON` on: the capture hot path (time-boxed, fail-open, must stay network-free); the `gh` runner (public side effect from the user's account); the model seam (token spend, daily cap); the scrub tripwire stage (withholding, not masking).
- `@MX:NOTE` on each recover allowlist entry (the reason is the note) and on the frame filter (why file paths are never read).
- `@MX:DEBT` with `@MX:CEILING` and `@MX:UPGRADE` on the interim `AmbiguousPolicy=local` constant, naming clarification 1 as the upgrade trigger.

## §E Risks

| Risk | Vector | Mitigation |
|---|---|---|
| Stack frames leak user data | frame file paths embed the builder's absolute paths because no `-trimpath` is set; closure or method names could in principle carry data | frames use `Function` only, never file, line, or args; module-prefix filter; allowlist regex; canary test (AC-ANON-012) |
| Panic values or error text leak | `os.PathError` carries a full path; `internal/core/project/root.go:114` panics with a message that embeds an error | the builder takes no error, panic value, or string; attribution uses `errors.Is` and `errors.As` only; type-level signature test and canary (AC-ANON-013) |
| Handler identity leaks | `%T` of a handler | only the type name is used, validated by an allowlist regex; event must be a registered constant |
| Real-name exposure | the issue and comment are public, tied to the user's account, and their creation time is public | consent text states it in four locales (AC-ANON-006); caps keep volume low; no client timestamp is sent |
| False attribution to moai | an environment or user cause that no rule recognises | ordered rules put environment and user first; unmatched becomes `ambiguous`, never `moai`; table test per row |
| Hook timeouts from machine load | a slow host looks like a bug | `hook_timeout` defaults to `ambiguous`; per-fingerprint and global caps |
| Vulnerability disclosed publicly by a crash report | a path-traversal sentinel or a security-flavoured function name | `ErrPathTraversal` reports are always withheld; classifier tripwire; a function name that trips the classifier blocks its own report (fails safe) |
| Duplicate issues from concurrent first filers | search index latency (unverified) | accepted; the consumer merges by fingerprint; comments are append-only so counts do not lose updates |
| Count drift | deleted comments, repeat users | accepted as approximate; documented in `design.md` section 7 |
| Surprise prompt at update | a user who toggled off in the console before being asked is asked once | documented edge (`design.md` section 9) |
| Token spend | model called more than once per issue | single call per publish, daily cap, counting-stub budget table (AC-ANON-020) |
| Flush slows `moai update` | network work at the end of update | 10 s time box; failure never fails the command |
| Pin-test churn | adding a question moves a stepper denominator and a golden | tests named in REQ-ANON-003 and updated in the same commit |
| Test env coupling | tests that read `CI` | every such test sets the variable with `t.Setenv` |

## §F Release ordering

No release may carry M2 (the consent question) without M5 (the sender). Verification at release time: the release checklist confirms `runParticipationStep` and the sender are both present on the release branch. Rationale: asking for consent to a feature that cannot act would record a consent the user cannot exercise.

## §G Anti-patterns to avoid

- Building a transport abstraction for anonymity; the transport is `gh`, full stop.
- Reading `err.Error()` anywhere in the capture or payload path.
- Calling the model before the duplicate lookup.
- Editing an issue body to count occurrences.
- Adding an init question to `DefaultQuestions` (it would reach the reconfigure path and break its pin).

## §H Cross-references

`spec.md`, `acceptance.md`, `design.md`, `research.md`; `.claude/rules/moai/development/spec-frontmatter-schema.md`; `.claude/rules/moai/development/verification-completeness.md` (RED-now ledger form).
