package cluster

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const (
	// learningHistoryDir 는 리포트 아티팩트가 기록되는 디렉터리(projectRoot 기준 상대)다.
	// 클러스터러가 write 하는 유일한 표면이다(REQ-OBL-013, AC-OBL-005).
	learningHistoryDir = ".moai/harness/learning-history"

	// reportFileName 은 리포트 아티팩트 파일명이다. 고정 파일명을 써서 동일 입력에
	// 대해 동일 경로/내용을 보장한다(REQ-OBL-009, AC-OBL-009).
	reportFileName = "failure-clusters.json"

	// reportSchemaVersion 은 리포트 아티팩트의 스키마 버전이다.
	reportSchemaVersion = "v1"

	// reportFilePerm / reportDirPerm 는 리포트 파일/디렉터리 권한이다.
	reportFilePerm = 0o600
	reportDirPerm  = 0o755
)

// Report 는 결정론적 리포트 아티팩트의 직렬화 루트다. 생성 시각(time.Now())을
// 절대 포함하지 않는다 — 그래야 동일 입력에 대해 바이트 동일 리포트를 보장한다
// (REQ-OBL-007/009, AC-OBL-009). first/last-seen 타임스탬프는 입력 이벤트에서
// 오는 고정값이다.
type Report struct {
	// SchemaVersion 은 리포트 스키마 버전이다.
	SchemaVersion string `json:"schema_version"`

	// ClusterCount 는 실패 클러스터 수다.
	ClusterCount int `json:"cluster_count"`

	// Clusters 는 시그니처 사전순으로 정렬된 실패 클러스터 목록이다.
	Clusters []FailureCluster `json:"clusters"`
}

// BuildReport 는 클러스터 슬라이스를 결정론적 Report 값으로 변환한다(순수 함수 —
// I/O 없음, time.Now() 없음). nil 클러스터는 빈(0-length) 슬라이스로 정규화하여
// 직렬화 시 `null` 이 아닌 `[]` 가 나오도록 한다(바이트 안정성).
func BuildReport(clusters []FailureCluster) Report {
	if clusters == nil {
		clusters = []FailureCluster{}
	}
	return Report{
		SchemaVersion: reportSchemaVersion,
		ClusterCount:  len(clusters),
		Clusters:      clusters,
	}
}

// MarshalReport 는 Report 를 결정론적 바이트열로 직렬화한다(2-space 들여쓰기 +
// 끝 개행). encoding/json 의 struct 필드 순서는 결정론적이며, 클러스터/멤버는 이미
// 안정 정렬되어 있으므로 동일 입력 → 동일 바이트다(REQ-OBL-009, AC-OBL-009).
func MarshalReport(r Report) ([]byte, error) {
	b, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("cluster: 리포트 직렬화 실패: %w", err)
	}
	b = append(b, '\n')
	return b, nil
}

// ReportPath 는 projectRoot 기준 리포트 아티팩트의 절대 경로를 결합한다
// (learning-history/failure-clusters.json).
func ReportPath(projectRoot string) string {
	return filepath.Join(projectRoot, learningHistoryDir, reportFileName)
}

// WriteReport 는 클러스터 슬라이스를 결정론적 리포트로 projectRoot 아래
// learning-history/failure-clusters.json 에 기록한다(REQ-OBL-009). 클러스터러가
// write 하는 유일한 파일이다(REQ-OBL-013). 기록된 리포트 경로를 반환한다.
//
// 동일 입력에 대해 바이트 동일 내용을 보장한다(AC-OBL-009): 생성 시각을 포함하지
// 않고, 클러스터/멤버는 안정 정렬되어 있으며, 직렬화는 결정론적이다.
func WriteReport(projectRoot string, clusters []FailureCluster) (string, error) {
	dir := filepath.Join(projectRoot, learningHistoryDir)
	if err := os.MkdirAll(dir, reportDirPerm); err != nil {
		return "", fmt.Errorf("cluster: learning-history 디렉터리 생성 실패 %s: %w", dir, err)
	}

	data, err := MarshalReport(BuildReport(clusters))
	if err != nil {
		return "", err
	}

	path := filepath.Join(dir, reportFileName)
	if err := os.WriteFile(path, data, reportFilePerm); err != nil {
		return "", fmt.Errorf("cluster: 리포트 기록 실패 %s: %w", path, err)
	}
	return path, nil
}
