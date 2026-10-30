// user_owned_templates.go — deploy-side protection for the Agentic Factory
// provider templates (issue forms, pull/merge request templates, comment
// templates) that live in a user's own .github/ and .gitlab/ trees.
//
// Update mode (forceUpdate) skips the provenance check on every path except
// the protected scopes (see published_skills.go). The Agentic Factory files sit
// inside directories the project owns, and a user commonly edits them, so a file
// the manifest records user_modified or user_created at one of these paths must
// survive an update. The scope is limited to files whose path names the feature;
// the template-managed reconcile roots (.claude, .moai, .agents) and the rest of
// .github/ and .gitlab/ keep their existing update behavior.
package template

import "strings"

// agenticFactoryPathMarker is the path fragment every Agentic Factory template
// file carries (file names agentic-factory-NN-*, directories agentic-factory/).
const agenticFactoryPathMarker = "agentic-factory"

// isUserOwnedTemplatePath reports whether a deploy-relative path is an
// Agentic Factory provider template under .github/ or .gitlab/.
func isUserOwnedTemplatePath(deployRelPath string) bool {
	if !strings.HasPrefix(deployRelPath, ".github/") && !strings.HasPrefix(deployRelPath, ".gitlab/") {
		return false
	}
	return strings.Contains(deployRelPath, agenticFactoryPathMarker)
}
