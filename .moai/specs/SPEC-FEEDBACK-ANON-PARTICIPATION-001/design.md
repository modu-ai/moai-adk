# Design — SPEC-FEEDBACK-ANON-PARTICIPATION-001

Design notes for the run phase. The spec states what must hold; this file states the shape that makes it hold. Citations were re-read in this tree (commit `2f492df19`) unless marked unverified.

## 1. Package layout and the import-cycle constraint

Three new packages, one per concern, so the capture path stays tiny and the network path stays isolated.

| Package | Holds | Imports (allowed) |
|---|---|---|
| `internal/bugreport` | kind enum, verdict enum, attribution rules, frame filter, fingerprint, payload type and validators, capture entry point, spool | `internal/config` (to read `feedback.participation`), standard library only otherwise |
| `internal/feedback/outbox` | drain, local ledger, dedupe and caps, scrub tripwire, queue (reusing `feedback.QueueStore`), outbox log, preview, purge | `internal/bugreport`, `internal/feedback` |
| `internal/feedback/publish` | sender, `gh` runner seam, duplicate lookup, occurrence comment, issue contract rendering, deterministic template text, model seam, daily model-call cap | `internal/bugreport`, `internal/feedback/outbox` |

Why `bugreport` is a leaf: hook handlers, the template deployer, and the CLI must call capture, and `internal/feedback` already imports `internal/hook` (classifier policy). Verified this run with `go list -deps`: `internal/config` depends on none of `hook`, `template`, `feedback`, `cli`; `internal/template` depends on `internal/config` and on none of `hook`, `feedback`, `cli`; `internal/hook` depends on `internal/template` and `internal/config` and not on `feedback` or `cli`. So `hook`, `template`, and `cli` may import `bugreport`, and `bugreport` may import `config`, with no cycle. Putting capture inside `internal/feedback` would make `hook` import `feedback`, which imports `hook`: a cycle.

Simplicity ladder: reuse before build. Reused unchanged: `feedback.Scrub` and the classifier (`internal/feedback/scrub.go`, `classify.go`), `feedback.QueueStore` (`queue.go`, locked atomic `Mutate`, 0600) pointed at a second file, the mask log (`masklog.go`), `version.GetVersion()` and `version.GetCommit()` (`pkg/version/version.go`), `runtime.GOOS` and `runtime.GOARCH`, `settings.ApplySchemaEdits` (the one writer both the wizard and the console drive), the Jev precedent (`JevQuestionID`, `saveBoolAnswer`, `applyJevFromWizard`, translations), `isInteractiveStdin`, and `promptBool`. New code is limited to the capture, attribution, fingerprint, ledger, publication, and rendering that do not exist.

## 2. Kind, verdict, and emit-site register

Six kinds. Each row maps to a default verdict and to the verified sites that call capture. Sites were read in this tree; line numbers are leads for the run phase, which re-reads before editing.

| Kind | Default verdict | Emit site (verified) | What the site sees |
|---|---|---|---|
| `panic` | `moai` | new deferred recover in `cmd/moai/main.go` `main` (verified: no recover exists there, and no top-level recover in `internal/cli/root.go:69` `Execute`); then re-panic so the runtime's own crash output and exit code survive | a Go panic anywhere in the process, including a hook invocation, which runs through the same `main` |
| `panic` | `moai` | the existing swallowing `recover()` sites, each gaining one capture call: `internal/guardliveness/evaluator.go:160`, `internal/cli/codex_stop_chain.go:426`, `internal/resilience/circuit.go:229`, `internal/navigator/route/run.go:51`, `internal/navigator/fix/request.go:105`, `internal/escalation/detector.go:137`, `internal/hook/session_start_guard_liveness.go:205`, `internal/hook/user_decision_capture.go:92`, `internal/hook/session_start_binary_lag.go:72`, `internal/hook/navigator_detect.go:158` and `:180` | a recovered panic that today leaves no trace |
| `panic` (allowlisted, not captured) | n/a | `internal/hook/trace/writer.go:86`, a recover that defends a closed-channel send race and drops one trace entry by design | benign by construction; the allowlist entry carries that reason |
| `hook_handler_failure` | `moai` after rules A and U | `internal/hook/registry.go:133` (`handler %d for event %s: %w`) | the handler's error chain and the registered event constant; the handler type via `%T` is the only identity used |
| `hook_timeout` | `ambiguous` | `internal/hook/registry.go:125` (`ErrHookTimeout`) | the sentinel only; load and a stuck handler look alike without more data |
| `internal_error` | `moai` | `internal/cli/preference/cmd.go:221` (`preference: internal error: store is not *fileStore`) through an explicit marker wrapper; the marker is the attribution | a violated internal invariant |
| `template_deploy_failure` | `moai` for the sentinels below, rules A and U otherwise | `internal/cli/update_template_sync.go:535` (`deploy templates`) and `internal/cli/update_clean_install.go:599` (`PRESERVE integrity violation`); sentinels `ErrPathTraversal` (`internal/template/deployer.go:457,462,474`) and `ErrTemplateNotFound` (`deployer.go:359`, `renderer.go:122`) | the error chain; non-sentinel filesystem errors resolve to `environment` by rule A1 |
| `harness_defect` | `moai` for `ErrMissingTemplateKey`; `ambiguous` for `ErrUnexpandedToken` and `ErrInvalidJSON` | `internal/cli/update_template_sync.go:439` (`template validation`) and `:535`; sentinels at `internal/template/renderer.go:135,149` and `internal/template/validator.go:56,59` | a shipped asset that fails to render or validate; the two ambiguous sentinels can also be caused by a user-supplied value rendered into the output |

The recover guard (REQ-ANON-006) is a go/parser walk over non-test files in `internal/`, `cmd/`, and `pkg/`: every `recover()` call site either sits in a function that calls the capture entry point or appears on the allowlist with a reason. Inventory verified this run: twelve sites in `internal/` (listed above plus the allowlisted one) and none in `cmd/` or `pkg/`.

Unverified, recorded as a gap: a runtime defect of a deployed agent or skill (as opposed to a hook handler or a shipped template that fails to render) has no machine-detectable signal in the tree at this commit. The `harness_defect` kind covers only the render and validate sentinels above; plan.md carries the open question.

## 3. Attribution rules (deterministic, ordered, first match wins)

Evaluated in Go by inspecting error types with `errors.Is` and `errors.As` only. Error text is never read, which is also why the panic value and `err.Error()` never reach the payload.

| Order | Rule | Verdict | Verified basis |
|---|---|---|---|
| A1 | chain contains a filesystem or syscall error (`*fs.PathError`, `*os.SyscallError`, `syscall.Errno`) | `environment` | the standard library error types; the permission, disk, and missing-path failures the card lists |
| A2 | chain contains a network error (`net.Error`, `*url.Error`, `*net.OpError`, `*net.DNSError`) | `environment` | the standard library error types |
| A3 | chain contains `*exec.Error`, `exec.ErrNotFound`, `os.ErrPermission`, `context.Canceled`, or `claudeNotFoundError` | `environment` | `internal/cli/claude_binary.go:46` defines `claudeNotFoundError` |
| U1 | chain contains `config.ErrConfigNotFound`, `ErrInvalidConfig`, `ErrInvalidYAML`, `ErrSectionTypeMismatch`, or `ErrInvalidDevelopmentMode` | `user` | `internal/config/errors.go:17-38` |
| K | kind-specific rows of the register above | `moai` or `ambiguous` as listed | section 2 |
| F | no row matched | `ambiguous` | fallback; under the policy constant it is retained locally or adjudicated once |

Only `moai` proceeds. `user` and `environment` stay local and end there (no queue, no model call). The unverified part: the exact error type the YAML library returns for a syntax error in a user file is not confirmed here; rule U1 relies on the `config` sentinels, and the run phase adds the YAML type after reading the loader (`internal/config/loader.go:372,426,478` are the decode sites).

The ambiguous policy constant `AmbiguousPolicy` takes `local` or `adjudicate`. Under `local` the signal stays local with zero model calls. Under `adjudicate` one model call receives only the validated payload fields and must answer exactly `moai` or `not-moai`; any other answer keeps the signal local. Both values are tested (AC-ANON-011). Interim shipped value until the operator resolves the clarification in plan.md: `local`, because it spends no tokens and publishes nothing doubtful; flipping it is a one-constant change in M6.

## 4. Fingerprint and frame filter

Canonical input string, newline separated: schema marker `v1`, `version.GetVersion()`, `version.GetCommit()`, `GOOS/GOARCH`, kind, then the ordered frame names. The fingerprint is the first 16 lowercase hex digits of the SHA-256 of that string.

Frames come from `runtime.Callers` and `runtime.CallersFrames`. Keep a frame only when `Frame.Function` begins with `github.com/modu-ai/moai-adk/`; strip that prefix; drop the capture package's own helper frames and `runtime.gopanic`; keep at most 12, innermost first. File, line, and arguments are never read from the frame. This matters because `.goreleaser.yml:20-24` and the `Makefile` `LDFLAGS` line carry `-s -w` and `-X` flags and no `-trimpath` (verified: a search for `trimpath` in both files returns nothing), so file paths in a frame embed the builder's absolute paths. Function names are safe; for generic functions the runtime prints `[...]` for type arguments rather than concrete types (inferred, not verified here; a run-phase test must pin it with a generic canary).

A panic with zero moai frames stays local. Other kinds capture from the call site, which is moai code, so a zero-frame result cannot occur for them.

The `detail` field exists for two cases only: hook kinds carry `<event>/<handler type>` and template kinds carry a closed token (`path_traversal`, `not_found`, `preserve_integrity`, `missing_key`, `unexpanded_token`, `invalid_json`). A value must match `^[A-Za-z0-9_.*/-]{1,96}$` and, for hook kinds, the event must be a registered event constant.

## 5. Payload schema v1

Fixed fields: `schema` (`"v1"`), `kind` (enum), `fingerprint` (16 hex), `version` (semver-like allowlist), `commit` (7 to 40 hex), `os` and `arch` (the Go value sets), `frames` (array of allowlisted names), `detail` (optional). No timestamp is sent, so the report does not disclose when the user worked. No session id, user name, host, path, repository name, or environment value exists in any field.

Construction takes typed inputs only (kind, frames, detail token, build identity); there is no `error`, `any`, or free-string parameter, and a type-level test pins the signature (AC-ANON-013). The existing scrubber and classifier then run over the rendered title and body as a tripwire: a clean payload yields zero findings; any finding means a validator gap and the payload is withheld rather than masked and sent.

Classifier caveat, accepted: the classifier vocabulary is English words such as "vulnerability" or "exploit"; a moai function named with such a word would block its own report. That fails safe (nothing is published) and is listed under risks.

## 6. Local pipeline

Capture appends one JSONL line (kind, frames, detail) to `.moai/state/bugreport/spool.jsonl`, bounded at 200 lines and 64 KiB; when full the signal is dropped. Capture does no attribution beyond the cheap environment-type check, no hashing of large inputs, no network, no model, and abandons its work after a 50 ms time box (the hook-path discipline in the repository's Advisory-Check rule: advisory work is time-boxed and fail-open).

The drain (run by flush) reads the spool and, per signal: attribute, then continue only for `moai`; fingerprint; local ledger check (per-fingerprint window); caps; build and validate the payload; render title and body; scrub tripwire; enqueue to `.moai/state/bugreport/queue.json` through `feedback.QueueStore`; append the outbox row. Each stage that stops a signal appends one log row naming the reason (`user`, `environment`, `deduped`, `capped`, `withheld`) and no payload text for non-queued outcomes.

The outbox log `.moai/logs/bugreport-outbox.log` is append-only JSONL, mode 0600, holding the exact payload for `queued`, `sent`, and `withheld` rows. Preview prints the queued payloads through the same render function the sender uses, which is how AC-ANON-016 proves byte identity rather than similarity.

Flush triggers are `moai feedback participation flush` and the end of `moai update`; further triggers are an open question in plan.md. Flush never runs inside a hook dispatch and is time-boxed to 10 seconds in total.

## 7. Publication through the user's own gh

All steps are Go code invoking `gh`. The flag surface below was read from `gh 2.92.0` help on this machine (`gh issue list --search --state --json --limit`, `gh issue create --title --body-file --repo`, `gh issue comment --body-file --repo`, and the `--json` fields `number,title,state,comments,body,url`). Not verified: any live call against GitHub, the search tokenisation of a 16-hex token under `in:title`, and index latency for a just-created issue.

1. Check consent and `gh` presence and authentication; otherwise leave the item queued, quietly.
2. Duplicate lookup: `gh issue list --repo <feedback.repository> --state all --search "<fingerprint> in:title" --json number,title,state,url --limit 10`, then an exact match on the title key `[auto-report] <kind> <fingerprint>` among the returned titles. The query carries only the fingerprint, which leaks nothing.
3. Match found: `gh issue comment <number> --repo <repo> --body-file -` with the occurrence marker block. No model call, no body edit.
4. No match and verdict `moai`: build the body (Go template plus an optional model summary section), then `gh issue create --repo <repo> --title <title> --body-file -`. No `--label` is passed: labels depend on repository permission and a non-collaborator's labels are dropped or rejected (the earlier research recorded labels as dropped without push access; not re-measured here). Triage labelling is a maintainer-side concern of the separate repair task.

Occurrence counting, chosen mechanism: an append-only comment per occurrence; the consumer's count is one plus the number of marker comments. Rejected alternatives: editing a counter in the issue body is a read-modify-write that loses updates when several users write at once and is unavailable to a non-collaborator; a reaction counts distinct accounts and cannot count repeat occurrences, though it is a candidate if the consumer prefers affected-user counts. The count is approximate by design: local per-fingerprint windows bound repeats per user, concurrent first filers can create two issues that the consumer merges by fingerprint, and a deleted comment lowers the count.

## 8. Model-call budget flow

Order of stages, with the model reachable at exactly one place per path:

```
capture (Go) -> attribute (Go) -> user/environment: stop
                              -> ambiguous: policy local: stop | adjudicate: [model call 1 of 1]
                              -> moai: fingerprint -> dedupe/caps (Go) -> scrub tripwire (Go)
                                       -> remote duplicate lookup (gh) -> exists: comment, stop
                                       -> absent: [model call: issue summary] -> create
```

Guards: the model seam is behind one interface per use (summary, adjudication) so a counting stub proves the budget at every entry point; a rolling daily cap `MaxModelCallsPerDay` bounds worst-case spend; model input is a fixed prompt template plus the validated payload fields, asserted by a golden test; model output is length- and character-allowlisted, scrubbed, and classified before use, and any failure falls back to the deterministic template, so the issue is filed either way and the model is an enhancement rather than a dependency.

How Go reaches a model is an open question in plan.md. Verified only that no text-generation call exists in the Go tree: the sole `exec.Command` of the `claude` or `codex` binary in non-test code is the availability probe in `internal/cli/doctor.go`. The seam keeps that choice out of every deterministic milestone.

## 9. Consent surfaces

- **Init**: the question lands in `Page3Questions` after `jev_enabled`, in its own group, default `"false"`, not required, absent from `DefaultQuestions` and `ReconfigureQuestions` (the Jev slot and placement rule, `internal/cli/wizard/questions.go`). The stepper denominator and the group golden change because the question count rises from 5 to 6; those pin tests are named in REQ-ANON-003. Rejected alternative: sharing the Jev page, because the Jev code comment records that a taller page scrolls the step indicator off screen.
- **Persistence**: `applyParticipationFromWizard(wizardRan, res, root)` mirrors `applyJevFromWizard`: no write unless the wizard ran; additionally no write when `CI` is non-empty. It writes both keys through `settings.ApplySchemaEdits`.
- **Update**: after template synchronisation succeeds, `runParticipationStep` runs only when `isInteractiveStdin()` is true, `CI` is empty, `--yes` is unset, no other mode flag is set, and both keys are false. It reuses the init question text per locale (single source of truth) and persists both keys. `moai update` previously carried no prompt of its own; the profile-wizard guard test counts profile-wizard seam calls only, so it is unaffected (verified by reading `internal/cli/init_update_profile_path_test.go`, which asserts a profile-wizard seam count of zero).
- **Marker semantics**: `participation_asked` is true once the question has been shown and answered at init or update. It exists because the shipped default `false` cannot distinguish "never asked" from "asked, said no". Known edge: a user who turns the toggle off in the console before ever being asked will be asked once at the next interactive update. Rejected alternative: one tri-state key, because the console toggle must be a boolean.
- **Web**: one `TypeBool` row in `internal/settings/schema_sections.go` beside `auto_submit`; four-locale title and description keys in `internal/web/assets/i18n.js`; the tab blurb in `internal/web/settings_shell.go:158` extended; guard tests as listed in AC-ANON-022. Prefer bool over text: `widget_policy_test.go` keeps a free-text whitelist this row must not join.

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
| daily model-call cap | 6 | summary plus adjudication combined |
| ambiguous policy | `local` (interim) | `local` or `adjudicate` |

## 11. Out-of-design

No relay, bot, shipped credential, or non-GitHub intake is designed; the earlier anonymity options were dropped by the operator and are not carried here. No repair automation is designed; the contract in section 7 is the only interface offered to it, and REQ-ANON-023 keeps every repair artifact out of deployed templates.
