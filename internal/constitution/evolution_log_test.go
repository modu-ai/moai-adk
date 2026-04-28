package constitution

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestLoadEvolutionLogs는 evolution-log 로드를 테스트한다.
func TestLoadEvolutionLogs(t *testing.T) {
	t.Run("파일 없으면 빈 목록 반환", func(t *testing.T) {
		logs, err := LoadEvolutionLogs("/nonexistent/path/evolution-log.md")
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if len(logs) != 0 {
			t.Errorf("expected empty logs, got %d", len(logs))
		}
	})

	t.Run("유효한 YAML frontmatter 파싱", func(t *testing.T) {
		tmpDir := t.TempDir()
		logPath := filepath.Join(tmpDir, "evolution-log.md")

		content := `# Evolution Log

---
id: LEARN-20260428-001
rule_id: CONST-V3R2-008
zone_before: Evolvable
zone_after: Evolvable
clause_before: "Old clause"
clause_after: "New clause"
canary_verdict: passed
contradictions: []
approved_by: human
approved_at: 2026-04-28T10:00:00Z
rolled_back: false
---
`
		if err := os.WriteFile(logPath, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		logs, err := LoadEvolutionLogs(logPath)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if len(logs) != 1 {
			t.Errorf("expected 1 log, got %d", len(logs))
		}
		if logs[0].ID != "LEARN-20260428-001" {
			t.Errorf("ID mismatch: got %s", logs[0].ID)
		}
	})

	t.Run("여러 엔트리 파싱", func(t *testing.T) {
		tmpDir := t.TempDir()
		logPath := filepath.Join(tmpDir, "evolution-log.md")

		content := `# Evolution Log

---
id: LEARN-20260428-001
rule_id: CONST-V3R2-008
zone_before: Evolvable
zone_after: Evolvable
clause_before: "Old 1"
clause_after: "New 1"
canary_verdict: passed
contradictions: []
approved_by: human
approved_at: 2026-04-28T10:00:00Z
rolled_back: false
---
---
id: LEARN-20260428-002
rule_id: CONST-V3R2-009
zone_before: Evolvable
zone_after: Evolvable
clause_before: "Old 2"
clause_after: "New 2"
canary_verdict: passed
contradictions: []
approved_by: human
approved_at: 2026-04-28T11:00:00Z
rolled_back: false
---
`
		if err := os.WriteFile(logPath, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		logs, err := LoadEvolutionLogs(logPath)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if len(logs) != 2 {
			t.Errorf("expected 2 logs, got %d", len(logs))
		}
	})
}

// TestAppendEvolutionLog는 로그 추가를 테스트한다.
func TestAppendEvolutionLog(t *testing.T) {
	t.Run("새 파일에 엔트리 추가", func(t *testing.T) {
		tmpDir := t.TempDir()
		logPath := filepath.Join(tmpDir, "evolution-log.md")

		log := &AmendmentLog{
			ID:            "LEARN-20260428-001",
			RuleID:        "CONST-V3R2-008",
			ZoneBefore:    ZoneEvolvable,
			ZoneAfter:     ZoneEvolvable,
			ClauseBefore:  "Old clause",
			ClauseAfter:   "New clause",
			CanaryVerdict: "passed",
			Contradictions: []string{},
			ApprovedBy:    "human",
			ApprovedAt:    time.Date(2026, 4, 28, 10, 0, 0, 0, time.UTC),
			RolledBack:    false,
		}

		if err := AppendEvolutionLog(logPath, log); err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		// 파일 존재 확인
		if _, err := os.Stat(logPath); os.IsNotExist(err) {
			t.Error("file was not created")
		}

		// 로드 확인
		logs, err := LoadEvolutionLogs(logPath)
		if err != nil {
			t.Errorf("failed to load logs: %v", err)
		}
		if len(logs) != 1 {
			t.Errorf("expected 1 log, got %d", len(logs))
		}
		if logs[0].ID != log.ID {
			t.Errorf("ID mismatch: got %s, want %s", logs[0].ID, log.ID)
		}
	})

	t.Run("기존 파일에 엔트리 추가", func(t *testing.T) {
		tmpDir := t.TempDir()
		logPath := filepath.Join(tmpDir, "evolution-log.md")

		// 첫 번째 엔트리
		log1 := &AmendmentLog{
			ID:           "LEARN-20260428-001",
			RuleID:       "CONST-V3R2-008",
			ZoneBefore:   ZoneEvolvable,
			ZoneAfter:    ZoneEvolvable,
			ClauseBefore: "Old 1",
			ClauseAfter:  "New 1",
			CanaryVerdict: "passed",
			ApprovedBy:   "human",
			ApprovedAt:   time.Now(),
			RolledBack:   false,
		}
		if err := AppendEvolutionLog(logPath, log1); err != nil {
			t.Fatal(err)
		}

		// 두 번째 엔트리
		log2 := &AmendmentLog{
			ID:           "LEARN-20260428-002",
			RuleID:       "CONST-V3R2-009",
			ZoneBefore:   ZoneEvolvable,
			ZoneAfter:    ZoneEvolvable,
			ClauseBefore: "Old 2",
			ClauseAfter:  "New 2",
			CanaryVerdict: "passed",
			ApprovedBy:   "human",
			ApprovedAt:   time.Now(),
			RolledBack:   false,
		}
		if err := AppendEvolutionLog(logPath, log2); err != nil {
			t.Fatal(err)
		}

		// 로드 확인
		logs, err := LoadEvolutionLogs(logPath)
		if err != nil {
			t.Errorf("failed to load logs: %v", err)
		}
		if len(logs) != 2 {
			t.Errorf("expected 2 logs, got %d", len(logs))
		}
	})

	t.Run("�증 실패 시 에러 반환", func(t *testing.T) {
		tmpDir := t.TempDir()
		logPath := filepath.Join(tmpDir, "evolution-log.md")

		// ID 없는 로그
		log := &AmendmentLog{
			RuleID:       "CONST-V3R2-008",
			ZoneBefore:   ZoneEvolvable,
			ZoneAfter:    ZoneEvolvable,
			ClauseBefore: "Old",
			ClauseAfter:  "New",
			ApprovedBy:   "human",
			ApprovedAt:   time.Now(),
		}

		if err := AppendEvolutionLog(logPath, log); err == nil {
			t.Error("expected error for invalid log, got nil")
		}
	})
}

// TestMarkRolledBack은 rollback 마킹을 테스트한다.
func TestMarkRolledBack(t *testing.T) {
	t.Run("활성 로그를 rolled_back으로 변경", func(t *testing.T) {
		tmpDir := t.TempDir()
		logPath := filepath.Join(tmpDir, "evolution-log.md")

		// 초기 로그 생성
		log := &AmendmentLog{
			ID:           "LEARN-20260428-001",
			RuleID:       "CONST-V3R2-008",
			ZoneBefore:   ZoneEvolvable,
			ZoneAfter:    ZoneEvolvable,
			ClauseBefore: "Old",
			ClauseAfter:  "New",
			CanaryVerdict: "passed",
			ApprovedBy:   "human",
			ApprovedAt:   time.Now(),
			RolledBack:   false,
		}
		if err := AppendEvolutionLog(logPath, log); err != nil {
			t.Fatal(err)
		}

		// Rollback 마킹
		reason := "Score drop 0.15 > threshold 0.10 detected in next SPEC evaluation"
		if err := MarkRolledBack(logPath, "CONST-V3R2-008", reason); err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		// 확인
		logs, err := LoadEvolutionLogs(logPath)
		if err != nil {
			t.Errorf("failed to load logs: %v", err)
		}
		if len(logs) != 1 {
			t.Errorf("expected 1 log, got %d", len(logs))
		}
		if !logs[0].RolledBack {
			t.Error("expected RolledBack to be true")
		}
		if logs[0].RollbackReason != reason {
			t.Errorf("rollback reason mismatch: got %s, want %s", logs[0].RollbackReason, reason)
		}
		if logs[0].RollbackAt == nil {
			t.Error("expected RollbackAt to be set")
		}
	})

	t.Run("존재하지 않는 rule은 에러 반환", func(t *testing.T) {
		tmpDir := t.TempDir()
		logPath := filepath.Join(tmpDir, "evolution-log.md")

		// 빈 파일 생성
		if err := os.WriteFile(logPath, []byte("# Evolution Log\n"), 0644); err != nil {
			t.Fatal(err)
		}

		err := MarkRolledBack(logPath, "CONST-V3R2-999", "reason")
		if err == nil {
			t.Error("expected error for non-existent rule, got nil")
		}
	})
}
