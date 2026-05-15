package harness

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// HasWebFrontend detects if a project has a web frontend by checking:
// 1. package.json with frontend framework dependencies (react, vue, next, etc.)
// 2. index.html in project root
func HasWebFrontend(projectRoot string) bool {
	// Check for package.json
	pkgJSONPath := filepath.Join(projectRoot, "package.json")
	if data, err := os.ReadFile(pkgJSONPath); err == nil {
		var pkgJSON struct {
			Dependencies map[string]string `json:"dependencies"`
		}
		if err := json.Unmarshal(data, &pkgJSON); err == nil {
			for dep := range pkgJSON.Dependencies {
				if isWebFramework(dep) {
					return true
				}
			}
		}
	}

	// Check for index.html
	indexHTMLPath := filepath.Join(projectRoot, "index.html")
	if _, err := os.Stat(indexHTMLPath); err == nil {
		return true
	}

	return false
}

// isWebFramework checks if a dependency name is a web frontend framework
func isWebFramework(dep string) bool {
	dep = strings.ToLower(dep)
	webFrameworks := []string{
		"react", "vue", "angular", "svelte", "solid",
		"next", "nuxt", "gatsby", "remix",
		"react-dom", "vue-router", "@angular/core",
		"sveltekit", "@sveltejs/kit",
	}

	for _, framework := range webFrameworks {
		if strings.HasPrefix(dep, framework) {
			return true
		}
	}
	return false
}
