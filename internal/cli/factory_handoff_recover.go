package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/modu-ai/moai-adk/internal/homestate"
	"github.com/spf13/cobra"
)

func newFactoryCommand() *cobra.Command {
	factory := &cobra.Command{Use: "factory", Short: "Factory runtime maintenance"}
	handoff := &cobra.Command{Use: "handoff", Short: "Factory handoff maintenance"}
	var id int64
	var token, decision string
	recover := &cobra.Command{Use: "recover-resume", Short: "Recover an indeterminate legacy resume claim", RunE: func(cmd *cobra.Command, _ []string) error {
		if id <= 0 || token == "" || (decision != "requeue" && decision != "fail") {
			return fmt.Errorf("--id, --expected-token, and --decision requeue|fail are required")
		}
		root, err := os.Getwd()
		if err != nil {
			return err
		}
		db, err := homestate.OpenFactory(root)
		if err != nil {
			return err
		}
		defer func() { _ = db.Close() }()
		return db.RecoverLegacyResume(context.Background(), id, token, decision, homestate.ProbeProcessIdentity, func() error {
			census, err := homestate.ReadRuntimeCensus(root)
			if err != nil {
				return err
			}
			if census.Total() != 0 {
				return fmt.Errorf("active runtimes=%d", census.Total())
			}
			return nil
		})
	}}
	recover.Flags().Int64Var(&id, "id", 0, "Resume handoff row id")
	recover.Flags().StringVar(&token, "expected-token", "", "Current claim token")
	recover.Flags().StringVar(&decision, "decision", "", "Recovery decision: requeue or fail")
	handoff.AddCommand(recover)
	factory.AddCommand(handoff)
	return factory
}

func init() { rootCmd.AddCommand(newFactoryCommand()) }
