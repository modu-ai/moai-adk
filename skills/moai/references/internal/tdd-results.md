# Internal: tdd-results

When to read: when you record the result of a test run while writing code test first, as `moai-impl` in the red, green and refactor steps.

Record what a run means, not only its exit code. Every run gets exactly one of four results:

| Result | Meaning | Effect |
| --- | --- | --- |
| `EXPECTED_RED` | A new test reached the assertion it was written for and failed because the behaviour does not exist yet | Allowed only in the red step, for that test |
| `REGRESSION_FAILURE` | A test that passed before fails, or the new test still fails after the green step | Stops the cycle until the cause is found |
| `TOOL_FAILURE` | A compile error, missing test, broken fixture, tool error or timeout kept the assertion from running | Stops the cycle; never counts as red |
| `PASS` | The selected tests ran and every expected assertion passed | Required to finish the green and refactor steps |

- A red record names the new test, the command, its exit code and the output lines that show the intended assertion failing. A non-zero exit without that output is `TOOL_FAILURE` or `REGRESSION_FAILURE`, never `EXPECTED_RED`.
- The green step runs the new test and the existing tests it could affect. Refactoring keeps both results.
- When a failure is reclassified (for example a red that turns out to be a missing fixture), the report gives the new result and the reason.
- A run that selected no test is none of the four: report it as not_run (see `internal/test-design`).
