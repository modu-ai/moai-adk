package quality

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// typecheckStepName is the axis name, deliberately language-independent: the
// gate reports "typecheck" whether the underlying tool is tsc, mypy, or a
// project's own script. It doubles as the disabled_steps key.
const typecheckStepName = "typecheck"

// nodeTypecheckScript is the package.json script the tier-(b) lookup honours.
const nodeTypecheckScript = "typecheck"

// nodeTypecheckStep is the Node placeholder resolved by resolveTypecheckStep.
//
// It carries no command of its own because which command is right depends on
// the project: a script if the author declared one, otherwise tsc. Resolution
// happens against the project directory at run time.
func nodeTypecheckStep() *gateStep {
	return &gateStep{name: typecheckStepName, binary: "", args: nil}
}

// resolveTypecheckStep decides what — if anything — to run for the typecheck
// axis, in three tiers.
//
//	(a) override:  gate.typecheck.command, honoured for ANY language
//	(b) script:    package.json scripts.typecheck  -> npm run typecheck
//	(c) tsconfig:  tsconfig.json present           -> npx --no-install tsc --noEmit
//
// The third return reports whether a step was produced. When it is false the
// second return explains why, and that explanation is surfaced rather than
// dropped: a silent skip is the exact failure this axis repairs, where a
// consumer repository's gate passed while its build was broken.
//
// A skip is not a failure. Tool availability stays fail-open — the point is
// that the operator can always see which axes actually ran.
func resolveTypecheckStep(base *gateStep, dir, override string) (gateStep, string, bool) {
	// Tier (a): an explicit command outranks everything and works for any
	// language, which is how a Python or Go project opts into the axis.
	if cmd := strings.TrimSpace(override); cmd != "" {
		fields := strings.Fields(cmd)
		if len(fields) == 0 {
			return gateStep{}, "typecheck: skipped (gate.typecheck.command is empty)", false
		}
		return gateStep{
			name:     typecheckStepName,
			binary:   fields[0],
			args:     fields[1:],
			optional: true,
		}, "", true
	}

	if base == nil {
		return gateStep{}, fmt.Sprintf(
			"%s: skipped (no default for this language; set gate.typecheck.command to enable one)",
			typecheckStepName), false
	}

	if dir == "" {
		return gateStep{}, fmt.Sprintf("%s: skipped (project directory unknown)", typecheckStepName), false
	}

	// Tier (b): the project's own script. It outranks the tsconfig shape check
	// below, so a monorepo root that delegates to turbo is not penalised for
	// having a solution-style tsconfig.
	if scripts, ok := readPackageJSONScripts(filepath.Join(dir, "package.json")); ok {
		if strings.TrimSpace(scripts[nodeTypecheckScript]) != "" {
			return gateStep{
				name:     typecheckStepName,
				binary:   "npm",
				args:     []string{"run", nodeTypecheckScript},
				optional: true,
			}, "", true
		}
	}

	// Tier (c): tsc against the project's tsconfig.
	tsconfig := filepath.Join(dir, "tsconfig.json")
	if _, err := os.Stat(tsconfig); err != nil {
		return gateStep{}, fmt.Sprintf(
			"%s: skipped (no scripts.typecheck and no tsconfig.json; set gate.typecheck.command to enable one)",
			typecheckStepName), false
	}

	if isSolutionStyleTsconfig(tsconfig) {
		return gateStep{}, fmt.Sprintf(
			"%s: skipped (solution-style tsconfig type-checks nothing and would pass vacuously; "+
				"add a scripts.typecheck that builds the referenced projects, or set gate.typecheck.command)",
			typecheckStepName), false
	}

	return gateStep{
		name:     typecheckStepName,
		binary:   "npx",
		args:     []string{"--no-install", "tsc", "--noEmit"},
		optional: true,
	}, "", true
}

// isSolutionStyleTsconfig reports whether the tsconfig is a solution file: an
// empty files array plus project references, and nothing to compile itself.
//
// Running tsc --noEmit against one exits 0 having checked nothing, so treating
// it as coverage would rebuild the very blind spot this axis closes. A parse
// failure answers false: refusing to run on an unreadable config would be a
// worse failure mode than attempting the check.
//
// tsconfig.json is JSONC, not JSON — TypeScript accepts comments and trailing
// commas, and real configs use them. encoding/json rejects both, and because a
// parse failure answers false, an unstripped comment would send a solution
// config down the "run tsc" path and straight back into the vacuous pass this
// function exists to prevent. Comments are stripped first.
func isSolutionStyleTsconfig(path string) bool {
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}

	var cfg struct {
		Files      *[]string         `json:"files"`
		Include    *[]string         `json:"include"`
		References []json.RawMessage `json:"references"`
	}
	if err := json.Unmarshal(stripJSONC(data), &cfg); err != nil {
		return false
	}

	if len(cfg.References) == 0 {
		return false
	}
	// An include list means the config does compile something of its own.
	if cfg.Include != nil && len(*cfg.Include) > 0 {
		return false
	}
	// files: [] is the solution marker. An absent files key with references and
	// no include also compiles nothing under the default exclusions in
	// practice, but only the explicit empty array is treated as conclusive.
	return cfg.Files != nil && len(*cfg.Files) == 0
}

// appendReason joins non-empty step notices with a newline.
//
// Notices accumulate rather than overwrite: before this, a later step's reason
// replaced an earlier one, so a run that skipped two axes reported one.
func appendReason(existing, add string) string {
	add = strings.TrimSpace(add)
	if add == "" {
		return existing
	}
	if strings.TrimSpace(existing) == "" {
		return add
	}
	return existing + "\n" + add
}
