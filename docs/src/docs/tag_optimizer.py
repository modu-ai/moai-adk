"""
TAG 시스템 최적화 유틸리티
이 모듈은 문서 시스템의 TAG를 관리하고 최적화합니다.

:TAG-OPTIMIZER
"""

import re
import json
from pathlib import Path
from typing import Dict, List, Optional, Set, Tuple
from dataclasses import dataclass


@dataclass
class TagInfo:
    """TAG 정보 데이터 클래스"""
    full_tag: str
    code: str
    language: Optional[str]
    category: str
    line_number: int
    file_path: str


class TagOptimizer:
    """TAG 시스템 최적화 클래스"""

    def __init__(self, base_dir: str = "."):
        self.base_dir = Path(base_dir)
        self.document_manager = self._load_document_manager()

    def _load_document_manager(self):
        """문서 관리자 로드 (간단한 버전)"""
        config_path = self.base_dir / "multilingual_config.json"
        if config_path.exists():
            with open(config_path, 'r', encoding='utf-8') as f:
                config = json.load(f)

            # 간단한 문서 관리자 생성
            class SimpleDocManager:
                def __init__(self, config):
                    self.tag_migration = config['document_system']['tag_system']['tag_migration']
                    self.deprecated_tags = config['document_system']['tag_system']['deprecated_tags']
                    self.language_tags = config['document_system']['tag_system']['language_tags']

            return SimpleDocManager(config)
        return None

    def find_all_tags(self) -> List[TagInfo]:
        """모든 파일에서 TAG를 검색"""
        all_tags = []

        # 문서 파일만 검색
        doc_files = list(self.base_dir.glob("*.md"))
        doc_files.extend(self.base_dir.glob("README-*.md"))

        for file_path in doc_files:
            try:
                with open(file_path, 'r', encoding='utf-8') as f:
                    content = f.read()

                tags = self._extract_tags_from_content(content, str(file_path))
                all_tags.extend(tags)
            except (FileNotFoundError, PermissionError):
                continue

        return all_tags

    def _extract_tags_from_content(self, content: str, file_path: str) -> List[TagInfo]:
        """콘텐츠에서 TAG 추출"""
        tags = []

        # @CODE: 패턴 검색
        pattern = r'#\s*@CODE:([^\s\n]+)'
        matches = re.finditer(pattern, content)

        for match in matches:
            full_tag = match.group(0)
            tag_code = match.group(1)

            # TAG 분석
            tag_parts = tag_code.split(':')
            language = None

            if len(tag_parts) >= 3:
                # DOC-ONLINE-001:KO 형식
                category = tag_parts[0]
                spec_id = tag_parts[1]
                language = tag_parts[2]
            elif len(tag_parts) == 2:
                # DOCS-002 형식
                category = tag_parts[0]
                spec_id = tag_parts[1]
            else:
                # 단일 TAG
                category = tag_parts[0]
                spec_id = None

            # 라인 번호 계산
            line_number = content[:match.start()].count('\n') + 1

            tag_info = TagInfo(
                full_tag=full_tag,
                code=tag_code,
                language=language,
                category=category,
                line_number=line_number,
                file_path=file_path
            )
            tags.append(tag_info)

        return tags

    def analyze_tag_usage(self) -> Dict[str, any]:
        """TAG 사용 현 분석"""
        all_tags = self.find_all_tags()

        analysis = {
            'total_tags': len(all_tags),
            'unique_tags': len(set(tag.code for tag in all_tags)),
            'tag_by_category': {},
            'tag_by_language': {},
            'deprecated_tags': [],
            'duplicate_tags': [],
            'unused_tags': [],
            'files_with_tags': {}
        }

        # 카테고리별 분류
        for tag in all_tags:
            category = tag.category
            if category not in analysis['tag_by_category']:
                analysis['tag_by_category'][category] = 0
            analysis['tag_by_category'][category] += 1

            # 언어별 분류
            if tag.language:
                if tag.language not in analysis['tag_by_language']:
                    analysis['tag_by_language'][tag.language] = 0
                analysis['tag_by_language'][tag.language] += 1

            # 파일별 분류
            if tag.file_path not in analysis['files_with_tags']:
                analysis['files_with_tags'][tag.file_path] = []
            analysis['files_with_tags'][tag.file_path].append(tag.code)

        # 사용되지 않는 태그 식별
        if self.document_manager:
            expected_tags = set()
            expected_tags.update(self.document_manager.language_tags)
            expected_tags.add("DOC-ONLINE-001:MAIN")

            used_tags = set(tag.code for tag in all_tags)
            analysis['unused_tags'] = list(expected_tags - used_tags)

            # 사용되지 않는 예전 태그 식별
            analysis['deprecated_tags'] = [
                tag for tag in all_tags
                if tag.code in self.document_manager.deprecated_tags
            ]

            # 중복된 태그 식별
            tag_counts = {}
            for tag in all_tags:
                tag_counts[tag.code] = tag_counts.get(tag.code, 0) + 1

            analysis['duplicate_tags'] = [
                tag.code for tag, count in tag_counts.items()
                if count > 1
            ]

        return analysis

    def optimize_tags(self) -> Dict[str, any]:
        """TAG 최적화 수행"""
        analysis = self.analyze_tag_usage()
        optimization_results = {
            'files_processed': 0,
            'tags_migrated': 0,
            'tags_cleaned': 0,
            'errors': []
        }

        # 모든 문서 파일 처리
        doc_files = list(self.base_dir.glob("*.md"))
        doc_files.extend(self.base_dir.glob("README-*.md"))

        for file_path in doc_files:
            try:
                with open(file_path, 'r', encoding='utf-8') as f:
                    content = f.read()

                # TAG 마이그레이션
                migrated_count = self._migrate_tags_in_content(content)
                if migrated_count > 0:
                    optimization_results['tags_migrated'] += migrated_count

                # 사용되지 않는 태그 정리
                cleaned_count = self._clean_deprecated_tags_in_content(content)
                if cleaned_count > 0:
                    optimization_results['tags_cleaned'] += cleaned_count

                optimization_results['files_processed'] += 1

            except (FileNotFoundError, PermissionError) as e:
                optimization_results['errors'].append(f"Error processing {file_path}: {str(e)}")

        return optimization_results

    def _migrate_tags_in_content(self, content: str) -> int:
        """콘텐츠 내에서 TAG 마이그레이션"""
        if not self.document_manager:
            return 0

        migrated_count = 0
        for old_tag, new_tag in self.document_manager.tag_migration.items():
            if old_tag in content:
                content = content.replace(old_tag, new_tag)
                migrated_count += 1

        return migrated_count

    def _clean_deprecated_tags_in_content(self, content: str) -> int:
        """콘텐츠 내에서 사용되지 않는 태그 정리"""
        if not self.document_manager:
            return 0

        cleaned_count = 0
        for deprecated_tag in self.document_manager.deprecated_tags:
            if deprecated_tag in content:
                # 적절한 새 태그로 교체
                if deprecated_tag == "@CODE:DOCS-003":
                    content = content.replace(deprecated_tag, ":KO")
                elif deprecated_tag in ["@CODE:DOCS-001", "@CODE:DOCS-002"]:
                    content = content.replace(deprecated_tag, ":MAIN")
                cleaned_count += 1

        return cleaned_count

    def generate_tag_report(self) -> str:
        """TAG 사용 현황 보고서 생성"""
        analysis = self.analyze_tag_usage()

        report = f"""# TAG 시스템 현황 보고서

**생성일**: 2025-11-05
**대상 디렉토리**: {self.base_dir}
**총 파일 수**: {len(list(self.base_dir.glob("*.md")) + list(self.base_dir.glob("README-*.md")))}

## 📊 TAG 사용 현황

- **총 TAG 수**: {analysis['total_tags']}
- **고유 TAG 수**: {analysis['unique_tags']}
- **파일 수**: {len(analysis['files_with_tags'])}

### 🗂️ 카테고리별 분류
"""

        for category, count in analysis['tag_by_category'].items():
            report += f"- **{category}**: {count}개\n"

        if analysis['tag_by_language']:
            report += "\n### 🌍 언어별 분류\n"
            for language, count in analysis['tag_by_language'].items():
                report += f"- **{language}**: {count}개\n"

        if analysis['deprecated_tags']:
            report += "\n### ⚠️ 사용되지 않는 TAG\n"
            for tag in analysis['deprecated_tags']:
                report += f"- `{tag.code}` ({Path(tag.file_path).name}:{tag.line_number})\n"

        if analysis['duplicate_tags']:
            report += "\n### 🔄 중복된 TAG\n"
            for tag in analysis['duplicate_tags']:
                report += f"- `{tag}`\n"

        if analysis['unused_tags']:
            report += "\n### 📦 사용되지 않는 예상 TAG\n"
            for tag in analysis['unused_tags']:
                report += f"- `{tag}`\n"

        report += "\n## 📁 파일별 TAG 현황\n"
        for file_path, tags in analysis['files_with_tags'].items():
            file_name = Path(file_path).name
            report += f"### {file_name}\n"
            report += f"TAG 수: {len(tags)}\n"
            for tag in tags:
                report += f"- `{tag}`\n"
            report += "\n"

        return report

    def validate_tag_consistency(self) -> Dict[str, any]:
        """TAG 일관성 검사"""
        analysis = self.analyze_tag_usage()

        validation = {
            'is_consistent': True,
            'warnings': [],
            'errors': [],
            'recommendations': []
        }

        # 중복된 TAG 확인
        if analysis['duplicate_tags']:
            validation['warnings'].append(
                f"중복된 TAG 발견: {', '.join(analysis['duplicate_tags'])}"
            )
            validation['is_consistent'] = False

        # 사용되지 않는 예상 TAG 확인
        if analysis['unused_tags']:
            validation['recommendations'].append(
                f"사용되지 않는 예상 TAG: {', '.join(analysis['unused_tags'])}"
            )

        # 사용되지 않는 예전 TAG 확인
        if analysis['deprecated_tags']:
            validation['errors'].append(
                f"사용되지 않는 예전 TAG 발견: {', '.join([tag.code for tag in analysis['deprecated_tags']])}"
            )
            validation['is_consistent'] = False

        # 언어 TAG 균형 검사
        if analysis['tag_by_language']:
            max_count = max(analysis['tag_by_language'].values())
            min_count = min(analysis['tag_by_language'].values())

            if max_count - min_count > 2:
                validation['warnings'].append(
                    f"언어별 TAG 수 차이가 큼: {max_count} vs {min_count}"
                )

        return validation


# 인스턴스 생성
tag_optimizer = TagOptimizer()

# 함수들을 모듈 레벨에서 제공
def analyze_tag_usage() -> Dict[str, any]:
    """TAG 사용 현황 분석"""
    return tag_optimizer.analyze_tag_usage()

def optimize_tags() -> Dict[str, any]:
    """TAG 최적화 수행"""
    return tag_optimizer.optimize_tags()

def generate_tag_report() -> str:
    """TAG 시스템 보고서 생성"""
    return tag_optimizer.generate_tag_report()

def validate_tag_consistency() -> Dict[str, any]:
    """TAG 일관성 검사"""
    return tag_optimizer.validate_tag_consistency()