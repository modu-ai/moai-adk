package cli

import (
	"strings"
	"testing"

	"github.com/spf13/pflag"
)

// t1972 F4 (carried from t1966): `moai init --help` showed "(default: true) (true)"
// for --enforce-quality and --enable-design. The help renderer already appends a
// bool flag's true default as "(true)" (as it does for --enable-lsp), so a usage
// string that also states "(default ...)" prints the default twice. The default
// must come from one place: the registered flag value, never the usage text.
func TestInitBoolFlagUsageDoesNotRepeatTheDefaultTheRendererPrints(t *testing.T) {
	seen := 0
	initCmd.Flags().VisitAll(func(f *pflag.Flag) {
		if f.Value.Type() != "bool" || f.DefValue != "true" {
			return
		}
		seen++
		if strings.Contains(strings.ToLower(f.Usage), "default") {
			t.Errorf("--%s defaults to true and the renderer prints that; its usage must not state the default again: %q", f.Name, f.Usage)
		}
	})
	// premise: the scan saw the flags this finding is about (an empty scan passes vacuously)
	if seen < 3 {
		t.Fatalf("saw %d bool flags defaulting to true, want at least 3 (--enable-lsp, --enforce-quality, --enable-design)", seen)
	}
}
