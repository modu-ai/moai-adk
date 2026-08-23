package feedback

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	ghsecret "github.com/modu-ai/moai-adk/internal/github"
)

// Credential-shaped fixtures are assembled at run time from a prefix constant
// plus a deterministic dummy tail. No contiguous credential-looking literal is
// committed to source: the repository's PreToolUse guard rejects such content,
// and a committed fixture would itself be a leak surface.
const (
	ghTokenPrefix   = "ghp_"
	googleKeyPrefix = "AIza"
	awsKeyPrefix    = "AKIA"
	armorOpen       = "-----BEGIN"
	armorClose      = "-----END"
	keyArmorLabel   = " RSA PRIVATE KEY"
	armorDashes     = "-----"
)

// dummyTail returns n deterministic alphanumeric characters.
func dummyTail(n int) string {
	const alphabet = "abcdefghij0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	var b strings.Builder
	for i := 0; i < n; i++ {
		b.WriteByte(alphabet[i%len(alphabet)])
	}
	return b.String()
}

func fakeGitHubToken() string  { return ghTokenPrefix + dummyTail(36) }
func fakeGoogleAPIKey() string { return googleKeyPrefix + dummyTail(35) }

func fakePrivateKeyBlock(withTerminator bool) string {
	lines := []string{
		armorOpen + keyArmorLabel + armorDashes,
		dummyTail(44),
		dummyTail(40),
	}
	if withTerminator {
		lines = append(lines, armorClose+keyArmorLabel+armorDashes)
	}
	return strings.Join(lines, "\n")
}

// testOptions isolates a Scrub call from the ambient process environment and
// from the real home directory, so table cases stay deterministic.
func testOptions() Options {
	return Options{
		Environ: func() []string { return nil },
		Home:    "/nonexistent-home-for-tests",
	}
}

func findingFor(res Result, kind, where string) (Finding, bool) {
	for _, f := range res.Findings {
		if f.Kind == kind && f.Where == where {
			return f, true
		}
	}
	return Finding{}, false
}

// AC-F-005 — findings carry kind/where/count only, never a raw value.
func TestFindingsCarryNoRawValue(t *testing.T) {
	t.Parallel()

	token := fakeGitHubToken()
	res, err := Scrub(Input{Body: "the token is " + token + " ok"}, testOptions())
	if err != nil {
		t.Fatalf("Scrub returned error: %v", err)
	}
	if len(res.Findings) == 0 {
		t.Fatalf("expected at least one finding, got none")
	}

	rendered := fmt.Sprintf("%+v", res.Findings)
	for i := 0; i+8 <= len(token); i++ {
		if strings.Contains(rendered, token[i:i+8]) {
			t.Fatalf("findings leak an 8-char substring of the raw token at offset %d: %s", i, rendered)
		}
	}
}

// AC-F-006 — GitHub token masked in body AND in title, with Where attribution.
func TestScrubMasksGitHubToken(t *testing.T) {
	t.Parallel()

	token := fakeGitHubToken()

	t.Run("body", func(t *testing.T) {
		t.Parallel()
		res, err := Scrub(Input{Title: "plain title", Body: "token is " + token + " here"}, testOptions())
		if err != nil {
			t.Fatalf("Scrub returned error: %v", err)
		}
		if strings.Contains(res.Body, token) {
			t.Fatalf("raw token survived in body: %q", res.Body)
		}
		f, ok := findingFor(res, KindSecret, WhereBody)
		if !ok {
			t.Fatalf("expected a %q finding in %q, got %+v", KindSecret, WhereBody, res.Findings)
		}
		if f.Count != 1 {
			t.Fatalf("expected Count 1, got %d", f.Count)
		}
	})

	t.Run("title", func(t *testing.T) {
		t.Parallel()
		res, err := Scrub(Input{Title: "crash with " + token, Body: "plain body"}, testOptions())
		if err != nil {
			t.Fatalf("Scrub returned error: %v", err)
		}
		if strings.Contains(res.Title, token) {
			t.Fatalf("raw token survived in title: %q", res.Title)
		}
		if _, ok := findingFor(res, KindSecret, WhereTitle); !ok {
			t.Fatalf("expected a %q finding in %q, got %+v", KindSecret, WhereTitle, res.Findings)
		}
	})
}

// AC-F-007 — the AIza pattern exists only via the union, so a pass proves the
// union is actually applied.
func TestScrubMasksGoogleAPIKey(t *testing.T) {
	t.Parallel()

	key := fakeGoogleAPIKey()
	res, err := Scrub(Input{Body: "key " + key}, testOptions())
	if err != nil {
		t.Fatalf("Scrub returned error: %v", err)
	}
	if strings.Contains(res.Body, key) {
		t.Fatalf("raw Google API key survived: %q", res.Body)
	}
	if _, ok := findingFor(res, KindSecret, WhereBody); !ok {
		t.Fatalf("expected a %q finding, got %+v", KindSecret, res.Findings)
	}
}

// AC-F-008 — benign prose is untouched on all three axes, and the rewrite path
// does not inherit the detector's case-insensitive over-reach.
func TestScrubBenignBodyUntouchedAndAllowed(t *testing.T) {
	t.Parallel()

	t.Run("benign prose", func(t *testing.T) {
		t.Parallel()
		body := "the " + ghTokenPrefix + " prefix is how GitHub tokens start. moai init 실행 시 마법사가 두 번 뜹니다."
		res, err := Scrub(Input{Body: body}, testOptions())
		if err != nil {
			t.Fatalf("Scrub returned error: %v", err)
		}
		if res.Body != body {
			t.Fatalf("benign body was rewritten:\n got: %q\nwant: %q", res.Body, body)
		}
		if len(res.Findings) != 0 {
			t.Fatalf("expected no findings, got %+v", res.Findings)
		}
		if res.Verdict != VerdictOK {
			t.Fatalf("expected verdict %q, got %q", VerdictOK, res.Verdict)
		}
	})

	t.Run("lowercase prose run", func(t *testing.T) {
		t.Parallel()
		// The detector compiles its patterns with (?i); a rewriter that inherits
		// that flag eats this lowercase run of ordinary prose.
		body := "sequence " + strings.ToLower(awsKeyPrefix) + "abcdefghijklmnop follows"
		res, err := Scrub(Input{Body: body}, testOptions())
		if err != nil {
			t.Fatalf("Scrub returned error: %v", err)
		}
		if res.Body != body {
			t.Fatalf("case-insensitive over-masking:\n got: %q\nwant: %q", res.Body, body)
		}
	})
}

// Falsification guard for the AC-F-008 lowercase case: it only observes
// something if the detector's own form WOULD have matched that prose. The
// subtests below show both sides of that asymmetry, so a future reader can see
// the case-sensitivity requirement is load-bearing rather than decorative.
func TestRewritePatternsAreCaseSensitive(t *testing.T) {
	t.Parallel()

	body := "sequence " + strings.ToLower(awsKeyPrefix) + "abcdefghijklmnop follows"

	patterns := rewritePatterns(nil)
	if len(patterns) == 0 {
		t.Fatalf("expected a non-empty rewrite pattern set")
	}

	var awsSource string
	for _, p := range patterns {
		if strings.HasPrefix(p.re.String(), caseInsensitiveFlag) {
			t.Fatalf("rewrite pattern kept the case-insensitive flag: %q", p.re.String())
		}
		if p.re.MatchString(body) {
			t.Fatalf("rewrite pattern %q matched lowercase prose", p.re.String())
		}
		if strings.HasPrefix(p.re.String(), awsKeyPrefix) {
			awsSource = p.re.String()
		}
	}

	if awsSource == "" {
		t.Fatalf("expected the policy to carry an %s-prefixed pattern", awsKeyPrefix)
	}
	detector := regexp.MustCompile(caseInsensitiveFlag + awsSource)
	if !detector.MatchString(body) {
		t.Fatalf("the detector form does not match the prose, so this guard observes nothing")
	}
}

// AC-F-009 — the masked span equals the adopted masker's return value exactly.
func TestMaskOutputShapeMatchesExistingMasker(t *testing.T) {
	t.Parallel()

	token := fakeGitHubToken()
	res, err := Scrub(Input{Body: "prefix " + token + " suffix"}, testOptions())
	if err != nil {
		t.Fatalf("Scrub returned error: %v", err)
	}
	want := "prefix " + ghsecret.MaskSecret(token) + " suffix"
	if res.Body != want {
		t.Fatalf("mask shape mismatch:\n got: %q\nwant: %q", res.Body, want)
	}
}

// AC-F-010 — home prefix collapse follows the paths.Home() HOME-first contract.
func TestScrubCollapsesHomePath(t *testing.T) {
	// No t.Parallel: t.Setenv is used in the subtests below.
	opt := func() Options {
		return Options{Environ: func() []string { return nil }}
	}

	t.Run("first home", func(t *testing.T) {
		home := t.TempDir()
		t.Setenv("HOME", home)
		res, err := Scrub(Input{Body: "see " + home + "/proj/main.go"}, opt())
		if err != nil {
			t.Fatalf("Scrub returned error: %v", err)
		}
		if !strings.Contains(res.Body, "~/proj/main.go") {
			t.Fatalf("home prefix not collapsed: %q", res.Body)
		}
		if strings.Contains(res.Body, home+"/") {
			t.Fatalf("absolute home prefix survived: %q", res.Body)
		}
		if _, ok := findingFor(res, KindHomePath, WhereBody); !ok {
			t.Fatalf("expected a %q finding, got %+v", KindHomePath, res.Findings)
		}
	})

	t.Run("second home", func(t *testing.T) {
		home := t.TempDir()
		t.Setenv("HOME", home)
		res, err := Scrub(Input{Body: "see " + home + "/x"}, opt())
		if err != nil {
			t.Fatalf("Scrub returned error: %v", err)
		}
		if !strings.Contains(res.Body, "~/x") || strings.Contains(res.Body, home+"/") {
			t.Fatalf("collapse did not follow the new HOME: %q", res.Body)
		}
	})
}

// AC-F-011 — environment values are masked by name vocabulary, extensible via
// security.sandbox.env_scrub_extra.
func TestScrubMasksEnvValues(t *testing.T) {
	const value = "abcd1234efgh"

	t.Run("default deny list", func(t *testing.T) {
		t.Parallel()
		opt := Options{
			Environ: func() []string { return []string{"GITHUB_TOKEN=" + value} },
			Home:    "/nonexistent-home-for-tests",
		}
		res, err := Scrub(Input{Body: "the value " + value + " appeared"}, opt)
		if err != nil {
			t.Fatalf("Scrub returned error: %v", err)
		}
		if strings.Contains(res.Body, value) {
			t.Fatalf("env value survived: %q", res.Body)
		}
		if _, ok := findingFor(res, KindEnv, WhereBody); !ok {
			t.Fatalf("expected a %q finding, got %+v", KindEnv, res.Findings)
		}
	})

	t.Run("env_scrub_extra", func(t *testing.T) {
		t.Parallel()
		opt := Options{
			Environ:       func() []string { return []string{"MY_CUSTOM_TOKEN=" + value} },
			EnvScrubExtra: []string{"MY_CUSTOM_TOKEN"},
			Home:          "/nonexistent-home-for-tests",
		}
		res, err := Scrub(Input{Body: "the value " + value + " appeared"}, opt)
		if err != nil {
			t.Fatalf("Scrub returned error: %v", err)
		}
		if strings.Contains(res.Body, value) {
			t.Fatalf("extra env value survived: %q", res.Body)
		}
	})

	t.Run("process environment", func(t *testing.T) {
		// No t.Parallel: t.Setenv is used, and the default Environ source is
		// os.Environ — the production path.
		t.Setenv("GITHUB_TOKEN", value)
		res, err := Scrub(Input{Body: "the value " + value + " appeared"}, Options{Home: "/nonexistent-home-for-tests"})
		if err != nil {
			t.Fatalf("Scrub returned error: %v", err)
		}
		if strings.Contains(res.Body, value) {
			t.Fatalf("process env value survived: %q", res.Body)
		}
	})
}

// AC-F-014 — the pipeline is idempotent; the retry queue and the confirmation
// gate's edit-and-resubmit option both depend on it.
func TestScrubIsIdempotent(t *testing.T) {
	t.Parallel()

	const value = "abcd1234efgh"
	opt := Options{
		Environ: func() []string { return []string{"GITHUB_TOKEN=" + value} },
		Home:    "/home/tester",
	}
	body := strings.Join([]string{
		"token " + fakeGitHubToken(),
		"key " + fakeGoogleAPIKey(),
		"env " + value,
		"path /home/tester/proj/main.go",
		fakePrivateKeyBlock(true),
	}, "\n")

	first, err := Scrub(Input{Title: "title " + fakeGitHubToken(), Body: body}, opt)
	if err != nil {
		t.Fatalf("first Scrub returned error: %v", err)
	}
	second, err := Scrub(Input{Title: first.Title, Body: first.Body}, opt)
	if err != nil {
		t.Fatalf("second Scrub returned error: %v", err)
	}
	if second.Body != first.Body {
		t.Fatalf("body is not idempotent:\nfirst:  %q\nsecond: %q", first.Body, second.Body)
	}
	if second.Title != first.Title {
		t.Fatalf("title is not idempotent:\nfirst:  %q\nsecond: %q", first.Title, second.Title)
	}
}

// AC-F-024 — a private-key block is masked from its header through the block
// terminator, and to end-of-input when the terminator is absent.
func TestScrubMasksPrivateKeyBlockEntirely(t *testing.T) {
	t.Parallel()

	t.Run("with terminator", func(t *testing.T) {
		t.Parallel()
		block := fakePrivateKeyBlock(true)
		res, err := Scrub(Input{Body: "here it is:\n" + block + "\nthat was the key"}, testOptions())
		if err != nil {
			t.Fatalf("Scrub returned error: %v", err)
		}
		for _, line := range strings.Split(block, "\n") {
			if strings.Contains(res.Body, line) {
				t.Fatalf("key block line survived: %q in %q", line, res.Body)
			}
		}
		if !strings.Contains(res.Body, "that was the key") {
			t.Fatalf("masking swallowed the trailing prose: %q", res.Body)
		}
	})

	t.Run("truncated block", func(t *testing.T) {
		t.Parallel()
		block := fakePrivateKeyBlock(false)
		res, err := Scrub(Input{Body: "here it is:\n" + block}, testOptions())
		if err != nil {
			t.Fatalf("Scrub returned error: %v", err)
		}
		for _, line := range strings.Split(block, "\n") {
			if strings.Contains(res.Body, line) {
				t.Fatalf("truncated key block line survived: %q in %q", line, res.Body)
			}
		}
	})
}
