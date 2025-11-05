#!/usr/bin/env python3
"""
Enhanced TAG De-duplication PreToolUse Hook
GPT-5 Pro 분석 기반 실시간 TAG 중복 검증 시스템
"""

import json
import os
import re
import sys
import time
from datetime import datetime, timedelta
from pathlib import Path
from typing import Dict, List, Optional, Tuple, Any

# Configuration paths
POLICY_PATH = Path(".moai/tag-policy-updated.json")
DEDUP_POLICY_PATH = Path(".moai/tag-dedup-policy.json")
LEDGER_PATH = Path(".moai/tags/ledger.jsonl")
INDEX_PATH = Path(".moai/tags/index.json")
DUPLICATE_CACHE_PATH = Path(".moai/cache/duplicate_cache.json")

class TagDedupValidator:
    """TAG 중복 검증기"""
    
    def __init__(self):
        self.policy = self._load_policy()
        self.dedup_policy = self._load_dedup_policy()
        self.duplicate_cache = self._load_duplicate_cache()
        
    def _load_policy(self) -> Dict[str, Any]:
        """TAG 정책 로드"""
        if POLICY_PATH.exists():
            with open(POLICY_PATH, 'r', encoding='utf-8') as f:
                return json.load(f)
        return {}
    
    def _load_dedup_policy(self) -> Dict[str, Any]:
        """중복 제거 정책 로드"""
        if DEDUP_POLICY_PATH.exists():
            with open(DEDUP_POLICY_PATH, 'r', encoding='utf-8') as f:
                return json.load(f)
        return {}
    
    def _load_duplicate_cache(self) -> Dict[str, Any]:
        """중복 캐시 로드"""
        if DUPLICATE_CACHE_PATH.exists():
            with open(DUPLICATE_CACHE_PATH, 'r', encoding='utf-8') as f:
                cache = json.load(f)
                # 1시간 이상 된 캐시는 초기화
                cache_time = datetime.fromisoformat(cache.get("timestamp", "1970-01-01"))
                if datetime.now() - cache_time > timedelta(hours=1):
                    return {"timestamp": datetime.now().isoformat(), "duplicates": {}}
                return cache
        return {"timestamp": datetime.now().isoformat(), "duplicates": {}}
    
    def _save_duplicate_cache(self) -> None:
        """중복 캐시 저장"""
        DUPLICATE_CACHE_PATH.parent.mkdir(exist_ok=True)
        with open(DUPLICATE_CACHE_PATH, 'w', encoding='utf-8') as f:
            json.dump(self.duplicate_cache, f, indent=2)
    
    def extract_tags_from_content(self, content: str, file_path: str) -> List[Dict[str, Any]]:
        """내용에서 TAG 추출"""
        tag_pattern = re.compile(r'@([A-Z]+):([A-Z_]+)-([0-9]{3,})')
        lines = content.split('\n')
        tags = []
        
        for line_num, line in enumerate(lines, 1):
            matches = tag_pattern.finditer(line)
            for match in matches:
                tag_type, domain, id_num = match.groups()
                full_tag = match.group(0)
                
                # Topline 여부 확인 (처음 20줄)
                is_topline = line_num <= 20
                
                # 권한 점수 계산
                authority_score = self._calculate_authority_score(file_path, line_num)
                
                tag_info = {
                    "tag": full_tag,
                    "tag_type": tag_type,
                    "domain": domain,
                    "id_number": id_num,
                    "line_number": line_num,
                    "is_topline": is_topline,
                    "authority_score": authority_score,
                    "pattern": f"{tag_type}:{domain}-{id_num}"
                }
                tags.append(tag_info)
        
        return tags
    
    def _calculate_authority_score(self, file_path: str, line_number: int) -> int:
        """파일 권한 점수 계산"""
        path = Path(file_path)
        authority_hierarchy = self.policy.get("topline_rules", {}).get("authority_hierarchy", {})
        
        # 최고 권한
        for pattern in authority_hierarchy.get("highest", []):
            if path.match(pattern):
                return 100
                
        # 높은 권한
        for pattern in authority_hierarchy.get("high", []):
            if path.match(pattern):
                return 80
                
        # 중간 권한
        for pattern in authority_hierarchy.get("medium", []):
            if path.match(pattern):
                return 60
                
        # 낮은 권한
        for pattern in authority_hierarchy.get("low", []):
            if path.match(pattern):
                return 40
                
        # 기본 권한
        return 20
    
    def _is_whitelisted_path(self, file_path: str) -> bool:
        """화이트리스트 경로 확인"""
        path = Path(file_path)
        whitelist = self.dedup_policy.get("gpt5_pro_analysis", {}).get("exception_whitelist", {}).get("paths", [])
        
        for pattern in whitelist:
            if path.match(pattern):
                return True
        
        return False
    
    def validate_topline_uniqueness(self, new_tags: List[Dict[str, Any]], file_path: str) -> List[Dict[str, Any]]:
        """Topline 유일성 검증"""
        violations = []
        
        # 화이트리스트 경로는 건너뜀
        if self._is_whitelisted_path(file_path):
            return violations
        
        for new_tag in new_tags:
            if not new_tag["is_topline"]:
                continue
            
            pattern = new_tag["pattern"]
            
            # 캐시에서 중복 확인
            if pattern in self.duplicate_cache["duplicates"]:
                existing = self.duplicate_cache["duplicates"][pattern]
                
                # 더 높은 권한을 가진 파일이 있는지 확인
                if existing["authority_score"] >= new_tag["authority_score"]:
                    violation = {
                        "type": "topline_duplicate",
                        "severity": "critical",
                        "tag": new_tag["tag"],
                        "pattern": pattern,
                        "file_path": file_path,
                        "line_number": new_tag["line_number"],
                        "existing_file": existing["file_path"],
                        "existing_authority": existing["authority_score"],
                        "new_authority": new_tag["authority_score"],
                        "message": f"Topline duplicate detected: {new_tag['tag']} already exists in {existing['file_path']} with higher authority",
                        "suggestion": "Use /alfred:tag-dedup to resolve duplicates"
                    }
                    violations.append(violation)
        
        return violations
    
    def update_duplicate_cache(self, new_tags: List[Dict[str, Any]], file_path: str) -> None:
        """중복 캐시 업데이트"""
        for tag in new_tags:
            if tag["is_topline"]:
                pattern = tag["pattern"]
                
                # 캐시에 없거나 더 높은 권한을 가진 TAG면 업데이트
                if (pattern not in self.duplicate_cache["duplicates"] or 
                    tag["authority_score"] > self.duplicate_cache["duplicates"][pattern]["authority_score"]):
                    
                    self.duplicate_cache["duplicates"][pattern] = {
                        "tag": tag["tag"],
                        "file_path": file_path,
                        "line_number": tag["line_number"],
                        "authority_score": tag["authority_score"],
                        "timestamp": datetime.now().isoformat()
                    }
        
        self._save_duplicate_cache()
    
    def validate_chain_integrity(self, new_tags: List[Dict[str, Any]], file_path: str) -> List[Dict[str, Any]]:
        """TAG 체인 무결성 검증"""
        violations = []
        
        # TODO: 체인 무결성 검증 로직 구현
        # SPEC → TEST → CODE → DOC 연결 확인
        
        return violations
    
    def validate_scope_compliance(self, new_tags: List[Dict[str, Any]], file_path: str) -> List[Dict[str, Any]]:
        """스코프 규정 준수 검증"""
        violations = []
        
        monorepo_scoping = self.policy.get("monorepo_scoping", {})
        if not monorepo_scoping.get("enabled", False):
            return violations
        
        for tag in new_tags:
            # 패키지 스코프핑 규칙 확인
            if "#" not in tag["tag"] and tag["is_topline"]:
                # 스코프가 없는 topline TAG 경고
                violation = {
                    "type": "scope_violation",
                    "severity": "warning",
                    "tag": tag["tag"],
                    "file_path": file_path,
                    "line_number": tag["line_number"],
                    "message": f"Package-scoped tag format recommended: {tag['tag']} should use domain scoping",
                    "suggestion": f"Consider using format like @TYPE:DOMAIN#PACKAGE-{tag['id_number']}"
                }
                violations.append(violation)
        
        return violations
    
    def validate_tags(self, content: str, file_path: str) -> Dict[str, Any]:
        """TAG 검증 실행"""
        try:
            # TAG 추출
            new_tags = self.extract_tags_from_content(content, file_path)
            
            if not new_tags:
                return {"valid": True, "violations": []}
            
            # 검증 실행
            violations = []
            
            # 1. Topline 유일성 검증
            topline_violations = self.validate_topline_uniqueness(new_tags, file_path)
            violations.extend(topline_violations)
            
            # 2. 체인 무결성 검증
            chain_violations = self.validate_chain_integrity(new_tags, file_path)
            violations.extend(chain_violations)
            
            # 3. 스코프 규정 준수 검증
            scope_violations = self.validate_scope_compliance(new_tags, file_path)
            violations.extend(scope_violations)
            
            # 4. 캐시 업데이트 (위반 없는 경우만)
            if not any(v["severity"] == "critical" for v in violations):
                self.update_duplicate_cache(new_tags, file_path)
            
            return {
                "valid": len(violations) == 0,
                "violations": violations,
                "tags_found": len(new_tags),
                "topline_tags": sum(1 for t in new_tags if t["is_topline"])
            }
            
        except Exception as e:
            return {
                "valid": False,
                "violations": [{
                    "type": "validation_error",
                    "severity": "error",
                    "message": f"TAG validation failed: {str(e)}",
                    "file_path": file_path
                }]
            }

def main():
    """메인 함수"""
    # 인자 파싱
    if len(sys.argv) < 4:
        print("Usage: PreToolUse.py <tool_name> <file_path> <content>")
        sys.exit(1)
    
    tool_name = sys.argv[1]
    file_path = sys.argv[2]
    content = sys.argv[3]
    
    # 파일 타입 확인
    file_extension = Path(file_path).suffix.lower()
    supported_extensions = ['.py', '.md', '.js', '.ts', '.jsx', '.tsx', '.java', '.go', '.rs']
    
    if file_extension not in supported_extensions:
        # 지원하지 않는 파일 타입은 통과
        print(json.dumps({"valid": True, "violations": []}))
        sys.exit(0)
    
    # 검증기 생성 및 실행
    validator = TagDedupValidator()
    result = validator.validate_tags(content, file_path)
    
    # 결과 출력
    print(json.dumps(result, indent=2))
    
    # 치명적 위반이 있으면 비제로 종료
    critical_violations = [v for v in result["violations"] if v["severity"] == "critical"]
    if critical_violations:
        print(f"\n🚨 Critical TAG violations detected:", file=sys.stderr)
        for violation in critical_violations:
            print(f"  • {violation['message']}", file=sys.stderr)
            print(f"    Suggestion: {violation.get('suggestion', 'Contact administrator')}", file=sys.stderr)
        sys.exit(1)
    
    sys.exit(0)

if __name__ == "__main__":
    main()
