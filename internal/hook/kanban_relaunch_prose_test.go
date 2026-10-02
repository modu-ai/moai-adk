package hook

import (
	"fmt"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// SPEC-FACTORY-STALE-RUN-HEAL-001 REQ-SRH-015 / AC-SRH-016: the kanban branch
// of the stale-run notice (roleValueRelaunch) is Kanban-only prose and stays
// byte-identical across this SPEC. The golden strings below are the
// pre-change bytes (tree 802a72235), written out in full so a one-character
// edit to any locale fails here instead of passing a substring check.
var kanbanRelaunchGolden = map[string]string{
	langEnglish: "stale run: this session carries legacy role value %[1]q from a binary before the leader/lane rename — " +
		"end this session and relaunch it under the current vocabulary",
	"ko": "stale run: 이 세션은 리더/레인 개칭 이전 바이너리의 레거시 역할 값 %[1]q 을(를) 담고 있습니다 — " +
		"이 세션을 끝내고 현행 어휘로 다시 띄우세요.",
	"ja": "stale run: このセッションはリーダー/レーン改名前のバイナリのレガシー役割値 %[1]q を持っています — " +
		"このセッションを終了し、現行の語彙で起動し直してください。",
	"zh": "stale run：本会话携带主导/泳道改名前二进制文件的遗留角色值 %[1]q —— " +
		"请结束本会话，并按现行词汇重新启动。",
}

func TestKanbanRelaunchProseUnchanged(t *testing.T) {
	if len(kanbanRelaunchGolden) != 4 {
		t.Fatalf("golden covers %d locales, want 4", len(kanbanRelaunchGolden))
	}
	// A kanban session: no factory run id, no factory workers stamp.
	t.Setenv(config.EnvMoaiKanbanID, "")
	t.Setenv(config.EnvMoaiFactoryWorkers, "")
	for lang, golden := range kanbanRelaunchGolden {
		t.Run(lang, func(t *testing.T) {
			if got := staleRunLocales[lang].roleValueRelaunch; got != golden {
				t.Errorf("roleValueRelaunch[%s] changed:\n got  %q\n want %q", lang, got, golden)
			}
			want := fmt.Sprintf(golden, "leader")
			if got := staleRunNotice("leader", lang); got != want {
				t.Errorf("kanban stale-run notice[%s] = %q, want the unchanged prose %q", lang, got, want)
			}
		})
	}
}
