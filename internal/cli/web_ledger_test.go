package cli

// Defect answer (review finding P2): runWeb built web.Config without Ledger, so /board and /api/board always said
// "ledger reader not provided" even when the project's AF ledger existed. These tests pin that runWeb hands web.Run the
// SAME ledger afLedgerPath resolves, and that a project without a ledger file keeps a nil reader and gets no file made.

import (
	"context"
	"os"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/ledger"
	"github.com/modu-ai/moai-adk/internal/web"
)

func pinWebLedgerEnv(t *testing.T) string {
	t.Helper()
	t.Setenv("MOAI_HOME", t.TempDir())
	t.Setenv("CLAUDE_PROJECT_DIR", t.TempDir())
	path, err := afLedgerPath("")
	if err != nil {
		t.Fatalf("afLedgerPath: %v", err)
	}
	return path
}

func TestRunWeb_ReadsTheProjectLedgerThatAFUses(t *testing.T) {
	root := t.TempDir()
	seams := materializeSeams(t, root) // sets its own MOAI_HOME; pin ours after it
	path := pinWebLedgerEnv(t)
	ctx := context.Background()
	st, err := ledger.Open(ctx, path)
	if err != nil {
		t.Fatalf("open the ledger: %v", err)
	}
	if err := st.AddCard(ctx, ledger.Card{
		ID: "w1", Title: "card w1", Position: core.Position{Stage: core.StageReady}, Priority: 1,
		Spec: ledger.Spec{Acceptance: []string{"TestX"}, Class: "A"},
	}); err != nil {
		t.Fatalf("add card: %v", err)
	}
	if err := st.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}

	var got web.Config
	runWebHarness(t, seams, withRootDir(root), func(c *runWebHarnessCfg) {
		c.captured = &got
		c.onRun = func(cfg web.Config) {
			if cfg.Ledger == nil {
				t.Error("runWeb gave web.Run no Ledger: the board would always read \"ledger reader not provided\"")
				return
			}
			cards, err := cfg.Ledger.Cards(ctx)
			if err != nil {
				t.Errorf("read cards through Config.Ledger while the server runs: %v", err)
				return
			}
			if len(cards) != 1 || cards[0].ID != "w1" {
				t.Errorf("Config.Ledger cards = %+v, want the one card w1", cards)
			}
		}
	})
	// The server has stopped: the ledger runWeb opened must be closed again.
	if got.Ledger != nil {
		if _, err := got.Ledger.Cards(ctx); err == nil {
			t.Error("the ledger is still readable after the server stopped: runWeb did not close it")
		}
	}
}

func TestRunWeb_NoLedgerFileKeepsNilReaderAndCreatesNothing(t *testing.T) {
	root := t.TempDir()
	seams := materializeSeams(t, root)
	path := pinWebLedgerEnv(t)
	var got web.Config
	runWebHarness(t, seams, withRootDir(root), func(c *runWebHarnessCfg) { c.captured = &got })
	if got.Ledger != nil {
		t.Errorf("Config.Ledger = %T, want nil when the project has no ledger file", got.Ledger)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Errorf("moai web created or touched the ledger file %s (stat err: %v)", path, err)
	}
}
