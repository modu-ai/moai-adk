package cli

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/tui"
)

// Card t1527 D1 — one version identity surface per update run.
//
// The re-executed pass (MOAI_UPDATE_REEXEC=1, set by markReexecPass before the
// freshly installed binary replaces the process) suppresses runUpdate's top
// "Current version" KV line; the template-sync identity band is that pass's
// single version surface. These tests pin the suppression helper and the
// renderIdentityBand contract the suppression leans on.

func TestReexecPassActive_EnvGated(t *testing.T) {
	t.Setenv(config.EnvUpdateReexec, "")
	if reexecPassActive() {
		t.Error("an unset MOAI_UPDATE_REEXEC must not read as a re-exec pass")
	}

	t.Setenv(config.EnvUpdateReexec, "1")
	if !reexecPassActive() {
		t.Error("MOAI_UPDATE_REEXEC=1 must read as a re-exec pass")
	}
}

func TestMarkReexecPass_SetsBothMarkers(t *testing.T) {
	t.Setenv("MOAI_SKIP_BINARY_UPDATE", "")
	t.Setenv(config.EnvUpdateReexec, "")

	if err := markReexecPass(); err != nil {
		t.Fatalf("markReexecPass: %v", err)
	}
	if got := os.Getenv("MOAI_SKIP_BINARY_UPDATE"); got != "1" {
		t.Errorf("MOAI_SKIP_BINARY_UPDATE = %q, want 1", got)
	}
	if got := os.Getenv(config.EnvUpdateReexec); got != "1" {
		t.Errorf("%s = %q, want 1", config.EnvUpdateReexec, got)
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
