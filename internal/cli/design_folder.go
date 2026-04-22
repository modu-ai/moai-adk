package cli

import (
	"crypto/sha256"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/modu-ai/moai-adk/internal/template"
)

// designDir is the relative path of the design folder under project root.
const designDir = ".moai/design"

// designTemplateFiles lists the template source paths (relative to embedded FS root)
// and their deployment destination names (without .tmpl suffix) inside designDir.
var designTemplateFiles = []struct {
	src  string // path in embedded FS
	dest string // path under .moai/design/ after rendering
}{
	{"README.md", "README.md"},
	{"research.md.tmpl", "research.md"},
	{"system.md.tmpl", "system.md"},
	{"spec.md.tmpl", "spec.md"},
}

// designSubdirs are the empty subdirectories created under .moai/design/.
var designSubdirs = []string{"wireframes", "screenshots"}

// reservedExact contains exact filenames that are reserved for auto-generated artifacts.
var reservedExact = []string{
	"tokens.json",
	"components.json",
	"import-warnings.json",
}

// reservedGlobs contains filepath.Match patterns (relative to .moai/design/)
// for reserved auto-generated paths.
var reservedGlobs = []string{
	"brief/BRIEF-*.md",
}

// designDirHasRegularFile reports whether dir contains at least one regular (non-hidden) file.
// Hidden files (names starting with ".") and empty subdirectories are ignored.
func designDirHasRegularFile(dir string) (bool, error) {
	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	for _, e := range entries {
		if !e.IsDir() && !strings.HasPrefix(e.Name(), ".") {
			return true, nil
		}
	}
	return false, nil
}

// checkReservedCollision returns an error if any existing user file in projectRoot/.moai/design/
// collides with a reserved filename (exact match or glob match).
func checkReservedCollision(projectRoot string, errOut io.Writer) error {
	base := filepath.Join(projectRoot, designDir)

	for _, name := range reservedExact {
		target := filepath.Join(base, name)
		if _, err := os.Stat(target); err == nil {
			if errOut != nil {
				_, _ = fmt.Fprintf(errOut, "error: reserved filename: %s\n", name)
			}
			return fmt.Errorf("reserved filename: %q collides with reserved name", name)
		}
	}

	for _, pattern := range reservedGlobs {
		// Walk the base dir looking for files matching the pattern
		walkErr := fs.WalkDir(os.DirFS(base), ".", func(path string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return nil
			}
			matched, matchErr := filepath.Match(pattern, path)
			if matchErr != nil {
				return nil
			}
			if matched {
				if errOut != nil {
					_, _ = fmt.Fprintf(errOut, "error: reserved filename: %s\n", path)
				}
				return fmt.Errorf("reserved filename: %q matches reserved pattern %q", path, pattern)
			}
			return nil
		})
		if walkErr != nil {
			return walkErr
		}
	}

	return nil
}

// hashFile computes the SHA-256 hash of the file at path.
func hashFile(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return nil, err
	}
	return h.Sum(nil), nil
}

// hashBytes computes the SHA-256 hash of data.
func hashBytes(data []byte) []byte {
	h := sha256.Sum256(data)
	return h[:]
}

// loadDesignTemplateFS returns the sub-filesystem rooted at ".moai/design" within
// the embedded templates.
func loadDesignTemplateFS() (fs.FS, error) {
	embedded, err := template.EmbeddedTemplates()
	if err != nil {
		return nil, fmt.Errorf("load embedded templates: %w", err)
	}
	sub, err := fs.Sub(embedded, ".moai/design")
	if err != nil {
		return nil, fmt.Errorf("sub .moai/design: %w", err)
	}
	return sub, nil
}

// scaffoldDesignDir deploys .moai/design/ templates to projectRoot.
//
// Returns (true, nil) on successful deployment, (false, nil) when the directory
// is non-empty and deployment is skipped, or (false, err) on I/O error.
// Warnings are written to warnOut (e.g. stderr or a strings.Builder in tests).
//
// REQ-001, REQ-002, REQ-003, REQ-010.
func scaffoldDesignDir(projectRoot string, warnOut io.Writer) (bool, error) {
	base := filepath.Join(projectRoot, designDir)

	// REQ-010: Skip if .moai/design/ already has a regular file.
	hasFile, err := designDirHasRegularFile(base)
	if err != nil {
		return false, fmt.Errorf("check design dir: %w", err)
	}
	if hasFile {
		if warnOut != nil {
			_, _ = fmt.Fprintf(warnOut, "warning: .moai/design/ already contains files — skip template deployment\n")
		}
		return false, nil
	}

	// Ensure base directory exists.
	if err := os.MkdirAll(base, 0o755); err != nil {
		return false, fmt.Errorf("mkdir .moai/design/: %w", err)
	}

	// REQ-003: Create empty subdirectories with .gitkeep.
	for _, sub := range designSubdirs {
		subPath := filepath.Join(base, sub)
		if err := os.MkdirAll(subPath, 0o755); err != nil {
			return false, fmt.Errorf("mkdir %s: %w", sub, err)
		}
		gitkeep := filepath.Join(subPath, ".gitkeep")
		if _, err := os.Stat(gitkeep); os.IsNotExist(err) {
			if err := os.WriteFile(gitkeep, []byte{}, 0o644); err != nil {
				return false, fmt.Errorf("create .gitkeep in %s: %w", sub, err)
			}
		}
	}

	// REQ-002: Deploy template files.
	designFS, err := loadDesignTemplateFS()
	if err != nil {
		return false, err
	}

	for _, tf := range designTemplateFiles {
		destPath := filepath.Join(base, tf.dest)

		// Skip .gitkeep files — they come from subdirs already handled above.
		if strings.HasSuffix(tf.src, ".gitkeep") {
			continue
		}

		content, err := fs.ReadFile(designFS, tf.src)
		if err != nil {
			return false, fmt.Errorf("read template %s: %w", tf.src, err)
		}

		if err := os.WriteFile(destPath, content, 0o644); err != nil {
			return false, fmt.Errorf("write %s: %w", tf.dest, err)
		}
	}

	return true, nil
}

// updateDesignDir syncs .moai/design/ template files during `moai update`.
//
// Rules:
//   - REQ-005: Files whose on-disk content differs from the canonical template
//     (SHA-256 mismatch) are treated as user-modified and are NOT overwritten.
//   - REQ-008: If any user file collides with a reserved filename, the function
//     returns a non-zero error and writes "reserved filename" to errOut.
//
// Returns nil on success, error on reserved filename collision or I/O failure.
func updateDesignDir(projectRoot string, errOut io.Writer) error {
	base := filepath.Join(projectRoot, designDir)

	// REQ-008: Reject reserved filename collisions before touching anything.
	if err := checkReservedCollision(projectRoot, errOut); err != nil {
		return err
	}

	// Load embedded design templates.
	designFS, err := loadDesignTemplateFS()
	if err != nil {
		return err
	}

	for _, tf := range designTemplateFiles {
		if strings.HasSuffix(tf.src, ".gitkeep") {
			continue
		}

		canonical, err := fs.ReadFile(designFS, tf.src)
		if err != nil {
			return fmt.Errorf("read canonical template %s: %w", tf.src, err)
		}

		canonicalHash := hashBytes(canonical)
		destPath := filepath.Join(base, tf.dest)

		// If the file does not exist yet, deploy it.
		if _, statErr := os.Stat(destPath); os.IsNotExist(statErr) {
			_ = os.MkdirAll(filepath.Dir(destPath), 0o755)
			if err := os.WriteFile(destPath, canonical, 0o644); err != nil {
				return fmt.Errorf("write %s: %w", tf.dest, err)
			}
			continue
		}

		// REQ-005: Compare on-disk hash with canonical hash.
		diskHash, err := hashFile(destPath)
		if err != nil {
			return fmt.Errorf("hash %s: %w", destPath, err)
		}

		if string(diskHash) == string(canonicalHash) {
			// File matches canonical template — safe to overwrite with latest.
			if err := os.WriteFile(destPath, canonical, 0o644); err != nil {
				return fmt.Errorf("update %s: %w", tf.dest, err)
			}
		}
		// Hash differs → user modified → skip (preserve user edit).
	}

	// REQ-003: Ensure subdirectories exist.
	for _, sub := range designSubdirs {
		subPath := filepath.Join(base, sub)
		if err := os.MkdirAll(subPath, 0o755); err != nil {
			return fmt.Errorf("mkdir %s: %w", sub, err)
		}
		gitkeep := filepath.Join(subPath, ".gitkeep")
		if _, err := os.Stat(gitkeep); os.IsNotExist(err) {
			if err := os.WriteFile(gitkeep, []byte{}, 0o644); err != nil {
				return fmt.Errorf("create .gitkeep in %s: %w", sub, err)
			}
		}
	}

	return nil
}
