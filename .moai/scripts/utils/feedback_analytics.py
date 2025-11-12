#!/usr/bin/env python3
"""
종합 사용자 피드백 분석 및 관리 시스템

기존 피드백 수집 시스템을 확장하여, 이슈 생성 전 분석, 패턴 감지, 개선 제안 기능을 제공합니다.

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
from collections import Counter, defaultdict


class FeedbackAnalyticsSystem:
    """종합 피드백 분석 및 관리 시스템"""

    def __init__(self, config_file: str = None):
        self.config = self._load_config(config_file)
        self.feedback_history = []
        self.patterns = {}

    def _load_config(self, config_file: str) -> Dict:
        """설정 파일 로드"""
        default_config = {
            "feedback_file": ".moai/feedback/feedback_history.json",
            "issue_categories": ["bug", "feature", "improvement", "refactor", "docs", "question"],
            "priority_mapping": {
                "긴급": "critical",
                "높음": "high",
                "중간": "medium",
                "낮음": "low"
            },
            "auto_analysis": True,
            "pattern_detection": True,
            "suggestion_engine": True
        }

        if config_file and Path(config_file).exists():
            import yaml
            with open(config_file, 'r', encoding='utf-8') as f:
                return {**default_config, **yaml.safe_load(f)}
        return default_config

    def collect_current_feedback_context(self) -> Dict[str, Any]:
        """현재 환경에서의 피드백 컨텍스트 수집"""
        base_info = self._collect_basic_info()
        enhanced_context = self._collect_enhanced_context()

        return {
            "timestamp": datetime.now().isoformat(),
            "basic_info": base_info,
            "enhanced_context": enhanced_context,
            "analysis_summary": self._generate_analysis_summary(base_info, enhanced_context)
        }

    def _collect_basic_info(self) -> Dict[str, Any]:
        """기본 환경 정보 수집 (기존 스크립트 확장)"""
        try:
            result = subprocess.run(
                ["python3", ".moai/scripts/utils/feedback-collect-info.py", "--json"],
                capture_output=True,
                text=True,
                timeout=5
            )
            if result.returncode == 0:
                return json.loads(result.stdout)
        except Exception:
            pass

        return {
            "moai_version": "unknown",
            "python_version": "unknown",
            "os_info": "unknown",
            "project_mode": "unknown",
            "current_branch": "unknown",
            "uncommitted_changes": 0,
            "current_spec": "",
            "recent_git_commits": ""
        }

    def _collect_enhanced_context(self) -> Dict[str, Any]:
        """향상된 컨텍스트 정보 수집"""
        context = {}

        # 현재 작업 상태
        context["work_in_progress"] = self._detect_current_work()

        # 최근 문제점 패턴
        context["recent_issues"] = self._analyze_recent_issues()

        # 개선 요구사항
        context["improvement_requests"] = self._analyze_improvement_patterns()

        # 사용자 활동 패턴
        context["user_activity"] = self._analyze_user_activity()

        return context

    def _detect_current_work(self) -> List[str]:
        """현재 진행 중인 작업 감지"""
        work_patterns = []

        # Git 브랜치 분석
        try:
            result = subprocess.run(
                ["git", "branch", "--show-current"],
                capture_output=True,
                text=True,
                timeout=2
            )
            branch = result.stdout.strip()
            if branch.startswith("feature/SPEC-"):
                work_patterns.append(f"SPEC 개작 중: {branch}")
            elif branch.startswith("hotfix/"):
                work_patterns.append(f"긴급 수정 중: {branch}")
            elif branch != "main" and branch != "develop":
                work_patterns.append(f"브랜치 작업 중: {branch}")
        except Exception:
            pass

        # 커밋 상태 분석
        try:
            result = subprocess.run(
                ["git", "diff", "--name-only"],
                capture_output=True,
                text=True,
                timeout=2
            )
            if result.stdout.strip():
                modified_files = result.stdout.strip().split('\n')
                if len(modified_files) > 5:
                    work_patterns.append(f"다중 파일 수정 중: {len(modified_files)}개 파일")
                else:
                    work_patterns.append("소규모 파일 수정 중")
        except Exception:
            pass

        return work_patterns

    def _analyze_recent_issues(self) -> List[Dict[str, Any]]:
        """최근 이슈 패턴 분석"""
        issues_data = []

        # GitHub 이슈 조회 (최근 10개)
        try:
            result = subprocess.run([
                "gh", "issue", "list",
                "--limit", "10",
                "--state", "all",
                "--json", "title,number,state,labels,createdAt"
            ], capture_output=True, text=True, timeout=10)

            if result.returncode == 0:
                issues = json.loads(result.stdout)
                for issue in issues:
                    issues_data.append({
                        "title": issue.get("title", ""),
                        "number": issue.get("number", 0),
                        "state": issue.get("state", ""),
                        "labels": issue.get("labels", []),
                        "created_at": issue.get("createdAt", "")
                    })
        except Exception:
            pass

        return issues_data

    def _analyze_improvement_patterns(self) -> List[str]:
        """개선 요구사항 패턴 분석"""
        patterns = []

        # CLAUDE.md 검증 결과 분석
        try:
            result = subprocess.run([
                "python3", ".moai/scripts/validation/validate_claude_md_compliance.py",
                "--file", "CLAUDE.md"
            ], capture_output=True, text=True, timeout=10)

            if result.returncode != 0:
                # 검증에서 문제가 발견되면 개선 패턴으로 분류
                patterns.append("문서 준수성 개선 필요")
        except Exception:
            pass

        # Skill 검증 결과 분석
        try:
            result = subprocess.run([
                "python3", ".moai/scripts/validation/validate_all_skills.py"
            ], capture_output=True, text=True, timeout=10)

            if result.returncode != 0:
                patterns.append("Skill 표준 준수성 개선 필요")
        except Exception:
            pass

        return patterns

    def _analyze_user_activity(self) -> Dict[str, Any]:
        """사용자 활동 패턴 분석"""
        activity = {
            "command_usage": self._analyze_command_patterns(),
            "session_duration": self._estimate_session_duration(),
            "interaction_frequency": self._estimate_interaction_frequency()
        }
        return activity

    def _analyze_command_patterns(self) -> List[str]:
        """명령어 사용 패턴 분석"""
        patterns = []

        # 최근 커밋 메시지 분석
        try:
            result = subprocess.run([
                "git", "log", "--oneline", "-10", "--grep", "/alfred:"
            ], capture_output=True, text=True, timeout=5)

            if result.stdout:
                alfred_commands = result.stdout.strip().split('\n')
                if len(alfred_commands) > 5:
                    patterns.append("알프레드 명령어 자주 사용")
                else:
                    patterns.append("알프레드 명령어 주기적 사용")
        except Exception:
            pass

        return patterns

    def _estimate_session_duration(self) -> str:
        """세션 지속 시간 추정"""
        try:
            logs_dir = Path(".moai/logs/sessions")
            if logs_dir.exists():
                log_files = sorted(logs_dir.glob("*.log"), reverse=True)
                if log_files:
                    # 가장 최근 로그 파일의 마지막 수정 시간으로 추정
                    latest_log = log_files[0]
                    time_diff = datetime.now() - datetime.fromtimestamp(latest_log.stat().st_mtime)

                    if time_diff < timedelta(minutes=30):
                        return "짧은 세션 (< 30분)"
                    elif time_diff < timedelta(hours=2):
                        return "중간 세션 (30분 - 2시간)"
                    else:
                        return "긴 세션 (> 2시간)"
        except Exception:
            pass

        return "알 수 없음"

    def _estimate_interaction_frequency(self) -> str:
        """상호작용 빈도 추정"""
        try:
            # Git 커밋 빈도로 추정
            result = subprocess.run([
                "git", "log", "--oneline", "--since", "1.week.ago"
            ], capture_output=True, text=True, timeout=5)

            if result.stdout:
                commits_this_week = len(result.stdout.strip().split('\n'))
                if commits_this_week > 10:
                    return "고빈도 (주 10+ 커밋)"
                elif commits_this_week > 3:
                    return "중빈도 (주 3-10 커밋)"
                else:
                    return "저빈도 (주 < 3 커밋)"
        except Exception:
            pass

        return "알 수 없음"

    def _generate_analysis_summary(self, basic_info: Dict, enhanced_context: Dict) -> Dict[str, Any]:
        """분석 요약 생성"""
        summary = {
            "overall_context": "안정적인 개발 환경",
            "priority_signals": [],
            "recommendations": [],
            "risk_factors": []
        }

        # 우선순위 신호 분석
        if basic_info.get("uncommitted_changes", 0) > 10:
            summary["priority_signals"].append("다수의 미커밋 변경사항")
            summary["risk_factors"].append("변경사항 누적")

        if enhanced_context.get("improvement_requests"):
            summary["priority_signals"].append("개선 요구사항 존재")
            summary["recommendations"].append("개선 사항 우선순위화")

        # 활동 패턴 기반 추천
        activity = enhanced_context.get("user_activity", {})
        if activity.get("session_duration") == "긴 세션 (> 2시간)":
            summary["recommendations"].append("세션 분할 제안")

        return summary

    def analyze_feedback_patterns(self, time_period: str = "1month") -> Dict[str, Any]:
        """피드백 패턴 분석"""
        patterns = {
            "issue_type_distribution": self._analyze_issue_type_distribution(time_period),
            "priority_trends": self._analyze_priority_trends(time_period),
            "common_problems": self._identify_common_problems(time_period),
            "improvement_opportunities": self._identify_improvement_opportunities(time_period)
        }

        self.patterns = patterns
        return patterns

    def _analyze_issue_type_distribution(self, time_period: str) -> Dict[str, int]:
        """이슈 타입 분포 분석"""
        # TODO: GitHub API를 통한 실제 이슈 분석 구현
        return {
            "bug": 15,
            "feature": 8,
            "improvement": 12,
            "docs": 5,
            "question": 3
        }

    def _analyze_priority_trends(self, time_period: str) -> Dict[str, int]:
        """우선순위 트렌드 분석"""
        # TODO: 실제 우선순위 분석 구현
        return {
            "critical": 2,
            "high": 5,
            "medium": 20,
            "low": 16
        }

    def _identify_common_problems(self, time_period: str) -> List[str]:
        """공통 문제점 식별"""
        common_issues = []

        # 검증 스크립트 오류 분석
        try:
            result = subprocess.run([
                "python3", ".moai/scripts/validation/validate_claude_md_compliance.py"
            ], capture_output=True, text=True, timeout=10)

            if result.stderr:
                # 오류 메시지에서 패턴 추출
                error_patterns = re.findall(r"Missing|Error|Failed", result.stderr)
                if error_patterns:
                    common_issues.append("문서 준수성 문제")
        except Exception:
            pass

        return common_issues

    def _identify_improvement_opportunities(self, time_period: str) -> List[str]:
        """개선 기회 식별"""
        opportunities = []

        # Skill 검증 결과 기반 개선 기회
        try:
            result = subprocess.run([
                "python3", ".moai/scripts/validation/validate_all_skills.py", "--detailed"
            ], capture_output=True, text=True, timeout=10)

            if result.returncode != 0:
                opportunities.append("Skill 메타데이터 개선")
        except Exception:
            pass

        return opportunities

    def generate_intelligent_suggestions(self) -> List[str]:
        """AI 기반 개선 제안 생성"""
        suggestions = []

        # 현재 상태 기반 제안
        context = self.collect_current_feedback_context()

        # 문서 개선 제안
        if "문서 준수성 문제" in context.get("enhanced_context", {}).get("recent_issues", []):
            suggestions.append("📚 CLAUDE.md 문서를 공식 표준에 맞게 개선하세요")

        # Skill 개선 제안
        if len(context.get("enhanced_context", {}).get("improvement_requests", [])) > 0:
            suggestions.append("🔧 Skill 표준 준수성 검증을 실행하세요")

        # 사용 경험 개선 제안
        if context.get("analysis_summary", {}).get("session_duration") == "긴 세션 (> 2시간)":
            suggestions.append("⏱️ 세션을 분할하여 작업 효율을 높이세요")

        # 우선순위 기반 제안
        priority_signals = context.get("analysis_summary", {}).get("priority_signals", [])
        if "다수의 미커밋 변경사항" in priority_signals:
            suggestions.append("💾 변경사항을 정기적으로 커밋하세요")

        return suggestions

    def generate_comprehensive_report(self) -> str:
        """종합 분석 보고서 생성"""
        context = self.collect_current_feedback_context()
        patterns = self.analyze_feedback_patterns()
        suggestions = self.generate_intelligent_suggestions()

        report = []
        report.append("📊 종합 피드백 분석 보고서")
        report.append("=" * 50)
        report.append(f"생성 시간: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}")
        report.append("")

        # 현재 상태 요약
        report.append("🔍 현재 개발 환경 상태")
        report.append("-" * 30)

        basic_info = context.get("basic_info", {})
        report.append(f"MoAI-ADK 버전: {basic_info.get('moai_version', 'unknown')}")
        report.append(f"현재 브랜치: `{basic_info.get('current_branch', 'unknown')}`")
        report.append(f"미커밋 변경사항: {basic_info.get('uncommitted_changes', 0)}개")
        report.append(f"세션 지속 시간: {context.get('analysis_summary', {}).get('session_duration', '알 수 없음')}")
        report.append("")

        # 작업 진행 상태
        work_items = context.get("enhanced_context", {}).get("work_in_progress", [])
        if work_items:
            report.append("📋 현재 진행 중인 작업")
            report.append("-" * 30)
            for item in work_items:
                report.append(f"• {item}")
            report.append("")

        # 분석 결과
        report.append("📈 분석 결과")
        report.append("-" * 30)

        # 이슈 패턴
        recent_issues = context.get("enhanced_context", {}).get("recent_issues", [])
        if recent_issues:
            report.append("최근 이슈:")
            for issue in recent_issues[:3]:  # 최근 3개만 표시
                report.append(f"  - #{issue['number']}: {issue['title']} ({issue['state']})")
        report.append("")

        # 개선 요구사항
        improvements = context.get("enhanced_context", {}).get("improvement_requests", [])
        if improvements:
            report.append("개선 요구사항:")
            for improvement in improvements:
                report.append(f"  - {improvement}")

        # 추천 제안
        if suggestions:
            report.append("")
            report.append("💡 AI 기반 추천 제안")
            report.append("-" * 30)
            for suggestion in suggestions:
                report.append(suggestion)

        return "\n".join(report)

    def save_feedback_data(self, feedback_data: Dict, filename: str = None):
        """피드백 데이터 저장"""
        if not filename:
            filename = f"feedback_analysis_{datetime.now().strftime('%Y%m%d_%H%M%S')}.json"

        feedback_file = Path(self.config["feedback_file"])
        feedback_file.parent.mkdir(parents=True, exist_ok=True)

        # 기존 데이터 로드
        if feedback_file.exists():
            try:
                with open(feedback_file, 'r', encoding='utf-8') as f:
                    existing_data = json.load(f)
                    if isinstance(existing_data, list):
                        self.feedback_history = existing_data
            except Exception:
                self.feedback_history = []

        # 새 데이터 추가
        self.feedback_history.append(feedback_data)

        # 저장
        with open(feedback_file, 'w', encoding='utf-8') as f:
            json.dump(self.feedback_history, f, ensure_ascii=False, indent=2)

        print(f"📄 피드백 데이터가 저장되었습니다: {feedback_file}")


def main():
    parser = argparse.ArgumentParser(description="종합 피드백 분석 시스템")
    parser.add_argument("--config", "-c", help="설정 파일 경로")
    parser.add_argument("--report", "-r", help="보고서 파일 경로")
    parser.add_argument("--save-feedback", "-s", action="store_true", help="피드백 데이터 저장")
    parser.add_argument("--analyze-patterns", "-p", action="store_true", help="패턴 분석 실행")
    parser.add_argument("--suggestions", action="store_true", help="개선 제안만 표시")

    args = parser.parse_args()

    try:
        analytics = FeedbackAnalyticsSystem(args.config)

        if args.suggestions:
            suggestions = analytics.generate_intelligent_suggestions()
            for suggestion in suggestions:
                print(suggestion)
            return

        # 현재 컨텍스트 수집
        current_context = analytics.collect_current_feedback_context()

        print("🔍 피드백 컨텍스트 분석 중...")
        print(analytics.generate_comprehensive_report())

        if args.analyze_patterns:
            print("\n📈 패턴 분석 중...")
            patterns = analytics.analyze_feedback_patterns()
            print(json.dumps(patterns, ensure_ascii=False, indent=2))

        if args.save_feedback:
            print("\n💾 피드백 데이터 저장 중...")
            analytics.save_feedback_data(current_context)

        if args.report:
            report = analytics.generate_comprehensive_report()
            with open(args.report, 'w', encoding='utf-8') as f:
                f.write(report)
            print(f"📄 보고서가 저장되었습니다: {args.report}")

    except Exception as e:
        print(f"❌ 오류: {str(e)}", file=sys.stderr)
        sys.exit(1)


if __name__ == "__main__":
    main()