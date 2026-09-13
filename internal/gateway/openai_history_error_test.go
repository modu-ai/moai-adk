package gateway

import (
	"errors"
	"fmt"
	"github.com/modu-ai/moai-adk/internal/gateway/translate"
	"strings"
	"testing"
)

// History replay keeps the guided-recovery message without the wrapping chain;
// ordinary translation failures surface their validation reason (card t695 D1).
func TestOpenAITranslationErrorSurfacesReasonWithoutWrappingChain(t *testing.T) {
	r := openAITranslationError(fmt.Errorf("secret detail: %w", translate.HistoryReplayError{}))
	body := oaiRead(t, r, nil)
	if r.StatusCode != 400 || !strings.Contains(body, "start a new conversation") || strings.Contains(body, "secret detail") {
		t.Fatal(body)
	}
	r = openAITranslationError(errors.New("unsupported effort"))
	body = oaiRead(t, r, nil)
	if r.StatusCode != 400 || !strings.Contains(body, "unsupported effort") || !strings.Contains(body, "invalid_request_error") {
		t.Fatal(body)
	}
}
