package cli

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"
)

// Card t2021 repair: the entry-flag refusals must reach the operator with the
// exact flag spelling they typed. fang's ErrorText style title-cases the first
// word of every error, so a message that STARTS with a flag prints "-L",
// "--Lane" or "-F/--Factory" — uppercase -F/-L are separate flags, so the text
// would name a flag the operator never typed. These tests go through the real
// output path (runFang -> fang.Execute -> moaiErrorHandler), not err.Error().

// renderEntryError runs the real root command through runFang and returns what
// the error handler printed on stderr.
func renderEntryError(t *testing.T, args ...string) string {
	t.Helper()
	t.Setenv("CLAUDE_PROJECT_DIR", t.TempDir())
	var errOut bytes.Buffer
	rootCmd.SetErr(&errOut)
	rootCmd.SetOut(io.Discard)
	rootCmd.SetArgs(args)
	t.Cleanup(func() {
		rootCmd.SetErr(nil)
		rootCmd.SetOut(nil)
		rootCmd.SetArgs(nil)
	})
	if err := runFang(context.Background(), rootCmd); err == nil {
		t.Fatalf("%v: runFang returned nil, want the entry refusal", args)
	}
	return errOut.String()
}

func TestT2021_EntryRefusalsKeepTheTypedFlagCase(t *testing.T) {
	for _, tc := range []struct {
		name    string
		args    []string
		want    string // exact text that must appear
		notWant string // the capitalized corruption that must not appear
	}{
		{"cc -l", []string{"cc", "-l"}, "-l flag was removed; use -a", "-L"},
		{"cc --lane", []string{"cc", "--lane"}, "--lane flag was removed; use -a", "--Lane"},
		{"glm -l", []string{"glm", "-l"}, "-l flag was removed; use -a", "-L"},
		{"glm --lane", []string{"glm", "--lane"}, "--lane flag was removed; use -a", "--Lane"},
		{"cc -f -a", []string{"cc", "-f", "-a"}, "-f/--factory and -a are two entry tokens", "-F/--Factory"},
		{"cc -a -f", []string{"cc", "-a", "-f"}, "-f/--factory and -a are two entry tokens", "-F/--Factory"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := renderEntryError(t, tc.args...)
			if !strings.Contains(got, tc.want) {
				t.Errorf("printed error lacks %q:\n%s", tc.want, got)
			}
			if strings.Contains(got, tc.notWant) {
				t.Errorf("printed error carries the re-cased flag %q:\n%s", tc.notWant, got)
			}
		})
	}
}
