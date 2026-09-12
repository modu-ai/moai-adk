package cli

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/modu-ai/moai-adk/internal/gateway"
	"github.com/spf13/cobra"
)

type gatewayHandlerFactory func(json.RawMessage) (http.Handler, error)

// newGatewayChildCommand consumes private pipe configuration. A nil factory
// keeps production closed until the transport and credential gates are verified.
func newGatewayChildCommand(factory gatewayHandlerFactory) *cobra.Command {
	return &cobra.Command{
		Use: "internal-gateway", Hidden: true, Args: cobra.NoArgs, SilenceUsage: true, SilenceErrors: true,
		// The private child does not execute root startup hooks or project mutation.
		PersistentPreRunE: func(*cobra.Command, []string) error { return nil },
		RunE: func(cmd *cobra.Command, _ []string) error {
			if factory == nil {
				return errors.New("gateway transport verification is not complete")
			}
			cfg, err := gateway.ReadChildConfig(cmd.InOrStdin())
			if err != nil {
				return err
			}
			handler, err := factory(cfg.Payload)
			if closer, ok := handler.(io.Closer); ok {
				defer closer.Close()
			}
			if err != nil || handler == nil {
				return errors.New("gateway handler initialization failed")
			}
			control, ok := cmd.InOrStdin().(io.ReadCloser)
			if !ok {
				control = io.NopCloser(cmd.InOrStdin())
			}
			return gateway.RunChildWithControl(cmd.Context(), cfg, handler, cmd.OutOrStdout(), control)
		},
	}
}
