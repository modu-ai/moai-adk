# @CODE:LINK-VALIDATION-CLI-001 | @TEST:LINK-VALIDATION-CLI-001

"""
Link Validation CLI Command
링크 검증 CLI 명령어
"""

import argparse
import asyncio
from pathlib import Path

from moai_adk.utils.link_validator import LinkValidator


def create_parser(subparsers) -> argparse.ArgumentParser:
    """링크 검증 파서 생성"""
    parser = subparsers.add_parser(
        'validate-links',
        help='온라인 문서 링크 검증',
        description='README.ko.md에 있는 모든 온라인 문서 링크를 자동으로 검증합니다'
    )

    parser.add_argument(
        '--file', '-f',
        type=str,
        default='README.ko.md',
        help='검증할 파일 경로 (기본값: README.ko.md)'
    )

    parser.add_argument(
        '--max-concurrent', '-c',
        type=int,
        default=3,
        help='동시에 검증할 최대 링크 수 (기본값: 3)'
    )

    parser.add_argument(
        '--timeout', '-t',
        type=int,
        default=8,
        help='요청 타임아웃 (초) (기본값: 8)'
    )

    parser.add_argument(
        '--output', '-o',
        type=str,
        help='결과를 저장할 파일 경로'
    )

    parser.add_argument(
        '--verbose', '-v',
        action='store_true',
        help='상세한 진행 상황 표시'
    )

    return parser


def run_command(args) -> int:
    """링크 검증 명령 실행"""
    try:
        # 파일 경로 설정
        file_path = Path(args.file)
        if not file_path.exists():
            print(f"오류: 파일이 존재하지 않습니다: {file_path}")
            return 1

        # 검증기 생성
        validator = LinkValidator(
            max_concurrent=args.max_concurrent,
            timeout=args.timeout
        )

        if args.verbose:
            print(f"파일에서 링크 추출 중: {file_path}")

        # 링크 추출
        links = validator.extract_links_from_file(file_path)

        if not links:
            print("검증할 링크가 없습니다.")
            return 0

        if args.verbose:
            print(f"총 {len(links)}개의 링크를 검증합니다...")

        # 비동기 검증 실행
        async def validate_links():
            async with validator:
                result = await validator.validate_all_links(links)
                return result

        result = asyncio.run(validate_links())

        # 결과 생성
        report = validator.generate_report(result)

        # 출력
        print(report)

        # 파일 저장
        if args.output:
            output_path = Path(args.output)
            output_path.write_text(report, encoding='utf-8')
            print(f"\n결과가 저장되었습니다: {output_path}")

        # 종료 코드 반환
        if result.invalid_links > 0:
            print(f"\n⚠️  {result.invalid_links}개의 링크가 검증에 실패했습니다.")
            return 1
        else:
            print("\n✅ 모든 링크가 정상적으로 검증되었습니다.")
            return 0

    except KeyboardInterrupt:
        print("\n검증이 사용자에 의해 중단되었습니다.")
        return 1
    except Exception as e:
        print(f"오류 발생: {e}")
        return 1
