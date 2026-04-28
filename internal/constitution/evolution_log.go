package constitution

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// evolutionLogPath는 기본 evolution-log.md 경로를 반환한다.
// SPEC-V3R2-CON-003에서 사용 예정.
var _ = filepath.Join // referenced by LoadEvolutionLogs path construction

// LoadEvolutionLogs는 evolution-log.md 파일에서 로그 목록을 로드한다.
// 파일이 없으면 빈 목록과 nil 에러를 반환한다.
func LoadEvolutionLogs(path string) ([]AmendmentLog, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []AmendmentLog{}, nil
		}
		return nil, fmt.Errorf("evolution-log.md 읽기 오류: %w", err)
	}

	// 마크다운에서 YAML frontmatter 추출
	entries := strings.Split(string(data), "---")
	var logs []AmendmentLog

	for i := 1; i < len(entries); i += 2 {
		if i+1 >= len(entries) {
			break
		}
		yamlBlock := entries[i]

		var log AmendmentLog
		if err := yaml.Unmarshal([]byte(yamlBlock), &log); err != nil {
			// 파싱 오류는 무시하고 다음 엔트리로 진행
			continue
		}

		if log.ID != "" {
			logs = append(logs, log)
		}
	}

	return logs, nil
}

// AppendEvolutionLog는 evolution-log.md 파일에 새 로그를 추가한다.
// Append-only 전략: 기존 내용을 보존하고 파일 끝에 추가한다.
func AppendEvolutionLog(path string, log *AmendmentLog) error {
	// 검증
	if err := log.Validate(); err != nil {
		return fmt.Errorf("amendment log 검증 오류: %w", err)
	}

	// YAML frontmatter 생성
	yamlData, err := yaml.Marshal(log)
	if err != nil {
		return fmt.Errorf("YAML 마샬링 오류: %w", err)
	}

	// 파일 열기 (쓰기 전용, 생성, 추가)
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("파일 열기 오류: %w", err)
	}
	defer func() { _ = f.Close() }()

	// 엔트리 작성: --- 구분자 + YAML + ---
	entry := fmt.Sprintf("---\n%s---\n", string(yamlData))
	if _, err := f.WriteString(entry); err != nil {
		return fmt.Errorf("파일 쓰기 오류: %w", err)
	}

	return nil
}

// MarkRolledBack은 evolution-log.md에서 해당 rule의 로그를 찾아 rolled_back을 true로 설정한다.
// SPEC-V3R2-CON-002 REQ-CON-002-008 구현.
//
// Rollback 트리거:
// - Amendment 이후 다음 SPEC 평가에서 score가 0.10 이상 하락
// - evaluator-active가 regression 탐지
func MarkRolledBack(path, ruleID string, reason string) error {
	logs, err := LoadEvolutionLogs(path)
	if err != nil {
		return err
	}

	// 해당 rule의 가장 최신 로그 찾기
	var found bool
	for i := len(logs) - 1; i >= 0; i-- {
		if logs[i].RuleID == ruleID && !logs[i].RolledBack {
			now := time.Now()
			logs[i].RolledBack = true
			logs[i].RollbackReason = reason
			logs[i].RollbackAt = &now
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("rule %s의 active 로그를 찾을 수 없음", ruleID)
	}

	// 전체 파일 다시 쓰기 (append-only이나 rollback은 기존 엔트리 수정)
	return rewriteEvolutionLog(path, logs)
}

// rewriteEvolutionLog는 로그 목록으로 파일을 다시 쓴다.
func rewriteEvolutionLog(path string, logs []AmendmentLog) error {
	// 디렉토리 확인
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("디렉토리 생성 오류: %w", err)
	}

	// 임시 파일에 쓰기
	tmpPath := path + ".tmp"
	f, err := os.Create(tmpPath)
	if err != nil {
		return fmt.Errorf("임시 파일 생성 오류: %w", err)
	}
	defer func() { _ = f.Close() }()

	// 헤더 작성
	if _, err := f.WriteString("# Evolution Log\n\n"); err != nil {
		return err
	}

	// 각 로그 작성
	for _, log := range logs {
		yamlData, err := yaml.Marshal(log)
		if err != nil {
			return fmt.Errorf("YAML 마샬링 오류: %w", err)
		}

		entry := fmt.Sprintf("---\n%s---\n", string(yamlData))
		if _, err := f.WriteString(entry); err != nil {
			return err
		}
	}

	// 원자적 교체 (rename)
	if err := f.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("파일 교체 오류: %w", err)
	}

	return nil
}

// GenerateLogID는 새 로그 ID를 생성한다.
// LEARN-YYYYMMDD-NNN 형식.
func GenerateLogID(now time.Time, lastLogs []AmendmentLog) string {
	dateStr := now.Format("20060102")

	// 해당 날짜의 마지막 시퀀스 번호 찾기
	maxSeq := 0
	for _, log := range lastLogs {
		prefix := "LEARN-" + dateStr + "-"
		if strings.HasPrefix(log.ID, prefix) {
			seqStr := strings.TrimPrefix(log.ID, prefix)
			var seq int
			if _, err := fmt.Sscanf(seqStr, "%d", &seq); err == nil {
				if seq > maxSeq {
					maxSeq = seq
				}
			}
		}
	}

	// 다음 시퀀스 번호
	return fmt.Sprintf("LEARN-%s-%03d", dateStr, maxSeq+1)
}
