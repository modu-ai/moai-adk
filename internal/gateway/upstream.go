package gateway

import (
	"context"
	"fmt"
	"net/http"
)

// upstreamAttempts bounds egress per routed request: one initial attempt plus
// two retries. The bound keeps a wedged upstream from turning one client turn
// into an unbounded wait.
const upstreamAttempts = 3

// upstreamSend runs attempt until it yields a terminal egress outcome. A
// transport error or an upstream 5xx is retryable: each attempt opens a fresh
// connection (the adapters run DisableKeepAlives), so a wedged or stale
// upstream socket is never re-dialed. Every attempt happens strictly before
// any response byte is exposed to the gateway client, which keeps the retried
// request pre-stream idempotent; a started stream is never re-driven here.
//
// The request body is re-created from req.GetBody before every retry: a
// *bytes.Reader body is consumed by the first send, and re-driving the same
// request without a reset would fail with a body-length transport error.
//
// The returned count is the number of egress attempts made. On retryable
// exhaustion the final outcome is returned as-is — for a 5xx, the last
// response's headers still carry Retry-After, and its body is already closed
// — so the caller's existing status mapping renders the error.
func upstreamSend(ctx context.Context, req *http.Request, attempt func() (*http.Response, error)) (*http.Response, int, error) {
	var resp *http.Response
	var err error
	attempts := 0
	for {
		if e := ctx.Err(); e != nil {
			return nil, attempts, e
		}
		attempts++
		if attempts > 1 && req != nil && req.GetBody != nil {
			b, e := req.GetBody()
			if e != nil {
				return nil, attempts, e
			}
			req.Body = b
		}
		resp, err = attempt()
		if err == nil && (resp == nil || resp.Body == nil || resp.StatusCode < 500 || resp.StatusCode > 599) {
			return resp, attempts, nil
		}
		if resp != nil && resp.Body != nil {
			_ = resp.Body.Close()
		}
		if attempts >= upstreamAttempts {
			return resp, attempts, err
		}
	}
}

// gatewayConnectMessage labels the 502 the gateway itself renders when no
// upstream connection could be established, carrying the attempt count and
// the underlying reason, so a connection-failure storm reads differently from
// an upstream-declared failure in client debug logs.
func gatewayConnectMessage(attempts int, err error) string {
	return fmt.Sprintf("gateway upstream connection failed after %d attempts: %v", attempts, err)
}

// upstreamStatusMessage labels a 5xx the upstream itself returned after the
// bounded pre-stream retries, keeping upstream-declared and gateway-rendered
// 502s distinguishable in the body.
func upstreamStatusMessage(status, attempts int) string {
	return fmt.Sprintf("upstream returned %d after %d attempts", status, attempts)
}
