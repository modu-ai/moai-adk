package mx

import (
	"fmt"
	"sort"
)

// Resolver provides AnchorID resolution services.
// This is a placeholder for SPC-004 which will implement full fan-in analysis.
type Resolver struct {
	manager *Manager
}

// NewResolver creates a new AnchorID resolver.
func NewResolver(manager *Manager) *Resolver {
	return &Resolver{
		manager: manager,
	}
}

// ResolveAnchor looks up an AnchorID and returns the corresponding tag.
// Returns error if AnchorID is not found.
//
// @MX:ANCHOR: [AUTO] ResolveAnchor — AnchorID 단일 조회 API의 불변 계약
// @MX:REASON: fan_in >= 3 — ResolveAll 내부 호출, CLI, 코드맵 생성 도구에서 사용
func (r *Resolver) ResolveAnchor(anchorID string) (Tag, error) {
	tags := r.manager.GetAllTags()

	for _, tag := range tags {
		if tag.Kind == MXAnchor && tag.AnchorID == anchorID {
			return tag, nil
		}
	}

	return Tag{}, fmt.Errorf("anchor ID not found: %s", anchorID)
}

// ResolveAll returns all tags for a given set of AnchorIDs.
func (r *Resolver) ResolveAll(anchorIDs []string) ([]Tag, error) {
	var result []Tag
	var missing []string

	for _, anchorID := range anchorIDs {
		tag, err := r.ResolveAnchor(anchorID)
		if err != nil {
			missing = append(missing, anchorID)
		} else {
			result = append(result, tag)
		}
	}

	if len(missing) > 0 {
		return result, fmt.Errorf("missing anchors: %v", missing)
	}

	return result, nil
}

// ListAnchors returns all ANCHOR tags in the project, sorted by file path.
func (r *Resolver) ListAnchors() []Tag {
	tags := r.manager.GetAllTags()

	var anchors []Tag
	for _, tag := range tags {
		if tag.Kind == MXAnchor {
			anchors = append(anchors, tag)
		}
	}

	// Sort by file path for consistent output
	sort.Slice(anchors, func(i, j int) bool {
		return anchors[i].File < anchors[j].File
	})

	return anchors
}

// AuditLowFanIn returns ANCHOR tags with low fan-in (< 3 callers).
// This is a placeholder implementation - SPC-004 will implement actual fan-in counting.
func (r *Resolver) AuditLowFanIn() []Tag {
	anchors := r.ListAnchors()

	// Placeholder: Return all anchors as "low fan-in" until SPC-004 implementation
	// In SPC-004, this will use actual static analysis to count callers
	return anchors
}
