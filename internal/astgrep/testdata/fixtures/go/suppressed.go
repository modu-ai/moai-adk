package fixtures

// Suppressed Go fixture — correctly paired @MX:REASON + suppression marker.
// Scanning this file should return 0 findings: the violation below is
// suppressed by the marker line adjacent to the target code.
//
// Ordering matters: ast-grep attaches the suppression of the line
// IMMEDIATELY preceding the target, so the @MX:REASON line comes first and
// the marker stays adjacent to the code (mirrors the ruby fixture
// convention used by the suppression pairing checker).

// Suppressed keeps an interface{} on purpose for the fixture.
func Suppressed() {
	// @MX:REASON test fixture for suppression policy; interface{} intentional
	// ast-grep-ignore
	var payload interface{}
	_ = payload
}
