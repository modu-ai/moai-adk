package core

import (
	"math/rand"
	"reflect"
	"strings"
	"testing"
)

// REQ-AF-013 coverage (DEC-AF-29): contact numbers, a name beside one, a stack
// trace and a passphrase are text a ledger must not carry.

func TestSanitizeMasksPhoneNumbers(t *testing.T) {
	for _, c := range []struct{ name, in, number string }{
		{"mobile with dashes", "call 010-1234-5678 today", "010-1234-5678"},
		{"mobile with three digits", "call 010-234-5678 today", "010-234-5678"},
		{"mobile without separators", "call 01012345678 today", "01012345678"},
		{"mobile with spaces", "call 010 1234 5678 today", "010 1234 5678"},
		{"mobile with dots", "call 010.1234.5678 today", "010.1234.5678"},
		{"international", "call +82 10-1234-5678 today", "+82 10-1234-5678"},
		{"international without plus", "call 82-10-1234-5678 today", "82-10-1234-5678"},
		{"seoul landline", "call 02-1234-5678 today", "02-1234-5678"},
		{"seoul landline short", "call 02-123-4567 today", "02-123-4567"},
		{"regional landline", "call 031-123-4567 today", "031-123-4567"},
		{"internet phone", "call 070-1234-5678 today", "070-1234-5678"},
		{"in parentheses", "call (010-1234-5678) today", "010-1234-5678"},
	} {
		t.Run(c.name, func(t *testing.T) {
			out, cats := Sanitize(c.in)
			if strings.Contains(out, c.number) || strings.Contains(out, "1234-5678") || strings.Contains(out, "12345678") {
				t.Errorf("the number survived: %q", out)
			}
			if !reflect.DeepEqual(cats, []Category{CategoryPhone}) {
				t.Errorf("categories = %v, want the phone number", cats)
			}
			if again, more := Sanitize(out); again != out || len(more) != 0 {
				t.Errorf("not idempotent: %q -> %q %v", out, again, more)
			}
		})
	}
}

func TestSanitizeMasksANameBesideAPhoneNumber(t *testing.T) {
	for _, c := range []struct{ name, in, person string }{
		{"name then number", "담당 홍길동 010-1234-5678 입니다", "홍길동"},
		{"name with honorific", "연락 홍길동님 010-1234-5678 로", "홍길동"},
		{"name with san", "연락 김철수씨 010-1234-5678 로", "김철수"},
		{"name and number in parentheses", "작성자 이영희 (010-1234-5678)", "이영희"},
		{"name, comma, number", "신청인 박민준, 010-1234-5678", "박민준"},
		{"name with colon", "박지민: 010-1234-5678", "박지민"},
		{"four-syllable name", "남궁민수 010-1234-5678", "남궁민수"},
		{"four-syllable name with honorific", "남궁민수님 010-1234-5678", "남궁민수"},
	} {
		t.Run(c.name, func(t *testing.T) {
			out, cats := Sanitize(c.in)
			if strings.Contains(out, c.person) || strings.Contains(out, "5678") {
				t.Errorf("name or number survived: %q", out)
			}
			if !reflect.DeepEqual(cats, []Category{CategoryPersonName, CategoryPhone}) {
				t.Errorf("categories = %v, want person-name and phone", cats)
			}
			if again, more := Sanitize(out); again != out || len(more) != 0 {
				t.Errorf("not idempotent: %q -> %q %v", out, again, more)
			}
		})
	}
}

// A name alone cannot be told from any other Korean word, so only one beside a
// contact number is masked; ordinary Korean text and numbers stay.
func TestSanitizeLeavesKoreanTextAndOrdinaryNumbersAlone(t *testing.T) {
	for _, in := range []string{
		"홍길동이 작성한 보고서를 검토했습니다",
		"담당자는 김철수이고 리뷰어는 이영희입니다",
		"카드 t1948 의 상태를 확인했습니다",
		"2026-10-30 09:41 에 커밋했습니다",
		"20261029T031506Z",
		"issue-1948-attempt-1-repair-3",
		"0123456789",
		"1234-5678",
		"port 8080 and 443",
		"v3.2.0-rc.38 build 1948",
		"head 6b61a2cc07 passed",
		"digits 0123456789012345678901234567890123456789",
		"range 010-1234 only",
		"id 1010-1234-5678-9012",
	} {
		if out, cats := Sanitize(in); out != in || len(cats) != 0 {
			t.Errorf("over-sanitized %q -> %q %v", in, out, cats)
		}
	}
}

const sampleStack = "run failed:\n" +
	"panic: runtime error: index out of range [3] with length 2\n" +
	"\n" +
	"goroutine 1 [running]:\n" +
	"main.process(0xc000012345, {0x1234, 0x5})\n" +
	"\t/home/ci/work/src/app/main.go:42 +0x1d\n" +
	"main.main()\n" +
	"\t/usr/local/go/src/main.go:12 +0x1d\n" +
	"created by main.start in goroutine 1\n" +
	"\t/usr/local/go/src/start.go:7 +0x5c\n" +
	"exit status 2\n" +
	"after the stack\n"

func TestSanitizeMasksAStackTrace(t *testing.T) {
	out, cats := Sanitize(sampleStack)
	for _, leak := range []string{"main.process", "0xc000012345", "/home/ci", "/usr/local/go", "index out of range", "goroutine 1", "created by"} {
		if strings.Contains(out, leak) {
			t.Errorf("%q survived: %q", leak, out)
		}
	}
	if !strings.HasPrefix(out, "run failed:\n") || !strings.HasSuffix(out, "after the stack\n") {
		t.Errorf("the text around the stack must stay: %q", out)
	}
	found := false
	for _, c := range cats {
		if c == CategoryStack {
			found = true
		}
	}
	if !found {
		t.Errorf("categories = %v, want the stack", cats)
	}
	if again, more := Sanitize(out); again != out || len(more) != 0 {
		t.Errorf("not idempotent: %q -> %q %v", out, again, more)
	}
}

func TestSanitizeMasksALonePanicLineAndAGoroutineHeader(t *testing.T) {
	for _, in := range []string{
		"panic: open /tmp/secret-file: permission denied",
		"goroutine 7 [running]:\nmain.main()\n\t/usr/local/go/src/main.go:12 +0x1d",
	} {
		out, cats := Sanitize(in)
		found := false
		for _, c := range cats {
			if c == CategoryStack {
				found = true
			}
		}
		if !found || strings.Contains(out, "main.main") || strings.Contains(out, "secret-file") {
			t.Errorf("%q -> %q %v", in, out, cats)
		}
	}
}

func TestSanitizeLeavesTalkAboutStacksAlone(t *testing.T) {
	for _, in := range []string{
		"the test panics when the input is empty",
		"a panic: in the handler was fixed",
		"goroutine leak found in the monitor",
		"see the stack in the CI log",
		"main.go:42 needs a nil check",
	} {
		if out, cats := Sanitize(in); out != in || len(cats) != 0 {
			t.Errorf("over-sanitized %q -> %q %v", in, out, cats)
		}
	}
}

func TestSanitizeMasksAPassphraseAssignment(t *testing.T) {
	for _, in := range []string{
		"passphrase: correct-horse-battery-staple",
		"PASSPHRASE=correct-horse-battery-staple ./run",
		`{"passphrase": "correct-horse-battery-staple"}`,
		"ssh_passphrase = 'correct horse battery staple'",
	} {
		out, cats := Sanitize(in)
		if strings.Contains(out, "horse") || !reflect.DeepEqual(cats, []Category{CategoryCredentialAssignment}) {
			t.Errorf("%q -> %q %v", in, out, cats)
		}
	}
	for _, in := range []string{
		"the passphrase policy changed",
		"rotate every passphrase quarterly",
	} {
		if out, cats := Sanitize(in); out != in || len(cats) != 0 {
			t.Errorf("over-sanitized %q -> %q %v", in, out, cats)
		}
	}
}

func TestSanitizeMasksNumbersSharingASeparator(t *testing.T) {
	out, cats := Sanitize("call 010-1234-5678,010-8765-4321 now")
	if strings.ContainsAny(out, "0123456789") || !reflect.DeepEqual(cats, []Category{CategoryPhone}) {
		t.Errorf("%q %v", out, cats)
	}
}

// A word that labels a number is not a person, and a longer word is not a name.
func TestSanitizeKeepsALabelAndALongWordBesideANumber(t *testing.T) {
	for _, c := range []struct{ in, kept string }{
		{"연락처 010-1234-5678", "연락처"},
		{"전화번호: 010-1234-5678", "전화번호"},
		{"국민은행계좌 010-1234-5678", "국민은행계좌"},
	} {
		out, cats := Sanitize(c.in)
		if !strings.Contains(out, c.kept) || !reflect.DeepEqual(cats, []Category{CategoryPhone}) {
			t.Errorf("%q -> %q %v", c.in, out, cats)
		}
	}
}

func TestSanitizeStackReplacementKeepsTheLineStructure(t *testing.T) {
	out, _ := Sanitize(sampleStack)
	if out != "run failed:\n[REDACTED]\nafter the stack\n" {
		t.Errorf("got %q", out)
	}
}

// A commit or content hash is full of digit runs that look like telephone numbers.
// The characters beside a number must not be ASCII letters or digits, so no digit run
// inside a hex token is one; ledgers and evidence carry the full head SHA, and a
// masked SHA would turn every one of them into a refusal.
func TestSanitizeNeverTouchesAHexDigest(t *testing.T) {
	rng := rand.New(rand.NewSource(11))
	const hex = "0123456789abcdef"
	digest := func(n int) string {
		b := make([]byte, n)
		for i := range b {
			b[i] = hex[rng.Intn(16)]
		}
		return string(b)
	}
	// digests that carry a number-shaped digit run by construction
	fixed := []string{
		"a01012345678b0123456789abcdef01234567890",
		"ff0212345678ab0123456789abcdef0123456701",
		"0101234567890123456789abcdef0123456789ab",
		"abcdef0123456789abcdef0101234567890123ab",
		"c0311234567d0123456789abcdef0123456789ab",
	}
	// digests that start with a number-shaped digit run
	fixed = append(fixed,
		"01012345678abcdef0123456789abcdef0123456",
		"0212345678abcdef0123456789abcdef01234567",
		("01012345678" + strings.Repeat("abcdef0123456789", 4))[:64],
	)
	for i := 0; i < 4000; i++ {
		fixed = append(fixed, digest(40))
	}
	for i := 0; i < 1000; i++ {
		fixed = append(fixed, digest(64))
	}
	contexts := []string{"%s", "head %s ok", "sha:%s", "(%s)", "%s.", "refs/heads/x@%s", "rev=%s;"}
	changed := 0
	for _, d := range fixed {
		for _, c := range contexts {
			in := strings.Replace(c, "%s", d, 1)
			if out, cats := Sanitize(in); out != in || len(cats) != 0 {
				if changed < 5 {
					t.Errorf("a hex digest was touched: %q -> %q %v", in, out, cats)
				}
				changed++
			}
		}
	}
	if changed > 0 {
		t.Errorf("%d of %d digest sightings changed", changed, len(fixed)*len(contexts))
	}
}

func TestSanitizeMasksAParenthesisedAreaCode(t *testing.T) {
	for _, in := range []string{"call (02) 123-4567 now", "call (031)123-4567 now", "call (02)1234-5678 now", "call (010) 1234-5678 now"} {
		out, cats := Sanitize(in)
		if strings.ContainsAny(out, "0123456789") || !reflect.DeepEqual(cats, []Category{CategoryPhone}) {
			t.Errorf("%q -> %q %v", in, out, cats)
		}
	}
}

// A job title between the name and the number does not hide the name.
func TestSanitizeMasksANameBehindATitle(t *testing.T) {
	for _, c := range []struct{ in, person string }{
		{"담당 홍길동 대리 010-1234-5678", "홍길동"},
		{"연락 김철수 과장님 010-1234-5678", "김철수"},
		{"이영희 팀장 (010-1234-5678)", "이영희"},
		{"박민준 센터장: 02-123-4567", "박민준"},
	} {
		out, cats := Sanitize(c.in)
		if strings.Contains(out, c.person) || strings.ContainsAny(out, "0123456789") || !reflect.DeepEqual(cats, []Category{CategoryPersonName, CategoryPhone}) {
			t.Errorf("%q -> %q %v", c.in, out, cats)
		}
	}
}

// Words that label a number, or a title standing alone, are not a person.
func TestSanitizeKeepsMoreLabelsBesideANumber(t *testing.T) {
	for _, label := range []string{"고객센터", "본사", "지점", "콜센터", "대표번호", "문의처", "담당자", "사무실", "자택", "휴대전화", "비상연락망", "대리", "팀장"} {
		in := label + " 010-1234-5678"
		out, cats := Sanitize(in)
		if !strings.Contains(out, label) || !reflect.DeepEqual(cats, []Category{CategoryPhone}) {
			t.Errorf("%q -> %q %v", in, out, cats)
		}
	}
}

const pythonStack = "run failed:\n" +
	"Traceback (most recent call last):\n" +
	"  File \"/home/ci/app/main.py\", line 10, in <module>\n" +
	"    run(secret_token)\n" +
	"  File \"/home/ci/app/lib.py\", line 4, in run\n" +
	"    raise ValueError(token)\n" +
	"ValueError: bad token abc123\n" +
	"after the stack\n"

const javaStack = "run failed:\n" +
	"Exception in thread \"main\" java.lang.IllegalStateException: boom user=bob\n" +
	"\tat com.acme.App.run(App.java:42)\n" +
	"\tat com.acme.App.main(App.java:10)\n" +
	"Caused by: java.io.IOException: disk /srv/data\n" +
	"\tat com.acme.Io.read(Io.java:7)\n" +
	"\t... 2 more\n" +
	"after the stack\n"

const nodeStack = "run failed:\n" +
	"TypeError: Cannot read properties of undefined (reading 'secret')\n" +
	"    at Object.<anonymous> (/home/ci/app/index.js:10:15)\n" +
	"    at Module._compile (node:internal/modules/cjs/loader:1256:14)\n" +
	"    at /home/ci/app/run.js:3:1\n" +
	"after the stack\n"

func TestSanitizeMasksPythonJavaAndNodeStacks(t *testing.T) {
	for _, c := range []struct {
		name, in string
		leaks    []string
	}{
		{"python", pythonStack, []string{"/home/ci", "secret_token", "abc123", "ValueError"}},
		{"java", javaStack, []string{"com.acme", "user=bob", "/srv/data", "IllegalState"}},
		{"node", nodeStack, []string{"/home/ci", "secret", "Module._compile"}},
	} {
		t.Run(c.name, func(t *testing.T) {
			out, cats := Sanitize(c.in)
			for _, leak := range c.leaks {
				if strings.Contains(out, leak) {
					t.Errorf("%q survived: %q", leak, out)
				}
			}
			if out != "run failed:\n[REDACTED]\nafter the stack\n" {
				t.Errorf("one block replaced, the lines around it kept: %q", out)
			}
			found := false
			for _, k := range cats {
				if k == CategoryStack {
					found = true
				}
			}
			if !found {
				t.Errorf("categories = %v, want the stack", cats)
			}
			if again, more := Sanitize(out); again != out || len(more) != 0 {
				t.Errorf("not idempotent: %q -> %q %v", out, again, more)
			}
		})
	}
}

func TestSanitizeLeavesIndentedProseAndErrorNamesAlone(t *testing.T) {
	for _, in := range []string{
		"the handler throws a ValueError: bad input when it is called",
		"steps:\n   at the beginning\n   at the end\n",
		"Traceback is the word for it",
		"  at home (office)\n",
		"java.lang.NullPointerException is thrown by the helper",
		"  File names are listed below",
	} {
		if out, cats := Sanitize(in); out != in || len(cats) != 0 {
			t.Errorf("over-sanitized %q -> %q %v", in, out, cats)
		}
	}
}

// A Java trace without a cause, ending in an elision line.
func TestSanitizeMasksAJavaTraceWithoutACause(t *testing.T) {
	in := "start\njava.lang.IllegalStateException: boom\n\tat com.acme.App.run(App.java:42)\n\t... 3 more\nend\n"
	out, cats := Sanitize(in)
	if out != "start\n[REDACTED]\nend\n" || len(cats) != 1 || cats[0] != CategoryStack {
		t.Errorf("got %q %v", out, cats)
	}
}
