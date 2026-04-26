package gopls

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"
	"testing"
)

// TestWriter_Write_RoundTrip verifies that Writer correctly serializes JSON with
// the Content-Length header and that Reader recovers the same data — a
// round-trip check.
func TestWriter_Write_RoundTrip(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		msg     any
		wantErr bool
	}{
		{
			name:    "simple object",
			msg:     map[string]string{"method": "initialize", "jsonrpc": "2.0"},
			wantErr: false,
		},
		{
			name:    "nested struct",
			msg:     map[string]any{"id": 1, "params": map[string]string{"rootUri": "file:///tmp"}},
			wantErr: false,
		},
		{
			name:    "empty object",
			msg:     map[string]any{},
			wantErr: false,
		},
		{
			name:    "with unicode",
			msg:     map[string]string{"message": "오류: 파일을 찾을 수 없음 — héllo wörld"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer
			w := NewWriter(&buf)

			err := w.Write(tt.msg)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Write() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}

			// Read the written data back through Reader and compare with the original.
			r := NewReader(&buf)
			raw, err := r.Read()
			if err != nil {
				t.Fatalf("Read() error = %v", err)
			}

			// Unmarshal JSON and compare with the original.
			var got any
			if err := json.Unmarshal(raw, &got); err != nil {
				t.Fatalf("Unmarshal error = %v", err)
			}

			// Marshal/unmarshal the original too to normalize types for comparison.
			orig, _ := json.Marshal(tt.msg)
			var want any
			_ = json.Unmarshal(orig, &want)

			gotJSON, _ := json.Marshal(got)
			wantJSON, _ := json.Marshal(want)
			if string(gotJSON) != string(wantJSON) {
				t.Errorf("round-trip mismatch:\n got = %s\nwant = %s", gotJSON, wantJSON)
			}
		})
	}
}

// TestReader_Read_MultipleMessages verifies that when multiple messages are
// concatenated in a buffer, each one is read in order.
func TestReader_Read_MultipleMessages(t *testing.T) {
	t.Parallel()

	msgs := []map[string]any{
		{"id": 1, "method": "initialize"},
		{"id": 2, "method": "textDocument/didOpen"},
		{"id": 3, "result": map[string]any{"capabilities": map[string]any{}}},
	}

	var buf bytes.Buffer
	w := NewWriter(&buf)
	for _, m := range msgs {
		if err := w.Write(m); err != nil {
			t.Fatalf("Write() error = %v", err)
		}
	}

	r := NewReader(&buf)
	for i, want := range msgs {
		raw, err := r.Read()
		if err != nil {
			t.Fatalf("message %d Read() error = %v", i, err)
		}
		var got map[string]any
		if err := json.Unmarshal(raw, &got); err != nil {
			t.Fatalf("message %d Unmarshal error = %v", i, err)
		}
		wantJSON, _ := json.Marshal(want)
		gotJSON, _ := json.Marshal(got)
		if string(gotJSON) != string(wantJSON) {
			t.Errorf("message %d mismatch:\n got = %s\nwant = %s", i, gotJSON, wantJSON)
		}
	}
}

// TestReader_Read_MalformedHeader verifies that parsing a malformed
// Content-Length header returns an error.
func TestReader_Read_MalformedHeader(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "missing Content-Length",
			input:   "Content-Type: application/json\r\n\r\n{}",
			wantErr: true,
		},
		{
			name:    "non-numeric Content-Length",
			input:   "Content-Length: abc\r\n\r\n{}",
			wantErr: true,
		},
		{
			name:    "negative Content-Length",
			input:   "Content-Length: -1\r\n\r\n{}",
			wantErr: true,
		},
		{
			name:    "missing header delimiter",
			input:   "Content-Length: 2{}",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			r := NewReader(strings.NewReader(tt.input))
			_, err := r.Read()
			if (err != nil) != tt.wantErr {
				t.Errorf("Read() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestReader_Read_TruncatedStream verifies that when the actual body is shorter
// than the declared Content-Length, an error is returned.
func TestReader_Read_TruncatedStream(t *testing.T) {
	t.Parallel()

	// Content-Length: 100 but only 2 bytes are actually provided.
	input := "Content-Length: 100\r\n\r\n{}"
	r := NewReader(strings.NewReader(input))
	_, err := r.Read()
	if err == nil {
		t.Fatal("expected error on truncated stream, got nil")
	}
}

// TestReader_Read_EOF verifies that calling Read on an empty stream returns io.EOF.
func TestReader_Read_EOF(t *testing.T) {
	t.Parallel()

	r := NewReader(strings.NewReader(""))
	_, err := r.Read()
	if err != io.EOF && err != io.ErrUnexpectedEOF {
		t.Errorf("expected io.EOF or io.ErrUnexpectedEOF for empty stream, got: %v", err)
	}
}

// TestParseHeaders_Valid verifies that a valid header sequence is parsed and
// the Content-Length is returned.
func TestParseHeaders_Valid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		input      string
		wantLength int
		wantErr    bool
	}{
		{
			name:       "single header",
			input:      "Content-Length: 42\r\n\r\n",
			wantLength: 42,
			wantErr:    false,
		},
		{
			name:       "ignore extra header",
			input:      "Content-Length: 7\r\nContent-Type: application/vscode-jsonrpc\r\n\r\n",
			wantLength: 7,
			wantErr:    false,
		},
		{
			name:       "Content-Length with spaces",
			input:      "Content-Length:  15\r\n\r\n",
			wantLength: 15,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			n, err := parseHeaders(strings.NewReader(tt.input))
			if (err != nil) != tt.wantErr {
				t.Fatalf("parseHeaders() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && n != tt.wantLength {
				t.Errorf("parseHeaders() = %d, want %d", n, tt.wantLength)
			}
		})
	}
}
