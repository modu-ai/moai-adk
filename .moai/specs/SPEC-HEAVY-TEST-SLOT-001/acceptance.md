---
id: SPEC-HEAVY-TEST-SLOT-001
title: "Acceptance — heavy-test slot discipline doc + lane-protocol pointer"
version: "0.1.0"
status: draft
created: 2026-09-14
updated: 2026-09-14
author: lane-6
tier: S
---

## Acceptance Criteria

- **AC-HTS-001** — `.moai/docs/heavy-test-slot-protocol.md` exists and contains, verbatim-checkable: the WHEN rule naming at least `internal/cli`; the three procedure commands (`slot status`, `slot acquire --resource <pkg>-tests …`, `slot release`); the enforcement fact with the literal exit code 3; and the `moai integration` distinction sentence.
  - Verify: `/usr/bin/grep -c 'internal/cli' <doc>` ≥ 1; `grep -c 'slot acquire --resource' <doc>` = 1; `grep -c 'exit 3' <doc>` ≥ 1; `grep -c 'moai integration' <doc>` ≥ 1.
- **AC-HTS-002** — The document records the control-group obligation: the 2026-09-10 incident as no-surface evidence and the exit-3 demonstration as enforcement evidence, plus an explicit sentence that concurrent heavy-suite re-runs must not be used to re-demonstrate.
  - Verify: `grep -c '2026-09-10' <doc>` ≥ 1; `grep -c '재현\|re-demonstrate\|재시연' <doc>` ≥ 1 (either locale form acceptable; document language is Korean).
- **AC-HTS-003** — `.claude/rules/local/gitflow-lane-protocol.md` carries a pointer paragraph into the protocol document, and the paragraph does NOT restate the procedure commands (pointer, not a second copy).
  - Verify: `grep -c 'heavy-test-slot-protocol' <rule>` ≥ 1; `grep -c 'slot acquire' <rule>` = 0.
- **AC-HTS-004** — Dev-only boundary holds: no new or modified file under `internal/template/templates/` and none under `docs-site/` in the card's changed-file set.
  - Verify: `git show <sync-commit> --stat` contains neither prefix.
- **AC-HTS-005** — Product code untouched: `internal/cli/slot.go` and its tests are byte-identical to the base commit `d416f8162`.
  - Verify: `git diff d416f8162 <HEAD> -- internal/cli/slot.go` is empty.

## Evidence contract

Every AC records command + verbatim output in the card verdict (`.moai/reports/t774/verdict.md`); a claim whose cited path no longer resolves is an unattributed claim.
