package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/modu-ai/moai-adk/internal/spec"
)

// newSpecViewCmd는 'moai spec view' 서브커맨드를 생성합니다.
func newSpecViewCmd() *cobra.Command {
	var shapeTrace bool

	cmd := &cobra.Command{
		Use:   "view <SPEC-ID>",
		Short: "View acceptance criteria in tree structure",
		Long: `Display acceptance criteria from a SPEC document in a hierarchical tree format.

Examples:
  moai spec view SPEC-SPC-001
  moai spec view SPEC-SPC-001 --shape-trace`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return cmd.Help()
			}

			specID := args[0]
			return viewAcceptanceCriteria(cmd, specID, shapeTrace)
		},
	}

	cmd.Flags().BoolVar(&shapeTrace, "shape-trace", false, "Include node depth and parent ID in output")

	return cmd
}

// viewAcceptanceCriteria는 SPEC 문서에서 Acceptance Criteria를 읽어 트리 구조로 출력합니다.
func viewAcceptanceCriteria(cmd *cobra.Command, specID string, shapeTrace bool) error {
	// 프로젝트 루트 찾기
	projectRoot, err := findProjectRootFn()
	if err != nil {
		return fmt.Errorf("failed to find project root: %w", err)
	}

	// SPEC 디렉토리 위치
	specDir := filepath.Join(projectRoot, ".moai", "specs", specID)
	specPath := filepath.Join(specDir, "spec.md")

	// spec.md 파일 존재 확인
	if _, err := os.Stat(specPath); os.IsNotExist(err) {
		return fmt.Errorf("spec.md not found for %s at %s", specID, specPath)
	}

	// 파일 읽기
	content, err := os.ReadFile(specPath)
	if err != nil {
		return fmt.Errorf("failed to read spec.md: %w", err)
	}

	// Acceptance Criteria 파싱
	criteria, parseErrors := spec.ParseAcceptanceCriteria(string(content), false)

	if len(parseErrors) > 0 {
		// 경고는 출력하지만 계속 진행
		for _, err := range parseErrors {
			switch e := err.(type) {
			case *spec.DanglingRequirementReference:
				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Warning: %v\n", e)
			case *spec.MissingRequirementMapping:
				// 리프 노드의 누락된 REQ 맵핑은 경고로만 처리
				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Warning: %v\n", e)
			default:
				// 다른 에러는 치명적
				return fmt.Errorf("parse error: %w", err)
			}
		}
	}

	// 트리 출력
	if len(criteria) == 0 {
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "No acceptance criteria found in %s\n", specID)
		return nil
	}

	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Acceptance Criteria for %s:\n\n", specID)
	printTree(cmd, criteria, "", shapeTrace, 0, "")

	return nil
}

// printTree는 Acceptance Criteria를 트리 형태로 재귀적으로 출력합니다.
func printTree(cmd *cobra.Command, criteria []spec.Acceptance, prefix string, shapeTrace bool, depth int, parentID string) {
	for i, ac := range criteria {
		// 트리 글리프 결정
		var glyph string
		var childPrefix string

		if i == len(criteria)-1 {
			glyph = "└── "
			childPrefix = prefix + "    "
		} else {
			glyph = "├── "
			childPrefix = prefix + "│   "
		}

		// 기본 정보 출력
		line := prefix + glyph + formatAcceptanceNode(ac, shapeTrace, depth, parentID)
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), line)

		// 자식 노드 재귀 출력
		if len(ac.Children) > 0 {
			printTree(cmd, ac.Children, childPrefix, shapeTrace, depth+1, ac.ID)
		}
	}
}

// formatAcceptanceNode는 단일 Acceptance Criteria 노드를 포맷팅합니다.
func formatAcceptanceNode(ac spec.Acceptance, shapeTrace bool, depth int, parentID string) string {
	var parts []string

	// ID
	parts = append(parts, ac.ID)

	// Given/When/Then
	if ac.Given != "" {
		parts = append(parts, ac.Given)
	}
	if ac.When != "" {
		parts = append(parts, ac.When)
	}
	if ac.Then != "" {
		parts = append(parts, ac.Then)
	}

	// REQ 맵핑
	if len(ac.RequirementIDs) > 0 {
		reqList := make([]string, len(ac.RequirementIDs))
		for i, reqID := range ac.RequirementIDs {
			reqList[i] = "REQ-" + reqID
		}
		parts = append(parts, fmt.Sprintf("(maps %s)", strings.Join(reqList, ", ")))
	}

	// Shape trace 정보
	if shapeTrace {
		traceInfo := fmt.Sprintf("[depth:%d", depth)
		if parentID != "" {
			traceInfo += fmt.Sprintf(", parent:%s", parentID)
		}
		traceInfo += "]"
		parts = append(parts, traceInfo)
	}

	return strings.Join(parts, ": ")
}
