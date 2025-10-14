"""백업 명령어"""
import click
from pathlib import Path
from rich.console import Console
from moai_adk.core.template.processor import TemplateProcessor

console = Console()


@click.command()
@click.option(
    "--path",
    type=click.Path(exists=True),
    default=".",
    help="프로젝트 경로 (기본: 현재 디렉토리)"
)
def backup(path: str) -> None:
    """현재 프로젝트를 백업합니다.

    백업 내용:
    - .claude/ (전체)
    - .moai/ (specs, reports 제외)
    - CLAUDE.md

    백업 위치: .moai-backup/YYYYMMDD-HHMMSS/
    """
    try:
        project_path = Path(path).resolve()

        # 프로젝트 초기화 확인
        if not (project_path / ".moai").exists():
            console.print("[yellow]⚠ Project not initialized[/yellow]")
            raise click.Abort()

        # 백업 생성
        console.print("[cyan]💾 백업 생성 중...[/cyan]")
        processor = TemplateProcessor(project_path)
        backup_path = processor.create_backup()

        # 성공 메시지
        console.print(f"[green]✓ 백업 완료: {backup_path.relative_to(project_path)}[/green]")

        # 백업 내용 표시
        backup_items = list(backup_path.iterdir())
        for item in backup_items:
            if item.is_dir():
                file_count = len(list(item.rglob("*")))
                console.print(f"   ├─ {item.name}/ ({file_count}개 파일)")
            else:
                console.print(f"   └─ {item.name}")

    except Exception as e:
        console.print(f"[red]✗ Backup failed: {e}[/red]")
        raise click.ClickException(str(e)) from e
