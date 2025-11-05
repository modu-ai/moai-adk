# @CODE:USER-EXPERIENCE-CLI-001 | @TEST:USER-EXPERIENCE-CLI-001

"""
User Experience Improvement CLI Command
사용자 경험 개선 CLI 명령어
"""

import argparse
import asyncio
from pathlib import Path

from rich.console import Console
from rich.panel import Panel
from rich.progress import Progress, SpinnerColumn, TextColumn
from rich.table import Table

from moai_adk.utils.user_experience import (
    UserExperienceAnalyzer,
    generate_improvement_plan,
)

console = Console()


def create_parser(subparsers) -> argparse.ArgumentParser:
    """사용자 경험 개선 파서 생성"""
    parser = subparsers.add_parser(
        'improve-ux',
        help='사용자 경험 개선 분석',
        description='온라인 문서 포털의 사용자 경험을 분석하고 개선 계획을 제공합니다'
    )

    parser.add_argument(
        '--url', '-u',
        type=str,
        default='https://adk.mo.ai.kr',
        help='분석할 URL (기본값: https://adk.mo.ai.kr)'
    )

    parser.add_argument(
        '--output', '-o',
        type=str,
        help='분석 결과를 저장할 파일 경로'
    )

    parser.add_argument(
        '--format', '-f',
        type=str,
        choices=['json', 'markdown', 'text'],
        default='markdown',
        help='출력 형식 (기본값: markdown)'
    )

    parser.add_argument(
        '--verbose', '-v',
        action='store_true',
        help='상세한 진행 상황 표시'
    )

    parser.add_argument(
        '--max-workers', '-w',
        type=int,
        default=5,
        help='동시에 처리할 최대 작업 수 (기본값: 5)'
    )

    return parser


def format_metrics_table(metrics: dict) -> Table:
    """지표 테이블 형식화"""
    table = Table(title="사용자 경험 지표")
    table.add_column("분야", style="cyan")
    table.add_column("점수", style="magenta")
    table.add_column("상태", style="green")

    # 성능 지표
    perf = metrics.get('performance')
    table.add_row("성능", f"{perf.success_rate if perf else 0:.2f}",
                  "✅ 좋음" if (perf and perf.is_good) else "❌ 개선 필요")

    # 네비게이션 지표
    nav = metrics.get('navigation')
    table.add_row("네비게이션", f"{nav.structure_score if nav else 0:.2f}",
                  "✅ 좋음" if (nav and nav.is_good) else "❌ 개선 필요")

    # 콘텐츠 지표
    content = metrics.get('content')
    table.add_row("콘텐츠", f"{content.accuracy_score if content else 0:.2f}",
                  "✅ 좋음" if (content and content.is_good) else "❌ 개선 필요")

    # 접근성 지표
    acc = metrics.get('accessibility')
    table.add_row("접근성", f"{1.0 if (acc and acc.is_good) else 0.0:.2f}",
                  "✅ 좋음" if (acc and acc.is_good) else "❌ 개선 필요")

    return table


def format_recommendations(recommendations: list) -> Table:
    """개선 제안 테이블 형식화"""
    table = Table(title="개선 제안")
    table.add_column("제안", style="yellow")
    table.add_column("우선순위", style="red")

    for i, rec in enumerate(recommendations, 1):
        priority = "중간"  # 기본값
        if "에러" in rec or "키보드" in rec or "정확성" in rec:
            priority = "높음"
        elif "성능" in rec or "구조" in rec:
            priority = "중간"
        else:
            priority = "낮음"

        table.add_row(f"{i}. {rec}", priority)

    return table


def format_timeline(timeline: dict) -> Table:
    """실행 계획 테이블 형식화"""
    table = Table(title="실행 계획")
    table.add_column("기간", style="cyan")
    table.add_column("작업", style="magenta")

    for period, tasks in timeline.items():
        if tasks:
            table.add_row(
                period.replace('_', ' ').title(),
                '\n'.join(f"• {task}" for task in tasks)
            )

    return table


async def analyze_user_experience(url: str, max_workers: int = 5, verbose: bool = False) -> dict:
    """사용자 경험 분석 수행"""
    if verbose:
        console.print(f"[blue]사용자 경험 분석 시작: {url}[/blue]")

    analyzer = UserExperienceAnalyzer(url, max_workers=max_workers)

    with Progress(
        SpinnerColumn(),
        TextColumn("[progress.description]{task.description}"),
        console=console,
        transient=True
    ) as progress:
        # 분석 작업 생성
        analysis_task = progress.add_task("사용자 경험 분석 중...", total=None)

        # 분석 실행
        analysis_report = await analyzer.generate_report()

        progress.update(analysis_task, completed=True)

    return analysis_report


def generate_markdown_report(analysis_report: dict, improvement_plan: dict) -> str:
    """마크다운 보고서 생성"""
    report = []

    # 헤더
    report.append("# 사용자 경험 개선 보고서")
    report.append("")
    report.append(f"**분석 대상**: {analysis_report.get('base_url', 'N/A')}")
    report.append(f"**분석 시간**: {analysis_report.get('generated_at', 'N/A')}")
    report.append("")

    # 전체 점수
    overall_score = analysis_report.get('overall_score', 0)
    report.append(f"## 전체 점수: {overall_score:.2f}")
    report.append("")

    # 지표 상세
    report.append("### 지표 상세")
    report.append("")

    # 성능 지표
    performance = analysis_report.get('performance')
    report.append("#### 성능")
    report.append(f"- 성공률: {performance.success_rate if performance else 0:.2f}")
    report.append(f"- 평균 로드 시간: {performance.load_time if performance else 0:.2f}초")
    report.append(f"- 응답 시간: {performance.response_time if performance else 0:.2f}초")
    report.append(f"- 에러율: {performance.error_rate if performance else 0:.2f}")
    report.append("")

    # 네비게이션 지표
    navigation = analysis_report.get('navigation')
    report.append("#### 네비게이션")
    report.append(f"- 구조 점수: {navigation.structure_score if navigation else 0:.2f}")
    report.append(f"- 링크 수: {navigation.link_count if navigation else 0}")
    report.append(f"- 깊이: {navigation.depth if navigation else 0}")
    report.append(f"- 완성도: {navigation.completeness if navigation else 0:.2f}")
    report.append("")

    # 콘텐츠 지표
    content = analysis_report.get('content')
    report.append("#### 콘텐츠")
    report.append(f"- 정확성: {content.accuracy_score if content else 0:.2f}")
    report.append(f"- 완성도: {content.completeness_score if content else 0:.2f}")
    report.append(f"- 조직성: {content.organization_score if content else 0:.2f}")
    report.append(f"- 가독성: {content.readability_score if content else 0:.2f}")
    report.append("")

    # 접근성 지표
    accessibility = analysis_report.get('accessibility')
    report.append("#### 접근성")
    report.append(f"- 키보드 네비게이션: {'✅' if (accessibility and accessibility.keyboard_navigation) else '❌'}")
    report.append(f"- 스크린 리더 지원: {'✅' if (accessibility and accessibility.screen_reader_support) else '❌'}")
    report.append(f"- 색상 대비: {'✅' if (accessibility and accessibility.color_contrast) else '❌'}")
    report.append(f"- 반응형 디자인: {'✅' if (accessibility and accessibility.responsive_design) else '❌'}")
    report.append(f"- ARIA 레이블: {'✅' if (accessibility and accessibility.aria_labels) else '❌'}")
    report.append("")

    # 개선 제안
    report.append("### 개선 제안")
    report.append("")
    for rec in analysis_report.get('recommendations', []):
        report.append(f"- {rec}")
    report.append("")

    # 실행 계획
    report.append("### 실행 계획")
    report.append("")
    report.append(f"**예상 소요 시간**: {improvement_plan.get('estimated_duration', 'N/A')}")
    report.append("")

    timeline = improvement_plan.get('timeline', {})
    for period, tasks in timeline.items():
        if tasks:
            report.append(f"#### {period.replace('_', ' ').title()}")
            for task in tasks:
                report.append(f"- {task}")
            report.append("")

    return "\n".join(report)


def run_command(args) -> int:
    """사용자 경험 개선 명령 실행"""
    try:
        # 분석 실행
        analysis_report = asyncio.run(
            analyze_user_experience(args.url, args.max_workers, args.verbose)
        )

        # 개선 계획 생성
        improvement_plan = generate_improvement_plan(analysis_report)

        # 결과 출력
        if args.verbose:
            console.print(Panel.fit(
                f"전체 점수: {analysis_report['overall_score']:.2f}",
                title="분석 결과",
                style="blue"
            ))

        # 지표 테이블 표시
        metrics_table = format_metrics_table(analysis_report)
        console.print(metrics_table)

        # 개선 제안 표시
        if analysis_report.get('recommendations'):
            recommendations_table = format_recommendations(analysis_report['recommendations'])
            console.print(recommendations_table)

        # 실행 계획 표시
        if improvement_plan.get('timeline'):
            timeline_table = format_timeline(improvement_plan['timeline'])
            console.print(timeline_table)

        # 결과 저장
        if args.output:
            if args.format == 'json':
                import json
                result = {
                    "analysis": analysis_report,
                    "improvement_plan": improvement_plan
                }
                output_content = json.dumps(result, indent=2, ensure_ascii=False)
            else:
                output_content = generate_markdown_report(analysis_report, improvement_plan)

            output_path = Path(args.output)
            output_path.write_text(output_content, encoding='utf-8')
            console.print(f"\n[green]결과가 저장되었습니다:[/green] {output_path}")

        # 종료 코드 반환
        if analysis_report['overall_score'] >= 0.85:
            console.print(
            f"\n[green]✅ 사용자 경험이 우수합니다 (점수: {analysis_report['overall_score']:.2f})[/green]"
        )
            return 0
        else:
            console.print(
                f"\n[yellow]⚠️  사용자 경험이 개선 필요합니다 (점수: {analysis_report['overall_score']:.2f})[/yellow]"
            )
            return 1

    except KeyboardInterrupt:
        console.print("\n[yellow]분석이 사용자에 의해 중단되었습니다.[/yellow]")
        return 1
    except Exception as e:
        console.print(f"[red]오류 발생:[/red] {e}")
        return 1
