---
id: SPEC-FEEDBACK-ANON-PARTICIPATION-001
title: "Opt-in improvement participation — automatic filing of moai-adk-attributed tool bugs from the user's own gh account, with user-scoped consent, allowlist attribution, and a zero-token default path"
version: "0.3.0"
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
tags: "feedback, participation, opt-in, attribution, fingerprint, token-budget, privacy, wizard, web-console, user-scoped-consent"
---

# SPEC-FEEDBACK-ANON-PARTICIPATION-001

## HISTORY

- 0.3.0 (2026-10-04) — Revision after plan-audit iteration 1 (score 0.62, FAIL). D2: consent moved out of the tracked project file into a user-scoped file under the user's moai home directory; a project-tier value never enables anything and the sender's repository is the compiled default (REQ-001, REQ-024, AC-001, AC-003). D3, D14: attribution became an allowlist, a hook handler error without a marker is `ambiguous`, exec-exit, JSON and YAML decode errors joined the non-moai rows, and every register row now states whether it is derived from an importable sentinel or asserted by the call site (REQ-009, AC-008, design section 2 and 3). D4: `detail` became closed-set membership in an enum type (REQ-011, AC-011). D5: every Verify line requires a `--- PASS: <Name>` line per named test (acceptance preface). D6: the console toggle is the existing radio pair with its present-companion, and `participation_asked` is stored in the user-scoped file and is not a schema field (REQ-020, AC-021). D7: the false premise that no text-generation call exists was corrected; the model decision is recorded (design section 8, plan Decisions). D8: pin-test blast radius enumerated by search (plan M2, AC-004). D9: AC-023 gained commands for the skill bodies, the mirror, and a positive per-locale docs pattern. D10: the off-means-nothing criterion was split by owning milestone (AC-002, AC-003). D11: defaults and the update prompt default are asserted (AC-001, AC-005). D12: the summary is persisted before create and the per-item call bound is stated (REQ-017, REQ-018, AC-018). D13: a static reachability guard was added (REQ-025, AC-025). D15: panic coverage is stated as the main goroutine plus registered recover sites, with the gap recorded. D16: the consent text gained five statements, the sender skips commenting past a per-issue cap, and occurrence markers are untrusted input (REQ-005, REQ-016, REQ-019). D17: test names left the requirement text; REQ-013 and REQ-022 stay bundled because the 25-requirement ceiling is full. D18: raw RED output is pasted with the proxy rule stated. D19: the narrowing of "deployed harness defect" is a recorded scope decision awaiting operator acceptance. D20: the update prompt asks only on a plain template-sync run, with the flag inventory verified from `update.go`. D21: a coexistence guard was added (REQ-023, AC-024). D1: the unresolved clarifications were reduced to two operator-owned ones in plan.md; the rest became recorded decisions. D22 and D23 are recorded in progress.md and plan M5. Acceptance criteria were renumbered; the old-to-new map is in acceptance.md.
- 0.2.0 (2026-10-04) — Rewritten after the operator corrected the card scope: participation is real-name (the issue is filed from the participating user's own `gh` account), so the anonymity design was dropped entirely; automatic filing happens only when a deterministic attribution rule assigns the cause to moai-adk; token minimisation became a requirement (all detection, fingerprinting, duplicate checking, masking, caps, and deterministic attribution are Go code; a language model is called only at the moment a new issue is about to be published, and for ambiguous attribution only under the policy the operator selects); and the issue-to-repair-to-pull-request pipeline is confirmed development-repository-only and must not ship in deployed templates. The 0.1.0 draft (anonymity options, relay, bot identity) was never committed and is not recorded further.
- 0.1.0 (2026-10-04) — Superseded draft; not committed.

## 1. Overview

MoAI-ADK users hit bugs in the tool itself, and each user's experience is lost unless they write a report. This SPEC adds an opt-in participation mode: where a user consents, the tool detects its own malfunctions, decides by deterministic rules whether moai-adk (not the user, not the environment) is responsible, reduces the case to a fixed-field fingerprint, and files a public issue in the moai-adk repository from the user's own GitHub account. A second user hitting the same fingerprint adds an occurrence comment instead of a new issue.

Four properties bind everything below:

1. **Off by default, asked once, interactively only, and user-scoped.** The question is asked at `moai init` and `moai update` only when a person can answer it. The consent lives in a file under the user's own moai home directory, outside every project tree: a value that arrives in a tracked project file never enables capture or sending, and a repository target arriving in a project file never redirects the user's account. The consent text says plainly that the issue is public, tied to the user's GitHub account, filed without per-report confirmation, and may spend the user's own model tokens.
2. **moai responsibility only, by allowlist.** A report is filed only when an ordered, deterministic rule set attributes the cause to moai-adk, and `moai` is an allowlist: only a panic, an explicit internal marker, or an enumerated moai sentinel qualifies. User code failures, user configuration errors, environment problems (network, permissions, disk, missing tools, subprocess exits), Claude Code tool failures, and every unrecognised error are never reported as moai's.
3. **Machine-generated fixed fields only.** A payload carries a closed set of fields. No error text, panic text, user path, repository name, user name, host name, environment value, session id, user code, or prompt is read into it, and the one optional field is a member of a closed set.
4. **Zero tokens by default.** Detection, fingerprinting, duplicate checking, masking, frequency caps, and deterministic attribution are Go code and spend no language-model tokens. A model is reached through one seam, called at most once per queue item for the issue summary (the validated result is stored and reused across retries) and, under the adjudication policy, at most once per queue item for an adjudication; never for a duplicate fingerprint, never when a cap is hit, and never for a non-moai verdict.

The issue-to-auto-repair-to-pull-request pipeline that will consume these reports is a separate, development-repository-only task; it is not part of this SPEC and must not be added to deployed templates.

## 2. Relationship to existing SPECs

The new SPEC extends the completed feedback subsystem and amends one recorded wizard decision. No file of a completed SPEC is edited; the amendments below are recorded here and in the run-phase commit messages.

| Existing requirement | Disposition | Detail |
|---|---|---|
| SPEC-FEEDBACK-AUTO-SUBMIT-001 REQ-1, REQ-2 (`auto_submit`, confirmation gate) | Keep | They govern the manual `/moai feedback` flow only. `auto_submit` still means "skip the confirmation gate"; it never means "participate". The project-tier `feedback.repository` also keeps serving the manual flow only; the automatic pipeline ignores it (REQ-ANON-024). |
| REQ-3 to REQ-6 (`moai feedback scrub`, secret patterns, home-path collapse, env masking) | Keep | The new pipeline reuses `feedback.Scrub` unchanged as a tripwire over every outbound payload. |
| REQ-7 (vulnerability classification, never a public issue) | Keep and extend | The classifier runs on every automatic payload; a blocked result stays local. A path-traversal-derived report is also withheld (REQ-ANON-012). |
| REQ-8 (mask log) | Keep | Automatic payloads that trigger masking append to the same mask log. |
| REQ-9 (retry queue, D4 draft-versus-queue split) | Amend | The automatic path uses its own queue file and outbox log so user-authored text and machine-generated payloads never share a store. The existing queue's `Attempts` field, never incremented today, is incremented by the new sender. |
| REQ-10 (skill-body obligations) | Keep and extend | The three skill-body copies gain a participation description; the four existing HARD obligations stay. |
| REQ-11 (`feedback_auto_submit` init question) | Superseded in intent, not restored | SPEC-INIT-QUIET-WIZARD-001 removed that question and its pin tests keep it removed. This SPEC adds a different question, `feedback_participation`; the removed-set pins are unchanged. |
| REQ-12, REQ-13 (web toggle, Template-First mirror) | Keep and extend | The `feedback` section is already `RouteSeam`; one boolean row is added whose value persists to the user-scoped file rather than to the project section file. |
| SPEC-INIT-QUIET-WIZARD-001 REQ-IQW-002 and the init question-set pins | Amend | Re-adding a question to the init wizard amends the quiet-wizard decision. REQ-ANON-022 requires every pin test the question affects to be updated; plan.md lists them. The least-invasive slot follows the Jev opt-in precedent (SPEC-JEV-OPTIN-MEASURE-001). |

## 3. Requirements (GEARS)

Six requirement modules.

### Module A — Consent and the participation setting

- **REQ-ANON-001** (Ubiquitous) — The participation setting shall be stored only in a user-scoped file under the user's moai home directory, outside every project tree, holding the booleans `participation.enabled` and `participation.asked` and an optional `participation.repository`; both booleans shall read false when the file, a key, or the home directory is absent or unreadable; and the shipped template section files and their local mirrors shall carry none of these keys.
- **REQ-ANON-002** (State-driven) — While the user-scoped `participation.enabled` value is false or absent, the bug-report pipeline shall not capture, record, queue, drain, or transmit any signal, shall make no network request, and shall invoke no language model.
- **REQ-ANON-003** (Event-driven) — When `moai init` runs its interactive wizard, the wizard shall present exactly one participation confirmation question, in its own group placed after the Jev opt-in question and absent from the shared and reconfigure question sets, defaulting to no unless the user-scoped value is already true; and the answer shall be persisted only when the wizard actually ran and the `CI` environment variable is empty.
- **REQ-ANON-004** (Event-driven) — When a plain template-synchronisation run of `moai update` finishes in an interactive terminal while the user-scoped `participation.asked` value is false, the command shall ask the participation question once, defaulting to no, and persist both values, an empty answer persisting `enabled` false and `asked` true; and the command shall not ask when stdin is not a terminal, when the `CI` environment variable is non-empty, or when any mode flag selects another operation; consent shall never be inferred or auto-confirmed.
- **REQ-ANON-005** (Ubiquitous) — The participation question text shall, in each of the four wizard locales (en, ko, ja, zh), state that the filed issue is public and tied to the user's GitHub account; that the issue and comment creation time is public; that filing is automatic with no per-report confirmation; that GitHub subscribes the filer to later comments on the issue; that the account becomes publicly associated with using moai-adk at that version and operating system; which fixed fields are sent; the preview command and the `moai web` settings screen; that an issue already filed cannot be recalled by the tool; and that composing the issue summary may spend the user's own model-subscription tokens, with a template text used when no model is available.

### Module B — Detection, attribution, and fingerprint

- **REQ-ANON-006** (Ubiquitous) — The pipeline shall define a closed enumeration of error kinds (`panic`, `hook_handler_failure`, `hook_timeout`, `internal_error`, `template_deploy_failure`, `harness_defect`), each mapped in `design.md` to a default attribution verdict, a derivation mode, and the exact code sites that emit it, shall accept a signal only from a registered site, and shall ensure every recovered-panic site in `internal/`, `cmd/`, and `pkg/` either calls the capture entry point or appears on a reasoned allowlist; panic coverage is the main goroutine plus the registered recover sites, and a panic on any other goroutine is outside coverage and recorded as a residual gap in `design.md`.
- **REQ-ANON-007** (Event-detected unwanted) — When a candidate signal arises from the execution of user code (a user test, lint, or build run), from the Claude Code tool-failure handlers, or from a user configuration error, the pipeline shall discard it without recording it.
- **REQ-ANON-008** (Ubiquitous) — The capture entry point shall be fail-open: it shall never return an error to its caller, never panic, never perform network input or output, never invoke a language model, abandon its work after a fixed time box, and append to a bounded local spool only.
- **REQ-ANON-009** (Ubiquitous) — The pipeline shall assign every captured signal one verdict of `moai`, `user`, `environment`, or `ambiguous` by an ordered, first-match, deterministic rule set evaluated in Go, environmental error classes first and user-input error classes second; `moai` shall be assigned only to a panic, to an explicit internal-marker wrapper, or to an enumerated moai sentinel, so that every other signal, including a hook handler error that carries no marker, is `user`, `environment`, or `ambiguous`; the pipeline shall continue only for `moai`, keep `user` and `environment` verdicts local without any further step, and handle `ambiguous` according to a single policy constant that selects either one bounded model adjudication or local retention without a model call.
- **REQ-ANON-010** (Ubiquitous) — The fingerprint shall be derived solely from the moai version, the build commit, the operating system and architecture, the error kind, and the ordered function names of moai-internal stack frames, where a frame is moai-internal when its function name begins with the module path `github.com/modu-ai/moai-adk/`; file paths, line numbers, arguments, and non-moai frames shall never enter the fingerprint or the payload, and a panic that yields no moai-internal frame shall stay local.

### Module C — Local pipeline and payload minimisation

- **REQ-ANON-011** (Ubiquitous) — The report payload shall consist only of the fixed fields `schema`, `kind`, `fingerprint`, `version`, `commit`, `os`, `arch`, `frames`, and an optional `detail`, each validated against an anchored allowlist; `detail` shall be a member of a closed set per kind (an enumerated token for template and harness kinds; a registered event together with a registered handler name for hook kinds), carried as a dedicated type and never as a free string, and re-validated by membership whenever a queued payload is read back; and the payload's construction shall accept no error value, panic value, or free-form string.
- **REQ-ANON-012** (Event-detected unwanted) — When the rendered title or body of a payload produces a `blocked` verdict from the existing classifier or any masking finding from the existing scrubber, or when the report derives from a path-traversal sentinel, the pipeline shall withhold it: it shall not be queued, shall not be sent, shall trigger no model call, and shall be recorded in the local outbox log as withheld.
- **REQ-ANON-013** (Ubiquitous) — The pipeline shall keep a local ledger and shall not queue a fingerprint already queued or sent within the per-fingerprint window, shall not queue more reports than the global caps per rolling day and per rolling week, shall bound the queue length, and shall drop an item after the attempt limit, with every constant defined in `internal/config/defaults.go`.
- **REQ-ANON-014** (Event-driven) — When the user runs the preview command, the system shall print exactly the bytes the sender would hand to `gh` for each queued item without any network request, and the pipeline shall append every queued, sent, withheld, deduplicated, capped, dropped, or discarded outcome to a local append-only outbox log readable only by the user.

### Module D — Publication, occurrence counting, and the model-call budget

- **REQ-ANON-015** (State-driven) — While the user-scoped `participation.enabled` value is true, the sender, implemented as Go code that invokes the user's own `gh` CLI, shall re-read that value before each item and skip every item when it is false, shall target the compiled default repository unless the user-scoped file names another, shall hand only a validated payload to its single publication seam, shall never run inside the synchronous path of a hook event, shall finish within a fixed time box, and shall leave an item queued without prompting the user when `gh` is missing or unauthenticated.
- **REQ-ANON-016** (Event-driven) — When the sender finds, by an exact title-key match among the results of a `gh` issue search for the fingerprint, an existing issue (open or closed) for the same fingerprint, the sender shall add one occurrence comment to that issue and shall invoke no language model, except that it shall add nothing when the issue already carries the per-issue occurrence-comment cap, a count it reads from the `gh` JSON output and treats as advisory.
- **REQ-ANON-017** (Event-driven) — When the sender finds no issue for the fingerprint and the verdict is `moai`, the sender shall obtain the issue summary through the single model seam at most once per queue item, shall persist the validated summary, or the decision to use the template text, on the queue item before it creates the issue and reuse it on every later attempt, shall re-validate model output through the scrubber, the classifier, and a length and character allowlist, and shall fall back to a deterministic Go template text when the model is unavailable, fails, or returns an output that fails validation, so that the issue is filed either way.
- **REQ-ANON-018** (Unwanted) — The pipeline shall not invoke a language model when participation is off, when the verdict is `user` or `environment`, when the verdict is `ambiguous` under the local-retention policy, when the fingerprint is deduplicated or capped locally, when the payload is withheld, when an issue for the fingerprint already exists, or when a rolling daily model-call cap is reached; and across every retry of one queue item and under either ambiguity policy it shall invoke a model at most once for the summary and at most once for an adjudication.
- **REQ-ANON-019** (Ubiquitous) — The filed issue shall use the title `[auto-report] <kind> <fingerprint>` and shall begin its body with a machine-readable block `<!-- moai-bugreport:v1 ... -->` carrying the fields `schema`, `fingerprint`, `kind`, `version`, `commit`, `os_arch`, and `frames`; each occurrence comment shall consist of a machine-readable block `<!-- moai-bugreport:occurrence v1 ... -->` carrying the fingerprint, version, commit, and `os_arch`; the issue count a consumer computes shall be one plus the number of occurrence comments and is advisory; occurrence markers shall be documented as untrusted input, so that a consumer re-derives every field from the title key and the issue body block and ignores a marker field that disagrees; and the client shall not edit an issue body or attach labels.

### Module E — Console, withdrawal, mirrors, and shipping boundary

- **REQ-ANON-020** (Event-driven) — When the user opens the web settings screen, the console shall render `feedback.participation` as the console's existing two-option radio pair with its present-companion, never as a checkbox or a free-text field, in the `feedback` section with titles and descriptions in all four locales, the description stating that filed issues are public and tied to the user's GitHub account; the console shall read and write the user-scoped value rather than the project tier, shall set `participation.asked` true whenever it changes `participation.enabled`, and shall not render `participation.asked`.
- **REQ-ANON-021** (Event-driven) — When a flush starts with the user-scoped `participation.enabled` value false, the pipeline shall discard every unsent queue item and the capture spool and record each discard in the outbox log; and when the user runs the purge command, the system shall remove the queue, the spool, the ledger, and the outbox log; neither action shall make a network request or invoke a language model.
- **REQ-ANON-022** (Ubiquitous) — The change set shall correct the three skill-body copies of `feedback.md` to describe participation, replace the statement that an issue is "created automatically" with an accurate statement in the four docs-site locale pages of `moai-feedback.md`, update every init-wizard pin test that the new question affects while leaving the removed-question set unchanged, and record the amendment of the quiet-wizard decision in the commit message of the milestone that adds the question.
- **REQ-ANON-023** (Unwanted) — The deployed template tree (`internal/template/templates/`) and the plugin tree (`plugins/moai/`) shall not contain any auto-repair artifact: no file whose path contains `auto-repair`, `autorepair`, or `auto_repair`, and no file whose content contains the issue-contract marker `moai-bugreport`; and the wizard question set shall not carry the participation question in a tree where the sender package is absent.

### Module F — Cross-cutting invariants

- **REQ-ANON-024** (Event-detected unwanted) — When a project-tier file (any file under the project's `.moai/config/` tree) carries a `participation` key, a `participation_asked` key, or a feedback repository value, the automatic pipeline shall ignore those values: they shall neither enable nor disable capture, queuing, or sending, and the repository the sender targets shall be the compiled default unless the user-scoped file names another.
- **REQ-ANON-025** (Ubiquitous) — The pipeline shall reach a language model only through one model seam: the bug-report capture package and the outbox package shall import neither `os/exec` nor `net/http` nor any CLI package; the publication package shall import `net/http` nowhere and `os/exec` only in its `gh` runner file; and the seam's production implementation shall be injected by the CLI.

## 4. Acceptance

The acceptance criteria are in `acceptance.md` (25 criteria, each mapped to a requirement, each with named verification commands). Given/When/Then scenarios are the verification layer; the requirements above are the GEARS layer.

## 5. Constraints

- Template comments in shipped YAML stay neutral: no SPEC identifiers and no requirement tokens.
- Instruction documents and skill bodies stay English; wizard question text carries en, ko, ja, and zh.
- Local verification is scoped to the packages the change touches; the full suite is CI's.
- No time estimates appear in any artifact; ordering and priority labels only.
- The transport is the user's own `gh`; no relay, bot identity, shipped credential, or anonymity mechanism is introduced.
- No consent value is stored in, or read from, a tracked file; every test that reaches the user-scoped store sets `MOAI_HOME` to a temporary directory so the real home directory is never touched.

## 6. Exclusions

### Out of Scope — Issue to auto-repair to pull request pipeline
- The consumer that reads the issue contract and produces a repair pull request is a separate, development-repository-only task. It is not built here and is never shipped in deployed templates (REQ-ANON-023 guards the shipping boundary).

### Out of Scope — User project bugs and user content
- Failures of user code (tests, lint, build), user configuration errors, environment problems (network, permissions, disk, missing tools), user prompts, user project content, and Claude Code tool failures are never captured (REQ-ANON-007, REQ-ANON-009). No telemetry beyond tool bug reports is added.

### Out of Scope — Anonymous reporting
- No relay service, bot account, GitHub App, shipped credential, or non-GitHub intake is built. Participation is real-name by design.

### Out of Scope — Team-level or project-level consent
- A repository, team, or organisation cannot grant participation on a user's behalf: a project file carrying a participation value is ignored (REQ-ANON-024). Consent is per user.

### Out of Scope — Behavioural defects of deployed agents and skills
- The signal "deployed harness defect" is narrowed to the shipped-template render and validate sentinels, because a behavioural defect of an agent or skill has no machine-detectable signal in the tree. The narrowing is a recorded scope decision awaiting operator acceptance (plan.md section B).

### Out of Scope — Panics on goroutines other than the main goroutine
- Wrapping every goroutine entry point with a capture call is not built here; the gap is recorded in `design.md` and in the residual-risk notes (REQ-ANON-006).

### Out of Scope — Changing the vulnerability policy
- `SECURITY.md` and the classifier vocabulary are not changed. Vulnerability-class reports stay local, exactly as the existing classifier decides.

### Out of Scope — Free-text bug reports from automatic detection
- Automatic reports never carry user-written text. Free-text reporting remains the manual `/moai feedback` flow, whose confirmation gate and scrubber are unchanged.

## 7. Cross-References

- `.moai/specs/SPEC-FEEDBACK-AUTO-SUBMIT-001/` — the completed feedback subsystem this SPEC extends.
- `.moai/specs/SPEC-INIT-QUIET-WIZARD-001/` — the quiet-wizard decision amended by REQ-ANON-003 and REQ-ANON-022.
- `.moai/specs/SPEC-JEV-OPTIN-MEASURE-001/` — the init-only opt-in slot and persistence precedent.
- `design.md` — user-scoped store, kind-to-verdict-to-site register with derivation mode, attribution allowlist, payload schema, publication flow, model seam and budget.
- `research.md` — verified premises, the corrected model-call premise, and the note on dropped anonymity options.
