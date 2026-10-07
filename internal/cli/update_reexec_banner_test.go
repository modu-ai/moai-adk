package cli

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/tui"
	"github.com/spf13/cobra"
)

// Card t1527 D1 — one version identity surface per update run.
//
// The re-executed pass carries a hidden ARGV marker (inserted by
// reexecNewBinary via reexecChildArgv): runUpdate suppresses its top
// "Current version" KV line there, and the template-sync identity band is
// that pass's single version surface. The marker rides ARGV, NOT the
// environment — an inherited env var cannot fake the pass (card review
// round 2).

func TestReexecPassActive_FlagGated(t *testing.T) {
	// A command WITHOUT the marker set reads as an ordinary pass.
	plain := &cobra.Command{Use: "plain"}
	if reexecPassActive(plain) {
		t.Error("a command without the marker flag must not read as a re-exec pass")
	}

	// A command whose flag chain carries the marker (the re-exec child's
	// shape: the hidden persistent flag parsed from argv) reads as one.
	marked := &cobra.Command{Use: "marked"}
	marked.Flags().Bool(reexecMarkerFlag, true, "")
	if !reexecPassActive(marked) {
		t.Error("the marker flag set must read as a re-exec pass")
	}
}

func TestReexecChildArgv_InsertsMarker(t *testing.T) {
	orig := os.Args
	defer func() { os.Args = orig }()
	fake := []string{"moai", "update", "--yes", "--force"}
	os.Args = fake

	got := reexecChildArgv()

	if len(got) == 0 || got[0] != "--"+reexecMarkerFlag {
		t.Errorf("child argv must lead with the hidden marker, got %v", got)
	}
	// The tail must be the CURRENT os.Args[1:] (what the child would have
	// received anyway), element for element after the inserted marker.
	wantTail := os.Args[1:]
	if len(got) != 1+len(wantTail) {
		t.Fatalf("child argv length = %d, want 1 + %d", len(got), len(wantTail))
	}
	for i, want := range wantTail {
		if got[i+1] != want {
			t.Errorf("child argv[%d] = %q, want the original argument %q", i+1, got[i+1], want)
		}
	}
}

// TestReexecMarkerFlag_IsHiddenAndPersistent pins the registration shape: the
// marker parses wherever the original invocation placed its flags, and never
// appears in help or completion.
func TestReexecMarkerFlag_IsHiddenAndPersistent(t *testing.T) {
	f := rootCmd.PersistentFlags().Lookup(reexecMarkerFlag)
	if f == nil {
		t.Fatal("the hidden re-exec marker must be registered as a root persistent flag")
	}
	if !f.Hidden {
		t.Error("the re-exec marker flag must be Hidden")
	}
	if f.Value.String() != "false" {
		t.Errorf("the marker flag default = %q, want false", f.Value.String())
	}
}

// TestUpdateBanner_IdentityBandIsTheSingleSyncSurface pins the D1 contract at
// the render layer: the band carries the version, and no helper in the sync
// header emits a second "Current version" row.
func TestUpdateBanner_IdentityBandIsTheSingleSyncSurface(t *testing.T) {
	th := tui.LightTheme()
	band := renderIdentityBand("v9.9.9-test", th)

	plain := stripSGR(band)
	if !strings.Contains(plain, "v9.9.9-test") {
		t.Errorf("identity band must carry the version, got %q", plain)
	}
	if strings.Contains(plain, "Current version") {
		t.Errorf("identity band must not duplicate a 'Current version' KV row, got %q", plain)
	}

	var buf bytes.Buffer
	_, _ = buf.WriteString(renderIdentityBand("v9.9.9-test", th))
	// The template-sync header is band + progress line, nothing else versioned.
	if n := strings.Count(stripSGR(buf.String()), "v9.9.9-test"); n != 1 {
		t.Errorf("sync header must show the version exactly once, got %d in %q", n, stripSGR(buf.String()))
	}
}
