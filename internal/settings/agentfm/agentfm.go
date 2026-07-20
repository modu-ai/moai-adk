// Package agentfm은 sub-agent 파일(.claude/agents/moai/*.md)의 YAML frontmatter
// 전용 read/patch 레이어다 (SPEC-WEB-CONSOLE-011 M3, REQ-WC11-025/027..029).
//
// 파일 모델: `---\n<frontmatter yaml>\n---\n<body>` — 레이어는 첫 두 `---` 구분선
// 으로 frontmatter 구간만 식별해 yaml.Node로 패치하고, body는 파싱조차 하지 않고
// 원본 bytes를 그대로 이어붙여 재조립한다 (byte 보존이 구조적으로 보장, design.md
// §C.1). 쓰기는 LIVE 파일만 대상이며 template 미러로의 자동 dual-write는 없다
// (REQ-WC11-027 — 미러 정합은 메인테이너 수동 판단).
//
// 연산 집합: model/effort 스칼라 교체 + upsert + effort 키 삭제 (EC-7 —
// effort 부재는 유효 상태이므로 "(absent)"로 되돌리기가 필수 연산이다). 섹션
// seam(yamlpatch)과 달리 삭제를 지원하는 유일한 레이어다.
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

// AgentInfo는 agent 파일 frontmatter의 표시/편집 대상 상태다.
type AgentInfo struct {
	Name          string // 파일 base name (확장자 제외)
	Path          string // 파일 절대/상대 경로
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

// Patch는 agent 파일의 frontmatter만 패치한다. model/effort가 빈 문자열이면
// 해당 키는 건드리지 않는다(preserve). deleteEffort가 true면 effort 키를
// 제거한다 (EC-7 "(absent)" 복귀; effort 값과 동시 지정 불가).
// body는 원본 bytes 그대로 재조립된다 — 동일 패치 2회는 byte-identical이다
// (AC-WC11-027 idempotency).
func Patch(path, model, effort string, deleteEffort bool) error {
	if deleteEffort && effort != "" {
		return fmt.Errorf("agentfm: cannot both set and delete effort")
	}
	if model == "" && effort == "" && !deleteEffort {
		return nil // no-op
	}

	fm, body, err := splitFrontmatter(path)
	if err != nil {
		return err
	}

	var doc yaml.Node
	if err := yaml.Unmarshal(fm, &doc); err != nil {
		return fmt.Errorf("agentfm: parse frontmatter %s: %w", path, err)
	}
	if doc.Kind != yaml.DocumentNode || len(doc.Content) == 0 || doc.Content[0].Kind != yaml.MappingNode {
		return fmt.Errorf("agentfm: %s: frontmatter is not a yaml mapping", path)
	}
	root := doc.Content[0]

	if model != "" {
		upsertScalar(root, "model", model)
	}
	if effort != "" {
		upsertScalar(root, "effort", effort)
	}
	if deleteEffort {
		deleteKey(root, "effort")
	}

	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(&doc); err != nil {
		_ = enc.Close()
		return fmt.Errorf("agentfm: encode frontmatter: %w", err)
	}
	if err := enc.Close(); err != nil {
		return fmt.Errorf("agentfm: close encoder: %w", err)
	}

	// 재조립: --- + frontmatter + --- + 원본 body bytes (무접촉).
	var out bytes.Buffer
	out.WriteString(fmDelimiter)
	out.Write(buf.Bytes())
	out.WriteString(fmDelimiter)
	out.Write(body)

	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("agentfm: stat %s: %w", path, err)
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".agentfm-*.tmp")
	if err != nil {
		return fmt.Errorf("agentfm: create temp: %w", err)
	}
	tmpName := tmp.Name()
	if _, err := tmp.Write(out.Bytes()); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpName)
		return fmt.Errorf("agentfm: write temp: %w", err)
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpName)
		return fmt.Errorf("agentfm: close temp: %w", err)
	}
	if err := os.Chmod(tmpName, info.Mode().Perm()); err != nil {
		_ = os.Remove(tmpName)
		return fmt.Errorf("agentfm: chmod temp: %w", err)
	}
	if err := os.Rename(tmpName, path); err != nil {
		_ = os.Remove(tmpName)
		return fmt.Errorf("agentfm: rename: %w", err)
	}
	return nil
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

// upsertScalar는 매핑의 스칼라 키를 교체하거나 없으면 추가한다. 기존 노드의
// Style/주석은 보존된다.
func upsertScalar(mapping *yaml.Node, key, value string) {
	for i := 0; i+1 < len(mapping.Content); i += 2 {
		if mapping.Content[i].Value == key {
			val := mapping.Content[i+1]
			val.Value = value
			if val.Style != yaml.DoubleQuotedStyle && val.Style != yaml.SingleQuotedStyle {
				val.Tag = ""
				val.Style = 0
			} else {
				val.Tag = "!!str"
			}
			return
		}
	}
	mapping.Content = append(mapping.Content,
		&yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: key},
		&yaml.Node{Kind: yaml.ScalarNode, Value: value},
	)
}

// deleteKey는 매핑에서 키(+값) 쌍을 제거한다. 부재 시 no-op.
func deleteKey(mapping *yaml.Node, key string) {
	for i := 0; i+1 < len(mapping.Content); i += 2 {
		if mapping.Content[i].Value == key {
			mapping.Content = append(mapping.Content[:i], mapping.Content[i+2:]...)
			return
		}
	}
}
