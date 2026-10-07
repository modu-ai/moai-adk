package quality

// gate_node_pm.go — Node package-manager-aware runner selection.
//
// Per-language runner-selection ambiguity classes (survey, card t1551):
//
//	JS/TS — MULTIPLE runners (npm/yarn/pnpm/bun/deno) declared via the
//	    packageManager field plus lockfiles (package-lock.json / yarn.lock /
//	    pnpm-lock.yaml / bun.lock / deno.json). Before this file: entirely
//	    missed — npm/npx was unconditional, so a bun-declared project
//	    received npm invocations (operator report from a consumer project).
//	    Now: a VALID packageManager "bun@<version>" declaration selects the
//	    bun path on every Node axis. Declaration wins; lockfiles are NOT
//	    consulted — a lockfile alone never selects a runner, and a parse
//	    failure or malformed value is never guessed into bun. Every other
//	    declaration (npm/yarn/pnpm/deno/absent/malformed) keeps the existing
//	    npm/npx vectors byte-identical; yarn/pnpm/deno declarations remain a
//	    known, deliberately unguessed gap.
//	Python — pip/poetry/uv/pdm exist as env wrappers, but the gate invokes
//	    the tools themselves (ruff/mypy/pytest) as bare binaries: no wrapper
//	    indirection reaches the gate, so there is no runner ambiguity to
//	    resolve. Single-runner in gate terms.
//	Java/Kotlin — mvn vs gradle is a MARKER ambiguity (pom.xml vs
//	    build.gradle[.kts]), not a declared-runner one: the toolchains
//	    table's Java entry matches build.gradle* and runs `mvn test`
//	    (optional, so a gradle-only machine silently skips it). Pre-existing
//	    class, distinct from package-manager declaration, unchanged here.
//	Go / Rust / Ruby / PHP / Swift / Dart-Flutter / C#-.NET — one canonical
//	    runner each (go / cargo / rspec / phpunit / swift|flutter / dotnet);
//	    no ambiguity class.
//
// The bun arm's contract, per the reviewed design note (t1361 bun-unification):
//   - scripts.test:run present → `bun run test:run`, no appended flags;
//   - only scripts.test → the EXPLICIT script form `bun run test` (never the
//     builtin `bun test`, which would hand a vitest/jest project to bun's own
//     runner), preserving the Vitest --run / Jest --ci non-watch flags and
//     the --passWithNoTests policy;
//   - scripts.lint / scripts.typecheck run via `bun run`, mandatory;
//   - config-gated linters (eslint/biome/oxlint) and tsc resolve ONLY
//     locally-installed binaries from the nearest node_modules/.bin up to
//     (and including) the repository root — no PATH/global fallback exists.
//     A resolved tool execs through the bun runtime (`bun x <tool>`, tool
//     carried in args): the entry carries a node shebang, so a direct exec
//     needs node — absent on bun-only machines (card t1572). An unresolved
//     tool keeps the expected-path exec, so bunx never gets the chance to
//     download from the network;
//   - a missing bun binary or a missing required local binary FAILS the
//     axis rather than skipping it to green;
//   - gate.typecheck.command is honoured verbatim on bun projects too —
//     translating an explicit npm/npx override is a separate contract
//     decision this file does not make.

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// nodePM names the package manager a Node project declared. The zero value
// is npm — the pre-existing behavior every non-bun shape keeps.
type nodePM int

const (
	nodePMNpm nodePM = iota
	nodePMBun
)

// bunRunStepName is the label prefix of every bun-path script step.
const bunRunStepName = "bun run "

// bunTestStepName is the run-summary label of the bun script-form test step.
const bunTestStepName = bunRunStepName + "test"

// nodePMFromDeclaration parses a packageManager field value (corepack form
// `<name>@<version>[+<hash>]`) and reports which runner it declares. Only a
// valid bun declaration selects the bun path; every other shape — other
// names, a missing @version, wrong case, garbage — answers npm, which is the
// "never guess" half of the contract: a malformed declaration is preserved
// behavior, not an inference opportunity.
func nodePMFromDeclaration(decl string) nodePM {
	name, version, ok := strings.Cut(strings.TrimSpace(decl), "@")
	if !ok || name != "bun" {
		return nodePMNpm
	}
	// corepack appends an integrity hash after '+'; it is not part of the
	// version. An empty version (after hash stripping) is malformed.
	if v, _, _ := strings.Cut(version, "+"); strings.TrimSpace(v) == "" {
		return nodePMNpm
	}
	return nodePMBun
}

// nodeManifest carries what the Node resolvers need from a package.json: the
// scripts map and the declared package manager.
type nodeManifest struct {
	scripts map[string]string
	pm      nodePM
}

// readNodeManifest parses the package.json at path. ok is false when the
// file is missing, unreadable, not valid JSON, or the root is not an object.
// A manifest without a scripts key is still ok (nil map reads as empty) —
// the tiers treat "no scripts" as tier (iii), which is what they always did;
// what changed is that a scriptless manifest can still declare bun.
//
// The packageManager field is captured raw and decoded separately: a
// malformed VALUE there (a non-string, say) must degrade to "no declaration"
// (npm), never take the scripts down with it — the scripts-only parser this
// replaced ignored the field entirely, and losing scripts.test:run because
// of a bad declaration would be a regression on valid manifests.
func readNodeManifest(path string) (nodeManifest, bool) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nodeManifest{}, false
	}
	var pkg struct {
		Scripts        map[string]string `json:"scripts"`
		PackageManager json.RawMessage   `json:"packageManager"`
	}
	if err := json.Unmarshal(data, &pkg); err != nil {
		return nodeManifest{}, false
	}
	var decl string
	if len(pkg.PackageManager) > 0 {
		if err := json.Unmarshal(pkg.PackageManager, &decl); err != nil {
			decl = ""
		}
	}
	return nodeManifest{
		scripts: pkg.Scripts,
		pm:      nodePMFromDeclaration(decl),
	}, true
}

// resolveNodeLocalBin finds tool in the nearest node_modules/.bin at or
// above dir, never crossing the repository boundary: the walk checks each
// level's bin directory (a nested package's own first, then the workspace
// root's hoisted one), and stops after the first ancestor carrying .git —
// that root's bin is checked, nothing above it is. Without a .git witness
// anywhere, the walk is bounded only by the filesystem root, where Dir(cur)
// == cur ends it.
//
// Windows note: node_modules/.bin also carries .cmd shims there, which Go's
// exec refuses to launch directly; only the plain name and .exe shapes are
// candidates, so a bun-declared project on Windows may fail a config-gated
// linter step that macOS/Linux resolves — a known limitation of the
// plain/.exe candidate set, not a PATH fallback (none exists on any platform).
func resolveNodeLocalBin(dir, tool string) (string, bool) {
	if dir == "" {
		return "", false
	}
	cur := dir
	for {
		binDir := filepath.Join(cur, "node_modules", ".bin")
		for _, cand := range nodeBinCandidates(tool) {
			p := filepath.Join(binDir, cand)
			if info, err := os.Stat(p); err == nil && !info.IsDir() && isExecutableFile(info) {
				return p, true
			}
		}
		if isRepoRoot(cur) {
			return "", false
		}
		parent := filepath.Dir(cur)
		if parent == cur {
			return "", false
		}
		cur = parent
	}
}

// expectedNodeLocalBin is the path a missing local tool SHOULD have had:
// the toolchain root's own node_modules/.bin. Unresolved tools carry it as
// their binary, so the failed exec names the exact expected location and
// cannot fall back to a PATH or global binary.
func expectedNodeLocalBin(dir, tool string) string {
	return filepath.Join(dir, "node_modules", ".bin", tool)
}

// nodeBinCandidates lists the node_modules/.bin filename shapes a tool can
// take on the current platform.
func nodeBinCandidates(tool string) []string {
	if runtime.GOOS == "windows" {
		return []string{tool + ".exe", tool}
	}
	return []string{tool}
}

// isExecutableFile reports whether a non-directory stat result carries an
// execute bit. Windows permissions are synthesized by the runtime, so every
// existing file counts there.
func isExecutableFile(info os.FileInfo) bool {
	if runtime.GOOS == "windows" {
		return true
	}
	return info.Mode().Perm()&0o111 != 0
}

// isRepoRoot reports whether dir carries a .git entry. A linked worktree's
// .git is a FILE pointing at the real gitdir, so both shapes witness a root.
func isRepoRoot(dir string) bool {
	_, err := os.Stat(filepath.Join(dir, ".git"))
	return err == nil
}

// bunTestStep rewrites the Node test step for a bun-declared project,
// mirroring the npm tiers: (i) scripts.test:run → `bun run test:run` with no
// flags; (ii) watch-prone scripts.test → the runner's non-watch flag
// appended after --passWithNoTests (`--` is stripped by bun and the flag
// forwarded, measured on bun 1.4 — identical to npm's separator semantics);
// (iii) anything else → `bun run test -- --passWithNoTests`. Every tier is
// the explicit script form (`bun run test`), never the builtin `bun test`.
// No tier is optional: a missing bun binary fails the step.
func bunTestStep(scripts map[string]string) gateStep {
	if strings.TrimSpace(scripts[nodeTestRunScript]) != "" {
		return gateStep{
			name:   bunRunStepName + nodeTestRunScript,
			binary: "bun",
			args:   []string{"run", nodeTestRunScript},
		}
	}
	if flag := nodeNonWatchFlag(scripts["test"]); flag != "" {
		return gateStep{
			name:   bunTestStepName + " " + flag,
			binary: "bun",
			args:   append([]string{"run", "test", "--", "--passWithNoTests"}, flag),
		}
	}
	return gateStep{
		name:   bunTestStepName,
		binary: "bun",
		args:   []string{"run", "test", "--", "--passWithNoTests"},
	}
}

// bunConfigLintSteps rewrites the config-gated lint table entries
// (eslint/biome/oxlint) for a bun-declared project: a locally-resolved tool
// execs through the bun runtime with the tool carried in args (the
// node_modules/.bin entry carries a node shebang — a direct exec needs
// node, absent on bun-only machines), an unresolved one keeps the
// expected-path binary, and the step becomes mandatory either way. The
// configFiles guard is preserved verbatim — a project without the linter's
// config still skips the entry; only a project whose config EXISTS but
// whose local binary is absent fails.
func bunConfigLintSteps(steps []gateStep, dir string) []gateStep {
	resolved := make([]gateStep, len(steps))
	for i, step := range steps {
		if step.binary == "npx" && len(step.args) > 0 {
			tool := step.args[0]
			if _, ok := resolveNodeLocalBin(dir, tool); ok {
				// Bun-runtime exec: bun substitutes its own runtime for the
				// entry's node shebang (measured, card t1572).
				step.binary = "bun"
				step.args = append([]string{"x", tool}, step.args[1:]...)
			} else {
				// Unresolved tool keeps the expected-path form: the failed
				// exec names the location, and bunx never downloads.
				step.binary = expectedNodeLocalBin(dir, tool)
				step.args = step.args[1:]
			}
			step.optional = false
		}
		resolved[i] = step
	}
	return resolved
}
