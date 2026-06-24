// Package cli — `moai harness clusters` read 표면 (SPEC-DIVECC-OBSERVABILITY-LOOP-001).
//
// 이 파일은 실패 시그니처 클러스터링 엔진(internal/harness/cluster)의 CLI read 표면을
// 제공한다. runHarnessStatus 와 동일한 형태(resolveProjectRoot → JSONL 읽기 → 집계 →
// 출력)를 그대로 미러링하며, 새로운/분기된 경로 해석 경로를 절대 도입하지 않는다
// (REQ-OBL-011). 클러스터러는 read-only이므로 이 명령도 어떤 파일도 write 하지 않는다.
//
// [HARD] Subagent boundary (C-HRA-008 / REQ-OBL-015): AskUserQuestion / mcp__askuser
// 호출 금지 — 이 CLI 는 subagent 컨텍스트에서 동작하며 사용자 상호작용은 orchestrator 가
// 소유한다. 상호작용 대신 stdout(기계 판독 JSON --json) / 텍스트 + 종료 코드만 쓴다.
package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/harness/cluster"
)

// newHarnessClustersCmd 은 `moai harness clusters` subcommand 팩토리다.
// LIVE harness 트리(newHarnessRouterCmd, harness_route.go)에 status 와 나란히
// 등록된다(AC-OBL-004). 비활성 deprecation-marker 트리(newHarnessCmd)에는 등록하지 않는다.
func newHarnessClustersCmd() *cobra.Command {
	var asJSON bool
	cmd := &cobra.Command{
		Use:   "clusters",
		Short: "Show deterministic failure-signature clusters from apply outcomes (read-only)",
		Long: `Read apply_outcome events from .moai/harness/usage-log.jsonl and group
failure/rolled-back events into deterministic failure-signature clusters.

The signature key is derived ONLY from fields present on the apply_outcome
event: the sorted outcome_regressed dimension set + outcome_verdict +
outcome_decision. kept outcomes are excluded. No machine learning, no
randomness — the output is byte-identical across repeated runs.

This command is strictly read-only with respect to the proposal/apply path.
It reads usage-log.jsonl and prints clusters; it never writes back into the
proposal/apply path and never alters any Apply decision.

Use --json for machine-readable output (stdout). When there are zero failure
clusters the command reports an empty result and exits 0 (not an error).

[HARD] This command does not call AskUserQuestion — the CLI runs in subagent
context and the orchestrator owns user interaction.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runHarnessClusters(cmd, args, asJSON)
		},
	}
	cmd.Flags().BoolVar(&asJSON, "json", false, "emit machine-readable JSON on stdout")
	return cmd
}

// runHarnessClusters 는 clusters verb 를 실행한다.
//
// runHarnessStatus 와 동일하게 resolveProjectRoot(cmd) 로 프로젝트 루트를 해석하고
// (REQ-OBL-011 — 새로운 경로 해석 경로 도입 금지), cluster.DefaultLogPath 로 입력
// usage-log.jsonl 경로를 구성한 뒤, cluster.Cluster 로 결정론적 클러스터를 계산한다.
//
// 클러스터가 0개여도 빈 결과를 보고하고 exit 0 으로 종료한다(REQ-OBL-012, 에러 아님).
func runHarnessClusters(cmd *cobra.Command, _ []string, asJSON bool) error {
	// runHarnessStatus 와 동일한 공유 헬퍼 재사용(REQ-OBL-011): --project-root 플래그
	// (상속 포함) → 빈 값이면 os.Getwd() 폴백. 분기된 경로 해석 함수를 새로 만들지 않는다.
	root, err := resolveProjectRoot(cmd)
	if err != nil {
		return err
	}

	logPath := cluster.DefaultLogPath(root)
	clusters, err := cluster.Cluster(logPath)
	if err != nil {
		return fmt.Errorf("harness clusters: %w", err)
	}

	out := cmd.OutOrStdout()

	if asJSON {
		// 기계 판독 출력은 stdout 으로(--json). 결정론적 리포트 구조를 직렬화한다.
		report := cluster.BuildReport(clusters)
		data, mErr := cluster.MarshalReport(report)
		if mErr != nil {
			return mErr
		}
		_, _ = out.Write(data)
		return nil
	}

	// 사람이 읽는 텍스트 출력(기본).
	if len(clusters) == 0 {
		// 빈 결과 보고(REQ-OBL-012) — 에러가 아니라 exit 0.
		_, _ = fmt.Fprintln(out, "No failure-signature clusters found.")
		return nil
	}

	_, _ = fmt.Fprintf(out, "=== Failure-Signature Clusters (%d) ===\n\n", len(clusters))
	for i, c := range clusters {
		_, _ = fmt.Fprintf(out, "[%d] signature: %s\n", i+1, c.Signature)
		_, _ = fmt.Fprintf(out, "    count             : %d\n", c.Count)
		_, _ = fmt.Fprintf(out, "    verdict           : %s\n", c.Verdict)
		_, _ = fmt.Fprintf(out, "    decision          : %s\n", c.Decision)
		_, _ = fmt.Fprintf(out, "    regressed dims    : %v\n", c.RepresentativeDimensions)
		_, _ = fmt.Fprintf(out, "    first seen        : %s\n", c.FirstSeen.Format("2006-01-02T15:04:05Z07:00"))
		_, _ = fmt.Fprintf(out, "    last seen         : %s\n", c.LastSeen.Format("2006-01-02T15:04:05Z07:00"))
		_, _ = fmt.Fprintln(out)
	}

	return nil
}
