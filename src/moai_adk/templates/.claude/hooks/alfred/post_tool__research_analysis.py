#!/usr/bin/env python3
# @CODE:HOOK-RESEARCH-003 | @SPEC:HOOK-RESEARCH-ANALYSIS-001 | @TEST: tests/hooks/test_research_analysis.py
"""작업 후 연구 분석 Hook

PostToolUse 단계에서 작업 결과를 분석하고 연구 인사이트를 추출.
지식 베이스 업데이트와 연구 데이터 동기화.

기능:
- 작업 결과 연구 분석
- 연구 인사이트 추출
- 지식 베이스 자동 업데이트
- 연구 데이터 동기화
- 성능 및 품질 분석

사용법:
    python3 post_tool__research_analysis.py <tool_name> <tool_result_json> <execution_time_ms>
"""

import json
import os
import sys
import time
from pathlib import Path
from typing import Any, Dict, List, Optional

# 모듈 경로 추가
sys.path.insert(0, str(Path(__file__).parent.parent.parent.parent / "src"))

from moai_adk.core.tags.validator import CentralValidationResult, CentralValidator, ValidationConfig
from moai_adk.statusline.version_reader import VersionReader

# Local hook configuration functions
from utils.hook_config import get_graceful_degradation, load_hook_timeout


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
        "auto_update_knowledge": True,
        "research_categories": ["RESEARCH", "ANALYSIS", "KNOWLEDGE", "INSIGHT"]
    }


def analyze_tool_result(tool_name: str, tool_result: Dict[str, Any]) -> Dict[str, Any]:
    """도구 결과 분석

    Args:
        tool_name: 도구 이름
        tool_result: 도구 실행 결과

    Returns:
        분석 결과
    """
    analysis = {
        "tool_name": tool_name,
        "result_type": type(tool_result).__name__,
        "execution_successful": True,
        "analysis_timestamp": time.strftime("%Y-%m-%d %H:%M:%S"),
        "insights": [],
        "patterns": [],
        "optimizations": [],
        "errors": []
    }

    # 성공/실패 분석
    if tool_result.get("continue", True) is False:
        analysis["execution_successful"] = False
        analysis["errors"].append("도구 실행이 중단되었습니다")

    # 에러 분석
    if "error" in tool_result:
        analysis["execution_successful"] = False
        analysis["errors"].append(tool_result["error"])

    # 파일 작업 결과 분석
    if tool_name in ["Edit", "Write", "MultiEdit"]:
        if "hookSpecificOutput" in tool_result:
            hook_output = tool_result["hookSpecificOutput"]
            if isinstance(hook_output, dict):
                if "file_path" in hook_output:
                    analysis["modified_files"] = [hook_output["file_path"]]
                if "changes_applied" in hook_output:
                    analysis["changes_applied"] = hook_output["changes_applied"]

    # 탐색 도구 결과 분석
    if tool_name in ["Task", "Explore", "Plan"]:
        if "result" in tool_result:
            result = tool_result["result"]
            if isinstance(result, dict):
                if "findings" in result:
                    analysis["insights"].extend(result["findings"])
                if "recommendations" in result:
                    analysis["optimizations"].extend(result["recommendations"])

    # 검증 도구 결과 분석
    if "validation_result" in tool_result:
        validation = tool_result["validation_result"]
        if isinstance(validation, dict):
            if "is_valid" in validation:
                analysis["validation_passed"] = validation["is_valid"]
            if "errors" in validation:
                analysis["errors"].extend(validation["errors"])

    return analysis


def extract_insights(analysis: Dict[str, Any]) -> List[str]:
    """인사이트 추출

    Args:
        analysis: 분석 결과

    Returns:
        추출된 인사이트 목록
    """
    insights = []

    # 성공 분석 인사이트
    if analysis["execution_successful"]:
        insights.append(f"{analysis['tool_name']} 작업이 성공적으로 완료되었습니다")

    # 에러 분석 인사이트
    if analysis["errors"]:
        insights.append(f"작업 중 에러 발생: {', '.join(analysis['errors'])}")
        insights.append("에러 패턴 분석이 필요합니다")

    # 패턴 인사이트
    if analysis["patterns"]:
        insights.append(f"발견된 패턴: {', '.join(analysis['patterns'])}")

    # 최적화 인사이트
    if analysis["optimizations"]:
        insights.append(f"제안된 최적화: {', '.join(analysis['optimizations'])}")

    return insights


def update_knowledge_base(insights: List[str], tool_name: str, analysis: Dict[str, Any]) -> Dict[str, Any]:
    """지식 베이스 업데이트

    Args:
        insights: 인사이트 목록
        tool_name: 도구 이름
        analysis: 분석 결과

    Returns:
        업데이트 결과
    """
    try:
        # 연구 지식 디렉토리 확인
        knowledge_dir = Path(".moai/research/knowledge/")
        knowledge_dir.mkdir(parents=True, exist_ok=True)

        # 지식 파일 경로
        knowledge_file = knowledge_dir / f"{tool_name}_insights.json"

        # 기존 지식 로드
        existing_knowledge = {}
        if knowledge_file.exists():
            with open(knowledge_file, 'r', encoding='utf-8') as f:
                existing_knowledge = json.load(f)

        # 새로운 인사이트 추가
        current_timestamp = time.strftime("%Y-%m-%d %H:%M:%S")
        new_entry = {
            "timestamp": current_timestamp,
            "insights": insights,
            "analysis": analysis,
            "execution_successful": analysis["execution_successful"]
        }

        # 지식 병합 (최대 100개 항목 유지)
        if "insights_history" not in existing_knowledge:
            existing_knowledge["insights_history"] = []

        existing_knowledge["insights_history"].append(new_entry)
        existing_knowledge["insights_history"] = existing_knowledge["insights_history"][-100:]

        # 지식 파일 저장
        with open(knowledge_file, 'w', encoding='utf-8') as f:
            json.dump(existing_knowledge, f, ensure_ascii=False, indent=2)

        return {
            "knowledge_updated": True,
            "knowledge_file": str(knowledge_file),
            "insights_count": len(insights),
            "total_history": len(existing_knowledge["insights_history"])
        }

    except Exception as e:
        return {
            "knowledge_updated": False,
            "error": f"지식 베이스 업데이트 실패: {str(e)}"
        }


def generate_research_report(analysis: Dict[str, Any], knowledge_update: Dict[str, Any],
                           execution_time_ms: float) -> Dict[str, Any]:
    """연구 보고서 생성

    Args:
        analysis: 분석 결과
        knowledge_update: 지식 업데이트 결과
        execution_time_ms: 실행 시간

    Returns:
        연구 보고서
    """
    report = {
        "research_analysis_completed": True,
        "tool_name": analysis["tool_name"],
        "execution_successful": analysis["execution_successful"],
        "execution_time_ms": execution_time_ms,
        "timestamp": analysis["analysis_timestamp"],
        "insights_count": len(analysis["insights"]),
        "errors_count": len(analysis["errors"]),
        "knowledge_update": knowledge_update,
        "recommendations": []
    }

    # 추천 생성
    if not analysis["execution_successful"]:
        report["recommendations"].append("에러 패턴 분석 및 재시도 권장")

    if analysis["insights"]:
        report["recommendations"].append("추출된 인사이트를 프로젝트 문서에 반영")

    if analysis["patterns"]:
        report["recommendations"].append("반복되는 패턴을 자동화로 개선")

    if analysis["optimizations"]:
        report["recommendations"].append("제안된 최적화 사항 적용")

    return report


def create_analysis_summary(report: Dict[str, Any]) -> str:
    """분석 요약 생성

    Args:
        report: 연구 보고서

    Returns:
        포맷된 요약 메시지
    """
    summary_parts = [
        f"🔬 연구 분석 완료",
        f"📝 도구: {report['tool_name']}",
        f"⏱️ 실행 시간: {report['execution_time_ms']:.1f}ms",
        f"✅ 성공: {'Yes' if report['execution_successful'] else 'No'}"
    ]

    if report["insights_count"] > 0:
        summary_parts.append(f"💡 인사이트: {report['insights_count']}개")

    if report["errors_count"] > 0:
        summary_parts.append(f"❌ 에러: {report['errors_count']}개")

    if report["knowledge_update"]["knowledge_updated"]:
        summary_parts.append(f"📚 지식 업데이트: {report['knowledge_update']['insights_count']}개")

    if report["recommendations"]:
        summary_parts.append(f"🎯 추천: {len(report['recommendations'])}개")

    return "\n".join(summary_parts)


def main() -> None:
    """메인 함수"""
    try:
        # 설정 로드
        timeout_seconds = load_hook_timeout() / 1000
        graceful_degradation = get_graceful_degradation()

        # 인자 파싱
        if len(sys.argv) < 4:
            print(json.dumps({
                "research_analysis_completed": False,
                "error": "Invalid arguments. Usage: python3 post_tool__research_analysis.py <tool_name> <tool_result_json> <execution_time_ms>"
            }))
            sys.exit(1)

        tool_name = sys.argv[1]
        try:
            tool_result = json.loads(sys.argv[2])
            execution_time_ms = float(sys.argv[3])
        except (json.JSONDecodeError, ValueError) as e:
            print(json.dumps({
                "research_analysis_completed": False,
                "error": f"Invalid arguments: {str(e)}"
            }))
            sys.exit(1)

        # 시작 시간 기록
        start_time = time.time()

        # 설정 로드
        research_config = load_research_config()

        # 도구 결과 분석
        analysis = analyze_tool_result(tool_name, tool_result)

        # 인사이트 추출
        insights = extract_insights(analysis)

        # 지식 베이스 업데이트
        knowledge_update = {}
        if research_config.get("auto_update_knowledge", True):
            knowledge_update = update_knowledge_base(insights, tool_name, analysis)

        # 연구 보고서 생성
        report = generate_research_report(analysis, knowledge_update, execution_time_ms)

        # 타임아웃 체크
        if time.time() - start_time > timeout_seconds:
            error_response = {
                "research_analysis_completed": False,
                "error": "연구 분석 타임아웃",
                "message": "연구 분석 실패 - 정상 작동으로 간주",
                "graceful_degradation": graceful_degradation
            }
            print(json.dumps(error_response, ensure_ascii=False))
            sys.exit(1)

        # 실행 시간 계산
        analysis_time_ms = (time.time() - start_time) * 1000

        # 최종 응답
        response = {
            **report,
            "analysis_time_ms": analysis_time_ms,
            "message": create_analysis_summary(report)
        }

        # 성능 경고
        timeout_warning_ms = timeout_seconds * 1000 * 0.8
        if analysis_time_ms > timeout_warning_ms:
            response["performance_warning"] = f"분석 시간이 타임아웃의 80%를 초과했습니다 ({analysis_time_ms:.0f}ms / {timeout_warning_ms:.0f}ms)"

        print(json.dumps(response, ensure_ascii=False, indent=2))

    except Exception as e:
        # 예외 처리
        error_response = {
            "research_analysis_completed": False,
            "error": f"연구 분석 오류: {str(e)}",
            "message": "연구 분석 실패 - 정상 작동으로 간주"
        }

        if graceful_degradation:
            error_response["graceful_degradation"] = True

        print(json.dumps(error_response, ensure_ascii=False))
        sys.exit(1)


if __name__ == "__main__":
    main()