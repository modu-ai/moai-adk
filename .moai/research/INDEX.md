# Figma 디자인 분석 프로젝트 - 인덱스

**생성 일시**: 2025-11-19
**프로젝트**: 모듈의사주-프로젝트-와디즈
**파일 키**: m2odCIWVPWv84ygT5w43Ur
**노드 ID**: 689:1242

---

## 🎯 빠른 네비게이션

### 📖 문서 (최우선)

| 문서 | 대상 사용자 | 읽는 시간 | 링크 |
|------|-----------|---------|------|
| **README.md** | 모두 (필독) | 5분 | 프로젝트 개요, 빠른 시작 |
| **FIGMA_EXTRACTION_GUIDE.md** | 처음 사용자 | 15분 | 7가지 추출 방법, 상세 설명 |
| **FIGMA_WORKFLOW.md** | 구현자 | 20분 | 5-Phase 워크플로우, 실행 예제 |

### 🔧 도구 (자동 실행)

| 도구 | 용도 | 실행 명령 |
|-----|------|---------|
| **figma_analyzer.py** | 색상/타이포 분석 | `uv run .moai/research/figma_analyzer.py --json figma-metadata.json --css --report` |

### 📊 참고 자료 (추가 정보)

| 파일 | 내용 |
|-----|------|
| **figma-mcp-official-docs.md** | Figma MCP 공식 문서 (40KB) |
| **figma-mcp-params.md** | API 파라미터 상세 설명 |
| **figma-mcp-error-mapping.md** | 에러 메시지 및 해결책 |
| **figma-mcp-compatibility-matrix.md** | MCP 호환성 매트릭스 |
| **figma-mcp-research-summary.md** | 연구 결과 요약 |

---

## 🚀 3단계 빠른 시작

### Step 1: 메타데이터 추출 (5분)

```bash
# Figma에서 Copy as JSON 또는 API로 추출
cat > .moai/research/figma-metadata.json << 'EOF'
# Figma JSON 메타데이터 붙여넣기
EOF
```

**또는** (API 사용):
```bash
export FIGMA_TOKEN=your_token
curl -H "X-Figma-Token: $FIGMA_TOKEN" \
  "https://api.figma.com/v1/files/m2odCIWVPWv84ygT5w43Ur/nodes?ids=689:1242" \
  > .moai/research/figma-metadata.json
```

### Step 2: 분석 실행 (2분)

```bash
uv run .moai/research/figma_analyzer.py \
  --json .moai/research/figma-metadata.json \
  --analyze colors,typography,components,images \
  --css \
  --report
```

### Step 3: 결과 확인 (3분)

```bash
# 생성된 파일 확인
ls -lh .moai/research/design-tokens.css
ls -lh .moai/research/analysis-report.md
cat .moai/research/design-tokens.css | head -20
```

---

## 📁 생성된 파일 구조

```
.moai/research/
│
├── 📚 시작 (먼저 읽기)
│   ├── INDEX.md                           ← 현재 파일
│   ├── README.md                          ← 프로젝트 개요
│   ├── FIGMA_EXTRACTION_GUIDE.md          ← 상세 추출 방법
│   └── FIGMA_WORKFLOW.md                  ← 5-Phase 워크플로우
│
├── 🔧 도구 및 스크립트
│   ├── figma_analyzer.py                  ← Python 분석 스크립트
│   └── scripts/
│       ├── fetch-figma-metadata.sh        ← API 메타데이터 다운로드
│       ├── download-figma-images.sh       ← 이미지 다운로드
│       └── check-contrast.py              ← WCAG 대조 검증
│
├── 📖 참고 자료 (추가 정보)
│   ├── figma-mcp-official-docs.md         ← Figma MCP 공식 문서
│   ├── figma-mcp-params.md                ← API 파라미터
│   ├── figma-mcp-error-mapping.md         ← 에러 해결
│   ├── figma-mcp-compatibility-matrix.md  ← 호환성
│   └── figma-mcp-research-summary.md      ← 연구 요약
│
├── 📊 분석 결과 (추출 후 생성)
│   ├── figma-metadata.json                ← Figma JSON 메타데이터
│   ├── figma-images.json                  ← 이미지 URL 목록
│   ├── design-tokens.css                  ← CSS 변수
│   ├── analysis-report.md                 ← 분석 리포트
│   ├── analysis-metadata.json             ← 추출 메타데이터
│   ├── color-palette.json                 ← 색상 팔레트
│   └── figma-node-689-1242-analysis.md   ← 최종 분석
│
└── 🎨 자산 (이미지 다운로드 후)
    └── figma-assets/
        ├── node-689-1242.png              ← 1배 해상도
        ├── node-689-1242@2x.png           ← 2배 해상도
        └── images/                        ← 삽입된 이미지
```

---

## 🎓 학습 경로

### 초급 (30분)
1. `README.md` 읽기 (5분)
2. 빠른 시작 3단계 실행 (25분)
   - 메타데이터 추출
   - 분석 스크립트 실행
   - 결과 확인

### 중급 (1-2시간)
1. `FIGMA_WORKFLOW.md` 읽기 (20분)
2. 각 Phase별 상세 실행 (1.5시간)
   - Phase 1: 메타데이터 수집
   - Phase 2: 색상/타이포 분석
   - Phase 3: 이미지 다운로드
   - Phase 4: 접근성 검증
   - Phase 5: 문서화

### 고급 (2-4시간)
1. `FIGMA_EXTRACTION_GUIDE.md` 읽기 (15분)
2. 자동화 스크립트 커스터마이징 (30분)
3. React 컴포넌트 생성 시작 (1.5-3시간)
   ```bash
   /moai:1-plan "노드 689:1242 React 컴포넌트 생성"
   /clear
   /moai:2-run SPEC-XXX
   ```

---

## 💡 주요 기능

### 1. 색상 자동 추출
```css
/* 자동 생성되는 CSS 변수 */
:root {
  --color-primary-500: #0EA5E9;
  --color-text: #0F172A;
  --color-background: #F8FAFC;
  /* 20+ 색상 */
}
```

### 2. 타이포그래피 분석
- 폰트 패밀리
- 크기 및 가중치
- 라인 높이
- 자간

### 3. 컴포넌트 추출
- Frame/Component 식별
- 자식 노드 구조
- 크기 및 위치
- 속성 (스타일, 효과)

### 4. 접근성 검증
```python
색상 대조 비율: 15.3:1
✅ WCAG AA: PASS (4.5:1 요구)
✅ WCAG AAA: PASS (7.0:1 요구)
```

---

## 🔗 다음 단계

### 즉시 (오늘)
- [ ] README.md 읽기
- [ ] 메타데이터 추출
- [ ] 분석 스크립트 실행
- [ ] 결과 검토

### 내일
- [ ] FIGMA_WORKFLOW.md 읽기
- [ ] 5-Phase 상세 실행
- [ ] 이미지 자산 다운로드
- [ ] 접근성 검증 완료

### 다음주
- [ ] React 컴포넌트 생성 시작
  ```bash
  /moai:1-plan "노드 689:1242 React 컴포넌트 (TypeScript, Tailwind)"
  /clear
  /moai:2-run SPEC-001
  ```
- [ ] Storybook 설정
- [ ] Playwright 테스트 작성

---

## 🆘 FAQ

### Q: Figma 파일에 접근할 수 없습니다.
**A**: 파일이 공개되지 않았을 수 있습니다. Figma에서 공유 설정 확인:
1. 파일 열기
2. 공유 버튼 (우측 상단)
3. "Anyone with the link can view" 선택

### Q: 토큰이 없어도 분석할 수 있나요?
**A**: 네! UI에서 "Copy as JSON" 사용:
1. Figma 파일에서 노드 선택
2. 우클릭 → Copy as JSON
3. `.moai/research/figma-metadata.json`에 저장
4. 분석 스크립트 실행

### Q: CSS 대신 Tailwind config로 생성할 수 있나요?
**A**: 스크립트 수정 필요. PR 환영합니다!
```python
# figma_analyzer.py 수정:
# generate_tailwind_config() 메서드 추가
```

### Q: 생성된 React 컴포넌트는 어디에?
**A**: 분석 후 `/moai:1-plan`으로 SPEC 생성:
```bash
/moai:1-plan "노드 689:1242 기반 React 컴포넌트"
/clear
/moai:2-run SPEC-001
```

---

## 📞 지원

- **문제 해결**: `FIGMA_EXTRACTION_GUIDE.md` 끝의 "문제 해결" 섹션
- **API 오류**: `figma-mcp-error-mapping.md` 참고
- **기술 지원**: `figma-mcp-official-docs.md` 참고

---

## 🔄 워크플로우 요약

```
메타데이터 추출 (5분)
    ↓
분석 스크립트 (2분)
    ↓
CSS/JSON 생성 (자동)
    ↓
이미지 다운로드 (5-10분)
    ↓
접근성 검증 (5분)
    ↓
React 컴포넌트 생성 (1-2시간)
    ↓
Storybook 통합 (30분)
    ↓
Playwright 테스트 (1시간)
    ↓
완료 ✅
```

**총 소요 시간**: 3-5시간 (자동화 포함)

---

**Version**: 1.0.0
**Last Updated**: 2025-11-19
**Status**: 준비 완료 ✅

시작하려면 **README.md**를 먼저 읽으세요!
