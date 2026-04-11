package gopls

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"
	"testing"
)

// TestWriter_Write_RoundTrip는 Writer가 Content-Length 헤더와 함께 JSON을 올바르게
// 직렬화하고 Reader가 동일한 데이터를 복원하는지 라운드트립 검증한다.
func TestWriter_Write_RoundTrip(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		msg     any
		wantErr bool
	}{
		{
			name:    "단순 객체",
			msg:     map[string]string{"method": "initialize", "jsonrpc": "2.0"},
			wantErr: false,
		},
		{
			name:    "중첩 구조체",
			msg:     map[string]any{"id": 1, "params": map[string]string{"rootUri": "file:///tmp"}},
			wantErr: false,
		},
		{
			name:    "빈 객체",
			msg:     map[string]any{},
			wantErr: false,
		},
		{
			name:    "유니코드 포함",
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
				t.Fatalf("Write() 에러 = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}

			// 작성된 데이터를 Reader로 다시 읽어서 원본과 비교한다.
			r := NewReader(&buf)
			raw, err := r.Read()
			if err != nil {
				t.Fatalf("Read() 에러 = %v", err)
			}

			// JSON 역직렬화 후 원본과 비교한다.
			var got any
			if err := json.Unmarshal(raw, &got); err != nil {
				t.Fatalf("Unmarshal 에러 = %v", err)
			}

			// 원본도 JSON 직렬화/역직렬화 후 비교하여 타입 정규화를 맞춘다.
			orig, _ := json.Marshal(tt.msg)
			var want any
			json.Unmarshal(orig, &want)

			gotJSON, _ := json.Marshal(got)
			wantJSON, _ := json.Marshal(want)
			if string(gotJSON) != string(wantJSON) {
				t.Errorf("라운드트립 불일치:\n got = %s\nwant = %s", gotJSON, wantJSON)
			}
		})
	}
}

// TestReader_Read_MultipleMessages는 버퍼에 여러 메시지가 연속으로 있을 때
// 각 메시지를 순서대로 읽는지 검증한다.
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
			t.Fatalf("Write() 에러 = %v", err)
		}
	}

	r := NewReader(&buf)
	for i, want := range msgs {
		raw, err := r.Read()
		if err != nil {
			t.Fatalf("메시지 %d Read() 에러 = %v", i, err)
		}
		var got map[string]any
		if err := json.Unmarshal(raw, &got); err != nil {
			t.Fatalf("메시지 %d Unmarshal 에러 = %v", i, err)
		}
		wantJSON, _ := json.Marshal(want)
		gotJSON, _ := json.Marshal(got)
		if string(gotJSON) != string(wantJSON) {
			t.Errorf("메시지 %d 불일치:\n got = %s\nwant = %s", i, gotJSON, wantJSON)
		}
	}
}

// TestReader_Read_MalformedHeader는 잘못된 Content-Length 헤더를 파싱할 때
// 에러를 반환하는지 검증한다.
func TestReader_Read_MalformedHeader(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "Content-Length 누락",
			input:   "Content-Type: application/json\r\n\r\n{}",
			wantErr: true,
		},
		{
			name:    "Content-Length 값이 숫자가 아님",
			input:   "Content-Length: abc\r\n\r\n{}",
			wantErr: true,
		},
		{
			name:    "Content-Length가 음수",
			input:   "Content-Length: -1\r\n\r\n{}",
			wantErr: true,
		},
		{
			name:    "헤더 구분자 없음",
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
				t.Errorf("Read() 에러 = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestReader_Read_TruncatedStream는 선언된 Content-Length보다 실제 바디가 짧을 때
// 에러를 반환하는지 검증한다.
func TestReader_Read_TruncatedStream(t *testing.T) {
	t.Parallel()

	// Content-Length: 100이지만 실제로는 2바이트만 제공한다.
	input := "Content-Length: 100\r\n\r\n{}"
	r := NewReader(strings.NewReader(input))
	_, err := r.Read()
	if err == nil {
		t.Fatal("짧은 스트림에서 에러를 기대했지만 nil 반환")
	}
}

// TestReader_Read_EOF는 빈 스트림에서 Read를 호출할 때 io.EOF를 반환하는지 검증한다.
func TestReader_Read_EOF(t *testing.T) {
	t.Parallel()

	r := NewReader(strings.NewReader(""))
	_, err := r.Read()
	if err != io.EOF && err != io.ErrUnexpectedEOF {
		t.Errorf("빈 스트림에서 io.EOF 또는 io.ErrUnexpectedEOF를 기대했지만: %v", err)
	}
}

// TestParseHeaders_Valid는 올바른 헤더 시퀀스를 파싱하여 Content-Length를 반환하는지 검증한다.
func TestParseHeaders_Valid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		input      string
		wantLength int
		wantErr    bool
	}{
		{
			name:       "단일 헤더",
			input:      "Content-Length: 42\r\n\r\n",
			wantLength: 42,
			wantErr:    false,
		},
		{
			name:       "추가 헤더 무시",
			input:      "Content-Length: 7\r\nContent-Type: application/vscode-jsonrpc\r\n\r\n",
			wantLength: 7,
			wantErr:    false,
		},
		{
			name:       "공백 있는 Content-Length",
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
				t.Fatalf("parseHeaders() 에러 = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && n != tt.wantLength {
				t.Errorf("parseHeaders() = %d, want %d", n, tt.wantLength)
			}
		})
	}
}
