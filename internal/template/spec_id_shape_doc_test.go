package template

import (
	"io/fs"
	"regexp"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/cli/specid"
)

// Card t2007 guard: the SPEC id shape that the shipped docs state for the
// `id` frontmatter field must accept and reject exactly what the enforced
// validator (internal/cli/specid, the authority behind `moai spec lint`)
// accepts and rejects. A doc that states a narrower shape makes the pre-write
// checklist reject legal ids (multi-domain ids, one-letter domains).
//
// Lexical extraction of the stated literal from the embedded copy, then a
// behavioural comparison on fixtures; it proves the documented shape, not that
// an agent applies it.

var specIDDocLiteral = regexp.MustCompile("`(\\^SPEC[^`]*\\$)`")

func TestSpecIDShapeDocsMatchValidator(t *testing.T) {
	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates(): %v", err)
	}

	docs := []struct {
		path string
		// marker selects the single line that states the id shape.
		marker string
	}{
		{".claude/skills/moai/workflows/plan/spec-assembly.md", "`id: SPEC-{DOMAIN}-{NUM}`"},
		{".claude/rules/moai/development/spec-frontmatter-schema.md", "| `id` | string |"},
	}

	fixtures := []string{
		"SPEC-AUTH-001",
		"SPEC-V3R6-SPEC-ID-VALIDATION-001",
		"SPEC-EXAMPLE-DOMAIN-001",
		"SPEC-A-001",
		"spec-auth-001",
		"SPEC-AUTH-001A",
		"SPEC-001",
		"SPEC-AUTH-01",
	}

	for _, d := range docs {
		raw, err := fs.ReadFile(fsys, d.path)
		if err != nil {
			t.Errorf("read %s: %v", d.path, err)
			continue
		}
		var literal string
		for _, line := range strings.Split(string(raw), "\n") {
			if !strings.Contains(line, d.marker) {
				continue
			}
			if m := specIDDocLiteral.FindStringSubmatch(line); m != nil {
				literal = m[1]
				break
			}
		}
		if literal == "" {
			t.Errorf("%s: no id-shape literal found on the line containing %q", d.path, d.marker)
			continue
		}
		re, err := regexp.Compile(literal)
		if err != nil {
			t.Errorf("%s: stated literal %q does not compile: %v", d.path, literal, err)
			continue
		}
		for _, id := range fixtures {
			if got, want := re.MatchString(id), specid.HasCanonicalSpecIDShape(id); got != want {
				t.Errorf("%s: literal %q gives %v for %q, validator specid.HasCanonicalSpecIDShape gives %v",
					d.path, literal, got, id, want)
			}
		}
	}
}
