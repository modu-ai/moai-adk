package contract

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"slices"
	"strings"
	"testing"
)

const acceptanceLogicalText = "# Acceptance\n" +
	"\n" +
	"### AC-FIX-001 — first criterion\n" +
	"### AC-FIX-002 [RETIRED] — withdrawn criterion\n" +
	"| AC-FIX-003 | table row |\n" +
	"- AC-FIX-004a sub-lettered criterion\n"

func sha256Hex(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

// TestAC_CONTRACT_005 — acceptance hash normalization (REQ-CONTRACT-005).
func TestAC_CONTRACT_005(t *testing.T) {
	lf := []byte(acceptanceLogicalText)
	crlf := []byte(strings.ReplaceAll(acceptanceLogicalText, "\n", "\r\n"))
	bom := append([]byte("\xEF\xBB\xBF"), lf...)
	variants := []struct {
		name string
		raw  []byte
	}{{"lf", lf}, {"crlf", crlf}, {"bom+lf", bom}}

	want := sha256Hex(lf)
	wantCount := ACCountResult{Prefix: "AC", Live: 3, Excluded: 1}

	for _, v := range variants {
		t.Run(v.name, func(t *testing.T) {
			if got := AcceptanceHash(v.raw); got != want {
				t.Fatalf("AcceptanceHash(%s) = %q, want %q (SHA-256 of the LF bytes)", v.name, got, want)
			}
			got, err := CountAC(NormalizeAcceptance(v.raw))
			if err != nil {
				t.Fatalf("CountAC(%s): %v", v.name, err)
			}
			if !equalCount(got, wantCount) {
				t.Fatalf("CountAC(%s) = %+v, want %+v", v.name, got, wantCount)
			}
		})
	}

	t.Run("one changed character changes the hash", func(t *testing.T) {
		changed := []byte(strings.Replace(acceptanceLogicalText, "first criterion", "first criterioN", 1))
		if AcceptanceHash(changed) == want {
			t.Fatal("hash unchanged after a one-character edit")
		}
		if AcceptanceHash(changed) == "" {
			t.Fatal("hash empty")
		}
	})
}

func equalCount(a, b ACCountResult) bool {
	return a.Prefix == b.Prefix && a.Live == b.Live && a.Excluded == b.Excluded &&
		slices.Equal(a.Ambiguous, b.Ambiguous)
}

func TestNormalizeAcceptance(t *testing.T) {
	cases := []struct {
		name, in, want string
	}{
		{"empty", "", ""},
		{"lf untouched", "a\nb\n", "a\nb\n"},
		{"crlf to lf", "a\r\nb\r\n", "a\nb\n"},
		{"lone cr kept", "a\rb\n", "a\rb\n"},
		{"cr cr lf keeps first cr", "a\r\r\n", "a\r\n"},
		{"one leading bom stripped", "\xEF\xBB\xBFa\n", "a\n"},
		{"only one bom stripped", "\xEF\xBB\xBF\xEF\xBB\xBFa", "\xEF\xBB\xBFa"},
		{"inner bom kept", "a\xEF\xBB\xBFb", "a\xEF\xBB\xBFb"},
		{"bom then crlf", "\xEF\xBB\xBFa\r\n", "a\n"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := NormalizeAcceptance([]byte(c.in)); string(got) != c.want {
				t.Fatalf("NormalizeAcceptance(%q) = %q, want %q", c.in, got, c.want)
			}
		})
	}
}

func TestNormalizeAcceptance_DoesNotMutateOrAliasInput(t *testing.T) {
	in := []byte("a\nb\n")
	orig := bytes.Clone(in)
	out := NormalizeAcceptance(in)
	if len(out) > 0 {
		out[0] = 'Z'
	}
	if !bytes.Equal(in, orig) {
		t.Fatalf("input mutated: %q", in)
	}
	crlf := []byte("a\r\n")
	_ = NormalizeAcceptance(crlf)
	if string(crlf) != "a\r\n" {
		t.Fatalf("input mutated: %q", crlf)
	}
}

func TestAcceptanceHash_Empty(t *testing.T) {
	if got, want := AcceptanceHash(nil), sha256Hex(nil); got != want {
		t.Fatalf("AcceptanceHash(nil) = %q, want %q", got, want)
	}
	if got, want := AcceptanceHash([]byte("\xEF\xBB\xBF")), sha256Hex(nil); got != want {
		t.Fatalf("AcceptanceHash(BOM only) = %q, want %q", got, want)
	}
}

func TestCountAC_Branches(t *testing.T) {
	cases := []struct {
		name string
		text string
		want ACCountResult
	}{
		{
			name: "empty file counts zero",
			text: "",
			want: ACCountResult{Prefix: "AC"},
		},
		{
			name: "markup shapes and distinct ids",
			text: "### AC-SYN-001 — heading\n| AC-SYN-003 | cell |\n- AC-SYN-05 inline\nAC-001 bare AC-001 again\n",
			want: ACCountResult{Prefix: "AC", Live: 4},
		},
		{
			name: "sub-letter is its own criterion",
			text: "AC-003\nAC-003a\n",
			want: ACCountResult{Prefix: "AC", Live: 2},
		},
		{
			name: "digit-bearing domains and four segments",
			text: "AC-1A-B2-007 and AC-X-Y-Z-9\n",
			want: ACCountResult{Prefix: "AC", Live: 2},
		},
		{
			name: "several ids on one line",
			text: "AC-001, AC-002 [REF], AC-003\n",
			want: ACCountResult{Prefix: "AC", Live: 2, Excluded: 1},
		},
		{
			name: "retired and ref markers with spaces and tabs",
			text: "AC-001 [RETIRED]\nAC-002\t [REF] x\nAC-003[REF]\n",
			want: ACCountResult{Prefix: "AC", Excluded: 3},
		},
		{
			name: "marker must follow immediately",
			text: "AC-001 is [RETIRED]\nAC-002 retired\nAC-003 [retired]\n",
			want: ACCountResult{Prefix: "AC", Live: 3},
		},
		{
			name: "ambiguity in first-seen order",
			text: "AC-002 [REF]\nAC-001\nAC-001 [RETIRED]\nAC-002\nAC-003 [REF]\nAC-004\n",
			want: ACCountResult{Prefix: "AC", Live: 1, Excluded: 1, Ambiguous: []string{"AC-002", "AC-001"}},
		},
		{
			name: "prefix declaration replaces default and applies to its own line",
			text: "AC-001 before\n<!-- moai-ac-prefix: CR --> CR-09\nCR-01\nCR-02 [RETIRED]\nAC-EXT-01 [REF]\nAC-002\n",
			want: ACCountResult{Prefix: "CR", Live: 3, Excluded: 1},
		},
		{
			name: "first declaration wins",
			text: "<!-- moai-ac-prefix: CR -->\n<!-- moai-ac-prefix: AC -->\nCR-01\nAC-01\n",
			want: ACCountResult{Prefix: "CR", Live: 1},
		},
		{
			name: "empty declaration does not count as declared",
			text: "<!-- moai-ac-prefix:  -->\n<!-- moai-ac-prefix: CR -->\nCR-01\nAC-01\n",
			want: ACCountResult{Prefix: "CR", Live: 1},
		},
		{
			name: "declaration must start at column zero",
			text: " <!-- moai-ac-prefix: CR -->\nCR-01\nAC-01\n",
			want: ACCountResult{Prefix: "AC", Live: 1},
		},
		{
			name: "declaration without spaces",
			text: "<!--moai-ac-prefix:CR-->\nCR-01\n",
			want: ACCountResult{Prefix: "CR", Live: 1},
		},
		{
			name: "declaration spaces inside are removed",
			text: "<!-- moai-ac-prefix: C R -->\nCR-01\n",
			want: ACCountResult{Prefix: "CR", Live: 1},
		},
		{
			name: "declaration without closing marker keeps the rest",
			text: "<!-- moai-ac-prefix: CR\nCR-01\n",
			want: ACCountResult{Prefix: "CR", Live: 1},
		},
		{
			name: "alternation prefix",
			text: "<!-- moai-ac-prefix: AC|FLH -->\nAC-001\nFLH-002\nCR-003\n",
			want: ACCountResult{Prefix: "AC|FLH", Live: 2},
		},
		{
			name: "comma list is inserted literally, as awk does",
			text: "<!-- moai-ac-prefix: CR, FLH -->\nCR-01\nFLH-02\nCR,FLH-03\n",
			want: ACCountResult{Prefix: "CR,FLH", Live: 1},
		},
		{
			name: "text after the closing marker is dropped",
			text: "<!-- moai-ac-prefix: CR --> trailing words\nCR-01\n",
			want: ACCountResult{Prefix: "CR", Live: 1},
		},
		{
			name: "no trailing newline",
			text: "AC-001\nAC-002",
			want: ACCountResult{Prefix: "AC", Live: 2},
		},
		{
			name: "non-ascii text around ids",
			text: "### AC-KO-001 — 한국어 설명 AC-KO-002 [REF]\n",
			want: ACCountResult{Prefix: "AC", Live: 1, Excluded: 1},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := CountAC([]byte(c.text))
			if err != nil {
				t.Fatalf("CountAC: %v", err)
			}
			if !equalCount(got, c.want) {
				t.Fatalf("CountAC = %+v, want %+v", got, c.want)
			}
			if got.IsAmbiguous() != (len(c.want.Ambiguous) > 0) {
				t.Fatalf("IsAmbiguous = %v", got.IsAmbiguous())
			}
		})
	}
}

func TestCountAC_InvalidPrefixIsUnusable(t *testing.T) {
	_, err := CountAC([]byte("<!-- moai-ac-prefix: AC( -->\nAC-001\n"))
	if !errors.Is(err, ErrACPrefixInvalid) {
		t.Fatalf("err = %v, want ErrACPrefixInvalid", err)
	}
}
