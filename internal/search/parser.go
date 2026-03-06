package search

import (
	"bufio"
	"encoding/json"
	"os"
	"strings"
)

// Message는 JSONL에서 파싱된 단일 대화 메시지를 나타낸다.
type Message struct {
	SessionID   string
	Role        string
	Text        string
	Timestamp   string
	GitBranch   string
	ProjectPath string
}

// noisePatterns는 필터링할 XML 태그 패턴 목록이다.
// Claude Code 내부 제어 태그 및 시스템 메시지를 제외한다.
var noisePatterns = []string{
	"<local-command-caveat>",
	"<command-name>",
	"<system-reminder>",
	"<function_calls>",
}

// minTextLength는 인덱싱할 최소 텍스트 길이이다.
// 너무 짧은 텍스트는 검색 가치가 없으므로 제외한다.
const minTextLength = 20

// jsonlRecord는 JSONL 파일의 단일 레코드 구조이다.
type jsonlRecord struct {
	Type      string        `json:"type"`
	Timestamp string        `json:"timestamp"`
	SessionID string        `json:"sessionId"`
	Message   jsonlMessage  `json:"message"`
}

// jsonlMessage는 레코드 내의 메시지 객체이다.
type jsonlMessage struct {
	Role    string         `json:"role"`
	Content []jsonlContent `json:"content"`
}

// jsonlContent는 메시지 내 콘텐츠 블록이다.
type jsonlContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// ParseJSONL은 JSONL 파일을 파싱하여 검색 가능한 메시지 목록을 반환한다.
// gitBranch와 projectPath는 각 메시지에 메타데이터로 첨부된다.
//
// 필터링 규칙:
//   - type이 "user" 또는 "assistant"가 아닌 레코드 제외
//   - XML 노이즈 태그를 포함하는 메시지 제외
//   - 정제 후 텍스트가 minTextLength 미만인 메시지 제외
//   - 잘못된 JSON 라인은 건너뛴다
func ParseJSONL(path, gitBranch, projectPath string) ([]Message, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	var messages []Message
	scanner := bufio.NewScanner(f)
	// 최대 라인 버퍼 크기를 4MB로 증가 (대용량 메시지 대응)
	scanner.Buffer(make([]byte, 4*1024*1024), 4*1024*1024)

	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}

		var rec jsonlRecord
		if err := json.Unmarshal([]byte(line), &rec); err != nil {
			// 잘못된 JSON은 건너뛴다
			continue
		}

		// user/assistant 타입만 처리
		if rec.Type != "user" && rec.Type != "assistant" {
			continue
		}

		// 역할 결정: message.role을 우선, 없으면 type 사용
		role := rec.Message.Role
		if role == "" {
			role = rec.Type
		}
		if role != "user" && role != "assistant" {
			continue
		}

		// 텍스트 타입 콘텐츠만 추출
		var textParts []string
		for _, c := range rec.Message.Content {
			if c.Type == "text" && c.Text != "" {
				textParts = append(textParts, c.Text)
			}
		}
		if len(textParts) == 0 {
			continue
		}

		text := strings.Join(textParts, " ")

		// 노이즈 패턴 필터링
		if containsNoise(text) {
			continue
		}

		// 최소 길이 확인
		if len([]rune(strings.TrimSpace(text))) < minTextLength {
			continue
		}

		messages = append(messages, Message{
			SessionID:   rec.SessionID,
			Role:        role,
			Text:        text,
			Timestamp:   rec.Timestamp,
			GitBranch:   gitBranch,
			ProjectPath: projectPath,
		})
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	// nil 대신 빈 슬라이스 반환 (호출자 nil 체크 불필요)
	if messages == nil {
		messages = []Message{}
	}

	return messages, nil
}

// containsNoise는 텍스트에 노이즈 패턴이 포함되어 있으면 true를 반환한다.
func containsNoise(text string) bool {
	for _, pattern := range noisePatterns {
		if strings.Contains(text, pattern) {
			return true
		}
	}
	return false
}
