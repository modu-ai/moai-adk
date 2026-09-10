// Stdlib-only premise check for card t582: mirrors the classifier's decision
// sequence (IsAbs -> backslash -> "~/" peel) and the expansion
// filepath.Join(home, p[2:]) without compiling internal/cli.
package main

import (
	"fmt"
	"path/filepath"
	"strings"
)

func main() {
	const home = "/Users/example"
	for _, p := range []string{
		"~/ok/SKILL.md",     // control: clean home-relative
		"~/../../etc/x",     // defect arm: dotdot escape
		`~/x\SKILL.md`,      // control: backslash (t571 guard)
		"/etc/x",            // control: absolute
		"~/a/../b/SKILL.md", // inner dotdot that stays inside home
		"~/..",              // escape to the parent of home
	} {
		shape := "relative"
		switch {
		case filepath.IsAbs(p):
			shape = "absolute"
		case strings.ContainsRune(p, '\\'):
			shape = "oddly-formed"
		case p == "~" || strings.HasPrefix(p, "~/"):
			shape = "home-relative"
		}
		expanded, inside := "-", "-"
		if shape == "home-relative" {
			expanded = filepath.Join(home, p[2:])
			rel, err := filepath.Rel(home, expanded)
			inside = fmt.Sprint(err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)))
		}
		fmt.Printf("%-22q shape=%-14s expanded=%-24s inside_home=%s\n", p, shape, expanded, inside)
	}
}
