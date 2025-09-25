#!/usr/bin/env python3
"""
@FEATURE:SQLITE-MIGRATION-CLI-001 - SQLite 마이그레이션 CLI 도구

SPEC-009 구현체를 위한 사용자 친화적 마이그레이션 인터페이스
"""

import json
import sys
import time
from pathlib import Path
from typing import Optional

import click

from ..core.tag_system.adapter import TagIndexAdapter
from ..core.tag_system.migration import TagMigrationTool
from ..core.tag_system.database import TagDatabaseManager


@click.group()
def sqlite_migration():
    """SQLite 기반 TAG 인덱싱 마이그레이션 도구"""
    pass


@sqlite_migration.command()
@click.option(
    "--config-path",
    "-c",
    type=click.Path(exists=True, path_type=Path),
    default=".moai/config.json",
    help="MoAI 설정 파일 경로",
)
@click.option("--dry-run", "-n", is_flag=True, help="실제 변경 없이 시뮬레이션만 실행")
@click.option("--force", "-f", is_flag=True, help="백업 없이 강제 마이그레이션")
def migrate(config_path: Path, dry_run: bool, force: bool):
    """JSON에서 SQLite로 TAG 인덱스 마이그레이션"""

    click.echo("🚀 MoAI-ADK SQLite 마이그레이션 시작")
    click.echo("=" * 50)

    try:
        # 설정 로드
        config = load_moai_config(config_path)
        tags_config = config.get("tags", {})
        backend_config = tags_config.get("backend", {})

        # 경로 설정
        json_path = Path(tags_config.get("index_path", ".moai/indexes/tags.json"))
        sqlite_path = Path(
            backend_config.get("sqlite", {}).get(
                "database_path", ".moai/indexes/tags.db"
            )
        )

        if not json_path.exists():
            click.echo(f"❌ JSON 인덱스 파일을 찾을 수 없습니다: {json_path}")
            sys.exit(1)

        # 파일 크기 확인
        json_size = json_path.stat().st_size
        click.echo(
            f"📊 JSON 파일 크기: {json_size:,} bytes ({json_size / 1024 / 1024:.2f} MB)"
        )

        if dry_run:
            click.echo("🔍 DRY-RUN 모드: 실제 변경 없이 시뮬레이션")
            estimate_migration_performance(json_path)
            return

        # 백업 생성
        if not force and backend_config.get("sqlite", {}).get("migration", {}).get(
            "backup_before_migration", True
        ):
            backup_path = create_backup(json_path)
            click.echo(f"💾 백업 생성: {backup_path}")

        # 마이그레이션 실행
        click.echo("⚡ 마이그레이션 진행 중...")
        start_time = time.perf_counter()

        migration_tool = TagMigrationTool(
            database_path=sqlite_path,
            json_path=json_path,
            backup_directory=json_path.parent / "backups",
        )

        def progress_callback(progress):
            """진행률 콜백"""
            click.echo(
                f"📈 진행률: {progress.percentage:.1f}% ({progress.processed}/{progress.total})"
            )

        result = migration_tool.migrate_json_to_sqlite(
            validate_data=True,
            create_backup=not force,
            progress_callback=progress_callback,
            detailed_reporting=True,
        )

        end_time = time.perf_counter()
        migration_time = end_time - start_time

        if result.success:
            click.echo("✅ 마이그레이션 성공!")
            click.echo(f"⏱️  소요 시간: {migration_time:.2f}초")
            click.echo(f"📝 마이그레이션된 TAG: {result.tags_migrated}개")
            click.echo(f"📍 참조 수: {result.references_migrated}개")

            # 성능 비교
            show_performance_comparison(json_path, sqlite_path)

            # 설정 업데이트 제안
            if click.confirm("🔄 SQLite 백엔드를 활성화하시겠습니까?"):
                enable_sqlite_backend(config_path)

        else:
            click.echo("❌ 마이그레이션 실패!")
            for error in result.errors:
                click.echo(f"   - {error}")
            sys.exit(1)

    except Exception as e:
        click.echo(f"💥 예상치 못한 오류: {e}")
        sys.exit(1)


@sqlite_migration.command()
@click.option(
    "--config-path",
    "-c",
    type=click.Path(exists=True, path_type=Path),
    default=".moai/config.json",
    help="MoAI 설정 파일 경로",
)
def rollback(config_path: Path):
    """SQLite에서 JSON으로 롤백"""

    click.echo("🔄 SQLite → JSON 롤백 시작")
    click.echo("=" * 50)

    try:
        config = load_moai_config(config_path)
        tags_config = config.get("tags", {})
        backend_config = tags_config.get("backend", {})

        json_path = Path(tags_config.get("index_path", ".moai/indexes/tags.json"))
        sqlite_path = Path(
            backend_config.get("sqlite", {}).get(
                "database_path", ".moai/indexes/tags.db"
            )
        )

        if not sqlite_path.exists():
            click.echo(f"❌ SQLite 데이터베이스를 찾을 수 없습니다: {sqlite_path}")
            sys.exit(1)

        click.echo("⚡ 롤백 진행 중...")
        start_time = time.perf_counter()

        migration_tool = TagMigrationTool(
            database_path=sqlite_path, json_path=json_path
        )

        result = migration_tool.migrate_sqlite_to_json()

        end_time = time.perf_counter()
        rollback_time = end_time - start_time

        if result.success:
            click.echo("✅ 롤백 성공!")
            click.echo(f"⏱️  소요 시간: {rollback_time:.2f}초")

            # 설정 업데이트
            disable_sqlite_backend(config_path)

        else:
            click.echo("❌ 롤백 실패!")
            for error in result.errors:
                click.echo(f"   - {error}")
            sys.exit(1)

    except Exception as e:
        click.echo(f"💥 예상치 못한 오류: {e}")
        sys.exit(1)


@sqlite_migration.command()
@click.option(
    "--config-path",
    "-c",
    type=click.Path(exists=True, path_type=Path),
    default=".moai/config.json",
    help="MoAI 설정 파일 경로",
)
@click.option("--iterations", "-i", default=10, help="성능 테스트 반복 횟수")
def benchmark(config_path: Path, iterations: int):
    """JSON vs SQLite 성능 벤치마크"""

    click.echo("🏃 성능 벤치마크 시작")
    click.echo("=" * 50)

    try:
        config = load_moai_config(config_path)
        tags_config = config.get("tags", {})

        json_path = Path(tags_config.get("index_path", ".moai/indexes/tags.json"))

        if not json_path.exists():
            click.echo(f"❌ JSON 인덱스 파일을 찾을 수 없습니다: {json_path}")
            sys.exit(1)

        # 임시 SQLite 생성 (벤치마크용)
        temp_sqlite = json_path.parent / "benchmark_test.db"

        click.echo("📊 벤치마크 실행 중...")

        # 마이그레이션 (벤치마크용)
        migration_tool = TagMigrationTool(
            database_path=temp_sqlite, json_path=json_path
        )

        migration_result = migration_tool.migrate_json_to_sqlite()
        if not migration_result.success:
            click.echo("❌ 벤치마크용 마이그레이션 실패")
            sys.exit(1)

        # 성능 테스트 실행
        json_times = []
        sqlite_times = []

        # JSON 성능 측정
        with open(json_path, "r") as f:
            json_data = json.load(f)

        for i in range(iterations):
            start = time.perf_counter()
            # 샘플 검색 작업
            search_in_json(json_data)
            end = time.perf_counter()
            json_times.append((end - start) * 1000)  # ms

        # SQLite 성능 측정
        db_manager = TagDatabaseManager(temp_sqlite)

        for i in range(iterations):
            start = time.perf_counter()
            # 동일한 검색 작업
            search_in_sqlite(db_manager)
            end = time.perf_counter()
            sqlite_times.append((end - start) * 1000)  # ms

        # 결과 표시
        json_avg = sum(json_times) / len(json_times)
        sqlite_avg = sum(sqlite_times) / len(sqlite_times)
        speedup = json_avg / sqlite_avg

        click.echo("📈 벤치마크 결과:")
        click.echo(f"   JSON 평균:    {json_avg:.2f}ms")
        click.echo(f"   SQLite 평균:  {sqlite_avg:.2f}ms")
        click.echo(f"   성능 향상:    {speedup:.1f}x 빠름")

        # 메모리 사용량 비교
        json_memory = estimate_json_memory(json_data)
        sqlite_memory = estimate_sqlite_memory(temp_sqlite)
        memory_savings = (json_memory - sqlite_memory) / json_memory

        click.echo(f"   JSON 메모리:  {json_memory / 1024 / 1024:.1f}MB")
        click.echo(f"   SQLite 메모리: {sqlite_memory / 1024 / 1024:.1f}MB")
        click.echo(f"   메모리 절약:  {memory_savings:.1%}")

        # 정리
        temp_sqlite.unlink()

        if speedup >= 10.0:
            click.echo("🎉 SPEC-009 성능 목표(10배) 달성!")
        else:
            click.echo("⚠️  성능 목표 미달성. 추가 최적화 필요")

    except Exception as e:
        click.echo(f"💥 벤치마크 오류: {e}")
        sys.exit(1)


def load_moai_config(config_path: Path) -> dict:
    """MoAI 설정 로드"""
    with open(config_path, "r") as f:
        return json.load(f)


def create_backup(json_path: Path) -> Path:
    """JSON 파일 백업 생성"""
    timestamp = time.strftime("%Y%m%d_%H%M%S")
    backup_path = json_path.parent / "backups" / f"tags_backup_{timestamp}.json"
    backup_path.parent.mkdir(exist_ok=True)

    import shutil

    shutil.copy2(json_path, backup_path)
    return backup_path


def estimate_migration_performance(json_path: Path):
    """마이그레이션 성능 추정"""
    with open(json_path, "r") as f:
        data = json.load(f)

    total_tags = data.get("statistics", {}).get("total_tags", 0)
    estimated_time = total_tags * 0.01  # 가정: TAG당 10ms

    click.echo(f"📊 예상 마이그레이션 시간: {estimated_time:.1f}초")
    click.echo(f"📝 마이그레이션할 TAG: {total_tags}개")


def show_performance_comparison(json_path: Path, sqlite_path: Path):
    """성능 비교 표시"""
    json_size = json_path.stat().st_size
    sqlite_size = sqlite_path.stat().st_size
    size_reduction = (json_size - sqlite_size) / json_size

    click.echo("📊 파일 크기 비교:")
    click.echo(f"   JSON:    {json_size:,} bytes")
    click.echo(f"   SQLite:  {sqlite_size:,} bytes")
    click.echo(f"   절약:    {size_reduction:.1%}")


def enable_sqlite_backend(config_path: Path):
    """SQLite 백엔드 활성화"""
    with open(config_path, "r") as f:
        config = json.load(f)

    config["tags"]["backend"]["type"] = "sqlite"
    config["tags"]["backend"]["sqlite"]["enabled"] = True

    with open(config_path, "w") as f:
        json.dump(config, f, indent=2, ensure_ascii=False)

    click.echo("✅ SQLite 백엔드 활성화됨")


def disable_sqlite_backend(config_path: Path):
    """SQLite 백엔드 비활성화"""
    with open(config_path, "r") as f:
        config = json.load(f)

    config["tags"]["backend"]["type"] = "json"
    config["tags"]["backend"]["sqlite"]["enabled"] = False

    with open(config_path, "w") as f:
        json.dump(config, f, indent=2, ensure_ascii=False)

    click.echo("✅ JSON 백엔드로 복원됨")


def search_in_json(json_data: dict) -> list:
    """JSON에서 샘플 검색"""
    # 샘플 TAG 검색 시뮬레이션
    index = json_data.get("index", {})
    return [key for key in index.keys() if "REQ" in key][:10]


def search_in_sqlite(db_manager) -> list:
    """SQLite에서 샘플 검색"""
    # 동일한 검색을 SQLite로 수행
    try:
        results = db_manager.search_by_category("REQ", limit=10)
        return [r["tag_key"] for r in results]
    except:
        return []  # 에러 시 빈 결과


def estimate_json_memory(json_data: dict) -> int:
    """JSON 메모리 사용량 추정"""
    import sys

    return sys.getsizeof(str(json_data))


def estimate_sqlite_memory(sqlite_path: Path) -> int:
    """SQLite 메모리 사용량 추정 (파일 크기 기반)"""
    return sqlite_path.stat().st_size


if __name__ == "__main__":
    sqlite_migration()
