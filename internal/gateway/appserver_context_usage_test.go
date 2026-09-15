package gateway

import (
	"context"
	"encoding/json"
	"github.com/modu-ai/moai-adk/internal/codexbridge"
	"io"
	"strings"
	"testing"
)

func TestAppServerContextInputAndSegmentOutputRemainSeparate(t *testing.T) {
	segment := codexbridge.Segment{Usage: &codexbridge.Usage{InputTokens: 831857, CachedInputTokens: 828416, OutputTokens: 448}, ContextUsage: &codexbridge.Usage{InputTokens: 208730, CachedInputTokens: 208384, OutputTokens: 116}}
	got := appServerSegmentUsage(segment)
	if got["input_tokens"] != 346 || got["cache_read_input_tokens"] != 208384 || got["output_tokens"] != 448 {
		t.Fatalf("occupancy and billing conflated: %+v", got)
	}
	if segment.Usage.InputTokens != 831857 {
		t.Fatal("billing mutated")
	}
	unknown := appServerSegmentUsage(codexbridge.Segment{})
	if unknown["input_tokens"] != 0 || unknown["output_tokens"] != 0 {
		t.Fatal("fabricated unknown usage")
	}
}

func TestAppServerContextUsageReachesBufferedAndLiveStreamingResponses(t *testing.T) {
	segment := codexbridge.Segment{Done: true, Text: "done", Usage: &codexbridge.Usage{InputTokens: 831857, OutputTokens: 448}, ContextUsage: &codexbridge.Usage{InputTokens: 208730, OutputTokens: 116}}
	buffered, err := appServerResponse("gpt-5.6-sol", false, segment)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if closeErr := buffered.Body.Close(); closeErr != nil {
			t.Errorf("close buffered body: %v", closeErr)
		}
	}()
	var result struct {
		Usage map[string]int64 `json:"usage"`
	}
	if err = json.NewDecoder(buffered.Body).Decode(&result); err != nil {
		t.Fatal(err)
	}
	if result.Usage["input_tokens"] != 208730 || result.Usage["output_tokens"] != 448 {
		t.Fatalf("buffered=%+v", result.Usage)
	}
	stream, err := appServerStream(context.Background(), "gpt-5.6-sol", func(context.Context, func(string) error) (codexbridge.Segment, error) { return segment, nil })
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if closeErr := stream.Body.Close(); closeErr != nil {
			t.Errorf("close stream body: %v", closeErr)
		}
	}()
	raw, err := io.ReadAll(stream.Body)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, line := range strings.Split(string(raw), "\n") {
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		var event struct {
			Type  string           `json:"type"`
			Usage map[string]int64 `json:"usage"`
		}
		if err = json.Unmarshal([]byte(strings.TrimPrefix(line, "data: ")), &event); err != nil {
			t.Fatal(err)
		}
		if event.Type == "message_delta" {
			found = true
			if event.Usage["input_tokens"] != 208730 || event.Usage["output_tokens"] != 448 {
				t.Fatalf("stream=%+v", event.Usage)
			}
		}
	}
	if !found {
		t.Fatal("no terminal usage")
	}
	bufferedStream, err := appServerResponse("gpt-5.6-sol", true, segment)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if closeErr := bufferedStream.Body.Close(); closeErr != nil {
			t.Errorf("close buffered stream body: %v", closeErr)
		}
	}()
	raw, err = io.ReadAll(bufferedStream.Body)
	if err != nil || !strings.Contains(string(raw), `"input_tokens":208730`) {
		t.Fatalf("buffered stream lost context: %s %v", raw, err)
	}
}
