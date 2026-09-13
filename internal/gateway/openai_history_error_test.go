package gateway

import (
	"errors"
	"fmt"
	"github.com/modu-ai/moai-adk/internal/gateway/translate"
	"strings"
	"testing"
)

func TestOpenAIHistoryErrorGuidesRecoveryWithoutExposingDetails(t *testing.T) {
	r := openAITranslationError(fmt.Errorf("secret detail: %w", translate.HistoryReplayError{}))
	body := oaiRead(t, r, nil)
	if r.StatusCode != 400 || !strings.Contains(body, "start a new conversation") || strings.Contains(body, "secret detail") {
		t.Fatal(body)
	}
	r = openAITranslationError(errors.New("secret detail"))
	body = oaiRead(t, r, nil)
	if strings.Contains(body, "secret detail") || !strings.Contains(body, "Bad Request") {
		t.Fatal(body)
	}
}
