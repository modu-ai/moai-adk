package cli

// verify_receipts.go — receipt helpers shared by the staying verify surface.
// (card t1909 step ⑤: the out-of-hook Codex Stop-chain receipt producers this
// file hosted — SPEC-DUAL-HARNESS-HOOK-PARITY-001 M2d, design §D3.3 — were
// removed together with the stop chain; verifyReceiptContext remains for the
// staying receipt emitters.)

import (
	"context"

	"github.com/spf13/cobra"
)

// verifyReceiptContext returns the command's context, or a background one when
// the command runs outside Execute (tests call RunE directly).
func verifyReceiptContext(c *cobra.Command) context.Context {
	if ctx := c.Context(); ctx != nil {
		return ctx
	}
	return context.Background()
}
