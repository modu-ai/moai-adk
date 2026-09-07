// emit.go — publication orchestration.
//
// Emit reads every command source under opts.CommandsRoot, parses each into
// the neutral form, derives the moai-<command> skill identity, and renders
// the deterministic SKILL.md set under opts.EmittedRoot. The command
// sources are never written; the returned publication is the only output.
package commandemit

import (
	"bytes"
	"fmt"
	"io/fs"
	"sort"
	"strings"
)

// Emit produces the publication of the command set under opts.CommandsRoot.
// On any validation error it returns (nil, err) — no partial artifact set.
// Collision policy is refuse-and-report: a derived skill name matching an
// existing canonical skill directory name under opts.SkillsRoot aborts the
// run naming every offending pair; it is never resolved by suffixing,
// renaming, or overwriting.
func Emit(fsys fs.FS, opts Options) (*Publication, error) {
	var files []string
	walkErr := fs.WalkDir(fsys, opts.CommandsRoot, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if strings.HasSuffix(p, ".md") || strings.HasSuffix(p, ".md.tmpl") {
			files = append(files, p)
		}
		return nil
	})
	if walkErr != nil {
		return nil, fmt.Errorf("commandemit: walk %s: %w", opts.CommandsRoot, walkErr)
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("commandemit: no command sources under %s", opts.CommandsRoot)
	}
	sort.Strings(files)

	docs := make([]CommandDoc, 0, len(files))
	for _, file := range files {
		data, err := fs.ReadFile(fsys, file)
		if err != nil {
			return nil, fmt.Errorf("commandemit: read %s: %w", file, err)
		}
		doc, err := ParseCommandDoc(file, data)
		if err != nil {
			return nil, err
		}
		docs = append(docs, doc)
	}

	if err := checkCollisions(fsys, opts, docs); err != nil {
		return nil, err
	}

	pub := &Publication{Skills: make(map[string][]byte, len(docs))}
	for _, doc := range docs {
		emittedPath := opts.EmittedRoot + "/" + doc.SkillName + "/SKILL.md"
		pub.Skills[emittedPath] = renderSkill(doc)
		pub.Report = append(pub.Report, SkillReport{
			Command:      doc.Command,
			Skill:        doc.SkillName,
			Path:         emittedPath,
			Description:  doc.Description,
			BoundaryFlag: BoundaryFlag,
		})
	}
	return pub, nil
}

// checkCollisions refuses emission when any derived skill name matches an
// existing canonical skill directory name. All collisions are gathered and
// reported in one diagnostic — one error per run, never a silent rename.
func checkCollisions(fsys fs.FS, opts Options, docs []CommandDoc) error {
	existing, err := canonicalSkillNames(fsys, opts.SkillsRoot)
	if err != nil {
		return err
	}
	var collisions []string
	for _, doc := range docs {
		if _, hit := existing[doc.SkillName]; hit {
			collisions = append(collisions, fmt.Sprintf(
				"%s: derived skill name %q collides with existing canonical skill directory %q — refusing to emit (no suffixing, no renaming, no overwriting)",
				doc.File, doc.SkillName, doc.SkillName))
		}
	}
	if len(collisions) > 0 {
		return fmt.Errorf("commandemit: skill-name collision(s), no artifacts produced:\n%s",
			strings.Join(collisions, "\n"))
	}
	return nil
}

// canonicalSkillNames collects the directory names under skillsRoot.
func canonicalSkillNames(fsys fs.FS, skillsRoot string) (map[string]struct{}, error) {
	entries, err := fs.ReadDir(fsys, skillsRoot)
	if err != nil {
		return nil, fmt.Errorf("commandemit: read canonical skills root %s: %w", skillsRoot, err)
	}
	names := make(map[string]struct{}, len(entries))
	for _, e := range entries {
		if e.IsDir() {
			names[e.Name()] = struct{}{}
		}
	}
	return names, nil
}

// renderSkill renders one SKILL.md: the YAML frontmatter (generated-header
// comment, name, quoted description) followed by the verbatim body bytes.
// The body starts with whatever followed the source's closing delimiter —
// typically a blank line — so the published layout is conventional
// frontmatter + blank line + body while the body bytes stay identical.
// @MX:ANCHOR: [AUTO] Sole rendering path for published command skills; the golden drift check and the byte-equality ACs all judge its output.
// @MX:REASON: [AUTO] The body half must stay byte-identical to the command source (verbatim-provenance contract); any transform here is a provenance violation.
func renderSkill(doc CommandDoc) []byte {
	var b bytes.Buffer
	b.WriteString("---\n")
	b.WriteString("# " + GeneratedHeader + "\n")
	b.WriteString("name: " + doc.SkillName + "\n")
	b.WriteString("description: " + quoteYAMLScalar(doc.Description) + "\n")
	b.WriteString("---\n")
	b.Write(doc.Body)
	return b.Bytes()
}
