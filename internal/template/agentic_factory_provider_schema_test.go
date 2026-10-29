// agentic_factory_provider_schema_test.go: guards for the DRAFT provider
// configuration schema fixture (testdata/agentic_factory). The schema is a
// fixture, not a shipped section, because the existing "new section" pattern
// edits shared registries (see the fixture header). These tests pin the key
// names (the downstream reader consumes them), the rejection style, the
// UNKNOWN rule and the "documented defaults stay uncollated" marker.
//
// afValidate / afResolve below are test-local mirrors of the rules the schema
// states. They are NOT the production reader: parity with the committed reader
// is checked by name and type only, and recorded as a gap.
//
// Sentinels on failure: AF_SCHEMA_KEYS / AF_SCHEMA_REJECT / AF_SCHEMA_UNKNOWN /
// AF_SCHEMA_SOURCE.
package template_test

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

const (
	afSchemaPath  = "testdata/agentic_factory/provider-config-schema-draft.yaml"
	afExamplePath = "testdata/agentic_factory/provider-config-example.yaml"
	afFetchedDate = "2026-10-29"
)

type afSchemaKey struct {
	Key            string   `yaml:"key"`
	Provider       string   `yaml:"provider"`
	Type           string   `yaml:"type"`
	Min            *int     `yaml:"min"`
	Allowed        []string `yaml:"allowed"`
	Default        any      `yaml:"default"`
	Unit           string   `yaml:"unit"`
	AppliesTo      string   `yaml:"applies_to"`
	OnOutOfRange   string   `yaml:"on_out_of_range"`
	OnInvalid      string   `yaml:"on_invalid"`
	WhenUnresolved string   `yaml:"when_unresolved"`
	SourceURLs     []string `yaml:"source_urls"`
	Fetched        string   `yaml:"fetched"`
	Status         string   `yaml:"status"`
	Provenance     string   `yaml:"provenance"`
	Note           string   `yaml:"note"`
}

type afSchema struct {
	SchemaVersion   string        `yaml:"schema_version"`
	SchemaStatus    string        `yaml:"schema_status"`
	Fetched         string        `yaml:"fetched"`
	UnknownRule     string        `yaml:"unknown_rule"`
	ResolutionOrder []string      `yaml:"resolution_order"`
	ExistingKeys    []afSchemaKey `yaml:"existing_reader_keys"`
	Keys            []afSchemaKey `yaml:"keys"`
}

func afLoadSchema(t *testing.T) afSchema {
	t.Helper()
	raw, err := os.ReadFile(filepath.FromSlash(afSchemaPath))
	if err != nil {
		t.Fatalf("AF_SCHEMA_KEYS: read schema fixture: %v", err)
	}
	var s afSchema
	if err := yaml.Unmarshal(raw, &s); err != nil {
		t.Fatalf("AF_SCHEMA_KEYS: schema fixture does not parse: %v", err)
	}
	return s
}

// afExpectedRateKeys pins the stable key names and their documented defaults
// (nil = no documented value). Downstream code consumes these names, so a
// rename or a changed default must be a deliberate edit of this table.
var afExpectedRateKeys = []struct {
	Suffix  string
	Default any
}{
	{"github.rate_limit.primary_unauthenticated_per_hour", 60},
	{"github.rate_limit.primary_authenticated_user_per_hour", 5000},
	{"github.rate_limit.primary_actions_token_per_repo_per_hour", 1000},
	{"github.rate_limit.secondary_concurrent_requests_ceiling", 100},
	{"github.rate_limit.recommended_concurrent_requests", 1},
	{"github.rate_limit.secondary_rest_points_per_minute", 900},
	{"github.rate_limit.secondary_cpu_seconds_per_60s", 90},
	{"github.rate_limit.secondary_content_requests_per_minute", 80},
	{"github.rate_limit.secondary_content_requests_per_hour", 500},
	{"github.rate_limit.min_seconds_between_mutating_requests", 1},
	{"github.rate_limit.secondary_unclassified_min_wait_seconds", 60},
	{"github.rate_limit.retry_backoff_multiplier", nil},
	{"github.rate_limit.retry_max_attempts", nil},
	{"github.rate_limit.poll_interval_seconds", nil},
	{"gitlab.rate_limit.authenticated_api_per_user_per_minute", 2000},
	{"gitlab.rate_limit.issue_creation_per_minute", 200},
	{"gitlab.rate_limit.note_creation_per_minute", 60},
	{"gitlab.rate_limit.unauthenticated_per_ip_per_minute", 500},
	{"gitlab.rate_limit.min_seconds_between_mutating_requests", nil},
	{"gitlab.rate_limit.retry_backoff_multiplier", nil},
	{"gitlab.rate_limit.retry_max_attempts", nil},
	{"gitlab.rate_limit.poll_interval_seconds", nil},
}

const afKeyPrefix = "workflow.agentic_factory."

// afValidate mirrors the schema's rejection rule for one value: out-of-set,
// out-of-range and wrong-typed values are errors; nothing is clamped or
// coerced. A nil value (unset) is valid.
func afValidate(k afSchemaKey, v any) error {
	if v == nil {
		return nil
	}
	switch k.Type {
	case "string":
		s, ok := v.(string)
		if !ok {
			return fmt.Errorf("%s: must be a string, got %T", k.Key, v)
		}
		for _, a := range k.Allowed {
			if s == a {
				return nil
			}
		}
		return fmt.Errorf("%s: must be one of: %s (or unset), got %q", k.Key, strings.Join(k.Allowed, ", "), s)
	case "boolean":
		if _, ok := v.(bool); !ok {
			return fmt.Errorf("%s: must be a bool, got %T", k.Key, v)
		}
		return nil
	case "integer":
		n, ok := v.(int)
		if !ok {
			return fmt.Errorf("%s: must be an integer, got %T", k.Key, v)
		}
		if k.Min != nil && n < *k.Min {
			return fmt.Errorf("%s: must be an integer >= %d (or unset), got %d", k.Key, *k.Min, n)
		}
		return nil
	}
	return fmt.Errorf("%s: unknown schema type %q", k.Key, k.Type)
}

// afResolve mirrors the resolution order: response header, configured value,
// schema default, UNKNOWN. UNKNOWN is reported as (nil, "UNKNOWN") so a caller
// cannot mistake it for a number.
func afResolve(k afSchemaKey, header, configured any) (value any, source string) {
	switch {
	case header != nil:
		return header, "response_header"
	case configured != nil:
		return configured, "config_value"
	case k.Default != nil:
		return k.Default, "schema_default"
	}
	return nil, "UNKNOWN"
}

func afIndexKeys(s afSchema) map[string]afSchemaKey {
	m := map[string]afSchemaKey{}
	for _, k := range s.Keys {
		m[k.Key] = k
	}
	return m
}

// TestAgenticFactorySchemaKeysStable pins the key set, types, defaults and
// document-level rules.
func TestAgenticFactorySchemaKeysStable(t *testing.T) {
	t.Parallel()
	s := afLoadSchema(t)
	if s.SchemaVersion != "draft-1" || s.SchemaStatus != "DRAFT" {
		t.Errorf("AF_SCHEMA_KEYS: schema must be marked draft-1 / DRAFT, got %q / %q", s.SchemaVersion, s.SchemaStatus)
	}
	if want := []string{"response_header", "config_value", "schema_default", "UNKNOWN"}; strings.Join(s.ResolutionOrder, ">") != strings.Join(want, ">") {
		t.Errorf("AF_SCHEMA_KEYS: resolution order = %v, want %v", s.ResolutionOrder, want)
	}

	idx := afIndexKeys(s)
	if len(idx) != len(s.Keys) {
		t.Errorf("AF_SCHEMA_KEYS: duplicate key in schema (%d unique of %d)", len(idx), len(s.Keys))
	}

	// the rate-limit keys: exact set, exact defaults
	rate := 0
	for _, k := range s.Keys {
		if strings.Contains(k.Key, ".rate_limit.") {
			rate++
		}
	}
	if rate != len(afExpectedRateKeys) {
		t.Errorf("AF_SCHEMA_KEYS: schema has %d rate-limit keys, table pins %d", rate, len(afExpectedRateKeys))
	}
	keyRe := regexp.MustCompile(`^workflow\.agentic_factory\.(github|gitlab)\.rate_limit\.[a-z0-9_]+$`)
	for _, e := range afExpectedRateKeys {
		full := afKeyPrefix + e.Suffix
		k, ok := idx[full]
		if !ok {
			t.Errorf("AF_SCHEMA_KEYS: key %s missing", full)
			continue
		}
		if !keyRe.MatchString(k.Key) {
			t.Errorf("AF_SCHEMA_KEYS: key %s does not match the stable naming pattern", k.Key)
		}
		if got := strings.Split(k.Key, ".")[2]; got != k.Provider {
			t.Errorf("AF_SCHEMA_KEYS: key %s names provider %q but provider field is %q", k.Key, got, k.Provider)
		}
		if k.Type != "integer" {
			t.Errorf("AF_SCHEMA_KEYS: %s type = %q, want integer", k.Key, k.Type)
		}
		if k.Unit == "" {
			t.Errorf("AF_SCHEMA_KEYS: %s has no unit", k.Key)
		}
		if fmt.Sprint(k.Default) != fmt.Sprint(e.Default) {
			t.Errorf("AF_SCHEMA_KEYS: %s default = %v, table pins %v", k.Key, k.Default, e.Default)
		}
	}

	// the proposed provider key mirrors browser.selected's rejection style
	p, ok := idx[afKeyPrefix+"provider.selected"]
	if !ok {
		t.Fatalf("AF_SCHEMA_KEYS: proposed key provider.selected missing")
	}
	if p.Type != "string" || strings.Join(p.Allowed, ",") != "github,gitlab" || p.Default != nil || p.OnInvalid != "reject" || p.Status != "proposed" {
		t.Errorf("AF_SCHEMA_KEYS: provider.selected must be string [github,gitlab], no default, on_invalid reject, status proposed; got %+v", p)
	}

	// consistency with the committed reader's key names and types
	want := map[string]string{
		afKeyPrefix + "browser.selected":   "string",
		afKeyPrefix + "codex.enabled":      "boolean",
		afKeyPrefix + "pr_queue.wip_limit": "integer",
	}
	for _, k := range s.ExistingKeys {
		if want[k.Key] != k.Type {
			t.Errorf("AF_SCHEMA_KEYS: existing reader key %s typed %q, reader has %q", k.Key, k.Type, want[k.Key])
		}
		delete(want, k.Key)
		if k.OnInvalid != "reject" {
			t.Errorf("AF_SCHEMA_KEYS: existing reader key %s must be on_invalid reject", k.Key)
		}
	}
	if len(want) != 0 {
		t.Errorf("AF_SCHEMA_KEYS: existing reader keys missing from the fixture: %v", want)
	}
}

// TestAgenticFactorySchemaSources: every documented default carries a source
// URL, the fetch date and the uncollated marker; keys without a documented
// value carry the no-documented-value marker and no number.
func TestAgenticFactorySchemaSources(t *testing.T) {
	t.Parallel()
	s := afLoadSchema(t)
	for _, k := range s.Keys {
		if !strings.Contains(k.Key, ".rate_limit.") {
			continue
		}
		if len(k.SourceURLs) == 0 {
			t.Errorf("AF_SCHEMA_SOURCE: %s has no source_urls", k.Key)
		}
		for _, u := range k.SourceURLs {
			if !strings.HasPrefix(u, "https://docs.github.com/") && !strings.HasPrefix(u, "https://docs.gitlab.com/") {
				t.Errorf("AF_SCHEMA_SOURCE: %s source %q is not an official documentation URL", k.Key, u)
			}
		}
		if k.Fetched != afFetchedDate {
			t.Errorf("AF_SCHEMA_SOURCE: %s fetched = %q, want %q", k.Key, k.Fetched, afFetchedDate)
		}
		if k.Default != nil {
			if k.Status != "uncollated" {
				t.Errorf("AF_SCHEMA_SOURCE: %s documents a default and must be marked uncollated, got status %q", k.Key, k.Status)
			}
			if k.Provenance == "" {
				t.Errorf("AF_SCHEMA_SOURCE: %s must say how its value reached the schema (provenance)", k.Key)
			}
		} else if k.Status != "no-documented-value" {
			t.Errorf("AF_SCHEMA_SOURCE: %s has no default and must be marked no-documented-value, got %q", k.Key, k.Status)
		}
		if k.Provider == "gitlab" && k.Default != nil && k.AppliesTo != "gitlab.com" {
			t.Errorf("AF_SCHEMA_SOURCE: GitLab default %s must be scoped applies_to: gitlab.com (self-managed limits are operator-set)", k.Key)
		}
	}
}

// TestAgenticFactorySchemaRejectsOutOfRange: invalid values are errors that
// name the key; the validator never clamps, coerces, or falls back.
func TestAgenticFactorySchemaRejectsOutOfRange(t *testing.T) {
	t.Parallel()
	s := afLoadSchema(t)
	idx := afIndexKeys(s)

	for _, k := range s.Keys {
		if k.Type != "integer" {
			continue
		}
		if k.OnOutOfRange != "reject" {
			t.Errorf("AF_SCHEMA_REJECT: %s must declare on_out_of_range: reject", k.Key)
		}
		if k.Min == nil || *k.Min < 1 {
			t.Errorf("AF_SCHEMA_REJECT: %s needs a min of at least 1", k.Key)
			continue
		}
		for _, bad := range []any{0, -1, *k.Min - 1, "5", true, 1.5} {
			if err := afValidate(k, bad); err == nil {
				t.Errorf("AF_SCHEMA_REJECT: %s accepted invalid value %#v", k.Key, bad)
			} else if !strings.Contains(err.Error(), k.Key) {
				t.Errorf("AF_SCHEMA_REJECT: error for %s does not name the key: %v", k.Key, err)
			}
		}
		if err := afValidate(k, *k.Min); err != nil {
			t.Errorf("AF_SCHEMA_REJECT: %s rejected its own minimum: %v", k.Key, err)
		}
		if err := afValidate(k, nil); err != nil {
			t.Errorf("AF_SCHEMA_REJECT: %s rejected unset: %v", k.Key, err)
		}
		if k.Default != nil {
			if err := afValidate(k, k.Default); err != nil {
				t.Errorf("AF_SCHEMA_REJECT: %s default fails its own range: %v", k.Key, err)
			}
		}
	}

	// documented minimums are floors, not suggestions
	mut := idx[afKeyPrefix+"github.rate_limit.min_seconds_between_mutating_requests"]
	if afValidate(mut, 0) == nil {
		t.Errorf("AF_SCHEMA_REJECT: a pause below the documented minimum must be rejected")
	}
	wait := idx[afKeyPrefix+"github.rate_limit.secondary_unclassified_min_wait_seconds"]
	if afValidate(wait, 59) == nil || afValidate(wait, 60) != nil {
		t.Errorf("AF_SCHEMA_REJECT: the unclassified wait floor must reject 59 and accept 60")
	}

	// provider selection: invalid value rejected, no fallback
	prov := idx[afKeyPrefix+"provider.selected"]
	for _, bad := range []any{"bitbucket", "GitHub", "", 1, true} {
		if afValidate(prov, bad) == nil {
			t.Errorf("AF_SCHEMA_REJECT: provider.selected accepted %#v", bad)
		}
	}
	for _, ok := range []any{"github", "gitlab", nil} {
		if err := afValidate(prov, ok); err != nil {
			t.Errorf("AF_SCHEMA_REJECT: provider.selected rejected %#v: %v", ok, err)
		}
	}
	if err := afValidate(prov, "bitbucket"); err == nil || !strings.Contains(err.Error(), "github, gitlab") {
		t.Errorf("AF_SCHEMA_REJECT: the rejection must list the allowed values, got %v", err)
	}
}

// TestAgenticFactorySchemaUnknownHandling: a missing value with no documented
// default is UNKNOWN (nil plus a marker), never a number; the example fixture
// leaves such keys unset on purpose.
func TestAgenticFactorySchemaUnknownHandling(t *testing.T) {
	t.Parallel()
	s := afLoadSchema(t)
	idx := afIndexKeys(s)

	if !strings.Contains(s.UnknownRule, "UNKNOWN") || !strings.Contains(s.UnknownRule, "conservatively") {
		t.Errorf("AF_SCHEMA_UNKNOWN: unknown_rule must state UNKNOWN and conservative handling, got %q", s.UnknownRule)
	}
	for _, k := range s.Keys {
		if strings.Contains(k.Key, ".rate_limit.") && k.WhenUnresolved != "UNKNOWN" {
			t.Errorf("AF_SCHEMA_UNKNOWN: %s must declare when_unresolved: UNKNOWN, got %q", k.Key, k.WhenUnresolved)
		}
	}

	// example instance: every key it sets exists in the schema and validates
	raw, err := os.ReadFile(filepath.FromSlash(afExamplePath))
	if err != nil {
		t.Fatalf("AF_SCHEMA_UNKNOWN: read example: %v", err)
	}
	var doc map[string]any
	if err := yaml.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("AF_SCHEMA_UNKNOWN: example does not parse: %v", err)
	}
	configured := map[string]any{}
	var walk func(prefix string, v any)
	walk = func(prefix string, v any) {
		m, ok := v.(map[string]any)
		if !ok || len(m) == 0 {
			if prefix != "" && ok == false {
				configured[prefix] = v
			}
			return
		}
		for key, child := range m {
			p := key
			if prefix != "" {
				p = prefix + "." + key
			}
			walk(p, child)
		}
	}
	walk("", doc)
	if len(configured) == 0 {
		t.Fatalf("AF_SCHEMA_UNKNOWN: example sets no keys")
	}
	for key, v := range configured {
		k, ok := idx[key]
		if !ok {
			t.Errorf("AF_SCHEMA_UNKNOWN: example sets %s, which is not in the schema", key)
			continue
		}
		if err := afValidate(k, v); err != nil {
			t.Errorf("AF_SCHEMA_UNKNOWN: example value invalid: %v", err)
		}
	}

	// resolution: header > config > default > UNKNOWN
	mut := idx[afKeyPrefix+"github.rate_limit.min_seconds_between_mutating_requests"]
	if v, src := afResolve(mut, nil, 2); v != 2 || src != "config_value" {
		t.Errorf("AF_SCHEMA_UNKNOWN: configured value must win over the default, got %v from %s", v, src)
	}
	if v, src := afResolve(mut, nil, nil); v != 1 || src != "schema_default" {
		t.Errorf("AF_SCHEMA_UNKNOWN: unset key must resolve to the schema default, got %v from %s", v, src)
	}
	auth := idx[afKeyPrefix+"github.rate_limit.primary_authenticated_user_per_hour"]
	if v, src := afResolve(auth, 37, 5000); v != 37 || src != "response_header" {
		t.Errorf("AF_SCHEMA_UNKNOWN: a response header must win over config and default, got %v from %s", v, src)
	}
	for _, e := range afExpectedRateKeys {
		if e.Default != nil {
			continue
		}
		k := idx[afKeyPrefix+e.Suffix]
		v, src := afResolve(k, nil, nil)
		if v != nil || src != "UNKNOWN" {
			t.Errorf("AF_SCHEMA_UNKNOWN: %s unset must be UNKNOWN (nil), got %v from %s", k.Key, v, src)
		}
		if _, set := configured[k.Key]; set {
			t.Errorf("AF_SCHEMA_UNKNOWN: example sets %s; the example must leave undocumented keys unset", k.Key)
		}
	}
	// the other provider's number is never borrowed
	gl := idx[afKeyPrefix+"gitlab.rate_limit.min_seconds_between_mutating_requests"]
	if v, _ := afResolve(gl, nil, nil); v != nil {
		t.Errorf("AF_SCHEMA_UNKNOWN: GitLab pause must be UNKNOWN, not %v (the GitHub figure must not be borrowed)", v)
	}
}

// TestAgenticFactoryDefaultsStayOutOfProse: the documented default numbers
// appear in the schema fixture and nowhere in shipped template text.
func TestAgenticFactoryDefaultsStayOutOfProse(t *testing.T) {
	t.Parallel()
	s := afLoadSchema(t)
	// numbers that are specific enough to scan for; 1, 60 and 100 collide
	// with ordinary prose (and the hierarchy limits) and are covered by the
	// unit-bound detector in the template tests instead.
	numRe := map[int]*regexp.Regexp{}
	for _, k := range s.Keys {
		if n, ok := k.Default.(int); ok && n >= 80 && n != 100 {
			numRe[n] = regexp.MustCompile(`(^|[^0-9.,/-])` + regexp.QuoteMeta(commaInt(n)) + `([^0-9]|$)|(^|[^0-9.,/-])` + fmt.Sprint(n) + `([^0-9]|$)`)
		}
	}
	if len(numRe) == 0 {
		t.Fatalf("AF_SCHEMA_SOURCE: no scannable default numbers found")
	}
	for rel, text := range afAllFeatureTexts(t) {
		// SHA-256 digests and markers legitimately contain digits; strip
		// hex runs and the version marker before scanning.
		clean := regexp.MustCompile(`[0-9a-fA-F]{32,}`).ReplaceAllString(text, "")
		clean = regexp.MustCompile(`moai-template:[^\s]*`).ReplaceAllString(clean, "")
		for n, re := range numRe {
			if re.MatchString(clean) {
				t.Errorf("AF_TEMPLATE_HARDCODE: %s contains the documented limit number %d outside the schema", rel, n)
			}
		}
	}
}

// commaInt formats n with thousands separators (5000 -> 5,000).
func commaInt(n int) string {
	s := fmt.Sprint(n)
	var out []byte
	for i, c := range []byte(s) {
		if i > 0 && (len(s)-i)%3 == 0 {
			out = append(out, ',')
		}
		out = append(out, c)
	}
	return string(out)
}
