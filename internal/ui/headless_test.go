package ui

import (
	"testing"
)

func TestNewHeadlessManager_NotHeadlessByDefault(t *testing.T) {
	hm := NewHeadlessManager()
	// In test environment, we don't know TTY state, so just test the override path.
	// The default depends on os.Stdin; we test ForceHeadless instead.
	if hm == nil {
		t.Fatal("NewHeadlessManager returned nil")
	}
}

func TestHeadlessManager_ForceHeadless(t *testing.T) {
	hm := NewHeadlessManager()
	hm.ForceHeadless(true)
	if !hm.IsHeadless() {
		t.Error("expected IsHeadless true after ForceHeadless(true)")
	}
}

func TestHeadlessManager_ForceHeadlessFalse(t *testing.T) {
	hm := NewHeadlessManager()
	hm.ForceHeadless(true)
	if !hm.IsHeadless() {
		t.Error("expected IsHeadless true after ForceHeadless(true)")
	}
	hm.ForceHeadless(false)
	// ForceHeadless(false) forces interactive (non-headless) mode.
	if hm.IsHeadless() {
		t.Error("expected IsHeadless false after ForceHeadless(false)")
	}
}

func TestHeadlessManager_ClearForce(t *testing.T) {
	hm := NewHeadlessManager()
	hm.ForceHeadless(true)
	if !hm.IsHeadless() {
		t.Error("expected IsHeadless true after ForceHeadless(true)")
	}
	hm.ClearForce()
	// After ClearForce, it reverts to TTY auto-detection.
	// We just verify no panic; actual result depends on test environment TTY state.
}

func TestHeadlessManager_SetDefaults(t *testing.T) {
	hm := NewHeadlessManager()
	defaults := map[string]string{
		"project_name": "test-project",
		"language":     "Go",
	}
	hm.SetDefaults(defaults)

	got, ok := hm.GetDefault("project_name")
	if !ok {
		t.Error("expected project_name default to exist")
	}
	if got != "test-project" {
		t.Errorf("expected 'test-project', got %q", got)
	}
}

func TestHeadlessManager_GetDefault_Missing(t *testing.T) {
	hm := NewHeadlessManager()
	_, ok := hm.GetDefault("nonexistent")
	if ok {
		t.Error("expected ok=false for missing default key")
	}
}

func TestHeadlessManager_SetDefaults_Overwrites(t *testing.T) {
	hm := NewHeadlessManager()
	hm.SetDefaults(map[string]string{"key": "v1"})
	hm.SetDefaults(map[string]string{"key": "v2"})

	got, ok := hm.GetDefault("key")
	if !ok {
		t.Fatal("expected key to exist")
	}
	if got != "v2" {
		t.Errorf("expected 'v2', got %q", got)
	}
}

func TestHeadlessManager_HasDefaults(t *testing.T) {
	hm := NewHeadlessManager()
	if hm.HasDefaults() {
		t.Error("expected HasDefaults false when none set")
	}

	hm.SetDefaults(map[string]string{"key": "val"})
	if !hm.HasDefaults() {
		t.Error("expected HasDefaults true after SetDefaults")
	}
}

func TestHeadlessManager_SetDefaults_NilMap(t *testing.T) {
	hm := NewHeadlessManager()
	hm.SetDefaults(nil)
	if hm.HasDefaults() {
		t.Error("expected HasDefaults false for nil map")
	}
}

func TestHeadlessManager_SetDefaults_EmptyMap(t *testing.T) {
	hm := NewHeadlessManager()
	hm.SetDefaults(map[string]string{})
	if hm.HasDefaults() {
		t.Error("expected HasDefaults false for empty map")
	}
}
