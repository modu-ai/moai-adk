package cli

// factory_fastcheck_entrykey_t2013_test.go — tracked item 3 (leader review of
// c14ea0bfb9): the fast check compares the card tree key after its steps only
// when the key taken at entry could be read. An entry key that cannot be read
// must fail the check, not switch the comparison off: otherwise a step that edits
// the tree passes, and the pass speaks for no tree at all.

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/modu-ai/moai-adk/internal/verify"
)

func TestT2013FastCheckUnreadableEntryKeyFailsBeforeAnyStep(t *testing.T) {
	if runtime.GOOS == "windows" || os.Geteuid() == 0 {
		t.Skip("the fixture relies on a file mode that denies the owner")
	}
	base := map[string]string{
		"go.mod":       "module example.com/m\n",
		"base.txt":     "original\n",
		"calc/calc.go": "package calc\n",
	}
	mutate := "package calc\n\nimport (\n\t\"os\"\n\t\"testing\"\n)\n\nfunc TestMutate(t *testing.T) {\n\tif err := os.WriteFile(\"../base.txt\", []byte(\"mutated while the check ran\\n\"), 0o644); err != nil {\n\t\tt.Fatal(err)\n\t}\n}\n"
	dir := fastCheckRepo(t, base, map[string]string{"calc/mutate_test.go": mutate}, nil)

	// An untracked file the owner cannot read makes verify.Key fail, while the
	// build, vet and test steps still run to completion.
	if err := os.WriteFile(filepath.Join(dir, "unreadable.bin"), []byte("x"), 0o000); err != nil {
		t.Fatal(err)
	}
	if _, err := verify.Key(context.Background(), dir); err == nil {
		t.Fatal("premise: verify.Key over the fixture succeeded, want an error")
	}

	_, tail, err := factoryFastCheck(context.Background(), io.Discard, dir, "main")
	t.Logf("factoryFastCheck: err=%v tail=%q", err, tail)
	if err == nil {
		t.Error("the check passed although the card tree key could not be read at entry")
	}
	got, readErr := os.ReadFile(filepath.Join(dir, "base.txt"))
	if readErr != nil {
		t.Fatal(readErr)
	}
	if string(got) != "original\n" {
		t.Errorf("base.txt = %q: a step ran and edited the tree although the entry key was unreadable", got)
	}
}
