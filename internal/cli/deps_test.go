package cli

import (
	"testing"
)

func TestInitDependencies(t *testing.T) {
	// Save and restore original deps
	origDeps := deps
	defer func() { deps = origDeps }()

	deps = nil
	InitDependencies()

	if deps == nil {
		t.Fatal("InitDependencies should set deps")
	}

	if deps.Config == nil {
		t.Error("deps.Config should not be nil")
	}
	if deps.HookProtocol == nil {
		t.Error("deps.HookProtocol should not be nil")
	}
	if deps.HookRegistry == nil {
		t.Error("deps.HookRegistry should not be nil")
	}
	if deps.RankCredStore == nil {
		t.Error("deps.RankCredStore should not be nil")
	}
	if deps.Logger == nil {
		t.Error("deps.Logger should not be nil")
	}
}

func TestGetDeps_ReturnsNilBeforeInit(t *testing.T) {
	origDeps := deps
	defer func() { deps = origDeps }()

	deps = nil
	if GetDeps() != nil {
		t.Error("GetDeps should return nil before InitDependencies")
	}
}

func TestGetDeps_ReturnsAfterInit(t *testing.T) {
	origDeps := deps
	defer func() { deps = origDeps }()

	InitDependencies()
	if GetDeps() == nil {
		t.Error("GetDeps should return non-nil after InitDependencies")
	}
}

func TestSetDeps(t *testing.T) {
	origDeps := deps
	defer func() { deps = origDeps }()

	custom := &Dependencies{}
	SetDeps(custom)

	if GetDeps() != custom {
		t.Error("SetDeps should replace the global deps")
	}
}
