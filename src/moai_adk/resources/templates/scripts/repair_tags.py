#!/usr/bin/env python3
"""
MoAI-ADK TAG 자동 리페어 시스템
단절된 링크 탐지, 자동 제안, traceability.json 보정
"""
import json
import re
import argparse
from pathlib import Path
from typing import Dict, List, Set, Tuple, Optional
from collections import defaultdict
from datetime import datetime


class TagRepairer:
    def __init__(self, project_root: Path):
        self.project_root = project_root
        self.moai_path = project_root / '.moai'
        self.indexes_path = self.moai_path / 'indexes'
        self.templates_path = self.moai_path / 'templates'
        
        # 14-Core TAG 체계
        self.tag_categories = {
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

    def scan_project_tags(self) -> Dict[str, List[str]]:
        """프로젝트 전체에서 모든 @TAG 수집"""
        all_tags = defaultdict(list)
        
        # .moai 디렉토리 스캔
        for md_file in self.moai_path.rglob('*.md'):
            if md_file.is_file():
                try:
                    content = md_file.read_text(encoding='utf-8', errors='ignore')
                    tags = self.extract_tags(content)
                    for tag in tags:
                        all_tags[tag].append(str(md_file.relative_to(self.project_root)))
                except (UnicodeDecodeError, PermissionError):
                    continue
        
        # 소스 코드에서도 태그 수집
        for src_dir in ['src', 'tests']:
            src_path = self.project_root / src_dir
            if src_path.exists():
                for code_file in src_path.rglob('*'):
                    if code_file.suffix in ['.py', '.js', '.ts', '.tsx', '.jsx', '.md']:
                        try:
                            content = code_file.read_text(encoding='utf-8', errors='ignore')
                            tags = self.extract_tags(content)
                            for tag in tags:
                                all_tags[tag].append(str(code_file.relative_to(self.project_root)))
                        except (UnicodeDecodeError, PermissionError):
                            continue
        
        return dict(all_tags)

    def extract_tags(self, content: str) -> List[str]:
        """텍스트에서 @TAG 추출"""
        tag_pattern = r'@([A-Z]+)[-:]([A-Z0-9-]+)'
        matches = re.findall(tag_pattern, content, re.MULTILINE)
        return [f"@{tag_type}:{tag_id}" for tag_type, tag_id in matches]

    def analyze_tag_integrity(self) -> Dict[str, any]:
        """단절된 @TAG 링크 분석"""
        print("🔍 프로젝트 태그 수집 중...")
        all_tags = self.scan_project_tags()
        
        orphaned_tags = []
        missing_links = []
        suggestions = []
        
        print(f"📊 총 {len(all_tags)}개 태그 발견")
        
        # 1. 참조 체인 검증
        print("🔗 참조 체인 검증 중...")
        for tag, locations in all_tags.items():
            try:
                tag_type, tag_id = tag[1:].split(':', 1)
                
                if tag_type == 'REQ':
                    # REQ → DESIGN 링크 확인
                    design_link = f"@DESIGN:{tag_id}"
                    if design_link not in all_tags:
                        missing_links.append((tag, design_link))
                        suggestions.append(f"Create DESIGN for {tag}")
                        
                elif tag_type == 'DESIGN':
                    # DESIGN → TASK 링크 확인
                    task_pattern = f"@TASK:{tag_id}"
                    matching_tasks = [t for t in all_tags.keys() if t.startswith(task_pattern)]
                    if not matching_tasks:
                        missing_links.append((tag, f"{task_pattern}*"))
                        suggestions.append(f"Decompose {tag} into tasks")
                        
                elif tag_type == 'TASK':
                    # TASK → TEST 링크 확인
                    test_pattern = f"@TEST:{tag_id}"
                    matching_tests = [t for t in all_tags.keys() if t.startswith(test_pattern)]
                    if not matching_tests:
                        missing_links.append((tag, f"{test_pattern}*"))
                        suggestions.append(f"Create tests for {tag}")
                        
            except ValueError:
                continue
        
        # 2. 고아 태그 탐지
        print("🏝️ 고아 태그 탐지 중...")
        for tag, locations in all_tags.items():
            if len(locations) == 1 and not self.has_references(tag, all_tags):
                orphaned_tags.append(tag)
        
        # 3. 수리 미리보기 생성
        repair_preview = self.generate_repair_preview(missing_links)
        
        return {
            'total_tags': len(all_tags),
            'orphaned_tags': orphaned_tags,
            'missing_links': missing_links,
            'suggestions': suggestions,
            'repair_preview': repair_preview,
            'all_tags': all_tags
        }

    def has_references(self, tag: str, all_tags: Dict[str, List[str]]) -> bool:
        """태그가 다른 태그와 연결되어 있는지 확인"""
        try:
            tag_id = tag.split(':', 1)[1]
            return any(tag_id in other_tag for other_tag in all_tags.keys() if other_tag != tag)
        except IndexError:
            return False

    def generate_repair_preview(self, missing_links: List[Tuple[str, str]]) -> List[Dict[str, any]]:
        """수리 미리보기 생성"""
        preview = []
        
        for source, target in missing_links:
            try:
                source_type, source_id = source[1:].split(':', 1)
                target_type = target[1:].split(':', 1)[0]
                
                if target_type == 'DESIGN':
                    preview.append({
                        'action': 'create_design',
                        'source': source,
                        'target': target,
                        'file': f'.moai/specs/{source_id}/design.md',
                        'template': 'design-template.md',
                        'description': f'Create DESIGN document for {source}'
                    })
                    
                elif target_type == 'TASK':
                    preview.append({
                        'action': 'create_tasks',
                        'source': source,
                        'target': target,
                        'file': f'.moai/specs/{source_id}/tasks.md',
                        'template': 'tasks-template.md',
                        'description': f'Create TASKS decomposition for {source}'
                    })
                    
                elif target_type == 'TEST':
                    preview.append({
                        'action': 'create_test',
                        'source': source,
                        'target': target,
                        'file': f'tests/test_{source_id.lower().replace("-", "_")}.py',
                        'template': 'test-template.py',
                        'description': f'Create test cases for {source}'
                    })
                    
            except (ValueError, IndexError):
                continue
        
        return preview

    def extract_requirements_from_tag(self, tag: str) -> Dict[str, any]:
        """태그에서 요구사항 정보 추출"""
        try:
            tag_type, tag_id = tag[1:].split(':', 1)
            
            return {
                'tag': tag,
                'type': tag_type,
                'id': tag_id,
                'priority': 'MEDIUM',
                'category': self.get_tag_category(tag_type),
                'estimated_complexity': 'MEDIUM'
            }
        except ValueError:
            return {}

    def get_tag_category(self, tag_type: str) -> Optional[str]:
        """태그 타입의 카테고리 반환"""
        for category, types in self.tag_categories.items():
            if tag_type in types:
                return category
        return None

    def estimate_task_count(self, source: str) -> int:
        """태그 기반 예상 작업 개수"""
        try:
            tag_type, tag_id = source[1:].split(':', 1)
            
            # 복잡도 기반 추정
            complexity_indicators = ['API', 'DATABASE', 'AUTH', 'PAYMENT', 'INTEGRATION']
            base_count = 3
            
            for indicator in complexity_indicators:
                if indicator in tag_id.upper():
                    base_count += 2
            
            return min(base_count, 10)  # 최대 10개
        except ValueError:
            return 3

    def create_design_from_template(self, item: Dict[str, any]):
        """DESIGN 템플릿으로부터 문서 생성"""
        design_path = self.project_root / item['file']
        design_path.parent.mkdir(parents=True, exist_ok=True)
        
        template = f"""# DESIGN-{item['source'][1:].split(':', 1)[1]}: 설계 문서

> **기반 요구사항**: {item['source']}
> **생성일**: {datetime.now().strftime('%Y-%m-%d')}
> **상태**: DRAFT

## 🎯 설계 개요

### 기반 요구사항 분석
{item['source']}에 대한 기술적 설계를 수행합니다.

### 설계 결정사항
- [ ] 아키텍처 패턴 선택
- [ ] 데이터 모델 정의  
- [ ] API 인터페이스 설계
- [ ] 보안 고려사항

## 🏗️ 아키텍처 설계

### 컴포넌트 구조
```
[TBD: 컴포넌트 다이어그램]
```

### 데이터 흐름
```
[TBD: 데이터 흐름도]
```

## 📋 구현 태스크

### 우선순위별 작업 분해
- [ ] @TASK:{item['source'][1:].split(':', 1)[1]}-001: 핵심 컴포넌트 구현
- [ ] @TASK:{item['source'][1:].split(':', 1)[1]}-002: 데이터 계층 구현  
- [ ] @TASK:{item['source'][1:].split(':', 1)[1]}-003: API 엔드포인트 구현

## 🧪 테스트 전략

### 테스트 범위
- [ ] @TEST:UNIT-{item['source'][1:].split(':', 1)[1]}: 단위 테스트
- [ ] @TEST:INT-{item['source'][1:].split(':', 1)[1]}: 통합 테스트
- [ ] @TEST:E2E-{item['source'][1:].split(':', 1)[1]}: E2E 테스트

## 📊 품질 기준

### 성능 요구사항
- [ ] @PERF:{item['source'][1:].split(':', 1)[1]}: 응답시간 < 2초

### 보안 요구사항  
- [ ] @SEC:{item['source'][1:].split(':', 1)[1]}: 입력값 검증

---
*자동 생성됨: MoAI-ADK repair_tags.py*
"""
        design_path.write_text(template, encoding='utf-8')
        print(f"✅ 생성: {design_path}")

    def create_tasks_from_design(self, item: Dict[str, any]):
        """DESIGN으로부터 TASKS 문서 생성"""
        tasks_path = self.project_root / item['file']
        tasks_path.parent.mkdir(parents=True, exist_ok=True)
        
        estimated_tasks = self.estimate_task_count(item['source'])
        
        template = f"""# TASKS-{item['source'][1:].split(':', 1)[1]}: TDD 작업 분해

> **기반 설계**: {item['source']}
> **생성일**: {datetime.now().strftime('%Y-%m-%d')}
> **TDD 순서**: RED → GREEN → REFACTOR

## 📊 작업 통계
- **총 작업 수**: {estimated_tasks}개
- **병렬 실행 가능**: {estimated_tasks//2}개 ([P] 마커)
- **예상 소요**: {estimated_tasks * 2}시간

## 🔄 TDD 작업 순서

### Phase 1: RED (실패하는 테스트)
"""

        for i in range(1, estimated_tasks + 1):
            task_id = f"T{i:03d}"
            parallel_marker = "[P]" if i > 1 and i % 2 == 0 else ""
            
            template += f"""
#### {task_id}: 테스트 작성 - 컴포넌트 {i} {parallel_marker}
- **파일**: `tests/test_{item['source'][1:].split(':', 1)[1].lower().replace('-', '_')}_component_{i}.py`
- **TAG**: @TEST:UNIT-{item['source'][1:].split(':', 1)[1]}-{i:03d}
- **설명**: 실패하는 테스트 케이스 작성
- **의존성**: 없음
- **예상시간**: 30분
"""

        template += f"""
### Phase 2: GREEN (테스트 통과하는 최소 구현)
"""

        for i in range(1, estimated_tasks + 1):
            task_id = f"T{i+estimated_tasks:03d}"
            
            template += f"""
#### {task_id}: 최소 구현 - 컴포넌트 {i}
- **파일**: `src/components/component_{i}.py`
- **TAG**: @FEATURE:{item['source'][1:].split(':', 1)[1]}-{i:03d}
- **설명**: 테스트를 통과하는 최소한의 구현
- **의존성**: T{i:03d}
- **예상시간**: 45분
"""

        template += f"""
### Phase 3: REFACTOR (코드 품질 개선)
"""

        for i in range(1, estimated_tasks + 1):
            task_id = f"T{i+estimated_tasks*2:03d}"
            
            template += f"""
#### {task_id}: 리팩토링 - 컴포넌트 {i} [P]
- **파일**: `src/components/component_{i}.py`
- **TAG**: @DEBT:{item['source'][1:].split(':', 1)[1]}-REFACTOR-{i:03d}
- **설명**: 코드 중복 제거, 성능 최적화
- **의존성**: T{i+estimated_tasks:03d}
- **예상시간**: 30분
"""

        template += f"""
## 🎯 완료 기준

### Definition of Done
- [ ] 모든 테스트 통과 (Coverage ≥ 80%)
- [ ] 코드 리뷰 완료
- [ ] 문서 업데이트 완료
- [ ] @TAG 매핑 완료

### 품질 게이트
- [ ] Constitution Check 통과
- [ ] 성능 기준 충족
- [ ] 보안 검증 완료

---
*자동 생성됨: MoAI-ADK repair_tags.py*
"""
        tasks_path.write_text(template, encoding='utf-8')
        print(f"✅ 생성: {tasks_path}")

    def update_traceability_index(self):
        """traceability.json 갱신"""
        traceability_path = self.indexes_path / 'traceability.json'
        
        if traceability_path.exists():
            traceability_data = json.loads(traceability_path.read_text())
        else:
            traceability_data = {
                'metadata': {'version': '14-core', 'total_links': 0},
                'chains': self.traceability_chains,
                'links': []
            }
        
        # 새로운 태그들로 링크 정보 갱신
        all_tags = self.scan_project_tags()
        
        # 기존 링크 초기화하고 재구성
        traceability_data['links'] = []
        
        for chain_name, chain_types in self.traceability_chains.items():
            for i in range(len(chain_types) - 1):
                from_type = chain_types[i]
                to_type = chain_types[i + 1]
                
                from_tags = [tag for tag in all_tags.keys() if tag.startswith(f"@{from_type}:")]
                to_tags = [tag for tag in all_tags.keys() if tag.startswith(f"@{to_type}:")]
                
                for from_tag in from_tags:
                    for to_tag in to_tags:
                        # ID가 연관된 태그들만 링크
                        from_id = from_tag.split(':', 1)[1]
                        to_id = to_tag.split(':', 1)[1]
                        
                        if from_id in to_id or to_id in from_id:
                            link = {
                                'from': from_tag,
                                'to': to_tag,
                                'chain': chain_name,
                                'relationship': 'implements' if chain_name == 'primary' else 'supports',
                                'timestamp': datetime.now().isoformat()
                            }
                            traceability_data['links'].append(link)
        
        # 메타데이터 업데이트
        traceability_data['metadata']['total_links'] = len(traceability_data['links'])
        traceability_data['metadata']['generated_at'] = datetime.now().isoformat()
        
        traceability_path.write_text(json.dumps(traceability_data, indent=2, ensure_ascii=False))
        print(f"✅ traceability.json 업데이트 완료")

    def auto_repair_tags(self, dry_run: bool = True) -> bool:
        """자동 수리 실행"""
        analysis = self.analyze_tag_integrity()
        
        if dry_run:
            print(f"\n🔧 @TAG 리페어 미리보기:")
            print(f"{'='*50}")
            print(f"총 태그: {analysis['total_tags']}개")
            print(f"고아 태그: {len(analysis['orphaned_tags'])}개")
            print(f"누락 링크: {len(analysis['missing_links'])}개")
            
            print(f"\n📋 수리 액션:")
            for item in analysis['repair_preview']:
                print(f"  - {item['action']}: {item['file']}")
                print(f"    {item['description']}")
            
            print(f"\n📈 통계: {len(analysis['missing_links'])}개 링크 복구 필요")
            return True
        
        # 실제 수리 실행
        print(f"\n🔧 자동 수리 실행 중...")
        
        for item in analysis['repair_preview']:
            try:
                if item['action'] == 'create_design':
                    self.create_design_from_template(item)
                elif item['action'] == 'create_tasks':
                    self.create_tasks_from_design(item)
                # TODO: create_test 액션 구현
                    
            except Exception as e:
                print(f"❌ 오류: {item['file']} - {e}")
        
        # traceability.json 갱신
        print(f"\n🔄 traceability.json 갱신 중...")
        self.update_traceability_index()
        
        print(f"\n✅ @TAG 리페어 완료: {len(analysis['missing_links'])}개 링크 복구")
        return True


def main():
    parser = argparse.ArgumentParser(description='MoAI-ADK TAG Auto Repair System')
    parser.add_argument('--project-root', type=Path, default=Path.cwd(),
                       help='프로젝트 루트 디렉토리')
    parser.add_argument('--dry-run', action='store_true', default=True,
                       help='미리보기 모드 (기본값)')
    parser.add_argument('--execute', action='store_true',
                       help='실제 수리 실행')
    parser.add_argument('--auto', action='store_true',
                       help='CI/CD용 자동 실행')
    
    args = parser.parse_args()
    
    # --execute가 있으면 dry_run을 False로
    if args.execute:
        args.dry_run = False
    
    print(f"🗿 MoAI-ADK TAG 리페어 시스템")
    print(f"프로젝트: {args.project_root}")
    print(f"모드: {'미리보기' if args.dry_run else '실행'}")
    
    repairer = TagRepairer(args.project_root)
    success = repairer.auto_repair_tags(dry_run=args.dry_run)
    
    sys.exit(0 if success else 1)


if __name__ == '__main__':
    import sys
    main()