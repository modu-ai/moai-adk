# TAG 검증 및 Skill 검증 도구

> **용도**: MoAI-ADK TAG 시스템 및 Skill 메타데이터 검증
>
> **대상**: 최종 사용자 및 개발자

---

## 📋 스크립트 목록

### 1. `tag_dedup_manager.py` (통합)

**목적**: TAG 중복 탐지 및 자동 수정 (통합 도구)

**통합 배경**:
- `tag_dedup_detector.py` (중복 탐지) + `tag_auto_corrector.py` (자동 수정)을 하나로 병합
- 중복 코드 제거, 단일 CLI로 통일
- 사용자 경험 개선

**사용 대상**:
- TAG 시스템 정기 유지보수
- SPEC 파일 추가 후 중복 확인
- TAG 정합성 검증

**기능**:
- TAG 중복 스캔 (변경 없음)
- 수정 계획 검토 (dry-run)
- 자동 수정 적용
- 전체 워크플로우 (스캔 → 검토 → 적용)

**사용 방법**:

```bash
# 1️⃣ 스캔 전용 (중복 탐지, 변경 없음)
python3 .moai/scripts/validation/tag_dedup_manager.py --scan-only

# 2️⃣ 검토 모드 (수정 계획 보기)
python3 .moai/scripts/validation/tag_dedup_manager.py --dry-run

# 3️⃣ 실행 모드 (중복 수정 적용)
python3 .moai/scripts/validation/tag_dedup_manager.py --apply

# 4️⃣ 전체 워크플로우 (스캔 → 검토 → 적용)
python3 .moai/scripts/validation/tag_dedup_manager.py --full

# 커스텀 설정 사용
python3 .moai/scripts/validation/tag_dedup_manager.py --config .moai/my-config.json --apply
```

**정상 실행 예시**:
```
🔍 TAG 중복 스캔 중...

⚠️  중복 그룹 발견: 3
   총 중복 TAG 수: 7

📋 TAG 중복 검토 중 (변경 없음)...

📊 수정 계획:
   적용될 수정: 7
   - @CODE:SPEC-GENERATOR-001 → @CODE:SPEC-GENERATOR-002 (.moai/specs/...)
   - ...
```

---

### 2. `validate_all_skills.py`

**목적**: 모든 Skills의 메타데이터와 구조 검증

**사용 대상**:
- Skill 추가/수정 후 검증
- 패키지 배포 전 Skill 표준 준수 확인
- Skill 문서화 완성도 점검

**기능**:
- SKILL.md 메타데이터 검증 (name, version, status)
- 필수 섹션 존재 확인
- 파일 구조 일관성 검사
- @TAG 연계 검증

**사용 방법**:
```bash
# 모든 Skills 검증
python3 .moai/scripts/validation/validate_all_skills.py

# 상세 리포트 생성
python3 .moai/scripts/validation/validate_all_skills.py --detailed

# 특정 Skill만 검증
python3 .moai/scripts/validation/validate_all_skills.py --skill moai-lang-python
```

---

## 🚀 일반적인 워크플로우

### 새 SPEC 추가 후
```bash
# 1. TAG 중복 확인
python3 .moai/scripts/validation/tag_dedup_manager.py --scan-only

# 2. 중복이 있으면 수정
python3 .moai/scripts/validation/tag_dedup_manager.py --dry-run
python3 .moai/scripts/validation/tag_dedup_manager.py --apply
```

### 패키지 배포 전
```bash
# 1. Skill 검증
python3 .moai/scripts/validation/validate_all_skills.py

# 2. TAG 검증
python3 .moai/scripts/validation/tag_dedup_manager.py --scan-only
```

---

## 📊 성능 및 실행 시간

| 도구 | 실행 시간 | 범위 |
|------|---------|------|
| tag_dedup_manager (--scan-only) | ~10초 | 전체 코드베이스 |
| tag_dedup_manager (--full) | ~30초 | 포함: 스캔 + 검토 + 적용 |
| validate_all_skills | ~5초 | 모든 Skills |

---

## 📚 관련 문서

- **TAG 시스템**: `.moai/specs/TAG-REFERENCE.md`
- **Skill 시스템**: `.moai/skills/`
- **개발 가이드**: `CONTRIBUTING.md`

---

**마지막 업데이트**: 2025-11-13
**상태**: Production Ready
