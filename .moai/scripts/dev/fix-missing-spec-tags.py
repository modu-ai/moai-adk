#!/usr/bin/env python3
# @CODE:SPEC-TAG-FIXER-001 | @SPEC:TAG-COVERAGE-ENFORCEMENT-001
"""
SPEC 파일에서 누락된 @SPEC: 태그를 자동으로 추가하는 스크립트

이 스크립트는 .moai/specs/ 디렉토리의 모든 파일을 검사하여:
1. plan.md, acceptance.md에 누락된 @SPEC: 태그 추가
2. 제목 라인(# SPEC-XXX-...)에서 @SPEC:{ID} 형식으로 변환
3. 변경 내용을 기록

사용법:
    python3 .moai/scripts/fix-missing-spec-tags.py
    python3 .moai/scripts/fix-missing-spec-tags.py --dry-run
    python3 .moai/scripts/fix-missing-spec-tags.py --spec-id SPEC-CLI-ANALYSIS-001
"""

import re
import sys
from pathlib import Path
from typing import Dict, List, Tuple

def extract_spec_id_from_path(filepath: str) -> str:
    """
    파일 경로에서 SPEC ID 추출

    예: .moai/specs/SPEC-CLI-ANALYSIS-001/plan.md → SPEC-CLI-ANALYSIS-001
    """
    match = re.search(r'(SPEC-[A-Z0-9-]+)', filepath)
    if match:
        return match.group(1)
    return None

def extract_spec_id_from_title(content: str) -> str:
    """
    파일 첫 줄의 제목에서 SPEC ID 추출

    예: # SPEC-CLI-ANALYSIS-001 구현 계획 → SPEC-CLI-ANALYSIS-001
    """
    lines = content.split('\n')
    if lines:
        first_line = lines[0]
        match = re.search(r'(SPEC-[A-Z0-9-]+)', first_line)
        if match:
            return match.group(1)
    return None

def has_spec_tag(content: str, spec_id: str) -> bool:
    """
    파일에 @SPEC:{ID} 태그가 있는지 확인
    """
    pattern = rf'@SPEC:{spec_id}'
    return bool(re.search(pattern, content))

def add_spec_tag_to_file(content: str, spec_id: str) -> Tuple[str, bool]:
    """
    파일의 제목 줄에 @SPEC:{ID} 태그 추가

    반환: (수정된 컨텐트, 변경 여부)
    """
    if has_spec_tag(content, spec_id):
        return content, False  # 이미 태그가 있음

    lines = content.split('\n')

    # YAML front matter 건너뛰기
    header_end = 0
    if lines[0].strip() == '---':
        for i in range(1, len(lines)):
            if lines[i].strip() == '---':
                header_end = i + 1
                break

    # YAML 이후부터 첫 번째 제목 찾기
    title_line_idx = None
    for i in range(header_end, len(lines)):
        if lines[i].startswith('#'):
            title_line_idx = i
            break

    if title_line_idx is None:
        # 마크다운 제목이 없음
        return content, False

    title_line = lines[title_line_idx]

    # 이미 @SPEC:로 시작하면 그냥 반환
    if '@SPEC:' in title_line:
        return content, False

    # SPEC-를 포함하는 제목이면 @SPEC:로 변환
    if 'SPEC-' in title_line and '@SPEC:' not in title_line:
        # "# 구현 계획: SPEC-BAAS-ECOSYSTEM-001" → "# @SPEC:BAAS-ECOSYSTEM-001 구현 계획: SPEC-BAAS-ECOSYSTEM-001"
        # 또는 "# SPEC-CLI-ANALYSIS-001 구현 계획" → "# @SPEC:CLI-ANALYSIS-001 구현 계획"

        # spec_id에서 "SPEC-" 제거한 부분
        short_spec_id = spec_id.replace('SPEC-', '')

        # 방법 1: 제목 맨 앞에 @SPEC:ID 추가
        modified_line = title_line.replace('SPEC-', f'@SPEC:{short_spec_id} ', 1)

        # 만약 위의 처리가 이상하면
        if not modified_line.startswith('# @SPEC:'):
            # 방법 2: 제목 앞에 @SPEC: 태그 삽입
            if title_line.startswith('# '):
                modified_line = f'# @SPEC:{short_spec_id} ' + title_line[2:]
            else:
                modified_line = title_line  # 변경 안 함

        lines[title_line_idx] = modified_line
        modified_content = '\n'.join(lines)
        return modified_content, True

    return content, False

def process_spec_files(spec_id: str = None, dry_run: bool = False) -> Dict:
    """
    모든 SPEC 파일을 처리

    Args:
        spec_id: 특정 SPEC ID만 처리 (None이면 모두)
        dry_run: True이면 실제 변경하지 않음

    Returns:
        처리 결과 딕셔너리
    """
    specs_dir = Path('.moai/specs')
    results = {
        'total_files': 0,
        'files_without_tags': 0,
        'files_fixed': 0,
        'errors': [],
        'changes': []
    }

    if not specs_dir.exists():
        results['errors'].append(f"디렉토리 없음: {specs_dir}")
        return results

    # 모든 SPEC 디렉토리 찾기
    spec_dirs = sorted([d for d in specs_dir.iterdir() if d.is_dir() and d.name.startswith('SPEC-')])

    for spec_dir in spec_dirs:
        current_spec_id = spec_dir.name

        # 특정 SPEC ID만 처리하는 경우
        if spec_id and current_spec_id != spec_id:
            continue

        # 중요한 파일들만 처리
        target_files = ['plan.md', 'acceptance.md']

        for filename in target_files:
            filepath = spec_dir / filename

            if not filepath.exists():
                continue

            results['total_files'] += 1

            try:
                # 파일 읽기
                content = filepath.read_text(encoding='utf-8')

                # SPEC ID 추출 (여러 방법 시도)
                extracted_id = extract_spec_id_from_title(content)
                if not extracted_id:
                    extracted_id = current_spec_id

                # 태그 확인
                if has_spec_tag(content, extracted_id):
                    # 이미 태그가 있음
                    continue

                results['files_without_tags'] += 1

                # 태그 추가
                modified_content, changed = add_spec_tag_to_file(content, extracted_id)

                if changed:
                    if not dry_run:
                        filepath.write_text(modified_content, encoding='utf-8')

                    results['files_fixed'] += 1
                    results['changes'].append({
                        'file': str(filepath),
                        'spec_id': extracted_id,
                        'action': 'added_tag'
                    })
                    try:
                        rel_path = filepath.relative_to(Path.cwd())
                    except ValueError:
                        rel_path = filepath
                    print(f"✅ 수정: {rel_path} → @SPEC:{extracted_id} 태그 추가")

            except Exception as e:
                results['errors'].append(f"오류 처리 중 {filepath}: {str(e)}")
                print(f"❌ 오류: {filepath} → {str(e)}")

    return results

def main():
    """메인 함수"""
    import argparse

    parser = argparse.ArgumentParser(description='SPEC 파일의 누락된 @SPEC: 태그 자동 추가')
    parser.add_argument('--dry-run', action='store_true', help='실제 변경하지 않고 미리보기만')
    parser.add_argument('--spec-id', type=str, help='특정 SPEC ID만 처리')

    args = parser.parse_args()

    print(f"🔍 SPEC 파일 검사 시작{'(Dry-run 모드)' if args.dry_run else ''}...")
    print()

    results = process_spec_files(spec_id=args.spec_id, dry_run=args.dry_run)

    print()
    print("=" * 60)
    print("📊 처리 결과 요약")
    print("=" * 60)
    print(f"총 처리한 파일: {results['total_files']}")
    print(f"태그 누락 파일: {results['files_without_tags']}")
    print(f"수정된 파일: {results['files_fixed']}")

    if results['errors']:
        print()
        print("⚠️ 발생한 오류:")
        for error in results['errors']:
            print(f"  - {error}")

    if results['changes']:
        print()
        print("📝 변경 사항:")
        for change in results['changes']:
            print(f"  - {change['file']}")

    if args.dry_run and results['files_without_tags'] > 0:
        print()
        print("💡 실제 수정을 하려면 --dry-run 옵션 없이 다시 실행하세요:")
        print(f"   python3 {Path(__file__).relative_to(Path.cwd())} {f'--spec-id {args.spec_id}' if args.spec_id else ''}")

if __name__ == '__main__':
    main()
