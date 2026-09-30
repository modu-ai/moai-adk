package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/factorylane"
	"github.com/modu-ai/moai-adk/internal/sessionmsg"
	"github.com/spf13/cobra"
)

// factoryLaneFallbackStore anchors the lane-fallback state store at the
// command's project root.
func factoryLaneFallbackStore() (*factorylane.Store, error) {
	root, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	return factorylane.NewStore(root, nil), nil
}

// newFactoryMessagingCommand groups the messaging-availability surfaces of
// SPEC-FACTORY-LANE-AUTONOMY-001 fragment 1 (design.md D1): the availability
// probe, the directed-request record, and its ack. The probe answers "is the
// channel available"; it sends nothing, and an unavailable verdict is a
// normal diagnostic outcome, not a command failure (REQ-FLA-001). User-facing
// strings use the leader noun (REQ-RNC-006/-007 family) — the M7 negative
// source scan enforces it.
func newFactoryMessagingCommand() *cobra.Command {
	messaging := &cobra.Command{
		Use:   "messaging",
		Short: "Cross-session messaging availability for factory lanes (probe + directed requests)",
	}
	messaging.AddCommand(newFactoryMessagingProbeCommand(), newFactoryMessagingRequestCommand(), newFactoryMessagingAckCommand())
	return messaging
}

// newFactoryMessagingProbeCommand is the REQ-FLA-001 probe: registered-peer
// listing via the sessionmsg registry, the named leader's heartbeat age against
// the configured availability bound, and the lane's active no-response
// observations. The probe is also the sweep point that materializes a
// directed request's no-response fact once its timer expires unanswered
// (REQ-FLA-002).
func newFactoryMessagingProbeCommand() *cobra.Command {
	var lead string
	var asJSON bool
	cmd := &cobra.Command{
		Use:   "probe",
		Short: "Report the channel verdict: available | channel-unavailable",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			root, err := os.Getwd()
			if err != nil {
				return err
			}
			registry, err := sessionmsg.NewStore(filepath.Join(root, sessionmsg.DefaultStateRoot), nil).ListAgents()
			if err != nil {
				return fmt.Errorf("read messaging registry: %w", err)
			}
			var observations []factorylane.Observation
			store := factorylane.NewStore(root, nil)
			if lane := os.Getenv(config.EnvMoaiFactoryWorker); lane != "" {
				if observations, err = store.SweepNoResponse(lane); err != nil {
					return fmt.Errorf("sweep directed requests: %w", err)
				}
			}
			verdict := factorylane.EvaluateAvailability(factorylane.ProbeInput{
				Now:          time.Now().UTC(),
				LeadPeer:     lead,
				OfflineBound: time.Duration(config.DefaultSessionMsgAgentOfflineMinutes) * time.Minute,
				Registry:     registry,
				Observations: observations,
			})
			if asJSON {
				data, err := json.Marshal(verdict)
				if err != nil {
					return err
				}
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), string(data))
				return nil
			}
			out := fmt.Sprintf("channel: %s\nreason: %s", verdict.Verdict, verdict.Reason)
			if verdict.UnavailableUntil != "" {
				out += fmt.Sprintf("\nunavailable-until: %s", verdict.UnavailableUntil)
			}
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), out)
			return nil
		},
	}
	cmd.Flags().StringVar(&lead, "leader", "", "Named leader peer to check (default: any registered peer)")
	cmd.Flags().BoolVar(&asJSON, "json", false, "Emit machine-readable JSON")
	return cmd
}

// newFactoryMessagingRequestCommand records a directed leader request: the
// observation that starts the bounded no-response window (REQ-FLA-002).
func newFactoryMessagingRequestCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "request",
		Short: "Record a directed leader request (starts the bounded no-response observation)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			lane, err := factoryLaneLabelFromEnv("messaging request")
			if err != nil {
				return err
			}
			store, err := factoryLaneFallbackStore()
			if err != nil {
				return err
			}
			obs, err := store.RecordRequest(lane)
			if err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "directed request recorded at %s (timer %s, bound %s)\nack with: moai factory messaging ack\n",
				obs.RequestedAt.Format(time.RFC3339), obs.Timer, obs.Bound)
			return nil
		},
	}
}

// newFactoryMessagingAckCommand marks the lane's pending directed request
// acked — the reply arrived before the timer expired.
func newFactoryMessagingAckCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "ack",
		Short: "Ack the lane's pending directed request",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			lane, err := factoryLaneLabelFromEnv("messaging ack")
			if err != nil {
				return err
			}
			store, err := factoryLaneFallbackStore()
			if err != nil {
				return err
			}
			acked, err := store.AckPending(lane)
			if err != nil {
				return err
			}
			if acked {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "directed request acked")
			} else {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "no pending directed request")
			}
			return nil
		},
	}
}

// newFactoryFallbackCommand is the fallback-transition event log surface
// (REQ-FLA-003): the bare command queries the log (count-by-lane is the
// AC-FLA-003 check); declare/restore record the mode switches.
func newFactoryFallbackCommand() *cobra.Command {
	var asJSON bool
	fb := &cobra.Command{
		Use:   "fallback",
		Short: "Query the fallback-transition event log",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			lane, err := factoryLaneLabelFromEnv("fallback")
			if err != nil {
				return err
			}
			store, err := factoryLaneFallbackStore()
			if err != nil {
				return err
			}
			events, err := store.Transitions(lane)
			if err != nil {
				return err
			}
			if asJSON {
				data, err := json.Marshal(events)
				if err != nil {
					return err
				}
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), string(data))
				return nil
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "fallback transitions for %s (%d event(s)):\n", lane, len(events))
			for _, ev := range events {
				card := ev.Card
				if card == "" {
					card = "-"
				}
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  %s  %s  card=%s\n", ev.At.Format(time.RFC3339), ev.Trigger, card)
			}
			return nil
		},
	}
	fb.Flags().BoolVar(&asJSON, "json", false, "Emit machine-readable JSON")
	fb.AddCommand(newFactoryFallbackDeclareCommand(), newFactoryFallbackRestoreCommand())
	return fb
}

// newFactoryFallbackDeclareCommand records one fallback activation — exactly
// one event per mode switch; while the lane is already in fallback the
// declare is refused with nothing written (REQ-FLA-003). The switch ACT
// itself — invoking /moai:todo --auto self-service pickup — is the lane's
// doctrine move; this verb only makes it a recorded, queryable event.
func newFactoryFallbackDeclareCommand() *cobra.Command {
	var trigger, card string
	cmd := &cobra.Command{
		Use:   "declare",
		Short: "Record the switch to self-service fallback (exactly once per switch)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			lane, err := factoryLaneLabelFromEnv("fallback declare")
			if err != nil {
				return err
			}
			store, err := factoryLaneFallbackStore()
			if err != nil {
				return err
			}
			ev, err := store.DeclareFallback(lane, factorylane.Trigger(trigger), card)
			if err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "fallback declared: trigger=%s card=%s at %s\nswitch to /moai:todo --auto self-service pickup (REQ-FLA-001)\n",
				ev.Trigger, orDash(ev.Card), ev.At.Format(time.RFC3339))
			return nil
		},
	}
	cmd.Flags().StringVar(&trigger, "trigger", "", "Fallback trigger: channel-unavailable | no-response")
	_ = cmd.MarkFlagRequired("trigger")
	cmd.Flags().StringVar(&card, "card", "", "In-progress card id at switch time (empty before pickup)")
	return cmd
}

// newFactoryFallbackRestoreCommand records the switch back to messaging so a
// later real fallback switch stays separately countable.
func newFactoryFallbackRestoreCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "restore",
		Short: "Record the return to messaging availability",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			lane, err := factoryLaneLabelFromEnv("fallback restore")
			if err != nil {
				return err
			}
			store, err := factoryLaneFallbackStore()
			if err != nil {
				return err
			}
			ev, err := store.Restore(lane)
			if err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "fallback restored for %s at %s\n", lane, ev.At.Format(time.RFC3339))
			return nil
		},
	}
}

// orDash renders an empty record field as "-" in human-readable lines.
func orDash(s string) string {
	if s == "" {
		return "-"
	}
	return s
}
