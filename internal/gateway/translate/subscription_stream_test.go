package translate

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"testing"
	"testing/iotest"
	"time"
)

func sparseTerminal(t *testing.T, events []string, absent bool) []string {
	t.Helper()
	out := append([]string(nil), events...)
	last := out[len(out)-1]
	raw := strings.TrimSpace(strings.SplitN(last, "data: ", 2)[1])
	var m map[string]any
	if err := json.Unmarshal([]byte(raw), &m); err != nil {
		t.Fatal(err)
	}
	response := m["response"].(map[string]any)
	if absent {
		delete(response, "output")
	} else {
		response["output"] = []any{}
	}
	typ := m["type"].(string)
	out[len(out)-1] = event(typ, m)
	return out
}
func TestSubscriptionStreamReconstructsOnlyValidatedOrderedItems(t *testing.T) {
	for _, tools := range []bool{false, true} {
		for _, absent := range []bool{false, true} {
			c := responseCtx(t)
			events := textEvents("completed")
			if tools {
				events = toolEvents(c)
			}
			var expected, actual bytes.Buffer
			if err := c.Stream(context.Background(), io.NopCloser(strings.NewReader(strings.Join(events, ""))), &expected); err != nil {
				t.Fatal(err)
			}
			raw := strings.Join(sparseTerminal(t, events, absent), "")
			if err := c.SubscriptionStream(context.Background(), io.NopCloser(strings.NewReader(raw)), &actual); err != nil {
				t.Fatal(err)
			}
			if actual.String() != expected.String() {
				t.Fatal("sparse terminal changed ordered content")
			}
			var strict bytes.Buffer
			if err := c.Stream(context.Background(), io.NopCloser(strings.NewReader(raw)), &strict); err == nil {
				t.Fatal("API stream gained sparse exception")
			}
		}
	}
}
func TestSubscriptionStreamSparseTerminalDoesNotRepairInvalidEvents(t *testing.T) {
	c := responseCtx(t)
	events := sparseTerminal(t, textEvents("completed"), false)
	cases := map[string]string{
		"missing item done":    strings.Join(append(append([]string{}, events[:6]...), events[7:]...), ""),
		"duplicate done":       strings.Join(events[:7], "") + events[6] + events[7],
		"wrong identity":       strings.Join(events[:6], "") + strings.ReplaceAll(events[6], "item-t", "other") + events[7],
		"wrong index":          strings.Join(events[:6], "") + strings.ReplaceAll(events[6], `"output_index":0`, `"output_index":1`) + events[7],
		"text mismatch":        strings.Join(events[:6], "") + strings.ReplaceAll(events[6], "hello", "changed") + events[7],
		"no public":            events[0] + events[7],
		"HTML":                 "<html>not SSE</html>",
		"invalid UTF8 comment": ": " + string([]byte{255}) + "\n\n" + strings.Join(events, ""),
		"null output":          strings.Join(events[:7], "") + strings.ReplaceAll(events[7], `"output":[]`, `"output":null`),
		"explicit mismatch":    strings.Join(events[:7], "") + event("response.completed", map[string]any{"response": finalResponse("completed", textItem("item-t", "different"))}),
	}
	for i := 0; i < len(events); i++ {
		cases[string(rune('a'+i))+" truncated"] = strings.Join(events[:i], "")
	}
	for name, raw := range cases {
		t.Run(name, func(t *testing.T) {
			var out bytes.Buffer
			if err := c.SubscriptionStream(context.Background(), io.NopCloser(strings.NewReader(raw)), &out); err == nil || strings.Contains(out.String(), "event: message_stop") {
				t.Fatal("invalid stream completed")
			}
		})
	}
}

func TestSubscriptionStreamBoundsCancelAndTerminalOrder(t *testing.T) {
	c := responseCtx(t)
	events := toolEvents(c)
	reversed := finalResponse("completed", toolItem("item-1", "call-1", alias(c, "a_b"), `{"b":2}`), toolItem("item-0", "call-0", alias(c, "a.b"), `{"a":1}`))
	raw := strings.Join(events[:len(events)-1], "") + event("response.completed", map[string]any{"response": reversed})
	var out bytes.Buffer
	if err := c.SubscriptionStream(context.Background(), io.NopCloser(strings.NewReader(raw)), &out); err == nil || strings.Contains(out.String(), "message_stop") {
		t.Fatal("reordered terminal accepted")
	}
	for _, limit := range []string{"event", "total"} {
		c := responseCtx(t)
		if limit == "event" {
			c.limits.MaxEventBytes = 64
		} else {
			c.limits.MaxOutputBytes = 64
		}
		var out bytes.Buffer
		if err := c.SubscriptionStream(context.Background(), io.NopCloser(strings.NewReader(strings.Join(textEvents("completed"), ""))), &out); err == nil || strings.Contains(out.String(), "message_stop") {
			t.Fatal("limit not enforced")
		}
	}
	reader, writer := io.Pipe()
	defer writer.Close()
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- c.SubscriptionStream(ctx, reader, io.Discard) }()
	cancel()
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		t.Fatal("cancel blocked")
	}
}

func TestStreamRawWireByteBoundaries(t *testing.T) {
	for _, subscription := range []bool{false, true} {
		events := textEvents("completed")
		if subscription {
			events = sparseTerminal(t, events, false)
		}
		lf := strings.Join(events, "")
		mixed := strings.Replace(lf, "\n", "\r\n", 3)
		for name, raw := range map[string]string{"LF": lf, "CRLF": strings.ReplaceAll(lf, "\n", "\r\n"), "mixed": mixed, "embedded CR": ": a\rb\r\n" + mixed} {
			for _, offset := range []int{-1, 0, 1} {
				c := responseCtx(t)
				c.limits.MaxOutputBytes = len(raw) + offset
				var out bytes.Buffer
				var err error
				if subscription {
					err = c.SubscriptionStream(context.Background(), io.NopCloser(iotest.OneByteReader(strings.NewReader(raw))), &out)
				} else {
					err = c.Stream(context.Background(), io.NopCloser(iotest.OneByteReader(strings.NewReader(raw))), &out)
				}
				terminal := strings.Contains(out.String(), "event: message_stop")
				if offset < 0 && (err == nil || terminal) {
					t.Errorf("%s subscription=%t limit offset=%d accepted excess wire bytes", name, subscription, offset)
				}
				if offset >= 0 && (err != nil || !terminal) {
					t.Errorf("%s subscription=%t offset=%d exact/under limit failed: %v", name, subscription, offset, err)
				}
			}
		}
	}
}
