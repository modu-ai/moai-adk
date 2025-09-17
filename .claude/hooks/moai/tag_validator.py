#!/usr/bin/env python3
"""
MoAI-ADK Tag Validator PreToolUse Hook v0.1.17
16-Core @TAG 시스템 품질 검증 및 규칙 강제

이 Hook은 모든 파일 편집 시 @TAG 시스템의 품질을 자동으로 검증합니다.
- 16-Core 태그 체계 준수 검증
- 태그 네이밍 규칙 및 일관성 검사
- 품질 점수 계산 및 개선 제안
"""

import json
import sys
import re
import os
from pathlib import Path
from datetime import datetime
from typing import Dict, List, Tuple, Optional, Any

# Import security manager for safe operations
sys.path.insert(0, str(Path(__file__).parent.parent.parent.parent.parent / 'moai_adk'))
try:
    from security import SecurityManager, SecurityError
except ImportError:
    # Fallback if security module not available
    SecurityManager = None
    class SecurityError(Exception):
        pass

class MoAITagValidator:
    """MoAI-ADK 16-Core @TAG 시스템 검증기"""
    
    def __init__(self, project_root: Path):
        self.project_root = project_root
        self.security_manager = SecurityManager() if SecurityManager else None
        self.config_path = project_root / ".moai" / "config.json"
        self.tags_index_path = project_root / ".moai" / "indexes" / "tags.json"
        
        # 16-Core 태그 체계 정의
        self.tag_categories = {
            'Spec': ['REQ', 'SPEC', 'DESIGN', 'TASK'],
            'Steering': ['VISION', 'STRUCT', 'TECH', 'ADR'],
            'Implementation': ['FEATURE', 'API', 'TEST', 'DATA'],
            'Quality': ['PERF', 'SEC', 'DEBT', 'TODO'],
            'Legacy': ['US', 'FR', 'NFR', 'BUG', 'REVIEW']
        }
        
        # 모든 유효한 태그 타입
        self.valid_tag_types = []
        for category_tags in self.tag_categories.values():
            self.valid_tag_types.extend(category_tags)
            
        # 태그별 네이밍 규칙
        self.naming_rules = {
            # REQ:[CATEGORY]-[DESCRIPTION]-[NNN] → REQ:USER-LOGIN-001
            'REQ': r'^[A-Z]+-[A-Z0-9-]+-\d{3}$',
            'API': r'^(GET|POST|PUT|DELETE|PATCH)-.+$',      # API:GET-USERS
            'TEST': r'^(UNIT|INT|E2E|LOAD)-.+$',             # TEST:UNIT-LOGIN
            'PERF': r'^[A-Z]+-(\d{3}MS|FAST|SLOW)$',         # PERF:API-500MS
            'SEC': r'^[A-Z]+-(HIGH|MED|LOW)$',               # SEC:XSS-HIGH
            'BUG': r'^(CRITICAL|HIGH|MED|LOW)-\d{3}$',       # BUG:CRITICAL-001
        }

    def safe_regex_search(self, pattern: str, text: str, max_length: int = 10000) -> List[Tuple[str, str]]:
        """
        Safe regex search to prevent ReDoS attacks.

        Args:
            pattern: Regex pattern to search
            text: Text to search in
            max_length: Maximum text length to process

        Returns:
            List of matches as tuples
        """
        # Limit text length to prevent ReDoS
        if len(text) > max_length:
            if self.security_manager:
                # Log potential attack attempt
                print(f"Warning: Text too long for regex processing ({len(text)} > {max_length})", file=sys.stderr)
            text = text[:max_length]

        try:
            return re.findall(pattern, text)
        except re.error as e:
            print(f"Warning: Regex error in pattern {pattern}: {e}", file=sys.stderr)
            return []

    def safe_file_validation(self, file_path: str) -> bool:
        """
        Validate file path for security.

        Args:
            file_path: Path to validate

        Returns:
            bool: True if file is safe to process
        """
        if not self.security_manager:
            return True  # Skip validation if security manager unavailable

        try:
            path_obj = Path(file_path)

            # Check file size (max 1MB for tag validation)
            if not self.security_manager.validate_file_size(path_obj, 1):
                print(f"Warning: File too large for tag validation: {file_path}", file=sys.stderr)
                return False

            # Check if path is within project boundaries
            if not self.security_manager.validate_path_safety_enhanced(path_obj, self.project_root):
                print(f"Warning: File outside project scope: {file_path}", file=sys.stderr)
                return False

            return True
        except Exception as e:
            print(f"Warning: File validation error for {file_path}: {e}", file=sys.stderr)
            return False

    def validate_content(self, content: str, file_path: str) -> Dict[str, Any]:
        """파일 내용의 @TAG 검증"""

        # Validate file safety first
        if not self.safe_file_validation(file_path):
            return {
                'valid': False,
                'quality_score': 0.0,
                'message': 'File failed security validation'
            }

        # 태그 패턴 찾기 (@TYPE:ID 또는 @TYPE-ID 형식) - Safe regex search
        tag_pattern = r'@([A-Z]+)[-:]([A-Z0-9-_]+)'
        found_tags = self.safe_regex_search(tag_pattern, content)
        
        if not found_tags:
            return {
                'valid': True, 
                'quality_score': 1.0,
                'message': 'No tags found - validation skipped'
            }
        
        validation_results = []
        quality_issues = []
        
        for tag_type, tag_id in found_tags:
            result = self.validate_single_tag(tag_type, tag_id, file_path)
            validation_results.append(result)
            
            if not result['valid']:
                return {
                    'valid': False,
                    'error': result['error'],
                    'suggestion': result['suggestion'],
                    'quality_score': 0.0
                }
            
            if result.get('quality_issues'):
                quality_issues.extend(result['quality_issues'])
        
        # 품질 점수 계산
        quality_score = self.calculate_quality_score(found_tags, quality_issues)
        
        return {
            'valid': True,
            'quality_score': quality_score,
            'quality_issues': quality_issues,
            'tags_found': len(found_tags),
            'message': f'Validated {len(found_tags)} tags successfully'
        }
    
    def validate_single_tag(self, tag_type: str, tag_id: str, file_path: str) -> Dict[str, Any]:
        """단일 태그 검증"""
        
        # 1. 유효한 태그 타입 검증
        if tag_type not in self.valid_tag_types:
            return {
                'valid': False,
                'error': f"'{tag_type}' is not a valid 16-Core tag type",
                'suggestion': self.suggest_similar_tag(tag_type)
            }
        
        # 2. 태그 ID 형식 검증
        if not self.is_valid_tag_id(tag_id, tag_type):
            return {
                'valid': False,
                'error': f"'{tag_id}' doesn't match naming convention for {tag_type}",
                'suggestion': self.get_naming_example(tag_type)
            }
        
        # 3. 파일 경로와 태그의 일관성 검사
        consistency_check = self.check_file_tag_consistency(file_path, tag_type)
        
        quality_issues = []
        if not consistency_check['consistent']:
            quality_issues.append(consistency_check['warning'])
        
        return {
            'valid': True,
            'quality_issues': quality_issues
        }
    
    def is_valid_tag_id(self, tag_id: str, tag_type: str) -> bool:
        """태그 ID 형식 유효성 검사"""
        
        # 기본 형식: 대문자, 숫자, 하이픈, 언더스코어 허용
        basic_pattern = r'^[A-Z0-9-_]+$'
        if not re.match(basic_pattern, tag_id):
            return False
        
        # 특정 태그 타입의 특수 규칙 검사
        if tag_type in self.naming_rules:
            return bool(re.match(self.naming_rules[tag_type], tag_id))
        
        # 기본 규칙: 최소 2자, 최대 50자
        return 2 <= len(tag_id) <= 50
    
    def check_file_tag_consistency(self, file_path: str, tag_type: str) -> Dict[str, Any]:
        """파일 경로와 태그 타입의 일관성 검사"""
        
        consistency_rules = {
            'SPEC': ['.moai/specs/', 'spec.md'],
            'REQ': ['.moai/specs/', 'spec.md', 'requirements.md'],
            'DESIGN': ['plan.md', 'research.md', 'data-model.md', 'contracts/'],
            'TASK': ['.moai/specs/', 'tasks.md'],
            'TEST': ['test/', 'tests/', '__test__', '.test.'],
            'API': ['api/', 'routes/', 'endpoints/'],
            'DATA': ['models/', 'schema/', 'database/'],
            'ADR': ['.moai/memory/decisions', 'ADR']
        }
        
        if tag_type in consistency_rules:
            expected_paths = consistency_rules[tag_type]
            if not any(path in file_path for path in expected_paths):
                return {
                    'consistent': False,
                    'warning': f"{tag_type} tag usually belongs in files containing: {', '.join(expected_paths)}"
                }
        
        return {'consistent': True}
    
    def suggest_similar_tag(self, invalid_tag: str) -> str:
        """유사한 유효 태그 제안"""
        suggestions = []
        
        for valid_tag in self.valid_tag_types:
            # 간단한 유사도 계산 (편집 거리 기반)
            if self.levenshtein_distance(invalid_tag.lower(), valid_tag.lower()) <= 2:
                suggestions.append(valid_tag)
        
        if suggestions:
            return f"Did you mean: {', '.join(suggestions[:3])}"
        else:
            return f"Valid tag types: {', '.join(self.valid_tag_types[:5])}..."
    
    def get_naming_example(self, tag_type: str) -> str:
        """태그 타입별 네이밍 예시 제공"""
        examples = {
            'REQ': 'REQ:USER-LOGIN-001, REQ:PERF-RESPONSE-001',
            'SPEC': 'SPEC:AUTH-OVERVIEW, SPEC:CART-SCOPE',
            'DESIGN': 'DESIGN:AUTH-ARCH, DESIGN:PAYMENT-SEQ',
            'TASK': 'TASK:AUTH-SERVICE-001, TASK:CART-UI-002',
            'API': 'API:GET-USERS, API:POST-LOGIN',  
            'TEST': 'TEST:UNIT-AUTH, TEST:E2E-CHECKOUT',
            'PERF': 'PERF:API-500MS, PERF:DB-FAST',
            'SEC': 'SEC:XSS-HIGH, SEC:SQL-MED',
            'ADR': 'ADR:ARCH-DECISION-001',
            'BUG': 'BUG:CRITICAL-001, BUG:HIGH-002'
        }
        
        if tag_type in examples:
            return f"Example: {examples[tag_type]}"
        else:
            return f"Use format: {tag_type}:DESCRIPTION (uppercase, no spaces)"
    
    def calculate_quality_score(self, found_tags: List[Tuple[str, str]], quality_issues: List[str]) -> float:
        """태그 품질 점수 계산 (0.0 ~ 1.0)"""
        
        if not found_tags:
            return 1.0
        
        base_score = 1.0
        
        # 품질 이슈로 인한 감점
        issue_penalty = len(quality_issues) * 0.1
        base_score -= issue_penalty
        
        # 태그 다양성 보너스 (여러 카테고리 사용)
        used_categories = set()
        for tag_type, _ in found_tags:
            for category, types in self.tag_categories.items():
                if tag_type in types:
                    used_categories.add(category)
                    break
        
        diversity_bonus = len(used_categories) * 0.05
        base_score += diversity_bonus
        
        # 네이밍 규칙 준수 보너스
        rule_following_count = 0
        for tag_type, tag_id in found_tags:
            if tag_type in self.naming_rules and self.is_valid_tag_id(tag_id, tag_type):
                rule_following_count += 1
        
        if found_tags:
            rule_bonus = (rule_following_count / len(found_tags)) * 0.1
            base_score += rule_bonus
        
        return max(0.0, min(1.0, base_score))
    
    def levenshtein_distance(self, s1: str, s2: str) -> int:
        """두 문자열 간의 편집 거리 계산"""
        if len(s1) < len(s2):
            return self.levenshtein_distance(s2, s1)
        
        if len(s2) == 0:
            return len(s1)
        
        previous_row = list(range(len(s2) + 1))
        for i, c1 in enumerate(s1):
            current_row = [i + 1]
            for j, c2 in enumerate(s2):
                insertions = previous_row[j + 1] + 1
                deletions = current_row[j] + 1
                substitutions = previous_row[j] + (c1 != c2)
                current_row.append(min(insertions, deletions, substitutions))
            previous_row = current_row
        
        return previous_row[-1]

def main():
    """Hook 진입점"""
    
    try:
        # Claude Code Hook 데이터 읽기
        hook_data = json.loads(sys.stdin.read())
        
        tool_name = hook_data.get('tool_name', '')
        tool_input = hook_data.get('tool_input', {})
        
        # 파일 편집 도구에만 태그 검증 적용
        if tool_name not in ['Write', 'Edit', 'MultiEdit']:
            sys.exit(0)  # 다른 도구는 통과
        
        # 편집 내용 추출
        content = tool_input.get('content', '') or tool_input.get('new_string', '')
        file_path = tool_input.get('file_path', '')
        
        if not content or '@' not in content:
            sys.exit(0)  # 태그가 없으면 통과
        
        # 프로젝트 루트 경로
        project_root = Path(os.environ.get('CLAUDE_PROJECT_DIR', os.getcwd()))
        
        # 태그 검증 실행
        validator = MoAITagValidator(project_root)
        result = validator.validate_content(content, file_path)
        
        if not result['valid']:
            print("\n🏷️  16-Core @TAG 검증 실패", file=sys.stderr)
            if file_path:
                print(f"- 파일: {file_path}", file=sys.stderr)
            print(f"- 오류: {result['error']}", file=sys.stderr)
            if 'suggestion' in result and result['suggestion']:
                print(f"- 제안: {result['suggestion']}", file=sys.stderr)
            print("- 참고: docs/sections/12-tag-system.md (태그 규칙/예시)", file=sys.stderr)
            sys.exit(2)  # Hook 차단
        
        # 품질 피드백
        if result['quality_score'] >= 0.9:
            print(f"✨ 우수한 태그 품질! (점수: {result['quality_score']:.2f})", file=sys.stderr)
        elif result['quality_score'] < 0.7:
            print(f"⚠️  태그 품질 개선 필요 (점수: {result['quality_score']:.2f})", file=sys.stderr)
            if result.get('quality_issues'):
                for issue in result['quality_issues']:
                    print(f"   • {issue}", file=sys.stderr)
        
        sys.exit(0)  # 검증 통과
        
    except Exception as error:
        print(f"🔧 Tag validator error: {error}", file=sys.stderr)
        sys.exit(0)  # 오류 시에도 통과 (검증 실패로 개발 차단 방지)

if __name__ == "__main__":
    main()
