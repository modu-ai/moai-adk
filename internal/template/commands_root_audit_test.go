package template

import (
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// TestRootLevelCommandsThinPattern verifies that root-level slash command files
// in the project's .claude/commands/ directory (not in the embedded template tree)
// follow the thin command pattern: YAML frontmatter with required fields,
// a body of less than 20 non-empty lines, a Skill() invocation, and for any
// Skill("<name>") reference, the corresponding .claude/skills/<name>/ directory
// must exist (partial migration gate).
//
// Source: SPEC-V3R2-WF-002 REQ-WF002-001, REQ-WF002-002, REQ-WF002-015
func TestRootLevelCommandsThinPattern(t *testing.T) {
	t.Parallel()

	// 이 테스트 파일 위치(internal/template/)에서 프로젝트 루트로 두 단계 상승
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller(0) failed")
	}
	projectRoot := filepath.Join(filepath.Dir(currentFile), "..", "..")
	commandsDir := filepath.Join(projectRoot, ".claude", "commands")
	skillsDir := filepath.Join(projectRoot, ".claude", "skills")

	// 경로 존재 확인
	if _, err := os.Stat(commandsDir); os.IsNotExist(err) {
		t.Fatalf("commands directory does not exist: %s", commandsDir)
	}

	fsys := os.DirFS(commandsDir)

	// 첫 번째 레벨(.md, .md.tmpl)만 수집 — agency/ 등 서브디렉토리는 제외
	var cmdFiles []string
	err := fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		// 루트 이외의 디렉토리는 진입하지 않음
		if d.IsDir() {
			if path != "." {
				return fs.SkipDir
			}
			return nil
		}
		if strings.HasSuffix(path, ".md") || strings.HasSuffix(path, ".md.tmpl") {
			cmdFiles = append(cmdFiles, path)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("WalkDir error: %v", err)
	}

	if len(cmdFiles) == 0 {
		t.Fatal("no root-level command files found under .claude/commands/")
	}

	for _, path := range cmdFiles {
		path := path // 루프 변수 캡처
		t.Run(path, func(t *testing.T) {
			t.Parallel()

			data, readErr := fs.ReadFile(fsys, path)
			if readErr != nil {
				t.Fatalf("ReadFile(%q) error: %v", path, readErr)
			}

			content := string(data)
			fm, body, parseErr := parseFrontmatterAndBody(content)
			if parseErr != "" {
				t.Fatalf("parse error: %s", parseErr)
			}

			// R1: 필수 frontmatter 필드 존재 확인
			if _, ok := fm["description"]; !ok {
				t.Error("missing required frontmatter field: description")
			}
			if _, ok := fm["allowed-tools"]; !ok {
				t.Error("missing required frontmatter field: allowed-tools")
			}

			// R2: allowed-tools는 CSV 문자열이어야 함 (YAML 배열 불가)
			if at, ok := fm["allowed-tools"]; ok {
				if strings.HasPrefix(strings.TrimSpace(at), "-") {
					t.Error("allowed-tools must be CSV string, not YAML array")
				}
			}

			// R3: body 비어있지 않은 줄 수 < 20 (REQ-WF002-001, REQ-WF002-002)
			bodyLines := countNonEmptyLines(body)
			if bodyLines >= 20 {
				t.Errorf("body has %d non-empty lines (max 19 for thin commands)", bodyLines)
			}

			// R4: body에 Skill() 호출 포함
			if !strings.Contains(body, "Skill(") {
				t.Errorf("body does not contain Skill() invocation")
			}

			// R5: 부분 마이그레이션 게이트 (REQ-WF002-015)
			// body에 Skill("<name>") 참조가 있을 때 .claude/skills/<name>/ 디렉토리가 존재해야 함
			checkSkillDirExists(t, body, skillsDir, path)
		})
	}

	t.Logf("audited %d root-level command files", len(cmdFiles))
}

// checkSkillDirExists는 body에서 Skill("...") 참조를 추출하고,
// 대응하는 .claude/skills/<name>/ 디렉토리가 존재하는지 확인한다.
// REQ-WF002-015 부분 마이그레이션 게이트.
func checkSkillDirExists(t *testing.T, body, skillsDir, cmdPath string) {
	t.Helper()

	// body에서 Skill("name") 패턴 추출
	remaining := body
	for {
		idx := strings.Index(remaining, `Skill("`)
		if idx < 0 {
			break
		}
		after := remaining[idx+len(`Skill("`):]
		end := strings.Index(after, `"`)
		if end < 0 {
			break
		}
		skillName := after[:end]
		if skillName != "" {
			skillPath := filepath.Join(skillsDir, skillName)
			if _, err := os.Stat(skillPath); os.IsNotExist(err) {
				t.Errorf(
					"THIN_WRAPPER_PARTIAL_MIGRATION: %s references Skill(%q) but .claude/skills/%s/ does not exist (REQ-WF002-015)",
					cmdPath, skillName, skillName,
				)
			}
		}
		remaining = after[end+1:]
	}
}
