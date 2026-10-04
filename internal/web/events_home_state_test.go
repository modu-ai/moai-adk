package web

import (
	"os"
	"testing"

	"github.com/modu-ai/moai-adk/internal/factory"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

func TestResolvedWatchPathsIncludeHomeTodoAndFactory(t *testing.T) {
	t.Setenv("MOAI_HOME", t.TempDir())
	root, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := homestate.EnsureProjectLayout(root); err != nil {
		t.Fatal(err)
	}

	paths := resolvedWatchPaths(root)
	if got := paths[factory.StateDirForRoot(root)]; got != "factory" {
		t.Fatalf("todo watch event = %q, want factory", got)
	}
	factoryDir, err := homestate.FactoryDir(root)
	if err != nil {
		t.Fatal(err)
	}
	if got := paths[factoryDir]; got != "factory" {
		t.Fatalf("factory watch event = %q, want factory", got)
	}
}
