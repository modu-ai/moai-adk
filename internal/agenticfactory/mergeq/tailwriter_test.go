package mergeq

// Defect answer: the writer that captures a gate's output must keep the end of it, never fail on volume, and never
// hold more than its limit. Each case writes a known sequence and compares what was kept with the exact tail.

import (
	"bytes"
	"testing"
)

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
