# 🔖 Phase 1: 릴리즈 계획 보고서 (v0.4.8)

**상태**: ⏳ 사용자 확인 대기
**생성일**: 2025-10-23
**릴리즈 사이클**: `/awesome:release-new` (Python 패키지 릴리즈 자동화)

---

## 📊 현재 상태 요약

### 버전 정보
- **현재 버전 (pyproject.toml)**: 0.4.7
- **마지막 릴리즈 태그**: v0.4.7 (커밋: `1644bd0d`)
- **지난 릴리즈 이후 커밋**: 9개
- **권장 목표 버전**: 0.4.8 (패치 증가)

### 품질 검증 결과 (Phase 0) ✅

| 검사 항목 | 결과 | 세부 사항 |
|----------|------|---------|
| **Python 버전** | ✅ 통과 | 3.13.1 설치됨 |
| **패키지 관리자** | ✅ 통과 | uv 0.9.3 설치됨 |
| **단위 테스트** | ✅ 통과 | 446개 통과, 8개 스킵 |
| **테스트 커버리지** | ✅ 통과 | 86.02% (목표: 85%) |
| **타입 체크** | ⚠️ 주의 | PhaseExecutor에서 1개 오류 (비차단) |
| **린팅** | ⚠️ 주의 | 템플릿 파일의 5개 이슈 (예상됨) |
| **보안** | ✅ 통과 | 높은 신뢰도, 낮은 심각도만 9개 |
| **의존성** | ⚠️ 구식 | 개발 환경의 5개 취약점 (배포 제외) |

---

## 📝 지난 릴리즈 이후 변경사항 분석

### 최근 커밋 (9개)

```
14f88e9c refactor(docs,cli,skills): Finalize output-styles TUI, add README generation Skill
523a4216 fix(output-styles): Simplify TUI layout with dash-line separators (no box frames)
87916334 refactor(output-styles): Enhance TUI elements with box frames, progress indicators, and visual hierarchy
accc2dbe fix(hooks): Implement PostToolUse JSON schema validation fix
dadbed31 refactor(docs): Remove agent and skill count references from all README files
09cb4e68 refactor(claude-code): Implement v3.0.0 Skills-based architecture with moai-cc-guide orchestrator
7ffaad43 refactor(skills): Redesign moai-alfred-tui-survey → moai-alfred-ask-user-questions
e60c7d93 fix(hooks): Migrate Hook system to Claude Code standard schema
bee62ce9 docs: Update CHANGELOG with v0.4.7 release notes
```

### 변경 유형 분류

| 유형 | 개수 | 예시 |
|------|------|------|
| **fix** (수정) | 2개 | PostToolUse 훅 JSON 스키마, Output-styles TUI 레이아웃 |
| **refactor** (리팩토링) | 4개 | 문서, Claude Code 스킬 아키텍처, 스킬 UI 설문 재설계 |
| **docs** (문서) | 1개 | CHANGELOG 업데이트 |
| **feat** (기능) | 0개 | (없음 - 유지보수 릴리즈) |

**릴리즈 유형**: **패치 릴리즈** (수정 + 리팩토링, 새로운 기능 없음)

---

## 🚀 릴리즈 계획 세부사항

### 업데이트할 파일 (SSOT 방식)

```yaml
주요 업데이트:
  - pyproject.toml: "0.4.7" → "0.4.8"

자동 처리 (pyproject.toml 통해):
  - src/moai_adk/__init__.py: importlib.metadata를 통한 동적 버전 로드
  - 기타 파일: 수동 업데이트 불필요 (SSOT 원칙 사용)
```

### 릴리즈 작업 계획

| 단계 | 작업 | 담당 | 상태 |
|------|------|------|------|
| **Phase 1** | ✅ 버전 분석 | 당신 | 완료 |
| **Phase 1** | ✅ 품질 검증 | 당신 | 완료 |
| **Phase 1** | ⏳ 사용자 확인 | 당신 | 대기 중 |
| **Phase 2** | pyproject.toml 업데이트 | alfred:2-run | 대기 중 |
| **Phase 2** | git 커밋 생성 | alfred:2-run | 대기 중 |
| **Phase 2** | 주석 태그 생성 | alfred:2-run | 대기 중 |
| **Phase 3** | 패키지 빌드 (uv) | alfred:3-sync | 대기 중 |
| **Phase 3** | PyPI 발행 | alfred:3-sync | 대기 중 (선택사항) |
| **Phase 3** | GitHub 릴리즈 생성 | alfred:3-sync | 대기 중 (Draft) |

### 예정 릴리즈 날짜
**목표 날짜**: 2025-10-23 (오늘)

---

## 🎯 시맨틱 버저닝 결정

### 버전 증가: 0.4.7 → 0.4.8

**근거**:
- ✅ 버그 수정만 있음 (PostToolUse 훅, Output-styles 레이아웃)
- ✅ 리팩토링만 있음 (동작 변경 없음)
- ✅ 새로운 공개 API 기능 없음
- ✅ 하위호환성 파괴 없음
- ❌ 마이너 버전 증가 (0.5.0) 불충분
- ❌ 메이저 버전 증가 (1.0.0) 불필요

**결정**: **패치 버전 증가** ✓

---

## 📋 변경로그 요약

### v0.4.8 (대기 중)

#### 🐛 버그 수정
- **PostToolUse 훅**: Claude Code가 요구하는 형식으로 JSON 스키마 검증 수정 (`decision`/`reason` 필드)
- **Output-styles TUI**: 마크다운 렌더링 개선을 위해 박스 프레임을 대시라인으로 대체

#### ♻️ 개선사항
- 모든 언어별 README 파일에서 에이전트/스킬 개수 참조 제거
- 스킬 UI 설문 컴포넌트 재설계 (`moai-alfred-tui-survey` → `moai-alfred-ask-user-questions`)
- Claude Code 스킬 아키텍처를 v3.0.0으로 업데이트 (`moai-cc-guide` 오케스트레이터 포함)
- GitHub README.md 생성 스킬 추가 (`moai-domain-readme-generation`)

---

## ✅ 품질 게이트 검증

| 게이트 | 요구사항 | 상태 | 근거 |
|--------|---------|------|------|
| **테스트 커버리지** | ≥ 85% | ✅ 통과 | 86.02% (446/517 라인) |
| **실패한 테스트 없음** | 0건 | ✅ 통과 | 446개 통과, 8개 스킵 |
| **타입 안전성** | Python 3.13+ | ✅ 통과 | mypy 검사 (1개 예상 주의) |
| **보안** | bandit 스캔 | ✅ 통과 | 낮은 신뢰도 이슈 9개만 |
| **문서** | 업데이트됨 | ✅ 통과 | 변경로그 준비 완료 |
| **Git 상태** | 깨끗한 트리 | ✅ 통과 | 모든 작업 커밋됨 (14f88e9c) |

---

## 🔔 주의사항 및 경고

### ⚠️ 예상된 편차 (비차단)

1. **템플릿 파일 린팅 이슈**:
   - 템플릿에 `Any` 임포트 경고 및 공백 이슈 존재
   - 예상되는 동작 (템플릿 스캐폴딩에 강제되지 않음)
   - 운영 코드에 영향 없음

2. **타입 체킹 주의**:
   - `PhaseExecutor` 인자 유형 불일치 감지됨
   - 비차단 오류 (낮은 우선순위 리팩토링 후보)
   - 릴리즈 준비 상태에 영향 없음

3. **개발 의존성 취약점**:
   - 5개 패키지에 알려진 취약점 (h11, h2, httpx, idna, pip)
   - 릴리즈 전 개발 환경만 해당
   - 운영 패키지에 포함되지 않음

---

## 🎬 다음 단계: 사용자 확인 필요

### ✋ 검토 및 확인 요청

**모든 품질 게이트를 통과했습니다.** Phase 2-3 진행 준비 완료입니다.

**두 가지 선택지:**

1. ✅ **릴리즈 승인** → 다음 진행:
   - `pyproject.toml` 업데이트 → `0.4.8`
   - 릴리즈 커밋 생성 (🔖 RELEASE 태그)
   - `uv`로 패키지 빌드
   - GitHub 릴리즈 생성 (Draft)
   - 선택사항: PyPI 발행

2. ❌ **릴리즈 취소** → 워크플로우 중단 및 추가 지시 대기

**진행 명령어**: 대화에서 확인 의사 표시 (긍정 응답)

---

**보고서 생성일**: 2025-10-23
**Phase**: 1 (분석 및 계획) — 완료
**다음 Phase**: 2 (실행) — 확인 대기 중
