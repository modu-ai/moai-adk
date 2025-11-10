# 통합 문서 관리 가이드

MoAI-ADK의 문서 검증, 린팅, 품질 보증 시스템을 완벽히 이해하세요.

## 개요

**moai-docs-unified** 는 한국어 문서 품질을 보장하는 5단계 검증 파이프라인입니다:

```
Phase 1: 마크다운 린팅
    ↓
Phase 2: Mermaid 다이어그램 검증
    ↓
Phase 2.5: Mermaid 코드 추출
    ↓
Phase 3: 한글 타이포그래피 검증
    ↓
Phase 4: 종합 품질 리포트
```

## 빠른 시작

### 전체 검증 실행

프로젝트 루트에서:

```bash
# 모든 Phase 실행
uv run .claude/skills/moai-docs-unified/scripts/lint_korean_docs.py
uv run .claude/skills/moai-docs-unified/scripts/validate_mermaid_diagrams.py
uv run .claude/skills/moai-docs-unified/scripts/validate_korean_typography.py
uv run .claude/skills/moai-docs-unified/scripts/generate_final_comprehensive_report.py
```

### 결과 확인

생성된 리포트 파일들:

```
.moai/reports/
├── lint_report_ko.txt              # Phase 1 결과
├── mermaid_validation_report.txt   # Phase 2 결과
├── mermaid_detail_report.txt       # Phase 2.5 결과
├── korean_typography_report.txt    # Phase 3 결과
└── korean_docs_comprehensive_review.txt  # Phase 4 최종 리포트
```

---

## 상세 설명

### Phase 1: 마크다운 린팅

**목적**: 마크다운 구조 및 형식 검증

**검증 항목**:

| 항목 | 설명 | 예시 |
|------|------|------|
| **제목(Header)** | H1 유일성, 계층 구조 | `# Title` (1개만) → `## Section` → `### Subsection` |
| **코드 블록** | 언어 선언, 일치하는 구분자 | `` ```python ... ``` `` |
| **링크** | 상대 경로, 파일 존재 여부, https | `[텍스트](../path/to/file.md)` |
| **리스트** | 마커 일관성, 들여쓰기 | `- Item 1` → `  - Nested` |
| **테이블** | 열 개수 일치, 정렬 | `\| Column 1 \| Column 2 \|` |
| **공백** | 후행 공백, UTF-8 인코딩 | 라인 끝에 공백 제거 |

**실행**:

```bash
uv run .claude/skills/moai-docs-unified/scripts/lint_korean_docs.py \
  --path docs/src/ko \
  --output .moai/reports/lint_report_ko.txt
```

**출력 예시**:

```
검사 완료: 53개 파일
  - 코드블록: 정상
  - 링크: 351개 (거짓양성: 상대경로)
  - 리스트: 241개 항목 검증
  - 헤더: 1,241개 거짓양성 (HTML 스팬)
```

---

### Phase 2: Mermaid 다이어그램 검증

**목적**: Mermaid 다이어그램 타입 및 문법 검증

**지원 다이어그램 타입**:

```
✅ graph TD/BT/LR/RL     (플로우차트)
✅ stateDiagram-v2       (상태 머신)
✅ sequenceDiagram       (시퀀스)
✅ classDiagram          (클래스)
✅ erDiagram             (ER 다이어그램)
✅ gantt                 (간트)
```

**실행**:

```bash
uv run .claude/skills/moai-docs-unified/scripts/validate_mermaid_diagrams.py \
  --path docs/src \
  --output .moai/reports/mermaid_validation_report.txt
```

**결과 해석**:

```
📊 Diagram Type: graph TD
   ✅ Valid: 유효한 타입
   ✅ Syntax: 문법 정상
   📍 Line: 125 (파일 내 위치)
   📏 Height: 15 lines
```

---

### Phase 2.5: Mermaid 코드 추출

**목적**: 모든 Mermaid 코드 추출 및 렌더링 테스트 가이드 제공

**렌더링 테스트 가이드**:

```bash
uv run .claude/skills/moai-docs-unified/scripts/extract_mermaid_details.py
```

생성된 파일에서:
1. `코드:` 섹션 전체 복사
2. https://mermaid.live 접속
3. 좌측 편집기에 붙여넣기
4. 우측에서 렌더링 확인

---

### Phase 3: 한글 타이포그래피 검증

**목적**: 한글 문서 특화 검증

**검증 항목**:

| 항목 | 권장 사항 | 예시 |
|------|----------|------|
| **인코딩** | UTF-8 (필수) | `한글 문서` |
| **전각 공백** | 반각 사용 | `` `` (O) vs `　` (X) |
| **전각 문자** | 반각 사용 | `()` (O) vs `（）` (X) |
| **마침표** | `.` 사용 | `.` (O) vs `。` (X) |
| **쉼표** | `,` 사용 | `,` (O) vs `、` (X) |
| **한영 공백** | 공백 추가 | `한글 English` (O) vs `한글English` (X) |

**실행**:

```bash
uv run .claude/skills/moai-docs-unified/scripts/validate_korean_typography.py \
  --path docs/src \
  --output .moai/reports/korean_typography_report.txt
```

**결과 예시**:

```
✅ UTF-8 인코딩: 100% 정상
✅ 전각 문자 사용 최소화: 권장됨
⚠️  전각 공백 발견: 12개 (수정 필요)
```

---

### Phase 4: 종합 품질 리포트

**목적**: 모든 Phase 결과 통합 및 우선순위 지정

**보고서 구성**:

1. **Executive Summary** - 종합 품질 점수 (0-10)
2. **Priority 1 (긴급)** - 즉시 수정 필요
3. **Priority 2 (높음)** - 중요 개선 사항
4. **Priority 3 (낮음)** - 선택 사항
5. **Action Items** - Immediate/Short-term/Long-term

**실행**:

```bash
uv run .claude/skills/moai-docs-unified/scripts/generate_final_comprehensive_report.py \
  --report-dir .moai/reports \
  --output .moai/reports/korean_docs_comprehensive_review.txt
```

**품질 점수 해석**:

```
📊 Overall Quality Score: 8.5/10 ⭐⭐⭐⭐

점수 범위:
  9.0-10.0  - 탁월함 (Excellent)
  8.0-8.9   - 우수함 (Good) ← 현재
  7.0-7.9   - 양호함 (Fair)
  6.0-6.9   - 개선 필요 (Needs Work)
  < 6.0     - 긴급 수정 (Critical)
```

---

## 고급 사용법

### 단일 파일 린팅

```bash
# 특정 파일만 검사
uv run .claude/skills/moai-docs-unified/scripts/lint_korean_docs.py \
  --path docs/src/ko/guides/specific-guide.md
```

### 커스텀 리포트 경로

```bash
# 다른 위치에 리포트 저장
uv run .claude/skills/moai-docs-unified/scripts/validate_korean_typography.py \
  --output my-custom-report.txt
```

### CI/CD 파이프라인 통합

GitHub Actions를 사용한 자동 검증:

```yaml
# .github/workflows/docs-validation.yml
name: Documentation Validation

on:
  pull_request:
    paths:
      - 'docs/**'

jobs:
  validate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-python@v4
      - run: pip install uv
      - run: |
          uv run .claude/skills/moai-docs-unified/scripts/lint_korean_docs.py
          uv run .claude/skills/moai-docs-unified/scripts/validate_korean_typography.py
          uv run .claude/skills/moai-docs-unified/scripts/generate_final_comprehensive_report.py
```

---

## 트러블슈팅

### "Project root not found" 오류

**원인**: 스크립트가 MoAI-ADK 프로젝트 루트에서 멀리 떨어진 곳에서 실행됨

**해결책**:
```bash
# 프로젝트 루트로 이동
cd /Users/goos/MoAI/MoAI-ADK

# 다시 실행
uv run .claude/skills/moai-docs-unified/scripts/lint_korean_docs.py
```

### "uv: command not found" 오류

**원인**: uv가 설치되지 않음

**해결책**:
```bash
pip install uv
```

### 리포트 파일이 생성되지 않음

**원인**: `.moai/reports/` 디렉토리 없음

**확인**:
```bash
mkdir -p .moai/reports
ls -la .moai/reports/
```

### 거짓양성 오류가 많음

**원인**:
- Phase 1: HTML 스팬 (Material Icons)이 헤더로 인식됨
- Phase 1: 상대 경로 링크가 깨진 링크로 인식됨

**해결책**:
- 오류 목록을 검토하여 실제 문제만 수정
- 거짓양성은 무시 가능

---

## 품질 메트릭

### 현재 상태

| 메트릭 | 값 | 상태 |
|--------|-----|------|
| 종합 품질 점수 | 8.5/10 | ✅ 우수 |
| UTF-8 인코딩 | 100% | ✅ 완벽 |
| Mermaid 유효성 | 100% (16/16) | ✅ 완벽 |
| 검증된 라인 | 28,543 | ✅ 광범위 |
| 검증된 파일 | 53개 | ✅ 전수 |

### 목표

| 메트릭 | 목표 | 주기 |
|--------|------|------|
| 품질 점수 | ≥ 8.0 | 매주 |
| 새 오류 발생률 | < 5% | PR 단위 |
| Mermaid 유효성 | 100% | 매 커밋 |
| UTF-8 준수 | 100% | 매 커밋 |

---

## 에이전트 활용

### docs-manager 에이전트 호출

task를 통해 문서 검증 자동화:

```python
# Alfred sub-agent에서
Task(
    description="문서 품질 검증",
    prompt="""
    전체 문서 검증 실행:
    1. Phase 1-3 실행
    2. 품질 점수 추출
    3. 우선순위 3개 항목 보고
    """,
    subagent_type="docs-manager"
)
```

---

## 참고 자료

- **[Skill 상세](../../reference/skills/index.md)** - moai-docs-unified 스킬 설명
- **[Agent 상세](../../reference/agents/index.md)** - docs-manager 에이전트 역할
- **[마크다운 가이드](https://www.markdownguide.org/)** - 마크다운 기본 문법
- **[Mermaid 공식](https://mermaid.live)** - Mermaid 다이어그램 렌더링

---

## 다음 단계

다음 방법 중 선택:

- **[스크립트 상세 가이드](scripts.md)** - 각 스크립트 심화 사용법
- **[에이전트 가이드](agent.md)** - docs-manager 활용법
- **[자주 묻는 질문](faq.md)** - 일반적인 문제 해결

---

**생성 일시**: 2025-11-10
**품질 점수**: 8.5/10
**최종 검증**: Phase 4 (종합 리포트)
