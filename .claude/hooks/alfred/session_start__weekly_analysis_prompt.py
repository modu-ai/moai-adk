#!/usr/bin/env python3
"""
SessionStart Hook: Weekly Session Analysis Reminder
세션 시작 시 주간 분석 리마인더 표시

역할:
- 마지막 분석 이후 경과 일수 확인
- 7일 이상 경과했으면 사용자에게 안내
- 수동 실행 명령어 제공
"""

import sys
import json
from pathlib import Path
from datetime import datetime, timedelta


def get_last_analysis_date():
    """마지막 분석 보고서의 날짜 가져오기"""
    reports_dir = Path.cwd() / ".moai" / "reports"

    if not reports_dir.exists():
        return None

    # weekly-YYYY-MM-DD.md 형식 파일 찾기
    weekly_reports = sorted(reports_dir.glob("weekly-*.md"))

    if not weekly_reports:
        return None

    # 가장 최신 리포트의 날짜 추출
    latest_report = weekly_reports[-1]
    # 파일명: weekly-2024-11-04.md → 2024-11-04
    try:
        date_str = latest_report.stem.split("-", 1)[1]
        return datetime.strptime(date_str, "%Y-%m-%d")
    except (ValueError, IndexError):
        return None


def days_since_last_analysis():
    """마지막 분석 이후 경과 일수"""
    last_date = get_last_analysis_date()

    if last_date is None:
        return 999  # 첫 실행

    return (datetime.now() - last_date).days


def main():
    """SessionStart 훅 메인 함수"""

    days_elapsed = days_since_last_analysis()

    # 7일 이상 경과했으면 리마인더 표시
    if days_elapsed >= 7:
        print("\n" + "=" * 70)
        print("📊 Weekly Session Analysis Reminder")
        print("=" * 70)
        print()

        if days_elapsed == 999:
            print("🔍 첫 세션 분석입니다. 한 주일 사용 후 분석을 시작할 수 있습니다.")
        else:
            print(f"⏰ 마지막 분석: {days_elapsed}일 전")
            print()
            print("💡 세션 분석 실행:")
            print()
            print("   Option 1: Python 스크립트로 실행")
            print("   python3 .moai/scripts/session_analyzer.py --days 7 --verbose")
            print()
            print("   Option 2: 셸 스크립트로 실행")
            print("   bash .moai/scripts/weekly_analysis.sh")
            print()
            print("분석 결과는 .moai/reports/weekly-YYYY-MM-DD.md에 저장됩니다.")

        print()
        print("=" * 70)
        print()


if __name__ == "__main__":
    try:
        main()
    except Exception as e:
        # Hook 오류는 조용히 처리 (세션 시작을 방해하지 않음)
        if "--verbose" in sys.argv:
            print(f"⚠️ Hook error: {e}", file=sys.stderr)
