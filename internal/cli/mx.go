package cli

import (
	"github.com/spf13/cobra"
)

// newMxCmd는 'moai mx' 부모 명령을 생성합니다.
// @MX TAG 관련 서브커맨드(query 등)를 포함합니다.
func newMxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "mx",
		Short:   "@MX TAG 관리 도구",
		Long:    `@MX TAG 사이드카 인덱스를 관리하고 조회하는 도구입니다.`,
		GroupID: "tools",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	// SPEC-V3R2-SPC-004: query 서브커맨드 등록
	cmd.AddCommand(newMxQueryCmd())

	return cmd
}

func init() {
	rootCmd.AddCommand(newMxCmd())
}
