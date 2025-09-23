#!/usr/bin/env python3
"""Minimal document/tag sync helper for /moai:3-sync.

Performs lightweight automated steps (currently TAG index refresh)
and prints follow-up instructions for manual actions that remain.
"""

from __future__ import annotations

import argparse
import os
import subprocess
import sys
from datetime import datetime
from pathlib import Path


def run(cmd: list[str], cwd: Path) -> tuple[int, str, str]:
    result = subprocess.run(
        cmd,
        cwd=cwd,
        capture_output=True,
        text=True,
    )
    return result.returncode, result.stdout.strip(), result.stderr.strip()


def update_tag_index(project_root: Path) -> dict[str, object]:
    script = project_root / ".moai" / "scripts" / "check-traceability.py"
    if not script.exists():
        print("⚠️  TAG 추적성 스크립트를 찾을 수 없습니다. (.moai/scripts/check-traceability.py)")
        return {"success": False, "stdout": "", "stderr": "missing"}

    code, out, err = run(
        [sys.executable or "python3", str(script), "--update"],
        project_root,
    )
    if code == 0:
        print("✅ TAG 추적성 인덱스를 갱신했습니다.")
        if out:
            print(out)
    else:
        print("⚠️  TAG 인덱스 갱신 중 문제가 발생했습니다.")
        if err:
            print(err)
    return {"success": code == 0, "stdout": out, "stderr": err}


def show_git_hint(project_root: Path) -> None:
    code, out, _ = run(["git", "status", "--short"], project_root)
    if code == 0 and out:
        print("📂 변경된 파일:")
        print(out)
        print(
            "💡 필요 시 `git add README.md docs/ .moai/indexes/tags.json` 후 커밋하세요."
        )


def collect_git_changes(project_root: Path) -> dict[str, object]:
    code, out, _ = run(["git", "status", "--short"], project_root)
    lines = out.splitlines() if code == 0 and out else []
    return {
        "count": len(lines),
        "examples": lines[:5],
    }


def _spec_status(spec_dir: Path) -> dict[str, object]:
    spec_file = spec_dir / "spec.md"
    plan_file = spec_dir / "plan.md"
    tasks_file = spec_dir / "tasks.md"
    status = "missing"
    note = "spec.md 누락"
    if spec_file.exists():
        try:
            content = spec_file.read_text(encoding="utf-8")
        except Exception:
            content = ""
        stripped = content.strip()
        if "[NEEDS CLARIFICATION" in stripped:
            status, note = "needs_clarification", "clarification 표시"
        elif len(stripped) < 400:
            status, note = "draft", "내용 부족"
        else:
            status, note = "ready", ""
    return {
        "id": spec_dir.name,
        "status": status,
        "has_plan": plan_file.exists(),
        "has_tasks": tasks_file.exists(),
        "updated": datetime.fromtimestamp(spec_dir.stat().st_mtime).isoformat(),
        "note": note,
    }


def collect_spec_status(project_root: Path) -> dict[str, object]:
    specs_root = project_root / ".moai" / "specs"
    summary = {
        "total": 0,
        "ready": 0,
        "draft": 0,
        "needs_clarification": 0,
        "missing": 0,
        "items": [],
    }
    if not specs_root.exists():
        return summary

    for spec_dir in sorted(specs_root.iterdir()):
        if not spec_dir.is_dir() or spec_dir.name.startswith("_") or not spec_dir.name.startswith("SPEC-"):
            continue
        data = _spec_status(spec_dir)
        summary["items"].append(data)
        summary["total"] += 1
        summary[data["status"]] += 1
    return summary


def _render_tag_section(tag_result: dict[str, object]) -> list[str]:
    lines = ["## TAG Traceability"]
    lines.append("- Status: ✅ Updated" if tag_result.get("success") else "- Status: ⚠️ Failed (see stderr)")
    if tag_result.get("stdout"):
        lines.append("- Output:")
        lines.append("  ```")
        lines.extend(f"  {line}" for line in str(tag_result["stdout"]).splitlines())
        lines.append("  ```")
    if tag_result.get("stderr") and tag_result.get("stderr") != "missing":
        lines.append("- Errors:")
        lines.append("  ```")
        lines.extend(f"  {line}" for line in str(tag_result["stderr"]).splitlines())
        lines.append("  ```")
    return lines


def _render_spec_section(spec_summary: dict[str, object]) -> list[str]:
    lines = ["## SPEC Overview"]
    lines.append(
        f"- Total: {spec_summary['total']} | Ready: {spec_summary['ready']} | Draft: {spec_summary['draft']} | Needs Clarification: {spec_summary['needs_clarification']} | Missing: {spec_summary['missing']}"
    )
    if spec_summary["items"]:
        lines.append("")
        lines.append("| SPEC | Status | Plan | Tasks | Updated | Note |")
        lines.append("| --- | --- | --- | --- | --- | --- |")
        for item in spec_summary["items"]:
            lines.append(
                "| {id} | {status} | {plan} | {tasks} | {updated} | {note} |".format(
                    id=item["id"],
                    status=item["status"],
                    plan="✅" if item["has_plan"] else "",
                    tasks="✅" if item["has_tasks"] else "",
                    updated=item["updated"],
                    note=item.get("note") or "",
                )
            )
    return lines


def _render_git_section(git_changes: dict[str, object]) -> list[str]:
    lines = ["## Git Working Tree"]
    lines.append(f"- Pending changes: {git_changes['count']}")
    if git_changes["examples"]:
        lines.append("- Samples:")
        lines.extend(f"  - {entry}" for entry in git_changes["examples"])
    return lines


def render_sync_report(
    project_root: Path,
    tag_result: dict[str, object],
    spec_summary: dict[str, object],
    git_changes: dict[str, object],
) -> str:
    timestamp = datetime.now().isoformat(timespec="seconds")
    lines = ["# MoAI-ADK Sync Report", ""]
    lines.append(f"- Generated: {timestamp}")
    lines.append(f"- Project: {project_root.name}")
    lines.append("")
    lines.extend(_render_tag_section(tag_result))
    lines.append("")
    lines.extend(_render_spec_section(spec_summary))
    lines.append("")
    lines.extend(_render_git_section(git_changes))
    lines.append("")
    lines.append("## Next Manual Steps")
    lines.append("1. README 및 관련 문서를 최신 상태로 정리")
    lines.append("2. 필요 시 PR 라벨과 체크리스트 업데이트")
    lines.append("3. 변경 사항 검토 후 커밋 및 푸시")
    return "\n".join(lines) + "\n"


def write_sync_report(project_root: Path, report: str) -> Path:
    report_dir = project_root / "docs" / "status"
    report_dir.mkdir(parents=True, exist_ok=True)
    report_path = report_dir / "sync-report.md"
    report_path.write_text(report, encoding="utf-8")
    return report_path


def update_doc_index_metadata(project_root: Path) -> None:
    index_path = project_root / "docs" / "sections" / "index.md"
    if not index_path.exists():
        return
    lines = index_path.read_text(encoding="utf-8").splitlines()
    today = datetime.now().date().isoformat()
    updated = []
    replaced = False
    for line in lines:
        if line.startswith("> **Last Updated**:"):
            parts = line.split("|", 1)
            suffix = f" | {parts[1].strip()}" if len(parts) > 1 else ""
            updated.append(f"> **Last Updated**: {today}{suffix}")
            replaced = True
        else:
            updated.append(line)
    if replaced:
        index_path.write_text("\n".join(updated) + "\n", encoding="utf-8")

def main() -> None:
    parser = argparse.ArgumentParser(description="MoAI 문서/태그 동기화 헬퍼")
    parser.add_argument("mode", nargs="?", default="auto", help="auto|force|status|project 등")
    parser.add_argument("target", nargs="?", help="동기화 대상 경로 (선택)")
    args = parser.parse_args()

    project_root = Path(os.environ.get("CLAUDE_PROJECT_DIR", Path.cwd()))

    print("🔄 MoAI 문서/태그 동기화를 실행합니다.")
    print(f"  - 모드: {args.mode}")
    if args.target:
        print(f"  - 대상: {args.target}")

    tag_result = update_tag_index(project_root)
    spec_summary = collect_spec_status(project_root)
    git_changes = collect_git_changes(project_root)
    report = render_sync_report(project_root, tag_result, spec_summary, git_changes)
    report_path = write_sync_report(project_root, report)
    update_doc_index_metadata(project_root)
    show_git_hint(project_root)

    print("📋 나머지 단계:")
    print("  1. 문서/README를 수동으로 업데이트합니다.")
    print("  2. 필요 시 API/아키텍처 문서를 갱신합니다.")
    print("  3. GitHub PR 상태를 확인하고 라벨/리뷰어를 조정합니다.")
    print(f"📝 동기화 리포트: {report_path.relative_to(project_root)}")


if __name__ == "__main__":
    main()
