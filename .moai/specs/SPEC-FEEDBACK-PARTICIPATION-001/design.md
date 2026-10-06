# Design — SPEC-FEEDBACK-PARTICIPATION-001

Design notes for the run phase (version 0.4.0). The spec states what must hold; this file states the shape that makes it hold. Citations were re-read in this tree (commit `bb54f2903`, code tree identical to `2f492df19`) unless marked unverified. Line numbers are leads for the run phase, which re-reads before editing.

## 1. Package layout and the import graph

Three new packages, one per concern, so the capture path stays tiny and the network path stays isolated. The user-scoped consent reader is a file in the existing `internal/config` leaf, not a package.

| Package | Holds | Imports (allowed) |
|---|---|---|
| `internal/bugreport` | kind enum, verdict enum, attribution rules, frame filter, fingerprint, payload type and closed-set validators, capture entry point, spool, internal-marker wrapper | standard library except `os/exec` and `net/http`, plus `internal/config`; nothing else from this module |
| `internal/feedback/outbox` | drain, local ledger, dedupe and caps, scrub tripwire, queue (reusing `feedback.QueueStore`), outbox log, preview, purge | `internal/bugreport`, `internal/feedback`, `internal/config`; no `os/exec`, no `net/http` |
| `internal/feedback/publish` | sender, `gh` runner seam, duplicate lookup, occurrence comment, issue contract rendering, deterministic template text, model seam interface, daily model-call cap | `internal/bugreport`, `internal/feedback/outbox`, `internal/feedback`, `internal/config`; `os/exec` only in `ghrunner.go`; no `net/http` |

Measured import facts (`go list -deps <pkg>`, filtered for this module's packages, tree `bb54f2903`):

- `internal/config` depends on `internal/config/atomicfile`, `internal/defs`, `internal/paths`, `pkg/models` and nothing else of this module: a leaf `bugreport` may import it.
- `internal/template` depends on `internal/config` and `internal/manifest` only. `bugreport` still does not import it: that would embed the template tree into every package that must call capture (the recover sites in `resilience`, `escalation`, `guardliveness`, `navigator`), so template sentinels are mapped to closed tokens by the call site in `internal/cli`, which already imports `internal/template`.
- `internal/hook` depends on `internal/cli/preference`, `internal/template`, `internal/config`, `internal/resilience`, `internal/escalation`, `internal/guardliveness` and does not depend on `internal/cli` or `internal/feedback`.
- `internal/feedback` depends on `internal/hook`, `internal/cli/preference`, `internal/escalation`, `internal/resilience`, `internal/guardliveness` and does not depend on `internal/cli`.
- `internal/cli` imports `internal/feedback` (`internal/cli/feedback.go`) and will import `publish` for the flush call. Therefore `publish` cannot import `internal/cli` (a cycle), and the production model implementation is injected from `internal/cli` (section 8). `bugreport` and the hook, template, and CLI callers cannot form a cycle because `bugreport` imports none of them. Putting capture inside `internal/feedback` would make `hook` import `feedback`, which imports `hook`: a cycle.

Simplicity ladder: reuse before build. Reused unchanged: `feedback.Scrub` and the classifier (`internal/feedback/scrub.go`, `classify.go`), `feedback.QueueStore` (`queue.go`, locked atomic `Mutate`, 0600) pointed at a second file, the mask log (`masklog.go`), `version.GetVersion()` and `version.GetCommit()` (`pkg/version/version.go`), `runtime.GOOS` and `runtime.GOARCH`, `settings.ApplySchemaEdits` (the one writer the console drives), the Jev precedent for the wizard question and its persistence call (`saveBoolAnswer`, `applyJevFromWizard`, translations), `isInteractiveStdin`, `promptBool`, the existing headless `claude` runner (section 8), and `paths.MoaiHome()` for the user-scoped file.

## 2. Kind, verdict, and emit-site register

Six kinds. Each row states the verdict rule, the **derivation mode**, and the verified emit sites. Derivation mode `D` means the verdict is derived inside `bugreport` from a type or sentinel `bugreport` can import (standard library, `internal/config`, or the marker type `bugreport` itself defines). Derivation mode `A` means the call site asserts the kind (and, for template and harness kinds, a closed token), because the sentinel lives in a package `bugreport` cannot import (`internal/hook`, `internal/template`, `internal/cli`).

| Kind | Verdict rule | Mode | Emit site (verified) | What the site sees |
|---|---|---|---|---|
| `panic` (main) | `moai` (a panic is a moai defect by definition) | A | new deferred recover in `cmd/moai/main.go` `main` (no recover exists there and none in `internal/cli/root.go:69` `Execute`); then re-panic so the runtime's crash output and exit code survive | a Go panic on the main goroutine, including a hook invocation, which runs through the same `main` |
| `panic` (recover sites) | `moai` | A | `internal/guardliveness/evaluator.go:160`, `internal/cli/codex_stop_chain.go:426`, `internal/resilience/circuit.go:229`, `internal/navigator/route/run.go:51`, `internal/navigator/fix/request.go:105`, `internal/escalation/detector.go:137`, `internal/hook/session_start_guard_liveness.go:205`, `internal/hook/user_decision_capture.go:92`, `internal/hook/session_start_binary_lag.go:72`, `internal/hook/navigator_detect.go:158` and `:180` | a recovered panic that today leaves no trace |
| `panic` (allowlisted, not captured) | n/a | n/a | `internal/hook/trace/writer.go:86`, a recover that defends a closed-channel send race and drops one trace entry by design | benign by construction; the allowlist entry carries that reason |
| `hook_handler_failure` | `moai` only when the error chain carries the `bugreport` internal marker; else the environment and user rows of section 3; else `ambiguous` | kind A, verdict D | `internal/hook/registry.go:133` (`handler %d for event %s: %w`) | the handler's error chain and the registered event constant; the handler's registered name is the only identity used |
| `hook_timeout` | `ambiguous` | A | `internal/hook/registry.go:125` (`ErrHookTimeout`, defined in `internal/hook`, so the call site asserts it) | the sentinel only; load and a stuck handler look alike without more data |
| `internal_error` | `moai` (the marker is the attribution) | D | `internal/cli/preference/cmd.go:221` (`preference: internal error: store is not *fileStore`) wrapped by the `bugreport` marker | a violated internal invariant |
| `template_deploy_failure` | token `not_found`, `preserve_integrity`: `moai`; token `path_traversal`: `moai` but withheld by REQ-ANON-012; unmatched chain: the section 3 rows, else `ambiguous` | kind and token A, chain rows D | `internal/cli/update_template_sync.go:535` (`deploy templates`) and `internal/cli/update_clean_install.go:599` (`PRESERVE integrity violation`); sentinels `ErrPathTraversal` (`internal/template/deployer.go:457,462,474`) and `ErrTemplateNotFound` (`deployer.go:359`, `renderer.go:122`) mapped to tokens by `errors.Is` in `internal/cli` | the error chain; non-sentinel filesystem errors resolve to `environment` by row A1 |
| `harness_defect` | token `missing_key`: `moai`; tokens `unexpanded_token`, `invalid_json`: `ambiguous` | A | `internal/cli/update_template_sync.go:439` (`template validation`) and `:535`; sentinels at `internal/template/renderer.go:135,149` and `internal/template/validator.go:56,59` mapped to tokens in `internal/cli` | a shipped asset that fails to render or validate; the two ambiguous tokens can also be caused by a user-supplied value rendered into the output |

The recover guard (REQ-ANON-006) is a go/parser walk over non-test files in `internal/`, `cmd/`, and `pkg/`: every `recover()` call site either sits in a function that calls the capture entry point or appears on the allowlist with a reason. Inventory verified this run: twelve sites in `internal/` (listed above plus the allowlisted one) and none in `cmd/` or `pkg/`.

**Panic coverage (stated scope).** A deferred `recover` in `main` recovers only panics on the main goroutine; a panic on any other goroutine terminates the process without running `main`'s defer. Hooks run synchronously on the main goroutine (`registry.go:108`, `h.Handle`), so hook panics are covered. Goroutine launch statements in non-test code: 43 files contain one (`grep -rEl '^[[:space:]]+go (func|[A-Za-z_.]+\()' internal cmd pkg --include='*.go' --exclude='*_test.go'`, measured at `bb54f2903`). Wrapping those entry points is out of scope; the gap is a residual risk, and a later helper (a goroutine launcher that recovers into capture) is the repair.

The signal "deployed harness defect" is covered only by the machine-detectable ones: hook failures, shipped-template render and validate failures, binary panics, and CLI internal errors (the kinds above). A behavioural defect of a deployed agent or skill has no machine-detectable signal in the tree at this commit, and the narrowing of the card text is an accepted scope decision (plan.md section B, DEC-8); a follow-up SPEC may add a deployed-asset self-check as a new signal source. The closed-kind test (AC-007) pins the token set, so extending coverage later is one visible change.

## 3. Attribution rules (deterministic, ordered, first match wins, `moai` is an allowlist)

Evaluated in Go by inspecting error types with `errors.Is` and `errors.As` only. Error text is never read, which is also why the panic value and `err.Error()` never reach the payload. The derivation column uses the section 2 vocabulary.

| Order | Rule | Verdict | Mode | Verified basis |
|---|---|---|---|---|
| A1 | chain contains a filesystem or syscall error (`*fs.PathError`, `*os.SyscallError`, `syscall.Errno`) | `environment` | D | standard library types; the permission, disk, and missing-path failures the card lists |
| A2 | chain contains a network error (`net.Error`, `*url.Error`, `*net.OpError`, `*net.DNSError`) | `environment` | D | standard library types |
| A3 | chain contains `*exec.Error`, `exec.ErrNotFound`, `*exec.ExitError`, `os.ErrPermission`, or `context.Canceled` | `environment` | D | standard library types; a subprocess exit (for example `git` in a user's repository state) is not moai's defect |
| U1 | chain contains `config.ErrConfigNotFound`, `ErrInvalidConfig`, `ErrInvalidYAML`, `ErrSectionTypeMismatch`, or `ErrInvalidDevelopmentMode` | `user` | D | `internal/config/errors.go:17-38`; `internal/config` is importable |
| U2 | chain contains `*json.SyntaxError` or `*json.UnmarshalTypeError` | `user` | D | standard library types; a decode failure is an input problem, not a moai invariant |
| U3 | chain contains `*yaml.TypeError` | `user` | D | `gopkg.in/yaml.v3` is already a dependency of `internal/config`; the type is confirmed in the run phase (a yaml syntax error is untyped in v3 and reaches row F, which is safe because F is not `moai`) |
| M1 | the signal is a panic | `moai` | A (the recover site asserts it) | a panic is a defect of whatever code panicked |
| M2 | chain carries the `bugreport` internal marker | `moai` | D | the marker type is defined in `bugreport` |
| M3 | the call site supplied a closed token whose register row says `moai` (section 2: `not_found`, `preserve_integrity`, `path_traversal`, `missing_key`) | `moai` | A | `internal/cli` maps template sentinels to tokens |
| F | no row matched | `ambiguous` | n/a | the fallback: every unrecognised error, a hook handler error without a marker, the hook timeout, and the tokens `unexpanded_token` and `invalid_json` |

Rows A and U run before M, so an internal marker wrapping an `*fs.PathError` resolves to `environment` (the cause is the environment even though moai noticed it). The register no longer lists `claudeNotFoundError`: it is unexported in `internal/cli` (`claude_binary.go:46`), `bugreport` cannot name it, and no registered emit site produces it. A missing `claude` binary at an exec site surfaces as `*exec.Error` and resolves by row A3 if it ever reached capture.

Only `moai` proceeds. `user` and `environment` stay local and end there (no queue, no model call). Because `moai` is an allowlist, a new error type that no row recognises can never become a public report by omission; the defect form where an unrecognised error "falls through to moai" does not exist.

An `ambiguous` verdict is retained: the signal stays local with zero model calls, nothing is queued, and the drain appends one `ambiguous` log row (section 6). There is no policy constant and no model adjudication path: the former `adjudicate` variant — one bounded model call per ambiguous item, its answer persisted on the queue item so a retry never asked again — was removed by the operator's decision (plan.md section B, DEC-7) and a follow-up SPEC may add it. AC-009 tests the retention.

## 4. Fingerprint, frame filter, and the `detail` closed sets

Canonical input string, newline separated: schema marker `v1`, `version.GetVersion()`, `version.GetCommit()`, `GOOS/GOARCH`, kind, then the ordered frame names. The fingerprint is the first 16 lowercase hex digits of the SHA-256 of that string.

Frames come from `runtime.Callers` and `runtime.CallersFrames`. Keep a frame only when `Frame.Function` begins with `github.com/modu-ai/moai-adk/`; strip that prefix; drop the capture package's own helper frames and `runtime.gopanic`; keep at most 12, innermost first. File, line, and arguments are never read from the frame. This matters because `.goreleaser.yml:20-24` and the `Makefile` `LDFLAGS` line carry `-s -w` and `-X` flags and no `-trimpath` (verified: a search for `trimpath` in both files returns nothing), so file paths in a frame embed the builder's absolute paths. Function names are safe; for generic functions the runtime prints `[...]` for type arguments rather than concrete types (inferred, not verified: the M1 first test pins it with a generic canary).

A panic with zero moai frames stays local. Other kinds capture from the call site, which is moai code, so a zero-frame result cannot occur for them.

**`detail` is closed-set membership, in an enum type.** The builder's `detail` parameter is a dedicated type, never a `string`:

- Template and harness kinds: a `TemplateToken` enum with exactly six members (`path_traversal`, `not_found`, `preserve_integrity`, `missing_key`, `unexpanded_token`, `invalid_json`), each valid only for the kinds its register row names.
- Hook kinds: a pair of opaque identifiers, `(EventID, HandlerID)`, obtainable only by registration. `internal/hook` registers, at package initialisation, the set of its `EventType` constants and the registered names of the production handlers (the type name of each handler with the package qualifier and pointer marker stripped; the `Handler` interface has no name method, so `%T` is the only identity, `internal/hook/types.go:624`). A name that does not match `^[A-Za-z][A-Za-z0-9]{0,63}$` is refused at registration, so no `/`, `.`, or path shape can enter. A guard in `internal/cli` (which wires every production handler in `internal/cli/deps.go:233-293`) asserts every registered handler's name is in the table and the table has no extra names (`TestEveryRegisteredHookHandlerHasBugreportName`), and a go/parser guard asserts the registered event set equals the `EventType` constants declared in `internal/hook/types.go`.
- Read-back: a queued payload is untrusted because the queue file is a local file. `ParseDetail(kind, s)` returns membership or a rejection; the sender re-validates every field, including `detail`, before it renders.

The canary test feeds `CANARY-/Users/leak/secret-token`, `WorktreeCreate//srv/customer/private-project`, and an empty string directly to the registration and read-back validators and expects rejection in every case, so an all-accepting validator or a builder taking `string` cannot survive.

## 5. Payload schema v1

Fixed fields: `schema` (`"v1"`), `kind` (enum), `fingerprint` (16 hex), `version` (semver-like allowlist), `commit` (7 to 40 hex), `os` and `arch` (the Go value sets), `frames` (array of allowlisted names), `detail` (optional, the section 4 type). No timestamp is sent, so the report does not disclose when the user worked. No session id, user name, host, path, repository name, or environment value exists in any field.

Construction takes typed inputs only (kind, frames, detail, build identity); there is no `error`, `any`, or free-string parameter, and a type-level test pins the signature (AC-011). The existing scrubber and classifier then run over the rendered title and body as a tripwire: a clean payload yields zero findings; any finding means a validator gap and the payload is withheld rather than masked and sent.

Classifier caveat, accepted: the classifier vocabulary is English words such as "vulnerability" or "exploit"; a moai function named with such a word would block its own report. That fails safe (nothing is published) and is listed under risks.

## 6. Local pipeline

Capture appends one JSONL line (kind, frames, detail) to `.moai/state/bugreport/spool.jsonl`, bounded at 200 lines and 64 KiB; when full the signal is dropped. Capture first reads the user-scoped value and returns when it is not true; it does no attribution beyond the cheap environment-type check, no hashing of large inputs, no network, no model, and abandons its work after a 50 ms time box (the hook-path discipline in the repository's Advisory-Check rule: advisory work is time-boxed and fail-open). The spool and queue live under the project's `.moai/state/` (gitignored in this repository, and in the shipped template `.gitignore` line `.moai/state/`); they hold only closed-schema fields.

The drain (run by flush) reads the spool and, per signal: attribute, then continue only for `moai`; fingerprint; local ledger check (per-fingerprint window); caps; build and validate the payload; render title and body; scrub tripwire; enqueue to `.moai/state/bugreport/queue.json` through `feedback.QueueStore`; append the outbox row. Each stage that stops a signal appends one log row naming the reason (`user`, `environment`, `ambiguous`, `deduped`, `capped`, `withheld`) and no payload text for non-queued outcomes. Queue items gain optional `omitempty` fields (fingerprint, kind, stored summary, summary decision `model` or `template`), so existing readers of the manual queue stay compatible.

The outbox log `.moai/logs/bugreport-outbox.log` is append-only JSONL, mode 0600, holding the exact payload for `queued`, `sent`, and `withheld` rows. Preview prints the queued payloads through the same render function the sender uses, which is how AC-014 proves byte identity rather than similarity.

Flush triggers are `moai feedback participation flush` and the end of a plain `moai update`; the decision record is in plan.md. Flush never runs inside a hook dispatch and is time-boxed to 10 seconds in total.

## 7. Publication through the user's own gh

All steps are Go code invoking `gh`. The flag surface below was read from `gh 2.92.0` help on this machine (`gh issue list --search --state --json --limit --repo`, `gh issue create --title --body-file --repo`, `gh issue comment --body-file --repo`, and the `--json` fields `number,title,state,comments,body,url`). A read-only probe against `modu-ai/moai-adk` confirmed that `gh issue list --json comments` returns, per issue, an array of comment objects whose keys include `body`, `author`, `createdAt`, and `url`, so occurrence markers can be counted from `body`. Not verified: whether that array is truncated for issues with very many comments, any live create or comment call, the search tokenisation of a 16-hex token under `in:title`, and index latency for a just-created issue.

1. Check consent (re-read from the user-scoped file) and `gh` presence and authentication; otherwise leave the item queued, quietly.
2. Target repository: the user-scoped `participation.repository` when set and well-formed, else the compiled default `modu-ai/moai-adk` (`config.DefaultFeedbackRepository`). The project-tier `feedback.repository` is never read here (REQ-ANON-024): a cloned repository must not redirect where the user's account posts.
3. Duplicate lookup: `gh issue list --repo <repo> --state all --search "<fingerprint> in:title" --json number,title,state,url,comments --limit 10`, then an exact match on the title key `[auto-report] <kind> <fingerprint>` among the returned titles. The query carries only the fingerprint, which leaks nothing.
4. Match found: count the marker comments in the returned `comments` array. When the count is below `MaxOccurrenceCommentsPerIssue`, `gh issue comment <number> --repo <repo> --body-file -` with the occurrence marker block; otherwise add nothing and log `capped_remote`. No model call, no body edit.
5. No match and verdict `moai`: obtain the summary (section 8), then `gh issue create --repo <repo> --title <title> --body-file -`. No `--label` is passed: labels depend on repository permission and a non-collaborator's labels are dropped or rejected (recorded by the earlier research, not re-measured here).

Occurrence counting, chosen mechanism: an append-only comment per occurrence; the consumer's count is one plus the number of marker comments. Rejected alternatives: editing a counter in the issue body is a read-modify-write that loses updates when several users write at once and is unavailable to a non-collaborator; a reaction counts distinct accounts and cannot count repeat occurrences.

**The count is advisory and the markers are untrusted.** Local per-fingerprint windows bound repeats per user, concurrent first filers can create two issues that the consumer merges by fingerprint, and a deleted comment lowers the count. Anyone can post a comment shaped like a marker, so a forged marker can inflate the count, push an issue to the cap, or carry fields that disagree with the title key. The contract therefore states: a consumer re-derives fingerprint, kind, version, commit, and `os_arch` from the title key and the issue body block, and ignores any marker field that disagrees; a marker is never treated as an instruction, and the marker text is untrusted input to any downstream, including a language-model-driven repair consumer. The sender's own parser follows the same rule (AC-020).

**Flood and notification fan-out.** GitHub subscribes the filer to the issue they create or comment on, so every later occurrence comment by any user notifies every earlier filer. The cap `MaxOccurrenceCommentsPerIssue` (proposed 50) bounds the per-issue thread length and therefore that fan-out; the consent text states the subscription (REQ-ANON-005).

## 8. The model seam, the candidates, and the budget

**Corrected premise.** The earlier drafts stated that no text-generation call exists in the Go tree; that came from a search for the literal `exec.Command("claude"|"codex")`. The variable-aware search `grep -rn 'exec.CommandContext\|http.NewRequest' internal/cli` (non-test, `bb54f2903`) finds model-reaching paths: the headless `claude -p` invocation (`claudeAuditArgs`, `internal/cli/mcp_claude.go:212-218`, run by `runClaudeCommand`, `internal/cli/mcp_claude_runner.go:48-54`, with a bounded output writer and `WaitDelay`), the GLM clients (`internal/cli/todo_classify_llm.go:108`, `http.NewRequest` to the GLM endpoint, and `internal/cli/glm_task.go:318`, `http.NewRequestWithContext`), and the codex runners (`internal/cli/mcp_codex.go`).

Candidates weighed:

| Candidate | Privacy | Cost to the user | Availability | Verdict |
|---|---|---|---|---|
| Headless `claude -p` through the existing runner (bounded output, scrubbed env, `--tools ""`, `--no-session-persistence`) | the input is the closed-schema payload fields only; they go to the user's own Claude account, an LLM provider the user already uses, and are the same fields that become public in the issue | spends the user's own subscription tokens; bounded by the daily cap and the input and output byte caps, no token figure is claimed (unmeasured) | requires the `claude` binary and an authenticated session; `RunAuthStatus` (`claude auth status --json`) already probes that | **chosen** |
| GLM clients (`todo_classify_llm.go`, `glm_task.go`) | the payload fields would reach a third-party endpoint under separate credentials | spends GLM quota the user may not have | requires GLM credentials | rejected |
| Codex runners (`mcp_codex.go`) | payload fields reach the user's Codex account | spends Codex quota | requires the codex binary and login | rejected: no benefit over the Claude path for a one-paragraph summary |
| No model (template text only) | none | none | always | the fallback for every unavailable or failed model path |

**Decision (recorded, plan.md Decisions):** reuse the existing headless `claude` runner through a single model seam; the consent text names that the summary may spend the user's own subscription tokens; the deterministic template text is used whenever the model is unavailable, unauthenticated, fails, or returns output that fails validation.

**Seam shape.** `publish` owns the `Summarizer` interface whose methods take validated payload fields only (today the interface serves the summary alone). The production implementation lives in `internal/cli` (`feedback_participation_model.go`), reusing the runner and a flag set derived from `claudeAuditArgs` without the audit `--json-schema`, and is injected where `internal/cli` calls flush. `publish` imports no model helper and no `internal/cli`; the static guard (REQ-ANON-025, AC-025) enforces it. The M6 first test records that the flag set `claudeAuditArgs` passes today is accepted by the installed `claude` for a summary prompt.

**Per-item call bound.** A queue item records its decisions as it goes, so a retry never repeats a paid step:

| Stage | Model call | Stored on the queue item before the next stage |
|---|---|---|
| remote lookup finds an issue | none | none |
| remote lookup finds none, verdict `moai` | at most 1 summary, only when no stored summary exists | the validated summary or the template decision |
| create fails and the item is retried | none (the stored summary is reused) | attempt count |

So per queue item the total is at most 1 call across any number of retries. Every call counts against `MaxModelCallsPerDay`.

```
capture (Go) -> attribute (Go) -> user/environment: stop
                              -> ambiguous: stop (retained locally, logged)
                              -> moai: fingerprint -> dedupe/caps (Go) -> scrub tripwire (Go)
                                       -> remote duplicate lookup (gh) -> exists: comment (or skip at cap), stop
                                       -> absent: [model: summary, stored] -> create (retry reuses stored summary)
```

Guards: the model seam is one interface so a counting stub proves the budget at every entry point; a rolling daily cap `MaxModelCallsPerDay` bounds worst-case spend; model input is a fixed prompt template plus the validated payload fields, asserted by a golden test; model output is length- and character-allowlisted, scrubbed, and classified before use, and any failure falls back to the deterministic template, so the issue is filed either way and the model is an enhancement rather than a dependency.

## 9. Consent surfaces

- **The user-scoped store.** The consent lives in `<moai home>/config/participation.yaml`, resolved through `paths.MoaiHome()` (the `MOAI_HOME` environment variable overrides it, which is how tests isolate it; the `config` directory is created 0700 by the existing home layout contract, `internal/homestate/paths.go` `EnsureHomeLayout`, and the writer creates the file 0600). Keys: `participation.enabled`, `participation.asked`, optional `participation.repository`. The file sits beside, not inside, `config/sections/`: the config resolver merges `~/.moai/config/sections/*.yaml` as its user tier together with the project and local tiers by priority (`internal/config/resolver.go:312-335`, `source.go`), so a merged read lets a project-tier file decide the result, and a consent value must never allow that. Precedent for a direct home-tier read with its own wrapper struct: `loadHomeRetentionDays`, `internal/cli/clean_home.go:97-120`.
- **The reader** (`internal/config`, new file, imported by `bugreport`, `outbox`, `publish`, `settings`, and `cli`) is fail-closed: a missing file, key, or directory, an unreadable or unparseable file, or a wrongly typed value yields `false`/`false`/empty. `participation.repository` is used only when it matches `^[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+$`. The reader never opens a project file, so a tracked file that says `participation: true` cannot enable anything by construction; the mutant fixture in AC-001 and AC-003 (tracked file says true, no user file) proves it.
- **The writer** is one function in `internal/settings` (a patch of the home file through `yamlpatch.PatchFile`, atomic, preserving unknown keys), driven by the wizard, the update step, and the console branch of `ApplySchemaEdits`.
- **Init**: the question lands in `Page3Questions` after `jev_enabled`, in its own group, not required, absent from `DefaultQuestions` and `ReconfigureQuestions` (the Jev slot and placement rule, `internal/cli/wizard/questions.go`). Its default is the stored user-scoped value (no when none is stored), so re-running init never silently withdraws an existing consent. `applyParticipationFromWizard(wizardRan, res, root)` mirrors `applyJevFromWizard`: no write unless the wizard ran; additionally no write when `CI` is non-empty; it writes both values through the writer. Rejected alternative: sharing the Jev page, because the Jev code comment records that a taller page scrolls the step indicator off screen.
- **Update**: after template synchronisation succeeds, `runParticipationStep` runs only on a plain template-sync run (section below), in an interactive terminal, with `CI` empty and `participation.asked` false. It reuses the init question text per locale (single source of truth), defaults to no, and persists both values; an empty answer persists `enabled` false and `asked` true. `moai update` previously carried no prompt of its own; the profile-wizard guard test counts profile-wizard seam calls only, so it is unaffected (read in `internal/cli/init_update_profile_path_test.go`, which asserts a profile-wizard seam count of zero).
- **The update flag inventory** (read from `internal/cli/update.go:71-93`): `--check`, `--shell-env`, `--config`/`-c`, `--force`, `--yes`, `--templates-only`, `--binary`, `--dry-run`, `--no-hooks`, `--no-plugin`, `--restore`, `--verbose`, `--profile` (retired, accepted), `--version`. Mode flags, after which the step never asks: `--check`, `--shell-env`, `--config`, `--binary`, `--dry-run`, `--restore`, `--version`, `--templates-only`, `--yes`. Modifier flags that leave the run a plain template sync: `--force`, `--no-hooks`, `--no-plugin`, `--verbose`, `--profile`. The rule is "only a plain template-sync run asks", and a test enumerates the command's flag set and fails when a flag is in neither list, so a new flag forces a decision rather than silently asking.
- **Marker semantics**: `participation.asked` is true once the question has been answered at init or update, or once the console changes `participation.enabled`. It exists because the default `false` cannot distinguish "never asked" from "asked, said no". Because the console write also sets `asked` true, a user who toggled in the console before any prompt is not asked again; an unchanged console submission records nothing.
- **Web**: one `TypeBool` field `feedback.participation` in `SectionFeedback` beside `auto_submit`, with a new persist kind `PersistUserScoped` (`internal/settings/schema.go`). The kind's wiring: `sectionFileFor` returns "" for it (it has no project section file); `SchemaCurrentValues` reads its value from the user-scoped reader; `ApplySchemaEdits` gains a branch that keeps the value-invariant gate (a submission equal to the persisted value, or an absent key submitted false, writes nothing) and otherwise writes `participation.enabled` and `participation.asked: true` in one patch through the writer; `schemaEditableField` (`internal/web/schemaform.go:362`) admits the new kind. `participation.asked` has no `FieldDef`, so it is not rendered, not counted by `TestFeedbackPanelFieldsWired`, and needs no i18n keys. The console renders every bool as the existing two-option radio pair with a hidden `<name>__present` companion (`internal/web/widget_policy_test.go:59-93`: `TestBoolFieldsRenderAsRadio`, `TestSchemaTogglePresentCompanion`), so the field renders that way with no widget code; the earlier "checkbox" wording was wrong and a literal checkbox would fail `TestBoolFieldsRenderAsRadio`. Four-locale title and description keys land in `internal/web/assets/i18n.js`; the tab blurb in `internal/web/settings_shell.go:158` is extended. The comment at `internal/settings/schema_sections.go:442` mentions an `isJevFieldName` routing function; no such function exists in the tree (a search finds only that comment), so there is no routing precedent to follow and none is needed. Prefer bool over text: `widget_policy_test.go` keeps a free-text whitelist this row must not join.

## 10. Constants (proposed; defined in `internal/config/defaults.go`)

| Constant | Proposed value | Meaning |
|---|---|---|
| per-fingerprint window | 7 days | no re-queue of the same fingerprint inside it |
| global cap, daily | 3 per rolling 24 hours | reports queued |
| global cap, weekly | 10 per rolling 7 days | reports queued |
| queue bound | 20 items | oldest dropped beyond it |
| attempt limit | 5 | then dropped with a log row |
| spool bound | 200 lines and 64 KiB | capture drops beyond it |
| capture time box | 50 ms | on the hook path |
| flush time box | 10 s | whole flush |
| frame limit | 12 | innermost first |
| daily model-call cap | 6 | summary calls |
| per-issue occurrence-comment cap | 50 | the sender adds no comment at or beyond it (advisory) |
| model input and output caps | fixed byte limits | bound token spend; values fixed in the run phase |
| ambiguous handling | retained locally (DEC-7, final) | no model call, nothing queued; no policy constant |

## 11. Out-of-design

No relay, bot, shipped credential, or non-GitHub intake is designed; the earlier anonymity options were dropped by the operator and are not carried here. No repair automation is designed; the contract in section 7 is the only interface offered to it, and REQ-ANON-023 keeps every repair artifact out of deployed templates. Goroutine entry-point wrapping is not designed (section 2).
