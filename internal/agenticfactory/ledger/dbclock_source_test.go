package ledger

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"
)

// AC-AFL-004: every timestamp the ledger writes comes from the database clock
// (the strftime(...,'now') text), never from a Go time value. The check reads the
// package source, so it covers every write site, including ones no test reaches.

const (
	// dbClock is the database-clock text; nowSQL in schema.go must fold to it.
	dbClock = `strftime('%Y-%m-%dT%H:%M:%fZ','now')`
	// offsetClock stands for a folded offsetSQL(...) call: still the database clock, but not the plain one.
	offsetClock = `strftime('%Y-%m-%dT%H:%M:%fZ','now','<offset>')`
	// dyn marks an operand the folder cannot resolve to a constant (a variable, a call, time.Now().Format()).
	dyn = "\x00"
)

var timestampColumns = []string{"created_at", "updated_at", "at", "heartbeat_at", "enqueued_at"}

var forbiddenTimeCalls = map[string]bool{"Now": true, "Since": true, "Until": true, "After": true}

type clockReport struct {
	sites     map[string]int // column -> write sites seen
	problems  []string       // a value written to a timestamp column that is not the database clock
	timeCalls []string       // time.Now/Since/Until/After calls
}

// foldConsts collects the package-level string constants, resolving constants that reference other constants.
func foldConsts(files []*ast.File) map[string]string {
	consts := map[string]string{}
	for pass := 0; pass < 5; pass++ {
		for _, f := range files {
			for _, d := range f.Decls {
				gd, ok := d.(*ast.GenDecl)
				if !ok || gd.Tok != token.CONST {
					continue
				}
				for _, sp := range gd.Specs {
					vs := sp.(*ast.ValueSpec)
					for i, n := range vs.Names {
						if i < len(vs.Values) {
							if v := foldString(vs.Values[i], consts); !strings.Contains(v, dyn) {
								consts[n.Name] = v
							}
						}
					}
				}
			}
		}
	}
	return consts
}

// foldString folds a string expression built from literals, constants and "+". An operand it cannot resolve
// becomes dyn; offsetSQL(...) becomes offsetClock.
func foldString(e ast.Expr, consts map[string]string) string {
	switch x := e.(type) {
	case *ast.BasicLit:
		if x.Kind == token.STRING {
			if s, err := strconv.Unquote(x.Value); err == nil {
				return s
			}
		}
	case *ast.Ident:
		if v, ok := consts[x.Name]; ok {
			return v
		}
	case *ast.ParenExpr:
		return foldString(x.X, consts)
	case *ast.BinaryExpr:
		if x.Op == token.ADD {
			return foldString(x.X, consts) + foldString(x.Y, consts)
		}
	case *ast.CallExpr:
		if id, ok := x.Fun.(*ast.Ident); ok && id.Name == "offsetSQL" {
			return offsetClock
		}
	}
	return dyn
}

// splitTop splits s at top-level commas (outside parentheses and single-quoted text).
func splitTop(s string) []string {
	var out []string
	depth, start, inQ := 0, 0, false
	for i := 0; i < len(s); i++ {
		switch c := s[i]; {
		case c == '\'':
			inQ = !inQ
		case inQ:
		case c == '(':
			depth++
		case c == ')':
			depth--
		case c == ',' && depth == 0:
			out = append(out, strings.TrimSpace(s[start:i]))
			start = i + 1
		}
	}
	return append(out, strings.TrimSpace(s[start:]))
}

// cutAtTop returns s up to the first top-level occurrence of the upper-case keyword kw ("" kw = whole string).
func cutAtTop(s, kw string) string {
	depth, inQ := 0, false
	up := strings.ToUpper(s)
	for i := 0; i < len(s); i++ {
		switch c := s[i]; {
		case c == '\'':
			inQ = !inQ
		case inQ:
		case c == '(':
			depth++
		case c == ')':
			depth--
		case depth == 0 && strings.HasPrefix(up[i:], kw) && i > 0 && (s[i-1] == ' ' || s[i-1] == '\n' || s[i-1] == '\t'):
			return s[:i]
		}
	}
	return s
}

// parenGroup returns the text inside the first balanced "(...)" of s, or "".
func parenGroup(s string) string {
	depth, inQ, start := 0, false, -1
	for i := 0; i < len(s); i++ {
		switch c := s[i]; {
		case c == '\'':
			inQ = !inQ
		case inQ:
		case c == '(':
			if depth == 0 {
				start = i + 1
			}
			depth++
		case c == ')':
			depth--
			if depth == 0 && start >= 0 {
				return s[start:i]
			}
		}
	}
	return ""
}

var (
	insertRe = regexp.MustCompile(`(?is)^\s*INSERT\s+(?:OR\s+\w+\s+)?INTO\s+\w+\s*\(([^)]*)\)\s*VALUES\s*(\(.*)$`)
	updateRe = regexp.MustCompile(`(?is)^\s*UPDATE\s+\w+\s+SET\s+(.*)$`)
	doUpdRe  = regexp.MustCompile(`(?is)\bDO\s+UPDATE\s+SET\s+(.*)$`)
)

func isTracked(col string) bool {
	for _, c := range timestampColumns {
		if c == col {
			return true
		}
	}
	return false
}

// judge records one value written to col in the statement.
func (r *clockReport) judge(where, col, val string) {
	if !isTracked(col) {
		return
	}
	r.sites[col]++
	switch {
	case val == dbClock:
	case val == "excluded."+col: // an upsert copying the value its own INSERT carried; that INSERT is judged separately
	default:
		r.problems = append(r.problems, where+": "+col+" is written from "+strconv.Quote(strings.ReplaceAll(val, dyn, "<dynamic>"))+", not the database clock")
	}
}

func (r *clockReport) assignments(where, list string) {
	for _, a := range splitTop(cutAtTop(list, "WHERE")) {
		if col, val, ok := strings.Cut(a, "="); ok {
			r.judge(where, strings.TrimSpace(col), strings.TrimSpace(val))
		}
	}
}

func (r *clockReport) statement(where, sql string) {
	if m := insertRe.FindStringSubmatch(sql); m != nil {
		if strings.Contains(m[1], dyn) {
			return
		}
		cols, vals := splitTop(m[1]), splitTop(parenGroup(m[2]))
		if len(cols) == len(vals) {
			for i, c := range cols {
				r.judge(where, c, vals[i])
			}
		}
		if u := doUpdRe.FindStringSubmatch(sql); u != nil {
			r.assignments(where, u[1])
		}
		return
	}
	if m := updateRe.FindStringSubmatch(sql); m != nil {
		r.assignments(where, m[1])
	}
}

// checkClockSources parses the given sources (file name -> text) and reports every write to a timestamp column.
func checkClockSources(t *testing.T, srcs map[string]string) clockReport {
	t.Helper()
	fset := token.NewFileSet()
	var files []*ast.File
	names := make([]string, 0, len(srcs))
	for n := range srcs {
		names = append(names, n)
	}
	sort.Strings(names)
	for _, n := range names {
		f, err := parser.ParseFile(fset, n, srcs[n], 0)
		if err != nil {
			t.Fatalf("parse %s: %v", n, err)
		}
		files = append(files, f)
	}
	consts := foldConsts(files)
	rep := clockReport{sites: map[string]int{}}
	for _, f := range files {
		inChain := map[ast.Node]bool{} // parts of a "a"+b+"c" chain already folded as one statement
		visit := func(n ast.Node) bool {
			switch x := n.(type) {
			case *ast.BinaryExpr:
				if x.Op == token.ADD && !inChain[x] {
					rep.statement(fset.Position(x.Pos()).String(), foldString(x, consts))
					ast.Inspect(x, func(c ast.Node) bool {
						switch cc := c.(type) {
						case *ast.BinaryExpr:
							if cc.Op == token.ADD {
								inChain[cc] = true
							}
						case *ast.BasicLit:
							inChain[cc] = true
						}
						return true
					})
				}
			case *ast.BasicLit:
				if x.Kind == token.STRING && !inChain[x] {
					rep.statement(fset.Position(x.Pos()).String(), foldString(x, consts))
				}
			case *ast.CallExpr:
				if sel, ok := x.Fun.(*ast.SelectorExpr); ok {
					if id, ok := sel.X.(*ast.Ident); ok && id.Name == "time" && forbiddenTimeCalls[sel.Sel.Name] {
						rep.timeCalls = append(rep.timeCalls, fset.Position(x.Pos()).String()+": time."+sel.Sel.Name)
					}
				}
			}
			return true
		}
		ast.Inspect(f, visit)
	}
	return rep
}

const fixtureHead = "package fx\n\nconst nowSQL = `strftime('%Y-%m-%dT%H:%M:%fZ','now')`\n\n"

func TestLedgerSource_TimestampColumnsUseDBClock(t *testing.T) {
	// Fixtures first: the checker must see both mutants fail and a clean source pass.
	t.Run("fixture_clean_source_passes", func(t *testing.T) {
		rep := checkClockSources(t, map[string]string{"clean.go": fixtureHead +
			"func f(db DB) { db.Exec(`INSERT INTO notes(kind,card_id,text,at) VALUES('x',?,?,`+nowSQL+`)`, 1, 2) }\n"})
		if len(rep.problems) != 0 || len(rep.timeCalls) != 0 || rep.sites["at"] != 1 {
			t.Fatalf("clean fixture: sites=%v problems=%v timeCalls=%v", rep.sites, rep.problems, rep.timeCalls)
		}
	})
	t.Run("fixture_m1_placeholder_bound_to_a_go_time", func(t *testing.T) {
		rep := checkClockSources(t, map[string]string{"m1.go": fixtureHead +
			"func f(db DB, when T) { db.Exec(`INSERT INTO notes(kind,card_id,text,at) VALUES('x',?,?,?)`, 1, 2, when) }\n"})
		if len(rep.problems) != 1 || !strings.Contains(rep.problems[0], "at is written from") {
			t.Fatalf("m1 must be reported once, got problems=%v", rep.problems)
		}
	})
	t.Run("fixture_m2_time_now_format_concatenated_into_sql", func(t *testing.T) {
		rep := checkClockSources(t, map[string]string{"m2.go": fixtureHead +
			"func f(db DB) { db.Exec(`UPDATE cards SET stage=?, updated_at='`+time.Now().Format(time.RFC3339)+`' WHERE id=?`, 1, 2) }\n"})
		if len(rep.problems) != 1 || !strings.Contains(rep.problems[0], "<dynamic>") {
			t.Fatalf("m2 must be reported as a dynamic value, got problems=%v", rep.problems)
		}
		if len(rep.timeCalls) != 1 {
			t.Fatalf("m2 must also be reported as a time.Now call, got %v", rep.timeCalls)
		}
	})

	// The package itself.
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	srcs := map[string]string{}
	for _, e := range entries {
		n := e.Name()
		if e.IsDir() || !strings.HasSuffix(n, ".go") || strings.HasSuffix(n, "_test.go") {
			continue
		}
		b, err := os.ReadFile(n)
		if err != nil {
			t.Fatal(err)
		}
		srcs[n] = string(b)
	}
	if len(srcs) == 0 {
		t.Fatal("no non-test source files found: the sweep is empty")
	}
	rep := checkClockSources(t, srcs)
	for _, col := range timestampColumns {
		if rep.sites[col] == 0 {
			t.Errorf("column %s has no write site in %d source files: the sweep would pass vacuously", col, len(srcs))
		}
	}
	for _, p := range rep.problems {
		t.Error(p)
	}
	for _, c := range rep.timeCalls {
		t.Error(c)
	}
	t.Logf("swept %d source files; write sites per column: %v", len(srcs), rep.sites)
}
