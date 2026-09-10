// Command ts measures what filepath.ToSlash / FromSlash actually do on the
// host running it, for card t540's plan-audit finding D1.
//
// Not part of the build: it lives under .moai/reports/, outside any package
// the module compiles.
package main

import (
	"fmt"
	"path/filepath"
	"runtime"
)

func main() {
	back := `C:\Users\u\.agents\skills\foo\SKILL.md`
	slash := "C:/Users/u/.agents/skills/foo/SKILL.md"
	fmt.Printf("GOOS=%s Separator=%q\n", runtime.GOOS, filepath.Separator)
	fmt.Printf("ToSlash(backslash)  = %q changed=%v\n", filepath.ToSlash(back), filepath.ToSlash(back) != back)
	fmt.Printf("FromSlash(slash)    = %q changed=%v\n", filepath.FromSlash(slash), filepath.FromSlash(slash) != slash)
	fmt.Printf("IsAbs(slash form)   = %v\n", filepath.IsAbs(slash))
	fmt.Printf("IsAbs(backslash)    = %v\n", filepath.IsAbs(back))
}
