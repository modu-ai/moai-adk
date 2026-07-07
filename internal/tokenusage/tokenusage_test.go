package tokenusage

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// writeTranscript writes the given JSONL lines to a fresh file under dir and
// returns its absolute path. Test-only helper; all fixtures live under
// t.TempDir() so no real ~/.claude transcript is ever touched (AC-TA-005).
func writeTranscript(t *testing.T, dir, name string, lines []string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	var buf []byte
	for _, l := range lines {
		buf = append(buf, l...)
		buf = append(buf, '\n')
	}
	if err := os.WriteFile(path, buf, 0o644); err != nil {
		t.Fatalf("writeTranscript: %v", err)
	}
	return path
}

// assistantLine builds a well-formed assistant transcript record carrying a
// message.usage object with the four accounted fields plus representative
// extra fields (service_tier, cache_creation) that the parser must ignore.
func assistantLine(in, out, cacheCreate, cacheRead int) string {
	return `{"type":"assistant","message":{"usage":{` +
		`"input_tokens":` + itoa(in) + `,` +
		`"output_tokens":` + itoa(out) + `,` +
		`"cache_creation_input_tokens":` + itoa(cacheCreate) + `,` +
		`"cache_read_input_tokens":` + itoa(cacheRead) + `,` +
		`"service_tier":"standard",` +
		`"cache_creation":{"ephemeral_5m_input_tokens":0}}}}`
}

// itoa is a tiny stdlib-free int formatter to keep fixture builders readable.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var b [20]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		b[i] = '-'
	}
	return string(b[i:])
}

// TestSumSession is the core happy-path: three identical assistant turns must
// sum arithmetically across all four fields (AC-TA-001, REQ-TA-001/002).
func TestSumSession(t *testing.T) {
	dir := t.TempDir()
	lines := []string{
		assistantLine(100, 20, 0, 500),
		assistantLine(100, 20, 0, 500),
		assistantLine(100, 20, 0, 500),
	}
	path := writeTranscript(t, dir, "session.jsonl", lines)

	u, err := SumSession(path)
	if err != nil {
		t.Fatalf("SumSession returned error: %v", err)
	}
	if u.TokensInput != 300 {
		t.Errorf("TokensInput = %d, want 300", u.TokensInput)
	}
	if u.TokensOutput != 60 {
		t.Errorf("TokensOutput = %d, want 60", u.TokensOutput)
	}
	if u.TokensCacheCreation != 0 {
		t.Errorf("TokensCacheCreation = %d, want 0", u.TokensCacheCreation)
	}
	if u.TokensCacheRead != 1500 {
		t.Errorf("TokensCacheRead = %d, want 1500", u.TokensCacheRead)
	}
	// tokens_spent = arithmetic sum of the four fields (REQ-TA-002).
	if u.TokensSpent != 1860 {
		t.Errorf("TokensSpent = %d, want 1860 (300+60+0+1500)", u.TokensSpent)
	}
}

// TestMalformedTolerant mixes a usage-absent assistant record, a JSON-unparseable
// line, a non-assistant record carrying a usage block, blank lines, and one valid
// record. Only the valid record contributes, and the parser never panics
// (AC-TA-002, AC-TA-003, REQ-TA-013).
func TestMalformedTolerant(t *testing.T) {
	dir := t.TempDir()
	lines := []string{
		`{"type":"assistant","message":{}}`, // (a) usage absent -> 0 contribution
		`{this is not valid json`,           // (b) malformed -> skip
		``,                                  // blank -> skip
		`{"type":"user","message":{"usage":{"input_tokens":9999}}}`, // non-assistant -> ignored
		assistantLine(100, 20, 0, 500),                              // (c) valid -> counted
	}
	path := writeTranscript(t, dir, "mixed.jsonl", lines)

	u, err := SumSession(path)
	if err != nil {
		t.Fatalf("SumSession returned error on tolerant input: %v", err)
	}
	if u.TokensInput != 100 {
		t.Errorf("TokensInput = %d, want 100 (only valid assistant record counts)", u.TokensInput)
	}
	if u.TokensOutput != 20 {
		t.Errorf("TokensOutput = %d, want 20", u.TokensOutput)
	}
	if u.TokensCacheRead != 500 {
		t.Errorf("TokensCacheRead = %d, want 500", u.TokensCacheRead)
	}
	if u.TokensSpent != 620 {
		t.Errorf("TokensSpent = %d, want 620", u.TokensSpent)
	}
}

// TestEmptyTranscript: a transcript with zero assistant turns yields the zero
// Usage and a zero ratio, with no panic (edge case §D.2).
func TestEmptyTranscript(t *testing.T) {
	dir := t.TempDir()
	path := writeTranscript(t, dir, "empty.jsonl", nil)

	u, err := SumSession(path)
	if err != nil {
		t.Fatalf("SumSession on empty transcript: %v", err)
	}
	if u.TokensSpent != 0 {
		t.Errorf("TokensSpent = %d, want 0", u.TokensSpent)
	}
	if u.CacheHitRatio != 0 {
		t.Errorf("CacheHitRatio = %v, want 0", u.CacheHitRatio)
	}
}

// TestAbsentFile: a missing transcript file returns a not-exist error and the
// zero Usage, without panicking. The caller (attribution layer, M2) is expected
// to skip-and-continue on this error (REQ-TA-013 file-absent branch).
func TestAbsentFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "does-not-exist.jsonl")

	u, err := SumSession(path)
	if err == nil {
		t.Fatalf("SumSession on absent file: want error, got nil")
	}
	if !errors.Is(err, os.ErrNotExist) {
		t.Errorf("SumSession absent-file error = %v, want wrapping os.ErrNotExist", err)
	}
	if u != (Usage{}) {
		t.Errorf("Usage on absent file = %+v, want zero value", u)
	}
}

// TestCacheHitRatioHelper exercises the ratio computation boundaries directly,
// including the 0-denominator guard and the read-only-cache extreme
// (AC-TA-004, REQ-TA-003).
func TestCacheHitRatioHelper(t *testing.T) {
	tests := []struct {
		name                          string
		input, cacheCreate, cacheRead int
		want                          float64
	}{
		{"zero denominator -> 0", 0, 0, 0, 0},
		{"pure cache read -> 1", 0, 0, 500, 1},
		{"typical 500/600", 100, 0, 500, 500.0 / 600.0},
		{"with creation in denom", 100, 100, 800, 800.0 / 1000.0},
		{"no cache read -> 0", 100, 0, 0, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CacheHitRatio(tt.input, tt.cacheCreate, tt.cacheRead)
			if got != tt.want {
				t.Errorf("CacheHitRatio(%d,%d,%d) = %v, want %v",
					tt.input, tt.cacheCreate, tt.cacheRead, got, tt.want)
			}
			if got < 0 || got > 1 {
				t.Errorf("CacheHitRatio out of [0,1]: %v", got)
			}
		})
	}
}

// TestCacheHitRatioFromSession confirms SumSession finalizes the ratio from the
// summed fields: read=500 over denom input(100)+creation(0)+read(500)=600
// (AC-TA-004 via the summation path).
func TestCacheHitRatioFromSession(t *testing.T) {
	dir := t.TempDir()
	path := writeTranscript(t, dir, "ratio.jsonl", []string{assistantLine(100, 20, 0, 500)})

	u, err := SumSession(path)
	if err != nil {
		t.Fatalf("SumSession: %v", err)
	}
	want := 500.0 / 600.0
	if u.CacheHitRatio != want {
		t.Errorf("CacheHitRatio = %v, want %v", u.CacheHitRatio, want)
	}
}

// TestReadOnlyInvariant verifies SumSession never mutates the transcript
// directory: file count, sizes, and modtimes are identical before and after
// (AC-TA-005, REQ-TA-004). This mirrors the ~/.claude/projects/** read-only
// contract using an isolated t.TempDir() so no real transcript is touched.
func TestReadOnlyInvariant(t *testing.T) {
	dir := t.TempDir()
	writeTranscript(t, dir, "a.jsonl", []string{assistantLine(100, 20, 0, 500)})
	writeTranscript(t, dir, "b.jsonl", []string{assistantLine(50, 10, 5, 200)})

	type snap struct {
		size    int64
		modTime int64
	}
	capture := func() map[string]snap {
		entries, err := os.ReadDir(dir)
		if err != nil {
			t.Fatalf("ReadDir: %v", err)
		}
		m := make(map[string]snap, len(entries))
		for _, e := range entries {
			info, err := e.Info()
			if err != nil {
				t.Fatalf("Info: %v", err)
			}
			m[e.Name()] = snap{size: info.Size(), modTime: info.ModTime().UnixNano()}
		}
		return m
	}

	before := capture()
	// Read every transcript in the directory.
	for name := range before {
		if _, err := SumSession(filepath.Join(dir, name)); err != nil {
			t.Fatalf("SumSession(%s): %v", name, err)
		}
	}
	after := capture()

	if len(before) != len(after) {
		t.Fatalf("file count changed: before=%d after=%d", len(before), len(after))
	}
	for name, b := range before {
		a, ok := after[name]
		if !ok {
			t.Errorf("file %s disappeared", name)
			continue
		}
		if a.size != b.size {
			t.Errorf("file %s size changed: before=%d after=%d", name, b.size, a.size)
		}
		if a.modTime != b.modTime {
			t.Errorf("file %s modtime changed (write side-effect): before=%d after=%d", name, b.modTime, a.modTime)
		}
	}
}

// TestLongLineTolerance ensures a very long JSON line (exceeding bufio.Scanner's
// default 64KB token cap) is parsed rather than silently dropped — transcript
// assistant records routinely exceed 64KB. Guards against a Scanner-based
// regression.
func TestLongLineTolerance(t *testing.T) {
	dir := t.TempDir()
	// Build a valid assistant record padded with a large ignored string field.
	pad := make([]byte, 200*1024) // 200KB
	for i := range pad {
		pad[i] = 'x'
	}
	big := `{"type":"assistant","message":{"usage":{"input_tokens":100,"output_tokens":20,` +
		`"cache_creation_input_tokens":0,"cache_read_input_tokens":500}},"pad":"` + string(pad) + `"}`
	path := writeTranscript(t, dir, "big.jsonl", []string{big})

	u, err := SumSession(path)
	if err != nil {
		t.Fatalf("SumSession on long line: %v", err)
	}
	if u.TokensSpent != 620 {
		t.Errorf("TokensSpent = %d, want 620 (long line must not be dropped)", u.TokensSpent)
	}
}
