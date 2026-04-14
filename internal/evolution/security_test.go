package evolution_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/evolution"
)

// --- CRITICAL 1: CheckFrozenGuard 경로 순회 취약점 재현 테스트 ---

// TestCheckFrozenGuard_RejectsPathTraversal verifies that CheckFrozenGuard rejects
// path traversal and absolute path attacks.
func TestCheckFrozenGuard_RejectsPathTraversal(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		targetFile  string
		expectBlock bool
	}{
		{
			name:        "../../../etc/passwd 거부",
			targetFile:  "../../../etc/passwd",
			expectBlock: true,
		},
		{
			name:        ".. 우회 경로 거부",
			targetFile:  ".claude/rules/moai/core/../../../CLAUDE.md",
			expectBlock: true,
		},
		{
			name:        "절대 경로 거부",
			targetFile:  "/etc/passwd",
			expectBlock: true,
		},
		{
			name:        "유효한 스킬 파일 허용",
			targetFile:  ".claude/skills/moai-lang-go/SKILL.md",
			expectBlock: false,
		},
		{
			name:        "./로 시작하는 유효 경로 허용",
			targetFile:  ".claude/skills/moai-lang-go/SKILL.md",
			expectBlock: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := evolution.CheckFrozenGuard(tt.targetFile)
			if tt.expectBlock && err == nil {
				t.Errorf("CheckFrozenGuard(%q): 차단 예상, 하지만 nil 반환", tt.targetFile)
			}
			if !tt.expectBlock && err != nil {
				t.Errorf("CheckFrozenGuard(%q): 허용 예상, 하지만 오류 반환: %v", tt.targetFile, err)
			}
			if tt.expectBlock && err != nil && err != evolution.ErrFrozenPath {
				t.Errorf("CheckFrozenGuard(%q): ErrFrozenPath 예상, 하지만 반환: %v", tt.targetFile, err)
			}
		})
	}
}

// TestApplyProposal_RejectsProjectRootEscape verifies that ApplyProposal cannot
// write files outside projectRoot even with path traversal in TargetFile.
func TestApplyProposal_RejectsProjectRootEscape(t *testing.T) {
	t.Parallel()

	projectRoot := t.TempDir()
	// 외부 디렉터리 생성 (프로젝트 루트의 형제)
	outsideRoot := t.TempDir()

	// outsideRoot 내 파일을 대상으로 하는 상대 경로 계산
	// filepath.Rel 로 계산하면 ../../tmp/xxx/evil.md 형태
	rel, err := filepath.Rel(projectRoot, filepath.Join(outsideRoot, "evil.md"))
	if err != nil {
		t.Fatalf("filepath.Rel: %v", err)
	}

	proposal := &evolution.ProposedChange{
		TargetFile: rel,
		ZoneID:     "some-zone",
		Addition:   "evil content\n",
	}

	applyErr := evolution.ApplyProposal(projectRoot, proposal)
	if applyErr == nil {
		t.Error("ApplyProposal: 프로젝트 루트 탈출 시 오류 예상, 하지만 nil 반환")
	}

	// 외부 경로에 파일이 생성되지 않아야 함
	if _, statErr := os.Stat(filepath.Join(outsideRoot, "evil.md")); !os.IsNotExist(statErr) {
		t.Error("ApplyProposal: 프로젝트 루트 외부에 파일이 생성되면 안 됨")
	}
}

// --- CRITICAL 2: LearningEntry.ID 경로 순회 취약점 재현 테스트 ---

// TestCreateLearning_RejectsPathTraversalID verifies that CreateLearning rejects
// IDs that would escape the learnings directory via path traversal.
func TestCreateLearning_RejectsPathTraversalID(t *testing.T) {
	t.Parallel()

	projectRoot := t.TempDir()
	mustInitMoAI(t, projectRoot)

	// 에이전트 파일을 덮어쓸 수 있는 악성 ID
	maliciousID := "../../.claude/agents/evil"

	entry := &evolution.LearningEntry{
		ID:      maliciousID,
		SkillID: "moai-lang-go",
		ZoneID:  "best-practices",
	}

	err := evolution.CreateLearning(projectRoot, entry)
	if err == nil {
		t.Error("CreateLearning: 경로 순회 ID에 오류 예상, 하지만 nil 반환")
	}
	if err != evolution.ErrInvalidLearningID {
		t.Errorf("CreateLearning: ErrInvalidLearningID 예상, 하지만 반환: %v", err)
	}

	// 순회 경로에 파일이 생성되지 않아야 함
	escapedPath := filepath.Join(projectRoot, ".moai", "evolution", "learnings", maliciousID+".md")
	if _, statErr := os.Stat(escapedPath); !os.IsNotExist(statErr) {
		t.Error("CreateLearning: 경로 순회 ID로 파일이 생성되면 안 됨")
	}
}

// TestCreateLearning_AcceptsValidID verifies that valid LEARN-YYYYMMDD-NNN IDs pass.
func TestCreateLearning_AcceptsValidID(t *testing.T) {
	t.Parallel()

	projectRoot := t.TempDir()
	mustInitMoAI(t, projectRoot)

	validIDs := []string{
		"LEARN-20260415-001",
		"LEARN-20260415-999",
		"LEARN-20000101-000",
	}

	for _, id := range validIDs {
		id := id
		t.Run(id, func(t *testing.T) {
			t.Parallel()
			entry := &evolution.LearningEntry{
				ID:      id,
				SkillID: "moai-lang-go",
				ZoneID:  "best-practices",
			}
			if err := evolution.CreateLearning(projectRoot, entry); err != nil {
				t.Errorf("CreateLearning(%q): 유효한 ID에 오류 반환: %v", id, err)
			}
		})
	}
}

// TestCreateLearning_RejectsInvalidIDFormats verifies various invalid ID formats.
func TestCreateLearning_RejectsInvalidIDFormats(t *testing.T) {
	t.Parallel()

	projectRoot := t.TempDir()
	mustInitMoAI(t, projectRoot)

	invalidIDs := []string{
		"../../etc/passwd",
		"learn-20260415-001",      // 소문자 LEARN
		"LEARN-2026415-001",       // 날짜 형식 오류
		"LEARN-20260415-01",       // NNN이 2자리
		"LEARN-20260415-0001",     // NNN이 4자리
		"LEARN-20260415-001/evil", // 슬래시 포함
		"",                        // 빈 문자열
		"LEARN-20260415-001.bak",  // 확장자 포함
	}

	for _, id := range invalidIDs {
		id := id
		t.Run("invalid_"+strings.ReplaceAll(id, "/", "_"), func(t *testing.T) {
			t.Parallel()
			entry := &evolution.LearningEntry{
				ID:      id,
				SkillID: "moai-lang-go",
				ZoneID:  "best-practices",
			}
			err := evolution.CreateLearning(projectRoot, entry)
			if err == nil {
				t.Errorf("CreateLearning(%q): 잘못된 ID에 오류 예상, 하지만 nil 반환", id)
			}
			if err != evolution.ErrInvalidLearningID {
				t.Errorf("CreateLearning(%q): ErrInvalidLearningID 예상, 하지만 반환: %v", id, err)
			}
		})
	}
}

// --- CRITICAL 4: FrozenGuard Agency 경로 누락 재현 테스트 ---

// TestCheckFrozenGuard_ProtectsAgencyConstitution verifies that agency constitution
// files, fork-manifest, and lsp-client.md are blocked.
func TestCheckFrozenGuard_ProtectsAgencyConstitution(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		targetFile string
	}{
		{
			name:       "Agency 헌법 파일 차단",
			targetFile: ".claude/rules/agency/constitution.md",
		},
		{
			name:       "Agency 규칙 디렉터리 내 다른 파일 차단",
			targetFile: ".claude/rules/agency/some-other-rule.md",
		},
		{
			name:       "fork-manifest.yaml 차단",
			targetFile: ".agency/fork-manifest.yaml",
		},
		{
			name:       "lsp-client.md 차단 (.claude/rules/moai/core/ 접두사)",
			targetFile: ".claude/rules/moai/core/lsp-client.md",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := evolution.CheckFrozenGuard(tt.targetFile)
			if err == nil {
				t.Errorf("CheckFrozenGuard(%q): ErrFrozenPath 예상, 하지만 nil 반환", tt.targetFile)
			}
			if err != evolution.ErrFrozenPath {
				t.Errorf("CheckFrozenGuard(%q): ErrFrozenPath 예상, 하지만 반환: %v", tt.targetFile, err)
			}
		})
	}
}
