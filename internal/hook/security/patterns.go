package security

// Guardian pattern table — the single-source, language-neutral dangerous-pattern
// catalogue consumed by the in-session 3-layer security guardian (Layer 1
// PostToolUse buffer scan, Layer 2 Stop turn-diff scan, Layer 3 opt-in commit
// cross-file review). SPEC-SEC-GUARDIAN-001.
//
// Design invariant (REQ-SG-011 / REQ-SG-053): the table is organized by
// vulnerability CLASS, never by programming language. Each pattern is a
// token-shaped, language-portable regex (a `yaml.load(` or `.innerHTML =`
// idiom reads the same across the 16 supported languages), so NO single
// language is treated as PRIMARY. `Langs` narrows a class ONLY when a pattern
// is genuinely language-specific; an empty `Langs` means the class applies
// across all 16 languages. This file is the sole pattern source — patterns are
// never scattered into the shell wrappers or duplicated per language.
//
// This table is DISTINCT from the pre-existing ast-grep `getDefaultOWASPRules`
// (used by the on-demand ast-grep scanner). The guardian table is regex-only,
// in-process, and advisory-first — it never shells out and never blocks.

import "regexp"

// VulnSeverity ranks a guardian vulnerability class. It is distinct from the
// ast-grep `Severity` type (error/warning/info/hint) used elsewhere in this
// package: the guardian uses the OWASP-style critical/high/medium/low scale.
type VulnSeverity string

const (
	// SevCritical is an immediately-exploitable, high-impact class (e.g. a
	// hardcoded private key).
	SevCritical VulnSeverity = "critical"
	// SevHigh is a serious vulnerability class (e.g. unsafe deserialization).
	SevHigh VulnSeverity = "high"
	// SevMedium is a review-worthy class where a false positive is cheap
	// (e.g. weak crypto, insecure randomness).
	SevMedium VulnSeverity = "medium"
	// SevLow is an informational class.
	SevLow VulnSeverity = "low"
)

// VulnClass groups dangerous patterns by vulnerability class (REQ-SG-011).
// A class applies across all 16 supported languages when Langs is empty.
type VulnClass struct {
	// Name is the vulnerability class identifier (e.g. "unsafe-deserialization").
	Name string
	// Severity is the class severity on the critical/high/medium/low scale.
	Severity VulnSeverity
	// Description is a one-line human explanation surfaced in advisory findings.
	Description string
	// Patterns are the language-agnostic regexes for the class.
	Patterns []*regexp.Regexp
	// Langs narrows the class to specific languages; empty means all 16.
	Langs []string
}

// GuardianFinding is a single advisory finding produced by the guardian scanner.
type GuardianFinding struct {
	Class    string       `json:"class"`
	Severity VulnSeverity `json:"severity"`
	Message  string       `json:"message"`
	Line     int          `json:"line"`
	Match    string       `json:"match"`
}

// mp is a small helper that compiles a case-sensitive regex at package-init time.
// A malformed pattern is a programmer error and panics via MustCompile.
func mp(expr string) *regexp.Regexp { return regexp.MustCompile(expr) }

// guardianClasses is the single-source pattern table (28 patterns across 10
// classes). It is a package-level var (compiled once) rather than a function
// so the compiled regexes are shared across all three layers.
//
// @MX:ANCHOR: [AUTO] guardianClasses is the single-source dangerous-pattern table consumed by all 3 guardian layers.
// @MX:REASON: [AUTO] fan_in >= 3 — ScanBuffer (Layer 1), ScanDiff (Layer 2), CrossFileScan (Layer 3) all iterate this table; changing its shape changes every layer.
var guardianClasses = []VulnClass{
	{
		Name:        "unsafe-deserialization",
		Severity:    SevHigh,
		Description: "Deserializing untrusted data can execute arbitrary code",
		Patterns: []*regexp.Regexp{
			mp(`yaml\.load\s*\(`),                 // PyYAML without SafeLoader
			mp(`pickle\.loads?\s*\(`),             // Python pickle
			mp(`torch\.load\s*\(`),                // PyTorch (ML model poisoning)
			mp(`Marshal\.load\s*\(`),              // Ruby Marshal
			mp(`new\s+ObjectInputStream\s*\(`),    // Java native deserialization
			mp(`\bunserialize\s*\(`),              // PHP unserialize
		},
	},
	{
		Name:        "dom-injection-xss",
		Severity:    SevHigh,
		Description: "Assigning untrusted data to a DOM sink enables XSS",
		Patterns: []*regexp.Regexp{
			mp(`\.innerHTML\s*=`),          // raw innerHTML assignment
			mp(`dangerouslySetInnerHTML`),  // React escape hatch
			mp(`document\.write\s*\(`),     // document.write sink
			mp(`v-html\s*=`),               // Vue raw HTML binding
		},
	},
	{
		Name:        "hardcoded-secret",
		Severity:    SevCritical,
		Description: "Hardcoded credential or private key committed to source",
		Patterns: []*regexp.Regexp{
			// Assignment operator covers = / := (Go) / <- (R) across the 16 languages.
			mp(`(?i)api[_-]?key\s*(:?=|<-)\s*["'][^"']+["']`), // api_key = "..." / apiKey := "..."
			mp(`(?i)password\s*(:?=|<-)\s*["'][^"']+["']`),    // password = "..."
			mp(`-----BEGIN [A-Z ]*PRIVATE KEY-----`),           // PEM private key block
			mp(`\bAKIA[0-9A-Z]{16}\b`),                         // AWS access key id
		},
	},
	{
		Name:        "code-injection-eval",
		Severity:    SevHigh,
		Description: "Dynamic code evaluation of untrusted input",
		Patterns: []*regexp.Regexp{
			mp(`\beval\s*\(`),          // eval() across many languages
			mp(`new\s+Function\s*\(`),  // JS Function constructor
		},
	},
	{
		Name:        "sql-injection",
		Severity:    SevHigh,
		Description: "SQL built by string concatenation instead of parameters",
		Patterns: []*regexp.Regexp{
			mp(`(?i)["'\x60]\s*(SELECT|INSERT|UPDATE|DELETE|DROP)\b[^"'\x60]*["'\x60]\s*\+`), // "SELECT ..." +
			mp(`(?i)(execute|query)\s*\([^)]*\+[^)]*\)`),                                     // query("..." + x)
		},
	},
	{
		Name:        "command-injection",
		Severity:    SevHigh,
		Description: "Spawning a shell with unsanitized input",
		Patterns: []*regexp.Regexp{
			mp(`shell\s*=\s*True`),            // Python subprocess(shell=True)
			mp(`os\.system\s*\(`),             // os.system(cmd)
			mp(`child_process\.exec\s*\(`),    // Node child_process.exec
			mp(`["'\x60](sh|bash)\s+-c\b`),    // sh -c / bash -c string
		},
	},
	{
		Name:        "path-traversal",
		Severity:    SevMedium,
		Description: "Path traversal sequence in a file path (review for user input)",
		Patterns: []*regexp.Regexp{
			mp(`\.\./\.\./`), // ../../ traversal
		},
	},
	{
		Name:        "ssrf",
		Severity:    SevMedium,
		Description: "Outbound request with a concatenated URL (review for user input)",
		Patterns: []*regexp.Regexp{
			mp(`(?i)(requests\.(get|post)|urllib\.request\.urlopen|http\.Get)\s*\([^)]*\+`), // requests.get(url + x)
		},
	},
	{
		Name:        "weak-crypto",
		Severity:    SevMedium,
		Description: "Weak hash or cipher mode for security-sensitive data",
		Patterns: []*regexp.Regexp{
			mp(`(?i)\b(MD5|SHA1)\b`), // md5/sha1 for password/token
			mp(`(?i)\bECB\b`),        // ECB cipher mode
		},
	},
	{
		Name:        "insecure-random",
		Severity:    SevMedium,
		Description: "Non-cryptographic randomness used where a CSPRNG is required",
		Patterns: []*regexp.Regexp{
			mp(`Math\.random\s*\(\s*\)`), // JS Math.random()
			mp(`(?i)\brand\s*\(\s*\)`),   // C/PHP rand()
		},
	},
}

// GuardianPatterns returns the single-source vulnerability-class table. All three
// guardian layers consume this one table (REQ-SG-053).
func GuardianPatterns() []VulnClass { return guardianClasses }

// Cross-file heuristic markers (Layer 3, REQ-SG-031). These detect the classic
// multi-file IDOR / broken-authorization shape: a user-supplied object id flows
// to an object-access sink WITHOUT an authorization check anywhere in the
// changed + related file set. They are token-shaped and language-neutral.
var (
	// crossFileIDSourceMarkers detect a user-supplied object identifier source.
	crossFileIDSourceMarkers = []*regexp.Regexp{
		mp(`(?i)req(uest)?\.(params|query|args|body)\b`), // req.params.id / request.args
		mp(`(?i)params\[["':]?id`),                       // params[:id] / params["id"]
		mp(`(?i)request\.GET\b`),                          // Django request.GET
	}
	// crossFileObjectSinkMarkers detect an object-access-by-id sink.
	crossFileObjectSinkMarkers = []*regexp.Regexp{
		mp(`(?i)find(ById|One)?\s*\(`),        // findById( / findOne(
		mp(`(?i)WHERE\s+id\s*=`),              // WHERE id =
		mp(`(?i)get_object_or_404\s*\(`),      // Django get_object_or_404
		mp(`(?i)\.get\s*\(\s*(pk|id)\b`),      // Model.objects.get(id=...)
	}
	// crossFileAuthzMarkers detect an authorization / ownership check.
	crossFileAuthzMarkers = []*regexp.Regexp{
		mp(`(?i)(authorize|authorized|current_user|owner_id|user_id\s*==|has_permission|can_access|verify_owner)`),
	}
)
