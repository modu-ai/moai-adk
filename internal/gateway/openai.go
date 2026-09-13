package gateway

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/modu-ai/moai-adk/internal/gateway/auth"
	"github.com/modu-ai/moai-adk/internal/gateway/translate"
)

// OpenAIConfig has no implicit login, API key, model capability or token estimate.
// MeasureInput must compute locally without egress; EstimateInputTokens provides
// an explicitly approximate option, not an exact token count. Subscription references must
// belong to the configured Store, checked with Store.Owns before sending.
type OpenAIConfig struct {
	Subscription *auth.Store
	Transport    *http.Transport
	Limits       translate.Limits
	MeasureInput func(ModelEntry, []byte) (int64, error)
	SendOptions  auth.SendOptions
}
type OpenAIAdapter struct{ config OpenAIConfig }

// NewOpenAIAdapter freezes transport settings and never reads proxy environment.
func NewOpenAIAdapter(c OpenAIConfig) (*OpenAIAdapter, error) {
	if c.Transport == nil || c.Transport.Proxy != nil || c.Transport.TLSClientConfig != nil && c.Transport.TLSClientConfig.InsecureSkipVerify {
		return nil, errors.New("OpenAI transport requires verified TLS without a proxy")
	}
	c.Transport = c.Transport.Clone()
	c.Transport.DisableKeepAlives = true
	c.Transport.ForceAttemptHTTP2 = false
	if c.Transport.ResponseHeaderTimeout <= 0 {
		c.Transport.ResponseHeaderTimeout = 30 * time.Second
	}
	return &OpenAIAdapter{config: c}, nil
}

// Send translates and validates before touching credentials or the fixed provider
// endpoint. Local failures are protocol responses, so Server preserves their 4xx.
// @MX:ANCHOR: [AUTO] Ordinary OpenAI adapter egress boundary.
// @MX:REASON: API-key and subscription paths share validation, never credentials.
func (a *OpenAIAdapter) Send(ctx context.Context, q RoutedRequest) (*http.Response, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if q.Entry.Provider != ProviderOpenAI || q.Entry.UpstreamID == "" {
		return openAIError(400), nil
	}
	limits := a.config.Limits
	if q.Entry.Capabilities.ContextTokens <= 0 || a.config.MeasureInput == nil {
		return openAIError(400), nil
	}
	measured, e := a.config.MeasureInput(q.Entry, q.Body)
	if e != nil {
		return openAIError(400), nil
	}
	limits.ContextTokens = int64(q.Entry.Capabilities.ContextTokens)
	limits.InputTokens = &measured
	if limits.History != nil {
		if a.config.Subscription == nil || q.Entry.AuthMethod != AuthPKCE {
			return openAIError(400), nil
		}
		var err error
		limits.CredentialScope, err = a.config.Subscription.ReplayScope(ctx, q.Credential)
		if err != nil {
			return openAIError(401), nil
		}
	}
	encoded, conversion, e := translate.RequestContext(ctx, q.Entry.UpstreamID, q.Body, limits)
	if e != nil {
		return openAITranslationError(e), nil
	}
	var requestShape struct {
		Stream bool              `json:"stream"`
		Tools  []json.RawMessage `json:"tools"`
		Input  []struct {
			Type string `json:"type"`
		} `json:"input"`
	}
	if json.Unmarshal(encoded, &requestShape) != nil {
		return openAIError(400), nil
	}
	clientStream := requestShape.Stream
	upstreamStream := requestShape.Stream
	if requestShape.Stream && !q.Entry.Capabilities.Streaming {
		return openAIError(400), nil
	}
	hasTools := len(requestShape.Tools) > 0
	for _, item := range requestShape.Input {
		hasTools = hasTools || item.Type == "function_call" || item.Type == "function_call_output"
	}
	if hasTools && !q.Entry.Capabilities.Tools {
		return openAIError(400), nil
	}
	if q.Credential == nil || q.Credential.Provider() != ProviderOpenAI {
		return openAIError(401), nil
	}
	current, e := q.Credential.Generation()
	if e != nil || current != q.Generation {
		return openAIError(401), nil
	}
	endpoint := auth.APIEndpoint
	switch q.Entry.AuthMethod {
	case AuthAPIKey:
		// The public API rejects oversized context instead of dropping input. Do
		// not send this field to the separately gated subscription endpoint.
		var body map[string]json.RawMessage
		if json.Unmarshal(encoded, &body) != nil {
			return openAIError(400), nil
		}
		body["truncation"] = json.RawMessage(`"disabled"`)
		encoded, e = marshalRawObject(body)
		if e != nil {
			return openAIError(400), nil
		}
	case AuthPKCE:
		if a.config.Subscription == nil || !a.config.Subscription.Owns(q.Credential) {
			return openAIError(401), nil
		}
		// The fixed subscription endpoint rejects max_output_tokens. Claude's
		// positive integer input was already validated above; this route uses the
		// subscription server's output policy, not a per-request generated-token cap.
		var body map[string]json.RawMessage
		if json.Unmarshal(encoded, &body) != nil {
			return openAIError(400), nil
		}
		delete(body, "max_output_tokens")
		if !upstreamStream {
			// The fixed subscription endpoint accepts streaming requests only.
			// Collect its validated SSE below when Claude requested JSON.
			body["stream"] = json.RawMessage("true")
			upstreamStream = true
		}
		encoded, e = marshalRawObject(body)
		if e != nil {
			return openAIError(400), nil
		}
		endpoint = auth.SubscriptionEndpoint
	default:
		return openAIError(400), nil
	}
	req, e := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(encoded))
	if e != nil {
		return nil, errors.New("OpenAI request construction failed")
	}
	req.Header.Set("Content-Type", "application/json")
	if upstreamStream {
		req.Header.Set("Accept", "text/event-stream")
	} else {
		req.Header.Set("Accept", "application/json")
	}
	var upstream *http.Response
	if q.Entry.AuthMethod == AuthPKCE {
		upstream, e = a.config.Subscription.SendAuthorized(ctx, q.Generation, req, a.config.Transport, a.config.SendOptions)
	} else {
		if e = q.Credential.Apply(req); e != nil {
			return openAIError(401), nil
		}
		// A credential implementation must not redirect the request.
		if req.URL == nil || req.URL.String() != auth.APIEndpoint || req.Host != "api.openai.com" && req.Host != "" || req.Method != http.MethodPost {
			return openAIError(401), nil
		}
		current, e = q.Credential.Generation()
		if e != nil || current != q.Generation {
			return openAIError(401), nil
		}
		upstream, e = a.config.Transport.RoundTrip(req)
	}
	if e != nil {
		if upstream != nil && upstream.Body != nil {
			_ = upstream.Body.Close()
		}
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		if errors.Is(e, auth.ErrCredentialAbsent) || errors.Is(e, auth.ErrCredentialChanged) || errors.Is(e, auth.ErrWrongProvider) {
			return openAIError(401), nil
		}
		return openAIError(502), nil
	}
	if upstream == nil || upstream.Body == nil {
		return openAIError(502), nil
	}
	if upstream.StatusCode != http.StatusOK {
		_ = upstream.Body.Close()
		status := upstream.StatusCode
		if status < 400 || status > 599 {
			status = 502
		}
		r := openAIError(status)
		if status == 429 || status == 503 {
			if retry := validRetryAfter(upstream.Header.Get("Retry-After")); retry != "" {
				r.Header.Set("Retry-After", retry)
			}
		}
		return r, nil
	}
	media, e := openAIResponseMedia(upstream.Header, q.Entry.AuthMethod == AuthPKCE, upstreamStream)
	if e != nil {
		_ = upstream.Body.Close()
		return openAIError(502), nil
	}
	if upstreamStream {
		if media != "text/event-stream" {
			_ = upstream.Body.Close()
			return openAIError(502), nil
		}
		if clientStream {
			return &http.Response{StatusCode: 200, Header: http.Header{"Content-Type": {"text/event-stream"}}, Body: convertedOpenAIStream(ctx, upstream.Body, conversion, q.Entry.AuthMethod == AuthPKCE)}, nil
		}
		converted, e := conversion.SubscriptionResponse(ctx, upstream.Body)
		if e != nil {
			return openAIError(502), nil
		}
		return &http.Response{StatusCode: 200, Header: http.Header{"Content-Type": {"application/json"}}, Body: io.NopCloser(bytes.NewReader(converted)), ContentLength: int64(len(converted))}, nil
	}
	defer upstream.Body.Close()
	if media != "application/json" {
		return openAIError(502), nil
	}
	max := limits.MaxOutputBytes
	if max <= 0 {
		max = 8 << 20
	}
	body, e := io.ReadAll(io.LimitReader(upstream.Body, int64(max)+1))
	if e != nil || len(body) > max {
		return openAIError(502), nil
	}
	converted, e := conversion.Response(body)
	if e != nil {
		return openAIError(502), nil
	}
	return &http.Response{StatusCode: 200, Header: http.Header{"Content-Type": {"application/json"}}, Body: io.NopCloser(bytes.NewReader(converted)), ContentLength: int64(len(converted))}, nil
}

// openAIResponseMedia preserves the API-key contract. Only fixed subscription
// streaming may omit Content-Type; the strict SSE parser must still complete.
func openAIResponseMedia(headers http.Header, subscription, stream bool) (string, error) {
	if !subscription {
		media, _, err := mime.ParseMediaType(headers.Get("Content-Type"))
		return media, err
	}
	var types, encodings []string
	typePresent := false
	for key, values := range headers {
		if strings.EqualFold(key, "Content-Type") {
			typePresent = true
			types = append(types, values...)
		}
		if strings.EqualFold(key, "Content-Encoding") {
			encodings = append(encodings, values...)
		}
	}
	if len(encodings) > 1 || len(encodings) == 1 && !strings.EqualFold(strings.TrimSpace(encodings[0]), "identity") {
		return "", errors.New("unsupported subscription encoding")
	}
	if !typePresent && stream {
		return "text/event-stream", nil
	}
	if len(types) != 1 {
		return "", errors.New("ambiguous subscription media")
	}
	media, _, err := mime.ParseMediaType(types[0])
	if err != nil || stream && media != "text/event-stream" || !stream && media != "application/json" {
		return "", errors.New("unsupported subscription media")
	}
	return media, nil
}

// openAITranslationError surfaces the translate failure reason in the 400 body.
// Translation errors are this adapter's own validation strings (never upstream
// payloads or credentials), so a reason-less "Bad Request" only masks the cause
// from client debug logs. History replay keeps its guided-recovery message.
func openAITranslationError(err error) *http.Response {
	var replay translate.HistoryReplayError
	if errors.As(err, &replay) {
		return openAIErrorMessage(400, replay.Error())
	}
	return openAIErrorMessage(400, err.Error())
}
func openAIError(status int) *http.Response {
	return openAIErrorMessage(status, http.StatusText(status))
}
func openAIErrorMessage(status int, message string) *http.Response {
	kind := "api_error"
	switch status {
	case 400, 413, 422:
		kind = "invalid_request_error"
	case 401:
		kind = "authentication_error"
	case 403:
		kind = "permission_error"
	case 404:
		kind = "not_found_error"
	case 429:
		kind = "rate_limit_error"
	}
	b, _ := json.Marshal(map[string]any{"type": "error", "error": map[string]string{"type": kind, "message": message}})
	return &http.Response{StatusCode: status, Header: http.Header{"Content-Type": {"application/json"}}, Body: io.NopCloser(bytes.NewReader(b)), ContentLength: int64(len(b))}
}
func validRetryAfter(value string) string {
	if len(value) > 64 || strings.ContainsAny(value, "\r\n") {
		return ""
	}
	if n, e := strconv.ParseUint(value, 10, 32); e == nil {
		return strconv.FormatUint(n, 10)
	}
	if t, e := http.ParseTime(value); e == nil {
		return t.UTC().Format(http.TimeFormat)
	}
	return ""
}

type openAIStreamBody struct {
	reader *io.PipeReader
	cancel context.CancelFunc
	done   chan struct{}
	once   sync.Once
}

// @MX:WARN: [AUTO] A pipe converter runs until completion or caller cancellation.
// @MX:REASON: Close cancels upstream, closes the reader and joins the converter.
func convertedOpenAIStream(ctx context.Context, upstream io.ReadCloser, c *translate.ResponseContext, subscription bool) io.ReadCloser {
	ctx, cancel := context.WithCancel(ctx)
	reader, writer := io.Pipe()
	b := &openAIStreamBody{reader: reader, cancel: cancel, done: make(chan struct{})}
	go func() {
		defer close(b.done)
		defer cancel()
		var e error
		if subscription {
			e = c.SubscriptionStream(ctx, upstream, writer)
		} else {
			e = c.Stream(ctx, upstream, writer)
		}
		if e != nil {
			e = errors.New("OpenAI response stream failed")
		}
		_ = writer.CloseWithError(e)
	}()
	return b
}
func (b *openAIStreamBody) Read(p []byte) (int, error) {
	n, e := b.reader.Read(p)
	if e != nil {
		_ = b.Close()
	}
	return n, e
}
func (b *openAIStreamBody) Close() error {
	b.once.Do(func() { b.cancel(); _ = b.reader.Close(); <-b.done })
	return nil
}

// marshalRawObject preserves raw values through final endpoint policy changes.
// encoding/json compacts RawMessage, which would alter receipt-bound item bytes.
func marshalRawObject(body map[string]json.RawMessage) ([]byte, error) {
	keys := make([]string, 0, len(body))
	for key := range body {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	var out bytes.Buffer
	out.WriteByte('{')
	for i, key := range keys {
		value := body[key]
		if !json.Valid(value) {
			return nil, errors.New("invalid raw request value")
		}
		if i > 0 {
			out.WriteByte(',')
		}
		name, _ := json.Marshal(key)
		out.Write(name)
		out.WriteByte(':')
		out.Write(value)
	}
	out.WriteByte('}')
	return out.Bytes(), nil
}
