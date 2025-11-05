#!/usr/bin/env python3
"""
MoAI-ADK TAG De-duplication Detector
GPT-5 Pro analysis 기반 TAG 중복 탐지 및 수정 시스템
"""

import json
import re
import os
import sys
from pathlib import Path
from typing import Dict, List, Tuple, Set, Optional
from collections import defaultdict, Counter
from dataclasses import dataclass
from datetime import datetime

@dataclass
class TagInfo:
    """TAG 정보 구조체"""
    tag: str
    file_path: str
    line_number: int
    line_content: str
    is_topline: bool
    tag_type: str  # SPEC, TEST, CODE, DOC
    domain: str
    id_number: str
    authority_score: int

@dataclass
class DuplicateGroup:
    """중복 그룹 정보"""
    tag_pattern: str
    occurrences: List[TagInfo]
    primary_candidate: Optional[TagInfo] = None
    duplicates: List[TagInfo] = None
    
    def __post_init__(self):
        if self.duplicates is None:
            self.duplicates = []

class TagDedupDetector:
    """TAG 중복 탐지기"""
    
    def __init__(self, config_path: str = None):
        self.project_root = Path.cwd()
        self.config = self._load_config(config_path)
        self.tag_pattern = re.compile(r'@([A-Z]+):([A-Z_]+)-([0-9]{3,})')
        self.all_tags: List[TagInfo] = []
        self.duplicate_groups: List[DuplicateGroup] = []
        
    def _load_config(self, config_path: str = None) -> Dict:
        """설정 파일 로드"""
        if config_path is None:
            config_path = self.project_root / ".moai" / "tag-policy-updated.json"
        
        try:
            with open(config_path, 'r', encoding='utf-8') as f:
                return json.load(f)
        except FileNotFoundError:
            print(f"❌ 설정 파일을 찾을 수 없습니다: {config_path}")
            sys.exit(1)
    
    def _calculate_authority_score(self, file_path: str, line_number: int) -> int:
        """파일 권한 점수 계산"""
        path = Path(file_path)
        authority_hierarchy = self.config.get("topline_rules", {}).get("authority_hierarchy", {})
        
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
    
    def _is_topline_tag(self, line_content: str, line_number: int) -> bool:
        """Topline TAG 여부 확인"""
        if line_number > 20:  # topline은 처음 20줄 이내
            return False
            
        # 주석 블록 스킵 확인
        skip_patterns = self.config.get("topline_rules", {}).get("skip_header_block", [])
        for pattern in skip_patterns:
            if pattern.lower() in line_content.lower():
                return False
                
        return True
    
    def _extract_tags_from_file(self, file_path: str) -> List[TagInfo]:
        """파일에서 TAG 추출"""
        tags = []
        
        try:
            with open(file_path, 'r', encoding='utf-8') as f:
                lines = f.readlines()
                
            for line_num, line in enumerate(lines, 1):
                matches = self.tag_pattern.finditer(line)
                for match in matches:
                    tag_type, domain, id_num = match.groups()
                    full_tag = match.group(0)
                    
                    tag_info = TagInfo(
                        tag=full_tag,
                        file_path=file_path,
                        line_number=line_num,
                        line_content=line.strip(),
                        is_topline=self._is_topline_tag(line, line_num),
                        tag_type=tag_type,
                        domain=domain,
                        id_number=id_num,
                        authority_score=self._calculate_authority_score(file_path, line_num)
                    )
                    tags.append(tag_info)
                    
        except Exception as e:
            print(f"⚠️  파일 처리 오류 {file_path}: {e}")
            
        return tags
    
    def scan_all_files(self) -> None:
        """전체 파일 스캔"""
        print("🔍 전체 파일 스캔 시작...")
        
        eligible_patterns = self.config.get("eligible", [])
        excluded_patterns = self.config.get("excluded", [])
        
        for pattern in eligible_patterns:
            for file_path in self.project_root.glob(pattern):
                # 제외 패턴 확인
                if any(file_path.match(exclude) for exclude in excluded_patterns):
                    continue
                    
                if file_path.is_file():
                    file_tags = self._extract_tags_from_file(str(file_path))
                    self.all_tags.extend(file_tags)
        
        print(f"✅ {len(self.all_tags)}개의 TAG 발견")
    
    def find_duplicates(self) -> None:
        """중복 TAG 탐지"""
        print("🔍 중복 TAG 탐지 시작...")
        
        # TAG 패턴별로 그룹화
        tag_groups = defaultdict(list)
        for tag in self.all_tags:
            # TAG 패턴 정규화 (타입:도메인-ID)
            pattern = f"{tag.tag_type}:{tag.domain}-{tag.id_number}"
            tag_groups[pattern].append(tag)
        
        # 중복 그룹 식별
        for pattern, occurrences in tag_groups.items():
            if len(occurrences) > 1:
                # Primary 후보 선택 (최고 권한 점수)
                primary_candidate = max(occurrences, key=lambda x: x.authority_score)
                
                # 나머지는 중복으로 분류
                duplicates = [tag for tag in occurrences if tag != primary_candidate]
                
                duplicate_group = DuplicateGroup(
                    tag_pattern=pattern,
                    occurrences=occurrences,
                    primary_candidate=primary_candidate,
                    duplicates=duplicates
                )
                self.duplicate_groups.append(duplicate_group)
        
        print(f"✅ {len(self.duplicate_groups)}개의 중복 그룹 발견")
    
    def analyze_duplicates(self) -> Dict:
        """중복 분석 리포트 생성 (Reference duplicates 구분)"""
        analysis = {
            "summary": {
                "total_tags": len(self.all_tags),
                "unique_patterns": len(set(f"{t.tag_type}:{t.domain}-{t.id_number}" for t in self.all_tags)),
                "duplicate_groups": len(self.duplicate_groups),
                "total_duplicate_occurrences": sum(len(g.duplicates) for g in self.duplicate_groups),
                "topline_duplicates": 0,
                "reference_duplicates": 0,
                "allowed_reference_duplicates": 0,
                "critical_duplicates": 0
            },
            "duplicate_groups": [],
            "recommended_actions": []
        }

        # 정책 설정 확인
        ref_dup_allowed = self.config.get("deduplication_policy", {}).get("reference_duplicates", {}).get("allowed", False)

        for group in self.duplicate_groups:
            # 중복 유형 분류
            has_topline = group.primary_candidate.is_topline or any(dup.is_topline for dup in group.duplicates)

            group_info = {
                "tag_pattern": group.tag_pattern,
                "total_occurrences": len(group.occurrences),
                "duplicate_type": "topline" if has_topline else "reference",
                "severity": "critical" if has_topline else ("info" if ref_dup_allowed else "warning"),
                "primary_candidate": {
                    "file": group.primary_candidate.file_path,
                    "line": group.primary_candidate.line_number,
                    "authority_score": group.primary_candidate.authority_score,
                    "is_topline": group.primary_candidate.is_topline
                },
                "duplicates": [
                    {
                        "file": dup.file_path,
                        "line": dup.line_number,
                        "authority_score": dup.authority_score,
                        "is_topline": dup.is_topline
                    }
                    for dup in group.duplicates
                ]
            }
            analysis["duplicate_groups"].append(group_info)

            # 중복 유형별 카운트
            if has_topline:
                analysis["summary"]["topline_duplicates"] += 1
                analysis["summary"]["critical_duplicates"] += 1
            else:
                if ref_dup_allowed:
                    analysis["summary"]["allowed_reference_duplicates"] += 1
                else:
                    analysis["summary"]["reference_duplicates"] += 1

        # 추천 액션 생성 (정책 기반)
        if analysis["summary"]["topline_duplicates"] > 0:
            analysis["recommended_actions"].append({
                "priority": "critical",
                "action": "renumber_or_remove_topline_duplicates",
                "description": f"{analysis['summary']['topline_duplicates']}개의 topline 중복은 즉시 처리 필요"
            })

        # Reference duplicates 처리
        if ref_dup_allowed:
            if analysis["summary"]["allowed_reference_duplicates"] > 0:
                analysis["recommended_actions"].append({
                    "priority": "info",
                    "action": "acknowledge_reference_duplicates",
                    "description": f"{analysis['summary']['allowed_reference_duplicates']}개의 참조 중복은 정책상 허용됨 (TRACEABILITY용)"
                })
        else:
            if analysis["summary"]["reference_duplicates"] > 0:
                analysis["recommended_actions"].append({
                    "priority": "warning",
                    "action": "review_reference_duplicates",
                    "description": f"{analysis['summary']['reference_duplicates']}개의 참조 중복 검토 필요"
                })

        # 정책 준수 여부 판단
        analysis["policy_compliance"] = {
            "compliant": analysis["summary"]["critical_duplicates"] == 0,
            "critical_issues": analysis["summary"]["topline_duplicates"],
            "allowed_reference_count": analysis["summary"]["allowed_reference_duplicates"]
        }

        return analysis
    
    def generate_correction_plan(self, analysis: Dict) -> Dict:
        """수정 계획 생성"""
        correction_plan = {
            "timestamp": datetime.now().isoformat(),
            "dry_run": True,
            "corrections": [],
            "rollback_info": {
                "backup_strategy": "git_tag",
                "restore_command": "git checkout TAG-DEDUP-BACKUP"
            }
        }
        
        for group in self.duplicate_groups:
            # Primary만 유지하고 나머지는 수정/삭제
            primary = group.primary_candidate
            
            for duplicate in group.duplicates:
                if duplicate.is_topline:
                    # Topline 중복은 재번호화 또는 제거
                    new_tag = self._generate_new_tag(duplicate)
                    correction = {
                        "type": "renumber",
                        "file_path": duplicate.file_path,
                        "line_number": duplicate.line_number,
                        "old_tag": duplicate.tag,
                        "new_tag": new_tag,
                        "confidence": 0.9,
                        "impact": "medium"
                    }
                    correction_plan["corrections"].append(correction)
                else:
                    # 참조 중복은 추적만
                    correction = {
                        "type": "track_only",
                        "file_path": duplicate.file_path,
                        "line_number": duplicate.line_number,
                        "tag": duplicate.tag,
                        "note": "Reference duplicate - no action needed"
                    }
                    correction_plan["corrections"].append(correction)
        
        return correction_plan
    
    def _generate_new_tag(self, tag: TagInfo) -> str:
        """새로운 TAG 생성"""
        # 도메인별 카운터 관리 (간단한 구현)
        domain_counters = defaultdict(int)
        
        for existing_tag in self.all_tags:
            if existing_tag.domain == tag.domain and existing_tag.tag_type == tag.tag_type:
                domain_counters[existing_tag.domain] += 1
        
        new_id = domain_counters[tag.domain] + 1
        new_tag = f"@{tag.tag_type}:{tag.domain}-{new_id:03d}"
        
        return new_tag
    
    def save_analysis_report(self, analysis: Dict, output_path: str = None) -> None:
        """분석 리포트 저장"""
        if output_path is None:
            output_path = self.project_root / ".moai" / "reports" / f"tag-dedup-analysis-{datetime.now().strftime('%Y%m%d-%H%M%S')}.json"
        
        os.makedirs(output_path.parent, exist_ok=True)
        
        with open(output_path, 'w', encoding='utf-8') as f:
            json.dump(analysis, f, indent=2, ensure_ascii=False)
        
        print(f"📊 분석 리포트 저장: {output_path}")
    
    def run_full_scan(self) -> Dict:
        """전체 스캔 실행"""
        print("🚀 TAG 중복 탐지 시작...")
        print(f"📁 프로젝트 루트: {self.project_root}")
        
        # 1. 전체 파일 스캔
        self.scan_all_files()
        
        # 2. 중복 탐지
        self.find_duplicates()
        
        # 3. 분석 리포트 생성
        analysis = self.analyze_duplicates()
        
        # 4. 리포트 저장
        self.save_analysis_report(analysis)
        
        # 5. 요약 출력
        self._print_summary(analysis)
        
        return analysis
    
    def _print_summary(self, analysis: Dict) -> None:
        """요약 정보 출력"""
        summary = analysis["summary"]
        
        print("\n" + "="*60)
        print("📊 TAG 중복 분석 요약")
        print("="*60)
        print(f"총 TAG 수: {summary['total_tags']}")
        print(f"고유 TAG 패턴: {summary['unique_patterns']}")
        print(f"중복 그룹: {summary['duplicate_groups']}")
        print(f"총 중복 발생: {summary['total_duplicate_occurrences']}")
        print(f"Topline 중복: {summary['topline_duplicates']} (치명적)")
        print(f"참조 중복: {summary['reference_duplicates']} (정보)")
        
        print("\n🎯 추천 액션:")
        for action in analysis["recommended_actions"]:
            priority_emoji = "🔴" if action["priority"] == "critical" else "🔵"
            print(f"{priority_emoji} {action['description']}")
        
        print("\n📈 상위 중복 그룹:")
        for group in sorted(analysis["duplicate_groups"], key=lambda x: x["total_occurrences"], reverse=True)[:5]:
            print(f"  • {group['tag_pattern']}: {group['total_occurrences']}회 발생")

def main():
    """메인 함수"""
    import argparse
    
    parser = argparse.ArgumentParser(description="MoAI-ADK TAG 중복 탐지기")
    parser.add_argument("--config", help="설정 파일 경로")
    parser.add_argument("--output", help="출력 파일 경로")
    parser.add_argument("--dry-run", action="store_true", default=True, help="시뮬레이션 모드")
    
    args = parser.parse_args()
    
    try:
        detector = TagDedupDetector(args.config)
        analysis = detector.run_full_scan()
        
        if analysis["summary"]["topline_duplicates"] > 0:
            print(f"\n⚠️  {analysis['summary']['topline_duplicates']}개의 치명적 topline 중복 발견!")
            print("수정을 위해서는 --apply-corrections 옵션을 사용하세요.")
            return 1
        else:
            print("\n✅ 치명적 중복 없음")
            return 0
            
    except Exception as e:
        print(f"❌ 오류 발생: {e}")
        return 1

if __name__ == "__main__":
    sys.exit(main())
