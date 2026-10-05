// integration_policy.go — `moai integration policy open|hold --reason`
// (card t1479, SPEC-MERGE-WINDOW-QUEUE-001 REQ-MWQ-012): the leader's window
// control. `open` by absence; a hold refuses every acquire without --wait
// (naming the reason) and suspends every promotion; the policy verb refuses
// a session that declares the lane role — the lane does not close the
// window on itself or on the other lanes.
package cli

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/factory"
	"github.com/spf13/cobra"
)

// newIntegrationPolicyCmd builds the policy command with its two
// subcommands.
func newIntegrationPolicyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "policy open|hold --reason <text>",
		Short: "Read or set the release-integration window policy (open | hold)",
	}
	cmd.AddCommand(newIntegrationPolicyOpenCmd(), newIntegrationPolicyHoldCmd())
	return cmd
}

// integrationLaneRoleSet reports whether this session declares the lane
// role — the claim that refuses policy writes (REQ-MWQ-012). An unset
// variable makes no claim.
func integrationLaneRoleSet() bool {
	return strings.TrimSpace(os.Getenv(config.EnvFactoryRole)) != ""
}

func newIntegrationPolicyOpenCmd() *cobra.Command {
	var sessionFlag string
	cmd := &cobra.Command{
		Use:   "open",
		Short: "Set the window policy to open (promotion resumes in queue order)",
		RunE: func(cmd *cobra.Command, args []string) error {
			return integrationSetWindowPolicy(cmd, factory.IntegrationWindowPolicy{Policy: factory.PolicyOpen, SetBy: sessionFlag, SetAt: factory.WindowClock().Format(time.RFC3339)})
		},
	}
	cmd.Flags().StringVar(&sessionFlag, "session", "", "Session id recorded as the policy setter")
	return cmd
}

func newIntegrationPolicyHoldCmd() *cobra.Command {
	var sessionFlag, reason string
	cmd := &cobra.Command{
		Use:   "hold --reason <text>",
		Short: "Set the window policy to hold: no promotion until policy open",
		RunE: func(cmd *cobra.Command, args []string) error {
			if strings.TrimSpace(reason) == "" {
				return fmt.Errorf("integration policy hold: --reason <text> is required (an unlabeled hold cannot be lifted by the reader who finds it)")
			}
			return integrationSetWindowPolicy(cmd, factory.IntegrationWindowPolicy{Policy: factory.PolicyHold, Reason: reason, SetBy: sessionFlag, SetAt: factory.WindowClock().Format(time.RFC3339)})
		},
	}
	cmd.Flags().StringVar(&sessionFlag, "session", "", "Session id recorded as the policy setter")
	cmd.Flags().StringVar(&reason, "reason", "", "Why the window is held (shown by status and by every refused acquire)")
	return cmd
}

// integrationSetWindowPolicy writes the policy after the lane-role refusal.
// The refusal runs against the ENVIRONMENT, not the record: a lane that
// cannot write the policy must not be able to read-and-decide otherwise,
// and the record's bytes are untouched on a refusal.
func integrationSetWindowPolicy(cmd *cobra.Command, policy factory.IntegrationWindowPolicy) error {
	if integrationLaneRoleSet() {
		return fmt.Errorf("integration policy: refused — this session declares the lane role (MOAI_FACTORY_ROLE); the window policy is the leader's control (REQ-MWQ-012)")
	}
	root := integrationLockRoot()
	if err := factory.WriteIntegrationWindowPolicy(root, policy); err != nil {
		return err
	}
	if policy.Policy == factory.PolicyHold {
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "integration window policy: hold (%s)\n", policy.Reason)
		return nil
	}
	_, _ = fmt.Fprintln(cmd.OutOrStdout(), "integration window policy: open")
	return nil
}
