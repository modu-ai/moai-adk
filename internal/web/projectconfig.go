package web

import (
	"fmt"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/pkg/models"
)

// @MX:WARN: [AUTO] readProjectConfig/writeProjectConfig는 프로필 스토어가 아닌 *프로젝트 설정*
// (.moai/config/sections/quality.yaml + git-convention.yaml)을 디스크에서 읽고 쓰는 두 번째 영속화 경계다
// (SPEC-WEB-CONSOLE-003). handlers.go handleSave의 첫 번째 경계(WritePreferences + SyncToProjectConfig)와는 별개다.
// @MX:REASON: [AUTO] 영속화는 반드시 config.NewConfigManager()/LoadRaw/SetSection/Save 를 통해서만 수행한다 —
// 웹 레이어에서 YAML을 직접 marshal/os.WriteFile 하는 것은 금지된 안티패턴(REQ-WC3-008). scope는 quality(development_mode)
// + git_convention(convention) 두 섹션으로 엄격히 한정되며 workflow/harness/git-strategy/llm은 절대 건드리지 않는다
// (REQ-WC3-007). 비어있는 제출값은 기존 영속값을 덮어쓰지 않는다(empty = "keep existing", EC-1).

// readProjectConfig is the real read seam (REQ-WC3-004). It loads the project
// config via the config manager (LoadRaw — no validation, write-intent path)
// and returns the persisted development_mode + git_convention.convention. An
// absent config dir yields empty values (LoadRaw default behavior, EC-5).
func readProjectConfig(projectRoot string) (devMode, convention string, err error) {
	mgr := config.NewConfigManager()
	cfg, err := mgr.LoadRaw(projectRoot)
	if err != nil {
		return "", "", fmt.Errorf("read project config: %w", err)
	}
	return string(cfg.Quality.DevelopmentMode), cfg.GitConvention.Convention, nil
}

// writeProjectConfig is the real write seam (REQ-WC3-005/007). It persists each
// non-empty value into its project-config section via the config-manager API
// (LoadRaw → mutate only non-empty → SetSection → Save). Empty submissions leave
// the existing persisted value unchanged (EC-1). It writes ONLY the quality
// (development_mode) and git_convention (convention) sections — Save() round-trips
// every other section's content unchanged. No direct yaml.Marshal/os.WriteFile.
func writeProjectConfig(projectRoot, devMode, convention string) error {
	mgr := config.NewConfigManager()
	cfg, err := mgr.LoadRaw(projectRoot)
	if err != nil {
		return fmt.Errorf("load project config: %w", err)
	}

	changed := false

	if devMode != "" && string(cfg.Quality.DevelopmentMode) != devMode {
		quality := cfg.Quality
		quality.DevelopmentMode = models.DevelopmentMode(devMode)
		if err := mgr.SetSection("quality", quality); err != nil {
			return fmt.Errorf("set quality section: %w", err)
		}
		changed = true
	}

	if convention != "" && cfg.GitConvention.Convention != convention {
		gc := cfg.GitConvention
		gc.Convention = convention
		if err := mgr.SetSection("git_convention", gc); err != nil {
			return fmt.Errorf("set git_convention section: %w", err)
		}
		changed = true
	}

	if changed {
		if err := mgr.Save(); err != nil {
			return fmt.Errorf("save project config: %w", err)
		}
	}
	return nil
}
