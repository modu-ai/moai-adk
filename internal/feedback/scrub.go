// Package feedback owns the text transformations applied to a feedback report
// before it can leave the machine: environment-value masking, secret-pattern
// masking, home-path collapse, and the pre-mask vulnerability classification
// that decides whether the report may be submitted to a public channel at all.
//
// The package is deliberately free of CLI concerns. internal/cli wires stdin,
// stdout and exit codes around it; everything here is a pure function over
// strings so the transformations can be table-tested.
package feedback

import (
	"github.com/modu-ai/moai-adk/internal/hook"
)

// Verdict values carried by Result.Verdict.
const (
	VerdictOK      = "ok"
	VerdictBlocked = "blocked"
)

// Finding kinds. One kind per transformation stage of the pipeline.
const (
	KindEnv      = "env"
	KindSecret   = "secret"
	KindHomePath = "homepath"
)

// Finding locations. The confirmation gate needs to be able to say "one match
// was masked in the title" — a count without a location leaves the user
// approving a title they do not know was rewritten.
const (
	WhereTitle = "title"
	WhereBody  = "body"
)

// Input carries every piece of user text that will be submitted. The title is
// a separate argument of `gh issue create` and is free-form user text, so a
// scrubber that only sees the body leaks through the title.
type Input struct {
	Title string
	Body  string
}

// Finding reports what was masked, where, and how many times. It never carries
// a raw value: the mask log and the confirmation gate both render findings, so
// a value here would turn the control into a leak surface.
type Finding struct {
	Kind  string `json:"kind"`
	Where string `json:"where"`
	Count int    `json:"count"`
}

// Result is the scrubber's whole output. JSON field names are lower snake so
// the skill body can read them with jq.
type Result struct {
	Verdict  string    `json:"verdict"`
	Title    string    `json:"title"`
	Body     string    `json:"body"`
	Findings []Finding `json:"findings"`
	Reason   string    `json:"reason"`
}

// Options carries the collaborators a Scrub call needs. Every field has a
// working zero value, so Scrub(in, Options{}) is the production configuration.
type Options struct {
	// Policy supplies the sensitive-content pattern set. Nil resolves to
	// hook.DefaultSecurityPolicy(). Callers that want the user's
	// security.extra_sensitive_content_patterns extensions pass a policy that
	// has already been through MergeExtraPatterns — the extension contract is
	// inherited rather than re-implemented here.
	Policy *hook.SecurityPolicy

	// EnvScrubExtra lists additional environment-variable names whose values
	// are masked, mirroring security.sandbox.env_scrub_extra.
	EnvScrubExtra []string

	// Environ supplies the environment as "KEY=VALUE" pairs. Nil resolves to
	// os.Environ.
	Environ func() []string

	// Home overrides the home directory used for path collapse. Empty resolves
	// through paths.Home(), which is HOME-first.
	Home string

	// ProjectRoot is the directory whose .moai tree receives the mask log and
	// the retry queue. Empty resolves by walking up from the working directory
	// looking for the .moai marker. Consumed by the on-disk artefacts.
	ProjectRoot string
}

// Scrub classifies the raw input, then masks it.
//
// [HARD] The classification reads the RAW input, before any masking. Masking
// removes exactly the signals the classifier needs (a secret-pattern hit, a
// key-file path mention), so classifying the masked text turns "a dangerous
// report containing a credential" into "an ordinary report". That inversion is
// a silent false negative, which is why the ordering is fixed here rather than
// left to the caller.
//
// The transformation order is env values, then secret patterns, then home
// paths. Env values are an exact match and therefore the most precise
// attribution; running the regexes first would attribute a token-shaped env
// value to "secret" instead of "env". Home collapse runs last so the mask
// tokens produced by the earlier stages cannot interfere with path matching.
//
// The pipeline is idempotent: scrubbing an already-scrubbed text returns it
// unchanged. The retry queue re-scrubs what it dequeues, and the confirmation
// gate re-scrubs a user-edited body, so both depend on that property.
func Scrub(in Input, opt Options) (Result, error) {
	verdict, reason := classify(in)

	patterns := rewritePatterns(opt.Policy)
	envValues := envMaskValues(environOf(opt), opt.EnvScrubExtra)
	home := resolveHome(opt)

	findings := make([]Finding, 0, 6)
	title, titleFindings := transform(in.Title, WhereTitle, patterns, envValues, home)
	body, bodyFindings := transform(in.Body, WhereBody, patterns, envValues, home)
	findings = append(findings, titleFindings...)
	findings = append(findings, bodyFindings...)

	return Result{
		Verdict:  verdict,
		Title:    title,
		Body:     body,
		Findings: findings,
		Reason:   reason,
	}, nil
}

// transform applies the three masking stages to one field and reports what each
// stage matched. Findings are emitted in pipeline order so the output is
// deterministic.
func transform(s, where string, patterns []rewritePattern, envValues []string, home string) (string, []Finding) {
	var findings []Finding

	out, envCount := maskEnvValues(s, envValues)
	if envCount > 0 {
		findings = append(findings, Finding{Kind: KindEnv, Where: where, Count: envCount})
	}

	out, secretCount := maskSecrets(out, patterns)
	if secretCount > 0 {
		findings = append(findings, Finding{Kind: KindSecret, Where: where, Count: secretCount})
	}

	out, homeCount := collapseHome(out, home)
	if homeCount > 0 {
		findings = append(findings, Finding{Kind: KindHomePath, Where: where, Count: homeCount})
	}

	return out, findings
}

// @MX:TODO: [AUTO] placeholder classifier — the vulnerability signals live in classify.go
//
// classify is the pre-mask classification seam. It exists here so that the
// pipeline ordering above is fixed by the code that owns the pipeline, not by
// the classifier that will later fill it in. The signal set (secret-pattern
// hit, key-file path mention, vulnerability vocabulary) and the SECURITY.md
// routing message are a separate deliverable; until then every input that
// survives the scrub is submittable.
func classify(Input) (verdict, reason string) {
	return VerdictOK, ""
}
