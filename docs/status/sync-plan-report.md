# 문서 동기화 계획 보고서: TAG System v5.0

**작성일**: 2025-10-01
**작성자**: doc-syncer agent
**대상**: MoAI-ADK TAG System v4.0 → v5.0 전환 프로젝트
**버전**: 동기화 계획 v1.0

---

## 📊 상태 분석 결과

### 1. Git 상태 분석

#### 변경된 파일 목록
```
Modified (M):
- moai-adk-ts/package.json (버전 업데이트 예상)
- moai-adk-ts/src/cli/commands/init.ts (코드 변경)

Deleted (D):
- .github/workflows/README.md
- .github/workflows/docs.yml
- .github/workflows/release.yml
- moai-adk-ts/scripts/validate-tags.ts (v4.0 TAG 검증 스크립트 제거됨)

Untracked (??):
- moai-adk-ts/.npmignore (신규 파일)
- moai-adk-ts/src/claude/ (신규 디렉토리)
- moai-adk-ts/templates/.github/ (신규 템플릿)
- moai-adk-ts/tsup.hooks.config.ts (신규 빌드 설정)
```

**최근 커밋**:
```
c4a288e Merge branch 'cleanup/remove-unused-files-phase2' into develop
08e45c3 🧹 refactor: Phase 2 불필요한 파일 대량 제거 (1,845 LOC)
499e687 Merge branch 'cleanup/remove-duplicate-files' into develop
5ec9645 ♻️ refactor: 중복 파일 제거 및 로거 시스템 통합
3ac4b54 🐛 fix: development 환경에서 logs 폴더 자동 생성 비활성화
```

**현재 브랜치**: `develop`
**Main 브랜치**: 미설정 (develop이 사실상 메인)

### 2. TAG 시스템 v5.0 전환 완료 확인

#### ✅ 완료된 작업
1. **핵심 설계 문서 작성**
   - `docs/analysis/tag-system-v5-design.md` (644줄) - v5.0 4-Core 전체 설계
   - `docs/analysis/tag-system-critical-analysis.md` (647줄) - v4.0 비판적 분석

2. **핵심 가이드 업데이트**
   - `CLAUDE.md` - 4-Core TAG 체계 반영
   - `.moai/memory/development-guide.md` - v5.0 업데이트 완료
   - `docs/guide/tag-system.md` (641줄) - 전면 개편 완료

3. **템플릿 시스템 업데이트**
   - `moai-adk-ts/templates/` - 전체 파일 v5.0 반영 완료

4. **CHANGELOG 작성**
   - `CHANGELOG.md` - v0.0.2 릴리스 노트 작성 (v5.0 4-Core 전환 내용)

### 3. TAG 스캔 결과 분석

#### v4.0 8-Core TAG 잔여 현황
**총 발견**: 450+ 건 (추정)

**주요 위치**:
- **예시/문서**: `README.md`, `docs/guide/workflow.md`, `docs/help/faq.md` 등
  - v4.0 예시 코드 (`@FEATURE:AUTH-001`, `@REQ:AUTH-001` 등)
  - 마이그레이션 가이드 참조 (의도적 보존)

- **아카이브**: `.archive/`, `MOAI-ADK-GUIDE.md` (레거시 참조용)
  - 역사적 기록 보존 필요

- **템플릿 예시**: `examples/specs/` (SPEC 예시 문서)
  - 교육용 예시 (v5.0 업데이트 필요)

- **테스트 코드**: `moai-adk-ts/src/__tests__/` (v4.0 TAG 참조)
  - 주석 레벨 참조 (코드는 동작 유지)

#### v5.0 4-Core TAG 현황
**총 발견**: 150+ 건

**주요 위치**:
- **설계 문서**: `docs/analysis/tag-system-v5-design.md`
- **가이드**: `docs/guide/tag-system.md`, `CLAUDE.md`, `development-guide.md`
- **CHANGELOG**: v0.0.2 릴리스 노트

**검증 결과**:
- ✅ v5.0 체계 (`@SPEC:ID`, `@TEST:ID`, `@CODE:ID`, `@DOC:ID`) 정상 사용
- ✅ TAG BLOCK 템플릿 단순화 (`// @CODE:AUTH-001 | SPEC: ... | TEST: ...`)
- ✅ 서브 카테고리 주석화 (`@CODE:ID:API`, `@CODE:ID:DOMAIN` 등)

### 4. 문서-코드 일치성 검사

#### ✅ 일치하는 영역
- **TRUST 원칙**: development-guide.md와 CLAUDE.md 완전 일치
- **TAG 체계**: v5.0 4-Core 정의가 모든 핵심 문서에 일관되게 적용
- **워크플로우**: 3단계 (/moai:1-spec → 2-build → 3-sync) 일관성 유지
- **CODE-FIRST 원칙**: "TAG의 진실은 코드 자체에만 존재" 통일

#### ⚠️ 불일치 영역
1. **README.md**: 여전히 v4.0 8-Core TAG 예시 사용
2. **docs/guide/workflow.md**: v4.0 TAG BLOCK 템플릿 예시
3. **examples/specs/**: v4.0 TAG 구조 예시 파일들
4. **테스트 주석**: 일부 테스트 파일에 v4.0 TAG 참조

---

## 🎯 동기화 전략

### 선택된 모드: **선택적 동기화 (Selective Sync)**

**근거**:
- ✅ 핵심 문서는 이미 v5.0 전환 완료
- ⚠️ 예시/교육 문서는 일부 v4.0 잔여
- ❌ v4.0 전면 제거는 역사적 맥락 손실 위험
- ✅ 마이그레이션 가이드 보존 필요

### 동기화 범위

#### 🔴 우선순위 1: 즉시 동기화 필요 (Critical)

**1. README.md** - 프로젝트 대표 문서
- **문제**: v4.0 TAG 예시 코드 다수 포함
- **조치**: v5.0 예시로 전면 교체
- **예상 시간**: 30분
- **영향도**: 높음 (사용자 첫 접점)

**2. docs/guide/workflow.md** - 핵심 워크플로우 가이드
- **문제**: TAG BLOCK 템플릿이 v4.0 형식
- **조치**: v5.0 템플릿으로 교체, 마이그레이션 섹션 추가
- **예상 시간**: 20분
- **영향도**: 높음 (개발자 주 참조 문서)

**3. examples/specs/** - SPEC 예시 문서
- **문제**: `@REQ:`, `@DESIGN:`, `@TASK:` 사용
- **조치**: v5.0 `@SPEC:`, `@CODE:` 형식으로 변환
- **예상 시간**: 15분
- **영향도**: 중간 (교육용)

#### 🟡 우선순위 2: 선택적 동기화 (Optional)

**4. docs/help/faq.md** - FAQ 문서
- **문제**: v4.0 TAG 검색 예시
- **조치**: v5.0 검색 패턴 추가, v4.0은 "레거시" 섹션으로 이동
- **예상 시간**: 10분
- **영향도**: 중간

**5. docs/reference/cli-cheatsheet.md** - CLI 치트시트
- **문제**: `rg "@REQ:"` 등 v4.0 검색 예시
- **조치**: v5.0 패턴 우선 표시, v4.0은 하위 호환성 참고로 표시
- **예상 시간**: 10분
- **영향도**: 중간

**6. 테스트 코드 주석** - `moai-adk-ts/src/__tests__/`
- **문제**: 주석에 v4.0 TAG 참조
- **조치**: 주석만 업데이트 (코드 동작 유지)
- **예상 시간**: 30분
- **영향도**: 낮음 (내부 코드)

#### 🟢 우선순위 3: 보존 (Archive)

**7. .archive/, MOAI-ADK-GUIDE.md** - 아카이브 문서
- **조치**: 그대로 보존
- **근거**: 역사적 기록 및 마이그레이션 참조용

**8. docs/status/ai-tag-sync-report.md** - 이전 동기화 리포트
- **조치**: 그대로 보존
- **근거**: 과거 작업 이력 참조용

### TAG 체인 검증 계획

#### 검증 방법
```bash
# v5.0 4-Core TAG 전체 스캔
rg '@(SPEC|TEST|CODE|DOC):' -n .moai/specs/ tests/ src/ docs/

# 고아 TAG 탐지
rg '@CODE:AUTH-001' -n src/    # CODE 존재 확인
rg '@SPEC:AUTH-001' -n .moai/specs/  # SPEC 존재 확인

# v4.0 잔여 TAG 스캔 (클린업용)
rg '@(REQ|DESIGN|TASK|FEATURE|API|UI|DATA):' -n
```

#### 검증 체크리스트
- [ ] 모든 `@CODE:ID`에 대응하는 `@SPEC:ID` 존재
- [ ] 모든 `@CODE:ID`에 대응하는 `@TEST:ID` 존재
- [ ] TAG BLOCK 형식 일관성 (`// @CODE:ID | SPEC: ... | TEST: ...`)
- [ ] v4.0 TAG는 예시/마이그레이션 문서에만 존재
- [ ] 실제 코드/SPEC 파일에는 v5.0 TAG만 사용

---

## 🚨 주의사항

### 1. 잠재적 충돌

#### Git 충돌 가능성
- **낮음**: 현재 브랜치 `develop`, 최근 커밋 모두 클린업 작업
- **대비책**: 동기화 전 현재 상태 커밋 권장

#### TAG 마이그레이션 충돌
- **v4.0 → v5.0 혼재 위험**: 일부 문서 업데이트 시 중간 상태 발생
- **대비책**: 우선순위 순서대로 작업하여 단계적 전환

### 2. TAG 체인 무결성

#### 끊어진 체인 가능성
- **원인**: 일부 SPEC만 v5.0 전환 시 연결 끊김
- **대비책**: SPEC-TEST-CODE 세트 단위로 검증

#### 중복 TAG 위험
- **원인**: v4.0과 v5.0 TAG가 동일 ID로 공존
- **대비책**: 동기화 후 중복 스캔 실행

### 3. 성능 영향

#### 예상 동기화 시간
- **우선순위 1**: 약 1시간 5분
- **우선순위 2**: 약 50분
- **TAG 검증**: 약 15분
- **총 예상 시간**: 약 2시간 10분

#### 메모리/디스크 영향
- **낮음**: 문서 파일만 수정, 빌드 불필요
- **백업 권장**: 동기화 전 현재 상태 Git 커밋

---

## ✅ 예상 산출물

### 1. 동기화 완료 문서

#### 업데이트 예정 파일 (6개)
```
/Users/goos/MoAI/MoAI-ADK/README.md
/Users/goos/MoAI/MoAI-ADK/docs/guide/workflow.md
/Users/goos/MoAI/MoAI-ADK/examples/specs/SPEC-002-quality-system.md
/Users/goos/MoAI/MoAI-ADK/examples/specs/SPEC-010-documentation.md
/Users/goos/MoAI/MoAI-ADK/examples/specs/README.md
/Users/goos/MoAI/MoAI-ADK/docs/help/faq.md
```

#### 선택적 업데이트 파일 (2개)
```
/Users/goos/MoAI/MoAI-ADK/docs/reference/cli-cheatsheet.md
/Users/goos/MoAI/MoAI-ADK/moai-adk-ts/src/__tests__/ (주석만)
```

### 2. TAG 체인 검증 리포트

**생성 위치**: `docs/status/sync-report.md`

**포함 내용**:
- v5.0 TAG 전체 스캔 결과
- TAG 체인 완결성 검증
- 고아 TAG 목록 (SPEC 없는 CODE)
- v4.0 잔여 TAG 위치 (마이그레이션 참조용)

### 3. Living Documents 갱신

**자동 갱신 예정**:
- `docs/sections/index.md` - Last Updated 메타 반영
- TAG 추적성 매트릭스 (필요 시 생성)

---

## 📋 실행 계획

### Phase 1: 백업 및 준비 (5분)
1. 현재 Git 상태 커밋
   ```bash
   git add -A
   git commit -m "📸 snapshot: Pre-TAG-v5.0-sync"
   ```

2. 백업 태그 생성
   ```bash
   git tag backup-pre-tag-v5-sync
   ```

### Phase 2: 우선순위 1 동기화 (1시간 5분)
1. README.md 업데이트 (30분)
   - v4.0 TAG BLOCK 예시 → v5.0
   - 검색 명령어 업데이트

2. docs/guide/workflow.md 업데이트 (20분)
   - TAG BLOCK 템플릿 v5.0
   - 마이그레이션 가이드 추가

3. examples/specs/ 업데이트 (15분)
   - 3개 파일 TAG 형식 변환

### Phase 3: TAG 체인 검증 (15분)
1. v5.0 TAG 전체 스캔
2. 고아 TAG 탐지
3. 검증 리포트 생성

### Phase 4: Living Document 갱신 (10분)
1. sync-report.md 생성
2. index.md 메타 업데이트

### Phase 5: Git 정리 (5분)
1. 변경 사항 커밋
   ```bash
   git add -A
   git commit -m "📝 docs: TAG System v5.0 동기화 완료"
   ```

2. 브랜치 정리 (사용자 확인 필요)
   - develop 브랜치 유지
   - PR 생성 여부는 사용자 결정

---

## 🎯 성공 지표

### 동기화 완료 기준
- [x] 핵심 문서 v5.0 전환 완료 (CLAUDE.md, development-guide.md, tag-system.md)
- [ ] README.md v5.0 예시 적용
- [ ] workflow.md v5.0 템플릿 적용
- [ ] examples/specs/ v5.0 형식 변환
- [ ] TAG 체인 검증 100% 통과
- [ ] v4.0 TAG는 레거시 참조용으로만 존재

### 품질 기준
- [ ] 모든 v5.0 TAG가 SPEC-TEST-CODE 체인 완결
- [ ] TAG BLOCK 템플릿 일관성 100%
- [ ] Living Document 자동 갱신 확인
- [ ] v4.0 → v5.0 마이그레이션 가이드 명확

---

## ❓승인 요청

**동기화 실행 계획을 승인하시겠습니까?**

**실행 시 변경 사항**:
- 6개 문서 파일 업데이트 (README.md, workflow.md, examples/specs/ 등)
- TAG 체인 검증 리포트 생성 (`docs/status/sync-report.md`)
- Living Document 메타 갱신

**Git 작업 (사용자 확인 필요)**:
- 현재 상태 커밋 권장
- 백업 태그 생성 권장
- 최종 커밋 메시지: "📝 docs: TAG System v5.0 동기화 완료"

**예상 소요 시간**: 약 1시간 35분 (우선순위 1 + 검증 + 정리)

**롤백 방법**: `git reset --hard backup-pre-tag-v5-sync` (백업 태그 생성 시)

---

**응답 옵션**:
- `y` 또는 `yes`: 전체 계획 승인 및 실행
- `p1`: 우선순위 1만 실행 (README, workflow, examples)
- `verify`: TAG 검증만 먼저 실행
- `n` 또는 `no`: 계획 검토 후 수정

---

**작성**: doc-syncer agent
**문서 버전**: v1.0
**최종 수정**: 2025-10-01
