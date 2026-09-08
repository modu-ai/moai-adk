package spec

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// zz_probe2_test.go — t528: separate DISCRIMINATOR drift from CORPUS drift.
//
// Run A used declRe_A (id must end in a digit) and reported 1160 declarations.
// Run B used declRe_B (dots allowed, id may end in a letter) and reported 1167.
// Those two numbers are NOT comparable. This probe runs both discriminators
// over the same tree in the same process, so the difference is attributable.

var declRe_A = regexp.MustCompile(`^\s*[-*+]\s+\*{0,2}(AC-[A-Za-z0-9-]*[0-9])\*{0,2}\s*(.*)$`)
var declRe_B = regexp.MustCompile(`^\s*[-*+]\s+\*{0,2}(AC-[A-Za-z0-9.-]*[A-Za-z0-9])\*{0,2}\s*(.*)$`)

func TestT528DiscriminatorDrift(t *testing.T) {
	root := "../../.moai/specs"
	ownCard := "SPEC-AC-COLLECTOR-ANCHOR-001"

	count := func(re *regexp.Regexp, excludeOwn bool) (files, decls int) {
		_ = filepath.Walk(root, func(p string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() || info.Name() != "spec.md" {
				return nil
			}
			if excludeOwn && strings.Contains(p, ownCard) {
				return nil
			}
			b, e := os.ReadFile(p)
			if e != nil {
				return nil
			}
			files++
			lines := strings.Split(string(b), "\n")
			start := findACSectionStart(lines)
			if start < 0 {
				return nil
			}
			for i := start; i < len(lines); i++ {
				if strings.HasPrefix(strings.TrimSpace(lines[i]), "##") {
					break
				}
				if re.MatchString(lines[i]) {
					decls++
				}
			}
			return nil
		})
		return
	}

	fA, dA := count(declRe_A, false)
	fB, dB := count(declRe_B, false)
	fAx, dAx := count(declRe_A, true)
	fBx, dBx := count(declRe_B, true)

	t.Logf("discriminator A (digit-final), own card INCLUDED : files=%d decls=%d", fA, dA)
	t.Logf("discriminator B (letter/dot ok), own card INCLUDED: files=%d decls=%d", fB, dB)
	t.Logf("discriminator A, own card EXCLUDED               : files=%d decls=%d", fAx, dAx)
	t.Logf("discriminator B, own card EXCLUDED               : files=%d decls=%d", fBx, dBx)
	t.Logf("=> corpus effect of own card (A): files %+d decls %+d", fA-fAx, dA-dAx)
	t.Logf("=> discriminator effect A->B (own excluded): decls %+d", dBx-dAx)
}

// TestT528HeadingFalsePositive — findACSectionStart matches ANY "##"-prefixed
// heading whose lowercased text contains "acceptance", including a heading that
// merely names the FILE acceptance.md, and it returns the FIRST such match.
func TestT528HeadingFalsePositive(t *testing.T) {
	root := "../../.moai/specs"
	var subHeading, fileMention int
	var samples []string
	_ = filepath.Walk(root, func(p string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || info.Name() != "spec.md" {
			return nil
		}
		b, e := os.ReadFile(p)
		if e != nil {
			return nil
		}
		lines := strings.Split(string(b), "\n")
		start := findACSectionStart(lines)
		if start < 1 {
			return nil
		}
		h := strings.TrimSpace(lines[start-1])
		if !strings.HasPrefix(h, "## ") {
			subHeading++ // anchored on ### or deeper
		}
		if strings.Contains(strings.ToLower(h), "acceptance.md") {
			fileMention++
			if len(samples) < 5 {
				samples = append(samples, p+"  ->  "+h)
			}
		}
		return nil
	})
	t.Logf("AC section anchored on a ###-or-deeper heading = %d", subHeading)
	t.Logf("AC section anchored on a heading naming the FILE acceptance.md = %d", fileMention)
	for _, s := range samples {
		t.Logf("  %s", s)
	}
}
