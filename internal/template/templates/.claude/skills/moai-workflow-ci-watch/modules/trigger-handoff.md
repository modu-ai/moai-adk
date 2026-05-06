# T3 Trigger Handoff Schema

Metadata schema for Wave 3 (expert-debug auto-fix loop) consumption.
Produced on exit code 2 from `scripts/ci-watch/run.sh`.

## JSON Schema (Wave 2 stable)

```json
{
  "prNumber": 785,
  "branch": "feat/SPEC-V3R3-CI-AUTONOMY-001-wave-2",
  "failedChecks": [
    {
      "name": "Lint",
      "runId": "12345678",
      "logUrl": "https://github.com/modu-ai/moai-adk/actions/runs/12345678",
      "conclusionDetail": ""
    }
  ],
  "auxiliaryFailCount": 1,
  "totalRequired": 6
}
```

## Field Definitions

| Field | Type | Description |
|-------|------|-------------|
| `prNumber` | int | GitHub PR number |
| `branch` | string | Head branch name |
| `failedChecks` | array | Required checks that reached failure/cancelled/timed_out |
| `failedChecks[].name` | string | Check context name (matches SSoT) |
| `failedChecks[].runId` | string | GitHub Actions workflow run ID |
| `failedChecks[].logUrl` | string | Direct URL to run logs |
| `failedChecks[].conclusionDetail` | string | Optional human-readable failure summary |
| `auxiliaryFailCount` | int | Count of advisory failures (NOT blocking) |
| `totalRequired` | int | Total required checks tracked |

## Stability Contract

- Fields `name`, `runId`, `logUrl` are stable for Wave 3 (do not rename/remove)
- `conclusionDetail` is optional; Wave 3 may be empty string
- `auxiliaryFailCount` is informational; Wave 3 should NOT treat as blocking
- Additional fields may be added in Wave 4+ (backward compatible)

## Orchestrator Consumption Pattern

```bash
# On exit 2, read JSON from stdout:
handoff_json="$(MOAI_CIWATCH_GH=gh sh scripts/ci-watch/run.sh $PR_NUMBER $BRANCH)"
exit_code=$?

if [ "$exit_code" = "2" ]; then
    # Inject into Wave 3 expert-debug spawn prompt:
    # "Fix the failing CI checks for PR #N. Handoff metadata: $handoff_json"
    # expert-debug reads logUrl to fetch logs and diagnoses root cause.
fi
```

## Go Types (`internal/ciwatch/handoff.go`)

```go
type FailedCheck struct {
    Name             string `json:"name"`
    RunID            string `json:"runId,omitempty"`
    LogURL           string `json:"logUrl,omitempty"`
    ConclusionDetail string `json:"conclusionDetail,omitempty"`
}

type Handoff struct {
    PRNumber           int           `json:"prNumber"`
    Branch             string        `json:"branch"`
    FailedChecks       []FailedCheck `json:"failedChecks"`
    AuxiliaryFailCount int           `json:"auxiliaryFailCount"`
    TotalRequired      int           `json:"totalRequired"`
}
```
