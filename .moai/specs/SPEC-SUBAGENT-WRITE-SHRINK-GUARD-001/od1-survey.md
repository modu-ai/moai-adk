---
id: SPEC-SUBAGENT-WRITE-SHRINK-GUARD-001
title: "M1 false-positive survey record (OD-1 disposition)"
created: 2026-09-22
author: manager-develop
tier: M
---

# OD-1 Survey Record — SPEC-SUBAGENT-WRITE-SHRINK-GUARD-001

Mechanical record required by AC-SWG-011. Full survey procedure, verbatim outputs, and
per-hit classification: `.moai/reports/t1057/m1-survey.md` (card evidence path,
machine-local).

od1_answer: OD-1a

The operator answered OD-1a on 2026-09-22 (spec.md §D.3 [RESOLVED — OD-1a]): the §D.2
threshold pair (`SWG-T1` = 2000 bytes, `SWG-T2` = 25%) ships as the guard's starting values,
adopted by operator decision. The pair is a PROPOSAL, not a measurement, in every artifact
that cites it.

counts: 22 survivors, 0 destructive accidents, 22 deliberate, 0 in-place SPEC amendments

- Survey surface: `git log --numstat` over `.claude/ .moai/ internal/ pkg/ cmd/ *.md`;
  6173 candidate rows; 288 rows reached the byte-level ratio filter; 5883 rows were
  git-level file deletions (post-image absent — outside the Write-shaped population).
- Positive control: `ce79ef7caf` / `internal/cli/cg.go` (pre 4438 → post 499, 11%)
  observed surviving steps 1-3 before any count was read.
- Classification: all 22 survivors deliberate (stub-plus-companion extraction, roster
  consolidation, absorption, dead-code sweep, CHANGELOG consolidation); 0 destructive
  accidents; 0 in-place SPEC amendments in committed history.
- The survey measures committed history only. It measures the false-positive side; it is
  not evidence about the true-positive rate (the motivating incident never reached a
  commit — spec.md §A.2).

🗿 MoAI
