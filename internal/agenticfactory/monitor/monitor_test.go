package monitor

import (
	"context"
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

type fakeClaims struct {
	rows map[int]ClaimRow
	err  error
}

func (f fakeClaims) ReadClaim(_ context.Context, issue int) (ClaimRow, error) {
	if f.err != nil {
		return ClaimRow{}, f.err
	}
	r, ok := f.rows[issue]
	if !ok {
		return ClaimRow{}, ErrNotFound
	}
	return r, nil
}

type fakeRemote struct {
	checks []CheckRun
	err    error
}

func (f fakeRemote) RequiredChecks(context.Context, string) ([]CheckRun, error) {
	return f.checks, f.err
}

type fakeLedger struct{ next string }

func (f fakeLedger) NextAction(context.Context, int) (string, error) { return f.next, nil }

var t0 = time.Date(2026, 10, 29, 6, 0, 0, 0, time.UTC)

const head = "0123456789abcdef0123456789abcdef01234567"

func collector(c ClaimReader, r RemoteReader, l LedgerReader) *Collector {
	return &Collector{Claims: c, Remote: r, Ledger: l, Now: func() time.Time { return t0 }}
}

func TestCollectFillsRowWithSourcesAndReadTime(t *testing.T) {
	c := collector(
		fakeClaims{rows: map[int]ClaimRow{7: {Project: "p", Issue: 7, Owner: "lane-18", State: "implementing", Stage: "implementing", Attempt: 2, Epoch: 3, HeadSHA: head, StoreSeq: 9, MirrorState: "current", LastCheckedAt: t0.Add(-time.Minute)}}},
		fakeRemote{checks: []CheckRun{{Name: "build", Conclusion: "success", HeadSHA: head}}},
		fakeLedger{next: "request verification"},
	)
	rows := c.Collect(context.Background(), []int{7})
	if len(rows) != 1 {
		t.Fatalf("rows = %d", len(rows))
	}
	r := rows[0]
	if r.Owner.Value != "lane-18" || r.Owner.Source != SrcClaimStore || !r.Owner.ReadAt.Equal(t0) {
		t.Errorf("owner field = %+v", r.Owner)
	}
	if r.Checks.Source != SrcRemote || len(r.Checks.Value) != 1 {
		t.Errorf("checks field = %+v", r.Checks)
	}
	if r.NextAction.Source != SrcLedger || r.NextAction.Value != "request verification" {
		t.Errorf("next action = %+v", r.NextAction)
	}
}

func TestUnavailableSourceIsUnconfirmedNotGuessed(t *testing.T) {
	c := collector(fakeClaims{err: errors.New("store unreadable")}, fakeRemote{err: errors.New("gh 502")}, fakeLedger{})
	r := c.Collect(context.Background(), []int{7})[0]
	if r.State.Confirmed || r.State.Value != "" {
		t.Errorf("state = %+v, want unconfirmed empty", r.State)
	}
	if r.Checks.Confirmed {
		t.Errorf("checks = %+v, want unconfirmed", r.Checks)
	}
}

func TestStaleChecksAreFlagged(t *testing.T) {
	other := strings.Repeat("a", 40)
	c := collector(
		fakeClaims{rows: map[int]ClaimRow{7: {Issue: 7, HeadSHA: head}}},
		fakeRemote{checks: []CheckRun{{Name: "build", Conclusion: "success", HeadSHA: other}}},
		fakeLedger{},
	)
	r := c.Collect(context.Background(), []int{7})[0]
	if !r.ChecksStale {
		t.Error("checks for a different HEAD must be flagged stale")
	}
}

// The read path must not be able to write: its source interfaces carry no write-verb methods and
// the package imports nothing that mutates (design §8.1, AC-AF-022).
func TestReadBoundaryHasNoWriteVerbs(t *testing.T) {
	fset := token.NewFileSet()
	files, _ := filepath.Glob("*.go")
	verbs := []string{"Set", "Write", "Put", "Delete", "Remove", "Create", "Update", "Claim", "Send", "Merge", "Push", "Resume", "Hold", "Stop", "Kill", "Dispatch", "Advance", "Renew", "Exec", "Run"}
	allowedImports := map[string]bool{"context": true, "encoding/json": true, "errors": true, "fmt": true, "regexp": true, "sort": true, "strings": true, "time": true}
	for _, f := range files {
		if strings.HasSuffix(f, "_test.go") {
			continue
		}
		af, err := parser.ParseFile(fset, f, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		for _, im := range af.Imports {
			p := strings.Trim(im.Path.Value, `"`)
			if !allowedImports[p] {
				t.Errorf("%s imports %q: monitor may import only read-safe stdlib packages", f, p)
			}
		}
		ast.Inspect(af, func(n ast.Node) bool {
			it, ok := n.(*ast.InterfaceType)
			if !ok {
				return true
			}
			for _, m := range it.Methods.List {
				for _, name := range m.Names {
					for _, v := range verbs {
						if strings.HasPrefix(name.Name, v) {
							t.Errorf("%s: interface method %s starts with write verb %s", f, name.Name, v)
						}
					}
				}
			}
			return true
		})
	}
	if _, err := os.Stat("mask.go"); err != nil {
		t.Fatal("mask.go missing")
	}
}
