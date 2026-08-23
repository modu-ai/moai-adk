package feedback

import (
	"os"
	"sort"
	"strings"

	ghsecret "github.com/modu-ai/moai-adk/internal/github"
	"github.com/modu-ai/moai-adk/internal/sandbox"
)

// minEnvValueLength is the shortest environment value the scrubber will mask.
// A very short value is almost never a credential and is very often an ordinary
// word ("on", "true", a two-letter locale), and masking those shreds the report
// the user is trying to file. The threshold trades a theoretical miss on a
// sub-8-character secret for not destroying prose.
const minEnvValueLength = 8

// envMaskValues resolves the environment values to mask, by NAME vocabulary:
// the sandbox default deny list, extended by security.sandbox.env_scrub_extra.
// The name list is read from internal/sandbox rather than duplicated, so the
// scrubber and the sandbox cannot drift apart.
//
// The sandbox's AWS_ prefix rule is deliberately NOT adopted. In the sandbox it
// strips variables from a child process, where removing AWS_REGION costs
// nothing; here it would mask the literal string "us-east-1" wherever it
// appears in the report. The prefix covers non-secret variables, so applying it
// to a text rewriter is over-masking.
func envMaskValues(environ []string, extra []string) []string {
	names := make(map[string]bool)
	for _, n := range sandbox.DefaultEnvDenyList() {
		names[n] = true
	}
	for _, n := range extra {
		if n != "" {
			names[n] = true
		}
	}

	seen := make(map[string]bool)
	values := make([]string, 0, len(names))
	for _, kv := range environ {
		i := strings.IndexByte(kv, '=')
		if i < 0 {
			continue
		}
		name, value := kv[:i], kv[i+1:]
		if !names[name] || len(value) < minEnvValueLength || seen[value] {
			continue
		}
		seen[value] = true
		values = append(values, value)
	}

	// Longest first: when one value is a substring of another, masking the
	// longer one first keeps the shorter replacement from splitting it.
	sort.Slice(values, func(i, j int) bool { return len(values[i]) > len(values[j]) })
	return values
}

// maskEnvValues replaces every exact occurrence of a known environment value.
func maskEnvValues(s string, values []string) (string, int) {
	count := 0
	for _, v := range values {
		n := strings.Count(s, v)
		if n == 0 {
			continue
		}
		s = strings.ReplaceAll(s, v, ghsecret.MaskSecret(v))
		count += n
	}
	return s, count
}

// environOf resolves the environment source, defaulting to the process
// environment.
func environOf(opt Options) []string {
	if opt.Environ != nil {
		return opt.Environ()
	}
	return os.Environ()
}
