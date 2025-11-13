"""Claude Code Headless-based Merge Analyzer

Analyzes template merge differences using Claude Code headless mode
for intelligent backup vs new template comparison and recommendations.
"""

import json
import subprocess
import sys
from difflib import unified_diff
from pathlib import Path
from typing import Any, Optional

import click
from rich.console import Console
from rich.panel import Panel
from rich.table import Table

console = Console()


class MergeAnalyzer:
    """분석기: Claude Code를 사용한 지능형 병합 분석

    백업된 사용자 설정과 새 템플릿을 비교하여 Claude AI가 분석하고
    병합 권장사항을 제시합니다.
    """

    # 분석할 주요 파일 목록
    ANALYZED_FILES = [
        "CLAUDE.md",
        ".claude/settings.json",
        ".moai/config/config.json",
        ".gitignore",
    ]

    # Claude headless 실행 설정
    CLAUDE_TIMEOUT = 120  # 최대 2분
    CLAUDE_MODEL = "claude-sonnet-4-5-20250929"  # 최신 Sonnet
    CLAUDE_TOOLS = ["Read", "Glob", "Grep"]  # 읽기 전용

    def __init__(self, project_path: Path):
        """Initialize analyzer with project path."""
        self.project_path = project_path

    def analyze_merge(
        self, backup_path: Path, template_path: Path
    ) -> dict[str, Any]:
        """Claude Code headless로 병합 분석 수행

        Args:
            backup_path: 백업된 설정 디렉토리 경로
            template_path: 새 템플릿 디렉토리 경로

        Returns:
            분석 결과를 담은 딕셔너리
                - files: 파일별 변경사항 리스트
                - safe_to_auto_merge: 자동 병합 안전 여부
                - user_action_required: 사용자 개입 필요 여부
                - summary: 종합 요약
                - error: 오류 메시지 (있는 경우)
        """
        console.print("🔍 분석 중: Claude Code로 병합 분석 진행 중...")

        # 1. 비교할 파일 수집
        diff_files = self._collect_diff_files(backup_path, template_path)

        # 2. Claude headless 프롬프트 작성
        prompt = self._create_analysis_prompt(
            backup_path, template_path, diff_files
        )

        # 3. Claude Code headless 실행
        try:
            result = subprocess.run(
                self._build_claude_command(),
                input=prompt,
                capture_output=True,
                text=True,
                timeout=self.CLAUDE_TIMEOUT,
            )

            if result.returncode == 0:
                try:
                    analysis = json.loads(result.stdout)
                    return analysis
                except json.JSONDecodeError as e:
                    console.print(
                        f"⚠️  Claude 응답 파싱 오류: {e}",
                        style="yellow",
                    )
                    return self._fallback_analysis(
                        backup_path, template_path, diff_files
                    )
            else:
                console.print(
                    f"⚠️  Claude 실행 오류: {result.stderr[:200]}",
                    style="yellow",
                )
                return self._fallback_analysis(
                    backup_path, template_path, diff_files
                )

        except subprocess.TimeoutExpired:
            console.print(
                "⚠️  Claude 분석 타임아웃 (120초 초과)",
                style="yellow",
            )
            return self._fallback_analysis(
                backup_path, template_path, diff_files
            )
        except FileNotFoundError:
            console.print(
                "❌ Claude Code를 찾을 수 없습니다.",
                style="red",
            )
            console.print(
                "   Claude Code 설치: https://claude.com/claude-code",
                style="cyan",
            )
            return self._fallback_analysis(
                backup_path, template_path, diff_files
            )

    def ask_user_confirmation(self, analysis: dict[str, Any]) -> bool:
        """분석 결과를 표시하고 사용자 승인 요청

        Args:
            analysis: analyze_merge() 결과

        Returns:
            True: 진행, False: 취소
        """
        # 1. 분석 결과 표시
        self._display_analysis(analysis)

        # 2. 사용자 확인
        if analysis.get("user_action_required", False):
            console.print(
                "\n⚠️  사용자 개입이 필요합니다. 아래 사항을 검토하세요:",
                style="warning",
            )
            for file_info in analysis.get("files", []):
                if file_info.get("conflict_severity") in ["medium", "high"]:
                    console.print(
                        f"   • {file_info['filename']}: {file_info.get('note', '')}",
                    )

        # 3. 확인 프롬프트
        proceed = click.confirm(
            "\n병합을 진행하시겠습니까?",
            default=analysis.get("safe_to_auto_merge", False),
        )

        return proceed

    def _collect_diff_files(
        self, backup_path: Path, template_path: Path
    ) -> dict[str, dict[str, Any]]:
        """백업과 템플릿 간 차이 파일 수집

        Returns:
            파일별 diff 정보 딕셔너리
        """
        diff_files = {}

        for file_name in self.ANALYZED_FILES:
            backup_file = backup_path / file_name
            template_file = template_path / file_name

            if not backup_file.exists() and not template_file.exists():
                continue

            diff_info = {
                "backup_exists": backup_file.exists(),
                "template_exists": template_file.exists(),
                "has_diff": False,
                "diff_lines": 0,
            }

            if backup_file.exists() and template_file.exists():
                backup_content = backup_file.read_text(encoding="utf-8")
                template_content = template_file.read_text(encoding="utf-8")

                if backup_content != template_content:
                    diff = list(
                        unified_diff(
                            backup_content.splitlines(),
                            template_content.splitlines(),
                            lineterm="",
                        )
                    )
                    diff_info["has_diff"] = True
                    diff_info["diff_lines"] = len(diff)

            diff_files[file_name] = diff_info

        return diff_files

    def _create_analysis_prompt(
        self,
        backup_path: Path,
        template_path: Path,
        diff_files: dict[str, dict[str, Any]],
    ) -> str:
        """Claude headless 분석 프롬프트 생성

        Returns:
            Claude에게 전달할 분석 프롬프트
        """
        return f"""당신은 MoAI-ADK 설정 파일 병합 전문가입니다.

## 컨텍스트
- 백업된 사용자 설정: {backup_path}
- 새 템플릿: {template_path}
- 분석할 파일: {', '.join(self.ANALYZED_FILES)}

## 분석 대상 파일
{self._format_diff_summary(diff_files)}

## 분석 작업
다음 항목을 분석하고 JSON 응답을 제공하세요:

1. 각 파일별 변경사항 식별
2. 충돌 위험도 평가 (low/medium/high)
3. 병합 권장사항 (use_template/keep_existing/smart_merge)
4. 전반적 안전성 평가

## 응답 형식 (JSON)
{{
  "files": [
    {{
      "filename": "CLAUDE.md",
      "changes": "변경사항 설명",
      "recommendation": "use_template|keep_existing|smart_merge",
      "conflict_severity": "low|medium|high",
      "note": "추가 설명 (선택사항)"
    }}
  ],
  "safe_to_auto_merge": true/false,
  "user_action_required": true/false,
  "summary": "병합 가능 여부와 이유",
  "risk_assessment": "위험도 평가"
}}

## 병합 규칙 참고
- CLAUDE.md: Project Information 섹션 보존
- settings.json: env 변수는 병합, permissions.deny는 템플릿 우선
- config.json: 사용자 메타데이터 보존, 스키마 업데이트
- .gitignore: 추가만 (기존 항목 보존)

## 추가 고려사항
- 사용자 커스터마이징이 손실될 위험 평가
- Alfred 인프라 파일의 강제 덮어쓰기 여부
- 롤백 가능성 검토
"""

    def _display_analysis(self, analysis: dict[str, Any]) -> None:
        """분석 결과를 Rich 형식으로 표시"""
        # 제목
        console.print("\n📊 병합 분석 결과 (Claude Code 분석)", style="bold")

        # 요약
        summary = analysis.get("summary", "분석 결과 없음")
        console.print(f"\n📝 {summary}")

        # 위험도 평가
        risk_assessment = analysis.get("risk_assessment", "")
        if risk_assessment:
            risk_style = "green" if "safe" in risk_assessment.lower() else "yellow"
            console.print(f"⚠️  위험도: {risk_assessment}", style=risk_style)

        # 파일별 변경사항 테이블
        if analysis.get("files"):
            table = Table(title="파일별 변경사항")
            table.add_column("파일", style="cyan")
            table.add_column("변경사항", style="white")
            table.add_column("권장", style="yellow")
            table.add_column("위험도", style="red")

            for file_info in analysis["files"]:
                severity_style = {
                    "low": "green",
                    "medium": "yellow",
                    "high": "red",
                }.get(file_info.get("conflict_severity", "low"), "white")

                table.add_row(
                    file_info.get("filename", "?"),
                    file_info.get("changes", "")[:30],
                    file_info.get("recommendation", "?"),
                    file_info.get("conflict_severity", "?"),
                    style=severity_style,
                )

            console.print(table)

            # 추가 설명
            for file_info in analysis["files"]:
                if file_info.get("note"):
                    console.print(
                        f"\n💡 {file_info['filename']}: {file_info['note']}",
                        style="dim",
                    )

    def _build_claude_command(self) -> list[str]:
        """Claude Code headless 명령어 구축"""
        return [
            "claude",
            "-p",
            "--output-format",
            "json",
            "--tools",
            ",".join(self.CLAUDE_TOOLS),
        ]

    def _format_diff_summary(
        self, diff_files: dict[str, dict[str, Any]]
    ) -> str:
        """diff_files를 프롬프트 형식으로 정렬"""
        summary = []
        for file_name, info in diff_files.items():
            if info["backup_exists"] and info["template_exists"]:
                status = (
                    f"✏️  변경됨 ({info['diff_lines']} 줄)"
                    if info["has_diff"]
                    else "✓ 동일"
                )
            elif info["backup_exists"]:
                status = "❌ 템플릿에서 삭제됨"
            else:
                status = "✨ 새 파일 (템플릿)"

            summary.append(f"- {file_name}: {status}")

        return "\n".join(summary)

    def _fallback_analysis(
        self,
        backup_path: Path,
        template_path: Path,
        diff_files: dict[str, dict[str, Any]],
    ) -> dict[str, Any]:
        """Claude 호출 실패 시 기본 분석 (difflib 기반)

        Claude를 사용할 수 없을 때 기본적인 분석 결과 반환
        """
        console.print(
            "⚠️  Claude Code를 사용할 수 없습니다. 기본 분석을 사용합니다.",
            style="yellow",
        )

        files_analysis = []
        has_high_risk = False

        for file_name, info in diff_files.items():
            if not info["has_diff"]:
                continue

            # 간단한 위험도 평가
            severity = "low"
            if file_name in [".claude/settings.json", ".moai/config/config.json"]:
                severity = "medium" if info["diff_lines"] > 10 else "low"

            files_analysis.append({
                "filename": file_name,
                "changes": f"{info['diff_lines']} 줄 변경됨",
                "recommendation": "smart_merge",
                "conflict_severity": severity,
            })

            if severity == "high":
                has_high_risk = True

        return {
            "files": files_analysis,
            "safe_to_auto_merge": not has_high_risk,
            "user_action_required": has_high_risk,
            "summary": f"{len(files_analysis)}개 파일 변경 감지 (기본 분석)",
            "risk_assessment": "높음 - Claude 분석 불가, 수동 검토 권장" if has_high_risk else "낮음",
            "fallback": True,
        }
