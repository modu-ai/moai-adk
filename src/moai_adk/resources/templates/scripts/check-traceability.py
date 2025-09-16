#!/usr/bin/env python3
"""
MoAI-ADK Traceability Checker
14-Core TAG 시스템의 추적성 검증
"""
import sys
import json
import argparse
from pathlib import Path
import re
from typing import Dict, List, Set, Tuple
from collections import defaultdict


class TraceabilityChecker:
    def __init__(self, project_root: Path):
        self.project_root = project_root
        self.moai_path = project_root / '.moai'
        self.indexes_path = self.moai_path / 'indexes'
        
        # 14-Core TAG 패턴
        self.tag_patterns = {
            'SPEC': ['REQ', 'DESIGN', 'TASK'],
            'STEERING': ['VISION', 'STRUCT', 'TECH', 'STACK'],
            'IMPLEMENTATION': ['FEATURE', 'API', 'TEST', 'DATA'],
            'QUALITY': ['PERF', 'SEC', 'DEBT', 'TODO']
        }
        
        # 추적성 체인
        self.traceability_chains = {
            'primary': ['REQ', 'DESIGN', 'TASK', 'TEST'],
            'steering': ['VISION', 'STRUCT', 'TECH', 'STACK'],
            'implementation': ['FEATURE', 'API', 'DATA'],
            'quality': ['PERF', 'SEC', 'DEBT', 'TODO']
        }
    
    def scan_all_tags(self) -> Dict[str, List[str]]:
        """프로젝트 전체에서 모든 @TAG 수집"""
        all_tags = defaultdict(list)
        
        # .moai 디렉토리 스캔
        for md_file in self.moai_path.rglob('*.md'):
            if md_file.is_file():
                content = md_file.read_text(encoding='utf-8', errors='ignore')
                tags = self.extract_tags_from_content(content)
                for tag in tags:
                    all_tags[tag].append(str(md_file.relative_to(self.project_root)))
        
        # 소스 코드에서도 태그 수집 (src, tests 디렉토리)
        for src_dir in ['src', 'tests']:
            src_path = self.project_root / src_dir
            if src_path.exists():
                for code_file in src_path.rglob('*'):
                    if code_file.suffix in ['.py', '.js', '.ts', '.tsx', '.jsx', '.md']:
                        try:
                            content = code_file.read_text(encoding='utf-8', errors='ignore')
                            tags = self.extract_tags_from_content(content)
                            for tag in tags:
                                all_tags[tag].append(str(code_file.relative_to(self.project_root)))
                        except (UnicodeDecodeError, PermissionError):
                            continue
        
        return dict(all_tags)
    
    def extract_tags_from_content(self, content: str) -> List[str]:
        """텍스트에서 @TAG 추출"""
        tag_pattern = r'@([A-Z]+)[-:]([A-Z0-9-]+)'
        matches = re.findall(tag_pattern, content, re.MULTILINE)
        return [f"@{tag_type}:{tag_id}" for tag_type, tag_id in matches]
    
    def validate_tag_naming(self, tags: Dict[str, List[str]]) -> Dict[str, any]:
        """태그 네이밍 규칙 검증"""
        results = {
            'valid_tags': 0,
            'invalid_tags': 0,
            'naming_errors': [],
            'category_distribution': defaultdict(int)
        }
        
        # 유효한 태그 타입 수집
        valid_types = set()
        for category_tags in self.tag_patterns.values():
            valid_types.update(category_tags)
        
        for tag, locations in tags.items():
            try:
                tag_type, tag_id = tag[1:].split(':', 1)
                
                # 유효한 태그 타입인지 확인
                if tag_type not in valid_types:
                    results['invalid_tags'] += 1
                    results['naming_errors'].append({
                        'tag': tag,
                        'error': f"'{tag_type}'는 14-Core 체계에 없는 태그 타입",
                        'locations': locations[:3]  # 최대 3개 위치만 표시
                    })
                    continue
                
                # ID 형식 검증
                if not re.match(r'^[A-Z0-9-]+$', tag_id):
                    results['invalid_tags'] += 1
                    results['naming_errors'].append({
                        'tag': tag,
                        'error': f"'{tag_id}'는 올바른 태그 ID 형식이 아님 (대문자, 숫자, 하이픈만 허용)",
                        'locations': locations[:3]
                    })
                    continue
                
                # 카테고리별 분포 계산
                for category, category_tags in self.tag_patterns.items():
                    if tag_type in category_tags:
                        results['category_distribution'][category] += 1
                        break
                
                results['valid_tags'] += 1
                
            except ValueError:
                results['invalid_tags'] += 1
                results['naming_errors'].append({
                    'tag': tag,
                    'error': "태그 형식이 @TYPE:ID 패턴과 맞지 않음",
                    'locations': locations[:3]
                })
        
        return results
    
    def check_traceability_chains(self, tags: Dict[str, List[str]]) -> Dict[str, any]:
        """추적성 체인 검증"""
        results = {
            'chain_coverage': {},
            'missing_links': [],
            'orphaned_tags': [],
            'complete_chains': 0,
            'incomplete_chains': 0
        }
        
        # 태그를 타입별로 그룹화
        tags_by_type = defaultdict(list)
        for tag in tags.keys():
            try:
                tag_type = tag.split(':', 1)[0][1:]  # @제거하고 타입 추출
                tags_by_type[tag_type].append(tag)
            except IndexError:
                continue
        
        # 각 체인별 커버리지 확인
        for chain_name, chain_types in self.traceability_chains.items():
            chain_tags = {}
            missing_types = []
            
            for tag_type in chain_types:
                if tag_type in tags_by_type:
                    chain_tags[tag_type] = tags_by_type[tag_type]
                else:
                    missing_types.append(tag_type)
            
            coverage = len(chain_tags) / len(chain_types)
            results['chain_coverage'][chain_name] = {
                'coverage': coverage,
                'present_types': list(chain_tags.keys()),
                'missing_types': missing_types,
                'tag_counts': {t: len(chain_tags.get(t, [])) for t in chain_types}
            }
            
            if coverage == 1.0:
                results['complete_chains'] += 1
            else:
                results['incomplete_chains'] += 1
        
        # 링크 검증 (REQ -> DESIGN -> TASK -> TEST)
        req_tags = tags_by_type.get('REQ', [])
        design_tags = tags_by_type.get('DESIGN', [])
        task_tags = tags_by_type.get('TASK', [])
        test_tags = tags_by_type.get('TEST', [])
        
        for req_tag in req_tags:
            req_id = req_tag.split(':', 1)[1]
            
            # 해당하는 DESIGN 태그 찾기
            related_design = [d for d in design_tags if req_id in d]
            if not related_design:
                results['missing_links'].append({
                    'from': req_tag,
                    'to': f"@DESIGN:*{req_id}*",
                    'type': 'REQ->DESIGN'
                })
        
        # 고아 태그 찾기 (다른 태그와 연결되지 않은 태그)
        for tag in tags.keys():
            tag_id = tag.split(':', 1)[1]
            related_count = sum(1 for other_tag in tags.keys() 
                              if other_tag != tag and tag_id in other_tag)
            if related_count == 0 and len(tags[tag]) == 1:
                results['orphaned_tags'].append(tag)
        
        return results
    
    def check_consistency(self, tags: Dict[str, List[str]]) -> Dict[str, any]:
        """일관성 검증"""
        results = {
            'file_consistency': [],
            'duplicate_definitions': [],
            'naming_inconsistencies': []
        }
        
        # 파일별 태그 일관성 확인
        file_tags = defaultdict(list)
        for tag, locations in tags.items():
            for location in locations:
                file_tags[location].append(tag)
        
        for file_path, file_tag_list in file_tags.items():
            # 같은 파일에서 일관되지 않은 네이밍 패턴 찾기
            tag_patterns_in_file = set()
            for tag in file_tag_list:
                try:
                    tag_type = tag.split(':', 1)[0][1:]
                    tag_patterns_in_file.add(tag_type)
                except IndexError:
                    continue
            
            if len(tag_patterns_in_file) > 5:  # 한 파일에 너무 많은 다른 태그 타입
                results['file_consistency'].append({
                    'file': file_path,
                    'tag_types': list(tag_patterns_in_file),
                    'count': len(tag_patterns_in_file),
                    'warning': '한 파일에 너무 많은 태그 타입'
                })
        
        # 중복 정의 찾기
        tag_ids = defaultdict(list)
        for tag in tags.keys():
            try:
                tag_id = tag.split(':', 1)[1]
                tag_ids[tag_id].append(tag)
            except IndexError:
                continue
        
        for tag_id, tag_list in tag_ids.items():
            if len(tag_list) > 1:
                # 같은 ID를 가진 다른 타입의 태그들
                different_types = set(t.split(':', 1)[0] for t in tag_list)
                if len(different_types) > 1:
                    results['duplicate_definitions'].append({
                        'tag_id': tag_id,
                        'conflicting_tags': tag_list,
                        'types': list(different_types)
                    })
        
        return results
    
    def generate_report(self) -> Dict[str, any]:
        """전체 추적성 보고서 생성"""
        print("🔍 14-Core TAG 시스템 스캔 중...")
        tags = self.scan_all_tags()
        
        print(f"📊 총 {len(tags)}개 태그 발견")
        
        print("🏷️ 태그 네이밍 검증 중...")
        naming_results = self.validate_tag_naming(tags)
        
        print("🔗 추적성 체인 검증 중...")
        chain_results = self.check_traceability_chains(tags)
        
        print("⚖️ 일관성 검증 중...")
        consistency_results = self.check_consistency(tags)
        
        report = {
            'summary': {
                'total_tags': len(tags),
                'valid_tags': naming_results['valid_tags'],
                'invalid_tags': naming_results['invalid_tags'],
                'complete_chains': chain_results['complete_chains'],
                'incomplete_chains': chain_results['incomplete_chains'],
                'orphaned_tags': len(chain_results['orphaned_tags'])
            },
            'tags': tags,
            'naming': naming_results,
            'traceability': chain_results,
            'consistency': consistency_results
        }
        
        return report


def print_report(report: Dict[str, any], verbose: bool = False):
    """보고서 출력"""
    summary = report['summary']
    
    print(f"\n📈 MoAI-ADK 추적성 검증 보고서")
    print(f"{'='*50}")
    print(f"총 태그 수: {summary['total_tags']}")
    print(f"유효한 태그: {summary['valid_tags']}")
    print(f"잘못된 태그: {summary['invalid_tags']}")
    print(f"완성된 체인: {summary['complete_chains']}")
    print(f"불완전한 체인: {summary['incomplete_chains']}")
    print(f"고아 태그: {summary['orphaned_tags']}")
    
    # 전체 결과 판정
    overall_score = (summary['valid_tags'] / max(summary['total_tags'], 1)) * 0.4 + \
                   (summary['complete_chains'] / 4) * 0.4 + \
                   (1 - summary['orphaned_tags'] / max(summary['total_tags'], 1)) * 0.2
    
    print(f"\n📊 추적성 품질 점수: {overall_score:.2%}")
    
    if overall_score >= 0.8:
        print("✅ 우수한 추적성 품질")
    elif overall_score >= 0.6:
        print("⚠️ 보통 추적성 품질 - 개선 필요")
    else:
        print("❌ 낮은 추적성 품질 - 즉시 개선 필요")
    
    # 상세 정보
    if verbose or overall_score < 0.8:
        print(f"\n🏷️ 카테고리별 태그 분포:")
        for category, count in report['naming']['category_distribution'].items():
            print(f"  {category}: {count}개")
        
        print(f"\n🔗 체인별 커버리지:")
        for chain_name, chain_info in report['traceability']['chain_coverage'].items():
            coverage = chain_info['coverage']
            status = '✅' if coverage == 1.0 else '⚠️' if coverage >= 0.5 else '❌'
            print(f"  {status} {chain_name}: {coverage:.1%}")
            if verbose and coverage < 1.0:
                print(f"    누락된 타입: {', '.join(chain_info['missing_types'])}")
        
        if report['naming']['naming_errors']:
            print(f"\n❌ 네이밍 오류 ({len(report['naming']['naming_errors'])}개):")
            for error in report['naming']['naming_errors'][:5]:  # 최대 5개만 표시
                print(f"  {error['tag']}: {error['error']}")
        
        if report['traceability']['orphaned_tags']:
            print(f"\n🏝️ 고아 태그 ({len(report['traceability']['orphaned_tags'])}개):")
            for tag in report['traceability']['orphaned_tags'][:5]:  # 최대 5개만 표시
                print(f"  {tag}")


def main():
    parser = argparse.ArgumentParser(description='MoAI-ADK Traceability Checker')
    parser.add_argument('--project-root', type=Path, default=Path.cwd(),
                       help='프로젝트 루트 디렉토리')
    parser.add_argument('--verbose', '-v', action='store_true',
                       help='상세한 출력')
    parser.add_argument('--output', '-o', type=Path,
                       help='JSON 보고서 파일 출력')
    parser.add_argument('--threshold', type=float, default=0.8,
                       help='통과 기준 점수 (0.0-1.0)')
    
    args = parser.parse_args()
    
    checker = TraceabilityChecker(args.project_root)
    report = checker.generate_report()
    
    print_report(report, args.verbose)
    
    if args.output:
        args.output.write_text(json.dumps(report, indent=2, ensure_ascii=False))
        print(f"\n📄 보고서가 {args.output}에 저장되었습니다")
    
    # 점수 기반 exit code
    summary = report['summary']
    overall_score = (summary['valid_tags'] / max(summary['total_tags'], 1)) * 0.4 + \
                   (summary['complete_chains'] / 4) * 0.4 + \
                   (1 - summary['orphaned_tags'] / max(summary['total_tags'], 1)) * 0.2
    
    sys.exit(0 if overall_score >= args.threshold else 1)


if __name__ == '__main__':
    main()