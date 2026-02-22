// @MX:ANCHOR: [AUTO] main 함수는 moai CLI의 진입점입니다. 오류 발생 시 종료 코드 1을 반환합니다.
// @MX:REASON: 실행 가능한 바이너리의 유일한 진입점이며 CLI 명령 실행을 위임합니다
package main

import (
	"os"

	"github.com/modu-ai/moai-adk/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}
