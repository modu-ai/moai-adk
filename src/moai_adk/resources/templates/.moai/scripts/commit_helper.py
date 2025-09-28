#!/usr/bin/env python3
"""
MoAI 커밋 도우미 v0.3.0 (TRUST 모듈화 시스템)
자동 메시지 생성 및 TRUST 원칙 준수 커밋 시스템 – 모듈화된 아키텍처

@REQ:GIT-COMMIT-001
@FEATURE:AUTO-COMMIT-001
@API:GET-COMMIT
@DESIGN:COMMIT-WORKFLOW-003
@TECH:CLAUDE-CODE-STD-001
@TRUST:UNIFIED
"""

import argparse
import sys
from pathlib import Path
from typing import Any

import click

# Import modular system components
scripts_path = str(Path(__file__).parent)
utils_path = str(Path(__file__).parent / "utils")

for path in [scripts_path, utils_path]:
    if path not in sys.path:
        sys.path.insert(0, path)

from commit_validator import CommitValidator
from file_analyzer import FileAnalyzer
from git_workflow import GitWorkflow, GitWorkflowError
from message_generator import MessageGenerator
from project_helper import ProjectHelper


class CommitHelper:
    """커밋 도우미 (TRUST 모듈화 시스템)"""

    def __init__(self):
        self.project_root = Path(__file__).resolve().parents[2]
        self.git_workflow = GitWorkflow(self.project_root)
        self.config = ProjectHelper.load_config(self.project_root)
        self.mode = self.config.get("mode", "personal")

        # Initialize modular components
        self.validator = CommitValidator()
        self.analyzer = FileAnalyzer()
        self.message_generator = MessageGenerator()

    def get_changed_files(self) -> dict[str, Any]:
        """변경된 파일 목록 조회 (모듈화된 분석 사용)"""
        try:
            result = self.git_workflow.git.run_command(["git", "status", "--porcelain"])

            changed_files = []
            for line in result.stdout.splitlines():
                if line.strip():
                    status = line[:2]
                    filename = line[3:].strip()
                    changed_files.append(
                        {
                            "status": status,
                            "filename": filename,
                            "type": self.analyzer.classify_file_change(status),
                        }
                    )

            return {
                "success": True,
                "files": changed_files,
                "count": len(changed_files),
                "has_changes": len(changed_files) > 0,
                "analysis": self.analyzer.analyze_file_changes(changed_files),
            }
        except Exception as e:
            return {"success": False, "error": str(e)}

    def create_smart_commit(
        self, message: str | None = None, files: list[str] | None = None
    ) -> dict[str, Any]:
        """스마트 커밋 생성 (모듈화된 검증 및 메시지 생성)"""
        try:
            # 변경사항 확인 및 검증
            changes = self.get_changed_files()
            if not changes["success"]:
                return changes

            change_validation = self.validator.validate_change_context(changes)
            if not change_validation["valid"]:
                return {"success": False, "error": change_validation["reason"]}

            # 파일 목록 검증
            file_validation = self.validator.validate_file_list(files)
            if not file_validation["valid"]:
                return {"success": False, "error": file_validation["reason"]}

            # 메시지가 없으면 자동 생성
            if not message:
                message = self.message_generator.generate_smart_message(changes["files"])

            # 커밋 실행
            commit_hash = self.git_workflow.create_constitution_commit(message, files)

            return {
                "success": True,
                "commit_hash": commit_hash,
                "message": message,
                "files_changed": changes["count"],
                "mode": self.mode,
                "analysis": changes.get("analysis", {}),
            }
        except GitWorkflowError as e:
            return {"success": False, "error": str(e)}

    def create_constitution_commit(
        self, message: str, files: list[str] | None = None
    ) -> dict[str, Any]:
        """TRUST 원칙 기반 커밋 생성 (모듈화된 검증)"""
        try:
            # 메시지 검증
            validation = self.validator.validate_commit_message(message)
            if not validation["valid"]:
                return {
                    "success": False,
                    "error": f"메시지 검증 실패: {validation['reason']}",
                }

            # 파일 목록 검증
            file_validation = self.validator.validate_file_list(files)
            if not file_validation["valid"]:
                return {"success": False, "error": file_validation["reason"]}

            # 커밋 실행
            commit_hash = self.git_workflow.create_constitution_commit(message, files)

            return {
                "success": True,
                "commit_hash": commit_hash,
                "message": message,
                "validation": validation,
                "file_validation": file_validation,
                "mode": self.mode,
            }
        except GitWorkflowError as e:
            return {"success": False, "error": str(e)}

    def suggest_commit_message(self, context: str | None = None) -> dict[str, Any]:
        """커밋 메시지 제안 (모듈화된 메시지 생성)"""
        try:
            changes = self.get_changed_files()
            if not changes["success"]:
                return changes

            change_validation = self.validator.validate_change_context(changes)
            if not change_validation["valid"]:
                return {"success": False, "error": change_validation["reason"]}

            suggestions = []

            # 파일 변경 기반 제안
            auto_message = self.message_generator.generate_smart_message(changes["files"])
            suggestions.append(
                {
                    "type": "auto",
                    "message": auto_message,
                    "confidence": self.message_generator.calculate_confidence(changes["files"]),
                }
            )

            # 컨텍스트 기반 제안
            if context:
                context_message = self.message_generator.generate_context_message(
                    context, changes["files"]
                )
                suggestions.append(
                    {"type": "context", "message": context_message, "confidence": 0.8}
                )

            # 템플릿 기반 제안
            template_suggestions = self.message_generator.generate_template_suggestions()
            suggestions.extend(template_suggestions)

            return {
                "success": True,
                "suggestions": suggestions,
                "files_changed": changes["count"],
                "change_summary": changes.get("analysis", {}).get("summary", {}),
                "analysis": changes.get("analysis", {}),
            }
        except Exception as e:
            return {"success": False, "error": str(e)}



def main():
    """CLI 엔트리포인트"""
    parser = argparse.ArgumentParser(description="MoAI 커밋 도우미 시스템")

    subparsers = parser.add_subparsers(dest="command", help="사용 가능한 명령어")

    # status 명령어
    subparsers.add_parser("status", help="변경된 파일 상태 조회")

    # commit 명령어
    commit_parser = subparsers.add_parser("commit", help="커밋 생성")
    commit_parser.add_argument("message", help="커밋 메시지")
    commit_parser.add_argument("--files", nargs="*", help="특정 파일만 커밋")

    # smart 명령어
    smart_parser = subparsers.add_parser("smart", help="스마트 커밋 (자동 메시지)")
    smart_parser.add_argument("--message", help="커스텀 메시지")
    smart_parser.add_argument("--files", nargs="*", help="특정 파일만 커밋")

    # suggest 명령어
    suggest_parser = subparsers.add_parser("suggest", help="커밋 메시지 제안")
    suggest_parser.add_argument("--context", help="추가 컨텍스트")

    args = parser.parse_args()

    if not args.command:
        parser.print_help()
        return

    helper = CommitHelper()

    try:
        if args.command == "status":
            result = helper.get_changed_files()
            if result["success"]:
                click.echo(f"\n변경된 파일 ({result['count']}개):")
                click.echo("-" * 60)
                for file in result["files"]:
                    status_symbol = {
                        "added": "+",
                        "modified": "M",
                        "deleted": "-",
                        "renamed": "R",
                    }.get(file["type"], "?")
                    click.echo(f"  {status_symbol} {file['filename']} ({file['type']})")

                if not result["has_changes"]:
                    click.echo("  변경사항이 없습니다")

                # 분석 정보 표시
                if "analysis" in result and result["analysis"]["total_files"] > 0:
                    analysis = result["analysis"]
                    click.echo("\n📊 변경사항 분석:")
                    click.echo(f"  파일 카테고리: Python({analysis['categories']['python']}) "
                             f"Docs({analysis['categories']['docs']}) "
                             f"Config({analysis['categories']['config']}) "
                             f"Test({analysis['categories']['tests']}) "
                             f"기타({analysis['categories']['other']})")
                    if analysis["is_simple_change"]:
                        click.echo("  🟢 단순 변경 (권장: smart commit)")
                    else:
                        click.echo("  🟡 복합 변경 (권장: 수동 메시지 작성)")
            else:
                click.echo(f"❌ 상태 조회 실패: {result['error']}")

        elif args.command == "commit":
            result = helper.create_constitution_commit(args.message, args.files)
            if result["success"]:
                click.echo(f"✅ 커밋 생성 완료: {result['commit_hash'][:8]}")
                click.echo(f"   메시지: {result['message']}")
                click.echo(f"   모드: {result['mode']}")
                if "validation" in result:
                    click.echo(f"   검증: {result['validation']['reason']}")
            else:
                click.echo(f"❌ 커밋 실패: {result['error']}")

        elif args.command == "smart":
            result = helper.create_smart_commit(args.message, args.files)
            if result["success"]:
                click.echo(f"✅ 스마트 커밋 완료: {result['commit_hash'][:8]}")
                click.echo(f"   메시지: {result['message']}")
                click.echo(f"   변경 파일: {result['files_changed']}개")
                click.echo(f"   모드: {result['mode']}")
            else:
                click.echo(f"❌ 스마트 커밋 실패: {result['error']}")

        elif args.command == "suggest":
            result = helper.suggest_commit_message(args.context)
            if result["success"]:
                click.echo(
                    f"\n커밋 메시지 제안 (변경 파일: {result['files_changed']}개):"
                )
                click.echo("-" * 60)
                for i, suggestion in enumerate(result["suggestions"], 1):
                    confidence_bar = "★" * int(suggestion["confidence"] * 5)
                    click.echo(f"{i}. {suggestion['message']}")
                    click.echo(
                        f"   유형: {suggestion['type']}, 신뢰도: {confidence_bar} ({suggestion['confidence']:.1f})"
                    )
                    click.echo()

                click.echo("변경사항 요약:")
                summary = result["change_summary"]
                for change_type, count in summary.items():
                    if count > 0:
                        click.echo(f"  {change_type}: {count}개")

                # 분석 정보 추가 표시
                if "analysis" in result:
                    analysis = result["analysis"]
                    click.echo(f"\n📊 추가 분석:")
                    click.echo(f"  커밋 유형 제안: {helper.analyzer.suggest_commit_type(analysis)}")
                    if analysis["has_tests"]:
                        click.echo("  ✅ 테스트 파일 포함")
                    if analysis["is_documentation_only"]:
                        click.echo("  📚 문서 전용 변경")
            else:
                click.echo(f"❌ 제안 실패: {result['error']}")

    except Exception as e:
        click.echo(f"❌ 오류 발생: {e}")


if __name__ == "__main__":
    main()
