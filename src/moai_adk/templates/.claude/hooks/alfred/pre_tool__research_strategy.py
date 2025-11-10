#!/usr/bin/env python3
# @CODE:HOOK-RESEARCH-002 | @SPEC:HOOK-RESEARCH-STRATEGY-001 | @TEST: tests/hooks/test_research_strategy.py
"""실시간 연구 전략 선택 Hook

PreToolUse 단계에서 작업 유형에 맞는 연구 전략을 자동으로 선택하고 최적화.
작업 전 연구 컨텍스트를 설정하고 자원을 할당.

기능:
- 작업 유형 분석 기반 연구 전략 선택
- 자동 연구 컨텍스트 설정
- 자원 최적화 및 메모리 관리
- 병렬 연구 처리 준비
- 연구 지식 JIT 로딩

사용법:
    python3 pre_tool__research_strategy.py <tool_name> <tool_args_json>
"""

import os

import json
import sys
import time
from pathlib import Path
from typing import Any, Dict, List, Optional, Tuple

# 모듈 경로 추가
sys.path.insert(0, str(Path(__file__).parent.parent.parent.parent / "src"))

from moai_adk.core.tags.validator import CentralValidationResult, CentralValidator, ValidationConfig
from moai_adk.statusline.version_reader import VersionReader

# Local hook configuration functions
def get_graceful_degradation() -> bool:
    """우아한 저하 degrade 모드 설정 반환"""
    return os.environ.get("MOAI_GRACEFUL_DEGRADATION", "true").lower() == "true"

def load_hook_timeout() -> int:
    """Hook 타임아웃 로드 (밀리초 단위)"""
    try:
        return int(os.environ.get("MOAI_HOOK_TIMEOUT_MS", "30000"))
    except ValueError:
        return 30000


def load_research_config() -> Dict[str, Any]:
    """연구 설정 로드

    Returns:
        연구 설정 딕셔너리
    """
    try:
        config_file = Path(".moai/config.json")
        if config_file.exists():
            with open(config_file, 'r', encoding='utf-8') as f:
                config = json.load(f)
                return config.get("tags", {}).get("research_tags", {})
    except Exception:
        pass

    return {
        "auto_discovery": True,
        "pattern_matching": True,
        "cross_reference": True,
        "knowledge_graph": True,
        "research_categories": ["RESEARCH", "ANALYSIS", "KNOWLEDGE", "INSIGHT"],
        "research_patterns": {
            "RESEARCH": ["@RESEARCH:", "research", "investigate", "analyze"],
            "ANALYSIS": ["@ANALYSIS:", "analysis", "evaluate", "assess"],
            "KNOWLEDGE": ["@KNOWLEDGE:", "knowledge", "learn", "pattern"],
            "INSIGHT": ["@INSIGHT:", "insight", "innovate", "optimize"]
        }
    }


def classify_tool_type(tool_name: str, tool_args: Dict[str, Any]) -> str:
    """도구 유형 분류

    Args:
        tool_name: 도구 이름
        tool_args: 도구 인자

    Returns:
        분류된 도구 유형
    """
    # 코드 작업 도구
    code_tools = {"Edit", "Write", "MultiEdit", "Read", "Grep", "Glob"}

    # 테스트 관련 도구
    test_tools = {"Bash"}

    # 문서 작업 도구
    doc_tools = {"NotebookEdit"}

    # 탐색 및 분석 도구
    explore_tools = {"Task", "Explore", "Plan", "WebFetch", "WebSearch"}

    # 프로젝트 관리 도구
    project_tools = {"Bash(git:*)", "Bash(gh:*)"}

    if tool_name in code_tools:
        return "code"
    elif tool_name in test_tools:
        return "test"
    elif tool_name in doc_tools:
        return "documentation"
    elif tool_name in explore_tools:
        return "exploration"
    elif tool_name in project_tools:
        return "project"
    else:
        return "general"


def get_research_strategies_for_tool(tool_type: str) -> Dict[str, Any]:
    """도구 유형에 적합한 연구 전략 선택

    Args:
        tool_type: 도구 유형

    Returns:
        선택된 연구 전략
    """
    config = load_research_config()

    # 전략 매핑 (도구 유형 -> 적합한 전략)
    strategy_mapping = {
        "code": {
            "primary_strategies": [
                "pattern_recognition",
                "systematic_elimination",
                "first_principles"
            ],
            "secondary_strategies": [
                "root_cause_analysis",
                "cross_domain_analysis"
            ],
            "focus_areas": [
                "code_quality",
                "architecture_consistency",
                "performance_optimization"
            ],
            "knowledge_categories": ["KNOWLEDGE", "INSIGHT"]
        },
        "test": {
            "primary_strategies": [
                "probabilistic_thinking",
                "systematic_elimination",
                "resource_optimization"
            ],
            "secondary_strategies": [
                "pattern_recognition",
                "continuous_learning"
            ],
            "focus_areas": [
                "test_coverage",
                "edge_case_coverage",
                "performance_testing"
            ],
            "knowledge_categories": ["ANALYSIS", "RESEARCH"]
        },
        "documentation": {
            "primary_strategies": [
                "pattern_recognition",
                "cross_domain_analysis",
                "first_principles"
            ],
            "secondary_strategies": [
                "knowledge_graph_building",
                "continuous_learning"
            ],
            "focus_areas": [
                "clarity",
                "completeness",
                "consistency"
            ],
            "knowledge_categories": ["KNOWLEDGE", "INSIGHT"]
        },
        "exploration": {
            "primary_strategies": [
                "cross_domain_analysis",
                "pattern_recognition",
                "first_principles"
            ],
            "secondary_strategies": [
                "probabilistic_thinking",
                "continuous_learning"
            ],
            "focus_areas": [
                "comprehensive_analysis",
                "deep_understanding",
                "insight_generation"
            ],
            "knowledge_categories": ["RESEARCH", "ANALYSIS", "KNOWLEDGE", "INSIGHT"]
        },
        "project": {
            "primary_strategies": [
                "resource_optimization",
                "systematic_elimination",
                "probabilistic_thinking"
            ],
            "secondary_strategies": [
                "cross_domain_analysis",
                "continuous_learning"
            ],
            "focus_areas": [
                "efficiency",
                "scalability",
                "maintainability"
            ],
            "knowledge_categories": ["ANALYSIS", "INSIGHT"]
        },
        "general": {
            "primary_strategies": [
                "pattern_recognition",
                "resource_optimization",
                "continuous_learning"
            ],
            "secondary_strategies": [
                "systematic_elimination",
                "probabilistic_thinking"
            ],
            "focus_areas": [
                "general_optimization",
                "best_practice_application"
            ],
            "knowledge_categories": ["RESEARCH", "KNOWLEDGE"]
        }
    }

    return strategy_mapping.get(tool_type, strategy_mapping["general"])


def load_jit_knowledge(tool_type: str, focus_areas: List[str]) -> Dict[str, Any]:
    """Just-In-Time 지식 로딩

    Args:
        tool_type: 도구 유형
        focus_areas: 초점 영역

    Returns:
        JIT 로딩된 지식
    """
    knowledge_base = {}
    knowledge_dir = Path(".moai/research/knowledge/")

    if not knowledge_dir.exists():
        return knowledge_base

    # 도구 유형과 초점 영역에 맞는 지식 파일 로드
    relevant_files = []

    # 일치하는 파일 패턴
    patterns = [
        f"{tool_type}_*.json",
        f"*.json"
    ]

    for pattern in patterns:
        try:
            for file_path in knowledge_dir.glob(pattern):
                if file_path.is_file():
                    relevant_files.append(file_path)
        except Exception:
            continue

    # 파일 로드
    for file_path in relevant_files:
        try:
            import json as json_module
            with open(file_path, 'r', encoding='utf-8') as f:
                knowledge_data = json_module.load(f)

                # 초점 영역과 관련된 지식만 필터링
                if focus_areas:
                    relevant_knowledge = {}
                    for area in focus_areas:
                        if area.lower() in str(file_path).lower() or area.lower() in json.dumps(knowledge_data).lower():
                            relevant_knowledge.update(knowledge_data)
                    knowledge_base[file_path.stem] = relevant_knowledge
                else:
                    knowledge_base[file_path.stem] = knowledge_data

        except Exception:
            continue

    return knowledge_base


def optimize_resources(tool_type: str, strategies: List[str]) -> Dict[str, Any]:
    """자원 최적화 설정

    Args:
        tool_type: 도구 유형
        strategies: 선택된 전략

    Returns:
        최적화된 자원 설정
    """
    # 도구 유형에 따른 자원 할당
    resource_configs = {
        "code": {
            "memory_limit": "512MB",
            "timeout_seconds": 45,
            "priority": "high",
            "parallel_processing": False,
            "cache_enabled": True
        },
        "test": {
            "memory_limit": "256MB",
            "timeout_seconds": 30,
            "priority": "medium",
            "parallel_processing": True,
            "cache_enabled": True
        },
        "documentation": {
            "memory_limit": "128MB",
            "timeout_seconds": 20,
            "priority": "low",
            "parallel_processing": False,
            "cache_enabled": False
        },
        "exploration": {
            "memory_limit": "1GB",
            "timeout_seconds": 60,
            "priority": "high",
            "parallel_processing": True,
            "cache_enabled": True
        },
        "project": {
            "memory_limit": "256MB",
            "timeout_seconds": 30,
            "priority": "medium",
            "parallel_processing": False,
            "cache_enabled": True
        },
        "general": {
            "memory_limit": "256MB",
            "timeout_seconds": 25,
            "priority": "medium",
            "parallel_processing": False,
            "cache_enabled": True
        }
    }

    config = resource_configs.get(tool_type, resource_configs["general"])

    # 전략 수에 따라 추가 최적화
    if len(strategies) > 3:
        config["memory_limit"] = f"{int(config['memory_limit'].replace('MB', '')) * 1.5:.0f}MB"
        config["parallel_processing"] = True

    return config


def create_research_context(tool_name: str, tool_args: Dict[str, Any],
                          strategy_config: Dict[str, Any],
                          knowledge_base: Dict[str, Any],
                          resource_config: Dict[str, Any]) -> Dict[str, Any]:
    """연구 컨텍스트 생성

    Args:
        tool_name: 도구 이름
        tool_args: 도구 인자
        strategy_config: 전략 설정
        knowledge_base: 지식 베이스
        resource_config: 자원 설정

    Returns:
        연구 컨텍스트
    """
    tool_type = classify_tool_type(tool_name, tool_args)

    return {
        "tool_context": {
            "name": tool_name,
            "type": tool_type,
            "args": tool_args
        },
        "research_strategy": strategy_config,
        "knowledge_base": knowledge_base,
        "resource_config": resource_config,
        "session_id": f"research_{int(time.time())}",
        "timestamp": time.strftime("%Y-%m-%d %H:%M:%S"),
        "optimization_applied": True
    }


def format_strategy_message(context: Dict[str, Any]) -> str:
    """전략 선택 메시지 생성

    Args:
        context: 연구 컨텍스트

    Returns:
        포맷된 메시지
    """
    strategy_names = context["research_strategy"]["primary_strategies"]
    focus_areas = context["research_strategy"]["focus_areas"]
    resource_config = context["resource_config"]

    message = [
        f"🔬 연구 전략 선택 완료",
        f"📝 도구: {context['tool_context']['name']} ({context['tool_context']['type']})",
        f"⚡ 활성화된 전략: {', '.join(strategy_names)}",
        f"🎯 초점 영역: {', '.join(focus_areas)}",
        f"💾 자원 설정: {resource_config['memory_limit']}, 타임아웃: {resource_config['timeout_seconds']}초"
    ]

    if context["knowledge_base"]:
        message.append(f"📚 JIT 지식 로딩: {len(context['knowledge_base'])} 개 항목")

    return "\n".join(message)


def main() -> None:
    """메인 함수"""
    try:
        # 설정 로드
        timeout_seconds = load_hook_timeout() / 1000
        graceful_degradation = get_graceful_degradation()

        # 인자 파싱
        if len(sys.argv) < 3:
            if os.environ.get("MOAI_SILENT_RESEARCH", "false").lower() != "true":
                print(json.dumps({
                    "research_strategy_selected": False,
                    "error": "Invalid arguments. Usage: python3 pre_tool__research_strategy.py <tool_name> <tool_args_json>"
                }))
            sys.exit(0)

        tool_name = sys.argv[1]
        try:
            tool_args = json.loads(sys.argv[2])
        except json.JSONDecodeError:
            if os.environ.get("MOAI_SILENT_RESEARCH", "false").lower() != "true":
                print(json.dumps({
                    "research_strategy_selected": False,
                    "error": "Invalid tool_args JSON"
                }))
            sys.exit(0)

        # 시작 시간 기록
        start_time = time.time()

        # 도구 유형 분류
        tool_type = classify_tool_type(tool_name, tool_args)

        # 연구 전략 선택
        strategy_config = get_research_strategies_for_tool(tool_type)

        # 초점 영역 추출
        focus_areas = strategy_config.get("focus_areas", [])

        # JIT 지식 로딩
        knowledge_base = load_jit_knowledge(tool_type, focus_areas)

        # 자원 최적화
        resource_config = optimize_resources(tool_type, strategy_config["primary_strategies"])

        # 연구 컨텍스트 생성
        research_context = create_research_context(
            tool_name, tool_args, strategy_config,
            knowledge_base, resource_config
        )

        # 타임아웃 체크
        if time.time() - start_time > timeout_seconds:
            timeout_response = {
                "research_strategy_selected": False,
                "error": "연구 전략 선택 타임아웃",
                "message": "연구 전략 선택 실패 - 기본 전략으로 진행",
                "graceful_degradation": graceful_degradation
            }
            if os.environ.get("MOAI_SILENT_RESEARCH", "false").lower() != "true":
                print(json.dumps(timeout_response, ensure_ascii=False))
            sys.exit(0)

        # 실행 시간 계산
        execution_time_ms = (time.time() - start_time) * 1000

        # 최종 응답
        response = {
            **research_context,
            "execution_time_ms": execution_time_ms,
            "research_strategy_selected": True,
            "message": format_strategy_message(research_context)
        }

        # 성능 경고
        timeout_warning_ms = timeout_seconds * 1000 * 0.8
        if execution_time_ms > timeout_warning_ms:
            response["performance_warning"] = f"전략 선택 시간이 타임아웃의 80%를 초과했습니다 ({execution_time_ms:.0f}ms / {timeout_warning_ms:.0f}ms)"

        if os.environ.get("MOAI_SILENT_RESEARCH", "false").lower() != "true":
            print(json.dumps(response, ensure_ascii=False, indent=2))

    except Exception as e:
        # 예외 처리
        error_response = {
            "research_strategy_selected": False,
            "error": f"연구 전략 선택 오류: {str(e)}",
            "message": "연구 전략 선택 실패 - 기본 전략으로 진행"
        }

        if graceful_degradation:
            error_response["graceful_degradation"] = True

        if os.environ.get("MOAI_SILENT_RESEARCH", "false").lower() != "true":
            print(json.dumps(error_response, ensure_ascii=False))
        sys.exit(1)


if __name__ == "__main__":
    main()