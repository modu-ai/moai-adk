package ui

import (
	"testing"
)

// === Framework options full coverage ===

func TestFrameworkOptions_AllLanguages(t *testing.T) {
	languages := []string{"Go", "Python", "TypeScript", "Java", "Rust", "PHP", "UnknownLang"}
	for _, lang := range languages {
		t.Run(lang, func(t *testing.T) {
			opts := frameworkOptions(lang)
			if len(opts) == 0 {
				t.Errorf("expected non-empty options for %s", lang)
			}
		})
	}
}

func TestFrameworkOptions_TypeScript(t *testing.T) {
	opts := frameworkOptions("TypeScript")
	found := false
	for _, o := range opts {
		if o.Value == "Next.js" {
			found = true
		}
	}
	if !found {
		t.Error("expected Next.js in TypeScript frameworks")
	}
}

func TestFrameworkOptions_Java(t *testing.T) {
	opts := frameworkOptions("Java")
	found := false
	for _, o := range opts {
		if o.Value == "Spring Boot" {
			found = true
		}
	}
	if !found {
		t.Error("expected Spring Boot in Java frameworks")
	}
}

func TestFrameworkOptions_Rust(t *testing.T) {
	opts := frameworkOptions("Rust")
	found := false
	for _, o := range opts {
		if o.Value == "Axum" {
			found = true
		}
	}
	if !found {
		t.Error("expected Axum in Rust frameworks")
	}
}

func TestFrameworkOptions_PHP(t *testing.T) {
	opts := frameworkOptions("PHP")
	found := false
	for _, o := range opts {
		if o.Value == "Laravel" {
			found = true
		}
	}
	if !found {
		t.Error("expected Laravel in PHP frameworks")
	}
}
