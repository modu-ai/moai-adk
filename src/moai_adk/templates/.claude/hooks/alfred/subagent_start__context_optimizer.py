#!/usr/bin/env python3
"""
SubagentStart Hook: Optimize context for each subagent

Claude Code v2.0.43 신규 기능:
- agent_id: 시작하는 에이전트 식별자
- agent_name: 에이전트 이름 (spec-builder, tdd-implementer 등)

역할:
1. 에이전트별 최적의 컨텍스트 전략 정의
2. 필요한 Skill 자동 로드
3. 토큰 예산 최적화
4. 메타데이터 저장
"""

import json
import sys
from pathlib import Path
from datetime import datetime


def optimize_agent_context(agent_name: str, agent_id: str) -> dict:
    """
    에이전트별 최적화된 컨텍스트 전략 제공

    v2.0.43: Skills frontmatter 자동 로딩으로 스킬 로드는 자동화됨
    이 Hook은 컨텍스트 최적화 및 메타데이터 기록 담당
    """

    # 에이전트별 컨텍스트 전략
    context_strategies = {
        "spec-builder": {
            "description": "SPEC 작성 - 최소 컨텍스트 로드",
            "max_tokens": 20000,  # Sonnet: 충분한 컨텍스트
            "priority_files": [".moai/specs/", ".moai/config/config.json"],
            "auto_load_skills": True,
        },
        "tdd-implementer": {
            "description": "TDD 구현 - 코드/테스트만 로드",
            "max_tokens": 30000,
            "priority_files": ["src/", "tests/", "pyproject.toml"],
            "auto_load_skills": True,
        },
        "backend-expert": {
            "description": "백엔드 설계 - API/DB 파일 로드",
            "max_tokens": 30000,
            "priority_files": ["src/", "pyproject.toml"],
            "auto_load_skills": True,
        },
        "frontend-expert": {
            "description": "프론트엔드 설계 - UI 컴포넌트만 로드",
            "max_tokens": 25000,
            "priority_files": ["src/components/", "src/pages/", "package.json"],
            "auto_load_skills": True,
        },
        "database-expert": {
            "description": "데이터베이스 설계 - 스키마 파일 로드",
            "max_tokens": 20000,
            "priority_files": [".moai/docs/schema/", "migrations/", "pyproject.toml"],
            "auto_load_skills": True,
        },
        "security-expert": {
            "description": "보안 분석 - 전체 코드베이스 분석",
            "max_tokens": 50000,  # 광범위한 분석 필요
            "priority_files": ["src/", "tests/", ".moai/config/"],
            "auto_load_skills": True,
        },
        "docs-manager": {
            "description": "문서 생성 - 최소 컨텍스트",
            "max_tokens": 15000,
            "priority_files": [".moai/specs/", "README.md", "src/"],
            "auto_load_skills": True,
        },
        "quality-gate": {
            "description": "품질 검증 - 현재 코드만 로드",
            "max_tokens": 15000,
            "priority_files": ["src/", "tests/"],
            "auto_load_skills": True,
        },
    }

    # 기본 전략
    default_strategy = {
        "description": "기타 에이전트 - 표준 컨텍스트",
        "max_tokens": 20000,
        "priority_files": ["src/", ".moai/config/"],
        "auto_load_skills": True,
    }

    strategy = context_strategies.get(agent_name, default_strategy)

    # 에이전트 메타데이터 저장 (v2.0.42의 agent_transcript_path와 연계)
    metadata = {
        "agent_id": agent_id,
        "agent_name": agent_name,
        "started_at": datetime.now().isoformat(),
        "strategy": strategy["description"],
        "max_tokens": strategy["max_tokens"],
        "auto_load_skills": strategy["auto_load_skills"],
        "priority_files": strategy["priority_files"],
    }

    # 메타데이터 저장
    metadata_dir = Path(".moai/logs/agent-transcripts")
    metadata_dir.mkdir(parents=True, exist_ok=True)

    metadata_file = metadata_dir / f"agent-{agent_id}.json"
    metadata_file.write_text(json.dumps(metadata, indent=2, ensure_ascii=False))

    return {
        "continue": True,
        "systemMessage": f"🎯 {strategy['description']} (Max {strategy['max_tokens']} tokens)"
    }


def main():
    """
    Claude Code Hook Entry Point

    STDIN: JSON with:
    - agentId (v2.0.43)
    - agentName (v2.0.43)
    - prompt: 에이전트에 전달되는 초기 프롬프트
    """
    try:
        hook_input = json.loads(sys.stdin.read())

        # v2.0.43 필드
        agent_id = hook_input.get("agentId", "unknown")
        agent_name = hook_input.get("agentName", "unknown")

        result = optimize_agent_context(agent_name, agent_id)
        print(json.dumps(result))

    except Exception as e:
        # Hook 실패 시 graceful degradation
        print(json.dumps({
            "continue": True,
            "systemMessage": f"⚠️ Context optimization skipped: {str(e)}"
        }))


if __name__ == "__main__":
    main()
