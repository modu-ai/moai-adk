package statusline

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"sync"

	gitpkg "github.com/modu-ai/moai-adk/internal/core/git"
	"github.com/modu-ai/moai-adk/pkg/version"
)

// defaultBuilder implements the Builder interface by orchestrating
// data collection from multiple sources and rendering the statusline.
type defaultBuilder struct {
	gitProvider    GitDataProvider
	updateProvider UpdateProvider
	usageProvider  UsageProvider // @MX:NOTE: API 사용량 수집 (Phase 5, REQ-V3-API-001)
	renderer       *Renderer
	mode           StatuslineMode
	mu             sync.RWMutex
}

// Options configures a new Builder instance.
type Options struct {
	// GitProvider collects git repository status. May be nil if no git repo.
	// If nil, git repository will be opened automatically.
	GitProvider GitDataProvider

	// UpdateProvider checks for version updates. May be nil to skip.
	// If nil, version will be read from config file automatically.
	UpdateProvider UpdateProvider

	// UsageProvider collects API usage (5H/7D). May be nil to skip.
	// If nil, usage collection is disabled.
	UsageProvider UsageProvider

	// RootDir is the project root directory for auto-detecting git repo.
	// If empty, current directory is used.
	RootDir string

	// HomeDir is the user's home directory for usage cache.
	// If empty, os.UserHomeDir() is used.
	HomeDir string

	// ThemeName selects the rendering theme: "default", "minimal", "nerd".
	ThemeName string

	// Mode sets the initial display mode.
	Mode StatuslineMode

	// NoColor disables all ANSI color output when true.
	NoColor bool

	// SegmentConfig maps segment keys to enabled state.
	// When nil or empty, all segments are displayed (backward compatible).
	SegmentConfig map[string]bool
}

// New creates a new Builder with the given options.
// If Mode is empty, defaults to ModeDefault.
// If GitProvider is nil, attempts to open a git repository at RootDir (or ".") automatically.
// If UpdateProvider is nil, attempts to read version from config file automatically.
// If UsageProvider is nil, attempts to create usage collector automatically (Phase 5).
//
// @MX:ANCHOR: statusline 패키지의 공개 생성자 - 모든 외부 호출자가 이 함수를 통해 Builder를 생성함
// @MX:REASON: 공개 API 경계점; auto-detection 로직(git/update/usage provider) 포함
func New(opts Options) Builder {
	mode := opts.Mode
	if mode == "" {
		mode = ModeDefault
	}
	// deprecated 모드 이름을 현재 이름으로 정규화한다 (REQ-V3-MODE-001, REQ-V3-MODE-002)
	mode = NormalizeMode(mode)

	gitProvider := opts.GitProvider
	updateProvider := opts.UpdateProvider
	usageProvider := opts.UsageProvider

	// Auto-create git provider if not provided
	if gitProvider == nil {
		rootDir := opts.RootDir
		if rootDir == "" {
			rootDir = "."
		}
		if repo, err := gitpkg.NewRepository(rootDir); err == nil {
			gitProvider = NewGitCollector(repo)
			slog.Debug("auto-opened git repository for statusline", "root", repo.Root())
		}
		// If git repo not found, continue without git provider
	}

	// Auto-create version provider if not provided
	if updateProvider == nil {
		updateProvider = NewVersionCollector(version.GetVersion())
		slog.Debug("auto-created version collector for statusline")
	}

	// Auto-create usage provider if not provided (Phase 5, REQ-V3-API-001)
	if usageProvider == nil {
		homeDir := opts.HomeDir
		if homeDir == "" {
			// 홈 디렉토리 자동 감지
			if h, err := os.UserHomeDir(); err == nil {
				homeDir = h
			}
		}
		if homeDir != "" {
			usageProvider = NewUsageCollector(homeDir)
			slog.Debug("auto-created usage collector for statusline")
		}
		// If home dir not found, continue without usage provider
	}

	return &defaultBuilder{
		gitProvider:    gitProvider,
		updateProvider: updateProvider,
		usageProvider:  usageProvider,
		renderer:       NewRenderer(opts.ThemeName, opts.NoColor, opts.SegmentConfig),
		mode:           mode,
	}
}

// Build reads JSON from r, collects data from all sources in parallel,
// and returns a formatted single-line statusline string.
// On any input error, it produces a safe fallback output.
// The output never contains newline characters.
//
// @MX:ANCHOR: 전체 statusline 빌드 파이프라인의 핵심 메서드
// @MX:REASON: Builder 인터페이스 구현; parseStdin + collectAll + Render 3단계 파이프라인 조율
func (b *defaultBuilder) Build(ctx context.Context, r io.Reader) (string, error) {
	mode := b.getMode()

	// Parse stdin JSON
	input := b.parseStdin(r)

	// Collect data from all sources
	data := b.collectAll(ctx, input)

	// renderer가 v3 모드를 직접 지원하므로 mode를 그대로 전달한다 (Phase 4, REQ-V3-LAYOUT-001~003)
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
// deprecated 모드 이름("minimal", "verbose")은 자동으로 정규화된다 (REQ-V3-MODE-001, REQ-V3-MODE-002).
func (b *defaultBuilder) SetMode(mode StatuslineMode) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.mode = NormalizeMode(mode)
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
	// 활성 태스크 정보 수집 (Phase 4에서 렌더링 활성화됨, REQ-V3 Cycle 5)
	if task := CollectTask(); task != nil {
		data.Task = *task
	}

	// Extract directory name from workspace (prefer project_dir per documentation)
	if input != nil {
		data.Directory = extractProjectDirectory(input)
	}

	// Extract output style from nested structure
	if input != nil && input.OutputStyle != nil {
		data.OutputStyle = input.OutputStyle.Name
	}

	// Extract Claude Code version from JSON input
	if input != nil && input.Version != "" {
		data.ClaudeCodeVersion = input.Version
	}

	// Parallel collectors (may involve I/O)
	var wg sync.WaitGroup
	var gitResult *GitStatusData
	var versionResult *VersionData
	var usageResult *UsageResult // @MX:NOTE: API 사용량 병렬 수집 (Phase 5)

	if b.gitProvider != nil {
		wg.Go(func() {
			result, err := b.gitProvider.CollectGitStatus(ctx)
			if err != nil {
				slog.Debug("git collection failed", "error", err)
				return
			}
			gitResult = result
		})
	}

	if b.updateProvider != nil {
		wg.Go(func() {
			result, err := b.updateProvider.CheckUpdate(ctx)
			if err != nil {
				slog.Debug("update check failed", "error", err)
				return
			}
			versionResult = result
		})
	}

	// API 사용량 수집 (Phase 5, REQ-V3-API-001)
	if b.usageProvider != nil {
		wg.Go(func() {
			result, err := b.usageProvider.CollectUsage(ctx)
			if err != nil {
				slog.Debug("usage collection failed", "error", err)
				return
			}
			usageResult = result
		})
	}

	wg.Wait()

	if gitResult != nil {
		data.Git = *gitResult
	}
	if versionResult != nil {
		data.Version = *versionResult
	}
	if usageResult != nil {
		data.Usage = usageResult // @MX:NOTE: 사용량 데이터 할당 (Phase 5)
	}

	return data
}

// extractProjectDirectory extracts the project directory name from workspace.
// Priority: workspace.project_dir > workspace.current_dir > cwd (legacy)
// Per https://code.claude.com/docs/en/statusline documentation.
func extractProjectDirectory(input *StdinData) string {
	if input == nil {
		return ""
	}

	// Priority 1: Use workspace.project_dir (preferred)
	if input.Workspace != nil && input.Workspace.ProjectDir != "" {
		return filepath.Base(input.Workspace.ProjectDir)
	}

	// Priority 2: Use workspace.current_dir
	if input.Workspace != nil && input.Workspace.CurrentDir != "" {
		return filepath.Base(input.Workspace.CurrentDir)
	}

	// Priority 3: Fall back to legacy cwd field
	if input.CWD != "" {
		return filepath.Base(input.CWD)
	}

	return ""
}
