---
id: SPEC-FEEDBACK-ANON-PARTICIPATION-001
title: "Opt-in improvement participation — automatic filing of moai-adk-attributed tool bugs from the user's own gh account, with deterministic attribution and a zero-token default path"
version: "0.2.0"
status: draft
created: 2026-10-04
updated: 2026-10-04
author: manager-spec
priority: P1
phase: "v3.3.0"
module: "internal/feedback"
lifecycle: spec-anchored
tier: L
depends_on: [SPEC-FEEDBACK-AUTO-SUBMIT-001]
related_specs: [SPEC-INIT-QUIET-WIZARD-001, SPEC-JEV-OPTIN-MEASURE-001, SPEC-WEBCONF-SIMPLIFY-001]
tags: "feedback, participation, opt-in, attribution, fingerprint, token-budget, privacy, wizard, web-console, template-first"
---

# SPEC-FEEDBACK-ANON-PARTICIPATION-001

## HISTORY

- 0.2.0 (2026-10-04) — Rewritten after the operator corrected the card scope: participation is real-name (the issue is filed from the participating user's own `gh` account), so the anonymity design was dropped entirely; automatic filing happens only when a deterministic attribution rule assigns the cause to moai-adk; token minimisation became a requirement (all detection, fingerprinting, duplicate checking, masking, caps, and deterministic attribution are Go code; a language model is called only at the moment a new issue is about to be published, and for ambiguous attribution only under the policy the operator selects); and the issue-to-repair-to-pull-request pipeline is confirmed development-repository-only and must not ship in deployed templates. The 0.1.0 draft (anonymity options, relay, bot identity) was never committed and is not recorded further.
- 0.1.0 (2026-10-04) — Superseded draft; not committed.

## 1. Overview

MoAI-ADK users hit bugs in the tool itself, and each user's experience is lost unless they write a report. This SPEC adds an opt-in participation mode: where a user consents, the tool detects its own malfunctions, decides by deterministic rules whether moai-adk (not the user, not the environment) is responsible, reduces the case to a fixed-field fingerprint, and files a public issue in the moai-adk repository from the user's own GitHub account. A second user hitting the same fingerprint adds an occurrence comment instead of a new issue.

Four properties bind everything below:

1. **Off by default, asked once, interactively only.** The question is asked at `moai init` and `moai update` only when a person can answer it. The consent text says plainly that the issue is public and tied to the user's GitHub account.
2. **moai responsibility only.** A report is filed only when an ordered, deterministic rule set attributes the cause to moai-adk. User code failures, user configuration errors, environment problems (network, permissions, disk, missing tools), and Claude Code tool failures are never reported.
3. **Machine-generated fixed fields only.** A payload carries a closed set of fields. No error text, panic text, user path, repository name, user name, host name, environment value, session id, user code, or prompt is read into it.
4. **Zero tokens by default.** Detection, fingerprinting, duplicate checking, masking, frequency caps, and deterministic attribution are Go code and spend no language-model tokens. A model is called at most once per newly published issue, never for a duplicate fingerprint, never when a cap is hit, and never for a non-moai verdict.

The issue-to-auto-repair-to-pull-request pipeline that will consume these reports is a separate, development-repository-only task; it is not part of this SPEC and must not be added to deployed templates.

## 2. Relationship to existing SPECs

The new SPEC extends the completed feedback subsystem and amends one recorded wizard decision. No file of a completed SPEC is edited; the amendments below are recorded here and in the run-phase commit messages.

| Existing requirement | Disposition | Detail |
|---|---|---|
| SPEC-FEEDBACK-AUTO-SUBMIT-001 REQ-1, REQ-2 (`auto_submit`, confirmation gate) | Keep | They govern the manual `/moai feedback` flow only. `auto_submit` still means "skip the confirmation gate"; it never means "participate". The two keys are independent. |
| REQ-3 to REQ-6 (`moai feedback scrub`, secret patterns, home-path collapse, env masking) | Keep | The new pipeline reuses `feedback.Scrub` unchanged as a tripwire over every outbound payload. |
| REQ-7 (vulnerability classification, never a public issue) | Keep and extend | The classifier runs on every automatic payload; a blocked result stays local. A path-traversal-derived report is also withheld (REQ-ANON-012). |
| REQ-8 (mask log) | Keep | Automatic payloads that trigger masking append to the same mask log. |
| REQ-9 (retry queue, D4 draft-versus-queue split) | Amend | The automatic path uses its own queue file and outbox log so user-authored text and machine-generated payloads never share a store. The existing queue's `Attempts` field, never incremented today, is incremented by the new sender. |
| REQ-10 (skill-body obligations) | Keep and extend | The three skill-body copies gain a participation description; the four existing HARD obligations stay. |
| REQ-11 (`feedback_auto_submit` init question) | Superseded in intent, not restored | SPEC-INIT-QUIET-WIZARD-001 removed that question and its pin tests keep it removed. This SPEC adds a different question, `feedback_participation`; the removed-set pins are unchanged. |
| REQ-12, REQ-13 (web toggle, Template-First mirror) | Keep and extend | The `feedback` section is already `RouteSeam`; a boolean row is added. |
| SPEC-INIT-QUIET-WIZARD-001 REQ-IQW-002 and the init question-set pins | Amend | Re-adding a question to the init wizard amends the quiet-wizard decision. REQ-ANON-003 names the pin tests that change. The least-invasive slot follows the Jev opt-in precedent (SPEC-JEV-OPTIN-MEASURE-001). |

## 3. Requirements (GEARS)

Five requirement modules.

### Module A — Consent and the participation setting

- **REQ-ANON-001** (Ubiquitous) — The feedback configuration shall carry a boolean `feedback.participation` and a boolean `feedback.participation_asked`, both shipped with the value `false`, in the template section file, the local mirror, and every settings fixture that mirrors the shipped file.
- **REQ-ANON-002** (State-driven) — While `feedback.participation` is false or absent, the bug-report pipeline shall not capture, record, queue, or transmit any signal, shall make no network request, and shall invoke no language model.
- **REQ-ANON-003** (Event-driven) — When `moai init` runs its interactive wizard, the wizard shall present exactly one participation confirmation question, defaulting to no, in its own group placed after the Jev opt-in question, absent from the shared and reconfigure question sets; the answer shall be persisted through the shared `settings.ApplySchemaEdits` seam only when the wizard actually ran and the `CI` environment variable is empty; and the change shall update the pin tests `TestInitQuestions_QuietSet`, `TestInitStepper_Denominator4`, `TestInitRegroup_SecondGroupGolden`, and the question-count assertion in `internal/cli/wizard/expansion_test.go`, leave the removed-question set unchanged, and record the amendment of the quiet-wizard decision in its commit message.
- **REQ-ANON-004** (Event-driven) — When `moai update` finishes its template synchronisation in an interactive terminal, with `feedback.participation_asked` false and `feedback.participation` false, the command shall ask the participation question once and persist both keys; and the command shall not ask when stdin is not a terminal, when the `CI` environment variable is non-empty, when `--yes` is given, or when `--check`, `--binary`, `--dry-run`, `--restore`, or `--config` selects another mode; consent shall never be inferred or auto-confirmed.
- **REQ-ANON-005** (Ubiquitous) — The participation question text shall, in each of the four wizard locales (en, ko, ja, zh), state that the filed issue is public and tied to the user's GitHub account, enumerate the fixed fields sent, name the preview command and the `moai web` settings screen, and state that an issue already filed cannot be recalled by the tool.

### Module B — Detection, attribution, and fingerprint

- **REQ-ANON-006** (Ubiquitous) — The pipeline shall define a closed enumeration of error kinds (`panic`, `hook_handler_failure`, `hook_timeout`, `internal_error`, `template_deploy_failure`, `harness_defect`), each mapped in `design.md` to a default attribution verdict and to the exact code sites that emit it, shall accept a signal only from a registered site, and shall ensure every recovered-panic site in `internal/`, `cmd/`, and `pkg/` either calls the capture entry point or appears on a reasoned allowlist.
- **REQ-ANON-007** (Event-detected unwanted) — When a candidate signal arises from the execution of user code (a user test, lint, or build run), from the Claude Code tool-failure handlers, or from a user configuration error, the pipeline shall discard it without recording it.
- **REQ-ANON-008** (Ubiquitous) — The capture entry point shall be fail-open: it shall never return an error to its caller, never panic, never perform network input or output, never invoke a language model, abandon its work after a fixed time box, and append to a bounded local spool only.
- **REQ-ANON-009** (Ubiquitous) — The pipeline shall assign every captured signal one verdict of `moai`, `user`, `environment`, or `ambiguous` by an ordered, first-match, deterministic rule set evaluated in Go (environmental error classes first, user-configuration error classes second, kind-specific moai rules third, `ambiguous` for the remainder), shall continue only for `moai`, shall keep `user` and `environment` verdicts local without any further step, and shall handle `ambiguous` according to a single policy constant that selects either one bounded model adjudication or local retention without a model call.
- **REQ-ANON-010** (Ubiquitous) — The fingerprint shall be derived solely from the moai version, the build commit, the operating system and architecture, the error kind, and the ordered function names of moai-internal stack frames, where a frame is moai-internal when its function name begins with the module path `github.com/modu-ai/moai-adk/`; file paths, line numbers, arguments, and non-moai frames shall never enter the fingerprint or the payload, and a panic that yields no moai-internal frame shall stay local.

### Module C — Local pipeline and payload minimisation

- **REQ-ANON-011** (Ubiquitous) — The report payload shall consist only of the fixed fields `schema`, `kind`, `fingerprint`, `version`, `commit`, `os`, `arch`, `frames`, and an optional constrained `detail`, each validated against an anchored allowlist, and its construction shall accept no error value, panic value, or free-form string.
- **REQ-ANON-012** (Event-detected unwanted) — When the rendered title or body of a payload produces a `blocked` verdict from the existing classifier or any masking finding from the existing scrubber, or when the report derives from a path-traversal sentinel, the pipeline shall withhold it: it shall not be queued, shall not be sent, shall trigger no model call, and shall be recorded in the local outbox log as withheld.
- **REQ-ANON-013** (Ubiquitous) — The pipeline shall keep a local ledger and shall not queue a fingerprint already queued or sent within the per-fingerprint window, shall not queue more reports than the global caps per rolling day and per rolling week, shall bound the queue length, and shall drop an item after the attempt limit, with every constant defined in `internal/config/defaults.go`.
- **REQ-ANON-014** (Event-driven) — When the user runs the preview command, the system shall print exactly the bytes the sender would hand to `gh` for each queued item without any network request, and the pipeline shall append every queued, sent, withheld, deduplicated, capped, dropped, or discarded outcome to a local append-only outbox log readable only by the user.

### Module D — Publication, occurrence counting, and the model-call budget

- **REQ-ANON-015** (State-driven) — While `feedback.participation` is true, the sender, implemented as Go code that invokes the user's own `gh` CLI, shall re-read the setting before each item and skip every item when it is false, shall hand only a validated payload to its single publication seam, shall never run inside the synchronous path of a hook event, shall finish within a fixed time box, and shall leave an item queued without prompting the user when `gh` is missing or unauthenticated.
- **REQ-ANON-016** (Event-driven) — When the sender finds, by an exact title-key match among the results of a `gh` issue search for the fingerprint, an existing issue (open or closed) for the same fingerprint, the sender shall add one occurrence comment to that issue and shall invoke no language model.
- **REQ-ANON-017** (Event-driven) — When the sender finds no issue for the fingerprint and the verdict is `moai`, the sender shall invoke a language model at most once to write the issue summary from the validated payload fields alone, shall re-validate the model output through the scrubber, the classifier, and a length and character allowlist, and shall fall back to a deterministic Go template text when the model is unavailable, fails, or returns an output that fails validation, so that the issue is filed either way.
- **REQ-ANON-018** (Unwanted) — The pipeline shall not invoke a language model when participation is off, when the verdict is `user` or `environment`, when the verdict is `ambiguous` under the local-retention policy, when the fingerprint is deduplicated or capped locally, when the payload is withheld, when an issue for the fingerprint already exists, or when a rolling daily model-call cap is reached.
- **REQ-ANON-019** (Ubiquitous) — The filed issue shall use the title `[auto-report] <kind> <fingerprint>` and shall begin its body with a machine-readable block `<!-- moai-bugreport:v1 ... -->` carrying the fields `schema`, `fingerprint`, `kind`, `version`, `commit`, `os_arch`, and `frames`; each occurrence comment shall consist of a machine-readable block `<!-- moai-bugreport:occurrence v1 ... -->` carrying the fingerprint, version, commit, and `os_arch`; the issue count a consumer computes shall be one plus the number of occurrence comments; and the client shall not edit an issue body or attach labels.

### Module E — Console, withdrawal, mirrors, and shipping boundary

- **REQ-ANON-020** (Event-driven) — When the user opens the web settings screen, the console shall render `feedback.participation` as a boolean toggle (never a free-text field) in the `feedback` section with titles and descriptions in all four locales, the description stating that filed issues are public and tied to the user's GitHub account, persist changes through `settings.ApplySchemaEdits`, and not render `feedback.participation_asked`.
- **REQ-ANON-021** (Event-driven) — When a flush starts with `feedback.participation` false, the pipeline shall discard every unsent queue item and the capture spool and record each discard in the outbox log; and when the user runs the purge command, the system shall remove the queue, the spool, the ledger, and the outbox log; neither action shall make a network request or invoke a language model.
- **REQ-ANON-022** (Ubiquitous) — The change set shall mirror both new keys into the template section file, register each in the shipped-key inventory with a production reader, correct the three skill-body copies of `feedback.md`, and replace the statement that an issue is "created automatically" with an accurate statement in the four docs-site locale pages of `moai-feedback.md`.
- **REQ-ANON-023** (Unwanted) — The deployed template tree (`internal/template/templates/`) and the plugin tree (`plugins/moai/`) shall not contain any auto-repair artifact: no file whose path contains `auto-repair`, `autorepair`, or `auto_repair`, and no file whose content contains the issue-contract marker `moai-bugreport`.

## 4. Acceptance

The acceptance criteria are in `acceptance.md` (25 criteria, each mapped to a requirement, each with one verification command). Given/When/Then scenarios are the verification layer; the requirements above are the GEARS layer.

## 5. Constraints

- Template comments in shipped YAML stay neutral: no SPEC identifiers and no requirement tokens.
- Instruction documents and skill bodies stay English; wizard question text carries en, ko, ja, and zh.
- Local verification is scoped to the packages the change touches; the full suite is CI's.
- No time estimates appear in any artifact; ordering and priority labels only.
- The transport is the user's own `gh`; no relay, bot identity, shipped credential, or anonymity mechanism is introduced.

## 6. Exclusions

### Out of Scope — Issue to auto-repair to pull request pipeline
- The consumer that reads the issue contract and produces a repair pull request is a separate, development-repository-only task. It is not built here and is never shipped in deployed templates (REQ-ANON-023 guards the shipping boundary).

### Out of Scope — User project bugs and user content
- Failures of user code (tests, lint, build), user configuration errors, environment problems (network, permissions, disk, missing tools), user prompts, user project content, and Claude Code tool failures are never captured (REQ-ANON-007, REQ-ANON-009). No telemetry beyond tool bug reports is added.

### Out of Scope — Anonymous reporting
- No relay service, bot account, GitHub App, shipped credential, or non-GitHub intake is built. Participation is real-name by design.

### Out of Scope — Changing the vulnerability policy
- `SECURITY.md` and the classifier vocabulary are not changed. Vulnerability-class reports stay local, exactly as the existing classifier decides.

### Out of Scope — Free-text bug reports from automatic detection
- Automatic reports never carry user-written text. Free-text reporting remains the manual `/moai feedback` flow, whose confirmation gate and scrubber are unchanged.

## 7. Cross-References

- `.moai/specs/SPEC-FEEDBACK-AUTO-SUBMIT-001/` — the completed feedback subsystem this SPEC extends.
- `.moai/specs/SPEC-INIT-QUIET-WIZARD-001/` — the quiet-wizard decision amended by REQ-ANON-003.
- `.moai/specs/SPEC-JEV-OPTIN-MEASURE-001/` — the init-only opt-in slot and persistence precedent.
- `design.md` — kind-to-verdict-to-site register, attribution rules, payload schema, publication flow, model-call budget.
- `research.md` — verified premises and the note on dropped anonymity options.
