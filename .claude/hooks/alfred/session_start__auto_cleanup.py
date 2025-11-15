#!/usr/bin/env python3

"""SessionStart Hook: 자동 정리 및 보고서 생성

세션 시작 시 오래된 임시 파일, 보고서 등을 정리하고
일일 분석 보고서를 생성합니다.

기능:
- 오래된 보고서 파일 자동 정리
- 임시 파일 정리
- 일일 분석 보고서 생성
- 세션 로그 분석
"""

import json
import logging
import shutil
import sys
from datetime import datetime, timedelta
from pathlib import Path
from typing import Dict, List, Optional

# 모듈 경로 추가
sys.path.insert(0, str(Path(__file__).parent.parent.parent.parent / "src"))

try:
    from moai_adk.utils.common import format_duration, get_summary_stats
except ImportError:
    # Fallback implementations if module not found
    def format_duration(seconds):
        """Format duration in seconds to readable string"""
        if seconds < 60:
            return f"{seconds:.1f}s"
        minutes = seconds / 60
        if minutes < 60:
            return f"{minutes:.1f}m"
        hours = minutes / 60
        return f"{hours:.1f}h"

    def get_summary_stats(values):
        """Get summary statistics for a list of values"""
        if not values:
            return {"mean": 0, "min": 0, "max": 0, "std": 0}
        import statistics
        return {
            "mean": statistics.mean(values),
            "min": min(values),
            "max": max(values),
            "std": statistics.stdev(values) if len(values) > 1 else 0
        }

logger = logging.getLogger(__name__)


def load_hook_timeout() -> int:
    """Load hook timeout from config.json (default: 3000ms)"""
    try:
        config_file = Path(".moai/config/config.json")
        if config_file.exists():
            with open(config_file, 'r', encoding='utf-8') as f:
                config = json.load(f)
                return config.get("hooks", {}).get("timeout_ms", 3000)
    except Exception:
        pass
    return 3000


def get_graceful_degradation() -> bool:
    """Load graceful_degradation setting from config.json (default: true)"""
    try:
        config_file = Path(".moai/config/config.json")
        if config_file.exists():
            with open(config_file, 'r', encoding='utf-8') as f:
                config = json.load(f)
                return config.get("hooks", {}).get("graceful_degradation", True)
    except Exception:
        pass
    return True


def load_config() -> Dict:
    """설정 파일 로드"""
    try:
        config_file = Path(".moai/config/config.json")
        if config_file.exists():
            with open(config_file, 'r', encoding='utf-8') as f:
                return json.load(f)
    except Exception:
        pass

    return {}


def should_cleanup_today(last_cleanup: Optional[str], cleanup_days: int = 7) -> bool:
    """오늘 정리가 필요한지 확인

    Args:
        last_cleanup: 마지막 정리 날짜 (YYYY-MM-DD)
        cleanup_days: 정리 주기 (일)

    Returns:
        정리 필요 여부
    """
    if not last_cleanup:
        return True

    try:
        last_date = datetime.strptime(last_cleanup, "%Y-%m-%d")
        next_cleanup = last_date + timedelta(days=cleanup_days)
        return datetime.now() >= next_cleanup
    except Exception:
        return True


def cleanup_old_files(config: Dict) -> Dict[str, int]:
    """오래된 파일 정리

    Args:
        config: 설정 딕셔너리

    Returns:
        정리된 파일 수 통계
    """
    stats = {
        "reports_cleaned": 0,
        "cache_cleaned": 0,
        "temp_cleaned": 0,
        "total_cleaned": 0
    }

    try:
        cleanup_config = config.get("auto_cleanup", {})
        if not cleanup_config.get("enabled", True):
            return stats

        cleanup_days = cleanup_config.get("cleanup_days", 7)
        max_reports = cleanup_config.get("max_reports", 10)
        _cleanup_targets = cleanup_config.get("cleanup_targets", [])

        cutoff_date = datetime.now() - timedelta(days=cleanup_days)

        # 보고서 파일 정리
        reports_dir = Path(".moai/reports")
        if reports_dir.exists():
            stats["reports_cleaned"] = cleanup_directory(
                reports_dir,
                cutoff_date,
                max_reports,
                patterns=["*.json", "*.md"]
            )

        # 캐시 파일 정리
        cache_dir = Path(".moai/cache")
        if cache_dir.exists():
            stats["cache_cleaned"] = cleanup_directory(
                cache_dir,
                cutoff_date,
                None,  # 캐시는 개수 제한 없음
                patterns=["*"]
            )

        # 임시 파일 정리
        temp_dir = Path(".moai/temp")
        if temp_dir.exists():
            stats["temp_cleaned"] = cleanup_directory(
                temp_dir,
                cutoff_date,
                None,
                patterns=["*"]
            )

        stats["total_cleaned"] = (
            stats["reports_cleaned"] +
            stats["cache_cleaned"] +
            stats["temp_cleaned"]
        )

    except Exception as e:
        logger.error(f"File cleanup failed: {e}")

    return stats


def cleanup_directory(
    directory: Path,
    cutoff_date: datetime,
    max_files: Optional[int],
    patterns: List[str]
) -> int:
    """디렉토리 파일 정리

    Args:
        directory: 대상 디렉토리
        cutoff_date: 자를 기준 날짜
        max_files: 최대 유지 파일 수
        patterns: 삭제할 파일 패턴 목록

    Returns:
        삭제된 파일 수
    """
    if not directory.exists():
        return 0

    cleaned_count = 0

    try:
        # 패턴에 해당하는 파일 목록 수집
        files_to_check = []
        for pattern in patterns:
            files_to_check.extend(directory.glob(pattern))

        # 날짜순 정렬 (오래된 것부터)
        files_to_check.sort(key=lambda f: f.stat().st_mtime)

        # 파일 삭제
        for file_path in files_to_check:
            try:
                # 파일 수정 시간 확인
                file_mtime = datetime.fromtimestamp(file_path.stat().st_mtime)

                # 기준 날짜 이전이면 삭제
                if file_mtime < cutoff_date:
                    if file_path.is_file():
                        file_path.unlink()
                        cleaned_count += 1
                    elif file_path.is_dir():
                        shutil.rmtree(file_path)
                        cleaned_count += 1

                # 최대 파일 수 제한
                elif max_files is not None:
                    remaining_files = len([f for f in files_to_check
                                         if f.exists() and
                                         datetime.fromtimestamp(f.stat().st_mtime) >= cutoff_date])
                    if remaining_files > max_files:
                        if file_path.is_file():
                            file_path.unlink()
                            cleaned_count += 1
                        elif file_path.is_dir():
                            shutil.rmtree(file_path)
                            cleaned_count += 1

            except Exception as e:
                logger.warning(f"Failed to delete {file_path}: {e}")
                continue

    except Exception as e:
        logger.error(f"Directory cleanup failed for {directory}: {e}")

    return cleaned_count


def generate_daily_analysis(config: Dict) -> Optional[str]:
    """일일 분석 보고서 생성

    Args:
        config: 설정 딕셔너리

    Returns:
        생성된 보고서 파일 경로 또는 None
    """
    try:
        analysis_config = config.get("daily_analysis", {})
        if not analysis_config.get("enabled", True):
            return None

        # 세션 로그 분석
        report_path = analyze_session_logs(analysis_config)

        # 설정에 마지막 분석 날짜 업데이트
        if report_path:
            config_file = Path(".moai/config/config.json")
            if config_file.exists():
                with open(config_file, 'r', encoding='utf-8') as f:
                    config_data = json.load(f)

                config_data["daily_analysis"]["last_analysis"] = datetime.now().strftime("%Y-%m-%d")

                with open(config_file, 'w', encoding='utf-8') as f:
                    json.dump(config_data, f, indent=2, ensure_ascii=False)

        return report_path

    except Exception as e:
        logger.error(f"Daily analysis failed: {e}")
        return None


def analyze_session_logs(analysis_config: Dict) -> Optional[str]:
    """세션 로그 분석

    Args:
        analysis_config: 분석 설정

    Returns:
        보고서 파일 경로 또는 None
    """
    try:
        # Claude Code 세션 로그 경로
        session_logs_dir = Path.home() / ".claude" / "projects"
        project_name = Path.cwd().name

        # 현재 프로젝트의 세션 로그 찾기
        project_sessions = []
        for project_dir in session_logs_dir.iterdir():
            if project_dir.name.endswith(project_name):
                session_files = list(project_dir.glob("session-*.json"))
                project_sessions.extend(session_files)

        if not project_sessions:
            return None

        # 최근 세션 로그 분석
        recent_sessions = sorted(project_sessions, key=lambda f: f.stat().st_mtime, reverse=True)[:10]

        # 분석 데이터 수집
        analysis_data = {
            "total_sessions": len(recent_sessions),
            "date_range": "",
            "tools_used": {},
            "errors_found": [],
            "duration_stats": {},
            "recommendations": []
        }

        if recent_sessions:
            first_session = datetime.fromtimestamp(recent_sessions[-1].stat().st_mtime)
            last_session = datetime.fromtimestamp(recent_sessions[0].stat().st_mtime)
            analysis_data["date_range"] = f"{first_session.strftime('%Y-%m-%d')} ~ {last_session.strftime('%Y-%m-%d')}"

            # 각 세션 분석
            all_durations = []
            for session_file in recent_sessions:
                try:
                    with open(session_file, 'r', encoding='utf-8') as f:
                        session_data = json.load(f)

                    # 도구 사용 분석
                    if "tool_use" in session_data:
                        for tool_use in session_data["tool_use"]:
                            tool_name = tool_use.get("name", "unknown")
                            analysis_data["tools_used"][tool_name] = analysis_data["tools_used"].get(tool_name, 0) + 1

                    # 오류 분석
                    if "errors" in session_data:
                        for error in session_data["errors"]:
                            analysis_data["errors_found"].append({
                                "timestamp": error.get("timestamp", ""),
                                "error": error.get("message", "")[:100]  # 첫 100자만
                            })

                    # 세션 길이 분석
                    if "start_time" in session_data and "end_time" in session_data:
                        start = session_data["start_time"]
                        end = session_data["end_time"]
                        if start and end:
                            try:
                                duration = float(end) - float(start)
                                all_durations.append(duration)
                            except (ValueError, TypeError):
                                pass

                except Exception as e:
                    logger.warning(f"Failed to analyze session {session_file}: {e}")
                    continue

            # 세션 길이 통계
            if all_durations:
                analysis_data["duration_stats"] = get_summary_stats(all_durations)

        # 보고서 생성
        report_content = format_analysis_report(analysis_data)

        # 보고서 저장
        _report_location = analysis_config.get("report_location", ".moai/reports/daily-")
        base_path = Path(".moai/reports")
        base_path.mkdir(exist_ok=True)

        timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
        report_file = base_path / f"daily-analysis-{timestamp}.md"

        with open(report_file, 'w', encoding='utf-8') as f:
            f.write(report_content)

        return str(report_file)

    except Exception as e:
        logger.error(f"Session log analysis failed: {e}")
        return None


def format_analysis_report(analysis_data: Dict) -> str:
    """분석 결과를 보고서 형식으로 변환

    Args:
        analysis_data: 분석 데이터

    Returns:
        형식화된 보고서 내용
    """
    report_lines = [
        "# 일일 세션 분석 보고서",
        "",
        f"생성 시간: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}",
        f"분석 기간: {analysis_data.get('date_range', 'N/A')}",
        f"총 세션 수: {analysis_data.get('total_sessions', 0)}",
        "",
        "## 📊 도구 사용 현황",
        ""
    ]

    # 도구 사용 순위
    tools_used = analysis_data.get("tools_used", {})
    if tools_used:
        sorted_tools = sorted(tools_used.items(), key=lambda x: x[1], reverse=True)
        for tool_name, count in sorted_tools[:10]:  # TOP 10
            report_lines.append(f"- **{tool_name}**: {count}회")
    else:
        report_lines.append("- 사용된 도구가 없습니다")

    report_lines.extend([
        "",
        "## ⚠️ 오류 현황",
        ""
    ])

    # 오류 현황
    errors = analysis_data.get("errors_found", [])
    if errors:
        for i, error in enumerate(errors[:5], 1):  # 최근 5개
            report_lines.append(f"{i}. {error.get('error', 'N/A')} ({error.get('timestamp', 'N/A')})")
    else:
        report_lines.append("- 발견된 오류가 없습니다")

    # 세션 길이 통계
    duration_stats = analysis_data.get("duration_stats", {})
    if duration_stats.get("mean", 0) > 0:
        report_lines.extend([
            "",
            "## ⏱️ 세션 길이 통계",
            "",
            f"- 평균: {format_duration(duration_stats['mean'])}",
            f"- 최소: {format_duration(duration_stats['min'])}",
            f"- 최대: {format_duration(duration_stats['max'])}",
            f"- 표준편차: {format_duration(duration_stats['std'])}"
        ])

    # 개선 제안
    report_lines.extend([
        "",
        "## 💡 개선 제안",
        ""
    ])

    # 도구 사용 패턴 기반 제안
    if tools_used:
        most_used_tool = max(tools_used.items(), key=lambda x: x[1])[0]
        if "Bash" in most_used_tool and tools_used[most_used_tool] > 10:
            report_lines.append("- 🔧 Bash 명령어 사용이 잦습니다. 스크립트 자동화를 고려해보세요")

    if len(errors) > 3:
        report_lines.append("- ⚠️ 오류 발생이 잦습니다. 안정성 검토가 필요합니다")

    if duration_stats.get("mean", 0) > 1800:  # 30분 이상
        report_lines.append("- ⏰ 세션 시간이 깁니다. 작업 분할을 고려해보세요")

    if not report_lines[-1].startswith("-"):
        report_lines.append("- 현재 세션 패턴이 양호합니다")

    report_lines.extend([
        "",
        "---",
        "*보고서는 Alfred의 SessionStart Hook으로 자동 생성되었습니다*",
        "*분석 설정은 `.moai/config/config.json`의 `daily_analysis` 섹션에서 관리할 수 있습니다*"
    ])

    return "\n".join(report_lines)


def update_cleanup_stats(cleanup_stats: Dict[str, int]):
    """정리 통계 업데이트

    Args:
        cleanup_stats: 정리 통계
    """
    try:
        stats_file = Path(".moai/cache/cleanup_stats.json")
        stats_file.parent.mkdir(exist_ok=True)

        # 기존 통계 로드
        existing_stats = {}
        if stats_file.exists():
            with open(stats_file, 'r', encoding='utf-8') as f:
                existing_stats = json.load(f)

        # 새 통계 추가
        today = datetime.now().strftime("%Y-%m-%d")
        existing_stats[today] = {
            "cleaned_files": cleanup_stats["total_cleaned"],
            "reports_cleaned": cleanup_stats["reports_cleaned"],
            "cache_cleaned": cleanup_stats["cache_cleaned"],
            "temp_cleaned": cleanup_stats["temp_cleaned"],
            "timestamp": datetime.now().isoformat()
        }

        # 최근 30일 통계만 유지
        cutoff_date = datetime.now() - timedelta(days=30)
        filtered_stats = {}
        for date, stats in existing_stats.items():
            try:
                stat_date = datetime.strptime(date, "%Y-%m-%d")
                if stat_date >= cutoff_date:
                    filtered_stats[date] = stats
            except ValueError:
                continue

        # 통계 저장
        with open(stats_file, 'w', encoding='utf-8') as f:
            json.dump(filtered_stats, f, indent=2, ensure_ascii=False)

    except Exception as e:
        logger.error(f"Failed to update cleanup stats: {e}")


def main():
    """메인 함수"""
    try:
        # Hook timeout 설정 로드
        timeout_seconds = load_hook_timeout() / 1000
        graceful_degradation = get_graceful_degradation()

        # 타임아웃 체크
        import signal
        import time

        def timeout_handler(signum, frame):
            raise TimeoutError("Hook execution timeout")

        signal.signal(signal.SIGALRM, timeout_handler)
        signal.alarm(int(timeout_seconds))

        try:
            start_time = time.time()

            # 설정 로드
            config = load_config()

            # 마지막 정리 날짜 확인
            last_cleanup = config.get("auto_cleanup", {}).get("last_cleanup")
            cleanup_days = config.get("auto_cleanup", {}).get("cleanup_days", 7)

            cleanup_stats = {"total_cleaned": 0, "reports_cleaned": 0, "cache_cleaned": 0, "temp_cleaned": 0}
            report_path = None

            # 정리 필요 시 실행
            if should_cleanup_today(last_cleanup, cleanup_days):
                cleanup_stats = cleanup_old_files(config)

                # 마지막 정리 날짜 업데이트
                config_file = Path(".moai/config/config.json")
                if config_file.exists():
                    with open(config_file, 'r', encoding='utf-8') as f:
                        config_data = json.load(f)

                    config_data["auto_cleanup"]["last_cleanup"] = datetime.now().strftime("%Y-%m-%d")

                    with open(config_file, 'w', encoding='utf-8') as f:
                        json.dump(config_data, f, indent=2, ensure_ascii=False)

                # 정리 통계 업데이트
                update_cleanup_stats(cleanup_stats)

            # 일일 분석 보고서 생성
            last_analysis = config.get("daily_analysis", {}).get("last_analysis")
            if should_cleanup_today(last_analysis, 1):  # 매일 실행
                report_path = generate_daily_analysis(config)

            # 실행 시간 기록
            execution_time = time.time() - start_time

            # 결과 출력
            result = {
                "hook": "session_start__auto_cleanup",
                "success": True,
                "execution_time_seconds": round(execution_time, 2),
                "cleanup_stats": cleanup_stats,
                "daily_analysis_report": report_path,
                "timestamp": datetime.now().isoformat()
            }

            print(json.dumps(result, ensure_ascii=False, indent=2))

        finally:
            signal.alarm(0)  # 타임아웃 해제

    except TimeoutError as e:
        # 타임아웃 처리
        result = {
            "hook": "session_start__auto_cleanup",
            "success": False,
            "error": f"Hook execution timeout: {str(e)}",
            "graceful_degradation": graceful_degradation,
            "timestamp": datetime.now().isoformat()
        }

        if graceful_degradation:
            result["message"] = "Hook timeout but continuing due to graceful degradation"

        print(json.dumps(result, ensure_ascii=False, indent=2))

    except Exception as e:
        # 예외 처리
        result = {
            "hook": "session_start__auto_cleanup",
            "success": False,
            "error": f"Hook execution failed: {str(e)}",
            "graceful_degradation": graceful_degradation,
            "timestamp": datetime.now().isoformat()
        }

        if graceful_degradation:
            result["message"] = "Hook failed but continuing due to graceful degradation"

        print(json.dumps(result, ensure_ascii=False, indent=2))


if __name__ == "__main__":
    main()
