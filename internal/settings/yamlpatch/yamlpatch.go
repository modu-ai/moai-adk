// Package yamlpatch는 주석·미모델링 키·키 순서를 보존하는 YAML 부분 패치 seam을
// 제공한다 (SPEC-WEB-CONSOLE-011 M1, REQ-WC11-003/004/017).
//
// ConfigManager.Save()의 typed struct 재직렬화는 yaml 주석 전량과 미모델링 키
// (예: workflow.yaml `team.patterns`, role-profile `effort`)를 파괴한다. 본 패키지는
// gopkg.in/yaml.v3 노드 트리 수술로 대상 스칼라만 교체/생성(upsert)하고 나머지 문서
// 구조를 보존한다. Save() 경로가 없는 8개 섹션(workflow, harness, ralph, research,
// feedback, observability, security, db)의 유일한 쓰기 경로다 (REQ-WC11-017).
//
// 한계 (design.md §A.4): yaml.v3 Encoder는 재직렬화 시 일부 포매팅(빈 줄, 긴 스칼라
// 줄바꿈)을 정규화할 수 있다. byte-stability는 보증이 아니라 검증 대상이며, 섹션별
// golden round-trip 테스트가 그 범위를 고정한다. 노드 삭제는 지원하지 않는다.
package yamlpatch

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// KeyEdit은 yaml 문서 내 단일 스칼라 교체/생성을 기술한다 (design.md §A.2).
type KeyEdit struct {
	// Path는 문서 루트 매핑부터의 키 경로다.
	// 예: ["workflow", "team", "role_profiles", "implementer", "model"].
	Path []string
	// Value는 기록할 스칼라 값이다. 기존 노드의 인용 스타일(Style)은 보존을
	// 시도하고, plain 스칼라의 타입 태그는 인코더가 값에서 재해석한다.
	Value string
}

// PatchFile은 path의 YAML 파일을 yaml.Node 문서로 로드해 edits의 각 스칼라를
// 교체하고, 경로가 없으면 mapping 노드를 생성(upsert)한 뒤 원자적으로 재기록한다.
// 주석(Head/Line/FootComment)·키 순서·미모델링 키는 보존된다. upsert는 명시적으로
// 편집된 경로에만 적용되며(EC-3), 노드 삭제는 지원하지 않는다 (design.md §A.2).
// edits가 비어 있으면 파일을 건드리지 않고 nil을 반환한다.
//
// REQ-WWS-005 (SPEC-WEB-WRITE-SAFETY-001): when every edit targets an EXISTING
// scalar, the write takes the line-splice path (lineSplice) — only the target
// lines are rewritten and the rest of the original bytes (blank lines, comments,
// key order, unknown keys) survive untouched. The former re-encode path remains
// as the fallback for upserts and unresolvable edits; re-encoding normalizes
// blank lines away (the limitation documented in this package header), which
// M1(d) observed as the feedback.yaml blank-line loss.
//
// @MX:ANCHOR: [AUTO] PatchFile은 seam 섹션 yaml의 공유 부분-쓰기 진입점이다 —
// 호출 파일 3개 5호출점(sectionwrite, initializer_expansion ×3, init_workflow_flags)이 같은 계약에 의존한다.
// @MX:REASON: [AUTO] REQ-WWS-005 (SPEC-WEB-WRITE-SAFETY-001): 기존 스칼라 교체는
// lineSplice(대상 라인만 재작성 — 빈 줄·주석·키 순서·unknown key 원문 바이트 보존)를
// 먼저 시도하고, upsert·해소 불가 편집만 재직렬화 폴백으로 보낸다. 재직렬화는 빈 줄을
// 정규화해 버리므로(패키지 헤더 문서화 한계) 폴백 강제 뮤턴트는
// TestPatchFileValueInvariantPreservesBytes가 RED로 잡는다 — 이 분기 구조를
// 단순화하려는 시도는 이 테스트부터 읽는다.
func PatchFile(path string, edits []KeyEdit) error {
	if len(edits) == 0 {
		return nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("yamlpatch: read %s: %w", path, err)
		}
		// A missing section file starts from an empty mapping document. The
		// section loaders treat an absent file as defaults (greenfield
		// tolerance); the seam write mirrors that — the first edit creates
		// the file rather than erroring.
		data = []byte("{}\n")
	}

	// Fast path: every edit resolves to an existing scalar → splice only the
	// target lines, preserving every other byte of the original document.
	if out, ok, serr := lineSplice(data, edits); serr != nil {
		return fmt.Errorf("yamlpatch: %s: %w", path, serr)
	} else if ok {
		return atomicWrite(path, out)
	}

	var doc yaml.Node
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return fmt.Errorf("yamlpatch: parse %s: %w", path, err)
	}
	if doc.Kind != yaml.DocumentNode || len(doc.Content) == 0 {
		return fmt.Errorf("yamlpatch: %s: empty or non-document YAML", path)
	}
	root := doc.Content[0]
	if root.Kind != yaml.MappingNode {
		return fmt.Errorf("yamlpatch: %s: top-level node is not a mapping", path)
	}

	for _, e := range edits {
		if err := applyEdit(root, e); err != nil {
			return fmt.Errorf("yamlpatch: %s: %w", path, err)
		}
	}

	out, err := encode(&doc, detectIndent(data))
	if err != nil {
		return fmt.Errorf("yamlpatch: encode %s: %w", path, err)
	}
	return atomicWrite(path, out)
}

// lineSplice applies edits by rewriting only the source lines of the targeted
// scalars. It reports ok=false whenever the splice path cannot be proven
// correct — an unresolvable path (upsert), a non-scalar target, two edits on
// one line, or a post-splice re-parse whose values disagree — and the caller
// falls back to the re-encode path. The re-parse verification is the safety
// net for hand-rolled quoting: a spliced file that does not parse back to the
// requested values is never written.
func lineSplice(data []byte, edits []KeyEdit) ([]byte, bool, error) {
	var doc yaml.Node
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return nil, false, nil // unparseable input → let the main path report it
	}
	if doc.Kind != yaml.DocumentNode || len(doc.Content) == 0 {
		return nil, false, nil
	}
	root := doc.Content[0]
	if root.Kind != yaml.MappingNode {
		return nil, false, nil
	}

	type splice struct {
		lineIdx int
		node    *yaml.Node
		value   string
		desc    string
	}
	var splices []splice
	seenLines := map[int]bool{}
	for _, e := range edits {
		cur := root
		resolved := false
		for i, key := range e.Path {
			if cur.Kind != yaml.MappingNode {
				return nil, false, nil
			}
			idx := findKey(cur, key)
			if idx < 0 {
				return nil, false, nil // upsert — needs the re-encode path
			}
			val := cur.Content[idx+1]
			if i == len(e.Path)-1 {
				if val.Kind != yaml.ScalarNode {
					return nil, false, nil
				}
				if val.Line <= 0 || val.Line > len(strings.Split(string(data), "\n")) {
					return nil, false, nil
				}
				lineIdx := val.Line - 1 // yaml.Node.Line is 1-based
				if seenLines[lineIdx] {
					return nil, false, nil // two edits on one line — ambiguous splice
				}
				seenLines[lineIdx] = true
				splices = append(splices, splice{lineIdx: lineIdx, node: val, value: e.Value, desc: strings.Join(e.Path, ".")})
				resolved = true
				break
			}
			cur = val
		}
		if !resolved {
			return nil, false, nil
		}
	}

	lines := strings.Split(string(data), "\n")
	for _, sp := range splices {
		spliced, ok := replaceScalarInLine(lines[sp.lineIdx], sp.node, sp.value)
		if !ok {
			return nil, false, nil
		}
		lines[sp.lineIdx] = spliced
	}

	out := []byte(strings.Join(lines, "\n"))

	// Verify the spliced document parses back to exactly the requested values
	// at each edited path — the guard that makes hand-rolled quoting safe.
	var check yaml.Node
	if err := yaml.Unmarshal(out, &check); err != nil {
		return nil, false, nil
	}
	if check.Kind != yaml.DocumentNode || len(check.Content) == 0 {
		return nil, false, nil
	}
	checkRoot := check.Content[0]
	for _, sp := range splices {
		cur := checkRoot
		for _, key := range strings.Split(sp.desc, ".") {
			if cur.Kind != yaml.MappingNode {
				return nil, false, nil
			}
			idx := findKey(cur, key)
			if idx < 0 {
				return nil, false, nil
			}
			cur = cur.Content[idx+1]
		}
		if cur.Kind != yaml.ScalarNode || cur.Value != sp.value {
			return nil, false, nil
		}
	}
	return out, true, nil
}

// replaceScalarInLine rewrites the value token of a `key: value` line in place,
// keeping the key, indentation, and any trailing comment untouched. The new
// value is rendered in the original node's quoting style; the caller verifies
// the result by re-parsing, so a style that cannot render safely falls back
// (ok=false) rather than writing a corrupted line.
func replaceScalarInLine(line string, node *yaml.Node, value string) (string, bool) {
	if !strings.Contains(line, ":") {
		return "", false
	}
	oldRendered := renderScalar(node, node.Value)
	if oldRendered == "" {
		return "", false
	}
	// Replace only the first occurrence AFTER the key separator so a value that
	// happens to embed the key text is not mangled.
	colon := strings.Index(line, ":")
	head, tail := line[:colon+1], line[colon+1:]
	pos := strings.Index(tail, oldRendered)
	if pos < 0 {
		return "", false
	}
	newRendered := renderScalar(&yaml.Node{Style: node.Style, Value: value}, value)
	if newRendered == "" {
		return "", false
	}
	return head + tail[:pos] + newRendered + tail[pos+len(oldRendered):], true
}

// renderScalar renders a scalar's value in the quoting style carried by the
// node (or, for a synthetic node, the style it was given). Plain style is
// returned as-is; the re-parse verification in lineSplice is what rejects a
// plain value that would not round-trip.
func renderScalar(node *yaml.Node, value string) string {
	switch node.Style {
	case yaml.DoubleQuotedStyle:
		return strconv.Quote(value)
	case yaml.SingleQuotedStyle:
		return "'" + strings.ReplaceAll(value, "'", "''") + "'"
	default:
		return value
	}
}

// applyEdit은 root 매핑에서 e.Path를 탐색해 최종 스칼라를 교체한다. 중간/최종
// 경로 세그먼트가 없으면 생성한다(upsert). 최종 대상이 스칼라가 아니면 오류다.
func applyEdit(root *yaml.Node, e KeyEdit) error {
	if len(e.Path) == 0 {
		return errors.New("empty edit path")
	}
	pathDesc := strings.Join(e.Path, ".")

	cur := root
	for i, key := range e.Path {
		last := i == len(e.Path)-1
		if cur.Kind != yaml.MappingNode {
			return fmt.Errorf("path %q: segment %q parent is not a mapping", pathDesc, key)
		}

		idx := findKey(cur, key)
		if idx < 0 {
			// upsert: 부재 키 생성 — 최종 세그먼트면 스칼라, 중간이면 매핑.
			keyNode := &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: key}
			var valNode *yaml.Node
			if last {
				valNode = &yaml.Node{Kind: yaml.ScalarNode, Value: e.Value}
			} else {
				valNode = &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
			}
			cur.Content = append(cur.Content, keyNode, valNode)
			if last {
				return nil
			}
			cur = valNode
			continue
		}

		val := cur.Content[idx+1]
		if last {
			return setScalar(val, e.Value, pathDesc)
		}
		cur = val
	}
	return nil
}

// findKey는 매핑 노드에서 key와 일치하는 키 노드의 Content 인덱스를 반환한다.
// 없으면 -1. 매핑 Content는 [k0, v0, k1, v1, ...] 형상이다.
func findKey(mapping *yaml.Node, key string) int {
	for i := 0; i+1 < len(mapping.Content); i += 2 {
		if mapping.Content[i].Value == key {
			return i
		}
	}
	return -1
}

// setScalar는 기존 스칼라 노드의 값을 교체한다. double/single-quoted 스타일은
// 문자열 태그와 함께 보존하고, plain 스칼라는 태그를 비워 인코더가 새 값에서
// 타입(!!bool/!!int/!!float/!!str)을 재해석하게 한다.
func setScalar(n *yaml.Node, value, pathDesc string) error {
	if n.Kind != yaml.ScalarNode {
		return fmt.Errorf("path %q: target is not a scalar", pathDesc)
	}
	n.Value = value
	switch n.Style {
	case yaml.DoubleQuotedStyle, yaml.SingleQuotedStyle:
		n.Tag = "!!str" // 인용 형상 유지 — 문자열로 고정.
	default:
		n.Tag = ""  // 인코더 재해석.
		n.Style = 0 // plain 유지 (필요 시 인코더가 안전 스타일 선택).
	}
	return nil
}

// indentRe는 첫 들여쓰기 라인의 선행 공백을 잡는다 (파일별 indent 폭 검출).
var indentRe = regexp.MustCompile(`(?m)^( +)\S`)

// detectIndent는 원본 파일의 들여쓰기 폭을 검출한다. 검출 불가 시 섹션 yaml
// 관례인 4를 쓰고, yaml.v3 인코더 제약에 맞춰 최소 2로 클램프한다.
func detectIndent(data []byte) int {
	m := indentRe.FindSubmatch(data)
	if m == nil {
		return 4
	}
	n := len(m[1])
	if n < 2 {
		return 2
	}
	return n
}

// encode는 문서 노드를 지정 indent로 재직렬화한다.
func encode(doc *yaml.Node, indent int) ([]byte, error) {
	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(indent)
	if err := enc.Encode(doc); err != nil {
		_ = enc.Close()
		return nil, err
	}
	if err := enc.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// defaultFilePerm은 absent 대상에 대한 문서화된 기본 모드다
// (SPEC-SEAM-GREENFIELD-001 §4). 템플릿 섹션 파일의 측정된 관례는 0644다
// (spec.md C7); os.CreateTemp의 0600은 이후 편집의 stat 보존이 그 어긋남을
// 영구 보존하므로 기각됐다 (§4 근거 기록).
const defaultFilePerm os.FileMode = 0o644

// atomicWrite는 동일 디렉터리 temp 파일 + rename으로 원자적으로 기록한다
// (ConfigManager.Save의 원자성 관례를 따른다). 원본 파일 모드를 보존한다.
//
// @MX:NOTE: [AUTO] absent 대상은 greenfield로 생성한다 — 읽기 계층
// (PatchFile)이 absent를 빈 문서로 시작하므로 쓰기도 그 계약을 거울처럼
// 따른다 (REQ-1). absent에는 보존할 모드가 없어 패키지 단일 정의점인
// defaultFilePerm(0644)으로 만든다 (REQ-2); 그 외 stat 오류(ENOTDIR·권한)는
// 기존 래핑으로 유지된다 (REQ-3). absent 경로도 직접 쓰기로 전환하지 않고
// temp+rename을 유지한다 (REQ-4).
// @MX:SPEC: SPEC-SEAM-GREENFIELD-001
func atomicWrite(path string, data []byte) error {
	mode := defaultFilePerm
	info, err := os.Stat(path)
	if err != nil {
		// REQ-1 (SPEC-SEAM-GREENFIELD-001): an absent target is greenfield —
		// the read layer starts from an empty document, so the write creates
		// the file instead of failing on the stat. Any other stat error
		// (ENOTDIR, permissions) keeps its wrapped error (REQ-3).
		if !os.IsNotExist(err) {
			return fmt.Errorf("yamlpatch: stat %s: %w", path, err)
		}
	} else {
		mode = info.Mode().Perm()
	}
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".yamlpatch-*.tmp")
	if err != nil {
		return fmt.Errorf("yamlpatch: create temp: %w", err)
	}
	tmpName := tmp.Name()
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpName)
		return fmt.Errorf("yamlpatch: write temp: %w", err)
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpName)
		return fmt.Errorf("yamlpatch: close temp: %w", err)
	}
	if err := os.Chmod(tmpName, mode); err != nil {
		_ = os.Remove(tmpName)
		return fmt.Errorf("yamlpatch: chmod temp: %w", err)
	}
	if err := os.Rename(tmpName, path); err != nil {
		_ = os.Remove(tmpName)
		return fmt.Errorf("yamlpatch: rename: %w", err)
	}
	return nil
}
