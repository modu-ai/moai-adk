package config

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
)

func snapshotConfigTree(t *testing.T, root string) map[string]string {
	t.Helper()
	result := map[string]string{}
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		if entry.IsDir() {
			result[rel] = "directory"
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		result[rel] = string(data)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func TestConfigSaveRefusesExternalSectionParents(t *testing.T) {
	for _, method := range []string{"Save", "saveLLM"} {
		for _, linkRel := range []string{".moai", filepath.Join(".moai", "config"), filepath.Join(".moai", "config", "sections")} {
			for _, present := range []bool{false, true} {
				t.Run(fmt.Sprintf("%s/%s/present=%t", method, linkRel, present), func(t *testing.T) {
					root, outside := t.TempDir(), t.TempDir()
					m := NewConfigManager()
					if _, err := m.LoadRaw(root); err != nil {
						t.Fatal(err)
					}
					llm := m.Get().LLM
					llm.TeamMode = "glm"
					if err := m.SetSection("llm", llm); err != nil {
						t.Fatal(err)
					}
					link := filepath.Join(root, linkRel)
					sections := filepath.Join(root, ".moai", "config", "sections")
					remaining, err := filepath.Rel(link, sections)
					if err != nil {
						t.Fatal(err)
					}
					if present {
						target := filepath.Join(outside, remaining, "llm.yaml")
						if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
							t.Fatal(err)
						}
						if err := os.WriteFile(target, []byte(llmSaveFixture), 0o644); err != nil {
							t.Fatal(err)
						}
					}
					if err := os.MkdirAll(filepath.Dir(link), 0o755); err != nil {
						t.Fatal(err)
					}
					if err := os.Symlink(outside, link); err != nil {
						t.Fatalf("create section-parent link: %v", err)
					}
					before := snapshotConfigTree(t, outside)
					if method == "Save" {
						err = m.Save()
					} else {
						err = m.saveLLM(sections)
					}
					if err == nil {
						t.Error("linked section parent accepted")
					}
					if after := snapshotConfigTree(t, outside); !reflect.DeepEqual(before, after) {
						var changed []string
						for path, value := range after {
							if previous, present := before[path]; !present || previous != value {
								changed = append(changed, path)
							}
						}
						for path := range before {
							if _, present := after[path]; !present {
								changed = append(changed, path)
							}
						}
						sort.Strings(changed)
						t.Errorf("outside directory/payload changed at %v", changed)
					}
					if got, err := os.Readlink(link); err != nil || got != outside {
						t.Errorf("section link changed: %q, %v", got, err)
					}
				})
			}
		}
	}
}

func TestConfigSaveRootedTypedSectionsPreserveBytesAndModes(t *testing.T) {
	for _, alias := range []bool{false, true} {
		t.Run(fmt.Sprintf("trustedAlias=%t", alias), func(t *testing.T) {
			root := t.TempDir()
			project := root
			if alias {
				project = filepath.Join(t.TempDir(), "project")
				if err := os.Symlink(root, project); err != nil {
					t.Fatalf("create trusted root alias: %v", err)
				}
			}
			m := NewConfigManager()
			if _, err := m.LoadRaw(project); err != nil {
				t.Fatal(err)
			}
			cfg := m.Get()
			sections := []struct {
				name, filename string
				value, wrapper any
			}{
				{"user", "user.yaml", cfg.User, userFileWrapper{User: cfg.User}},
				{"language", "language.yaml", cfg.Language, languageFileWrapper{Language: cfg.Language}},
				{"quality", "quality.yaml", cfg.Quality, qualityFileWrapper{Constitution: cfg.Quality}},
				{"git_convention", "git-convention.yaml", cfg.GitConvention, gitConventionFileWrapper{GitConvention: cfg.GitConvention}},
				{"git_strategy", "git-strategy.yaml", cfg.GitStrategy, gitStrategyFileWrapper{GitStrategy: cfg.GitStrategy}},
				{"llm", "llm.yaml", cfg.LLM, llmFileWrapper{LLM: cfg.LLM}},
			}
			reference := t.TempDir()
			for _, section := range sections {
				if err := saveSection(reference, section.filename, section.wrapper); err != nil {
					t.Fatal(err)
				}
			}
			for _, existing := range []bool{false, true} {
				if existing {
					for _, section := range sections {
						if err := m.SetSection(section.name, section.value); err != nil {
							t.Fatal(err)
						}
						for _, dir := range []string{reference, filepath.Join(root, ".moai", "config", "sections")} {
							if err := os.Chmod(filepath.Join(dir, section.filename), 0o640); err != nil {
								t.Fatal(err)
							}
						}
					}
				}
				if err := m.Save(); err != nil {
					t.Fatalf("Save(existing=%t): %v", existing, err)
				}
				for _, section := range sections {
					target := filepath.Join(root, ".moai", "config", "sections", section.filename)
					want, err := os.ReadFile(filepath.Join(reference, section.filename))
					if err != nil {
						t.Fatal(err)
					}
					if got, err := os.ReadFile(target); err != nil || string(got) != string(want) {
						t.Errorf("%s typed bytes differ from generic writer: %v", section.filename, err)
					}
					wantInfo, err := os.Stat(filepath.Join(reference, section.filename))
					if err != nil {
						t.Fatal(err)
					}
					if info, err := os.Stat(target); err != nil || info.Mode().Perm() != wantInfo.Mode().Perm() {
						t.Errorf("%s mode differs from generic writer: %v", section.filename, err)
					}
				}
			}
		})
	}
}
