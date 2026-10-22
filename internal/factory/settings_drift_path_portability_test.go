package factory

import "testing"

func TestSettingsDriftWorktreePathKeepsHostSeparatorSemantics(t *testing.T) {
	for _, tc := range []struct {
		name              string
		recorded, current string
		separator         uint8
		want              bool
	}{
		{"windows native and git slash", `C:\Users\worker\card`, "C:/Users/worker/card", '\\', true},
		{"windows reverse spelling", "C:/Users/worker/card", `C:\Users\worker\card`, '\\', true},
		{"windows foreign worktree", `C:\Users\worker\card-other`, "C:/Users/worker/card", '\\', false},
		{"windows foreign volume", `D:\Users\worker\card`, "C:/Users/worker/card", '\\', false},
		{"never collapse parent traversal", `C:\Users\worker\other\..\card`, "C:/Users/worker/card", '\\', false},
		{"never guess case aliases", `C:\Users\worker\CARD`, "C:/Users/worker/card", '\\', false},
		{"unix literal backslash remains distinct", `/repo/a\b`, "/repo/a/b", '/', false},
		{"unix exact path", "/repo/card", "/repo/card", '/', true},
		{"unix parent traversal remains distinct", "/repo/other/../card", "/repo/card", '/', false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := settingsDriftWorktreeEqual(tc.recorded, tc.current, tc.separator); got != tc.want {
				t.Fatalf("ledger worktree equivalence(%q,%q,%q)=%v want=%v", tc.recorded, tc.current, tc.separator, got, tc.want)
			}
		})
	}
}
