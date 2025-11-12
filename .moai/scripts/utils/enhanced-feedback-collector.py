#!/usr/bin/env python3
"""
향상된 자동 피드백 정보 수집 및 분석 시스템

기존 시스템을 확장하여 상황 인지 기반 자동 분석, 개선 제안, 동적 템플릿 생성 기능을 제공합니다.

Version: 1.0.0 (2025-11-13)
Maintained by: MoAI-ADK Team
"""

import json
import sys
import os
import subprocess
from pathlib import Path
from datetime import datetime, timedelta
from typing import Dict, List, Any, Optional
import argparse
import re
from collections import defaultdict


class EnhancedFeedbackCollector:
    """향상된 피드백 수집 및 분석 시스템"""

    def __init__(self, config_file: str = None):
        self.config = self._load_config(config_file)
        self.context_data = {}
        self.analysis_results = {}

    def _load_config(self, config_file: str) -> Dict:
        """설정 파일 로드"""
        default_config = {
            "auto_analysis": True,
            "pattern_detection": True,
            "suggestion_engine": True,
            "context_awareness": True,
            "template_adaptation": True,
            "output_format": "korean"  # korean or json
        }

        if config_file and Path(config_file).exists():
            try:
                import yaml
                with open(config_file, 'r', encoding='utf-8') as f:
                    return {**default_config, **yaml.safe_load(f)}
            except Exception:
                pass

        return default_config

    def collect_comprehensive_feedback(self) -> Dict[str, Any]:
        """종합 피드백 정보 수집"""
        print("🔍 종합 피드백 정보 수집 중...")

        # 1. 기본 환경 정보 수집
        basic_info = self._collect_basic_environment()
        print("✅ 기본 환경 정보 수집 완료")

        # 2. 상황 인지 기반 분석
        context_analysis = self._perform_context_analysis(basic_info)
        print("✅ 상황 인지 분석 완료")

        # 3. 패턴 분석 실행
        pattern_analysis = self._analyze_patterns()
        print("✅ 패턴 분석 완료")

        # 4. 개선 제안 생성
        suggestions = self._generate_suggestions(basic_info, context_analysis, pattern_analysis)
        print("✅ 개선 제안 생성 완료")

        # 5. 동적 템플릿 생성
        template = self._create_dynamic_template(basic_info, context_analysis, suggestions)
        print("✅ 동적 템플릿 생성 완료")

        # 종합 데이터 구성
        comprehensive_data = {
            "timestamp": datetime.now().isoformat(),
            "collection_type": "enhanced_comprehensive",
            "basic_environment": basic_info,
            "context_analysis": context_analysis,
            "pattern_analysis": pattern_analysis,
            "suggestions": suggestions,
            "dynamic_template": template,
            "analysis_summary": self._generate_analysis_summary(
                basic_info, context_analysis, pattern_analysis, suggestions
            )
        }

        self.context_data = comprehensive_data
        return comprehensive_data

    def _collect_basic_environment(self) -> Dict[str, Any]:
        """기본 환경 정보 수집 (기존 스크립트 통합)"""
        environment = {}

        # MoAI-ADK 환경 정보
        try:
            config_path = Path(".moai/config/config.json")
            if config_path.exists():
                with open(config_path, 'r', encoding='utf-8') as f:
                    config = json.load(f)
                    environment["moai_config"] = config
                    environment["moai_version"] = config.get("moai", {}).get("version", "unknown")
                    environment["project_mode"] = config.get("project", {}).get("mode", "unknown")
                    environment["conversation_language"] = config.get("language", {}).get("conversation_language", "unknown")
        except Exception:
            pass

        # 시스템 정보
        try:
            import platform
            environment["system_info"] = f"{platform.system()} {platform.release()}"
            environment["python_version"] = f"{platform.python_version()}"
        except Exception:
            pass

        # Git 상태
        try:
            result = subprocess.run(
                ["git", "branch", "--show-current"],
                capture_output=True,
                text=True,
                timeout=2
            )
            environment["current_branch"] = result.stdout.strip() if result.returncode == 0 else "unknown"

            result = subprocess.run(
                ["git", "status", "--porcelain"],
                capture_output=True,
                text=True,
                timeout=2
            )
            if result.returncode == 0:
                uncommitted = [line for line in result.stdout.strip().split('\n') if line]
                environment["uncommitted_changes"] = len(uncommitted)
                environment["uncommitted_files"] = uncommitted[:3]  # 최대 3개 파일 표시
        except Exception:
            pass

        # 현재 작업 SPEC 감지
        current_branch = environment.get("current_branch", "")
        if current_branch.startswith("feature/SPEC-"):
            environment["current_spec"] = current_branch.replace("feature/", "")
        else:
            environment["current_spec"] = ""

        # 최근 활동 분석
        environment["recent_activity"] = self._analyze_recent_activity()

        return environment

    def _analyze_recent_activity(self) -> Dict[str, Any]:
        """최근 활동 분석"""
        activity = {}

        # 최근 커밋
        try:
            result = subprocess.run(
                ["git", "log", "-5", "--oneline"],
                capture_output=True,
                text=True,
                timeout=2
            )
            if result.returncode == 0:
                activity["recent_commits"] = result.stdout.strip().split('\n')
        except Exception:
            pass

        # 최근 로그 분석
        try:
            logs_dir = Path(".moai/logs/sessions")
            if logs_dir.exists():
                log_files = sorted(logs_dir.glob("*.log"), reverse=True)[:1]
                if log_files:
                    with open(log_files[0], 'r', encoding='utf-8') as f:
                        log_content = f.read()
                        # 마지막 활동 감지
                        if "/alfred:" in log_content:
                            activity["alfred_activity"] = "최근 알프레드 명령어 사용 감지"
                        if "error" in log_content.lower():
                            activity["error_activity"] = "최근 오류 로그 감지"
        except Exception:
            pass

        return activity

    def _perform_context_analysis(self, basic_info: Dict) -> Dict[str, Any]:
        """상황 인지 기반 분석"""
        analysis = {}

        # 개발 환경 평가
        analysis["development_environment"] = self._evaluate_development_environment(basic_info)

        # 작업 진행 상태 평가
        analysis["work_progress"] = self._evaluate_work_progress(basic_info)

        # 팀 협업 상태 평가
        analysis["collaboration_status"] = self._evaluate_collaboration_status(basic_info)

        # 위험 요소 평가
        analysis["risk_factors"] = self._identify_risk_factors(basic_info)

        return analysis

    def _evaluate_development_environment(self, basic_info: Dict) -> Dict[str, Any]:
        """개발 환경 평가"""
        evaluation = {"status": "stable", "issues": [], "recommendations": []}

        # 버전 호환성 확인
        moai_version = basic_info.get("moai_version", "")
        if moai_version != "unknown" and moai_version < "0.22.0":
            evaluation["status"] = "warning"
            evaluation["issues"].append("사용 중인 MoAI-ADK 버전이 최신이 아님")
            evaluation["recommendations"].append("최신 버전으로 업그레이드 권장")

        # Python 버전 확인
        python_version = basic_info.get("python_version", "")
        if python_version:
            major, minor = map(int, python_version.split('.')[:2])
            if major < 3 or (major == 3 and minor < 9):
                evaluation["status"] = "warning"
                evaluation["issues"].append("Python 버전이 최적화된 버전이 아님")
                evaluation["recommendations"].append("Python 3.9+ 사용 권장")

        return evaluation

    def _evaluate_work_progress(self, basic_info: Dict) -> Dict[str, Any]:
        """작업 진행 상태 평가"""
        evaluation = {"status": "normal", "work_items": [], "recommendations": []}

        uncommitted_count = basic_info.get("uncommitted_changes", 0)
        if uncommitted_count > 20:
            evaluation["status"] = "high"
            evaluation["recommendations"].append(f"미커밋 변경사항이 많습니다 ({uncommitted_count}개). 정기 커밋 권장")
        elif uncommitted_count > 5:
            evaluation["status"] = "medium"
            evaluation["recommendations"].append("변경사항을 정리할 시점입니다")

        current_spec = basic_info.get("current_spec", "")
        if current_spec:
            evaluation["work_items"].append(f"SPEC 작업 진행 중: {current_spec}")

        return evaluation

    def _evaluate_collaboration_status(self, basic_info: Dict) -> Dict[str, Any]:
        """팀 협업 상태 평가"""
        evaluation = {"status": "normal", "collaboration_indicators": []}

        project_mode = basic_info.get("project_mode", "")
        if project_mode == "team":
            evaluation["collaboration_indicators"].append("팀 모드로 작업 중")
            evaluation["recommendations"] = ["정기적인 코드 리뷰 권장", "커밋 메시지 표준화"]
        else:
            evaluation["collaboration_indicators"].append("개인 모드로 작업 중")

        # 브랜치 상태 확인
        current_branch = basic_info.get("current_branch", "")
        if current_branch.startswith("feature/"):
            evaluation["collaboration_indicators"].append("기능 개발 브랜치 사용 중")

        return evaluation

    def _identify_risk_factors(self, basic_info: Dict) -> List[str]:
        """위험 요소 식별"""
        risks = []

        uncommitted_count = basic_info.get("uncommitted_changes", 0)
        if uncommitted_count > 50:
            risks.append(f"미커밋 변경사항 과다 ({uncommitted_count}개)")

        # 긴 브랜치 이름
        current_branch = basic_info.get("current_branch", "")
        if len(current_branch) > 50:
            risks.append("브랜치 이름이 너무 깁니다")

        # 최근 오류 활동
        recent_activity = basic_info.get("recent_activity", {})
        if "error_activity" in recent_activity:
            risks.append("최근 오류 활동 감지")

        return risks

    def _analyze_patterns(self) -> Dict[str, Any]:
        """패턴 분석"""
        patterns = {}

        # 문서 패턴 분석
        patterns["documentation"] = self._analyze_documentation_patterns()

        # 코드 패턴 분석
        patterns["code"] = self._analyze_code_patterns()

        # 사용자 행동 패턴 분석
        patterns["behavior"] = self._analyze_behavior_patterns()

        return patterns

    def _analyze_documentation_patterns(self) -> Dict[str, Any]:
        """문서 패턴 분석"""
        doc_patterns = {"status": "good", "issues": []}

        try:
            result = subprocess.run([
                "python3", ".moai/scripts/validation/validate_claude_md_compliance.py"
            ], capture_output=True, text=True, timeout=10)

            if result.returncode != 0:
                doc_patterns["status"] = "needs_improvement"
                doc_patterns["issues"].append("CLAUDE.md 문서 준수성 문제")
        except Exception:
            pass

        return doc_patterns

    def _analyze_code_patterns(self) -> Dict[str, Any]:
        """코드 패턴 분석"""
        code_patterns = {"status": "good", "issues": []}

        try:
            result = subprocess.run([
                "python3", ".moai/scripts/validation/validate_all_skills.py"
            ], capture_output=True, text=True, timeout=10)

            if result.returncode != 0:
                code_patterns["status"] = "needs_review"
                code_patterns["issues"].append("Skill 표준 준수성 문제")
        except Exception:
            pass

        return code_patterns

    def _analyze_behavior_patterns(self) -> Dict[str, Any]:
        """사용자 행동 패턴 분석"""
        behavior_patterns = {"usage_frequency": "unknown", "preferences": []}

        # Git 커밋 빈도 분석
        try:
            result = subprocess.run([
                "git", "log", "--oneline", "--since", "1.week.ago"
            ], capture_output=True, text=True, timeout=5)

            if result.stdout:
                commits = result.stdout.strip().split('\n')
                if len(commits) > 15:
                    behavior_patterns["usage_frequency"] = "high"
                elif len(commits) > 5:
                    behavior_patterns["usage_frequency"] = "medium"
                else:
                    behavior_patterns["usage_frequency"] = "low"
        except Exception:
            pass

        return behavior_patterns

    def _generate_suggestions(self, basic_info: Dict, context_analysis: Dict,
                              pattern_analysis: Dict) -> List[str]:
        """개선 제안 생성"""
        suggestions = []

        # 환경 기반 제안
        dev_env = context_analysis.get("development_environment", {})
        if dev_env["status"] == "warning":
            suggestions.extend(dev_env.get("recommendations", []))

        # 작업 진행 상태 기반 제안
        work_progress = context_analysis.get("work_progress", {})
        suggestions.extend(work_progress.get("recommendations", []))

        # 패턴 분석 기반 제안
        doc_patterns = pattern_analysis.get("documentation", {})
        if doc_patterns["status"] != "good":
            suggestions.append("📚 문서 표준화 작업을 수행하세요")

        code_patterns = pattern_analysis.get("code", {})
        if code_patterns["status"] != "good":
            suggestions.append("🔧 Skill 표준 준수성 검증을 실행하세요")

        # 상황 기반 동적 제안
        uncommitted_count = basic_info.get("uncommitted_changes", 0)
        if uncommitted_count > 10:
            suggestions.append("💾 변경사항을 정리할 시점입니다 (새로운 브랜치 생성 고려)")

        current_spec = basic_info.get("current_spec", "")
        if current_spec:
            suggestions.append(f"🎯 {current_spec} SPEC 완료 목표를 설정하세요")

        # 중복 제안 제거
        unique_suggestions = list(dict.fromkeys(suggestions))

        return unique_suggestions[:10]  # 최대 10개 제안

    def _create_dynamic_template(self, basic_info: Dict, context_analysis: Dict,
                                 suggestions: List[str]) -> Dict[str, Any]:
        """동적 템플릿 생성"""
        template = {
            "type": "dynamic_issue_template",
            "context_based": True,
            "adapted_to": basic_info.get("current_branch", "unknown"),
            "sections": []
        }

        # 기본 정보 섹션
        template["sections"].append({
            "title": "🔍 환경 정보",
            "content": self._generate_environment_section(basic_info)
        })

        # 상황 분석 섹션
        template["sections"].append({
            "title": "📋 상황 분석",
            "content": self._generate_context_section(context_analysis)
        })

        # 제안 섹션
        if suggestions:
            template["sections"].append({
                "title": "💡 개선 제안",
                "content": "\n".join(f"- {suggestion}" for suggestion in suggestions)
            })

        # 이슈 유형별 동적 조정
        current_spec = basic_info.get("current_spec", "")
        if current_spec:
            template["sections"].append({
                "title": "🎯 SPEC 작업 연결",
                "content": f"이 이슈는 {current_spec} SPEC와 관련될 수 있습니다."
            })

        return template

    def _generate_environment_section(self, basic_info: Dict) -> str:
        """환경 정보 섹션 생성"""
        lines = []

        lines.append(f"**MoAI-ADK 버전**: {basic_info.get('moai_version', 'unknown')}")
        lines.append(f"**프로젝트 모드**: {basic_info.get('project_mode', 'unknown')}")
        lines.append(f"**현재 브랜치**: `{basic_info.get('current_branch', 'unknown')}`")

        uncommitted = basic_info.get('uncommitted_changes', 0)
        if uncommitted > 0:
            lines.append(f"**미커밋 변경사항**: {uncommitted}개")

        current_spec = basic_info.get('current_spec', '')
        if current_spec:
            lines.append(f"**작업 중 SPEC**: {current_spec}")

        return "\n".join(lines)

    def _generate_context_section(self, context_analysis: Dict) -> str:
        """상황 분석 섹션 생성"""
        lines = []

        # 개발 환경 평가
        dev_env = context_analysis.get("development_environment", {})
        lines.append(f"**개발 환경**: {dev_env.get('status', 'unknown')}")

        # 작업 진행 상태
        work_progress = context_analysis.get("work_progress", {})
        lines.append(f"**작업 진행**: {work_progress.get('status', 'unknown')}")

        # 위험 요소
        risks = context_analysis.get("risk_factors", [])
        if risks:
            lines.append("**위험 요소**:")
            for risk in risks:
                lines.append(f"- {risk}")

        return "\n".join(lines)

    def _generate_analysis_summary(self, basic_info: Dict, context_analysis: Dict,
                                  pattern_analysis: Dict, suggestions: List[str]) -> Dict[str, Any]:
        """분석 요약 생성"""
        summary = {
            "overall_status": "good",
            "priority_areas": [],
            "immediate_actions": [],
            "long_term_goals": []
        }

        # 종합 상태 평가
        risk_factors = context_analysis.get("risk_factors", [])
        if risk_factors:
            summary["overall_status"] = "needs_attention"
            summary["priority_areas"] = risk_factors

        # 즉시 필요한 행동
        if len(suggestions) > 0:
            summary["immediate_actions"] = suggestions[:3]

        # 장기 목표
        if pattern_analysis.get("code", {}).get("status") == "needs_review":
            summary["long_term_goals"].append("Skill 표준화")
        if pattern_analysis.get("documentation", {}).get("status") == "needs_improvement":
            summary["long_term_goals"].append("문서 품질 향상")

        return summary

    def format_output(self, data: Dict, format_type: str = "korean") -> str:
        """출력 형식화"""
        if format_type == "json":
            return json.dumps(data, ensure_ascii=False, indent=2)

        # 한국어 형식
        output = []
        output.append("🎯 향상된 피드백 수집 및 분석 보고서")
        output.append("=" * 50)
        output.append(f"생성 시간: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}")
        output.append("")

        # 분석 요약
        summary = data.get("analysis_summary", {})
        output.append("📊 분석 요약")
        output.append("-" * 30)
        output.append(f"전체 상태: {summary.get('overall_status', 'unknown')}")

        priority_areas = summary.get("priority_areas", [])
        if priority_areas:
            output.append("우선순위 영역:")
            for area in priority_areas:
                output.append(f"  • {area}")

        # 개선 제안
        suggestions = data.get("suggestions", [])
        if suggestions:
            output.append("")
            output.append("💡 개선 제안")
            output.append("-" * 30)
            for suggestion in suggestions:
                output.append(suggestion)

        # 동적 템플릿 표시
        template = data.get("dynamic_template", {})
        if template:
            output.append("")
            output.append("📋 동적 템플릿")
            output.append("-" * 30)
            for section in template.get("sections", []):
                output.append(f"**{section['title']}**")
                output.append(section['content'])
                output.append("")

        return "\n".join(output)

    def save_feedback_data(self, data: Dict, filename: str = None):
        """피드백 데이터 저장"""
        if not filename:
            timestamp = datetime.now().strftime('%Y%m%d_%H%M%S')
            filename = f"enhanced_feedback_{timestamp}.json"

        feedback_dir = Path(".moai/feedback")
        feedback_dir.mkdir(parents=True, exist_ok=True)

        filepath = feedback_dir / filename
        with open(filepath, 'w', encoding='utf-8') as f:
            json.dump(data, f, ensure_ascii=False, indent=2)

        print(f"📄 향상된 피드백 데이터가 저장되었습니다: {filepath}")


def main():
    parser = argparse.ArgumentParser(description="향상된 자동 피드백 수집 및 분석 시스템")
    parser.add_argument("--config", "-c", help="설정 파일 경로")
    parser.add_argument("--output", "-o", help="출력 파일 경로")
    parser.add_argument("--format", "-f", choices=["korean", "json"], default="korean", help="출력 형식")
    parser.add_argument("--save", "-s", action="store_true", help="피드백 데이터 저장")
    parser.add_argument("--quick", "-q", action="store_true", help="빠른 모드 (기본 분석만)")

    args = parser.parse_args()

    try:
        collector = EnhancedFeedbackCollector(args.config)

        # 종합 피드백 수집
        feedback_data = collector.collect_comprehensive_feedback()

        # 출력 생성
        output = collector.format_output(feedback_data, args.format)
        print(output)

        # 데이터 저장
        if args.save:
            collector.save_feedback_data(feedback_data, args.output)

    except Exception as e:
        print(f"❌ 오류: {str(e)}", file=sys.stderr)
        sys.exit(1)


if __name__ == "__main__":
    main()