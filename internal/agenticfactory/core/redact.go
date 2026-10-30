package core

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// Category names a kind of sensitive value the sanitizer finds.
type Category string

const (
	CategoryToken                Category = "token"                 // service tokens and key ids (gh*, github_pat, sk-, AWS)
	CategoryAuthorization        Category = "authorization"         // an Authorization header or a bearer credential
	CategoryCookie               Category = "cookie"                // a Cookie or Set-Cookie header
	CategoryURLCredential        Category = "url-credential"        // a password inside a URL
	CategoryCredentialAssignment Category = "credential-assignment" // name=value or "name": "value" for a secret-named field
	CategoryPrivateKey           Category = "private-key"           // a PEM private key block
	CategoryEmail                Category = "email"                 // an e-mail address, ssh user@host included
	CategoryHomePath             Category = "home-path"             // an absolute home directory and what is under it
	CategoryPhone                Category = "phone"                 // a Korean or +82 telephone number
	CategoryPersonName           Category = "person-name"           // a Korean name written beside a telephone number
	CategoryStack                Category = "stack"                 // a panic line or a goroutine stack trace
	CategoryOpaqueToken          Category = "opaque-token"          // a long mixed letter-and-digit run no other rule knows
)

const redacted = "[REDACTED]"

// sanRule finds one category. repl builds the replacement from the match's groups
// and reports whether it changed anything: a value that is already the placeholder
// is left alone, so sanitizing twice finds nothing the second time.
type sanRule struct {
	cat  Category
	re   *regexp.Regexp
	repl func(g []string) (string, bool)
}

// whole replaces the whole match.
func whole(g []string) (string, bool) { return redacted, true }

// keepFirst keeps group 1 (the name or the header) and replaces group 2 (the value);
// groups after 2 are kept (a closing quote, the @ of a URL).
func keepGroups(valueGroup int) func(g []string) (string, bool) {
	return func(g []string) (string, bool) {
		if g[valueGroup] == redacted {
			return g[0], false
		}
		var b strings.Builder
		for i := 1; i < len(g); i++ {
			if i == valueGroup {
				b.WriteString(redacted)
			} else {
				b.WriteString(g[i])
			}
		}
		return b.String(), true
	}
}

// The rules are applied in order; an earlier rule hides a value from a later one
// (an Authorization header is masked before the bare bearer rule sees it).
var sanRules = []sanRule{
	{CategoryPrivateKey, regexp.MustCompile(`(?s)-----BEGIN [A-Z0-9 ]*PRIVATE KEY-----.*?(?:-----END [A-Z0-9 ]*PRIVATE KEY-----|$)`), whole},
	{CategoryAuthorization, regexp.MustCompile(`(?i)(authorization\s*:\s*(?:bearer|basic|token)\s+)(\S+)`), keepGroups(2)},
	{CategoryAuthorization, regexp.MustCompile(`(?i)bearer\s+[A-Za-z0-9._~+/=-]{8,}`), whole},
	{CategoryCookie, regexp.MustCompile(`(?i)((?:set-)?cookie\s*:\s*)([^\n]+)`), keepGroups(2)},
	{CategoryURLCredential, regexp.MustCompile(`(?i)(https?://[^/\s:@]+:)([^@\s/]+)(@)`), keepGroups(2)},
	{CategoryCredentialAssignment, regexp.MustCompile(`(?i)("[a-z_]*(?:api_?key|token|secret|pass(?:phrase|w(?:or)?d)|credential)[a-z_]*"\s*:\s*")([^"]*)(")`), keepGroups(2)},
	{CategoryCredentialAssignment, regexp.MustCompile(`(?i)\b([a-z0-9_]*(?:token|secret|pass(?:phrase|w(?:or)?d)|api[_-]?key|credential|cookie)[a-z0-9_]*\s*[:=]\s*)("[^"]*"|'[^']*'|[^\s",}]+)`), keepGroups(2)},
	{CategoryCredentialAssignment, regexp.MustCompile(`(?i)\b([a-z0-9_]*key[a-z0-9_]*=)("[^"]*"|'[^']*'|\S+)`), keepGroups(2)},
	{CategoryToken, regexp.MustCompile(`gh[pousr]_[A-Za-z0-9]{20,}`), whole},
	{CategoryToken, regexp.MustCompile(`github_pat_[A-Za-z0-9_]{20,}`), whole},
	{CategoryToken, regexp.MustCompile(`sk-[A-Za-z0-9_-]{16,}`), whole},
	{CategoryToken, regexp.MustCompile(`\b(?:AKIA|ASIA)[A-Z0-9]{16}\b`), whole},
	// An e-mail address, including the user@host part of an ssh remote.
	{CategoryEmail, regexp.MustCompile(`[A-Za-z0-9._%+-]+@[A-Za-z0-9-]+(?:\.[A-Za-z0-9-]+)*\.[A-Za-z]{2,}`), whole},
	// An absolute home directory and everything under it. The pattern matches one
	// character of left context so a path is only masked where it starts; that
	// character is put back.
	// A panic and goroutine stack: frame arguments, absolute paths and the panic text
	// can carry values and user names, so the whole block goes. A lone "panic:" line
	// at the start of a line goes too; prose that merely mentions a panic stays.
	{CategoryStack, stackBlock, func(g []string) (string, bool) {
		if strings.HasSuffix(g[0], "\n") {
			return redacted + "\n", true
		}
		return redacted, true
	}},
	{CategoryStack, regexp.MustCompile(`(?m)^panic: [^\n]*`), whole},
	// Python: the header line, the indented File/code lines and the unindented exception line.
	{CategoryStack, regexp.MustCompile(`(?m)^Traceback \(most recent call last\):\n(?:[ \t]+[^\n]*\n)*[^\s][^\n]*(?:\n|$)`), keepNewline},
	// Java and Node: an optional error line, frame lines ("at ..."), and "Caused by:" groups.
	{CategoryStack, javaNodeStack, keepNewline},
	{CategoryHomePath, regexp.MustCompile(`(?:^|[^A-Za-z0-9_./-])/(?:Users|home)/[^\s"'/\\]+(?:/[^\s"']*)?`), func(g []string) (string, bool) {
		if g[0] != "" && g[0][0] != '/' {
			return g[0][:1] + redacted, true
		}
		return redacted, true
	}},
	{CategoryHomePath, regexp.MustCompile(`[A-Za-z]:\\Users\\[^\s"'\\]+(?:\\[^\s"']*)?`), whole},
}

// sanitize masks s and counts what it masked, per category.
func sanitize(s string) (string, map[Category]int) {
	found := map[Category]int{}
	for _, r := range sanRules {
		r := r
		s = r.re.ReplaceAllStringFunc(s, func(m string) string {
			g := r.re.FindStringSubmatch(m)
			out, changed := r.repl(g)
			if changed {
				found[r.cat]++
			}
			return out
		})
	}
	s = maskContacts(s, found)
	// Last, so every named shape has had its turn: a long mixed run that no rule
	// knows may still be a credential in a shape nobody listed, and is masked.
	s = opaqueRun.ReplaceAllStringFunc(s, func(m string) string {
		if looksOpaque(m) {
			found[CategoryOpaqueToken]++
			return redacted
		}
		return m
	})
	// And a mixed run that '/' or '+' cuts into pieces under 32 characters, such as an
	// unnamed AWS secret access key.
	s = maskCutSecrets(s, found)
	return s, found
}

var (
	stackBlock = regexp.MustCompile(`(?m)(?:^panic: [^\n]*\n(?:\[signal [^\n]*\n)?\n?)?^goroutine \d+ \[[^\]\n]*\]:\n(?:(?:\t[^\n]*|[A-Za-z_][^\n ]*\([^\n]*\)|created by [^\n]*|exit status \d+)(?:\n|$))*`)
	cutRun     = regexp.MustCompile(`[A-Za-z0-9/+]{40,}`)
	opaqueRun  = regexp.MustCompile(`[A-Za-z0-9_+=\-]{32,}`)
	hexDigest  = regexp.MustCompile(`^(?:[0-9a-f]{40}|[0-9a-f]{64})$`)
	uuidShape  = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
)

// opaqueTransitionRatio is where an identifier made of words stops and a random
// run begins, for runs that have separators: words change character class rarely
// (a number after a letter), random text almost every other character.
const opaqueTransitionRatio = 0.30

// looksOpaque reports a run of 32 or more characters that mixes letters and digits
// and is not a known harmless shape: a full commit or content hash (40 or 64 hex,
// also as a segment behind a prefix such as "ev-"), a UUID, or an identifier made
// of words and numbers separated by '-' or '_' (a branch, a card or a run name).
// A run without separators that mixes letters and digits is opaque; one with
// separators is opaque only when its segments change character class like random
// text does.
func looksOpaque(run string) bool {
	if uuidShape.MatchString(run) {
		return false
	}
	segs := strings.FieldsFunc(run, func(r rune) bool { return r == '-' || r == '_' || r == '=' })
	var keep []string
	for _, sg := range segs {
		if !hexDigest.MatchString(sg) {
			keep = append(keep, sg)
		}
	}
	var letters, digits bool
	pairs, flips, n := 0, 0, 0
	for _, sg := range keep {
		prev := byte(0)
		for i := 0; i < len(sg); i++ {
			c := classOf(sg[i])
			switch c {
			case 'l', 'u':
				letters = true
			case 'd':
				digits = true
			}
			n++
			if i > 0 {
				pairs++
				if c != prev {
					flips++
				}
			}
			prev = c
		}
	}
	if n < 32-len(segs) || !letters || !digits { // too little left after the harmless parts, or not mixed
		return false
	}
	if len(segs) == 1 {
		return true
	}
	return pairs > 0 && float64(flips)/float64(pairs) >= opaqueTransitionRatio
}

// classOf is the character class a run changes between: lower, upper, digit, other.
func classOf(c byte) byte {
	switch {
	case c >= 'a' && c <= 'z':
		return 'l'
	case c >= 'A' && c <= 'Z':
		return 'u'
	case c >= '0' && c <= '9':
		return 'd'
	}
	return 'o'
}

// Sanitize masks every secret and piece of personal data it knows and reports
// which categories it found, each once, sorted. It is the one sanitizer: what
// leaves the process (an Issue comment, a manifest, a stored error, a ledger)
// passes through it or is refused by RefuseIfSensitive.
func Sanitize(s string) (string, []Category) {
	out, found := sanitize(s)
	cats := make([]Category, 0, len(found))
	for c := range found {
		cats = append(cats, c)
	}
	sort.Slice(cats, func(i, j int) bool { return cats[i] < cats[j] })
	return out, cats
}

// Redact is Sanitize with a count: the masked text and how many values were
// masked. Text that leaves the process passes through it first.
func Redact(s string) (string, int) {
	out, found := sanitize(s)
	n := 0
	for _, c := range found {
		n += c
	}
	return out, n
}

// RefuseIfSensitive is for the callers that must not send anything sensitive at
// all, masked or not: it returns ErrUnsanitized naming the categories found, never
// the values.
func RefuseIfSensitive(s string) error {
	_, cats := Sanitize(s)
	if len(cats) == 0 {
		return nil
	}
	names := make([]string, len(cats))
	for i, c := range cats {
		names[i] = string(c)
	}
	return fmt.Errorf("%w: found %s", ErrUnsanitized, strings.Join(names, ", "))
}

// wordShare is the share of s that reads as words: an optional capital followed by
// two or more lower-case letters. A path or an identifier made of words is mostly
// that; random text of this alphabet is about a third (99% of 40-character random
// runs sit under 0.65), a title-case path 0.86 and up.
func wordShare(s string) float64 {
	if s == "" {
		return 0
	}
	inWord := 0
	for i := 0; i < len(s); {
		j := i
		if s[j] >= 'A' && s[j] <= 'Z' {
			j++
		}
		k := j
		for k < len(s) && s[k] >= 'a' && s[k] <= 'z' {
			k++
		}
		if k-j >= 2 {
			inWord += k - i
			i = k
			continue
		}
		i++
	}
	return float64(inWord) / float64(len(s))
}

// maxWordShare is where a run stops being a secret and becomes a path of words.
const maxWordShare = 0.80

// looksCutSecret reports a run of [A-Za-z0-9/+] that is a secret cut into short
// pieces by '/' or '+'. A path is made of the same characters, so it is judged by
// what the text is and not by where it stands: it must hold a '/' or '+', mix upper
// case, lower case and digits, read as words for less than maxWordShare of it, and
// change character class like random text does once the '/' and '+' are taken out.
// What stands beside it decides nothing, because a secret is written after a
// sentence's last word and a full stop, after "key-" or "key_", inside a URL path
// and behind a leading or trailing '/'. A bare hex hash holds no '/' or '+' and is
// not touched.
func looksCutSecret(run string) bool {
	if !strings.ContainsAny(run, "/+") {
		return false
	}
	body := strings.Map(func(r rune) rune {
		if r == '/' || r == '+' {
			return -1
		}
		return r
	}, run)
	var lower, upper, digit bool
	pairs, flips := 0, 0
	for i := 0; i < len(body); i++ {
		c := classOf(body[i])
		switch c {
		case 'l':
			lower = true
		case 'u':
			upper = true
		case 'd':
			digit = true
		}
		if i > 0 {
			pairs++
			if c != classOf(body[i-1]) {
				flips++
			}
		}
	}
	if !lower || !upper || !digit || wordShare(body) >= maxWordShare || pairs == 0 {
		return false
	}
	return float64(flips)/float64(pairs) >= opaqueTransitionRatio
}

// tempRoot is the per-user temp root of macOS, /var/folders/xx/<random>/T/ (also C/
// or 0/, also behind /private): a directory named at random on purpose, printed by
// every test and build tool, holding no secret. Only this shape is exempt; whatever
// follows it is judged like any other text.
var tempRoot = regexp.MustCompile(`^(?:/private)?/var/folders/[a-z0-9]{2}/[a-z0-9_]{20,40}/[A-Za-z0-9](?:/|$)`)

// maxSuffixTries bounds the '/' boundaries tried in one run, so a very long path of
// many segments costs a bounded number of judgements.
const maxSuffixTries = 32

// cutSecretStart returns the index in run from which the text is a cut secret, or -1.
// The run is judged whole and then from after each '/', so a secret written behind a
// path of words (whose own share of words would otherwise lift the whole run above
// maxWordShare) is found; a macOS temp root at the start is skipped. The earliest
// start wins, and what precedes it stays.
//
// Known limit: a secret that is itself mostly readable words and digits (the
// documentation example keys are such samples) can still pass when it stands in a
// path of words; random secrets do not.
func cutSecretStart(run string) int {
	i := 0
	if loc := tempRoot.FindStringIndex(run); loc != nil {
		i = loc[1]
	}
	for try := 0; try < maxSuffixTries && len(run)-i >= 40; try++ {
		if looksCutSecret(run[i:]) {
			return i
		}
		j := strings.IndexByte(run[i:], '/')
		if j < 0 {
			break
		}
		i += j + 1
	}
	return -1
}

// maskCutSecrets masks the runs looksCutSecret accepts and counts them as opaque
// tokens.
func maskCutSecrets(s string, found map[Category]int) string {
	idx := cutRun.FindAllStringIndex(s, -1)
	if len(idx) == 0 {
		return s
	}
	var b strings.Builder
	last := 0
	for _, m := range idx {
		start, end := m[0], m[1]
		for end < len(s) && s[end] == '=' && end-m[1] < 2 { // base64 padding belongs to the token
			end++
		}
		k := cutSecretStart(s[m[0]:m[1]])
		if k < 0 {
			continue
		}
		start += k
		b.WriteString(s[last:start])
		b.WriteString(redacted)
		last = end
		found[CategoryOpaqueToken]++
	}
	b.WriteString(s[last:])
	return b.String()
}

// javaFrame is one stack-frame line of a Java or Node trace: "at fn (file:line:col)",
// "at Class.method(File.java:12)", "at /path/file.js:10:15" or "... 2 more". The shapes
// need a file position or a known marker, so indented prose that starts with "at" is
// not one.
const javaFrame = `(?:[ \t]+(?:at (?:[^\n(]+ \([^)\n]*[:.<][^)\n]*\)|[^\s(]+\([^)\n]*(?:[:.]|Native Method|Unknown Source)[^)\n]*\)|[^\s(]+:\d+(?::\d+)?)|\.\.\. \d+ (?:more|common frames omitted))[^\n]*(?:\n|$))`

// javaNodeStack is an optional error line, frame lines, and "Caused by:" groups.
var javaNodeStack = regexp.MustCompile(`(?m)(?:^[^\n]*(?:Error|Exception)[^\n]*\n)?` + javaFrame + `+(?:^Caused by: [^\n]*\n` + javaFrame + `*)*`)

// keepNewline masks a matched block and keeps the line break that ended it.
func keepNewline(g []string) (string, bool) {
	if strings.HasSuffix(g[0], "\n") {
		return redacted + "\n", true
	}
	return redacted, true
}

// A Korean or +82 telephone number. The left and right context characters are
// captured and must not be ASCII letters or digits, so a number inside a longer digit
// run (an id, a timestamp) and one inside a hex digest are left alone;
// the area part is one of the real prefixes, so an arbitrary ten-digit number is not
// a phone number.
var (
	phoneRe = regexp.MustCompile(`(^|[^0-9A-Za-z])((?:\+?82[-. ]?(?:10|1[16-9]|2|3[1-3]|4[1-4]|5[1-5]|6[1-4]|70|80)|0(?:1[016789]|2|3[1-3]|4[1-4]|5[1-5]|6[1-4]|70|80)\)?)[-. ]?\d{3,4}[-. ]?\d{4})([^0-9A-Za-z]|$)`)
	nameRe  = regexp.MustCompile(`(^|[^가-힣])([가-힣]{2,4})((?:\s*(?:대리|과장|차장|부장|팀장|실장|본부장|센터장|사원|주임|대표|이사|사장|교수|박사|선생|매니저|연구원|위원|의원))?(?:님|씨)?[\s:,(]*)\x00PHONE\x00`)
)

// notNames are words that label a number and are not a person: a name rule that
// took them would mask the label and report a person that is not there.
// Debt, not handled: a label or title this list does not name, placed straight before a
// number, is read as a name and masked (harmless, and reported as a person-name).
var notNames = map[string]bool{
	"전화": true, "연락": true, "연락처": true, "휴대폰": true, "핸드폰": true, "전화번호": true,
	"번호": true, "문의": true, "대표": true, "팩스": true, "모바일": true, "고객센터": true, "본사": true,
	"지사": true, "지점": true, "상담": true, "상담센터": true, "콜센터": true, "대표번호": true,
	"문의처": true, "담당": true, "담당자": true, "직통": true, "사무실": true, "회사": true, "자택": true,
	"비상연락망": true, "긴급연락": true, "휴대전화": true, "유선": true, "무선": true, "대리": true,
	"과장": true, "차장": true, "부장": true, "팀장": true, "실장": true, "본부장": true, "센터장": true,
	"사원": true, "주임": true, "이사": true, "사장": true, "교수": true, "박사": true, "매니저": true,
	"연구원": true, "위원": true, "의원": true,
}

// maskContacts masks telephone numbers, and a Korean name written directly before
// one. A name alone cannot be told from any other Korean word, so only a name beside
// a number is touched.
func maskContacts(s string, found map[Category]int) string {
	const ph = "\x00PHONE\x00"
	for pass := 0; pass < 3; pass++ { // a shared separator hides the next number from one pass
		next := phoneRe.ReplaceAllString(s, "${1}"+ph+"${3}")
		if next == s {
			break
		}
		s = next
	}
	phones := strings.Count(s, ph)
	if phones == 0 {
		return s
	}
	found[CategoryPhone] += phones
	s = nameRe.ReplaceAllStringFunc(s, func(m string) string {
		g := nameRe.FindStringSubmatch(m)
		if notNames[g[2]] {
			return m
		}
		found[CategoryPersonName]++
		return g[1] + ph
	})
	return strings.ReplaceAll(s, ph, redacted)
}
