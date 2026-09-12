package template

// workflow_worktree_key_honesty_test.go — SPEC-CONFIG-KEY-HONESTY-001, for the
// worktree toggles as the SHIPPED file states them (#1705).
//
// The template's comment told the user that `auto_cleanup` was declared but
// not read, grouped into one sentence with `auto_merge`, which really has no
// reader. `auto_cleanup` gates both auto-cleanup paths, and each of them runs
// a `git worktree remove` immediately after reading it. A user who believed
// the file would either flip the key while reasoning about something else, or
// rule it out after an unexplained deletion; both readings lose work.
//
// The wording is not what is pinned here — the claim is. For each toggle, what
// the shipped comment says about being read has to match whether the Go tree
// reads the field, in both directions.

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"
)

const workflowTemplatePath = ".moai/config/sections/workflow.yaml"

// worktreeToggles maps each `workflow.worktree.*` key to its struct field on
// WorkflowWorktreeConfig (internal/config/types.go).
var worktreeToggles = map[string]string{
	"auto_create":          "AutoCreate",
	"auto_merge":           "AutoMerge",
	"auto_cleanup":         "AutoCleanup",
	"session_name_pattern": "SessionNamePattern",
}

// unreadClaims are the phrasings the template uses to call a key inert.
var unreadClaims = []string{"not read", "reserved"}

// keySection matches the start of a per-key comment clause, `# auto_cleanup:`.
var keySection = regexp.MustCompile(`^#\s*([a-z_]+):`)

// goFile matches a source path named inside a comment.
var goFile = regexp.MustCompile(`[\w./-]+\.go`)

// commentSections groups the `#` comment lines of a YAML document into the
// clause each one belongs to: a clause starts at a `# <key>:` line and runs
// until the next such line or the end of the comment block. The claim about a
// key spans more than the line that names it — `auto_cleanup`'s names its two
// reader files on the lines after — so a per-line reading would miss half of
// what the file asserts.
func commentSections(content string) map[string][]string {
	sections := map[string][]string{}
	current := ""
	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "#") {
			current = ""
			continue
		}
		if match := keySection.FindStringSubmatch(trimmed); match != nil {
			current = match[1]
		}
		if current != "" {
			sections[current] = append(sections[current], trimmed)
		}
	}
	return sections
}

// sortedKeys renders a set in a stable order, so a failure reads the same twice.
func sortedKeys(set map[string]bool) []string {
	out := make([]string, 0, len(set))
	for key := range set {
		out = append(out, key)
	}
	sort.Strings(out)
	return out
}

// namedGoFiles returns the base names of the source files a clause names.
func namedGoFiles(lines []string) map[string]bool {
	named := map[string]bool{}
	for _, line := range lines {
		for _, match := range goFile.FindAllString(line, -1) {
			named[filepath.Base(match)] = true
		}
	}
	return named
}

// fieldReaders returns the `internal/` source lines that READ
// WorkflowWorktreeConfig.<field>. A line that assigns to the field is not a
// reader, and neither is a comment — the point of the audit is code that acts
// on the value.
// worktreeConfigType is the struct the toggles live on. A value of it can be
// copied into a variable or taken as a parameter, and a read through that copy
// is still a read.
const worktreeConfigType = "WorkflowWorktreeConfig"

// isWorktreeConfigType reports whether an AST type expression names
// WorkflowWorktreeConfig, qualified or not.
func isWorktreeConfigType(expr ast.Expr) bool {
	switch typed := expr.(type) {
	case *ast.Ident:
		return typed.Name == worktreeConfigType
	case *ast.SelectorExpr:
		return typed.Sel != nil && typed.Sel.Name == worktreeConfigType
	case *ast.StarExpr:
		return isWorktreeConfigType(typed.X)
	}
	return false
}

// endsInWorktree reports whether an expression is a selector chain ending in
// `.Worktree`, which is how the config exposes the struct.
func endsInWorktree(expr ast.Expr) bool {
	selector, ok := expr.(*ast.SelectorExpr)
	return ok && selector.Sel != nil && selector.Sel.Name == "Worktree"
}

// worktreeAliases returns the identifiers in one file that hold a
// WorkflowWorktreeConfig value: parameters declared with that type, and
// variables assigned from a `....Worktree` selector.
func worktreeAliases(file *ast.File) map[string]bool {
	aliases := map[string]bool{}
	ast.Inspect(file, func(node ast.Node) bool {
		switch typed := node.(type) {
		case *ast.FuncDecl:
			if typed.Type == nil || typed.Type.Params == nil {
				return true
			}
			for _, field := range typed.Type.Params.List {
				if !isWorktreeConfigType(field.Type) {
					continue
				}
				for _, name := range field.Names {
					aliases[name.Name] = true
				}
			}
		case *ast.AssignStmt:
			for i, rhs := range typed.Rhs {
				if !endsInWorktree(rhs) || i >= len(typed.Lhs) {
					continue
				}
				if ident, ok := typed.Lhs[i].(*ast.Ident); ok {
					aliases[ident.Name] = true
				}
			}
		case *ast.ValueSpec:
			for i, value := range typed.Values {
				if endsInWorktree(value) && i < len(typed.Names) {
					aliases[typed.Names[i].Name] = true
				}
			}
		}
		return true
	})
	return aliases
}

// fieldReadersIn returns `path:line` for every place under root that READS
// WorkflowWorktreeConfig.<field>.
//
// The scan is over the AST rather than the text, because a field name alone
// cannot be trusted: `internal/github` has an unrelated `opts.AutoMerge`, and
// matching on `.AutoMerge` would report the worktree toggle as read when
// nothing reads it. Reads through a copy of the struct count, which a
// `.Worktree.<field>` text match misses (review feedback): a value bound by a
// parameter or an assignment is followed within its file.
//
// What it still does not see is a copy handed across files through something
// other than a parameter of that type, e.g. stored in a field of another
// struct. Resolving that needs a type checker over the whole tree, which is a
// large price for a claim about four keys; a reader added that way would show
// up as an unnamed file in the direct scan of whichever file reads it.
func fieldReadersIn(t *testing.T, root string, field string) []string {
	t.Helper()

	var readers []string
	fset := token.NewFileSet()
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		file, parseErr := parser.ParseFile(fset, path, nil, parser.SkipObjectResolution)
		if parseErr != nil {
			// A file this package cannot parse is not a file it can audit.
			t.Logf("skipping %s: %v", path, parseErr)
			return nil
		}

		aliases := worktreeAliases(file)
		written := map[ast.Node]bool{}
		ast.Inspect(file, func(node ast.Node) bool {
			assign, ok := node.(*ast.AssignStmt)
			if !ok {
				return true
			}
			for _, lhs := range assign.Lhs {
				written[lhs] = true
			}
			return true
		})

		ast.Inspect(file, func(node ast.Node) bool {
			selector, ok := node.(*ast.SelectorExpr)
			if !ok || selector.Sel == nil || selector.Sel.Name != field || written[node] {
				return true
			}
			// `cfg.Workflow.Worktree.AutoCleanup`, or `w.AutoCleanup` where w
			// holds the struct.
			reader := endsInWorktree(selector.X)
			if ident, isIdent := selector.X.(*ast.Ident); isIdent && aliases[ident.Name] {
				reader = true
			}
			if reader {
				position := fset.Position(selector.Pos())
				readers = append(
					readers,
					filepath.ToSlash(path)+":"+strconv.Itoa(position.Line),
				)
			}
			return true
		})
		return nil
	})
	if err != nil {
		t.Fatalf("walking %s: %v", root, err)
	}
	return readers
}

// fieldReaders audits the production tree. The test's own package directory is
// internal/template, so `..` is internal/.
func fieldReaders(t *testing.T, field string) []string {
	t.Helper()
	return fieldReadersIn(t, filepath.Join(".."), field)
}

// TestFieldReadersIn_FollowsCopiesAndIgnoresOtherStructs pins the scan itself
// against a fixture, in both directions. The field name alone cannot decide:
// `internal/github` reads an unrelated `opts.AutoMerge`, and counting that
// would report the worktree toggle as live while nothing reads it. A read
// through a copy of the struct, meanwhile, is a real read that a
// `.Worktree.<field>` text match does not see (review feedback).
func TestFieldReadersIn_FollowsCopiesAndIgnoresOtherStructs(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	write := func(name, body string) {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o600); err != nil {
			t.Fatalf("writing %s: %v", name, err)
		}
	}

	write("direct.go", `package fixture

func direct(cfg Config) bool {
	return cfg.Workflow.Worktree.AutoCleanup
}
`)
	write("copied.go", `package fixture

func copied(cfg Config) bool {
	worktree := cfg.Workflow.Worktree
	return worktree.AutoCleanup
}
`)
	write("parameter.go", `package fixture

func parameter(worktree WorkflowWorktreeConfig) bool {
	return worktree.AutoCleanup
}
`)
	write("other_struct.go", `package fixture

func other(opts MergeOptions) bool {
	return opts.AutoCleanup
}
`)
	write("written.go", `package fixture

func written(cfg *Config) {
	cfg.Workflow.Worktree.AutoCleanup = true
}
`)
	write("commented.go", `package fixture

// cfg.Workflow.Worktree.AutoCleanup is what this used to read.
func commented() bool {
	return false
}
`)

	readers := fieldReadersIn(t, dir, "AutoCleanup")

	files := map[string]int{}
	for _, reader := range readers {
		files[filepath.Base(strings.SplitN(reader, ":", 2)[0])]++
	}
	for _, name := range []string{"direct.go", "copied.go", "parameter.go"} {
		if files[name] != 1 {
			t.Errorf("%s: counted %d read(s), want 1 (readers: %v)", name, files[name], readers)
		}
	}
	for _, name := range []string{"other_struct.go", "written.go", "commented.go"} {
		if files[name] != 0 {
			t.Errorf("%s: counted %d read(s), want 0 (readers: %v)", name, files[name], readers)
		}
	}
}

func TestWorkflowTemplate_WorktreeToggleReaderClaimsAreTrue(t *testing.T) {
	t.Parallel()

	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error: %v", err)
	}
	data, err := fs.ReadFile(fsys, workflowTemplatePath)
	if err != nil {
		t.Fatalf("ReadFile(%s) error: %v", workflowTemplatePath, err)
	}
	sections := commentSections(string(data))

	for key, field := range worktreeToggles {
		readers := fieldReaders(t, field)
		clause, mentioned := sections[key]

		claimedUnread := false
		for _, line := range clause {
			lower := strings.ToLower(line)
			for _, claim := range unreadClaims {
				if strings.Contains(lower, claim) {
					claimedUnread = true
				}
			}
		}

		if len(readers) == 0 {
			if !claimedUnread {
				t.Errorf(
					"%s: no production line reads it, and the shipped %s does not say so. "+
						"An inert key that reads as live invites the user to set it and wait",
					key, workflowTemplatePath,
				)
			}
			// A clause that calls the key unread must not also name a reader
			// file: that pairing is what a stale comment looks like after the
			// last reader is removed (review feedback).
			if stale := namedGoFiles(clause); len(stale) > 0 {
				t.Errorf(
					"%s: the shipped %s calls it unread yet names %v as a reader. "+
						"Nothing there reads the field",
					key, workflowTemplatePath, sortedKeys(stale),
				)
			}
			continue
		}

		if !mentioned {
			t.Errorf(
				"%s: %d production line(s) read it (%v) and the shipped %s has no clause for the key. "+
					"A toggle that acts must say so where the user sets it",
				key, len(readers), readers, workflowTemplatePath,
			)
			continue
		}
		if claimedUnread {
			t.Errorf(
				"%s: the shipped %s calls it unread (%q), but %d production line(s) read it: %v. "+
					"auto_cleanup gates a `git worktree remove`, so a wrong claim here loses work",
				key, workflowTemplatePath, clause, len(readers), readers,
			)
			continue
		}

		// Every reading file has to be named, and every named file has to
		// read: the clause for auto_cleanup names two paths, and losing
		// either one would leave the file describing a reader that is gone
		// while the remaining one keeps the claim technically alive.
		readingFiles := map[string]bool{}
		for _, reader := range readers {
			readingFiles[filepath.Base(strings.SplitN(reader, ":", 2)[0])] = true
		}
		named := namedGoFiles(clause)
		for file := range readingFiles {
			if !named[file] {
				t.Errorf(
					"%s: %s reads the field but the shipped %s clause does not name it (%q). "+
						"The reader list in the comment is what the user checks the key against",
					key, file, workflowTemplatePath, clause,
				)
			}
		}
		for file := range named {
			if !readingFiles[file] {
				t.Errorf(
					"%s: the shipped %s names %s as a reader, but nothing there reads the field. "+
						"Readers found: %v",
					key, workflowTemplatePath, file, readers,
				)
			}
		}
	}
}
