package merge

// base.go derives the base side of the per-file 3-way merge that runs after a
// template deploy.
//
// A 3-way merge needs three inputs, and only two of them are ever in hand at
// this point: the user's file (current) and the freshly deployed template
// (updated). The third — the template content the user's file was originally
// deployed from — is not stored anywhere. The manifest records a hash per file,
// which can tell whether the template changed but cannot reconstruct what it
// used to say.
//
// Feeding the current template in as the base is not a workable stand-in, and it
// fails in the one direction that matters. With base and updated byte-identical,
// every key the user's file happens not to carry reads as a deletion the user
// made, and the merge dutifully preserves it — so a key the template newly
// introduces can never reach a project that already has the file. That is how a
// new default (a new MCP server entry, a new settings key) silently failed to
// arrive on every existing install while landing correctly on fresh ones.
//
// What is derivable is a base that is *consistent with* both sides: the deployed
// template narrowed to the keys the user's file actually has. Read as a 3-way
// input it says "the user started from a template that carried exactly these
// keys", which is true of every key it mentions and silent about the rest. That
// leaves the merge free to treat a key only the template has as an addition
// rather than a deletion, while a key only the user has stays an addition of
// theirs, and a value the user edited still reads as their change.
//
// The one thing this base cannot express is a value change the template made to
// a key the user never touched: with base and updated agreeing on that value,
// the template's edit is invisible and the user's copy stands. That limitation
// is inherited, not introduced — the previous base had it too, for the same
// reason — and closing it needs the deployed template content to be snapshotted
// at deploy time so a genuine base exists on the next update.

import (
	"encoding/json"
	"io/fs"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// templateManaged reports whether path is a file the embedded templates ship,
// counting the .tmpl form because a rendered file lands at the path without its
// suffix.
//
// This is the discriminator between a file the deploy just replaced and a file
// the user created themselves. Merging template content into the latter would
// inject keys it was never deployed from, so only a template-managed path is
// eligible for the derived base.
func templateManaged(embedded fs.FS, path string) bool {
	for _, candidate := range []string{path, path + ".tmpl"} {
		if _, err := fs.Stat(embedded, candidate); err == nil {
			return true
		}
	}
	return false
}

// deriveTemplateBase returns the merge base for path, and reports whether one
// could be derived at all.
//
// Derivation needs a structured format whose merge strategy compares keys —
// JSON and YAML. Anything else (a rendered shell script, plain text) has no key
// structure to narrow, so it reports false and the caller preserves the user's
// content, which is what it did before for these files.
func deriveTemplateBase(path string, current, updated []byte) ([]byte, bool) {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".json":
		return derive(current, updated, json.Unmarshal, func(v map[string]any) ([]byte, error) {
			return json.MarshalIndent(v, "", "  ")
		})
	case ".yaml", ".yml":
		return derive(current, updated, yaml.Unmarshal, func(v map[string]any) ([]byte, error) {
			return yaml.Marshal(v)
		})
	default:
		return nil, false
	}
}

// derive narrows updated to the keys current also carries, using the supplied
// codec. A side that does not parse yields no base rather than a wrong one.
func derive(
	current, updated []byte,
	unmarshal func([]byte, any) error,
	marshal func(map[string]any) ([]byte, error),
) ([]byte, bool) {
	var currentMap, updatedMap map[string]any
	if unmarshal(current, &currentMap) != nil || currentMap == nil {
		return nil, false
	}
	if unmarshal(updated, &updatedMap) != nil || updatedMap == nil {
		return nil, false
	}
	data, err := marshal(pruneToShared(updatedMap, currentMap))
	if err != nil {
		return nil, false
	}
	return data, true
}

// pruneToShared returns updated with every key absent from current removed,
// recursing wherever both sides hold a nested map so the narrowing reaches keys
// like mcpServers.moai rather than stopping at the container.
//
// Values always come from updated. A key present on both sides therefore enters
// the base carrying the template's value, which is what makes a user's edit to
// that key read as their change during the merge.
func pruneToShared(updated, current map[string]any) map[string]any {
	pruned := make(map[string]any, len(updated))
	for key, updatedVal := range updated {
		currentVal, shared := current[key]
		if !shared {
			// A key the user's file does not carry stays out of the base, so the
			// merge sees the template introducing it rather than the user having
			// removed it.
			continue
		}
		updatedChild, updatedIsMap := updatedVal.(map[string]any)
		currentChild, currentIsMap := currentVal.(map[string]any)
		if updatedIsMap && currentIsMap {
			pruned[key] = pruneToShared(updatedChild, currentChild)
			continue
		}
		pruned[key] = updatedVal
	}
	return pruned
}
