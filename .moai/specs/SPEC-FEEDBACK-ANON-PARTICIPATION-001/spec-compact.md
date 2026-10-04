# spec-compact — SPEC-FEEDBACK-ANON-PARTICIPATION-001

Run-phase digest: requirements, acceptance map, files, exclusions. Full text in `spec.md` and `acceptance.md`.

## Requirements (GEARS)

Module A — consent
- REQ-ANON-001 keys `feedback.participation` and `feedback.participation_asked`, default false, all mirrors.
- REQ-ANON-002 While participation is off: no capture, no network, no model call.
- REQ-ANON-003 When init runs its wizard: one confirm question after Jev, default no, absent from shared and reconfigure sets, persisted via `settings.ApplySchemaEdits` only if the wizard ran and `CI` is empty; update the named pin tests; record the amendment.
- REQ-ANON-004 When update finishes template sync interactively and nothing was asked yet: ask once; never when non-interactive, CI, `--yes`, or another mode; no auto-consent.
- REQ-ANON-005 Consent text in en, ko, ja, zh: public, tied to the user's GitHub account, fields, preview, `moai web`, no recall.

Module B — detection and attribution
- REQ-ANON-006 closed kind enum (6), registered emit sites, every recover site reports or is allowlisted.
- REQ-ANON-007 discard user-code, tool-failure, and user-config signals.
- REQ-ANON-008 capture is fail-open, bounded, network-free, model-free.
- REQ-ANON-009 deterministic ordered attribution (environment, user, moai kind rules, ambiguous); only `moai` continues; ambiguous follows one policy constant (local or one adjudication).
- REQ-ANON-010 fingerprint from version, commit, os/arch, kind, moai-internal frame names (module-prefix filter, no paths or lines).

Module C — local pipeline
- REQ-ANON-011 closed-schema payload; no error, panic, or free-string input.
- REQ-ANON-012 scrub and classify tripwire; block, masking finding, or path traversal means withheld.
- REQ-ANON-013 local ledger, per-fingerprint window, global caps, queue bound, attempt limit.
- REQ-ANON-014 preview prints the exact bytes handed to `gh`; append-only 0600 outbox log.

Module D — publication
- REQ-ANON-015 Go sender over the user's own `gh`; consent re-read per item; never in a hook dispatch; time-boxed; quiet when `gh` is absent.
- REQ-ANON-016 existing fingerprint issue: occurrence comment, no model call.
- REQ-ANON-017 new moai issue: at most one model call, validated output, template fallback.
- REQ-ANON-018 no model call in any excluded case or when the daily cap is reached.
- REQ-ANON-019 issue contract: title `[auto-report] <kind> <fingerprint>`, body marker block, occurrence comment marker, count = 1 + comments, no body edit, no labels.

Module E — console and boundary
- REQ-ANON-020 web bool toggle with four-locale text; marker key not rendered.
- REQ-ANON-021 off-state flush discards unsent items; purge removes all local state.
- REQ-ANON-022 Template-First mirrors, inventory, skill-body copies, docs-site correction.
- REQ-ANON-023 no auto-repair artifact in `internal/template/templates/` or `plugins/moai/`.

## Acceptance map (25)

AC-ANON-001 to 025 map as in `acceptance.md` §D: 001 R1, 002 R2, 003-004 R3, 005 R4, 006 R5, 007 R6, 008 R7, 009 R8, 010-011 R9, 012 R10, 013 R11, 014 R12, 015 R13, 016 R14, 017 R15, 018 R16, 019 R17, 020 R18, 021 R19, 022 R20, 023 R21, 024 R22, 025 R23.

## Files (by milestone)

M1 `internal/bugreport/{kind,verdict,payload,validate,fingerprint,frames}.go`; M2 config, settings, web, wizard, init and update files per `plan.md` §C; M3 capture, spool, attribution, register, `cmd/moai/main.go`, ten recover sites, hook registry, preference cmd, update sync and clean-install; M4 `internal/feedback/outbox/`, `internal/cli/feedback_participation.go`; M5 `internal/feedback/publish/` and golden; M6 model, adjudicate, budget; M7 three `feedback.md` copies, four docs pages, `internal/template/auto_repair_guard_test.go`.

## Exclusions

- Issue to auto-repair to pull request pipeline (development-repository-only; never shipped in templates).
- User project bugs, user config errors, environment problems, Claude Code tool failures, user prompts and content.
- Anonymous reporting (no relay, bot, App, shipped credential, non-GitHub intake).
- Changing the vulnerability policy or classifier vocabulary.
- Free-text reports from automatic detection.
