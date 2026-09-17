package hook

// Card t804. `moai glm setup` escapes a backslash, a double quote and a dollar
// sign before writing ~/.moai/.env.glm, and glmcred.Load reverses that on the
// way out. The session-start hook reads the same file with its own parser,
// which strips the surrounding quotes but never unescapes — so a value
// containing any of those three characters was injected into
// settings.local.json in its escaped form, i.e. as a token the provider never
// issued.
//
// Measured before the repair (probe over the real Save/Load round trip, one
// fixture per character): glmcred.Load returned the original in every case;
// the hook's reader returned the escaped form for all three, and for a value
// carrying all three at once. A plain value matched on both paths, which is
// why nothing has been noticed — real GLM keys are alphanumeric.
//
// The fixtures below are invented strings, not credentials.

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/glmcred"
)

func TestLoadGLMKeyFromEnvFileReturnsTheValueThatWasSaved(t *testing.T) {
	for _, tc := range []struct {
		name  string
		value string
	}{
		{"plain", "probe-value-no-specials"},
		{"backslash", `probe\value`},
		{"double-quote", `probe"value`},
		{"dollar", `probe$value`},
		{"all-three", `probe\all"three$here`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			home := t.TempDir()
			t.Setenv("MOAI_HOME", filepath.Join(home, ".moai"))
			// glmcred.Load short-circuits on this one; an inherited value would
			// make the test pass without reading the file at all.
			t.Setenv(glmcred.EnvTestGLMKey, "")

			if err := glmcred.Save(tc.value); err != nil {
				t.Fatalf("Save: %v", err)
			}

			got := loadGLMKeyFromEnvFile()
			if got != tc.value {
				raw, _ := os.ReadFile(glmcred.Path())
				t.Errorf("hook reader = %q, want %q\nfile on disk:\n%s", got, tc.value, raw)
			}
			// Positive control: the canonical reader already round-trips, so a
			// failure here would mean the fixture, not the hook, is wrong.
			if viaLoad := glmcred.Load(); viaLoad != tc.value {
				t.Errorf("glmcred.Load = %q, want %q — fixture is not exercising the real save path", viaLoad, tc.value)
			}
		})
	}
}
