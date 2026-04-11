// Package gopls는 gopls 서브프로세스와 통신하는 자체 구현 LSP 클라이언트를 제공한다.
// 외부 라이브러리 없이 표준 라이브러리만으로 JSON-RPC 2.0 Content-Length 프레이밍을 구현한다.
// REQ-GB-060..062: encoding/json, bufio, os/exec, context, sync, log/slog만 사용한다.
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

// Writer는 LSP Content-Length 프레이밍을 적용하여 JSON-RPC 메시지를 기록한다.
// REQ-GB-030, REQ-GB-032: Content-Length 헤더 형식 `Content-Length: N\r\n\r\n<json>`으로 직렬화한다.
type Writer struct {
	w io.Writer
}

// NewWriter는 주어진 io.Writer를 감싸는 Writer를 생성한다.
func NewWriter(w io.Writer) *Writer {
	return &Writer{w: w}
}

// Write는 msg를 JSON으로 직렬화하고 Content-Length 헤더와 함께 기록한다.
// REQ-GB-032: Content-Length를 계산하고 프레임 전체를 원자적으로 기록한다.
func (w *Writer) Write(msg any) error {
	body, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("gopls: write: JSON 직렬화 실패: %w", err)
	}

	var buf bytes.Buffer
	// LSP Content-Length 헤더 형식: `Content-Length: N\r\n\r\n`
	fmt.Fprintf(&buf, "Content-Length: %d\r\n\r\n", len(body))
	buf.Write(body)

	_, err = w.w.Write(buf.Bytes())
	if err != nil {
		return fmt.Errorf("gopls: write: 스트림 기록 실패: %w", err)
	}
	return nil
}

// Reader는 LSP Content-Length 프레이밍으로 인코딩된 스트림에서 JSON-RPC 메시지를 읽는다.
// REQ-GB-031: Content-Length 헤더를 파싱하고 정확히 N바이트를 읽는다.
type Reader struct {
	r *bufio.Reader
}

// NewReader는 주어진 io.Reader를 감싸는 Reader를 생성한다.
// 내부적으로 bufio.Reader를 사용하여 헤더 파싱 성능을 높인다.
func NewReader(r io.Reader) *Reader {
	return &Reader{r: bufio.NewReader(r)}
}

// Read는 다음 LSP 메시지를 읽어 JSON 원시 바이트를 반환한다.
// 스트림이 끝나면 io.EOF를 반환한다.
func (r *Reader) Read() (json.RawMessage, error) {
	length, err := parseHeaders(r.r)
	if err != nil {
		return nil, err
	}

	// 정확히 length 바이트를 읽는다.
	body := make([]byte, length)
	if _, err := io.ReadFull(r.r, body); err != nil {
		return nil, fmt.Errorf("gopls: read: 바디 읽기 실패 (선언 길이 %d): %w", length, err)
	}
	return json.RawMessage(body), nil
}

// parseHeaders는 LSP 헤더 섹션(`\r\n\r\n` 종료)을 파싱하여 Content-Length 값을 반환한다.
// REQ-GB-031: Content-Length 헤더를 찾을 때까지 줄을 읽고, 이중 CRLF에서 멈춘다.
//
// 테스트에서도 호출 가능하도록 io.Reader를 받는다. 내부적으로 bufio.Reader로 변환한다.
//
// 지원 형식:
//
//	Content-Length: N\r\n
//	Content-Type: ...\r\n  (선택적, 무시)
//	\r\n
func parseHeaders(r io.Reader) (int, error) {
	// Reader.Read()는 이미 bufio.Reader를 전달하지만, 테스트에서는 strings.Reader를 전달한다.
	// bufio.NewReader는 이미 *bufio.Reader인 경우 그대로 반환하지 않으므로 타입 스위치를 사용한다.
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
				// EOF이면서 아직 헤더 종료를 못 만났을 때
				if line == "" {
					return 0, io.EOF
				}
			} else {
				return 0, fmt.Errorf("gopls: 헤더 읽기 실패: %w", err)
			}
		}

		// CRLF 정규화: \r\n 또는 \n 처리
		line = strings.TrimRight(line, "\r\n")

		// 빈 줄 = 헤더 종료 신호
		if line == "" {
			if contentLength == -1 {
				return 0, fmt.Errorf("gopls: 헤더에 Content-Length 없음")
			}
			return contentLength, nil
		}

		// Content-Length 헤더만 파싱한다. 다른 헤더(Content-Type 등)는 무시한다.
		if strings.HasPrefix(line, "Content-Length:") {
			val := strings.TrimSpace(strings.TrimPrefix(line, "Content-Length:"))
			n, err := strconv.Atoi(val)
			if err != nil {
				return 0, fmt.Errorf("gopls: Content-Length 파싱 실패 %q: %w", val, err)
			}
			if n < 0 {
				return 0, fmt.Errorf("gopls: Content-Length 음수 값: %d", n)
			}
			contentLength = n
		}
	}
}
