package labels

import (
	"errors"
	"net/http"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/effect"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/forge"
)

// Whatever builds a request, the guard sees it before the transport does.
func TestHostRefusesWeakeningRequestsBeforeTheTransportSeesThem(t *testing.T) {
	m := newMini()
	h := ghHost(m, true, t0)
	for _, r := range []forge.Request{
		{Method: http.MethodPut, Path: "/repos/o/r/branches/main/protection"},
		{Method: http.MethodPost, Path: "/repos/o/r/rulesets"},
		{Method: http.MethodPost, Path: "/repos/o/r/hooks"},
		{Method: http.MethodPost, Path: "/repos/o/r/labels", Header: map[string]string{"X-HTTP-Method-Override": "DELETE"}},
	} {
		_, err := h.do(t.Context(), r)
		if !errors.Is(err, effect.ErrRefused) {
			t.Errorf("%s %s: err = %v", r.Method, r.Path, err)
		}
	}
	if len(m.reqs) != 0 {
		t.Errorf("the transport saw %d requests", len(m.reqs))
	}
}

func TestLabelNamesWithSlashesStillReachTheHost(t *testing.T) {
	m := newMini()
	m.labels["kind/bug"] = map[string]string{"name": "kind/bug", "color": "000000", "description": ""}
	h := ghHost(m, true, t0)
	if err := h.Update(t.Context(), "kind/bug", Label{Name: "kind/bug", Color: "ffffff"}); err != nil {
		t.Fatalf("update of kind/bug refused: %v", err)
	}
}
