package cli

// Launcher verb wiring for the stripped-replay envelope repair
// (SPEC-GATEWAY-ENVELOPE-REPAIR-001 REQ-EVR-004/005). The repair runs only
// when the user explicitly passes --repair-envelope alongside --resume; a
// classification alone never modifies anything. On refusal the gateway's
// classified guidance is surfaced unchanged — constructed from the validator's
// own error type, never duplicated.

import (
	"context"
	"errors"
	"fmt"

	"github.com/modu-ai/moai-adk/internal/gateway/conversation"
	"github.com/modu-ai/moai-adk/internal/gateway/translate"
)

// repairGatewayEnvelopeFlag is the explicit launcher invocation surface.
const repairGatewayEnvelopeFlag = "--repair-envelope"

// gatewayRepairGuidance returns the frozen classified reasoning guidance by
// reading it off the validator's own error type (single SSOT; REQ-EVR-007).
func gatewayRepairGuidance() string {
	return translate.HistoryReplayError{Cause: translate.CauseReasoning}.Error()
}

// repairGatewayEnvelope runs the launcher-side repair before the resume flow
// when the user asked for it. Without the flag it is a strict no-op.
func repairGatewayEnvelope(in gatewayLaunchRequest, families *conversation.Manager) error {
	requested := false
	for _, arg := range in.Args {
		if arg == repairGatewayEnvelopeFlag {
			requested = true
			break
		}
	}
	if !requested {
		return nil
	}
	if families == nil {
		return errors.New("gateway envelope repair requires a family conversation")
	}
	if in.Continue {
		return errors.New("gateway envelope repair requires --resume <uuid>, not --continue")
	}
	id := ""
	for i := 0; i < len(in.Args); i++ {
		if in.Args[i] == "--resume" && i+1 < len(in.Args) {
			id = in.Args[i+1]
			break
		}
	}
	if id == "" {
		return errors.New("gateway envelope repair requires --resume <uuid>")
	}
	if _, err := families.RepairEnvelope(context.Background(), id); err != nil {
		if errors.Is(err, conversation.ErrRepairAlreadyAttempted) || errors.Is(err, conversation.ErrRepairNotRepairable) || errors.Is(err, conversation.ErrMissing) {
			return fmt.Errorf("gateway envelope repair: %s: %w", gatewayRepairGuidance(), err)
		}
		return fmt.Errorf("gateway envelope repair: %w", err)
	}
	return nil
}
