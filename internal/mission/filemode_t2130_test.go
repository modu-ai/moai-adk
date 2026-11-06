package mission

import (
	"io/fs"
	"testing"
)

// Card t2130 (#1969, 21 Windows failures): on Windows the permission bits are
// not meaningful, so the readers must not reject a receipt for its mode; on
// POSIX the exact checks stay.
func TestPrivateModeOK(t *testing.T) {
	cases := []struct {
		goos string
		perm fs.FileMode
		want bool
	}{
		{"windows", 0o666, true},
		{"windows", 0o444, true},
		{"linux", 0o600, true},
		{"linux", 0o666, false},
		{"linux", 0o640, false},
		{"darwin", 0o644, false},
	}
	for _, c := range cases {
		if got := privateModeOK(c.goos, c.perm, 0o600); got != c.want {
			t.Errorf("privateModeOK(%q, %o, 0600) = %v, want %v", c.goos, c.perm, got, c.want)
		}
	}
}

func TestNoGroupOrOtherAccess(t *testing.T) {
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
		if got := noGroupOrOtherAccess(c.goos, c.perm); got != c.want {
			t.Errorf("noGroupOrOtherAccess(%q, %o) = %v, want %v", c.goos, c.perm, got, c.want)
		}
	}
}
