# MoAI-ADK GitFlow PR

> 🗿 **GitFlow 완전 투명성** - 에이전트가 자동으로 정보를 채웁니다

## 📝 SPEC 정보

- **관련 SPEC**: `SPEC-AUTH-001` (예: JWT 인증 시스템)
- **디렉토리**: `.moai/specs/SPEC-AUTH-001/`
- **@TAG 연결**: @SPEC:AUTH-001 @CODE:AUTH-001 (자동 태깅)

## ✅ SPEC 품질 체크

- [ ] **YAML Front Matter**: 7개 필수 필드 (id, version, status, created, updated, author, priority)
- [ ] **HISTORY 섹션**: 버전별 변경 이력 기록 (v0.0.1 INITIAL 항목 포함)
- [ ] **EARS 요구사항**: Ubiquitous, Event-driven, State-driven, Optional, Constraints
- [ ] **@SPEC:ID TAG**: 문서에 TAG 포함 및 중복 확인 (`rg "@SPEC:<ID>" -n`)

## 🤖 자동 검증 상태

<!-- 아래 체크리스트는 에이전트가 자동으로 업데이트합니다 -->
<!-- /alfred:1-spec → feature 브랜치 생성 → Draft PR -->
<!-- /alfred:2-build → TDD 구현 → 체크박스 자동 체크 -->
<!-- /alfred:3-sync → 문서 동기화 → Ready for Review -->

- [ ] **spec-builder**: EARS 명세 완성 및 feature 브랜치 생성
- [ ] **code-builder**: TDD RED-GREEN-REFACTOR 완료
- [ ] **doc-syncer**: Living Document 동기화 및 PR Ready

## 📊 품질 지표 (자동 계산)

- **TRUST 5원칙**: ✅ 준수
- **테스트 커버리지**: XX% (85% 이상 목표)
- **@TAG 추적성**: 100%

## 🌍 Locale 설정

- **프로젝트 언어**: <!-- ko/en/ja/zh -->
- **커밋 메시지**: <!-- locale에 따라 자동 생성 -->

## 🎯 변경 사항

<!-- code-builder가 TDD 결과를 자동으로 채움 -->

### 🔴 RED (테스트 작성)
- **테스트 파일**: `tests/auth/service.test.ts`
- **테스트 설명**: [실패하는 테스트 설명]

### 🟢 GREEN (구현)
- **구현 파일**: `src/auth/service.ts`
- **구현 완료**: [기능 설명]

### ♻️ REFACTOR (개선)
- **리팩토링 내역**: [코드 품질 개선 사항]

## 📚 문서 동기화

<!-- doc-syncer가 자동으로 채움 -->

- [ ] README 업데이트
- [ ] API 문서 동기화
- [ ] TAG 인덱스 업데이트
- [ ] HISTORY 섹션 업데이트 (SPEC 문서)

---

🚀 **MoAI-ADK**: 3단계 파이프라인으로 Git 명령어 없이도 프로페셔널 개발!

**리뷰어**: TRUST 5원칙 준수 여부와 SPEC 메타데이터 완성도만 확인하면 됩니다.
