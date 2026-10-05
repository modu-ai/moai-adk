package cli

// Seam-bypass guard (card t775, follows t581's identity and USE guards).
//
// t581 pinned two things about the config-path seam: that fromConfigPath is the
// identity under a '/' separator, and that the '/'-pinned absolute arm is
// reached and hands the declared form to stat. Both compare VALUES.
//
// A bypass that copies the seam's body into the call site instead of calling it
//
//	if configPathSeparator == '/' {
//	    statPath = e.Path
//	} else {
//	    statPath = strings.ReplaceAll(e.Path, "/", string(configPathSeparator))
//	}
//
// produces the same value and runs the same arm, so BOTH t581 guards pass. That
// was measured, not reasoned: the mutant was injected at doctor_codex.go's stat
// site and the whole seam suite stayed green (.moai/reports/t775/probe.md).
//
// The conclusion is structural rather than incidental. A seam exists to force a
// single passage; "did this go through the seam?" is not a question about the
// resulting value, and no value comparison can answer it. So this guard reads
// the SHAPE of the source instead: outside the seam's own file, no production
// code in this package may branch on configPathSeparator == '/'. That is the
// copied body's signature, and it cannot be written without it.
//
// Chosen over a call-observing seam (a fromConfigPathFn indirection a test can
// count, the shape cc.go:20 and codex_contract.go:52 use): that would catch the
// same bypass, but only by adding test-only wiring to production code, and only
// for callers that go through the variable — a copied body skips the variable
// too. Source shape has neither cost.
//
// Deliberate limits: this guard is scoped to the ONE seam t581 covers, per the
// card's over-design warning. A bypass spelled differently (a helper named
// something else, a switch, a lookup table) is not caught — the guard blocks the
// copy-the-body recurrence, not every conceivable route around the seam.

import (
	"go/ast"
	"go/token"
	"strings"
	"testing"
)

// seamOwnerFile is the one file allowed to branch on the separator: the seam
// itself lives there, and its body is exactly the shape this guard refuses
// everywhere else.
const seamOwnerFile = "codex_config_path.go"

// separatorSlashBranch reports whether expr compares configPathSeparator to the
// rune literal '/', in either operand order and with either == or !=.
func separatorSlashBranch(expr ast.Expr) bool {
	bin, ok := expr.(*ast.BinaryExpr)
	if !ok || (bin.Op != token.EQL && bin.Op != token.NEQ) {
		return false
	}
	// Any identifier, not just configPathSeparator: a copied body is as likely to
	// compare a local (sep, s, hostSep) as the package global, and naming the
	// global only would make the guard trivially evadable by renaming. An indexed
	// byte (path[i] == '/', harness_validate.go) is not an identifier and is not
	// this shape, so it stays out.
	isSep := func(e ast.Expr) bool {
		_, ok := e.(*ast.Ident)
		return ok
	}
	isSlashRune := func(e ast.Expr) bool {
		lit, ok := e.(*ast.BasicLit)
		return ok && lit.Kind == token.CHAR && lit.Value == `'/'`
	}
	return (isSep(bin.X) && isSlashRune(bin.Y)) || (isSep(bin.Y) && isSlashRune(bin.X))
}

// separatorBranchFiles returns the non-test files that carry such a branch.
func separatorBranchFiles(files []parsedFile) []string {
	var found []string
	for _, pf := range files {
		hit := false
		ast.Inspect(pf.ast, func(n ast.Node) bool {
			if n == nil || hit {
				return !hit
			}
			if expr, ok := n.(ast.Expr); ok && separatorSlashBranch(expr) {
				hit = true
				return false
			}
			return true
		})
		if hit {
			found = append(found, pf.name)
		}
	}
	return found
}

func TestConfigPathSeamIsNotBypassedByACopiedBody(t *testing.T) {
	nonTest, _ := parseDirFiles(t, ".")

	// Premise assertion first: the detector must actually fire on the seam's own
	// file. Without this, a detector that silently matches nothing would report a
	// clean package forever — the vacuous-guard failure this card is about.
	owner := false
	var offenders []string
	for _, name := range separatorBranchFiles(nonTest) {
		if name == seamOwnerFile {
			owner = true
			continue
		}
		offenders = append(offenders, name)
	}
	if !owner {
		t.Fatalf("the detector found no configPathSeparator == '/' branch in %s — "+
			"either the seam moved or this guard stopped matching; it proves nothing until it does",
			seamOwnerFile)
	}

	if len(offenders) > 0 {
		t.Errorf("production files branch on <separator> == '/' outside %s: %v — "+
			"that is the seam's body copied to the call site, which passes every value-comparing "+
			"guard; call fromConfigPath/toConfigPath instead", seamOwnerFile, strings.Join(offenders, ", "))
	}
}
