// Package gopls provides a self-implemented LSP client for communicating with the gopls subprocess.
// Implements JSON-RPC 2.0 Content-Length framing using only the standard library without external dependencies.
// REQ-GB-060..062: only encoding/json, bufio, os/exec, context, sync, log/slog are used.
package gopls

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// Writer writes JSON-RPC messages with LSP Content-Length framing applied.
// REQ-GB-030, REQ-GB-032: serializes in Content-Length header format `Content-Length: N\r\n\r\n<json>`.
type Writer struct {
	w io.Writer
}

// NewWriter creates a Writer wrapping the given io.Writer.
func NewWriter(w io.Writer) *Writer {
	return &Writer{w: w}
}

// Write serializes msg as JSON and writes it with a Content-Length header.
// REQ-GB-032: computes Content-Length and writes the entire frame atomically.
func (w *Writer) Write(msg any) error {
	body, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("gopls: write: JSON serialization failed: %w", err)
	}

	var buf bytes.Buffer
	// LSP Content-Length header format: `Content-Length: N\r\n\r\n`
	fmt.Fprintf(&buf, "Content-Length: %d\r\n\r\n", len(body))
	buf.Write(body)

	_, err = w.w.Write(buf.Bytes())
	if err != nil {
		return fmt.Errorf("gopls: write: stream write failed: %w", err)
	}
	return nil
}

// Reader reads JSON-RPC messages from a stream encoded with LSP Content-Length framing.
// REQ-GB-031: parses the Content-Length header and reads exactly N bytes.
type Reader struct {
	r *bufio.Reader
}

// NewReader creates a Reader wrapping the given io.Reader.
// Internally uses bufio.Reader to improve header parsing performance.
func NewReader(r io.Reader) *Reader {
	return &Reader{r: bufio.NewReader(r)}
}

// Read reads the next LSP message and returns raw JSON bytes.
// Returns io.EOF when the stream ends.
func (r *Reader) Read() (json.RawMessage, error) {
	length, err := parseHeaders(r.r)
	if err != nil {
		return nil, err
	}

	// Read exactly length bytes.
	body := make([]byte, length)
	if _, err := io.ReadFull(r.r, body); err != nil {
		return nil, fmt.Errorf("gopls: read: body read failed (declared length %d): %w", length, err)
	}
	return json.RawMessage(body), nil
}

// parseHeaders parses the LSP header section (terminated by `\r\n\r\n`) and returns the Content-Length value.
// REQ-GB-031: reads lines until it finds the Content-Length header, stopping at the double CRLF.
//
// Accepts io.Reader so it can be called from tests. Converts internally to bufio.Reader.
//
// Supported format:
//
//	Content-Length: N\r\n
//	Content-Type: ...\r\n  (optional, ignored)
//	\r\n
func parseHeaders(r io.Reader) (int, error) {
	// Reader.Read() passes a bufio.Reader, but tests may pass a strings.Reader.
	// bufio.NewReader does not return the existing *bufio.Reader as-is, so use a type switch.
	var br *bufio.Reader
	if b, ok := r.(*bufio.Reader); ok {
		br = b
	} else {
		br = bufio.NewReader(r)
	}

	contentLength := -1

	for {
		line, err := br.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				// EOF reached before finding end of headers
				if line == "" {
					return 0, io.EOF
				}
			} else {
				return 0, fmt.Errorf("gopls: header read failed: %w", err)
			}
		}

		// Normalize CRLF: handle \r\n or \n
		line = strings.TrimRight(line, "\r\n")

		// Empty line signals end of headers
		if line == "" {
			if contentLength == -1 {
				return 0, fmt.Errorf("gopls: no Content-Length in headers")
			}
			return contentLength, nil
		}

		// Parse only the Content-Length header; ignore other headers (Content-Type, etc.).
		if strings.HasPrefix(line, "Content-Length:") {
			val := strings.TrimSpace(strings.TrimPrefix(line, "Content-Length:"))
			n, err := strconv.Atoi(val)
			if err != nil {
				return 0, fmt.Errorf("gopls: Content-Length parse failed %q: %w", val, err)
			}
			if n < 0 {
				return 0, fmt.Errorf("gopls: Content-Length negative value: %d", n)
			}
			contentLength = n
		}
	}
}
