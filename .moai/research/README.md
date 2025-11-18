# Figma 설계 분석 프로젝트

**프로젝트명**: 모듈의사주-프로젝트-와디즈
**파일 키**: m2odCIWVPWv84ygT5w43Ur
**노드 ID**: 689:1242
**상태**: 준비 완료
**최종 업데이트**: 2025-11-19

---

## 📋 개요

이 프로젝트는 Figma 디자인 파일 (`노드 689:1242`)에서 메타데이터를 추출하고 분석하여 React 컴포넌트를 생성하는 end-to-end 워크플로우입니다.

### 주요 산출물

1. **디자인 메타데이터** - Figma JSON 구조 전체
2. **컬러 팔레트** - CSS 변수 형태의 설계 토큰
3. **타이포그래피** - 폰트, 크기, 가중치 정보
4. **컴포넌트 라이브러리** - 재사용 가능한 설계 요소
5. **이미지 자산** - 고해상도 PNG 파일
6. **접근성 검증** - WCAG 2.1 AA 규정 확인
7. **분석 리포트** - 마크다운 형식의 종합 문서

---

## 🚀 빠른 시작

### 1단계: Figma 메타데이터 추출 (5분)

**Option A: UI 클릭 (가장 간단)**

```bash
# 1. Figma 파일 열기
# https://www.figma.com/file/m2odCIWVPWv84ygT5w43Ur

# 2. 노드 689:1242 선택 → 우클릭 → Copy as JSON

# 3. 파일에 저장
cat > .moai/research/figma-metadata.json << 'EOF'
# 여기에 복사된 JSON 붙여넣기
EOF
```

**Option B: API (자동화)**

```bash
# 토큰 설정
export FIGMA_TOKEN=your_personal_access_token

# 메타데이터 다운로드
curl -H "X-Figma-Token: $FIGMA_TOKEN" \
  "https://api.figma.com/v1/files/m2odCIWVPWv84ygT5w43Ur/nodes?ids=689:1242" \
  > .moai/research/figma-metadata.json
```

**개인 액세스 토큰 발급**: https://www.figma.com/settings/account

### 2단계: 분석 실행 (2분)

```bash
# 색상, 타이포그래피, 컴포넌트 분석
# CSS 토큰 및 마크다운 리포트 생성
uv run .moai/research/figma_analyzer.py \
  --json .moai/research/figma-metadata.json \
  --analyze colors,typography,components,images \
  --output .moai/research \
  --css \
  --report
```

**생성 결과**:
```
Extracted 24 colors
Extracted 12 typography styles
Found 8 components
Found 3 images
✅ CSS tokens saved to .moai/research/design-tokens.css
✅ Report saved to .moai/research/analysis-report.md
✅ Metadata saved to .moai/research/analysis-metadata.json
```

### 3단계: 결과 검토 (3분)

```bash
# 생성된 CSS 변수 확인
cat .moai/research/design-tokens.css

# 상세 분석 리포트 확인
cat .moai/research/analysis-report.md

# 메타데이터 구조 확인
cat .moai/research/analysis-metadata.json | python3 -m json.tool
```

---

## 📚 문서 가이드

| 문서 | 내용 | 대상 |
|------|------|------|
| **FIGMA_EXTRACTION_GUIDE.md** | 상세한 추출 방법 (7가지 방법 설명) | 처음 사용자 |
| **FIGMA_WORKFLOW.md** | 단계별 워크플로우 (5 Phase) | 구현자 |
| **figma_analyzer.py** | Python 분석 스크립트 | 자동화 담당자 |
| **design-system.md** | 생성된 디자인 시스템 | 디자이너 |
| **README.md** | 이 문서 (개요) | 모든 사용자 |

---

## 🔄 전체 워크플로우

```
┌─────────────────────────────────────────────────────────┐
│  Phase 1: 메타데이터 수집 (10분)                        │
│  - Figma 파일 접근                                       │
│  - 노드 689:1242 JSON 추출                               │
│  - .moai/research/figma-metadata.json 저장              │
└────────────────────┬────────────────────────────────────┘
                     ↓
┌─────────────────────────────────────────────────────────┐
│  Phase 2: 색상 및 타이포그래피 분석 (15분)             │
│  - figma_analyzer.py 실행                                │
│  - 색상 팔레트 추출                                      │
│  - 타이포그래피 정보 분석                                │
│  - CSS 변수 생성                                         │
└────────────────────┬────────────────────────────────────┘
                     ↓
┌─────────────────────────────────────────────────────────┐
│  Phase 3: 이미지 자산 다운로드 (10분)                   │
│  - 고해상도 PNG 2배 내보내기                             │
│  - .moai/research/figma-assets/ 저장                    │
└────────────────────┬────────────────────────────────────┘
                     ↓
┌─────────────────────────────────────────────────────────┐
│  Phase 4: 접근성 검증 (15분)                            │
│  - WCAG AA 색상 대조 확인                                │
│  - 키보드 네비게이션 검증                                │
│  - 터치 대상 크기 확인                                   │
└────────────────────┬────────────────────────────────────┘
                     ↓
┌─────────────────────────────────────────────────────────┐
│  Phase 5: 문서화 (20분)                                 │
│  - 최종 분석 리포트 생성                                 │
│  - 디자인 시스템 문서작성                                │
│  - 구현 가이드 제시                                      │
└────────────────────┬────────────────────────────────────┘
                     ↓
┌─────────────────────────────────────────────────────────┐
│  Phase 6: React 컴포넌트 생성 (MOAI 에이전트)          │
│  - /moai:1-plan "React 컴포넌트 생성"                   │
│  - /moai:2-run SPEC-XXX                                 │
│  - Storybook 통합                                        │
│  - Playwright 테스트                                     │
└─────────────────────────────────────────────────────────┘
```

**소요 시간**: 약 1-2시간 (자동화)

---

## 📂 파일 구조

```
.moai/research/
│
├── 📖 문서 (Guides & Docs)
│   ├── README.md                          ← 현재 파일
│   ├── FIGMA_EXTRACTION_GUIDE.md          ← 상세 추출 가이드
│   ├── FIGMA_WORKFLOW.md                  ← 5-Phase 워크플로우
│   ├── figma-node-689-1242-analysis.md   ← 최종 분석 리포트
│   └── design-system.md                   ← 디자인 시스템
│
├── 🔧 도구 (Tools)
│   ├── figma_analyzer.py                  ← Python 분석 스크립트
│   └── scripts/
│       ├── fetch-figma-metadata.sh        ← 메타데이터 다운로드
│       ├── download-figma-images.sh       ← 이미지 다운로드
│       └── check-contrast.py              ← WCAG 대조 검증
│
├── 📊 데이터 (Extracted Data)
│   ├── figma-metadata.json                ← Figma JSON 메타데이터
│   ├── figma-images.json                  ← 이미지 URL 목록
│   ├── design-tokens.css                  ← CSS 변수
│   ├── analysis-report.md                 ← 분석 리포트
│   ├── analysis-metadata.json             ← 추출 결과 메타데이터
│   └── color-palette.json                 ← 색상 팔레트
│
└── 🎨 자산 (Assets)
    └── figma-assets/
        ├── node-689-1242.png              ← 1배 해상도
        ├── node-689-1242@2x.png           ← 2배 해상도
        └── images/                        ← 삽입된 이미지들
```

---

## 🎯 주요 기능

### 1. 색상 추출 및 분석

```css
/* 자동 생성된 CSS 변수 */
:root {
  --color-primary-500: #0EA5E9;
  --color-text: #0F172A;
  --color-background: #F8FAFC;
  --color-success: #10B981;
  --color-error: #EF4444;
  /* ... 20+ 색상 */
}
```

### 2. 타이포그래피 관리

```css
:root {
  --font-family-sans: -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
  --font-size-lg: 18px;
  --font-size-xl: 20px;
  --line-height-tight: 1.2;
  --line-height-normal: 1.5;
  /* ... 15+ 타이포그래피 토큰 */
}
```

### 3. 접근성 검증

```python
색상 대조 비율: 15.3:1
✅ WCAG AA (4.5:1): PASS
✅ WCAG AAA (7.0:1): PASS

색상 조합 검증:
✅ AA | ✅ AAA | 15.3:1 | Primary Text on Background
✅ AA | ✅ AAA | 5.2:1  | Primary CTA on White
```

### 4. 컴포넌트 카탈로그

```markdown
## 추출된 컴포넌트

- **Button** (Primary, Secondary, Ghost)
  - 크기: Small (32px), Medium (40px), Large (48px)
  - 상태: Default, Hover, Active, Disabled

- **Input** (Text, Email, Password)
  - 상태: Normal, Focus, Error, Disabled

- **Card** (Hero, Feature, Product)
  - 레이아웃: Vertical, Horizontal
  - 이미지: Optional, Required
```

---

## 💻 사용 예제

### 예제 1: 분석 및 CSS 토큰 생성

```bash
# 메타데이터 추출
export FIGMA_TOKEN=figd_XXXXXXXXXXXX
curl -H "X-Figma-Token: $FIGMA_TOKEN" \
  "https://api.figma.com/v1/files/m2odCIWVPWv84ygT5w43Ur/nodes?ids=689:1242" \
  > .moai/research/figma-metadata.json

# 분석 및 CSS 생성
uv run .moai/research/figma_analyzer.py \
  --json .moai/research/figma-metadata.json \
  --css

# 결과 확인
cat .moai/research/design-tokens.css
```

### 예제 2: 이미지 자산 다운로드

```bash
# 1. 이미지 URL 조회
curl -H "X-Figma-Token: $FIGMA_TOKEN" \
  "https://api.figma.com/v1/images/m2odCIWVPWv84ygT5w43Ur?ids=689:1242&scale=2" \
  > .moai/research/figma-images.json

# 2. URL에서 다운로드
cat .moai/research/figma-images.json | python3 << 'EOF'
import json
import urllib.request
import sys
from pathlib import Path

data = json.load(sys.stdin)
Path(".moai/research/figma-assets").mkdir(exist_ok=True)

for node_id, url in data.get("images", {}).items():
    filename = f".moai/research/figma-assets/{node_id}@2x.png"
    urllib.request.urlretrieve(url, filename)
    print(f"✅ Downloaded {filename}")
EOF
```

### 예제 3: React 컴포넌트 생성 (MOAI)

```bash
# 분석 완료 후, 컴포넌트 생성 시작
/moai:1-plan "노드 689:1242 기반 React 컴포넌트 구현

요구사항:
- TypeScript 지원
- Tailwind CSS 스타일링
- 접근성 (WCAG 2.1 AA)
- 반응형 디자인
- Storybook 통합
- Playwright 테스트"

# 토큰 절약
/clear

# TDD 구현
/moai:2-run SPEC-001
```

---

## ✅ 체크리스트

### 초기 설정

- [ ] Figma 파일 열기: https://www.figma.com/file/m2odCIWVPWv84ygT5w43Ur
- [ ] 개인 액세스 토큰 발급 (Settings → Personal access tokens)
- [ ] 환경 변수 설정: `export FIGMA_TOKEN=...`

### Phase 1: 메타데이터 추출

- [ ] 메타데이터 JSON 다운로드
- [ ] `.moai/research/figma-metadata.json` 저장
- [ ] 파일 크기 확인 (1MB 이상 정상)

### Phase 2: 분석 실행

- [ ] `figma_analyzer.py` 실행
- [ ] 색상 추출 확인 (20+ colors)
- [ ] 타이포그래피 분석 확인 (10+ styles)
- [ ] CSS 토큰 생성 확인

### Phase 3: 이미지 다운로드

- [ ] 이미지 URL 조회
- [ ] `figma-assets/` 디렉토리에 저장
- [ ] PNG 파일 확인

### Phase 4: 접근성 검증

- [ ] WCAG AA 색상 대조 검증
- [ ] 모든 조합 4.5:1 이상 확인
- [ ] 문서화 완료

### Phase 5: 문서화

- [ ] 분석 리포트 생성
- [ ] 디자인 시스템 문서작성
- [ ] README 및 가이드 검토

### Phase 6: 컴포넌트 생성

- [ ] SPEC 생성: `/moai:1-plan`
- [ ] `/clear` 실행
- [ ] TDD 구현: `/moai:2-run SPEC-XXX`
- [ ] Storybook 통합
- [ ] 테스트 작성

---

## 🆘 문제 해결

### Q1: "FIGMA_TOKEN not found"
**A**: 토큰 설정 필요
```bash
export FIGMA_TOKEN=your_personal_access_token
# 또는 .env에 저장 (⚠️ .gitignore 추가)
```

### Q2: "JSON 파일이 너무 크다"
**A**: 정상입니다. Figma 전체 구조 포함. 분석에 시간 걸릴 수 있음.
```bash
# 파일 크기 확인
du -h .moai/research/figma-metadata.json
```

### Q3: "색상이 추출되지 않음"
**A**: 노드에 Fill이 없을 수 있음. 파일 구조 확인:
```bash
grep -o '"fills"' .moai/research/figma-metadata.json | wc -l
```

### Q4: "이미지 다운로드 실패"
**A**: URL 만료 가능. 새로 생성:
```bash
curl -H "X-Figma-Token: $FIGMA_TOKEN" \
  "https://api.figma.com/v1/images/m2odCIWVPWv84ygT5w43Ur?ids=689:1242&scale=2&format=png" \
  > .moai/research/figma-images.json
```

---

## 📞 다음 단계

1. **즉시 (오늘)**: 위의 "빠른 시작" 3단계 완료
2. **내일**: React 컴포넌트 생성 시작 (`/moai:1-plan`)
3. **다음주**: Storybook + Playwright 테스트 통합

---

## 📚 참고 자료

- **Figma API 문서**: https://www.figma.com/developers/api
- **WCAG 2.1**: https://www.w3.org/WAI/WCAG21/quickref/
- **React 컴포넌트 패턴**: https://react.dev
- **Tailwind CSS**: https://tailwindcss.com
- **Storybook**: https://storybook.js.org

---

**Last Updated**: 2025-11-19
**Version**: 1.0.0
**Status**: 준비 완료
**Contact**: claude@anthropic.com
