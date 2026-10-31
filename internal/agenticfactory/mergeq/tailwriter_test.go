package mergeq

// Defect answer: the writer that captures a gate's output must keep the end of it, never fail on volume, and never
// hold more than its limit. Each case writes a known sequence and compares what was kept with the exact tail.

import (
	"bytes"
	"strings"
	"testing"
)

// Truncated says whether the start of the output was dropped, so a caller knows the first kept line may be a
// fragment. Everything that fits is not truncated, whatever the number of writes.
func TestTailWriter_TruncatedSaysWhetherTheStartWasDropped(t *testing.T) {
	for _, tc := range []struct {
		name   string
		max    int
		writes []string
		want   bool
	}{
		{"nothing written", 4, nil, false},
		{"under the limit", 6, []string{"abc", "d"}, false},
		{"exactly the limit in one write", 4, []string{"abcd"}, false},
		{"exactly the limit in two writes", 4, []string{"ab", "cd"}, false},
		{"one write over the limit", 4, []string{"abcde"}, true},
		{"a second write pushes the first out", 4, []string{"abc", "de"}, true},
		{"a full-size write after earlier bytes", 4, []string{"a", "bcde"}, true},
		{"a full-size write onto an empty buffer", 4, []string{"", "abcd"}, false},
	} {
		w := NewTailWriter(tc.max)
		for _, s := range tc.writes {
			if _, err := w.Write([]byte(s)); err != nil {
				t.Fatalf("%s: Write(%q): %v", tc.name, s, err)
			}
		}
		if got := w.Truncated(); got != tc.want {
			t.Errorf("%s: Truncated() = %v, want %v (kept %q)", tc.name, got, tc.want, w.Bytes())
		}
	}
}

func TestTailWriter_KeepsTheLastBytesAndNeverFails(t *testing.T) {
	for _, tc := range []struct {
		name   string
		max    int
		writes []string
		want   string
	}{
		{"under the limit keeps everything", 10, []string{"abc", "def"}, "abcdef"},
		{"exactly the limit", 6, []string{"abc", "def"}, "abcdef"},
		{"a second write pushes the first out", 6, []string{"abcd", "efgh"}, "cdefgh"},
		{"one write bigger than the limit keeps its tail", 4, []string{"0123456789"}, "6789"},
		{"many small writes keep the last window", 5, []string{"a", "b", "c", "d", "e", "f", "g"}, "cdefg"},
		{"an empty write changes nothing", 4, []string{"abcd", ""}, "abcd"},
	} {
		w := NewTailWriter(tc.max)
		for _, s := range tc.writes {
			n, err := w.Write([]byte(s))
			if err != nil || n != len(s) {
				t.Fatalf("%s: Write(%q) = %d, %v; want %d, nil", tc.name, s, n, err, len(s))
			}
		}
		if got := string(w.Bytes()); got != tc.want {
			t.Errorf("%s: kept %q, want %q", tc.name, got, tc.want)
		}
		if len(w.Bytes()) > tc.max {
			t.Errorf("%s: holds %d bytes, over the limit %d", tc.name, len(w.Bytes()), tc.max)
		}
	}
}

func TestTailWriter_NineMegabytesStayWithinTheLimit(t *testing.T) {
	w := NewTailWriter(64 << 10)
	chunk := bytes.Repeat([]byte("x"), 1<<20)
	for range 9 {
		if _, err := w.Write(chunk); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := w.Write([]byte("LAST")); err != nil {
		t.Fatal(err)
	}
	if len(w.Bytes()) != 64<<10 || !bytes.HasSuffix(w.Bytes(), []byte("xLAST")) {
		t.Errorf("kept %d bytes ending %q", len(w.Bytes()), w.Bytes()[len(w.Bytes())-6:])
	}
}

// MaskedPrefix says how many leading kept bytes are a secret value whose marker the window no longer holds.
func TestTailWriter_MaskedPrefixCoversTheKeptPartOfAMarkedLine(t *testing.T) {
	for _, tc := range []struct {
		name   string
		max    int
		writes []string
		want   int
	}{
		{"nothing dropped, marker kept", 64, []string{"password=abc\n"}, 0},
		{"value past the window, no newline", 16, []string{"password=", strings.Repeat("Q", 40)}, 16},
		{"marker exactly the dropped part", 16, []string{"password=", strings.Repeat("Q", 16)}, 16},
		{"marker cut by one byte", 16, []string{"password=", strings.Repeat("Q", 8)}, 16},
		{"marker wholly kept at the front", 16, []string{"password=", strings.Repeat("Q", 7)}, 0},
		{"the marked line ends inside the window", 20, []string{"password=xx\nlast line ok\n"}, 7},
		{"marker split across writes", 16, []string{"pa", "ssword=", strings.Repeat("Q", 40)}, 16},
		{"marker split by a control character", 16, []string{"pass\a", "word=", strings.Repeat("Q", 40)}, 16},
		{"marker ends its line, value line is cut", 16, []string{"password=\n", strings.Repeat("Q", 40)}, 16},
		{"marker ends its line, value line starts the window", 16, []string{"password=\n", strings.Repeat("Q", 16)}, 16},
		{"a line with no marker", 16, []string{"plain ", strings.Repeat("Q", 40)}, 0},
		{"the line before the window was marked, this one is not", 16, []string{"password=x\n", "plain ", strings.Repeat("Q", 40)}, 0},
	} {
		w := NewTailWriter(tc.max)
		for _, s := range tc.writes {
			_, _ = w.Write([]byte(s))
		}
		if got := w.MaskedPrefix(); got != tc.want {
			t.Errorf("%s: MaskedPrefix() = %d, want %d (kept %q)", tc.name, got, tc.want, w.Bytes())
		}
	}
}

// The watch keeps bounded state however much is written: spans only for lines still in the window, and a carry of at
// most tailCarry bytes. A marker whose raw form is longer than that carry is not recognised (the stated ceiling).
func TestTailWriter_WatchStateIsBounded(t *testing.T) {
	w := NewTailWriter(64)
	for range 3000 {
		_, _ = w.Write([]byte("key=1\n"))
	}
	_, _ = w.Write([]byte(strings.Repeat("Z", 5000)))
	if len(w.spans) > 64 || len(w.carry) > tailCarry {
		t.Errorf("state grew: %d spans, %d carry bytes", len(w.spans), len(w.carry))
	}
	w = NewTailWriter(16)
	for _, s := range []string{"pass", strings.Repeat("\a", tailCarry+88), "word=" + strings.Repeat("Q", 40)} {
		_, _ = w.Write([]byte(s))
	}
	if w.MaskedPrefix() != 0 {
		t.Errorf("a marker spread over more than %d raw bytes is documented as unseen", tailCarry)
	}
}
