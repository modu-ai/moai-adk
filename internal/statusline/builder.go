package statusline

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"sync"
)

// defaultBuilder implements the Builder interface by orchestrating
// data collection from multiple sources and rendering the statusline.
type defaultBuilder struct {
	gitProvider    GitDataProvider
	updateProvider UpdateProvider
	renderer       *Renderer
	mode           StatuslineMode
	mu             sync.RWMutex
}

// Options configures a new Builder instance.
type Options struct {
	// GitProvider collects git repository status. May be nil if no git repo.
	GitProvider GitDataProvider

	// UpdateProvider checks for version updates. May be nil to skip.
	UpdateProvider UpdateProvider

	// ThemeName selects the rendering theme: "default", "minimal", "nerd".
	ThemeName string

	// Mode sets the initial display mode.
	Mode StatuslineMode

	// NoColor disables all ANSI color output when true.
	NoColor bool
}

// New creates a new Builder with the given options.
// If Mode is empty, defaults to ModeDefault.
func New(opts Options) Builder {
	mode := opts.Mode
	if mode == "" {
		mode = ModeDefault
	}

	return &defaultBuilder{
		gitProvider:    opts.GitProvider,
		updateProvider: opts.UpdateProvider,
		renderer:       NewRenderer(opts.ThemeName, opts.NoColor),
		mode:           mode,
	}
}

// Build reads JSON from r, collects data from all sources in parallel,
// and returns a formatted single-line statusline string.
// On any input error, it produces a safe fallback output.
// The output never contains newline characters.
func (b *defaultBuilder) Build(ctx context.Context, r io.Reader) (string, error) {
	mode := b.getMode()

	// Parse stdin JSON
	input := b.parseStdin(r)

	// Collect data from all sources
	data := b.collectAll(ctx, input)

	// Render the statusline
	result := b.renderer.Render(data, mode)

	return result, nil
}

// getMode returns the current display mode. Thread-safe.
func (b *defaultBuilder) getMode() StatuslineMode {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.mode
}

// SetMode switches the display mode. Thread-safe.
func (b *defaultBuilder) SetMode(mode StatuslineMode) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.mode = mode
}

// parseStdin reads and parses JSON from the reader.
// Returns nil on any error (empty stdin, invalid JSON, etc.).
func (b *defaultBuilder) parseStdin(r io.Reader) *StdinData {
	if r == nil {
		return nil
	}

	var input StdinData
	decoder := json.NewDecoder(r)
	if err := decoder.Decode(&input); err != nil {
		slog.Debug("stdin JSON parse failed", "error", err)
		return nil
	}

	return &input
}

// collectAll gathers data from all sources in parallel.
// Individual collector failures are non-fatal; partial data is used.
func (b *defaultBuilder) collectAll(ctx context.Context, input *StdinData) *StatusData {
	data := &StatusData{}

	// Instant collectors (no I/O, no concurrency needed)
	if mem := CollectMemory(input); mem != nil {
		data.Memory = *mem
	}
	if met := CollectMetrics(input); met != nil {
		data.Metrics = *met
	}

	// Parallel collectors (may involve I/O)
	var wg sync.WaitGroup
	var gitResult *GitStatusData
	var versionResult *VersionData

	if b.gitProvider != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			result, err := b.gitProvider.CollectGitStatus(ctx)
			if err != nil {
				slog.Debug("git collection failed", "error", err)
				return
			}
			gitResult = result
		}()
	}

	if b.updateProvider != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			result, err := b.updateProvider.CheckUpdate(ctx)
			if err != nil {
				slog.Debug("update check failed", "error", err)
				return
			}
			versionResult = result
		}()
	}

	wg.Wait()

	if gitResult != nil {
		data.Git = *gitResult
	}
	if versionResult != nil {
		data.Version = *versionResult
	}

	return data
}
