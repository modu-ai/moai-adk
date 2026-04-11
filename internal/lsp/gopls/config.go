package gopls

// @MX:ANCHOR: [AUTO] Config 타입 및 로더 — bridge.go, NewBridge에서 의존하는 설정 진입점
// @MX:REASON: fan_in >= 3 (NewBridge, 테스트, 외부 의존 주입 경로)

import (
	"errors"
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config는 gopls 브릿지 동작 설정을 담는다.
// REQ-GB-050: .moai/config/sections/lsp.yaml의 gopls_bridge 섹션에서 로드한다.
// REQ-GB-022: 타임아웃과 디바운스 창은 lsp.yaml에서 설정 가능하다.
type Config struct {
	// Enabled는 gopls 브릿지 활성화 여부다.
	// REQ-GB-051: false(기본값)이면 GoFeedbackGenerator는 CLI 전용 경로를 유지한다.
	Enabled bool
	// Binary는 gopls 실행 파일 경로 또는 이름이다. 기본값: "gopls".
	Binary string
	// Args는 gopls 서브프로세스에 전달할 추가 인수다.
	Args []string
	// InitOptions는 LSP initialize 요청의 initializationOptions다.
	// REQ-GB-013: staticcheck: true를 포함해야 한다.
	InitOptions map[string]any
	// Timeout은 개별 LSP 요청(didOpen, 진단 등)의 타임아웃이다. 기본값: 30s.
	// REQ-GB-020: didOpen 후 publishDiagnostics 대기 타임아웃.
	Timeout time.Duration
	// InitTimeout은 LSP initialize 핸드셰이크 타임아웃이다.
	// REQ-GB-012: 초기화 타임아웃 30초.
	InitTimeout time.Duration
	// ShutdownTimeout은 graceful shutdown 타임아웃이다.
	// REQ-GB-004: 5초 타임아웃 후 SIGKILL.
	ShutdownTimeout time.Duration
	// DebounceWindow는 publishDiagnostics 디바운스 창이다.
	// REQ-GB-021: 기본값 150ms.
	DebounceWindow time.Duration
}

// DefaultConfig는 합리적인 기본값으로 채워진 Config를 반환한다.
// binary=gopls, timeout=30s, initTimeout=30s, shutdownTimeout=5s, debounce=150ms.
func DefaultConfig() *Config {
	return &Config{
		Enabled:         false,
		Binary:          "gopls",
		Args:            []string{},
		InitOptions:     map[string]any{"staticcheck": true},
		Timeout:         30 * time.Second,
		InitTimeout:     30 * time.Second,
		ShutdownTimeout: 5 * time.Second,
		DebounceWindow:  150 * time.Millisecond,
	}
}

// ─── YAML 파싱용 내부 구조체 ─────────────────────────────────────────────────
// lsp.yaml의 구조와 1:1 매핑한다.

type lspYAMLRoot struct {
	LSP lspYAMLConfig `yaml:"lsp"`
}

type lspYAMLConfig struct {
	GoplsBridge goplsBridgeYAML `yaml:"gopls_bridge"`
}

type goplsBridgeYAML struct {
	Enabled     bool           `yaml:"enabled"`
	Binary      string         `yaml:"binary"`
	Args        []string       `yaml:"args"`
	InitOptions map[string]any `yaml:"init_options"`
	Timeouts    timeoutsYAML   `yaml:"timeouts"`
}

type timeoutsYAML struct {
	InitializeSeconds      int `yaml:"initialize_seconds"`
	RequestSeconds         int `yaml:"request_seconds"`
	ShutdownSeconds        int `yaml:"shutdown_seconds"`
	DiagnosticsDebounceMs  int `yaml:"diagnostics_debounce_ms"`
}

// LoadConfig는 configPath의 YAML 파일에서 gopls 브릿지 설정을 로드한다.
// 파일이 없으면 DefaultConfig를 반환한다 (오류 아님).
// YAML 구문 오류 등 다른 오류는 error를 반환한다.
//
// REQ-GB-050: lsp.yaml에서 설정을 읽는다.
// REQ-GB-002 연계: gopls 없음과 동일하게 설정 없음도 graceful하게 처리한다.
func LoadConfig(configPath string) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			// 파일이 없으면 기본값을 반환한다. 오류가 아니다.
			return DefaultConfig(), nil
		}
		return nil, fmt.Errorf("gopls: config: 파일 읽기 실패 %q: %w", configPath, err)
	}

	var root lspYAMLRoot
	if err := yaml.Unmarshal(data, &root); err != nil {
		return nil, fmt.Errorf("gopls: config: YAML 파싱 실패 %q: %w", configPath, err)
	}

	return mergeWithDefaults(&root.LSP.GoplsBridge), nil
}

// mergeWithDefaults는 YAML에서 파싱한 gopls_bridge 설정과 기본값을 병합한다.
// 0 값인 필드는 기본값으로 채운다.
func mergeWithDefaults(y *goplsBridgeYAML) *Config {
	def := DefaultConfig()
	cfg := &Config{
		Enabled:     y.Enabled,
		Binary:      def.Binary,
		Args:        def.Args,
		InitOptions: def.InitOptions,
	}

	if y.Binary != "" {
		cfg.Binary = y.Binary
	}
	if len(y.Args) > 0 {
		cfg.Args = y.Args
	}
	if len(y.InitOptions) > 0 {
		cfg.InitOptions = y.InitOptions
	}

	// 타임아웃: 0이면 기본값 사용
	cfg.InitTimeout = durationOrDefault(y.Timeouts.InitializeSeconds, 0, def.InitTimeout, time.Second)
	cfg.Timeout = durationOrDefault(y.Timeouts.RequestSeconds, 0, def.Timeout, time.Second)
	cfg.ShutdownTimeout = durationOrDefault(y.Timeouts.ShutdownSeconds, 0, def.ShutdownTimeout, time.Second)
	cfg.DebounceWindow = durationOrDefault(y.Timeouts.DiagnosticsDebounceMs, 0, def.DebounceWindow, time.Millisecond)

	return cfg
}

// durationOrDefault는 값이 0이면 defaultVal을 반환하고, 아니면 value * unit을 반환한다.
func durationOrDefault(value, zero int, defaultVal time.Duration, unit time.Duration) time.Duration {
	if value == zero {
		return defaultVal
	}
	return time.Duration(value) * unit
}
