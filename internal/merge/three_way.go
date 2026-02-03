package merge

import (
	"context"
	"fmt"
)

// engine is the concrete implementation of the Engine interface.
type engine struct {
	selector StrategySelector
}

// NewEngine creates a new Engine with the default strategy selector.
func NewEngine() Engine {
	return &engine{
		selector: NewStrategySelector(),
	}
}

// NewEngineWithSelector creates a new Engine with a custom strategy selector.
func NewEngineWithSelector(selector StrategySelector) Engine {
	return &engine{
		selector: selector,
	}
}

// ThreeWayMerge performs a generic line-based 3-way merge on byte slices.
// It always uses the LineMerge strategy regardless of file type.
func (e *engine) ThreeWayMerge(base, current, updated []byte) (*MergeResult, error) {
	return mergeLineBased(base, current, updated)
}

// MergeFile performs a strategy-aware 3-way merge, selecting the merge
// algorithm based on the file path and extension.
func (e *engine) MergeFile(ctx context.Context, path string, base, current, updated []byte) (*MergeResult, error) {
	// Check for context cancellation before starting.
	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("merge file %q: %w", path, ctx.Err())
	default:
	}

	strategy := e.selector.SelectStrategy(path)

	switch strategy {
	case LineMerge:
		return mergeLineBased(base, current, updated)
	case YAMLDeep:
		return mergeYAML(base, current, updated)
	case JSONMerge:
		return mergeJSON(base, current, updated)
	case SectionMerge:
		return mergeSectionBased(base, current, updated)
	case EntryMerge:
		return mergeEntryBased(base, current, updated)
	case Overwrite:
		return mergeOverwrite(current, updated)
	default:
		return nil, fmt.Errorf("%w: strategy %q for %q", ErrMergeUnsupported, strategy, path)
	}
}
