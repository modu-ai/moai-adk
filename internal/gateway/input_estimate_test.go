package gateway

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestEstimateInputMatchesCountProjection(t *testing.T) {
	for _, tc := range []struct {
		body string
		want int64
	}{
		{`{}`, 1},
		{`{"messages":[],"model":"ignored","max_tokens":999}`, 4},
		{`{"system":"가나","messages":[],"tools":[]}`, 10},
	} {
		got, e := EstimateInputTokens(ModelEntry{}, []byte(tc.body))
		if e != nil || got != tc.want {
			t.Fatalf("estimate %s = %d, %v; want %d", tc.body, got, e, tc.want)
		}
		var p map[string]json.RawMessage
		_ = json.Unmarshal([]byte(tc.body), &p)
		w := httptest.NewRecorder()
		writeCount(w, ModelEntry{Provider: ProviderOpenAI}, p)
		var out struct {
			Count     int64  `json:"input_tokens"`
			Accuracy  string `json:"accuracy"`
			Algorithm string `json:"algorithm"`
		}
		if json.Unmarshal(w.Body.Bytes(), &out) != nil || out.Count != got || out.Accuracy != "estimate" || out.Algorithm != "json-runes-div4" {
			t.Fatal(w.Body.String())
		}
	}
}

func TestOpenAIEstimatedOverflowAndFullInputUpstream400(t *testing.T) {
	first, last := strings.Repeat("처음", 300), strings.Repeat("끝", 300)
	raw, _ := json.Marshal(map[string]any{"model": "gpt-5.6-sol", "max_tokens": 10, "system": "system preserved", "messages": []any{map[string]string{"role": "user", "content": first}, map[string]string{"role": "assistant", "content": "middle"}, map[string]string{"role": "user", "content": last}}})
	tr, served := oaiTLS(t, func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Instructions string `json:"instructions"`
			Truncation   string `json:"truncation"`
			Input        []struct {
				Role    string `json:"role"`
				Content []struct {
					Text string `json:"text"`
				} `json:"content"`
			} `json:"input"`
		}
		if json.NewDecoder(r.Body).Decode(&body) != nil || len(body.Input) != 3 {
			t.Error("full input missing")
		} else if body.Input[0].Content[0].Text != first || body.Input[1].Content[0].Text != "middle" || body.Input[2].Content[0].Text != last || body.Instructions != "system preserved" || body.Truncation != "disabled" {
			t.Error("public input changed")
		}
		w.WriteHeader(400)
		if _, err := io.WriteString(w, `{"error":{"message":"synthetic upstream private details"}}`); err != nil {
			t.Error(err)
		}
	})
	cfg := oaiConfig(tr)
	cfg.MeasureInput = EstimateInputTokens
	a, e := NewOpenAIAdapter(cfg)
	if e != nil {
		t.Fatal(e)
	}
	q := oaiRequest(t)
	q.Body = raw
	estimate, e := EstimateInputTokens(q.Entry, raw)
	if e != nil {
		t.Fatal(e)
	}
	q.Entry.Capabilities.ContextTokens = int(estimate) - 1
	r, e := a.Send(context.Background(), q)
	_ = oaiRead(t, r, e)
	if r.StatusCode != 400 || served.Load() != 0 {
		t.Fatal("estimated overflow crossed egress", r.StatusCode, served.Load())
	}
	q.Entry.Capabilities.ContextTokens = int(estimate)
	r, e = a.Send(context.Background(), q)
	body := oaiRead(t, r, e)
	if r.StatusCode != 400 || served.Load() != 1 || strings.Contains(body, "private") || !strings.Contains(body, "invalid_request_error") {
		t.Fatal("upstream context rejection not preserved safely", r.StatusCode, served.Load(), body)
	}
}

func TestEstimateInputRejectsAmbiguousJSON(t *testing.T) {
	for _, body := range [][]byte{[]byte(`{"messages":[],"messages":[]}`), []byte(`{"system":"\ud800"}`), []byte(`[]`), []byte("{\"system\":\"\xff\"}")} {
		if _, e := EstimateInputTokens(ModelEntry{}, body); e == nil {
			t.Fatal("invalid JSON accepted")
		}
	}
}
