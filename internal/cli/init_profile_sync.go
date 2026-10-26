package cli

// init_profile_sync.go — the profile synchronization a completed init runs after its
// template deploy, shared by the healthy tail (init.go) and the resume dispatch
// (init_resume.go) so the two paths cannot drift (card t1591, Item D).

import (
	"fmt"
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

// userNameUsesAnchor reports whether the `user` node or its `name` node is an
// alias or carries an anchor.
func userNameUsesAnchor(doc *yaml.Node) bool {
	top := doc
	if top.Kind == yaml.DocumentNode && len(top.Content) > 0 {
		top = top.Content[0]
	}
	child := func(m *yaml.Node, key string) *yaml.Node {
		if m.Kind != yaml.MappingNode {
			return nil
		}
		for i := 0; i+1 < len(m.Content); i += 2 {
			if m.Content[i].Value == key {
				return m.Content[i+1]
			}
		}
		return nil
	}
	user := child(top, "user")
	if user == nil {
		return false
	}
	if user.Kind == yaml.AliasNode || user.Anchor != "" {
		return true
	}
	name := child(user, "name")
	return name != nil && (name.Kind == yaml.AliasNode || name.Anchor != "")
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
	var root yaml.Node
	if yaml.Unmarshal(data, &root) == nil && userNameUsesAnchor(&root) {
		return fmt.Errorf("clear user name: %s uses YAML anchors/aliases; edit user.name by hand", userPath)
	}
	if err := yamlpatch.PatchFile(userPath, []yamlpatch.KeyEdit{{Path: []string{"user", "name"}, Value: ""}}); err != nil {
		return fmt.Errorf("clear user name: %w", err)
	}
	return nil
}
