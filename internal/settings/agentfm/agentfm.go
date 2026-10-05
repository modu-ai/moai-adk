// Package agentfm is a read-only scan layer over sub-agent definition files
// (.claude/agents/**/*.md): it lists each file's YAML frontmatter state for
// the moai web console's agent-overrides sub-section
// (SPEC-WEB-AGENTFM-RESTORE-001 M3; the surface SPEC-WEB-CONSOLE-011 M3
// originally built).
//
// File model: `---\n<frontmatter yaml>\n---\n<body>` — the layer identifies
// the frontmatter span by the two `---` delimiters and parses only that span;
// the body is never parsed. The directory scan reads LIVE files only; there is
// no template-mirror dual-write.
//
// READ-ONLY BY CONTRACT (REQ-AFR-005): the original package also carried a
// Patch write layer (upsertScalar/deleteKey frontmatter mutation). The console
// persist moved to llm.agent_overrides under SPEC-MODEL-PROFILE-MATRIX-001
// REQ-MPM-040, and that write layer is deliberately NOT re-ported —
// .claude/agents/**/*.md is a read-only scan surface and every save path must
// leave it byte-identical. Nothing in this package writes a file.
package agentfm

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// fmDelimiter는 frontmatter 구분선이다.
const fmDelimiter = "---\n"

// AgentInfo는 agent 파일 frontmatter의 표시 상태다.
type AgentInfo struct {
	Name          string // 파일 base name (확장자 제외)
	Path          string // 파일 경로
	Model         string // frontmatter model: 값 ("" = 키 부재)
	Effort        string // frontmatter effort: 값 ("" = 키 부재)
	EffortPresent bool   // effort 키 존재 여부 (부재는 유효 상태 — EC-7)
	ParseOK       bool   // frontmatter 파싱 성공 여부 (실패 행은 편집 비활성)
	Description   string // frontmatter description: 값 (role/tooltip 표시용, "" = 키 부재)
}

// List는 agentsDir의 *.md agent 파일 frontmatter 상태를 이름순으로 반환한다.
// 디렉터리 부재는 빈 슬라이스다(오류 아님 — greenfield 루트). 개별 파일의
// 파싱 실패는 ParseOK=false 행으로 강등한다 — 페이지 전체 실패 금지 (design.md
// §C.1 견고성).
func List(agentsDir string) ([]AgentInfo, error) {
	entries, err := os.ReadDir(agentsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("agentfm: read dir: %w", err)
	}
	var out []AgentInfo
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		path := filepath.Join(agentsDir, e.Name())
		info := AgentInfo{
			Name: strings.TrimSuffix(e.Name(), ".md"),
			Path: path,
		}
		if fm, _, err := splitFrontmatter(path); err == nil {
			var doc struct {
				Model       string  `yaml:"model"`
				Effort      *string `yaml:"effort"`
				Description string  `yaml:"description"`
			}
			if yaml.Unmarshal(fm, &doc) == nil {
				info.ParseOK = true
				info.Model = doc.Model
				info.Description = doc.Description
				if doc.Effort != nil {
					info.Effort = *doc.Effort
					info.EffortPresent = true
				}
			}
		}
		out = append(out, info)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}

// splitFrontmatter는 파일을 (frontmatter yaml bytes, body bytes)로 분리한다.
// body는 두 번째 `---\n` 직후의 원본 bytes 전체다.
func splitFrontmatter(path string) (fm, body []byte, err error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, fmt.Errorf("agentfm: read %s: %w", path, err)
	}
	if !bytes.HasPrefix(data, []byte(fmDelimiter)) {
		return nil, nil, fmt.Errorf("agentfm: %s: no frontmatter delimiter", path)
	}
	rest := data[len(fmDelimiter):]
	idx := bytes.Index(rest, []byte("\n---\n"))
	if idx < 0 {
		return nil, nil, fmt.Errorf("agentfm: %s: unterminated frontmatter", path)
	}
	fm = rest[:idx+1]                // 마지막 개행 포함
	body = rest[idx+len("\n---\n"):] // 구분선 직후 원본 bytes
	return fm, body, nil
}
