#!/usr/bin/env python3
"""
MoAI-ADK Stage Validator
4단계 파이프라인의 각 Gate 검수 자동화
"""
import sys
import json
import argparse
from pathlib import Path
import re
import subprocess
from typing import Dict, List, Optional, Tuple


class StageValidator:
    def __init__(self, project_root: Path):
        self.project_root = project_root
        self.moai_path = project_root / '.moai'
        self.indexes_path = self.moai_path / 'indexes'
        
    def validate_specify_stage(self) -> Dict[str, any]:
        """SPECIFY Gate 검수"""
        results = {
            'stage': 'SPECIFY',
            'passed': True,
            'checks': [],
            'errors': [],
            'warnings': []
        }
        
        # EARS 형식 확인
        spec_files = list(self.moai_path.glob('specs/*/spec.md'))
        for spec_file in spec_files:
            if spec_file.exists():
                content = spec_file.read_text()
                
                # EARS 키워드 확인
                ears_keywords = re.findall(r'\b(WHEN|IF|WHILE|WHERE|UBIQUITOUS)\b', content)
                if not ears_keywords:
                    results['errors'].append(f"{spec_file}: EARS 형식 키워드 누락")
                    results['passed'] = False
                else:
                    results['checks'].append(f"{spec_file}: EARS 키워드 {len(ears_keywords)}개 확인")
                
                # [NEEDS CLARIFICATION] 확인
                clarifications = re.findall(r'\[NEEDS CLARIFICATION[^\]]*\]', content)
                if clarifications:
                    results['warnings'].append(f"{spec_file}: 미해결 명확화 항목 {len(clarifications)}개")
                
                # User Story 확인
                user_stories = re.findall(r'US-\d{3}', content)
                if not user_stories:
                    results['errors'].append(f"{spec_file}: User Story 번호 누락")
                    results['passed'] = False
                else:
                    results['checks'].append(f"{spec_file}: User Story {len(user_stories)}개 확인")
        
        return results
    
    def validate_plan_stage(self) -> Dict[str, any]:
        """PLAN Gate 검수"""
        results = {
            'stage': 'PLAN',
            'passed': True,
            'checks': [],
            'errors': [],
            'warnings': []
        }
        
        # Constitution Check 확인
        plan_files = list(self.moai_path.glob('specs/*/plan.md'))
        for plan_file in plan_files:
            if plan_file.exists():
                content = plan_file.read_text()
                
                # Constitution Check 항목
                if 'Constitution Check' not in content:
                    results['errors'].append(f"{plan_file}: Constitution Check 누락")
                    results['passed'] = False
                else:
                    results['checks'].append(f"{plan_file}: Constitution Check 확인")
                
                # research.md 존재 확인
                research_path = plan_file.parent / 'research.md'
                if not research_path.exists():
                    results['warnings'].append(f"{research_path}: research.md 파일 누락")
        
        # ADR 확인
        adr_files = list((self.moai_path / 'memory' / 'decisions').glob('ADR-*.md'))
        if not adr_files:
            results['warnings'].append("ADR 문서가 없습니다")
        else:
            results['checks'].append(f"ADR 문서 {len(adr_files)}개 확인")
        
        return results
    
    def validate_tasks_stage(self) -> Dict[str, any]:
        """TASKS Gate 검수"""  
        results = {
            'stage': 'TASKS',
            'passed': True,
            'checks': [],
            'errors': [],
            'warnings': []
        }
        
        # TDD 순서 확인
        tasks_files = list(self.moai_path.glob('specs/*/tasks.md'))
        for tasks_file in tasks_files:
            if tasks_file.exists():
                content = tasks_file.read_text()
                
                # TDD 마커 확인
                if 'Tests First (TDD)' not in content and 'RED-GREEN-REFACTOR' not in content:
                    results['errors'].append(f"{tasks_file}: TDD 순서 마커 누락")
                    results['passed'] = False
                else:
                    results['checks'].append(f"{tasks_file}: TDD 순서 확인")
                
                # [P] 병렬 마커 확인
                parallel_markers = re.findall(r'\[P\]', content)
                if not parallel_markers:
                    results['warnings'].append(f"{tasks_file}: 병렬 실행 마커 없음")
                else:
                    results['checks'].append(f"{tasks_file}: 병렬 작업 {len(parallel_markers)}개 확인")
                
                # 태스크 번호 확인
                task_numbers = re.findall(r'T\d{3}', content)
                if not task_numbers:
                    results['errors'].append(f"{tasks_file}: 태스크 번호 형식 오류")
                    results['passed'] = False
                else:
                    results['checks'].append(f"{tasks_file}: 태스크 {len(task_numbers)}개 확인")
        
        return results
    
    def validate_implement_stage(self) -> Dict[str, any]:
        """IMPLEMENT Gate 검수"""
        results = {
            'stage': 'IMPLEMENT', 
            'passed': True,
            'checks': [],
            'errors': [],
            'warnings': []
        }
        
        # 테스트 실행
        try:
            result = subprocess.run(['pytest', '--tb=short'], 
                                 capture_output=True, text=True, cwd=self.project_root)
            if result.returncode == 0:
                results['checks'].append("모든 테스트 통과")
            else:
                results['errors'].append(f"테스트 실패: {result.stderr}")
                results['passed'] = False
        except FileNotFoundError:
            results['warnings'].append("pytest를 찾을 수 없습니다")
        
        # 커버리지 확인
        config_path = self.moai_path / 'config.json'
        if config_path.exists():
            config = json.loads(config_path.read_text())
            target = config.get('quality_gates', {}).get('coverageTarget', 0.8)
            
            try:
                result = subprocess.run(['pytest', '--cov=.', '--cov-report=json'], 
                                     capture_output=True, text=True, cwd=self.project_root)
                
                coverage_file = self.project_root / 'coverage.json'
                if coverage_file.exists():
                    coverage_data = json.loads(coverage_file.read_text())
                    actual = coverage_data['totals']['percent_covered'] / 100
                    
                    if actual >= target:
                        results['checks'].append(f"커버리지 목표 달성: {actual:.1%}")
                    else:
                        results['errors'].append(f"커버리지 부족: {actual:.1%} < {target:.0%}")
                        results['passed'] = False
            except (FileNotFoundError, json.JSONDecodeError):
                results['warnings'].append("커버리지 측정 실패")
        
        return results
    
    def validate_all_stages(self) -> Dict[str, any]:
        """모든 단계 검수"""
        all_results = {
            'overall_passed': True,
            'total_checks': 0,
            'total_errors': 0,
            'total_warnings': 0,
            'stages': {}
        }
        
        validators = {
            'SPECIFY': self.validate_specify_stage,
            'PLAN': self.validate_plan_stage,
            'TASKS': self.validate_tasks_stage, 
            'IMPLEMENT': self.validate_implement_stage
        }
        
        for stage_name, validator in validators.items():
            result = validator()
            all_results['stages'][stage_name] = result
            all_results['total_checks'] += len(result['checks'])
            all_results['total_errors'] += len(result['errors'])
            all_results['total_warnings'] += len(result['warnings'])
            
            if not result['passed']:
                all_results['overall_passed'] = False
        
        return all_results


def print_validation_results(results: Dict[str, any], verbose: bool = False):
    """검증 결과 출력"""
    if 'overall_passed' in results:
        # 전체 결과
        print(f"\n🔍 MoAI-ADK 전체 검증 결과")
        print(f"{'='*50}")
        print(f"전체 결과: {'✅ PASS' if results['overall_passed'] else '❌ FAIL'}")
        print(f"검사 항목: {results['total_checks']}개")
        print(f"오류: {results['total_errors']}개") 
        print(f"경고: {results['total_warnings']}개")
        
        for stage_name, stage_result in results['stages'].items():
            status = '✅ PASS' if stage_result['passed'] else '❌ FAIL'
            print(f"\n{stage_name} Gate: {status}")
            
            if verbose or not stage_result['passed']:
                for check in stage_result['checks']:
                    print(f"  ✓ {check}")
                for error in stage_result['errors']:
                    print(f"  ❌ {error}")
                for warning in stage_result['warnings']:
                    print(f"  ⚠️ {warning}")
    else:
        # 단일 단계 결과
        stage_name = results['stage']
        status = '✅ PASS' if results['passed'] else '❌ FAIL'
        print(f"\n🔍 {stage_name} Gate 검증 결과: {status}")
        print(f"{'='*50}")
        
        for check in results['checks']:
            print(f"✓ {check}")
        for error in results['errors']:
            print(f"❌ {error}")
        for warning in results['warnings']:
            print(f"⚠️ {warning}")


def main():
    parser = argparse.ArgumentParser(description='MoAI-ADK Stage Validator')
    parser.add_argument('--stage', choices=['SPECIFY', 'PLAN', 'TASKS', 'IMPLEMENT', 'ALL'],
                       default='ALL', help='검증할 단계')
    parser.add_argument('--project-root', type=Path, default=Path.cwd(),
                       help='프로젝트 루트 디렉토리')
    parser.add_argument('--verbose', '-v', action='store_true',
                       help='상세한 출력')
    
    args = parser.parse_args()
    
    validator = StageValidator(args.project_root)
    
    if args.stage == 'ALL':
        results = validator.validate_all_stages()
    else:
        stage_methods = {
            'SPECIFY': validator.validate_specify_stage,
            'PLAN': validator.validate_plan_stage,
            'TASKS': validator.validate_tasks_stage,
            'IMPLEMENT': validator.validate_implement_stage
        }
        results = stage_methods[args.stage]()
    
    print_validation_results(results, args.verbose)
    
    # Exit code 설정
    if 'overall_passed' in results:
        sys.exit(0 if results['overall_passed'] else 1)
    else:
        sys.exit(0 if results['passed'] else 1)


if __name__ == '__main__':
    main()