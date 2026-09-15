package cli

import (
	"context"
	"encoding/json"
	"net/http"
	"path/filepath"

	"github.com/modu-ai/moai-adk/internal/codexbridge"
	"github.com/modu-ai/moai-adk/internal/gateway/conversation"
	"github.com/modu-ai/moai-adk/internal/gateway/receipt"
	"github.com/modu-ai/moai-adk/internal/gateway/translate"
)

func authorizeManagedGPTNativeReceipt(ctx context.Context, body []byte, policy translate.NativePolicy, c gatewayPrivateConversation, store *receipt.Store) error {
	if policy.UserID != "" {
		if _, err := managedGPTConversationForRequest(ctx, c, nil, body); err != nil {
			return err
		}
	}
	// The root receipt remains unchanged. A cleared App Server thread owns a new
	// conversation ID and cannot gain history or pending tools from the old ID.
	policy.UserID = ""
	return authorizeGatewayNativeReceipt(ctx, body, policy, c, store)
}

func managedGPTConversationForRequest(ctx context.Context, c gatewayPrivateConversation, headers http.Header, body []byte) (string, error) {
	var request struct {
		Metadata struct {
			UserID string `json:"user_id"`
		} `json:"metadata"`
	}
	if json.Unmarshal(body, &request) != nil {
		return "", codexbridge.ErrScope
	}
	session := headers.Get("X-Claude-Code-Session-Id")
	if len(headers.Values("X-Claude-Code-Session-Id")) > 1 {
		return "", codexbridge.ErrScope
	}
	if request.Metadata.UserID != "" {
		var identity struct {
			SessionID string `json:"session_id"`
			DeviceID  string `json:"device_id"`
		}
		if json.Unmarshal([]byte(request.Metadata.UserID), &identity) != nil || identity.SessionID == "" || identity.DeviceID == "" {
			return "", codexbridge.ErrScope
		}
		if session != "" && session != identity.SessionID {
			return "", codexbridge.ErrScope
		}
		session = identity.SessionID
	}
	if session == "" {
		session = c.SessionID
	}
	if err := authorizeGatewayNativeSession(ctx, c, session); err != nil {
		return "", codexbridge.ErrScope
	}
	return session, nil
}

func authorizeGatewayNativeSession(ctx context.Context, c gatewayPrivateConversation, session string) error {
	if session == c.SessionID {
		return nil
	}
	// Only canonical launcher family paths can locate a native profile. Arbitrary
	// factory fixtures and request-supplied paths cannot supply clear authority.
	root := filepath.Dir(c.ReceiptDir)
	if filepath.Base(root) == c.SessionID && filepath.Base(filepath.Dir(root)) == "forks" {
		root = filepath.Dir(filepath.Dir(root))
	}
	if filepath.Base(c.ReceiptDir) != "receipt" || filepath.Base(root) != c.FamilyID {
		return codexbridge.ErrScope
	}
	if err := conversation.ValidateNativeClear(ctx, filepath.Join(root, "native"), c.CWD, session); err != nil {
		return codexbridge.ErrScope
	}
	return nil
}
