package core

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

// One sanitizer decides what may leave the process. Each category has a case with
// the secret that must not survive and the category that must be reported, so a
// shape that one of the old sanitizers knew and another did not is pinned here.
var sanitizeCases = []struct {
	name     string
	in       string
	secret   string
	category Category
}{
	{"github token", "use ghp_abcdefghijklmnopqrstuvwxyz0123 now", "ghp_abcdefghijklmnopqrstuvwxyz0123", CategoryToken},
	{"fine-grained token", "github_pat_11ABCDEFG0abcdefghijklmnopqrstuvwxyz_0123456789abcdef", "github_pat_11ABCDEFG0abcdefghijklmnopqrstuvwxyz_0123456789abcdef", CategoryToken},
	{"sk key", "key sk-ant-api03-AbCdEf0123456789abcdef used", "sk-ant-api03-AbCdEf0123456789abcdef", CategoryToken},
	{"aws key id", "id AKIAIOSFODNN7EXAMPLE here", "AKIAIOSFODNN7EXAMPLE", CategoryToken},
	{"aws temporary key id", "id ASIAIOSFODNN7EXAMPLE here", "ASIAIOSFODNN7EXAMPLE", CategoryToken},
	{"bearer word", "send Bearer abcdefgh12345678 along", "abcdefgh12345678", CategoryAuthorization},
	{"authorization bearer", "Authorization: Bearer eyJhbGciOiJIUzI1NiJ9.payload.sig", "eyJhbGciOiJIUzI1NiJ9.payload.sig", CategoryAuthorization},
	{"authorization basic", "authorization: basic dXNlcjpwYXNzd29yZA==", "dXNlcjpwYXNzd29yZA==", CategoryAuthorization},
	{"cookie header", "Cookie: session=abcdef0123456789; theme=dark", "abcdef0123456789", CategoryCookie},
	{"set-cookie header", "Set-Cookie: sid=zzzz9999yyyy; Path=/", "zzzz9999yyyy", CategoryCookie},
	{"url password", "clone https://user:s3cr3tPass@github.com/o/r.git", "s3cr3tPass", CategoryURLCredential},
	{"env assignment", "GITHUB_TOKEN=abc123def456ghi789 ./run", "abc123def456ghi789", CategoryCredentialAssignment},
	{"password assignment", "DB_PASSWORD=hunter2hunter2 start", "hunter2hunter2", CategoryCredentialAssignment},
	{"colon assignment", "api_key: zzz999 ok", "zzz999", CategoryCredentialAssignment},
	{"json secret", `{"api_key": "zzzz-1234-yyyy-5678"}`, "zzzz-1234-yyyy-5678", CategoryCredentialAssignment},
	{"private key block", "x\n-----BEGIN RSA PRIVATE KEY-----\nMIIEowIBAAKCAQEA0123456789abcdef\n-----END RSA PRIVATE KEY-----\ny", "MIIEowIBAAKCAQEA0123456789abcdef", CategoryPrivateKey},
	{"openssh private key", "-----BEGIN OPENSSH PRIVATE KEY-----\nb3BlbnNzaC1rZXktdjEAAAAABG5vbmU=\n-----END OPENSSH PRIVATE KEY-----", "b3BlbnNzaC1rZXktdjEAAAAABG5vbmU=", CategoryPrivateKey},
	{"truncated private key", "head\n-----BEGIN PRIVATE KEY-----\nMIIBVQIBADANBgkqhkiG9w0BAQEFAASC", "MIIBVQIBADANBgkqhkiG9w0BAQEFAASC", CategoryPrivateKey},
	{"email", "ask dev.person+x@example.co.kr about it", "dev.person+x@example.co.kr", CategoryEmail},
	{"ssh remote", "git@github.com:org/repo.git", "git@github.com", CategoryEmail},
	{"mac home", "see /Users/goos/MoAI/moai-adk-go/internal/x.go for details", "/Users/goos", CategoryHomePath},
	{"linux home", "log at /home/runner/work/_temp/out.log", "/home/runner", CategoryHomePath},
	{"windows home", `file C:\Users\Goos\AppData\x.txt here`, `C:\Users\Goos`, CategoryHomePath},
}

func TestSanitizeMasksEveryCategoryAndNamesIt(t *testing.T) {
	for _, c := range sanitizeCases {
		t.Run(c.name, func(t *testing.T) {
			out, cats := Sanitize(c.in)
			if strings.Contains(out, c.secret) {
				t.Errorf("the secret survived: %q", out)
			}
			found := false
			for _, got := range cats {
				if got == c.category {
					found = true
				}
			}
			if !found {
				t.Errorf("category %q not reported, got %v", c.category, cats)
			}
			if again, more := Sanitize(out); again != out || len(more) != 0 {
				t.Errorf("sanitizing is not idempotent: %q -> %q %v", out, again, more)
			}
		})
	}
}

func TestSanitizeLeavesOrdinaryTextAlone(t *testing.T) {
	for _, in := range []string{
		"internal/agenticfactory/claim/claim.go:42",
		"/usr/local/bin/moai and /tmp/work",
		"@MX:NOTE reviewed by the team",
		"user@host without a domain suffix",
		"state=verifying head=cb30e05b3e",
		"head 0123456789abcdef0123456789abcdef01234567 build passed",
		".moai/specs/SPEC-AGENTIC-FACTORY-001/spec.md",
		"-----BEGIN CERTIFICATE-----\nMIIB\n-----END CERTIFICATE-----",
		"the key was rotated and the password policy changed",
	} {
		if out, cats := Sanitize(in); out != in || len(cats) != 0 {
			t.Errorf("over-sanitized %q -> %q %v", in, out, cats)
		}
	}
}

func TestSanitizeReportsEachCategoryOnceInOrder(t *testing.T) {
	in := "mail a@b.example.com and a@c.example.com, token=abc12345 and Cookie: s=1234abcd"
	_, cats := Sanitize(in)
	want := []Category{CategoryCookie, CategoryCredentialAssignment, CategoryEmail}
	if !reflect.DeepEqual(cats, want) {
		t.Errorf("categories = %v, want %v (each once, sorted)", cats, want)
	}
}

// The old Redact keeps its contract: the masked text and how many values were
// masked, and it now knows every shape the sanitizer knows.
func TestRedactIsTheSanitizerWithACount(t *testing.T) {
	for _, c := range sanitizeCases {
		out, n := Redact(c.in)
		if n == 0 || strings.Contains(out, c.secret) {
			t.Errorf("%s: Redact missed it: %q n=%d", c.name, out, n)
		}
	}
}

// A caller that must not send anything sensitive refuses instead of masking: the
// error names the categories and never the values.
func TestRefuseIfSensitive(t *testing.T) {
	if err := RefuseIfSensitive("nothing to hide here"); err != nil {
		t.Errorf("clean text refused: %v", err)
	}
	err := RefuseIfSensitive("mail dev@example.com with token=abc12345def")
	if !errors.Is(err, ErrUnsanitized) {
		t.Fatalf("sensitive text: %v", err)
	}
	for _, leak := range []string{"dev@example.com", "abc12345def"} {
		if strings.Contains(err.Error(), leak) {
			t.Errorf("the refusal repeats the value %q: %v", leak, err)
		}
	}
	for _, name := range []string{string(CategoryEmail), string(CategoryCredentialAssignment)} {
		if !strings.Contains(err.Error(), name) {
			t.Errorf("the refusal must name %q: %v", name, err)
		}
	}
}
