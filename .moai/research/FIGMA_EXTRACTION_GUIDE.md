# Figma Design 추출 및 분석 가이드

**프로젝트**: 모듈의사주-프로젝트-와디즈
**파일 키**: m2odCIWVPWv84ygT5w43Ur
**노드 ID**: 689:1242
**작성 날짜**: 2025-11-19

---

## 1단계: Figma 파일 접근 및 메타데이터 추출

### 방법 1: Figma 플러그인 사용 (권장)

**설치 단계**:
1. Figma 파일 열기: https://www.figma.com/file/m2odCIWVPWv84ygT5w43Ur
2. **플러그인 > 커뮤니티 검색** → "Export JSON"
3. JSON으로 전체 문서 구조 다운로드
4. `.moai/research/figma-metadata.json` 저장

**추출할 정보**:
```json
{
  "node_id": "689:1242",
  "node_name": "[추출]",
  "node_type": "[Frame|Component|Group|Board]",
  "bounds": {
    "x": 0,
    "y": 0,
    "width": 0,
    "height": 0
  },
  "fills": [
    {
      "type": "SOLID",
      "color": "#RRGGBB",
      "opacity": 1.0
    }
  ],
  "strokes": [
    {
      "type": "SOLID",
      "color": "#RRGGBB",
      "weight": 2
    }
  ],
  "effects": [],
  "children": [
    {
      "id": "...",
      "name": "...",
      "type": "...",
      "bounds": { ... }
    }
  ]
}
```

### 방법 2: Figma REST API 사용

**준비**:
- Figma 개인 액세스 토큰 필요 (Settings → Account → Personal access tokens)
- curl 또는 Python 스크립트로 API 호출

**API 엔드포인트**:
```bash
# 파일 메타데이터
curl -H "X-Figma-Token: YOUR_TOKEN" \
  "https://api.figma.com/v1/files/m2odCIWVPWv84ygT5w43Ur"

# 특정 노드 정보
curl -H "X-Figma-Token: YOUR_TOKEN" \
  "https://api.figma.com/v1/files/m2odCIWVPWv84ygT5w43Ur/nodes?ids=689:1242"

# 이미지 URL 생성
curl -H "X-Figma-Token: YOUR_TOKEN" \
  "https://api.figma.com/v1/images/m2odCIWVPWv84ygT5w43Ur?ids=689:1242&scale=2&format=png"
```

### 방법 3: Figma 설계 파일 직접 다운로드

**수동 단계**:
1. Figma 파일 열기
2. 노드 `689:1242` 선택
3. **우클릭 → Copy as JSON** (한 번에 내보내기)
4. 텍스트 에디터에 붙여넣기 → `.moai/research/node-689-1242.json` 저장

---

## 2단계: 이미지 자산 다운로드

### 고해상도 PNG 내보내기

**Figma UI에서**:
1. 노드 `689:1242` 선택
2. **오른쪽 패널 → Export (내보내기)**
3. 설정:
   - Format: PNG
   - Scale: 2x (고해상도)
   - Suffix: @2x
4. **Export 클릭** → `.moai/research/figma-assets/` 저장

**자산 구조**:
```
figma-assets/
├── node-689-1242.png        # 메인 컴포넌트
├── node-689-1242@2x.png     # 2배 해상도
├── component-variants/       # 컴포넌트 변형
│   ├── variant-1.png
│   ├── variant-2.png
│   └── variant-3.png
└── images/                   # 삽입된 이미지
    ├── image-1.png
    └── image-2.png
```

---

## 3단계: 컬러 팔레트 및 디자인 토큰 추출

### 색상 분석

```python
#!/usr/bin/env python3
"""
Figma 색상 추출 스크립트
"""

import json
import re

def extract_colors(figma_json):
    """Figma JSON에서 모든 색상 추출"""
    colors = {}

    def traverse(node, path=""):
        if "fills" in node:
            for fill in node["fills"]:
                if fill.get("type") == "SOLID":
                    color = fill.get("color", {})
                    hex_color = rgb_to_hex(color)
                    colors[f"{path}/{node['name']}"] = {
                        "hex": hex_color,
                        "opacity": fill.get("opacity", 1.0),
                        "rgba": color
                    }

        if "children" in node:
            for child in node["children"]:
                traverse(child, f"{path}/{node['name']}")

    traverse(figma_json)
    return colors

def rgb_to_hex(color_dict):
    """RGB를 Hex로 변환"""
    r = int(color_dict.get("r", 0) * 255)
    g = int(color_dict.get("g", 0) * 255)
    b = int(color_dict.get("b", 0) * 255)
    return f"#{r:02x}{g:02x}{b:02x}"

# 사용법
with open(".moai/research/figma-metadata.json") as f:
    figma_data = json.load(f)

colors = extract_colors(figma_data)

# 저장
with open(".moai/research/color-palette.json", "w") as f:
    json.dump(colors, f, indent=2)
```

### 디자인 토큰 (CSS 변수)

```css
/* .moai/research/design-tokens.css */

:root {
  /* Primary Colors */
  --color-primary-50: #F0F9FF;
  --color-primary-500: #0EA5E9;
  --color-primary-900: #0C2D4A;

  /* Semantic Colors */
  --color-success: #10B981;
  --color-error: #EF4444;
  --color-warning: #F59E0B;
  --color-info: #3B82F6;

  /* Spacing Scale (8px base) */
  --spacing-xs: 4px;
  --spacing-sm: 8px;
  --spacing-md: 16px;
  --spacing-lg: 24px;
  --spacing-xl: 32px;
  --spacing-2xl: 48px;

  /* Typography */
  --font-family-sans: -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
  --font-family-mono: "Monaco", "Menlo", "Courier New", monospace;

  /* Font Sizes */
  --font-size-xs: 12px;
  --font-size-sm: 14px;
  --font-size-base: 16px;
  --font-size-lg: 18px;
  --font-size-xl: 20px;
  --font-size-2xl: 24px;
  --font-size-3xl: 30px;
  --font-size-4xl: 36px;

  /* Line Heights */
  --line-height-tight: 1.2;
  --line-height-normal: 1.5;
  --line-height-relaxed: 1.75;

  /* Border Radius */
  --radius-none: 0;
  --radius-sm: 2px;
  --radius-md: 4px;
  --radius-lg: 8px;
  --radius-xl: 12px;
  --radius-full: 9999px;

  /* Shadows */
  --shadow-sm: 0 1px 2px 0 rgba(0, 0, 0, 0.05);
  --shadow-md: 0 4px 6px -1px rgba(0, 0, 0, 0.1);
  --shadow-lg: 0 10px 15px -3px rgba(0, 0, 0, 0.1);
  --shadow-xl: 0 20px 25px -5px rgba(0, 0, 0, 0.1);
}
```

---

## 4단계: 컴포넌트 구조 분석

### 컴포넌트 매핑

```markdown
### 노드 689:1242 구조

**타입**: [Frame|Component|Group]
**이름**: [추출]
**크기**: [width]x[height]px

#### 자식 요소

| ID | 이름 | 타입 | 크기 | 역할 |
|---|---|---|---|---|
| xxx | [이름] | Frame | 100x50 | [설명] |
| xxx | [이름] | Text | 80x20 | [설명] |
| xxx | [이름] | Image | 60x60 | [설명] |
| xxx | [이름] | Group | 150x100 | [설명] |

#### 레이어 계층 (Tree View)

```
689:1242 (Frame)
├── Header (Group)
│   ├── Logo (Image)
│   ├── Title (Text)
│   └── Subtitle (Text)
├── Content (Group)
│   ├── Card 1 (Component)
│   ├── Card 2 (Component)
│   └── Card 3 (Component)
└── Footer (Group)
    ├── Copyright (Text)
    └── Links (Group)
```

### 컴포넌트 변형 (Variants)

```markdown
#### 변형 분석

**Button 컴포넌트**:
- 크기: Small (32px), Medium (40px), Large (48px)
- 상태: Default, Hover, Active, Disabled
- 종류: Primary, Secondary, Ghost, Danger

**변형 조합**: 3 × 4 × 4 = 48개

**선택 우선순위**:
1. Primary / Medium / Default (가장 일반적)
2. Secondary / Medium / Default
3. Ghost / Small / Disabled
```

---

## 5단계: 접근성 및 색상 대조 검증

### WCAG 2.1 AA 규정 확인

```python
#!/usr/bin/env python3
"""
색상 대조 검증 스크립트
"""

from colorsys import rgb_to_hls

def contrast_ratio(color1, color2):
    """WCAG 대조 비율 계산"""
    def luminance(rgb):
        r, g, b = [x / 255.0 for x in rgb]
        r = r / 12.92 if r <= 0.03928 else ((r + 0.055) / 1.055) ** 2.4
        g = g / 12.92 if g <= 0.03928 else ((g + 0.055) / 1.055) ** 2.4
        b = b / 12.92 if b <= 0.03928 else ((b + 0.055) / 1.055) ** 2.4
        return 0.2126 * r + 0.7152 * g + 0.0722 * b

    l1 = luminance(color1)
    l2 = luminance(color2)
    lmax, lmin = max(l1, l2), min(l1, l2)
    return round((lmax + 0.05) / (lmin + 0.05), 2)

# 사용 예
text_color = (15, 23, 42)      # #0F172A
background = (248, 250, 252)   # #F8FAFC

ratio = contrast_ratio(text_color, background)
print(f"대조 비율: {ratio}:1")
print(f"WCAG AA 준수: {'✅' if ratio >= 4.5 else '❌'}")
print(f"WCAG AAA 준수: {'✅' if ratio >= 7.0 else '❌'}")
```

### 체크리스트

```markdown
## 접근성 검증 체크리스트

- [ ] 모든 텍스트 색상 대조 4.5:1 이상 (WCAG AA)
- [ ] 활성화 상태 포커스 지시자 명확
- [ ] 아이콘 + 텍스트로 정보 전달 (색상만으로 아님)
- [ ] 버튼 최소 크기 44x44px
- [ ] 터치 대상 간 최소 8px 여백
- [ ] 밝기 기반 색상 대비 확인
```

---

## 6단계: 코드 생성 가능성 평가

### 난이도 분류

| 난이도 | 특징 | 예상 시간 | 권장 라이브러리 |
|---|---|---|---|
| **쉬움** | 단순 텍스트, 기본 도형 | 1-2시간 | HTML/CSS |
| **중간** | 그리드 레이아웃, 기본 컴포넌트 | 4-8시간 | React, Vue, Tailwind |
| **어려움** | 복잡한 상호작용, 애니메이션 | 1-2일 | React, Framer Motion, Storybook |
| **매우 어려움** | 실시간 렌더링, WebGL, 맞춤형 | 1주+ | Three.js, Canvas API |

### 코드 생성 평가

```markdown
#### 노드 689:1242 분석

**유형**: [추출]
**복잡도**: [낮음/중간/높음]
**예상 라인 수**: [추정]

**코드 생성 가능성**:
- ✅ HTML/CSS: [가능/부분 가능/어려움]
- ✅ React: [가능/부분 가능/어려움]
- ✅ Vue: [가능/부분 가능/어려움]
- ✅ Tailwind: [가능/부분 가능/어려움]
- ✅ CSS-in-JS: [가능/부분 가능/어려움]

**권장 기술 스택**:
1. React 19 + TypeScript
2. Tailwind CSS 또는 Styled Components
3. Framer Motion (애니메이션)
4. Storybook (컴포넌트 라이브러리)

**생성 순서**:
1. 레이아웃 구조 (HTML)
2. 스타일링 (CSS/Tailwind)
3. 상호작용 (JavaScript/React)
4. 접근성 강화 (ARIA, 키보드 네비게이션)
5. 성능 최적화 (code splitting, lazy loading)
```

---

## 7단계: 최종 분석 리포트 작성

### 리포트 템플릿

```markdown
# Figma 디자인 분석 리포트: 노드 689:1242

## 📊 기본 정보

- **노드명**: [추출]
- **노드 타입**: [Frame/Component/Group]
- **파일**: 모듈의사주-프로젝트-와디즈
- **파일 키**: m2odCIWVPWv84ygT5w43Ur
- **검증 일시**: 2025-11-19
- **추출자**: [이름]

## 🎨 디자인 요소

### 색상 팔레트

| 이름 | Hex | RGB | 사용 |
|---|---|---|---|
| Primary | #0EA5E9 | (14, 165, 233) | 주요 CTA, 강조 |
| Text | #0F172A | (15, 23, 42) | 본문 텍스트 |
| Background | #F8FAFC | (248, 250, 252) | 배경 |
| Success | #10B981 | (16, 185, 129) | 성공 메시지 |
| Error | #EF4444 | (239, 68, 68) | 에러 메시지 |

### 타이포그래피

| 이름 | 폰트 | 크기 | 두께 | 라인높이 | 사용 |
|---|---|---|---|---|---|
| Heading L | -apple-system | 32px | 700 | 1.25 | h1, h2 |
| Body | -apple-system | 16px | 400 | 1.5 | 본문 |
| Caption | -apple-system | 12px | 500 | 1.25 | 작은 레이블 |

### 간격 시스템

- xs: 4px
- sm: 8px
- md: 16px
- lg: 24px
- xl: 32px
- 2xl: 48px

## 📐 구조 분석

### 계층 구조

```
노드 689:1242
├── Header (Group)
├── Content (Group)
└── Footer (Group)
```

- **총 자식 노드**: N개
- **그룹**: N개
- **컴포넌트 참조**: N개
- **이미지**: N개
- **텍스트**: N개

### 크기 및 위치

| 속성 | 값 |
|---|---|
| Width | 0px |
| Height | 0px |
| X Position | 0px |
| Y Position | 0px |

## ♿ 접근성 검증

| 항목 | 상태 | 노트 |
|---|---|---|
| 색상 대조 (WCAG AA) | ✅ | 4.5:1 이상 |
| 포커스 표시자 | ✅ | 명확함 |
| 키보드 네비게이션 | ✅ | 모든 요소 탭 접근 가능 |
| 스크린 리더 지원 | ⚠️ | 요청: alt 텍스트 추가 |
| 터치 대상 크기 | ✅ | 44x44px 이상 |

## 💻 코드 생성 가능성

- **난이도**: [낮음/중간/높음]
- **예상 구현 시간**: N-M시간
- **예상 코드 라인**: ~N라인 (공백 포함)
- **권장 라이브러리**: React + Tailwind + Framer Motion

## 📦 산출물

### 다운로드

- 이미지: `.moai/research/figma-assets/node-689-1242.png` (2x)
- 메타데이터: `.moai/research/figma-metadata.json`
- 색상: `.moai/research/color-palette.json`
- 토큰: `.moai/research/design-tokens.css`

### 코드 생성 (권장)

```bash
# React 컴포넌트 생성
/moai:1-plan "노드 689:1242 React 컴포넌트 구현"

# 또는 수동 생성
uv run src/figma_to_code.py \
  --file-key m2odCIWVPWv84ygT5w43Ur \
  --node-id 689:1242 \
  --output ./src/components/
```

## ✅ 권장사항

1. **즉시**: Figma에서 JSON 메타데이터 다운로드
2. **다음**: 이미지 자산 2배 해상도로 내보내기
3. **검증**: WCAG AA 색상 대조 확인
4. **코드**: React 컴포넌트 생성 (Storybook 포함)
5. **테스트**: Playwright로 시각 회귀 테스트

---

**다음 단계**: `.moai/research/` 디렉토리의 다른 파일 참고
**지원**: 추가 분석이 필요하면 새로운 SPEC 생성 (`/moai:1-plan`)
```

---

## 파일 저장 위치

| 파일 | 경로 | 설명 |
|---|---|---|
| 메타데이터 | `.moai/research/figma-metadata.json` | Figma JSON 구조 |
| 색상 팔레트 | `.moai/research/color-palette.json` | 추출된 색상 |
| 디자인 토큰 | `.moai/research/design-tokens.css` | CSS 변수 |
| 이미지 자산 | `.moai/research/figma-assets/` | PNG, SVG 등 |
| 분석 리포트 | `.moai/research/figma-node-689-1242-analysis.md` | 최종 보고서 |

---

## 다음 단계

1. **Figma 파일 접근**
   - 파일 링크: https://www.figma.com/file/m2odCIWVPWv84ygT5w43Ur
   - 개인 액세스 토큰 발급 (필요시)

2. **메타데이터 추출**
   - 플러그인 또는 REST API 사용
   - `.moai/research/` 폴더에 저장

3. **분석 스크립트 실행**
   - Python 스크립트로 색상, 레이아웃 분석
   - 접근성 검증 자동화

4. **코드 생성**
   - React 컴포넌트 생성 (`/moai:2-run`)
   - Storybook 통합
   - Playwright 테스트

5. **최종 리포트**
   - 분석 결과 정리
   - 권장사항 제시
   - 구현 계획 수립

---

**문의**: Figma 추출 과정에서 막히는 부분이 있으면 언제든지 물어보세요!
