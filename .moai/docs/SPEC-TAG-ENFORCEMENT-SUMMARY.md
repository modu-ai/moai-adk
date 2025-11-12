# @DOC:SPEC-TAG-ENFORCEMENT-SUMMARY SPEC 파일 TAG 강제 및 예방 시스템 수행 결과

**작성일**: 2025-11-13
**상태**: 완료 (진행 중)
**목표**: SPEC 파일의 모든 @SPEC: 태그 누락 문제를 체계적으로 해결

---

## 📊 실행 현황

### 문제 규모
- **총 SPEC 디렉토리**: 97개
- **분석 대상 파일**: 194개 (plan.md, acceptance.md)
- **태그 누락 파일**: 122개 (약 63%)

### 즉시 조치 완료 ✅

#### 1. 현재 세션의 SPEC 파일 수정
- ✅ `SPEC-CLI-ANALYSIS-001/plan.md` - @SPEC:CLI-ANALYSIS-001 추가
- ✅ `SPEC-CLI-ANALYSIS-001/acceptance.md` - @SPEC:CLI-ANALYSIS-001 추가

#### 2. spec-builder 에이전트 강화
- ✅ `.claude/agents/alfred/spec-builder.md` - TAG 검증 체크리스트 추가
  - spec.md, plan.md, acceptance.md 모두 @SPEC: 태그 필수 명시
  - 검증 명령어 제공: `rg '@SPEC:{ID}' .moai/specs/SPEC-{ID}/`
  - 누락 시 자동 복구 지침 제공

#### 3. 설정 정책 강화
- ✅ `.moai/config/config.json` - spec_file_requirements 섹션 추가
  ```json
  {
    "spec_file_requirements": {
      "enabled": true,
      "required_files": ["spec.md", "plan.md", "acceptance.md"],
      "required_tag_in_all": "@SPEC:",
      "tag_position": "file_header",
      "enforce_in_template": true,
      "enforce_in_generation": true,
      "auto_fix_enabled": true
    }
  }
  ```

#### 4. 자동 복구 스크립트
- ✅ `.moai/scripts/fix-missing-spec-tags.py` - SPEC 파일 태그 자동 수정 스크립트
  - YAML front matter 지원
  - 제목 라인 자동 감지
  - Dry-run 모드 지원
  - 부분 수정 또는 전체 수정 가능

#### 5. 근본 원인 분석 문서
- ✅ `.moai/docs/TAG-MISSING-ROOT-CAUSE-ANALYSIS.md` - 완벽한 분석 및 예방 계획

---

## 🛡️ 예방 시스템 아키텍처

### Level 1: Template Layer (템플릿 단계)
**상태**: 🟡 부분 적용

- ✅ spec-builder 에이전트에 TAG 검증 체크리스트 추가
- ⏳ 패키지 템플릿 동기화 (로컬 프로젝트만 적용, 패키지는 SPEC 보관 안 함)

### Level 2: Generation Layer (생성 단계)
**상태**: 🟡 부분 적용

- ✅ spec-builder 에이전트에 검증 단계 추가
- ⏳ 자동 복구 로직 통합 (스크립트로 제공됨)

### Level 3: Policy Layer (정책 단계)
**상태**: ✅ 완료

- ✅ config.json에 spec_file_requirements 추가
- ✅ strict 모드 정책과 템플릿 동기화

### Level 4: Validation Layer (검증 단계)
**상태**: 🟡 부분 적용

- ✅ 훅이 이미 TAG 미포함을 감지 및 경고
- ⏳ Pre-Tool 훅으로 생성 단계에서 차단 (향후 구현)

### Level 5: Automated Recovery Layer (자동 복구 단계)
**상태**: ✅ 완료

- ✅ 자동 복구 스크립트 제공
- ✅ Dry-run으로 미리보기 가능
- ✅ 선택적 또는 전체 수정 가능

---

## 📋 실행 계획

### Phase 1: 즉시 조치 (NOW) ✅
- [x] 현재 SPEC 파일들 수정 (SPEC-CLI-ANALYSIS-001)
- [x] spec-builder 강화
- [x] config.json 업데이트
- [x] 근본 원인 분석 문서 작성
- [x] 자동 복구 스크립트 작성

### Phase 2: 단기 조치 (이번 주)
- [ ] 모든 SPEC 파일 자동 수정 실행:
  ```bash
  python3 .moai/scripts/fix-missing-spec-tags.py
  ```
- [ ] 수정 결과 검증:
  ```bash
  find .moai/specs -name "*.md" | xargs grep -c "@SPEC:"
  ```
- [ ] Git 커밋:
  ```bash
  git add .moai/specs/
  git add .moai/scripts/fix-missing-spec-tags.py
  git add .moai/config/config.json
  git add .claude/agents/alfred/spec-builder.md
  git commit -m "feat(spec-tag-enforcement): 모든 SPEC 파일에 @SPEC: 태그 추가 및 예방 시스템 구축"
  ```

### Phase 3: 중기 조치 (v0.23.0)
- [ ] Pre-Tool 훅 구현: `.claude/hooks/alfred/pre_tool__spec_tag_validator.py`
- [ ] spec-builder 자동 복구 로직 통합
- [ ] 테스트 추가:
  - `tests/unit/test_spec_builder_tags.py`
  - `tests/hooks/test_spec_tag_validator.py`

### Phase 4: 장기 조치 (v0.24.0+)
- [ ] GitHub Actions CI/CD에 SPEC TAG 검증 추가
- [ ] 주간 정기 감사 자동화
- [ ] 모니터링 대시보드 구축

---

## 🔍 검증 체크리스트

### 현재 상태 확인

```bash
# 1. 현재 남은 태그 누락 파일 확인
find .moai/specs -name "plan.md" -o -name "acceptance.md" \
  | xargs grep -L "@SPEC:" | wc -l

# 2. 구체적인 누락 파일 목록
find .moai/specs -name "plan.md" -o -name "acceptance.md" \
  | xargs grep -L "@SPEC:"

# 3. 수정 후 검증
find .moai/specs -name "*.md" -exec grep "@SPEC:" {} \; | wc -l
```

### 예상 결과
- 모든 SPEC/plan.md 파일: @SPEC: 태그 포함
- 모든 SPEC/acceptance.md 파일: @SPEC: 태그 포함
- spec.md 파일: 이미 태그 포함

---

## 📚 관련 문서

1. **근본 원인 분석**: `.moai/docs/TAG-MISSING-ROOT-CAUSE-ANALYSIS.md`
   - 문제 원인 체인
   - 4가지 계층별 원인
   - 4단계 예방 시스템

2. **예방 정책**: `.moai/config/config.json` → `tags.policy.spec_file_requirements`
   - 정책 설정
   - 자동 수정 옵션

3. **자동 복구 스크립트**: `.moai/scripts/fix-missing-spec-tags.py`
   - YAML front matter 지원
   - Dry-run 모드
   - 선택적 수정

4. **생성 단계 강화**: `.claude/agents/alfred/spec-builder.md`
   - TAG 검증 체크리스트
   - 검증 명령어
   - 자동 복구 지침

---

## 🎓 핵심 학습

### 왜 이런 문제가 발생했는가?

**1. 비일관적인 템플릿**
- spec.md: @SPEC: 태그 포함
- plan.md, acceptance.md: 태그 누락

**2. 검증 단계 부재**
- 생성 후 검증 로직 없음
- 훅이 경고만 표시 (강제 안 함)

**3. 정책-템플릿 동기화 미흡**
- Config는 "strict" 모드
- 템플릿은 이를 반영 안 함

### 해결 원칙

```
Prevention > Detection > Recovery

1. 정책: 무엇이 필요한가
   ↓
2. 템플릿: 정책을 반영한 템플릿
   ↓
3. 생성: 정책과 템플릿을 준수하는 생성
   ↓
4. 검증: 생성 후 정책 준수 확인
   ↓
5. 복구: 위반 시 자동 복구
```

---

## 🚀 다음 액션 아이템

### 이번 세션
- ✅ 즉시 조치 완료
- ✅ 예방 시스템 구축
- ✅ 문서화 완료

### 다음 세션 (Priority: HIGH)
```bash
# 1. 모든 SPEC 파일 수정
python3 .moai/scripts/fix-missing-spec-tags.py

# 2. 변경 사항 검증
find .moai/specs -name "plan.md" -o -name "acceptance.md" \
  | xargs grep -c "@SPEC:" | grep -v ":1$" | head -5

# 3. Git 커밋
git add .moai/specs/ .moai/scripts/ .moai/config/ .claude/agents/
git commit -m "fix(spec-tags): SPEC 파일 @SPEC: 태그 일괄 추가 및 예방 시스템 구축"
git push origin feature/SPEC-SKILLS-EXPERT-UPGRADE-001
```

---

## 💡 기억할 점

> **"올바른 정책을 설정하고, 그에 맞는 템플릿을 제공하고, 생성 단계에서 검증하고, 만약을 위한 자동 복구를 준비하라"**

이 원칙을 지키면 **다시는 이런 태깅 실수가 발생하지 않을 것입니다.**

---

**문서 ID**: @DOC:SPEC-TAG-ENFORCEMENT-SUMMARY
**생성일**: 2025-11-13
**상태**: 완료 (실행 대기 중)
**다음 검토**: 스크립트 실행 후
