# acceptance.md — SPEC-FACTORY-WORKER-NAMING-001 (card t1085)

## §A Scope of Verification

Verification layer for the three work items: GTD naming closure (M1), serialization gate + inventory (M2), vocabulary rename (M3), compatibility decision + embed + verification (M4). Requirements layer lives in spec.md §B (GEARS); this file carries the observable, binary-testable criteria.

## §D AC Matrix

| AC | Requirement | Milestone | Severity | Verification |
|----|-------------|-----------|----------|--------------|
| AC-001 | REQ-001 | M1 | MUST | file read + grep |
| AC-002 | REQ-002 | M2 | MUST | command + output record |
| AC-003 | REQ-003 | M2 | MUST | blocker-path exercise or gate-pass evidence |
| AC-004 | REQ-004 | M3 | MUST | grep + build + run surface |
| AC-005 | REQ-005 | M3 | MUST | 4-locale grep parity |
| AC-006 | REQ-006 | M3 | MUST | per-file grep + make build + embed check |
| AC-007 | REQ-007/008 | M4 | MUST | progress.md decision record vs inventory |
| AC-008 | REQ-009 | M1/M4 | MUST | TestGTD run output |
| AC-009 | REQ-010 | M3 | MUST | affected-package test run |
| AC-010 | REQ-011 | M3/M4 | MUST | grep sweep vs exception list |

## §D.1 Acceptance Criteria (Given-When-Then)

- **AC-001** Given the closure note at `.moai/specs/SPEC-FACTORY-WORKER-NAMING-001/gtd-todo-naming-closure.md`, When its content is read, Then it states the operator decision dated 2026-09-22 that `moai todo` keeps its name (no GTD-branded rename), names the 8-tree GTD family as investigation-only, records t855 as zero-work-commits, and names t1084 as the disposal owner.
- **AC-002** Given the develop tree is readable, When the M2 gate runs, Then progress.md records both observations with command + verbatim output: `internal/factorymsg/` presence in develop AND `SPEC-FACTORY-MIXED-HOOK-001` status read from develop.
- **AC-003** Given the M2 gate result, When the gate fails, Then a blocker report is returned and zero rename edits exist in the tree (verifiable: `git status --porcelain` shows only plan artifacts); When the gate passes, Then the recorded pass evidence exists before the first M3 edit.
- **AC-004** Given the post-M3 tree, When `grep -n '"agent"' internal/cli/factory.go` and the help-text surfaces are inspected, Then the join token value is `"worker"`, help text advertises `-f worker`, and no user-facing string advertises `-f agent` (except M4-retained alias entries listed in progress.md).
- **AC-005** Given `internal/hook/session_start_factory_i18n.go`, When `grep -c 'lane-'` is run on it post-M3, Then the count is 0 (or every residual hit appears on the REQ-007/M4 exception list recorded in progress.md — e.g. a retained alias hint), and each of the four locales (en/ko/ja/zh) carries the worker-axis wording in lockstep (grep for `worker-1..worker-` returns one hit per locale block).
- **AC-006** Given the four doc-twin files (local + template mirror × 2 names), When `grep -c 'lane-'` runs per file post-M3, Then notation hits are 0 or per-line dispositions are recorded; and When a template-mirror edit occurred, Then `make build` ran after it and the rebuilt binary embeds the new content (embed-check passes).
- **AC-007** Given the M2 inventory table, When the M4 decision is written, Then every old form (`-f agent`, `-f lane-<n>`, `--name lane-<n>`) has an explicit keep-alias-or-remove decision with its measured basis, and no old form was removed at any point before M4 (verifiable via commit history order).
- **AC-008** Given the GTD compat surface, When `go test ./internal/cli/ -run TestGTD` runs post-M1 and post-M4, Then it passes — the closure record changed nothing in the `gtd`↔`todo` alias surface.
- **AC-009** Given the six measured test files, When the affected packages are tested post-M3, Then `go test ./internal/cli/ ./internal/kanban/ ./internal/hook` passes and no test file still constructs `lane-`/`agent-<n>` join labels except cases explicitly retained under the M4 alias decision.
- **AC-010** Given the final tree, When the closing sweep runs (`grep -rn -- '-f agent' internal/` and `grep -rn 'lane-' internal/cli/factory.go internal/hook/session_start_factory_i18n.go`), Then every hit is either zero or appears on the M4 exception list recorded in progress.md.

## §D.2 Edge Cases

- A locale string updated but its sibling hint line (the `-f lane-<n>` help variant) missed — AC-005's lockstep grep catches it.
- Template mirror edited without `make build` — AC-006's embed check catches it.
- t1074 lands with DIFFERENT naming than this SPEC assumes (e.g. it already renamed some slots) — M2 measures the LANDED tree; the plan's line numbers are plan-time baselines, not run-time guarantees.
- Old tokens referenced by operator-side scripts outside the repo — inventory informs keep-alias; write-reach does not extend there.

## §D.3 Quality Gates

- Standard harness: `go vet` → `golangci-lint run` → `go test` on affected packages only (never full suite locally — load discipline).
- `moai spec lint` on this SPEC directory (post-commit re-measure — ownership lint is blind before commit).
- Template neutrality: renamed doc content free of SPEC IDs, internal dates, commit SHAs.

## §D.4 Definition of Done

- AC-001..AC-010 all PASS with recorded evidence paths.
- M2 gate + inventory recorded in progress.md.
- M4 decision recorded with measured basis.
- Affected-package tests green; CI verdict on the pushed develop head (lead batch push) is the integration judgment.
