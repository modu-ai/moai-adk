package cli

import (
	"context"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/factorymsg"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

func TestFactoryRunSelectionAtomicSlotsAndArgv(t *testing.T) {
	root := t.TempDir()
	t.Setenv("MOAI_HOME", t.TempDir())
	if _, err := factorymsg.ResolveActiveRun(context.Background(), root, ""); err == nil || !strings.Contains(err.Error(), "NO_ACTIVE_FACTORY") {
		t.Fatalf("zero=%v", err)
	}
	db, err := homestate.OpenFactory(root)
	if err != nil {
		t.Fatal(err)
	}
	if err := db.RecordRun(context.Background(), homestate.FactoryRun{RunID: "run-a", Backend: "codex", ManifestJSON: "{}"}); err != nil {
		t.Fatal(err)
	}
	if got, err := factorymsg.ResolveActiveRun(context.Background(), root, ""); err != nil || got != "run-a" {
		t.Fatalf("one=%q %v", got, err)
	}
	if err := db.RecordRun(context.Background(), homestate.FactoryRun{RunID: "run-b", Backend: "claude", ManifestJSON: "{}"}); err != nil {
		t.Fatal(err)
	}
	_ = db.Close()
	if _, err := factorymsg.ResolveActiveRun(context.Background(), root, ""); err == nil || !strings.Contains(err.Error(), "AMBIGUOUS_FACTORY") {
		t.Fatalf("many=%v", err)
	}
	if got, err := factorymsg.ResolveActiveRun(context.Background(), root, "run-b"); err != nil || got != "run-b" {
		t.Fatalf("explicit=%q %v", got, err)
	}
	p, err := parseFactoryFlag([]string{"-f", "agent", "--factory-run", "run-b", "--", "--factory-run", "child", "x"})
	if err != nil {
		t.Fatal(err)
	}
	if p.RunID != "run-b" || !p.AgentRole || len(p.Rest) != 4 || p.Rest[0] != "--" || p.Rest[1] != "--factory-run" || p.Rest[2] != "child" || p.Rest[3] != "x" {
		t.Fatalf("parse=%+v", p)
	}
}
