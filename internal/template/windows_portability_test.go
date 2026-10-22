package template

import "testing"

func TestSmartPATHUsesTargetWindowsPaths(t *testing.T) {
	env := t515EnvLookup(map[string]string{"SystemRoot": `C:\Windows`, "ProgramFiles": `C:\Program Files`})
	got := buildSmartPATHFor("windows", `C:\Users\fixture`, env, func(p string) bool { return p == `C:\Program Files\Git\bin` }, false)
	want := `C:\Users\fixture\.local\bin;C:\Users\fixture\go\bin;C:\Windows\System32;C:\Program Files\Git\bin`
	if got != want {
		t.Fatalf("target Windows PATH = %q, want %q", got, want)
	}
}

func TestSmartPATHPreservesWindowsRootAndUNC(t *testing.T) {
	for home, prefix := range map[string]string{
		`C:\`:                 `C:\.local\bin;C:\go\bin`,
		`\\server\share\user`: `\\server\share\user\.local\bin;\\server\share\user\go\bin`,
	} {
		if got := buildSmartPATHFor("windows", home, t515EnvLookup(nil), func(string) bool { return false }, false); got != prefix {
			t.Errorf("PATH for %q = %q, want %q", home, got, prefix)
		}
	}
}
