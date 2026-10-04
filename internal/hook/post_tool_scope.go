package hook

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/modu-ai/moai-adk/internal/config"
)

// postToolTargetOutsideProject reports whether the Write/Edit/MultiEdit target
// lies outside the project, so the lint (Quality Gate) and security scans skip
// it: a throwaway script under a session scratchpad or /tmp is not project code
// and its findings are noise (card t1499).
//
// "Inside" means inside CLAUDE_PROJECT_DIR OR inside the session cwd. The cwd
// counts because a lane works in a worktree whose root is not the primary
// checkout CLAUDE_PROJECT_DIR names. Both sides are compared after
// symlink resolution, so a spelling (macOS /tmp versus /private/tmp) never
// decides scope.
//
// It fails open: with no root known, no file_path, or an unparseable payload it
// reports false and the scans run exactly as before.
func postToolTargetOutsideProject(input *HookInput) bool {
	if input == nil {
		return false
	}
	var ti struct {
		FilePath string `json:"file_path"`
	}
	if err := json.Unmarshal(input.ToolInput, &ti); err != nil || ti.FilePath == "" {
		return false
	}

	var roots []string
	for _, r := range []string{os.Getenv(config.EnvClaudeProjectDir), input.CWD} {
		if r != "" {
			roots = append(roots, r)
		}
	}
	if len(roots) == 0 {
		return false
	}

	target := ti.FilePath
	if !filepath.IsAbs(target) {
		target = filepath.Join(roots[0], target)
	}
	realTarget := evalSymlinksExistingPrefix(target)
	for _, r := range roots {
		if pathWithin(evalSymlinksExistingPrefix(r), realTarget) {
			return false
		}
	}
	return true
}

// evalSymlinksExistingPrefix resolves symlinks in path. A path that does not
// exist yet (a file being created) is resolved through its deepest existing
// ancestor, with the missing tail appended unchanged.
func evalSymlinksExistingPrefix(path string) string {
	path = filepath.Clean(path)
	var tail []string
	for p := path; ; {
		if real, err := filepath.EvalSymlinks(p); err == nil {
			for i := len(tail) - 1; i >= 0; i-- {
				real = filepath.Join(real, tail[i])
			}
			return real
		}
		parent := filepath.Dir(p)
		if parent == p {
			return path
		}
		tail = append(tail, filepath.Base(p))
		p = parent
	}
}

// pathWithin reports whether target equals root or sits beneath it. Both are
// cleaned absolute paths.
func pathWithin(root, target string) bool {
	rel, err := filepath.Rel(root, target)
	if err != nil {
		return false
	}
	return rel == "." || (rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) && !filepath.IsAbs(rel))
}
