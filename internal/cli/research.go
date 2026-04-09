package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

// newResearchCmd 는 research 커맨드 트리의 루트를 생성한다.
func newResearchCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "research",
		Short:   "Self-Research System 관리",
		Long:    "moai-adk 컴포넌트를 자율적으로 실험하고 개선하는 Self-Research System",
		GroupID: "tools",
	}
	cmd.AddCommand(newResearchStatusCmd())
	cmd.AddCommand(newResearchBaselineCmd())
	cmd.AddCommand(newResearchListCmd())
	return cmd
}

// newResearchStatusCmd 는 연구 대시보드를 표시하는 서브커맨드를 생성한다.
func newResearchStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "연구 대시보드 표시",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("get working directory: %w", err)
			}
			return runResearchStatus(cmd.OutOrStdout(), cwd)
		},
	}
}

// runResearchStatus 는 projectDir 기준으로 연구 상태를 출력한다.
func runResearchStatus(w io.Writer, projectDir string) error {
	researchDir := filepath.Join(projectDir, ".moai", "research")
	if _, err := os.Stat(researchDir); os.IsNotExist(err) {
		_, _ = fmt.Fprintln(w, "No research data found. Run 'moai research baseline <target>' first.")
		return nil
	}

	// 기본 대시보드 출력
	pairs := []kvPair{
		{"Directory", filepath.Join(".moai", "research")},
	}

	// 실험 결과 디렉터리 수 확인
	experimentsDir := filepath.Join(researchDir, "experiments")
	expCount := countDirs(experimentsDir)
	pairs = append(pairs, kvPair{"Experiments", fmt.Sprintf("%d found", expCount)})

	// eval suite 수 확인
	evalsDir := filepath.Join(researchDir, "evals")
	evalCount := countEvalFiles(evalsDir)
	pairs = append(pairs, kvPair{"Eval Suites", fmt.Sprintf("%d registered", evalCount)})

	_, _ = fmt.Fprintln(w, renderCard("Research Status", renderKeyValueLines(pairs)))
	return nil
}

// newResearchBaselineCmd 는 baseline 측정 서브커맨드를 생성한다 (미구현).
func newResearchBaselineCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "baseline [target]",
		Short: "Baseline 측정 (coming soon)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Baseline measurement is not yet implemented. Coming in a future release.")
			return nil
		},
	}
}

// newResearchListCmd 는 eval suite 목록을 출력하는 서브커맨드를 생성한다.
func newResearchListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "등록된 eval suite 목록",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("get working directory: %w", err)
			}
			return runResearchList(cmd.OutOrStdout(), cwd)
		},
	}
}

// runResearchList 는 projectDir 기준으로 .moai/research/evals/ 하위의 eval 파일을 나열한다.
func runResearchList(w io.Writer, projectDir string) error {
	evalsDir := filepath.Join(projectDir, ".moai", "research", "evals")
	files, err := findEvalFiles(evalsDir)
	if err != nil || len(files) == 0 {
		_, _ = fmt.Fprintln(w, "No eval suites found. Create one at .moai/research/evals/")
		return nil
	}

	_, _ = fmt.Fprintf(w, "Eval suites (%d):\n", len(files))
	for _, f := range files {
		// 프로젝트 루트 기준 상대 경로로 출력
		rel, relErr := filepath.Rel(projectDir, f)
		if relErr != nil {
			rel = f
		}
		_, _ = fmt.Fprintf(w, "  %s\n", rel)
	}
	return nil
}

// findEvalFiles 는 dir 하위에서 *.eval.yaml 파일을 재귀 탐색한다.
func findEvalFiles(dir string) ([]string, error) {
	var results []string
	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(d.Name(), ".eval.yaml") {
			results = append(results, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return results, nil
}

// countEvalFiles 는 dir 하위의 *.eval.yaml 파일 수를 반환한다.
func countEvalFiles(dir string) int {
	files, err := findEvalFiles(dir)
	if err != nil {
		return 0
	}
	return len(files)
}

func init() {
	rootCmd.AddCommand(newResearchCmd())
}
