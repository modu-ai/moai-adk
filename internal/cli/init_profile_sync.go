package cli

// init_profile_sync.go — the profile synchronization a completed init runs after its
// template deploy, shared by the healthy tail (init.go) and the resume dispatch
// (init_resume.go) so the two paths cannot drift (card t1591, Item D).

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"

	"github.com/modu-ai/moai-adk/internal/cli/wizard"
	"github.com/modu-ai/moai-adk/internal/profile"
	"github.com/modu-ai/moai-adk/internal/settings/yamlpatch"
)

// syncInitProfile writes the profile preferences into the project config
// (.moai/config/sections/user.yaml and language.yaml). The wizard's answers for the
// user name and the conversation language take precedence over the profile values,
// the same precedence the template deploy gives opts. A nil wizardResult, or an empty
// answer that is not an explicit clear, leaves the profile value in place.
func syncInitProfile(projectRoot string, prefs profile.ProfilePreferences, wizardResult *wizard.WizardResult) error {
	clearName := false
	if wizardResult != nil {
		if wizardResult.ConversationLang != "" {
			prefs.ConversationLang = wizardResult.ConversationLang
		}
		if wizardResult.UserName != "" {
			prefs.UserName = wizardResult.UserName
		} else if wizardResult.UserNameCleared {
			// Card t1966: an explicit clear must not re-save the profile name.
			prefs.UserName = ""
			clearName = true
		}
	}
	if err := profile.SyncToProjectConfig(projectRoot, prefs); err != nil {
		return err
	}
	if clearName {
		return clearProjectUserName(projectRoot)
	}
	return nil
}

// userNameEditSafety is the allow-list behind the clear: it reports whether the
// file parses (parsed) and, if so, whether yamlpatch is guaranteed to write the
// very node the loader reads and to rewrite nothing else (safe). The patch
// re-encodes the first document only and matches keys by their Value text, so
// the file must be exactly one document, a top-level mapping with no alias,
// anchored or top-level merge KEY; `user` must be a direct plain mapping value
// (no alias, no anchor) whose keys are not aliases or anchored; and `name`, when
// present, a direct plain scalar. A `<<` merge inside `user` stays allowed: the
// patch then only adds or blanks the explicit name row (matrix rows B04-B08).
func userNameEditSafety(data []byte) (parsed, safe bool) {
	dec := yaml.NewDecoder(bytes.NewReader(data))
	var first yaml.Node
	if dec.Decode(&first) != nil {
		return false, false
	}
	var extra yaml.Node
	if err := dec.Decode(&extra); err != io.EOF {
		return true, false // several documents (or a broken later one)
	}
	top := &first
	if top.Kind == yaml.DocumentNode && len(top.Content) > 0 {
		top = top.Content[0]
	}
	if top.Kind != yaml.MappingNode {
		return true, false
	}
	plainKeys := func(m *yaml.Node, topLevel bool) bool {
		for i := 0; i+1 < len(m.Content); i += 2 {
			k := m.Content[i]
			if k.Kind != yaml.ScalarNode || k.Anchor != "" || (topLevel && k.ShortTag() == "!!merge") {
				return false
			}
		}
		return true
	}
	value := func(m *yaml.Node, key string) *yaml.Node {
		for i := 0; i+1 < len(m.Content); i += 2 {
			if m.Content[i].Value == key {
				return m.Content[i+1]
			}
		}
		return nil
	}
	if !plainKeys(top, true) {
		return true, false
	}
	user := value(top, "user")
	if user == nil || user.Kind != yaml.MappingNode || user.Anchor != "" || !plainKeys(user, false) {
		return true, false
	}
	name := value(user, "name")
	if name != nil && (name.Kind != yaml.ScalarNode || name.Anchor != "") {
		return true, false
	}
	return true, true
}

// clearProjectUserName blanks user.name in the project's user.yaml. SyncToProjectConfig
// only writes a non-empty name, and a resumed init redeploys no template, so a name
// the interrupted run left behind would survive an explicit clear (card t1966, F1).
// A missing user.yaml has nothing to clear.
func clearProjectUserName(projectRoot string) error {
	userPath := filepath.Join(projectRoot, ".moai", "config", "sections", "user.yaml")
	data, err := os.ReadFile(userPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("clear user name: %w", err)
	}
	// Skip unless a non-empty user.name exists: a zero-byte file, a missing
	// user mapping or name key, or an already-empty name has nothing to clear,
	// and patching would error (empty document) or invent a null `name:` row.
	// A file that does not parse is NOT skipped: it falls through to the patch,
	// which reports the error (or clears the name), so a malformed file never
	// silently keeps a name the user cleared.
	var doc struct {
		User struct {
			Name string `yaml:"name"`
		} `yaml:"user"`
	}
	if yaml.Unmarshal(data, &doc) == nil && doc.User.Name == "" {
		return nil
	}
	// yamlpatch does not resolve aliases and edits the anchor node itself, so a
	// user or user.name node that is an alias or carries an anchor is refused
	// untouched (card t1966, F8). A file that does not parse has no node tree to
	// inspect and falls through to the patch as above.
	if parsed, safe := userNameEditSafety(data); parsed && !safe {
		return fmt.Errorf("clear user name: %s uses YAML structure this command cannot edit safely (anchors, aliases, merge keys, or several documents); edit user.name by hand", userPath)
	}
	if err := yamlpatch.PatchFile(userPath, []yamlpatch.KeyEdit{{Path: []string{"user", "name"}, Value: ""}}); err != nil {
		return fmt.Errorf("clear user name: %w", err)
	}
	return nil
}
