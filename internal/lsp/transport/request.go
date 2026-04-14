package transport

import (
	"context"
	"errors"
	"fmt"
)

// ErrRequestTimeout is returned by CallWithTimeout when the context deadline is
// exceeded before the language server responds (REQ-LC-041).
//
// Callers may use errors.Is to distinguish timeouts from protocol errors.
//
// @MX:ANCHOR: [AUTO] ErrRequestTimeout — sentinel used by core.Client, Quality Gates, and test assertions
// @MX:REASON: fan_in >= 3 — core.Client degraded-state handler, test assertions, and upstream callers all branch on this sentinel
var ErrRequestTimeout = errors.New("lsp: request timeout")

// WrapCallError wraps err with contextual information: method name, file URI,
// and language server identifier (REQ-LC-040).
//
// Returns nil when err is nil, so callers can use:
//
//	return WrapCallError(method, uri, lang, transport.Call(...))
//
// @MX:NOTE: [AUTO] 에러 래핑 규칙: fmt.Errorf("%w" 패턴 사용 — errors.Is/As 체인 유지
func WrapCallError(method, uri, language string, err error) error {
	if err == nil {
		return nil
	}
	if uri != "" {
		return fmt.Errorf("lsp call %s (uri=%s lang=%s): %w", method, uri, language, err)
	}
	return fmt.Errorf("lsp call %s (lang=%s): %w", method, language, err)
}

// CallWithTimeout sends a JSON-RPC request via t and enforces the caller's
// context deadline (REQ-LC-040, REQ-LC-041).
//
// If ctx is nil, CallWithTimeout returns an error immediately.
//
// On context cancellation / deadline exceeded, the error is wrapped with
// ErrRequestTimeout so callers can branch:
//
//	if errors.Is(err, transport.ErrRequestTimeout) { ... }
//
// Protocol errors from the underlying transport are wrapped with WrapCallError,
// embedding method and language for observability.
func CallWithTimeout(ctx context.Context, t Transport, method string, params, result any, language string) error {
	if ctx == nil {
		return fmt.Errorf("lsp: CallWithTimeout %q: nil context", method)
	}

	err := t.Call(ctx, method, params, result)
	if err == nil {
		return nil
	}

	// 컨텍스트 타임아웃/취소 여부 확인 (REQ-LC-041)
	if isContextErr(ctx, err) {
		return fmt.Errorf("lsp call %s (lang=%s): %w", method, language, ErrRequestTimeout)
	}

	// 일반 프로토콜 에러: method + language 컨텍스트 추가
	return WrapCallError(method, "", language, err)
}

// isContextErr reports whether the error is due to context cancellation or
// deadline exceeded, either from the context itself or from a wrapped error.
func isContextErr(ctx context.Context, err error) bool {
	if ctx.Err() != nil {
		return true
	}
	return errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled)
}
