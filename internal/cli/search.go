package cli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/defs"
	"github.com/modu-ai/moai-adk/internal/search"
)

// searchCmd는 세션 히스토리 검색 서브커맨드이다.
var searchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search indexed session history",
	Long: `Search indexed Claude Code session history using SQLite FTS5 full-text search.

Supports Korean/CJK text search via trigram tokenizer.
Sessions are automatically indexed via SessionEnd hook.

Examples:
  moai search "JWT 인증"           # Full-text search
  moai search "인덱스" --branch main  # Filter by branch
  moai search "플로우" --role user    # Filter by role
  moai search --index-session abc123  # Manually index a session`,
	GroupID: "tools",
	RunE:    runSearch,
}

func init() {
	rootCmd.AddCommand(searchCmd)

	// 검색 옵션 플래그
	searchCmd.Flags().StringP("branch", "b", "", "git 브랜치 필터")
	searchCmd.Flags().String("since", "", "시작 날짜 필터 (RFC3339, 예: 2026-01-01T00:00:00Z)")
	searchCmd.Flags().String("until", "", "종료 날짜 필터 (RFC3339, 예: 2026-12-31T23:59:59Z)")
	searchCmd.Flags().String("role", "", "역할 필터: user|assistant")
	searchCmd.Flags().IntP("limit", "n", 20, "최대 결과 수")

	// 인덱싱 옵션 플래그
	searchCmd.Flags().String("index-session", "", "특정 세션 ID를 수동으로 인덱싱")
	searchCmd.Flags().String("project-path", "", "인덱싱 시 사용할 프로젝트 경로")
	searchCmd.Flags().String("git-branch", "", "인덱싱 시 사용할 git 브랜치")
}

// runSearch는 search 커맨드의 실행 핸들러이다.
// --index-session 플래그가 있으면 세션을 인덱싱하고 종료한다.
// 없으면 쿼리 인수를 받아 검색 결과를 출력한다.
func runSearch(cmd *cobra.Command, args []string) error {
	sessionID, _ := cmd.Flags().GetString("index-session")

	// 인덱싱 모드
	if sessionID != "" {
		return runIndexSession(cmd, sessionID)
	}

	// 검색 모드: 쿼리 인수 필수
	if len(args) == 0 {
		return errors.New("검색 쿼리가 필요합니다. 예: moai search \"JWT 인증\"")
	}

	return runSearchQuery(cmd, args[0])
}

// runIndexSession은 지정된 세션 ID를 JSONL 파일에서 찾아 인덱싱한다.
func runIndexSession(cmd *cobra.Command, sessionID string) error {
	projectPath, _ := cmd.Flags().GetString("project-path")
	gitBranch, _ := cmd.Flags().GetString("git-branch")

	dbPath, err := searchDBPath()
	if err != nil {
		return err
	}

	db, err := search.OpenDB(dbPath)
	if err != nil {
		return fmt.Errorf("DB 오픈 실패: %w", err)
	}
	defer func() { _ = db.Close() }()

	if err := search.CreateTables(db); err != nil {
		return fmt.Errorf("DB 테이블 초기화 실패: %w", err)
	}

	// ~/.claude/projects/ 하위에서 JSONL 파일 검색
	filePath, err := findSessionJSONL(sessionID)
	if err != nil {
		return fmt.Errorf("세션 JSONL 파일을 찾을 수 없음 (session=%s): %w", sessionID, err)
	}

	if err := search.IndexSession(db, sessionID, filePath, gitBranch, projectPath); err != nil {
		return fmt.Errorf("세션 인덱싱 실패: %w", err)
	}

	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "세션 인덱싱 완료: %s\n", sessionID)
	return nil
}

// runSearchQuery는 쿼리를 실행하고 결과를 출력한다.
func runSearchQuery(cmd *cobra.Command, query string) error {
	branch, _ := cmd.Flags().GetString("branch")
	since, _ := cmd.Flags().GetString("since")
	until, _ := cmd.Flags().GetString("until")
	role, _ := cmd.Flags().GetString("role")
	limit, _ := cmd.Flags().GetInt("limit")

	dbPath, err := searchDBPath()
	if err != nil {
		return err
	}

	db, err := search.OpenDB(dbPath)
	if err != nil {
		return fmt.Errorf("DB 오픈 실패: %w", err)
	}
	defer func() { _ = db.Close() }()

	if err := search.CreateTables(db); err != nil {
		return fmt.Errorf("DB 테이블 초기화 실패: %w", err)
	}

	opts := search.SearchOptions{
		Query:  query,
		Branch: branch,
		Since:  since,
		Until:  until,
		Role:   role,
		Limit:  limit,
	}

	results, err := search.Search(db, opts)
	if err != nil {
		return fmt.Errorf("검색 실패: %w", err)
	}

	printSearchResults(cmd, query, results)
	return nil
}

// printSearchResults는 검색 결과를 터미널에 출력한다.
func printSearchResults(cmd *cobra.Command, query string, results []search.SearchResult) {
	out := cmd.OutOrStdout()

	if len(results) == 0 {
		_, _ = fmt.Fprintf(out, "검색 결과 없음: %q\n", query)
		return
	}

	// 헤더 출력
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.AdaptiveColor{
		Light: "#C45A3C", Dark: "#DA7756",
	})
	_, _ = fmt.Fprintln(out, headerStyle.Render(fmt.Sprintf("검색 결과: %q (%d건)", query, len(results))))
	_, _ = fmt.Fprintln(out)

	// 각 결과를 카드 형태로 출력
	roleStyle := lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{
		Light: "#059669", Dark: "#10B981",
	})
	mutedStyle := lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{
		Light: "#9CA3AF", Dark: "#6B7280",
	})
	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.AdaptiveColor{Light: "#D1D5DB", Dark: "#4B5563"}).
		Padding(0, 2)

	for i, r := range results {
		// 카드 내용 구성
		meta := fmt.Sprintf("%s  %s  %s",
			roleStyle.Render(r.Role),
			mutedStyle.Render(r.GitBranch),
			mutedStyle.Render(r.Timestamp),
		)
		excerpt := r.Excerpt
		if excerpt == "" {
			excerpt = "(내용 없음)"
		}
		// 세션 ID 단축 표시
		shortSession := r.SessionID
		if len(shortSession) > 12 {
			shortSession = shortSession[:12] + "..."
		}

		content := meta + "\n" + excerpt + "\n" + mutedStyle.Render(fmt.Sprintf("session: %s", shortSession))
		_, _ = fmt.Fprintln(out, borderStyle.Render(content))

		// 마지막 결과 이후 구분선 없음
		if i < len(results)-1 {
			_, _ = fmt.Fprintln(out)
		}
	}
}

// searchDBPath는 검색 DB 파일 경로를 반환한다.
// 경로: ~/.moai/search/sessions.db
func searchDBPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("홈 디렉터리를 찾을 수 없음: %w", err)
	}
	return filepath.Join(homeDir, defs.MoAIDir, defs.SearchSubdir, defs.SearchDB), nil
}

// findSessionJSONL은 ~/.claude/projects/ 하위 디렉터리에서
// {sessionId}.jsonl 파일을 찾아 경로를 반환한다.
func findSessionJSONL(sessionID string) (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("홈 디렉터리를 찾을 수 없음: %w", err)
	}

	projectsDir := filepath.Join(homeDir, ".claude", "projects")
	targetFile := sessionID + ".jsonl"

	// 하위 디렉터리를 순회하여 JSONL 파일 검색
	var found string
	err = filepath.WalkDir(projectsDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			// 접근 불가 항목은 건너뜀
			return nil
		}
		if !d.IsDir() && d.Name() == targetFile {
			found = path
			return filepath.SkipAll
		}
		return nil
	})
	if err != nil {
		return "", err
	}

	if found == "" {
		return "", fmt.Errorf("%s.jsonl 파일을 %s에서 찾을 수 없음",
			sessionID, projectsDir)
	}

	return found, nil
}
