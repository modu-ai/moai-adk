package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/modu-ai/moai-adk/internal/homestate"
)

func profileLeaseTokenFromEnv(env []string) string {
	for _, item := range env {
		if strings.HasPrefix(item, "MOAI_PROFILE_LEASE_TOKEN=") {
			return strings.TrimPrefix(item, "MOAI_PROFILE_LEASE_TOKEN=")
		}
	}
	return ""
}
func transferProfileLeaseToChild(env []string, parentPID int, parentFingerprint string, childPID int, childFingerprint string) error {
	token := profileLeaseTokenFromEnv(env)
	if token == "" {
		return nil
	}
	if childPID <= 0 || childFingerprint == "" {
		return fmt.Errorf("child profile lease identity indeterminate")
	}
	store, err := homestate.OpenProfileLeases()
	if err != nil {
		return err
	}
	defer store.Close()
	return store.TransferToChild(context.Background(), token, parentPID, parentFingerprint, childPID, childFingerprint)
}
