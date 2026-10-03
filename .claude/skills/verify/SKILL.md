---
description: >
  Pre-commit verification entry point. Claude Code 2.1.286+ instructs the model
  to run a skill named `verify` right before committing (docs-only and
  tests-only commits exempt); this thin wrapper delegates that call to the MoAI
  quality gate instead of hand-rolling substitute checks.
metadata:
  version: "1.0.0"
  category: "workflow"
  status: "active"
  updated: "2026-10-03"
  tags: "verify, gate, pre-commit, quality"
---

# Verify — Delegates to the MoAI Gate

Run the MoAI pre-commit quality gate: invoke `Skill("moai")` with arguments `gate $ARGUMENTS`.

Rules for the caller:

- Do not substitute your own checks — the gate owns language detection, lint,
  format, type-check, and test execution, plus shared-snapshot reuse/recording.
- On FAIL with fixable lint/format findings, prefer re-running the gate with
  `--fix` before editing code by hand.
- The gate's PASS/FAIL table is the verification evidence for the upcoming
  commit. Resolve a FAIL (auto-fix or manual) before committing; in
  MoAI-managed projects the mechanical pre-commit hook independently re-runs
  the same gate at the commit tier and owns the blocking decision.
