package cli

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGatewayPeerRegistryBidirectionalAndPrivateHistory(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	source := t.TempDir()
	native := t.TempDir()
	in := gatewayLaunchRequest{OriginalConfig: source, OriginalConfigSet: true}
	os.Mkdir(filepath.Join(source, "sessions"), 0700)
	os.WriteFile(filepath.Join(source, "sessions", "lead.json"), []byte("lead"), 0600)
	if err := shareGatewayPeerRegistry(native, in); err != nil {
		t.Fatal(err)
	}
	if b, err := os.ReadFile(filepath.Join(native, "sessions", "lead.json")); err != nil || string(b) != "lead" {
		t.Fatal("lead invisible", err)
	}
	os.WriteFile(filepath.Join(native, "sessions", "lane.json"), []byte("lane"), 0600)
	if b, err := os.ReadFile(filepath.Join(source, "sessions", "lane.json")); err != nil || string(b) != "lane" {
		t.Fatal("lane invisible", err)
	}
	if err := shareGatewayPeerRegistry(native, in); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"projects", ".claude.json", ".credentials.json"} {
		if _, err := os.Lstat(filepath.Join(native, name)); !os.IsNotExist(err) {
			t.Fatal("private state shared", name)
		}
	}
	other := t.TempDir()
	if err := shareGatewayPeerRegistry(native, gatewayLaunchRequest{OriginalConfig: other, OriginalConfigSet: true}); err == nil {
		t.Fatal("foreign profile silently reused")
	}
}
