#!/usr/bin/env python3
"""
MoAI 커밋 도우미 v0.2.0 (통합 시스템 기반)
자동 메시지 생성 및 TRUST 원칙 준수 커밋 시스템 – 통합 Git 워크플로우 사용

@REQ:GIT-COMMIT-001
@FEATURE:AUTO-COMMIT-001
@API:GET-COMMIT
@DESIGN:COMMIT-WORKFLOW-002
@TECH:CLAUDE-CODE-STD-001
"""

import argparse
import re
import sys
from pathlib import Path
from typing import Any

import click

# 새로운 통합 시스템 import
utils_path = str(Path(__file__).parent / "utils")
if utils_path not in sys.path:
    sys.path.insert(0, utils_path)

from git_workflow import GitWorkflow, GitWorkflowError
from project_helper import ProjectHelper


class CommitHelper:
    """커밋 도우미 (통합 시스템 래퍼)"""

    def __init__(self):
        self.project_root = Path(__file__).resolve().parents[2]
        self.git_workflow = GitWorkflow(self.project_root)
        self.config = ProjectHelper.load_config(self.project_root)
        self.mode = self.config.get("mode", "personal")

    def get_changed_files(self) -> dict[str, Any]:
        """변경된 파일 목록 조회"""
        try:
            result = self.git_workflow.git.run_command(["git", "status", "--porcelain"])

            changed_files = []
            for line in result.stdout.splitlines():
                if line.strip():
                    status = line[:2]
                    filename = line[3:].strip()
                    changed_files.append({
                        "status": status,
                        "filename": filename,
                        "type": self._classify_file_change(status)
                    })

            return {
                "success": True,
                "files": changed_files,
                "count": len(changed_files),
                "has_changes": len(changed_files) > 0
            }
        except Exception as e:
            return {"success": False, "error": str(e)}

    def create_smart_commit(self, message: str | None = None, files: list[str] | None = None) -> dict[str, Any]:
        """스마트 커밋 생성"""
        try:
            # 변경사항 확인
            changes = self.get_changed_files()
            if not changes["success"]:
                return changes

            if not changes["has_changes"]:
                return {"success": False, "error": "커밋할 변경사항이 없습니다"}

            # 메시지가 없으면 자동 생성
            if not message:
                message = self._generate_smart_message(changes["files"])

            # 커밋 실행
            commit_hash = self.git_workflow.create_constitution_commit(message, files)

            return {
                "success": True,
                "commit_hash": commit_hash,
                "message": message,
                "files_changed": changes["count"],
                "mode": self.mode
            }
        except GitWorkflowError as e:
            return {"success": False, "error": str(e)}

    def create_constitution_commit(self, message: str, files: list[str] | None = None) -> dict[str, Any]:
        """TRUST 원칙 기반 커밋 생성"""
        try:
            # 메시지 검증
            validation = self._validate_commit_message(message)
            if not validation["valid"]:
                return {"success": False, "error": f"메시지 검증 실패: {validation['reason']}"}

            # 커밋 실행
            commit_hash = self.git_workflow.create_constitution_commit(message, files)

            return {
                "success": True,
                "commit_hash": commit_hash,
                "message": message,
                "validation": validation,
                "mode": self.mode
            }
        except GitWorkflowError as e:
            return {"success": False, "error": str(e)}

    def suggest_commit_message(self, context: str | None = None) -> dict[str, Any]:
        """커밋 메시지 제안"""
        try:
            changes = self.get_changed_files()
            if not changes["success"]:
                return changes

            if not changes["has_changes"]:
                return {"success": False, "error": "변경사항이 없습니다"}

            suggestions = []

            # 파일 변경 기반 제안
            auto_message = self._generate_smart_message(changes["files"])
            suggestions.append({
                "type": "auto",
                "message": auto_message,
                "confidence": self._calculate_confidence(changes["files"])
            })

            # 컨텍스트 기반 제안
            if context:
                context_message = self._generate_context_message(context, changes["files"])
                suggestions.append({
                    "type": "context",
                    "message": context_message,
                    "confidence": 0.8
                })

            # 템플릿 기반 제안
            template_suggestions = self._generate_template_suggestions(changes["files"])
            suggestions.extend(template_suggestions)

            return {
                "success": True,
                "suggestions": suggestions,
                "files_changed": changes["count"],
                "change_summary": self._summarize_changes(changes["files"])
            }
        except Exception as e:
            return {"success": False, "error": str(e)}

    def _classify_file_change(self, status: str) -> str:
        """파일 변경 유형 분류"""
        if status.startswith("A"):
            return "added"
        elif status.startswith("M"):
            return "modified"
        elif status.startswith("D"):
            return "deleted"
        elif status.startswith("R"):
            return "renamed"
        elif status.startswith("C"):
            return "copied"
        else:
            return "unknown"

    def _generate_smart_message(self, files: list[dict[str, Any]]) -> str:
        """스마트 커밋 메시지 생성"""
        if not files:
            return "🔧 Minor updates"

        # 변경 유형별 분류
        added = [f for f in files if f["type"] == "added"]
        modified = [f for f in files if f["type"] == "modified"]
        deleted = [f for f in files if f["type"] == "deleted"]

        # 파일 확장자별 분류
        py_files = [f for f in files if f["filename"].endswith(".py")]
        md_files = [f for f in files if f["filename"].endswith(".md")]
        json_files = [f for f in files if f["filename"].endswith(".json")]

        # 메시지 생성 로직
        if len(added) > 0 and len(modified) == 0 and len(deleted) == 0:
            if len(added) == 1:
                return f"✨ Add {added[0]['filename']}"
            else:
                return f"✨ Add {len(added)} new files"

        elif len(modified) > 0 and len(added) == 0 and len(deleted) == 0:
            if len(py_files) > len(md_files):
                return "🔧 Update Python modules"
            elif len(md_files) > 0:
                return "📚 Update documentation"
            elif len(json_files) > 0:
                return "🔧 Update configuration"
            else:
                return "🔧 Update files"

        elif len(deleted) > 0:
            return f"🗑️ Remove {len(deleted)} files"

        else:
            # 혼합 변경
            total = len(files)
            if total <= 3:
                return f"🔧 Update {total} files"
            else:
                return f"♻️ Refactor multiple files ({total} files)"

    def _generate_context_message(self, context: str, files: list[dict[str, Any]]) -> str:
        """컨텍스트 기반 메시지 생성"""
        context_lower = context.lower()

        if "fix" in context_lower or "bug" in context_lower:
            return f"🐛 Fix: {context}"
        elif "feat" in context_lower or "feature" in context_lower:
            return f"✨ Feature: {context}"
        elif "test" in context_lower:
            return f"🧪 Test: {context}"
        elif "doc" in context_lower:
            return f"📚 Docs: {context}"
        elif "refactor" in context_lower:
            return f"♻️ Refactor: {context}"
        else:
            return f"🔧 {context}"

    def _generate_template_suggestions(self, files: list[dict[str, Any]]) -> list[dict[str, Any]]:
        """템플릿 기반 제안 생성"""
        templates = [
            {"type": "feature", "message": "✨ feat: ", "confidence": 0.6},
            {"type": "fix", "message": "🐛 fix: ", "confidence": 0.6},
            {"type": "docs", "message": "📚 docs: ", "confidence": 0.6},
            {"type": "refactor", "message": "♻️ refactor: ", "confidence": 0.6},
            {"type": "test", "message": "🧪 test: ", "confidence": 0.6},
            {"type": "chore", "message": "🔧 chore: ", "confidence": 0.6}
        ]
        return templates

    def _calculate_confidence(self, files: list[dict[str, Any]]) -> float:
        """제안 신뢰도 계산"""
        if not files:
            return 0.0

        # 단일 파일 변경 시 높은 신뢰도
        if len(files) == 1:
            return 0.9

        # 동일 유형 파일들만 변경 시 높은 신뢰도
        file_types = set(f["filename"].split(".")[-1] for f in files if "." in f["filename"])
        if len(file_types) == 1:
            return 0.8

        # 혼합 변경 시 중간 신뢰도
        return 0.6

    def _summarize_changes(self, files: list[dict[str, Any]]) -> dict[str, int]:
        """변경사항 요약"""
        summary = {"added": 0, "modified": 0, "deleted": 0, "renamed": 0}
        for file in files:
            file_type = file.get("type", "unknown")
            if file_type in summary:
                summary[file_type] += 1
        return summary

    def _validate_commit_message(self, message: str) -> dict[str, Any]:
        """커밋 메시지 검증"""
        if not message or not message.strip():
            return {"valid": False, "reason": "메시지가 비어있습니다"}

        if len(message) < 10:
            return {"valid": False, "reason": "메시지가 너무 짧습니다 (최소 10자)"}

        if len(message) > 200:
            return {"valid": False, "reason": "메시지가 너무 깁니다 (최대 200자)"}

        # 이모지 체크 (선택사항)
        has_emoji = bool(re.search(r'[\U0001F600-\U0001F64F\U0001F300-\U0001F5FF\U0001F680-\U0001F6FF\U0001F1E0-\U0001F1FF]', message))

        return {
            "valid": True,
            "reason": "검증 통과",
            "has_emoji": has_emoji,
            "length": len(message)
        }


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
                    status_symbol = {"added": "+", "modified": "M", "deleted": "-", "renamed": "R"}.get(file["type"], "?")
                    click.echo(f"  {status_symbol} {file['filename']} ({file['type']})")

                if not result["has_changes"]:
                    click.echo("  변경사항이 없습니다")
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
                click.echo(f"\n커밋 메시지 제안 (변경 파일: {result['files_changed']}개):")
                click.echo("-" * 60)
                for i, suggestion in enumerate(result["suggestions"], 1):
                    confidence_bar = "★" * int(suggestion["confidence"] * 5)
                    click.echo(f"{i}. {suggestion['message']}")
                    click.echo(f"   유형: {suggestion['type']}, 신뢰도: {confidence_bar} ({suggestion['confidence']:.1f})")
                    click.echo()

                click.echo("변경사항 요약:")
                summary = result["change_summary"]
                for change_type, count in summary.items():
                    if count > 0:
                        click.echo(f"  {change_type}: {count}개")
            else:
                click.echo(f"❌ 제안 실패: {result['error']}")

    except Exception as e:
        click.echo(f"❌ 오류 발생: {e}")


if __name__ == "__main__":
    main()
