#!/usr/bin/env python3
"""
MoAI-ADK TAG Auto-Correction System
GPT-5 Pro 권장사항 기반 자동 TAG 수정 시스템
"""

import json
import re
import os
import sys
import subprocess
from pathlib import Path
from typing import Dict, List, Optional, Tuple
from dataclasses import dataclass
from datetime import datetime
from tag_dedup_detector import TagDedupDetector, TagInfo, DuplicateGroup

@dataclass
class CorrectionAction:
    """수정 액션 정보"""
    action_type: str  # renumber, remove, update_reference
    file_path: str
    line_number: int
    old_tag: str
    new_tag: Optional[str]
    confidence: float
    impact: str  # low, medium, high, critical
    description: str

@dataclass 
class CorrectionResult:
    """수정 결과 정보"""
    action: CorrectionAction
    success: bool
    error_message: Optional[str] = None
    backup_path: Optional[str] = None

class TagAutoCorrector:
    """TAG 자동 수정기"""
    
    def __init__(self, config_path: str = None, dry_run: bool = True):
        self.project_root = Path.cwd()
        self.config = self._load_config(config_path)
        self.dry_run = dry_run
        self.detector = TagDedupDetector(config_path)
        self.corrections: List[CorrectionAction] = []
        self.results: List[CorrectionResult] = []
        
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
    
    def _create_backup(self) -> str:
        """백업 생성"""
        timestamp = datetime.now().strftime("%Y%m%d-%H%M%S")
        backup_tag = f"TAG-DEDUP-BACKUP-{timestamp}"
        
        try:
            # Git 태그 생성
            subprocess.run(
                ["git", "tag", backup_tag],
                check=True,
                capture_output=True,
                text=True
            )
            print(f"✅ 백업 생성: {backup_tag}")
            return backup_tag
        except subprocess.CalledProcessError as e:
            print(f"⚠️  Git 태그 생성 실패: {e}")
            return None
    
    def _validate_correction(self, action: CorrectionAction) -> Tuple[bool, str]:
        """수정 액션 검증"""
        # 파일 존재 확인
        if not os.path.exists(action.file_path):
            return False, f"파일이 존재하지 않음: {action.file_path}"
        
        # 신뢰도 확인
        if action.confidence < self.config.get("auto_correction", {}).get("confidence_threshold", 0.8):
            return False, f"신뢰도 부족: {action.confidence}"
        
        # 영향도 확인
        if action.impact == "critical" and self.dry_run:
            return False, "치명적 영향도 - dry-run 모드에서 건너뜀"
        
        return True, "검증 통과"
    
    def _apply_line_correction(self, action: CorrectionAction) -> bool:
        """라인 단위 수정 적용"""
        try:
            file_path = Path(action.file_path)
            
            # 파일 읽기
            with open(file_path, 'r', encoding='utf-8') as f:
                lines = f.readlines()
            
            # 해당 라인 수정
            if action.line_number <= len(lines):
                old_line = lines[action.line_number - 1]
                
                if action.new_tag:
                    # TAG 교체
                    new_line = old_line.replace(action.old_tag, action.new_tag)
                else:
                    # TAG 제거
                    new_line = re.sub(r'\s*' + re.escape(action.old_tag), '', old_line)
                
                lines[action.line_number - 1] = new_line
                
                # 파일 쓰기
                with open(file_path, 'w', encoding='utf-8') as f:
                    f.writelines(lines)
                
                print(f"✅ 수정 완료: {action.file_path}:{action.line_number}")
                return True
            else:
                print(f"❌ 라인 번호 초과: {action.file_path}:{action.line_number}")
                return False
                
        except Exception as e:
            print(f"❌ 수정 실패: {action.file_path}:{action.line_number} - {e}")
            return False
    
    def generate_corrections(self, duplicate_groups: List[DuplicateGroup]) -> None:
        """수정 계획 생성"""
        print("🔧 수정 계획 생성 중...")
        
        for group in duplicate_groups:
            primary = group.primary_candidate
            
            for duplicate in group.duplicates:
                # Topline 중복 처리
                if duplicate.is_topline:
                    # 새로운 TAG 생성
                    new_tag = self._generate_new_tag(duplicate.tag, duplicate.domain, duplicate.tag_type)
                    
                    action = CorrectionAction(
                        action_type="renumber",
                        file_path=duplicate.file_path,
                        line_number=duplicate.line_number,
                        old_tag=duplicate.tag,
                        new_tag=new_tag,
                        confidence=0.9,
                        impact="high",
                        description=f"Topline 중복 처리: {duplicate.tag} → {new_tag}"
                    )
                    self.corrections.append(action)
                    
                    # 관련 참조 업데이트 필요 확인
                    self._plan_reference_updates(duplicate.tag, new_tag)
                
                # 참조 중복 처리 (높은 권한 파일의 참조만)
                elif duplicate.authority_score >= 60:
                    action = CorrectionAction(
                        action_type="update_reference",
                        file_path=duplicate.file_path,
                        line_number=duplicate.line_number,
                        old_tag=duplicate.tag,
                        new_tag=primary.tag,
                        confidence=0.8,
                        impact="medium",
                        description=f"참조 업데이트: {duplicate.tag} → {primary.tag}"
                    )
                    self.corrections.append(action)
        
        print(f"✅ {len(self.corrections)}개의 수정 액션 생성 완료")
    
    def _plan_reference_updates(self, old_tag: str, new_tag: str) -> None:
        """관련 참조 업데이트 계획"""
        # 전체 파일에서 old_tag 참조 검색
        for tag_info in self.detector.all_tags:
            if tag_info.tag == old_tag and not tag_info.is_topline:
                action = CorrectionAction(
                    action_type="update_reference",
                    file_path=tag_info.file_path,
                    line_number=tag_info.line_number,
                    old_tag=old_tag,
                    new_tag=new_tag,
                    confidence=0.9,
                    impact="medium",
                    description=f"관련 참조 업데이트: {old_tag} → {new_tag}"
                )
                self.corrections.append(action)
    
    def _generate_new_tag(self, old_tag: str, domain: str, tag_type: str) -> str:
        """새로운 TAG 생성"""
        # 기존 최대 ID 찾기
        max_id = 0
        tag_pattern = re.compile(rf'@{tag_type}:{domain}-(\d+)')
        
        for tag_info in self.detector.all_tags:
            match = tag_pattern.search(tag_info.tag)
            if match:
                id_num = int(match.group(1))
                max_id = max(max_id, id_num)
        
        # 새로운 ID 생성
        new_id = max_id + 1
        new_tag = f"@{tag_type}:{domain}-{new_id:03d}"
        
        return new_tag
    
    def apply_corrections(self) -> List[CorrectionResult]:
        """수정 적용"""
        print("🔧 수정 적용 중...")
        
        if self.dry_run:
            print("🔍 DRY RUN 모드 - 실제 수정하지 않음")
        
        # 백업 생성
        backup_tag = self._create_backup() if not self.dry_run else None
        
        # 수정 적용
        for action in self.corrections:
            # 검증
            is_valid, message = self._validate_correction(action)
            if not is_valid:
                print(f"⚠️  검증 실패: {message}")
                continue
            
            # 적용
            if self.dry_run:
                print(f"[DRY RUN] {action.description}")
                result = CorrectionResult(
                    action=action,
                    success=True,
                    backup_path=backup_tag
                )
            else:
                success = self._apply_line_correction(action)
                result = CorrectionResult(
                    action=action,
                    success=success,
                    backup_path=backup_tag
                )
            
            self.results.append(result)
        
        # 결과 요약
        self._print_results_summary()
        
        return self.results
    
    def _print_results_summary(self) -> None:
        """결과 요약 출력"""
        total = len(self.results)
        successful = sum(1 for r in self.results if r.success)
        failed = total - successful
        
        print(f"\n📊 수정 결과 요약:")
        print(f"총 수정: {total}")
        print(f"성공: {successful}")
        print(f"실패: {failed}")
        
        if failed > 0:
            print(f"\n❌ 실패한 수정:")
            for result in self.results:
                if not result.success:
                    print(f"  • {result.action.file_path}:{result.action.line_number} - {result.error_message}")
        
        # 영향도별 요약
        impact_summary = {}
        for result in self.results:
            impact = result.action.impact
            if impact not in impact_summary:
                impact_summary[impact] = 0
            impact_summary[impact] += 1
        
        print(f"\n📈 영향도별 요약:")
        for impact, count in sorted(impact_summary.items()):
            emoji = {"low": "🟢", "medium": "🟡", "high": "🟠", "critical": "🔴"}.get(impact, "⚪")
            print(f"  {emoji} {impact}: {count}")
    
    def save_correction_report(self, output_path: str = None) -> None:
        """수정 리포트 저장"""
        if output_path is None:
            timestamp = datetime.now().strftime("%Y%m%d-%H%M%S")
            output_path = self.project_root / ".moai" / "reports" / f"tag-correction-{timestamp}.json"
        
        os.makedirs(output_path.parent, exist_ok=True)
        
        report = {
            "timestamp": datetime.now().isoformat(),
            "dry_run": self.dry_run,
            "total_corrections": len(self.corrections),
            "results": [
                {
                    "action": {
                        "type": result.action.action_type,
                        "file_path": result.action.file_path,
                        "line_number": result.action.line_number,
                        "old_tag": result.action.old_tag,
                        "new_tag": result.action.new_tag,
                        "confidence": result.action.confidence,
                        "impact": result.action.impact,
                        "description": result.action.description
                    },
                    "success": result.success,
                    "error_message": result.error_message,
                    "backup_path": result.backup_path
                }
                for result in self.results
            ],
            "summary": {
                "total": len(self.results),
                "successful": sum(1 for r in self.results if r.success),
                "failed": sum(1 for r in self.results if not r.success)
            }
        }
        
        with open(output_path, 'w', encoding='utf-8') as f:
            json.dump(report, f, indent=2, ensure_ascii=False)
        
        print(f"📊 수정 리포트 저장: {output_path}")
    
    def run_correction(self, analysis_file: str = None) -> List[CorrectionResult]:
        """전체 수정 프로세스 실행"""
        print("🚀 TAG 자동 수정 시작...")
        
        # 1. 중복 탐지
        analysis = self.detector.run_full_scan()
        
        # 2. 수정 계획 생성
        self.generate_corrections(self.detector.duplicate_groups)
        
        # 3. 수정 적용
        results = self.apply_corrections()
        
        # 4. 리포트 저장
        self.save_correction_report()
        
        return results

def main():
    """메인 함수"""
    import argparse
    
    parser = argparse.ArgumentParser(description="MoAI-ADK TAG 자동 수정기")
    parser.add_argument("--config", help="설정 파일 경로")
    parser.add_argument("--analysis", help="분석 파일 경로")
    parser.add_argument("--output", help="출력 파일 경로")
    parser.add_argument("--apply", action="store_true", help="실제 수정 적용 (dry-run 아님)")
    
    args = parser.parse_args()
    
    try:
        corrector = TagAutoCorrector(
            config_path=args.config,
            dry_run=not args.apply
        )
        
        results = corrector.run_correction(args.analysis)
        
        # 실패한 수정이 있으면 비제로 리턴
        failed_count = sum(1 for r in results if not r.success)
        if failed_count > 0:
            print(f"\n⚠️  {failed_count}개의 수정 실패")
            return 1
        else:
            print(f"\n✅ 모든 수정 성공")
            return 0
            
    except Exception as e:
        print(f"❌ 오류 발생: {e}")
        return 1

if __name__ == "__main__":
    sys.exit(main())
