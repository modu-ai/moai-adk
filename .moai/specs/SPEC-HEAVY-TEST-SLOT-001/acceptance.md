---
id: SPEC-HEAVY-TEST-SLOT-001
title: "Acceptance — §8 slot-paragraph extension"
version: "0.2.0"
created: 2026-09-14
updated: 2026-09-14
author: lane-6
tier: S
---

## Acceptance Criteria

The base for every check is the merge base of the card branch and the base commit recorded in the verdict (`git merge-base <base-commit> HEAD`), so a develop absorb cannot false-FAIL an anchor.

- **AC-HTS-001** — The §8 slot paragraph in `.claude/rules/local/gitflow-lane-protocol.md` names the enforcement code and the heavy packages.
  - Verify: `grep -c 'exit 3\|종료 코드 3' <rule>` ≥ 1 AND `grep -c 'internal/cli' <rule>` ≥ 1.
- **AC-HTS-002** — The §8 paragraph records the control group: the 2026-09-10 incident and the explicit prohibition on concurrent heavy-suite re-demonstration.
  - Verify: `grep -c '2026-09-10' <rule>` ≥ 1 AND `grep -cE '재시연|재현|re-demonstrate' <rule>` ≥ 1.
- **AC-HTS-003** — The extension stays dev-only: no changed file under `internal/template/templates/`, `docs-site/`, or `.claude/rules/moai/workflow/resource-slot-lease.md`.
  - Verify: `git diff --name-only <merge-base> HEAD` matches none of the three prefixes.
- **AC-HTS-004** — Product code untouched: `internal/cli/slot.go` and its tests are byte-identical to the merge base.
  - Verify: `git diff <merge-base> HEAD -- internal/cli/slot.go internal/cli/slot_test.go` is empty.

## Evidence contract

Every AC records command + verbatim output in the card verdict (`.moai/reports/t774/verdict.md`); a claim whose cited path no longer resolves is an unattributed claim.
