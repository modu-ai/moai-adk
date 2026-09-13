package gateway

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestOpenAISubscriptionMediaBoundary(t *testing.T) {
	for _, tc := range []struct {
		h                        http.Header
		subscription, stream, ok bool
	}{
		{nil, true, true, true}, {nil, false, true, false}, {nil, true, false, false},
		{http.Header{"Content-Type": {""}}, true, true, false},
		{http.Header{"Content-Type": {"text/event-stream"}}, true, true, true},
		{http.Header{"Content-Type": {"text/html"}}, true, true, false},
		{http.Header{"Content-Type": {"text/event-stream", "text/event-stream"}}, true, true, false},
		{http.Header{"Content-Type": {"text/event-stream"}, "content-type": {"application/json"}}, true, true, false},
		{http.Header{"Content-Type": {"text/event-stream;broken"}}, true, true, false},
		{http.Header{"Content-Encoding": {"br"}}, true, true, false},
	} {
		_, err := openAIResponseMedia(tc.h, tc.subscription, tc.stream)
		if (err == nil) != tc.ok {
			t.Fatalf("media boundary subscription=%v stream=%v", tc.subscription, tc.stream)
		}
	}
}
func TestOpenAISubscriptionAbsentMediaStrictSSE(t *testing.T) {
	store, ref, gen := oaiStore(t)
	for _, valid := range []bool{true, false} {
		tr, calls := oaiTLS(t, func(w http.ResponseWriter, r *http.Request) {
			w.Header()["Content-Type"] = nil
			w.WriteHeader(200)
			if !valid {
				if _, err := io.WriteString(w, "<html>invalid</html>"); err != nil {
					t.Error(err)
				}
				return
			}
			var final map[string]any
			_ = json.Unmarshal([]byte(oaiOutput), &final)
			final["output"] = []any{}
			b, _ := json.Marshal(final)
			if _, err := io.WriteString(w, strings.Replace(oaiSSE(), oaiOutput, string(b), 1)); err != nil {
				t.Error(err)
			}
		})
		cfg := oaiConfig(tr)
		cfg.Subscription = store
		a, err := NewOpenAIAdapter(cfg)
		if err != nil {
			t.Fatal(err)
		}
		q := RoutedRequest{Entry: oaiEntry(AuthPKCE), Credential: ref, Generation: gen, Body: []byte(strings.Replace(oaiInput, `"max_tokens":10`, `"stream":true,"max_tokens":10`, 1))}
		r, err := a.Send(context.Background(), q)
		if err != nil {
			t.Fatal(err)
		}
		body, readErr := io.ReadAll(r.Body)
		if cerr := r.Body.Close(); cerr != nil {
			t.Fatal(cerr)
		}
		if valid && (r.StatusCode != 200 || readErr != nil || !strings.Contains(string(body), "message_stop")) {
			t.Fatalf("valid subscription SSE failed: %d %v", r.StatusCode, readErr)
		}
		if !valid && readErr == nil {
			t.Fatal("arbitrary response admitted")
		}
		if calls.Load() != 1 {
			t.Fatal("unexpected retry")
		}
	}
}

func TestOpenAISubscriptionNonStreamCollectsSSE(t *testing.T) {
	store, ref, gen := oaiStore(t)
	tr, calls := oaiTLS(t, func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("invalid outgoing JSON: %v", err)
		}
		if body["stream"] != true {
			t.Errorf("subscription request was not forced to stream: %#v", body)
		}
		w.Header().Set("Content-Type", "text/event-stream")
		if _, err := io.WriteString(w, oaiSSE()); err != nil {
			t.Error(err)
		}
	})
	cfg := oaiConfig(tr)
	cfg.Subscription = store
	a, err := NewOpenAIAdapter(cfg)
	if err != nil {
		t.Fatal(err)
	}
	r, err := a.Send(context.Background(), RoutedRequest{
		Entry: oaiEntry(AuthPKCE), Body: []byte(oaiInput),
		Credential: ref, Generation: gen,
	})
	body := oaiRead(t, r, err)
	if r.StatusCode != http.StatusOK || !strings.Contains(body, `"stop_reason":"end_turn"`) || strings.Contains(body, "event:") {
		t.Fatalf("non-stream subscription response was not collected: %d %s", r.StatusCode, body)
	}
	if calls.Load() != 1 {
		t.Fatalf("unexpected upstream calls: %d", calls.Load())
	}
}

func TestOpenAISubscriptionOutputPolicyWireAndValidation(t *testing.T) {
	store, ref, gen := oaiStore(t)
	for _, method := range []AuthMethod{AuthPKCE, AuthAPIKey} {
		tr, calls := oaiTLS(t, func(w http.ResponseWriter, r *http.Request) {
			var body map[string]json.RawMessage
			if json.NewDecoder(r.Body).Decode(&body) != nil {
				t.Error("invalid outgoing JSON")
			}
			cap, exists := body["max_output_tokens"]
			if method == AuthPKCE {
				if exists || r.Host != "chatgpt.com" || r.URL.Path != "/backend-api/codex/responses" {
					t.Error("subscription policy or endpoint mismatch")
				}
			} else if !exists || string(cap) != "10" || r.Host != "api.openai.com" || r.URL.Path != "/v1/responses" {
				t.Error("API key policy or endpoint mismatch")
			}
			if method == AuthPKCE {
				w.Header().Set("Content-Type", "text/event-stream")
				if _, err := io.WriteString(w, oaiSSE()); err != nil {
					t.Error(err)
				}
			} else {
				w.Header().Set("Content-Type", "application/json")
				if _, err := io.WriteString(w, oaiOutput); err != nil {
					t.Error(err)
				}
			}
		})
		cfg := oaiConfig(tr)
		cfg.Subscription = store
		a, e := NewOpenAIAdapter(cfg)
		if e != nil {
			t.Fatal(e)
		}
		q := oaiRequest(t)
		if method == AuthPKCE {
			q.Entry = oaiEntry(method)
			q.Credential = ref
			q.Generation = gen
		}
		r, e := a.Send(context.Background(), q)
		oaiRead(t, r, e)
		if calls.Load() != 1 {
			t.Fatal("valid request not sent once")
		}
		for _, value := range []string{"0", "-1", "1.5", `"10"`, "null", "9223372036854775808", "1e100"} {
			q.Body = []byte(strings.Replace(oaiInput, `"max_tokens":10`, `"max_tokens":`+value, 1))
			r, e = a.Send(context.Background(), q)
			if e != nil || r.StatusCode != 400 {
				t.Fatal("invalid max_tokens accepted")
			}
			if cerr := r.Body.Close(); cerr != nil {
				t.Fatal(cerr)
			}
		}
		for _, body := range []string{strings.Replace(oaiInput, `"max_tokens":10,`, "", 1), strings.Replace(oaiInput, `"max_tokens":10`, `"max_tokens":0,"max_tokens":10`, 1), strings.Replace(oaiInput, `"max_tokens":10`, `"max_tokens":10,"max_tokens":0`, 1), strings.Repeat(" ", cfg.Limits.MaxBodyBytes) + oaiInput} {
			q.Body = []byte(body)
			r, e = a.Send(context.Background(), q)
			if e != nil || r.StatusCode != 400 {
				t.Fatal("missing/duplicate/oversized input accepted")
			}
			if cerr := r.Body.Close(); cerr != nil {
				t.Fatal(cerr)
			}
		}
		if calls.Load() != 1 {
			t.Fatal("invalid input reached external transport")
		}
	}
}
