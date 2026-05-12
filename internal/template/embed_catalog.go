// Package template — SPEC-V3R4-CATALOG-002 helpers.
//
// embed_catalog.go provides the orchestration surface between the catalog
// manifest (CATALOG-001) and the slim deploy filter (CATALOG-002). The
// embedded raw filesystem (embedded.FS) remains an UNEXPORTED package
// variable; external callers MUST go through LoadEmbeddedCatalog() and
// NewSlimDeployerWithRenderer() rather than accessing embeddedRaw directly.
//
// DEFECT-5 INVARIANT: No exported function returns embeddedRaw or fs.FS over
// the raw (unfiltered) embed. The only externally visible names are
// LoadEmbeddedCatalog and NewSlimDeployerWithRenderer. See spec.md §"Overview"
// [INVARIANT] block.
package template

import (
	"fmt"
)

// @MX:NOTE: [AUTO] CATALOG-001 + CATALOG-002 wrapper convenience. Loads embedded catalog.yaml and returns typed *Catalog.
//
// LoadEmbeddedCatalog parses catalog.yaml from the embedded filesystem and
// returns a fully-typed *Catalog. This is a convenience wrapper around
// LoadCatalog(embeddedRaw) — external packages SHOULD use this entry point
// rather than constructing their own catalog loaders.
//
// Returns CATALOG_LOAD_FAILED-tagged error (via caller's error wrapping) when
// the manifest is missing or corrupt. See SPEC-V3R4-CATALOG-001 §"Loader API"
// for the typed contract.
func LoadEmbeddedCatalog() (*Catalog, error) {
	cat, err := LoadCatalog(embeddedRaw)
	if err != nil {
		return nil, fmt.Errorf("load embedded catalog: %w", err)
	}
	return cat, nil
}

// @MX:ANCHOR: [AUTO] CATALOG-002 encapsulated entry point — sole external surface for slim deploy.
// @MX:REASON: [AUTO] fan_in=1 now (init.go); future CATALOG-003/004 will route through this; preserves embeddedRaw encapsulation (DEFECT-5).
//
// NewSlimDeployerWithRenderer constructs a Deployer that writes only
// core-tier catalog entries (plus non-catalog template files) by wrapping
// the raw embedded filesystem in SlimFS before passing it to
// NewDeployerWithRenderer. The caller-provided renderer is reused as-is.
//
// This is the encapsulated entry point for slim mode. External packages
// MUST use this constructor rather than calling SlimFS(embeddedRaw, cat)
// directly — the raw embed FS is an unexported package variable.
//
// Returns an error if cat is nil or if SlimFS wrapping fails.
func NewSlimDeployerWithRenderer(cat *Catalog, renderer Renderer) (Deployer, error) {
	if cat == nil {
		return nil, fmt.Errorf("nil catalog")
	}
	slim, err := SlimFS(embeddedRaw, cat)
	if err != nil {
		return nil, fmt.Errorf("slim fs: %w", err)
	}
	return NewDeployerWithRenderer(slim, renderer), nil
}
