#!/usr/bin/env python3
"""
MoAI-ADK Dashboard Renderer - v1.0.0
프로젝트 상태를 시각적으로 표시하는 대시보드 스크립트

Rich 라이브러리를 사용하여 컬러풀한 터미널 출력을 생성합니다.
"""

import json
import os
import sys
import subprocess
from pathlib import Path
from datetime import datetime, timedelta
from typing import Dict, List, Any, Optional, Tuple
import argparse

try:
    from rich.console import Console
    from rich.table import Table
    from rich.panel import Panel
    from rich.progress import Progress, BarColumn, TextColumn, SpinnerColumn
    from rich.layout import Layout
    from rich.text import Text
    from rich.tree import Tree
    from rich.align import Align
    from rich.columns import Columns
    from rich import box
except ImportError:
    print("❌ Rich 라이브러리가 필요합니다. 설치하려면: pip install rich")
    sys.exit(1)


class MoAIDashboard:
    """MoAI-ADK 프로젝트 대시보드 렌더러"""

    def __init__(self, project_root: Path):
        self.project_root = project_root
        self.console = Console(width=100)
        self.state_file = project_root / ".moai" / "indexes" / "state.json"
        self.tags_file = project_root / ".moai" / "indexes" / "tags.json"
        self.version_file = project_root / ".moai" / "indexes" / "version.json"
        self.config_file = project_root / ".moai" / "config.json"

        # 캐시된 데이터
        self._state_data: Optional[Dict] = None
        self._git_info: Optional[Dict] = None
        self._tags_data: Optional[Dict] = None

    def load_state_data(self) -> Dict[str, Any]:
        """프로젝트 상태 데이터 로드"""
        if self._state_data is not None:
            return self._state_data

        try:
            if self.state_file.exists():
                with open(self.state_file, 'r', encoding='utf-8') as f:
                    self._state_data = json.load(f)
            else:
                self._state_data = self._create_default_state()
        except json.JSONDecodeError:
            self._state_data = self._create_default_state()

        return self._state_data

    def _create_default_state(self) -> Dict[str, Any]:
        """기본 상태 데이터 생성"""
        return {
            "metadata": {
                "project_name": self.project_root.name,
                "created_at": datetime.now().strftime("%Y-%m-%d"),
                "last_updated": datetime.now().strftime("%Y-%m-%d")
            },
            "project_state": {
                "current_phase": "INIT",
                "available_phases": ["INIT", "SPECIFY", "PLAN", "TASKS", "IMPLEMENT", "SYNC"],
                "completion_percentage": 0.0
            },
            "pipeline_status": {
                "specify": {"status": "NOT_STARTED", "specs_count": 0},
                "plan": {"status": "NOT_STARTED", "plans_count": 0},
                "tasks": {"status": "NOT_STARTED", "tasks_count": 0},
                "implement": {"status": "NOT_STARTED", "implementations_count": 0}
            },
            "constitution_compliance": {
                "simplicity": {"status": "UNKNOWN", "project_count": 0, "max_allowed": 3},
                "architecture": {"status": "UNKNOWN", "library_count": 0},
                "testing": {"status": "UNKNOWN", "current_coverage": 0.0, "coverage_target": 0.8},
                "observability": {"status": "UNKNOWN", "structured_logging": False},
                "versioning": {"status": "UNKNOWN", "current_version": "0.0.0"}
            },
            "tag_system": {
                "total_tags": 0,
                "by_category": {},
                "traceability_coverage": 0.0
            }
        }

    def load_git_info(self) -> Dict[str, Any]:
        """Git 정보 수집"""
        if self._git_info is not None:
            return self._git_info

        git_info = {
            "is_git_repo": False,
            "current_branch": "unknown",
            "last_commit": {"hash": "", "message": "", "date": ""},
            "status": {"modified": 0, "added": 0, "deleted": 0, "untracked": 0},
            "remote_status": {"ahead": 0, "behind": 0}
        }

        try:
            # Git 저장소 확인
            subprocess.run(["git", "rev-parse", "--git-dir"],
                         capture_output=True, check=True, cwd=self.project_root)
            git_info["is_git_repo"] = True

            # 현재 브랜치
            result = subprocess.run(["git", "branch", "--show-current"],
                                  capture_output=True, text=True, cwd=self.project_root)
            if result.returncode == 0:
                git_info["current_branch"] = result.stdout.strip()

            # 최근 커밋
            result = subprocess.run(["git", "log", "-1", "--pretty=format:%H|%s|%cr"],
                                  capture_output=True, text=True, cwd=self.project_root)
            if result.returncode == 0 and result.stdout:
                parts = result.stdout.split("|")
                if len(parts) >= 3:
                    git_info["last_commit"] = {
                        "hash": parts[0][:8],
                        "message": parts[1],
                        "date": parts[2]
                    }

            # 상태 정보
            result = subprocess.run(["git", "status", "--porcelain"],
                                  capture_output=True, text=True, cwd=self.project_root)
            if result.returncode == 0:
                lines = result.stdout.strip().split('\n') if result.stdout.strip() else []
                for line in lines:
                    if line.startswith('M '):
                        git_info["status"]["modified"] += 1
                    elif line.startswith('A '):
                        git_info["status"]["added"] += 1
                    elif line.startswith('D '):
                        git_info["status"]["deleted"] += 1
                    elif line.startswith('??'):
                        git_info["status"]["untracked"] += 1

        except (subprocess.CalledProcessError, FileNotFoundError):
            pass

        self._git_info = git_info
        return git_info

    def analyze_specs(self) -> List[Dict[str, Any]]:
        """SPEC 디렉토리 분석"""
        specs_dir = self.project_root / ".moai" / "specs"
        specs = []

        if not specs_dir.exists():
            return specs

        for spec_dir in specs_dir.iterdir():
            if (spec_dir.is_dir() and
                spec_dir.name.startswith("SPEC-") and
                not spec_dir.name.endswith("-sample")):

                spec_file = spec_dir / "spec.md"
                plan_file = spec_dir / "plan.md"
                tasks_file = spec_dir / "tasks.md"

                # 제목 추출
                title = "제목 없음"
                if spec_file.exists():
                    try:
                        with open(spec_file, 'r', encoding='utf-8') as f:
                            content = f.read()
                            lines = content.split('\n')
                            for line in lines:
                                if line.startswith('# '):
                                    title = line[2:].strip()
                                    break
                    except:
                        pass

                # 상태 및 진행률 계산
                status = "🔄 진행"
                progress = 0

                if spec_file.exists():
                    progress += 33
                if plan_file.exists():
                    progress += 33
                if tasks_file.exists():
                    progress += 34

                if progress >= 100:
                    status = "✅ 완료"
                elif progress == 0:
                    status = "⏸️ 대기"

                specs.append({
                    "id": spec_dir.name,
                    "title": title,
                    "status": status,
                    "progress": progress,
                    "has_spec": spec_file.exists(),
                    "has_plan": plan_file.exists(),
                    "has_tasks": tasks_file.exists()
                })

        return sorted(specs, key=lambda x: x["id"])

    def calculate_pipeline_progress(self, state_data: Dict) -> Dict[str, Dict]:
        """파이프라인 진행률 계산"""
        pipeline = state_data.get("pipeline_status", {})
        specs = self.analyze_specs()

        # 각 단계별 진행률 계산
        stages = {
            "SPECIFY": {"completed": 0, "total": 0, "percentage": 0},
            "PLAN": {"completed": 0, "total": 0, "percentage": 0},
            "TASKS": {"completed": 0, "total": 0, "percentage": 0},
            "IMPLEMENT": {"completed": 0, "total": 0, "percentage": 0},
            "SYNC": {"completed": 0, "total": 0, "percentage": 0}
        }

        total_specs = len(specs)
        if total_specs > 0:
            # SPECIFY: spec.md 파일이 있는 경우
            specify_completed = sum(1 for s in specs if s["has_spec"])
            stages["SPECIFY"] = {
                "completed": specify_completed,
                "total": total_specs,
                "percentage": (specify_completed / total_specs) * 100
            }

            # PLAN: plan.md 파일이 있는 경우
            plan_completed = sum(1 for s in specs if s["has_plan"])
            stages["PLAN"] = {
                "completed": plan_completed,
                "total": total_specs,
                "percentage": (plan_completed / total_specs) * 100
            }

            # TASKS: tasks.md 파일이 있는 경우
            tasks_completed = sum(1 for s in specs if s["has_tasks"])
            stages["TASKS"] = {
                "completed": tasks_completed,
                "total": total_specs,
                "percentage": (tasks_completed / total_specs) * 100
            }

        return stages

    def render_header(self, state_data: Dict) -> Panel:
        """헤더 섹션 렌더링"""
        project_name = state_data.get("metadata", {}).get("project_name", "MoAI-ADK")
        version = state_data.get("constitution_compliance", {}).get("versioning", {}).get("current_version", "0.0.0")
        current_phase = state_data.get("project_state", {}).get("current_phase", "INIT")
        current_time = datetime.now().strftime("%Y-%m-%d %H:%M")

        header_text = f"프로젝트: {project_name} v{version} | 단계: {current_phase} | {current_time}"
        return Panel(
            Align.center(Text(header_text, style="bold white")),
            title="🗿 MoAI-ADK Dashboard",
            box=box.DOUBLE,
            style="blue"
        )

    def render_pipeline_progress(self, stages: Dict) -> Panel:
        """파이프라인 진행률 렌더링"""
        content = []

        stage_names = ["SPECIFY", "PLAN", "TASKS", "IMPLEMENT", "SYNC"]
        stage_colors = ["green", "blue", "yellow", "magenta", "cyan"]

        for i, (stage, color) in enumerate(zip(stage_names, stage_colors)):
            stage_data = stages.get(stage, {"completed": 0, "total": 0, "percentage": 0})
            percentage = stage_data["percentage"]
            completed = stage_data["completed"]
            total = stage_data["total"]

            # 프로그레스 바 생성
            bar_length = 20
            filled = int((percentage / 100) * bar_length)
            bar = "█" * filled + "░" * (bar_length - filled)

            # 상태 표시
            status_icon = "✅" if percentage == 100 else ("←" if i == self._get_current_stage_index(stages) else "")

            line = f"{stage:<9} {bar} {percentage:3.0f}% [{completed}/{total}] {status_icon}"
            content.append(Text(line, style=color))

        return Panel(
            "\n".join(str(line) for line in content),
            title="📊 개발 파이프라인 진행 상황",
            box=box.ROUNDED
        )

    def _get_current_stage_index(self, stages: Dict) -> int:
        """현재 활성 단계 인덱스 반환"""
        stage_names = ["SPECIFY", "PLAN", "TASKS", "IMPLEMENT", "SYNC"]
        for i, stage in enumerate(stage_names):
            stage_data = stages.get(stage, {"percentage": 0})
            if stage_data["percentage"] < 100:
                return i
        return len(stage_names) - 1

    def render_specs_table(self, specs: List[Dict]) -> Table:
        """SPEC 현황 테이블 렌더링"""
        table = Table(title="📋 SPEC 현황", box=box.ROUNDED)
        table.add_column("SPEC ID", style="cyan", no_wrap=True)
        table.add_column("제목", style="white")
        table.add_column("상태", style="white", justify="center")
        table.add_column("진행률", style="white", justify="center")

        if not specs:
            table.add_row("없음", "SPEC이 생성되지 않았습니다", "⏸️ 대기", "0%")
        else:
            for spec in specs:
                table.add_row(
                    spec["id"],
                    spec["title"][:40] + "..." if len(spec["title"]) > 40 else spec["title"],
                    spec["status"],
                    f"{spec['progress']}%"
                )

        return table

    def render_tag_system(self, state_data: Dict) -> Tree:
        """TAG 시스템 트리 렌더링"""
        tag_data = state_data.get("tag_system", {})
        total_tags = tag_data.get("total_tags", 0)
        by_category = tag_data.get("by_category", {})
        coverage = tag_data.get("traceability_coverage", 0.0)

        health_icon = "✅" if coverage >= 0.9 else "⚠️" if coverage >= 0.7 else "❌"
        tree = Tree(f"🏷️ TAG 시스템 (건강도: {coverage*100:.0f}% {health_icon})")

        if by_category:
            for category, count in by_category.items():
                if isinstance(count, int):
                    tree.add(f"{category}: {count}개")
                else:
                    tree.add(f"{category}: {count}")
        else:
            tree.add("STEERING: 3개 (@VISION, @STRUCT, @TECH)")
            tree.add("SPEC: 0개")
            tree.add("IMPLEMENTATION: 0개")
            tree.add("QUALITY: 0개")

        tree.add(f"전체 연결성: {'완전' if coverage >= 0.95 else '부분적'} (태그 수: {total_tags})")

        return tree

    def render_constitution_status(self, state_data: Dict) -> Table:
        """Constitution 준수 현황 렌더링"""
        constitution = state_data.get("constitution_compliance", {})

        table = Table(title="⚖️ Constitution 5원칙 준수 현황", box=box.ROUNDED)
        table.add_column("원칙", style="white", no_wrap=True)
        table.add_column("상태", style="white")
        table.add_column("세부사항", style="white")

        # Simplicity
        simplicity = constitution.get("simplicity", {})
        project_count = simplicity.get("project_count", 0)
        max_allowed = simplicity.get("max_allowed", 3)
        status_icon = "✅" if project_count <= max_allowed else "❌"
        table.add_row(
            "Simplicity",
            status_icon,
            f"프로젝트 수 {project_count}/{max_allowed}개"
        )

        # Architecture
        architecture = constitution.get("architecture", {})
        arch_status = architecture.get("status", "UNKNOWN")
        arch_icon = "✅" if arch_status == "COMPLIANT" else "⚠️"
        table.add_row(
            "Architecture",
            arch_icon,
            "모든 기능이 라이브러리 구조" if arch_status == "COMPLIANT" else "구조 검증 필요"
        )

        # Testing
        testing = constitution.get("testing", {})
        coverage = testing.get("current_coverage", 0.0)
        target = testing.get("coverage_target", 0.8)
        test_icon = "✅" if coverage >= target else "⚠️"
        table.add_row(
            "Testing",
            test_icon,
            f"커버리지 {coverage*100:.0f}% (목표: {target*100:.0f}%)"
        )

        # Observability
        observability = constitution.get("observability", {})
        logging = observability.get("structured_logging", False)
        obs_icon = "✅" if logging else "⚠️"
        table.add_row(
            "Observability",
            obs_icon,
            "구조화 로깅 활성화됨" if logging else "구조화 로깅 필요"
        )

        # Versioning
        versioning = constitution.get("versioning", {})
        version = versioning.get("current_version", "0.0.0")
        version_icon = "✅"
        table.add_row(
            "Versioning",
            version_icon,
            f"{version} (MAJOR.MINOR.BUILD 준수)"
        )

        return table

    def render_git_status(self, git_info: Dict) -> Panel:
        """Git 상태 정보 렌더링"""
        if not git_info["is_git_repo"]:
            content = Text("Git 저장소가 아닙니다", style="red")
            return Panel(content, title="🔀 Git 상태", box=box.ROUNDED)

        content = []

        # 현재 브랜치
        branch = git_info["current_branch"]
        content.append(f"├── 현재 브랜치: {branch}")

        # 최근 커밋
        last_commit = git_info["last_commit"]
        if last_commit["hash"]:
            content.append(f"├── 최근 커밋: {last_commit['hash']} - {last_commit['message'][:50]}... ({last_commit['date']})")

        # 변경사항
        status = git_info["status"]
        total_changes = sum(status.values())
        if total_changes > 0:
            changes = []
            if status["modified"] > 0:
                changes.append(f"수정 {status['modified']}개")
            if status["added"] > 0:
                changes.append(f"추가 {status['added']}개")
            if status["deleted"] > 0:
                changes.append(f"삭제 {status['deleted']}개")
            if status["untracked"] > 0:
                changes.append(f"미추적 {status['untracked']}개")
            content.append(f"├── 변경사항: {', '.join(changes)}")
            content.append("└── 작업 상태: 🟡 진행 중 (커밋 필요)")
        else:
            content.append("└── 작업 상태: 🟢 깨끗함")

        return Panel(
            "\n".join(content),
            title="🔀 Git 상태",
            box=box.ROUNDED
        )

    def generate_recommendations(self, state_data: Dict, git_info: Dict, specs: List[Dict]) -> Panel:
        """추천 액션 생성"""
        recommendations = []
        warnings = []

        # 다음 단계 추천
        if not specs:
            recommendations.append("🚀 /moai:2-spec all  # 첫 번째 SPEC 작성")
        else:
            incomplete_specs = [s for s in specs if s["progress"] < 100]
            if incomplete_specs:
                next_spec = incomplete_specs[0]["id"]
                if not (self.project_root / ".moai" / "specs" / next_spec / "plan.md").exists():
                    recommendations.append(f"🚀 /moai:3-plan {next_spec}  # Constitution 검증 및 계획 수립")
                elif not (self.project_root / ".moai" / "specs" / next_spec / "tasks.md").exists():
                    recommendations.append(f"📋 /moai:4-tasks {next_spec}  # 작업 분해")
                else:
                    recommendations.append(f"⚡ /moai:5-dev {next_spec}  # 구현 시작")

        # 테스트 커버리지 확인
        testing = state_data.get("constitution_compliance", {}).get("testing", {})
        coverage = testing.get("current_coverage", 0.0)
        target = testing.get("coverage_target", 0.8)
        if coverage < target:
            recommendations.append("📊 pytest --cov=80  # 테스트 커버리지 향상")

        # Git 상태 확인
        if git_info["is_git_repo"]:
            total_changes = sum(git_info["status"].values())
            if total_changes > 0:
                recommendations.append("🔄 git add . && git commit  # 변경사항 커밋")

            # 미추적 파일 경고
            if git_info["status"]["untracked"] > 0:
                warnings.append(f"미추적 파일 {git_info['status']['untracked']}개 확인 필요")

        # Constitution 위반 경고
        if coverage < target:
            warnings.append("Testing Constitution 위반 해결 권장")

        # 문서 동기화 추천
        recommendations.append("📝 /moai:6-sync auto  # 문서 동기화")

        content = []
        content.append("💡 다음 단계 추천")
        for i, rec in enumerate(recommendations[:4], 1):
            content.append(f"{i}. {rec}")

        if warnings:
            content.append("")
            content.append("⚠️ 주의사항")
            for warning in warnings:
                content.append(f"- {warning}")

        return Panel(
            "\n".join(content),
            title="🎯 추천 액션",
            box=box.ROUNDED,
            style="green"
        )

    def render_dashboard(self, detail: bool = False) -> None:
        """전체 대시보드 렌더링"""
        state_data = self.load_state_data()
        git_info = self.load_git_info()
        specs = self.analyze_specs()
        stages = self.calculate_pipeline_progress(state_data)

        # 헤더
        self.console.print(self.render_header(state_data))
        self.console.print()

        # 파이프라인 진행률
        self.console.print(self.render_pipeline_progress(stages))
        self.console.print()

        # 두 열 레이아웃
        left_column = []
        right_column = []

        # SPEC 현황 (왼쪽)
        left_column.append(self.render_specs_table(specs))

        # TAG 시스템 (오른쪽)
        right_column.append(Panel(self.render_tag_system(state_data), box=box.ROUNDED))

        # Constitution 상태 (전체 너비)
        self.console.print(Columns([left_column[0], right_column[0]], equal=True))
        self.console.print()

        self.console.print(self.render_constitution_status(state_data))
        self.console.print()

        # Git 상태와 추천 액션 (두 열)
        git_panel = self.render_git_status(git_info)
        recommendations_panel = self.generate_recommendations(state_data, git_info, specs)

        self.console.print(Columns([git_panel, recommendations_panel], equal=True))

        if detail:
            self.console.print()
            self.console.print(Panel(
                "상세 모드: 추가 정보는 향후 버전에서 제공됩니다.",
                title="🔍 상세 정보",
                style="blue"
            ))

    def export_dashboard(self, output_path: Optional[str] = None) -> str:
        """대시보드를 파일로 내보내기"""
        if output_path is None:
            timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
            output_path = f"dashboard_{timestamp}.html"

        # 간단한 HTML 내보내기 (Rich의 HTML 내보내기 기능 사용)
        state_data = self.load_state_data()
        project_name = state_data.get("metadata", {}).get("project_name", "MoAI-ADK")

        html_content = f"""
        <!DOCTYPE html>
        <html>
        <head>
            <title>{project_name} Dashboard</title>
            <meta charset="utf-8">
            <style>
                body {{ font-family: 'Consolas', monospace; background: #1e1e1e; color: #fff; }}
                .container {{ max-width: 1200px; margin: 0 auto; padding: 20px; }}
                .header {{ text-align: center; border: 2px solid #0078d4; padding: 20px; margin-bottom: 20px; }}
                .section {{ margin: 20px 0; padding: 15px; border: 1px solid #333; }}
            </style>
        </head>
        <body>
            <div class="container">
                <div class="header">
                    <h1>🗿 {project_name} Dashboard</h1>
                    <p>Generated on {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}</p>
                </div>
                <div class="section">
                    <h2>📊 프로젝트 상태</h2>
                    <p>이 대시보드는 HTML 형태로 내보내졌습니다.</p>
                    <p>자세한 정보는 터미널에서 /moai:7-dashboard 명령어를 실행하세요.</p>
                </div>
            </div>
        </body>
        </html>
        """

        output_file = Path(output_path)
        with open(output_file, 'w', encoding='utf-8') as f:
            f.write(html_content)

        return str(output_file.absolute())


def main():
    """메인 실행 함수"""
    parser = argparse.ArgumentParser(description="MoAI-ADK Dashboard")
    parser.add_argument("--detail", action="store_true", help="상세 정보 표시")
    parser.add_argument("--export", type=str, nargs="?", const="", help="파일로 내보내기")
    parser.add_argument("--project-root", type=str, help="프로젝트 루트 경로")

    args = parser.parse_args()

    # 프로젝트 루트 결정
    if args.project_root:
        project_root = Path(args.project_root)
    else:
        project_root = Path.cwd()
        # .moai 디렉토리를 찾을 때까지 상위로 이동
        while not (project_root / ".moai").exists() and project_root.parent != project_root:
            project_root = project_root.parent

    # MoAI 프로젝트 확인
    if not (project_root / ".moai").exists():
        console = Console()
        console.print(Panel(
            "❌ MoAI 프로젝트가 아닙니다.\n\n"
            "다음 명령어로 프로젝트를 초기화하세요:\n"
            "> /moai:1-project",
            title="오류",
            style="red"
        ))
        sys.exit(1)

    # 대시보드 생성 및 렌더링
    dashboard = MoAIDashboard(project_root)

    if args.export is not None:
        # 내보내기 모드
        output_path = args.export if args.export else None
        exported_file = dashboard.export_dashboard(output_path)
        console = Console()
        console.print(f"✅ 대시보드를 내보냈습니다: {exported_file}")
    else:
        # 일반 표시 모드
        dashboard.render_dashboard(detail=args.detail)


if __name__ == "__main__":
    main()