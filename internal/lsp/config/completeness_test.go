package config_test

import (
	"sort"
	"testing"

	"github.com/modu-ai/moai-adk/internal/lsp/config"
)

// canonicalLanguages is the exact 16-language list from REQ-LM-006.
// Matches .claude/skills/moai/workflows/sync.md Phase 0.6.1 Language Detection table.
// "flutter" is the canonical name for the Dart/Flutter ecosystem.
var canonicalLanguages = []string{
	"cpp", "csharp", "elixir", "flutter", "go", "java",
	"javascript", "kotlin", "php", "python", "r", "ruby",
	"rust", "scala", "swift", "typescript",
}

// canonicalLanguageSet is a set version for quick lookup.
var canonicalLanguageSet = func() map[string]bool {
	m := make(map[string]bool, len(canonicalLanguages))
	for _, l := range canonicalLanguages {
		m[l] = true
	}
	return m
}()

// sampleAllLanguagesYAML is the REQ-LM-006 compliance fixture.
// It contains all 16 canonical languages with install_hint, project_markers,
// and fallback_binaries as required by REQ-LM-004/001/008.
const sampleAllLanguagesYAML = `
lsp:
  servers:
    cpp:
      command: clangd
      fallback_binaries: []
      args: []
      project_markers:
        - CMakeLists.txt
        - compile_commands.json
      install_hint: "Install via LLVM distribution"
      file_extensions:
        - .cpp
        - .cc
        - .cxx
        - .h
        - .hpp
    csharp:
      command: omnisharp
      fallback_binaries:
        - roslyn-ls
      args:
        - -lsp
      project_markers:
        - "*.csproj"
        - "*.sln"
      install_hint: "Install OmniSharp"
      file_extensions:
        - .cs
    elixir:
      command: elixir-ls
      fallback_binaries:
        - lexical
      args: []
      project_markers:
        - mix.exs
      install_hint: "Install elixir-ls"
      file_extensions:
        - .ex
        - .exs
    flutter:
      command: dart
      fallback_binaries: []
      args:
        - language-server
        - --protocol=lsp
      project_markers:
        - pubspec.yaml
      install_hint: "Install Dart SDK or Flutter SDK"
      file_extensions:
        - .dart
    go:
      command: gopls
      fallback_binaries: []
      args: []
      project_markers:
        - go.mod
        - go.sum
      install_hint: "go install golang.org/x/tools/gopls@latest"
      file_extensions:
        - .go
    java:
      command: jdtls
      fallback_binaries: []
      args: []
      project_markers:
        - pom.xml
        - build.gradle
      install_hint: "Install Eclipse JDT Language Server"
      file_extensions:
        - .java
    javascript:
      command: typescript-language-server
      fallback_binaries: []
      args:
        - --stdio
      project_markers:
        - package.json
        - jsconfig.json
      install_hint: "npm install -g typescript-language-server typescript"
      file_extensions:
        - .js
        - .jsx
        - .mjs
    kotlin:
      command: kotlin-language-server
      fallback_binaries: []
      args: []
      project_markers:
        - build.gradle.kts
        - build.gradle
      install_hint: "Install kotlin-language-server"
      file_extensions:
        - .kt
        - .kts
    php:
      command: phpactor
      fallback_binaries:
        - intelephense
      args:
        - language-server
      project_markers:
        - composer.json
      install_hint: "Install phpactor"
      file_extensions:
        - .php
    python:
      command: pylsp
      fallback_binaries:
        - pyright-langserver
        - basedpyright-langserver
      args: []
      project_markers:
        - pyproject.toml
        - requirements.txt
      install_hint: "pip install python-lsp-server"
      file_extensions:
        - .py
    r:
      command: R
      fallback_binaries: []
      args:
        - --slave
        - -e
        - "languageserver::run()"
      project_markers:
        - DESCRIPTION
        - renv.lock
      install_hint: "install.packages('languageserver') in R"
      file_extensions:
        - .R
        - .r
    ruby:
      command: ruby-lsp
      fallback_binaries:
        - solargraph
      args: []
      project_markers:
        - Gemfile
        - .ruby-version
      install_hint: "gem install ruby-lsp"
      file_extensions:
        - .rb
    rust:
      command: rust-analyzer
      fallback_binaries: []
      args: []
      project_markers:
        - Cargo.toml
        - Cargo.lock
      install_hint: "rustup component add rust-analyzer"
      file_extensions:
        - .rs
    scala:
      command: metals
      fallback_binaries: []
      args: []
      project_markers:
        - build.sbt
        - build.sc
      install_hint: "Install metals"
      file_extensions:
        - .scala
        - .sc
    swift:
      command: sourcekit-lsp
      fallback_binaries: []
      args: []
      project_markers:
        - Package.swift
      install_hint: "Install Xcode Command Line Tools"
      file_extensions:
        - .swift
    typescript:
      command: typescript-language-server
      fallback_binaries: []
      args:
        - --stdio
      project_markers:
        - tsconfig.json
        - package.json
      install_hint: "npm install -g typescript-language-server typescript"
      file_extensions:
        - .ts
        - .tsx
`

// Load16LangConfig is a test helper that calls config.Load and returns (cfg, err).
func Load16LangConfig(t *testing.T, path string) (*config.ServersConfig, error) {
	t.Helper()
	return config.Load(path)
}

// TestLoad_All16LanguagesPresent verifies the YAML fixture contains all 16 canonical
// languages as required by REQ-LM-006.
func TestLoad_All16LanguagesPresent(t *testing.T) {
	t.Parallel()

	path := writeTempYAML(t, sampleAllLanguagesYAML)
	cfg, err := Load16LangConfig(t, path)
	if err != nil {
		t.Fatalf("Load error = %v", err)
	}

	if len(cfg.Servers) != 16 {
		t.Errorf("server count = %d, want 16", len(cfg.Servers))
	}

	for _, lang := range canonicalLanguages {
		if _, ok := cfg.Servers[lang]; !ok {
			t.Errorf("missing canonical language %q from servers config", lang)
		}
	}
}

// TestLoad_NoLanguageMissingInstallHint verifies every server entry has a non-empty
// InstallHint (REQ-LM-004).
func TestLoad_NoLanguageMissingInstallHint(t *testing.T) {
	t.Parallel()

	path := writeTempYAML(t, sampleAllLanguagesYAML)
	cfg, err := Load16LangConfig(t, path)
	if err != nil {
		t.Fatalf("Load error = %v", err)
	}

	for _, lang := range canonicalLanguages {
		sc, ok := cfg.Servers[lang]
		if !ok {
			continue // already reported by TestLoad_All16LanguagesPresent
		}
		if sc.InstallHint == "" {
			t.Errorf("language %q missing install_hint (REQ-LM-004)", lang)
		}
	}
}

// TestLoad_NoLanguageMissingProjectMarkers verifies every server entry has at least
// one ProjectMarker (REQ-LM-001).
func TestLoad_NoLanguageMissingProjectMarkers(t *testing.T) {
	t.Parallel()

	path := writeTempYAML(t, sampleAllLanguagesYAML)
	cfg, err := Load16LangConfig(t, path)
	if err != nil {
		t.Fatalf("Load error = %v", err)
	}

	for _, lang := range canonicalLanguages {
		sc, ok := cfg.Servers[lang]
		if !ok {
			continue
		}
		if len(sc.ProjectMarkers) == 0 {
			t.Errorf("language %q has no project_markers (REQ-LM-001)", lang)
		}
	}
}

// TestLoad_CanonicalLanguageListSorted verifies the 16-language list is exactly as specified
// in REQ-LM-006 (alphabetical order, no extras, no missing).
func TestLoad_CanonicalLanguageListSorted(t *testing.T) {
	t.Parallel()

	sorted := make([]string, len(canonicalLanguages))
	copy(sorted, canonicalLanguages)
	sort.Strings(sorted)

	for i, lang := range sorted {
		if canonicalLanguages[i] != lang {
			t.Errorf("canonicalLanguages[%d] = %q, want %q (list must be alphabetical)", i, canonicalLanguages[i], lang)
		}
	}

	if len(canonicalLanguages) != 16 {
		t.Errorf("canonical language count = %d, want 16", len(canonicalLanguages))
	}
}

// TestLoad_LanguageFieldInjectedFromKey verifies the Language field is injected from the
// YAML map key (not required in YAML body).
func TestLoad_LanguageFieldInjectedFromKey(t *testing.T) {
	t.Parallel()

	path := writeTempYAML(t, sampleAllLanguagesYAML)
	cfg, err := Load16LangConfig(t, path)
	if err != nil {
		t.Fatalf("Load error = %v", err)
	}

	for _, lang := range canonicalLanguages {
		sc, ok := cfg.Servers[lang]
		if !ok {
			continue
		}
		if sc.Language != lang {
			t.Errorf("language %q: Language field = %q, want %q (should be injected from map key)", lang, sc.Language, lang)
		}
	}
}
