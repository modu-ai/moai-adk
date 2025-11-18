#!/usr/bin/env python3
"""
SubagentStop Hook: Track agent lifecycle completion

Claude Code v2.0.42 신규 기능:
- agent_id: 에이전트 식별자
- agent_transcript_path: 에이전트 대화 기록 파일 경로
- execution_time_ms: 에이전트 실행 시간 (밀리초)

역할:
1. 에이전트 실행 완료 기록
2. 실행 시간 측정 및 저장
3. 성능 데이터 수집
4. 워크플로우 체인 분석
"""

import json
import sys
from datetime import datetime
from pathlib import Path


def track_agent_lifecycle(
    agent_id: str,
    agent_name: str,
    transcript_path: str,
    execution_time_ms: int,
    success: bool,
) -> dict:
    """
    에이전트 실행 라이프사이클 추적

    v2.0.42: agent_transcript_path를 활용한 상세 추적
    """

    # 에이전트 메타데이터 업데이트
    metadata_dir = Path(".moai/logs/agent-transcripts")
    metadata_dir.mkdir(parents=True, exist_ok=True)

    metadata_file = metadata_dir / f"agent-{agent_id}.json"

    # 기존 메타데이터 읽기
    metadata = {}
    if metadata_file.exists():
        try:
            metadata = json.loads(metadata_file.read_text())
        except json.JSONDecodeError:
            pass

    # 실행 완료 정보 추가
    metadata.update({
        "agent_id": agent_id,
        "agent_name": agent_name,
        "transcript_path": transcript_path,  # v2.0.42
        "execution_time_ms": execution_time_ms,
        "execution_time_seconds": execution_time_ms / 1000,
        "success": success,
        "completed_at": datetime.now().isoformat(),
        "status": "completed" if success else "failed",
    })

    # 메타데이터 저장
    metadata_file.write_text(json.dumps(metadata, indent=2, ensure_ascii=False))

    # 에이전트별 성능 통계 기록
    stats_file = Path(".moai/logs/agent-performance.jsonl")
    stats_file.parent.mkdir(parents=True, exist_ok=True)

    performance_record = {
        "timestamp": datetime.now().isoformat(),
        "agent_id": agent_id,
        "agent_name": agent_name,
        "execution_time_ms": execution_time_ms,
        "success": success,
    }

    # JSONL 포맷으로 append
    with open(stats_file, "a", encoding="utf-8") as f:
        f.write(json.dumps(performance_record, ensure_ascii=False) + "\n")

    # 시스템 메시지 반환
    status_emoji = "✅" if success else "❌"
    time_str = f"{execution_time_ms / 1000:.1f}s"

    return {
        "continue": True,
        "systemMessage": f"{status_emoji} {agent_name} completed in {time_str}"
    }


def main():
    """
    Claude Code Hook Entry Point

    STDIN: JSON with:
    - agentId (v2.0.42)
    - agentName (v2.0.42)
    - agentTranscriptPath (v2.0.42)
    - executionTime: 실행 시간 (밀리초)
    - success: 성공 여부
    """
    try:
        hook_input = json.loads(sys.stdin.read())

        # v2.0.42 필드
        agent_id = hook_input.get("agentId", "unknown")
        agent_name = hook_input.get("agentName", "unknown")
        transcript_path = hook_input.get("agentTranscriptPath", "")
        execution_time_ms = hook_input.get("executionTime", 0)
        success = hook_input.get("success", False)

        result = track_agent_lifecycle(
            agent_id=agent_id,
            agent_name=agent_name,
            transcript_path=transcript_path,
            execution_time_ms=execution_time_ms,
            success=success,
        )
        print(json.dumps(result))

    except Exception as e:
        # Hook 실패 시 graceful degradation
        print(json.dumps({
            "continue": True,
            "systemMessage": f"⚠️ Lifecycle tracking failed: {str(e)}"
        }))


if __name__ == "__main__":
    main()
