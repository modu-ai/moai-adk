// emit.go — dual-publication orchestration.
//
// EmitAll reads every .md under agentsRoot, parses each into the neutral
// form, validates it against the mapping manifest (fail-closed), and renders
// the deterministic Codex TOML set. The .md publication is identity: source
// bytes pass through untouched (Option A — the emitter never re-renders the
// neutral layer, so the moai update regression ban holds by construction).
package agentemit

import (
	"fmt"
	"io/fs"
	"path"
	"sort"
	"strings"
)

// EmitAll produces the dual publication of the agent set under agentsRoot.
// On any validation error it returns (nil, err) — no partial artifact set.
func EmitAll(fsys fs.FS, agentsRoot string, man Manifest) (*Publication, error) {
	var files []string
	walkErr := fs.WalkDir(fsys, agentsRoot, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if strings.HasSuffix(p, ".md") {
			files = append(files, p)
		}
		return nil
	})
	if walkErr != nil {
		return nil, fmt.Errorf("agentemit: walk %s: %w", agentsRoot, walkErr)
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("agentemit: no .md sources under %s", agentsRoot)
	}
	sort.Strings(files)

	type artifact struct {
		path string
		data []byte
	}
	var tomls []artifact
	seenName := map[string]string{}
	md := map[string][]byte{}

	for _, file := range files {
		data, err := fs.ReadFile(fsys, file)
		if err != nil {
			return nil, fmt.Errorf("agentemit: read %s: %w", file, err)
		}
		doc, err := ParseAgentDoc(file, data)
		if err != nil {
			return nil, err
		}
		if prev, dup := seenName[doc.Name]; dup {
			return nil, fmt.Errorf("%s: duplicate agent name %q (also declared by %s) — Codex namespace collision", file, doc.Name, prev)
		}
		seenName[doc.Name] = file

		hasMCP := false
		for _, tok := range doc.Tools {
			class, ok := classifyToken(man, tok)
			if !ok {
				return nil, fmt.Errorf("%s: unknown tool token %q — not mapped to any class in the Codex mapping manifest", file, tok)
			}
			if class == "moai-mcp" {
				hasMCP = true
			}
		}

		tomlData, err := renderTOML(doc, man, hasMCP)
		if err != nil {
			return nil, err
		}
		tomls = append(tomls, artifact{path: codexTOMLPath(man, file), data: tomlData})

		// Identity pass-through: the published .md bytes ARE the source bytes.
		md[file] = data
	}

	pub := &Publication{Markdown: md, CodexTOML: make(map[string][]byte, len(tomls))}
	for _, a := range tomls {
		pub.CodexTOML[a.path] = a.data
	}
	return pub, nil
}

// codexTOMLPath resolves the emitted TOML path for one .md source according
// to the manifest layout knob (subdirectory preferred; flat + prefix
// fallback). Paths are forward-slash relative, deploy-root anchored.
func codexTOMLPath(man Manifest, mdPath string) string {
	stem := strings.TrimSuffix(path.Base(mdPath), ".md")
	switch man.Layout.Mode {
	case "flat_prefix":
		return ".codex/agents/" + man.Layout.FlatPrefix + stem + ".toml"
	default: // subdirectory
		return ".codex/agents/" + man.Layout.Subdirectory + "/" + stem + ".toml"
	}
}
