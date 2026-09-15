package gateway

import (
	"bufio"
	"context"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/codexbridge"
)

func TestAppServerIncrementalStreamBeforeCompletion(t *testing.T) {
	finish := make(chan struct{})
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	response, streamErr := appServerStream(ctx, "test-model", func(ctx context.Context, emit func(string) error) (codexbridge.Segment, error) {
		if err := emit("visible now"); err != nil {
			return codexbridge.Segment{}, err
		}
		select {
		case <-finish:
			return codexbridge.Segment{Text: "visible now", Done: true}, nil
		case <-ctx.Done():
			return codexbridge.Segment{}, ctx.Err()
		}
	})
	if streamErr != nil {
		t.Fatal(streamErr)
	}
	defer func() {
		if err := response.Body.Close(); err != nil {
			t.Errorf("close stream body: %v", err)
		}
	}()
	reader := bufio.NewReader(response.Body)
	var early strings.Builder
	for !strings.Contains(early.String(), "visible now") {
		line, err := reader.ReadString('\n')
		if err != nil {
			t.Fatal(err)
		}
		early.WriteString(line)
	}
	if strings.Contains(early.String(), "message_stop") {
		t.Fatal("terminal before durable completion")
	}
	close(finish)
	rest, err := io.ReadAll(reader)
	if err != nil {
		t.Fatal(err)
	}
	all := early.String() + string(rest)
	if strings.Count(all, "visible now") != 1 || strings.Count(all, "event: message_start") != 1 || !strings.Contains(all, "event: message_stop") {
		t.Fatal(all)
	}
}

func TestAppServerIncrementalStreamErrorHasNoSuccessTerminal(t *testing.T) {
	response, streamErr := appServerStream(context.Background(), "test", func(_ context.Context, emit func(string) error) (codexbridge.Segment, error) {
		if err := emit("partial"); err != nil {
			return codexbridge.Segment{}, err
		}
		return codexbridge.Segment{}, errors.New("private upstream detail")
	})
	if streamErr != nil {
		t.Fatal(streamErr)
	}
	defer func() {
		if err := response.Body.Close(); err != nil {
			t.Errorf("close stream body: %v", err)
		}
	}()
	raw, _ := io.ReadAll(response.Body)
	if strings.Contains(string(raw), "message_stop") || strings.Contains(string(raw), "private upstream detail") || !strings.Contains(string(raw), "event: error") {
		t.Fatal(string(raw))
	}
}

func TestAppServerStreamEarlyRejectionRemainsHTTPError(t *testing.T) {
	response, err := appServerStream(context.Background(), "test", func(context.Context, func(string) error) (codexbridge.Segment, error) {
		return codexbridge.Segment{}, codexbridge.ErrScope
	})
	if response != nil || !errors.Is(err, codexbridge.ErrScope) {
		t.Fatal(response, err)
	}
}

func TestAppServerStreamBodyCloseCancelsProducer(t *testing.T) {
	done := make(chan struct{})
	response, err := appServerStream(context.Background(), "test", func(ctx context.Context, emit func(string) error) (codexbridge.Segment, error) {
		defer close(done)
		if err := emit("blocked text"); err != nil {
			return codexbridge.Segment{}, err
		}
		<-ctx.Done()
		return codexbridge.Segment{}, ctx.Err()
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := response.Body.Close(); err != nil {
		t.Fatal(err)
	}
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("closed stream left producer running")
	}
}

type queueCauseFixture struct{ cause string }

func (e queueCauseFixture) Error() string     { return "CANARY provider private text" }
func (e queueCauseFixture) CauseCode() string { return e.cause }
func (e queueCauseFixture) Unwrap() error     { return codexbridge.ErrLimit }

func TestAppServerErrorClassificationNeverLeaksProviderText(t *testing.T) {
	for _, test := range []struct {
		err    error
		status int
		cause  string
	}{
		{codexbridge.ErrScope, 400, "appserver_scope_mismatch"},
		{codexbridge.ErrRecovery, 400, "appserver_recovery_required"},
		{codexbridge.ErrProtocol, 400, "appserver_protocol_error"},
		{codexbridge.ErrLimit, 400, "appserver_limit_exceeded"},
		{queueCauseFixture{"appserver_event_queue_bytes_exceeded"}, 400, "appserver_event_queue_bytes_exceeded"},
		{queueCauseFixture{"CANARY arbitrary cause"}, 400, "appserver_limit_exceeded"},
		{errors.New("secret private provider text"), 502, "appserver_transport_error"},
	} {
		status, _, cause := appServerError(test.err)
		if status != test.status || cause != test.cause {
			t.Fatal(status, cause)
		}
	}
}

func TestAppServerMeasuredUsageMapsCacheWithoutDoubleCounting(t *testing.T) {
	segment := codexbridge.Segment{Done: true, Usage: &codexbridge.Usage{InputTokens: 100, CachedInputTokens: 60, CacheWriteInputTokens: 10, OutputTokens: 25}}
	for _, stream := range []bool{false, true} {
		response, err := appServerResponse("test", stream, segment)
		if err != nil {
			t.Fatal(err)
		}
		raw, err := io.ReadAll(response.Body)
		if closeErr := response.Body.Close(); closeErr != nil {
			t.Errorf("close response body: %v", closeErr)
		}
		if err != nil {
			t.Fatal(err)
		}
		for _, want := range []string{`"input_tokens":30`, `"cache_read_input_tokens":60`, `"cache_creation_input_tokens":10`, `"output_tokens":25`} {
			if !strings.Contains(string(raw), want) {
				t.Fatal(want, string(raw))
			}
		}
	}
	response, err := appServerStream(context.Background(), "test", func(context.Context, func(string) error) (codexbridge.Segment, error) { return segment, nil })
	if err != nil {
		t.Fatal(err)
	}
	raw, err := io.ReadAll(response.Body)
	if closeErr := response.Body.Close(); closeErr != nil {
		t.Errorf("close stream body: %v", closeErr)
	}
	if err != nil || !strings.Contains(string(raw), `"input_tokens":30`) || !strings.Contains(string(raw), `"output_tokens":25`) {
		t.Fatal(string(raw), err)
	}
}
