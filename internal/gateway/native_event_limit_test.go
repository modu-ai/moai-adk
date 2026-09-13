package gateway

import (
	"context"
	"github.com/modu-ai/moai-adk/internal/gateway/translate"
	"io"
	"strings"
	"testing"
	"testing/iotest"
)

func TestNativeEventRawByteBoundaries(t *testing.T) {
	for _, crlf := range []bool{false, true} {
		for _, fragmented := range []bool{false, true} {
			raw := nativeSSE()
			sep := "\n\n"
			if crlf {
				raw = strings.ReplaceAll(raw, "\n", "\r\n")
				sep = "\r\n\r\n"
			}
			longest := 0
			for _, frame := range strings.Split(raw, sep) {
				if len(frame) > 0 && len(frame)+len(sep) > longest {
					longest = len(frame) + len(sep)
				}
			}
			for _, offset := range []int{-1, 0, 1} {
				var reader io.Reader = strings.NewReader(raw)
				if fragmented {
					reader = iotest.OneByteReader(reader)
				}
				b := newNativeBody(context.Background(), io.NopCloser(reader), "canonical", translate.Limits{MaxOutputBytes: 1 << 20, MaxEventBytes: longest + offset})
				out, err := io.ReadAll(b)
				_ = b.Close() // nativeBody.Close always returns nil
				success := strings.Contains(string(out), "event: message_stop")
				t.Logf("CRLF=%v fragmented=%v largest_raw_event=%d event_limit=%d success=%v err=%v", crlf, fragmented, longest, longest+offset, success, err)
				if offset < 0 && (err == nil || success) {
					t.Error("raw event over limit accepted")
				}
				if offset >= 0 && (err != nil || !success) {
					t.Error("within event limit rejected")
				}
			}
		}
	}
}
