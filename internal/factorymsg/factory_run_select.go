package factorymsg

import "time"

// kstZone is a fixed UTC+9 zone. LoadLocation("Asia/Seoul") would need tzdata
// and time.Local would follow the host TZ; a fixed zone renders the same bytes
// everywhere.
var kstZone = time.FixedZone("KST", 9*60*60)

// FormatKST renders a stored RFC 3339 time as "YYYY-MM-DD HH:MM:SS KST"
// (SPEC-LANE-JOIN-RUN-SELECT-001 REQ-LJR-002). A value that does not parse is
// returned unchanged, so a malformed stored time is shown rather than hidden.
func FormatKST(raw string) string {
	at, err := time.Parse(time.RFC3339Nano, raw)
	if err != nil {
		return raw
	}
	return at.In(kstZone).Format("2006-01-02 15:04:05") + " KST"
}
