---
description: "(Deprecated) AI Agency - use /moai design instead. Redirects to /moai design workflow."
argument-hint: "[brief|build|review|profile] [args]"
allowed-tools: Skill
---

> DEPRECATED — /agency has been absorbed into /moai design (SPEC-AGENCY-ABSORB-001)

The /agency command is deprecated as of this release. Its functionality is now part of /moai design
with hybrid Claude Design (path A) and code-based skill (path B) execution paths.

## Migration

| Old /agency subcommand | New /moai design equivalent |
|---|---|
| /agency brief | /moai plan (manager-spec with BRIEF Goal/Audience/Brand sections) |
| /agency build | /moai design |
| /agency review | /moai e2e + evaluator-active |
| /agency learn | /moai-workflow-research |
| /agency evolve | /moai-workflow-research |
| /agency profile | /moai project |
| /agency resume | /moai run SPEC-XXX (resume from SPEC document) |

This wrapper will be removed in the next minor version per SPEC-AGENCY-ABSORB-001 REQ-DEPRECATE-003.

Use Skill("moai") with arguments: design $ARGUMENTS
