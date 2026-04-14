package telemetry

// DetermineOutcome applies conservative heuristics to infer the outcome of a
// skill invocation based on subsequent session events.
//
// Heuristic rules (applied in priority order):
//  1. No events → OutcomeUnknown (conservative default)
//  2. Error signal only → OutcomeError
//  3. Test pass + no error → OutcomeSuccess
//  4. Test pass + test fail → OutcomePartial
//  5. Test pass + error → OutcomePartial (mixed signal)
//  6. Test fail only → OutcomeError
//  7. No signal flags set → OutcomeUnknown
//
// The function intentionally defaults to OutcomeUnknown to avoid polluting
// evolution proposals with false positives.
func DetermineOutcome(events []Event) string {
	if len(events) == 0 {
		return OutcomeUnknown
	}

	var hasError, hasTestPass, hasTestFail bool

	for _, e := range events {
		if e.IsError {
			hasError = true
		}
		if e.IsTestPass {
			hasTestPass = true
		}
		if e.IsTestFail {
			hasTestFail = true
		}
	}

	// No signal flags set — cannot determine outcome.
	if !hasError && !hasTestPass && !hasTestFail {
		return OutcomeUnknown
	}

	// Mixed signal: test pass + test fail → partial.
	if hasTestPass && hasTestFail {
		return OutcomePartial
	}

	// Mixed signal: test pass + error → partial.
	if hasTestPass && hasError {
		return OutcomePartial
	}

	// Pure error signals.
	if hasError || hasTestFail {
		return OutcomeError
	}

	// Pure success: test pass, no error, no fail.
	if hasTestPass {
		return OutcomeSuccess
	}

	return OutcomeUnknown
}
