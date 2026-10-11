# spec-compact — SPEC-FEEDBACK-PARTICIPATION-001 (version 0.5.1)

Run-phase digest: requirements, acceptance map, files, exclusions. Full text in `spec.md` and `acceptance.md`.

## Requirements (GEARS)

Module A — consent
- REQ-ANON-001 Consent lives only in a user-scoped file under the user's moai home (`participation.enabled`, `participation.asked`, optional `participation.repository`); absent or unreadable reads false; no key ships in templates or mirrors.
- REQ-ANON-002 While the user-scoped value is false or absent: no capture, record, queue, drain, or transmit; no network; no model (the withdrawal actions of REQ-ANON-021 are excepted — they discard and log local state only).
- REQ-ANON-003 When init runs its wizard: one confirm question after Jev, own group, absent from shared and reconfigure sets, default no unless the stored value is true; persisted only when the wizard ran and `CI` is empty.
- REQ-ANON-004 When a plain template-sync `moai update` finishes interactively and `asked` is false: ask once, default no, persist both values (empty answer persists enabled false, asked true); never under non-terminal, `CI`, or any mode flag; no auto-consent.
- REQ-ANON-005 Consent text in en, ko, ja, zh: public and account-tied; creation time public; automatic, no per-report confirmation; filer subscribed to later comments; account publicly associated with using moai-adk at that version and OS; fields; preview and `moai web`; no recall; summary may spend the user's own model tokens.

Module B — detection and attribution
- REQ-ANON-006 Closed kind enum (6) with verdict, derivation mode, registered sites; every recover site reports or is allowlisted; coverage is the main goroutine plus registered sites, other goroutines are a recorded gap.
- REQ-ANON-007 Discard user-code, tool-failure, and user-config signals.
- REQ-ANON-008 Capture is fail-open, bounded, network-free, model-free; the spool lives under the user's moai home, never under the project tree (D36).
- REQ-ANON-009 Ordered first-match attribution, environment then user, evaluated at capture while the error chain is alive; `moai` only for a panic, an internal marker, or an enumerated moai sentinel; everything else (including an unmarked hook handler error) is user, environment, or ambiguous; only `moai` continues; ambiguous is retained locally — no model call, nothing queued (out of scope: the former adjudication variant, DEC-7).
- REQ-ANON-010 Fingerprint from version, commit, os/arch, kind, moai-internal frame names (module-prefix filter, no paths or lines).

Module C — local pipeline
- REQ-ANON-011 Closed-schema payload; `detail` is closed-set membership in a dedicated type (template token, or registered event plus registered handler name), re-validated on read-back; no error, panic, or free-string input.
- REQ-ANON-012 Scrub and classify tripwire; block, masking finding, or path traversal means withheld.
- REQ-ANON-013 Local ledger, per-fingerprint window, global caps, queue bound, attempt limit.
- REQ-ANON-014 Preview prints the queued payload's rendered bytes (the bytes the create path hands to `gh` for a new issue); append-only 0600 outbox log.

Module D — publication
- REQ-ANON-015 Go sender over the user's own `gh`; consent re-read per item; target is the compiled default repository unless the user-scoped file names another; never in a hook dispatch; time-boxed; quiet when `gh` is absent.
- REQ-ANON-016 Existing fingerprint issue: one occurrence comment, none at the per-issue cap (advisory count from `gh` JSON); no model call.
- REQ-ANON-017 New moai issue: one model call through the seam at most per queue item, a durable `summary_requested` marker persisted before the call (recovery with a marker but no summary takes the template text; the queue lock is owner-verified and stale-breakable, so a crash cannot wedge the queue, D37), validated summary persisted before create and reused on retry, template fallback.
- REQ-ANON-018 No model call in any excluded case (including `ambiguous`) or at the daily cap; across retries at most one summary call per queue item.
- REQ-ANON-019 Issue contract: title `[auto-report] <kind> <fingerprint>`, body marker block, occurrence marker, count = 1 + comments (advisory), markers untrusted (consumers re-derive from the title key), no body edit, no labels.

Module E — console and boundary
- REQ-ANON-020 Web toggle is the existing radio pair with its present-companion, user-scoped read and write, `asked` set on change and never rendered, four-locale text carrying the full REQ-ANON-005 disclosure (console-first enablement is informed).
- REQ-ANON-021 Off-state flush discards unsent items; purge removes all local state.
- REQ-ANON-022 Correct the three skill-body copies and four docs pages; update every init pin test the question affects; record the quiet-wizard amendment.
- REQ-ANON-023 No auto-repair artifact in `internal/template/templates/` or `plugins/moai/`; the wizard question never ships where the sender package is absent.

Module F — cross-cutting
- REQ-ANON-024 A participation or repository value in any project-tier file is ignored by the automatic pipeline; the sender targets the compiled default repository unless the user-scoped file names another.
- REQ-ANON-025 One model seam: no `os/exec` or `net/http` in bugreport or outbox, no `net/http` in publish and `os/exec` only in its `gh` runner, production implementation injected by the CLI.

## Acceptance map (25)

AC-001 R1+R24, AC-002 R2 (capture)+R8, AC-003 R2 (drain, sender)+R24, AC-004 R3, AC-005 R4, AC-006 R5, AC-007 R6+R7, AC-008 R9, AC-009 R9 (ambiguous retained locally, M4), AC-010 R10, AC-011 R11, AC-012 R12, AC-013 R13, AC-014 R14, AC-015 R15, AC-016 R16, AC-017 R17, AC-018 R17+R18, AC-019 R18, AC-020 R19+R14, AC-021 R20, AC-022 R21, AC-023 R22+R1, AC-024 R23, AC-025 R25. The old-to-new numbering map is in `acceptance.md`.

## Files (by milestone)

M1 `internal/bugreport/{kind,verdict,payload,validate,detail,fingerprint,frames,marker}.go`; M2 user-scoped reader in `internal/config`, writer and `PersistUserScoped` in `internal/settings`, web predicate and i18n, wizard, init and update files per `plan.md` §C (no template key, no inventory row); M3 capture (with capture-time attribution), spool (user-scoped store under `<moai home>/state/bugreport/`, D36), attribution, register, `cmd/moai/main.go`, the recover sites the design walk enumerates (non-exhaustive list; the walk is authoritative), hook registry and handler-name registration, preference cmd, update sync and clean-install; M4 `internal/feedback/outbox/`, the owner-verified stale-lock break in `queue.go` `Mutate` (D37), `internal/cli/feedback_participation.go`; M5 `internal/feedback/publish/` and golden; M6 model seam, `summary_requested` marker, budget, CLI production implementation, static guards; M7 three `feedback.md` copies, four docs pages, `internal/template/auto_repair_guard_test.go`, wizard coexistence guard.

## Exclusions

- Issue to auto-repair to pull request pipeline (development-repository-only; never shipped in templates).
- User project bugs, user config errors, environment problems, Claude Code tool failures, user prompts and content.
- Anonymous reporting (no relay, bot, App, shipped credential, non-GitHub intake).
- Team-level or project-level consent.
- Behavioural defects of deployed agents and skills (narrowing accepted, DEC-8; a follow-up SPEC may add a deployed-asset self-check).
- Model adjudication of ambiguous attribution (retention is the shipped policy, DEC-7; a follow-up SPEC may add the variant).
- Panics on goroutines other than the main goroutine.
- Changing the vulnerability policy or classifier vocabulary.
- Free-text reports from automatic detection.
