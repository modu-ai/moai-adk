package cli

import "testing"

func TestCodexFindingWindowsPathSpacesRemainWhole(t *testing.T) {
	for _, tc := range []struct {
		name, body, path string
		line             int
	}{
		{"backtick-drive", "Fix `C:\\Users\\First Last\\repo\\file.go:42` without losing scope", `C:\Users\First Last\repo\file.go`, 42},
		{"quoted-drive", `Fix "D:\Team Work\repo\file.go:9" now`, `D:\Team Work\repo\file.go`, 9},
		{"raw-drive", `The location C:\Users\First Last\repo\file.go:42 must survive`, `C:\Users\First Last\repo\file.go`, 42},
		{"quoted-unc", "Fix `\\\\server\\Team Share\\repo\\file.go:17`", `\\server\Team Share\repo\file.go`, 17},
		{"quoted-relative", "Fix `internal\\team docs\\file.go:23`", `internal\team docs\file.go`, 23},
		{"first-location", "`C:\\Team Work\\file.go:3` and `C:\\Team Work\\file.go:7`", `C:\Team Work\file.go`, 3},
	} {
		t.Run(tc.name, func(t *testing.T) {
			findings := codexFindingsOf("- [P2] " + tc.body)
			if len(findings) != 1 || findings[0].File != tc.path || findings[0].Line != tc.line {
				t.Fatalf("findings=%+v, want whole path %q:%d", findings, tc.path, tc.line)
			}
		})
	}
}

func TestCodexFindingWindowsPathSpaceAmbiguityAndURLsStayStrict(t *testing.T) {
	for _, body := range []string{
		"`C:\\Team Work\\one.go:4` conflicts with `C:\\Team Work\\two.go:5`",
		"https://example.invalid/repo/file.go:42",
		"`https://example.invalid/repo/file.go:42`",
	} {
		if path, line, ok := codexFindingAnchorOf(body); ok || path != "" || line != 0 {
			t.Errorf("invented anchor=%q:%d from %q", path, line, body)
		}
	}
}
