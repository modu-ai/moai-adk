#!/usr/bin/env python3
"""
TAG 시스템 검증 스크립트

95% 추적성 목표 달성 여부를 검증하고 상세 보고서를 생성합니다.
"""

import json
import time
from pathlib import Path
from typing import Dict, Any, List

from moai_adk.core.tags.validator import CentralValidator, ValidationConfig
from moai_adk.core.tags.auto_corrector import TagAutoCorrector, AutoCorrectionConfig
from moai_adk.core.tags.tags import suggest_tag_for_file, validate_tag_chain


def validate_tag_system_integrity() -> Dict[str, Any]:
    """TAG 시스템 무결성 검증"""
    print("🔍 TAG 시스템 무결성 검증 시작...")

    # 검증 설정
    validator_config = ValidationConfig(
        strict_mode=False,
        check_duplicates=True,
        check_orphans=True,
        check_chain_integrity=True,
        allowed_file_types=["py", "js", "ts", "jsx", "tsx", "md", "txt"]
    )

    # 검증기 생성
    validator = CentralValidator(config=validator_config)

    # 전체 프로젝트 검증
    start_time = time.time()
    result = validator.validate_directory("/Users/goos/MoAI/MoAI-ADK")
    validation_time = time.time() - start_time

    # 결과 분석
    analysis = {
        "validation_result": result.to_dict(),
        "validation_time_ms": validation_time,
        "timestamp": result.timestamp.isoformat(),
        "traceability_percentage": result.statistics.coverage_percentage,
        "total_issues": len(result.issues),
        "error_count": len(result.errors),
        "warning_count": len(resultwarnings),
        "target_achievement": result.statistics.coverage_percentage >= 95.0
    }

    return analysis


def validate_tag_chains() -> Dict[str, Any]:
    """TAG 체인 무결성 검증"""
    print("🔗 TAG 체인 무결성 검증 시작...")

    chains_tested = 0
    chains_valid = 0
    chains_invalid = []

    # 테스트할 체인 목록
    test_chains = [
        {
            "tag": "@DOC:TAG-IMPROVEMENT-PLAN-001",
            "chain": "@SPEC:TAG-IMPROVEMENT-PLAN-001 -> @CODE:TAG-IMPROVEMENT-PLAN-001 -> @TEST:TAG-IMPROVEMENT-PLAN-001 -> @DOC:TAG-IMPROVEMENT-PLAN-001"
        },
        {
            "tag": "@DOC:TAG-SYSTEM-VALIDATION-REPORT-001",
            "chain": "@SPEC:TAG-SYSTEM-VALIDATION-REPORT-001 -> @CODE:TAG-SYSTEM-VALIDATION-REPORT-001 -> @TEST:TAG-SYSTEM-VALIDATION-REPORT-001 -> @DOC:TAG-SYSTEM-VALIDATION-REPORT-001"
        },
        {
            "tag": "@CODE:TEMPLATE-ENGINE-001",
            "chain": "@SPEC:TEMPLATE-SYSTEM-001 -> @CODE:TEMPLATE-ENGINE-001 -> @TEST:TEMPLATE-SYSTEM-001 -> @DOC:TEMPLATE-SYSTEM-001"
        }
    ]

    for chain_info in test_chains:
        is_valid = validate_tag_chain(chain_info["tag"], chain_info["chain"])
        chains_tested += 1
        if is_valid:
            chains_valid += 1
        else:
            chains_invalid.append(chain_info)

    return {
        "chains_tested": chains_tested,
        "chains_valid": chains_valid,
        "chains_invalid": chains_invalid,
        "chain_integrity_percentage": (chains_valid / chains_tested) * 100 if chains_tested > 0 else 0
    }


def validate_auto_correction_system() -> Dict[str, Any]:
    """자동 수정 시스템 검증"""
    print("🔧 자동 수정 시스템 검증 시작...")

    # 자동 수정 설정
    correction_config = AutoCorrectionConfig(
        enable_auto_fix=True,
        confidence_threshold=0.8,
        create_missing_specs=False,
        create_missing_tests=False,
        remove_duplicates=True
    )

    corrector = TagAutoCorrector(config=correction_config)

    # 신뢰도 테스트
    confidence_tests = []
    test_files = [
        "src/moai_adk/core/tags/validator.py",
        "src/moai_adk/core/tags/auto_corrector.py",
        "src/moai_adk/core/template_engine.py"
    ]

    for file_path in test_files:
        tag_suggestion = corrector.suggest_tag_for_code_file(file_path)
        if tag_suggestion:
            tag, confidence = tag_suggestion
            confidence_tests.append({
                "file": file_path,
                "suggested_tag": tag,
                "confidence": confidence,
                "meets_threshold": confidence >= correction_config.confidence_threshold
            })

    return {
        "confidence_threshold": correction_config.confidence_threshold,
        "confidence_tests": confidence_tests,
        "average_confidence": sum(test["confidence"] for test in confidence_tests) / len(confidence_tests) if confidence_tests else 0,
        "high_confidence_ratio": len([test for test in confidence_tests if test["confidence"] >= 0.9]) / len(confidence_tests) if confidence_tests else 0
    }


def generate_comprehensive_report() -> Dict[str, Any]:
    """종합 보고서 생성"""
    print("📊 종합 보고서 생성 시작...")

    # 각 시스템 검증
    validation_result = validate_tag_system_integrity()
    chain_result = validate_tag_chains()
    auto_correction_result = validate_auto_correction_system()

    # 종합 평가
    target_achievement = validation_result["traceability_percentage"] >= 95.0
    overall_score = (
        validation_result["traceability_percentage"] * 0.4 +
        chain_result["chain_integrity_percentage"] * 0.3 +
        auto_correction_result["average_confidence"] * 0.3
    )

    return {
        "validation_timestamp": time.strftime("%Y-%m-%d %H:%M:%S"),
        "target_achievement": target_achievement,
        "overall_score": round(overall_score, 2),
        "traceability_target": 95.0,
        "actual_traceability": validation_result["traceability_percentage"],
        "chain_integrity_percentage": chain_result["chain_integrity_percentage"],
        "auto_correction_confidence": auto_correction_result["average_confidence"],
        "validation_details": validation_result,
        "chain_details": chain_result,
        "auto_correction_details": auto_correction_result,
        "recommendations": generate_recommendations(validation_result, chain_result, auto_correction_result)
    }


def generate_recommendations(validation: Dict, chain: Dict, auto_correction: Dict) -> List[str]:
    """개선 권장사항 생성"""
    recommendations = []

    if validation["traceability_percentage"] < 95.0:
        recommendations.append(f"TAG 추적성이 {validation['traceability_percentage']}%로 목표 95%에 미달합니다. 잔여 TAG를 추가로 배치하세요.")

    if chain["chain_integrity_percentage"] < 95.0:
        recommendations.append(f"TAG 체인 무결성이 {chain['chain_integrity_percentage']}%로 개선이 필요합니다. 체인 연결을 강화하세요.")

    if auto_correction["average_confidence"] < 0.8:
        recommendations.append(f"자동 수정 신뢰도가 {auto_correction['average_confidence']:.2f}로 낮습니다. 알고리즘 개선이 필요합니다.")

    if validation["total_issues"] > 0:
        recommendations.append(f"총 {validation['total_issues']}개의 TAG 이슈가 발견되었습니다. 상세히 검토하세요.")

    if not recommendations:
        recommendations.append("TAG 시스템이 목표를 달성했습니다. 지속적인 모니터링을 유지하세요.")

    return recommendations


def main():
    """메인 실행 함수"""
    print("🚀 TAG 시스템 종합 검증 시작")
    print("=" * 50)

    # 종합 보고서 생성
    report = generate_comprehensive_report()

    # 결과 출력
    print("\n📋 검증 결과 요약")
    print("=" * 50)
    print(f"검증 시간: {report['validation_timestamp']}")
    print(f"목표 달성 여부: {'✅ 달성' if report['target_achievement'] else '❌ 미달성'}")
    print(f"종합 점수: {report['overall_score']}/100")
    print(f"TAG 추적성: {report['actual_traceability']:.1f}% (목표: {report['traceability_target']:.1f}%)")
    print(f"체인 무결성: {report['chain_integrity_percentage']:.1f}%")
    print(f"자동 수정 신뢰도: {report['auto_correction_confidence']:.1f}")

    print("\n🎯 개선 권장사항")
    print("=" * 50)
    for i, recommendation in enumerate(report['recommendations'], 1):
        print(f"{i}. {recommendation}")

    # 보고서 저장
    report_file = Path(".moai/reports/tag-system-validation-report.json")
    report_file.parent.mkdir(parents=True, exist_ok=True)

    with open(report_file, 'w', encoding='utf-8') as f:
        json.dump(report, f, indent=2, ensure_ascii=False)

    print(f"\n📄 상세 보고서 저장 완료: {report_file}")

    # 성공 메시지
    if report['target_achievement']:
        print("\n🎉 목표 달성! TAG 시스템이 95% 추적성 목표를 성공적으로 달성했습니다.")
    else:
        print(f"\n⚠️ 목표 미달성. 현재 추적성: {report['actual_traceability']:.1f}%")

    return report


if __name__ == "__main__":
    main()