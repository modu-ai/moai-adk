package translate

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestGPTNativeImageInputAndToolResult(t *testing.T) {
	image := map[string]any{"type": "image", "source": map[string]string{"type": "base64", "media_type": "image/png", "data": "aW1hZ2U="}}
	tool := map[string]any{"name": "Read", "input_schema": map[string]any{"type": "object"}}
	for _, toolResult := range []bool{false, true} {
		messages := []any{map[string]any{"role": "user", "content": []any{image}}}
		if toolResult {
			messages = []any{map[string]any{"role": "user", "content": "read image"}, map[string]any{"role": "assistant", "content": []any{map[string]any{"type": "tool_use", "id": "read1", "name": "Read", "input": map[string]any{}}}}, map[string]any{"role": "user", "content": []any{map[string]any{"type": "tool_result", "tool_use_id": "read1", "content": []any{map[string]string{"type": "text", "text": "before"}, image, map[string]string{"type": "text", "text": "after"}}}}}}
		}
		body, _ := json.Marshal(map[string]any{"model": "gpt-5.6-sol", "max_tokens": 1024, "messages": messages, "tools": []any{tool}})
		out, _, err := Request("gpt-5.6-sol", body, Limits{PolicyProfile: PolicyGPTNative})
		if err != nil {
			t.Fatalf("toolResult=%v: %v", toolResult, err)
		}
		if !strings.Contains(string(out), `"input_image"`) || !strings.Contains(string(out), "data:image/png;base64,aW1hZ2U=") {
			t.Fatalf("image dropped: %s", out)
		}
		if toolResult && (strings.Index(string(out), "before") >= strings.Index(string(out), "data:image/png") || strings.Index(string(out), "data:image/png") >= strings.Index(string(out), "after")) {
			t.Fatalf("tool content ordering lost: %s", out)
		}
		if _, _, err := Request("gpt-5.6-sol", body, Limits{}); err == nil {
			t.Error("ordinary policy unexpectedly accepts image")
		}
	}
	for _, source := range []map[string]string{{"type": "base64", "media_type": "text/plain", "data": "YQ=="}, {"type": "base64", "media_type": "image/png", "data": "%%%"}, {"type": "url", "url": "file:///etc/passwd"}, {"type": "url", "url": "http://example.com/image.png"}, {"type": "url", "url": "https://user:secret@example.com/a"}} {
		image["source"] = source
		body, _ := json.Marshal(map[string]any{"max_tokens": 100, "messages": []any{map[string]any{"role": "user", "content": []any{image}}}})
		if _, _, err := Request("gpt-5.6-sol", body, Limits{PolicyProfile: PolicyGPTNative}); err == nil {
			t.Errorf("unsafe image accepted: %+v", source)
		}
	}
}

func TestImageSourceValidationBoundary(t *testing.T) {
	for _, raw := range []string{
		`{`,
		`{"type":"image","unexpected":true}`,
		`{"type":"text","source":{}}`,
		`{"type":"image","cache_control":{"type":"unknown"},"source":{}}`,
		`{"type":"image","source":null}`,
		`{"type":"image","source":{"type":"localImage","path":"/etc/passwd"}}`,
		`{"type":"image","source":{"type":"base64","media_type":1,"data":"YQ=="}}`,
		`{"type":"image","source":{"type":"base64","media_type":"image/png","data":""}}`,
		`{"type":"image","source":{"type":"base64","media_type":"image/png","data":"YQ==","extra":true}}`,
		`{"type":"image","source":{"type":"url","url":"https://example.com/a","extra":true}}`,
		`{"type":"image","source":{"type":"url","url":123}}`,
	} {
		if _, err := ImageSourceURL([]byte(raw)); err == nil {
			t.Errorf("invalid source accepted: %s", raw)
		}
	}
	const good = `{"type":"image","source":{"type":"url","url":"https://example.com/image.png"}}`
	if value, err := ImageSourceURL([]byte(good)); err != nil || value != "https://example.com/image.png" {
		t.Fatalf("HTTPS image: %q %v", value, err)
	}
	oversize, _ := json.Marshal(map[string]any{"type": "image", "source": map[string]string{"type": "base64", "media_type": "image/png", "data": strings.Repeat("a", (8<<20)+1)}})
	if _, err := ImageSourceURL(oversize); err == nil {
		t.Error("oversized source accepted")
	}
}
