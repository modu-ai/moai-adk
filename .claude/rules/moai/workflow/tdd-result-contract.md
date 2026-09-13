---
description: "Semantic result classification for RED-GREEN-REFACTOR"
paths:
  - ".claude/skills/moai/workflows/run/**"
  - ".agents/skills/moai-workflow-tdd/**"
  - ".claude/rules/moai/workflow/**"
---

# TDD result contract

Every TDD command records a semantic result, not only an exit code:

| Result | Meaning | Pipeline effect |
|---|---|---|
| `EXPECTED_RED` | A new AC test reached its intended assertion and failed because the behavior is not implemented | Allowed only in RED for that AC |
| `REGRESSION_FAILURE` | An existing or previously passing test failed, or a RED test fails after GREEN | Blocks the cycle and requires diagnosis |
| `TOOL_FAILURE` | Compile/discovery/fixture/tool/timeout failure prevented the intended assertion from running | Blocks; never counts as RED |
| `PASS` | The selected test set completed with all expected assertions passing | Required for GREEN and REFACTOR completion |

The RED record includes the new test identity, command, exit status, and
verbatim output showing the intended assertion. A non-zero exit code without
that semantic evidence is `TOOL_FAILURE` or `REGRESSION_FAILURE`, never an
expected RED. GREEN runs both the new AC test and the regression set. Refactor
must preserve both results. The phase report carries the classification and a
rerun reason whenever a failure is reclassified.
