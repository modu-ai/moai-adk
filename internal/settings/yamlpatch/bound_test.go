package yamlpatch

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBoundPatchRejectsDirectoryLeafBeforeRead(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "section.yaml")
	if err := os.Mkdir(path, 0o755); err != nil {
		t.Fatal(err)
	}
	err := PatchFileAtRoot(root, "section.yaml", []KeyEdit{{Path: []string{"a"}, Value: "new"}}, RootOptions{})
	if err == nil || !strings.Contains(err.Error(), "yamlpatch: stat") {
		t.Fatalf("directory leaf = %v; want refusal before read", err)
	}
}

func TestBoundFullDocumentPreservesPayloadAndPublicationPolicy(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "config", "section.yaml")
	data := []byte("# typed section\nsection:\n  values: [one, two]\n  empty: {}\n")
	if err := WriteFileAtRoot(root, filepath.Join("config", "section.yaml"), data, RootOptions{ParentPerm: 0o755}); err != nil {
		t.Fatal(err)
	}
	if got, err := os.ReadFile(path); err != nil || string(got) != string(data) {
		t.Fatalf("full-document bytes = %q, %v", got, err)
	}
	outside := filepath.Join(t.TempDir(), "outside.yaml")
	if err := os.WriteFile(outside, []byte("user bytes"), 0o644); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(root, "linked.yaml")
	if err := os.Symlink(outside, link); err != nil {
		t.Fatalf("create full-document linked fixture: %v", err)
	}
	if err := WriteFileAtRoot(root, "linked.yaml", data, RootOptions{}); err == nil || !strings.Contains(err.Error(), "yamlpatch: stat") {
		t.Fatalf("linked full-document target = %v; want stat refusal", err)
	}
	if got, err := os.ReadFile(outside); err != nil || string(got) != "user bytes" {
		t.Fatalf("outside full-document payload changed: %q, %v", got, err)
	}
	if err := WriteFileAtRoot(root, filepath.Join("missing", "section.yaml"), data, RootOptions{}); err == nil || !strings.Contains(err.Error(), "yamlpatch: create temp") {
		t.Fatalf("missing full-document parent = %v; want create-temp error", err)
	}
}

func TestBoundPatchRefusesLinkedComponents(t *testing.T) {
	for _, linkRel := range []string{".moai", filepath.Join(".moai", "config"), filepath.Join(".moai", "config", "sections")} {
		for _, present := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/present=%t", linkRel, present), func(t *testing.T) {
				root, outside := t.TempDir(), t.TempDir()
				rel := filepath.Join(".moai", "config", "sections", "section.yaml")
				link := filepath.Join(root, linkRel)
				targetDir := filepath.Dir(filepath.Join(root, rel))
				tail, err := filepath.Rel(link, targetDir)
				if err != nil {
					t.Fatal(err)
				}
				outsideDir := filepath.Join(outside, tail)
				if err := os.MkdirAll(outsideDir, 0o755); err != nil {
					t.Fatal(err)
				}
				outsideFile := filepath.Join(outsideDir, "section.yaml")
				const original = "a: original\n"
				if present {
					if err := os.WriteFile(outsideFile, []byte(original), 0o644); err != nil {
						t.Fatal(err)
					}
				}
				if err := os.MkdirAll(filepath.Dir(link), 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(outside, link); err != nil {
					t.Fatalf("create section ancestor link: %v", err)
				}
				err = PatchFileAtRoot(root, rel, []KeyEdit{{Path: []string{"a"}, Value: "updated"}}, RootOptions{ParentPerm: 0o755})
				if err == nil || !strings.Contains(err.Error(), "yamlpatch: stat") {
					t.Errorf("linked component = %v; want stat refusal", err)
				}
				if got, err := os.Readlink(link); err != nil || got != outside {
					t.Errorf("section ancestor link changed: %q, %v", got, err)
				}
				if present {
					if got, err := os.ReadFile(outsideFile); err != nil || string(got) != original {
						t.Errorf("outside section changed: %q, %v", got, err)
					}
				} else if _, err := os.Lstat(outsideFile); !os.IsNotExist(err) {
					t.Errorf("missing outside section was created: %v", err)
				}
				entries, err := os.ReadDir(outsideDir)
				want := 0
				if present {
					want = 1
				}
				if err != nil || len(entries) != want {
					t.Errorf("outside directory entries = %v, %v; want %d unchanged entries", entries, err, want)
				}
			})
		}
	}
}

func TestBoundPatchAcceptsTrustedRootAlias(t *testing.T) {
	for _, present := range []bool{false, true} {
		t.Run(fmt.Sprintf("present=%t", present), func(t *testing.T) {
			root := t.TempDir()
			alias := filepath.Join(t.TempDir(), "project-alias")
			if err := os.Symlink(root, alias); err != nil {
				t.Fatalf("create trusted root alias: %v", err)
			}
			relative := filepath.Join(".moai", "config", "sections", "section.yaml")
			target := filepath.Join(root, relative)
			if present {
				if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(target, []byte("# keep\na: old\nunknown: preserved\n"), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			if err := PatchFileAtRoot(alias, relative, []KeyEdit{{Path: []string{"a"}, Value: "new"}}, RootOptions{ParentPerm: 0o755}); err != nil {
				t.Fatalf("patch trusted root alias: %v", err)
			}
			got, err := os.ReadFile(target)
			want := "a: new\n"
			if present {
				want = "# keep\na: new\nunknown: preserved\n"
			}
			if err != nil || string(got) != want {
				t.Fatalf("canonical target = %q, %v; want %q", got, err, want)
			}
		})
	}
}

func TestGenericPatchAcceptsOwnedParentAlias(t *testing.T) {
	root := t.TempDir()
	alias := filepath.Join(t.TempDir(), "owned-parent")
	if err := os.Symlink(root, alias); err != nil {
		t.Fatalf("create owned parent alias: %v", err)
	}
	path := filepath.Join(alias, "section.yaml")
	for _, value := range []string{"first", "second"} {
		if err := PatchFile(path, []KeyEdit{{Path: []string{"a"}, Value: value}}); err != nil {
			t.Fatalf("generic owned path: %v", err)
		}
		if got, err := os.ReadFile(filepath.Join(root, "section.yaml")); err != nil || string(got) != "a: "+value+"\n" {
			t.Fatalf("owned alias target = %q, %v", got, err)
		}
	}
}

func TestBoundPatchPreservesGenericModeAndEmptyDocumentContracts(t *testing.T) {
	for _, tc := range []struct {
		name    string
		seed    string
		present bool
		options RootOptions
	}{
		{name: "greenfield"},
		{name: "preserve existing mode and presentation", present: true, seed: "# keep\n\na: old # note\nunknown: kept\n"},
		{name: "explicit mode and blank document", present: true, seed: "# keep\n", options: RootOptions{Mode: 0o600, AllowEmpty: true}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			generic := filepath.Join(t.TempDir(), "generic.yaml")
			bound := filepath.Join(root, "section.yaml")
			if tc.present {
				for _, path := range []string{generic, bound} {
					if err := os.WriteFile(path, []byte(tc.seed), 0o640); err != nil {
						t.Fatal(err)
					}
				}
			}
			edits := []KeyEdit{{Path: []string{"a"}, Value: "new"}}
			var err error
			if tc.options.AllowEmpty {
				err = PatchFileWithMode(generic, edits, tc.options.Mode)
			} else {
				err = PatchFile(generic, edits)
			}
			if err != nil {
				t.Fatalf("reference generic write: %v", err)
			}
			if err := PatchFileAtRoot(root, "section.yaml", edits, tc.options); err != nil {
				t.Fatal(err)
			}
			want, err := os.ReadFile(generic)
			if err != nil {
				t.Fatal(err)
			}
			got, err := os.ReadFile(bound)
			if err != nil || string(got) != string(want) {
				t.Fatalf("bound bytes = %q, %v; generic bytes %q", got, err, want)
			}
			wantInfo, err := os.Stat(generic)
			if err != nil {
				t.Fatal(err)
			}
			gotInfo, err := os.Stat(bound)
			if err != nil || gotInfo.Mode().Perm() != wantInfo.Mode().Perm() {
				t.Fatalf("bound mode differs from generic mode %o: %v", wantInfo.Mode().Perm(), err)
			}
		})
	}
}

func TestBoundPatchErrorBoundariesAndCleanup(t *testing.T) {
	root := t.TempDir()
	edits := []KeyEdit{{Path: []string{"a"}, Value: "new"}}
	for _, path := range []string{"../outside.yaml", filepath.Join(root, "absolute.yaml"), "."} {
		if err := PatchFileAtRoot(root, path, edits, RootOptions{}); err == nil || !strings.Contains(err.Error(), "yamlpatch: stat") {
			t.Errorf("non-local file %q = %v; want stat refusal", path, err)
		}
	}
	if err := os.WriteFile(filepath.Join(root, "regular"), []byte("user bytes"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := PatchFileAtRoot(root, filepath.Join("regular", "section.yaml"), edits, RootOptions{ParentPerm: 0o755}); err == nil || !strings.Contains(err.Error(), "yamlpatch: stat") {
		t.Errorf("regular parent = %v; want stat refusal", err)
	}
	if err := PatchFileAtRoot(root, filepath.Join("missing", "section.yaml"), edits, RootOptions{}); err == nil || !strings.Contains(err.Error(), "yamlpatch: create temp") {
		t.Errorf("missing parent = %v; want create-temp error", err)
	}
	for _, target := range []string{filepath.Join(root, "regular"), filepath.Join(root, "dangling")} {
		name := filepath.Join(root, "linked.yaml")
		if err := os.Symlink(target, name); err != nil {
			t.Fatalf("create leaf link: %v", err)
		}
		if err := PatchFileAtRoot(root, "linked.yaml", edits, RootOptions{}); err == nil || !strings.Contains(err.Error(), "yamlpatch: stat") {
			t.Errorf("linked leaf = %v; want stat refusal", err)
		}
		if err := os.Remove(name); err != nil {
			t.Fatal(err)
		}
	}
	if got, err := os.ReadFile(filepath.Join(root, "regular")); err != nil || string(got) != "user bytes" {
		t.Fatalf("regular parent payload changed: %q, %v", got, err)
	}
	if err := PatchFileAtRoot(filepath.Join(root, "absent-root"), "section.yaml", nil, RootOptions{ParentPerm: 0o755}); err != nil {
		t.Fatalf("empty edit set touched absent root: %v", err)
	}
	if err := os.Mkdir(filepath.Join(root, "directory.yaml"), 0o755); err != nil {
		t.Fatal(err)
	}
	r, err := os.OpenRoot(root)
	if err != nil {
		t.Fatal(err)
	}
	err = atomicWriteAtRoot(r, "directory.yaml", []byte("a: new\n"), 0)
	if closeErr := r.Close(); closeErr != nil {
		t.Errorf("close fixture root: %v", closeErr)
	}
	if err == nil || !strings.Contains(err.Error(), "yamlpatch: rename") {
		t.Fatalf("directory replacement = %v; want rename failure", err)
	}
	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".yamlpatch-") || strings.HasPrefix(entry.Name(), ".moai-config-") {
			t.Errorf("failed publication left temp %q", entry.Name())
		}
	}
}
