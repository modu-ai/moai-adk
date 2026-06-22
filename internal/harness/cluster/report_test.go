package cluster

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/harness"
)

// ── AC-OBL-009: 리포트 아티팩트 결정론적 emit (REQ-OBL-009) ──

// TestClusterReport: 비어 있지 않은 클러스터 집합에 대해 리포트를 두 번 emit 하면
// learning-history/ 아래에 바이트 동일 내용이 기록된다. time.Now() 누수 없음.
func TestClusterReport(t *testing.T) {
	events := []harness.Event{
		mkEvent(verdictRolledBack, "regression-blocked", "PROPOSAL-A", []string{"coverage"}, ts(1)),
		mkEvent(verdictRolledBack, "regression-blocked", "PROPOSAL-B", []string{"coverage"}, ts(2)),
		mkEvent(verdictRolledBack, "regression-blocked", "PROPOSAL-C", []string{"lint"}, ts(3)),
	}
	clusters := ClusterEvents(events)

	root1 := t.TempDir()
	root2 := t.TempDir()

	path1, err := WriteReport(root1, clusters)
	if err != nil {
		t.Fatalf("WriteReport(1) err = %v", err)
	}
	path2, err := WriteReport(root2, clusters)
	if err != nil {
		t.Fatalf("WriteReport(2) err = %v", err)
	}

	// 리포트 경로가 learning-history/ 아래인지 검증(AC-OBL-009).
	if !strings.Contains(path1, filepath.Join(".moai", "harness", "learning-history")) {
		t.Errorf("리포트 경로가 learning-history/ 아래가 아님: %q", path1)
	}
	if filepath.Base(path1) != reportFileName {
		t.Errorf("리포트 파일명 = %q, want %q", filepath.Base(path1), reportFileName)
	}

	b1, err := os.ReadFile(path1)
	if err != nil {
		t.Fatalf("read report1: %v", err)
	}
	b2, err := os.ReadFile(path2)
	if err != nil {
		t.Fatalf("read report2: %v", err)
	}

	if string(b1) != string(b2) {
		t.Errorf("두 리포트가 바이트 동일하지 않음(결정론 위반)\n--- report1 ---\n%s\n--- report2 ---\n%s", b1, b2)
	}

	// 리포트에 생성 시각/날짜 류 키가 들어가면 안 된다(time.Now() 누수 방지).
	for _, forbidden := range []string{"generated_at", "created_at", "now"} {
		if strings.Contains(string(b1), forbidden) {
			t.Errorf("리포트에 비결정론적 키 %q 가 포함됨: %s", forbidden, b1)
		}
	}
}

// TestWriteReportIsOnlyFileWritten: 클러스터러가 learning-history/ 아래에 write 하는
// 파일은 리포트 하나뿐이다(AC-OBL-005 의 "유일한 write 파일" 불변식). 빈 디렉터리에
// 리포트를 쓴 뒤, learning-history/ 직속에 reportFileName 외 파일이 없어야 한다.
func TestWriteReportIsOnlyFileWritten(t *testing.T) {
	root := t.TempDir()
	clusters := ClusterEvents([]harness.Event{
		mkEvent(verdictRolledBack, "regression-blocked", "PROPOSAL-A", []string{"coverage"}, ts(1)),
	})
	if _, err := WriteReport(root, clusters); err != nil {
		t.Fatalf("WriteReport err = %v", err)
	}

	lhDir := filepath.Join(root, learningHistoryDir)
	entries, err := os.ReadDir(lhDir)
	if err != nil {
		t.Fatalf("read learning-history dir: %v", err)
	}
	if len(entries) != 1 {
		var names []string
		for _, e := range entries {
			names = append(names, e.Name())
		}
		t.Fatalf("learning-history 직속 항목 수 = %d (%v), want 1 (리포트만)", len(entries), names)
	}
	if entries[0].Name() != reportFileName {
		t.Errorf("유일 항목 = %q, want %q", entries[0].Name(), reportFileName)
	}
}

// TestWriteReportEmptyClusters: 빈 클러스터 집합도 유효한 리포트(cluster_count: 0,
// clusters: [])를 emit 한다 — null 이 아닌 빈 배열.
func TestWriteReportEmptyClusters(t *testing.T) {
	root := t.TempDir()
	path, err := WriteReport(root, nil)
	if err != nil {
		t.Fatalf("WriteReport(nil) err = %v", err)
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read report: %v", err)
	}
	content := string(b)
	if !strings.Contains(content, `"cluster_count": 0`) {
		t.Errorf("빈 리포트에 cluster_count: 0 가 없음: %s", content)
	}
	if !strings.Contains(content, `"clusters": []`) {
		t.Errorf("빈 리포트의 clusters 가 [] 가 아님(null 누수): %s", content)
	}
}

// TestBuildReportNilNormalization: BuildReport 는 nil 클러스터를 빈 슬라이스로
// 정규화하여 직렬화 시 null 이 아닌 [] 를 보장한다.
func TestBuildReportNilNormalization(t *testing.T) {
	r := BuildReport(nil)
	if r.Clusters == nil {
		t.Error("BuildReport(nil) 은 nil 이 아닌 빈 슬라이스를 줘야 함")
	}
	if r.ClusterCount != 0 {
		t.Errorf("ClusterCount = %d, want 0", r.ClusterCount)
	}
	if r.SchemaVersion != reportSchemaVersion {
		t.Errorf("SchemaVersion = %q, want %q", r.SchemaVersion, reportSchemaVersion)
	}
}

// TestMarshalReportTrailingNewline: 직렬화 결과는 끝 개행을 포함한다(POSIX 텍스트 규약).
func TestMarshalReportTrailingNewline(t *testing.T) {
	b, err := MarshalReport(BuildReport(nil))
	if err != nil {
		t.Fatalf("MarshalReport err = %v", err)
	}
	if len(b) == 0 || b[len(b)-1] != '\n' {
		t.Errorf("리포트 직렬화는 끝 개행으로 끝나야 함")
	}
}

// TestReportPath: ReportPath 가 projectRoot 기준 절대 경로를 올바르게 결합한다.
func TestReportPath(t *testing.T) {
	got := ReportPath("/proj")
	want := filepath.Join("/proj", ".moai", "harness", "learning-history", reportFileName)
	if got != want {
		t.Errorf("ReportPath() = %q, want %q", got, want)
	}
}
