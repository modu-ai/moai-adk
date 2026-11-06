// Package sqliteuri builds the file: URI the pure-Go SQLite driver is opened with.
//
// A DSN built by concatenating the file: scheme and the path, or a url.URL given a raw Windows path, hands the
// driver file://C:/… and the drive letter parses as the URI authority ("invalid
// uri authority: C:"). File emits the root-absolute form file:///C:/… instead,
// and percent-encodes spaces, '#', '%' and '?' so they stay part of the path.
package sqliteuri

import (
	"net/url"
	"path/filepath"
	"strings"
)

// File returns the driver DSN for the database file at path with the given
// connection query (may be nil). It is pure string work, so Windows drive
// paths are handled identically on every OS. POSIX paths keep their bytes.
// A relative path is made absolute first: a bare relative path in a URI with
// an authority marker would be read as a host name.
//
// @MX:ANCHOR: [AUTO] the one builder for file: SQLite DSNs
// @MX:REASON: ledger, homestate, factorymsg, discovery and cli all open databases through it; a drive-path regression fails every Windows open.
func File(path string, query url.Values) string {
	p := path
	if hasDrive(p) {
		p = strings.ReplaceAll(p, `\`, "/")
	} else {
		p = filepath.ToSlash(p)
		if !strings.HasPrefix(p, "/") {
			if abs, err := filepath.Abs(path); err == nil {
				p = filepath.ToSlash(abs)
			}
		}
	}
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	return (&url.URL{Scheme: "file", Path: p, RawQuery: query.Encode()}).String()
}

// hasDrive reports a leading Windows drive designator such as C: or d:.
func hasDrive(p string) bool {
	if len(p) < 2 || p[1] != ':' {
		return false
	}
	c := p[0]
	return c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z'
}
