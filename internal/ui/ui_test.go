package ui

import (
	"testing"
)

func TestErrCancelled_Error(t *testing.T) {
	err := ErrCancelled
	if err.Error() != "operation cancelled" {
		t.Errorf("expected 'operation cancelled', got %q", err.Error())
	}
}

func TestErrNoItems_Error(t *testing.T) {
	err := ErrNoItems
	if err.Error() != "no items to select from" {
		t.Errorf("expected 'no items to select from', got %q", err.Error())
	}
}

func TestErrHeadlessNoDefaults_Error(t *testing.T) {
	err := ErrHeadlessNoDefaults
	if err.Error() != "headless mode requires defaults for all wizard fields" {
		t.Errorf("expected headless no defaults message, got %q", err.Error())
	}
}

func TestSelectItem_Fields(t *testing.T) {
	item := SelectItem{Label: "Go", Value: "go", Desc: "Language"}
	if item.Label != "Go" {
		t.Errorf("expected Label 'Go', got %q", item.Label)
	}
	if item.Value != "go" {
		t.Errorf("expected Value 'go', got %q", item.Value)
	}
	if item.Desc != "Language" {
		t.Errorf("expected Desc 'Language', got %q", item.Desc)
	}
}

func TestWizardResult_Fields(t *testing.T) {
	r := WizardResult{
		ProjectName: "proj",
		Language:    "Go",
		Framework:   "Cobra",
		Features:    []string{"LSP"},
		UserName:    "user",
		ConvLang:    "en",
	}
	if r.ProjectName != "proj" {
		t.Error("ProjectName mismatch")
	}
	if r.Language != "Go" {
		t.Error("Language mismatch")
	}
	if r.Framework != "Cobra" {
		t.Error("Framework mismatch")
	}
	if len(r.Features) != 1 || r.Features[0] != "LSP" {
		t.Error("Features mismatch")
	}
	if r.UserName != "user" {
		t.Error("UserName mismatch")
	}
	if r.ConvLang != "en" {
		t.Error("ConvLang mismatch")
	}
}

func TestNewProgress_ReturnsNonNil(t *testing.T) {
	theme := testTheme()
	hm := NewHeadlessManager()
	p := NewProgress(theme, hm)
	if p == nil {
		t.Error("NewProgress should not return nil")
	}
}
