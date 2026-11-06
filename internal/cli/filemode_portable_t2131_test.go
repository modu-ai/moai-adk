package cli

import (
	"io/fs"
	"runtime"
	"testing"
)

// permMatchesOnHost is the test-side counterpart of the production rule: POSIX
// keeps the exact permission-bit equality, Windows (no POSIX bits) accepts any
// reported mode (card t2131, #1969).
func permMatchesOnHost(perm, want fs.FileMode) bool {
	return runtime.GOOS == "windows" || perm == want
}

// weakModeRefusalExpected reports whether a deliberately weakened file mode
// (chmod 0644) is refused on this platform; Windows has no POSIX bits, so the
// negative step that relies on it runs only elsewhere while every other step of
// the same test still runs there.
func weakModeRefusalExpected() bool { return runtime.GOOS != "windows" }

func TestNoGroupOrOtherAccessOn(t *testing.T) {
	cases := []struct {
		goos string
		perm fs.FileMode
		want bool
	}{
		{"windows", 0o666, true},
		{"linux", 0o600, true},
		{"linux", 0o400, true},
		{"linux", 0o640, false},
		{"linux", 0o604, false},
		{"darwin", 0o666, false},
	}
	for _, c := range cases {
		if got := noGroupOrOtherAccessOn(c.goos, c.perm); got != c.want {
			t.Errorf("noGroupOrOtherAccessOn(%q, %o) = %v, want %v", c.goos, c.perm, got, c.want)
		}
	}
}
