package monitor

import (
	"context"
	"strings"
	"testing"
)

// Each case holds a secret that must not survive masking in any output.
var secretCases = []struct{ name, in, secret string }{
	{"github pat", "token ghp_abcdefghijklmnopqrstuvwxyz0123456789 used", "ghp_abcdefghijklmnopqrstuvwxyz0123456789"},
	{"fine-grained pat", "github_pat_11ABCDEFG0abcdefghijklmnopqrstuvwxyz_0123456789abcdef", "github_pat_11ABCDEFG0abcdefghijklmnopqrstuvwxyz_0123456789abcdef"},
	{"api key sk", "key=sk-ant-api03-AbCdEf0123456789abcdef", "sk-ant-api03-AbCdEf0123456789abcdef"},
	{"bearer header", "Authorization: Bearer eyJhbGciOiJIUzI1NiJ9.payload.sig", "eyJhbGciOiJIUzI1NiJ9.payload.sig"},
	{"basic header", "authorization: basic dXNlcjpwYXNzd29yZA==", "dXNlcjpwYXNzd29yZA=="},
	{"url userinfo", "clone https://user:s3cr3tPass@github.com/o/r.git", "s3cr3tPass"},
	{"env assignment", "GITHUB_TOKEN=abc123def456ghi789 ./run", "abc123def456ghi789"},
	{"env password", "DB_PASSWORD=hunter2hunter2 start", "hunter2hunter2"},
	{"cookie", "Cookie: session=abcdef0123456789; theme=dark", "abcdef0123456789"},
	{"aws key id", "id AKIAIOSFODNN7EXAMPLE here", "AKIAIOSFODNN7EXAMPLE"},
	{"json secret", `{"api_key": "zzzz-1234-yyyy-5678"}`, "zzzz-1234-yyyy-5678"},
}

func TestMaskRemovesSecrets(t *testing.T) {
	for _, c := range secretCases {
		t.Run(c.name, func(t *testing.T) {
			out, masked := Mask(c.in)
			if strings.Contains(out, c.secret) {
				t.Errorf("secret survived: %q", out)
			}
			if !masked {
				t.Error("masked flag not set")
			}
		})
	}
}

func TestMaskLeavesPlainTextAlone(t *testing.T) {
	in := "issue 1948 head 0123456789abcdef0123456789abcdef01234567 build passed"
	out, masked := Mask(in)
	if out != in || masked {
		t.Errorf("plain text changed: %q masked=%v", out, masked)
	}
}

// Fail-closed: a field whose text still looks like a secret after masking is omitted, never shown.
func TestFailClosedOmitsField(t *testing.T) {
	f := textField("next", "run with ghp_abcdefghijklmnopqrstuvwxyz0123456789", SrcLedger, t0)
	if strings.Contains(f.Value, "ghp_") {
		t.Errorf("secret in field: %q", f.Value)
	}
	if !f.Redacted {
		t.Error("redaction not recorded")
	}
}

func TestFailClosedOmitsOpaqueBlob(t *testing.T) {
	f := textField("note", "see Zx9Kq2Lm7Pv4Rt8Wn3Yb6Hc1Dj5Fg0Aa2Ss4Dd", SrcLedger, t0)
	if f.Value != "" || !f.Redacted {
		t.Errorf("opaque blob shown: %+v", f)
	}
	h := textField("head", "head 0123456789abcdef0123456789abcdef01234567", SrcClaimStore, t0)
	if h.Redacted || !strings.Contains(h.Value, "0123456789abcdef") {
		t.Errorf("HEAD SHA wrongly redacted: %+v", h)
	}
}

func TestCollectedRowNeverCarriesSecret(t *testing.T) {
	c := collector(
		fakeClaims{rows: map[int]ClaimRow{7: {Issue: 7, Owner: "lane-18", BlockReason: "auth-failed", ResumeCondition: "token ghp_abcdefghijklmnopqrstuvwxyz0123456789 rotated"}}},
		fakeRemote{}, fakeLedger{next: "Authorization: Bearer abc.def.ghi"},
	)
	r := c.Collect(context.Background(), []int{7})[0]
	all := r.ResumeCondition.Value + r.NextAction.Value
	for _, s := range []string{"ghp_abcdefghijklmnopqrstuvwxyz0123456789", "abc.def.ghi"} {
		if strings.Contains(all, s) {
			t.Errorf("row leaks %q", s)
		}
	}
}
