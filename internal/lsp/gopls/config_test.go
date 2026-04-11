package gopls

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestDefaultConfig는 DefaultConfig가 합리적인 기본값을 반환하는지 검증한다.
func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg == nil {
		t.Fatal("DefaultConfig()가 nil을 반환해서는 안 된다")
	}
	if cfg.Binary != "gopls" {
		t.Errorf("기본 Binary = %q, 원하는 값 = %q", cfg.Binary, "gopls")
	}
	if cfg.Timeout != 30*time.Second {
		t.Errorf("기본 Timeout = %v, 원하는 값 = %v", cfg.Timeout, 30*time.Second)
	}
	if cfg.Enabled {
		t.Error("기본 Enabled는 false여야 한다")
	}
}

// TestLoadConfig_ValidYAML은 올바른 YAML 파일에서 설정을 로드하는지 검증한다.
func TestLoadConfig_ValidYAML(t *testing.T) {
	yaml := `lsp:
    gopls_bridge:
        enabled: true
        binary: /usr/local/bin/gopls
        args:
            - "-remote=auto"
        init_options:
            staticcheck: true
        timeouts:
            initialize_seconds: 60
            request_seconds: 10
            shutdown_seconds: 10
            diagnostics_debounce_ms: 200
`
	dir := t.TempDir()
	configPath := filepath.Join(dir, "lsp.yaml")
	if err := os.WriteFile(configPath, []byte(yaml), 0644); err != nil {
		t.Fatalf("테스트 파일 생성 실패: %v", err)
	}

	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig 오류: %v", err)
	}
	if !cfg.Enabled {
		t.Error("Enabled = false, true를 기대했다")
	}
	if cfg.Binary != "/usr/local/bin/gopls" {
		t.Errorf("Binary = %q, 원하는 값 = %q", cfg.Binary, "/usr/local/bin/gopls")
	}
	if len(cfg.Args) != 1 || cfg.Args[0] != "-remote=auto" {
		t.Errorf("Args = %v, 원하는 값 = [-remote=auto]", cfg.Args)
	}
	// 타임아웃 검증
	if cfg.Timeout != 10*time.Second {
		t.Errorf("Timeout (request) = %v, 원하는 값 = %v", cfg.Timeout, 10*time.Second)
	}
	if cfg.InitTimeout != 60*time.Second {
		t.Errorf("InitTimeout = %v, 원하는 값 = %v", cfg.InitTimeout, 60*time.Second)
	}
	if cfg.ShutdownTimeout != 10*time.Second {
		t.Errorf("ShutdownTimeout = %v, 원하는 값 = %v", cfg.ShutdownTimeout, 10*time.Second)
	}
	if cfg.DebounceWindow != 200*time.Millisecond {
		t.Errorf("DebounceWindow = %v, 원하는 값 = %v", cfg.DebounceWindow, 200*time.Millisecond)
	}
	// init_options 검증
	if v, ok := cfg.InitOptions["staticcheck"]; !ok || v != true {
		t.Errorf("InitOptions[staticcheck] = %v, true를 기대했다", v)
	}
}

// TestLoadConfig_MissingFile은 파일이 없으면 기본값을 반환하는지 검증한다.
// REQ-GB-002 대응: 파일 없음은 오류가 아니라 기본값을 반환한다.
func TestLoadConfig_MissingFile(t *testing.T) {
	cfg, err := LoadConfig("/nonexistent/path/lsp.yaml")
	if err != nil {
		t.Fatalf("파일 없음 시 오류가 반환됐다: %v (기본값을 기대했다)", err)
	}
	if cfg == nil {
		t.Fatal("파일 없음 시 nil이 반환됐다 (기본값을 기대했다)")
	}
	// 기본값 확인
	if cfg.Binary != "gopls" {
		t.Errorf("기본 Binary = %q, 원하는 값 = gopls", cfg.Binary)
	}
}

// TestLoadConfig_InvalidYAML은 잘못된 YAML 파일에서 오류를 반환하는지 검증한다.
func TestLoadConfig_InvalidYAML(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "lsp.yaml")
	if err := os.WriteFile(configPath, []byte("not: valid: yaml: ["), 0644); err != nil {
		t.Fatalf("테스트 파일 생성 실패: %v", err)
	}

	_, err := LoadConfig(configPath)
	if err == nil {
		t.Error("잘못된 YAML에서 오류가 반환되지 않았다")
	}
}

// TestLoadConfig_PartialYAML은 일부 필드만 지정된 YAML에서 기본값이 채워지는지 검증한다.
func TestLoadConfig_PartialYAML(t *testing.T) {
	yaml := `lsp:
    gopls_bridge:
        enabled: true
`
	dir := t.TempDir()
	configPath := filepath.Join(dir, "lsp.yaml")
	if err := os.WriteFile(configPath, []byte(yaml), 0644); err != nil {
		t.Fatalf("테스트 파일 생성 실패: %v", err)
	}

	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig 오류: %v", err)
	}
	if !cfg.Enabled {
		t.Error("Enabled = false, true를 기대했다")
	}
	// 지정하지 않은 필드는 기본값
	if cfg.Binary != "gopls" {
		t.Errorf("부분 YAML: Binary = %q, 기본값 gopls를 기대했다", cfg.Binary)
	}
	if cfg.Timeout != 30*time.Second {
		t.Errorf("부분 YAML: Timeout = %v, 기본값 30s를 기대했다", cfg.Timeout)
	}
}

// TestLoadConfig_RealLSPYaml은 실제 .moai/config/sections/lsp.yaml을 로드할 수 있는지 검증한다.
// 이 파일이 없으면 테스트를 건너뛴다.
func TestLoadConfig_RealLSPYaml(t *testing.T) {
	// 프로젝트 루트에서 실제 파일 경로를 찾는다.
	// 테스트는 패키지 디렉토리에서 실행되므로 상위 경로로 이동한다.
	candidates := []string{
		"../../../.moai/config/sections/lsp.yaml",
		"../../../../.moai/config/sections/lsp.yaml",
	}
	var found string
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			found = c
			break
		}
	}
	if found == "" {
		t.Skip("실제 lsp.yaml 파일을 찾을 수 없어 테스트를 건너뜁니다")
	}

	cfg, err := LoadConfig(found)
	if err != nil {
		t.Fatalf("실제 lsp.yaml 로드 오류: %v", err)
	}
	if cfg == nil {
		t.Fatal("실제 lsp.yaml 로드 결과 nil")
	}
	// 실제 파일에서는 enabled가 false여야 한다 (기본값)
	if cfg.Binary != "gopls" {
		t.Errorf("Binary = %q, gopls를 기대했다", cfg.Binary)
	}
}
