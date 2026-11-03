package cli

// AC-AF-021 case (a): the launcher entry refusals are a contract an operator reads
// on a terminal. Each fixed input must keep refusing with byte-identical text.

import (
	"os"
	"strings"
	"testing"
)

const launcherEntryRefusalsGolden = "testdata/agentic_factory_compat/entry_refusals.golden"

// launcherEntryRefusalRows are the inputs, in golden order. Every one is refused today.
var launcherEntryRefusalRows = [][]string{
	{"-f", "4"},
	{"-f", "worker"},
	{"-f", "lane-2"},
	{"-f", "-l"},
	{"-f", "-a"},
	{"-l", "-a"},
	{"-l", "x"},
	{"-a", "x"},
	{"-l", "--name", "foo"},
	{"-f", "--name", "lane-2"},
	{"-f", "--leader", "leader-2"},
	{"--leader"},
	{"-l", "--leader", "lead-1"},
	{"-l", "--clear-policy", "bogus"},
	{"-f", "--no-auto-dispatch"},
}

func TestCompat_LauncherEntryRefusalsMatchGolden(t *testing.T) {
	raw, readErr := os.ReadFile(launcherEntryRefusalsGolden)

	var actual []string
	for _, args := range launcherEntryRefusalRows {
		_, err := parseLauncherEntry(args)
		if err == nil {
			t.Errorf("%q: parseLauncherEntry accepted the input, want a refusal", args)
			continue
		}
		actual = append(actual, strings.Join(args, " ")+"\t"+err.Error())
	}
	if readErr != nil {
		t.Fatalf("golden %s unreadable (%v); current refusals:\n%s", launcherEntryRefusalsGolden, readErr, strings.Join(actual, "\n"))
	}

	want := strings.Split(strings.TrimSuffix(string(raw), "\n"), "\n")
	if len(want) == 0 || want[0] == "" {
		t.Fatal("golden holds no rows: the comparison would be vacuous")
	}
	if len(want) != len(launcherEntryRefusalRows) || len(actual) != len(want) {
		t.Fatalf("golden has %d rows, the table %d, refused %d: they must be equal", len(want), len(launcherEntryRefusalRows), len(actual))
	}
	for i := range want {
		if actual[i] != want[i] {
			t.Errorf("row %d differs\n got: %q\nwant: %q", i+1, actual[i], want[i])
		}
	}
	t.Logf("compared %d refusal rows byte-for-byte", len(want))
}
