package gateway

import (
	"context"
	"github.com/modu-ai/moai-adk/internal/gateway/auth"
	"github.com/modu-ai/moai-adk/internal/glmcred"
	"io"
	"net/http"
	"testing"
)

func TestGLMNativeStoredKeyAndBetaRemoval(t *testing.T) {
	t.Setenv("MOAI_HOME", t.TempDir())
	t.Setenv(glmcred.EnvTestGLMKey, "")
	if e := glmcred.Save("synthetic-stored"); e != nil {
		t.Fatal(e)
	}
	ref, e := auth.NewGLMCredential()
	if e != nil {
		t.Fatal(e)
	}
	seen := make(chan *http.Request, 1)
	tr, served := nativeTLS(t, func(w http.ResponseWriter, r *http.Request) {
		// A retried attempt re-enters this handler while the first capture is
		// still buffered; a blocking send wedges the handler goroutine and
		// cleanup. First capture wins; excess ones are dropped.
		select {
		case seen <- r:
		default:
		}
		w.Header().Set("Content-Type", "application/json")
		if _, err := io.WriteString(w, nativeOutput); err != nil {
			t.Error(err)
		}
	})
	a, e := NewGLMAdapter(nativeConfig(tr))
	if e != nil {
		t.Fatal(e)
	}
	q := nativeQ(t)
	q.Entry.Provider = ProviderZAI
	q.Entry.AuthMethod = AuthExistingGLM
	q.Credential = ref
	r, e := a.Send(context.Background(), q)
	_ = oaiRead(t, r, e)
	request := <-seen
	if request.Host != "api.z.ai" || request.URL.Path != "/api/anthropic/v1/messages" || request.Header.Get("Authorization") != "Bearer synthetic-stored" || request.Header.Get("Anthropic-Beta") != "" || request.Header.Get("X-Api-Key") != "" {
		t.Fatal("GLM boundary")
	}
	if e = glmcred.Save("rotated"); e != nil {
		t.Fatal(e)
	}
	r, e = a.Send(context.Background(), q)
	_ = oaiRead(t, r, e)
	if r.StatusCode != 401 || served.Load() != 1 {
		t.Fatal("rotated key sent")
	}
}
