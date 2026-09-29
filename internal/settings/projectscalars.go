package settings

// 이 파일은 development_mode + git_convention.convention 스칼라 편집의 공유
// 쓰기 seam이다 (SPEC-WEB-SAVE-LOSSLESS-001 M1 — plan-audit iter-2 F5 처분
// (a), REQ-WSL-002/003 + AC-WSL-009). 구 경로는 internal/web
// writeProjectConfig와 internal/cli persistProjectConfig(TUI 쌍둥이)가 각자
// LoadRaw → SetSection → Save 전체-재마샬을 수행해 GitHub issue #1731이
// 지목한 quality.yaml 미모델링 키(constitution.session_effort_default)와
// 주석을 devMode/convention 편집 한 번에 소실시켰다. 웹/TUI가 하나의 중립
// seam을 공유한다 (AP-2 — no parallel writer).

import (
	"fmt"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/settings/yamlpatch"
)

// WriteProjectScalars는 development_mode + git_convention.convention 편집을
// yamlpatch 라인-스플라이스로 영속화한다. 빈 제출값은 기존 영속값을 덮어쓰지
// 않는다(empty = keep existing, EC-1 — 구 writeProjectConfig 계약 승계).
//
// no-op 게이트(REQ-WSL-001 — 동치 제출은 mtime 포함 무기록): 제출값이 로드된
// 현재값과 같으면 편집 대상에서 제외되고, (env-override 등으로 구조체 값과
// 디스크 스칼라가 갈라져 있어) 디스크 스칼라가 이미 제출값과 같으면 스플라이스가
// 바이트 불변일 뿐이므로 역시 기록하지 않는다.
func WriteProjectScalars(projectRoot, devMode, convention string) error {
	mgr := config.NewConfigManager()
	cfg, err := mgr.LoadRaw(projectRoot)
	if err != nil {
		return fmt.Errorf("load project config: %w", err)
	}

	if devMode != "" && string(cfg.Quality.DevelopmentMode) != devMode {
		path := []string{"constitution", "development_mode"}
		if disk, ok := readSeamScalar(projectRoot, "quality", path); !ok || disk != devMode {
			if err := WriteSectionViaSeam(projectRoot, "quality", []yamlpatch.KeyEdit{
				{Path: path, Value: devMode},
			}); err != nil {
				return fmt.Errorf("persist development_mode: %w", err)
			}
		}
	}

	if convention != "" && cfg.GitConvention.Convention != convention {
		path := []string{"git_convention", "convention"}
		if disk, ok := readSeamScalar(projectRoot, "git-convention", path); !ok || disk != convention {
			if err := WriteSectionViaSeam(projectRoot, "git-convention", []yamlpatch.KeyEdit{
				{Path: path, Value: convention},
			}); err != nil {
				return fmt.Errorf("persist git_convention: %w", err)
			}
		}
	}

	return nil
}
