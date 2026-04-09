package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"sync"
)

// Handler processes MCP method calls.
type Handler interface {
	HandleMethod(ctx context.Context, method string, params json.RawMessage) (any, error)
}

// Server is an MCP stdio server that communicates via newline-delimited JSON.
type Server struct {
	handler Handler
	mu      sync.Mutex
}

// NewServer creates a new MCP stdio server backed by the given handler.
func NewServer(handler Handler) *Server {
	return &Server{handler: handler}
}

// Serve reads JSON-RPC requests from reader and writes responses to writer.
// It processes requests one at a time; MCP over stdio is inherently sequential.
// Serve returns when the reader is exhausted or the context is cancelled.
// @MX:WARN: scanner.Scan()이 I/O에서 블로킹되면 ctx.Done()으로 중단 불가. 호출자가 reader를 닫아야 Serve가 반환됨.
// @MX:REASON: 블로킹 I/O와 context 취소의 상호작용 제한
func (s *Server) Serve(ctx context.Context, reader io.Reader, writer io.Writer) error {
	if s.handler == nil {
		return fmt.Errorf("mcp: nil handler")
	}
	scanner := bufio.NewScanner(reader)
	// Use a 1 MiB buffer to accommodate large tool responses.
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var req JSONRPCRequest
		if err := json.Unmarshal(line, &req); err != nil {
			slog.Warn("mcp: invalid JSON-RPC request", "error", err)
			continue
		}

		// Notifications have no ID — they do not expect a response.
		if req.ID == nil {
			s.handleNotification(ctx, req)
			continue
		}

		result, callErr := s.handler.HandleMethod(ctx, req.Method, req.Params)

		resp := JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
		}
		if callErr != nil {
			resp.Error = &JSONRPCError{
				Code:    -32603,
				Message: callErr.Error(),
			}
		} else {
			resp.Result = result
		}

		data, marshalErr := json.Marshal(resp)
		if marshalErr != nil {
			slog.Warn("mcp: failed to marshal response",
				"method", req.Method,
				"error", marshalErr,
			)
			continue
		}
		s.mu.Lock()
		_, _ = fmt.Fprintf(writer, "%s\n", data)
		s.mu.Unlock()
	}

	return scanner.Err()
}

// handleNotification handles a JSON-RPC notification (e.g. notifications/initialized).
func (s *Server) handleNotification(_ context.Context, req JSONRPCRequest) {
	slog.Debug("mcp: notification received", "method", req.Method)
}
