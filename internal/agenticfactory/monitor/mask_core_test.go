package monitor

import (
	"strings"
	"testing"
)

// B4: the monitor mask is the one core sanitizer, so what it hides includes e-mail addresses and home
// paths, and an unknown long mixed run is still dropped from a displayed field.
func TestMaskHidesEmailAndHomePathLikeTheCoreSanitizer(t *testing.T) {
	for _, in := range []string{"owner someone@example.com asked", "log at /Users/someone/work/out.log"} {
		got, changed := Mask(in)
		if !changed || strings.Contains(got, "someone") {
			t.Errorf("Mask(%q) = %q changed=%v, want the personal data hidden", in, got, changed)
		}
	}
}

func TestMaskFieldDropsAnOpaqueRunButKeepsACommitSHA(t *testing.T) {
	if v, red := MaskField("token Zx9Qm2Lp7Rt4Vw8Yb3Nc6Hd1Fj5Ks0Ag"); v != "" || !red {
		t.Errorf("opaque run: value=%q redacted=%v, want dropped", v, red)
	}
	sha := strings.Repeat("a1", 20)
	if v, red := MaskField(sha); v != sha || red {
		t.Errorf("commit sha: value=%q redacted=%v, want kept", v, red)
	}
	if v, red := MaskField("plain text"); v != "plain text" || red {
		t.Errorf("plain: value=%q redacted=%v", v, red)
	}
}
