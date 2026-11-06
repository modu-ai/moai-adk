// Package sqliteuri builds the file: URI the pure-Go SQLite driver is opened with.
package sqliteuri

import "net/url"

// File returns the driver DSN for a database file at path with the given
// connection query. RED stub: reproduces the hand-built "file:"+path form.
func File(path string, query url.Values) string {
	return "file:" + path + "?" + query.Encode()
}
