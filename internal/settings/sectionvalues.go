package settings

// 이 파일은 M2b 확장 섹션의 제네릭 읽기 seam이다 (SPEC-WEB-CONSOLE-011).
// 섹션 파일을 yaml 문서로 읽어 FieldDef.Path / ReadOnly / RawBlock 경로의 현재
// 값을 문자열로 추출한다. 읽기 전용이며 어떤 쓰기 경로도 갖지 않는다 — 쓰기는
// sectionapply.go(typed) / sectionwrite.go(seam)가 담당한다.

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// typedSectionFiles는 typed 섹션 이름 → 섹션 파일 base name 매핑이다.
var typedSectionFiles = map[string]string{
	"git_strategy": "git-strategy",
	"llm":          "llm",
	"quality":      "quality",
}

// sectionFileFor는 FieldDef의 영속화 대상 섹션 파일 base name을 반환한다.
func sectionFileFor(f FieldDef) string {
	switch f.Persist.Kind {
	case PersistSeam:
		return f.Persist.Section
	case PersistTypedSection:
		return typedSectionFiles[f.Persist.Section]
	default:
		return ""
	}
}

// fieldYAMLPath는 FieldDef의 섹션 파일 내 yaml 경로를 반환한다. seam 필드는
// Path 그대로, typed 필드는 파일 최상위 키 + dot-key 분해다 (quality.yaml의
// 최상위 키는 "constitution", git-strategy.yaml은 "git_strategy", llm.yaml은 "llm").
func fieldYAMLPath(f FieldDef) []string {
	if f.Persist.Kind == PersistSeam {
		return f.Persist.Path
	}
	root := f.Persist.Section
	if f.Persist.Section == "quality" {
		root = "constitution"
	}
	return append([]string{root}, strings.Split(f.Persist.Key, ".")...)
}

// SchemaCurrentValues는 M2b 확장 필드(PersistSeam + PersistTypedSection) 전체의
// 디스크 현재 값을 문자열 맵(FieldDef.Name → 스칼라 문자열)으로 반환한다.
// 파일/키 부재는 빈 문자열이다 (오류 아님 — greenfield 루트 허용).
func SchemaCurrentValues(projectRoot string) (map[string]string, error) {
	docs := map[string]*yaml.Node{}
	out := map[string]string{}
	for _, f := range allFields() {
		file := sectionFileFor(f)
		if file == "" {
			continue // profile-store / 기존 project-config 필드는 기존 read seam 소관
		}
		doc, err := loadSectionDoc(docs, projectRoot, file)
		if err != nil {
			return nil, err
		}
		out[f.Name] = scalarAtPath(doc, fieldYAMLPath(f))
	}
	for _, ro := range ReadOnlyDisplayFields() {
		doc, err := loadSectionDoc(docs, projectRoot, ro.File)
		if err != nil {
			return nil, err
		}
		out[ro.Name] = scalarAtPath(doc, ro.Path)
	}
	return out, nil
}

// RawBlockValues는 REQ-WC11-062 raw view 서브블록의 표시용 yaml 텍스트를
// 반환한다 (Name → 재직렬화 텍스트). 부재 블록은 빈 문자열이다. 표시 전용
// 재직렬화이므로 포매팅 정규화는 무해하다.
func RawBlockValues(projectRoot string) (map[string]string, error) {
	docs := map[string]*yaml.Node{}
	out := map[string]string{}
	for _, rb := range RawViewBlocks() {
		doc, err := loadSectionDoc(docs, projectRoot, rb.File)
		if err != nil {
			return nil, err
		}
		node := nodeAtPath(doc, rb.Path)
		if node == nil {
			out[rb.Name] = ""
			continue
		}
		raw, err := yaml.Marshal(node)
		if err != nil {
			return nil, fmt.Errorf("settings: marshal raw block %s: %w", rb.Name, err)
		}
		out[rb.Name] = string(raw)
	}
	return out, nil
}

// loadSectionDoc는 섹션 파일을 yaml 문서로 로드해 캐시한다. 파일 부재는 nil
// 문서(값 전부 빈 문자열)로 처리한다.
func loadSectionDoc(cache map[string]*yaml.Node, projectRoot, file string) (*yaml.Node, error) {
	if doc, ok := cache[file]; ok {
		return doc, nil
	}
	path := filepath.Join(projectRoot, ".moai", "config", "sections", file+".yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			cache[file] = nil
			return nil, nil
		}
		return nil, fmt.Errorf("settings: read section %s: %w", file, err)
	}
	var doc yaml.Node
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("settings: parse section %s: %w", file, err)
	}
	cache[file] = &doc
	return &doc, nil
}

// nodeAtPath는 문서에서 매핑 경로를 탐색한다. 부재/형상 불일치는 nil이다.
func nodeAtPath(doc *yaml.Node, path []string) *yaml.Node {
	if doc == nil || doc.Kind != yaml.DocumentNode || len(doc.Content) == 0 {
		return nil
	}
	cur := doc.Content[0]
	for _, key := range path {
		if cur == nil || cur.Kind != yaml.MappingNode {
			return nil
		}
		var next *yaml.Node
		for i := 0; i+1 < len(cur.Content); i += 2 {
			if cur.Content[i].Value == key {
				next = cur.Content[i+1]
				break
			}
		}
		cur = next
	}
	return cur
}

// scalarAtPath는 경로의 스칼라 값을 문자열로 반환한다 (부재/비-스칼라 → "").
func scalarAtPath(doc *yaml.Node, path []string) string {
	node := nodeAtPath(doc, path)
	if node == nil || node.Kind != yaml.ScalarNode {
		return ""
	}
	return node.Value
}
