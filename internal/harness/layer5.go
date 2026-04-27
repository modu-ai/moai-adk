// Package harness — Layer 5 user directory scaffolder.
package harness

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ScaffoldOpts controls which files ScaffoldHarnessDir creates and what
// project metadata is embedded in their headers.
type ScaffoldOpts struct {
	// Domain is the project domain (e.g. "ios-mobile") captured during the
	// Phase 5 Socratic interview.
	Domain string
	// SpecID is the project-init SPEC identifier (e.g. "SPEC-PROJ-INIT-001").
	SpecID string
	// IncludeDesignExtension creates design-extension.md when Q13 = "Advanced"
	// (REQ-PH-012). When false, only the 7 baseline files are created.
	IncludeDesignExtension bool
}

// scaffoldFile describes a single file produced by ScaffoldHarnessDir.
type scaffoldFile struct {
	name    string
	purpose string
	body    string
}

// ScaffoldHarnessDir creates the .moai/harness/ user directory layout. The
// caller is responsible for path validation; layer code does NOT call
// EnsureAllowed. Tests freely use t.TempDir() paths.
//
// Files always created (7):
//   - main.md, plan-extension.md, run-extension.md, sync-extension.md,
//     chaining-rules.yaml, interview-results.md, README.md
//
// File created only when opts.IncludeDesignExtension is true (8th file):
//   - design-extension.md
//
// Each file's first non-empty line states its purpose so that human readers
// can identify the editable zone at a glance.
func ScaffoldHarnessDir(harnessDir string, opts ScaffoldOpts) error {
	if harnessDir == "" {
		return errors.New("ScaffoldHarnessDir: empty harness directory")
	}
	if err := os.MkdirAll(harnessDir, 0o755); err != nil {
		return fmt.Errorf("ScaffoldHarnessDir: mkdir %s: %w", harnessDir, err)
	}
	files := scaffoldFiles(opts)
	for _, f := range files {
		path := filepath.Join(harnessDir, f.name)
		if err := os.WriteFile(path, []byte(f.body), 0o644); err != nil {
			return fmt.Errorf("ScaffoldHarnessDir: write %s: %w", path, err)
		}
	}
	return nil
}

func scaffoldFiles(opts ScaffoldOpts) []scaffoldFile {
	now := time.Now().UTC().Format("2006-01-02")
	files := []scaffoldFile{
		{
			name:    "main.md",
			purpose: "CLAUDE.md @import 진입점 — 프로젝트 메타데이터 + 도메인 요약",
			body:    mainMD(opts, now),
		},
		{
			name:    "plan-extension.md",
			purpose: "manager-spec chain 명시 — /moai plan 시 활성",
			body:    extensionMD("plan", opts, "manager-spec"),
		},
		{
			name:    "run-extension.md",
			purpose: "manager-tdd / manager-ddd chain rules — /moai run 시 활성",
			body:    extensionMD("run", opts, "manager-tdd"),
		},
		{
			name:    "sync-extension.md",
			purpose: "manager-docs chain rules — /moai sync 시 활성",
			body:    extensionMD("sync", opts, "manager-docs"),
		},
		{
			name:    "chaining-rules.yaml",
			purpose: "machine-readable chain rules — manager-tdd가 read",
			body:    chainingRulesPlaceholder(opts),
		},
		{
			name:    "interview-results.md",
			purpose: "Phase 5 인터뷰 답변 — Buffer.Commit 후 WriteResultsToFile이 갱신",
			body:    interviewResultsPlaceholder(opts),
		},
		{
			name:    "README.md",
			purpose: "사용자 가시성 — 5-Layer 설명 + 편집 가능 영역 표시",
			body:    readmeMD(opts),
		},
	}
	if opts.IncludeDesignExtension {
		files = append(files, scaffoldFile{
			name:    "design-extension.md",
			purpose: "/moai design 워크플로우 확장 — Q13='Advanced' 분기에서만 생성 (REQ-PH-012)",
			body:    extensionMD("design", opts, "expert-frontend"),
		})
	}
	return files
}

func mainMD(opts ScaffoldOpts, now string) string {
	var b strings.Builder
	b.WriteString("# Harness Main\n")
	b.WriteString("<!-- 진입점: CLAUDE.md @import가 이 파일을 따라옵니다. -->\n\n")
	fmt.Fprintf(&b, "**Domain**: %s\n", opts.Domain)
	fmt.Fprintf(&b, "**SPEC**: %s\n", opts.SpecID)
	fmt.Fprintf(&b, "**Updated**: %s\n\n", now)
	b.WriteString("## Domain Summary\n\n")
	b.WriteString("이 프로젝트는 ")
	fmt.Fprintf(&b, "%s 도메인 기반입니다. 도메인 특화 패턴은 my-harness-* skills에 정의됩니다.\n\n", opts.Domain)
	b.WriteString("## Linked Files\n\n")
	b.WriteString("- `plan-extension.md` — Plan phase chain\n")
	b.WriteString("- `run-extension.md` — Run phase chain\n")
	b.WriteString("- `sync-extension.md` — Sync phase chain\n")
	b.WriteString("- `chaining-rules.yaml` — machine-readable rules\n")
	b.WriteString("- `interview-results.md` — original interview answers\n")
	return b.String()
}

func extensionMD(phase string, opts ScaffoldOpts, primaryAgent string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# %s Phase Harness Extension\n", capitalize(phase))
	fmt.Fprintf(&b, "<!-- %s 단계에서 자동 활성. 사용자 편집 가능 영역. -->\n\n", phase)
	fmt.Fprintf(&b, "**Primary agent**: `%s`\n", primaryAgent)
	fmt.Fprintf(&b, "**Domain**: %s\n", opts.Domain)
	fmt.Fprintf(&b, "**SPEC origin**: %s\n\n", opts.SpecID)
	b.WriteString("## Chain Rules\n\n")
	b.WriteString("```yaml\n")
	fmt.Fprintf(&b, "phase: %s\nwhen:\n  agent: %s\ninsert_before: []\ninsert_after: []\n", phase, primaryAgent)
	b.WriteString("```\n\n")
	b.WriteString("## Notes\n\n")
	b.WriteString("- 이 파일은 사용자가 자유롭게 편집할 수 있습니다.\n")
	b.WriteString("- `moai update`는 이 파일을 절대 수정하지 않습니다 (REQ-PH-009).\n")
	return b.String()
}

func chainingRulesPlaceholder(opts ScaffoldOpts) string {
	var b strings.Builder
	b.WriteString("# .moai/harness/chaining-rules.yaml\n")
	b.WriteString("# machine-readable chain rules — manager-tdd가 read\n")
	fmt.Fprintf(&b, "# Generated by Phase 5 interview for %s\n\n", opts.SpecID)
	b.WriteString("version: 1\n")
	b.WriteString("chains:\n")
	b.WriteString("  - phase: run\n")
	b.WriteString("    when:\n")
	b.WriteString("      agent: manager-tdd\n")
	b.WriteString("    insert_before: []\n")
	b.WriteString("    insert_after: []\n")
	return b.String()
}

func interviewResultsPlaceholder(opts ScaffoldOpts) string {
	var b strings.Builder
	b.WriteString("---\n")
	fmt.Fprintf(&b, "spec_id: %s\n", opts.SpecID)
	b.WriteString("generated_at: TBD\n")
	b.WriteString("project_root: TBD\n")
	b.WriteString("conversation_language: ko\n")
	b.WriteString("---\n\n")
	b.WriteString("# Interview Results (placeholder)\n")
	b.WriteString("<!-- Phase 5 인터뷰 후 WriteResultsToFile이 이 파일을 갱신합니다. -->\n")
	return b.String()
}

func readmeMD(opts ScaffoldOpts) string {
	var b strings.Builder
	b.WriteString("# .moai/harness/ — User Customization Zone\n")
	b.WriteString("<!-- 사용자 편집 가능. moai update는 이 디렉터리를 절대 수정하지 않음 (REQ-PH-009). -->\n\n")
	fmt.Fprintf(&b, "**Domain**: %s · **SPEC**: %s\n\n", opts.Domain, opts.SpecID)
	b.WriteString("## 5-Layer Integration\n\n")
	b.WriteString("이 디렉터리는 Layer 5 (사용자 영역) 콘텐츠입니다. 5-Layer 통합 장치:\n\n")
	b.WriteString("- **L1**: my-harness-* skill frontmatter triggers (paths/keywords/agents/phases)\n")
	b.WriteString("- **L2**: `.moai/config/sections/workflow.yaml` `harness:` 섹션\n")
	b.WriteString("- **L3**: `CLAUDE.md` `<!-- moai:harness-start -->` ~ `<!-- moai:harness-end -->` marker\n")
	b.WriteString("- **L4**: `.claude/skills/moai/workflows/{plan,run,sync,design}.md` 정적 import line\n")
	b.WriteString("- **L5**: `.moai/harness/` (이 디렉터리)\n\n")
	b.WriteString("## Editable Files\n\n")
	b.WriteString("- `main.md` — CLAUDE.md @import 진입점 (편집 가능)\n")
	b.WriteString("- `*-extension.md` — phase별 chain 확장 (편집 가능)\n")
	b.WriteString("- `chaining-rules.yaml` — chain rules (편집 가능, schema 준수)\n")
	b.WriteString("- `interview-results.md` — 인터뷰 답변 (참조용, 편집 비권장)\n")
	return b.String()
}

func capitalize(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
