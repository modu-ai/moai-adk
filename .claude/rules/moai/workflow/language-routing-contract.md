---
description: "Manifest and source-aware language routing for monorepo sync gates"
paths:
  - ".claude/hooks/moai/sync-phase-quality-gate.sh"
  - ".claude/skills/moai/workflows/sync/**"
---

# Language routing contract

Language detection is evidence collection, not a first-match guess. Resolve
all candidates from repository markers and changed-file suffixes, de-duplicate
them, and record the candidate list in the gate/quality report.

`build.gradle.kts` is Kotlin only when its contents show a Kotlin plugin or
Kotlin source is present; otherwise it is Java/Gradle. A monorepo may return
multiple candidates. The fast hook may choose one primary command for its
bounded Stop-time check, but the full sync quality phase must shard or queue
the remaining candidates and must not skip a changed language because another
marker appeared first.

An unknown candidate is a recorded routing gap, not a successful no-op. A
source-suffix candidate with no manifest is still eligible for the language
toolchain check, subject to the tool-availability contract.
