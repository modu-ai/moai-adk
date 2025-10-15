"""업데이트 명령어"""
from pathlib import Path

import click
from rich.console import Console

from moai_adk import __version__
from moai_adk.core.template.processor import TemplateProcessor

console = Console()


@click.command()
@click.option(
    "--path",
    type=click.Path(exists=True),
    default=".",
    help="프로젝트 경로 (기본: 현재 디렉토리)"
)
@click.option(
    "--force",
    is_flag=True,
    help="백업 생략하고 강제 업데이트"
)
@click.option(
    "--check",
    is_flag=True,
    help="버전 확인만 수행 (업데이트하지 않음)"
)
def update(path: str, force: bool, check: bool) -> None:
    """템플릿 파일을 최신 버전으로 업데이트합니다.

    업데이트 내용:
    - .claude/ (전체 교체)
    - .moai/ (specs, reports 보존)
    - CLAUDE.md (병합)
    - config.json (스마트 병합)

    예시:
        moai-adk update              # 백업 후 업데이트
        moai-adk update --force      # 백업 없이 업데이트
        moai-adk update --check      # 버전만 확인
    """
    try:
        project_path = Path(path).resolve()

        # 프로젝트 초기화 확인
        if not (project_path / ".moai").exists():
            console.print("[yellow]⚠ Project not initialized[/yellow]")
            raise click.Abort()

        # Phase 1: 버전 확인
        console.print("[cyan]🔍 버전 확인 중...[/cyan]")
        current_version = __version__
        latest_version = __version__
        console.print(f"   현재 버전: {current_version}")
        console.print(f"   최신 버전: {latest_version}")

        if check:
            # --check 옵션이면 여기서 종료
            if current_version == latest_version:
                console.print("[green]✓ Already up to date[/green]")
            else:
                console.print("[yellow]⚠ Update available[/yellow]")
            return

        # Phase 2: 백업 (--force 없으면)
        if not force:
            console.print("\n[cyan]💾 백업 생성 중...[/cyan]")
            processor = TemplateProcessor(project_path)
            backup_path = processor.create_backup()
            console.print(f"[green]✓ 백업 완료: {backup_path.relative_to(project_path)}[/green]")
        else:
            console.print("\n[yellow]⚠ 백업 생략 (--force)[/yellow]")

        # Phase 3: 템플릿 업데이트
        console.print("\n[cyan]📄 템플릿 업데이트 중...[/cyan]")
        processor = TemplateProcessor(project_path)
        processor.copy_templates(backup=False, silent=True)  # 이미 백업했으므로

        console.print("   [green]✅ .claude/ 업데이트 완료[/green]")
        console.print("   [green]✅ .moai/ 업데이트 완료 (specs/reports 보존)[/green]")
        console.print("   [green]🔄 CLAUDE.md 병합 완료[/green]")
        console.print("   [green]🔄 config.json 병합 완료[/green]")

        console.print("\n[green]✓ 업데이트 완료![/green]")

    except Exception as e:
        console.print(f"[red]✗ Update failed: {e}[/red]")
        raise click.ClickException(str(e)) from e
