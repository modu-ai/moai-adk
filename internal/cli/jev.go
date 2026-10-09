package cli

import (
	"fmt"
	"strings"

	"github.com/modu-ai/moai-adk/internal/jevcred"
	"github.com/spf13/cobra"
)

// jevCmd manages the Jev (TypeSafe) credential. SPEC-GLM-JEV-KEY-001 gives the
// stored credential its first CLI surface: `moai jev --key <credential>`
// stores through the existing jevcred writer; bare `moai jev` is help, exit 0,
// storage untouched.
//
// @MX:NOTE: [AUTO] storage SSOT lives in internal/jevcred — this command only
// validates input and delegates; a second writer or a reveal route must never
// be introduced here (REQ-GJK-007/010, plan.md §G).
var jevCmd = &cobra.Command{
	Use:   "jev [--key <api-key>]",
	Short: "Manage the Jev (TypeSafe) credential",
	Long: `Manage the Jev (TypeSafe) credential.

The credential is stored at ~/.moai/.env.typesafe (mode 0600) — the same file
'moai doctor', the web console, and the jev_ask gate read.

Use 'moai jev --key <credential>' to store the credential.
Run 'moai jev' without --key to print this help.`,
	GroupID: "tools",
	RunE:    runJev,
}

func init() {
	jevCmd.Flags().String("key", "", "Store the TypeSafe API credential")
	rootCmd.AddCommand(jevCmd)
}

// runJev stores the TypeSafe credential through the jevcred writer when --key
// carries a value, and prints help otherwise (REQ-GJK-009: bare `moai jev` is
// help + exit 0).
func runJev(cmd *cobra.Command, _ []string) error {
	key, err := cmd.Flags().GetString("key")
	if err != nil {
		return err
	}
	if !cmd.Flags().Changed("key") {
		return cmd.Help()
	}

	trimmed := strings.TrimSpace(key)
	if trimmed == "" {
		return fmt.Errorf("empty Jev credential")
	}
	// REQ-GJK-012: reject-before-write, mirroring the web validator's
	// trim-then-ContainsAny ordering (internal/web/jevkey.go). jevcred cannot
	// round-trip a newline (EscapeValue deliberately does not escape it; Load
	// reads line-by-line), so a line-bearing value must never reach the
	// writer — the existing credential file stays byte-for-byte unchanged.
	if strings.ContainsAny(trimmed, "\r\n") {
		return fmt.Errorf("Jev credential must not contain line breaks")
	}
	if err := jevcred.Save(trimmed); err != nil {
		return fmt.Errorf("save Jev credential: %w", err)
	}

	// REQ-GJK-008: the confirmation discloses at most what jevcred.View
	// permits — the configured fact plus, for a credential longer than four
	// characters, its final four characters (REQ-JEVC-020).
	if hint := jevcred.View(); hint.Hint != "" {
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Jev credential stored (…%s)\n", hint.Hint)
	} else {
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Jev credential stored")
	}
	return nil
}
