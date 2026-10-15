// bundle_list_test.go — U-04 (AC-RC30-05, REQ-RC30-006): `moai bundle list`
// names every catalog bundle with its skill and agent counts, and
// `moai bundle add --help` names the valid bundles.
package cli

import (
	"bytes"
	"slices"
	"strconv"
	"strings"
	"testing"
)

// TestBundleList: the list output carries one line per catalog bundle with that
// pack's skill and agent counts, and the add help names exactly the catalog's
// bundles.
func TestBundleList(t *testing.T) {
	t.Run("list_names_each_bundle_with_counts", func(t *testing.T) {
		out := runBundleCommand(t, "list")
		packs := loadCatalogForBundle().Catalog.OptionalPacks
		for _, name := range bundleNames() {
			want := []string{name, strconv.Itoa(len(packs[name].Skills)), "skills", strconv.Itoa(len(packs[name].Agents)), "agents"}
			if got := bundleLineFields(out, name); !slices.Equal(got, want) {
				t.Errorf("bundle %s: list line fields = %v, want %v\n%s", name, got, want, out)
			}
		}
	})
	t.Run("add_help_names_valid_bundles", func(t *testing.T) {
		out := runBundleCommand(t, "add", "--help")
		var named []string
		for _, line := range strings.Split(out, "\n") {
			if rest, ok := strings.CutPrefix(line, "Valid bundles: "); ok {
				named = strings.Split(rest, ", ")
			}
		}
		if !slices.Equal(named, bundleNames()) {
			t.Errorf("add --help names %v, want the catalog bundles %v\n%s", named, bundleNames(), out)
		}
	})
}

// runBundleCommand runs `moai bundle <args>` through the production command
// tree and returns its output; a command error fails the test.
func runBundleCommand(t *testing.T, args ...string) string {
	t.Helper()
	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	rootCmd.SetArgs(append([]string{"bundle"}, args...))
	t.Cleanup(func() {
		rootCmd.SetArgs(nil)
		rootCmd.SetOut(nil)
		rootCmd.SetErr(nil)
	})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("moai bundle %s: %v\n%s", strings.Join(args, " "), err, out.String())
	}
	return out.String()
}

// bundleLineFields returns the whitespace-separated fields of the output line
// that opens with name, or nil when no line does.
func bundleLineFields(out, name string) []string {
	for _, line := range strings.Split(out, "\n") {
		if fields := strings.Fields(line); len(fields) > 0 && fields[0] == name {
			return fields
		}
	}
	return nil
}
