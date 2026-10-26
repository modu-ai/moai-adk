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
	"reflect"
	"regexp"
	"sort"
	"strings"

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

// clearRefusal is the refusal for a user.yaml the clear cannot edit provably
// safely: the allow-list rejected its structure, or the patched copy failed the
// equivalence check.
func clearRefusal(userPath string) error {
	return fmt.Errorf("clear user name: %s uses YAML structure this command cannot edit safely (anchors, aliases, merge keys, explicit tags, non-string keys, document markers, or several documents); edit user.name by hand", userPath)
}

var docEndMarker = regexp.MustCompile(`(?m)^\.\.\.(\s|#|$)`)

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
	// The YAML parser also reads U+0085 / U+2028 / U+2029 as line breaks, so a `...`
	// after one is a document end the line-start pattern below cannot see, and the
	// comments after it are invisible to the node-level equivalence check (audit
	// F16). Such a file is refused outright.
	// A lone CR (not part of a CRLF pair) is a line break too and slips the
	// line-start pattern (audit F17).
	loneCR := bytes.Contains(bytes.ReplaceAll(data, []byte("\r\n"), nil), []byte("\r"))
	// NUL byte rejection (audit F19): UTF-16 text contains NUL bytes and its LF
	// plus `...` marker slips the byte checks above, so a file with a NUL byte
	// takes this refusal path. This is not a general non-UTF-8 check.
	if bytes.ContainsAny(data, "\u0085\u2028\u2029") || loneCR || bytes.IndexByte(data, 0) >= 0 || docEndMarker.Match(data) {
		// A `...` marker: the re-encode path drops the comments after it.
		return true, false
	}
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
	// plainKeys accepts only plain string keys. yamlpatch finds a key by its raw
	// text while the loader decodes it through its tag, so a tagged key (!!binary,
	// !!int, an explicit !!merge, a custom tag) can point the two at different
	// nodes. The one exception is the plain `<<` merge key inside `user`.
	plainKeys := func(m *yaml.Node, topLevel bool) bool {
		for i := 0; i+1 < len(m.Content); i += 2 {
			k := m.Content[i]
			if k.Kind != yaml.ScalarNode || k.Anchor != "" {
				return false
			}
			switch k.ShortTag() {
			case "!!str":
			case "!!merge":
				// The plain `<<` key, or the explicit `!!merge <<` spelling the
				// re-encoder itself writes (round trip, audit N13), inside user.
				if topLevel || k.Value != "<<" {
					return false
				}
			default:
				return false
			}
		}
		return true
	}
	tagged := func(n *yaml.Node) bool { return n.Style&yaml.TaggedStyle != 0 }
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
	if user == nil || user.Kind != yaml.MappingNode || user.Anchor != "" || tagged(user) || !plainKeys(user, false) {
		return true, false
	}
	name := value(user, "name")
	if name != nil && (name.Kind != yaml.ScalarNode || name.Anchor != "" || tagged(name)) {
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
		return clearRefusal(userPath)
	}
	return patchUserNameVerified(userPath, data)
}

// patchUserNameVerified is the runtime safety net behind the allow-list: it
// patches a temporary copy in the same directory, proves the result equivalent
// to the original (apart from user.name, which must be exactly ""), and only then
// renames the copy over the original. On any failure the original is left
// byte-identical and the temporary copy removed.
func patchUserNameVerified(userPath string, before []byte) error {
	info, err := os.Stat(userPath)
	if err != nil {
		return fmt.Errorf("clear user name: %w", err)
	}
	tmp, err := os.CreateTemp(filepath.Dir(userPath), ".user.yaml.clear-*")
	if err != nil {
		return fmt.Errorf("clear user name: %w", err)
	}
	tmpPath := tmp.Name()
	defer func() { _ = os.Remove(tmpPath) }()
	_, werr := tmp.Write(before)
	if cerr := tmp.Close(); werr == nil {
		werr = cerr
	}
	if werr != nil {
		return fmt.Errorf("clear user name: %w", werr)
	}
	if err := os.Chmod(tmpPath, info.Mode().Perm()); err != nil {
		return fmt.Errorf("clear user name: %w", err)
	}
	if err := yamlpatch.PatchFile(tmpPath, []yamlpatch.KeyEdit{{Path: []string{"user", "name"}, Value: "", Quoted: true}}); err != nil {
		return fmt.Errorf("clear user name: %w", err)
	}
	after, err := os.ReadFile(tmpPath)
	if err != nil {
		return fmt.Errorf("clear user name: %w", err)
	}
	if !clearIsEquivalent(before, after) {
		return clearRefusal(userPath)
	}
	// Re-read right before the rename: if another writer changed the file since
	// `before` was read, renaming the verified copy would revert that change.
	// This narrows the window but cannot eliminate it - the other user.yaml
	// writers take no lock this code could share.
	if current, err := os.ReadFile(userPath); err != nil || !bytes.Equal(current, before) {
		return fmt.Errorf("clear user name: %s changed during the edit; nothing was written, retry", userPath)
	}
	if err := os.Rename(tmpPath, userPath); err != nil {
		return fmt.Errorf("clear user name: %w", err)
	}
	return nil
}

// decodeYAMLStream decodes every document of b into generic values (merge keys
// resolved by the decoder); ok is false when any document fails to decode.
func decodeYAMLStream(b []byte) (docs []any, ok bool) {
	dec := yaml.NewDecoder(bytes.NewReader(b))
	for {
		var v any
		err := dec.Decode(&v)
		if err == io.EOF {
			return docs, true
		}
		if err != nil {
			return nil, false
		}
		docs = append(docs, v)
	}
}

// streamComments collects the head, line and foot comments of every node of
// every document, sorted, so a lost comment (including one after a document
// end marker) changes the set.
func streamComments(b []byte) []string {
	var out []string
	var walk func(n *yaml.Node)
	walk = func(n *yaml.Node) {
		for _, c := range []string{n.HeadComment, n.LineComment, n.FootComment} {
			for _, l := range strings.Split(c, "\n") {
				if l = strings.TrimSpace(l); l != "" {
					out = append(out, l)
				}
			}
		}
		for _, c := range n.Content {
			walk(c)
		}
	}
	dec := yaml.NewDecoder(bytes.NewReader(b))
	for {
		var n yaml.Node
		if dec.Decode(&n) != nil {
			break
		}
		walk(&n)
	}
	sort.Strings(out)
	return out
}

// clearIsEquivalent reports whether after differs from before only in user.name,
// which must decode to the empty string (not null): same decoded values, same
// number of documents, same comments.
func clearIsEquivalent(before, after []byte) bool {
	b, okB := decodeYAMLStream(before)
	a, okA := decodeYAMLStream(after)
	if !okB || !okA || len(b) == 0 || len(b) != len(a) {
		return false
	}
	userOf := func(doc any) map[string]any {
		top, _ := doc.(map[string]any)
		user, _ := top["user"].(map[string]any)
		return user
	}
	bu, au := userOf(b[0]), userOf(a[0])
	if bu == nil || au == nil {
		return false
	}
	if name, present := au["name"]; !present || name != "" {
		return false
	}
	delete(bu, "name")
	delete(au, "name")
	return reflect.DeepEqual(b, a) && reflect.DeepEqual(streamComments(before), streamComments(after))
}
